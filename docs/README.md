# AgenticGoKit Documentation

**The Complete Guide to Building AI Agent Systems in Go**

AgenticGoKit is a production-ready Go framework for building intelligent agent workflows with dynamic tool integration, multi-provider LLM support, and enterprise-grade patterns.

## 📚 Documentation Structure

### **🚀 Learning Paths**

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

### **Getting Started**
- **[5-Minute Quickstart](tutorials/getting-started/quickstart.md)** - Get running immediately
- **[Your First Agent](tutorials/getting-started/your-first-agent.md)** - Build a simple agent from scratch
- **[Multi-Agent Collaboration](tutorials/getting-started/multi-agent-collaboration.md)** - Agents working together
- **[Memory & RAG](tutorials/getting-started/memory-and-rag.md)** - Add knowledge capabilities
- **[Tool Integration](tutorials/getting-started/tool-integration.md)** - Connect external tools
- **[Production Deployment](tutorials/getting-started/production-deployment.md)** - Deploy to production

### **Core Concepts**  
- **[Agent Fundamentals](tutorials/core-concepts/agent-lifecycle.md)** - Understanding AgentHandler interface and patterns
- **[Memory & RAG](tutorials/memory-systems/README.md)** - Persistent memory, vector search, and knowledge bases
- **[Multi-Agent Orchestration](tutorials/core-concepts/orchestration-patterns.md)** - Orchestration patterns and API reference
- **[Orchestration Configuration](guides/setup/orchestration-configuration.md)** - Complete guide to configuration-based orchestration
- **[Examples & Tutorials](guides/Examples.md)** - Practical examples and code samples
- **[Tool Integration](tutorials/mcp/README.md)** - MCP protocol and dynamic tool discovery
- **[LLM Providers](guides/setup/llm-providers.md)** - Azure, OpenAI, Ollama, and custom providers
- **[Configuration](reference/api/configuration.md)** - Managing agentflow.toml and environment setup

### **Advanced Usage**
- **[Advanced Patterns](tutorials/advanced/README.md)** - Advanced orchestration patterns and configuration
- **[RAG Configuration](guides/RAGConfiguration.md)** - Retrieval-Augmented Generation setup and tuning
- **[Memory Provider Setup](guides/setup/vector-databases.md)** - PostgreSQL, Weaviate, and in-memory setup guides
- **[Workflow Visualization](guides/development/visualization.md)** - Generate and customize Mermaid diagrams
- **[Production Deployment](guides/deployment/README.md)** - Scaling, monitoring, and best practices  
- **[Error Handling](tutorials/core-concepts/error-handling.md)** - Resilient agent workflows
- **[Custom Tools](guides/CustomTools.md)** - Building your own MCP servers
- **[Performance Tuning](guides/Performance.md)** - Optimization and benchmarking

