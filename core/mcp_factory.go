// Package core provides global MCP factory and management functions.
package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// Global MCP manager instance and registry
var (
	globalMCPManager        MCPManager
	globalMCPRegistry       FunctionToolRegistry
	globalCacheManager      MCPCacheManager
	mcpManagerMutex         sync.RWMutex
	mcpRegistryMutex        sync.RWMutex
	cacheManagerMutex       sync.RWMutex
	mcpManagerInitialized   bool
	cacheManagerInitialized bool
)

// InitializeMCPManager initializes the global MCP manager with the provided configuration.
func InitializeMCPManager(config MCPConfig) error {
	mcpManagerMutex.Lock()
	defer mcpManagerMutex.Unlock()

	if mcpManagerInitialized {
		GetLogger().Debug().Msg("MCP manager already initialized")
		return nil
	}

	// This will be implemented by the internal factory
	manager, err := createMCPManagerInternal(config)
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	globalMCPManager = manager
	mcpManagerInitialized = true

	GetLogger().Info().Msg("MCP manager initialized successfully")
	return nil
}

// GetMCPManager returns the global MCP manager instance.
func GetMCPManager() MCPManager {
	mcpManagerMutex.RLock()
	defer mcpManagerMutex.RUnlock()
	return globalMCPManager
}

// InitializeMCPToolRegistry initializes the global MCP tool registry.
func InitializeMCPToolRegistry() error {
	mcpRegistryMutex.Lock()
	defer mcpRegistryMutex.Unlock()

	if globalMCPRegistry != nil {
		GetLogger().Debug().Msg("MCP tool registry already initialized")
		return nil
	}

	// This will be implemented by the internal factory
	registry, err := createMCPToolRegistryInternal()
	if err != nil {
		return fmt.Errorf("failed to create MCP tool registry: %w", err)
	}

	globalMCPRegistry = registry
	GetLogger().Info().Msg("MCP tool registry initialized successfully")
	return nil
}

// GetMCPToolRegistry returns the global MCP tool registry.
func GetMCPToolRegistry() FunctionToolRegistry {
	mcpRegistryMutex.RLock()
	defer mcpRegistryMutex.RUnlock()
	return globalMCPRegistry
}

// RegisterMCPToolsWithRegistry discovers and registers all available MCP tools with the registry.
func RegisterMCPToolsWithRegistry(ctx context.Context) error {
	manager := GetMCPManager()
	if manager == nil {
		return fmt.Errorf("MCP manager not initialized")
	}

	registry := GetMCPToolRegistry()
	if registry == nil {
		return fmt.Errorf("MCP tool registry not initialized")
	}

	// Refresh tools from all connected servers
	if err := manager.RefreshTools(ctx); err != nil {
		GetLogger().Warn().Err(err).Msg("Failed to refresh tools from some MCP servers")
	}

	// Get available tools
	tools := manager.GetAvailableTools()

	// Register each tool with the registry
	for _, toolInfo := range tools {
		mcpTool, err := createMCPToolAdapterInternal(toolInfo, manager)
		if err != nil {
			GetLogger().Warn().
				Str("tool", toolInfo.Name).
				Err(err).
				Msg("Failed to create tool adapter")
			continue
		}

		if err := registry.Register(mcpTool); err != nil {
			GetLogger().Warn().
				Str("tool", toolInfo.Name).
				Err(err).
				Msg("Failed to register MCP tool")
			continue
		}

		GetLogger().Debug().
			Str("tool", toolInfo.Name).
			Str("server", toolInfo.ServerName).
			Msg("Registered MCP tool")
	}

	GetLogger().Info().
		Int("total_tools", len(tools)).
		Msg("MCP tools registered with tool registry")

	return nil
}

// ShutdownMCPManager gracefully shuts down the MCP manager and all connections.
func ShutdownMCPManager() error {
	mcpManagerMutex.Lock()
	defer mcpManagerMutex.Unlock()

	if globalMCPManager == nil {
		return nil
	}

	if err := globalMCPManager.DisconnectAll(); err != nil {
		GetLogger().Error().Err(err).Msg("Failed to disconnect from some MCP servers")
		return err
	}

	globalMCPManager = nil
	mcpManagerInitialized = false

	GetLogger().Info().Msg("MCP manager shutdown completed")
	return nil
}

