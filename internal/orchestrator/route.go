package orchestrator

import (
	agentflow "kunalkushwaha/agentflow/internal/core"
	"log"
	"sync"
	"sync/atomic"
)

// RouteOrchestrator delivers each event to one handler via round-robin.
type RouteOrchestrator struct {
	mu          sync.RWMutex
	handlers    []agentflow.EventHandler // Use EventHandler, renamed field
	handlerKeys []string                 // Renamed field
	next        uint64
}

// NewRouteOrchestrator creates a new round-robin orchestrator.
func NewRouteOrchestrator() *RouteOrchestrator {
	return &RouteOrchestrator{
		handlers:    make([]agentflow.EventHandler, 0), // Use EventHandler
		handlerKeys: make([]string, 0),
	}
}

// RegisterAgent adds an event handler.
func (r *RouteOrchestrator) RegisterAgent(name string, handler agentflow.EventHandler) { // Use EventHandler
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers = append(r.handlers, handler)    // Use renamed field
	r.handlerKeys = append(r.handlerKeys, name) // Use renamed field
}

// Dispatch sends the event to the next handler in the rotation.
func (r *RouteOrchestrator) Dispatch(event agentflow.Event) {
	r.mu.RLock()
	handlerCount := len(r.handlers) // Use renamed field
	if handlerCount == 0 {
		r.mu.RUnlock()
		log.Printf("RouteOrchestrator: No handlers registered to handle event %s", event.GetID())
		return
	}

	idx := atomic.AddUint64(&r.next, 1) - 1
	handler := r.handlers[int(idx%uint64(handlerCount))] // Use renamed field
	r.mu.RUnlock()

	if err := handler.Handle(event); err != nil { // Call Handle on EventHandler
		log.Printf("RouteOrchestrator: Handler failed handling event %s: %v", event.GetID(), err)
	}
}

// Stop is a no-op for this simple orchestrator.
func (r *RouteOrchestrator) Stop() {}
