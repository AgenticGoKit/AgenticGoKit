// Package core provides error routing validation and chain builder functionality for AgentFlow.
package core

import (
	"fmt"
	"strings"
	"time"
)

// ErrorRoutingValidator validates error routing configurations
type ErrorRoutingValidator struct {
	availableAgents map[string]bool
	requiredAgents  []string
}

// NewErrorRoutingValidator creates a new error routing validator
func NewErrorRoutingValidator(availableAgents map[string]AgentHandler) *ErrorRoutingValidator {
	agentMap := make(map[string]bool)
	for agentName := range availableAgents {
		agentMap[agentName] = true
	}

	return &ErrorRoutingValidator{
		availableAgents: agentMap,
		requiredAgents: []string{
			"error-handler",
			"validation-error-handler",
			"timeout-error-handler",
			"critical-error-handler",
		},
	}
}

// ValidationResult contains the results of error routing validation
type ValidationResult struct {
	IsValid         bool     `json:"is_valid"`
	Errors          []string `json:"errors"`
	Warnings        []string `json:"warnings"`
	MissingAgents   []string `json:"missing_agents"`
	UnusedAgents    []string `json:"unused_agents"`
	ConfiguredPaths int      `json:"configured_paths"`
}

// ValidateConfiguration validates an error router configuration
func (v *ErrorRoutingValidator) ValidateConfiguration(config *ErrorRouterConfig) *ValidationResult {
	result := &ValidationResult{
		IsValid:       true,
		Errors:        make([]string, 0),
		Warnings:      make([]string, 0),
		MissingAgents: make([]string, 0),
		UnusedAgents:  make([]string, 0),
	}

	if config == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "error router configuration is nil")
		return result
	}

	// Validate basic configuration
	v.validateBasicConfig(config, result)

	// Validate category handlers
	v.validateCategoryHandlers(config, result)

	// Validate severity handlers
	v.validateSeverityHandlers(config, result)

	// Validate default error handler
	v.validateDefaultErrorHandler(config, result)

	// Check for missing critical agents
	v.validateCriticalAgents(config, result)

	// Check for unused agents
	v.validateUnusedAgents(config, result)

	// Count configured paths
	result.ConfiguredPaths = len(config.CategoryHandlers) + len(config.SeverityHandlers)
	if config.ErrorHandlerName != "" {
		result.ConfiguredPaths++
	}

	return result
}

// validateBasicConfig validates basic configuration parameters
func (v *ErrorRoutingValidator) validateBasicConfig(config *ErrorRouterConfig, result *ValidationResult) {
	if config.MaxRetries < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "max_retries cannot be negative")
	}

	if config.MaxRetries > 10 {
		result.Warnings = append(result.Warnings, "max_retries is very high (>10), consider reducing")
	}

	if config.RetryDelayMs < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "retry_delay_ms cannot be negative")
	}

	if config.RetryDelayMs > 30000 {
		result.Warnings = append(result.Warnings, "retry_delay_ms is very high (>30s), consider reducing")
	}
}

// validateCategoryHandlers validates category-specific error handlers
func (v *ErrorRoutingValidator) validateCategoryHandlers(config *ErrorRouterConfig, result *ValidationResult) {
	validCategories := map[string]bool{
		ErrorCodeValidation: true,
		ErrorCodeTimeout:    true,
		ErrorCodeLLM:        true,
		ErrorCodeNetwork:    true,
		ErrorCodeAuth:       true,
		ErrorCodeResource:   true,
		ErrorCodeUnknown:    true,
	}

	for category, handler := range config.CategoryHandlers {
		// Check if category is valid
		if !validCategories[category] {
			result.Warnings = append(result.Warnings, fmt.Sprintf("unknown error category: %s", category))
		}

		// Check if handler agent exists
		if !v.availableAgents[handler] {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("category handler '%s' for category '%s' does not exist", handler, category))
			result.MissingAgents = append(result.MissingAgents, handler)
		}
	}

	// Check for missing critical category handlers
	criticalCategories := []string{ErrorCodeValidation, ErrorCodeTimeout}
	for _, category := range criticalCategories {
		if _, exists := config.CategoryHandlers[category]; !exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("missing handler for critical category: %s", category))
		}
	}
}

// validateSeverityHandlers validates severity-specific error handlers
func (v *ErrorRoutingValidator) validateSeverityHandlers(config *ErrorRouterConfig, result *ValidationResult) {
	validSeverities := map[string]bool{
		SeverityLow:      true,
		SeverityMedium:   true,
		SeverityHigh:     true,
		SeverityCritical: true,
	}

	for severity, handler := range config.SeverityHandlers {
		// Check if severity is valid
		if !validSeverities[severity] {
			result.Warnings = append(result.Warnings, fmt.Sprintf("unknown error severity: %s", severity))
		}

		// Check if handler agent exists
		if !v.availableAgents[handler] {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("severity handler '%s' for severity '%s' does not exist", handler, severity))
			result.MissingAgents = append(result.MissingAgents, handler)
		}
	}

	// Check for critical severity handler
	if _, exists := config.SeverityHandlers[SeverityCritical]; !exists {
		result.Warnings = append(result.Warnings, "missing handler for critical severity errors")
	}
}

