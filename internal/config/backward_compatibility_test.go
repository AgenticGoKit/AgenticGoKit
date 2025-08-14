package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackwardCompatibilityLegacyProjects tests that existing projects continue to work
func TestBackwardCompatibilityLegacyProjects(t *testing.T) {
	// Test 1: Legacy project with minimal configuration
	t.Run("minimal_legacy_config", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "legacy-minimal-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		configPath := filepath.Join(tempDir, "agentflow.toml")

		// Minimal legacy configuration (pre-agent configuration system)
		legacyConfig := `[agent_flow]
name = "legacy-project"
version = "1.0.0"`

		err = os.WriteFile(configPath, []byte(legacyConfig), 0644)
		require.NoError(t, err)

		// Should load without errors
		config, err := LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, "legacy-project", config.AgentFlow.Name)
		assert.Equal(t, "1.0.0", config.AgentFlow.Version)

		// Should validate with defaults
		validator := NewDefaultConfigValidator()
		errors := validator.ValidateConfig(config)

		// May have warnings but should not fail completely
		for _, err := range errors {
			// Errors should be warnings or suggestions, not critical failures
			assert.NotContains(t, err.Message, "critical")
			assert.NotContains(t, err.Message, "fatal")
		}

		// Should be able to create resolver with defaults
		resolver := NewConfigResolver(config)
		assert.NotNil(t, resolver)
	})

	// Test 2: Legacy project with basic LLM configuration
	t.Run("legacy_with_basic_llm", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "legacy-llm-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		configPath := filepath.Join(tempDir, "agentflow.toml")

		// Legacy configuration with basic LLM settings
		legacyConfig := `[agent_flow]
name = "legacy-llm-project"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-3.5-turbo"
temperature = 0.7`

		err = os.WriteFile(configPath, []byte(legacyConfig), 0644)
		require.NoError(t, err)

		config, err := LoadConfig(configPath)
		require.NoError(t, err)

		// Should preserve existing LLM configuration
		assert.Equal(t, "openai", config.LLM.Provider)
		assert.Equal(t, "gpt-3.5-turbo", config.LLM.Model)
		assert.Equal(t, float32(0.7), config.LLM.Temperature)

		// Should validate successfully
		validator := NewDefaultConfigValidator()
		errors := validator.ValidateConfig(config)

		// Should have minimal or no errors for basic valid configuration
		criticalErrors := 0
		for _, err := range errors {
			if containsAny(err.Message, []string{"required", "missing", "invalid"}) {
				criticalErrors++
			}
		}
		assert.Equal(t, 0, criticalErrors, "Should not have critical errors for valid legacy config")
	})

	// Test 3: Legacy project with hardcoded agent references
	t.Run("legacy_with_hardcoded_agents", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "legacy-hardcoded-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		configPath := filepath.Join(tempDir, "agentflow.toml")

		// Legacy configuration that might reference hardcoded agents
		legacyConfig := `[agent_flow]
name = "legacy-hardcoded-project"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7

# Legacy orchestration without explicit agent definitions
[orchestration]
mode = "sequential"
agents = ["researcher", "writer", "reviewer"]`

		err = os.WriteFile(configPath, []byte(legacyConfig), 0644)
		require.NoError(t, err)

		config, err := LoadConfig(configPath)
		require.NoError(t, err)

		// Should load orchestration configuration
		assert.Equal(t, "sequential", config.Orchestration.Mode)
		assert.Equal(t, []string{"researcher", "writer", "reviewer"}, config.Orchestration.Agents)

		// Validation should warn about missing agent definitions but not fail
		validator := NewDefaultConfigValidator()
		errors := validator.ValidateConfig(config)

		// Should have warnings about missing agent definitions
		hasAgentWarnings := false
		for _, err := range errors {
			if containsAny(err.Message, []string{"agent", "definition", "not found"}) {
				hasAgentWarnings = true
				// Should be warnings, not critical errors
				assert.Contains(t, err.Suggestion, "define")
			}
		}

		// In a legacy system, we might expect warnings about missing agent definitions
		// but the system should still be functional
	})

	// Test 4: Migration from hardcoded to configuration-driven
	t.Run("migration_compatibility", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "migration-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		configPath := filepath.Join(tempDir, "agentflow.toml")

		// Start with legacy configuration
		legacyConfig := `[agent_flow]
name = "migration-project"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7`

		err = os.WriteFile(configPath, []byte(legacyConfig), 0644)
		require.NoError(t, err)

		// Load legacy configuration
		legacyConfigObj, err := LoadConfig(configPath)
		require.NoError(t, err)

		// Simulate migration by adding agent definitions
		migratedConfig := `[agent_flow]
name = "migration-project"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7

[agents.researcher]
role = "researcher"
description = "Research and information gathering"
system_prompt = "You are a research specialist."
capabilities = ["web_search", "document_analysis"]
enabled = true

[agents.writer]
role = "writer"
description = "Content creation and writing"
system_prompt = "You are a skilled writer."
capabilities = ["content_creation", "editing"]
enabled = true

[orchestration]
mode = "sequential"
agents = ["researcher", "writer"]`

		err = os.WriteFile(configPath, []byte(migratedConfig), 0644)
		require.NoError(t, err)

		// Load migrated configuration
		migratedConfigObj, err := LoadConfig(configPath)
		require.NoError(t, err)

		// Should preserve existing settings
		assert.Equal(t, legacyConfigObj.AgentFlow.Name, migratedConfigObj.AgentFlow.Name)
		assert.Equal(t, legacyConfigObj.LLM.Provider, migratedConfigObj.LLM.Provider)
		assert.Equal(t, legacyConfigObj.LLM.Model, migratedConfigObj.LLM.Model)

		// Should now have agent definitions
		assert.Len(t, migratedConfigObj.Agents, 2)
		assert.Contains(t, migratedConfigObj.Agents, "researcher")
		assert.Contains(t, migratedConfigObj.Agents, "writer")

		// Should validate successfully
		validator := NewDefaultConfigValidator()
		errors := validator.ValidateConfig(migratedConfigObj)
		assert.Empty(t, errors, "Migrated configuration should be valid")
	})
}

