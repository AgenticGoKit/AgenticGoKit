package core

import (
	"context"
	"strings"
	"testing"
	"time"
)

// --- Test stubs ---

type stubManager struct{}

func (s *stubManager) Connect(ctx context.Context, serverName string) error         { return nil }
func (s *stubManager) Disconnect(serverName string) error                           { return nil }
func (s *stubManager) DisconnectAll() error                                         { return nil }
func (s *stubManager) DiscoverServers(ctx context.Context) ([]MCPServerInfo, error) { return nil, nil }
func (s *stubManager) ListConnectedServers() []string                               { return nil }
func (s *stubManager) GetServerInfo(serverName string) (*MCPServerInfo, error)      { return nil, nil }
func (s *stubManager) RefreshTools(ctx context.Context) error                       { return nil }
func (s *stubManager) GetAvailableTools() []MCPToolInfo                             { return nil }
func (s *stubManager) GetToolsFromServer(serverName string) []MCPToolInfo           { return nil }
func (s *stubManager) HealthCheck(ctx context.Context) map[string]MCPHealthStatus {
	return map[string]MCPHealthStatus{}
}
func (s *stubManager) GetMetrics() MCPMetrics { return MCPMetrics{} }

type stubCache struct{}

func (c *stubCache) Get(ctx context.Context, key MCPCacheKey) (*MCPCachedResult, error) {
	return nil, nil
}
func (c *stubCache) Set(ctx context.Context, key MCPCacheKey, result MCPToolResult, ttl time.Duration) error {
	return nil
}
func (c *stubCache) Delete(ctx context.Context, key MCPCacheKey) error         { return nil }
func (c *stubCache) Clear(ctx context.Context) error                           { return nil }
func (c *stubCache) Exists(ctx context.Context, key MCPCacheKey) (bool, error) { return false, nil }
func (c *stubCache) Stats(ctx context.Context) (MCPCacheStats, error)          { return MCPCacheStats{}, nil }
func (c *stubCache) Cleanup(ctx context.Context) error                         { return nil }
func (c *stubCache) Close() error                                              { return nil }

type stubCacheManager struct{}

func (cm *stubCacheManager) GetCache(toolName, serverName string) MCPCache { return &stubCache{} }
func (cm *stubCacheManager) ExecuteWithCache(ctx context.Context, execution MCPToolExecution) (MCPToolResult, error) {
	return MCPToolResult{ToolName: execution.ToolName, ServerName: execution.ServerName, Success: true}, nil
}
func (cm *stubCacheManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	return nil
}
func (cm *stubCacheManager) GetGlobalStats(ctx context.Context) (MCPCacheStats, error) {
	return MCPCacheStats{}, nil
}
func (cm *stubCacheManager) Shutdown() error                       { return nil }
func (cm *stubCacheManager) Configure(config MCPCacheConfig) error { return nil }

type stubTool struct{ name string }

func (t *stubTool) Name() string { return t.name }
func (t *stubTool) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
	return map[string]any{"ok": true}, nil
}

type stubRegistry struct{ tools map[string]FunctionTool }

func (r *stubRegistry) Register(tool FunctionTool) error {
	if r.tools == nil {
		r.tools = map[string]FunctionTool{}
	}
	r.tools[tool.Name()] = tool
	return nil
}
func (r *stubRegistry) Get(name string) (FunctionTool, bool) { t, ok := r.tools[name]; return t, ok }
func (r *stubRegistry) List() []string {
	out := make([]string, 0, len(r.tools))
	for n := range r.tools {
		out = append(out, n)
	}
	return out
}
func (r *stubRegistry) CallTool(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	if t, ok := r.Get(name); ok {
		return t.Call(ctx, args)
	}
	return nil, nil
}

// --- Tests ---

func TestNoMCPManagerPluginReturnsActionableError(t *testing.T) {
	// Reset global state
	_ = ShutdownMCP()
	// Ensure no factory is registered
	mcpManagerFactory = nil
	err := InitializeMCP(DefaultMCPConfig())
	if err == nil {
		t.Fatalf("expected error when no MCP manager plugin is registered")
	}
	if !strings.Contains(err.Error(), "no MCP manager plugin registered") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetMCPManagerFactoryAllowsInitialization(t *testing.T) {
	_ = ShutdownMCP()
	SetMCPManagerFactory(func(cfg MCPConfig) (MCPManager, error) { return &stubManager{}, nil })
	if err := InitializeMCP(DefaultMCPConfig()); err != nil {
		t.Fatalf("InitializeMCP failed with stub factory: %v", err)
	}
	if GetMCPManager() == nil {
		t.Fatalf("expected MCP manager to be set")
	}
}

func TestNoMCPCacheManagerPluginReturnsActionableError(t *testing.T) {
	// Reset cache globals
	cacheManagerMutex.Lock()
	globalCacheManager = nil
	cacheManagerInitialized = false
	cacheManagerMutex.Unlock()
	mcpCacheManagerFactory = nil
	err := InitializeMCPCacheManager(DefaultMCPCacheConfig())
	if err == nil {
		t.Fatalf("expected error when no MCP cache manager plugin is registered")
	}
	if !strings.Contains(err.Error(), "no MCP cache manager plugin registered") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetMCPCacheManagerFactoryAllowsInitialization(t *testing.T) {
	cacheManagerMutex.Lock()
	globalCacheManager = nil
	cacheManagerInitialized = false
	cacheManagerMutex.Unlock()
	SetMCPCacheManagerFactory(func(cfg MCPCacheConfig) (MCPCacheManager, error) { return &stubCacheManager{}, nil })
	if err := InitializeMCPCacheManager(DefaultMCPCacheConfig()); err != nil {
		t.Fatalf("InitializeMCPCacheManager failed with stub factory: %v", err)
	}
	if GetMCPCacheManager() == nil {
		t.Fatalf("expected MCP cache manager to be set")
	}
}

func TestNoFunctionToolRegistryPluginReturnsActionableError(t *testing.T) {
	mcpRegistryMutex.Lock()
	globalMCPRegistry = nil
	mcpRegistryMutex.Unlock()
	functionToolRegistryFactory = nil
	if _, err := createMCPToolRegistryInternal(); err == nil {
		t.Fatalf("expected error when no FunctionToolRegistry plugin is registered")
	}
}

func TestSetFunctionToolRegistryFactoryAllowsInitialization(t *testing.T) {
	mcpRegistryMutex.Lock()
	globalMCPRegistry = nil
	mcpRegistryMutex.Unlock()
	SetFunctionToolRegistryFactory(func() (FunctionToolRegistry, error) { return &stubRegistry{}, nil })
	if err := InitializeMCPToolRegistry(); err != nil {
		t.Fatalf("InitializeMCPToolRegistry failed with stub factory: %v", err)
	}
	if GetMCPToolRegistry() == nil {
		t.Fatalf("expected FunctionToolRegistry to be set")
	}
}
