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
    startTime := time.Now()
    
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

    // Create output state by cloning input state
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("processed_by", a.name)
    outputState.Set("timestamp", time.Now().Format(time.RFC3339))

    endTime := time.Now()
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     endTime,
        Duration:    endTime.Sub(startTime),
    }, nil
}

func main() {
    // Set log level
    agentflow.SetLogLevel(agentflow.INFO)

    // Create agents
    agents := map[string]agentflow.AgentHandler{
        "processor": NewSimpleAgent("ProcessorAgent"),
    }

    // Create runner with configuration and tracing
    traceLogger := agentflow.NewInMemoryTraceLogger()
    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:      agents,
        QueueSize:   10,
        TraceLogger: traceLogger,
    })

    // Start the runner
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        log.Fatalf("Failed to start runner: %v", err)
    }
    defer runner.Stop()

    fmt.Println("AgentFlow runner started successfully!")

    // Create and emit event with proper session handling
    sessionID := "library-test-session"
    eventData := agentflow.EventData{
        "message": "Hello from AgentFlow library!",
        "type":    "test_message",
    }

    metadata := map[string]string{
        agentflow.RouteMetadataKey: "processor",
        agentflow.SessionIDKey:     sessionID,
    }

    event := agentflow.NewEvent("processor", eventData, metadata)

    fmt.Printf("Emitting event: %s\n", event.GetID())
    if err := runner.Emit(event); err != nil {
        log.Fatalf("Failed to emit event: %v", err)
    }

    // Wait for processing
    time.Sleep(time.Second * 2)
    
    // Retrieve and display traces
    traces, err := runner.DumpTrace(sessionID)
    if err != nil {
        log.Printf("Error getting traces: %v", err)
    } else {
        fmt.Printf("Found %d trace entries for session %s\n", len(traces), sessionID)
    }
    
    fmt.Println("Event processing completed!")
}
```

## Advanced Features

### Circuit Breaker and Retry Logic

AgentFlow includes built-in circuit breaker and retry capabilities for resilient agent execution:

```go
// Agent with circuit breaker and retry configuration
type ResilientAgent struct {
    name string
}

func (a *ResilientAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Simulate potentially failing operation
    if shouldSimulateFailure() {
        return agentflow.AgentResult{}, fmt.Errorf("simulated failure for circuit breaker testing")
    }
    
    outputState := state.Clone()
    outputState.Set("processed_by", a.name)
    outputState.Set("success", true)
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}

// Create runner with circuit breaker configuration
func createResilientRunner() *agentflow.Runner {
    agents := map[string]agentflow.AgentHandler{
        "resilient": &ResilientAgent{name: "ResilientAgent"},
    }
    
    return agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:        agents,
        QueueSize:     10,
        TraceLogger:   agentflow.NewInMemoryTraceLogger(),
    })
}
```

### Responsible AI Integration

Implement responsible AI checks in your workflows:

```go
// ResponsibleAIAgent performs content safety and ethical checks
type ResponsibleAIAgent struct{}

func (a *ResponsibleAIAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Get content to check
    content, exists := event.GetData()["content"]
    if !exists {
        return agentflow.AgentResult{}, fmt.Errorf("no content provided for responsible AI check")
    }
    
    // Perform safety checks (placeholder implementation)
    if isContentUnsafe(content) {
        outputState := state.Clone()
        outputState.Set("responsible_ai_check", "failed")
        outputState.Set("reason", "unsafe content detected")
        outputState.SetMeta(agentflow.RouteMetadataKey, "error-handler")
        
        return agentflow.AgentResult{
            OutputState: outputState,
            StartTime:   startTime,
            EndTime:     time.Now(),
            Duration:    time.Since(startTime),
        }, nil
    }
    
    // Content is safe, continue processing
    outputState := state.Clone()
    outputState.Set("responsible_ai_check", "passed")
    outputState.SetMeta(agentflow.RouteMetadataKey, "content-processor")
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}

func isContentUnsafe(content interface{}) bool {
    // Implement your content safety logic here
    contentStr, ok := content.(string)
    if !ok {
        return true
    }
    // Simple example: check for harmful keywords
    harmfulKeywords := []string{"violence", "hate", "harmful"}
    for _, keyword := range harmfulKeywords {
        if strings.Contains(strings.ToLower(contentStr), keyword) {
            return true
        }
    }
    return false
}
```

### Enhanced Error Routing

Implement sophisticated error handling with routing:

```go
// ErrorRoutingAgent demonstrates enhanced error handling
type ErrorRoutingAgent struct{}