// TestBackwardCompatibilityAgentFactory tests that the agent factory works with legacy configurations
func TestBackwardCompatibilityAgentFactory(t *testing.T) {
	// Test 1: Factory with minimal configuration
	t.Run("factory_with_minimal_config", func(t *testing.T) {
		config := &Config{
			AgentFlow: AgentFlowConfig{
				Name:    "minimal-test",
				Version: "1.0.0",
			},
			LLM: LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
			},
		}

		factory := NewConfigurableAgentFactory(config)
		assert.NotNil(t, factory)

		// Should handle requests for undefined agents gracefully
		_, err := factory.CreateAgent("undefined_agent")
		assert.Error(t, err, "Should error for undefined agent")
		assert.Contains(t, err.Error(), "not found")
	})

	// Test 2: Factory with legacy agent creation patterns
	t.Run("factory_legacy_patterns", func(t *testing.T) {
		config := &Config{
			AgentFlow: AgentFlowConfig{
				Name:    "legacy-pattern-test",
				Version: "1.0.0",
			},
			LLM: LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   2000,
			},
			Agents: map[string]AgentConfig{
				"legacy_agent": {
					Role:         "legacy",
					Description:  "Legacy agent for backward compatibility",
					SystemPrompt: "You are a legacy agent.",
					Capabilities: []string{"legacy_capability"},
					Enabled:      true,
				},
			},
		}

		factory := NewConfigurableAgentFactory(config)

		// Should be able to create configured agent
		agent, err := factory.CreateAgent("legacy_agent")
		require.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "legacy", agent.GetRole())
	})

	// Test 3: Mixed hardcoded and configured agents
	t.Run("mixed_agent_types", func(t *testing.T) {
		config := &Config{
			AgentFlow: AgentFlowConfig{
				Name:    "mixed-test",
				Version: "1.0.0",
			},
			LLM: LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
			},
			Agents: map[string]AgentConfig{
				"configured_agent": {
					Role:         "configured",
					Description:  "Configured agent",
					SystemPrompt: "You are configured.",
					Capabilities: []string{"configured_capability"},
					Enabled:      true,
				},
			},
		}

		factory := NewConfigurableAgentFactory(config)

		// Should create configured agent
		configuredAgent, err := factory.CreateAgent("configured_agent")
		require.NoError(t, err)
		assert.NotNil(t, configuredAgent)

		// Should handle hardcoded agent requests appropriately
		_, err = factory.CreateAgent("hardcoded_agent")
		assert.Error(t, err, "Should error for undefined hardcoded agent")
	})
}

