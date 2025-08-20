package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigDrivenScaffold tests that the scaffold generates configuration-driven projects
func TestConfigDrivenScaffold(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "config-driven-scaffold-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Test configuration
	config := ProjectConfig{
		Name:              "test-config-project",
		NumAgents:         3,
		Provider:          "openai",
		OrchestrationMode: "sequential",
		MemoryEnabled:     true,
		MemoryProvider:    "pgvector",
		EmbeddingProvider: "openai",
		EmbeddingModel:    "text-embedding-ada-002",
		RAGEnabled:        true,
		MCPEnabled:        true,
	}

	// Create the project
	err = CreateAgentProjectModular(config)
	require.NoError(t, err)

	// Test 1: Verify agentflow.toml contains agent definitions
	configPath := filepath.Join(config.Name, "agentflow.toml")
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configStr := string(configContent)

	// Should contain agent definitions
	assert.Contains(t, configStr, "[agents.agent1]", "Should contain agent1 definition")
	assert.Contains(t, configStr, "[agents.agent2]", "Should contain agent2 definition")
	assert.Contains(t, configStr, "[agents.agent3]", "Should contain agent3 definition")

	// Should contain agent-specific LLM settings
	assert.Contains(t, configStr, "[agents.agent1.llm]", "Should contain agent1 LLM config")
	assert.Contains(t, configStr, "temperature =", "Should contain temperature settings")

	// Should contain auto_llm configuration
	assert.Contains(t, configStr, "auto_llm = true", "Should contain auto_llm configuration")

	// Should contain retry policies
	assert.Contains(t, configStr, "[agents.agent1.retry_policy]", "Should contain retry policy")
	assert.Contains(t, configStr, "max_retries = 3", "Should contain retry configuration")

	// Should contain global LLM configuration
	assert.Contains(t, configStr, "[llm]", "Should contain global LLM config")
	assert.Contains(t, configStr, "provider = \"openai\"", "Should contain provider config")
	assert.Contains(t, configStr, "model = \"gpt-4\"", "Should contain model config")

	// Test 2: Verify main.go uses ConfigurableAgentFactory
	mainPath := filepath.Join(config.Name, "main.go")
	mainContent, err := os.ReadFile(mainPath)
	require.NoError(t, err)

	mainStr := string(mainContent)

	// Should use configuration-driven approach
	assert.Contains(t, mainStr, "core.NewConfigurableAgentFactory", "Should use ConfigurableAgentFactory")
	assert.Contains(t, mainStr, "core.NewAgentManager", "Should use AgentManager")
	assert.Contains(t, mainStr, "agentManager.GetActiveAgents()", "Should get active agents from manager")

	// Should not contain hardcoded agent creation
	assert.NotContains(t, mainStr, "agents.NewAgent1", "Should not contain hardcoded agent creation")
	assert.NotContains(t, mainStr, "agents.NewAgent2", "Should not contain hardcoded agent creation")

	// Should contain configuration loading
	assert.Contains(t, mainStr, "core.LoadConfig(\"agentflow.toml\")", "Should load configuration")
	assert.Contains(t, mainStr, "core.NewRunnerFromConfig", "Should create runner from config")

	// Test 3: Verify agent files are configuration-aware
	agentPath := filepath.Join(config.Name, "agents", "agent1.go")
	agentContent, err := os.ReadFile(agentPath)
	require.NoError(t, err)

	agentStr := string(agentContent)

	// Should mention configuration-driven approach
	assert.Contains(t, agentStr, "configuration-driven", "Should mention configuration-driven approach")
	assert.Contains(t, agentStr, "ConfigurableAgentFactory", "Should mention ConfigurableAgentFactory")
	assert.Contains(t, agentStr, "agentflow.toml", "Should mention configuration file")

	// Should contain configuration-aware methods
	assert.Contains(t, agentStr, "ResolvedAgentConfig", "Should use ResolvedAgentConfig")
	assert.Contains(t, agentStr, "GetRole()", "Should have GetRole method")
	assert.Contains(t, agentStr, "GetCapabilities()", "Should have GetCapabilities method")

	// Test 4: Verify README mentions configuration system
	readmePath := filepath.Join(config.Name, "README.md")
	readmeContent, err := os.ReadFile(readmePath)
	require.NoError(t, err)

	readmeStr := string(readmeContent)

	// Should mention configuration-driven architecture
	assert.Contains(t, readmeStr, "Configuration-Driven Architecture", "Should mention config-driven architecture")
	assert.Contains(t, readmeStr, "agentcli validate", "Should mention validation command")
	assert.Contains(t, readmeStr, "agentflow.toml", "Should mention config file")
	assert.Contains(t, readmeStr, "No hardcoded agents", "Should mention no hardcoded agents")

	// Test 5: Verify configuration structure is valid
	// Check that agent definitions have required fields
	assert.Contains(t, configStr, "role =", "Should contain agent roles")
	assert.Contains(t, configStr, "description =", "Should contain agent descriptions")
	assert.Contains(t, configStr, "system_prompt =", "Should contain system prompts")
	assert.Contains(t, configStr, "capabilities =", "Should contain capabilities")
	assert.Contains(t, configStr, "enabled = true", "Should contain enabled flags")

	// Test 6: Verify orchestration configuration
	assert.Contains(t, configStr, "[orchestration]", "Should contain orchestration config")
	assert.Contains(t, configStr, "mode = \"sequential\"", "Should contain orchestration mode")

	// Test 7: Verify memory configuration is present
	assert.Contains(t, configStr, "[agent_memory]", "Should contain memory config")
	assert.Contains(t, configStr, "provider = \"pgvector\"", "Should contain memory provider")
	assert.Contains(t, configStr, "[agent_memory.embedding]", "Should contain embedding config")

	// Test 8: Verify MCP configuration is present
	assert.Contains(t, configStr, "[mcp]", "Should contain MCP config")
	assert.Contains(t, configStr, "enabled = true", "Should enable MCP")
}

