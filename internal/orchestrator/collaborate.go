package orchestrator

import (
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
// It aggregates any errors returned by the handlers.
func (o *CollaborativeOrchestrator) Dispatch(event agentflow.Event) []error {
	// FIX: Add nil check for the event
	if event == nil {
		log.Println("CollaborativeOrchestrator: Received nil event, skipping dispatch.")
		// Return an error or an empty slice depending on desired behavior for nil events
		// Returning an empty slice might be less surprising than returning an error.
		return nil // Or return []error{fmt.Errorf("cannot dispatch nil event")}
	}

	o.mu.RLock()
	handlersCopy := make([]agentflow.EventHandler, len(o.handlers))
	copy(handlersCopy, o.handlers)
	o.mu.RUnlock()

	if len(handlersCopy) == 0 {
		log.Printf("CollaborativeOrchestrator: No handlers registered, skipping dispatch for event ID %s", event.GetID())
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(handlersCopy))
	wg.Add(len(handlersCopy))

	log.Printf("CollaborativeOrchestrator: Dispatching event ID %s to %d handlers", event.GetID(), len(handlersCopy))

	for _, handler := range handlersCopy {
		go func(h agentflow.EventHandler) {
			defer wg.Done()
			// FIX: Add nil check for handler as well (belt-and-suspenders)
			if h == nil {
				log.Printf("CollaborativeOrchestrator: Encountered nil handler during dispatch for event ID %s", event.GetID())
				errChan <- fmt.Errorf("encountered nil handler")
				return
			}
			// The type assertion check might be redundant if RegisterAgent prevents non-EventHandler types
			if handlerWithHandle, ok := h.(interface{ Handle(agentflow.Event) error }); ok {
				if err := handlerWithHandle.Handle(event); err != nil {
					log.Printf("CollaborativeOrchestrator: Handler error for event ID %s: %v", event.GetID(), err)
					errChan <- err
				}
			} else {
				// This case should ideally not happen
				errChan <- fmt.Errorf("handler does not implement Handle(Event) error")
			}
		}(handler)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		log.Printf("CollaborativeOrchestrator: Finished dispatch for event ID %s with %d errors", event.GetID(), len(errs))
	}

	return errs
}

// DispatchAll sends the event to all registered handlers concurrently.
// It's similar to Dispatch but might have different semantics (e.g., error handling).
// Currently, it's identical to Dispatch. Consider merging or differentiating.
func (o *CollaborativeOrchestrator) DispatchAll(event agentflow.Event) []error {
	// For now, delegate to Dispatch. If different logic is needed, implement here.
	return o.Dispatch(event)
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
