// Package core provides integration tests for circuit breaker and retry logic functionality.
package core

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestCircuitBreakerBasicFunctionality tests the circuit breaker functionality
func TestCircuitBreakerBasicFunctionality(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold:   2,
		SuccessThreshold:   2,
		Timeout:            1 * time.Second,
		MaxConcurrentCalls: 2,
	}

	cb := NewCircuitBreaker(config)

	// Test successful calls
	for i := 0; i < 3; i++ {
		if cb.GetState() != CircuitBreakerClosed {
			t.Errorf("Expected circuit breaker to be closed, got %v", cb.GetState())
		}

		err := cb.Call(func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}

	// Test failures to trigger circuit opening
	for i := 0; i < 2; i++ {
		err := cb.Call(func() error {
			return errors.New("failure")
		})

		if err == nil {
			t.Error("Expected error, got nil")
		}
	}

	// Circuit should be open now
	if cb.GetState() != CircuitBreakerOpen {
		t.Errorf("Expected circuit breaker to be open, got %v", cb.GetState())
	}

	// Test that calls are rejected when circuit is open
	err := cb.Call(func() error {
		return nil
	})

	if err == nil || !strings.Contains(err.Error(), "circuit breaker is open") {
		t.Errorf("Expected circuit breaker open error, got %v", err)
	}

	// Wait for reset timeout
	time.Sleep(config.Timeout + 100*time.Millisecond)

	// Circuit should transition to half-open on next call
	err = cb.Call(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error in half-open state, got %v", err)
	}

	if cb.GetState() != CircuitBreakerHalfOpen {
		t.Errorf("Expected circuit breaker to be half-open, got %v", cb.GetState())
	}
}

// TestCircuitBreakerStateTransitions tests all state transitions
func TestCircuitBreakerStateTransitions(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold:   3,
		SuccessThreshold:   2,
		Timeout:            500 * time.Millisecond,
		MaxConcurrentCalls: 2,
	}

	cb := NewCircuitBreaker(config)

	// Test CLOSED to OPEN transition
	t.Run("ClosedToOpen_FailureThreshold", func(t *testing.T) {
		// Cause failures to reach threshold
		for i := 0; i < config.FailureThreshold; i++ {
			err := cb.Call(func() error {
				return errors.New("failure")
			})

			if err == nil {
				t.Errorf("Expected error for failure %d", i+1)
			}
		}

		// Should be OPEN now
		if cb.GetState() != CircuitBreakerOpen {
			t.Errorf("Expected OPEN state after %d failures, got: %v", config.FailureThreshold, cb.GetState())
		}
	})

	// Test OPEN to HALF_OPEN transition after timeout
	t.Run("OpenToHalfOpen_Timeout", func(t *testing.T) {
		// Wait for reset timeout
		time.Sleep(config.Timeout + 100*time.Millisecond)

		// First call should transition to HALF_OPEN
		err := cb.Call(func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error in half-open state, got: %v", err)
		}

		if cb.GetState() != CircuitBreakerHalfOpen {
			t.Errorf("Expected HALF_OPEN state, got: %v", cb.GetState())
		}
	})

	// Test HALF_OPEN to CLOSED transition
	t.Run("HalfOpenToClosed_SuccessThreshold", func(t *testing.T) {
		// Make enough successful calls to close the circuit
		for i := 0; i < config.SuccessThreshold-1; i++ { // -1 because we already made one successful call
			err := cb.Call(func() error {
				return nil
			})

			if err != nil {
				t.Errorf("Expected no error in half-open state, got: %v", err)
			}
		}

		// Should be CLOSED now
		if cb.GetState() != CircuitBreakerClosed {
			t.Errorf("Expected CLOSED state after %d successes, got: %v", config.SuccessThreshold, cb.GetState())
		}
	})

	// Test HALF_OPEN to OPEN transition on failure
	t.Run("HalfOpenToOpen_OnFailure", func(t *testing.T) {
		// Reset and trigger open state
		cb = NewCircuitBreaker(config)
		for i := 0; i < config.FailureThreshold; i++ {
			cb.Call(func() error {
				return errors.New("failure")
			})
		}

		// Wait for reset timeout to enter HALF_OPEN
		time.Sleep(config.Timeout + 100*time.Millisecond)

		// Make one successful call to enter HALF_OPEN
		cb.Call(func() error {
			return nil
		})

		// Now fail - should go back to OPEN
		err := cb.Call(func() error {
			return errors.New("failure in half-open")
		})

		if err == nil {
			t.Error("Expected error for failure in half-open state")
		}

		if cb.GetState() != CircuitBreakerOpen {
			t.Errorf("Expected OPEN state after failure in half-open, got: %v", cb.GetState())
		}
	})
}

