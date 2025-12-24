// Package bentoml provides a BentoML LLM provider plugin for AgenticGoKit.
package bentoml

import (
	"context"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/agenticgokit/agenticgokit/internal/llm"
)

// providerAdapter adapts internal llm.PublicProviderAdapter to core.ModelProvider
type providerAdapter struct {
	adapter *llm.PublicProviderAdapter
}

func (a *providerAdapter) Call(ctx context.Context, p core.Prompt) (core.Response, error) {
	ip := llm.PublicPrompt{
		System: p.System,
		User:   p.User,
		Parameters: llm.PublicModelParameters{
			Temperature: p.Parameters.Temperature,
			MaxTokens:   p.Parameters.MaxTokens,
		},
	}
	resp, err := a.adapter.Call(ctx, ip)
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

func (a *providerAdapter) Stream(ctx context.Context, p core.Prompt) (<-chan core.Token, error) {
	ip := llm.PublicPrompt{
		System: p.System,
		User:   p.User,
		Parameters: llm.PublicModelParameters{
			Temperature: p.Parameters.Temperature,
			MaxTokens:   p.Parameters.MaxTokens,
		},
	}
	ich, err := a.adapter.Stream(ctx, ip)
	if err != nil {
		return nil, err
	}
	och := make(chan core.Token)
	go func() {
		defer close(och)
		for t := range ich {
			och <- core.Token{Content: t.Content, Error: t.Error}
		}
	}()
	return och, nil
}

func (a *providerAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return a.adapter.Embeddings(ctx, texts)
}

// factory creates a BentoML provider from configuration
func factory(cfg core.LLMProviderConfig) (core.ModelProvider, error) {
	// Build BentoML config from core config
	bentomlConfig := llm.BentoMLConfig{
		BaseURL:          cfg.BaseURL,
		APIKey:           cfg.APIKey,
		Model:            cfg.Model,
		MaxTokens:        cfg.MaxTokens,
		Temperature:      float32(cfg.Temperature),
		TopP:             float32(cfg.BentoMLTopP),
		TopK:             cfg.BentoMLTopK,
		PresencePenalty:  float32(cfg.BentoMLPresencePenalty),
		FrequencyPenalty: float32(cfg.BentoMLFrequencyPenalty),
		Stop:             cfg.BentoMLStop,
		ServiceName:      cfg.BentoMLServiceName,
		Runners:          cfg.BentoMLRunners,
		ExtraHeaders:     cfg.BentoMLExtraHeaders,
		MaxRetries:       cfg.BentoMLMaxRetries,
		RetryDelay:       cfg.BentoMLRetryDelay,
		HTTPTimeout:      cfg.HTTPTimeout,
	}

	wrapper, err := llm.NewBentoMLAdapterWrapped(bentomlConfig)
	if err != nil {
		return nil, err
	}
	return &providerAdapter{adapter: llm.NewPublicProviderAdapter(wrapper)}, nil
}

func init() {
	core.RegisterModelProviderFactory("bentoml", factory)
}
