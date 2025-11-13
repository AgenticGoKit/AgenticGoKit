// Package error_handling provides internal retry logic implementation for AgentFlow.
package error_handling

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

// RetrierImplementation implements retry logic with various strategies
type RetrierImplementation struct {
	policy *core.RetryPolicy
}

// NewRetrierImplementation creates a new retrier implementation with the given policy
func NewRetrierImplementation(policy *core.RetryPolicy) *RetrierImplementation {
	if policy == nil {
		policy = core.DefaultRetryPolicy()
	}

	return &RetrierImplementation{
		policy: policy,
	}
}

// Execute runs the function with retry logic
func (r *RetrierImplementation) Execute(ctx context.Context, fn func() error) *retryResultInternal {
	result := &retryResultInternal{ErrorHistory: make([]string, 0)}

	startTime := time.Now()

	for attempt := 1; attempt <= r.policy.MaxRetries+1; attempt++ {
		result.AttemptCount = attempt

		// Execute the function
		err := fn()

		if err == nil {
			result.Success = true
			result.TotalDuration = time.Since(startTime)
			return result
		}

		// Record the error
		result.LastError = err
		result.ErrorHistory = append(result.ErrorHistory, err.Error())

		// Check if we should retry
		if attempt > r.policy.MaxRetries {
			// No more retries
			result.TotalDuration = time.Since(startTime)
			break
		}

		// Check if error is retryable
		if !r.isRetryableError(err) {
			result.TotalDuration = time.Since(startTime)
			break
		}

		// Calculate delay for next attempt
		delay := r.calculateDelay(attempt)

		// no callback wiring in simplified adapter

		// Wait for the delay, respecting context cancellation
		select {
		case <-ctx.Done():
			result.LastError = ctx.Err()
			result.ErrorHistory = append(result.ErrorHistory, ctx.Err().Error())
			result.TotalDuration = time.Since(startTime)
			return result
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return result
}

// ExecuteWithCircuitBreaker combines retry logic with circuit breaker
func (r *RetrierImplementation) ExecuteWithCircuitBreaker(ctx context.Context, cb *CircuitBreakerImplementation, fn func() error) *retryResultInternal {
	wrappedFn := func() error {
		return cb.Call(fn)
	}

	return r.Execute(ctx, wrappedFn)
}

// isRetryableError checks if an error should be retried based on the policy
func (r *RetrierImplementation) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()

	// Check against retryable error codes
	for _, retryableCode := range r.policy.RetryableErrors {
		if containsErrorCode(errorStr, retryableCode) {
			return true
		}
	}

	return false
}

// calculateDelay calculates the delay for the next retry attempt
func (r *RetrierImplementation) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(r.policy.InitialDelay) * math.Pow(r.policy.BackoffFactor, float64(attempt-1))

	// Apply maximum delay limit
	if delay > float64(r.policy.MaxDelay) {
		delay = float64(r.policy.MaxDelay)
	}

	// Add jitter if enabled
	if r.policy.Jitter {
		jitterRange := delay * 0.1                     // 10% jitter
		jitter := (rand.Float64()*2 - 1) * jitterRange // Random value between -jitterRange and +jitterRange
		delay += jitter
	}

	// Ensure delay is not negative
	if delay < 0 {
		delay = float64(r.policy.InitialDelay)
	}

	return time.Duration(delay)
}

// containsErrorCode checks if an error message contains a specific error code
func containsErrorCode(errorMsg, errorCode string) bool {
	// Simple string matching - could be enhanced with more sophisticated logic
	return fmt.Sprintf("%v", errorMsg) == errorCode ||
		fmt.Sprintf("error code: %s", errorCode) == errorMsg ||
		fmt.Sprintf("code=%s", errorCode) == errorMsg
}

// RetryManagerImplementation manages multiple retriers and provides metrics
type RetryManagerImplementation struct {
	retriers map[string]*RetrierImplementation
}

// NewRetryManagerImplementation creates a new retry manager implementation
func NewRetryManagerImplementation() *RetryManagerImplementation {
	return &RetryManagerImplementation{
		retriers: make(map[string]*RetrierImplementation),
	}
}

// AddRetrier adds a named retrier to the manager
func (rm *RetryManagerImplementation) AddRetrier(name string, retrier *RetrierImplementation) {
	rm.retriers[name] = retrier
}

// GetRetrier returns a retrier by name
func (rm *RetryManagerImplementation) GetRetrier(name string) (*RetrierImplementation, bool) {
	retrier, exists := rm.retriers[name]
	return retrier, exists
}

// GetMetrics returns metrics for a named retrier
// Note: Metrics removed in this simplified internal manager

// Minimal internal result type used by adapter
type retryResultInternal struct {
	Success       bool
	AttemptCount  int
	LastError     error
	ErrorHistory  []string
	TotalDuration time.Duration
}

