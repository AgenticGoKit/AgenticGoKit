package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kunalkushwaha/agenticgokit/core"
)

func TestConfigResolverIntegrationWithLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "agentflow_resolver_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test configuration file
	configContent := `
[agent_flow]
name = "resolver-test"
version = "1.0.0"
provider = "openai"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 800

[agents.test_agent]
role = "test_role"
description = "Test agent for resolver"
system_prompt = "You are a test agent"
capabilities = ["testing"]
enabled = true
timeout_seconds = 30

[providers.openai]
api_key = "test-key"
`

	configPath := filepath.Join(tempDir, "agentflow.toml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set up environment variables to override configuration
	testEnvVars := map[string]string{
		"AGENTFLOW_LLM_PROVIDER":                       "azure",
		"AGENTFLOW_LLM_TEMPERATURE":                    "0.5",
		"AGENTFLOW_AGENT_TEST_AGENT_ROLE":              "overridden_role",
		"AGENTFLOW_AGENT_TEST_AGENT_SYSTEM_PROMPT":     "You are an overridden test agent",
		"AGENTFLOW_AGENT_TEST_AGENT_LLM_MODEL":         "gpt-3.5-turbo",
		"AGENTFLOW_AGENT_TEST_AGENT_LLM_MAX_TOKENS":    "1200",
	}

	// Set environment variables
	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}
	defer func() {
		// Clean up environment variables
		for key := range testEnvVars {
			os.Unsetenv(key)
		}
	}()

	// Load configuration (should apply environment overrides automatically)
	config, err := core.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify that global environment overrides were applied during LoadConfig
	if config.LLM.Provider != "azure" {
		t.Errorf("Expected global LLM provider 'azure', got '%s'", config.LLM.Provider)
	}
	if config.LLM.Temperature != 0.5 {
		t.Errorf("Expected global LLM temperature 0.5, got %f", config.LLM.Temperature)
	}

	// Verify that agent environment overrides were applied during LoadConfig
	if config.Agents["test_agent"].Role != "overridden_role" {
		t.Errorf("Expected agent role 'overridden_role', got '%s'", config.Agents["test_agent"].Role)
	}
	if config.Agents["test_agent"].SystemPrompt != "You are an overridden test agent" {
		t.Errorf("Expected overridden system prompt, got '%s'", config.Agents["test_agent"].SystemPrompt)
	}

	// Test agent resolution with the loaded configuration
	resolved, err := config.ResolveAgentConfig("test_agent")
	if err != nil {
		t.Fatalf("Failed to resolve agent config: %v", err)
	}

	// Verify that the resolved configuration includes all overrides
	if resolved.Role != "overridden_role" {
		t.Errorf("Expected resolved role 'overridden_role', got '%s'", resolved.Role)
	}
	if resolved.SystemPrompt != "You are an overridden test agent" {
		t.Errorf("Expected resolved system prompt to be overridden, got '%s'", resolved.SystemPrompt)
	}
	if resolved.LLMConfig.Provider != "azure" {
		t.Errorf("Expected resolved LLM provider 'azure', got '%s'", resolved.LLMConfig.Provider)
	}
	if resolved.LLMConfig.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected resolved LLM model 'gpt-3.5-turbo', got '%s'", resolved.LLMConfig.Model)
	}
	if resolved.LLMConfig.MaxTokens != 1200 {
		t.Errorf("Expected resolved LLM max tokens 1200, got %d", resolved.LLMConfig.MaxTokens)
	}
	if resolved.LLMConfig.Temperature != 0.5 {
		t.Errorf("Expected resolved LLM temperature 0.5, got %f", resolved.LLMConfig.Temperature)
	}
}

