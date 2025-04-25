package agentflow

// Orchestrator determines which AgentHandler should process an event.
type Orchestrator interface {
	// Dispatch routes the event to the appropriate handler(s).
	// It might return an error if routing fails or the handler fails synchronously.
	Dispatch(event Event) error

	// RegisterAgent associates a name with a handler.
	// Uses AgentHandler now.
	RegisterAgent(name string, handler AgentHandler) error

	// GetCallbackRegistry retrieves the registry associated with this orchestrator/runner.
	// This registry should be the same instance used by the Runner.
	GetCallbackRegistry() *CallbackRegistry

	// Stop performs any necessary cleanup for the orchestrator (e.g., closing connections).
	Stop()
}
