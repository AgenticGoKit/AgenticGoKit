// package v1beta provides streamlined workflow orchestration for multi-agent systems.
// This file implements SubWorkflow composition - the ability to use workflows as agents.
package v1beta

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// =============================================================================
// SUBWORKFLOW AGENT - CORE WRAPPER
// =============================================================================

// SubWorkflowAgent wraps a Workflow to implement the Agent interface.
// This allows workflows to be used as agents within other workflows,
// enabling recursive composition and hierarchical workflow patterns.
//
// Example:
//
//	parallelWorkflow, _ := vnext.NewParallelWorkflow(&vnext.WorkflowConfig{Name: "Analysis"})
//	analysisAgent := vnext.NewSubWorkflowAgent("analysis", parallelWorkflow)
//	mainWorkflow.AddStep(vnext.WorkflowStep{Name: "analyze", Agent: analysisAgent})
type SubWorkflowAgent struct {
	workflow    Workflow
	name        string
	description string
	config      *Config

	// Nesting tracking
	depth      int
	maxDepth   int
	parentPath string // e.g., "main/analyze/parallel"

	// Statistics
	execCount     int64
	totalDuration time.Duration
	mu            sync.RWMutex
}

// SubWorkflowOption is a functional option for configuring SubWorkflowAgent
type SubWorkflowOption func(*SubWorkflowAgent)

// NewSubWorkflowAgent creates an agent that wraps a workflow.
//
// This is the primary entry point for creating subworkflows. It wraps any
// Workflow implementation to act as an Agent, enabling hierarchical composition.
//
// Example:
//
//	parallelWorkflow, _ := vnext.NewParallelWorkflow(&vnext.WorkflowConfig{Name: "Analysis"})
//	parallelWorkflow.AddStep(vnext.WorkflowStep{Name: "sentiment", Agent: sentimentAgent})
//
//	// Wrap as agent
//	analysisAgent := vnext.NewSubWorkflowAgent("analysis", parallelWorkflow)
//
//	// Use in main workflow
//	mainWorkflow.AddStep(vnext.WorkflowStep{
//	    Name: "analyze",
//	    Agent: analysisAgent, // SubWorkflow!
//	})
func NewSubWorkflowAgent(name string, workflow Workflow, opts ...SubWorkflowOption) Agent {
	wa := &SubWorkflowAgent{
		workflow:    workflow,
		name:        name,
		description: fmt.Sprintf("Subworkflow agent wrapping %s workflow", name),
		maxDepth:    10, // Default safety limit
		config: &Config{
			Name: name,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(wa)
	}

	return wa
}

// =============================================================================
// AGENT INTERFACE IMPLEMENTATION
// =============================================================================

// Run implements Agent.Run by executing the wrapped workflow and converting results
func (wa *SubWorkflowAgent) Run(ctx context.Context, input string) (*Result, error) {
	return wa.RunWithOptions(ctx, input, nil)
}

// RunWithOptions implements Agent.RunWithOptions
func (wa *SubWorkflowAgent) RunWithOptions(ctx context.Context, input string, runOpts *RunOptions) (*Result, error) {
	wa.mu.Lock()
	wa.execCount++
	wa.mu.Unlock()

	startTime := time.Now()

	// Check nesting depth for safety
	if wa.depth > 0 && wa.depth >= wa.maxDepth {
		return nil, fmt.Errorf("maximum workflow nesting depth (%d) exceeded at %s",
			wa.maxDepth, wa.getFullPath())
	}

	// Apply timeout from RunOptions if specified
	if runOpts != nil && runOpts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, runOpts.Timeout)
		defer cancel()
	}

	// Add workflow metadata to context for tracing
	ctx = wa.enrichContext(ctx)

	// Execute the wrapped workflow
	workflowResult, err := wa.workflow.Run(ctx, input)

	duration := time.Since(startTime)
	wa.mu.Lock()
	wa.totalDuration += duration
	wa.mu.Unlock()

	if err != nil {
		return &Result{
			Success:  false,
			Error:    fmt.Sprintf("subworkflow %s failed: %v", wa.name, err),
			Duration: duration,
			Metadata: wa.buildMetadata(workflowResult, false),
		}, err
	}

	// Convert WorkflowResult to Agent Result
	return wa.convertToResult(workflowResult, duration), nil
}

