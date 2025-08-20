package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/internal/core"
	"github.com/kunalkushwaha/agenticgokit/core"
)

// ParallelAgentConfig holds configuration for ParallelAgent.
type ParallelAgentConfig struct {
	Timeout time.Duration // Optional timeout for the entire parallel execution.
}

// ParallelAgent runs multiple sub-agents concurrently.
// It merges the results from successful agents.
// If any agent errors, it collects the errors but allows others to complete (unless cancelled).
type ParallelAgent struct {
	name         string
	agents       []agenticgokit.Agent
	config       ParallelAgentConfig
	agentConfig  *core.ResolvedAgentConfig // Add agent configuration support
}

// NewParallelAgent creates a new ParallelAgent.
// It filters out any nil agents provided in the variadic agents argument.
func NewParallelAgent(name string, config ParallelAgentConfig, agents ...agenticgokit.Agent) *ParallelAgent {
	validAgents := make([]agenticgokit.Agent, 0, len(agents))
	for i, agent := range agents {
		if agent != nil {
			validAgents = append(validAgents, agent)
		} else {
			// Log a warning if a nil agent is skipped
			agenticgokit.Logger().Warn().
				Str("parallel_agent", name).
				Int("index", i).
				Msg("ParallelAgent: received a nil agent, skipping.")
		}
	}
	return &ParallelAgent{
		name:        name,
		agents:      validAgents,
		config:      config,
		agentConfig: nil, // Configuration can be set later
	}
}

// NewParallelAgentWithConfig creates a new ParallelAgent with agent configuration.
func NewParallelAgentWithConfig(name string, config ParallelAgentConfig, agentConfig *core.ResolvedAgentConfig, agents ...agenticgokit.Agent) *ParallelAgent {
	validAgents := make([]agenticgokit.Agent, 0, len(agents))
	for i, agent := range agents {
		if agent != nil {
			validAgents = append(validAgents, agent)
		} else {
			// Log a warning if a nil agent is skipped
			agenticgokit.Logger().Warn().
				Str("parallel_agent", name).
				Int("index", i).
				Msg("ParallelAgent: received a nil agent, skipping.")
		}
	}
	return &ParallelAgent{
		name:        name,
		agents:      validAgents,
		config:      config,
		agentConfig: agentConfig,
	}
}

// Name returns the name of the parallel agent.
func (a *ParallelAgent) Name() string {
	return a.name
}

// Run executes all sub-agents in parallel.
func (a *ParallelAgent) Run(ctx context.Context, initialState agenticgokit.State) (agenticgokit.State, error) {
	// Check if agent is enabled
	if !a.IsEnabled() {
		agenticgokit.Logger().Debug().
			Str("parallel_agent", a.name).
			Msg("ParallelAgent: Agent is disabled, skipping execution.")
		return initialState, nil
	}

	if len(a.agents) == 0 {
		agenticgokit.Logger().Warn().
			Str("parallel_agent", a.name).
			Msg("ParallelAgent: No sub-agents to run.")
		return initialState.Clone(), nil
	}

	var wg sync.WaitGroup
	resultsChan := make(chan agenticgokit.State, len(a.agents))
	errChan := make(chan error, len(a.agents))
	mergedState := initialState.Clone() // Start with a clone to merge into
	var mergeMutex sync.Mutex           // Mutex to protect mergedState during concurrent merges
	var collectedErrors []error

	// Add configuration metadata to merged state
	if a.agentConfig != nil {
		mergedState.Set("parallel_agent_role", a.agentConfig.Role)
		mergedState.Set("parallel_agent_description", a.agentConfig.Description)
		mergedState.Set("parallel_agent_capabilities", a.agentConfig.Capabilities)
		
		// Apply system prompt if configured
		if a.agentConfig.SystemPrompt != "" {
			mergedState.Set("parallel_system_prompt", a.agentConfig.SystemPrompt)
			mergedState.Set("system_prompt_applied", true)
		}
	}

	runCtx, cancel := context.WithCancel(ctx)
	
	// Apply timeout from agent configuration first, then from parallel config
	if a.agentConfig != nil && a.agentConfig.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, a.agentConfig.Timeout)
	} else if a.config.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, a.config.Timeout)
	}
	defer cancel() // Ensure context is cancelled on exit

	wg.Add(len(a.agents))

	for i, agent := range a.agents {
		go func(idx int, ag agenticgokit.Agent) {
			defer wg.Done()

			agentInputState := initialState.Clone()
			agentResultState, err := ag.Run(runCtx, agentInputState)

			select {
			case <-runCtx.Done():
				if err == nil {
					err = fmt.Errorf("agent '%s' cancelled: %w", ag.Name(), runCtx.Err())
				}
				agenticgokit.Logger().Warn().
					Str("parallel_agent", a.name).
					Int("agent_index", idx).
					Str("agent_name", ag.Name()).
					Err(runCtx.Err()).
					Msg("ParallelAgent: Context done for agent")
				errChan <- err
				return
			default:
			}

			if err != nil {
				err = fmt.Errorf("agent '%s' error: %w", ag.Name(), err)
				agenticgokit.Logger().Error().
					Str("parallel_agent", a.name).
					Int("agent_index", idx).
					Str("agent_name", ag.Name()).
					Err(err).
					Msg("ParallelAgent: Agent error")
				errChan <- err
				return
			}

			if agentResultState != nil {
				resultsChan <- agentResultState
			}

		}(i, agent)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
		close(errChan)
	}()

	for resultsChan != nil || errChan != nil {
		select {
		case result, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
				continue
			}
			mergeMutex.Lock()
			mergedState.Merge(result)
			mergeMutex.Unlock()

		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			mergeMutex.Lock()
			collectedErrors = append(collectedErrors, err)
			mergeMutex.Unlock()
		}
	}

	if len(collectedErrors) > 0 {
		multiErr := agenticgokit.NewMultiError(collectedErrors)
		agenticgokit.Logger().Error().
			Str("parallel_agent", a.name).
			Err(multiErr).
			Msg("ParallelAgent: Finished with errors")
		return mergedState, multiErr
	}

	// Add execution metadata
	mergedState.Set("executed_by", a.name)
	if a.agentConfig != nil {
		mergedState.Set("execution_role", a.agentConfig.Role)
	}
	mergedState.Set("execution_timestamp", time.Now().Unix())
	mergedState.Set("parallel_execution_completed", true)

	return mergedState, nil
}

