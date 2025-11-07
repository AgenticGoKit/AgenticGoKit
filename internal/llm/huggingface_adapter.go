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

// HuggingFaceAdapter implements the ModelProvider interface for Hugging Face APIs.
// Supports Inference API, Inference Endpoints, Text Generation Inference (TGI), and Chat API.
type HuggingFaceAdapter struct {
	apiKey            string
	model             string
	baseURL           string
	apiType           HFAPIType
	maxTokens         int
	temperature       float32
	topP              float32
	topK              int
	doSample          bool
	waitForModel      bool
	useCache          bool
	stopSequences     []string
	repetitionPenalty float32
	httpClient        *http.Client
}

// HFAPIType represents the type of Hugging Face API being used
type HFAPIType string

const (
	HFAPITypeInference HFAPIType = "inference" // Hosted Inference API
	HFAPITypeEndpoint  HFAPIType = "endpoint"  // Inference Endpoints
	HFAPITypeTGI       HFAPIType = "tgi"       // Text Generation Inference
	HFAPITypeChat      HFAPIType = "chat"      // OpenAI-compatible chat API
)

// HFAdapterOptions holds optional configuration for HuggingFaceAdapter
type HFAdapterOptions struct {
	TopP              float32
	TopK              int
	DoSample          bool
	WaitForModel      bool
	UseCache          bool
	StopSequences     []string
	RepetitionPenalty float32
	HTTPTimeout       time.Duration
}

// NewHuggingFaceAdapter creates a new HuggingFaceAdapter instance
func NewHuggingFaceAdapter(
	apiKey string,
	model string,
	baseURL string,
	apiType HFAPIType,
	maxTokens int,
	temperature float32,
	options HFAdapterOptions,
) (*HuggingFaceAdapter, error) {
	// Validate required fields
	if apiType != HFAPITypeTGI && apiKey == "" {
		return nil, errors.New("API key is required for Hugging Face API (except for local TGI)")
	}

	if model == "" {
		model = "gpt2" // Default model
	}

	if apiType == "" {
		apiType = HFAPITypeInference // Default to Inference API
	}

	// Validate baseURL requirement for non-inference API types
	if (apiType == HFAPITypeEndpoint || apiType == HFAPITypeTGI || apiType == HFAPITypeChat) && baseURL == "" {
		return nil, fmt.Errorf("base URL is required for API type: %s", apiType)
	}

	// Set defaults
	if maxTokens == 0 {
		maxTokens = 250
	}
	if temperature == 0 {
		temperature = 0.7
	}

	// Set HTTP timeout
	timeout := 120 * time.Second
	if options.HTTPTimeout > 0 {
		timeout = options.HTTPTimeout
	}

	return &HuggingFaceAdapter{
		apiKey:            apiKey,
		model:             model,
		baseURL:           strings.TrimSuffix(baseURL, "/"),
		apiType:           apiType,
		maxTokens:         maxTokens,
		temperature:       temperature,
		topP:              options.TopP,
		topK:              options.TopK,
		doSample:          options.DoSample,
		waitForModel:      options.WaitForModel,
		useCache:          options.UseCache,
		stopSequences:     options.StopSequences,
		repetitionPenalty: options.RepetitionPenalty,
		httpClient:        &http.Client{Timeout: timeout},
	}, nil
}

// buildAPIURL constructs the appropriate API URL based on apiType and model
func (h *HuggingFaceAdapter) buildAPIURL(streaming bool) string {
	switch h.apiType {
	case HFAPITypeInference:
		// As of late 2024, HuggingFace migrated to a router-based architecture
		// The new endpoint uses OpenAI-compatible chat completions format
		// See: https://huggingface.co/docs/inference-providers/index
		return "https://router.huggingface.co/v1/chat/completions"
	case HFAPITypeEndpoint:
		return h.baseURL
	case HFAPITypeTGI:
		if streaming {
			return fmt.Sprintf("%s/generate_stream", h.baseURL)
		}
		return fmt.Sprintf("%s/generate", h.baseURL)
	case HFAPITypeChat:
		return fmt.Sprintf("%s/v1/chat/completions", h.baseURL)
	default:
		return h.baseURL
	}
}

