package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// UnifiedAgent represents a production-ready agent that supports all capabilities
// through a composable, capability-based architecture.
type UnifiedAgent struct {
	name         string
	capabilities map[CapabilityType]AgentCapability
	handler      AgentHandler
	config       *ResolvedAgentConfig // Add configuration support
}

// NewUnifiedAgent creates a new unified agent with the given name and capabilities
func NewUnifiedAgent(name string, capabilities map[CapabilityType]AgentCapability, handler AgentHandler) *UnifiedAgent {
	if capabilities == nil {
		capabilities = make(map[CapabilityType]AgentCapability)
	}

	return &UnifiedAgent{
		name:         name,
		capabilities: capabilities,
		handler:      handler,
		config:       nil, // Configuration can be set later
	}
}

// NewUnifiedAgentWithConfig creates a new unified agent with configuration
func NewUnifiedAgentWithConfig(name string, config *ResolvedAgentConfig, capabilities map[CapabilityType]AgentCapability, handler AgentHandler) *UnifiedAgent {
	if capabilities == nil {
		capabilities = make(map[CapabilityType]AgentCapability)
	}

	return &UnifiedAgent{
		name:         name,
		capabilities: capabilities,
		handler:      handler,
		config:       config,
	}
}

// Name returns the agent's name
func (u *UnifiedAgent) Name() string {
	return u.name
}

// Run executes the agent with the given state and context
func (u *UnifiedAgent) Run(ctx context.Context, state State) (State, error) {
	log.Debug().
		Str("agent", u.name).
		Int("capabilities", len(u.capabilities)).
		Msg("UnifiedAgent starting execution")

	// Clone the input state to avoid mutations
	workingState := state.Clone()

	// Pre-execution: Apply capability pre-processing
	var err error
	workingState, err = u.applyCapabilityPreProcessing(ctx, workingState)
	if err != nil {
		log.Error().
			Err(err).
			Str("agent", u.name).
			Msg("Capability pre-processing failed")
		return state, fmt.Errorf("capability pre-processing failed: %w", err)
	}
	// Execute the core agent logic
	var result State
	if u.handler != nil {
		// Use custom handler if provided
		agentResult, err := u.handler.Run(ctx, NewEvent(u.name, nil, nil), workingState)
		if err != nil {
			log.Error().
				Err(err).
				Str("agent", u.name).
				Msg("Agent handler execution failed")
			return state, fmt.Errorf("agent handler execution failed: %w", err)
		}
		result = agentResult.OutputState
	} else {
		// Default behavior: add processed metadata
		result = workingState.Clone()
		result.Set("processed_by", u.name)

		// Add capability metadata
		capabilityTypes := make([]string, 0, len(u.capabilities))
		for capType := range u.capabilities {
			capabilityTypes = append(capabilityTypes, string(capType))
		}
		result.Set("capabilities", capabilityTypes)
	}

	// Post-execution: Apply capability post-processing
	finalState, err := u.applyCapabilityPostProcessing(ctx, result)
	if err != nil {
		log.Error().
			Err(err).
			Str("agent", u.name).
			Msg("Capability post-processing failed")
		return state, fmt.Errorf("capability post-processing failed: %w", err)
	}

	log.Debug().
		Str("agent", u.name).
		Msg("UnifiedAgent execution completed successfully")

	return finalState, nil
}

// GetCapability returns the capability of the specified type, if present
func (u *UnifiedAgent) GetCapability(capType CapabilityType) (AgentCapability, bool) {
	cap, exists := u.capabilities[capType]
	return cap, exists
}

// HasCapability checks if the agent has a capability of the specified type
func (u *UnifiedAgent) HasCapability(capType CapabilityType) bool {
	_, exists := u.capabilities[capType]
	return exists
}

// ListCapabilities returns a list of all capability types this agent has
func (u *UnifiedAgent) ListCapabilities() []CapabilityType {
	types := make([]CapabilityType, 0, len(u.capabilities))
	for capType := range u.capabilities {
		types = append(types, capType)
	}
	return types
}

// Configure implements CapabilityConfigurable for runtime configuration
func (u *UnifiedAgent) Configure(configs map[CapabilityType]interface{}) error {
	for capType, config := range configs {
		if capability, exists := u.capabilities[capType]; exists {
			// Apply capability-specific configuration based on type
			switch capType {
			case CapabilityTypeLLM:
				if llm, ok := capability.(*LLMCapability); ok {
					if llmConfig, ok := config.(LLMConfig); ok {
						llm.Config = llmConfig
						log.Debug().
							Str("agent", u.name).
							Str("capability", string(capType)).
							Msg("LLM capability configured")
					}
				}
			case CapabilityTypeMetrics:
				if metrics, ok := capability.(*MetricsCapability); ok {
					if metricsConfig, ok := config.(MetricsConfig); ok {
						metrics.Config = metricsConfig
						log.Debug().
							Str("agent", u.name).
							Str("capability", string(capType)).
							Msg("Metrics capability configured")
					}
				}
			default:
				log.Debug().
					Str("agent", u.name).
					Str("capability", string(capType)).
					Msg("Configuration not implemented for capability type")
			}
		}
	}
	return nil
}

