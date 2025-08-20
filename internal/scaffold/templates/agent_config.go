package templates

// ConfigDrivenAgentTemplate provides a template for configuration-driven agents
const ConfigDrivenAgentTemplate = `// Package agents contains reference implementations for custom agent types.
//
// IMPORTANT: This package is now optional. The main application uses
// ConfigurableAgentFactory to create agents directly from agentflow.toml.
//
// Use this package only if you need:
// - Custom agent types with specialized business logic
// - Reference implementations for learning
// - Migration examples from hardcoded to configuration-driven agents
package agents

import (
	"context"
	"fmt"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// Example{{.Agent.DisplayName}}Agent shows how to create a custom agent type
// that can be registered with the ConfigurableAgentFactory.
//
// This is a reference implementation - the main application uses
// configuration-driven agents created automatically from agentflow.toml.
type Example{{.Agent.DisplayName}}Agent struct {
	config core.ResolvedAgentConfig
	llm    core.ModelProvider
}

// NewExample{{.Agent.DisplayName}}Agent creates a configuration-aware agent instance.
//
// This shows how to create agents that use configuration instead of hardcoded values.
func NewExample{{.Agent.DisplayName}}Agent(config core.ResolvedAgentConfig) (*Example{{.Agent.DisplayName}}Agent, error) {
	// Initialize LLM provider from resolved configuration
	var llm core.ModelProvider
	if config.LLMConfig != nil {
		providerCfg := core.LLMProviderConfig{
			Type:        config.LLMConfig.Provider,
			Model:       config.LLMConfig.Model,
			Temperature: config.LLMConfig.Temperature,
			MaxTokens:   config.LLMConfig.MaxTokens,
			HTTPTimeout: config.LLMConfig.Timeout,
		}
		var err error
		llm, err = core.NewModelProviderFromConfig(providerCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize LLM provider: %w", err)
		}
	}

	return &Example{{.Agent.DisplayName}}Agent{
		config: config,
		llm:    llm,
	}, nil
}

// GetRole implements the Agent interface
func (a *Example{{.Agent.DisplayName}}Agent) GetRole() string {
	return a.config.Role
}

// GetCapabilities implements the Agent interface  
func (a *Example{{.Agent.DisplayName}}Agent) GetCapabilities() []string {
	return a.config.Capabilities
}

// IsEnabled implements the Agent interface
func (a *Example{{.Agent.DisplayName}}Agent) IsEnabled() bool {
	return a.config.Enabled
}

// Run implements the AgentHandler interface with configuration-driven behavior
func (a *Example{{.Agent.DisplayName}}Agent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	logger := core.Logger()
	logger.Debug().
		Str("agent", a.config.Role).
		Str("event_id", event.GetID()).
		Msg("Configuration-driven agent processing")

	// Check if agent is enabled
	if !a.config.Enabled {
		return core.AgentResult{
			OutputState: core.NewStateWithData(map[string]interface{}{
				"response": fmt.Sprintf("Agent %s is disabled", a.config.Role),
				"skipped":  true,
			}),
		}, nil
	}

	// Extract input message
	eventData := event.GetData()
	message, exists := eventData["message"]
	if !exists {
		return core.AgentResult{}, fmt.Errorf("no message found in event data")
	}

	messageStr, ok := message.(string)
	if !ok {
		return core.AgentResult{}, fmt.Errorf("message is not a string")
	}

	{{if .Config.MemoryEnabled}}
	// Use memory for context if available
	var contextInfo string
	mem := core.GetMemory(ctx)
	if mem != nil { // GetMemory never returns nil (NoOpMemory), but keep guard
		results, err := mem.Query(ctx, messageStr, 3)
		if err == nil && len(results) > 0 {
			contextInfo = fmt.Sprintf("\\n\\nRelevant context from memory:\\n%s", results[0].Content)
		}
	}
	{{end}}

	// Use system prompt from configuration
	systemPrompt := a.config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = fmt.Sprintf("You are %s, a helpful AI assistant.", a.config.Role)
	}

	// Build LLM prompt
	userPrompt := fmt.Sprintf("User: %s{{if .Config.MemoryEnabled}}%s{{end}}", messageStr{{if .Config.MemoryEnabled}}, contextInfo{{end}})
	prompt := core.Prompt{System: systemPrompt, User: userPrompt}

	// Generate response using configured LLM settings
	resp, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return core.AgentResult{}, fmt.Errorf("LLM generation failed: %w", err)
	}

	{{if .Config.MemoryEnabled}}
	// Store interaction in memory if available
	mem = core.GetMemory(ctx)
	if mem != nil {
		interactionContent := fmt.Sprintf("User: %s\\nAgent (%s): %s", messageStr, a.config.Role, resp.Content)
		if err := mem.Store(ctx, interactionContent, fmt.Sprintf("%s-interaction", a.config.Role)); err != nil {
			logger.Warn().Err(err).Msg("Failed to store interaction in memory")
		}
	}
	{{end}}

	// Return result with configuration-aware metadata
	return core.AgentResult{
		OutputState: core.NewStateWithData(map[string]interface{}{
			"response":     resp.Content,
			"agent_role":   a.config.Role,
			"capabilities": a.config.Capabilities,
			"message":      messageStr,
		}),
	}, nil
}

// RegisterCustomAgentTypes shows how to register custom agent types
// with the ConfigurableAgentFactory for use in configuration files.
//
// Call this function during application initialization to make custom
// agent types available in agentflow.toml configurations.
func RegisterCustomAgentTypes() error {
	// Example of registering a custom agent type
	// factory := core.GetConfigurableAgentFactory()
	// if factory != nil {
	//     factory.RegisterAgentType("example_{{.Agent.Name}}", func(config core.ResolvedAgentConfig) (core.Agent, error) {
	//         return NewExample{{.Agent.DisplayName}}Agent(config)
	//     })
	// }
	
	return nil
}
`
