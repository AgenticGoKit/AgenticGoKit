package agentflow

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// HookPoint defines specific moments in the agent/runner lifecycle where callbacks can be invoked.
type HookPoint string

const (
	// Runner Hooks
	HookBeforeEventHandling HookPoint = "BeforeEventHandling" // Before an event is routed to any handler
	HookAfterEventHandling  HookPoint = "AfterEventHandling"  // After an event has been handled (or failed)

	// Agent Hooks (invoked by Runner during handling)
	HookBeforeAgentRun HookPoint = "BeforeAgentRun" // Before an agent's Run method is called
	HookAfterAgentRun  HookPoint = "AfterAgentRun"  // After an agent's Run method completes (successfully or with error state)

	// HookAll is a special value used to register a callback for all hook points.
	// Note: The CallbackRegistry handles iterating through specific hooks when HookAll is used.
	HookAll HookPoint = "AllHooks"

	// TODO: Add more granular hooks later if needed:
	// HookBeforeModelCall HookPoint = "BeforeModelCall"
	// HookAfterModelCall  HookPoint = "AfterModelCall"
	// HookBeforeToolCall  HookPoint = "BeforeToolCall"
	// HookAfterToolCall   HookPoint = "AfterToolCall"
	// HookOnStateChange   HookPoint = "OnStateChange" // Might be tricky to implement efficiently
)

// CallbackArgs holds the arguments passed to a callback function.
// Different hook points might populate different fields.
type CallbackArgs struct {
	Ctx   context.Context
	Hook  HookPoint
	Event Event // Use the interface type directly
	State State
	// Add other relevant context as needed
}

// CallbackFunc defines the signature for callback functions.
// It receives the context, the current state, and the event that triggered the hook.
// It can optionally return a new State to replace the current one, or an error.
// FIX: Accept Event interface type directly
type CallbackFunc func(ctx context.Context, currentState State, event Event) (newState State, err error)

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
	log.Printf("Callback '%s' registered for hook '%s'", name, hook)
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
			log.Printf("Callback '%s' unregistered from hook '%s'", name, hook)
			return
		}
	}
	log.Printf("Warning: Callback '%s' not found for hook '%s' during unregister", name, hook)
}

// Invoke calls all registered callbacks for a given hook point.
// It passes necessary arguments like context, event, state, error.
func (r *CallbackRegistry) Invoke(args CallbackArgs) {
	r.mu.RLock()
	// Get specific hooks and potentially 'AllHooks'
	specificHooks := r.callbacks[args.Hook]
	allHooks := r.callbacks[HookAll]
	combinedHooks := append([]*CallbackRegistration{}, specificHooks...) // Create copy
	combinedHooks = append(combinedHooks, allHooks...)
	r.mu.RUnlock()

	if len(combinedHooks) == 0 {
		return // No callbacks for this hook
	}

	//log.Printf("Invoking %d callbacks for hook '%s'", len(combinedHooks), args.Hook)

	for _, reg := range combinedHooks {
		//eventID := "nil"
		// FIX: Remove eventForCallback variable
		// var eventForCallback *Event // Variable to hold the event pointer for the callback
		// if args.Event != nil {
		// 	//eventID = (*args.Event).GetID() // Dereference for GetID
		// 	eventForCallback = args.Event // Use the original pointer for the callback
		// }

		//log.Printf("Invoking callback '%s' for hook '%s', event '%s'", reg.ID, args.Hook, eventID)

		// Call the actual callback function, passing the interface value directly
		// FIX: Pass args.Event directly and use args.State
		returnedState, callbackErr := reg.CallbackFunc(args.Ctx, args.State, args.Event) // Pass interface value
		if callbackErr != nil {
			log.Printf("Error executing callback '%s' for hook '%s': %v", reg.ID, args.Hook, callbackErr)
			// Optionally: Store/aggregate errors? Stop invoking further callbacks?
		}
		if returnedState != nil {
			log.Printf("Callback '%s' returned new state for hook '%s'.", reg.ID, args.Hook)
			// FIX: Use args.State
			args.State = returnedState // Update state for subsequent callbacks in this invocation
		}
	}
	// Note: The updated state (args.CurrentState) is not automatically propagated back
	// to the caller (e.g., the Runner loop) unless the caller explicitly uses it.
}
