# Model Context Protocol (MCP) in AgenticGoKit

> **Navigation:** [Documentation Home](../../README.md) → [Tutorials](../README.md) → **MCP (Tools)**

## Overview

The Model Context Protocol (MCP) is a powerful framework within AgenticGoKit that enables agents to interact with external tools, APIs, and services. MCP bridges the gap between language models and the outside world, allowing agents to perform actions beyond text generation.

With MCP, agents can search the web, access databases, call APIs, manipulate files, perform calculations, and much more. This capability transforms agents from simple text processors into powerful assistants that can take meaningful actions.

## Key Concepts

### What is MCP?

MCP (Model Context Protocol) is a standardized interface for connecting language models to external tools and capabilities. It defines:

1. **Tool Registration**: How tools are defined and registered with the system
2. **Tool Discovery**: How agents discover available tools
3. **Tool Invocation**: How agents call tools and receive results
4. **Tool Response Handling**: How tool results are processed and incorporated into agent responses

### MCP Architecture

```
┌─────────────┐     ┌───────────────┐     ┌─────────────┐
│             │     │               │     │             │
│    Agent    │────▶│  MCP Manager  │────▶│    Tool     │
│             │     │               │     │             │
└─────────────┘     └───────────────┘     └─────────────┘
       ▲                     │                   │
       │                     ▼                   ▼
       │              ┌───────────────┐    ┌─────────────┐
       └──────────────│  Tool Result  │◀───│   External  │
                      │   Processor   │    │   Service   │
                      └───────────────┘    └─────────────┘
```

### Tool Types

AgenticGoKit supports various types of tools:

1. **Built-in Tools**: Core functionality provided by the framework
2. **Custom Tools**: User-defined tools for specific use cases
3. **API Tools**: Wrappers around external APIs and services
4. **Stateful Tools**: Tools that maintain state between invocations
5. **Composite Tools**: Tools composed of multiple sub-tools

## Why Use MCP?

MCP provides several key benefits:

- **Extended Capabilities**: Enables agents to perform actions beyond text generation
- **Modularity**: Tools can be developed and maintained independently
- **Flexibility**: Mix and match tools based on specific requirements
- **Standardization**: Consistent interface for all tool interactions
- **Security**: Controlled access to external systems

## MCP vs. Function Calling

MCP is similar to function calling in LLMs but provides additional capabilities:

| Feature | MCP | Function Calling |
|---------|-----|------------------|
| Tool Discovery | Dynamic | Static |
| Tool Registration | Runtime | Design time |
| Tool Composition | Supported | Limited |
| State Management | Built-in | Manual |
| Error Handling | Comprehensive | Basic |
| Security Controls | Fine-grained | Limited |

## Getting Started with MCP

To start using MCP in AgenticGoKit, you'll need to:

1. **Create Tools**: Define the tools your agents will use
2. **Register Tools**: Make tools available to the MCP manager
3. **Configure Agents**: Set up agents to use MCP
4. **Handle Tool Results**: Process and incorporate tool outputs

The following tutorials will guide you through these steps in detail:

- [Tool Development](tool-development.md) - Creating custom tools
- [Tool Integration](tool-integration.md) - Integrating tools with agents
- [Advanced Tool Patterns](advanced-tool-patterns.md) - Complex tool usage patterns

## Example: Simple MCP Setup

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Initialize MCP for tool discovery
    core.QuickStartMCP()
    
    // Create LLM provider
    provider, err := core.NewOpenAIAdapter(
        os.Getenv("OPENAI_API_KEY"),
        "gpt-3.5-turbo",
        1000,
        0.7,
    )
    if err != nil {
        log.Fatalf("Failed to create provider: %v", err)
    }
    
    // Create agent with MCP tools
    agent := core.NewLLMAgent("assistant", provider).
        WithSystemPrompt("You are a helpful assistant with access to tools. Use them when needed.").
        WithTools([]string{"calculator", "weather"}) // Tools discovered via MCP
    
    // Create agents map
    agents := map[string]core.AgentHandler{
        "assistant": agent,
    }
    
    // Create runner
    runner := core.CreateRouteRunner(agents)
    
    // Start runner
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        log.Fatalf("Failed to start runner: %v", err)
    }
    defer runner.Stop()
    
    // Create event with user query
    event := core.NewEvent("assistant", map[string]interface{}{
        "message": "What's 25 * 16 and what's the weather in New York?",
    })
    event.SetMetadata(core.RouteMetadataKey, "assistant")
    
    // Emit the event
    if err := runner.Emit(event); err != nil {
        log.Fatalf("Failed to emit event: %v", err)
    }
    
    // Wait for processing
    time.Sleep(5 * time.Second)
    
    fmt.Println("MCP tool integration complete!")
}
```

## Next Steps

Now that you understand the basics of MCP, proceed to the following tutorials to learn more:

- [Tool Development](tool-development.md) - Learn how to create custom tools
- [Tool Integration](tool-integration.md) - Integrate tools with your agents
- [Advanced Tool Patterns](advanced-tool-patterns.md) - Explore complex tool usage patterns

## Further Reading

- [API Reference: MCP](../../reference/api/agent.md#mcp)
- [Examples: Tool Usage](../../examples/)
- [Advanced Patterns: Tool Composition](../advanced-patterns/multi-agent-collaboration.md)