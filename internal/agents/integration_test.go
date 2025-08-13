package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigurableAgentFactoryIntegration(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "agentflow_factory_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a comprehensive test configuration file
	configContent := `
[agent_flow]
name = "factory-integration-test"
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
system_prompt = "You are a research specialist who excels at finding and analyzing information from various sources. Always provide detailed, accurate, and well-sourced information."
capabilities = ["information_gathering", "fact_checking", "source_identification"]
enabled = true
timeout_seconds = 45

[agents.researcher.llm]
temperature = 0.3
max_tokens = 1200

[agents.analyzer]
role = "analysis_specialist"
description = "Provides deep insights and implications"
system_prompt = "You are an analysis specialist who identifies patterns, trends, and provides actionable insights from data and information."
capabilities = ["pattern_recognition", "trend_analysis", "insight_generation"]
enabled = true

[agents.writer]
role = "content_writer"
description = "Creates well-structured written content"
system_prompt = "You are a skilled content writer who creates clear, engaging, and well-structured written content."
capabilities = ["content_creation", "editing", "summarization"]
enabled = true
timeout_seconds = 60

[agents.writer.llm]
temperature = 0.8
max_tokens = 1500

[agents.disabled_agent]
role = "disabled_role"
description = "This agent is disabled for testing"
system_prompt = "You are disabled"
capabilities = ["disabled_capability"]
enabled = false

[orchestration]
mode = "collaborative"
timeout_seconds = 120
collaborative_agents = ["researcher", "analyzer", "writer"]

[providers.openai]
api_key = "test-key-for-integration"
model = "gpt-4"
max_tokens = 1000
temperature = 0.7
`

	configPath := filepath.Join(tempDir, "agentflow.toml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set up environment variables to test overrides
	testEnvVars := map[string]string{
		"AGENTFLOW_LLM_TEMPERATURE":                     "0.5",
		"AGENTFLOW_AGENT_RESEARCHER_LLM_MAX_TOKENS":     "1000",
		"AGENTFLOW_AGENT_WRITER_SYSTEM_PROMPT":          "You are an enhanced content writer with advanced capabilities.",
		"AGENTFLOW_AGENT_ANALYZER_CAPABILITIES":         "advanced_analysis,predictive_modeling,data_visualization",
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
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create configurable agent factory
	factory := NewConfigurableAgentFactory(config)

	// Test 1: Create individual agents
	t.Run("CreateIndividualAgents", func(t *testing.T) {
		// Create researcher agent
		researcher, err := factory.CreateAgentFromConfig("researcher", config)
		if err != nil {
			t.Fatalf("Failed to create researcher agent: %v", err)
		}

		configuredResearcher, ok := researcher.(*ConfiguredAgent)
		if !ok {
			t.Fatal("Expected researcher to be a ConfiguredAgent")
		}

		// Verify researcher configuration
		if configuredResearcher.GetRole() != "research_specialist" {
			t.Errorf("Expected researcher role 'research_specialist', got '%s'", configuredResearcher.GetRole())
		}

		if configuredResearcher.GetTimeout() != 45*time.Second {
			t.Errorf("Expected researcher timeout 45s, got %v", configuredResearcher.GetTimeout())
		}

		// Verify LLM configuration with environment override
		llmConfig := configuredResearcher.GetLLMConfig()
		if llmConfig.MaxTokens != 1000 { // Should be overridden by env var
			t.Errorf("Expected researcher max tokens 1000 (env override), got %d", llmConfig.MaxTokens)
		}

		if llmConfig.Temperature != 0.3 { // Should use agent-specific config
			t.Errorf("Expected researcher temperature 0.3, got %f", llmConfig.Temperature)
		}

		// Create writer agent
		writer, err := factory.CreateAgentFromConfig("writer", config)
		if err != nil {
			t.Fatalf("Failed to create writer agent: %v", err)
		}

		configuredWriter, ok := writer.(*ConfiguredAgent)
		if !ok {
			t.Fatal("Expected writer to be a ConfiguredAgent")
		}

		// Verify writer configuration with environment override
		if configuredWriter.GetSystemPrompt() != "You are an enhanced content writer with advanced capabilities." {
			t.Errorf("Expected overridden system prompt, got '%s'", configuredWriter.GetSystemPrompt())
		}

		// Create analyzer agent
		analyzer, err := factory.CreateAgentFromConfig("analyzer", config)
		if err != nil {
			t.Fatalf("Failed to create analyzer agent: %v", err)
		}

		configuredAnalyzer, ok := analyzer.(*ConfiguredAgent)
		if !ok {
			t.Fatal("Expected analyzer to be a ConfiguredAgent")
		}

		// Verify analyzer capabilities with environment override
		capabilities := configuredAnalyzer.GetCapabilities()
		expectedCaps := []string{"advanced_analysis", "predictive_modeling", "data_visualization"}
		if len(capabilities) != 3 {
			t.Errorf("Expected 3 capabilities, got %d: %v", len(capabilities), capabilities)
		}
		for i, expected := range expectedCaps {
			if i >= len(capabilities) || capabilities[i] != expected {
				t.Errorf("Expected capability '%s' at index %d, got '%s'", expected, i, capabilities[i])
			}
		}
	})

	// Test 2: Create all enabled agents
	t.Run("CreateAllEnabledAgents", func(t *testing.T) {
		agents, err := factory.CreateAllEnabledAgents()
		if err != nil {
			t.Fatalf("Failed to create all enabled agents: %v", err)
		}

		// Should have 3 enabled agents (researcher, analyzer, writer)
		if len(agents) != 3 {
			t.Errorf("Expected 3 enabled agents, got %d", len(agents))
		}

		// Verify all expected agents are present
		expectedAgents := []string{"researcher", "analyzer", "writer"}
		for _, agentName := range expectedAgents {
			if _, exists := agents[agentName]; !exists {
				t.Errorf("Expected agent '%s' to be created", agentName)
			}
		}

		// Verify disabled agent is not present
		if _, exists := agents["disabled_agent"]; exists {
			t.Error("Expected disabled_agent to NOT be created")
		}
	})

	// Test 3: Validate agent configurations
	t.Run("ValidateAgentConfigurations", func(t *testing.T) {
		// Valid agents should pass validation
		validAgents := []string{"researcher", "analyzer", "writer"}
		for _, agentName := range validAgents {
			err := factory.ValidateAgentConfiguration(agentName)
			if err != nil {
				t.Errorf("Expected agent '%s' to be valid, got error: %v", agentName, err)
			}
		}

		// Disabled agent should fail validation
		err := factory.ValidateAgentConfiguration("disabled_agent")
		if err == nil {
			t.Error("Expected disabled_agent to fail validation")
		}

		// Non-existent agent should fail validation
		err = factory.ValidateAgentConfiguration("nonexistent_agent")
		if err == nil {
			t.Error("Expected nonexistent_agent to fail validation")
		}
	})

	// Test 4: Test agent execution with configuration
	t.Run("AgentExecution", func(t *testing.T) {
		// Create a researcher agent
		researcher, err := factory.CreateAgentFromConfig("researcher", config)
		if err != nil {
			t.Fatalf("Failed to create researcher agent: %v", err)
		}

		// Test running the agent
		ctx := context.Background()
		inputState := NewState()
		inputState.Set("task", "research_task")
		inputState.Set("query", "test query")

		outputState, err := researcher.Run(ctx, inputState)
		if err != nil {
			t.Fatalf("Failed to run researcher agent: %v", err)
		}

		// Verify configuration metadata was added to state
		if agentName, exists := outputState.Get("agent_name"); !exists || agentName != "researcher" {
			t.Errorf("Expected agent_name 'researcher', got '%v'", agentName)
		}

		if role, exists := outputState.Get("agent_role"); !exists || role != "research_specialist" {
			t.Errorf("Expected agent_role 'research_specialist', got '%v'", role)
		}

		if executedBy, exists := outputState.Get("executed_by"); !exists || executedBy != "researcher" {
			t.Errorf("Expected executed_by 'researcher', got '%v'", executedBy)
		}

		// Verify capabilities were added
		if capabilities, exists := outputState.Get("agent_capabilities"); !exists {
			t.Error("Expected agent_capabilities to be set")
		} else {
			caps, ok := capabilities.([]string)
			if !ok {
				t.Errorf("Expected capabilities to be []string, got %T", capabilities)
			} else if len(caps) == 0 {
				t.Error("Expected capabilities to be non-empty")
			}
		}
	})

	// Test 5: Test helper methods
	t.Run("HelperMethods", func(t *testing.T) {
		// Test GetAgentCapabilities
		capabilities := factory.GetAgentCapabilities("researcher")
		if len(capabilities) != 3 {
			t.Errorf("Expected 3 capabilities for researcher, got %d", len(capabilities))
		}

		// Test IsAgentEnabled
		if !factory.IsAgentEnabled("researcher") {
			t.Error("Expected researcher to be enabled")
		}
		if factory.IsAgentEnabled("disabled_agent") {
			t.Error("Expected disabled_agent to be disabled")
		}
		if factory.IsAgentEnabled("nonexistent_agent") {
			t.Error("Expected nonexistent_agent to be disabled")
		}
	})
}

func TestConfigurableAgentFactoryWithoutEnvironmentOverrides(t *testing.T) {
	// Test that the factory works correctly without any environment variables
	tempDir, err := os.MkdirTemp("", "agentflow_no_env_factory_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configContent := `
[agent_flow]
name = "no-env-factory-test"
version = "1.0.0"
provider = "openai"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 800

[agents.simple_agent]
role = "simple_role"
description = "Simple agent without env overrides"
system_prompt = "You are a simple agent"
capabilities = ["simple_capability"]
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

	// Load configuration without any environment variables
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create factory and agent
	factory := NewConfigurableAgentFactory(config)
	agent, err := factory.CreateAgentFromConfig("simple_agent", config)
	if err != nil {
		t.Fatalf("Failed to create simple agent: %v", err)
	}

	// Verify agent was created with original configuration values
	configuredAgent, ok := agent.(*ConfiguredAgent)
	if !ok {
		t.Fatal("Expected agent to be a ConfiguredAgent")
	}

	if configuredAgent.GetRole() != "simple_role" {
		t.Errorf("Expected role 'simple_role', got '%s'", configuredAgent.GetRole())
	}

	if configuredAgent.GetDescription() != "Simple agent without env overrides" {
		t.Errorf("Expected original description, got '%s'", configuredAgent.GetDescription())
	}

	if configuredAgent.GetSystemPrompt() != "You are a simple agent" {
		t.Errorf("Expected original system prompt, got '%s'", configuredAgent.GetSystemPrompt())
	}

	if configuredAgent.GetTimeout() != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", configuredAgent.GetTimeout())
	}

	// Verify LLM configuration uses global defaults
	llmConfig := configuredAgent.GetLLMConfig()
	if llmConfig.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", llmConfig.Provider)
	}
	if llmConfig.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", llmConfig.Model)
	}
	if llmConfig.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", llmConfig.Temperature)
	}
}