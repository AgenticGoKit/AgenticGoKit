package agents

import (
	"context"
	"fmt"
	"log"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// SequentialAgent runs a series of sub-agents one after another.
type SequentialAgent struct {
	name   string
	agents []agentflow.Agent
}

// Name returns the name of the sequential agent.
func (a *SequentialAgent) Name() string {
	return a.name
}

// NewSequentialAgent creates a new SequentialAgent.
// It filters out any nil agents provided in the list.
func NewSequentialAgent(name string, agents ...agentflow.Agent) *SequentialAgent {
	validAgents := make([]agentflow.Agent, 0, len(agents))
	for i, agent := range agents {
		if agent == nil {
			log.Printf("Warning: SequentialAgent '%s' received a nil agent at index %d, skipping.", name, i)
			continue
		}
		validAgents = append(validAgents, agent)
	}
	return &SequentialAgent{
		agents: validAgents,
		name:   name,
	}
}

// Run executes the sequence of sub-agents.
// It iterates through the configured agents, passing state sequentially.
// Execution halts immediately if a sub-agent returns an error or if the context is cancelled.
func (s *SequentialAgent) Run(ctx context.Context, initialState agentflow.State) (agentflow.State, error) {
	if len(s.agents) == 0 {
		log.Printf("SequentialAgent '%s': No sub-agents to run.", s.name)
		return initialState, nil // Return input state if no agents
	}

	var err error
	nextState := initialState // Start with the initial state

	for i, agent := range s.agents {
		// Check for context cancellation before running each sub-agent
		select {
		case <-ctx.Done():
			log.Printf("SequentialAgent '%s': Context cancelled before running agent %d.", s.name, i)
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
			log.Printf("%v", err) // Log the error
			// Return the state *before* the error occurred and the error itself
			return nextState, err
		}
		// Update the state for the next iteration
		nextState = outputState
	}

	// Return the final state after all agents completed successfully
	return nextState, nil
}
