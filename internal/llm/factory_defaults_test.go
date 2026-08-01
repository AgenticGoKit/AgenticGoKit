package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCreateProviderDefaultMaxTokens verifies the factory no longer applies
// the surprising 150-token default (issue #143).
func TestCreateProviderDefaultMaxTokens(t *testing.T) {
	f := NewProviderFactory()
	provider, err := f.CreateProvider(ProviderConfig{Type: ProviderTypeOllama})
	if err != nil {
		t.Fatalf("CreateProvider failed: %v", err)
	}

	adapter, ok := provider.(*OllamaAdapter)
	if !ok {
		t.Fatalf("unexpected provider type %T", provider)
	}
	if adapter.maxTokens != DefaultMaxTokens {
		t.Errorf("maxTokens = %d, want DefaultMaxTokens (%d)", adapter.maxTokens, DefaultMaxTokens)
	}
	if DefaultMaxTokens < 1024 {
		t.Errorf("DefaultMaxTokens = %d; agent workloads need a generous default", DefaultMaxTokens)
	}
	if adapter.temperature != DefaultTemperature {
		t.Errorf("temperature = %v, want DefaultTemperature (%v)", adapter.temperature, DefaultTemperature)
	}
}

// TestOllamaCallHonorsExplicitZeroTemperature verifies that an explicit
// per-call temperature of 0 (deterministic sampling) reaches the request
// payload instead of being replaced by the adapter default (issue #143).
func TestOllamaCallHonorsExplicitZeroTemperature(t *testing.T) {
	var gotPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Errorf("failed to decode request payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message": {"role": "assistant", "content": "ok"}, "done": true}`))
	}))
	defer server.Close()

	adapter, err := NewOllamaAdapter(server.URL, "test-model", 128, 0.9, 0)
	if err != nil {
		t.Fatalf("NewOllamaAdapter failed: %v", err)
	}

	zero := float32(0)
	_, err = adapter.Call(context.Background(), Prompt{
		User:       "hello",
		Parameters: ModelParameters{Temperature: &zero},
	})
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	temp, ok := gotPayload["temperature"].(float64)
	if !ok {
		t.Fatalf("temperature missing from payload: %v", gotPayload)
	}
	if temp != 0 {
		t.Errorf("payload temperature = %v, want explicit 0 (not adapter default 0.9)", temp)
	}
}
