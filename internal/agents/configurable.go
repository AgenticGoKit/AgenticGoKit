// Package agents provides internal agent implementations for AgentFlow.
package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// ConfigAwareUnifiedAgent extends UnifiedAgent with configuration awareness
type ConfigAwareUnifiedAgent struct {
	*core.UnifiedAgent
	config *core.ResolvedAgentConfig
}

// NewConfigAwareUnifiedAgent creates a new configuration-aware unified agent
func NewConfigAwareUnifiedAgent(name string, config *core.ResolvedAgentConfig, capabilities map[core.CapabilityType]core.AgentCapability, handler core.AgentHandler) *ConfigAwareUnifiedAgent {
	unifiedAgent := core.NewUnifiedAgent(name, capabilities, handler)

	return &ConfigAwareUnifiedAgent{
		UnifiedAgent: unifiedAgent,
		config:       config,
	}
}

// Run executes the agent with configuration-aware behavior
func (ca *ConfigAwareUnifiedAgent) Run(ctx context.Context, state core.State) (core.State, error) {
	// Check if agent is enabled
	if !ca.IsEnabled() {
		return state, fmt.Errorf("agent '%s' is disabled", ca.Name())
	}

	// Apply timeout from configuration
	if ca.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ca.config.Timeout)
		defer cancel()
	}

	// Add configuration metadata to state
	workingState := state.Clone()
	workingState.Set("agent_name", ca.Name())
	workingState.Set("agent_role", ca.GetRole())
	workingState.Set("agent_capabilities", ca.GetCapabilities())
	workingState.Set("agent_description", ca.GetDescription())

	// Apply system prompt if configured
	if ca.config.SystemPrompt != "" {
		var err error
		workingState, err = ca.ApplySystemPrompt(ctx, workingState)
		if err != nil {
			core.Logger().Error().
				Str("agent", ca.Name()).
				Err(err).
				Msg("Failed to apply system prompt")
			return state, fmt.Errorf("failed to apply system prompt: %w", err)
		}
	}

	// Execute the underlying unified agent
	result, err := ca.UnifiedAgent.Run(ctx, workingState)
	if err != nil {
		return result, err
	}

	// Add execution metadata
	result.Set("executed_by", ca.Name())
	result.Set("execution_role", ca.GetRole())
	result.Set("execution_timestamp", time.Now().Unix())

	return result, nil
}

// Configuration interface implementations
func (ca *ConfigAwareUnifiedAgent) GetRole() string {
	if ca.config != nil {
		return ca.config.Role
	}
	return ""
}

func (ca *ConfigAwareUnifiedAgent) GetDescription() string {
	if ca.config != nil {
		return ca.config.Description
	}
	return ""
}

func (ca *ConfigAwareUnifiedAgent) GetSystemPrompt() string {
	if ca.config != nil {
		return ca.config.SystemPrompt
	}
	return ""
}

func (ca *ConfigAwareUnifiedAgent) GetCapabilities() []string {
	if ca.config != nil {
		return ca.config.Capabilities
	}
	return []string{}
}

func (ca *ConfigAwareUnifiedAgent) IsEnabled() bool {
	if ca.config != nil {
		return ca.config.Enabled
	}
	return true // Default to enabled if no config
}

func (ca *ConfigAwareUnifiedAgent) GetTimeout() time.Duration {
	if ca.config != nil {
		return ca.config.Timeout
	}
	return 0
}

func (ca *ConfigAwareUnifiedAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	if ca.config != nil {
		return ca.config.LLMConfig
	}
	return nil
}

