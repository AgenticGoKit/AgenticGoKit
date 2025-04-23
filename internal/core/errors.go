package agentflow

import (
	"errors"
	"fmt"
	"strings"
)

// MultiError aggregates multiple errors that occurred during an operation,
// typically used by agents that run multiple sub-operations concurrently (like ParallelAgent).
type MultiError struct {
	// Errors contains the list of non-nil errors that occurred.
	Errors []error
}

// NewMultiError creates a MultiError from a slice of errors.
// It filters out any nil errors present in the input slice.
// If the resulting list of errors is empty, it returns nil.
func NewMultiError(errs []error) error {
	filteredErrs := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			filteredErrs = append(filteredErrs, err)
		}
	}
	if len(filteredErrs) == 0 {
		return nil // No errors, return nil
	}
	return &MultiError{Errors: filteredErrs}
}

// Error returns a string representation summarizing all aggregated errors.
// It provides a count and lists each individual error.
func (m *MultiError) Error() string {
	if len(m.Errors) == 0 {
		return "no errors"
	}
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "%d errors occurred:\n", len(m.Errors))
	for i, err := range m.Errors {
		fmt.Fprintf(&builder, "\t* [%d] %s\n", i+1, err.Error())
	}
	return builder.String()
}

// Unwrap returns the underlying slice of errors. This allows MultiError
// to be used with errors.Is and errors.As for checking against the
// MultiError type itself, but not directly against the contained errors.
// To check for specific contained errors, iterate through the Errors slice.
func (m *MultiError) Unwrap() []error {
	return m.Errors
}

// ErrMaxIterationsReached indicates that a LoopAgent terminated because
// it reached its configured maximum iteration limit without the stop
// condition being met.
var ErrMaxIterationsReached = errors.New("maximum loop iterations reached")
