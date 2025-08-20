package core

import (
	"context"
	"strings"
	"testing"
)

// fakeProvider implements ModelProvider for testing registration
type fakeProvider struct{}

func (f *fakeProvider) Call(ctx context.Context, p Prompt) (Response, error) {
	return Response{Content: "ok"}, nil
}
func (f *fakeProvider) Stream(ctx context.Context, p Prompt) (<-chan Token, error) {
	ch := make(chan Token)
	close(ch)
	return ch, nil
}
func (f *fakeProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return [][]float64{}, nil
}

func TestModelProviderRegistry_RegisterAndResolve(t *testing.T) {
	name := "dummy"
	RegisterModelProviderFactory(name, func(cfg LLMProviderConfig) (ModelProvider, error) {
		return &fakeProvider{}, nil
	})

	p, err := NewModelProviderFromConfig(LLMProviderConfig{Type: name})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p == nil {
		t.Fatal("expected provider, got nil")
	}
	resp, err := p.Call(context.Background(), Prompt{User: "ping"})
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}
	if resp.Content != "ok" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
}

func TestModelProviderRegistry_MissingProvider_YieldsActionableError(t *testing.T) {
	_, err := NewModelProviderFromConfig(LLMProviderConfig{Type: "not-registered"})
	if err == nil {
		t.Fatal("expected error for missing provider, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "not registered") {
		t.Fatalf("expected 'not registered' hint, got: %s", msg)
	}
}
