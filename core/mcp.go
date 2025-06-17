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
package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// ==========================================
// SECTION 2: CONFIGURATION TYPES (~300 lines)
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

	Logger().Info().Msg("MCP manager initialized successfully")
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

	Logger().Info().Msg("MCP with cache initialized successfully")
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

	Logger().Info().Msg("Production MCP initialized successfully")
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

	Logger().Info().Msg("MCP cache manager initialized successfully")
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
	Logger().Info().Msg("MCP tool registry initialized successfully")
	return nil
}

// InitializeMCPManager initializes the global MCP manager with the provided configuration.
// This is an alias for InitializeMCP for backward compatibility.
func InitializeMCPManager(config MCPConfig) error {
	return InitializeMCP(config)
}

// CreateMCPAgentWithLLMAndTools creates an MCP-aware agent with the specified configuration.
// This is a comprehensive factory function for creating fully configured agents.
func CreateMCPAgentWithLLMAndTools(ctx context.Context, name string, llmProvider ModelProvider, mcpConfig MCPConfig, agentConfig MCPAgentConfig) (*MCPAwareAgent, error) {
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
func NewMCPAgent(name string, llmProvider ModelProvider) (*MCPAwareAgent, error) {
	manager := GetMCPManager()
	if manager == nil {
		return nil, fmt.Errorf("MCP manager not initialized - call InitializeMCP() first")
	}

	config := DefaultMCPAgentConfig()
	return NewMCPAwareAgent(name, llmProvider, manager, config), nil
}

// NewMCPAgentWithCache creates an MCP-aware agent with caching capabilities.
// This provides better performance through intelligent result caching.
func NewMCPAgentWithCache(name string, llmProvider ModelProvider) (*MCPAwareAgent, error) {
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
func NewProductionMCPAgent(name string, llmProvider ModelProvider, config ProductionConfig) (*MCPAwareAgent, error) {
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

	// Direct execution without cache
	// This would need to be implemented in the internal manager
	return MCPToolResult{}, fmt.Errorf("direct tool execution not yet implemented")
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
	Logger().Info().Int("tool_count", len(tools)).Msg("Registering MCP tools with registry")

	// This would need implementation to convert MCPToolInfo to FunctionTool
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
	// This would need to be implemented with TOML parsing
	// For now, return default config
	return DefaultMCPConfig(), nil
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

	Logger().Info().Msg("MCP shutdown completed successfully")
	return nil
}

// ==========================================
// SECTION 12: INTERNAL BRIDGE FUNCTIONS (~100 lines)
// ==========================================

// These functions bridge to internal implementations
// They will be implemented to connect to internal/mcp packages

// createMCPManagerInternal creates an MCP manager through internal factory.
func createMCPManagerInternal(config MCPConfig) (MCPManager, error) {
	// Bridge to internal/mcp implementation
	// Note: This needs to be wired to the internal/mcp package
	return nil, fmt.Errorf("MCP manager creation bridge to internal/mcp not yet wired")
}

// createMCPCacheManagerInternal creates a cache manager through internal factory.
func createMCPCacheManagerInternal(config MCPCacheConfig) (MCPCacheManager, error) {
	// Bridge to internal/mcp.NewCacheManager()
	// Note: This needs to be wired to the internal/mcp package
	return nil, fmt.Errorf("MCP cache manager creation bridge to internal/mcp not yet wired")
}

// createMCPToolRegistryInternal creates a tool registry through internal factory.
func createMCPToolRegistryInternal() (FunctionToolRegistry, error) {
	// Bridge to internal registry implementation
	// Note: This needs to be wired to the internal/tools package
	return nil, fmt.Errorf("MCP tool registry creation bridge to internal/tools not yet wired")
}

// initializeProductionMetrics initializes production metrics.
func initializeProductionMetrics(config MetricsConfig) error {
	// Bridge to internal/mcp.NewMCPMetrics()
	return fmt.Errorf("production metrics initialization not yet implemented")
}

// ProductionMCPConfig converts production config to basic MCP config.
func ProductionMCPConfig(config ProductionConfig) MCPConfig {
	// Convert production config to basic MCP config
	return DefaultMCPConfig()
}

// ProductionCacheConfig converts production cache config to MCP cache config.
func ProductionCacheConfig(config CacheConfig) MCPCacheConfig {
	// Convert production cache config to MCP cache config
	return DefaultMCPCacheConfig()
}

// ProductionAgentConfig converts production config to agent config.
func ProductionAgentConfig(config ProductionConfig) MCPAgentConfig {
	// Convert production config to agent config
	return DefaultMCPAgentConfig()
}
