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

	// 2. Run the agent handler
	log.Printf("RouteOrchestrator: Running agent '%s' for event %s", targetName, event.GetID())
	agentResult, agentErr = handler.Run(ctx, event, currentState) // Pass context

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

	// FIX: Return the captured agentResult and agentErr
	return agentResult, agentErr
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