// TestAgentConfigGeneration tests the agent configuration generation specifically
func TestAgentConfigGeneration(t *testing.T) {
	config := ProjectConfig{
		Name:              "test-agent-config",
		NumAgents:         2,
		Provider:          "openai",
		OrchestrationMode: "collaborative",
	}

	// Test agent config generation
	agentConfig := generateAgentConfig(config)

	// Should contain agent definitions
	assert.Contains(t, agentConfig, "[agents.agent1]", "Should contain agent1")
	assert.Contains(t, agentConfig, "[agents.agent2]", "Should contain agent2")

	// Should contain different temperatures for different agents
	assert.Contains(t, agentConfig, "temperature =", "Should contain temperature settings")

	// Should contain capabilities arrays
	assert.Contains(t, agentConfig, "capabilities = [", "Should contain capabilities arrays")

	// Should contain retry policies
	assert.Contains(t, agentConfig, "[agents.agent1.retry_policy]", "Should contain retry policies")
}

// TestTemperatureVariation tests that different agents get different temperatures
func TestTemperatureVariation(t *testing.T) {
	// Test temperature assignment for different agent types
	tests := []struct {
		agentName           string
		expectedTemperature float64
	}{
		{"researcher", 0.3},
		{"creative_writer", 0.8},
		{"reviewer", 0.2},
		{"analyst", 0.5},
		{"agent1", 0.7}, // First agent gets 0.7
	}

	for _, tt := range tests {
		t.Run(tt.agentName, func(t *testing.T) {
			temp := getTemperatureForAgent(tt.agentName, 0)
			assert.Equal(t, tt.expectedTemperature, temp,
				"Agent %s should have temperature %.1f", tt.agentName, tt.expectedTemperature)
		})
	}
}

// TestMaxTokensVariation tests that different agents get different max_tokens
func TestMaxTokensVariation(t *testing.T) {
	tests := []struct {
		agentName      string
		expectedTokens int
	}{
		{"writer", 3000},
		{"content_creator", 3000},
		{"researcher", 2500},
		{"analyst", 2500},
		{"reviewer", 1500},
		{"summary_agent", 1000},
		{"agent1", 2000}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.agentName, func(t *testing.T) {
			tokens := getMaxTokensForAgent(tt.agentName, 0)
			assert.Equal(t, tt.expectedTokens, tokens,
				"Agent %s should have max_tokens %d", tt.agentName, tt.expectedTokens)
		})
	}
}

// TestCapabilitiesFormatting tests the capabilities array formatting
func TestCapabilitiesFormatting(t *testing.T) {
	tests := []struct {
		capabilities []string
		expected     string
	}{
		{[]string{}, `["general_assistance"]`},
		{[]string{"processing"}, `["processing"]`},
		{[]string{"web_search", "analysis"}, `["web_search", "analysis"]`},
		{[]string{"research", "fact_checking", "writing"}, `["research", "fact_checking", "writing"]`},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.capabilities, ","), func(t *testing.T) {
			result := formatCapabilitiesArray(tt.capabilities)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDefaultModelForProvider tests the default model selection
func TestDefaultModelForProvider(t *testing.T) {
	tests := []struct {
		provider string
		expected string
	}{
		{"openai", "gpt-4"},
		{"azure", "gpt-4"},
		{"ollama", "llama2"},
		{"mock", "mock-model"},
		{"unknown", "gpt-4"}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			result := getDefaultModelForProvider(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}