// TestRetryLogicBasicFunctionality tests the retry logic functionality
func TestRetryLogicBasicFunctionality(t *testing.T) {
	policy := &RetryPolicy{
		MaxRetries:      3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        2 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          false, // Disable for predictable testing
		RetryableErrors: []string{ErrorCodeTimeout, ErrorCodeNetwork},
	}

	retrier := NewRetrier(policy)

	// Test successful execution without retries
	t.Run("SuccessfulExecution_NoRetries", func(t *testing.T) {
		attempts := 0
		result := retrier.Execute(context.Background(), func() error {
			attempts++
			return nil
		})

		if !result.Success {
			t.Errorf("Expected success, got failure: %v", result.LastError)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got: %d", attempts)
		}
	})

	// Test retry on retryable errors
	t.Run("RetryOnRetryableErrors", func(t *testing.T) {
		attempts := 0
		result := retrier.Execute(context.Background(), func() error {
			attempts++
			return errors.New("timeout: operation timed out") // Retryable error
		})

		if result.Success {
			t.Error("Expected failure after all retries exhausted")
		}
		expectedAttempts := policy.MaxRetries + 1 // Original attempt + retries
		if attempts != expectedAttempts {
			t.Errorf("Expected %d attempts, got: %d", expectedAttempts, attempts)
		}
	})

	// Test no retry on non-retryable errors
	t.Run("NoRetryOnNonRetryableErrors", func(t *testing.T) {
		attempts := 0
		result := retrier.Execute(context.Background(), func() error {
			attempts++
			return errors.New("auth failed: invalid credentials") // Non-retryable error
		})

		if result.Success {
			t.Error("Expected failure")
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt for non-retryable error, got: %d", attempts)
		}
	})

	// Test eventual success after retries
	t.Run("EventualSuccess_AfterRetries", func(t *testing.T) {
		attempts := 0
		result := retrier.Execute(context.Background(), func() error {
			attempts++
			if attempts < 3 {
				return errors.New("timeout: temporary failure")
			}
			return nil
		})

		if !result.Success {
			t.Errorf("Expected success after eventual success, got: %v", result.LastError)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got: %d", attempts)
		}
	})
}

