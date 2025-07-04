// Package core provides the unified agent builder for creating composable agents.
package core

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/rs/zerolog"
)

// =============================================================================
// UNIFIED AGENT BUILDER
// =============================================================================

// AgentBuilder provides a fluent interface for building agents with capabilities.
// It allows for easy composition of different agent features through a builder pattern.
type AgentBuilder struct {
	name         string
	capabilities []AgentCapability
	errors       []error
	config       AgentBuilderConfig
	// Multi-agent composition fields
	compositionMode string
	subAgents       []Agent
	multiConfig     MultiAgentConfig
	loopConfig      LoopConfig
}

// AgentBuilderConfig contains configuration for the agent builder
type AgentBuilderConfig struct {
	ValidateCapabilities bool // Whether to validate capability combinations
	SortByPriority       bool // Whether to sort capabilities by priority
	StrictMode           bool // Whether to fail on any capability error
}

// DefaultAgentBuilderConfig returns sensible defaults for agent builder configuration
func DefaultAgentBuilderConfig() AgentBuilderConfig {
	return AgentBuilderConfig{
		ValidateCapabilities: true,
		SortByPriority:       true,
		StrictMode:           true,
	}
}

// NewAgent creates a new agent builder with the specified name
func NewAgent(name string) *AgentBuilder {
	return &AgentBuilder{
		name:         name,
		capabilities: make([]AgentCapability, 0),
		errors:       make([]error, 0),
		config:       DefaultAgentBuilderConfig(),
	}
}

// NewAgentWithConfig creates a new agent builder with custom configuration
func NewAgentWithConfig(name string, config AgentBuilderConfig) *AgentBuilder {
	return &AgentBuilder{
		name:         name,
		capabilities: make([]AgentCapability, 0),
		errors:       make([]error, 0),
		config:       config,
	}
}

// =============================================================================
// BUILDER CONFIGURATION METHODS
// =============================================================================

// WithValidation enables or disables capability validation
func (b *AgentBuilder) WithValidation(validate bool) *AgentBuilder {
	b.config.ValidateCapabilities = validate
	return b
}

// WithStrictMode enables or disables strict mode
func (b *AgentBuilder) WithStrictMode(strict bool) *AgentBuilder {
	b.config.StrictMode = strict
	return b
}

// =============================================================================
// MULTI-AGENT COMPOSITION METHODS
// =============================================================================

// WithParallelAgents configures the agent to run sub-agents in parallel
func (b *AgentBuilder) WithParallelAgents(agents ...Agent) *AgentBuilder {
	b.compositionMode = "parallel"
	b.subAgents = append(b.subAgents, agents...)
	return b
}

// WithSequentialAgents configures the agent to run sub-agents sequentially
func (b *AgentBuilder) WithSequentialAgents(agents ...Agent) *AgentBuilder {
	b.compositionMode = "sequential"
	b.subAgents = append(b.subAgents, agents...)
	return b
}

// WithLoopAgent configures the agent to repeatedly run a sub-agent
func (b *AgentBuilder) WithLoopAgent(agent Agent, maxIterations int, condition func(State) bool) *AgentBuilder {
	b.compositionMode = "loop"
	b.subAgents = []Agent{agent}
	b.loopConfig = LoopConfig{
		MaxIterations: maxIterations,
		Condition:     condition,
	}
	return b
}

// WithMultiAgentConfig sets the configuration for multi-agent composition
func (b *AgentBuilder) WithMultiAgentConfig(config MultiAgentConfig) *AgentBuilder {
	b.multiConfig = config
	return b
}

// WithMultiAgentTimeout sets the timeout for multi-agent composition
func (b *AgentBuilder) WithMultiAgentTimeout(timeout time.Duration) *AgentBuilder {
	if b.multiConfig.Timeout == 0 {
		b.multiConfig = DefaultMultiAgentConfig()
	}
	b.multiConfig.Timeout = timeout
	return b
}

