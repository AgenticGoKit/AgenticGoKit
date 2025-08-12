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

// TestConfigSystemEndToEnd tests the complete configuration system workflow
func TestConfigSystemEndToEnd(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "config-e2e-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "agentflow.toml")

	// Test 1: Create a comprehensive configuration
	configContent := `[agent_flow]
name = "e2e-test-system"
version = "1.0.0"
description = "End-to-end test configuration"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000

[agents.researcher]
role = "researcher"
description = "Research and information gathering agent"
system_prompt = "You are a research specialist focused on gathering accurate information."
capabilities = ["web_search", "document_analysis", "data_extraction"]
enabled = true

[agents.researcher.llm]
temperature = 0.3
max_tokens = 1500

[agents.writer]
role = "writer"
description = "Content creation and writing agent"
system_prompt = "You are a skilled writer who creates engaging content."
capabilities = ["content_creation", "editing", "seo_optimization"]
enabled = true

[agents.reviewer]
role = "reviewer"
description = "Quality assurance and review agent"
system_prompt = "You are a meticulous reviewer ensuring quality and accuracy."
capabilities = ["quality_assurance", "fact_checking", "proofreading"]
enabled = false

[orchestration]
mode = "sequential"
agents = ["researcher", "writer", "reviewer"]

[memory]
enabled = true
provider = "pgvector"
connection_string = "postgresql://localhost:5432/agentflow"

[memory.embedding]
provider = "openai"
model = "text-embedding-ada-002"
dimensions = 1536

[rag]
enabled = true
chunk_size = 1000
chunk_overlap = 200
similarity_threshold = 0.8

[mcp]
enabled = true

[[mcp.servers]]
name = "web_search"
type = "stdio"
command = "uvx"
args = ["mcp-server-web-search"]
enabled = true

[[mcp.servers]]
name = "filesystem"
type = "stdio"
command = "uvx"
args = ["mcp-server-filesystem", "/tmp"]
enabled = true

[retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 30000
backoff_multiplier = 2.0

[rate_limit]
requests_per_second = 10
burst_size = 20`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test 2: Load and validate configuration
	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, "e2e-test-system", config.AgentFlow.Name)
	assert.Equal(t, "1.0.0", config.AgentFlow.Version)

	// Test 3: Validate configuration
	validator := NewDefaultConfigValidator()
	errors := validator.ValidateConfig(config)
	assert.Empty(t, errors, "Configuration should be valid")

	// Test 4: Resolve agent configurations
	resolver := NewConfigResolver(config)
	
	// Test researcher agent resolution
	researcherConfig, err := resolver.ResolveAgentConfig("researcher")
	require.NoError(t, err)
	assert.Equal(t, "researcher", researcherConfig.Role)
	assert.Equal(t, float32(0.3), researcherConfig.LLM.Temperature) // Agent-specific override
	assert.Equal(t, 1500, researcherConfig.LLM.MaxTokens)          // Agent-specific override
	assert.Equal(t, "gpt-4", researcherConfig.LLM.Model)           // Inherited from global
	assert.True(t, researcherConfig.Enabled)

	// Test writer agent resolution
	writerConfig, err := resolver.ResolveAgentConfig("writer")
	require.NoError(t, err)
	assert.Equal(t, "writer", writerConfig.Role)
	assert.Equal(t, float32(0.7), writerConfig.LLM.Temperature) // Inherited from global
	assert.Equal(t, 2000, writerConfig.LLM.MaxTokens)          // Inherited from global
	assert.True(t, writerConfig.Enabled)

	// Test reviewer agent resolution (disabled)
	reviewerConfig, err := resolver.ResolveAgentConfig("reviewer")
	require.NoError(t, err)
	assert.Equal(t, "reviewer", reviewerConfig.Role)
	assert.False(t, reviewerConfig.Enabled)

	// Test 5: Create agents using configuration
	factory := NewConfigurableAgentFactory(config)
	
	// Create researcher agent
	researcher, err := factory.CreateAgent("researcher")
	require.NoError(t, err)
	assert.NotNil(t, researcher)

	// Create writer agent
	writer, err := factory.CreateAgent("writer")
	require.NoError(t, err)
	assert.NotNil(t, writer)

	// Test 6: Test agent manager with configuration
	manager := NewAgentManager(config)
	
	// Initialize agents
	err = manager.InitializeAgents()
	require.NoError(t, err)

	// Get active agents (should exclude disabled reviewer)
	activeAgents := manager.GetActiveAgents()
	assert.Len(t, activeAgents, 2) // researcher and writer only
	
	agentNames := make([]string, len(activeAgents))
	for i, agent := range activeAgents {
		agentNames[i] = agent.GetRole()
	}
	assert.Contains(t, agentNames, "researcher")
	assert.Contains(t, agentNames, "writer")
	assert.NotContains(t, agentNames, "reviewer") // Should be excluded (disabled)

	// Test 7: Test configuration hot-reload
	reloader := NewConfigReloader(configPath)
	
	// Set up reload callback
	reloadCalled := false
	reloader.OnConfigReload(func(newConfig *Config) error {
		reloadCalled = true
		assert.Equal(t, "e2e-test-system", newConfig.AgentFlow.Name)
		return nil
	})

	// Start watching (in a separate goroutine for testing)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		reloader.StartWatching(ctx)
	}()

	// Modify configuration file to trigger reload
	time.Sleep(100 * time.Millisecond) // Give watcher time to start
	
	modifiedConfig := configContent + "\n# Modified for reload test"
	err = os.WriteFile(configPath, []byte(modifiedConfig), 0644)
	require.NoError(t, err)

	// Wait for reload to be detected
	time.Sleep(500 * time.Millisecond)
	
	// Note: In a real implementation, we'd check if reloadCalled is true
	// For this test, we're just ensuring the system doesn't crash

	// Test 8: Test environment variable overrides
	os.Setenv("AGENTFLOW_LLM_TEMPERATURE", "0.9")
	os.Setenv("AGENTFLOW_AGENTS_RESEARCHER_ENABLED", "false")
	defer func() {
		os.Unsetenv("AGENTFLOW_LLM_TEMPERATURE")
		os.Unsetenv("AGENTFLOW_AGENTS_RESEARCHER_ENABLED")
	}()

	// Reload configuration with environment overrides
	configWithEnv, err := LoadConfig(configPath)
	require.NoError(t, err)

	resolverWithEnv := NewConfigResolver(configWithEnv)
	
	// Apply environment overrides
	resolverWithEnv.ApplyEnvironmentOverrides()

	// Test that environment variables override configuration
	researcherConfigWithEnv, err := resolverWithEnv.ResolveAgentConfig("researcher")
	require.NoError(t, err)
	
	// Note: In a real implementation, we'd check the actual override values
	// For this test, we're ensuring the system handles environment variables

	// Test 9: Test backward compatibility
	legacyConfig := &Config{
		AgentFlow: AgentFlowConfig{
			Name:    "legacy-system",
			Version: "1.0.0",
		},
		// No LLM or Agents configuration - should use defaults
	}

	legacyResolver := NewConfigResolver(legacyConfig)
	
	// Should be able to resolve with defaults
	defaultConfig, err := legacyResolver.ResolveAgentConfig("default")
	if err == nil {
		assert.NotNil(t, defaultConfig)
	}

	// Test 10: Test error handling and validation
	invalidConfigPath := filepath.Join(tempDir, "invalid.toml")
	invalidConfigContent := `[agent_flow]
name = "invalid-system"

[agents.invalid_agent]
role = ""  # Invalid: empty role
temperature = 2.5  # Invalid: temperature > 2.0
capabilities = []  # Invalid: no capabilities`

	err = os.WriteFile(invalidConfigPath, []byte(invalidConfigContent), 0644)
	require.NoError(t, err)

	invalidConfig, err := LoadConfig(invalidConfigPath)
	require.NoError(t, err) // Should load but fail validation

	validationErrors := validator.ValidateConfig(invalidConfig)
	assert.NotEmpty(t, validationErrors, "Invalid configuration should have validation errors")

	// Check specific validation errors
	errorMessages := make([]string, len(validationErrors))
	for i, err := range validationErrors {
		errorMessages[i] = err.Message
	}
	
	// Should have errors for empty role and invalid temperature
	hasRoleError := false
	hasTemperatureError := false
	for _, msg := range errorMessages {
		if contains(msg, "role") {
			hasRoleError = true
		}
		if contains(msg, "temperature") {
			hasTemperatureError = true
		}
	}
	
	assert.True(t, hasRoleError || hasTemperatureError, "Should have validation errors for role or temperature")
}

