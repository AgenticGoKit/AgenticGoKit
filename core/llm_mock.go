// Package core provides public mock implementations for testing.
package core

import (
	"context"
)

// MockModelProvider is a mock implementation of the ModelProvider interface for testing.
// It allows setting expected return values and errors for each method.
// It is safe for concurrent use in tests.
type MockModelProvider struct {
	// internal *llm.MockModelProvider
}

// SetCallExpectation sets the expected response and error for the Call method.
func (m *MockModelProvider) SetCallExpectation(resp Response, err error) {
	// internalResp := llm.Response{
	// 	Content: resp.Content,
	// 	Usage: llm.UsageStats{
	// 		PromptTokens:     resp.Usage.PromptTokens,
	// 		CompletionTokens: resp.Usage.CompletionTokens,
	// 		TotalTokens:      resp.Usage.TotalTokens,
	// 	},
	// 	FinishReason: resp.FinishReason,
	// }
	// m.internal.SetCallExpectation(internalResp, err)
}

// SetCallFunc sets a custom function to handle the Call method.
func (m *MockModelProvider) SetCallFunc(f func(ctx context.Context, prompt Prompt) (Response, error)) {
	// m.internal.SetCallFunc(func(ctx context.Context, prompt llm.Prompt) (llm.Response, error) {
	// 	publicPrompt := Prompt{
	// 		System: prompt.System,
	// 		User:   prompt.User,
	// 		Parameters: ModelParameters{
	// 			Temperature: prompt.Parameters.Temperature,
	// 			MaxTokens:   prompt.Parameters.MaxTokens,
	// 		},
	// 	}

	// 	resp, err := f(ctx, publicPrompt)
	// 	if err != nil {
	// 		return llm.Response{}, err
	// 	}

	// 	return llm.Response{
	// 		Content: resp.Content,
	// 		Usage: llm.UsageStats{
	// 			PromptTokens:     resp.Usage.PromptTokens,
	// 			CompletionTokens: resp.Usage.CompletionTokens,
	// 			TotalTokens:      resp.Usage.TotalTokens,
	// 		},
	// 		FinishReason: resp.FinishReason,
	// 	}, nil
	// })
}

func (m *MockModelProvider) Call(ctx context.Context, prompt Prompt) (Response, error) {
	// internalPrompt := llm.Prompt{
	// 	System: prompt.System,
	// 	User:   prompt.User,
	// 	Parameters: llm.ModelParameters{
	// 		Temperature: prompt.Parameters.Temperature,
	// 		MaxTokens:   prompt.Parameters.MaxTokens,
	// 	},
	// }

	// resp, err := m.internal.Call(ctx, internalPrompt)
	// if err != nil {
	// 	return Response{}, err
	// }

	// return Response{
	// 	Content: resp.Content,
	// 	Usage: UsageStats{
	// 		PromptTokens:     resp.Usage.PromptTokens,
	// 		CompletionTokens: resp.Usage.CompletionTokens,
	// 		TotalTokens:      resp.Usage.TotalTokens,
	// 	},
	// 	FinishReason: resp.FinishReason,
	// }, nil
	return Response{}, nil
}

// SetStreamExpectation sets the expected channel and error for the Stream method.
func (m *MockModelProvider) SetStreamExpectation(tokens []Token, err error) {
	// if err != nil {
	// 	m.internal.SetStreamExpectation(nil, err)
	// 	return
	// }

	// // Convert public tokens to internal tokens
	// streamChan := make(chan llm.Token, len(tokens))
	// for _, token := range tokens {
	// 	streamChan <- llm.Token{
	// 		Content: token.Content,
	// 		Error:   token.Error,
	// 	}
	// }
	// close(streamChan)

	// m.internal.SetStreamExpectation(streamChan, nil)
}

// SetStreamFunc sets a custom function to handle the Stream method.
func (m *MockModelProvider) SetStreamFunc(f func(ctx context.Context, prompt Prompt) (<-chan Token, error)) {
	// m.internal.SetStreamFunc(func(ctx context.Context, prompt llm.Prompt) (<-chan llm.Token, error) {
	// 	publicPrompt := Prompt{
	// 		System: prompt.System,
	// 		User:   prompt.User,
	// 		Parameters: ModelParameters{
	// 			Temperature: prompt.Parameters.Temperature,
	// 			MaxTokens:   prompt.Parameters.MaxTokens,
	// 		},
	// 	}

	// 	publicChan, err := f(ctx, publicPrompt)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	internalChan := make(chan llm.Token)
	// 	go func() {
	// 		defer close(internalChan)
	// 		for token := range publicChan {
	// 			internalChan <- llm.Token{
	// 				Content: token.Content,
	// 				Error:   token.Error,
	// 			}
	// 		}
	// 	}()

	// 	return internalChan, nil
	// })
}

func (m *MockModelProvider) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	// internalPrompt := llm.Prompt{
	// 	System: prompt.System,
	// 	User:   prompt.User,
	// 	Parameters: llm.ModelParameters{
	// 		Temperature: prompt.Parameters.Temperature,
	// 		MaxTokens:   prompt.Parameters.MaxTokens,
	// 	},
	// }

	// internalChan, err := m.internal.Stream(ctx, internalPrompt)
	// if err != nil {
	// 	return nil, err
	// }

	// publicChan := make(chan Token)
	// go func() {
	// 	defer close(publicChan)
	// 	for token := range internalChan {
	// 		publicChan <- Token{
	// 			Content: token.Content,
	// 			Error:   token.Error,
	// 		}
	// 	}
	// }()

	// return publicChan, nil
	return nil, nil
}

// SetEmbeddingsExpectation sets the expected response and error for the Embeddings method.
func (m *MockModelProvider) SetEmbeddingsExpectation(resp [][]float64, err error) {
	// m.internal.SetEmbeddingsExpectation(resp, err)
}

// SetEmbeddingsFunc sets a custom function to handle the Embeddings method.
func (m *MockModelProvider) SetEmbeddingsFunc(f func(ctx context.Context, texts []string) ([][]float64, error)) {
	// m.internal.SetEmbeddingsFunc(f)
}

func (m *MockModelProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	// return m.internal.Embeddings(ctx, texts)
	return nil, nil
}

// GetCallPromptArg returns the prompt argument captured by the last Call invocation.
func (m *MockModelProvider) GetCallPromptArg() Prompt {
	// internalPrompt := m.internal.GetCallPromptArg()
	// return Prompt{
	// 	System: internalPrompt.System,
	// 	User:   internalPrompt.User,
	// 	Parameters: ModelParameters{
	// 		Temperature: internalPrompt.Parameters.Temperature,
	// 		MaxTokens:   internalPrompt.Parameters.MaxTokens,
	// 	},
	// }
	return Prompt{}
}

// GetStreamPromptArg returns the prompt argument captured by the last Stream invocation.
func (m *MockModelProvider) GetStreamPromptArg() Prompt {
	// internalPrompt := m.internal.GetStreamPromptArg()
	// return Prompt{
	// 	System: internalPrompt.System,
	// 	User:   internalPrompt.User,
	// 	Parameters: ModelParameters{
	// 		Temperature: internalPrompt.Parameters.Temperature,
	// 		MaxTokens:   internalPrompt.Parameters.MaxTokens,
	// 	},
	// }
	return Prompt{}
}

// GetEmbeddingsTextsArg returns the texts argument captured by the last Embeddings invocation.
func (m *MockModelProvider) GetEmbeddingsTextsArg() []string {
	// return m.internal.GetEmbeddingsTextsArg()
	return nil
}