// WithMultiAgentErrorStrategy sets the error handling strategy for multi-agent composition
func (b *AgentBuilder) WithMultiAgentErrorStrategy(strategy ErrorHandlingStrategy) *AgentBuilder {
	if b.multiConfig.Timeout == 0 {
		b.multiConfig = DefaultMultiAgentConfig()
	}
	b.multiConfig.ErrorStrategy = strategy
	return b
}

// WithMultiAgentConcurrency sets the maximum concurrency for multi-agent composition
func (b *AgentBuilder) WithMultiAgentConcurrency(maxConcurrency int) *AgentBuilder {
	if b.multiConfig.Timeout == 0 {
		b.multiConfig = DefaultMultiAgentConfig()
	}
	b.multiConfig.MaxConcurrency = maxConcurrency
	return b
}

// =============================================================================
// CAPABILITY ADDITION METHODS
// =============================================================================

// WithMCP adds MCP capability to the agent
func (b *AgentBuilder) WithMCP(manager MCPManager) *AgentBuilder {
	if manager == nil {
		if b.config.StrictMode {
			b.errors = append(b.errors, fmt.Errorf("MCP manager cannot be nil"))
			return b
		}
		// In non-strict mode, create a stub capability or skip
		return b
	}

	capability := NewMCPCapability(manager, DefaultMCPAgentConfig())
	b.capabilities = append(b.capabilities, capability)
	return b
}

// WithMCPAndConfig adds MCP capability with custom configuration
func (b *AgentBuilder) WithMCPAndConfig(manager MCPManager, config MCPAgentConfig) *AgentBuilder {
	if manager == nil {
		if b.config.StrictMode {
			b.errors = append(b.errors, fmt.Errorf("MCP manager cannot be nil"))
			return b
		}
		// In non-strict mode, create a stub capability or skip
		return b
	}

	capability := NewMCPCapability(manager, config)
	b.capabilities = append(b.capabilities, capability)
	return b
}

// WithMCPAndCache adds MCP capability with caching
func (b *AgentBuilder) WithMCPAndCache(manager MCPManager, cacheManager MCPCacheManager) *AgentBuilder {
	if manager == nil {
		if b.config.StrictMode {
			b.errors = append(b.errors, fmt.Errorf("MCP manager cannot be nil"))
			return b
		}
		// In non-strict mode, skip MCP capability
		return b
	}
	if cacheManager == nil {
		if b.config.StrictMode {
			b.errors = append(b.errors, fmt.Errorf("MCP cache manager cannot be nil"))
			return b
		}
		// In non-strict mode, skip MCP capability
		return b
	}

	capability := NewMCPCapabilityWithCache(manager, DefaultMCPAgentConfig(), cacheManager)
	b.capabilities = append(b.capabilities, capability)
	return b
}

// WithLLM adds LLM capability to the agent
func (b *AgentBuilder) WithLLM(provider ModelProvider) *AgentBuilder {
	if provider == nil {
		b.errors = append(b.errors, fmt.Errorf("LLM provider cannot be nil"))
		return b
	}

	capability := NewLLMCapability(provider, DefaultLLMConfig())
	b.capabilities = append(b.capabilities, capability)
	return b
}

// WithLLMAndConfig adds LLM capability with custom configuration
func (b *AgentBuilder) WithLLMAndConfig(provider ModelProvider, config LLMConfig) *AgentBuilder {
	if provider == nil {
		b.errors = append(b.errors, fmt.Errorf("LLM provider cannot be nil"))
		return b
	}

	capability := NewLLMCapability(provider, config)
	b.capabilities = append(b.capabilities, capability)
	return b
}

// WithCache adds cache capability to the agent
func (b *AgentBuilder) WithCache(manager interface{}, config interface{}) *AgentBuilder {
	if manager == nil {
		if b.config.StrictMode {
			b.errors = append(b.errors, fmt.Errorf("cache manager cannot be nil"))
			return b
		}
		// In non-strict mode, skip cache capability
		return b
	}

	capability := NewCacheCapability(manager, config)
	b.capabilities = append(b.capabilities, capability)
	return b
}