func TestConfigResolverPriorityOrder(t *testing.T) {
	// Test the priority order: Agent-specific env > Global env > Agent config > Global config
	tempDir, err := os.MkdirTemp("", "agentflow_priority_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create configuration with all levels defined
	configContent := `
[agent_flow]
name = "priority-test"
version = "1.0.0"
provider = "openai"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 800

[agents.priority_agent]
role = "config_role"
system_prompt = "Config system prompt"
capabilities = ["testing"]
enabled = true

[agents.priority_agent.llm]
provider = "azure"
model = "gpt-3.5-turbo"
temperature = 0.5
max_tokens = 1000

[providers.openai]
api_key = "test-key"
`

	configPath := filepath.Join(tempDir, "agentflow.toml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variables at different levels
	testEnvVars := map[string]string{
		// Global LLM overrides
		"AGENTFLOW_LLM_PROVIDER":    "ollama",
		"AGENTFLOW_LLM_MODEL":       "llama2",
		"AGENTFLOW_LLM_TEMPERATURE": "0.3",
		"AGENTFLOW_LLM_MAX_TOKENS":  "1500",
		
		// Agent-specific LLM overrides (should have highest priority)
		"AGENTFLOW_AGENT_PRIORITY_AGENT_LLM_PROVIDER":    "anthropic",
		"AGENTFLOW_AGENT_PRIORITY_AGENT_LLM_MODEL":       "claude-3",
		"AGENTFLOW_AGENT_PRIORITY_AGENT_LLM_TEMPERATURE": "0.1",
		"AGENTFLOW_AGENT_PRIORITY_AGENT_LLM_MAX_TOKENS":  "2000",
	}

	// Set environment variables
	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}
	defer func() {
		// Clean up environment variables
		for key := range testEnvVars {
			os.Unsetenv(key)
		}
	}()

	// Load configuration
	config, err := core.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Resolve agent configuration
	resolved, err := config.ResolveAgentConfig("priority_agent")
	if err != nil {
		t.Fatalf("Failed to resolve agent config: %v", err)
	}

	// Verify that agent-specific environment variables have the highest priority
	if resolved.LLMConfig.Provider != "anthropic" {
		t.Errorf("Expected provider 'anthropic' (agent-specific env), got '%s'", resolved.LLMConfig.Provider)
	}
	if resolved.LLMConfig.Model != "claude-3" {
		t.Errorf("Expected model 'claude-3' (agent-specific env), got '%s'", resolved.LLMConfig.Model)
	}
	if resolved.LLMConfig.Temperature != 0.1 {
		t.Errorf("Expected temperature 0.1 (agent-specific env), got %f", resolved.LLMConfig.Temperature)
	}
	if resolved.LLMConfig.MaxTokens != 2000 {
		t.Errorf("Expected max tokens 2000 (agent-specific env), got %d", resolved.LLMConfig.MaxTokens)
	}
}

func TestConfigResolverWithoutEnvironmentVariables(t *testing.T) {
	// Test that configuration works normally without any environment variables
	tempDir, err := os.MkdirTemp("", "agentflow_no_env_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configContent := `
[agent_flow]
name = "no-env-test"
version = "1.0.0"
provider = "openai"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 800

[agents.normal_agent]
role = "normal_role"
description = "Normal agent without env overrides"
system_prompt = "You are a normal agent"
capabilities = ["testing"]
enabled = true

[providers.openai]
api_key = "test-key"
`

	configPath := filepath.Join(tempDir, "agentflow.toml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load configuration without any environment variables
	config, err := core.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Resolve agent configuration
	resolved, err := config.ResolveAgentConfig("normal_agent")
	if err != nil {
		t.Fatalf("Failed to resolve agent config: %v", err)
	}

	// Verify that original configuration values are preserved
	if resolved.Role != "normal_role" {
		t.Errorf("Expected role 'normal_role', got '%s'", resolved.Role)
	}
	if resolved.Description != "Normal agent without env overrides" {
		t.Errorf("Expected original description, got '%s'", resolved.Description)
	}
	if resolved.LLMConfig.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", resolved.LLMConfig.Provider)
	}
	if resolved.LLMConfig.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", resolved.LLMConfig.Model)
	}
	if resolved.LLMConfig.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", resolved.LLMConfig.Temperature)
	}
}