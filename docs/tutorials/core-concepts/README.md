# Core Concepts Overview

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
result, err := agent.Run(ctx, inputState)
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
func main() {
    // Create agent
    agent := core.NewLLMAgent("assistant", openaiProvider)
    
    // Create state with input
    state := core.NewState()
    state.Set("message", "Hello, world!")
    
    // Run agent
    result, err := agent.Run(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get response
    response, _ := result.Get("response")
    fmt.Println(response)
}
```

### 2. Event-Driven Processing
```go
func main() {
    // Create runner with agents
    runner := core.NewRunnerWithConfig(core.RunnerConfig{
        Agents: map[string]core.AgentHandler{
            "processor": processorAgent,
            "responder": responderAgent,
        },
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
        "researcher": researchAgent,
        "analyzer":   analysisAgent,
        "writer":     writingAgent,
    }
    
    // All agents work on the same input
    runner := core.CreateCollaborativeRunner(agents, 60*time.Second)
    runner.Start(context.Background())
    
    event := core.NewEvent("", // No specific target - all agents get it
        core.EventData{"topic": "AI trends"}, nil)
    runner.Emit(event)
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
- **[Orchestration Overview](../orchestration/README.md)** - Learn about different orchestration modes

## Key Takeaways

- **Events** carry messages and data through the system
- **State** objects persist data between agent interactions
- **Agents** are the core processing units that transform state
- **Runners** manage the event processing loop
- **Orchestrators** coordinate how agents work together
- **Memory** provides persistent storage and RAG capabilities
- The system is designed for **scalability**, **fault tolerance**, and **flexibility**

Understanding these concepts will help you build more effective and maintainable multi-agent systems with AgenticGoKit.