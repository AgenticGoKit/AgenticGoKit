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

For complete documentation including server discovery, caching, production deployment, and custom tool development, see the [Agent API reference](agent.md#mcp).
