package agents

import "github.com/kunalkushwaha/agenticgokit/core"

import (
	"context"
	"testing"
	"time"
)

func TestConfigurableAgentFactory_CreateAgent(t *testing.T) {
	// Create test configuration
	config := &Config{
		LLM: AgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   800,
		},
		Providers: map[string]map[string]interface{}{
			"openai": {
				"api_key": "test-key",
				"model":   "gpt-4",
			},
		},
	}

	// Create resolved agent configuration
	resolvedConfig := &ResolvedAgentConfig{
		Name:         "test_agent",
		Role:         "test_role",
		Description:  "Test agent for factory testing",
		SystemPrompt: "You are a test agent",
		Capabilities: []string{"testing", "validation"},
		Enabled:      true,
		LLMConfig: &ResolvedLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   800,
			Timeout:     30 * time.Second,
		},
		Timeout: 45 * time.Second,
	}

	factory := NewConfigurableAgentFactory(config)

	// Create a mock LLM provider
	mockProvider := &MockModelProvider{}

	// Test creating agent
	agent, err := factory.CreateAgent("test_agent", resolvedConfig, mockProvider)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Verify agent was created
	if agent == nil {
		t.Fatal("Expected agent to be created, got nil")
	}

	if agent.Name() != "test_agent" {
		t.Errorf("Expected agent name 'test_agent', got '%s'", agent.Name())
	}

	// Test that the agent is a ConfiguredAgent
	configuredAgent, ok := agent.(*ConfiguredAgent)
	if !ok {
		t.Fatal("Expected agent to be a ConfiguredAgent")
	}

	if configuredAgent.GetRole() != "test_role" {
		t.Errorf("Expected role 'test_role', got '%s'", configuredAgent.GetRole())
	}

	if len(configuredAgent.GetCapabilities()) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(configuredAgent.GetCapabilities()))
	}

	if configuredAgent.GetTimeout() != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", configuredAgent.GetTimeout())
	}
}

func TestConfigurableAgentFactory_CreateAgentFromConfig(t *testing.T) {
	// Create test configuration
	config := &Config{
		LLM: AgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   800,
		},
		Agents: map[string]AgentConfig{
			"config_agent": {
				Role:         "config_role",
				Description:  "Agent created from config",
				SystemPrompt: "You are a configuration-driven agent",
				Capabilities: []string{"configuration", "testing"},
				Enabled:      true,
				Timeout:      60,
			},
		},
		Providers: map[string]map[string]interface{}{
			"openai": {
				"api_key": "test-key",
				"model":   "gpt-4",
			},
		},
	}

	factory := NewConfigurableAgentFactory(config)

	// Test creating agent from config
	agent, err := factory.CreateAgentFromConfig("config_agent", config)
	if err != nil {
		t.Fatalf("Failed to create agent from config: %v", err)
	}

	// Verify agent was created correctly
	if agent == nil {
		t.Fatal("Expected agent to be created, got nil")
	}

	configuredAgent, ok := agent.(*ConfiguredAgent)
	if !ok {
		t.Fatal("Expected agent to be a ConfiguredAgent")
	}

	if configuredAgent.GetRole() != "config_role" {
		t.Errorf("Expected role 'config_role', got '%s'", configuredAgent.GetRole())
	}

	if configuredAgent.GetDescription() != "Agent created from config" {
		t.Errorf("Expected description 'Agent created from config', got '%s'", configuredAgent.GetDescription())
	}

	if configuredAgent.GetTimeout() != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", configuredAgent.GetTimeout())
	}
}

func TestConfigurableAgentFactory_CreateAllEnabledAgents(t *testing.T) {
	// Create test configuration with multiple agents
	config := &Config{
		LLM: AgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   800,
		},
		Agents: map[string]AgentConfig{
			"agent1": {
				Role:         "role1",
				Description:  "First agent",
				SystemPrompt: "You are agent 1",
				Capabilities: []string{"capability1"},
				Enabled:      true,
			},
			"agent2": {
				Role:         "role2",
				Description:  "Second agent",
				SystemPrompt: "You are agent 2",
				Capabilities: []string{"capability2"},
				Enabled:      true,
			},
			"disabled_agent": {
				Role:         "disabled_role",
				Description:  "Disabled agent",
				SystemPrompt: "You are disabled",
				Capabilities: []string{"disabled_capability"},
				Enabled:      false,
			},
		},
		Providers: map[string]map[string]interface{}{
			"openai": {
				"api_key": "test-key",
				"model":   "gpt-4",
			},
		},
	}

	factory := NewConfigurableAgentFactory(config)

	// Test creating all enabled agents
	agents, err := factory.CreateAllEnabledAgents()
	if err != nil {
		t.Fatalf("Failed to create all enabled agents: %v", err)
	}

	// Should have 2 enabled agents
	if len(agents) != 2 {
		t.Errorf("Expected 2 enabled agents, got %d", len(agents))
	}

	// Check that the correct agents were created
	if _, exists := agents["agent1"]; !exists {
		t.Error("Expected agent1 to be created")
	}
	if _, exists := agents["agent2"]; !exists {
		t.Error("Expected agent2 to be created")
	}
	if _, exists := agents["disabled_agent"]; exists {
		t.Error("Expected disabled_agent to NOT be created")
	}
}

