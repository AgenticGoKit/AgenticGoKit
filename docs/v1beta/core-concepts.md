# Core Concepts

This guide explains the major building blocks in AgenticGoKit v1beta and how they relate. It stays high-level and points to the right detailed guides (configuration, streaming, tools, memory/RAG, workflows).

---

## What You Will Learn

- What an agent is and how to create one (presets vs builder)
- How handlers plug in custom logic with access to LLM, tools, and memory
- How tools and memory are configured via functional options
- How runtime execution works (Run/RunWithOptions and streaming)
- Where to go next for deeper topics

---

## Components at a Glance

- **Agents**: The runtime unit that executes LLM calls, tools, and memory.
- **Handlers**: Your custom logic; you control how the agent responds.
- **Tools**: Extend capabilities (MCP, discovery, timeouts, concurrency, caching).
- **Memory**: Context retention (chromem default) with RAG options.

For detailed configuration of knobs, see the dedicated [configuration guide](configuration.md).

---

## Agents

Agents can be created quickly with presets or assembled with the builder.

### Preset Constructors (fastest)

```go
agent, err := v1beta.NewChatAgent("Assistant",
    v1beta.WithLLM("openai", "gpt-4"),
)

// Other presets
v1beta.NewResearchAgent("Researcher")
v1beta.NewDataAgent("Analyst")
v1beta.NewWorkflowAgent("Orchestrator")
```

Use a preset when you want sensible defaults for temperature, memory, and (where relevant) tools/workflow.

### Builder (full control)

```go
agent, err := v1beta.NewBuilder("CustomAgent").
    WithPreset(v1beta.ChatAgent). // or WithConfig(cfg) or WithLLM(provider, model) via options
    WithTools(
        v1beta.WithMCPDiscovery(),
        v1beta.WithToolTimeout(30 * time.Second),
    ).
    WithMemory(
        v1beta.WithMemoryProvider("chromem"),
        v1beta.WithRAG(4096, 0.7, 0.3),
    ).
    WithHandler(myHandler).
    Build()
```

- Start with **one** of: `WithPreset`, or `WithConfig`, or the preset constructors plus `WithLLM` options.
- Add features with `WithTools`, `WithMemory`, `WithWorkflow`, `WithSubWorkflow`, then set `WithHandler`.
- Prefer presets for consistency; drop to `WithConfig` only when composing config programmatically.

### Agent Interface (for reference)

```go
type Agent interface {
    Name() string
    Run(ctx context.Context, input string) (*Result, error)
    RunWithOptions(ctx context.Context, input string, opts *RunOptions) (*Result, error)
    RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error)
    RunStreamWithOptions(ctx context.Context, input string, runOpts *RunOptions, streamOpts ...StreamOption) (Stream, error)
    Config() *Config
    Capabilities() []string
    Memory() Memory
    Initialize(ctx context.Context) error
    Cleanup(ctx context.Context) error
}
```

---

## Handlers

Handlers are your customization point. v1beta uses a single handler signature with capabilities.

```go
handler := func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Call LLM
    text, err := caps.LLM("You are helpful", input)
    if err != nil {
        return "", err
    }

    // Optional tools
    if caps.Tools != nil {
        // caps.Tools.Execute(ctx, name, args) or higher-level helpers
    }

    // Optional memory
    if caps.Memory != nil {
        // Use memory to build context or store signals
    }

    return text, nil
}

agent, err := v1beta.NewBuilder("Custom").
    WithPreset(v1beta.ChatAgent).
    WithHandler(handler).
    Build()
```

- `Capabilities` exposes `LLM`, `Tools`, and `Memory`—no separate handler types are required.
- Keep handlers small; delegate heavy lifting to tools and memory.

---

## Tools

Tools add capabilities beyond plain LLM responses.

**Tool interface (brief):**
```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}
```

