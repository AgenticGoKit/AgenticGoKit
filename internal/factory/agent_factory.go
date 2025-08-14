package factory

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/internal/llm"
	"github.com/kunalkushwaha/agenticgokit/internal/mcp" // Import the MCP package
	"github.com/kunalkushwaha/agenticgokit/internal/tools"
)

// convertTomlToMCPConfig converts MCPConfigToml to MCPConfig
func convertTomlToMCPConfig(tomlConfig core.MCPConfigToml) core.MCPConfig {
	config := core.MCPConfig{
		EnableDiscovery:   tomlConfig.EnableDiscovery,
		DiscoveryTimeout:  time.Duration(tomlConfig.DiscoveryTimeout) * time.Millisecond,
		ScanPorts:         tomlConfig.ScanPorts,
		ConnectionTimeout: time.Duration(tomlConfig.ConnectionTimeout) * time.Millisecond,
		MaxRetries:        tomlConfig.MaxRetries,
		RetryDelay:        time.Duration(tomlConfig.RetryDelay) * time.Millisecond,
		EnableCaching:     tomlConfig.EnableCaching,
		CacheTimeout:      time.Duration(tomlConfig.CacheTimeout) * time.Millisecond,
		MaxConnections:    tomlConfig.MaxConnections,
	}

	// Convert servers
	for _, serverToml := range tomlConfig.Servers {
		server := core.MCPServerConfig{
			Name:    serverToml.Name,
			Type:    serverToml.Type,
			Host:    serverToml.Host,
			Port:    serverToml.Port,
			Command: serverToml.Command,
			Enabled: serverToml.Enabled,
		}
		config.Servers = append(config.Servers, server)
	}

	return config
}

// RunnerConfig allows customization but provides sensible defaults.
type RunnerConfig = core.RunnerConfig

// NewRunnerWithConfig wires up everything, registers agents, and returns a ready-to-use runner.
var NewRunnerWithConfig = core.NewRunnerWithConfig

// NewDefaultToolRegistry returns a ToolRegistry with built-in tools registered.
func NewDefaultToolRegistry() *tools.ToolRegistry {
	registry := tools.NewToolRegistry()
	_ = registry.Register(&tools.WebSearchTool{})
	_ = registry.Register(&tools.ComputeMetricTool{})
	return registry
}

// NewDefaultLLMAdapter returns an Azure OpenAI LLM adapter using environment variables.
func NewDefaultLLMAdapter() llm.ModelProvider {
	options := llm.AzureOpenAIAdapterOptions{
		Endpoint:            os.Getenv("AZURE_OPENAI_ENDPOINT"),
		APIKey:              os.Getenv("AZURE_OPENAI_API_KEY"),
		ChatDeployment:      os.Getenv("AZURE_OPENAI_DEPLOYMENT_ID"),
		EmbeddingDeployment: os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT"),
	}
	azureLLM, err := llm.NewAzureOpenAIAdapter(options)
	if err != nil {
		log.Fatalf("Failed to create Azure OpenAI adapter: %v", err)
	}
	return azureLLM
}

// NewMCPEnabledToolRegistry returns a ToolRegistry with built-in tools and MCP integration.
func NewMCPEnabledToolRegistry(mcpConfig core.MCPConfig) (*tools.ToolRegistry, core.MCPManager, error) {
	// Start with the default tool registry
	registry := NewDefaultToolRegistry()

	// Create MCP manager and add MCP tools
	mcpManager, err := createMCPManager(mcpConfig, registry)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create MCP manager: %w", err)
	}

	return registry, mcpManager, nil
}

// NewMCPEnabledRunnerWithConfig creates a runner with MCP capabilities.
func NewMCPEnabledRunnerWithConfig(config RunnerConfig, mcpConfig core.MCPConfig) (core.Runner, core.MCPManager, error) {
	// Create MCP-enabled tool registry and store globally
	registry, mcpManager, err := NewMCPEnabledToolRegistry(mcpConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create MCP-enabled tool registry: %w", err)
	}

	// Store globally for access by agents
	mcpEnabledRegistry = registry
	globalMCPManager = mcpManager

	// Create the runner with the provided config
	runner := core.NewRunnerWithConfig(config)

	return runner, mcpManager, nil
}

