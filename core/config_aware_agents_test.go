package core

import (
	"context"
	"testing"
	"time"
)

func TestConfigAwareUnifiedAgent(t *testing.T) {
	// Create resolved configuration
	config := &ResolvedAgentConfig{
		Name:         "test_agent",
		Role:         "test_role",
		Description:  "Test configuration-aware agent",
		SystemPrompt: "You are a test agent configured from TOML",
		Capabilities: []string{"testing", "configuration"},
		Enabled:      true,
		Timeout:      30 * time.Second,
		LLMConfig: &ResolvedLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   800,
			Timeout:     30 * time.Second,
		},
	}

	// Create capabilities
	capabilities := map[CapabilityType]AgentCapability{
		CapabilityTypeMetrics: NewMetricsCapability(DefaultMetricsConfig()),
	}

	// Create config-aware agent
	agent := NewConfigAwareUnifiedAgent("test_agent", config, capabilities, nil)

	// Test configuration interface methods
	if agent.GetRole() != "test_role" {
		t.Errorf("Expected role 'test_role', got '%s'", agent.GetRole())
	}

	if agent.GetDescription() != "Test configuration-aware agent" {
		t.Errorf("Expected description 'Test configuration-aware agent', got '%s'", agent.GetDescription())
	}

	if agent.GetSystemPrompt() != "You are a test agent configured from TOML" {
		t.Errorf("Expected system prompt 'You are a test agent configured from TOML', got '%s'", agent.GetSystemPrompt())
	}

	capabilities_list := agent.GetCapabilities()
	if len(capabilities_list) != 2 || capabilities_list[0] != "testing" || capabilities_list[1] != "configuration" {
		t.Errorf("Expected capabilities ['testing', 'configuration'], got %v", capabilities_list)
	}

	if !agent.IsEnabled() {
		t.Error("Expected agent to be enabled")
	}

	if agent.GetTimeout() != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", agent.GetTimeout())
	}

	llmConfig := agent.GetLLMConfig()
	if llmConfig == nil {
		t.Fatal("Expected LLM config to be present")
	}
	if llmConfig.Provider != "openai" {
		t.Errorf("Expected LLM provider 'openai', got '%s'", llmConfig.Provider)
	}
}

func TestConfigAwareUnifiedAgent_Run(t *testing.T) {
	config := &ResolvedAgentConfig{
		Name:         "test_agent",
		Role:         "test_role",
		Description:  "Test agent",
		SystemPrompt: "You are a helpful test agent",
		Capabilities: []string{"testing"},
		Enabled:      true,
		Timeout:      5 * time.Second,
	}

	capabilities := map[CapabilityType]AgentCapability{
		CapabilityTypeMetrics: NewMetricsCapability(DefaultMetricsConfig()),
	}

	agent := NewConfigAwareUnifiedAgent("test_agent", config, capabilities, nil)

	// Test running the agent
	ctx := context.Background()
	inputState := NewState()
	inputState.Set("input", "test")

	outputState, err := agent.Run(ctx, inputState)
	if err != nil {
		t.Fatalf("Failed to run agent: %v", err)
	}

	// Verify configuration metadata was added
	if agentName, _ := outputState.Get("agent_name"); agentName != "test_agent" {
		t.Errorf("Expected agent_name 'test_agent', got '%v'", agentName)
	}

	if role, _ := outputState.Get("agent_role"); role != "test_role" {
		t.Errorf("Expected agent_role 'test_role', got '%v'", role)
	}

	if executedBy, _ := outputState.Get("executed_by"); executedBy != "test_agent" {
		t.Errorf("Expected executed_by 'test_agent', got '%v'", executedBy)
	}

	if systemPrompt, _ := outputState.Get("system_prompt"); systemPrompt != "You are a helpful test agent" {
		t.Errorf("Expected system_prompt to be set, got '%v'", systemPrompt)
	}
}

func TestConfigAwareUnifiedAgent_DisabledAgent(t *testing.T) {
	config := &ResolvedAgentConfig{
		Name:         "disabled_agent",
		Role:         "disabled_role",
		Description:  "Disabled test agent",
		SystemPrompt: "You are disabled",
		Capabilities: []string{"testing"},
		Enabled:      false, // Disabled
		Timeout:      30 * time.Second,
	}

	agent := NewConfigAwareUnifiedAgent("disabled_agent", config, nil, nil)

	// Test that disabled agent fails to run
	ctx := context.Background()
	inputState := NewState()

	_, err := agent.Run(ctx, inputState)
	if err == nil {
		t.Error("Expected error when running disabled agent")
	}

	if !agent.IsEnabled() == false {
		t.Error("Expected IsEnabled() to return false for disabled agent")
	}
}

