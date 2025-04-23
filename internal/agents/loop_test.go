package agents

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// --- Test Helper Agents ---

// --- Test Cases ---

func TestLoopAgent_Run_ConditionMet(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("count", 0)

	subAgent := &CounterAgent{}
	stopCondition := func(s agentflow.State) bool {
		countVal, _ := s.Get("count")
		count, _ := countVal.(int)
		return count >= 3 // Stop when count reaches 3 or more
	}

	config := LoopAgentConfig{
		Condition:     stopCondition,
		MaxIterations: 5, // Should stop before reaching this
	}
	loopAgent := NewLoopAgent("test-cond-met", config, subAgent)
	if loopAgent == nil {
		t.Fatal("NewLoopAgent returned nil")
	}

	finalState, err := loopAgent.Run(ctx, initialState)

	if err != nil {
		t.Fatalf("LoopAgent.Run() returned an unexpected error: %v", err)
	}

	// Verify final state count
	finalCountVal, _ := finalState.Get("count")
	finalCount, ok := finalCountVal.(int)
	if !ok || finalCount != 3 {
		t.Errorf("Expected final count to be 3, got %v (type %T)", finalCountVal, finalCountVal)
	}
}

func TestLoopAgent_Run_MaxIterationsReached(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("count", 0)

	subAgent := &CounterAgent{}
	// Condition that is never met
	stopCondition := func(s agentflow.State) bool {
		return false
	}

	config := LoopAgentConfig{
		Condition:     stopCondition,
		MaxIterations: 4, // Set a specific max
	}
	loopAgent := NewLoopAgent("test-max-iter", config, subAgent)
	if loopAgent == nil {
		t.Fatal("NewLoopAgent returned nil")
	}

	finalState, err := loopAgent.Run(ctx, initialState)

	if err == nil {
		t.Fatalf("LoopAgent.Run() did not return an error when max iterations expected.")
	}

	// Check if the specific error is returned
	if !errors.Is(err, agentflow.ErrMaxIterationsReached) {
		t.Errorf("Expected error '%v', but got: %v", agentflow.ErrMaxIterationsReached, err)
	}
	t.Logf("Received expected error: %v", err)

	// Verify final state count - should be equal to MaxIterations
	finalCountVal, _ := finalState.Get("count")
	finalCount, ok := finalCountVal.(int)
	if !ok || finalCount != config.MaxIterations {
		t.Errorf("Expected final count to be %d (MaxIterations), got %v", config.MaxIterations, finalCountVal)
	}
}

func TestLoopAgent_Run_SubAgentErrors(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("count", 0)

	simulatedError := errors.New("sub-agent failed")
	subAgent := &CounterAgent{
		FailOnCount: 3, // Fail on the 3rd iteration
		ReturnError: simulatedError,
	}
	// Condition that would eventually be met, but error happens first
	stopCondition := func(s agentflow.State) bool {
		countVal, _ := s.Get("count")
		count, _ := countVal.(int)
		return count >= 5
	}

	config := LoopAgentConfig{
		Condition:     stopCondition,
		MaxIterations: 10,
	}
	loopAgent := NewLoopAgent("test-sub-error", config, subAgent)
	if loopAgent == nil {
		t.Fatal("NewLoopAgent returned nil")
	}

	finalState, err := loopAgent.Run(ctx, initialState)

	if err == nil {
		t.Fatalf("LoopAgent.Run() did not return an error when sub-agent failure expected.")
	}

	// Check if the specific error is wrapped
	if !errors.Is(err, simulatedError) {
		t.Errorf("Expected error '%v' to be wrapped, but got: %v", simulatedError, err)
	}
	if !strings.Contains(err.Error(), "iteration 3") {
		t.Errorf("Expected error message to mention iteration 3, got: %v", err)
	}
	t.Logf("Received expected wrapped error: %v", err)

	// Verify final state count - should be the count *before* the error occurred (2)
	finalCountVal, _ := finalState.Get("count")
	finalCount, ok := finalCountVal.(int)
	if !ok || finalCount != 2 {
		t.Errorf("Expected final count to be 2 (before error), got %v", finalCountVal)
	}
}

func TestLoopAgent_Run_DefaultMaxIterations(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("count", 0)

	subAgent := &CounterAgent{}
	// No condition provided, should hit default max iterations
	config := LoopAgentConfig{
		Condition:     nil,
		MaxIterations: 0, // Trigger default
	}
	loopAgent := NewLoopAgent("test-default-max", config, subAgent)
	if loopAgent == nil {
		t.Fatal("NewLoopAgent returned nil")
	}
	if loopAgent.config.MaxIterations != defaultMaxIterations {
		t.Fatalf("LoopAgent did not set default MaxIterations. Got %d, want %d", loopAgent.config.MaxIterations, defaultMaxIterations)
	}

	finalState, err := loopAgent.Run(ctx, initialState)

	if !errors.Is(err, agentflow.ErrMaxIterationsReached) {
		t.Fatalf("Expected ErrMaxIterationsReached, got: %v", err)
	}

	// Verify final state count
	finalCountVal, _ := finalState.Get("count")
	finalCount, ok := finalCountVal.(int)
	if !ok || finalCount != defaultMaxIterations {
		t.Errorf("Expected final count to be %d (default MaxIterations), got %v", defaultMaxIterations, finalCountVal)
	}
}