// UpdateConfiguration updates the agent's configuration
func (ca *ConfigAwareUnifiedAgent) UpdateConfiguration(config *core.ResolvedAgentConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	ca.config = config

	// Update LLM capability if present and LLM config is provided
	if config.LLMConfig != nil {
		if _, exists := ca.UnifiedAgent.GetCapability(core.CapabilityTypeLLM); exists {
			// Convert ResolvedLLMConfig to LLMConfig if agent supports it
			llmConfig := core.LLMConfig{
				Temperature:    config.LLMConfig.Temperature,
				MaxTokens:      config.LLMConfig.MaxTokens,
				TimeoutSeconds: int(config.LLMConfig.Timeout.Seconds()),
			}
			// Best-effort: if underlying agent supports SetLLMProvider via CapabilityConfigurable, it will be configured by builder
			core.Logger().Debug().
				Str("agent", ca.Name()).
				Str("provider", config.LLMConfig.Provider).
				Float64("temperature", config.LLMConfig.Temperature).
				Int("max_tokens", config.LLMConfig.MaxTokens).
				Msg("LLM configuration available for capability")
			_ = llmConfig
		}
	}

	core.Logger().Debug().
		Str("agent", ca.Name()).
		Str("role", config.Role).
		Bool("enabled", config.Enabled).
		Strs("capabilities", config.Capabilities).
		Msg("Agent configuration updated")

	return nil
}

// ApplySystemPrompt applies the configured system prompt to the state
func (ca *ConfigAwareUnifiedAgent) ApplySystemPrompt(ctx context.Context, state core.State) (core.State, error) {
	if ca.config == nil || ca.config.SystemPrompt == "" {
		return state, nil
	}

	workingState := state.Clone()
	workingState.Set("system_prompt", ca.config.SystemPrompt)
	workingState.Set("system_prompt_applied", true)

	// If we have an LLM capability, we can potentially use the system prompt
	if _, exists := ca.UnifiedAgent.GetCapability(core.CapabilityTypeLLM); exists {
		core.Logger().Debug().
			Str("agent", ca.Name()).
			Str("system_prompt", ca.config.SystemPrompt).
			Msg("System prompt prepared for LLM capability")
	}

	return workingState, nil
}

// ConfigAwareSequentialAgent extends SequentialAgent with configuration awareness
type ConfigAwareSequentialAgent struct {
	name   string
	agents []core.ConfigAwareAgent
	config *core.ResolvedAgentConfig
}

// NewConfigAwareSequentialAgent creates a new configuration-aware sequential agent
func NewConfigAwareSequentialAgent(name string, config *core.ResolvedAgentConfig, agents ...core.ConfigAwareAgent) *ConfigAwareSequentialAgent {
	// Filter out disabled agents
	enabledAgents := make([]core.ConfigAwareAgent, 0, len(agents))
	for _, agent := range agents {
		if agent != nil && agent.IsEnabled() {
			enabledAgents = append(enabledAgents, agent)
		} else if agent != nil {
			core.Logger().Debug().
				Str("sequential_agent", name).
				Str("disabled_agent", agent.Name()).
				Msg("Skipping disabled agent in sequential execution")
		}
	}

	return &ConfigAwareSequentialAgent{
		name:   name,
		agents: enabledAgents,
		config: config,
	}
}

// Name returns the agent's name
func (ca *ConfigAwareSequentialAgent) Name() string {
	return ca.name
}

// Run executes the sequence of enabled sub-agents
func (ca *ConfigAwareSequentialAgent) Run(ctx context.Context, initialState core.State) (core.State, error) {
	if !ca.IsEnabled() {
		return initialState, fmt.Errorf("sequential agent '%s' is disabled", ca.name)
	}

	// Apply timeout from configuration
	if ca.config != nil && ca.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ca.config.Timeout)
		defer cancel()
	}

	if len(ca.agents) == 0 {
		core.Logger().Warn().
			Str("sequential_agent", ca.name).
			Msg("No enabled sub-agents to run")
		return initialState, nil
	}

	nextState := initialState.Clone()
	nextState.Set("sequential_agent", ca.name)
	nextState.Set("sequential_agent_role", ca.GetRole())

	for i, agent := range ca.agents {
		select {
		case <-ctx.Done():
			core.Logger().Warn().
				Str("sequential_agent", ca.name).
				Int("agent_index", i).
				Msg("Context cancelled before running agent")
			return nextState, fmt.Errorf("sequential agent '%s': context cancelled: %w", ca.name, ctx.Err())
		default:
		}

		inputState := nextState.Clone()
		outputState, err := agent.Run(ctx, inputState)
		if err != nil {
			core.Logger().Error().
				Str("sequential_agent", ca.name).
				Int("agent_index", i).
				Str("sub_agent", agent.Name()).
				Err(err).
				Msg("Error in sequential sub-agent")
			return nextState, fmt.Errorf("sequential agent '%s': error in sub-agent %s: %w", ca.name, agent.Name(), err)
		}
		nextState = outputState
	}

	nextState.Set("sequential_execution_completed", true)
	return nextState, nil
}

