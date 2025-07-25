# Debugging Multi-Agent Systems in AgenticGoKit

## Overview

Debugging multi-agent systems requires specialized techniques due to their distributed, asynchronous nature. This guide covers practical strategies for identifying, isolating, and resolving issues in complex agent interactions using AgenticGoKit's built-in tracing and debugging capabilities.

AgenticGoKit provides comprehensive tracing through the `agentcli` tool and structured logging through the `zerolog` library, making it easier to understand agent interactions and diagnose issues.

## Prerequisites

- Understanding of [Agent Lifecycle](../core-concepts/agent-lifecycle.md)
- Familiarity with [Orchestration Patterns](../core-concepts/orchestration-patterns.md)
- Basic knowledge of Go debugging tools
- AgenticGoKit project with `agentflow.toml` configuration

## Common Debugging Scenarios

### 1. Agent Not Responding

When an agent stops responding to events, follow this systematic approach:

#### Symptoms
- Events are queued but not processed
- Agent appears "stuck" or unresponsive
- Timeout errors in orchestrator

#### Debugging Steps

```go
// 1. Check agent health status
func checkAgentHealth(ctx context.Context, agent core.Agent) error {
    // Create a simple health check event
    healthEvent := core.NewEvent("health_check", map[string]interface{}{
        "timestamp": time.Now(),
        "check_id":  uuid.New().String(),
    })
    
    // Set a short timeout for health check
    healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Try to process the health check
    result, err := agent.Process(healthCtx, healthEvent, core.NewState())
    if err != nil {
        return fmt.Errorf("agent health check failed: %w", err)
    }
    
    log.Printf("Agent health check passed: %+v", result)
    return nil
}

// 2. Check for deadlocks or blocking operations
func debugAgentExecution(agent core.Agent) {
    // Enable detailed logging
    logger := log.With().
        Str("component", "agent_debug").
        Str("agent_id", agent.ID()).
        Logger()
    
    // Wrap agent with debugging middleware
    debugAgent := &DebuggingAgent{
        agent:  agent,
        logger: logger,
    }
    
    // Use the debug-wrapped agent
    // This will log all method calls and their durations
}

type DebuggingAgent struct {
    agent  core.Agent
    logger zerolog.Logger
}

func (d *DebuggingAgent) Process(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    start := time.Now()
    eventID := event.GetID()
    
    d.logger.Info().
        Str("event_id", eventID).
        Str("event_type", event.GetType()).
        Msg("Agent processing started")
    
    // Check for context cancellation
    select {
    case <-ctx.Done():
        d.logger.Warn().
            Str("event_id", eventID).
            Err(ctx.Err()).
            Msg("Context cancelled before processing")
        return core.AgentResult{}, ctx.Err()
    default:
    }
    
    result, err := d.agent.Process(ctx, event, state)
    
    duration := time.Since(start)
    if err != nil {
        d.logger.Error().
            Str("event_id", eventID).
            Dur("duration", duration).
            Err(err).
            Msg("Agent processing failed")
    } else {
        d.logger.Info().
            Str("event_id", eventID).
            Dur("duration", duration).
            Msg("Agent processing completed")
    }
    
    return result, err
}
```

#### Common Causes and Solutions

1. **Blocking I/O Operations**
   ```go
   // Problem: Blocking without timeout
   resp, err := http.Get("https://api.example.com/data")
   
   // Solution: Use context with timeout
   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   defer cancel()
   
   req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
   resp, err := http.DefaultClient.Do(req)
   ```

2. **Resource Exhaustion**
   ```go
   // Monitor resource usage
   func monitorAgentResources(agent core.Agent) {
       ticker := time.NewTicker(30 * time.Second)
       defer ticker.Stop()
       
       for range ticker.C {
           var m runtime.MemStats
           runtime.ReadMemStats(&m)
           
           log.Printf("Agent %s - Memory: %d KB, Goroutines: %d", 
               agent.ID(), 
               m.Alloc/1024, 
               runtime.NumGoroutine())
       }
   }
   ```

3. **Deadlocks**
   ```go
   // Use Go's race detector and deadlock detector
   // go run -race -tags deadlock main.go
   
   // Add timeout to all blocking operations
   func safeChannelSend(ch chan<- interface{}, data interface{}, timeout time.Duration) error {
       select {
       case ch <- data:
           return nil
       case <-time.After(timeout):
           return fmt.Errorf("channel send timeout after %v", timeout)
       }
   }
   ```

