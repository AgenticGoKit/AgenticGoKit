# Logging and Tracing in AgenticGoKit

## Overview

AgenticGoKit provides comprehensive logging and tracing capabilities through structured logging with `zerolog` and built-in trace collection. This guide covers how to configure logging, use the tracing system, and analyze execution traces to debug multi-agent systems effectively.

## Prerequisites

- Understanding of [Agent Lifecycle](../core-concepts/agent-lifecycle.md)
- Familiarity with [Debugging Multi-Agent Systems](debugging-multi-agent-systems.md)
- AgenticGoKit project with `agentflow.toml` configuration

## Logging System

### Configuration

AgenticGoKit uses structured logging with configurable levels and formats. Configure logging in your `agentflow.toml`:

```toml
[logging]
level = "debug"  # debug, info, warn, error
format = "json"  # json or text

[agent_flow]
name = "my-agent-system"
version = "1.0.0"
```

### Log Levels

AgenticGoKit supports four log levels:

- **DEBUG**: Detailed information for debugging
- **INFO**: General information about system operation
- **WARN**: Warning messages for potential issues
- **ERROR**: Error messages for failures

### Using the Logger

AgenticGoKit provides a global logger that can be used throughout your application:

```go
package main

import (
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Get the global logger
    logger := core.Logger()
    
    // Log at different levels
    logger.Debug().Msg("Debug message")
    logger.Info().Msg("Info message")
    logger.Warn().Msg("Warning message")
    logger.Error().Msg("Error message")
    
    // Log with structured data
    logger.Info().
        Str("agent_id", "research-agent").
        Int("event_count", 5).
        Dur("duration", time.Second*2).
        Msg("Agent processing completed")
}
```

### Agent-Specific Logging

Each agent has its own logger with contextual information:

```go
// In your agent implementation
type MyAgent struct {
    name   string
    logger zerolog.Logger
}

func NewMyAgent(name string) *MyAgent {
    return &MyAgent{
        name:   name,
        logger: core.Logger().With().Str("agent", name).Logger(),
    }
}

func (a *MyAgent) Process(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    a.logger.Info().
        Str("event_id", event.GetID()).
        Str("event_type", event.GetType()).
        Msg("Processing event")
    
    // Your agent logic here
    
    a.logger.Info().
        Str("event_id", event.GetID()).
        Msg("Event processing completed")
    
    return core.AgentResult{}, nil
}
```

## Tracing System

### How Tracing Works

AgenticGoKit automatically collects trace entries during execution. Traces are stored as JSON files with the pattern `<session-id>.trace.json` in your project directory.

### Trace Entry Structure

Each trace entry contains:

```go
type TraceEntry struct {
    Timestamp     time.Time       `json:"timestamp"`
    Type          string          `json:"type"`
    EventID       string          `json:"event_id"`
    SessionID     string          `json:"session_id"`
    AgentID       string          `json:"agent_id"`
    State         *State          `json:"state"`
    AgentResult   *AgentResult    `json:"agent_result,omitempty"`
    Hook          HookPoint       `json:"hook"`
    Error         string          `json:"error,omitempty"`
    TargetAgentID string          `json:"target_agent_id,omitempty"`
    SourceAgentID string          `json:"source_agent_id,omitempty"`
}
```

### Hook Points

AgenticGoKit traces execution at specific hook points:

- **BeforeEventHandling**: Before an event is processed
- **AfterEventHandling**: After an event is processed
- **BeforeAgentRun**: Before an agent processes an event
- **AfterAgentRun**: After an agent processes an event

### Configuring Tracing

Tracing is enabled by default. You can configure it in your runner setup:

```go
package main

import (
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Option 1: Use in-memory trace logger (default)
    traceLogger := core.NewInMemoryTraceLogger()
    
    // Option 2: Use file-based trace logger for persistent traces
    fileTraceLogger, err := core.NewFileTraceLogger("./traces")
    if err != nil {
        log.Fatal("Failed to create file trace logger:", err)
    }
    defer fileTraceLogger.Close() // Important: Close to finalize JSON files
    
    // Create runner from config (plugins set default trace logger)
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil {
        log.Fatal(err)
    }
    // Optionally override trace logger via callbacks
    runner.RegisterCallback(core.HookBeforeEventHandling, "traceDir", func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        return args.State, nil
    })
    
    // Start the runner
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer runner.Stop()
    
    // Dump traces programmatically if needed
    traces, err := runner.DumpTrace("my-session")
    if err != nil {
        log.Printf("Failed to dump traces: %v", err)
    } else {
        log.Printf("Retrieved %d trace entries", len(traces))
    }
}
```

