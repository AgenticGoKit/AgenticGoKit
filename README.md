 ## ⚠️ This project was renamed from `agentflow` to `agenticgokit`. Please update your references. 



# AgenticGoKit
**The Go Framework for Building Multi-Agent AI Systems**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kunalkushwaha/agenticgokit)](https://goreportcard.com/report/github.com/kunalkushwaha/agenticgokit)
[![Documentation](https://img.shields.io/badge/docs-latest-blue)](docs/README.md)

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

**That's it!** Multi-agent collaboration with one CLI command.

---

## 📦 Installation

### **🚀 One-Line Installation (Recommended)**

#### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.ps1 | iex
```

#### Linux/macOS (Bash)
```bash
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.sh | bash
```

### **⚡ Alternative Methods**

<table>
<tr>
<td width="50%">

#### **Go Install**
```bash
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest
```

#### **Specific Version**
```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.sh | bash -s -- --version v0.3.0

# Windows
iwr -useb 'https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.ps1' | iex -Version v0.3.0
```

</td>
<td width="50%">

#### **Manual Download**
1. Go to [Releases](https://github.com/kunalkushwaha/agenticgokit/releases)
2. Download binary for your platform
3. Add to PATH

#### **Build from Source**
```bash
git clone https://github.com/kunalkushwaha/agenticgokit.git
cd agenticgokit
make build
```

</td>
</tr>
</table>

### **🔧 Post-Installation**

```bash
# Verify installation
agentcli version

# Enable shell completion (optional)
# Bash: source <(agentcli completion bash)
# Zsh: agentcli completion zsh > "${fpath[1]}/_agentcli"
# PowerShell: agentcli completion powershell | Out-String | Invoke-Expression

# Create your first project
agentcli create my-project --template basic
```

**📖 [Complete Installation Guide](INSTALL.md)** - Detailed instructions, troubleshooting, and advanced options

---

## 🚀 Why AgenticGoKit?

<table>
<tr>
<td width="50%">

### **🏃‍♂️ For Developers**
- **Go-Native Performance**: Compiled binaries, efficient memory usage
- **Type Safety**: Compile-time error checking prevents runtime issues
- **Simple Deployment**: Single binary, no complex Python environments
- **Native Concurrency**: Goroutines for true parallel agent execution

</td>
<td width="50%">

### **🤖 For AI Systems**
- **Multi-Agent Focus**: Built specifically for agent orchestration
- **Memory & RAG**: Built-in vector databases and knowledge management
- **Tool Integration**: MCP protocol for dynamic tool discovery
- **Production Ready**: Error handling, monitoring, scaling patterns

</td>
</tr>
</table>

**[📊 Complete Framework Comparison](docs/guides/framework-comparison.md)** - Detailed comparison with LangChain, AutoGen, CrewAI, Semantic Kernel + migration guides

---

## 🏗️ What You Can Build

<table>
<tr>
<td width="50%">

### 🔍 **Research Assistants**
Multi-agent research teams with web search, analysis, and synthesis
```bash
agentcli create research-team \
  --orchestration-mode collaborative \
  --agents 3 --mcp-enabled --visualize
```

### 📊 **Data Processing Pipelines** 
Sequential workflows with error handling and monitoring
```bash
agentcli create data-pipeline \
  --orchestration-mode sequential \
  --agents 4 --visualize
```

</td>
<td width="50%">

### 💬 **Conversational Systems**
Chat agents with persistent memory and context
```bash
agentcli create chat-system \
  --agents 2 --visualize
```

### 📚 **Knowledge Bases**
RAG-powered Q&A with document ingestion and vector search
```bash
agentcli create knowledge-base \
  --orchestration-mode collaborative \
  --agents 3 --visualize
```

</td>
</tr>
</table>

---

## 🎯 Quick Start Paths

<table>
<tr>
<td width="33%">

### 🏃‍♂️ **5-Minute Start**
Get your first agent running immediately

```bash
go get github.com/kunalkushwaha/agenticgokit
```

**[→ Start Building](docs/tutorials/getting-started/quickstart.md)**

</td>
<td width="33%">

### 🎓 **Learn Step-by-Step**
Follow guided tutorials to master concepts

- [Your First Agent](docs/tutorials/getting-started/your-first-agent.md)
- [Multi-Agent Collaboration](docs/tutorials/getting-started/multi-agent-collaboration.md)
- [Memory & RAG](docs/tutorials/getting-started/memory-and-rag.md)
- [Tool Integration](docs/tutorials/getting-started/tool-integration.md)

**[→ Start Learning](docs/tutorials/getting-started/README.md)**

</td>
<td width="33%">

### 🚀 **Explore Examples**
Run working examples and demos

```bash
git clone https://github.com/kunalkushwaha/agenticgokit
cd examples/04-rag-knowledge-base
docker-compose up -d
go run main.go
```

**[→ Browse Examples](examples/README.md)**

</td>
</tr>
</table>

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

runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
result, err := runner.ProcessEvent(ctx, event)
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

## ⚙️ Environment Setup

### **LLM Provider Configuration**
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

## 📚 Documentation

### **📖 [Complete Documentation](docs/README.md)**

<table>
<tr>
<td width="50%">

### **🎓 Learning Path**
- **[Getting Started](docs/tutorials/getting-started/README.md)** - Step-by-step tutorials
- **[Core Concepts](docs/tutorials/core-concepts/README.md)** - Fundamental concepts
- **[Memory Systems](docs/tutorials/memory-systems/README.md)** - RAG and knowledge management
- **[MCP Tools](docs/tutorials/mcp/README.md)** - Tool integration
- **[Advanced Patterns](docs/tutorials/advanced/README.md)** - Complex orchestration
- **[Debugging](docs/tutorials/debugging/README.md)** - Troubleshooting

</td>
<td width="50%">

### **🛠️ Practical Guides**
- **[Setup Guides](docs/guides/setup/README.md)** - Configuration and environment
- **[Development](docs/guides/development/README.md)** - Development patterns
- **[Deployment](docs/guides/deployment/README.md)** - Production deployment
- **[Troubleshooting](docs/guides/troubleshooting.md)** - Common issues
- **[Framework Comparison](docs/guides/framework-comparison.md)** - vs LangChain, AutoGen

</td>
</tr>
<tr>
<td width="50%">

### **📋 Reference**
- **[API Reference](docs/reference/README.md)** - Complete API documentation
- **[CLI Reference](docs/reference/cli.md)** - Command-line interface
- **[Configuration](docs/reference/api/configuration.md)** - Configuration options

</td>
<td width="50%">

### **👥 Contributors**
- **[Contributor Guide](docs/contributors/ContributorGuide.md)** - Development setup
- **[Code Style](docs/contributors/CodeStyle.md)** - Coding standards
- **[Testing](docs/contributors/Testing.md)** - Testing strategies

</td>
</tr>
</table>

---

## 🌍 Community & Support

<table>
<tr>
<td width="33%">

### **💬 Get Help**
- [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions) - Q&A and community
- [GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues) - Bug reports and features
- [Troubleshooting Guide](docs/guides/troubleshooting.md) - Common solutions

</td>
<td width="33%">

### **🤝 Contribute**
- [Contributor Guide](docs/contributors/ContributorGuide.md) - How to contribute
- [Good First Issues](https://github.com/kunalkushwaha/agenticgokit/labels/good%20first%20issue) - Start here
- [Roadmap](docs/ROADMAP.md) - Future plans

</td>
<td width="33%">

### **📢 Stay Updated**
- [GitHub Releases](https://github.com/kunalkushwaha/agenticgokit/releases) - Latest updates
- [Star the Repo](https://github.com/kunalkushwaha/agenticgokit) - Get notifications
- [Follow Development](https://github.com/kunalkushwaha/agenticgokit/pulse) - Activity

</td>
</tr>
</table>

---

## 🏆 Why Choose AgenticGoKit?

<table>
<tr>
<td width="50%">

### **🚀 Performance**
- **Compiled Go**: Native performance, efficient memory usage
- **Concurrent Processing**: True parallel agent execution with goroutines
- **Single Binary**: No complex runtime dependencies
- **Fast Startup**: Instant initialization, no warm-up time

</td>
<td width="50%">

### **🛠️ Developer Experience**
- **Type Safety**: Compile-time error checking
- **CLI Scaffolding**: Generate complete projects instantly
- **Configuration-Driven**: Change behavior without code changes
- **Workflow Visualization**: Automatic Mermaid diagrams

</td>
</tr>
<tr>
<td width="50%">

### **🤖 AI-First Design**
- **Multi-Agent Focus**: Built specifically for agent orchestration
- **Memory Integration**: Built-in vector databases and RAG
- **Tool Ecosystem**: MCP protocol for dynamic capabilities
- **Production Patterns**: Error handling, retry logic, monitoring

</td>
<td width="50%">

### **🏭 Production Ready**
- **Error Handling**: Comprehensive error routing and recovery
- **Monitoring**: Built-in logging and tracing capabilities
- **Scalability**: Designed for horizontal scaling patterns
- **Configuration**: Environment-based configuration management

</td>
</tr>
</table>

---

## 🚀 Ready to Build?

<div align="center">

### [🏃‍♂️ **Start with 5-Minute Quickstart**](docs/tutorials/getting-started/quickstart.md)

*Build your first multi-agent system in 5 minutes*

### [🎓 **Follow the Learning Path**](docs/tutorials/getting-started/README.md)

*Master AgenticGoKit with step-by-step tutorials*

### [🚀 **Explore Live Examples**](examples/README.md)

*See working multi-agent systems in action*

---

**[⭐ Star us on GitHub](https://github.com/kunalkushwaha/agenticgokit)** • **[📖 Read the Docs](docs/README.md)** • **[💬 Join Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)**

</div>

---

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.

---

*AgenticGoKit: Where Go performance meets AI agent intelligence.*