func TestConfigAwareUnifiedAgent_UpdateConfiguration(t *testing.T) {
	initialConfig := &ResolvedAgentConfig{
		Name:         "test_agent",
		Role:         "initial_role",
		Description:  "Initial description",
		SystemPrompt: "Initial prompt",
		Capabilities: []string{"initial"},
		Enabled:      true,
		Timeout:      30 * time.Second,
	}

	agent := NewConfigAwareUnifiedAgent("test_agent", initialConfig, nil, nil)

	// Update configuration
	newConfig := &ResolvedAgentConfig{
		Name:         "test_agent",
		Role:         "updated_role",
		Description:  "Updated description",
		SystemPrompt: "Updated prompt",
		Capabilities: []string{"updated", "enhanced"},
		Enabled:      true,
		Timeout:      60 * time.Second,
	}

	err := agent.UpdateConfiguration(newConfig)
	if err != nil {
		t.Fatalf("Failed to update configuration: %v", err)
	}

	// Verify updated configuration
	if agent.GetRole() != "updated_role" {
		t.Errorf("Expected updated role 'updated_role', got '%s'", agent.GetRole())
	}

	if agent.GetDescription() != "Updated description" {
		t.Errorf("Expected updated description 'Updated description', got '%s'", agent.GetDescription())
	}

	if agent.GetSystemPrompt() != "Updated prompt" {
		t.Errorf("Expected updated system prompt 'Updated prompt', got '%s'", agent.GetSystemPrompt())
	}

	capabilities := agent.GetCapabilities()
	if len(capabilities) != 2 || capabilities[0] != "updated" || capabilities[1] != "enhanced" {
		t.Errorf("Expected updated capabilities ['updated', 'enhanced'], got %v", capabilities)
	}

	if agent.GetTimeout() != 60*time.Second {
		t.Errorf("Expected updated timeout 60s, got %v", agent.GetTimeout())
	}
}

func TestConfigAwareSequentialAgent(t *testing.T) {
	// Create sub-agents
	agent1Config := &ResolvedAgentConfig{
		Name:         "agent1",
		Role:         "first_agent",
		Description:  "First agent in sequence",
		SystemPrompt: "You are the first agent",
		Capabilities: []string{"first"},
		Enabled:      true,
		Timeout:      10 * time.Second,
	}

	agent2Config := &ResolvedAgentConfig{
		Name:         "agent2",
		Role:         "second_agent",
		Description:  "Second agent in sequence",
		SystemPrompt: "You are the second agent",
		Capabilities: []string{"second"},
		Enabled:      true,
		Timeout:      10 * time.Second,
	}

	agent1 := NewConfigAwareUnifiedAgent("agent1", agent1Config, nil, nil)
	agent2 := NewConfigAwareUnifiedAgent("agent2", agent2Config, nil, nil)

	// Create sequential agent configuration
	seqConfig := &ResolvedAgentConfig{
		Name:         "sequential_agent",
		Role:         "coordinator",
		Description:  "Coordinates sequential execution",
		SystemPrompt: "You coordinate agent execution",
		Capabilities: []string{"coordination", "sequencing"},
		Enabled:      true,
		Timeout:      30 * time.Second,
	}

	seqAgent := NewConfigAwareSequentialAgent("sequential_agent", seqConfig, agent1, agent2)

	// Test configuration methods
	if seqAgent.GetRole() != "coordinator" {
		t.Errorf("Expected role 'coordinator', got '%s'", seqAgent.GetRole())
	}

	if !seqAgent.IsEnabled() {
		t.Error("Expected sequential agent to be enabled")
	}

	// Test running the sequential agent
	ctx := context.Background()
	inputState := NewState()
	inputState.Set("input", "test")

	outputState, err := seqAgent.Run(ctx, inputState)
	if err != nil {
		t.Fatalf("Failed to run sequential agent: %v", err)
	}

	// Verify sequential execution metadata
	if seqAgentName, _ := outputState.Get("sequential_agent"); seqAgentName != "sequential_agent" {
		t.Errorf("Expected sequential_agent 'sequential_agent', got '%v'", seqAgentName)
	}

	if completed, _ := outputState.Get("sequential_execution_completed"); completed != true {
		t.Error("Expected sequential_execution_completed to be true")
	}
}

