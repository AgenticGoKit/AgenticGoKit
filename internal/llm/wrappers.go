// Package llm provides internal LLM adapter implementations and wrappers.
package llm

import (
	"context"
	"net/http"
	"time"
)

// PublicModelProvider defines the public interface that wrappers implement
type PublicModelProvider interface {
	Call(ctx context.Context, prompt PublicPrompt) (PublicResponse, error)
	Stream(ctx context.Context, prompt PublicPrompt) (<-chan PublicToken, error)
	Embeddings(ctx context.Context, texts []string) ([][]float64, error)
}

// PublicLLMAdapter defines the public interface for simple LLM interaction
type PublicLLMAdapter interface {
	Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
}

// Public types that match the core package types
type PublicModelParameters struct {
	Temperature *float32
	MaxTokens   *int32
}

type PublicPrompt struct {
	System     string
	User       string
	Parameters PublicModelParameters
}

type PublicUsageStats struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type PublicResponse struct {
	Content      string
	Usage        PublicUsageStats
	FinishReason string
}

type PublicToken struct {
	Content string
	Error   error
}

// ModelProviderWrapper wraps internal ModelProvider to public interface
type ModelProviderWrapper struct {
	internal ModelProvider
}

func NewModelProviderWrapper(internal ModelProvider) *ModelProviderWrapper {
	return &ModelProviderWrapper{internal: internal}
}

func (w *ModelProviderWrapper) Call(ctx context.Context, prompt PublicPrompt) (PublicResponse, error) {
	internalPrompt := Prompt{
		System: prompt.System,
		User:   prompt.User,
		Parameters: ModelParameters{
			Temperature: prompt.Parameters.Temperature,
			MaxTokens:   prompt.Parameters.MaxTokens,
		},
	}

	resp, err := w.internal.Call(ctx, internalPrompt)
	if err != nil {
		return PublicResponse{}, err
	}

	return PublicResponse{
		Content: resp.Content,
		Usage: PublicUsageStats{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		FinishReason: resp.FinishReason,
	}, nil
}

func (w *ModelProviderWrapper) Stream(ctx context.Context, prompt PublicPrompt) (<-chan PublicToken, error) {
	internalPrompt := Prompt{
		System: prompt.System,
		User:   prompt.User,
		Parameters: ModelParameters{
			Temperature: prompt.Parameters.Temperature,
			MaxTokens:   prompt.Parameters.MaxTokens,
		},
	}

	internalChan, err := w.internal.Stream(ctx, internalPrompt)
	if err != nil {
		return nil, err
	}

	publicChan := make(chan PublicToken)
	go func() {
		defer close(publicChan)
		for token := range internalChan {
			publicChan <- PublicToken{
				Content: token.Content,
				Error:   token.Error,
			}
		}
	}()

	return publicChan, nil
}

func (w *ModelProviderWrapper) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return w.internal.Embeddings(ctx, texts)
}

// LLMAdapterWrapper adapts public ModelProvider to LLMAdapter
type LLMAdapterWrapper struct {
	provider PublicModelProvider
}

func NewLLMAdapterWrapper(provider PublicModelProvider) *LLMAdapterWrapper {
	return &LLMAdapterWrapper{provider: provider}
}

