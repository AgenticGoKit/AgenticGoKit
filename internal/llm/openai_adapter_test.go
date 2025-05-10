package llm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIAdapter_Complete(t *testing.T) {
	t.Run("Valid prompts", func(t *testing.T) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY environment variable is not set")
		}

		adapter, err := NewOpenAIAdapter(apiKey, "gpt-4o-mini", 50, 0.7)
		require.NoError(t, err)

		ctx := context.Background()
		response, err := adapter.Complete(ctx, "System prompt", "User prompt")

		// Assertions
		assert.NoError(t, err)
		assert.NotEmpty(t, response)
	})

	t.Run("Empty prompts", func(t *testing.T) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY environment variable is not set")
		}

		adapter, err := NewOpenAIAdapter(apiKey, "gpt-4o-mini", 50, 0.7)
		require.NoError(t, err)

		ctx := context.Background()
		response, err := adapter.Complete(ctx, "", "")

		// Assertions
		assert.Error(t, err)
		assert.Empty(t, response)
	})
}
