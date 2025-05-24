# AgentFlow Library Usage Guide

## Overview

AgentFlow is now available as a fully-featured Go library that you can import and use in your own projects. This guide provides comprehensive examples and best practices for using AgentFlow as an external dependency.

## Installation

Add AgentFlow to your Go project:

```bash
go get github.com/kunalkushwaha/agentflow@latest
```

## Quick Start

### Basic Agent Implementation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    agentflow "github.com/kunalkushwaha/agentflow/core"
)

// SimpleAgent demonstrates basic agent implementation
type SimpleAgent struct {
    name string
}

func NewSimpleAgent(name string) *SimpleAgent {
    return &SimpleAgent{name: name}
}

func (a *SimpleAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    // Log agent execution
    agentflow.Logger().Info().
        Str("agent", a.name).
        Str("event_id", event.GetID()).
        Msg("Processing event")

    // Get data from event
    eventData := event.GetData()
    message, ok := eventData["message"]
    if !ok {
        message = "No message provided"
    }

    // Process the message
    response := fmt.Sprintf("%s processed: %v", a.name, message)

    // Create output state
    outputState := agentflow.NewState()
    outputState.Set("response", response)
    outputState.Set("processed_by", a.name)
    outputState.Set("timestamp", time.Now().Format(time.RFC3339))

    // Copy input state to output
    for _, key := range state.Keys() {
        if value, exists := state.Get(key); exists {
            outputState.Set(key, value)
        }
    }

    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   time.Now(),
        EndTime:     time.Now(),
        Duration:    time.Millisecond * 10,
    }, nil
}

func main() {
    // Set log level
    agentflow.SetLogLevel(agentflow.INFO)

    // Create agents
    agents := map[string]agentflow.AgentHandler{
        "processor": NewSimpleAgent("ProcessorAgent"),
    }

    // Create runner with configuration
    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:    agents,
        QueueSize: 10,
    })

    // Start the runner
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        log.Fatalf("Failed to start runner: %v", err)
    }
    defer runner.Stop()

    fmt.Println("AgentFlow runner started successfully!")

    // Create and emit event
    eventData := agentflow.EventData{
        "message": "Hello from AgentFlow library!",
        "type":    "test_message",
    }

    metadata := map[string]string{
        agentflow.RouteMetadataKey: "processor",
        "session_id":               "library-test-session",
    }

    event := agentflow.NewEvent("processor", eventData, metadata)

    fmt.Printf("Emitting event: %s\n", event.GetID())
    if err := runner.Emit(event); err != nil {
        log.Fatalf("Failed to emit event: %v", err)
    }

    // Wait for processing
    time.Sleep(time.Second * 2)
    fmt.Println("Event processing completed!")
}
```

## Advanced Features

### Agent Chaining

Chain multiple agents together to create complex workflows:

```go
// ChainedAgent demonstrates agent chaining
type ChainedAgent struct {
    name     string
    nextAgent string
    step     int
}

func (a *ChainedAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    agentflow.Logger().Info().
        Str("agent", a.name).
        Str("event_id", event.GetID()).
        Int("step", a.step).
        Msg("ChainedAgent processing")

    // Process current step
    outputState := state.Clone()
    outputState.Set(fmt.Sprintf("step_%d_completed", a.step), true)
    outputState.Set("current_step", a.step)

    result := agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   time.Now(),
        EndTime:     time.Now(),
        Duration:    time.Millisecond * 100,
    }

    // Chain to next agent if specified
    if a.nextAgent != "" && a.step < 3 {
        agentflow.Logger().Info().
            Str("agent", a.name).
            Str("next_agent", a.nextAgent).
            Int("step", a.step).
            Msg("Routing to next agent")
        
        result.OutputState.SetMeta(agentflow.RouteMetadataKey, a.nextAgent)
    } else {
        agentflow.Logger().Info().
            Str("agent", a.name).
            Int("final_step", a.step).
            Msg("Chain completed")
    }

    return result, nil
}

// Usage
func createChainedAgents() map[string]agentflow.AgentHandler {
    return map[string]agentflow.AgentHandler{
        "step1": &ChainedAgent{name: "Step1Agent", nextAgent: "step2", step: 1},
        "step2": &ChainedAgent{name: "Step2Agent", nextAgent: "step3", step: 2},
        "step3": &ChainedAgent{name: "Step3Agent", nextAgent: "", step: 3},
    }
}
```

### Error Handling

Implement robust error handling and recovery:

```go
type ErrorHandlingAgent struct {
    shouldError bool
}

func (a *ErrorHandlingAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    if a.shouldError {
        return agentflow.AgentResult{}, fmt.Errorf("simulated agent error")
    }

    // Normal processing
    outputState := agentflow.NewState()
    outputState.Set("success", true)
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   time.Now(),
        EndTime:     time.Now(),
        Duration:    time.Millisecond * 50,
    }, nil
}

// Error handler agent
type ErrorHandlerAgent struct{}

