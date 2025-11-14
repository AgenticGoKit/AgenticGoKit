# AgenticGoKit Documentation

**The Complete Guide to Building AI Agent Systems in Go**

AgenticGoKit is a production-ready Go framework for building intelligent agent workflows with dynamic tool integration, multi-provider LLM support, and enterprise-grade patterns.

---

## 🚀 Quick Start

**New to AgenticGoKit?** Start here:

1. **[Build Your First Agent](v1beta/getting-started.md)** - Working agent in 5 minutes
2. **[Core Concepts](v1beta/core-concepts.md)** - Understand the fundamentals
3. **[Examples](v1beta/examples/)** - Real-world code patterns

---

## 📚 v1beta Documentation (Recommended)

**v1beta** is the modern, production-ready API for AgenticGoKit. All new projects should use v1beta.

### **Getting Started**
- **[Getting Started Guide](v1beta/getting-started.md)** - Install and build your first agent
- **[Core Concepts](v1beta/core-concepts.md)** - Agents, workflows, memory, tools
- **[Configuration](v1beta/configuration.md)** - Environment setup and options
- **[Quick Reference](v1beta/quick-reference.md)** - Common patterns and snippets

### **Complete Examples**
Working code for common patterns:

- **[Basic Agent](v1beta/examples/basic-agent.md)** - Simple chat agent
- **[Streaming Agent](v1beta/examples/streaming-agent.md)** - Real-time streaming responses
- **[Sequential Workflow](v1beta/examples/workflow-sequential.md)** - Step-by-step pipeline
- **[Parallel Workflow](v1beta/examples/workflow-parallel.md)** - Concurrent execution
- **[DAG Workflow](v1beta/examples/workflow-dag.md)** - Complex dependencies
- **[Loop Workflow](v1beta/examples/workflow-loop.md)** - Iterative refinement
- **[Subworkflows](v1beta/examples/subworkflow-composition.md)** - Nested workflows
- **[Memory & RAG](v1beta/examples/memory-rag.md)** - Knowledge bases
- **[Custom Handlers](v1beta/examples/custom-handlers.md)** - Custom business logic

### **Feature Guides**
- **[Memory & RAG](v1beta/memory-and-rag.md)** - Memory providers, vector search, knowledge bases
- **[Tool Integration](v1beta/tool-integration.md)** - Adding tools to agents
- **[Custom Handlers](v1beta/custom-handlers.md)** - Implement custom logic
- **[Error Handling](v1beta/error-handling.md)** - Robust error patterns
- **[Performance](v1beta/performance.md)** - Optimization and tuning
- **[Troubleshooting](v1beta/troubleshooting.md)** - Common issues and solutions

### **Migration**
- **[Migration Guide](MIGRATION.md)** - Migrate from legacy APIs to v1beta
- **[API Versioning](API_VERSIONING.md)** - Version strategy and stability

---

## 📖 Learning Paths

### **Beginner Path**
Perfect for developers new to AgenticGoKit:

1. **[Build Your First Agent](v1beta/getting-started.md)** - 5-minute start
2. **[Basic Agent Example](v1beta/examples/basic-agent.md)** - Complete working code
3. **[Core Concepts](v1beta/core-concepts.md)** - Understanding the framework
4. **[Configuration Guide](v1beta/configuration.md)** - Setup and options

### **Intermediate Path**
Build production-ready systems:

1. **[Streaming Agent](v1beta/examples/streaming-agent.md)** - Real-time responses
2. **[Sequential Workflow](v1beta/examples/workflow-sequential.md)** - Multi-step pipelines
3. **[Memory & RAG](v1beta/examples/memory-rag.md)** - Add knowledge capabilities
4. **[Tool Integration](v1beta/tool-integration.md)** - Connect external tools

### **Advanced Path**
Master complex patterns:

1. **[Parallel Workflows](v1beta/examples/workflow-parallel.md)** - Concurrent execution
2. **[DAG Workflows](v1beta/examples/workflow-dag.md)** - Complex dependencies
3. **[Subworkflows](v1beta/examples/subworkflow-composition.md)** - Nested patterns
4. **[Performance Tuning](v1beta/performance.md)** - Optimization strategies


---

## 🔧 Legacy Documentation

