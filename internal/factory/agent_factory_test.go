package factory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"kunalkushwaha/agentflow/internal/agents"
	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/orchestrator"
)

// Converts an agentflow.Agent to agentflow.AgentHandler
type agentAdapter struct {
	agent agentflow.Agent
}

func (a *agentAdapter) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	out, err := a.agent.Run(ctx, state)
	// Defensive: clear route in both output state and meta
	out.SetMeta(agentflow.RouteMetadataKey, "")
	outState := out
	outState.SetMeta(agentflow.RouteMetadataKey, "")
	fmt.Println("agentAdapter: Cleared route metadata after agent run")
	return agentflow.AgentResult{OutputState: outState}, err
}

func AgentToHandler(agent agentflow.Agent) agentflow.AgentHandler {
	return &agentAdapter{agent: agent}
}

// Allows using a function as an agentflow.Agent
type AgentFunc func(ctx context.Context, state agentflow.State) (agentflow.State, error)

func (f AgentFunc) Run(ctx context.Context, state agentflow.State) (agentflow.State, error) {
	return f(ctx, state)
}

func (f AgentFunc) Name() string { return "func-agent" }

// Allows using a function as an agentflow.AgentHandler
type AgentHandlerFunc func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error)

func (f AgentHandlerFunc) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	return f(ctx, event, state)
}

// --- Dummy agent for simple workflows ---
type DummyAgent struct {
	id      string
	called  *bool
	counter *int
}

func (a *DummyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	if a.called != nil {
		*a.called = true
	}
	if a.counter != nil {
		*a.counter++
	}
	out := state.Clone()
	out.Set("handled_by", a.id)
	return agentflow.AgentResult{OutputState: out}, nil
}

func (a *DummyAgent) Name() string {
	return a.id
}

// --- StateAgent: Agent with state manipulation ---
type StateAgent struct {
	id      string
	counter *int
}

func (a *StateAgent) Run(ctx context.Context, state agentflow.State) (agentflow.State, error) {
	if a.counter != nil {
		*a.counter++
	}
	fmt.Printf("StateAgent %s called, counter now %d\n", a.id, *a.counter)
	out := state.Clone()
	out.Set("handled_by", a.id)
	return out, nil
}

func (a *StateAgent) Name() string { return a.id }

// --- Route Orchestrator: Event routed to one agent ---
func TestRouteOrchestrator(t *testing.T) {
	called := false
	agents := map[string]agentflow.AgentHandler{
		"route-agent": &DummyAgent{id: "route-agent", called: &called},
	}
	runner := NewRunnerWithConfig(RunnerConfig{
		Agents: agents,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runner.Start(ctx)
	if err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}

	event := agentflow.NewEvent("test", agentflow.EventData{}, map[string]string{
		agentflow.RouteMetadataKey: "route-agent",
		agentflow.SessionIDKey:     "route-session",
	})
	_ = runner.Emit(event)
	time.Sleep(50 * time.Millisecond)
	if !called {
		t.Error("Route agent was not called")
	}
	runner.Stop()
}

// --- Collaborative Orchestrator: Event routed to all agents ---
func TestCollaborativeOrchestrator(t *testing.T) {
	called1, called2 := false, false
	agents := map[string]agentflow.AgentHandler{
		"a1": &DummyAgent{id: "a1", called: &called1},
		"a2": &DummyAgent{id: "a2", called: &called2},
	}
	runner := NewRunnerWithConfig(RunnerConfig{
		Agents:       agents,
		Orchestrator: orchestrator.NewCollaborativeOrchestrator(),
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runner.Start(ctx)
	if err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}

	event := agentflow.NewEvent("test", agentflow.EventData{}, map[string]string{
		agentflow.SessionIDKey: "collab-session",
	})
	_ = runner.Emit(event)
	time.Sleep(50 * time.Millisecond)
	if !called1 || !called2 {
		t.Error("Both agents should have been called in collaborative orchestrator")
	}
	runner.Stop()
}

// --- SequentialAgent: Chained agent execution ---
func TestSequentialAgentWorkflow(t *testing.T) {
	calls1, calls2 := 0, 0
	agent1 := &StateAgent{id: "seq1", counter: &calls1}
	agent2 := &StateAgent{id: "seq2", counter: &calls2}
	seqAgent := agents.NewSequentialAgent("seq", agent1, agent2)
	handler := AgentToHandler(seqAgent)
	agentsMap := map[string]agentflow.AgentHandler{"seq": handler}
	runner := NewRunnerWithConfig(RunnerConfig{Agents: agentsMap})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runner.Start(ctx)
	if err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}

	event := agentflow.NewEvent("test", agentflow.EventData{}, map[string]string{
		agentflow.RouteMetadataKey: "seq",
		agentflow.SessionIDKey:     "seq-session",
	})
	_ = runner.Emit(event)
	time.Sleep(50 * time.Millisecond)
	if calls1 != 1 || calls2 != 1 {
		t.Errorf("Expected each sequential agent to be called once, got %d and %d", calls1, calls2)
	}

	runner.Stop()
}

// --- ParallelAgent: Parallel agent execution ---
func TestParallelAgentWorkflow(t *testing.T) {
	calls1 := 0
	calls2 := 0
	agent1 := &StateAgent{id: "par1", counter: &calls1}
	agent2 := &StateAgent{id: "par2", counter: &calls2}
	config := agents.ParallelAgentConfig{Timeout: 2 * time.Second}
	parAgent := agents.NewParallelAgent("par", config, agent1, agent2)

	handler := AgentToHandler(parAgent)
	agentsMap := map[string]agentflow.AgentHandler{"par": handler}
	runner := NewRunnerWithConfig(RunnerConfig{Agents: agentsMap})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runner.Start(ctx)
	if err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}

	event := agentflow.NewEvent("test", agentflow.EventData{}, map[string]string{
		agentflow.RouteMetadataKey: "par",
		agentflow.SessionIDKey:     "par-session",
	})
	_ = runner.Emit(event)
	time.Sleep(100 * time.Millisecond)
	if calls1 != 1 || calls2 != 1 {
		t.Errorf("Both parallel agents should have been called once, got %d and %d", calls1, calls2)
	}
	runner.Stop()
}

