package core

import (
	"testing"
)

// TestNewAgentManager tests the NewAgentManager factory function
func TestNewAgentManager(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string // expected implementation type in logs
	}{
		{
			name:   "with nil config",
			config: nil,
			want:   "basic",
		},
		{
			name:   "with empty config",
			config: &Config{},
			want:   "basic",
		},
		{
			name: "with agents config",
			config: &Config{
				Agents: map[string]AgentConfig{
					"test-agent": {
						Role:         "tester",
						Description:  "Test agent",
						SystemPrompt: "You are a test agent",
						Capabilities: []string{"testing"},
						Enabled:      true,
					},
				},
			},
			want: "basic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewAgentManager(tt.config)

			// Verify we get a valid manager
			if manager == nil {
				t.Fatal("NewAgentManager returned nil")
			}

			// Verify it implements AgentManager interface
			_, ok := manager.(AgentManager)
			if !ok {
				t.Fatal("Returned manager does not implement AgentManager interface")
			}

			// Test all interface methods work
			err := manager.InitializeAgents()
			if err != nil {
				t.Fatalf("InitializeAgents failed: %v", err)
			}

			agents := manager.GetActiveAgents()
			if agents == nil {
				t.Fatal("GetActiveAgents returned nil")
			}

			currentAgents := manager.GetCurrentAgents()
			if currentAgents == nil {
				t.Fatal("GetCurrentAgents returned nil")
			}
		})
	}
}

// TestNewConfigurableAgentFactory tests the NewConfigurableAgentFactory factory function
func TestNewConfigurableAgentFactory(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "with nil config",
			config: nil,
		},
		{
			name:   "with empty config",
			config: &Config{},
		},
		{
			name: "with agent config",
			config: &Config{
				Agents: map[string]AgentConfig{
					"test-agent": {
						Role:         "tester",
						Capabilities: []string{"testing"},
						Enabled:      true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewConfigurableAgentFactory(tt.config)

			// Verify we get a valid factory
			if factory == nil {
				t.Fatal("NewConfigurableAgentFactory returned nil")
			}

			// Verify it implements ConfigurableAgentFactory interface
			_, ok := factory.(ConfigurableAgentFactory)
			if !ok {
				t.Fatal("Returned factory does not implement ConfigurableAgentFactory interface")
			}

			// Test GetAgentCapabilities method
			capabilities := factory.GetAgentCapabilities("test-agent")
			if capabilities == nil {
				t.Fatal("GetAgentCapabilities returned nil")
			}
		})
	}
}

// TestFactoryFunctionConsistency tests that factory functions have consistent behavior
func TestFactoryFunctionConsistency(t *testing.T) {
	config := &Config{
		Agents: map[string]AgentConfig{
			"agent1": {
				Role:         "processor",
				Capabilities: []string{"processing"},
				Enabled:      true,
			},
			"agent2": {
				Role:         "analyzer",
				Capabilities: []string{"analysis"},
				Enabled:      false,
			},
		},
	}

	// Test that both factory functions work with the same config
	manager := NewAgentManager(config)
	factory := NewConfigurableAgentFactory(config)

	if manager == nil {
		t.Fatal("NewAgentManager returned nil")
	}
	if factory == nil {
		t.Fatal("NewConfigurableAgentFactory returned nil")
	}

	// Test that manager initialization works
	err := manager.InitializeAgents()
	if err != nil {
		t.Fatalf("InitializeAgents failed: %v", err)
	}

	// Test that we get expected agents
	activeAgents := manager.GetActiveAgents()
	if len(activeAgents) == 0 {
		t.Log("No active agents found (expected with basic implementation)")
	}

	// Test factory capabilities
	capabilities1 := factory.GetAgentCapabilities("agent1")
	capabilities2 := factory.GetAgentCapabilities("agent2")

	if len(capabilities1) == 0 && len(capabilities2) == 0 {
		t.Log("No capabilities found (expected with basic implementation)")
	}
}

// TestBasicImplementationWarnings tests that basic implementations provide appropriate warnings
func TestBasicImplementationWarnings(t *testing.T) {
	// This test ensures that when no internal packages are imported,
	// we get basic implementations with appropriate warnings

	config := &Config{
		Agents: map[string]AgentConfig{
			"test-agent": {
				Role:    "tester",
				Enabled: true,
			},
		},
	}

	manager := NewAgentManager(config)

	// Initialize agents
	err := manager.InitializeAgents()
	if err != nil {
		t.Fatalf("InitializeAgents failed: %v", err)
	}

	// Get active agents
	agents := manager.GetActiveAgents()
	if len(agents) == 0 {
		t.Fatal("Expected at least one active agent")
	}

	// Test that each agent implements the required interface
	for _, agent := range agents {
		if agent.Name() == "" {
			t.Error("Agent name is empty")
		}
		if agent.GetRole() == "" {
			t.Error("Agent role is empty")
		}
		if !agent.IsEnabled() {
			t.Error("Agent should be enabled by default")
		}
		if agent.GetTimeout() <= 0 {
			t.Error("Agent timeout should be positive")
		}

		capabilities := agent.GetCapabilities()
		if len(capabilities) == 0 {
			t.Error("Agent should have at least one capability")
		}
	}
}

// TestConfigurationIntegration tests that factory functions work with real configuration
func TestConfigurationIntegration(t *testing.T) {
	// Test with a more comprehensive configuration
	config := &Config{
		Agents: map[string]AgentConfig{
			"agent1": {
				Role:         "data_processor",
				Description:  "Processes incoming data",
				SystemPrompt: "You are a data processing agent. Process incoming data efficiently.",
				Capabilities: []string{"data_processing", "validation", "transformation"},
				Enabled:      true,
				Timeout:      45, // seconds
			},
			"agent2": {
				Role:         "analyzer",
				Description:  "Analyzes processed data",
				SystemPrompt: "You are an analysis agent. Analyze data for patterns and insights.",
				Capabilities: []string{"analysis", "pattern_recognition", "reporting"},
				Enabled:      true,
				Timeout:      60, // seconds
			},
		},
	}

	// Test factory functions
	manager := NewAgentManager(config)
	factory := NewConfigurableAgentFactory(config)

	// Initialize and test manager
	err := manager.InitializeAgents()
	if err != nil {
		t.Fatalf("Failed to initialize agents: %v", err)
	}

	// Get active agents and verify configuration integration
	agents := manager.GetActiveAgents()

	agentMap := make(map[string]Agent)
	for _, agent := range agents {
		agentMap[agent.Name()] = agent
	}

	// Test agent1 configuration
	if agent1, exists := agentMap["agent1"]; exists {
		if agent1.GetRole() != "data_processor" {
			t.Errorf("Expected agent1 role 'data_processor', got '%s'", agent1.GetRole())
		}
		if agent1.GetTimeout().Seconds() != 45 {
			t.Errorf("Expected agent1 timeout 45s, got %v", agent1.GetTimeout())
		}
		capabilities := agent1.GetCapabilities()
		expectedCaps := map[string]bool{
			"data_processing": true,
			"validation":      true,
			"transformation":  true,
		}
		for _, cap := range capabilities {
			if !expectedCaps[cap] && cap != "basic_processing" {
				t.Errorf("Unexpected capability '%s' for agent1", cap)
			}
		}
	}

	// Test factory capabilities
	agent1Caps := factory.GetAgentCapabilities("agent1")
	agent2Caps := factory.GetAgentCapabilities("agent2")

	if len(agent1Caps) == 0 {
		t.Error("Expected agent1 to have capabilities")
	}
	if len(agent2Caps) == 0 {
		t.Error("Expected agent2 to have capabilities")
	}
}