### 2. Incorrect Agent Interactions

When agents don't interact as expected, use these debugging techniques:

#### Event Flow Tracing

```go
// Create a request tracer to follow events through the system
type RequestTracer struct {
    requestID string
    events    []TraceEvent
    mu        sync.Mutex
}

type TraceEvent struct {
    Timestamp time.Time
    AgentID   string
    EventType string
    Action    string // "received", "processing", "completed", "failed"
    Data      map[string]interface{}
}

func NewRequestTracer(requestID string) *RequestTracer {
    return &RequestTracer{
        requestID: requestID,
        events:    make([]TraceEvent, 0),
    }
}

func (rt *RequestTracer) TraceEvent(agentID, eventType, action string, data map[string]interface{}) {
    rt.mu.Lock()
    defer rt.mu.Unlock()
    
    rt.events = append(rt.events, TraceEvent{
        Timestamp: time.Now(),
        AgentID:   agentID,
        EventType: eventType,
        Action:    action,
        Data:      data,
    })
}

func (rt *RequestTracer) PrintTrace() {
    rt.mu.Lock()
    defer rt.mu.Unlock()
    
    fmt.Printf("\\n=== Request Trace: %s ===\\n", rt.requestID)
    for _, event := range rt.events {
        fmt.Printf("[%s] %s:%s - %s (%+v)\\n",
            event.Timestamp.Format("15:04:05.000"),
            event.AgentID,
            event.EventType,
            event.Action,
            event.Data)
    }
    fmt.Printf("=== End Trace ===\\n\\n")
}

// Use the tracer in your orchestrator
func debugOrchestration(orchestrator core.Orchestrator, event core.Event) {
    requestID := event.GetID()
    tracer := NewRequestTracer(requestID)
    
    // Wrap agents with tracing
    wrappedAgents := make([]core.Agent, 0)
    for _, agent := range orchestrator.GetAgents() {
        wrappedAgent := &TracingAgent{
            agent:  agent,
            tracer: tracer,
        }
        wrappedAgents = append(wrappedAgents, wrappedAgent)
    }
    
    // Create new orchestrator with wrapped agents
    debugOrchestrator := core.NewOrchestrator(wrappedAgents...)
    
    // Process the event
    result, err := debugOrchestrator.Process(context.Background(), event, core.NewState())
    
    // Print the trace
    tracer.PrintTrace()
    
    if err != nil {
        log.Printf("Orchestration failed: %v", err)
    } else {
        log.Printf("Orchestration completed: %+v", result)
    }
}

type TracingAgent struct {
    agent  core.Agent
    tracer *RequestTracer
}

func (ta *TracingAgent) Process(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    agentID := ta.agent.ID()
    eventType := event.GetType()
    
    ta.tracer.TraceEvent(agentID, eventType, "received", map[string]interface{}{
        "state_keys": state.Keys(),
    })
    
    ta.tracer.TraceEvent(agentID, eventType, "processing", nil)
    
    result, err := ta.agent.Process(ctx, event, state)
    
    if err != nil {
        ta.tracer.TraceEvent(agentID, eventType, "failed", map[string]interface{}{
            "error": err.Error(),
        })
    } else {
        ta.tracer.TraceEvent(agentID, eventType, "completed", map[string]interface{}{
            "result_type": fmt.Sprintf("%T", result.Data),
        })
    }
    
    return result, err
}
```

#### State Flow Analysis

