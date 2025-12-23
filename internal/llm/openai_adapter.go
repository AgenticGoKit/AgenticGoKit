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

// DefaultOpenAIBaseURL is the default OpenAI API endpoint
const DefaultOpenAIBaseURL = "https://api.openai.com/v1"

// OpenAIAdapterConfig holds extended configuration for OpenAI-compatible adapters
type OpenAIAdapterConfig struct {
	APIKey       string
	Model        string
	MaxTokens    int
	Temperature  float32
	BaseURL      string            // Custom base URL (for vLLM, MLFlow, etc.)
	ExtraHeaders map[string]string // Custom headers (for gateways)
	HTTPTimeout  time.Duration     // HTTP client timeout

	// Extended sampling parameters (for vLLM compatibility)
	TopP              float32
	TopK              int
	PresencePenalty   float32
	FrequencyPenalty  float32
	RepetitionPenalty float32
	Stop              []string
}

// OpenAIAdapter implements the ModelProvider interface for OpenAI-compatible APIs.
// This adapter can be used for OpenAI, vLLM, MLFlow Gateway, and any other
// OpenAI-compatible endpoint by configuring the baseURL.
type OpenAIAdapter struct {
	apiKey       string
	model        string
	maxTokens    int
	temperature  float32
	baseURL      string            // Default: https://api.openai.com/v1
	extraHeaders map[string]string // Custom headers for gateways
	httpClient   *http.Client

	// Extended sampling parameters
	topP              float32
	topK              int
	presencePenalty   float32
	frequencyPenalty  float32
	repetitionPenalty float32
	stop              []string
}

// NewOpenAIAdapter creates a new OpenAIAdapter instance.
// SIGNATURE REMAINS UNCHANGED for backward compatibility.
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
		baseURL:     DefaultOpenAIBaseURL, // Default OpenAI URL
		httpClient:  &http.Client{Timeout: 120 * time.Second},
	}, nil
}

// NewOpenAIAdapterWithConfig creates an OpenAI-compatible adapter with extended configuration.
// Use this for vLLM, MLFlow Gateway, or any OpenAI-compatible endpoint.
func NewOpenAIAdapterWithConfig(config OpenAIAdapterConfig) (*OpenAIAdapter, error) {
	if config.Model == "" {
		return nil, errors.New("model is required")
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultOpenAIBaseURL
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 2048
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 120 * time.Second
	}

	return &OpenAIAdapter{
		apiKey:            config.APIKey,
		model:             config.Model,
		maxTokens:         config.MaxTokens,
		temperature:       config.Temperature,
		baseURL:           strings.TrimSuffix(config.BaseURL, "/"),
		extraHeaders:      config.ExtraHeaders,
		httpClient:        &http.Client{Timeout: config.HTTPTimeout},
		topP:              config.TopP,
		topK:              config.TopK,
		presencePenalty:   config.PresencePenalty,
		frequencyPenalty:  config.FrequencyPenalty,
		repetitionPenalty: config.RepetitionPenalty,
		stop:              config.Stop,
	}, nil
}

// SetBaseURL allows overriding the default OpenAI endpoint (e.g., for vLLM or MLFlow)
func (o *OpenAIAdapter) SetBaseURL(url string) {
	if url != "" {
		o.baseURL = strings.TrimSuffix(url, "/")
	}
}

// SetExtraHeaders allows adding custom headers (e.g., for MLFlow Gateway)
func (o *OpenAIAdapter) SetExtraHeaders(headers map[string]string) {
	o.extraHeaders = headers
}

// setHeaders sets common headers for requests
func (o *OpenAIAdapter) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
	}
	for key, value := range o.extraHeaders {
		req.Header.Set(key, value)
	}
}

// getBaseURL returns the base URL, defaulting to OpenAI if not set
func (o *OpenAIAdapter) getBaseURL() string {
	if o.baseURL == "" {
		return DefaultOpenAIBaseURL
	}
	return o.baseURL
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
		userContent = BuildMultimodalContent(userPrompt, prompt)
	} else {
		// Text-only content
		userContent = userPrompt
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userContent,
	})

	// Build request body with extended parameters
	reqBody := map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
	}

	// Add extended sampling parameters if set (for vLLM compatibility)
	if o.topP > 0 {
		reqBody["top_p"] = o.topP
	}
	if o.topK > 0 {
		reqBody["top_k"] = o.topK
	}
	if o.presencePenalty != 0 {
		reqBody["presence_penalty"] = o.presencePenalty
	}
	if o.frequencyPenalty != 0 {
		reqBody["frequency_penalty"] = o.frequencyPenalty
	}
	if o.repetitionPenalty != 0 {
		reqBody["repetition_penalty"] = o.repetitionPenalty
	}
	if len(o.stop) > 0 {
		reqBody["stop"] = o.stop
	}

	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.getBaseURL()+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return Response{}, err
	}
	o.setHeaders(req)

	resp, err := o.httpClient.Do(req)
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
		userContent = BuildMultimodalContent(userPrompt, prompt)
	} else {
		// Text-only content
		userContent = userPrompt
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userContent,
	})

	// Build request body with extended parameters
	reqBody := map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
		"stream":      true, // Enable streaming
	}

	// Add extended sampling parameters if set (for vLLM compatibility)
	if o.topP > 0 {
		reqBody["top_p"] = o.topP
	}
	if o.topK > 0 {
		reqBody["top_k"] = o.topK
	}
	if o.presencePenalty != 0 {
		reqBody["presence_penalty"] = o.presencePenalty
	}
	if o.frequencyPenalty != 0 {
		reqBody["frequency_penalty"] = o.frequencyPenalty
	}
	if o.repetitionPenalty != 0 {
		reqBody["repetition_penalty"] = o.repetitionPenalty
	}
	if len(o.stop) > 0 {
		reqBody["stop"] = o.stop
	}

	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP request for streaming
	req, err := http.NewRequestWithContext(ctx, "POST", o.getBaseURL()+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	o.setHeaders(req)

	// Make the request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
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

	// Use the configured model for embeddings, or default to OpenAI's embedding model
	embeddingModel := o.model
	if o.baseURL == DefaultOpenAIBaseURL || o.baseURL == "" {
		// For OpenAI, use appropriate embedding model
		embeddingModel = "text-embedding-3-small"
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": embeddingModel,
		"input": texts,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.getBaseURL()+"/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	o.setHeaders(req)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embeddings API error: %s", string(body))
	}

	var response struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Data) != len(texts) {
		return nil, errors.New("number of embeddings returned does not match input")
	}

	embeddings := make([][]float64, len(texts))
	for _, item := range response.Data {
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

// Model returns the model name
func (o *OpenAIAdapter) Model() string {
	return o.model
}

// BaseURL returns the base URL
func (o *OpenAIAdapter) BaseURL() string {
	return o.getBaseURL()
}
