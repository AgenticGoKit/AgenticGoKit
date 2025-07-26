// Package mcp provides internal implementation for Model Context Protocol (MCP) integration.
package mcp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kunalkushwaha/AgenticGoKit/core"
	"github.com/kunalkushwaha/AgenticGoKit/internal/tools"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/discovery"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

// MCPManagerImpl provides the concrete implementation of core.MCPManager.
type MCPManagerImpl struct {
	config       core.MCPConfig
	toolRegistry *tools.ToolRegistry
	logger       *log.Logger

	// Connection management
	clients    map[string]*client.Client
	transports map[string]transport.Transport

	// Discovery
	discovery *discovery.Discovery

	// Tool management
	mcpTools    map[string]*MCPTool // tool name -> MCPTool
	serverTools map[string][]string // server name -> tool names

	// Metrics and monitoring
	metrics     MCPMetricsImpl
	serverStats map[string]*ServerStats

	// Thread safety
	mu sync.RWMutex
}

// ServerStats tracks statistics for an individual server.
type ServerStats struct {
	ConnectedAt     time.Time
	LastActivity    time.Time
	ToolExecutions  int64
	SuccessfulCalls int64
	FailedCalls     int64
	TotalLatency    time.Duration
	ToolCount       int
	mu              sync.RWMutex
}

// MCPMetricsImpl implements the metrics tracking for MCP operations.
type MCPMetricsImpl struct {
	ConnectedServers int                              `json:"connected_servers"`
	TotalTools       int                              `json:"total_tools"`
	ToolExecutions   int64                            `json:"tool_executions"`
	TotalLatency     time.Duration                    `json:"total_latency"`
	TotalErrors      int64                            `json:"total_errors"`
	ServerMetrics    map[string]core.MCPServerMetrics `json:"server_metrics"`
	mu               sync.RWMutex
}

// createMCPManager creates a new MCP manager with the given configuration.
func createMCPManager(config core.MCPConfig, registry *tools.ToolRegistry, logger *log.Logger) (*MCPManagerImpl, error) {
	if logger == nil {
		logger = log.Default()
	}

	// Validate configuration
	if err := validateMCPConfig(config); err != nil {
		return nil, fmt.Errorf("invalid MCP configuration: %w", err)
	}
	manager := &MCPManagerImpl{
		config:       config,
		toolRegistry: registry,
		logger:       logger,
		clients:      make(map[string]*client.Client),
		transports:   make(map[string]transport.Transport),
		discovery:    discovery.NewDiscovery(logger),
		mcpTools:     make(map[string]*MCPTool),
		serverTools:  make(map[string][]string),
		serverStats:  make(map[string]*ServerStats),
		metrics: MCPMetricsImpl{
			ServerMetrics: make(map[string]core.MCPServerMetrics),
		},
	}

	// Note: Discovery timeout configuration would need to be implemented
	// in the MCP Navigator library. For now, we'll use the default timeout.

	return manager, nil
}

// Connect establishes a connection to the specified MCP server.
func (m *MCPManagerImpl) Connect(ctx context.Context, serverName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already connected
	if _, exists := m.clients[serverName]; exists {
		return nil // Already connected
	}

	// Find server configuration
	var serverConfig *core.MCPServerConfig
	for _, config := range m.config.Servers {
		if config.Name == serverName {
			serverConfig = &config
			break
		}
	}

	if serverConfig == nil {
		return fmt.Errorf("server configuration not found for '%s'", serverName)
	}

	if !serverConfig.Enabled {
		return fmt.Errorf("server '%s' is disabled", serverName)
	}

	// Create transport based on server type
	transport, err := m.createTransport(*serverConfig)
	if err != nil {
		return fmt.Errorf("failed to create transport for server '%s': %w", serverName, err)
	}

	// Create client
	clientConfig := client.ClientConfig{
		Name:    "AgentFlow",
		Version: "1.0.0",
		Logger:  m.logger,
		Timeout: m.config.ConnectionTimeout,
	}
	mcpClient := client.NewClient(transport, clientConfig)

	// Connect with timeout
	connectCtx, cancel := context.WithTimeout(ctx, m.config.ConnectionTimeout)
	defer cancel()

	if err := mcpClient.Connect(connectCtx); err != nil {
		return fmt.Errorf("failed to connect to server '%s': %w", serverName, err)
	}

	// Initialize the MCP protocol
	clientInfo := mcp.ClientInfo{
		Name:    "AgentFlow",
		Version: "1.0.0",
	}

	if err := mcpClient.Initialize(connectCtx, clientInfo); err != nil {
		mcpClient.Disconnect()
		return fmt.Errorf("failed to initialize MCP protocol with server '%s': %w", serverName, err)
	}

	// Store connection
	m.clients[serverName] = mcpClient
	m.transports[serverName] = transport
	m.serverStats[serverName] = &ServerStats{
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
	}

	m.logger.Printf("Successfully connected to MCP server: %s", serverName)

	// Register tools from this server
	if err := m.registerServerTools(ctx, serverName); err != nil {
		m.logger.Printf("Warning: failed to register tools from server '%s': %v", serverName, err)
		// Don't fail the connection for tool registration issues
	}

	return nil
}

