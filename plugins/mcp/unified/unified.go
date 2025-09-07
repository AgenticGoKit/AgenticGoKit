package mcp_unified

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

// unifiedMCPManager supports multiple transport types: TCP, HTTP SSE, HTTP Streaming, WebSocket, STDIO
type unifiedMCPManager struct {
	config           core.MCPConfig
	connectedServers map[string]bool
	tools            []core.MCPToolInfo
	mu               sync.RWMutex
}

func newUnifiedManager(cfg core.MCPConfig) (core.MCPManager, error) {
	return &unifiedMCPManager{
		config:           cfg,
		connectedServers: make(map[string]bool),
		tools:            []core.MCPToolInfo{},
	}, nil
}

func (m *unifiedMCPManager) Connect(ctx context.Context, serverName string) error {
	// Find server configuration
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
	if !server.Enabled {
		return fmt.Errorf("server %s is disabled", serverName)
	}

	// Mark as connected; actual connectivity is tested during tool operations
	m.mu.Lock()
	m.connectedServers[serverName] = true
	m.mu.Unlock()
	return nil
}

func (m *unifiedMCPManager) Disconnect(serverName string) error {
	m.mu.Lock()
	delete(m.connectedServers, serverName)
	m.mu.Unlock()
	return nil
}

func (m *unifiedMCPManager) DisconnectAll() error {
	m.mu.Lock()
	m.connectedServers = make(map[string]bool)
	m.mu.Unlock()
	return nil
}

func (m *unifiedMCPManager) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
	servers := make([]core.MCPServerInfo, 0, len(m.config.Servers))
	for _, s := range m.config.Servers {
		if !s.Enabled {
			continue
		}
		status := "discovered"
		m.mu.RLock()
		if m.connectedServers[s.Name] {
			status = "connected"
		}
		m.mu.RUnlock()

		address := s.Host
		if s.Endpoint != "" {
			address = s.Endpoint
		}

		servers = append(servers, core.MCPServerInfo{
			Name:    s.Name,
			Type:    s.Type,
			Address: address,
			Port:    s.Port,
			Status:  status,
			Version: "",
		})
	}
	return servers, nil
}

func (m *unifiedMCPManager) ListConnectedServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []string
	for name := range m.connectedServers {
		out = append(out, name)
	}
	return out
}

func (m *unifiedMCPManager) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
	for _, s := range m.config.Servers {
		if s.Name == serverName {
			status := "disconnected"
			m.mu.RLock()
			if m.connectedServers[serverName] {
				status = "connected"
			}
			m.mu.RUnlock()

			address := s.Host
			if s.Endpoint != "" {
				address = s.Endpoint
			}

			info := &core.MCPServerInfo{
				Name:    s.Name,
				Type:    s.Type,
				Address: address,
				Port:    s.Port,
				Status:  status,
				Version: "",
			}
			return info, nil
		}
	}
	return nil, fmt.Errorf("server %s not found", serverName)
}

func (m *unifiedMCPManager) RefreshTools(ctx context.Context) error {
	// For each enabled server, connect and list tools
	var all []core.MCPToolInfo
	for _, s := range m.config.Servers {
		if !s.Enabled {
			continue
		}
		tools, err := m.discoverToolsFromServer(ctx, s.Name)
		if err != nil {
			core.Logger().Warn().
				Str("server_name", s.Name).
				Err(err).
				Msg("Failed to discover tools from server")
			continue
		}
		all = append(all, tools...)
	}
	m.mu.Lock()
	m.tools = all
	m.mu.Unlock()
	return nil
}

func (m *unifiedMCPManager) GetAvailableTools() []core.MCPToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]core.MCPToolInfo(nil), m.tools...)
}

func (m *unifiedMCPManager) GetToolsFromServer(serverName string) []core.MCPToolInfo {
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

func (m *unifiedMCPManager) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	health := make(map[string]core.MCPHealthStatus)
	for _, s := range m.config.Servers {
		if !s.Enabled {
			continue
		}
		status := core.MCPHealthStatus{Status: "unknown", LastCheck: time.Now()}

		// Try to create a client and connect briefly for health check
		client, err := m.createClientForServer(&s)
		if err != nil {
			status.Status = "unhealthy"
			status.Error = fmt.Sprintf("Failed to create client: %v", err)
		} else {
			start := time.Now()
			if err := client.Connect(ctx); err != nil {
				status.Status = "unhealthy"
				status.Error = fmt.Sprintf("Connection failed: %v", err)
			} else {
				status.Status = "healthy"
				status.ResponseTime = time.Since(start)
				client.Disconnect()
			}
		}
		health[s.Name] = status
	}
	return health
}

