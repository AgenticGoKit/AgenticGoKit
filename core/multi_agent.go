// Package core provides multi-agent composition and orchestration capabilities
package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// MULTI-AGENT CONFIGURATION TYPES
// =============================================================================

// MultiAgentConfig provides configuration for multi-agent compositions
type MultiAgentConfig struct {
	Timeout        time.Duration
	MaxConcurrency int
	ErrorStrategy  ErrorHandlingStrategy
	StateStrategy  StateHandlingStrategy
}

// ErrorHandlingStrategy defines how errors are handled in multi-agent compositions
type ErrorHandlingStrategy string

const (
	ErrorStrategyFailFast   ErrorHandlingStrategy = "fail_fast"   // Stop on first error
	ErrorStrategyCollectAll ErrorHandlingStrategy = "collect_all" // Collect all errors
	ErrorStrategyContinue   ErrorHandlingStrategy = "continue"    // Ignore errors
)

// StateHandlingStrategy defines how state is managed in multi-agent compositions
type StateHandlingStrategy string

const (
	StateStrategyMerge     StateHandlingStrategy = "merge"     // Merge all states
	StateStrategyOverwrite StateHandlingStrategy = "overwrite" // Use last state
	StateStrategyIsolate   StateHandlingStrategy = "isolate"   // Keep states separate
)

// DefaultMultiAgentConfig returns sensible defaults for multi-agent configurations
func DefaultMultiAgentConfig() MultiAgentConfig {
	return MultiAgentConfig{
		Timeout:        30 * time.Second,
		MaxConcurrency: 10,
		ErrorStrategy:  ErrorStrategyCollectAll,
		StateStrategy:  StateStrategyMerge,
	}
}

// =============================================================================
// MULTI-AGENT CONSTRUCTORS
// =============================================================================

// NewParallelAgent creates an agent that runs sub-agents concurrently
// All sub-agents receive the same input state and their results are merged
func NewParallelAgent(name string, timeout time.Duration, subAgents ...Agent) Agent {
	return &parallelAgent{
		name:      name,
		timeout:   timeout,
		subAgents: subAgents,
	}
}

// NewParallelAgentWithConfig creates a parallel agent with full configuration options
func NewParallelAgentWithConfig(name string, config MultiAgentConfig, subAgents ...Agent) Agent {
	return &parallelAgent{
		name:      name,
		timeout:   config.Timeout,
		subAgents: subAgents,
		config:    config,
	}
}

// NewSequentialAgent creates an agent that runs sub-agents one after another
// Each sub-agent receives the output state from the previous agent
func NewSequentialAgent(name string, subAgents ...Agent) Agent {
	return &sequentialAgent{
		name:      name,
		subAgents: subAgents,
	}
}

// NewLoopAgent creates an agent that repeats a sub-agent with conditions
// The sub-agent continues executing until the condition returns true or max iterations reached
func NewLoopAgent(name string, maxIterations int, timeout time.Duration,
	condition func(State) bool, subAgent Agent) Agent {
	return &loopAgent{
		name:          name,
		maxIterations: maxIterations,
		timeout:       timeout,
		condition:     condition,
		subAgent:      subAgent,
	}
}

// =============================================================================
// AGENT COMPOSITION BUILDER PATTERN
// =============================================================================

// CompositionBuilder provides fluent interface for building multi-agent compositions
type CompositionBuilder struct {
	name       string
	agents     []Agent
	mode       string
	config     MultiAgentConfig
	loopConfig LoopConfig
}

// LoopConfig contains configuration specific to loop compositions
type LoopConfig struct {
	MaxIterations int
	Condition     func(State) bool
}

// NewComposition creates a new composition builder with the given name
func NewComposition(name string) *CompositionBuilder {
	return &CompositionBuilder{
		name:   name,
		agents: make([]Agent, 0),
		config: DefaultMultiAgentConfig(),
	}
}

// WithAgents adds one or more agents to the composition
func (cb *CompositionBuilder) WithAgents(agents ...Agent) *CompositionBuilder {
	cb.agents = append(cb.agents, agents...)
	return cb
}

// AsParallel configures the composition to run agents concurrently
func (cb *CompositionBuilder) AsParallel() *CompositionBuilder {
	cb.mode = "parallel"
	return cb
}

// AsSequential configures the composition to run agents one after another
func (cb *CompositionBuilder) AsSequential() *CompositionBuilder {
	cb.mode = "sequential"
	return cb
}