// buildInferenceRequest creates a request body for Inference API, Endpoints, or TGI
func (h *HuggingFaceAdapter) buildInferenceRequest(prompt Prompt, maxTokens int, temperature float32, stream bool) map[string]interface{} {
	// Build input string
	input := ""
	if prompt.System != "" {
		input = fmt.Sprintf("System: %s\n\nUser: %s", prompt.System, prompt.User)
	} else {
		input = prompt.User
	}

	request := map[string]interface{}{
		"inputs": input,
		"parameters": map[string]interface{}{
			"max_new_tokens":   maxTokens,
			"temperature":      temperature,
			"return_full_text": false,
			"do_sample":        h.doSample,
		},
	}

	// Add optional parameters
	if h.topP > 0 {
		request["parameters"].(map[string]interface{})["top_p"] = h.topP
	}
	if h.topK > 0 {
		request["parameters"].(map[string]interface{})["top_k"] = h.topK
	}
	if h.repetitionPenalty > 0 {
		request["parameters"].(map[string]interface{})["repetition_penalty"] = h.repetitionPenalty
	}
	if len(h.stopSequences) > 0 {
		request["parameters"].(map[string]interface{})["stop"] = h.stopSequences
	}

	// Add options for Inference API
	if h.apiType == HFAPITypeInference {
		request["options"] = map[string]bool{
			"use_cache":      h.useCache,
			"wait_for_model": h.waitForModel,
		}
	}

	// Add stream flag for TGI
	if h.apiType == HFAPITypeTGI && stream {
		request["stream"] = true
	}

	return request
}

// buildChatRequest creates a request body for Chat API (OpenAI-compatible)
func (h *HuggingFaceAdapter) buildChatRequest(prompt Prompt, maxTokens int, temperature float32, stream bool) map[string]interface{} {
	messages := []map[string]string{}

	if prompt.System != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": prompt.System,
		})
	}

	if prompt.User != "" {
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": prompt.User,
		})
	}

	request := map[string]interface{}{
		"model":       h.model,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
		"stream":      stream,
	}

	// Add optional parameters
	if h.topP > 0 {
		request["top_p"] = h.topP
	}
	if len(h.stopSequences) > 0 {
		request["stop"] = h.stopSequences
	}

	return request
}

// Call implements the ModelProvider interface for synchronous requests
func (h *HuggingFaceAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	userPrompt := prompt.User
	if userPrompt == "" {
		return Response{}, errors.New("user prompt cannot be empty")
	}

	// Get parameters with overrides
	maxTokens := h.maxTokens
	if prompt.Parameters.MaxTokens != nil {
		maxTokens = int(*prompt.Parameters.MaxTokens)
	}
	temperature := h.temperature
	if prompt.Parameters.Temperature != nil {
		temperature = *prompt.Parameters.Temperature
	}

	// Build request based on API type
	var requestBody interface{}
	switch h.apiType {
	case HFAPITypeChat, HFAPITypeInference:
		// Both Chat and Inference now use the OpenAI-compatible chat completions format
		requestBody = h.buildChatRequest(prompt, maxTokens, temperature, false)
	case HFAPITypeEndpoint, HFAPITypeTGI:
		requestBody = h.buildInferenceRequest(prompt, maxTokens, temperature, false)
	default:
		return Response{}, fmt.Errorf("unsupported API type: %s", h.apiType)
	}

	// Marshal request
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Response{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL
	url := h.buildAPIURL(false)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}

	// Make request with retry for model loading
	var resp *http.Response
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err = h.httpClient.Do(req)
		if err != nil {
			return Response{}, fmt.Errorf("request failed: %w", err)
		}

		// Check for model loading error
		if resp.StatusCode == http.StatusServiceUnavailable {
			loadingErr := h.handleModelLoading(resp)
			if loadingErr != nil && i < maxRetries-1 {
				resp.Body.Close()
				continue // Retry
			}
		}

		break
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return Response{}, h.parseError(resp.StatusCode, body)
	}

	// Parse response based on API type
	switch h.apiType {
	case HFAPITypeChat, HFAPITypeInference:
		// Both use the OpenAI-compatible chat completions format
		return h.parseChatResponse(body)
	case HFAPITypeEndpoint, HFAPITypeTGI:
		return h.parseInferenceResponse(body)
	default:
		return Response{}, fmt.Errorf("unsupported API type: %s", h.apiType)
	}
}

