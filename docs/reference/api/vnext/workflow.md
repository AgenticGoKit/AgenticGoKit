# Workflow API (vNext)

**Multi-agent orchestration with sequential, parallel, DAG, and loop modes**

The vNext workflow engine wraps multiple agents into an executable graph. It streams intermediate events, records step-level metrics, and shares memory between steps when configured.

## 🔑 Workflow Interface

```go
type Workflow interface {
    Run(ctx context.Context, input string) (*WorkflowResult, error)
    RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error)

    AddStep(step WorkflowStep) error
    SetMemory(memory Memory)
    GetConfig() *WorkflowConfig

    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

`WorkflowConfig` controls mode, timeout, max iterations, optional shared memory config, and named agent list for declarative setups.

## 🧱 Steps

```go
type WorkflowStep struct {
    Name         string
    Agent        Agent
    Condition    func(context.Context, *WorkflowContext) bool
    Dependencies []string
    Transform    func(string) string
    Metadata     map[string]interface{}
}
```

- `Condition` → skip steps dynamically (returns `false` to skip)
- `Dependencies` → DAG ordering
- `Transform` → mutate input before the step runs

## 🚀 Constructing Workflows

```go
workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{Timeout: 2 * time.Minute})
if err != nil {
    log.Fatal(err)
}

workflow.SetMemory(memory)
workflow.AddStep(vnext.WorkflowStep{Name: "gather", Agent: gatherAgent})
workflow.AddStep(vnext.WorkflowStep{Name: "summarise", Agent: summariseAgent})

result, err := workflow.Run(ctx, "Analyse customer tickets")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Final output:", result.FinalOutput)
```

Alternate constructors:

- `NewParallelWorkflow(config)`
- `NewDAGWorkflow(config)`
- `NewLoopWorkflow(config)`
- `NewWorkflow(config)` → respects `config.Mode`

## 📊 WorkflowResult

```go
type WorkflowResult struct {
    Success       bool
    FinalOutput   string
    StepResults   []StepResult
    Duration      time.Duration
    TotalTokens   int
    ExecutionPath []string
    Metadata      map[string]interface{}
    Error         string
}
```

`StepResult` captures per-step output, duration, token usage, and errors. `ExecutionPath` lists the order of successfully executed steps.

## 🔁 Modes

- **Sequential**: executes steps in order, piping output to the next step
- **Parallel**: fires all steps concurrently, aggregating outputs
- **DAG**: respects dependency graph (detects deadlocks if dependencies never resolve)
- **Loop**: replays the step list until `MaxIterations` or a context flag (`loop_continue`) stops the cycle

Example loop exit condition inside an agent:

```go
workflowCtx.Set("loop_continue", false) // stored via WorkflowContext in custom logic
```

## 🧠 WorkflowContext

```go
type WorkflowContext struct {
    WorkflowID   string
    SharedMemory Memory
    StepResults  map[string]*StepResult
    Variables    map[string]interface{}
}
```

Accessors:

```go
prev, ok := ctx.GetStepResult("summarise")
ctx.Set("loop_continue", prev.Success)
value, ok := ctx.Get("customer_id")
```

The context is shared across steps and exposes thread-safe getters/setters.

## 🌊 Streaming Workflows

**✅ Production Ready**: Workflow streaming is fully implemented with reliable context management and real-time token display.

```go
// Create multi-agent workflow
workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
    Mode:    vnext.Sequential,
    Timeout: 300 * time.Second,
})
workflow.AddStep(vnext.WorkflowStep{Name: "research", Agent: researchAgent})  
workflow.AddStep(vnext.WorkflowStep{Name: "analyze", Agent: analyzeAgent})

// Stream execution with real-time output
stream, err := workflow.RunStream(ctx, "Research project status")
if err != nil {
    log.Fatal(err)
}

for chunk := range stream.Chunks() {
    switch chunk.Type {
    case vnext.ChunkTypeMetadata:
        if stepName, ok := chunk.Metadata["step_name"].(string); ok {
            log.Printf("[%s] %s", stepName, chunk.Content)
        }
    case vnext.ChunkTypeDelta:
        fmt.Print(chunk.Delta) // Real-time token streaming
    case vnext.ChunkTypeDone:
        fmt.Println("\n✅ Step completed")
    }
}

final, _ := stream.Wait()
fmt.Printf("\n🎉 Workflow complete: %s\n", final.Content)
```

**Performance**: Workflow streaming adds ~36-160% overhead vs direct agents (measured) but provides multi-agent orchestration and real-time feedback. Streaming is reliable and context-cancel bugs have been resolved.

## 🧮 Shared Memory

Call `workflow.SetMemory(memory)` to give steps a shared provider. Steps can read/write via the agent’s builder configuration or manual memory usage inside custom handlers. Input/output snippets are stored with content types `workflow_step_input` and `workflow_step_output` when memory is present.

## 🧩 Dependency Graph Helpers

`buildInputFromDependencies` automatically aggregates outputs from dependency steps for DAG mode. Supply dependency names in `WorkflowStep.Dependencies` to signal ordering:

```go
workflow.AddStep(vnext.WorkflowStep{Name: "fetch", Agent: fetchAgent})
workflow.AddStep(vnext.WorkflowStep{Name: "analyse", Agent: analyseAgent, Dependencies: []string{"fetch"}})
workflow.AddStep(vnext.WorkflowStep{Name: "report", Agent: reportAgent, Dependencies: []string{"analyse"}})
```

## 🛠 Plugins

Override the default workflow engine by registering a factory:

```go
vnext.SetWorkflowFactory(func(cfg *vnext.WorkflowConfig) (vnext.Workflow, error) {
    return newCustomWorkflow(cfg), nil
})
```

## ✅ Best Practices

- Set `WorkflowConfig.Timeout` to cap the entire orchestration duration
- Use `MaxIterations` to avoid runaway loops
- Attach shared memory to allow later steps to inspect earlier outputs
- Stream for long-running workflows to give users live feedback
- Initialise/Shutdown workflows when managing long-lived agent instances (call once per lifecycle)

## 🔗 Related Docs

- [agent.md](agent.md) for agent execution details
- [streaming.md](streaming.md) for stream configuration
- [memory.md](memory.md) to configure shared RAG for workflows
- [tools.md](tools.md) when combining tool-enabled agents inside flows
