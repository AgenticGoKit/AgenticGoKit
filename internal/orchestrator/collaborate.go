package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// CollaborativeOrchestrator dispatches events to all registered handlers concurrently.
type CollaborativeOrchestrator struct {
	handlers         map[string]core.AgentHandler
	callbackRegistry *core.CallbackRegistry
	mu               sync.RWMutex
}

// NewCollaborativeOrchestrator creates a new CollaborativeOrchestrator.
func NewCollaborativeOrchestrator(registry *core.CallbackRegistry) *CollaborativeOrchestrator {
	return &CollaborativeOrchestrator{
		handlers:         make(map[string]core.AgentHandler),
		callbackRegistry: registry,
	}
}

// RegisterAgent adds an agent handler to the orchestrator.
func (o *CollaborativeOrchestrator) RegisterAgent(name string, handler core.AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Check if agent name already exists
	if _, exists := o.handlers[name]; exists {
		return fmt.Errorf("agent with name '%s' already registered", name)
	}
	// Store handler by name
	o.handlers[name] = handler
	core.Logger().Debug().
		Str("agent", name).
		Msg("CollaborativeOrchestrator: Registered agent")
	return nil
}

// Dispatch sends the event to all registered handlers concurrently.
func (o *CollaborativeOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
	if event == nil {
		core.Logger().Warn().Msg("CollaborativeOrchestrator: Received nil event, skipping dispatch.")
		err := errors.New("cannot dispatch nil event")
		return core.AgentResult{Error: err.Error()}, err
	}

	o.mu.RLock()
	defer o.mu.RUnlock()

	if len(o.handlers) == 0 {
		core.Logger().Warn().Msg("CollaborativeOrchestrator: No agents registered")
		err := errors.New("no agents registered")
		return core.AgentResult{Error: err.Error()}, err
	}

	// Create a channel to collect results from all agents
	resultChan := make(chan core.AgentResult, len(o.handlers))
	var wg sync.WaitGroup

	// Extract the current state from the event data
	currentState := core.NewState()
	if event != nil {
		// Create state from event data - using event metadata and data
		for k, v := range event.GetMetadata() {
			currentState.SetMeta(k, v)
		}
		// Add event data to state
		eventData := event.GetData()
		for key, value := range eventData {
			currentState.Set(key, value)
		}
	}

	// Launch goroutines for each agent handler
	for name, handler := range o.handlers {
		wg.Add(1)
		go func(agentName string, h core.AgentHandler) {
			defer wg.Done()
			core.Logger().Debug().
				Str("agent", agentName).
				Msg("CollaborativeOrchestrator: Dispatching to agent")

			result, err := h.Run(ctx, event, currentState)
			if err != nil {
				result.Error = err.Error()
			}
			resultChan <- result
		}(name, handler)
	}

	// Wait for all agents to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all results
	var results []core.AgentResult
	var errors []string
	hasSuccess := false
	combinedState := core.NewState()

	for result := range resultChan {
		results = append(results, result)
		if result.Error != "" {
			errors = append(errors, result.Error)
		} else {
			hasSuccess = true
			// Merge output states (iterate through keys since State is an interface)
			for _, key := range result.OutputState.Keys() {
				if value, ok := result.OutputState.Get(key); ok {
					combinedState.Set(key, value)
				}
			}
			// Also merge metadata
			for _, key := range result.OutputState.MetaKeys() {
				if value, ok := result.OutputState.GetMeta(key); ok {
					combinedState.SetMeta(key, value)
				}
			}
		}
	}

	// Create combined result
	combinedResult := core.AgentResult{
		OutputState: combinedState,
		StartTime:   time.Now(),
		EndTime:     time.Now(),
	}

	// If all agents failed, return error
	if !hasSuccess {
		combinedResult.Error = fmt.Sprintf("all agents failed: %v", errors)
		return combinedResult, fmt.Errorf("collaborative dispatch failed: all agents returned errors")
	}

	core.Logger().Debug().
		Int("total_agents", len(o.handlers)).
		Int("successful", len(results)-len(errors)).
		Int("failed", len(errors)).
		Msg("CollaborativeOrchestrator: Dispatch completed")

	return combinedResult, nil
}

// GetCallbackRegistry returns the callback registry
func (o *CollaborativeOrchestrator) GetCallbackRegistry() *core.CallbackRegistry {
	return o.callbackRegistry
}

// Stop halts the orchestrator
func (o *CollaborativeOrchestrator) Stop() {
	core.Logger().Debug().Msg("CollaborativeOrchestrator: Stopped")
}
