# Message Passing and Event Flow in AgenticGoKit

## Overview

At the heart of AgenticGoKit is an event-driven architecture that enables flexible communication between agents. This tutorial explains how messages flow through the system, how the Runner orchestrates this flow, and how you can leverage these patterns in your own applications.

Understanding message passing is crucial because it's the foundation of how agents communicate, share data, and coordinate their work in AgenticGoKit.

## Prerequisites

- Basic understanding of Go programming
- Familiarity with interfaces and goroutines
- Completed the [5-minute quickstart](../getting-started/quickstart.md)

## Core Concepts

### Events: The Communication Currency

In AgenticGoKit, all communication happens through **Events**. An Event is more than just a message - it's a structured packet of information with routing metadata.

```go
// The Event interface - the core communication unit
type Event interface {
    GetID() string                 // Unique identifier
    GetTimestamp() time.Time       // When event was created
    GetTargetAgentID() string      // Destination agent
    GetSourceAgentID() string      // Source agent
    GetData() EventData            // Actual payload data
    GetMetadata() map[string]string // Routing metadata
    GetSessionID() string          // Session identifier
    // ... other methods for setting values
}
```

Events carry:
- **Data**: The actual payload (questions, responses, etc.)
- **Metadata**: Routing information, session IDs, etc.
- **Source/Target**: Which agent sent it and where it's going
- **Timestamps**: When the event was created

### EventData: The Payload

```go
// EventData holds the payload of an event
type EventData map[string]any

// Example usage
data := core.EventData{
    "message": "What's the weather today?",
    "user_id": "user-123",
    "priority": "high",
}
```

### Runners: The Traffic Controllers

The **Runner** is responsible for routing events to the appropriate agents and managing the overall flow of communication.

```go
// The Runner interface - manages event flow
type Runner interface {
    Emit(event Event) error        // Send an event into the system
    RegisterAgent(name string, handler AgentHandler) error // Register an agent
    Start(ctx context.Context) error // Start processing
    Stop()                         // Stop processing
    // ... other methods
}
```

The Runner:
1. Receives events via `Emit()`
2. Determines which agent(s) should handle the event
3. Delivers the event to the appropriate agent(s)
4. Collects results and routes them to the next destination

## How Message Passing Works

### 1. Creating Events

Events are typically created using the `NewEvent()` function:

```go
// Create a new event
event := core.NewEvent(
    "weather-agent",           // Target agent ID
    core.EventData{           // Payload data
        "message": "What's the weather in Paris?",
        "location": "Paris",
    },
    map[string]string{        // Metadata
        "session_id": "user-123",
        "priority": "normal",
    },
)
```

### 2. Emitting Events

Events are sent into the system using the Runner's `Emit()` method:

```go
// Send the event into the system
err := runner.Emit(event)
if err != nil {
    log.Fatalf("Failed to emit event: %v", err)
}
```

This is an **asynchronous operation** - `Emit()` returns immediately, and processing happens in the background.

### 3. Event Processing Flow

Behind the scenes, the Runner follows this flow:

```
┌─────────┐     ┌──────────┐     ┌─────────┐
│ Client  │────▶│  Runner  │────▶│ Agent A │
└─────────┘     └──────────┘     └─────────┘
                     │                │
                     │                ▼
                     │           ┌─────────┐
                     │◀──────────│ Result  │
                     │           └─────────┘
                     ▼
                ┌─────────┐
                │ Agent B │
                └─────────┘
```

1. **Queues** the event for processing
2. **Routes** it to the target agent(s) via the Orchestrator
3. **Collects** the results
4. **Forwards** results to the next agent or back to the caller

### 4. Handling Results

Results are typically handled through callbacks:

```go
// Register a callback for when an agent completes processing
runner.RegisterCallback(core.HookAfterAgentRun, "my-callback", 
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        fmt.Printf("Agent %s completed\n", args.AgentID)
        return args.State, nil
    },
)
```

## Under the Hood: The Runner Implementation (Internals)

The Runner implementation uses a combination of channels, goroutines, and queues to manage event flow:

```go
// Simplified version of the internal runner loop
func (r *RunnerImpl) loop(ctx context.Context) {
    defer r.wg.Done()
    for {
        select {
        case <-ctx.Done():
            return
        case <-r.stopChan:
            return
        case event := <-r.queue:
            // Process event in the main goroutine
            // internal: r.processEvent(ctx, event) — use runner.Emit(event) in public API
        }
    }
}

// internal: func (r *RunnerImpl) processEvent(ctx context.Context, event Event) {
    // 1. Invoke BeforeEventHandling callbacks
    // 2. Route to orchestrator for agent dispatch
    // 3. Handle agent result
    // 4. Invoke AfterEventHandling callbacks
    // 5. Emit new events if needed
}
```

