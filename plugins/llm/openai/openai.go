package openai

import (
	"context"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/agenticgokit/agenticgokit/internal/llm"
)

// providerAdapter adapts internal llm.PublicProviderAdapter to core.ModelProvider
// (duplicated minimal adapter to avoid relying on unexported core types)
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
	// Map config to internal wrapper directly
	wrapper, err := llm.NewOpenAIAdapterWrapped(cfg.APIKey, cfg.Model, cfg.MaxTokens, float32(cfg.Temperature))
	if err != nil {
		return nil, err
	}
	return &providerAdapter{adapter: llm.NewPublicProviderAdapter(wrapper)}, nil
}

func init() {
	core.RegisterModelProviderFactory("openai", factory)
}

