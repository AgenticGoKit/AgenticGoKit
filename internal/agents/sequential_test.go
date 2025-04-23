package agents

import (
	"context"
	"errors"
	"reflect"
	"testing"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// --- Test Helper Agent ---

// --- Test Cases ---

func TestSequentialAgent_Run_Success(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("initial", "data")

	agent1 := &SpyAgent{Name: "agent1"}
	agent2 := &SpyAgent{Name: "agent2"}
	agent3 := &SpyAgent{Name: "agent3"}

	seqAgent := NewSequentialAgent("test-success", agent1, agent2, agent3)

	finalState, err := seqAgent.Run(ctx, initialState)

	if err != nil {
		t.Fatalf("SequentialAgent.Run() returned an unexpected error: %v", err)
	}

	// Verify final state data
	expectedData := map[string]interface{}{
		"initial":        "data",
		"agent1":         "processed_by_agent1",
		"agent2":         "processed_by_agent2",
		"agent3":         "processed_by_agent3",
		"last_processed": "agent3", // Agent 3 was the last one
	}
	if !reflect.DeepEqual(finalState.GetData(), expectedData) {
		t.Errorf("Final state data mismatch:\ngot:  %#v\nwant: %#v", finalState.GetData(), expectedData)
	}

	// Verify intermediate inputs (optional but good for understanding flow)
	if agent1.InputData["initial"] != "data" {
		t.Errorf("Agent1 did not receive initial data correctly")
	}
	if agent2.InputData["agent1"] != "processed_by_agent1" {
		t.Errorf("Agent2 did not receive agent1's output correctly")
	}
	if agent3.InputData["agent2"] != "processed_by_agent2" {
		t.Errorf("Agent3 did not receive agent2's output correctly")
	}
}

func TestSequentialAgent_Run_Failure(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("initial", "data")

	simulatedError := errors.New("agent2 failed")
	agent1 := &SpyAgent{Name: "agent1"}
	agent2 := &SpyAgent{Name: "agent2", ReturnError: simulatedError} // Agent 2 will fail
	agent3 := &SpyAgent{Name: "agent3"}

	seqAgent := NewSequentialAgent("test-failure", agent1, agent2, agent3)

	finalState, err := seqAgent.Run(ctx, initialState)

	if err == nil {
		t.Fatalf("SequentialAgent.Run() did not return an error when expected.")
	}

	// Check if the specific error is wrapped
	if !errors.Is(err, simulatedError) {
		t.Errorf("Expected error '%v' to be wrapped, but got: %v", simulatedError, err)
	}
	t.Logf("Received expected error: %v", err) // Log the wrapped error

	// Verify the state returned is the state *before* the failing agent (agent2) ran
	// This means it should be the output state of agent1.
	expectedData := map[string]interface{}{
		"initial":        "data",
		"agent1":         "processed_by_agent1",
		"last_processed": "agent1", // Agent 1 was the last successful one
	}
	if !reflect.DeepEqual(finalState.GetData(), expectedData) {
		t.Errorf("State before failure mismatch:\ngot:  %#v\nwant: %#v", finalState.GetData(), expectedData)
	}

	// Verify agent3 was never run (its InputData should be nil)
	if agent3.InputData != nil {
		t.Errorf("Agent3 should not have been run after agent2 failed, but its InputData is not nil: %#v", agent3.InputData)
	}
}

func TestSequentialAgent_Run_EdgeCases(t *testing.T) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	initialState.Set("initial", "data")

	t.Run("ZeroSubAgents", func(t *testing.T) {
		seqAgent := NewSequentialAgent("test-zero")                // No agents added
		finalState, err := seqAgent.Run(ctx, initialState.Clone()) // Pass clone

		if err != nil {
			t.Fatalf("Run with zero agents returned an error: %v", err)
		}
		// Should return the initial state unmodified
		if !reflect.DeepEqual(finalState.GetData(), initialState.GetData()) {
			t.Errorf("Run with zero agents modified the state:\ngot:  %#v\nwant: %#v", finalState.GetData(), initialState.GetData())
		}
	})

	t.Run("NilAgentEntries", func(t *testing.T) {
		agent1 := &SpyAgent{Name: "agent1"}
		agent3 := &SpyAgent{Name: "agent3"}
		// Pass nil explicitly in the middle
		seqAgent := NewSequentialAgent("test-nil", agent1, nil, agent3)

		if len(seqAgent.agents) != 2 {
			t.Fatalf("NewSequentialAgent did not filter out nil agent. Got %d agents, want 2.", len(seqAgent.agents))
		}

		finalState, err := seqAgent.Run(ctx, initialState.Clone()) // Pass clone

		if err != nil {
			t.Fatalf("Run with filtered nil agents returned an error: %v", err)
		}

		// Verify final state data (only agent1 and agent3 should run)
		expectedData := map[string]interface{}{
			"initial":        "data",
			"agent1":         "processed_by_agent1",
			"agent3":         "processed_by_agent3", // Added by agent3
			"last_processed": "agent3",              // Agent 3 was the last one
		}
		if !reflect.DeepEqual(finalState.GetData(), expectedData) {
			t.Errorf("Final state data mismatch after filtering nil:\ngot:  %#v\nwant: %#v", finalState.GetData(), expectedData)
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		agent1 := &SpyAgent{Name: "agent1"}
		agent2 := &SpyAgent{Name: "agent2"} // This one won't run

		seqAgent := NewSequentialAgent("test-cancel", agent1, agent2)

		// Run agent1 successfully - we don't need its output state for this specific test logic
		_, err := agent1.Run(ctx, initialState.Clone()) // Use blank identifier _
		if err != nil {
			t.Fatalf("Agent1 failed unexpectedly: %v", err)
		}

		// Cancel the context *before* running the sequential agent's second step
		cancel()

		// Run the sequential agent (it should detect cancellation before agent2)
		// Start from initial again for simplicity, as the test focuses on cancellation detection
		finalState, err := seqAgent.Run(ctx, initialState.Clone())

		if err == nil {
			t.Fatalf("SequentialAgent did not return context cancellation error")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got: %v", err)
		}

		// The returned state should be the state *before* cancellation was detected
		// In this run, it started from initial state and cancelled before agent1 ran within seqAgent
		if !reflect.DeepEqual(finalState.GetData(), initialState.GetData()) {
			t.Errorf("State on cancellation mismatch:\ngot:  %#v\nwant: %#v", finalState.GetData(), initialState.GetData())
		}

		// Verify agent2 was not run
		if agent2.InputData != nil {
			t.Errorf("Agent2 should not have run after context cancellation")
		}
	})
}

// --- Benchmark ---

func BenchmarkSequentialAgent_Run(b *testing.B) {
	ctx := context.Background()
	initialState := agentflow.NewState()
	numAgents := 5
	agents := make([]agentflow.Agent, numAgents)
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