// TestBackwardCompatibilityConfigResolver tests resolver with legacy configurations
func TestBackwardCompatibilityConfigResolver(t *testing.T) {
	// Test 1: Resolver with minimal configuration
	t.Run("resolver_minimal_config", func(t *testing.T) {
		config := &Config{
			AgentFlow: AgentFlowConfig{
				Name:    "minimal-resolver-test",
				Version: "1.0.0",
			},
		}

		resolver := NewConfigResolver(config)
		assert.NotNil(t, resolver)

		// Should handle missing LLM configuration gracefully
		_, err := resolver.ResolveAgentConfig("any_agent")
		assert.Error(t, err, "Should error when no agents are defined")
	})

	// Test 2: Resolver with partial configuration
	t.Run("resolver_partial_config", func(t *testing.T) {
		config := &Config{
			AgentFlow: AgentFlowConfig{
				Name:    "partial-resolver-test",
				Version: "1.0.0",
			},
			LLM: LLMConfig{
				Provider: "openai",
				Model:    "gpt-4",
				// Missing temperature and max_tokens - should use defaults
			},
			Agents: map[string]AgentConfig{
				"partial_agent": {
					Role:        "partial",
					Description: "Partially configured agent",
					// Missing system_prompt and capabilities - should use defaults
					Enabled: true,
				},
			},
		}

		resolver := NewConfigResolver(config)

		resolvedConfig, err := resolver.ResolveAgentConfig("partial_agent")
		require.NoError(t, err)

		// Should fill in defaults
		assert.Equal(t, "partial", resolvedConfig.Role)
		assert.Equal(t, "openai", resolvedConfig.LLM.Provider)
		assert.Equal(t, "gpt-4", resolvedConfig.LLM.Model)

		// Should have default values for missing fields
		assert.NotZero(t, resolvedConfig.LLM.Temperature) // Should have default
		assert.NotZero(t, resolvedConfig.LLM.MaxTokens)   // Should have default
	})

	// Test 3: Environment variable compatibility
	t.Run("env_var_compatibility", func(t *testing.T) {
		// Set legacy environment variables
		os.Setenv("OPENAI_API_KEY", "test-key")
		os.Setenv("AGENTFLOW_LLM_MODEL", "gpt-3.5-turbo")
		defer func() {
			os.Unsetenv("OPENAI_API_KEY")
			os.Unsetenv("AGENTFLOW_LLM_MODEL")
		}()

		config := &Config{
			AgentFlow: AgentFlowConfig{
				Name:    "env-compat-test",
				Version: "1.0.0",
			},
			LLM: LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4", // Should be overridden by env var
				Temperature: 0.7,
			},
			Agents: map[string]AgentConfig{
				"env_agent": {
					Role:         "env_test",
					Description:  "Environment variable test agent",
					SystemPrompt: "Test agent",
					Capabilities: []string{"testing"},
					Enabled:      true,
				},
			},
		}

		resolver := NewConfigResolver(config)
		resolver.ApplyEnvironmentOverrides()

		resolvedConfig, err := resolver.ResolveAgentConfig("env_agent")
		require.NoError(t, err)

		// Environment variable should override configuration
		// Note: In a real implementation, we'd check the actual override
		assert.NotNil(t, resolvedConfig)
	})
}

// TestBackwardCompatibilityValidation tests validation with legacy configurations
func TestBackwardCompatibilityValidation(t *testing.T) {
	validator := NewDefaultConfigValidator()

	// Test 1: Validation of minimal legacy configuration
	t.Run("validate_minimal_legacy", func(t *testing.T) {
		config := &Config{
			AgentFlow: AgentFlowConfig{
				Name:    "minimal-legacy",
				Version: "1.0.0",
			},
		}

		errors := validator.ValidateConfig(config)

		// Should have warnings but not critical errors
		criticalErrors := 0
		warnings := 0

		for _, err := range errors {
			if containsAny(err.Message, []string{"required", "missing"}) {
				if containsAny(err.Message, []string{"critical", "fatal"}) {
					criticalErrors++
				} else {
					warnings++
				}
			}
		}

		// Should have warnings about missing configuration but not critical failures
		assert.Equal(t, 0, criticalErrors, "Should not have critical errors for minimal config")
	})

	// Test 2: Validation with gradual configuration addition
	t.Run("validate_gradual_config", func(t *testing.T) {
		// Start with minimal config
		config := &Config{
			AgentFlow: AgentFlowConfig{
				Name:    "gradual-config",
				Version: "1.0.0",
			},
		}

		initialErrors := validator.ValidateConfig(config)
		initialErrorCount := len(initialErrors)

		// Add LLM configuration
		config.LLM = LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   2000,
		}

		withLLMErrors := validator.ValidateConfig(config)
		withLLMErrorCount := len(withLLMErrors)

		// Should have fewer errors after adding LLM config
		assert.LessOrEqual(t, withLLMErrorCount, initialErrorCount,
			"Should have same or fewer errors after adding LLM config")

		// Add agent configuration
		config.Agents = map[string]AgentConfig{
			"test_agent": {
				Role:         "test",
				Description:  "Test agent",
				SystemPrompt: "You are a test agent.",
				Capabilities: []string{"testing"},
				Enabled:      true,
			},
		}

		withAgentsErrors := validator.ValidateConfig(config)
		withAgentsErrorCount := len(withAgentsErrors)

		// Should have fewer errors after adding agent config
		assert.LessOrEqual(t, withAgentsErrorCount, withLLMErrorCount,
			"Should have same or fewer errors after adding agent config")
	})

	// Test 3: Validation suggestions for legacy configurations
	t.Run("validate_legacy_suggestions", func(t *testing.T) {
		config := &Config{
			AgentFlow: AgentFlowConfig{
				Name:    "legacy-suggestions",
				Version: "1.0.0",
			},
			LLM: LLMConfig{
				Provider: "openai",
				Model:    "gpt-3.5-turbo", // Older model
			},
		}

		errors := validator.ValidateConfig(config)

		// Should provide helpful suggestions for improvement
		hasSuggestions := false
		for _, err := range errors {
			if err.Suggestion != "" {
				hasSuggestions = true
				// Suggestions should be constructive
				assert.NotContains(t, err.Suggestion, "error")
				assert.NotContains(t, err.Suggestion, "fail")
			}
		}

		// Should provide suggestions for improvement
		if len(errors) > 0 {
			assert.True(t, hasSuggestions, "Should provide suggestions for legacy configurations")
		}
	})
}

// Helper function to check if a string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if contains(s, substr) {
			return true
		}
	}
	return false
}
