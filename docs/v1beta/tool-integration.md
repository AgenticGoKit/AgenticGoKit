# Tool Integration

Learn how to add tool capabilities to your agents, enabling them to interact with external systems, APIs, and services through the Model Context Protocol (MCP).

---

## 🎯 Overview

AgenticGoKit v1beta provides comprehensive tool integration through:

- **MCP Protocol** - Standard protocol for tool communication
- **Multiple Transports** - TCP, stdio, WebSocket, HTTP SSE, HTTP streaming
- **Tool Discovery** - Automatic discovery of available tools
- **Caching** - Intelligent result caching with TTL
- **Circuit Breaker** - Fault tolerance for external services
- **Metrics** - Comprehensive monitoring and observability

---

## 🚀 Quick Start

### Basic MCP Server Configuration

```go
package main

import (
    "context"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Define MCP servers
    mcpServers := []v1beta.MCPServer{
        {
            Name:    "filesystem",
            Type:    "stdio",
            Command: "mcp-server-filesystem",
            Enabled: true,
        },
        {
            Name:    "web-api",
            Type:    "http_sse",
            Address: "localhost",
            Port:    8080,
            Enabled: true,
        },
    }
    
    // Create agent with MCP tools
    agent, err := v1beta.NewBuilder("ToolAgent").
        WithPreset(v1beta.ChatAgent).
        WithLLM("openai", "gpt-4").
        WithTools(
            v1beta.WithMCP(mcpServers...),
        ).
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    // Agent can now use MCP tools
    result, _ := agent.Run(context.Background(), "List files in current directory")
}
```

---

## 🔧 MCP Server Types

### 1. Stdio Transport

Execute MCP servers as child processes:

```go
mcpServer := v1beta.MCPServer{
    Name:    "filesystem",
    Type:    "stdio",
    Command: "mcp-server-filesystem",
    Enabled: true,
}

agent, _ := v1beta.NewBuilder("FSAgent").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCP(mcpServer),
    ).
    Build()
```

**Use Cases:**
- Local file operations
- Command-line tools
- Sandboxed environments

### 2. TCP Transport

Connect to MCP servers over TCP:

```go
mcpServer := v1beta.MCPServer{
    Name:    "database",
    Type:    "tcp",
    Address: "localhost",
    Port:    9090,
    Enabled: true,
}

agent, _ := v1beta.NewBuilder("DBAgent").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCP(mcpServer),
    ).
    Build()
```

**Use Cases:**
- Remote services
- Microservices
- Database connections

### 3. WebSocket Transport

Bi-directional communication with MCP servers:

```go
mcpServer := v1beta.MCPServer{
    Name:    "realtime-data",
    Type:    "websocket",
    Address: "ws://localhost:8081/mcp",
    Enabled: true,
}

agent, _ := v1beta.NewBuilder("RealtimeAgent").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCP(mcpServer),
    ).
    Build()
```

**Use Cases:**
- Real-time updates
- Streaming data
- Interactive services

### 4. HTTP SSE Transport

Server-Sent Events for streaming:

```go
mcpServer := v1beta.MCPServer{
    Name:    "web-api",
    Type:    "http_sse",
    Address: "localhost",
    Port:    8080,
    Enabled: true,
}

agent, _ := v1beta.NewBuilder("WebAgent").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCP(mcpServer),
    ).
    Build()
```

**Use Cases:**
- Web APIs
- Event streams
- HTTP-based services

### 5. HTTP Streaming Transport

Chunked transfer encoding for large responses:

```go
mcpServer := v1beta.MCPServer{
    Name:    "data-processor",
    Type:    "http_streaming",
    Address: "api.example.com",
    Port:    443,
    Enabled: true,
}

agent, _ := v1beta.NewBuilder("DataAgent").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCP(mcpServer),
    ).
    Build()
```

**Use Cases:**
- Large data transfers
- Progressive processing
- Media streaming

---

## 🔍 Tool Discovery

### Automatic Discovery

Automatically discover MCP servers on the network:

```go
agent, _ := v1beta.NewBuilder("DiscoveryAgent").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCPDiscovery(8080, 8081, 8090, 8100), // Scan these ports
    ).
    Build()
```

**Configuration:**

```go
// Custom discovery settings
agent, _ := v1beta.NewBuilder("CustomDiscovery").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCPDiscovery(), // Uses default ports
        v1beta.WithToolTimeout(30 * time.Second),
    ).
    Build()
```

### Manual Tool Registration