// RunStream implements Agent.RunStream by streaming the wrapped workflow
func (wa *SubWorkflowAgent) RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error) {
	wa.mu.Lock()
	wa.execCount++
	wa.mu.Unlock()

	// Check nesting depth for safety
	if wa.depth > 0 && wa.depth >= wa.maxDepth {
		return nil, fmt.Errorf("maximum workflow nesting depth (%d) exceeded at %s",
			wa.maxDepth, wa.getFullPath())
	}

	// Add workflow metadata to context for tracing
	ctx = wa.enrichContext(ctx)

	// Create wrapper stream that adds subworkflow metadata
	stream, err := wa.workflow.RunStream(ctx, input, opts...)
	if err != nil {
		return nil, fmt.Errorf("subworkflow %s streaming failed: %w", wa.name, err)
	}

	// Wrap stream to inject metadata
	return wa.wrapStream(stream), nil
}

// RunStreamWithOptions implements Agent.RunStreamWithOptions
func (wa *SubWorkflowAgent) RunStreamWithOptions(ctx context.Context, input string, runOpts *RunOptions, streamOpts ...StreamOption) (Stream, error) {
	// Delegate to RunStream for now
	return wa.RunStream(ctx, input, streamOpts...)
}

// Name implements Agent.Name
func (wa *SubWorkflowAgent) Name() string {
	return wa.name
}

// Config implements Agent.Config
func (wa *SubWorkflowAgent) Config() *Config {
	return wa.config
}

// Capabilities implements Agent.Capabilities
func (wa *SubWorkflowAgent) Capabilities() []string {
	return []string{
		"workflow_execution",
		"workflow_composition",
		string(wa.workflow.GetConfig().Mode),
	}
}

// Initialize implements Agent.Initialize
func (wa *SubWorkflowAgent) Initialize(ctx context.Context) error {
	return wa.workflow.Initialize(ctx)
}

// Cleanup implements Agent.Cleanup
func (wa *SubWorkflowAgent) Cleanup(ctx context.Context) error {
	return wa.workflow.Shutdown(ctx)
}

// =============================================================================
// HELPER METHODS
// =============================================================================

// convertToResult converts a WorkflowResult to an Agent Result
func (wa *SubWorkflowAgent) convertToResult(workflowResult *WorkflowResult, duration time.Duration) *Result {
	return &Result{
		Success:    workflowResult.Success,
		Content:    workflowResult.FinalOutput,
		TokensUsed: workflowResult.TotalTokens,
		Duration:   duration,
		Metadata:   wa.buildMetadata(workflowResult, true),
	}
}

// buildMetadata creates metadata for the result
func (wa *SubWorkflowAgent) buildMetadata(workflowResult *WorkflowResult, includeDetails bool) map[string]interface{} {
	metadata := map[string]interface{}{
		"type":          "subworkflow",
		"workflow_name": wa.name,
		"workflow_path": wa.getFullPath(),
		"depth":         wa.depth,
	}

	if workflowResult != nil {
		metadata["step_count"] = len(workflowResult.StepResults)
		metadata["execution_path"] = workflowResult.ExecutionPath
		metadata["workflow_duration"] = workflowResult.Duration.String()

		if includeDetails {
			metadata["step_results"] = workflowResult.StepResults
			metadata["workflow_metadata"] = workflowResult.Metadata
		}
	}

	wa.mu.RLock()
	metadata["execution_count"] = wa.execCount
	if wa.execCount > 0 {
		metadata["avg_duration"] = (wa.totalDuration / time.Duration(wa.execCount)).String()
	}
	wa.mu.RUnlock()

	return metadata
}

