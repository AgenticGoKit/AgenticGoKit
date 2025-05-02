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
const RouteMetadataKey = agentflow.RouteMetadataKey // Use the agentflow constant

// RouteOrchestrator routes events to a single registered handler based on metadata.
type RouteOrchestrator struct {
	handlers map[string]agentflow.AgentHandler
	registry *agentflow.CallbackRegistry
	emitter  EventEmitter // Interface for emitting events
	mu       sync.RWMutex
}

// EventEmitter is an interface for components that can emit events
type EventEmitter interface {
	Emit(event agentflow.Event) error
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
// FIX: Ensure this method returns an error to match the agentflow.Orchestrator interface
func (o *RouteOrchestrator) RegisterAgent(agentID string, handler agentflow.AgentHandler) error { // <<< MUST return error
	o.mu.Lock()
	defer o.mu.Unlock()
	if handler == nil {
		log.Printf("Warning: Attempted to register a nil handler for agent ID '%s'", agentID)
		// FIX: Return an error for nil handler
		return fmt.Errorf("cannot register a nil handler for agent ID '%s'", agentID)
	}
	if _, exists := o.handlers[agentID]; exists {
		log.Printf("Warning: Overwriting handler for agent '%s'", agentID)
		// Optionally return an error here if overwriting is disallowed
		// return fmt.Errorf("agent '%s' already registered", agentID)
	}
	o.handlers[agentID] = handler
	log.Printf("RouteOrchestrator: Registered agent '%s'", agentID)
	return nil // <<< MUST return nil on success
}

// Dispatch routes the event based on the RouteMetadataKey and executes the agent.
// FIX: Update signature to return AgentResult, error
func (o *RouteOrchestrator) Dispatch(ctx context.Context, event agentflow.Event) (agentflow.AgentResult, error) {
	if event == nil {
		err := errors.New("cannot dispatch nil event")
		return agentflow.AgentResult{Error: err.Error()}, err
	}

	o.mu.RLock() // Lock for reading handlers map

	// FIX: Use GetMetadataValue and constant
	targetName, targetNameOK := event.GetMetadataValue(agentflow.RouteMetadataKey)
	// FIX: Return specific error if routing key is missing
	if !targetNameOK {
		o.mu.RUnlock() // Unlock before returning
		err := fmt.Errorf("missing routing key '%s' in event metadata (event %s)", agentflow.RouteMetadataKey, event.GetID())
		log.Printf("RouteOrchestrator: Error - %v", err)
		return agentflow.AgentResult{Error: err.Error()}, err
	}

	handler, exists := o.handlers[targetName]
	o.mu.RUnlock() // Unlock after accessing handlers map

	if !exists {
		err := fmt.Errorf("no agent handler registered for target '%s' (event %s)", targetName, event.GetID())
		log.Printf("RouteOrchestrator: Error - %v", err)
		return agentflow.AgentResult{Error: err.Error()}, err
	}

	// --- Agent Execution with Hooks ---
	var agentResult agentflow.AgentResult
	var agentErr error
	// FIX: Initialize currentState as the interface type
	var currentState agentflow.State = agentflow.NewState() // Example: Start with empty state

	// 1. Invoke BeforeAgentRun hooks
	if o.registry != nil {
		beforeArgs := agentflow.CallbackArgs{Ctx: ctx, Hook: agentflow.HookBeforeAgentRun, Event: event, State: currentState, AgentID: targetName}
		// FIX: Assign returned state interface directly
		newState, hookErr := o.registry.Invoke(ctx, beforeArgs) // Pass ctx
		if hookErr != nil {
			// Handle hook error appropriately - log, maybe return?
			log.Printf("RouteOrchestrator: Error in BeforeAgentRun hooks for agent '%s': %v", targetName, hookErr)
			// Decide if this should halt execution
		}
		if newState != nil { // Check if hook returned a new state
			currentState = newState // Update currentState (which is agentflow.State)
		}
	}

	// <<< START FIX: Merge event data into the current state >>>
	eventData := event.GetData()
	if eventData != nil {
		log.Printf("RouteOrchestrator: Merging event data into state for agent '%s'", targetName)
		for key, value := range eventData {
			// Avoid overwriting existing state keys from callbacks? Or allow event data to override?
			// Current approach: Event data overrides. Adjust if needed.
			currentState.Set(key, value)
		}
	}
	// <<< END FIX >>>

	// 2. Run the agent handler
	log.Printf("RouteOrchestrator: Running agent '%s' for event %s with state keys: %v", targetName, event.GetID(), currentState.Keys()) // Log state keys
	agentResult, agentErr = handler.Run(ctx, event, currentState)                                                                        // Pass context

	// 3. Invoke AfterAgentRun hooks (always, even on error)
	if o.registry != nil {
		// FIX: Use the state returned by the agent *if* the agent didn't error, otherwise use the state *before* the agent ran.
		// FIX: Declare stateForAfterHook as the interface type
		var stateForAfterHook agentflow.State = currentState // Default to state before agent run
		if agentErr == nil && agentResult.OutputState != nil {
			// FIX: Assign returned state interface directly
			stateForAfterHook = agentResult.OutputState // Use agent's output state (which is agentflow.State)
		}

		afterArgs := agentflow.CallbackArgs{
			Ctx:         ctx,
			Hook:        agentflow.HookAfterAgentRun, // Or HookAgentError if agentErr != nil
			Event:       event,
			State:       stateForAfterHook, // Pass the interface value
			AgentID:     targetName,
			AgentResult: agentResult, // Pass the full result
			Error:       agentErr,    // Pass the agent error
		}
		if agentErr != nil {
			afterArgs.Hook = agentflow.HookAgentError // Use specific hook on error
		}

		// FIX: Assign returned state interface directly (though often After hooks don't modify state)
		finalStateFromHooks, hookErr := o.registry.Invoke(ctx, afterArgs) // Pass ctx
		if hookErr != nil {
			log.Printf("RouteOrchestrator: Error in %s hooks for agent '%s': %v", afterArgs.Hook, targetName, hookErr)
		}
		// Decide how to handle state changes from After hooks, if necessary.
		// For simplicity, we might ignore state changes from After hooks unless specifically needed.
		_ = finalStateFromHooks // Avoid unused variable error if not using the result
	}
	// --- End Agent Execution ---

	// --- Azure Best Practice: Ensure Routing Metadata Consistency ---
	if agentErr == nil && agentResult.OutputState != nil {
		// Check if routing metadata was set in the result state
		if newRoute, hasNewRoute := agentResult.OutputState.GetMeta(RouteMetadataKey); hasNewRoute && newRoute != "" {
			// Create new event with proper routing for the next stage
			fixedEvent := o.EnsureProperRouting(event, agentResult)

			// If routing has changed (EnsureProperRouting returns a different event)
			if fixedEvent != event {
				// Fix: Properly handle multiple return values
				currentRoute, hasCurrentRoute := event.GetMetadataValue(RouteMetadataKey)
				routeDisplay := "<none>"
				if hasCurrentRoute {
					routeDisplay = currentRoute
				}

				log.Printf("RouteOrchestrator: Processing route change from '%s' to '%s'",
					routeDisplay, newRoute)

				// Queue the new event for processing
				// This requires access to a Runner or event emitter
				// If runner is available as a field, use:
				// if o.runner != nil {
				//     if err := o.runner.Emit(fixedEvent); err != nil {
				//         log.Printf("RouteOrchestrator: Failed to emit new event with updated routing: %v", err)
				//     }
				// }

				// Alternative: Add a RunnerEmitter field to RouteOrchestrator for this purpose
				if o.emitter != nil {
					if err := o.emitter.Emit(fixedEvent); err != nil {
						log.Printf("RouteOrchestrator: Failed to emit new event with updated routing: %v", err)
					} else {
						log.Printf("RouteOrchestrator: Successfully queued event with updated routing to '%s'", newRoute)
					}
				} else {
					log.Printf("RouteOrchestrator: No emitter available to queue event with updated routing")
				}
			}
		}
	}
	// --- End Azure Best Practice ---

	// FIX: Return the captured agentResult and agentErr
	return agentResult, agentErr
}

// Add this method right after the Dispatch method

// EnsureProperRouting ensures that agent result state metadata is correctly reflected in event routing
func (o *RouteOrchestrator) EnsureProperRouting(event agentflow.Event, result agentflow.AgentResult) agentflow.Event {
	// Skip if there's no output state
	if result.OutputState == nil {
		return event
	}

	// Check if the agent specified a new route in its output state
	newRoute, hasNewRoute := result.OutputState.GetMeta(RouteMetadataKey)
	if !hasNewRoute || newRoute == "" {
		return event // No routing change needed
	}

	// Get current routing from event metadata - FIX: Properly handle both return values
	currentRoute, hasCurrentRoute := event.GetMetadataValue(RouteMetadataKey)
	if !hasCurrentRoute {
		currentRoute = "" // Default to empty string if not found
		log.Printf("RouteOrchestrator: No routing metadata in current event, will create new event with route: %s", newRoute)
	}

	// Only create a new event if routing has changed
	if currentRoute != newRoute {
		log.Printf("RouteOrchestrator: Detected routing change from '%s' to '%s', creating new event",
			currentRoute, newRoute)

		// Get the data from the output state for the new event
		stateData := make(map[string]interface{})
		for _, key := range result.OutputState.Keys() {
			if val, ok := result.OutputState.Get(key); ok {
				stateData[key] = val
			}
		}

		// Create new metadata, preserving session ID and other metadata
		newMeta := make(map[string]string)
		if meta := event.GetMetadata(); meta != nil {
			for k, v := range meta {
				newMeta[k] = v
			}
		}

		// Set the new route
		newMeta[RouteMetadataKey] = newRoute

		// Create a new event with the updated routing
		newEvent := agentflow.NewEvent(
			newRoute,  // targetAgent
			stateData, // data from output state
			newMeta,   // metadata with updated route
		)

		// Preserve event ID chain for tracing
		newEvent.SetID(fmt.Sprintf("%s-route-%s", event.GetID(), newRoute))

		return newEvent
	}

	return event
}

// GetCallbackRegistry returns the associated registry.
func (o *RouteOrchestrator) GetCallbackRegistry() *agentflow.CallbackRegistry {
	return o.registry // Return the stored registry
}

// Add this method after GetCallbackRegistry
// SetEmitter sets the event emitter for the orchestrator
func (o *RouteOrchestrator) SetEmitter(emitter EventEmitter) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.emitter = emitter
	log.Println("RouteOrchestrator: Emitter configured successfully")
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

// DispatchAll implements the Orchestrator interface but provides more graceful handling
// for events without routing information.
func (o *RouteOrchestrator) DispatchAll(ctx context.Context, event agentflow.Event) (agentflow.AgentResult, error) {
	// Check if event has routing information
	_, hasRouting := event.GetMetadataValue(agentflow.RouteMetadataKey)

	if !hasRouting {
		// Check if this might be an error event (special case handling)
		if errorData, hasError := event.GetData()["error"]; hasError {
			log.Printf("RouteOrchestrator: Processing error event %s without routing key", event.GetID())

			// Special case for error events - create an empty result
			return agentflow.AgentResult{
				OutputState: agentflow.NewState(), // Empty but valid state
				Error:       fmt.Sprintf("Error event processed: %v", errorData),
			}, nil // Return nil error to break the cascade
		}
	}

	// Otherwise, delegate to normal Dispatch
	log.Printf("RouteOrchestrator: DispatchAll forwarding to Dispatch for event %s", event.GetID())
	return o.Dispatch(ctx, event)
}