func TestLoopAgent_Run_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	initialState := agentflow.NewState()
	initialState.Set("run_count", 0) // Use a different key to track runs

	// Use DelayAgent with a delay longer than the cancellation timer
	subAgent := &DelayAgent{
		Name:  "delaySubAgent",
		Delay: 20 * time.Millisecond, // Delay > cancellation timer
		DataToAdd: map[string]interface{}{
			"status": "processing", // Some data the agent might add
		},
	}

	// Condition that would eventually be met if not cancelled
	stopCondition := func(s agentflow.State) bool {
		runCountVal, _ := s.Get("run_count")
		runCount, _ := runCountVal.(int)
		return runCount >= 5
	}

	// Need a way to increment run_count *outside* the DelayAgent's delay
	// Let's wrap the DelayAgent in a simple sequential agent for the test
	wrapperAgent := NewSequentialAgent("wrapper", &SimpleUpdateAgent{Key: "run_count"}, subAgent)

	config := LoopAgentConfig{
		Condition:     stopCondition,
		MaxIterations: 10,
	}
	// Use the wrapper agent in the loop
	loopAgent := NewLoopAgent("test-cancel-loop", config, wrapperAgent)
	if loopAgent == nil {
		t.Fatal("NewLoopAgent returned nil")
	}

	// Cancel context after a short delay, during the first agent's delay
	cancelDelay := 10 * time.Millisecond
	go func() {
		time.Sleep(cancelDelay)
		cancel()
	}()

	startTime := time.Now()
	finalState, err := loopAgent.Run(ctx, initialState)
	duration := time.Since(startTime)

	if err == nil {
		t.Fatalf("LoopAgent.Run() did not return an error when context cancellation expected.")
	}

	// Check if context.Canceled is wrapped
	// The error might come from the LoopAgent check or the DelayAgent check
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error to be wrapped, got: %v", err)
	}
	t.Logf("Received expected cancellation error: %v", err)

	// Check timing - should be roughly cancelDelay + loop overhead
	// Allow a generous buffer for scheduling variance
	if duration < cancelDelay || duration > cancelDelay+(30*time.Millisecond) {
		t.Errorf("Execution time (%v) was significantly different from expected cancellation time (~%v)", duration, cancelDelay)
	}

	// Verify final state - should be the state from before or after the iteration where cancellation was detected.
	finalRunCountVal, _ := finalState.Get("run_count")
	finalRunCount, ok := finalRunCountVal.(int)
	if !ok {
		t.Fatalf("Final run_count is not an int: %v", finalRunCountVal)
	}

	// Accept either 0 (cancelled during iter 1) or 1 (cancelled before iter 2)
	if !(finalRunCount == 0 || finalRunCount == 1) {
		t.Errorf("Expected final run_count to be 0 or 1 due to cancellation timing, got %d", finalRunCount)
	}

	// Data from DelayAgent should still not be present as cancellation happened before it could add data.
	if _, exists := finalState.GetData()["status"]; exists {
		t.Errorf("Expected final state not to contain 'status' data from cancelled agent run")
	}
	t.Logf("Final run_count on cancellation: %d (0 or 1 expected)", finalRunCount)
}

func TestNewLoopAgent_NilSubAgent(t *testing.T) {
	config := LoopAgentConfig{MaxIterations: 5}
	loopAgent := NewLoopAgent("test-nil-sub", config, nil) // Pass nil sub-agent
	if loopAgent != nil {
		t.Fatal("NewLoopAgent should return nil when subAgent is nil")
	}
	// No error expected here, just nil return (logged internally)
}

// --- Benchmark ---

func BenchmarkLoopAgent_Run(b *testing.B) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("count", 0)

	// Sub-agent that does minimal work (just increments count)
	subAgent := &CounterAgent{}

	// Condition to stop after exactly 10 iterations
	targetIterations := 10
	stopCondition := func(s agentflow.State) bool {
		countVal, _ := s.Get("count")
		count, _ := countVal.(int)
		return count >= targetIterations
	}

	config := LoopAgentConfig{
		Condition:     stopCondition,
		MaxIterations: targetIterations + 5, // Set higher max to ensure condition stops it
	}
	loopAgent := NewLoopAgent("benchmark-loop", config, subAgent)
	if loopAgent == nil {
		b.Fatal("NewLoopAgent returned nil")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Pass a clone to avoid interference between benchmark iterations
		// Reset initial state data for each benchmark run if necessary, though counter handles it here
		_, err := loopAgent.Run(ctx, initialState.Clone())
		if err != nil {
			// Benchmark shouldn't error in this configuration
			b.Fatalf("Benchmark run failed unexpectedly: %v", err)
		}
	}
}
