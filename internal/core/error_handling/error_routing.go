// Package error_handling provides internal error routing implementation for AgentFlow.
package error_handling

import (
	"fmt"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

// ErrorRoutingImplementation provides the concrete implementation of error routing
type ErrorRoutingImplementation struct {
	config *core.ErrorRouterConfig
}

// NewErrorRoutingImplementation creates a new error routing implementation
func NewErrorRoutingImplementation(config *core.ErrorRouterConfig) *ErrorRoutingImplementation {
	if config == nil {
		config = core.DefaultErrorRouterConfig()
	}
	return &ErrorRoutingImplementation{
		config: config,
	}
}

// CreateEnhancedErrorEvent creates a structured error event with enhanced metadata
func (eri *ErrorRoutingImplementation) CreateEnhancedErrorEvent(originalEvent core.Event, agentID string, err error) core.Event {
	sessionID, _ := originalEvent.GetMetadataValue(core.SessionIDKey)

	// Categorize the error
	errorCode, severity, category := eri.categorizeError(err)

	// Determine recovery action based on error type and retry count
	retryCount := eri.getRetryCount(originalEvent)
	recoveryAction := eri.determineRecoveryAction(errorCode, retryCount, eri.config.MaxRetries)

	errorData := core.ErrorEventData{
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
	payload := core.EventData{
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
	targetHandler := eri.determineErrorHandler(errorData)

	// Create metadata
	metadata := map[string]string{
		core.SessionIDKey:     sessionID,
		core.RouteMetadataKey: targetHandler,
		"status":              "error",
		"error_code":          errorCode,
		"severity":            severity,
		"recovery_action":     recoveryAction,
	}

	if agentID != "" && agentID != "unknown" {
		metadata["failed_agent_id"] = agentID
	}

	errorEvent := core.NewEvent(targetHandler, payload, metadata)
	errorEvent.SetSourceAgentID(agentID)

	return errorEvent
}

// categorizeError analyzes an error and returns error code, severity, and category
func (eri *ErrorRoutingImplementation) categorizeError(err error) (errorCode, severity, category string) {
	errorMsg := err.Error()

	// Basic error categorization based on error message patterns
	switch {
	case containsAny(errorMsg, []string{"validation", "invalid", "required", "missing"}):
		return core.ErrorCodeValidation, core.SeverityMedium, "validation"
	case containsAny(errorMsg, []string{"timeout", "deadline", "context canceled"}):
		return core.ErrorCodeTimeout, core.SeverityHigh, "timeout"
	case containsAny(errorMsg, []string{"llm", "openai", "azure", "model", "completion"}):
		return core.ErrorCodeLLM, core.SeverityMedium, "llm"
	case containsAny(errorMsg, []string{"network", "connection", "dial", "http"}):
		return core.ErrorCodeNetwork, core.SeverityHigh, "network"
	case containsAny(errorMsg, []string{"auth", "unauthorized", "forbidden", "token"}):
		return core.ErrorCodeAuth, core.SeverityCritical, "auth"
	case containsAny(errorMsg, []string{"memory", "resource", "limit", "quota"}):
		return core.ErrorCodeResource, core.SeverityCritical, "resource"
	default:
		return core.ErrorCodeUnknown, core.SeverityMedium, "unknown"
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
func (eri *ErrorRoutingImplementation) getRetryCount(event core.Event) int {
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
func (eri *ErrorRoutingImplementation) determineRecoveryAction(errorCode string, retryCount, maxRetries int) string {
	switch errorCode {
	case core.ErrorCodeAuth, core.ErrorCodeResource:
		return core.RecoveryEscalate // Don't retry auth or resource errors
	case core.ErrorCodeValidation:
		return core.RecoveryTerminate // Don't retry validation errors
	case core.ErrorCodeTimeout, core.ErrorCodeNetwork, core.ErrorCodeLLM:
		if retryCount >= maxRetries {
			return core.RecoveryFallback
		}
		return core.RecoveryRetry
	default:
		if retryCount >= maxRetries {
			return core.RecoveryEscalate
		}
		return core.RecoveryRetry
	}
}

// determineErrorHandler selects the appropriate error handler based on configuration
func (eri *ErrorRoutingImplementation) determineErrorHandler(errorData core.ErrorEventData) string {
	// First check for severity-specific handlers
	if handler, exists := eri.config.SeverityHandlers[errorData.Severity]; exists {
		return handler
	}

	// Then check for category-specific handlers
	if handler, exists := eri.config.CategoryHandlers[errorData.ErrorCode]; exists {
		return handler
	}

	// Fall back to default error handler
	return eri.config.ErrorHandlerName
}

// IsRetryableError determines if an error should be retried based on its characteristics
func (eri *ErrorRoutingImplementation) IsRetryableError(errorData core.ErrorEventData) bool {
	switch errorData.RecoveryAction {
	case core.RecoveryRetry:
		return true
	case core.RecoveryFallback, core.RecoveryEscalate, core.RecoveryTerminate:
		return false
	default:
		return false
	}
}

// IncrementRetryCount creates a new event with incremented retry count
func (eri *ErrorRoutingImplementation) IncrementRetryCount(event core.Event) core.Event {
	retryCount := eri.getRetryCount(event) + 1

	// Clone the event with updated retry count
	newMetadata := make(map[string]string)
	for k, v := range event.GetMetadata() {
		newMetadata[k] = v
	}
	newMetadata["retry_count"] = fmt.Sprintf("%d", retryCount)

	newEvent := core.NewEvent(event.GetTargetAgentID(), event.GetData(), newMetadata)
	newEvent.SetSourceAgentID(event.GetSourceAgentID())

	return newEvent
}

