# MCP API Migration Guide

**Upgrading to Consolidated MCP API**

If you were using the previous fragmented MCP files, this guide helps you migrate to the new consolidated API.

## 🔄 What Changed

### **Before (Fragmented)**
```
core/
├── mcp.go          # Core interfaces only
├── mcp_factory.go  # Factory functions  
├── mcp_cache.go    # Cache interfaces
├── mcp_helpers.go  # Configuration helpers
├── mcp_production.go # Production features
└── mcp_agent.go    # Agent implementation
```

### **After (Consolidated)**
```
core/
├── mcp.go          # 🎯 Everything: interfaces + factories + config + helpers
└── mcp_agent.go    # 🤖 Agent implementation only
```

## 📦 Import Changes

### **Before**
```go
import (
    "github.com/kunalkushwaha/agentflow/core"
    // No changes needed - all functions still in core package
)
```

### **After**  
```go
import (
    "github.com/kunalkushwaha/agentflow/core"
    // Same imports work! API is backward compatible
)
```

## 🔧 Function Migration

**Good news: All existing function calls work unchanged!**

| Old Function | New Function | Status |
|-------------|--------------|---------|
| `core.InitializeMCP()` | `core.InitializeMCP()` | ✅ Unchanged |
| `core.NewMCPAgent()` | `core.NewMCPAgent()` | ✅ Unchanged |
| `core.DefaultMCPConfig()` | `core.DefaultMCPConfig()` | ✅ Unchanged |
| `core.ShutdownMCP()` | `core.ShutdownMCP()` | ✅ Unchanged |

## 🆕 New Functions Available

The consolidated API adds several new convenience functions:

```go
// Quick start (new!)
err := core.QuickStartMCP()

// Enhanced agent creation (new!)
agent, err := core.CreateMCPAgentWithLLMAndTools(
    ctx, "agent", llmProvider, mcpConfig, agentConfig)

// Backward compatibility aliases (new!)
err := core.InitializeMCPManager(config)  // alias for InitializeMCP
err := core.ShutdownMCPManager()          // alias for ShutdownMCP

// Direct tool execution (new!)
result, err := core.ExecuteMCPTool(ctx, "search", args)
```

## 📋 Step-by-Step Migration

### **Step 1: Update Your Go Modules**
```bash
go mod tidy
```

### **Step 2: Test Existing Code**
Your existing code should work without changes:

```go
// This still works exactly the same
config := core.DefaultMCPConfig()
err := core.InitializeMCP(config)
agent, err := core.NewMCPAgent("agent", llmProvider)
```

### **Step 3: Consider Upgrades (Optional)**

#### **Upgrade to QuickStart**
```go
// Old way
config := core.DefaultMCPConfig()
err := core.InitializeMCP(config)

// New simpler way  
err := core.QuickStartMCP()
```

#### **Upgrade to Enhanced Agent Creation**
```go
// Old way
err := core.InitializeMCP(mcpConfig)
agent, err := core.NewMCPAgent("agent", llmProvider)

// New comprehensive way
agent, err := core.CreateMCPAgentWithLLMAndTools(
    ctx, "agent", llmProvider, mcpConfig, agentConfig)
```

#### **Upgrade to Production Features**
```go
// Add production features
prodConfig := core.DefaultProductionConfig()
err := core.InitializeProductionMCP(ctx, prodConfig)
agent, err := core.NewProductionMCPAgent("agent", llmProvider, prodConfig)
```

## 🎯 Benefits of Migration

### **Simplified API**
- One file (`core/mcp.go`) contains everything
- Consistent function naming
- Progressive complexity (basic → cached → production)

### **New Features**
- QuickStart for rapid prototyping
- Production-ready configurations
- Enhanced caching capabilities
- Better error handling
- Comprehensive monitoring

### **Better Documentation**
- Complete usage examples
- Progressive learning path
- Production best practices

## ⚠️ Breaking Changes

**None!** The consolidation is 100% backward compatible.

All existing function signatures, types, and behavior remain identical.

## 🧪 Testing Your Migration

Create a simple test to verify everything works:

```go
package main

import (
    "context"
    "fmt" 
    "log"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func testMigration() {
    // Test 1: Basic initialization (should work as before)
    config := core.DefaultMCPConfig()
    if err := core.InitializeMCP(config); err != nil {
        log.Printf("❌ Basic init failed: %v", err)
        return
    }
    fmt.Println("✅ Basic initialization works")
    
    // Test 2: New QuickStart feature  
    if err := core.QuickStartMCP(); err != nil {
        log.Printf("❌ QuickStart failed: %v", err)
        return
    }
    fmt.Println("✅ QuickStart works")
    
    // Test 3: Agent creation (should work as before)
    llmProvider := &MockLLMProvider{} // Your LLM implementation
    agent, err := core.NewMCPAgent("test-agent", llmProvider)
    if err != nil {
        log.Printf("❌ Agent creation failed: %v", err)
        return
    }
    fmt.Printf("✅ Agent created: %s\n", agent.Name())
    
    // Test 4: Cleanup (should work as before)
    if err := core.ShutdownMCP(); err != nil {
        log.Printf("❌ Shutdown failed: %v", err)
        return  
    }
    fmt.Println("✅ Shutdown works")
    
    fmt.Println("🎉 Migration successful!")
}
```

## 📞 Need Help?

If you encounter issues during migration:

1. **Check the error message** - Most issues are configuration-related
2. **Review the [Usage Guide](MCP_API_Usage_Guide.md)** - Complete examples available
3. **Check existing examples** - See `examples/` directory
4. **File an issue** - We're here to help!

## 🔮 Future-Proofing

The new consolidated API is designed for:

- **Stability** - Function signatures won't change
- **Extensibility** - New features added without breaking changes  
- **Simplicity** - One import, progressive complexity
- **Production** - Enterprise-ready from day one

Your migration investment ensures compatibility with future AgentFlow releases.

---

**Migration Status: ✅ Complete and Ready**

The new consolidated MCP API is production-ready and backward compatible. Upgrade at your own pace!
