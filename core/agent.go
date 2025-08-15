// Package core provides essential agent interfaces and types for AgentFlow.
package core

import (
	"context"
	"time"
)

// =============================================================================
// CORE AGENT INTERFACES
// =============================================================================

// Agent defines the interface for any component that can process a State.
type Agent interface {
	// Run processes the input State and returns an output State or an error.
	// The context can be used for cancellation or deadlines.
	Run(ctx context.Context, inputState State) (State, error)
	// Name returns the unique identifier name of the agent.
	Name() string
}

// AgentHandler defines the interface for executing agent logic.
type AgentHandler interface {
	Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

// AgentHandlerFunc allows using a function as an AgentHandler.
type AgentHandlerFunc func(ctx context.Context, event Event, state State) (AgentResult, error)

func (f AgentHandlerFunc) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
	return f(ctx, event, state)
}

// ConfigAwareAgent represents an agent that can be configured from ResolvedAgentConfig
type ConfigAwareAgent interface {
	Agent
	// Configuration methods
	GetRole() string
	GetDescription() string
	GetSystemPrompt() string
	GetCapabilities() []string
	IsEnabled() bool
	GetTimeout() time.Duration
	GetLLMConfig() *ResolvedLLMConfig
	
	// Configuration update methods
	UpdateConfiguration(config *ResolvedAgentConfig) error
	ApplySystemPrompt(ctx context.Context, state State) (State, error)
}

// =============================================================================
// AGENT CAPABILITY FRAMEWORK
// =============================================================================

// AgentCapability represents a feature that can be added to any agent.
// Capabilities are composable and can be mixed and matched to create
// agents with different functionality.
type AgentCapability interface {
	// Name returns a unique identifier for this capability
	Name() string

	// Configure applies this capability to an agent during creation
	Configure(agent CapabilityConfigurable) error

	// Validate checks if this capability can be applied with others.
	// This allows capabilities to declare incompatibilities or dependencies.
	Validate(others []AgentCapability) error

	// Priority returns the initialization priority for this capability.
	// Lower numbers are initialized first. Used for dependency ordering.
	Priority() int
}

// CapabilityConfigurable represents an agent that can have capabilities configured on it.
type CapabilityConfigurable interface {
	// SetLLMProvider sets the LLM provider for the agent
	SetLLMProvider(provider ModelProvider, config AgentLLMConfig)

	// SetCacheManager sets the cache manager for the agent
	SetCacheManager(manager interface{}, config interface{})

	// SetMetricsConfig sets the metrics configuration for the agent
	SetMetricsConfig(config MetricsConfig)

	// GetLogger returns the agent's logger for capability configuration
	GetLogger() interface{} // Using interface{} to avoid zerolog dependency in core
}

// CapabilityType defines the type of capability for validation and ordering
type CapabilityType string

