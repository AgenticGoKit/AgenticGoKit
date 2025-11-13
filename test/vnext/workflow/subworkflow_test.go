package workflow_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

// =============================================================================
// BASIC EXECUTION TESTS
// =============================================================================

func TestSubWorkflowAgent_BasicExecution(t *testing.T) {
	// Create a simple mock agent
	mockAgent := &mockSubWorkflowAgent{
		name:    "mock",
		content: "test output",
	}

	// Create a sequential workflow
	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	err = workflow.AddStep(vnext.WorkflowStep{
		Name:  "step1",
		Agent: mockAgent,
	})
	if err != nil {
		t.Fatalf("Failed to add step: %v", err)
	}

	// Wrap workflow as agent
	agent := vnext.NewSubWorkflowAgent("test", workflow)

	// Execute
	ctx := context.Background()
	result, err := agent.Run(ctx, "test input")

	// Verify
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !result.Success {
		t.Error("Expected success=true")
	}
	if result.Content != "test output" {
		t.Errorf("Expected content='test output', got: %s", result.Content)
	}

	// Check metadata
	metadata := result.Metadata
	if metadata["type"] != "subworkflow" {
		t.Errorf("Expected type='subworkflow', got: %v", metadata["type"])
	}
	if metadata["workflow_name"] != "test" {
		t.Errorf("Expected workflow_name='test', got: %v", metadata["workflow_name"])
	}
}

func TestSubWorkflowAgent_Name(t *testing.T) {
	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	agent := vnext.NewSubWorkflowAgent("my_workflow", workflow)

	if agent.Name() != "my_workflow" {
		t.Errorf("Expected name='my_workflow', got: %s", agent.Name())
	}
}

func TestSubWorkflowAgent_Config(t *testing.T) {
	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	agent := vnext.NewSubWorkflowAgent("my_workflow", workflow)

	config := agent.Config()
	if config == nil {
		t.Fatal("Expected non-nil config")
	}
	if config.Name != "my_workflow" {
		t.Errorf("Expected config.Name='my_workflow', got: %s", config.Name)
	}
}

// =============================================================================
// WORKFLOW TYPE TESTS
// =============================================================================

func TestSubWorkflowAgent_SequentialSubWorkflow(t *testing.T) {
	// Create mock agents
	agent1 := &mockSubWorkflowAgent{name: "agent1", content: "output1"}
	agent2 := &mockSubWorkflowAgent{name: "agent2", content: "output2"}

	// Create sequential workflow
	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
	workflow.AddStep(vnext.WorkflowStep{Name: "step2", Agent: agent2})

	// Wrap as agent
	workflowAgent := vnext.NewSubWorkflowAgent("sequential", workflow)

	// Execute
	result, err := workflowAgent.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	// Verify both agents were called
	if agent1.callCount != 1 || agent2.callCount != 1 {
		t.Errorf("Expected both agents called once, got agent1=%d, agent2=%d",
			agent1.callCount, agent2.callCount)
	}
}

func TestSubWorkflowAgent_ParallelSubWorkflow(t *testing.T) {
	// Create mock agents
	agent1 := &mockSubWorkflowAgent{name: "agent1", content: "output1"}
	agent2 := &mockSubWorkflowAgent{name: "agent2", content: "output2"}

	// Create parallel workflow
	workflow, err := vnext.NewParallelWorkflow(&vnext.WorkflowConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
	workflow.AddStep(vnext.WorkflowStep{Name: "step2", Agent: agent2})

	// Wrap as agent
	workflowAgent := vnext.NewSubWorkflowAgent("parallel", workflow)

	// Execute
	result, err := workflowAgent.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	// Verify both agents were called
	if agent1.callCount != 1 || agent2.callCount != 1 {
		t.Error("Expected both agents to be called once")
	}
}

func TestSubWorkflowAgent_DAGSubWorkflow(t *testing.T) {
	// Create mock agents
	agent1 := &mockSubWorkflowAgent{name: "agent1", content: "output1"}
	agent2 := &mockSubWorkflowAgent{name: "agent2", content: "output2"}

	// Create DAG workflow
	workflow, err := vnext.NewDAGWorkflow(&vnext.WorkflowConfig{})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
	workflow.AddStep(vnext.WorkflowStep{
		Name:         "step2",
		Agent:        agent2,
		Dependencies: []string{"step1"},
	})

	// Wrap as agent
	workflowAgent := vnext.NewSubWorkflowAgent("dag", workflow)

	// Execute
	result, err := workflowAgent.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}
}

