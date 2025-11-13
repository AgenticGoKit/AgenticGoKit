package agents

import (
	"context"
	"fmt"
	"time"

	agenticgokit "github.com/agenticgokit/agenticgokit/internal/core"
	"github.com/agenticgokit/agenticgokit/core"
)

// defaultMaxIterations is the default limit if LoopAgentConfig.MaxIterations is not set.
const defaultMaxIterations = 100

// ConditionFunc is a function type used by LoopAgent to determine if the loop should stop.
// It receives the current state *after* a sub-agent run and returns true to stop the loop,
// or false to continue (up to MaxIterations).
type ConditionFunc func(currentState agenticgokit.State) bool

// LoopAgentConfig holds configuration for the LoopAgent.
type LoopAgentConfig struct {
	Condition     ConditionFunc
	MaxIterations int
	Timeout       time.Duration
}

// LoopAgent repeatedly executes a sub-agent until a condition is met,
// max iterations are reached, or the context is cancelled.
type LoopAgent struct {
	name        string
	subAgent    agenticgokit.Agent
	config      LoopAgentConfig
	agentConfig *core.ResolvedAgentConfig // Add agent configuration support
}

// NewLoopAgent creates a new LoopAgent.
// It requires a non-nil subAgent to execute in the loop.
// It applies the default MaxIterations if the provided value is invalid.
func NewLoopAgent(name string, config LoopAgentConfig, subAgent agenticgokit.Agent) *LoopAgent {
	if subAgent == nil {
		agenticgokit.Logger().Error().
			Str("agent", name).
			Msg("LoopAgent requires a non-nil subAgent.")
		return nil // Cannot create a loop agent without a sub-agent
	}

	maxIter := config.MaxIterations
	if maxIter <= 0 {
		maxIter = defaultMaxIterations
		agenticgokit.Logger().Warn().
			Str("agent", name).
			Int("default_max_iterations", defaultMaxIterations).
			Msg("LoopAgent: MaxIterations not specified or invalid, defaulting to defaultMaxIterations.")
	}

	return &LoopAgent{
		subAgent: subAgent,
		config: LoopAgentConfig{
			Condition:     config.Condition,
			MaxIterations: maxIter,
			Timeout:       config.Timeout,
		},
		name:        name,
		agentConfig: nil, // Configuration can be set later
	}
}

// NewLoopAgentWithConfig creates a new LoopAgent with agent configuration.
func NewLoopAgentWithConfig(name string, config LoopAgentConfig, agentConfig *core.ResolvedAgentConfig, subAgent agenticgokit.Agent) *LoopAgent {
	if subAgent == nil {
		agenticgokit.Logger().Error().
			Str("agent", name).
			Msg("LoopAgent requires a non-nil subAgent.")
		return nil // Cannot create a loop agent without a sub-agent
	}

	maxIter := config.MaxIterations
	if maxIter <= 0 {
		maxIter = defaultMaxIterations
		agenticgokit.Logger().Warn().
			Str("agent", name).
			Int("default_max_iterations", defaultMaxIterations).
			Msg("LoopAgent: MaxIterations not specified or invalid, defaulting to defaultMaxIterations.")
	}

	return &LoopAgent{
		subAgent: subAgent,
		config: LoopAgentConfig{
			Condition:     config.Condition,
			MaxIterations: maxIter,
			Timeout:       config.Timeout,
		},
		name:        name,
		agentConfig: agentConfig,
	}
}

// Name returns the name of the loop agent.
func (a *LoopAgent) Name() string {
	return a.name
}

