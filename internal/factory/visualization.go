// Package factory provides factory functions for creating internal implementations
package factory

import (
	"github.com/agenticgokit/agenticgokit/core"
	"github.com/agenticgokit/agenticgokit/internal/validation"
	"github.com/agenticgokit/agenticgokit/internal/visualization"
)

// NewMermaidGenerator creates a new Mermaid generator implementation
func NewMermaidGenerator() core.MermaidGenerator {
	return visualization.NewMermaidGeneratorImplementation()
}

// NewWorkflowValidator creates a new workflow validator implementation
func NewWorkflowValidator() core.WorkflowValidator {
	return validation.NewWorkflowValidatorImplementation()
}

