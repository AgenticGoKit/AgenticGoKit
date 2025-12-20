// Package orchestrator provides internal mixed orchestration functionality.
package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/agenticgokit/agenticgokit/core"
)

// MixedOrchestrator implements hybrid orchestration combining collaborative and sequential patterns
type MixedOrchestrator struct {
	handlers                map[string]core.AgentHandler
	collaborativeAgents     map[string]core.AgentHandler // Agents that run in parallel
	collaborativeAgentNames []string                     // Names of collaborative agents
	sequentialAgents        []string                     // Agent names that run in sequence
	callbackRegistry        *core.CallbackRegistry
	mu                      sync.RWMutex
}

// NewMixedOrchestrator creates an orchestrator that combines collaborative and sequential execution
func NewMixedOrchestrator(registry *core.CallbackRegistry, collaborativeAgentNames, sequentialAgentNames []string) *MixedOrchestrator {
	return &MixedOrchestrator{
		handlers:                make(map[string]core.AgentHandler),
		collaborativeAgents:     make(map[string]core.AgentHandler),
		collaborativeAgentNames: collaborativeAgentNames,
		sequentialAgents:        sequentialAgentNames,
		callbackRegistry:        registry,
	}
}

// RegisterAgent adds an agent to the appropriate execution group
func (o *MixedOrchestrator) RegisterAgent(name string, handler core.AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil for agent %s", name)
	}

	o.handlers[name] = handler

	// Determine which group this agent belongs to
	isCollaborative := false
	for _, collabName := range o.getCollaborativeAgentNames() {
		if collabName == name {
			o.collaborativeAgents[name] = handler
			isCollaborative = true
			break
		}
	}

	core.Logger().Debug().
		Str("agent", name).
		Bool("collaborative", isCollaborative).
		Bool("sequential", !isCollaborative && o.isSequentialAgent(name)).
		Msg("MixedOrchestrator: Agent registered")

	return nil
}

// Dispatch implements hybrid execution: collaborative agents run in parallel, then sequential agents run in order
func (o *MixedOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
	core.Logger().Debug().
		Str("event_id", event.GetID()).
		Int("collaborative_agents", len(o.collaborativeAgents)).
		Int("sequential_agents", len(o.sequentialAgents)).
		Msg("MixedOrchestrator: Starting hybrid dispatch")

	// Get initial state from event data
	combinedState := core.NewState()

	// Copy event data into state
	for key, value := range event.GetData() {
		combinedState.Set(key, value)
	}

	// Phase 1: Execute collaborative agents in parallel
	if len(o.collaborativeAgents) > 0 {
		core.Logger().Debug().
			Int("agents", len(o.collaborativeAgents)).
			Msg("MixedOrchestrator: Phase 1 - Collaborative execution")

		collabResult, err := o.executeCollaborativePhase(ctx, event, combinedState)
		if err != nil {
			return core.AgentResult{}, fmt.Errorf("collaborative phase failed: %w", err)
		}

		// Merge collaborative results into combined state
		for _, key := range collabResult.OutputState.Keys() {
			if value, ok := collabResult.OutputState.Get(key); ok {
				combinedState.Set(key, value)
			}
		}
		for _, key := range collabResult.OutputState.MetaKeys() {
			if value, ok := collabResult.OutputState.GetMeta(key); ok {
				combinedState.SetMeta(key, value)
			}
		}
	}

	// Phase 2: Execute sequential agents in order
	if len(o.sequentialAgents) > 0 {
		core.Logger().Debug().
			Int("agents", len(o.sequentialAgents)).
			Msg("MixedOrchestrator: Phase 2 - Sequential execution")

		seqResult, err := o.executeSequentialPhase(ctx, event, combinedState)
		if err != nil {
			return core.AgentResult{}, fmt.Errorf("sequential phase failed: %w", err)
		}

		// Merge sequential result into combined state
		for _, key := range seqResult.OutputState.Keys() {
			if value, ok := seqResult.OutputState.Get(key); ok {
				combinedState.Set(key, value)
			}
		}
		for _, key := range seqResult.OutputState.MetaKeys() {
			if value, ok := seqResult.OutputState.GetMeta(key); ok {
				combinedState.SetMeta(key, value)
			}
		}
	}

	core.Logger().Debug().
		Str("event_id", event.GetID()).
		Msg("MixedOrchestrator: Hybrid dispatch completed successfully")

	return core.AgentResult{
		OutputState: combinedState,
	}, nil
}

