# Tool Integration

Learn how to add tool capabilities to your v1beta agents using Model Context Protocol (MCP) servers, discovery, caching, and monitoring.

---

## Overview

- Configure tools through builder WithTools using ToolOption helpers
- Connect MCP servers explicitly or via discovery; supported transports: stdio, tcp, websocket, http_sse, http_streaming
- Control reliability with timeouts, concurrency caps, caching, and circuit breakers
- Observe tool health and metrics directly from capabilities.Tools

---

## Quick Start (MCP tools)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    servers := []v1beta.MCPServer{
        {Name: "filesystem", Type: "stdio", Command: "mcp-server-filesystem", Enabled: true},
        {Name: "web-api", Type: "http_sse", Address: "localhost", Port: 8080, Enabled: true},
    }

    agent, err := v1beta.NewBuilder("ToolAgent").
        WithPreset(v1beta.ChatAgent).
        WithTools(
            v1beta.WithMCP(servers...),
            v1beta.WithToolTimeout(30*time.Second),
        ).
        WithHandler(func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
            if caps.Tools == nil {
                return caps.LLM("You are a helpful assistant.", input)
            }
            res, err := caps.Tools.Execute(ctx, "filesystem.list", map[string]interface{}{"path": "."})
            if err != nil {
                return "", err
            }
            return caps.LLM("Summarize these files", fmt.Sprint(res.Content))
        }).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    _, _ = agent.Run(context.Background(), "List files in current directory")
}
```

---

## MCP Server Types

Define servers with v1beta.MCPServer and pass them to WithMCP.

- stdio: {Name: "filesystem", Type: "stdio", Command: "mcp-server-filesystem"}
- tcp: {Name: "database", Type: "tcp", Address: "localhost", Port: 9090}
- websocket: {Name: "realtime", Type: "websocket", Address: "ws://localhost:8081/mcp"}
- http_sse: {Name: "web-api", Type: "http_sse", Address: "localhost", Port: 8080}
- http_streaming: {Name: "data", Type: "http_streaming", Address: "api.example.com", Port: 443}

---

## Tool Discovery

Enable MCP discovery instead of listing servers manually.

```go
agent, _ := v1beta.NewBuilder("DiscoveryAgent").
    WithPreset(v1beta.ChatAgent).
    WithTools(
        v1beta.WithMCPDiscovery(8080, 8081, 8090),
        v1beta.WithToolTimeout(30*time.Second),
    ).
    Build()
```

If no ports are provided, defaults are used (8080, 8081, 8090, 8100, 3000, 3001).

---

## Internal Tools (non-MCP)

Register internal tools with RegisterInternalTool and execute via capabilities.Tools.

```go
// Define a tool
type searchTool struct{}
func (t *searchTool) Name() string        { return "web_search" }
func (t *searchTool) Description() string { return "Search the web" }
func (t *searchTool) Execute(ctx context.Context, args map[string]interface{}) (*v1beta.ToolResult, error) {
    return &v1beta.ToolResult{Success: true, Content: "results"}, nil
}

// Register once (init or main)
v1beta.RegisterInternalTool("web_search", func() v1beta.Tool { return &searchTool{} })

// Use in a handler
handler := func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    if caps.Tools == nil || !caps.Tools.IsAvailable("web_search") {
        return caps.LLM("You are a helpful assistant.", input)
    }
    res, err := caps.Tools.Execute(ctx, "web_search", map[string]interface{}{"query": input})
    if err != nil || !res.Success {
        return "", fmt.Errorf("tool failed: %v", err)
    }
    return caps.LLM("Summarize these results", fmt.Sprint(res.Content))
}
```

---

## Tool Configuration (options)

Pass ToolOption functions to WithTools:
- WithMCP(servers ...MCPServer) enables MCP with explicit servers.
- WithMCPDiscovery(scanPorts ...int) enables auto-discovery.
- WithToolTimeout(timeout time.Duration) sets per-call timeout.
- WithMaxConcurrentTools(max int) caps parallel executions.
- WithToolCaching(ttl time.Duration) enables result caching.

---

## Tool Configuration (TOML)

```toml
[tools]
enabled = true
max_retries = 3
timeout = "30s"
rate_limit = 100
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
max_size_mb = 100
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

