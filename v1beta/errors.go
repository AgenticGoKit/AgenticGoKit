// package v1beta provides the next generation streamlined API for AgenticGoKit
package v1beta

import (
	"errors"
	"fmt"
	"strings"
)

// =============================================================================
// EXTENDED ERROR CODES
// =============================================================================

// NOTE: ErrorCode and AgentError are defined in agent.go
// This file extends the error codes and provides helper functions

const (
	// LLM errors (extending existing codes)
	ErrCodeLLMNotConfigured ErrorCode = "LLM_NOT_CONFIGURED"
	ErrCodeLLMCallFailed    ErrorCode = "LLM_CALL_FAILED"
	ErrCodeLLMTimeout       ErrorCode = "LLM_TIMEOUT"
	ErrCodeLLMRateLimited   ErrorCode = "LLM_RATE_LIMITED"
	ErrCodeLLMInvalidModel  ErrorCode = "LLM_INVALID_MODEL"
	ErrCodeLLMConnection    ErrorCode = "LLM_CONNECTION"
	ErrCodeLLMAuth          ErrorCode = "LLM_AUTH"
	ErrCodeLLMQuotaExceeded ErrorCode = "LLM_QUOTA_EXCEEDED"

	// Tool errors
	ErrCodeToolNotFound     ErrorCode = "TOOL_NOT_FOUND"
	ErrCodeToolExecute      ErrorCode = "TOOL_EXECUTE"
	ErrCodeToolTimeout      ErrorCode = "TOOL_TIMEOUT"
	ErrCodeToolInvalidArgs  ErrorCode = "TOOL_INVALID_ARGS"
	ErrCodeToolNotAvailable ErrorCode = "TOOL_NOT_AVAILABLE"

	// Memory errors
	ErrCodeMemoryNotConfigured  ErrorCode = "MEMORY_NOT_CONFIGURED"
	ErrCodeMemoryStore          ErrorCode = "MEMORY_STORE"
	ErrCodeMemoryQuery          ErrorCode = "MEMORY_QUERY"
	ErrCodeMemoryConnection     ErrorCode = "MEMORY_CONNECTION"
	ErrCodeMemoryInvalidBackend ErrorCode = "MEMORY_INVALID_BACKEND"

	// Workflow errors
	ErrCodeWorkflowInvalid       ErrorCode = "WORKFLOW_INVALID"
	ErrCodeWorkflowStepFailed    ErrorCode = "WORKFLOW_STEP_FAILED"
	ErrCodeWorkflowTimeout       ErrorCode = "WORKFLOW_TIMEOUT"
	ErrCodeWorkflowCycleDetected ErrorCode = "WORKFLOW_CYCLE_DETECTED"
	ErrCodeWorkflowMaxIterations ErrorCode = "WORKFLOW_MAX_ITERATIONS"

	// MCP errors
	ErrCodeMCPServerNotFound  ErrorCode = "MCP_SERVER_NOT_FOUND"
	ErrCodeMCPConnection      ErrorCode = "MCP_CONNECTION"
	ErrCodeMCPTimeout         ErrorCode = "MCP_TIMEOUT"
	ErrCodeMCPInvalidResponse ErrorCode = "MCP_INVALID_RESPONSE"
	ErrCodeMCPServerUnhealthy ErrorCode = "MCP_SERVER_UNHEALTHY"

	// Handler errors
	ErrCodeHandlerFailed  ErrorCode = "HANDLER_FAILED"
	ErrCodeHandlerTimeout ErrorCode = "HANDLER_TIMEOUT"
	ErrCodeHandlerPanic   ErrorCode = "HANDLER_PANIC"

	// Agent lifecycle errors
	ErrCodeAgentNotInitialized ErrorCode = "AGENT_NOT_INITIALIZED"
	ErrCodeAgentShutdown       ErrorCode = "AGENT_SHUTDOWN"
	ErrCodeAgentInvalidState   ErrorCode = "AGENT_INVALID_STATE"

	// Validation errors
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeMissingRequired  ErrorCode = "MISSING_REQUIRED"

	// Runtime errors
	ErrCodeTimeout        ErrorCode = "TIMEOUT"
	ErrCodeCancelled      ErrorCode = "CANCELLED"
	ErrCodeInternal       ErrorCode = "INTERNAL"
	ErrCodeNotImplemented ErrorCode = "NOT_IMPLEMENTED"
	ErrCodeUnsupported    ErrorCode = "UNSUPPORTED"

	// Configuration errors (extending existing)
	ErrCodeConfigNotFound   ErrorCode = "CONFIG_NOT_FOUND"
	ErrCodeConfigParse      ErrorCode = "CONFIG_PARSE"
	ErrCodeConfigValidation ErrorCode = "CONFIG_VALIDATION"
)