// validateDefaultErrorHandler validates the default error handler
func (v *ErrorRoutingValidator) validateDefaultErrorHandler(config *ErrorRouterConfig, result *ValidationResult) {
	if config.ErrorHandlerName == "" {
		result.Warnings = append(result.Warnings, "no default error handler configured")
		return
	}

	if !v.availableAgents[config.ErrorHandlerName] {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("default error handler '%s' does not exist", config.ErrorHandlerName))
		result.MissingAgents = append(result.MissingAgents, config.ErrorHandlerName)
	}
}

// validateCriticalAgents checks for presence of critical error handling agents
func (v *ErrorRoutingValidator) validateCriticalAgents(config *ErrorRouterConfig, result *ValidationResult) {
	for _, requiredAgent := range v.requiredAgents {
		if !v.availableAgents[requiredAgent] {
			result.Warnings = append(result.Warnings, fmt.Sprintf("recommended agent '%s' is not available", requiredAgent))
		}
	}
}

// validateUnusedAgents identifies error handling agents that are not configured
func (v *ErrorRoutingValidator) validateUnusedAgents(config *ErrorRouterConfig, result *ValidationResult) {
	usedAgents := make(map[string]bool)

	// Mark category handlers as used
	for _, handler := range config.CategoryHandlers {
		usedAgents[handler] = true
	}

	// Mark severity handlers as used
	for _, handler := range config.SeverityHandlers {
		usedAgents[handler] = true
	}

	// Mark default handler as used
	if config.ErrorHandlerName != "" {
		usedAgents[config.ErrorHandlerName] = true
	}

	// Find unused error handling agents
	for agentName := range v.availableAgents {
		if strings.Contains(agentName, "error") && !usedAgents[agentName] {
			result.UnusedAgents = append(result.UnusedAgents, agentName)
		}
	}
}

// ErrorHandlingChain represents a chain of error handlers
type ErrorHandlingChain struct {
	handlers []ErrorHandlerLink
	config   *ErrorRouterConfig
}

// ErrorHandlerLink represents a single link in the error handling chain
type ErrorHandlerLink struct {
	Name        string             `json:"name"`
	Handler     AgentHandler       `json:"-"`
	Condition   ErrorConditionFunc `json:"-"`
	OnSuccess   ChainActionFunc    `json:"-"`
	OnFailure   ChainActionFunc    `json:"-"`
	MaxAttempts int                `json:"max_attempts"`
	Timeout     time.Duration      `json:"timeout"`
}

// ErrorConditionFunc determines if this handler should process the error
type ErrorConditionFunc func(errorData *ErrorEventData) bool

// ChainActionFunc defines actions to take after handler execution
type ChainActionFunc func(result AgentResult, errorData *ErrorEventData) ChainAction

// ChainAction represents what to do next in the chain
type ChainAction int

const (
	// ChainActionContinue - continue to next handler in chain
	ChainActionContinue ChainAction = iota
	// ChainActionStop - stop processing, error is resolved
	ChainActionStop
	// ChainActionRetry - retry current handler
	ChainActionRetry
	// ChainActionEscalate - escalate to higher priority handler
	ChainActionEscalate
)

// ErrorChainBuilder helps build error handling chains
type ErrorChainBuilder struct {
	chain    *ErrorHandlingChain
	registry map[string]AgentHandler
}

// NewErrorChainBuilder creates a new error chain builder
func NewErrorChainBuilder(registry map[string]AgentHandler) *ErrorChainBuilder {
	return &ErrorChainBuilder{
		chain: &ErrorHandlingChain{
			handlers: make([]ErrorHandlerLink, 0),
		},
		registry: registry,
	}
}

// AddHandler adds a handler to the chain
func (b *ErrorChainBuilder) AddHandler(name string) *ErrorHandlerBuilder {
	handler, exists := b.registry[name]
	if !exists {
		Logger().Warn().Str("handler_name", name).Msg("Handler not found in registry")
		handler = nil
	}

	link := ErrorHandlerLink{
		Name:        name,
		Handler:     handler,
		MaxAttempts: 1,
		Timeout:     30 * time.Second,
	}

	return &ErrorHandlerBuilder{
		chainBuilder: b,
		link:         &link,
	}
}

// Build creates the final error handling chain
func (b *ErrorChainBuilder) Build() *ErrorHandlingChain {
	return b.chain
}

