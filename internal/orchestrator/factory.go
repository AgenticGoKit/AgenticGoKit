// Package orchestrator provides internal orchestrator factory functionality.
package orchestrator

import (
	"fmt"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// OrchestratorType represents the type of orchestrator
type OrchestratorType string

const (
	OrchestratorTypeRoute         OrchestratorType = "route"
	OrchestratorTypeCollaborative OrchestratorType = "collaborative"
	OrchestratorTypeSequential    OrchestratorType = "sequential"
	OrchestratorTypeLoop          OrchestratorType = "loop"
	OrchestratorTypeMixed         OrchestratorType = "mixed"
	OrchestratorTypeParallel      OrchestratorType = "parallel"
)

// OrchestratorConfig holds configuration for creating orchestrators
type OrchestratorConfig struct {
	Type                    OrchestratorType `json:"type" toml:"type"`
	Timeout                 time.Duration    `json:"timeout,omitempty" toml:"timeout,omitempty"`
	MaxConcurrency          int              `json:"max_concurrency,omitempty" toml:"max_concurrency,omitempty"`
	FailureThreshold        float64          `json:"failure_threshold,omitempty" toml:"failure_threshold,omitempty"`
	
	// Sequential and Loop specific
	AgentSequence           []string         `json:"agent_sequence,omitempty" toml:"agent_sequence,omitempty"`
	MaxIterations           int              `json:"max_iterations,omitempty" toml:"max_iterations,omitempty"`
	
	// Mixed orchestrator specific
	CollaborativeAgents     []string         `json:"collaborative_agents,omitempty" toml:"collaborative_agents,omitempty"`
	SequentialAgents        []string         `json:"sequential_agents,omitempty" toml:"sequential_agents,omitempty"`
}

// OrchestratorFactory creates orchestrators based on configuration
type OrchestratorFactory struct {
	registry *core.CallbackRegistry
}

// NewOrchestratorFactory creates a new orchestrator factory
func NewOrchestratorFactory(registry *core.CallbackRegistry) *OrchestratorFactory {
	return &OrchestratorFactory{
		registry: registry,
	}
}

// CreateOrchestrator creates an orchestrator based on the configuration
func (f *OrchestratorFactory) CreateOrchestrator(config OrchestratorConfig) (core.Orchestrator, error) {
	switch config.Type {
	case OrchestratorTypeRoute:
		return f.createRouteOrchestrator(config)
	case OrchestratorTypeCollaborative, OrchestratorTypeParallel:
		return f.createCollaborativeOrchestrator(config)
	case OrchestratorTypeSequential:
		return f.createSequentialOrchestrator(config)
	case OrchestratorTypeLoop:
		return f.createLoopOrchestrator(config)
	case OrchestratorTypeMixed:
		return f.createMixedOrchestrator(config)
	default:
		return nil, fmt.Errorf("unsupported orchestrator type: %s", config.Type)
	}
}

// createRouteOrchestrator creates a route orchestrator
func (f *OrchestratorFactory) createRouteOrchestrator(config OrchestratorConfig) (core.Orchestrator, error) {
	return NewRouteOrchestrator(f.registry), nil
}

// createCollaborativeOrchestrator creates a collaborative orchestrator
func (f *OrchestratorFactory) createCollaborativeOrchestrator(config OrchestratorConfig) (core.Orchestrator, error) {
	return NewCollaborativeOrchestrator(f.registry), nil
}

// createSequentialOrchestrator creates a sequential orchestrator
func (f *OrchestratorFactory) createSequentialOrchestrator(config OrchestratorConfig) (core.Orchestrator, error) {
	if len(config.AgentSequence) == 0 {
		return nil, fmt.Errorf("sequential orchestrator requires agent sequence")
	}
	return NewSequentialOrchestrator(f.registry, config.AgentSequence), nil
}

// createLoopOrchestrator creates a loop orchestrator
func (f *OrchestratorFactory) createLoopOrchestrator(config OrchestratorConfig) (core.Orchestrator, error) {
	if len(config.AgentSequence) == 0 {
		return nil, fmt.Errorf("loop orchestrator requires at least one agent")
	}
	
	orchestrator := NewLoopOrchestrator(f.registry, config.AgentSequence)
	
	// Set max iterations if specified
	if config.MaxIterations > 0 {
		orchestrator.SetMaxIterations(config.MaxIterations)
	}
	
	return orchestrator, nil
}

// createMixedOrchestrator creates a mixed orchestrator
func (f *OrchestratorFactory) createMixedOrchestrator(config OrchestratorConfig) (core.Orchestrator, error) {
	return NewMixedOrchestrator(f.registry, config.CollaborativeAgents, config.SequentialAgents), nil
}

// DefaultFactory is a global factory instance for convenience
var DefaultFactory *OrchestratorFactory

// SetDefaultRegistry sets the registry for the default factory
func SetDefaultRegistry(registry *core.CallbackRegistry) {
	DefaultFactory = NewOrchestratorFactory(registry)
}

// CreateOrchestratorFromConfig is a convenience function that uses the default factory
func CreateOrchestratorFromConfig(config OrchestratorConfig) (core.Orchestrator, error) {
	if DefaultFactory == nil {
		return nil, fmt.Errorf("default factory not initialized - call SetDefaultRegistry first")
	}
	return DefaultFactory.CreateOrchestrator(config)
}

// Convenience functions for common orchestrator types

// CreateRouteOrchestrator creates a route orchestrator
func CreateRouteOrchestrator(registry *core.CallbackRegistry) core.Orchestrator {
	return NewRouteOrchestrator(registry)
}

// CreateCollaborativeOrchestrator creates a collaborative orchestrator
func CreateCollaborativeOrchestrator(registry *core.CallbackRegistry) core.Orchestrator {
	return NewCollaborativeOrchestrator(registry)
}

// CreateSequentialOrchestrator creates a sequential orchestrator
func CreateSequentialOrchestrator(registry *core.CallbackRegistry, agentNames []string) core.Orchestrator {
	return NewSequentialOrchestrator(registry, agentNames)
}

// CreateLoopOrchestrator creates a loop orchestrator
func CreateLoopOrchestrator(registry *core.CallbackRegistry, agentNames []string) core.Orchestrator {
	return NewLoopOrchestrator(registry, agentNames)
}

// CreateMixedOrchestrator creates a mixed orchestrator
func CreateMixedOrchestrator(registry *core.CallbackRegistry, collaborativeAgents, sequentialAgents []string) core.Orchestrator {
	return NewMixedOrchestrator(registry, collaborativeAgents, sequentialAgents)
}

// init registers the orchestrator factory with the core package
func init() {
	// Register our factory function with the core package
	core.RegisterOrchestratorFactory(func(config core.OrchestratorConfig, registry *core.CallbackRegistry) (core.Orchestrator, error) {
		// Convert core config to internal config
		internalConfig := OrchestratorConfig{
			Type:                OrchestratorType(config.Type),
			AgentSequence:       config.AgentNames,
			MaxIterations:       config.MaxIterations,
			CollaborativeAgents: config.CollaborativeAgentNames,
			SequentialAgents:    config.SequentialAgentNames,
		}
		
		// Handle legacy field names
		if len(internalConfig.AgentSequence) == 0 && len(config.SequentialAgentNames) > 0 {
			internalConfig.AgentSequence = config.SequentialAgentNames
		}
		
		// Create factory and orchestrator
		factory := NewOrchestratorFactory(registry)
		return factory.CreateOrchestrator(internalConfig)
	})
}