## Using agentcli for Trace Analysis

### Listing Available Traces

```bash
# List all available trace sessions
agentcli list
```

### Viewing Traces

```bash
# View complete trace for a session
agentcli trace <session-id>

# View only agent flow without state details
agentcli trace --flow-only <session-id>

# Filter trace to specific agent
agentcli trace --filter agent=<agent-name> <session-id>

# View verbose trace with full state details
agentcli trace --verbose <session-id>

# Debug trace structure
agentcli trace --debug <session-id>
```

### Example Trace Output

```
Trace for session req-17461596:

┌───────────────────┬─────────────────────────┬────────────────┬────────────────────────────────────────┬──────────────────────────────┐
│ TIMESTAMP         │ HOOK                    │ AGENT          │ STATE                                  │ ERROR                        │
├───────────────────┼─────────────────────────┼────────────────┼────────────────────────────────────────┼──────────────────────────────┤
│ 14:32:15.123      │ BeforeEventHandling     │ planner        │ {message:"Analyze this data", ...}     │ -                            │
│ 14:32:15.145      │ AfterEventHandling      │ planner        │ {analysis:"Data shows trends", ...}    │ -                            │
│ 14:32:15.146      │ BeforeEventHandling     │ summarizer     │ {analysis:"Data shows trends", ...}    │ -                            │
│ 14:32:15.167      │ AfterEventHandling      │ summarizer     │ {summary:"Key findings: ...", ...}     │ -                            │
└───────────────────┴─────────────────────────┴────────────────┴────────────────────────────────────────┴──────────────────────────────┘

Agent request flow for session req-17461596:

TIME             AGENT          NEXT           HOOK                    EVENT ID
14:32:15.123     planner        summarizer     AfterEventHandling      req-1746...
14:32:15.167     summarizer     (end)          AfterEventHandling      req-1746...

Sequence diagram:
----------------
1. planner → summarizer
2. summarizer → (end)

Condensed route:
planner → summarizer
```

## Advanced Logging Patterns

### Correlation IDs

Use correlation IDs to track requests across multiple agents:

```go
func processWithCorrelation(ctx context.Context, event core.Event, state core.State) {
    // Extract or generate correlation ID
    correlationID := event.GetID()
    
    // Add correlation ID to context
    logger := core.Logger().With().
        Str("correlation_id", correlationID).
        Logger()
    
    logger.Info().Msg("Starting request processing")
    
    // Pass correlation ID through state
    state.SetMeta("correlation_id", correlationID)
    
    // Your processing logic here
    
    logger.Info().Msg("Request processing completed")
}
```

### Performance Logging

Log performance metrics for monitoring:

```go
func logPerformanceMetrics(agentID string, duration time.Duration, success bool) {
    logger := core.Logger().With().
        Str("component", "performance").
        Str("agent_id", agentID).
        Logger()
    
    if success {
        logger.Info().
            Dur("duration", duration).
            Msg("Agent execution completed successfully")
    } else {
        logger.Warn().
            Dur("duration", duration).
            Msg("Agent execution failed")
    }
}
```

### Error Context Logging

Provide rich context when logging errors:

```go
func handleAgentError(err error, agentID string, eventID string, state core.State) {
    logger := core.Logger().With().
        Str("agent_id", agentID).
        Str("event_id", eventID).
        Logger()
    
    // Log error with context
    logger.Error().
        Err(err).
        Strs("state_keys", state.Keys()).
        Msg("Agent processing failed")
    
    // Log additional context if available
    if errorCode, exists := state.GetMeta("error_code"); exists {
        logger.Error().
            Str("error_code", errorCode).
            Msg("Error code context")
    }
}
```

## Custom Trace Logging

