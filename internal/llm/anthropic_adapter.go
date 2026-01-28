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

// DefaultAnthropicBaseURL is the default Anthropic API endpoint
const DefaultAnthropicBaseURL = "https://api.anthropic.com"

// DefaultAnthropicAPIVersion is the API version header required by Anthropic
const DefaultAnthropicAPIVersion = "2023-06-01"

// AnthropicAdapterConfig holds configuration for Anthropic adapters
type AnthropicAdapterConfig struct {
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float32
	BaseURL     string            // Custom base URL (for proxies)
	APIVersion  string            // API version (default: 2023-06-01)
	HTTPTimeout time.Duration     // HTTP client timeout
	TopP        float32           // Nucleus sampling
	TopK        int               // Top-K sampling
	Stop        []string          // Stop sequences
	System      string            // Default system prompt (optional)
	Metadata    map[string]string // Request metadata
}

// AnthropicAdapter implements the ModelProvider interface for Anthropic Claude APIs.
type AnthropicAdapter struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float32
	baseURL     string
	apiVersion  string
	httpClient  *http.Client
	topP        float32
	topK        int
	stop        []string
}

// NewAnthropicAdapter creates a new AnthropicAdapter instance.
func NewAnthropicAdapter(apiKey, model string, maxTokens int, temperature float32) (*AnthropicAdapter, error) {
	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
	}
	if model == "" {
		model = "claude-sonnet-4-20250514" // Latest model
	}
	if maxTokens == 0 {
		maxTokens = 1024 // Default max tokens
	}
	if temperature == 0 {
		temperature = 0.7 // Default temperature
	}

	return &AnthropicAdapter{
		apiKey:      apiKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		baseURL:     DefaultAnthropicBaseURL,
		apiVersion:  DefaultAnthropicAPIVersion,
		httpClient:  NewOptimizedHTTPClient(120 * time.Second),
	}, nil
}

// NewAnthropicAdapterWithConfig creates an Anthropic adapter with extended configuration.
func NewAnthropicAdapterWithConfig(config AnthropicAdapterConfig) (*AnthropicAdapter, error) {
	if config.APIKey == "" {
		return nil, errors.New("API key is required")
	}
	if config.Model == "" {
		config.Model = "claude-sonnet-4-20250514"
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultAnthropicBaseURL
	}
	if config.APIVersion == "" {
		config.APIVersion = DefaultAnthropicAPIVersion
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 1024
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 120 * time.Second
	}

	return &AnthropicAdapter{
		apiKey:      config.APIKey,
		model:       config.Model,
		maxTokens:   config.MaxTokens,
		temperature: config.Temperature,
		baseURL:     strings.TrimSuffix(config.BaseURL, "/"),
		apiVersion:  config.APIVersion,
		httpClient:  NewOptimizedHTTPClient(config.HTTPTimeout),
		topP:        config.TopP,
		topK:        config.TopK,
		stop:        config.Stop,
	}, nil
}

// setHeaders sets common headers for requests
func (a *AnthropicAdapter) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", a.apiVersion)
}

// getBaseURL returns the base URL
func (a *AnthropicAdapter) getBaseURL() string {
	if a.baseURL == "" {
		return DefaultAnthropicBaseURL
	}
	return a.baseURL
}

// buildMessages converts the Prompt into Anthropic's message format
func (a *AnthropicAdapter) buildMessages(prompt Prompt) (string, []map[string]interface{}) {
	systemPrompt := prompt.System
	messages := []map[string]interface{}{}

	// Build user message content
	var content interface{}
	if len(prompt.Images) > 0 {
		// Multimodal content with images
		contentParts := []map[string]interface{}{}

		// Add images first
		for _, img := range prompt.Images {
			if img.Base64 != "" {
				// Anthropic expects base64 without data URI prefix
				base64Data := img.Base64
				mediaType := "image/jpeg" // Default

				// Extract media type and data if it's a data URI
				if strings.HasPrefix(base64Data, "data:") {
					parts := strings.SplitN(base64Data, ",", 2)
					if len(parts) == 2 {
						// Extract media type from "data:image/jpeg;base64"
						mediaTypePart := strings.TrimPrefix(parts[0], "data:")
						mediaTypePart = strings.TrimSuffix(mediaTypePart, ";base64")
						if mediaTypePart != "" {
							mediaType = mediaTypePart
						}
						base64Data = parts[1]
					}
				}

				contentParts = append(contentParts, map[string]interface{}{
					"type": "image",
					"source": map[string]interface{}{
						"type":       "base64",
						"media_type": mediaType,
						"data":       base64Data,
					},
				})
			} else if img.URL != "" {
				// URL-based image
				contentParts = append(contentParts, map[string]interface{}{
					"type": "image",
					"source": map[string]interface{}{
						"type": "url",
						"url":  img.URL,
					},
				})
			}
		}

		// Add text content
		if prompt.User != "" {
			contentParts = append(contentParts, map[string]interface{}{
				"type": "text",
				"text": prompt.User,
			})
		}

		content = contentParts
	} else {
		// Text-only content
		content = prompt.User
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": content,
	})

	return systemPrompt, messages
}