// WithMetrics adds metrics capability to the agent
func (b *AgentBuilder) WithMetrics(config MetricsConfig) *AgentBuilder {
	capability := NewMetricsCapability(config)
	b.capabilities = append(b.capabilities, capability)
	return b
}

// WithDefaultMetrics adds metrics capability with default configuration
func (b *AgentBuilder) WithDefaultMetrics() *AgentBuilder {
	capability := NewMetricsCapability(DefaultMetricsConfig())
	b.capabilities = append(b.capabilities, capability)
	return b
}

// WithCapability adds a custom capability to the agent
func (b *AgentBuilder) WithCapability(capability AgentCapability) *AgentBuilder {
	if capability == nil {
		b.errors = append(b.errors, fmt.Errorf("capability cannot be nil"))
		return b
	}

	b.capabilities = append(b.capabilities, capability)
	return b
}

// WithCapabilities adds multiple capabilities to the agent
func (b *AgentBuilder) WithCapabilities(capabilities ...AgentCapability) *AgentBuilder {
	for _, cap := range capabilities {
		b.WithCapability(cap)
	}
	return b
}

// =============================================================================
// BUILDER INTROSPECTION METHODS
// =============================================================================

// HasCapability checks if the builder has a specific capability type
func (b *AgentBuilder) HasCapability(capType CapabilityType) bool {
	return HasCapabilityType(b.capabilities, capType)
}

// GetCapability returns a capability of a specific type if present
func (b *AgentBuilder) GetCapability(capType CapabilityType) AgentCapability {
	return GetCapabilityByType(b.capabilities, capType)
}

// ListCapabilities returns all capability types currently in the builder
func (b *AgentBuilder) ListCapabilities() []CapabilityType {
	var types []CapabilityType
	for _, cap := range b.capabilities {
		types = append(types, CapabilityType(cap.Name()))
	}
	return types
}

// CapabilityCount returns the number of capabilities in the builder
func (b *AgentBuilder) CapabilityCount() int {
	return len(b.capabilities)
}

// =============================================================================
// VALIDATION AND ERROR HANDLING
// =============================================================================

// Validate validates the current capability configuration
func (b *AgentBuilder) Validate() error {
	// Check for builder errors first
	if len(b.errors) > 0 {
		return fmt.Errorf("builder has %d errors: %v", len(b.errors), b.errors)
	}

	// Validate capability combinations if enabled
	if b.config.ValidateCapabilities {
		return ValidateCapabilityCombination(b.capabilities)
	}

	return nil
}

// GetErrors returns any errors that occurred during building
func (b *AgentBuilder) GetErrors() []error {
	return b.errors
}

// HasErrors checks if the builder has any errors
func (b *AgentBuilder) HasErrors() bool {
	return len(b.errors) > 0
}

// ClearErrors clears all builder errors
func (b *AgentBuilder) ClearErrors() *AgentBuilder {
	b.errors = make([]error, 0)
	return b
}

// =============================================================================
// BUILD METHODS
// =============================================================================

// Build creates the final agent with all configured capabilities
func (b *AgentBuilder) Build() (Agent, error) {
	// Validate the configuration
	if err := b.Validate(); err != nil {
		return nil, fmt.Errorf("agent validation failed: %w", err)
	}

	// Check if this is a multi-agent composition
	if b.compositionMode != "" {
		return b.buildCompositeAgent()
	}

	// Sort capabilities by priority if enabled
	capabilities := b.capabilities
	if b.config.SortByPriority {
		capabilities = SortCapabilitiesByPriority(b.capabilities)
	}

	// Create the unified agent
	agent, err := createUnifiedAgent(b.name, capabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to create unified agent: %w", err)
	}
	// Configure each capability on the agent
	logger := Logger().With().Str("agent", b.name).Logger()

	// We need to cast the agent to CapabilityConfigurable to configure capabilities
	configurableAgent, ok := agent.(CapabilityConfigurable)
	if !ok {
		return nil, fmt.Errorf("agent does not implement CapabilityConfigurable interface")
	}

	for _, cap := range capabilities {
		if err := cap.Configure(configurableAgent); err != nil {
			if b.config.StrictMode {
				return nil, fmt.Errorf("failed to configure capability %s: %w", cap.Name(), err)
			} else {
				logger.Warn().
					Str("capability", cap.Name()).
					Err(err).
					Msg("Failed to configure capability (non-strict mode)")
			}
		}
	}

	logger.Info().
		Int("capabilities", len(capabilities)).
		Strs("capability_types", capabilityNames(capabilities)).
		Msg("Agent built successfully")

	return agent, nil
}

