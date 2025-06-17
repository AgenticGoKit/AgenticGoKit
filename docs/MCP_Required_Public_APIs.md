# Required Public API Additions for MCP Integration

**Status**: SPECIFIC IMPLEMENTATION REQUIREMENTS  
**Priority**: CRITICAL for `agentcli create` command  
**Focus**: Bridge internal/mcp implementation to public core/ API  

## Critical Missing Functions

### 1. Cache Manager Public APIs (Required)

**File**: `core/mcp_cache.go` - ADD THESE FUNCTIONS:

```go
// InitializeMCPCacheManager initializes the global MCP cache manager
func InitializeMCPCacheManager(config MCPCacheConfig) error {
    // Bridge to internal/mcp.NewCacheManager()
}

// GetMCPCacheManager returns the global cache manager instance  
func GetMCPCacheManager() MCPCacheManager {
    // Return global cache manager instance
}

// NewMCPCacheManager creates a new cache manager (for custom setups)
func NewMCPCacheManager(config MCPCacheConfig) (MCPCacheManager, error) {
    // Direct bridge to internal implementation
}

// ShutdownMCPCacheManager cleanly shuts down the cache manager
func ShutdownMCPCacheManager() error {
    // Cleanup global cache manager
}
```

### 2. Unified MCP Setup (Required)

**File**: `core/mcp_unified.go` - NEW FILE:

```go
// MCPSetupConfig combines all MCP features for agentcli create
type MCPSetupConfig struct {
    Basic      MCPConfig        `toml:"mcp"`
    Cache      MCPCacheConfig   `toml:"cache"`
    Production *ProductionConfig `toml:"production"`
    
    // Feature flags
    CacheEnabled      bool   `toml:"cache_enabled"`
    ProductionEnabled bool   `toml:"production_enabled"`
    MetricsEnabled    bool   `toml:"metrics_enabled"`
    MetricsPort       int    `toml:"metrics_port"`
}

// InitializeUnifiedMCP sets up complete MCP stack based on config
func InitializeUnifiedMCP(ctx context.Context, config MCPSetupConfig) error

// CreateMCPEnabledAgent creates agent with all configured features
func CreateMCPEnabledAgent(name string, llmProvider LLMProvider) (*MCPAwareAgent, error)

// LoadMCPConfigFromTOML loads unified config from agentflow.toml
func LoadMCPConfigFromTOML(configPath string) (MCPSetupConfig, error)
```

### 3. Simple MCP Helpers (Required)

**File**: `core/mcp_simple.go` - NEW FILE:

```go
// QuickStartMCP initializes MCP with minimal configuration
func QuickStartMCP(tools ...string) error

// CreateBasicMCPAgent creates agent with default tools
func CreateBasicMCPAgent(name string, llmProvider LLMProvider, tools ...string) (*MCPAwareAgent, error)

// ExecuteMCPTool simple one-shot tool execution
func ExecuteMCPTool(ctx context.Context, toolName string, args map[string]interface{}) (MCPToolResult, error)
```

## Bridge Functions Needed

### Internal to Public Bridges

**File**: `core/mcp_factory.go` - ADD THESE:

```go
// createMCPManagerInternal - ALREADY EXISTS, needs to be implemented
func createMCPManagerInternal(config MCPConfig) (MCPManager, error) {
    // Bridge to internal/mcp implementation
}

// createMCPCacheManagerInternal - NEW BRIDGE FUNCTION
func createMCPCacheManagerInternal(config MCPCacheConfig) (MCPCacheManager, error) {
    // Bridge to internal/mcp.NewCacheManager()
}

// createMCPToolRegistryInternal - ALREADY EXISTS, needs implementation
func createMCPToolRegistryInternal() (FunctionToolRegistry, error) {
    // Bridge to internal registry
}
```

## Generated Project Template Requirements

### agentflow.toml Template Enhancement

**Current**: Basic MCP config exists in `core/config.go`
**Required**: Support unified MCP configuration