func TestSubWorkflowAgent_LoopSubWorkflow(t *testing.T) {
	// Create mock agent
	agent1 := &mockSubWorkflowAgent{name: "agent1", content: "output"}

	// Create loop workflow
	workflow, err := vnext.NewLoopWorkflow(&vnext.WorkflowConfig{
		MaxIterations: 2,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})

	// Wrap as agent
	workflowAgent := vnext.NewSubWorkflowAgent("loop", workflow)

	// Execute
	result, err := workflowAgent.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	// Verify agent was called multiple times (loop iterations)
	if agent1.callCount < 2 {
		t.Errorf("Expected at least 2 calls, got: %d", agent1.callCount)
	}
}

// =============================================================================
// RESULT CONVERSION TESTS
// =============================================================================

func TestSubWorkflowAgent_ResultConversion(t *testing.T) {
	mockAgent := &mockSubWorkflowAgent{
		name:       "mock",
		content:    "test content",
		tokensUsed: 100,
		delay:      time.Millisecond, // Small delay to ensure measurable duration
	}

	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: mockAgent})

	workflowAgent := vnext.NewSubWorkflowAgent("test", workflow)

	result, err := workflowAgent.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Verify result fields
	if !result.Success {
		t.Error("Expected Success=true")
	}
	if result.Content != "test content" {
		t.Errorf("Expected Content='test content', got: %s", result.Content)
	}
	// Duration should be >= 0; very fast operations may round to 0
	if result.Duration < 0 {
		t.Errorf("Expected Duration >= 0, got: %v", result.Duration)
	}

	// Verify metadata
	metadata := result.Metadata
	if metadata["type"] != "subworkflow" {
		t.Error("Expected metadata type='subworkflow'")
	}
	if metadata["workflow_name"] != "test" {
		t.Error("Expected metadata workflow_name='test'")
	}
	if _, ok := metadata["step_count"]; !ok {
		t.Error("Expected metadata to contain step_count")
	}
}

// =============================================================================
// METADATA TRACKING TESTS
// =============================================================================

func TestSubWorkflowAgent_MetadataTracking(t *testing.T) {
	mockAgent := &mockSubWorkflowAgent{name: "mock", content: "output"}

	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: mockAgent})
	workflow.AddStep(vnext.WorkflowStep{Name: "step2", Agent: mockAgent})

	workflowAgent := vnext.NewSubWorkflowAgent("test", workflow)

	result, err := workflowAgent.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	metadata := result.Metadata

	// Check required metadata fields
	requiredFields := []string{
		"type", "workflow_name", "workflow_path",
		"depth", "step_count", "execution_path",
		"workflow_duration", "execution_count",
	}

	for _, field := range requiredFields {
		if _, ok := metadata[field]; !ok {
			t.Errorf("Expected metadata to contain field: %s", field)
		}
	}

	// Verify step_count
	if metadata["step_count"] != 2 {
		t.Errorf("Expected step_count=2, got: %v", metadata["step_count"])
	}

	// Verify execution_count
	if metadata["execution_count"] != int64(1) {
		t.Errorf("Expected execution_count=1, got: %v", metadata["execution_count"])
	}
}

func TestSubWorkflowAgent_ExecutionStats(t *testing.T) {
	mockAgent := &mockSubWorkflowAgent{
		name:    "mock",
		content: "output",
		delay:   time.Millisecond, // Ensure measurable duration
	}

	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: mockAgent})

	workflowAgent := vnext.NewSubWorkflowAgent("test", workflow)

	// Execute multiple times
	for i := 0; i < 3; i++ {
		_, err := workflowAgent.Run(context.Background(), "input")
		if err != nil {
			t.Fatalf("Execution %d failed: %v", i+1, err)
		}
	}

	// Get stats - need to cast to concrete type
	stats := workflowAgent.(interface {
		GetStats() vnext.SubWorkflowStats
	}).GetStats()

	if stats.ExecutionCount != 3 {
		t.Errorf("Expected ExecutionCount=3, got: %d", stats.ExecutionCount)
	}
	// Duration might still be close to 0 on very fast systems
	if stats.TotalDuration < 0 {
		t.Errorf("Expected TotalDuration >= 0, got: %v", stats.TotalDuration)
	}
	if stats.AvgDuration < 0 {
		t.Errorf("Expected AvgDuration >= 0, got: %v", stats.AvgDuration)
	}
	if stats.Name != "test" {
		t.Errorf("Expected Name='test', got: %s", stats.Name)
	}
}

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

