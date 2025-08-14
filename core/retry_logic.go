// Package core provides retry logic functionality for AgentFlow error handling.
package core

import (
	"context"
	"time"
)

// =============================================================================
// RETRY POLICY AND CONFIGURATION
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

// =============================================================================
// RETRY TYPES AND INTERFACES
// =============================================================================

// RetryFunc represents a function that can be retried
type RetryFunc func() error

// RetryResult contains the result of a retry operation
type RetryResult struct {
	Success       bool          `json:"success"`
	AttemptCount  int           `json:"attempt_count"`
	TotalDuration time.Duration `json:"total_duration"`
	LastError     error         `json:"-"`
	ErrorHistory  []string      `json:"error_history"`
}

// RetryCallbacks allows monitoring retry attempts
type RetryCallbacks struct {
	OnRetry   func(attempt int, err error, delay time.Duration)
	OnGiveUp  func(attempt int, err error)
	OnSuccess func(attempt int)
}

// RetryMetrics contains metrics about retry operations
type RetryMetrics struct {
	TotalAttempts   int           `json:"total_attempts"`
	SuccessfulCalls int           `json:"successful_calls"`
	FailedCalls     int           `json:"failed_calls"`
	AverageAttempts float64       `json:"average_attempts"`
	AverageDuration time.Duration `json:"average_duration"`
}

// =============================================================================
// RETRIER INTERFACE AND FACTORY
// =============================================================================

// Retrier defines the interface for retry functionality
type Retrier interface {
	Execute(ctx context.Context, fn RetryFunc) *RetryResult
	SetCallbacks(callbacks RetryCallbacks)
}

// RetryManager defines the interface for managing multiple retriers
type RetryManager interface {
	AddRetrier(name string, retrier Retrier)
	GetRetrier(name string) (Retrier, bool)
	GetMetrics(name string) (*RetryMetrics, bool)
}

// RetrierFactory is the function signature for creating retriers
type RetrierFactory func(policy *RetryPolicy) Retrier

// RetryManagerFactory is the function signature for creating retry managers
type RetryManagerFactory func() RetryManager

// retrierFactory holds the registered factory function
var retrierFactory RetrierFactory

// retryManagerFactory holds the registered factory function
var retryManagerFactory RetryManagerFactory

// RegisterRetrierFactory registers the retrier factory function
func RegisterRetrierFactory(factory RetrierFactory) {
	retrierFactory = factory
}

// RegisterRetryManagerFactory registers the retry manager factory function
func RegisterRetryManagerFactory(factory RetryManagerFactory) {
	retryManagerFactory = factory
}

// NewRetrier creates a retrier from policy
// This function requires the internal error handling factory to be registered
func NewRetrier(policy *RetryPolicy) Retrier {
	if retrierFactory == nil {
		// Return a basic retrier implementation if no factory is registered
		return &basicRetrier{policy: policy}
	}
	return retrierFactory(policy)
}

// NewRetryManager creates a retry manager
// This function requires the internal error handling factory to be registered
func NewRetryManager() RetryManager {
	if retryManagerFactory == nil {
		// Return a basic retry manager implementation if no factory is registered
		return &basicRetryManager{retriers: make(map[string]Retrier)}
	}
	return retryManagerFactory()
}

// =============================================================================
// BASIC IMPLEMENTATIONS (FALLBACK)
// =============================================================================

// basicRetrier provides a minimal implementation when no factory is registered
type basicRetrier struct {
	policy    *RetryPolicy
	callbacks RetryCallbacks
}

func (br *basicRetrier) Execute(ctx context.Context, fn RetryFunc) *RetryResult {
	// Basic implementation - just try once
	err := fn()
	return &RetryResult{
		Success:      err == nil,
		AttemptCount: 1,
		LastError:    err,
	}
}

func (br *basicRetrier) SetCallbacks(callbacks RetryCallbacks) {
	br.callbacks = callbacks
}

// basicRetryManager provides a minimal implementation when no factory is registered
type basicRetryManager struct {
	retriers map[string]Retrier
}

func (brm *basicRetryManager) AddRetrier(name string, retrier Retrier) {
	brm.retriers[name] = retrier
}

func (brm *basicRetryManager) GetRetrier(name string) (Retrier, bool) {
	retrier, exists := brm.retriers[name]
	return retrier, exists
}

func (brm *basicRetryManager) GetMetrics(name string) (*RetryMetrics, bool) {
	// Basic implementation - return empty metrics
	return &RetryMetrics{}, false
}