// =============================================================================
// COMPONENT-SPECIFIC ERROR HELPERS
// =============================================================================

// LLMError creates an LLM-related error with helpful suggestions
func LLMError(code ErrorCode, message string, cause error) *AgentError {
	err := NewAgentErrorWithError(code, message, cause)
	err.AddDetail("component", "LLM")

	// Add helpful suggestions based on error code
	switch code {
	case ErrCodeLLMNotConfigured:
		err.AddDetail("suggestion", "Configure an LLM provider using WithLLM() or set LLM configuration in your config file")
	case ErrCodeLLMCallFailed:
		err.AddDetail("suggestion", "Check your API key, network connection, and LLM provider status")
	case ErrCodeLLMTimeout:
		err.AddDetail("suggestion", "Increase timeout or try with a smaller prompt")
	case ErrCodeLLMRateLimited:
		err.AddDetail("suggestion", "Implement rate limiting or use exponential backoff with retry logic")
	case ErrCodeLLMAuth:
		err.AddDetail("suggestion", "Verify your API key is correct and has not expired")
	case ErrCodeLLMQuotaExceeded:
		err.AddDetail("suggestion", "Check your account quota and upgrade if necessary")
	}

	return err
}

// ToolError creates a tool-related error with helpful suggestions
func ToolError(code ErrorCode, toolName string, message string, cause error) *AgentError {
	err := NewAgentErrorWithError(code, message, cause)
	err.AddDetail("component", "Tool")
	err.AddDetail("tool_name", toolName)

	// Add helpful suggestions
	switch code {
	case ErrCodeToolNotFound:
		err.AddDetail("suggestion", fmt.Sprintf("Tool '%s' is not registered. Use WithTools() to register tools", toolName))
	case ErrCodeToolExecute:
		err.AddDetail("suggestion", "Check tool implementation and arguments")
	case ErrCodeToolInvalidArgs:
		err.AddDetail("suggestion", "Verify the tool arguments match the expected schema")
	case ErrCodeToolTimeout:
		err.AddDetail("suggestion", "Increase tool timeout or optimize tool implementation")
	}

	return err
}

// MemoryError creates a memory-related error with helpful suggestions
func MemoryError(code ErrorCode, message string, cause error) *AgentError {
	err := NewAgentErrorWithError(code, message, cause)
	err.AddDetail("component", "Memory")

	// Add helpful suggestions
	switch code {
	case ErrCodeMemoryNotConfigured:
		err.AddDetail("suggestion", "Configure memory using WithMemory() or set memory configuration in your config file")
	case ErrCodeMemoryStore:
		err.AddDetail("suggestion", "Check memory backend connection and storage capacity")
	case ErrCodeMemoryQuery:
		err.AddDetail("suggestion", "Verify query syntax and memory backend availability")
	case ErrCodeMemoryConnection:
		err.AddDetail("suggestion", "Check memory backend connection settings and network")
	case ErrCodeMemoryInvalidBackend:
		err.AddDetail("suggestion", "Use a supported memory backend: local, redis, postgres, etc.")
	}

	return err
}

// WorkflowError creates a workflow-related error with helpful suggestions
func WorkflowError(code ErrorCode, message string, cause error) *AgentError {
	err := NewAgentErrorWithError(code, message, cause)
	err.AddDetail("component", "Workflow")

	// Add helpful suggestions
	switch code {
	case ErrCodeWorkflowInvalid:
		err.AddDetail("suggestion", "Check workflow configuration and step definitions")
	case ErrCodeWorkflowStepFailed:
		err.AddDetail("suggestion", "Review step configuration and check agent availability")
	case ErrCodeWorkflowTimeout:
		err.AddDetail("suggestion", "Increase workflow timeout or optimize step execution")
	case ErrCodeWorkflowCycleDetected:
		err.AddDetail("suggestion", "Remove circular dependencies in workflow steps")
	case ErrCodeWorkflowMaxIterations:
		err.AddDetail("suggestion", "Increase max iterations or review loop termination condition")
	}

	return err
}

