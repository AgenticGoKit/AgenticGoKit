package validation

import "github.com/kunalkushwaha/agenticgokit/core"

// ValidationErrorType represents different types of validation errors
type ValidationErrorType string

const (
	ValidationErrorOrphanedAgent      ValidationErrorType = "ORPHANED_AGENT"
	ValidationErrorCircularDependency ValidationErrorType = "CIRCULAR_DEPENDENCY"
	ValidationErrorMissingAgent       ValidationErrorType = "MISSING_AGENT"
	ValidationErrorInvalidRouting     ValidationErrorType = "INVALID_ROUTING"
	ValidationErrorNoEntryPoint       ValidationErrorType = "NO_ENTRY_POINT"
	ValidationErrorMultipleEndpoints  ValidationErrorType = "MULTIPLE_ENDPOINTS"
)

// WorkflowValidationError represents a validation error with details
type WorkflowValidationError struct {
	Type        string                 `json:"type"`
	Message     string                 `json:"message"`
	Severity    string                 `json:"severity"`
	Component   string                 `json:"component"`
	Suggestions []string               `json:"suggestions"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// WorkflowValidationWarning represents a validation warning
type WorkflowValidationWarning struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Component string `json:"component"`
}

// WorkflowValidationResult contains the results of validation
type WorkflowValidationResult struct {
	IsValid  bool                        `json:"is_valid"`
	Errors   []WorkflowValidationError   `json:"errors"`
	Warnings []WorkflowValidationWarning `json:"warnings"`
}

// AgentNode represents an agent in the workflow graph
type AgentNode struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Dependencies []string               `json:"dependencies"`
	Routes       []string               `json:"routes"`
	IsEntryPoint bool                   `json:"is_entry_point"`
	IsEndpoint   bool                   `json:"is_endpoint"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowGraph represents a workflow graph structure
type WorkflowGraph struct {
	Nodes map[string]AgentNode `json:"nodes"`
	Edges []WorkflowGraphEdge  `json:"edges"`
}

// WorkflowGraphEdge represents an edge in the workflow graph
type WorkflowGraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type,omitempty"`
}

// WorkflowValidator interface for validating workflows
type WorkflowValidator interface {
	ValidateComposition(agents []core.Agent, mode string) core.WorkflowValidationResult
	ValidateOrchestration(agents map[string]core.AgentHandler, mode core.OrchestrationMode) core.WorkflowValidationResult
	ValidateWorkflowGraph(graph core.WorkflowGraph) core.WorkflowValidationResult
}

// WorkflowValidatorFactory is the function signature for creating WorkflowValidator instances
type WorkflowValidatorFactory func() WorkflowValidator

// workflowValidatorFactory holds the registered factory function
var workflowValidatorFactory WorkflowValidatorFactory

// RegisterWorkflowValidatorFactory registers the WorkflowValidator factory function
func RegisterWorkflowValidatorFactory(factory WorkflowValidatorFactory) {
	workflowValidatorFactory = factory
}

// NewWorkflowValidator creates a new WorkflowValidator instance
func NewWorkflowValidator() WorkflowValidator {
	if workflowValidatorFactory != nil {
		return workflowValidatorFactory()
	}

	// Fallback to simple implementation if internal package not imported
	return &simpleWorkflowValidator{}
}

// Simple implementation for core package
type simpleWorkflowValidator struct{}

func (wv *simpleWorkflowValidator) ValidateComposition(agents []Agent, mode string) WorkflowValidationResult {
	result := WorkflowValidationResult{
		IsValid:  true,
		Errors:   []WorkflowValidationError{},
		Warnings: []WorkflowValidationWarning{},
	}

	// Basic validation
	if len(agents) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, WorkflowValidationError{
			Type:      string(ValidationErrorNoEntryPoint),
			Message:   "No agents defined",
			Severity:  "error",
			Component: "composition",
		})
	}

	return result
}

func (wv *simpleWorkflowValidator) ValidateOrchestration(agents map[string]AgentHandler, mode OrchestrationMode) WorkflowValidationResult {
	result := WorkflowValidationResult{
		IsValid:  true,
		Errors:   []WorkflowValidationError{},
		Warnings: []WorkflowValidationWarning{},
	}

	// Basic validation
	if len(agents) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, WorkflowValidationError{
			Type:      string(ValidationErrorNoEntryPoint),
			Message:   "No agents registered",
			Severity:  "error",
			Component: "orchestration",
		})
	}

	return result
}

func (wv *simpleWorkflowValidator) ValidateWorkflowGraph(graph WorkflowGraph) WorkflowValidationResult {
	result := WorkflowValidationResult{
		IsValid:  true,
		Errors:   []WorkflowValidationError{},
		Warnings: []WorkflowValidationWarning{},
	}

	// Basic validation
	if len(graph.Nodes) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, WorkflowValidationError{
			Type:      string(ValidationErrorNoEntryPoint),
			Message:   "No nodes in graph",
			Severity:  "error",
			Component: "graph",
		})
	}

	return result
}