func (a *ErrorHandlerAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    agentflow.Logger().Error().
        Str("event_id", event.GetID()).
        Msg("ErrorHandlerAgent handling error")

    outputState := agentflow.NewState()
    outputState.Set("error_handled", true)
    outputState.Set("recovery_time", time.Now().Format(time.RFC3339))

    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   time.Now(),
        EndTime:     time.Now(),
        Duration:    time.Millisecond * 5,
    }, nil
}
```

### Tracing and Observability

Enable comprehensive tracing for debugging and monitoring:

```go
func createRunnerWithTracing() *agentflow.Runner {
    // Create in-memory trace logger
    traceLogger := agentflow.NewInMemoryTraceLogger()

    // Create runner with tracing enabled
    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:      agents,
        QueueSize:   10,
        TraceLogger: traceLogger,
    })

    return runner
}

func analyzeTraces(runner *agentflow.Runner, sessionID string) {
    traces, err := runner.DumpTrace(sessionID)
    if err != nil {
        fmt.Printf("Error getting traces: %v\n", err)
        return
    }

    fmt.Printf("Found %d trace entries:\n", len(traces))
    for i, trace := range traces {
        fmt.Printf("  %d. %s: Type=%s, EventID=%s, AgentID=%s\n",
            i+1,
            trace.Timestamp.Format("15:04:05.000"),
            trace.Type,
            trace.EventID,
            trace.AgentID)
    }
}
```

### Concurrent Processing

Handle multiple events concurrently:

```go
func concurrentProcessing(runner *agentflow.Runner) {
    var wg sync.WaitGroup
    
    // Emit multiple events concurrently
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            eventData := agentflow.EventData{
                "message": fmt.Sprintf("Concurrent event %d", id),
                "id":      id,
            }
            
            metadata := map[string]string{
                agentflow.RouteMetadataKey: "processor",
                "session_id":               fmt.Sprintf("concurrent-session-%d", id),
            }
            
            event := agentflow.NewEvent("processor", eventData, metadata)
            
            if err := runner.Emit(event); err != nil {
                fmt.Printf("Failed to emit event %d: %v\n", id, err)
            }
        }(i)
    }
    
    wg.Wait()
    fmt.Println("All concurrent events emitted")
}
```

## Configuration Options

### Runner Configuration

```go
runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
    Agents:    agents,           // Map of agent handlers
    QueueSize: 100,             // Event queue size
    TraceLogger: traceLogger,   // Optional trace logger
})
```

### Logging Configuration

```go
// Set log level
agentflow.SetLogLevel(agentflow.DEBUG) // DEBUG, INFO, WARN, ERROR

// Get logger for custom logging
logger := agentflow.Logger()
logger.Info().Str("key", "value").Msg("Custom log message")
```

## Best Practices

### 1. Agent Design

- Keep agents focused on single responsibilities
- Use meaningful names for agents and state keys
- Implement proper error handling and recovery
- Include comprehensive logging

### 2. State Management

- Clone state when making modifications
- Use consistent key naming conventions
- Include metadata for routing and debugging
- Keep state data serializable

### 3. Event Handling

- Use unique session IDs for tracing
- Include relevant metadata for routing
- Handle context cancellation appropriately
- Validate event data before processing

### 4. Performance

- Configure appropriate queue sizes
- Monitor trace logs for bottlenecks
- Use concurrent processing for independent events
- Implement circuit breakers for error-prone agents

### 5. Testing

- Create comprehensive test agents
- Use in-memory trace loggers for testing
- Test error scenarios and recovery
- Validate state transitions

## Production Considerations

### File-Based Tracing

```go
traceLogger := agentflow.NewFileTraceLogger("./production-traces")
runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
    Agents:      agents,
    QueueSize:   1000,
    TraceLogger: traceLogger,
})
```

### Monitoring and Metrics

```go
// Track custom metrics
type MetricsAgent struct {
    successCount int64
    errorCount   int64
}

func (a *MetricsAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Process event
    result, err := a.processEvent(ctx, event, state)
    
    // Track metrics
    duration := time.Since(startTime)
    if err != nil {
        atomic.AddInt64(&a.errorCount, 1)
        agentflow.Logger().Error().
            Err(err).
            Dur("duration", duration).
            Msg("Agent execution failed")
    } else {
        atomic.AddInt64(&a.successCount, 1)
        agentflow.Logger().Info().
            Dur("duration", duration).
            Msg("Agent execution succeeded")
    }
    
    return result, err
}
```

## Migration from Internal Usage

If you were using AgentFlow internally, update your imports:

**Old** (internal use):
```go
import "kunalkushwaha/agentflow/internal/core"
```

**New** (library use):
```go
import agentflow "github.com/kunalkushwaha/agentflow/core"
```

## Support and Community

- **Repository**: [github.com/kunalkushwaha/agentflow](https://github.com/kunalkushwaha/agentflow)
- **Examples**: See the `examples/` directory in the repository
- **Documentation**: Complete documentation in the `docs/` directory
- **Issues**: Report bugs or request features via GitHub Issues

## Conclusion

AgentFlow provides a powerful, flexible framework for building agent-based systems in Go. With proper design patterns and best practices, you can create sophisticated, scalable, and maintainable agent workflows that handle complex business logic with ease.

The library is production-ready and actively maintained. We welcome contributions and feedback from the community!
