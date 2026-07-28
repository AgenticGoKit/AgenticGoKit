package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/agenticgokit/agenticgokit/internal/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// DefaultOpenAIBaseURL is the default OpenAI API endpoint
const DefaultOpenAIBaseURL = "https://api.openai.com/v1"

// APIStatusError is returned when the OpenAI-compatible endpoint responds
// with a non-200 status. StatusCode is preserved (previously discarded —
// the error was a flat string built from the response body, giving callers
// no structured way to classify e.g. a 502/503/524 gateway blip as
// retry-worthy vs. a genuine 4xx client error) so downstream code can use
// errors.As instead of matching substrings against the error text.
type APIStatusError struct {
	StatusCode int
	Body       string
}

func (e *APIStatusError) Error() string {
	return fmt.Sprintf("OpenAI API error (status %d): %s", e.StatusCode, e.Body)
}

// OpenAIAdapterConfig holds extended configuration for OpenAI-compatible adapters
type OpenAIAdapterConfig struct {
	APIKey       string
	Model        string
	MaxTokens    int
	Temperature  float32
	BaseURL      string            // Custom base URL (for vLLM, MLFlow, etc.)
	ExtraHeaders map[string]string // Custom headers (for gateways)
	HTTPTimeout  time.Duration     // HTTP client timeout

	// Extended sampling parameters (for vLLM compatibility)
	TopP              float32
	TopK              int
	PresencePenalty   float32
	FrequencyPenalty  float32
	RepetitionPenalty float32
	Stop              []string

	// ResponseFormat, when non-nil, is passed through verbatim as the
	// OpenAI-compatible "response_format" request field (e.g.
	// map[string]interface{}{"type": "json_object"}). nil (the zero
	// value) omits the field entirely — no behavior change for existing
	// callers.
	ResponseFormat interface{}

	// CachePrompt, when true, sets the request body's "cache_prompt" field —
	// llama.cpp's server flag to reuse a matching KV-cache prefix instead of
	// re-prefilling it. false (the zero value) omits the field entirely, so
	// non-llama.cpp OpenAI-compatible backends never see it.
	CachePrompt bool
}

// OpenAIAdapter implements the ModelProvider interface for OpenAI-compatible APIs.
// This adapter can be used for OpenAI, vLLM, MLFlow Gateway, and any other
// OpenAI-compatible endpoint by configuring the baseURL.
type OpenAIAdapter struct {
	apiKey       string
	model        string
	maxTokens    int
	temperature  float32
	baseURL      string            // Default: https://api.openai.com/v1
	extraHeaders map[string]string // Custom headers for gateways
	httpClient   *http.Client

	// Extended sampling parameters
	topP              float32
	topK              int
	presencePenalty   float32
	frequencyPenalty  float32
	repetitionPenalty float32
	stop              []string

	responseFormat interface{}
	cachePrompt    bool
}

// NewOpenAIAdapter creates a new OpenAIAdapter instance.
// SIGNATURE REMAINS UNCHANGED for backward compatibility.
func NewOpenAIAdapter(apiKey, model string, maxTokens int, temperature float32) (*OpenAIAdapter, error) {
	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
	}
	if model == "" {
		model = "gpt-4o-mini" // Default model
	}
	if maxTokens == 0 {
		maxTokens = 150 // Default max tokens
	}
	if temperature == 0 {
		temperature = 0.7 // Default temperature
	}

	return &OpenAIAdapter{
		apiKey:      apiKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		baseURL:     DefaultOpenAIBaseURL, // Default OpenAI URL
		httpClient:  NewOptimizedHTTPClient(120 * time.Second),
	}, nil
}

// NewOpenAIAdapterWithConfig creates an OpenAI-compatible adapter with extended configuration.
// Use this for vLLM, MLFlow Gateway, or any OpenAI-compatible endpoint.
func NewOpenAIAdapterWithConfig(config OpenAIAdapterConfig) (*OpenAIAdapter, error) {
	if config.Model == "" {
		return nil, errors.New("model is required")
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultOpenAIBaseURL
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 2048
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 120 * time.Second
	}

	return &OpenAIAdapter{
		apiKey:            config.APIKey,
		model:             config.Model,
		maxTokens:         config.MaxTokens,
		temperature:       config.Temperature,
		baseURL:           strings.TrimSuffix(config.BaseURL, "/"),
		extraHeaders:      config.ExtraHeaders,
		httpClient:        NewOptimizedHTTPClient(config.HTTPTimeout),
		topP:              config.TopP,
		topK:              config.TopK,
		presencePenalty:   config.PresencePenalty,
		frequencyPenalty:  config.FrequencyPenalty,
		repetitionPenalty: config.RepetitionPenalty,
		stop:              config.Stop,
		responseFormat:    config.ResponseFormat,
		cachePrompt:       config.CachePrompt,
	}, nil
}