// executeCollaborativePhase runs all collaborative agents in parallel
func (o *MixedOrchestrator) executeCollaborativePhase(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	if len(o.collaborativeAgents) == 0 {
		return core.AgentResult{OutputState: state}, nil
	}

	var wg sync.WaitGroup
	resultChan := make(chan core.AgentResult, len(o.collaborativeAgents))

	// Execute all collaborative agents in parallel
	for name, handler := range o.collaborativeAgents {
		wg.Add(1)
		go func(agentName string, h core.AgentHandler) {
			defer wg.Done()

			result, err := h.Run(ctx, event, state)
			if err != nil {
				result = core.AgentResult{
					OutputState: state,
					Error:       fmt.Sprintf("Agent %s failed: %v", agentName, err),
				}
			}

			resultChan <- result
		}(name, handler)
	}

	wg.Wait()
	close(resultChan)

	// Collect and merge results
	combinedState := core.NewState()
	var errors []string
	hasSuccess := false

	for result := range resultChan {
		if result.Error != "" {
			errors = append(errors, result.Error)
		} else {
			hasSuccess = true
			// Merge successful results
			for _, key := range result.OutputState.Keys() {
				if value, ok := result.OutputState.Get(key); ok {
					combinedState.Set(key, value)
				}
			}
			for _, key := range result.OutputState.MetaKeys() {
				if value, ok := result.OutputState.GetMeta(key); ok {
					combinedState.SetMeta(key, value)
				}
			}
		}
	}

	if !hasSuccess && len(errors) > 0 {
		return core.AgentResult{}, fmt.Errorf("all collaborative agents failed: %v", errors)
	}

	return core.AgentResult{OutputState: combinedState}, nil
}

// executeSequentialPhase runs sequential agents one after another
func (o *MixedOrchestrator) executeSequentialPhase(ctx context.Context, event core.Event, initialState core.State) (core.AgentResult, error) {
	if len(o.sequentialAgents) == 0 {
		return core.AgentResult{OutputState: initialState}, nil
	}

	currentState := initialState // State interface, not *SimpleState

	for i, agentName := range o.sequentialAgents {
		o.mu.RLock()
		handler, exists := o.handlers[agentName]
		o.mu.RUnlock()

		if !exists {
			core.Logger().Warn().
				Str("agent", agentName).
				Int("position", i).
				Msg("MixedOrchestrator: Sequential agent not found, skipping")
			continue
		}

		core.Logger().Debug().
			Str("agent", agentName).
			Int("position", i).
			Int("total", len(o.sequentialAgents)).
			Msg("MixedOrchestrator: Executing sequential agent")

		result, err := handler.Run(ctx, event, currentState)
		if err != nil {
			return core.AgentResult{}, fmt.Errorf("sequential agent %s (position %d) failed: %w", agentName, i, err)
		}

		// Use this agent's output as input for the next agent
		currentState = result.OutputState
	}

	return core.AgentResult{OutputState: currentState}, nil
}

// Helper methods
func (o *MixedOrchestrator) getCollaborativeAgentNames() []string {
	return o.collaborativeAgentNames
}

func (o *MixedOrchestrator) isSequentialAgent(name string) bool {
	for _, seqName := range o.sequentialAgents {
		if seqName == name {
			return true
		}
	}
	return false
}

// GetCallbackRegistry returns the callback registry for this orchestrator
func (o *MixedOrchestrator) GetCallbackRegistry() *core.CallbackRegistry {
	return o.callbackRegistry
}

// Stop halts the mixed orchestrator (cleanup if needed)
func (o *MixedOrchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Clear handlers
	o.handlers = make(map[string]core.AgentHandler)
	o.collaborativeAgents = make(map[string]core.AgentHandler)
	o.sequentialAgents = nil

	core.Logger().Debug().Msg("MixedOrchestrator: Stopped and cleaned up")
}

