---
title: Logging and Tracing
description: Configure logging and analyze execution traces using AgenticGoKit's current API
sidebar: true
outline: deep
editLink: true
lastUpdated: true
prev:
  text: 'Debugging Multi-Agent Systems'
  link: './debugging-multi-agent-systems'
next:
  text: 'Core Concepts'
  link: '../core-concepts/'
tags:
  - logging
  - tracing
  - agentcli
  - monitoring
  - structured-logging
  - callbacks
head:
  - - meta
    - name: keywords
      content: AgenticGoKit, logging, tracing, agentcli, structured logging, monitoring
  - - meta
    - property: og:title
      content: Logging and Tracing - AgenticGoKit
  - - meta
    - property: og:description
      content: Configure logging and analyze execution traces using AgenticGoKit's current API
---

# Logging and Tracing in AgenticGoKit

## Overview

AgenticGoKit provides comprehensive logging and tracing capabilities through structured logging with `zerolog` and built-in trace collection. This guide covers how to configure logging, use the tracing system with the current Runner API, and analyze execution traces using agentcli to debug multi-agent systems effectively.

## Prerequisites

::: tip Required Knowledge
Build on these fundamentals to master logging and tracing techniques.
:::

- Understanding of [Agent Lifecycle](../core-concepts/agent-lifecycle) and the unified Agent interface
- Familiarity with [Debugging Multi-Agent Systems](./debugging-multi-agent-systems) and current debugging techniques
- AgenticGoKit project with `agentflow.toml` configuration and Runner setup

## Logging System

### Configuration

AgenticGoKit uses structured logging with configurable levels and formats. Configure logging in your `agentflow.toml`:

::: code-group

```toml [Development Config]
[logging]
level = "debug"
format = "text"

[agent_flow]
name = "my-agent-system-dev"
version = "1.0.0"

[llm]
provider = "ollama"
model = "gemma2:2b"
temperature = 0.7
max_tokens = 800
```

```toml [Production Config]
[logging]
level = "info"
format = "json"

[agent_flow]
name = "my-agent-system"
version = "1.0.0"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.3
max_tokens = 1000
```

:::

### Log Levels

::: info Log Level Configuration
Configure log levels in `agentflow.toml` under the `[logging]` section.
:::

AgenticGoKit supports four log levels:

::: details Log Level Details
- **DEBUG**: Detailed information for debugging (use sparingly in production)
- **INFO**: General information about system operation  
- **WARN**: Warning messages for potential issues
- **ERROR**: Error messages for failures
:::

### Using the Logger

AgenticGoKit provides a global logger that can be used throughout your application:

```go
import (

package main
    "time"
)
func main() {
    // Get the global logger
    logger := core.Logger()
    // Log at different levels
    logger.Debug().Msg("Debug message")
    logger.Info().Msg("Info message")
    logger.Warn().Msg("Warning message")
    logger.Error().Msg("Error message")
    // Log with structured data for agent operations
    logger.Info().
        Str("agent_name", "research-agent").
        Str("agent_role", "researcher").
        Int("state_keys", 5).
        Dur("duration", time.Second*2).
        Msg("Agent execution completed")
}

```

### Agent-Specific Logging

Each agent should have its own logger with contextual information using the current Agent interface:

```go
package main

import (

    "time"
)
// In your agent implementation
type MyAgent struct {
    name   string
    role   string
    logger zerolog.Logger
}
func NewMyAgent(name, role string) *MyAgent {
    return &MyAgent{
        name:   name,
        role:   role,
        logger: core.Logger().With().
            Str("agent", name).
            Str("role", role).
            Logger(),
    }
}
func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    a.logger.Info().
        Strs("input_keys", state.Keys()).
        Strs("input_meta", state.MetaKeys()).
        Msg("Starting agent execution")
    // Your agent logic here
    result := state.Clone()
    result.Set("processed_by", a.name)
    result.Set("processed_at", time.Now())
    a.logger.Info().
        Strs("output_keys", result.Keys()).
        Msg("Agent execution completed")
    return result, nil
}
// Implement Agent interface methods
func (a *MyAgent) Name() string { return a.name }
func (a *MyAgent) GetRole() string { return a.role }
func (a *MyAgent) GetDescription() string { return fmt.Sprintf("Agent %s with role %s", a.name, a.role) }
func (a *MyAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Default implementation - can be overridden
    return core.AgentResult{}, nil
}
func (a *MyAgent) GetCapabilities() []string { return []string{} }
func (a *MyAgent) GetSystemPrompt() string { return "" }
func (a *MyAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (a *MyAgent) IsEnabled() bool { return true }
func (a *MyAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (a *MyAgent) Initialize(ctx context.Context) error { return nil }
func (a *MyAgent) Shutdown(ctx context.Context) error { return nil }

```

## Tracing System

### How Tracing Works

AgenticGoKit automatically collects trace entries during execution through the Runner's callback system. Traces are stored as JSON files with the pattern `<session-id>.trace.json` in your project directory and can be analyzed using the agentcli tool.

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

::: tip Callback Registration
Register callbacks at these hook points to implement custom tracing and monitoring.
:::

AgenticGoKit traces execution at specific hook points through the callback system:

