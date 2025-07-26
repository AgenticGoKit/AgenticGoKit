package mcp

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/rs/zerolog"
)

// RetryPolicy defines the interface for retry policies
type RetryPolicy interface {
	// ShouldRetry determines if an operation should be retried
	ShouldRetry(attempt int, err error, duration time.Duration) bool

	// NextDelay calculates the delay before the next retry
	NextDelay(attempt int) time.Duration

	// Reset resets the retry policy state
	Reset()
}

// RetryClassification categorizes errors for retry decisions
type RetryClassification int

const (
	RetryableError RetryClassification = iota
	NonRetryableError
	ThrottledError
	NetworkError
	TimeoutError
)

// ClassifyError classifies an error for retry decisions
func ClassifyError(err error) RetryClassification {
	if err == nil {
		return NonRetryableError
	}

	errStr := err.Error()

	// Network-related errors
	if isNetworkError(errStr) {
		return NetworkError
	}

	// Timeout errors
	if isTimeoutError(errStr) {
		return TimeoutError
	}

	// Throttling errors
	if isThrottledError(errStr) {
		return ThrottledError
	}

	// Context cancellation is not retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return NonRetryableError
	}

	// Default to retryable for unknown errors
	return RetryableError
}

// isNetworkError checks if an error is network-related
func isNetworkError(errStr string) bool {
	networkKeywords := []string{
		"connection refused",
		"connection reset",
		"network unreachable",
		"no route to host",
		"connection timeout",
		"dial tcp",
		"i/o timeout",
	}

	for _, keyword := range networkKeywords {
		if containsIgnoreCase(errStr, keyword) {
			return true
		}
	}
	return false
}

// isTimeoutError checks if an error is timeout-related
func isTimeoutError(errStr string) bool {
	timeoutKeywords := []string{
		"timeout",
		"deadline exceeded",
		"context deadline exceeded",
	}

	for _, keyword := range timeoutKeywords {
		if containsIgnoreCase(errStr, keyword) {
			return true
		}
	}
	return false
}

// isThrottledError checks if an error indicates throttling
func isThrottledError(errStr string) bool {
	throttleKeywords := []string{
		"rate limit",
		"throttled",
		"too many requests",
		"quota exceeded",
		"rate exceeded",
	}

	for _, keyword := range throttleKeywords {
		if containsIgnoreCase(errStr, keyword) {
			return true
		}
	}
	return false
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					fmt.Sprintf(" %s ", s)[1:len(s)+1] != s))
}

// ExponentialBackoffPolicy implements exponential backoff with jitter
type ExponentialBackoffPolicy struct {
	BaseDelay      time.Duration
	MaxDelay       time.Duration
	Multiplier     float64
	Jitter         float64
	MaxAttempts    int
	RetryableTypes map[RetryClassification]bool
}

// NewExponentialBackoffPolicy creates a new exponential backoff policy
func NewExponentialBackoffPolicy(baseDelay, maxDelay time.Duration, maxAttempts int) *ExponentialBackoffPolicy {
	return &ExponentialBackoffPolicy{
		BaseDelay:   baseDelay,
		MaxDelay:    maxDelay,
		Multiplier:  2.0,
		Jitter:      0.1,
		MaxAttempts: maxAttempts,
		RetryableTypes: map[RetryClassification]bool{
			RetryableError: true,
			NetworkError:   true,
			TimeoutError:   true,
			ThrottledError: true,
		},
	}
}

// ShouldRetry determines if an operation should be retried
func (p *ExponentialBackoffPolicy) ShouldRetry(attempt int, err error, duration time.Duration) bool {
	if attempt >= p.MaxAttempts {
		return false
	}

	classification := ClassifyError(err)
	return p.RetryableTypes[classification]
}

// NextDelay calculates the delay before the next retry
func (p *ExponentialBackoffPolicy) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return p.BaseDelay
	}

	// Calculate exponential delay
	delay := float64(p.BaseDelay) * math.Pow(p.Multiplier, float64(attempt))

	// Add jitter
	if p.Jitter > 0 {
		jitterAmount := delay * p.Jitter
		jitter := (rand.Float64() - 0.5) * 2 * jitterAmount
		delay += jitter
	}

	// Cap at max delay
	if delay > float64(p.MaxDelay) {
		delay = float64(p.MaxDelay)
	}

	return time.Duration(delay)
}

// Reset resets the retry policy state
func (p *ExponentialBackoffPolicy) Reset() {
	// No state to reset for exponential backoff
}

// LinearBackoffPolicy implements linear backoff
type LinearBackoffPolicy struct {
	BaseDelay      time.Duration
	MaxDelay       time.Duration
	Increment      time.Duration
	MaxAttempts    int
	RetryableTypes map[RetryClassification]bool
}

// NewLinearBackoffPolicy creates a new linear backoff policy
func NewLinearBackoffPolicy(baseDelay, increment, maxDelay time.Duration, maxAttempts int) *LinearBackoffPolicy {
	return &LinearBackoffPolicy{
		BaseDelay:   baseDelay,
		MaxDelay:    maxDelay,
		Increment:   increment,
		MaxAttempts: maxAttempts,
		RetryableTypes: map[RetryClassification]bool{
			RetryableError: true,
			NetworkError:   true,
			TimeoutError:   true,
		},
	}
}

