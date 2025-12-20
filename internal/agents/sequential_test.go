package agents

import (
	"context"
	"errors"
	"reflect"
	"testing"

	agenticgokit "github.com/agenticgokit/agenticgokit/internal/core"
)

// --- Test Helper Agents ---
// Assume SpyAgent is defined in agents_test_helpers.go

// --- Test Cases ---

func TestSequentialAgent_Run_AllSuccess(t *testing.T) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	initialState.Set("initial", "value")

	agent1 := NewSpyAgent("agent1")
	agent2 := NewSpyAgent("agent2")
	agent3 := NewSpyAgent("agent3")

	seqAgent := NewSequentialAgent("test-all-success", agent1, agent2, agent3)
	finalState, err := seqAgent.Run(ctx, initialState)

	if err != nil {
		t.Fatalf("SequentialAgent.Run() returned an unexpected error: %v", err)
	}
	if finalState == nil {
		t.Fatal("SequentialAgent.Run() returned nil state")
	}

	// Verify final state data
	expectedData := map[string]interface{}{
		"initial":        "value",
		"agent1":         "processed_by_agent1",
		"agent2":         "processed_by_agent2",
		"agent3":         "processed_by_agent3",
		"last_processed": "agent3", // Agent 3 was last
	}
	// Collect only expected keys (ignore execution metadata)
	finalData := make(map[string]interface{})
	for k := range expectedData {
		if val, ok := finalState.Get(k); ok {
			finalData[k] = val
		}
	}
	if !reflect.DeepEqual(finalData, expectedData) {
		t.Errorf("Final state data mismatch:\ngot:  %v\nwant: %v", finalData, expectedData)
	}

	// Verify intermediate states recorded by SpyAgents
	if agent1.InputData == nil || agent1.InputData["initial"] != "value" || len(agent1.InputData) != 1 {
		t.Errorf("Agent1 did not receive the expected initial state: got %v", agent1.InputData)
	}
	expectedAgent2Input := map[string]interface{}{"initial": "value", "agent1": "processed_by_agent1", "last_processed": "agent1"}
	if !reflect.DeepEqual(agent2.InputData, expectedAgent2Input) {
		t.Errorf("Agent2 did not receive the expected state from Agent1:\ngot:  %v\nwant: %v", agent2.InputData, expectedAgent2Input)
	}
	expectedAgent3Input := map[string]interface{}{"initial": "value", "agent1": "processed_by_agent1", "agent2": "processed_by_agent2", "last_processed": "agent2"}
	if !reflect.DeepEqual(agent3.InputData, expectedAgent3Input) {
		t.Errorf("Agent3 did not receive the expected state from Agent2:\ngot:  %v\nwant: %v", agent3.InputData, expectedAgent3Input)
	}
}

func TestSequentialAgent_Run_PartialFailure(t *testing.T) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	initialState.Set("initial", "value")

	simulatedError := errors.New("agent2 failed deliberately")
	agent1 := NewSpyAgent("agent1")
	agent2 := NewSpyAgent("agent2")
	agent2.ReturnError = simulatedError // Set error after creation
	agent3 := NewSpyAgent("agent3")     // Agent 3 should not run

	seqAgent := NewSequentialAgent("test-partial-fail", agent1, agent2, agent3)
	finalState, err := seqAgent.Run(ctx, initialState)

	if err == nil {
		t.Fatalf("SequentialAgent.Run() did not return an error when expected.")
	}

	// Check if the specific error is returned (SequentialAgent should return it directly)
	if !errors.Is(err, simulatedError) {
		t.Errorf("Expected error '%v', but got: %v", simulatedError, err)
	}
	t.Logf("Received expected error: %v", err)

	// Verify final state - should be the state *before* the failing agent (output of agent1)
	expectedData := map[string]interface{}{
		"initial":        "value",
		"agent1":         "processed_by_agent1",
		"last_processed": "agent1",
	}
	// Collect only expected keys (ignore execution metadata)
	finalData := make(map[string]interface{})
	if finalState == nil {
		t.Log("Final state was nil on partial failure, cannot compare data.")
		// Depending on desired behavior, you might want to fail here if state shouldn't be nil
	} else {
		for k := range expectedData {
			if val, ok := finalState.Get(k); ok {
				finalData[k] = val
			}
		}
	}

	if finalState != nil && !reflect.DeepEqual(finalData, expectedData) {
		t.Errorf("Final state data mismatch on partial failure:\ngot:  %v\nwant: %v", finalData, expectedData)
	}

	// Verify agent3 did not run
	if agent3.InputData != nil {
		t.Errorf("Agent3 should not have run after Agent2 failed, but its InputData is: %v", agent3.InputData)
	}
}