// NewMCPEnabledRunnerFromConfig creates a runner with MCP capabilities from AgentFlow config.
func NewMCPEnabledRunnerFromConfig(config *core.Config, agents map[string]core.AgentHandler) (core.Runner, core.MCPManager, error) {
	if !config.MCP.Enabled {
		// If MCP is disabled, create a standard runner
		runnerConfig := core.RunnerConfig{
			Config: config,
			Agents: agents,
		}
		runner := core.NewRunnerWithConfig(runnerConfig)
		return runner, nil, nil
	}

	// MCP is enabled, create MCP-enabled runner
	mcpConfig := convertTomlToMCPConfig(config.MCP)

	// Create MCP-enabled tool registry and store globally
	registry, mcpManager, err := NewMCPEnabledToolRegistry(mcpConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create MCP manager: %w", err)
	}

	// Store globally for access by agents
	mcpEnabledRegistry = registry
	globalMCPManager = mcpManager

	runnerConfig := core.RunnerConfig{
		Config: config,
		Agents: agents,
	}

	runner := core.NewRunnerWithConfig(runnerConfig)

	return runner, mcpManager, nil
}

// NewMCPManagerOnly creates just an MCP manager without a full tool registry.
// This is useful when you want to manage tools separately.
func NewMCPManagerOnly(mcpConfig core.MCPConfig) (core.MCPManager, error) {
	// Create a basic tool registry for MCP integration
	registry := tools.NewToolRegistry()

	// Create and return the MCP manager
	return createMCPManager(mcpConfig, registry)
}

// createMCPManager is a helper to create and configure an MCP manager
func createMCPManager(mcpConfig core.MCPConfig, registry *tools.ToolRegistry) (core.MCPManager, error) {
	// Import the internal MCP package for implementation
	mcpManager, err := mcp.NewMCPManager(mcpConfig, registry, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP manager: %w", err)
	}

	return mcpManager, nil
}

// Package-level variables to store MCP components
var (
	mcpEnabledRegistry *tools.ToolRegistry
	globalMCPManager   core.MCPManager
)

// GetMCPEnabledToolRegistry returns the globally available MCP-enabled tool registry
func GetMCPEnabledToolRegistry() *tools.ToolRegistry {
	if mcpEnabledRegistry == nil {
		// Return default registry if MCP is not enabled
		return NewDefaultToolRegistry()
	}
	return mcpEnabledRegistry
}

// GetGlobalMCPManager returns the globally available MCP manager
func GetGlobalMCPManager() core.MCPManager {
	return globalMCPManager
}

// MCP Tool Discovery and Registration

// DiscoverAndRegisterMCPTools discovers available MCP tools and registers them in the provided registry.
func DiscoverAndRegisterMCPTools(ctx context.Context, registry *tools.ToolRegistry, mcpManager core.MCPManager) error {
	if mcpManager == nil {
		return fmt.Errorf("MCP manager is nil")
	}

	// Refresh tools from all connected servers
	if err := mcpManager.RefreshTools(ctx); err != nil {
		log.Printf("Warning: Failed to refresh tools from some MCP servers: %v", err)
	}

	// Get available tools
	availableTools := mcpManager.GetAvailableTools()

	registeredCount := 0
	for _, toolInfo := range availableTools { // Create MCP tool adapter
		mcpTool, err := CreateMCPToolFromInfo(toolInfo, mcpManager)
		if err != nil {
			log.Printf("Warning: Failed to create adapter for tool '%s': %v", toolInfo.Name, err)
			continue
		}

		// Register with the tool registry
		if err := registry.Register(mcpTool); err != nil {
			log.Printf("Warning: Failed to register MCP tool '%s': %v", toolInfo.Name, err)
			continue
		}

		log.Printf("Registered MCP tool: %s (from %s)", toolInfo.Name, toolInfo.ServerName)
		registeredCount++
	}

	log.Printf("Successfully registered %d MCP tools out of %d available", registeredCount, len(availableTools))
	return nil
}

// AutoDiscoverMCPTools automatically discovers and connects to MCP servers, then registers their tools.
func AutoDiscoverMCPTools(ctx context.Context, registry *tools.ToolRegistry, mcpManager core.MCPManager) error {
	if mcpManager == nil {
		return fmt.Errorf("MCP manager is nil")
	}

	// Discover available servers
	servers, err := mcpManager.DiscoverServers(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover MCP servers: %w", err)
	}

	log.Printf("Discovered %d MCP servers", len(servers))

	// Connect to each discovered server
	connectedCount := 0
	for _, server := range servers {
		if err := mcpManager.Connect(ctx, server.Name); err != nil {
			log.Printf("Warning: Failed to connect to MCP server '%s': %v", server.Name, err)
			continue
		}
		log.Printf("Connected to MCP server: %s (%s:%d)", server.Name, server.Address, server.Port)
		connectedCount++
	}

	log.Printf("Successfully connected to %d out of %d MCP servers", connectedCount, len(servers))

	// Register tools from connected servers
	return DiscoverAndRegisterMCPTools(ctx, registry, mcpManager)
}