// BuildOrPanic builds the agent and panics if there are any errors.
// This is useful for testing or when you're certain the configuration is valid.
func (b *AgentBuilder) BuildOrPanic() Agent {
	agent, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build agent '%s': %v", b.name, err))
	}
	return agent
}

// =============================================================================
// MULTI-AGENT CONVENIENCE METHODS
// =============================================================================

// CreateDataProcessingPipeline creates a sequential agent for data processing workflows
// Usage: input -> processing -> output
func CreateDataProcessingPipeline(name string, inputAgent, processingAgent, outputAgent Agent) Agent {
	pipeline, _ := NewAgent(name).
		WithSequentialAgents(inputAgent, processingAgent, outputAgent).
		Build()
	return pipeline
}

// CreateParallelAnalysis creates a parallel agent for analysis workflows
// Usage: Multiple analysis agents process the same input concurrently
func CreateParallelAnalysisWorkflow(name string, timeout time.Duration, analysisAgents ...Agent) Agent {
	workflow, _ := NewAgent(name).
		WithParallelAgents(analysisAgents...).
		WithMultiAgentTimeout(timeout).
		WithMultiAgentErrorStrategy(ErrorStrategyCollectAll).
		Build()
	return workflow
}

// CreateResilientWorkflow creates a parallel agent with fault tolerance
func CreateResilientWorkflow(name string, agents ...Agent) Agent {
	workflow, _ := NewAgent(name).
		WithParallelAgents(agents...).
		WithMultiAgentTimeout(60 * time.Second).
		WithMultiAgentErrorStrategy(ErrorStrategyContinue).
		WithMultiAgentConcurrency(20).
		Build()
	return workflow
}

// CreateConditionalProcessor creates a loop agent with a simple condition
func CreateConditionalProcessor(name string, maxIterations int, conditionKey string, expectedValue interface{}, agent Agent) Agent {
	condition := func(state State) bool {
		if value, exists := state.Get(conditionKey); exists {
			return value == expectedValue
		}
		return false
	}

	processor, _ := NewAgent(name).
		WithLoopAgent(agent, maxIterations, condition).
		WithMultiAgentTimeout(120 * time.Second).
		Build()
	return processor
}

// =============================================================================
// CONFIGURATION-DRIVEN CREATION
// =============================================================================

// SimpleAgentConfig represents a complete agent configuration that can be loaded from files
type SimpleAgentConfig struct {
	Name string `toml:"name"`

	// Capability configurations
	LLM     *LLMConfig     `toml:"llm"`
	MCP     *MCPConfig     `toml:"mcp"`
	Cache   *interface{}   `toml:"cache"` // Flexible cache configuration
	Metrics *MetricsConfig `toml:"metrics"`

	// Feature flags
	LLMEnabled     bool `toml:"llm_enabled"`
	MCPEnabled     bool `toml:"mcp_enabled"`
	CacheEnabled   bool `toml:"cache_enabled"`
	MetricsEnabled bool `toml:"metrics_enabled"`
}

