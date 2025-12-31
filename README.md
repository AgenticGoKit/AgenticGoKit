# AgenticGoKit

> **🚀 BETA RELEASE** - The v1beta API is now stable and recommended for all new projects. While still in beta, the core APIs are working well and ready for testing. We continue to refine features and welcome feedback and contributions!
>
> **📋 API Versioning Plan:**
> - **Current (v0.x)**: `v1beta` package is the recommended API (formerly `vnext`)
> - **v1.0 Release**: `v1beta` will become the primary `v1` package
> - **Legacy APIs**: Both `core` and `core/vnext` packages will be removed in v1.0

**Production-ready Go framework for building intelligent multi-agent AI systems**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kunalkushwaha/agenticgokit)](https://goreportcard.com/report/github.com/kunalkushwaha/agenticgokit)
[![Build Status](https://github.com/kunalkushwaha/agenticgokit/workflows/CI/badge.svg)](https://github.com/kunalkushwaha/agenticgokit/actions)
[![Documentation](https://img.shields.io/badge/docs-latest-blue)](docs/README.md)

**The most productive way to build AI agents in Go.** AgenticGoKit provides a unified, streaming-first API for creating intelligent agents with built-in workflow orchestration, tool integration, and memory management. Start with simple single agents and scale to complex multi-agent workflows.

## Why Choose AgenticGoKit?

- **v1beta APIs**: Modern, streaming-first agent interface with comprehensive error handling
- **Multimodal Support**: Native support for images, audio, and video inputs alongside text
- **Real-time Streaming**: Watch your agents think and respond in real-time  
- **Multi-Agent Workflows**: Sequential, parallel, DAG, and loop orchestration patterns
- **Multiple LLM Providers**: Seamlessly switch between OpenAI, Ollama, Azure OpenAI, HuggingFace, and more
- **High Performance**: Compiled Go binaries with minimal overhead
- **Rich Integrations**: Memory providers, tool discovery, MCP protocol support
- **Active Development**: Beta status with stable core APIs and ongoing improvements

---

## Quick Start

**Start building immediately with the modern v1beta API:**

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
    // Create a chat agent with Ollama
    agent, err := v1beta.NewBuilder("ChatAgent").
        WithConfig(&v1beta.Config{
            Name:         "ChatAgent",
            SystemPrompt: "You are a helpful assistant",
            Timeout:      30 * time.Second,
            LLM: v1beta.LLMConfig{
                Provider: "ollama",
                Model:    "gemma3:1b",
                BaseURL:  "http://localhost:11434",
            },
        }).
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
    researcher, _ := v1beta.NewBuilder("researcher").
        WithConfig(&v1beta.Config{
            Name:         "researcher",
            SystemPrompt: "You are a research specialist",
            Timeout:      60 * time.Second,
            LLM: v1beta.LLMConfig{
                Provider: "ollama",
                Model:    "gemma3:1b",
                BaseURL:  "http://localhost:11434",
            },
        }).
        Build()
    
    analyzer, _ := v1beta.NewBuilder("analyzer").
        WithConfig(&v1beta.Config{
            Name:         "analyzer",
            SystemPrompt: "You are a data analyst",
            Timeout:      60 * time.Second,
            LLM: v1beta.LLMConfig{
                Provider: "ollama",
                Model:    "gemma3:1b",
                BaseURL:  "http://localhost:11434",
            },
        }).
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
- **Multimodal Input**: Support for images, audio, and video alongside text
- **Memory & RAG**: Built-in persistence, retrieval, and direct memory access (**chromem enabled by default**)
- **Tool Integration**: MCP protocol, function calling, tool discovery
- **Subworkflows**: Compose workflows as agents for complex hierarchies
- **Multiple LLM Providers**: OpenAI, Azure OpenAI, Ollama, HuggingFace, OpenRouter, and custom providers
- **Flexible Configuration**: Builder pattern with type-safe options

### Multimodal Capabilities

AgenticGoKit provides native support for multimodal inputs, allowing your agents to process images, audio, and video alongside text:

```go
// Create options with multimodal input
opts := v1beta.NewRunOptions()
opts.Images = []v1beta.ImageData{
    {
        URL:      "https://example.com/image.jpg",
        Metadata: map[string]interface{}{"source": "web"},
    },
    {
        Base64:   base64EncodedImageData,
        Metadata: map[string]interface{}{"type": "screenshot"},
    },
}
opts.Audio = []v1beta.AudioData{
    {
        URL:      "https://example.com/audio.mp3",
        Format:   "mp3",
        Metadata: map[string]interface{}{"duration": "30s"},
    },
}
opts.Video = []v1beta.VideoData{
    {
        Base64:   base64EncodedVideoData,
        Format:   "mp4",
        Metadata: map[string]interface{}{"resolution": "1080p"},
    },
}

// Run agent with multimodal input
result, err := agent.RunWithOptions(ctx, "Describe this image and summarize the audio", opts)
```

**Supported Modalities:**
- **Images**: JPG, PNG, GIF (via URL or Base64)
- **Audio**: MP3, WAV, OGG (via URL or Base64)
- **Video**: MP4, WebM (via URL or Base64)

**Compatible Providers:** OpenAI GPT-4 Vision, Gemini Pro Vision, and other multimodal LLMs

### Supported LLM Providers

AgenticGoKit works with all major LLM providers out of the box:

| Provider | Model Examples | Use Case |
|----------|---------------|----------|
| **OpenAI** | GPT-4, GPT-4 Vision, GPT-3.5-turbo | Production-grade conversational and multimodal AI |
| **Azure OpenAI** | GPT-4, GPT-3.5-turbo | Enterprise deployments with Azure |
| **Ollama** | Llama 3, Gemma, Mistral, Phi | Local development and privacy-focused apps |
| **HuggingFace** | Llama-2, Mistral, Falcon | Open-source model experimentation |
| **OpenRouter** | Multiple models | Access to various providers via single API |
| **Custom** | Any OpenAI-compatible API | Bring your own provider |

Switch providers with a simple configuration change—no code modifications required.

## API Usage

**v1beta provides clean, modern APIs for building AI agents:**

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

// Single agent with configuration
agent, _ := v1beta.NewBuilder("MyAgent").
    WithConfig(&v1beta.Config{
        Name: "MyAgent",
        LLM:  v1beta.LLMConfig{Provider: "ollama", Model: "gemma3:1b"},
    }).
    Build()
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
// Basic chat agent using preset
agent, _ := v1beta.NewBuilder("helper").
    WithPreset(v1beta.ChatAgent).
    Build()

// With custom configuration and RAG
agent, _ := v1beta.NewBuilder("helper").
    WithConfig(&v1beta.Config{
        Name:         "helper",
        SystemPrompt: "You are a helpful assistant",
        LLM: v1beta.LLMConfig{
            Provider: "ollama",
            Model:    "gemma3:1b",
            BaseURL:  "http://localhost:11434",
        },
        // Memory is ENABLED by default using chromem!
        // You only need to add config if you want to customize it or use RAG settings.
    }).
    WithMemory(
        v1beta.WithRAG(4000, 0.3, 0.7),
    ).
    WithTools(
        v1beta.WithMCPDiscovery(), // Enable MCP auto-discovery
    ).
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
agent, _ := v1beta.NewBuilder("OpenAIAgent").
    WithConfig(&v1beta.Config{
        Name: "OpenAIAgent",
        LLM: v1beta.LLMConfig{
            Provider: "openai",
            Model:    "gpt-4",
            APIKey:   os.Getenv("OPENAI_API_KEY"),
        },
    }).
    Build()

// Azure OpenAI
agent, _ := v1beta.NewBuilder("AzureAgent").
    WithConfig(&v1beta.Config{
        Name: "AzureAgent",
        LLM: v1beta.LLMConfig{
            Provider: "azure",
            Model:    "gpt-4",
            BaseURL:  "https://your-resource.openai.azure.com",
            APIKey:   os.Getenv("AZURE_OPENAI_API_KEY"),
        },
    }).
    Build()

// Ollama (Local)
agent, _ := v1beta.NewBuilder("OllamaAgent").
    WithConfig(&v1beta.Config{
        Name: "OllamaAgent",
        LLM: v1beta.LLMConfig{
            Provider: "ollama",
            Model:    "gemma3:1b",
            BaseURL:  "http://localhost:11434",
        },
    }).
    Build()

// HuggingFace
agent, _ := v1beta.NewBuilder("HFAgent").
    WithConfig(&v1beta.Config{
        Name: "HFAgent",
        LLM: v1beta.LLMConfig{
            Provider: "huggingface",
            Model:    "meta-llama/Llama-2-7b-chat-hf",
            APIKey:   os.Getenv("HUGGINGFACE_API_KEY"),
        },
    }).
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

### Current Status (v0.x - Beta)

- **Recommended**: Use `v1beta` package for all new projects
- **Import Path**: `github.com/agenticgokit/agenticgokit/v1beta`
- **Stability**: Beta - Core APIs are stable and functional, suitable for testing and development
- **Status**: APIs may evolve based on feedback before v1.0 release
- **Note**: `v1beta` is the evolution of the former `core/vnext` package

### v1.0 Release Plan

**What's Changing:**
- `v1beta` package will become the primary `v1` API
- Legacy `core` and `core/vnext` packages will be **removed entirely**
- Clean, stable API with semantic versioning guarantees

**Migration Path:**
- If you're using `v1beta` or `vnext`: Minimal changes (import path update only)
- If you're using `core`: Migrate to `v1beta` now to prepare
- **`core/vnext` users**: `vnext` has been renamed to `v1beta` - update imports

**Timeline:**
- v0.x (Current): `v1beta` stabilization and testing
- v1.0 (Planned): `v1beta` → `v1`, remove `core` package

### Why v1beta Now?

The `v1beta` package represents our next-generation API design:
- ✅ Streaming-first architecture
- ✅ Unified builder pattern
- ✅ Better error handling
- ✅ Workflow composition
- ✅ Stable core APIs (beta status)
- ⚠️ Minor changes possible before v1.0

By using `v1beta` today, you're getting access to the latest features and helping shape the v1.0 release with your feedback.

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
