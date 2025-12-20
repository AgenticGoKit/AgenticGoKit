package mcp_http_streaming

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

// httpStreamingMCPManager is a transport plugin that connects to MCP servers over HTTP Streaming.
type httpStreamingMCPManager struct {
	config           core.MCPConfig
	connectedServers map[string]bool
	tools            []core.MCPToolInfo
	mu               sync.RWMutex
}

func newHTTPStreamingManager(cfg core.MCPConfig) (core.MCPManager, error) {
	return &httpStreamingMCPManager{
		config:           cfg,
		connectedServers: make(map[string]bool),
		tools:            []core.MCPToolInfo{},
	}, nil
}

func (m *httpStreamingMCPManager) Connect(ctx context.Context, serverName string) error {
	// Validate server exists and is HTTP Streaming
	var server *core.MCPServerConfig
	for i := range m.config.Servers {
		s := &m.config.Servers[i]
		if s.Name == serverName {
			server = s
			break
		}
	}
	if server == nil {
		return fmt.Errorf("server %s not found in configuration", serverName)
	}
	if server.Type != "http_streaming" {
		return fmt.Errorf("server %s is type %s; http_streaming plugin only supports http_streaming", serverName, server.Type)
	}

	// For now, mark as connected; the first real client call will validate connectivity
	m.mu.Lock()
	m.connectedServers[serverName] = true
	m.mu.Unlock()
	return nil
}

func (m *httpStreamingMCPManager) Disconnect(serverName string) error {
	m.mu.Lock()
	delete(m.connectedServers, serverName)
	m.mu.Unlock()
	return nil
}

func (m *httpStreamingMCPManager) DisconnectAll() error {
	m.mu.Lock()
	m.connectedServers = make(map[string]bool)
	m.mu.Unlock()
	return nil
}

func (m *httpStreamingMCPManager) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
	servers := make([]core.MCPServerInfo, 0, len(m.config.Servers))
	for _, s := range m.config.Servers {
		if !s.Enabled || s.Type != "http_streaming" {
			continue
		}
		status := "discovered"
		m.mu.RLock()
		if m.connectedServers[s.Name] {
			status = "connected"
		}
		m.mu.RUnlock()
		servers = append(servers, core.MCPServerInfo{
			Name:    s.Name,
			Type:    s.Type,
			Address: s.Endpoint,
			Status:  status,
			Version: "",
		})
	}
	return servers, nil
}

func (m *httpStreamingMCPManager) ListConnectedServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []string
	for name := range m.connectedServers {
		out = append(out, name)
	}
	return out
}

func (m *httpStreamingMCPManager) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
	for _, s := range m.config.Servers {
		if s.Name == serverName {
			status := "disconnected"
			m.mu.RLock()
			if m.connectedServers[serverName] {
				status = "connected"
			}
			m.mu.RUnlock()
			info := &core.MCPServerInfo{
				Name:    s.Name,
				Type:    s.Type,
				Address: s.Endpoint,
				Status:  status,
				Version: "",
			}
			return info, nil
		}
	}
	return nil, fmt.Errorf("server %s not found", serverName)
}

func (m *httpStreamingMCPManager) RefreshTools(ctx context.Context) error {
	// For each http_streaming server, connect and list tools
	var all []core.MCPToolInfo
	for _, s := range m.config.Servers {
		if !s.Enabled || s.Type != "http_streaming" {
			continue
		}
		tools, err := m.discoverToolsFromServer(ctx, s.Name)
		if err != nil {
			continue
		}
		all = append(all, tools...)
	}
	m.mu.Lock()
	m.tools = all
	m.mu.Unlock()
	return nil
}

func (m *httpStreamingMCPManager) GetAvailableTools() []core.MCPToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]core.MCPToolInfo(nil), m.tools...)
}

func (m *httpStreamingMCPManager) GetToolsFromServer(serverName string) []core.MCPToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []core.MCPToolInfo
	for _, t := range m.tools {
		if t.ServerName == serverName {
			out = append(out, t)
		}
	}
	return out
}

func (m *httpStreamingMCPManager) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	health := make(map[string]core.MCPHealthStatus)
	for _, s := range m.config.Servers {
		if !s.Enabled || s.Type != "http_streaming" {
			continue
		}
		status := core.MCPHealthStatus{Status: "unknown"}
		if err := m.Connect(ctx, s.Name); err == nil {
			status.Status = "healthy"
		} else {
			status.Status = "unhealthy"
			status.Error = err.Error()
		}
		health[s.Name] = status
	}
	return health
}

