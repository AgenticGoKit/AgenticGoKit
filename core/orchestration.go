// Package core provides orchestration capabilities for multi-agent systems
package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// ORCHESTRATION TYPES AND CONSTANTS
// =============================================================================

// OrchestrationMode defines how events are distributed to agents
type OrchestrationMode string

const (
	// OrchestrationRoute sends each event to a single agent based on routing metadata (default behavior)
	OrchestrationRoute OrchestrationMode = "route"
	// OrchestrationCollaborate sends each event to ALL registered agents in parallel
	OrchestrationCollaborate OrchestrationMode = "collaborate"
	// OrchestrationSequential processes agents one after another
	OrchestrationSequential OrchestrationMode = "sequential"
	// OrchestrationParallel processes agents in parallel (similar to collaborate but different semantics)
	OrchestrationParallel OrchestrationMode = "parallel"
	// OrchestrationLoop repeats processing with a single agent
	OrchestrationLoop OrchestrationMode = "loop"
)

// OrchestrationConfig contains configuration for orchestration behavior
type OrchestrationConfig struct {
	Timeout          time.Duration // Overall timeout for orchestration operations
	MaxConcurrency   int           // Maximum number of concurrent agent executions
	FailureThreshold float64       // Percentage of failures before stopping (0.0-1.0)
	RetryPolicy      *RetryPolicy  // Policy for retrying failed operations (uses existing RetryPolicy from core)
}

// DefaultOrchestrationConfig returns sensible defaults for orchestration configuration
func DefaultOrchestrationConfig() OrchestrationConfig {
	return OrchestrationConfig{
		Timeout:          30 * time.Second,
		MaxConcurrency:   10,
		FailureThreshold: 0.5, // Stop if 50% of agents fail
		RetryPolicy:      DefaultRetryPolicy(),
	}
}

// =============================================================================
// ENHANCED RUNNER CONFIGURATION
// =============================================================================

// EnhancedRunnerConfig extends RunnerConfig with orchestration options
type EnhancedRunnerConfig struct {
	RunnerConfig
	OrchestrationMode OrchestrationMode   // Mode of orchestration (route or collaborate)
	Config            OrchestrationConfig // Orchestration-specific configuration
}

// =============================================================================
// ORCHESTRATION CONSTRUCTORS
// =============================================================================

// NewCollaborativeOrchestrator creates an orchestrator that runs all agents in parallel
// Each event is sent to ALL registered agents simultaneously
func NewCollaborativeOrchestrator(registry *CallbackRegistry) Orchestrator {
	return &collaborativeOrchestrator{
		handlers:         make(map[string]AgentHandler),
		callbackRegistry: registry,
	}
}

// NewRunnerWithOrchestration creates a runner with specified orchestration mode
func NewRunnerWithOrchestration(cfg EnhancedRunnerConfig) Runner {
	// Create base runner with standard configuration
	runner := NewRunnerWithConfig(cfg.RunnerConfig)

	// Override orchestrator based on mode
	callbackRegistry := runner.GetCallbackRegistry()

	var orch Orchestrator
	switch cfg.OrchestrationMode {
	case OrchestrationCollaborate:
		orch = NewCollaborativeOrchestrator(callbackRegistry)
	default:
		orch = NewRouteOrchestrator(callbackRegistry)
	}

	// Set the orchestrator on the runner
	if runnerImpl, ok := runner.(*RunnerImpl); ok {
		runnerImpl.SetOrchestrator(orch)

		// Re-register all agents with the new orchestrator since SetOrchestrator replaces it
		for name, agent := range cfg.RunnerConfig.Agents {
			if err := orch.RegisterAgent(name, agent); err != nil {
				Logger().Error().Str("agent", name).Err(err).Msg("Failed to register agent with new orchestrator")
			}
		}

		// Re-register the default error handler if it wasn't provided
		if _, exists := cfg.RunnerConfig.Agents["error-handler"]; !exists {
			orch.RegisterAgent("error-handler", AgentHandlerFunc(
				func(ctx context.Context, event Event, state State) (AgentResult, error) {
					state.SetMeta(RouteMetadataKey, "")
					return AgentResult{OutputState: state}, nil
				},
			))
		}
	}

	return runner
}

// =============================================================================
// ORCHESTRATION BUILDER PATTERN
// =============================================================================

// OrchestrationBuilder provides fluent interface for orchestration setup
type OrchestrationBuilder struct {
	mode   OrchestrationMode
	agents map[string]AgentHandler
	config OrchestrationConfig
}

// NewOrchestrationBuilder creates a new orchestration builder with the specified mode
func NewOrchestrationBuilder(mode OrchestrationMode) *OrchestrationBuilder {
	return &OrchestrationBuilder{
		mode:   mode,
		agents: make(map[string]AgentHandler),
		config: DefaultOrchestrationConfig(),
	}
}

