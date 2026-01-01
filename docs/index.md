# AgenticGoKit Documentation

**The Go Framework for Building Multi-Agent AI Systems**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/agenticgokit/agenticgokit/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/agenticgokit/agenticgokit)](https://goreportcard.com/report/github.com/agenticgokit/agenticgokit)
[![GitHub Stars](https://img.shields.io/github/stars/agenticgokit/agenticgokit?style=social)](https://github.com/agenticgokit/agenticgokit)

Build intelligent agent workflows with dynamic tool integration, multi-provider LLM support, and enterprise-grade orchestration patterns. **Go-native performance meets AI agent systems.**

---

## ⚡ Build Your First Agent

Create a simple chat agent in minutes:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create agent with builder pattern
    agent, err := v1beta.NewBuilder("ChatAgent").
        WithLLM("openai", "gpt-4").
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Run agent
    result, err := agent.Run(context.Background(), "Hello! Tell me about AgenticGoKit.")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Content)
}
```

**[→ Get Started](v1beta/getting-started.md)** • **[→ Installation](v1beta/installation.md)** • **[→ Examples](v1beta/examples/)**

---

## 🚀 Why AgenticGoKit?

<div class="feature-grid">
<div class="feature-card">

### 🏃‍♂️ **For Developers**
- **Streamlined API**: 8 core builder methods (down from 30+)
- **Type Safety**: Compile-time error checking
- **Single Binary**: No complex Python environments
- **Native Concurrency**: True parallel execution with goroutines

</div>
<div class="feature-card">

### 🤖 **For AI Systems**
- **Battle-Tested**: Built from 2+ years of real-world use
- **Memory & RAG**: Built-in vector databases and knowledge management
- **Tool Integration**: MCP protocol for dynamic tool discovery
- **4 Workflow Types**: Sequential, Parallel, DAG, Loop + Subworkflows

</div>
</div>

---

## 🎯 Quick Start Paths

<div class="quickstart-grid">
<div class="quickstart-card">

### 🏃‍♂️ **Quick Start**
Build and run your first agent

```bash
go get github.com/agenticgokit/agenticgokit/v1beta
```

**[→ Start Building](v1beta/getting-started.md)**

</div>
<div class="quickstart-card">

### 🎓 **Learn Concepts**
Understand the architecture and patterns

- [Core Concepts](v1beta/core-concepts.md)
- [Builder Patterns](v1beta/configuration.md)
- [Streaming](v1beta/streaming.md)
- [Workflows](v1beta/workflows.md)

**[→ Learn More](v1beta/README.md)**

</div>
<div class="quickstart-card">

### � **Starting with v1beta**
v1beta is the current production API. Core and vnext packages are deprecated.

- [Getting Started](v1beta/getting-started.md)
- [Core Concepts](v1beta/core-concepts.md)
- [Examples](v1beta/examples/)

**[→ v1beta Documentation](v1beta/README.md)**

</div>
</div>

---

## 📚 Documentation

### **🌟 Start Here**

Modern API designed for building real-world agent systems:

| Guide | Description |
|-------|-------------|
| **[Getting Started](v1beta/getting-started.md)** | Build your first agent |
| **[Core Concepts](v1beta/core-concepts.md)** | Agents, handlers, tools, memory |
| **[Installation](v1beta/installation.md)** | Setup and configuration |
| **[Configuration](v1beta/configuration.md)** | Builder patterns and options |
| **[Workflows](v1beta/workflows.md)** | Sequential, Parallel, DAG, Loop |
| **[Streaming](v1beta/streaming.md)** | Real-time streaming patterns |
| **[Memory & RAG](v1beta/memory-and-rag.md)** | Knowledge integration |
| **[Custom Handlers](v1beta/custom-handlers.md)** | Custom business logic |
| **[Tool Integration](v1beta/tool-integration.md)** | MCP and tool development |
| **[Error Handling](v1beta/error-handling.md)** | Robust error patterns |
| **[Performance](v1beta/performance.md)** | Optimization strategies |
| **[Troubleshooting](v1beta/troubleshooting.md)** | Common issues and solutions |

### **📖 Examples**

Complete, runnable examples:

- **[Basic Agent](v1beta/examples/basic-agent.md)** - Simple chat agent
- **[Streaming Agent](v1beta/examples/streaming-agent.md)** - Real-time responses
- **[Sequential Workflow](v1beta/examples/workflow-sequential.md)** - Step-by-step processing
- **[Parallel Workflow](v1beta/examples/workflow-parallel.md)** - Concurrent execution
- **[DAG Workflow](v1beta/examples/workflow-dag.md)** - Complex dependencies
- **[Loop Workflow](v1beta/examples/workflow-loop.md)** - Iterative processing
- **[Memory & RAG](v1beta/examples/memory-rag.md)** - Knowledge-powered agents
- **[Custom Handlers](v1beta/examples/custom-handlers.md)** - Business logic integration
- **[Subworkflows](v1beta/examples/subworkflow-composition.md)** - Nested workflows

**[→ Browse All Examples](v1beta/examples/)**

---

## 🏗️ What You Can Build

<div class="use-cases-grid">
<div class="use-case-card">

### 🔍 **Research Assistants**
Multi-agent research with web search and analysis

```go
agent, _ := v1beta.NewBuilder("ResearchAgent").
    WithLLM("openai", "gpt-4").
    WithTools(v1beta.WithMCP(webSearchServer)).
    Build()
```

</div>
<div class="use-case-card">

### 📊 **Data Pipelines** 
Sequential workflows with error handling

```go
workflow, _ := v1beta.NewSequentialWorkflow("pipeline",
    v1beta.Step("extract", extractAgent, "Extract data"),
    v1beta.Step("transform", transformAgent, "Transform data"),
    v1beta.Step("load", loadAgent, "Load data"),
)
```

</div>
<div class="use-case-card">

### 💬 **Chat Systems**
Conversational agents with memory

```go
agent, _ := v1beta.NewBuilder("ChatAgent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
        v1beta.WithSessionScoped(),
    ).
    Build()
```

</div>
<div class="use-case-card">

### 📚 **Knowledge Bases**
RAG-powered Q&A systems

```go
agent, _ := v1beta.NewBuilder("QAAgent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithRAG(4000, 0.3, 0.7),
    ).
    Build()
```

</div>
</div>

---

## 🌟 Key Features

### **Highlights**

- **🎯 Simplified API**: 8 core builder methods (was 30+)
- **🔄 4 Workflow Types**: Sequential, Parallel, DAG, Loop
- **🧩 Subworkflows**: Compose complex agent systems
- **📡 Streaming**: Real-time responses with chunking
- **🧠 Memory & RAG**: Built-in vector databases
- **🔧 MCP Tools**: Dynamic tool discovery and integration
- **⚙️ Functional Options**: Clean configuration patterns
- **🎛️ Custom Handlers**: Full control over agent logic
- **❌ Error Handling**: Structured errors with suggestions
- **📊 Performance**: Optimized for production workloads

---

## 📖 Additional Documentation

<div class="docs-grid">
<div class="docs-section">

### 📚 **[Tutorials](tutorials/README.md)**
Step-by-step learning guides:
- **[Getting Started](tutorials/getting-started/README.md)** - Beginner tutorials
- **[Core Concepts](tutorials/core-concepts/README.md)** - Fundamental concepts
- **[Memory Systems](tutorials/memory-systems/README.md)** - RAG and knowledge
- **[MCP Tools](tutorials/mcp/README.md)** - Tool integration
- **[Advanced Patterns](tutorials/advanced/README.md)** - Complex patterns

</div>
<div class="docs-section">

### 🛠️ **[How-To Guides](guides/README.md)**
Task-oriented guides:
- **[Configuration](guides/Configuration.md)** - Setup and config
- **[Memory Setup](guides/MemoryProviderSetup.md)** - Memory providers
- **[Tool Integration](guides/ToolIntegration.md)** - Custom tools
- **[Deployment](guides/deployment/README.md)** - Production deployment

</div>
<div class="docs-section">

### 📋 **[API Reference](reference/README.md)**
Technical documentation:
- **[v1beta API](reference/v1beta-api/README.md)** - Complete v1beta reference
- **[Configuration Reference](reference/api/configuration.md)** - All config options

</div>
<div class="docs-section">

### 👥 **[Contributors](contributors/README.md)**
For contributors:
- **[Contributor Guide](contributors/ContributorGuide.md)** - How to contribute
- **[Code Style](contributors/CodeStyle.md)** - Coding standards
- **[Testing](contributors/Testing.md)** - Testing guidelines

</div>
</div>

---

## 🚀 Installation

### **Quick Install**

```bash
go get github.com/agenticgokit/agenticgokit/v1beta
```

### **Environment Setup**

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Azure OpenAI
export AZURE_OPENAI_API_KEY="your-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export AZURE_OPENAI_DEPLOYMENT="gpt-4"

# Ollama (local)
export OLLAMA_HOST="http://localhost:11434"
```

**[→ Complete Installation Guide](v1beta/installation.md)**

---

## 🔄 Migrating from core/vnext?

The v1beta package is the production-ready API:

```go
// ❌ Old (core/vnext - Deprecated)
import "github.com/agenticgokit/agenticgokit/core/vnext"

agent := vnext.NewBuilder("agent").
    WithConfig(&vnext.Config{...}).
    Build()

// ✅ New (v1beta - Recommended)
import "github.com/agenticgokit/agenticgokit/v1beta"

agent, err := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    Build()
```

**[→ See More Examples](v1beta/examples/)**

---

## 🧠 Core Example

### **Multi-Agent Workflow**

```go
package main

import (
    "context"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create specialized agents
    researcher, _ := v1beta.NewBuilder("Researcher").
        WithLLM("openai", "gpt-4").
        WithTools(v1beta.WithMCP(webSearchServer)).
        Build()

    analyzer, _ := v1beta.NewBuilder("Analyzer").
        WithLLM("openai", "gpt-4").
        Build()

    // Create parallel workflow
    workflow, _ := v1beta.NewParallelWorkflow("Research",
        v1beta.Step("research", researcher, "Research topic"),
        v1beta.Step("analyze", analyzer, "Analyze findings"),
    )

    // Execute workflow
    results, err := workflow.Run(context.Background(), "AI agent frameworks")
    if err != nil {
        log.Fatal(err)
    }

    // Process results
    for step, result := range results {
        log.Printf("%s: %s", step, result.Content)
    }
}
```

**[→ See More Examples](v1beta/examples/)**

---

## 🌍 Community & Support

<div class="community-grid">
<div class="community-card">

### 💬 **Get Help**
- [GitHub Discussions](https://github.com/agenticgokit/agenticgokit/discussions) - Q&A and community
- [GitHub Issues](https://github.com/agenticgokit/agenticgokit/issues) - Bug reports
- [Troubleshooting](v1beta/troubleshooting.md) - Common solutions

</div>
<div class="community-card">

### 🤝 **Contribute**
- [Contributor Guide](contributors/ContributorGuide.md) - How to contribute
- [Good First Issues](https://github.com/agenticgokit/agenticgokit/labels/good%20first%20issue) - Start here
- [Roadmap](ROADMAP.md) - Future plans

</div>
<div class="community-card">

### 📢 **Stay Updated**
- [GitHub Releases](https://github.com/agenticgokit/agenticgokit/releases) - Latest updates
- [Star the Repo](https://github.com/agenticgokit/agenticgokit) - Get notifications
- [Changelog](https://github.com/agenticgokit/agenticgokit/blob/main/CHANGELOG.md) - Version history

</div>
</div>

---

## 🏆 Why Choose AgenticGoKit?

### 🚀 **Go-Native Performance**
Built with Go's strengths in mind - compiled binaries, true concurrency with goroutines, single-binary deployment, and instant startup times. No Python interpreter overhead.

### 🛠️ **Developer-Friendly API**
Streamlined from 30+ methods to 8 core builder methods. Type-safe with compile-time checking. Clean functional options pattern. Comprehensive documentation and examples.

### 🤖 **AI-First Architecture**
Purpose-built for multi-agent systems with 4 workflow types (Sequential, Parallel, DAG, Loop), built-in memory & RAG, MCP tool integration, and nested subworkflow composition.

### 🏭 **Production-Ready**
Structured error handling with recovery strategies, distributed tracing and monitoring, horizontal scalability. Designed for real-world deployment scenarios.

---

## 🚀 Ready to Start?

**[→ Build Your First Agent](v1beta/getting-started.md)** - Get up and running in minutes

**[→ Browse Code Examples](v1beta/examples/)** - See patterns in action

**[→ Migrate from core/vnext](MIGRATION.md)** - Upgrade existing projects

---

**[⭐ Star on GitHub](https://github.com/agenticgokit/agenticgokit)** • **[📖 Documentation](v1beta/README.md)** • **[💬 Community](https://github.com/agenticgokit/agenticgokit/discussions)**

---

## 📜 License

Apache 2.0 - see [LICENSE](https://github.com/agenticgokit/agenticgokit/blob/main/LICENSE)

---

*Build intelligent agents. Ship production systems. All in Go.*
