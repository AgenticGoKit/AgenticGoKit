package workflow_test

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// mockAgent implements a simple agent for workflow testing
type mockAgent struct {
	name         string
	responseFunc func(input string) string
	callCount    int32
	shouldFail   bool
	executeDelay time.Duration
}

func newMockAgent(name string, responseFunc func(input string) string) *mockAgent {
	return &mockAgent{
		name:         name,
		responseFunc: responseFunc,
	}
}

func (m *mockAgent) Name() string { return m.name }

func (m *mockAgent) Run(ctx context.Context, input string) (*vnext.Result, error) {
	atomic.AddInt32(&m.callCount, 1)

	if m.executeDelay > 0 {
		select {
		case <-time.After(m.executeDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.shouldFail {
		return &vnext.Result{
			Success: false,
			Content: "",
			Error:   "mock agent error",
		}, fmt.Errorf("mock agent error")
	}

	response := input
	if m.responseFunc != nil {
		response = m.responseFunc(input)
	}

	return &vnext.Result{
		Success:  true,
		Content:  response,
		Duration: 10 * time.Millisecond,
	}, nil
}

func (m *mockAgent) RunWithOptions(ctx context.Context, input string, opts *vnext.RunOptions) (*vnext.Result, error) {
	return m.Run(ctx, input)
}

func (m *mockAgent) RunStream(ctx context.Context, input string, opts ...vnext.StreamOption) (vnext.Stream, error) {
	return nil, fmt.Errorf("streaming not implemented in mock")
}

func (m *mockAgent) RunStreamWithOptions(ctx context.Context, input string, runOpts *vnext.RunOptions, streamOpts ...vnext.StreamOption) (vnext.Stream, error) {
	return nil, fmt.Errorf("streaming not implemented in mock")
}

func (m *mockAgent) Config() *vnext.Config {
	return &vnext.Config{Name: m.name}
}

func (m *mockAgent) Capabilities() []string {
	return []string{"test"}
}

func (m *mockAgent) Initialize(ctx context.Context) error {
	return nil
}

func (m *mockAgent) Cleanup(ctx context.Context) error {
	return nil
}

func (m *mockAgent) GetCallCount() int {
	return int(atomic.LoadInt32(&m.callCount))
}

// TestWorkflowModeConstants tests workflow mode constants
func TestWorkflowModeConstants(t *testing.T) {
	modes := []vnext.WorkflowMode{
		vnext.Sequential,
		vnext.Parallel,
		vnext.DAG,
		vnext.Loop,
	}

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			if mode == "" {
				t.Errorf("Workflow mode is empty")
			}
		})
	}
}