// SetBaseURL allows overriding the default OpenAI endpoint (e.g., for vLLM or MLFlow)
func (o *OpenAIAdapter) SetBaseURL(url string) {
	if url != "" {
		o.baseURL = strings.TrimSuffix(url, "/")
	}
}

// SetExtraHeaders allows adding custom headers (e.g., for MLFlow Gateway)
func (o *OpenAIAdapter) SetExtraHeaders(headers map[string]string) {
	o.extraHeaders = headers
}

// setSessionHeader forwards a per-call session identifier as an outbound
// X-Session-Id header, when the caller supplied one via
// RunOptions.SessionID (agent_impl.go already threads this into ctx as
// context.WithValue(ctx, "session_id", ...), but nothing previously read it
// back out — every call looked session-less to the provider regardless of
// what the caller passed). Several OpenAI-compatible gateways (session-aware
// routers/caching proxies) use this header for sticky routing and cache
// affinity; a static extraHeaders map (set once at adapter construction,
// shared by every call for the adapter's lifetime) can't carry a value that
// varies per call, so this reads it from ctx instead, per call.
func setSessionHeader(req *http.Request, ctx context.Context) {
	if sid, ok := ctx.Value("session_id").(string); ok && sid != "" {
		req.Header.Set("X-Session-Id", sid)
	}
}

// setHeaders sets common headers for requests
func (o *OpenAIAdapter) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
	}
	for key, value := range o.extraHeaders {
		req.Header.Set(key, value)
	}
}

// getBaseURL returns the base URL, defaulting to OpenAI if not set
func (o *OpenAIAdapter) getBaseURL() string {
	if o.baseURL == "" {
		return DefaultOpenAIBaseURL
	}
	return o.baseURL
}

// Call implements the ModelProvider interface for a single request/response.
// buildChatRequestBody assembles the Chat Completions request body shared by
// Call (stream=false) and Stream (stream=true) — single source of truth so
// the two call sites can't drift apart on it again. "stream" is always sent
// explicitly, never omitted: some OpenAI-compatible gateways ignore an
// absent field and default to SSE regardless (observed live against a real
// gateway — a missing "stream" key on Call() made every response come back
// as "data: {...}" chunks, which the non-streaming response parser then
// failed to decode as a single JSON object, a json.SyntaxError on every
// call). vLLM/MLFlow Gateway/BentoML/AI Foundry Local embed *OpenAIAdapter
// for code reuse (see their own adapter files), so this fix covers them too
// without touching each one individually.
func (o *OpenAIAdapter) buildChatRequestBody(messages []map[string]interface{}, maxTokens int, temperature float32, stream bool) map[string]interface{} {
	reqBody := map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
		"stream":      stream,
	}
	// Extended sampling parameters, if set (for vLLM compatibility).
	if o.topP > 0 {
		reqBody["top_p"] = o.topP
	}
	if o.topK > 0 {
		reqBody["top_k"] = o.topK
	}
	if o.presencePenalty != 0 {
		reqBody["presence_penalty"] = o.presencePenalty
	}
	if o.frequencyPenalty != 0 {
		reqBody["frequency_penalty"] = o.frequencyPenalty
	}
	if o.repetitionPenalty != 0 {
		reqBody["repetition_penalty"] = o.repetitionPenalty
	}
	if len(o.stop) > 0 {
		reqBody["stop"] = o.stop
	}
	if o.responseFormat != nil {
		reqBody["response_format"] = o.responseFormat
	}
	if o.cachePrompt {
		reqBody["cache_prompt"] = true
	}
	return reqBody
}