### Using this TOML config

Load the file, then hand it to the builder with your handler:

```go
cfg, err := v1beta.LoadConfigFromFile("agent-tools.toml")
if err != nil {
    log.Fatal(err)
}

handler := func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    res, err := caps.Tools.Execute(ctx, "web_search", map[string]interface{}{"query": input})
    if err != nil || !res.Success {
        return "", fmt.Errorf("tool failed: %v", err)
    }
    return caps.LLM("Use this tool result.", fmt.Sprint(res.Content))
}

agent, err := v1beta.NewBuilder(cfg.Name).
    WithConfig(cfg).
    WithHandler(handler).
    Build()
if err != nil {
    log.Fatal(err)
}
```

Notes:
- `WithConfig` replaces builder state; do not combine it with `WithPreset` in the same chain.
- You can still add option helpers after `WithConfig` (e.g., `WithTools(v1beta.WithToolTimeout(...))`) to override specific knobs.

---

## Using Tools in Handlers

### Direct execution
```go
handler := func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    result, err := caps.Tools.Execute(ctx, "web_search", map[string]interface{}{
        "query": input,
        "max_results": 5,
    })
    if err != nil || !result.Success {
        return "", fmt.Errorf("tool error: %v", err)
    }
    return caps.LLM("Use these results to answer.", fmt.Sprint(result.Content))
}
```

### List and pick
```go
handler := func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    for _, tool := range caps.Tools.List() {
        if strings.Contains(tool.Name, "search") {
            res, err := caps.Tools.Execute(ctx, tool.Name, map[string]interface{}{"query": input})
            if err == nil {
                return fmt.Sprint(res.Content), nil
            }
        }
    }
    return caps.LLM("No tool matched; answer directly.", input)
}
```

### LLM-directed selection
Use the LLM to pick a tool name, then validate and execute it before responding. This pattern keeps the LLM’s choice constrained to known tools and fails safely if the choice is missing or unavailable.

```go
handler := func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    toolNames := caps.Tools.Available()
    choice, err := caps.LLM("Choose a tool name from this list and nothing else: "+strings.Join(toolNames, ", "), input)
    if err != nil {
        return "", err
    }

    selected := strings.TrimSpace(choice)
    if selected == "" || !caps.Tools.IsAvailable(selected) {
        return caps.LLM("No suitable tool found; answer directly.", input)
    }

    res, err := caps.Tools.Execute(ctx, selected, map[string]interface{}{"query": input})
    if err != nil || !res.Success {
        return "", fmt.Errorf("tool %s failed: %v", selected, err)
    }

    return caps.LLM("Use this tool result in the answer.", fmt.Sprint(res.Content))
}
```

---

## Caching

Enable caching via WithToolCaching(ttl) or TOML tools.cache. Use per-tool TTLs under tools.cache.tool_ttls for fine-grained control.

---

## Circuit Breaker

Configure circuit breaker limits under tools.circuit_breaker to avoid cascading failures (matches CircuitBreakerConfig).

---

## Monitoring and Health

```go
handler := func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    metrics := caps.Tools.GetMetrics()
    health := caps.Tools.HealthCheck(ctx)
    _ = metrics
    _ = health
    return caps.LLM("You are a helpful assistant.", input)
}
```

ToolMetrics and MCPHealthStatus are defined in the API for deeper inspection.

---

## Common Patterns

- Tool-first: attempt a tool, fall back to LLM on failure.
- Multi-tool pipeline: chain Execute calls and feed outputs to the LLM.
- Parallel tools: run independent tools concurrently and merge results.
- Conditional tools: ask the LLM whether external data is needed, then execute.

---

## Troubleshooting

- MCP server not connecting: verify server is running; check address/port/command.
- Tools not discovered: ensure discovery is enabled and ports are included.
- Tool timeouts: raise WithToolTimeout or server-side limits.

---

## Next Steps

- Custom handlers: custom-handlers.md
- Error handling: error-handling.md
- Performance: performance.md
- Examples: examples/