```go
// Track state changes through the system
type StateTracker struct {
    states []StateSnapshot
    mu     sync.Mutex
}

type StateSnapshot struct {
    Timestamp time.Time
    AgentID   string
    Action    string // "input", "output"
    Keys      []string
    Values    map[string]interface{}
}

func NewStateTracker() *StateTracker {
    return &StateTracker{
        states: make([]StateSnapshot, 0),
    }
}

func (st *StateTracker) TrackState(agentID, action string, state core.State) {
    st.mu.Lock()
    defer st.mu.Unlock()
    
    // Create a snapshot of the state
    snapshot := StateSnapshot{
        Timestamp: time.Now(),
        AgentID:   agentID,
        Action:    action,
        Keys:      state.Keys(),
        Values:    make(map[string]interface{}),
    }
    
    // Copy state values (be careful with large objects)
    for _, key := range state.Keys() {
        if value, exists := state.Get(key); exists {
            // Only copy serializable values for debugging
            if isSerializable(value) {
                snapshot.Values[key] = value
            } else {
                snapshot.Values[key] = fmt.Sprintf("<%T>", value)
            }
        }
    }
    
    st.states = append(st.states, snapshot)
}

func (st *StateTracker) PrintStateFlow() {
    st.mu.Lock()
    defer st.mu.Unlock()
    
    fmt.Printf("\\n=== State Flow Analysis ===\\n")
    for i, snapshot := range st.states {
        fmt.Printf("[%d] %s - %s:%s\\n",
            i,
            snapshot.Timestamp.Format("15:04:05.000"),
            snapshot.AgentID,
            snapshot.Action)
        
        fmt.Printf("    Keys: %v\\n", snapshot.Keys)
        for key, value := range snapshot.Values {
            fmt.Printf("    %s: %v\\n", key, value)
        }
        fmt.Printf("\\n")
    }
    fmt.Printf("=== End State Flow ===\\n\\n")
}

func isSerializable(value interface{}) bool {
    switch value.(type) {
    case string, int, int64, float64, bool, nil:
        return true
    case []interface{}, map[string]interface{}:
        return true
    default:
        return false
    }
}
```

### 3. Performance Issues

When agents are slow or consuming too many resources:

#### Performance Profiling

```go
import (
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "time"
)

// Enable pprof endpoint for profiling
func enableProfiling() {
    go func() {
        log.Println("Starting pprof server on :6060")
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}

// Profile agent execution
func profileAgentExecution(agent core.Agent, event core.Event, state core.State) {
    // CPU profiling
    start := time.Now()
    var m1, m2 runtime.MemStats
    
    runtime.ReadMemStats(&m1)
    runtime.GC() // Force GC to get accurate memory measurements
    
    result, err := agent.Process(context.Background(), event, state)
    
    runtime.ReadMemStats(&m2)
    duration := time.Since(start)
    
    // Calculate memory usage
    memUsed := m2.Alloc - m1.Alloc
    
    log.Printf("Agent Performance Profile:")
    log.Printf("  Duration: %v", duration)
    log.Printf("  Memory Used: %d bytes", memUsed)
    log.Printf("  Goroutines: %d", runtime.NumGoroutine())
    log.Printf("  GC Runs: %d", m2.NumGC-m1.NumGC)
    
    if err != nil {
        log.Printf("  Error: %v", err)
    } else {
        log.Printf("  Result: %+v", result)
    }
}

// Benchmark agent performance
func benchmarkAgent(agent core.Agent, iterations int) {
    event := core.NewEvent("benchmark", map[string]interface{}{
        "test": "performance",
    })
    state := core.NewState()
    
    durations := make([]time.Duration, iterations)
    var totalMemory uint64
    
    for i := 0; i < iterations; i++ {
        var m1, m2 runtime.MemStats
        runtime.ReadMemStats(&m1)
        
        start := time.Now()
        _, err := agent.Process(context.Background(), event, state)
        duration := time.Since(start)
        
        runtime.ReadMemStats(&m2)
        
        durations[i] = duration
        totalMemory += (m2.Alloc - m1.Alloc)
        
        if err != nil {
            log.Printf("Iteration %d failed: %v", i, err)
        }
    }
    
    // Calculate statistics
    var totalDuration time.Duration
    minDuration := durations[0]
    maxDuration := durations[0]
    
    for _, d := range durations {
        totalDuration += d
        if d < minDuration {
            minDuration = d
        }
        if d > maxDuration {
            maxDuration = d
        }
    }
    
    avgDuration := totalDuration / time.Duration(iterations)
    avgMemory := totalMemory / uint64(iterations)
    
    log.Printf("Agent Benchmark Results (%d iterations):", iterations)
    log.Printf("  Average Duration: %v", avgDuration)
    log.Printf("  Min Duration: %v", minDuration)
    log.Printf("  Max Duration: %v", maxDuration)
    log.Printf("  Average Memory: %d bytes", avgMemory)
}
```

#### Resource Monitoring

