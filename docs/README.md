# AgentFlow Documentation

Welcome to the AgentFlow documentation! This guide helps you build intelligent agent workflows with Model Context Protocol (MCP) integration.

## 🚀 Getting Started

- **[Quick Start Guide](../README.md)** - Get AgentFlow running in minutes
- **[MCP API Quick Reference](MCP_API_Quick_Reference.md)** - 30-second MCP setup
- **[MCP API Usage Guide](MCP_API_Usage_Guide.md)** - Complete MCP integration guide

## 📚 Core Documentation

### **MCP Integration (Model Context Protocol)**
- **[MCP API Usage Guide](MCP_API_Usage_Guide.md)** - Complete usage guide with examples
- **[MCP API Quick Reference](MCP_API_Quick_Reference.md)** - Developer cheat sheet
- **[MCP API Migration Guide](MCP_API_Migration_Guide.md)** - Upgrade from old API
- **[MCP Public API Design](MCP_Public_API_Design.md)** - Technical design decisions
- **[MCP Technical Specification](MCP_Technical_Specification.md)** - Implementation details

### **Architecture & Design**
- **[Architecture Overview](Architecture.md)** - System architecture and components
- **[Development Guide](DevGuide.md)** - Contributing and development setup
- **[Library Usage Guide](LibraryUsageGuide.md)** - Using AgentFlow as a library
- **[Tracing Guide](TracingGuide.md)** - Debugging and monitoring

### **Planning & Roadmap**
- **[Roadmap](ROADMAP.md)** - Future development plans
- **[MCP Integration Plan](MCP_Integration_Plan.md)** - MCP feature planning

## 🔧 API Reference

### **Core Package (`core/`)**
The consolidated public API for AgentFlow:

```go
import "github.com/kunalkushwaha/agentflow/core"
```

**Key Files:**
- **`core/mcp.go`** - Complete MCP public API (interfaces, factories, config)
- **`core/mcp_agent.go`** - MCP-aware agent implementation
- **`core/agent.go`** - Basic agent interfaces
- **`core/factory.go`** - Agent and runner factories

### **Usage Patterns**

#### **Basic MCP Usage**
```go
// Initialize and create agent
core.QuickStartMCP()
agent, err := core.NewMCPAgent("my-agent", llmProvider)

// Use agent
state := core.NewState()
state.Set("query", "search for AI news")
result, err := agent.Run(ctx, state)
```

#### **Production MCP Usage**
```go
// Production setup
prodConfig := core.DefaultProductionConfig()
core.InitializeProductionMCP(ctx, prodConfig)
agent, err := core.NewProductionMCPAgent("prod-agent", llmProvider, prodConfig)
```

## 📁 Examples

Explore practical examples in the [`examples/`](../examples/) directory:

### **MCP Examples**
- **[MCP Integration](../examples/mcp_integration/)** - Basic MCP server connection
- **[MCP Working Demo](../examples/mcp_working_demo/)** - Complete working example
- **[MCP Production Demo](../examples/mcp_production_demo/)** - Production-ready setup
- **[MCP Agent Demo](../examples/mcp_agent_demo/)** - Agent-focused examples

### **Core Examples**  
- **[Minimal Example](../examples/minimal-example/)** - Simplest possible agent
- **[OpenAI Example](../examples/openai_example/)** - Using OpenAI models
- **[Ollama Example](../examples/ollama_example/)** - Using local Ollama models
- **[Multi-Agent](../examples/multi_agent/)** - Multiple cooperating agents

### **Advanced Examples**
- **[Memory Agent](../examples/memory_agent/)** - Agents with persistent memory
- **[Tools Usage](../examples/tools/)** - Custom tool integration
- **[Clean Multi-Agent](../examples/clean_multi_agent/)** - Production multi-agent setup

## 🏗️ Architecture Overview

AgentFlow follows a clean architecture with clear separation:

```
🎯 Public API (core/)
├── Agent interfaces and factories
├── MCP integration (complete in mcp.go)
├── Configuration management
└── State and result types

🔧 Internal Implementation (internal/)
├── Agent runners and orchestration
├── MCP protocol implementation  
├── Tool registry and management
└── Factory implementations

📦 Examples & Tools (examples/, cmd/)
├── Usage examples for all features
├── CLI tools (agentcli)
└── Integration demonstrations
```

## 🎯 Key Concepts

### **Agents**
Intelligent entities that can:
- Accept input state
- Process using LLMs and tools
- Return output state
- Integrate with MCP tools seamlessly

### **MCP Integration**
Model Context Protocol support for:
- **Dynamic tool discovery** - Find tools automatically
- **Multi-server connections** - Connect to multiple MCP servers
- **Intelligent caching** - Cache tool results for performance
- **Production features** - Pooling, retries, metrics, load balancing

### **State Management**
- **Input State** - Parameters and context for agent execution
- **Output State** - Results and computed values
- **State Transformation** - Agents transform input → output state

### **Progressive Complexity**
1. **Basic** - Simple agents with core functionality
2. **Enhanced** - Add caching, better configuration  
3. **Production** - Full enterprise features

## 🛠️ Development

### **Setup Development Environment**
```bash
git clone https://github.com/kunalkushwaha/agentflow
cd agentflow
go mod download
go test ./...
```

### **Project Structure**
```
agentflow/
├── core/           # 🎯 Public API 
├── internal/       # 🔧 Implementation
├── examples/       # 📚 Usage examples
├── cmd/           # 🚀 CLI tools
├── docs/          # 📖 Documentation
└── tests/         # 🧪 Integration tests
```

### **Contributing**
1. Read the [Development Guide](DevGuide.md)
2. Check existing [issues](https://github.com/kunalkushwaha/agentflow/issues)
3. Follow the contribution guidelines
4. Submit pull requests with tests

## 📖 Learning Path

### **New to AgentFlow?**
1. Start with [README](../README.md) - Basic concepts
2. Try [Minimal Example](../examples/minimal-example/) - First agent
3. Read [MCP Quick Reference](MCP_API_Quick_Reference.md) - MCP basics
4. Explore [MCP Usage Guide](MCP_API_Usage_Guide.md) - Complete MCP

### **Upgrading from Previous Versions?**
1. Check [Migration Guide](MCP_API_Migration_Guide.md) - API changes
2. Review [MCP Usage Guide](MCP_API_Usage_Guide.md) - New features
3. Test with your existing code - Should work unchanged!

### **Building Production Systems?**
1. Study [MCP Usage Guide](MCP_API_Usage_Guide.md) - Production patterns
2. Review [Architecture](Architecture.md) - System design
3. Check [Production Examples](../examples/mcp_production_demo/) - Real implementations
4. Follow [Best Practices](MCP_API_Usage_Guide.md#best-practices) - Proven patterns

## 🆘 Support

- **Documentation Issues** - File GitHub issues
- **API Questions** - Check usage guides first
- **Bug Reports** - Include minimal reproduction case
- **Feature Requests** - Describe use case and benefits

## 📜 License

AgentFlow is open source software. See the [LICENSE](../LICENSE) file for details.

---

**Ready to build intelligent agents? Start with the [MCP Quick Reference](MCP_API_Quick_Reference.md)! 🚀**