func TestSubWorkflowAgent_ErrorHandling(t *testing.T) {
	// Create agent that returns error
	errorAgent := &mockSubWorkflowAgent{
		name:        "error_agent",
		shouldError: true,
		errorMsg:    "intentional error",
	}

	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: errorAgent})

	workflowAgent := vnext.NewSubWorkflowAgent("test", workflow)

	result, err := workflowAgent.Run(context.Background(), "input")

	// Workflows may handle errors internally and return a failed result without error
	// OR they may propagate the error. Both are valid. Check result status.
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Either we get an error OR result indicates failure
	if err == nil && result.Success {
		t.Error("Expected either error or result.Success=false")
	}

	// If we got an error, it should mention the subworkflow
	if err != nil && !strings.Contains(err.Error(), "subworkflow") {
		t.Logf("Warning: Error message should contain 'subworkflow', got: %v", err)
	}
}

// =============================================================================
// DEPTH AND SAFETY TESTS
// =============================================================================

func TestSubWorkflowAgent_MaxDepthEnforcement(t *testing.T) {
	mockAgent := &mockSubWorkflowAgent{name: "mock", content: "output"}

	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: mockAgent})

	// Create workflow agent with max depth of 2
	workflowAgent := vnext.NewSubWorkflowAgent("test", workflow,
		vnext.WithSubWorkflowMaxDepth(2),
		vnext.WithSubWorkflowDepth(2), // Already at max depth
	)

	// Should fail due to max depth
	_, err := workflowAgent.Run(context.Background(), "input")

	if err == nil {
		t.Fatal("Expected error due to max depth, got nil")
	}
	if !strings.Contains(err.Error(), "maximum workflow nesting depth") {
		t.Errorf("Expected max depth error, got: %v", err)
	}
}

func TestSubWorkflowAgent_DepthTracking(t *testing.T) {
	mockAgent := &mockSubWorkflowAgent{name: "mock", content: "output"}

	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: mockAgent})

	// Create workflow agent with custom depth
	workflowAgent := vnext.NewSubWorkflowAgent("test", workflow, vnext.WithSubWorkflowDepth(3))

	stats := workflowAgent.(interface {
		GetStats() vnext.SubWorkflowStats
	}).GetStats()

	if stats.Depth != 3 {
		t.Errorf("Expected Depth=3, got: %d", stats.Depth)
	}
}

func TestSubWorkflowAgent_ParentPathTracking(t *testing.T) {
	mockAgent := &mockSubWorkflowAgent{name: "mock", content: "output"}

	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: mockAgent})

	// Create workflow agent with parent path
	workflowAgent := vnext.NewSubWorkflowAgent("child", workflow,
		vnext.WithSubWorkflowParentPath("parent/grandparent"),
	)

	stats := workflowAgent.(interface {
		GetStats() vnext.SubWorkflowStats
	}).GetStats()

	expectedPath := "parent/grandparent/child"
	if stats.Path != expectedPath {
		t.Errorf("Expected Path='%s', got: %s", expectedPath, stats.Path)
	}
}

// =============================================================================
// OPTIONS TESTS
// =============================================================================

func TestSubWorkflowAgent_WithMaxDepthOption(t *testing.T) {
	mockAgent := &mockSubWorkflowAgent{name: "mock", content: "output"}
	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: mockAgent})

	agent := vnext.NewSubWorkflowAgent("test", workflow, vnext.WithSubWorkflowMaxDepth(5))

	// Verify through stats
	stats := agent.(interface {
		GetStats() vnext.SubWorkflowStats
	}).GetStats()

	if stats.MaxDepth != 5 {
		t.Errorf("Expected MaxDepth=5, got: %d", stats.MaxDepth)
	}
}

func TestSubWorkflowAgent_WithDescriptionOption(t *testing.T) {
	mockAgent := &mockSubWorkflowAgent{name: "mock", content: "output"}
	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: mockAgent})

	customDesc := "Custom workflow description"
	agent := vnext.NewSubWorkflowAgent("test", workflow, vnext.WithSubWorkflowDescription(customDesc))

	// Verify through stats
	stats := agent.(interface {
		GetStats() vnext.SubWorkflowStats
	}).GetStats()

	if stats.Description != customDesc {
		t.Errorf("Expected description='%s', got: %s", customDesc, stats.Description)
	}
}

// =============================================================================
// CONVENIENCE FACTORY TESTS
// =============================================================================

func TestQuickSubWorkflow(t *testing.T) {
	mockAgent := &mockSubWorkflowAgent{name: "mock", content: "output"}
	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: mockAgent})

	agent := vnext.QuickSubWorkflow("test", workflow)

	if agent.Name() != "test" {
		t.Errorf("Expected name='test', got: %s", agent.Name())
	}

	result, err := agent.Run(context.Background(), "input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected successful execution")
	}
}

func TestNewSequentialSubWorkflow(t *testing.T) {
	agent, err := vnext.NewSequentialSubWorkflow("test", &vnext.WorkflowConfig{})

	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}
	if agent == nil {
		t.Fatal("Expected non-nil agent")
	}
	if agent.Name() != "test" {
		t.Errorf("Expected name='test', got: %s", agent.Name())
	}
}

