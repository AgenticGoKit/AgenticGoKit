package workflow_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSequenceAgent returns different outputs based on call count
type mockSequenceAgent struct {
	name    string
	outputs []string
	callNum int
	config  *vnext.Config
}

func newMockSequenceAgent(name string, outputs []string) *mockSequenceAgent {
	return &mockSequenceAgent{
		name:    name,
		outputs: outputs,
		config:  &vnext.Config{Name: name},
	}
}

func (m *mockSequenceAgent) Run(ctx context.Context, input string) (*vnext.Result, error) {
	if m.callNum >= len(m.outputs) {
		m.callNum = len(m.outputs) - 1 // Use last output if exceeded
	}
	output := m.outputs[m.callNum]
	m.callNum++

	return &vnext.Result{
		Success: true,
		Content: output,
	}, nil
}

func (m *mockSequenceAgent) RunWithOptions(ctx context.Context, input string, opts *vnext.RunOptions) (*vnext.Result, error) {
	return m.Run(ctx, input)
}

func (m *mockSequenceAgent) RunStream(ctx context.Context, input string, opts ...vnext.StreamOption) (vnext.Stream, error) {
	result, err := m.Run(ctx, input)
	if err != nil {
		return nil, err
	}
	// Create a simple stream using vnext.NewStream
	metadata := &vnext.StreamMetadata{
		AgentName: m.name,
		StartTime: time.Now(),
	}
	stream, writer := vnext.NewStream(ctx, metadata, opts...)

	// Write the result as a single chunk
	go func() {
		defer writer.Close()
		writer.Write(&vnext.StreamChunk{
			Type:    vnext.ChunkTypeText,
			Content: result.Content,
		})
	}()

	return stream, nil
}

func (m *mockSequenceAgent) RunStreamWithOptions(ctx context.Context, input string, runOpts *vnext.RunOptions, streamOpts ...vnext.StreamOption) (vnext.Stream, error) {
	return m.RunStream(ctx, input, streamOpts...)
}

func (m *mockSequenceAgent) Name() string {
	return m.name
}

func (m *mockSequenceAgent) Config() *vnext.Config {
	return m.config
}

func (m *mockSequenceAgent) Capabilities() []string {
	return []string{"chat"}
}

func (m *mockSequenceAgent) Initialize(ctx context.Context) error {
	return nil
}

func (m *mockSequenceAgent) Cleanup(ctx context.Context) error {
	return nil
} // TestLoopWorkflow_OutputContainsCondition tests loop exit on specific text
func TestLoopWorkflow_OutputContainsCondition(t *testing.T) {
	// Create agents that return different outputs
	writer := newMockSequenceAgent("writer", []string{
		"Draft 1",
		"Draft 2 (revised)",
		"Draft 3 (revised)",
	})

	editor := newMockSequenceAgent("editor", []string{
		"Needs improvement",
		"APPROVED: Draft 2 looks good!",
		"APPROVED: Draft 3 looks good!",
	})

	// Create loop workflow with approval condition
	condition := func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		if lastResult != nil && strings.Contains(lastResult.FinalOutput, "APPROVED") {
			return false, nil // Stop looping
		}
		return true, nil // Continue
	}

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		Mode:          vnext.Loop,
		MaxIterations: 5,
		Timeout:       10 * time.Second,
	}, condition)
	require.NoError(t, err)

	// Add steps
	err = workflow.AddStep(vnext.WorkflowStep{Name: "write", Agent: writer})
	require.NoError(t, err)
	err = workflow.AddStep(vnext.WorkflowStep{Name: "review", Agent: editor})
	require.NoError(t, err)

	// Execute
	result, err := workflow.Run(context.Background(), "Write a story")

	// Assertions
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.IterationInfo)

	// Note: 2 full iterations were executed (0 and 1)
	// Iteration 0: Draft 1 -> "Needs improvement"
	// Iteration 1: Draft 2 -> "APPROVED: Draft 2 looks good!"
	// Iteration 2: Condition checked, sees APPROVED, stops before executing
	// But since we check BEFORE iteration 2, the condition returns false at the START of iteration 2
	// So we record that we were about to do iteration 2 but stopped
	assert.Equal(t, 2, result.IterationInfo.TotalIterations) // 2 iterations completed
	assert.Equal(t, vnext.ExitConditionMet, result.IterationInfo.ExitReason)
	assert.Contains(t, result.FinalOutput, "APPROVED")
}