// Memory returns the memory provider for this agent (delegating to the workflow)
func (a *SubWorkflowAgent) Memory() Memory {
	// SubWorkflowAgent wraps a Workflow. We need to expose its memory.
	// Since we are also updating Workflow interface to have Memory(),
	// we can delegate this call once Workflow is updated.
	// However, currently Workflow interface might not have Memory() yet
	// (it's the next step in my plan).
	// To avoid compile errors if I do this out of order, I should assume Workflow will have it.
	// But `a.workflow` is of type `Workflow` interface?
	// Let's check struct definition.
	// line 12: type SubWorkflowAgent struct { workflow Workflow ... }
	// So yes, I rely on Workflow interface change.
	// I should update Workflow interface FIRST or concurrently.
	// Since I am already here, I will write the code assuming Workflow has it.
	// If I compile/test now, it will fail until Step 5 is done.
	// But since I am doing sequential edits, I will finish this file then do workflow.go.
	return a.workflow.Memory()
}

// getFullPath returns the full hierarchical path of this workflow agent
func (wa *SubWorkflowAgent) getFullPath() string {
	if wa.parentPath == "" {
		return wa.name
	}
	return wa.parentPath + "/" + wa.name
}

// enrichContext adds workflow nesting information to context for tracing
func (wa *SubWorkflowAgent) enrichContext(ctx context.Context) context.Context {
	type contextKey string
	const (
		workflowPathKey  contextKey = "workflow_path"
		workflowDepthKey contextKey = "workflow_depth"
	)

	ctx = context.WithValue(ctx, workflowPathKey, wa.getFullPath())
	ctx = context.WithValue(ctx, workflowDepthKey, wa.depth)

	return ctx
}

// wrapStream wraps a stream to add subworkflow metadata to chunks
// and forward nested agent lifecycle events to parent stream
func (wa *SubWorkflowAgent) wrapStream(stream Stream) Stream {
	// Add subworkflow metadata to stream metadata
	metadata := stream.Metadata()
	if metadata.Extra == nil {
		metadata.Extra = make(map[string]interface{})
	}
	metadata.Extra["subworkflow_name"] = wa.name
	metadata.Extra["subworkflow_path"] = wa.getFullPath()
	metadata.Extra["subworkflow_depth"] = wa.depth

	// Create wrapped stream that intercepts chunks and adds subworkflow context
	return &wrappedSubworkflowStream{
		inner:            stream,
		subworkflowName:  wa.name,
		subworkflowPath:  wa.getFullPath(),
		subworkflowDepth: wa.depth,
	}
}

// wrappedSubworkflowStream wraps a stream to enrich chunks with subworkflow context
// This ensures nested agent lifecycle events (ChunkTypeAgentStart, ChunkTypeAgentComplete)
// are forwarded to parent workflows so the UI can display nested agent outputs
type wrappedSubworkflowStream struct {
	inner            Stream
	subworkflowName  string
	subworkflowPath  string
	subworkflowDepth int
}

func (ws *wrappedSubworkflowStream) Chunks() <-chan *StreamChunk {
	outChan := make(chan *StreamChunk, 100)

	go func() {
		defer close(outChan)

		for chunk := range ws.inner.Chunks() {
			// Enrich chunk metadata with subworkflow context
			// This is crucial for nested agent lifecycle events (agent_start, agent_complete)
			// to be properly displayed in the UI
			if chunk.Metadata == nil {
				chunk.Metadata = make(map[string]interface{})
			}

			// Add parent subworkflow context (preserves existing step_name from nested agents)
			chunk.Metadata["parent_subworkflow"] = ws.subworkflowName
			chunk.Metadata["subworkflow_path"] = ws.subworkflowPath
			chunk.Metadata["subworkflow_depth"] = ws.subworkflowDepth

			// Forward the enriched chunk to parent stream
			outChan <- chunk
		}
	}()

	return outChan
}

func (ws *wrappedSubworkflowStream) Wait() (*Result, error) {
	return ws.inner.Wait()
}

func (ws *wrappedSubworkflowStream) Cancel() {
	ws.inner.Cancel()
}

func (ws *wrappedSubworkflowStream) Metadata() *StreamMetadata {
	return ws.inner.Metadata()
}

func (ws *wrappedSubworkflowStream) AsReader() io.Reader {
	return ws.inner.AsReader()
}

// =============================================================================
// PUBLIC METHODS FOR INTROSPECTION
// =============================================================================