- **HookBeforeEventHandling**: Before an event is processed by the Runner
- **HookAfterEventHandling**: After an event is processed by the Runner
- **HookBeforeAgentRun**: Before an agent executes (Run method called)
- **HookAfterAgentRun**: After an agent executes (Run method completed)
- **HookAgentError**: When an agent execution fails

### Configuring Tracing

Tracing is enabled by default through the Runner's callback system. You can configure it in your runner setup:

```go
import (

package main
    "log"
)
func main() {
    // Create runner from config (tracing is automatically enabled)
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil {
        log.Fatal(err)
    }
    // Register custom tracing callbacks if needed
    runner.RegisterCallback(core.HookBeforeAgentRun, "custom-tracer",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            core.Logger().Debug().
                Str("agent_id", args.AgentID).
                Strs("state_keys", args.State.Keys()).
                Msg("Agent execution starting")
            return args.State, nil
        })
    runner.RegisterCallback(core.HookAfterAgentRun, "custom-tracer",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            if args.Error != nil {
                core.Logger().Error().
                    Str("agent_id", args.AgentID).
                    Err(args.Error).
                    Msg("Agent execution failed")
            } else {
                core.Logger().Info().
                    Str("agent_id", args.AgentID).
                    Msg("Agent execution completed successfully")
            }
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
        core.Logger().Error().Err(err).Msg("Failed to dump traces")
    } else {
        core.Logger().Info().Int("trace_count", len(traces)).Msg("Retrieved trace entries")
    }
}

```

## Using agentcli for Trace Analysis

### Listing Available Traces

::: code-group

```bash [List Sessions]
# List all available trace sessions
agentcli list
```

```bash [List with Details]
# List sessions with additional details
agentcli list --verbose

# List sessions from specific time range
agentcli list --since "2024-01-01"
```

:::

### Viewing Traces

::: code-group

```bash [Basic Commands]
# View complete trace for a session
agentcli trace <session-id>

# View only agent flow without state details
agentcli trace --flow-only <session-id>

# Filter trace to specific agent
agentcli trace --filter agent=<agent-name> <session-id>
```

```bash [Advanced Analysis]
# View verbose trace with full state details
agentcli trace --verbose <session-id>

# Debug trace structure
agentcli trace --debug <session-id>

# Analyze trace performance
agentcli trace --analyze <session-id>
```

```bash [Export & Monitoring]
# Export trace to different formats
agentcli trace --format json <session-id> > trace.json
agentcli trace --format csv <session-id> > trace.csv

# Real-time trace monitoring
agentcli trace --follow <session-id>
```

:::

### Example Trace Output

```text
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

Use correlation IDs to track requests across multiple agents using the State metadata:

```go
package main

import (

    "time"
)
processWithCorrelation(ctx context.Context, state core.State) (core.State, error) {
    // Extract or generate correlation ID
    correlationID, exists := state.GetMeta("correlation_id")
    if !exists {
        correlationID = uuid.New().String()
        state.SetMeta("correlation_id", correlationID)
    }
    // Add correlation ID to logger context
    logger := core.Logger().With().
        Str("correlation_id", correlationID).
        Logger()
    logger.Info().Msg("Starting request processing")
    // Your processing logic here
    result := state.Clone()
    result.Set("processed", true)
    result.Set("processed_at", time.Now())
    logger.Info().Msg("Request processing completed")
    return result, nil
}

```

### Performance Logging

Log performance metrics for monitoring using the current Agent interface:

```go
package main

import (

    "time"
)
logPerformanceMetrics(agent core.Agent, duration time.Duration, success bool, stateSize int) {
    logger := core.Logger().With().
        Str("component", "performance").
        Str("agent_name", agent.Name()).
        Str("agent_role", agent.GetRole()).
        Logger()
    if success {
        logger.Info().
            Dur("duration", duration).
            Int("state_keys", stateSize).
            Strs("capabilities", agent.GetCapabilities()).
            Msg("Agent execution completed successfully")
    } else {
        logger.Warn().
            Dur("duration", duration).
            Int("state_keys", stateSize).
            Msg("Agent execution failed")
    }
}

```

### Error Context Logging

Provide rich context when logging errors using the current Agent interface:

```go
package main

