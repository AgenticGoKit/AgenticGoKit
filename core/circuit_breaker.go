// Package core provides circuit breaker functionality for AgentFlow error handling.
package core

import (
	"time"
)

// =============================================================================
// CIRCUIT BREAKER TYPES AND CONSTANTS
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

// =============================================================================
// CIRCUIT BREAKER CONFIGURATION
// =============================================================================

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold   int           `json:"failure_threshold"`    // Number of failures before opening
	SuccessThreshold   int           `json:"success_threshold"`    // Number of successes to close from half-open
	Timeout            time.Duration `json:"timeout"`              // How long to wait before transitioning to half-open
	MaxConcurrentCalls int           `json:"max_concurrent_calls"` // Maximum concurrent calls in half-open state
}

// DefaultCircuitBreakerConfig returns a sensible default configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold:   5,
		SuccessThreshold:   3,
		Timeout:            30 * time.Second,
		MaxConcurrentCalls: 2,
	}
}

// =============================================================================
// CIRCUIT BREAKER METRICS
// =============================================================================

// CircuitBreakerMetrics contains metrics about circuit breaker performance
type CircuitBreakerMetrics struct {
	State           CircuitBreakerState `json:"state"`
	FailureCount    int                 `json:"failure_count"`
	SuccessCount    int                 `json:"success_count"`
	LastFailureTime time.Time           `json:"last_failure_time"`
	ConcurrentCalls int                 `json:"concurrent_calls"`
}

// =============================================================================
// CIRCUIT BREAKER INTERFACE AND FACTORY
// =============================================================================

// CircuitBreaker defines the interface for circuit breaker functionality
type CircuitBreaker interface {
	Call(fn func() error) error
	SetStateChangeCallback(callback func(from, to CircuitBreakerState))
	GetState() CircuitBreakerState
	GetMetrics() CircuitBreakerMetrics
}

// CircuitBreakerFactory is the function signature for creating circuit breakers
type CircuitBreakerFactory func(config *CircuitBreakerConfig) CircuitBreaker

// circuitBreakerFactory holds the registered factory function
var circuitBreakerFactory CircuitBreakerFactory

// RegisterCircuitBreakerFactory registers the circuit breaker factory function
func RegisterCircuitBreakerFactory(factory CircuitBreakerFactory) {
	circuitBreakerFactory = factory
}

// NewCircuitBreaker creates a circuit breaker from configuration
// This function requires the internal error handling factory to be registered
func NewCircuitBreaker(config *CircuitBreakerConfig) CircuitBreaker {
	if circuitBreakerFactory == nil {
		// Return a basic circuit breaker implementation if no factory is registered
		return &basicCircuitBreaker{config: config}
	}
	return circuitBreakerFactory(config)
}

// =============================================================================
// BASIC CIRCUIT BREAKER IMPLEMENTATION (FALLBACK)
// =============================================================================

// basicCircuitBreaker provides a minimal implementation when no factory is registered
type basicCircuitBreaker struct {
	config *CircuitBreakerConfig
	state  CircuitBreakerState
}

func (bcb *basicCircuitBreaker) Call(fn func() error) error {
	// Basic implementation - just call the function
	return fn()
}

func (bcb *basicCircuitBreaker) SetStateChangeCallback(callback func(from, to CircuitBreakerState)) {
	// Basic implementation - do nothing
}

func (bcb *basicCircuitBreaker) GetState() CircuitBreakerState {
	return bcb.state
}

func (bcb *basicCircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	return CircuitBreakerMetrics{
		State: bcb.state,
	}
}