```toml
[agentflow]
provider = "{{.Provider}}"

[mcp]
enabled = {{.MCPEnabled}}
enable_discovery = true
discovery_timeout_ms = 10000

{{if .WithCache}}
[cache]
enabled = true
backend = "{{.CacheBackend}}"
default_ttl = "5m"
max_size_mb = 100

[cache.tool_ttls]
web_search = "2m"
summarize = "10m"
{{end}}

{{if .WithMetrics}}
[metrics]
enabled = true
port = {{.MetricsPort}}
path = "/metrics"
{{end}}

{{if .MCPProduction}}
[production]
connection_pool_size = {{.ConnectionPoolSize}}
retry_policy = "{{.RetryPolicy}}"
enable_load_balancer = {{.WithLoadBalancer}}
{{end}}

{{range .MCPServers}}
[[mcp.servers]]
name = "{{.Name}}"
type = "stdio"
command = "mcp-server-{{.Name}}"
enabled = true
{{end}}
```

### Generated main.go Template

**Required Pattern**: Use ONLY public core/ APIs

```go
package main

import (
    "context"
    "log"
    
    "github.com/kunalkushwaha/agentflow/core"  // ONLY public import
)

func main() {
    ctx := context.Background()
    
    // 1. Load unified MCP configuration
    config, err := core.LoadMCPConfigFromTOML("agentflow.toml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 2. Initialize unified MCP stack
    if err := core.InitializeUnifiedMCP(ctx, config); err != nil {
        log.Fatalf("Failed to initialize MCP: %v", err)
    }
    defer core.ShutdownUnifiedMCP()
    
    // 3. Create MCP-enabled agents
    llmProvider := // ... based on config.Provider
    
    {{range .Agents}}
    {{.Name}}, err := core.CreateMCPEnabledAgent("{{.Name}}", llmProvider)
    if err != nil {
        log.Fatalf("Failed to create {{.Name}}: %v", err)
    }
    {{end}}
    
    // 4. Setup and run workflow
    // ... rest of generated workflow code
}
```

## Implementation Order

### Phase 1: Critical Public APIs (BLOCKING)
1. ✅ **Add cache manager factory functions** - `core/mcp_cache.go`
2. ✅ **Create unified setup APIs** - `core/mcp_unified.go`  
3. ✅ **Add simple helper functions** - `core/mcp_simple.go`
4. ✅ **Implement internal bridges** - `core/mcp_factory.go`

### Phase 2: Configuration Integration (BLOCKING)
1. ✅ **Extend config loading** - Support unified MCP config in `core/config.go`
2. ✅ **Update TOML parsing** - Handle all MCP sections
3. ✅ **Add validation** - Ensure config compatibility

### Phase 3: Generated Templates (DEPENDS ON 1-2)
1. ✅ **Update scaffold system** - Add MCP project generation
2. ✅ **Create CLI command** - `cmd/agentcli/cmd/create.go`
3. ✅ **Add TOML templates** - Complete agentflow.toml generation
4. ✅ **Create code templates** - MCP-enabled main.go and agent files

## Validation Criteria

### ✅ **Public API Completeness**
- Generated projects import ONLY `github.com/kunalkushwaha/agentflow/core`
- No `internal/mcp` imports in generated code
- All MCP features accessible through public APIs

### ✅ **Functional Requirements**
- Basic MCP: Agent + tool execution
- Enhanced MCP: + caching
- Production MCP: + metrics, pooling, load balancing
- All features configurable via `agentflow.toml`

### ✅ **Generated Project Quality**
- Compiles out-of-box: `go build .`
- Runs successfully: `go run .`
- Proper error handling and logging
- Production-ready patterns

---

**DECISION POINT**: Should we proceed with implementing these missing public APIs before continuing with the `agentcli create` command, or would you like to modify the approach?

**RECOMMENDATION**: Implement the missing public APIs first to ensure a clean, maintainable architecture for generated projects.
