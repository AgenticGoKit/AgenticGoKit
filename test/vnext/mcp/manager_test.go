package mcp_test

import (
	"context"
	"testing"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// mockToolManager implements ToolManager for testing
type mockToolManager struct {
	tools        map[string]vnext.ToolInfo
	mcpServers   map[string]vnext.MCPServerInfo
	healthStatus map[string]vnext.MCPHealthStatus
	metrics      vnext.ToolMetrics
	initialized  bool
}

func newMockToolManager() *mockToolManager {
	return &mockToolManager{
		tools:        make(map[string]vnext.ToolInfo),
		mcpServers:   make(map[string]vnext.MCPServerInfo),
		healthStatus: make(map[string]vnext.MCPHealthStatus),
		metrics: vnext.ToolMetrics{
			ToolMetrics:      make(map[string]vnext.ToolSpecificMetrics),
			MCPServerMetrics: make(map[string]vnext.MCPServerMetrics),
		},
	}
}

func (m *mockToolManager) Execute(ctx context.Context, name string, args map[string]interface{}) (*vnext.ToolResult, error) {
	if !m.initialized {
		return nil, vnext.NewAgentError(vnext.ErrCodeToolNotFound, "Tool manager not initialized")
	}

	if _, exists := m.tools[name]; !exists {
		return nil, vnext.NewAgentError(vnext.ErrCodeToolNotFound, "Tool not found: "+name)
	}

	// Update metrics
	m.metrics.TotalExecutions++
	m.metrics.SuccessfulCalls++

	return &vnext.ToolResult{
		Success: true,
		Content: "mock result for " + name,
	}, nil
}

func (m *mockToolManager) List() []vnext.ToolInfo {
	result := make([]vnext.ToolInfo, 0, len(m.tools))
	for _, tool := range m.tools {
		result = append(result, tool)
	}
	return result
}

func (m *mockToolManager) Available() []string {
	result := make([]string, 0, len(m.tools))
	for name := range m.tools {
		result = append(result, name)
	}
	return result
}

func (m *mockToolManager) IsAvailable(name string) bool {
	_, exists := m.tools[name]
	return exists
}

func (m *mockToolManager) ConnectMCP(ctx context.Context, servers ...vnext.MCPServer) error {
	for _, server := range servers {
		if !server.Enabled {
			continue
		}

		// Add server info
		m.mcpServers[server.Name] = vnext.MCPServerInfo{
			Name:      server.Name,
			Type:      server.Type,
			Address:   server.Address,
			Port:      server.Port,
			Status:    "connected",
			ToolCount: 2, // Mock 2 tools per server
		}

		// Add mock tools for this server
		m.tools[server.Name+"_tool1"] = vnext.ToolInfo{
			Name:        server.Name + "_tool1",
			Description: "Test tool 1 from " + server.Name,
			Category:    "test",
			Parameters:  map[string]interface{}{"param1": "string"},
		}
		m.tools[server.Name+"_tool2"] = vnext.ToolInfo{
			Name:        server.Name + "_tool2",
			Description: "Test tool 2 from " + server.Name,
			Category:    "test",
			Parameters:  map[string]interface{}{"param2": "number"},
		}

		// Set health status
		m.healthStatus[server.Name] = vnext.MCPHealthStatus{
			Status:       "healthy",
			LastCheck:    time.Now(),
			ResponseTime: 10 * time.Millisecond,
			ToolCount:    2,
		}

		// Update metrics
		m.metrics.MCPServerMetrics[server.Name] = vnext.MCPServerMetrics{
			ToolCount:       2,
			LastActivity:    time.Now(),
			SuccessfulCalls: 0,
			FailedCalls:     0,
		}
	}
	return nil
}

func (m *mockToolManager) DisconnectMCP(serverName string) error {
	if _, exists := m.mcpServers[serverName]; !exists {
		return vnext.NewAgentError(vnext.ErrCodeToolNotFound, "Server not found: "+serverName)
	}

	delete(m.mcpServers, serverName)
	delete(m.healthStatus, serverName)
	delete(m.metrics.MCPServerMetrics, serverName)

	// Remove tools from this server
	for name := range m.tools {
		if len(name) > len(serverName) && name[:len(serverName)] == serverName {
			delete(m.tools, name)
		}
	}
	return nil
}

func (m *mockToolManager) DiscoverMCP(ctx context.Context) ([]vnext.MCPServerInfo, error) {
	result := make([]vnext.MCPServerInfo, 0, len(m.mcpServers))
	for _, info := range m.mcpServers {
		result = append(result, info)
	}
	return result, nil
}

func (m *mockToolManager) HealthCheck(ctx context.Context) map[string]vnext.MCPHealthStatus {
	return m.healthStatus
}

func (m *mockToolManager) GetMetrics() vnext.ToolMetrics {
	return m.metrics
}

func (m *mockToolManager) Initialize(ctx context.Context) error {
	m.initialized = true
	return nil
}

func (m *mockToolManager) Shutdown(ctx context.Context) error {
	m.initialized = false
	m.tools = make(map[string]vnext.ToolInfo)
	m.mcpServers = make(map[string]vnext.MCPServerInfo)
	m.healthStatus = make(map[string]vnext.MCPHealthStatus)
	return nil
}

// =============================================================================
// TESTS
// =============================================================================

// TestToolManagerBasic tests basic tool manager functionality
func TestToolManagerBasic(t *testing.T) {
	manager := newMockToolManager()
	if manager == nil {
		t.Fatal("Failed to create tool manager")
	}

	// Test initial state before initialization
	tools := manager.List()
	if len(tools) != 0 {
		t.Errorf("Expected 0 tools initially, got %d", len(tools))
	}

	// Initialize
	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize manager: %v", err)
	}

	if !manager.initialized {
		t.Error("Manager should be initialized")
	}

	// Test Available() on empty manager
	available := manager.Available()
	if len(available) != 0 {
		t.Errorf("Expected 0 available tools, got %d", len(available))
	}
}

