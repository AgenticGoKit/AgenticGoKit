// Package error_handling provides factory functions for error handling components.
package error_handling

import (
	"github.com/agenticgokit/agenticgokit/core"
)

// ErrorHandlingFactory provides factory functions for error handling components
type ErrorHandlingFactory struct{}

// NewErrorHandlingFactory creates a new error handling factory
func NewErrorHandlingFactory() *ErrorHandlingFactory {
	return &ErrorHandlingFactory{}
}

// CreateErrorRoutingImplementation creates a new error routing implementation
func (f *ErrorHandlingFactory) CreateErrorRoutingImplementation(config *core.ErrorRouterConfig) *ErrorRoutingImplementation {
	return NewErrorRoutingImplementation(config)
}

// CreateCircuitBreakerImplementation creates a new circuit breaker implementation
func (f *ErrorHandlingFactory) CreateCircuitBreakerImplementation(config *core.CircuitBreakerConfig) *CircuitBreakerImplementation {
	return NewCircuitBreakerImplementation(config)
}

// CreateRetrierImplementation creates a new retrier implementation
func (f *ErrorHandlingFactory) CreateRetrierImplementation(policy *core.RetryPolicy) *RetrierImplementation {
	return NewRetrierImplementation(policy)
}

// CreateRetryManagerImplementation creates a new retry manager implementation
func (f *ErrorHandlingFactory) CreateRetryManagerImplementation() *RetryManagerImplementation {
	return NewRetryManagerImplementation()
}

// Default factory instance
var defaultFactory = NewErrorHandlingFactory()

// GetDefaultFactory returns the default error handling factory
func GetDefaultFactory() *ErrorHandlingFactory {
	return defaultFactory
}

