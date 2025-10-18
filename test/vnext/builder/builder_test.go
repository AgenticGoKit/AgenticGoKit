package builder_test

import (
	"context"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

// TestBuilderBasic tests basic builder functionality
func TestBuilderBasic(t *testing.T) {
	// Test basic builder creation
	builder := vnext.NewBuilder("test-agent")
	if builder == nil {
		t.Fatal("NewBuilder returned nil")
	}

	// Test building an agent with default config
	agent, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if agent.Name() != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", agent.Name())
	}

	// Verify config has defaults
	config := agent.Config()
	if config.Timeout <= 0 {
		t.Error("Expected positive timeout in default config")
	}

	if config.LLM.Provider == "" {
		t.Error("Expected LLM provider in default config")
	}
}

// TestBuilderPresets tests all preset configurations
func TestBuilderPresets(t *testing.T) {
	presets := []vnext.PresetType{vnext.ChatAgent, vnext.ResearchAgent, vnext.DataAgent, vnext.WorkflowAgent}

	for _, preset := range presets {
		t.Run(string(preset), func(t *testing.T) {
			agent, err := vnext.NewBuilder("test-" + string(preset)).
				WithPreset(preset).
				Build()

			if err != nil {
				t.Fatalf("Failed to build %s: %v", preset, err)
			}

			capabilities := agent.Capabilities()
			if len(capabilities) == 0 {
				t.Errorf("Agent %s has no capabilities", preset)
			}

			// All agents should have LLM capability
			found := false
			for _, cap := range capabilities {
				if cap == "llm" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Agent %s missing LLM capability", preset)
			}
		})
	}
}

// TestBuilderClone tests the Clone functionality
func TestBuilderClone(t *testing.T) {
	// Create base builder
	baseBuilder := vnext.NewBuilder("base").
		WithPreset(vnext.ChatAgent).
		WithMemory(vnext.WithMemoryProvider("memory"))

	// Clone it
	clonedBuilder := baseBuilder.Clone()

	// Modify clone
	clonedBuilder.WithMemory(vnext.WithMemoryProvider("pgvector"))

	// Build both
	baseAgent, err1 := baseBuilder.Build()
	clonedAgent, err2 := clonedBuilder.Build()

	if err1 != nil || err2 != nil {
		t.Fatalf("Build failed: %v, %v", err1, err2)
	}

	// Both should work independently
	if baseAgent.Name() != "base" {
		t.Errorf("Base agent name incorrect: %s", baseAgent.Name())
	}

	if clonedAgent.Name() != "base" {
		t.Errorf("Cloned agent name incorrect: %s", clonedAgent.Name())
	}

	// Memory providers should be different
	baseConfig := baseAgent.Config()
	clonedConfig := clonedAgent.Config()

	if baseConfig.Memory.Provider == clonedConfig.Memory.Provider {
		t.Error("Clone did not properly isolate memory configuration")
	}
}

// TestBuilderFunctionalOptions tests functional options
func TestBuilderFunctionalOptions(t *testing.T) {
	agent, err := vnext.NewBuilder("test-options").
		WithMemory(
			vnext.WithMemoryProvider("pgvector"),
			vnext.WithRAG(4000, 0.7, 0.3),
			vnext.WithSessionScoped(),
		).
		WithTools(
			vnext.WithToolTimeout(60*time.Second),
			vnext.WithMaxConcurrentTools(10),
		).
		Build()

	if err != nil {
		t.Fatalf("Build with options failed: %v", err)
	}

	config := agent.Config()

	// Check memory configuration
	if config.Memory == nil {
		t.Fatal("Memory config is nil")
	}

	if config.Memory.Provider != "pgvector" {
		t.Errorf("Expected memory provider 'pgvector', got '%s'", config.Memory.Provider)
	}

	// Check session scoped in options
	if config.Memory.Options == nil || config.Memory.Options["session_scoped"] != "true" {
		t.Error("Expected session-scoped memory to be enabled")
	}

	if config.Memory.RAG == nil {
		t.Fatal("RAG config is nil")
	}

	if config.Memory.RAG.MaxTokens != 4000 {
		t.Errorf("Expected RAG max tokens 4000, got %d", config.Memory.RAG.MaxTokens)
	}

	// Check tools configuration
	if config.Tools == nil {
		t.Fatal("Tools config is nil")
	}

	if config.Tools.Timeout != 60*time.Second {
		t.Errorf("Expected tool timeout 60s, got %v", config.Tools.Timeout)
	}

	if config.Tools.MaxConcurrent != 10 {
		t.Errorf("Expected max concurrent tools 10, got %d", config.Tools.MaxConcurrent)
	}
}

