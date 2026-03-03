# AgenticGoKit Documentation

Welcome to the AgenticGoKit documentation - the production-ready framework for building AI agents in Go.

> **API Version**: This documentation covers the `v1beta` package, which will become `v1` upon stable release.

---

## Quick Links

- **[Getting Started](./getting-started.md)** - Start here! Build your first v1beta agent (includes setup)
- **[Core Concepts](./core-concepts.md)** - Understand agents, handlers, tools, and memory
- **[Streaming](./streaming.md)** - Real-time streaming patterns and chunk types
- **[Workflows](./workflows.md)** - Sequential, Parallel, DAG, Loop, and Subworkflows

---

## Documentation Overview

### Getting Started
Start your journey with v1beta APIs:
- [Getting Started](./getting-started.md) - Your first v1beta agent in 5 minutes (includes setup)
- [Core Concepts](./core-concepts.md) - Fundamental concepts and architecture

### Core Features
Explore the main features:
- [Custom Handlers](./custom-handlers.md) - CustomHandlerFunc and AgentHandlerFunc patterns
- [Streaming](./streaming.md) - Real-time streaming with 13 chunk types
- [Workflows](./workflows.md) - Multi-agent orchestration patterns
- [Memory & RAG](./memory-and-rag.md) - Memory integration and retrieval-augmented generation
- [Tool Integration](./tool-integration.md) - Tool registration and MCP support
- [Observability](./observability.md) - Distributed tracing and structured logging

### Advanced Topics
- [Error Handling](./error-handling.md) - Error patterns and best practices
- [Performance](./performance.md) - Optimization tips and performance tuning
- [Troubleshooting](./troubleshooting.md) - Common issues and solutions

### Examples
Complete, runnable examples:
- [Basic Agent](./examples/basic-agent.md) - Simple agent with preset builder
- [Streaming Agent](./examples/streaming-agent.md) - Real-time streaming example
- [Sequential Workflow](./examples/workflow-sequential.md) - Sequential agent workflow
- [Parallel Workflow](./examples/workflow-parallel.md) - Parallel agent execution
- [DAG Workflow](./examples/workflow-dag.md) - Directed acyclic graph workflow
- [Loop Workflow](./examples/workflow-loop.md) - Iterative loop workflow
- [Subworkflow Composition](./examples/subworkflow-composition.md) - Nested workflow patterns
- [Memory & RAG](./examples/memory-rag.md) - Memory and RAG integration
- [Custom Handlers](./examples/custom-handlers.md) - Custom handler implementations

---

## Why AgenticGoKit?

AgenticGoKit provides a modern, streamlined API with:

### Streamlined API Surface
- **8 core methods** (reduced from 30+)
- **Unified RunOptions** for all execution modes
- **Preset builders** for common agent types
- **Functional options** for clean configuration

### Built-in Streaming
- **13 chunk types**: Content, Delta, Thought, ToolCall, ToolResult, Metadata, Error, Done, AgentStart, AgentComplete, Image, Audio, Video
- **Multiple patterns**: Channel-based, callback-based, io.Reader
- **Full lifecycle control** with cancellation and error handling

### Multi-Agent Workflows
- **4 workflow types**: Sequential, Parallel, DAG, Loop
- **Subworkflow composition** for nested patterns
- **Context sharing** between agents
- **Step-by-step streaming** with progress tracking

### Flexible Memory & RAG
- **Multiple backends**: In-memory, PostgreSQL (pgvector), Weaviate
- **RAG support** with configurable weights
- **Session management** and history tracking

### Production-Ready Observability
- **Distributed tracing** with OpenTelemetry
- **Workflow visibility** for Sequential, Parallel, DAG, Loop patterns
- **Subworkflow tracking** with hierarchical nesting
- **Multiple exporters**: Console, File, OTLP (Jaeger, Tempo)
- **Complete visibility**: Agent execution, LLM calls, tool usage, workflows

### Comprehensive Tooling
- **Tool registration** and discovery
- **MCP integration** for Model Context Protocol
- **Caching** and rate limiting
- **Timeout** and retry handling

