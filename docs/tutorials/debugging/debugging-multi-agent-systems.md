---
title: Debugging Multi-Agent Systems
description: Practical strategies for debugging complex agent interactions using AgenticGoKit's current API
sidebar: true
outline: deep
editLink: true
lastUpdated: true
prev:
  text: 'Debugging Overview'
  link: './'
next:
  text: 'Logging and Tracing'
  link: './logging-and-tracing'
tags:
  - debugging
  - multi-agent
  - troubleshooting
  - performance
  - monitoring
  - agent-lifecycle
head:
  - - meta
    - name: keywords
      content: AgenticGoKit, debugging, multi-agent systems, agent lifecycle, performance monitoring
  - - meta
    - property: og:title
      content: Debugging Multi-Agent Systems - AgenticGoKit
  - - meta
    - property: og:description
      content: Practical strategies for debugging complex agent interactions using AgenticGoKit's current API
---

# Debugging Multi-Agent Systems in AgenticGoKit

## Overview

Debugging multi-agent systems requires specialized techniques due to their concurrent execution, asynchronous communication, and complex state management. This guide covers practical strategies for identifying, isolating, and resolving issues in complex agent interactions using AgenticGoKit's built-in tracing and debugging capabilities.

AgenticGoKit provides comprehensive tracing through the `agentcli` tool, structured logging through the `zerolog` library, and callback-based monitoring, making it easier to understand agent interactions and diagnose issues using the current unified Agent interface.

## Prerequisites

::: tip Required Knowledge
Ensure you have these fundamentals before proceeding with debugging techniques.
:::

- Understanding of [Agent Lifecycle](../core-concepts/agent-lifecycle) and the unified Agent interface
- Familiarity with [Orchestration Patterns](../core-concepts/orchestration-patterns) and Runner configuration
- Basic knowledge of Go debugging tools and context handling
- AgenticGoKit project with `agentflow.toml` configuration and Runner setup

## Common Debugging Scenarios

### 1. Agent Not Responding

When an agent stops responding to events, follow this systematic approach:

#### Symptoms

::: warning Common Symptoms
- Events are queued but not processed
- Agent appears "stuck" or unresponsive
- Timeout errors in orchestrator
:::

#### Debugging Steps

```go
package main

import (

    "time"
)
// 1. Check agent health status using current Agent interface
checkAgentHealth(ctx context.Context, agent core.Agent) error {
    // Create a simple health check state
    healthState := core.NewState()
    healthState.Set("health_check", true)
    healthState.Set("timestamp", time.Now())
    healthState.SetMeta("check_id", uuid.New().String())
    // Set a short timeout for health check
    healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    // Try to process the health check using Run method
    result, err := agent.Run(healthCtx, healthState)
    if err != nil {
        return fmt.Errorf("agent %s health check failed: %w", agent.Name(), err)
    }
    core.Logger().Info().
        Str("agent", agent.Name()).
        Str("role", agent.GetRole()).
        Msg("Agent health check passed")
    return nil
}
// 2. Check for deadlocks or blocking operations using current Agent interface
debugAgentExecution(agent core.Agent) {
    // Enable detailed logging
    logger := core.Logger().With().
        Str("component", "agent_debug").
        Str("agent_name", agent.Name()).
        Str("agent_role", agent.GetRole()).
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
func (d *DebuggingAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    start := time.Now()
    d.logger.Info().
        Strs("state_keys", state.Keys()).
        Msg("Agent execution started")
    // Check for context cancellation
    select {
    case <-ctx.Done():
        d.logger.Warn().
            Err(ctx.Err()).
            Msg("Context cancelled before processing")
        return state, ctx.Err()
    default:
    }
    result, err := d.agent.Run(ctx, state)
    duration := time.Since(start)
    if err != nil {
        d.logger.Error().
            Dur("duration", duration).
            Err(err).
            Msg("Agent execution failed")
    } else {
        d.logger.Info().
            Dur("duration", duration).
            Strs("result_keys", result.Keys()).
            Msg("Agent execution completed")
    }
    return result, err
}
// Implement other Agent interface methods for debugging wrapper
func (d *DebuggingAgent) Name() string { return d.agent.Name() }
func (d *DebuggingAgent) GetRole() string { return d.agent.GetRole() }
func (d *DebuggingAgent) GetDescription() string { return d.agent.GetDescription() }
func (d *DebuggingAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return d.agent.HandleEvent(ctx, event, state)
}
func (d *DebuggingAgent) GetCapabilities() []string { return d.agent.GetCapabilities() }
func (d *DebuggingAgent) GetSystemPrompt() string { return d.agent.GetSystemPrompt() }
func (d *DebuggingAgent) GetTimeout() time.Duration { return d.agent.GetTimeout() }
func (d *DebuggingAgent) IsEnabled() bool { return d.agent.IsEnabled() }
func (d *DebuggingAgent) GetLLMConfig() *core.ResolvedLLMConfig { return d.agent.GetLLMConfig() }
func (d *DebuggingAgent) Initialize(ctx context.Context) error { return d.agent.Initialize(ctx) }
func (d *DebuggingAgent) Shutdown(ctx context.Context) error { return d.agent.Shutdown(ctx) }

```

#### Common Causes and Solutions

::: tip Best Practice
Always use context with timeouts in agent implementations to prevent blocking operations.
:::

1. **Blocking I/O Operations**

::: code-group

```go [Problem: No Timeout]
// Problem: Blocking without timeout in agent execution
func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Problem: No timeout
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return state, fmt.Errorf("API call failed: %w", err)
    }
    
    result := state.Clone()
    result.Set("api_response", resp.Status)
    return result, nil
}
```

```go [Solution: With Context Timeout]
// Solution: Use context with timeout
func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return state, fmt.Errorf("API call failed: %w", err)
    }
    
    result := state.Clone()
    result.Set("api_response", resp.Status)
    return result, nil
}
```

:::

2. **Resource Exhaustion**
   ```go
   // Monitor resource usage for agents
   monitorAgentResources(agent core.Agent) {
       ticker := time.NewTicker(30 * time.Second)
       defer ticker.Stop()
       
       for range ticker.C {
           var m runtime.MemStats
           runtime.ReadMemStats(&m)
           
           core.Logger().Info().
               Str("agent", agent.Name()).
               Str("role", agent.GetRole()).
               Uint64("memory_kb", m.Alloc/1024).
               Int("goroutines", runtime.NumGoroutine()).
               Msg("Agent resource usage")
       }
   }
   ```

3. **Deadlocks and Context Handling**

::: warning Race Detection
Always use Go's race detector during development: `go run -race main.go`
:::

::: code-group

```bash [Enable Race Detection]
# Use Go's race detector during development
go run -race main.go

# For testing
go test -race ./...

# For building with race detection
go build -race
```

```go [Proper Context Handling]
// Proper context handling in agent execution
func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Check context before starting work
    select {
    case <-ctx.Done():
        return state, ctx.Err()
    default:
    }
    
    // Use context timeout for operations
    timeout := a.GetTimeout()
    workCtx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    // Perform work with context
    result, err := a.performWork(workCtx, state)
    if err != nil {
        return state, fmt.Errorf("work failed: %w", err)
    }
    
    return result, nil
}
```

