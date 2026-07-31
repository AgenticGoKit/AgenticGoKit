package agentkit

// Build-time diagnostics.
//
// Non-fatal findings are values, not log lines. A framework that only logs a
// problem is invisible to any consumer that wires its own logger, so anything
// worth telling a developer is returned where code can act on it.

// Severity classifies a diagnostic.
type Severity string

const (
	// SevInfo is expected behavior worth knowing about.
	SevInfo Severity = "info"
	// SevWarning is probably-unintended configuration that still works.
	SevWarning Severity = "warning"
	// SevError is a degraded mode: the agent runs, but a feature will not
	// behave as advertised until it is fixed.
	SevError Severity = "error"
)

// DiagnosticCode identifies a finding for programmatic handling.
type DiagnosticCode string

const (
	// DiagNoNativeTools: tools are registered but the model cannot call them
	// natively, so a fallback strategy is in use.
	DiagNoNativeTools DiagnosticCode = "NO_NATIVE_TOOLS"

	// DiagNoNativeJSON: structured output was requested but the model has no
	// native JSON mode, so a tool or prompted strategy is in use.
	DiagNoNativeJSON DiagnosticCode = "NO_NATIVE_JSON"

	// DiagModalityUnsupported: the request carries content the model cannot
	// accept.
	DiagModalityUnsupported DiagnosticCode = "MODALITY_UNSUPPORTED"
)

// Diagnostic is a non-fatal finding produced while building an agent.
type Diagnostic struct {
	Severity Severity          `json:"severity"`
	Code     DiagnosticCode    `json:"code"`
	Message  string            `json:"message"`
	Details  map[string]string `json:"details,omitempty"`
}

// DiagnosticHandler receives diagnostics as they are produced.
type DiagnosticHandler func(Diagnostic)
