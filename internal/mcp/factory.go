package mcp

import (
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/internal/tools"
)

// NewMCPManager creates a new MCP manager with the given configuration.
//
// The manager handles MCP server connections, tool discovery, and integration
// with the AgentFlow tool registry. It provides a high-level interface for
// working with MCP servers and their tools.
//
// Example usage:
//
//	config := core.MCPConfig{
//	    EnableDiscovery: true,
//	    ConnectionTimeout: 30 * time.Second,
//	    Servers: []core.MCPServerConfig{
//	        {
//	            Name: "local-server",
//	            Type: "tcp",
//	            Host: "localhost",
//	            Port: 8080,
//	            Enabled: true,
//	        },
//	    },
//	}
//
//	registry := tools.NewToolRegistry()
//	manager, err := mcp.NewMCPManager(config, registry, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewMCPManager(config core.MCPConfig, registry *tools.ToolRegistry, logger *log.Logger) (core.MCPManager, error) {
	if registry == nil {
		return nil, fmt.Errorf("tool registry cannot be nil")
	}

	// Set default values for required configuration
	if config.ConnectionTimeout == 0 {
		config.ConnectionTimeout = 30 * time.Second
	}
	if config.DiscoveryTimeout == 0 {
		config.DiscoveryTimeout = 10 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.MaxConnections == 0 {
		config.MaxConnections = 10
	}
	if config.CacheTimeout == 0 {
		config.CacheTimeout = 5 * time.Minute
	}

	return NewMCPManagerImpl(config, registry, logger)
}

// NewMCPManagerImpl creates the concrete implementation.
func NewMCPManagerImpl(config core.MCPConfig, registry *tools.ToolRegistry, logger *log.Logger) (*MCPManagerImpl, error) {
	// Call the original constructor that was renamed to avoid conflicts
	return createMCPManager(config, registry, logger)
}
