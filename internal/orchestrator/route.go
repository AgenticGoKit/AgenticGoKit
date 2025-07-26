package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/internal/core"
)

// RouteMetadataKey is the key used in event metadata to specify a target handler name.
const RouteMetadataKey = agenticgokit.RouteMetadataKey // Use the agenticgokit constant

// RouteOrchestrator routes events to a single registered handler based on metadata.
type RouteOrchestrator struct {
	handlers map[string]agenticgokit.AgentHandler
	registry *agenticgokit.CallbackRegistry
	emitter  EventEmitter // Interface for emitting events
	mu       sync.RWMutex
}

// EventEmitter is an interface for components that can emit events
type EventEmitter interface {
	Emit(event agenticgokit.Event) error
}

// NewRouteOrchestrator creates a simple routing orchestrator.
// It requires the CallbackRegistry from the Runner.
func NewRouteOrchestrator(registry *agenticgokit.CallbackRegistry) *RouteOrchestrator { // Accept registry
	if registry == nil {
		agenticgokit.Logger().Warn().Msg("NewRouteOrchestrator created with a nil CallbackRegistry")
	}
	return &RouteOrchestrator{
		handlers: make(map[string]agenticgokit.AgentHandler),
		registry: registry, // Store the registry
	}
}

// RegisterAgent adds an agent handler.
func (o *RouteOrchestrator) RegisterAgent(agentID string, handler agenticgokit.AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if handler == nil {
		agenticgokit.Logger().Warn().
			Str("agent_id", agentID).
			Msg("Attempted to register a nil handler")
		return fmt.Errorf("cannot register a nil handler for agent ID '%s'", agentID)
	}
	if _, exists := o.handlers[agentID]; exists {
		agenticgokit.Logger().Warn().
			Str("agent_id", agentID).
			Msg("Overwriting handler for agent")
	}
	o.handlers[agentID] = handler
	agenticgokit.Logger().Debug().
		Str("agent_id", agentID).
		Msg("RouteOrchestrator: Registered agent")
	return nil
}

// Dispatch routes the event based on the RouteMetadataKey and executes the agent.
func (o *RouteOrchestrator) Dispatch(ctx context.Context, event agenticgokit.Event) (agenticgokit.AgentResult, error) {
	if event == nil {
		err := errors.New("cannot dispatch nil event")
		return agenticgokit.AgentResult{Error: err.Error()}, err
	}

	o.mu.RLock() // Lock for reading handlers map

	targetName, targetNameOK := event.GetMetadataValue(agenticgokit.RouteMetadataKey)
	if !targetNameOK {
		o.mu.RUnlock()
		err := fmt.Errorf("missing routing key '%s' in event metadata (event %s)", agenticgokit.RouteMetadataKey, event.GetID())
		agenticgokit.Logger().Error().
			Str("event_id", event.GetID()).
			Str("route_key", agenticgokit.RouteMetadataKey).
			Msgf("RouteOrchestrator: Error - %v", err)
		return agenticgokit.AgentResult{Error: err.Error()}, err
	}

	handler, exists := o.handlers[targetName]
	o.mu.RUnlock()

	if !exists {
		err := fmt.Errorf("no agent handler registered for target '%s' (event %s)", targetName, event.GetID())
		agenticgokit.Logger().Error().
			Str("event_id", event.GetID()).
			Str("target", targetName).
			Msgf("RouteOrchestrator: Error - %v", err)
		return agenticgokit.AgentResult{Error: err.Error()}, err
	}

	var agentResult agenticgokit.AgentResult
	var agentErr error
	var currentState agenticgokit.State = agenticgokit.NewState()

	// 1. Invoke BeforeAgentRun hooks
	if o.registry != nil {
		beforeArgs := agenticgokit.CallbackArgs{Ctx: ctx, Hook: agenticgokit.HookBeforeAgentRun, Event: event, State: currentState, AgentID: targetName}
		newState, hookErr := o.registry.Invoke(ctx, beforeArgs)
		if hookErr != nil {
			agenticgokit.Logger().Error().
				Str("agent_id", targetName).
				Err(hookErr).
				Msg("RouteOrchestrator: Error in BeforeAgentRun hooks")
		}
		if newState != nil {
			currentState = newState
		}
	}

	// Merge event data into the current state
	eventData := event.GetData()
	if eventData != nil {
		agenticgokit.Logger().Debug().
			Str("agent_id", targetName).
			Msg("RouteOrchestrator: Merging event data into state")
		for key, value := range eventData {
			currentState.Set(key, value)
		}
	}

	// 2. Run the agent handler
	agenticgokit.Logger().Debug().
		Str("agent_id", targetName).
		Str("event_id", event.GetID()).
		Interface("state_keys", currentState.Keys()).
		Msg("RouteOrchestrator: Running agent")
	agentResult, agentErr = handler.Run(ctx, event, currentState)

	// 3. Invoke AfterAgentRun hooks (always, even on error)
	if o.registry != nil {
		var stateForAfterHook agenticgokit.State = currentState
		if agentErr == nil && agentResult.OutputState != nil {
			stateForAfterHook = agentResult.OutputState
		}

		afterArgs := agenticgokit.CallbackArgs{
			Ctx:         ctx,
			Hook:        agenticgokit.HookAfterAgentRun,
			Event:       event,
			State:       stateForAfterHook,
			AgentID:     targetName,
			AgentResult: agentResult,
			Error:       agentErr,
		}
		if agentErr != nil {
			afterArgs.Hook = agenticgokit.HookAgentError
		}

		finalStateFromHooks, hookErr := o.registry.Invoke(ctx, afterArgs)
		if hookErr != nil {
			agenticgokit.Logger().Error().
				Str("agent_id", targetName).
				Str("hook", string(afterArgs.Hook)).
				Err(hookErr).
				Msg("RouteOrchestrator: Error in after hooks")
		}
		_ = finalStateFromHooks
	}

	// Ensure Routing Metadata Consistency
	if agentErr == nil && agentResult.OutputState != nil {
		if newRoute, hasNewRoute := agentResult.OutputState.GetMeta(RouteMetadataKey); hasNewRoute && newRoute != "" {
			fixedEvent := o.EnsureProperRouting(event, agentResult)
			if fixedEvent != event {
				currentRoute, hasCurrentRoute := event.GetMetadataValue(RouteMetadataKey)
				routeDisplay := "<none>"
				if hasCurrentRoute {
					routeDisplay = currentRoute
				}

				agenticgokit.Logger().Debug().
					Str("from", routeDisplay).
					Str("to", newRoute).
					Msg("RouteOrchestrator: Processing route change")

				if o.emitter != nil {
					if err := o.emitter.Emit(fixedEvent); err != nil {
						agenticgokit.Logger().Error().
							Str("to", newRoute).
							Err(err).
							Msg("RouteOrchestrator: Failed to emit new event with updated routing")
					} else {
						agenticgokit.Logger().Debug().
							Str("to", newRoute).
							Msg("RouteOrchestrator: Successfully queued event with updated routing")
					}
				} else {
					agenticgokit.Logger().Warn().
						Str("to", newRoute).
						Msg("RouteOrchestrator: No emitter available to queue event with updated routing")
				}
			}
		}
	}

	return agentResult, agentErr
}

