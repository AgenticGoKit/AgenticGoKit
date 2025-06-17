# AgentFlow MCP Public API Design (REVISED - CONSOLIDATED)

**Created**: June 17, 2025  
**Status**: Design Phase - SIMPLIFIED  
**Focus**: Clean, Consolidated Public API Surface  

## ❌ Current Problem: Too Many Files
**Current**: 6 MCP files (1,444 lines) - `mcp.go`, `mcp_agent.go`, `mcp_cache.go`, `mcp_factory.go`, `mcp_helpers.go`, `mcp_production.go`  
**Problem**: Fragmented API, complex imports, maintenance overhead

## ✅ Proposed Solution: 2-File Consolidation

### **File 1: `core/mcp.go` (ALL MCP APIs)**
**Purpose**: Complete MCP public interface in one place  
**Size**: ~500-600 lines  

```go
package core

// ==================== INTERFACES ====================
type MCPManager interface {
    Connect(ctx context.Context, serverName string) error
    Disconnect(serverName string) error
    DiscoverServers(ctx context.Context) ([]MCPServerInfo, error)
    RefreshTools(ctx context.Context) error
    GetAvailableTools() []MCPToolInfo
    HealthCheck(ctx context.Context) map[string]MCPHealthStatus
}

type MCPAgent interface {
    Agent
    SelectTools(ctx context.Context, query string, stateContext State) ([]string, error)
    ExecuteTools(ctx context.Context, tools []MCPToolExecution) ([]MCPToolResult, error)
    GetAvailableMCPTools() []MCPToolInfo
}

type MCPCache interface {
    Get(ctx context.Context, key MCPCacheKey) (*MCPCachedResult, error)
    Set(ctx context.Context, key MCPCacheKey, result MCPToolResult, ttl time.Duration) error
    Stats(ctx context.Context) (MCPCacheStats, error)
}

type MCPCacheManager interface {
    GetCache(toolName, serverName string) MCPCache
    ExecuteWithCache(ctx context.Context, execution MCPToolExecution) (MCPToolResult, error)
}

// ==================== CONFIGURATION TYPES ====================
type MCPConfig struct { ... }
type MCPCacheConfig struct { ... }  
type ProductionConfig struct { ... }
type MCPToolExecution struct { ... }
type MCPToolResult struct { ... }

// ==================== INITIALIZATION (3 LEVELS) ====================
// Level 1: Basic MCP
func InitializeMCP(config MCPConfig) error
func QuickStartMCP(tools ...string) error

// Level 2: MCP + Caching  
func InitializeMCPWithCache(mcpConfig MCPConfig, cacheConfig MCPCacheConfig) error

// Level 3: Production MCP (All Features)
func InitializeProductionMCP(ctx context.Context, config ProductionConfig) error

// ==================== AGENT FACTORIES (3 LEVELS) ====================
// Level 1: Basic MCP Agent
func NewMCPAgent(name string, llmProvider LLMProvider) (*MCPAwareAgent, error)

// Level 2: MCP Agent + Caching
func NewMCPAgentWithCache(name string, llmProvider LLMProvider) (*MCPAwareAgent, error)

// Level 3: Production MCP Agent (All Features)
func NewProductionMCPAgent(name string, llmProvider LLMProvider, config ProductionConfig) (*MCPAwareAgent, error)

// ==================== SIMPLE HELPERS ====================
func ConnectMCPServer(name, serverType, endpoint string) error
func ExecuteMCPTool(ctx context.Context, toolName string, args map[string]interface{}) (MCPToolResult, error)

// ==================== CONFIGURATION HELPERS ====================
func DefaultMCPConfig() MCPConfig
func DefaultMCPCacheConfig() MCPCacheConfig
func DefaultProductionConfig() ProductionConfig
func LoadMCPConfigFromTOML(path string) (MCPConfig, error)

// ==================== GLOBAL ACCESS ====================
func GetMCPManager() MCPManager
func GetMCPCacheManager() MCPCacheManager

// ==================== SHUTDOWN ====================
func ShutdownMCP() error
```

### **File 2: `core/mcp_agent.go` (Implementation Only)**
**Purpose**: MCPAwareAgent implementation  
**Size**: ~400-500 lines (existing implementation)

## Simplified Public API Usage Patterns

