// Package orchestrator provides internal loop orchestration functionality.
package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// LoopOrchestrator implements loop execution with a single agent
type LoopOrchestrator struct {
	handlers         map[string]core.AgentHandler
	agentName        string
	maxIterations    int
	callbackRegistry *core.CallbackRegistry
	mu               sync.RWMutex
}

// NewLoopOrchestrator creates an orchestrator that runs a single agent in a loop
func NewLoopOrchestrator(registry *core.CallbackRegistry, agentNames []string) *LoopOrchestrator {
	agentName := ""
	if len(agentNames) > 0 {
		agentName = agentNames[0] // Use first agent for loop
	}
	return &LoopOrchestrator{
		handlers:         make(map[string]core.AgentHandler),
		agentName:        agentName,
		maxIterations:    5, // Default iterations
		callbackRegistry: registry,
	}
}

// RegisterAgent adds an agent to the loop orchestrator
func (o *LoopOrchestrator) RegisterAgent(name string, handler core.AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil for agent %s", name)
	}

	o.handlers[name] = handler
	core.Logger().Debug().Str("agent", name).Msg("LoopOrchestrator: Agent registered")
	return nil
}

// Dispatch executes the specified agent in a loop
func (o *LoopOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.agentName == "" {
		return core.AgentResult{}, fmt.Errorf("no agent specified for loop")
	}

	handler, exists := o.handlers[o.agentName]
	if !exists {
		return core.AgentResult{}, fmt.Errorf("loop agent %s not found", o.agentName)
	}

	// Initialize state from event data
	currentState := core.NewState()
	for key, value := range event.GetData() {
		currentState.Set(key, value)
	}
	for key, value := range event.GetMetadata() {
		currentState.SetMeta(key, value)
	}

	var state core.State = currentState // Use State interface for chaining

	// Execute agent in loop
	for i := 0; i < o.maxIterations; i++ {
		core.Logger().Debug().
			Str("agent", o.agentName).
			Int("iteration", i+1).
			Int("max_iterations", o.maxIterations).
			Msg("LoopOrchestrator: Executing agent iteration")

		result, err := handler.Run(ctx, event, state)
		if err != nil {
			return core.AgentResult{}, fmt.Errorf("loop agent %s (iteration %d) failed: %w", o.agentName, i+1, err)
		}

		// Check for completion signal in state
		if completed, ok := result.OutputState.Get("loop_completed"); ok {
			if completedBool, isBool := completed.(bool); isBool && completedBool {
				core.Logger().Info().
					Str("agent", o.agentName).
					Int("iteration", i+1).
					Msg("LoopOrchestrator: Agent signaled completion, stopping loop")
				return core.AgentResult{OutputState: result.OutputState}, nil
			}
		}

		// Use this iteration's output as input for the next iteration
		state = result.OutputState
	}

	core.Logger().Info().
		Str("agent", o.agentName).
		Int("iterations", o.maxIterations).
		Msg("LoopOrchestrator: Completed maximum iterations")

	return core.AgentResult{OutputState: state}, nil
}

// SetMaxIterations sets the maximum number of loop iterations
func (o *LoopOrchestrator) SetMaxIterations(max int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.maxIterations = max
}

// GetCallbackRegistry returns the callback registry
func (o *LoopOrchestrator) GetCallbackRegistry() *core.CallbackRegistry {
	return o.callbackRegistry
}

// Stop halts the loop orchestrator
func (o *LoopOrchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.handlers = make(map[string]core.AgentHandler)
	core.Logger().Debug().Msg("LoopOrchestrator: Stopped")
}