// ShouldRetry determines if an operation should be retried
func (p *LinearBackoffPolicy) ShouldRetry(attempt int, err error, duration time.Duration) bool {
	if attempt >= p.MaxAttempts {
		return false
	}

	classification := ClassifyError(err)
	return p.RetryableTypes[classification]
}

// NextDelay calculates the delay before the next retry
func (p *LinearBackoffPolicy) NextDelay(attempt int) time.Duration {
	delay := p.BaseDelay + time.Duration(attempt)*p.Increment
	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	return delay
}

// Reset resets the retry policy state
func (p *LinearBackoffPolicy) Reset() {
	// No state to reset for linear backoff
}

// ToolSpecificRetryPolicy allows different retry policies per tool
type ToolSpecificRetryPolicy struct {
	DefaultPolicy  RetryPolicy
	ToolPolicies   map[string]RetryPolicy
	ServerPolicies map[string]RetryPolicy
}

// NewToolSpecificRetryPolicy creates a new tool-specific retry policy
func NewToolSpecificRetryPolicy(defaultPolicy RetryPolicy) *ToolSpecificRetryPolicy {
	return &ToolSpecificRetryPolicy{
		DefaultPolicy:  defaultPolicy,
		ToolPolicies:   make(map[string]RetryPolicy),
		ServerPolicies: make(map[string]RetryPolicy),
	}
}

// SetToolPolicy sets a retry policy for a specific tool
func (p *ToolSpecificRetryPolicy) SetToolPolicy(toolName string, policy RetryPolicy) {
	p.ToolPolicies[toolName] = policy
}

// SetServerPolicy sets a retry policy for a specific server
func (p *ToolSpecificRetryPolicy) SetServerPolicy(serverID string, policy RetryPolicy) {
	p.ServerPolicies[serverID] = policy
}

// GetPolicy returns the appropriate retry policy for a tool/server combination
func (p *ToolSpecificRetryPolicy) GetPolicy(toolName, serverID string) RetryPolicy {
	// Check for tool-specific policy first
	if policy, exists := p.ToolPolicies[toolName]; exists {
		return policy
	}

	// Check for server-specific policy
	if policy, exists := p.ServerPolicies[serverID]; exists {
		return policy
	}

	// Return default policy
	return p.DefaultPolicy
}

// RetryExecutor handles retry logic with circuit breaker integration
type RetryExecutor struct {
	policy         RetryPolicy
	circuitBreaker *core.CircuitBreaker
	logger         *zerolog.Logger
	metrics        *RetryMetrics
}

// RetryMetrics tracks retry statistics
type RetryMetrics struct {
	TotalAttempts     int64
	TotalRetries      int64
	SuccessfulRetries int64
	FailedRetries     int64
	AverageAttempts   float64
}

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(policy RetryPolicy, circuitBreaker *core.CircuitBreaker, logger *zerolog.Logger) *RetryExecutor {
	return &RetryExecutor{
		policy:         policy,
		circuitBreaker: circuitBreaker,
		logger:         logger,
		metrics:        &RetryMetrics{},
	}
}