### Implementing Custom Trace Logger

You can implement a custom trace logger for specific needs:

```go
type FileTraceLogger struct {
    filePath string
    mu       sync.Mutex
}

func NewFileTraceLogger(filePath string) *FileTraceLogger {
    return &FileTraceLogger{
        filePath: filePath,
    }
}

func (f *FileTraceLogger) Log(entry core.TraceEntry) error {
    f.mu.Lock()
    defer f.mu.Unlock()
    
    // Open file for appending
    file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Marshal entry to JSON
    data, err := json.Marshal(entry)
    if err != nil {
        return err
    }
    
    // Write to file
    _, err = file.Write(append(data, '\n'))
    return err
}

func (f *FileTraceLogger) GetTrace(sessionID string) ([]core.TraceEntry, error) {
    // Implementation to read and filter traces by session ID
    // This is a simplified example
    return nil, fmt.Errorf("not implemented")
}
```

### Using Custom Trace Logger

```go
func main() {
    // Create runner with default tracing from plugins
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the runner
    // ...
}
```

## Monitoring and Alerting

### Log-Based Monitoring

Set up monitoring based on log patterns:

```go
// Monitor error rates
func monitorErrorRate() {
    logger := core.Logger().With().
        Str("component", "monitoring").
        Logger()
    
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        // Calculate error rate from logs
        errorRate := calculateErrorRate()
        
        logger.Info().
            Float64("error_rate", errorRate).
            Msg("Error rate metric")
        
        if errorRate > 0.1 { // 10% threshold
            logger.Warn().
                Float64("error_rate", errorRate).
                Msg("High error rate detected")
        }
    }
}
```

### Health Check Logging

Log health check results for monitoring:

```go
func logHealthCheck(component string, healthy bool, details map[string]interface{}) {
    logger := core.Logger().With().
        Str("component", "health_check").
        Str("check_component", component).
        Logger()
    
    if healthy {
        logger.Info().
            Fields(details).
            Msg("Health check passed")
    } else {
        logger.Error().
            Fields(details).
            Msg("Health check failed")
    }
}
```

## Best Practices

### 1. Structured Logging
- Always use structured logging with key-value pairs
- Use consistent field names across your application
- Include relevant context in every log message

### 2. Log Levels
- Use DEBUG for detailed debugging information
- Use INFO for normal operational messages
- Use WARN for potential issues that don't stop execution
- Use ERROR for actual failures

### 3. Performance Considerations
- Avoid logging large objects in production
- Use appropriate log levels to control verbosity
- Consider async logging for high-throughput scenarios

### 4. Security
- Never log sensitive information (passwords, tokens, etc.)
- Sanitize user input before logging
- Use log rotation to manage disk space

### 5. Trace Analysis
- Use correlation IDs to track requests across agents
- Analyze trace patterns to identify bottlenecks
- Monitor trace file sizes and implement rotation

## Troubleshooting Common Issues

### Missing Traces

If traces are not being generated:

1. Check that tracing is enabled in your configuration
2. Verify that the runner has write permissions to the directory
3. Ensure that the callback registry is properly configured

### Large Trace Files

If trace files become too large:

1. Implement trace rotation
2. Filter traces to specific sessions or agents
3. Use sampling for high-volume scenarios

### Performance Impact

If logging impacts performance:

1. Reduce log level in production
2. Use async logging
3. Monitor logging overhead

## Conclusion

Effective logging and tracing are essential for debugging multi-agent systems. AgenticGoKit's built-in capabilities provide comprehensive visibility into system behavior, making it easier to identify and resolve issues.

In the next tutorial, we'll explore [Performance Monitoring](performance-monitoring.md) techniques for optimizing system performance.

## Next Steps

- [Performance Monitoring](performance-monitoring.md)
- [Production Troubleshooting](production-troubleshooting.md)
- [Debugging Multi-Agent Systems](debugging-multi-agent-systems.md)

## Further Reading

- [Zerolog Documentation](https://github.com/rs/zerolog)
- [Structured Logging Best Practices](https://engineering.grab.com/structured-logging)
- [Distributed Tracing Concepts](https://opentracing.io/docs/overview/what-is-tracing/)