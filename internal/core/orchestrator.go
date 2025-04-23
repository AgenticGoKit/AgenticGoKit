package agentflow

// Orchestrator defines the strategy for dispatching an event to handlers.
type Orchestrator interface {
	// RegisterAgent adds an event handler to the orchestrator. (Method name kept for consistency, but type changed)
	RegisterAgent(name string, handler EventHandler) // Use EventHandler
	// Dispatch sends the event according to the orchestration strategy.
	Dispatch(event Event)
	// Stop allows the orchestrator to clean up resources if needed.
	Stop()
}