import (

    "runtime"
    "time"
)
handleAgentError(err error, agent core.Agent, state core.State) {
    logger := core.Logger().With().
        Str("agent_name", agent.Name()).
        Str("agent_role", agent.GetRole()).
        Logger()
    // Log error with context
    logger.Error().
        Err(err).
        Strs("state_keys", state.Keys()).
        Strs("state_meta", state.MetaKeys()).
        Strs("capabilities", agent.GetCapabilities()).
        Msg("Agent execution failed")
    // Log additional context if available
    if errorCode, exists := state.GetMeta("error_code"); exists {
        logger.Error().
            Str("error_code", errorCode).
            Msg("Error code context")
    }
    if sessionID, exists := state.GetMeta("session_id"); exists {
        logger.Error().
            Str("session_id", sessionID).
            Msg("Session context for error")
    }
}
// Enhanced error handling with recovery patterns
executeAgentWithRecovery(ctx context.Context, agent core.Agent, state core.State) (result core.State, err error) {
    defer func() {
        if r := recover(); r != nil {
            // Capture stack trace
            buf := make([]byte, 4096)
            n := runtime.Stack(buf, false)
            core.Logger().Error().
                Str("agent_name", agent.Name()).
                Str("panic", fmt.Sprintf("%v", r)).
                Str("stack_trace", string(buf[:n])).
                Msg("Agent execution panicked")
            err = fmt.Errorf("agent %s panicked: %v", agent.Name(), r)
        }
    }()
    // Execute agent with timeout and error handling
    timeout := agent.GetTimeout()
    if timeout == 0 {
        timeout = 30 * time.Second
    }
    agentCtx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    result, err = agent.Run(agentCtx, state)
    if err != nil {
        handleAgentError(err, agent, state)
        return state, fmt.Errorf("agent execution failed: %w", err)
    }
    return result, nil
}
// Error callback registration for comprehensive error tracking
registerErrorCallbacks(runner *core.Runner) {
    runner.RegisterCallback(core.HookAgentError, "error-tracker",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            if args.Error != nil {
                core.Logger().Error().
                    Str("agent_id", args.AgentID).
                    Err(args.Error).
                    Strs("state_keys", args.State.Keys()).
                    Msg("Agent error captured by callback")
                // Add error metadata to state for downstream processing
                args.State.SetMeta("last_error", args.Error.Error())
                args.State.SetMeta("error_timestamp", time.Now().Format(time.RFC3339))
            }
            return args.State, nil
        })
}

```

## Custom Trace Logging

### Implementing Custom Trace Callbacks

You can implement custom trace callbacks for specific needs using the Runner's callback system:

```go
package main

import (

    "encoding/json"
    "os"
    "sync"
    "time"
)
type CustomTraceLogger struct {
    filePath string
    mu       sync.Mutex
    file     *os.File
}
func &CustomTraceLogger{filePath string) (*CustomTraceLogger, error) {
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return nil, err
    }
    return &CustomTraceLogger{
        filePath: filePath,
        file:     file,
    }, nil
}
func (c *CustomTraceLogger) Close() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.file.Close()
}
func (c *CustomTraceLogger) RegisterCallbacks(runner *core.Runner) {
    // Register callback for agent execution start
    runner.RegisterCallback(core.HookBeforeAgentRun, "custom-trace-logger",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            entry := map[string]interface{}{
                "timestamp":  time.Now(),
                "hook":       "BeforeAgentRun",
                "agent_id":   args.AgentID,
                "state_keys": args.State.Keys(),
                "meta_keys":  args.State.MetaKeys(),
            }
            c.logEntry(entry)
            return args.State, nil
        })
    // Register callback for agent execution completion
    runner.RegisterCallback(core.HookAfterAgentRun, "custom-trace-logger",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            entry := map[string]interface{}{
                "timestamp":  time.Now(),
                "hook":       "AfterAgentRun",
                "agent_id":   args.AgentID,
                "state_keys": args.State.Keys(),
                "success":    args.Error == nil,
            }
            if args.Error != nil {
                entry["error"] = args.Error.Error()
            }
            c.logEntry(entry)
            return args.State, nil
        })
}
func (c *CustomTraceLogger) logEntry(entry map[string]interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    data, err := json.Marshal(entry)
    if err != nil {
        core.Logger().Error().Err(err).Msg("Failed to marshal trace entry")
        return
    }
    _, err = c.file.Write(append(data, '\n'))
    if err != nil {
        core.Logger().Error().Err(err).Msg("Failed to write trace entry")
    }
}

```

### Using Custom Trace Logger

```go
func main() {
    // Create runner with default tracing
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create and register custom trace logger
    customLogger, err := &CustomTraceLogger{"./custom-traces.jsonl")
    if err != nil {
        log.Fatal(err)
    }
    defer customLogger.Close()
    
    customLogger.RegisterCallbacks(runner)
    
    // Start the runner
    ctx := context.Background()
    if err := runner.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer runner.Stop()
    
    // Your application logic here
}
```

## Monitoring and Alerting

### Log-Based Monitoring

Set up monitoring based on log patterns using callback-based metrics:

```go
package main

import (

    "sync"
    "sync/atomic"
    "time"
)
// Monitor error rates using callback system
type ErrorRateMonitor struct {
    totalExecutions int64
    errorCount      int64
}
func &ErrorRateMonitor{) *ErrorRateMonitor {
    return &ErrorRateMonitor{}
}
func (erm *ErrorRateMonitor) RegisterCallbacks(runner *core.Runner) {
    runner.RegisterCallback(core.HookAfterAgentRun, "error-rate-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            atomic.AddInt64(&erm.totalExecutions, 1)
            if args.Error != nil {
                atomic.AddInt64(&erm.errorCount, 1)
            }
            return args.State, nil
        })
}
func (erm *ErrorRateMonitor) StartMonitoring() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        total := atomic.LoadInt64(&erm.totalExecutions)
        errors := atomic.LoadInt64(&erm.errorCount)
        var errorRate float64
        if total > 0 {
            errorRate = float64(errors) / float64(total)
        }
        core.Logger().Info().
            Int64("total_executions", total).
            Int64("error_count", errors).
            Float64("error_rate", errorRate).
            Msg("Error rate metric")
        if errorRate > 0.1 { // 10% threshold
            core.Logger().Warn().
                Float64("error_rate", errorRate).
                Msg("High error rate detected")
        }
    }
}