func (o *OpenAIAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	// Create observability span
	tracer := otel.Tracer("agenticgokit.llm")
	ctx, span := tracer.Start(ctx, "llm.openai.call")
	defer span.End()

	// Set span attributes
	span.SetAttributes(
		attribute.String(observability.AttrLLMProvider, "openai"),
		attribute.String(observability.AttrLLMModel, o.model),
		attribute.Int(observability.AttrLLMMaxTokens, o.maxTokens),
		attribute.Float64(observability.AttrLLMTemperature, float64(o.temperature)),
	)

	// Track start time for latency
	startTime := time.Now()

	userPrompt := prompt.User
	if userPrompt == "" {
		err := errors.New("user prompt cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "empty user prompt")
		return Response{}, err
	}

	maxTokens := o.maxTokens
	if prompt.Parameters.MaxTokens != nil {
		maxTokens = int(*prompt.Parameters.MaxTokens)
		span.SetAttributes(attribute.Int(observability.AttrLLMMaxTokens, maxTokens))
	}
	temperature := o.temperature
	if prompt.Parameters.Temperature != nil {
		temperature = *prompt.Parameters.Temperature
		span.SetAttributes(attribute.Float64(observability.AttrLLMTemperature, float64(temperature)))
	}

	// Build messages array for Chat Completions API
	messages := []map[string]interface{}{}

	// Add system message if provided
	if prompt.System != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": prompt.System,
		})
	}

	// Construct user message content
	var userContent interface{}
	if len(prompt.Images) > 0 {
		// Multimodal content
		userContent = BuildMultimodalContent(userPrompt, prompt)
		span.SetAttributes(attribute.Bool("llm.multimodal", true))
		span.SetAttributes(attribute.Int("llm.image_count", len(prompt.Images)))
	} else {
		// Text-only content
		userContent = userPrompt
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userContent,
	})

	reqBody := o.buildChatRequestBody(messages, maxTokens, temperature, false)

	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal request")
		return Response{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.getBaseURL()+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create HTTP request")
		return Response{}, err
	}
	o.setHeaders(req)
	setSessionHeader(req, ctx)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "HTTP request failed")
		return Response{}, err
	}
	defer resp.Body.Close()

	// Record HTTP status
	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := &APIStatusError{StatusCode: resp.StatusCode, Body: string(body)}
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("API error: status %d", resp.StatusCode))
		return Response{}, err
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode response")
		return Response{}, err
	}

	if len(response.Choices) == 0 {
		err := errors.New("no completion choices returned")
		span.RecordError(err)
		span.SetStatus(codes.Error, "no choices in response")
		return Response{}, err
	}

	// Calculate latency
	latencyMs := time.Since(startTime).Milliseconds()

	// Record token usage and latency
	span.SetAttributes(
		attribute.Int(observability.AttrLLMPromptTokens, response.Usage.PromptTokens),
		attribute.Int(observability.AttrLLMCompletionTokens, response.Usage.CompletionTokens),
		attribute.Int(observability.AttrLLMTotalTokens, response.Usage.TotalTokens),
		attribute.Int64("llm.latency_ms", latencyMs),
		attribute.String("llm.finish_reason", response.Choices[0].FinishReason),
	)

	// Capture full prompt and response at detailed trace level
	if observability.IsDetailedTracing() {
		span.SetAttributes(
			attribute.String(observability.AttrPromptSystem, observability.TruncateForTrace(prompt.System, observability.MaxContentLength)),
			attribute.String(observability.AttrPromptUser, observability.TruncateForTrace(prompt.User, observability.MaxContentLength)),
			attribute.String(observability.AttrLLMResponse, observability.TruncateForTrace(response.Choices[0].Message.Content, observability.MaxContentLength)),
		)
	}

	span.SetStatus(codes.Ok, "LLM call successful")

	return Response{
		Content: response.Choices[0].Message.Content,
		Usage: UsageStats{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
		FinishReason: response.Choices[0].FinishReason,
	}, nil
}

