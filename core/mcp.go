// Package core provides the complete public API for Model Context Protocol (MCP) integration.
//
// This package exposes a comprehensive, consolidated API for MCP integration in AgentFlow.
// All MCP functionality is available through this single interface, providing progressive
// complexity from simple tool usage to production-ready deployments.
//
// Usage Patterns:
//   - Basic MCP: QuickStartMCP(), NewMCPAgent()
//   - Enhanced MCP: InitializeMCPWithCache(), NewMCPAgentWithCache()
//   - Production MCP: InitializeProductionMCP(), NewProductionMCPAgent()
//
// NOTE: This package currently includes real implementations for manager/cache/registry.
// Upcoming change will move those into plugins (stdio/tcp/websocket, cache backends),
// keeping only interfaces + factory hooks here. A default plugin is available at
// plugins/mcp/default to preserve current behavior via SetMCPManagerFactory.
package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ==========================================
// SECTION 1: CORE INTERFACES (~200 lines)
// ==========================================

// MCPManager provides the main interface for managing MCP connections and tools.
// It handles server discovery, connection management, and tool registration.
type MCPManager interface {
	// Connection Management
	Connect(ctx context.Context, serverName string) error
	Disconnect(serverName string) error
	DisconnectAll() error

	// Server Discovery and Management
	DiscoverServers(ctx context.Context) ([]MCPServerInfo, error)
	ListConnectedServers() []string
	GetServerInfo(serverName string) (*MCPServerInfo, error)

	// Tool Management
	RefreshTools(ctx context.Context) error
	GetAvailableTools() []MCPToolInfo
	GetToolsFromServer(serverName string) []MCPToolInfo

	// Health and Monitoring
	HealthCheck(ctx context.Context) map[string]MCPHealthStatus
	GetMetrics() MCPMetrics
}

// MCPAgent represents an agent that can utilize MCP tools.
// It extends the basic Agent interface with MCP-specific capabilities.
type MCPAgent interface {
	Agent
	// MCP-specific methods
	SelectTools(ctx context.Context, query string, stateContext State) ([]string, error)
	ExecuteTools(ctx context.Context, tools []MCPToolExecution) ([]MCPToolResult, error)
	GetAvailableMCPTools() []MCPToolInfo
}

// MCPCache defines the interface for caching MCP tool results.
type MCPCache interface {
	// Get retrieves a cached result by key
	Get(ctx context.Context, key MCPCacheKey) (*MCPCachedResult, error)

	// Set stores a result in the cache
	Set(ctx context.Context, key MCPCacheKey, result MCPToolResult, ttl time.Duration) error

	// Delete removes a specific key from the cache
	Delete(ctx context.Context, key MCPCacheKey) error

	// Clear removes all entries from the cache
	Clear(ctx context.Context) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key MCPCacheKey) (bool, error)

	// Stats returns cache performance statistics
	Stats(ctx context.Context) (MCPCacheStats, error)

	// Cleanup performs maintenance operations (e.g., TTL expiration)
	Cleanup(ctx context.Context) error

	// Close closes the cache and releases resources
	Close() error
}

// MCPCacheManager manages multiple cache instances and provides cache-aware tool execution.
type MCPCacheManager interface {
	// GetCache returns a cache instance for a specific tool or server
	GetCache(toolName, serverName string) MCPCache

	// ExecuteWithCache executes a tool with caching support
	ExecuteWithCache(ctx context.Context, execution MCPToolExecution) (MCPToolResult, error)

	// InvalidateByPattern invalidates cache entries matching a pattern
	InvalidateByPattern(ctx context.Context, pattern string) error

	// GetGlobalStats returns aggregated cache statistics
	GetGlobalStats(ctx context.Context) (MCPCacheStats, error)

	// Shutdown cleanly shuts down the cache manager
	Shutdown() error

	// Configure updates cache configuration
	Configure(config MCPCacheConfig) error
}

// FunctionTool defines the interface for a callable tool that agents can use.
type FunctionTool interface {
	// Name returns the unique identifier for the tool.
	Name() string
	// Call executes the tool's logic with the given arguments.
	Call(ctx context.Context, args map[string]any) (map[string]any, error)
}

// FunctionToolRegistry defines the interface for managing function tools.
type FunctionToolRegistry interface {
	// Register adds a tool to the registry.
	Register(tool FunctionTool) error
	// Get retrieves a tool by name.
	Get(name string) (FunctionTool, bool)
	// List returns all registered tool names.
	List() []string
	// CallTool executes a tool by name with the given arguments.
	CallTool(ctx context.Context, name string, args map[string]any) (map[string]any, error)
}

// MCPToolExecutor is an optional interface that MCPManager implementations
// can satisfy to enable direct tool execution via ExecuteMCPTool/Cache.
// Transport plugins should implement this when they support direct calls.
type MCPToolExecutor interface {
	ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (MCPToolResult, error)
}

// ==========================================
// SECTION 2: CONFIGURATION TYPES (~150 lines)
// ==========================================

// MCPConfig holds configuration for MCP integration.
type MCPConfig struct {
	// Discovery settings
	EnableDiscovery  bool          `toml:"enable_discovery"`
	DiscoveryTimeout time.Duration `toml:"discovery_timeout"`
	ScanPorts        []int         `toml:"scan_ports"`

	// Connection settings
	ConnectionTimeout time.Duration `toml:"connection_timeout"`
	MaxRetries        int           `toml:"max_retries"`
	RetryDelay        time.Duration `toml:"retry_delay"`

	// Server configurations
	Servers []MCPServerConfig `toml:"servers"`

	// Performance settings
	EnableCaching  bool          `toml:"enable_caching"`
	CacheTimeout   time.Duration `toml:"cache_timeout"`
	MaxConnections int           `toml:"max_connections"`
}

// MCPCacheConfig holds configuration for the cache system.
type MCPCacheConfig struct {
	// Cache behavior
	Enabled    bool          `toml:"enabled"`
	DefaultTTL time.Duration `toml:"default_ttl"`
	MaxSize    int64         `toml:"max_size_mb"`
	MaxKeys    int           `toml:"max_keys"`

	// Eviction policy
	EvictionPolicy  string        `toml:"eviction_policy"` // "lru", "lfu", "ttl"
	CleanupInterval time.Duration `toml:"cleanup_interval"`

	// Per-tool TTL overrides
	ToolTTLs map[string]time.Duration `toml:"tool_ttls"`

	// Backend configuration
	Backend       string            `toml:"backend"` // "memory", "redis", "file"
	BackendConfig map[string]string `toml:"backend_config"`
}

