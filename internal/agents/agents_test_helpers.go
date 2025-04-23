package agents

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
)

// --- Common Test Helper Agents ---

// SpyAgent is a simple agent for testing sequential flow.
// It records the input state's data and can be configured to return an error.
type SpyAgent struct {
	Name        string
	ReturnError error
	InputData   map[string]interface{} // Records the data map of the input state
}

func (s *SpyAgent) Run(ctx context.Context, inputState agentflow.State) (agentflow.State, error) {
	s.InputData = inputState.GetData() // Record input data

	if s.ReturnError != nil {
		return inputState, s.ReturnError // Return input state on error
	}

	// Modify the state for the next agent
	outputState := inputState.Clone()
	newData := fmt.Sprintf("processed_by_%s", s.Name)
	outputState.Set(s.Name, newData)          // Add agent-specific data
	outputState.Set("last_processed", s.Name) // Overwrite who processed last

	return outputState, nil
}

// DelayAgent is a simple agent that adds data to the state after a delay.
// Used for testing parallel execution, timeouts, and cancellations.
type DelayAgent struct {
	Name        string
	Delay       time.Duration
	ReturnError error
	RunCount    atomic.Int32           // Track how many times Run was actually invoked
	DataToAdd   map[string]interface{} // Data to add/overwrite in the state
}

func (d *DelayAgent) Run(ctx context.Context, inputState agentflow.State) (agentflow.State, error) {
	d.RunCount.Add(1)
	select {
	case <-time.After(d.Delay):
		// Delay completed
		if d.ReturnError != nil {
			return inputState, d.ReturnError // Return input state on error
		}
		outputState := inputState.Clone()
		if d.DataToAdd != nil {
			for k, v := range d.DataToAdd {
				outputState.Set(k, v)
			}
		} else {
			// Default behavior if DataToAdd is nil
			outputState.Set(d.Name, fmt.Sprintf("processed_by_%s", d.Name))
		}
		return outputState, nil
	case <-ctx.Done():
		// Context cancelled during delay
		return inputState, fmt.Errorf("agent %s cancelled during delay: %w", d.Name, ctx.Err())
	}
}

// CounterAgent increments a "count" value in the state.
// Used for testing LoopAgent.
type CounterAgent struct {
	FailOnCount int // If > 0, return an error when count reaches this value
	ReturnError error
}

func (c *CounterAgent) Run(ctx context.Context, inputState agentflow.State) (agentflow.State, error) {
	outputState := inputState.Clone()
	countVal, _ := outputState.Get("count")
	count, _ := countVal.(int) // Assume int, default 0 if not present or wrong type

	count++
	outputState.Set("count", count)

	if c.FailOnCount > 0 && count == c.FailOnCount {
		if c.ReturnError == nil {
			c.ReturnError = fmt.Errorf("counter agent failed deliberately at count %d", count)
		}
		// Return state *before* the failing increment for consistency in tests
		return inputState, c.ReturnError
	}

	return outputState, nil
}

// NoOpAgent does nothing, used for benchmarking overhead.
type NoOpAgent struct{}

func (n *NoOpAgent) Run(ctx context.Context, inputState agentflow.State) (agentflow.State, error) {
	return inputState, nil // Pass state through
}

// SimpleUpdateAgent increments a specific key in the state immediately.
// Used for LoopAgent cancellation test setup.
type SimpleUpdateAgent struct {
	Key string
}

func (a *SimpleUpdateAgent) Run(ctx context.Context, inputState agentflow.State) (agentflow.State, error) {
	outputState := inputState.Clone()
	countVal, _ := outputState.Get(a.Key)
	count, _ := countVal.(int)
	count++
	outputState.Set(a.Key, count)
	return outputState, nil
}
