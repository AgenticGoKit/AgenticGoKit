package orchestrator

import (
	agentflow "kunalkushwaha/agentflow/internal/core"
	"log"
	"strings"
	"sync"
)

// CollaborateOrchestrator dispatches each event to all registered handlers concurrently.
type CollaborateOrchestrator struct {
	mu       sync.RWMutex
	handlers map[string]agentflow.EventHandler // Use EventHandler, renamed field
}

// NewCollaborateOrchestrator creates a new fan-out orchestrator.
func NewCollaborateOrchestrator() *CollaborateOrchestrator {
	return &CollaborateOrchestrator{
		handlers: make(map[string]agentflow.EventHandler), // Use EventHandler
	}
}

// RegisterAgent adds an event handler. Safe for concurrent use.
func (c *CollaborateOrchestrator) RegisterAgent(name string, handler agentflow.EventHandler) { // Use EventHandler
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[name] = handler // Use renamed field
}

// Dispatch sends the event to all registered handlers in parallel and collects errors.
func (c *CollaborateOrchestrator) Dispatch(event agentflow.Event) {
	c.mu.RLock()
	// Create a snapshot of handlers to call Handle on
	currentHandlers := make([]agentflow.EventHandler, 0, len(c.handlers)) // Use EventHandler
	for _, handler := range c.handlers {                                  // Use renamed field
		currentHandlers = append(currentHandlers, handler)
	}
	c.mu.RUnlock()

	if len(currentHandlers) == 0 {
		log.Printf("CollaborateOrchestrator: No handlers registered for event %s", event.GetID())
		return
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(currentHandlers))

	wg.Add(len(currentHandlers))
	for _, handler := range currentHandlers { // Use EventHandler
		go func(h agentflow.EventHandler) { // Use EventHandler
			defer wg.Done()
			if err := h.Handle(event); err != nil { // Call Handle on EventHandler
				errs <- err
			}
		}(handler)
	}

	wg.Wait()
	close(errs)

	var errorMessages []string
	for err := range errs {
		errorMessages = append(errorMessages, err.Error())
	}

	if len(errorMessages) > 0 {
		log.Printf("CollaborateOrchestrator: Errors handling event %s: %s",
			event.GetID(), strings.Join(errorMessages, "; "))
	}
}

// Stop is a no-op for this orchestrator.
func (c *CollaborateOrchestrator) Stop() {}
