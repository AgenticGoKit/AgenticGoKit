package core

import (
	"context"
	"testing"
)

func TestAgentBuilder_BasicFunctionality(t *testing.T) {
	// Test creating a basic agent
	agent, err := NewAgent("test-agent").Build()
	if err != nil {
		t.Fatalf("Failed to create basic agent: %v", err)
	}

	if agent.Name() != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", agent.Name())
	}
}

func TestAgentBuilder_WithMetrics(t *testing.T) {
	// Test creating an agent with metrics capability
	agent, err := NewAgent("metrics-agent").
		WithDefaultMetrics().
		Build()

	if err != nil {
		t.Fatalf("Failed to create agent with metrics: %v", err)
	}

	if agent.Name() != "metrics-agent" {
		t.Errorf("Expected agent name 'metrics-agent', got '%s'", agent.Name())
	}
}

func TestAgentBuilder_CapabilityIntrospection(t *testing.T) {
	// Test builder introspection methods
	builder := NewAgent("test-agent").
		WithDefaultMetrics()

	if !builder.HasCapability(CapabilityTypeMetrics) {
		t.Error("Expected builder to have metrics capability")
	}

	if builder.CapabilityCount() != 1 {
		t.Errorf("Expected 1 capability, got %d", builder.CapabilityCount())
	}

	capabilities := builder.ListCapabilities()
	if len(capabilities) != 1 || capabilities[0] != CapabilityTypeMetrics {
		t.Errorf("Expected [metrics] capabilities, got %v", capabilities)
	}
}

func TestAgentBuilder_ValidationErrors(t *testing.T) {
	// Test that nil capabilities cause errors
	builder := NewAgent("test-agent").
		WithMCP(nil) // This should cause an error

	if !builder.HasErrors() {
		t.Error("Expected builder to have errors after adding nil MCP manager")
	}

	errors := builder.GetErrors()
	if len(errors) == 0 {
		t.Error("Expected at least one error")
	}
}

func TestAgentRun(t *testing.T) {
	// Test that the placeholder agent can run
	agent, err := NewAgent("run-test").
		WithDefaultMetrics().
		Build()

	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Create a test state
	inputState := NewState()
	inputState.Set("test", "value")

	// Run the agent
	outputState, err := agent.Run(context.Background(), inputState)
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	// Check that the agent processed the state
	processedBy, exists := outputState.Get("processed_by")
	if !exists {
		t.Error("Expected 'processed_by' field in output state")
	}

	if processedBy != "run-test" {
		t.Errorf("Expected processed_by='run-test', got '%v'", processedBy)
	}

	// Check that capabilities were added
	capabilities, exists := outputState.Get("capabilities")
	if !exists {
		t.Error("Expected 'capabilities' field in output state")
	}

	capSlice, ok := capabilities.([]string)
	if !ok {
		t.Errorf("Expected capabilities to be []string, got %T", capabilities)
	}

	if len(capSlice) != 1 || capSlice[0] != "metrics" {
		t.Errorf("Expected capabilities=['metrics'], got %v", capSlice)
	}
}

func TestCapabilityRegistry(t *testing.T) {
	// Test the global capability registry
	registry := GlobalCapabilityRegistry

	// Check that default capabilities are registered
	types := registry.List()
	expectedTypes := []CapabilityType{
		CapabilityTypeLLM,
		CapabilityTypeCache,
		CapabilityTypeMetrics,
		CapabilityTypeMCP,
	}

	for _, expectedType := range expectedTypes {
		found := false
		for _, actualType := range types {
			if actualType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected capability type %s to be registered", expectedType)
		}
	}
}

func TestCapabilityValidation(t *testing.T) {
	// Test capability validation
	validator := &CapabilityValidator{}

	// Test ValidateUnique
	capabilities := []AgentCapability{
		NewMetricsCapability(DefaultMetricsConfig()),
		NewMetricsCapability(DefaultMetricsConfig()), // Duplicate
	}

	err := validator.ValidateUnique(CapabilityTypeMetrics, capabilities[1:])
	if err == nil {
		t.Error("Expected error for duplicate metrics capability")
	}
}
