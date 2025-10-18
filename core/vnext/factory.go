// Package vnext provides the next-generation Agent API for AgenticGoKit
package vnext

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// FACTORY.GO - Unified Factory and Initialization System
// =============================================================================
//
// This file provides:
// - Agent preset registry and factories
// - QuickStart functions for common scenarios
// - Global initialization and configuration helpers
//
// Factory patterns:
// - Preset registry for pre-configured agent types
// - QuickStart functions using NewBuilder() for minimal setup
// - Integration with existing factory systems (Memory, Tools, Workflow)
//
// =============================================================================

// =============================================================================
// AGENT PRESET REGISTRY
// =============================================================================

// AgentPreset represents a pre-configured agent type
//
// Presets provide named configurations for common agent patterns.
// Each preset includes a builder function that creates a fully configured agent.
type AgentPreset struct {
	Name        string                       // Preset name (e.g., "chat", "research")
	Description string                       // Human-readable description
	Builder     func(*Config) (Agent, error) // Function to build the agent
}

var (
	agentPresets = make(map[string]*AgentPreset)
	presetMutex  sync.RWMutex
)

// RegisterAgentPreset registers a named agent preset
//
// Presets allow you to define reusable agent configurations with custom names.
//
// Example:
//
//	RegisterAgentPreset(&AgentPreset{
//	    Name: "my-agent",
//	    Description: "Custom agent",
//	    Builder: func(cfg *Config) (Agent, error) {
//	        return NewBuilder("my-agent").WithConfig(cfg).Build()
//	    },
//	})
func RegisterAgentPreset(preset *AgentPreset) error {
	if preset == nil {
		err := ConfigError("preset", "preset", nil)
		err.Details["message"] = "Agent preset cannot be nil"
		return err
	}
	if preset.Name == "" {
		err := ConfigError("preset", "name", nil)
		err.Details["message"] = "Agent preset name is required"
		return err
	}
	if preset.Builder == nil {
		err := ConfigError("preset", "builder", nil)
		err.Details["message"] = "Agent preset builder function is required"
		return err
	}

	presetMutex.Lock()
	defer presetMutex.Unlock()
	agentPresets[preset.Name] = preset
	return nil
}

// GetAgentPreset retrieves a registered agent preset by name
//
// Returns nil if the preset is not found.
func GetAgentPreset(name string) *AgentPreset {
	presetMutex.RLock()
	defer presetMutex.RUnlock()
	return agentPresets[name]
}

// ListAgentPresets returns all registered preset names
func ListAgentPresets() []string {
	presetMutex.RLock()
	defer presetMutex.RUnlock()

	names := make([]string, 0, len(agentPresets))
	for name := range agentPresets {
		names = append(names, name)
	}
	return names
}

// =============================================================================
// QUICKSTART FUNCTIONS
// =============================================================================

// QuickChatAgent creates a chat agent with minimal configuration
//
// Example:
//
//	agent, err := QuickChatAgent("gpt-4")
//	result, _ := agent.Run(ctx, "Hello!")
func QuickChatAgent(model string) (Agent, error) {
	return QuickChatAgentWithConfig(model, nil)
}

// QuickChatAgentWithConfig creates a chat agent with custom configuration
//
// Example:
//
//	config := &Config{
//	    LLM: LLMConfig{Model: "gpt-4"},
//	}
//	agent, err := QuickChatAgentWithConfig("gpt-4", config)
func QuickChatAgentWithConfig(model string, config *Config) (Agent, error) {
	if config == nil {
		config = &Config{
			Name:         "chat-agent",
			SystemPrompt: "You are a helpful assistant",
			Timeout:      30 * time.Second,
		}
	}

	// Ensure LLM config is set
	if config.LLM.Model == "" {
		config.LLM.Model = model
	}
	if config.LLM.Provider == "" {
		config.LLM.Provider = "openai"
	}

	// Use the builder with ChatAgent preset
	return NewBuilder(config.Name).
		WithConfig(config).
		WithPreset(ChatAgent).
		Build()
}

// QuickResearchAgent creates a research agent with minimal configuration
//
// Example:
//
//	agent, err := QuickResearchAgent("gpt-4")
//	result, _ := agent.Run(ctx, "Research quantum computing")
func QuickResearchAgent(model string) (Agent, error) {
	return QuickResearchAgentWithConfig(model, nil)
}

