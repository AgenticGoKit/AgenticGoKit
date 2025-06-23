package core

import (
	"context"
	"testing"
)

func TestUnifiedAgent_BasicFunctionality(t *testing.T) {
	// Test creating a unified agent directly
	capabilities := map[CapabilityType]AgentCapability{
		CapabilityTypeMetrics: NewMetricsCapability(DefaultMetricsConfig()),
	}

	agent := NewUnifiedAgent("test-unified", capabilities, nil)

	if agent.Name() != "test-unified" {
		t.Errorf("Expected agent name 'test-unified', got '%s'", agent.Name())
	}

	if !agent.HasCapability(CapabilityTypeMetrics) {
		t.Error("Expected agent to have metrics capability")
	}

	if agent.HasCapability(CapabilityTypeLLM) {
		t.Error("Expected agent to not have LLM capability")
	}
}

func TestUnifiedAgent_CapabilityManagement(t *testing.T) {
	// Start with empty capabilities
	agent := NewUnifiedAgent("capability-test", nil, nil)

	// Check initial state
	if len(agent.ListCapabilities()) != 0 {
		t.Errorf("Expected 0 capabilities, got %d", len(agent.ListCapabilities()))
	}

	// Test getting non-existent capability
	_, exists := agent.GetCapability(CapabilityTypeMetrics)
	if exists {
		t.Error("Expected metrics capability to not exist")
	}

	// Add a capability through CapabilityConfigurable interface
	metricsConfig := DefaultMetricsConfig()
	agent.SetMetricsConfig(metricsConfig)

	// Verify it was added
	if !agent.HasCapability(CapabilityTypeMetrics) {
		t.Error("Expected agent to have metrics capability after SetMetricsConfig")
	}

	metricsCap, exists := agent.GetCapability(CapabilityTypeMetrics)
	if !exists {
		t.Error("Expected to retrieve metrics capability")
	}

	if metrics, ok := metricsCap.(*MetricsCapability); ok {
		if metrics.Config.Port != metricsConfig.Port {
			t.Errorf("Expected metrics port %d, got %d", metricsConfig.Port, metrics.Config.Port)
		}
	} else {
		t.Error("Expected metrics capability to be of type *MetricsCapability")
	}
}

func TestUnifiedAgent_Run(t *testing.T) {
	// Create agent with metrics capability
	capabilities := map[CapabilityType]AgentCapability{
		CapabilityTypeMetrics: NewMetricsCapability(DefaultMetricsConfig()),
	}

	agent := NewUnifiedAgent("run-test", capabilities, nil)

	// Create test state
	inputState := NewState()
	inputState.Set("test_input", "value")

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

	// Check that capability metadata was added
	capabilities_field, exists := outputState.Get("capabilities")
	if !exists {
		t.Error("Expected 'capabilities' field in output state")
	}

	capSlice, ok := capabilities_field.([]string)
	if !ok {
		t.Errorf("Expected capabilities to be []string, got %T", capabilities_field)
	}

	if len(capSlice) != 1 || capSlice[0] != "metrics" {
		t.Errorf("Expected capabilities=['metrics'], got %v", capSlice)
	}

	// Check that metrics pre-processing was applied
	metricsEnabled, exists := outputState.Get("metrics_enabled")
	if !exists {
		t.Error("Expected 'metrics_enabled' field in output state")
	}

	if metricsEnabled != true {
		t.Errorf("Expected metrics_enabled=true, got %v", metricsEnabled)
	}

	// Check that metrics post-processing was applied
	metricsCollected, exists := outputState.Get("metrics_collected")
	if !exists {
		t.Error("Expected 'metrics_collected' field in output state")
	}

	if metricsCollected != true {
		t.Errorf("Expected metrics_collected=true, got %v", metricsCollected)
	}

	// Verify original input is preserved
	testInput, exists := outputState.Get("test_input")
	if !exists {
		t.Error("Expected original 'test_input' field to be preserved")
	}

	if testInput != "value" {
		t.Errorf("Expected test_input='value', got '%v'", testInput)
	}
}

