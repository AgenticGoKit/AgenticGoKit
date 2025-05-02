package agents

import (
	"context"
	"fmt"
	"log"
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
			log.Printf("Warning: ParallelAgent '%s' received a nil agent at index %d, skipping.", name, i)
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
		log.Printf("ParallelAgent '%s': No sub-agents to run.", a.name)
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
					// FIX: Use ag.Name() instead of agentflow.AgentName(ag)
					err = fmt.Errorf("agent '%s' cancelled: %w", ag.Name(), runCtx.Err())
				}
				log.Printf("ParallelAgent '%s': Context done for agent %d (%s): %v", a.name, idx, ag.Name(), runCtx.Err())
				errChan <- err
				return
			default:
			}

			if err != nil {
				// Add agent name to the error for better context if not already present
				err = fmt.Errorf("agent '%s' error: %w", ag.Name(), err)
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
			// FIX: Use mergedState.Merge(result)
			mergedState.Merge(result)
			mergeMutex.Unlock()

		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			mergeMutex.Lock() // Protect collectedErrors as well, just in case
			collectedErrors = append(collectedErrors, err)
			mergeMutex.Unlock()
		}
	}

	if len(collectedErrors) > 0 {
		multiErr := agentflow.NewMultiError(collectedErrors)
		log.Printf("ParallelAgent '%s': Finished with errors: %v", a.name, multiErr)
		return mergedState, multiErr
	}

	return mergedState, nil
}
