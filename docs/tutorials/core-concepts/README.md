---
title: Core Concepts Overview
description: Essential guide to AgenticGoKit's fundamental building blocks and multi-agent system architecture
sidebar: true
prev: false
next: ./message-passing
outline: deep
tags:
  - core-concepts
  - architecture
  - multi-agent
  - tutorial
  - overview
---

# Core Concepts Overview

Understanding AgenticGoKit's core concepts is essential for building effective multi-agent systems. This section covers the fundamental building blocks that power the framework and provides accurate, up-to-date examples using the current API.

::: tip Learning Path
Start with this overview, then progress through [Message Passing](./message-passing.md) → [State Management](./state-management.md) → [Agent Lifecycle](./agent-lifecycle.md) → [Orchestration Patterns](./orchestration-patterns.md) → [Error Handling](./error-handling.md)
:::

## The Big Picture

AgenticGoKit is built around a configuration-driven, event-based architecture that enables flexible multi-agent coordination:

```mermaid
graph TB
    Config[agentflow.toml] --> Runner[Runner]
    Client[Client Code] --> Event[Event]
    Event --> Runner
    Runner --> Orchestrator[Orchestrator]
    Orchestrator --> Agent[Unified Agent Interface]
    Agent --> State[Thread-Safe State]
    State --> Memory[Memory & RAG]
    
    subgraph "Agent Capabilities"
        Agent --> LLM[LLM Integration]
        Agent --> MCP[MCP Tools]
        Agent --> Metrics[Metrics & Tracing]
    end
    
    subgraph "Orchestration Modes"
        Orchestrator --> Route[Route]
        Orchestrator --> Collab[Collaborative]
        Orchestrator --> Seq[Sequential]
        Orchestrator --> Loop[Loop]
        Orchestrator --> Mixed[Mixed]
    end
```

## Key Components

### 1. Configuration-Driven Architecture

AgenticGoKit uses `agentflow.toml` files to configure the entire system:

```toml
[agent_flow]
name = "my-agent-system"
version = "1.0.0"

[llm]
provider = "ollama"
model = "gemma3:1b"
temperature = 0.7

[orchestration]
mode = "collaborative"
timeout_seconds = 30
collaborative_agents = ["analyzer", "validator"]

[agents.analyzer]
role = "data_analyzer"
system_prompt = "You are a data analysis expert."
enabled = true
capabilities = ["analysis", "reporting"]
```

### 2. Unified Agent Interface

AgenticGoKit provides two interfaces for agent implementation:

```go
// Full Agent interface (for advanced agents)
type Agent interface {
    // Core identification
    Name() string
    GetRole() string
    GetDescription() string
    
    // Primary execution pattern (recommended)
    Run(ctx context.Context, inputState State) (State, error)
    
    // Event-driven pattern (for compatibility with Runner)
    HandleEvent(ctx context.Context, event Event, state State) (AgentResult, error)
    
    // Configuration and capabilities
    GetCapabilities() []string
    GetSystemPrompt() string
    IsEnabled() bool
    
    // Lifecycle management
    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
}

// AgentHandler interface (for simple agents)
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

// Function type for easy agent creation
type AgentHandlerFunc func(ctx context.Context, event Event, state State) (AgentResult, error)
```

::: tip Agent vs AgentHandler Usage
- **Agent.Run()**: Use for direct state processing and composition patterns
- **Agent.HandleEvent()**: Used internally by Runner for event-driven orchestration
- **AgentHandlerFunc**: Lightweight wrapper for simple processing functions
- Both Agent and AgentHandler can be registered with `Runner.RegisterAgent()`
:::

### 3. Thread-Safe State Management

State objects provide thread-safe data and metadata storage:

```go
// Create and manipulate state
state := core.NewState()
state.Set("user_query", "What's the weather like?")
state.SetMeta("session_id", "user-123")
state.SetMeta("confidence", "0.95")

// Thread-safe operations
clonedState := state.Clone()
state.Merge(otherState)

// Type-safe access
if query, ok := state.Get("user_query"); ok {
    if queryStr, ok := query.(string); ok {
        // Process query safely
    }
}
```

### 4. Event-Driven Communication

Events carry structured data and routing information through the system:

