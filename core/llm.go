// Package core provides public LLM interfaces and essential types for AgentFlow.
package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/internal/llm"
)

// Essential public types for LLM interaction

// ModelParameters holds common configuration options for language model calls.
type ModelParameters struct {
	Temperature *float32 // Sampling temperature. nil uses the provider's default.
	MaxTokens   *int32   // Max tokens to generate. nil uses the provider's default.
}

// Prompt represents the input to a language model call.
type Prompt struct {
	System     string          // System message sets the context or instructions for the model.
	User       string          // User message is the primary input or question.
	Parameters ModelParameters // Parameters specify model configuration for this call.
}

// UsageStats contains token usage information for a model call.
type UsageStats struct {
	PromptTokens     int // Tokens in the input prompt.
	CompletionTokens int // Tokens generated in the response.
	TotalTokens      int // Total tokens processed.
}

// Response represents the output from a language model call.
type Response struct {
	Content      string     // The primary text response from the model.
	Usage        UsageStats // Token usage statistics for the call.
	FinishReason string     // Why the model stopped generating tokens.
}

// Token represents a single token streamed from a language model.
type Token struct {
	Content string // The text chunk of the token.
	Error   error  // Any error that occurred during streaming.
}

// ModelProvider defines the interface for interacting with different language model backends.
// This is the primary interface for LLM operations.
type ModelProvider interface {
	// Call sends a prompt to the model and returns a complete response.
	Call(ctx context.Context, prompt Prompt) (Response, error)

	// Stream sends a prompt to the model and returns a channel of tokens.
	Stream(ctx context.Context, prompt Prompt) (<-chan Token, error)

	// Embeddings generates vector embeddings for the provided texts.
	Embeddings(ctx context.Context, texts []string) ([][]float64, error)
}

// LLMAdapter defines a simplified interface for basic LLM interaction.
// Use this interface when you only need simple completion functionality.
type LLMAdapter interface {
	// Complete sends a prompt to the LLM and returns the completion
	Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
}

// =============================================================================
// REGISTRY FOR MODEL PROVIDERS (Plugins register here)
// =============================================================================

// ModelProviderFactory is a constructor for a ModelProvider based on config.
type ModelProviderFactory func(LLMProviderConfig) (ModelProvider, error)

var (
	modelProviderFactoriesMu sync.RWMutex
	modelProviderFactories   = map[string]ModelProviderFactory{}
)

// RegisterModelProviderFactory registers a factory under a provider type key (e.g., "openai").
// Keys are stored in lowercase.
func RegisterModelProviderFactory(name string, factory ModelProviderFactory) {
	if name == "" || factory == nil {
		return
	}
	modelProviderFactoriesMu.Lock()
	modelProviderFactories[strings.ToLower(name)] = factory
	modelProviderFactoriesMu.Unlock()
}

func getModelProviderFactory(name string) (ModelProviderFactory, bool) {
	modelProviderFactoriesMu.RLock()
	f, ok := modelProviderFactories[strings.ToLower(name)]
	modelProviderFactoriesMu.RUnlock()
	return f, ok
}

// Helper functions for creating parameter pointers
func FloatPtr(f float32) *float32 {
	return &f
}

func Int32Ptr(i int32) *int32 {
	return &i
}

// =============================================================================
// PUBLIC FACTORY FUNCTIONS
// =============================================================================

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
	internalOptions := llm.PublicAzureOpenAIAdapterOptions{
		Endpoint:            options.Endpoint,
		APIKey:              options.APIKey,
		ChatDeployment:      options.ChatDeployment,
		EmbeddingDeployment: options.EmbeddingDeployment,
		HTTPClient:          options.HTTPClient,
	}

	wrapper, err := llm.NewAzureOpenAIAdapterWrapped(internalOptions)
	if err != nil {
		return nil, err
	}

	return &coreModelProviderAdapter{adapter: llm.NewPublicProviderAdapter(wrapper)}, nil
}

// NewOpenAIAdapter creates a new OpenAI adapter
func NewOpenAIAdapter(apiKey, model string, maxTokens int, temperature float32) (ModelProvider, error) {
	wrapper, err := llm.NewOpenAIAdapterWrapped(apiKey, model, maxTokens, temperature)
	if err != nil {
		return nil, err
	}

	return &coreModelProviderAdapter{adapter: llm.NewPublicProviderAdapter(wrapper)}, nil
}

// NewOllamaAdapter creates a new Ollama adapter
func NewOllamaAdapter(baseURL, model string, maxTokens int, temperature float32) (ModelProvider, error) {
	wrapper, err := llm.NewOllamaAdapterWrapped(baseURL, model, maxTokens, temperature)
	if err != nil {
		return nil, err
	}

	return &coreModelProviderAdapter{adapter: llm.NewPublicProviderAdapter(wrapper)}, nil
}

