package agents

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/internal/core"
)

// Helper: filter out execution metadata keys added by agents
func filterExecutionDataKeys(keys []string) []string {
	skip := map[string]bool{
		"executed_by":                    true,
		"execution_timestamp":            true,
		"execution_role":                 true,
		"parallel_execution_completed":   true,
		"sequential_execution_completed": true,
	}
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if !skip[k] {
			out = append(out, k)
		}
	}
	return out
}

// --- Test Helper Agents ---
// Definitions for DelayAgent and NoOpAgent are REMOVED from here.
// They reside in agents_test_helpers.go

// --- Test Cases ---

func TestParallelAgent_Run_AllSuccess(t *testing.T) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	initialState.Set("initial", "value")
	initialState.SetMeta("initial_meta", "meta_value")

	// These types come from agents_test_helpers.go
	agent1 := NewDelayAgent("agent1", 10*time.Millisecond, nil)
	agent2 := NewDelayAgent("agent2", 20*time.Millisecond, nil)
	agent3 := NewDelayAgent("agent3", 5*time.Millisecond, nil)

	parAgent := NewParallelAgent("test-all-success", ParallelAgentConfig{}, agent1, agent2, agent3)

	finalState, err := parAgent.Run(ctx, initialState)

	// FIX: Check if err is nil. NewMultiError returns nil if no errors.
	if err != nil {
		t.Fatalf("ParallelAgent.Run() returned an unexpected error: %v", err)
	}
	if finalState == nil {
		t.Fatal("ParallelAgent.Run() returned nil state")
	}

	expectedData := map[string]interface{}{
		"initial": "value",
		"agent1":  "processed_by_agent1",
		"agent2":  "processed_by_agent2",
		"agent3":  "processed_by_agent3",
	}
	finalKeys := filterExecutionDataKeys(finalState.Keys())
	if len(finalKeys) != len(expectedData) {
		t.Errorf("Final state data key count mismatch: got %d (%v), want %d (%v)", len(finalKeys), finalKeys, len(expectedData), keysFromMap(expectedData))
	}
	for k, expectedV := range expectedData {
		actualV, ok := finalState.Get(k)
		if !ok {
			t.Errorf("Final state missing expected data key: %s", k)
		} else if !reflect.DeepEqual(actualV, expectedV) {
			t.Errorf("Final state data mismatch for key '%s': got %v (%T), want %v (%T)", k, actualV, actualV, expectedV, expectedV)
		}
	}
	expectedMeta := map[string]string{
		"initial_meta": "meta_value",
		"agent1_meta":  "meta_from_agent1",
		"agent2_meta":  "meta_from_agent2",
		"agent3_meta":  "meta_from_agent3",
	}
	finalMetaKeys := finalState.MetaKeys()
	if len(finalMetaKeys) != len(expectedMeta) {
		t.Errorf("Final state metadata key count mismatch: got %d (%v), want %d (%v)", len(finalMetaKeys), finalMetaKeys, len(expectedMeta), keysFromMapStr(expectedMeta))
	}
	for k, expectedV := range expectedMeta {
		actualV, ok := finalState.GetMeta(k)
		if !ok {
			t.Errorf("Final state missing expected metadata key: %s", k)
		} else if actualV != expectedV {
			t.Errorf("Final state metadata mismatch for key '%s': got %q, want %q", k, actualV, expectedV)
		}
	}
}

