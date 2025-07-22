# AgenticGoKit Tutorials

Welcome to the AgenticGoKit tutorial collection! These tutorials are designed to take you from basic concepts to advanced patterns, helping you build robust multi-agent systems with confidence.

## Learning Path

### 🚀 Getting Started
- [5-Minute Quickstart](../quickstart.md) - Get your first agent running
- [Core Concepts Overview](core-concepts/README.md) - Understand the fundamentals

### 🏗️ Core Concepts
- [Message Passing & Event Flow](core-concepts/message-passing.md) - How data flows through the system
- [State Management](core-concepts/state-management.md) - Managing data between agents
- [Agent Lifecycle](core-concepts/agent-lifecycle.md) - Agent creation and execution
- [Error Handling](core-concepts/error-handling.md) - Robust error management

### 🎭 Orchestration Patterns
- [Orchestration Overview](orchestration/README.md) - Understanding orchestration modes
- [Routing Mode](orchestration/routing-mode.md) - Single agent routing (default)
- [Collaborative Mode](orchestration/collaborative-mode.md) - Parallel agent execution
- [Sequential Mode](orchestration/sequential-mode.md) - Pipeline processing
- [Mixed Mode](orchestration/mixed-mode.md) - Hybrid workflows
- [Loop Mode](orchestration/loop-mode.md) - Iterative processing

### 🧠 Memory Systems
- [Memory Overview](memory-systems/README.md) - Memory system fundamentals
- [Basic Memory](memory-systems/basic-memory.md) - In-memory storage
- [Vector Databases](memory-systems/vector-databases.md) - pgvector, Weaviate integration
- [RAG Implementation](memory-systems/rag-implementation.md) - Retrieval-Augmented Generation
- [Knowledge Bases](memory-systems/knowledge-bases.md) - Document ingestion and search

### 🔧 MCP Integration
- [MCP Overview](mcp-integration/README.md) - Model Context Protocol basics
- [Tool Discovery](mcp-integration/tool-discovery.md) - Finding and connecting to tools
- [Tool Execution](mcp-integration/tool-execution.md) - Using tools in agents
- [Custom Tools](mcp-integration/custom-tools.md) - Building MCP servers
- [Production MCP](mcp-integration/production-mcp.md) - Enterprise patterns

### 🐛 Debugging & Monitoring
- [Debugging Overview](debugging-monitoring/README.md) - Debugging fundamentals
- [Tracing Events](debugging-monitoring/tracing-events.md) - Event flow tracing
- [Logging Patterns](debugging-monitoring/logging-patterns.md) - Structured logging
- [Performance Profiling](debugging-monitoring/performance-profiling.md) - Performance analysis
- [Troubleshooting Guide](debugging-monitoring/troubleshooting-guide.md) - Common issues

### 🚀 Advanced Patterns
- [Advanced Overview](advanced-patterns/README.md) - Advanced patterns introduction
- [Circuit Breakers](advanced-patterns/circuit-breakers.md) - Fault tolerance
- [Retry Policies](advanced-patterns/retry-policies.md) - Retry strategies
- [Load Balancing](advanced-patterns/load-balancing.md) - Scaling patterns
- [Testing Strategies](advanced-patterns/testing-strategies.md) - Testing multi-agent systems

## Tutorial Format

Each tutorial follows a consistent structure:

- **Overview** - What you'll learn and why it matters
- **Prerequisites** - What you need to know first
- **Step-by-Step Guide** - Hands-on implementation
- **Code Examples** - Working code you can run
- **Best Practices** - Production-ready patterns
- **Troubleshooting** - Common issues and solutions
- **Next Steps** - Where to go from here

## Running Examples

All tutorial examples are designed to be runnable. Most require:

```bash
# Install AgenticGoKit
go mod init my-tutorial
go get github.com/kunalkushwaha/agenticgokit

# Set up environment (if needed)
export OPENAI_API_KEY=your-key-here

# Run the example
go run main.go
```

For examples requiring databases or external services, we provide Docker Compose files for one-command setup.

## Contributing

Found an issue or want to improve a tutorial? We welcome contributions! See our [Contributing Guide](../../CONTRIBUTING.md) for details.

## Support

- 💬 [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions) - Ask questions
- 🐛 [GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues) - Report bugs
- 📖 [API Documentation](../api/README.md) - Detailed API reference