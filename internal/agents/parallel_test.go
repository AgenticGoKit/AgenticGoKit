package agents

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// --- Test Helper Agent ---

// --- Test Cases ---

func TestParallelAgent_Run_AllSuccess(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("initial", "value")

	agent1 := &DelayAgent{Name: "agent1", Delay: 10 * time.Millisecond}
	agent2 := &DelayAgent{Name: "agent2", Delay: 20 * time.Millisecond}
	agent3 := &DelayAgent{Name: "agent3", Delay: 5 * time.Millisecond}

	parAgent := NewParallelAgent("test-all-success", ParallelAgentConfig{}, agent1, agent2, agent3)

	finalState, err := parAgent.Run(ctx, initialState)

	if err != nil {
		t.Fatalf("ParallelAgent.Run() returned an unexpected error: %v", err)
	}

	// Verify final state data - should contain initial data + data from all agents
	expectedData := map[string]interface{}{
		"initial": "value",
		"agent1":  "processed_by_agent1",
		"agent2":  "processed_by_agent2",
		"agent3":  "processed_by_agent3",
	}
	if !reflect.DeepEqual(finalState.GetData(), expectedData) {
		t.Errorf("Final state data mismatch:\ngot:  %#v\nwant: %#v", finalState.GetData(), expectedData)
	}
}

func TestParallelAgent_Run_PartialFailure(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("initial", "value")

	simulatedError := errors.New("agent2 failed deliberately")
	agent1 := &DelayAgent{Name: "agent1", Delay: 10 * time.Millisecond}
	agent2 := &DelayAgent{Name: "agent2", Delay: 5 * time.Millisecond, ReturnError: simulatedError} // Fails quickly
	agent3 := &DelayAgent{Name: "agent3", Delay: 15 * time.Millisecond}                             // Succeeds later

	parAgent := NewParallelAgent("test-partial-fail", ParallelAgentConfig{}, agent1, agent2, agent3)

	finalState, err := parAgent.Run(ctx, initialState)

	if err == nil {
		t.Fatalf("ParallelAgent.Run() did not return an error when expected.")
	}

	// Check if it's a MultiError
	multiErr, ok := err.(*agentflow.MultiError)
	if !ok {
		t.Fatalf("Expected a *MultiError, but got type %T: %v", err, err)
	}

	// Check if the specific error is contained within the MultiError
	found := false
	for _, containedErr := range multiErr.Errors {
		if errors.Is(containedErr, simulatedError) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error '%v' to be contained in MultiError, but it wasn't: %v", simulatedError, multiErr)
	}
	if len(multiErr.Errors) != 1 {
		t.Errorf("Expected 1 error in MultiError, got %d: %v", len(multiErr.Errors), multiErr)
	}
	t.Logf("Received expected MultiError: %v", multiErr)

	// Verify final state data - should contain initial data + data from successful agents (1 and 3)
	expectedData := map[string]interface{}{
		"initial": "value",
		"agent1":  "processed_by_agent1",
		// agent2 data should be missing
		"agent3": "processed_by_agent3",
	}
	if !reflect.DeepEqual(finalState.GetData(), expectedData) {
		t.Errorf("Final state data mismatch on partial failure:\ngot:  %#v\nwant: %#v", finalState.GetData(), expectedData)
	}
}

