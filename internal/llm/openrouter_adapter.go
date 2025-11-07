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
		httpClient:  &http.Client{Timeout: 120 * time.Second},
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

	// Add user message if provided
	if prompt.User != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "user",
			"content": prompt.User,
		})
	}

	return messages
}

// Call implements the ModelProvider interface for a single request/response.
func (o *OpenRouterAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	userPrompt := prompt.User
	if userPrompt == "" {
		return Response{}, errors.New("user prompt cannot be empty")
	}

	maxTokens := o.maxTokens
	if prompt.Parameters.MaxTokens != nil {
		maxTokens = int(*prompt.Parameters.MaxTokens)
	}
	temperature := o.temperature
	if prompt.Parameters.Temperature != nil {
		temperature = *prompt.Parameters.Temperature
	}

	// Build messages array for Chat Completions API
	messages := buildOpenRouterMessages(prompt)

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
	})
	if err != nil {
		return Response{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
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
		return Response{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

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
			return Response{}, fmt.Errorf("OpenRouter API error [%s]: %s", errorResp.Error.Code, errorResp.Error.Message)
		}

		return Response{}, fmt.Errorf("OpenRouter API error: %d - %s", resp.StatusCode, string(body))
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
		return Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return Response{}, errors.New("no completion choices returned")
	}

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
	userPrompt := prompt.User
	if userPrompt == "" {
		return nil, errors.New("user prompt cannot be empty")
	}

	maxTokens := o.maxTokens
	if prompt.Parameters.MaxTokens != nil {
		maxTokens = int(*prompt.Parameters.MaxTokens)
	}
	temperature := o.temperature
	if prompt.Parameters.Temperature != nil {
		temperature = *prompt.Parameters.Temperature
	}

	// Build messages array for Chat Completions API
	messages := buildOpenRouterMessages(prompt)

	// Create streaming request
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
		"stream":      true, // Enable streaming
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP request for streaming
	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
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
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
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
			return nil, fmt.Errorf("OpenRouter API error [%s]: %s", errorResp.Error.Code, errorResp.Error.Message)
		}

		return nil, fmt.Errorf("OpenRouter API error: %d - %s", resp.StatusCode, string(body))
	}

	// Create token channel
	tokenChan := make(chan Token, 10)

	// Start goroutine to process streaming response
	go func() {
		defer close(tokenChan)
		defer resp.Body.Close()

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
				return
			default:
			}

			// Process SSE data lines
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
				if data == "[DONE]" {
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
					return
				}

				// Extract content delta
				if len(streamResponse.Choices) > 0 {
					content := streamResponse.Choices[0].Delta.Content
					if content != "" {
						select {
						case tokenChan <- Token{Content: content}:
						case <-ctx.Done():
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