func (a *ErrorRoutingAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Get error information from state
    errorType, _ := state.Get("error_type")
    errorMessage, _ := state.Get("error_message")
    
    outputState := state.Clone()
    
    // Route based on error type
    switch errorType {
    case "validation_error":
        outputState.Set("recovery_action", "request_user_input")
        outputState.SetMeta(agentflow.RouteMetadataKey, "validation-handler")
    case "timeout_error":
        outputState.Set("recovery_action", "retry_with_backoff")
        outputState.SetMeta(agentflow.RouteMetadataKey, "retry-handler")
    case "critical_error":
        outputState.Set("recovery_action", "escalate_to_admin")
        outputState.SetMeta(agentflow.RouteMetadataKey, "admin-notifier")
    default:
        outputState.Set("recovery_action", "generic_error_handling")
    }
    
    outputState.Set("error_handled", true)
    outputState.Set("handled_by", "ErrorRoutingAgent")
    outputState.Set("original_error", errorMessage)
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}
```

### Workflow Validation

Implement workflow validation for complex agent chains:

```go
// WorkflowValidationAgent validates workflow state and transitions
type WorkflowValidationAgent struct{}

func (a *WorkflowValidationAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Validate required workflow fields
    requiredFields := []string{"workflow_id", "step", "user_id"}
    for _, field := range requiredFields {
        if _, exists := state.Get(field); !exists {
            return agentflow.AgentResult{}, fmt.Errorf("missing required field: %s", field)
        }
    }
    
    // Validate workflow step progression
    currentStep, _ := state.Get("step")
    if !isValidStepTransition(currentStep) {
        outputState := state.Clone()
        outputState.Set("validation_error", "invalid step transition")
        outputState.SetMeta(agentflow.RouteMetadataKey, "error-handler")
        
        return agentflow.AgentResult{
            OutputState: outputState,
            StartTime:   startTime,
            EndTime:     time.Now(),
            Duration:    time.Since(startTime),
        }, nil
    }
    
    // Validation passed
    outputState := state.Clone()
    outputState.Set("validation_status", "passed")
    outputState.Set("validated_at", time.Now().Format(time.RFC3339))
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}

func isValidStepTransition(step interface{}) bool {
    // Implement your workflow validation logic
    stepStr, ok := step.(string)
    if !ok {
        return false
    }
    validSteps := map[string]bool{
        "init": true, "process": true, "validate": true, "complete": true,
    }
    return validSteps[stepStr]
}
```

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
// Comprehensive RunnerConfig options
runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
    Agents:      agents,                                    // Map of agent handlers
    QueueSize:   100,                                      // Event queue size
    TraceLogger: agentflow.NewInMemoryTraceLogger(),       // In-memory tracing
    // Or use file-based tracing for production:
    // TraceLogger: agentflow.NewFileTraceLogger("./traces"),
})
```

### Logging Configuration

```go
// Set log level
agentflow.SetLogLevel(agentflow.DEBUG) // DEBUG, INFO, WARN, ERROR

// Get logger for custom logging
logger := agentflow.Logger()
logger.Info().
    Str("session_id", sessionID).
    Str("agent", "my-agent").
    Msg("Custom log message")
```

### Key Constants and Metadata Keys

AgentFlow provides several important constants for metadata handling:

```go
// Standard metadata keys
const (
    RouteMetadataKey = "route_to"     // Agent routing
    SessionIDKey     = "session_id"   // Session tracking
)

// Usage in event creation
metadata := map[string]string{
    agentflow.RouteMetadataKey: "target-agent",
    agentflow.SessionIDKey:     "session-123",
    "custom_key":               "custom_value",
}

// Access in agents
func (a *MyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    // Get routing information
    targetAgent, exists := event.GetMetadataValue(agentflow.RouteMetadataKey)
    if exists {
        fmt.Printf("Event routed to: %s\n", targetAgent)
    }
    
    // Get session ID for tracing
    sessionID, exists := event.GetMetadataValue(agentflow.SessionIDKey)
    if exists {
        fmt.Printf("Session ID: %s\n", sessionID)
    }
    
    // Set routing for next agent
    outputState := state.Clone()
    outputState.SetMeta(agentflow.RouteMetadataKey, "next-agent")
    
    return agentflow.AgentResult{OutputState: outputState}, nil
}
```

