// Package core provides circuit breaker functionality for AgentFlow error handling.
package core

import (
	"fmt"
	"sync"
	"time"
)

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

// CircuitBreaker implements the circuit breaker pattern for error handling
type CircuitBreaker struct {
	config          *CircuitBreakerConfig
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	concurrentCalls int
	mu              sync.RWMutex
	onStateChange   func(from, to CircuitBreakerState)
}

// NewCircuitBreaker creates a new circuit breaker instance
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

// SetStateChangeCallback sets a callback function that is called when the circuit breaker state changes
func (cb *CircuitBreaker) SetStateChangeCallback(callback func(from, to CircuitBreakerState)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = callback
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we can make the call
	if err := cb.canCall(); err != nil {
		return err
	}

	// Increment concurrent calls for half-open state
	if cb.state == CircuitBreakerHalfOpen {
		cb.concurrentCalls++
	}

	// Release the lock while executing the function
	cb.mu.Unlock()

	// Execute the function
	err := fn()

	// Re-acquire the lock to update state
	cb.mu.Lock()

	// Update state based on result
	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}

	// Decrement concurrent calls for half-open state
	if cb.state == CircuitBreakerHalfOpen {
		cb.concurrentCalls--
	}

	return err
}

// canCall checks if a call can be made based on current state
func (cb *CircuitBreaker) canCall() error {
	switch cb.state {
	case CircuitBreakerClosed:
		return nil
	case CircuitBreakerOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailureTime) >= cb.config.Timeout {
			cb.setState(CircuitBreakerHalfOpen)
			cb.successCount = 0
			cb.concurrentCalls = 0
			return nil
		}
		return fmt.Errorf("circuit breaker is open")
	case CircuitBreakerHalfOpen:
		// Check if we can make another concurrent call
		if cb.concurrentCalls >= cb.config.MaxConcurrentCalls {
			return fmt.Errorf("circuit breaker is half-open and at max concurrent calls")
		}
		return nil
	default:
		return fmt.Errorf("unknown circuit breaker state")
	}
}

// onSuccess handles successful call completion
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case CircuitBreakerClosed:
		// Reset failure count on success
		cb.failureCount = 0
	case CircuitBreakerHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.setState(CircuitBreakerClosed)
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// onFailure handles failed call completion
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitBreakerClosed:
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.setState(CircuitBreakerOpen)
		}
	case CircuitBreakerHalfOpen:
		cb.setState(CircuitBreakerOpen)
		cb.successCount = 0
	}
}

// setState changes the circuit breaker state and calls the callback if set
func (cb *CircuitBreaker) setState(newState CircuitBreakerState) {
	oldState := cb.state
	cb.state = newState

	if cb.onStateChange != nil && oldState != newState {
		cb.onStateChange(oldState, newState)
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerMetrics{
		State:           cb.state,
		FailureCount:    cb.failureCount,
		SuccessCount:    cb.successCount,
		LastFailureTime: cb.lastFailureTime,
		ConcurrentCalls: cb.concurrentCalls,
	}
}

// CircuitBreakerMetrics contains metrics about circuit breaker performance
type CircuitBreakerMetrics struct {
	State           CircuitBreakerState `json:"state"`
	FailureCount    int                 `json:"failure_count"`
	SuccessCount    int                 `json:"success_count"`
	LastFailureTime time.Time           `json:"last_failure_time"`
	ConcurrentCalls int                 `json:"concurrent_calls"`
}

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
