package workflow_test

import (
	"context"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConditions_OutputContains tests the OutputContains condition builder
func TestConditions_OutputContains(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{
		"Working on it...",
		"Still working...",
		"Done! APPROVED",
	})

	condition := vnext.Conditions.OutputContains("APPROVED")

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		MaxIterations: 10,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, 3, result.IterationInfo.TotalIterations)
	assert.Equal(t, vnext.ExitConditionMet, result.IterationInfo.ExitReason)
	assert.Contains(t, result.FinalOutput, "APPROVED")
}

// TestConditions_MaxTokens tests the MaxTokens condition builder
func TestConditions_MaxTokens(t *testing.T) {
	agent := &mockAgentWithTokens{
		name:          "agent",
		outputText:    "response",
		tokensPerCall: 100,
	}

	condition := vnext.Conditions.MaxTokens(250) // Should stop after 2-3 iterations

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		MaxIterations: 10,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.LessOrEqual(t, result.IterationInfo.TotalIterations, 3)
	assert.Equal(t, vnext.ExitConditionMet, result.IterationInfo.ExitReason)
}

// TestConditions_Convergence tests the Convergence condition builder
func TestConditions_Convergence(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{
		"First draft of the story",
		"First draft of the story with minor edits",
		"First draft of the story with minor edits and polish",
		"First draft of the story with minor edits and polish.", // Very small change
	})

	condition := vnext.Conditions.Convergence(0.1) // Stop when changes are < 10%

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		MaxIterations: 10,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, 4, result.IterationInfo.TotalIterations)
	assert.Equal(t, vnext.ExitConditionMet, result.IterationInfo.ExitReason)
}

// TestConditions_And tests the And combinator
func TestConditions_And(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{
		"iteration 0",
		"iteration 1",
		"iteration 2 APPROVED",
	})

	condition := vnext.Conditions.And(
		vnext.Conditions.OutputContains("APPROVED"),
		func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
			return iteration < 5, nil // Must be less than 5 iterations
		},
	)

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		MaxIterations: 10,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, result.IterationInfo.TotalIterations)
	assert.Equal(t, vnext.ExitConditionMet, result.IterationInfo.ExitReason)
}

// TestConditions_Or tests the Or combinator
func TestConditions_Or(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{
		"iteration 0",
		"iteration 1",
		"iteration 2",
	})

	// OR: Continue if EITHER condition says continue
	// This creates a condition that stops after 2 iterations OR if we see "NEVERFOUND"
	condition := vnext.Conditions.Or(
		func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
			return iteration < 2, nil // Continue for first 2 iterations
		},
		vnext.Conditions.OutputContains("NEVERFOUND"), // Will always return true (continue) since not found
	)

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		MaxIterations: 10,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	require.NoError(t, err)
	assert.NotNil(t, result)
	// Should continue beyond 2 iterations because OutputContains("NEVERFOUND") always says continue
	assert.Greater(t, result.IterationInfo.TotalIterations, 2)
}

// TestConditions_Not tests the Not combinator
func TestConditions_Not(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{
		"iteration 0",
		"iteration 1",
		"iteration 2",
	})

	// Create a condition that stops after 5 iterations
	// NOT of that should stop BEFORE 5 iterations
	// Actually, NOT inverts the logic: if original says "continue", NOT says "stop"
	// So NOT(iteration < 5) means: continue when iteration >= 5
	// But we check BEFORE execution, so at iteration 0: 0<5 is true, NOT(true)=false, stop immediately

	// Let's test NOT properly: stop when a condition is true
	stopWhenGreaterThan2 := func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		return iteration <= 2, nil // Continue while <= 2 (stop when > 2)
	}

	condition := vnext.Conditions.Not(stopWhenGreaterThan2)
	// NOT(continue while <=2) = stop while <=2 = continue when >2
	// But at iter 0: 0<=2 is true (continue), NOT(true)=false (stop)

	// This is confusing. Let me use a clearer example:
	// Condition that says "stop after first iteration"
	stopAfterFirst := func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		return iteration == 0, nil // Only continue on iteration 0
	}
	condition = vnext.Conditions.Not(stopAfterFirst)
	// NOT(only continue on 0) = don't continue on 0 = skip iteration 0
	// At iter 0: (0==0)=true, NOT(true)=false, stop

	// OK, let me think differently. NOT should just invert whatever the condition returns.
	// Simple test: run 3 iterations total
	continueForFirst3 := func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		return iteration >= 3, nil // Continue when >= 3 (which is never for first few)
	}
	condition = vnext.Conditions.Not(continueForFirst3)
	// At iter 0: (0>=3)=false (stop), NOT(false)=true (continue)
	// At iter 1: (1>=3)=false (stop), NOT(false)=true (continue)
	// At iter 2: (2>=3)=false (stop), NOT(false)=true (continue)
	// At iter 3: (3>=3)=true (continue), NOT(true)=false (stop)

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		MaxIterations: 5,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, result.IterationInfo.TotalIterations) // Should stop at iteration 3
	assert.Equal(t, vnext.ExitConditionMet, result.IterationInfo.ExitReason)
}

