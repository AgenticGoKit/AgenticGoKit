# MCP 2-File Consolidation Implementation Plan

**Approved Design**: 2-file structure  
**Status**: Ready for Implementation  
**Target**: Clean, consolidated public API  

## Implementation Steps

### **Step 1: Create Consolidated `core/mcp.go`**

**Action**: Merge 4 files into 1 comprehensive MCP API file

**Source Files to Consolidate**:
- `core/mcp.go` (149 lines) - Core interfaces and types
- `core/mcp_factory.go` (204 lines) - Factory functions  
- `core/mcp_helpers.go` (84 lines) - Configuration helpers
- `core/mcp_cache.go` (143 lines) - Cache interfaces and helpers

**Total**: ~580 lines → Single `core/mcp.go` file

### **Step 2: Clean Up `core/mcp_agent.go`**

**Action**: Keep only agent implementation, remove redundant imports

**Current**: 585 lines (agent implementation)  
**Target**: ~400-500 lines (clean implementation only)

### **Step 3: Handle `core/mcp_production.go`**

**Decision Options**:
- **Option A**: Merge into main `core/mcp.go` (+279 lines = ~850 lines total)
- **Option B**: Keep separate (279 lines) for advanced users only

**Recommendation**: **Option A** - Merge everything into single file for simplicity

### **Step 4: Remove Redundant Files**

**Files to Delete**:
- `core/mcp_factory.go` (merged into mcp.go)
- `core/mcp_helpers.go` (merged into mcp.go)  
- `core/mcp_cache.go` (merged into mcp.go)
- `core/mcp_production.go` (merged into mcp.go) 

## Consolidated File Structure

### **File 1: `core/mcp.go` (~850 lines)**

