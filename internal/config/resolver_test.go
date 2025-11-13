package config

import (
	"os"
	"testing"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

func TestConfigResolverBasicResolution(t *testing.T) {
	config := &core.Config{
		LLM: core.AgentLLMConfig{
			Provider:       "openai",
			Model:          "gpt-4",
			Temperature:    0.7,
			MaxTokens:      800,
			TimeoutSeconds: 30,
		},
		Agents: map[string]core.AgentConfig{
			"test_agent": {
				Role:         "test_role",
				Description:  "Test agent",
				SystemPrompt: "You are a test agent",
				Capabilities: []string{"testing"},
				Enabled:      true,
				Timeout:      45,
			},
		},
	}

	resolver := NewConfigResolver(config)
	resolved, err := resolver.ResolveAgentConfigWithEnv("test_agent")
	if err != nil {
		t.Fatalf("Failed to resolve agent config: %v", err)
	}

	// Verify basic resolution
	if resolved.Name != "test_agent" {
		t.Errorf("Expected name 'test_agent', got '%s'", resolved.Name)
	}
	if resolved.Role != "test_role" {
		t.Errorf("Expected role 'test_role', got '%s'", resolved.Role)
	}
	if resolved.LLMConfig.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", resolved.LLMConfig.Provider)
	}
	if resolved.Timeout != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", resolved.Timeout)
	}
}

func TestConfigResolverEnvironmentOverrides(t *testing.T) {
	// Set up test environment variables
	testEnvVars := map[string]string{
		"AGENTFLOW_AGENT_TEST_AGENT_ROLE":            "overridden_role",
		"AGENTFLOW_AGENT_TEST_AGENT_DESCRIPTION":     "Overridden description",
		"AGENTFLOW_AGENT_TEST_AGENT_SYSTEM_PROMPT":   "You are an overridden agent",
		"AGENTFLOW_AGENT_TEST_AGENT_CAPABILITIES":    "cap1,cap2,cap3",
		"AGENTFLOW_AGENT_TEST_AGENT_ENABLED":         "false",
		"AGENTFLOW_AGENT_TEST_AGENT_TIMEOUT_SECONDS": "60",
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

	config := &core.Config{
		LLM: core.AgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   800,
		},
		Agents: map[string]core.AgentConfig{
			"test_agent": {
				Role:         "original_role",
				Description:  "Original description",
				SystemPrompt: "You are an original agent",
				Capabilities: []string{"original_cap"},
				Enabled:      true,
				Timeout:      30,
			},
		},
	}

	resolver := NewConfigResolver(config)
	resolved, err := resolver.ResolveAgentConfigWithEnv("test_agent")
	if err != nil {
		t.Fatalf("Failed to resolve agent config: %v", err)
	}

	// Verify environment overrides were applied
	if resolved.Role != "overridden_role" {
		t.Errorf("Expected role 'overridden_role', got '%s'", resolved.Role)
	}
	if resolved.Description != "Overridden description" {
		t.Errorf("Expected description 'Overridden description', got '%s'", resolved.Description)
	}
	if resolved.SystemPrompt != "You are an overridden agent" {
		t.Errorf("Expected system prompt 'You are an overridden agent', got '%s'", resolved.SystemPrompt)
	}
	if len(resolved.Capabilities) != 3 || resolved.Capabilities[0] != "cap1" {
		t.Errorf("Expected capabilities [cap1, cap2, cap3], got %v", resolved.Capabilities)
	}
	if resolved.Enabled != false {
		t.Errorf("Expected enabled false, got %v", resolved.Enabled)
	}
	if resolved.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", resolved.Timeout)
	}
}