// buildTools converts ToolDefinition to Anthropic's tool format
func (a *AnthropicAdapter) buildTools(tools []ToolDefinition) []map[string]interface{} {
	if len(tools) == 0 {
		return nil
	}

	anthropicTools := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		anthropicTool := map[string]interface{}{
			"name":        tool.Function.Name,
			"description": tool.Function.Description,
		}
		if tool.Function.Parameters != nil {
			anthropicTool["input_schema"] = tool.Function.Parameters
		} else {
			// Anthropic requires input_schema, even if empty
			anthropicTool["input_schema"] = map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			}
		}
		anthropicTools = append(anthropicTools, anthropicTool)
	}
	return anthropicTools
}

// Call implements the ModelProvider interface for a single request/response.
func (a *AnthropicAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	// Create observability span
	tracer := otel.Tracer("agenticgokit.llm")
	ctx, span := tracer.Start(ctx, "llm.anthropic.call")
	defer span.End()

	// Set span attributes
	span.SetAttributes(
		attribute.String(observability.AttrLLMProvider, "anthropic"),
		attribute.String(observability.AttrLLMModel, a.model),
		attribute.Int(observability.AttrLLMMaxTokens, a.maxTokens),
		attribute.Float64(observability.AttrLLMTemperature, float64(a.temperature)),
	)

	// Track start time for latency
	startTime := time.Now()

	if prompt.User == "" {
		err := errors.New("user prompt cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "empty user prompt")
		return Response{}, err
	}

	maxTokens := a.maxTokens
	if prompt.Parameters.MaxTokens != nil {
		maxTokens = int(*prompt.Parameters.MaxTokens)
		span.SetAttributes(attribute.Int(observability.AttrLLMMaxTokens, maxTokens))
	}
	temperature := a.temperature
	if prompt.Parameters.Temperature != nil {
		temperature = *prompt.Parameters.Temperature
		span.SetAttributes(attribute.Float64(observability.AttrLLMTemperature, float64(temperature)))
	}

	// Build messages
	systemPrompt, messages := a.buildMessages(prompt)

	// Build request body
	reqBody := map[string]interface{}{
		"model":       a.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		reqBody["system"] = systemPrompt
	}

	// Add optional parameters
	if a.topP > 0 {
		reqBody["top_p"] = a.topP
	}
	if a.topK > 0 {
		reqBody["top_k"] = a.topK
	}
	if len(a.stop) > 0 {
		reqBody["stop_sequences"] = a.stop
	}

	// Add tools if provided
	if tools := a.buildTools(prompt.Tools); tools != nil {
		reqBody["tools"] = tools
		span.SetAttributes(attribute.Int("llm.tool_count", len(tools)))
	}

	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal request")
		return Response{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.getBaseURL()+"/v1/messages", bytes.NewBuffer(requestBody))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create HTTP request")
		return Response{}, err
	}
	a.setHeaders(req)

	resp, err := a.httpClient.Do(req)
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
		err := fmt.Errorf("Anthropic API error: %s", string(body))
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("API error: status %d", resp.StatusCode))
		return Response{}, err
	}

	// Parse response
	var response anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode response")
		return Response{}, err
	}

	// Extract content and tool calls
	var contentBuilder strings.Builder
	var toolCalls []ToolCallResponse

	for _, block := range response.Content {
		switch block.Type {
		case "text":
			contentBuilder.WriteString(block.Text)
		case "tool_use":
			toolCalls = append(toolCalls, ToolCallResponse{
				ID:   block.ID,
				Type: "function",
				Function: FunctionCallResponse{
					Name:      block.Name,
					Arguments: block.Input,
				},
			})
		}
	}

	// Calculate latency
	latencyMs := time.Since(startTime).Milliseconds()

	// Record token usage and latency
	span.SetAttributes(
		attribute.Int(observability.AttrLLMPromptTokens, response.Usage.InputTokens),
		attribute.Int(observability.AttrLLMCompletionTokens, response.Usage.OutputTokens),
		attribute.Int(observability.AttrLLMTotalTokens, response.Usage.InputTokens+response.Usage.OutputTokens),
		attribute.Int64("llm.latency_ms", latencyMs),
		attribute.String("llm.finish_reason", response.StopReason),
	)

	span.SetStatus(codes.Ok, "LLM call successful")

	return Response{
		Content: contentBuilder.String(),
		Usage: UsageStats{
			PromptTokens:     response.Usage.InputTokens,
			CompletionTokens: response.Usage.OutputTokens,
			TotalTokens:      response.Usage.InputTokens + response.Usage.OutputTokens,
		},
		FinishReason: response.StopReason,
		ToolCalls:    toolCalls,
	}, nil
}

// anthropicResponse represents the Anthropic API response structure
type anthropicResponse struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Content      []anthropicContentBlock `json:"content"`
	Model        string                  `json:"model"`
	StopReason   string                  `json:"stop_reason"`
	StopSequence *string                 `json:"stop_sequence,omitempty"`
	Usage        anthropicUsage          `json:"usage"`
}

