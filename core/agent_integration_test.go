package core

import (
	"context"
	"testing"
)

// TestAgentInterface demonstrates the unified agent interface design
// This test shows how the consolidated Agent interface eliminates adapter patterns
func TestAgentInterface(t *testing.T) {
	config := &Config{
		Agents: map[string]AgentConfig{
			"unified-agent": {
				Role:         "unified_processor",
				Description:  "Demonstrates unified interface",
				SystemPrompt: "You are a unified agent supporting both patterns",
				Capabilities: []string{"state_processing", "event_handling"},
				Enabled:      true,
				Timeout:      30,
			},
		},
	}

	// Create agent through manager
	manager := NewAgentManager(config)
	err := manager.InitializeAgents()
	if err != nil {
		t.Fatalf("Failed to initialize agents: %v", err)
	}

	agents := manager.GetActiveAgents()
	if len(agents) == 0 {
		t.Fatal("No agents created")
	}

	agent := agents[0]

	// Test both execution patterns with the same agent instance
	t.Run("StateProcessingPattern", func(t *testing.T) {
		ctx := context.Background()
		inputState := NewState()
		inputState.Set("test_data", "state processing test")

		outputState, err := agent.Run(ctx, inputState)
		if err != nil {
			t.Fatalf("State processing failed: %v", err)
		}

		processedBy, exists := outputState.Get("processed_by")
		if !exists || processedBy != agent.Name() {
			t.Error("State not processed correctly")
		}
	})

	t.Run("EventDrivenPattern", func(t *testing.T) {
		ctx := context.Background()
		event := NewEvent("test-target", map[string]any{"event_type": "test_event"}, map[string]string{})
		state := NewState()
		state.Set("test_data", "event processing test")

		result, err := agent.HandleEvent(ctx, event, state)
		if err != nil {
			t.Fatalf("Event handling failed: %v", err)
		}

		processedBy, exists := result.OutputState.Get("processed_by")
		if !exists || processedBy != agent.Name() {
			t.Error("Event not processed correctly")
		}

		if result.Duration < 0 {
			t.Error("Result should have non-negative execution duration")
		}
	})

	t.Run("AgentCapabilities", func(t *testing.T) {
		// Test that agent implements all required capabilities
		if agent.Name() == "" {
			t.Error("Agent should have a name")
		}
		if agent.GetRole() == "" {
			t.Error("Agent should have a role")
		}
		if len(agent.GetCapabilities()) == 0 {
			t.Error("Agent should have capabilities")
		}
		if agent.GetTimeout() <= 0 {
			t.Error("Agent should have a positive timeout")
		}
		if !agent.IsEnabled() {
			t.Error("Agent should be enabled")
		}
	})

	t.Run("LifecycleManagement", func(t *testing.T) {
		ctx := context.Background()

		// Test initialization
		err := agent.Initialize(ctx)
		if err != nil {
			t.Fatalf("Agent initialization failed: %v", err)
		}

		// Test shutdown
		err = agent.Shutdown(ctx)
		if err != nil {
			t.Fatalf("Agent shutdown failed: %v", err)
		}
	})
}

// TestAgentAdapterElimination demonstrates how adapters are no longer needed
func TestAgentAdapterElimination(t *testing.T) {
	config := &Config{
		Agents: map[string]AgentConfig{
			"multi-pattern-agent": {
				Role:    "multi_processor",
				Enabled: true,
			},
		},
	}

	manager := NewAgentManager(config)
	err := manager.InitializeAgents()
	if err != nil {
		t.Fatalf("Failed to initialize agents: %v", err)
	}

	agents := manager.GetActiveAgents()
	if len(agents) == 0 {
		t.Fatal("No agents created")
	}

	agent := agents[0]

	// Simulate orchestrator usage - no adapter needed!
	t.Run("DirectOrchestrationUsage", func(t *testing.T) {
		// Before: Required AgentToHandlerAdapter
		// agentHandler := &AgentToHandlerAdapter{agent: agent}
		// result, err := agentHandler.Run(ctx, event, state)

		// After: Direct usage, no adapter needed
		ctx := context.Background()
		event := NewEvent("test-target", map[string]any{"orchestrator_event": "direct_call"}, map[string]string{})
		state := NewState()
		state.Set("orchestrator_data", "test")

		result, err := agent.HandleEvent(ctx, event, state)
		if err != nil {
			t.Fatalf("Direct event handling failed: %v", err)
		}

		if result.OutputState == nil {
			t.Fatal("Result should have output state")
		}
	})

	t.Run("DirectStateProcessingUsage", func(t *testing.T) {
		// Direct usage for state processing workflows
		ctx := context.Background()
		inputState := NewState()
		inputState.Set("workflow_data", "process this")

		outputState, err := agent.Run(ctx, inputState)
		if err != nil {
			t.Fatalf("Direct state processing failed: %v", err)
		}

		if outputState == nil {
			t.Fatal("Should have output state")
		}
	})

	t.Run("UnifiedInterfaceTypes", func(t *testing.T) {
		// Verify type compatibility - no casting needed
		var agentInterface Agent = agent
		var handlerFunc func(context.Context, Event, State) (AgentResult, error) = agent.HandleEvent
		var runFunc func(context.Context, State) (State, error) = agent.Run

		if agentInterface == nil || handlerFunc == nil || runFunc == nil {
			t.Error("Agent should be directly compatible with both patterns")
		}

		// Test polymorphic usage
		agents := []Agent{agent} // Can store any agent implementation
		for _, a := range agents {
			// Both patterns work with the same interface
			_, err1 := a.Run(context.Background(), NewState())
			_, err2 := a.HandleEvent(context.Background(), NewEvent("test", map[string]any{}, map[string]string{}), NewState())

			if err1 != nil || err2 != nil {
				t.Error("Polymorphic usage should work seamlessly")
			}
		}
	})
}

