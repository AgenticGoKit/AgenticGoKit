package agents

import (
	"context"
	"fmt"
	"log"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// defaultMaxIterations is the default limit if LoopAgentConfig.MaxIterations is not set.
const defaultMaxIterations = 10

// ConditionFunc is a function type used by LoopAgent to determine if the loop should stop.
// It receives the current state *after* a sub-agent run and returns true to stop the loop,
// or false to continue (up to MaxIterations).
type ConditionFunc func(currentState agentflow.State) bool

// LoopAgentConfig holds configuration for the LoopAgent.
type LoopAgentConfig struct {
	// Condition is evaluated after each successful sub-agent run.
	// If it returns true, the loop stops successfully, returning the current state.
	// If nil, the loop will only stop if the sub-agent errors, the context is cancelled,
	// or MaxIterations is reached.
	Condition ConditionFunc
	// MaxIterations provides a safety limit on the number of times the sub-agent is run.
	// If this limit is reached and the Condition (if provided) has not returned true,
	// the loop terminates and returns ErrMaxIterationsReached along with the state
	// from the last successful iteration.
	// If 0 or negative, it defaults to defaultMaxIterations (10).
	MaxIterations int
}

// LoopAgent repeatedly executes a single sub-agent until a condition is met,
// an error occurs, the context is cancelled, or a maximum number of iterations is reached.
type LoopAgent struct {
	subAgent agentflow.Agent
	config   LoopAgentConfig
	name     string // Optional name for logging/identification
}

// NewLoopAgent creates a new LoopAgent.
// It requires a non-nil subAgent to execute in the loop.
// It applies the default MaxIterations if the provided value is invalid.
func NewLoopAgent(name string, config LoopAgentConfig, subAgent agentflow.Agent) *LoopAgent {
	if subAgent == nil {
		log.Printf("Error: LoopAgent '%s' requires a non-nil subAgent.", name)
		return nil // Cannot create a loop agent without a sub-agent
	}

	maxIter := config.MaxIterations
	if maxIter <= 0 {
		maxIter = defaultMaxIterations
		log.Printf("LoopAgent '%s': MaxIterations not specified or invalid, defaulting to %d.", name, maxIter)
	}

	return &LoopAgent{
		subAgent: subAgent,
		config: LoopAgentConfig{ // Store potentially modified config
			Condition:     config.Condition,
			MaxIterations: maxIter,
		},
		name: name,
	}
}

// Run executes the sub-agent in a loop according to the configuration.
// In each iteration:
// 1. Checks for context cancellation.
// 2. Clones the current state.
// 3. Runs the sub-agent with the cloned state.
// 4. If the sub-agent errors or context is cancelled during run, returns the state *before* that iteration and the error.
// 5. If successful, updates the current state with the sub-agent's output.
// 6. Evaluates the Condition function (if provided) with the new state. If true, returns the current state and nil error.
// 7. If MaxIterations is reached, returns the current state and ErrMaxIterationsReached.
func (l *LoopAgent) Run(ctx context.Context, initialState agentflow.State) (agentflow.State, error) {
	currentState := initialState // Start with the initial state
	var err error
	iteration := 0

	for iteration < l.config.MaxIterations {
		iteration++
		//log.Printf("LoopAgent '%s': Starting iteration %d/%d.", l.name, iteration, l.config.MaxIterations)

		// Check for context cancellation before running the sub-agent
		select {
		case <-ctx.Done():
			log.Printf("LoopAgent '%s': Context cancelled before iteration %d.", l.name, iteration)
			// Return the state from the *previous* successful iteration
			return currentState, fmt.Errorf("LoopAgent '%s': context cancelled: %w", l.name, ctx.Err())
		default:
			// Context is not cancelled, proceed
		}

		// Clone state for the sub-agent run
		inputState := currentState.Clone()
		outputState, agentErr := l.subAgent.Run(ctx, inputState)

		if agentErr != nil {
			err = fmt.Errorf("LoopAgent '%s': error in sub-agent during iteration %d: %w", l.name, iteration, agentErr)
			log.Printf("%v", err)
			// Return the state *before* the error occurred and the error itself
			return currentState, err
		}

		// Update the current state for the next iteration or condition check
		currentState = outputState

		// Evaluate the condition function if provided
		if l.config.Condition != nil {
			stop := l.config.Condition(currentState)
			if stop {
				//log.Printf("LoopAgent '%s': Condition met at iteration %d. Stopping loop.", l.name, iteration)
				return currentState, nil // Condition met, loop succeeded
			}
		}
	}

	// If loop finished due to reaching max iterations without condition being met
	log.Printf("LoopAgent '%s': Reached max iterations (%d) without condition being met.", l.name, l.config.MaxIterations)
	// Return the state from the last successful iteration and the specific error
	return currentState, agentflow.ErrMaxIterationsReached
}
