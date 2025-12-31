# AgenticGoKit Documentation

Welcome to the AgenticGoKit documentation - the production-ready framework for building AI agents in Go.

> **API Version**: This documentation covers the `v1beta` package, which will become `v1` upon stable release.

---

## 🚀 Quick Links

- **[Getting Started](./getting-started.md)** - Start here! Build your first v1beta agent
- **[Installation](./installation.md)** - Setup and installation instructions
- **[Core Concepts](./core-concepts.md)** - Understand agents, handlers, tools, and memory
- **[Streaming](./streaming.md)** - Real-time streaming patterns and chunk types
- **[Workflows](./workflows.md)** - Sequential, Parallel, DAG, Loop, and Subworkflows
- **[Migration Guide](./migration-from-core.md)** - Migrate from core/vnext to v1beta

---

## 📚 Documentation Overview

### Getting Started
Start your journey with v1beta APIs:
- [Installation](./installation.md) - Install v1beta and set up your environment
- [Getting Started](./getting-started.md) - Your first v1beta agent in 5 minutes
- [Core Concepts](./core-concepts.md) - Fundamental concepts and architecture

### Core Features
Explore the main features:
- [Custom Handlers](./custom-handlers.md) - CustomHandlerFunc and AgentHandlerFunc patterns
- [Streaming](./streaming.md) - Real-time streaming with 8 chunk types
- [Workflows](./workflows.md) - Multi-agent orchestration patterns
- [Memory & RAG](./memory-and-rag.md) - Memory integration and retrieval-augmented generation
- [Tool Integration](./tool-integration.md) - Tool registration and MCP support
- [Configuration](./configuration.md) - Builder patterns and functional options

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

## ✨ Why AgenticGoKit?

AgenticGoKit provides a modern, streamlined API with:

### 🎯 Streamlined API Surface
- **8 core methods** (reduced from 30+)
- **Unified RunOptions** for all execution modes
- **Preset builders** for common agent types
- **Functional options** for clean configuration

### ⚡ Built-in Streaming
- **8 chunk types**: Text, Delta, Thought, ToolCall, ToolResult, Metadata, Error, Done
- **Multiple patterns**: Channel-based, callback-based, io.Reader
- **Full lifecycle control** with cancellation and error handling

### 🔄 Multi-Agent Workflows
- **4 workflow types**: Sequential, Parallel, DAG, Loop
- **Subworkflow composition** for nested patterns
- **Context sharing** between agents
- **Step-by-step streaming** with progress tracking

### 💾 Flexible Memory & RAG
- **Multiple backends**: In-memory, PostgreSQL (pgvector), Weaviate
- **RAG support** with configurable weights
- **Session management** and history tracking

### 🛠️ Comprehensive Tooling
- **Tool registration** and discovery
- **MCP integration** for Model Context Protocol
- **Caching** and rate limiting
- **Timeout** and retry handling

---

## 🔄 Migrating from core/vnext?

If you're using the deprecated `core` or `core/vnext` packages:

1. **Read the [Migration Guide](./migration-from-core.md)** - Step-by-step instructions
2. **Check the [API Comparison](./migration-from-core.md#api-comparison)** - Side-by-side examples
3. **Review [Breaking Changes](./migration-from-core.md#breaking-changes)** - What's different
4. **Explore [Examples](./examples/)** - See v1beta in action

### Quick Migration Overview

```go
// ❌ Old (core/vnext - Deprecated)
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
- **OpenAI** - GPT-4, GPT-3.5-turbo, and other OpenAI models
- **Azure AI** - Azure OpenAI Service with your deployments
- **Ollama** - Local models (Llama, Mistral, Gemma, etc.)
- **HuggingFace** - Inference API for HuggingFace models
- **OpenRouter** - Access to multiple LLM providers

See [Installation](./installation.md) for setup instructions for each provider.

---

## 📖 API Reference

For complete API documentation, see:
- **[API Reference](./api-reference.md)** - Full API documentation
- **[GoDoc](https://pkg.go.dev/github.com/agenticgokit/agenticgokit/v1beta)** - Auto-generated reference

---

## 🆘 Need Help?

- **[Troubleshooting Guide](./troubleshooting.md)** - Common issues and solutions
- **[Examples](./examples/)** - Complete, runnable code examples
- **[GitHub Issues](https://github.com/agenticgokit/agenticgokit/issues)** - Report bugs or request features
- **[Discussions](https://github.com/agenticgokit/agenticgokit/discussions)** - Ask questions and share ideas

---

## 🗺️ Documentation Navigation

```
v1beta/
├── README.md (you are here)
├── getting-started.md
├── installation.md
├── core-concepts.md
├── streaming.md
├── workflows.md
├── memory-and-rag.md
├── configuration.md
├── custom-handlers.md
├── tool-integration.md
├── error-handling.md
├── performance.md
├── troubleshooting.md
├── migration-from-core.md
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

## 🚦 Getting Started Checklist

- [ ] [Install v1beta](./installation.md)
- [ ] [Build your first agent](./getting-started.md)
- [ ] [Understand core concepts](./core-concepts.md)
- [ ] [Try streaming](./streaming.md)
- [ ] [Explore workflows](./workflows.md)
- [ ] [Review examples](./examples/)

---

**Ready to build?** Start with the [Getting Started Guide](./getting-started.md) →
