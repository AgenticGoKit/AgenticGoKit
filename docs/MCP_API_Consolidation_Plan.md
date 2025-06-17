# MCP Public API Consolidation Design

**Problem**: 6 separate MCP files (1,444 lines) create a fragmented public API
**Solution**: Consolidate into 2-3 focused files with clear responsibilities

## Current Fragmentation Analysis

### Current Files (TOO MANY):
- `mcp.go` (149 lines) - Core interfaces and types
- `mcp_agent.go` (585 lines) - Agent implementation and interface
- `mcp_cache.go` (143 lines) - Cache interfaces and helpers  
- `mcp_factory.go` (204 lines) - Global factory functions
- `mcp_helpers.go` (84 lines) - Configuration helpers
- `mcp_production.go` (279 lines) - Production configuration

**Total**: 6 files, 1,444 lines

## Proposed Consolidation (CLEAN)

### **File 1: `core/mcp.go` (Primary MCP API)**
**Purpose**: All core MCP interfaces, types, and main factory functions
**Content**: ~400-500 lines

```go
package core

// INTERFACES
type MCPManager interface { ... }
type MCPAgent interface { ... } 
type MCPCache interface { ... }
type MCPCacheManager interface { ... }

// CORE TYPES  
type MCPConfig struct { ... }
type MCPCacheConfig struct { ... }
type MCPToolExecution struct { ... }
type MCPToolResult struct { ... }
type MCPServerInfo struct { ... }

// FACTORY FUNCTIONS
func InitializeMCP(config MCPConfig) error
func InitializeMCPWithCache(mcpConfig MCPConfig, cacheConfig MCPCacheConfig) error  
func InitializeProductionMCP(ctx context.Context, config ProductionConfig) error
func GetMCPManager() MCPManager
func GetMCPCacheManager() MCPCacheManager

// AGENT FACTORIES
func NewMCPAgent(name string, llmProvider LLMProvider) (*MCPAwareAgent, error)
func NewMCPAgentWithCache(name string, llmProvider LLMProvider) (*MCPAwareAgent, error)
func NewProductionMCPAgent(name string, llmProvider LLMProvider, config ProductionConfig) (*MCPAwareAgent, error)

// SIMPLE HELPERS
func QuickStartMCP(tools ...string) error
func ConnectMCPServer(name, serverType, endpoint string) error
func ExecuteMCPTool(ctx context.Context, toolName string, args map[string]interface{}) (MCPToolResult, error)

// CONFIGURATION HELPERS
func DefaultMCPConfig() MCPConfig
func DefaultMCPCacheConfig() MCPCacheConfig  
func DefaultProductionConfig() ProductionConfig
func LoadMCPConfigFromTOML(path string) (MCPConfig, error)

// SHUTDOWN
func ShutdownMCP() error
```

### **File 2: `core/mcp_agent.go` (Agent Implementation)**
**Purpose**: MCPAwareAgent implementation only
**Content**: ~300-400 lines (keep existing implementation)

```go
package core

// MCPAwareAgent implementation with all methods
type MCPAwareAgent struct { ... }

func (a *MCPAwareAgent) Run(ctx context.Context, inputState State) (State, error) { ... }
func (a *MCPAwareAgent) SelectTools(ctx context.Context, query string, stateContext State) ([]string, error) { ... }
func (a *MCPAwareAgent) ExecuteTools(ctx context.Context, tools []MCPToolExecution) ([]MCPToolResult, error) { ... }
// ... all existing agent implementation methods
```

### **File 3: `core/mcp_production.go` (Optional - Production Only)**
**Purpose**: Advanced production configurations and types
**Content**: ~200-250 lines

```go
package core

// ProductionConfig with all advanced options
type ProductionConfig struct { ... }

// Production-specific helpers
func ValidateProductionConfig(config ProductionConfig) error { ... }
func OptimizeForProduction(config ProductionConfig) ProductionConfig { ... }
// ... production-specific functions only
```

## Consolidation Strategy

### **Merge Plan**:

1. **Merge into `core/mcp.go`**:
   - All interfaces from `mcp.go`
   - All factory functions from `mcp_factory.go`  
   - All configuration helpers from `mcp_helpers.go`
   - Cache interfaces and helpers from `mcp_cache.go`
   - Simple helper functions (new)

2. **Keep separate `core/mcp_agent.go`**:
   - Only the MCPAwareAgent implementation
   - All agent-specific methods and logic

3. **Keep separate `core/mcp_production.go`** (optional):
   - Only for advanced production configurations
   - Could be merged into main `mcp.go` if small enough

## Simplified Public API for Generated Projects

### **Single Import Pattern**:
```go
import "github.com/kunalkushwaha/agentflow/core"

// Everything MCP-related available through core package
```

### **Basic MCP (10 lines)**:
```go
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

### **Production MCP (15 lines)**:
```go
func main() {
    ctx := context.Background()
    
    // Load config from agentflow.toml
    config, _ := core.LoadMCPConfigFromTOML("agentflow.toml")
    
    // Initialize production MCP stack
    core.InitializeProductionMCP(ctx, config.Production)
    defer core.ShutdownMCP()
    
    // Create production agent
    llmProvider := core.NewAzureLLMProvider(azureConfig)
    agent, _ := core.NewProductionMCPAgent("prod_agent", llmProvider, config.Production)
    
    // Run workflow...
}
```

## Benefits of Consolidation

### ✅ **Developer Experience**:
- **Single source of truth**: All MCP APIs in one place
- **Reduced cognitive load**: Don't need to remember which file has what
- **Better IDE support**: All functions discoverable in one file
- **Simpler documentation**: One file to document

### ✅ **Maintenance**:
- **Fewer files to maintain**: 2-3 files instead of 6
- **Easier refactoring**: Related functions in same file
- **Simpler testing**: Less cross-file dependencies
- **Cleaner imports**: Generated projects import less

### ✅ **API Clarity**:
- **Clear progression**: Basic → Enhanced → Production in one file
- **Logical grouping**: Related functions together
- **Easier discovery**: All MCP capabilities visible at once

## Migration Strategy

### **Phase 1: Create Consolidated `core/mcp.go`**
1. Move all interfaces from existing `mcp.go`
2. Move all factory functions from `mcp_factory.go`
3. Move configuration helpers from `mcp_helpers.go` 
4. Move cache interfaces from `mcp_cache.go`
5. Add missing factory functions for cache and production
6. Add simple helper functions

### **Phase 2: Clean Up Remaining Files**
1. Keep only agent implementation in `mcp_agent.go`
2. Decide on `mcp_production.go` (merge or keep separate)
3. Remove redundant files: `mcp_factory.go`, `mcp_helpers.go`, `mcp_cache.go`

### **Phase 3: Update All References**
1. Update generated project templates
2. Update examples to use consolidated API
3. Update documentation

## Decision Point

**Question**: Should we:
- **Option A**: Consolidate to 1 large file (`core/mcp.go` ~800-900 lines)
- **Option B**: Consolidate to 2 files (`core/mcp.go` ~500 lines + `core/mcp_agent.go` ~400 lines)  
- **Option C**: Consolidate to 3 files (+ keep `mcp_production.go` separate)

**Recommendation**: **Option B** - 2 files is the sweet spot:
- `core/mcp.go`: All interfaces, types, factories, helpers
- `core/mcp_agent.go`: Agent implementation only

This keeps the public API simple while maintaining reasonable file sizes.