:::

### 2. Incorrect Agent Interactions

When agents don't interact as expected, use these debugging techniques:

#### Event Flow Tracing

```go
package main

import (

    "strings"
    "sync"
    "time"
)
// Create a request tracer to follow execution through the system
type RequestTracer struct {
    sessionID string
    events    []TraceEvent
    mu        sync.Mutex
}
type TraceEvent struct {
    Timestamp time.Time
    AgentName string
    AgentRole string
    Action    string // "started", "processing", "completed", "failed"
    StateKeys []string
    Metadata  map[string]string
}
func &RequestTracer{sessionID string) *RequestTracer {
    return &RequestTracer{
        sessionID: sessionID,
        events:    make([]TraceEvent, 0),
    }
}
func (rt *RequestTracer) TraceEvent(agentName, agentRole, action string, stateKeys []string, metadata map[string]string) {
    rt.mu.Lock()
    defer rt.mu.Unlock()
    rt.events = append(rt.events, TraceEvent{
        Timestamp: time.Now(),
        AgentName: agentName,
        AgentRole: agentRole,
        Action:    action,
        StateKeys: stateKeys,
        Metadata:  metadata,
    })
}
func (rt *RequestTracer) PrintTrace() {
    rt.mu.Lock()
    defer rt.mu.Unlock()
    fmt.Printf("\\n=== Session Trace: %s ===\\n", rt.sessionID)
    for _, event := range rt.events {
        fmt.Printf("[%s] %s (%s) - %s | Keys: %v\\n",
            event.Timestamp.Format("15:04:05.000"),
            event.AgentName,
            event.AgentRole,
            event.Action,
            event.StateKeys)
    }
    fmt.Printf("=== End Trace ===\\n\\n")
}
// Use the tracer with Runner and callbacks
debugRunnerExecution(runner *core.Runner, sessionID string) {
    tracer := &RequestTracer{sessionID)
    // Register tracing callbacks
    runner.RegisterCallback(core.HookBeforeAgentRun, "tracer",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            tracer.TraceEvent(args.AgentID, "unknown", "started", args.State.Keys(), 
                map[string]string{"session_id": sessionID})
            return args.State, nil
        })
    runner.RegisterCallback(core.HookAfterAgentRun, "tracer",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            action := "completed"
            if args.Error != nil {
                action = "failed"
            }
            tracer.TraceEvent(args.AgentID, "unknown", action, args.State.Keys(),
                map[string]string{"session_id": sessionID})
            return args.State, nil
        })
    // After execution, print the trace
    defer tracer.PrintTrace()
}
type TracingAgent struct {
    agent  core.Agent
    tracer *RequestTracer
}
func (ta *TracingAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    agentName := ta.agent.Name()
    agentRole := ta.agent.GetRole()
    ta.tracer.TraceEvent(agentName, agentRole, "received", state.Keys(), 
        map[string]string{"capabilities": strings.Join(ta.agent.GetCapabilities(), ",")})
    ta.tracer.TraceEvent(agentName, agentRole, "processing", nil, nil)
    result, err := ta.agent.Run(ctx, state)
    if err != nil {
        ta.tracer.TraceEvent(agentName, agentRole, "failed", state.Keys(),
            map[string]string{"error": err.Error()})
    } else {
        ta.tracer.TraceEvent(agentName, agentRole, "completed", result.Keys(),
            map[string]string{"success": "true"})
    }
    return result, err
}
// Implement other Agent interface methods for tracing wrapper
func (ta *TracingAgent) Name() string { return ta.agent.Name() }
func (ta *TracingAgent) GetRole() string { return ta.agent.GetRole() }
func (ta *TracingAgent) GetDescription() string { return ta.agent.GetDescription() }
func (ta *TracingAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return ta.agent.HandleEvent(ctx, event, state)
}
func (ta *TracingAgent) GetCapabilities() []string { return ta.agent.GetCapabilities() }
func (ta *TracingAgent) GetSystemPrompt() string { return ta.agent.GetSystemPrompt() }
func (ta *TracingAgent) GetTimeout() time.Duration { return ta.agent.GetTimeout() }
func (ta *TracingAgent) IsEnabled() bool { return ta.agent.IsEnabled() }
func (ta *TracingAgent) GetLLMConfig() *core.ResolvedLLMConfig { return ta.agent.GetLLMConfig() }
func (ta *TracingAgent) Initialize(ctx context.Context) error { return ta.agent.Initialize(ctx) }
func (ta *TracingAgent) Shutdown(ctx context.Context) error { return ta.agent.Shutdown(ctx) }

```

#### State Flow Analysis

```go
// Track state changes through the system using current State interface
type StateTracker struct {
    states []StateSnapshot
    mu     sync.Mutex
}

type StateSnapshot struct {
    Timestamp time.Time
    AgentName string
    AgentRole string
    Action    string // "input", "output"
    Keys      []string
    MetaKeys  []string
    Values    map[string]interface{}
    Metadata  map[string]string
}

func &StateTracker{) *StateTracker {
    return &StateTracker{
        states: make([]StateSnapshot, 0),
    }
}

func (st *StateTracker) TrackState(agentName, agentRole, action string, state core.State) {
    st.mu.Lock()
    defer st.mu.Unlock()
    
    // Create a snapshot of the state
    snapshot := StateSnapshot{
        Timestamp: time.Now(),
        AgentName: agentName,
        AgentRole: agentRole,
        Action:    action,
        Keys:      state.Keys(),
        MetaKeys:  state.MetaKeys(),
        Values:    make(map[string]interface{}),
        Metadata:  make(map[string]string),
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
    
    // Copy metadata
    for _, key := range state.MetaKeys() {
        if value, exists := state.GetMeta(key); exists {
            snapshot.Metadata[key] = value
        }
    }
    
    st.states = append(st.states, snapshot)
}

func (st *StateTracker) PrintStateFlow() {
    st.mu.Lock()
    defer st.mu.Unlock()
    
    fmt.Printf("\\n=== State Flow Analysis ===\\n")
    for i, snapshot := range st.states {
        fmt.Printf("[%d] %s - %s (%s) - %s\\n",
            i,
            snapshot.Timestamp.Format("15:04:05.000"),
            snapshot.AgentName,
            snapshot.AgentRole,
            snapshot.Action)
        
        fmt.Printf("    Data Keys: %v\\n", snapshot.Keys)
        fmt.Printf("    Meta Keys: %v\\n", snapshot.MetaKeys)
        
        // Print sample values (limit output)
        for key, value := range snapshot.Values {
            fmt.Printf("    %s: %v\\n", key, value)
        }
        
        // Print metadata
        if len(snapshot.Metadata) > 0 {
            fmt.Printf("    Metadata: %v\\n", snapshot.Metadata)
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

::: info pprof Integration
Enable pprof for detailed performance analysis. Access the web interface at `http://localhost:6060/debug/pprof/`
:::

