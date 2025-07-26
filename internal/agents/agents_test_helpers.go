package agents

import (
	"context" // Ensure errors is imported
	"fmt"     // Ensure log is imported
	"sync"
	"sync/atomic"
	"time"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/internal/core"
)

// --- Common Test Helper Agents ---

// SpyAgent is a simple agent for testing sequential flow.
// It records the input state's data and can be configured to return an error.
type SpyAgent struct {
	name        string // FIX: Rename field to lowercase 'name'
	ReturnError error
	InputData   map[string]interface{} // Records the data map of the input state
}

// NewSpyAgent creates a new SpyAgent.
func NewSpyAgent(name string) *SpyAgent {
	return &SpyAgent{name: name} // FIX: Initialize lowercase 'name'
}

// Name returns the name of the spy agent.
func (s *SpyAgent) Name() string {
	return s.name // FIX: Return lowercase 'name' field
}

func (s *SpyAgent) Run(ctx context.Context, inputState agenticgokit.State) (agenticgokit.State, error) {
	s.InputData = make(map[string]interface{})
	for _, key := range inputState.Keys() {
		if val, ok := inputState.Get(key); ok {
			s.InputData[key] = val
		}
	}

	if s.ReturnError != nil {
		// Wrap the error for context
		return inputState, fmt.Errorf("agent '%s' failed: %w", s.name, s.ReturnError) // FIX: Use s.name
	}

	// Modify the state for the next agent
	outputState := inputState.Clone()
	// FIX: Use s.name in fmt.Sprintf and Set calls
	newData := fmt.Sprintf("processed_by_%s", s.name)
	outputState.Set(s.name, newData)          // Add agent-specific data using its name as key
	outputState.Set("last_processed", s.name) // Overwrite who processed last

	return outputState, nil
}

// DelayAgent simulates work by delaying and optionally returns an error.
type DelayAgent struct {
	name        string // FIX: Rename field to lowercase 'name'
	Delay       time.Duration
	ReturnError error
	RunCount    atomic.Int64 // Track how many times Run was entered
}

// NewDelayAgent creates a new DelayAgent.
func NewDelayAgent(name string, delay time.Duration, returnError error) *DelayAgent {
	return &DelayAgent{ // FIX: Initialize lowercase 'name'
		name:        name,
		Delay:       delay,
		ReturnError: returnError,
	}
}

// Name returns the name of the delay agent.
func (a *DelayAgent) Name() string {
	return a.name // FIX: Return lowercase 'name' field
}

// Run implements the agenticgokit.Agent interface for DelayAgent.
func (a *DelayAgent) Run(ctx context.Context, input agenticgokit.State) (agenticgokit.State, error) {
	a.RunCount.Add(1)
	select {
	case <-time.After(a.Delay):
		if a.ReturnError != nil {
			// FIX: Use a.name
			return nil, fmt.Errorf("agent '%s' failed: %w", a.name, a.ReturnError)
		}
		output := input.Clone()
		// FIX: Use a.name in Set calls and string formatting
		output.Set(a.name, "processed_by_"+a.name)
		output.SetMeta(a.name+"_meta", "meta_from_"+a.name)
		return output, nil
	case <-ctx.Done():
		// FIX: Use a.name
		return nil, fmt.Errorf("agent '%s' cancelled: %w", a.name, ctx.Err())
	}
}

// CounterAgent increments a "count" value in the state.
// Used for testing LoopAgent.
type CounterAgent struct {
	FailOnCount int // If > 0, return an error when count reaches this value
	ReturnError error
}

// Name returns the name of the counter agent.
func (c *CounterAgent) Name() string {
	return "CounterAgent" // Or make it configurable if needed
}

func (c *CounterAgent) Run(ctx context.Context, inputState agenticgokit.State) (agenticgokit.State, error) {
	outputState := inputState.Clone()
	countVal, _ := outputState.Get("count")
	count, _ := countVal.(int) // Assume int, default 0 if not present or wrong type

	count++
	outputState.Set("count", count)

	if c.FailOnCount > 0 && count == c.FailOnCount {
		err := c.ReturnError
		if err == nil {
			err = fmt.Errorf("counter agent failed deliberately at count %d", count)
		}
		// Wrap error with agent name
		return inputState, fmt.Errorf("agent '%s' error: %w", c.Name(), err)
	}

	return outputState, nil
}

// NoOpAgent does nothing, used for benchmarking minimal overhead.
type NoOpAgent struct{}

// Name returns the name of the no-op agent.
func (a *NoOpAgent) Name() string {
	return "NoOpAgent"
}

// Run implements the agenticgokit.Agent interface for NoOpAgent.
func (a *NoOpAgent) Run(ctx context.Context, input agenticgokit.State) (agenticgokit.State, error) {
	return input, nil // Return input state immediately
}

// SimpleUpdateAgent increments a specific key in the state immediately.
// Used for LoopAgent cancellation test setup.
type SimpleUpdateAgent struct {
	Key string
}

// Name returns the name of the simple update agent.
func (a *SimpleUpdateAgent) Name() string {
	return fmt.Sprintf("SimpleUpdateAgent(%s)", a.Key) // More descriptive name
}

func (a *SimpleUpdateAgent) Run(ctx context.Context, inputState agenticgokit.State) (agenticgokit.State, error) {
	outputState := inputState.Clone()
	countVal, _ := outputState.Get(a.Key)
	count, _ := countVal.(int)
	count++
	outputState.Set(a.Key, count)
	return outputState, nil
}

// MockAgent for testing purposes
type MockAgent struct {
	NameVal      string
	RunFunc      func(ctx context.Context, inputState agenticgokit.State) (agenticgokit.State, error) // Adjusted signature
	RunCallCount int
	mu           sync.Mutex
}

// Name returns the name of the mock agent.
func (m *MockAgent) Name() string {
	if m.NameVal == "" {
		return "MockAgent"
	}
	return m.NameVal
}

func (m *MockAgent) Run(ctx context.Context, inputState agenticgokit.State) (agenticgokit.State, error) {
	m.mu.Lock()
	m.RunCallCount++
	m.mu.Unlock()
	if m.RunFunc != nil {
		// FIX: Use Keys() and Get() to access data if needed by RunFunc logic
		// Example: Log input data
		// for _, key := range inputState.Keys() {
		//  val, _ := inputState.Get(key)
		//  fmt.Printf("MockAgent %s received input %s: %v\n", m.NameVal, key, val)
		// }
		return m.RunFunc(ctx, inputState)
	}
	// Default behavior: return input state
	return inputState, nil
}

func (m *MockAgent) GetRunCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.RunCallCount
}

// Helper to create a simple state for testing
// Use agenticgokit.State and agenticgokit.NewState
func createState(data map[string]any, meta map[string]string) agenticgokit.State {
	s := agenticgokit.NewState()
	if data != nil {
		for k, v := range data {
			s.Set(k, v)
		}
	}
	if meta != nil {
		for k, v := range meta {
			s.SetMeta(k, v)
		}
	}
	return s
}