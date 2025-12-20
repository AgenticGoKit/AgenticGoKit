package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
	agenticgokit "github.com/agenticgokit/agenticgokit/internal/core"
)

// SequentialAgent runs a series of sub-agents one after another.
type SequentialAgent struct {
	name   string
	agents []agenticgokit.Agent
	config *core.ResolvedAgentConfig // Add configuration support
}

// Name returns the name of the sequential agent.
func (a *SequentialAgent) Name() string {
	return a.name
}

// NewSequentialAgent creates a new SequentialAgent.
// It filters out any nil agents provided in the list.
func NewSequentialAgent(name string, agents ...agenticgokit.Agent) *SequentialAgent {
	validAgents := make([]agenticgokit.Agent, 0, len(agents))
	for i, agent := range agents {
		if agent == nil {
			agenticgokit.Logger().Warn().
				Str("sequential_agent", name).
				Int("index", i).
				Msg("SequentialAgent: received a nil agent, skipping.")
			continue
		}
		validAgents = append(validAgents, agent)
	}
	return &SequentialAgent{
		agents: validAgents,
		name:   name,
		config: nil, // Configuration can be set later
	}
}

// NewSequentialAgentWithConfig creates a new SequentialAgent with configuration.
func NewSequentialAgentWithConfig(name string, config *core.ResolvedAgentConfig, agents ...agenticgokit.Agent) *SequentialAgent {
	validAgents := make([]agenticgokit.Agent, 0, len(agents))
	for i, agent := range agents {
		if agent == nil {
			agenticgokit.Logger().Warn().
				Str("sequential_agent", name).
				Int("index", i).
				Msg("SequentialAgent: received a nil agent, skipping.")
			continue
		}
		validAgents = append(validAgents, agent)
	}
	return &SequentialAgent{
		agents: validAgents,
		name:   name,
		config: config,
	}
}

// Run executes the sequence of sub-agents.
// It iterates through the configured agents, passing state sequentially.
// Execution halts immediately if a sub-agent returns an error or if the context is cancelled.
func (s *SequentialAgent) Run(ctx context.Context, initialState agenticgokit.State) (agenticgokit.State, error) {
	// Check if agent is enabled
	if !s.IsEnabled() {

		return initialState, nil
	}

	// Apply timeout from configuration
	if s.config != nil && s.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.config.Timeout)
		defer cancel()
	}

	if len(s.agents) == 0 {
		agenticgokit.Logger().Warn().
			Str("sequential_agent", s.name).
			Msg("SequentialAgent: No sub-agents to run.")
		return initialState, nil // Return input state if no agents
	}

	var err error
	nextState := initialState.Clone() // Start with a clone of the initial state

	// Add configuration metadata to state
	if s.config != nil {
		nextState.Set("sequential_agent_role", s.config.Role)
		nextState.Set("sequential_agent_description", s.config.Description)
		nextState.Set("sequential_agent_capabilities", s.config.Capabilities)

		// Apply system prompt if configured
		if s.config.SystemPrompt != "" {
			nextState.Set("sequential_system_prompt", s.config.SystemPrompt)
			nextState.Set("system_prompt_applied", true)
		}
	}

	for i, agent := range s.agents {
		// Check for context cancellation before running each sub-agent
		select {
		case <-ctx.Done():
			agenticgokit.Logger().Warn().
				Str("sequential_agent", s.name).
				Int("agent_index", i).
				Msg("SequentialAgent: Context cancelled before running agent.")
			return nextState, fmt.Errorf("SequentialAgent '%s': context cancelled: %w", s.name, ctx.Err())
		default:
			// Context is not cancelled, proceed
		}

		// It's crucial to clone the state before passing it to the next agent
		// to prevent unintended side effects if agents modify the state concurrently
		// or if the caller reuses the initial state.
		inputState := nextState.Clone()

		// Run the sub-agent
		outputState, agentErr := agent.Run(ctx, inputState)
		if agentErr != nil {
			err = fmt.Errorf("SequentialAgent '%s': error in sub-agent %d: %w", s.name, i, agentErr)
			agenticgokit.Logger().Error().
				Str("sequential_agent", s.name).
				Int("agent_index", i).
				Err(agentErr).
				Msg("SequentialAgent: Error in sub-agent.")
			// Return the state *before* the error occurred and the error itself
			return nextState, err
		}
		// Update the state for the next iteration
		nextState = outputState
	}

	// Add execution metadata
	nextState.Set("executed_by", s.name)
	if s.config != nil {
		nextState.Set("execution_role", s.config.Role)
	}
	nextState.Set("execution_timestamp", time.Now().Unix())
	nextState.Set("sequential_execution_completed", true)

	// Return the final state after all agents completed successfully
	return nextState, nil
}

// =============================================================================
// CONFIGURATION INTERFACE METHODS
// =============================================================================

// SetConfiguration sets the agent's configuration
func (s *SequentialAgent) SetConfiguration(config *core.ResolvedAgentConfig) {
	s.config = config
}

// GetConfiguration returns the agent's configuration
func (s *SequentialAgent) GetConfiguration() *core.ResolvedAgentConfig {
	return s.config
}

// GetRole returns the agent's configured role
func (s *SequentialAgent) GetRole() string {
	if s.config != nil {
		return s.config.Role
	}
	return "sequential_coordinator" // Default role
}

// GetDescription returns the agent's configured description
func (s *SequentialAgent) GetDescription() string {
	if s.config != nil {
		return s.config.Description
	}
	return "Executes agents sequentially" // Default description
}

// GetSystemPrompt returns the agent's configured system prompt
func (s *SequentialAgent) GetSystemPrompt() string {
	if s.config != nil {
		return s.config.SystemPrompt
	}
	return "" // No default system prompt
}

// GetCapabilities returns the agent's configured capabilities
func (s *SequentialAgent) GetCapabilities() []string {
	if s.config != nil {
		return s.config.Capabilities
	}
	return []string{"sequential_execution", "agent_coordination"} // Default capabilities
}

// IsEnabled returns whether the agent is enabled in configuration
func (s *SequentialAgent) IsEnabled() bool {
	if s.config != nil {
		return s.config.Enabled
	}
	return true // Default to enabled
}

// GetTimeout returns the agent's configured timeout
func (s *SequentialAgent) GetTimeout() time.Duration {
	if s.config != nil {
		return s.config.Timeout
	}
	return 0 // No timeout
}

// GetLLMConfig returns the agent's configured LLM settings
func (s *SequentialAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	if s.config != nil {
		return s.config.LLMConfig
	}
	return nil
}

// UpdateConfiguration updates the agent's configuration
func (s *SequentialAgent) UpdateConfiguration(config *core.ResolvedAgentConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}
	s.config = config

	agenticgokit.Logger().Debug().
		Str("agent", s.name).
		Str("role", config.Role).
		Bool("enabled", config.Enabled).
		Strs("capabilities", config.Capabilities).
		Msg("SequentialAgent configuration updated")

	return nil
}

// ApplySystemPrompt applies the configured system prompt to the state
func (s *SequentialAgent) ApplySystemPrompt(ctx context.Context, state agenticgokit.State) (agenticgokit.State, error) {
	if s.config == nil || s.config.SystemPrompt == "" {
		return state, nil
	}

	workingState := state.Clone()
	workingState.Set("system_prompt", s.config.SystemPrompt)
	workingState.Set("system_prompt_applied", true)
	workingState.Set("system_prompt_source", "sequential_agent")

	agenticgokit.Logger().Debug().
		Str("agent", s.name).
		Str("system_prompt", s.config.SystemPrompt).
		Msg("System prompt applied to state")

	return workingState, nil
}