The following documentation covers legacy APIs and the `agentcli` tool. **For new projects, use [v1beta documentation](#-v1beta-documentation-recommended) instead.**

<details>
<summary>Click to expand legacy documentation</summary>

### **Getting Started (Legacy)**
- **[5-Minute Quickstart](tutorials/getting-started/quickstart.md)** - Get running immediately
- **[Your First Agent](tutorials/getting-started/your-first-agent.md)** - Build a simple agent from scratch
- **[Multi-Agent Collaboration](tutorials/getting-started/multi-agent-collaboration.md)** - Agents working together
- **[Memory & RAG](tutorials/getting-started/memory-and-rag.md)** - Add knowledge capabilities
- **[Tool Integration](tutorials/getting-started/tool-integration.md)** - Connect external tools
- **[Production Deployment](tutorials/getting-started/production-deployment.md)** - Deploy to production

### **Core Concepts (Legacy)**  
- **[Agent Fundamentals](tutorials/core-concepts/agent-lifecycle.md)** - Understanding AgentHandler interface and patterns
- **[Memory & RAG](tutorials/memory-systems/README.md)** - Persistent memory, vector search, and knowledge bases
- **[Multi-Agent Orchestration](tutorials/core-concepts/orchestration-patterns.md)** - Orchestration patterns and API reference
- **[Orchestration Configuration](guides/setup/orchestration-configuration.md)** - Complete guide to configuration-based orchestration
- **[Examples & Tutorials](guides/Examples.md)** - Practical examples and code samples
- **[Tool Integration](tutorials/mcp/README.md)** - MCP protocol and dynamic tool discovery
- **[LLM Providers](guides/setup/llm-providers.md)** - Azure, OpenAI, Ollama, and custom providers
- **[Configuration](reference/api/configuration.md)** - Managing agentflow.toml and environment setup

### **Advanced Usage (Legacy)**
- **[Advanced Patterns](tutorials/advanced/README.md)** - Advanced orchestration patterns and configuration
- **[RAG Configuration](guides/RAGConfiguration.md)** - Retrieval-Augmented Generation setup and tuning
- **[Memory Provider Setup](guides/setup/vector-databases.md)** - PostgreSQL, Weaviate, and in-memory setup guides
- **[Workflow Visualization](guides/development/visualization.md)** - Generate and customize Mermaid diagrams
- **[Production Deployment](guides/deployment/README.md)** - Scaling, monitoring, and best practices  
- **[Error Handling](tutorials/core-concepts/error-handling.md)** - Resilient agent workflows
- **[Custom Tools](guides/CustomTools.md)** - Building your own MCP servers
- **[Performance Tuning](guides/Performance.md)** - Optimization and benchmarking

### **API Reference (Legacy)**
- **[Core Package API](reference/api/agent.md)** - Complete public API reference
- **[Agent Interface](reference/api/agent.md)** - AgentHandler and related types
- **[Memory API](reference/api/agent.md#memory)** - Memory system and RAG APIs
- **[MCP Integration](reference/api/agent.md#mcp)** - Tool discovery and execution APIs
- **[CLI Commands](reference/cli.md)** - agentcli reference

</details>

---

## 🎯 Use Cases

Build production systems for:

- **Customer Support**: Multi-agent systems with memory and tool integration
- **Content Generation**: Sequential and parallel workflows for content creation
- **Research Systems**: RAG-powered agents with knowledge base access
- - **Data Processing**: ETL pipelines with validation and transformation
- **Code Analysis**: Iterative refinement loops for quality assurance

---

## 💡 Why AgenticGoKit?

### **Go-Native Performance**
Built from the ground up in Go for maximum performance and reliability:
- Goroutine-based concurrency for parallel workflows
- Minimal memory footprint and fast startup times
- Native compilation for deployment anywhere

### **Production-Ready**
Enterprise-grade features out of the box:
- Comprehensive error handling and recovery
- Structured logging and observability hooks
- Memory management and resource cleanup
- Type-safe APIs with full IDE support

### **Multi-Provider Support**
Work with any LLM provider:
- OpenAI (GPT-4, GPT-3.5)
- Azure OpenAI
- Ollama (local models)
- Anthropic Claude
- Custom providers via simple interface

### **Flexible Architecture**
Build exactly what you need:
- Single agents or complex multi-agent systems
- Sequential, parallel, or DAG-based workflows
- Custom handlers for business logic
- Pluggable memory and tool providers

---

## 📦 Installation

```bash
go get github.com/agenticgokit/agenticgokit/v1beta
```

**Requirements:**
- Go 1.21 or higher
- LLM provider API key (OpenAI, Azure, etc.) or local Ollama installation

---

## 🔧 For Contributors

**Want to contribute to AgenticGoKit?** See our [Contributor Documentation](contributors/README.md):

- **[Contributor Guide](contributors/ContributorGuide.md)** - Development setup and workflow
- **[Code Style](contributors/CodeStyle.md)** - Coding standards and conventions
- **[Testing](contributors/Testing.md)** - Testing strategies and guidelines
- **[Adding Features](contributors/AddingFeatures.md)** - Feature development process
- **[Documentation Standards](contributors/DocsStandards.md)** - Writing great docs

```bash
# Quick start for contributors
git clone https://github.com/agenticgokit/agenticgokit.git
cd agenticgokit
go mod tidy
go test ./...
```

---

## 🌐 Community

- **[GitHub Discussions](https://github.com/agenticgokit/agenticgokit/discussions)** - Ask questions and share ideas
- **[Issues](https://github.com/agenticgokit/agenticgokit/issues)** - Report bugs and request features
- **[Contributing](contributors/ContributorGuide.md)** - How to contribute

---

## 📄 Additional Resources

- **[Changelog](../RELEASE.md)** - Release notes and version history
- **[Roadmap](ROADMAP.md)** - Planned features and improvements
- **[Design Documents](design/)** - Architecture and design decisions

---

**[⭐ Star us on GitHub](https://github.com/agenticgokit/agenticgokit)** | **[🚀 Get Started](v1beta/getting-started.md)** | **[📖 Examples](v1beta/examples/)**