// TestConditions_Custom tests the Custom wrapper
func TestConditions_Custom(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{
		"a",
		"b",
		"c",
	})

	condition := vnext.Conditions.Custom(func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
		// Stop after 2 iterations
		return iteration < 2, nil
	})

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		MaxIterations: 10,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.IterationInfo.TotalIterations)
	assert.Equal(t, vnext.ExitConditionMet, result.IterationInfo.ExitReason)
}

// TestConditions_ComplexComposition tests complex condition composition
func TestConditions_ComplexComposition(t *testing.T) {
	agent := newMockSequenceAgent("agent", []string{
		"Working...",
		"Still working...",
		"Almost done...",
		"APPROVED by editor",
	})

	// Continue while: (NOT found APPROVED) AND (iteration < 10)
	// OutputContains("APPROVED") returns false (stop) when found
	// NOT(OutputContains("APPROVED")) returns true (continue) when found - that's backwards!
	//
	// Let's use a simpler composition: stop when approved OR max iterations
	condition := vnext.Conditions.And(
		vnext.Conditions.OutputContains("APPROVED"), // Continue while NOT found (returns !found)
		func(ctx context.Context, iteration int, lastResult *vnext.WorkflowResult) (bool, error) {
			return iteration < 10, nil // Continue while < 10
		},
	)

	workflow, err := vnext.NewLoopWorkflowWithCondition(&vnext.WorkflowConfig{
		MaxIterations: 20,
	}, condition)
	require.NoError(t, err)

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step", Agent: agent})
	require.NoError(t, err)

	result, err := workflow.Run(context.Background(), "input")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 4, result.IterationInfo.TotalIterations)
	assert.Equal(t, vnext.ExitConditionMet, result.IterationInfo.ExitReason)
	assert.Contains(t, result.FinalOutput, "APPROVED")
}

// Mock agent that returns specific token counts
type mockAgentWithTokens struct {
	name          string
	outputText    string
	tokensPerCall int
	callCount     int
	config        *vnext.Config
}

func (m *mockAgentWithTokens) Run(ctx context.Context, input string) (*vnext.Result, error) {
	m.callCount++
	return &vnext.Result{
		Success:    true,
		Content:    m.outputText,
		TokensUsed: m.tokensPerCall,
		Metadata: map[string]interface{}{
			"call_count": m.callCount,
		},
	}, nil
}

func (m *mockAgentWithTokens) RunWithOptions(ctx context.Context, input string, opts *vnext.RunOptions) (*vnext.Result, error) {
	return m.Run(ctx, input)
}

func (m *mockAgentWithTokens) RunStream(ctx context.Context, input string, opts ...vnext.StreamOption) (vnext.Stream, error) {
	metadata := &vnext.StreamMetadata{
		AgentName: m.name,
		StartTime: time.Now(),
	}
	stream, writer := vnext.NewStream(ctx, metadata, opts...)

	go func() {
		defer writer.Close()
		writer.Write(&vnext.StreamChunk{
			Type:    vnext.ChunkTypeText,
			Content: m.outputText,
		})
	}()

	return stream, nil
}

func (m *mockAgentWithTokens) RunStreamWithOptions(ctx context.Context, input string, runOpts *vnext.RunOptions, streamOpts ...vnext.StreamOption) (vnext.Stream, error) {
	return m.RunStream(ctx, input, streamOpts...)
}

func (m *mockAgentWithTokens) Name() string {
	return m.name
}

func (m *mockAgentWithTokens) Config() *vnext.Config {
	if m.config == nil {
		m.config = &vnext.Config{Name: m.name}
	}
	return m.config
}

func (m *mockAgentWithTokens) Capabilities() []string {
	return []string{"chat"}
}

func (m *mockAgentWithTokens) Initialize(ctx context.Context) error {
	return nil
}

func (m *mockAgentWithTokens) Cleanup(ctx context.Context) error {
	return nil
}
