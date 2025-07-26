# Core Concepts Overview

> **Navigation:** [Documentation Home](../../README.md) → [Tutorials](../README.md) → **Core Concepts**

Understanding AgenticGoKit's core concepts is essential for building effective multi-agent systems. This section covers the fundamental building blocks that power the framework.

## The Big Picture

AgenticGoKit is built around a few key concepts that work together to create a powerful multi-agent system:

```mermaid
graph TB
    Event[Event] --> Runner[Runner]
    Runner --> Orchestrator[Orchestrator]
    Orchestrator --> Agent[Agent]
    Agent --> State[State]
    State --> Memory[Memory]
    
    subgraph "Core Flow"
        Event --> State
        State --> Agent
        Agent --> State
    end
    
    subgraph "Orchestration"
        Orchestrator --> AgentA[Agent A]
        Orchestrator --> AgentB[Agent B]
        Orchestrator --> AgentC[Agent C]
    end
```

## Key Components

### 1. Events - The Message System
Events are the messages that flow through your agent system. They carry data, metadata, and routing information.

```go
// Create an event
event := core.NewEvent("target-agent", 
    core.EventData{"message": "Hello, world!"}, 
    map[string]string{"priority": "high"})

// Events have IDs, timestamps, and routing info
fmt.Println("Event ID:", event.GetID())
fmt.Println("Target:", event.GetTargetAgentID())
```

### 2. State - The Data Container
State objects carry data between agents and persist information across interactions.

```go
// Create and manipulate state
state := core.NewState()
state.Set("user_input", "What's the weather like?")
state.SetMeta("session_id", "user-123")

// State is thread-safe and can be cloned/merged
clonedState := state.Clone()
```

### 3. Agents - The Processing Units
Agents are the core processing units that transform input state into output state.

```go
// Agents implement a simple interface
type Agent interface {
    Run(ctx context.Context, inputState State) (State, error)
    Name() string
}

// Create a simple agent
agent := core.NewLLMAgent("assistant", llmProvider)
agentResult, err := agent.Run(ctx, event, state)
```

### 4. Runner - The Event Processor
The Runner manages the event processing loop, routing events to the appropriate agents.

```go
// Create and start a runner
runner := core.NewRunner(100) // queue size
runner.RegisterAgent("assistant", agentHandler)
runner.Start(ctx)

// Emit events for processing
runner.Emit(event)
```

### 5. Orchestrator - The Coordination Engine
Orchestrators determine how events are distributed to agents (single, parallel, sequential, etc.).

```go
// Different orchestration modes
collaborativeRunner := core.CreateCollaborativeRunner(agents, 30*time.Second)
sequentialRunner := core.NewOrchestrationBuilder(core.OrchestrationSequential).
    WithAgents(agents).
    Build()
```

## Data Flow Architecture

Understanding how data flows through the system is crucial:

```mermaid
sequenceDiagram
    participant Client
    participant Runner
    participant Orchestrator
    participant Agent
    participant Memory
    
    Client->>Runner: Emit Event
    Runner->>Orchestrator: Dispatch Event
    Orchestrator->>Agent: Run(ctx, event, state)
    Agent->>Memory: Store/Retrieve Data
    Memory-->>Agent: Return Results
    Agent-->>Orchestrator: Return AgentResult
    Orchestrator-->>Runner: Return Result
    Runner->>Runner: Process Result (emit new events)
```

## Core Patterns

### 1. Simple Agent Execution
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/AgenticGoKit/core"
)

// Simple agent implementation
type SimpleAgent struct {
    name string
    llm  core.ModelProvider
}

func (a *SimpleAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get message from event
    message, ok := event.GetData()["message"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("no message found")
    }
    
    // Create prompt
    prompt := core.Prompt{
        System: "You are a helpful assistant.",
        User:   message,
    }
    
    // Call LLM
    response, err := a.llm.Call(ctx, prompt)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Create output state
    outputState := core.NewState()
    outputState.Set("response", response.Content)
    
    return core.AgentResult{OutputState: outputState}, nil
}

