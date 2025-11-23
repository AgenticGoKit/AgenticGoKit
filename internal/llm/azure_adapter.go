package llm

import (
	"bufio" // For SSE streaming
	"bytes" // For request body
	"context"
	"encoding/json" // For JSON handling
	"errors"
	"fmt"
	"io"
	"log"
	"net/http" // For HTTP requests
	"strings"  // For SSE parsing
	"time"     // For HTTP client timeout
)

// --- API Specific Structs ---

const (
	// Define a specific API version to use
	azureAPIVersion = "2024-02-15-preview" // Or choose another appropriate version
)

// Structure for chat messages sent to the API
type azureChatMessage struct {
	Role    string      `json:"role"`              // "system", "user", "assistant"
	Content interface{} `json:"content,omitempty"` // Text content or multimodal content array
}

// Request structure for the Chat Completions API
type azureChatCompletionsRequest struct {
	Messages    []azureChatMessage `json:"messages"`
	Stream      bool               `json:"stream,omitempty"`
	Temperature *float32           `json:"temperature,omitempty"`
	MaxTokens   *int32             `json:"max_tokens,omitempty"`
	// TODO: Add other parameters like top_p, stop, presence_penalty etc.
}

// Response structure for non-streaming Chat Completions API
type azureChatCompletionsResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Index        int              `json:"index"`
		Message      azureChatMessage `json:"message"`
		FinishReason string           `json:"finish_reason"` // e.g., "stop", "length"
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	// TODO: Add fields for content filtering results if needed
}

// Response structure for streamed Chat Completions API chunks (SSE data)
type azureChatCompletionsStreamChunk struct {
	ID      string `json:"id"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`    // Usually only present in the first chunk for assistant
			Content string `json:"content,omitempty"` // The token delta
		} `json:"delta"`
		FinishReason *string `json:"finish_reason,omitempty"` // Present in the last chunk for a choice
	} `json:"choices"`
	Usage *struct { // Usually nil until the very end, sometimes in a separate final chunk
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// Request structure for the Embeddings API
type azureEmbeddingsRequest struct {
	Input []string `json:"input"`
	// Model string `json:"model"` // Deployment name is used in URL for Azure
}

// Response structure for the Embeddings API
type azureEmbeddingsResponse struct {
	Data []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"` // Note: API returns float32
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// Structure for API errors
type azureErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// --- Adapter Implementation ---

// AzureOpenAIAdapter implements the ModelProvider interface using direct HTTP calls.
type AzureOpenAIAdapter struct {
	httpClient          *http.Client // Use standard HTTP client
	endpointBaseURL     string       // Base URL like "https://your-resource.openai.azure.com"
	apiKey              string
	chatDeployment      string // Deployment name for chat models
	embeddingDeployment string // Deployment name for embedding models
}

// AzureOpenAIAdapterOptions holds configuration options for the AzureOpenAIAdapter.
type AzureOpenAIAdapterOptions struct {
	Endpoint            string       // Example: "https://your-resource-name.openai.azure.com"
	APIKey              string       // Your Azure OpenAI API Key
	ChatDeployment      string       // Deployment name for chat models
	EmbeddingDeployment string       // Deployment name for embedding models
	HTTPClient          *http.Client // Optional: Provide a custom client
}

// NewAzureOpenAIAdapter creates a new adapter for Azure OpenAI using direct HTTP calls.
func NewAzureOpenAIAdapter(opts AzureOpenAIAdapterOptions) (*AzureOpenAIAdapter, error) {
	if opts.Endpoint == "" || opts.APIKey == "" || opts.ChatDeployment == "" || opts.EmbeddingDeployment == "" {
		return nil, errors.New("azure adapter requires endpoint, api key, chat deployment, and embedding deployment")
	}

	// Ensure endpoint doesn't have trailing slash for easier URL joining
	endpoint := strings.TrimSuffix(opts.Endpoint, "/")

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second} // Default client with timeout
	}

	return &AzureOpenAIAdapter{
		httpClient:          client,
		endpointBaseURL:     endpoint,
		apiKey:              opts.APIKey,
		chatDeployment:      opts.ChatDeployment,
		embeddingDeployment: opts.EmbeddingDeployment,
	}, nil
}

