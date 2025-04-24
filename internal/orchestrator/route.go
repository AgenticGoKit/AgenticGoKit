package orchestrator

import (
	agentflow "kunalkushwaha/agentflow/internal/core"
	"log"
	"sync"
	"sync/atomic"
)

// RouteMetadataKey is the key used in event metadata to specify a target handler name.
const RouteMetadataKey = "agentflow_route_to"

// RouteOrchestrator delivers each event to one handler via round-robin,
// unless a specific handler is requested via metadata.
type RouteOrchestrator struct {
	mu          sync.RWMutex
	handlers    []agentflow.EventHandler          // Keep ordered list for round-robin
	handlerKeys []string                          // Keep ordered list for round-robin
	handlerMap  map[string]agentflow.EventHandler // Map for direct lookup by name
	next        uint64
}

// NewRouteOrchestrator creates a new round-robin orchestrator.
func NewRouteOrchestrator() *RouteOrchestrator {
	return &RouteOrchestrator{
		handlers:    make([]agentflow.EventHandler, 0),
		handlerKeys: make([]string, 0),
		handlerMap:  make(map[string]agentflow.EventHandler), // Initialize map
	}
}

// RegisterAgent adds an event handler.
func (r *RouteOrchestrator) RegisterAgent(name string, handler agentflow.EventHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Prevent duplicate registration names, which would break map lookup
	if _, exists := r.handlerMap[name]; exists {
		log.Printf("RouteOrchestrator: Warning - handler with name '%s' already registered. Overwriting.", name)
	}
	// Add to map first
	r.handlerMap[name] = handler

	// Keep track of order for round-robin (rebuild slices on registration)
	// This is less efficient but simpler than managing indices on deregistration (if added later)
	r.handlerKeys = make([]string, 0, len(r.handlerMap))
	r.handlers = make([]agentflow.EventHandler, 0, len(r.handlerMap))
	for key, h := range r.handlerMap {
		r.handlerKeys = append(r.handlerKeys, key)
		r.handlers = append(r.handlers, h)
	}
	// Note: Round-robin order might change slightly on registration if map iteration order differs.
	// If strict registration order is needed for round-robin, append logic needs adjustment.
}

// Dispatch sends the event to a specific handler if requested via metadata,
// otherwise sends to the next handler in the round-robin rotation.
func (r *RouteOrchestrator) Dispatch(event agentflow.Event) {
	r.mu.RLock() // Use RLock for reading map/slices initially

	// 1. Check for explicit routing via metadata
	targetHandlerName := ""
	if event.GetMetadata() != nil {
		targetHandlerName = event.GetMetadata()[RouteMetadataKey]
	}

	if targetHandlerName != "" {
		handler, found := r.handlerMap[targetHandlerName]
		if found {
			r.mu.RUnlock() // Unlock before calling handler
			log.Printf("RouteOrchestrator: Routing event %s directly to handler '%s' via metadata", event.GetID(), targetHandlerName)
			if err := handler.Handle(event); err != nil {
				log.Printf("RouteOrchestrator: Direct handler '%s' failed handling event %s: %v", targetHandlerName, event.GetID(), err)
			}
			return // Handled via metadata routing
		}
		// If target specified but not found, log and fall through to round-robin
		log.Printf("RouteOrchestrator: Target handler '%s' specified in metadata for event %s not found. Falling back to round-robin.", targetHandlerName, event.GetID())
	}

	// 2. Fallback to Round-Robin
	handlerCount := len(r.handlers) // Use slice length for round-robin index
	if handlerCount == 0 {
		r.mu.RUnlock()
		log.Printf("RouteOrchestrator: No handlers registered to handle event %s", event.GetID())
		return
	}

	// Use atomic counter for round-robin index based on the ordered slice
	idx := atomic.AddUint64(&r.next, 1) - 1
	handlerIndex := int(idx % uint64(handlerCount))
	handler := r.handlers[handlerIndex]        // Get handler from the ordered slice
	handlerName := r.handlerKeys[handlerIndex] // Get corresponding name for logging
	r.mu.RUnlock()                             // Unlock before calling handler

	log.Printf("RouteOrchestrator: Routing event %s via round-robin to handler '%s' (index %d)", event.GetID(), handlerName, handlerIndex)
	if err := handler.Handle(event); err != nil {
		log.Printf("RouteOrchestrator: Round-robin handler '%s' failed handling event %s: %v", handlerName, event.GetID(), err)
	}
}

// Stop is a no-op for this simple orchestrator.
func (r *RouteOrchestrator) Stop() {}
