package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/agenticgokit/agenticgokit/internal/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// OllamaAdapter implements the LLMAdapter interface for Ollama's API.
type OllamaAdapter struct {
	baseURL        string
	model          string
	embeddingModel string
	maxTokens      int
	temperature    float32
	httpClient     *http.Client
}

// NewOllamaAdapter creates a new OllamaAdapter instance.
// baseURL should include scheme and host, e.g. http://localhost:11434
func NewOllamaAdapter(baseURL, model string, maxTokens int, temperature float32, httpTimeout time.Duration) (*OllamaAdapter, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2:latest" // Use llama3.2 as default - good general purpose model
	}
	if maxTokens == 0 {
		maxTokens = 150 // Default max tokens
	}
	if temperature == 0 {
		temperature = 0.7 // Default temperature
	}
	if httpTimeout == 0 {
		httpTimeout = 120 * time.Second // Default timeout
	}

	// Reuse one HTTP client with keep-alive to avoid connection churn and model reload latency
	// Use optimized transport configuration for best performance
	client := NewOptimizedHTTPClient(httpTimeout)

	return &OllamaAdapter{
		baseURL:        baseURL,
		model:          model,
		embeddingModel: "nomic-embed-text:latest", // Default embedding model
		maxTokens:      maxTokens,
		temperature:    temperature,
		httpClient:     client,
	}, nil
}

// SetEmbeddingModel allows setting a custom embedding model for the adapter
func (o *OllamaAdapter) SetEmbeddingModel(model string) {
	if model != "" {
		o.embeddingModel = model
	}
}

