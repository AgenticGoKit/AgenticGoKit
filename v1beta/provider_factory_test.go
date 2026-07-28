package v1beta

import (
	"testing"
	"time"

	"github.com/agenticgokit/agenticgokit/internal/llm"
)

func TestCreateLLMProvider_NoWrapWhenRetryAndCircuitBreakerUnset(t *testing.T) {
	p, err := createLLMProvider(LLMConfig{Provider: "mock"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, wrapped := p.(*llm.CircuitBreakerProvider); wrapped {
		t.Fatal("provider should not be wrapped when MaxRetries/CircuitBreaker are unset (default LLMConfig)")
	}
}

func TestCreateLLMProvider_WrapsWhenMaxRetriesSet(t *testing.T) {
	p, err := createLLMProvider(LLMConfig{Provider: "mock", MaxRetries: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, wrapped := p.(*llm.CircuitBreakerProvider); !wrapped {
		t.Fatal("provider should be wrapped when MaxRetries is set")
	}
}

func TestCreateLLMProvider_WrapsWhenCircuitBreakerEnabled(t *testing.T) {
	p, err := createLLMProvider(LLMConfig{
		Provider: "mock",
		CircuitBreaker: &CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 5,
			SuccessThreshold: 2,
			Timeout:          30 * time.Second,
			HalfOpenMaxCalls: 2,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, wrapped := p.(*llm.CircuitBreakerProvider); !wrapped {
		t.Fatal("provider should be wrapped when CircuitBreaker.Enabled is true")
	}
}

func TestCreateLLMProvider_NoWrapWhenCircuitBreakerConfigDisabled(t *testing.T) {
	// A non-nil CircuitBreaker config with Enabled: false must behave like
	// no config at all — the struct being present isn't the trigger.
	p, err := createLLMProvider(LLMConfig{
		Provider:       "mock",
		CircuitBreaker: &CircuitBreakerConfig{Enabled: false},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, wrapped := p.(*llm.CircuitBreakerProvider); wrapped {
		t.Fatal("provider should not be wrapped when CircuitBreaker.Enabled is false")
	}
}