func TestConfigurableAgentFactory_ValidateAgentConfiguration(t *testing.T) {
	// Create test configuration
	config := &Config{
		LLM: AgentLLMConfig{
			Provider: "openai",
			Model:    "gpt-4",
		},
		Agents: map[string]AgentConfig{
			"valid_agent": {
				Role:         "valid_role",
				Description:  "Valid agent",
				SystemPrompt: "You are a valid agent",
				Capabilities: []string{"validation"},
				Enabled:      true,
			},
			"invalid_agent": {
				Role:         "", // Missing role
				Description:  "Invalid agent",
				SystemPrompt: "", // Missing system prompt
				Capabilities: []string{}, // No capabilities
				Enabled:      true,
			},
			"disabled_agent": {
				Role:         "disabled_role",
				Description:  "Disabled agent",
				SystemPrompt: "You are disabled",
				Capabilities: []string{"disabled"},
				Enabled:      false,
			},
		},
	}

	factory := NewConfigurableAgentFactory(config)

	// Test valid agent
	err := factory.ValidateAgentConfiguration("valid_agent")
	if err != nil {
		t.Errorf("Expected valid_agent to be valid, got error: %v", err)
	}

	// Test invalid agent (missing role)
	err = factory.ValidateAgentConfiguration("invalid_agent")
	if err == nil {
		t.Error("Expected invalid_agent to be invalid due to missing role")
	}

	// Test disabled agent
	err = factory.ValidateAgentConfiguration("disabled_agent")
	if err == nil {
		t.Error("Expected disabled_agent to be invalid due to being disabled")
	}

	// Test non-existent agent
	err = factory.ValidateAgentConfiguration("nonexistent_agent")
	if err == nil {
		t.Error("Expected nonexistent_agent to be invalid")
	}
}

func TestConfigurableAgentFactory_HelperMethods(t *testing.T) {
	// Create test configuration
	config := &Config{
		Agents: map[string]AgentConfig{
			"test_agent": {
				Role:         "test_role",
				Capabilities: []string{"cap1", "cap2"},
				Enabled:      true,
			},
			"disabled_agent": {
				Role:    "disabled_role",
				Enabled: false,
			},
		},
	}

	factory := NewConfigurableAgentFactory(config)

	// Test GetAgentCapabilities
	capabilities := factory.GetAgentCapabilities("test_agent")
	if len(capabilities) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(capabilities))
	}
	if capabilities[0] != "cap1" || capabilities[1] != "cap2" {
		t.Errorf("Expected ['cap1', 'cap2'], got %v", capabilities)
	}

	// Test IsAgentEnabled
	if !factory.IsAgentEnabled("test_agent") {
		t.Error("Expected test_agent to be enabled")
	}
	if factory.IsAgentEnabled("disabled_agent") {
		t.Error("Expected disabled_agent to be disabled")
	}
	if factory.IsAgentEnabled("nonexistent_agent") {
		t.Error("Expected nonexistent_agent to be disabled")
	}
}

