package llm

import "context"

// LLMAdapter defines a simplified interface for LLM interaction
// Following Azure best practices for interface segregation principle
type LLMAdapter interface {
	// Complete sends a prompt to the LLM and returns the completion
	Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
}

// ModelProviderAdapter adapts a ModelProvider to the LLMAdapter interface
type ModelProviderAdapter struct {
	provider ModelProvider
}

// NewModelProviderAdapter creates a new adapter that wraps a ModelProvider
func NewModelProviderAdapter(provider ModelProvider) *ModelProviderAdapter {
	return &ModelProviderAdapter{provider: provider}
}

// Complete implements the LLMAdapter interface
func (a *ModelProviderAdapter) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	resp, err := a.provider.Call(ctx, Prompt{
		System: systemPrompt,
		User:   userPrompt,
		Parameters: ModelParameters{
			Temperature: floatPtr(0.7),
			MaxTokens:   int32Ptr(2000),
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// Helper functions
func floatPtr(f float32) *float32 {
	return &f
}

func int32Ptr(i int32) *int32 {
	return &i
}
