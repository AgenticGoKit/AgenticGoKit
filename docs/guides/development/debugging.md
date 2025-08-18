# Debugging Agent Interactions

**Understand and troubleshoot multi-agent workflows effectively**

This guide teaches you how to debug complex agent interactions in AgenticGoKit. You'll learn to use tracing, logging, and debugging tools to understand what's happening in your multi-agent systems and resolve issues quickly.

## Quick Start (5 minutes)

### 1. Enable Debug Logging

Update your `agentflow.toml`:

```toml
[logging]
level = "debug"  # Change from "info" to "debug"
format = "json"
file = "debug.log"  # Optional: log to file
```

### 2. Run with Verbose Output

```bash
# Run your agent with debug output
go run . -m "Test message" --verbose

# Or set environment variable
export AGENTFLOW_LOG_LEVEL=debug
go run . -m "Test message"
```

### 3. Use Built-in Tracing

```bash
# Check if tracing is available
agentcli trace --help

# View recent traces (if available)
agentcli trace --recent
```

## Understanding Agent Flow

### Basic Flow Visualization

AgenticGoKit processes events through this flow:

```
Event → Runner → Agent Selection → Agent Execution → Result Processing → Next Agent
```

### Tracing Agent Execution

Enable detailed tracing in your code:

```go
import (
    "context"
    "log"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Enable debug logging
    core.SetLogLevel("debug")
    
    // Create agents with tracing
    agents := map[string]core.AgentHandler{
        "analyzer": NewTracedAgent("analyzer", llmProvider),
        "processor": NewTracedAgent("processor", llmProvider),
    }
    
    // Create runner with debug info
    runner := core.CreateSequentialRunner(agents, []string{"analyzer", "processor"}, 30*time.Second)
    
    // Process event with context
    ctx := context.WithValue(context.Background(), "trace_id", generateTraceID())
    event := core.NewEvent("analyze", map[string]interface{}{
        "input": "Debug this workflow",
    })
    
    _ = runner.Start(ctx)
    defer runner.Stop()
    err := runner.Emit(event)
    if err != nil {
        log.Printf("Error processing event: %v", err)
    }
    
    log.Printf("Results: %+v", results)
}
```

### Creating a Traced Agent

```go
type TracedAgent struct {
    name        string
    llmProvider core.ModelProvider
}

func NewTracedAgent(name string, provider core.ModelProvider) *TracedAgent {
    return &TracedAgent{
        name:        name,
        llmProvider: provider,
    }
}

func (a *TracedAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    traceID := ctx.Value("trace_id")
    log.Printf("[TRACE:%v] Agent %s starting execution", traceID, a.name)
    log.Printf("[TRACE:%v] Event type: %s, data: %+v", traceID, event.Type, event.Data)
    log.Printf("[TRACE:%v] State: %+v", traceID, state.Data)
    
    start := time.Now()
    
    // Your agent logic here
    prompt := fmt.Sprintf("Process this input: %v", event.Data)
    response, err := a.llmProvider.GenerateResponse(ctx, prompt, nil)
    
    duration := time.Since(start)
    
    if err != nil {
        log.Printf("[TRACE:%v] Agent %s failed after %v: %v", traceID, a.name, duration, err)
        return nil, err
    }
    
    result := &core.AgentResult{
        Data: map[string]interface{}{
            "response": response,
            "agent":    a.name,
            "duration": duration.String(),
        },
    }
    
    log.Printf("[TRACE:%v] Agent %s completed in %v", traceID, a.name, duration)
    log.Printf("[TRACE:%v] Result: %+v", traceID, result.Data)
    
    return result, nil
}

func generateTraceID() string {
    return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}
```

## Common Debugging Scenarios

### 1. Agent Not Executing

**Symptoms:**
- No output from specific agents
- Workflow stops unexpectedly
- Silent failures

**Debug Steps:**

```go
// Add execution checks
func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    log.Printf("DEBUG: Agent %s received event type: %s", a.name, event.Type)
    
    // Check if agent should handle this event
    if !a.shouldHandle(event) {
        log.Printf("DEBUG: Agent %s skipping event type: %s", a.name, event.Type)
        return nil, nil  // Return nil to pass to next agent
    }
    
    log.Printf("DEBUG: Agent %s processing event", a.name)
    
    // Your processing logic...
    
    return result, nil
}

func (a *MyAgent) shouldHandle(event core.Event) bool {
    // Add your logic to determine if this agent should handle the event
    return event.Type == "analyze" || event.Type == "process"
}
```

### 2. State Management Issues

**Symptoms:**
- Data not passed between agents
- Unexpected state modifications
- Missing context

**Debug Steps:**

```go
func debugState(ctx context.Context, state *core.State, agentName string, phase string) {
    log.Printf("DEBUG: [%s] Agent %s state %s:", generateTraceID(), agentName, phase)
    log.Printf("  Data keys: %v", getKeys(state.Data))
    log.Printf("  Metadata keys: %v", getKeys(state.Metadata))
    
    // Log specific important fields
    if val, exists := state.Data["important_field"]; exists {
        log.Printf("  important_field: %v", val)
    }
    
    // Check for common issues
    if len(state.Data) == 0 {
        log.Printf("  WARNING: State data is empty")
    }
}

func getKeys(m map[string]interface{}) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

// Use in your agent
func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    debugState(ctx, state, a.name, "before")
    
    // Your processing...
    
    debugState(ctx, state, a.name, "after")
    return result, nil
}
```

