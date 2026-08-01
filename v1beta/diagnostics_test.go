package v1beta

import (
	"testing"
)

// TestDiagnosticsSurfaceDummyEmbeddingFallback verifies that an agent whose
// LLM provider has no embedding backend exposes the degraded-RAG condition as
// a value — via DiagnosticsOf and the builder's DiagnosticHandler — instead
// of only a log line.
func TestDiagnosticsSurfaceDummyEmbeddingFallback(t *testing.T) {
	var handled []Diagnostic
	agent, err := NewBuilder("anthropic-bot").
		WithConfig(&Config{
			// "mock" has no embedding backend, like anthropic
			LLM: LLMConfig{Provider: "mock", Model: "mock-model"},
		}).
		WithDiagnosticHandler(func(d Diagnostic) { handled = append(handled, d) }).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	diags := DiagnosticsOf(agent)
	if !hasDiagnostic(diags, DiagEmbeddingFallbackDummy) {
		t.Errorf("DiagnosticsOf missing %s, got %v", DiagEmbeddingFallbackDummy, diagCodes(diags))
	}
	if !hasDiagnostic(handled, DiagEmbeddingFallbackDummy) {
		t.Errorf("DiagnosticHandler did not receive %s, got %v", DiagEmbeddingFallbackDummy, diagCodes(handled))
	}

	for _, d := range diags {
		if d.Code == DiagEmbeddingFallbackDummy {
			if d.Severity != DiagError {
				t.Errorf("severity = %s, want %s", d.Severity, DiagError)
			}
			if d.Details["llm_provider"] != "mock" {
				t.Errorf("details.llm_provider = %q, want %q", d.Details["llm_provider"], "mock")
			}
		}
	}
}

// TestDiagnosticsSurfaceDisabledMemoryWithSettings verifies the
// likely-forgotten-Enabled condition is exposed programmatically.
func TestDiagnosticsSurfaceDisabledMemoryWithSettings(t *testing.T) {
	agent, err := NewBuilder("forgot-enabled").
		WithConfig(&Config{
			LLM:    LLMConfig{Provider: "mock", Model: "mock-model"},
			Memory: &MemoryConfig{Provider: "chromem"}, // Enabled left false
		}).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if !hasDiagnostic(DiagnosticsOf(agent), DiagMemoryDisabledWithConfig) {
		t.Errorf("expected %s diagnostic, got %v", DiagMemoryDisabledWithConfig, diagCodes(DiagnosticsOf(agent)))
	}
}

// TestDiagnosticsSurfaceDimensionMismatch verifies a dimensions/model
// disagreement is exposed before the first vector-store write fails.
func TestDiagnosticsSurfaceDimensionMismatch(t *testing.T) {
	agent, err := NewBuilder("dims").
		WithConfig(&Config{
			LLM: LLMConfig{Provider: "mock", Model: "mock-model"},
			Memory: &MemoryConfig{
				Enabled:  true,
				Provider: "chromem",
				Options: map[string]string{
					"embedding_provider": "ollama",
					"embedding_model":    "nomic-embed-text", // 768 dims
					"dimensions":         "1536",             // stale from a dummy-vector store
				},
			},
		}).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	diags := DiagnosticsOf(agent)
	if !hasDiagnostic(diags, DiagEmbeddingDimensionMismatch) {
		t.Fatalf("expected %s diagnostic, got %v", DiagEmbeddingDimensionMismatch, diagCodes(diags))
	}
	for _, d := range diags {
		if d.Code == DiagEmbeddingDimensionMismatch {
			if d.Details["model_dimensions"] != "768" || d.Details["configured_dimensions"] != "1536" {
				t.Errorf("details = %v, want model 768 / configured 1536", d.Details)
			}
		}
	}
}

// TestDiagnosticsCleanConfigIsQuiet verifies a healthy configuration emits no
// diagnostics — the surface must stay high-signal.
func TestDiagnosticsCleanConfigIsQuiet(t *testing.T) {
	agent, err := NewBuilder("clean").
		WithConfig(&Config{
			LLM: LLMConfig{Provider: "ollama", Model: "llama3.2"},
		}).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if diags := DiagnosticsOf(agent); len(diags) != 0 {
		t.Errorf("clean config produced diagnostics: %v", diagCodes(diags))
	}
}

func hasDiagnostic(diags []Diagnostic, code DiagnosticCode) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func diagCodes(diags []Diagnostic) []DiagnosticCode {
	out := make([]DiagnosticCode, len(diags))
	for i, d := range diags {
		out[i] = d.Code
	}
	return out
}
