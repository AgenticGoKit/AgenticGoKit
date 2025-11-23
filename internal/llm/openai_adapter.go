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

// OpenAIAdapter implements the ModelProvider interface for OpenAI's API.
type OpenAIAdapter struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float32
}

// NewOpenAIAdapter creates a new OpenAIAdapter instance.
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
	}, nil
}

// Call implements the ModelProvider interface for a single request/response.
func (o *OpenAIAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
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
		contentParts := []map[string]interface{}{
			{
				"type": "text",
				"text": userPrompt,
			},
		}

		for _, img := range prompt.Images {
			imgObj := map[string]interface{}{
				"type": "image_url",
				"image_url": map[string]string{
					"url": img.URL,
				},
			}
			// If Base64 is provided and URL is empty, construct data URL
			if img.URL == "" && img.Base64 != "" {
				// Assume jpeg if not specified, or try to detect? 
				// For now, let's assume the Base64 string might already contain the prefix or we add a generic one.
				// Standard format: data:image/jpeg;base64,{base64_image}
				if !strings.HasPrefix(img.Base64, "data:") {
					imgObj["image_url"] = map[string]string{
						"url": fmt.Sprintf("data:image/jpeg;base64,%s", img.Base64),
					}
				} else {
					imgObj["image_url"] = map[string]string{
						"url": img.Base64,
					}
				}
			}
			contentParts = append(contentParts, imgObj)
		}
		userContent = contentParts
	} else {
		// Text-only content
		userContent = userPrompt
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userContent,
	})

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
	})
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Response{}, errors.New("OpenAI API error: " + string(body))
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
		return Response{}, err
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
func (o *OpenAIAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
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
		contentParts := []map[string]interface{}{
			{
				"type": "text",
				"text": userPrompt,
			},
		}

		for _, img := range prompt.Images {
			imgObj := map[string]interface{}{
				"type": "image_url",
				"image_url": map[string]string{
					"url": img.URL,
				},
			}
			// If Base64 is provided and URL is empty, construct data URL
			if img.URL == "" && img.Base64 != "" {
				if !strings.HasPrefix(img.Base64, "data:") {
					imgObj["image_url"] = map[string]string{
						"url": fmt.Sprintf("data:image/jpeg;base64,%s", img.Base64),
					}
				} else {
					imgObj["image_url"] = map[string]string{
						"url": img.Base64,
					}
				}
			}
			contentParts = append(contentParts, imgObj)
		}
		userContent = contentParts
	} else {
		// Text-only content
		userContent = userPrompt
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userContent,
	})

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
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	// Make the request
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
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
func (o *OpenAIAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	// Use appropriate embedding model instead of chat model
	embeddingModel := "text-embedding-3-small"

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": embeddingModel,
		"input": texts,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New("OpenAI API error: " + string(body))
	}

	var response struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Data) != len(texts) {
		return nil, errors.New("number of embeddings returned does not match input")
	}

	embeddings := make([][]float64, len(texts))
	for i, item := range response.Data {
		embeddings[i] = item.Embedding
	}

	return embeddings, nil
}
