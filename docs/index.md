# AgenticGoKit Documentation

**The Go Framework for Building Multi-Agent AI Systems**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kunalkushwaha/agenticgokit/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kunalkushwaha/agenticgokit)](https://goreportcard.com/report/github.com/kunalkushwaha/agenticgokit)
[![GitHub Stars](https://img.shields.io/github/stars/kunalkushwaha/agenticgokit?style=social)](https://github.com/kunalkushwaha/agenticgokit)

Build intelligent agent workflows with dynamic tool integration, multi-provider LLM support, and enterprise-grade orchestration patterns. **Go-native performance meets AI agent systems.**

---

## ⚡ 5-Minute Demo

Create a collaborative multi-agent system with one command:

```bash
# Install the CLI
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest

# Create a multi-agent research team
agentcli create research-team --template research-assistant --visualize

cd research-team

# Set your API key
export AZURE_OPENAI_API_KEY=your-key-here
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
export AZURE_OPENAI_DEPLOYMENT=your-deployment-name

# Run the collaborative system
go run . -m "Research the latest developments in AI agent frameworks"
```

**What you get:**
- ✅ Complete Go project with `main.go`, `agentflow.toml`, and `go.mod`
- ✅ Three specialized agents working in parallel
- ✅ Automatic result synthesis and error handling
- ✅ Mermaid workflow diagrams generated
- ✅ Production-ready project structure

---

## 🚀 Why AgenticGoKit?

<div class="feature-grid">
<div class="feature-card">

### 🏃‍♂️ **For Developers**
- **Go-Native Performance**: Compiled binaries, efficient memory usage
- **Type Safety**: Compile-time error checking prevents runtime issues
- **Simple Deployment**: Single binary, no complex Python environments
- **Native Concurrency**: Goroutines for true parallel agent execution

</div>
<div class="feature-card">

### 🤖 **For AI Systems**
- **Multi-Agent Focus**: Built specifically for agent orchestration
- **Memory & RAG**: Built-in vector databases and knowledge management
- **Tool Integration**: MCP protocol for dynamic tool discovery
- **Production Ready**: Error handling, monitoring, scaling patterns

</div>
</div>

---

## 🎯 Quick Start Paths

<div class="quickstart-grid">
<div class="quickstart-card">

### 🏃‍♂️ **5-Minute Start**
Get your first agent running immediately

```bash
go get github.com/kunalkushwaha/agenticgokit
```

**[→ Start Building](tutorials/getting-started/quickstart.md)**

</div>
<div class="quickstart-card">

### 🎓 **Learn Step-by-Step**
Follow guided tutorials to master concepts

- [Your First Agent](tutorials/getting-started/your-first-agent.md)
- [Multi-Agent Collaboration](tutorials/getting-started/multi-agent-collaboration.md)
- [Memory & RAG](tutorials/getting-started/memory-and-rag.md)
- [Tool Integration](tutorials/getting-started/tool-integration.md)

**[→ Start Learning](tutorials/getting-started/README.md)**

</div>
<div class="quickstart-card">

### 🚀 **Explore Examples**
Run working examples and demos

```bash
git clone https://github.com/kunalkushwaha/agenticgokit
cd examples/04-rag-knowledge-base
docker-compose up -d
go run main.go
```

**[→ Browse Examples](https://github.com/kunalkushwaha/agenticgokit/tree/main/examples)**

</div>
</div>

---

## 🏗️ What You Can Build

<div class="use-cases-grid">
<div class="use-case-card">

### 🔍 **Research Assistants**
Multi-agent research teams with web search, analysis, and synthesis
```bash
agentcli create research-team --template research-assistant
```

</div>
<div class="use-case-card">

### 📊 **Data Processing Pipelines** 
Sequential workflows with error handling and monitoring
```bash
agentcli create data-pipeline --template data-pipeline --visualize
```

</div>
<div class="use-case-card">

### 💬 **Conversational Systems**
Chat agents with persistent memory and context
```bash
agentcli create chat-system --template chat-system
```

</div>
<div class="use-case-card">

### 📚 **Knowledge Bases**
RAG-powered Q&A with document ingestion and vector search
```bash
agentcli create knowledge-base --template rag-system
```

</div>
</div>

---

## 📚 Documentation Structure

### 🚀 **Learning Paths**

**New to AgenticGoKit?** Follow these guided paths:

#### **Beginner Path** (30 minutes)
1. [5-Minute Quickstart](tutorials/getting-started/quickstart.md) - Get running immediately
2. [Your First Agent](tutorials/getting-started/your-first-agent.md) - Build a simple agent
3. [Multi-Agent Collaboration](tutorials/getting-started/multi-agent-collaboration.md) - Agents working together

#### **Intermediate Path** (1 hour)
1. [Memory & RAG](tutorials/getting-started/memory-and-rag.md) - Add knowledge capabilities
2. [Tool Integration](tutorials/getting-started/tool-integration.md) - Connect external tools
3. [Core Concepts](tutorials/core-concepts/README.md) - Deep dive into fundamentals

#### **Advanced Path** (2+ hours)
1. [Advanced Patterns](tutorials/advanced/README.md) - Complex orchestration patterns
2. [Production Deployment](tutorials/getting-started/production-deployment.md) - Deploy to production
3. [Performance Optimization](tutorials/advanced/load-balancing-scaling.md) - Scale your systems

---

## 📖 Documentation Sections

<div class="docs-grid">
<div class="docs-section">

### 📚 **[Tutorials](tutorials/README.md)**
Learning-oriented guides to help you understand AgenticGoKit:
- **[Getting Started](tutorials/getting-started/README.md)** - Step-by-step beginner tutorials
- **[Core Concepts](tutorials/core-concepts/README.md)** - Fundamental concepts and patterns
- **[Memory Systems](tutorials/memory-systems/README.md)** - RAG and knowledge management
- **[MCP Tools](tutorials/mcp/README.md)** - Tool integration and development
- **[Advanced Patterns](tutorials/advanced/README.md)** - Complex orchestration patterns
- **[Debugging](tutorials/debugging/README.md)** - Debugging and troubleshooting

</div>
<div class="docs-section">

### 🛠️ **[How-To Guides](guides/README.md)**
Task-oriented guides for specific scenarios:
- **[Setup](guides/setup/README.md)** - Configuration and environment setup
- **[Development](guides/development/README.md)** - Development patterns and best practices
- **[Deployment](guides/deployment/README.md)** - Production deployment and scaling
- **[Framework Comparison](guides/framework-comparison.md)** - vs LangChain, AutoGen, CrewAI

</div>
<div class="docs-section">

### 📋 **[Reference](reference/README.md)**
Information-oriented documentation:
- **[API Reference](README.md)** - Complete API documentation
- **[CLI Reference](reference/cli.md)** - Command-line interface documentation
- **[Configuration Reference](reference/api/configuration.md)** - Configuration options

</div>
<div class="docs-section">

### 👥 **[Contributors](contributors/README.md)**
For developers contributing to AgenticGoKit:
- **[Contributor Guide](contributors/ContributorGuide.md)** - Development setup and workflow
- **[Code Style](contributors/CodeStyle.md)** - Coding standards and conventions
- **[Testing](contributors/Testing.md)** - Testing strategies and guidelines

</div>
</div>

---

## 🧠 Core Concepts

### **Multi-Agent Orchestration**
```go
// Collaborative agents (parallel execution)
agents := map[string]core.AgentHandler{
    "researcher": NewResearchAgent(),
    "analyzer":   NewAnalysisAgent(),
    "validator":  NewValidationAgent(),
}

runner, _ := core.NewRunnerFromConfig("agentflow.toml")
_ = runner.Start(context.Background())
defer runner.Stop()
_ = runner.Emit(core.NewEvent("all", map[string]any{"task": "analyze"}, nil))
```

### **Configuration-Based Setup**
```toml
# agentflow.toml
[orchestration]
mode = "collaborative"
timeout_seconds = 30

[agent_memory]
provider = "pgvector"
enable_rag = true
chunk_size = 1000

[mcp]
enabled = true
```

### **Memory & RAG Integration**
```go
// Configure persistent memory with vector search
memory, err := core.NewMemory(core.AgentMemoryConfig{
    Provider: "pgvector",
    EnableRAG: true,
    EnableKnowledgeBase: true,
    ChunkSize: 1000,
})
```

### **Tool Integration (MCP)**
```go
// MCP tools are automatically discovered and integrated
// Agents can use web search, file operations, and custom tools
agent := agents.NewToolEnabledAgent("assistant", llmProvider, toolManager)
```

---

## 🌟 Current Features

- **🤖 Multi-Agent Orchestration**: Collaborative, sequential, loop, and mixed patterns
- **🧠 Memory & RAG**: PostgreSQL pgvector, Weaviate, and in-memory providers  
- **🔧 Tool Integration**: MCP protocol support for dynamic tool discovery
- **⚙️ Configuration Management**: TOML-based configuration with environment overrides
- **📊 Workflow Visualization**: Automatic Mermaid diagram generation
- **🎯 CLI Scaffolding**: Generate complete projects with one command
- **📈 Production Patterns**: Error handling, retry logic, and monitoring hooks

---

## 🚀 Installation & Setup

### **Option 1: CLI Tool (Recommended)**
```bash
# Install the CLI
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest

# Create your first project
agentcli create my-agents --template research-assistant --visualize

cd my-agents
```

### **Option 2: Go Module**
```bash
go mod init my-agent-project
go get github.com/kunalkushwaha/agenticgokit

# Create agentflow.toml configuration file
# See reference/api/configuration.md for details
```

### **Environment Setup**
```bash
# For Azure OpenAI (recommended)
export AZURE_OPENAI_API_KEY=your-key-here
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
export AZURE_OPENAI_DEPLOYMENT=your-deployment-name

# For OpenAI
export OPENAI_API_KEY=your-key-here

# For Ollama (local)
export OLLAMA_HOST=http://localhost:11434
```

---

## 🌍 Community & Support

<div class="community-grid">
<div class="community-card">

### 💬 **Get Help**
- [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions) - Q&A and community
- [GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues) - Bug reports and features
- [Troubleshooting Guide](guides/troubleshooting.md) - Common solutions

</div>
<div class="community-card">

### 🤝 **Contribute**
- [Contributor Guide](contributors/ContributorGuide.md) - How to contribute
- [Good First Issues](https://github.com/kunalkushwaha/agenticgokit/labels/good%20first%20issue) - Start here
- [Roadmap](https://github.com/kunalkushwaha/agenticgokit/projects) - Future plans

</div>
<div class="community-card">

### 📢 **Stay Updated**
- [GitHub Releases](https://github.com/kunalkushwaha/agenticgokit/releases) - Latest updates
- [Star the Repo](https://github.com/kunalkushwaha/agenticgokit) - Get notifications
- [Follow Development](https://github.com/kunalkushwaha/agenticgokit/pulse) - Activity

</div>
</div>

---

## 🏆 Why Choose AgenticGoKit?

<div class="benefits-grid">
<div class="benefit-card">

### 🚀 **Performance**
- **Compiled Go**: Native performance, efficient memory usage
- **Concurrent Processing**: True parallel agent execution with goroutines
- **Single Binary**: No complex runtime dependencies
- **Fast Startup**: Instant initialization, no warm-up time

</div>
<div class="benefit-card">

### 🛠️ **Developer Experience**
- **Type Safety**: Compile-time error checking
- **CLI Scaffolding**: Generate complete projects instantly
- **Configuration-Driven**: Change behavior without code changes
- **Workflow Visualization**: Automatic Mermaid diagrams

</div>
<div class="benefit-card">

### 🤖 **AI-First Design**
- **Multi-Agent Focus**: Built specifically for agent orchestration
- **Memory Integration**: Built-in vector databases and RAG
- **Tool Ecosystem**: MCP protocol for dynamic capabilities
- **Production Patterns**: Error handling, retry logic, monitoring

</div>
<div class="benefit-card">

### 🏭 **Production Ready**
- **Error Handling**: Comprehensive error routing and recovery
- **Monitoring**: Built-in logging and tracing capabilities
- **Scalability**: Designed for horizontal scaling patterns
- **Configuration**: Environment-based configuration management

</div>
</div>

---

## 🚀 Ready to Build?

<div class="cta-section">

### [🏃‍♂️ **Start with 5-Minute Quickstart**](tutorials/getting-started/quickstart.md)

*Build your first multi-agent system in 5 minutes*

### [🎓 **Follow the Learning Path**](tutorials/getting-started/README.md)

*Master AgenticGoKit with step-by-step tutorials*

### [🚀 **Explore Live Examples**](https://github.com/kunalkushwaha/agenticgokit/tree/main/examples)

*See working multi-agent systems in action*

---

**[⭐ Star us on GitHub](https://github.com/kunalkushwaha/agenticgokit)** • **[📖 Read the Docs](README.md)** • **[💬 Join Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)**

</div>

---

## License

Apache 2.0 - see [LICENSE](https://github.com/kunalkushwaha/agenticgokit/blob/main/LICENSE) for details.

---

*AgenticGoKit: Where Go performance meets AI agent intelligence.*
