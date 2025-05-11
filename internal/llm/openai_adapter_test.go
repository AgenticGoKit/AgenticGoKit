package llm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIAdapter_Call(t *testing.T) {
	t.Run("Valid prompt", func(t *testing.T) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY environment variable is not set")
		}

		adapter, err := NewOpenAIAdapter(apiKey, "gpt-4o-mini", 50, 0.7)
		require.NoError(t, err)

		ctx := context.Background()
		response, err := adapter.Call(ctx, "User prompt")

		// Assertions
		assert.NoError(t, err)
		assert.NotEmpty(t, response)
	})

	t.Run("Empty prompt", func(t *testing.T) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY environment variable is not set")
		}

		adapter, err := NewOpenAIAdapter(apiKey, "gpt-4o-mini", 50, 0.7)
		require.NoError(t, err)

		ctx := context.Background()
		response, err := adapter.Call(ctx, "")

		// Assertions
		assert.Error(t, err)
		assert.Empty(t, response)
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
		embeddings, err := adapter.Embeddings(ctx, "Test input")

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
		embeddings, err := adapter.Embeddings(ctx, "")

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, embeddings)
	})
}