// ProductionConfig contains all production-level configuration.
type ProductionConfig struct {
	// Connection pooling configuration
	ConnectionPool ConnectionPoolConfig `toml:"connection_pool"`

	// Retry policy configuration
	RetryPolicy RetryPolicyConfig `toml:"retry_policy"`

	// Load balancing configuration
	LoadBalancer LoadBalancerConfig `toml:"load_balancer"`

	// Metrics configuration
	Metrics MetricsConfig `toml:"metrics"`

	// Health check configuration
	HealthCheck HealthCheckConfig `toml:"health_check"`

	// Cache configuration
	Cache CacheConfig `toml:"cache"`

	// Circuit breaker configuration
	CircuitBreaker ProductionCircuitBreakerConfig `toml:"circuit_breaker"`
}

// MCPServerConfig defines configuration for individual MCP servers.
type MCPServerConfig struct {
	Name    string `toml:"name"`
	Type    string `toml:"type"` // tcp, stdio, docker, websocket
	Host    string `toml:"host,omitempty"`
	Port    int    `toml:"port,omitempty"`
	Command string `toml:"command,omitempty"` // for stdio transport
	Enabled bool   `toml:"enabled"`
}

// ConnectionPoolConfig contains connection pooling settings.
type ConnectionPoolConfig struct {
	MinConnections       int           `toml:"min_connections"`
	MaxConnections       int           `toml:"max_connections"`
	MaxIdleTime          time.Duration `toml:"max_idle_time"`
	HealthCheckInterval  time.Duration `toml:"health_check_interval"`
	HealthCheckTimeout   time.Duration `toml:"health_check_timeout"`
	ReconnectBackoff     time.Duration `toml:"reconnect_backoff"`
	MaxReconnectBackoff  time.Duration `toml:"max_reconnect_backoff"`
	MaxReconnectAttempts int           `toml:"max_reconnect_attempts"`
	ConnectionTimeout    time.Duration `toml:"connection_timeout"`
	MaxConnectionAge     time.Duration `toml:"max_connection_age"`
}

// RetryPolicyConfig contains retry policy settings.
type RetryPolicyConfig struct {
	Strategy             string                     `toml:"strategy"` // exponential, linear, adaptive
	BaseDelay            time.Duration              `toml:"base_delay"`
	MaxDelay             time.Duration              `toml:"max_delay"`
	MaxAttempts          int                        `toml:"max_attempts"`
	Multiplier           float64                    `toml:"multiplier"`
	Jitter               float64                    `toml:"jitter"`
	RetryableErrors      []string                   `toml:"retryable_errors"`
	NonRetryableErrors   []string                   `toml:"non_retryable_errors"`
	ToolSpecificPolicies map[string]ToolRetryConfig `toml:"tool_specific_policies"`
}

// ToolRetryConfig contains tool-specific retry configuration.
type ToolRetryConfig struct {
	Strategy    string        `toml:"strategy"`
	BaseDelay   time.Duration `toml:"base_delay"`
	MaxDelay    time.Duration `toml:"max_delay"`
	MaxAttempts int           `toml:"max_attempts"`
}

// LoadBalancerConfig contains load balancer settings.
type LoadBalancerConfig struct {
	Strategy              string        `toml:"strategy"` // round_robin, least_connections, etc.
	HealthCheckInterval   time.Duration `toml:"health_check_interval"`
	HealthCheckTimeout    time.Duration `toml:"health_check_timeout"`
	UnhealthyThreshold    int           `toml:"unhealthy_threshold"`
	HealthyThreshold      int           `toml:"healthy_threshold"`
	FailoverEnabled       bool          `toml:"failover_enabled"`
	CircuitBreakerEnabled bool          `toml:"circuit_breaker_enabled"`
}

// MetricsConfig contains metrics settings.
type MetricsConfig struct {
	Enabled           bool          `toml:"enabled"`
	Port              int           `toml:"port"`
	Path              string        `toml:"path"`
	UpdateInterval    time.Duration `toml:"update_interval"`
	HistogramBuckets  []float64     `toml:"histogram_buckets"`
	PrometheusEnabled bool          `toml:"prometheus_enabled"`
}

// HealthCheckConfig contains health check settings.
type HealthCheckConfig struct {
	Enabled        bool          `toml:"enabled"`
	Port           int           `toml:"port"`
	Path           string        `toml:"path"`
	Interval       time.Duration `toml:"interval"`
	Timeout        time.Duration `toml:"timeout"`
	ChecksRequired int           `toml:"checks_required"`
}

// CacheConfig contains cache settings (extending existing).
type CacheConfig struct {
	// Existing cache config
	Type    string        `toml:"type"`
	TTL     time.Duration `toml:"ttl"`
	MaxSize int           `toml:"max_size"`

	// Production-specific settings
	BackgroundCleanup  bool          `toml:"background_cleanup"`
	CleanupInterval    time.Duration `toml:"cleanup_interval"`
	MemoryLimit        int64         `toml:"memory_limit"`
	CompressionEnabled bool          `toml:"compression_enabled"`
	PersistenceEnabled bool          `toml:"persistence_enabled"`
	PersistencePath    string        `toml:"persistence_path"`

	// Distributed cache settings
	Redis RedisConfig `toml:"redis"`
}

// RedisConfig contains Redis cache settings.
type RedisConfig struct {
	Enabled    bool          `toml:"enabled"`
	Address    string        `toml:"address"`
	Password   string        `toml:"password"`
	Database   int           `toml:"database"`
	PoolSize   int           `toml:"pool_size"`
	Timeout    time.Duration `toml:"timeout"`
	MaxRetries int           `toml:"max_retries"`
}

// ProductionCircuitBreakerConfig contains circuit breaker settings.
type ProductionCircuitBreakerConfig struct {
	// Existing circuit breaker config
	FailureThreshold int           `toml:"failure_threshold"`
	SuccessThreshold int           `toml:"success_threshold"`
	Timeout          time.Duration `toml:"timeout"`

	// Production-specific settings
	HalfOpenMaxCalls    int           `toml:"half_open_max_calls"`
	OpenStateTimeout    time.Duration `toml:"open_state_timeout"`
	MetricsEnabled      bool          `toml:"metrics_enabled"`
	NotificationEnabled bool          `toml:"notification_enabled"`
}

// ==========================================
// SECTION 3: DATA TYPES (~150 lines)
// ==========================================

// MCPServerInfo represents information about an MCP server.
type MCPServerInfo struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Address      string                 `json:"address"`
	Port         int                    `json:"port"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Capabilities map[string]interface{} `json:"capabilities"`
	Status       string                 `json:"status"` // connected, disconnected, error
}

// MCPToolInfo represents metadata about an available MCP tool.
type MCPToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	ServerName  string                 `json:"server_name"`
}

// MCPToolExecution represents a tool execution request.
type MCPToolExecution struct {
	ToolName   string                 `json:"tool_name"`
	Arguments  map[string]interface{} `json:"arguments"`
	ServerName string                 `json:"server_name,omitempty"`
}

// MCPToolResult represents the result of an MCP tool execution.
type MCPToolResult struct {
	ToolName   string        `json:"tool_name"`
	ServerName string        `json:"server_name"`
	Success    bool          `json:"success"`
	Content    []MCPContent  `json:"content,omitempty"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
}