---

## � Note: Deprecated Packages

The `core` and `core/vnext` packages are deprecated. New projects should use v1beta. For existing projects, gradual migration is recommended—both versions can coexist in your codebase.

### Quick Start with v1beta

```go
// ✅ New (v1beta - Current)
import "github.com/agenticgokit/agenticgokit/core/vnext"

agent := vnext.NewBuilder("agent").
    WithConfig(&vnext.Config{...}).
    Build()

// ✅ New (v1beta - Recommended)
import "github.com/agenticgokit/agenticgokit/v1beta"

agent, err := v1beta.NewChatAgent("agent",
    v1beta.WithLLM("openai", "gpt-4"),
)
```

---

## 🔌 Supported LLM Providers

AgenticGoKit supports the following LLM providers:
- **OpenAI** (`plugins/llm/openai`) - GPT-4o, GPT-4 Vision, GPT-3.5-turbo
- **Anthropic** (`plugins/llm/anthropic`) - Claude 3.5 Sonnet, Claude 3 Opus, Haiku
- **Azure OpenAI** (`plugins/llm/azureopenai`) - Azure OpenAI Service with your deployments
- **Azure AI Foundry Local** (`plugins/llm/foundrylocal`) - On-device inference via Foundry Local; supports Phi, Qwen, Mistral, Llama and any model loaded locally (no API key required)
- **Ollama** (`plugins/llm/ollama`) - Local models (Llama 3, Mistral, Gemma, Phi, etc.)
- **HuggingFace** (`plugins/llm/huggingface`) - Inference API for HuggingFace models
- **OpenRouter** (`plugins/llm/openrouter`) - Access to 100+ models via a single API key
- **BentoML** (`plugins/llm/bentoml`) - Self-hosted ML models with production features (batching, observability)
- **MLFlow** (`plugins/llm/mlflow`) - Models deployed via MLFlow AI Gateway
- **vLLM** (`plugins/llm/vllm`) - High-throughput LLM serving with PagedAttention optimization

See [Getting Started](./getting-started.md) for setup instructions and [examples/](../../examples/) for provider-specific quickstarts.

---

## 📖 API Reference

For complete API documentation, see:
- **[API Reference](./api-reference.md)** - Full API documentation
- **[GoDoc](https://pkg.go.dev/github.com/agenticgokit/agenticgokit/v1beta)** - Auto-generated reference

---

## Need Help?

- **[Troubleshooting Guide](./troubleshooting.md)** - Common issues and solutions
- **[Examples](./examples/)** - Complete, runnable code examples
- **[GitHub Issues](https://github.com/agenticgokit/agenticgokit/issues)** - Report bugs or request features
- **[Discussions](https://github.com/agenticgokit/agenticgokit/discussions)** - Ask questions and share ideas

---

## Documentation Navigation

```
v1beta/
├── README.md (you are here)
├── getting-started.md
├── core-concepts.md
├── streaming.md
├── workflows.md
├── memory-and-rag.md
├── observability.md
├── configuration.md
├── custom-handlers.md
├── tool-integration.md
├── error-handling.md
├── performance.md
├── troubleshooting.md
├── api-reference.md
└── examples/
    ├── basic-agent.md
    ├── streaming-agent.md
    ├── workflow-sequential.md
    ├── workflow-parallel.md
    ├── workflow-dag.md
    ├── workflow-loop.md
    ├── subworkflow-composition.md
    ├── memory-rag.md
    └── custom-handlers.md
```

---

## Getting Started Checklist

- [ ] [Install and setup v1beta](./getting-started.md#installation)
- [ ] [Build your first agent](./getting-started.md)
- [ ] [Understand core concepts](./core-concepts.md)
- [ ] [Try streaming](./streaming.md)
- [ ] [Explore workflows](./workflows.md)
- [ ] [Review examples](./examples/)

---

**Ready to build?** Start with the [Getting Started Guide](./getting-started.md) →