// TestCircuitBreakerWithRetryIntegration tests circuit breaker and retry working together
func TestCircuitBreakerWithRetryIntegration(t *testing.T) {
	cbConfig := &CircuitBreakerConfig{
		FailureThreshold:   3,
		SuccessThreshold:   2,
		Timeout:            500 * time.Millisecond,
		MaxConcurrentCalls: 2,
	}

	retryPolicy := &RetryPolicy{
		MaxRetries:      2,
		InitialDelay:    50 * time.Millisecond,
		MaxDelay:        1 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          false,
		RetryableErrors: []string{ErrorCodeNetwork, ErrorCodeTimeout},
	}

	retrier := NewRetrier(retryPolicy)

	// Test circuit breaker preventing retries when open
	t.Run("CircuitBreakerPreventsRetries", func(t *testing.T) {
		cb := NewCircuitBreaker(cbConfig)

		// Force circuit breaker to OPEN state
		for i := 0; i < cbConfig.FailureThreshold; i++ {
			cb.Call(func() error {
				return errors.New("failure")
			})
		}

		// Now try with retry + circuit breaker - should fail immediately
		attempts := 0
		start := time.Now()

		result := retrier.ExecuteWithCircuitBreaker(context.Background(), cb, func() error {
			attempts++
			return errors.New("network error")
		})

		duration := time.Since(start)

		if result.Success {
			t.Error("Expected failure when circuit breaker is open")
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt when circuit breaker is open, got: %d", attempts)
		}
		if duration > 100*time.Millisecond {
			t.Errorf("Expected immediate failure with open circuit breaker, took: %v", duration)
		}
	})

	// Test successful recovery with circuit breaker and retries
	t.Run("SuccessfulRecovery_WithRetries", func(t *testing.T) {
		cb := NewCircuitBreaker(cbConfig)

		// Simulate service recovery scenario
		serviceDown := true
		attempts := 0

		// First, make the circuit breaker go to OPEN state
		for i := 0; i < cbConfig.FailureThreshold; i++ {
			cb.Call(func() error {
				return errors.New("service down")
			})
		}

		// Wait for reset timeout
		time.Sleep(cbConfig.Timeout + 100*time.Millisecond)

		// Now service is "recovering" - should succeed with retries
		result := retrier.ExecuteWithCircuitBreaker(context.Background(), cb, func() error {
			attempts++
			if serviceDown && attempts < 2 {
				return errors.New("timeout: service still recovering")
			}
			serviceDown = false
			return nil
		})

		if !result.Success {
			t.Errorf("Expected success after service recovery, got: %v", result.LastError)
		}
		if cb.GetState() != CircuitBreakerHalfOpen && cb.GetState() != CircuitBreakerClosed {
			t.Errorf("Expected circuit breaker to be in HALF_OPEN or CLOSED state, got: %v", cb.GetState())
		}
	})
}

// TestConcurrentCircuitBreakerAccess tests circuit breaker under concurrent access
func TestConcurrentCircuitBreakerAccess(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold:   5,
		SuccessThreshold:   3,
		Timeout:            100 * time.Millisecond,
		MaxConcurrentCalls: 3,
	}

	cb := NewCircuitBreaker(config)

	// Test concurrent successful calls
	t.Run("ConcurrentSuccessfulCalls", func(t *testing.T) {
		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cb.Call(func() error {
					return nil
				})
				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		if successCount != 10 {
			t.Errorf("Expected 10 successful calls, got: %d", successCount)
		}
		if cb.GetState() != CircuitBreakerClosed {
			t.Errorf("Expected CLOSED state, got: %v", cb.GetState())
		}
	})

	// Test concurrent calls with failures
	t.Run("ConcurrentFailureCalls", func(t *testing.T) {
		cb = NewCircuitBreaker(config) // Reset
		var wg sync.WaitGroup
		errorCount := 0
		var mu sync.Mutex

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := cb.Call(func() error {
					return errors.New("failure")
				})
				if err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		if errorCount != 10 {
			t.Errorf("Expected 10 failed calls, got: %d", errorCount)
		}
		if cb.GetState() != CircuitBreakerOpen {
			t.Errorf("Expected OPEN state after failures, got: %v", cb.GetState())
		}
	})
}

// TestRetryWithJitter tests retry logic with jitter enabled
func TestRetryWithJitter(t *testing.T) {
	policy := &RetryPolicy{
		MaxRetries:      3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        2 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          true,
		RetryableErrors: []string{ErrorCodeTimeout},
	}

	retrier := NewRetrier(policy)

	// Test that jitter introduces variability in delays
	t.Run("JitterIntroducesVariability", func(t *testing.T) {
		var durations []time.Duration

		for i := 0; i < 5; i++ {
			start := time.Now()

			retrier.Execute(context.Background(), func() error {
				return errors.New("timeout: consistent failure")
			})

			durations = append(durations, time.Since(start))
		}

		// Check that not all durations are exactly the same (indicating jitter is working)
		firstDuration := durations[0]
		variabilityFound := false
		tolerance := 50 * time.Millisecond

		for _, duration := range durations[1:] {
			if duration < firstDuration-tolerance || duration > firstDuration+tolerance {
				variabilityFound = true
				break
			}
		}

		if !variabilityFound {
			t.Error("Expected variability in retry delays with jitter enabled, but all durations were similar")
		}
	})
}
