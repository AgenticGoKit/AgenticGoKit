// Package core provides orchestration capabilities for multi-agent systems
package core

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// ORCHESTRATOR INTERFACE
// =============================================================================

// Orchestrator defines the interface for routing events to agents.
type Orchestrator interface {
	// Dispatch the event to the appropriate agent.
	Dispatch(ctx context.Context, event Event) (AgentResult, error)
	// RegisterAgent registers a new agent with the given name and handler.
	RegisterAgent(name string, handler AgentHandler) error
	// GetCallbackRegistry returns the callback registry.
	GetCallbackRegistry() *CallbackRegistry
	// Stop halts the orchestrator.
	Stop()
}

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
	// OrchestrationMixed combines collaborative and sequential patterns in hybrid workflows
	OrchestrationMixed OrchestrationMode = "mixed"
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
	OrchestrationMode   OrchestrationMode   // Mode of orchestration (route, collaborate, mixed, etc.)
	Config              OrchestrationConfig // Orchestration-specific configuration
	CollaborativeAgents []string            // List of agent names for collaborative execution
	SequentialAgents    []string            // List of agent names for sequential execution
}

// =============================================================================
// ORCHESTRATION CONSTRUCTORS
// =============================================================================

// OrchestratorConfig holds configuration for creating orchestrators
type OrchestratorConfig struct {
	Type                    string   `json:"type" toml:"type"`
	AgentNames              []string `json:"agent_names,omitempty" toml:"agent_names,omitempty"`
	CollaborativeAgentNames []string `json:"collaborative_agents,omitempty" toml:"collaborative_agents,omitempty"`
	SequentialAgentNames    []string `json:"sequential_agents,omitempty" toml:"sequential_agents,omitempty"`
	MaxIterations           int      `json:"max_iterations,omitempty" toml:"max_iterations,omitempty"`
	TimeoutSeconds          int      `json:"timeout_seconds,omitempty" toml:"timeout_seconds,omitempty"`
}

// OrchestratorFactoryFunc is the function signature for creating orchestrators
type OrchestratorFactoryFunc func(config OrchestratorConfig, registry *CallbackRegistry) (Orchestrator, error)

// orchestratorFactory holds the registered factory function
var orchestratorFactory OrchestratorFactoryFunc

// RegisterOrchestratorFactory registers the orchestrator factory function
func RegisterOrchestratorFactory(factory OrchestratorFactoryFunc) {
	orchestratorFactory = factory
}

// NewOrchestrator creates an orchestrator from configuration
// This function requires the internal orchestrator factory to be registered
func NewOrchestrator(config OrchestratorConfig, registry *CallbackRegistry) (Orchestrator, error) {
	if orchestratorFactory == nil {
		return nil, fmt.Errorf("orchestrator factory not registered - import _ \"github.com/kunalkushwaha/agenticgokit/internal/orchestrator\" to register")
	}
	return orchestratorFactory(config, registry)
}

// =============================================================================
// ORCHESTRATOR FACTORY FUNCTIONS
// =============================================================================

// NewRouteOrchestrator creates a simple routing orchestrator.
// It requires the CallbackRegistry from the Runner.
// The implementation is provided by the internal/orchestrator package.
func NewRouteOrchestrator(registry *CallbackRegistry) Orchestrator {
	config := OrchestratorConfig{Type: "route"}
	orch, err := NewOrchestrator(config, registry)
	if err != nil {
		// This should not happen if the internal orchestrator package is properly imported
		Logger().Error().Err(err).Msg("Failed to create route orchestrator")
		return nil
	}
	return orch
}

// NewCollaborativeOrchestrator creates an orchestrator that runs all agents in parallel
// Each event is sent to ALL registered agents simultaneously
func NewCollaborativeOrchestrator(registry *CallbackRegistry) Orchestrator {
	config := OrchestratorConfig{Type: "collaborative"}
	orch, err := NewOrchestrator(config, registry)
	if err != nil {
		// Fallback to route orchestrator if factory not available
		return NewRouteOrchestrator(registry)
	}
	return orch
}

// NewSequentialOrchestrator creates an orchestrator that runs agents in sequence
func NewSequentialOrchestrator(registry *CallbackRegistry, agentNames []string) Orchestrator {
	config := OrchestratorConfig{
		Type:                 "sequential",
		SequentialAgentNames: agentNames,
	}
	orch, err := NewOrchestrator(config, registry)
	if err != nil {
		// Fallback to route orchestrator if factory not available
		return NewRouteOrchestrator(registry)
	}
	return orch
}

// NewLoopOrchestrator creates an orchestrator that runs a single agent in a loop
func NewLoopOrchestrator(registry *CallbackRegistry, agentNames []string) Orchestrator {
	config := OrchestratorConfig{
		Type:       "loop",
		AgentNames: agentNames,
	}
	orch, err := NewOrchestrator(config, registry)
	if err != nil {
		// Fallback to route orchestrator if factory not available
		return NewRouteOrchestrator(registry)
	}
	return orch
}

// NewMixedOrchestrator creates an orchestrator that combines collaborative and sequential execution
func NewMixedOrchestrator(registry *CallbackRegistry, collaborativeAgentNames, sequentialAgentNames []string) Orchestrator {
	config := OrchestratorConfig{
		Type:                    "mixed",
		CollaborativeAgentNames: collaborativeAgentNames,
		SequentialAgentNames:    sequentialAgentNames,
	}
	orch, err := NewOrchestrator(config, registry)
	if err != nil {
		// Fallback to route orchestrator if factory not available
		return NewRouteOrchestrator(registry)
	}
	return orch
}

// =============================================================================
// RUNNER INTEGRATION
// =============================================================================

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
	case OrchestrationMixed:
		orch = NewMixedOrchestrator(callbackRegistry, cfg.CollaborativeAgents, cfg.SequentialAgents)
	case OrchestrationSequential:
		orch = NewSequentialOrchestrator(callbackRegistry, cfg.SequentialAgents)
	case OrchestrationLoop:
		orch = NewLoopOrchestrator(callbackRegistry, cfg.SequentialAgents)
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
	// Ensure we have memory and sessionID to satisfy Runner requirements
	memory := QuickMemory()
	sessionID := GenerateSessionID()

	return NewRunnerWithOrchestration(EnhancedRunnerConfig{
		RunnerConfig: RunnerConfig{
			Agents:    ob.agents,
			Memory:    memory,
			SessionID: sessionID,
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

// Note: All orchestrator implementations have been moved to internal/orchestrator package
// This keeps the core package focused on interfaces and factory functions
