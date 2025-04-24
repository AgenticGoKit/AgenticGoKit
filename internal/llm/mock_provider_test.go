package llm

import (
	"context"
	"sync"
)

// MockModelProvider is a mock implementation of the ModelProvider interface for testing.
// It allows setting expected return values and errors for each method.
// It is safe for concurrent use in tests.
type MockModelProvider struct {
	mu sync.RWMutex

	// Call expectations
	CallFunc      func(ctx context.Context, prompt Prompt) (Response, error)
	CallPromptArg Prompt
	CallResp      Response
	CallErr       error

	// Stream expectations
	StreamFunc      func(ctx context.Context, prompt Prompt) (<-chan Token, error)
	StreamPromptArg Prompt
	StreamChan      chan Token // Pre-filled channel to return
	StreamErr       error

	// Embeddings expectations
	EmbeddingsFunc     func(ctx context.Context, texts []string) ([][]float64, error)
	EmbeddingsTextsArg []string
	EmbeddingsResp     [][]float64
	EmbeddingsErr      error
}

// NewMockModelProvider creates a new mock provider.
func NewMockModelProvider() *MockModelProvider {
	return &MockModelProvider{}
}

// --- Call Method ---

// SetCallExpectation sets the expected response and error for the Call method.
func (m *MockModelProvider) SetCallExpectation(resp Response, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CallResp = resp
	m.CallErr = err
	m.CallFunc = nil // Clear any custom function
}

// SetCallFunc sets a custom function to handle the Call method.
func (m *MockModelProvider) SetCallFunc(f func(ctx context.Context, prompt Prompt) (Response, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CallFunc = f
}

func (m *MockModelProvider) Call(ctx context.Context, prompt Prompt) (Response, error) {
	m.mu.Lock() // Lock to record args and check func
	m.CallPromptArg = prompt
	callFunc := m.CallFunc
	resp := m.CallResp
	err := m.CallErr
	m.mu.Unlock() // Unlock before potentially calling custom func

	if callFunc != nil {
		return callFunc(ctx, prompt)
	}
	return resp, err
}

// --- Stream Method ---

// SetStreamExpectation sets the expected channel and error for the Stream method.
// The provided channel should be pre-filled with tokens and closed appropriately by the test.
func (m *MockModelProvider) SetStreamExpectation(streamChan chan Token, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StreamChan = streamChan
	m.StreamErr = err
	m.StreamFunc = nil // Clear any custom function
}

// SetStreamFunc sets a custom function to handle the Stream method.
func (m *MockModelProvider) SetStreamFunc(f func(ctx context.Context, prompt Prompt) (<-chan Token, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StreamFunc = f
}

func (m *MockModelProvider) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	m.mu.Lock() // Lock to record args and check func
	m.StreamPromptArg = prompt
	streamFunc := m.StreamFunc
	streamChan := m.StreamChan
	err := m.StreamErr
	m.mu.Unlock() // Unlock before potentially calling custom func

	if streamFunc != nil {
		return streamFunc(ctx, prompt)
	}
	// Return a nil channel if the expected channel is nil and no error is set,
	// otherwise return the expected channel/error.
	if streamChan == nil && err == nil {
		// Return a closed nil channel if no expectation set, mimicking no output
		return nil, nil
	}
	return streamChan, err
}

// --- Embeddings Method ---

// SetEmbeddingsExpectation sets the expected response and error for the Embeddings method.
func (m *MockModelProvider) SetEmbeddingsExpectation(resp [][]float64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EmbeddingsResp = resp
	m.EmbeddingsErr = err
	m.EmbeddingsFunc = nil // Clear any custom function
}

// SetEmbeddingsFunc sets a custom function to handle the Embeddings method.
func (m *MockModelProvider) SetEmbeddingsFunc(f func(ctx context.Context, texts []string) ([][]float64, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EmbeddingsFunc = f
}

func (m *MockModelProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	m.mu.Lock() // Lock to record args and check func
	m.EmbeddingsTextsArg = texts
	embeddingsFunc := m.EmbeddingsFunc
	resp := m.EmbeddingsResp
	err := m.EmbeddingsErr
	m.mu.Unlock() // Unlock before potentially calling custom func

	if embeddingsFunc != nil {
		return embeddingsFunc(ctx, texts)
	}
	return resp, err
}

// --- Helper Methods for Assertions ---

// GetCallPromptArg returns the prompt argument captured by the last Call invocation.
func (m *MockModelProvider) GetCallPromptArg() Prompt {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.CallPromptArg
}

// GetStreamPromptArg returns the prompt argument captured by the last Stream invocation.
func (m *MockModelProvider) GetStreamPromptArg() Prompt {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.StreamPromptArg
}

// GetEmbeddingsTextsArg returns the texts argument captured by the last Embeddings invocation.
func (m *MockModelProvider) GetEmbeddingsTextsArg() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return a copy to prevent race conditions if the test modifies the slice later
	textsCopy := make([]string, len(m.EmbeddingsTextsArg))
	copy(textsCopy, m.EmbeddingsTextsArg)
	return textsCopy
}
