package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAIAdapter_Call_ResponseFormat is hermetic (httptest, no live API
// key needed) — asserts the "response_format" field's presence/absence in
// the actual outbound request body, unlike TestOpenAIAdapter_Call above
// which only exercises success/failure against a real endpoint.
func TestOpenAIAdapter_Call_ResponseFormat(t *testing.T) {
	t.Run("set when configured", func(t *testing.T) {
		var gotBody map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			require.NoError(t, json.Unmarshal(body, &gotBody))
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"choices":[{"message":{"content":"{}"},"finish_reason":"stop"}]}`))
		}))
		defer server.Close()

		adapter, err := NewOpenAIAdapterWithConfig(OpenAIAdapterConfig{
			APIKey:         "test-key",
			Model:          "test-model",
			BaseURL:        server.URL,
			ResponseFormat: map[string]interface{}{"type": "json_object"},
		})
		require.NoError(t, err)

		_, err = adapter.Call(context.Background(), Prompt{User: "hello"})
		require.NoError(t, err)

		rf, ok := gotBody["response_format"].(map[string]interface{})
		require.True(t, ok, "response_format missing from request body: %v", gotBody)
		assert.Equal(t, "json_object", rf["type"])
	})

	t.Run("absent when not configured", func(t *testing.T) {
		var gotBody map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			require.NoError(t, json.Unmarshal(body, &gotBody))
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
		}))
		defer server.Close()

		adapter, err := NewOpenAIAdapterWithConfig(OpenAIAdapterConfig{
			APIKey:  "test-key",
			Model:   "test-model",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		_, err = adapter.Call(context.Background(), Prompt{User: "hello"})
		require.NoError(t, err)

		_, ok := gotBody["response_format"]
		assert.False(t, ok, "response_format should be absent when not configured, got: %v", gotBody)
	})
}

// TestOpenAIAdapter_Call_CachePrompt is hermetic (httptest, no live API key
// needed) — asserts the "cache_prompt" field's presence/absence in the
// actual outbound request body.
func TestOpenAIAdapter_Call_CachePrompt(t *testing.T) {
	t.Run("set when configured", func(t *testing.T) {
		var gotBody map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			require.NoError(t, json.Unmarshal(body, &gotBody))
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
		}))
		defer server.Close()

		adapter, err := NewOpenAIAdapterWithConfig(OpenAIAdapterConfig{
			APIKey:      "test-key",
			Model:       "test-model",
			BaseURL:     server.URL,
			CachePrompt: true,
		})
		require.NoError(t, err)

		_, err = adapter.Call(context.Background(), Prompt{User: "hello"})
		require.NoError(t, err)

		cp, ok := gotBody["cache_prompt"].(bool)
		require.True(t, ok, "cache_prompt missing from request body: %v", gotBody)
		assert.True(t, cp)
	})

	t.Run("absent when not configured", func(t *testing.T) {
		var gotBody map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			require.NoError(t, json.Unmarshal(body, &gotBody))
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
		}))
		defer server.Close()

		adapter, err := NewOpenAIAdapterWithConfig(OpenAIAdapterConfig{
			APIKey:  "test-key",
			Model:   "test-model",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		_, err = adapter.Call(context.Background(), Prompt{User: "hello"})
		require.NoError(t, err)

		_, ok := gotBody["cache_prompt"]
		assert.False(t, ok, "cache_prompt should be absent when not configured, got: %v", gotBody)
	})
}

func TestOpenAIAdapter_Call(t *testing.T) {
	t.Run("Valid prompt", func(t *testing.T) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY environment variable is not set")
		}

		adapter, err := NewOpenAIAdapter(apiKey, "gpt-4o-mini", 50, 0.7)
		require.NoError(t, err)

		ctx := context.Background()
		prompt := Prompt{
			System: "Test system",
			User:   "User prompt",
			Parameters: ModelParameters{
				Temperature: floatPtr(0.7),
				MaxTokens:   int32Ptr(50),
			},
		}
		response, err := adapter.Call(ctx, prompt)

		// Assertions
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Content)
	})

	t.Run("Empty prompt", func(t *testing.T) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY environment variable is not set")
		}

		adapter, err := NewOpenAIAdapter(apiKey, "gpt-4o-mini", 50, 0.7)
		require.NoError(t, err)

		ctx := context.Background()
		prompt := Prompt{System: "", User: "", Parameters: ModelParameters{}}
		response, err := adapter.Call(ctx, prompt)

		// Assertions
		assert.Error(t, err)
		assert.Empty(t, response.Content)
	})
}

func TestOpenAIAdapter_Embeddings(t *testing.T) {
	t.Run("Valid input", func(t *testing.T) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY environment variable is not set")
		}

		adapter, err := NewOpenAIAdapter(apiKey, "text-embedding-ada-002", 0, 0)
		require.NoError(t, err)

		ctx := context.Background()
		inputs := []string{"Test input"}
		embeddings, err := adapter.Embeddings(ctx, inputs)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, embeddings)
		assert.Greater(t, len(embeddings), 0)
	})

	t.Run("Empty input", func(t *testing.T) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY environment variable is not set")
		}

		adapter, err := NewOpenAIAdapter(apiKey, "text-embedding-ada-002", 0, 0)
		require.NoError(t, err)

		ctx := context.Background()
		inputs := []string{}
		embeddings, err := adapter.Embeddings(ctx, inputs)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, embeddings)
		assert.Equal(t, 0, len(embeddings))
	})
}