// Configuration interface implementations for ConfigAwareSequentialAgent
func (ca *ConfigAwareSequentialAgent) GetRole() string {
	if ca.config != nil {
		return ca.config.Role
	}
	return "sequential_coordinator"
}

func (ca *ConfigAwareSequentialAgent) GetDescription() string {
	if ca.config != nil {
		return ca.config.Description
	}
	return "Executes agents sequentially"
}

func (ca *ConfigAwareSequentialAgent) GetSystemPrompt() string {
	if ca.config != nil {
		return ca.config.SystemPrompt
	}
	return ""
}

func (ca *ConfigAwareSequentialAgent) GetCapabilities() []string {
	if ca.config != nil {
		return ca.config.Capabilities
	}
	// Aggregate capabilities from sub-agents
	capSet := make(map[string]bool)
	for _, agent := range ca.agents {
		for _, cap := range agent.GetCapabilities() {
			capSet[cap] = true
		}
	}
	caps := make([]string, 0, len(capSet))
	for cap := range capSet {
		caps = append(caps, cap)
	}
	return caps
}

func (ca *ConfigAwareSequentialAgent) IsEnabled() bool {
	if ca.config != nil {
		return ca.config.Enabled
	}
	return true
}

func (ca *ConfigAwareSequentialAgent) GetTimeout() time.Duration {
	if ca.config != nil {
		return ca.config.Timeout
	}
	return 0
}

func (ca *ConfigAwareSequentialAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	if ca.config != nil {
		return ca.config.LLMConfig
	}
	return nil
}

func (ca *ConfigAwareSequentialAgent) UpdateConfiguration(config *core.ResolvedAgentConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}
	ca.config = config
	return nil
}

func (ca *ConfigAwareSequentialAgent) ApplySystemPrompt(ctx context.Context, state core.State) (core.State, error) {
	if ca.config == nil || ca.config.SystemPrompt == "" {
		return state, nil
	}

	workingState := state.Clone()
	workingState.Set("sequential_system_prompt", ca.config.SystemPrompt)
	return workingState, nil
}

// ConfigAwareParallelAgent extends ParallelAgent with configuration awareness
type ConfigAwareParallelAgent struct {
	name   string
	agents []core.ConfigAwareAgent
	config *core.ResolvedAgentConfig
}

// NewConfigAwareParallelAgent creates a new configuration-aware parallel agent
func NewConfigAwareParallelAgent(name string, config *core.ResolvedAgentConfig, agents ...core.ConfigAwareAgent) *ConfigAwareParallelAgent {
	// Filter out disabled agents
	enabledAgents := make([]core.ConfigAwareAgent, 0, len(agents))
	for _, agent := range agents {
		if agent != nil && agent.IsEnabled() {
			enabledAgents = append(enabledAgents, agent)
		} else if agent != nil {
			core.Logger().Debug().
				Str("parallel_agent", name).
				Str("disabled_agent", agent.Name()).
				Msg("Skipping disabled agent in parallel execution")
		}
	}

	return &ConfigAwareParallelAgent{
		name:   name,
		agents: enabledAgents,
		config: config,
	}
}

// Name returns the agent's name
func (ca *ConfigAwareParallelAgent) Name() string {
	return ca.name
}