```go
package main

import (

    "net/http"
    _ "net/http/pprof"
    "runtime"
    "time"
)
// Enable pprof endpoint for profiling
enableProfiling() {
    go func() {
        core.Logger().Info().Msg("Starting pprof server on :6060")
        core.Logger().Error().Err(http.ListenAndServe("localhost:6060", nil)).Msg("pprof server stopped")
    }()
}
// Profile agent execution using current Agent interface
profileAgentExecution(agent core.Agent, state core.State) {
    start := time.Now()
    var m1, m2 runtime.MemStats
    runtime.ReadMemStats(&m1)
    runtime.GC() // Force GC to get accurate memory measurements
    ctx := context.Background()
    result, err := agent.Run(ctx, state)
    runtime.ReadMemStats(&m2)
    duration := time.Since(start)
    // Calculate memory usage
    memUsed := m2.Alloc - m1.Alloc
    core.Logger().Info().
        Str("agent", agent.Name()).
        Str("role", agent.GetRole()).
        Dur("duration", duration).
        Uint64("memory_used", memUsed).
        Int("goroutines", runtime.NumGoroutine()).
        Uint32("gc_runs", m2.NumGC-m1.NumGC).
        Msg("Agent performance profile")
    if err != nil {
        core.Logger().Error().
            Str("agent", agent.Name()).
            Err(err).
            Msg("Agent execution failed during profiling")
    } else {
        core.Logger().Info().
            Str("agent", agent.Name()).
            Strs("result_keys", result.Keys()).
            Msg("Agent execution completed successfully")
    }
}
// Benchmark agent performance using current Agent interface
benchmarkAgent(agent core.Agent, iterations int) {
    state := core.NewState()
    state.Set("benchmark", true)
    state.Set("test", "performance")
    durations := make([]time.Duration, iterations)
    var totalMemory uint64
    for i := 0; i < iterations; i++ {
        var m1, m2 runtime.MemStats
        runtime.ReadMemStats(&m1)
        start := time.Now()
        ctx := context.Background()
        _, err := agent.Run(ctx, state)
        duration := time.Since(start)
        runtime.ReadMemStats(&m2)
        durations[i] = duration
        totalMemory += (m2.Alloc - m1.Alloc)
        if err != nil {
            core.Logger().Error().
                Str("agent", agent.Name()).
                Int("iteration", i).
                Err(err).
                Msg("Benchmark iteration failed")
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
    core.Logger().Info().
        Str("agent", agent.Name()).
        Str("role", agent.GetRole()).
        Int("iterations", iterations).
        Dur("avg_duration", avgDuration).
        Dur("min_duration", minDuration).
        Dur("max_duration", maxDuration).
        Uint64("avg_memory", avgMemory).
        Msg("Agent benchmark results")
}

```

#### Resource Monitoring

```go
// Monitor agent resource usage over time using current Agent interface
type ResourceMonitor struct {
    agent     core.Agent
    metrics   []ResourceMetric
    mu        sync.Mutex
    stopChan  chan struct{}
    interval  time.Duration
}

type ResourceMetric struct {
    Timestamp   time.Time
    AgentName   string
    AgentRole   string
    CPUPercent  float64
    MemoryBytes uint64
    Goroutines  int
    GCPauses    time.Duration
}

func &ResourceMonitor{agent core.Agent, interval time.Duration) *ResourceMonitor {
    return &ResourceMonitor{
        agent:    agent,
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
        AgentName:   rm.agent.Name(),
        AgentRole:   rm.agent.GetRole(),
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
        core.Logger().Warn().
            Str("agent", rm.agent.Name()).
            Msg("No metrics collected for agent")
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
    
    core.Logger().Info().
        Str("agent", rm.agent.Name()).
        Str("role", rm.agent.GetRole()).
        Int("samples", len(metrics)).
        Float64("avg_cpu_percent", avgCPU).
        Float64("avg_memory_mb", avgMemory/1024/1024).
        Float64("peak_memory_mb", float64(maxMemory)/1024/1024).
        Int("max_goroutines", maxGoroutines).
        Msg("Resource summary for agent")
}
```

## Debugging Tools and Utilities

### 1. Agent Inspector

```go
// Comprehensive agent inspection utility using current Agent interface
type AgentInspector struct {
    agent   core.Agent
    history []InspectionRecord
    mu      sync.Mutex
}

type InspectionRecord struct {
    Timestamp   time.Time
    AgentName   string
    AgentRole   string
    InputState  core.State
    OutputState core.State
    Error       error
    Duration    time.Duration
}

func &AgentInspector{agent core.Agent) *AgentInspector {
    return &AgentInspector{
        agent:   agent,
        history: make([]InspectionRecord, 0),
    }
}

func (ai *AgentInspector) Run(ctx context.Context, state core.State) (core.State, error) {
    start := time.Now()
    
    // Clone input state to preserve it
    inputState := state.Clone()
    
    result, err := ai.agent.Run(ctx, state)
    
    duration := time.Since(start)
    
    record := InspectionRecord{
        Timestamp:   start,
        AgentName:   ai.agent.Name(),
        AgentRole:   ai.agent.GetRole(),
        InputState:  inputState,
        OutputState: result,
        Error:       err,
        Duration:    duration,
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

// Implement other Agent interface methods for inspector wrapper
func (ai *AgentInspector) Name() string { return ai.agent.Name() }
func (ai *AgentInspector) GetRole() string { return ai.agent.GetRole() }
func (ai *AgentInspector) GetDescription() string { return ai.agent.GetDescription() }
func (ai *AgentInspector) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return ai.agent.HandleEvent(ctx, event, state)
}
func (ai *AgentInspector) GetCapabilities() []string { return ai.agent.GetCapabilities() }
func (ai *AgentInspector) GetSystemPrompt() string { return ai.agent.GetSystemPrompt() }
func (ai *AgentInspector) GetTimeout() time.Duration { return ai.agent.GetTimeout() }
func (ai *AgentInspector) IsEnabled() bool { return ai.agent.IsEnabled() }
func (ai *AgentInspector) GetLLMConfig() *core.ResolvedLLMConfig { return ai.agent.GetLLMConfig() }
func (ai *AgentInspector) Initialize(ctx context.Context) error { return ai.agent.Initialize(ctx) }
func (ai *AgentInspector) Shutdown(ctx context.Context) error { return ai.agent.Shutdown(ctx) }

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
        
        fmt.Printf("[%d] %s - %s (%s) (%v) - %s\\n",
            i,
            record.Timestamp.Format("15:04:05.000"),
            record.AgentName,
            record.AgentRole,
            record.Duration,
            status)
        
        fmt.Printf("    Input Keys: %v\\n", record.InputState.Keys())
        if record.OutputState != nil {
            fmt.Printf("    Output Keys: %v\\n", record.OutputState.Keys())
        }
        
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
            "total_executions": 0,
        }
    }
    
    var totalDuration time.Duration
    var errorCount int
    agentRoles := make(map[string]int)
    
    for _, record := range history {
        totalDuration += record.Duration
        if record.Error != nil {
            errorCount++
        }
        agentRoles[record.AgentRole]++
    }
    
    return map[string]interface{}{
        "total_executions":  len(history),
        "error_count":       errorCount,
        "error_rate":        float64(errorCount) / float64(len(history)),
        "average_duration":  totalDuration / time.Duration(len(history)),
        "agent_roles":       agentRoles,
        "agent_name":        ai.agent.Name(),
        "capabilities":      ai.agent.GetCapabilities(),
    }
}
```

