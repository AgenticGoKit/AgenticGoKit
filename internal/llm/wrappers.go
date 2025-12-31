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

	// vLLM-specific fields
	VLLMTopK              int      `json:"vllm_top_k,omitempty" toml:"vllm_top_k,omitempty"`
	VLLMTopP              float64  `json:"vllm_top_p,omitempty" toml:"vllm_top_p,omitempty"`
	VLLMMinP              float64  `json:"vllm_min_p,omitempty" toml:"vllm_min_p,omitempty"`
	VLLMPresencePenalty   float64  `json:"vllm_presence_penalty,omitempty" toml:"vllm_presence_penalty,omitempty"`
	VLLMFrequencyPenalty  float64  `json:"vllm_frequency_penalty,omitempty" toml:"vllm_frequency_penalty,omitempty"`
	VLLMRepetitionPenalty float64  `json:"vllm_repetition_penalty,omitempty" toml:"vllm_repetition_penalty,omitempty"`
	VLLMBestOf            int      `json:"vllm_best_of,omitempty" toml:"vllm_best_of,omitempty"`
	VLLMUseBeamSearch     bool     `json:"vllm_use_beam_search,omitempty" toml:"vllm_use_beam_search,omitempty"`
	VLLMLengthPenalty     float64  `json:"vllm_length_penalty,omitempty" toml:"vllm_length_penalty,omitempty"`
	VLLMStopTokenIds      []int    `json:"vllm_stop_token_ids,omitempty" toml:"vllm_stop_token_ids,omitempty"`
	VLLMSkipSpecialTokens bool     `json:"vllm_skip_special_tokens,omitempty" toml:"vllm_skip_special_tokens,omitempty"`
	VLLMIgnoreEOS         bool     `json:"vllm_ignore_eos,omitempty" toml:"vllm_ignore_eos,omitempty"`
	VLLMStop              []string `json:"vllm_stop,omitempty" toml:"vllm_stop,omitempty"`

	// MLFlow Gateway-specific fields
	MLFlowChatRoute        string            `json:"mlflow_chat_route,omitempty" toml:"mlflow_chat_route,omitempty"`
	MLFlowEmbeddingsRoute  string            `json:"mlflow_embeddings_route,omitempty" toml:"mlflow_embeddings_route,omitempty"`
	MLFlowCompletionsRoute string            `json:"mlflow_completions_route,omitempty" toml:"mlflow_completions_route,omitempty"`
	MLFlowExtraHeaders     map[string]string `json:"mlflow_extra_headers,omitempty" toml:"mlflow_extra_headers,omitempty"`
	MLFlowMaxRetries       int               `json:"mlflow_max_retries,omitempty" toml:"mlflow_max_retries,omitempty"`
	MLFlowRetryDelay       time.Duration     `json:"mlflow_retry_delay,omitempty" toml:"mlflow_retry_delay,omitempty"`
	MLFlowTopP             float64           `json:"mlflow_top_p,omitempty" toml:"mlflow_top_p,omitempty"`
	MLFlowStop             []string          `json:"mlflow_stop,omitempty" toml:"mlflow_stop,omitempty"`

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

func NewOpenRouterAdapterWrapped(
	apiKey, model, baseURL string,
	maxTokens int,
	temperature float32,
	siteURL, siteName string,
) (PublicModelProvider, error) {
	adapter, err := NewOpenRouterAdapter(
		apiKey,
		model,
		baseURL,
		maxTokens,
		temperature,
		siteURL,
		siteName,
	)
	if err != nil {
		return nil, err
	}

	return NewModelProviderWrapper(adapter), nil
}

