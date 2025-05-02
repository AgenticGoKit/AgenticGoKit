package agentflow

import "context"

// Orchestrator defines the interface for routing events to agents.
type Orchestrator interface {
	// Dispatch routes an event to the appropriate agent(s).
	// FIX: Add context and return AgentResult, error
	Dispatch(ctx context.Context, event Event) (AgentResult, error)

	// RegisterAgent associates an agent name with its handler.
	RegisterAgent(name string, handler AgentHandler) error

	// GetCallbackRegistry returns the associated callback registry.
	GetCallbackRegistry() *CallbackRegistry

	// Stop allows the orchestrator to clean up resources.
	Stop()
}
