// Package core provides essential agent interfaces and types for AgentFlow.
// The Agent interface is the primary public API for all agent implementations.
package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// SECTION 1: CORE INTERFACES (Public API Contracts)
// =============================================================================

// Agent defines the unified interface for any component that can process state and events.
// This interface combines both state processing and event-driven execution patterns.
type Agent interface {
	// Core identification
	Name() string
	GetRole() string
	GetDescription() string

	// Execution capabilities - unified interface supporting both patterns
	Run(ctx context.Context, inputState State) (State, error)                       // State processing pattern
	HandleEvent(ctx context.Context, event Event, state State) (AgentResult, error) // Event-driven pattern

	// Configuration and lifecycle
	GetCapabilities() []string
	GetSystemPrompt() string
	GetTimeout() time.Duration
	IsEnabled() bool
	GetLLMConfig() *ResolvedLLMConfig

	// Lifecycle management
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// AgentHandler is now just an alias for backward compatibility during migration
// TODO: Remove in next major version - use Agent.HandleEvent directly
type AgentHandler interface {
	Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

// AgentHandlerFunc allows using a function as an AgentHandler.
type AgentHandlerFunc func(ctx context.Context, event Event, state State) (AgentResult, error)

func (f AgentHandlerFunc) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
	return f(ctx, event, state)
}

// ConfigurableAgentInterface is now an alias for Agent since all agents are configurable
// TODO: Remove in next major version - use Agent directly
type ConfigurableAgentInterface = Agent

// AgentManager interface for managing agent lifecycle
type AgentManager interface {
	UpdateAgentConfigurations(config *Config) error
	GetCurrentAgents() map[string]Agent
	CreateAgent(name string, config *ResolvedAgentConfig) (Agent, error)
	DisableAgent(name string) error

	// Additional methods expected by public API
	InitializeAgents() error
	GetActiveAgents() []Agent // Updated to use unified Agent interface
}

// ConfigurableAgentFactory creates agents based on resolved configuration
type ConfigurableAgentFactory interface {
	CreateAgent(name string, resolvedConfig *ResolvedAgentConfig, llmProvider ModelProvider) (Agent, error)
	CreateAgentFromConfig(name string, globalConfig *Config) (Agent, error)
	GetAgentCapabilities(name string) []string
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
	SetLLMProvider(provider ModelProvider, config LLMConfig)

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

// Multi-Agent Configuration Types
type MultiAgentConfig struct {
	Timeout        time.Duration
	MaxConcurrency int
	ErrorStrategy  ErrorHandlingStrategy
	StateStrategy  StateHandlingStrategy
}

type ErrorHandlingStrategy string

const (
	ErrorStrategyFailFast   ErrorHandlingStrategy = "fail_fast"   // Stop on first error
	ErrorStrategyCollectAll ErrorHandlingStrategy = "collect_all" // Collect all errors
	ErrorStrategyContinue   ErrorHandlingStrategy = "continue"    // Ignore errors
)

type StateHandlingStrategy string

const (
	StateStrategyMerge     StateHandlingStrategy = "merge"     // Merge all states
	StateStrategyOverwrite StateHandlingStrategy = "overwrite" // Use last state
	StateStrategyIsolate   StateHandlingStrategy = "isolate"   // Keep states separate
)

type LoopConfig struct {
	MaxIterations  int
	Timeout        time.Duration
	BreakCondition func(State) bool
}

// =============================================================================
// SECTION 2: PUBLIC FACTORY FUNCTIONS (User Entry Points)
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

// NewAgentManager creates a new agent manager from configuration
// Implementation is provided by internal packages via registration
func NewAgentManager(config *Config) AgentManager {
	if config == nil {
		config = &Config{}
	}

	if agentManagerFactory != nil {
		manager := agentManagerFactory(config)
		Logger().Debug().
			Str("implementation", "enhanced").
			Msg("UNIQUE_MARKER_Created agent manager using registered enhanced implementation")
		return manager
	}

	Logger().Debug().
		Str("implementation", "core").
		Msg("Using core agent manager implementation")

	// Use core basic implementation
	return &basicAgentManager{
		config: config,
		agents: make(map[string]Agent),
	}
}

// NewConfigurableAgentFactory creates a new configurable agent factory
// Implementation is provided by internal packages via registration
func NewConfigurableAgentFactory(config *Config) ConfigurableAgentFactory {
	if config == nil {
		config = &Config{}
	}

	if configurableAgentFactoryFunc != nil {
		factory := configurableAgentFactoryFunc(config)
		Logger().Debug().
			Str("implementation", "enhanced").
			Msg("Created configurable agent factory using registered enhanced implementation")
		return factory
	}

	Logger().Debug().
		Str("implementation", "core").
		Msg("Using core agent factory implementation")

	// Use core basic implementation
	return &basicConfigurableAgentFactory{
		config: config,
	}
}

// =============================================================================
// SECTION 3: REGISTRATION SYSTEM (Internal Package Integration)
// =============================================================================

var (
	// agentManagerFactory holds the registered agent manager factory function
	agentManagerFactory func(*Config) AgentManager

	// configurableAgentFactoryFunc holds the registered configurable agent factory function
	configurableAgentFactoryFunc func(*Config) ConfigurableAgentFactory
)

// RegisterAgentManagerFactory registers an agent manager factory function
// This allows internal packages to provide full implementations
func RegisterAgentManagerFactory(factory func(*Config) AgentManager) {
	agentManagerFactory = factory
	Logger().Debug().Msg("Registered full-featured agent manager implementation")
}

// RegisterConfigurableAgentFactory registers a configurable agent factory function
// This allows internal packages to provide full implementations
func RegisterConfigurableAgentFactory(factory func(*Config) ConfigurableAgentFactory) {
	configurableAgentFactoryFunc = factory
	Logger().Debug().Msg("Registered full-featured configurable agent factory implementation")
}

// validateRegistrations logs information about which implementations are available
func validateRegistrations() {
	if agentManagerFactory == nil {
		Logger().Debug().Msg("No agent manager factory registered - using basic implementation")
	} else {
		Logger().Debug().Msg("Full agent manager factory is registered")
	}

	if configurableAgentFactoryFunc == nil {
		Logger().Debug().Msg("No configurable agent factory registered - using basic implementation")
	} else {
		Logger().Debug().Msg("Full configurable agent factory is registered")
	}
}

// =============================================================================
// SECTION 4: BASIC IMPLEMENTATIONS (Core Functionality)
// =============================================================================

// These basic implementations provide full working functionality out of the box.
// Internal packages can enhance these with additional features via the registration system.

// basicAgentManager provides a complete, working agent manager implementation
type basicAgentManager struct {
	config *Config
	agents map[string]Agent
	mutex  sync.RWMutex
}

// basicConfigurableAgentFactory provides a complete, working factory implementation
type basicConfigurableAgentFactory struct {
	config *Config
}

// basicAgent provides a minimal but complete Agent implementation
type basicAgent struct {
	name string
}

// basicConfigurableAgent provides a complete configurable Agent implementation
type basicConfigurableAgent struct {
	name   string
	role   string
	config *Config
}

// Implementation methods for basicAgentManager
func (am *basicAgentManager) UpdateAgentConfigurations(config *Config) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.config = config
	return nil
}

func (am *basicAgentManager) GetCurrentAgents() map[string]Agent {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	agents := make(map[string]Agent, len(am.agents))
	for name, agent := range am.agents {
		agents[name] = agent
	}
	return agents
}

func (am *basicAgentManager) CreateAgent(name string, config *ResolvedAgentConfig) (Agent, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	agent := &basicAgent{name: name}
	am.agents[name] = agent
	return agent, nil
}

func (am *basicAgentManager) DisableAgent(name string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	delete(am.agents, name)
	return nil
}

func (am *basicAgentManager) InitializeAgents() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	if am.config != nil && len(am.config.Agents) > 0 {
		for agentName := range am.config.Agents {
			if _, exists := am.agents[agentName]; !exists {
				am.agents[agentName] = &basicAgent{name: agentName}
			}
		}
	}
	return nil
}