func TestConfigResolverLLMEnvironmentOverrides(t *testing.T) {
	// Set up LLM environment variables
	testEnvVars := map[string]string{
		"AGENTFLOW_LLM_PROVIDER":                         "azure",
		"AGENTFLOW_LLM_MODEL":                            "gpt-3.5-turbo",
		"AGENTFLOW_LLM_TEMPERATURE":                      "0.5",
		"AGENTFLOW_LLM_MAX_TOKENS":                       "1000",
		"AGENTFLOW_AGENT_TEST_AGENT_LLM_PROVIDER":        "ollama",
		"AGENTFLOW_AGENT_TEST_AGENT_LLM_MODEL":           "llama2",
		"AGENTFLOW_AGENT_TEST_AGENT_LLM_TEMPERATURE":     "0.3",
		"AGENTFLOW_AGENT_TEST_AGENT_LLM_MAX_TOKENS":      "1200",
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

	config := &core.Config{
		LLM: core.AgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   800,
		},
		Agents: map[string]core.AgentConfig{
			"test_agent": {
				Role:         "test_role",
				SystemPrompt: "You are a test agent",
				Capabilities: []string{"testing"},
				Enabled:      true,
			},
			"other_agent": {
				Role:         "other_role",
				SystemPrompt: "You are another agent",
				Capabilities: []string{"testing"},
				Enabled:      true,
			},
		},
	}

	resolver := NewConfigResolver(config)

	// Test agent with specific overrides
	resolved, err := resolver.ResolveAgentConfigWithEnv("test_agent")
	if err != nil {
		t.Fatalf("Failed to resolve test_agent config: %v", err)
	}

	// Should use agent-specific overrides (highest priority)
	if resolved.LLMConfig.Provider != "ollama" {
		t.Errorf("Expected provider 'ollama', got '%s'", resolved.LLMConfig.Provider)
	}
	if resolved.LLMConfig.Model != "llama2" {
		t.Errorf("Expected model 'llama2', got '%s'", resolved.LLMConfig.Model)
	}
	if resolved.LLMConfig.Temperature != 0.3 {
		t.Errorf("Expected temperature 0.3, got %f", resolved.LLMConfig.Temperature)
	}
	if resolved.LLMConfig.MaxTokens != 1200 {
		t.Errorf("Expected max tokens 1200, got %d", resolved.LLMConfig.MaxTokens)
	}

	// Test agent without specific overrides (should use global overrides)
	otherResolved, err := resolver.ResolveAgentConfigWithEnv("other_agent")
	if err != nil {
		t.Fatalf("Failed to resolve other_agent config: %v", err)
	}

	// Should use global overrides
	if otherResolved.LLMConfig.Provider != "azure" {
		t.Errorf("Expected provider 'azure', got '%s'", otherResolved.LLMConfig.Provider)
	}
	if otherResolved.LLMConfig.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected model 'gpt-3.5-turbo', got '%s'", otherResolved.LLMConfig.Model)
	}
	if otherResolved.LLMConfig.Temperature != 0.5 {
		t.Errorf("Expected temperature 0.5, got %f", otherResolved.LLMConfig.Temperature)
	}
	if otherResolved.LLMConfig.MaxTokens != 1000 {
		t.Errorf("Expected max tokens 1000, got %d", otherResolved.LLMConfig.MaxTokens)
	}
}