// QuickResearchAgentWithConfig creates a research agent with custom configuration
//
// Example:
//
//	config := &Config{
//	    LLM: LLMConfig{Model: "gpt-4"},
//	    Memory: &MemoryConfig{Provider: "postgres"},
//	}
//	agent, err := QuickResearchAgentWithConfig("gpt-4", config)
func QuickResearchAgentWithConfig(model string, config *Config) (Agent, error) {
	if config == nil {
		config = &Config{
			Name:         "research-agent",
			SystemPrompt: "You are a research assistant",
			Timeout:      60 * time.Second,
		}
	}

	// Ensure LLM config is set
	if config.LLM.Model == "" {
		config.LLM.Model = model
	}
	if config.LLM.Provider == "" {
		config.LLM.Provider = "openai"
	}

	// Use the builder with ResearchAgent preset
	return NewBuilder(config.Name).
		WithConfig(config).
		WithPreset(ResearchAgent).
		Build()
}

// QuickWorkflow creates a simple sequential workflow
//
// Example:
//
//	agent1, _ := QuickChatAgent("gpt-4")
//	agent2, _ := QuickResearchAgent("gpt-4")
//	workflow, _ := QuickWorkflow([]Agent{agent1, agent2})
func QuickWorkflow(agents []Agent) (Workflow, error) {
	if len(agents) == 0 {
		err := WorkflowError("validation", "agents", nil)
		err.Details["message"] = "at least one agent is required for workflow"
		return nil, err
	}

	config := &WorkflowConfig{
		Mode:    Sequential,
		Timeout: 5 * time.Minute,
	}

	workflow, err := NewSequentialWorkflow(config)
	if err != nil {
		return nil, err
	}

	// Add agents as steps
	for i, agent := range agents {
		step := WorkflowStep{
			Name:  fmt.Sprintf("step-%d", i+1),
			Agent: agent,
		}
		if err := workflow.AddStep(step); err != nil {
			return nil, err
		}
	}

	return workflow, nil
}

// QuickParallelWorkflow creates a parallel workflow
//
// Example:
//
//	agents := []Agent{agent1, agent2}
//	workflow, _ := QuickParallelWorkflow(agents)
func QuickParallelWorkflow(agents []Agent) (Workflow, error) {
	if len(agents) == 0 {
		err := WorkflowError("validation", "agents", nil)
		err.Details["message"] = "at least one agent is required for workflow"
		return nil, err
	}

	config := &WorkflowConfig{
		Mode:    Parallel,
		Timeout: 5 * time.Minute,
	}

	workflow, err := NewParallelWorkflow(config)
	if err != nil {
		return nil, err
	}

	// Add agents as steps
	for i, agent := range agents {
		step := WorkflowStep{
			Name:  fmt.Sprintf("step-%d", i+1),
			Agent: agent,
		}
		if err := workflow.AddStep(step); err != nil {
			return nil, err
		}
	}

	return workflow, nil
}

// QuickAgentFromPreset creates an agent using a registered preset
//
// Example:
//
//	agent, err := QuickAgentFromPreset("my-custom-preset", config)
func QuickAgentFromPreset(presetName string, config *Config) (Agent, error) {
	preset := GetAgentPreset(presetName)
	if preset == nil {
		err := ConfigError("preset", "name", nil)
		err.Details["message"] = fmt.Sprintf("Agent preset '%s' not found. Use RegisterAgentPreset() first", presetName)
		return nil, err
	}

	return preset.Builder(config)
}

// =============================================================================
// INITIALIZATION AND SETUP
// =============================================================================

var (
	initOnce      sync.Once
	isInitialized bool
	initError     error
	initMutex     sync.RWMutex
)

// InitializeDefaults initializes the vNext API with built-in presets
//
// This should be called once at application startup.
// Calling multiple times is safe - it only initializes once.
//
// Example:
//
//	func main() {
//	    if err := vnext.InitializeDefaults(); err != nil {
//	        log.Fatal(err)
//	    }
//	    agent, _ := vnext.QuickChatAgent("gpt-4")
//	}
func InitializeDefaults() error {
	initOnce.Do(func() {
		// Register built-in presets
		if err := registerBuiltinPresets(); err != nil {
			initError = err
			return
		}

		// Mark as initialized
		initMutex.Lock()
		isInitialized = true
		initMutex.Unlock()
	})

	return initError
}

// IsInitialized returns whether InitializeDefaults has been called
func IsInitialized() bool {
	initMutex.RLock()
	defer initMutex.RUnlock()
	return isInitialized
}

// registerBuiltinPresets registers the standard agent presets
func registerBuiltinPresets() error {
	// Register chat preset
	chatPreset := &AgentPreset{
		Name:        "chat",
		Description: "Conversational agent optimized for interactive chat",
		Builder: func(cfg *Config) (Agent, error) {
			builder := NewBuilder("chat-agent").WithPreset(ChatAgent)
			if cfg != nil {
				builder = builder.WithConfig(cfg)
			}
			return builder.Build()
		},
	}
	if err := RegisterAgentPreset(chatPreset); err != nil {
		return err
	}

	// Register research preset
	researchPreset := &AgentPreset{
		Name:        "research",
		Description: "Research agent optimized for information gathering",
		Builder: func(cfg *Config) (Agent, error) {
			builder := NewBuilder("research-agent").WithPreset(ResearchAgent)
			if cfg != nil {
				builder = builder.WithConfig(cfg)
			}
			return builder.Build()
		},
	}
	if err := RegisterAgentPreset(researchPreset); err != nil {
		return err
	}

	return nil
}

