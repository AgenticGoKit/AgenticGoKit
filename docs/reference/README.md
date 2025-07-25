# AgenticGoKit API Reference

> **Navigation:** [Documentation Home](../README.md) → **API Reference**

**Complete API documentation for building multi-agent systems**

This section provides comprehensive documentation for all public APIs in AgenticGoKit. The framework is designed with a clean, intuitive interface that makes building complex agent systems straightforward.

## 📚 API Overview

AgenticGoKit's API is organized around several core concepts:

- **[Agent API](api/agent.md)** - Individual agents that process events and states
- **[Orchestration API](api/orchestration.md)** - Multi-agent coordination and workflow patterns
- **[State & Events](api/state-event.md)** - Data flow and communication between agents
- **[Memory API](api/memory.md)** - Persistent storage, RAG, and knowledge management
- **[MCP Integration](api/mcp.md)** - Tool integration via Model Context Protocol
- **[Configuration API](api/configuration.md)** - System configuration and setup

## 🏗️ Architecture Overview

```mermaid
graph TB
    subgraph "Core Interfaces"
        Agent[Agent Interface]
        AgentHandler[AgentHandler Interface]
        Runner[Runner Interface]
        Orchestrator[Orchestrator Interface]
    end
    
    subgraph "Data Flow"
        Event[Event]
        State[State]
        AgentResult[AgentResult]
    end
    
    subgraph "Advanced Features"
        Memory[Memory System]
        MCP[MCP Tools]
        Config[Configuration]
    end
    
    Agent --> AgentHandler
    AgentHandler --> Runner
    Runner --> Orchestrator
    
    Event --> AgentHandler
    State --> AgentHandler
    AgentHandler --> AgentResult
    
    Memory --> Agent
    MCP --> Agent
    Config --> Runner
```

## 🚀 Quick Start

### Basic Agent Creation

```go
package main

import (
    "context"
    "fmt"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create a simple agent
    agent := core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        message := event.Data["message"].(string)
        
        return core.AgentResult{
            Data: map[string]interface{}{
                "response": fmt.Sprintf("Processed: %s", message),
            },
        }, nil
    })
    
    // Create a runner and register the agent
    runner := core.NewRunner()
    runner.RegisterAgent("processor", agent)
    
    // Process an event
    event := core.NewEvent("process", map[string]interface{}{
        "message": "Hello, AgenticGoKit!",
    })
    
    results, err := runner.ProcessEvent(context.Background(), event)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Response: %s\n", results["processor"].Data["response"])
}
```

### Multi-Agent Collaboration

```go
func collaborativeExample() {
    // Create multiple agents
    agents := map[string]core.AgentHandler{
        "analyzer": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            text := event.Data["text"].(string)
            analysis := analyzeText(text) // Your analysis logic
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "analysis": analysis,
                    "word_count": len(strings.Fields(text)),
                },
            }, nil
        }),
        
        "summarizer": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            text := event.Data["text"].(string)
            summary := summarizeText(text) // Your summarization logic
            
            return core.AgentResult{
                Data: map[string]interface{}{
                    "summary": summary,
                },
            }, nil
        }),
    }
    
    // Create collaborative runner
    runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
    
    // Process with multiple agents
    event := core.NewEvent("analyze", map[string]interface{}{
        "text": "Long document text here...",
    })
    
    results, _ := runner.ProcessEvent(context.Background(), event)
    
    // Both agents processed the event in parallel
    fmt.Printf("Analysis: %v\n", results["analyzer"].Data["analysis"])
    fmt.Printf("Summary: %s\n", results["summarizer"].Data["summary"])
}
```

## 🔧 Core Interfaces

### Agent vs AgentHandler

AgenticGoKit provides two main interfaces for creating agents:

**Agent Interface** - Simple, state-based processing:
```go
type Agent interface {
    Run(ctx context.Context, inputState State) (State, error)
    Name() string
}
```

**AgentHandler Interface** - Event-driven processing with rich results:
```go
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}
```

**When to use which:**
- Use `Agent` for simple, stateful transformations
- Use `AgentHandler` for event-driven systems with complex orchestration
- Use `AgentHandlerFunc` for quick function-based agents

### Conversion Between Interfaces