func (am *basicAgentManager) GetActiveAgents() []Agent {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	agents := make([]Agent, 0, len(am.agents))
	for name := range am.agents {
		role := "basic_processor"
		if am.config != nil {
			if agentConfig, exists := am.config.Agents[name]; exists {
				if agentConfig.Role != "" {
					role = agentConfig.Role
				}
			}
		}
		agents = append(agents, &basicConfigurableAgent{
			name:   name,
			role:   role,
			config: am.config,
		})
	}
	return agents
}

// Implementation methods for basicConfigurableAgentFactory
func (f *basicConfigurableAgentFactory) CreateAgent(name string, resolvedConfig *ResolvedAgentConfig, llmProvider ModelProvider) (Agent, error) {
	return &basicAgent{name: name}, nil
}

func (f *basicConfigurableAgentFactory) CreateAgentFromConfig(name string, globalConfig *Config) (Agent, error) {
	return &basicAgent{name: name}, nil
}

func (f *basicConfigurableAgentFactory) GetAgentCapabilities(name string) []string {
	if f.config != nil {
		return f.config.GetAgentCapabilities(name)
	}
	return []string{"basic_processing"}
}

// Implementation methods for basicAgent
func (a *basicAgent) Run(ctx context.Context, inputState State) (State, error) {
	result := inputState.Clone()
	result.Set("processed_by", a.name)
	result.Set("agent_type", "basic")
	result.Set("processed_at", time.Now().Unix())
	return result, nil
}

