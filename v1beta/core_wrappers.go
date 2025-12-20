// Package v1beta provides wrapper functions for core functionality.
// This allows v1beta to gradually become independent from the core package.
package v1beta

import (
	"context"

	"github.com/agenticgokit/agenticgokit/core"
)

// =============================================================================
// LOGGING WRAPPERS
// =============================================================================

// Logger returns the configured logger instance
// This directly returns core's logger for now
func Logger() core.CoreLogger {
	return core.Logger()
}

// =============================================================================
// MEMORY FACTORY WRAPPERS
// =============================================================================

// newCoreMemory creates a memory provider from core.AgentMemoryConfig
// This wraps core.NewMemory for internal use
func newCoreMemory(config core.AgentMemoryConfig) (core.Memory, error) {
	return core.NewMemory(config)
}

// =============================================================================
// MCP WRAPPERS
// =============================================================================

// InitializeMCPWithCache initializes MCP with cache configuration
func InitializeMCPWithCache(mcpConfig core.MCPConfig, cacheConfig core.MCPCacheConfig) error {
	return core.InitializeMCPWithCache(mcpConfig, cacheConfig)
}

// GetMCPManager returns the MCP manager instance
func GetMCPManager() core.MCPManager {
	return core.GetMCPManager()
}

// ExecuteMCPTool executes an MCP tool
func ExecuteMCPTool(ctx context.Context, toolName string, args map[string]interface{}) (core.MCPToolResult, error) {
	return core.ExecuteMCPTool(ctx, toolName, args)
}