```

### Health Check Logging

Log health check results for monitoring using the Agent interface:

```go
logHealthCheck(agent core.Agent, healthy bool, duration time.Duration, details map[string]interface{}) {
    logger := core.Logger().With().
        Str("component", "health_check").
        Str("agent_name", agent.Name()).
        Str("agent_role", agent.GetRole()).
        Logger()
    
    if healthy {
        logger.Info().
            Dur("duration", duration).
            Bool("enabled", agent.IsEnabled()).
            Strs("capabilities", agent.GetCapabilities()).
            Fields(details).
            Msg("Agent health check passed")
    } else {
        logger.Error().
            Dur("duration", duration).
            Bool("enabled", agent.IsEnabled()).
            Fields(details).
            Msg("Agent health check failed")
    }
}
```

## Best Practices

::: warning Security
Never log sensitive information like passwords, tokens, or personal data.
:::

### 1. Structured Logging
- Always use structured logging with key-value pairs using `core.Logger()`
- Use consistent field names across your application (agent_name, agent_role, etc.)
- Include relevant context in every log message

### 2. Log Levels
- Use DEBUG for detailed debugging information (disable in production)
- Use INFO for normal operational messages
- Use WARN for potential issues that don't stop execution
- Use ERROR for actual failures

### 3. Performance Considerations
- Avoid logging large State objects in production
- Use appropriate log levels to control verbosity
- Consider callback-based logging for high-throughput scenarios

### 4. Security
- Never log sensitive information (passwords, tokens, etc.)
- Sanitize user input before logging
- Use log rotation to manage disk space

### 5. Trace Analysis
- Use correlation IDs in State metadata to track requests across agents
- Analyze trace patterns with agentcli to identify bottlenecks
- Monitor trace file sizes and implement rotation

## Troubleshooting Common Issues

### Missing Traces

::: details Troubleshooting Missing Traces
If traces are not being generated:

1. **Check Runner Configuration**: Verify tracing is enabled in your Runner setup
2. **File Permissions**: Ensure the runner has write permissions to the directory
3. **Callback Registry**: Confirm the callback registry is properly configured
4. **Agent Execution**: Verify that agents are actually being executed
5. **Session IDs**: Check that session IDs are being properly set
:::

### Large Trace Files

::: details Managing Large Trace Files
If trace files become too large:

1. **Trace Rotation**: Implement trace rotation using custom callbacks
2. **Filtering**: Filter traces to specific sessions or agents using agentcli
3. **Sampling**: Use sampling for high-volume scenarios
4. **Cleanup**: Clean up old trace files regularly
5. **Compression**: Consider compressing archived trace files
:::

### Performance Impact

::: details Optimizing Logging Performance
If logging impacts performance:

1. **Log Levels**: Reduce log level in production (INFO or WARN)
2. **Callback Efficiency**: Use callback-based logging for efficiency
3. **Profiling**: Monitor logging overhead with profiling tools
4. **State Size**: Avoid logging large State objects
5. **Async Logging**: Consider asynchronous logging for high-throughput scenarios
:::

## Quick Reference

::: details Common agentcli Commands
```bash
# Essential commands for daily debugging
agentcli list                           # List all sessions
agentcli trace <session-id>             # View trace
agentcli trace --flow-only <session-id> # Agent flow only
agentcli trace --verbose <session-id>   # Detailed trace
```
:::

::: details Logger Usage Patterns
```go
// Essential logging patterns
logger := core.Logger()
logger.Info().Str("agent", name).Msg("message")
logger.Error().Err(err).Msg("error occurred")
```
:::

## Conclusion

Effective logging and tracing are essential for debugging multi-agent systems. AgenticGoKit's built-in capabilities through the Runner's callback system and structured logging provide comprehensive visibility into system behavior, making it easier to identify and resolve issues using the current API.

## Next Steps

::: info Continue Learning
Explore these related topics to deepen your debugging expertise.
:::

- [Debugging Multi-Agent Systems](./debugging-multi-agent-systems) - Core debugging techniques
- [Agent Lifecycle](../core-concepts/agent-lifecycle) - Understanding agent execution
- [Message Passing](../core-concepts/message-passing) - Event flow and callbacks

## Further Reading

::: info External Resources
Explore these external resources to deepen your understanding of logging and tracing concepts.
:::

- **[Zerolog Documentation](https://github.com/rs/zerolog)** - Official documentation for the structured logging library used by AgenticGoKit
- **[Structured Logging Best Practices](https://engineering.grab.com/structured-logging)** - Industry best practices for implementing structured logging
- **[Distributed Tracing Concepts](https://opentracing.io/docs/overview/what-is-tracing/)** - Understanding distributed tracing principles and patterns##
 Advanced Performance Optimization

### High-Performance Logging Patterns

For high-throughput systems, implement optimized logging patterns that minimize performance impact:

```go
import (
    "context"
    "sync"
    "time"
    "github.com/kunalkushwaha/agenticgokit/core"
)

// High-performance async logger
type AsyncLogger struct {
    logChannel chan LogEntry
    buffer     []LogEntry
    batchSize  int
    flushTimer *time.Timer
    mu         sync.Mutex
    wg         sync.WaitGroup
}

