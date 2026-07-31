package agentkit

import (
	"errors"
	"fmt"
)

// Sentinel errors. Every error this package returns wraps one of these, so
// callers branch with errors.Is rather than matching message text.
var (
	// ErrInvalidConfig means the agent could not be constructed as described.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrMaxSteps means the run hit its step budget without producing a final
	// answer — usually a tool loop that never converges.
	ErrMaxSteps = errors.New("max steps exceeded")

	// ErrUsageLimit means the run hit a token or call budget.
	ErrUsageLimit = errors.New("usage limit exceeded")

	// ErrToolNotFound means the model called a tool that is not registered.
	ErrToolNotFound = errors.New("tool not found")

	// ErrNoOutput means the model finished without producing output matching
	// the requested type.
	ErrNoOutput = errors.New("no structured output produced")

	// ErrUnsupported means the provider or model lacks a required capability.
	// Callers should consult Provider.Capabilities to branch before calling.
	ErrUnsupported = errors.New("unsupported capability")

	// ErrProvider means the model backend failed. Inspect the wrapped error
	// for provider-specific detail.
	ErrProvider = errors.New("provider error")
)

// RetryError is returned by a tool or output validator to send a corrective
// message back to the model instead of failing the run. It turns validation
// into an autonomous self-correction loop.
type RetryError struct {
	Message string
}

func (e *RetryError) Error() string { return e.Message }

// Retry returns an error that asks the model to try again with the given
// guidance.
func Retry(format string, args ...any) error {
	return &RetryError{Message: fmt.Sprintf(format, args...)}
}

// ToolError identifies which tool failed while preserving the cause for
// errors.Is and errors.As.
type ToolError struct {
	Tool string
	Err  error
}

func (e *ToolError) Error() string { return fmt.Sprintf("tool %s: %v", e.Tool, e.Err) }
func (e *ToolError) Unwrap() error { return e.Err }

// StepError identifies which step of a run failed while preserving the cause.
type StepError struct {
	Step  int
	Phase string
	Err   error
}

func (e *StepError) Error() string {
	return fmt.Sprintf("step %d (%s): %v", e.Step, e.Phase, e.Err)
}
func (e *StepError) Unwrap() error { return e.Err }
