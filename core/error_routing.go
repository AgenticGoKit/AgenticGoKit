// Package core provides enhanced error routing capabilities for AgentFlow.
package core

import (
	"time"
)

// =============================================================================
// ERROR EVENT DATA AND TYPES
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

// =============================================================================
// ERROR CODE CONSTANTS
// =============================================================================

// ErrorCode constants for consistent error categorization
const (
	ErrorCodeValidation = "VALIDATION_ERROR"
	ErrorCodeTimeout    = "TIMEOUT_ERROR"
	ErrorCodeLLM        = "LLM_ERROR"
	ErrorCodeNetwork    = "NETWORK_ERROR"
	ErrorCodeAuth       = "AUTH_ERROR"
	ErrorCodeResource   = "RESOURCE_ERROR"
	ErrorCodeUnknown    = "UNKNOWN_ERROR"
)

// ErrorSeverity constants
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// RecoveryAction constants
const (
	RecoveryRetry     = "retry"
	RecoveryFallback  = "fallback"
	RecoveryEscalate  = "escalate"
	RecoveryTerminate = "terminate"
)

// =============================================================================
// ERROR ROUTER CONFIGURATION
// =============================================================================

// ErrorRouterConfig configures the enhanced error routing behavior
type ErrorRouterConfig struct {
	MaxRetries           int               `json:"max_retries"`
	RetryDelayMs         int               `json:"retry_delay_ms"`
	EnableCircuitBreaker bool              `json:"enable_circuit_breaker"`
	ErrorHandlerName     string            `json:"error_handler_name"`
	CategoryHandlers     map[string]string `json:"category_handlers"` // Maps error categories to specific handlers
	SeverityHandlers     map[string]string `json:"severity_handlers"` // Maps severity levels to specific handlers
}

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

// =============================================================================
// ERROR ROUTING INTERFACE AND FACTORY
// =============================================================================

// ErrorRouter defines the interface for error routing functionality
type ErrorRouter interface {
	CreateEnhancedErrorEvent(originalEvent Event, agentID string, err error) Event
	IsRetryableError(errorData ErrorEventData) bool
	IncrementRetryCount(event Event) Event
}

// ErrorRouterFactory is the function signature for creating error routers
type ErrorRouterFactory func(config *ErrorRouterConfig) ErrorRouter

// errorRouterFactory holds the registered factory function
var errorRouterFactory ErrorRouterFactory

// RegisterErrorRouterFactory registers the error router factory function
func RegisterErrorRouterFactory(factory ErrorRouterFactory) {
	errorRouterFactory = factory
}

// NewErrorRouter creates an error router from configuration
// This function requires the internal error handling factory to be registered
func NewErrorRouter(config *ErrorRouterConfig) ErrorRouter {
	if errorRouterFactory == nil {
		// Return a basic error router implementation if no factory is registered
		return &basicErrorRouter{config: config}
	}
	return errorRouterFactory(config)
}

// =============================================================================
// PUBLIC CONVENIENCE FUNCTIONS
// =============================================================================

// CreateEnhancedErrorEvent creates a structured error event with enhanced metadata
func CreateEnhancedErrorEvent(originalEvent Event, agentID string, err error, config *ErrorRouterConfig) Event {
	router := NewErrorRouter(config)
	return router.CreateEnhancedErrorEvent(originalEvent, agentID, err)
}

// IsRetryableError determines if an error should be retried based on its characteristics
func IsRetryableError(errorData ErrorEventData) bool {
	router := NewErrorRouter(nil)
	return router.IsRetryableError(errorData)
}

// IncrementRetryCount creates a new event with incremented retry count
func IncrementRetryCount(event Event) Event {
	router := NewErrorRouter(nil)
	return router.IncrementRetryCount(event)
}

// =============================================================================
// BASIC ERROR ROUTER IMPLEMENTATION (FALLBACK)
// =============================================================================

// basicErrorRouter provides a minimal implementation when no factory is registered
type basicErrorRouter struct {
	config *ErrorRouterConfig
}

func (ber *basicErrorRouter) CreateEnhancedErrorEvent(originalEvent Event, agentID string, err error) Event {
	// Minimal implementation - just create a basic error event
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
	// Basic implementation - retry on certain error codes
	switch errorData.ErrorCode {
	case ErrorCodeTimeout, ErrorCodeNetwork, ErrorCodeLLM:
		return errorData.RetryCount < 3
	default:
		return false
	}
}

func (ber *basicErrorRouter) IncrementRetryCount(event Event) Event {
	// Basic implementation - just clone the event
	newMetadata := make(map[string]string)
	for k, v := range event.GetMetadata() {
		newMetadata[k] = v
	}

	newEvent := NewEvent(event.GetTargetAgentID(), event.GetData(), newMetadata)
	newEvent.SetSourceAgentID(event.GetSourceAgentID())
	return newEvent
}