```go
// Create events with data and metadata
event := core.NewEvent("target-agent", 
    core.EventData{
        "message": "Hello, world!",
        "priority": "high",
    }, 
    map[string]string{
        "session_id": "user-123",
        "route": "target-agent",
    })

// Events are immutable with thread-safe access
fmt.Println("Event ID:", event.GetID())
fmt.Println("Target:", event.GetTargetAgentID())
fmt.Println("Session:", event.GetSessionID())
```

### 5. Configuration-Based Runners

Runners are created from configuration and manage the entire event processing lifecycle:

```go
// Create runner with configuration
config := core.DefaultRunnerConfig()
runner := core.NewRunnerWithConfig(config)

// Or create a basic runner
// runner := core.NewRunner(1000) // queue size

// Register agents by name (matching config)
err = runner.RegisterAgent("analyzer", analyzerHandler)
err = runner.RegisterAgent("validator", validatorHandler)

// Start → Emit → Stop lifecycle
ctx := context.Background()
err = runner.Start(ctx)
defer runner.Stop()

// Emit events for processing
event := core.NewEvent("analyzer", 
    core.EventData{"data": inputData}, 
    map[string]string{"session_id": "session-123"})
err = runner.Emit(event)
```

## Data Flow Architecture

Understanding how data flows through the system is crucial for building effective multi-agent systems:

```mermaid
sequenceDiagram
    participant Client
    participant Runner
    participant Callbacks
    participant Orchestrator
    participant Agent
    participant State
    participant Memory
    
    Client->>Runner: Emit Event
    Runner->>Callbacks: BeforeEventHandling
    Runner->>Orchestrator: Dispatch Event
    Orchestrator->>Callbacks: BeforeAgentRun
    Orchestrator->>Agent: HandleEvent(ctx, event, state)
    Agent->>State: Get/Set data and metadata
    Agent->>Memory: Store/Retrieve (optional)
    Memory-->>Agent: Return results
    Agent-->>Orchestrator: Return AgentResult
    Orchestrator->>Callbacks: AfterAgentRun
    Orchestrator-->>Runner: Return Result
    Runner->>Runner: Process routing (emit new events)
    Runner->>Callbacks: AfterEventHandling
```

::: info Key Flow Points
- **Events** trigger the processing pipeline
- **Callbacks** provide hooks at every stage
- **State** flows through agents and can be transformed
- **Routing** enables chaining and complex workflows
- **Memory** provides persistent storage and RAG capabilities
:::

## Core Patterns

### 1. Configuration-First Development

Start with configuration, then implement agents:

::: code-group

```toml [agentflow.toml]
[agent_flow]
name = "simple-assistant"

[llm]
provider = "ollama"
model = "gemma3:1b"
temperature = 0.7

[orchestration]
mode = "route"
timeout_seconds = 30

[agents.assistant]
role = "helpful_assistant"
system_prompt = "You are a helpful AI assistant."
enabled = true
```

```go [main.go]
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create runner with configuration
    config := core.DefaultRunnerConfig()
    runner := core.NewRunnerWithConfig(config)
    
    // Register agent handler
    err = runner.RegisterAgent("assistant", core.AgentHandlerFunc(
        func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            // Get message from event data
            message, ok := event.GetData()["message"].(string)
            if !ok {
                return core.AgentResult{}, fmt.Errorf("no message found")
            }
            
            // Create output state with response
            outputState := state.Clone()
            outputState.Set("response", fmt.Sprintf("Hello! You said: %s", message))
            outputState.SetMeta("processed_by", "assistant")
            
            return core.AgentResult{OutputState: outputState}, nil
        }))
    if err != nil {
        log.Fatal(err)
    }
    
    // Start → Emit → Stop lifecycle
    ctx := context.Background()
    err = runner.Start(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer runner.Stop()
    
    // Create and emit event
    event := core.NewEvent("assistant", 
        core.EventData{"message": "Hello, AgenticGoKit!"}, 
        map[string]string{
            "session_id": "demo-session",
            "route": "assistant",
        })
    
    err = runner.Emit(event)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Event emitted successfully!")
}
```

:::

### 2. Multi-Agent Orchestration

Configure multiple agents to work together:

::: code-group

```toml [agentflow.toml]
[orchestration]
mode = "collaborative"
timeout_seconds = 45
collaborative_agents = ["analyzer", "validator", "reporter"]

[agents.analyzer]
role = "data_analyzer"
system_prompt = "Analyze the provided data for insights."
enabled = true

[agents.validator]
role = "data_validator"
system_prompt = "Validate data quality and accuracy."
enabled = true

[agents.reporter]
role = "report_generator"
system_prompt = "Generate comprehensive reports."
enabled = true
```