// AsLoop configures the composition to repeatedly run a single agent
// Requires exactly one agent in the composition
func (cb *CompositionBuilder) AsLoop(maxIterations int, condition func(State) bool) *CompositionBuilder {
	cb.mode = "loop"
	cb.loopConfig = LoopConfig{
		MaxIterations: maxIterations,
		Condition:     condition,
	}
	return cb
}

// WithTimeout sets the overall timeout for the composition
func (cb *CompositionBuilder) WithTimeout(timeout time.Duration) *CompositionBuilder {
	cb.config.Timeout = timeout
	return cb
}

// WithErrorStrategy sets how errors are handled during composition execution
func (cb *CompositionBuilder) WithErrorStrategy(strategy ErrorHandlingStrategy) *CompositionBuilder {
	cb.config.ErrorStrategy = strategy
	return cb
}

// WithStateStrategy sets how state is managed between agents in the composition
func (cb *CompositionBuilder) WithStateStrategy(strategy StateHandlingStrategy) *CompositionBuilder {
	cb.config.StateStrategy = strategy
	return cb
}

// WithMaxConcurrency sets the maximum number of concurrent agents (for parallel mode)
func (cb *CompositionBuilder) WithMaxConcurrency(max int) *CompositionBuilder {
	cb.config.MaxConcurrency = max
	return cb
}

// GenerateMermaidDiagramWithConfig generates a Mermaid diagram for this composition
func (cb *CompositionBuilder) GenerateMermaidDiagramWithConfig(config MermaidConfig) string {
	// Create a MermaidGenerator instance from internal package
	generator := NewMermaidGenerator()
	return generator.GenerateCompositionDiagram(cb.mode, cb.name, cb.agents, config)
}

// Build creates the composed agent based on the configuration
func (cb *CompositionBuilder) Build() (Agent, error) {
	if len(cb.agents) == 0 {
		return nil, fmt.Errorf("composition '%s' requires at least one agent", cb.name)
	}

	switch cb.mode {
	case "parallel":
		return NewParallelAgentWithConfig(cb.name, cb.config, cb.agents...), nil
	case "sequential":
		return NewSequentialAgent(cb.name, cb.agents...), nil
	case "loop":
		if len(cb.agents) != 1 {
			return nil, fmt.Errorf("loop composition '%s' requires exactly one agent, got %d", cb.name, len(cb.agents))
		}
		return NewLoopAgent(cb.name, cb.loopConfig.MaxIterations, cb.config.Timeout,
			cb.loopConfig.Condition, cb.agents[0]), nil
	case "":
		return nil, fmt.Errorf("composition '%s' mode not specified - use AsParallel(), AsSequential(), or AsLoop()", cb.name)
	default:
		return nil, fmt.Errorf("unknown composition mode '%s' for composition '%s'", cb.mode, cb.name)
	}
}

// =============================================================================
// CONVENIENCE FUNCTIONS
// =============================================================================

// CreateParallelWorkflow creates a parallel workflow with multiple agents
// This is a convenience function for common parallel agent patterns
func CreateParallelWorkflow(name string, timeout time.Duration, agents ...Agent) Agent {
	return NewParallelAgent(name, timeout, agents...)
}

// CreateSequentialWorkflow creates a sequential workflow with multiple agents
// This is a convenience function for common sequential agent patterns
func CreateSequentialWorkflow(name string, agents ...Agent) Agent {
	return NewSequentialAgent(name, agents...)
}

// CreateProcessingPipeline creates a common data processing pipeline
// Input -> Processing -> Output pattern
func CreateProcessingPipeline(name string, inputAgent, processingAgent, outputAgent Agent) Agent {
	return NewSequentialAgent(name, inputAgent, processingAgent, outputAgent)
}

// CreateParallelAnalysis creates a parallel analysis workflow
// Multiple analysis agents process the same input concurrently
func CreateParallelAnalysis(name string, timeout time.Duration, analysisAgents ...Agent) Agent {
	return NewParallelAgent(name, timeout, analysisAgents...)
}

// CreateConditionalLoop creates a loop agent with a simple condition function
// Convenience function for creating loop agents with common patterns
func CreateConditionalLoop(name string, maxIterations int, timeout time.Duration,
	conditionKey string, expectedValue interface{}, agent Agent) Agent {
	condition := func(state State) bool {
		if value, exists := state.Get(conditionKey); exists {
			return value == expectedValue
		}
		return false
	}
	return NewLoopAgent(name, maxIterations, timeout, condition, agent)
}

