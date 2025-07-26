package agents

import (
	"context"
	"fmt"
	"time"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/internal/core"
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
	name     string
	subAgent agenticgokit.Agent
	config   LoopAgentConfig
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
		name: name,
	}
}

// Name returns the name of the loop agent.
func (a *LoopAgent) Name() string {
	return a.name
}

// Run executes the sub-agent in a loop according to the configuration.
func (l *LoopAgent) Run(ctx context.Context, initialState agenticgokit.State) (agenticgokit.State, error) {
	currentState := initialState
	var err error
	iteration := 0

	var loopCtx context.Context
	var cancel context.CancelFunc

	// Apply overall loop timeout if configured
	if l.config.Timeout > 0 {
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
				agenticgokit.Logger().Info().
					Str("agent", l.name).
					Int("iteration", iteration).
					Msg("LoopAgent: Condition met, stopping loop.")
				return currentState, nil // Condition met, loop succeeded
			}
		}
	}

	// If loop finished due to reaching max iterations without condition being met
	agenticgokit.Logger().Warn().
		Str("agent", l.name).
		Int("max_iterations", l.config.MaxIterations).
		Msg("LoopAgent: Reached max iterations without condition being met.")
	return currentState, agenticgokit.ErrMaxIterationsReached
}