// Helper to build the full API URL
func (a *AzureOpenAIAdapter) buildURL(deploymentName, pathSegment string) string {
	// Example: https://{endpoint}/openai/deployments/{deployment}/chat/completions?api-version={version}
	return fmt.Sprintf("%s/openai/deployments/%s/%s?api-version=%s",
		a.endpointBaseURL, deploymentName, pathSegment, azureAPIVersion)
}

// Helper to execute HTTP requests
func (a *AzureOpenAIAdapter) doRequest(ctx context.Context, method, url string, requestBody interface{}) (*http.Response, error) {
	var reqBodyBytes []byte
	var err error

	if requestBody != nil {
		reqBodyBytes, err = json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", a.apiKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	// Check for non-success status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		errorBodyBytes, _ := io.ReadAll(resp.Body)
		apiError := azureErrorResponse{}
		if json.Unmarshal(errorBodyBytes, &apiError) == nil && apiError.Error.Message != "" {
			return nil, fmt.Errorf("api request failed: status %d, type %s, code %s, message: %s",
				resp.StatusCode, apiError.Error.Type, apiError.Error.Code, apiError.Error.Message)
		}
		// Fallback error message
		return nil, fmt.Errorf("api request failed: status %d, body: %s", resp.StatusCode, string(errorBodyBytes))
	}

	return resp, nil
}

// mapInternalPrompt maps our internal Prompt to the API's chat message format.
func mapInternalPrompt(prompt Prompt) []azureChatMessage {
	messages := []azureChatMessage{}
	if prompt.System != "" {
		messages = append(messages, azureChatMessage{Role: "system", Content: prompt.System})
	}
	
	// Build user message with potential multimodal content
	if prompt.User != "" || len(prompt.Images) > 0 {
		var userContent interface{}
		
		if len(prompt.Images) > 0 {
			// Build multimodal content array
			contentArray := []map[string]interface{}{}
			
			// Add text content if present
			if prompt.User != "" {
				contentArray = append(contentArray, map[string]interface{}{
					"type": "text",
					"text": prompt.User,
				})
			}
			
			// Add images
			for _, img := range prompt.Images {
				imageContent := map[string]interface{}{
					"type": "image_url",
				}
				
				if img.URL != "" {
					imageContent["image_url"] = map[string]interface{}{
						"url": img.URL,
					}
				} else if img.Base64 != "" {
					// Azure OpenAI supports base64 images in data URL format
					dataURL := img.Base64
					if !strings.HasPrefix(dataURL, "data:") {
						dataURL = "data:image/jpeg;base64," + dataURL
					}
					imageContent["image_url"] = map[string]interface{}{
						"url": dataURL,
					}
				}
				
				contentArray = append(contentArray, imageContent)
			}
			
			userContent = contentArray
		} else {
			// Text-only content
			userContent = prompt.User
		}
		
		messages = append(messages, azureChatMessage{Role: "user", Content: userContent})
	}
	
	return messages
}

// Call implements the ModelProvider interface for a single request/response.
func (a *AzureOpenAIAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	apiReq := azureChatCompletionsRequest{
		Messages:    mapInternalPrompt(prompt),
		Stream:      false,
		Temperature: prompt.Parameters.Temperature,
		MaxTokens:   prompt.Parameters.MaxTokens,
	}

	url := a.buildURL(a.chatDeployment, "chat/completions")
	httpResp, err := a.doRequest(ctx, http.MethodPost, url, apiReq)
	if err != nil {
		return Response{}, err // Error already formatted by doRequest
	}
	defer httpResp.Body.Close()

	var apiResp azureChatCompletionsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		return Response{}, fmt.Errorf("failed to decode api response: %w", err)
	}

	// Map API response back to internal response
	llmResp := Response{
		Usage: UsageStats{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		},
	}

	if len(apiResp.Choices) > 0 {
		// Handle Content as interface{} - it should be a string for text responses
		if contentStr, ok := apiResp.Choices[0].Message.Content.(string); ok {
			llmResp.Content = contentStr
		}
		llmResp.FinishReason = apiResp.Choices[0].FinishReason
	} else {
		// This case should ideally be covered by non-2xx status code, but check just in case
		return Response{}, errors.New("api returned success but no choices")
	}

	return llmResp, nil
}

