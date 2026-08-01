package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestOllamaEmbeddingModelNotFoundHint verifies the error for a missing model
// tells the user exactly how to fix it.
func TestOllamaEmbeddingModelNotFoundHint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"model \"nomic-embed-text\" not found, try pulling it first"}`))
	}))
	defer server.Close()

	svc := NewOllamaEmbeddingService("nomic-embed-text", server.URL)
	_, err := svc.GenerateEmbedding(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for missing model, got nil")
	}
	if !strings.Contains(err.Error(), "ollama pull nomic-embed-text") {
		t.Errorf("error should include the `ollama pull` fix, got: %v", err)
	}
}

// TestOllamaEmbeddingConnectionRefusedHint verifies the error for an
// unreachable Ollama instance points at starting/aiming Ollama.
func TestOllamaEmbeddingConnectionRefusedHint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // immediately close so the port refuses connections

	svc := NewOllamaEmbeddingService("nomic-embed-text", server.URL)
	_, err := svc.GenerateEmbedding(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for unreachable Ollama, got nil")
	}
	if !strings.Contains(err.Error(), "ollama serve") {
		t.Errorf("error should suggest starting Ollama, got: %v", err)
	}
	if !strings.Contains(err.Error(), server.URL) {
		t.Errorf("error should include the attempted base URL, got: %v", err)
	}
}