// NewMCPEnabledRunnerWithRegistry creates a runner that includes MCP tools from the registry.
func NewMCPEnabledRunnerWithRegistry(ctx context.Context, registry FunctionToolRegistry) (Runner, error) {
	if registry == nil {
		return nil, fmt.Errorf("tool registry cannot be nil")
	}

	// This would create a runner with the MCP tools
	// For now, return a basic runner - this needs to be implemented based on the actual runner interface
	GetLogger().Info().Msg("Creating MCP-enabled runner with tool registry")

	// TODO: Implement actual runner creation with MCP tools
	return nil, fmt.Errorf("MCP-enabled runner creation not yet implemented")
}

// CreateMCPAgentWithLLMAndTools creates a complete MCP agent setup with LLM and tool discovery.
func CreateMCPAgentWithLLMAndTools(ctx context.Context, name string, llmProvider LLMProvider, mcpConfig MCPConfig, agentConfig MCPAgentConfig) (*MCPAwareAgent, error) {
	// Initialize MCP infrastructure
	if err := InitializeMCPManager(mcpConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP manager: %w", err)
	}

	// Initialize tool registry
	if err := InitializeMCPToolRegistry(); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP tool registry: %w", err)
	}

	// Discover and connect to MCP servers
	mcpManager := GetMCPManager()
	if mcpManager != nil {
		if err := mcpManager.RefreshTools(ctx); err != nil {
			// Log warning but don't fail - agent can still work with manually configured tools
			GetLogger().Warn().Err(err).Msg("Failed to refresh MCP tools, agent will use manually configured tools")
		}
	}

	// Create the agent
	agent := NewMCPAwareAgent(name, llmProvider, mcpManager, agentConfig)

	GetLogger().Info().
		Str("agent", name).
		Int("available_tools", len(agent.GetAvailableMCPTools())).
		Msg("MCP-aware agent created successfully")

	return agent, nil
}

// NewMCPAwareAgentWithDefaults creates a new MCP-aware agent with default configuration.
func NewMCPAwareAgentWithDefaults(name string, llmProvider LLMProvider) (*MCPAwareAgent, error) {
	config := DefaultMCPAgentConfig()
	return NewMCPAwareAgentWithConfig(name, llmProvider, config)
}

// NewMCPAwareAgentWithConfig creates a new MCP-aware agent with custom configuration.
func NewMCPAwareAgentWithConfig(name string, llmProvider LLMProvider, config MCPAgentConfig) (*MCPAwareAgent, error) {
	// Get the global MCP manager
	mcpManager := GetMCPManager()
	if mcpManager == nil {
		return nil, fmt.Errorf("MCP manager not initialized. Call InitializeMCPManager first")
	}

	if llmProvider == nil {
		return nil, fmt.Errorf("LLM provider cannot be nil")
	}

	agent := NewMCPAwareAgent(name, llmProvider, mcpManager, config)
	return agent, nil
}

// GetLogger returns a logger instance.
// This is a temporary implementation - should use the actual logging system
func GetLogger() *zerolog.Logger {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	return &logger
}

// LLMProvider is an alias for ModelProvider for compatibility
type LLMProvider = ModelProvider

// These functions will be implemented by the internal factory to avoid circular imports
var (
	createMCPManagerInternal      func(MCPConfig) (MCPManager, error)
	createMCPToolRegistryInternal func() (FunctionToolRegistry, error)
	createMCPToolAdapterInternal  func(MCPToolInfo, MCPManager) (FunctionTool, error)
)

// SetMCPFactoryFunctions sets the internal factory functions to avoid circular imports.
func SetMCPFactoryFunctions(
	managerFactory func(MCPConfig) (MCPManager, error),
	registryFactory func() (FunctionToolRegistry, error),
	toolAdapterFactory func(MCPToolInfo, MCPManager) (FunctionTool, error),
) {
	createMCPManagerInternal = managerFactory
	createMCPToolRegistryInternal = registryFactory
	createMCPToolAdapterInternal = toolAdapterFactory
}