// Stream implements the ModelProvider interface for streaming responses.
func (a *AzureOpenAIAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	apiReq := azureChatCompletionsRequest{
		Messages:    mapInternalPrompt(prompt),
		Stream:      true, // Enable streaming
		Temperature: prompt.Parameters.Temperature,
		MaxTokens:   prompt.Parameters.MaxTokens,
	}

	url := a.buildURL(a.chatDeployment, "chat/completions")
	httpResp, err := a.doRequest(ctx, http.MethodPost, url, apiReq)
	if err != nil {
		// If error occurs before stream starts, return error directly
		return nil, err
	}

	tokenChan := make(chan Token)

	// Start goroutine to process the SSE stream
	go func() {
		defer close(tokenChan)
		defer httpResp.Body.Close()

		scanner := bufio.NewScanner(httpResp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue // Skip empty lines
			}

			// Check for context cancellation during processing
			select {
			case <-ctx.Done():
				log.Printf("Azure stream context cancelled during SSE processing")
				return
			default:
				// continue processing
			}

			if strings.HasPrefix(line, "data:") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				if data == "[DONE]" {
					return // Stream finished successfully
				}

				var chunk azureChatCompletionsStreamChunk
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					log.Printf("Error decoding stream chunk JSON: %v, data: %s", err, data)
					tokenChan <- Token{Error: fmt.Errorf("stream decode error: %w", err)}
					return
				}

				// Extract content delta
				if len(chunk.Choices) > 0 {
					contentDelta := chunk.Choices[0].Delta.Content
					if contentDelta != "" {
						select {
						case tokenChan <- Token{Content: contentDelta}:
						case <-ctx.Done():
							log.Printf("Azure stream context cancelled during token send")
							return
						}
					}
					// TODO: Could potentially capture finish_reason from the last chunk here
				}
			} else {
				log.Printf("Unexpected SSE line: %s", line)
			}
		}

		// Check for scanner errors after loop
		if err := scanner.Err(); err != nil {
			// Don't send if context was cancelled, as that's the likely cause
			if ctx.Err() == nil {
				log.Printf("Error reading stream body: %v", err)
				tokenChan <- Token{Error: fmt.Errorf("stream read error: %w", err)}
			}
		}
	}()

	return tokenChan, nil
}

// Embeddings implements the ModelProvider interface for generating embeddings.
func (a *AzureOpenAIAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	apiReq := azureEmbeddingsRequest{
		Input: texts,
	}

	url := a.buildURL(a.embeddingDeployment, "embeddings")
	httpResp, err := a.doRequest(ctx, http.MethodPost, url, apiReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var apiResp azureEmbeddingsResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode embeddings response: %w", err)
	}

	if len(apiResp.Data) != len(texts) {
		return nil, fmt.Errorf("embeddings api returned %d embeddings, expected %d", len(apiResp.Data), len(texts))
	}

	// Convert float32 embeddings to float64
	embeddings := make([][]float64, len(apiResp.Data))
	// Assuming the order is preserved by the API
	for _, item := range apiResp.Data {
		if item.Index < 0 || item.Index >= len(embeddings) {
			return nil, fmt.Errorf("embeddings api returned invalid index %d", item.Index)
		}
		float64Embedding := make([]float64, len(item.Embedding))
		for j, val := range item.Embedding {
			float64Embedding[j] = float64(val)
		}
		embeddings[item.Index] = float64Embedding // Place using index
	}

	// Verify all embeddings were received
	for i, emb := range embeddings {
		if emb == nil {
			return nil, fmt.Errorf("embeddings api did not return embedding for index %d", i)
		}
	}

	// TODO: Return usage info if needed (apiResp.Usage)

	return embeddings, nil
}
