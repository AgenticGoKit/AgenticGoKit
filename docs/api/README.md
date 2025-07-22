# AgenticGoKit API Reference

**Complete API documentation for building multi-agent systems**

This section provides comprehensive documentation for all public APIs in AgenticGoKit. The framework is designed with a clean, intuitive interface that makes building complex agent systems straightforward.

## 📚 API Overview

AgenticGoKit's API is organized around several core concepts:

- **[Agent](agent.md)** - Individual agents that process events and states
- **[Orchestration](orchestration.md)** - Multi-agent coordination and workflow patterns
- **[State & Events](state-event.md)** - Data flow and communication between agents
- **[Memory](memory.md)** - Persistent storage, RAG, and knowledge management
- **[MCP Integration](mcp.md)** - Tool integration via Model Context Protocol
- **[Configuration](configuration.md)** - System configuration and setup

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

## 📖 Detailed Documentation

Each section provides comprehensive documentation with examples:

- **[Agent API](agent.md)** - Creating and configuring individual agents
- **[Orchestration API](orchestration.md)** - Multi-agent coordination patterns
- **[State & Event API](state-event.md)** - Data flow and communication
- **[Memory API](memory.md)** - Persistent storage and RAG systems
- **[MCP API](mcp.md)** - Tool integration and external services
- **[Configuration API](configuration.md)** - System setup and configuration

## 🧪 Testing Your Agents

AgenticGoKit provides testing utilities for agent systems:

```go
func TestAgent(t *testing.T) {
    agent := core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        return core.AgentResult{
            Data: map[string]interface{}{
                "processed": true,
            },
        }, nil
    })
    
    event := core.NewEvent("test", nil)
    state := core.NewState()
    
    result, err := agent.Run(context.Background(), event, state)
    
    assert.NoError(t, err)
    assert.True(t, result.Data["processed"].(bool))
}
```

## 🔗 Related Resources

- **[Tutorials](../tutorials/)** - Step-by-step learning guides
- **[How-To Guides](../how-to/)** - Task-oriented instructions
- **[Examples](../../examples/)** - Complete working examples
- **[Production Guide](../production/)** - Deployment and scaling

---

*This API documentation covers AgenticGoKit v0.1.0. The framework is actively developed, so some interfaces may evolve. Check the [changelog](../../CHANGELOG.md) for updates.*