func TestConfiguredAgent_Run(t *testing.T) {
	// Create a mock underlying agent
	mockAgent := &MockAgent{
		name: "mock_agent",
		runFunc: func(ctx context.Context, state State) (State, error) {
			outputState := state.Clone()
			outputState.Set("processed", true)
			return outputState, nil
		},
	}

	// Create resolved configuration
	resolvedConfig := &ResolvedAgentConfig{
		Name:         "configured_agent",
		Role:         "test_role",
		Description:  "Test configured agent",
		SystemPrompt: "You are a test agent",
		Capabilities: []string{"testing"},
		Enabled:      true,
		Timeout:      30 * time.Second,
	}

	// Create configured agent
	configuredAgent := &ConfiguredAgent{
		Agent:     mockAgent,
		AgentName: "configured_agent",
		Config:    resolvedConfig,
	}

	// Test running the agent
	ctx := context.Background()
	inputState := NewState()
	inputState.Set("input", "test")

	outputState, err := configuredAgent.Run(ctx, inputState)
	if err != nil {
		t.Fatalf("Failed to run configured agent: %v", err)
	}

	// Verify configuration metadata was added
	if agentName, _ := outputState.Get("agent_name"); agentName != "configured_agent" {
		t.Errorf("Expected agent_name 'configured_agent', got '%v'", agentName)
	}

	if role, _ := outputState.Get("agent_role"); role != "test_role" {
		t.Errorf("Expected agent_role 'test_role', got '%v'", role)
	}

	if executedBy, _ := outputState.Get("executed_by"); executedBy != "configured_agent" {
		t.Errorf("Expected executed_by 'configured_agent', got '%v'", executedBy)
	}

	// Verify underlying agent was called
	if processed, _ := outputState.Get("processed"); processed != true {
		t.Error("Expected underlying agent to be called")
	}
}

func TestConfiguredAgent_InterfaceMethods(t *testing.T) {
	// Create resolved configuration
	resolvedConfig := &ResolvedAgentConfig{
		Name:         "test_agent",
		Role:         "test_role",
		Description:  "Test description",
		SystemPrompt: "Test system prompt",
		Capabilities: []string{"cap1", "cap2"},
		Enabled:      true,
		Timeout:      45 * time.Second,
		LLMConfig: &ResolvedLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
		},
	}

	// Create configured agent
	configuredAgent := &ConfiguredAgent{
		Agent:     &MockAgent{name: "mock"},
		AgentName: "test_agent",
		Config:    resolvedConfig,
	}

	// Test interface methods
	if configuredAgent.Name() != "test_agent" {
		t.Errorf("Expected name 'test_agent', got '%s'", configuredAgent.Name())
	}

	if configuredAgent.GetRole() != "test_role" {
		t.Errorf("Expected role 'test_role', got '%s'", configuredAgent.GetRole())
	}

	if configuredAgent.GetDescription() != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", configuredAgent.GetDescription())
	}

	if configuredAgent.GetSystemPrompt() != "Test system prompt" {
		t.Errorf("Expected system prompt 'Test system prompt', got '%s'", configuredAgent.GetSystemPrompt())
	}

	capabilities := configuredAgent.GetCapabilities()
	if len(capabilities) != 2 || capabilities[0] != "cap1" || capabilities[1] != "cap2" {
		t.Errorf("Expected capabilities ['cap1', 'cap2'], got %v", capabilities)
	}

	if !configuredAgent.IsEnabled() {
		t.Error("Expected agent to be enabled")
	}

	if configuredAgent.GetTimeout() != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", configuredAgent.GetTimeout())
	}

	llmConfig := configuredAgent.GetLLMConfig()
	if llmConfig == nil {
		t.Fatal("Expected LLM config to be present")
	}
	if llmConfig.Provider != "openai" {
		t.Errorf("Expected LLM provider 'openai', got '%s'", llmConfig.Provider)
	}
}

func TestConfigurableAgentFactory_DisabledAgent(t *testing.T) {
	// Create resolved configuration for disabled agent
	resolvedConfig := &ResolvedAgentConfig{
		Name:         "disabled_agent",
		Role:         "disabled_role",
		Description:  "Disabled agent",
		SystemPrompt: "You are disabled",
		Capabilities: []string{"disabled"},
		Enabled:      false, // Disabled
	}

	config := &Config{}
	factory := NewConfigurableAgentFactory(config)

	// Test creating disabled agent should fail
	_, err := factory.CreateAgent("disabled_agent", resolvedConfig, nil)
	if err == nil {
		t.Error("Expected error when creating disabled agent")
	}
}

// Mock implementations for testing

type MockModelProvider struct{}

func (m *MockModelProvider) Call(ctx context.Context, prompt Prompt) (Response, error) {
	return Response{
		Content: "Mock response",
	}, nil
}

func (m *MockModelProvider) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
	ch := make(chan Token, 1)
	ch <- Token{Content: "Mock response"}
	close(ch)
	return ch, nil
}

func (m *MockModelProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	for i := range texts {
		embeddings[i] = []float64{0.1, 0.2, 0.3}
	}
	return embeddings, nil
}

type MockAgent struct {
	name    string
	runFunc func(context.Context, State) (State, error)
}

func (m *MockAgent) Name() string {
	return m.name
}

func (m *MockAgent) Run(ctx context.Context, state State) (State, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, state)
	}
	return state, nil
}