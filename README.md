# AgenticGoKit
**The Go Framework for Building Multi-Agent Systems**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kunalkushwaha/agenticgokit)](https://goreportcard.com/report/github.com/kunalkushwaha/agenticgokit)
[![Documentation](https://img.shields.io/badge/docs-agenticgokit.dev-blue)](https://agenticgokit.dev)

Build intelligent agent workflows with dynamic tool integration, multi-provider LLM support, and enterprise-grade orchestration patterns. **Go-native performance for building agentic systems.**

---

## 🚀 Why AgenticGoKit?

| Feature | AgenticGoKit | LangChain | AutoGen |
|---------|-------------|-----------|---------|
| **Language** | Go (compiled, fast) | Python (interpreted) | Python (interpreted) |
| **Performance** | Go-native performance | Moderate | Moderate |
| **Production Ready** | Working toward production | Requires additional setup | Research-focused |
| **Type Safety** | Full compile-time safety | Runtime errors common | Runtime errors common |
| **Deployment** | Single binary | Complex Python environment | Complex Python environment |
| **Concurrency** | Native goroutines | Threading limitations | Threading limitations |
| **Memory Usage** | Lower memory footprint | High memory usage | High memory usage |

**[📊 Complete Framework Comparison](docs/framework-comparison.md)** - Detailed comparison with LangChain, AutoGen, CrewAI, Semantic Kernel, and migration guides

## ⚡ 5-Minute Demo

Create a collaborative multi-agent system using the CLI:

```bash
# Create a multi-agent project with collaborative orchestration
agentcli create research-team --orchestration-mode collaborative \
  --collaborative-agents "researcher,analyzer,writer" \
  --provider openai --agents 3

cd research-team

# Set your API key
export OPENAI_API_KEY=your-key-here

# Run the collaborative system
go run main.go "Research the latest developments in AI agent frameworks"
```

**What you get:**
- ✅ Complete Go project with `main.go`, `agentflow.toml`, and `go.mod`
- ✅ Three specialized agents working in parallel
- ✅ Automatic result synthesis and error handling
- ✅ Production-ready project structure
- ✅ Docker Compose files for databases (if using memory features)

**That's it!** Multi-agent collaboration with one CLI command.

## 🏗️ What You Can Build

### 🔍 **Research Assistants**
Multi-agent research teams with web search, analysis, and synthesis
```bash
agentcli create research-team --orchestration-mode collaborative \
  --collaborative-agents "searcher,analyzer,writer" --mcp-enabled
```

### 📊 **Data Processing Pipelines** 
Sequential workflows with memory, error handling, and monitoring
```bash
agentcli create data-pipeline --orchestration-mode sequential \
  --sequential-agents "collector,processor,validator"
```

### 💬 **Conversational Systems**
Chat agents with persistent memory, context, and multi-turn conversations
```bash
agentcli create chat-system --memory-enabled --memory-provider pgvector \
  --rag-enabled --provider openai --agents 2
```

### 📚 **Knowledge Bases**
RAG-powered Q&A with document ingestion, vector search, and source attribution
```bash
agentcli create knowledge-base --memory-enabled --memory-provider weaviate \
  --rag-enabled --hybrid-search --orchestration-mode collaborative
```

## 🎯 Quick Start Options

<table>
<tr>
<td width="50%">

### 🏃‍♂️ **5-Minute Quickstart**
Get your first agent running immediately
```bash
go get github.com/kunalkushwaha/agenticgokit
```
[→ Start Building](docs/quickstart.md)

</td>
<td width="50%">

### 🎓 **15-Minute Tutorials**
Learn core concepts with hands-on examples
- Multi-Agent Collaboration
- Memory & RAG Systems  
- Tool Integration
- Production Deployment

[→ Learn Step-by-Step](docs/tutorials/)

</td>
</tr>
<tr>
<td width="50%">

### 🚀 **Live Examples**
Run impressive demos with one command
```bash
git clone https://github.com/kunalkushwaha/agenticgokit
cd examples/research-assistant
docker-compose up -d
go run main.go
```
[→ Explore Examples](examples/)

</td>
<td width="50%">

### 🏗️ **Build Something Cool**
Ready to build a real application? Try these CLI commands:

```bash
# Research assistant with web search and analysis
agentcli create research-assistant --orchestration-mode collaborative \
  --collaborative-agents "searcher,analyzer,writer" --mcp-enabled

# Data processing pipeline with error handling  
agentcli create data-pipeline --orchestration-mode sequential \
  --sequential-agents "collector,processor,validator"

# Chat system with persistent memory
agentcli create chat-system --memory-enabled --memory-provider pgvector \
  --rag-enabled --provider openai

# Knowledge base with document ingestion
agentcli create knowledge-base --memory-enabled --memory-provider weaviate \
  --rag-enabled --hybrid-search
```

[→ Explore Examples](examples/)

</td>
</tr>
</table>

## 🧠 Core Concepts

### **Agent Builder Pattern**
```go
// Build agents with fluent interface
agent := core.NewAgent("assistant").
    WithLLM(llmConfig).
    WithMemory(memoryConfig).
    WithMCP(mcpConfig).
    Build()
```

### **Multi-Agent Orchestration**
```go
// Collaborative agents (parallel execution)
runner := core.CreateCollaborativeRunner(agents, timeout)

// Sequential pipeline (step-by-step processing)
runner := core.NewRunnerWithOrchestration(core.EnhancedRunnerConfig{
    OrchestrationMode: core.OrchestrationSequential,
    SequentialAgents: []string{"agent1", "agent2"},
})

// Configuration-based setup
runner, err := core.NewRunnerFromConfig("agentflow.toml", agents)
```

### **Memory & RAG**
```go
// Configure persistent memory with vector search
memory, err := core.NewMemory(core.AgentMemoryConfig{
    Provider: "pgvector",
    Connection: "postgres://localhost/agentdb",
    EnableRAG: true,
    Search: core.SearchConfigToml{HybridSearch: true},
})
```

### **Tool Integration (MCP)**
```go
// Initialize MCP for tool discovery
err := core.InitializeMCP(core.DefaultMCPConfig())

// Create MCP-aware agents
agent, err := core.NewMCPAgent("assistant", provider)
```

## 📊 Performance & Scale

- **🚀 Go-Native Performance**: Compiled binary with efficient memory management
- **⚡ Concurrent Processing**: Native goroutine support for parallel agent execution
- **💾 Memory Efficient**: Lower memory footprint compared to Python frameworks
- **🔄 Error Handling**: Built-in retry logic and error routing capabilities
- **📈 Scalable Architecture**: Designed for horizontal scaling (implementation in progress)

## 🌟 Current Features

- **🤖 Multi-Agent Orchestration**: Collaborative, sequential, loop, and mixed patterns
- **🧠 Memory & RAG**: PostgreSQL pgvector, Weaviate, and in-memory providers
- **🔧 Tool Integration**: MCP protocol support for dynamic tool discovery
- **⚙️ Configuration Management**: TOML-based configuration with environment overrides
- **🎯 Agent Builder**: Fluent interface for composing agent capabilities
- **📊 Basic Monitoring**: Logging and trace capabilities (expanding)

## 🚀 Installation

### Option 1: Go Module (Recommended)
```bash
go mod init my-agent-project
go get github.com/kunalkushwaha/agenticgokit
```

### Option 2: CLI Tool
```bash
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest
agentcli create my-project --orchestration-mode collaborative --agents 3
```



## 📚 Documentation

### **For Developers**
- **[🚀 5-Minute Quickstart](docs/quickstart.md)** - Get running immediately
- **[📊 Framework Comparison](docs/framework-comparison.md)** - Compare with LangChain, AutoGen, CrewAI + migration guides
- **[🎓 Tutorials](docs/tutorials/)** - Step-by-step learning path
- **[💡 How-To Guides](docs/how-to/)** - Task-oriented solutions
- **[📖 API Reference](docs/api/)** - Complete API documentation

### **For Production**
- **[🏭 Deployment Guide](docs/production/deployment.md)** - Docker deployment (coming soon)
- **[📊 Monitoring](docs/production/monitoring.md)** - Observability setup (coming soon)
- **[⚡ Performance](docs/production/performance.md)** - Optimization guide (coming soon)

### **For Contributors**
- **[🤝 Contributing Guide](docs/contributing/)** - How to contribute
- **[🏗️ Architecture](docs/architecture.md)** - System design deep-dive
- **[🧪 Testing](docs/testing.md)** - Testing strategies
- **[📋 Roadmap](ROADMAP.md)** - Future plans

## 🌍 Community & Ecosystem

### **Get Help**
- **[💬 Discord Community](https://discord.gg/dnKWFKgW)** - Real-time chat and support
- **[💡 GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)** - Q&A and ideas
- **[🐛 Issue Tracker](https://github.com/kunalkushwaha/agenticgokit/issues)** - Bug reports and feature requests

## 🏆 Why Developers Choose AgenticGoKit

- **🚀 Go Performance**: Compiled binaries with efficient memory management
- **🔧 Simple Deployment**: Single binary, no complex Python environments
- **🤖 Multi-Agent Focus**: Built specifically for agent orchestration patterns
- **📊 Type Safety**: Compile-time error checking prevents runtime issues
- **🧠 Memory Integration**: Built-in support for vector databases and RAG
- **🔄 Active Development**: Rapidly evolving toward production readiness

## 🚀 Ready to Build?

<div align="center">

### [🏃‍♂️ **Start with 5-Minute Quickstart**](docs/quickstart.md)

*Build your first multi-agent system in 5 minutes*

---

**[⭐ Star us on GitHub](https://github.com/kunalkushwaha/agenticgokit)** • **[📖 Read the Docs](https://agenticgokit.dev)** • **[💬 Join Discord](https://discord.gg/dnKWFKgW)**

</div>

---

## Legacy Documentation

### **Core Concepts (Legacy)**  
- **[Agent Fundamentals](guides/AgentBasics.md)** - Understanding AgentHandler interface and patterns
- **[Memory & RAG](guides/Memory.md)** - Persistent memory, vector search, and knowledge bases
- **[Multi-Agent Orchestration](multi_agent_orchestration.md)** - Orchestration patterns and API reference
- **[Orchestration Configuration](guides/OrchestrationConfiguration.md)** - Complete guide to configuration-based orchestration
- **[Examples & Tutorials](guides/Examples.md)** - Practical examples and code samples
- **[Tool Integration](guides/ToolIntegration.md)** - MCP protocol and dynamic tool discovery
- **[LLM Providers](guides/Providers.md)** - Azure, OpenAI, Ollama, and custom providers
- **[Configuration](guides/Configuration.md)** - Managing agentflow.toml and environment setup

### **Advanced Usage (Legacy)**
- **[⚡ Performance Optimization](docs/guides/Performance.md)** - Speed and efficiency
- **[🛠️ Custom Tools](docs/guides/CustomTools.md)** - Build your own MCP servers

### **For AgentFlow Contributors**
- **[👨‍💻 Contributor Guide](docs/contributors/ContributorGuide.md)** - Development setup
- **[🏗️ Architecture Deep Dive](docs/contributors/CoreVsInternal.md)** - Internal structure
- **[🧪 Testing Strategy](docs/contributors/Testing.md)** - Testing best practices
- **[📝 Code Style](docs/contributors/CodeStyle.md)** - Standards and conventions

### **API Reference**
- **[📖 Core Package](docs/api/core.md)** - Complete public API
- **[🤖 Agent Interface](docs/api/agents.md)** - Agent types and methods
- **[🔧 MCP Integration](docs/api/mcp.md)** - Tool discovery APIs
- **[⌨️ CLI Commands](docs/api/cli.md)** - agentcli reference

### Learn More
- **[📚 Complete Documentation](docs/README.md)** - User guides, API reference, and contributor docs
- **[🚀 Getting Started](docs/guides/AgentBasics.md)** - Build your first agent in 5 minutes
- **[💡 Examples & Tutorials](docs/guides/Examples.md)** - Practical code samples and patterns
- **[🏗️ Architecture Overview](docs/Architecture.md)** - How AgentFlow works under the hood

## License

MIT License - see [LICENSE](LICENSE) for details.
