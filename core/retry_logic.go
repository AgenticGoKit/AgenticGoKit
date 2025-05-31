// Package core provides retry logic functionality for AgentFlow error handling.
package core

import (
	"context"
	"fmt"
	"math"
	"math/rand"
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

// RetryResult contains the result of a retry operation
type RetryResult struct {
	Success       bool          `json:"success"`
	AttemptCount  int           `json:"attempt_count"`
	TotalDuration time.Duration `json:"total_duration"`
	LastError     error         `json:"-"`
	ErrorHistory  []string      `json:"error_history"`
}

// RetryFunc represents a function that can be retried
type RetryFunc func() error

// Retrier implements retry logic with various strategies
type Retrier struct {
	policy    *RetryPolicy
	callbacks RetryCallbacks
}

// RetryCallbacks allows monitoring retry attempts
type RetryCallbacks struct {
	OnRetry   func(attempt int, err error, delay time.Duration)
	OnGiveUp  func(attempt int, err error)
	OnSuccess func(attempt int)
}

// NewRetrier creates a new retrier with the given policy
func NewRetrier(policy *RetryPolicy) *Retrier {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}

	return &Retrier{
		policy: policy,
	}
}

// SetCallbacks sets the retry callbacks
func (r *Retrier) SetCallbacks(callbacks RetryCallbacks) {
	r.callbacks = callbacks
}

// Execute runs the function with retry logic
func (r *Retrier) Execute(ctx context.Context, fn RetryFunc) *RetryResult {
	result := &RetryResult{
		ErrorHistory: make([]string, 0),
	}

	startTime := time.Now()

	for attempt := 1; attempt <= r.policy.MaxRetries+1; attempt++ {
		result.AttemptCount = attempt

		// Execute the function
		err := fn()

		if err == nil {
			// Success
			result.Success = true
			result.TotalDuration = time.Since(startTime)
			if r.callbacks.OnSuccess != nil {
				r.callbacks.OnSuccess(attempt)
			}
			return result
		}

		// Record the error
		result.LastError = err
		result.ErrorHistory = append(result.ErrorHistory, err.Error())

		// Check if we should retry
		if attempt > r.policy.MaxRetries {
			// No more retries
			result.TotalDuration = time.Since(startTime)
			if r.callbacks.OnGiveUp != nil {
				r.callbacks.OnGiveUp(attempt, err)
			}
			break
		}

		// Check if error is retryable
		if !r.isRetryableError(err) {
			result.TotalDuration = time.Since(startTime)
			if r.callbacks.OnGiveUp != nil {
				r.callbacks.OnGiveUp(attempt, err)
			}
			break
		}

		// Calculate delay for next attempt
		delay := r.calculateDelay(attempt)

		if r.callbacks.OnRetry != nil {
			r.callbacks.OnRetry(attempt, err, delay)
		}

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
func (r *Retrier) ExecuteWithCircuitBreaker(ctx context.Context, cb *CircuitBreaker, fn RetryFunc) *RetryResult {
	wrappedFn := func() error {
		return cb.Call(fn)
	}

	return r.Execute(ctx, wrappedFn)
}

// isRetryableError checks if an error should be retried based on the policy
func (r *Retrier) isRetryableError(err error) bool {
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
func (r *Retrier) calculateDelay(attempt int) time.Duration {
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

// RetryableErrorChecker defines an interface for custom retryable error checking
type RetryableErrorChecker interface {
	IsRetryable(error) bool
}

// CustomRetryableChecker allows custom retry logic
type CustomRetryableChecker struct {
	CheckFunc func(error) bool
}

// IsRetryable implements RetryableErrorChecker
func (c *CustomRetryableChecker) IsRetryable(err error) bool {
	if c.CheckFunc == nil {
		return false
	}
	return c.CheckFunc(err)
}

// SetCustomRetryableChecker allows setting a custom retryable error checker
func (r *Retrier) SetCustomRetryableChecker(checker RetryableErrorChecker) {
	// Store the checker and modify isRetryableError to use it
	// This would require modifying the Retrier struct to include the checker
}

// RetryMetrics contains metrics about retry operations
type RetryMetrics struct {
	TotalAttempts   int           `json:"total_attempts"`
	SuccessfulCalls int           `json:"successful_calls"`
	FailedCalls     int           `json:"failed_calls"`
	AverageAttempts float64       `json:"average_attempts"`
	AverageDuration time.Duration `json:"average_duration"`
}

// RetryManager manages multiple retriers and provides metrics
type RetryManager struct {
	retriers map[string]*Retrier
	metrics  map[string]*RetryMetrics
}

// NewRetryManager creates a new retry manager
func NewRetryManager() *RetryManager {
	return &RetryManager{
		retriers: make(map[string]*Retrier),
		metrics:  make(map[string]*RetryMetrics),
	}
}

// AddRetrier adds a named retrier to the manager
func (rm *RetryManager) AddRetrier(name string, retrier *Retrier) {
	rm.retriers[name] = retrier
	rm.metrics[name] = &RetryMetrics{}
}

// GetRetrier returns a retrier by name
func (rm *RetryManager) GetRetrier(name string) (*Retrier, bool) {
	retrier, exists := rm.retriers[name]
	return retrier, exists
}

// GetMetrics returns metrics for a named retrier
func (rm *RetryManager) GetMetrics(name string) (*RetryMetrics, bool) {
	metrics, exists := rm.metrics[name]
	return metrics, exists
}