// SetupLogging configures the logging level
//
// Valid levels: "debug", "info", "warn", "error"
func SetupLogging(level string) error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[level] {
		err := ConfigError("logging", "level", nil)
		err.Details["message"] = fmt.Sprintf("Invalid log level '%s'. Valid levels: debug, info, warn, error", level)
		return err
	}

	// TODO: Implement actual logging configuration
	return nil
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// Initialize initializes the vNext API with defaults (panics on error)
//
// This is a convenience wrapper around InitializeDefaults that panics on error.
// Use this for simple scripts where you want to fail fast.
//
// Example:
//
//	func main() {
//	    vnext.Initialize()  // panics if initialization fails
//	    agent, _ := vnext.QuickChatAgent("gpt-4")
//	}
func Initialize() {
	if err := InitializeDefaults(); err != nil {
		panic(fmt.Sprintf("Failed to initialize vNext API: %v", err))
	}
}

// Setup is an alias for Initialize for more user-friendly naming
func Setup() {
	Initialize()
}

// CreateAgent creates an agent with default configuration (panics on error)
//
// This is the simplest way to create an agent - just provide a model and prompt.
// For error handling, use QuickChatAgent instead.
//
// Example:
//
//	agent := vnext.CreateAgent("gpt-4", "You are helpful")
func CreateAgent(model string, systemPrompt string) Agent {
	config := &Config{
		Name:         "default-agent",
		SystemPrompt: systemPrompt,
		Timeout:      30 * time.Second,
		LLM: LLMConfig{
			Model:    model,
			Provider: "openai",
		},
	}

	agent, err := QuickChatAgentWithConfig(model, config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create agent: %v", err))
	}

	return agent
}

// Run is a convenience function for one-shot agent execution
//
// This is the simplest way to run an agent - creates, initializes, executes, and cleans up.
//
// Example:
//
//	result, err := vnext.Run(ctx, "gpt-4", "Hello!")
//	fmt.Println(result.Output)
//
// Note: Creates a new agent for each call. For production use, create an agent once and reuse it.
func Run(ctx context.Context, model string, input string) (*Result, error) {
	agent, err := QuickChatAgent(model)
	if err != nil {
		return nil, err
	}

	if err := agent.Initialize(ctx); err != nil {
		return nil, err
	}
	defer agent.Cleanup(ctx)

	return agent.Run(ctx, input)
}

// =============================================================================
// EXAMPLES
// =============================================================================
//
// EXAMPLE 1: Simplest Possible Usage
//
//	func main() {
//	    result, _ := vnext.Run(context.Background(), "gpt-4", "Hello!")
//	    fmt.Println(result.Output)
//	}
//
// EXAMPLE 2: Quick Start with Agent
//
//	func main() {
//	    vnext.Setup()  // or vnext.Initialize()
//	    agent, _ := vnext.QuickChatAgent("gpt-4")
//	    result, _ := agent.Run(context.Background(), "Hello!")
//	    fmt.Println(result.Output)
//	}
//
// EXAMPLE 3: Custom Configuration
//
//	func main() {
//	    vnext.Initialize()
//	    config := &vnext.Config{
//	        LLM: vnext.LLMConfig{Model: "gpt-4"},
//	        Memory: &vnext.MemoryConfig{Provider: "postgres"},
//	    }
//	    agent, _ := vnext.QuickResearchAgentWithConfig("gpt-4", config)
//	}
//
// EXAMPLE 4: Custom Agent Preset
//
//	func init() {
//	    vnext.RegisterAgentPreset(&vnext.AgentPreset{
//	        Name: "code-reviewer",
//	        Description: "Code review agent",
//	        Builder: func(cfg *vnext.Config) (vnext.Agent, error) {
//	            return vnext.NewBuilder("reviewer").
//	                WithConfig(cfg).
//	                WithSystemPrompt("You are a code reviewer").
//	                Build()
//	        },
//	    })
//	}
//
// EXAMPLE 5: Multi-Agent Workflow
//
//	func main() {
//	    vnext.Setup()
//	    researcher, _ := vnext.QuickResearchAgent("gpt-4")
//	    summarizer, _ := vnext.QuickChatAgent("gpt-4")
//	    workflow, _ := vnext.QuickWorkflow([]vnext.Agent{researcher, summarizer})
//	    result, _ := workflow.Execute(context.Background(), "Research AI")
//	}