// Disconnect closes the connection to the specified server.
func (m *MCPManagerImpl) Disconnect(serverName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[serverName]
	if !exists {
		return nil // Not connected
	}

	// Unregister tools from this server
	m.unregisterServerTools(serverName)

	// Close connection
	if err := client.Disconnect(); err != nil {
		m.logger.Printf("Error disconnecting from server '%s': %v", serverName, err)
	}

	// Clean up
	delete(m.clients, serverName)
	delete(m.transports, serverName)
	delete(m.serverStats, serverName)

	m.logger.Printf("Disconnected from MCP server: %s", serverName)
	return nil
}

// DisconnectAll closes all MCP connections.
func (m *MCPManagerImpl) DisconnectAll() error {
	m.mu.RLock()
	serverNames := make([]string, 0, len(m.clients))
	for name := range m.clients {
		serverNames = append(serverNames, name)
	}
	m.mu.RUnlock()

	var lastErr error
	for _, serverName := range serverNames {
		if err := m.Disconnect(serverName); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// DiscoverServers discovers available MCP servers.
func (m *MCPManagerImpl) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
	if !m.config.EnableDiscovery {
		return nil, fmt.Errorf("server discovery is disabled")
	}

	m.logger.Println("Starting MCP server discovery...")

	var allServers []core.MCPServerInfo

	// Default ports to scan if not specified
	ports := m.config.ScanPorts
	if len(ports) == 0 {
		ports = []int{8080, 8081, 8090, 8100, 3000, 3001} // Common MCP ports
	}

	// Discover TCP servers
	discoveredServers := m.discovery.DiscoverTCPServers(ctx, "localhost", ports)

	for _, server := range discoveredServers {
		serverInfo := core.MCPServerInfo{
			Name:        server.Name,
			Type:        server.Type,
			Address:     server.Address,
			Port:        server.Port,
			Description: server.Description,
			Status:      "discovered",
		}
		allServers = append(allServers, serverInfo)
	}

	m.logger.Printf("Discovered %d MCP servers", len(allServers))
	return allServers, nil
}

// ListConnectedServers returns the names of all connected servers.
func (m *MCPManagerImpl) ListConnectedServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]string, 0, len(m.clients))
	for name := range m.clients {
		servers = append(servers, name)
	}
	return servers
}

// GetServerInfo returns information about a specific server.
func (m *MCPManagerImpl) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[serverName]
	if !exists {
		return nil, fmt.Errorf("server '%s' is not connected", serverName)
	}

	serverInfo := client.GetServerInfo()
	if serverInfo == nil {
		return nil, fmt.Errorf("no server info available for '%s'", serverName)
	}

	// Find server configuration
	var config *core.MCPServerConfig
	for _, cfg := range m.config.Servers {
		if cfg.Name == serverName {
			config = &cfg
			break
		}
	}

	info := &core.MCPServerInfo{
		Name:    serverInfo.Name,
		Version: serverInfo.Version,
		Status:  "connected",
	}

	if config != nil {
		info.Type = config.Type
		info.Address = config.Host
		info.Port = config.Port
	}

	// Add capabilities
	caps := client.GetServerCapabilities()
	if caps != nil {
		info.Capabilities = make(map[string]interface{})
		if caps.Tools != nil {
			info.Capabilities["tools"] = true
		}
		if caps.Resources != nil {
			info.Capabilities["resources"] = true
		}
		if caps.Prompts != nil {
			info.Capabilities["prompts"] = true
		}
	}

	return info, nil
}

// RefreshTools refreshes the tool list from all connected servers.
func (m *MCPManagerImpl) RefreshTools(ctx context.Context) error {
	m.mu.RLock()
	serverNames := make([]string, 0, len(m.clients))
	for name := range m.clients {
		serverNames = append(serverNames, name)
	}
	m.mu.RUnlock()

	var lastErr error
	for _, serverName := range serverNames {
		if err := m.registerServerTools(ctx, serverName); err != nil {
			m.logger.Printf("Failed to refresh tools from server '%s': %v", serverName, err)
			lastErr = err
		}
	}

	return lastErr
}