func TestParallelAgent_Run_PartialFailure(t *testing.T) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	initialState.Set("initial", "value")
	initialState.SetMeta("initial_meta", "meta_value")

	simulatedError := errors.New("agent2 failed deliberately")
	agent1 := NewDelayAgent("agent1", 10*time.Millisecond, nil)
	agent2 := NewDelayAgent("agent2", 5*time.Millisecond, simulatedError)
	agent3 := NewDelayAgent("agent3", 15*time.Millisecond, nil)

	parAgent := NewParallelAgent("test-partial-fail", ParallelAgentConfig{}, agent1, agent2, agent3)

	finalState, err := parAgent.Run(ctx, initialState)

	if err == nil {
		t.Fatalf("ParallelAgent.Run() did not return an error when expected.")
	}

	multiErr, ok := err.(*agenticgokit.MultiError)
	if !ok {
		t.Fatalf("Expected a *MultiError, but got type %T: %v", err, err)
	}

	found := false
	// FIX: Access Errors field directly
	for _, containedErr := range multiErr.Errors {
		if errors.Is(containedErr, simulatedError) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error '%v' to be contained/wrapped in MultiError, but it wasn't: %v", simulatedError, multiErr)
	}
	// FIX: Access Errors field directly for length check
	if len(multiErr.Errors) != 1 {
		t.Errorf("Expected 1 error in MultiError, got %d: %v", len(multiErr.Errors), multiErr)
	}
	t.Logf("Received expected MultiError: %v", multiErr)

	if finalState == nil {
		t.Fatal("ParallelAgent.Run() returned nil state even on partial failure")
	}

	expectedData := map[string]interface{}{
		"initial": "value",
		"agent1":  "processed_by_agent1",
		"agent3":  "processed_by_agent3",
	}
	finalKeys := filterExecutionDataKeys(finalState.Keys())
	if len(finalKeys) != len(expectedData) {
		t.Errorf("Final state data key count mismatch on partial failure: got %d (%v), want %d (%v)", len(finalKeys), finalKeys, len(expectedData), keysFromMap(expectedData))
	}
	for k, expectedV := range expectedData {
		actualV, ok := finalState.Get(k)
		if !ok {
			t.Errorf("Final state missing expected data key on partial failure: %s", k)
		} else if !reflect.DeepEqual(actualV, expectedV) {
			t.Errorf("Final state data mismatch for key '%s' on partial failure: got %v (%T), want %v (%T)", k, actualV, actualV, expectedV, expectedV)
		}
	}
	expectedMeta := map[string]string{
		"initial_meta": "meta_value",
		"agent1_meta":  "meta_from_agent1",
		"agent3_meta":  "meta_from_agent3",
	}
	finalMetaKeys := finalState.MetaKeys()
	if len(finalMetaKeys) != len(expectedMeta) {
		t.Errorf("Final state metadata key count mismatch on partial failure: got %d (%v), want %d (%v)", len(finalMetaKeys), finalMetaKeys, len(expectedMeta), keysFromMapStr(expectedMeta))
	}
	for k, expectedV := range expectedMeta {
		actualV, ok := finalState.GetMeta(k)
		if !ok {
			t.Errorf("Final state missing expected metadata key on partial failure: %s", k)
		} else if actualV != expectedV {
			t.Errorf("Final state metadata mismatch for key '%s' on partial failure: got %q, want %q", k, actualV, expectedV)
		}
	}
}