```go
// Monitor agent resource usage over time
type ResourceMonitor struct {
    agentID   string
    metrics   []ResourceMetric
    mu        sync.Mutex
    stopChan  chan struct{}
    interval  time.Duration
}

type ResourceMetric struct {
    Timestamp   time.Time
    CPUPercent  float64
    MemoryBytes uint64
    Goroutines  int
    GCPauses    time.Duration
}

func NewResourceMonitor(agentID string, interval time.Duration) *ResourceMonitor {
    return &ResourceMonitor{
        agentID:  agentID,
        metrics:  make([]ResourceMetric, 0),
        stopChan: make(chan struct{}),
        interval: interval,
    }
}

func (rm *ResourceMonitor) Start() {
    go rm.monitor()
}

func (rm *ResourceMonitor) Stop() {
    close(rm.stopChan)
}

func (rm *ResourceMonitor) monitor() {
    ticker := time.NewTicker(rm.interval)
    defer ticker.Stop()
    
    var lastCPUTime time.Duration
    var lastTimestamp time.Time
    
    for {
        select {
        case <-rm.stopChan:
            return
        case <-ticker.C:
            rm.collectMetrics(&lastCPUTime, &lastTimestamp)
        }
    }
}

func (rm *ResourceMonitor) collectMetrics(lastCPUTime *time.Duration, lastTimestamp *time.Time) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    now := time.Now()
    
    // Calculate CPU usage (simplified)
    var cpuPercent float64
    if !lastTimestamp.IsZero() {
        // This is a simplified CPU calculation
        // In production, use proper CPU monitoring libraries
        timeDiff := now.Sub(*lastTimestamp)
        if timeDiff > 0 {
            cpuPercent = float64(runtime.NumCPU()) * 100.0 / float64(timeDiff.Nanoseconds()) * 1000000
        }
    }
    
    metric := ResourceMetric{
        Timestamp:   now,
        CPUPercent:  cpuPercent,
        MemoryBytes: m.Alloc,
        Goroutines:  runtime.NumGoroutine(),
        GCPauses:    time.Duration(m.PauseTotalNs),
    }
    
    rm.mu.Lock()
    rm.metrics = append(rm.metrics, metric)
    
    // Keep only last 1000 metrics to prevent memory growth
    if len(rm.metrics) > 1000 {
        rm.metrics = rm.metrics[len(rm.metrics)-1000:]
    }
    rm.mu.Unlock()
    
    *lastTimestamp = now
}

func (rm *ResourceMonitor) GetMetrics() []ResourceMetric {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    // Return a copy to prevent race conditions
    metrics := make([]ResourceMetric, len(rm.metrics))
    copy(metrics, rm.metrics)
    return metrics
}

func (rm *ResourceMonitor) PrintSummary() {
    metrics := rm.GetMetrics()
    if len(metrics) == 0 {
        log.Printf("No metrics collected for agent %s", rm.agentID)
        return
    }
    
    var totalCPU, totalMemory float64
    var maxMemory uint64
    var maxGoroutines int
    
    for _, metric := range metrics {
        totalCPU += metric.CPUPercent
        totalMemory += float64(metric.MemoryBytes)
        
        if metric.MemoryBytes > maxMemory {
            maxMemory = metric.MemoryBytes
        }
        if metric.Goroutines > maxGoroutines {
            maxGoroutines = metric.Goroutines
        }
    }
    
    count := float64(len(metrics))
    avgCPU := totalCPU / count
    avgMemory := totalMemory / count
    
    log.Printf("Resource Summary for Agent %s:", rm.agentID)
    log.Printf("  Samples: %d", len(metrics))
    log.Printf("  Average CPU: %.2f%%", avgCPU)
    log.Printf("  Average Memory: %.2f MB", avgMemory/1024/1024)
    log.Printf("  Peak Memory: %.2f MB", float64(maxMemory)/1024/1024)
    log.Printf("  Max Goroutines: %d", maxGoroutines)
}
```

## Debugging Tools and Utilities

### 1. Agent Inspector