func main() {
    // Create LLM provider from configuration
    provider, err := core.NewProviderFromWorkingDir()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create agent
    agent := &SimpleAgent{
        name: "assistant",
        llm:  provider,
    }
    
    // Create state with input
    state := core.NewState()
    state.Set("message", "Hello, world!")
    
    // Create event
    event := core.NewEvent("assistant", core.EventData{
        "message": "Hello, world!",
    }, nil)
    
    // Run agent
    result, err := agent.Run(context.Background(), event, state)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get response from output state
    if response, ok := result.OutputState.Get("response"); ok {
        fmt.Println(response)
    }
}
```

### 2. Event-Driven Processing
```go
func main() {
    // Create agents
    agents := map[string]core.AgentHandler{
        "processor": &ProcessorAgent{},
        "responder": &ResponderAgent{},
    }
    
    // Create collaborative runner
    runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
    })
    
    // Start processing
    runner.Start(context.Background())
    
    // Emit event
    event := core.NewEvent("processor", 
        core.EventData{"task": "analyze data"}, nil)
    runner.Emit(event)
}
```

### 3. Multi-Agent Collaboration
```go
func main() {
    agents := map[string]core.AgentHandler{
        "researcher": &ResearchAgent{},
        "analyzer":   &AnalysisAgent{},
        "writer":     &WritingAgent{},
    }
    
    // All agents work on the same input
    runner := core.CreateCollaborativeRunner(agents, 60*time.Second)
    
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    event := core.NewEvent("researcher", // Start with researcher
        core.EventData{"topic": "AI trends"}, 
        map[string]string{"route": "researcher"})
    
    err := runner.Emit(event)
    if err != nil {
        log.Fatal(err)
    }
    
    // Wait for processing
    time.Sleep(10 * time.Second)
}
```

## Memory Integration

AgenticGoKit provides powerful memory capabilities for persistent storage and RAG:

```go
// Configure memory
memoryConfig := core.AgentMemoryConfig{
    Provider:   "pgvector",
    Connection: "postgres://user:pass@localhost/db",
    EnableRAG:  true,
}

memory, err := core.NewMemory(memoryConfig)

// Store information
memory.Store(ctx, core.MemoryItem{
    Content: "Important information",
    Tags:    []string{"important", "user-data"},
})

// Search with RAG
results, err := memory.Search(ctx, "find important information")
```

## Error Handling Patterns

AgenticGoKit provides sophisticated error handling and recovery:

```go
// Configure error routing
errorConfig := core.DefaultErrorRouterConfig()
errorConfig.MaxRetries = 3
errorConfig.BackoffFactor = 2.0

runner := core.NewRunnerWithConfig(core.RunnerConfig{
    ErrorRouterConfig: errorConfig,
    Agents: map[string]core.AgentHandler{
        "main-agent":    mainAgent,
        "error-handler": errorHandlerAgent,
    },
})
```

## Next Steps

Now that you understand the core concepts, dive deeper into specific areas:

1. **[Message Passing](message-passing.md)** - Learn how events flow through the system
2. **[State Management](state-management.md)** - Master data handling between agents
3. **[Agent Lifecycle](agent-lifecycle.md)** - Understand agent creation and execution
4. **[Error Handling](error-handling.md)** - Build robust error management

Or jump to specific orchestration patterns:
- **[Orchestration Overview](../README.md)** - Learn about different orchestration modes

## Key Takeaways

- **Events** carry messages and data through the system
- **State** objects persist data between agent interactions
- **Agents** are the core processing units that transform state
- **Runners** manage the event processing loop
- **Orchestrators** coordinate how agents work together
- **Memory** provides persistent storage and RAG capabilities
- The system is designed for **scalability**, **fault tolerance**, and **flexibility**

Understanding these concepts will help you build more effective and maintainable multi-agent systems with AgenticGoKit.