// =============================================================================
// CONFIGURATION INTERFACE METHODS
// =============================================================================

// SetConfiguration sets the agent's configuration
func (a *ParallelAgent) SetConfiguration(config *core.ResolvedAgentConfig) {
	a.agentConfig = config
}

// GetConfiguration returns the agent's configuration
func (a *ParallelAgent) GetConfiguration() *core.ResolvedAgentConfig {
	return a.agentConfig
}

// GetRole returns the agent's configured role
func (a *ParallelAgent) GetRole() string {
	if a.agentConfig != nil {
		return a.agentConfig.Role
	}
	return "parallel_coordinator" // Default role
}

// GetDescription returns the agent's configured description
func (a *ParallelAgent) GetDescription() string {
	if a.agentConfig != nil {
		return a.agentConfig.Description
	}
	return "Executes agents in parallel" // Default description
}

// GetSystemPrompt returns the agent's configured system prompt
func (a *ParallelAgent) GetSystemPrompt() string {
	if a.agentConfig != nil {
		return a.agentConfig.SystemPrompt
	}
	return "" // No default system prompt
}

// GetCapabilities returns the agent's configured capabilities
func (a *ParallelAgent) GetCapabilities() []string {
	if a.agentConfig != nil {
		return a.agentConfig.Capabilities
	}
	return []string{"parallel_execution", "agent_coordination"} // Default capabilities
}

// IsEnabled returns whether the agent is enabled in configuration
func (a *ParallelAgent) IsEnabled() bool {
	if a.agentConfig != nil {
		return a.agentConfig.Enabled
	}
	return true // Default to enabled
}

// GetTimeout returns the agent's configured timeout
func (a *ParallelAgent) GetTimeout() time.Duration {
	if a.agentConfig != nil {
		return a.agentConfig.Timeout
	}
	return a.config.Timeout // Fall back to parallel config timeout
}

// GetLLMConfig returns the agent's configured LLM settings
func (a *ParallelAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	if a.agentConfig != nil {
		return a.agentConfig.LLMConfig
	}
	return nil
}

// UpdateConfiguration updates the agent's configuration
func (a *ParallelAgent) UpdateConfiguration(config *core.ResolvedAgentConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}
	a.agentConfig = config
	
	agenticgokit.Logger().Debug().
		Str("agent", a.name).
		Str("role", config.Role).
		Bool("enabled", config.Enabled).
		Strs("capabilities", config.Capabilities).
		Msg("ParallelAgent configuration updated")
	
	return nil
}

// ApplySystemPrompt applies the configured system prompt to the state
func (a *ParallelAgent) ApplySystemPrompt(ctx context.Context, state agenticgokit.State) (agenticgokit.State, error) {
	if a.agentConfig == nil || a.agentConfig.SystemPrompt == "" {
		return state, nil
	}

	workingState := state.Clone()
	workingState.Set("system_prompt", a.agentConfig.SystemPrompt)
	workingState.Set("system_prompt_applied", true)
	workingState.Set("system_prompt_source", "parallel_agent")

	agenticgokit.Logger().Debug().
		Str("agent", a.name).
		Str("system_prompt", a.agentConfig.SystemPrompt).
		Msg("System prompt applied to state")

	return workingState, nil
}
