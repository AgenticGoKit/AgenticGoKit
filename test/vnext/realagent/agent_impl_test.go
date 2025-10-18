package realagent_test

import (
	"context"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRealAgent_ValidConfig tests creating an agent with valid configuration
func TestNewRealAgent_ValidConfig(t *testing.T) {
	config := &vnext.Config{
		Name:         "test-agent",
		SystemPrompt: "You are a helpful assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.7,
			MaxTokens:   100,
			BaseURL:     "http://localhost:11434",
		},
	}

	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		Build()

	require.NoError(t, err, "Should create agent without error")
	require.NotNil(t, agent, "Agent should not be nil")

	// Test agent properties
	assert.Equal(t, "test-agent", agent.Name(), "Agent name should match config")
	assert.NotNil(t, agent.Config(), "Agent config should not be nil")
	assert.Equal(t, config.Name, agent.Config().Name, "Config name should match")
}

// TestNewRealAgent_InvalidConfig tests error handling with invalid configurations
func TestNewRealAgent_InvalidConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *vnext.Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "", // Error message may vary based on which validation fails first
		},
		{
			name: "Empty agent name",
			config: &vnext.Config{
				Name:         "",
				SystemPrompt: "Test",
				LLM: vnext.LLMConfig{
					Provider: "ollama",
					Model:    "gemma3:1b",
				},
			},
			expectError: true,
			errorMsg:    "agent name is required",
		},
		{
			name: "Empty LLM provider",
			config: &vnext.Config{
				Name:         "test",
				SystemPrompt: "Test",
				LLM: vnext.LLMConfig{
					Provider: "",
					Model:    "gemma3:1b",
				},
			},
			expectError: true,
			errorMsg:    "LLM provider is required",
		},
		{
			name: "Empty LLM model",
			config: &vnext.Config{
				Name:         "test",
				SystemPrompt: "Test",
				LLM: vnext.LLMConfig{
					Provider: "ollama",
					Model:    "",
				},
			},
			expectError: true,
			errorMsg:    "LLM model is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var agent vnext.Agent
			var err error

			if tt.config == nil {
				// Test with nil config directly
				agent, err = vnext.NewBuilder("test").Build()
			} else {
				agent, err = vnext.NewBuilder(tt.config.Name).
					WithConfig(tt.config).
					Build()
			}

			if tt.expectError {
				assert.Error(t, err, "Should return error for invalid config")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
				assert.Nil(t, agent, "Agent should be nil on error")
			} else {
				assert.NoError(t, err, "Should not return error")
				assert.NotNil(t, agent, "Agent should not be nil")
			}
		})
	}
}

// TestAgentCapabilities tests the Capabilities() method
func TestAgentCapabilities(t *testing.T) {
	tests := []struct {
		name               string
		config             *vnext.Config
		expectedCapabilies []string
	}{
		{
			name: "LLM only",
			config: &vnext.Config{
				Name:         "llm-agent",
				SystemPrompt: "Test",
				LLM: vnext.LLMConfig{
					Provider: "ollama",
					Model:    "gemma3:1b",
					BaseURL:  "http://localhost:11434",
				},
				Timeout: 30 * time.Second,
			},
			expectedCapabilies: []string{"llm", "streaming"},
		},
		{
			name: "LLM with memory",
			config: &vnext.Config{
				Name:         "memory-agent",
				SystemPrompt: "Test",
				LLM: vnext.LLMConfig{
					Provider: "ollama",
					Model:    "gemma3:1b",
					BaseURL:  "http://localhost:11434",
				},
				Memory: &vnext.MemoryConfig{
					Provider:   "memory",
					Connection: "memory",
				},
				Timeout: 30 * time.Second,
			},
			expectedCapabilies: []string{"llm", "memory", "streaming"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := vnext.NewBuilder(tt.config.Name).
				WithConfig(tt.config).
				Build()

			require.NoError(t, err)
			require.NotNil(t, agent)

			capabilities := agent.Capabilities()
			assert.NotEmpty(t, capabilities, "Should have capabilities")

			// Check that expected capabilities are present
			for _, expected := range tt.expectedCapabilies {
				assert.Contains(t, capabilities, expected, "Should contain capability: "+expected)
			}
		})
	}
}