```go
package core

import (
    "context"
    "time"
    // ... other imports
)

// ==========================================
// SECTION 1: CORE INTERFACES (~150 lines)
// ==========================================

// MCPManager provides the main interface for managing MCP connections and tools
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

// MCPAgent represents an agent that can utilize MCP tools
type MCPAgent interface {
    Agent
    SelectTools(ctx context.Context, query string, stateContext State) ([]string, error)
    ExecuteTools(ctx context.Context, tools []MCPToolExecution) ([]MCPToolResult, error)
    GetAvailableMCPTools() []MCPToolInfo
}

// MCPCache defines the interface for caching MCP tool results
type MCPCache interface {
    Get(ctx context.Context, key MCPCacheKey) (*MCPCachedResult, error)
    Set(ctx context.Context, key MCPCacheKey, result MCPToolResult, ttl time.Duration) error
    Delete(ctx context.Context, key MCPCacheKey) error
    Clear(ctx context.Context) error
    Exists(ctx context.Context, key MCPCacheKey) (bool, error)
    Stats(ctx context.Context) (MCPCacheStats, error)
    Cleanup(ctx context.Context) error
    Close() error
}

// MCPCacheManager manages multiple cache instances and provides cache-aware tool execution
type MCPCacheManager interface {
    GetCache(toolName, serverName string) MCPCache
    ExecuteWithCache(ctx context.Context, execution MCPToolExecution) (MCPToolResult, error)
    InvalidateByPattern(ctx context.Context, pattern string) error
    GetGlobalStats(ctx context.Context) (MCPCacheStats, error)
    Shutdown() error
    Configure(config MCPCacheConfig) error
}

// ==========================================
// SECTION 2: CONFIGURATION TYPES (~200 lines)
// ==========================================

// MCPConfig holds configuration for MCP integration
type MCPConfig struct {
    EnableDiscovery   bool          `toml:"enable_discovery"`
    DiscoveryTimeout  time.Duration `toml:"discovery_timeout"`
    ScanPorts         []int         `toml:"scan_ports"`
    ConnectionTimeout time.Duration `toml:"connection_timeout"`
    MaxRetries        int           `toml:"max_retries"`
    RetryDelay        time.Duration `toml:"retry_delay"`
    EnableCaching     bool          `toml:"enable_caching"`
    CacheTimeout      time.Duration `toml:"cache_timeout"`
    MaxConnections    int           `toml:"max_connections"`
    Servers           []MCPServerConfig `toml:"servers"`
}

// MCPCacheConfig holds configuration for the cache system
type MCPCacheConfig struct {
    Enabled         bool                      `toml:"enabled"`
    DefaultTTL      time.Duration            `toml:"default_ttl"`
    MaxSize         int64                    `toml:"max_size_mb"`
    MaxKeys         int                      `toml:"max_keys"`
    EvictionPolicy  string                   `toml:"eviction_policy"`
    CleanupInterval time.Duration            `toml:"cleanup_interval"`
    ToolTTLs        map[string]time.Duration `toml:"tool_ttls"`
    Backend         string                   `toml:"backend"`
    BackendConfig   map[string]string        `toml:"backend_config"`
}

// ProductionConfig holds advanced production configuration
type ProductionConfig struct {
    // Connection pooling
    ConnectionPoolSize int           `toml:"connection_pool_size"`
    PoolTimeout        time.Duration `toml:"pool_timeout"`
    
    // Load balancing  
    LoadBalancingEnabled bool   `toml:"load_balancing_enabled"`
    LoadBalancingStrategy string `toml:"load_balancing_strategy"`
    
    // Advanced retry
    RetryPolicy       string        `toml:"retry_policy"`
    MaxRetryAttempts  int           `toml:"max_retry_attempts"`
    RetryBackoffBase  time.Duration `toml:"retry_backoff_base"`
    RetryBackoffMax   time.Duration `toml:"retry_backoff_max"`
    EnableJitter      bool          `toml:"enable_jitter"`
    
    // Circuit breaker
    CircuitBreakerEnabled   bool          `toml:"circuit_breaker_enabled"`
    FailureThreshold        int           `toml:"failure_threshold"`
    RecoveryTimeout         time.Duration `toml:"recovery_timeout"`
    HalfOpenMaxCalls        int           `toml:"half_open_max_calls"`
    
    // Monitoring
    MetricsEnabled bool   `toml:"metrics_enabled"`
    MetricsPort    int    `toml:"metrics_port"`
    MetricsPath    string `toml:"metrics_path"`
    HealthCheckEnabled bool `toml:"health_check_enabled"`
    HealthCheckInterval time.Duration `toml:"health_check_interval"`
    
    // Tracing
    TracingEnabled bool   `toml:"tracing_enabled"`
    TracingSampler string `toml:"tracing_sampler"`
    
    // Performance
    RequestTimeout       time.Duration `toml:"request_timeout"`
    MaxConcurrentRequests int          `toml:"max_concurrent_requests"`
}

// ... other types (MCPToolExecution, MCPToolResult, etc.)

// ==========================================
// SECTION 3: INITIALIZATION FUNCTIONS (~100 lines)
// ==========================================

// Level 1: Basic MCP
func InitializeMCP(config MCPConfig) error {
    // Initialize basic MCP manager
}

func QuickStartMCP(tools ...string) error {
    // Simple initialization with default config
}

// Level 2: MCP + Caching
func InitializeMCPWithCache(mcpConfig MCPConfig, cacheConfig MCPCacheConfig) error {
    // Initialize MCP manager + cache manager
}

// Level 3: Production MCP (All Features)
func InitializeProductionMCP(ctx context.Context, config ProductionConfig) error {
    // Initialize full production stack: manager, cache, metrics, pooling, etc.
}

// ==========================================
// SECTION 4: AGENT FACTORIES (~100 lines)
// ==========================================

// Level 1: Basic MCP Agent
func NewMCPAgent(name string, llmProvider LLMProvider) (*MCPAwareAgent, error) {
    // Create agent with basic MCP capabilities
}

// Level 2: MCP Agent + Caching
func NewMCPAgentWithCache(name string, llmProvider LLMProvider) (*MCPAwareAgent, error) {
    // Create agent with MCP + caching
}

// Level 3: Production MCP Agent (All Features) 
func NewProductionMCPAgent(name string, llmProvider LLMProvider, config ProductionConfig) (*MCPAwareAgent, error) {
    // Create agent with all production features
}

// ==========================================
// SECTION 5: CONFIGURATION HELPERS (~100 lines)
// ==========================================

func DefaultMCPConfig() MCPConfig {
    // Return sensible defaults for basic MCP
}

func DefaultMCPCacheConfig() MCPCacheConfig {
    // Return sensible defaults for caching
}

func DefaultProductionConfig() ProductionConfig {
    // Return production-ready defaults
}

func LoadMCPConfigFromTOML(path string) (MCPConfig, error) {
    // Load configuration from agentflow.toml
}

// ==========================================
// SECTION 6: SIMPLE HELPERS (~50 lines)
// ==========================================

func ConnectMCPServer(name, serverType, endpoint string) error {
    // Connect to single MCP server
}

func ExecuteMCPTool(ctx context.Context, toolName string, args map[string]interface{}) (MCPToolResult, error) {
    // Execute single tool with simple interface
}

// ==========================================
// SECTION 7: GLOBAL ACCESS & SHUTDOWN (~50 lines)
// ==========================================

func GetMCPManager() MCPManager {
    // Return global MCP manager
}

func GetMCPCacheManager() MCPCacheManager {
    // Return global cache manager
}

func ShutdownMCP() error {
    // Clean shutdown of all MCP components
}

// ==========================================
// SECTION 8: CACHE UTILITIES (~50 lines)
// ==========================================

func GenerateCacheKey(toolName, serverName string, args map[string]string) MCPCacheKey {
    // Generate standardized cache key
}

// ... other cache utilities
```

