package core

import (
	"context"
	"strings"
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
	autoLLM      bool // Controls whether to automatically call LLM when provider is configured

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
		autoLLM:      false, // Default to false for safety - user must explicitly enable
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

	// If an LLM provider is configured and auto-LLM is enabled, perform a default completion.
	// This gives UnifiedAgent a sensible behavior out-of-the-box for
	// configuration-driven agents created by the factory/builder.
	if u.llmProvider != nil && u.autoLLM {
		Logger().Debug().Str("agent", u.name).Msg("UnifiedAgent: LLM provider detected; preparing default completion")
		// Prefer a system prompt set in state (e.g., by a config-aware wrapper),
		// otherwise fall back to the agent's configured systemPrompt.
		system := u.systemPrompt
		if v, ok := state.Get("system_prompt"); ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				system = s
			}
		}

		// Extract user input from state. Orchestrators usually merge event data
		// into state before invoking the agent, so "message" is a common key.
		var user string
		if v, ok := state.Get("message"); ok {
			if s, ok2 := v.(string); ok2 {
				user = s
			}
		}

		// Only call the LLM if we have a non-empty user prompt.
		if strings.TrimSpace(user) != "" {
			Logger().Info().Str("agent", u.name).Msg("UnifiedAgent: calling LLM provider")
			params := ModelParameters{}
			if u.llmConfig != nil {
				if u.llmConfig.Temperature != 0 {
					// Convert float64 -> float32 pointer
					t := float32(u.llmConfig.Temperature)
					params.Temperature = &t
				}
				if u.llmConfig.MaxTokens > 0 {
					mt := int32(u.llmConfig.MaxTokens)
					params.MaxTokens = &mt
				}
			}

			resp, err := u.llmProvider.Call(ctx, Prompt{System: system, User: user, Parameters: params})
			if err != nil {
				return out, err
			}

			if resp.Content != "" {
				// Standardize keys so downstream result collectors can display output.
				out.Set("response", resp.Content)
				out.Set("message", resp.Content)
			}
		}
	} else {
		Logger().Warn().Str("agent", u.name).Msg("UnifiedAgent: no LLM provider configured; skipping LLM call")
	}

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
func (u *UnifiedAgent) SetLLMProvider(provider ModelProvider, config LLMConfig) {
	u.llmProvider = provider
	// Map to ResolvedLLMConfig-lite as available
	u.llmConfig = &ResolvedLLMConfig{
		Provider:         config.Provider,
		Model:            config.Model,
		Temperature:      config.Temperature,
		MaxTokens:        config.MaxTokens,
		Timeout:          TimeoutFromSeconds(config.TimeoutSeconds),
		TopP:             config.TopP,
		FrequencyPenalty: config.FrequencyPenalty,
		PresencePenalty:  config.PresencePenalty,
	}
}

func (u *UnifiedAgent) SetCacheManager(manager interface{}, config interface{}) {
	u.cacheMgr = manager
}

func (u *UnifiedAgent) SetMetricsConfig(config MetricsConfig) { u.metricsCfg = config }

// SetAutoLLM configures whether the agent should automatically call the LLM provider
// when one is configured. Set to true to enable automatic LLM calls, false to disable.
func (u *UnifiedAgent) SetAutoLLM(enabled bool) { u.autoLLM = enabled }

// GetAutoLLM returns whether automatic LLM calls are enabled.
func (u *UnifiedAgent) GetAutoLLM() bool { return u.autoLLM }

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
