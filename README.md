# AgenticGoKit

> **🚀 BETA RELEASE** - The v1beta API is now stable and recommended for all new projects. While still in beta, the core APIs are working well and ready for testing. We continue to refine features and welcome feedback and contributions!
>
> **📋 API Versioning Plan:**
> - **Current (v0.x)**: `v1beta` package is the recommended API (formerly `vnext`)
> - **v1.0 Release**: `v1beta` will become the primary `v1` package
> - **Legacy APIs**: Both `core` and `core/vnext` packages will be removed in v1.0

**Robust Go framework for building intelligent multi-agent AI systems**

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
- **Batteries Included**: Built-in memory and RAG by default (zero config needed, swappable with pgvector/custom)
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
}
```

> **Note:** The `agentcli` scaffolding tool is being deprecated and will be replaced by the `agk` CLI in a future release.

## Core Capabilities

AgenticGoKit handles the complexities of building AI systems so you can focus on logic.

### 🔄 [Workflow Orchestration](docs/v1beta/workflows.md)
Orchestrate multiple agents using robust patterns. Pass data between agents, handle errors, and manage state automatically.
- **Patterns**: Sequential, Parallel, DAG, Loop.
- **Example**: [Sequential Workflow Demo](examples/sequential-workflow-demo/)

### ⚡ [Real-time Streaming](docs/v1beta/streaming.md)
Built from the ground up for streaming. Receive tokens and tool updates as they happen, suitable for real-time UI experiences.
- **Example**: [Streaming Workflow](examples/streaming_workflow/)

### 🧠 [Memory & RAG](docs/v1beta/memory-and-rag.md)
**Batteries Included**: Agents come with valid memory out-of-the-box (`chromem` embedded vector DB).
- **Features**: Chat history preservation, semantic search, and document ingestion.
- **Configurable**: Swap the default with `pgvector` or custom providers easily.

### 👁️ [Multimodal Input](docs/v1beta/README.md#multimodal-capabilities)
Native support for Images, Audio, and Video inputs. Works seamlessly with models like GPT-4 Vision, Gemini Pro Vision, etc.

### 🛠️ [Tool Integration](docs/v1beta/tool-integration.md)
Extend agents with tools using standard Go functions or the **Model Context Protocol (MCP)** for standardized tool discovery.

## Supported LLM Providers

AgenticGoKit works with all major LLM providers out of the box:

| Provider | Model Examples | Use Case |
|----------|---------------|----------|
| **OpenAI** | GPT-4, GPT-4 Vision, GPT-3.5-turbo | Production-grade conversational and multimodal AI |
| **Azure OpenAI** | GPT-4, GPT-3.5-turbo | Enterprise deployments with Azure |
| **Ollama** | Llama 3, Gemma, Mistral, Phi | Local development and privacy-focused apps |
| **HuggingFace** | Llama-2, Mistral, Falcon | Open-source model experimentation |
| **OpenRouter** | Multiple models | Access to various providers via single API |
| **BentoML** | Any model packaged as Bento | Self-hosted ML models with production features |
| **MLFlow** | Models via MLFlow AI Gateway | ML model deployment and management |
| **vLLM** | Llama-2, Mistral, etc. | High-throughput LLM serving with PagedAttention |
| **Custom** | Any OpenAI-compatible API | Bring your own provider |

## Learning Resources

### 📚 Documentation
- **[Getting Started](docs/v1beta/getting-started.md)** - Build your first agent
- **[API Reference](v1beta/README.md)** - Comprehensive API docs
- **[Memory & RAG](docs/v1beta/memory-and-rag.md)** - Deep dive into memory systems

### 💡 Examples
- **[Story Writer Chat v2](examples/story-writer-chat-v2/)** - Complete Real-time collaborative writing app
- **[Ollama Quickstart](examples/ollama-quickstart/)** - Local LLM development
- **[MCP Integration](examples/mcp-integration/)** - Using Model Context Protocol
- **[HuggingFace Quickstart](examples/huggingface-quickstart/)** - Using HF Inference Endpoints
- **[BentoML Quickstart](examples/bentoml-quickstart/)** - Self-hosted ML models
- **[MLFlow Gateway Demo](examples/mlflow-gateway-demo/)** - MLFlow AI Gateway integration
- **[vLLM Quickstart](examples/vllm-quickstart/)** - High-throughput inference

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
