# Multi-Agent Orchestration API Documentation

This document describes the newly exposed multi-agent orchestration capabilities in AgentFlow's public core API.

## Overview

The multi-agent orchestration API provides patterns for coordinating multiple agents in complex workflows. It supports several orchestration modes including collaborative (parallel) execution, sequential pipelines, and looping patterns.

## Key Components

### 1. Orchestration Modes

```go
type OrchestrationMode string

const (
    OrchestrationRoute       OrchestrationMode = "route"       // Route to single agent
    OrchestrationCollaborate OrchestrationMode = "collaborate" // Send to all agents
)
```

### 2. Orchestrator Interface

```go
type Orchestrator interface {
    Dispatch(ctx context.Context, event Event) (AgentResult, error)
    RegisterAgent(name string, handler AgentHandler) error
    GetCallbackRegistry() *CallbackRegistry
    Stop()
}
```

### 3. Multi-Agent Composition (Planned)

The multi-agent composition API will provide:
- Parallel execution of multiple agents
- Sequential agent pipelines 
- Loop-based agent execution
- Configurable failure handling strategies

## Usage Examples

### Collaborative Orchestration

```go
// Create collaborative orchestrator
registry := core.NewCallbackRegistry()
orchestrator := core.NewCollaborativeOrchestrator(registry)

// Register agents
for name, handler := range agents {
    orchestrator.RegisterAgent(name, handler)
}

// Create and dispatch event
eventData := make(core.EventData)
eventData["task"] = "collaborative_analysis"

event := core.NewEvent("any", eventData, nil)
result, err := orchestrator.Dispatch(ctx, event)
```

### Fault-Tolerant Runner

```go
// Create fault-tolerant runner with retry policies
runner := core.CreateFaultTolerantRunner(agents)

// The runner provides built-in retry logic and failure tolerance
// when processing events through the orchestrator
```

### Orchestration Builder Pattern

```go
// Build custom orchestration with specific configuration
runner := core.NewOrchestrationBuilder(core.OrchestrationCollaborate).
    WithAgents(agents).
    WithTimeout(30 * time.Second).
    WithFailureThreshold(0.8).
    WithRetryPolicy(retryPolicy).
    Build()
```

## Configuration Options

### OrchestrationConfig

```go
type OrchestrationConfig struct {
    Timeout          time.Duration  // Overall orchestration timeout
    MaxConcurrency   int           // Maximum concurrent agents
    FailureThreshold float64       // Failure threshold (0.0-1.0)
    RetryPolicy      *RetryPolicy  // Retry configuration
}
```

### RetryPolicy

```go
type RetryPolicy struct {
    MaxRetries      int           // Maximum retry attempts
    InitialDelay    time.Duration // Initial delay before first retry
    MaxDelay        time.Duration // Maximum delay between retries  
    BackoffFactor   float64       // Exponential backoff multiplier
    Jitter          bool          // Add random jitter to delays
    RetryableErrors []string      // List of retryable error codes
}
```

## Convenience Functions

### CreateCollaborativeRunner
Creates a runner where all agents process events in parallel.

```go
func CreateCollaborativeRunner(agents map[string]AgentHandler, timeout time.Duration) Runner
```

### CreateFaultTolerantRunner
Creates a collaborative runner with aggressive retry policies for environments with transient failures.

```go
func CreateFaultTolerantRunner(agents map[string]AgentHandler) Runner
```

### CreateLoadBalancedRunner
Creates a runner that distributes load across multiple agent instances.

```go
func CreateLoadBalancedRunner(agents map[string]AgentHandler, maxConcurrency int) Runner
```

## Agent Handler Interface

Agents must implement the `AgentHandler` interface:

```go
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}
```

## Event and State Management

### Creating Events

```go
// Create event with data and metadata
eventData := make(core.EventData)
eventData["task"] = "process_data"

metadata := map[string]string{
    "priority": "high",
    "source":   "user_request",
}

event := core.NewEvent("target_agent", eventData, metadata)
```

### Working with State

```go
// Create and populate state
state := core.NewState()
state.Set("key", "value")
state.SetMeta("metadata_key", "metadata_value")

// Access state data
if value, ok := state.Get("key"); ok {
    // Use value
}
```

## Error Handling

The orchestration system provides comprehensive error handling:

- **Individual Agent Failures**: Captured in `AgentResult.Error`
- **Orchestration Failures**: Returned as error from `Dispatch`
- **Timeout Handling**: Configurable timeouts with context cancellation
- **Retry Logic**: Automatic retries with exponential backoff
- **Failure Thresholds**: Configurable tolerance for partial failures

## Best Practices

1. **Use Collaborative Mode** for independent parallel processing
2. **Set Appropriate Timeouts** to prevent hanging operations
3. **Configure Retry Policies** for resilient operation
4. **Handle Partial Failures** by checking individual agent results
5. **Use State Management** to pass data between agents effectively

## Migration from Internal APIs

If you were previously using internal orchestrator packages:

1. Replace `internal/orchestrator` imports with `core`
2. Use `core.NewCollaborativeOrchestrator()` instead of internal constructors
3. Update agent handlers to use public `core.AgentHandler` interface
4. Use public `core.Event` and `core.State` types

## Future Enhancements

The following features are planned for future releases:

- **Sequential Agent Pipelines**: Chain agents in sequence with state passing
- **Loop-Based Execution**: Repeat agent execution until conditions are met
- **Advanced Routing**: Content-based routing and load balancing
- **Agent Composition**: Build complex agents from simpler components
- **Visual Workflow Designer**: GUI for designing multi-agent workflows

## Performance Considerations

- Collaborative orchestration runs agents concurrently for better performance
- Configure `MaxConcurrency` to limit resource usage
- Use timeouts to prevent resource leaks
- Monitor agent execution times and optimize slow agents
- Consider using `FailureThreshold` to fail fast when many agents are failing

## Troubleshooting

### Common Issues

1. **Agents Not Registered**: Ensure all agents are registered before dispatching events
2. **Timeout Errors**: Increase timeout values or optimize agent performance  
3. **State Conflicts**: Use unique keys when multiple agents modify state
4. **Memory Leaks**: Always call `Stop()` on orchestrators when done

### Debugging

Enable debug logging to see orchestration flow:
```go
core.Logger().Debug().Msg("Orchestration debug message")
```

### Monitoring

Monitor orchestration metrics:
- Agent execution times
- Failure rates
- Concurrent agent counts
- Event processing throughput