### 2. System Health Checker

```go
// Comprehensive system health checking using current Agent interface
type HealthChecker struct {
    agents []core.Agent
    runner *core.Runner
    checks []HealthCheck
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

func &HealthChecker{runner *core.Runner, agents ...core.Agent) *HealthChecker {
    hc := &HealthChecker{
        agents: agents,
        runner: runner,
        checks: make([]HealthCheck, 0),
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
    healthState := core.NewState()
    healthState.Set("health_check", true)
    healthState.Set("timestamp", time.Now())
    healthState.SetMeta("check_id", uuid.New().String())
    
    for _, agent := range hc.agents {
        if !agent.IsEnabled() {
            continue // Skip disabled agents
        }
        
        _, err := agent.Run(ctx, healthState)
        if err != nil {
            return fmt.Errorf("agent %s (%s) failed health check: %w", 
                agent.Name(), agent.GetRole(), err)
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

::: tip Debugging Philosophy
Focus on observability first - implement comprehensive logging and tracing before you need it.
:::

### 1. Structured Logging
- Use consistent log formats across all agents with the `core.Logger()`
- Include correlation IDs in State metadata to track requests
- Log at appropriate levels (DEBUG, INFO, WARN, ERROR)
- Include context information (agent name, role, capabilities)

### 2. Error Handling
- Wrap errors with context using `fmt.Errorf`
- Use typed errors for different failure modes
- Implement callback-based error handling for resilience
- Provide actionable error messages with agent context

### 3. Testing Strategies
- Write unit tests for individual agents using the Agent interface
- Create integration tests for agent interactions through Runner
- Use mocking to isolate components with interface composition
- Implement chaos testing for resilience using callback injection

### 4. Monitoring
- Set up callback-based metrics for key performance indicators
- Use agentcli for trace analysis of complex workflows
- Implement health checks using the Agent interface
- Monitor resource usage trends through structured logging

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
ecks))
    
    for _, check := range hc.checks {
        result := hc.runSingleCheck(ctx, check)
        results = append(results, result)
    }
    
    return results
}

func (hc *HealthChecker) runSingleCheck(ctx context.Context, check HealthCheck) HealthResult {
    start := time.Now()
    
    // Create timeout context for this check
    checkCtx, cancel := context.WithTimeout(ctx, check.Timeout)
    defer cancel()
    
    result := HealthResult{
        CheckName: check.Name,
        Timestamp: start,
    }
    
    // Run the check with timeout
    done := make(chan error, 1)
    go func() {
        done <- check.CheckFunc(checkCtx)
    }()
    
    select {
    case err := <-done:
        result.Duration = time.Since(start)
        if err != nil {
            result.Status = "FAIL"
            result.Error = err
        } else {
            result.Status = "PASS"
        }
    case <-checkCtx.Done():
        result.Duration = time.Since(start)
        result.Status = "TIMEOUT"
        result.Error = checkCtx.Err()
    }
    
    return result
}

func (hc *HealthChecker) checkAgentResponsiveness(ctx context.Context) error {
    for _, agent := range hc.agents {
        if err := checkAgentHealth(ctx, agent); err != nil {
            return fmt.Errorf("agent %s failed health check: %w", agent.Name(), err)
        }
    }
    return nil
}

func (hc *HealthChecker) checkMemoryUsage(ctx context.Context) error {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    // Check if memory usage is above threshold (1GB)
    if m.Alloc > 1024*1024*1024 {
        return fmt.Errorf("high memory usage: %d MB", m.Alloc/1024/1024)
    }
    
    return nil
}

func (hc *HealthChecker) checkGoroutineCount(ctx context.Context) error {
    count := runtime.NumGoroutine()
    
    // Check if goroutine count is above threshold (1000)
    if count > 1000 {
        return fmt.Errorf("high goroutine count: %d", count)
    }
    
    return nil
}

func (hc *HealthChecker) PrintResults(results []HealthResult) {
    fmt.Printf("\n=== Health Check Results ===\n")
    
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
        
        fmt.Printf("[%s] %s - %s (%v)\n",
            status,
            result.CheckName,
            result.Timestamp.Format("15:04:05.000"),
            result.Duration)
        
        if result.Error != nil {
            fmt.Printf("    Error: %v\n", result.Error)
        }
    }
    
    fmt.Printf("\nSummary: %d passed, %d failed, %d timed out\n", passCount, failCount, timeoutCount)
    fmt.Printf("=== End Health Check Results ===\n\n")
}
```

## Using agentcli for Debugging

### Trace Collection and Analysis

The `agentcli` tool provides powerful tracing capabilities for debugging multi-agent systems:

```bash
# Start trace collection for a specific session
agentcli trace start --session-id "debug-session-001" --output traces.json

# Run your agent system
go run main.go

# Stop trace collection
agentcli trace stop --session-id "debug-session-001"

# Analyze collected traces
agentcli trace analyze --file traces.json --format table
```

### Trace Analysis Examples

```go
// Custom trace analysis using agentcli output
analyzeTraceFile(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("failed to read trace file: %w", err)
    }
    
    var traces []core.TraceEntry
    if err := json.Unmarshal(data, &traces); err != nil {
        return fmt.Errorf("failed to parse trace data: %w", err)
    }
    
    // Analyze agent execution patterns
    agentStats := make(map[string]*AgentStats)
    
    for _, trace := range traces {
        if _, exists := agentStats[trace.AgentID]; !exists {
            agentStats[trace.AgentID] = &AgentStats{
                AgentID: trace.AgentID,
                Executions: 0,
                TotalDuration: 0,
                Errors: 0,
            }
        }
        
        stats := agentStats[trace.AgentID]
        stats.Executions++
        
        if trace.Error != "" {
            stats.Errors++
        }
        
        // Calculate duration if we have start/end pairs
        if trace.Hook == core.HookAfterAgentRun {
            // Find corresponding start trace
            for _, startTrace := range traces {
                if startTrace.AgentID == trace.AgentID && 
                   startTrace.Hook == core.HookBeforeAgentRun &&
                   startTrace.EventID == trace.EventID {
                    duration := trace.Timestamp.Sub(startTrace.Timestamp)
                    stats.TotalDuration += duration
                    break
                }
            }
        }
    }
    
    // Print analysis results
    fmt.Printf("\n=== Trace Analysis Results ===\n")
    for agentID, stats := range agentStats {
        avgDuration := time.Duration(0)
        if stats.Executions > 0 {
            avgDuration = stats.TotalDuration / time.Duration(stats.Executions)
        }
        
        errorRate := float64(stats.Errors) / float64(stats.Executions) * 100
        
        fmt.Printf("Agent: %s\n", agentID)
        fmt.Printf("  Executions: %d\n", stats.Executions)
        fmt.Printf("  Average Duration: %v\n", avgDuration)
        fmt.Printf("  Error Rate: %.2f%%\n", errorRate)
        fmt.Printf("  Total Errors: %d\n", stats.Errors)
        fmt.Printf("\n")
    }
    
    return nil
}

type AgentStats struct {
    AgentID       string
    Executions    int
    TotalDuration time.Duration
    Errors        int
}
```