func TestParallelAgent_Run_Timeout(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("initial", "value")

	agent1 := &DelayAgent{Name: "agent1", Delay: 10 * time.Millisecond}  // Finishes before timeout
	agent2 := &DelayAgent{Name: "agent2", Delay: 100 * time.Millisecond} // Will be cancelled by timeout
	agent3 := &DelayAgent{Name: "agent3", Delay: 5 * time.Millisecond}   // Finishes before timeout

	// Configure timeout shorter than agent2's delay
	config := ParallelAgentConfig{Timeout: 50 * time.Millisecond}
	parAgent := NewParallelAgent("test-timeout", config, agent1, agent2, agent3)

	startTime := time.Now()
	finalState, err := parAgent.Run(ctx, initialState)
	duration := time.Since(startTime)

	if err == nil {
		t.Fatalf("ParallelAgent.Run() did not return an error when timeout expected.")
	}

	// Check for MultiError containing context.DeadlineExceeded
	multiErr, ok := err.(*agentflow.MultiError)
	if !ok {
		t.Fatalf("Expected a *MultiError on timeout, but got type %T: %v", err, err)
	}

	foundDeadline := false
	for _, containedErr := range multiErr.Errors {
		// The error from the cancelled agent should wrap context.DeadlineExceeded
		if errors.Is(containedErr, context.DeadlineExceeded) {
			foundDeadline = true
			t.Logf("Found expected deadline error within MultiError: %v", containedErr)
			break
		}
	}
	if !foundDeadline {
		t.Errorf("Expected context.DeadlineExceeded to be wrapped in MultiError, but it wasn't: %v", multiErr)
	}
	// We expect only agent2 to time out
	if len(multiErr.Errors) != 1 {
		t.Errorf("Expected 1 error (timeout) in MultiError, got %d: %v", len(multiErr.Errors), multiErr)
	}

	// Check if execution time is roughly the timeout duration
	if duration < config.Timeout || duration > config.Timeout+(20*time.Millisecond) { // Allow some buffer
		t.Errorf("Execution time (%v) was significantly different from timeout (%v)", duration, config.Timeout)
	}

	// Verify final state data - should contain initial data + data from successful agents (1 and 3)
	expectedData := map[string]interface{}{
		"initial": "value",
		"agent1":  "processed_by_agent1",
		// agent2 data should be missing
		"agent3": "processed_by_agent3",
	}
	if !reflect.DeepEqual(finalState.GetData(), expectedData) {
		t.Errorf("Final state data mismatch on timeout:\ngot:  %#v\nwant: %#v", finalState.GetData(), expectedData)
	}

	// Verify agent2's Run was likely entered but cancelled
	if agent2.RunCount.Load() == 0 {
		t.Errorf("Agent2 (timed out) Run method was expected to be called at least once, but count is 0")
	}
}

func TestParallelAgent_Run_ExternalContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	initialState := agentflow.NewState()
	initialState.Set("initial", "value")

	agent1 := &DelayAgent{Name: "agent1", Delay: 100 * time.Millisecond} // Will be cancelled
	agent2 := &DelayAgent{Name: "agent2", Delay: 150 * time.Millisecond} // Will be cancelled
	agent3 := &DelayAgent{Name: "agent3", Delay: 5 * time.Millisecond}   // Finishes before cancel

	parAgent := NewParallelAgent("test-cancel", ParallelAgentConfig{}, agent1, agent2, agent3)

	// Cancel the context shortly after starting
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	startTime := time.Now()
	finalState, err := parAgent.Run(ctx, initialState)
	duration := time.Since(startTime)

	if err == nil {
		t.Fatalf("ParallelAgent.Run() did not return an error when cancellation expected.")
	}

	// Check for MultiError containing context.Canceled
	multiErr, ok := err.(*agentflow.MultiError)
	if !ok {
		t.Fatalf("Expected a *MultiError on cancellation, but got type %T: %v", err, err)
	}

	foundCanceled := 0
	for _, containedErr := range multiErr.Errors {
		if errors.Is(containedErr, context.Canceled) {
			foundCanceled++
			t.Logf("Found expected canceled error within MultiError: %v", containedErr)
		}
	}
	if foundCanceled == 0 {
		t.Errorf("Expected context.Canceled to be wrapped in MultiError, but it wasn't: %v", multiErr)
	}
	// Expect errors from agent1 and agent2 due to cancellation
	if len(multiErr.Errors) != 2 {
		t.Errorf("Expected 2 errors (cancellations) in MultiError, got %d: %v", len(multiErr.Errors), multiErr)
	}

	// Check if execution time is roughly the cancellation time
	if duration < 25*time.Millisecond || duration > 60*time.Millisecond { // Allow buffer
		t.Errorf("Execution time (%v) was significantly different from expected cancellation time (~30ms)", duration)
	}

	// Verify final state data - should contain initial data + data from successful agent (3)
	expectedData := map[string]interface{}{
		"initial": "value",
		// agent1 data missing
		// agent2 data missing
		"agent3": "processed_by_agent3",
	}
	if !reflect.DeepEqual(finalState.GetData(), expectedData) {
		t.Errorf("Final state data mismatch on cancellation:\ngot:  %#v\nwant: %#v", finalState.GetData(), expectedData)
	}

	// Verify agents 1 and 2 were likely entered but cancelled
	if agent1.RunCount.Load() == 0 {
		t.Errorf("Agent1 (cancelled) Run method was expected to be called at least once, but count is 0")
	}
	if agent2.RunCount.Load() == 0 {
		t.Errorf("Agent2 (cancelled) Run method was expected to be called at least once, but count is 0")
	}
}