func TestConfigResolverApplyEnvironmentOverrides(t *testing.T) {
	// Set up environment variables
	testEnvVars := map[string]string{
		"AGENTFLOW_LLM_PROVIDER":              "azure",
		"AGENTFLOW_LLM_TEMPERATURE":           "0.5",
		"AGENTFLOW_AGENT_TEST_AGENT_ROLE":     "env_role",
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

	config := &core.Config{
		LLM: core.AgentLLMConfig{
			Provider:    "openai",
			Temperature: 0.7,
		},
		Agents: map[string]core.AgentConfig{
			"test_agent": {
				Role:         "original_role",
				SystemPrompt: "You are a test agent",
				Capabilities: []string{"testing"},
				Enabled:      true,
			},
		},
	}

	resolver := NewConfigResolver(config)
	err := resolver.ApplyEnvironmentOverrides()
	if err != nil {
		t.Fatalf("Failed to apply environment overrides: %v", err)
	}

	// Verify global overrides were applied to the config
	if config.LLM.Provider != "azure" {
		t.Errorf("Expected global LLM provider 'azure', got '%s'", config.LLM.Provider)
	}
	if config.LLM.Temperature != 0.5 {
		t.Errorf("Expected global LLM temperature 0.5, got %f", config.LLM.Temperature)
	}

	// Verify agent overrides were applied to the config
	if config.Agents["test_agent"].Role != "env_role" {
		t.Errorf("Expected agent role 'env_role', got '%s'", config.Agents["test_agent"].Role)
	}
}

func TestConfigResolverResolveAllAgents(t *testing.T) {
	config := &core.Config{
		LLM: core.AgentLLMConfig{
			Provider:    "openai",
			Temperature: 0.7,
		},
		Agents: map[string]core.AgentConfig{
			"agent1": {
				Role:         "role1",
				SystemPrompt: "You are agent 1",
				Capabilities: []string{"cap1"},
				Enabled:      true,
			},
			"agent2": {
				Role:         "role2",
				SystemPrompt: "You are agent 2",
				Capabilities: []string{"cap2"},
				Enabled:      true,
			},
		},
	}

	resolver := NewConfigResolver(config)
	allResolved, err := resolver.ResolveAllAgents()
	if err != nil {
		t.Fatalf("Failed to resolve all agents: %v", err)
	}

	if len(allResolved) != 2 {
		t.Errorf("Expected 2 resolved agents, got %d", len(allResolved))
	}

	if _, exists := allResolved["agent1"]; !exists {
		t.Error("Expected agent1 to be resolved")
	}
	if _, exists := allResolved["agent2"]; !exists {
		t.Error("Expected agent2 to be resolved")
	}

	if allResolved["agent1"].Role != "role1" {
		t.Errorf("Expected agent1 role 'role1', got '%s'", allResolved["agent1"].Role)
	}
	if allResolved["agent2"].Role != "role2" {
		t.Errorf("Expected agent2 role 'role2', got '%s'", allResolved["agent2"].Role)
	}
}

func TestConfigResolverInvalidEnvironmentValues(t *testing.T) {
	// Set up invalid environment variables
	testEnvVars := map[string]string{
		"AGENTFLOW_AGENT_TEST_AGENT_ENABLED":         "invalid_bool",
		"AGENTFLOW_AGENT_TEST_AGENT_TIMEOUT_SECONDS": "invalid_int",
		"AGENTFLOW_LLM_TEMPERATURE":                  "invalid_float",
		"AGENTFLOW_LLM_MAX_TOKENS":                   "invalid_int",
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

	config := &core.Config{
		LLM: core.AgentLLMConfig{
			Provider:    "openai",
			Temperature: 0.7,
			MaxTokens:   800,
		},
		Agents: map[string]core.AgentConfig{
			"test_agent": {
				Role:         "test_role",
				SystemPrompt: "You are a test agent",
				Capabilities: []string{"testing"},
				Enabled:      true,
				Timeout:      30,
			},
		},
	}

	resolver := NewConfigResolver(config)

	// Should handle invalid values gracefully
	resolved, err := resolver.ResolveAgentConfigWithEnv("test_agent")
	if err != nil {
		t.Fatalf("Failed to resolve agent config with invalid env vars: %v", err)
	}

	// Should keep original values when env vars are invalid
	if resolved.Enabled != true {
		t.Errorf("Expected enabled true (original), got %v", resolved.Enabled)
	}
	if resolved.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s (original), got %v", resolved.Timeout)
	}

	// Test global overrides with invalid values
	err = resolver.ApplyEnvironmentOverrides()
	if err == nil {
		t.Error("Expected error when applying invalid global environment overrides")
	}
}

func TestConfigResolverNonExistentAgent(t *testing.T) {
	config := &core.Config{
		Agents: map[string]core.AgentConfig{
			"existing_agent": {
				Role:         "test_role",
				SystemPrompt: "You are a test agent",
				Capabilities: []string{"testing"},
				Enabled:      true,
			},
		},
	}

	resolver := NewConfigResolver(config)
	_, err := resolver.ResolveAgentConfigWithEnv("nonexistent_agent")
	if err == nil {
		t.Error("Expected error for nonexistent agent")
	}
}

func TestConfigResolverValidateResolvedConfig(t *testing.T) {
	config := &core.Config{
		LLM: core.AgentLLMConfig{
			Provider:    "openai",
			Temperature: 0.7,
		},
		Agents: map[string]core.AgentConfig{
			"valid_agent": {
				Role:         "test_role",
				SystemPrompt: "You are a test agent",
				Capabilities: []string{"information_gathering"},
				Enabled:      true,
			},
			"invalid_agent": {
				Role:         "", // Invalid: missing role
				SystemPrompt: "You are a test agent",
				Capabilities: []string{"unknown_capability"}, // Invalid: unknown capability
				Enabled:      true,
			},
		},
	}

	resolver := NewConfigResolver(config)
	validationErrors := resolver.ValidateResolvedConfig()

	// Should have validation errors for the invalid agent
	if len(validationErrors) == 0 {
		t.Error("Expected validation errors for invalid configuration")
	}

	// Check for specific errors
	foundRoleError := false
	foundCapabilityError := false
	for _, err := range validationErrors {
		if err.Field == "agents.invalid_agent.role" {
			foundRoleError = true
		}
		if err.Field == "agents.invalid_agent.capabilities" {
			foundCapabilityError = true
		}
	}

	if !foundRoleError {
		t.Error("Expected to find role validation error")
	}
	if !foundCapabilityError {
		t.Error("Expected to find capability validation error")
	}
}
