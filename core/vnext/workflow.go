// Package vnext provides streamlined workflow orchestration for multi-agent systems.
// This file consolidates workflow functionality into a clean, easy-to-use interface.
package vnext

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// WORKFLOW INTERFACE
// =============================================================================

// Workflow defines the interface for multi-agent workflow orchestration
// This provides a simplified interface for sequential, parallel, DAG, and loop workflows
type Workflow interface {
	// Run executes the workflow with the given input
	Run(ctx context.Context, input string) (*WorkflowResult, error)

	// RunStream executes the workflow with streaming output
	RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error)

	// AddStep adds a step to the workflow
	// For sequential/loop workflows, steps are added in order
	// For parallel workflows, all steps run concurrently
	// For DAG workflows, steps should specify dependencies
	AddStep(step WorkflowStep) error

	// SetMemory configures shared memory for the workflow
	SetMemory(memory Memory)

	// GetConfig returns the workflow configuration
	GetConfig() *WorkflowConfig

	// Lifecycle methods
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Name         string                                       // Step identifier
	Agent        Agent                                        // Agent to execute this step
	Condition    func(context.Context, *WorkflowContext) bool // Optional condition function
	Dependencies []string                                     // Dependencies for DAG workflows
	Transform    func(string) string                          // Optional input transformation
	Metadata     map[string]interface{}                       // Additional step metadata
}

// WorkflowResult represents the result of workflow execution
type WorkflowResult struct {
	Success       bool                   `json:"success"`
	FinalOutput   string                 `json:"final_output"`
	StepResults   []StepResult           `json:"step_results"`
	Duration      time.Duration          `json:"duration"`
	TotalTokens   int                    `json:"total_tokens"`
	ExecutionPath []string               `json:"execution_path"` // Order of executed steps
	Metadata      map[string]interface{} `json:"metadata"`
	Error         string                 `json:"error,omitempty"`
}