// ConfigError creates a configuration-related error with helpful suggestions
func ConfigError(code ErrorCode, message string, cause error) *AgentError {
	err := NewAgentErrorWithError(code, message, cause)
	err.AddDetail("component", "Config")

	// Add helpful suggestions
	switch code {
	case ErrConfigInvalid:
		err.AddDetail("suggestion", "Review configuration syntax and required fields")
	case ErrCodeConfigNotFound:
		err.AddDetail("suggestion", "Ensure configuration file exists at the specified path")
	case ErrCodeConfigParse:
		err.AddDetail("suggestion", "Check TOML syntax and structure")
	case ErrCodeConfigValidation:
		err.AddDetail("suggestion", "Review validation errors and fix configuration values")
	case ErrConfigMissing:
		err.AddDetail("suggestion", "Provide required configuration fields")
	}

	return err
}

// MCPError creates an MCP-related error with helpful suggestions
func MCPError(code ErrorCode, serverName string, message string, cause error) *AgentError {
	err := NewAgentErrorWithError(code, message, cause)
	err.AddDetail("component", "MCP")
	err.AddDetail("server_name", serverName)

	// Add helpful suggestions
	switch code {
	case ErrCodeMCPServerNotFound:
		err.AddDetail("suggestion", fmt.Sprintf("MCP server '%s' is not registered. Use ConnectMCP() to connect servers", serverName))
	case ErrCodeMCPConnection:
		err.AddDetail("suggestion", "Check MCP server address, port, and network connectivity")
	case ErrCodeMCPTimeout:
		err.AddDetail("suggestion", "Increase MCP timeout or check server responsiveness")
	case ErrCodeMCPServerUnhealthy:
		err.AddDetail("suggestion", "Check MCP server health and restart if necessary")
	}

	return err
}

// HandlerError creates a handler-related error with helpful suggestions
func HandlerError(code ErrorCode, message string, cause error) *AgentError {
	err := NewAgentErrorWithError(code, message, cause)
	err.AddDetail("component", "Handler")

	// Add helpful suggestions
	switch code {
	case ErrCodeHandlerFailed:
		err.AddDetail("suggestion", "Review handler implementation and error handling")
	case ErrCodeHandlerTimeout:
		err.AddDetail("suggestion", "Increase handler timeout or optimize handler logic")
	case ErrCodeHandlerPanic:
		err.AddDetail("suggestion", "Add panic recovery in handler or fix the panic cause")
	}

	return err
}

// NewAgentLifecycleError creates an agent lifecycle error with helpful suggestions
func NewAgentLifecycleError(code ErrorCode, message string, cause error) *AgentError {
	err := NewAgentErrorWithError(code, message, cause)
	err.AddDetail("component", "Agent")

	// Add helpful suggestions
	switch code {
	case ErrCodeAgentNotInitialized:
		err.AddDetail("suggestion", "Call Initialize() before using the agent")
	case ErrCodeAgentShutdown:
		err.AddDetail("suggestion", "Agent has been shutdown. Create a new agent instance")
	case ErrCodeAgentInvalidState:
		err.AddDetail("suggestion", "Check agent state and lifecycle methods")
	}

	return err
}

// NewValidationError creates a validation error with field context
func NewValidationError(field string, message string) *AgentError {
	err := NewAgentError(ErrCodeValidationFailed, message)
	err.AddDetail("component", "Validation")
	err.AddDetail("field", field)
	err.AddDetail("suggestion", "Check input values and ensure all required fields are provided")
	return err
}

// =============================================================================
// ERROR CHECKING HELPERS
// =============================================================================

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	var agErr *AgentError
	if errors.As(err, &agErr) {
		return agErr.Code == code
	}
	return false
}

