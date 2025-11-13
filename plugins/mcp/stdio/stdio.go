package stdiomcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/agenticgokit/agenticgokit/core"
)

// Placeholder STDIO MCP plugin. Implements core.MCPManager minimally and reserved for future expansion.
// For now, it mirrors the default behavior but is split to a transport-specific package name.
type stdioMCPManager struct {
	config           core.MCPConfig
	connectedServers map[string]bool
	tools            []core.MCPToolInfo
	mu               sync.RWMutex
}

func newSTDIOManager(cfg core.MCPConfig) (core.MCPManager, error) {
	return &stdioMCPManager{
		config:           cfg,
		connectedServers: make(map[string]bool),
		tools:            []core.MCPToolInfo{},
	}, nil
}

func (m *stdioMCPManager) Connect(ctx context.Context, serverName string) error {
	return fmt.Errorf("stdio MCP not implemented")
}
func (m *stdioMCPManager) Disconnect(serverName string) error { return nil }
func (m *stdioMCPManager) DisconnectAll() error               { return nil }
func (m *stdioMCPManager) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
	return nil, fmt.Errorf("stdio MCP not implemented")
}
func (m *stdioMCPManager) ListConnectedServers() []string { return nil }
func (m *stdioMCPManager) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
	return nil, fmt.Errorf("not found")
}
func (m *stdioMCPManager) RefreshTools(ctx context.Context) error                  { return nil }
func (m *stdioMCPManager) GetAvailableTools() []core.MCPToolInfo                   { return nil }
func (m *stdioMCPManager) GetToolsFromServer(serverName string) []core.MCPToolInfo { return nil }
func (m *stdioMCPManager) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	return map[string]core.MCPHealthStatus{}
}
func (m *stdioMCPManager) GetMetrics() core.MCPMetrics { return core.MCPMetrics{} }

// Register the STDIO manager factory on import.
func init() {
	core.SetMCPManagerFactory(func(cfg core.MCPConfig) (core.MCPManager, error) {
		return newSTDIOManager(cfg)
	})
}