// WithAgent adds a single agent to the orchestration
func (ob *OrchestrationBuilder) WithAgent(name string, handler AgentHandler) *OrchestrationBuilder {
	ob.agents[name] = handler
	return ob
}

// WithAgents adds multiple agents to the orchestration from a map
func (ob *OrchestrationBuilder) WithAgents(agents map[string]AgentHandler) *OrchestrationBuilder {
	for name, handler := range agents {
		ob.agents[name] = handler
	}
	return ob
}

// WithTimeout sets the orchestration timeout
func (ob *OrchestrationBuilder) WithTimeout(timeout time.Duration) *OrchestrationBuilder {
	ob.config.Timeout = timeout
	return ob
}

// WithMaxConcurrency sets the maximum number of concurrent agents
func (ob *OrchestrationBuilder) WithMaxConcurrency(max int) *OrchestrationBuilder {
	ob.config.MaxConcurrency = max
	return ob
}

// WithFailureThreshold sets the failure threshold (0.0-1.0)
// When this percentage of agents fail, the orchestration will stop
func (ob *OrchestrationBuilder) WithFailureThreshold(threshold float64) *OrchestrationBuilder {
	ob.config.FailureThreshold = threshold
	return ob
}

// WithRetryPolicy sets the retry policy for failed agents
func (ob *OrchestrationBuilder) WithRetryPolicy(policy *RetryPolicy) *OrchestrationBuilder {
	ob.config.RetryPolicy = policy
	return ob
}

// WithConfig sets the complete orchestration configuration
func (ob *OrchestrationBuilder) WithConfig(config OrchestrationConfig) *OrchestrationBuilder {
	ob.config = config
	return ob
}

// Build creates the configured runner with the specified orchestration mode
func (ob *OrchestrationBuilder) Build() Runner {
	return NewRunnerWithOrchestration(EnhancedRunnerConfig{
		RunnerConfig: RunnerConfig{
			Agents: ob.agents,
		},
		OrchestrationMode: ob.mode,
		Config:            ob.config,
	})
}

// =============================================================================
// CONVENIENCE FUNCTIONS
// =============================================================================

// CreateCollaborativeRunner creates a runner where all agents process events in parallel
// Each event is sent to ALL registered agents simultaneously
func CreateCollaborativeRunner(agents map[string]AgentHandler, timeout time.Duration) Runner {
	return NewOrchestrationBuilder(OrchestrationCollaborate).
		WithAgents(agents).
		WithTimeout(timeout).
		Build()
}

// CreateRouteRunner creates a standard routing runner (existing behavior)
// Each event is sent to a single agent based on routing metadata
func CreateRouteRunner(agents map[string]AgentHandler) Runner {
	return NewOrchestrationBuilder(OrchestrationRoute).
		WithAgents(agents).
		Build()
}

// CreateHighThroughputRunner creates a collaborative runner optimized for high throughput
// Uses higher concurrency limits and more tolerant failure thresholds
func CreateHighThroughputRunner(agents map[string]AgentHandler) Runner {
	return NewOrchestrationBuilder(OrchestrationCollaborate).
		WithAgents(agents).
		WithMaxConcurrency(50).
		WithTimeout(60 * time.Second).
		WithFailureThreshold(0.8). // Tolerate 80% failures
		Build()
}

// CreateFaultTolerantRunner creates a collaborative runner with aggressive retry policies
// Designed for environments where transient failures are common
func CreateFaultTolerantRunner(agents map[string]AgentHandler) Runner {
	retryPolicy := RetryPolicy{
		MaxRetries:    5,
		BackoffFactor: 1.5,
		MaxDelay:      30 * time.Second,
	}

	return NewOrchestrationBuilder(OrchestrationCollaborate).
		WithAgents(agents).
		WithRetryPolicy(&retryPolicy).
		WithFailureThreshold(0.9). // Very tolerant of failures
		Build()
}

// CreateLoadBalancedRunner creates a runner that distributes load across multiple agent instances
// Useful for scaling horizontally with multiple instances of the same agent type
func CreateLoadBalancedRunner(agents map[string]AgentHandler, maxConcurrency int) Runner {
	return NewOrchestrationBuilder(OrchestrationRoute).
		WithAgents(agents).
		WithMaxConcurrency(maxConcurrency).
		WithTimeout(30 * time.Second).
		Build()
}

// =============================================================================
// ORCHESTRATION UTILITIES
// =============================================================================

// ConvertAgentToHandler converts an Agent to an AgentHandler for use in orchestration
// This is a utility function to bridge between Agent and AgentHandler interfaces
func ConvertAgentToHandler(agent Agent) AgentHandler {
	return AgentHandlerFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
		// Merge event data into state
		if eventData := event.GetData(); eventData != nil {
			for key, value := range eventData {
				state.Set(key, value)
			}
		}

		// Run the agent
		outputState, err := agent.Run(ctx, state)
		if err != nil {
			return AgentResult{Error: err.Error()}, err
		}

		return AgentResult{OutputState: outputState}, nil
	})
}

