package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// CollaborativeOrchestrator dispatches events to all registered handlers concurrently.
type CollaborativeOrchestrator struct {
	handlers map[string]agentflow.AgentHandler // Use AgentHandler interface and map by name
	mu       sync.RWMutex
}

// NewCollaborativeOrchestrator creates a new CollaborativeOrchestrator.
func NewCollaborativeOrchestrator() *CollaborativeOrchestrator {
	return &CollaborativeOrchestrator{
		handlers: make(map[string]agentflow.AgentHandler), // Initialize map
	}
}

// RegisterAgent adds an agent handler to the orchestrator.
func (o *CollaborativeOrchestrator) RegisterAgent(name string, handler agentflow.AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Check if agent name already exists
	if _, exists := o.handlers[name]; exists {
		return fmt.Errorf("agent with name '%s' already registered", name)
	}
	// Store handler by name
	o.handlers[name] = handler
	agentflow.Logger().Info().
		Str("agent", name).
		Msg("CollaborativeOrchestrator: Registered agent")
	return nil
}

// Dispatch sends the event to all registered handlers concurrently.
func (o *CollaborativeOrchestrator) Dispatch(ctx context.Context, event agentflow.Event) (agentflow.AgentResult, error) {
	if event == nil {
		agentflow.Logger().Warn().Msg("CollaborativeOrchestrator: Received nil event, skipping dispatch.")
		err := errors.New("cannot dispatch nil event")
		return agentflow.AgentResult{Error: err.Error()}, err
	}

	o.mu.RLock()
	// Create a slice of handlers to run from the map values
	handlersToRun := make([]agentflow.AgentHandler, 0, len(o.handlers))
	for _, handler := range o.handlers {
		handlersToRun = append(handlersToRun, handler)
	}
	o.mu.RUnlock()

	if len(handlersToRun) == 0 {
		agentflow.Logger().Warn().
			Str("event_id", event.GetID()).
			Msg("CollaborativeOrchestrator: No handlers registered, skipping dispatch")
		return agentflow.AgentResult{}, nil
	}

	var wg sync.WaitGroup
	errsChan := make(chan error, len(handlersToRun))
	resultsChan := make(chan agentflow.AgentResult, len(handlersToRun))

	wg.Add(len(handlersToRun))
	agentflow.Logger().Info().
		Str("event_id", event.GetID()).
		Int("handler_count", len(handlersToRun)).
		Msg("CollaborativeOrchestrator: Dispatching event to handlers")

	// Create initial state from event
	initialState := agentflow.NewState()
	if eventData := event.GetData(); eventData != nil {
		for k, v := range eventData {
			initialState.Set(k, v)
		}
	}
	// Copy metadata too
	if eventMeta := event.GetMetadata(); eventMeta != nil {
		for k, v := range eventMeta {
			initialState.SetMeta(k, v)
		}
	}

	for _, handler := range handlersToRun {
		go func(h agentflow.AgentHandler) {
			defer wg.Done()
			if h == nil {
				err := fmt.Errorf("encountered nil handler during dispatch for event ID %s", event.GetID())
				agentflow.Logger().Error().
					Str("event_id", event.GetID()).
					Msg(err.Error())
				errsChan <- err
				return
			}

			stateForHandler := initialState.Clone()
			result, err := h.Run(ctx, event, stateForHandler)

			if err != nil {
				agentflow.Logger().Error().
					Str("event_id", event.GetID()).
					Err(err).
					Msg("CollaborativeOrchestrator: Handler error")
				errsChan <- err
			} else {
				agentflow.Logger().Info().
					Str("event_id", event.GetID()).
					Msg("CollaborativeOrchestrator: Handler finished successfully")
				resultsChan <- result
			}

		}(handler)
	}

	wg.Wait()
	close(resultsChan)
	close(errsChan)

	var collectedErrors []error
	for err := range errsChan {
		collectedErrors = append(collectedErrors, err)
	}

	finalState := agentflow.NewState()
	if sessionID, ok := initialState.GetMeta(agentflow.SessionIDKey); ok {
		finalState.SetMeta(agentflow.SessionIDKey, sessionID)
	}
	for _, key := range initialState.Keys() {
		if val, ok := initialState.Get(key); ok {
			finalState.Set(key, val)
		}
	}

	for result := range resultsChan {
		if result.OutputState != nil {
			for _, key := range result.OutputState.Keys() {
				if val, ok := result.OutputState.Get(key); ok {
					finalState.Set(key, val)
				}
			}
			for _, key := range result.OutputState.MetaKeys() {
				if key != agentflow.SessionIDKey && key != agentflow.RouteMetadataKey {
					if val, ok := result.OutputState.GetMeta(key); ok {
						finalState.SetMeta(key, val)
					}
				}
			}
		}
	}

	if len(collectedErrors) > 0 {
		agentflow.Logger().Warn().
			Str("event_id", event.GetID()).
			Int("error_count", len(collectedErrors)).
			Msg("CollaborativeOrchestrator: Finished dispatch with errors")
		aggErr := errors.Join(collectedErrors...)
		return agentflow.AgentResult{OutputState: finalState, Error: aggErr.Error()}, aggErr
	}

	agentflow.Logger().Info().
		Str("event_id", event.GetID()).
		Msg("CollaborativeOrchestrator: Finished dispatch successfully")
	return agentflow.AgentResult{OutputState: finalState}, nil
}

// DispatchAll needs the same signature update and logic adjustment
func (o *CollaborativeOrchestrator) DispatchAll(ctx context.Context, event agentflow.Event) (agentflow.AgentResult, error) {
	return o.Dispatch(ctx, event)
}

// Stop is a placeholder for potential cleanup tasks.
func (o *CollaborativeOrchestrator) Stop() {
	agentflow.Logger().Info().Msg("CollaborativeOrchestrator stopping...")
}

// GetCallbackRegistry returns nil as CollaborativeOrchestrator doesn't manage callbacks directly.
func (o *CollaborativeOrchestrator) GetCallbackRegistry() *agentflow.CallbackRegistry {
	return nil
}

// Compile-time check to ensure CollaborativeOrchestrator implements Orchestrator
var _ agentflow.Orchestrator = (*CollaborativeOrchestrator)(nil)
