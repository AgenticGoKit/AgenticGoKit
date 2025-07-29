package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
	messages := []map[string]interface{}{
		{
			"role":    "user",
			"content": userPrompt,
		},
	}
	
	// Add system message if provided
	if prompt.System != "" {
		messages = append([]map[string]interface{}{
			{
				"role":    "system", 
				"content": prompt.System,
			},
		}, messages...)
	}

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
	ch := make(chan Token)
	go func() {
		defer close(ch)
		// For now, just call Call and send the whole response as one token (streaming can be implemented later with SSE)
		resp, err := o.Call(ctx, prompt)
		if err != nil {
			ch <- Token{Error: err}
			return
		}
		ch <- Token{Content: resp.Content}
	}()
	return ch, nil
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
