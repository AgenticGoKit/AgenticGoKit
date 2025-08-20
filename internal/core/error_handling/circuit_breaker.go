// Package error_handling provides internal circuit breaker implementation for AgentFlow.
package error_handling

import (
	"fmt"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// CircuitBreakerImplementation implements the circuit breaker pattern for error handling
type CircuitBreakerImplementation struct {
	config          *core.CircuitBreakerConfig
	state           core.CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	concurrentCalls int
	mu              sync.RWMutex
	onStateChange   func(from, to core.CircuitBreakerState)
}

// NewCircuitBreakerImplementation creates a new circuit breaker implementation
func NewCircuitBreakerImplementation(config *core.CircuitBreakerConfig) *CircuitBreakerImplementation {
	if config == nil {
		config = core.DefaultCircuitBreakerConfig()
	}

	return &CircuitBreakerImplementation{
		config: config,
		state:  core.CircuitBreakerClosed,
	}
}

// SetStateChangeCallback sets a callback function that is called when the circuit breaker state changes
func (cb *CircuitBreakerImplementation) SetStateChangeCallback(callback func(from, to core.CircuitBreakerState)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = callback
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreakerImplementation) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we can make the call
	if err := cb.canCall(); err != nil {
		return err
	}

	// Increment concurrent calls for half-open state
	if cb.state == core.CircuitBreakerHalfOpen {
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
	if cb.state == core.CircuitBreakerHalfOpen {
		cb.concurrentCalls--
	}

	return err
}

// canCall checks if a call can be made based on current state
func (cb *CircuitBreakerImplementation) canCall() error {
	switch cb.state {
	case core.CircuitBreakerClosed:
		return nil
	case core.CircuitBreakerOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailureTime) >= cb.config.Timeout {
			cb.setState(core.CircuitBreakerHalfOpen)
			cb.successCount = 0
			cb.concurrentCalls = 0
			return nil
		}
		return fmt.Errorf("circuit breaker is open")
	case core.CircuitBreakerHalfOpen:
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
func (cb *CircuitBreakerImplementation) onSuccess() {
	switch cb.state {
	case core.CircuitBreakerClosed:
		// Reset failure count on success
		cb.failureCount = 0
	case core.CircuitBreakerHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.setState(core.CircuitBreakerClosed)
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// onFailure handles failed call completion
func (cb *CircuitBreakerImplementation) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case core.CircuitBreakerClosed:
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.setState(core.CircuitBreakerOpen)
		}
	case core.CircuitBreakerHalfOpen:
		cb.setState(core.CircuitBreakerOpen)
		cb.successCount = 0
	}
}

// setState changes the circuit breaker state and calls the callback if set
func (cb *CircuitBreakerImplementation) setState(newState core.CircuitBreakerState) {
	oldState := cb.state
	cb.state = newState

	if cb.onStateChange != nil && oldState != newState {
		cb.onStateChange(oldState, newState)
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreakerImplementation) GetState() core.CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreakerImplementation) GetMetrics() core.CircuitBreakerMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return core.CircuitBreakerMetrics{
		State:           cb.state,
		FailureCount:    cb.failureCount,
		SuccessCount:    cb.successCount,
		LastFailureTime: cb.lastFailureTime,
		ConcurrentCalls: cb.concurrentCalls,
	}
}
