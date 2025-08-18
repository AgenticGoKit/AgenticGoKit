// Package agents provides the capability framework for building composable agents.
package agents

import (
	"fmt"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// =============================================================================
// CORE CAPABILITY FRAMEWORK
// =============================================================================

// CapabilityConfigurable represents an agent that can have capabilities configured on it.
// This interface allows capabilities to be configured without depending on the concrete UnifiedAgent type.
type CapabilityConfigurable interface {
	// SetLLMProvider sets the LLM provider for the agent
	SetLLMProvider(provider core.ModelProvider, config core.LLMConfig)

	// SetCacheManager sets the cache manager for the agent
	SetCacheManager(manager interface{}, config interface{})

	// SetMetricsConfig sets the metrics configuration for the agent
	SetMetricsConfig(config core.MetricsConfig)

	// GetLogger removed to avoid direct logger dependency; use core.Logger() instead
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

// CapabilityType defines the type of capability for validation and ordering
type CapabilityType string

const (
	CapabilityTypeLLM     CapabilityType = "llm"
	CapabilityTypeMCP     CapabilityType = "mcp"
	CapabilityTypeCache   CapabilityType = "cache"
	CapabilityTypeMetrics CapabilityType = "metrics"
	CapabilityTypeMemory  CapabilityType = "memory"
	CapabilityTypeTools   CapabilityType = "tools"
)

// CapabilityError represents an error during capability configuration
type CapabilityError struct {
	CapabilityName string
	Operation      string
	Err            error
}

func (e *CapabilityError) Error() string {
	return fmt.Sprintf("capability '%s' %s: %v", e.CapabilityName, e.Operation, e.Err)
}

func (e *CapabilityError) Unwrap() error {
	return e.Err
}

// NewCapabilityError creates a new capability error
func NewCapabilityError(capName, operation string, err error) *CapabilityError {
	return &CapabilityError{
		CapabilityName: capName,
		Operation:      operation,
		Err:            err,
	}
}

// =============================================================================
// CAPABILITY VALIDATION HELPERS
// =============================================================================

// CapabilityValidator provides common validation patterns for capabilities
type CapabilityValidator struct{}

// ValidateUnique ensures only one capability of a given type exists
func (v *CapabilityValidator) ValidateUnique(capType CapabilityType, others []AgentCapability) error {
	count := 0
	for _, other := range others {
		if other.Name() == string(capType) {
			count++
		}
	}

	if count > 0 {
		return fmt.Errorf("only one %s capability allowed per agent", capType)
	}

	return nil
}

// ValidateRequires ensures required capabilities are present
func (v *CapabilityValidator) ValidateRequires(required []CapabilityType, others []AgentCapability) error {
	available := make(map[CapabilityType]bool)
	for _, other := range others {
		available[CapabilityType(other.Name())] = true
	}

	var missing []CapabilityType
	for _, req := range required {
		if !available[req] {
			missing = append(missing, req)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required capabilities: %v", missing)
	}

	return nil
}

// ValidateIncompatible ensures incompatible capabilities are not present
func (v *CapabilityValidator) ValidateIncompatible(incompatible []CapabilityType, others []AgentCapability) error {
	for _, other := range others {
		otherType := CapabilityType(other.Name())
		for _, incompat := range incompatible {
			if otherType == incompat {
				return fmt.Errorf("capability %s is incompatible with %s", otherType, incompat)
			}
		}
	}

	return nil
}

// =============================================================================
// BASE CAPABILITY IMPLEMENTATIONS
// =============================================================================

// BaseCapability provides common functionality for capabilities
type BaseCapability struct {
	name     string
	priority int
}

func (c *BaseCapability) Name() string {
	return c.name
}

func (c *BaseCapability) Priority() int {
	return c.priority
}

// Default validation allows all combinations
func (c *BaseCapability) Validate(others []AgentCapability) error {
	return nil
}

// =============================================================================
// LLM CAPABILITY
// =============================================================================

// LLMCapability adds language model integration to agents
type LLMCapability struct {
	BaseCapability
	Provider core.ModelProvider
	Config   core.LLMConfig
}

// NewLLMCapability creates a new LLM capability
func NewLLMCapability(provider core.ModelProvider, config core.LLMConfig) *LLMCapability {
	return &LLMCapability{
		BaseCapability: BaseCapability{
			name:     string(CapabilityTypeLLM),
			priority: 10, // Initialize early
		},
		Provider: provider,
		Config:   config,
	}
}

func (c *LLMCapability) Configure(agent CapabilityConfigurable) error {
	if c.Provider == nil {
		return NewCapabilityError(c.Name(), "configuration",
			fmt.Errorf("LLM provider is required"))
	}

	agent.SetLLMProvider(c.Provider, c.Config)

	core.Logger().Info().
		Str("capability", c.Name()).
		Msg("LLM capability configured")

	return nil
}

func (c *LLMCapability) Validate(others []AgentCapability) error {
	validator := &CapabilityValidator{}
	return validator.ValidateUnique(CapabilityTypeLLM, others)
}

// =============================================================================
// CACHE CAPABILITY
// =============================================================================

// CacheCapability adds caching functionality to agents
type CacheCapability struct {
	BaseCapability
	Manager interface{} // Flexible cache manager interface
	Config  interface{} // Flexible cache configuration
}

// NewCacheCapability creates a new cache capability
func NewCacheCapability(manager interface{}, config interface{}) *CacheCapability {
	return &CacheCapability{
		BaseCapability: BaseCapability{
			name:     string(CapabilityTypeCache),
			priority: 20, // Initialize after core capabilities
		},
		Manager: manager,
		Config:  config,
	}
}

func (c *CacheCapability) Configure(agent CapabilityConfigurable) error {
	if c.Manager == nil {
		return NewCapabilityError(c.Name(), "configuration",
			fmt.Errorf("cache manager is required"))
	}

	agent.SetCacheManager(c.Manager, c.Config)

	core.Logger().Info().
		Str("capability", c.Name()).
		Msg("Cache capability configured")

	return nil
}

// =============================================================================
// METRICS CAPABILITY
// =============================================================================

// MetricsCapability adds metrics collection to agents
type MetricsCapability struct {
	BaseCapability
	Config core.MetricsConfig
}

// NewMetricsCapability creates a new metrics capability
func NewMetricsCapability(config core.MetricsConfig) *MetricsCapability {
	return &MetricsCapability{
		BaseCapability: BaseCapability{
			name:     string(CapabilityTypeMetrics),
			priority: 30, // Initialize after other capabilities
		},
		Config: config,
	}
}

func (c *MetricsCapability) Configure(agent CapabilityConfigurable) error {
	agent.SetMetricsConfig(c.Config)

	core.Logger().Info().
		Str("capability", c.Name()).
		Int("port", c.Config.Port).
		Msg("Metrics capability configured")

	return nil
}

// =============================================================================
// CAPABILITY CONFIGURATION TYPES
// =============================================================================

// DefaultLLMConfig returns sensible defaults for LLM configuration using core types
func DefaultLLMConfig() core.LLMConfig {
	return core.LLMConfig{
		Temperature:    0.7,
		MaxTokens:      1000,
		TimeoutSeconds: 30,
	}
}

// DefaultMetricsConfig returns sensible defaults for metrics configuration
func DefaultMetricsConfig() core.MetricsConfig {
	return core.MetricsConfig{
		Enabled:           true,
		Port:              8080,
		Path:              "/metrics",
		PrometheusEnabled: true,
	}
}

// =============================================================================
// CAPABILITY REGISTRY
// =============================================================================

// CapabilityRegistry manages available capability types
type CapabilityRegistry struct {
	capabilities map[CapabilityType]func() AgentCapability
}

// GlobalCapabilityRegistry is the default registry
var GlobalCapabilityRegistry = &CapabilityRegistry{
	capabilities: make(map[CapabilityType]func() AgentCapability),
}

// Register adds a capability factory to the registry
func (r *CapabilityRegistry) Register(capType CapabilityType, factory func() AgentCapability) {
	r.capabilities[capType] = factory
}

// Create creates a new capability instance
func (r *CapabilityRegistry) Create(capType CapabilityType) (AgentCapability, error) {
	factory, exists := r.capabilities[capType]
	if !exists {
		return nil, fmt.Errorf("unknown capability type: %s", capType)
	}
	return factory(), nil
}

// List returns all registered capability types
func (r *CapabilityRegistry) List() []CapabilityType {
	var types []CapabilityType
	for capType := range r.capabilities {
		types = append(types, capType)
	}
	return types
}

// =============================================================================
// CAPABILITY UTILITIES
// =============================================================================

// SortCapabilitiesByPriority sorts capabilities by their priority for initialization
func SortCapabilitiesByPriority(capabilities []AgentCapability) []AgentCapability {
	sorted := make([]AgentCapability, len(capabilities))
	copy(sorted, capabilities)

	// Simple bubble sort by priority (fine for small numbers of capabilities)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Priority() > sorted[j+1].Priority() {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// ValidateCapabilityCombination validates that a set of capabilities can work together
func ValidateCapabilityCombination(capabilities []AgentCapability) error {
	for i, cap := range capabilities {
		others := append(capabilities[:i], capabilities[i+1:]...)
		if err := cap.Validate(others); err != nil {
			return NewCapabilityError(cap.Name(), "validation", err)
		}
	}
	return nil
}

// GetCapabilityByType finds a capability of a specific type in a slice
func GetCapabilityByType(capabilities []AgentCapability, capType CapabilityType) AgentCapability {
	for _, cap := range capabilities {
		if cap.Name() == string(capType) {
			return cap
		}
	}
	return nil
}

// HasCapabilityType checks if a capability of a specific type exists in a slice
func HasCapabilityType(capabilities []AgentCapability, capType CapabilityType) bool {
	return GetCapabilityByType(capabilities, capType) != nil
}

// =============================================================================
// INITIALIZATION
// =============================================================================

func init() {
	// Register default capability factories
	GlobalCapabilityRegistry.Register(CapabilityTypeLLM, func() AgentCapability {
		return &LLMCapability{
			BaseCapability: BaseCapability{name: string(CapabilityTypeLLM), priority: 10},
		}
	})

	GlobalCapabilityRegistry.Register(CapabilityTypeCache, func() AgentCapability {
		return &CacheCapability{
			BaseCapability: BaseCapability{name: string(CapabilityTypeCache), priority: 20},
		}
	})

	GlobalCapabilityRegistry.Register(CapabilityTypeMetrics, func() AgentCapability {
		return &MetricsCapability{
			BaseCapability: BaseCapability{name: string(CapabilityTypeMetrics), priority: 30},
		}
	})
}
