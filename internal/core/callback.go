package agentflow

import (
	"context"
	"fmt"
	"sync"
)

// HookPoint defines specific points in the execution flow where callbacks can be triggered.
type HookPoint string

const (
	// HookBeforeEventHandling is triggered before any processing of an incoming event begins.
	HookBeforeEventHandling HookPoint = "BeforeEventHandling"
	// HookAfterEventHandling is triggered after all processing for an event is complete.
	HookAfterEventHandling HookPoint = "AfterEventHandling"
	// HookBeforeAgentRun is triggered just before an agent's Run method is called.
	HookBeforeAgentRun HookPoint = "BeforeAgentRun"
	// HookAfterAgentRun is triggered just after an agent's Run method completes (successfully or with error).
	HookAfterAgentRun HookPoint = "AfterAgentRun"
	// FIX: Define HookAgentError
	HookAgentError HookPoint = "AgentError" // Triggered specifically when an agent Run returns an error
	HookAll        HookPoint = "AllHooks"   // Special hook for callbacks that run on all points
)

// CallbackArgs encapsulates all arguments passed to a callback function.
type CallbackArgs struct {
	Ctx         context.Context
	Hook        HookPoint
	Event       Event       // The event that triggered the hook
	State       State       // The current state, can be modified by callbacks
	AgentID     string      // ID of the agent involved (for agent-related hooks)
	AgentResult AgentResult // Result of the agent execution (for AfterAgentRun, AgentError) // <<< RENAMED from Result
	Output      AgentResult // <<< ADDED (Alias for AgentResult for consistency if needed elsewhere)
	Error       error       // Error information (for AgentError)
}

// CallbackFunc defines the signature for callback functions.
// It receives the context and all relevant arguments for the specific hook.
// It can optionally return a new State to replace the current one, or an error.
type CallbackFunc func(ctx context.Context, args CallbackArgs) (State, error)

// CallbackRegistration holds details about a registered callback.
type CallbackRegistration struct {
	ID           string       // Unique name for the callback within its hook point
	Hook         HookPoint    // The hook point this callback is registered for
	CallbackFunc CallbackFunc // The function to execute
	// Add AgentName if callbacks are agent-specific
	AgentName string // Optional: Name of the agent this callback is associated with
}

// CallbackRegistry manages registered callback functions.
type CallbackRegistry struct {
	mu        sync.RWMutex
	callbacks map[HookPoint][]*CallbackRegistration // Store pointers to registrations
}

// NewCallbackRegistry creates a new callback registry.
func NewCallbackRegistry() *CallbackRegistry {
	return &CallbackRegistry{
		callbacks: make(map[HookPoint][]*CallbackRegistration),
	}
}

// Register adds a callback function for a specific hook point.
func (r *CallbackRegistry) Register(hook HookPoint, name string, cb CallbackFunc) error {
	if name == "" {
		return fmt.Errorf("callback name cannot be empty")
	}
	if cb == nil {
		return fmt.Errorf("callback function cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	registration := &CallbackRegistration{ // Store pointer
		ID:           name,
		Hook:         hook,
		CallbackFunc: cb,
		// AgentName: agentName, // Set if needed
	}

	// Check for duplicates
	for _, existing := range r.callbacks[hook] {
		if existing.ID == name {
			return fmt.Errorf("callback '%s' already registered for hook '%s'", name, hook)
		}
	}

	r.callbacks[hook] = append(r.callbacks[hook], registration)
	Logger().Debug().
		Str("callback", name).
		Str("hook", string(hook)).
		Msg("Callback registered")
	return nil
}

// Unregister removes a callback function.
func (r *CallbackRegistry) Unregister(hook HookPoint, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hooks := r.callbacks[hook]
	for i, reg := range hooks {
		if reg.ID == name {
			// Remove the element
			r.callbacks[hook] = append(hooks[:i], hooks[i+1:]...)
			Logger().Debug().
				Str("callback", name).
				Str("hook", string(hook)).
				Msg("Callback unregistered")
			return
		}
	}
	Logger().Warn().
		Str("callback", name).
		Str("hook", string(hook)).
		Msg("Callback not found during unregister")
}

// Invoke calls all registered callbacks for a specific hook and HookAll.
// It propagates the state between callbacks and returns the final state.
func (r *CallbackRegistry) Invoke(ctx context.Context, args CallbackArgs) (State, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	currentState := args.State
	if currentState == nil {
		currentState = &SimpleState{data: make(map[string]interface{})}
		Logger().Warn().
			Str("hook", string(args.Hook)).
			Msg("Initial state was nil, created new SimpleState")
	}

	// Get registrations for the specific hook and HookAll
	hookRegistrations := r.callbacks[args.Hook]
	allRegistrations := r.callbacks[HookAll]

	// Combine callbacks: specific first, then HookAll
	// Extract the CallbackFunc from each registration
	callbacksToRun := make([]CallbackFunc, 0, len(hookRegistrations)+len(allRegistrations))
	for _, reg := range hookRegistrations {
		if reg != nil { // Add nil check for safety
			callbacksToRun = append(callbacksToRun, reg.CallbackFunc)
		}
	}
	for _, reg := range allRegistrations {
		if reg != nil { // Add nil check for safety
			callbacksToRun = append(callbacksToRun, reg.CallbackFunc)
		}
	}

	var lastErr error
	// Iterate directly over the combined CallbackFunc slice
	for _, callback := range callbacksToRun {
		// Create args for this specific call, ensuring the latest state is passed
		currentArgs := args
		currentArgs.State = currentState // Pass the current state

		Logger().Debug().
			Str("hook", string(args.Hook)).
			Msg("Executing callback")

		returnedState, err := callback(ctx, currentArgs) // Call the function directly
		if err != nil {
			// Decide how to handle errors - log and continue, or stop?
			Logger().Error().
				Str("hook", string(args.Hook)).
				Err(err).
				Msg("Error executing callback")
			lastErr = err // Store the last error
		}

		// Update currentState only if the callback returned a non-nil state
		if returnedState != nil {
			Logger().Debug().
				Str("hook", string(args.Hook)).
				Msg("Callback returned updated state")
			currentState = returnedState
		} else {
			Logger().Debug().
				Str("hook", string(args.Hook)).
				Msg("Callback returned nil state, state remains unchanged")
		}
	}

	Logger().Debug().
		Str("hook", string(args.Hook)).
		Msg("Finished invoking callbacks, returning final state")
	return currentState, lastErr
}
