package options_test

import (
	"context"
	"testing"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// createTestAgent creates a test agent with standard configuration
func createTestAgent(t *testing.T, name string) vnext.Agent {
	agent, err := vnext.NewBuilder(name).
		WithConfig(&vnext.Config{
			Name:         name,
			SystemPrompt: "You are a test assistant. Respond concisely.",
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: 0.7,
				MaxTokens:   50,
			},
			Timeout: 30 * time.Second,
		}).
		Build()

	require.NoError(t, err)
	require.NotNil(t, agent)

	// Initialize the agent
	err = agent.Initialize(context.Background())
	require.NoError(t, err)

	return agent
} // =============================================================================
// TESTS FOR RunWithOptions
// =============================================================================

func TestRunWithOptions_NilOptions(t *testing.T) {
	agent := createTestAgent(t, "test-agent")
	defer agent.Cleanup(context.Background())

	ctx := context.Background()

	// When opts is nil, should delegate to Run()
	result, err := agent.RunWithOptions(ctx, "Say 'hello'", nil)

	// Should work with Ollama
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Content)
}

func TestRunWithOptions_Timeout(t *testing.T) {
	agent := createTestAgent(t, "test-agent")
	defer agent.Cleanup(context.Background())

	ctx := context.Background()

	opts := &vnext.RunOptions{
		Timeout: 30 * time.Second,
	}

	// The timeout should be applied to the context
	result, err := agent.RunWithOptions(ctx, "What is 2+2?", opts)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Content)
}

func TestRunWithOptions_TemperatureOverride(t *testing.T) {
	agent := createTestAgent(t, "test-agent")
	defer agent.Cleanup(context.Background())

	// Access the internal config to verify override
	originalTemp := agent.Config().LLM.Temperature

	ctx := context.Background()
	newTemp := 0.3
	opts := &vnext.RunOptions{
		Temperature: &newTemp,
	}

	// Execute with temperature override
	result, err := agent.RunWithOptions(ctx, "Say 'test'", opts)
	require.NoError(t, err)
	assert.True(t, result.Success)

	// After execution, config should be restored to original
	// (defer in RunWithOptions should restore it)
	finalTemp := agent.Config().LLM.Temperature
	assert.Equal(t, originalTemp, finalTemp, "Temperature should be restored after run")
}

func TestRunWithOptions_MaxTokensOverride(t *testing.T) {
	agent := createTestAgent(t, "test-agent")
	defer agent.Cleanup(context.Background())

	originalMaxTokens := agent.Config().LLM.MaxTokens

	ctx := context.Background()
	opts := &vnext.RunOptions{
		MaxTokens: 30, // Override to smaller value
	}

	result, err := agent.RunWithOptions(ctx, "Count to 5", opts)
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Config should be restored
	finalMaxTokens := agent.Config().LLM.MaxTokens
	assert.Equal(t, originalMaxTokens, finalMaxTokens, "MaxTokens should be restored")
}

func TestRunWithOptions_ConfigurationRestoration(t *testing.T) {
	// Test that all configuration is properly restored after RunWithOptions
	agent := createTestAgent(t, "test-agent")
	defer agent.Cleanup(context.Background())

	// Capture original config
	originalConfig := agent.Config()
	originalTemp := originalConfig.LLM.Temperature
	originalMaxTokens := originalConfig.LLM.MaxTokens

	ctx := context.Background()
	newTemp := 0.3
	opts := &vnext.RunOptions{
		Temperature: &newTemp,
		MaxTokens:   30,
		Timeout:     30 * time.Second,
	}

	// Execute with overrides
	result, err := agent.RunWithOptions(ctx, "Say hello", opts)
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify config was restored
	finalConfig := agent.Config()
	assert.Equal(t, originalTemp, finalConfig.LLM.Temperature, "Temperature restored")
	assert.Equal(t, originalMaxTokens, finalConfig.LLM.MaxTokens, "MaxTokens restored")
}

func TestRunWithOptions_MultipleOptionsSimultaneously(t *testing.T) {
	// Test applying multiple options at once
	agent := createTestAgent(t, "test-agent")
	defer agent.Cleanup(context.Background())

	ctx := context.Background()
	temp := 0.5
	opts := &vnext.RunOptions{
		Timeout:        30 * time.Second,
		Temperature:    &temp,
		MaxTokens:      40,
		SessionID:      "multi-test-session",
		DetailedResult: true,
		ToolMode:       "auto",
		MaxRetries:     3,
	}

	// Execute with all options
	result, err := agent.RunWithOptions(ctx, "What is 1+1?", opts)

	// Should work with all options applied
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Content)

	// Verify detailed result metadata was added
	if opts.DetailedResult {
		assert.NotNil(t, result.Metadata)
		assert.Contains(t, result.Metadata, "timeout")
		assert.Contains(t, result.Metadata, "temperature_override")
	}
}

// =============================================================================
// TESTS REQUIRING REAL PROVIDERS (SKIPPED)
// =============================================================================

func TestRunWithOptions_ToolMode(t *testing.T) {
	t.Skip("Tool configuration requires real LLM provider setup")
}

func TestRunWithOptions_DetailedResult(t *testing.T) {
	t.Skip("Requires real LLM provider to fully test result enhancement")
}

func TestRunWithOptions_IncludeTrace(t *testing.T) {
	t.Skip("Requires real tracing system to fully test")
}

func TestRunWithOptions_MemoryOptions(t *testing.T) {
	t.Skip("Requires real memory provider to fully test")
}



