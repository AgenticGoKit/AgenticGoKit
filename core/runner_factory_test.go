package core

import (
	"context"
	"testing"
	"time"
)

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
	runner := NewRunnerWithConfig(RunnerConfig{Agents: agents})
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