func TestUnifiedAgent_WithCustomHandler(t *testing.T) {
	// Create a custom handler
	customHandler := AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		outputState := state.Clone()
		outputState.Set("custom_processing", true)
		outputState.Set("handler_used", "custom")
		return AgentResult{OutputState: outputState}, nil
	})

	// Create agent with custom handler and metrics capability
	capabilities := map[CapabilityType]AgentCapability{
		CapabilityTypeMetrics: NewMetricsCapability(DefaultMetricsConfig()),
	}

	agent := NewUnifiedAgent("custom-handler-test", capabilities, customHandler)

	// Run the agent
	inputState := NewState()
	inputState.Set("input", "test")

	outputState, err := agent.Run(context.Background(), inputState)
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	// Check that custom handler was used
	handlerUsed, exists := outputState.Get("handler_used")
	if !exists {
		t.Error("Expected 'handler_used' field in output state")
	}

	if handlerUsed != "custom" {
		t.Errorf("Expected handler_used='custom', got '%v'", handlerUsed)
	}

	// Check that custom processing was applied
	customProcessing, exists := outputState.Get("custom_processing")
	if !exists {
		t.Error("Expected 'custom_processing' field in output state")
	}

	if customProcessing != true {
		t.Errorf("Expected custom_processing=true, got %v", customProcessing)
	}

	// Check that capability pre/post-processing still happened
	metricsEnabled, exists := outputState.Get("metrics_enabled")
	if !exists {
		t.Error("Expected 'metrics_enabled' field in output state (capability pre-processing)")
	}

	metricsCollected, exists := outputState.Get("metrics_collected")
	if !exists {
		t.Error("Expected 'metrics_collected' field in output state (capability post-processing)")
	}

	if metricsEnabled != true || metricsCollected != true {
		t.Error("Expected capability pre/post-processing to occur even with custom handler")
	}
}

func TestUnifiedAgent_Configuration(t *testing.T) {
	// Create agent
	agent := NewUnifiedAgent("config-test", nil, nil)
	// Test LLM configuration
	mockProvider := &MockProvider{}
	llmConfig := LLMConfig{
		Temperature: 0.5,
		MaxTokens:   500,
	}

	agent.SetLLMProvider(mockProvider, llmConfig)

	// Verify LLM capability was created and configured
	if !agent.HasCapability(CapabilityTypeLLM) {
		t.Error("Expected agent to have LLM capability after SetLLMProvider")
	}

	llmCap, exists := agent.GetCapability(CapabilityTypeLLM)
	if !exists {
		t.Error("Expected to retrieve LLM capability")
	}

	if llm, ok := llmCap.(*LLMCapability); ok {
		if llm.Provider != mockProvider {
			t.Error("Expected LLM provider to be set correctly")
		}
		if llm.Config.Temperature != 0.5 {
			t.Errorf("Expected LLM temperature 0.5, got %f", llm.Config.Temperature)
		}
		if llm.Config.MaxTokens != 500 {
			t.Errorf("Expected LLM max tokens 500, got %d", llm.Config.MaxTokens)
		}
	} else {
		t.Error("Expected LLM capability to be of type *LLMCapability")
	}
}

func TestUnifiedAgent_String(t *testing.T) {
	capabilities := map[CapabilityType]AgentCapability{
		CapabilityTypeMetrics: NewMetricsCapability(DefaultMetricsConfig()),
		CapabilityTypeLLM:     NewLLMCapability(&MockProvider{}, LLMConfig{}),
	}

	agent := NewUnifiedAgent("string-test", capabilities, nil)

	str := agent.String()
	expected := "UnifiedAgent{name=string-test, capabilities=[metrics llm]}"

	// Since map iteration order is not guaranteed, we need to check if both expected variations are valid
	alternative := "UnifiedAgent{name=string-test, capabilities=[llm metrics]}"

	if str != expected && str != alternative {
		t.Errorf("Expected string representation to be '%s' or '%s', got '%s'", expected, alternative, str)
	}
}