```go [collaborative.go]
func setupCollaborativeSystem() {
    config := core.DefaultRunnerConfig()
    runner := core.NewRunnerWithConfig(config)
    
    // Register all agents
    runner.RegisterAgent("analyzer", &AnalyzerAgent{})
    runner.RegisterAgent("validator", &ValidatorAgent{})
    runner.RegisterAgent("reporter", &ReporterAgent{})
    
    // Add callback to monitor results
    runner.RegisterCallback(core.HookAfterAgentRun, "monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            fmt.Printf("Agent %s completed processing\n", args.AgentID)
            return args.State, nil
        })
    
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // All agents will process this event in parallel
    event := core.NewEvent("analyzer", 
        core.EventData{"dataset": "sales_data.csv"}, 
        map[string]string{"session_id": "analysis-session"})
    
    runner.Emit(event)
}
```

:::

### 3. Agent Composition and Capabilities

Build sophisticated agents with multiple capabilities:

```go
// Create a composable agent with multiple capabilities
capabilities := map[core.CapabilityType]core.AgentCapability{
    core.CapabilityTypeLLM:     llmCapability,
    core.CapabilityTypeMemory:  memoryCapability,
    core.CapabilityTypeMetrics: metricsCapability,
}

agent := core.NewComposableAgent("advanced-agent", capabilities, handler)

// Configure from resolved configuration
config := &core.ResolvedAgentConfig{
    Name:         "advanced-agent",
    Role:         "research_specialist",
    SystemPrompt: "You are an advanced research agent with memory and metrics.",
    Capabilities: []string{"research", "analysis", "memory"},
    Enabled:      true,
}

agent = core.NewComposableAgentWithConfig("advanced-agent", config, capabilities, handler)
```

## Advanced Features

### Memory and RAG Integration

Configure persistent memory and retrieval-augmented generation:

::: code-group

```toml [agentflow.toml]
[agent_memory]
provider = "pgvector"
connection = "postgres://user:pass@localhost/agentflow"
enable_rag = true
embedding_model = "text-embedding-ada-002"
```

```go [memory.go]
// Memory is configured automatically from agentflow.toml
// Access through agent capabilities or direct factory

memoryConfig := core.AgentMemoryConfig{
    Provider:   "pgvector",
    Connection: "postgres://user:pass@localhost/agentflow",
    EnableRAG:  true,
}

memory, err := core.NewMemory(memoryConfig)
if err != nil {
    log.Fatal(err)
}

// Store information with metadata
err = memory.Store(ctx, core.MemoryItem{
    Content:   "AgenticGoKit supports multiple orchestration patterns",
    Tags:      []string{"documentation", "orchestration"},
    Metadata:  map[string]string{"source": "core-concepts"},
})

// Search with semantic similarity
results, err := memory.Search(ctx, "orchestration patterns")
```

:::

### Error Handling and Recovery

Built-in error routing and recovery mechanisms:

```go
// Register error handling callbacks
runner.RegisterCallback(core.HookAgentError, "error-recovery",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        log.Printf("Agent %s failed: %v", args.AgentID, args.Error)
        
        // Implement recovery logic
        if isRetryableError(args.Error) {
            // Emit retry event
            retryEvent := core.NewEvent(args.AgentID,
                args.Event.GetData(),
                map[string]string{
                    "retry": "true",
                    "session_id": args.Event.GetSessionID(),
                })
            runner.Emit(retryEvent)
        }
        
        return args.State, nil
    })

// Configure error routing in agentflow.toml
// [error_routing]
// enabled = true
// max_retries = 3
// retry_delay_ms = 1000
```

### Monitoring and Observability

Built-in tracing and metrics collection:

```go
// Enable tracing
traceLogger := core.NewInMemoryTraceLogger()
runner.SetTraceLogger(traceLogger)

// Register trace hooks
err := core.RegisterTraceHooks(runner.GetCallbackRegistry(), traceLogger)

// Dump traces for analysis
traces, err := runner.DumpTrace("session-id")
for _, trace := range traces {
    fmt.Printf("Trace: %s at %s\n", trace.Type, trace.Timestamp)
}
```