// LLMProviderConfig holds configuration for creating LLM providers
type LLMProviderConfig struct {
	Type        string  `json:"type" toml:"type"`
	APIKey      string  `json:"api_key,omitempty" toml:"api_key,omitempty"`
	Model       string  `json:"model,omitempty" toml:"model,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty" toml:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty" toml:"temperature,omitempty"`

	// Azure-specific fields
	Endpoint            string `json:"endpoint,omitempty" toml:"endpoint,omitempty"`
	ChatDeployment      string `json:"chat_deployment,omitempty" toml:"chat_deployment,omitempty"`
	EmbeddingDeployment string `json:"embedding_deployment,omitempty" toml:"embedding_deployment,omitempty"`

	// Ollama-specific fields
	BaseURL string `json:"base_url,omitempty" toml:"base_url,omitempty"`

	// OpenRouter-specific fields
	SiteURL  string `json:"site_url,omitempty" toml:"site_url,omitempty"`
	SiteName string `json:"site_name,omitempty" toml:"site_name,omitempty"`

	// HuggingFace-specific fields
	HFAPIType           string   `json:"hf_api_type,omitempty" toml:"hf_api_type,omitempty"`
	HFWaitForModel      bool     `json:"hf_wait_for_model,omitempty" toml:"hf_wait_for_model,omitempty"`
	HFUseCache          bool     `json:"hf_use_cache,omitempty" toml:"hf_use_cache,omitempty"`
	HFTopP              float64  `json:"hf_top_p,omitempty" toml:"hf_top_p,omitempty"`
	HFTopK              int      `json:"hf_top_k,omitempty" toml:"hf_top_k,omitempty"`
	HFDoSample          bool     `json:"hf_do_sample,omitempty" toml:"hf_do_sample,omitempty"`
	HFStopSequences     []string `json:"hf_stop_sequences,omitempty" toml:"hf_stop_sequences,omitempty"`
	HFRepetitionPenalty float64  `json:"hf_repetition_penalty,omitempty" toml:"hf_repetition_penalty,omitempty"`

	// HTTP client configuration
	HTTPTimeout time.Duration `json:"http_timeout,omitempty" toml:"http_timeout,omitempty"`
}

// NewModelProviderFromConfig creates a ModelProvider from configuration
func NewModelProviderFromConfig(config LLMProviderConfig) (ModelProvider, error) {
	// 1) Try plugin registry first
	if f, ok := getModelProviderFactory(config.Type); ok {
		return f(config)
	}

	// 2) Fallback to legacy internal implementation to avoid breaking existing users during transition
	internalConfig := llm.PublicLLMProviderConfig{
		Type:                config.Type,
		APIKey:              config.APIKey,
		Model:               config.Model,
		MaxTokens:           config.MaxTokens,
		Temperature:         config.Temperature,
		Endpoint:            config.Endpoint,
		ChatDeployment:      config.ChatDeployment,
		EmbeddingDeployment: config.EmbeddingDeployment,
		BaseURL:             config.BaseURL,
		SiteURL:             config.SiteURL,
		SiteName:            config.SiteName,
		HFAPIType:           config.HFAPIType,
		HFWaitForModel:      config.HFWaitForModel,
		HFUseCache:          config.HFUseCache,
		HFTopP:              config.HFTopP,
		HFTopK:              config.HFTopK,
		HFDoSample:          config.HFDoSample,
		HFStopSequences:     config.HFStopSequences,
		HFRepetitionPenalty: config.HFRepetitionPenalty,
		HTTPTimeout:         config.HTTPTimeout,
	}

	wrapper, err := llm.NewModelProviderFromConfigWrapped(internalConfig)
	if err != nil {
		// Provide actionable error guiding users to import a plugin
		if strings.TrimSpace(config.Type) != "" {
			return nil, fmt.Errorf("llm provider '%s' not registered. Import the plugin: _ 'github.com/kunalkushwaha/agenticgokit/plugins/llm/%s' (original error: %w)", config.Type, strings.ToLower(config.Type), err)
		}
		return nil, fmt.Errorf("llm provider type not specified; set LLMProviderConfig.Type and import the matching plugin (original error: %w)", err)
	}

	return &coreModelProviderAdapter{adapter: llm.NewPublicProviderAdapter(wrapper)}, nil
}

// NewModelProviderAdapter creates an LLMAdapter from a ModelProvider
func NewModelProviderAdapter(provider ModelProvider) LLMAdapter {
	if adapter, ok := provider.(*coreModelProviderAdapter); ok {
		return &coreLLMAdapter{adapter: llm.NewPublicLLMAdapterWrapper(llm.NewModelProviderAdapterWrapped(adapter.adapter))}
	}

	// Fallback for external implementations
	return &directCoreLLMAdapter{provider: provider}
}

// =============================================================================
// MINIMAL CORE ADAPTERS
// =============================================================================

