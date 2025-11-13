// Package orchestrator provides internal sequential orchestration functionality.
package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/agenticgokit/agenticgokit/core"
)

// SequentialOrchestrator implements sequential execution of agents
type SequentialOrchestrator struct {
	handlers         map[string]core.AgentHandler
	agentSequence    []string
	callbackRegistry *core.CallbackRegistry
	mu               sync.RWMutex
}

// NewSequentialOrchestrator creates an orchestrator that runs agents in sequence
func NewSequentialOrchestrator(registry *core.CallbackRegistry, agentNames []string) *SequentialOrchestrator {
	return &SequentialOrchestrator{
		handlers:         make(map[string]core.AgentHandler),
		agentSequence:    agentNames,
		callbackRegistry: registry,
	}
}

// RegisterAgent adds an agent to the sequential orchestrator
func (o *SequentialOrchestrator) RegisterAgent(name string, handler core.AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil for agent %s", name)
	}

	o.handlers[name] = handler
	core.Logger().Debug().Str("agent", name).Msg("SequentialOrchestrator: Agent registered")
	return nil
}

// Dispatch executes agents in the specified sequence
func (o *SequentialOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if len(o.agentSequence) == 0 {
		return core.AgentResult{}, fmt.Errorf("no agent sequence defined")
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

	// Execute agents in sequence
	for i, agentName := range o.agentSequence {
		handler, exists := o.handlers[agentName]
		if !exists {
			core.Logger().Warn().Str("agent", agentName).Msg("SequentialOrchestrator: Agent not found, skipping")
			continue
		}

		core.Logger().Debug().
			Str("agent", agentName).
			Int("position", i).
			Msg("SequentialOrchestrator: Executing agent")

		result, err := handler.Run(ctx, event, state)
		if err != nil {
			return core.AgentResult{}, fmt.Errorf("sequential agent %s failed: %w", agentName, err)
		}

		// Pass output state to next agent
		state = result.OutputState
	}

	return core.AgentResult{OutputState: state}, nil
}

// GetCallbackRegistry returns the callback registry
func (o *SequentialOrchestrator) GetCallbackRegistry() *core.CallbackRegistry {
	return o.callbackRegistry
}

// Stop halts the sequential orchestrator
func (o *SequentialOrchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.handlers = make(map[string]core.AgentHandler)
	core.Logger().Debug().Msg("SequentialOrchestrator: Stopped")
}