```go
// Comprehensive agent inspection utility
type AgentInspector struct {
    agent   core.Agent
    history []InspectionRecord
    mu      sync.Mutex
}

type InspectionRecord struct {
    Timestamp time.Time
    Event     core.Event
    State     core.State
    Result    core.AgentResult
    Error     error
    Duration  time.Duration
}

func NewAgentInspector(agent core.Agent) *AgentInspector {
    return &AgentInspector{
        agent:   agent,
        history: make([]InspectionRecord, 0),
    }
}

func (ai *AgentInspector) Process(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    start := time.Now()
    
    result, err := ai.agent.Process(ctx, event, state)
    
    duration := time.Since(start)
    
    record := InspectionRecord{
        Timestamp: start,
        Event:     event,
        State:     state.Clone(), // Clone to preserve state at this point
        Result:    result,
        Error:     err,
        Duration:  duration,
    }
    
    ai.mu.Lock()
    ai.history = append(ai.history, record)
    
    // Keep only last 100 records
    if len(ai.history) > 100 {
        ai.history = ai.history[len(ai.history)-100:]
    }
    ai.mu.Unlock()
    
    return result, err
}

func (ai *AgentInspector) GetHistory() []InspectionRecord {
    ai.mu.Lock()
    defer ai.mu.Unlock()
    
    history := make([]InspectionRecord, len(ai.history))
    copy(history, ai.history)
    return history
}

func (ai *AgentInspector) PrintHistory() {
    history := ai.GetHistory()
    
    fmt.Printf("\\n=== Agent Inspection History ===\\n")
    for i, record := range history {
        status := "SUCCESS"
        if record.Error != nil {
            status = "ERROR"
        }
        
        fmt.Printf("[%d] %s - %s (%v) - %s\\n",
            i,
            record.Timestamp.Format("15:04:05.000"),
            record.Event.GetType(),
            record.Duration,
            status)
        
        if record.Error != nil {
            fmt.Printf("    Error: %v\\n", record.Error)
        }
    }
    fmt.Printf("=== End History ===\\n\\n")
}

func (ai *AgentInspector) GetStats() map[string]interface{} {
    history := ai.GetHistory()
    
    if len(history) == 0 {
        return map[string]interface{}{
            "total_requests": 0,
        }
    }
    
    var totalDuration time.Duration
    var errorCount int
    eventTypes := make(map[string]int)
    
    for _, record := range history {
        totalDuration += record.Duration
        if record.Error != nil {
            errorCount++
        }
        eventTypes[record.Event.GetType()]++
    }
    
    return map[string]interface{}{
        "total_requests":    len(history),
        "error_count":       errorCount,
        "error_rate":        float64(errorCount) / float64(len(history)),
        "average_duration":  totalDuration / time.Duration(len(history)),
        "event_types":       eventTypes,
    }
}
```

### 2. System Health Checker

