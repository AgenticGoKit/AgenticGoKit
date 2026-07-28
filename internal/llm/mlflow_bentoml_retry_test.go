package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Regression coverage for PR #157 review's nit: MLFlow/BentoML adapters had
// their own internal retry loop that retried unconditionally, including a
// 4xx (bad request/auth failure) that retrying can never fix. Both loops now
// stop after the first non-retryable error, same classification as
// DefaultIsRetryable/CircuitBreakerProvider.

func TestMLFlowGatewayAdapter_Call_DoesNotRetryNonRetryableStatus(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"bad request","type":"invalid_request_error"}}`))
	}))
	defer server.Close()

	adapter, err := NewMLFlowGatewayAdapter(MLFlowGatewayConfig{
		BaseURL:    server.URL,
		ChatRoute:  "test-route",
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error constructing adapter: %v", err)
	}

	_, callErr := adapter.Call(context.Background(), Prompt{User: "hello"})
	if callErr == nil {
		t.Fatal("expected an error from a 400 response, got nil")
	}
	if requests != 1 {
		t.Fatalf("got %d requests, want 1 (a 400 must not be retried)", requests)
	}
}

func TestMLFlowGatewayAdapter_Call_RetriesRetryableStatus(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":{"message":"unavailable"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	adapter, err := NewMLFlowGatewayAdapter(MLFlowGatewayConfig{
		BaseURL:    server.URL,
		ChatRoute:  "test-route",
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error constructing adapter: %v", err)
	}

	resp, callErr := adapter.Call(context.Background(), Prompt{User: "hello"})
	if callErr != nil {
		t.Fatalf("expected eventual success after retries, got error: %v", callErr)
	}
	if resp.Content != "ok" {
		t.Fatalf("got content %q, want %q", resp.Content, "ok")
	}
	if requests != 3 {
		t.Fatalf("got %d requests, want 3 (2 retryable failures + 1 success)", requests)
	}
}

func TestBentoMLAdapter_Call_DoesNotRetryNonRetryableStatus(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"invalid api key","type":"invalid_request_error"}}`))
	}))
	defer server.Close()

	adapter, err := NewBentoMLAdapter(BentoMLConfig{
		BaseURL:    server.URL,
		Model:      "test-model",
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error constructing adapter: %v", err)
	}

	_, callErr := adapter.Call(context.Background(), Prompt{User: "hello"})
	if callErr == nil {
		t.Fatal("expected an error from a 401 response, got nil")
	}
	if requests != 1 {
		t.Fatalf("got %d requests, want 1 (a 401 must not be retried)", requests)
	}
}

func TestBentoMLAdapter_Call_RetriesRetryableStatus(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"message":"rate limited"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	adapter, err := NewBentoMLAdapter(BentoMLConfig{
		BaseURL:    server.URL,
		Model:      "test-model",
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error constructing adapter: %v", err)
	}

	resp, callErr := adapter.Call(context.Background(), Prompt{User: "hello"})
	if callErr != nil {
		t.Fatalf("expected eventual success after retries, got error: %v", callErr)
	}
	if resp.Content != "ok" {
		t.Fatalf("got content %q, want %q", resp.Content, "ok")
	}
	if requests != 2 {
		t.Fatalf("got %d requests, want 2 (1 retryable failure + 1 success)", requests)
	}
}