type LogEntry struct {
    Level       string
    Message     string
    Fields      map[string]interface{}
    Timestamp   time.Time
    AgentName   string
    SessionID   string
}

func NewAsyncLogger(batchSize int, flushInterval time.Duration) *AsyncLogger {
    al := &AsyncLogger{
        logChannel: make(chan LogEntry, 1000),
        buffer:     make([]LogEntry, 0, batchSize),
        batchSize:  batchSize,
        flushTimer: time.NewTimer(flushInterval),
    }
    
    al.wg.Add(1)
    go al.processingLoop(flushInterval)
    
    return al
}

func (al *AsyncLogger) LogAsync(level, message, agentName, sessionID string, fields map[string]interface{}) {
    entry := LogEntry{
        Level:     level,
        Message:   message,
        Fields:    fields,
        Timestamp: time.Now(),
        AgentName: agentName,
        SessionID: sessionID,
    }
    
    select {
    case al.logChannel <- entry:
        // Successfully queued
    default:
        // Channel full, log synchronously as fallback
        core.Logger().Warn().
            Str("agent", agentName).
            Str("session", sessionID).
            Msg("Async log channel full, falling back to sync logging")
        
        al.logSync(entry)
    }
}

func (al *AsyncLogger) processingLoop(flushInterval time.Duration) {
    defer al.wg.Done()
    
    for {
        select {
        case entry, ok := <-al.logChannel:
            if !ok {
                // Channel closed, flush remaining entries
                al.flushBuffer()
                return
            }
            
            al.mu.Lock()
            al.buffer = append(al.buffer, entry)
            
            if len(al.buffer) >= al.batchSize {
                al.flushBuffer()
                al.resetTimer(flushInterval)
            }
            al.mu.Unlock()
            
        case <-al.flushTimer.C:
            al.mu.Lock()
            if len(al.buffer) > 0 {
                al.flushBuffer()
            }
            al.resetTimer(flushInterval)
            al.mu.Unlock()
        }
    }
}

func (al *AsyncLogger) flushBuffer() {
    if len(al.buffer) == 0 {
        return
    }
    
    // Batch write to underlying logger
    for _, entry := range al.buffer {
        al.logSync(entry)
    }
    
    // Clear buffer
    al.buffer = al.buffer[:0]
}

func (al *AsyncLogger) logSync(entry LogEntry) {
    logger := core.Logger()
    
    event := logger.Info()
    if entry.Level == "error" {
        event = logger.Error()
    } else if entry.Level == "warn" {
        event = logger.Warn()
    } else if entry.Level == "debug" {
        event = logger.Debug()
    }
    
    event = event.
        Time("timestamp", entry.Timestamp).
        Str("agent", entry.AgentName).
        Str("session", entry.SessionID)
    
    // Add custom fields
    for key, value := range entry.Fields {
        switch v := value.(type) {
        case string:
            event = event.Str(key, v)
        case int:
            event = event.Int(key, v)
        case int64:
            event = event.Int64(key, v)
        case float64:
            event = event.Float64(key, v)
        case bool:
            event = event.Bool(key, v)
        case time.Duration:
            event = event.Dur(key, v)
        default:
            event = event.Interface(key, v)
        }
    }
    
    event.Msg(entry.Message)
}

func (al *AsyncLogger) resetTimer(interval time.Duration) {
    if !al.flushTimer.Stop() {
        select {
        case <-al.flushTimer.C:
        default:
        }
    }
    al.flushTimer.Reset(interval)
}

func (al *AsyncLogger) Close() {
    close(al.logChannel)
    al.wg.Wait()
    
    if !al.flushTimer.Stop() {
        select {
        case <-al.flushTimer.C:
        default:
        }
    }
}
```

### Sampling for High-Volume Tracing

Implement intelligent sampling to reduce tracing overhead in high-volume systems:

```go
import (
    "context"
    "hash/fnv"
    "math/rand"
    "sync/atomic"
    "time"
    "github.com/kunalkushwaha/agenticgokit/core"
)

// Intelligent trace sampler
type TraceSampler struct {
    baseRate        float64  // Base sampling rate (0.0-1.0)
    errorRate       float64  // Always sample errors at this rate
    slowThreshold   time.Duration
    slowRate        float64  // Sample slow requests at this rate
    requestCount    int64
    sampledCount    int64
}

func NewTraceSampler(baseRate, errorRate, slowRate float64, slowThreshold time.Duration) *TraceSampler {
    return &TraceSampler{
        baseRate:      baseRate,
        errorRate:     errorRate,
        slowThreshold: slowThreshold,
        slowRate:      slowRate,
    }
}

func (ts *TraceSampler) ShouldSample(sessionID string, duration time.Duration, hasError bool) bool {
    atomic.AddInt64(&ts.requestCount, 1)
    
    // Always sample errors at higher rate
    if hasError && rand.Float64() < ts.errorRate {
        atomic.AddInt64(&ts.sampledCount, 1)
        return true
    }
    
    // Always sample slow requests
    if duration > ts.slowThreshold && rand.Float64() < ts.slowRate {
        atomic.AddInt64(&ts.sampledCount, 1)
        return true
    }
    
    // Deterministic sampling based on session ID
    if ts.deterministicSample(sessionID) {
        atomic.AddInt64(&ts.sampledCount, 1)
        return true
    }
    
    return false
}