// GetErrorCode returns the error code from an error, or empty string if not an AgentError
func GetErrorCode(err error) ErrorCode {
	var agErr *AgentError
	if errors.As(err, &agErr) {
		return agErr.Code
	}
	return ""
}

// GetErrorDetails returns the details from an error, or nil if not an AgentError
func GetErrorDetails(err error) map[string]interface{} {
	var agErr *AgentError
	if errors.As(err, &agErr) {
		return agErr.Details
	}
	return nil
}

// GetErrorSuggestion returns the suggestion from an error's details
func GetErrorSuggestion(err error) string {
	details := GetErrorDetails(err)
	if details != nil {
		if suggestion, ok := details["suggestion"].(string); ok {
			return suggestion
		}
	}
	return ""
}

// IsConfigError checks if an error is a configuration error
func IsConfigError(err error) bool {
	code := GetErrorCode(err)
	return strings.HasPrefix(string(code), "CONFIG_") || code == ErrConfigInvalid || code == ErrConfigMissing
}

// IsLLMError checks if an error is an LLM error
func IsLLMError(err error) bool {
	code := GetErrorCode(err)
	return strings.HasPrefix(string(code), "LLM_")
}

// IsToolError checks if an error is a tool error
func IsToolError(err error) bool {
	code := GetErrorCode(err)
	return strings.HasPrefix(string(code), "TOOL_")
}

// IsMemoryError checks if an error is a memory error
func IsMemoryError(err error) bool {
	code := GetErrorCode(err)
	return strings.HasPrefix(string(code), "MEMORY_")
}

// IsWorkflowError checks if an error is a workflow error
func IsWorkflowError(err error) bool {
	code := GetErrorCode(err)
	return strings.HasPrefix(string(code), "WORKFLOW_")
}

// IsMCPError checks if an error is an MCP error
func IsMCPError(err error) bool {
	code := GetErrorCode(err)
	return strings.HasPrefix(string(code), "MCP_")
}

// IsHandlerError checks if an error is a handler error
func IsHandlerError(err error) bool {
	code := GetErrorCode(err)
	return strings.HasPrefix(string(code), "HANDLER_")
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	code := GetErrorCode(err)
	switch code {
	case ErrCodeLLMTimeout, ErrCodeLLMRateLimited, ErrCodeLLMConnection:
		return true
	case ErrCodeToolTimeout, ErrCodeMCPTimeout:
		return true
	case ErrCodeMemoryConnection:
		return true
	case ErrCodeTimeout:
		return true
	default:
		return false
	}
}

// IsFatal checks if an error is fatal (non-recoverable)
func IsFatal(err error) bool {
	code := GetErrorCode(err)
	switch code {
	case ErrCodeLLMAuth, ErrCodeLLMQuotaExceeded:
		return true
	case ErrConfigInvalid, ErrCodeConfigNotFound:
		return true
	case ErrCodeAgentShutdown:
		return true
	default:
		return false
	}
}

// =============================================================================
// ERROR COLLECTION
// =============================================================================

// ErrorCollection represents a collection of errors
type ErrorCollection struct {
	Errors []*AgentError
}

// NewErrorCollection creates a new error collection
func NewErrorCollection() *ErrorCollection {
	return &ErrorCollection{
		Errors: make([]*AgentError, 0),
	}
}

// Add adds an error to the collection
func (ec *ErrorCollection) Add(err *AgentError) {
	if err != nil {
		ec.Errors = append(ec.Errors, err)
	}
}

// HasErrors returns true if the collection has any errors
func (ec *ErrorCollection) HasErrors() bool {
	return len(ec.Errors) > 0
}