// GetWorkflow returns the wrapped workflow for introspection
func (wa *SubWorkflowAgent) GetWorkflow() Workflow {
	return wa.workflow
}

// GetStats returns execution statistics for this workflow agent
func (wa *SubWorkflowAgent) GetStats() SubWorkflowStats {
	wa.mu.RLock()
	defer wa.mu.RUnlock()

	avgDuration := time.Duration(0)
	if wa.execCount > 0 {
		avgDuration = wa.totalDuration / time.Duration(wa.execCount)
	}

	return SubWorkflowStats{
		Name:           wa.name,
		Path:           wa.getFullPath(),
		Depth:          wa.depth,
		MaxDepth:       wa.maxDepth,
		Description:    wa.description,
		ExecutionCount: wa.execCount,
		TotalDuration:  wa.totalDuration,
		AvgDuration:    avgDuration,
	}
}

// SubWorkflowStats contains execution statistics for a subworkflow agent
type SubWorkflowStats struct {
	Name           string
	Path           string
	Depth          int
	MaxDepth       int
	Description    string
	ExecutionCount int64
	TotalDuration  time.Duration
	AvgDuration    time.Duration
}

// =============================================================================
// FUNCTIONAL OPTIONS
// =============================================================================

// WithSubWorkflowMaxDepth sets the maximum nesting depth to prevent infinite recursion
func WithSubWorkflowMaxDepth(depth int) SubWorkflowOption {
	return func(wa *SubWorkflowAgent) {
		wa.maxDepth = depth
	}
}

// WithSubWorkflowDescription sets a custom description for the subworkflow agent
func WithSubWorkflowDescription(desc string) SubWorkflowOption {
	return func(wa *SubWorkflowAgent) {
		wa.description = desc
	}
}

// WithSubWorkflowParentPath sets the parent workflow path for hierarchy tracking
func WithSubWorkflowParentPath(path string) SubWorkflowOption {
	return func(wa *SubWorkflowAgent) {
		wa.parentPath = path
	}
}

// WithSubWorkflowDepth sets the current nesting depth
func WithSubWorkflowDepth(depth int) SubWorkflowOption {
	return func(wa *SubWorkflowAgent) {
		wa.depth = depth
	}
}

// =============================================================================
// CONVENIENCE FACTORY FUNCTIONS
// =============================================================================

// QuickSubWorkflow creates a subworkflow agent with minimal configuration.
// This is a convenience wrapper around NewSubWorkflowAgent for simple use cases.
//
// Example:
//
//	workflow, _ := vnext.NewParallelWorkflow(&vnext.WorkflowConfig{Name: "Analysis"})
//	agent := vnext.QuickSubWorkflow("analysis", workflow)
func QuickSubWorkflow(name string, workflow Workflow) Agent {
	return NewSubWorkflowAgent(name, workflow)
}

// NewSequentialSubWorkflow creates a sequential workflow wrapped as an agent
func NewSequentialSubWorkflow(name string, config *WorkflowConfig) (Agent, error) {
	workflow, err := NewSequentialWorkflow(config)
	if err != nil {
		return nil, err
	}
	return NewSubWorkflowAgent(name, workflow), nil
}

// NewParallelSubWorkflow creates a parallel workflow wrapped as an agent
func NewParallelSubWorkflow(name string, config *WorkflowConfig) (Agent, error) {
	workflow, err := NewParallelWorkflow(config)
	if err != nil {
		return nil, err
	}
	return NewSubWorkflowAgent(name, workflow), nil
}

// NewDAGSubWorkflow creates a DAG workflow wrapped as an agent
func NewDAGSubWorkflow(name string, config *WorkflowConfig) (Agent, error) {
	workflow, err := NewDAGWorkflow(config)
	if err != nil {
		return nil, err
	}
	return NewSubWorkflowAgent(name, workflow), nil
}

// NewLoopSubWorkflow creates a loop workflow wrapped as an agent
func NewLoopSubWorkflow(name string, config *WorkflowConfig) (Agent, error) {
	workflow, err := NewLoopWorkflow(config)
	if err != nil {
		return nil, err
	}
	return NewSubWorkflowAgent(name, workflow), nil
}
