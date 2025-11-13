// Package error_handling provides factory registration for error handling components.
package error_handling

import (
	"context"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

// init automatically registers the error handling factories when the package is imported
func init() {
	// Register error router factory
	core.RegisterErrorRouterFactory(func(config *core.ErrorRouterConfig) core.ErrorRouter {
		return NewErrorRoutingImplementation(config)
	})

	// Register circuit breaker factory
	core.RegisterCircuitBreakerFactory(func(config *core.CircuitBreakerConfig) core.CircuitBreaker {
		impl := NewCircuitBreakerImplementation(config)
		return &circuitBreakerAdapter{impl: impl}
	})

	// Register retry handler factory (align with core.RetryHandler)
	core.RegisterRetryHandlerFactory(func(policy *core.RetryPolicy) core.RetryHandler {
		impl := NewRetrierImplementation(policy)
		return &retryHandlerAdapter{impl: impl}
	})
}

// =============================================================================
// ADAPTER TYPES FOR INTERFACE COMPATIBILITY
// =============================================================================

// circuitBreakerAdapter adapts the internal implementation to the core interface
type circuitBreakerAdapter struct {
	impl *CircuitBreakerImplementation
}

func (cba *circuitBreakerAdapter) Call(fn func() error) error {
	return cba.impl.Call(fn)
}

func (cba *circuitBreakerAdapter) SetStateChangeCallback(callback func(from, to core.CircuitBreakerState)) {
	cba.impl.SetStateChangeCallback(callback)
}

func (cba *circuitBreakerAdapter) GetState() core.CircuitBreakerState {
	return cba.impl.GetState()
}

func (cba *circuitBreakerAdapter) GetMetrics() core.CircuitBreakerMetrics {
	return cba.impl.GetMetrics()
}

// retrierAdapter adapts the internal implementation to the core interface
type retryHandlerAdapter struct {
	impl *RetrierImplementation
}

func (ra *retryHandlerAdapter) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	// Bridge to internal Execute which returns detailed result; here we mimic core interface
	res := ra.impl.Execute(ctx, func() error { return operation() })
	if res != nil && !res.Success {
		if res.LastError != nil {
			return res.LastError
		}
		return context.Canceled // generic
	}
	return nil
}

func (ra *retryHandlerAdapter) ShouldRetry(attempt int, err error) bool {
	// Use policy from impl
	if ra.impl == nil || ra.impl.policy == nil {
		return false
	}
	if attempt >= ra.impl.policy.MaxRetries {
		return false
	}
	return ra.impl.isRetryableError(err)
}

func (ra *retryHandlerAdapter) CalculateDelay(attempt int) time.Duration {
	if ra.impl == nil {
		return 0
	}
	return ra.impl.calculateDelay(attempt)
}

// retryManagerAdapter adapts the internal implementation to the core interface
// Remove retry manager adapter for now; core exposes RetryHandler only

