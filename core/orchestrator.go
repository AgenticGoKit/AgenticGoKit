// Package core provides the public Orchestrator interface and related types for AgentFlow.
package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// EventEmitter is an interface for components that can emit events
type EventEmitter interface {
	Emit(event Event) error
}

// Orchestrator defines the interface for routing events to agents.
type Orchestrator interface {
	// Dispatch the event to the appropriate agent.
	Dispatch(ctx context.Context, event Event) (AgentResult, error)
	// RegisterAgent registers a new agent with the given name and handler.
	RegisterAgent(name string, handler AgentHandler) error
	// GetCallbackRegistry returns the callback registry.
	GetCallbackRegistry() *CallbackRegistry
	// Stop halts the orchestrator.
	Stop()
}

// RouteOrchestrator routes events to a single registered handler based on metadata.
type RouteOrchestrator struct {
	handlers map[string]AgentHandler
	registry *CallbackRegistry
	emitter  EventEmitter // Interface for emitting events
	mu       sync.RWMutex
}

// NewRouteOrchestrator creates a simple routing orchestrator.
// It requires the CallbackRegistry from the Runner.
func NewRouteOrchestrator(registry *CallbackRegistry) *RouteOrchestrator {
	if registry == nil {
		Logger().Warn().Msg("NewRouteOrchestrator created with a nil CallbackRegistry")
	}
	return &RouteOrchestrator{
		handlers: make(map[string]AgentHandler),
		registry: registry,
	}
}

// RegisterAgent adds an agent handler.
func (o *RouteOrchestrator) RegisterAgent(agentID string, handler AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if handler == nil {
		Logger().Warn().
			Str("agent_id", agentID).
			Msg("Attempted to register a nil handler")
		return fmt.Errorf("cannot register a nil handler for agent ID '%s'", agentID)
	}
	if _, exists := o.handlers[agentID]; exists {
		Logger().Warn().
			Str("agent_id", agentID).
			Msg("Overwriting handler for agent")
	}
	o.handlers[agentID] = handler
	Logger().Debug().
		Str("agent_id", agentID).
		Msg("RouteOrchestrator: Registered agent")
	return nil
}

// Dispatch routes the event based on the RouteMetadataKey and executes the agent.
func (o *RouteOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
	if event == nil {
		err := errors.New("cannot dispatch nil event")
		return AgentResult{Error: err.Error()}, err
	}

	o.mu.RLock() // Lock for reading handlers map

	targetName, targetNameOK := event.GetMetadataValue(RouteMetadataKey)
	if !targetNameOK {
		o.mu.RUnlock()
		err := fmt.Errorf("missing routing key '%s' in event metadata (event %s)", RouteMetadataKey, event.GetID())
		Logger().Error().
			Str("event_id", event.GetID()).
			Str("route_key", RouteMetadataKey).
			Msgf("RouteOrchestrator: Error - %v", err)
		return AgentResult{Error: err.Error()}, err
	}

	handler, exists := o.handlers[targetName]
	o.mu.RUnlock()

	if !exists {
		err := fmt.Errorf("no agent handler registered for target '%s' (event %s)", targetName, event.GetID())
		Logger().Error().
			Str("event_id", event.GetID()).
			Str("target", targetName).
			Msgf("RouteOrchestrator: Error - %v", err)
		return AgentResult{Error: err.Error()}, err
	}

	var agentResult AgentResult
	var agentErr error
	var currentState State = NewState()

	// 1. Invoke BeforeAgentRun hooks
	if o.registry != nil {
		beforeArgs := CallbackArgs{Ctx: ctx, Hook: HookBeforeAgentRun, Event: event, State: currentState, AgentID: targetName}
		newState, hookErr := o.registry.Invoke(ctx, beforeArgs)
		if hookErr != nil {
			Logger().Error().
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
		Logger().Debug().
			Str("agent_id", targetName).
			Msg("RouteOrchestrator: Merging event data into state")
		for key, value := range eventData {
			currentState.Set(key, value)
		}
	}

	// 2. Run the agent handler
	Logger().Debug().
		Str("agent_id", targetName).
		Str("event_id", event.GetID()).
		Interface("state_keys", currentState.Keys()).
		Msg("RouteOrchestrator: Running agent")
	agentResult, agentErr = handler.Run(ctx, event, currentState)

	// 3. Invoke AfterAgentRun hooks (always, even on error)
	if o.registry != nil {
		var stateForAfterHook State = currentState
		if agentErr == nil && agentResult.OutputState != nil {
			stateForAfterHook = agentResult.OutputState
		}

		afterArgs := CallbackArgs{
			Ctx:         ctx,
			Hook:        HookAfterAgentRun,
			Event:       event,
			State:       stateForAfterHook,
			AgentID:     targetName,
			AgentResult: agentResult,
			Error:       agentErr,
		}
		if agentErr != nil {
			afterArgs.Hook = HookAgentError
		}

		finalStateFromHooks, hookErr := o.registry.Invoke(ctx, afterArgs)
		if hookErr != nil {
			Logger().Error().
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

				Logger().Debug().
					Str("from", routeDisplay).
					Str("to", newRoute).
					Msg("RouteOrchestrator: Processing route change")

				if o.emitter != nil {
					if err := o.emitter.Emit(fixedEvent); err != nil {
						Logger().Error().
							Str("to", newRoute).
							Err(err).
							Msg("RouteOrchestrator: Failed to emit new event with updated routing")
					} else {
						Logger().Debug().
							Str("to", newRoute).
							Msg("RouteOrchestrator: Successfully queued event with updated routing")
					}
				} else {
					Logger().Warn().
						Str("to", newRoute).
						Msg("RouteOrchestrator: No emitter available to queue event with updated routing")
				}
			}
		}
	}

	return agentResult, agentErr
}

// EnsureProperRouting ensures that agent result state metadata is correctly reflected in event routing
func (o *RouteOrchestrator) EnsureProperRouting(event Event, result AgentResult) Event {
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
		Logger().Warn().
			Str("new_route", newRoute).
			Msg("RouteOrchestrator: No routing metadata in current event, will create new event with route")
	}

	if currentRoute != newRoute {
		Logger().Debug().
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

		newEvent := NewEvent(
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
func (o *RouteOrchestrator) GetCallbackRegistry() *CallbackRegistry {
	return o.registry
}

// SetEmitter sets the event emitter for the orchestrator
func (o *RouteOrchestrator) SetEmitter(emitter EventEmitter) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.emitter = emitter
	Logger().Debug().Msg("RouteOrchestrator: Emitter configured successfully")
}

// Stop performs cleanup (currently none needed for RouteOrchestrator).
func (o *RouteOrchestrator) Stop() {
	Logger().Info().Msg("RouteOrchestrator stopping.")
}
