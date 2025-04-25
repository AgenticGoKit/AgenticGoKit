package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// RouteMetadataKey is the key used in event metadata to specify a target handler name.
const RouteMetadataKey = "route_to"

// RouteOrchestrator routes events to a single registered handler based on metadata.
type RouteOrchestrator struct {
	handlers map[string]agentflow.AgentHandler
	registry *agentflow.CallbackRegistry // Ensure this field exists
	mu       sync.RWMutex
}

// NewRouteOrchestrator creates a simple routing orchestrator.
// It requires the CallbackRegistry from the Runner.
func NewRouteOrchestrator(registry *agentflow.CallbackRegistry) *RouteOrchestrator { // Accept registry
	if registry == nil {
		// Decide how to handle nil registry: panic, return error, or create a default one?
		// Creating a default one might hide issues, panic might be too harsh.
		// Let's log a warning and proceed with nil, but tests should provide one.
		log.Println("Warning: NewRouteOrchestrator created with a nil CallbackRegistry")
	}
	return &RouteOrchestrator{
		handlers: make(map[string]agentflow.AgentHandler),
		registry: registry, // Store the registry
	}
}

// RegisterAgent adds an agent handler.
func (o *RouteOrchestrator) RegisterAgent(agentID string, handler agentflow.AgentHandler) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if handler == nil {
		log.Printf("Warning: Attempted to register a nil handler for agent ID '%s'", agentID)
		return
	}
	o.handlers[agentID] = handler
	log.Printf("RouteOrchestrator: Registered agent '%s'", agentID)
}

// Dispatch routes the event to the appropriate handler based on TargetAgentID or metadata.
func (o *RouteOrchestrator) Dispatch(event agentflow.Event) error {
	if event == nil {
		// FIX: Use errors.New for static error messages
		return errors.New("RouteOrchestrator: received nil event")
	}

	targetAgentID := event.GetTargetAgentID()

	// Attempt to get target from metadata ONLY if explicit target is missing
	if targetAgentID == "" {
		// Ensure metadata is not nil before accessing it
		metadata := event.GetMetadata()
		if metadata != nil {
			// Use the defined constant RouteMetadataKey if available, otherwise "route"
			routeKey := RouteMetadataKey // Assuming RouteMetadataKey = "route" or similar
			if routeMeta, ok := metadata[routeKey]; ok && routeMeta != "" {
				targetAgentID = routeMeta
			}
		}
		// DO NOT add any fallback to event type here.
	}

	// log.Printf("DEBUG: targetAgentID before check is: '%s'", targetAgentID) // Keep for debugging if needed

	// Check IMMEDIATELY after determining targetAgentID if it's still empty.
	if targetAgentID == "" {
		// Use the specific error message substring the test expects.
		return errors.New("RouteOrchestrator: event has no target agent ID or route metadata key")
	}

	o.mu.RLock()
	handler, exists := o.handlers[targetAgentID]
	o.mu.RUnlock()

	if !exists {
		// This error should only occur if targetAgentID was found (non-empty), but no handler matches.
		return fmt.Errorf("RouteOrchestrator: no handler registered for agent '%s'", targetAgentID)
	}

	// Proceed with dispatch if handler exists
	// FIX: Use context.Background() for top-level context
	ctx := context.Background()
	// FIX: Create a new state for each dispatch if not passed in
	state := agentflow.NewState() // Assuming state is managed per dispatch

	log.Printf("RouteOrchestrator: Dispatching event %s to handler for agent '%s'", event.GetID(), targetAgentID)
	// FIX: Pass context and state to the handler's Run method
	result, err := handler.Run(ctx, event, state) // Assuming Run takes ctx, event, state
	if err != nil {
		// FIX: Log the specific error from the handler
		log.Printf("RouteOrchestrator: Error from handler '%s' processing event %s: %v", targetAgentID, event.GetID(), err)
		// FIX: Wrap the handler error for context
		return fmt.Errorf("handler '%s' failed processing event %s: %w", targetAgentID, event.GetID(), err) // Wrap error
	}

	// FIX: Log the result details using %+v
	log.Printf("RouteOrchestrator: Successfully dispatched event %s to handler '%s'. Result: %+v", event.GetID(), targetAgentID, result)
	return nil
}

// GetCallbackRegistry returns the associated registry.
func (o *RouteOrchestrator) GetCallbackRegistry() *agentflow.CallbackRegistry {
	return o.registry // Return the stored registry
}

// Stop performs cleanup (currently none needed for RouteOrchestrator).
func (o *RouteOrchestrator) Stop() {
	log.Println("RouteOrchestrator stopping.")
	// No specific resources (like network connections) to clean up in this simple version.
	// Clear handlers map? Optional, depends on desired reuse behavior.
	// o.mu.Lock()
	// o.handlers = make(map[string]agentflow.AgentHandler)
	// o.mu.Unlock()
}
