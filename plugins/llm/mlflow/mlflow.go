// Package mlflow provides an MLFlow AI Gateway LLM provider plugin for AgenticGoKit.
package mlflow

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

// factory creates an MLFlow Gateway provider from configuration
func factory(cfg core.LLMProviderConfig) (core.ModelProvider, error) {
	// Build MLFlow config from core config
	mlflowConfig := llm.MLFlowGatewayConfig{
		BaseURL:          cfg.BaseURL,
		ChatRoute:        cfg.MLFlowChatRoute,
		EmbeddingsRoute:  cfg.MLFlowEmbeddingsRoute,
		CompletionsRoute: cfg.MLFlowCompletionsRoute,
		Model:            cfg.Model,
		APIKey:           cfg.APIKey,
		ExtraHeaders:     cfg.MLFlowExtraHeaders,
		MaxTokens:        cfg.MaxTokens,
		Temperature:      float32(cfg.Temperature),
		TopP:             float32(cfg.MLFlowTopP),
		Stop:             cfg.MLFlowStop,
		MaxRetries:       cfg.MLFlowMaxRetries,
		RetryDelay:       cfg.MLFlowRetryDelay,
		HTTPTimeout:      cfg.HTTPTimeout,
	}

	wrapper, err := llm.NewMLFlowGatewayAdapterWrapped(mlflowConfig)
	if err != nil {
		return nil, err
	}
	return &providerAdapter{adapter: llm.NewPublicProviderAdapter(wrapper)}, nil
}

func init() {
	core.RegisterModelProviderFactory("mlflow", factory)
}