### **API Reference**
- **[Core Package API](reference/api/agent.md)** - Complete public API reference
- **[Agent Interface](reference/api/agent.md)** - AgentHandler and related types
- **[Memory API](reference/api/agent.md#memory)** - Memory system and RAG APIs
- **[MCP Integration](reference/api/agent.md#mcp)** - Tool discovery and execution APIs
- **[CLI Commands](reference/cli.md)** - agentcli reference

## 🔧 For AgenticGoKit Contributors

**Want to contribute to AgenticGoKit?** See our [Contributor Documentation](contributors/README.md) for:

- **Development Setup** - Getting started with the codebase
- **Architecture Overview** - Understanding the project structure  
- **Contribution Guidelines** - Code style, testing, and documentation standards
- **Development Workflow** - How to submit changes and new features

---

## Quick Start

### Installation
```bash
# Install the CLI
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest

# Prereq: have Ollama running with the gemma3:1b model (recommended for examples)
# On Windows/macOS/Linux: https://ollama.com — then pull the model:
#   ollama pull gemma3:1b

# Create a collaborative multi-agent system (config-driven)
agentcli create research-system \
    --orchestration-mode collaborative \
    --agents 3 \
    --visualize \
    --mcp-enabled

cd research-system

# Configure LLM provider via agentflow.toml (defaults are fine for Ollama gemma3:1b)
# Then run with any message — agents work together in parallel
go run . -m "research AI trends and provide comprehensive analysis"
```

### Multi-Agent Orchestration
```bash
# Sequential processing pipeline
agentcli create data-pipeline \
  --orchestration-mode sequential \
  --agents 3 \
  --orchestration-timeout 45 \
  --visualize

# Loop-based workflow with conditions
agentcli create quality-loop \
  --orchestration-mode loop \
  --agents 1 \
  --max-iterations 5 \
  --orchestration-timeout 120 \
  --visualize

# Mixed collaborative + sequential workflow
agentcli create complex-workflow \
    --orchestration-mode mixed \
    --agents 4 \
    --orchestration-timeout 90 \
    --visualize-output "docs/workflows"
```

### Configuration-Based Orchestration
All generated projects use **configuration-driven orchestration** via `agentflow.toml`:

```toml
[orchestration]
mode = "sequential"                    # sequential, collaborative, loop, mixed, route
timeout_seconds = 30                   # Timeout for orchestration operations
sequential_agents = ["agent1", "agent2", "agent3"]
```

```go
package main

import (
    "context"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

// Example showing config-driven orchestration with Start/Emit/Stop.
func main() {
    // Build runner from agentflow.toml
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil {
        log.Fatal(err)
    }

    // Register your agents (names must match those referenced in config)
    _ = runner.RegisterAgent("agent1", NewAgent1())
    _ = runner.RegisterAgent("agent2", NewAgent2())
    _ = runner.RegisterAgent("agent3", NewAgent3())

    // Start the runner and emit an event
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer runner.Stop()

    // Target a specific agent or broadcast (e.g., "all") depending on mode
    evt := core.NewEvent("agent1", map[string]interface{}{"message": "hello"}, nil)
    if err := runner.Emit(evt); err != nil {
        log.Fatal(err)
    }
}
```

### First Agent
```bash
# Generate a single agent project
agentcli create simple-agent --template basic --visualize

# The generated agent1.go will look like this:
```

```go
package main

import (
    "context"
    "fmt"
    agentflow "github.com/kunalkushwaha/agenticgokit/core"
)

type Agent1Handler struct {
    llm        agentflow.ModelProvider
    mcpManager agentflow.MCPManager
}

func NewAgent1(llm agentflow.ModelProvider, mcp agentflow.MCPManager) *Agent1Handler {
    return &Agent1Handler{llm: llm, mcpManager: mcp}
}

func (a *Agent1Handler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    // Extract user message
    message := event.GetData()["message"]
    
    // Build prompt with available tools
    systemPrompt := "You are a helpful assistant that uses tools when needed."
    toolPrompt := agentflow.FormatToolsForPrompt(ctx, a.mcpManager)
    fullPrompt := fmt.Sprintf("%s\n%s\nUser: %s", systemPrompt, toolPrompt, message)
    
    // Get LLM response
    response, err := a.llm.Generate(ctx, fullPrompt)
    if err != nil {
        return agentflow.AgentResult{}, err
    }
    
    // Execute any tool calls
    toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
    if len(toolResults) > 0 {
        // Synthesize tool results with response
        finalPrompt := fmt.Sprintf("Response: %s\nTool Results: %v\nProvide final answer:", response, toolResults)
        response, _ = a.llm.Generate(ctx, finalPrompt)
    }
    
    // Return result
    state.Set("response", response)
    return agentflow.AgentResult{Result: response, State: state}, nil
}
```

### Multi-Agent
```bash
# Generate a collaborative multi-agent workflow
agentcli create research-system \
  --orchestration-mode collaborative \
  --collaborative-agents "researcher,analyzer,validator" \
  --visualize

# This creates:
# - researcher.go (Research agent - gathers information)
# - analyzer.go (Analysis agent - processes data)  
# - validator.go (Validation agent - ensures quality)
# - main.go (Collaborative orchestration)
# - workflow.mmd (Mermaid diagram)
```

**Collaborative Orchestration Code (config-driven):**
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Build runner from configuration (mode: collaborative)
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil {
        log.Fatal(err)
    }

    // Register agents used by the orchestration
    _ = runner.RegisterAgent("researcher", NewResearcher())
    _ = runner.RegisterAgent("analyzer", NewAnalyzer())
    _ = runner.RegisterAgent("validator", NewValidator())

    // Create event and broadcast to all agents
    event := core.NewEvent("all", map[string]interface{}{
        "task": "research AI trends and provide comprehensive analysis",
    }, nil)

    // Start runner and emit event. Agents process in parallel.
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer runner.Stop()

    if err := runner.Emit(event); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Collaborative workflow started; check logs for agent outputs.")
}
```

## Why AgenticGoKit?

### **For Users:**
- **⚡ Fast Setup**: Working agents in 5 minutes with CLI scaffolding
- **🔧 Tool-Rich**: Dynamic tool discovery via MCP protocol
- **🌐 Provider Agnostic**: Works with any LLM (Azure, OpenAI, Ollama)
- **🏗️ Production Ready**: Built-in error handling, monitoring, scaling patterns

### **For Contributors:**
- **🎯 Clear Architecture**: Separation between core (public API) and internal (implementation)
- **📝 Documentation First**: Every feature documented with examples
- **🧪 Test Coverage**: Comprehensive unit and integration tests
- **🔄 Continuous Integration**: Automated testing and release workflows

---

## 📖 Documentation Sections

### **[📚 Tutorials](tutorials/README.md)**
Learning-oriented guides to help you understand AgenticGoKit:
- **[Getting Started](tutorials/getting-started/README.md)** - Step-by-step beginner tutorials
- **[Core Concepts](tutorials/core-concepts/README.md)** - Fundamental concepts and patterns
- **[Memory Systems](tutorials/memory-systems/README.md)** - RAG and knowledge management
- **[MCP Tools](tutorials/mcp/README.md)** - Tool integration and development
- **[Advanced Patterns](tutorials/advanced/README.md)** - Complex orchestration patterns
- **[Debugging](tutorials/debugging/README.md)** - Debugging and troubleshooting

### **[🛠️ How-To Guides](guides/README.md)**
Task-oriented guides for specific scenarios:
- **[Setup](guides/setup/README.md)** - Configuration and environment setup
- **[Development](guides/development/README.md)** - Development patterns and best practices
- **[Deployment](guides/deployment/README.md)** - Production deployment and scaling

### **[📋 Reference](reference/README.md)**
Information-oriented documentation:
- **[API Reference](reference/README.md)** - Complete API documentation
- **[CLI Reference](reference/cli.md)** - Command-line interface documentation
- **[Configuration Reference](reference/api/configuration.md)** - Configuration options

### **[👥 Contributors](contributors/README.md)**
For developers contributing to AgenticGoKit:
- **[Contributor Guide](contributors/ContributorGuide.md)** - Development setup and workflow
- **[Code Style](contributors/CodeStyle.md)** - Coding standards and conventions
- **[Testing](contributors/Testing.md)** - Testing strategies and guidelines

## Contributing

We welcome contributions! See our [Contributor Guide](contributors/ContributorGuide.md) for details.

```bash
# Quick start for contributors
git clone https://github.com/kunalkushwaha/agenticgokit.git
cd agenticgokit
go mod tidy
go test ./...
```

## Community

- **[GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)** - Q&A and community
- **[Issues](https://github.com/kunalkushwaha/agenticgokit/issues)** - Bug reports and feature requests
- **[Contributing](contributors/ContributorGuide.md)** - How to contribute code and documentation

---

**[⭐ Star us on GitHub](https://github.com/kunalkushwaha/agenticgokit)** | **[📖 Full Documentation](https://agenticgokit.dev)** | **[🚀 Examples](examples/)**