// Run executes the sub-agent in a loop according to the configuration.
func (l *LoopAgent) Run(ctx context.Context, initialState agenticgokit.State) (agenticgokit.State, error) {
	// Check if agent is enabled
	if !l.IsEnabled() {
		agenticgokit.Logger().Debug().
			Str("loop_agent", l.name).
			Msg("LoopAgent: Agent is disabled, skipping execution.")
		return initialState, nil
	}

	currentState := initialState.Clone()
	var err error
	iteration := 0

	// Add configuration metadata to state
	if l.agentConfig != nil {
		currentState.Set("loop_agent_role", l.agentConfig.Role)
		currentState.Set("loop_agent_description", l.agentConfig.Description)
		currentState.Set("loop_agent_capabilities", l.agentConfig.Capabilities)
		
		// Apply system prompt if configured
		if l.agentConfig.SystemPrompt != "" {
			currentState.Set("loop_system_prompt", l.agentConfig.SystemPrompt)
			currentState.Set("system_prompt_applied", true)
		}
	}

	var loopCtx context.Context
	var cancel context.CancelFunc

	// Apply timeout from agent configuration first, then from loop config
	if l.agentConfig != nil && l.agentConfig.Timeout > 0 {
		loopCtx, cancel = context.WithTimeout(ctx, l.agentConfig.Timeout)
	} else if l.config.Timeout > 0 {
		loopCtx, cancel = context.WithTimeout(ctx, l.config.Timeout)
	} else {
		loopCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	for iteration < l.config.MaxIterations {
		iteration++
		agenticgokit.Logger().Debug().
			Str("agent", l.name).
			Int("iteration", iteration).
			Int("max_iterations", l.config.MaxIterations).
			Msg("LoopAgent: Starting iteration.")

		// Check for context cancellation before running the sub-agent
		select {
		case <-loopCtx.Done():
			agenticgokit.Logger().Warn().
				Str("agent", l.name).
				Int("iteration", iteration).
				Msg("LoopAgent: Context cancelled before iteration.")
			return currentState, fmt.Errorf("LoopAgent '%s': context cancelled: %w", l.name, loopCtx.Err())
		default:
			// Context is not cancelled, proceed
		}

		// Clone state for the sub-agent run
		inputState := currentState.Clone()
		outputState, agentErr := l.subAgent.Run(loopCtx, inputState)

		if agentErr != nil {
			err = fmt.Errorf("LoopAgent '%s': error in sub-agent during iteration %d: %w", l.name, iteration, agentErr)
			agenticgokit.Logger().Error().
				Str("agent", l.name).
				Int("iteration", iteration).
				Err(agentErr).
				Msg("LoopAgent: Error in sub-agent during iteration.")
			return currentState, err
		}

		// Update the current state for the next iteration or condition check
		currentState = outputState

		// Evaluate the condition function if provided
		if l.config.Condition != nil {
			stop := l.config.Condition(currentState)
			if stop {
				agenticgokit.Logger().Debug().
					Str("agent", l.name).
					Int("iteration", iteration).
					Msg("LoopAgent: Condition met, stopping loop.")
				
				// Add execution metadata for successful completion
				currentState.Set("executed_by", l.name)
				if l.agentConfig != nil {
					currentState.Set("execution_role", l.agentConfig.Role)
				}
				currentState.Set("execution_timestamp", time.Now().Unix())
				currentState.Set("loop_execution_completed", true)
				currentState.Set("loop_condition_met", true)
				currentState.Set("loop_iterations", iteration)
				
				return currentState, nil // Condition met, loop succeeded
			}
		}
	}

	// If loop finished due to reaching max iterations without condition being met
	agenticgokit.Logger().Warn().
		Str("agent", l.name).
		Int("max_iterations", l.config.MaxIterations).
		Msg("LoopAgent: Reached max iterations without condition being met.")
	
	// Add execution metadata
	currentState.Set("executed_by", l.name)
	if l.agentConfig != nil {
		currentState.Set("execution_role", l.agentConfig.Role)
	}
	currentState.Set("execution_timestamp", time.Now().Unix())
	currentState.Set("loop_execution_completed", false)
	currentState.Set("loop_max_iterations_reached", true)
	currentState.Set("loop_iterations", iteration)
	
	return currentState, agenticgokit.ErrMaxIterationsReached
}

// =============================================================================
// CONFIGURATION INTERFACE METHODS
// =============================================================================

// SetConfiguration sets the agent's configuration
func (l *LoopAgent) SetConfiguration(config *core.ResolvedAgentConfig) {
	l.agentConfig = config
}

// GetConfiguration returns the agent's configuration
func (l *LoopAgent) GetConfiguration() *core.ResolvedAgentConfig {
	return l.agentConfig
}

// GetRole returns the agent's configured role
func (l *LoopAgent) GetRole() string {
	if l.agentConfig != nil {
		return l.agentConfig.Role
	}
	return "loop_coordinator" // Default role
}

// GetDescription returns the agent's configured description
func (l *LoopAgent) GetDescription() string {
	if l.agentConfig != nil {
		return l.agentConfig.Description
	}
	return "Executes agent in a loop until condition is met" // Default description
}

// GetSystemPrompt returns the agent's configured system prompt
func (l *LoopAgent) GetSystemPrompt() string {
	if l.agentConfig != nil {
		return l.agentConfig.SystemPrompt
	}
	return "" // No default system prompt
}

// GetCapabilities returns the agent's configured capabilities
func (l *LoopAgent) GetCapabilities() []string {
	if l.agentConfig != nil {
		return l.agentConfig.Capabilities
	}
	return []string{"loop_execution", "condition_evaluation", "agent_coordination"} // Default capabilities
}

// IsEnabled returns whether the agent is enabled in configuration
func (l *LoopAgent) IsEnabled() bool {
	if l.agentConfig != nil {
		return l.agentConfig.Enabled
	}
	return true // Default to enabled
}

// GetTimeout returns the agent's configured timeout
func (l *LoopAgent) GetTimeout() time.Duration {
	if l.agentConfig != nil {
		return l.agentConfig.Timeout
	}
	return l.config.Timeout // Fall back to loop config timeout
}

// GetLLMConfig returns the agent's configured LLM settings
func (l *LoopAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	if l.agentConfig != nil {
		return l.agentConfig.LLMConfig
	}
	return nil
}

// UpdateConfiguration updates the agent's configuration
func (l *LoopAgent) UpdateConfiguration(config *core.ResolvedAgentConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}
	l.agentConfig = config
	
	agenticgokit.Logger().Debug().
		Str("agent", l.name).
		Str("role", config.Role).
		Bool("enabled", config.Enabled).
		Strs("capabilities", config.Capabilities).
		Msg("LoopAgent configuration updated")
	
	return nil
}

// ApplySystemPrompt applies the configured system prompt to the state
func (l *LoopAgent) ApplySystemPrompt(ctx context.Context, state agenticgokit.State) (agenticgokit.State, error) {
	if l.agentConfig == nil || l.agentConfig.SystemPrompt == "" {
		return state, nil
	}

	workingState := state.Clone()
	workingState.Set("system_prompt", l.agentConfig.SystemPrompt)
	workingState.Set("system_prompt_applied", true)
	workingState.Set("system_prompt_source", "loop_agent")

	agenticgokit.Logger().Debug().
		Str("agent", l.name).
		Str("system_prompt", l.agentConfig.SystemPrompt).
		Msg("System prompt applied to state")

	return workingState, nil
}