### Trace Logger Options

```go
// In-memory trace logger (for development/testing)
memoryLogger := agentflow.NewInMemoryTraceLogger()

// File-based trace logger (for production)
fileLogger := agentflow.NewFileTraceLogger("./application-traces")

// Using with runner
runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
    Agents:      agents,
    TraceLogger: fileLogger, // or memoryLogger
})

// Retrieve traces
traces, err := runner.DumpTrace(sessionID)
if err != nil {
    log.Printf("Error retrieving traces: %v", err)
    return
}

// Analyze trace entries
for _, trace := range traces {
    fmt.Printf("Trace: %s - Agent: %s - Type: %s - Duration: %v\n",
        trace.Timestamp.Format("15:04:05.000"),
        trace.AgentID,
        trace.Type,
        trace.Duration)
}
```

## Best Practices

### 1. Agent Design

- **Single Responsibility**: Keep agents focused on one specific task
- **Use Meaningful Names**: Choose descriptive names for agents and state keys
- **Proper Error Handling**: Implement comprehensive error handling with routing
- **Include Comprehensive Logging**: Use structured logging with context
- **State Cloning**: Always clone state to avoid mutations affecting other agents

```go
func (a *MyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    startTime := time.Now()
    
    // Always log execution start
    agentflow.Logger().Info().
        Str("agent", a.name).
        Str("event_id", event.GetID()).
        Msg("Agent execution started")
    
    // Always clone state
    outputState := state.Clone()
    
    // Your processing logic...
    
    // Always log completion with timing
    agentflow.Logger().Info().
        Str("agent", a.name).
        Dur("duration", time.Since(startTime)).
        Msg("Agent execution completed")
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   startTime,
        EndTime:     time.Now(),
        Duration:    time.Since(startTime),
    }, nil
}
```

### 2. State Management

- **Clone Before Modify**: Use `state.Clone()` to avoid side effects
- **Consistent Key Naming**: Use descriptive, consistent naming conventions
- **Include Metadata**: Use metadata for routing and debugging information
- **Validate Required Fields**: Check for required data before processing

```go
// Good state management patterns
outputState := state.Clone()

// Check for required fields
userID, exists := state.Get("user_id")
if !exists {
    return agentflow.AgentResult{}, fmt.Errorf("missing required field: user_id")
}

// Use descriptive keys
outputState.Set("processing_result", result)
outputState.Set("processed_by_agent", a.name)
outputState.Set("processing_timestamp", time.Now().Format(time.RFC3339))

// Set routing metadata
outputState.SetMeta(agentflow.RouteMetadataKey, "next-agent")
outputState.SetMeta("processing_stage", "validation")
```

### 3. Event Handling

- **Use Unique Session IDs**: Essential for tracing and debugging
- **Include Relevant Metadata**: Add routing and context information
- **Handle Context Cancellation**: Respect context cancellation signals
- **Validate Event Data**: Verify event data before processing

```go
// Create events with proper session handling
sessionID := fmt.Sprintf("workflow-%s-%d", workflowType, time.Now().UnixNano())

event := agentflow.NewEvent("source-system", agentflow.EventData{
    "task_type":    "analysis",
    "user_prompt":  userInput,
    "priority":     "high",
}, map[string]string{
    agentflow.RouteMetadataKey: "analyzer-agent",
    agentflow.SessionIDKey:     sessionID,
    "workflow_type":            workflowType,
    "user_id":                  userID,
})

// In agent, handle context cancellation
func (a *MyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    select {
    case <-ctx.Done():
        return agentflow.AgentResult{}, ctx.Err()
    default:
        // Continue processing
    }
    
    // Validate event data
    taskType, exists := event.GetData()["task_type"]
    if !exists {
        return agentflow.AgentResult{}, fmt.Errorf("missing task_type in event data")
    }
    
    // Your processing logic...
}
```

### 4. Error Handling and Resilience

