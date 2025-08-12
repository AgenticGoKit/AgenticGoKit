package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigurationSystemIntegration tests the complete configuration system end-to-end
func TestConfigurationSystemIntegration(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "config-system-integration")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Run("CompleteConfigurationWorkflow", func(t *testing.T) {
		// Step 1: Create comprehensive configuration
		configContent := `
[agent_flow]
name = "integration-test"
version = "1.0.0"
provider = "mock"

[logging]
level = "info"
format = "json"

[runtime]
max_concurrent_agents = 5
timeout_seconds = 30

[llm]
provider = "mock"
model = "test-model"
temperature = 0.7
max_tokens = 2000

[agents.researcher]
role = "research_specialist"
description = "Conducts comprehensive research"
system_prompt = "You are a research specialist"
capabilities = ["information_gathering", "fact_checking"]
enabled = true
timeout_seconds = 45

[agents.researcher.llm]
temperature = 0.3
max_tokens = 2500

[agents.researcher.retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 5000
backoff_factor = 2.0

[agents.researcher.rate_limit]
requests_per_second = 10
burst_size = 20

[agents.researcher.metadata]
specialization = "research"
priority = "high"

[agents.writer]
role = "content_writer"
description = "Creates engaging content"
system_prompt = "You are a content writer"
capabilities = ["content_creation", "editing"]
enabled = true
timeout_seconds = 30

[agents.disabled_agent]
role = "disabled_role"
description = "This agent is disabled"
system_prompt = "Disabled agent"
capabilities = ["testing"]
enabled = false

[orchestration]
mode = "sequential"
timeout_seconds = 300
sequential_agents = ["researcher", "writer"]

[agent_memory]
provider = "memory"
max_results = 10
dimensions = 1536
auto_embed = true

[mcp]
enabled = false
`

		err := os.WriteFile("agentflow.toml", []byte(configContent), 0644)
		require.NoError(t, err)

		// Step 2: Load and validate configuration
		config, err := LoadConfig("agentflow.toml")
		require.NoError(t, err)
		assert.Equal(t, "integration-test", config.AgentFlow.Name)
		assert.Equal(t, "1.0.0", config.AgentFlow.Version)
		assert.Equal(t, "mock", config.AgentFlow.Provider)

		// Step 3: Validate configuration with comprehensive validator
		validator := NewDefaultConfigValidator()
		validationErrors := validator.ValidateConfig(config)
		
		// Should have minimal validation errors for well-formed config
		errorCount := 0
		for _, err := range validationErrors {
			if err.Message != "" {
				t.Logf("Validation issue: %s - %s", err.Field, err.Message)
				errorCount++
			}
		}
		assert.LessOrEqual(t, errorCount, 5, "Should have minimal validation issues")

		// Step 4: Test configuration resolution
		resolver := NewConfigResolver()
		
		// Resolve researcher agent configuration
		researcherConfig, err := resolver.ResolveAgentConfig("researcher", config)
		require.NoError(t, err)
		
		assert.Equal(t, "researcher", researcherConfig.Name)
		assert.Equal(t, "research_specialist", researcherConfig.Role)
		assert.Equal(t, "You are a research specialist", researcherConfig.SystemPrompt)
		assert.True(t, researcherConfig.Enabled)
		assert.Equal(t, 45*time.Second, researcherConfig.Timeout)
		
		// Check LLM configuration resolution (agent-specific overrides global)
		assert.Equal(t, "mock", researcherConfig.LLMConfig.Provider)
		assert.Equal(t, "test-model", researcherConfig.LLMConfig.Model)
		assert.Equal(t, 0.3, researcherConfig.LLMConfig.Temperature) // Agent override
		assert.Equal(t, 2500, researcherConfig.LLMConfig.MaxTokens)  // Agent override

		// Step 5: Test agent factory
		factory := NewConfigurableAgentFactory(resolver)
		
		// Create individual agent
		researcherAgent, err := factory.CreateAgent("researcher", config)
		require.NoError(t, err)
		assert.NotNil(t, researcherAgent)

		// Create all enabled agents
		agents, err := factory.CreateAllEnabledAgents(config)
		require.NoError(t, err)
		assert.Len(t, agents, 2, "Should create 2 enabled agents")
		assert.Contains(t, agents, "researcher")
		assert.Contains(t, agents, "writer")
		assert.NotContains(t, agents, "disabled_agent", "Should not create disabled agent")

		// Step 6: Test agent execution
		ctx := context.Background()
		event := NewEvent("test", EventData{"message": "test query"}, nil)
		state := NewState()

		result, err := researcherAgent.Run(ctx, event, state)
		require.NoError(t, err)
		assert.NotNil(t, result.OutputState)
	})

	t.Run("EnvironmentVariableOverrides", func(t *testing.T) {
		// Set environment variables
		os.Setenv("AGENTFLOW_LLM_TEMPERATURE", "0.5")
		os.Setenv("AGENTFLOW_AGENT_RESEARCHER_ROLE", "senior_researcher")
		os.Setenv("AGENTFLOW_AGENT_RESEARCHER_LLM_MAX_TOKENS", "3000")
		defer func() {
			os.Unsetenv("AGENTFLOW_LLM_TEMPERATURE")
			os.Unsetenv("AGENTFLOW_AGENT_RESEARCHER_ROLE")
			os.Unsetenv("AGENTFLOW_AGENT_RESEARCHER_LLM_MAX_TOKENS")
		}()

		// Load configuration
		config, err := LoadConfig("agentflow.toml")
		require.NoError(t, err)

		// Resolve with environment overrides
		resolver := NewConfigResolver()
		researcherConfig, err := resolver.ResolveAgentConfig("researcher", config)
		require.NoError(t, err)

		// Check environment overrides were applied
		assert.Equal(t, "senior_researcher", researcherConfig.Role)
		assert.Equal(t, 0.5, researcherConfig.LLMConfig.Temperature) // Global override
		assert.Equal(t, 3000, researcherConfig.LLMConfig.MaxTokens)  // Agent-specific override
	})

	t.Run("ConfigurationHotReload", func(t *testing.T) {
		// Create initial configuration
		initialConfig := `
[agent_flow]
name = "hot-reload-test"
provider = "mock"

[agents.test_agent]
role = "test_role"
description = "Test agent"
system_prompt = "Initial prompt"
capabilities = ["testing"]
enabled = true
timeout_seconds = 30
`
		err := os.WriteFile("hot-reload-test.toml", []byte(initialConfig), 0644)
		require.NoError(t, err)

		// Create config reloader
		reloader, err := NewConfigReloader("hot-reload-test.toml")
		require.NoError(t, err)

		// Track configuration changes
		changeCount := 0
		var lastConfig *Config
		reloader.OnConfigChange(func(newConfig *Config) {
			changeCount++
			lastConfig = newConfig
		})

		// Start watching
		reloader.Start()
		defer reloader.Stop()

		// Wait for initial load
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 1, changeCount, "Should have initial config load")
		assert.NotNil(t, lastConfig)
		assert.Equal(t, "hot-reload-test", lastConfig.AgentFlow.Name)

		// Update configuration
		updatedConfig := `
[agent_flow]
name = "hot-reload-test-updated"
provider = "mock"

[agents.test_agent]
role = "updated_role"
description = "Updated test agent"
system_prompt = "Updated prompt"
capabilities = ["testing", "validation"]
enabled = true
timeout_seconds = 45
`
		err = os.WriteFile("hot-reload-test.toml", []byte(updatedConfig), 0644)
		require.NoError(t, err)

		// Wait for reload
		time.Sleep(200 * time.Millisecond)
		assert.Equal(t, 2, changeCount, "Should have detected config change")
		assert.Equal(t, "hot-reload-test-updated", lastConfig.AgentFlow.Name)
		assert.Equal(t, "updated_role", lastConfig.Agents["test_agent"].Role)
	})

	t.Run("ConfigurationValidationIntegration", func(t *testing.T) {
		// Create configuration with various validation issues
		invalidConfig := `
[agent_flow]
name = "validation-test"
# Missing provider

[agents.invalid_agent]
# Missing role
description = "Invalid agent"
system_prompt = "Hi"  # Too short
capabilities = ["unknown_capability"]
enabled = true
timeout_seconds = -10  # Invalid timeout

[agents.invalid_agent.llm]
temperature = 3.0  # Invalid temperature
max_tokens = 0     # Invalid max tokens

[orchestration]
mode = "sequential"
sequential_agents = ["invalid_agent", "nonexistent_agent"]
`

		err := os.WriteFile("invalid-config.toml", []byte(invalidConfig), 0644)
		require.NoError(t, err)

		// Load configuration (should succeed despite validation issues)
		config, err := LoadConfig("invalid-config.toml")
		require.NoError(t, err)

		// Validate configuration
		validator := NewDefaultConfigValidator()
		validationErrors := validator.ValidateConfig(config)

		// Should have multiple validation errors
		assert.Greater(t, len(validationErrors), 5, "Should have multiple validation errors")

		// Check for specific validation errors
		errorFields := make(map[string]bool)
		for _, err := range validationErrors {
			errorFields[err.Field] = true
		}

		// Should detect missing provider
		assert.True(t, errorFields["llm.provider"] || errorFields["agent_flow.provider"], "Should detect missing provider")
		
		// Should detect invalid agent configuration
		assert.True(t, errorFields["agents.invalid_agent.role"], "Should detect missing role")
		assert.True(t, errorFields["agents.invalid_agent.system_prompt"], "Should detect short system prompt")
		assert.True(t, errorFields["agents.invalid_agent.capabilities"], "Should detect unknown capabilities")
		assert.True(t, errorFields["agents.invalid_agent.timeout_seconds"], "Should detect invalid timeout")
		assert.True(t, errorFields["agents.invalid_agent.llm.temperature"], "Should detect invalid temperature")
		assert.True(t, errorFields["agents.invalid_agent.llm.max_tokens"], "Should detect invalid max tokens")
	})

	t.Run("BackwardCompatibilityTest", func(t *testing.T) {
		// Create minimal legacy-style configuration
		legacyConfig := `
[agent_flow]
name = "legacy-test"
provider = "mock"

[agents.simple_agent]
role = "simple_role"
description = "Simple agent"
capabilities = ["testing"]
enabled = true
`

		err := os.WriteFile("legacy-config.toml", []byte(legacyConfig), 0644)
		require.NoError(t, err)

		// Should load successfully
		config, err := LoadConfig("legacy-config.toml")
		require.NoError(t, err)

		// Should resolve agent configuration with defaults
		resolver := NewConfigResolver()
		agentConfig, err := resolver.ResolveAgentConfig("simple_agent", config)
		require.NoError(t, err)

		// Check defaults are applied
		assert.Equal(t, "simple_agent", agentConfig.Name)
		assert.Equal(t, "simple_role", agentConfig.Role)
		assert.True(t, agentConfig.Enabled)
		assert.Equal(t, 30*time.Second, agentConfig.Timeout) // Default timeout

		// Should create agent successfully
		factory := NewConfigurableAgentFactory(resolver)
		agent, err := factory.CreateAgent("simple_agent", config)
		require.NoError(t, err)
		assert.NotNil(t, agent)
	})

	t.Run("PerformanceOptimizationValidation", func(t *testing.T) {
		// Create configuration with performance issues
		performanceConfig := `
[agent_flow]
name = "performance-test"
provider = "mock"

[agents.slow_agent]
role = "slow_role"
description = "Slow agent"
system_prompt = "You are a slow agent"
capabilities = ["testing"]
enabled = true
timeout_seconds = 300  # Very high timeout

[agents.slow_agent.llm]
temperature = 1.8      # High temperature
max_tokens = 8000      # Very high token limit

[agents.expensive_agent]
role = "expensive_role"
description = "Expensive agent"
system_prompt = "You are an expensive agent"
capabilities = ["testing"]
enabled = true

[agents.expensive_agent.llm]
model = "gpt-4"        # Expensive model
max_tokens = 4000      # High token limit
`

		err := os.WriteFile("performance-config.toml", []byte(performanceConfig), 0644)
		require.NoError(t, err)

		// Load and validate
		config, err := LoadConfig("performance-config.toml")
		require.NoError(t, err)

		validator := NewDefaultConfigValidator()
		validationErrors := validator.ValidateConfig(config)

		// Should have performance-related warnings
		performanceWarnings := 0
		for _, err := range validationErrors {
			if strings.Contains(strings.ToLower(err.Message), "timeout") ||
			   strings.Contains(strings.ToLower(err.Message), "temperature") ||
			   strings.Contains(strings.ToLower(err.Message), "tokens") {
				performanceWarnings++
			}
		}

		assert.Greater(t, performanceWarnings, 0, "Should detect performance issues")
	})
}