func TestParallelAgent_Run_Timeout(t *testing.T) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	initialState.Set("initial", "value")
	initialState.SetMeta("initial_meta", "meta_value")

	agent1 := NewDelayAgent("agent1", 10*time.Millisecond, nil)
	agent2 := NewDelayAgent("agent2", 100*time.Millisecond, nil)
	agent3 := NewDelayAgent("agent3", 5*time.Millisecond, nil)

	config := ParallelAgentConfig{Timeout: 50 * time.Millisecond}
	parAgent := NewParallelAgent("test-timeout", config, agent1, agent2, agent3)

	startTime := time.Now()
	finalState, err := parAgent.Run(ctx, initialState)
	duration := time.Since(startTime)

	if err == nil {
		t.Fatalf("ParallelAgent.Run() did not return an error when timeout expected.")
	}

	multiErr, ok := err.(*agenticgokit.MultiError)
	if !ok {
		t.Fatalf("Expected a *MultiError on timeout, but got type %T: %v", err, err)
	}

	foundDeadline := false
	// FIX: Access Errors field directly
	for _, containedErr := range multiErr.Errors {
		if errors.Is(containedErr, context.DeadlineExceeded) {
			foundDeadline = true
			t.Logf("Found expected deadline error within MultiError: %v", containedErr)
			break
		}
	}
	if !foundDeadline {
		t.Errorf("Expected context.DeadlineExceeded to be wrapped in MultiError, but it wasn't: %v", multiErr)
	}
	// FIX: Access Errors field directly for length check
	if len(multiErr.Errors) != 1 {
		t.Errorf("Expected 1 error (timeout) in MultiError, got %d: %v", len(multiErr.Errors), multiErr)
	}

	if duration < config.Timeout || duration > config.Timeout+(30*time.Millisecond) {
		t.Errorf("Execution time (%v) was significantly different from timeout (%v)", duration, config.Timeout)
	}

	if finalState == nil {
		t.Fatal("ParallelAgent.Run() returned nil state even on timeout")
	}

	expectedData := map[string]interface{}{
		"initial": "value",
		"agent1":  "processed_by_agent1",
		"agent3":  "processed_by_agent3",
	}
	finalKeys := filterExecutionDataKeys(finalState.Keys())
	if len(finalKeys) != len(expectedData) {
		t.Errorf("Final state data key count mismatch on timeout: got %d (%v), want %d (%v)", len(finalKeys), finalKeys, len(expectedData), keysFromMap(expectedData))
	}
	for k, expectedV := range expectedData {
		actualV, ok := finalState.Get(k)
		if !ok {
			t.Errorf("Final state missing expected data key on timeout: %s", k)
		} else if !reflect.DeepEqual(actualV, expectedV) {
			t.Errorf("Final state data mismatch for key '%s' on timeout: got %v (%T), want %v (%T)", k, actualV, actualV, expectedV, expectedV)
		}
	}
	expectedMeta := map[string]string{
		"initial_meta": "meta_value",
		"agent1_meta":  "meta_from_agent1",
		"agent3_meta":  "meta_from_agent3",
	}
	finalMetaKeys := finalState.MetaKeys()
	if len(finalMetaKeys) != len(expectedMeta) {
		t.Errorf("Final state metadata key count mismatch on timeout: got %d (%v), want %d (%v)", len(finalMetaKeys), finalMetaKeys, len(expectedMeta), keysFromMapStr(expectedMeta))
	}
	for k, expectedV := range expectedMeta {
		actualV, ok := finalState.GetMeta(k)
		if !ok {
			t.Errorf("Final state missing expected metadata key on timeout: %s", k)
		} else if actualV != expectedV {
			t.Errorf("Final state metadata mismatch for key '%s' on timeout: got %q, want %q", k, actualV, expectedV)
		}
	}
	if agent2.RunCount.Load() == 0 {
		t.Errorf("Agent2 (timed out) Run method was expected to be called at least once, but count is 0")
	}
}

