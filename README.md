# AgenticGoKit

> **⚠️ ALPHA RELEASE** - This project is in active development. APIs may change and features are still being stabilized. Use in production at your own risk. We welcome feedback and contributions!
>
> **📋 API Versioning Plan:**
> - **Current (v0.x)**: `v1beta` package is the recommended API
> - **v1.0 Release**: `v1beta` will become the primary `v1` package
> - **Legacy APIs**: `core` package will be removed in v1.0

**Production-ready Go framework for building intelligent multi-agent AI systems**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kunalkushwaha/agenticgokit)](https://goreportcard.com/report/github.com/kunalkushwaha/agenticgokit)
[![Build Status](https://github.com/kunalkushwaha/agenticgokit/workflows/CI/badge.svg)](https://github.com/kunalkushwaha/agenticgokit/actions)
[![Documentation](https://img.shields.io/badge/docs-latest-blue)](docs/README.md)

**The most productive way to build AI agents in Go.** AgenticGoKit provides a unified, streaming-first API for creating intelligent agents with built-in workflow orchestration, tool integration, and memory management. Start with simple single agents and scale to complex multi-agent workflows.

## Why Choose AgenticGoKit?

- **v1beta APIs**: Modern, streaming-first agent interface with comprehensive error handling
- **Real-time Streaming**: Watch your agents think and respond in real-time  
- **Multi-Agent Workflows**: Sequential, parallel, DAG, and loop orchestration patterns
- **High Performance**: Compiled Go binaries with minimal overhead
- **Rich Integrations**: Memory providers, tool discovery, MCP protocol support
- **Production Ready**: Works with OpenAI, Ollama, Azure OpenAI, HuggingFace out of the box

---

## Quick Start

**Start building immediately with the modern v1beta API:**

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create a chat agent with Ollama
    agent, err := v1beta.NewBuilder().
        WithLLM("ollama", "gemma3:1b").
        WithBaseURL("http://localhost:11434").
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Basic execution
    result, err := agent.Run(context.Background(), "Explain Go channels in 50 words")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Response:", result.Content)
    fmt.Printf("Duration: %.2fs | Tokens: %d\n", result.Duration.Seconds(), result.TokensUsed)
}
```

> **Note:** The `agentcli` scaffolding tool is being deprecated and will be replaced by the `agk` CLI in a future release.

## Streaming Workflows

**Watch your multi-agent workflows execute in real-time:**

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
    // Create specialized agents
    researcher, _ := v1beta.NewBuilder().
        WithName("researcher").
        WithSystemPrompt("You are a research specialist").
        WithLLM("ollama", "gemma3:1b").
        Build()
    
    analyzer, _ := v1beta.NewBuilder().
        WithName("analyzer").
        WithSystemPrompt("You are a data analyst").
        WithLLM("ollama", "gemma3:1b").
        Build()
    
    // Build workflow
    workflow, _ := v1beta.NewSequentialWorkflow(&v1beta.WorkflowConfig{
        Timeout: 300 * time.Second,
    })
    workflow.AddStep(v1beta.WorkflowStep{Name: "research", Agent: researcher})
    workflow.AddStep(v1beta.WorkflowStep{Name: "analyze", Agent: analyzer})
    
    // Execute with streaming
    stream, _ := workflow.RunStream(context.Background(), "Research Go best practices")
    
    for chunk := range stream.Chunks() {
        if chunk.Type == v1beta.ChunkTypeDelta {
            fmt.Print(chunk.Delta) // Real-time token streaming!
        }
    }
    
    result, _ := stream.Wait()
    fmt.Printf("\nComplete: %s\n", result.Content)
}
```

## Core Features

### v1beta APIs (Production-Ready)
- **Unified Agent Interface**: Single API for all agent operations
- **Real-time Streaming**: Watch tokens generate in real-time
- **Multi-Agent Workflows**: Sequential, parallel, DAG, loop orchestration, and subworkflows
- **Memory & RAG**: Built-in persistence and retrieval
- **Tool Integration**: MCP protocol, function calling, tool discovery
- **Subworkflows**: Compose workflows as agents for complex hierarchies
- **Multiple LLM Providers**: OpenAI, Azure OpenAI, Ollama, HuggingFace
- **Flexible Configuration**: Builder pattern with type-safe options

## API Usage

**v1beta provides clean, modern APIs for building AI agents:**

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

// Single agent
agent, _ := v1beta.NewBuilder().WithLLM("ollama", "gemma3:1b").Build()
result, _ := agent.Run(ctx, "Hello world")

// Streaming agent  
stream, _ := agent.RunStream(ctx, "Write a story")
for chunk := range stream.Chunks() { /* real-time output */ }

// Multi-agent workflow
workflow, _ := v1beta.NewSequentialWorkflow(config)
stream, _ := workflow.RunStream(ctx, input) 
```

## Examples & Templates

## Examples

### Basic Agent
```go
// Basic chat agent
agent, _ := v1beta.NewBuilder().
    WithLLM("ollama", "gemma3:1b").
    Build()

// With custom configuration  
agent, _ := v1beta.NewBuilder().
    WithName("helper").
    WithLLM("ollama", "gemma3:1b").
    WithBaseURL("http://localhost:11434").
    WithMemory(&v1beta.MemoryConfig{
        Provider: "memory",
        RAG: &v1beta.RAGConfig{Enabled: true},
    }).
    WithTools(&v1beta.ToolsConfig{
        Enabled: true,
        MCP: &v1beta.MCPConfig{Enabled: true},
    }).
    Build()
```

### Workflow Orchestration

v1beta supports four workflow patterns for orchestrating multiple agents:

```go
// Sequential Workflow - Execute agents one after another
sequential, _ := v1beta.NewSequentialWorkflow(&v1beta.WorkflowConfig{
    Name: "research-pipeline",
    Timeout: 300 * time.Second,
})
sequential.AddStep(v1beta.WorkflowStep{Name: "research", Agent: researchAgent})
sequential.AddStep(v1beta.WorkflowStep{Name: "analyze", Agent: analyzerAgent})
sequential.AddStep(v1beta.WorkflowStep{Name: "summarize", Agent: summarizerAgent})

// Parallel Workflow - Execute agents concurrently
parallel, _ := v1beta.NewParallelWorkflow(&v1beta.WorkflowConfig{
    Name: "multi-analysis",
})
parallel.AddStep(v1beta.WorkflowStep{Name: "research", Agent: researchAgent})
parallel.AddStep(v1beta.WorkflowStep{Name: "fact-check", Agent: factChecker})
parallel.AddStep(v1beta.WorkflowStep{Name: "sentiment", Agent: sentimentAgent})

// DAG Workflow - Execute with dependencies
dag, _ := v1beta.NewDAGWorkflow(&v1beta.WorkflowConfig{
    Name: "dependent-workflow",
})
dag.AddStep(v1beta.WorkflowStep{Name: "fetch", Agent: fetchAgent})
dag.AddStep(v1beta.WorkflowStep{
    Name: "analyze",
    Agent: analyzerAgent,
    Dependencies: []string{"fetch"},
})
dag.AddStep(v1beta.WorkflowStep{
    Name: "report",
    Agent: reportAgent,
    Dependencies: []string{"analyze"},
})

// Loop Workflow - Iterate until condition met
loop, _ := v1beta.NewLoopWorkflow(&v1beta.WorkflowConfig{
    Name: "iterative-refinement",
    MaxIterations: 5,
})
loop.AddStep(v1beta.WorkflowStep{Name: "generate", Agent: generatorAgent})
loop.AddStep(v1beta.WorkflowStep{Name: "review", Agent: reviewAgent})
loop.SetLoopCondition(v1beta.Conditions.OutputContains("APPROVED"))
```

### Subworkflows - Workflows as Agents

Compose workflows as agents for complex hierarchical orchestration:

```go
// Create a research workflow
researchWorkflow, _ := v1beta.NewParallelWorkflow(&v1beta.WorkflowConfig{
    Name: "research-team",
})
researchWorkflow.AddStep(v1beta.WorkflowStep{Name: "web", Agent: webResearcher})
researchWorkflow.AddStep(v1beta.WorkflowStep{Name: "academic", Agent: academicResearcher})

// Wrap workflow as an agent
researchAgent := v1beta.NewSubWorkflowAgent("research", researchWorkflow)

// Use in main workflow
mainWorkflow, _ := v1beta.NewSequentialWorkflow(&v1beta.WorkflowConfig{
    Name: "report-generation",
})
mainWorkflow.AddStep(v1beta.WorkflowStep{
    Name: "research",
    Agent: researchAgent, // Workflow acting as agent!
})
mainWorkflow.AddStep(v1beta.WorkflowStep{Name: "write", Agent: writerAgent})

// Execute with nested streaming
stream, _ := mainWorkflow.RunStream(ctx, "Research AI safety")
for chunk := range stream.Chunks() {
    // Nested agent outputs are automatically forwarded
    if chunk.Type == v1beta.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
}
```

### Provider Support

AgenticGoKit supports multiple LLM providers out of the box:

```go
// OpenAI
agent, _ := v1beta.NewBuilder().
    WithLLM("openai", "gpt-4").
    WithAPIKey(os.Getenv("OPENAI_API_KEY")).
    Build()

// Azure OpenAI
agent, _ := v1beta.NewBuilder().
    WithLLM("azure", "gpt-4").
    WithBaseURL("https://your-resource.openai.azure.com").
    WithAPIKey(os.Getenv("AZURE_OPENAI_API_KEY")).
    Build()

// Ollama (Local)
agent, _ := v1beta.NewBuilder().
    WithLLM("ollama", "gemma3:1b").
    WithBaseURL("http://localhost:11434").
    Build()

// HuggingFace
agent, _ := v1beta.NewBuilder().
    WithLLM("huggingface", "meta-llama/Llama-2-7b-chat-hf").
    WithAPIKey(os.Getenv("HUGGINGFACE_API_KEY")).
    Build()
```

## Learning Resources

### Working Examples ([`examples/`](examples/))
- **[Story Writer Chat v2](examples/story-writer-chat-v2/)** - Real-time multi-agent workflow with streaming
- **[HuggingFace Integration](examples/huggingface-quickstart/)** - Using HuggingFace models
- **[MCP Integration](examples/mcp-integration/)** - Model Context Protocol tools
- **[Streaming Workflow](examples/streaming_workflow/)** - Streaming multi-agent workflows
- **[Simple Streaming](examples/simple-streaming/)** - Basic streaming examples
- **[Ollama Quickstart](examples/ollama-quickstart/)** - Getting started with Ollama

### Documentation
- **[v1beta API Reference](v1beta/README.md)** - Complete API documentation
- **[Examples Directory](examples/)** - Full collection of working examples

## Development

```bash
# Clone and build
git clone https://github.com/kunalkushwaha/agenticgokit.git
cd agenticgokit
make build

# Run tests
make test

# Run examples
cd examples/ollama-quickstart
go run .
```

## API Versioning & Roadmap

### Current Status (v0.x - Alpha)

- **Recommended**: Use `v1beta` package for all new projects
- **Import Path**: `github.com/agenticgokit/agenticgokit/v1beta`
- **Stability**: Beta - API is mostly stable, minor changes possible

### v1.0 Release Plan

**What's Changing:**
- `v1beta` package will become the primary `v1` API
- Legacy `core` package will be **removed entirely**
- Clean, stable API with semantic versioning guarantees

**Migration Path:**
- If you're using `v1beta`: Minimal changes (import path update only)
- If you're using `core`: Migrate to `v1beta` now to prepare

**Timeline:**
- v0.x (Current): `v1beta` stabilization and testing
- v1.0 (Planned): `v1beta` → `v1`, remove `core` package

### Why v1beta Now?

The `v1beta` package represents our next-generation API design:
- ✅ Streaming-first architecture
- ✅ Unified builder pattern
- ✅ Better error handling
- ✅ Workflow composition
- ✅ Production-ready

By using `v1beta` today, you're future-proofing your code for the v1.0 release.

## Resources

- **Website**: [www.agenticgokit.com](https://www.agenticgokit.com)
- **Documentation**: [docs.agenticgokit.com](https://docs.agenticgokit.com)
- **Examples**: [examples/](examples/)
- **Discussions**: [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)
- **Issues**: [GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues)

## Contributing

We welcome contributions! See [docs/contributors/ContributorGuide.md](docs/contributors/ContributorGuide.md) for getting started.

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
