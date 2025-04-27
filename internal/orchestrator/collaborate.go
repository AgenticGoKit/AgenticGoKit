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
	handlers []agentflow.EventHandler // Use EventHandler interface
	mu       sync.RWMutex
}

// NewCollaborativeOrchestrator creates a new CollaborativeOrchestrator.
func NewCollaborativeOrchestrator() *CollaborativeOrchestrator {
	return &CollaborativeOrchestrator{
		handlers: make([]agentflow.EventHandler, 0),
	}
}

// RegisterAgent adds a handler to the orchestrator.
// Note: agentID is ignored in this implementation but kept for potential future use
// or consistency with other orchestrator RegisterAgent methods.
func (o *CollaborativeOrchestrator) RegisterAgent(agentID string, handler agentflow.EventHandler) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if handler == nil {
		log.Printf("Warning: Attempted to register a nil handler for agent ID '%s'", agentID)
		return
	}
	o.handlers = append(o.handlers, handler)
	log.Printf("Registered handler for collaborative dispatch (agentID: %s)", agentID)
}

// Dispatch sends the event to all registered handlers concurrently.
func (o *CollaborativeOrchestrator) Dispatch(ctx context.Context, event agentflow.Event) (agentflow.AgentResult, error) {
	if event == nil {
		log.Println("CollaborativeOrchestrator: Received nil event, skipping dispatch.")
		err := errors.New("cannot dispatch nil event")
		return agentflow.AgentResult{Error: err.Error()}, err
	}

	o.mu.RLock()
	handlersCopy := make([]agentflow.EventHandler, len(o.handlers))
	copy(handlersCopy, o.handlers)
	o.mu.RUnlock()

	if len(handlersCopy) == 0 {
		log.Printf("CollaborativeOrchestrator: No handlers registered, skipping dispatch for event ID %s", event.GetID())
		return agentflow.AgentResult{}, nil
	}

	var wg sync.WaitGroup
	// FIX: Use a mutex to protect the errors slice instead of a channel
	var errs []error
	var errsMu sync.Mutex
	wg.Add(len(handlersCopy))

	log.Printf("CollaborativeOrchestrator: Dispatching event ID %s to %d handlers", event.GetID(), len(handlersCopy))

	for _, handler := range handlersCopy {
		go func(h agentflow.EventHandler) {
			defer wg.Done()
			if h == nil {
				log.Printf("CollaborativeOrchestrator: Encountered nil handler during dispatch for event ID %s", event.GetID())
				errsMu.Lock()
				errs = append(errs, fmt.Errorf("encountered nil handler"))
				errsMu.Unlock()
				return
			}
			// Assuming Handle doesn't need context for now.
			if err := h.Handle(event); err != nil {
				log.Printf("CollaborativeOrchestrator: Handler error for event ID %s: %v", event.GetID(), err)
				errsMu.Lock()
				errs = append(errs, err)
				errsMu.Unlock()
			}
		}(handler)
	}

	wg.Wait()

	// FIX: Aggregate errors into a single error if any occurred
	if len(errs) > 0 {
		log.Printf("CollaborativeOrchestrator: Finished dispatch for event ID %s with %d errors", event.GetID(), len(errs))
		// Simple aggregation: combine messages
		errMsg := ""
		for i, e := range errs {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += e.Error()
		}
		aggErr := errors.New(errMsg)
		// Return the aggregated error in both parts of the tuple
		return agentflow.AgentResult{Error: aggErr.Error()}, aggErr
	}

	log.Printf("CollaborativeOrchestrator: Finished dispatch for event ID %s successfully", event.GetID())
	return agentflow.AgentResult{}, nil // Return empty result and nil error on success
}

// DispatchAll needs the same signature update and logic adjustment
func (o *CollaborativeOrchestrator) DispatchAll(ctx context.Context, event agentflow.Event) (agentflow.AgentResult, error) {
	// For now, delegate to Dispatch.
	return o.Dispatch(ctx, event)
}

// Stop is a placeholder for potential cleanup tasks.
func (o *CollaborativeOrchestrator) Stop() {
	log.Println("CollaborativeOrchestrator stopping...")
	// No specific resources to clean up in this basic implementation.
}

// GetCallbackRegistry returns nil as CollaborativeOrchestrator doesn't manage callbacks directly.
// Callbacks are typically managed by the Runner.
func (o *CollaborativeOrchestrator) GetCallbackRegistry() *agentflow.CallbackRegistry {
	return nil // Or return the registry passed during creation if designed differently
}
