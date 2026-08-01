package core

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// saveEmbeddingFactories snapshots and clears the package-level embedding
// factory registrations so each test controls the registration state.
func saveEmbeddingFactories(t *testing.T) {
	t.Helper()
	savedOpenAI := openAIEmbeddingFactory
	savedOllama := ollamaEmbeddingFactory
	savedDummy := dummyEmbeddingFactory
	t.Cleanup(func() {
		openAIEmbeddingFactory = savedOpenAI
		ollamaEmbeddingFactory = savedOllama
		dummyEmbeddingFactory = savedDummy
	})
	openAIEmbeddingFactory = nil
	ollamaEmbeddingFactory = nil
	dummyEmbeddingFactory = nil
}

func TestNewEmbeddingServiceForConfig_UnregisteredFactoryFailsLoudly(t *testing.T) {
	saveEmbeddingFactories(t)

	for _, provider := range []string{"openai", "azure", "ollama"} {
		cfg := AgentMemoryConfig{
			Provider:  "chromem",
			Embedding: EmbeddingConfig{Provider: provider, Model: "some-model"},
		}
		svc, err := NewEmbeddingServiceForConfig(cfg)
		if err == nil {
			t.Errorf("provider %q: expected error when no factory registered, got service %T", provider, svc)
			continue
		}
		if !strings.Contains(err.Error(), "plugins/embedding") {
			t.Errorf("provider %q: error should point at the plugins/embedding import, got: %v", provider, err)
		}
		if !errors.Is(err, ErrEmbeddingFactoryNotRegistered) {
			t.Errorf("provider %q: error should wrap ErrEmbeddingFactoryNotRegistered for programmatic handling, got: %v", provider, err)
		}
	}
}

func TestNewEmbeddingServiceForConfig_RegisteredFactoryUsed(t *testing.T) {
	saveEmbeddingFactories(t)

	called := false
	RegisterOllamaEmbeddingFactory(func(model, baseURL string) EmbeddingService {
		called = true
		return &noOpEmbeddingService{dimensions: 768}
	})

	cfg := AgentMemoryConfig{
		Provider:  "chromem",
		Embedding: EmbeddingConfig{Provider: "ollama", Model: "nomic-embed-text"},
	}
	svc, err := NewEmbeddingServiceForConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("registered Ollama factory was not invoked")
	}
	if svc == nil {
		t.Fatal("expected a service, got nil")
	}
}

func TestNewEmbeddingServiceForConfig_UnknownProviderIsError(t *testing.T) {
	saveEmbeddingFactories(t)

	cfg := AgentMemoryConfig{
		Provider:  "chromem",
		Embedding: EmbeddingConfig{Provider: "anthropic"},
	}
	_, err := NewEmbeddingServiceForConfig(cfg)
	if err == nil {
		t.Fatal("expected error for unknown embedding provider, got nil")
	}
	if !errors.Is(err, ErrEmbeddingProviderUnsupported) {
		t.Errorf("error should wrap ErrEmbeddingProviderUnsupported for programmatic handling, got: %v", err)
	}
}

func TestNewEmbeddingServiceForConfig_ExplicitDummyAndEmptyStillConstruct(t *testing.T) {
	saveEmbeddingFactories(t)

	for _, provider := range []string{"dummy", ""} {
		cfg := AgentMemoryConfig{
			Provider:   "chromem",
			Dimensions: 4,
			Embedding:  EmbeddingConfig{Provider: provider},
		}
		svc, err := NewEmbeddingServiceForConfig(cfg)
		if err != nil {
			t.Errorf("provider %q: unexpected error: %v", provider, err)
			continue
		}
		vec, err := svc.GenerateEmbedding(context.Background(), "hello")
		if err != nil {
			t.Errorf("provider %q: GenerateEmbedding failed: %v", provider, err)
		}
		if len(vec) != 4 {
			t.Errorf("provider %q: got %d dimensions, want 4", provider, len(vec))
		}
	}
}
