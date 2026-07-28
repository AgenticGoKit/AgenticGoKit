package v1beta

import (
	"context"
	"errors"
	"testing"

	"github.com/agenticgokit/agenticgokit/internal/llm"
)

// fakeMiddlewareProvider is a minimal llm.ModelProvider for middleware tests
// — only Call is exercised by execute().
type fakeMiddlewareProvider struct {
	callFunc func(ctx context.Context, prompt llm.Prompt) (llm.Response, error)
}

func (f *fakeMiddlewareProvider) Call(ctx context.Context, prompt llm.Prompt) (llm.Response, error) {
	if f.callFunc != nil {
		return f.callFunc(ctx, prompt)
	}
	return llm.Response{Content: "default"}, nil
}
func (f *fakeMiddlewareProvider) Stream(ctx context.Context, prompt llm.Prompt) (<-chan llm.Token, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeMiddlewareProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, errors.New("not implemented")
}

// recordingMiddleware records BeforeRun/AfterRun invocations and can be
// configured to mutate input/result/err or abort.
type recordingMiddleware struct {
	name       string
	calls      *[]string
	beforeErr  error
	beforeTag  string // appended to input in BeforeRun, if non-empty
	afterErr   error
	afterClear bool // if true, AfterRun clears an incoming error
}

func (m *recordingMiddleware) Name() string { return m.name }

func (m *recordingMiddleware) BeforeRun(ctx context.Context, input string) (context.Context, string, error) {
	*m.calls = append(*m.calls, "before:"+m.name)
	if m.beforeErr != nil {
		return ctx, input, m.beforeErr
	}
	if m.beforeTag != "" {
		input += m.beforeTag
	}
	return ctx, input, nil
}

func (m *recordingMiddleware) AfterRun(ctx context.Context, input string, result *Result, err error) (*Result, error) {
	*m.calls = append(*m.calls, "after:"+m.name)
	if m.afterErr != nil {
		return result, m.afterErr
	}
	if m.afterClear {
		return result, nil
	}
	return result, err
}

func minimalAgent(provider llm.ModelProvider, middlewares []AgentMiddleware) *realAgent {
	return &realAgent{
		config: &Config{
			Name: "test-agent",
			LLM:  LLMConfig{Provider: "mock", Model: "mock-model"},
		},
		llmProvider: provider,
		initialized: true,
		metrics:     &agentMetrics{},
		middlewares: middlewares,
	}
}

func TestMiddleware_NilMiddlewaresIsNoOp(t *testing.T) {
	agent := minimalAgent(&fakeMiddlewareProvider{}, nil)
	result, err := agent.Run(context.Background(), "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "default" {
		t.Fatalf("got %q, want %q", result.Content, "default")
	}
}

func TestMiddleware_BeforeRunOrderAndInputMutation(t *testing.T) {
	var calls []string
	var gotInput string
	provider := &fakeMiddlewareProvider{
		callFunc: func(ctx context.Context, prompt llm.Prompt) (llm.Response, error) {
			gotInput = prompt.User
			return llm.Response{Content: "ok"}, nil
		},
	}
	mw1 := &recordingMiddleware{name: "mw1", calls: &calls, beforeTag: "+mw1"}
	mw2 := &recordingMiddleware{name: "mw2", calls: &calls, beforeTag: "+mw2"}

	agent := minimalAgent(provider, []AgentMiddleware{mw1, mw2})
	_, err := agent.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotInput != "input+mw1+mw2" {
		t.Fatalf("got input %q, want %q (BeforeRun must run in registration order)", gotInput, "input+mw1+mw2")
	}
	wantCalls := []string{"before:mw1", "before:mw2", "after:mw2", "after:mw1"}
	if !equalStrings(calls, wantCalls) {
		t.Fatalf("got call order %v, want %v (AfterRun must run LIFO)", calls, wantCalls)
	}
}

func TestMiddleware_BeforeRunErrorAbortsBeforeLLMCall(t *testing.T) {
	var calls []string
	llmCalled := false
	provider := &fakeMiddlewareProvider{
		callFunc: func(ctx context.Context, prompt llm.Prompt) (llm.Response, error) {
			llmCalled = true
			return llm.Response{Content: "should not happen"}, nil
		},
	}
	wantErr := errors.New("boom")
	mw := &recordingMiddleware{name: "mw", calls: &calls, beforeErr: wantErr}

	agent := minimalAgent(provider, []AgentMiddleware{mw})
	_, err := agent.Run(context.Background(), "input")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("got err %v, want wrapping %v", err, wantErr)
	}
	agentErr, ok := err.(*AgentError)
	if !ok || agentErr.Code != ErrMiddlewareBeforeRun {
		t.Fatalf("got err %v (type %T), want *AgentError with code %v", err, err, ErrMiddlewareBeforeRun)
	}
	if llmCalled {
		t.Fatal("LLM must not be called when BeforeRun aborts the run")
	}
	if len(calls) != 1 || calls[0] != "before:mw" {
		t.Fatalf("got calls %v, want only before:mw (no AfterRun on BeforeRun abort)", calls)
	}
}

func TestMiddleware_AfterRunCanClearError(t *testing.T) {
	provider := &fakeMiddlewareProvider{
		callFunc: func(ctx context.Context, prompt llm.Prompt) (llm.Response, error) {
			return llm.Response{}, errors.New("llm failed")
		},
	}
	var calls []string
	mw := &recordingMiddleware{name: "mw", calls: &calls, afterClear: true}

	agent := minimalAgent(provider, []AgentMiddleware{mw})
	_, err := agent.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("expected AfterRun to clear the error, got: %v", err)
	}
}

func TestMiddleware_AfterRunErrorIsNotWrappedAsMiddlewareError(t *testing.T) {
	provider := &fakeMiddlewareProvider{}
	wantErr := errors.New("deliberate replacement")
	var calls []string
	mw := &recordingMiddleware{name: "mw", calls: &calls, afterErr: wantErr}

	agent := minimalAgent(provider, []AgentMiddleware{mw})
	_, err := agent.Run(context.Background(), "input")

	if !errors.Is(err, wantErr) {
		t.Fatalf("got err %v, want %v unwrapped (not %v)", err, wantErr, ErrMiddlewareAfterRun)
	}
	if _, ok := err.(*AgentError); ok {
		t.Fatal("AfterRun's returned error must not be wrapped as *AgentError/ErrMiddlewareAfterRun")
	}
}

func TestMiddleware_SkipMiddlewareByName(t *testing.T) {
	var calls []string
	provider := &fakeMiddlewareProvider{}
	mw1 := &recordingMiddleware{name: "mw1", calls: &calls}
	mw2 := &recordingMiddleware{name: "mw2", calls: &calls}

	agent := minimalAgent(provider, []AgentMiddleware{mw1, mw2})
	_, err := agent.RunWithOptions(context.Background(), "input", &RunOptions{
		SkipMiddleware: []string{"mw1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantCalls := []string{"before:mw2", "after:mw2"}
	if !equalStrings(calls, wantCalls) {
		t.Fatalf("got calls %v, want %v (mw1 must be fully skipped)", calls, wantCalls)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