## Learning Path

Now that you understand the core concepts, follow this recommended learning path:

::: tip Recommended Order
1. **[Message Passing](./message-passing.md)** - Master event-driven communication
2. **[State Management](./state-management.md)** - Learn data flow and persistence  
3. **[Agent Lifecycle](./agent-lifecycle.md)** - Understand agent creation and management
4. **[Orchestration Patterns](./orchestration-patterns.md)** - Coordinate multiple agents
5. **[Error Handling](./error-handling.md)** - Build resilient systems
:::

### Quick Reference

| Component | Purpose | Key Interface |
|-----------|---------|---------------|
| **Agent** | Processing unit | `Agent` with `Run()` and `HandleEvent()` |
| **Event** | Message carrier | `Event` with data and metadata |
| **State** | Data container | `State` with thread-safe operations |
| **Runner** | Event processor | `Runner` with lifecycle management |
| **Config** | System setup | `agentflow.toml` configuration |

### Common Patterns

- **Configuration-First**: Define system in `agentflow.toml`, implement handlers
- **Event-Driven**: Use events for all agent communication
- **State Transformation**: Agents transform input state to output state
- **Callback Hooks**: Monitor and extend system behavior
- **Composable Agents**: Build complex agents from capabilities

## Key Takeaways

::: info Core Principles
- **Configuration-Driven**: Everything starts with `agentflow.toml`
- **Unified Interface**: All agents implement the same `Agent` interface
- **Thread-Safe**: State and events are designed for concurrent access
- **Event-Based**: Communication happens through structured events
- **Composable**: Build complex systems from simple, reusable components
- **Observable**: Built-in tracing, metrics, and error handling
:::

### Architecture Benefits

- **Scalability**: Event-driven architecture supports high throughput
- **Flexibility**: Multiple orchestration patterns for different use cases
- **Reliability**: Comprehensive error handling and recovery mechanisms
- **Maintainability**: Clear separation of concerns and configuration
- **Extensibility**: Plugin architecture and capability system
- **Observability**: Built-in tracing, logging, and metrics

Understanding these concepts provides the foundation for building sophisticated, production-ready multi-agent systems with AgenticGoKit.

## What's Next?

- **Start Building**: Try the [Getting Started Guide](../getting-started/quickstart.md)
- **Explore Examples**: Check out [practical examples](../../examples/)
- **Join Community**: Connect with other developers
- **Contribute**: Help improve AgenticGoKit
## Practical Examples and Use Cases

### Example 1: Simple Chat Assistant

A basic chat assistant that processes user messages:

::: code-group

```go [main.go]
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create runner
    runner := core.NewRunner(100)
    
    // Create simple chat agent
    chatAgent := core.AgentHandlerFunc(
        func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            startTime := time.Now()
            
            // Get message from event
            message, ok := event.GetData()["message"].(string)
            if !ok {
                return core.AgentResult{}, fmt.Errorf("no message in event")
            }
            
            // Simple response logic
            response := fmt.Sprintf("You said: %s. How can I help you further?", message)
            
            // Create output state
            outputState := state.Clone()
            outputState.Set("response", response)
            outputState.Set("timestamp", time.Now().Unix())
            outputState.SetMeta("agent", "chat-assistant")
            
            return core.AgentResult{
                OutputState: outputState,
                StartTime:   startTime,
                EndTime:     time.Now(),
                Duration:    time.Since(startTime),
            }, nil
        })
    
    // Register agent
    err := runner.RegisterAgent("chat", chatAgent)
    if err != nil {
        log.Fatal(err)
    }
    
    // Start runner
    ctx := context.Background()
    err = runner.Start(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer runner.Stop()
    
    // Send a message
    event := core.NewEvent("chat",
        core.EventData{"message": "Hello, how are you?"},
        map[string]string{
            "session_id": "user-123",
            "route": "chat",
        })
    
    err = runner.Emit(event)
    if err != nil {
        log.Fatal(err)
    }
    
    // Wait for processing
    time.Sleep(1 * time.Second)
    fmt.Println("Chat message processed!")
}
```

:::

### Example 2: Data Processing Pipeline

A multi-stage data processing system:

::: code-group