// TestConfigurationSystemStressTest tests the system under load
func TestConfigurationSystemStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "config-stress-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Run("ManyAgentsConfiguration", func(t *testing.T) {
		// Create configuration with many agents
		configContent := `
[agent_flow]
name = "stress-test"
provider = "mock"

[llm]
provider = "mock"
model = "test-model"
temperature = 0.7
`

		// Add 50 agents
		for i := 0; i < 50; i++ {
			agentConfig := fmt.Sprintf(`
[agents.agent_%d]
role = "role_%d"
description = "Agent %d description"
system_prompt = "You are agent %d"
capabilities = ["testing", "validation"]
enabled = true
timeout_seconds = 30
`, i, i, i, i)
			configContent += agentConfig
		}

		err := os.WriteFile("stress-config.toml", []byte(configContent), 0644)
		require.NoError(t, err)

		// Load configuration
		start := time.Now()
		config, err := LoadConfig("stress-config.toml")
		loadTime := time.Since(start)
		require.NoError(t, err)
		assert.Len(t, config.Agents, 50)
		t.Logf("Config load time for 50 agents: %v", loadTime)

		// Validate configuration
		start = time.Now()
		validator := NewDefaultConfigValidator()
		validationErrors := validator.ValidateConfig(config)
		validationTime := time.Since(start)
		t.Logf("Validation time for 50 agents: %v", validationTime)
		t.Logf("Validation errors: %d", len(validationErrors))

		// Resolve all agent configurations
		start = time.Now()
		resolver := NewConfigResolver()
		resolvedCount := 0
		for agentName := range config.Agents {
			_, err := resolver.ResolveAgentConfig(agentName, config)
			require.NoError(t, err)
			resolvedCount++
		}
		resolutionTime := time.Since(start)
		assert.Equal(t, 50, resolvedCount)
		t.Logf("Resolution time for 50 agents: %v", resolutionTime)

		// Create all agents
		start = time.Now()
		factory := NewConfigurableAgentFactory(resolver)
		agents, err := factory.CreateAllEnabledAgents(config)
		creationTime := time.Since(start)
		require.NoError(t, err)
		assert.Len(t, agents, 50)
		t.Logf("Agent creation time for 50 agents: %v", creationTime)

		// Performance assertions
		assert.Less(t, loadTime, 1*time.Second, "Config loading should be fast")
		assert.Less(t, validationTime, 2*time.Second, "Validation should be fast")
		assert.Less(t, resolutionTime, 1*time.Second, "Resolution should be fast")
		assert.Less(t, creationTime, 2*time.Second, "Agent creation should be fast")
	})

	t.Run("FrequentConfigurationReloads", func(t *testing.T) {
		// Create initial configuration
		initialConfig := `
[agent_flow]
name = "reload-stress-test"
provider = "mock"

[agents.test_agent]
role = "test_role"
description = "Test agent"
system_prompt = "Test prompt"
capabilities = ["testing"]
enabled = true
`

		err := os.WriteFile("reload-stress.toml", []byte(initialConfig), 0644)
		require.NoError(t, err)

		// Create reloader
		reloader, err := NewConfigReloader("reload-stress.toml")
		require.NoError(t, err)

		changeCount := 0
		reloader.OnConfigChange(func(newConfig *Config) {
			changeCount++
		})

		reloader.Start()
		defer reloader.Stop()

		// Wait for initial load
		time.Sleep(100 * time.Millisecond)
		initialChangeCount := changeCount

		// Perform multiple rapid updates
		start := time.Now()
		for i := 0; i < 10; i++ {
			updatedConfig := fmt.Sprintf(`
[agent_flow]
name = "reload-stress-test-%d"
provider = "mock"

[agents.test_agent]
role = "test_role_%d"
description = "Test agent %d"
system_prompt = "Test prompt %d"
capabilities = ["testing"]
enabled = true
`, i, i, i, i)

			err := os.WriteFile("reload-stress.toml", []byte(updatedConfig), 0644)
			require.NoError(t, err)
			
			// Small delay to allow file system events
			time.Sleep(50 * time.Millisecond)
		}

		// Wait for all reloads to complete
		time.Sleep(500 * time.Millisecond)
		totalTime := time.Since(start)

		finalChangeCount := changeCount
		reloadCount := finalChangeCount - initialChangeCount

		t.Logf("Performed %d reloads in %v", reloadCount, totalTime)
		assert.GreaterOrEqual(t, reloadCount, 5, "Should detect multiple reloads")
		assert.Less(t, totalTime, 10*time.Second, "Reloads should complete quickly")
	})
}

