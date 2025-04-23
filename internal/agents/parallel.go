package agents

import (
	"context"
	"log"
	"sync"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// ParallelAgentConfig holds configuration for the ParallelAgent.
type ParallelAgentConfig struct {
	// Timeout specifies an optional duration limit for the entire parallel execution.
	// If the timeout is reached before all agents complete, pending agents are cancelled
	// via context, and a MultiError containing context.DeadlineExceeded (wrapped)
	// will be returned along with results from completed agents.
	// A zero value means no timeout is applied beyond the parent context's deadline.
	Timeout time.Duration
}

// ParallelAgent executes a list of sub-agents concurrently.
// It waits for all sub-agents to complete (or be cancelled/timeout) and then
// aggregates their output states and errors.
// The final state is a merge of the initial state and the data from all successfully
// completed sub-agents (last write wins for conflicting keys).
// If any sub-agents return errors or are cancelled, a MultiError is returned.
type ParallelAgent struct {
	agents []agentflow.Agent
	config ParallelAgentConfig
	name   string // Optional name for logging/identification
}

// NewParallelAgent creates a new ParallelAgent.
// It filters out any nil agents provided in the list.
func NewParallelAgent(name string, config ParallelAgentConfig, agents ...agentflow.Agent) *ParallelAgent {
	validAgents := make([]agentflow.Agent, 0, len(agents))
	for i, agent := range agents {
		if agent == nil {
			log.Printf("Warning: ParallelAgent '%s' received a nil agent at index %d, skipping.", name, i)
			continue
		}
		validAgents = append(validAgents, agent)
	}
	return &ParallelAgent{
		agents: validAgents,
		config: config,
		name:   name,
	}
}

// subResult holds the output state and error from a single sub-agent execution within ParallelAgent.
type subResult struct {
	OutputState agentflow.State
	Error       error
	AgentIndex  int // Keep track of which agent produced this result
}

// Run executes all configured sub-agents concurrently in separate goroutines.
// It applies the configured timeout (if any) using context.WithTimeout.
// It collects results and errors from all goroutines.
// Data from successful runs is merged into the final state.
// Errors are aggregated into a MultiError.
func (p *ParallelAgent) Run(ctx context.Context, initialState agentflow.State) (agentflow.State, error) {
	numAgents := len(p.agents)
	if numAgents == 0 {
		log.Printf("ParallelAgent '%s': No sub-agents to run.", p.name)
		return initialState, nil // Return input state if no agents
	}

	// Prepare context with timeout if configured
	runCtx := ctx
	var cancel context.CancelFunc
	if p.config.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, p.config.Timeout)
	} else {
		runCtx, cancel = context.WithCancel(ctx) // Still need cancel for early exit
	}
	defer cancel() // Ensure cancellation propagates eventually

	resultsChan := make(chan subResult, numAgents)
	var wg sync.WaitGroup
	wg.Add(numAgents)

	// Fan-out: Launch goroutine for each agent
	for i, agent := range p.agents {
		go func(agentIdx int, currentAgent agentflow.Agent, input agentflow.State) {
			defer wg.Done()

			// Each goroutine gets a clone of the initial state
			clonedInput := input.Clone()
			output, err := currentAgent.Run(runCtx, clonedInput)

			// Check if context was cancelled *during* agent execution
			select {
			case <-runCtx.Done():
				// If context finished (timeout or external cancel), potentially override agent error
				if err == nil { // If agent finished successfully but context timed out/cancelled
					err = runCtx.Err() // Report context error instead
				}
				log.Printf("ParallelAgent '%s': Context done for agent %d: %v", p.name, agentIdx, runCtx.Err())
			default:
				// Context not done, proceed with agent's result
			}

			resultsChan <- subResult{
				OutputState: output,
				Error:       err,
				AgentIndex:  agentIdx,
			}
		}(i, agent, initialState)
	}

	// Fan-in: Wait for all goroutines to complete and close channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results and errors
	finalState := initialState.Clone() // Start with a clone of the initial state for aggregation
	allErrors := make([]error, 0, numAgents)
	processedData := make(map[string]interface{}) // Collect data from successful runs

	for result := range resultsChan {
		if result.Error != nil {
			allErrors = append(allErrors, result.Error)
			// Optionally log the individual error here
			// log.Printf("ParallelAgent '%s': Error from agent %d: %v", p.name, result.AgentIndex, result.Error)
		} else {
			// If there was no error, process the output state.
			// Aggregate data from successful results.
			// Simple merge: last write wins for conflicting keys.
			// More sophisticated merging might be needed depending on requirements.
			for k, v := range result.OutputState.GetData() {
				processedData[k] = v
			}
			// Merge metadata as well (last write wins)
			for k, v := range result.OutputState.GetMetadata() {
				finalState.SetMeta(k, v) // Directly set metadata on the final state
			}
		}
	}

	// Apply the aggregated data to the final state
	for k, v := range processedData {
		finalState.Set(k, v)
	}

	// Aggregate errors
	multiErr := agentflow.NewMultiError(allErrors)
	if multiErr != nil {
		log.Printf("ParallelAgent '%s': Finished with errors: %v", p.name, multiErr)
	}

	return finalState, multiErr
}
