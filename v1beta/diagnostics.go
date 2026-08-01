package v1beta

// Build-time diagnostics.
//
// A framework log line is easy to miss (or lost entirely when the consumer
// wires their own logger). Non-fatal configuration findings are therefore
// also surfaced as values: collect them with Builder.WithDiagnosticHandler,
// or read them from a built agent with DiagnosticsOf. Every diagnostic is
// still logged, so nothing changes for consumers that ignore this API.

// DiagnosticSeverity classifies how serious a diagnostic is.
type DiagnosticSeverity string

const (
	// DiagInfo is informational: expected behavior worth knowing about.
	DiagInfo DiagnosticSeverity = "info"
	// DiagWarning indicates configuration that is probably not what the
	// user intended, but the agent still works.
	DiagWarning DiagnosticSeverity = "warning"
	// DiagError indicates a degraded mode: the agent constructs and runs,
	// but a feature will not behave as advertised until fixed.
	DiagError DiagnosticSeverity = "error"
)

// DiagnosticCode identifies a specific finding so consumers can handle it
// programmatically instead of string-matching messages.
type DiagnosticCode string

const (
	// DiagEmbeddingFallbackDummy: memory is enabled but no real embedding
	// backend is configured or derivable from the LLM provider. Chat
	// history works; semantic search and RAG return meaningless results.
	DiagEmbeddingFallbackDummy DiagnosticCode = "EMBEDDING_FALLBACK_DUMMY"

	// DiagMemoryDisabledWithConfig: a MemoryConfig carrying settings was
	// provided with Enabled left false — likely a forgotten Enabled: true
	// after the presence-implies-enabled behavior was removed.
	DiagMemoryDisabledWithConfig DiagnosticCode = "MEMORY_DISABLED_WITH_CONFIG"

	// DiagEmbeddingDimensionMismatch: the configured vector dimensions
	// disagree with the embedding model's known dimensions; vector-store
	// writes may fail or similarity search may degrade.
	DiagEmbeddingDimensionMismatch DiagnosticCode = "EMBEDDING_DIMENSION_MISMATCH"
)

// Diagnostic is a non-fatal finding produced while building an agent.
type Diagnostic struct {
	Severity DiagnosticSeverity `json:"severity"`
	Code     DiagnosticCode     `json:"code"`
	// Message is human-readable and includes the suggested fix.
	Message string `json:"message"`
	// Details carries structured context (provider names, option keys,
	// dimension values, ...) for programmatic handling.
	Details map[string]string `json:"details,omitempty"`
}

// DiagnosticHandler receives diagnostics as they are produced during Build.
type DiagnosticHandler func(Diagnostic)

// diagnosticProvider is the optional interface implemented by agents that
// collect build-time diagnostics. It is deliberately not part of the Agent
// interface so existing Agent implementations remain valid.
type diagnosticProvider interface {
	Diagnostics() []Diagnostic
}

// DiagnosticsOf returns the build-time diagnostics collected by an agent,
// or nil if the agent implementation does not expose any.
//
//	agent, err := v1beta.NewBuilder("bot").WithConfig(cfg).Build()
//	for _, d := range v1beta.DiagnosticsOf(agent) {
//	    if d.Code == v1beta.DiagEmbeddingFallbackDummy {
//	        // fail deployment, alert, or reconfigure
//	    }
//	}
func DiagnosticsOf(agent Agent) []Diagnostic {
	if agent == nil {
		return nil
	}
	if dp, ok := agent.(diagnosticProvider); ok {
		return dp.Diagnostics()
	}
	return nil
}