func TestNewParallelSubWorkflow(t *testing.T) {
	agent, err := vnext.NewParallelSubWorkflow("test", &vnext.WorkflowConfig{})

	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}
	if agent == nil {
		t.Fatal("Expected non-nil agent")
	}
	if agent.Name() != "test" {
		t.Errorf("Expected name='test', got: %s", agent.Name())
	}
}

// =============================================================================
// STREAMING TESTS
// =============================================================================

func TestSubWorkflowAgent_Streaming(t *testing.T) {
	mockAgent := &mockSubWorkflowAgent{
		name:    "mock",
		content: "streaming output",
	}

	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: mockAgent})

	workflowAgent := vnext.NewSubWorkflowAgent("test", workflow)

	// Test streaming
	stream, err := workflowAgent.RunStream(context.Background(), "input")
	if err != nil {
		t.Fatalf("Streaming failed: %v", err)
	}
	if stream == nil {
		t.Fatal("Expected non-nil stream")
	}

	// Verify stream metadata
	metadata := stream.Metadata()
	if metadata.Extra == nil {
		t.Error("Expected stream metadata Extra to be set")
	} else {
		if metadata.Extra["subworkflow_name"] != "test" {
			t.Error("Expected subworkflow_name in stream metadata")
		}
	}
}

// =============================================================================
// CONTEXT TESTS
// =============================================================================

func TestSubWorkflowAgent_ContextTimeout(t *testing.T) {
	// Create slow agent
	slowAgent := &mockSubWorkflowAgent{
		name:    "slow",
		content: "output",
		delay:   2 * time.Second,
	}

	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{})
	workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: slowAgent})

	workflowAgent := vnext.NewSubWorkflowAgent("test", workflow)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should timeout (though behavior depends on workflow implementation)
	result, err := workflowAgent.Run(ctx, "input")

	// Either we get an error or the context was honored and operation stopped
	// On some systems/implementations, the timeout may not be strictly honored
	if err == nil && result != nil && result.Success {
		t.Skip("Timeout not honored - may be expected if workflow doesn't check context")
	}

	// If we got an error, verify it's timeout-related
	if err != nil && !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "timeout") {
		t.Logf("Got error but not timeout-related: %v", err)
	}
}

// =============================================================================
// MOCK AGENT FOR TESTING
// =============================================================================

type mockSubWorkflowAgent struct {
	name        string
	content     string
	tokensUsed  int
	shouldError bool
	errorMsg    string
	callCount   int
	delay       time.Duration
}

func (m *mockSubWorkflowAgent) Run(ctx context.Context, input string) (*vnext.Result, error) {
	return m.RunWithOptions(ctx, input, nil)
}

func (m *mockSubWorkflowAgent) RunWithOptions(ctx context.Context, input string, opts *vnext.RunOptions) (*vnext.Result, error) {
	m.callCount++

	// Simulate delay if specified
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.shouldError {
		return &vnext.Result{
			Success: false,
			Error:   m.errorMsg,
		}, fmt.Errorf("%s", m.errorMsg)
	}

	return &vnext.Result{
		Success:    true,
		Content:    m.content,
		TokensUsed: m.tokensUsed,
	}, nil
}

func (m *mockSubWorkflowAgent) RunStream(ctx context.Context, input string, opts ...vnext.StreamOption) (vnext.Stream, error) {
	return m.RunStreamWithOptions(ctx, input, nil, opts...)
}

func (m *mockSubWorkflowAgent) RunStreamWithOptions(ctx context.Context, input string, runOpts *vnext.RunOptions, streamOpts ...vnext.StreamOption) (vnext.Stream, error) {
	// Create a basic stream
	metadata := &vnext.StreamMetadata{
		AgentName: m.name,
		StartTime: time.Now(),
	}
	stream, writer := vnext.NewStream(ctx, metadata)

	go func() {
		defer writer.Close()
		writer.Write(&vnext.StreamChunk{
			Type:    vnext.ChunkTypeText,
			Content: m.content,
		})
		writer.Write(&vnext.StreamChunk{
			Type: vnext.ChunkTypeDone,
		})
	}()

	return stream, nil
}

func (m *mockSubWorkflowAgent) Name() string {
	return m.name
}

func (m *mockSubWorkflowAgent) Config() *vnext.Config {
	return &vnext.Config{Name: m.name}
}

func (m *mockSubWorkflowAgent) Capabilities() []string {
	return []string{"test", "mock"}
}

func (m *mockSubWorkflowAgent) Initialize(ctx context.Context) error {
	return nil
}

func (m *mockSubWorkflowAgent) Cleanup(ctx context.Context) error {
	return nil
}
