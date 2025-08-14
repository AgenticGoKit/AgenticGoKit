// Package factory provides factory functions for creating internal implementations
package factory

import (
	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/internal/validation"
	"github.com/kunalkushwaha/agenticgokit/internal/visualization"
)

// NewMermaidGenerator creates a new Mermaid generator implementation
func NewMermaidGenerator() core.MermaidGenerator {
	return visualization.NewMermaidGeneratorImplementation()
}

// NewWorkflowValidator creates a new workflow validator implementation
func NewWorkflowValidator() core.WorkflowValidator {
	return validation.NewWorkflowValidatorImplementation()
}
