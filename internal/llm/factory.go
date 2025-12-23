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
	ProviderTypeOpenAI        ProviderType = "openai"
	ProviderTypeAzureOpenAI   ProviderType = "azure"
	ProviderTypeOllama        ProviderType = "ollama"
	ProviderTypeOpenRouter    ProviderType = "openrouter"
	ProviderTypeHuggingFace   ProviderType = "huggingface"
	ProviderTypeVLLM          ProviderType = "vllm"
	ProviderTypeMLFlowGateway ProviderType = "mlflow"
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

	// OpenRouter-specific fields
	SiteURL  string `json:"site_url,omitempty" toml:"site_url,omitempty"`
	SiteName string `json:"site_name,omitempty" toml:"site_name,omitempty"`

	// HuggingFace-specific fields
	HFAPIType           string   `json:"hf_api_type,omitempty" toml:"hf_api_type,omitempty"`
	HFWaitForModel      bool     `json:"hf_wait_for_model,omitempty" toml:"hf_wait_for_model,omitempty"`
	HFUseCache          bool     `json:"hf_use_cache,omitempty" toml:"hf_use_cache,omitempty"`
	HFTopP              float32  `json:"hf_top_p,omitempty" toml:"hf_top_p,omitempty"`
	HFTopK              int      `json:"hf_top_k,omitempty" toml:"hf_top_k,omitempty"`
	HFDoSample          bool     `json:"hf_do_sample,omitempty" toml:"hf_do_sample,omitempty"`
	HFStopSequences     []string `json:"hf_stop_sequences,omitempty" toml:"hf_stop_sequences,omitempty"`
	HFRepetitionPenalty float32  `json:"hf_repetition_penalty,omitempty" toml:"hf_repetition_penalty,omitempty"`

	// vLLM-specific fields
	VLLMTopK              int      `json:"vllm_top_k,omitempty" toml:"vllm_top_k,omitempty"`
	VLLMTopP              float32  `json:"vllm_top_p,omitempty" toml:"vllm_top_p,omitempty"`
	VLLMMinP              float32  `json:"vllm_min_p,omitempty" toml:"vllm_min_p,omitempty"`
	VLLMPresencePenalty   float32  `json:"vllm_presence_penalty,omitempty" toml:"vllm_presence_penalty,omitempty"`
	VLLMFrequencyPenalty  float32  `json:"vllm_frequency_penalty,omitempty" toml:"vllm_frequency_penalty,omitempty"`
	VLLMRepetitionPenalty float32  `json:"vllm_repetition_penalty,omitempty" toml:"vllm_repetition_penalty,omitempty"`
	VLLMBestOf            int      `json:"vllm_best_of,omitempty" toml:"vllm_best_of,omitempty"`
	VLLMUseBeamSearch     bool     `json:"vllm_use_beam_search,omitempty" toml:"vllm_use_beam_search,omitempty"`
	VLLMLengthPenalty     float32  `json:"vllm_length_penalty,omitempty" toml:"vllm_length_penalty,omitempty"`
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
	MLFlowTopP             float32           `json:"mlflow_top_p,omitempty" toml:"mlflow_top_p,omitempty"`
	MLFlowStop             []string          `json:"mlflow_stop,omitempty" toml:"mlflow_stop,omitempty"`

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
	case ProviderTypeOpenRouter:
		return f.createOpenRouterProvider(config)
	case ProviderTypeHuggingFace:
		return f.createHuggingFaceProvider(config)
	case ProviderTypeVLLM:
		return f.createVLLMProvider(config)
	case ProviderTypeMLFlowGateway:
		return f.createMLFlowGatewayProvider(config)
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

// createOpenRouterProvider creates an OpenRouter provider
func (f *ProviderFactory) createOpenRouterProvider(config ProviderConfig) (ModelProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for OpenRouter provider")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	model := config.Model
	if model == "" {
		model = "openai/gpt-3.5-turbo" // Default model
	}

	return NewOpenRouterAdapter(
		config.APIKey,
		model,
		baseURL,
		config.MaxTokens,
		config.Temperature,
		config.SiteURL,
		config.SiteName,
	)
}