// Execute executes a function with retry logic
func (r *RetryExecutor) Execute(ctx context.Context, operation func() error) error {
	r.policy.Reset()

	for attempt := 0; ; attempt++ {
		r.metrics.TotalAttempts++

		start := time.Now()
		err := operation()
		duration := time.Since(start)
		if err == nil {
			if attempt > 0 {
				r.metrics.SuccessfulRetries++
				r.logger.Info().
					Int("attempts", attempt+1).
					Dur("duration", duration).
					Msg("Operation succeeded after retries")
			}
			return nil
		}
		// Check circuit breaker state
		if r.circuitBreaker != nil {
			// Test if circuit breaker would allow a call by checking if canCall would succeed
			if err := r.circuitBreaker.Call(func() error { return nil }); err != nil {
				return fmt.Errorf("circuit breaker blocking request: %w", err)
			}
		}

		// Check if we should retry
		if !r.policy.ShouldRetry(attempt, err, duration) {
			if attempt > 0 {
				r.metrics.FailedRetries++
			}
			return fmt.Errorf("retry policy exhausted after %d attempts: %w", attempt+1, err)
		}

		// Calculate delay
		delay := r.policy.NextDelay(attempt)
		r.metrics.TotalRetries++
		r.logger.Debug().
			Int("attempt", attempt+1).
			Str("error", err.Error()).
			Dur("delay", delay).
			Int("error_type", int(ClassifyError(err))).
			Msg("Retrying operation")

		// Wait for delay or context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
}

// ExecuteWithResult executes a function with retry logic and returns a result
func (r *RetryExecutor) ExecuteWithResult(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
	var result interface{}

	err := r.Execute(ctx, func() error {
		var err error
		result, err = operation()
		return err
	})

	return result, err
}

// GetMetrics returns retry metrics
func (r *RetryExecutor) GetMetrics() *RetryMetrics {
	avgAttempts := float64(r.metrics.TotalAttempts) / float64(r.metrics.TotalRetries+1)

	return &RetryMetrics{
		TotalAttempts:     r.metrics.TotalAttempts,
		TotalRetries:      r.metrics.TotalRetries,
		SuccessfulRetries: r.metrics.SuccessfulRetries,
		FailedRetries:     r.metrics.FailedRetries,
		AverageAttempts:   avgAttempts,
	}
}

// AdaptiveRetryPolicy adjusts retry behavior based on success/failure patterns
type AdaptiveRetryPolicy struct {
	basePolicy         RetryPolicy
	successWindow      []bool // Track recent successes/failures
	windowSize         int
	windowIndex        int
	successRate        float64
	adaptiveMultiplier float64
}

// NewAdaptiveRetryPolicy creates a new adaptive retry policy
func NewAdaptiveRetryPolicy(basePolicy RetryPolicy, windowSize int) *AdaptiveRetryPolicy {
	return &AdaptiveRetryPolicy{
		basePolicy:         basePolicy,
		successWindow:      make([]bool, windowSize),
		windowSize:         windowSize,
		windowIndex:        0,
		successRate:        1.0,
		adaptiveMultiplier: 1.0,
	}
}

// RecordResult records the result of an operation
func (p *AdaptiveRetryPolicy) RecordResult(success bool) {
	p.successWindow[p.windowIndex] = success
	p.windowIndex = (p.windowIndex + 1) % p.windowSize

	// Calculate success rate
	successCount := 0
	for _, s := range p.successWindow {
		if s {
			successCount++
		}
	}
	p.successRate = float64(successCount) / float64(p.windowSize)

	// Adjust adaptive multiplier based on success rate
	if p.successRate > 0.8 {
		p.adaptiveMultiplier = 0.5 // Reduce delays when success rate is high
	} else if p.successRate < 0.3 {
		p.adaptiveMultiplier = 2.0 // Increase delays when success rate is low
	} else {
		p.adaptiveMultiplier = 1.0 // Normal delays
	}
}

// ShouldRetry determines if an operation should be retried
func (p *AdaptiveRetryPolicy) ShouldRetry(attempt int, err error, duration time.Duration) bool {
	return p.basePolicy.ShouldRetry(attempt, err, duration)
}

// NextDelay calculates the delay before the next retry with adaptive adjustment
func (p *AdaptiveRetryPolicy) NextDelay(attempt int) time.Duration {
	baseDelay := p.basePolicy.NextDelay(attempt)
	adaptedDelay := time.Duration(float64(baseDelay) * p.adaptiveMultiplier)
	return adaptedDelay
}

// Reset resets the retry policy state
func (p *AdaptiveRetryPolicy) Reset() {
	p.basePolicy.Reset()
}

// GetSuccessRate returns the current success rate
func (p *AdaptiveRetryPolicy) GetSuccessRate() float64 {
	return p.successRate
}

// DefaultRetryPolicies provides common retry policy configurations
type DefaultRetryPolicies struct{}

// QuickRetry returns a policy for quick operations that should retry fast
func (DefaultRetryPolicies) QuickRetry() RetryPolicy {
	return NewExponentialBackoffPolicy(
		100*time.Millisecond, // Base delay
		5*time.Second,        // Max delay
		3,                    // Max attempts
	)
}

// StandardRetry returns a policy for standard operations
func (DefaultRetryPolicies) StandardRetry() RetryPolicy {
	return NewExponentialBackoffPolicy(
		1*time.Second,  // Base delay
		30*time.Second, // Max delay
		5,              // Max attempts
	)
}

// SlowRetry returns a policy for slow operations that can tolerate longer delays
func (DefaultRetryPolicies) SlowRetry() RetryPolicy {
	return NewExponentialBackoffPolicy(
		5*time.Second, // Base delay
		2*time.Minute, // Max delay
		3,             // Max attempts
	)
}

// NetworkRetry returns a policy optimized for network operations
func (DefaultRetryPolicies) NetworkRetry() RetryPolicy {
	policy := NewExponentialBackoffPolicy(
		500*time.Millisecond, // Base delay
		10*time.Second,       // Max delay
		4,                    // Max attempts
	)

	// Only retry network and timeout errors
	policy.RetryableTypes = map[RetryClassification]bool{
		NetworkError: true,
		TimeoutError: true,
	}

	return policy
}

// ThrottleRetry returns a policy optimized for handling throttling
func (DefaultRetryPolicies) ThrottleRetry() RetryPolicy {
	policy := NewLinearBackoffPolicy(
		2*time.Second,  // Base delay
		1*time.Second,  // Increment
		30*time.Second, // Max delay
		3,              // Max attempts
	)

	// Only retry throttling errors
	policy.RetryableTypes = map[RetryClassification]bool{
		ThrottledError: true,
	}

	return policy
}
