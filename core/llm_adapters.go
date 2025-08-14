// Package core provides public LLM adapter constructors for AgentFlow.
package core

import (
	"context"
	"net/http"
	"time"

	"github.com/kunalkushwaha/agenticgokit/internal/llm"
)

// AzureOpenAIAdapterOptions holds configuration for Azure OpenAI adapter
type AzureOpenAIAdapterOptions struct {
	Endpoint            string
	APIKey              string
	ChatDeployment      string
	EmbeddingDeployment string
	HTTPClient          *http.Client
}

// NewAzureOpenAIAdapter creates a new Azure OpenAI adapter
func NewAzureOpenAIAdapter(options AzureOpenAIAdapterOptions) (ModelProvider, error) {
	if options.HTTPClient == nil {
		options.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	internalOptions := llm.AzureOpenAIAdapterOptions{
		Endpoint:            options.Endpoint,
		APIKey:              options.APIKey,
		ChatDeployment:      options.ChatDeployment,
		EmbeddingDeployment: options.EmbeddingDeployment,
		HTTPClient:          options.HTTPClient,
	}

	adapter, err := llm.NewAzureOpenAIAdapter(internalOptions)
	if err != nil {
		return nil, err
	}

	return &modelProviderWrapper{internal: adapter}, nil
}

// NewOpenAIAdapter creates a new OpenAI adapter
func NewOpenAIAdapter(apiKey, model string, maxTokens int, temperature float32) (ModelProvider, error) {
	adapter, err := llm.NewOpenAIAdapter(apiKey, model, maxTokens, temperature)
	if err != nil {
		return nil, err
	}

	return &modelProviderWrapper{internal: adapter}, nil
}

// NewOllamaAdapter creates a new Ollama adapter
func NewOllamaAdapter(baseURL, model string, maxTokens int, temperature float32) (ModelProvider, error) {
	adapter, err := llm.NewOllamaAdapter(baseURL, model, maxTokens, temperature)
	if err != nil {
		return nil, err
	}

	return &modelProviderWrapper{internal: adapter}, nil
}

// NewModelProviderAdapter creates an LLMAdapter from a ModelProvider
func NewModelProviderAdapter(provider ModelProvider) LLMAdapter {
	// If it's our wrapper, use the internal provider directly
	if wrapper, ok := provider.(*modelProviderWrapper); ok {
		return llm.NewModelProviderAdapter(wrapper.internal)
	}

	// Otherwise create an adapter for the public interface
	return &llmAdapterWrapper{provider: provider}
}

// LLMProviderConfig holds configuration for creating LLM providers
type LLMProviderConfig struct {
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

// NewModelProviderFromConfig creates a ModelProvider from configuration
func NewModelProviderFromConfig(config LLMProviderConfig) (ModelProvider, error) {
	internalConfig := llm.ProviderConfig{
		Type:                llm.ProviderType(config.Type),
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
	
	adapter, err := llm.CreateProviderFromConfig(internalConfig)
	if err != nil {
		return nil, err
	}
	
	return &modelProviderWrapper{internal: adapter}, nil
}

// modelProviderWrapper wraps internal ModelProvider to public interface
type modelProviderWrapper struct {
	internal llm.ModelProvider
}

func (w *modelProviderWrapper) Call(ctx context.Context, prompt Prompt) (Response, error) {
	internalPrompt := llm.Prompt{
		System: prompt.System,
		User:   prompt.User,
		Parameters: llm.ModelParameters{
			Temperature: prompt.Parameters.Temperature,
			MaxTokens:   prompt.Parameters.MaxTokens,
		},
	}

	resp, err := w.internal.Call(ctx, internalPrompt)
	if err != nil {
		return Response{}, err
	}

	return Response{
		Content: resp.Content,
		Usage: UsageStats{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		FinishReason: resp.FinishReason,
	}, nil
}

func (w *modelProviderWrapper) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	internalPrompt := llm.Prompt{
		System: prompt.System,
		User:   prompt.User,
		Parameters: llm.ModelParameters{
			Temperature: prompt.Parameters.Temperature,
			MaxTokens:   prompt.Parameters.MaxTokens,
		},
	}

	internalChan, err := w.internal.Stream(ctx, internalPrompt)
	if err != nil {
		return nil, err
	}

	publicChan := make(chan Token)
	go func() {
		defer close(publicChan)
		for token := range internalChan {
			publicChan <- Token{
				Content: token.Content,
				Error:   token.Error,
			}
		}
	}()

	return publicChan, nil
}

func (w *modelProviderWrapper) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return w.internal.Embeddings(ctx, texts)
}

// llmAdapterWrapper adapts public ModelProvider to LLMAdapter
type llmAdapterWrapper struct {
	provider ModelProvider
}

func (w *llmAdapterWrapper) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	resp, err := w.provider.Call(ctx, Prompt{
		System: systemPrompt,
		User:   userPrompt,
		Parameters: ModelParameters{
			Temperature: FloatPtr(0.7),
			MaxTokens:   Int32Ptr(2000),
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
