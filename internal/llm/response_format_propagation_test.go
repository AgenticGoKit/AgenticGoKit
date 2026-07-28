package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests assert that ResponseFormat (and, for the OpenAI-compatible
// composed adapters, CachePrompt) actually reaches the outbound request
// body for every adapter that claims to support it — not just OpenAI's own
// adapter. vLLM/MLFlow/BentoML compose OpenAIAdapterConfig internally
// (verified by inspection: "// Embed OpenAI adapter for code reuse"), so a
// missing field there would silently no-op rather than error.

func TestVLLMAdapter_Call_ResponseFormat(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &gotBody))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	adapter, err := NewVLLMAdapter(VLLMConfig{
		BaseURL:        server.URL,
		Model:          "test-model",
		ResponseFormat: map[string]interface{}{"type": "json_object"},
	})
	require.NoError(t, err)

	_, err = adapter.Call(context.Background(), Prompt{User: "hello"})
	require.NoError(t, err)

	rf, ok := gotBody["response_format"].(map[string]interface{})
	require.True(t, ok, "response_format missing from vLLM request body: %v", gotBody)
	assert.Equal(t, "json_object", rf["type"])
}

func TestMLFlowGatewayAdapter_Call_ResponseFormat(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &gotBody))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	adapter, err := NewMLFlowGatewayAdapter(MLFlowGatewayConfig{
		BaseURL:        server.URL,
		ChatRoute:      "test-route",
		ResponseFormat: map[string]interface{}{"type": "json_object"},
	})
	require.NoError(t, err)

	_, err = adapter.Call(context.Background(), Prompt{User: "hello"})
	require.NoError(t, err)

	rf, ok := gotBody["response_format"].(map[string]interface{})
	require.True(t, ok, "response_format missing from MLFlow Gateway request body: %v", gotBody)
	assert.Equal(t, "json_object", rf["type"])
}

func TestBentoMLAdapter_Call_ResponseFormat(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &gotBody))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	adapter, err := NewBentoMLAdapter(BentoMLConfig{
		BaseURL:        server.URL,
		Model:          "test-model",
		ResponseFormat: map[string]interface{}{"type": "json_object"},
	})
	require.NoError(t, err)

	_, err = adapter.Call(context.Background(), Prompt{User: "hello"})
	require.NoError(t, err)

	rf, ok := gotBody["response_format"].(map[string]interface{})
	require.True(t, ok, "response_format missing from BentoML request body: %v", gotBody)
	assert.Equal(t, "json_object", rf["type"])
}

func TestOpenRouterAdapter_Call_ResponseFormat(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &gotBody))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	// factory.go sets these fields directly (same package) after
	// construction — replicate that here since NewOpenRouterAdapter's
	// constructor intentionally keeps its existing positional signature.
	adapter, err := NewOpenRouterAdapter("test-key", "test-model", server.URL, 100, 0.5, "", "")
	require.NoError(t, err)
	adapter.responseFormat = map[string]interface{}{"type": "json_object"}

	_, err = adapter.Call(context.Background(), Prompt{User: "hello"})
	require.NoError(t, err)

	rf, ok := gotBody["response_format"].(map[string]interface{})
	require.True(t, ok, "response_format missing from OpenRouter request body: %v", gotBody)
	assert.Equal(t, "json_object", rf["type"])
}

func TestAzureOpenAIAdapter_Call_ResponseFormat(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &gotBody))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}],"usage":{}}`))
	}))
	defer server.Close()

	adapter, err := NewAzureOpenAIAdapter(AzureOpenAIAdapterOptions{
		Endpoint:            server.URL,
		APIKey:              "test-key",
		ChatDeployment:      "test-chat",
		EmbeddingDeployment: "test-embed",
		ResponseFormat:      map[string]interface{}{"type": "json_object"},
	})
	require.NoError(t, err)

	_, err = adapter.Call(context.Background(), Prompt{User: "hello"})
	require.NoError(t, err)

	rf, ok := gotBody["response_format"].(map[string]interface{})
	require.True(t, ok, "response_format missing from Azure OpenAI request body: %v", gotBody)
	assert.Equal(t, "json_object", rf["type"])
}