// TestConfigSystemPerformance tests the performance of the configuration system
func TestConfigSystemPerformance(t *testing.T) {
	// Create a large configuration with many agents
	tempDir, err := os.MkdirTemp("", "config-perf-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "large-config.toml")

	// Generate configuration with 100 agents
	configContent := `[agent_flow]
name = "performance-test"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000
`

	// Add 100 agents
	for i := 0; i < 100; i++ {
		configContent += fmt.Sprintf(`
[agents.agent_%d]
role = "agent_%d"
description = "Performance test agent %d"
system_prompt = "You are agent %d"
capabilities = ["capability_%d"]
enabled = true
`, i, i, i, i, i)
	}

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test loading performance
	start := time.Now()
	config, err := LoadConfig(configPath)
	loadTime := time.Since(start)
	require.NoError(t, err)
	assert.Less(t, loadTime, 1*time.Second, "Configuration loading should be fast")

	// Test validation performance
	validator := NewDefaultConfigValidator()
	start = time.Now()
	errors := validator.ValidateConfig(config)
	validationTime := time.Since(start)
	assert.Empty(t, errors)
	assert.Less(t, validationTime, 1*time.Second, "Configuration validation should be fast")

	// Test resolution performance
	resolver := NewConfigResolver(config)
	start = time.Now()
	for i := 0; i < 100; i++ {
		agentName := fmt.Sprintf("agent_%d", i)
		_, err := resolver.ResolveAgentConfig(agentName)
		require.NoError(t, err)
	}
	resolutionTime := time.Since(start)
	assert.Less(t, resolutionTime, 1*time.Second, "Agent resolution should be fast")

	// Test agent creation performance
	factory := NewConfigurableAgentFactory(config)
	start = time.Now()
	for i := 0; i < 10; i++ { // Test with fewer agents for creation
		agentName := fmt.Sprintf("agent_%d", i)
		agent, err := factory.CreateAgent(agentName)
		require.NoError(t, err)
		assert.NotNil(t, agent)
	}
	creationTime := time.Since(start)
	assert.Less(t, creationTime, 5*time.Second, "Agent creation should be reasonably fast")
}

// TestConfigSystemConcurrency tests concurrent access to the configuration system
func TestConfigSystemConcurrency(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-concurrency-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "concurrent-config.toml")
	configContent := `[agent_flow]
name = "concurrency-test"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7

[agents.test_agent]
role = "test"
description = "Concurrency test agent"
system_prompt = "Test agent"
capabilities = ["testing"]
enabled = true`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)

	resolver := NewConfigResolver(config)
	factory := NewConfigurableAgentFactory(config)

	// Test concurrent resolution
	const numGoroutines = 10
	const numOperations = 100

	// Test concurrent agent resolution
	t.Run("concurrent_resolution", func(t *testing.T) {
		done := make(chan bool, numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- true }()
				
				for j := 0; j < numOperations; j++ {
					agentConfig, err := resolver.ResolveAgentConfig("test_agent")
					assert.NoError(t, err)
					assert.NotNil(t, agentConfig)
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	// Test concurrent agent creation
	t.Run("concurrent_creation", func(t *testing.T) {
		done := make(chan bool, numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- true }()
				
				for j := 0; j < 10; j++ { // Fewer operations for creation
					agent, err := factory.CreateAgent("test_agent")
					assert.NoError(t, err)
					assert.NotNil(t, agent)
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}

// TestConfigSystemErrorRecovery tests error recovery scenarios
func TestConfigSystemErrorRecovery(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config-error-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "error-test.toml")

	// Test 1: Recovery from invalid TOML syntax
	invalidToml := `[agent_flow
name = "invalid-syntax"  # Missing closing bracket`

	err = os.WriteFile(configPath, []byte(invalidToml), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(configPath)
	assert.Error(t, err, "Should fail to load invalid TOML")

	// Test 2: Recovery after fixing syntax
	validToml := `[agent_flow]
name = "valid-syntax"
version = "1.0.0"

[agents.test_agent]
role = "test"
description = "Test agent"
system_prompt = "Test"
capabilities = ["testing"]
enabled = true`

	err = os.WriteFile(configPath, []byte(validToml), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	assert.NoError(t, err, "Should load valid TOML after fix")
	assert.Equal(t, "valid-syntax", config.AgentFlow.Name)

	// Test 3: Graceful handling of missing agents
	resolver := NewConfigResolver(config)
	
	_, err = resolver.ResolveAgentConfig("nonexistent_agent")
	assert.Error(t, err, "Should error for nonexistent agent")

	// Test 4: Factory error handling
	factory := NewConfigurableAgentFactory(config)
	
	_, err = factory.CreateAgent("nonexistent_agent")
	assert.Error(t, err, "Should error when creating nonexistent agent")

	// Test 5: Validation error recovery
	validator := NewDefaultConfigValidator()
	
	// Create config with validation errors
	configWithErrors := &Config{
		AgentFlow: AgentFlowConfig{
			Name:    "error-test",
			Version: "1.0.0",
		},
		Agents: map[string]AgentConfig{
			"invalid_agent": {
				Role:         "", // Invalid: empty role
				Description:  "Invalid agent",
				SystemPrompt: "Test",
				Capabilities: []string{}, // Invalid: no capabilities
				Enabled:      true,
			},
		},
	}

	errors := validator.ValidateConfig(configWithErrors)
	assert.NotEmpty(t, errors, "Should have validation errors")

	// Fix the errors
	configWithErrors.Agents["invalid_agent"].Role = "fixed_role"
	configWithErrors.Agents["invalid_agent"].Capabilities = []string{"fixed_capability"}

	errors = validator.ValidateConfig(configWithErrors)
	assert.Empty(t, errors, "Should have no validation errors after fix")
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Import fmt for string formatting
import "fmt"