func (ts *TraceSampler) deterministicSample(sessionID string) bool {
    // Use hash of session ID for consistent sampling
    h := fnv.New32a()
    h.Write([]byte(sessionID))
    hash := h.Sum32()
    
    // Convert hash to 0.0-1.0 range
    hashFloat := float64(hash) / float64(^uint32(0))
    
    return hashFloat < ts.baseRate
}

func (ts *TraceSampler) GetStats() (total, sampled int64, rate float64) {
    total = atomic.LoadInt64(&ts.requestCount)
    sampled = atomic.LoadInt64(&ts.sampledCount)
    
    if total > 0 {
        rate = float64(sampled) / float64(total)
    }
    
    return total, sampled, rate
}

// Sampling-aware trace collector
type SamplingTraceCollector struct {
    sampler   *TraceSampler
    collector *CustomTraceLogger
}

func NewSamplingTraceCollector(sampler *TraceSampler, collector *CustomTraceLogger) *SamplingTraceCollector {
    return &SamplingTraceCollector{
        sampler:   sampler,
        collector: collector,
    }
}

func (stc *SamplingTraceCollector) RegisterCallbacks(runner *core.Runner) {
    // Track execution times for sampling decisions
    executionTimes := make(map[string]time.Time)
    var mu sync.Mutex
    
    runner.RegisterCallback(core.HookBeforeAgentRun, "sampling-tracer",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            sessionID, _ := args.State.GetMeta("session_id")
            
            mu.Lock()
            executionTimes[sessionID] = time.Now()
            mu.Unlock()
            
            return args.State, nil
        })
    
    runner.RegisterCallback(core.HookAfterAgentRun, "sampling-tracer",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            sessionID, _ := args.State.GetMeta("session_id")
            
            mu.Lock()
            startTime, exists := executionTimes[sessionID]
            if exists {
                delete(executionTimes, sessionID)
            }
            mu.Unlock()
            
            if !exists {
                return args.State, nil
            }
            
            duration := time.Since(startTime)
            hasError := args.Error != nil
            
            // Make sampling decision
            if stc.sampler.ShouldSample(sessionID, duration, hasError) {
                entry := map[string]interface{}{
                    "timestamp":  time.Now(),
                    "session_id": sessionID,
                    "agent_id":   args.AgentID,
                    "duration":   duration,
                    "success":    !hasError,
                    "sampled":    true,
                }
                
                if hasError {
                    entry["error"] = args.Error.Error()
                }
                
                stc.collector.logEntry(entry)
            }
            
            return args.State, nil
        })
}
```

### Memory-Efficient State Tracking

Implement memory-efficient patterns for tracking large amounts of state data:

```go
import (
    "compress/gzip"
    "encoding/json"
    "io"
    "sync"
    "time"
    "github.com/kunalkushwaha/agenticgokit/core"
)

// Memory-efficient state tracker with compression
type CompressedStateTracker struct {
    states     map[string]*CompressedState
    maxStates  int
    mu         sync.RWMutex
    evictChan  chan string
}

type CompressedState struct {
    SessionID     string
    Timestamp     time.Time
    CompressedData []byte
    OriginalSize  int
    CompressedSize int
}

func NewCompressedStateTracker(maxStates int) *CompressedStateTracker {
    cst := &CompressedStateTracker{
        states:    make(map[string]*CompressedState),
        maxStates: maxStates,
        evictChan: make(chan string, 100),
    }
    
    // Start background eviction goroutine
    go cst.evictionLoop()
    
    return cst
}

func (cst *CompressedStateTracker) TrackState(sessionID string, state core.State) error {
    // Serialize state to JSON
    stateData := make(map[string]interface{})
    
    // Copy state keys and values
    for _, key := range state.Keys() {
        if value, exists := state.Get(key); exists {
            stateData[key] = value
        }
    }
    
    // Add metadata
    metadata := make(map[string]string)
    for _, key := range state.MetaKeys() {
        if value, exists := state.GetMeta(key); exists {
            metadata[key] = value
        }
    }
    stateData["_metadata"] = metadata
    
    // Serialize to JSON
    jsonData, err := json.Marshal(stateData)
    if err != nil {
        return fmt.Errorf("failed to serialize state: %w", err)
    }
    
    // Compress the JSON data
    compressedData, err := cst.compressData(jsonData)
    if err != nil {
        return fmt.Errorf("failed to compress state: %w", err)
    }
    
    compressedState := &CompressedState{
        SessionID:      sessionID,
        Timestamp:      time.Now(),
        CompressedData: compressedData,
        OriginalSize:   len(jsonData),
        CompressedSize: len(compressedData),
    }
    
    cst.mu.Lock()
    defer cst.mu.Unlock()
    
    // Check if we need to evict old states
    if len(cst.states) >= cst.maxStates {
        // Find oldest state to evict
        var oldestSession string
        var oldestTime time.Time
        
        for session, state := range cst.states {
            if oldestSession == "" || state.Timestamp.Before(oldestTime) {
                oldestSession = session
                oldestTime = state.Timestamp
            }
        }
        
        if oldestSession != "" {
            delete(cst.states, oldestSession)
            
            // Notify eviction (non-blocking)
            select {
            case cst.evictChan <- oldestSession:
            default:
            }
        }
    }
    
    cst.states[sessionID] = compressedState
    
    return nil
}