// SetConfiguration sets the error router configuration for the chain
func (b *ErrorChainBuilder) SetConfiguration(config *ErrorRouterConfig) *ErrorChainBuilder {
	b.chain.config = config
	return b
}

// ErrorHandlerBuilder helps configure individual handlers in the chain
type ErrorHandlerBuilder struct {
	chainBuilder *ErrorChainBuilder
	link         *ErrorHandlerLink
}

// WithCondition sets the condition for when this handler should execute
func (hb *ErrorHandlerBuilder) WithCondition(condition ErrorConditionFunc) *ErrorHandlerBuilder {
	hb.link.Condition = condition
	return hb
}

// WithMaxAttempts sets the maximum attempts for this handler
func (hb *ErrorHandlerBuilder) WithMaxAttempts(attempts int) *ErrorHandlerBuilder {
	hb.link.MaxAttempts = attempts
	return hb
}

// WithTimeout sets the timeout for this handler
func (hb *ErrorHandlerBuilder) WithTimeout(timeout time.Duration) *ErrorHandlerBuilder {
	hb.link.Timeout = timeout
	return hb
}

// OnSuccess sets the action to take when handler succeeds
func (hb *ErrorHandlerBuilder) OnSuccess(action ChainActionFunc) *ErrorHandlerBuilder {
	hb.link.OnSuccess = action
	return hb
}

// OnFailure sets the action to take when handler fails
func (hb *ErrorHandlerBuilder) OnFailure(action ChainActionFunc) *ErrorHandlerBuilder {
	hb.link.OnFailure = action
	return hb
}

// Add finalizes this handler and adds it to the chain
func (hb *ErrorHandlerBuilder) Add() *ErrorChainBuilder {
	hb.chainBuilder.chain.handlers = append(hb.chainBuilder.chain.handlers, *hb.link)
	return hb.chainBuilder
}

// Common condition functions

// ValidationErrorCondition checks if error is a validation error
func ValidationErrorCondition(errorData *ErrorEventData) bool {
	return errorData.ErrorCategory == "validation" || errorData.ErrorCode == ErrorCodeValidation
}

// TimeoutErrorCondition checks if error is a timeout error
func TimeoutErrorCondition(errorData *ErrorEventData) bool {
	return errorData.ErrorCategory == "timeout" || errorData.ErrorCode == ErrorCodeTimeout
}

// CriticalErrorCondition checks if error is critical severity
func CriticalErrorCondition(errorData *ErrorEventData) bool {
	return errorData.Severity == SeverityCritical
}

// RetryCountCondition creates a condition based on retry count
func RetryCountCondition(maxRetries int) ErrorConditionFunc {
	return func(errorData *ErrorEventData) bool {
		return errorData.RetryCount < maxRetries
	}
}

// Common action functions

// StopOnSuccessAction stops the chain when handler succeeds
func StopOnSuccessAction(result AgentResult, errorData *ErrorEventData) ChainAction {
	return ChainActionStop
}

// ContinueOnSuccessAction continues the chain when handler succeeds
func ContinueOnSuccessAction(result AgentResult, errorData *ErrorEventData) ChainAction {
	return ChainActionContinue
}

// RetryOnFailureAction retries the handler when it fails
func RetryOnFailureAction(result AgentResult, errorData *ErrorEventData) ChainAction {
	return ChainActionRetry
}

// EscalateOnFailureAction escalates when handler fails
func EscalateOnFailureAction(result AgentResult, errorData *ErrorEventData) ChainAction {
	return ChainActionEscalate
}

// DefaultErrorChain creates a default error handling chain
func DefaultErrorChain(registry map[string]AgentHandler) *ErrorHandlingChain {
	builder := NewErrorChainBuilder(registry)

	return builder.
		// First: Try validation error handler for validation errors
		AddHandler("validation-error-handler").
		WithCondition(ValidationErrorCondition).
		WithMaxAttempts(2).
		OnSuccess(StopOnSuccessAction).
		OnFailure(ContinueOnSuccessAction).
		Add().

		// Second: Try timeout error handler for timeout errors
		AddHandler("timeout-error-handler").
		WithCondition(TimeoutErrorCondition).
		WithMaxAttempts(3).
		OnSuccess(StopOnSuccessAction).
		OnFailure(ContinueOnSuccessAction).
		Add().

		// Third: Try critical error handler for critical errors
		AddHandler("critical-error-handler").
		WithCondition(CriticalErrorCondition).
		WithMaxAttempts(1).
		OnSuccess(StopOnSuccessAction).
		OnFailure(EscalateOnFailureAction).
		Add().

		// Last: Default error handler for all other errors
		AddHandler("error-handler").
		WithCondition(func(errorData *ErrorEventData) bool { return true }).
		WithMaxAttempts(3).
		OnSuccess(StopOnSuccessAction).
		OnFailure(StopOnSuccessAction).
		Add().
		Build()
}
