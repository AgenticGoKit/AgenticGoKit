package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// OpenAIAdapter implements the LLMAdapter interface for OpenAI's API.
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

// Complete sends a prompt to OpenAI's API and returns the completion.
func (o *OpenAIAdapter) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	if systemPrompt == "" && userPrompt == "" {
		return "", errors.New("both systemPrompt and userPrompt cannot be empty")
	}

	prompt := systemPrompt + "\n" + userPrompt
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       o.model,
		"prompt":      prompt,
		"max_tokens":  o.maxTokens,
		"temperature": o.temperature,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", errors.New("OpenAI API error: " + string(body))
	}

	var response struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", errors.New("no completion choices returned")
	}

	return response.Choices[0].Text, nil
}