```go [pipeline.go]
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    runner := core.NewRunner(100)
    
    // Stage 1: Data Collector
    collector := core.AgentHandlerFunc(
        func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            startTime := time.Now()
            
            // Simulate data collection
            rawData := []string{"apple", "banana", "cherry", "date"}
            
            outputState := state.Clone()
            outputState.Set("raw_data", rawData)
            outputState.Set("stage", "collected")
            outputState.SetMeta("collector", "fruit-collector")
            
            // Emit to next stage
            go func() {
                time.Sleep(100 * time.Millisecond)
                nextEvent := core.NewEvent("processor",
                    core.EventData{"raw_data": rawData},
                    map[string]string{
                        "session_id": event.GetSessionID(),
                        "pipeline": "data-processing",
                        "stage": "processing",
                    })
                runner.Emit(nextEvent)
            }()
            
            return core.AgentResult{
                OutputState: outputState,
                StartTime:   startTime,
                EndTime:     time.Now(),
                Duration:    time.Since(startTime),
            }, nil
        })
    
    // Stage 2: Data Processor
    processor := core.AgentHandlerFunc(
        func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            startTime := time.Now()
            
            // Get data from event
            rawData, ok := event.GetData()["raw_data"].([]string)
            if !ok {
                return core.AgentResult{}, fmt.Errorf("no raw data found")
            }
            
            // Process data (uppercase)
            processedData := make([]string, len(rawData))
            for i, item := range rawData {
                processedData[i] = strings.ToUpper(item)
            }
            
            outputState := state.Clone()
            outputState.Set("processed_data", processedData)
            outputState.Set("stage", "processed")
            outputState.SetMeta("processor", "text-processor")
            
            // Emit to final stage
            go func() {
                time.Sleep(100 * time.Millisecond)
                nextEvent := core.NewEvent("formatter",
                    core.EventData{"processed_data": processedData},
                    map[string]string{
                        "session_id": event.GetSessionID(),
                        "pipeline": "data-processing",
                        "stage": "formatting",
                    })
                runner.Emit(nextEvent)
            }()
            
            return core.AgentResult{
                OutputState: outputState,
                StartTime:   startTime,
                EndTime:     time.Now(),
                Duration:    time.Since(startTime),
            }, nil
        })
    
    // Stage 3: Data Formatter
    formatter := core.AgentHandlerFunc(
        func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            startTime := time.Now()
            
            // Get processed data
            processedData, ok := event.GetData()["processed_data"].([]string)
            if !ok {
                return core.AgentResult{}, fmt.Errorf("no processed data found")
            }
            
            // Format as JSON-like string
            formatted := fmt.Sprintf(`{"fruits": ["%s"]}`, strings.Join(processedData, `", "`))
            
            outputState := state.Clone()
            outputState.Set("formatted_data", formatted)
            outputState.Set("stage", "complete")
            outputState.SetMeta("formatter", "json-formatter")
            
            fmt.Printf("Pipeline complete! Result: %s\n", formatted)
            
            return core.AgentResult{
                OutputState: outputState,
                StartTime:   startTime,
                EndTime:     time.Now(),
                Duration:    time.Since(startTime),
            }, nil
        })
    
    // Register all agents
    runner.RegisterAgent("collector", collector)
    runner.RegisterAgent("processor", processor)
    runner.RegisterAgent("formatter", formatter)
    
    // Start pipeline
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Trigger pipeline
    event := core.NewEvent("collector",
        core.EventData{"trigger": "start"},
        map[string]string{
            "session_id": "pipeline-001",
            "pipeline": "data-processing",
        })
    
    runner.Emit(event)
    
    // Wait for completion
    time.Sleep(2 * time.Second)
}
```

:::

### Example 3: Error Handling and Recovery

Robust error handling with retry mechanisms:

::: code-group

