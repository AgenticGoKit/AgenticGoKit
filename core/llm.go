// Package core provides public LLM interfaces and essential types for AgentFlow.
package core

import "context"

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

// Helper functions for creating parameter pointers
func FloatPtr(f float32) *float32 {
	return &f
}

func Int32Ptr(i int32) *int32 {
	return &i
}