// TestConfigurationSystemErrorHandling tests error handling scenarios
func TestConfigurationSystemErrorHandling(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "config-error-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	t.Run("InvalidTOMLSyntax", func(t *testing.T) {
		// Create configuration with invalid TOML syntax
		invalidTOML := `
[agent_flow
name = "invalid-syntax"  # Missing closing bracket
provider = "mock"

[agents.test_agent]
role = "test_role"
description = "Test agent"
`

		err := os.WriteFile("invalid-syntax.toml", []byte(invalidTOML), 0644)
		require.NoError(t, err)

		// Should fail to load
		_, err = LoadConfig("invalid-syntax.toml")
		assert.Error(t, err, "Should fail to load invalid TOML")
		assert.Contains(t, err.Error(), "parse", "Error should mention parsing issue")
	})

	t.Run("MissingConfigurationFile", func(t *testing.T) {
		// Try to load non-existent file
		_, err := LoadConfig("nonexistent.toml")
		assert.Error(t, err, "Should fail to load non-existent file")
	})

	t.Run("InvalidAgentConfiguration", func(t *testing.T) {
		// Create configuration with invalid agent reference
		invalidAgentConfig := `
[agent_flow]
name = "invalid-agent-test"
provider = "mock"

[orchestration]
mode = "sequential"
sequential_agents = ["nonexistent_agent"]
`

		err := os.WriteFile("invalid-agent.toml", []byte(invalidAgentConfig), 0644)
		require.NoError(t, err)

		// Should load but fail validation
		config, err := LoadConfig("invalid-agent.toml")
		require.NoError(t, err)

		validator := NewDefaultConfigValidator()
		validationErrors := validator.ValidateConfig(config)
		assert.Greater(t, len(validationErrors), 0, "Should have validation errors")

		// Should fail to create agents
		resolver := NewConfigResolver()
		factory := NewConfigurableAgentFactory(resolver)
		
		_, err = factory.CreateAgent("nonexistent_agent", config)
		assert.Error(t, err, "Should fail to create non-existent agent")
	})

	t.Run("ConfigurationReloadFailure", func(t *testing.T) {
		// Create valid initial configuration
		validConfig := `
[agent_flow]
name = "reload-failure-test"
provider = "mock"

[agents.test_agent]
role = "test_role"
description = "Test agent"
system_prompt = "Test prompt"
capabilities = ["testing"]
enabled = true
`

		err := os.WriteFile("reload-failure.toml", []byte(validConfig), 0644)
		require.NoError(t, err)

		// Create reloader
		reloader, err := NewConfigReloader("reload-failure.toml")
		require.NoError(t, err)

		successCount := 0
		errorCount := 0
		
		reloader.OnConfigChange(func(newConfig *Config) {
			successCount++
		})
		
		reloader.OnError(func(err error) {
			errorCount++
		})

		reloader.Start()
		defer reloader.Stop()

		// Wait for initial load
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 1, successCount, "Should have successful initial load")

		// Update with invalid configuration
		invalidConfig := `
[agent_flow
name = "invalid"  # Invalid TOML syntax
`

		err = os.WriteFile("reload-failure.toml", []byte(invalidConfig), 0644)
		require.NoError(t, err)

		// Wait for reload attempt
		time.Sleep(200 * time.Millisecond)
		
		// Should have detected error and not updated success count
		assert.Equal(t, 1, successCount, "Should not have additional successful loads")
		assert.Greater(t, errorCount, 0, "Should have detected reload error")
	})
}

// Helper function to check if string contains substring (case-insensitive)
func strings.Contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}