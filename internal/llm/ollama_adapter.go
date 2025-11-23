package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaAdapter implements the LLMAdapter interface for Ollama's API.
type OllamaAdapter struct {
	baseURL        string
	model          string
	embeddingModel string
	maxTokens      int
	temperature    float32
}

// NewOllamaAdapter creates a new OllamaAdapter instance.
// baseURL should include scheme and host, e.g. http://localhost:11434
func NewOllamaAdapter(baseURL, model string, maxTokens int, temperature float32) (*OllamaAdapter, error) {
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

	return &OllamaAdapter{
		baseURL:        baseURL,
		model:          model,
		embeddingModel: "nomic-embed-text:latest", // Default embedding model
		maxTokens:      maxTokens,
		temperature:    temperature,
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
	if prompt.System == "" && prompt.User == "" {
		return Response{}, errors.New("both system and user prompts cannot be empty")
	}

	// Determine final parameters, preferring explicit prompt settings
	var finalMaxTokens int
	if prompt.Parameters.MaxTokens != nil && *prompt.Parameters.MaxTokens > 0 {
		finalMaxTokens = int(*prompt.Parameters.MaxTokens)
	} else {
		finalMaxTokens = o.maxTokens
	}

	var finalTemperature float32
	if prompt.Parameters.Temperature != nil && *prompt.Parameters.Temperature > 0 {
		finalTemperature = *prompt.Parameters.Temperature
	} else {
		finalTemperature = o.temperature
	}

	// Build messages array
	messages := []map[string]interface{}{}
	if prompt.System != "" {
		messages = append(messages, map[string]interface{}{"role": "system", "content": prompt.System})
	}
	
	userMessage := map[string]interface{}{"role": "user", "content": prompt.User}
	
	// Add images if present
	if len(prompt.Images) > 0 {
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

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)

				if err == nil && resp.StatusCode == http.StatusOK {
					defer resp.Body.Close()
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
	
	messages = append(messages, userMessage)

	// Prepare the request payload
	requestBody := map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  finalMaxTokens,
		"temperature": finalTemperature,
		"stream":      false,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return Response{}, fmt.Errorf("failed to marshal request body: %w", err)
	}
	// Make the HTTP request with timeout
	client := &http.Client{
		Timeout: 30 * time.Second, // Add 30 second timeout
	}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/chat", o.baseURL), bytes.NewBuffer(payload))
	if err != nil {
		return Response{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Response{}, fmt.Errorf("Ollama API error: %s", string(body))
	}

	var apiResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return Response{
		Content: apiResp.Message.Content,
	}, nil
}

// Stream implements the ModelProvider interface for streaming responses.
func (o *OllamaAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
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

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)

				if err == nil && resp.StatusCode == http.StatusOK {
					defer resp.Body.Close()
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

	// Apply prompt parameters if provided
	if prompt.Parameters.Temperature != nil {
		payload["options"].(map[string]interface{})["temperature"] = *prompt.Parameters.Temperature
	}
	if prompt.Parameters.MaxTokens != nil {
		payload["options"].(map[string]interface{})["num_predict"] = *prompt.Parameters.MaxTokens
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP request for streaming
	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Create token channel
	tokenChan := make(chan Token, 10)

	// Start goroutine to process streaming response
	go func() {
		defer close(tokenChan)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)

		for {
			select {
			case <-ctx.Done():
				tokenChan <- Token{Error: ctx.Err()}
				return
			default:
			}

			var response struct {
				Response string `json:"response"`
				Done     bool   `json:"done"`
				Error    string `json:"error,omitempty"`
			}

			if err := decoder.Decode(&response); err != nil {
				if err == io.EOF {
					return // End of stream
				}
				tokenChan <- Token{Error: fmt.Errorf("failed to decode response: %w", err)}
				return
			}

			if response.Error != "" {
				tokenChan <- Token{Error: fmt.Errorf("ollama error: %s", response.Error)}
				return
			}

			if response.Response != "" {
				tokenChan <- Token{Content: response.Response}
			}

			if response.Done {
				return // Stream complete
			}
		}
	}()

	return tokenChan, nil
}

// Embeddings implements the ModelProvider interface for generating embeddings.
func (o *OllamaAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
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

		// Make the HTTP request with timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/embeddings", o.baseURL), bytes.NewBuffer(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP request for text %d: %w", i, err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
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
