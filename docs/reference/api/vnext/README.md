# vNext API Reference

**Unified agent and workflow surface for AgenticGoKit**

The `core/vnext` package delivers the next-generation agent runtime with a focused API surface, streamlined builder, and first-class integrations for memory, tools, MCP servers, and workflows. These documents mirror the existing `docs/reference/api` layout but focus specifically on vNext types so you can adopt the new primitives without digging through the source.

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
| Workflow orchestration & streaming | [workflow.md](workflow.md) |
| Streaming primitives & helpers | [streaming.md](streaming.md) |

## 🚀 Quick Start

```go
package main

import (
    "context"
    "log"

    vnext "github.com/kunalkushwaha/agenticgokit/core/vnext"
)

func main() {
    agent, err := vnext.NewChatAgent("demo-bot")
    if err != nil {
        log.Fatal(err)
    }

    result, err := agent.Run(context.Background(), "Summarise the vNext API")
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Response:", result.Content)
}
```

Read the individual documents for advanced topics like memory enrichment, MPC discovery, workflow streaming, and detailed execution traces.