### Real-time Monitoring with agentcli

```bash
# Monitor agent execution in real-time
agentcli monitor --agents "research-agent,analysis-agent" --interval 5s

# Monitor specific metrics
agentcli monitor --metrics "memory,cpu,goroutines" --format json

# Set up alerts for specific conditions
agentcli monitor --alert-on "error_rate>5%" --alert-on "memory>500MB"
```

## Advanced Debugging Patterns

### 1. Distributed Tracing Integration

```go
package main

import (

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)
// Agent wrapper with OpenTelemetry tracing
type TracedAgent struct {
    agent  core.Agent
    tracer trace.Tracer
}
func NewTracedAgent(agent core.Agent) *TracedAgent {
    return &TracedAgent{
        agent:  agent,
        tracer: otel.Tracer("agenticgokit"),
    }
}
func (ta *TracedAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Start a new span for this agent execution
    ctx, span := ta.tracer.Start(ctx, fmt.Sprintf("agent.%s.run", ta.agent.Name()))
    defer span.End()
    // Add agent metadata to span
    span.SetAttributes(
        attribute.String("agent.name", ta.agent.Name()),
        attribute.String("agent.role", ta.agent.GetRole()),
        attribute.StringSlice("agent.capabilities", ta.agent.GetCapabilities()),
        attribute.Int("state.keys.count", len(state.Keys())),
    )
    // Execute the agent
    result, err := ta.agent.Run(ctx, state)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    } else {
        span.SetStatus(codes.Ok, "Agent execution completed successfully")
        span.SetAttributes(
            attribute.Int("result.keys.count", len(result.Keys())),
        )
    }
    return result, err
}
// Implement other Agent interface methods...
func (ta *TracedAgent) Name() string { return ta.agent.Name() }
func (ta *TracedAgent) GetRole() string { return ta.agent.GetRole() }
func (ta *TracedAgent) GetDescription() string { return ta.agent.GetDescription() }
func (ta *TracedAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return ta.agent.HandleEvent(ctx, event, state)
}
func (ta *TracedAgent) GetCapabilities() []string { return ta.agent.GetCapabilities() }
func (ta *TracedAgent) GetSystemPrompt() string { return ta.agent.GetSystemPrompt() }
func (ta *TracedAgent) GetTimeout() time.Duration { return ta.agent.GetTimeout() }
func (ta *TracedAgent) IsEnabled() bool { return ta.agent.IsEnabled() }
func (ta *TracedAgent) GetLLMConfig() *core.ResolvedLLMConfig { return ta.agent.GetLLMConfig() }
func (ta *TracedAgent) Initialize(ctx context.Context) error { return ta.agent.Initialize(ctx) }
func (ta *TracedAgent) Shutdown(ctx context.Context) error { return ta.agent.Shutdown(ctx) }

```

### 2. Circuit Breaker Pattern for Debugging

```go
// Circuit breaker for agent debugging
type CircuitBreakerAgent struct {
    agent           core.Agent
    failureCount    int
    lastFailureTime time.Time
    state          CircuitState
    threshold      int
    timeout        time.Duration
    mu             sync.Mutex
}

type CircuitState int

const (
    Closed CircuitState = iota
    Open
    HalfOpen
)

func &CircuitBreakerAgent{agent core.Agent, threshold int, timeout time.Duration) *CircuitBreakerAgent {
    return &CircuitBreakerAgent{
        agent:     agent,
        threshold: threshold,
        timeout:   timeout,
        state:     Closed,
    }
}

func (cb *CircuitBreakerAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    // Check circuit state
    switch cb.state {
    case Open:
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.state = HalfOpen
            core.Logger().Info().
                Str("agent", cb.agent.Name()).
                Msg("Circuit breaker transitioning to half-open")
        } else {
            return state, fmt.Errorf("circuit breaker is open for agent %s", cb.agent.Name())
        }
    case HalfOpen:
        // Allow one request through
    case Closed:
        // Normal operation
    }
    
    // Execute the agent
    result, err := cb.agent.Run(ctx, state)
    
    if err != nil {
        cb.failureCount++
        cb.lastFailureTime = time.Now()
        
        if cb.failureCount >= cb.threshold {
            cb.state = Open
            core.Logger().Warn().
                Str("agent", cb.agent.Name()).
                Int("failure_count", cb.failureCount).
                Msg("Circuit breaker opened due to failures")
        }
        
        return result, err
    }
    
    // Success - reset failure count and close circuit
    cb.failureCount = 0
    cb.state = Closed
    
    return result, nil
}

// Implement other Agent interface methods...
func (cb *CircuitBreakerAgent) Name() string { return cb.agent.Name() }
func (cb *CircuitBreakerAgent) GetRole() string { return cb.agent.GetRole() }
func (cb *CircuitBreakerAgent) GetDescription() string { return cb.agent.GetDescription() }
func (cb *CircuitBreakerAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return cb.agent.HandleEvent(ctx, event, state)
}
func (cb *CircuitBreakerAgent) GetCapabilities() []string { return cb.agent.GetCapabilities() }
func (cb *CircuitBreakerAgent) GetSystemPrompt() string { return cb.agent.GetSystemPrompt() }
func (cb *CircuitBreakerAgent) GetTimeout() time.Duration { return cb.agent.GetTimeout() }
func (cb *CircuitBreakerAgent) IsEnabled() bool { return cb.agent.IsEnabled() }
func (cb *CircuitBreakerAgent) GetLLMConfig() *core.ResolvedLLMConfig { return cb.agent.GetLLMConfig() }
func (cb *CircuitBreakerAgent) Initialize(ctx context.Context) error { return cb.agent.Initialize(ctx) }
func (cb *CircuitBreakerAgent) Shutdown(ctx context.Context) error { return cb.agent.Shutdown(ctx) }
```

## Production Debugging Strategies

### 1. Graceful Degradation

```go
// Agent with graceful degradation for production debugging
type ResilientAgent struct {
    primary   core.Agent
    fallback  core.Agent
    monitor   *ResourceMonitor
}

func &ResilientAgent{primary, fallback core.Agent) *ResilientAgent {
    return &ResilientAgent{
        primary:  primary,
        fallback: fallback,
        monitor:  &ResourceMonitor{primary, 30*time.Second),
    }
}

func (ra *ResilientAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Try primary agent first
    result, err := ra.tryPrimaryAgent(ctx, state)
    if err == nil {
        return result, nil
    }
    
    // Log primary failure and try fallback
    core.Logger().Warn().
        Str("primary_agent", ra.primary.Name()).
        Str("fallback_agent", ra.fallback.Name()).
        Err(err).
        Msg("Primary agent failed, trying fallback")
    
    return ra.fallback.Run(ctx, state)
}

func (ra *ResilientAgent) tryPrimaryAgent(ctx context.Context, state core.State) (core.State, error) {
    // Set a timeout for primary agent
    timeout := ra.primary.GetTimeout()
    if timeout == 0 {
        timeout = 30 * time.Second
    }
    
    primaryCtx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    return ra.primary.Run(primaryCtx, state)
}

// Implement other Agent interface methods...
func (ra *ResilientAgent) Name() string { return ra.primary.Name() }
func (ra *ResilientAgent) GetRole() string { return ra.primary.GetRole() }
func (ra *ResilientAgent) GetDescription() string { return ra.primary.GetDescription() }
func (ra *ResilientAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return ra.primary.HandleEvent(ctx, event, state)
}
func (ra *ResilientAgent) GetCapabilities() []string { return ra.primary.GetCapabilities() }
func (ra *ResilientAgent) GetSystemPrompt() string { return ra.primary.GetSystemPrompt() }
func (ra *ResilientAgent) GetTimeout() time.Duration { return ra.primary.GetTimeout() }
func (ra *ResilientAgent) IsEnabled() bool { return ra.primary.IsEnabled() }
func (ra *ResilientAgent) GetLLMConfig() *core.ResolvedLLMConfig { return ra.primary.GetLLMConfig() }
func (ra *ResilientAgent) Initialize(ctx context.Context) error { return ra.primary.Initialize(ctx) }
func (ra *ResilientAgent) Shutdown(ctx context.Context) error { return ra.primary.Shutdown(ctx) }
```

