package factory

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestRouteOrchestrator is a simple orchestrator implementation for testing
type TestRouteOrchestrator struct {
	handlers map[string]AgentHandler
	registry *CallbackRegistry
}

func NewTestRouteOrchestrator(registry *CallbackRegistry) *TestRouteOrchestrator {
	return &TestRouteOrchestrator{
		handlers: make(map[string]AgentHandler),
		registry: registry,
	}
}

func (o *TestRouteOrchestrator) RegisterAgent(name string, handler AgentHandler) error {
	o.handlers[name] = handler
	return nil
}

func (o *TestRouteOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
	// Simple routing - get route from metadata
	metadata := event.GetMetadata()
	route, exists := metadata["route"]
	if !exists || route == "" {
		route = "default"
	}

	handler, exists := o.handlers[route]
	if !exists {
		return AgentResult{}, fmt.Errorf("no agent registered for route: %s", route)
	}

	state := NewState()
	return handler.Run(ctx, event, state)
}

func (o *TestRouteOrchestrator) GetCallbackRegistry() *CallbackRegistry {
	return o.registry
}

func (o *TestRouteOrchestrator) Stop() {
	// No cleanup needed
}

// DummyAgent implements AgentHandler for testing
type DummyAgent struct {
	id     string
	called *bool
}

func (a *DummyAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
	if a.called != nil {
		*a.called = true
	}
	out := state.Clone()
	out.Set("handled_by", a.id)
	return AgentResult{OutputState: out}, nil
}

func (a *DummyAgent) Name() string { return a.id }

func TestRunnerWithDummyAgent(t *testing.T) {
	called := false
	agents := map[string]AgentHandler{
		"dummy": &DummyAgent{id: "dummy", called: &called},
	}

	// Create a memory instance as required by the RunnerConfig
	memory := QuickMemory()

	// Create callback registry and orchestrator for the test
	callbackRegistry := NewCallbackRegistry()
	orchestrator := NewTestRouteOrchestrator(callbackRegistry)

	runner := NewRunnerWithConfig(RunnerConfig{
		Agents:       agents,
		Memory:       memory,
		SessionID:    "test-session",
		Orchestrator: orchestrator,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runner.Start(ctx); err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}
	event := NewEvent("test", EventData{}, map[string]string{"route": "dummy", "session_id": "test-session"})
	_ = runner.Emit(event)
	time.Sleep(50 * time.Millisecond)
	if !called {
		t.Error("Dummy agent was not called")
	}
	runner.Stop()
}