// applyCapabilityPreProcessing applies pre-processing logic from all capabilities
func (u *UnifiedAgent) applyCapabilityPreProcessing(ctx context.Context, state State) (State, error) {
	workingState := state
	// Apply LLM capability pre-processing
	if llmCap, exists := u.capabilities[CapabilityTypeLLM]; exists {
		if llm, ok := llmCap.(*LLMCapability); ok && llm.Provider != nil {
			// Apply LLM-specific pre-processing (e.g., prompt preparation)
			log.Debug().
				Str("agent", u.name).
				Msg("Applying LLM pre-processing")
		}
	}

	// Apply Cache capability pre-processing
	if cacheCap, exists := u.capabilities[CapabilityTypeCache]; exists {
		if cache, ok := cacheCap.(*CacheCapability); ok && cache.Manager != nil {
			// Check cache for existing results
			log.Debug().
				Str("agent", u.name).
				Msg("Applying Cache pre-processing")
		}
	}
	// Apply Metrics capability pre-processing
	if metricsCap, exists := u.capabilities[CapabilityTypeMetrics]; exists {
		if metrics, ok := metricsCap.(*MetricsCapability); ok {
			// Start metrics collection
			log.Debug().
				Str("agent", u.name).
				Msg("Applying Metrics pre-processing")
			workingState.Set("metrics_enabled", metrics.Config.Enabled)
		}
	}

	// Apply MCP capability pre-processing
	if mcpCap, exists := u.capabilities[CapabilityTypeMCP]; exists {
		if mcp, ok := mcpCap.(*MCPCapability); ok {
			// Apply MCP-specific pre-processing
			log.Debug().
				Str("agent", u.name).
				Msg("Applying MCP pre-processing")
			if mcp.Manager != nil {
				workingState.Set("mcp_enabled", true)
			}
		}
	}

	return workingState, nil
}

// applyCapabilityPostProcessing applies post-processing logic from all capabilities
func (u *UnifiedAgent) applyCapabilityPostProcessing(ctx context.Context, state State) (State, error) {
	workingState := state

	// Apply Metrics capability post-processing
	if metricsCap, exists := u.capabilities[CapabilityTypeMetrics]; exists {
		if _, ok := metricsCap.(*MetricsCapability); ok {
			// Record metrics
			log.Debug().
				Str("agent", u.name).
				Msg("Applying Metrics post-processing")
			workingState.Set("metrics_collected", true)
		}
	}

	// Apply Cache capability post-processing
	if cacheCap, exists := u.capabilities[CapabilityTypeCache]; exists {
		if cache, ok := cacheCap.(*CacheCapability); ok && cache.Manager != nil {
			// Store results in cache
			log.Debug().
				Str("agent", u.name).
				Msg("Applying Cache post-processing")
			workingState.Set("cache_updated", true)
		}
	}

	// Apply LLM capability post-processing
	if llmCap, exists := u.capabilities[CapabilityTypeLLM]; exists {
		if llm, ok := llmCap.(*LLMCapability); ok && llm.Provider != nil {
			// Apply LLM-specific post-processing
			log.Debug().
				Str("agent", u.name).
				Msg("Applying LLM post-processing")
		}
	}

	// Apply MCP capability post-processing
	if mcpCap, exists := u.capabilities[CapabilityTypeMCP]; exists {
		if mcp, ok := mcpCap.(*MCPCapability); ok && mcp.Manager != nil {
			// Apply MCP-specific post-processing
			log.Debug().
				Str("agent", u.name).
				Msg("Applying MCP post-processing")
		}
	}

	return workingState, nil
}

// String returns a string representation of the agent
func (u *UnifiedAgent) String() string {
	capTypes := make([]string, 0, len(u.capabilities))
	for capType := range u.capabilities {
		capTypes = append(capTypes, string(capType))
	}
	return fmt.Sprintf("UnifiedAgent{name=%s, capabilities=%v}", u.name, capTypes)
}

// =============================================================================
// CAPABILITY CONFIGURABLE INTERFACE IMPLEMENTATION
// =============================================================================

// SetLLMProvider sets the LLM provider for the agent
func (u *UnifiedAgent) SetLLMProvider(provider ModelProvider, config LLMConfig) {
	if llmCap, exists := u.capabilities[CapabilityTypeLLM]; exists {
		if llm, ok := llmCap.(*LLMCapability); ok {
			llm.Provider = provider
			llm.Config = config
			log.Debug().
				Str("agent", u.name).
				Msg("LLM provider configured")
		}
	} else {
		// Create a new LLM capability if it doesn't exist
		llmCap := NewLLMCapability(provider, config)
		u.capabilities[CapabilityTypeLLM] = llmCap
		log.Debug().
			Str("agent", u.name).
			Msg("LLM capability created and configured")
	}
}

