# AgentFlow Documentation

**The Complete Guide to Building AI Agent Systems in Go**

AgentFlow is a production-ready Go framework for building intelligent agent workflows with dynamic tool integration, multi-provider LLM support, and enterprise-grade patterns.

## 📚 For AgentFlow Users

### **Getting Started**
- **[Quick Start Guide](#quick-start)** - Get running in 5 minutes
- **[Installation & Setup](#installation)** - Go module setup and CLI installation
- **[Your First Agent](#first-agent)** - Build a simple agent from scratch
- **[Multi-Agent Workflows](#multi-agent)** - Orchestrate multiple agents

### **Core Concepts**  
- **[Agent Fundamentals](guides/AgentBasics.md)** - Understanding AgentHandler interface and patterns
- **[Examples & Tutorials](guides/Examples.md)** - Practical examples and code samples
- **[Tool Integration](guides/ToolIntegration.md)** - MCP protocol and dynamic tool discovery
- **[LLM Providers](guides/Providers.md)** - Azure, OpenAI, Ollama, and custom providers
- **[Configuration](guides/Configuration.md)** - Managing agentflow.toml and environment setup

### **Advanced Usage**
- **[Production Deployment](guides/Production.md)** - Scaling, monitoring, and best practices  
- **[Error Handling](guides/ErrorHandling.md)** - Resilient agent workflows
- **[Custom Tools](guides/CustomTools.md)** - Building your own MCP servers
- **[Performance Tuning](guides/Performance.md)** - Optimization and benchmarking

### **API Reference**
- **[Core Package API](api/core.md)** - Complete public API reference
- **[Agent Interface](api/agents.md)** - AgentHandler and related types
- **[MCP Integration](api/mcp.md)** - Tool discovery and execution APIs
- **[CLI Commands](api/cli.md)** - agentcli reference

## 🔧 For AgentFlow Contributors

### **Development Setup**
- **[Contributor Guide](contributors/ContributorGuide.md)** - Getting started with development
- **[Architecture Deep Dive](contributors/Architecture.md)** - Internal structure and design decisions
- **[Testing Strategy](contributors/Testing.md)** - Unit tests, integration tests, and benchmarks
- **[Release Process](contributors/ReleaseProcess.md)** - How releases are managed

### **Codebase Structure**
- **[Core vs Internal](contributors/CoreVsInternal.md)** - Public API vs implementation
- **[Adding Features](contributors/AddingFeatures.md)** - How to extend AgentFlow
- **[Code Style](contributors/CodeStyle.md)** - Go standards and project conventions
- **[Documentation Standards](contributors/DocsStandards.md)** - Writing user-focused docs

---

## Quick Start

### Installation
```bash
# Install the CLI
go install github.com/kunalkushwaha/agentflow/cmd/agentcli@latest

# Create your first project
agentcli create my-agent-app --agents 2 --mcp-enabled
cd my-agent-app

# Run your agents
go run . -m "search for the latest Go tutorials and summarize them"
```

### First Agent
```bash
# Generate a single agent project
agentcli create simple-agent

# The generated agent1.go will look like this:
```

```go
package main

import (
    "context"
    "fmt"
    agentflow "github.com/kunalkushwaha/agentflow/core"
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
# Generate a multi-agent workflow
agentcli create research-system --agents 3 --mcp-enabled --provider azure

# This creates:
# - agent1.go (Research agent - gathers information)
# - agent2.go (Analysis agent - processes data)  
# - agent3.go (Summary agent - final synthesis)
# - workflow orchestration in main.go
```

## Why AgentFlow?

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

## Contributing

We welcome contributions! See our [Contributor Guide](contributors/ContributorGuide.md) for details.

```bash
# Quick start for contributors
git clone https://github.com/kunalkushwaha/agentflow.git
cd agentflow
go mod tidy
go test ./...

# Generate docs
go run tools/docgen/main.go
```

## Community

- **[GitHub Discussions](https://github.com/kunalkushwaha/agentflow/discussions)** - Q&A and community
- **[Issues](https://github.com/kunalkushwaha/agentflow/issues)** - Bug reports and feature requests
- **[Contributing](CONTRIBUTING.md)** - How to contribute code and documentation

---

**[⭐ Star us on GitHub](https://github.com/kunalkushwaha/agentflow)** | **[📖 Full Documentation](https://agentflow.dev)** | **[🚀 Examples](examples/)**