## Practical Example: Building a Conversational Agent

Let's see how this works in practice with a simple conversational agent:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create a provider
    provider, err := core.NewOpenAIAdapter(
        os.Getenv("OPENAI_API_KEY"),
        "gpt-3.5-turbo",
        500,
        0.7,
    )
    if err != nil {
        panic(err)
    }
    
    // Create an agent
    agent, err := core.NewAgent("assistant").
        WithLLMAndConfig(provider, core.LLMConfig{
            SystemPrompt: "You are a helpful assistant.",
        }).
        Build()
    if err != nil {
        panic(err)
    }
    
    // Create a runner with orchestrator
    runner := core.NewRunner(100)
    orchestrator := core.NewRouteOrchestrator(runner.GetCallbackRegistry())
    runner.SetOrchestrator(orchestrator)
    
    // Register the agent
    agentHandler := core.ConvertAgentToHandler(agent)
    runner.RegisterAgent("assistant", agentHandler)
    
    // Start the runner
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        panic(err)
    }
    defer runner.Stop()
    
    // Set up result handling
    resultReceived := make(chan bool)
    runner.RegisterCallback(core.HookAfterAgentRun, "result-handler",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            if args.AgentResult.OutputState != nil {
                if response, ok := args.AgentResult.OutputState.Get("response"); ok {
                    fmt.Printf("Agent Response: %s\n", response)
                }
            }
            resultReceived <- true
            return args.State, nil
        },
    )
    
    // Create and emit an event
    event := core.NewEvent(
        "assistant",
        core.EventData{"message": "Tell me about AgenticGoKit"},
        map[string]string{
            "session_id": "user-123",
            "route": "assistant", // Required for routing
        },
    )
    
    if err := runner.Emit(event); err != nil {
        panic(err)
    }
    
    // Wait for result
    select {
    case <-resultReceived:
        fmt.Println("Processing complete!")
    case <-time.After(30 * time.Second):
        fmt.Println("Timeout waiting for response")
    }
}
```

## Advanced Message Passing Patterns

### 1. Session Management

Events can carry session IDs to maintain conversation context:

```go
event := core.NewEvent(
    "assistant",
    core.EventData{"message": "What was my last question?"},
    map[string]string{"session_id": "user-123"},
)
```

### 2. Event Chaining

You can create chains of events where each agent's output becomes the next agent's input:

```go
runner.RegisterCallback(core.HookAfterAgentRun, "chain-handler",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        if args.AgentID == "researcher" && args.Error == nil {
            // Create a new event for the analyzer
            nextEvent := core.NewEvent(
                "analyzer",
                core.EventData{"research_data": args.AgentResult.OutputState},
                map[string]string{
                    "session_id": args.Event.GetSessionID(),
                    "route": "analyzer",
                },
            )
            runner.Emit(nextEvent)
        }
        return args.State, nil
    },
)
```

### 3. Broadcast Events

Send the same event to multiple agents simultaneously using collaborative orchestration:

```go
// This requires using a collaborative orchestrator
// which we'll cover in the orchestration tutorial
event := core.NewEvent(
    "",  // Empty target for broadcast
    core.EventData{"message": "New data available"},
    map[string]string{"broadcast": "true"},
)
```

### 4. Priority Handling

Use metadata to implement priority queues:

```go
event := core.NewEvent(
    "processor",
    core.EventData{"task": "urgent_analysis"},
    map[string]string{
        "priority": "high",
        "deadline": time.Now().Add(5*time.Minute).Format(time.RFC3339),
    },
)
```

## Event Lifecycle and Hooks

AgenticGoKit provides several hooks where you can intercept and modify the event processing flow:

```go
// Available hooks
const (
    HookBeforeEventHandling // Before any processing
    HookBeforeAgentRun     // Before agent execution
    HookAfterAgentRun      // After successful agent execution
    HookAgentError         // When agent execution fails
    HookAfterEventHandling // After all processing
)
```

### Example: Adding Logging and Metrics

```go
// Add comprehensive logging
runner.RegisterCallback(core.HookBeforeEventHandling, "logger",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        fmt.Printf("[%s] Processing event %s\n", 
            time.Now().Format(time.RFC3339),
            args.Event.GetID(),
        )
        return args.State, nil
    },
)

