package llm

import (
	"context"
	"errors"
	"net"
	"time"
)

// CircuitBreaker is the minimal circuit-breaking seam CircuitBreakerProvider
// needs. Satisfied structurally by
// internal/core/error_handling.CircuitBreakerImplementation without this
// package importing it: internal/core/error_handling imports core, and core
// imports internal/llm (core/llm.go), so internal/llm -> internal/core/error_handling
// would close an import cycle. Callers that CAN safely import both (e.g.
// v1beta/provider_factory.go, which is downstream of both) construct the
// concrete breaker and pass it in through ProviderConfig.CircuitBreaker.
type CircuitBreaker interface {
	Call(fn func() error) error
}

// RetryPolicy configures the retry loop CircuitBreakerProvider wraps around
// a ModelProvider call.
type RetryPolicy struct {
	// MaxRetries is the number of retry attempts after the first try. Zero
	// means no retries (the call still runs once).
	MaxRetries int

	// IsRetryable classifies whether err should trigger a retry. Defaults to
	// DefaultIsRetryable when nil. Deliberately NOT the exact-string-match
	// classification core.RetryPolicy.RetryableErrors/
	// error_handling.RetrierImplementation.isRetryableError use — that
	// approach breaks silently when an error's message wording changes.
	IsRetryable func(err error) bool

	// BackoffFunc computes the delay before attempt N (1-based). Defaults to
	// exponential backoff with a 10s cap when nil.
	BackoffFunc func(attempt int) time.Duration
}

// DefaultIsRetryable classifies context deadline/cancellation, net.Error
// (including timeouts surfaced through wrapped chains, e.g. an http.Client
// body-read timeout), and an *APIStatusError with a 429 or 5xx status as
// retryable. errors.Is/As based, not string-matched, so it survives
// error-message wording changes upstream.
//
// A completed HTTP round-trip carrying a gateway error (502/503/524, a
// Cloudflare edge error like 530, or a 429 rate limit) is not a net.Error —
// it's a successful response with a non-2xx status, surfaced by this
// package's adapters as *APIStatusError (openai_adapter.go). Without this
// case, MaxRetries would only cover network/timeout failures and silently
// not retry the exact class of failure most likely to be transient: the
// origin/gateway itself being briefly unavailable. 4xx other than 429 (bad
// request, auth failure, not found) is not retryable — retrying won't fix a
// malformed request or bad credentials.
func DefaultIsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var apiErr *APIStatusError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 429 || apiErr.StatusCode >= 500
	}
	return false
}

func defaultBackoff(attempt int) time.Duration {
	const base = 100 * time.Millisecond
	const max = 10 * time.Second
	if attempt <= 0 {
		return base
	}
	d := base * time.Duration(uint(1)<<uint(attempt-1))
	if d > max || d <= 0 {
		d = max
	}
	return d
}

// CircuitBreakerProviderConfig configures CircuitBreakerProvider.
type CircuitBreakerProviderConfig struct {
	// Breaker, when non-nil, gates every call through Breaker.Call(fn) —
	// an open circuit short-circuits immediately without touching the inner
	// provider or sleeping through a retry loop. Nil disables
	// circuit-breaking (retry-only behavior).
	Breaker CircuitBreaker
	Retry   RetryPolicy
}

// CircuitBreakerProvider decorates any ModelProvider with circuit-breaking
// and retry, without the wrapped adapter knowing about either. It satisfies
// ModelProvider itself, so every existing adapter (OpenAI, Anthropic, Azure,
// Ollama, OpenRouter, HuggingFace, vLLM, FoundryLocal, BentoML, MLFlow) gets
// this by having its constructor's return value wrapped once, centrally, in
// ProviderFactory.CreateProvider — no per-adapter changes.
type CircuitBreakerProvider struct {
	inner  ModelProvider
	config CircuitBreakerProviderConfig
}

// NewCircuitBreakerProvider wraps inner with circuit-breaking/retry per
// config. Safe to call with a zero-value config.Retry (defaults apply) and a
// nil config.Breaker (retry-only, no circuit-breaking).
func NewCircuitBreakerProvider(inner ModelProvider, config CircuitBreakerProviderConfig) *CircuitBreakerProvider {
	if config.Retry.IsRetryable == nil {
		config.Retry.IsRetryable = DefaultIsRetryable
	}
	if config.Retry.BackoffFunc == nil {
		config.Retry.BackoffFunc = defaultBackoff
	}
	return &CircuitBreakerProvider{inner: inner, config: config}
}

// call runs fn through the circuit breaker (if configured), then retries
// per the retry policy, honoring ctx cancellation between attempts.
func (p *CircuitBreakerProvider) call(ctx context.Context, fn func() error) error {
	protected := fn
	if p.config.Breaker != nil {
		protected = func() error { return p.config.Breaker.Call(fn) }
	}

	var lastErr error
	for attempt := 0; attempt <= p.config.Retry.MaxRetries; attempt++ {
		lastErr = protected()
		if lastErr == nil {
			return nil
		}
		if !p.config.Retry.IsRetryable(lastErr) {
			return lastErr
		}
		if attempt < p.config.Retry.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(p.config.Retry.BackoffFunc(attempt + 1)):
			}
		}
	}
	return lastErr
}

// Call implements ModelProvider.
func (p *CircuitBreakerProvider) Call(ctx context.Context, prompt Prompt) (Response, error) {
	var resp Response
	err := p.call(ctx, func() error {
		var callErr error
		resp, callErr = p.inner.Call(ctx, prompt)
		return callErr
	})
	return resp, err
}

// Stream implements ModelProvider. Only the connection-setup call
// (inner.Stream returning before any token is delivered) is retried. Once a
// channel is returned, tokens flow through unmodified: retrying mid-stream
// would require buffering everything already sent to the caller and risks
// double-emitting content already delivered.
func (p *CircuitBreakerProvider) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	var out <-chan Token
	err := p.call(ctx, func() error {
		var streamErr error
		out, streamErr = p.inner.Stream(ctx, prompt)
		return streamErr
	})
	return out, err
}

// Embeddings implements ModelProvider.
func (p *CircuitBreakerProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	var embeddings [][]float64
	err := p.call(ctx, func() error {
		var callErr error
		embeddings, callErr = p.inner.Embeddings(ctx, texts)
		return callErr
	})
	return embeddings, err
}