func (cst *CompressedStateTracker) GetState(sessionID string) (core.State, error) {
    cst.mu.RLock()
    compressedState, exists := cst.states[sessionID]
    cst.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("state not found for session %s", sessionID)
    }
    
    // Decompress data
    jsonData, err := cst.decompressData(compressedState.CompressedData)
    if err != nil {
        return nil, fmt.Errorf("failed to decompress state: %w", err)
    }
    
    // Deserialize JSON
    var stateData map[string]interface{}
    if err := json.Unmarshal(jsonData, &stateData); err != nil {
        return nil, fmt.Errorf("failed to deserialize state: %w", err)
    }
    
    // Reconstruct state
    state := core.NewState()
    
    for key, value := range stateData {
        if key == "_metadata" {
            // Handle metadata
            if metadata, ok := value.(map[string]interface{}); ok {
                for metaKey, metaValue := range metadata {
                    if metaStr, ok := metaValue.(string); ok {
                        state.SetMeta(metaKey, metaStr)
                    }
                }
            }
        } else {
            state.Set(key, value)
        }
    }
    
    return state, nil
}

func (cst *CompressedStateTracker) compressData(data []byte) ([]byte, error) {
    var compressed bytes.Buffer
    
    gzWriter := gzip.NewWriter(&compressed)
    defer gzWriter.Close()
    
    if _, err := gzWriter.Write(data); err != nil {
        return nil, err
    }
    
    if err := gzWriter.Close(); err != nil {
        return nil, err
    }
    
    return compressed.Bytes(), nil
}

func (cst *CompressedStateTracker) decompressData(data []byte) ([]byte, error) {
    reader := bytes.NewReader(data)
    gzReader, err := gzip.NewReader(reader)
    if err != nil {
        return nil, err
    }
    defer gzReader.Close()
    
    var decompressed bytes.Buffer
    if _, err := io.Copy(&decompressed, gzReader); err != nil {
        return nil, err
    }
    
    return decompressed.Bytes(), nil
}

func (cst *CompressedStateTracker) evictionLoop() {
    for sessionID := range cst.evictChan {
        core.Logger().Debug().
            Str("session_id", sessionID).
            Msg("State evicted from compressed tracker")
    }
}

func (cst *CompressedStateTracker) GetStats() map[string]interface{} {
    cst.mu.RLock()
    defer cst.mu.RUnlock()
    
    totalOriginalSize := 0
    totalCompressedSize := 0
    
    for _, state := range cst.states {
        totalOriginalSize += state.OriginalSize
        totalCompressedSize += state.CompressedSize
    }
    
    compressionRatio := float64(0)
    if totalOriginalSize > 0 {
        compressionRatio = float64(totalCompressedSize) / float64(totalOriginalSize)
    }
    
    return map[string]interface{}{
        "tracked_states":       len(cst.states),
        "total_original_size":  totalOriginalSize,
        "total_compressed_size": totalCompressedSize,
        "compression_ratio":    compressionRatio,
        "memory_saved_bytes":   totalOriginalSize - totalCompressedSize,
    }
}
```

## Production Monitoring Integration

### Metrics Collection for Production Systems

Integrate with production monitoring systems for comprehensive observability:

```go
import (
    "context"
    "sync"
    "time"
    "github.com/kunalkushwaha/agenticgokit/core"
)

// Production metrics collector
type ProductionMetricsCollector struct {
    metrics map[string]*AgentMetrics
    mu      sync.RWMutex
    
    // Metric aggregation
    totalRequests     int64
    totalErrors       int64
    totalDuration     time.Duration
    
    // Time-based metrics
    requestsPerSecond float64
    errorsPerSecond   float64
    
    // Histogram buckets for response times
    responseBuckets map[string]int64
}

type AgentMetrics struct {
    Name              string
    ExecutionCount    int64
    ErrorCount        int64
    TotalDuration     time.Duration
    MinDuration       time.Duration
    MaxDuration       time.Duration
    LastExecution     time.Time
    LastError         time.Time
    ErrorRate         float64
    AvgDuration       time.Duration
}

func NewProductionMetricsCollector() *ProductionMetricsCollector {
    return &ProductionMetricsCollector{
        metrics: make(map[string]*AgentMetrics),
        responseBuckets: map[string]int64{
            "0-10ms":    0,
            "10-50ms":   0,
            "50-100ms":  0,
            "100-500ms": 0,
            "500ms-1s":  0,
            "1s-5s":     0,
            "5s+":       0,
        },
    }
}

