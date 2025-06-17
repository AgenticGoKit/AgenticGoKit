# MCP API Quick Reference

**AgentFlow MCP Integration - Developer Cheat Sheet**

## 🚀 Quick Start (30 seconds)

```go
// 1. Initialize MCP
core.QuickStartMCP()

// 2. Create agent  
agent, _ := core.NewMCPAgent("my-agent", llmProvider)

// 3. Use agent
state := core.NewState()
state.Set("query", "search for AI news")
result, _ := agent.Run(ctx, state)

// 4. Cleanup
defer core.ShutdownMCP()
```

## 📋 API Patterns

### **Basic Pattern**
```go
core.InitializeMCP(config) → core.NewMCPAgent() → agent.Run()
```

### **Cache Pattern** 
```go
core.InitializeMCPWithCache(mcp, cache) → core.NewMCPAgentWithCache() → agent.Run()
```

### **Production Pattern**
```go
core.InitializeProductionMCP(ctx, prod) → core.NewProductionMCPAgent() → agent.Run()
```

## 🔧 Core Functions

| Function | Purpose | Returns |
|----------|---------|---------|
| `QuickStartMCP()` | Initialize with defaults | `error` |
| `InitializeMCP(config)` | Basic initialization | `error` |
| `InitializeMCPWithCache(mcp, cache)` | Init with caching | `error` |
| `InitializeProductionMCP(ctx, prod)` | Full production stack | `error` |
| `NewMCPAgent(name, llm)` | Basic agent | `*MCPAwareAgent, error` |
| `NewMCPAgentWithCache(name, llm)` | Cached agent | `*MCPAwareAgent, error` |
| `NewProductionMCPAgent(name, llm, config)` | Production agent | `*MCPAwareAgent, error` |
| `ExecuteMCPTool(ctx, tool, args)` | Direct tool execution | `MCPToolResult, error` |
| `GetMCPManager()` | Global manager | `MCPManager` |
| `ShutdownMCP()` | Clean shutdown | `error` |

## ⚙️ Configuration Quick Setup

### Basic Config
```go
config := core.DefaultMCPConfig()
config.Servers = []core.MCPServerConfig{{
    Name: "web-tools", Type: "tcp", 
    Host: "localhost", Port: 8811,
}}
```

### Cache Config  
```go
cache := core.DefaultMCPCacheConfig()
cache.Enabled = true
cache.DefaultTTL = 10 * time.Minute
cache.ToolTTLs = map[string]time.Duration{
    "search": 30 * time.Minute,
}
```

### Production Config
```go
prod := core.DefaultProductionConfig()
prod.ConnectionPool.MaxConnections = 50
prod.Metrics.Enabled = true
prod.LoadBalancing.Enabled = true
```

## 🔍 Server Types

| Type | Usage | Example |
|------|-------|---------|
| `tcp` | Network server | `{Type: "tcp", Host: "localhost", Port: 8811}` |
| `stdio` | Local process | `{Type: "stdio", Command: "mcp-server", Args: ["--config", "..."]}` |
| `websocket` | WebSocket server | `{Type: "websocket", Host: "ws.example.com", Port: 443}` |

## 📊 Monitoring

```go
// Health check
manager := core.GetMCPManager()
health := manager.HealthCheck(ctx)

// Metrics
metrics := manager.GetMetrics()
fmt.Printf("Tools: %d, Executions: %d\n", 
    metrics.TotalTools, metrics.ToolExecutions)

// Cache stats
cache := core.GetMCPCacheManager()
stats, _ := cache.GetGlobalStats(ctx)
fmt.Printf("Hit rate: %.1f%%\n", stats.HitRate*100)
```

## 🐛 Common Issues & Fixes

| Error | Solution |
|-------|----------|
| "MCP manager not initialized" | Call `InitializeMCP()` first |
| "No tools available" | Check server connectivity & config |
| Tool timeout | Increase `RequestTimeout` in config |
| Cache miss rate high | Tune `ToolTTLs` in cache config |

## 📁 File Structure

```
core/
├── mcp.go          # 🎯 Main API (all interfaces, factories, config)
└── mcp_agent.go    # 🤖 Agent implementation only
```

## 🏗️ Agent Configuration

```go
agentConfig := core.DefaultMCPAgentConfig()
agentConfig.MaxToolsPerExecution = 5
agentConfig.ParallelExecution = true  
agentConfig.ExecutionTimeout = 2 * time.Minute
agentConfig.RetryFailedTools = true
agentConfig.EnableCaching = true
```

## 🔄 Lifecycle Management  

```go
// Initialize
core.InitializeMCP(config)

// Create agents
agent, _ := core.NewMCPAgent("agent", llm)

// Runtime access
manager := core.GetMCPManager()
tools := manager.GetAvailableTools()

// Shutdown
defer core.ShutdownMCP()
```

## 💡 Best Practices

✅ **DO:**
- Always call `ShutdownMCP()` for cleanup
- Use caching for repeated operations
- Enable metrics in production
- Handle initialization errors

❌ **DON'T:**
- Create agents before initialization
- Ignore health check results
- Set very short cache TTLs
- Skip error handling

---

**Need more details?** See [MCP_API_Usage_Guide.md](MCP_API_Usage_Guide.md)