// CreateMixedOrchestration creates a runner that combines multiple orchestration patterns
// This allows for complex workflows where different agent groups use different orchestration modes
func CreateMixedOrchestration(routeAgents, collaborativeAgents map[string]AgentHandler) Runner {
	// Combine all agents into a single map
	allAgents := make(map[string]AgentHandler)
	for name, handler := range routeAgents {
		allAgents[name] = handler
	}
	for name, handler := range collaborativeAgents {
		allAgents[name] = handler
	}

	// For now, use route orchestration as the base
	// In the future, this could be enhanced to support mixed modes
	return CreateRouteRunner(allAgents)
}

// =============================================================================
// COLLABORATIVE ORCHESTRATOR IMPLEMENTATION
// =============================================================================

// collaborativeOrchestrator implements the Orchestrator interface for collaborative mode
type collaborativeOrchestrator struct {
	handlers         map[string]AgentHandler
	callbackRegistry *CallbackRegistry
	mu               sync.RWMutex
}

// RegisterAgent adds an agent handler to the collaborative orchestrator
func (o *collaborativeOrchestrator) RegisterAgent(name string, handler AgentHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.handlers[name]; exists {
		return fmt.Errorf("agent with name '%s' already registered", name)
	}

	o.handlers[name] = handler
	Logger().Info().
		Str("agent", name).
		Msg("CollaborativeOrchestrator: Registered agent")
	return nil
}

// GetCallbackRegistry returns the callback registry
func (o *collaborativeOrchestrator) GetCallbackRegistry() *CallbackRegistry {
	return o.callbackRegistry
}

// Stop halts the orchestrator
func (o *collaborativeOrchestrator) Stop() {
	// Implementation for stopping the orchestrator
	Logger().Info().Msg("CollaborativeOrchestrator: Stopped")
}

// Dispatch sends the event to all registered agent handlers concurrently
func (o *collaborativeOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
	if event == nil {
		Logger().Warn().Msg("CollaborativeOrchestrator: Received nil event, skipping dispatch.")
		err := errors.New("cannot dispatch nil event")
		return AgentResult{Error: err.Error()}, err
	}

	o.mu.RLock()
	defer o.mu.RUnlock()

	if len(o.handlers) == 0 {
		Logger().Warn().Msg("CollaborativeOrchestrator: No agents registered")
		err := errors.New("no agents registered")
		return AgentResult{Error: err.Error()}, err
	}

	// Create a channel to collect results from all agents
	resultChan := make(chan AgentResult, len(o.handlers))
	var wg sync.WaitGroup

	// Extract the current state from the event data
	currentState := NewState()
	if event != nil {
		// Create state from event data - using event metadata and data
		for k, v := range event.GetMetadata() {
			currentState.SetMeta(k, v)
		}
		// Add event data to state
		eventData := event.GetData()
		for key, value := range eventData {
			currentState.Set(key, value)
		}
	}

	// Launch goroutines for each agent handler
	for name, handler := range o.handlers {
		wg.Add(1)
		go func(agentName string, h AgentHandler) {
			defer wg.Done()
			Logger().Debug().
				Str("agent", agentName).
				Msg("CollaborativeOrchestrator: Dispatching to agent")

			result, err := h.Run(ctx, event, currentState)
			if err != nil {
				result.Error = err.Error()
			}
			resultChan <- result
		}(name, handler)
	}

	// Wait for all agents to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all results
	var results []AgentResult
	var errors []string
	hasSuccess := false
	combinedState := NewState()

	for result := range resultChan {
		results = append(results, result)
		if result.Error != "" {
			errors = append(errors, result.Error)
		} else {
			hasSuccess = true
			// Merge output states (iterate through keys since State is an interface)
			for _, key := range result.OutputState.Keys() {
				if value, ok := result.OutputState.Get(key); ok {
					combinedState.Set(key, value)
				}
			}
			// Also merge metadata
			for _, key := range result.OutputState.MetaKeys() {
				if value, ok := result.OutputState.GetMeta(key); ok {
					combinedState.SetMeta(key, value)
				}
			}
		}
	}

	// Create combined result
	combinedResult := AgentResult{
		OutputState: combinedState,
		StartTime:   time.Now(),
		EndTime:     time.Now(),
	}

	// If all agents failed, return error
	if !hasSuccess {
		combinedResult.Error = fmt.Sprintf("all agents failed: %v", errors)
		return combinedResult, fmt.Errorf("collaborative dispatch failed: all agents returned errors")
	}

	Logger().Info().
		Int("total_agents", len(o.handlers)).
		Int("successful", len(results)-len(errors)).
		Int("failed", len(errors)).
		Msg("CollaborativeOrchestrator: Dispatch completed")

	return combinedResult, nil
}

// =============================================================================
// ORCHESTRATION CONSTRUCTORS (UPDATED)
// =============================================================================
