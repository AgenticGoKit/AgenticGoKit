# AgenticGoKit

> **⚠️ ALPHA RELEASE** - This project is in active development. APIs may change and features are still being stabilized. Use in production at your own risk. We welcome feedback and contributions!

**Production-ready Go framework for building intelligent multi-agent AI systems**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kunalkushwaha/agenticgokit)](https://goreportcard.com/report/github.com/kunalkushwaha/agenticgokit)
[![Build Status](https://github.com/kunalkushwaha/agenticgokit/workflows/CI/badge.svg)](https://github.com/kunalkushwaha/agenticgokit/actions)
[![Documentation](https://img.shields.io/badge/docs-latest-blue)](docs/README.md)

**The most productive way to build AI agents in Go.** AgenticGoKit provides a unified, streaming-first API for creating intelligent agents with built-in workflow orchestration, tool integration, and memory management. Start with simple single agents and scale to complex multi-agent workflows.

## Why Choose AgenticGoKit?

- **vnext APIs**: Modern, streaming-first agent interface with comprehensive error handling
- **Real-time Streaming**: Watch your agents think and respond in real-time  
- **Multi-Agent Workflows**: Sequential, parallel, and DAG orchestration patterns
- **High Performance**: Compiled Go binaries with minimal overhead
- **Rich Integrations**: Memory providers, tool discovery, MCP protocol support
- **Zero Dependencies**: Works with OpenAI, Ollama, Azure OpenAI out of the box

---

## Quick Start

### Option 1: Use vnext APIs (Recommended)

**Start building immediately with the modern vnext API:**

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core/vnext"
)

func main() {
    // Create a chat agent with Ollama
    agent, err := vnext.NewChatAgent("assistant", 
        vnext.WithLLM("ollama", "gemma3:1b", "http://localhost:11434"),
    )
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

### Option 2: Use CLI Generator (Scaffolding)

**Generate complete projects with agentcli:**

```bash
# Install CLI
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest

# Create project with scaffolding
agentcli create my-agent --template basic
cd my-agent

# Set up environment 
export AZURE_OPENAI_API_KEY=your-key
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/

# Run generated code
go run .
```

## Streaming Workflows

**Watch your multi-agent workflows execute in real-time:**

```go
package main

import (
    "context"
    "fmt" 
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core/vnext"
)

func main() {
    // Create specialized agents
    researcher, _ := vnext.NewResearchAgent("researcher")
    analyzer, _ := vnext.NewDataAgent("analyzer")
    
    // Build workflow
    workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
        Timeout: 300 * time.Second,
    })
    workflow.AddStep(vnext.WorkflowStep{Name: "research", Agent: researcher})
    workflow.AddStep(vnext.WorkflowStep{Name: "analyze", Agent: analyzer})
    
    // Execute with streaming
    stream, _ := workflow.RunStream(context.Background(), "Research Go best practices")
    
    for chunk := range stream.Chunks() {
        if chunk.Type == vnext.ChunkTypeDelta {
            fmt.Print(chunk.Delta) // Real-time token streaming!
        }
    }
    
    result, _ := stream.Wait()
    fmt.Printf("\nComplete: %s\n", result.Content)
}
```

## Core Features

### vnext APIs (Production-Ready)
- **Unified Agent Interface**: Single API for all agent operations
- **Real-time Streaming**: Watch tokens generate in real-time
- **Multi-Agent Workflows**: Sequential, parallel, DAG orchestration  
- **Memory & RAG**: Built-in persistence and retrieval
- **Tool Integration**: MCP protocol, function calling

### Project Generation  
- **CLI Scaffolding**: Generate complete projects with `agentcli create`
- **Multiple Templates**: Chat, research, RAG, workflow patterns
- **Configuration**: TOML-based with environment overrides
- **Visualization**: Auto-generated Mermaid workflow diagrams

## Choose Your Path

### Recommended: Direct vnext API Usage

**Start coding immediately with clean, modern APIs:**

```go
import "github.com/kunalkushwaha/agenticgokit/core/vnext"

// Single agent
agent, _ := vnext.NewChatAgent("bot")
result, _ := agent.Run(ctx, "Hello world")

// Streaming agent  
stream, _ := agent.RunStream(ctx, "Write a story")
for chunk := range stream.Chunks() { /* real-time output */ }

// Multi-agent workflow
workflow, _ := vnext.NewSequentialWorkflow(config)
stream, _ := workflow.RunStream(ctx, input) 
```

### Alternative: Project Scaffolding 

**Generate complete projects with configuration:**

```
my-agent/                 # Generated by agentcli create
├── main.go              # Entry point with vnext APIs
├── agentflow.toml       # Configuration 
├── go.mod               # Dependencies
├── agents/              # Custom agent implementations
└── docs/                # Generated diagrams
```

```toml
# agentflow.toml - Configuration for generated projects
[orchestration]
mode = "sequential" 
timeout_seconds = 300

[llm]
provider = "ollama"
model = "gemma3:1b"

[memory]
provider = "local"
enable_rag = true
```

## Examples & Templates

### Direct vnext Examples
```go
// Basic chat agent
agent, _ := vnext.NewChatAgent("helper")

// With custom configuration  
agent, _ := vnext.NewChatAgent("helper",
    vnext.WithLLM("ollama", "gemma3:1b", "http://localhost:11434"),
    vnext.WithMemory(vnext.EnableRAG()),
    vnext.WithTools("web_search", "calculator"),
)

// Workflow orchestration
workflow, _ := vnext.NewParallelWorkflow(config)
workflow.AddStep(vnext.WorkflowStep{Name: "research", Agent: researchAgent})
workflow.AddStep(vnext.WorkflowStep{Name: "fact-check", Agent: factChecker}) 
```

### CLI Templates (agentcli create)
```bash
agentcli create my-bot --template basic           # Simple chat agent
agentcli create research-team --template research  # Multi-agent research
agentcli create kb-system --template rag-system   # Knowledge base + RAG
agentcli create workflow --template chat-system   # Conversational workflows
```

### Provider Support
```go
// Works with any LLM provider
vnext.WithLLM("openai", "gpt-4", "")                    // OpenAI
vnext.WithLLM("azure", "gpt-4", "https://your.azure")  // Azure OpenAI  
vnext.WithLLM("ollama", "gemma3:1b", "http://localhost:11434") // Local Ollama
```

## Learning Resources

### Working Examples ([`examples/`](examples/))
- **[vnext Streaming Workflow](examples/vnext/streaming_workflow/)** - Real-time multi-agent workflows
- **[Simple Agent](examples/01-simple-agent/)** - Basic agent setup
- **[Multi-Agent Collaboration](examples/02-multi-agent-collab/)** - Team coordination  
- **[RAG Knowledge Base](examples/04-rag-knowledge-base/)** - Memory & retrieval
- **[MCP Tool Integration](examples/05-mcp-agent/)** - Dynamic tool discovery

### Documentation
- **[vnext API Reference](docs/reference/api/vnext/)** - Complete API documentation
- **[Workflow Streaming Guide](core/vnext/STREAMING_GUIDE.md)** - Real-time execution
- **[Migration Guide](core/vnext/MIGRATION_GUIDE.md)** - Upgrade from legacy APIs  
- **[Performance Benchmarks](test/vnext/benchmarks/)** - Overhead analysis

## Development

```bash
# Clone and build
git clone https://github.com/kunalkushwaha/agenticgokit.git
cd agenticgokit
make build

# Run tests
make test

# Run examples
cd examples/01-simple-agent
go run .
```

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