// TestAgentInitializeAndCleanup tests the Initialize and Cleanup lifecycle
func TestAgentInitializeAndCleanup(t *testing.T) {
	config := &vnext.Config{
		Name:         "lifecycle-agent",
		SystemPrompt: "Test",
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

	// Test Initialize
	err = agent.Initialize(ctx)
	assert.NoError(t, err, "Initialize should succeed")

	// Test Initialize is idempotent (can be called multiple times)
	err = agent.Initialize(ctx)
	assert.NoError(t, err, "Second Initialize should also succeed")

	// Test Cleanup
	err = agent.Cleanup(ctx)
	assert.NoError(t, err, "Cleanup should succeed")
}

// TestAgentName tests the Name() method
func TestAgentName(t *testing.T) {
	config := &vnext.Config{
		Name:         "named-agent",
		SystemPrompt: "Test",
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

	name := agent.Name()
	assert.Equal(t, "named-agent", name, "Agent name should match")
}

// TestAgentConfig tests the Config() method
func TestAgentConfig(t *testing.T) {
	config := &vnext.Config{
		Name:         "config-agent",
		SystemPrompt: "You are a test assistant.",
		Timeout:      60 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.5,
			MaxTokens:   150,
			BaseURL:     "http://localhost:11434",
		},
	}

	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		Build()

	require.NoError(t, err)
	require.NotNil(t, agent)

	retrievedConfig := agent.Config()
	require.NotNil(t, retrievedConfig, "Config should not be nil")

	// Verify config values
	assert.Equal(t, config.Name, retrievedConfig.Name)
	assert.Equal(t, config.SystemPrompt, retrievedConfig.SystemPrompt)
	assert.Equal(t, config.LLM.Provider, retrievedConfig.LLM.Provider)
	assert.Equal(t, config.LLM.Model, retrievedConfig.LLM.Model)
	assert.Equal(t, config.LLM.Temperature, retrievedConfig.LLM.Temperature)
}

// TestBuilderFreeze tests that builder cannot be modified after Build()
func TestBuilderFreeze(t *testing.T) {
	config := &vnext.Config{
		Name:         "test-agent",
		SystemPrompt: "Test",
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
		Timeout: 30 * time.Second,
	}

	builder := vnext.NewBuilder(config.Name).
		WithConfig(config)

	// Build the agent
	agent, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, agent)

	// Try to build again - should fail with frozen error
	_, err = builder.Build()
	assert.Error(t, err, "Should not allow building twice")
	assert.Contains(t, err.Error(), "frozen", "Error should mention frozen builder")
}

// TestBuilderWithLLMConfig tests the WithConfig builder method with LLM config
func TestBuilderWithLLMConfig(t *testing.T) {
	config := &vnext.Config{
		Name:         "llm-test-agent",
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

	agentConfig := agent.Config()
	assert.Equal(t, "ollama", agentConfig.LLM.Provider)
	assert.Equal(t, "gemma3:1b", agentConfig.LLM.Model)
}

// TestBuilderWithMemory tests the WithMemory builder method
func TestBuilderWithMemory(t *testing.T) {
	config := &vnext.Config{
		Name:         "memory-test-agent",
		SystemPrompt: "Test",
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
		Timeout: 30 * time.Second,
	}

	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		WithMemory().
		Build()

	require.NoError(t, err)
	require.NotNil(t, agent)

	capabilities := agent.Capabilities()
	assert.Contains(t, capabilities, "memory", "Should have memory capability")

	agentConfig := agent.Config()
	require.NotNil(t, agentConfig.Memory, "Memory config should be set")
}

// TestBuilderClone tests the Clone() method
func TestBuilderClone(t *testing.T) {
	config1 := &vnext.Config{
		Name:         "original-agent",
		SystemPrompt: "Original prompt",
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
		Timeout: 30 * time.Second,
	}

	original := vnext.NewBuilder(config1.Name).
		WithConfig(config1)

	// Clone the builder
	cloned := original.Clone()

	// Build from original
	agent1, err1 := original.Build()
	require.NoError(t, err1)

	// Modify cloned config and build
	config2 := &vnext.Config{
		Name:         "cloned-agent",
		SystemPrompt: "Cloned prompt",
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
		Timeout: 30 * time.Second,
	}

	agent2, err2 := cloned.WithConfig(config2).Build()
	require.NoError(t, err2)

	// Verify they are different
	assert.NotEqual(t, agent1.Config().SystemPrompt, agent2.Config().SystemPrompt)
	assert.Equal(t, "Original prompt", agent1.Config().SystemPrompt)
	assert.Equal(t, "Cloned prompt", agent2.Config().SystemPrompt)
}