// StepResult represents the result of a single workflow step
type StepResult struct {
	StepName  string        `json:"step_name"`
	Success   bool          `json:"success"`
	Output    string        `json:"output"`
	Duration  time.Duration `json:"duration"`
	Tokens    int           `json:"tokens"`
	Error     string        `json:"error,omitempty"`
	Skipped   bool          `json:"skipped,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// WorkflowContext provides shared context across workflow steps
type WorkflowContext struct {
	WorkflowID   string                 // Unique workflow execution ID
	SharedMemory Memory                 // Shared memory across steps
	StepResults  map[string]*StepResult // Results from completed steps
	Variables    map[string]interface{} // Shared variables
	CurrentStep  string                 // Currently executing step
	IterationNum int                    // For loop workflows
	mu           sync.RWMutex           // Protects concurrent access
}

// Get retrieves a variable from the workflow context
func (wc *WorkflowContext) Get(key string) (interface{}, bool) {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	val, ok := wc.Variables[key]
	return val, ok
}

// Set stores a variable in the workflow context
func (wc *WorkflowContext) Set(key string, value interface{}) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.Variables[key] = value
}

// GetStepResult retrieves the result of a completed step
func (wc *WorkflowContext) GetStepResult(stepName string) (*StepResult, bool) {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	result, ok := wc.StepResults[stepName]
	return result, ok
}

// =============================================================================
// FACTORY FUNCTIONS
// =============================================================================

// NewSequentialWorkflow creates a workflow that executes steps in order
func NewSequentialWorkflow(config *WorkflowConfig) (Workflow, error) {
	if config == nil {
		config = &WorkflowConfig{
			Mode:          Sequential,
			Timeout:       60 * time.Second,
			MaxIterations: 1,
		}
	}
	config.Mode = Sequential
	return newBasicWorkflow(config)
}

// NewParallelWorkflow creates a workflow that executes all steps concurrently
func NewParallelWorkflow(config *WorkflowConfig) (Workflow, error) {
	if config == nil {
		config = &WorkflowConfig{
			Mode:    Parallel,
			Timeout: 60 * time.Second,
		}
	}
	config.Mode = Parallel
	return newBasicWorkflow(config)
}

// NewDAGWorkflow creates a workflow that executes steps based on dependencies
func NewDAGWorkflow(config *WorkflowConfig) (Workflow, error) {
	if config == nil {
		config = &WorkflowConfig{
			Mode:    DAG,
			Timeout: 120 * time.Second,
		}
	}
	config.Mode = DAG
	return newBasicWorkflow(config)
}

// NewLoopWorkflow creates a workflow that repeats steps until a condition is met
func NewLoopWorkflow(config *WorkflowConfig) (Workflow, error) {
	if config == nil {
		config = &WorkflowConfig{
			Mode:          Loop,
			Timeout:       120 * time.Second,
			MaxIterations: 10,
		}
	}
	config.Mode = Loop
	return newBasicWorkflow(config)
}

// NewWorkflow creates a workflow with the specified configuration
func NewWorkflow(config *WorkflowConfig) (Workflow, error) {
	if config == nil {
		return nil, fmt.Errorf("workflow configuration is required")
	}

	// Use factory if registered (for plugin-based implementations)
	if factory := getWorkflowFactory(); factory != nil {
		return factory(config)
	}

	// Return basic implementation
	return newBasicWorkflow(config)
}

// =============================================================================
// BASIC WORKFLOW IMPLEMENTATION
// =============================================================================

// basicWorkflow provides a straightforward workflow implementation
type basicWorkflow struct {
	config  *WorkflowConfig
	steps   []WorkflowStep
	memory  Memory
	context *WorkflowContext
	mu      sync.RWMutex
}

// newBasicWorkflow creates a new basic workflow implementation
func newBasicWorkflow(config *WorkflowConfig) (*basicWorkflow, error) {
	if config == nil {
		return nil, fmt.Errorf("workflow configuration is required")
	}

	return &basicWorkflow{
		config: config,
		steps:  make([]WorkflowStep, 0),
		context: &WorkflowContext{
			WorkflowID:  fmt.Sprintf("wf_%d", time.Now().UnixNano()),
			StepResults: make(map[string]*StepResult),
			Variables:   make(map[string]interface{}),
		},
	}, nil
}

// Run implements Workflow.Run
func (w *basicWorkflow) Run(ctx context.Context, input string) (*WorkflowResult, error) {
	startTime := time.Now()

	// Apply timeout from config
	if w.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, w.config.Timeout)
		defer cancel()
	}

	// Initialize workflow context
	w.context.Set("initial_input", input)
	w.context.Set("start_time", startTime)

	// Execute based on workflow mode
	var stepResults []StepResult
	var finalOutput string
	var err error

	switch w.config.Mode {
	case Sequential:
		stepResults, finalOutput, err = w.executeSequential(ctx, input)
	case Parallel:
		stepResults, finalOutput, err = w.executeParallel(ctx, input)
	case DAG:
		stepResults, finalOutput, err = w.executeDAG(ctx, input)
	case Loop:
		stepResults, finalOutput, err = w.executeLoop(ctx, input)
	default:
		return nil, fmt.Errorf("unsupported workflow mode: %s", w.config.Mode)
	}

	// Build execution path
	executionPath := make([]string, 0, len(stepResults))
	totalTokens := 0
	for _, result := range stepResults {
		if !result.Skipped {
			executionPath = append(executionPath, result.StepName)
			totalTokens += result.Tokens
		}
	}

	return &WorkflowResult{
		Success:       err == nil,
		FinalOutput:   finalOutput,
		StepResults:   stepResults,
		Duration:      time.Since(startTime),
		TotalTokens:   totalTokens,
		ExecutionPath: executionPath,
		Metadata: map[string]interface{}{
			"workflow_id": w.context.WorkflowID,
			"mode":        string(w.config.Mode),
		},
		Error: errToString(err),
	}, nil
}

// RunStream implements Workflow.RunStream
func (w *basicWorkflow) RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error) {
	// Use the original context for agent execution
	// Let individual agents handle their own timeouts
	agentCtx := ctx

	// Create stream using the original context
	metadata := &StreamMetadata{
		AgentName: fmt.Sprintf("workflow_%s", w.config.Mode),
		StartTime: time.Now(),
		Extra: map[string]interface{}{
			"workflow_id": w.context.WorkflowID,
			"mode":        string(w.config.Mode),
		},
	}

	stream, writer := NewStream(agentCtx, metadata, opts...)

	// Start workflow execution in goroutine
	go func() {
		defer writer.Close()
		startTime := time.Now()

		// Debug: Check if context is already cancelled
		select {
		case <-agentCtx.Done():
			err := fmt.Errorf("agent context already cancelled before workflow start: %w", agentCtx.Err())
			writer.Write(&StreamChunk{
				Type:  ChunkTypeError,
				Error: err,
			})
			writer.CloseWithError(err)
			return
		default:
		}

		// Initialize workflow context
		w.context.Set("initial_input", input)
		w.context.Set("start_time", startTime)

		// Emit workflow start
		writer.Write(&StreamChunk{
			Type:    ChunkTypeMetadata,
			Content: fmt.Sprintf("Starting %s workflow", w.config.Mode),
			Metadata: map[string]interface{}{
				"workflow_id": w.context.WorkflowID,
				"mode":        string(w.config.Mode),
				"steps":       len(w.steps),
			},
		})

		// Execute based on workflow mode
		var stepResults []StepResult
		var finalOutput string
		var err error

		switch w.config.Mode {
		case Sequential:
			stepResults, finalOutput, err = w.executeSequentialStreaming(agentCtx, input, writer)
		case Parallel:
			stepResults, finalOutput, err = w.executeParallelStreaming(agentCtx, input, writer)
		case DAG:
			stepResults, finalOutput, err = w.executeDAGStreaming(agentCtx, input, writer)
		case Loop:
			stepResults, finalOutput, err = w.executeLoopStreaming(agentCtx, input, writer)
		default:
			err = fmt.Errorf("unsupported workflow mode: %s", w.config.Mode)
		}

		if err != nil {
			writer.Write(&StreamChunk{
				Type:  ChunkTypeError,
				Error: err,
			})
			writer.CloseWithError(err)
			return
		}

		// Emit final text chunk
		writer.Write(&StreamChunk{
			Type:    ChunkTypeText,
			Content: finalOutput,
		})

		// Emit done chunk
		writer.Write(&StreamChunk{
			Type: ChunkTypeDone,
		})

		// Build result
		executionPath := make([]string, 0, len(stepResults))
		totalTokens := 0
		for _, result := range stepResults {
			if !result.Skipped {
				executionPath = append(executionPath, result.StepName)
				totalTokens += result.Tokens
			}
		}

		workflowResult := &WorkflowResult{
			Success:       true,
			FinalOutput:   finalOutput,
			StepResults:   stepResults,
			Duration:      time.Since(startTime),
			TotalTokens:   totalTokens,
			ExecutionPath: executionPath,
			Metadata: map[string]interface{}{
				"workflow_id": w.context.WorkflowID,
				"mode":        string(w.config.Mode),
				"streamed":    true,
			},
		}

		// Convert to Result and set on stream
		result := &Result{
			Success:  true,
			Content:  finalOutput,
			Duration: time.Since(startTime),
			Metadata: workflowResult.Metadata,
		}

		if s, ok := stream.(*basicStream); ok {
			s.SetResult(result)
		}
	}()

	return stream, nil
}

// safeStreamWrite writes to the stream with panic recovery
func safeStreamWrite(writer StreamWriter, chunk *StreamChunk, stepName string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("stream write panic in step %s: %v", stepName, r)
		}
	}()

	if writer != nil && chunk != nil {
		writer.Write(chunk)
	}
	return nil
}

// executeSequentialStreaming runs steps one after another with streaming
func (w *basicWorkflow) executeSequentialStreaming(ctx context.Context, input string, writer StreamWriter) ([]StepResult, string, error) {
	w.mu.RLock()
	steps := w.steps
	w.mu.RUnlock()

	results := make([]StepResult, 0, len(steps))
	currentInput := input

	for i, step := range steps {
		// Check context cancellation
		select {
		case <-ctx.Done():
			contextErr := fmt.Errorf("sequential workflow cancelled at step %d/%d (%s): %w", i+1, len(steps), step.Name, ctx.Err())
			return results, currentInput, contextErr
		default:
		}

		// Emit step start with safe writing
		stepStartChunk := &StreamChunk{
			Type:    ChunkTypeMetadata,
			Content: fmt.Sprintf("Step %d/%d: %s", i+1, len(steps), step.Name),
			Metadata: map[string]interface{}{
				"step_name":   step.Name,
				"step_index":  i,
				"total_steps": len(steps),
			},
		}

		if writeErr := safeStreamWrite(writer, stepStartChunk, step.Name); writeErr != nil {
			// Log warning but continue execution
			fmt.Printf("Warning: Failed to write step start chunk: %v\n", writeErr)
		}

		// Execute step with streaming if agent supports it
		stepResult, output, err := w.executeStepStreaming(ctx, step, currentInput, writer)
		results = append(results, stepResult)

		if err != nil {
			// Enhanced error with step context
			stepErr := fmt.Errorf("sequential workflow step %d/%d (%s) failed after processing %d steps: %w",
				i+1, len(steps), step.Name, len(results), err)
			return results, currentInput, stepErr
		}

		currentInput = output
	}

	return results, currentInput, nil
}

// executeParallelStreaming runs all steps concurrently with streaming
func (w *basicWorkflow) executeParallelStreaming(ctx context.Context, input string, writer StreamWriter) ([]StepResult, string, error) {
	w.mu.RLock()
	steps := w.steps
	w.mu.RUnlock()

	if len(steps) == 0 {
		return []StepResult{}, input, nil
	}

	results := make([]StepResult, len(steps))
	outputs := make([]string, len(steps))
	errors := make([]error, len(steps))

	var wg sync.WaitGroup
	for i, step := range steps {
		wg.Add(1)
		go func(idx int, s WorkflowStep) {
			defer wg.Done()

			writer.Write(&StreamChunk{
				Type:    ChunkTypeMetadata,
				Content: fmt.Sprintf("Starting parallel step: %s", s.Name),
				Metadata: map[string]interface{}{
					"step_name": s.Name,
				},
			})

			result, output, err := w.executeStepStreaming(ctx, s, input, writer)
			results[idx] = result
			outputs[idx] = output
			errors[idx] = err
		}(i, step)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			return results, "", fmt.Errorf("parallel step %s failed: %w", steps[i].Name, err)
		}
	}

	// Aggregate outputs
	finalOutput := ""
	for _, output := range outputs {
		if output != "" {
			finalOutput += output + "\n"
		}
	}

	return results, finalOutput, nil
}

// executeDAGStreaming runs steps based on dependency graph with streaming
func (w *basicWorkflow) executeDAGStreaming(ctx context.Context, input string, writer StreamWriter) ([]StepResult, string, error) {
	// For simplicity, delegate to sequential for now
	// Full DAG implementation would handle dependencies
	return w.executeSequentialStreaming(ctx, input, writer)
}

// executeLoopStreaming runs steps iteratively with streaming
func (w *basicWorkflow) executeLoopStreaming(ctx context.Context, input string, writer StreamWriter) ([]StepResult, string, error) {
	w.mu.RLock()
	maxIterations := w.config.MaxIterations
	w.mu.RUnlock()

	if maxIterations == 0 {
		maxIterations = 10 // Default max iterations
	}

	allResults := make([]StepResult, 0)
	currentInput := input
	iteration := 0

	for iteration < maxIterations {
		writer.Write(&StreamChunk{
			Type:    ChunkTypeMetadata,
			Content: fmt.Sprintf("Loop iteration %d/%d", iteration+1, maxIterations),
			Metadata: map[string]interface{}{
				"iteration": iteration + 1,
			},
		})

		// Execute all steps in sequence
		results, output, err := w.executeSequentialStreaming(ctx, currentInput, writer)
		allResults = append(allResults, results...)

		if err != nil {
			return allResults, currentInput, err
		}

		currentInput = output
		iteration++

		// Check for convergence (simplified - would need proper condition)
		if output == input {
			break
		}
	}

	return allResults, currentInput, nil
}

// executeStepStreaming executes a single workflow step with streaming
func (w *basicWorkflow) executeStepStreaming(ctx context.Context, step WorkflowStep, input string, writer StreamWriter) (StepResult, string, error) {
	startTime := time.Now()

	// Check context cancellation before step execution
	select {
	case <-ctx.Done():
		contextErr := ctx.Err()
		stepErr := fmt.Errorf("step %s cancelled before execution: %w", step.Name, contextErr)
		return StepResult{
			StepName:  step.Name,
			Success:   false,
			Output:    "",
			Error:     stepErr.Error(),
			Timestamp: startTime,
			Duration:  time.Since(startTime),
		}, "", stepErr
	default:
	}

	// Check condition
	if step.Condition != nil && !step.Condition(ctx, w.context) {
		return StepResult{
			StepName:  step.Name,
			Success:   true,
			Skipped:   true,
			Timestamp: startTime,
			Duration:  time.Since(startTime),
		}, input, nil
	}

	// Apply transformation if provided
	if step.Transform != nil {
		input = step.Transform(input)
	}

	// Try to run with streaming if agent supports it
	stream, err := step.Agent.RunStream(ctx, input)
	if err != nil {
		// Enhanced error context
		enhancedErr := fmt.Errorf("step %s agent streaming failed: %w", step.Name, err)
		if ctx.Err() != nil {
			enhancedErr = fmt.Errorf("step %s cancelled during agent start (context: %v): %w", step.Name, ctx.Err(), err)
		}

		return StepResult{
			StepName:  step.Name,
			Success:   false,
			Output:    "",
			Error:     enhancedErr.Error(),
			Timestamp: startTime,
			Duration:  time.Since(startTime),
		}, "", enhancedErr
	}

	// Forward chunks from agent stream to workflow stream
	var output string
	chunkCount := 0
	for chunk := range stream.Chunks() {
		chunkCount++

		// Check for context cancellation during streaming
		select {
		case <-ctx.Done():
			contextErr := fmt.Errorf("step %s cancelled during streaming at chunk %d: %w", step.Name, chunkCount, ctx.Err())
			return StepResult{
				StepName:  step.Name,
				Success:   false,
				Output:    output,
				Error:     contextErr.Error(),
				Timestamp: startTime,
				Duration:  time.Since(startTime),
			}, output, contextErr
		default:
		}

		// Modify chunk metadata to include step name
		if chunk.Metadata == nil {
			chunk.Metadata = make(map[string]interface{})
		}
		chunk.Metadata["step_name"] = step.Name
		chunk.Metadata["chunk_count"] = chunkCount

		// Use safe stream writing
		if writeErr := safeStreamWrite(writer, chunk, step.Name); writeErr != nil {
			// Log warning but continue processing
			fmt.Printf("Warning: Failed to write chunk %d for step %s: %v\n", chunkCount, step.Name, writeErr)
		}

		// Collect text for final output
		if chunk.Type == ChunkTypeText || chunk.Type == ChunkTypeDelta {
			if chunk.Content != "" {
				output += chunk.Content
			} else {
				output += chunk.Delta
			}
		}
	}

	result, streamErr := stream.Wait()
	if streamErr != nil {
		// Enhanced error context for stream errors
		if ctx.Err() != nil {
			err = fmt.Errorf("step %s stream wait failed with context cancellation (context: %v): %w", step.Name, ctx.Err(), streamErr)
		} else {
			err = fmt.Errorf("step %s stream wait failed: %w", step.Name, streamErr)
		}
	}

	stepResult := StepResult{
		StepName:  step.Name,
		Success:   err == nil,
		Output:    output,
		Timestamp: startTime,
		Duration:  time.Since(startTime),
	}

	if result != nil {
		stepResult.Tokens = result.TokensUsed
	}

	if err != nil {
		stepResult.Error = err.Error()
		// Log step failure with context
		return stepResult, "", fmt.Errorf("workflow step %s failed after %.2fs: %w", step.Name, time.Since(startTime).Seconds(), err)
	}

	// Store result in context
	w.context.mu.Lock()
	w.context.StepResults[step.Name] = &stepResult
	w.context.mu.Unlock()

	return stepResult, output, nil
}

// executeSequential runs steps one after another
func (w *basicWorkflow) executeSequential(ctx context.Context, input string) ([]StepResult, string, error) {
	w.mu.RLock()
	steps := w.steps
	w.mu.RUnlock()

	results := make([]StepResult, 0, len(steps))
	currentInput := input

	for _, step := range steps {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return results, currentInput, ctx.Err()
		default:
		}

		// Check condition
		if step.Condition != nil && !step.Condition(ctx, w.context) {
			results = append(results, StepResult{
				StepName:  step.Name,
				Skipped:   true,
				Timestamp: time.Now(),
			})
			continue
		}

		// Apply input transformation
		stepInput := currentInput
		if step.Transform != nil {
			stepInput = step.Transform(currentInput)
		}

		// Execute step
		result := w.executeStep(ctx, &step, stepInput)
		results = append(results, result)

		// Store result in context
		w.context.mu.Lock()
		w.context.StepResults[step.Name] = &result
		w.context.mu.Unlock()

		// If step failed, stop workflow
		if !result.Success {
			return results, currentInput, fmt.Errorf("step %s failed: %s", step.Name, result.Error)
		}

		// Use output as input for next step
		currentInput = result.Output
	}

	return results, currentInput, nil
}

// executeParallel runs all steps concurrently
func (w *basicWorkflow) executeParallel(ctx context.Context, input string) ([]StepResult, string, error) {
	w.mu.RLock()
	steps := w.steps
	w.mu.RUnlock()

	results := make([]StepResult, len(steps))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	for i, step := range steps {
		wg.Add(1)
		go func(idx int, s WorkflowStep) {
			defer wg.Done()

			// Check condition
			if s.Condition != nil && !s.Condition(ctx, w.context) {
				mu.Lock()
				results[idx] = StepResult{
					StepName:  s.Name,
					Skipped:   true,
					Timestamp: time.Now(),
				}
				mu.Unlock()
				return
			}

			// Apply input transformation
			stepInput := input
			if s.Transform != nil {
				stepInput = s.Transform(input)
			}

			// Execute step
			result := w.executeStep(ctx, &s, stepInput)

			mu.Lock()
			results[idx] = result
			w.context.StepResults[s.Name] = &result
			if !result.Success {
				errors = append(errors, fmt.Errorf("step %s failed: %s", s.Name, result.Error))
			}
			mu.Unlock()
		}(i, step)
	}

	wg.Wait()

	// Combine outputs (concatenate all successful outputs)
	var finalOutput string
	for _, result := range results {
		if result.Success && !result.Skipped {
			if finalOutput != "" {
				finalOutput += "\n"
			}
			finalOutput += result.Output
		}
	}

	// Return first error if any
	var err error
	if len(errors) > 0 {
		err = errors[0]
	}

	return results, finalOutput, err
}

// executeDAG runs steps based on dependency order
func (w *basicWorkflow) executeDAG(ctx context.Context, input string) ([]StepResult, string, error) {
	w.mu.RLock()
	steps := w.steps
	w.mu.RUnlock()

	// Build dependency graph
	completed := make(map[string]bool)
	results := make([]StepResult, 0, len(steps))
	finalOutput := input

	// Execute steps in dependency order
	for len(completed) < len(steps) {
		executed := false

		for _, step := range steps {
			// Skip if already completed
			if completed[step.Name] {
				continue
			}

			// Check if all dependencies are satisfied
			canExecute := true
			for _, dep := range step.Dependencies {
				if !completed[dep] {
					canExecute = false
					break
				}
			}

			if !canExecute {
				continue
			}

			// Check condition
			if step.Condition != nil && !step.Condition(ctx, w.context) {
				result := StepResult{
					StepName:  step.Name,
					Skipped:   true,
					Timestamp: time.Now(),
				}
				results = append(results, result)
				completed[step.Name] = true
				executed = true
				continue
			}

			// Build input from dependencies
			stepInput := w.buildInputFromDependencies(step.Dependencies, input)
			if step.Transform != nil {
				stepInput = step.Transform(stepInput)
			}

			// Execute step
			result := w.executeStep(ctx, &step, stepInput)
			results = append(results, result)

			// Store result
			w.context.mu.Lock()
			w.context.StepResults[step.Name] = &result
			w.context.mu.Unlock()

			completed[step.Name] = true
			executed = true

			if !result.Success {
				return results, finalOutput, fmt.Errorf("step %s failed: %s", step.Name, result.Error)
			}

			finalOutput = result.Output
		}

		// Check for deadlock (circular dependencies)
		if !executed {
			return results, finalOutput, fmt.Errorf("workflow deadlock detected: circular dependencies or missing steps")
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return results, finalOutput, ctx.Err()
		default:
		}
	}

	return results, finalOutput, nil
}

// executeLoop repeats steps until max iterations or condition is met
func (w *basicWorkflow) executeLoop(ctx context.Context, input string) ([]StepResult, string, error) {
	maxIterations := w.config.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10
	}

	allResults := make([]StepResult, 0)
	currentInput := input

	for iteration := 0; iteration < maxIterations; iteration++ {
		w.context.IterationNum = iteration

		// Execute one iteration
		iterResults, output, err := w.executeSequential(ctx, currentInput)
		allResults = append(allResults, iterResults...)

		if err != nil {
			return allResults, output, err
		}

		// Check loop condition (stored in context)
		shouldContinue, _ := w.context.Get("loop_continue")
		if shouldContinue == false {
			return allResults, output, nil
		}

		currentInput = output
	}

	return allResults, currentInput, nil
}

// executeStep executes a single workflow step
func (w *basicWorkflow) executeStep(ctx context.Context, step *WorkflowStep, input string) StepResult {
	startTime := time.Now()
	w.context.CurrentStep = step.Name

	// Store input in workflow memory if available
	if w.memory != nil {
		_ = w.memory.Store(ctx, input, WithContentType("workflow_step_input"), WithSource(step.Name))
	}

	// Execute the agent
	result, err := step.Agent.Run(ctx, input)

	stepResult := StepResult{
		StepName:  step.Name,
		Duration:  time.Since(startTime),
		Timestamp: startTime,
	}

	if err != nil {
		stepResult.Success = false
		stepResult.Error = err.Error()
		return stepResult
	}

	stepResult.Success = result.Success
	stepResult.Output = result.Content
	stepResult.Tokens = result.TokensUsed
	if result.Error != "" {
		stepResult.Error = result.Error
	}

	// Store output in workflow memory if available
	if w.memory != nil && stepResult.Success {
		_ = w.memory.Store(ctx, stepResult.Output, WithContentType("workflow_step_output"), WithSource(step.Name))
	}

	return stepResult
}

// buildInputFromDependencies creates input from completed dependency outputs
func (w *basicWorkflow) buildInputFromDependencies(deps []string, defaultInput string) string {
	if len(deps) == 0 {
		return defaultInput
	}

	// Combine outputs from all dependencies
	var combined string
	for _, dep := range deps {
		if result, ok := w.context.GetStepResult(dep); ok && result.Success {
			if combined != "" {
				combined += "\n"
			}
			combined += result.Output
		}
	}

	if combined == "" {
		return defaultInput
	}
	return combined
}

// AddStep implements Workflow.AddStep
func (w *basicWorkflow) AddStep(step WorkflowStep) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Validate step
	if step.Name == "" {
		return fmt.Errorf("step name is required")
	}
	if step.Agent == nil {
		return fmt.Errorf("step agent is required")
	}

	// Check for duplicate names
	for _, existing := range w.steps {
		if existing.Name == step.Name {
			return fmt.Errorf("step with name %s already exists", step.Name)
		}
	}

	w.steps = append(w.steps, step)
	return nil
}

// SetMemory implements Workflow.SetMemory
func (w *basicWorkflow) SetMemory(memory Memory) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.memory = memory
	w.context.SharedMemory = memory
}

// GetConfig implements Workflow.GetConfig
func (w *basicWorkflow) GetConfig() *WorkflowConfig {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.config
}

// Initialize implements Workflow.Initialize
func (w *basicWorkflow) Initialize(ctx context.Context) error {
	// Initialize all step agents
	for _, step := range w.steps {
		if err := step.Agent.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize agent %s: %w", step.Name, err)
		}
	}
	return nil
}

// Shutdown implements Workflow.Shutdown
func (w *basicWorkflow) Shutdown(ctx context.Context) error {
	// Cleanup all step agents
	for _, step := range w.steps {
		if err := step.Agent.Cleanup(ctx); err != nil {
			return fmt.Errorf("failed to cleanup agent %s: %w", step.Name, err)
		}
	}
	return nil
}

// =============================================================================
// WORKFLOW FACTORY REGISTRY
// =============================================================================

// WorkflowFactory creates a Workflow implementation based on WorkflowConfig
type WorkflowFactory func(*WorkflowConfig) (Workflow, error)

var (
	workflowFactory WorkflowFactory
	workflowMutex   sync.RWMutex
)

// SetWorkflowFactory allows plugins to register a custom Workflow factory
func SetWorkflowFactory(factory WorkflowFactory) {
	workflowMutex.Lock()
	defer workflowMutex.Unlock()
	workflowFactory = factory
}

// getWorkflowFactory returns the registered Workflow factory
func getWorkflowFactory() WorkflowFactory {
	workflowMutex.RLock()
	defer workflowMutex.RUnlock()
	return workflowFactory
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// errToString converts an error to a string, returning empty string for nil
func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// =============================================================================
// EXAMPLE USAGE AND DOCUMENTATION
// =============================================================================

/*
Example usage of the workflow system:

Sequential workflow:
	workflow, err := NewSequentialWorkflow(&WorkflowConfig{
		Timeout: 60 * time.Second,
	})

	// Add steps
	workflow.AddStep(WorkflowStep{
		Name:  "analyze",
		Agent: analyzerAgent,
	})
	workflow.AddStep(WorkflowStep{
		Name:  "summarize",
		Agent: summarizerAgent,
	})

	result, err := workflow.Run(ctx, "Initial input")

Parallel workflow:
	workflow, err := NewParallelWorkflow(&WorkflowConfig{
		Timeout: 30 * time.Second,
	})

	// All steps run concurrently
	workflow.AddStep(WorkflowStep{Name: "fact_check", Agent: factChecker})
	workflow.AddStep(WorkflowStep{Name: "sentiment", Agent: sentimentAnalyzer})
	workflow.AddStep(WorkflowStep{Name: "summarize", Agent: summarizer})

	result, err := workflow.Run(ctx, "Article content")

DAG workflow with dependencies:
	workflow, err := NewDAGWorkflow(&WorkflowConfig{
		Timeout: 120 * time.Second,
	})

	// Steps execute based on dependencies
	workflow.AddStep(WorkflowStep{
		Name:  "fetch_data",
		Agent: dataFetcher,
	})
	workflow.AddStep(WorkflowStep{
		Name:         "analyze",
		Agent:        analyzer,
		Dependencies: []string{"fetch_data"},
	})
	workflow.AddStep(WorkflowStep{
		Name:         "visualize",
		Agent:        visualizer,
		Dependencies: []string{"fetch_data"},
	})
	workflow.AddStep(WorkflowStep{
		Name:         "report",
		Agent:        reporter,
		Dependencies: []string{"analyze", "visualize"},
	})

	result, err := workflow.Run(ctx, "Generate report")

Loop workflow:
	workflow, err := NewLoopWorkflow(&WorkflowConfig{
		Timeout:       300 * time.Second,
		MaxIterations: 5,
	})

	workflow.AddStep(WorkflowStep{
		Name:  "research",
		Agent: researchAgent,
	})
	workflow.AddStep(WorkflowStep{
		Name:  "refine",
		Agent: refineAgent,
	})

	// Stop loop by setting context variable
	// Inside agent logic: context.Set("loop_continue", false)

	result, err := workflow.Run(ctx, "Research topic")

Workflow with shared memory:
	// Create shared memory
	memory, _ := NewMemory(&MemoryConfig{
		Provider: "memory",
	})

	workflow, _ := NewSequentialWorkflow(nil)
	workflow.SetMemory(memory)

	// Steps can now access shared memory
	workflow.AddStep(WorkflowStep{
		Name:  "collect",
		Agent: collectorAgent,
	})
	workflow.AddStep(WorkflowStep{
		Name:  "analyze",
		Agent: analyzerAgent, // Can query memory from previous step
	})

Conditional steps:
	workflow.AddStep(WorkflowStep{
		Name:  "optional_step",
		Agent: optionalAgent,
		Condition: func(ctx context.Context, wc *WorkflowContext) bool {
			// Only run if previous step succeeded
			if result, ok := wc.GetStepResult("previous_step"); ok {
				return result.Success
			}
			return false
		},
	})

Input transformation:
	workflow.AddStep(WorkflowStep{
		Name:  "processor",
		Agent: processorAgent,
		Transform: func(input string) string {
			// Modify input before passing to agent
			return "Process this: " + input
		},
	})

Accessing workflow results:
	result, err := workflow.Run(ctx, "input")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Final output: %s\n", result.FinalOutput)
	fmt.Printf("Total duration: %v\n", result.Duration)
	fmt.Printf("Total tokens: %d\n", result.TotalTokens)
	fmt.Printf("Execution path: %v\n", result.ExecutionPath)

	// Access individual step results
	for _, stepResult := range result.StepResults {
		fmt.Printf("Step %s: %s (tokens: %d)\n",
			stepResult.StepName,
			stepResult.Output,
			stepResult.Tokens)
	}
*/