// Stream implements the ModelProvider interface for streaming requests
func (h *HuggingFaceAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	userPrompt := prompt.User
	if userPrompt == "" {
		return nil, errors.New("user prompt cannot be empty")
	}

	// Get parameters with overrides
	maxTokens := h.maxTokens
	if prompt.Parameters.MaxTokens != nil {
		maxTokens = int(*prompt.Parameters.MaxTokens)
	}
	temperature := h.temperature
	if prompt.Parameters.Temperature != nil {
		temperature = *prompt.Parameters.Temperature
	}

	// Build request based on API type
	var requestBody interface{}
	switch h.apiType {
	case HFAPITypeChat, HFAPITypeInference:
		// Both use the OpenAI-compatible chat completions format
		requestBody = h.buildChatRequest(prompt, maxTokens, temperature, true)
	case HFAPITypeEndpoint, HFAPITypeTGI:
		requestBody = h.buildInferenceRequest(prompt, maxTokens, temperature, true)
	default:
		return nil, fmt.Errorf("unsupported API type: %s", h.apiType)
	}

	// Marshal request
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL
	url := h.buildAPIURL(true)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}

	// Make request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, h.parseError(resp.StatusCode, body)
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
				continue
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
					return // Stream finished
				}

				// Parse based on API type
				var content string
				var parseErr error

				switch h.apiType {
				case HFAPITypeChat, HFAPITypeInference:
					// Both use the OpenAI-compatible chat completions format
					content, parseErr = h.parseChatStreamChunk(data)
				case HFAPITypeEndpoint, HFAPITypeTGI:
					content, parseErr = h.parseInferenceStreamChunk(data)
				}

				if parseErr != nil {
					tokenChan <- Token{Error: parseErr}
					return
				}

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

		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			if ctx.Err() == nil {
				tokenChan <- Token{Error: fmt.Errorf("stream read error: %w", err)}
			}
		}
	}()

	return tokenChan, nil
}

// Embeddings implements the ModelProvider interface for generating embeddings
func (h *HuggingFaceAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, errors.New("no texts provided for embeddings")
	}

	// Try the new router's models endpoint for feature extraction
	url := fmt.Sprintf("https://router.huggingface.co/hf-inference/models/%s", h.model)

	requestBody := map[string]interface{}{
		"inputs": texts,
	}

	if h.waitForModel {
		requestBody["options"] = map[string]bool{
			"wait_for_model": true,
		}
	}

	// Marshal request
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}

	// Make request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, h.parseError(resp.StatusCode, body)
	}

	// Parse embeddings
	var embeddings [][]float64
	if err := json.Unmarshal(body, &embeddings); err != nil {
		return nil, fmt.Errorf("failed to parse embeddings: %w", err)
	}

	return embeddings, nil
}

// parseInferenceResponse parses the response from Inference API, Endpoints, or TGI
func (h *HuggingFaceAdapter) parseInferenceResponse(body []byte) (Response, error) {
	var results []struct {
		GeneratedText string `json:"generated_text"`
	}

	if err := json.Unmarshal(body, &results); err != nil {
		return Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(results) == 0 {
		return Response{}, errors.New("no results returned from API")
	}

	return Response{
		Content:      results[0].GeneratedText,
		Usage:        UsageStats{}, // Not available in standard inference API
		FinishReason: "stop",
	}, nil
}

// parseChatResponse parses the response from Chat API (OpenAI-compatible)
func (h *HuggingFaceAdapter) parseChatResponse(body []byte) (Response, error) {
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

	if err := json.Unmarshal(body, &response); err != nil {
		return Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Choices) == 0 {
		return Response{}, errors.New("no choices returned from API")
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

// parseInferenceStreamChunk parses a streaming chunk from Inference API
func (h *HuggingFaceAdapter) parseInferenceStreamChunk(data string) (string, error) {
	var chunk struct {
		Token struct {
			Text string `json:"text"`
		} `json:"token"`
	}

	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return "", fmt.Errorf("failed to parse stream chunk: %w", err)
	}

	return chunk.Token.Text, nil
}

// parseChatStreamChunk parses a streaming chunk from Chat API
func (h *HuggingFaceAdapter) parseChatStreamChunk(data string) (string, error) {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}

	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return "", fmt.Errorf("failed to parse stream chunk: %w", err)
	}

	if len(chunk.Choices) > 0 {
		return chunk.Choices[0].Delta.Content, nil
	}

	return "", nil
}

// handleModelLoading handles model loading errors (503) and waits if needed
func (h *HuggingFaceAdapter) handleModelLoading(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errorResp struct {
		Error         string  `json:"error"`
		EstimatedTime float64 `json:"estimated_time"`
	}

	if json.Unmarshal(body, &errorResp) == nil {
		if strings.Contains(errorResp.Error, "loading") {
			// Model is loading, wait if estimated time is reasonable
			waitTime := time.Duration(errorResp.EstimatedTime) * time.Second
			if waitTime > 60*time.Second {
				waitTime = 60 * time.Second
			}
			time.Sleep(waitTime)
			return fmt.Errorf("model loading, waited %v seconds", waitTime.Seconds())
		}
	}

	return nil
}

// parseError parses error responses from Hugging Face APIs
func (h *HuggingFaceAdapter) parseError(statusCode int, body []byte) error {
	var errorResp struct {
		Error string `json:"error"`
	}

	if json.Unmarshal(body, &errorResp) == nil && errorResp.Error != "" {
		return fmt.Errorf("Hugging Face API error (%d): %s", statusCode, errorResp.Error)
	}

	return fmt.Errorf("Hugging Face API error (%d): %s", statusCode, string(body))
}
