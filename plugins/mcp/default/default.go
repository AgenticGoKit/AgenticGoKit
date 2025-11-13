package defaultmcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

// defaultMCPManager is a minimal MCP manager implementation registered by this plugin.
// It satisfies the core.MCPManager interface and enables initialization and basic inspection
// without pulling heavy dependencies into core. Tool discovery/execution can be expanded later.
type defaultMCPManager struct {
	config           core.MCPConfig
	connectedServers map[string]bool
	tools            []core.MCPToolInfo
	mu               sync.RWMutex
}

func newDefaultMCPManager(cfg core.MCPConfig) (core.MCPManager, error) {
	return &defaultMCPManager{
		config:           cfg,
		connectedServers: make(map[string]bool),
		tools:            []core.MCPToolInfo{},
	}, nil
}

// Connect marks a configured server as connected after basic validation.
func (m *defaultMCPManager) Connect(ctx context.Context, serverName string) error {
	// Validate server exists in config
	var found bool
	for _, s := range m.config.Servers {
		if s.Name == serverName && s.Enabled {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("server %s not found in configuration", serverName)
	}
	m.mu.Lock()
	m.connectedServers[serverName] = true
	m.mu.Unlock()
	return nil
}

func (m *defaultMCPManager) Disconnect(serverName string) error {
	m.mu.Lock()
	delete(m.connectedServers, serverName)
	m.mu.Unlock()
	return nil
}

func (m *defaultMCPManager) DisconnectAll() error {
	m.mu.Lock()
	m.connectedServers = make(map[string]bool)
	m.mu.Unlock()
	return nil
}

func (m *defaultMCPManager) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
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

func (m *defaultMCPManager) ListConnectedServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.connectedServers))
	for name := range m.connectedServers {
		out = append(out, name)
	}
	return out
}

func (m *defaultMCPManager) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
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

func (m *defaultMCPManager) RefreshTools(ctx context.Context) error {
	// No-op for now; discovery is transport-specific and will be added in dedicated plugins.
	return nil
}

func (m *defaultMCPManager) GetAvailableTools() []core.MCPToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]core.MCPToolInfo(nil), m.tools...)
}

func (m *defaultMCPManager) GetToolsFromServer(serverName string) []core.MCPToolInfo {
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

func (m *defaultMCPManager) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	health := make(map[string]core.MCPHealthStatus)
	for _, s := range m.config.Servers {
		if !s.Enabled {
			continue
		}
		status := core.MCPHealthStatus{Status: "unknown"}
		m.mu.RLock()
		if m.connectedServers[s.Name] {
			status.Status = "healthy"
		}
		m.mu.RUnlock()
		health[s.Name] = status
	}
	return health
}

