# Agent API (vNext)

**Unified execution surface for agents, streaming, and detailed results**

The vNext `Agent` interface collapses handlers, middleware, and result types into a single cohesive contract. Use it to run LLM-first agents, enable streaming, and opt into granular telemetry without juggling multiple abstractions.

## 🔑 Core Interface

```go
type Agent interface {
    Name() string
    Run(ctx context.Context, input string) (*Result, error)
    RunWithOptions(ctx context.Context, input string, opts *RunOptions) (*Result, error)

    RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error)
    RunStreamWithOptions(ctx context.Context, input string, runOpts *RunOptions, streamOpts ...StreamOption) (Stream, error)

    Config() *Config
    Capabilities() []string

    Initialize(ctx context.Context) error
    Cleanup(ctx context.Context) error
}
```

The agent always receives a simple string input and returns a rich `Result`. Use `RunWithOptions` when you need to customise tools, memory sessions, tracing or retries for a single call.

## 📦 Result Payload

`Result` aggregates everything you need after execution:

```go
type Result struct {
    Success     bool
    Content     string
    Duration    time.Duration
    TraceID     string
    Metadata    map[string]interface{}
    TokensUsed  int
    ToolsCalled []string
    MemoryUsed  bool
    // plus ToolExecutions, LLMInteractions, legacy fields, etc.
}
```

Helpers:

- `result.Text()` returns the primary text content (legacy compatibility)
- `result.IsSuccess()` includes error field checks
- `result.ToolExecutions`, `result.LLMInteractions` expose deep traces when `RunOptions.DetailedResult` is enabled

## ⚙️ Run Options

Use the factory helpers or chain methods directly:

```go
opts := vnext.NewRunOptions().
    SetTools("web_search", "calculator").
    SetTimeout(45 * time.Second).
    SetDetailedResult(true).
    SetTracing(true, "enhanced").
    SetMemory("session-42", &vnext.MemoryOptions{Enabled: true, SessionScoped: true}).
    AddContext("customer_id", "cust-99")

result, err := agent.RunWithOptions(ctx, "Plan a travel itinerary", opts)
```

Common option helpers:

- `RunWithTools(...)`, `RunWithMemory(sessionID, *MemoryOptions)`
- `RunWithStreaming()` for backwards-compatible streaming flag (prefer `RunStream`)
- `RunWithDetailedResult()` to fill metrics, tool execution info, and source attributions
- `RunWithTimeout(duration)` to bound a single request

## 🚀 Basic Usage

```go
agent, _ := vnext.NewChatAgent("support-bot")

result, err := agent.Run(ctx, "Summarise the last release notes")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Response: ", result.Content)
```

### With Per-Call Tool Control

```go
opts := vnext.RunWithTools("search", "fetch_docs")
opts.SetDetailedResult(true)

run := "Find latest API changes and list doc URLs"
result, err := agent.RunWithOptions(ctx, run, opts)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Tools used:", result.ToolsCalled)
```

### With Memory Sessions

```go
memOpts := &vnext.MemoryOptions{ // enables agent-specific memory bridge
    Enabled:       true,
    Provider:      "memory",
    SessionScoped: true,
}
opts := vnext.RunWithMemory("session-a1b2", memOpts).SetDetailedResult(true)

agent.RunWithOptions(ctx, "Remember that my name is Priya", opts)
agent.RunWithOptions(ctx, "Who am I?", opts)
```

### Streaming Responses

```go
stream, err := agent.RunStream(ctx, "Provide a live coding walkthrough",
    vnext.WithThoughts(),
    vnext.WithStreamHandler(func(chunk *vnext.StreamChunk) bool {
        switch chunk.Type {
        case vnext.ChunkTypeDelta:
            fmt.Print(chunk.Delta)
        case vnext.ChunkTypeThought:
            log.Printf("Thought: %s", chunk.Content)
        }
        return true
    }),
)
if err != nil {
    log.Fatal(err)
}

final, err := stream.Wait()
fmt.Println("\nFinal response:", final.Content)
```

Use `RunStreamWithOptions` when you also need `RunOptions` (memory sessions, tracing, etc.) alongside streaming configuration.

## 🧱 Capability Detection

`agent.Capabilities()` reports the features enabled by configuration:

- `llm` is always present
- `memory` and `rag` appear when memory is configured with RAG options
- `tools` shows up once tools/MCP are enabled
- `workflow` is present for workflow-aware agents
- `custom_handler` indicates a user-supplied handler

## 🛡️ Structured Errors

vNext agents surface typed errors through `*AgentError` with `ErrorCode` constants (`ErrToolExecutionFailed`, `ErrMemoryRetrieveFailed`, etc.). You can check them with `errors.As` and the helper methods:

```go
result, err := agent.Run(ctx, input)
if err != nil {
    var agentErr *vnext.AgentError
    if errors.As(err, &agentErr) && agentErr.IsErrorCode(vnext.ErrToolExecutionFailed) {
        log.Printf("tool failure: %s", agentErr)
    }
}
```

`Result.Error` is still populated for compatibility, but prefer the structured error to branch on failure types.

## 🧩 Middleware Hooks

Implement `AgentMiddleware` and register it via the builder (see [builder.md](builder.md)) to wrap execution:

```go
type auditMiddleware struct{}

func (m auditMiddleware) BeforeRun(ctx context.Context, input string) (context.Context, string, error) {
    log.Println("incoming input", input)
    return ctx, input, nil
}

func (m auditMiddleware) AfterRun(ctx context.Context, input string, res *vnext.Result, err error) (*vnext.Result, error) {
    log.Println("duration", res.Duration)
    return res, err
}
```

Attach middleware through custom builder implementations (see `core/vnext/unified_agent_builder.go` for the default stack).

## ✅ Next Steps

Continue with:

- [builder.md](builder.md) for constructing agents and attaching middleware
- [memory.md](memory.md) for persistence, RAG, and session handling
- [tools.md](tools.md) for MCP connectivity and tool orchestration
- [streaming.md](streaming.md) if you need deeper control over streaming semantics
