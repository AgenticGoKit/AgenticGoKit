// Package callbacks provides internal callback registry implementations for AgentFlow.
package callbacks

import (
	"context"
	"fmt"
	"sync"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// DefaultCallbackRegistry manages registered callback functions.
type DefaultCallbackRegistry struct {
	mu        sync.RWMutex
	callbacks map[core.HookPoint][]*core.CallbackRegistration
}

// NewDefaultCallbackRegistry creates a new callback registry.
func NewDefaultCallbackRegistry() *DefaultCallbackRegistry {
	return &DefaultCallbackRegistry{
		callbacks: make(map[core.HookPoint][]*core.CallbackRegistration),
	}
}

// Register adds a callback function for a specific hook point.
func (r *DefaultCallbackRegistry) Register(hook core.HookPoint, name string, cb core.CallbackFunc) error {
	if name == "" {
		return fmt.Errorf("callback name cannot be empty")
	}
	if cb == nil {
		return fmt.Errorf("callback function cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	registration := &core.CallbackRegistration{
		ID:           name,
		Hook:         hook,
		CallbackFunc: cb,
	}

	for _, existing := range r.callbacks[hook] {
		if existing.ID == name {
			return fmt.Errorf("callback '%s' already registered for hook '%s'", name, hook)
		}
	}

	r.callbacks[hook] = append(r.callbacks[hook], registration)
	core.Logger().Debug().
		Str("callback", name).
		Str("hook", string(hook)).
		Msg("Callback registered")
	return nil
}

// Unregister removes a callback function.
func (r *DefaultCallbackRegistry) Unregister(hook core.HookPoint, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hooks := r.callbacks[hook]
	for i, reg := range hooks {
		if reg.ID == name {
			r.callbacks[hook] = append(hooks[:i], hooks[i+1:]...)
			core.Logger().Info().
				Str("callback", name).
				Str("hook", string(hook)).
				Msg("Callback unregistered")
			return
		}
	}
	core.Logger().Warn().
		Str("callback", name).
		Str("hook", string(hook)).
		Msg("Callback not found during unregister")
}

// Invoke calls all registered callbacks for a specific hook and HookAll.
func (r *DefaultCallbackRegistry) Invoke(ctx context.Context, args core.CallbackArgs) (core.State, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	currentState := args.State
	if currentState == nil {
		core.Logger().Warn().
			Str("hook", string(args.Hook)).
			Msg("Initial state was nil, created new State")
		currentState = core.NewState()
	}

	hookRegistrations := r.callbacks[args.Hook]
	allRegistrations := r.callbacks[core.HookAll]

	callbacksToRun := make([]core.CallbackFunc, 0, len(hookRegistrations)+len(allRegistrations))
	for _, reg := range hookRegistrations {
		if reg != nil {
			callbacksToRun = append(callbacksToRun, reg.CallbackFunc)
		}
	}
	for _, reg := range allRegistrations {
		if reg != nil {
			callbacksToRun = append(callbacksToRun, reg.CallbackFunc)
		}
	}

	var lastErr error
	for _, callback := range callbacksToRun {
		currentArgs := args
		currentArgs.State = currentState

		core.Logger().Debug().
			Str("hook", string(args.Hook)).
			Msg("Executing callback")

		returnedState, err := callback(ctx, currentArgs)
		if err != nil {
			core.Logger().Error().
				Str("hook", string(args.Hook)).
				Err(err).
				Msg("Error executing callback")
			lastErr = err
		}

		if returnedState != nil {
			core.Logger().Debug().
				Str("hook", string(args.Hook)).
				Msg("Callback returned updated state")
			currentState = returnedState
		} else {
			core.Logger().Debug().
				Str("hook", string(args.Hook)).
				Msg("Callback returned nil state, state remains unchanged")
		}
	}

	core.Logger().Debug().
		Str("hook", string(args.Hook)).
		Msg("Finished invoking callbacks, returning final state")
	return currentState, lastErr
}
