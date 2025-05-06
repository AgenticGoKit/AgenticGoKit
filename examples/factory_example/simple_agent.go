package main

import (
	"context"
	"fmt"
	"log"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// SimpleAgentHandler is a basic agent handler for demonstration purposes.
type SimpleAgentHandler struct{}

// NewSimpleAgentBuilder creates a builder for SimpleAgentHandler.
func NewSimpleAgentBuilder() *SimpleAgentBuilder {
	return &SimpleAgentBuilder{}
}

// SimpleAgentBuilder is a builder for SimpleAgentHandler.
type SimpleAgentBuilder struct{}

// Build creates a new SimpleAgentHandler.
func (b *SimpleAgentBuilder) Build() agentflow.AgentHandler {
	return &SimpleAgentHandler{}
}

// Run processes an event and returns a response.
func (h *SimpleAgentHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	// Extract the message from the event data.
	messageObj, ok := event.GetData()["message"]
	if !ok {
		return agentflow.AgentResult{}, fmt.Errorf("missing 'message' in event data")
	}

	message, ok := messageObj.(string)
	if !ok {
		return agentflow.AgentResult{}, fmt.Errorf("'message' is not a string")
	}

	// Log the received message.
	log.Printf("Simple Agent Received message: %s", message)

	// Create a response message.
	response := fmt.Sprintf("Agent received: %s", message)

	// Create a new state and set the response.
	outputState := state.Clone()
	outputState.Set("response", response)

	// Return the result.
	return agentflow.AgentResult{
		OutputState: outputState,
	}, nil
}

// LoopConditionAgent checks a counter in the state to determine if a loop should continue.
type LoopConditionAgent struct{}

// NewLoopConditionAgentBuilder creates a builder for LoopConditionAgent.
func NewLoopConditionAgentBuilder() *LoopConditionAgentBuilder {
	return &LoopConditionAgentBuilder{}
}

// LoopConditionAgentBuilder is a builder for LoopConditionAgent.
type LoopConditionAgentBuilder struct{}

// Build creates a new LoopConditionAgent.
func (b *LoopConditionAgentBuilder) Build() agentflow.AgentHandler {
	return &LoopConditionAgent{}
}

// Run checks the "counter" in the state and returns a result indicating whether to continue the loop.
func (h *LoopConditionAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	counterObj, ok := state.Get("counter")
	counter := 0
	if ok {
		counterFloat, ok := counterObj.(float64)
		if ok {
			counter = int(counterFloat)
		}
	}

	outputState := state.Clone()
	if counter < 3 {
		log.Printf("LoopConditionAgent: Counter is %d, continuing loop.", counter)
		return agentflow.AgentResult{OutputState: outputState}, nil // Return empty output state to continue
	} else {
		log.Printf("LoopConditionAgent: Counter is %d, ending loop.", counter)
		outputState.Set("loop_ended", true)
		return agentflow.AgentResult{OutputState: outputState}, nil // Set loop_ended to true to stop
	}
}

// IncrementCounterAgent increments a counter in the state.
type IncrementCounterAgent struct{}

// NewIncrementCounterAgentBuilder creates a builder for IncrementCounterAgent.
func NewIncrementCounterAgentBuilder() *IncrementCounterAgentBuilder {
	return &IncrementCounterAgentBuilder{}
}

// IncrementCounterAgentBuilder is a builder for IncrementCounterAgent.
type IncrementCounterAgentBuilder struct{}

// Build creates a new IncrementCounterAgent.
func (b *IncrementCounterAgentBuilder) Build() agentflow.AgentHandler {
	return &IncrementCounterAgent{}
}

// Run increments the "counter" in the state.
func (h *IncrementCounterAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	counter := state.Get("counter").(int)
	state.Set("counter", counter+1)
	return agentflow.AgentResult{OutputState: state}, nil
}