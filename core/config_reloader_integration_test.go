package core

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigReloaderIntegration(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "agentflow.toml")

	// Initial configuration
	initialConfig := `
[agent_flow]
name = "test-flow"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 800

[agents.researcher]
role = "research_specialist"
description = "Gathers comprehensive information"
system_prompt = "You are a research specialist..."
capabilities = ["information_gathering", "fact_checking"]
enabled = true

[agents.researcher.llm]
temperature = 0.3
max_tokens = 1200

[agents.writer]
role = "content_writer"
description = "Creates engaging content"
system_prompt = "You are a content writer..."
capabilities = ["content_creation", "editing"]
enabled = true
`

	// Write initial config
	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Create real components
	validator := NewDefaultConfigValidator()
	config := &Config{} // We'll load the real config when we start watching
	factory := NewConfigurableAgentFactory(config)
	agentManager := NewAgentManager(factory)
	reloader := NewConfigReloader(validator, agentManager)

	// Set shorter debounce for testing
	reloader.SetDebouncePeriod(100 * time.Millisecond)

	// Track configuration changes
	var configChanges []string
	var changesMutex sync.Mutex
	var changesWg sync.WaitGroup

	reloader.OnConfigChanged(func(config *Config, err error) {
		changesMutex.Lock()
		defer changesMutex.Unlock()

		if err != nil {
			configChanges = append(configChanges, "ERROR: "+err.Error())
		} else if config != nil {
			configChanges = append(configChanges, "SUCCESS: "+config.AgentFlow.Name)
		}
		changesWg.Done()
	})

	// Start watching
	err = reloader.StartWatching(configPath)
	require.NoError(t, err)
	defer reloader.StopWatching()

	// Verify initial state
	assert.True(t, reloader.IsWatching())
	assert.NotZero(t, reloader.GetLastReloadTime())

	// Check initial agents
	agents := agentManager.GetCurrentAgents()
	assert.Len(t, agents, 2)
	assert.Contains(t, agentManager.ListAgents(), "researcher")
	assert.Contains(t, agentManager.ListAgents(), "writer")
	assert.Equal(t, 2, agentManager.GetEnabledAgentCount())

	// Test 1: Update existing agent configuration
	changesWg.Add(1)
	updatedConfig1 := `
[agent_flow]
name = "test-flow"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 800

[agents.researcher]
role = "senior_research_specialist"
description = "Advanced research capabilities"
system_prompt = "You are a senior research specialist with advanced capabilities..."
capabilities = ["information_gathering", "fact_checking", "source_identification"]
enabled = true

[agents.researcher.llm]
temperature = 0.2
max_tokens = 1500

[agents.writer]
role = "content_writer"
description = "Creates engaging content"
system_prompt = "You are a content writer..."
capabilities = ["content_creation", "editing"]
enabled = true
`

	err = os.WriteFile(configPath, []byte(updatedConfig1), 0644)
	require.NoError(t, err)

	// Wait for change to be processed
	done := make(chan struct{})
	go func() {
		changesWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for first config change")
	}

	// Verify changes were applied
	researcherConfig, exists := agentManager.GetAgentConfig("researcher")
	require.True(t, exists)
	assert.Equal(t, "senior_research_specialist", researcherConfig.Role)
	assert.Equal(t, "Advanced research capabilities", researcherConfig.Description)
	assert.Contains(t, researcherConfig.Capabilities, "source_identification")

	// Test 2: Add new agent
	changesWg.Add(1)
	updatedConfig2 := `
[agent_flow]
name = "test-flow"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 800

[agents.researcher]
role = "senior_research_specialist"
description = "Advanced research capabilities"
system_prompt = "You are a senior research specialist with advanced capabilities..."
capabilities = ["information_gathering", "fact_checking", "source_identification"]
enabled = true

[agents.researcher.llm]
temperature = 0.2
max_tokens = 1500

[agents.writer]
role = "content_writer"
description = "Creates engaging content"
system_prompt = "You are a content writer..."
capabilities = ["content_creation", "editing"]
enabled = true

[agents.analyzer]
role = "data_analyst"
description = "Analyzes data and provides insights"
system_prompt = "You are a data analyst..."
capabilities = ["data_analysis", "pattern_recognition"]
enabled = true

[agents.analyzer.llm]
temperature = 0.5
max_tokens = 1000
`

	err = os.WriteFile(configPath, []byte(updatedConfig2), 0644)
	require.NoError(t, err)

	// Wait for change to be processed
	done2 := make(chan struct{})
	go func() {
		changesWg.Wait()
		close(done2)
	}()

	select {
	case <-done2:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for second config change")
	}

	// Add a small delay to ensure all processing is complete
	time.Sleep(200 * time.Millisecond)

	// Debug: Print current agent state
	t.Logf("Current agent count: %d", agentManager.GetAgentCount())
	t.Logf("Current agents: %v", agentManager.ListAgents())

	// Verify new agent was added
	assert.Equal(t, 3, agentManager.GetAgentCount())
	assert.Contains(t, agentManager.ListAgents(), "analyzer")

	analyzerConfig, exists := agentManager.GetAgentConfig("analyzer")
	require.True(t, exists)
	assert.Equal(t, "data_analyst", analyzerConfig.Role)
	assert.True(t, analyzerConfig.Enabled)

	// Test 3: Disable an agent
	changesWg.Add(1)
	updatedConfig3 := `
[agent_flow]
name = "test-flow"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 800

[agents.researcher]
role = "senior_research_specialist"
description = "Advanced research capabilities"
system_prompt = "You are a senior research specialist with advanced capabilities..."
capabilities = ["information_gathering", "fact_checking", "source_identification"]
enabled = true

[agents.researcher.llm]
temperature = 0.2
max_tokens = 1500

[agents.writer]
role = "content_writer"
description = "Creates engaging content"
system_prompt = "You are a content writer..."
capabilities = ["content_creation", "editing"]
enabled = false

[agents.analyzer]
role = "data_analyst"
description = "Analyzes data and provides insights"
system_prompt = "You are a data analyst..."
capabilities = ["data_analysis", "pattern_recognition"]
enabled = true

[agents.analyzer.llm]
temperature = 0.5
max_tokens = 1000
`

	err = os.WriteFile(configPath, []byte(updatedConfig3), 0644)
	require.NoError(t, err)

	// Wait for change to be processed
	done3 := make(chan struct{})
	go func() {
		changesWg.Wait()
		close(done3)
	}()

	select {
	case <-done3:
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for third config change")
	}

	// Verify writer was disabled
	assert.Equal(t, 2, agentManager.GetEnabledAgentCount())
	enabledAgents := agentManager.ListEnabledAgents()
	assert.Contains(t, enabledAgents, "researcher")
	assert.Contains(t, enabledAgents, "analyzer")
	assert.NotContains(t, enabledAgents, "writer")

	writerConfig, exists := agentManager.GetAgentConfig("writer")
	require.True(t, exists)
	assert.False(t, writerConfig.Enabled)

	// Test 4: Remove an agent from configuration
	changesWg.Add(1)
	updatedConfig4 := `
[agent_flow]
name = "test-flow"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 800

[agents.researcher]
role = "senior_research_specialist"
description = "Advanced research capabilities"
system_prompt = "You are a senior research specialist with advanced capabilities..."
capabilities = ["information_gathering", "fact_checking", "source_identification"]
enabled = true

[agents.researcher.llm]
temperature = 0.2
max_tokens = 1500

[agents.analyzer]
role = "data_analyst"
description = "Analyzes data and provides insights"
system_prompt = "You are a data analyst..."
capabilities = ["data_analysis", "pattern_recognition"]
enabled = true

[agents.analyzer.llm]
temperature = 0.5
max_tokens = 1000
`

	err = os.WriteFile(configPath, []byte(updatedConfig4), 0644)
	require.NoError(t, err)

	// Wait for change to be processed
	done4 := make(chan struct{})
	go func() {
		changesWg.Wait()
		close(done4)
	}()

	select {
	case <-done4:
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for fourth config change")
	}

	// Verify writer was removed/disabled
	enabledAgents = agentManager.ListEnabledAgents()
	assert.Len(t, enabledAgents, 2)
	assert.Contains(t, enabledAgents, "researcher")
	assert.Contains(t, enabledAgents, "analyzer")
	assert.NotContains(t, enabledAgents, "writer")

	// Test 5: Invalid configuration should not break the system
	changesWg.Add(1)
	invalidConfig := `
[agent_flow
name = "invalid-config"  # Missing closing bracket
`

	err = os.WriteFile(configPath, []byte(invalidConfig), 0644)
	require.NoError(t, err)

	// Wait for change to be processed
	done5 := make(chan struct{})
	go func() {
		changesWg.Wait()
		close(done5)
	}()

	select {
	case <-done5:
		// Success - error should be handled gracefully
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for invalid config change")
	}

	// Verify system still works with previous valid configuration
	enabledAgents = agentManager.ListEnabledAgents()
	assert.Len(t, enabledAgents, 2) // Should still have the agents from the last valid config

	// Check that we received error notification
	changesMutex.Lock()
	hasError := false
	for _, change := range configChanges {
		if len(change) > 6 && change[:6] == "ERROR:" {
			hasError = true
			break
		}
	}
	changesMutex.Unlock()
	assert.True(t, hasError, "Should have received error notification for invalid config")

	// Test 6: Manual reload
	// First, fix the config
	err = os.WriteFile(configPath, []byte(updatedConfig4), 0644)
	require.NoError(t, err)

	// Manual reload
	err = reloader.ReloadConfig()
	assert.NoError(t, err)

	// Verify reload time was updated
	assert.True(t, time.Since(reloader.GetLastReloadTime()) < time.Second)

	// Final verification
	assert.True(t, reloader.IsWatching())
	assert.Equal(t, 2, agentManager.GetEnabledAgentCount())

	Logger().Info().
		Strs("config_changes", configChanges).
		Int("total_changes", len(configChanges)).
		Msg("Configuration hot-reload integration test completed")
}

