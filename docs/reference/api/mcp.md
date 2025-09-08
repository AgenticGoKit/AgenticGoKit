# MCP Integration API

**Tool integration via Model Context Protocol**

This document covers AgenticGoKit's MCP (Model Context Protocol) integration API, which enables agents to discover, connect to, and use external tools and services. MCP provides a standardized way to integrate with various tools, from web search to database operations.

## 📋 Core Concepts

### MCP Overview

MCP (Model Context Protocol) is a protocol for connecting AI agents with external tools and services. AgenticGoKit provides comprehensive MCP integration with three levels of complexity:

- **Basic MCP**: Simple tool discovery and execution
- **Enhanced MCP**: Caching and performance optimization
- **Production MCP**: Enterprise-grade features with monitoring and scaling

## 🚀 Basic Usage

### CLI Integration

The easiest way to get started with MCP is using AgentCLI:

```bash
# Create project with MCP enabled
agentcli create my-project --enable-mcp

# Create with specific MCP level
agentcli create my-project --mcp minimal    # Basic MCP
agentcli create my-project --mcp standard   # With caching
agentcli create my-project --mcp advanced   # Full enterprise features
```

For complete CLI documentation, see the [MCP CLI Guide](../../guides/MCP-CLI-Guide.md).

### Quick Start with MCP

```go
// Initialize MCP with automatic discovery
err := core.QuickStartMCP()
if err != nil {
    panic(fmt.Sprintf("Failed to initialize MCP: %v", err))
}

// Create an MCP-aware agent
llmProvider, err := core.NewOpenAIProvider()
if err != nil {
    panic(fmt.Sprintf("Failed to create LLM provider: %v", err))
}

agent, err := core.NewMCPAgent("assistant", llmProvider)
if err != nil {
    panic(fmt.Sprintf("Failed to create MCP agent: %v", err))
}
```

## 📚 Configuration Reference

### MCP Levels Comparison

| Level | CLI Flag | Caching | Metrics | Load Balancing | Use Case |
|-------|----------|---------|---------|---------------|----------|
| **Minimal** | `--mcp minimal` | ❌ | ❌ | ❌ | Development, Simple Apps |
| **Standard** | `--mcp standard` | ✅ | ❌ | ❌ | Production Applications |
| **Advanced** | `--mcp advanced` | ✅ | ✅ | ✅ | Enterprise Systems |

### Generated Configuration

When using AgentCLI with MCP flags, the following configuration is generated in `agentflow.toml`:

```toml
[mcp]
enabled = true
transport = "tcp"
enable_discovery = true
connection_timeout = 5000
max_retries = 3
retry_delay = 1000

# Additional features based on MCP level:
# --mcp standard adds:
enable_caching = true
cache_timeout = 300000

# --mcp advanced adds:
enable_metrics = true
enable_load_balancing = true
max_connections = 10

# Tool server examples (commented by default)
[[mcp.servers]]
name = "docker-http-sse"
type = "http_sse"
host = "localhost"
port = 8812
enabled = false

[[mcp.servers]]
name = "brave-search"
type = "stdio"
command = "npx @modelcontextprotocol/server-brave-search"
enabled = false
```

For complete documentation including server discovery, caching, production deployment, and custom tool development, see the [Agent API reference](agent.md#mcp).