func TestConfigAwareSequentialAgent_DisabledSubAgent(t *testing.T) {
	// Create enabled and disabled sub-agents
	enabledConfig := &ResolvedAgentConfig{
		Name:         "enabled_agent",
		Role:         "enabled_role",
		Description:  "Enabled agent",
		SystemPrompt: "You are enabled",
		Capabilities: []string{"enabled"},
		Enabled:      true,
		Timeout:      10 * time.Second,
	}

	disabledConfig := &ResolvedAgentConfig{
		Name:         "disabled_agent",
		Role:         "disabled_role",
		Description:  "Disabled agent",
		SystemPrompt: "You are disabled",
		Capabilities: []string{"disabled"},
		Enabled:      false, // Disabled
		Timeout:      10 * time.Second,
	}

	enabledAgent := NewConfigAwareUnifiedAgent("enabled_agent", enabledConfig, nil, nil)
	disabledAgent := NewConfigAwareUnifiedAgent("disabled_agent", disabledConfig, nil, nil)

	seqConfig := &ResolvedAgentConfig{
		Name:         "sequential_agent",
		Role:         "coordinator",
		Description:  "Coordinates sequential execution",
		SystemPrompt: "You coordinate agent execution",
		Capabilities: []string{"coordination"},
		Enabled:      true,
		Timeout:      30 * time.Second,
	}

	// Create sequential agent with both enabled and disabled sub-agents
	seqAgent := NewConfigAwareSequentialAgent("sequential_agent", seqConfig, enabledAgent, disabledAgent)

	// The sequential agent should only include enabled agents
	// We can't directly test the internal agents list, but we can verify behavior
	ctx := context.Background()
	inputState := NewState()

	outputState, err := seqAgent.Run(ctx, inputState)
	if err != nil {
		t.Fatalf("Failed to run sequential agent: %v", err)
	}

	// The execution should complete successfully with only the enabled agent
	if completed, _ := outputState.Get("sequential_execution_completed"); completed != true {
		t.Error("Expected sequential execution to complete successfully")
	}
}

func TestConfigAwareParallelAgent(t *testing.T) {
	// Create sub-agents
	agent1Config := &ResolvedAgentConfig{
		Name:         "parallel_agent1",
		Role:         "parallel_role1",
		Description:  "First parallel agent",
		SystemPrompt: "You are the first parallel agent",
		Capabilities: []string{"parallel1"},
		Enabled:      true,
		Timeout:      10 * time.Second,
	}

	agent2Config := &ResolvedAgentConfig{
		Name:         "parallel_agent2",
		Role:         "parallel_role2",
		Description:  "Second parallel agent",
		SystemPrompt: "You are the second parallel agent",
		Capabilities: []string{"parallel2"},
		Enabled:      true,
		Timeout:      10 * time.Second,
	}

	agent1 := NewConfigAwareUnifiedAgent("parallel_agent1", agent1Config, nil, nil)
	agent2 := NewConfigAwareUnifiedAgent("parallel_agent2", agent2Config, nil, nil)

	// Create parallel agent configuration
	parallelConfig := &ResolvedAgentConfig{
		Name:         "parallel_coordinator",
		Role:         "parallel_coordinator",
		Description:  "Coordinates parallel execution",
		SystemPrompt: "You coordinate parallel agent execution",
		Capabilities: []string{"parallel_coordination"},
		Enabled:      true,
		Timeout:      30 * time.Second,
	}

	parallelAgent := NewConfigAwareParallelAgent("parallel_coordinator", parallelConfig, agent1, agent2)

	// Test configuration methods
	if parallelAgent.GetRole() != "parallel_coordinator" {
		t.Errorf("Expected role 'parallel_coordinator', got '%s'", parallelAgent.GetRole())
	}

	if !parallelAgent.IsEnabled() {
		t.Error("Expected parallel agent to be enabled")
	}

	// Test running the parallel agent
	ctx := context.Background()
	inputState := NewState()
	inputState.Set("input", "test")

	outputState, err := parallelAgent.Run(ctx, inputState)
	if err != nil {
		t.Fatalf("Failed to run parallel agent: %v", err)
	}

	// Verify parallel execution metadata
	if parallelAgentName, _ := outputState.Get("parallel_agent"); parallelAgentName != "parallel_coordinator" {
		t.Errorf("Expected parallel_agent 'parallel_coordinator', got '%v'", parallelAgentName)
	}

	if completed, _ := outputState.Get("parallel_execution_completed"); completed != true {
		t.Error("Expected parallel_execution_completed to be true")
	}
}

func TestConfigAwareAgent_Timeout(t *testing.T) {
	config := &ResolvedAgentConfig{
		Name:         "timeout_agent",
		Role:         "timeout_role",
		Description:  "Agent with timeout",
		SystemPrompt: "You have a timeout",
		Capabilities: []string{"timeout"},
		Enabled:      true,
		Timeout:      100 * time.Millisecond, // Very short timeout
	}

	// Create a mock handler that takes longer than the timeout
	handler := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		time.Sleep(200 * time.Millisecond) // Sleep longer than timeout
		return AgentResult{OutputState: state}, nil
	})

	agent := NewConfigAwareUnifiedAgent("timeout_agent", config, nil, handler)

	ctx := context.Background()
	inputState := NewState()

	_, err := agent.Run(ctx, inputState)
	if err == nil {
		t.Error("Expected timeout error")
	}

	// The error should be related to context cancellation/timeout
	if err != nil && err.Error() != "" {
		// We expect some kind of timeout or cancellation error
		t.Logf("Got expected timeout-related error: %v", err)
	}
}