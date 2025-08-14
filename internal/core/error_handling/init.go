// Package error_handling provides factory registration for error handling components.
package error_handling

import (
	"context"

	"github.com/kunalkushwaha/agenticgokit/core"
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

	// Register retrier factory
	core.RegisterRetrierFactory(func(policy *core.RetryPolicy) core.Retrier {
		impl := NewRetrierImplementation(policy)
		return &retrierAdapter{impl: impl}
	})

	// Register retry manager factory
	core.RegisterRetryManagerFactory(func() core.RetryManager {
		impl := NewRetryManagerImplementation()
		return &retryManagerAdapter{impl: impl}
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
type retrierAdapter struct {
	impl *RetrierImplementation
}

func (ra *retrierAdapter) Execute(ctx context.Context, fn core.RetryFunc) *core.RetryResult {
	return ra.impl.Execute(ctx, fn)
}

func (ra *retrierAdapter) SetCallbacks(callbacks core.RetryCallbacks) {
	ra.impl.SetCallbacks(callbacks)
}

// retryManagerAdapter adapts the internal implementation to the core interface
type retryManagerAdapter struct {
	impl *RetryManagerImplementation
}

func (rma *retryManagerAdapter) AddRetrier(name string, retrier core.Retrier) {
	// This would need proper adaptation between core.Retrier and internal implementation
	// For now, this is a placeholder
}

func (rma *retryManagerAdapter) GetRetrier(name string) (core.Retrier, bool) {
	// This would need proper adaptation
	return nil, false
}

func (rma *retryManagerAdapter) GetMetrics(name string) (*core.RetryMetrics, bool) {
	return rma.impl.GetMetrics(name)
}
