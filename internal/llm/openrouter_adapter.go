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

// OpenRouterAdapter implements the ModelProvider interface for OpenRouter's API.
// OpenRouter provides unified access to multiple LLM providers through an OpenAI-compatible API.
type OpenRouterAdapter struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float32
	baseURL     string // Default: https://openrouter.ai/api/v1
	siteURL     string // Optional: for rankings (HTTP-Referer header)
	siteName    string // Optional: for rankings (X-Title header)
	httpClient  *http.Client

	// responseFormat/cachePrompt are set directly by factory.go
	// (createOpenRouterProvider), same package, rather than adding params
	// to NewOpenRouterAdapter's constructor and breaking its existing
	// positional-argument callers (e.g. wrappers.go's
	// NewOpenRouterAdapterWrapped).
	responseFormat interface{}
	cachePrompt    bool
}

// NewOpenRouterAdapter creates a new OpenRouterAdapter instance.
func NewOpenRouterAdapter(apiKey, model, baseURL string, maxTokens int, temperature float32, siteURL, siteName string) (*OpenRouterAdapter, error) {
	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
	}
	if model == "" {
		model = "openai/gpt-3.5-turbo" // Default model
	}
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	if maxTokens == 0 {
		maxTokens = 2000 // Default max tokens
	}
	if temperature == 0 {
		temperature = 0.7 // Default temperature
	}

	return &OpenRouterAdapter{
		apiKey:      apiKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		siteURL:     siteURL,
		siteName:    siteName,
		httpClient:  NewOptimizedHTTPClient(120 * time.Second),
	}, nil
}

// buildMessages converts our internal Prompt to OpenRouter's message format
func buildOpenRouterMessages(prompt Prompt) []map[string]interface{} {
	messages := []map[string]interface{}{}

	// Add system message if provided
	if prompt.System != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": prompt.System,
		})
	}

	// Build user message with potential multimodal content
	if prompt.User != "" || len(prompt.Images) > 0 || len(prompt.Audio) > 0 || len(prompt.Video) > 0 {
		userMessage := map[string]interface{}{
			"role": "user",
		}

		if len(prompt.Images) > 0 || len(prompt.Audio) > 0 || len(prompt.Video) > 0 {
			// Use shared multimodal content builder
			userMessage["content"] = BuildMultimodalContent(prompt.User, prompt)
		} else {
			// Text-only content
			userMessage["content"] = prompt.User
		}

		messages = append(messages, userMessage)
	}

	return messages
}

// buildChatRequestBody assembles the Chat Completions request body shared by
// Call (stream=false) and Stream (stream=true) — single source of truth so
// the two call sites can't drift apart on it, mirroring OpenAIAdapter's own
// buildChatRequestBody (openai_adapter.go). "stream" is always sent
// explicitly, never omitted: some OpenAI-compatible gateways ignore an
// absent field and default to SSE regardless (observed live — a missing
// "stream" key on a non-streaming Call() made the gateway return
// "data: {...}" chunks, which the non-streaming response parser then failed
// to decode as a single JSON object).
func (o *OpenRouterAdapter) buildChatRequestBody(messages []map[string]interface{}, maxTokens int, temperature float32, stream bool) map[string]interface{} {
	reqBody := map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
		"stream":      stream,
	}
	if o.responseFormat != nil {
		reqBody["response_format"] = o.responseFormat
	}
	if o.cachePrompt {
		reqBody["cache_prompt"] = true
	}
	return reqBody
}