// ExecuteTool implements core.MCPToolExecutor for unified transport support
func (m *unifiedMCPManager) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (core.MCPToolResult, error) {
	// Find server containing this tool
	var target string
	m.mu.RLock()
	for _, t := range m.tools {
		if t.Name == toolName {
			target = t.ServerName
			break
		}
	}
	m.mu.RUnlock()

	// If tool not found in cache, try first enabled server
	if target == "" {
		for _, s := range m.config.Servers {
			if s.Enabled {
				target = s.Name
				break
			}
		}
	}
	if target == "" {
		return core.MCPToolResult{}, fmt.Errorf("no enabled MCP server found for tool %s", toolName)
	}

	// Find server config
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

	// Create client for this server
	client, err := m.createClientForServer(server)
	if err != nil {
		return core.MCPToolResult{}, fmt.Errorf("failed to create client: %w", err)
	}

	start := time.Now()
	if err := client.Connect(ctx); err != nil {
		return core.MCPToolResult{}, fmt.Errorf("failed to connect to MCP server %s: %w", target, err)
	}
	defer client.Disconnect()

	if err := client.Initialize(ctx, mcp.ClientInfo{Name: "agentflow-mcp-client", Version: "1.0.0"}); err != nil {
		return core.MCPToolResult{}, fmt.Errorf("failed to initialize MCP session: %w", err)
	}

	res, err := client.CallTool(ctx, toolName, args)
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
		if len(res.Content) > 0 && res.Content[0].Text != "" {
			out.Error = res.Content[0].Text
		}
	}
	return out, nil
}

func (m *unifiedMCPManager) discoverToolsFromServer(ctx context.Context, serverName string) ([]core.MCPToolInfo, error) {
	// Find server config
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

	// Create client for this server
	client, err := m.createClientForServer(server)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for %s: %w", serverName, err)
	}

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", serverName, err)
	}
	defer client.Disconnect()

	if err := client.Initialize(ctx, mcp.ClientInfo{Name: "agentflow-mcp-client", Version: "1.0.0"}); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP session with %s: %w", serverName, err)
	}

	tools, err := client.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools from %s: %w", serverName, err)
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

// createClientForServer creates appropriate client based on server type
func (m *unifiedMCPManager) createClientForServer(server *core.MCPServerConfig) (*client.Client, error) {
	builder := client.NewClientBuilder().
		WithName("agentflow-mcp-client").
		WithVersion("1.0.0").
		WithTimeout(30 * time.Second)

	switch server.Type {
	case "tcp":
		return builder.WithTCPTransport(server.Host, server.Port).Build(), nil

	case "http_sse":
		endpoint := server.Endpoint
		if endpoint == "" {
			if server.Host != "" && server.Port > 0 {
				endpoint = fmt.Sprintf("http://%s:%d", server.Host, server.Port)
			} else {
				return nil, fmt.Errorf("http_sse server %s requires either endpoint or host:port configuration", server.Name)
			}
		}
		// Create SSE transport
		sseTransport := transport.NewSSETransport(endpoint, "/sse")
		return builder.WithTransport(sseTransport).Build(), nil

	case "http_streaming":
		endpoint := server.Endpoint
		if endpoint == "" {
			if server.Host != "" && server.Port > 0 {
				endpoint = fmt.Sprintf("http://%s:%d", server.Host, server.Port)
			} else {
				return nil, fmt.Errorf("http_streaming server %s requires either endpoint or host:port configuration", server.Name)
			}
		}
		// Create Streaming HTTP transport
		streamingTransport := transport.NewStreamingHTTPTransport(endpoint, "/stream")
		return builder.WithTransport(streamingTransport).Build(), nil

	case "websocket":
		url := fmt.Sprintf("ws://%s:%d", server.Host, server.Port)
		return builder.WithWebSocketTransport(url).Build(), nil

	case "stdio":
		return builder.WithSTDIOTransport(server.Command, []string{}).Build(), nil

	default:
		return nil, fmt.Errorf("unsupported transport type: %s", server.Type)
	}
}

func (m *unifiedMCPManager) GetMetrics() core.MCPMetrics {
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

// Register the unified manager factory - this replaces other transport plugins
func init() {
	core.SetMCPManagerFactory(func(cfg core.MCPConfig) (core.MCPManager, error) {
		return newUnifiedManager(cfg)
	})
}
