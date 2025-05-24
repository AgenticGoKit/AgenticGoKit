// Package core provides the public Agent interface and related types for AgentFlow.
package core

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

// AgentHandler defines the interface for executing agent logic.
type AgentHandler interface {
	Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

// AgentHandlerFunc allows using a function as an AgentHandler.
type AgentHandlerFunc func(ctx context.Context, event Event, state State) (AgentResult, error)

func (f AgentHandlerFunc) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
	return f(ctx, event, state)
}

// ... other agent related code ...
