package builder_test

import (
	"context"
	"strings"
	"testing"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

// TestBuilderSubWorkflow tests the WithSubWorkflow builder method
func TestBuilderSubWorkflow(t *testing.T) {
	// Create simple agents
	agent1 := &mockAgent{name: "agent1"}
	agent2 := &mockAgent{name: "agent2"}

	// Create a sequential workflow
	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
	if err != nil {
		t.Fatalf("Failed to add step1: %v", err)
	}

	err = workflow.AddStep(vnext.WorkflowStep{Name: "step2", Agent: agent2})
	if err != nil {
		t.Fatalf("Failed to add step2: %v", err)
	}

	// Wrap workflow as agent using builder
	subAgent, err := vnext.NewBuilder("sub-agent").
		WithSubWorkflow(
			vnext.WithWorkflowInstance(workflow),
			vnext.WithSubWorkflowMaxDepthBuilder(5),
			vnext.WithSubWorkflowDescriptionBuilder("Test SubWorkflow"),
		).
		Build()

	if err != nil {
		t.Fatalf("Failed to build SubWorkflow agent: %v", err)
	}

	if subAgent == nil {
		t.Fatal("SubWorkflow agent should not be nil")
	}

	// Verify it's a SubWorkflowAgent
	if subAgent.Name() != "sub-agent" {
		t.Errorf("Expected agent name 'sub-agent', got '%s'", subAgent.Name())
	}

	// Test execution
	result, err := subAgent.Run(context.Background(), "test input")
	if err != nil {
		t.Errorf("SubWorkflow execution failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}
}

// TestBuilderSubWorkflowWithoutInstance tests error when workflow instance is missing
func TestBuilderSubWorkflowWithoutInstance(t *testing.T) {
	_, err := vnext.NewBuilder("sub-agent").
		WithSubWorkflow(
			vnext.WithSubWorkflowMaxDepthBuilder(5),
		).
		Build()

	if err == nil {
		t.Fatal("Expected error when building SubWorkflow without workflow instance")
	}

	if !strings.Contains(err.Error(), "workflow instance") {
		t.Errorf("Error should mention 'workflow instance', got: %s", err.Error())
	}
}

// TestBuilderSubWorkflowWithLoop tests SubWorkflow with conditional loop
func TestBuilderSubWorkflowWithLoop(t *testing.T) {
	// Create agents for loop
	agent1 := &mockAgent{name: "writer", response: "Draft story"}
	agent2 := &mockAgent{name: "editor", response: "APPROVED: Good story"}

	// Create loop workflow with condition
	loopWorkflow, err := vnext.NewLoopWorkflowWithCondition(
		&vnext.WorkflowConfig{MaxIterations: 3},
		vnext.Conditions.OutputContains("APPROVED"),
	)
	if err != nil {
		t.Fatalf("Failed to create loop workflow: %v", err)
	}

	loopWorkflow.AddStep(vnext.WorkflowStep{Name: "write", Agent: agent1})
	loopWorkflow.AddStep(vnext.WorkflowStep{Name: "edit", Agent: agent2})

	// Wrap as agent
	loopAgent, err := vnext.NewBuilder("revision-agent").
		WithSubWorkflow(
			vnext.WithWorkflowInstance(loopWorkflow),
			vnext.WithSubWorkflowMaxDepthBuilder(10),
			vnext.WithSubWorkflowDescriptionBuilder("Writer-Editor revision loop"),
		).
		Build()

	if err != nil {
		t.Fatalf("Failed to build loop SubWorkflow: %v", err)
	}

	// Execute
	result, err := loopAgent.Run(context.Background(), "Write a story")
	if err != nil {
		t.Errorf("Loop SubWorkflow execution failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}

	// Result should contain "APPROVED" from editor
	if !strings.Contains(result.Content, "APPROVED") {
		t.Errorf("Expected result to contain 'APPROVED', got: %s", result.Content)
	}
}

// TestBuilderSubWorkflowNesting tests nested SubWorkflows
func TestBuilderSubWorkflowNesting(t *testing.T) {
	// Level 1: Inner agents
	agent1 := &mockAgent{name: "analyzer", response: "Analysis complete"}
	agent2 := &mockAgent{name: "reviewer", response: "Review complete"}

	// Level 2: Inner workflow
	innerWorkflow, err := vnext.NewParallelWorkflow(&vnext.WorkflowConfig{})
	if err != nil {
		t.Fatalf("Failed to create inner workflow: %v", err)
	}
	innerWorkflow.AddStep(vnext.WorkflowStep{Name: "analyze", Agent: agent1})
	innerWorkflow.AddStep(vnext.WorkflowStep{Name: "review", Agent: agent2})

	// Level 3: Wrap as agent
	innerAgent, err := vnext.NewBuilder("inner-agent").
		WithSubWorkflow(vnext.WithWorkflowInstance(innerWorkflow)).
		Build()

	if err != nil {
		t.Fatalf("Failed to build inner SubWorkflow: %v", err)
	}

	// Level 4: Outer workflow
	agent3 := &mockAgent{name: "summarizer", response: "Summary complete"}
	outerWorkflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	if err != nil {
		t.Fatalf("Failed to create outer workflow: %v", err)
	}
	outerWorkflow.AddStep(vnext.WorkflowStep{Name: "inner", Agent: innerAgent})
	outerWorkflow.AddStep(vnext.WorkflowStep{Name: "summarize", Agent: agent3})

	// Level 5: Wrap as agent
	outerAgent, err := vnext.NewBuilder("outer-agent").
		WithSubWorkflow(
			vnext.WithWorkflowInstance(outerWorkflow),
			vnext.WithSubWorkflowMaxDepthBuilder(5),
		).
		Build()

	if err != nil {
		t.Fatalf("Failed to build outer SubWorkflow: %v", err)
	}

	// Execute nested workflow
	result, err := outerAgent.Run(context.Background(), "test input")
	if err != nil {
		t.Errorf("Nested SubWorkflow execution failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}
}

// TestBuilderSubWorkflowMinimal tests minimal SubWorkflow creation
func TestBuilderSubWorkflowMinimal(t *testing.T) {
	agent := &mockAgent{name: "test-agent"}
	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}
	workflow.AddStep(vnext.WorkflowStep{Name: "test", Agent: agent})

	// Minimal configuration - only workflow instance
	subAgent, err := vnext.NewBuilder("simple-sub").
		WithSubWorkflow(vnext.WithWorkflowInstance(workflow)).
		Build()

	if err != nil {
		t.Fatalf("Failed to build minimal SubWorkflow: %v", err)
	}

	if subAgent == nil {
		t.Fatal("SubWorkflow agent should not be nil")
	}

	result, err := subAgent.Run(context.Background(), "test")
	if err != nil {
		t.Errorf("Execution failed: %v", err)
	}

	if result == nil || !result.Success {
		t.Error("Expected successful execution")
	}
}

// TestBuilderSubWorkflowImmutability tests that builder is immutable after Build()
func TestBuilderSubWorkflowImmutability(t *testing.T) {
	agent := &mockAgent{name: "test-agent"}
	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}
	workflow.AddStep(vnext.WorkflowStep{Name: "test", Agent: agent})

	builder := vnext.NewBuilder("test-sub").
		WithSubWorkflow(vnext.WithWorkflowInstance(workflow))

	// First build
	_, err = builder.Build()
	if err != nil {
		t.Fatalf("First build failed: %v", err)
	}

	// Second build should fail (builder is frozen)
	_, err = builder.Build()
	if err == nil {
		t.Fatal("Expected error when building twice")
	}

	if !strings.Contains(err.Error(), "frozen") {
		t.Errorf("Error should mention 'frozen', got: %s", err.Error())
	}
}

// TestBuilderSubWorkflowClone tests cloning SubWorkflow builders
func TestBuilderSubWorkflowClone(t *testing.T) {
	agent := &mockAgent{name: "test-agent"}
	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}
	workflow.AddStep(vnext.WorkflowStep{Name: "test", Agent: agent})

	// Create base builder
	baseBuilder := vnext.NewBuilder("base-sub").
		WithSubWorkflow(
			vnext.WithWorkflowInstance(workflow),
			vnext.WithSubWorkflowMaxDepthBuilder(5),
		)

	// Clone and modify
	clone1 := baseBuilder.Clone()
	agent1, err := clone1.Build()
	if err != nil {
		t.Fatalf("Clone1 build failed: %v", err)
	}

	// Original can still be cloned and built
	clone2 := baseBuilder.Clone()
	agent2, err := clone2.Build()
	if err != nil {
		t.Fatalf("Clone2 build failed: %v", err)
	}

	// Both agents should work
	if agent1 == nil || agent2 == nil {
		t.Fatal("Cloned agents should not be nil")
	}

	// Test both
	result1, _ := agent1.Run(context.Background(), "test")
	result2, _ := agent2.Run(context.Background(), "test")

	if result1 == nil || result2 == nil {
		t.Fatal("Results should not be nil")
	}
}

// mockAgent for testing
type mockAgent struct {
	name     string
	response string
}

func (m *mockAgent) Name() string {
	return m.name
}

func (m *mockAgent) Run(ctx context.Context, input string) (*vnext.Result, error) {
	response := m.response
	if response == "" {
		response = "Mock response from " + m.name
	}
	return &vnext.Result{
		Success: true,
		Content: response,
	}, nil
}

func (m *mockAgent) RunWithOptions(ctx context.Context, input string, opts *vnext.RunOptions) (*vnext.Result, error) {
	return m.Run(ctx, input)
}

func (m *mockAgent) RunStream(ctx context.Context, input string, opts ...vnext.StreamOption) (vnext.Stream, error) {
	return nil, nil
}

func (m *mockAgent) RunStreamWithOptions(ctx context.Context, input string, runOpts *vnext.RunOptions, streamOpts ...vnext.StreamOption) (vnext.Stream, error) {
	return nil, nil
}

func (m *mockAgent) Config() *vnext.Config {
	return &vnext.Config{Name: m.name}
}

func (m *mockAgent) Capabilities() []string {
	return []string{"mock"}
}

func (m *mockAgent) Initialize(ctx context.Context) error {
	return nil
}

func (m *mockAgent) Cleanup(ctx context.Context) error {
	return nil
}