```go [error-handling.go]
package main

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "strconv"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    runner := core.NewRunner(100)
    
    // Unreliable agent that sometimes fails
    unreliableAgent := core.AgentHandlerFunc(
        func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            startTime := time.Now()
            
            // Simulate random failures
            if rand.Float64() < 0.6 { // 60% failure rate
                return core.AgentResult{}, fmt.Errorf("simulated failure")
            }
            
            // Success case
            outputState := state.Clone()
            outputState.Set("result", "Success!")
            outputState.Set("processed_at", time.Now().Unix())
            
            return core.AgentResult{
                OutputState: outputState,
                StartTime:   startTime,
                EndTime:     time.Now(),
                Duration:    time.Since(startTime),
            }, nil
        })
    
    // Register error handling callback
    runner.RegisterCallback(core.HookAgentError, "retry-handler",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            // Get retry count
            retryCount := 0
            if countStr, exists := args.Event.GetMetadata()["retry_count"]; exists {
                if count, err := strconv.Atoi(countStr); err == nil {
                    retryCount = count
                }
            }
            
            maxRetries := 3
            if retryCount < maxRetries {
                // Create retry event
                retryMetadata := make(map[string]string)
                for k, v := range args.Event.GetMetadata() {
                    retryMetadata[k] = v
                }
                retryMetadata["retry_count"] = strconv.Itoa(retryCount + 1)
                retryMetadata["is_retry"] = "true"
                
                retryEvent := core.NewEvent(
                    args.Event.GetTargetAgentID(),
                    args.Event.GetData(),
                    retryMetadata,
                )
                
                // Schedule retry with backoff
                backoffDelay := time.Duration(retryCount+1) * time.Second
                go func() {
                    fmt.Printf("Retrying in %v (attempt %d/%d)\n", backoffDelay, retryCount+1, maxRetries)
                    time.Sleep(backoffDelay)
                    runner.Emit(retryEvent)
                }()
                
                // Update state with retry info
                retryState := args.State.Clone()
                retryState.SetMeta("retry_scheduled", "true")
                retryState.SetMeta("retry_count", strconv.Itoa(retryCount+1))
                
                return retryState, nil
            }
            
            // Max retries exceeded
            fmt.Printf("Max retries exceeded for event %s\n", args.Event.GetID())
            failedState := args.State.Clone()
            failedState.Set("failed", true)
            failedState.SetMeta("max_retries_exceeded", "true")
            
            return failedState, nil
        })
    
    // Register success callback
    runner.RegisterCallback(core.HookAfterAgentRun, "success-handler",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            if retryCount, exists := args.Event.GetMetadata()["retry_count"]; exists {
                fmt.Printf("Success after %s retries!\n", retryCount)
            } else {
                fmt.Println("Success on first attempt!")
            }
            return args.State, nil
        })
    
    // Register agent and start
    runner.RegisterAgent("unreliable", unreliableAgent)
    
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Send test event
    event := core.NewEvent("unreliable",
        core.EventData{"task": "process_data"},
        map[string]string{
            "session_id": "error-test-001",
        })
    
    runner.Emit(event)
    
    // Wait for completion (including retries)
    time.Sleep(10 * time.Second)
}
```

:::

### Example 4: State Management Patterns

Advanced state management with validation and transformation:

::: code-group