func (m *defaultMCPManager) GetMetrics() core.MCPMetrics {
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

// Register the default MCP manager factory on import.
func init() {
	core.SetMCPManagerFactory(func(cfg core.MCPConfig) (core.MCPManager, error) {
		return newDefaultMCPManager(cfg)
	})
	// Provide default cache manager (in-memory) and function tool registry implementations
	core.SetMCPCacheManagerFactory(func(cfg core.MCPCacheConfig) (core.MCPCacheManager, error) {
		// minimal in-memory cache manager using core's real implementation for now
		// this preserves behavior until cache is extracted into dedicated plugins
		// create a tiny wrapper that defers to the internal helper through the public hook path
		type inMemory struct{ core.MCPCacheManager }
		// Fallback to core internal real implementation if present; otherwise a no-op mock
		// Try initialize via core helper; if it returns error, create a pass-through stub
		if mgr, err := func() (core.MCPCacheManager, error) {
			// Use a small shim: call InitializeMCPCacheManager with cfg in a temp path then capture the instance
			// We cannot call unexported helpers here; instead return a simple ephemeral cache implementation.
			return newSimpleCacheManager(cfg), nil
		}(); err == nil {
			return mgr, nil
		}
		return newSimpleCacheManager(cfg), nil
	})
	core.SetFunctionToolRegistryFactory(func() (core.FunctionToolRegistry, error) {
		return newSimpleFunctionToolRegistry(), nil
	})
}

// Simple, minimal in-memory cache manager to satisfy interfaces without heavy deps.
type simpleCache struct {
	data map[string]*core.MCPCachedResult
	mu   sync.RWMutex
}
type simpleCacheManager struct {
	cfg    core.MCPCacheConfig
	caches map[string]*simpleCache
	mu     sync.RWMutex
}

func newSimpleCacheManager(cfg core.MCPCacheConfig) core.MCPCacheManager {
	return &simpleCacheManager{cfg: cfg, caches: map[string]*simpleCache{}}
}
func (m *simpleCacheManager) GetCache(toolName, serverName string) core.MCPCache {
	key := toolName + ":" + serverName
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.caches[key]; ok {
		return c
	}
	c := &simpleCache{data: map[string]*core.MCPCachedResult{}}
	m.caches[key] = c
	return c
}
func (m *simpleCacheManager) ExecuteWithCache(ctx context.Context, e core.MCPToolExecution) (core.MCPToolResult, error) {
	args := map[string]string{}
	for k, v := range e.Arguments {
		args[k] = fmt.Sprintf("%v", v)
	}
	key := core.GenerateCacheKey(e.ToolName, e.ServerName, args)
	cache := m.GetCache(e.ToolName, e.ServerName)
	if m.cfg.Enabled {
		if r, err := cache.Get(ctx, key); err == nil {
			return r.Result, nil
		}
	}
	exec, ok := core.GetMCPManager().(core.MCPToolExecutor)
	if !ok {
		return core.MCPToolResult{}, fmt.Errorf("MCP manager does not support direct tool execution")
	}
	res, err := exec.ExecuteTool(ctx, e.ToolName, e.Arguments)
	if err != nil {
		return res, err
	}
	if m.cfg.Enabled && res.Success {
		_ = cache.Set(ctx, key, res, m.cfg.DefaultTTL)
	}
	return res, nil
}
func (m *simpleCacheManager) InvalidateByPattern(ctx context.Context, pattern string) error {
	return nil
}
func (m *simpleCacheManager) GetGlobalStats(ctx context.Context) (core.MCPCacheStats, error) {
	return core.MCPCacheStats{}, nil
}
func (m *simpleCacheManager) Shutdown() error {
	m.mu.Lock()
	m.caches = map[string]*simpleCache{}
	m.mu.Unlock()
	return nil
}
func (m *simpleCacheManager) Configure(cfg core.MCPCacheConfig) error { m.cfg = cfg; return nil }

func (c *simpleCache) Get(ctx context.Context, key core.MCPCacheKey) (*core.MCPCachedResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.data[c.key(key)]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("cache miss")
}
func (c *simpleCache) Set(ctx context.Context, key core.MCPCacheKey, result core.MCPToolResult, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[c.key(key)] = &core.MCPCachedResult{Key: key, Result: result, Timestamp: time.Now(), TTL: ttl}
	return nil
}
func (c *simpleCache) Delete(ctx context.Context, key core.MCPCacheKey) error {
	c.mu.Lock()
	delete(c.data, c.key(key))
	c.mu.Unlock()
	return nil
}
func (c *simpleCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	c.data = map[string]*core.MCPCachedResult{}
	c.mu.Unlock()
	return nil
}
func (c *simpleCache) Exists(ctx context.Context, key core.MCPCacheKey) (bool, error) {
	c.mu.RLock()
	_, ok := c.data[c.key(key)]
	c.mu.RUnlock()
	return ok, nil
}
func (c *simpleCache) Stats(ctx context.Context) (core.MCPCacheStats, error) {
	c.mu.RLock()
	n := len(c.data)
	c.mu.RUnlock()
	return core.MCPCacheStats{TotalKeys: n}, nil
}
func (c *simpleCache) Cleanup(ctx context.Context) error { return nil }
func (c *simpleCache) Close() error {
	c.mu.Lock()
	c.data = map[string]*core.MCPCachedResult{}
	c.mu.Unlock()
	return nil
}
func (c *simpleCache) key(k core.MCPCacheKey) string {
	return k.ToolName + ":" + k.ServerName + ":" + k.Hash
}

// Simple function tool registry
type simpleToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]core.FunctionTool
}

func newSimpleFunctionToolRegistry() core.FunctionToolRegistry {
	return &simpleToolRegistry{tools: map[string]core.FunctionTool{}}
}
func (r *simpleToolRegistry) Register(t core.FunctionTool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t == nil {
		return fmt.Errorf("tool cannot be nil")
	}
	n := t.Name()
	if n == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if _, ok := r.tools[n]; ok {
		return fmt.Errorf("tool %s already registered", n)
	}
	r.tools[n] = t
	return nil
}
func (r *simpleToolRegistry) Get(n string) (core.FunctionTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[n]
	return t, ok
}
func (r *simpleToolRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.tools))
	for n := range r.tools {
		out = append(out, n)
	}
	return out
}
func (r *simpleToolRegistry) CallTool(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	t, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return t.Call(ctx, args)
}