func TestParallelAgent_Run_ExternalContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	initialState := agenticgokit.NewState()
	initialState.Set("initial", "value")
	initialState.SetMeta("initial_meta", "meta_value")

	agent1 := NewDelayAgent("agent1", 100*time.Millisecond, nil)
	agent2 := NewDelayAgent("agent2", 150*time.Millisecond, nil)
	agent3 := NewDelayAgent("agent3", 5*time.Millisecond, nil)

	parAgent := NewParallelAgent("test-cancel", ParallelAgentConfig{}, agent1, agent2, agent3)

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

	multiErr, ok := err.(*agenticgokit.MultiError)
	if !ok {
		t.Fatalf("Expected a *MultiError on cancellation, but got type %T: %v", err, err)
	}

	foundCanceled := 0
	// FIX: Access Errors field directly
	for _, containedErr := range multiErr.Errors {
		if errors.Is(containedErr, context.Canceled) {
			foundCanceled++
			t.Logf("Found expected canceled error within MultiError: %v", containedErr)
		}
	}
	if foundCanceled == 0 {
		t.Errorf("Expected context.Canceled to be wrapped in MultiError, but it wasn't: %v", multiErr)
	}
	// FIX: Access Errors field directly for length check
	if len(multiErr.Errors) != 2 {
		t.Errorf("Expected 2 errors (cancellations) in MultiError, got %d: %v", len(multiErr.Errors), multiErr)
	}

	if duration < 25*time.Millisecond || duration > 70*time.Millisecond {
		t.Errorf("Execution time (%v) was significantly different from expected cancellation time (~30ms)", duration)
	}

	if finalState == nil {
		t.Fatal("ParallelAgent.Run() returned nil state even on cancellation")
	}

	expectedData := map[string]interface{}{
		"initial": "value",
		"agent3":  "processed_by_agent3",
	}
	finalKeys := filterExecutionDataKeys(finalState.Keys())
	if len(finalKeys) != len(expectedData) {
		t.Errorf("Final state data key count mismatch on cancellation: got %d (%v), want %d (%v)", len(finalKeys), finalKeys, len(expectedData), keysFromMap(expectedData))
	}
	for k, expectedV := range expectedData {
		actualV, ok := finalState.Get(k)
		if !ok {
			t.Errorf("Final state missing expected data key on cancellation: %s", k)
		} else if !reflect.DeepEqual(actualV, expectedV) {
			t.Errorf("Final state data mismatch for key '%s' on cancellation: got %v (%T), want %v (%T)", k, actualV, actualV, expectedV, expectedV)
		}
	}
	expectedMeta := map[string]string{
		"initial_meta": "meta_value",
		"agent3_meta":  "meta_from_agent3",
	}
	finalMetaKeys := finalState.MetaKeys()
	if len(finalMetaKeys) != len(expectedMeta) {
		t.Errorf("Final state metadata key count mismatch on cancellation: got %d (%v), want %d (%v)", len(finalMetaKeys), finalMetaKeys, len(expectedMeta), keysFromMapStr(expectedMeta))
	}
	for k, expectedV := range expectedMeta {
		actualV, ok := finalState.GetMeta(k)
		if !ok {
			t.Errorf("Final state missing expected metadata key on cancellation: %s", k)
		} else if actualV != expectedV {
			t.Errorf("Final state metadata mismatch for key '%s' on cancellation: got %q, want %q", k, actualV, expectedV)
		}
	}
	if agent1.RunCount.Load() == 0 {
		t.Errorf("Agent1 (cancelled) Run method was expected to be called at least once, but count is 0")
	}
	if agent2.RunCount.Load() == 0 {
		t.Errorf("Agent2 (cancelled) Run method was expected to be called at least once, but count is 0")
	}
}

func TestParallelAgent_Run_ZeroAgents(t *testing.T) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	initialState.Set("initial", "value")
	initialState.SetMeta("initial_meta", "meta_value")

	parAgent := NewParallelAgent("test-zero", ParallelAgentConfig{})
	finalState, err := parAgent.Run(ctx, initialState)

	// FIX: Check if err is nil
	if err != nil {
		t.Fatalf("Run with zero agents returned an error: %v", err)
	}
	if finalState == nil {
		t.Fatal("Run with zero agents returned nil state")
	}

	if len(finalState.Keys()) != 1 || len(initialState.Keys()) != 1 {
		t.Errorf("Data key count mismatch: got %d, want 1", len(finalState.Keys()))
	} else {
		finalVal, _ := finalState.Get("initial")
		initialVal, _ := initialState.Get("initial")
		if !reflect.DeepEqual(finalVal, initialVal) {
			t.Errorf("Data mismatch: got %v, want %v", finalVal, initialVal)
		}
	}
	if len(finalState.MetaKeys()) != 1 || len(initialState.MetaKeys()) != 1 {
		t.Errorf("Meta key count mismatch: got %d, want 1", len(finalState.MetaKeys()))
	} else {
		finalMetaVal, _ := finalState.GetMeta("initial_meta")
		initialMetaVal, _ := initialState.GetMeta("initial_meta")
		if finalMetaVal != initialMetaVal {
			t.Errorf("Metadata mismatch: got %q, want %q", finalMetaVal, initialMetaVal)
		}
	}
	if finalState == initialState {
		t.Errorf("Run with zero agents returned the exact same state instance, expected a clone.")
	}
}

