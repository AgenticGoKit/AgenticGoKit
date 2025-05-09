package agentflow

import (
	"context"
)

// Agent defines the interface for any component that can process a State.
type Agent interface {
	// Run processes the input State and returns an output State or an error.
	// The context can be used for cancellation or deadlines.
	Run(ctx context.Context, inputState State) (State, error)
	// Name returns the unique identifier name of the agent.
	Name() string
}

// Note: The previous Agent interface (with Handle(Event)) might need to be
// renamed or refactored depending on how event handling and workflow execution
// will coexist or be integrated. For now, we define the new one as requested.

// AgentHandler defines the interface for executing agent logic.
type AgentHandler interface {
	Run(ctx context.Context, event Event, state State) (AgentResult, error)
	// It should NOT have a Handle method here.
}

// AgentHandlerFunc allows using a function as an AgentHandler.
type AgentHandlerFunc func(ctx context.Context, event Event, state State) (AgentResult, error)

func (f AgentHandlerFunc) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
	return f(ctx, event, state)
}

// ... other agent related code ...