**Configure tools with ToolOption:**
- `WithMCP(servers ...MCPServer)` – attach MCP servers
- `WithMCPDiscovery(scanPorts ...int)` – auto-discover servers
- `WithToolTimeout(timeout time.Duration)` – per-call timeout
- `WithMaxConcurrentTools(max int)` – concurrency cap
- `WithToolCaching(ttl time.Duration)` – cache tool results

**Example:**
```go
agent, err := v1beta.NewBuilder("ToolAgent").
    WithPreset(v1beta.ChatAgent).
    WithTools(
        v1beta.WithMCPDiscovery(),
        v1beta.WithToolTimeout(30 * time.Second),
    ).
    Build()
```

See the [tool integration guide](tool-integration.md) for deeper examples.

---

## Memory

Memory is enabled by default with the embedded `chromem` provider. Configure it with MemoryOption.

- `WithMemoryProvider(provider string)` – e.g., "chromem", "pgvector", "weaviate"
- `WithRAG(maxTokens, personalWeight, knowledgeWeight)` – enable retrieval augmentation
- `WithSessionScoped()` – session-isolated memory
- `WithContextAware()` – context-aware retrieval

**Example:**
```go
agent, err := v1beta.NewBuilder("Memo").
    WithPreset(v1beta.ChatAgent).
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithRAG(4000, 0.6, 0.4),
        v1beta.WithSessionScoped(),
    ).
    Build()
```

To disable memory entirely, pass a Config with `Memory.Enabled = false` via `WithConfig`.

See the [memory & RAG guide](memory-and-rag.md) for details.

---

## Configuration (where it fits)

Core concepts focus on what the pieces are. For how to wire every knob (builder options, Config struct, TOML), see the dedicated [configuration guide](configuration.md). Use presets or the builder with functional options for most cases; drop to `Config` when assembling settings programmatically.

---

## Runtime Execution

- `Run(ctx, input)` – simplest path
- `RunWithOptions(ctx, input, opts)` – per-call overrides via `RunOptions`
- `RunStream` / `RunStreamWithOptions` – streaming responses (see streaming guide)

**RunOptions (key fields):**
```go
type RunOptions struct {
    Tools       []string
    ToolMode    string // "auto", "specific", "none"
    Memory      *MemoryOptions
    SessionID   string
    Timeout     time.Duration
    Context     map[string]interface{}
    MaxRetries  int
    MaxTokens   int
    Temperature *float64
    DetailedResult bool
    IncludeTrace   bool
    IncludeSources bool
    Images      []ImageData
    Audio       []AudioData
    Video       []VideoData
}
```

**Example override:**
```go
temp := 0.8
opts := &v1beta.RunOptions{
    Temperature: &temp,
    MaxTokens:   800,
    SessionID:   "user-42",
}

result, err := agent.RunWithOptions(ctx, "Summarize this", opts)
```

See the [streaming guide](streaming.md) for chunking patterns and stream options.

---

## Results (high level)

`Result` includes core fields like `Content`, `Success`, `Duration`, `Metadata`, and modality fields (`Images`, `Audio`, `Video`). For tool calls, memory context, and tracing, inspect `Metadata` or use detailed results. Streaming returns chunks (delta, thought, tool_call, tool_result, metadata, error, done); see the streaming guide for full types.

---

## Middleware

v1beta defines an `AgentMiddleware` interface with `BeforeRun/AfterRun`. It is not wired through the streamlined builder yet; you can wrap handlers or compose at the workflow layer. If you add middleware, ensure you propagate context and respect timeouts.

---

## Next Steps

- Configuration details: [configuration.md](configuration.md)
- Streaming: [streaming.md](streaming.md)
- Tools: [tool-integration.md](tool-integration.md)
- Memory & RAG: [memory-and-rag.md](memory-and-rag.md)
- Workflows: [workflows.md](workflows.md)
- Examples: [examples/](examples/)

Use this overview to pick your path: start with a preset + handler, add tools and memory via options, then refine with runtime options and workflows as needed.