```go [state-patterns.go]
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

// Custom state validator
func validateUserData(state core.State) error {
    // Check required fields
    requiredFields := []string{"user_id", "email", "name"}
    for _, field := range requiredFields {
        if _, ok := state.Get(field); !ok {
            return fmt.Errorf("missing required field: %s", field)
        }
    }
    
    // Validate email format (simplified)
    if email, ok := state.Get("email"); ok {
        if emailStr, ok := email.(string); ok {
            if len(emailStr) == 0 || !contains(emailStr, "@") {
                return fmt.Errorf("invalid email format")
            }
        }
    }
    
    return nil
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && s[len(s)-len(substr):] != substr[:len(substr)]
}

func main() {
    runner := core.NewRunner(100)
    
    // Validator agent
    validator := core.AgentHandlerFunc(
        func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            startTime := time.Now()
            
            // Merge event data into state
            for key, value := range event.GetData() {
                state.Set(key, value)
            }
            
            // Validate state
            if err := validateUserData(state); err != nil {
                return core.AgentResult{}, fmt.Errorf("validation failed: %w", err)
            }
            
            // Add validation metadata
            outputState := state.Clone()
            outputState.SetMeta("validated", "true")
            outputState.SetMeta("validated_at", time.Now().Format(time.RFC3339))
            outputState.SetMeta("validator", "user-data-validator")
            
            fmt.Println("User data validated successfully")
            
            return core.AgentResult{
                OutputState: outputState,
                StartTime:   startTime,
                EndTime:     time.Now(),
                Duration:    time.Since(startTime),
            }, nil
        })
    
    // Enricher agent
    enricher := core.AgentHandlerFunc(
        func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            startTime := time.Now()
            
            // Get user data from state
            userID, _ := state.Get("user_id")
            name, _ := state.Get("name")
            
            // Enrich with additional data
            outputState := state.Clone()
            outputState.Set("full_profile", map[string]interface{}{
                "user_id": userID,
                "name": name,
                "created_at": time.Now().Unix(),
                "status": "active",
                "preferences": map[string]string{
                    "theme": "dark",
                    "language": "en",
                },
            })
            
            // Add enrichment metadata
            outputState.SetMeta("enriched", "true")
            outputState.SetMeta("enriched_at", time.Now().Format(time.RFC3339))
            outputState.SetMeta("enricher", "profile-enricher")
            
            fmt.Println("User profile enriched")
            
            return core.AgentResult{
                OutputState: outputState,
                StartTime:   startTime,
                EndTime:     time.Now(),
                Duration:    time.Since(startTime),
            }, nil
        })
    
    // Register agents
    runner.RegisterAgent("validator", validator)
    runner.RegisterAgent("enricher", enricher)
    
    // Add callback to chain processing
    runner.RegisterCallback(core.HookAfterAgentRun, "chain-processor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            if args.AgentID == "validator" {
                // After validation, trigger enrichment
                enrichEvent := core.NewEvent("enricher",
                    core.EventData{}, // Empty data, enricher will use state
                    map[string]string{
                        "session_id": args.Event.GetSessionID(),
                        "chain": "validation-enrichment",
                    })
                
                go func() {
                    time.Sleep(100 * time.Millisecond)
                    runner.Emit(enrichEvent)
                }()
            }
            
            return args.State, nil
        })
    
    // Start system
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Test with valid data
    validEvent := core.NewEvent("validator",
        core.EventData{
            "user_id": "user-123",
            "email": "user@example.com",
            "name": "John Doe",
        },
        map[string]string{
            "session_id": "state-test-001",
        })
    
    runner.Emit(validEvent)
    
    // Test with invalid data
    time.Sleep(1 * time.Second)
    
    invalidEvent := core.NewEvent("validator",
        core.EventData{
            "user_id": "user-456",
            "email": "invalid-email", // Missing @
            "name": "Jane Doe",
        },
        map[string]string{
            "session_id": "state-test-002",
        })
    
    runner.Emit(invalidEvent)
    
    // Wait for processing
    time.Sleep(3 * time.Second)
}
```

:::

## Common Pitfalls and Troubleshooting

### Pitfall 1: State Mutation Issues

::: danger Avoid Direct State Mutation
```go
// ❌ Wrong: Mutating shared state directly
func badAgent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    state.Set("result", "modified") // This affects other agents!
    return core.AgentResult{OutputState: state}, nil
}

// ✅ Correct: Always clone state before modification
func goodAgent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    outputState := state.Clone()
    outputState.Set("result", "modified")
    return core.AgentResult{OutputState: outputState}, nil
}
```
:::

### Pitfall 2: Blocking Operations in Agents

::: danger Avoid Blocking Operations
```go
// ❌ Wrong: Blocking operation without timeout
func blockingAgent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    time.Sleep(30 * time.Second) // Blocks entire system!
    return core.AgentResult{}, nil
}

// ✅ Correct: Use context timeout
func timeoutAgent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    select {
    case <-time.After(1 * time.Second):
        // Do work
        return core.AgentResult{OutputState: state.Clone()}, nil
    case <-ctx.Done():
        return core.AgentResult{}, ctx.Err()
    }
}
```
:::

### Pitfall 3: Missing Error Handling

::: danger Handle All Errors
```go
// ❌ Wrong: Ignoring errors
func unreliableAgent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    result, _ := riskyOperation() // Ignoring error!
    outputState := state.Clone()
    outputState.Set("result", result)
    return core.AgentResult{OutputState: outputState}, nil
}

// ✅ Correct: Proper error handling
func reliableAgent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    result, err := riskyOperation()
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("operation failed: %w", err)
    }
    
    outputState := state.Clone()
    outputState.Set("result", result)
    return core.AgentResult{OutputState: outputState}, nil
}

func riskyOperation() (string, error) {
    return "success", nil // Placeholder
}
```
:::

## Performance Considerations

### Memory Management

- **Clone Wisely**: Only clone state when necessary
- **Clean Up**: Remove unused data from state
- **Limit Size**: Keep event data and state reasonably sized

