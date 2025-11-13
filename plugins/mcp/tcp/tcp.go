package mcp_tcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
)

// tcpMCPManager is a transport plugin that connects to MCP servers over TCP and executes tools.
type tcpMCPManager struct {
	config           core.MCPConfig
	connectedServers map[string]bool
	tools            []core.MCPToolInfo
	mu               sync.RWMutex
}

func newTCPManager(cfg core.MCPConfig) (core.MCPManager, error) {
	return &tcpMCPManager{
		config:           cfg,
		connectedServers: make(map[string]bool),
		tools:            []core.MCPToolInfo{},
	}, nil
}

func (m *tcpMCPManager) Connect(ctx context.Context, serverName string) error {
	// Validate server exists and is TCP
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
	if server.Type != "tcp" {
		return fmt.Errorf("server %s is type %s; tcp plugin only supports tcp", serverName, server.Type)
	}
	// For now, mark as connected; the first real client call will validate connectivity
	m.mu.Lock()
	m.connectedServers[serverName] = true
	m.mu.Unlock()
	return nil
}

func (m *tcpMCPManager) Disconnect(serverName string) error {
	m.mu.Lock()
	delete(m.connectedServers, serverName)
	m.mu.Unlock()
	return nil
}

func (m *tcpMCPManager) DisconnectAll() error {
	m.mu.Lock()
	m.connectedServers = make(map[string]bool)
	m.mu.Unlock()
	return nil
}

func (m *tcpMCPManager) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
	servers := make([]core.MCPServerInfo, 0, len(m.config.Servers))
	for _, s := range m.config.Servers {
		if !s.Enabled || s.Type != "tcp" {
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
			Address: s.Host,
			Port:    s.Port,
			Status:  status,
			Version: "",
		})
	}
	return servers, nil
}

func (m *tcpMCPManager) ListConnectedServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []string
	for name := range m.connectedServers {
		out = append(out, name)
	}
	return out
}

func (m *tcpMCPManager) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
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
				Address: s.Host,
				Port:    s.Port,
				Status:  status,
				Version: "",
			}
			return info, nil
		}
	}
	return nil, fmt.Errorf("server %s not found", serverName)
}

func (m *tcpMCPManager) RefreshTools(ctx context.Context) error {
	// For each tcp server, connect and list tools
	var all []core.MCPToolInfo
	for _, s := range m.config.Servers {
		if !s.Enabled || s.Type != "tcp" {
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

func (m *tcpMCPManager) GetAvailableTools() []core.MCPToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]core.MCPToolInfo(nil), m.tools...)
}

func (m *tcpMCPManager) GetToolsFromServer(serverName string) []core.MCPToolInfo {
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

func (m *tcpMCPManager) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	health := make(map[string]core.MCPHealthStatus)
	for _, s := range m.config.Servers {
		if !s.Enabled || s.Type != "tcp" {
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

// ExecuteTool implements core.MCPToolExecutor using the navigator TCP transport.
func (m *tcpMCPManager) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (core.MCPToolResult, error) {
	// Find server containing this tool (fallback: first tcp server if unknown)
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
			if s.Enabled && s.Type == "tcp" {
				target = s.Name
				break
			}
		}
	}
	if target == "" {
		return core.MCPToolResult{}, fmt.Errorf("no tcp MCP server configured for tool %s", toolName)
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

	c := client.NewClientBuilder().
		WithTCPTransport(server.Host, server.Port).
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

func (m *tcpMCPManager) discoverToolsFromServer(ctx context.Context, serverName string) ([]core.MCPToolInfo, error) {
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
	c := client.NewClientBuilder().
		WithTCPTransport(server.Host, server.Port).
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

func (m *tcpMCPManager) GetMetrics() core.MCPMetrics {
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

// Register the TCP manager factory on import; this overrides the default plugin when both are imported.
func init() {
	core.SetMCPManagerFactory(func(cfg core.MCPConfig) (core.MCPManager, error) {
		return newTCPManager(cfg)
	})
}