// CreateFanOutFanIn creates a pattern where input is processed by multiple agents and results are collected
// This is equivalent to parallel processing but with explicit naming for the pattern
func CreateFanOutFanIn(name string, timeout time.Duration, processors ...Agent) Agent {
	return NewParallelAgent(name, timeout, processors...)
}

// =============================================================================
// MULTI-AGENT IMPLEMENTATION TYPES
// =============================================================================

// parallelAgent runs sub-agents concurrently and merges their results
type parallelAgent struct {
	name      string
	timeout   time.Duration
	subAgents []Agent
	config    MultiAgentConfig
}

func (pa *parallelAgent) Name() string {
	return pa.name
}

func (pa *parallelAgent) Run(ctx context.Context, inputState State) (State, error) {
	if len(pa.subAgents) == 0 {
		return inputState.Clone(), nil
	}

	// Create context with timeout if specified
	runCtx := ctx
	if pa.timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, pa.timeout)
		defer cancel()
	}

	var wg sync.WaitGroup
	resultsChan := make(chan State, len(pa.subAgents))
	errChan := make(chan error, len(pa.subAgents))

	wg.Add(len(pa.subAgents))

	// Run all agents concurrently
	for i, agent := range pa.subAgents {
		go func(ag Agent, index int) {
			defer wg.Done()

			// Each agent gets a clone of the input state
			agentInput := inputState.Clone()
			result, err := ag.Run(runCtx, agentInput)

			if err != nil {
				errChan <- fmt.Errorf("agent %s (index %d): %w", ag.Name(), index, err)
				return
			}

			resultsChan <- result
		}(agent, i)
	}

	// Wait for all agents to complete
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errChan)
	}()

	// Collect results and errors
	mergedState := inputState.Clone()
	var errors []error

	// Collect all results
	for result := range resultsChan {
		// Merge state based on strategy
		switch pa.config.StateStrategy {
		case StateStrategyMerge:
			mergedState.Merge(result)
		case StateStrategyOverwrite:
			mergedState = result
		case StateStrategyIsolate:
			// Store in separate keys by agent name
			// This would need additional metadata to track which agent produced what
			mergedState.Merge(result)
		default:
			mergedState.Merge(result)
		}
	}

	// Collect all errors
	for err := range errChan {
		errors = append(errors, err)
	}

	// Handle errors based on strategy
	if len(errors) > 0 {
		switch pa.config.ErrorStrategy {
		case ErrorStrategyFailFast:
			return mergedState, errors[0]
		case ErrorStrategyCollectAll:
			return mergedState, fmt.Errorf("parallel execution errors: %v", errors)
		case ErrorStrategyContinue:
			// Continue with partial results
			return mergedState, nil
		default:
			return mergedState, fmt.Errorf("parallel execution errors: %v", errors)
		}
	}

	return mergedState, nil
}

// sequentialAgent runs sub-agents one after another
type sequentialAgent struct {
	name      string
	subAgents []Agent
}

func (sa *sequentialAgent) Name() string {
	return sa.name
}

func (sa *sequentialAgent) Run(ctx context.Context, inputState State) (State, error) {
	if len(sa.subAgents) == 0 {
		return inputState.Clone(), nil
	}

	currentState := inputState.Clone()

	// Run agents sequentially
	for i, agent := range sa.subAgents {
		result, err := agent.Run(ctx, currentState)
		if err != nil {
			return currentState, fmt.Errorf("agent %s (index %d): %w", agent.Name(), i, err)
		}
		currentState = result
	}

	return currentState, nil
}

// loopAgent repeats a sub-agent with conditions
type loopAgent struct {
	name          string
	maxIterations int
	timeout       time.Duration
	condition     func(State) bool
	subAgent      Agent
}

func (la *loopAgent) Name() string {
	return la.name
}

func (la *loopAgent) Run(ctx context.Context, inputState State) (State, error) {
	// Create context with timeout if specified
	runCtx := ctx
	if la.timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, la.timeout)
		defer cancel()
	}

	currentState := inputState.Clone()

	for i := 0; i < la.maxIterations; i++ {
		// Check if context is done
		select {
		case <-runCtx.Done():
			return currentState, fmt.Errorf("loop agent %s: context cancelled after %d iterations: %w", la.name, i, runCtx.Err())
		default:
		}

		// Check loop condition
		if la.condition != nil && la.condition(currentState) {
			break
		}

		// Run the sub-agent
		result, err := la.subAgent.Run(runCtx, currentState)
		if err != nil {
			return currentState, fmt.Errorf("loop agent %s iteration %d: %w", la.name, i, err)
		}

		currentState = result
	}

	return currentState, nil
}