// Stream implements the ModelProvider interface for streaming responses.
func (o *OpenAIAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	// Create observability span
	tracer := otel.Tracer("agenticgokit.llm")
	ctx, span := tracer.Start(ctx, "llm.openai.stream")
	defer span.End()

	// Set span attributes
	span.SetAttributes(
		attribute.String(observability.AttrLLMProvider, "openai"),
		attribute.String(observability.AttrLLMModel, o.model),
		attribute.Int(observability.AttrLLMMaxTokens, o.maxTokens),
		attribute.Float64(observability.AttrLLMTemperature, float64(o.temperature)),
		attribute.Bool("llm.streaming", true),
	)

	// Track start time for latency
	startTime := time.Now()

	userPrompt := prompt.User
	if userPrompt == "" {
		err := errors.New("user prompt cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "empty user prompt")
		return nil, err
	}

	maxTokens := o.maxTokens
	if prompt.Parameters.MaxTokens != nil {
		maxTokens = int(*prompt.Parameters.MaxTokens)
		span.SetAttributes(attribute.Int(observability.AttrLLMMaxTokens, maxTokens))
	}
	temperature := o.temperature
	if prompt.Parameters.Temperature != nil {
		temperature = *prompt.Parameters.Temperature
		span.SetAttributes(attribute.Float64(observability.AttrLLMTemperature, float64(temperature)))
	}

	// Build messages array for Chat Completions API
	messages := []map[string]interface{}{}

	// Add system message if provided
	if prompt.System != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": prompt.System,
		})
	}

	// Construct user message content
	var userContent interface{}
	if len(prompt.Images) > 0 {
		// Multimodal content
		userContent = BuildMultimodalContent(userPrompt, prompt)
		span.SetAttributes(attribute.Bool("llm.multimodal", true))
		span.SetAttributes(attribute.Int("llm.image_count", len(prompt.Images)))
	} else {
		// Text-only content
		userContent = userPrompt
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userContent,
	})

	reqBody := o.buildChatRequestBody(messages, maxTokens, temperature, true)

	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal request")
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP request for streaming
	req, err := http.NewRequestWithContext(ctx, "POST", o.getBaseURL()+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	o.setHeaders(req)
	setSessionHeader(req, ctx)

	// Make the request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "HTTP request failed")
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Record HTTP status
	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		// Read before Close — reading an already-closed response body
		// returns empty/error, which silently dropped the actual error body
		// on every non-200 stream-setup failure before this fix.
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		err := &APIStatusError{StatusCode: resp.StatusCode, Body: string(body)}
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("API error: status %d", resp.StatusCode))
		return nil, err
	}

	// Create token channel
	tokenChan := make(chan Token, 10)

	// Start goroutine to process streaming response
	go func() {
		defer close(tokenChan)
		defer resp.Body.Close()

		// Track chunks and total bytes
		chunkCount := 0
		totalBytes := 0
		var fullResponseBuilder strings.Builder

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue // Skip empty lines
			}

			// Check for context cancellation
			select {
			case <-ctx.Done():
				tokenChan <- Token{Error: ctx.Err()}
				span.RecordError(ctx.Err())
				span.SetStatus(codes.Error, "context canceled during stream")
				return
			default:
			}

			// Process SSE data lines
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
				if data == "[DONE]" {
					// Record final metrics
					latencyMs := time.Since(startTime).Milliseconds()
					span.SetAttributes(
						attribute.Int("llm.stream.chunk_count", chunkCount),
						attribute.Int("llm.stream.total_bytes", totalBytes),
						attribute.Int64("llm.latency_ms", latencyMs),
					)
					if observability.IsDetailedTracing() {
						span.SetAttributes(
							attribute.String(observability.AttrPromptSystem, observability.TruncateForTrace(prompt.System, observability.MaxContentLength)),
							attribute.String(observability.AttrPromptUser, observability.TruncateForTrace(prompt.User, observability.MaxContentLength)),
							attribute.String(observability.AttrLLMResponse, observability.TruncateForTrace(fullResponseBuilder.String(), observability.MaxContentLength)),
						)
					}
					span.SetStatus(codes.Ok, "stream completed successfully")
					return // Stream finished successfully
				}

				// Parse the JSON chunk
				var streamResponse struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
						FinishReason *string `json:"finish_reason"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(data), &streamResponse); err != nil {
					tokenChan <- Token{Error: fmt.Errorf("failed to decode stream chunk: %w", err)}
					span.RecordError(err)
					span.SetStatus(codes.Error, "failed to decode stream chunk")
					return
				}

				// Extract content delta
				if len(streamResponse.Choices) > 0 {
					content := streamResponse.Choices[0].Delta.Content
					if content != "" {
						chunkCount++
						totalBytes += len(content)
						fullResponseBuilder.WriteString(content)
						select {
						case tokenChan <- Token{Content: content}:
						case <-ctx.Done():
							tokenChan <- Token{Error: ctx.Err()}
							span.RecordError(ctx.Err())
							span.SetStatus(codes.Error, "context canceled during send")
							return
						}
					}
				}
			}
		}

		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			if ctx.Err() == nil {
				tokenChan <- Token{Error: fmt.Errorf("stream read error: %w", err)}
				span.RecordError(err)
				span.SetStatus(codes.Error, "stream read error")
			}
		}
	}()

	return tokenChan, nil
}

// Embeddings implements the ModelProvider interface for generating embeddings.
func (o *OpenAIAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	// Use the configured model for embeddings, or default to OpenAI's embedding model
	embeddingModel := o.model
	if o.baseURL == DefaultOpenAIBaseURL || o.baseURL == "" {
		// For OpenAI, use appropriate embedding model
		embeddingModel = "text-embedding-3-small"
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": embeddingModel,
		"input": texts,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.getBaseURL()+"/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	o.setHeaders(req)
	setSessionHeader(req, ctx)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIStatusError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var response struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Data) != len(texts) {
		return nil, errors.New("number of embeddings returned does not match input")
	}

	embeddings := make([][]float64, len(texts))
	for _, item := range response.Data {
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

// Model returns the model name
func (o *OpenAIAdapter) Model() string {
	return o.model
}

// BaseURL returns the base URL
func (o *OpenAIAdapter) BaseURL() string {
	return o.getBaseURL()
}
