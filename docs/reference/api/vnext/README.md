# vNext API Reference

**Unified agent and workflow surface for AgenticGoKit**

The `core/vnext` package delivers the next-generation agent runtime with a focused API surface, streamlined builder, and first-class integrations for memory, tools, MCP servers, and workflows. This is the **recommended API** for new projects, offering production-ready streaming, workflow orchestration, and comprehensive error handling.

## 📦 Package Overview

- `github.com/kunalkushwaha/agenticgokit/core/vnext`: primary entry point
- Consolidated interfaces: `Agent`, `Workflow`, `ToolManager`, `Memory`
- Functional options and presets for rapid agent construction
- Configuration loaders for single agents and multi-agent projects (TOML)
- Streaming, detailed results, tracing, and error codes baked in

## 🧭 How to Use This Reference

| Topic | Document |
| ----- | -------- |
| Core agent interface, results, run options | [agent.md](agent.md) |
| Streamlined builder & presets | [builder.md](builder.md) |
| Configuration schemas & loaders | [configuration.md](configuration.md) |
| Memory APIs, RAG helpers, sessions | [memory.md](memory.md) |
| Tool manager, MCP integration, caching | [tools.md](tools.md) |
| Workflow orchestration, streaming & composition | [workflow.md](workflow.md) |
| Streaming primitives & helpers | [streaming.md](streaming.md) |

## 🔑 Key Features

### SubWorkflow Composition
**Workflows can be used as agents within other workflows**, enabling powerful modular designs. Wrap any workflow as an agent using `NewSubWorkflowAgent()` or the builder's `WithSubWorkflow()` method.

```go
// Create nested workflows
analysisWorkflow, _ := vnext.NewParallelWorkflow(&vnext.WorkflowConfig{Name: "Analysis"})
analysisWorkflow.AddStep(vnext.WorkflowStep{Name: "sentiment", Agent: sentimentAgent})

// Wrap as agent
subAgent := vnext.NewSubWorkflowAgent("analysis", analysisWorkflow)

// Use in parent workflow
mainWorkflow.AddStep(vnext.WorkflowStep{Name: "analyze", Agent: subAgent})
```

See [workflow.md](workflow.md#-subworkflows-workflow-composition) for details.

## 🚀 Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/kunalkushwaha/agenticgokit/core/vnext"
)

func main() {
    // Quick agent creation
    agent, err := vnext.NewChatAgent("demo-bot")
    if err != nil {
        log.Fatal(err)
    }

    // Basic execution
    result, err := agent.Run(context.Background(), "Explain vnext workflows")
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Response:", result.Content)
    
    // Streaming execution  
    stream, err := agent.RunStream(ctx, "Generate a detailed report")
    if err != nil {
        log.Fatal(err)
    }
    
    for chunk := range stream.Chunks() {
        if chunk.Type == vnext.ChunkTypeDelta {
            fmt.Print(chunk.Delta) // Real-time token display
        }
    }
    
    finalResult, _ := stream.Wait()
    log.Println("Final:", finalResult.Content)
}
```

Read the individual documents for advanced topics like memory enrichment, MPC discovery, workflow streaming, and detailed execution traces.