func (w *LLMAdapterWrapper) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	resp, err := w.provider.Call(ctx, PublicPrompt{
		System: systemPrompt,
		User:   userPrompt,
		Parameters: PublicModelParameters{
			Temperature: floatPtr(0.7),
			MaxTokens:   int32Ptr(2000),
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// Helper functions are already defined in adapter.go

// Public configuration types
type PublicAzureOpenAIAdapterOptions struct {
	Endpoint            string
	APIKey              string
	ChatDeployment      string
	EmbeddingDeployment string
	HTTPClient          *http.Client
}

type PublicLLMProviderConfig struct {
	Type        string        `json:"type" toml:"type"`
	APIKey      string        `json:"api_key,omitempty" toml:"api_key,omitempty"`
	Model       string        `json:"model,omitempty" toml:"model,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty" toml:"max_tokens,omitempty"`
	Temperature float32       `json:"temperature,omitempty" toml:"temperature,omitempty"`
	
	// Azure-specific fields
	Endpoint            string `json:"endpoint,omitempty" toml:"endpoint,omitempty"`
	ChatDeployment      string `json:"chat_deployment,omitempty" toml:"chat_deployment,omitempty"`
	EmbeddingDeployment string `json:"embedding_deployment,omitempty" toml:"embedding_deployment,omitempty"`
	
	// Ollama-specific fields
	BaseURL string `json:"base_url,omitempty" toml:"base_url,omitempty"`
	
	// HTTP client configuration
	HTTPTimeout time.Duration `json:"http_timeout,omitempty" toml:"http_timeout,omitempty"`
}

// Factory functions that create wrapped providers
func NewAzureOpenAIAdapterWrapped(options PublicAzureOpenAIAdapterOptions) (PublicModelProvider, error) {
	if options.HTTPClient == nil {
		options.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	internalOptions := AzureOpenAIAdapterOptions{
		Endpoint:            options.Endpoint,
		APIKey:              options.APIKey,
		ChatDeployment:      options.ChatDeployment,
		EmbeddingDeployment: options.EmbeddingDeployment,
		HTTPClient:          options.HTTPClient,
	}

	adapter, err := NewAzureOpenAIAdapter(internalOptions)
	if err != nil {
		return nil, err
	}

	return NewModelProviderWrapper(adapter), nil
}

func NewOpenAIAdapterWrapped(apiKey, model string, maxTokens int, temperature float32) (PublicModelProvider, error) {
	adapter, err := NewOpenAIAdapter(apiKey, model, maxTokens, temperature)
	if err != nil {
		return nil, err
	}

	return NewModelProviderWrapper(adapter), nil
}

func NewOllamaAdapterWrapped(baseURL, model string, maxTokens int, temperature float32) (PublicModelProvider, error) {
	adapter, err := NewOllamaAdapter(baseURL, model, maxTokens, temperature)
	if err != nil {
		return nil, err
	}

	return NewModelProviderWrapper(adapter), nil
}

func NewModelProviderFromConfigWrapped(config PublicLLMProviderConfig) (PublicModelProvider, error) {
	internalConfig := ProviderConfig{
		Type:                ProviderType(config.Type),
		APIKey:              config.APIKey,
		Model:               config.Model,
		MaxTokens:           config.MaxTokens,
		Temperature:         config.Temperature,
		Endpoint:            config.Endpoint,
		ChatDeployment:      config.ChatDeployment,
		EmbeddingDeployment: config.EmbeddingDeployment,
		BaseURL:             config.BaseURL,
		HTTPTimeout:         config.HTTPTimeout,
	}
	
	adapter, err := CreateProviderFromConfig(internalConfig)
	if err != nil {
		return nil, err
	}
	
	return NewModelProviderWrapper(adapter), nil
}

func NewModelProviderAdapterWrapped(provider PublicModelProvider) PublicLLMAdapter {
	// If it's our wrapper, use the internal provider directly
	if wrapper, ok := provider.(*ModelProviderWrapper); ok {
		return NewModelProviderAdapter(wrapper.internal)
	}

	// Otherwise create an adapter for the public interface
	return NewLLMAdapterWrapper(provider)
}

// =============================================================================
// PUBLIC INTERFACE ADAPTERS
// =============================================================================

// PublicProviderAdapter adapts internal wrapper to public interface
type PublicProviderAdapter struct {
	wrapper PublicModelProvider
}

func NewPublicProviderAdapter(wrapper PublicModelProvider) *PublicProviderAdapter {
	return &PublicProviderAdapter{wrapper: wrapper}
}

func (a *PublicProviderAdapter) Call(ctx context.Context, prompt PublicPrompt) (PublicResponse, error) {
	return a.wrapper.Call(ctx, prompt)
}

func (a *PublicProviderAdapter) Stream(ctx context.Context, prompt PublicPrompt) (<-chan PublicToken, error) {
	return a.wrapper.Stream(ctx, prompt)
}

func (a *PublicProviderAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return a.wrapper.Embeddings(ctx, texts)
}

// PublicLLMAdapterWrapper adapts internal LLM adapter to public interface
type PublicLLMAdapterWrapper struct {
	wrapper PublicLLMAdapter
}

func NewPublicLLMAdapterWrapper(wrapper PublicLLMAdapter) *PublicLLMAdapterWrapper {
	return &PublicLLMAdapterWrapper{wrapper: wrapper}
}

func (a *PublicLLMAdapterWrapper) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	return a.wrapper.Complete(ctx, systemPrompt, userPrompt)
}

// PublicDirectLLMAdapter provides fallback for external ModelProvider implementations
type PublicDirectLLMAdapter struct {
	provider PublicModelProvider
}

func NewPublicDirectLLMAdapter(provider PublicModelProvider) *PublicDirectLLMAdapter {
	return &PublicDirectLLMAdapter{provider: provider}
}

func (a *PublicDirectLLMAdapter) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	resp, err := a.provider.Call(ctx, PublicPrompt{
		System: systemPrompt,
		User:   userPrompt,
		Parameters: PublicModelParameters{
			Temperature: floatPtr(0.7),
			MaxTokens:   int32Ptr(2000),
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}