// Error implements the error interface
func (ec *ErrorCollection) Error() string {
	if !ec.HasErrors() {
		return ""
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Multiple errors occurred (%d):", len(ec.Errors)))

	for i, err := range ec.Errors {
		parts = append(parts, fmt.Sprintf("  %d. %s", i+1, err.Error()))
	}

	return strings.Join(parts, "\n")
}

// First returns the first error in the collection
func (ec *ErrorCollection) First() *AgentError {
	if len(ec.Errors) == 0 {
		return nil
	}
	return ec.Errors[0]
}

// Filter filters errors by a predicate function
func (ec *ErrorCollection) Filter(predicate func(*AgentError) bool) []*AgentError {
	var filtered []*AgentError
	for _, err := range ec.Errors {
		if predicate(err) {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// HasFatal checks if the collection has any fatal errors
func (ec *ErrorCollection) HasFatal() bool {
	for _, err := range ec.Errors {
		if IsFatal(err) {
			return true
		}
	}
	return false
}

// =============================================================================
// EXAMPLE USAGE AND DOCUMENTATION
// =============================================================================

/*
Package errors provides a unified error handling system for AgenticGoKit vNext.

# Basic Error Creation

Create component-specific errors with helpful suggestions:

	// LLM error
	err := LLMError(ErrCodeLLMCallFailed, "Failed to call OpenAI API", originalErr)

	// Tool error
	err := ToolError(ErrCodeToolNotFound, "calculator", "Tool not found", nil)

	// Memory error
	err := MemoryError(ErrCodeMemoryStore, "Failed to store in memory", originalErr)

	// Workflow error
	err := WorkflowError(ErrCodeWorkflowStepFailed, "Step 2 failed", originalErr)

	// Config error
	err := ConfigError(ErrCodeConfigNotFound, "Config file not found", nil)

# Error Checking

Check for specific error types and codes:

	if IsLLMError(err) {
		// Handle LLM error
		suggestion := GetErrorSuggestion(err)
		fmt.Println("Suggestion:", suggestion)
	}

	if IsErrorCode(err, ErrCodeLLMRateLimited) {
		// Handle rate limiting specifically
		time.Sleep(time.Second * 5)
	}

	if IsRetryable(err) {
		// Retry the operation
		return retry(operation)
	}

	if IsFatal(err) {
		// Fatal error - don't retry
		log.Fatal(err)
	}

# Error Details

Access error details and suggestions:

	if agErr, ok := err.(*AgentError); ok {
		fmt.Printf("Code: %s\n", agErr.Code)
		fmt.Printf("Message: %s\n", agErr.Message)

		if suggestion, ok := agErr.Details["suggestion"].(string); ok {
			fmt.Printf("Suggestion: %s\n", suggestion)
		}

		if component, ok := agErr.Details["component"].(string); ok {
			fmt.Printf("Component: %s\n", component)
		}
	}

# Error Collections

Collect multiple errors:

	collection := NewErrorCollection()
	collection.Add(LLMError(ErrCodeLLMCallFailed, "Call 1 failed", nil))
	collection.Add(ToolError(ErrCodeToolExecute, "tool1", "Execution failed", nil))

	if collection.HasErrors() {
		fmt.Println(collection.Error())
	}

	if collection.HasFatal() {
		log.Fatal("Fatal errors encountered")
	}

# Common Patterns

## Retry on retryable errors

	func callWithRetry(fn func() error, maxRetries int) error {
		for attempt := 1; attempt <= maxRetries; attempt++ {
			err := fn()
			if err == nil {
				return nil
			}

			if !IsRetryable(err) {
				return err
			}

			if attempt < maxRetries {
				backoff := time.Duration(attempt*attempt) * time.Second
				time.Sleep(backoff)
			}
		}
		return fmt.Errorf("max retries exceeded")
	}

## Handle component-specific errors

	result, err := agent.Run(ctx, input)
	if err != nil {
		switch {
		case IsLLMError(err):
			log.Printf("LLM error: %v\nSuggestion: %s", err, GetErrorSuggestion(err))
		case IsToolError(err):
			log.Printf("Tool error: %v\nSuggestion: %s", err, GetErrorSuggestion(err))
		case IsMemoryError(err):
			log.Printf("Memory error: %v\nSuggestion: %s", err, GetErrorSuggestion(err))
		default:
			log.Printf("Error: %v", err)
		}
		return err
	}

## Validate and collect errors

	func validateAgentConfig(config *Config) error {
		collection := NewErrorCollection()

		if config.LLM.Provider == "" {
			collection.Add(ConfigError(ErrConfigMissing, "LLM provider not specified", nil))
		}

		if config.Name == "" {
			collection.Add(NewValidationError("name", "Agent name is required"))
		}

		if collection.HasErrors() {
			return collection
		}

		return nil
	}
*/