// NewAgentFromConfig creates an agent from configuration
// Note: This is a placeholder implementation. Full implementation would require
// creating providers and managers from configuration.
func NewAgentFromConfig(name string, config SimpleAgentConfig) (Agent, error) {
	builder := NewAgent(name)

	// Add LLM capability if configured
	if config.LLMEnabled && config.LLM != nil {
		// TODO: Create provider from config
		// provider := createLLMProviderFromConfig(*config.LLM)
		// builder = builder.WithLLMAndConfig(provider, *config.LLM)
	}

	// Add MCP capability if enabled
	if config.MCPEnabled && config.MCP != nil {
		// TODO: Create MCP manager from config
		// manager := createMCPManagerFromConfig(*config.MCP)
		// builder = builder.WithMCP(manager)
	}

	// Add metrics if enabled
	if config.MetricsEnabled && config.Metrics != nil {
		builder = builder.WithMetrics(*config.Metrics)
	}

	return builder.Build()
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// createUnifiedAgent creates a UnifiedAgent instance.
// createUnifiedAgent creates a new unified agent with the specified capabilities.
// This replaces the previous placeholder implementation with the actual UnifiedAgent.
func createUnifiedAgent(name string, capabilities []AgentCapability) (Agent, error) {
	// Convert slice to map for efficient lookup
	capabilityMap := make(map[CapabilityType]AgentCapability)
	for _, cap := range capabilities {
		// Determine capability type from the concrete implementation
		var capType CapabilityType
		switch cap.(type) {
		case *LLMCapability:
			capType = CapabilityTypeLLM
		case *CacheCapability:
			capType = CapabilityTypeCache
		case *MetricsCapability:
			capType = CapabilityTypeMetrics
		case *MCPCapability:
			capType = CapabilityTypeMCP
		default:
			// For unknown types, use the capability name as the type
			capType = CapabilityType(cap.Name())
		}
		capabilityMap[capType] = cap
	}

	// Create and return the unified agent
	return NewUnifiedAgent(name, capabilityMap, nil), nil
}

// PlaceholderAgent is a temporary implementation of both Agent and CapabilityConfigurable interfaces
// This will be replaced by the actual UnifiedAgent implementation
type PlaceholderAgent struct {
	name         string
	capabilities []AgentCapability
	logger       zerolog.Logger

	// Capability-specific fields
	llmProvider     ModelProvider
	llmConfig       LLMConfig
	mcpManager      MCPManager
	mcpConfig       MCPAgentConfig
	mcpCacheManager MCPCacheManager
	cacheManager    interface{}
	cacheConfig     interface{}
	metricsConfig   MetricsConfig
}

// Name implements the Agent interface
func (p *PlaceholderAgent) Name() string {
	return p.name
}

// Run implements the Agent interface - placeholder implementation
func (p *PlaceholderAgent) Run(ctx context.Context, inputState State) (State, error) {
	// Placeholder implementation - will be replaced by actual UnifiedAgent logic
	outputState := inputState.Clone()
	outputState.Set("processed_by", p.name)
	outputState.Set("capabilities", p.getCapabilityNames())
	return outputState, nil
}

// CapabilityConfigurable interface implementation
func (p *PlaceholderAgent) SetLLMProvider(provider ModelProvider, config LLMConfig) {
	p.llmProvider = provider
	p.llmConfig = config
}

func (p *PlaceholderAgent) SetCacheManager(manager interface{}, config interface{}) {
	p.cacheManager = manager
	p.cacheConfig = config
}

func (p *PlaceholderAgent) SetMetricsConfig(config MetricsConfig) {
	p.metricsConfig = config
}

func (p *PlaceholderAgent) GetLogger() *zerolog.Logger {
	return &p.logger
}

// MCP-specific configuration methods
func (p *PlaceholderAgent) SetMCPManager(manager MCPManager, config MCPAgentConfig) {
	p.mcpManager = manager
	p.mcpConfig = config
}

func (p *PlaceholderAgent) SetMCPCacheManager(manager MCPCacheManager) {
	p.mcpCacheManager = manager
}

// Helper method to get capability names
func (p *PlaceholderAgent) getCapabilityNames() []string {
	names := make([]string, len(p.capabilities))
	for i, cap := range p.capabilities {
		names[i] = cap.Name()
	}
	return names
}

// capabilityNames extracts the names of capabilities for logging
func capabilityNames(capabilities []AgentCapability) []string {
	names := make([]string, len(capabilities))
	for i, cap := range capabilities {
		names[i] = cap.Name()
	}
	sort.Strings(names) // Sort for consistent output
	return names
}

// buildCompositeAgent creates a composite agent based on the composition mode
func (b *AgentBuilder) buildCompositeAgent() (Agent, error) {
	if len(b.subAgents) == 0 {
		return nil, fmt.Errorf("composite agent '%s' requires at least one sub-agent", b.name)
	}

	// Use default config if not set
	config := b.multiConfig
	if config.Timeout == 0 {
		config = DefaultMultiAgentConfig()
	}

	logger := Logger().With().Str("agent", b.name).Str("composition_mode", b.compositionMode).Logger()

	switch b.compositionMode {
	case "parallel":
		logger.Info().
			Int("sub_agents", len(b.subAgents)).
			Dur("timeout", config.Timeout).
			Str("error_strategy", string(config.ErrorStrategy)).
			Msg("Building parallel composite agent")
		return NewParallelAgentWithConfig(b.name, config, b.subAgents...), nil

	case "sequential":
		logger.Info().
			Int("sub_agents", len(b.subAgents)).
			Msg("Building sequential composite agent")
		return NewSequentialAgent(b.name, b.subAgents...), nil

	case "loop":
		if len(b.subAgents) != 1 {
			return nil, fmt.Errorf("loop composite agent '%s' requires exactly one sub-agent, got %d", b.name, len(b.subAgents))
		}
		logger.Info().
			Int("max_iterations", b.loopConfig.MaxIterations).
			Dur("timeout", config.Timeout).
			Msg("Building loop composite agent")
		return NewLoopAgent(b.name, b.loopConfig.MaxIterations, config.Timeout,
			b.loopConfig.Condition, b.subAgents[0]), nil

	default:
		return nil, fmt.Errorf("unknown composition mode '%s' for agent '%s'", b.compositionMode, b.name)
	}
}

// =============================================================================
// VISUALIZATION METHODS
// =============================================================================

// GenerateMermaidDiagram generates a Mermaid diagram for the multi-agent composition
// Returns empty string if not a multi-agent composition
func (b *AgentBuilder) GenerateMermaidDiagram() string {
	return b.GenerateMermaidDiagramWithConfig(DefaultMermaidConfig())
}

// GenerateMermaidDiagramWithConfig generates a Mermaid diagram with custom configuration
func (b *AgentBuilder) GenerateMermaidDiagramWithConfig(config MermaidConfig) string {
	if b.compositionMode == "" || len(b.subAgents) == 0 {
		return "" // Not a multi-agent composition
	}

	// Create a temporary composition builder to generate the diagram
	composition := NewComposition(b.name).WithAgents(b.subAgents...)

	// Apply composition mode
	switch b.compositionMode {
	case "parallel":
		composition = composition.AsParallel()
	case "sequential":
		composition = composition.AsSequential()
	case "loop":
		if len(b.subAgents) == 1 && b.loopConfig.Condition != nil {
			composition = composition.AsLoop(b.loopConfig.MaxIterations, b.loopConfig.Condition)
		}
	}

	// Apply multi-agent configuration
	if b.multiConfig.Timeout > 0 {
		composition = composition.WithTimeout(b.multiConfig.Timeout)
	}
	if b.multiConfig.ErrorStrategy != "" {
		composition = composition.WithErrorStrategy(b.multiConfig.ErrorStrategy)
	}

	return composition.GenerateMermaidDiagramWithConfig(config)
}

// CanVisualize returns true if the agent builder has a multi-agent composition that can be visualized
func (b *AgentBuilder) CanVisualize() bool {
	return b.compositionMode != "" && len(b.subAgents) > 0
}

// =============================================================================
