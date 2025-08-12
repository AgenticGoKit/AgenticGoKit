package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidationIntegration(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "agentflow_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Valid configuration
	validConfig := `
[agent_flow]
name = "test-project"
version = "1.0.0"
provider = "openai"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 800
timeout_seconds = 30

[agents.researcher]
role = "research_specialist"
description = "Gathers comprehensive information from multiple sources"
system_prompt = "You are a research specialist who excels at finding and analyzing information from various sources."
capabilities = ["information_gathering", "fact_checking", "source_identification"]
enabled = true
timeout_seconds = 45

[agents.researcher.llm]
temperature = 0.3
max_tokens = 1200

[agents.analyzer]
role = "analysis_specialist"
description = "Provides deep insights and implications"
system_prompt = "You are an analysis specialist who identifies patterns and provides actionable insights."
capabilities = ["pattern_recognition", "trend_analysis", "insight_generation"]
enabled = true

[orchestration]
mode = "collaborative"
timeout_seconds = 60
collaborative_agents = ["researcher", "analyzer"]

[providers.openai]
api_key = "test-key"
model = "gpt-4"
`

	validConfigPath := filepath.Join(tempDir, "valid_config.toml")
	err = os.WriteFile(validConfigPath, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write valid config: %v", err)
	}

	// Load and validate the valid configuration
	config, err := LoadConfig(validConfigPath)
	if err != nil {
		t.Fatalf("Failed to load valid config: %v", err)
	}

	// Verify the configuration was loaded correctly
	if config.LLM.Provider != "openai" {
		t.Errorf("Expected LLM provider 'openai', got '%s'", config.LLM.Provider)
	}

	if len(config.Agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(config.Agents))
	}

	// Test agent resolution
	researcherConfig, err := config.ResolveAgentConfig("researcher")
	if err != nil {
		t.Fatalf("Failed to resolve researcher config: %v", err)
	}

	if researcherConfig.LLMConfig.Temperature != 0.3 {
		t.Errorf("Expected researcher temperature 0.3, got %f", researcherConfig.LLMConfig.Temperature)
	}

	analyzerConfig, err := config.ResolveAgentConfig("analyzer")
	if err != nil {
		t.Fatalf("Failed to resolve analyzer config: %v", err)
	}

	if analyzerConfig.LLMConfig.Temperature != 0.7 {
		t.Errorf("Expected analyzer temperature 0.7 (global), got %f", analyzerConfig.LLMConfig.Temperature)
	}

	// Test case 2: Invalid configuration (should still load but with warnings)
	invalidConfig := `
[agent_flow]
name = "test-project"
version = "1.0.0"
provider = "openai"

[llm]
provider = "unknown_provider"
temperature = 2.5
max_tokens = -100

[agents.invalid_agent]
role = "Invalid-Role"
system_prompt = "Hi"
capabilities = ["unknown_capability", "unknown_capability"]
enabled = true
timeout_seconds = -10

[agents.invalid_agent.llm]
temperature = 3.0
max_tokens = 0

[orchestration]
mode = "collaborative"
timeout_seconds = 60
collaborative_agents = ["invalid_agent", "nonexistent_agent"]

[providers.openai]
api_key = "test-key"
`

	invalidConfigPath := filepath.Join(tempDir, "invalid_config.toml")
	err = os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Load the invalid configuration (should succeed but with warnings)
	invalidConfigObj, err := LoadConfig(invalidConfigPath)
	if err != nil {
		t.Fatalf("Failed to load invalid config: %v", err)
	}

	// The configuration should still be loaded despite validation errors
	if invalidConfigObj == nil {
		t.Error("Expected configuration to be loaded despite validation errors")
	}

	// Test manual validation
	validator := NewDefaultConfigValidator()
	validationErrors := validator.ValidateConfig(invalidConfigObj)

	if len(validationErrors) == 0 {
		t.Error("Expected validation errors for invalid configuration")
	}

	// Check that we have the expected types of errors
	foundLLMError := false
	foundAgentError := false
	foundOrchestrationError := false

	for _, validationError := range validationErrors {
		if validationError.Field == "llm.provider" {
			foundLLMError = true
		}
		if validationError.Field == "agents.invalid_agent.role" {
			foundAgentError = true
		}
		if validationError.Field == "orchestration.collaborative_agents" {
			foundOrchestrationError = true
		}
	}

	if !foundLLMError {
		t.Error("Expected to find LLM validation error")
	}
	if !foundAgentError {
		t.Error("Expected to find agent validation error")
	}
	if !foundOrchestrationError {
		t.Error("Expected to find orchestration validation error")
	}
}

func TestConfigDefaultsWithValidation(t *testing.T) {
	// Create a minimal configuration that relies on defaults
	tempDir, err := os.MkdirTemp("", "agentflow_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	minimalConfig := `
[agent_flow]
name = "minimal-project"
version = "1.0.0"
provider = "openai"

[agents.simple_agent]
role = "simple_agent"
description = "A simple test agent"
system_prompt = "You are a helpful assistant for testing purposes."
capabilities = ["information_gathering"]
enabled = true

[providers.openai]
api_key = "test-key"
`

	configPath := filepath.Join(tempDir, "minimal_config.toml")
	err = os.WriteFile(configPath, []byte(minimalConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write minimal config: %v", err)
	}

	// Load the configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load minimal config: %v", err)
	}

	// Verify defaults were applied
	if config.LLM.Temperature == 0 {
		t.Error("Expected default temperature to be set")
	}
	if config.LLM.MaxTokens == 0 {
		t.Error("Expected default max_tokens to be set")
	}

	// Verify agent defaults
	agent := config.Agents["simple_agent"]
	if agent.Timeout == 0 {
		t.Error("Expected default timeout to be set for agent")
	}

	// Validate the configuration with defaults
	validator := NewDefaultConfigValidator()
	validationErrors := validator.ValidateConfig(config)

	// Should have minimal or no validation errors
	if len(validationErrors) > 1 {
		t.Errorf("Expected minimal validation errors, got %d: %v", len(validationErrors), validationErrors)
	}
}