// TestSequentialWorkflowCreation tests sequential workflow creation
func TestSequentialWorkflowCreation(t *testing.T) {
	tests := []struct {
		name    string
		config  *vnext.WorkflowConfig
		wantErr bool
	}{
		{
			name:    "with_nil_config",
			config:  nil,
			wantErr: false, // Should use defaults
		},
		{
			name: "with_custom_config",
			config: &vnext.WorkflowConfig{
				Mode:    vnext.Sequential,
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := vnext.NewSequentialWorkflow(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSequentialWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && wf == nil {
				t.Error("NewSequentialWorkflow() returned nil workflow")
			}
		})
	}
}

// TestParallelWorkflowCreation tests parallel workflow creation
func TestParallelWorkflowCreation(t *testing.T) {
	tests := []struct {
		name    string
		config  *vnext.WorkflowConfig
		wantErr bool
	}{
		{
			name:    "with_nil_config",
			config:  nil,
			wantErr: false,
		},
		{
			name: "with_custom_config",
			config: &vnext.WorkflowConfig{
				Mode:    vnext.Parallel,
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := vnext.NewParallelWorkflow(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewParallelWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && wf == nil {
				t.Error("NewParallelWorkflow() returned nil workflow")
			}
		})
	}
}

// TestDAGWorkflowCreation tests DAG workflow creation
func TestDAGWorkflowCreation(t *testing.T) {
	tests := []struct {
		name    string
		config  *vnext.WorkflowConfig
		wantErr bool
	}{
		{
			name:    "with_nil_config",
			config:  nil,
			wantErr: false,
		},
		{
			name: "with_custom_config",
			config: &vnext.WorkflowConfig{
				Mode:    vnext.DAG,
				Timeout: 60 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := vnext.NewDAGWorkflow(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDAGWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && wf == nil {
				t.Error("NewDAGWorkflow() returned nil workflow")
			}
		})
	}
}

// TestLoopWorkflowCreation tests loop workflow creation
func TestLoopWorkflowCreation(t *testing.T) {
	tests := []struct {
		name    string
		config  *vnext.WorkflowConfig
		wantErr bool
	}{
		{
			name:    "with_nil_config",
			config:  nil,
			wantErr: false,
		},
		{
			name: "with_custom_config",
			config: &vnext.WorkflowConfig{
				Mode:          vnext.Loop,
				Timeout:       60 * time.Second,
				MaxIterations: 5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := vnext.NewLoopWorkflow(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLoopWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && wf == nil {
				t.Error("NewLoopWorkflow() returned nil workflow")
			}
		})
	}
}

// TestWorkflowAddStep tests adding steps to a workflow
func TestWorkflowAddStep(t *testing.T) {
	wf, err := vnext.NewSequentialWorkflow(nil)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	agent1 := newMockAgent("agent1", func(input string) string {
		return "response1"
	})

	agent2 := newMockAgent("agent2", func(input string) string {
		return "response2"
	})

	tests := []struct {
		name    string
		step    vnext.WorkflowStep
		wantErr bool
	}{
		{
			name: "valid_step",
			step: vnext.WorkflowStep{
				Name:  "step1",
				Agent: agent1,
			},
			wantErr: false,
		},
		{
			name: "step_with_metadata",
			step: vnext.WorkflowStep{
				Name:  "step2",
				Agent: agent2,
				Metadata: map[string]interface{}{
					"key": "value",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := wf.AddStep(tt.step)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddStep() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSequentialWorkflowExecution tests sequential workflow execution
func TestSequentialWorkflowExecution(t *testing.T) {
	wf, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	agent1 := newMockAgent("agent1", func(input string) string {
		return input + " -> step1"
	})

	agent2 := newMockAgent("agent2", func(input string) string {
		return input + " -> step2"
	})

	agent3 := newMockAgent("agent3", func(input string) string {
		return input + " -> step3"
	})

	wf.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
	wf.AddStep(vnext.WorkflowStep{Name: "step2", Agent: agent2})
	wf.AddStep(vnext.WorkflowStep{Name: "step3", Agent: agent3})

	ctx := context.Background()
	result, err := wf.Run(ctx, "input")

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful result, got error: %s", result.Error)
	}

	if len(result.StepResults) != 3 {
		t.Errorf("Expected 3 step results, got %d", len(result.StepResults))
	}

	// Verify execution order
	expectedOrder := []string{"step1", "step2", "step3"}
	if len(result.ExecutionPath) != len(expectedOrder) {
		t.Errorf("Execution path length = %d, want %d", len(result.ExecutionPath), len(expectedOrder))
	}

	for i, step := range expectedOrder {
		if i < len(result.ExecutionPath) && result.ExecutionPath[i] != step {
			t.Errorf("Execution path[%d] = %s, want %s", i, result.ExecutionPath[i], step)
		}
	}

	// Verify each agent was called once
	if agent1.GetCallCount() != 1 {
		t.Errorf("Agent1 call count = %d, want 1", agent1.GetCallCount())
	}
	if agent2.GetCallCount() != 1 {
		t.Errorf("Agent2 call count = %d, want 1", agent2.GetCallCount())
	}
	if agent3.GetCallCount() != 1 {
		t.Errorf("Agent3 call count = %d, want 1", agent3.GetCallCount())
	}
}

// TestParallelWorkflowExecution tests parallel workflow execution
func TestParallelWorkflowExecution(t *testing.T) {
	wf, err := vnext.NewParallelWorkflow(&vnext.WorkflowConfig{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	agent1 := newMockAgent("agent1", func(input string) string {
		time.Sleep(50 * time.Millisecond)
		return "response1"
	})

	agent2 := newMockAgent("agent2", func(input string) string {
		time.Sleep(50 * time.Millisecond)
		return "response2"
	})

	agent3 := newMockAgent("agent3", func(input string) string {
		time.Sleep(50 * time.Millisecond)
		return "response3"
	})

	wf.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
	wf.AddStep(vnext.WorkflowStep{Name: "step2", Agent: agent2})
	wf.AddStep(vnext.WorkflowStep{Name: "step3", Agent: agent3})

	ctx := context.Background()
	startTime := time.Now()
	result, err := wf.Run(ctx, "input")
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful result, got error: %s", result.Error)
	}

	if len(result.StepResults) != 3 {
		t.Errorf("Expected 3 step results, got %d", len(result.StepResults))
	}

	// Parallel execution should be faster than sequential
	// Sequential would take ~150ms, parallel should take ~50-100ms
	if duration > 120*time.Millisecond {
		t.Errorf("Parallel execution took too long: %v (expected < 120ms)", duration)
	}

	// Verify each agent was called once
	if agent1.GetCallCount() != 1 {
		t.Errorf("Agent1 call count = %d, want 1", agent1.GetCallCount())
	}
	if agent2.GetCallCount() != 1 {
		t.Errorf("Agent2 call count = %d, want 1", agent2.GetCallCount())
	}
	if agent3.GetCallCount() != 1 {
		t.Errorf("Agent3 call count = %d, want 1", agent3.GetCallCount())
	}
}

// TestDAGWorkflowExecution tests DAG workflow with dependencies
func TestDAGWorkflowExecution(t *testing.T) {
	wf, err := vnext.NewDAGWorkflow(&vnext.WorkflowConfig{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	agent1 := newMockAgent("agent1", func(input string) string {
		return "result1"
	})

	agent2 := newMockAgent("agent2", func(input string) string {
		return "result2"
	})

	agent3 := newMockAgent("agent3", func(input string) string {
		return "result3"
	})

	// DAG structure: step1 and step2 run first, step3 depends on both
	wf.AddStep(vnext.WorkflowStep{
		Name:  "step1",
		Agent: agent1,
	})

	wf.AddStep(vnext.WorkflowStep{
		Name:  "step2",
		Agent: agent2,
	})

	wf.AddStep(vnext.WorkflowStep{
		Name:         "step3",
		Agent:        agent3,
		Dependencies: []string{"step1", "step2"},
	})

	ctx := context.Background()
	result, err := wf.Run(ctx, "input")

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful result, got error: %s", result.Error)
	}

	if len(result.StepResults) != 3 {
		t.Errorf("Expected 3 step results, got %d", len(result.StepResults))
	}

	// Verify step3 executed after step1 and step2
	step3Index := -1
	step1Index := -1
	step2Index := -1

	for i, step := range result.ExecutionPath {
		switch step {
		case "step1":
			step1Index = i
		case "step2":
			step2Index = i
		case "step3":
			step3Index = i
		}
	}

	if step3Index <= step1Index || step3Index <= step2Index {
		t.Error("step3 should execute after step1 and step2")
	}
}

// TestLoopWorkflowExecution tests loop workflow execution
func TestLoopWorkflowExecution(t *testing.T) {
	wf, err := vnext.NewLoopWorkflow(&vnext.WorkflowConfig{
		Timeout:       10 * time.Second,
		MaxIterations: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	iterationCount := 0
	agent := newMockAgent("agent", func(input string) string {
		iterationCount++
		return fmt.Sprintf("iteration_%d", iterationCount)
	})

	wf.AddStep(vnext.WorkflowStep{
		Name:  "loop_step",
		Agent: agent,
		Condition: func(ctx context.Context, wc *vnext.WorkflowContext) bool {
			// Stop after 3 iterations
			return wc.IterationNum < 3
		},
	})

	ctx := context.Background()
	result, err := wf.Run(ctx, "input")

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful result, got error: %s", result.Error)
	}

	// Should have executed 3 times
	if agent.GetCallCount() != 3 {
		t.Errorf("Agent call count = %d, want 3", agent.GetCallCount())
	}
}

// TestWorkflowStepCondition tests conditional step execution
func TestWorkflowStepCondition(t *testing.T) {
	wf, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	agent1 := newMockAgent("agent1", func(input string) string {
		return "result1"
	})

	agent2 := newMockAgent("agent2", func(input string) string {
		return "result2"
	})

	agent3 := newMockAgent("agent3", func(input string) string {
		return "result3"
	})

	wf.AddStep(vnext.WorkflowStep{
		Name:  "step1",
		Agent: agent1,
	})

	// This step should be skipped due to condition
	wf.AddStep(vnext.WorkflowStep{
		Name:  "step2",
		Agent: agent2,
		Condition: func(ctx context.Context, wc *vnext.WorkflowContext) bool {
			return false // Always skip
		},
	})

	wf.AddStep(vnext.WorkflowStep{
		Name:  "step3",
		Agent: agent3,
	})

	ctx := context.Background()
	result, err := wf.Run(ctx, "input")

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// agent1 and agent3 should execute, agent2 should be skipped
	if agent1.GetCallCount() != 1 {
		t.Errorf("Agent1 call count = %d, want 1", agent1.GetCallCount())
	}
	if agent2.GetCallCount() != 0 {
		t.Errorf("Agent2 call count = %d, want 0 (should be skipped)", agent2.GetCallCount())
	}
	if agent3.GetCallCount() != 1 {
		t.Errorf("Agent3 call count = %d, want 1", agent3.GetCallCount())
	}

	// Check for skipped step in results
	foundSkipped := false
	for _, stepResult := range result.StepResults {
		if stepResult.StepName == "step2" && stepResult.Skipped {
			foundSkipped = true
			break
		}
	}

	if !foundSkipped {
		t.Error("Expected to find skipped step2 in results")
	}
}

// TestWorkflowStepTransform tests input transformation
func TestWorkflowStepTransform(t *testing.T) {
	wf, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	agent := newMockAgent("agent", func(input string) string {
		return input
	})

	wf.AddStep(vnext.WorkflowStep{
		Name:  "transform_step",
		Agent: agent,
		Transform: func(input string) string {
			return strings.ToUpper(input)
		},
	})

	ctx := context.Background()
	result, err := wf.Run(ctx, "lowercase input")

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful result")
	}

	// Check if the output contains uppercase text (transform was applied)
	if len(result.StepResults) > 0 {
		output := result.StepResults[0].Output
		if !strings.Contains(output, "LOWERCASE") && !strings.Contains(output, "INPUT") {
			t.Logf("Transform may not have been applied, output: %s", output)
		}
	}
}

// TestWorkflowContext tests workflow context sharing
func TestWorkflowContext(t *testing.T) {
	wf, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	agent1 := newMockAgent("agent1", func(input string) string {
		return "data_from_step1"
	})

	agent2 := newMockAgent("agent2", func(input string) string {
		return "data_from_step2"
	})

	wf.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
	wf.AddStep(vnext.WorkflowStep{Name: "step2", Agent: agent2})

	ctx := context.Background()
	result, err := wf.Run(ctx, "input")

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful result")
	}

	// Verify both steps executed
	if len(result.StepResults) != 2 {
		t.Errorf("Expected 2 step results, got %d", len(result.StepResults))
	}
}

// TestWorkflowErrorHandling tests error handling in workflows
func TestWorkflowErrorHandling(t *testing.T) {
	wf, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	agent1 := newMockAgent("agent1", func(input string) string {
		return "result1"
	})

	failingAgent := newMockAgent("failing", func(input string) string {
		return ""
	})
	failingAgent.shouldFail = true

	agent3 := newMockAgent("agent3", func(input string) string {
		return "result3"
	})

	wf.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
	wf.AddStep(vnext.WorkflowStep{Name: "failing_step", Agent: failingAgent})
	wf.AddStep(vnext.WorkflowStep{Name: "step3", Agent: agent3})

	ctx := context.Background()
	result, err := wf.Run(ctx, "input")

	// Workflow should report error
	if err == nil && result.Success {
		t.Error("Expected workflow to fail when step fails")
	}

	// step1 should have executed
	if agent1.GetCallCount() != 1 {
		t.Errorf("Agent1 call count = %d, want 1", agent1.GetCallCount())
	}

	// failingAgent should have executed
	if failingAgent.GetCallCount() != 1 {
		t.Errorf("Failing agent call count = %d, want 1", failingAgent.GetCallCount())
	}

	// step3 might not execute depending on error handling strategy
	t.Logf("Agent3 call count: %d", agent3.GetCallCount())
}

// TestWorkflowTimeout tests workflow timeout handling
func TestWorkflowTimeout(t *testing.T) {
	wf, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Timeout: 100 * time.Millisecond, // Very short timeout
	})
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	slowAgent := newMockAgent("slow", func(input string) string {
		time.Sleep(200 * time.Millisecond) // Longer than timeout
		return "result"
	})
	slowAgent.executeDelay = 200 * time.Millisecond

	wf.AddStep(vnext.WorkflowStep{Name: "slow_step", Agent: slowAgent})

	ctx := context.Background()
	result, err := wf.Run(ctx, "input")

	// Should timeout
	if err == nil && result.Success {
		t.Error("Expected workflow to timeout")
	}

	if result != nil && !strings.Contains(result.Error, "timeout") && !strings.Contains(result.Error, "context") {
		t.Logf("Expected timeout error, got: %s", result.Error)
	}
}

// TestWorkflowGetConfig tests workflow configuration retrieval
func TestWorkflowGetConfig(t *testing.T) {
	config := &vnext.WorkflowConfig{
		Mode:          vnext.Sequential,
		Timeout:       30 * time.Second,
		MaxIterations: 5,
	}

	wf, err := vnext.NewSequentialWorkflow(config)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	retrievedConfig := wf.GetConfig()
	if retrievedConfig == nil {
		t.Fatal("GetConfig() returned nil")
	}

	if retrievedConfig.Mode != vnext.Sequential {
		t.Errorf("Mode = %s, want %s", retrievedConfig.Mode, vnext.Sequential)
	}

	if retrievedConfig.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", retrievedConfig.Timeout)
	}
}

// TestWorkflowInitializeShutdown tests workflow lifecycle
func TestWorkflowInitializeShutdown(t *testing.T) {
	wf, err := vnext.NewSequentialWorkflow(nil)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	ctx := context.Background()

	// Test Initialize
	err = wf.Initialize(ctx)
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	// Test Shutdown
	err = wf.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

// TestWorkflowResult tests WorkflowResult structure
func TestWorkflowResult(t *testing.T) {
	result := &vnext.WorkflowResult{
		Success:     true,
		FinalOutput: "final result",
		StepResults: []vnext.StepResult{
			{
				StepName: "step1",
				Success:  true,
				Output:   "output1",
				Duration: 10 * time.Millisecond,
			},
		},
		Duration:      50 * time.Millisecond,
		TotalTokens:   100,
		ExecutionPath: []string{"step1"},
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	if !result.Success {
		t.Error("Result should be successful")
	}

	if result.FinalOutput != "final result" {
		t.Errorf("FinalOutput = %s, want 'final result'", result.FinalOutput)
	}

	if len(result.StepResults) != 1 {
		t.Errorf("StepResults length = %d, want 1", len(result.StepResults))
	}

	if len(result.ExecutionPath) != 1 {
		t.Errorf("ExecutionPath length = %d, want 1", len(result.ExecutionPath))
	}
}

// TestStepResult tests StepResult structure
func TestStepResult(t *testing.T) {
	stepResult := vnext.StepResult{
		StepName:  "test_step",
		Success:   true,
		Output:    "test output",
		Duration:  15 * time.Millisecond,
		Tokens:    50,
		Skipped:   false,
		Timestamp: time.Now(),
	}

	if stepResult.StepName != "test_step" {
		t.Errorf("StepName = %s, want 'test_step'", stepResult.StepName)
	}

	if !stepResult.Success {
		t.Error("Success should be true")
	}

	if stepResult.Skipped {
		t.Error("Skipped should be false")
	}
}

// TestWorkflowConfig tests WorkflowConfig structure
func TestWorkflowConfig(t *testing.T) {
	config := &vnext.WorkflowConfig{
		Mode:          vnext.DAG,
		Agents:        []string{"agent1", "agent2"},
		Timeout:       60 * time.Second,
		MaxIterations: 10,
	}

	if config.Mode != vnext.DAG {
		t.Errorf("Mode = %s, want %s", config.Mode, vnext.DAG)
	}

	if len(config.Agents) != 2 {
		t.Errorf("Agents length = %d, want 2", len(config.Agents))
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", config.Timeout)
	}

	if config.MaxIterations != 10 {
		t.Errorf("MaxIterations = %d, want 10", config.MaxIterations)
	}
}