// GetAvailableTools returns information about all available MCP tools.
func (m *MCPManagerImpl) GetAvailableTools() []core.MCPToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]core.MCPToolInfo, 0, len(m.mcpTools))
	for _, tool := range m.mcpTools {
		tools = append(tools, tool.ToMCPToolInfo())
	}
	return tools
}

// GetToolsFromServer returns tools available from a specific server.
func (m *MCPManagerImpl) GetToolsFromServer(serverName string) []core.MCPToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	toolNames, exists := m.serverTools[serverName]
	if !exists {
		return nil
	}

	tools := make([]core.MCPToolInfo, 0, len(toolNames))
	for _, toolName := range toolNames {
		if tool, exists := m.mcpTools[toolName]; exists {
			tools = append(tools, tool.ToMCPToolInfo())
		}
	}
	return tools
}

// HealthCheck performs a health check on all connected servers.
func (m *MCPManagerImpl) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	m.mu.RLock()
	clients := make(map[string]*client.Client)
	for name, client := range m.clients {
		clients[name] = client
	}
	m.mu.RUnlock()

	results := make(map[string]core.MCPHealthStatus)

	for serverName, client := range clients {
		status := m.checkServerHealth(ctx, serverName, client)
		results[serverName] = status
	}

	return results
}

// GetMetrics returns current MCP metrics.
func (m *MCPManagerImpl) GetMetrics() core.MCPMetrics {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	// Update current counts
	m.mu.RLock()
	connectedServers := len(m.clients)
	totalTools := len(m.mcpTools)
	m.mu.RUnlock()

	// Calculate average latency
	var avgLatency time.Duration
	if m.metrics.ToolExecutions > 0 {
		avgLatency = m.metrics.TotalLatency / time.Duration(m.metrics.ToolExecutions)
	}

	// Calculate error rate
	errorRate := 0.0
	if m.metrics.ToolExecutions > 0 {
		errorRate = float64(m.metrics.TotalErrors) / float64(m.metrics.ToolExecutions)
	}

	return core.MCPMetrics{
		ConnectedServers: connectedServers,
		TotalTools:       totalTools,
		ToolExecutions:   m.metrics.ToolExecutions,
		AverageLatency:   avgLatency,
		ErrorRate:        errorRate,
		ServerMetrics:    m.metrics.ServerMetrics,
	}
}

// Helper method implementations

// validateMCPConfig validates the MCP configuration.
func validateMCPConfig(config core.MCPConfig) error {
	if config.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection timeout must be positive")
	}
	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	if config.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}
	if config.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}

	// Validate server configurations
	serverNames := make(map[string]bool)
	for _, server := range config.Servers {
		if server.Name == "" {
			return fmt.Errorf("server name cannot be empty")
		}
		if serverNames[server.Name] {
			return fmt.Errorf("duplicate server name: %s", server.Name)
		}
		serverNames[server.Name] = true

		if err := validateServerConfig(server); err != nil {
			return fmt.Errorf("invalid server config for '%s': %w", server.Name, err)
		}
	}

	return nil
}

// validateServerConfig validates an individual server configuration.
func validateServerConfig(config core.MCPServerConfig) error {
	switch config.Type {
	case "tcp":
		if config.Host == "" {
			return fmt.Errorf("TCP server must specify host")
		}
		if config.Port <= 0 || config.Port > 65535 {
			return fmt.Errorf("TCP server must specify valid port (1-65535)")
		}
	case "stdio":
		if config.Command == "" {
			return fmt.Errorf("STDIO server must specify command")
		}
	case "websocket":
		if config.Host == "" {
			return fmt.Errorf("WebSocket server must specify host")
		}
		if config.Port <= 0 || config.Port > 65535 {
			return fmt.Errorf("WebSocket server must specify valid port (1-65535)")
		}
	case "docker":
		// Docker validation could be added here
	default:
		return fmt.Errorf("unsupported server type: %s", config.Type)
	}
	return nil
}

// createTransport creates a transport based on the server configuration.
func (m *MCPManagerImpl) createTransport(config core.MCPServerConfig) (transport.Transport, error) {
	switch config.Type {
	case "tcp":
		return transport.NewTCPTransport(config.Host, config.Port), nil
	case "stdio":
		// Parse command into command and args
		// For simplicity, we'll assume command is just the executable name
		// In a real implementation, you'd want to parse shell commands properly
		return transport.NewStdioTransport(config.Command, []string{}), nil
	case "websocket":
		url := fmt.Sprintf("ws://%s:%d", config.Host, config.Port)
		return transport.NewWebSocketTransport(url), nil
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", config.Type)
	}
}

