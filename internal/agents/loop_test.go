package agents

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/internal/core"
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

	// Create LoopAgent without a Condition, relying on MaxIterations
	// FIX: Use LoopAgentConfig type
	config := LoopAgentConfig{
		// Condition: nil, // Implicitly nil
		// MaxIterations: 0, // Will use default
	}
	loopAgent := NewLoopAgent("test-default-max", config, subAgent) // Pass config correctly
	if loopAgent == nil {
		t.Fatal("NewLoopAgent returned nil")
	}

	finalState, err := loopAgent.Run(ctx, initialState)

	// Expect ErrMaxIterationsReached
	if !errors.Is(err, agentflow.ErrMaxIterationsReached) {
		t.Fatalf("Expected ErrMaxIterationsReached, got: %v", err)
	}
	if finalState == nil {
		t.Fatal("LoopAgent returned nil state on reaching max iterations")
	}

	// Verify final state count
	finalCountVal, _ := finalState.Get("count")
	finalCount, ok := finalCountVal.(int)
	// FIX: Provide both arguments to Errorf
	if !ok || finalCount != defaultMaxIterations {
		t.Errorf("Expected final count to be %d (default MaxIterations), got %v", defaultMaxIterations, finalCountVal)
	}
}

func TestLoopAgent_Run_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	initialState := agentflow.NewState()
	initialState.Set("run_count", 0)

	// Use DelayAgent with a delay longer than the cancellation timer
	subAgent := NewDelayAgent("delaySubAgent", 20*time.Millisecond, nil)

	stopCondition := func(s agentflow.State) bool {
		runCountVal, _ := s.Get("run_count")
		runCount, _ := runCountVal.(int)
		return runCount >= 5
	}

	wrapperAgent := NewSequentialAgent("wrapper", &SimpleUpdateAgent{Key: "run_count"}, subAgent)

	config := LoopAgentConfig{
		Condition:     stopCondition,
		MaxIterations: 10,
	}
	loopAgent := NewLoopAgent("test-cancel-loop", config, wrapperAgent)
	if loopAgent == nil {
		t.Fatal("NewLoopAgent returned nil")
	}

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

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error to be wrapped, got: %v", err)
	}
	t.Logf("Received expected cancellation error: %v", err)

	if duration < cancelDelay || duration > cancelDelay+(30*time.Millisecond) {
		t.Errorf("Execution time (%v) was significantly different from expected cancellation time (~%v)", duration, cancelDelay)
	}

	finalRunCountVal, _ := finalState.Get("run_count")
	finalRunCount, ok := finalRunCountVal.(int)
	if !ok {
		// If cancellation happened very early, run_count might not exist yet. Treat as 0.
		if finalRunCountVal == nil {
			finalRunCount = 0
			ok = true
		} else {
			t.Fatalf("Final run_count is not an int: %v", finalRunCountVal)
		}
	}

	if !(finalRunCount == 0 || finalRunCount == 1) {
		t.Errorf("Expected final run_count to be 0 or 1 due to cancellation timing, got %d", finalRunCount)
	}

	t.Logf("Final run_count on cancellation: %d (0 or 1 expected)", finalRunCount)
}

func TestNewLoopAgent_NilSubAgent(t *testing.T) {
	config := LoopAgentConfig{MaxIterations: 5}
	loopAgent := NewLoopAgent("test-nil-sub", config, nil)
	if loopAgent != nil {
		t.Fatal("NewLoopAgent should return nil when subAgent is nil")
	}
}

func TestLoopAgent_Run_Timeout(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("count", 0)

	slowAgent := NewDelayAgent("slow", 100*time.Millisecond, nil)

	config := LoopAgentConfig{
		Timeout: 50 * time.Millisecond,
		Condition: func(s agentflow.State) bool {
			countVal, _ := s.Get("count")
			count, _ := countVal.(int)
			return count < 5
		},
	}
	loopAgent := NewLoopAgent(
		"test-loop-timeout",
		config,
		slowAgent,
	)
	if loopAgent == nil {
		t.Fatal("NewLoopAgent returned nil unexpectedly")
	}

	startTime := time.Now()
	finalState, err := loopAgent.Run(ctx, initialState)
	duration := time.Since(startTime)

	if err == nil {
		t.Fatalf("LoopAgent.Run() did not return an error when timeout expected.")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
	}

	if duration < loopAgent.config.Timeout || duration > loopAgent.config.Timeout+(30*time.Millisecond) {
		t.Errorf("Execution time (%v) was significantly different from timeout (%v)", duration, loopAgent.config.Timeout)
	}

	if finalState == nil {
		t.Fatal("LoopAgent.Run() returned nil state on timeout")
	}
	countVal, _ := finalState.Get("count")
	count, ok := countVal.(int)
	if !ok && countVal == nil {
		count = 0
		ok = true
	}
	if !ok || count != 0 {
		t.Errorf("Expected final count 0 on timeout, got %v (type %T)", countVal, countVal)
	}

	if _, exists := finalState.Get("slow"); exists {
		t.Error("Final state should not contain data from the timed-out agent")
	}
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