// Call implements the ModelProvider interface for a single request/response.
func (o *OllamaAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	// Create observability span
	tracer := otel.Tracer("agenticgokit.llm")
	ctx, span := tracer.Start(ctx, "llm.ollama.call")
	defer span.End()

	// Ensure HTTP client is initialized for tests that construct adapter directly
	if o.httpClient == nil {
		o.httpClient = NewOptimizedHTTPClient(120 * time.Second)
	}

	if prompt.System == "" && prompt.User == "" {
		err := errors.New("both system and user prompts cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "empty prompts")
		return Response{}, err
	}

	// Determine final parameters, preferring explicit prompt settings
	var finalMaxTokens int
	if prompt.Parameters.MaxTokens != nil && *prompt.Parameters.MaxTokens > 0 {
		finalMaxTokens = int(*prompt.Parameters.MaxTokens)
	} else {
		finalMaxTokens = o.maxTokens
	}

	var finalTemperature float32
	// A non-nil pointer is an explicit setting: temperature 0 (deterministic
	// sampling) is valid and must not fall back to the adapter default.
	if prompt.Parameters.Temperature != nil {
		finalTemperature = *prompt.Parameters.Temperature
	} else {
		finalTemperature = o.temperature
	}

	// Set span attributes
	span.SetAttributes(
		attribute.String(observability.AttrLLMProvider, "ollama"),
		attribute.String(observability.AttrLLMModel, o.model),
		attribute.Int(observability.AttrLLMMaxTokens, finalMaxTokens),
		attribute.Float64(observability.AttrLLMTemperature, float64(finalTemperature)),
	)

	// Track start time for latency
	startTime := time.Now()

	// Build messages array
	messages := []map[string]interface{}{}
	if prompt.System != "" {
		messages = append(messages, map[string]interface{}{"role": "system", "content": prompt.System})
	}

	userMessage := map[string]interface{}{"role": "user", "content": prompt.User}

	// Add images if present
	if len(prompt.Images) > 0 {
		span.SetAttributes(
			attribute.Bool("llm.multimodal", true),
			attribute.Int("llm.image_count", len(prompt.Images)),
		)
		images := []string{}
		for _, img := range prompt.Images {
			// Ollama expects base64 strings
			if img.Base64 != "" {
				// Strip prefix if present (e.g. data:image/jpeg;base64,)
				base64Data := img.Base64
				if idx := strings.Index(base64Data, ","); idx != -1 {
					base64Data = base64Data[idx+1:]
				}
				images = append(images, base64Data)
			} else if img.URL != "" {
				// Fetch URL and convert to base64
				req, err := http.NewRequestWithContext(ctx, "GET", img.URL, nil)
				if err != nil {
					continue
				}
				req.Header.Set("User-Agent", "AgenticGoKit/1.0")

				// PERFORMANCE: Reuse adapter's httpClient for connection pooling
				resp, err := o.httpClient.Do(req)
				if err != nil {
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					data, err := io.ReadAll(resp.Body)
					if err == nil {
						base64Data := base64.StdEncoding.EncodeToString(data)
						images = append(images, base64Data)
					}
				}
			}
		}
		if len(images) > 0 {
			userMessage["images"] = images
		}
	}

	// Log warning for unsupported Audio/Video inputs
	if len(prompt.Audio) > 0 {
		log.Printf("WARN: Ollama adapter does not currently support Audio inputs. Ignoring %d audio files.\n", len(prompt.Audio))
	}
	if len(prompt.Video) > 0 {
		log.Printf("WARN: Ollama adapter does not currently support Video inputs. Ignoring %d video files.\n", len(prompt.Video))
	}

	messages = append(messages, userMessage)

	// Prepare the request payload
	requestBody := map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  finalMaxTokens,
		"temperature": finalTemperature,
		"stream":      false,
	}

	// Include native tools if provided
	if len(prompt.Tools) > 0 {
		span.SetAttributes(attribute.Int("llm.tool_count", len(prompt.Tools)))
		tools := make([]map[string]interface{}, len(prompt.Tools))
		for i, tool := range prompt.Tools {
			tools[i] = map[string]interface{}{
				"type":     tool.Type,
				"function": tool.Function,
			}
		}
		requestBody["tools"] = tools
		// Hint the model to choose tools automatically when appropriate
		requestBody["tool_choice"] = "auto"
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal request")
		return Response{}, fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/chat", o.baseURL), bytes.NewBuffer(payload))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create HTTP request")
		return Response{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "HTTP request failed")
		return Response{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Record HTTP status
	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("Ollama API error: %s", string(body))
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("API error: status %d", resp.StatusCode))
		return Response{}, err
	}

	var apiResp struct {
		Message struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Function struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"message"`
		// Ollama returns token accounting on completion
		PromptEvalCount int   `json:"prompt_eval_count"`
		EvalCount       int   `json:"eval_count"`
		TotalDurationNs int64 `json:"total_duration"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode response")
		return Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// Capture token usage and latency when available
	if apiResp.PromptEvalCount > 0 {
		span.SetAttributes(
			attribute.Int(observability.AttrLLMPromptTokens, apiResp.PromptEvalCount),
			attribute.Int(observability.AttrLLMTokensIn, apiResp.PromptEvalCount),
		)
	}
	if apiResp.EvalCount > 0 {
		span.SetAttributes(
			attribute.Int(observability.AttrLLMCompletionTokens, apiResp.EvalCount),
			attribute.Int(observability.AttrLLMTokensOut, apiResp.EvalCount),
		)
	}
	if apiResp.PromptEvalCount > 0 || apiResp.EvalCount > 0 {
		total := apiResp.PromptEvalCount + apiResp.EvalCount
		span.SetAttributes(attribute.Int(observability.AttrLLMTotalTokens, total))
	}
	// Prefer server-reported duration; fallback to local measurement
	if apiResp.TotalDurationNs > 0 {
		span.SetAttributes(attribute.Int64(observability.AttrLLMLatencyMs, apiResp.TotalDurationNs/1_000_000))
	} else {
		latencyMs := time.Since(startTime).Milliseconds()
		span.SetAttributes(attribute.Int64(observability.AttrLLMLatencyMs, latencyMs))
	}

	response := Response{
		Content: apiResp.Message.Content,
	}

	if len(apiResp.Message.ToolCalls) > 0 {
		span.SetAttributes(attribute.Int("llm.tool_calls_count", len(apiResp.Message.ToolCalls)))
		response.ToolCalls = make([]ToolCallResponse, len(apiResp.Message.ToolCalls))
		for i, tc := range apiResp.Message.ToolCalls {
			response.ToolCalls[i] = ToolCallResponse{
				Type: "function",
				Function: FunctionCallResponse{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	// Capture full prompt and response at detailed trace level
	if observability.IsDetailedTracing() {
		span.SetAttributes(
			attribute.String(observability.AttrPromptSystem, observability.TruncateForTrace(prompt.System, observability.MaxContentLength)),
			attribute.String(observability.AttrPromptUser, observability.TruncateForTrace(prompt.User, observability.MaxContentLength)),
			attribute.String(observability.AttrLLMResponse, observability.TruncateForTrace(response.Content, observability.MaxContentLength)),
		)
	}

	span.SetStatus(codes.Ok, "LLM call successful")
	return response, nil
}

// Stream implements the ModelProvider interface for streaming responses.
func (o *OllamaAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	// Create observability span
	tracer := otel.Tracer("agenticgokit.llm")
	ctx, span := tracer.Start(ctx, "llm.ollama.stream")
	defer span.End()

	// Ensure HTTP client is initialized for tests that construct adapter directly
	if o.httpClient == nil {
		o.httpClient = NewOptimizedHTTPClient(120 * time.Second)
	}

	// Set span attributes
	span.SetAttributes(
		attribute.String(observability.AttrLLMProvider, "ollama"),
		attribute.String(observability.AttrLLMModel, o.model),
		attribute.Int(observability.AttrLLMMaxTokens, o.maxTokens),
		attribute.Float64(observability.AttrLLMTemperature, float64(o.temperature)),
		attribute.Bool("llm.streaming", true),
	)

	// Track start time for latency
	startTime := time.Now()

	// Create the request payload for Ollama streaming API
	payload := map[string]interface{}{
		"model":  o.model,
		"prompt": fmt.Sprintf("%s\n\nHuman: %s\n\nAssistant:", prompt.System, prompt.User),
		"stream": true,
		"options": map[string]interface{}{
			"temperature": o.temperature,
			"num_predict": o.maxTokens,
		},
	}

	// Add images if present
	if len(prompt.Images) > 0 {
		span.SetAttributes(
			attribute.Bool("llm.multimodal", true),
			attribute.Int("llm.image_count", len(prompt.Images)),
		)
		images := []string{}
		for _, img := range prompt.Images {
			if img.Base64 != "" {
				base64Data := img.Base64
				if idx := strings.Index(base64Data, ","); idx != -1 {
					base64Data = base64Data[idx+1:]
				}
				images = append(images, base64Data)
			} else if img.URL != "" {
				// Fetch URL and convert to base64
				req, err := http.NewRequestWithContext(ctx, "GET", img.URL, nil)
				if err != nil {
					continue
				}
				req.Header.Set("User-Agent", "AgenticGoKit/1.0")

				// PERFORMANCE: Reuse adapter's httpClient for connection pooling
				resp, err := o.httpClient.Do(req)
				if err != nil {
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					data, err := io.ReadAll(resp.Body)
					if err == nil {
						base64Data := base64.StdEncoding.EncodeToString(data)
						images = append(images, base64Data)
					}
				}
			}
		}
		if len(images) > 0 {
			payload["images"] = images
		}
	}

	// Log warning for unsupported Audio/Video inputs
	if len(prompt.Audio) > 0 {
		log.Printf("WARN: Ollama adapter does not currently support Audio inputs. Ignoring %d audio files.\n", len(prompt.Audio))
	}
	if len(prompt.Video) > 0 {
		log.Printf("WARN: Ollama adapter does not currently support Video inputs. Ignoring %d video files.\n", len(prompt.Video))
	}

	// Apply prompt parameters if provided
	if prompt.Parameters.Temperature != nil {
		payload["options"].(map[string]interface{})["temperature"] = *prompt.Parameters.Temperature
		span.SetAttributes(attribute.Float64(observability.AttrLLMTemperature, float64(*prompt.Parameters.Temperature)))
	}
	if prompt.Parameters.MaxTokens != nil {
		payload["options"].(map[string]interface{})["num_predict"] = *prompt.Parameters.MaxTokens
		span.SetAttributes(attribute.Int(observability.AttrLLMMaxTokens, int(*prompt.Parameters.MaxTokens)))
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal request")
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP request for streaming
	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewReader(payloadBytes))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
		resp.Body.Close()
		err := fmt.Errorf("HTTP error: %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP error: %d", resp.StatusCode))
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
		promptTokens := 0
		completionTokens := 0
		serverDurationNs := int64(0)
		var fullResponseBuilder strings.Builder

		decoder := json.NewDecoder(resp.Body)

		for {
			select {
			case <-ctx.Done():
				tokenChan <- Token{Error: ctx.Err()}
				span.RecordError(ctx.Err())
				span.SetStatus(codes.Error, "context canceled during stream")
				return
			default:
			}

			var response struct {
				Response        string `json:"response"`
				Done            bool   `json:"done"`
				Error           string `json:"error,omitempty"`
				PromptEvalCount int    `json:"prompt_eval_count"`
				EvalCount       int    `json:"eval_count"`
				TotalDurationNs int64  `json:"total_duration"`
			}

			if err := decoder.Decode(&response); err != nil {
				if err == io.EOF {
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
					return // End of stream
				}
				tokenChan <- Token{Error: fmt.Errorf("failed to decode response: %w", err)}
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to decode stream response")
				return
			}

			if response.Error != "" {
				err := fmt.Errorf("ollama error: %s", response.Error)
				tokenChan <- Token{Error: err}
				span.RecordError(err)
				span.SetStatus(codes.Error, "ollama API error")
				return
			}

			if response.Response != "" {
				chunkCount++
				totalBytes += len(response.Response)
				fullResponseBuilder.WriteString(response.Response)
				tokenChan <- Token{Content: response.Response}
			}

			if response.Done {
				// Capture token accounting if provided
				if response.PromptEvalCount > 0 {
					promptTokens = response.PromptEvalCount
				}
				if response.EvalCount > 0 {
					completionTokens = response.EvalCount
				}
				if response.TotalDurationNs > 0 {
					serverDurationNs = response.TotalDurationNs
				}

				// Record final metrics
				latencyMs := time.Since(startTime).Milliseconds()
				if serverDurationNs > 0 {
					latencyMs = serverDurationNs / 1_000_000
				}
				attrs := []attribute.KeyValue{
					attribute.Int("llm.stream.chunk_count", chunkCount),
					attribute.Int("llm.stream.total_bytes", totalBytes),
					attribute.Int64(observability.AttrLLMLatencyMs, latencyMs),
				}
				if promptTokens > 0 {
					attrs = append(attrs,
						attribute.Int(observability.AttrLLMPromptTokens, promptTokens),
						attribute.Int(observability.AttrLLMTokensIn, promptTokens),
					)
				}
				if completionTokens > 0 {
					attrs = append(attrs,
						attribute.Int(observability.AttrLLMCompletionTokens, completionTokens),
						attribute.Int(observability.AttrLLMTokensOut, completionTokens),
					)
				}
				if promptTokens+completionTokens > 0 {
					total := promptTokens + completionTokens
					attrs = append(attrs, attribute.Int(observability.AttrLLMTotalTokens, total))
				}
				span.SetAttributes(attrs...)
				if observability.IsDetailedTracing() {
					span.SetAttributes(
						attribute.String(observability.AttrPromptSystem, observability.TruncateForTrace(prompt.System, observability.MaxContentLength)),
						attribute.String(observability.AttrPromptUser, observability.TruncateForTrace(prompt.User, observability.MaxContentLength)),
						attribute.String(observability.AttrLLMResponse, observability.TruncateForTrace(fullResponseBuilder.String(), observability.MaxContentLength)),
					)
				}
				span.SetStatus(codes.Ok, "stream completed successfully")
				return // Stream complete
			}
		}
	}()

	return tokenChan, nil
}

// Embeddings implements the ModelProvider interface for generating embeddings.
func (o *OllamaAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	// Ensure HTTP client is initialized for tests that construct adapter directly
	if o.httpClient == nil {
		o.httpClient = NewOptimizedHTTPClient(120 * time.Second)
	}
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	// Use the configured embedding model
	// Common embedding models in Ollama: nomic-embed-text, all-minilm, mxbai-embed-large
	embeddingModel := o.embeddingModel

	embeddings := make([][]float64, len(texts))

	// Process each text individually as Ollama embeddings API typically handles one at a time
	for i, text := range texts {
		requestBody := map[string]interface{}{
			"model":  embeddingModel,
			"prompt": text,
		}

		payload, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body for text %d: %w", i, err)
		}

		// Make the HTTP request using the reusable httpClient
		// PERFORMANCE FIX: Reuse adapter's httpClient instead of creating new one per request
		// This enables connection pooling and keep-alive for embeddings
		req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/embeddings", o.baseURL), bytes.NewBuffer(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP request for text %d: %w", i, err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := o.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("HTTP request failed for text %d: %w", i, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("Ollama embeddings API error for text %d: %s", i, string(body))
		}

		var apiResp struct {
			Embedding []float64 `json:"embedding"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			return nil, fmt.Errorf("failed to decode embeddings response for text %d: %w", i, err)
		}

		if len(apiResp.Embedding) == 0 {
			return nil, fmt.Errorf("empty embedding returned for text %d", i)
		}

		embeddings[i] = apiResp.Embedding
	}

	return embeddings, nil
}
