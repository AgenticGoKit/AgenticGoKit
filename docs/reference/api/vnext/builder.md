# Agent Builder (vNext)

**8-method builder for configuring agents, presets, and handlers**

The vNext builder shrinks the agent construction API to a handful of expressive calls. Use presets for quick starts, functional options to tweak LLM/memory/tool behaviour, or supply a custom handler that drives execution with full access to capabilities.

## 🧱 Builder Interface

```go
type Builder interface {
    WithConfig(config *Config) Builder
    WithPreset(preset PresetType) Builder
    WithHandler(handler HandlerFunc) Builder
    Build() (Agent, error)

    WithMemory(opts ...MemoryOption) Builder
    WithTools(opts ...ToolOption) Builder
    WithWorkflow(opts ...WorkflowOption) Builder
    Clone() Builder
}
```

Call `Build()` once per builder instance; reuse `Clone()` for variations.

## 🚀 Quick Construction

```go
agent, err := vnext.NewBuilder("support-bot").
    WithPreset(vnext.ChatAgent).
    WithLLM("openai", "gpt-4o").
    WithAgentTimeout(45 * time.Second).
    WithMemory(
        vnext.WithMemoryProvider("memory"),
        vnext.WithSessionScoped(),
        vnext.WithContextAware(),
    ).
    WithTools(
        vnext.WithToolTimeout(30 * time.Second),
        vnext.WithMaxConcurrentTools(5),
        vnext.WithMCPDiscovery(8080, 8081),
    ).
    Build()
```

The free functions `WithLLM`, `WithSystemPrompt`, `WithAgentTimeout`, `WithDebugMode`, etc., are plain functional options imported from `core/vnext` (no builder method prefix).

## 🧩 Presets

`PresetType` values give you sensible defaults:

- `ChatAgent`: conversational assistant (higher temperature, session memory)
- `ResearchAgent`: long-form research with tools enabled and longer timeouts
- `DataAgent`: deterministic analytic helper with low temperature
- `WorkflowAgent`: orchestration-first agent prepared for workflows

```go
researcher, err := vnext.NewResearchAgent("insights",
    vnext.WithLLM("openai", "gpt-4-turbo"),
    vnext.WithAgentTimeout(90 * time.Second),
)
```

## 🛠️ Custom Handlers

Swap in your own logic while still tapping the LLM, tools, and memory stacks.

```go
handler := func(ctx context.Context, input string, caps *vnext.Capabilities) (string, error) {
    if strings.Contains(input, "weather") {
        res, err := caps.Tools.Execute(ctx, "get_weather", map[string]interface{}{"location": input})
        if err != nil || !res.Success {
            return "Weather service unavailable", err
        }
        return fmt.Sprintf("Weather: %v", res.Content), nil
    }

    return caps.LLM(
        "You are a concise assistant",
        input,
    )
}

agent, _ := vnext.NewBuilder("concierge").
    WithHandler(handler).
    WithTools(vnext.WithMCP(vnext.MCPServer{Name: "weather", Type: "http_sse", Address: "http://localhost:8080/mcp"})).
    Build()
```

`Capabilities` exposes:

- `LLM(systemPrompt, userPrompt string) (string, error)`
- `Tools` (the configured `ToolManager`)
- `Memory` (underlying memory provider)

## 🧠 Memory Options

```go
builder.WithMemory(
    vnext.WithMemoryProvider("pgvector"),
    vnext.WithSessionScoped(),
    vnext.WithRAG(4096, 0.4, 0.6),
)
```

- Provider names map to registered memory factories (see [memory.md](memory.md))
- `WithRAG` turns on RAG context building with weights and token limits
- Options are merged into the builder’s `MemoryConfig`

## 🛠 Tool Options

```go
builder.WithTools(
    vnext.WithToolTimeout(20 * time.Second),
    vnext.WithMaxConcurrentTools(3),
    vnext.WithToolCaching(10 * time.Minute),
    vnext.WithMCP(
        vnext.MCPServer{Name: "fs", Type: "stdio", Command: "mcp-fs"},
    ),
)
```

`WithMCPDiscovery` configures automatic MCP port scanning; combine with explicit servers if needed.

## 🔁 Workflow Options

```go
builder.WithWorkflow(
    vnext.WithWorkflowMode(string(vnext.Sequential)),
    vnext.WithWorkflowAgents("agent-one", "agent-two"),
    vnext.WithMaxIterations(8),
)
```

The workflow config is picked up when the agent is orchestrated via the vNext workflow engine.

## 🔄 SubWorkflow Options

**Build agents that wrap workflows**, enabling workflow composition and nesting.

```go
// Create a workflow first
parallelWorkflow, _ := vnext.NewParallelWorkflow(&vnext.WorkflowConfig{
    Name: "ParallelAnalysis",
})
parallelWorkflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
parallelWorkflow.AddStep(vnext.WorkflowStep{Name: "step2", Agent: agent2})

// Use builder to create SubWorkflow agent
subAgent, err := vnext.NewBuilder("sub-agent").
    WithSubWorkflow(
        vnext.WithWorkflowInstance(parallelWorkflow),
        vnext.WithSubWorkflowMaxDepthBuilder(5),
        vnext.WithSubWorkflowDescriptionBuilder("Parallel analysis pipeline"),
    ).
    Build()

// Now use subAgent in another workflow
mainWorkflow.AddStep(vnext.WorkflowStep{Name: "analyze", Agent: subAgent})
```

### SubWorkflow Builder Options

- `WithWorkflowInstance(workflow)` - **Required**: Provides the workflow to wrap
- `WithSubWorkflowMaxDepthBuilder(depth)` - Sets maximum nesting depth (default: 10)
- `WithSubWorkflowDescriptionBuilder(desc)` - Sets description for the SubWorkflow agent

### Alternative: Direct Construction

For simpler cases, use `NewSubWorkflowAgent()` directly:

```go
subAgent := vnext.NewSubWorkflowAgent("analysis", workflow,
    vnext.WithSubWorkflowMaxDepth(5),
    vnext.WithSubWorkflowDescription("Analysis pipeline"),
)
```

### Use Cases

- **Modular Workflows**: Break complex flows into reusable subworkflows
- **Conditional Branching**: Use different subworkflows based on conditions
- **Parallel Processing**: Nest parallel workflows within sequential flows
- **Multi-Level Orchestration**: Create hierarchical agent systems

See [workflow.md](workflow.md#-subworkflows-workflow-composition) for more details on SubWorkflow composition.

## ♻️ Cloning Builders

Freeze the base configuration and create custom variants:

```go
base := vnext.NewBuilder("support-base").
    WithPreset(vnext.ChatAgent).
    WithLLM("openai", "gpt-4o").
    WithMemory(vnext.WithSessionScoped())

agentUS, _ := base.Clone().WithSystemPrompt("Answer as a US-based agent").Build()
agentEU, _ := base.Clone().WithSystemPrompt("Answer as an EU-based agent").Build()
```

`Clone()` deep copies nested configs (Memory, Tools, Workflow, Tracing), so adjustments on clones will not mutate the base builder.

## 🧪 Validation

`Build()` validates the configuration (name, provider/model, positive timeout). Errors are returned as soon as validation fails:

```go
if _, err := vnext.NewBuilder("").Build(); err != nil {
    log.Printf("builder failed: %v", err)
}
```

## 🔗 Related Docs

- [agent.md](agent.md) for execution APIs
- [configuration.md](configuration.md) for loading TOML-based configs
- [tools.md](tools.md) if you need MCP servers or tool caching
- [workflow.md](workflow.md) for orchestrating multiple agents