func (a *basicAgent) HandleEvent(ctx context.Context, event Event, state State) (AgentResult, error) {
	startTime := time.Now()
	outputState, err := a.Run(ctx, state)
	endTime := time.Now()

	result := AgentResult{
		OutputState: outputState,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(startTime),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

func (a *basicAgent) Name() string {
	return a.name
}

func (a *basicAgent) GetRole() string {
	return "basic_processor"
}

func (a *basicAgent) GetDescription() string {
	return fmt.Sprintf("Basic agent: %s", a.name)
}

func (a *basicAgent) GetCapabilities() []string {
	return []string{"basic_processing"}
}

func (a *basicAgent) GetSystemPrompt() string {
	return "You are a basic processing agent."
}

func (a *basicAgent) GetTimeout() time.Duration {
	return 30 * time.Second
}

func (a *basicAgent) IsEnabled() bool {
	return true
}

func (a *basicAgent) GetLLMConfig() *ResolvedLLMConfig {
	return nil
}

func (a *basicAgent) Initialize(ctx context.Context) error {
	return nil
}

func (a *basicAgent) Shutdown(ctx context.Context) error {
	return nil
}

// Implementation methods for basicConfigurableAgent
func (a *basicConfigurableAgent) Run(ctx context.Context, inputState State) (State, error) {
	result := inputState.Clone()
	result.Set("processed_by", a.name)
	result.Set("agent_role", a.role)
	result.Set("agent_type", "basic_configurable")
	result.Set("processed_at", time.Now().Unix())
	result.Set("response", fmt.Sprintf("Processed by basic agent %s with role %s", a.name, a.role))
	return result, nil
}

func (a *basicConfigurableAgent) HandleEvent(ctx context.Context, event Event, state State) (AgentResult, error) {
	startTime := time.Now()
	outputState, err := a.Run(ctx, state)
	endTime := time.Now()

	result := AgentResult{
		OutputState: outputState,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(startTime),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

func (a *basicConfigurableAgent) Name() string {
	return a.name
}

func (a *basicConfigurableAgent) GetRole() string {
	return a.role
}

func (a *basicConfigurableAgent) GetDescription() string {
	return fmt.Sprintf("Basic configurable agent: %s (role: %s)", a.name, a.role)
}

func (a *basicConfigurableAgent) GetSystemPrompt() string {
	if a.config != nil {
		if agentConfig, exists := a.config.Agents[a.name]; exists {
			if agentConfig.SystemPrompt != "" {
				return agentConfig.SystemPrompt
			}
		}
	}
	return "You are a helpful assistant."
}

func (a *basicConfigurableAgent) GetCapabilities() []string {
	if a.config != nil {
		if agentConfig, exists := a.config.Agents[a.name]; exists {
			if len(agentConfig.Capabilities) > 0 {
				return agentConfig.Capabilities
			}
		}
	}
	return []string{"basic_processing"}
}

func (a *basicConfigurableAgent) IsEnabled() bool {
	if a.config != nil {
		if agentConfig, exists := a.config.Agents[a.name]; exists {
			return agentConfig.Enabled
		}
	}
	return true
}

func (a *basicConfigurableAgent) GetTimeout() time.Duration {
	if a.config != nil {
		if agentConfig, exists := a.config.Agents[a.name]; exists {
			if agentConfig.Timeout > 0 {
				return time.Duration(agentConfig.Timeout) * time.Second
			}
		}
	}
	return 30 * time.Second
}

func (a *basicConfigurableAgent) GetLLMConfig() *ResolvedLLMConfig {
	return nil
}

func (a *basicConfigurableAgent) Initialize(ctx context.Context) error {
	return nil
}

func (a *basicConfigurableAgent) Shutdown(ctx context.Context) error {
	return nil
}

// =============================================================================
// SECTION 5: COMPOSABLE AGENT IMPLEMENTATION (Advanced)
// =============================================================================

// ComposableAgent represents a production-ready agent that supports all capabilities
// through a composable, capability-based architecture. This is the recommended
// implementation for complex agents that need advanced features like capability
// composition, pre/post processing, and runtime configuration.
type ComposableAgent struct {
	name         string
	role         string
	description  string
	systemPrompt string
	timeout      time.Duration
	enabled      bool
	llmConfig    *ResolvedLLMConfig
	capabilities map[CapabilityType]AgentCapability
	handler      AgentHandler
	config       *ResolvedAgentConfig
}

// NewComposableAgent creates a new composable agent with the given name and capabilities.
// This is the recommended constructor for agents that need advanced capability management.
func NewComposableAgent(name string, capabilities map[CapabilityType]AgentCapability, handler AgentHandler) *ComposableAgent {
	if capabilities == nil {
		capabilities = make(map[CapabilityType]AgentCapability)
	}

	return &ComposableAgent{
		name:         name,
		role:         "composable_agent",
		description:  fmt.Sprintf("Composable agent: %s", name),
		systemPrompt: "You are a capable AI agent with composable capabilities.",
		timeout:      30 * time.Second,
		enabled:      true,
		capabilities: capabilities,
		handler:      handler,
	}
}

// NewComposableAgentWithConfig creates a new composable agent with full configuration.
// This constructor provides complete control over agent behavior and is suitable
// for production deployments.
func NewComposableAgentWithConfig(name string, config *ResolvedAgentConfig, capabilities map[CapabilityType]AgentCapability, handler AgentHandler) *ComposableAgent {
	if capabilities == nil {
		capabilities = make(map[CapabilityType]AgentCapability)
	}

	agent := &ComposableAgent{
		name:         name,
		capabilities: capabilities,
		handler:      handler,
		config:       config,
		enabled:      true,
	}

	// Apply configuration if provided
	if config != nil {
		agent.role = config.Role
		agent.description = config.Description
		agent.systemPrompt = config.SystemPrompt
		agent.timeout = time.Duration(config.Timeout) * time.Second
		agent.enabled = config.Enabled
		agent.llmConfig = config.LLMConfig
	}

	// Set defaults
	if agent.role == "" {
		agent.role = "composable_agent"
	}
	if agent.description == "" {
		agent.description = fmt.Sprintf("Composable agent: %s", name)
	}
	if agent.systemPrompt == "" {
		agent.systemPrompt = "You are a capable AI agent with composable capabilities."
	}
	if agent.timeout <= 0 {
		agent.timeout = 30 * time.Second
	}

	return agent
}

// Implement Agent interface

func (c *ComposableAgent) Name() string {
	return c.name
}

func (c *ComposableAgent) GetRole() string {
	return c.role
}

func (c *ComposableAgent) GetDescription() string {
	return c.description
}

func (c *ComposableAgent) GetCapabilities() []string {
	capabilities := make([]string, 0, len(c.capabilities))
	for capType := range c.capabilities {
		capabilities = append(capabilities, string(capType))
	}
	return capabilities
}

func (c *ComposableAgent) GetSystemPrompt() string {
	return c.systemPrompt
}

func (c *ComposableAgent) GetTimeout() time.Duration {
	return c.timeout
}

func (c *ComposableAgent) IsEnabled() bool {
	return c.enabled
}

func (c *ComposableAgent) GetLLMConfig() *ResolvedLLMConfig {
	return c.llmConfig
}

func (c *ComposableAgent) Initialize(ctx context.Context) error {
	Logger().Debug().
		Str("agent", c.name).
		Int("capabilities", len(c.capabilities)).
		Msg("Initializing ComposableAgent")

	// Initialize all capabilities
	for capType, capability := range c.capabilities {
		if initializer, ok := capability.(interface {
			Initialize(context.Context) error
		}); ok {
			if err := initializer.Initialize(ctx); err != nil {
				return fmt.Errorf("failed to initialize capability %s: %w", capType, err)
			}
		}
	}

	return nil
}

func (c *ComposableAgent) Shutdown(ctx context.Context) error {
	Logger().Debug().
		Str("agent", c.name).
		Msg("Shutting down ComposableAgent")

	// Shutdown all capabilities
	var errors []error
	for capType, capability := range c.capabilities {
		if shutdowner, ok := capability.(interface {
			Shutdown(context.Context) error
		}); ok {
			if err := shutdowner.Shutdown(ctx); err != nil {
				errors = append(errors, fmt.Errorf("failed to shutdown capability %s: %w", capType, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

func (c *ComposableAgent) Run(ctx context.Context, state State) (State, error) {
	Logger().Debug().
		Str("agent", c.name).
		Int("capabilities", len(c.capabilities)).
		Msg("ComposableAgent starting execution")

	// Clone the input state to avoid mutations
	workingState := state.Clone()

	// Pre-execution: Apply capability pre-processing
	var err error
	workingState, err = c.applyCapabilityPreProcessing(ctx, workingState)
	if err != nil {
		Logger().Error().
			Err(err).
			Str("agent", c.name).
			Msg("Capability pre-processing failed")
		return state, fmt.Errorf("capability pre-processing failed: %w", err)
	}

	// Execute the core agent logic
	var result State
	if c.handler != nil {
		// Use custom handler if provided
		agentResult, err := c.handler.Run(ctx, NewEvent(c.name, map[string]any{}, map[string]string{}), workingState)
		if err != nil {
			Logger().Error().
				Err(err).
				Str("agent", c.name).
				Msg("Agent handler execution failed")
			return state, fmt.Errorf("agent handler execution failed: %w", err)
		}
		result = agentResult.OutputState
	} else {
		// Default behavior: add processed metadata
		result = workingState.Clone()
		result.Set("processed_by", c.name)
		result.Set("agent_type", "composable")
		result.Set("processed_at", time.Now().Unix())

		// Add capability metadata
		result.Set("capabilities", c.GetCapabilities())
	}

	// Post-execution: Apply capability post-processing
	finalState, err := c.applyCapabilityPostProcessing(ctx, result)
	if err != nil {
		Logger().Error().
			Err(err).
			Str("agent", c.name).
			Msg("Capability post-processing failed")
		return state, fmt.Errorf("capability post-processing failed: %w", err)
	}

	Logger().Debug().
		Str("agent", c.name).
		Msg("ComposableAgent execution completed successfully")

	return finalState, nil
}

func (c *ComposableAgent) HandleEvent(ctx context.Context, event Event, state State) (AgentResult, error) {
	startTime := time.Now()
	outputState, err := c.Run(ctx, state)
	endTime := time.Now()

	result := AgentResult{
		OutputState: outputState,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(startTime),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

// Capability management methods

// GetCapability returns the capability of the specified type, if present
func (c *ComposableAgent) GetCapability(capType CapabilityType) (AgentCapability, bool) {
	cap, exists := c.capabilities[capType]
	return cap, exists
}

// HasCapability checks if the agent has a capability of the specified type
func (c *ComposableAgent) HasCapability(capType CapabilityType) bool {
	_, exists := c.capabilities[capType]
	return exists
}

// ListCapabilityTypes returns a list of all capability types this agent has
func (c *ComposableAgent) ListCapabilityTypes() []CapabilityType {
	types := make([]CapabilityType, 0, len(c.capabilities))
	for capType := range c.capabilities {
		types = append(types, capType)
	}
	return types
}

// AddCapability adds a new capability to the agent
func (c *ComposableAgent) AddCapability(capType CapabilityType, capability AgentCapability) {
	c.capabilities[capType] = capability
}

// RemoveCapability removes a capability from the agent
func (c *ComposableAgent) RemoveCapability(capType CapabilityType) {
	delete(c.capabilities, capType)
}

// Configure implements basic capability configuration
func (c *ComposableAgent) Configure(configs map[CapabilityType]interface{}) error {
	// TODO: Implement full capability configuration when CapabilityConfigurable is complete
	Logger().Debug().
		Int("capability_count", len(configs)).
		Str("agent", c.name).
		Msg("Capability configuration requested but not fully implemented")
	return nil
}

// SetConfiguration updates the agent's configuration
func (c *ComposableAgent) SetConfiguration(config *ResolvedAgentConfig) {
	c.config = config
	if config != nil {
		c.role = config.Role
		c.description = config.Description
		c.systemPrompt = config.SystemPrompt
		c.timeout = time.Duration(config.Timeout) * time.Second
		c.enabled = config.Enabled
		c.llmConfig = config.LLMConfig
	}
}

// GetConfiguration returns the agent's current configuration
func (c *ComposableAgent) GetConfiguration() *ResolvedAgentConfig {
	return c.config
}

// String returns a string representation of the agent
func (c *ComposableAgent) String() string {
	return fmt.Sprintf("ComposableAgent{name: %s, role: %s, capabilities: %v}", c.name, c.role, c.ListCapabilityTypes())
}

// Private helper methods

func (c *ComposableAgent) applyCapabilityPreProcessing(ctx context.Context, state State) (State, error) {
	workingState := state

	// Apply pre-processing from all capabilities
	for capType, capability := range c.capabilities {
		if preprocessor, ok := capability.(interface {
			PreProcess(context.Context, State) (State, error)
		}); ok {
			var err error
			workingState, err = preprocessor.PreProcess(ctx, workingState)
			if err != nil {
				return state, fmt.Errorf("pre-processing failed for capability %s: %w", capType, err)
			}
		}
	}

	return workingState, nil
}

func (c *ComposableAgent) applyCapabilityPostProcessing(ctx context.Context, state State) (State, error) {
	workingState := state

	// Apply post-processing from all capabilities
	for capType, capability := range c.capabilities {
		if postprocessor, ok := capability.(interface {
			PostProcess(context.Context, State) (State, error)
		}); ok {
			var err error
			workingState, err = postprocessor.PostProcess(ctx, workingState)
			if err != nil {
				return state, fmt.Errorf("post-processing failed for capability %s: %w", capType, err)
			}
		}
	}

	return workingState, nil
}