// createHuggingFaceProvider creates a Hugging Face provider
func (f *ProviderFactory) createHuggingFaceProvider(config ProviderConfig) (ModelProvider, error) {
	// API key validation (not required for local TGI)
	apiType := HFAPIType(config.HFAPIType)
	if apiType == "" {
		apiType = HFAPITypeInference
	}

	if apiType != HFAPITypeTGI && config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Hugging Face provider (except for local TGI)")
	}

	model := config.Model
	if model == "" {
		model = "gpt2" // Default model for free tier
	}

	baseURL := config.BaseURL
	// Validate baseURL requirement for non-inference API types
	if (apiType == HFAPITypeEndpoint || apiType == HFAPITypeTGI || apiType == HFAPITypeChat) && baseURL == "" {
		return nil, fmt.Errorf("base URL is required for API type: %s", apiType)
	}

	options := HFAdapterOptions{
		TopP:              config.HFTopP,
		TopK:              config.HFTopK,
		DoSample:          config.HFDoSample,
		WaitForModel:      config.HFWaitForModel,
		UseCache:          config.HFUseCache,
		StopSequences:     config.HFStopSequences,
		RepetitionPenalty: config.HFRepetitionPenalty,
		HTTPTimeout:       config.HTTPTimeout,
	}

	return NewHuggingFaceAdapter(
		config.APIKey,
		model,
		baseURL,
		apiType,
		config.MaxTokens,
		config.Temperature,
		options,
	)
}

// createVLLMProvider creates a vLLM provider
func (f *ProviderFactory) createVLLMProvider(config ProviderConfig) (ModelProvider, error) {
	if config.Model == "" {
		return nil, fmt.Errorf("model is required for vLLM provider")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	vllmConfig := VLLMConfig{
		BaseURL:           baseURL,
		APIKey:            config.APIKey,
		Model:             config.Model,
		MaxTokens:         config.MaxTokens,
		Temperature:       config.Temperature,
		TopK:              config.VLLMTopK,
		TopP:              config.VLLMTopP,
		MinP:              config.VLLMMinP,
		PresencePenalty:   config.VLLMPresencePenalty,
		FrequencyPenalty:  config.VLLMFrequencyPenalty,
		RepetitionPenalty: config.VLLMRepetitionPenalty,
		BestOf:            config.VLLMBestOf,
		UseBeamSearch:     config.VLLMUseBeamSearch,
		LengthPenalty:     config.VLLMLengthPenalty,
		StopTokenIds:      config.VLLMStopTokenIds,
		SkipSpecialTokens: config.VLLMSkipSpecialTokens,
		IgnoreEOS:         config.VLLMIgnoreEOS,
		Stop:              config.VLLMStop,
		HTTPTimeout:       config.HTTPTimeout,
	}

	return NewVLLMAdapter(vllmConfig)
}

// createMLFlowGatewayProvider creates an MLFlow AI Gateway provider
func (f *ProviderFactory) createMLFlowGatewayProvider(config ProviderConfig) (ModelProvider, error) {
	if config.MLFlowChatRoute == "" {
		return nil, fmt.Errorf("mlflow_chat_route is required for MLFlow Gateway provider")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:5001"
	}

	mlflowConfig := MLFlowGatewayConfig{
		BaseURL:          baseURL,
		ChatRoute:        config.MLFlowChatRoute,
		EmbeddingsRoute:  config.MLFlowEmbeddingsRoute,
		CompletionsRoute: config.MLFlowCompletionsRoute,
		Model:            config.Model,
		APIKey:           config.APIKey,
		ExtraHeaders:     config.MLFlowExtraHeaders,
		MaxTokens:        config.MaxTokens,
		Temperature:      config.Temperature,
		TopP:             config.MLFlowTopP,
		Stop:             config.MLFlowStop,
		MaxRetries:       config.MLFlowMaxRetries,
		RetryDelay:       config.MLFlowRetryDelay,
		HTTPTimeout:      config.HTTPTimeout,
	}

	return NewMLFlowGatewayAdapter(mlflowConfig)
}

// DefaultFactory is a global factory instance for convenience
var DefaultFactory = NewProviderFactory()

// CreateProviderFromConfig is a convenience function that uses the default factory
func CreateProviderFromConfig(config ProviderConfig) (ModelProvider, error) {
	return DefaultFactory.CreateProvider(config)
}