Register tools programmatically:

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

// Get tool manager from capabilities
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // List available tools
    tools := capabilities.Tools.List()
    for _, tool := range tools {
        fmt.Printf("Tool: %s - %s\n", tool.Name, tool.Description)
    }
    
    // Check if specific tool is available
    if capabilities.Tools.IsAvailable("web_search") {
        // Use the tool
        result, _ := capabilities.Tools.Execute(ctx, "web_search", map[string]interface{}{
            "query": "latest news",
        })
        return fmt.Sprintf("%v", result.Content), nil
    }
    
    return capabilities.LLM("You are a helpful assistant.", input)
}
```

---

## ⚙️ Tool Configuration

### Basic Configuration

```go
agent, _ := v1beta.NewBuilder("ConfiguredAgent").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCP(servers...),
        v1beta.WithToolTimeout(30 * time.Second),
        v1beta.WithMaxConcurrentTools(5),
    ).
    Build()
```

### Advanced Configuration via TOML

```toml
[tools]
enabled = true
max_retries = 3
timeout = "30s"
rate_limit = 100  # requests per second
max_concurrent = 10

[tools.mcp]
enabled = true
discovery = true
connection_timeout = "30s"
max_retries = 3
retry_delay = "1s"
discovery_timeout = "10s"
scan_ports = [8080, 8081, 8090, 8100]

[[tools.mcp.servers]]
name = "filesystem"
type = "stdio"
command = "mcp-server-filesystem"
enabled = true

[[tools.mcp.servers]]
name = "web-api"
type = "http_sse"
address = "localhost"
port = 8080
enabled = true

[tools.cache]
enabled = true
ttl = "15m"
max_size = 100  # MB
max_keys = 10000
eviction_policy = "lru"
cleanup_interval = "5m"
backend = "memory"

[tools.cache.tool_ttls]
web_search = "5m"
content_fetch = "30m"
summarize_text = "60m"

[tools.circuit_breaker]
enabled = true
failure_threshold = 5
success_threshold = 2
timeout = "60s"
half_open_max_calls = 3
```

---

## 🎯 Using Tools in Handlers

### Execute Tool Directly

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Execute a specific tool
    result, err := capabilities.Tools.Execute(ctx, "web_search", map[string]interface{}{
        "query":      input,
        "max_results": 5,
    })
    if err != nil {
        return "", fmt.Errorf("tool execution failed: %w", err)
    }
    
    if !result.Success {
        return "", fmt.Errorf("tool error: %s", result.Error)
    }
    
    // Use tool result
    return fmt.Sprintf("Search results: %v", result.Content), nil
}
```

### Tool Discovery Pattern

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // List all available tools
    tools := capabilities.Tools.List()
    
    // Find appropriate tool
    var selectedTool *v1beta.ToolInfo
    for _, tool := range tools {
        if strings.Contains(tool.Name, "search") {
            selectedTool = &tool
            break
        }
    }
    
    if selectedTool == nil {
        return "No search tool available", nil
    }
    
    // Execute the found tool
    result, err := capabilities.Tools.Execute(ctx, selectedTool.Name, map[string]interface{}{
        "query": input,
    })
    if err != nil {
        return "", err
    }
    
    return fmt.Sprintf("%v", result.Content), nil
}
```

### LLM-Directed Tool Selection

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Get available tools
    tools := capabilities.Tools.List()
    toolList := make([]string, len(tools))
    for i, tool := range tools {
        toolList[i] = fmt.Sprintf("%s: %s", tool.Name, tool.Description)
    }
    
    // Ask LLM which tool to use
    prompt := fmt.Sprintf("Available tools:\n%s\n\nWhich tool should be used for: %s\nReply with just the tool name.",
        strings.Join(toolList, "\n"), input)
    
    toolName, err := capabilities.LLM("You are a tool selection expert.", prompt)
    if err != nil {
        return "", err
    }
    
    // Execute selected tool
    result, err := capabilities.Tools.Execute(ctx, strings.TrimSpace(toolName), map[string]interface{}{
        "query": input,
    })
    if err != nil {
        return "", err
    }
    
    return fmt.Sprintf("%v", result.Content), nil
}
```

---

## 💾 Caching

### Enable Caching

```go
agent, _ := v1beta.NewBuilder("CachedAgent").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCP(servers...),
        v1beta.WithToolCaching(15 * time.Minute), // 15 minute TTL
    ).
    Build()
```

### Per-Tool TTL Configuration