func TestConfigReloaderIntegration_EnvironmentOverrides(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("AGENTFLOW_LLM_TEMPERATURE", "0.9")
	os.Setenv("AGENTFLOW_AGENT_RESEARCHER_SYSTEM_PROMPT", "Enhanced research prompt from environment")
	defer func() {
		os.Unsetenv("AGENTFLOW_LLM_TEMPERATURE")
		os.Unsetenv("AGENTFLOW_AGENT_RESEARCHER_SYSTEM_PROMPT")
	}()

	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "agentflow.toml")

	configContent := `
[agent_flow]
name = "env-test-flow"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.5

[agents.researcher]
role = "researcher"
system_prompt = "Original prompt"
enabled = true
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create components
	validator := NewDefaultConfigValidator()
	config := &Config{} // We'll load the real config when we start watching
	factory := NewConfigurableAgentFactory(config)
	agentManager := NewAgentManager(factory)
	reloader := NewConfigReloader(validator, agentManager)

	// Start watching
	err = reloader.StartWatching(configPath)
	require.NoError(t, err)
	defer reloader.StopWatching()

	// Verify environment overrides were applied
	researcherConfig, exists := agentManager.GetAgentConfig("researcher")
	require.True(t, exists)

	// The system prompt should be overridden by environment variable
	assert.Equal(t, "Enhanced research prompt from environment", researcherConfig.SystemPrompt)

	// The LLM temperature should be overridden by environment variable
	assert.Equal(t, 0.9, researcherConfig.LLMConfig.Temperature)
}
