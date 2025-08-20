// Package core provides essential error handling types for AgentFlow.
package core

import (
	"context"
	"time"
)

// =============================================================================
// ERROR EVENT DATA AND CONSTANTS
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

// =============================================================================
// CIRCUIT BREAKER TYPES
// =============================================================================

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// CircuitBreakerClosed - normal operation, requests are allowed
	CircuitBreakerClosed CircuitBreakerState = iota
	// CircuitBreakerOpen - circuit is open, requests are rejected
	CircuitBreakerOpen
	// CircuitBreakerHalfOpen - testing state, limited requests are allowed
	CircuitBreakerHalfOpen
)

// String returns a string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerClosed:
		return "CLOSED"
	case CircuitBreakerOpen:
		return "OPEN"
	case CircuitBreakerHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold   int           `json:"failure_threshold"`    // Number of failures before opening
	SuccessThreshold   int           `json:"success_threshold"`    // Number of successes to close from half-open
	Timeout            time.Duration `json:"timeout"`              // How long to wait before transitioning to half-open
	MaxConcurrentCalls int           `json:"max_concurrent_calls"` // Maximum concurrent calls in half-open state
}

// CircuitBreakerMetrics contains metrics about circuit breaker performance
type CircuitBreakerMetrics struct {
	State           CircuitBreakerState `json:"state"`
	FailureCount    int                 `json:"failure_count"`
	SuccessCount    int                 `json:"success_count"`
	LastFailureTime time.Time           `json:"last_failure_time"`
	ConcurrentCalls int                 `json:"concurrent_calls"`
}

// CircuitBreaker defines the interface for circuit breaker functionality
type CircuitBreaker interface {
	Call(fn func() error) error
	SetStateChangeCallback(callback func(from, to CircuitBreakerState))
	GetState() CircuitBreakerState
	GetMetrics() CircuitBreakerMetrics
}

// =============================================================================
// RETRY POLICY TYPES
// =============================================================================

// RetryPolicy defines the retry behavior configuration
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`      // Maximum number of retry attempts
	InitialDelay    time.Duration `json:"initial_delay"`    // Initial delay before first retry
	MaxDelay        time.Duration `json:"max_delay"`        // Maximum delay between retries
	BackoffFactor   float64       `json:"backoff_factor"`   // Exponential backoff multiplier
	Jitter          bool          `json:"jitter"`           // Add random jitter to delays
	RetryableErrors []string      `json:"retryable_errors"` // List of error codes that are retryable
}

// RetryHandler defines the interface for retry functionality
type RetryHandler interface {
	ExecuteWithRetry(ctx context.Context, operation func() error) error
	ShouldRetry(attempt int, err error) bool
	CalculateDelay(attempt int) time.Duration
}

// =============================================================================
// ERROR ROUTER TYPES
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

// ErrorRouter defines the interface for error routing functionality
type ErrorRouter interface {
	CreateEnhancedErrorEvent(originalEvent Event, agentID string, err error) Event
	IsRetryableError(errorData ErrorEventData) bool
	IncrementRetryCount(event Event) Event
}

// =============================================================================
// PUBLIC FACTORY FUNCTIONS
// =============================================================================

// DefaultCircuitBreakerConfig returns a sensible default configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold:   5,
		SuccessThreshold:   3,
		Timeout:            30 * time.Second,
		MaxConcurrentCalls: 2,
	}
}

// NewCircuitBreaker creates a circuit breaker from configuration
func NewCircuitBreaker(config *CircuitBreakerConfig) CircuitBreaker {
	if circuitBreakerFactory != nil {
		return circuitBreakerFactory(config)
	}
	return &basicCircuitBreaker{config: config}
}

// RegisterCircuitBreakerFactory registers the circuit breaker factory function
func RegisterCircuitBreakerFactory(factory func(config *CircuitBreakerConfig) CircuitBreaker) {
	circuitBreakerFactory = factory
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		RetryableErrors: []string{
			ErrorCodeTimeout,
			ErrorCodeNetwork,
			ErrorCodeLLM,
		},
	}
}

// NewRetryHandler creates a retry handler from policy
func NewRetryHandler(policy *RetryPolicy) RetryHandler {
	if retryHandlerFactory != nil {
		return retryHandlerFactory(policy)
	}
	return &basicRetryHandler{policy: policy}
}

// RegisterRetryHandlerFactory registers the retry handler factory function
func RegisterRetryHandlerFactory(factory func(policy *RetryPolicy) RetryHandler) {
	retryHandlerFactory = factory
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

// NewErrorRouter creates an error router from configuration
func NewErrorRouter(config *ErrorRouterConfig) ErrorRouter {
	if errorRouterFactory != nil {
		return errorRouterFactory(config)
	}
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
// INTERNAL IMPLEMENTATIONS
// =============================================================================

var (
	circuitBreakerFactory func(config *CircuitBreakerConfig) CircuitBreaker
	retryHandlerFactory   func(policy *RetryPolicy) RetryHandler
	errorRouterFactory    func(config *ErrorRouterConfig) ErrorRouter
)

// basicCircuitBreaker provides a minimal implementation
type basicCircuitBreaker struct {
	config *CircuitBreakerConfig
	state  CircuitBreakerState
}

func (bcb *basicCircuitBreaker) Call(fn func() error) error {
	return fn()
}

func (bcb *basicCircuitBreaker) SetStateChangeCallback(callback func(from, to CircuitBreakerState)) {
	// Basic implementation - do nothing
}

func (bcb *basicCircuitBreaker) GetState() CircuitBreakerState {
	return bcb.state
}

func (bcb *basicCircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	return CircuitBreakerMetrics{State: bcb.state}
}

// basicRetryHandler provides a minimal implementation
type basicRetryHandler struct {
	policy *RetryPolicy
}

func (rh *basicRetryHandler) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	var lastErr error
	for attempt := 0; attempt <= rh.policy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := rh.CalculateDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		lastErr = operation()
		if lastErr == nil {
			return nil
		}

		if !rh.ShouldRetry(attempt, lastErr) {
			break
		}
	}
	return lastErr
}

func (rh *basicRetryHandler) ShouldRetry(attempt int, err error) bool {
	if attempt >= rh.policy.MaxRetries {
		return false
	}

	errorStr := err.Error()
	for _, retryableError := range rh.policy.RetryableErrors {
		if errorStr == retryableError {
			return true
		}
	}
	return false
}

func (rh *basicRetryHandler) CalculateDelay(attempt int) time.Duration {
	delay := time.Duration(float64(rh.policy.InitialDelay) * float64(attempt) * rh.policy.BackoffFactor)
	if delay > rh.policy.MaxDelay {
		delay = rh.policy.MaxDelay
	}
	return delay
}

// basicErrorRouter provides a minimal implementation
type basicErrorRouter struct {
	config *ErrorRouterConfig
}

func (ber *basicErrorRouter) CreateEnhancedErrorEvent(originalEvent Event, agentID string, err error) Event {
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
	switch errorData.ErrorCode {
	case ErrorCodeTimeout, ErrorCodeNetwork, ErrorCodeLLM:
		return errorData.RetryCount < 3
	default:
		return false
	}
}

func (ber *basicErrorRouter) IncrementRetryCount(event Event) Event {
	newMetadata := make(map[string]string)
	for k, v := range event.GetMetadata() {
		newMetadata[k] = v
	}

	newEvent := NewEvent(event.GetTargetAgentID(), event.GetData(), newMetadata)
	newEvent.SetSourceAgentID(event.GetSourceAgentID())
	return newEvent
}