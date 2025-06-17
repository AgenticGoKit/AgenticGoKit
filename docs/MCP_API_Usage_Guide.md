# MCP API Usage Guide

**AgentFlow MCP Integration - Public API Reference**  
**Version**: 1.0  
**File**: `core/mcp.go` + `core/mcp_agent.go`  
**Status**: Production Ready  

This guide shows you how to use AgentFlow's consolidated MCP (Model Context Protocol) API for building intelligent agents that can dynamically discover and use external tools.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Progressive Usage Patterns](#progressive-usage-patterns)
3. [Basic MCP Usage](#basic-mcp-usage)
4. [MCP with Caching](#mcp-with-caching)
5. [Production MCP](#production-mcp)
6. [Configuration Reference](#configuration-reference)
7. [Complete Examples](#complete-examples)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

---

## Quick Start

Get started with MCP in just a few lines:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // 1. Initialize MCP with defaults
    if err := core.QuickStartMCP(); err != nil {
        log.Fatal(err)
    }

    // 2. Create your LLM provider (example with a mock)
    llmProvider := &MyLLMProvider{} // Your LLM implementation
    
    // 3. Create an MCP-aware agent
    agent, err := core.NewMCPAgent("my-agent", llmProvider)
    if err != nil {
        log.Fatal(err)
    }

    // 4. Use the agent
    ctx := context.Background()
    state := core.NewState()
    state.Set("query", "search for AI news")
    
    result, err := agent.Run(ctx, state)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Agent result: %+v\n", result)
    
    // 5. Clean shutdown
    defer core.ShutdownMCP()
}
```

---

## Progressive Usage Patterns

AgentFlow's MCP API follows a progressive complexity model:

### **Level 1: Basic MCP** 
*Simple tool usage with minimal configuration*

### **Level 2: MCP + Caching**
*Enhanced performance with intelligent result caching*

### **Level 3: Production MCP**
*Enterprise-grade: connection pooling, retry logic, metrics, load balancing*

---

## Basic MCP Usage

### Initialize MCP

```go
// Option 1: Quick start with defaults
err := core.QuickStartMCP()

// Option 2: Custom configuration
config := core.DefaultMCPConfig()
config.EnableDiscovery = true
config.ConnectionTimeout = 30 * time.Second
config.Servers = []core.MCPServerConfig{
    {
        Name: "web-tools",
        Type: "tcp", 
        Host: "localhost",
        Port: 8811,
        Enabled: true,
    },
}
err := core.InitializeMCP(config)
```

### Create MCP-Aware Agents

```go
// Basic agent
agent, err := core.NewMCPAgent("research-agent", llmProvider)

// Agent with full configuration
mcpConfig := core.DefaultMCPConfig()
agentConfig := core.DefaultMCPAgentConfig()
agentConfig.MaxToolsPerExecution = 3
agentConfig.ParallelExecution = true

agent, err := core.CreateMCPAgentWithLLMAndTools(
    ctx, "advanced-agent", llmProvider, mcpConfig, agentConfig)
```

### Execute Tools Directly

```go
// Execute a single tool
result, err := core.ExecuteMCPTool(ctx, "web_search", map[string]interface{}{
    "query": "latest AI developments",
    "limit": 10,
})

// Connect to a specific server
err := core.ConnectMCPServer("my-server", "tcp", "localhost:8811")
```

### Global Access

```go
// Access global instances
manager := core.GetMCPManager()
tools := manager.GetAvailableTools()

// Manual tool registration
err := core.RegisterMCPToolsWithRegistry(ctx)
```

---

## MCP with Caching

Enhanced performance through intelligent result caching:

### Initialize with Cache

```go
// Initialize MCP + Caching
mcpConfig := core.DefaultMCPConfig()
cacheConfig := core.DefaultMCPCacheConfig()
cacheConfig.Enabled = true
cacheConfig.DefaultTTL = 10 * time.Minute

err := core.InitializeMCPWithCache(mcpConfig, cacheConfig)
```

### Cache Configuration

```go
cacheConfig := core.MCPCacheConfig{
    Enabled:    true,
    DefaultTTL: 5 * time.Minute,
    MaxSize:    1000,
    EvictionPolicy: "lru",
    
    // Per-tool TTL overrides
    ToolTTLs: map[string]time.Duration{
        "web_search":    30 * time.Minute, // Search results cached longer
        "weather":       5 * time.Minute,  // Weather data cached briefly
        "stock_price":   1 * time.Minute,  // Financial data cached very briefly
    },
    
    // Redis configuration (optional)
    Redis: &core.RedisConfig{
        Enabled:  false, // Use in-memory by default
        Host:     "localhost",
        Port:     6379,
        DB:       0,
        Password: "",
    },
}
```

### Create Cache-Enabled Agents

```go
// Agent with caching
agent, err := core.NewMCPAgentWithCache("cached-agent", llmProvider)

// Access cache manager
cacheManager := core.GetMCPCacheManager()
stats, err := cacheManager.GetGlobalStats(ctx)
fmt.Printf("Cache hit rate: %.2f%%\n", stats.HitRate*100)

// Invalidate cache by pattern
err = cacheManager.InvalidateByPattern(ctx, "web_search")
```

---

## Production MCP

Enterprise-grade MCP with all advanced features:

### Initialize Production Stack

```go
prodConfig := core.DefaultProductionConfig()
prodConfig.Servers = []core.MCPServerConfig{
    {Name: "web-tools", Type: "tcp", Host: "web-tools.company.com", Port: 8811},
    {Name: "data-tools", Type: "tcp", Host: "data-tools.company.com", Port: 8812},
}

// Connection pooling
prodConfig.ConnectionPool.MaxConnections = 50
prodConfig.ConnectionPool.MaxIdleTime = 10 * time.Minute

// Retry logic
prodConfig.Retry.MaxRetries = 5
prodConfig.Retry.BackoffBase = 1 * time.Second
prodConfig.Retry.BackoffMax = 30 * time.Second

// Circuit breaker
prodConfig.CircuitBreaker.Enabled = true
prodConfig.CircuitBreaker.FailureThreshold = 10
prodConfig.CircuitBreaker.RecoveryTimeout = 60 * time.Second

// Load balancing
prodConfig.LoadBalancing.Enabled = true
prodConfig.LoadBalancing.Strategy = "round_robin"

// Metrics
prodConfig.Metrics.Enabled = true
prodConfig.Metrics.Port = 8080
prodConfig.Metrics.Path = "/metrics"

// Initialize
err := core.InitializeProductionMCP(ctx, prodConfig)
```

### Create Production Agents

```go
agent, err := core.NewProductionMCPAgent("production-agent", llmProvider, prodConfig)

// Production agent has all features:
// - Connection pooling
// - Automatic retries with exponential backoff
// - Circuit breaker protection
// - Load balancing across servers
// - Comprehensive metrics
// - Health monitoring
```

### Monitor Production Health

```go
manager := core.GetMCPManager()

// Health checks
healthStatus := manager.HealthCheck(ctx)
for server, status := range healthStatus {
    if status.Healthy {
        fmt.Printf("✅ %s: %s\n", server, status.Message)
    } else {
        fmt.Printf("❌ %s: %s\n", server, status.Message)
    }
}

// Metrics
metrics := manager.GetMetrics()
fmt.Printf("Connected servers: %d\n", metrics.ConnectedServers)
fmt.Printf("Total tools: %d\n", metrics.TotalTools)
fmt.Printf("Tool executions: %d\n", metrics.ToolExecutions)
fmt.Printf("Average latency: %v\n", metrics.AverageLatency)
```

---

## Configuration Reference

### MCPConfig (Basic Configuration)

```go
type MCPConfig struct {
    EnableDiscovery   bool                `toml:"enable_discovery"`
    DiscoveryTimeout  time.Duration       `toml:"discovery_timeout"`
    ConnectionTimeout time.Duration       `toml:"connection_timeout"`
    RequestTimeout    time.Duration       `toml:"request_timeout"`
    MaxRetries        int                 `toml:"max_retries"`
    RetryDelay        time.Duration       `toml:"retry_delay"`
    MaxConnections    int                 `toml:"max_connections"`
    CacheTimeout      time.Duration       `toml:"cache_timeout"`
    EnableMetrics     bool                `toml:"enable_metrics"`
    MetricsPort       int                 `toml:"metrics_port"`
    Servers           []MCPServerConfig   `toml:"servers"`
}
```

### MCPServerConfig (Server Configuration)

```go
type MCPServerConfig struct {
    Name       string            `toml:"name"`
    Type       string            `toml:"type"`        // "tcp", "stdio", "websocket"
    Host       string            `toml:"host"`
    Port       int               `toml:"port"`
    Command    string            `toml:"command"`     // For STDIO servers
    Args       []string          `toml:"args"`
    Env        map[string]string `toml:"env"`
    Enabled    bool              `toml:"enabled"`
    Priority   int               `toml:"priority"`
    HealthCheck *HealthCheckConfig `toml:"health_check"`
}
```

### MCPAgentConfig (Agent Configuration)

```go
type MCPAgentConfig struct {
    MaxToolsPerExecution int           `toml:"max_tools_per_execution"`
    ToolSelectionTimeout time.Duration `toml:"tool_selection_timeout"`
    ParallelExecution    bool          `toml:"parallel_execution"`
    ExecutionTimeout     time.Duration `toml:"execution_timeout"`
    RetryFailedTools     bool          `toml:"retry_failed_tools"`
    MaxRetries          int           `toml:"max_retries"`
    UseToolDescriptions bool          `toml:"use_tool_descriptions"`
    ToolSelectionPrompt string        `toml:"tool_selection_prompt"`
    ResultInterpretation bool          `toml:"result_interpretation"`
    EnableCaching       bool          `toml:"enable_caching"`
    CacheConfig         MCPCacheConfig `toml:"cache"`
}
```

---

## Complete Examples

### Example 1: Web Research Agent

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Configure MCP with web tools
    config := core.DefaultMCPConfig()
    config.Servers = []core.MCPServerConfig{
        {
            Name:    "web-search",
            Type:    "tcp",
            Host:    "localhost",
            Port:    8811,
            Enabled: true,
        },
        {
            Name:    "web-scraper",
            Type:    "tcp", 
            Host:    "localhost",
            Port:    8812,
            Enabled: true,
        },
    }

    // Initialize with caching for better performance
    cacheConfig := core.DefaultMCPCacheConfig()
    cacheConfig.Enabled = true
    cacheConfig.ToolTTLs = map[string]time.Duration{
        "search":      30 * time.Minute,
        "fetch_page":  10 * time.Minute,
        "summarize":   60 * time.Minute,
    }

    if err := core.InitializeMCPWithCache(config, cacheConfig); err != nil {
        log.Fatal(err)
    }
    defer core.ShutdownMCP()

    // Create LLM provider
    llmProvider := &OllamaProvider{Model: "llama3.1"}

    // Create research agent
    agentConfig := core.DefaultMCPAgentConfig()
    agentConfig.MaxToolsPerExecution = 5
    agentConfig.ParallelExecution = false // Sequential for research workflow
    agentConfig.UseToolDescriptions = true

    agent, err := core.CreateMCPAgentWithLLMAndTools(
        context.Background(), "research-agent", llmProvider, config, agentConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Research workflow
    queries := []string{
        "What are the latest developments in AI agents?",
        "Find information about MCP (Model Context Protocol)",
        "Research best practices for LLM agent architectures",
    }

    for _, query := range queries {
        fmt.Printf("\n🔍 Researching: %s\n", query)
        
        state := core.NewState()
        state.Set("query", query)
        state.Set("max_results", 5)
        state.Set("include_summaries", true)

        result, err := agent.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Research failed: %v\n", err)
            continue
        }

        // Extract research results
        if findings := result.Get("findings"); findings != nil {
            fmt.Printf("✅ Found: %v\n", findings)
        }
        if summary := result.Get("summary"); summary != nil {
            fmt.Printf("📋 Summary: %v\n", summary)
        }
    }

    // Show cache performance
    cacheManager := core.GetMCPCacheManager()
    if stats, err := cacheManager.GetGlobalStats(context.Background()); err == nil {
        fmt.Printf("\n📊 Cache Performance:\n")
        fmt.Printf("   Hit Rate: %.1f%%\n", stats.HitRate*100)
        fmt.Printf("   Total Hits: %d\n", stats.HitCount)
        fmt.Printf("   Cache Size: %d MB\n", stats.TotalSize/(1024*1024))
    }
}
```

### Example 2: Multi-Agent Production System

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/kunalkushwaha/agentflow/core"
)

func main() {
    // Production configuration
    prodConfig := core.DefaultProductionConfig()
    prodConfig.Servers = []core.MCPServerConfig{
        {Name: "web-tools", Type: "tcp", Host: "web-tools.internal", Port: 8811},
        {Name: "data-tools", Type: "tcp", Host: "data-tools.internal", Port: 8812},
        {Name: "ai-tools", Type: "tcp", Host: "ai-tools.internal", Port: 8813},
    }
    
    // High-performance settings
    prodConfig.ConnectionPool.MaxConnections = 100
    prodConfig.LoadBalancing.Enabled = true
    prodConfig.LoadBalancing.Strategy = "round_robin"
    prodConfig.Metrics.Enabled = true
    prodConfig.Metrics.Port = 9090

    // Initialize production MCP
    ctx := context.Background()
    if err := core.InitializeProductionMCP(ctx, prodConfig); err != nil {
        log.Fatal(err)
    }
    defer core.ShutdownMCP()

    // Create multiple specialized agents
    llmProvider := &OpenAIProvider{Model: "gpt-4"}

    agents := map[string]*core.MCPAwareAgent{}
    
    // Research agent
    if agent, err := core.NewProductionMCPAgent("researcher", llmProvider, prodConfig); err == nil {
        agents["researcher"] = agent
    }
    
    // Analysis agent  
    if agent, err := core.NewProductionMCPAgent("analyst", llmProvider, prodConfig); err == nil {
        agents["analyst"] = agent
    }
    
    // Writer agent
    if agent, err := core.NewProductionMCPAgent("writer", llmProvider, prodConfig); err == nil {
        agents["writer"] = agent
    }

    // Production workflow: concurrent agent execution
    tasks := []struct {
        agentName string
        query     string
    }{
        {"researcher", "Find latest market trends in AI"},
        {"analyst", "Analyze competitor pricing strategies"},  
        {"writer", "Generate product launch announcement"},
    }

    var wg sync.WaitGroup
    results := make(chan string, len(tasks))

    for _, task := range tasks {
        wg.Add(1)
        go func(agentName, query string) {
            defer wg.Done()
            
            agent := agents[agentName]
            state := core.NewState()
            state.Set("query", query)
            state.Set("priority", "high")
            
            start := time.Now()
            result, err := agent.Run(ctx, state)
            duration := time.Since(start)
            
            if err != nil {
                results <- fmt.Sprintf("❌ %s failed: %v", agentName, err)
            } else {
                results <- fmt.Sprintf("✅ %s completed in %v: %v", 
                    agentName, duration, result.Get("output"))
            }
        }(task.agentName, task.query)
    }

    // Wait for all agents and collect results
    go func() {
        wg.Wait()
        close(results)
    }()

    fmt.Println("🚀 Production agents running...")
    for result := range results {
        fmt.Println(result)
    }

    // Monitor production health
    manager := core.GetMCPManager()
    if healthStatus := manager.HealthCheck(ctx); len(healthStatus) > 0 {
        fmt.Println("\n🏥 System Health:")
        for server, status := range healthStatus {
            symbol := "✅"
            if !status.Healthy {
                symbol = "❌"
            }
            fmt.Printf("   %s %s: %s\n", symbol, server, status.Message)
        }
    }

    // Production metrics
    if metrics := manager.GetMetrics(); metrics.ConnectedServers > 0 {
        fmt.Printf("\n📊 Production Metrics:\n")
        fmt.Printf("   Connected Servers: %d\n", metrics.ConnectedServers)
        fmt.Printf("   Total Tools: %d\n", metrics.TotalTools)
        fmt.Printf("   Executions: %d\n", metrics.ToolExecutions)
        fmt.Printf("   Avg Latency: %v\n", metrics.AverageLatency)
        fmt.Printf("   Success Rate: %.1f%%\n", 
            float64(metrics.SuccessfulExecutions)/float64(metrics.ToolExecutions)*100)
    }
}
```

---

## Best Practices

### 1. Configuration Management

```go
// Load from configuration file
config, err := core.LoadMCPConfigFromTOML("agentflow.toml")

// Use environment-specific configs
var mcpConfig core.MCPConfig
switch os.Getenv("ENVIRONMENT") {
case "production":
    mcpConfig = core.DefaultProductionConfig().MCPConfig
case "staging":
    mcpConfig = core.DefaultMCPConfig()
    mcpConfig.ConnectionTimeout = 60 * time.Second
default:
    mcpConfig = core.DefaultMCPConfig()
}
```

### 2. Error Handling

```go
// Always handle initialization errors
if err := core.InitializeMCP(config); err != nil {
    log.Fatalf("Failed to initialize MCP: %v", err)
}

// Graceful shutdown
defer func() {
    if err := core.ShutdownMCP(); err != nil {
        log.Printf("Warning: MCP shutdown error: %v", err)
    }
}()

// Agent execution with retries
var result core.State
var err error
for retries := 0; retries < 3; retries++ {
    result, err = agent.Run(ctx, state)
    if err == nil {
        break
    }
    time.Sleep(time.Duration(retries+1) * time.Second)
}
```

### 3. Performance Optimization

```go
// Use caching for repeated operations
cacheConfig := core.DefaultMCPCacheConfig()
cacheConfig.ToolTTLs = map[string]time.Duration{
    "expensive_operation": 60 * time.Minute,
    "frequent_lookup":     5 * time.Minute,
}

// Enable parallel execution for independent tools
agentConfig := core.DefaultMCPAgentConfig()
agentConfig.ParallelExecution = true
agentConfig.MaxToolsPerExecution = 5

// Connection pooling for high throughput
prodConfig.ConnectionPool.MaxConnections = 50
prodConfig.ConnectionPool.KeepAlive = true
```

### 4. Monitoring and Observability

```go
// Enable metrics in production
prodConfig.Metrics.Enabled = true
prodConfig.Metrics.Port = 8080

// Regular health checks
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        manager := core.GetMCPManager()
        health := manager.HealthCheck(context.Background())
        for server, status := range health {
            if !status.Healthy {
                log.Printf("ALERT: Server %s unhealthy: %s", server, status.Message)
            }
        }
    }
}()

// Cache performance monitoring
if cacheManager := core.GetMCPCacheManager(); cacheManager != nil {
    stats, _ := cacheManager.GetGlobalStats(ctx)
    if stats.HitRate < 0.5 {
        log.Printf("WARNING: Low cache hit rate: %.2f%%", stats.HitRate*100)
    }
}
```

---

## Troubleshooting

### Common Issues

#### 1. "MCP manager not initialized"
```go
// Solution: Always initialize before creating agents
err := core.InitializeMCP(config)
if err != nil {
    log.Fatal(err)
}
agent, err := core.NewMCPAgent("my-agent", llmProvider)
```

#### 2. "No MCP tools available"
```go
// Solution: Check server configuration and connectivity
manager := core.GetMCPManager()
servers := manager.ListConnectedServers()
fmt.Printf("Connected servers: %v\n", servers)

tools := manager.GetAvailableTools()
fmt.Printf("Available tools: %d\n", len(tools))

// Test connectivity
health := manager.HealthCheck(context.Background())
for server, status := range health {
    fmt.Printf("%s: healthy=%v, message=%s\n", 
        server, status.Healthy, status.Message)
}
```

#### 3. "Tool execution timeout"
```go
// Solution: Increase timeouts
config := core.DefaultMCPConfig()
config.RequestTimeout = 60 * time.Second
config.ConnectionTimeout = 30 * time.Second

agentConfig := core.DefaultMCPAgentConfig()
agentConfig.ExecutionTimeout = 5 * time.Minute
```

#### 4. "Cache performance issues"
```go
// Solution: Monitor and tune cache settings
cacheManager := core.GetMCPCacheManager()
stats, _ := cacheManager.GetGlobalStats(ctx)

fmt.Printf("Hit Rate: %.2f%%\n", stats.HitRate*100)
fmt.Printf("Cache Size: %d MB\n", stats.TotalSize/(1024*1024))

// Clear cache if needed
err := cacheManager.InvalidateByPattern(ctx, "problematic_tool")
```

### Debug Mode

```go
// Enable detailed logging
import "github.com/rs/zerolog/log"

// Set debug level
zerolog.SetGlobalLevel(zerolog.DebugLevel)

// MCP operations will now show detailed logs
```

### Health Monitoring

```go
// Comprehensive health check
func healthCheck() {
    manager := core.GetMCPManager()
    
    // Server connectivity
    health := manager.HealthCheck(context.Background())
    allHealthy := true
    for server, status := range health {
        if !status.Healthy {
            fmt.Printf("❌ %s: %s\n", server, status.Message)
            allHealthy = false
        }
    }
    
    if allHealthy {
        fmt.Println("✅ All MCP servers healthy")
    }
    
    // Tool availability
    tools := manager.GetAvailableTools()
    fmt.Printf("📊 %d tools available\n", len(tools))
    
    // Metrics
    metrics := manager.GetMetrics()
    fmt.Printf("📈 Executions: %d, Success Rate: %.1f%%\n", 
        metrics.ToolExecutions,
        float64(metrics.SuccessfulExecutions)/float64(metrics.ToolExecutions)*100)
}
```

---

## Next Steps

1. **Explore Examples**: Check the `examples/` directory for more usage patterns
2. **Read Integration Guide**: See how MCP integrates with existing AgentFlow workflows  
3. **Production Checklist**: Review the production deployment guide
4. **Contributing**: Help improve the MCP integration

For more information:
- [AgentFlow Documentation](../README.md)
- [MCP Architecture Guide](Architecture.md)  
- [API Reference](../core/mcp.go)

---

**Happy Building! 🚀**