// ExecuteTool implements core.MCPToolExecutor using HTTP Streaming transport.
func (m *httpStreamingMCPManager) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (core.MCPToolResult, error) {
	// Find server containing this tool (fallback: first http_streaming server if unknown)
	var target string
	m.mu.RLock()
	for _, t := range m.tools {
		if t.Name == toolName {
			target = t.ServerName
			break
		}
	}
	m.mu.RUnlock()
	if target == "" {
		for _, s := range m.config.Servers {
			if s.Enabled && s.Type == "http_streaming" {
				target = s.Name
				break
			}
		}
	}
	if target == "" {
		return core.MCPToolResult{}, fmt.Errorf("no http_streaming MCP server configured for tool %s", toolName)
	}

	// Resolve server config
	var server *core.MCPServerConfig
	for i := range m.config.Servers {
		if m.config.Servers[i].Name == target {
			server = &m.config.Servers[i]
			break
		}
	}
	if server == nil {
		return core.MCPToolResult{}, fmt.Errorf("server config for %s not found", target)
	}

	endpoint := server.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("http://%s:%d", server.Host, server.Port)
	}

	// Parse endpoint to get base URL and path
	baseURL := endpoint
	endpointPath := "/stream" // Default streaming endpoint

	// Create Streaming HTTP transport
	streamingTransport := transport.NewStreamingHTTPTransport(baseURL, endpointPath)

	// Use HTTP Streaming transport
	c := client.NewClientBuilder().
		WithTransport(streamingTransport).
		WithName("agentflow-mcp-client").
		WithVersion("1.0.0").
		WithTimeout(30 * time.Second).
		Build()

	start := time.Now()
	if err := c.Connect(ctx); err != nil {
		return core.MCPToolResult{}, fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	defer c.Disconnect()

	if err := c.Initialize(ctx, mcp.ClientInfo{Name: "agentflow-mcp-client", Version: "1.0.0"}); err != nil {
		return core.MCPToolResult{}, fmt.Errorf("failed to initialize MCP session: %w", err)
	}

	res, err := c.CallTool(ctx, toolName, args)
	if err != nil {
		return core.MCPToolResult{}, fmt.Errorf("tool execution failed: %w", err)
	}

	out := core.MCPToolResult{
		ToolName:   toolName,
		ServerName: target,
		Success:    !res.IsError,
		Duration:   time.Since(start),
	}
	for _, content := range res.Content {
		out.Content = append(out.Content, core.MCPContent{
			Type:     content.Type,
			Text:     content.Text,
			Data:     content.Data,
			MimeType: content.MimeType,
		})
	}
	if res.IsError {
		out.Error = "Tool execution returned error"
	}
	return out, nil
}

func (m *httpStreamingMCPManager) discoverToolsFromServer(ctx context.Context, serverName string) ([]core.MCPToolInfo, error) {
	// Resolve server
	var server *core.MCPServerConfig
	for i := range m.config.Servers {
		if m.config.Servers[i].Name == serverName {
			server = &m.config.Servers[i]
			break
		}
	}
	if server == nil {
		return nil, fmt.Errorf("server %s not found", serverName)
	}

	// Create Streaming HTTP transport for tool discovery
	endpoint := server.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("http://%s:%d", server.Host, server.Port)
	}

	baseURL := endpoint
	endpointPath := "/stream" // Default streaming endpoint
	streamingTransport := transport.NewStreamingHTTPTransport(baseURL, endpointPath)

	c := client.NewClientBuilder().
		WithTransport(streamingTransport).
		WithName("agentflow-mcp-client").
		WithVersion("1.0.0").
		WithTimeout(30 * time.Second).
		Build()

	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	defer c.Disconnect()

	if err := c.Initialize(ctx, mcp.ClientInfo{Name: "agentflow-mcp-client", Version: "1.0.0"}); err != nil {
		return nil, err
	}

	tools, err := c.ListTools(ctx)
	if err != nil {
		return nil, err
	}

	var out []core.MCPToolInfo
	for _, t := range tools {
		out = append(out, core.MCPToolInfo{
			Name:        t.Name,
			Description: t.Description,
			Schema:      t.InputSchema,
			ServerName:  serverName,
		})
	}
	return out, nil
}

func (m *httpStreamingMCPManager) GetMetrics() core.MCPMetrics {
	m.mu.RLock()
	connected := len(m.connectedServers)
	tools := len(m.tools)
	m.mu.RUnlock()
	return core.MCPMetrics{
		ConnectedServers: connected,
		TotalTools:       tools,
		ServerMetrics:    map[string]core.MCPServerMetrics{},
	}
}

// Note: This plugin doesn't auto-register to avoid conflicts with other plugins.
// Import it explicitly when needed for HTTP Streaming transport.

