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

# Create a collaborative multi-agent system
agentcli create research-system \
  --orchestration-mode collaborative \
  --agents 3 \
  --visualize \
  --mcp-enabled

cd research-system

# Set your API key
export AZURE_OPENAI_API_KEY=your-key-here
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
export AZURE_OPENAI_DEPLOYMENT=your-deployment-name

# Run with any message - agents work together intelligently
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
  --orchestration-mode collaborative \
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
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Load configuration from agentflow.toml
    config, err := core.LoadConfigFromWorkingDir()
    if err != nil {
        log.Fatal(err)
    }

    // Create provider from configuration
    provider, err := config.InitializeProvider()
    if err != nil {
        log.Fatal(err)
    }

    // Create agents
    agents := map[string]core.AgentHandler{
        "agent1": core.NewLLMAgent("agent1", provider),
        "agent2": core.NewLLMAgent("agent2", provider),
        "agent3": core.NewLLMAgent("agent3", provider),
    }

    // Create runner based on configuration
    var runner core.Runner
    switch config.Orchestration.Mode {
    case "sequential":
        runner = core.NewOrchestrationBuilder(core.OrchestrationSequential).
            WithAgents(agents).
            WithTimeout(time.Duration(config.Orchestration.TimeoutSeconds)*time.Second).
            Build()
    case "collaborative":
        runner = core.CreateCollaborativeRunner(agents, 
            time.Duration(config.Orchestration.TimeoutSeconds)*time.Second)
    default:
        runner = core.CreateRouteRunner(agents)
    }

    // Start the runner
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer runner.Stop()
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

**Collaborative Orchestration Code:**
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Initialize agents
    agents := map[string]core.AgentHandler{
        "researcher": NewResearcher(),
        "analyzer":   NewAnalyzer(),
        "validator":  NewValidator(),
    }
    
    // Create collaborative orchestration
    runner := core.NewOrchestrationBuilder(core.OrchestrationCollaborate).
        WithAgents(agents).
        WithTimeout(2 * time.Minute).
        WithFailureThreshold(0.8).
        WithMaxConcurrency(10).
        Build()
    
    // Create event
    event := core.NewEvent("all", map[string]interface{}{
        "task": "research AI trends and provide comprehensive analysis",
    }, nil)
    
    // All agents process the event in parallel
    result, err := runner.Run(context.Background(), event)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Collaborative Result: %s\n", result.GetResult())
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
- **[API Reference](tutorials\getting-started\README.md)** - Complete API documentation
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