// Run executes enabled sub-agents in parallel
func (ca *ConfigAwareParallelAgent) Run(ctx context.Context, initialState core.State) (core.State, error) {
	if !ca.IsEnabled() {
		return initialState, fmt.Errorf("parallel agent '%s' is disabled", ca.name)
	}

	// Apply timeout from configuration
	if ca.config != nil && ca.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ca.config.Timeout)
		defer cancel()
	}

	if len(ca.agents) == 0 {
		core.Logger().Warn().
			Str("parallel_agent", ca.name).
			Msg("No enabled sub-agents to run")
		return initialState, nil
	}

	// Implement parallel execution directly
	return ca.runParallelExecution(ctx, initialState)
}

// Configuration interface implementations for ConfigAwareParallelAgent
func (ca *ConfigAwareParallelAgent) GetRole() string {
	if ca.config != nil {
		return ca.config.Role
	}
	return "parallel_coordinator"
}

func (ca *ConfigAwareParallelAgent) GetDescription() string {
	if ca.config != nil {
		return ca.config.Description
	}
	return "Executes agents in parallel"
}

func (ca *ConfigAwareParallelAgent) GetSystemPrompt() string {
	if ca.config != nil {
		return ca.config.SystemPrompt
	}
	return ""
}

func (ca *ConfigAwareParallelAgent) GetCapabilities() []string {
	if ca.config != nil {
		return ca.config.Capabilities
	}
	// Aggregate capabilities from sub-agents
	capSet := make(map[string]bool)
	for _, agent := range ca.agents {
		for _, cap := range agent.GetCapabilities() {
			capSet[cap] = true
		}
	}
	caps := make([]string, 0, len(capSet))
	for cap := range capSet {
		caps = append(caps, cap)
	}
	return caps
}

func (ca *ConfigAwareParallelAgent) IsEnabled() bool {
	if ca.config != nil {
		return ca.config.Enabled
	}
	return true
}

func (ca *ConfigAwareParallelAgent) GetTimeout() time.Duration {
	if ca.config != nil {
		return ca.config.Timeout
	}
	return 0
}

func (ca *ConfigAwareParallelAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	if ca.config != nil {
		return ca.config.LLMConfig
	}
	return nil
}

func (ca *ConfigAwareParallelAgent) UpdateConfiguration(config *core.ResolvedAgentConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}
	ca.config = config
	return nil
}

func (ca *ConfigAwareParallelAgent) ApplySystemPrompt(ctx context.Context, state core.State) (core.State, error) {
	if ca.config == nil || ca.config.SystemPrompt == "" {
		return state, nil
	}

	workingState := state.Clone()
	workingState.Set("parallel_system_prompt", ca.config.SystemPrompt)
	return workingState, nil
}

// runParallelExecution implements parallel execution logic
func (ca *ConfigAwareParallelAgent) runParallelExecution(ctx context.Context, initialState core.State) (core.State, error) {
	// For now, implement a simple sequential execution as a placeholder
	// In a full implementation, this would use goroutines and channels like the internal ParallelAgent

	mergedState := initialState.Clone()
	mergedState.Set("parallel_agent", ca.name)
	mergedState.Set("parallel_agent_role", ca.GetRole())

	// Execute each agent and merge results
	for i, agent := range ca.agents {
		inputState := initialState.Clone()
		outputState, err := agent.Run(ctx, inputState)
		if err != nil {
			core.Logger().Error().
				Str("parallel_agent", ca.name).
				Int("agent_index", i).
				Str("sub_agent", agent.Name()).
				Err(err).
				Msg("Error in parallel sub-agent")
			return mergedState, fmt.Errorf("parallel agent '%s': error in sub-agent %s: %w", ca.name, agent.Name(), err)
		}

		// Merge the output state
		mergedState.Merge(outputState)
	}

	mergedState.Set("parallel_execution_completed", true)
	return mergedState, nil
}
