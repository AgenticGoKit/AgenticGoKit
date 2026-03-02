package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultFoundryLocalBaseURL is the default Foundry Local API endpoint
const DefaultFoundryLocalBaseURL = "http://localhost:5272/v1"

// FoundryLocalAdapter wraps OpenAIAdapter for Azure AI Foundry Local.
// Foundry Local exposes an OpenAI-compatible API but rejects Authorization headers.
type FoundryLocalAdapter struct {
	*OpenAIAdapter
	managementURL string // base URL without /v1 suffix, for management endpoints
}

// FoundryLocalConfig holds configuration for the Foundry Local adapter
type FoundryLocalConfig struct {
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float32
	HTTPTimeout time.Duration
}

// NewFoundryLocalAdapter creates a new adapter for Azure AI Foundry Local.
func NewFoundryLocalAdapter(config FoundryLocalConfig) (*FoundryLocalAdapter, error) {
	if config.Model == "" {
		return nil, fmt.Errorf("model is required for Foundry Local")
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultFoundryLocalBaseURL
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

	// Create underlying OpenAI adapter with empty API key
	openaiAdapter, err := NewOpenAIAdapterWithConfig(OpenAIAdapterConfig{
		APIKey:      "", // Foundry Local doesn't need auth
		Model:       config.Model,
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		BaseURL:     config.BaseURL,
		HTTPTimeout: config.HTTPTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Foundry Local adapter: %w", err)
	}

	// Derive management URL (strip /v1 suffix if present)
	mgmtURL := config.BaseURL
	if len(mgmtURL) > 3 && mgmtURL[len(mgmtURL)-3:] == "/v1" {
		mgmtURL = mgmtURL[:len(mgmtURL)-3]
	}

	return &FoundryLocalAdapter{
		OpenAIAdapter: openaiAdapter,
		managementURL: mgmtURL,
	}, nil
}

// setHeaders overrides OpenAIAdapter.setHeaders to skip Authorization header.
// Foundry Local rejects Bearer tokens with 401.
func (f *FoundryLocalAdapter) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header — Foundry Local doesn't need/want it
}

// Call delegates to OpenAIAdapter but uses our header override.
func (f *FoundryLocalAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	// We need to intercept the HTTP requests to remove auth headers.
	// The simplest approach: temporarily clear the API key, call parent, restore.
	origKey := f.OpenAIAdapter.apiKey
	f.OpenAIAdapter.apiKey = ""
	defer func() { f.OpenAIAdapter.apiKey = origKey }()
	return f.OpenAIAdapter.Call(ctx, prompt)
}

// Stream delegates to OpenAIAdapter but uses our header override.
func (f *FoundryLocalAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	origKey := f.OpenAIAdapter.apiKey
	f.OpenAIAdapter.apiKey = ""
	defer func() { f.OpenAIAdapter.apiKey = origKey }()
	return f.OpenAIAdapter.Stream(ctx, prompt)
}

// Embeddings delegates to OpenAIAdapter but uses our header override.
func (f *FoundryLocalAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	origKey := f.OpenAIAdapter.apiKey
	f.OpenAIAdapter.apiKey = ""
	defer func() { f.OpenAIAdapter.apiKey = origKey }()
	return f.OpenAIAdapter.Embeddings(ctx, texts)
}

// =============================================================================
// FOUNDRY LOCAL MANAGEMENT APIs
// =============================================================================

// FoundryLocalStatus represents the server status
type FoundryLocalStatus struct {
	Endpoints    []string `json:"Endpoints"`
	ModelDirPath string   `json:"ModelDirPath"`
	PipeName     string   `json:"PipeName"`
}

// FoundryLocalModel represents a model in the Foundry catalog
type FoundryLocalModel struct {
	Name                string `json:"name"`
	DisplayName         string `json:"displayName"`
	Publisher           string `json:"publisher"`
	Task                string `json:"task"`
	FileSizeMB          int    `json:"fileSizeMb"`
	SupportsToolCalling bool   `json:"supportsToolCalling"`
	License             string `json:"license"`
	Alias               string `json:"alias"`
}

// Status returns Foundry Local server status (GET /openai/status)
func (f *FoundryLocalAdapter) Status(ctx context.Context) (*FoundryLocalStatus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", f.managementURL+"/openai/status", nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("foundry local status request failed: %w", err)
	}
	defer resp.Body.Close()

	var status FoundryLocalStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode foundry local status: %w", err)
	}
	return &status, nil
}

// ListCatalogModels returns available models from the Foundry catalog (GET /foundry/list)
func (f *FoundryLocalAdapter) ListCatalogModels(ctx context.Context) ([]FoundryLocalModel, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", f.managementURL+"/foundry/list", nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("foundry local list request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Models []FoundryLocalModel `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode foundry local models: %w", err)
	}
	return result.Models, nil
}

// LoadedModels returns currently loaded models (GET /openai/loadedmodels)
func (f *FoundryLocalAdapter) LoadedModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", f.managementURL+"/openai/loadedmodels", nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("foundry local loaded models request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var models []string
	if err := json.Unmarshal(body, &models); err != nil {
		return nil, fmt.Errorf("failed to decode loaded models: %w", err)
	}
	return models, nil
}

// LoadModel loads a model into memory (GET /openai/load/{name})
func (f *FoundryLocalAdapter) LoadModel(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/openai/load/%s", f.managementURL, name), nil)
	if err != nil {
		return err
	}
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("foundry local load model failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("foundry local load model failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
