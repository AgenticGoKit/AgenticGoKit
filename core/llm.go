// Package core provides public LLM interfaces and types for AgentFlow.
package core

import "context"

// ModelParameters holds common configuration options for language model calls.
type ModelParameters struct {
	Temperature *float32 // Sampling temperature. nil uses the provider's default.
	MaxTokens   *int32   // Max tokens to generate. nil uses the provider's default.
	// TODO: Add TopP, StopSequences, PresencePenalty, FrequencyPenalty etc.
}

// Prompt represents the input to a language model call.
type Prompt struct {
	// System message sets the context or instructions for the model.
	System string
	// User message is the primary input or question.
	User string
	// Parameters specify model configuration for this call.
	Parameters ModelParameters
	// TODO: Add fields for message history, function calls/definitions
}

// UsageStats contains token usage information for a model call.
type UsageStats struct {
	PromptTokens     int // Tokens in the input prompt.
	CompletionTokens int // Tokens generated in the response.
	TotalTokens      int // Total tokens processed.
}

// Response represents the output from a non-streaming language model call.
type Response struct {
	// Content is the primary text response from the model.
	Content string
	// Usage provides token usage statistics for the call.
	Usage UsageStats
	// FinishReason indicates why the model stopped generating tokens (e.g., "stop", "length", "content_filter").
	FinishReason string
	// TODO: Add fields for function call results, log probabilities, etc.
}

// Token represents a single token streamed from a language model.
type Token struct {
	// Content is the text chunk of the token.
	Content string
	// Error holds any error that occurred during streaming for this token or subsequent ones.
	// If non-nil, the stream should be considered terminated.
	Error error
	// TODO: Add fields for token index, log probabilities, finish reason (on last token), usage (on last token) if available.
}

// ModelProvider defines the interface for interacting with different language model backends.
// Implementations should be thread-safe.
type ModelProvider interface {
	// Call sends a prompt to the model and returns a complete response.
	// It blocks until the full response is generated or an error occurs.
	Call(ctx context.Context, prompt Prompt) (Response, error)

	// Stream sends a prompt to the model and returns a channel of tokens.
	// The channel is closed when the stream is complete or an error occurs.
	// Errors are sent as Token.Error in the stream.
	Stream(ctx context.Context, prompt Prompt) (<-chan Token, error)

	// Embeddings generates vector embeddings for the provided texts.
	// Returns a slice of float64 slices, one per input text.
	Embeddings(ctx context.Context, texts []string) ([][]float64, error)
}

// LLMAdapter defines a simplified interface for LLM interaction
// Following Azure best practices for interface segregation principle
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
