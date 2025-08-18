package core

import (
	"context"
	"time"
)

// UnifiedAgent is a concrete Agent implementation used by internal builders.
// It provides basic capability plumbing and a simple Run implementation.
type UnifiedAgent struct {
	name         string
	role         string
	description  string
	systemPrompt string
	timeout      time.Duration
	enabled      bool

	// Config/LLM
	llmConfig *ResolvedLLMConfig

	// Capabilities storage
	capabilities map[CapabilityType]AgentCapability
	handler      AgentHandler

	// Capability-configured dependencies
	llmProvider ModelProvider
	cacheMgr    interface{}
	metricsCfg  MetricsConfig

	// MCP-specific wiring to satisfy MCP capability
	mcpManager     MCPManager
	mcpAgentConfig MCPAgentConfig
	mcpCacheMgr    MCPCacheManager
}

// NewUnifiedAgent constructs a new UnifiedAgent with provided capabilities and optional handler.
func NewUnifiedAgent(name string, caps map[CapabilityType]AgentCapability, handler AgentHandler) *UnifiedAgent {
	if caps == nil {
		caps = map[CapabilityType]AgentCapability{}
	}
	return &UnifiedAgent{
		name:         name,
		role:         "unified_agent",
		description:  "Unified composable agent",
		systemPrompt: "You are a helpful AI agent.",
		timeout:      30 * time.Second,
		enabled:      true,
		capabilities: caps,
		handler:      handler,
	}
}

// Agent interface
func (u *UnifiedAgent) Name() string                     { return u.name }
func (u *UnifiedAgent) GetRole() string                  { return u.role }
func (u *UnifiedAgent) GetDescription() string           { return u.description }
func (u *UnifiedAgent) GetSystemPrompt() string          { return u.systemPrompt }
func (u *UnifiedAgent) GetTimeout() time.Duration        { return u.timeout }
func (u *UnifiedAgent) IsEnabled() bool                  { return u.enabled }
func (u *UnifiedAgent) GetLLMConfig() *ResolvedLLMConfig { return u.llmConfig }

func (u *UnifiedAgent) GetCapabilities() []string {
	out := make([]string, 0, len(u.capabilities))
	for ct := range u.capabilities {
		out = append(out, string(ct))
	}
	return out
}

func (u *UnifiedAgent) Initialize(ctx context.Context) error { return nil }
func (u *UnifiedAgent) Shutdown(ctx context.Context) error   { return nil }

func (u *UnifiedAgent) Run(ctx context.Context, state State) (State, error) {
	// Basic pre/post without complex hooks for now
	if u.handler != nil {
		res, err := u.handler.Run(ctx, NewEvent(u.name, map[string]any{}, map[string]string{}), state)
		if err != nil {
			return res.OutputState, err
		}
		return res.OutputState, nil
	}
	out := state.Clone()
	out.Set("processed_by", u.name)
	out.Set("agent_type", "unified")
	out.Set("capabilities", u.GetCapabilities())
	return out, nil
}

func (u *UnifiedAgent) HandleEvent(ctx context.Context, event Event, state State) (AgentResult, error) {
	start := time.Now()
	out, err := u.Run(ctx, state)
	end := time.Now()
	result := AgentResult{OutputState: out, StartTime: start, EndTime: end, Duration: end.Sub(start)}
	if err != nil {
		result.Error = err.Error()
	}
	return result, nil
}

// CapabilityConfigurable bridging (subset used by internal capabilities)
func (u *UnifiedAgent) SetLLMProvider(provider ModelProvider, config AgentLLMConfig) {
	u.llmProvider = provider
	// Map to ResolvedLLMConfig-lite as available
	u.llmConfig = &ResolvedLLMConfig{
		Provider:         config.Provider,
		Model:            config.Model,
		Temperature:      config.Temperature,
		MaxTokens:        config.MaxTokens,
		Timeout:          time.Duration(config.TimeoutSeconds) * time.Second,
		TopP:             config.TopP,
		FrequencyPenalty: config.FrequencyPenalty,
		PresencePenalty:  config.PresencePenalty,
	}
}

func (u *UnifiedAgent) SetCacheManager(manager interface{}, config interface{}) {
	u.cacheMgr = manager
}

func (u *UnifiedAgent) SetMetricsConfig(config MetricsConfig) { u.metricsCfg = config }

// Logger accessor to satisfy internal CapabilityConfigurable usage through core.Logger
func (u *UnifiedAgent) GetLogger() CoreLogger { return Logger() }

// MCP wiring to satisfy MCP capability Configure calls
func (u *UnifiedAgent) SetMCPManager(manager MCPManager, config MCPAgentConfig) {
	u.mcpManager = manager
	u.mcpAgentConfig = config
}

func (u *UnifiedAgent) SetMCPCacheManager(manager MCPCacheManager) { u.mcpCacheMgr = manager }

// GetCapability returns a capability by type if present (helper for internal bridges)
func (u *UnifiedAgent) GetCapability(t CapabilityType) (AgentCapability, bool) {
	if u.capabilities == nil {
		return nil, false
	}
	cap, ok := u.capabilities[t]
	return cap, ok
}
