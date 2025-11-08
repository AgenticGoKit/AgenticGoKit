package huggingface

import (
	"context"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/internal/llm"
)

// providerAdapter adapts internal llm.PublicProviderAdapter to core.ModelProvider
type providerAdapter struct {
	adapter *llm.PublicProviderAdapter
}

func (a *providerAdapter) Call(ctx context.Context, prompt core.Prompt) (core.Response, error) {
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
		return core.Response{}, err
	}

	return core.Response{
		Content: resp.Content,
		Usage: core.UsageStats{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		FinishReason: resp.FinishReason,
	}, nil
}

func (a *providerAdapter) Stream(ctx context.Context, prompt core.Prompt) (<-chan core.Token, error) {
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

	publicChan := make(chan core.Token)
	go func() {
		defer close(publicChan)
		for token := range internalChan {
			publicChan <- core.Token{Content: token.Content, Error: token.Error}
		}
	}()

	return publicChan, nil
}

func (a *providerAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return a.adapter.Embeddings(ctx, texts)
}

func factory(cfg core.LLMProviderConfig) (core.ModelProvider, error) {
	// Set default API type if not specified
	apiType := cfg.HFAPIType
	if apiType == "" {
		apiType = "inference" // Default to Inference API
	}

	// Set default base URL based on API type
	baseURL := cfg.BaseURL
	if baseURL == "" {
		switch apiType {
		case "inference":
			// New router-based endpoint (as of late 2024)
			baseURL = "https://router.huggingface.co"
		case "chat":
			baseURL = "https://router.huggingface.co"
			// For endpoint and tgi, base URL must be provided by user
		}
	}

	wrapper, err := llm.NewHuggingFaceAdapterWrapped(
		cfg.APIKey,
		cfg.Model,
		baseURL,
		apiType,
		cfg.MaxTokens,
		float32(cfg.Temperature),
		cfg.HFWaitForModel,
		cfg.HFUseCache,
		cfg.HFDoSample,
		float32(cfg.HFTopP),
		cfg.HFTopK,
		float32(cfg.HFRepetitionPenalty),
		cfg.HFStopSequences,
	)
	if err != nil {
		return nil, err
	}

	return &providerAdapter{adapter: llm.NewPublicProviderAdapter(wrapper)}, nil
}

func init() {
	core.RegisterModelProviderFactory("huggingface", factory)
}
