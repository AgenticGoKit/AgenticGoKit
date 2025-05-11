package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// OllamaAdapter implements the LLMAdapter interface for Ollama's API.
type OllamaAdapter struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float32
}

// NewOllamaAdapter creates a new OllamaAdapter instance.
func NewOllamaAdapter(apiKey, model string, maxTokens int, temperature float32) (*OllamaAdapter, error) {
	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
	}
	if model == "" {
		model = "gemma3:latest" // Replace with Ollama's default model
	}
	if maxTokens == 0 {
		maxTokens = 150 // Default max tokens
	}
	if temperature == 0 {
		temperature = 0.7 // Default temperature
	}

	return &OllamaAdapter{
		apiKey:      apiKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
	}, nil
}

// Call implements the ModelProvider interface for a single request/response.
func (o *OllamaAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	if prompt.System == "" && prompt.User == "" {
		return Response{}, errors.New("both system and user prompts cannot be empty")
	}

	// Prepare the request payload
	requestBody := map[string]interface{}{
		"model":       o.model,
		"messages":    []map[string]string{{"role": "system", "content": prompt.System}, {"role": "user", "content": prompt.User}},
		"max_tokens":  prompt.Parameters.MaxTokens,
		"temperature": prompt.Parameters.Temperature,
		"stream":      false,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return Response{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Make the HTTP request
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:11434/api/chat", bytes.NewBuffer(payload))
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
		body, _ := ioutil.ReadAll(resp.Body)
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
	return nil, errors.New("streaming not supported by OllamaAdapter")
}

// Embeddings implements the ModelProvider interface for generating embeddings.
func (o *OllamaAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, errors.New("embeddings not supported by OllamaAdapter")
}
