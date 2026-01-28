package llm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicAdapter_Call(t *testing.T) {
	t.Run("Valid prompt", func(t *testing.T) {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			t.Skip("ANTHROPIC_API_KEY environment variable is not set")
		}

		adapter, err := NewAnthropicAdapter(apiKey, "claude-sonnet-4-20250514", 100, 0.7)
		require.NoError(t, err)

		ctx := context.Background()
		prompt := Prompt{
			System: "You are a helpful assistant. Respond concisely.",
			User:   "Say hello in exactly three words.",
			Parameters: ModelParameters{
				Temperature: floatPtr(0.7),
				MaxTokens:   int32Ptr(50),
			},
		}
		response, err := adapter.Call(ctx, prompt)

		// Assertions
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Content)
		assert.Greater(t, response.Usage.TotalTokens, 0)
	})

	t.Run("Empty prompt", func(t *testing.T) {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			t.Skip("ANTHROPIC_API_KEY environment variable is not set")
		}

		adapter, err := NewAnthropicAdapter(apiKey, "claude-sonnet-4-20250514", 100, 0.7)
		require.NoError(t, err)

		ctx := context.Background()
		prompt := Prompt{System: "", User: "", Parameters: ModelParameters{}}
		response, err := adapter.Call(ctx, prompt)

		// Assertions
		assert.Error(t, err)
		assert.Empty(t, response.Content)
	})

	t.Run("Empty API key", func(t *testing.T) {
		_, err := NewAnthropicAdapter("", "claude-sonnet-4-20250514", 100, 0.7)
		assert.Error(t, err)
	})
}

func TestAnthropicAdapter_Stream(t *testing.T) {
	t.Run("Valid streaming", func(t *testing.T) {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			t.Skip("ANTHROPIC_API_KEY environment variable is not set")
		}

		adapter, err := NewAnthropicAdapter(apiKey, "claude-sonnet-4-20250514", 100, 0.7)
		require.NoError(t, err)

		ctx := context.Background()
		prompt := Prompt{
			System: "You are a helpful assistant.",
			User:   "Count from 1 to 5.",
			Parameters: ModelParameters{
				MaxTokens: int32Ptr(100),
			},
		}

		tokenChan, err := adapter.Stream(ctx, prompt)
		require.NoError(t, err)

		var content string
		for token := range tokenChan {
			if token.Error != nil {
				t.Fatalf("Stream error: %v", token.Error)
			}
			content += token.Content
		}

		assert.NotEmpty(t, content)
	})
}

func TestAnthropicAdapter_Embeddings(t *testing.T) {
	t.Run("Returns error", func(t *testing.T) {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			apiKey = "test-key" // Use test key since we expect an error anyway
		}

		adapter, err := NewAnthropicAdapter(apiKey, "claude-sonnet-4-20250514", 100, 0.7)
		require.NoError(t, err)

		ctx := context.Background()
		_, err = adapter.Embeddings(ctx, []string{"test"})

		// Anthropic doesn't support embeddings
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not provide an embeddings API")
	})
}

func TestAnthropicAdapter_DefaultValues(t *testing.T) {
	apiKey := "test-api-key"

	t.Run("Default model", func(t *testing.T) {
		adapter, err := NewAnthropicAdapter(apiKey, "", 100, 0.7)
		require.NoError(t, err)
		assert.Equal(t, "claude-sonnet-4-20250514", adapter.Model())
	})

	t.Run("Default max tokens", func(t *testing.T) {
		adapter, err := NewAnthropicAdapter(apiKey, "claude-sonnet-4-20250514", 0, 0.7)
		require.NoError(t, err)
		assert.Equal(t, 1024, adapter.maxTokens)
	})

	t.Run("Default temperature", func(t *testing.T) {
		adapter, err := NewAnthropicAdapter(apiKey, "claude-sonnet-4-20250514", 100, 0)
		require.NoError(t, err)
		assert.Equal(t, float32(0.7), adapter.temperature)
	})

	t.Run("Default base URL", func(t *testing.T) {
		adapter, err := NewAnthropicAdapter(apiKey, "claude-sonnet-4-20250514", 100, 0.7)
		require.NoError(t, err)
		assert.Equal(t, DefaultAnthropicBaseURL, adapter.BaseURL())
	})
}

func TestAnthropicAdapterWithConfig(t *testing.T) {
	t.Run("Custom config", func(t *testing.T) {
		config := AnthropicAdapterConfig{
			APIKey:      "test-key",
			Model:       "claude-3-haiku-20240307",
			MaxTokens:   2048,
			Temperature: 0.5,
			BaseURL:     "https://custom.anthropic.com",
			TopP:        0.9,
			TopK:        40,
			Stop:        []string{"\n\n"},
		}

		adapter, err := NewAnthropicAdapterWithConfig(config)
		require.NoError(t, err)
		assert.Equal(t, "claude-3-haiku-20240307", adapter.Model())
		assert.Equal(t, "https://custom.anthropic.com", adapter.BaseURL())
		assert.Equal(t, 2048, adapter.maxTokens)
		assert.Equal(t, float32(0.5), adapter.temperature)
		assert.Equal(t, float32(0.9), adapter.topP)
		assert.Equal(t, 40, adapter.topK)
		assert.Equal(t, []string{"\n\n"}, adapter.stop)
	})

	t.Run("Missing API key", func(t *testing.T) {
		config := AnthropicAdapterConfig{
			Model: "claude-sonnet-4-20250514",
		}
		_, err := NewAnthropicAdapterWithConfig(config)
		assert.Error(t, err)
	})
}