```go
// Convert Agent to AgentHandler
func ConvertAgentToHandler(agent Agent) AgentHandler {
    return core.ConvertAgentToHandler(agent)
}

// Convert function to AgentHandler
func ConvertFuncToHandler(fn func(context.Context, Event, State) (AgentResult, error)) AgentHandler {
    return core.AgentHandlerFunc(fn)
}
```

## 📊 Orchestration Patterns

AgenticGoKit supports multiple orchestration patterns:

| Pattern | Description | Use Case |
|---------|-------------|----------|
| **Route** | Single agent per event (default) | Simple request-response |
| **Collaborate** | All agents process in parallel | Analysis, multiple perspectives |
| **Sequential** | Agents process in order | Data pipelines, workflows |
| **Mixed** | Hybrid collaborative + sequential | Complex business processes |
| **Loop** | Single agent iterative processing | Refinement, optimization |

### Pattern Selection Guide

```go
// Choose orchestration pattern based on your needs
func chooseOrchestration(useCase string) core.OrchestrationMode {
    switch useCase {
    case "simple_processing":
        return core.OrchestrationRoute
    case "parallel_analysis":
        return core.OrchestrationCollaborate
    case "data_pipeline":
        return core.OrchestrationSequential
    case "complex_workflow":
        return core.OrchestrationMixed
    case "iterative_refinement":
        return core.OrchestrationLoop
    default:
        return core.OrchestrationRoute
    }
}
```

## 🧠 Memory and RAG

AgenticGoKit provides built-in support for persistent memory and RAG (Retrieval-Augmented Generation):

```go
// Create memory-enabled agent
memoryConfig := core.AgentMemoryConfig{
    Provider:   "pgvector",
    Connection: "postgres://user:pass@localhost:5432/agents",
    AutoEmbed:  true,
}

memory, _ := core.NewMemory(memoryConfig)
llmProvider, _ := core.NewOpenAIProvider()

agent := core.NewMemoryEnabledAgent("assistant", llmProvider, memory)
```

## 🔧 Tool Integration

Integrate external tools via MCP (Model Context Protocol):

```go
// Initialize MCP
core.QuickStartMCP()

// Create MCP-aware agent
llmProvider, _ := core.NewOpenAIProvider()
agent, _ := core.NewMCPAgent("assistant", llmProvider)

// Agent can now discover and use MCP tools automatically
```

## Command Line Interface

### [CLI Reference](cli.md)
Complete reference for the `agentcli` command-line tool.

**Available commands:**
- `create` - Create new projects with multi-agent orchestration
- `trace` - View execution traces and debugging information
- `mcp` - Manage MCP servers and tools
- `cache` - Cache management and optimization
- `run` - Run agents interactively for testing
- `test` - Run automated agent tests
- `benchmark` - Performance benchmarking
- `config` - Configuration management

## API Design Principles

AgenticGoKit APIs follow these principles:

- **Interface-based design** - Clear separation of concerns
- **Context-aware** - All operations accept Go context
- **Error handling** - Explicit error returns and handling
- **Configuration-driven** - Behavior controlled through configuration
- **Extensible** - Plugin architecture for custom implementations

## Type Definitions

All APIs use consistent type definitions:

```go
// Core types
type State map[string]interface{}
type Event struct { /* ... */ }
type AgentResult struct { /* ... */ }

// Configuration types
type Config struct { /* ... */ }
type OrchestrationMode string
```

## Usage Patterns

Common patterns across all APIs:

### Context Usage
```go
ctx := context.Background()
result, err := agent.Run(ctx, inputState)
```

### Error Handling
```go
if err != nil {
    log.Printf("Operation failed: %v", err)
    return err
}
```

### Configuration Loading
```go
config, err := core.LoadConfigFromWorkingDir()
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}
```

## Versioning

AgenticGoKit follows semantic versioning:
- **Major versions** - Breaking API changes
- **Minor versions** - New features, backward compatible
- **Patch versions** - Bug fixes, backward compatible

## Migration Guides

When APIs change between versions, migration guides are provided:
- Breaking changes are documented
- Migration examples are provided
- Deprecation notices give advance warning

## Getting Help

For API questions:
1. Check this reference documentation
2. Review [code examples](../tutorials/README.md) in tutorials
3. See [how-to guides](../guides/README.md) for specific use cases
4. Check the [troubleshooting guide](../guides/troubleshooting.md) for common issues