package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/internal/core"
)

// CollaborativeOrchestrator dispatches events to all registered handlers concurrently.
type CollaborativeOrchestrator struct {
	handlers map[string]agenticgokit.AgentHandler // Use AgentHandler interface and map by name
	mu       sync.RWMutex
}

// NewCollaborativeOrchestrator creates a new CollaborativeOrchestrator.
func NewCollaborativeOrchestrator() *CollaborativeOrchestrator {
	return &CollaborativeOrchestrator{
		handlers: make(map[string]agenticgokit.AgentHandler), // Initialize map
	}
}

// RegisterAgent adds an agent handler to the orchestrator.
func (o *CollaborativeOrchestrator) RegisterAgent(name string, handler agenticgokit.AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Check if agent name already exists
	if _, exists := o.handlers[name]; exists {
		return fmt.Errorf("agent with name '%s' already registered", name)
	}
	// Store handler by name
	o.handlers[name] = handler
	agenticgokit.Logger().Info().
		Str("agent", name).
		Msg("CollaborativeOrchestrator: Registered agent")
	return nil
}

// Dispatch sends the event to all registered handlers concurrently.
func (o *CollaborativeOrchestrator) Dispatch(ctx context.Context, event agenticgokit.Event) (agenticgokit.AgentResult, error) {
	if event == nil {
		agenticgokit.Logger().Warn().Msg("CollaborativeOrchestrator: Received nil event, skipping dispatch.")
		err := errors.New("cannot dispatch nil event")
		return agenticgokit.AgentResult{Error: err.Error()}, err
	}

	o.mu.RLock()
	// Create a slice of handlers to run from the map values
	handlersToRun := make([]agenticgokit.AgentHandler, 0, len(o.handlers))
	for _, handler := range o.handlers {
		handlersToRun = append(handlersToRun, handler)
	}
	o.mu.RUnlock()

	if len(handlersToRun) == 0 {
		agenticgokit.Logger().Warn().
			Str("event_id", event.GetID()).
			Msg("CollaborativeOrchestrator: No handlers registered, skipping dispatch")
		return agenticgokit.AgentResult{}, nil
	}

	var wg sync.WaitGroup
	errsChan := make(chan error, len(handlersToRun))
	resultsChan := make(chan agenticgokit.AgentResult, len(handlersToRun))

	wg.Add(len(handlersToRun))
	agenticgokit.Logger().Info().
		Str("event_id", event.GetID()).
		Int("handler_count", len(handlersToRun)).
		Msg("CollaborativeOrchestrator: Dispatching event to handlers")

	// Create initial state from event
	initialState := agenticgokit.NewState()
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
		go func(h agenticgokit.AgentHandler) {
			defer wg.Done()
			if h == nil {
				err := fmt.Errorf("encountered nil handler during dispatch for event ID %s", event.GetID())
				agenticgokit.Logger().Error().
					Str("event_id", event.GetID()).
					Msg(err.Error())
				errsChan <- err
				return
			}

			stateForHandler := initialState.Clone()
			result, err := h.Run(ctx, event, stateForHandler)

			if err != nil {
				agenticgokit.Logger().Error().
					Str("event_id", event.GetID()).
					Err(err).
					Msg("CollaborativeOrchestrator: Handler error")
				errsChan <- err
			} else {
				agenticgokit.Logger().Info().
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

	finalState := agenticgokit.NewState()
	if sessionID, ok := initialState.GetMeta(agenticgokit.SessionIDKey); ok {
		finalState.SetMeta(agenticgokit.SessionIDKey, sessionID)
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
				if key != agenticgokit.SessionIDKey && key != agenticgokit.RouteMetadataKey {
					if val, ok := result.OutputState.GetMeta(key); ok {
						finalState.SetMeta(key, val)
					}
				}
			}
		}
	}

	if len(collectedErrors) > 0 {
		agenticgokit.Logger().Warn().
			Str("event_id", event.GetID()).
			Int("error_count", len(collectedErrors)).
			Msg("CollaborativeOrchestrator: Finished dispatch with errors")
		aggErr := errors.Join(collectedErrors...)
		return agenticgokit.AgentResult{OutputState: finalState, Error: aggErr.Error()}, aggErr
	}

	agenticgokit.Logger().Info().
		Str("event_id", event.GetID()).
		Msg("CollaborativeOrchestrator: Finished dispatch successfully")
	return agenticgokit.AgentResult{OutputState: finalState}, nil
}

// DispatchAll needs the same signature update and logic adjustment
func (o *CollaborativeOrchestrator) DispatchAll(ctx context.Context, event agenticgokit.Event) (agenticgokit.AgentResult, error) {
	return o.Dispatch(ctx, event)
}

// Stop is a placeholder for potential cleanup tasks.
func (o *CollaborativeOrchestrator) Stop() {
	agenticgokit.Logger().Info().Msg("CollaborativeOrchestrator stopping...")
}

// GetCallbackRegistry returns nil as CollaborativeOrchestrator doesn't manage callbacks directly.
func (o *CollaborativeOrchestrator) GetCallbackRegistry() *agenticgokit.CallbackRegistry {
	return nil
}

// Compile-time check to ensure CollaborativeOrchestrator implements Orchestrator
var _ agenticgokit.Orchestrator = (*CollaborativeOrchestrator)(nil)