func TestSequentialAgent_Run_ZeroAgents(t *testing.T) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	initialState.Set("initial", "value")

	seqAgent := NewSequentialAgent("test-zero") // No agents
	finalState, err := seqAgent.Run(ctx, initialState)

	if err != nil {
		t.Fatalf("Run with zero agents returned an error: %v", err)
	}
	if finalState == nil {
		t.Fatal("Run with zero agents returned nil state")
	}

	// Should return the initial state unmodified (or a clone)
	// FIX: Remove unused expectedData variable
	// expectedData := map[string]interface{}{"initial": "value"} // REMOVE THIS LINE

	finalData := make(map[string]interface{})
	if val, ok := finalState.Get("initial"); ok {
		finalData["initial"] = val
	}
	initialData := make(map[string]interface{})
	for _, key := range initialState.Keys() { // Compare against original initialState data
		if val, ok := initialState.Get(key); ok {
			initialData[key] = val
		}
	}

	if !reflect.DeepEqual(finalData, initialData) {
		t.Errorf("Final state data mismatch for zero agents:\ngot:  %v\nwant: %v", finalData, initialData)
	}
}

func TestSequentialAgent_Run_NilAgentsFiltered(t *testing.T) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	initialState.Set("initial", "value")

	agent1 := NewSpyAgent("agent1")
	agent3 := NewSpyAgent("agent3")

	// Pass nil agents to the constructor
	seqAgent := NewSequentialAgent("test-nil-filtered", agent1, nil, agent3, nil)

	if seqAgent == nil {
		t.Fatal("NewSequentialAgent returned nil unexpectedly")
	}
	// Check internal agents slice if accessible, otherwise rely on behavior
	// if len(seqAgent.agents) != 2 {
	//  t.Fatalf("NewSequentialAgent did not filter nil agents correctly, got %d, want 2", len(seqAgent.agents))
	// }

	finalState, err := seqAgent.Run(ctx, initialState)

	if err != nil {
		t.Fatalf("Run after filtering nil agents returned an error: %v", err)
	}
	if finalState == nil {
		t.Fatal("Run after filtering nil agents returned nil state")
	}

	// Verify final state data - should contain initial data + data from non-nil agents (1 and 3)
	expectedData := map[string]interface{}{
		"initial":        "value",
		"agent1":         "processed_by_agent1",
		"agent3":         "processed_by_agent3",
		"last_processed": "agent3", // Agent 3 was last
	}
	finalData := make(map[string]interface{})
	for k := range expectedData {
		if val, ok := finalState.Get(k); ok {
			finalData[k] = val
		}
	}
	if !reflect.DeepEqual(finalData, expectedData) {
		t.Errorf("Final state data mismatch after filtering nil:\ngot:  %v\nwant: %v", finalData, expectedData)
	}

	// Verify intermediate states
	expectedAgent3Input := map[string]interface{}{"initial": "value", "agent1": "processed_by_agent1", "last_processed": "agent1"}
	if !reflect.DeepEqual(agent3.InputData, expectedAgent3Input) {
		t.Errorf("Agent3 did not receive the expected state from Agent1 after filtering:\ngot:  %v\nwant: %v", agent3.InputData, expectedAgent3Input)
	}
}

// --- Benchmark ---

func BenchmarkSequentialAgent_Run(b *testing.B) {
	ctx := context.Background()
	initialState := agenticgokit.NewState()
	numAgents := 5
	agents := make([]agenticgokit.Agent, numAgents)
	for i := 0; i < numAgents; i++ {
		agents[i] = &NoOpAgent{}
	}
	seqAgent := NewSequentialAgent("benchmark", agents...)

	b.ResetTimer() // Start timing after setup
	for i := 0; i < b.N; i++ {
		// It's important to pass a clone in benchmark to avoid state pollution across iterations
		_, err := seqAgent.Run(ctx, initialState.Clone())
		if err != nil {
			b.Fatalf("Benchmark run failed: %v", err) // Fail benchmark on error
		}
	}
}