func TestParallelAgent_Run_ZeroAgents(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("initial", "value")

	parAgent := NewParallelAgent("test-zero", ParallelAgentConfig{}) // No agents
	finalState, err := parAgent.Run(ctx, initialState)

	if err != nil {
		t.Fatalf("Run with zero agents returned an error: %v", err)
	}
	// Should return the initial state unmodified
	if !reflect.DeepEqual(finalState.GetData(), initialState.GetData()) {
		t.Errorf("Run with zero agents modified the state:\ngot:  %#v\nwant: %#v", finalState.GetData(), initialState.GetData())
	}
	// Ensure it's not just the same pointer
	if &finalState == &initialState {
		t.Errorf("Run with zero agents returned the exact same state instance, expected a clone or original.")
	}
}

func TestParallelAgent_Run_NilAgentsFiltered(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("initial", "value")

	agent1 := &DelayAgent{Name: "agent1", Delay: 10 * time.Millisecond}
	agent3 := &DelayAgent{Name: "agent3", Delay: 5 * time.Millisecond}

	parAgent := NewParallelAgent("test-nil-filtered", ParallelAgentConfig{}, agent1, nil, agent3, nil)

	if len(parAgent.agents) != 2 {
		t.Fatalf("NewParallelAgent did not filter nil agents correctly, got %d, want 2", len(parAgent.agents))
	}

	finalState, err := parAgent.Run(ctx, initialState)

	if err != nil {
		t.Fatalf("Run after filtering nil agents returned an error: %v", err)
	}

	// Verify final state data - should contain initial data + data from non-nil agents (1 and 3)
	expectedData := map[string]interface{}{
		"initial": "value",
		"agent1":  "processed_by_agent1",
		"agent3":  "processed_by_agent3",
	}
	if !reflect.DeepEqual(finalState.GetData(), expectedData) {
		t.Errorf("Final state data mismatch after filtering nil:\ngot:  %#v\nwant: %#v", finalState.GetData(), expectedData)
	}
}

// --- Benchmark ---

func BenchmarkParallelAgent_Run(b *testing.B) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	numAgents := 50 // As per requirement
	agents := make([]agentflow.Agent, numAgents)
	for i := 0; i < numAgents; i++ {
		// Use NoOpAgent for minimal overhead benchmark
		agents[i] = &NoOpAgent{}
	}
	// No timeout for benchmark
	parAgent := NewParallelAgent("benchmark-parallel", ParallelAgentConfig{}, agents...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Pass a clone to avoid interference between benchmark iterations
		_, err := parAgent.Run(ctx, initialState.Clone())
		if err != nil {
			// Benchmarks shouldn't error with NoOpAgents
			b.Fatalf("Benchmark run failed unexpectedly: %v", err)
		}
	}
}