// MCPContent represents content returned by MCP tools.
type MCPContent struct {
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	Data     string                 `json:"data,omitempty"`
	MimeType string                 `json:"mime_type,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MCPHealthStatus represents the health status of an MCP server connection.
type MCPHealthStatus struct {
	Status       string        `json:"status"` // healthy, unhealthy, unknown
	LastCheck    time.Time     `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	ToolCount    int           `json:"tool_count"`
}

// MCPMetrics provides metrics about MCP operations.
type MCPMetrics struct {
	ConnectedServers int                         `json:"connected_servers"`
	TotalTools       int                         `json:"total_tools"`
	ToolExecutions   int64                       `json:"tool_executions"`
	AverageLatency   time.Duration               `json:"average_latency"`
	ErrorRate        float64                     `json:"error_rate"`
	ServerMetrics    map[string]MCPServerMetrics `json:"server_metrics"`
}

// MCPServerMetrics provides metrics for individual servers.
type MCPServerMetrics struct {
	ToolCount        int           `json:"tool_count"`
	Executions       int64         `json:"executions"`
	SuccessfulCalls  int64         `json:"successful_calls"`
	FailedCalls      int64         `json:"failed_calls"`
	AverageLatency   time.Duration `json:"average_latency"`
	LastActivity     time.Time     `json:"last_activity"`
	ConnectionUptime time.Duration `json:"connection_uptime"`
}

// MCPCacheKey represents a unique identifier for cached tool results.
type MCPCacheKey struct {
	ToolName   string            `json:"tool_name"`
	ServerName string            `json:"server_name"`
	Args       map[string]string `json:"args"`
	Hash       string            `json:"hash"` // SHA256 hash of normalized args
}

