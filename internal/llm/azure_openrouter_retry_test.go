package llm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Regression coverage for PR #157 review's "Medium" item: Azure and
// OpenRouter didn't actually benefit from the new MaxRetries/circuit-breaker
// support because DefaultIsRetryable only recognizes *APIStatusError
// (produced by the OpenAI-family adapters), while azure_adapter.go and
// openrouter_adapter.go returned plain fmt.Errorf on non-2xx — silently
// classifying a transient 429/503 from either provider as non-retryable.

func TestAzureAdapter_Call_NonOKStatus_ReturnsAPIStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":{"message":"backend overloaded","type":"server_error","code":"503"}}`))
	}))
	defer server.Close()

	adapter, err := NewAzureOpenAIAdapter(AzureOpenAIAdapterOptions{
		Endpoint:            server.URL,
		APIKey:              "test-key",
		ChatDeployment:      "test-chat",
		EmbeddingDeployment: "test-embed",
	})
	if err != nil {
		t.Fatalf("unexpected error constructing adapter: %v", err)
	}

	_, callErr := adapter.Call(context.Background(), Prompt{User: "hello"})
	if callErr == nil {
		t.Fatal("expected an error from a 503 response, got nil")
	}

	var apiErr *APIStatusError
	if !errors.As(callErr, &apiErr) {
		t.Fatalf("expected *APIStatusError (so DefaultIsRetryable can classify it), got %T: %v", callErr, callErr)
	}
	if apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("got status %d, want %d", apiErr.StatusCode, http.StatusServiceUnavailable)
	}
	if !DefaultIsRetryable(callErr) {
		t.Fatal("a 503 from Azure must be classified retryable by DefaultIsRetryable")
	}
}

func TestAzureAdapter_Call_AuthFailure_NotRetryable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"invalid api key","type":"invalid_request_error","code":"401"}}`))
	}))
	defer server.Close()

	adapter, err := NewAzureOpenAIAdapter(AzureOpenAIAdapterOptions{
		Endpoint:            server.URL,
		APIKey:              "test-key",
		ChatDeployment:      "test-chat",
		EmbeddingDeployment: "test-embed",
	})
	if err != nil {
		t.Fatalf("unexpected error constructing adapter: %v", err)
	}

	_, callErr := adapter.Call(context.Background(), Prompt{User: "hello"})
	if callErr == nil {
		t.Fatal("expected an error from a 401 response, got nil")
	}
	if DefaultIsRetryable(callErr) {
		t.Fatal("a 401 (bad credentials) must NOT be classified retryable")
	}
}

func TestOpenRouterAdapter_Call_NonOKStatus_ReturnsAPIStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"rate limited","code":"429","type":"rate_limit"}}`))
	}))
	defer server.Close()

	adapter, err := NewOpenRouterAdapter("test-key", "test-model", server.URL, 0, 0, "", "")
	if err != nil {
		t.Fatalf("unexpected error constructing adapter: %v", err)
	}

	_, callErr := adapter.Call(context.Background(), Prompt{User: "hello"})
	if callErr == nil {
		t.Fatal("expected an error from a 429 response, got nil")
	}

	var apiErr *APIStatusError
	if !errors.As(callErr, &apiErr) {
		t.Fatalf("expected *APIStatusError (so DefaultIsRetryable can classify it), got %T: %v", callErr, callErr)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("got status %d, want %d", apiErr.StatusCode, http.StatusTooManyRequests)
	}
	if !DefaultIsRetryable(callErr) {
		t.Fatal("a 429 from OpenRouter must be classified retryable by DefaultIsRetryable")
	}
}

// TestOpenRouterAdapter_Stream_NonOKStatus_BodyNotLostToEarlyClose is the
// regression test for the actual bug: Stream previously called
// resp.Body.Close() BEFORE io.ReadAll(resp.Body), so a non-200 stream error
// always carried an empty body — this asserts the real error body text
// (not just a non-nil error) survives into the returned error.
func TestOpenRouterAdapter_Stream_NonOKStatus_BodyNotLostToEarlyClose(t *testing.T) {
	const wantMessage = "context length exceeded for this model"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"` + wantMessage + `","code":"400","type":"invalid_request_error"}}`))
	}))
	defer server.Close()

	adapter, err := NewOpenRouterAdapter("test-key", "test-model", server.URL, 0, 0, "", "")
	if err != nil {
		t.Fatalf("unexpected error constructing adapter: %v", err)
	}

	_, streamErr := adapter.Stream(context.Background(), Prompt{User: "hello"})
	if streamErr == nil {
		t.Fatal("expected an error from a 400 response, got nil")
	}

	var apiErr *APIStatusError
	if !errors.As(streamErr, &apiErr) {
		t.Fatalf("expected *APIStatusError, got %T: %v", streamErr, streamErr)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
	if apiErr.Body == "" {
		t.Fatal("APIStatusError.Body is empty — the body-read-after-close bug regressed")
	}
	if !contains(apiErr.Body, wantMessage) {
		t.Fatalf("APIStatusError.Body = %q, want it to contain %q", apiErr.Body, wantMessage)
	}
	if DefaultIsRetryable(streamErr) {
		t.Fatal("a 400 (bad request) must NOT be classified retryable")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}