### 3. Performance Issues

**Symptoms:**
- Slow agent execution
- High memory usage
- Timeout errors

**Debug Steps:**

```go
type PerformanceTracker struct {
    agentTimes map[string][]time.Duration
    mutex      sync.RWMutex
}

func NewPerformanceTracker() *PerformanceTracker {
    return &PerformanceTracker{
        agentTimes: make(map[string][]time.Duration),
    }
}

func (pt *PerformanceTracker) TrackAgent(agentName string, duration time.Duration) {
    pt.mutex.Lock()
    defer pt.mutex.Unlock()
    
    pt.agentTimes[agentName] = append(pt.agentTimes[agentName], duration)
}

func (pt *PerformanceTracker) GetStats(agentName string) (avg, min, max time.Duration, count int) {
    pt.mutex.RLock()
    defer pt.mutex.RUnlock()
    
    times := pt.agentTimes[agentName]
    if len(times) == 0 {
        return 0, 0, 0, 0
    }
    
    var total time.Duration
    min = times[0]
    max = times[0]
    
    for _, t := range times {
        total += t
        if t < min {
            min = t
        }
        if t > max {
            max = t
        }
    }
    
    avg = total / time.Duration(len(times))
    count = len(times)
    return
}

func (pt *PerformanceTracker) PrintStats() {
    pt.mutex.RLock()
    defer pt.mutex.RUnlock()
    
    fmt.Println("Agent Performance Stats:")
    for agent, times := range pt.agentTimes {
        if len(times) > 0 {
            avg, min, max, count := pt.GetStats(agent)
            fmt.Printf("  %s: avg=%v, min=%v, max=%v, count=%d\n", agent, avg, min, max, count)
        }
    }
}
```

## Production Debugging

### Structured Logging

```go
import (
    "github.com/sirupsen/logrus"
)

func setupProductionLogging() *logrus.Logger {
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetLevel(logrus.InfoLevel)
    
    // Add common fields
    logger = logger.WithFields(logrus.Fields{
        "service": "agenticgokit",
        "version": "1.0.0",
    }).Logger
    
    return logger
}

// Use in agents
func (a *MyAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    logger := setupProductionLogging()
    
    logger.WithFields(logrus.Fields{
        "agent":      a.name,
        "event_type": event.Type,
        "trace_id":   ctx.Value("trace_id"),
    }).Info("Agent execution started")
    
    // Your processing...
    
    logger.WithFields(logrus.Fields{
        "agent":    a.name,
        "duration": time.Since(start).String(),
        "success":  err == nil,
    }).Info("Agent execution completed")
    
    return result, err
}
```

### Health Checks

```go
func setupHealthChecks(agents map[string]core.AgentHandler) {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        health := map[string]interface{}{
            "status":    "healthy",
            "timestamp": time.Now(),
            "agents":    make(map[string]string),
        }
        
        // Check each agent's health
        for name, agent := range agents {
            if healthChecker, ok := agent.(HealthChecker); ok {
                if healthChecker.IsHealthy() {
                    health["agents"].(map[string]string)[name] = "healthy"
                } else {
                    health["agents"].(map[string]string)[name] = "unhealthy"
                    health["status"] = "degraded"
                }
            } else {
                health["agents"].(map[string]string)[name] = "unknown"
            }
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(health)
    })
    
    log.Println("Health check endpoint available at /health")
    go http.ListenAndServe(":8081", nil)
}

type HealthChecker interface {
    IsHealthy() bool
}
```

## Troubleshooting Checklist

### When Agents Don't Respond

- [ ] Check if the agent is registered with the runner
- [ ] Verify the event type matches what the agent expects
- [ ] Ensure the agent's `Execute` method is being called
- [ ] Check for panics or unhandled errors
- [ ] Verify LLM provider credentials and connectivity

### When Performance is Poor

- [ ] Measure individual agent execution times
- [ ] Check for memory leaks or excessive allocations
- [ ] Monitor goroutine count for potential leaks
- [ ] Verify database connection pooling (if using memory providers)
- [ ] Check LLM provider response times

### When State is Lost

- [ ] Verify state is being passed correctly between agents
- [ ] Check for state mutations that might cause issues
- [ ] Ensure proper error handling doesn't lose state
- [ ] Verify orchestration mode is appropriate for your use case

## Tools and Utilities

### Debug Command Line Tool

```bash
# Create a simple debug script
cat > debug.sh << 'EOF'
#!/bin/bash
echo "Starting AgenticGoKit Debug Session"
echo "=================================="

# Set debug environment
export AGENTFLOW_LOG_LEVEL=debug
export DEBUG_PAUSE=false

# Run with debug output
go run . -m "$1" 2>&1 | tee debug_output.log

echo "Debug output saved to debug_output.log"
EOF

chmod +x debug.sh

# Usage
./debug.sh "Debug this interaction"
```

### Log Analysis

```bash
# Analyze debug logs
grep "ERROR" debug_output.log
grep "Agent.*failed" debug_output.log
grep "duration" debug_output.log | sort -k3 -n
```

## Next Steps

- **[Testing Agents](testing-agents.md)** - Comprehensive testing strategies
- **[Best Practices](best-practices.md)** - Development best practices
- **[Production Deployment](../deployment/README.md)** - Production deployment