```go
// Comprehensive system health checking
type HealthChecker struct {
    agents       []core.Agent
    orchestrator core.Orchestrator
    checks       []HealthCheck
}

type HealthCheck struct {
    Name        string
    Description string
    CheckFunc   func(context.Context) error
    Timeout     time.Duration
    Critical    bool
}

type HealthResult struct {
    CheckName   string
    Status      string // "PASS", "FAIL", "TIMEOUT"
    Error       error
    Duration    time.Duration
    Timestamp   time.Time
}

func NewHealthChecker(orchestrator core.Orchestrator, agents ...core.Agent) *HealthChecker {
    hc := &HealthChecker{
        agents:       agents,
        orchestrator: orchestrator,
        checks:       make([]HealthCheck, 0),
    }
    
    // Add default health checks
    hc.AddDefaultChecks()
    
    return hc
}

func (hc *HealthChecker) AddDefaultChecks() {
    // Agent responsiveness check
    hc.AddCheck(HealthCheck{
        Name:        "agent_responsiveness",
        Description: "Check if all agents respond to health check events",
        CheckFunc:   hc.checkAgentResponsiveness,
        Timeout:     10 * time.Second,
        Critical:    true,
    })
    
    // Memory usage check
    hc.AddCheck(HealthCheck{
        Name:        "memory_usage",
        Description: "Check system memory usage",
        CheckFunc:   hc.checkMemoryUsage,
        Timeout:     5 * time.Second,
        Critical:    false,
    })
    
    // Goroutine leak check
    hc.AddCheck(HealthCheck{
        Name:        "goroutine_count",
        Description: "Check for goroutine leaks",
        CheckFunc:   hc.checkGoroutineCount,
        Timeout:     5 * time.Second,
        Critical:    false,
    })
}

func (hc *HealthChecker) AddCheck(check HealthCheck) {
    hc.checks = append(hc.checks, check)
}

func (hc *HealthChecker) RunHealthChecks(ctx context.Context) []HealthResult {
    results := make([]HealthResult, 0, len(hc.checks))
    
    for _, check := range hc.checks {
        result := hc.runSingleCheck(ctx, check)
        results = append(results, result)
    }
    
    return results
}

func (hc *HealthChecker) runSingleCheck(ctx context.Context, check HealthCheck) HealthResult {
    start := time.Now()
    
    // Create timeout context
    checkCtx, cancel := context.WithTimeout(ctx, check.Timeout)
    defer cancel()
    
    // Run the check
    err := check.CheckFunc(checkCtx)
    
    duration := time.Since(start)
    status := "PASS"
    
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            status = "TIMEOUT"
        } else {
            status = "FAIL"
        }
    }
    
    return HealthResult{
        CheckName: check.Name,
        Status:    status,
        Error:     err,
        Duration:  duration,
        Timestamp: start,
    }
}

func (hc *HealthChecker) checkAgentResponsiveness(ctx context.Context) error {
    healthEvent := core.NewEvent("health_check", map[string]interface{}{
        "timestamp": time.Now(),
        "check_id":  uuid.New().String(),
    })
    
    for _, agent := range hc.agents {
        _, err := agent.Process(ctx, healthEvent, core.NewState())
        if err != nil {
            return fmt.Errorf("agent %s failed health check: %w", agent.ID(), err)
        }
    }
    
    return nil
}

func (hc *HealthChecker) checkMemoryUsage(ctx context.Context) error {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    // Check if memory usage is above threshold (e.g., 1GB)
    const maxMemoryBytes = 1024 * 1024 * 1024
    if m.Alloc > maxMemoryBytes {
        return fmt.Errorf("memory usage too high: %d bytes (max: %d)", m.Alloc, maxMemoryBytes)
    }
    
    return nil
}

func (hc *HealthChecker) checkGoroutineCount(ctx context.Context) error {
    count := runtime.NumGoroutine()
    
    // Check if goroutine count is above threshold
    const maxGoroutines = 1000
    if count > maxGoroutines {
        return fmt.Errorf("too many goroutines: %d (max: %d)", count, maxGoroutines)
    }
    
    return nil
}

func (hc *HealthChecker) PrintHealthReport(results []HealthResult) {
    fmt.Printf("\\n=== System Health Report ===\\n")
    fmt.Printf("Timestamp: %s\\n\\n", time.Now().Format("2006-01-02 15:04:05"))
    
    passCount := 0
    failCount := 0
    timeoutCount := 0
    
    for _, result := range results {
        status := result.Status
        switch status {
        case "PASS":
            passCount++
        case "FAIL":
            failCount++
        case "TIMEOUT":
            timeoutCount++
        }
        
        fmt.Printf("[%s] %s (%v)\\n", status, result.CheckName, result.Duration)
        if result.Error != nil {
            fmt.Printf("    Error: %v\\n", result.Error)
        }
    }
    
    fmt.Printf("\\nSummary: %d PASS, %d FAIL, %d TIMEOUT\\n", passCount, failCount, timeoutCount)
    
    if failCount > 0 || timeoutCount > 0 {
        fmt.Printf("⚠️  System health issues detected!\\n")
    } else {
        fmt.Printf("✅ All health checks passed\\n")
    }
    
    fmt.Printf("=== End Health Report ===\\n\\n")
}
```

## Best Practices

### 1. Structured Logging
- Use consistent log formats across all agents
- Include correlation IDs to track requests
- Log at appropriate levels (DEBUG, INFO, WARN, ERROR)
- Include context information (agent ID, event type, etc.)

### 2. Error Handling
- Wrap errors with context using `fmt.Errorf`
- Use typed errors for different failure modes
- Implement circuit breakers for external dependencies
- Provide actionable error messages

### 3. Testing Strategies
- Write unit tests for individual agents
- Create integration tests for agent interactions
- Use mocking to isolate components
- Implement chaos testing for resilience

### 4. Monitoring
- Set up metrics for key performance indicators
- Use distributed tracing for complex workflows
- Implement health checks and alerting
- Monitor resource usage trends

## Conclusion

Debugging multi-agent systems requires systematic approaches and specialized tools. By implementing comprehensive logging, tracing, and monitoring, you can effectively identify and resolve issues in complex agent interactions.

In the next tutorial, we'll explore [Logging and Tracing](logging-and-tracing.md) techniques for better observability.

## Next Steps

- [Logging and Tracing](logging-and-tracing.md)
- [Performance Monitoring](performance-monitoring.md)
- [Production Troubleshooting](production-troubleshooting.md)

## Further Reading

- [Go Debugging Tools](https://golang.org/doc/gdb)
- [pprof Profiling](https://golang.org/pkg/net/http/pprof/)
- [Distributed Tracing Best Practices](https://opentracing.io/guides/)