// TestToolManagerMCPConnection tests MCP server connection
func TestToolManagerMCPConnection(t *testing.T) {
	manager := newMockToolManager()
	ctx := context.Background()

	// Initialize
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Connect to MCP server
	server := vnext.MCPServer{
		Name:    "test-server",
		Type:    "tcp",
		Address: "localhost",
		Port:    8080,
		Enabled: true,
	}

	if err := manager.ConnectMCP(ctx, server); err != nil {
		t.Fatalf("Failed to connect MCP: %v", err)
	}

	// Check server was added
	discovered, err := manager.DiscoverMCP(ctx)
	if err != nil {
		t.Fatalf("Failed to discover servers: %v", err)
	}

	if len(discovered) != 1 {
		t.Errorf("Expected 1 server, got %d", len(discovered))
	}

	if discovered[0].Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", discovered[0].Name)
	}

	// Check tools were added
	tools := manager.List()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	available := manager.Available()
	if len(available) != 2 {
		t.Errorf("Expected 2 available tool names, got %d", len(available))
	}
}

// TestToolManagerToolExecution tests tool execution
func TestToolManagerToolExecution(t *testing.T) {
	manager := newMockToolManager()
	ctx := context.Background()

	// Initialize
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Connect server and add tools
	server := vnext.MCPServer{
		Name:    "exec-server",
		Type:    "tcp",
		Address: "localhost",
		Port:    8081,
		Enabled: true,
	}

	if err := manager.ConnectMCP(ctx, server); err != nil {
		t.Fatalf("Failed to connect MCP: %v", err)
	}

	// Execute tool
	toolName := "exec-server_tool1"
	result, err := manager.Execute(ctx, toolName, map[string]interface{}{"test": "value"})
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful tool execution")
	}

	if result.Content == nil {
		t.Error("Expected non-nil result content")
	}

	// Check metrics
	metrics := manager.GetMetrics()
	if metrics.TotalExecutions != 1 {
		t.Errorf("Expected 1 execution, got %d", metrics.TotalExecutions)
	}

	if metrics.SuccessfulCalls != 1 {
		t.Errorf("Expected 1 successful call, got %d", metrics.SuccessfulCalls)
	}

	// Try executing non-existent tool
	_, err = manager.Execute(ctx, "nonexistent", nil)
	if err == nil {
		t.Error("Expected error for non-existent tool")
	}
}

// TestToolManagerHealthCheck tests health checking
func TestToolManagerHealthCheck(t *testing.T) {
	manager := newMockToolManager()
	ctx := context.Background()

	// Initialize
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Connect multiple servers
	servers := []vnext.MCPServer{
		{Name: "server1", Type: "tcp", Address: "localhost", Port: 8081, Enabled: true},
		{Name: "server2", Type: "tcp", Address: "localhost", Port: 8082, Enabled: true},
	}

	if err := manager.ConnectMCP(ctx, servers...); err != nil {
		t.Fatalf("Failed to connect MCPs: %v", err)
	}

	// Check health
	health := manager.HealthCheck(ctx)
	if len(health) != 2 {
		t.Errorf("Expected 2 health statuses, got %d", len(health))
	}

	for name, status := range health {
		if status.Status != "healthy" {
			t.Errorf("Expected healthy status for %s, got %s", name, status.Status)
		}

		if status.ToolCount != 2 {
			t.Errorf("Expected 2 tools for %s, got %d", name, status.ToolCount)
		}
	}
}

