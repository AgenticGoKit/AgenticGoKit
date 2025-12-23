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

// MLFlowGatewayConfig holds MLFlow AI Gateway configuration
type MLFlowGatewayConfig struct {
	// Base configuration
	BaseURL string `json:"base_url" toml:"base_url"` // e.g., "http://localhost:5001"

	// Route configuration - routes defined in the MLFlow Gateway config
	ChatRoute        string `json:"chat_route" toml:"chat_route"`               // Route name for chat completions
	EmbeddingsRoute  string `json:"embeddings_route" toml:"embeddings_route"`   // Route name for embeddings
	CompletionsRoute string `json:"completions_route" toml:"completions_route"` // Route name for completions

	// Model can override the route's default model
	Model string `json:"model,omitempty" toml:"model,omitempty"`

	// Authentication
	APIKey       string            `json:"api_key,omitempty" toml:"api_key"`
	ExtraHeaders map[string]string `json:"extra_headers,omitempty" toml:"extra_headers"`

	// Model parameters (forwarded to gateway)
	MaxTokens   int      `json:"max_tokens,omitempty" toml:"max_tokens"`
	Temperature float32  `json:"temperature,omitempty" toml:"temperature"`
	TopP        float32  `json:"top_p,omitempty" toml:"top_p"`
	Stop        []string `json:"stop,omitempty" toml:"stop"`

	// Retry configuration
	MaxRetries int           `json:"max_retries,omitempty" toml:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay,omitempty" toml:"retry_delay"`

	// HTTP configuration
	HTTPTimeout time.Duration `json:"http_timeout,omitempty" toml:"http_timeout"`
}

// MLFlowGatewayAdapter wraps the OpenAI adapter for MLFlow AI Gateway.
// MLFlow Gateway provides OpenAI-compatible routes, so we reuse the OpenAI adapter
// with a custom base URL pointing to the gateway's route endpoint.
type MLFlowGatewayAdapter struct {
	*OpenAIAdapter // Embed OpenAI adapter for code reuse
	config         MLFlowGatewayConfig
	maxRetries     int
	retryDelay     time.Duration
}

// NewMLFlowGatewayAdapter creates a new MLFlow AI Gateway adapter by configuring
// an OpenAI adapter with the gateway's route-based URL.
func NewMLFlowGatewayAdapter(config MLFlowGatewayConfig) (*MLFlowGatewayAdapter, error) {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:5001"
	}
	if config.ChatRoute == "" {
		return nil, fmt.Errorf("chat route is required for MLFlow Gateway")
	}

	// Set defaults
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	// Build the OpenAI-compatible endpoint URL for MLFlow Gateway
	// MLFlow Gateway uses: /gateway/{route_name}/v1
	baseURL := fmt.Sprintf("%s/gateway/%s/v1",
		strings.TrimSuffix(config.BaseURL, "/"),
		config.ChatRoute)

	// Create OpenAI adapter with MLFlow Gateway configuration
	openaiConfig := OpenAIAdapterConfig{
		APIKey:       config.APIKey,
		Model:        config.Model, // May be empty if route has a default model
		MaxTokens:    config.MaxTokens,
		Temperature:  config.Temperature,
		BaseURL:      baseURL,
		ExtraHeaders: config.ExtraHeaders,
		HTTPTimeout:  config.HTTPTimeout,
		TopP:         config.TopP,
		Stop:         config.Stop,
	}

	openaiAdapter, err := NewOpenAIAdapterWithConfig(openaiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MLFlow Gateway adapter: %w", err)
	}

	return &MLFlowGatewayAdapter{
		OpenAIAdapter: openaiAdapter,
		config:        config,
		maxRetries:    config.MaxRetries,
		retryDelay:    config.RetryDelay,
	}, nil
}

// Call delegates to the embedded OpenAI adapter with retry logic
func (m *MLFlowGatewayAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	var lastErr error
	for attempt := 0; attempt < m.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return Response{}, ctx.Err()
			case <-time.After(m.retryDelay * time.Duration(attempt)):
			}
		}

		resp, err := m.OpenAIAdapter.Call(ctx, prompt)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		// Continue retrying on error
	}
	return Response{}, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Stream delegates to the embedded OpenAI adapter
func (m *MLFlowGatewayAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	return m.OpenAIAdapter.Stream(ctx, prompt)
}

// Embeddings delegates to the embedded OpenAI adapter with the embeddings route
func (m *MLFlowGatewayAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if m.config.EmbeddingsRoute == "" {
		return nil, fmt.Errorf("embeddings route not configured for MLFlow Gateway")
	}

	// Create a temporary adapter for the embeddings route
	embeddingsBaseURL := fmt.Sprintf("%s/gateway/%s/v1",
		strings.TrimSuffix(m.config.BaseURL, "/"),
		m.config.EmbeddingsRoute)

	embeddingsConfig := OpenAIAdapterConfig{
		APIKey:       m.config.APIKey,
		Model:        m.config.Model,
		BaseURL:      embeddingsBaseURL,
		ExtraHeaders: m.config.ExtraHeaders,
		HTTPTimeout:  m.config.HTTPTimeout,
	}

	embeddingsAdapter, err := NewOpenAIAdapterWithConfig(embeddingsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings adapter: %w", err)
	}

	return embeddingsAdapter.Embeddings(ctx, texts)
}

// RouteInfo contains information about an MLFlow Gateway route
type RouteInfo struct {
	Name      string `json:"name"`
	RouteType string `json:"route_type"`
	Model     struct {
		Name     string `json:"name"`
		Provider string `json:"provider"`
	} `json:"model"`
}

// ListRoutes returns available routes from the MLFlow Gateway
func (m *MLFlowGatewayAdapter) ListRoutes(ctx context.Context) ([]RouteInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", m.config.BaseURL+"/api/2.0/gateway/routes/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create routes request: %w", err)
	}
	m.OpenAIAdapter.setHeaders(req)

	resp, err := m.OpenAIAdapter.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MLFlow Gateway unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MLFlow Gateway routes API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Routes []RouteInfo `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode routes response: %w", err)
	}

	return response.Routes, nil
}

// HealthCheck verifies the MLFlow Gateway is responsive
func (m *MLFlowGatewayAdapter) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", m.config.BaseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := m.OpenAIAdapter.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("MLFlow Gateway health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MLFlow Gateway unhealthy: status %d", resp.StatusCode)
	}
	return nil
}

// ChatRoute returns the configured chat route
func (m *MLFlowGatewayAdapter) ChatRoute() string {
	return m.config.ChatRoute
}

// EmbeddingsRoute returns the configured embeddings route
func (m *MLFlowGatewayAdapter) EmbeddingsRoute() string {
	return m.config.EmbeddingsRoute
}

// Config returns the MLFlow Gateway configuration
func (m *MLFlowGatewayAdapter) Config() MLFlowGatewayConfig {
	return m.config
}