func NewHuggingFaceAdapterWrapped(
	apiKey, model, baseURL, apiType string,
	maxTokens int,
	temperature float32,
	waitForModel, useCache, doSample bool,
	topP float32,
	topK int,
	repetitionPenalty float32,
	stopSequences []string,
) (PublicModelProvider, error) {
	adapter, err := NewHuggingFaceAdapter(
		apiKey,
		model,
		baseURL,
		HFAPIType(apiType),
		maxTokens,
		temperature,
		HFAdapterOptions{
			TopP:              topP,
			TopK:              topK,
			DoSample:          doSample,
			WaitForModel:      waitForModel,
			UseCache:          useCache,
			StopSequences:     stopSequences,
			RepetitionPenalty: repetitionPenalty,
		},
	)
	if err != nil {
		return nil, err
	}

	return NewModelProviderWrapper(adapter), nil
}

// NewVLLMAdapterWrapped creates a wrapped vLLM adapter
func NewVLLMAdapterWrapped(config VLLMConfig) (PublicModelProvider, error) {
	adapter, err := NewVLLMAdapter(config)
	if err != nil {
		return nil, err
	}

	return NewModelProviderWrapper(adapter), nil
}

// NewMLFlowGatewayAdapterWrapped creates a wrapped MLFlow AI Gateway adapter
func NewMLFlowGatewayAdapterWrapped(config MLFlowGatewayConfig) (PublicModelProvider, error) {
	adapter, err := NewMLFlowGatewayAdapter(config)
	if err != nil {
		return nil, err
	}

	return NewModelProviderWrapper(adapter), nil
}

// NewBentoMLAdapterWrapped creates a wrapped BentoML adapter
func NewBentoMLAdapterWrapped(config BentoMLConfig) (PublicModelProvider, error) {
	adapter, err := NewBentoMLAdapter(config)
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
		Temperature:         float32(config.Temperature), // Convert float64 to float32 at boundary
		Endpoint:            config.Endpoint,
		ChatDeployment:      config.ChatDeployment,
		EmbeddingDeployment: config.EmbeddingDeployment,
		BaseURL:             config.BaseURL,
		SiteURL:             config.SiteURL,
		SiteName:            config.SiteName,
		HFAPIType:           config.HFAPIType,
		HFWaitForModel:      config.HFWaitForModel,
		HFUseCache:          config.HFUseCache,
		HFTopP:              float32(config.HFTopP),
		HFTopK:              config.HFTopK,
		HFDoSample:          config.HFDoSample,
		HFStopSequences:     config.HFStopSequences,
		HFRepetitionPenalty: float32(config.HFRepetitionPenalty),
		HTTPTimeout:         config.HTTPTimeout,
		// vLLM fields
		VLLMTopK:              config.VLLMTopK,
		VLLMTopP:              float32(config.VLLMTopP),
		VLLMMinP:              float32(config.VLLMMinP),
		VLLMPresencePenalty:   float32(config.VLLMPresencePenalty),
		VLLMFrequencyPenalty:  float32(config.VLLMFrequencyPenalty),
		VLLMRepetitionPenalty: float32(config.VLLMRepetitionPenalty),
		VLLMBestOf:            config.VLLMBestOf,
		VLLMUseBeamSearch:     config.VLLMUseBeamSearch,
		VLLMLengthPenalty:     float32(config.VLLMLengthPenalty),
		VLLMStopTokenIds:      config.VLLMStopTokenIds,
		VLLMSkipSpecialTokens: config.VLLMSkipSpecialTokens,
		VLLMIgnoreEOS:         config.VLLMIgnoreEOS,
		VLLMStop:              config.VLLMStop,
		// MLFlow fields
		MLFlowChatRoute:        config.MLFlowChatRoute,
		MLFlowEmbeddingsRoute:  config.MLFlowEmbeddingsRoute,
		MLFlowCompletionsRoute: config.MLFlowCompletionsRoute,
		MLFlowExtraHeaders:     config.MLFlowExtraHeaders,
		MLFlowMaxRetries:       config.MLFlowMaxRetries,
		MLFlowRetryDelay:       config.MLFlowRetryDelay,
		MLFlowTopP:             float32(config.MLFlowTopP),
		MLFlowStop:             config.MLFlowStop,
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
