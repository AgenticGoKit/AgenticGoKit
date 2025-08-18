package core

import (
	"fmt"
	"time"
)

// UnifiedAgentBuilder provides a fluent API to construct a UnifiedAgent.
type UnifiedAgentBuilder struct {
	name         string
	role         string
	description  string
	systemPrompt string
	timeout      time.Duration
	enabled      *bool

	handler      AgentHandler
	capabilities map[CapabilityType]AgentCapability

	llmProvider ModelProvider
	llmConfig   *AgentLLMConfig

	mcpManager MCPManager
	mcpConfig  *MCPAgentConfig
}

// NewUnifiedAgentBuilder creates a new builder with the required agent name.
func NewUnifiedAgentBuilder(name string) *UnifiedAgentBuilder {
	return &UnifiedAgentBuilder{
		name:         name,
		capabilities: make(map[CapabilityType]AgentCapability),
	}
}

// WithRole sets the agent role.
func (b *UnifiedAgentBuilder) WithRole(role string) *UnifiedAgentBuilder {
	b.role = role
	return b
}

// WithDescription sets a human-friendly description for the agent.
func (b *UnifiedAgentBuilder) WithDescription(desc string) *UnifiedAgentBuilder {
	b.description = desc
	return b
}

// WithSystemPrompt sets the system prompt for LLM interactions.
func (b *UnifiedAgentBuilder) WithSystemPrompt(prompt string) *UnifiedAgentBuilder {
	b.systemPrompt = prompt
	return b
}

// WithTimeout sets the execution timeout for the agent.
func (b *UnifiedAgentBuilder) WithTimeout(timeout time.Duration) *UnifiedAgentBuilder {
	b.timeout = timeout
	return b
}

// Enabled sets whether the agent is enabled.
func (b *UnifiedAgentBuilder) Enabled(enabled bool) *UnifiedAgentBuilder {
	b.enabled = &enabled
	return b
}

// WithHandler attaches a custom AgentHandler for execution logic.
func (b *UnifiedAgentBuilder) WithHandler(h AgentHandler) *UnifiedAgentBuilder {
	b.handler = h
	return b
}

// WithCapability adds a single capability to the agent.
func (b *UnifiedAgentBuilder) WithCapability(t CapabilityType, c AgentCapability) *UnifiedAgentBuilder {
	if b.capabilities == nil {
		b.capabilities = make(map[CapabilityType]AgentCapability)
	}
	b.capabilities[t] = c
	return b
}

// WithCapabilities merges multiple capabilities into the builder.
func (b *UnifiedAgentBuilder) WithCapabilities(caps map[CapabilityType]AgentCapability) *UnifiedAgentBuilder {
	if caps == nil {
		return b
	}
	if b.capabilities == nil {
		b.capabilities = make(map[CapabilityType]AgentCapability)
	}
	for k, v := range caps {
		b.capabilities[k] = v
	}
	return b
}

// WithLLMConfig sets the LLM configuration to be applied to the agent.
func (b *UnifiedAgentBuilder) WithLLMConfig(cfg AgentLLMConfig) *UnifiedAgentBuilder {
	// Copy to avoid external mutation after call
	c := cfg
	b.llmConfig = &c
	return b
}

// WithLLMProvider sets a pre-created ModelProvider.
func (b *UnifiedAgentBuilder) WithLLMProvider(p ModelProvider) *UnifiedAgentBuilder {
	b.llmProvider = p
	return b
}

// WithMCP wires an MCP manager and agent config for MCP capability use-cases.
func (b *UnifiedAgentBuilder) WithMCP(manager MCPManager, cfg MCPAgentConfig) *UnifiedAgentBuilder {
	b.mcpManager = manager
	c := cfg
	b.mcpConfig = &c
	return b
}

// Build constructs a UnifiedAgent, applying defaults and basic validation.
func (b *UnifiedAgentBuilder) Build() (Agent, error) {
	if b.name == "" {
		return nil, fmt.Errorf("agent name is required")
	}
	if b.timeout < 0 {
		return nil, fmt.Errorf("timeout cannot be negative")
	}

	ua := NewUnifiedAgent(b.name, b.capabilities, b.handler)

	// Apply provided fields or keep defaults from constructor
	if b.role != "" {
		ua.role = b.role
	}
	if b.description != "" {
		ua.description = b.description
	}
	if b.systemPrompt != "" {
		ua.systemPrompt = b.systemPrompt
	}
	if b.timeout > 0 {
		ua.timeout = b.timeout
	}
	if b.enabled != nil {
		ua.enabled = *b.enabled
	}

	// Apply LLM configuration if provided
	if b.llmConfig != nil {
		ua.SetLLMProvider(b.llmProvider, *b.llmConfig)
	}

	// Apply MCP wiring if provided
	if b.mcpManager != nil && b.mcpConfig != nil {
		ua.SetMCPManager(b.mcpManager, *b.mcpConfig)
	}

	return ua, nil
}
