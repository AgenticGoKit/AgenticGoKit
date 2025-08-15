// Package core provides essential error routing types for AgentFlow.
package core

import (
	"time"
)

// =============================================================================
// ESSENTIAL ERROR TYPES
// =============================================================================

// ErrorEventData represents structured error information for enhanced error routing
type ErrorEventData struct {
	OriginalEvent  Event     `json:"original_event"`
	FailedAgent    string    `json:"failed_agent"`
	ErrorMessage   string    `json:"error_message"`
	ErrorCode      string    `json:"error_code"`
	RetryCount     int       `json:"retry_count"`
	Timestamp      time.Time `json:"timestamp"`
	SessionID      string    `json:"session_id"`
	Severity       string    `json:"severity"`        // "low", "medium", "high", "critical"
	ErrorCategory  string    `json:"error_category"`  // "validation", "timeout", "llm", "network", "unknown"
	RecoveryAction string    `json:"recovery_action"` // "retry", "fallback", "escalate", "terminate"
}

// Error code constants for consistent error categorization
const (
	ErrorCodeValidation = "VALIDATION_ERROR"
	ErrorCodeTimeout    = "TIMEOUT_ERROR"
	ErrorCodeLLM        = "LLM_ERROR"
	ErrorCodeNetwork    = "NETWORK_ERROR"
	ErrorCodeAuth       = "AUTH_ERROR"
	ErrorCodeResource   = "RESOURCE_ERROR"
	ErrorCodeUnknown    = "UNKNOWN_ERROR"
)

// Error severity constants
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Recovery action constants
const (
	RecoveryRetry     = "retry"
	RecoveryFallback  = "fallback"
	RecoveryEscalate  = "escalate"
	RecoveryTerminate = "terminate"
)

// ErrorRouterConfig configures the enhanced error routing behavior
type ErrorRouterConfig struct {
	MaxRetries           int               `json:"max_retries"`
	RetryDelayMs         int               `json:"retry_delay_ms"`
	EnableCircuitBreaker bool              `json:"enable_circuit_breaker"`
	ErrorHandlerName     string            `json:"error_handler_name"`
	CategoryHandlers     map[string]string `json:"category_handlers"` // Maps error categories to specific handlers
	SeverityHandlers     map[string]string `json:"severity_handlers"` // Maps severity levels to specific handlers
}

// ErrorRouter defines the interface for error routing functionality
type ErrorRouter interface {
	CreateEnhancedErrorEvent(originalEvent Event, agentID string, err error) Event
	IsRetryableError(errorData ErrorEventData) bool
	IncrementRetryCount(event Event) Event
}

// =============================================================================
// PUBLIC FACTORY FUNCTIONS
// =============================================================================

// DefaultErrorRouterConfig returns a sensible default configuration
func DefaultErrorRouterConfig() *ErrorRouterConfig {
	return &ErrorRouterConfig{
		MaxRetries:           3,
		RetryDelayMs:         1000,
		EnableCircuitBreaker: true,
		ErrorHandlerName:     "error-handler",
		CategoryHandlers: map[string]string{
			ErrorCodeValidation: "validation-error-handler",
			ErrorCodeTimeout:    "timeout-error-handler",
			ErrorCodeLLM:        "llm-error-handler",
			ErrorCodeNetwork:    "network-error-handler",
			ErrorCodeAuth:       "auth-error-handler",
		},
		SeverityHandlers: map[string]string{
			SeverityCritical: "critical-error-handler",
			SeverityHigh:     "high-priority-error-handler",
		},
	}
}

// NewErrorRouter creates an error router from configuration
// Implementation is provided by internal packages
func NewErrorRouter(config *ErrorRouterConfig) ErrorRouter {
	if errorRouterFactory != nil {
		return errorRouterFactory(config)
	}
	// Return a basic implementation - internal packages can register better implementations
	return &basicErrorRouter{config: config}
}

// RegisterErrorRouterFactory registers the error router factory function
func RegisterErrorRouterFactory(factory func(config *ErrorRouterConfig) ErrorRouter) {
	errorRouterFactory = factory
}

// Convenience functions for error handling
func CreateEnhancedErrorEvent(originalEvent Event, agentID string, err error, config *ErrorRouterConfig) Event {
	router := NewErrorRouter(config)
	return router.CreateEnhancedErrorEvent(originalEvent, agentID, err)
}

func IsRetryableError(errorData ErrorEventData) bool {
	router := NewErrorRouter(nil)
	return router.IsRetryableError(errorData)
}

func IncrementRetryCount(event Event) Event {
	router := NewErrorRouter(nil)
	return router.IncrementRetryCount(event)
}

// =============================================================================
// INTERNAL IMPLEMENTATION
// =============================================================================

var errorRouterFactory func(config *ErrorRouterConfig) ErrorRouter

// basicErrorRouter provides a minimal implementation
type basicErrorRouter struct {
	config *ErrorRouterConfig
}

func (ber *basicErrorRouter) CreateEnhancedErrorEvent(originalEvent Event, agentID string, err error) Event {
	// Minimal implementation - internal packages can provide more sophisticated implementations
	payload := EventData{
		"error":        err.Error(),
		"failed_agent": agentID,
	}

	metadata := map[string]string{
		RouteMetadataKey: "error-handler",
		"status":         "error",
	}

	errorEvent := NewEvent("error-handler", payload, metadata)
	errorEvent.SetSourceAgentID(agentID)
	return errorEvent
}

func (ber *basicErrorRouter) IsRetryableError(errorData ErrorEventData) bool {
	// Basic implementation - internal packages can provide more sophisticated logic
	switch errorData.ErrorCode {
	case ErrorCodeTimeout, ErrorCodeNetwork, ErrorCodeLLM:
		return errorData.RetryCount < 3
	default:
		return false
	}
}

func (ber *basicErrorRouter) IncrementRetryCount(event Event) Event {
	// Basic implementation - internal packages can provide more sophisticated logic
	newMetadata := make(map[string]string)
	for k, v := range event.GetMetadata() {
		newMetadata[k] = v
	}

	newEvent := NewEvent(event.GetTargetAgentID(), event.GetData(), newMetadata)
	newEvent.SetSourceAgentID(event.GetSourceAgentID())
	return newEvent
}