// TestToolManagerDisconnect tests server disconnection
func TestToolManagerDisconnect(t *testing.T) {
	manager := newMockToolManager()
	ctx := context.Background()

	// Initialize
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Connect server
	server := vnext.MCPServer{
		Name:    "disconnect-server",
		Type:    "tcp",
		Address: "localhost",
		Port:    8083,
		Enabled: true,
	}

	if err := manager.ConnectMCP(ctx, server); err != nil {
		t.Fatalf("Failed to connect MCP: %v", err)
	}

	// Verify tools exist
	if len(manager.Available()) != 2 {
		t.Errorf("Expected 2 tools before disconnect")
	}

	// Disconnect
	if err := manager.DisconnectMCP("disconnect-server"); err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}

	// Verify tools removed
	if len(manager.Available()) != 0 {
		t.Errorf("Expected 0 tools after disconnect, got %d", len(manager.Available()))
	}

	// Verify server removed
	discovered, _ := manager.DiscoverMCP(ctx)
	if len(discovered) != 0 {
		t.Errorf("Expected 0 servers after disconnect, got %d", len(discovered))
	}

	// Try disconnecting non-existent server
	err := manager.DisconnectMCP("nonexistent")
	if err == nil {
		t.Error("Expected error when disconnecting non-existent server")
	}
}

// TestToolManagerShutdown tests shutdown functionality
func TestToolManagerShutdown(t *testing.T) {
	manager := newMockToolManager()
	ctx := context.Background()

	// Initialize
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Connect server
	server := vnext.MCPServer{
		Name:    "shutdown-server",
		Type:    "tcp",
		Address: "localhost",
		Port:    8084,
		Enabled: true,
	}

	if err := manager.ConnectMCP(ctx, server); err != nil {
		t.Fatalf("Failed to connect MCP: %v", err)
	}

	// Shutdown
	if err := manager.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	// Verify state cleared
	if manager.initialized {
		t.Error("Manager should not be initialized after shutdown")
	}

	if len(manager.List()) != 0 {
		t.Error("Expected no tools after shutdown")
	}

	if len(manager.mcpServers) != 0 {
		t.Error("Expected no servers after shutdown")
	}
}

// TestToolManagerIsAvailable tests tool availability checking
func TestToolManagerIsAvailable(t *testing.T) {
	manager := newMockToolManager()
	ctx := context.Background()

	// Initialize
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Connect server
	server := vnext.MCPServer{
		Name:    "avail-server",
		Type:    "tcp",
		Address: "localhost",
		Port:    8085,
		Enabled: true,
	}

	if err := manager.ConnectMCP(ctx, server); err != nil {
		t.Fatalf("Failed to connect MCP: %v", err)
	}

	// Check available tools
	if !manager.IsAvailable("avail-server_tool1") {
		t.Error("Tool should be available")
	}

	if !manager.IsAvailable("avail-server_tool2") {
		t.Error("Tool should be available")
	}

	if manager.IsAvailable("nonexistent") {
		t.Error("Non-existent tool should not be available")
	}
}

// TestToolManagerMetrics tests metrics tracking
func TestToolManagerMetrics(t *testing.T) {
	manager := newMockToolManager()
	ctx := context.Background()

	// Initialize
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Connect server
	server := vnext.MCPServer{
		Name:    "metrics-server",
		Type:    "tcp",
		Address: "localhost",
		Port:    8086,
		Enabled: true,
	}

	if err := manager.ConnectMCP(ctx, server); err != nil {
		t.Fatalf("Failed to connect MCP: %v", err)
	}

	// Execute tools multiple times
	for i := 0; i < 5; i++ {
		_, _ = manager.Execute(ctx, "metrics-server_tool1", nil)
	}

	// Check metrics
	metrics := manager.GetMetrics()
	if metrics.TotalExecutions != 5 {
		t.Errorf("Expected 5 total executions, got %d", metrics.TotalExecutions)
	}

	if metrics.SuccessfulCalls != 5 {
		t.Errorf("Expected 5 successful calls, got %d", metrics.SuccessfulCalls)
	}

	// Check server-specific metrics
	serverMetrics, exists := metrics.MCPServerMetrics["metrics-server"]
	if !exists {
		t.Fatal("Expected metrics for metrics-server")
	}

	if serverMetrics.ToolCount != 2 {
		t.Errorf("Expected 2 tools in metrics, got %d", serverMetrics.ToolCount)
	}
}

// TestToolManagerDisabledServer tests that disabled servers are not connected
func TestToolManagerDisabledServer(t *testing.T) {
	manager := newMockToolManager()
	ctx := context.Background()

	// Initialize
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Try connecting disabled server
	server := vnext.MCPServer{
		Name:    "disabled-server",
		Type:    "tcp",
		Address: "localhost",
		Port:    8087,
		Enabled: false, // Disabled
	}

	if err := manager.ConnectMCP(ctx, server); err != nil {
		t.Fatalf("ConnectMCP should not error for disabled server: %v", err)
	}

	// Verify server was not added
	discovered, _ := manager.DiscoverMCP(ctx)
	if len(discovered) != 0 {
		t.Errorf("Expected 0 servers (disabled), got %d", len(discovered))
	}

	// Verify no tools added
	if len(manager.Available()) != 0 {
		t.Errorf("Expected 0 tools from disabled server, got %d", len(manager.Available()))
	}
}



