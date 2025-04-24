package llm

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestMockModelProvider_Call(t *testing.T) {
	mock := NewMockModelProvider()
	ctx := context.Background()
	testPrompt := Prompt{System: "sys", User: "usr"}
	expectedResp := Response{Content: "response"}
	expectedErr := errors.New("call error")

	// Test success case
	mock.SetCallExpectation(expectedResp, nil)
	resp, err := mock.Call(ctx, testPrompt)
	if err != nil {
		t.Errorf("Call() returned unexpected error: %v", err)
	}
	if !reflect.DeepEqual(resp, expectedResp) {
		t.Errorf("Call() response mismatch: got %v, want %v", resp, expectedResp)
	}
	if !reflect.DeepEqual(mock.GetCallPromptArg(), testPrompt) {
		t.Errorf("Call() prompt arg mismatch: got %v, want %v", mock.GetCallPromptArg(), testPrompt)
	}

	// Test error case
	mock.SetCallExpectation(Response{}, expectedErr)
	resp, err = mock.Call(ctx, testPrompt)
	if err == nil {
		t.Errorf("Call() expected an error, but got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Call() error mismatch: got %v, want %v", err, expectedErr)
	}
	if !reflect.DeepEqual(resp, Response{}) { // Expect zero value response on error
		t.Errorf("Call() response on error should be zero value, got %v", resp)
	}

	// Test context cancellation (using custom func)
	ctxCancelled, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately
	mock.SetCallFunc(func(ctx context.Context, prompt Prompt) (Response, error) {
		if ctx.Err() != nil {
			return Response{}, ctx.Err() // Simulate context check
		}
		return Response{}, errors.New("should have detected cancellation")
	})
	_, err = mock.Call(ctxCancelled, testPrompt)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Call() with cancelled context error mismatch: got %v, want %v", err, context.Canceled)
	}
}

func TestMockModelProvider_Stream(t *testing.T) {
	mock := NewMockModelProvider()
	ctx := context.Background()
	testPrompt := Prompt{System: "sys", User: "stream"}
	expectedErr := errors.New("stream error")

	// Test success case
	expectedTokens := []Token{{Content: "hello"}, {Content: " world"}}
	streamChan := make(chan Token, len(expectedTokens))
	for _, tk := range expectedTokens {
		streamChan <- tk
	}
	close(streamChan) // Close channel after filling

	mock.SetStreamExpectation(streamChan, nil)
	respChan, err := mock.Stream(ctx, testPrompt)
	if err != nil {
		t.Errorf("Stream() returned unexpected error: %v", err)
	}
	if respChan == nil {
		t.Fatal("Stream() returned nil channel unexpectedly")
	}

	// Drain the channel to verify content
	receivedTokens := []Token{}
	for tk := range respChan {
		receivedTokens = append(receivedTokens, tk)
	}
	if !reflect.DeepEqual(receivedTokens, expectedTokens) {
		t.Errorf("Stream() tokens mismatch:\ngot:  %v\nwant: %v", receivedTokens, expectedTokens)
	}
	if !reflect.DeepEqual(mock.GetStreamPromptArg(), testPrompt) {
		t.Errorf("Stream() prompt arg mismatch: got %v, want %v", mock.GetStreamPromptArg(), testPrompt)
	}

	// Test error case (error returned immediately)
	mock.SetStreamExpectation(nil, expectedErr) // Error returned, channel is nil
	respChan, err = mock.Stream(ctx, testPrompt)
	if err == nil {
		t.Errorf("Stream() expected an error, but got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Stream() error mismatch: got %v, want %v", err, expectedErr)
	}
	if respChan != nil {
		t.Errorf("Stream() channel should be nil on immediate error, got %v", respChan)
	}

	// Test error during stream (sent via token)
	errDuringStream := errors.New("error mid-stream")
	streamChanWithError := make(chan Token, 2)
	streamChanWithError <- Token{Content: "first "}
	streamChanWithError <- Token{Error: errDuringStream} // Send error in last token
	close(streamChanWithError)

	mock.SetStreamExpectation(streamChanWithError, nil) // No immediate error
	respChan, err = mock.Stream(ctx, testPrompt)
	if err != nil {
		t.Fatalf("Stream() returned unexpected immediate error: %v", err)
	}

	var lastToken Token
	for tk := range respChan {
		lastToken = tk
	}
	if lastToken.Error == nil {
		t.Errorf("Expected error in last token, but got nil")
	}
	if !errors.Is(lastToken.Error, errDuringStream) {
		t.Errorf("Error in last token mismatch: got %v, want %v", lastToken.Error, errDuringStream)
	}

	// Test context cancellation (using custom func)
	ctxCancelled, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately
	mock.SetStreamFunc(func(ctx context.Context, prompt Prompt) (<-chan Token, error) {
		if ctx.Err() != nil {
			return nil, ctx.Err() // Simulate context check before returning channel
		}
		return nil, errors.New("should have detected cancellation")
	})
	_, err = mock.Stream(ctxCancelled, testPrompt)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Stream() with cancelled context error mismatch: got %v, want %v", err, context.Canceled)
	}
}

func TestMockModelProvider_Embeddings(t *testing.T) {
	mock := NewMockModelProvider()
	ctx := context.Background()
	testTexts := []string{"text1", "text2"}
	expectedResp := [][]float64{{0.1, 0.2}, {0.3, 0.4}}
	expectedErr := errors.New("embeddings error")

	// Test success case
	mock.SetEmbeddingsExpectation(expectedResp, nil)
	resp, err := mock.Embeddings(ctx, testTexts)
	if err != nil {
		t.Errorf("Embeddings() returned unexpected error: %v", err)
	}
	if !reflect.DeepEqual(resp, expectedResp) {
		t.Errorf("Embeddings() response mismatch: got %v, want %v", resp, expectedResp)
	}
	if !reflect.DeepEqual(mock.GetEmbeddingsTextsArg(), testTexts) {
		t.Errorf("Embeddings() texts arg mismatch: got %v, want %v", mock.GetEmbeddingsTextsArg(), testTexts)
	}

	// Test error case
	mock.SetEmbeddingsExpectation(nil, expectedErr) // Response is nil on error
	resp, err = mock.Embeddings(ctx, testTexts)
	if err == nil {
		t.Errorf("Embeddings() expected an error, but got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Embeddings() error mismatch: got %v, want %v", err, expectedErr)
	}
	if resp != nil {
		t.Errorf("Embeddings() response on error should be nil, got %v", resp)
	}

	// Test context cancellation (using custom func)
	ctxCancelled, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately
	mock.SetEmbeddingsFunc(func(ctx context.Context, texts []string) ([][]float64, error) {
		if ctx.Err() != nil {
			return nil, ctx.Err() // Simulate context check
		}
		return nil, errors.New("should have detected cancellation")
	})
	_, err = mock.Embeddings(ctxCancelled, testTexts)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Embeddings() with cancelled context error mismatch: got %v, want %v", err, context.Canceled)
	}
}
