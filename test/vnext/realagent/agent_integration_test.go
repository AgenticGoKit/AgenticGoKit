package realagent_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isOllamaAvailable checks if Ollama is running locally
func isOllamaAvailable() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// TestRealAgent_IntegrationRun tests the agent with real Ollama (if available)
func TestRealAgent_IntegrationRun(t *testing.T) {
	if !isOllamaAvailable() {
		t.Skip("Skipping integration test: Ollama not available")
	}

	config := &vnext.Config{
		Name:         "integration-test-agent",
		SystemPrompt: "You are a helpful assistant. Keep answers very short and concise.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.7,
			MaxTokens:   50,
			BaseURL:     "http://localhost:11434",
		},
	}

	// Build agent
	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		Build()

	require.NoError(t, err, "Should build agent successfully")
	require.NotNil(t, agent, "Agent should not be nil")

	// Initialize
	ctx := context.Background()
	err = agent.Initialize(ctx)
	require.NoError(t, err, "Should initialize successfully")
	defer agent.Cleanup(ctx)

	// Test Run method with real LLM
	t.Run("Run with real LLM", func(t *testing.T) {
		queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		result, err := agent.Run(queryCtx, "What is 2+2?")

		require.NoError(t, err, "Should run successfully")
		require.NotNil(t, result, "Result should not be nil")

		// Verify result structure
		assert.True(t, result.Success, "Result should be successful")
		assert.NotEmpty(t, result.Content, "Should have content")
		assert.Greater(t, result.Duration, time.Duration(0), "Should have duration")

		// Verify we got a real response (not mock)
		assert.NotContains(t, result.Content, "processed:", "Should not contain mock text")

		// Check the response makes sense for a math question
		// The answer should contain "four" or "4"
		content := strings.ToLower(result.Content)
		containsFour := strings.Contains(content, "four") || strings.Contains(content, "4")

		assert.True(t, containsFour, "Response should contain the answer to 2+2")

		t.Logf("Response: %s", result.Content)
		t.Logf("Duration: %v", result.Duration)
	})

	// Test capabilities
	t.Run("Capabilities include LLM", func(t *testing.T) {
		capabilities := agent.Capabilities()
		assert.Contains(t, capabilities, "llm", "Should have LLM capability")
		assert.Contains(t, capabilities, "streaming", "Should have streaming capability")
	})
}

// TestRealAgent_ContextCancellation tests that context cancellation works
func TestRealAgent_ContextCancellation(t *testing.T) {
	if !isOllamaAvailable() {
		t.Skip("Skipping integration test: Ollama not available")
	}

	config := &vnext.Config{
		Name:         "cancellation-test-agent",
		SystemPrompt: "You are a helpful assistant.",
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
		Timeout: 30 * time.Second,
	}

	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		Build()

	require.NoError(t, err)
	require.NotNil(t, agent)

	ctx := context.Background()
	err = agent.Initialize(ctx)
	require.NoError(t, err)
	defer agent.Cleanup(ctx)

	// Create a context with very short timeout
	queryCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	cancel() // Cancel immediately

	result, err := agent.Run(queryCtx, "Tell me a long story")

	// Should get an error due to cancelled context
	assert.Error(t, err, "Should return error for cancelled context")
	assert.Nil(t, result, "Result should be nil on error")
}

// TestRealAgent_MultipleRuns tests running the agent multiple times
func TestRealAgent_MultipleRuns(t *testing.T) {
	if !isOllamaAvailable() {
		t.Skip("Skipping integration test: Ollama not available")
	}

	config := &vnext.Config{
		Name:         "multi-run-test-agent",
		SystemPrompt: "You are a helpful assistant. Answer in one short sentence.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.7,
			MaxTokens:   50,
			BaseURL:     "http://localhost:11434",
		},
	}

	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		Build()

	require.NoError(t, err)
	require.NotNil(t, agent)

	ctx := context.Background()
	err = agent.Initialize(ctx)
	require.NoError(t, err)
	defer agent.Cleanup(ctx)

	// Run multiple queries
	queries := []string{
		"What is 1+1?",
		"What is 2+2?",
		"What is 3+3?",
	}

	for i, query := range queries {
		t.Run(query, func(t *testing.T) {
			queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			result, err := agent.Run(queryCtx, query)

			require.NoError(t, err, "Run %d should succeed", i+1)
			require.NotNil(t, result, "Result %d should not be nil", i+1)
			assert.True(t, result.Success, "Result %d should be successful", i+1)
			assert.NotEmpty(t, result.Content, "Result %d should have content", i+1)

			t.Logf("Query %d: %s -> %s (Duration: %v)", i+1, query, result.Content, result.Duration)
		})
	}
}

// BenchmarkAgentRun benchmarks the agent Run method
func BenchmarkAgentRun(b *testing.B) {
	if !isOllamaAvailable() {
		b.Skip("Skipping benchmark: Ollama not available")
	}

	config := &vnext.Config{
		Name:         "benchmark-agent",
		SystemPrompt: "Answer very briefly.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.7,
			MaxTokens:   20,
			BaseURL:     "http://localhost:11434",
		},
	}

	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		Build()

	if err != nil {
		b.Fatalf("Failed to build agent: %v", err)
	}

	ctx := context.Background()
	if err := agent.Initialize(ctx); err != nil {
		b.Fatalf("Failed to initialize agent: %v", err)
	}
	defer agent.Cleanup(ctx)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, err := agent.Run(queryCtx, "Hi")
		cancel()

		if err != nil {
			b.Errorf("Run failed: %v", err)
		}
	}
}