// registerServerTools registers all tools from a specific server.
func (m *MCPManagerImpl) registerServerTools(ctx context.Context, serverName string) error {
	client, exists := m.clients[serverName]
	if !exists {
		return fmt.Errorf("server '%s' is not connected", serverName)
	}

	// List available tools from the server
	tools, err := client.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools from server '%s': %w", serverName, err)
	}

	m.logger.Printf("Found %d tools on server '%s'", len(tools), serverName)

	// Unregister existing tools from this server first
	m.unregisterServerTools(serverName)

	// Register new tools
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		mcpTool := NewMCPTool(tool, serverName, client, m)
		toolName := mcpTool.Name()

		// Register with AgentFlow tool registry
		if err := m.toolRegistry.Register(mcpTool); err != nil {
			m.logger.Printf("Warning: failed to register tool '%s': %v", toolName, err)
			continue
		}

		// Track in our internal registry
		m.mcpTools[toolName] = mcpTool
		toolNames = append(toolNames, toolName)

		m.logger.Printf("Registered MCP tool: %s", toolName)
	}

	// Update server tools mapping
	m.serverTools[serverName] = toolNames

	// Update server statistics
	if stats, exists := m.serverStats[serverName]; exists {
		stats.mu.Lock()
		stats.ToolCount = len(toolNames)
		stats.LastActivity = time.Now()
		stats.mu.Unlock()
	}

	return nil
}

// unregisterServerTools unregisters all tools from a specific server.
func (m *MCPManagerImpl) unregisterServerTools(serverName string) {
	toolNames, exists := m.serverTools[serverName]
	if !exists {
		return
	}

	for _, toolName := range toolNames {
		// Remove from AgentFlow tool registry
		// Note: The current tool registry doesn't have an Unregister method
		// We'll just remove from our internal tracking for now
		delete(m.mcpTools, toolName)
		m.logger.Printf("Unregistered MCP tool: %s", toolName)
	}

	delete(m.serverTools, serverName)
}

// checkServerHealth performs a health check on a specific server.
func (m *MCPManagerImpl) checkServerHealth(ctx context.Context, serverName string, client *client.Client) core.MCPHealthStatus {
	start := time.Now()

	status := core.MCPHealthStatus{
		Status:    "unknown",
		LastCheck: start,
	}

	// Check basic connection
	if !client.IsConnected() {
		status.Status = "unhealthy"
		status.Error = "not connected"
		return status
	}

	if !client.IsInitialized() {
		status.Status = "unhealthy"
		status.Error = "not initialized"
		return status
	}

	// Try to list tools as a health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tools, err := client.ListTools(healthCtx)
	responseTime := time.Since(start)
	status.ResponseTime = responseTime

	if err != nil {
		status.Status = "unhealthy"
		status.Error = err.Error()
		return status
	}

	status.Status = "healthy"
	status.ToolCount = len(tools)
	return status
}

// recordToolCall records metrics for a tool call.
func (m *MCPManagerImpl) recordToolCall(serverName string, startTime time.Time) {
	m.metrics.mu.Lock()
	m.metrics.ToolExecutions++
	m.metrics.mu.Unlock()

	if stats, exists := m.serverStats[serverName]; exists {
		stats.mu.Lock()
		stats.ToolExecutions++
		stats.LastActivity = time.Now()
		stats.mu.Unlock()
	}
}

// recordToolSuccess records metrics for a successful tool call.
func (m *MCPManagerImpl) recordToolSuccess(serverName string, startTime time.Time) {
	duration := time.Since(startTime)

	m.metrics.mu.Lock()
	m.metrics.TotalLatency += duration
	m.metrics.mu.Unlock()

	if stats, exists := m.serverStats[serverName]; exists {
		stats.mu.Lock()
		stats.SuccessfulCalls++
		stats.TotalLatency += duration
		stats.LastActivity = time.Now()
		stats.mu.Unlock()
	}
}

// recordToolError records metrics for a failed tool call.
func (m *MCPManagerImpl) recordToolError(serverName string, startTime time.Time, err error) {
	duration := time.Since(startTime)

	m.metrics.mu.Lock()
	m.metrics.TotalErrors++
	m.metrics.TotalLatency += duration
	m.metrics.mu.Unlock()

	if stats, exists := m.serverStats[serverName]; exists {
		stats.mu.Lock()
		stats.FailedCalls++
		stats.TotalLatency += duration
		stats.LastActivity = time.Now()
		stats.mu.Unlock()
	}
	m.logger.Printf("MCP tool execution failed on server '%s': %v", serverName, err)
}