### **File 2: `core/mcp_agent.go` (~400-500 lines)**

**Keep existing implementation with minimal changes**:
- MCPAwareAgent struct and all methods
- Agent implementation logic
- Remove redundant imports that are now in main mcp.go

## Implementation Order

### **Phase 1: Create New Consolidated File**
1. ✅ Create new `core/mcp_consolidated.go` (temporary name)
2. ✅ Copy and organize all sections from existing files
3. ✅ Add missing factory functions for cache and production
4. ✅ Ensure all imports and dependencies work

### **Phase 2: Update Agent File**  
1. ✅ Clean up `core/mcp_agent.go`
2. ✅ Remove redundant imports
3. ✅ Update any references to moved functions

### **Phase 3: Replace and Clean Up**
1. ✅ Rename `mcp_consolidated.go` → `mcp.go`
2. ✅ Delete old files: `mcp_factory.go`, `mcp_helpers.go`, `mcp_cache.go`, `mcp_production.go`
3. ✅ Update all imports and references throughout codebase

### **Phase 4: Validate**
1. ✅ Run `go build ./core/...` to ensure compilation
2. ✅ Run existing tests to ensure functionality preserved
3. ✅ Update examples to use consolidated API

## Benefits Verification

### ✅ **Single Import**
```go
import "github.com/kunalkushwaha/agentflow/core"
// ALL MCP functionality available through core.*
```

### ✅ **Progressive Complexity**  
```go
// Basic (3 lines)
core.QuickStartMCP("web_search")
agent, _ := core.NewMCPAgent("agent1", llmProvider)

// Enhanced (5 lines) 
core.InitializeMCPWithCache(mcpConfig, cacheConfig)
agent, _ := core.NewMCPAgentWithCache("agent1", llmProvider)

// Production (7 lines)
core.InitializeProductionMCP(ctx, productionConfig)  
agent, _ := core.NewProductionMCPAgent("agent1", llmProvider, productionConfig)
```

### ✅ **Clean Generated Projects**
```go
// agentcli create myproject --with-mcp --with-cache
// Generated main.go - SIMPLE AND CLEAN

package main

import (
    "context"
    "github.com/kunalkushwaha/agentflow/core"  // ONLY IMPORT
)

func main() {
    mcpConfig := core.DefaultMCPConfig()
    cacheConfig := core.DefaultMCPCacheConfig()
    
    core.InitializeMCPWithCache(mcpConfig, cacheConfig)
    defer core.ShutdownMCP()
    
    llmProvider := // ... from config
    agent1, _ := core.NewMCPAgentWithCache("agent1", llmProvider)
    agent2, _ := core.NewMCPAgentWithCache("agent2", llmProvider)
    
    // ... workflow execution
}
```

---

**Ready to proceed with implementation?** Should I start with creating the consolidated `core/mcp.go` file?