### 2. Debug Mode Configuration

```toml
[logging]
level = "debug"
format = "json"

[agent_flow]
name = "debug-system"
version = "1.0.0"

[runtime]
max_concurrent_agents = 5

[agents.debug-monitor]
role = "monitor"
description = "System monitoring and debugging agent"
enabled = true
capabilities = ["monitoring", "health-check", "profiling"]
system_prompt = "Monitor system health and performance."
timeout = 30
cpu_threshold = "80%"
error_rate_threshold = "5%"
```

## Best Practices Summary

### Development Best Practices

1. **Always Use Context with Timeouts**
   ```go
   func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
       timeout := a.GetTimeout()
       if timeout == 0 {
           timeout = 30 * time.Second // Default timeout
       }
       
       workCtx, cancel := context.WithTimeout(ctx, timeout)
       defer cancel()
       
       return a.doWork(workCtx, state)
   }
   ```

2. **Implement Proper Error Handling**
   ```go
   func (a *MyAgent) processData(ctx context.Context, data interface{}) error {
       if err := a.validateData(data); err != nil {
           return fmt.Errorf("data validation failed: %w", err)
       }
       
       if err := a.processValidData(ctx, data); err != nil {
           return fmt.Errorf("data processing failed: %w", err)
       }
       
       return nil
   }
   ```

3. **Use Structured Logging**
   ```go
   core.Logger().Info().
       Str("agent", a.Name()).
       Str("operation", "data_processing").
       Int("data_size", len(data)).
       Dur("duration", duration).
       Msg("Data processing completed")
   ```

### Production Best Practices

1. **Monitor Resource Usage**
2. **Implement Health Checks**
3. **Use Circuit Breakers for External Dependencies**
4. **Set Up Proper Alerting**
5. **Maintain Debug Logs with Appropriate Levels**

## Troubleshooting Common Issues

### Issue: Agent Deadlock

**Symptoms**: Agent stops responding, high CPU usage, no progress

**Solution**:
```go
// Use Go's race detector
go run -race main.go

// Add timeout to all operations
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Monitor goroutine count
go func() {
    for {
        count := runtime.NumGoroutine()
        if count > 1000 {
            log.Printf("High goroutine count: %d", count)
        }
        time.Sleep(10 * time.Second)
    }
}()
```

### Issue: Memory Leaks

**Symptoms**: Gradually increasing memory usage, eventual OOM

**Solution**:
```go
// Regular memory profiling
go func() {
    for {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        if m.Alloc > 500*1024*1024 { // 500MB threshold
            log.Printf("High memory usage: %d MB", m.Alloc/1024/1024)
            
            // Force garbage collection
            runtime.GC()
            
            // Create heap dump for analysis
            if m.Alloc > 1024*1024*1024 { // 1GB threshold
                createHeapDump()
            }
        }
        
        time.Sleep(60 * time.Second)
    }
}()
```

### Issue: Slow Agent Performance

**Symptoms**: High latency, timeouts, poor throughput

**Solution**:
```go
// Profile agent execution
profileSlowAgent(agent core.Agent) {
    // Enable CPU profiling
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
    
    // Run agent with profiling
    start := time.Now()
    result, err := agent.Run(context.Background(), core.NewState())
    duration := time.Since(start)
    
    log.Printf("Agent execution took %v", duration)
    
    // Analyze results
    if duration > 5*time.Second {
        log.Printf("Slow execution detected for agent %s", agent.Name())
    }
}
```

## Conclusion

Debugging multi-agent systems requires a systematic approach combining health monitoring, performance profiling, and trace analysis. AgenticGoKit's unified Agent interface and callback system provide powerful tools for identifying and resolving issues in complex agent interactions. By following the patterns and techniques outlined in this guide, you can effectively troubleshoot and optimize your multi-agent applications.

## Quick Reference

::: details Agent Health Check
```go
// Quick agent health check
quickHealthCheck(ctx context.Context, agent core.Agent) error {
    healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    state := core.NewState()
    state.Set("health_check", true)
    
    _, err := agent.Run(healthCtx, state)
    return err
}
```
:::

::: details Performance Monitoring
```go
// Basic performance monitoring
monitorAgent(agent core.Agent) {
    start := time.Now()
    var m1, m2 runtime.MemStats
    
    runtime.ReadMemStats(&m1)
    // Execute agent...
    runtime.ReadMemStats(&m2)
    
    duration := time.Since(start)
    memUsed := m2.Alloc - m1.Alloc
    
    core.Logger().Info().
        Str("agent", agent.Name()).
        Dur("duration", duration).
        Uint64("memory_used", memUsed).
        Msg("Agent performance")
}
```
:::

::: details Common Debug Commands
```bash
# Essential debugging commands
go run -race main.go              # Enable race detection
go tool pprof cpu.prof            # Analyze CPU profile
go tool pprof mem.prof            # Analyze memory profile
agentcli trace <session-id>       # View execution trace
```
:::

## Next Steps

::: info Continue Learning
Explore these related topics to master debugging multi-agent systems.
:::

- Learn about [Logging and Tracing](./logging-and-tracing) for detailed logging strategies
- Explore [Error Handling](../core-concepts/error-handling) patterns
- Review [Agent Lifecycle](../core-concepts/agent-lifecycle) for better understanding
- Check out [State Management](../core-concepts/state-management) debugging techniques

## Related Resources

::: info External Resources
Explore these external resources for additional debugging knowledge and tools.
:::