func (pmc *ProductionMetricsCollector) RecordExecution(agentName string, duration time.Duration, err error) {
    pmc.mu.Lock()
    defer pmc.mu.Unlock()
    
    // Get or create agent metrics
    metrics, exists := pmc.metrics[agentName]
    if !exists {
        metrics = &AgentMetrics{
            Name:        agentName,
            MinDuration: duration,
            MaxDuration: duration,
        }
        pmc.metrics[agentName] = metrics
    }
    
    // Update agent-specific metrics
    metrics.ExecutionCount++
    metrics.TotalDuration += duration
    metrics.LastExecution = time.Now()
    
    if duration < metrics.MinDuration {
        metrics.MinDuration = duration
    }
    if duration > metrics.MaxDuration {
        metrics.MaxDuration = duration
    }
    
    if err != nil {
        metrics.ErrorCount++
        metrics.LastError = time.Now()
    }
    
    // Calculate derived metrics
    metrics.ErrorRate = float64(metrics.ErrorCount) / float64(metrics.ExecutionCount)
    metrics.AvgDuration = metrics.TotalDuration / time.Duration(metrics.ExecutionCount)
    
    // Update global metrics
    pmc.totalRequests++
    pmc.totalDuration += duration
    
    if err != nil {
        pmc.totalErrors++
    }
    
    // Update response time buckets
    pmc.updateResponseBuckets(duration)
}

func (pmc *ProductionMetricsCollector) updateResponseBuckets(duration time.Duration) {
    switch {
    case duration < 10*time.Millisecond:
        pmc.responseBuckets["0-10ms"]++
    case duration < 50*time.Millisecond:
        pmc.responseBuckets["10-50ms"]++
    case duration < 100*time.Millisecond:
        pmc.responseBuckets["50-100ms"]++
    case duration < 500*time.Millisecond:
        pmc.responseBuckets["100-500ms"]++
    case duration < 1*time.Second:
        pmc.responseBuckets["500ms-1s"]++
    case duration < 5*time.Second:
        pmc.responseBuckets["1s-5s"]++
    default:
        pmc.responseBuckets["5s+"]++
    }
}

func (pmc *ProductionMetricsCollector) GetMetrics() map[string]interface{} {
    pmc.mu.RLock()
    defer pmc.mu.RUnlock()
    
    globalErrorRate := float64(0)
    globalAvgDuration := time.Duration(0)
    
    if pmc.totalRequests > 0 {
        globalErrorRate = float64(pmc.totalErrors) / float64(pmc.totalRequests)
        globalAvgDuration = pmc.totalDuration / time.Duration(pmc.totalRequests)
    }
    
    // Copy agent metrics
    agentMetrics := make(map[string]AgentMetrics)
    for name, metrics := range pmc.metrics {
        agentMetrics[name] = *metrics
    }
    
    // Copy response buckets
    responseBuckets := make(map[string]int64)
    for bucket, count := range pmc.responseBuckets {
        responseBuckets[bucket] = count
    }
    
    return map[string]interface{}{
        "global": map[string]interface{}{
            "total_requests":    pmc.totalRequests,
            "total_errors":      pmc.totalErrors,
            "error_rate":        globalErrorRate,
            "avg_duration":      globalAvgDuration,
            "requests_per_sec":  pmc.requestsPerSecond,
            "errors_per_sec":    pmc.errorsPerSecond,
        },
        "agents":           agentMetrics,
        "response_buckets": responseBuckets,
    }
}

func (pmc *ProductionMetricsCollector) StartPeriodicReporting(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    lastTime := time.Now()
    lastRequests := int64(0)
    lastErrors := int64(0)
    
    for {
        select {
        case <-ctx.Done():
            return
        case now := <-ticker.C:
            pmc.mu.Lock()
            
            // Calculate rates
            timeDiff := now.Sub(lastTime).Seconds()
            if timeDiff > 0 {
                pmc.requestsPerSecond = float64(pmc.totalRequests-lastRequests) / timeDiff
                pmc.errorsPerSecond = float64(pmc.totalErrors-lastErrors) / timeDiff
            }
            
            lastTime = now
            lastRequests = pmc.totalRequests
            lastErrors = pmc.totalErrors
            
            pmc.mu.Unlock()
            
            // Log metrics
            metrics := pmc.GetMetrics()
            core.Logger().Info().
                Interface("metrics", metrics).
                Msg("Production metrics report")
        }
    }
}

// Integration with Runner callbacks
func (pmc *ProductionMetricsCollector) RegisterCallbacks(runner *core.Runner) {
    executionTimes := make(map[string]time.Time)
    var mu sync.Mutex
    
    runner.RegisterCallback(core.HookBeforeAgentRun, "metrics-collector",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            mu.Lock()
            executionTimes[args.AgentID] = time.Now()
            mu.Unlock()
            
            return args.State, nil
        })
    
    runner.RegisterCallback(core.HookAfterAgentRun, "metrics-collector",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            mu.Lock()
            startTime, exists := executionTimes[args.AgentID]
            if exists {
                delete(executionTimes, args.AgentID)
            }
            mu.Unlock()
            
            if exists {
                duration := time.Since(startTime)
                pmc.RecordExecution(args.AgentID, duration, args.Error)
            }
            
            return args.State, nil
        })
}
```

## Conclusion

Advanced logging and tracing optimization requires careful consideration of performance, memory usage, and operational requirements. By implementing the patterns shown in this guide, you can build high-performance, observable multi-agent systems that provide comprehensive debugging capabilities without sacrificing system performance.

Key takeaways for production systems:

1. **Use async logging** for high-throughput scenarios
2. **Implement intelligent sampling** to reduce tracing overhead
3. **Compress state data** for memory efficiency
4. **Collect comprehensive metrics** for monitoring
5. **Balance observability with performance** based on your requirements

These advanced patterns enable you to maintain full observability of your multi-agent systems while ensuring they can handle production-level loads efficiently.