// TestAgentBackwardCompatibilityBridge shows how existing code can be migrated
func TestAgentBackwardCompatibilityBridge(t *testing.T) {
	// Legacy AgentHandler usage can be bridged temporarily
	config := &Config{
		Agents: map[string]AgentConfig{
			"legacy-agent": {
				Role:    "legacy_processor",
				Enabled: true,
			},
		},
	}

	manager := NewAgentManager(config)
	err := manager.InitializeAgents()
	if err != nil {
		t.Fatalf("Failed to initialize agents: %v", err)
	}

	agents := manager.GetActiveAgents()
	agent := agents[0]

	// Create a backward compatibility bridge if needed
	legacyHandler := AgentHandlerFunc(agent.HandleEvent)

	ctx := context.Background()
	event := NewEvent("legacy", map[string]any{}, map[string]string{})
	state := NewState()

	result, err := legacyHandler.Run(ctx, event, state)
	if err != nil {
		t.Fatalf("Legacy handler bridge failed: %v", err)
	}

	if result.OutputState == nil {
		t.Fatal("Legacy handler should produce result")
	}
}

// TestAgentDesignBenefits demonstrates the benefits of the unified design
func TestAgentDesignBenefits(t *testing.T) {
	t.Run("SingleImplementationBothPatterns", func(t *testing.T) {
		// Agents now support both patterns without duplicating logic
		config := &Config{
			Agents: map[string]AgentConfig{
				"benefit-demo": {Role: "demo", Enabled: true},
			},
		}

		manager := NewAgentManager(config)
		manager.InitializeAgents()
		agents := manager.GetActiveAgents()
		agent := agents[0]

		// Same agent, different execution patterns
		ctx := context.Background()
		testState := NewState()
		testState.Set("input", "test")
		testEvent := NewEvent("demo", map[string]any{}, map[string]string{})

		// State processing
		state1, _ := agent.Run(ctx, testState)
		// Event handling
		result1, _ := agent.HandleEvent(ctx, testEvent, testState)

		// Both should process through the same agent
		processedBy1, exists1 := state1.Get("processed_by")
		processedBy2, exists2 := result1.OutputState.Get("processed_by")
		if !exists1 || !exists2 || processedBy1 != processedBy2 {
			t.Error("Both patterns should use the same processing logic")
		}
	})

	t.Run("CleanerArchitecture", func(t *testing.T) {
		// No more adapter proliferation
		// No more interface conversion complexity
		// Single source of truth for agent behavior

		// Agents are directly usable in any context
		manager := NewAgentManager(&Config{
			Agents: map[string]AgentConfig{
				"clean-agent": {Role: "clean", Enabled: true},
			},
		})
		manager.InitializeAgents()
		agents := manager.GetActiveAgents()

		if len(agents) == 0 {
			t.Fatal("Should have agents")
		}

		// Direct usage in orchestrators, workflows, handlers, etc.
		agent := agents[0]

		// Can be used directly without type conversion
		orchestratorAgents := map[string]Agent{
			agent.Name(): agent,
		}

		workflowAgents := []Agent{agent}

		if len(orchestratorAgents) == 0 || len(workflowAgents) == 0 {
			t.Error("Agents should be directly usable in any context")
		}
	})

	t.Run("ExtensibilityAndMaintenance", func(t *testing.T) {
		// Adding new methods to Agent interface affects all implementations
		// No more keeping multiple interfaces in sync
		// Clear separation of concerns with lifecycle methods

		config := &Config{
			Agents: map[string]AgentConfig{
				"extensible-agent": {Role: "extensible", Enabled: true},
			},
		}

		manager := NewAgentManager(config)
		manager.InitializeAgents()
		agents := manager.GetActiveAgents()
		agent := agents[0]

		// All agents have full lifecycle support
		ctx := context.Background()

		if err := agent.Initialize(ctx); err != nil {
			t.Fatalf("Initialize should be available: %v", err)
		}

		if err := agent.Shutdown(ctx); err != nil {
			t.Fatalf("Shutdown should be available: %v", err)
		}

		// All agents have rich metadata
		if agent.GetRole() == "" || len(agent.GetCapabilities()) == 0 {
			t.Error("All agents should have rich metadata")
		}
	})
}