// coreModelProviderAdapter provides minimal adapter for core interface
type coreModelProviderAdapter struct {
	adapter *llm.PublicProviderAdapter
}

func (a *coreModelProviderAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
	internalPrompt := llm.PublicPrompt{
		System: prompt.System,
		User:   prompt.User,
		Parameters: llm.PublicModelParameters{
			Temperature: prompt.Parameters.Temperature,
			MaxTokens:   prompt.Parameters.MaxTokens,
		},
	}

	resp, err := a.adapter.Call(ctx, internalPrompt)
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

func (a *coreModelProviderAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	internalPrompt := llm.PublicPrompt{
		System: prompt.System,
		User:   prompt.User,
		Parameters: llm.PublicModelParameters{
			Temperature: prompt.Parameters.Temperature,
			MaxTokens:   prompt.Parameters.MaxTokens,
		},
	}

	internalChan, err := a.adapter.Stream(ctx, internalPrompt)
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

func (a *coreModelProviderAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return a.adapter.Embeddings(ctx, texts)
}

// coreLLMAdapter provides minimal adapter for core LLM interface
type coreLLMAdapter struct {
	adapter *llm.PublicLLMAdapterWrapper
}

func (a *coreLLMAdapter) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	return a.adapter.Complete(ctx, systemPrompt, userPrompt)
}

// directCoreLLMAdapter provides fallback for external ModelProvider implementations
type directCoreLLMAdapter struct {
	provider ModelProvider
}

func (a *directCoreLLMAdapter) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	resp, err := a.provider.Call(ctx, Prompt{
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

// NewLLMProvider creates a ModelProvider from AgentLLMConfig with environment variable support
func NewLLMProvider(config AgentLLMConfig) (ModelProvider, error) {
	providerConfig := LLMProviderConfig{
		Type:        config.Provider,
		Model:       config.Model,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
		HTTPTimeout: TimeoutFromSeconds(config.TimeoutSeconds),
	}

	// Read environment variables based on provider type
	switch config.Provider {
	case "openai":
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			providerConfig.APIKey = apiKey
		}
	case "azure", "azureopenai":
		if apiKey := os.Getenv("AZURE_OPENAI_API_KEY"); apiKey != "" {
			providerConfig.APIKey = apiKey
		}
		if endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT"); endpoint != "" {
			providerConfig.Endpoint = endpoint
		}
		if deployment := os.Getenv("AZURE_OPENAI_DEPLOYMENT"); deployment != "" {
			providerConfig.ChatDeployment = deployment
			providerConfig.EmbeddingDeployment = deployment
		}
	case "ollama":
		// Ollama typically doesn't need API keys, but we could support custom base URL
		if baseURL := os.Getenv("OLLAMA_BASE_URL"); baseURL != "" {
			providerConfig.BaseURL = baseURL
		} else {
			providerConfig.BaseURL = "http://localhost:11434"
		}
	case "openrouter":
		if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
			providerConfig.APIKey = apiKey
		}
		if baseURL := os.Getenv("OPENROUTER_BASE_URL"); baseURL != "" {
			providerConfig.BaseURL = baseURL
		} else {
			providerConfig.BaseURL = "https://openrouter.ai/api/v1"
		}
		if siteURL := os.Getenv("OPENROUTER_SITE_URL"); siteURL != "" {
			providerConfig.SiteURL = siteURL
		}
		if siteName := os.Getenv("OPENROUTER_SITE_NAME"); siteName != "" {
			providerConfig.SiteName = siteName
		}
	case "huggingface":
		if apiKey := os.Getenv("HUGGINGFACE_API_KEY"); apiKey != "" {
			providerConfig.APIKey = apiKey
		}
		if baseURL := os.Getenv("HUGGINGFACE_BASE_URL"); baseURL != "" {
			providerConfig.BaseURL = baseURL
		}
		if apiType := os.Getenv("HUGGINGFACE_API_TYPE"); apiType != "" {
			providerConfig.HFAPIType = apiType
		} else {
			providerConfig.HFAPIType = "inference" // Default to Inference API
		}
		// Optional HF-specific parameters from environment
		if waitForModel := os.Getenv("HUGGINGFACE_WAIT_FOR_MODEL"); waitForModel == "true" {
			providerConfig.HFWaitForModel = true
		}
		if useCache := os.Getenv("HUGGINGFACE_USE_CACHE"); useCache == "false" {
			providerConfig.HFUseCache = false
		} else {
			providerConfig.HFUseCache = true // Default to true
		}
	}

	return NewModelProviderFromConfig(providerConfig)
}

// =============================================================================
// TYPE ALIASES FOR BACKWARD COMPATIBILITY
// =============================================================================

// LLMConfig is an alias for AgentLLMConfig for backward compatibility
type LLMConfig = AgentLLMConfig