// CreateMCPEnabledRegistryWithAutoDiscovery creates a tool registry with automatic MCP discovery.
func CreateMCPEnabledRegistryWithAutoDiscovery(ctx context.Context, mcpConfig core.MCPConfig) (*tools.ToolRegistry, core.MCPManager, error) {
	// Start with the default tool registry
	registry := NewDefaultToolRegistry()

	// Create MCP manager
	mcpManager, err := createMCPManager(mcpConfig, registry)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create MCP manager: %w", err)
	}

	// Auto-discover and register MCP tools
	if err := AutoDiscoverMCPTools(ctx, registry, mcpManager); err != nil {
		// Log warning but don't fail - registry still has built-in tools
		log.Printf("Warning: MCP auto-discovery failed: %v", err)
	}

	return registry, mcpManager, nil
}

// ValidateToolRegistryIntegration checks that MCP tools are properly integrated with the registry.
func ValidateToolRegistryIntegration(registry *tools.ToolRegistry, mcpManager core.MCPManager) error {
	if registry == nil {
		return fmt.Errorf("tool registry is nil")
	}
	if mcpManager == nil {
		return fmt.Errorf("MCP manager is nil")
	}

	// Get all registered tools
	registeredTools := registry.List()

	// Get available MCP tools
	mcpTools := mcpManager.GetAvailableTools()

	// Check that MCP tools are registered
	mcpToolNames := make(map[string]bool)
	for _, tool := range mcpTools {
		mcpToolNames[tool.Name] = true
	}

	registeredMCPCount := 0
	for _, toolName := range registeredTools {
		if mcpToolNames[toolName] {
			registeredMCPCount++
		}
	}

	log.Printf("Tool registry validation: %d registered tools, %d MCP tools, %d MCP tools registered",
		len(registeredTools), len(mcpTools), registeredMCPCount)

	if registeredMCPCount != len(mcpTools) {
		log.Printf("Warning: Not all MCP tools are registered in the tool registry")
	}

	return nil
}

// GetMCPToolsFromRegistry returns only the MCP tools from the registry.
func GetMCPToolsFromRegistry(registry *tools.ToolRegistry, mcpManager core.MCPManager) []string {
	if registry == nil || mcpManager == nil {
		return []string{}
	}

	registeredTools := registry.List()
	mcpTools := mcpManager.GetAvailableTools()

	// Create a map of MCP tool names
	mcpToolNames := make(map[string]bool)
	for _, tool := range mcpTools {
		mcpToolNames[tool.Name] = true
	}

	// Filter registered tools to only include MCP tools
	var mcpRegisteredTools []string
	for _, toolName := range registeredTools {
		if mcpToolNames[toolName] {
			mcpRegisteredTools = append(mcpRegisteredTools, toolName)
		}
	}

	return mcpRegisteredTools
}

// CreateMCPToolFromInfo creates a simple MCP tool implementation for testing.
func CreateMCPToolFromInfo(toolInfo core.MCPToolInfo, mcpManager core.MCPManager) (tools.FunctionTool, error) {
	return &SimpleMCPToolAdapter{
		name:        toolInfo.Name,
		description: toolInfo.Description,
		serverName:  toolInfo.ServerName,
		manager:     mcpManager,
	}, nil
}

// SimpleMCPToolAdapter is a basic MCP tool adapter for registry integration.
type SimpleMCPToolAdapter struct {
	name        string
	description string
	serverName  string
	manager     core.MCPManager
}

func (t *SimpleMCPToolAdapter) Name() string {
	return t.name
}

func (t *SimpleMCPToolAdapter) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
	// For now, this is a placeholder implementation
	// In a real implementation, this would call the actual MCP tool
	return map[string]any{
		"result": fmt.Sprintf("MCP tool '%s' called with args: %v (from server: %s)",
			t.name, args, t.serverName),
		"success": true,
		"server":  t.serverName,
	}, nil
}

// InitRealMCPFactory sets up the core package to use the real MCP implementation
// This should be called during application initialization to enable real MCP functionality
func InitRealMCPFactory() {
	// Set the factory function in core to use our real implementation
	core.SetMCPManagerFactory(func(config core.MCPConfig) (core.MCPManager, error) {
		return NewMCPManagerOnly(config)
	})
}

// init automatically registers the real MCP factory when this package is imported
func init() {
	// Register the real MCP implementation factory with the core package
	core.SetMCPManagerFactory(func(config core.MCPConfig) (core.MCPManager, error) {
		return NewMCPManagerOnly(config)
	})
}