- **Implement Retry Logic**: For transient failures
- **Use Circuit Breaker Patterns**: Prevent cascade failures
- **Route Errors Appropriately**: Use error routing for recovery
- **Log Errors with Context**: Include sufficient debugging information

```go
type ResilientAgent struct {
    name       string
    maxRetries int
    retryDelay time.Duration
}

func (a *ResilientAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    var lastErr error
    
    for attempt := 0; attempt <= a.maxRetries; attempt++ {
        if attempt > 0 {
            agentflow.Logger().Warn().
                Str("agent", a.name).
                Int("attempt", attempt).
                Err(lastErr).
                Msg("Retrying agent execution")
                
            select {
            case <-ctx.Done():
                return agentflow.AgentResult{}, ctx.Err()
            case <-time.After(a.retryDelay):
            }
        }
        
        result, err := a.processEvent(ctx, event, state)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        if !a.isRetryableError(err) {
            break
        }
    }
    
    // Route to error handler
    outputState := state.Clone()
    outputState.Set("error_message", lastErr.Error())
    outputState.Set("failed_agent", a.name)
    outputState.SetMeta(agentflow.RouteMetadataKey, "error-handler")
    
    return agentflow.AgentResult{OutputState: outputState}, nil
}
```

### 5. Performance and Monitoring

- **Configure Appropriate Queue Sizes**: Balance memory usage and throughput
- **Monitor Trace Logs**: Use traces for bottleneck identification
- **Implement Custom Metrics**: Track business-relevant metrics
- **Use File-Based Tracing in Production**: For persistent trace storage

```go
// Production configuration
runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
    Agents:      agents,
    QueueSize:   1000,  // Larger queue for production
    TraceLogger: agentflow.NewFileTraceLogger("./production-traces"),
})

// Custom metrics collection
registry := runner.CallbackRegistry()
registry.Register(agentflow.HookAfterAgentRun, "metrics-collector", func(ctx context.Context, event agentflow.Event, state agentflow.State) error {
    // Collect custom metrics
    agentName := event.GetTargetAgentID()
    duration := state.Get("execution_duration")
    
    // Send to your metrics system
    metricsCollector.RecordAgentExecution(agentName, duration)
    return nil
})

// Performance analysis
func analyzePerformance(runner *agentflow.Runner, sessionID string) {
    traces, _ := runner.DumpTrace(sessionID)
    
    var totalDuration time.Duration
    agentCounts := make(map[string]int)
    
    for _, trace := range traces {
        if trace.Type == "agent_end" {
            totalDuration += trace.Duration
            agentCounts[trace.AgentID]++
        }
    }
    
    fmt.Printf("Performance Summary:\n")
    fmt.Printf("Total Duration: %v\n", totalDuration)
    for agent, count := range agentCounts {
        fmt.Printf("  %s: %d executions\n", agent, count)
    }
}
```

### 6. Testing

- **Use In-Memory Trace Loggers**: Perfect for unit tests
- **Create Test Agents**: Implement simple agents for testing
- **Test Error Scenarios**: Verify error handling and recovery
- **Validate State Transitions**: Ensure proper state management

```go
func TestAgentWorkflow(t *testing.T) {
    // Create test agents
    agents := map[string]agentflow.AgentHandler{
        "test-agent": &TestAgent{shouldSucceed: true},
        "error-handler": &TestErrorHandler{},
    }
    
    // Use in-memory tracing for tests
    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:      agents,
        QueueSize:   10,
        TraceLogger: agentflow.NewInMemoryTraceLogger(),
    })
    
    ctx := context.Background()
    err := runner.Start(ctx)
    require.NoError(t, err)
    defer runner.Stop()
    
    // Test successful execution
    sessionID := "test-session-123"
    event := agentflow.NewEvent("test", agentflow.EventData{
        "test_data": "example",
    }, map[string]string{
        agentflow.RouteMetadataKey: "test-agent",
        agentflow.SessionIDKey:     sessionID,
    })
    
    err = runner.Emit(event)
    require.NoError(t, err)
    
    // Wait and verify traces
    time.Sleep(100 * time.Millisecond)
    traces, err := runner.DumpTrace(sessionID)
    require.NoError(t, err)
    require.Greater(t, len(traces), 0)
}
```

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