// TestBuilderFactoryFunctions tests factory functions
func TestBuilderFactoryFunctions(t *testing.T) {
	testCases := []struct {
		name    string
		factory func(string, ...vnext.Option) (vnext.Agent, error)
		preset  vnext.PresetType
	}{
		{"ChatAgent", vnext.NewChatAgent, vnext.ChatAgent},
		{"ResearchAgent", vnext.NewResearchAgent, vnext.ResearchAgent},
		{"DataAgent", vnext.NewDataAgent, vnext.DataAgent},
		{"WorkflowAgent", vnext.NewWorkflowAgent, vnext.WorkflowAgent},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent, err := tc.factory("test-" + tc.name)
			if err != nil {
				t.Fatalf("Factory function failed: %v", err)
			}

			if agent.Name() != "test-"+tc.name {
				t.Errorf("Expected name 'test-%s', got '%s'", tc.name, agent.Name())
			}

			capabilities := agent.Capabilities()
			if len(capabilities) == 0 {
				t.Errorf("Agent has no capabilities")
			}
		})
	}
}

// TestBuilderCustomHandler tests custom handler functionality
func TestBuilderCustomHandler(t *testing.T) {
	customHandler := func(ctx context.Context, input string, capabilities *vnext.Capabilities) (string, error) {
		if input == "test" {
			return "custom response", nil
		}
		return "default response", nil
	}

	agent, err := vnext.NewBuilder("custom-handler-test").
		WithHandler(customHandler).
		Build()

	if err != nil {
		t.Fatalf("Build with custom handler failed: %v", err)
	}

	ctx := context.Background()

	// Test custom logic
	result, err := agent.Run(ctx, "test")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Content != "custom response" {
		t.Errorf("Expected 'custom response', got '%s'", result.Content)
	}

	// Test default logic
	result2, err := agent.Run(ctx, "other")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result2.Content != "default response" {
		t.Errorf("Expected 'default response', got '%s'", result2.Content)
	}
}

// TestBuilderRunOptions tests RunOptions functionality
func TestBuilderRunOptions(t *testing.T) {
	agent, err := vnext.NewBuilder("options-test").Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	ctx := context.Background()

	// Test with options
	opts := vnext.NewRunOptions().
		SetTools("calculator", "web_search").
		SetTimeout(30 * time.Second).
		SetDetailedResult(true)

	result, err := agent.RunWithOptions(ctx, "test input", opts)
	if err != nil {
		t.Fatalf("RunWithOptions failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful result")
	}

	// Check that detailed result metadata was added
	if result.Metadata == nil {
		t.Error("Expected metadata in detailed result")
	} else {
		if applied, ok := result.Metadata["options_applied"].(bool); !ok || !applied {
			t.Error("Expected options_applied to be true in metadata")
		}
	}
}

// TestBuilderValidation tests configuration validation
func TestBuilderValidation(t *testing.T) {
	// Test empty name validation
	builder := vnext.NewBuilder("")
	_, err := builder.Build()
	if err == nil {
		t.Error("Expected validation error for empty name")
	}

	// Test empty LLM provider validation
	builder2 := vnext.NewBuilder("test").WithConfig(&vnext.Config{
		Name: "test",
		LLM: vnext.LLMConfig{
			Provider: "", // Empty provider should fail
			Model:    "gpt-4",
		},
	})
	_, err = builder2.Build()
	if err == nil {
		t.Error("Expected validation error for empty LLM provider")
	}
}

// TestBuilderImmutability tests builder immutability after Build()
func TestBuilderImmutability(t *testing.T) {
	builder := vnext.NewBuilder("immutable-test")

	// Build the agent
	_, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Try to modify after build - should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when modifying frozen builder")
		}
	}()

	builder.WithPreset(vnext.ChatAgent) // This should panic
}