// --- LoopAgent: Repeated agent execution ---
func TestLoopAgentWorkflow(t *testing.T) {
	calls := 0
	loopBodyWithState := AgentFunc(func(ctx context.Context, in agentflow.State) (agentflow.State, error) {
		calls++
		out := in.Clone()
		out.Set("calls", calls)
		return out, nil
	})
	config := agents.LoopAgentConfig{
		Condition: func(s agentflow.State) bool {
			callsVal, _ := s.Get("calls")
			return callsVal == 3
		},
		MaxIterations: 5,
	}
	loopAgent := agents.NewLoopAgent("loop", config, loopBodyWithState)
	handler := AgentToHandler(loopAgent)
	agentsMap := map[string]agentflow.AgentHandler{"loop": handler}
	runner := NewRunnerWithConfig(RunnerConfig{Agents: agentsMap})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runner.Start(ctx)
	if err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}

	event := agentflow.NewEvent("test", agentflow.EventData{}, map[string]string{
		agentflow.RouteMetadataKey: "loop",
		agentflow.SessionIDKey:     "loop-session",
	})
	_ = runner.Emit(event)
	time.Sleep(100 * time.Millisecond)
	if calls != 3 {
		t.Errorf("Loop agent should have run 3 times, got %d", calls)
	}
	runner.Stop()
}

// --- Error handling and callback hooks ---
func TestAgentErrorAndCallbacks(t *testing.T) {
	errorAgent := func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
		return agentflow.AgentResult{Error: "fail"}, context.DeadlineExceeded
	}
	agentsMap := map[string]agentflow.AgentHandler{
		"err": AgentHandlerFunc(errorAgent),
		"error-handler": AgentHandlerFunc(func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
			// Optionally, just return the state or log the error
			return agentflow.AgentResult{OutputState: state}, nil
		}),
	}
	runner := NewRunnerWithConfig(RunnerConfig{Agents: agentsMap})

	beforeCalled := false
	afterCalled := false
	runner.RegisterCallback(agentflow.HookBeforeAgentRun, "before", func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
		beforeCalled = true
		return args.State, nil
	})
	runner.RegisterCallback(agentflow.HookAfterAgentRun, "after", func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
		afterCalled = true
		return args.State, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runner.Start(ctx)
	if err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}

	event := agentflow.NewEvent("test", agentflow.EventData{}, map[string]string{
		agentflow.RouteMetadataKey: "err",
		agentflow.SessionIDKey:     "err-session",
	})
	_ = runner.Emit(event)
	time.Sleep(50 * time.Millisecond)
	if !beforeCalled || !afterCalled {
		t.Error("Expected both before and after callbacks to be called")
	}
	runner.Stop()
}

// --- Session state preservation ---
func TestSessionStatePreserved(t *testing.T) {
	agentsMap := map[string]agentflow.AgentHandler{
		"session": AgentHandlerFunc(func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
			count, _ := state.Get("count")
			n := 0
			if count != nil {
				n = count.(int)
			}
			out := state.Clone()
			out.Set("count", n+1)
			return agentflow.AgentResult{OutputState: out}, nil
		}),
	}
	runner := NewRunnerWithConfig(RunnerConfig{Agents: agentsMap})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runner.Start(ctx)
	if err != nil {
		t.Fatalf("Runner failed to start: %v", err)
	}

	meta := map[string]string{
		agentflow.RouteMetadataKey: "session",
		agentflow.SessionIDKey:     "session-test",
	}
	event := agentflow.NewEvent("test", agentflow.EventData{}, meta)
	_ = runner.Emit(event)
	time.Sleep(50 * time.Millisecond)
	_ = runner.Emit(event)
	time.Sleep(50 * time.Millisecond)

	// Check session state (if your runner exposes session state)
	// This is a placeholder; adapt to your session store API if needed.
	// session, _ := runner.GetSession("session-test")
	// state := session.GetState()
	// count, _ := state.Get("count")
	// if count != 2 {
	//     t.Errorf("Expected count to be 2, got %v", count)
	// }
	runner.Stop()
}