### **Pattern 1: Basic MCP (5 lines)**
```go
import "github.com/kunalkushwaha/agentflow/core"

func main() {
    core.QuickStartMCP("web_search", "summarize")
    llmProvider := core.NewMockLLMProvider()
    agent, _ := core.NewMCPAgent("agent1", llmProvider)
    
    state := core.NewState()
    state.Set("input", "search for AI news")
    result, _ := agent.Run(context.Background(), state)
    fmt.Printf("Result: %v\n", result.Get("output"))
}
```

### **Pattern 2: MCP with Caching (8 lines)**
```go
import "github.com/kunalkushwaha/agentflow/core"

func main() {
    mcpConfig := core.DefaultMCPConfig()
    cacheConfig := core.DefaultMCPCacheConfig()
    
    core.InitializeMCPWithCache(mcpConfig, cacheConfig)
    defer core.ShutdownMCP()
    
    llmProvider := core.NewOllamaLLMProvider()
    agent, _ := core.NewMCPAgentWithCache("cached_agent", llmProvider)
    
    // ... workflow execution
}
```

### **Pattern 3: Production MCP (10 lines)**
```go
import "github.com/kunalkushwaha/agentflow/core"

func main() {
    ctx := context.Background()
    
    // Load from agentflow.toml
    config, _ := core.LoadMCPConfigFromTOML("agentflow.toml")
    productionConfig := core.DefaultProductionConfig()
    
    core.InitializeProductionMCP(ctx, productionConfig)
    defer core.ShutdownMCP()
    
    llmProvider := core.NewAzureLLMProvider(azureConfig)  
    agent, _ := core.NewProductionMCPAgent("prod_agent", llmProvider, productionConfig)
    
    // ... production workflow
}
```

## Key Benefits of Consolidation

### ✅ **Developer Experience**
- **Single import**: Only `github.com/kunalkushwaha/agentflow/core`
- **Progressive complexity**: Basic → Cached → Production APIs
- **Discoverable**: All MCP functions in one place
- **IDE friendly**: Better autocomplete and documentation

### ✅ **Generated Project Simplicity** 
```go
// agentcli create myproject --with-mcp --with-cache
// Generated main.go:

package main

import (
    "context"
    "log"
    "github.com/kunalkushwaha/agentflow/core"  // ONLY import needed
)

func main() {
    ctx := context.Background()
    
    // 1. Initialize MCP (method depends on flags)
    {{if .MCPProduction}}
    config := core.DefaultProductionConfig()
    core.InitializeProductionMCP(ctx, config)
    {{else if .WithCache}}
    mcpConfig := core.DefaultMCPConfig()
    cacheConfig := core.DefaultMCPCacheConfig()  
    core.InitializeMCPWithCache(mcpConfig, cacheConfig)
    {{else}}
    mcpConfig := core.DefaultMCPConfig()
    core.InitializeMCP(mcpConfig)
    {{end}}
    defer core.ShutdownMCP()
    
    // 2. Create agents (method depends on flags)
    llmProvider := // ... based on provider config
    {{range .Agents}}
    {{.Name}}, err := {{if $.MCPProduction}}core.NewProductionMCPAgent{{else if $.WithCache}}core.NewMCPAgentWithCache{{else}}core.NewMCPAgent{{end}}("{{.Name}}", llmProvider{{if $.MCPProduction}}, config{{end}})
    if err != nil {
        log.Fatalf("Failed to create {{.Name}}: %v", err)
    }
    {{end}}
    
    // 3. Run workflow
    // ... workflow logic
}
```

## Implementation Strategy (REVISED)

### **Phase 1: Consolidate Core APIs**
1. ✅ **Create new `core/mcp.go`** - All interfaces, factories, helpers in one file
2. ✅ **Clean up `core/mcp_agent.go`** - Keep only agent implementation  
3. ✅ **Remove redundant files** - Delete `mcp_factory.go`, `mcp_helpers.go`, `mcp_cache.go`
4. ✅ **Decide on `mcp_production.go`** - Merge into main or keep separate

### **Phase 2: Update Generated Templates**
1. ✅ **Simplify imports** - Only `core` package in generated projects
2. ✅ **Progressive API usage** - Basic/Cache/Production patterns
3. ✅ **Configuration integration** - Load from `agentflow.toml`

### **Phase 3: Validate and Test**
1. ✅ **Build generated projects** - Ensure compilation
2. ✅ **Test all patterns** - Basic, cached, production workflows
3. ✅ **Update documentation** - Reflect simplified API

---

**NEXT STEP**: Do you approve this consolidated 2-file approach? Should we proceed with implementing the consolidated `core/mcp.go` with all APIs in one place?
