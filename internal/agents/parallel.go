package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// ParallelAgentConfig holds configuration for ParallelAgent.
type ParallelAgentConfig struct {
	Timeout time.Duration // Optional timeout for the entire parallel execution.
}

// ParallelAgent runs multiple sub-agents concurrently.
// It merges the results from successful agents.
// If any agent errors, it collects the errors but allows others to complete (unless cancelled).
type ParallelAgent struct {
	name   string
	agents []agentflow.Agent
	config ParallelAgentConfig
}

// NewParallelAgent creates a new ParallelAgent.
// It filters out any nil agents provided in the variadic agents argument.
func NewParallelAgent(name string, config ParallelAgentConfig, agents ...agentflow.Agent) *ParallelAgent {
	validAgents := make([]agentflow.Agent, 0, len(agents))
	for i, agent := range agents {
		if agent != nil {
			validAgents = append(validAgents, agent)
		} else {
			// Log a warning if a nil agent is skipped
			agentflow.Logger().Warn().
				Str("parallel_agent", name).
				Int("index", i).
				Msg("ParallelAgent: received a nil agent, skipping.")
		}
	}
	return &ParallelAgent{
		name:   name,
		agents: validAgents,
		config: config,
	}
}

// Name returns the name of the parallel agent.
func (a *ParallelAgent) Name() string {
	return a.name
}

// Run executes all sub-agents in parallel.
func (a *ParallelAgent) Run(ctx context.Context, initialState agentflow.State) (agentflow.State, error) {
	if len(a.agents) == 0 {
		agentflow.Logger().Warn().
			Str("parallel_agent", a.name).
			Msg("ParallelAgent: No sub-agents to run.")
		return initialState.Clone(), nil
	}

	var wg sync.WaitGroup
	resultsChan := make(chan agentflow.State, len(a.agents))
	errChan := make(chan error, len(a.agents))
	mergedState := initialState.Clone() // Start with a clone to merge into
	var mergeMutex sync.Mutex           // Mutex to protect mergedState during concurrent merges
	var collectedErrors []error

	runCtx, cancel := context.WithCancel(ctx)
	if a.config.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, a.config.Timeout)
	}
	defer cancel() // Ensure context is cancelled on exit

	wg.Add(len(a.agents))

	for i, agent := range a.agents {
		go func(idx int, ag agentflow.Agent) {
			defer wg.Done()

			agentInputState := initialState.Clone()
			agentResultState, err := ag.Run(runCtx, agentInputState)

			select {
			case <-runCtx.Done():
				if err == nil {
					err = fmt.Errorf("agent '%s' cancelled: %w", ag.Name(), runCtx.Err())
				}
				agentflow.Logger().Warn().
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
				agentflow.Logger().Error().
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
		multiErr := agentflow.NewMultiError(collectedErrors)
		agentflow.Logger().Error().
			Str("parallel_agent", a.name).
			Err(multiErr).
			Msg("ParallelAgent: Finished with errors")
		return mergedState, multiErr
	}

	return mergedState, nil
}