const (
	CapabilityTypeLLM     CapabilityType = "llm"
	CapabilityTypeMemory  CapabilityType = "memory"
	CapabilityTypeCache   CapabilityType = "cache"
	CapabilityTypeMetrics CapabilityType = "metrics"
	CapabilityTypeMCP     CapabilityType = "mcp"
	CapabilityTypeTracing CapabilityType = "tracing"
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

// LoopConfig provides configuration for loop-based agent compositions
type LoopConfig struct {
	MaxIterations int
	Timeout       time.Duration
	BreakCondition func(State) bool
}

// =============================================================================
// AGENT FACTORY INTERFACES
// =============================================================================

// ConfigurableAgentInterface defines the interface for configuration-driven agents
type ConfigurableAgentInterface interface {
	Agent
	GetRole() string
	GetCapabilities() []string
	GetSystemPrompt() string
	GetTimeout() time.Duration
	IsEnabled() bool
	GetLLMConfig() *ResolvedLLMConfig
	GetDescription() string
}

// AgentManager interface for managing agent lifecycle
type AgentManager interface {
	UpdateAgentConfigurations(config *Config) error
	GetCurrentAgents() map[string]Agent
	CreateAgent(name string, config *ResolvedAgentConfig) (Agent, error)
	DisableAgent(name string) error
}

// =============================================================================
// PUBLIC FACTORY FUNCTIONS
// =============================================================================

// DefaultMultiAgentConfig returns sensible defaults for multi-agent configurations
func DefaultMultiAgentConfig() MultiAgentConfig {
	return MultiAgentConfig{
		Timeout:        30 * time.Second,
		MaxConcurrency: 10,
		ErrorStrategy:  ErrorStrategyCollectAll,
		StateStrategy:  StateStrategyMerge,
	}
}

// NewConfigurableAgentFactory creates a new configurable agent factory
// Implementation is provided by internal packages
func NewConfigurableAgentFactory(config *Config) ConfigurableAgentFactory {
	if configurableAgentFactoryFunc != nil {
		return configurableAgentFactoryFunc(config)
	}
	return &basicConfigurableAgentFactory{config: config}
}

// NewAgentManager creates a new agent manager
// Implementation is provided by internal packages
func NewAgentManager(factory ConfigurableAgentFactory) AgentManager {
	if agentManagerFactory != nil {
		return agentManagerFactory(factory)
	}
	return &basicAgentManager{factory: factory}
}

// RegisterConfigurableAgentFactory registers the configurable agent factory function
func RegisterConfigurableAgentFactory(factory func(*Config) ConfigurableAgentFactory) {
	configurableAgentFactoryFunc = factory
}

// RegisterAgentManagerFactory registers the agent manager factory function
func RegisterAgentManagerFactory(factory func(ConfigurableAgentFactory) AgentManager) {
	agentManagerFactory = factory
}

// =============================================================================
// FACTORY INTERFACES
// =============================================================================

// ConfigurableAgentFactory creates agents based on resolved configuration
type ConfigurableAgentFactory interface {
	CreateAgent(name string, resolvedConfig *ResolvedAgentConfig, llmProvider ModelProvider) (Agent, error)
	CreateAgentFromConfig(name string, globalConfig *Config) (Agent, error)
	GetAgentCapabilities(name string) []string
}

// =============================================================================
// INTERNAL IMPLEMENTATIONS
// =============================================================================

var (
	configurableAgentFactoryFunc func(*Config) ConfigurableAgentFactory
	agentManagerFactory          func(ConfigurableAgentFactory) AgentManager
)

// Basic implementations for when internal packages aren't registered
type basicConfigurableAgentFactory struct {
	config *Config
}

func (f *basicConfigurableAgentFactory) CreateAgent(name string, resolvedConfig *ResolvedAgentConfig, llmProvider ModelProvider) (Agent, error) {
	Logger().Warn().Msg("Using basic agent factory - import internal/agents for full functionality")
	return &basicAgent{name: name}, nil
}

func (f *basicConfigurableAgentFactory) CreateAgentFromConfig(name string, globalConfig *Config) (Agent, error) {
	Logger().Warn().Msg("Using basic agent factory - import internal/agents for full functionality")
	return &basicAgent{name: name}, nil
}

func (f *basicConfigurableAgentFactory) GetAgentCapabilities(name string) []string {
	if f.config != nil {
		return f.config.GetAgentCapabilities(name)
	}
	return []string{}
}

type basicAgentManager struct {
	factory ConfigurableAgentFactory
}

func (am *basicAgentManager) UpdateAgentConfigurations(config *Config) error {
	Logger().Warn().Msg("Using basic agent manager - import internal/agents for full functionality")
	return nil
}

func (am *basicAgentManager) GetCurrentAgents() map[string]Agent {
	return make(map[string]Agent)
}

func (am *basicAgentManager) CreateAgent(name string, config *ResolvedAgentConfig) (Agent, error) {
	Logger().Warn().Msg("Using basic agent manager - import internal/agents for full functionality")
	return &basicAgent{name: name}, nil
}

func (am *basicAgentManager) DisableAgent(name string) error {
	Logger().Warn().Msg("Using basic agent manager - import internal/agents for full functionality")
	return nil
}

// basicAgent provides a minimal implementation
type basicAgent struct {
	name string
}

func (a *basicAgent) Run(ctx context.Context, inputState State) (State, error) {
	// Basic implementation - just return the input state
	return inputState, nil
}

func (a *basicAgent) Name() string {
	return a.name
}