// Package llm provides internal LLM factory functionality.
package llm

import (
	"fmt"
	"net/http"
	"time"
)

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	ProviderTypeOpenAI      ProviderType = "openai"
	ProviderTypeAzureOpenAI ProviderType = "azure"
	ProviderTypeOllama      ProviderType = "ollama"
)

// ProviderConfig holds configuration for creating LLM providers
type ProviderConfig struct {
	Type        ProviderType `json:"type" toml:"type"`
	APIKey      string       `json:"api_key,omitempty" toml:"api_key,omitempty"`
	Model       string       `json:"model,omitempty" toml:"model,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty" toml:"max_tokens,omitempty"`
	Temperature float32      `json:"temperature,omitempty" toml:"temperature,omitempty"`
	
	// Azure-specific fields
	Endpoint            string `json:"endpoint,omitempty" toml:"endpoint,omitempty"`
	ChatDeployment      string `json:"chat_deployment,omitempty" toml:"chat_deployment,omitempty"`
	EmbeddingDeployment string `json:"embedding_deployment,omitempty" toml:"embedding_deployment,omitempty"`
	
	// Ollama-specific fields
	BaseURL string `json:"base_url,omitempty" toml:"base_url,omitempty"`
	
	// HTTP client configuration
	HTTPTimeout time.Duration `json:"http_timeout,omitempty" toml:"http_timeout,omitempty"`
}

// ProviderFactory creates LLM providers based on configuration
type ProviderFactory struct {
	httpClient *http.Client
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetHTTPClient sets a custom HTTP client for the factory
func (f *ProviderFactory) SetHTTPClient(client *http.Client) {
	f.httpClient = client
}

// CreateProvider creates a ModelProvider based on the configuration
func (f *ProviderFactory) CreateProvider(config ProviderConfig) (ModelProvider, error) {
	// Set defaults
	if config.MaxTokens == 0 {
		config.MaxTokens = 150
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.HTTPTimeout > 0 && f.httpClient.Timeout != config.HTTPTimeout {
		f.httpClient = &http.Client{Timeout: config.HTTPTimeout}
	}

	switch config.Type {
	case ProviderTypeOpenAI:
		return f.createOpenAIProvider(config)
	case ProviderTypeAzureOpenAI:
		return f.createAzureProvider(config)
	case ProviderTypeOllama:
		return f.createOllamaProvider(config)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}
}

// createOpenAIProvider creates an OpenAI provider
func (f *ProviderFactory) createOpenAIProvider(config ProviderConfig) (ModelProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for OpenAI provider")
	}
	if config.Model == "" {
		config.Model = "gpt-4o-mini" // Default model
	}
	
	return NewOpenAIAdapter(config.APIKey, config.Model, config.MaxTokens, config.Temperature)
}

// createAzureProvider creates an Azure OpenAI provider
func (f *ProviderFactory) createAzureProvider(config ProviderConfig) (ModelProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Azure OpenAI provider")
	}
	if config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required for Azure OpenAI provider")
	}
	if config.ChatDeployment == "" {
		return nil, fmt.Errorf("chat deployment is required for Azure OpenAI provider")
	}
	if config.EmbeddingDeployment == "" {
		return nil, fmt.Errorf("embedding deployment is required for Azure OpenAI provider")
	}
	
	options := AzureOpenAIAdapterOptions{
		Endpoint:            config.Endpoint,
		APIKey:              config.APIKey,
		ChatDeployment:      config.ChatDeployment,
		EmbeddingDeployment: config.EmbeddingDeployment,
		HTTPClient:          f.httpClient,
	}
	
	return NewAzureOpenAIAdapter(options)
}

// createOllamaProvider creates an Ollama provider
func (f *ProviderFactory) createOllamaProvider(config ProviderConfig) (ModelProvider, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434" // Default Ollama URL
	}
	
	model := config.Model
	if model == "" {
		model = "llama3.2:latest" // Default model
	}
	
	return NewOllamaAdapter(baseURL, model, config.MaxTokens, config.Temperature)
}

// DefaultFactory is a global factory instance for convenience
var DefaultFactory = NewProviderFactory()

// CreateProviderFromConfig is a convenience function that uses the default factory
func CreateProviderFromConfig(config ProviderConfig) (ModelProvider, error) {
	return DefaultFactory.CreateProvider(config)
}