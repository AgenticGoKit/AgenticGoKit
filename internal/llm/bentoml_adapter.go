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

// BentoMLConfig holds BentoML-specific configuration
type BentoMLConfig struct {
	// Base configuration
	BaseURL     string `json:"base_url" toml:"base_url"`         // e.g., "http://localhost:3000"
	APIKey      string `json:"api_key,omitempty" toml:"api_key"` // Optional - BentoML API token
	Model       string `json:"model" toml:"model"`               // Model name/path
	MaxTokens   int    `json:"max_tokens,omitempty" toml:"max_tokens"`
	Temperature float32 `json:"temperature,omitempty" toml:"temperature"`

	// Sampling parameters (OpenAI-compatible)
	TopP             float32  `json:"top_p,omitempty" toml:"top_p"`
	TopK             int      `json:"top_k,omitempty" toml:"top_k"`
	PresencePenalty  float32  `json:"presence_penalty,omitempty" toml:"presence_penalty"`
	FrequencyPenalty float32  `json:"frequency_penalty,omitempty" toml:"frequency_penalty"`
	Stop             []string `json:"stop,omitempty" toml:"stop"`

	// BentoML-specific options
	ServiceName string            `json:"service_name,omitempty" toml:"service_name"` // BentoML service name
	Runners     []string          `json:"runners,omitempty" toml:"runners"`           // Specific runners to use
	ExtraHeaders map[string]string `json:"extra_headers,omitempty" toml:"extra_headers"`

	// Retry configuration
	MaxRetries int           `json:"max_retries,omitempty" toml:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay,omitempty" toml:"retry_delay"`

	// HTTP configuration
	HTTPTimeout time.Duration `json:"http_timeout,omitempty" toml:"http_timeout"`
}

// BentoMLAdapter wraps the OpenAI adapter for BentoML inference servers.
// BentoML provides an OpenAI-compatible API, so we reuse the OpenAI adapter
// with a custom base URL pointing to the BentoML server.
type BentoMLAdapter struct {
	*OpenAIAdapter // Embed OpenAI adapter for code reuse
	config         BentoMLConfig
	maxRetries     int
	retryDelay     time.Duration
}

// NewBentoMLAdapter creates a new BentoML adapter by configuring an OpenAI adapter
// with the BentoML server's base URL.
func NewBentoMLAdapter(config BentoMLConfig) (*BentoMLAdapter, error) {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:3000"
	}
	if config.Model == "" {
		return nil, fmt.Errorf("model name is required for BentoML")
	}

	// Set defaults
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	// Build the OpenAI-compatible endpoint URL for BentoML
	// BentoML OpenAI-compatible endpoint uses /v1 prefix
	baseURL := strings.TrimSuffix(config.BaseURL, "/") + "/v1"

	// Create OpenAI adapter with BentoML configuration
	openaiConfig := OpenAIAdapterConfig{
		APIKey:           config.APIKey,
		Model:            config.Model,
		MaxTokens:        config.MaxTokens,
		Temperature:      config.Temperature,
		BaseURL:          baseURL,
		ExtraHeaders:     config.ExtraHeaders,
		HTTPTimeout:      config.HTTPTimeout,
		TopP:             config.TopP,
		TopK:             config.TopK,
		PresencePenalty:  config.PresencePenalty,
		FrequencyPenalty: config.FrequencyPenalty,
		Stop:             config.Stop,
	}

	openaiAdapter, err := NewOpenAIAdapterWithConfig(openaiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create BentoML adapter: %w", err)
	}

	return &BentoMLAdapter{
		OpenAIAdapter: openaiAdapter,
		config:        config,
		maxRetries:    config.MaxRetries,
		retryDelay:    config.RetryDelay,
	}, nil
}

// Call delegates to the embedded OpenAI adapter with retry logic
func (b *BentoMLAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	var lastErr error
	for attempt := 0; attempt < b.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return Response{}, ctx.Err()
			case <-time.After(b.retryDelay * time.Duration(attempt)):
			}
		}

		resp, err := b.OpenAIAdapter.Call(ctx, prompt)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return Response{}, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Stream delegates to the embedded OpenAI adapter
func (b *BentoMLAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	return b.OpenAIAdapter.Stream(ctx, prompt)
}

// Embeddings delegates to the embedded OpenAI adapter
func (b *BentoMLAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return b.OpenAIAdapter.Embeddings(ctx, texts)
}

// ListModels returns available models from the BentoML server
func (b *BentoMLAdapter) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", b.config.BaseURL+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create models request: %w", err)
	}
	b.OpenAIAdapter.setHeaders(req)

	resp, err := b.OpenAIAdapter.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("BentoML server unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("BentoML models API error (status %d): %s", resp.StatusCode, string(body))
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

// HealthCheck verifies the BentoML server is responsive
func (b *BentoMLAdapter) HealthCheck(ctx context.Context) error {
	// BentoML uses /healthz or /livez for health checks
	req, err := http.NewRequestWithContext(ctx, "GET", b.config.BaseURL+"/healthz", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := b.OpenAIAdapter.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("BentoML health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("BentoML server unhealthy: status %d", resp.StatusCode)
	}
	return nil
}

// Readiness checks if the BentoML server is ready to serve requests
func (b *BentoMLAdapter) Readiness(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", b.config.BaseURL+"/readyz", nil)
	if err != nil {
		return fmt.Errorf("failed to create readiness check request: %w", err)
	}

	resp, err := b.OpenAIAdapter.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("BentoML readiness check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("BentoML server not ready: status %d", resp.StatusCode)
	}
	return nil
}

// ServiceName returns the configured BentoML service name
func (b *BentoMLAdapter) ServiceName() string {
	return b.config.ServiceName
}

// Config returns the BentoML configuration
func (b *BentoMLAdapter) Config() BentoMLConfig {
	return b.config
}