- **[AgenticGoKit Core Documentation](https://github.com/kunalkushwaha/agenticgokit)** - Official repository and documentation
- **[Go Debugging Tools](https://golang.org/doc/debugging)** - Official Go debugging guide and tools
- **[pprof Profiling Guide](https://golang.org/pkg/net/http/pprof/)** - Go's built-in profiling tools documentation
- **[OpenTelemetry Go Documentation](https://opentelemetry.io/docs/go/)** - Advanced tracing and observability patterns
## Com
mon Debugging Pitfalls and Solutions

### Pitfall 1: Race Conditions in Agent State

::: danger Race Condition Issues
Race conditions occur when multiple agents access shared state simultaneously without proper synchronization.
:::

**Problem**: Agents modifying state concurrently leading to inconsistent results.

**Solution**: Use proper state management patterns with the current State interface.

```go
// ❌ Problematic: Direct state modification without synchronization
func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    counter, _ := state.Get("counter")
    newCounter := counter.(int) + 1
    state.Set("counter", newCounter) // Race condition!
    return state, nil
}

// ✅ Correct: Use state cloning and atomic operations
func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    result := state.Clone() // Create isolated copy
    
    counter, exists := result.Get("counter")
    if !exists {
        counter = 0
    }
    
    newCounter := counter.(int) + 1
    result.Set("counter", newCounter)
    result.SetMeta("updated_by", a.Name())
    
    return result, nil
}
```

### Pitfall 2: Memory Leaks in Long-Running Agents

::: danger Memory Management
Agents that don't properly clean up resources can cause memory leaks in long-running systems.
:::

**Problem**: Accumulating memory usage over time due to unclosed resources.

**Solution**: Implement proper resource management and cleanup patterns.

```go
// ❌ Problematic: No resource cleanup
type LeakyAgent struct {
    connections []*http.Client
    buffers     [][]byte
}

func (l *LeakyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    client := &http.Client{} // Never cleaned up
    l.connections = append(l.connections, client)
    
    buffer := make([]byte, 1024*1024) // 1MB buffer never freed
    l.buffers = append(l.buffers, buffer)
    
    return state, nil
}

// ✅ Correct: Proper resource management
type CleanAgent struct {
    name   string
    client *http.Client // Reuse connection
}

func NewCleanAgent(name string) *CleanAgent {
    return &CleanAgent{
        name: name,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *CleanAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Use pooled buffer
    buffer := make([]byte, 1024)
    defer func() {
        // Buffer will be garbage collected when function exits
        buffer = nil
    }()
    
    // Reuse HTTP client
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)
    resp, err := c.client.Do(req)
    if err != nil {
        return state, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close() // Always close response body
    
    result := state.Clone()
    result.Set("http_status", resp.StatusCode)
    
    return result, nil
}

func (c *CleanAgent) Shutdown(ctx context.Context) error {
    // Clean up resources
    if c.client != nil {
        c.client.CloseIdleConnections()
    }
    return nil
}
```

### Pitfall 3: Inadequate Error Context

::: danger Error Handling
Generic error messages make debugging extremely difficult in complex multi-agent systems.
:::

**Problem**: Errors without sufficient context for debugging.

**Solution**: Provide rich error context with agent and state information.

```go
// ❌ Problematic: Generic error messages
func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    if err := a.validateInput(state); err != nil {
        return state, fmt.Errorf("validation failed") // No context!
    }
    return state, nil
}

// ✅ Correct: Rich error context
func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    if err := a.validateInput(state); err != nil {
        // Include agent context, state information, and error chain
        return state, fmt.Errorf("agent %s validation failed for state with keys %v: %w", 
            a.Name(), state.Keys(), err)
    }
    
    // Add correlation ID for tracing
    correlationID, _ := state.GetMeta("correlation_id")
    
    core.Logger().Info().
        Str("agent", a.Name()).
        Str("correlation_id", correlationID).
        Strs("input_keys", state.Keys()).
        Msg("Agent processing started")
    
    return state, nil
}

func (a *MyAgent) validateInput(state core.State) error {
    requiredKeys := []string{"input_data", "user_id"}
    
    for _, key := range requiredKeys {
        if _, exists := state.Get(key); !exists {
            return fmt.Errorf("missing required key: %s", key)
        }
    }
    
    return nil
}
```

### Pitfall 4: Blocking Operations Without Timeouts

::: danger Timeout Management
Operations without timeouts can cause agents to hang indefinitely.
:::

**Problem**: Agents hanging on external calls or long-running operations.

**Solution**: Always use context with timeouts for external operations.

```go
// ❌ Problematic: No timeout on external calls
func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    resp, err := http.Get("https://slow-api.example.com") // Can hang forever
    if err != nil {
        return state, err
    }
    defer resp.Body.Close()
    
    return state, nil
}

// ✅ Correct: Proper timeout handling
func (a *MyAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Create timeout context for external call
    callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    // Create request with context
    req, err := http.NewRequestWithContext(callCtx, "GET", "https://slow-api.example.com", nil)
    if err != nil {
        return state, fmt.Errorf("failed to create request: %w", err)
    }
    
    client := &http.Client{
        Timeout: 15 * time.Second, // Additional client-level timeout
    }
    
    resp, err := client.Do(req)
    if err != nil {
        // Check if it was a timeout
        if callCtx.Err() == context.DeadlineExceeded {
            return state, fmt.Errorf("agent %s: external API call timed out after 10s", a.Name())
        }
        return state, fmt.Errorf("agent %s: external API call failed: %w", a.Name(), err)
    }
    defer resp.Body.Close()
    
    result := state.Clone()
    result.Set("api_response_status", resp.StatusCode)
    result.SetMeta("api_call_duration", time.Since(time.Now()).String())
    
    return result, nil
}
```

## Performance Debugging Strategies

### CPU Profiling for Agent Performance

Use Go's built-in profiling tools to identify CPU bottlenecks in agent execution:

```go
import (
    _ "net/http/pprof"
    "net/http"
    "runtime/pprof"
)

// Enable CPU profiling in your debugging setup
func enableCPUProfiling() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}

// Profile specific agent execution
func profileAgentCPU(agent core.Agent, state core.State) error {
    // Start CPU profiling
    f, err := os.Create("agent_cpu.prof")
    if err != nil {
        return err
    }
    defer f.Close()
    
    if err := pprof.StartCPUProfile(f); err != nil {
        return err
    }
    defer pprof.StopCPUProfile()
    
    // Execute agent multiple times for profiling
    ctx := context.Background()
    for i := 0; i < 100; i++ {
        _, err := agent.Run(ctx, state)
        if err != nil {
            core.Logger().Error().Err(err).Msg("Agent execution failed during profiling")
        }
    }
    
    return nil
}
```

### Memory Profiling for Agent Memory Usage

Track memory allocation patterns in agent execution:

```go
import (
    "runtime"
    "runtime/pprof"
)

// Profile agent memory usage
func profileAgentMemory(agent core.Agent, state core.State) error {
    // Force garbage collection before profiling
    runtime.GC()
    
    // Create memory profile
    f, err := os.Create("agent_mem.prof")
    if err != nil {
        return err
    }
    defer f.Close()
    
    // Execute agent
    ctx := context.Background()
    _, err = agent.Run(ctx, state)
    if err != nil {
        return err
    }
    
    // Write memory profile
    runtime.GC() // Force GC to get accurate memory stats
    if err := pprof.WriteHeapProfile(f); err != nil {
        return err
    }
    
    return nil
}

// Monitor memory usage during agent execution
func monitorAgentMemory(agent core.Agent, state core.State) {
    var m1, m2 runtime.MemStats
    
    runtime.ReadMemStats(&m1)
    
    ctx := context.Background()
    result, err := agent.Run(ctx, state)
    
    runtime.ReadMemStats(&m2)
    
    memUsed := m2.Alloc - m1.Alloc
    
    core.Logger().Info().
        Str("agent", agent.Name()).
        Uint64("memory_used_bytes", memUsed).
        Uint64("memory_used_kb", memUsed/1024).
        Int("goroutines", runtime.NumGoroutine()).
        Msg("Agent memory usage")
    
    // Alert on high memory usage
    if memUsed > 10*1024*1024 { // 10MB threshold
        core.Logger().Warn().
            Str("agent", agent.Name()).
            Uint64("memory_used_mb", memUsed/1024/1024).
            Msg("High memory usage detected")
    }
}
```

## Production Debugging Best Practices

### 1. Structured Logging for Production

Implement comprehensive structured logging for production debugging:

```go
// Production-ready logging setup
func setupProductionLogging(agent core.Agent) zerolog.Logger {
    return core.Logger().With().
        Str("agent_name", agent.Name()).
        Str("agent_role", agent.GetRole()).
        Strs("capabilities", agent.GetCapabilities()).
        Str("environment", "production").
        Logger()
}

// Enhanced agent wrapper with production logging
type ProductionAgent struct {
    agent  core.Agent
    logger zerolog.Logger
    metrics *AgentMetrics
}

type AgentMetrics struct {
    ExecutionCount    int64
    ErrorCount        int64
    TotalDuration     time.Duration
    LastExecution     time.Time
    mu                sync.Mutex
}

func NewProductionAgent(agent core.Agent) *ProductionAgent {
    return &ProductionAgent{
        agent:   agent,
        logger:  setupProductionLogging(agent),
        metrics: &AgentMetrics{},
    }
}

func (p *ProductionAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    start := time.Now()
    
    // Update metrics
    p.metrics.mu.Lock()
    p.metrics.ExecutionCount++
    p.metrics.LastExecution = start
    p.metrics.mu.Unlock()
    
    // Extract correlation ID for tracing
    correlationID, _ := state.GetMeta("correlation_id")
    
    // Log execution start
    p.logger.Info().
        Str("correlation_id", correlationID).
        Strs("input_keys", state.Keys()).
        Time("start_time", start).
        Msg("Agent execution started")
    
    // Execute agent
    result, err := p.agent.Run(ctx, state)
    
    duration := time.Since(start)
    
    // Update metrics
    p.metrics.mu.Lock()
    p.metrics.TotalDuration += duration
    if err != nil {
        p.metrics.ErrorCount++
    }
    p.metrics.mu.Unlock()
    
    // Log execution result
    if err != nil {
        p.logger.Error().
            Str("correlation_id", correlationID).
            Err(err).
            Dur("duration", duration).
            Msg("Agent execution failed")
    } else {
        p.logger.Info().
            Str("correlation_id", correlationID).
            Dur("duration", duration).
            Strs("output_keys", result.Keys()).
            Msg("Agent execution completed")
    }
    
    return result, err
}

func (p *ProductionAgent) GetMetrics() AgentMetrics {
    p.metrics.mu.Lock()
    defer p.metrics.mu.Unlock()
    
    return *p.metrics
}
```

### 2. Health Monitoring and Alerting

Implement comprehensive health monitoring for production systems:

```go
// Production health monitor
type ProductionHealthMonitor struct {
    agents          map[string]*ProductionAgent
    alertThresholds HealthThresholds
    alertChannel    chan HealthAlert
    mu              sync.RWMutex
}

type HealthThresholds struct {
    MaxErrorRate      float64       // Maximum acceptable error rate (0.0-1.0)
    MaxResponseTime   time.Duration // Maximum acceptable response time
    MinSuccessRate    float64       // Minimum acceptable success rate
}

type HealthAlert struct {
    AgentName   string
    AlertType   string
    Message     string
    Severity    string
    Timestamp   time.Time
    Metrics     AgentMetrics
}

func NewProductionHealthMonitor() *ProductionHealthMonitor {
    return &ProductionHealthMonitor{
        agents: make(map[string]*ProductionAgent),
        alertThresholds: HealthThresholds{
            MaxErrorRate:    0.05, // 5% error rate
            MaxResponseTime: 5 * time.Second,
            MinSuccessRate:  0.95, // 95% success rate
        },
        alertChannel: make(chan HealthAlert, 100),
    }
}

func (phm *ProductionHealthMonitor) RegisterAgent(agent *ProductionAgent) {
    phm.mu.Lock()
    defer phm.mu.Unlock()
    
    phm.agents[agent.agent.Name()] = agent
}

func (phm *ProductionHealthMonitor) StartMonitoring(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            phm.checkAgentHealth()
        }
    }
}

func (phm *ProductionHealthMonitor) checkAgentHealth() {
    phm.mu.RLock()
    defer phm.mu.RUnlock()
    
    for agentName, agent := range phm.agents {
        metrics := agent.GetMetrics()
        
        // Calculate error rate
        errorRate := float64(0)
        if metrics.ExecutionCount > 0 {
            errorRate = float64(metrics.ErrorCount) / float64(metrics.ExecutionCount)
        }
        
        // Calculate average response time
        avgResponseTime := time.Duration(0)
        if metrics.ExecutionCount > 0 {
            avgResponseTime = metrics.TotalDuration / time.Duration(metrics.ExecutionCount)
        }
        
        // Check thresholds and generate alerts
        if errorRate > phm.alertThresholds.MaxErrorRate {
            alert := HealthAlert{
                AgentName: agentName,
                AlertType: "HIGH_ERROR_RATE",
                Message:   fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", errorRate*100, phm.alertThresholds.MaxErrorRate*100),
                Severity:  "CRITICAL",
                Timestamp: time.Now(),
                Metrics:   metrics,
            }
            
            select {
            case phm.alertChannel <- alert:
            default:
                // Alert channel full, log the issue
                core.Logger().Warn().
                    Str("agent", agentName).
                    Msg("Alert channel full, dropping health alert")
            }
        }
        
        if avgResponseTime > phm.alertThresholds.MaxResponseTime {
            alert := HealthAlert{
                AgentName: agentName,
                AlertType: "SLOW_RESPONSE",
                Message:   fmt.Sprintf("Average response time %v exceeds threshold %v", avgResponseTime, phm.alertThresholds.MaxResponseTime),
                Severity:  "WARNING",
                Timestamp: time.Now(),
                Metrics:   metrics,
            }
            
            select {
            case phm.alertChannel <- alert:
            default:
            }
        }
    }
}

func (phm *ProductionHealthMonitor) GetAlerts() <-chan HealthAlert {
    return phm.alertChannel
}
```

## Conclusion

Effective debugging of multi-agent systems requires a systematic approach that combines proper tooling, structured logging, comprehensive monitoring, and proactive error handling. By following the patterns and practices outlined in this guide, you can build robust, debuggable multi-agent systems that are easier to maintain and troubleshoot in production environments.

The key principles to remember are:

1. **Always use timeouts and proper context handling**
2. **Implement comprehensive logging with correlation IDs**
3. **Monitor resource usage and performance metrics**
4. **Use circuit breakers and graceful degradation patterns**
5. **Provide rich error context for debugging**
6. **Test your debugging and monitoring systems regularly**

By applying these debugging strategies and avoiding common pitfalls, you'll be able to build more reliable and maintainable multi-agent systems with AgenticGoKit.