```toml
[tools.cache.tool_ttls]
web_search = "5m"           # Short TTL for dynamic data
content_fetch = "30m"       # Medium TTL for semi-static content
database_query = "1m"       # Very short for real-time data
static_api = "24h"          # Long TTL for static APIs
```

### Cache Backend Options

```toml
[tools.cache]
backend = "memory"  # Options: memory, redis, file

# Redis backend configuration
[tools.cache.backend_config]
redis_url = "redis://localhost:6379"
redis_db = "0"
redis_password = "${REDIS_PASSWORD}"
```

---

## 🔒 Circuit Breaker

Protect against cascading failures:

```toml
[tools.circuit_breaker]
enabled = true
failure_threshold = 5        # Open after 5 consecutive failures
success_threshold = 2        # Close after 2 consecutive successes
timeout = "60s"              # Time before attempting to half-open
half_open_max_calls = 3      # Max calls in half-open state
```

**States:**
- **Closed**: Normal operation, all requests go through
- **Open**: Circuit is tripped, requests fail immediately
- **Half-Open**: Testing if service has recovered

---

## 📊 Monitoring and Metrics

### Get Tool Metrics

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Get metrics
    metrics := capabilities.Tools.GetMetrics()
    
    fmt.Printf("Total Executions: %d\n", metrics.TotalExecutions)
    fmt.Printf("Success Rate: %.2f%%\n", 
        float64(metrics.SuccessfulCalls)/float64(metrics.TotalExecutions)*100)
    fmt.Printf("Average Latency: %v\n", metrics.AverageLatency)
    fmt.Printf("Cache Hit Rate: %.2f%%\n", metrics.CacheHitRate*100)
    
    // Per-tool metrics
    for toolName, toolMetrics := range metrics.ToolMetrics {
        fmt.Printf("\nTool: %s\n", toolName)
        fmt.Printf("  Executions: %d\n", toolMetrics.Executions)
        fmt.Printf("  Success Rate: %.2f%%\n", toolMetrics.SuccessRate*100)
        fmt.Printf("  Avg Latency: %v\n", toolMetrics.AverageLatency)
    }
    
    return capabilities.LLM("You are a helpful assistant.", input)
}
```

### Health Checks

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Check health of all MCP servers
    healthStatus := capabilities.Tools.HealthCheck(ctx)
    
    for serverName, status := range healthStatus {
        fmt.Printf("Server: %s\n", serverName)
        fmt.Printf("  Status: %s\n", status.Status)
        fmt.Printf("  Response Time: %v\n", status.ResponseTime)
        fmt.Printf("  Tool Count: %d\n", status.ToolCount)
        if status.Error != "" {
            fmt.Printf("  Error: %s\n", status.Error)
        }
    }
    
    return capabilities.LLM("You are a helpful assistant.", input)
}
```

---

## 🎨 Common Patterns

### Pattern 1: Tool-First Agent

Prefer tools over LLM for specific tasks:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Try to use appropriate tool first
    if strings.Contains(strings.ToLower(input), "search") {
        result, err := capabilities.Tools.Execute(ctx, "web_search", map[string]interface{}{
            "query": input,
        })
        if err == nil && result.Success {
            return fmt.Sprintf("%v", result.Content), nil
        }
    }
    
    // Fall back to LLM
    return capabilities.LLM("You are a helpful assistant.", input)
}
```

### Pattern 2: Multi-Tool Pipeline

Chain multiple tools:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Step 1: Search for information
    searchResult, err := capabilities.Tools.Execute(ctx, "web_search", map[string]interface{}{
        "query": input,
    })
    if err != nil {
        return "", err
    }
    
    // Step 2: Fetch content
    contentResult, err := capabilities.Tools.Execute(ctx, "content_fetch", map[string]interface{}{
        "url": searchResult.Content,
    })
    if err != nil {
        return "", err
    }
    
    // Step 3: Summarize with LLM
    return capabilities.LLM(
        "Summarize this content concisely.",
        fmt.Sprintf("%v", contentResult.Content),
    )
}
```

### Pattern 3: Parallel Tool Execution

