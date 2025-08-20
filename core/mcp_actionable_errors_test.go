package core

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// ensureCleanState resets global MCP state before each test
func ensureCleanState(t *testing.T) {
	t.Helper()
	_ = ShutdownMCP() // ignore error; just ensure globals are cleared
}

func TestInitializeMCP_NoFactory_ReturnsActionableError(t *testing.T) {
	ensureCleanState(t)

	err := InitializeMCP(DefaultMCPConfig())
	if err == nil {
		t.Fatalf("expected error when no MCP manager plugin registered")
	}
	if !strings.Contains(err.Error(), "no MCP manager plugin registered") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestInitializeMCPCacheManager_NoFactory_ReturnsActionableError(t *testing.T) {
	ensureCleanState(t)

	err := InitializeMCPCacheManager(DefaultMCPCacheConfig())
	if err == nil {
		t.Fatalf("expected error when no MCP cache manager plugin registered")
	}
	if !strings.Contains(err.Error(), "no MCP cache manager plugin registered") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestInitializeMCPToolRegistry_NoFactory_ReturnsActionableError(t *testing.T) {
	ensureCleanState(t)

	err := InitializeMCPToolRegistry()
	if err == nil {
		t.Fatalf("expected error when no MCP function tool registry plugin registered")
	}
	if !strings.Contains(err.Error(), "no MCP function tool registry plugin registered") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestExecuteMCPTool_NoManager_ReturnsError(t *testing.T) {
	ensureCleanState(t)

	_, err := ExecuteMCPTool(context.Background(), "some_tool", map[string]interface{}{"a": 1})
	if err == nil {
		t.Fatalf("expected error when MCP manager not initialized")
	}
	if !strings.Contains(err.Error(), "MCP manager not initialized") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

// localStubManager implements MCPManager but not MCPToolExecutor
type localStubManager struct{}

func (s *localStubManager) Connect(ctx context.Context, serverName string) error { return nil }
func (s *localStubManager) Disconnect(serverName string) error                   { return nil }
func (s *localStubManager) DisconnectAll() error                                 { return nil }
func (s *localStubManager) DiscoverServers(ctx context.Context) ([]MCPServerInfo, error) {
	return nil, nil
}
func (s *localStubManager) ListConnectedServers() []string                          { return nil }
func (s *localStubManager) GetServerInfo(serverName string) (*MCPServerInfo, error) { return nil, nil }
func (s *localStubManager) RefreshTools(ctx context.Context) error                  { return nil }
func (s *localStubManager) GetAvailableTools() []MCPToolInfo {
	return []MCPToolInfo{{Name: "t1", Description: "d", Schema: map[string]any{}, ServerName: "s"}}
}
func (s *localStubManager) GetToolsFromServer(serverName string) []MCPToolInfo { return nil }
func (s *localStubManager) HealthCheck(ctx context.Context) map[string]MCPHealthStatus {
	return map[string]MCPHealthStatus{}
}
func (s *localStubManager) GetMetrics() MCPMetrics { return MCPMetrics{} }

func TestExecuteMCPTool_ManagerWithoutExecutor_ReturnsActionableError(t *testing.T) {
	ensureCleanState(t)
	// Register stub manager factory
	SetMCPManagerFactory(func(cfg MCPConfig) (MCPManager, error) { return &localStubManager{}, nil })
	defer SetMCPManagerFactory(nil)

	if err := InitializeMCP(DefaultMCPConfig()); err != nil {
		t.Fatalf("init mcp: %v", err)
	}

	_, err := ExecuteMCPTool(context.Background(), "some_tool", map[string]interface{}{"a": 1})
	if err == nil {
		t.Fatalf("expected actionable error for missing MCPToolExecutor support")
	}
	if !strings.Contains(err.Error(), "does not support direct tool execution") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

// localStubRegistry is a simple in-memory FunctionToolRegistry
type localStubRegistry struct{ names []string }

func (r *localStubRegistry) Register(tool FunctionTool) error {
	if tool == nil {
		return fmt.Errorf("nil tool")
	}
	r.names = append(r.names, tool.Name())
	return nil
}
func (r *localStubRegistry) Get(name string) (FunctionTool, bool) { return nil, false }
func (r *localStubRegistry) List() []string                       { return r.names }
func (r *localStubRegistry) CallTool(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestRegisterMCPToolsWithRegistry_RegistersTools(t *testing.T) {
	ensureCleanState(t)
	// Register factories
	SetMCPManagerFactory(func(cfg MCPConfig) (MCPManager, error) { return &localStubManager{}, nil })
	defer SetMCPManagerFactory(nil)

	reg := &localStubRegistry{}
	SetFunctionToolRegistryFactory(func() (FunctionToolRegistry, error) { return reg, nil })
	defer SetFunctionToolRegistryFactory(nil)

	if err := InitializeMCP(DefaultMCPConfig()); err != nil {
		t.Fatalf("init mcp: %v", err)
	}
	if err := InitializeMCPToolRegistry(); err != nil {
		t.Fatalf("init registry: %v", err)
	}

	if err := RegisterMCPToolsWithRegistry(context.Background()); err != nil {
		t.Fatalf("register tools: %v", err)
	}
	if len(reg.names) == 0 {
		t.Fatalf("expected at least one tool registered")
	}
}