// MCPCachedResult represents a cached tool execution result.
type MCPCachedResult struct {
	Key         MCPCacheKey            `json:"key"`
	Result      MCPToolResult          `json:"result"`
	Timestamp   time.Time              `json:"timestamp"`
	TTL         time.Duration          `json:"ttl"`
	AccessCount int                    `json:"access_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MCPCacheStats provides statistics about cache performance.
type MCPCacheStats struct {
	TotalKeys      int           `json:"total_keys"`
	HitCount       int64         `json:"hit_count"`
	MissCount      int64         `json:"miss_count"`
	HitRate        float64       `json:"hit_rate"`
	EvictionCount  int64         `json:"eviction_count"`
	TotalSize      int64         `json:"total_size_bytes"`
	AverageLatency time.Duration `json:"average_latency"`
	LastCleanup    time.Time     `json:"last_cleanup"`
}

// ==========================================
// SECTION 4: GLOBAL STATE (~50 lines)
// ==========================================

// Global MCP manager instance and registry
var (
	globalMCPManager        MCPManager
	globalMCPRegistry       FunctionToolRegistry
	globalCacheManager      MCPCacheManager
	mcpManagerMutex         sync.RWMutex
	mcpRegistryMutex        sync.RWMutex
	cacheManagerMutex       sync.RWMutex
	mcpManagerInitialized   bool
	cacheManagerInitialized bool
)

// ==========================================
// SECTION 5: INITIALIZATION FUNCTIONS (~150 lines)
// ==========================================

// QuickStartMCP initializes MCP with minimal configuration for simple use cases.
// This is the fastest way to get started with MCP integration.
func QuickStartMCP(tools ...string) error {
	config := DefaultMCPConfig()

	// Enable discovery to find servers automatically
	config.EnableDiscovery = true

	return InitializeMCP(config)
}

// InitializeMCP initializes the basic MCP manager with the provided configuration.
// This provides core MCP functionality without advanced features.
func InitializeMCP(config MCPConfig) error {
	mcpManagerMutex.Lock()
	defer mcpManagerMutex.Unlock()

	if mcpManagerInitialized {
		Logger().Debug().Msg("MCP manager already initialized")
		return nil
	}

	// Create MCP manager through internal factory
	manager, err := createMCPManagerInternal(config)
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	globalMCPManager = manager
	mcpManagerInitialized = true

	Logger().Debug().Msg("MCP manager initialized successfully")
	return nil
}

// InitializeMCPWithCache initializes MCP with caching capabilities.
// This provides enhanced performance through intelligent result caching.
func InitializeMCPWithCache(mcpConfig MCPConfig, cacheConfig MCPCacheConfig) error {
	// First initialize basic MCP
	if err := InitializeMCP(mcpConfig); err != nil {
		return fmt.Errorf("failed to initialize MCP: %w", err)
	}

	// Then initialize cache manager
	if err := InitializeMCPCacheManager(cacheConfig); err != nil {
		return fmt.Errorf("failed to initialize MCP cache: %w", err)
	}

	Logger().Debug().Msg("MCP with cache initialized successfully")
	return nil
}

// InitializeProductionMCP initializes the complete production MCP stack.
// This provides all advanced features: connection pooling, retry logic, metrics, etc.
func InitializeProductionMCP(ctx context.Context, config ProductionConfig) error {
	// Initialize basic MCP with production settings
	mcpConfig := ProductionMCPConfig(config)
	if err := InitializeMCP(mcpConfig); err != nil {
		return fmt.Errorf("failed to initialize production MCP: %w", err)
	}

	// Initialize cache if enabled
	if config.Cache.Type != "" {
		cacheConfig := ProductionCacheConfig(config.Cache)
		if err := InitializeMCPCacheManager(cacheConfig); err != nil {
			return fmt.Errorf("failed to initialize production cache: %w", err)
		}
	}

	// Initialize metrics if enabled
	if config.Metrics.Enabled {
		if err := initializeProductionMetrics(config.Metrics); err != nil {
			return fmt.Errorf("failed to initialize production metrics: %w", err)
		}
	}

	Logger().Debug().Msg("Production MCP initialized successfully")
	return nil
}

// InitializeMCPCacheManager initializes the global MCP cache manager.
func InitializeMCPCacheManager(config MCPCacheConfig) error {
	cacheManagerMutex.Lock()
	defer cacheManagerMutex.Unlock()

	if cacheManagerInitialized {
		Logger().Debug().Msg("MCP cache manager already initialized")
		return nil
	}

	// Create cache manager through internal factory
	manager, err := createMCPCacheManagerInternal(config)
	if err != nil {
		return fmt.Errorf("failed to create MCP cache manager: %w", err)
	}

	globalCacheManager = manager
	cacheManagerInitialized = true

	Logger().Debug().Msg("MCP cache manager initialized successfully")
	return nil
}

// InitializeMCPToolRegistry initializes the global MCP tool registry.
func InitializeMCPToolRegistry() error {
	mcpRegistryMutex.Lock()
	defer mcpRegistryMutex.Unlock()

	if globalMCPRegistry != nil {
		Logger().Debug().Msg("MCP tool registry already initialized")
		return nil
	}

	// Create registry through internal factory
	registry, err := createMCPToolRegistryInternal()
	if err != nil {
		return fmt.Errorf("failed to create MCP tool registry: %w", err)
	}

	globalMCPRegistry = registry
	Logger().Debug().Msg("MCP tool registry initialized successfully")
	return nil
}

// InitializeMCPManager initializes the global MCP manager with the provided configuration.
// This is an alias for InitializeMCP for backward compatibility.
func InitializeMCPManager(config MCPConfig) error {
	return InitializeMCP(config)
}

// CreateMCPAgentWithLLMAndTools creates an MCP-aware agent with the specified configuration.
// This is a comprehensive factory function for creating fully configured agents.
func CreateMCPAgentWithLLMAndTools(ctx context.Context, name string, llmProvider ModelProvider, mcpConfig MCPConfig, agentConfig MCPAgentConfig) (MCPAwareAgent, error) {
	// Initialize MCP if not already done
	if err := InitializeMCP(mcpConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP: %w", err)
	}

	// Initialize tool registry if not already done
	if err := InitializeMCPToolRegistry(); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP tool registry: %w", err)
	}

	// Get the initialized manager
	manager := GetMCPManager()
	if manager == nil {
		return nil, fmt.Errorf("MCP manager not available after initialization")
	}

	// Create the agent
	return NewMCPAwareAgent(name, llmProvider, manager, agentConfig), nil
}

// ShutdownMCPManager gracefully shuts down the global MCP manager.
// This is an alias for ShutdownMCP for backward compatibility.
func ShutdownMCPManager() error {
	return ShutdownMCP()
}

// ==========================================
// SECTION 6: AGENT FACTORIES (~100 lines)
// ==========================================

// NewMCPAgent creates a basic MCP-aware agent with essential capabilities.
// This is the simplest way to create an agent that can use MCP tools.
func NewMCPAgent(name string, llmProvider ModelProvider) (MCPAwareAgent, error) {
	manager := GetMCPManager()
	if manager == nil {
		return nil, fmt.Errorf("MCP manager not initialized - call InitializeMCP() first")
	}

	config := DefaultMCPAgentConfig()
	return NewMCPAwareAgent(name, llmProvider, manager, config), nil
}

// NewMCPAgentWithCache creates an MCP-aware agent with caching capabilities.
// This provides better performance through intelligent result caching.
func NewMCPAgentWithCache(name string, llmProvider ModelProvider) (MCPAwareAgent, error) {
	manager := GetMCPManager()
	if manager == nil {
		return nil, fmt.Errorf("MCP manager not initialized - call InitializeMCPWithCache() first")
	}

	cacheManager := GetMCPCacheManager()
	if cacheManager == nil {
		return nil, fmt.Errorf("MCP cache manager not initialized - call InitializeMCPWithCache() first")
	}

	config := DefaultMCPAgentConfig()
	config.EnableCaching = true

	agent := NewMCPAwareAgent(name, llmProvider, manager, config)
	// Wire cache manager into agent (would need to extend MCPAwareAgent)

	return agent, nil
}

// NewProductionMCPAgent creates a production-ready MCP agent with all advanced features.
// This provides enterprise-grade capabilities: connection pooling, retry logic, metrics, etc.
func NewProductionMCPAgent(name string, llmProvider ModelProvider, config ProductionConfig) (MCPAwareAgent, error) {
	manager := GetMCPManager()
	if manager == nil {
		return nil, fmt.Errorf("production MCP not initialized - call InitializeProductionMCP() first")
	}

	agentConfig := ProductionAgentConfig(config)
	return NewMCPAwareAgent(name, llmProvider, manager, agentConfig), nil
}

// ==========================================
// SECTION 7: GLOBAL ACCESS FUNCTIONS (~50 lines)
// ==========================================

// GetMCPManager returns the global MCP manager instance.
func GetMCPManager() MCPManager {
	mcpManagerMutex.RLock()
	defer mcpManagerMutex.RUnlock()
	return globalMCPManager
}

// GetMCPCacheManager returns the global cache manager instance.
func GetMCPCacheManager() MCPCacheManager {
	cacheManagerMutex.RLock()
	defer cacheManagerMutex.RUnlock()
	return globalCacheManager
}

// GetMCPToolRegistry returns the global MCP tool registry.
func GetMCPToolRegistry() FunctionToolRegistry {
	mcpRegistryMutex.RLock()
	defer mcpRegistryMutex.RUnlock()
	return globalMCPRegistry
}

// ==========================================
// SECTION 8: SIMPLE HELPER FUNCTIONS (~80 lines)
// ==========================================

// ConnectMCPServer connects to a single MCP server with simple configuration.
// This is useful for quickly connecting to a known server.
func ConnectMCPServer(name, serverType, endpoint string) error {
	manager := GetMCPManager()
	if manager == nil {
		return fmt.Errorf("MCP manager not initialized")
	}

	return manager.Connect(context.Background(), name)
}

// ExecuteMCPTool executes a single MCP tool with a simple interface.
// This is the simplest way to execute an MCP tool without creating an agent.
func ExecuteMCPTool(ctx context.Context, toolName string, args map[string]interface{}) (MCPToolResult, error) {
	manager := GetMCPManager()
	if manager == nil {
		return MCPToolResult{}, fmt.Errorf("MCP manager not initialized")
	}

	// Check if cache manager is available
	cacheManager := GetMCPCacheManager()
	if cacheManager != nil {
		// Use cache-aware execution
		execution := MCPToolExecution{
			ToolName:  toolName,
			Arguments: args,
		}
		return cacheManager.ExecuteWithCache(ctx, execution)
	}
	// Direct execution without cache using a manager that supports MCPToolExecutor
	if exec, ok := manager.(MCPToolExecutor); ok {
		return exec.ExecuteTool(ctx, toolName, args)
	}
	return MCPToolResult{}, fmt.Errorf("MCP manager does not support direct tool execution. Import a transport plugin (e.g., plugins/mcp/tcp, plugins/mcp/stdio, or plugins/mcp/websocket)")
}

// RegisterMCPToolsWithRegistry discovers and registers all available MCP tools with the registry.
func RegisterMCPToolsWithRegistry(ctx context.Context) error {
	manager := GetMCPManager()
	if manager == nil {
		return fmt.Errorf("MCP manager not initialized")
	}

	registry := GetMCPToolRegistry()
	if registry == nil {
		return fmt.Errorf("MCP tool registry not initialized")
	}

	// Refresh tools from all connected servers
	if err := manager.RefreshTools(ctx); err != nil {
		Logger().Warn().Err(err).Msg("Failed to refresh tools from some MCP servers")
	}
	// Get available tools and register them
	tools := manager.GetAvailableTools()
	Logger().Debug().Int("tool_count", len(tools)).Msg("Registering MCP tools with registry")

	// Register each MCP tool as a FunctionTool
	for _, toolInfo := range tools {
		mcpTool := newMCPFunctionTool(toolInfo, manager)
		if err := registry.Register(mcpTool); err != nil {
			Logger().Warn().
				Str("tool", toolInfo.Name).
				Str("server", toolInfo.ServerName).
				Err(err).
				Msg("Failed to register MCP tool")
		} else {
			Logger().Debug().
				Str("tool", toolInfo.Name).
				Str("server", toolInfo.ServerName).
				Msg("Successfully registered MCP tool")
		}
	}

	Logger().Debug().Int("registered_tools", len(tools)).Msg("Completed MCP tool registration")
	return nil
}

// ==========================================
// SECTION 9: CONFIGURATION HELPERS (~100 lines)
// ==========================================

// DefaultMCPConfig returns a default MCP configuration suitable for development.
func DefaultMCPConfig() MCPConfig {
	return MCPConfig{
		EnableDiscovery:   true,
		DiscoveryTimeout:  10 * time.Second,
		ScanPorts:         []int{8080, 8081, 8090, 8100, 3000, 3001},
		ConnectionTimeout: 30 * time.Second,
		MaxRetries:        3,
		RetryDelay:        1 * time.Second,
		EnableCaching:     true,
		CacheTimeout:      5 * time.Minute,
		MaxConnections:    10,
		Servers:           []MCPServerConfig{}, // Empty by default
	}
}

// DefaultMCPCacheConfig returns a default cache configuration.
func DefaultMCPCacheConfig() MCPCacheConfig {
	return MCPCacheConfig{
		Enabled:         true,
		DefaultTTL:      15 * time.Minute,
		MaxSize:         100, // 100 MB
		MaxKeys:         10000,
		EvictionPolicy:  "lru",
		CleanupInterval: 5 * time.Minute,
		Backend:         "memory",
		ToolTTLs: map[string]time.Duration{
			"web_search":         5 * time.Minute,  // Search results change frequently
			"content_fetch":      30 * time.Minute, // Content is more stable
			"summarize_text":     60 * time.Minute, // Summaries are expensive to compute
			"sentiment_analysis": 45 * time.Minute, // Analysis results are stable
			"compute_metric":     20 * time.Minute, // Metrics may change
			"entity_extraction":  60 * time.Minute, // Entity extraction is expensive
		},
		BackendConfig: map[string]string{
			"redis_addr":     "localhost:6379",
			"redis_password": "",
			"redis_db":       "0",
			"file_path":      "./cache",
		},
	}
}

// DefaultProductionConfig returns production-ready default configuration.
func DefaultProductionConfig() ProductionConfig {
	return ProductionConfig{
		ConnectionPool: ConnectionPoolConfig{
			MinConnections:       5,
			MaxConnections:       50,
			MaxIdleTime:          10 * time.Minute,
			HealthCheckInterval:  30 * time.Second,
			HealthCheckTimeout:   5 * time.Second,
			ReconnectBackoff:     1 * time.Second,
			MaxReconnectBackoff:  30 * time.Second,
			MaxReconnectAttempts: 10,
			ConnectionTimeout:    30 * time.Second,
			MaxConnectionAge:     1 * time.Hour,
		},
		RetryPolicy: RetryPolicyConfig{
			Strategy:    "exponential",
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    30 * time.Second,
			MaxAttempts: 5,
			Multiplier:  2.0,
			Jitter:      0.1,
		},
		LoadBalancer: LoadBalancerConfig{
			Strategy:              "round_robin",
			HealthCheckInterval:   10 * time.Second,
			HealthCheckTimeout:    5 * time.Second,
			UnhealthyThreshold:    3,
			HealthyThreshold:      2,
			FailoverEnabled:       true,
			CircuitBreakerEnabled: true,
		},
		Metrics: MetricsConfig{
			Enabled:           true,
			Port:              8080,
			Path:              "/metrics",
			UpdateInterval:    10 * time.Second,
			PrometheusEnabled: true,
		},
		HealthCheck: HealthCheckConfig{
			Enabled:        true,
			Port:           8081,
			Path:           "/health",
			Interval:       30 * time.Second,
			Timeout:        5 * time.Second,
			ChecksRequired: 3,
		},
		Cache: CacheConfig{
			Type:               "redis",
			TTL:                15 * time.Minute,
			MaxSize:            1000,
			BackgroundCleanup:  true,
			CleanupInterval:    5 * time.Minute,
			MemoryLimit:        1024 * 1024 * 1024, // 1GB
			CompressionEnabled: true,
			PersistenceEnabled: true,
			Redis: RedisConfig{
				Enabled:    true,
				Address:    "localhost:6379",
				PoolSize:   20,
				Timeout:    5 * time.Second,
				MaxRetries: 3,
			},
		},
		CircuitBreaker: ProductionCircuitBreakerConfig{
			FailureThreshold:    10,
			SuccessThreshold:    5,
			Timeout:             60 * time.Second,
			HalfOpenMaxCalls:    5,
			OpenStateTimeout:    30 * time.Second,
			MetricsEnabled:      true,
			NotificationEnabled: true,
		},
	}
}

// ProductionMCPConfig maps ProductionConfig to MCPConfig for initialization.
// This stays minimal in core; transport/cache plugins can extend behavior.
func ProductionMCPConfig(cfg ProductionConfig) MCPConfig {
	return MCPConfig{
		EnableDiscovery:   false,
		DiscoveryTimeout:  0,
		ScanPorts:         nil,
		ConnectionTimeout: cfg.ConnectionPool.ConnectionTimeout,
		MaxRetries:        cfg.RetryPolicy.MaxAttempts,
		RetryDelay:        cfg.RetryPolicy.BaseDelay,
		Servers:           []MCPServerConfig{},
		EnableCaching:     cfg.Cache.Type != "",
		CacheTimeout:      cfg.Cache.TTL,
		MaxConnections:    cfg.ConnectionPool.MaxConnections,
	}
}

// ProductionCacheConfig maps CacheConfig to MCPCacheConfig minimally.
func ProductionCacheConfig(cfg CacheConfig) MCPCacheConfig {
	backend := "memory"
	if strings.ToLower(cfg.Type) == "redis" {
		backend = "redis"
	}
	return MCPCacheConfig{
		Enabled:         cfg.Type != "",
		DefaultTTL:      cfg.TTL,
		MaxSize:         int64(cfg.MaxSize),
		EvictionPolicy:  "lru",
		CleanupInterval: cfg.CleanupInterval,
		Backend:         backend,
		BackendConfig:   map[string]string{},
	}
}

// ProductionAgentConfig maps ProductionConfig to MCPAgentConfig.
func ProductionAgentConfig(cfg ProductionConfig) MCPAgentConfig {
	return MCPAgentConfig{
		MaxToolsPerExecution: 5,
		ToolSelectionTimeout: 30 * time.Second,
		ParallelExecution:    true,
		ExecutionTimeout:     60 * time.Second,
		RetryFailedTools:     true,
		MaxRetries:           cfg.RetryPolicy.MaxAttempts,
		UseToolDescriptions:  true,
		ResultInterpretation: true,
		EnableCaching:        cfg.Cache.Type != "",
		CacheConfig:          ProductionCacheConfig(cfg.Cache),
	}
}

// NewMCPServerConfig creates a new server configuration with validation.
func NewMCPServerConfig(name, serverType, host string, port int) (MCPServerConfig, error) {
	config := MCPServerConfig{
		Name:    name,
		Type:    serverType,
		Host:    host,
		Port:    port,
		Enabled: true,
	}

	// Basic validation
	if name == "" {
		return config, fmt.Errorf("server name cannot be empty")
	}

	switch serverType {
	case "tcp", "websocket":
		if host == "" {
			return config, fmt.Errorf("%s server must specify host", serverType)
		}
		if port <= 0 || port > 65535 {
			return config, fmt.Errorf("%s server must specify valid port (1-65535)", serverType)
		}
	case "stdio":
		// For STDIO, we use the host field as the command
		if host == "" {
			return config, fmt.Errorf("stdio server must specify command")
		}
		config.Command = host
		config.Host = ""
		config.Port = 0
	case "docker":
		// Docker configuration validation could be added here
	default:
		return config, fmt.Errorf("unsupported server type: %s", serverType)
	}

	return config, nil
}

// NewTCPServerConfig creates a TCP server configuration.
func NewTCPServerConfig(name, host string, port int) (MCPServerConfig, error) {
	return NewMCPServerConfig(name, "tcp", host, port)
}

// NewSTDIOServerConfig creates a STDIO server configuration.
func NewSTDIOServerConfig(name, command string) (MCPServerConfig, error) {
	config := MCPServerConfig{
		Name:    name,
		Type:    "stdio",
		Command: command,
		Enabled: true,
	}

	if name == "" {
		return config, fmt.Errorf("server name cannot be empty")
	}

	if command == "" {
		return config, fmt.Errorf("stdio server must specify command")
	}

	return config, nil
}

// NewWebSocketServerConfig creates a WebSocket server configuration.
func NewWebSocketServerConfig(name, host string, port int) (MCPServerConfig, error) {
	return NewMCPServerConfig(name, "websocket", host, port)
}

// LoadMCPConfigFromTOML loads MCP configuration from a TOML file.
func LoadMCPConfigFromTOML(path string) (MCPConfig, error) {
	// TODO: Implement TOML file loading with proper parsing
	// For now, return default config with a warning
	Logger().Warn().
		Str("path", path).
		Msg("TOML configuration loading not implemented, using default config")

	config := DefaultMCPConfig()

	// Add a basic server configuration for demo purposes
	if len(config.Servers) == 0 {
		config.Servers = []MCPServerConfig{
			{
				Name:    "docker-mcp",
				Type:    "tcp",
				Host:    "localhost",
				Port:    8811,
				Enabled: true,
			},
		}
	}

	return config, nil
}

// ==========================================
// SECTION 10: CACHE UTILITIES (~50 lines)
// ==========================================

// GenerateCacheKey creates a standardized cache key for tool execution.
func GenerateCacheKey(toolName, serverName string, args map[string]string) MCPCacheKey {
	return MCPCacheKey{
		ToolName:   toolName,
		ServerName: serverName,
		Args:       normalizeArgs(args),
		Hash:       generateArgHash(args),
	}
}

// normalizeArgs ensures consistent argument formatting for cache keys.
func normalizeArgs(args map[string]string) map[string]string {
	normalized := make(map[string]string)
	for k, v := range args {
		// Normalize whitespace and case for cache consistency
		normalized[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	return normalized
}

// generateArgHash creates a deterministic hash of the arguments.
func generateArgHash(args map[string]string) string {
	// Sort keys for deterministic hashing
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k + "=" + args[k] + "|"))
	}
	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars for brevity
}

// ==========================================
// SECTION 11: SHUTDOWN FUNCTIONS (~50 lines)
// ==========================================

// ShutdownMCP cleanly shuts down all MCP components.
func ShutdownMCP() error {
	var errors []error

	// Shutdown cache manager
	if globalCacheManager != nil {
		if err := globalCacheManager.Shutdown(); err != nil {
			errors = append(errors, fmt.Errorf("cache manager shutdown error: %w", err))
		}
	}

	// Shutdown MCP manager
	if globalMCPManager != nil {
		if err := globalMCPManager.DisconnectAll(); err != nil {
			errors = append(errors, fmt.Errorf("MCP manager shutdown error: %w", err))
		}
	}

	// Reset global state
	mcpManagerMutex.Lock()
	globalMCPManager = nil
	mcpManagerInitialized = false
	mcpManagerMutex.Unlock()

	cacheManagerMutex.Lock()
	globalCacheManager = nil
	cacheManagerInitialized = false
	cacheManagerMutex.Unlock()

	mcpRegistryMutex.Lock()
	globalMCPRegistry = nil
	mcpRegistryMutex.Unlock()

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	Logger().Debug().Msg("MCP shutdown completed successfully")
	return nil
}

// ==========================================
// SECTION 12: INTERNAL BRIDGE FUNCTIONS (~100 lines)
// ==========================================

// These functions bridge to internal implementations
// They will be implemented to connect to internal/mcp packages

// MCPManagerFactory is a function type for creating MCP managers
type MCPManagerFactory func(config MCPConfig) (MCPManager, error)

// Global variable to hold the factory function
var mcpManagerFactory MCPManagerFactory

// SetMCPManagerFactory allows setting a custom factory for creating MCP managers
// This enables dependency injection while keeping the core package free of internal imports
func SetMCPManagerFactory(factory MCPManagerFactory) {
	mcpManagerFactory = factory
}

// MCPCacheManagerFactory creates MCPCacheManager implementations.
type MCPCacheManagerFactory func(config MCPCacheConfig) (MCPCacheManager, error)

var mcpCacheManagerFactory MCPCacheManagerFactory

// SetMCPCacheManagerFactory sets the cache manager factory from a plugin.
func SetMCPCacheManagerFactory(factory MCPCacheManagerFactory) { mcpCacheManagerFactory = factory }

// FunctionToolRegistryFactory creates FunctionToolRegistry implementations.
type FunctionToolRegistryFactory func() (FunctionToolRegistry, error)

var functionToolRegistryFactory FunctionToolRegistryFactory

// SetFunctionToolRegistryFactory sets the function tool registry factory from a plugin.
func SetFunctionToolRegistryFactory(factory FunctionToolRegistryFactory) {
	functionToolRegistryFactory = factory
}

// createMCPManagerInternal creates an MCP manager through the configured factory.
func createMCPManagerInternal(config MCPConfig) (MCPManager, error) {
	if mcpManagerFactory != nil {
		// Use the real factory if it's been set
		return mcpManagerFactory(config)
	}
	// No plugin has registered a factory yet; provide actionable guidance.
	return nil, fmt.Errorf("no MCP manager plugin registered: add a blank import for github.com/kunalkushwaha/agenticgokit/plugins/mcp/default (or a transport plugin like /tcp, /stdio, or /websocket)")
}

// createMCPCacheManagerInternal creates a cache manager through internal factory.
func createMCPCacheManagerInternal(config MCPCacheConfig) (MCPCacheManager, error) {
	if mcpCacheManagerFactory != nil {
		return mcpCacheManagerFactory(config)
	}
	return nil, fmt.Errorf("no MCP cache manager plugin registered: add a blank import for github.com/kunalkushwaha/agenticgokit/plugins/mcp/default or a cache plugin")
}

// createMCPToolRegistryInternal creates a tool registry through internal factory.
func createMCPToolRegistryInternal() (FunctionToolRegistry, error) {
	if functionToolRegistryFactory != nil {
		return functionToolRegistryFactory()
	}
	return nil, fmt.Errorf("no MCP function tool registry plugin registered: add a blank import for github.com/kunalkushwaha/agenticgokit/plugins/mcp/default or a registry plugin")
}

// initializeProductionMetrics initializes production metrics.
func initializeProductionMetrics(config MetricsConfig) error {
	if !config.Enabled {
		Logger().Debug().Msg("Production metrics disabled")
		return nil
	}

	// Initialize basic metrics tracking
	Logger().Debug().
		Int("port", config.Port).
		Str("path", config.Path).
		Bool("prometheus", config.PrometheusEnabled).
		Msg("Initializing production metrics")

	// Real metrics exporters (Prometheus, OTEL) should be provided by plugins.
	return nil
}

// mcpFunctionTool wraps an MCP tool as a FunctionTool
type mcpFunctionTool struct {
	toolInfo MCPToolInfo
	manager  MCPManager
}

func newMCPFunctionTool(toolInfo MCPToolInfo, manager MCPManager) FunctionTool {
	return &mcpFunctionTool{
		toolInfo: toolInfo,
		manager:  manager,
	}
}

func (t *mcpFunctionTool) Name() string {
	return t.toolInfo.Name
}

func (t *mcpFunctionTool) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
	// Execute the MCP tool using a manager that implements MCPToolExecutor
	exec, ok := t.manager.(MCPToolExecutor)
	if !ok {
		return nil, fmt.Errorf("MCP manager does not support direct tool execution. Import a transport plugin (e.g., plugins/mcp/tcp, plugins/mcp/stdio, or plugins/mcp/websocket)")
	}

	result, err := exec.ExecuteTool(ctx, t.toolInfo.Name, args)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("tool execution failed: %s", result.Error)
	}

	// Convert MCPContent to simple map
	response := make(map[string]any)
	response["success"] = result.Success
	response["duration"] = result.Duration.String()

	if len(result.Content) > 0 {
		content := make([]map[string]any, len(result.Content))
		for i, c := range result.Content {
			content[i] = map[string]any{
				"type":     c.Type,
				"text":     c.Text,
				"data":     c.Data,
				"mimeType": c.MimeType,
				"metadata": c.Metadata,
			}
		}
		response["content"] = content
	}

	return response, nil
}

// (Real MCP initialization moved to transport plugins under plugins/mcp/*)

// ==========================================
// SECTION: MCP UTILITY FUNCTIONS FOR AGENTS
// ==========================================

// FormatToolsPromptForLLM creates a prompt section describing available MCP tools with their schemas
// This function formats MCP tool information into a comprehensive prompt that helps LLMs
// understand what tools are available and how to use them according to their schemas.
func FormatToolsPromptForLLM(tools []MCPToolInfo) string {
	if len(tools) == 0 {
		return ""
	}

	prompt := "\n\nAvailable MCP tools:\n"
	for _, tool := range tools {
		prompt += fmt.Sprintf("\n**%s**: %s\n", tool.Name, tool.Description)

		// Include the schema information from MCP discovery
		if len(tool.Schema) > 0 {
			prompt += "Schema: "
			schemaStr := FormatSchemaForLLM(tool.Schema)
			prompt += schemaStr + "\n"
		}
	}

	prompt += `
To use a tool, you MUST respond with a tool call in this exact JSON format:
TOOL_CALL{"name": "tool_name", "args": {arguments_according_to_schema}}

When a user asks you to search for something, use the search tool.
When a user asks you to fetch web content, use the fetch_content tool.
When a user asks about Docker, use the docker tool.

Important:
- Use the exact parameter names and types specified in each tool's schema
- Make tool calls immediately when they would help answer the user's question
- If the user specifically asks you to use a tool, you MUST use it

Example for a search (replace with actual search terms):
TOOL_CALL{"name": "search", "args": {"query": "search terms here", "max_results": 10}}

Use these tools to provide comprehensive and accurate responses.`

	return prompt
}

// FormatSchemaForLLM converts a tool schema map to a readable string format for the LLM
// This function takes a JSON schema from MCP tool discovery and formats it in a way
// that LLMs can easily understand and use to make proper tool calls.
func FormatSchemaForLLM(schema map[string]interface{}) string {
	if schema == nil {
		return "No schema available"
	}

	var result strings.Builder

	// Handle the "type" field
	if schemaType, ok := schema["type"].(string); ok {
		result.WriteString(fmt.Sprintf("Type: %s", schemaType))
	}

	// Handle "properties" field (for object types)
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		result.WriteString("\nParameters:\n")
		for propName, propDetails := range properties {
			if propMap, ok := propDetails.(map[string]interface{}); ok {
				propType := "unknown"
				if t, exists := propMap["type"]; exists {
					if typeStr, ok := t.(string); ok {
						propType = typeStr
					}
				}

				description := ""
				if desc, exists := propMap["description"]; exists {
					if descStr, ok := desc.(string); ok {
						description = fmt.Sprintf(" - %s", descStr)
					}
				}

				result.WriteString(fmt.Sprintf("  - %s (%s)%s\n", propName, propType, description))
			}
		}
	}

	// Handle "required" field
	if required, ok := schema["required"].([]interface{}); ok {
		if len(required) > 0 {
			result.WriteString("Required parameters: ")
			for i, req := range required {
				if reqStr, ok := req.(string); ok {
					if i > 0 {
						result.WriteString(", ")
					}
					result.WriteString(reqStr)
				}
			}
			result.WriteString("\n")
		}
	}

	return result.String()
}

// ParseLLMToolCalls extracts tool calls from LLM response content
// This function parses TOOL_CALL{} patterns from LLM responses and does NOT add
// any hardcoded auto-detection logic. It trusts the LLM to make proper tool calls
// based on the provided MCP schemas.
func ParseLLMToolCalls(content string) []map[string]interface{} {
	var toolCalls []map[string]interface{}

	// Debug: Log what we're trying to parse
	logger := Logger()
	logger.Debug().Str("content", content).Msg("Parsing tool calls from LLM response")

	// Parse TOOL_CALL{...} patterns from LLM response
	parts := strings.Split(content, "TOOL_CALL")
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		logger.Debug().Str("part", part).Msg("Processing TOOL_CALL part")

		if strings.HasPrefix(part, "{") {
			// Find the closing brace
			braceCount := 0
			endIndex := -1
			for j, char := range part {
				if char == '{' {
					braceCount++
				} else if char == '}' {
					braceCount--
					if braceCount == 0 {
						endIndex = j
						break
					}
				}
			}

			if endIndex > 0 {
				jsonStr := part[:endIndex+1]
				logger.Debug().Str("json_str", jsonStr).Msg("Extracted JSON string")

				// Parse the JSON string
				toolCall := ParseToolCallJSON(jsonStr)
				logger.Debug().Interface("parsed_tool_call", toolCall).Msg("Parsed tool call")

				if len(toolCall) > 0 {
					toolCalls = append(toolCalls, toolCall)
				}
			}
		}
	}

	// NO AUTO-DETECTION: The LLM should decide when to use tools based on the provided schemas
	// Trust the LLM to make proper tool calls when needed according to the MCP tool schemas
	logger.Debug().Interface("final_tool_calls", toolCalls).Msg("Final parsed tool calls")
	return toolCalls
}

// ParseToolCallJSON is a robust JSON parser for tool calls
// This function attempts to parse JSON using the standard library first,
// then falls back to a simple parser for malformed JSON from LLMs.
func ParseToolCallJSON(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})

	// Try to parse as proper JSON first
	if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
		return result
	}

	// Fall back to simple parsing if JSON unmarshal fails
	// Remove outer braces
	jsonStr = strings.Trim(jsonStr, "{}")

	// Split by commas (simple approach)
	parts := strings.Split(jsonStr, ",")

	for _, part := range parts {
		if strings.Contains(part, ":") {
			keyValue := strings.SplitN(part, ":", 2)
			if len(keyValue) == 2 {
				key := strings.Trim(keyValue[0], " \"")
				value := strings.Trim(keyValue[1], " \"")

				// Try to parse nested objects for args
				if key == "args" && strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
					argsMap := ParseToolCallJSON(value)
					result[key] = argsMap
				} else {
					result[key] = value
				}
			}
		}
	}

	return result
}

// init function to automatically set up real MCP implementation when available
func init() {
	// This will be called when the core package is initialized
	// If the factory package is available, it will register itself
	//Logger().Debug().Msg("Core MCP package initialized")
}

// =============================================================================
// MCP AGENT TYPES (MINIMAL PUBLIC INTERFACE)
// =============================================================================

// MCPAgentConfig holds configuration for MCP-aware agents.
type MCPAgentConfig struct {
	// Tool selection settings
	MaxToolsPerExecution int           `toml:"max_tools_per_execution"`
	ToolSelectionTimeout time.Duration `toml:"tool_selection_timeout"`

	// Execution settings
	ParallelExecution bool          `toml:"parallel_execution"`
	ExecutionTimeout  time.Duration `toml:"execution_timeout"`
	RetryFailedTools  bool          `toml:"retry_failed_tools"`
	MaxRetries        int           `toml:"max_retries"`

	// LLM integration settings
	UseToolDescriptions  bool   `toml:"use_tool_descriptions"`
	ToolSelectionPrompt  string `toml:"tool_selection_prompt"`
	ResultInterpretation bool   `toml:"result_interpretation"`

	// Cache settings
	EnableCaching bool           `toml:"enable_caching"`
	CacheConfig   MCPCacheConfig `toml:"cache"`
}

// MCPAwareAgent interface for MCP-enabled agents
type MCPAwareAgent interface {
	Agent
	GetMCPManager() MCPManager
	GetConfig() MCPAgentConfig
}

// DefaultMCPAgentConfig returns a default configuration for MCP agents.
func DefaultMCPAgentConfig() MCPAgentConfig {
	return MCPAgentConfig{
		MaxToolsPerExecution: 5,
		ToolSelectionTimeout: 30 * time.Second,
		ParallelExecution:    false,
		ExecutionTimeout:     60 * time.Second,
		RetryFailedTools:     true,
		MaxRetries:           3,
		UseToolDescriptions:  true,
		ResultInterpretation: true,
		EnableCaching:        true,
		CacheConfig:          DefaultMCPCacheConfig(),
	}
}

// NewMCPAwareAgent creates a new MCP-aware agent
// Implementation is provided by internal packages
func NewMCPAwareAgent(name string, llmProvider ModelProvider, mcpManager MCPManager, config MCPAgentConfig) MCPAwareAgent {
	if mcpAwareAgentFactory != nil {
		return mcpAwareAgentFactory(name, llmProvider, mcpManager, config)
	}
	Logger().Warn().Msg("Using basic MCP agent - import internal/mcp for full functionality")
	return &basicMCPAwareAgent{name: name, config: config}
}

// RegisterMCPAwareAgentFactory registers the MCP-aware agent factory function
func RegisterMCPAwareAgentFactory(factory func(string, ModelProvider, MCPManager, MCPAgentConfig) MCPAwareAgent) {
	mcpAwareAgentFactory = factory
}

var mcpAwareAgentFactory func(string, ModelProvider, MCPManager, MCPAgentConfig) MCPAwareAgent

// basicMCPAwareAgent provides a minimal implementation
type basicMCPAwareAgent struct {
	name   string
	config MCPAgentConfig
}

func (a *basicMCPAwareAgent) Run(ctx context.Context, inputState State) (State, error) {
	return inputState, nil
}

func (a *basicMCPAwareAgent) HandleEvent(ctx context.Context, event Event, state State) (AgentResult, error) {
	startTime := time.Now()
	outputState, err := a.Run(ctx, state)
	endTime := time.Now()

	result := AgentResult{
		OutputState: outputState,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(startTime),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

func (a *basicMCPAwareAgent) Name() string {
	return a.name
}

func (a *basicMCPAwareAgent) GetRole() string {
	return "mcp_basic_agent"
}

func (a *basicMCPAwareAgent) GetDescription() string {
	return fmt.Sprintf("Basic MCP-aware agent: %s", a.name)
}

func (a *basicMCPAwareAgent) GetCapabilities() []string {
	return []string{"mcp_processing", "basic_processing"}
}

func (a *basicMCPAwareAgent) GetSystemPrompt() string {
	return "You are a basic MCP-aware agent."
}

func (a *basicMCPAwareAgent) GetTimeout() time.Duration {
	return 30 * time.Second
}

func (a *basicMCPAwareAgent) IsEnabled() bool {
	return true
}

func (a *basicMCPAwareAgent) GetLLMConfig() *ResolvedLLMConfig {
	return nil
}

func (a *basicMCPAwareAgent) Initialize(ctx context.Context) error {
	return nil
}

func (a *basicMCPAwareAgent) Shutdown(ctx context.Context) error {
	return nil
}

func (a *basicMCPAwareAgent) GetMCPManager() MCPManager {
	return nil
}

func (a *basicMCPAwareAgent) GetConfig() MCPAgentConfig {
	return a.config
}