// TestLoopWorkflow_MaxIterationsReached tests that loop respects max iterations
func TestLoopWorkflow_MaxIterationsReached(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{
		"Output 1",
		"Output 2",
		"Output 3",
		"Output 4",
	})

	// Condition that never triggers (always continue)
	condition := func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		return true, nil // Always continue
	}

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		Mode:          vnext.Loop,
		MaxIterations: 3,
		Timeout:       10 * time.Second,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.IterationInfo)
	assert.Equal(t, 3, result.IterationInfo.TotalIterations)
	assert.Equal(t, vnext.ExitMaxIterations, result.IterationInfo.ExitReason)
}

// TestLoopWorkflow_ConditionError tests error handling in conditions
func TestLoopWorkflow_ConditionError(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{"output 1", "output 2"})

	// Condition that returns error on 2nd iteration
	condition := func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		if iteration == 1 {
			return false, fmt.Errorf("condition check failed")
		}
		return true, nil
	}

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		Mode:          vnext.Loop,
		MaxIterations: 5,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "condition check failed")
	assert.NotNil(t, result)
	assert.NotNil(t, result.IterationInfo)
	// Iteration 0 executed successfully, then condition checked at iteration 1 and failed
	assert.Equal(t, 1, result.IterationInfo.TotalIterations) // 1 iteration was executed before error
	assert.Equal(t, vnext.ExitError, result.IterationInfo.ExitReason)
}

// TestLoopWorkflow_SetLoopCondition tests setting condition after creation
func TestLoopWorkflow_SetLoopCondition(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{"v1", "DONE", "v3"})

	workflow, err := vnext.NewLoopWorkflow(&vnext.WorkflowConfig{
		MaxIterations: 5,
	})
	require.NoError(t, err)

	// Set condition after creation
	condition := func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		if lastResult != nil && lastResult.FinalOutput == "DONE" {
			return false, nil
		}
		return true, nil
	}

	err = workflow.SetLoopCondition(condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	assert.NoError(t, err)
	assert.Equal(t, 2, result.IterationInfo.TotalIterations)
	assert.Equal(t, vnext.ExitConditionMet, result.IterationInfo.ExitReason)
}

// TestLoopWorkflow_SetLoopCondition_WrongMode tests error when setting condition on non-loop workflow
func TestLoopWorkflow_SetLoopCondition_WrongMode(t *testing.T) {
	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	require.NoError(t, err)

	condition := func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		return true, nil
	}

	err = workflow.SetLoopCondition(condition)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only valid for Loop mode")
}

// TestLoopWorkflow_IterationTracking tests that iterations are correctly tracked
func TestLoopWorkflow_IterationTracking(t *testing.T) {
	callCount := 0
	agent := newMockSequenceAgent("agent", []string{"a", "b", "c"})

	// Condition that stops after 2 iterations
	condition := func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		callCount++
		return iteration < 2, nil // Stop after iteration 2
	}

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		MaxIterations: 10,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	assert.NoError(t, err)
	assert.Equal(t, 2, result.IterationInfo.TotalIterations)
	assert.Equal(t, 1, result.IterationInfo.LastIterationNum) // 0-indexed
	assert.Equal(t, 3, callCount)                             // Called before iteration 0, 1, and 2
}

// TestLoopWorkflow_NoCondition tests backward compatibility (no condition set)
func TestLoopWorkflow_NoCondition(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{"a", "b", "c", "d"})

	workflow, err := vnext.NewLoopWorkflow(&vnext.WorkflowConfig{
		MaxIterations: 3,
	})
	require.NoError(t, err)

	// Don't set any condition - should use max iterations
	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	assert.NoError(t, err)
	assert.Equal(t, 3, result.IterationInfo.TotalIterations)
	assert.Equal(t, vnext.ExitMaxIterations, result.IterationInfo.ExitReason)
}

// TestLoopWorkflow_ContextCancellation tests that loop respects context cancellation
func TestLoopWorkflow_ContextCancellation(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{"a", "b", "c"})

	condition := func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		// Add a small delay to allow context cancellation to happen
		time.Sleep(100 * time.Millisecond)
		return true, nil // Always continue
	}

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		MaxIterations: 100,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := workflow.Run(ctx, "input")

	assert.Error(t, err)
	assert.True(t, err == context.DeadlineExceeded || strings.Contains(err.Error(), "deadline"))
	assert.NotNil(t, result)
	if result != nil {
		assert.NotNil(t, result.IterationInfo)
		assert.Equal(t, vnext.ExitContextCancelled, result.IterationInfo.ExitReason)
	}
}