type anthropicContentBlock struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Stream implements the ModelProvider interface for streaming responses.
func (a *AnthropicAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	// Create observability span
	tracer := otel.Tracer("agenticgokit.llm")
	ctx, span := tracer.Start(ctx, "llm.anthropic.stream")
	defer span.End()

	// Set span attributes
	span.SetAttributes(
		attribute.String(observability.AttrLLMProvider, "anthropic"),
		attribute.String(observability.AttrLLMModel, a.model),
		attribute.Int(observability.AttrLLMMaxTokens, a.maxTokens),
		attribute.Float64(observability.AttrLLMTemperature, float64(a.temperature)),
		attribute.Bool("llm.streaming", true),
	)

	// Track start time for latency
	startTime := time.Now()

	if prompt.User == "" {
		err := errors.New("user prompt cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "empty user prompt")
		return nil, err
	}

	maxTokens := a.maxTokens
	if prompt.Parameters.MaxTokens != nil {
		maxTokens = int(*prompt.Parameters.MaxTokens)
	}
	temperature := a.temperature
	if prompt.Parameters.Temperature != nil {
		temperature = *prompt.Parameters.Temperature
	}

	// Build messages
	systemPrompt, messages := a.buildMessages(prompt)

	// Build request body with streaming enabled
	reqBody := map[string]interface{}{
		"model":       a.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
		"stream":      true,
	}

	if systemPrompt != "" {
		reqBody["system"] = systemPrompt
	}

	if a.topP > 0 {
		reqBody["top_p"] = a.topP
	}
	if a.topK > 0 {
		reqBody["top_k"] = a.topK
	}
	if len(a.stop) > 0 {
		reqBody["stop_sequences"] = a.stop
	}

	// Add tools if provided
	if tools := a.buildTools(prompt.Tools); tools != nil {
		reqBody["tools"] = tools
	}

	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal request")
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.getBaseURL()+"/v1/messages", bytes.NewBuffer(requestBody))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	a.setHeaders(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "HTTP request failed")
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("Anthropic API error: %d - %s", resp.StatusCode, string(body))
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

		chunkCount := 0
		totalBytes := 0

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
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

			// Process SSE event lines
			if strings.HasPrefix(line, "event: ") {
				// Skip event type line, we process data lines
				continue
			}

			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
				if data == "[DONE]" {
					latencyMs := time.Since(startTime).Milliseconds()
					span.SetAttributes(
						attribute.Int("llm.stream.chunk_count", chunkCount),
						attribute.Int("llm.stream.total_bytes", totalBytes),
						attribute.Int64("llm.latency_ms", latencyMs),
					)
					span.SetStatus(codes.Ok, "stream completed successfully")
					return
				}

				// Parse the JSON event
				var event anthropicStreamEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					// Skip malformed events
					continue
				}

				// Handle different event types
				switch event.Type {
				case "content_block_delta":
					if event.Delta != nil && event.Delta.Type == "text_delta" && event.Delta.Text != "" {
						chunkCount++
						totalBytes += len(event.Delta.Text)
						select {
						case tokenChan <- Token{Content: event.Delta.Text}:
						case <-ctx.Done():
							tokenChan <- Token{Error: ctx.Err()}
							span.RecordError(ctx.Err())
							span.SetStatus(codes.Error, "context canceled during send")
							return
						}
					}
				case "message_stop":
					latencyMs := time.Since(startTime).Milliseconds()
					span.SetAttributes(
						attribute.Int("llm.stream.chunk_count", chunkCount),
						attribute.Int("llm.stream.total_bytes", totalBytes),
						attribute.Int64("llm.latency_ms", latencyMs),
					)
					span.SetStatus(codes.Ok, "stream completed successfully")
					return
				case "error":
					if event.Error != nil {
						err := fmt.Errorf("Anthropic stream error: %s - %s", event.Error.Type, event.Error.Message)
						tokenChan <- Token{Error: err}
						span.RecordError(err)
						span.SetStatus(codes.Error, "stream error")
						return
					}
				}
			}
		}

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

// anthropicStreamEvent represents events in the Anthropic streaming response
type anthropicStreamEvent struct {
	Type  string                `json:"type"`
	Index int                   `json:"index,omitempty"`
	Delta *anthropicStreamDelta `json:"delta,omitempty"`
	Error *anthropicStreamError `json:"error,omitempty"`
}

type anthropicStreamDelta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type anthropicStreamError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Embeddings implements the ModelProvider interface.
// Note: Anthropic does not provide an embeddings API, so this returns an error.
func (a *AnthropicAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, errors.New("Anthropic does not provide an embeddings API; use a dedicated embedding provider (e.g., OpenAI, Ollama)")
}

// Model returns the model name
func (a *AnthropicAdapter) Model() string {
	return a.model
}

// BaseURL returns the base URL
func (a *AnthropicAdapter) BaseURL() string {
	return a.getBaseURL()
}
