package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	log.Printf("CollaborativeOrchestrator: Registered agent '%s'", name)
	return nil
}

// Dispatch sends the event to all registered handlers concurrently.
func (o *CollaborativeOrchestrator) Dispatch(ctx context.Context, event agentflow.Event) (agentflow.AgentResult, error) {
	if event == nil {
		log.Println("CollaborativeOrchestrator: Received nil event, skipping dispatch.")
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
		log.Printf("CollaborativeOrchestrator: No handlers registered, skipping dispatch for event ID %s", event.GetID())
		return agentflow.AgentResult{}, nil
	}

	var wg sync.WaitGroup
	// *** TASK 6: Replace mutex/slice with error channel ***
	errsChan := make(chan error, len(handlersToRun))
	// *** TASK 5: Introduce results channel ***
	resultsChan := make(chan agentflow.AgentResult, len(handlersToRun))

	wg.Add(len(handlersToRun))
	log.Printf("CollaborativeOrchestrator: Dispatching event ID %s to %d handlers", event.GetID(), len(handlersToRun))

	// *** TASK 4: Create initial state from event ***
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
				log.Printf("CollaborativeOrchestrator: %v", err)
				// *** TASK 6: Send error to channel ***
				errsChan <- err
				return
			}

			// *** TASK 4: Clone state for each handler ***
			stateForHandler := initialState.Clone()

			// *** TASK 2 & 3: Call h.Run with context and cloned state ***
			result, err := h.Run(ctx, event, stateForHandler)

			// *** TASK 5 & 6: Send result or error to appropriate channel ***
			if err != nil {
				log.Printf("CollaborativeOrchestrator: Handler error for event ID %s: %v", event.GetID(), err)
				// *** TASK 6: Send error to channel ***
				errsChan <- err
				// Optionally send partial result if needed: resultsChan <- result
			} else {
				log.Printf("CollaborativeOrchestrator: Handler finished successfully for event ID %s", event.GetID())
				resultsChan <- result // Send successful result
			}

		}(handler)
	}

	wg.Wait()
	// *** TASK 5 & 6: Close channels ***
	close(resultsChan)
	close(errsChan)

	// *** TASK 6: Collect errors from channel ***
	var collectedErrors []error
	for err := range errsChan {
		collectedErrors = append(collectedErrors, err)
	}

	// *** TASK 5: Aggregate results (State) ***
	finalState := agentflow.NewState()
	// Preserve initial metadata like session ID
	if sessionID, ok := initialState.GetMeta(agentflow.SessionIDKey); ok {
		finalState.SetMeta(agentflow.SessionIDKey, sessionID)
	}
	// Copy initial data
	for _, key := range initialState.Keys() {
		if val, ok := initialState.Get(key); ok {
			finalState.Set(key, val)
		}
	}

	for result := range resultsChan {
		// Merge output states
		if result.OutputState != nil {
			for _, key := range result.OutputState.Keys() {
				if val, ok := result.OutputState.Get(key); ok {
					finalState.Set(key, val) // Last write wins
				}
			}
			// Merge metadata
			for _, key := range result.OutputState.MetaKeys() {
				if key != agentflow.SessionIDKey && key != agentflow.RouteMetadataKey {
					if val, ok := result.OutputState.GetMeta(key); ok {
						finalState.SetMeta(key, val)
					}
				}
			}
		}
	}

	// *** TASK 6: Aggregate errors using errors.Join ***
	if len(collectedErrors) > 0 {
		log.Printf("CollaborativeOrchestrator: Finished dispatch for event ID %s with %d errors", event.GetID(), len(collectedErrors))
		aggErr := errors.Join(collectedErrors...) // Use errors.Join
		// *** TASK 5: Return aggregated state even on error ***
		return agentflow.AgentResult{OutputState: finalState, Error: aggErr.Error()}, aggErr
	}

	log.Printf("CollaborativeOrchestrator: Finished dispatch for event ID %s successfully", event.GetID())
	// *** TASK 5: Return aggregated state on success ***
	return agentflow.AgentResult{OutputState: finalState}, nil
}

// DispatchAll needs the same signature update and logic adjustment
func (o *CollaborativeOrchestrator) DispatchAll(ctx context.Context, event agentflow.Event) (agentflow.AgentResult, error) {
	// For now, delegate to Dispatch.
	return o.Dispatch(ctx, event)
}

// Stop is a placeholder for potential cleanup tasks.
func (o *CollaborativeOrchestrator) Stop() {
	log.Println("CollaborativeOrchestrator stopping...")
}

// GetCallbackRegistry returns nil as CollaborativeOrchestrator doesn't manage callbacks directly.
func (o *CollaborativeOrchestrator) GetCallbackRegistry() *agentflow.CallbackRegistry {
	return nil
}

// *** TASK 7: Ensure Interface Compliance ***
// Compile-time check to ensure CollaborativeOrchestrator implements Orchestrator
var _ agentflow.Orchestrator = (*CollaborativeOrchestrator)(nil)
