# Tool & MCP API (vNext)

**Unified tool orchestration with caching, MCP discovery, and LLM prompts**

Tools in vNext are managed through a single `ToolManager` abstraction. It coordinates built-in tools, custom providers, and MCP servers, exposing consistent execution semantics to agents and workflows.

## 🔑 ToolManager Interface

```go
type ToolManager interface {
    Execute(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error)
    List() []ToolInfo
    Available() []string
    IsAvailable(name string) bool

    ConnectMCP(ctx context.Context, servers ...MCPServer) error
    DisconnectMCP(serverName string) error
    DiscoverMCP(ctx context.Context) ([]MCPServerInfo, error)

    HealthCheck(ctx context.Context) map[string]MCPHealthStatus
    GetMetrics() ToolMetrics

    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

`ToolResult` contains `Success`, `Content`, and `Error` fields so you can propagate structured outcomes back into prompts or responses.

## ⚙️ Configuring Tools

Attach `ToolsConfig` via builder or TOML:

```go
builder.WithTools(
    vnext.WithToolTimeout(25 * time.Second),
    vnext.WithMaxConcurrentTools(4),
    vnext.WithToolCaching(15 * time.Minute),
    vnext.WithMCPDiscovery(),
)
```

TOML snippet:

```toml
[tools]
enabled = true
max_retries = 3
timeout = "30s"
rate_limit = 40
max_concurrent = 5

[tools.cache]
enabled = true
ttl = "15m"
eviction_policy = "lru"

[[tools.mcp.servers]]
name = "filesystem"
type = "stdio"
command = "mcp-fs"

[[tools.mcp.servers]]
name = "web"
type = "http_sse"
address = "http://localhost:8080/mcp"
```

`DefaultToolsConfig()` and `DefaultMCPConfig()` return fully populated defaults if you want to start programmatically.

## 🌐 MCP Integration

```go
manager, err := vnext.NewToolManagerWithMCP(vnext.DefaultToolsConfig(),
    vnext.MCPServer{Name: "web", Type: "http_sse", Address: "http://localhost:8080/mcp"},
)
if err != nil {
    log.Fatal(err)
}

defer manager.Shutdown(ctx)

if err := manager.Initialize(ctx); err != nil {
    log.Fatal(err)
}

info, _ := manager.DiscoverMCP(ctx)
log.Printf("discovered %d tools", len(info))
```

`ConnectMCP` adds servers at runtime, while `DiscoverMCP` performs auto-discovery when `ToolsConfig.MCP.Discovery` is enabled (port scan controlled by `ScanPorts`).

### MCP Health & Metrics

```go
health := manager.HealthCheck(ctx)
metrics := manager.GetMetrics()

log.Printf("cache hit rate: %.2f", metrics.CacheHitRate)
for server, status := range health {
    log.Printf("%s status: %s", server, status.Status)
}
```

`ToolMetrics` includes per-tool counters, latency, successes/failures, plus per-server details (`MCPServerMetrics`).

## 🧮 Executing Tools

```go
res, err := manager.Execute(ctx, "summarize", map[string]interface{}{
    "url": "https://example.com/blog",
})
if err != nil {
    log.Fatal(err)
}

if res.Success {
    fmt.Println("Summary:", res.Content)
} else {
    fmt.Println("Tool error:", res.Error)
}
```

Use `Available()` / `IsAvailable()` to guard tool calls before you emit plans to the LLM.

## 🧰 Formatting for LLM Prompts

Feed tool metadata into LLM system prompts so the model knows how to call them:

```go
tools := manager.List()
prompt := "You can call tools to gather data." + vnext.FormatToolsPromptForLLM(tools)
```

To parse tool usage attempts from the model, use the helper parsers from `tools.go` and `utils.go`:

```go
calls := vnext.ParseLLMToolCalls(response)
for _, call := range calls {
    name := call["name"].(string)
    args := call["args"].(map[string]interface{})
    res, _ := manager.Execute(ctx, name, args)
    fmt.Println(vnext.FormatToolResult(name, res))
}
```

## 🧭 Caching & Circuit Breaking

`CacheConfig` controls TTLs, eviction policy, backend, and per-tool overrides. Use `WithToolCaching(ttl)` to enable a simple in-memory cache or provide a bespoke backend through a plugin implementing `ToolCache`.

`CircuitBreakerConfig` guards against failing tools—configure thresholds, timeouts, and half-open behaviour via TOML or code (`tools.CircuitBreaker`).

## 🧩 Plugins & Factories

Register a custom tool manager factory when building extensions:

```go
vnext.SetToolManagerFactory(func(cfg *vnext.ToolsConfig) (vnext.ToolManager, error) {
    return newCustomManager(cfg), nil
})
```

Once registered, `NewToolManager` will return your implementation. This keeps the default fallback (`basicToolManager`) for situations where no plugin is loaded.

## 🧪 Validation

`ValidateToolsConfig`, `ValidateMCPConfig`, `ValidateMCPServer`, and `ValidateCacheConfig` surface configuration mistakes early.

```go
if err := vnext.ValidateToolsConfig(cfg.Tools); err != nil {
    log.Fatalf("invalid tools config: %v", err)
}
```

## 🔗 Related Docs

- [agent.md](agent.md): enabling tools per-call with `RunOptions`
- [builder.md](builder.md): wiring tool options into agent construction
- [workflow.md](workflow.md): using tool-enabled agents inside orchestration
- [streaming.md](streaming.md): streaming tool call events to clients