// EnsureProperRouting ensures that agent result state metadata is correctly reflected in event routing
func (o *RouteOrchestrator) EnsureProperRouting(event agenticgokit.Event, result agenticgokit.AgentResult) agenticgokit.Event {
	if result.OutputState == nil {
		return event
	}

	newRoute, hasNewRoute := result.OutputState.GetMeta(RouteMetadataKey)
	if !hasNewRoute || newRoute == "" {
		return event
	}

	currentRoute, hasCurrentRoute := event.GetMetadataValue(RouteMetadataKey)
	if !hasCurrentRoute {
		currentRoute = ""
		agenticgokit.Logger().Warn().
			Str("new_route", newRoute).
			Msg("RouteOrchestrator: No routing metadata in current event, will create new event with route")
	}

	if currentRoute != newRoute {
		agenticgokit.Logger().Debug().
			Str("from", currentRoute).
			Str("to", newRoute).
			Msg("RouteOrchestrator: Detected routing change, creating new event")

		stateData := make(map[string]interface{})
		for _, key := range result.OutputState.Keys() {
			if val, ok := result.OutputState.Get(key); ok {
				stateData[key] = val
			}
		}

		newMeta := make(map[string]string)
		if meta := event.GetMetadata(); meta != nil {
			for k, v := range meta {
				newMeta[k] = v
			}
		}

		newMeta[RouteMetadataKey] = newRoute

		newEvent := agenticgokit.NewEvent(
			newRoute,
			stateData,
			newMeta,
		)

		newEvent.SetID(fmt.Sprintf("%s-route-%s", event.GetID(), newRoute))

		return newEvent
	}

	return event
}

// GetCallbackRegistry returns the associated registry.
func (o *RouteOrchestrator) GetCallbackRegistry() *agenticgokit.CallbackRegistry {
	return o.registry
}

// SetEmitter sets the event emitter for the orchestrator
func (o *RouteOrchestrator) SetEmitter(emitter EventEmitter) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.emitter = emitter
	agenticgokit.Logger().Debug().Msg("RouteOrchestrator: Emitter configured successfully")
}

// Stop performs cleanup (currently none needed for RouteOrchestrator).
func (o *RouteOrchestrator) Stop() {
	agenticgokit.Logger().Info().Msg("RouteOrchestrator stopping.")
}

// DispatchAll implements the Orchestrator interface but provides more graceful handling
// for events without routing information.
func (o *RouteOrchestrator) DispatchAll(ctx context.Context, event agenticgokit.Event) (agenticgokit.AgentResult, error) {
	_, hasRouting := event.GetMetadataValue(agenticgokit.RouteMetadataKey)

	if !hasRouting {
		if errorData, hasError := event.GetData()["error"]; hasError {
			agenticgokit.Logger().Warn().
				Str("event_id", event.GetID()).
				Msg("RouteOrchestrator: Processing error event without routing key")

			return agenticgokit.AgentResult{
				OutputState: agenticgokit.NewState(),
				Error:       fmt.Sprintf("Error event processed: %v", errorData),
			}, nil
		}
	}

	agenticgokit.Logger().Debug().
		Str("event_id", event.GetID()).
		Msg("RouteOrchestrator: DispatchAll forwarding to Dispatch")
	return o.Dispatch(ctx, event)
}