Execute multiple tools concurrently:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    type result struct {
        name string
        data interface{}
        err  error
    }
    
    resultChan := make(chan result, 3)
    
    // Execute tools in parallel
    go func() {
        res, err := capabilities.Tools.Execute(ctx, "web_search", map[string]interface{}{"query": input})
        resultChan <- result{"search", res.Content, err}
    }()
    
    go func() {
        res, err := capabilities.Tools.Execute(ctx, "news_api", map[string]interface{}{"topic": input})
        resultChan <- result{"news", res.Content, err}
    }()
    
    go func() {
        res, err := capabilities.Tools.Execute(ctx, "social_media", map[string]interface{}{"hashtag": input})
        resultChan <- result{"social", res.Content, err}
    }()
    
    // Collect results
    var results []string
    for i := 0; i < 3; i++ {
        r := <-resultChan
        if r.err == nil {
            results = append(results, fmt.Sprintf("%s: %v", r.name, r.data))
        }
    }
    
    // Synthesize with LLM
    return capabilities.LLM(
        "Synthesize these results into a coherent response.",
        strings.Join(results, "\n\n"),
    )
}
```

### Pattern 4: Conditional Tool Usage

Use tools based on conditions:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Check if input requires external data
    needsData, err := capabilities.LLM(
        "Does this query require external data? Reply YES or NO only.",
        input,
    )
    if err != nil {
        return "", err
    }
    
    if strings.TrimSpace(needsData) == "YES" {
        // Use tools to fetch data
        result, err := capabilities.Tools.Execute(ctx, "web_search", map[string]interface{}{
            "query": input,
        })
        if err != nil {
            return "", err
        }
        
        // Use LLM with tool data
        return capabilities.LLM(
            fmt.Sprintf("Use this data to answer: %v", result.Content),
            input,
        )
    }
    
    // Use LLM directly
    return capabilities.LLM("You are a helpful assistant.", input)
}
```

---

## 🎯 Best Practices

### 1. Error Handling

Always handle tool errors gracefully:

```go
result, err := capabilities.Tools.Execute(ctx, "web_search", args)
if err != nil {
    // Log error
    log.Printf("Tool execution failed: %v", err)
    
    // Provide fallback
    return capabilities.LLM("Answer without external data.", input)
}

if !result.Success {
    // Handle tool-specific errors
    log.Printf("Tool returned error: %s", result.Error)
    return "Tool encountered an issue. Please try again.", nil
}
```

### 2. Timeout Configuration

Set appropriate timeouts:

```go
// Quick operations
v1beta.WithToolTimeout(5 * time.Second)

// Standard operations
v1beta.WithToolTimeout(30 * time.Second)

// Long-running operations
v1beta.WithToolTimeout(2 * time.Minute)
```

### 3. Caching Strategy

Choose TTL based on data freshness requirements:

```toml
[tools.cache.tool_ttls]
stock_price = "30s"        # Real-time data
weather = "15m"            # Changes frequently
wikipedia = "24h"          # Relatively static
company_info = "7d"        # Rarely changes
```

### 4. Rate Limiting

Respect API limits:

```toml
[tools]
rate_limit = 10  # 10 requests per second
max_concurrent = 3  # Max 3 parallel executions
```

### 5. Monitoring

Regularly check tool health:

```go
// Periodic health check
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        health := capabilities.Tools.HealthCheck(ctx)
        for name, status := range health {
            if status.Status != "healthy" {
                log.Printf("Unhealthy tool server: %s - %s", name, status.Error)
            }
        }
    }
}()
```

---

## 🐛 Troubleshooting

### Issue: MCP Server Not Connecting

**Cause**: Server not running or wrong configuration

**Solution**: Verify server configuration
```bash
# Test server manually
curl http://localhost:8080/health

# Check logs
cat /var/log/mcp-server.log
```

### Issue: Tools Not Discovered

**Cause**: Discovery disabled or ports not scanned

**Solution**: Enable discovery explicitly
```go
agent, _ := v1beta.NewBuilder("Agent").
    WithTools(
        v1beta.WithMCPDiscovery(8080, 8081, 8090),
    ).
    Build()
```

### Issue: Tool Execution Timeout

**Cause**: Operation taking too long

**Solution**: Increase timeout
```go
agent, _ := v1beta.NewBuilder("Agent").
    WithTools(
        v1beta.WithMCP(servers...),
        v1beta.WithToolTimeout(60 * time.Second),
    ).
    Build()
```

---

## 📚 Next Steps

- **[Custom Handlers](./custom-handlers.md)** - Advanced tool integration patterns
- **[Error Handling](./error-handling.md)** - Robust error management
- **[Performance](./performance.md)** - Tool optimization strategies
- **[Examples](./examples/)** - Complete tool integration examples

---

**Ready for error handling?** Continue to [Error Handling](./error-handling.md) →
