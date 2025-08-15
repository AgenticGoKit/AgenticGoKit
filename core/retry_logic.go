// Package core provides essential retry logic types for AgentFlow error handling.
package core

import (
	"context"
	"time"
)

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
// PUBLIC FACTORY FUNCTIONS
// =============================================================================

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
// Implementation is provided by internal packages
func NewRetryHandler(policy *RetryPolicy) RetryHandler {
	if retryHandlerFactory != nil {
		return retryHandlerFactory(policy)
	}
	// Return a basic implementation - internal packages can register better implementations
	return &basicRetryHandler{policy: policy}
}

// RegisterRetryHandlerFactory registers the retry handler factory function
func RegisterRetryHandlerFactory(factory func(policy *RetryPolicy) RetryHandler) {
	retryHandlerFactory = factory
}

// =============================================================================
// INTERNAL IMPLEMENTATION
// =============================================================================

var retryHandlerFactory func(policy *RetryPolicy) RetryHandler

// basicRetryHandler provides a minimal implementation
type basicRetryHandler struct {
	policy *RetryPolicy
}

func (rh *basicRetryHandler) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	// Basic implementation - internal packages can provide more sophisticated implementations
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
	// Basic implementation - internal packages can provide more sophisticated logic
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
	// Basic exponential backoff implementation
	delay := time.Duration(float64(rh.policy.InitialDelay) * float64(attempt) * rh.policy.BackoffFactor)
	if delay > rh.policy.MaxDelay {
		delay = rh.policy.MaxDelay
	}
	return delay
}