// Package llm provides internal LLM adapter implementations.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VLLMConfig holds vLLM-specific configuration
type VLLMConfig struct {
	// Base configuration
	BaseURL     string  `json:"base_url" toml:"base_url"`         // e.g., "http://localhost:8000"
	APIKey      string  `json:"api_key,omitempty" toml:"api_key"` // Optional
	Model       string  `json:"model" toml:"model"`               // Model name/path
	MaxTokens   int     `json:"max_tokens,omitempty" toml:"max_tokens"`
	Temperature float32 `json:"temperature,omitempty" toml:"temperature"`

	// vLLM-specific parameters
	BestOf            int      `json:"best_of,omitempty" toml:"best_of"`                         // Generate N completions, return best
	UseBeamSearch     bool     `json:"use_beam_search,omitempty" toml:"use_beam_search"`         // Use beam search
	TopK              int      `json:"top_k,omitempty" toml:"top_k"`                             // Top-k sampling
	TopP              float32  `json:"top_p,omitempty" toml:"top_p"`                             // Nucleus sampling
	MinP              float32  `json:"min_p,omitempty" toml:"min_p"`                             // Min-p sampling
	PresencePenalty   float32  `json:"presence_penalty,omitempty" toml:"presence_penalty"`       // Presence penalty
	FrequencyPenalty  float32  `json:"frequency_penalty,omitempty" toml:"frequency_penalty"`     // Frequency penalty
	RepetitionPenalty float32  `json:"repetition_penalty,omitempty" toml:"repetition_penalty"`   // Repetition penalty
	LengthPenalty     float32  `json:"length_penalty,omitempty" toml:"length_penalty"`           // Length penalty for beam search
	StopTokenIds      []int    `json:"stop_token_ids,omitempty" toml:"stop_token_ids"`           // Token IDs to stop generation
	SkipSpecialTokens bool     `json:"skip_special_tokens,omitempty" toml:"skip_special_tokens"` // Skip special tokens
	IgnoreEOS         bool     `json:"ignore_eos,omitempty" toml:"ignore_eos"`                   // Ignore EOS token
	Stop              []string `json:"stop,omitempty" toml:"stop"`                               // Stop sequences

	// HTTP configuration
	HTTPTimeout time.Duration `json:"http_timeout,omitempty" toml:"http_timeout"`
}

// VLLMAdapter wraps the OpenAI adapter for vLLM inference servers.
// vLLM provides an OpenAI-compatible API, so we reuse the OpenAI adapter
// with a custom base URL pointing to the vLLM server.
type VLLMAdapter struct {
	*OpenAIAdapter // Embed OpenAI adapter for code reuse
	config         VLLMConfig
}

// NewVLLMAdapter creates a new vLLM adapter by configuring an OpenAI adapter
// with the vLLM server's base URL.
func NewVLLMAdapter(config VLLMConfig) (*VLLMAdapter, error) {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8000"
	}
	if config.Model == "" {
		return nil, fmt.Errorf("model name is required for vLLM")
	}

	// Create OpenAI adapter with vLLM configuration
	openaiConfig := OpenAIAdapterConfig{
		APIKey:            config.APIKey,
		Model:             config.Model,
		MaxTokens:         config.MaxTokens,
		Temperature:       config.Temperature,
		BaseURL:           strings.TrimSuffix(config.BaseURL, "/") + "/v1", // vLLM uses /v1 prefix
		HTTPTimeout:       config.HTTPTimeout,
		TopP:              config.TopP,
		TopK:              config.TopK,
		PresencePenalty:   config.PresencePenalty,
		FrequencyPenalty:  config.FrequencyPenalty,
		RepetitionPenalty: config.RepetitionPenalty,
		Stop:              config.Stop,
	}

	openaiAdapter, err := NewOpenAIAdapterWithConfig(openaiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create vLLM adapter: %w", err)
	}

	return &VLLMAdapter{
		OpenAIAdapter: openaiAdapter,
		config:        config,
	}, nil
}

// Call delegates to the embedded OpenAI adapter
func (v *VLLMAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	return v.OpenAIAdapter.Call(ctx, prompt)
}

// Stream delegates to the embedded OpenAI adapter
func (v *VLLMAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	return v.OpenAIAdapter.Stream(ctx, prompt)
}

// Embeddings delegates to the embedded OpenAI adapter
func (v *VLLMAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return v.OpenAIAdapter.Embeddings(ctx, texts)
}

// ListModels returns available models from the vLLM server
func (v *VLLMAdapter) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", v.config.BaseURL+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create models request: %w", err)
	}
	v.OpenAIAdapter.setHeaders(req)

	resp, err := v.OpenAIAdapter.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vLLM server unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vLLM models API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	models := make([]string, len(response.Data))
	for i, m := range response.Data {
		models[i] = m.ID
	}
	return models, nil
}

// HealthCheck verifies the vLLM server is responsive
func (v *VLLMAdapter) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", v.config.BaseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := v.OpenAIAdapter.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("vLLM health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vLLM server unhealthy: status %d", resp.StatusCode)
	}
	return nil
}

// Config returns the vLLM configuration
func (v *VLLMAdapter) Config() VLLMConfig {
	return v.config
}
