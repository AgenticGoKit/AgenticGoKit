// Package agents: MCP capability using core MCP interfaces
package agents

import (
	"fmt"

	"github.com/agenticgokit/agenticgokit/core"
)

// MCPCapability adds MCP tool integration to any agent
type MCPCapability struct {
	BaseCapability
	Manager      core.MCPManager
	Config       core.MCPAgentConfig
	CacheManager core.MCPCacheManager // Optional caching
}

// NewMCPCapability creates a new MCP capability
func NewMCPCapability(manager core.MCPManager, config core.MCPAgentConfig) *MCPCapability {
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
func NewMCPCapabilityWithCache(manager core.MCPManager, config core.MCPAgentConfig, cacheManager core.MCPCacheManager) *MCPCapability {
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

	// Agents that support MCP should expose these methods
	if mcpConfigurable, ok := agent.(interface {
		SetMCPManager(manager core.MCPManager, config core.MCPAgentConfig)
		SetMCPCacheManager(manager core.MCPCacheManager)
	}); ok {
		mcpConfigurable.SetMCPManager(c.Manager, c.Config)
		if c.CacheManager != nil {
			mcpConfigurable.SetMCPCacheManager(c.CacheManager)
		}
	} else {
		return NewCapabilityError(c.Name(), "configuration",
			fmt.Errorf("agent does not support MCP capability"))
	}

	core.Logger().Debug().
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
		// It's okay; no hard requirement enforced
	}

	return nil
}

func init() {
	// Register MCP capability factory
	GlobalCapabilityRegistry.Register(CapabilityTypeMCP, func() AgentCapability {
		return &MCPCapability{
			BaseCapability: BaseCapability{name: string(CapabilityTypeMCP), priority: 15},
		}
	})
}

