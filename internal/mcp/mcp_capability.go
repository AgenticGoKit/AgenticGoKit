// Package core provides MCP capability implementation for the unified agent architecture.
package mcp

import (
	"fmt"
)

// =============================================================================
// MCP CAPABILITY IMPLEMENTATION
// =============================================================================

// MCPCapability adds MCP tool integration to any agent
type MCPCapability struct {
	BaseCapability
	Manager      MCPManager
	Config       MCPAgentConfig
	CacheManager MCPCacheManager // Optional caching
}

// NewMCPCapability creates a new MCP capability
func NewMCPCapability(manager MCPManager, config MCPAgentConfig) *MCPCapability {
	return &MCPCapability{
		BaseCapability: BaseCapability{
			name:     string(CapabilityTypeMCP),
			priority: 15, // Initialize after LLM but before cache
		},
		Manager: manager,
		Config:  config,
	}
}

// NewMCPCapabilityWithCache creates a new MCP capability with caching
func NewMCPCapabilityWithCache(manager MCPManager, config MCPAgentConfig, cacheManager MCPCacheManager) *MCPCapability {
	return &MCPCapability{
		BaseCapability: BaseCapability{
			name:     string(CapabilityTypeMCP),
			priority: 15,
		},
		Manager:      manager,
		Config:       config,
		CacheManager: cacheManager,
	}
}

func (c *MCPCapability) Configure(agent CapabilityConfigurable) error {
	if c.Manager == nil {
		return NewCapabilityError(c.Name(), "configuration",
			fmt.Errorf("MCP manager is required"))
	}

	// For now, we'll add a method to set MCP configuration
	// This will be properly implemented when we create the UnifiedAgent
	if mcpConfigurable, ok := agent.(interface {
		SetMCPManager(manager MCPManager, config MCPAgentConfig)
		SetMCPCacheManager(manager MCPCacheManager)
	}); ok {
		mcpConfigurable.SetMCPManager(c.Manager, c.Config)
		if c.CacheManager != nil {
			mcpConfigurable.SetMCPCacheManager(c.CacheManager)
		}
	} else {
		return NewCapabilityError(c.Name(), "configuration",
			fmt.Errorf("agent does not support MCP capability"))
	}

	agent.GetLogger().Info().
		Str("capability", c.Name()).
		Msg("MCP capability configured")

	return nil
}

func (c *MCPCapability) Validate(others []AgentCapability) error {
	validator := &CapabilityValidator{}

	// Only one MCP capability allowed per agent
	if err := validator.ValidateUnique(CapabilityTypeMCP, others); err != nil {
		return err
	}

	// MCP capability works best with LLM capability for intelligent tool selection
	if !HasCapabilityType(others, CapabilityTypeLLM) {
		agent := GetCapabilityByType(others, CapabilityTypeLLM)
		if agent == nil {
			// This is a warning, not an error - MCP can work without LLM
			// but won't have intelligent tool selection
		}
	}

	return nil
}

// =============================================================================
// MCP CAPABILITY REGISTRY INITIALIZATION
// =============================================================================

func init() {
	// Register MCP capability factory
	GlobalCapabilityRegistry.Register(CapabilityTypeMCP, func() AgentCapability {
		return &MCPCapability{
			BaseCapability: BaseCapability{name: string(CapabilityTypeMCP), priority: 15},
		}
	})
}
