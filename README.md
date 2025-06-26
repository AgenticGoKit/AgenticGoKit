# AgentFlow

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/kunalkushwaha/agentflow?style=social)](https://github.com/kunalkushwaha/agentflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/kunalkushwaha/agentflow)](https://goreportcard.com/report/github.com/kunalkushwaha/agentflow)

**The Go SDK for building production-ready multi-agent AI systems**

AgentFlow makes it incredibly simple to build, deploy, and scale AI agent workflows in Go. From a single intelligent agent to complex multi-agent orchestrations, AgentFlow provides the tools you need to ship AI applications that work reliably in production.

## What Makes AgentFlow Special?

- **30-Second Setup**: Generate working multi-agent systems with a single CLI command
- **LLM-Driven Tool Discovery**: Agents automatically find and use the right tools via MCP protocol  
- **Production-First**: Built-in error handling, observability, and enterprise patterns
- **Unified API**: One clean interface for all LLM providers and tool integrations
- **Zero Dependencies**: Pure Go with minimal external requirements
- **Developer Experience**: From prototype to production without rewriting code

## Perfect for

> **⚠️ Alpha Stage**: AgentFlow has production-grade features but APIs are rapidly evolving. Use for prototyping and research while we stabilize for production.

**Researchers**: Prototype multi-agent systems with enterprise-grade patterns  
**Developers**: Learn production-ready agent architectures and build proof-of-concepts  
**Experimenters**: Test multi-agent workflows with built-in observability and error handling  
**Early Adopters**: Explore cutting-edge agent frameworks with production features  

## Quick Start (30 seconds)

### 1. Install AgentFlow
```bash
go install github.com/kunalkushwaha/agentflow/cmd/agentcli@latest
```

### 2. Create Your First Multi-Agent System
```bash
# Generate a working project with intelligent agents
agentcli create my-ai-app --agents 2 --provider ollama

cd my-ai-app

# Run with any message - agents will use tools intelligently
go run . -m "search for the latest Go releases and summarize"
```

### 3. See the Magic
```
11:20AM INF MCP Tools discovered agent=agent1 tool_count=3
11:20AM INF Executing LLM-requested tools agent=agent1 tool_calls=1
11:20AM INF Tool execution successful agent=agent1 tool_name=search

=== WORKFLOW RESULTS ===
Based on the latest search results, here are the key Go releases:
- Go 1.23.8 released with improved performance...
- Go 1.24 upcoming features include enhanced generics...
=========================
```

That's it! You have a working multi-agent system that can search the web, process information, and provide intelligent responses.

## Core Concepts

AgentFlow is built around three simple concepts:

### **Agents**
Smart components that process information and make decisions
```go
// Create an agent that can use any LLM
agent, err := core.NewMCPAgent("research-agent", llmProvider)
```

### **Tools** 
External capabilities agents can discover and use via MCP (Model Context Protocol)
```go
// Agents automatically discover tools like web search, databases, APIs
core.QuickStartMCP() // Auto-discovers available tools
```

### **Workflows**
Orchestrated sequences of agents working together
```go
// Chain agents together for complex workflows
agent1 → agent2 → responsible_ai → finalizer
```

## Intelligent Tool Usage

AgentFlow agents don't just follow scripts - they **think** and **decide** which tools to use:

```bash
# Agent analyzes query and chooses appropriate tools
./my-app -m "search for latest Docker tutorials"
# → Agent chooses 'web_search' with query="latest Docker tutorials"

./my-app -m "list running containers"  
# → Agent chooses 'docker' tool with args=["ps"]

./my-app -m "explain quantum computing"
# → Agent uses no tools, provides direct explanation
```

**The LLM decides what tools to use, when to use them, and how to combine results.**

## Core Architecture

AgentFlow's power comes from its layered, event-driven architecture that separates concerns while enabling seamless integration:

```
┌─────────────────────────────────────────────────────┐
│                   Your Application                  │
├─────────────────────────────────────────────────────┤
│    Agent Layer (Multi-Agent Orchestration)          │
│  ┌─────────────┬─────────────┬─────────────────────┐│
│  │   Agent1    │   Agent2    │   ResponsibleAI     ││
│  │ (Research)  │ (Analysis)  │   (Validation)      ││
│  └─────────────┴─────────────┴─────────────────────┘│
├─────────────────────────────────────────────────────┤
│    Workflow Layer (Event-Driven Orchestration)      │
│  ┌─────────────┬─────────────┬─────────────────────┐│
│  │   Runner    │ Orchestrator│   State Manager     ││
│  │ (Execution) │ (Routing)   │   (Memory)          ││
│  └─────────────┴─────────────┴─────────────────────┘│
├─────────────────────────────────────────────────────┤
│    Tool Layer (MCP Integration)                     │
│  ┌─────────────┬─────────────┬─────────────────────┐│
│  │ MCP Manager │ Tool Registry│  Cache & Metrics   ││
│  │(Discovery)  │ (Execution) │  (Performance)      ││
│  └─────────────┴─────────────┴─────────────────────┘│
├─────────────────────────────────────────────────────┤
│    LLM Layer (Provider Abstraction)                 │
│  ┌─────────────┬─────────────┬─────────────────────┐│
│  │   OpenAI    │   Ollama    │     Azure AI        ││
│  │   Adapter   │   Adapter   │     Adapter         ││
│  └─────────────┴─────────────┴─────────────────────┘│
└─────────────────────────────────────────────────────┘
```

### Why This Architecture Matters

#### **1. Event-Driven Foundation**
```go
// Events flow through the system, enabling loose coupling
event := agentflow.NewEvent("research", query, metadata)
runner.Emit(event) // Automatic routing to appropriate agents
```

- **Scalability**: Add agents without changing existing code
- **Reliability**: Built-in error handling and retry mechanisms  
- **Observability**: Every event is tracked and traceable

#### **2. Intelligent Agent Orchestration**
```go
// Agents work together automatically
agent1 → agent2 → responsible_ai → finalizer
```

- **Sequential**: Step-by-step processing (research → analysis → summary)
- **Parallel**: Concurrent processing for speed
- **Conditional**: Smart routing based on content and context

#### **3. MCP-Powered Tool Discovery**
```go
// Tools are discovered dynamically, not hard-coded
core.QuickStartMCP() // Finds all available tools
agent.Run(query)     // LLM chooses which tools to use
```

- **Flexibility**: Connect to any MCP server (web, database, cloud APIs)
- **Intelligence**: LLM decides tool usage based on context
- **Extensibility**: Add new tools without code changes

#### **4. Provider-Agnostic LLM Integration**
```go
// Unified interface for all LLM providers
llm := core.NewOpenAIAdapter(config)    // or Ollama, Azure, etc.
agent := core.NewMCPAgent("agent", llm) // Same interface
```

- **Flexibility**: Switch providers without rewriting agents
- **Testing**: Use mock providers for development
- **Cost Control**: Choose appropriate providers per use case

### Key Design Principles

#### **Composition Over Configuration**
```go
// Build complex agents from simple capabilities
agent := core.NewAgentBuilder("research").
    WithLLM(llmProvider).
    WithMCP().
    WithCache().
    WithMetrics().
    Build()
```

#### **Observable by Default**
```go
// Every operation generates traces and metrics
traces := runner.DumpTrace("session-123")
metrics := agent.GetMetrics()
runner.RegisterCallback(core.HookAfterAgentRun, myCallback)
```

#### **Production-Ready**
```go
// Built-in patterns for enterprise deployment
- Circuit breakers and retries
- Connection pooling and load balancing  
- Input validation and rate limiting
- Health checks and monitoring
```

### **Performance Characteristics**

- **Fast Startup**: Agents initialize in ~3μs
- **Low Memory**: ~5KB per agent average footprint
- **High Throughput**: Handle thousands of concurrent events
- **Horizontal Scaling**: Add agents/servers without bottlenecks

This architecture enables AgentFlow to be simple for beginners yet powerful enough for enterprise production workloads.

## Supported Integrations

### LLM Providers
- **OpenAI** (GPT-4, GPT-3.5)
- **Azure OpenAI** (Enterprise-ready)  
- **Ollama** (Local models)
- **Mock** (Development/testing)

### Tool Ecosystem (via MCP)
- **Web Tools**: Search, scraping, content fetching
- **Development Tools**: Docker, GitHub, code execution
- **Database Tools**: PostgreSQL, MongoDB, Redis
- **Cloud APIs**: AWS, GCP, Azure services
- **Custom Tools**: Build your own MCP servers

### Memory & Storage
- **Vector Databases**: Weaviate, pgvector
- **MCP Tool Caching**: In-memory caching

## Examples & Tutorials

### Quick Examples

**Simple Agent (5 lines)**
```go
package main

import "github.com/kunalkushwaha/agentflow/core"

func main() {
    core.QuickStartMCP()
    agent, _ := core.NewMCPAgent("helper", &MockLLM{})
    state := core.NewState()
    state.Set("query", "help me understand Go interfaces")
    result, _ := agent.Run(context.Background(), state)
}
```

**Multi-Agent Workflow**
```go
// Generated by: agentcli create workflow --agents 3
func main() {
    // Agent1: Research and gather information
    // Agent2: Analyze and process data  
    // Agent3: Generate final response
    runner.Start() // Automatic orchestration
}
```

### Learn More
- **[Getting Started Guide](docs/DevGuide.md)** - Complete tutorial
- **[Architecture Overview](docs/Architecture.md)** - How it works
- **[Production Deployment](docs/TracingGuide.md)** - Enterprise setup
- **[API Reference](docs/MCP_API_Usage_Guide.md)** - Complete API docs

## Use Cases & Success Stories

### **Research & Analysis**
"*AgentFlow provided the perfect foundation for our market research system*"
```bash
agentcli create research-bot --agents 3 --provider openai
# → Generates scaffolding: agent1 → agent2 → responsible_ai → finalizer
# → You implement: search logic, analysis algorithms, data processing
```

### **Customer Support**
"*Cut development time by 70% with AgentFlow's multi-agent scaffolding*"
```bash  
agentcli create support-ai --agents 4 --provider azure
# → Generates scaffolding: sequential workflow with error handling
# → You implement: ticket classification, routing rules, response logic
```

### **Data Processing**
"*AgentFlow handled all the plumbing, we focused on the algorithms*"
```bash
agentcli create data-pipeline --agents 2 --mcp-production
# → Generates scaffolding: MCP-enabled agents with tool discovery
# → You implement: document parsing, processing rules, business logic
```

### **E-commerce Automation**
"*Rapid prototyping with production-ready agent infrastructure*"
```bash
agentcli create inventory-ai --agents 3 --mcp-enabled --with-cache
# → Generates scaffolding: cached multi-agent workflow
# → You implement: inventory algorithms, prediction models, business rules
```

## Developer Experience

### **Beautiful Defaults**
Every generated project works out of the box with sensible configurations, comprehensive logging, and production-ready patterns.

### **Progressive Complexity**
Start simple and add capabilities as you grow:
```bash
# Development - Basic setup
agentcli create myapp-basic --provider mock

# Enhanced - Add MCP tools  
agentcli create myapp-enhanced --mcp-enabled

# Production - Full features
agentcli create myapp-production --mcp-production --with-cache --with-metrics
```

*Note: Each command creates a separate project. To upgrade an existing project, modify the `agentflow.toml` configuration file or copy code between projects.*

### **Built-in Observability**
Every agent comes with tracing, metrics, and debugging tools:
```go
// Automatic session tracking
traces, _ := runner.DumpTrace("session-123")

// Built-in metrics
metrics := agent.GetMetrics()

// Hook into any lifecycle event
runner.RegisterCallback(core.HookAfterAgentRun, myCallback)
```

## Production Features

### �️ **Enterprise Ready**
- **Error Recovery**: Circuit breakers, retries, fallback strategies
- **Security**: Input validation, rate limiting, audit trails
- **Scalability**: Horizontal scaling, load balancing, connection pooling
- **Monitoring**: Health checks, metrics, alerting integration

### **CI/CD Integration**
```yaml
# GitHub Actions example
- name: Test AgentFlow App
  run: |
    go test ./...
    go run . -m "health check" --validate
```

## Join the AgentFlow Community

### **Contributing**
We welcome contributions! Check out our [Contributing Guide](CONTRIBUTING.md) to get started.

```bash
# Quick contribution setup
git clone https://github.com/kunalkushwaha/agentflow.git
cd agentflow
go mod tidy
go test ./...
```

### **Get Help**
- **[Documentation](docs/README.md)** - Complete guides and API reference
- **[GitHub Discussions](https://github.com/kunalkushwaha/agentflow/discussions)** - Community Q&A
- **[Issues](https://github.com/kunalkushwaha/agentflow/issues)** - Bug reports and feature requests
- **[GitHub](https://github.com/kunalkushwaha/agentflow)** - Star the project!

### **Roadmap**
- **Multi-modal agent support**: Enable agents to work with text, images, audio, and other data types
- **Distributed agent clusters**: Scale agent workflows across multiple machines and networks
- **Docker configurations and cloud deployment guides**: Simplified deployment tooling

## **Built by Developer, for Developers**

AgentFlow started as a hobby project for building tooling for AI-enabled applications and became good enough to share with other developers. I'm open-sourcing it because I believe every developer should have access to production-grade AI agent tools.

**Join us in building the future of AI development in Go.**

---

[![GitHub](https://img.shields.io/badge/GitHub-AgentFlow-blue?style=for-the-badge&logo=github)](https://github.com/kunalkushwaha/agentflow)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

**⭐ Star us on GitHub** · **🐛 Report Issues** · **💡 Suggest Features** · **🤝 Contribute**
