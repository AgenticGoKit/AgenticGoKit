package v1beta

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/agenticgokit/agenticgokit/internal/llm"
)

// =============================================================================
// Full-stack fail-loud memory construction (issue #137)
// =============================================================================

// TestNewMemoryPropagatesEmbeddingErrors verifies the whole chain fails
// loudly: v1beta.NewMemory → core memory factory → chromem plugin →
// core.NewEmbeddingServiceForConfig. Before the fix, this returned a silent
// no-op memory.
func TestNewMemoryPropagatesEmbeddingErrors(t *testing.T) {
	_, err := NewMemory(&MemoryConfig{
		Enabled:  true,
		Provider: "chromem",
		Options: map[string]string{
			"embedding_provider": "not-a-real-provider",
		},
	})
	if err == nil {
		t.Fatal("NewMemory with an unknown embedding provider should return an error, got nil (silent no-op memory)")
	}
	if !strings.Contains(err.Error(), "unsupported embedding provider") {
		t.Errorf("error should identify the unsupported embedding provider, got: %v", err)
	}
}

// =============================================================================
// Zero-config default memory path (issue #137)
// =============================================================================

// TestDefaultMemoryDerivesEmbeddingsFromLLM verifies the nil-Memory default
// path: memory defaults to enabled chromem AND derives a real embedding
// configuration from the LLM provider. It also proves the plugins/embedding
// blank import registers real factories: if it did not, building the chromem
// provider with embedding_provider=ollama would fail loudly.
func TestDefaultMemoryDerivesEmbeddingsFromLLM(t *testing.T) {
	agent, err := NewBuilder("default-memory").WithConfig(&Config{
		LLM: LLMConfig{
			Provider: "ollama",
			Model:    "llama3.2",
			BaseURL:  "http://localhost:11434",
		},
	}).Build()
	if err != nil {
		t.Fatalf("Build with default memory failed: %v", err)
	}
	if agent.Memory() == nil {
		t.Fatal("default memory should be enabled when no MemoryConfig is provided")
	}

	opts := agent.Config().Memory.Options
	if got := opts["embedding_provider"]; got != "ollama" {
		t.Errorf("embedding_provider = %q, want %q (derived from LLM provider)", got, "ollama")
	}
	if got := opts["embedding_model"]; got != defaultOllamaEmbeddingModel {
		t.Errorf("embedding_model = %q, want %q — the chat model must never be used", got, defaultOllamaEmbeddingModel)
	}
	if got := opts["dimensions"]; got != "768" {
		t.Errorf("dimensions = %q, want %q for %s", got, "768", defaultOllamaEmbeddingModel)
	}
}

// =============================================================================
// Per-run sampling parameter plumbing (issue #143)
// =============================================================================

// recordingProvider is a llm.ModelProvider fake that records the last prompt
// it received, so tests can assert what actually reaches the provider.
type recordingProvider struct {
	lastPrompt llm.Prompt
}

func (r *recordingProvider) Call(ctx context.Context, p llm.Prompt) (llm.Response, error) {
	r.lastPrompt = p
	return llm.Response{Content: "ok"}, nil
}

func (r *recordingProvider) Stream(ctx context.Context, p llm.Prompt) (<-chan llm.Token, error) {
	r.lastPrompt = p
	ch := make(chan llm.Token, 1)
	ch <- llm.Token{Content: "ok"}
	close(ch)
	return ch, nil
}

func (r *recordingProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, nil
}

func newRecordingAgent(t *testing.T) (Agent, *recordingProvider) {
	t.Helper()
	agent, err := NewBuilder("rec").WithConfig(&Config{
		Timeout: 30 * time.Second,
		LLM: LLMConfig{
			Provider:    "mock",
			Model:       "mock-model",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
		Memory: &MemoryConfig{Enabled: false},
	}).Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	ra, ok := agent.(*realAgent)
	if !ok {
		t.Fatalf("unexpected agent type %T", agent)
	}
	rec := &recordingProvider{}
	ra.llmProvider = rec
	return agent, rec
}

// TestConfigSamplingParametersReachProvider verifies configured sampling
// values are passed per call. Before the fix, execute() built prompts with no
// Parameters at all, so the provider only ever saw its construction-time
// values and per-run overrides were dead code.
func TestConfigSamplingParametersReachProvider(t *testing.T) {
	agent, rec := newRecordingAgent(t)

	if _, err := agent.Run(context.Background(), "hello"); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	p := rec.lastPrompt.Parameters
	if p.Temperature == nil || *p.Temperature != 0.7 {
		t.Errorf("Parameters.Temperature = %v, want configured 0.7", p.Temperature)
	}
	if p.MaxTokens == nil || *p.MaxTokens != 2048 {
		t.Errorf("Parameters.MaxTokens = %v, want configured 2048", p.MaxTokens)
	}
}

// TestRunOptionsOverridesReachProvider verifies RunOptions.Temperature and
// MaxTokens overrides reach the provider — including an explicit temperature
// of 0, which must not be conflated with "unset" (issue #143).
func TestRunOptionsOverridesReachProvider(t *testing.T) {
	agent, rec := newRecordingAgent(t)

	zero := 0.0
	opts := &RunOptions{
		Temperature: &zero,
		MaxTokens:   99,
		ToolMode:    "none",
	}
	if _, err := agent.RunWithOptions(context.Background(), "hello", opts); err != nil {
		t.Fatalf("RunWithOptions failed: %v", err)
	}

	p := rec.lastPrompt.Parameters
	if p.Temperature == nil {
		t.Fatal("Parameters.Temperature = nil; explicit 0 override was dropped")
	}
	if *p.Temperature != 0 {
		t.Errorf("Parameters.Temperature = %v, want explicit 0", *p.Temperature)
	}
	if p.MaxTokens == nil || *p.MaxTokens != 99 {
		t.Errorf("Parameters.MaxTokens = %v, want override 99", p.MaxTokens)
	}
}
