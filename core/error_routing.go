// Package core provides enhanced error routing capabilities for AgentFlow.
package core

import (
	"fmt"
	"time"
)

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

// CreateEnhancedErrorEvent creates a structured error event with enhanced metadata
func CreateEnhancedErrorEvent(originalEvent Event, agentID string, err error, config *ErrorRouterConfig) Event {
	if config == nil {
		config = DefaultErrorRouterConfig()
	}

	sessionID, _ := originalEvent.GetMetadataValue(SessionIDKey)

	// Categorize the error
	errorCode, severity, category := categorizeError(err)

	// Determine recovery action based on error type and retry count
	retryCount := getRetryCount(originalEvent)
	recoveryAction := determineRecoveryAction(errorCode, retryCount, config.MaxRetries)

	errorData := ErrorEventData{
		OriginalEvent:  originalEvent,
		FailedAgent:    agentID,
		ErrorMessage:   err.Error(),
		ErrorCode:      errorCode,
		RetryCount:     retryCount,
		Timestamp:      time.Now(),
		SessionID:      sessionID,
		Severity:       severity,
		ErrorCategory:  category,
		RecoveryAction: recoveryAction,
	}

	// Create event payload
	payload := EventData{
		"error_data":        errorData,
		"original_event_id": originalEvent.GetID(),
		"error":             err.Error(),
		"failed_agent":      agentID,
		"retry_count":       retryCount,
		"error_code":        errorCode,
		"severity":          severity,
		"recovery_action":   recoveryAction,
	}

	// Determine target handler based on configuration
	targetHandler := determineErrorHandler(errorData, config)

	// Create metadata
	metadata := map[string]string{
		SessionIDKey:      sessionID,
		RouteMetadataKey:  targetHandler,
		"status":          "error",
		"error_code":      errorCode,
		"severity":        severity,
		"recovery_action": recoveryAction,
	}

	if agentID != "" && agentID != "unknown" {
		metadata["failed_agent_id"] = agentID
	}

	errorEvent := NewEvent(targetHandler, payload, metadata)
	errorEvent.SetSourceAgentID(agentID)

	return errorEvent
}

// categorizeError analyzes an error and returns error code, severity, and category
func categorizeError(err error) (errorCode, severity, category string) {
	errorMsg := err.Error()

	// Basic error categorization based on error message patterns
	// In a real implementation, this could be more sophisticated
	switch {
	case containsAny(errorMsg, []string{"validation", "invalid", "required", "missing"}):
		return ErrorCodeValidation, SeverityMedium, "validation"
	case containsAny(errorMsg, []string{"timeout", "deadline", "context canceled"}):
		return ErrorCodeTimeout, SeverityHigh, "timeout"
	case containsAny(errorMsg, []string{"llm", "openai", "azure", "model", "completion"}):
		return ErrorCodeLLM, SeverityMedium, "llm"
	case containsAny(errorMsg, []string{"network", "connection", "dial", "http"}):
		return ErrorCodeNetwork, SeverityHigh, "network"
	case containsAny(errorMsg, []string{"auth", "unauthorized", "forbidden", "token"}):
		return ErrorCodeAuth, SeverityCritical, "auth"
	case containsAny(errorMsg, []string{"memory", "resource", "limit", "quota"}):
		return ErrorCodeResource, SeverityCritical, "resource"
	default:
		return ErrorCodeUnknown, SeverityMedium, "unknown"
	}
}

// containsAny checks if a string contains any of the provided substrings
func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// getRetryCount extracts retry count from event metadata or returns 0
func getRetryCount(event Event) int {
	if retryStr, ok := event.GetMetadataValue("retry_count"); ok {
		// Simple conversion - in real implementation might use strconv
		switch retryStr {
		case "1":
			return 1
		case "2":
			return 2
		case "3":
			return 3
		default:
			return 0
		}
	}
	return 0
}

// determineRecoveryAction decides what recovery action to take based on error and retry count
func determineRecoveryAction(errorCode string, retryCount, maxRetries int) string {
	switch errorCode {
	case ErrorCodeAuth, ErrorCodeResource:
		return RecoveryEscalate // Don't retry auth or resource errors
	case ErrorCodeValidation:
		return RecoveryTerminate // Don't retry validation errors
	case ErrorCodeTimeout, ErrorCodeNetwork, ErrorCodeLLM:
		if retryCount >= maxRetries {
			return RecoveryFallback
		}
		return RecoveryRetry
	default:
		if retryCount >= maxRetries {
			return RecoveryEscalate
		}
		return RecoveryRetry
	}
}

// determineErrorHandler selects the appropriate error handler based on configuration
func determineErrorHandler(errorData ErrorEventData, config *ErrorRouterConfig) string {
	// First check for severity-specific handlers
	if handler, exists := config.SeverityHandlers[errorData.Severity]; exists {
		return handler
	}

	// Then check for category-specific handlers
	if handler, exists := config.CategoryHandlers[errorData.ErrorCode]; exists {
		return handler
	}

	// Fall back to default error handler
	return config.ErrorHandlerName
}

// IsRetryableError determines if an error should be retried based on its characteristics
func IsRetryableError(errorData ErrorEventData) bool {
	switch errorData.RecoveryAction {
	case RecoveryRetry:
		return true
	case RecoveryFallback, RecoveryEscalate, RecoveryTerminate:
		return false
	default:
		return false
	}
}

// IncrementRetryCount creates a new event with incremented retry count
func IncrementRetryCount(event Event) Event {
	retryCount := getRetryCount(event) + 1

	// Clone the event with updated retry count
	newMetadata := make(map[string]string)
	for k, v := range event.GetMetadata() {
		newMetadata[k] = v
	}
	newMetadata["retry_count"] = fmt.Sprintf("%d", retryCount)

	newEvent := NewEvent(event.GetTargetAgentID(), event.GetData(), newMetadata)
	newEvent.SetSourceAgentID(event.GetSourceAgentID())

	return newEvent
}