### Concurrency

- **Thread Safety**: Always use state.Clone() before modification
- **Avoid Locks**: Don't use locks in agent handlers
- **Context Cancellation**: Respect context timeouts

### Monitoring

- **Add Metrics**: Use metadata to track performance
- **Log Appropriately**: Log errors and important events
- **Trace Execution**: Use callback hooks for monitoring

## Real-World Use Cases

### 1. Document Processing System

- **Collector**: Receives documents from various sources
- **Parser**: Extracts text and metadata
- **Analyzer**: Performs content analysis
- **Indexer**: Stores processed documents

### 2. Customer Support Automation

- **Classifier**: Categorizes incoming requests
- **Router**: Routes to appropriate specialist agents
- **Responder**: Generates responses based on knowledge base
- **Escalator**: Handles complex cases requiring human intervention

### 3. Data Pipeline Orchestration

- **Extractor**: Pulls data from multiple sources
- **Transformer**: Cleans and normalizes data
- **Validator**: Ensures data quality
- **Loader**: Stores data in target systems

### 4. Multi-Modal AI Assistant

- **Text Processor**: Handles text-based queries
- **Image Analyzer**: Processes image inputs
- **Audio Transcriber**: Converts speech to text
- **Response Synthesizer**: Combines results into coherent responses

Each use case demonstrates different orchestration patterns, error handling strategies, and state management approaches suitable for specific requirements.

## Quick Reference

### Core Interfaces

| Interface | Purpose | Key Methods |
|-----------|---------|-------------|
| `Agent` | Full-featured agent | `Run()`, `HandleEvent()`, `Initialize()`, `Shutdown()` |
| `AgentHandler` | Lightweight handler | `Run()` |
| `Runner` | Event orchestration | `Emit()`, `RegisterAgent()`, `Start()`, `Stop()` |
| `State` | Thread-safe data | `Get()`, `Set()`, `Clone()`, `GetMeta()`, `SetMeta()` |
| `Event` | Message container | `GetData()`, `GetMetadata()`, `GetID()`, `GetSessionID()` |

### Factory Functions

| Function | Purpose | Returns |
|----------|---------|---------|
| `core.NewRunner(size)` | Basic runner | `*RunnerImpl` |
| `core.NewRunnerWithConfig(config)` | Configured runner | `Runner` |
| `core.NewState()` | Empty state | `State` |
| `core.NewEvent(target, data, meta)` | Event | `Event` |
| `core.AgentHandlerFunc(fn)` | Function wrapper | `AgentHandler` |

### Common Patterns

```go
// Agent Registration
runner.RegisterAgent("name", agent)
runner.RegisterAgent("name", core.AgentHandlerFunc(fn))

// State Operations
outputState := inputState.Clone()
outputState.Set("key", value)
outputState.SetMeta("route", "next-agent")

// Event Creation
event := core.NewEvent("target-agent", 
    core.EventData{"key": "value"},
    map[string]string{"session_id": "123"})

// Error Handling
if err != nil {
    return core.AgentResult{}, fmt.Errorf("operation failed: %w", err)
}
```

## Navigation Guide

### Learning Path

1. **Start Here**: [Core Concepts Overview](./README.md) - Architecture and interfaces
2. **Agent Basics**: [Agent Lifecycle](./agent-lifecycle.md) - Creating and managing agents
3. **Communication**: [Message Passing](./message-passing.md) - Events and Runner orchestration
4. **Data Flow**: [State Management](./state-management.md) - Thread-safe data handling
5. **Coordination**: [Orchestration Patterns](./orchestration-patterns.md) - Multi-agent patterns
6. **Resilience**: [Error Handling](./error-handling.md) - Fault tolerance and recovery

### By Use Case

- **Simple Processing**: Start with [Agent Lifecycle](./agent-lifecycle.md) → [State Management](./state-management.md)
- **Event-Driven Systems**: Focus on [Message Passing](./message-passing.md) → [Orchestration Patterns](./orchestration-patterns.md)
- **Production Systems**: Cover all topics, emphasizing [Error Handling](./error-handling.md)

### By Experience Level

- **Beginner**: README → Agent Lifecycle → State Management
- **Intermediate**: Message Passing → Orchestration Patterns → Error Handling  
- **Advanced**: All topics with focus on performance and production patterns