func TestParallelAgent_Run_NilAgentsFiltered(t *testing.T) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	initialState.Set("initial", "value")
	initialState.SetMeta("initial_meta", "meta_value")

	agent1 := NewDelayAgent("agent1", 10*time.Millisecond, nil)
	agent3 := NewDelayAgent("agent3", 5*time.Millisecond, nil)

	parAgent := NewParallelAgent("test-nil-filtered", ParallelAgentConfig{}, agent1, nil, agent3, nil)

	if len(parAgent.agents) != 2 {
		t.Fatalf("NewParallelAgent did not filter nil agents correctly, got %d, want 2", len(parAgent.agents))
	}

	finalState, err := parAgent.Run(ctx, initialState)

	// FIX: Check if err is nil
	if err != nil {
		t.Fatalf("Run after filtering nil agents returned an error: %v", err)
	}
	if finalState == nil {
		t.Fatal("Run after filtering nil agents returned nil state")
	}

	expectedData := map[string]interface{}{
		"initial": "value",
		"agent1":  "processed_by_agent1",
		"agent3":  "processed_by_agent3",
	}
	finalKeys := filterExecutionDataKeys(finalState.Keys())
	if len(finalKeys) != len(expectedData) {
		t.Errorf("Final state data key count mismatch after filtering nil: got %d (%v), want %d (%v)", len(finalKeys), finalKeys, len(expectedData), keysFromMap(expectedData))
	}
	for k, expectedV := range expectedData {
		actualV, ok := finalState.Get(k)
		if !ok {
			t.Errorf("Final state missing expected data key after filtering nil: %s", k)
		} else if !reflect.DeepEqual(actualV, expectedV) {
			t.Errorf("Final state data mismatch for key '%s' after filtering nil: got %v (%T), want %v (%T)", k, actualV, actualV, expectedV, expectedV)
		}
	}
	expectedMeta := map[string]string{
		"initial_meta": "meta_value",
		"agent1_meta":  "meta_from_agent1",
		"agent3_meta":  "meta_from_agent3",
	}
	finalMetaKeys := finalState.MetaKeys()
	if len(finalMetaKeys) != len(expectedMeta) {
		t.Errorf("Final state metadata key count mismatch after filtering nil: got %d (%v), want %d (%v)", len(finalMetaKeys), finalMetaKeys, len(expectedMeta), keysFromMapStr(expectedMeta))
	}
	for k, expectedV := range expectedMeta {
		actualV, ok := finalState.GetMeta(k)
		if !ok {
			t.Errorf("Final state missing expected metadata key after filtering nil: %s", k)
		} else if actualV != expectedV {
			t.Errorf("Final state metadata mismatch for key '%s' after filtering nil: got %q, want %q", k, actualV, expectedV)
		}
	}
}

// --- Benchmark ---

func BenchmarkParallelAgent_Run(b *testing.B) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	numAgents := 50
	agents := make([]agenticgokit.Agent, numAgents)
	for i := 0; i < numAgents; i++ {
		agents[i] = &NoOpAgent{} // Assumes NoOpAgent is in agents_test_helpers.go
	}
	parAgent := NewParallelAgent("benchmark-parallel", ParallelAgentConfig{}, agents...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parAgent.Run(ctx, initialState.Clone())
		// FIX: Check if err is nil
		if err != nil {
			b.Fatalf("Benchmark run failed unexpectedly: %v", err)
		}
	}
}

// --- Helper Functions ---
func keysFromMap(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func keysFromMapStr(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
