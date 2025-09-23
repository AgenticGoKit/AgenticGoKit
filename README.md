# AgenticGoKit

**Production-ready Go framework for building intelligent multi-agent AI systems**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kunalkushwaha/agenticgokit)](https://goreportcard.com/report/github.com/kunalkushwaha/agenticgokit)
[![Build Status](https://github.com/kunalkushwaha/agenticgokit/workflows/CI/badge.svg)](https://github.com/kunalkushwaha/agenticgokit/actions)
[![Documentation](https://img.shields.io/badge/docs-latest-blue)](docs/README.md)

AgenticGoKit enables developers to build sophisticated agent workflows with dynamic tool integration, multi-provider LLM support, and enterprise-grade orchestration patterns. Designed for Go developers who need the performance and reliability of compiled binaries with the flexibility of modern AI agent systems.

> **⚠️ Alpha Release**: AgenticGoKit is currently in alpha development. APIs may change, and some features are still being developed. Suitable for experimentation and early adoption, but not recommended for production use yet.

---

## Quick Start

```bash
# Install CLI
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest

# Create your first agent project
agentcli create my-agent --template basic
cd my-agent

# Set up environment (Azure OpenAI example)
export AZURE_OPENAI_API_KEY=your-key
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
export AZURE_OPENAI_DEPLOYMENT=your-deployment

# Run it
go run .
```

## Core Example

```go
package main

import (
    "context"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create agents
    agents := map[string]core.AgentHandler{
        "researcher": NewResearchAgent(),
        "analyzer":   NewAnalysisAgent(),
    }
    
    // Run collaborative workflow
    runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
    result, err := runner.ProcessEvent(ctx, event)
}
```

## What's Included

- **Multi-Agent Orchestration**: Parallel, sequential, and loop patterns
- **Memory & RAG**: PostgreSQL pgvector, Weaviate, in-memory providers
- **Tool Integration**: MCP protocol for dynamic tool discovery  
- **CLI Scaffolding**: Generate complete projects instantly
- **Configuration**: TOML-based with environment overrides
- **Visualization**: Auto-generated Mermaid workflow diagrams

## Project Structure

```
my-agent/
├── main.go              # Entry point
├── agentflow.toml       # Configuration
├── go.mod               # Go module
├── agents/              # Agent implementations
└── docs/                # Generated diagrams
```

## Configuration

```toml
# agentflow.toml
[orchestration]
mode = "collaborative"
timeout_seconds = 30

[agent_memory]
provider = "pgvector"
enable_rag = true

[mcp]
enabled = true
```

## Templates

```bash
# Available project templates
agentcli create my-project --template basic           # Single agent
agentcli create my-project --template research        # Research team
agentcli create my-project --template rag-system      # Knowledge base
agentcli create my-project --template chat-system     # Conversational
```

## Plugin System

```go
import (
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/memory/pgvector"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/orchestrator/default"
)
```

## Examples

Check out [`examples/`](examples/) for working demos:
- [Simple Agent](examples/01-simple-agent/)
- [Multi-Agent Collaboration](examples/02-multi-agent-collab/)
- [RAG Knowledge Base](examples/04-rag-knowledge-base/)
- [MCP Tool Integration](examples/05-mcp-agent/)

## Development

```bash
# Clone and build
git clone https://github.com/kunalkushwaha/agenticgokit.git
cd agenticgokit
make build

# Run tests
make test

# Run examples
cd examples/01-simple-agent
go run .
```

## Resources

- **Website**: [www.agenticgokit.com](https://www.agenticgokit.com)
- **Documentation**: [docs.agenticgokit.com](https://docs.agenticgokit.com)
- **Examples**: [examples/](examples/)
- **Discussions**: [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)
- **Issues**: [GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues)

## Contributing

We welcome contributions! See [docs/contributors/ContributorGuide.md](docs/contributors/ContributorGuide.md) for getting started.

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