// Call implements the ModelProvider interface for a single request/response.
func (o *OpenRouterAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	// Create observability span
	tracer := otel.Tracer("agenticgokit.llm")
	ctx, span := tracer.Start(ctx, "llm.openrouter.call")
	defer span.End()

	// Set span attributes
	span.SetAttributes(
		attribute.String(observability.AttrLLMProvider, "openrouter"),
		attribute.String(observability.AttrLLMModel, o.model),
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
	} else {
		span.SetAttributes(attribute.Int(observability.AttrLLMMaxTokens, o.maxTokens))
	}
	temperature := o.temperature
	if prompt.Parameters.Temperature != nil {
		temperature = *prompt.Parameters.Temperature
		span.SetAttributes(attribute.Float64(observability.AttrLLMTemperature, float64(temperature)))
	} else {
		span.SetAttributes(attribute.Float64(observability.AttrLLMTemperature, float64(o.temperature)))
	}

	// Track multimodal content
	if len(prompt.Images) > 0 || len(prompt.Audio) > 0 || len(prompt.Video) > 0 {
		span.SetAttributes(attribute.Bool("llm.multimodal", true))
		if len(prompt.Images) > 0 {
			span.SetAttributes(attribute.Int("llm.image_count", len(prompt.Images)))
		}
	}

	// Build messages array for Chat Completions API
	messages := buildOpenRouterMessages(prompt)

	reqBody := o.buildChatRequestBody(messages, maxTokens, temperature, false)

	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal request")
		return Response{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create HTTP request")
		return Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	// Set optional headers for rankings
	if o.siteURL != "" {
		req.Header.Set("HTTP-Referer", o.siteURL)
	}
	if o.siteName != "" {
		req.Header.Set("X-Title", o.siteName)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "HTTP request failed")
		return Response{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Record HTTP status
	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	// Returned as *APIStatusError (the same type openai_adapter.go/azure_
	// adapter.go produce), not a plain fmt.Errorf, so CircuitBreakerProvider's
	// DefaultIsRetryable can classify a 429/5xx as retryable via errors.As
	// instead of silently treating every OpenRouter error as non-retryable.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		// Try to parse OpenRouter error format
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Code    string `json:"code"`
				Type    string `json:"type"`
			} `json:"error"`
		}

		if json.Unmarshal(body, &errorResp) == nil && errorResp.Error.Message != "" {
			err := &APIStatusError{StatusCode: resp.StatusCode, Body: fmt.Sprintf("[%s]: %s", errorResp.Error.Code, errorResp.Error.Message)}
			span.RecordError(err)
			span.SetStatus(codes.Error, fmt.Sprintf("API error: %s", errorResp.Error.Code))
			return Response{}, err
		}

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
		return Response{}, fmt.Errorf("failed to decode response: %w", err)
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
func (o *OpenRouterAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	// Create observability span
	tracer := otel.Tracer("agenticgokit.llm")
	ctx, span := tracer.Start(ctx, "llm.openrouter.stream")
	defer span.End()

	// Set span attributes
	span.SetAttributes(
		attribute.String(observability.AttrLLMProvider, "openrouter"),
		attribute.String(observability.AttrLLMModel, o.model),
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
	} else {
		span.SetAttributes(attribute.Int(observability.AttrLLMMaxTokens, o.maxTokens))
	}
	temperature := o.temperature
	if prompt.Parameters.Temperature != nil {
		temperature = *prompt.Parameters.Temperature
		span.SetAttributes(attribute.Float64(observability.AttrLLMTemperature, float64(temperature)))
	} else {
		span.SetAttributes(attribute.Float64(observability.AttrLLMTemperature, float64(o.temperature)))
	}

	// Track multimodal content
	if len(prompt.Images) > 0 || len(prompt.Audio) > 0 || len(prompt.Video) > 0 {
		span.SetAttributes(attribute.Bool("llm.multimodal", true))
		if len(prompt.Images) > 0 {
			span.SetAttributes(attribute.Int("llm.image_count", len(prompt.Images)))
		}
	}

	// Build messages array for Chat Completions API
	messages := buildOpenRouterMessages(prompt)

	// Create streaming request
	streamReqBody := o.buildChatRequestBody(messages, maxTokens, temperature, true)

	requestBody, err := json.Marshal(streamReqBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal request")
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP request for streaming
	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	// Set optional headers for rankings
	if o.siteURL != "" {
		req.Header.Set("HTTP-Referer", o.siteURL)
	}
	if o.siteName != "" {
		req.Header.Set("X-Title", o.siteName)
	}

	// Make the request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "HTTP request failed")
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Record HTTP status
	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	// Read the body BEFORE closing it (previously closed first, the same
	// bug this PR's openai_adapter.go fix addresses for the OpenAI adapter
	// — left in place here meant a non-200 stream error came back with an
	// empty body every time). Also returned as *APIStatusError so
	// CircuitBreakerProvider's DefaultIsRetryable can classify a 429/5xx as
	// retryable.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Try to parse OpenRouter error format
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Code    string `json:"code"`
				Type    string `json:"type"`
			} `json:"error"`
		}

		if json.Unmarshal(body, &errorResp) == nil && errorResp.Error.Message != "" {
			err := &APIStatusError{StatusCode: resp.StatusCode, Body: fmt.Sprintf("[%s]: %s", errorResp.Error.Code, errorResp.Error.Message)}
			span.RecordError(err)
			span.SetStatus(codes.Error, fmt.Sprintf("API error: %s", errorResp.Error.Code))
			return nil, err
		}

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

		var chunkCount int
		var totalBytes int64
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
				span.RecordError(ctx.Err())
				span.SetStatus(codes.Error, "context cancelled")
				tokenChan <- Token{Error: ctx.Err()}
				return
			default:
			}

			// Process SSE data lines
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
				if data == "[DONE]" {
					// Calculate final metrics
					latencyMs := time.Since(startTime).Milliseconds()
					span.SetAttributes(
						attribute.Int("llm.chunk_count", chunkCount),
						attribute.Int64("llm.total_bytes", totalBytes),
						attribute.Int64("llm.latency_ms", latencyMs),
					)
					if observability.IsDetailedTracing() {
						span.SetAttributes(
							attribute.String(observability.AttrPromptSystem, observability.TruncateForTrace(prompt.System, observability.MaxContentLength)),
							attribute.String(observability.AttrPromptUser, observability.TruncateForTrace(prompt.User, observability.MaxContentLength)),
							attribute.String(observability.AttrLLMResponse, observability.TruncateForTrace(fullResponseBuilder.String(), observability.MaxContentLength)),
						)
					}
					span.SetStatus(codes.Ok, "stream completed")
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
					span.RecordError(err)
					span.SetStatus(codes.Error, "failed to decode chunk")
					tokenChan <- Token{Error: fmt.Errorf("failed to decode stream chunk: %w", err)}
					return
				}

				// Extract content delta
				if len(streamResponse.Choices) > 0 {
					content := streamResponse.Choices[0].Delta.Content
					if content != "" {
						chunkCount++
						totalBytes += int64(len(content))
						fullResponseBuilder.WriteString(content)

						select {
						case tokenChan <- Token{Content: content}:
						case <-ctx.Done():
							span.RecordError(ctx.Err())
							span.SetStatus(codes.Error, "context cancelled during streaming")
							tokenChan <- Token{Error: ctx.Err()}
							return
						}
					}
				}
			}
		}

		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			if ctx.Err() == nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "stream read error")
				tokenChan <- Token{Error: fmt.Errorf("stream read error: %w", err)}
			}
		}
	}()

	return tokenChan, nil
}

// Embeddings implements the ModelProvider interface for generating embeddings.
// Note: OpenRouter may not support embeddings for all models. This implementation
// returns an error indicating the feature is not supported. If needed, specific
// embedding models can be added in the future.
func (o *OpenRouterAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	// OpenRouter doesn't have a standardized embeddings endpoint across all providers
	// For now, return an error indicating this is not supported
	// Future enhancement: Could support specific embedding models if needed
	return nil, fmt.Errorf("embeddings not currently supported by OpenRouter adapter; use a dedicated embedding provider")
}