runner.RegisterCallback(core.HookAfterAgentRun, "metrics",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        duration := time.Since(args.Event.GetTimestamp())
        fmt.Printf("Agent %s completed in %v\n", args.AgentID, duration)
        
        // Record metrics
        // metrics.RecordAgentDuration(args.AgentID, duration)
        
        return args.State, nil
    },
)
```

## Error Handling in Message Passing

When agents fail, AgenticGoKit provides sophisticated error routing:

```go
// Register error handler
runner.RegisterCallback(core.HookAgentError, "error-handler",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        fmt.Printf("Agent %s failed: %v\n", args.AgentID, args.Error)
        
        // Optionally emit a recovery event
        recoveryEvent := core.NewEvent(
            "error-recovery-agent",
            core.EventData{
                "original_event": args.Event,
                "error": args.Error.Error(),
                "failed_agent": args.AgentID,
            },
            map[string]string{
                "session_id": args.Event.GetSessionID(),
                "route": "error-recovery-agent",
            },
        )
        
        runner.Emit(recoveryEvent)
        return args.State, nil
    },
)
```

## Common Pitfalls and Solutions

### 1. Deadlocks

**Problem**: Waiting for results in the same goroutine that processes events.

**Solution**: Use callbacks or separate goroutines for waiting on results.

```go
// Bad - blocks the event loop
result := <-resultChannel

// Good - use callbacks
runner.RegisterCallback(core.HookAfterAgentRun, "handler", 
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        // Handle result here
        return args.State, nil
    },
)
```

### 2. Event Loops

**Problem**: Creating circular event chains that never terminate.

**Solution**: Implement loop detection or maximum hop counts in your metadata.

```go
// Add hop count to prevent infinite loops
metadata := map[string]string{
    "session_id": "user-123",
    "hop_count": "1",
    "max_hops": "5",
}
```

### 3. Lost Events

**Problem**: Events that never get processed because the target agent doesn't exist.

**Solution**: Implement error routing and default handlers for unknown targets.

```go
// The orchestrator will automatically handle unknown targets
// and emit error events that you can catch with error callbacks
```

### 4. Memory Leaks

**Problem**: Events accumulating in queues without being processed.

**Solution**: Monitor queue sizes and implement proper shutdown procedures.

```go
// Always stop the runner when done
defer runner.Stop()

// Monitor queue health
runner.RegisterCallback(core.HookBeforeEventHandling, "monitor",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        // Check queue sizes, memory usage, etc.
        return args.State, nil
    },
)
```

## Performance Considerations

### 1. Queue Sizing

```go
// Create runner with appropriate queue size
runner := core.NewRunner(1000) // Larger queue for high throughput
```

### 2. Event Batching

```go
// For high-volume scenarios, consider batching events
batchEvent := core.NewEvent(
    "batch-processor",
    core.EventData{
        "events": []core.Event{event1, event2, event3},
    },
    metadata,
)
```

### 3. Async Processing

```go
// Use goroutines for non-blocking operations
runner.RegisterCallback(core.HookAfterAgentRun, "async-handler",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        go func() {
            // Long-running operation
            processResult(args.AgentResult)
        }()
        return args.State, nil
    },
)
```

## Debugging Message Flow

### 1. Event Tracing

```go
// Enable detailed logging
runner.RegisterCallback(core.HookBeforeEventHandling, "tracer",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        fmt.Printf("Event %s: %+v\n", args.Event.GetID(), args.Event.GetData())
        return args.State, nil
    },
)
```

### 2. State Inspection

```go
// Inspect state at each step
runner.RegisterCallback(core.HookAfterAgentRun, "state-inspector",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        if args.AgentResult.OutputState != nil {
            fmt.Printf("State after %s: %+v\n", 
                args.AgentID, 
                args.AgentResult.OutputState.GetAll(),
            )
        }
        return args.State, nil
    },
)
```

## Best Practices

1. **Always set session IDs** for tracking related events
2. **Use meaningful event IDs** for debugging
3. **Include sufficient metadata** for routing and processing
4. **Handle errors gracefully** with error callbacks
5. **Monitor queue health** and processing times
6. **Use callbacks instead of blocking waits**
7. **Implement proper shutdown procedures**
8. **Add comprehensive logging** for production systems

## Conclusion

The event-driven architecture of AgenticGoKit provides a flexible foundation for building complex agent systems. By understanding how events flow through the Runner to agents and back, you can create sophisticated communication patterns between your agents.

Key takeaways:
- Events are the primary communication mechanism
- Runners manage event flow and routing
- Callbacks provide hooks for customization
- Proper error handling is crucial
- Async patterns prevent blocking

## Next Steps

- [Orchestration Patterns](../../guides/deployment/README.md) - Learn how different orchestration modes build on message passing
- [State Management](state-management.md) - Understand how data flows between agents
- [Error Handling](error-handling.md) - Master robust error management patterns
- [Debugging Guide](../debugging/README.md) - Learn to trace and debug event flows

## Further Reading

- [API Reference: Event Interface](../../reference/api/state-event.md#event)
- [API Reference: Runner Interface](../../reference/api/orchestration.md#runner)
- [Examples: Message Passing Patterns](../../examples/)