// SetCacheManager sets the cache manager for the agent
func (u *UnifiedAgent) SetCacheManager(manager interface{}, config interface{}) {
	if cacheCap, exists := u.capabilities[CapabilityTypeCache]; exists {
		if cache, ok := cacheCap.(*CacheCapability); ok {
			cache.Manager = manager
			log.Debug().
				Str("agent", u.name).
				Msg("Cache manager configured")
		}
	} else { // Create a new cache capability if it doesn't exist
		cacheCap := NewCacheCapability(manager, config)
		u.capabilities[CapabilityTypeCache] = cacheCap
		log.Debug().
			Str("agent", u.name).
			Msg("Cache capability created and configured")
	}
}

// SetMetricsConfig sets the metrics configuration for the agent
func (u *UnifiedAgent) SetMetricsConfig(config MetricsConfig) {
	if metricsCap, exists := u.capabilities[CapabilityTypeMetrics]; exists {
		if metrics, ok := metricsCap.(*MetricsCapability); ok {
			metrics.Config = config
			log.Debug().
				Str("agent", u.name).
				Msg("Metrics configuration updated")
		}
	} else {
		// Create a new metrics capability if it doesn't exist
		metricsCap := NewMetricsCapability(config)
		u.capabilities[CapabilityTypeMetrics] = metricsCap
		log.Debug().
			Str("agent", u.name).
			Msg("Metrics capability created and configured")
	}
}

// GetLogger returns the agent's logger for capability configuration
func (u *UnifiedAgent) GetLogger() *zerolog.Logger {
	logger := log.With().Str("agent", u.name).Logger()
	return &logger
}

// =============================================================================
// CONFIGURATION-AWARE METHODS
// =============================================================================

// SetConfiguration sets the agent's configuration
func (u *UnifiedAgent) SetConfiguration(config *ResolvedAgentConfig) {
	u.config = config
	
	// Update LLM capability with configuration if present
	if config != nil && config.LLMConfig != nil {
		if llmCap, exists := u.capabilities[CapabilityTypeLLM]; exists {
			if llm, ok := llmCap.(*LLMCapability); ok {
				llmConfig := LLMConfig{
					Temperature:    config.LLMConfig.Temperature,
					MaxTokens:      config.LLMConfig.MaxTokens,
					SystemPrompt:   config.SystemPrompt,
					TimeoutSeconds: int(config.LLMConfig.Timeout.Seconds()),
				}
				llm.Config = llmConfig
				
				Logger().Debug().
					Str("agent", u.name).
					Str("provider", config.LLMConfig.Provider).
					Float64("temperature", config.LLMConfig.Temperature).
					Int("max_tokens", config.LLMConfig.MaxTokens).
					Msg("Updated LLM configuration from agent config")
			}
		}
	}
}

// GetConfiguration returns the agent's configuration
func (u *UnifiedAgent) GetConfiguration() *ResolvedAgentConfig {
	return u.config
}

// GetRole returns the agent's configured role
func (u *UnifiedAgent) GetRole() string {
	if u.config != nil {
		return u.config.Role
	}
	return u.name + "_agent" // Default role
}

// GetDescription returns the agent's configured description
func (u *UnifiedAgent) GetDescription() string {
	if u.config != nil {
		return u.config.Description
	}
	return "Agent " + u.name // Default description
}

// GetSystemPrompt returns the agent's configured system prompt
func (u *UnifiedAgent) GetSystemPrompt() string {
	if u.config != nil {
		return u.config.SystemPrompt
	}
	return "" // No default system prompt
}

// GetConfiguredCapabilities returns the agent's configured capabilities
func (u *UnifiedAgent) GetConfiguredCapabilities() []string {
	if u.config != nil {
		return u.config.Capabilities
	}
	return []string{} // No configured capabilities
}

// IsConfigurationEnabled returns whether the agent is enabled in configuration
func (u *UnifiedAgent) IsConfigurationEnabled() bool {
	if u.config != nil {
		return u.config.Enabled
	}
	return true // Default to enabled
}

// GetConfiguredTimeout returns the agent's configured timeout
func (u *UnifiedAgent) GetConfiguredTimeout() time.Duration {
	if u.config != nil {
		return u.config.Timeout
	}
	return 0 // No timeout
}

// GetConfiguredLLMConfig returns the agent's configured LLM settings
func (u *UnifiedAgent) GetConfiguredLLMConfig() *ResolvedLLMConfig {
	if u.config != nil {
		return u.config.LLMConfig
	}
	return nil
}
