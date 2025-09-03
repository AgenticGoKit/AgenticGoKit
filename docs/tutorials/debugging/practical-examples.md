---
title: Practical Debugging Examples
description: Complete, runnable debugging examples for AgenticGoKit applications
sidebar: true
outline: deep
editLink: true
lastUpdated: true
prev:
  text: 'Logging and Tracing'
  link: './logging-and-tracing'
next:
  text: 'Core Concepts'
  link: '../core-concepts/'
tags:
  - debugging
  - examples
  - practical
  - use-cases
  - troubleshooting
head:
  - - meta
    - name: keywords
      content: AgenticGoKit, debugging examples, practical debugging, use cases
  - - meta
    - property: og:title
      content: Practical Debugging Examples - AgenticGoKit
  - - meta
    - property: og:description
      content: Complete, runnable debugging examples for AgenticGoKit applications
---

# Practical Debugging Examples

## Overview

This guide provides complete, runnable debugging examples that demonstrate real-world debugging scenarios in AgenticGoKit applications. Each example builds progressively from simple to complex patterns, showing practical solutions to common debugging challenges.

## Prerequisites

::: tip Required Setup
Ensure you have these components ready before running the examples.
:::

- AgenticGoKit installed and configured
- Go 1.21+ development environment
- Basic understanding of [Agent Lifecycle](../core-concepts/agent-lifecycle)
- Familiarity with [Debugging Multi-Agent Systems](./debugging-multi-agent-systems)

## Example 1: Basic Agent Health Monitoring

### Scenario
You need to monitor the health of agents in your system and detect when they become unresponsive.

### Complete Implementation
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
    // Create a simple agent for health monitoring
    agent := &HealthMonitorAgent{
        name: "health-monitor",
        role: "monitor",
    }
    
    // Create health checker
    checker := NewHealthChecker()
    
    // Run health checks
    ctx := context.Background()
    for i := 0; i < 5; i++ {
        result := checker.CheckAgent(ctx, agent)
        fmt.Printf("Health check %d: %s\n", i+1, result.Status)
        
        if !result.Healthy {
            fmt.Printf("  Issue: %s\n", result.Issue)
            fmt.Printf("  Suggestion: %s\n", result.Suggestion)
        }
        
        time.Sleep(2 * time.Second)
    }
}

type HealthMonitorAgent struct {
    name string
    role string
}

func (h *HealthMonitorAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Simulate some work
    time.Sleep(100 * time.Millisecond)
    
    result := state.Clone()
    result.Set("health_status", "healthy")
    result.Set("last_check", time.Now())
    
    return result, nil
}

// Implement Agent interface methods
func (h *HealthMonitorAgent) Name() string { return h.name }
func (h *HealthMonitorAgent) GetRole() string { return h.role }
func (h *HealthMonitorAgent) GetDescription() string { return "Health monitoring agent" }
func (h *HealthMonitorAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return core.AgentResult{}, nil
}
func (h *HealthMonitorAgent) GetCapabilities() []string { return []string{"monitoring"} }
func (h *HealthMonitorAgent) GetSystemPrompt() string { return "" }
func (h *HealthMonitorAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (h *HealthMonitorAgent) IsEnabled() bool { return true }
func (h *HealthMonitorAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (h *HealthMonitorAgent) Initialize(ctx context.Context) error { return nil }
func (h *HealthMonitorAgent) Shutdown(ctx context.Context) error { return nil }
```

```go [health_checker.go]
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/kunalkushwaha/agenticgokit/core"
)

type HealthChecker struct {
    timeout time.Duration
}

type HealthResult struct {
    AgentName  string
    Healthy    bool
    Status     string
    Issue      string
    Suggestion string
    Duration   time.Duration
}

func NewHealthChecker() *HealthChecker {
    return &HealthChecker{
        timeout: 5 * time.Second,
    }
}

func (hc *HealthChecker) CheckAgent(ctx context.Context, agent core.Agent) HealthResult {
    start := time.Now()
    
    // Create health check context with timeout
    healthCtx, cancel := context.WithTimeout(ctx, hc.timeout)
    defer cancel()
    
    // Create health check state
    healthState := core.NewState()
    healthState.Set("health_check", true)
    healthState.Set("timestamp", start)
    
    // Perform health check
    _, err := agent.Run(healthCtx, healthState)
    duration := time.Since(start)
    
    result := HealthResult{
        AgentName: agent.Name(),
        Duration:  duration,
    }
    
    if err != nil {
        result.Healthy = false
        result.Status = "UNHEALTHY"
        result.Issue = err.Error()
        result.Suggestion = "Check agent implementation and dependencies"
    } else if duration > 2*time.Second {
        result.Healthy = false
        result.Status = "SLOW"
        result.Issue = fmt.Sprintf("Response time %v exceeds threshold", duration)
        result.Suggestion = "Investigate performance bottlenecks"
    } else {
        result.Healthy = true
        result.Status = "HEALTHY"
    }
    
    return result
}
```

```toml [agentflow.toml]
[logging]
level = "info"
format = "text"

[agent_flow]
name = "health-monitoring-example"
version = "1.0.0"

[runtime]
max_concurrent_agents = 5
```

:::

### Key Learning Points

::: info Health Monitoring Patterns
- **Timeout Management**: Always use context with timeout for health checks
- **State Validation**: Create specific health check state for testing
- **Performance Thresholds**: Monitor response times and set appropriate thresholds
- **Error Classification**: Distinguish between different types of health issues
:::

### Running the Example

```bash
# Run the health monitoring example
go run main.go health_checker.go

# Expected output:
# Health check 1: HEALTHY
# Health check 2: HEALTHY
# Health check 3: HEALTHY
# Health check 4: HEALTHY
# Health check 5: HEALTHY
```

## Example 2: Performance Profiling and Bottleneck Detection

### Scenario
Your multi-agent system is running slowly, and you need to identify performance bottlenecks and optimize agent execution.

### Complete Implementation::: c
ode-group

```go [performance_example.go]
package main

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    "time"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create agents with different performance characteristics
    agents := []core.Agent{
        &FastAgent{name: "fast-agent"},
        &SlowAgent{name: "slow-agent"},
        &MemoryIntensiveAgent{name: "memory-agent"},
    }
    
    // Create performance profiler
    profiler := NewPerformanceProfiler()
    
    // Profile each agent
    ctx := context.Background()
    for _, agent := range agents {
        fmt.Printf("\n--- Profiling %s ---\n", agent.Name())
        
        // Run multiple iterations for accurate profiling
        for i := 0; i < 3; i++ {
            result := profiler.ProfileAgent(ctx, agent)
            fmt.Printf("Run %d: Duration=%v, Memory=%d KB, CPU=%.2f%%\n",
                i+1, result.Duration, result.MemoryUsed/1024, result.CPUUsage)
        }
        
        // Get performance summary
        summary := profiler.GetSummary(agent.Name())
        fmt.Printf("Summary: Avg Duration=%v, Peak Memory=%d KB\n",
            summary.AvgDuration, summary.PeakMemory/1024)
        
        // Check for performance issues
        issues := profiler.DetectIssues(agent.Name())
        if len(issues) > 0 {
            fmt.Printf("⚠️  Performance Issues:\n")
            for _, issue := range issues {
                fmt.Printf("  - %s\n", issue)
            }
        }
    }
    
    // Generate performance report
    fmt.Printf("\n--- Performance Report ---\n")
    profiler.GenerateReport()
}

type FastAgent struct {
    name string
}

func (f *FastAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Simulate fast processing
    time.Sleep(10 * time.Millisecond)
    
    result := state.Clone()
    result.Set("processed_by", f.name)
    result.Set("processing_time", "fast")
    
    return result, nil
}

func (f *FastAgent) Name() string { return f.name }
func (f *FastAgent) GetRole() string { return "processor" }
func (f *FastAgent) GetDescription() string { return "Fast processing agent" }
func (f *FastAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return core.AgentResult{}, nil
}
func (f *FastAgent) GetCapabilities() []string { return []string{"fast-processing"} }
func (f *FastAgent) GetSystemPrompt() string { return "" }
func (f *FastAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (f *FastAgent) IsEnabled() bool { return true }
func (f *FastAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (f *FastAgent) Initialize(ctx context.Context) error { return nil }
func (f *FastAgent) Shutdown(ctx context.Context) error { return nil }

type SlowAgent struct {
    name string
}

func (s *SlowAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Simulate slow processing
    time.Sleep(500 * time.Millisecond)
    
    result := state.Clone()
    result.Set("processed_by", s.name)
    result.Set("processing_time", "slow")
    
    return result, nil
}

func (s *SlowAgent) Name() string { return s.name }
func (s *SlowAgent) GetRole() string { return "processor" }
func (s *SlowAgent) GetDescription() string { return "Slow processing agent" }
func (s *SlowAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return core.AgentResult{}, nil
}
func (s *SlowAgent) GetCapabilities() []string { return []string{"slow-processing"} }
func (s *SlowAgent) GetSystemPrompt() string { return "" }
func (s *SlowAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (s *SlowAgent) IsEnabled() bool { return true }
func (s *SlowAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (s *SlowAgent) Initialize(ctx context.Context) error { return nil }
func (s *SlowAgent) Shutdown(ctx context.Context) error { return nil }

type MemoryIntensiveAgent struct {
    name string
}

func (m *MemoryIntensiveAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Simulate memory-intensive processing
    data := make([][]byte, 1000)
    for i := range data {
        data[i] = make([]byte, 1024) // 1KB each
    }
    
    time.Sleep(50 * time.Millisecond)
    
    result := state.Clone()
    result.Set("processed_by", m.name)
    result.Set("processing_time", "memory-intensive")
    result.Set("data_size", len(data))
    
    return result, nil
}

func (m *MemoryIntensiveAgent) Name() string { return m.name }
func (m *MemoryIntensiveAgent) GetRole() string { return "processor" }
func (m *MemoryIntensiveAgent) GetDescription() string { return "Memory intensive agent" }
func (m *MemoryIntensiveAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return core.AgentResult{}, nil
}
func (m *MemoryIntensiveAgent) GetCapabilities() []string { return []string{"memory-processing"} }
func (m *MemoryIntensiveAgent) GetSystemPrompt() string { return "" }
func (m *MemoryIntensiveAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (m *MemoryIntensiveAgent) IsEnabled() bool { return true }
func (m *MemoryIntensiveAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (m *MemoryIntensiveAgent) Initialize(ctx context.Context) error { return nil }
func (m *MemoryIntensiveAgent) Shutdown(ctx context.Context) error { return nil }
```

```go [profiler.go]
package main

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    "time"
    "github.com/kunalkushwaha/agenticgokit/core"
)

type PerformanceProfiler struct {
    results map[string][]ProfileResult
    mu      sync.Mutex
}

type ProfileResult struct {
    AgentName   string
    Duration    time.Duration
    MemoryUsed  uint64
    CPUUsage    float64
    Timestamp   time.Time
}

type PerformanceSummary struct {
    AgentName    string
    AvgDuration  time.Duration
    MinDuration  time.Duration
    MaxDuration  time.Duration
    PeakMemory   uint64
    AvgMemory    uint64
    RunCount     int
}

func NewPerformanceProfiler() *PerformanceProfiler {
    return &PerformanceProfiler{
        results: make(map[string][]ProfileResult),
    }
}

func (pp *PerformanceProfiler) ProfileAgent(ctx context.Context, agent core.Agent) ProfileResult {
    var m1, m2 runtime.MemStats
    
    // Get initial memory stats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Record start time
    start := time.Now()
    
    // Execute agent
    state := core.NewState()
    state.Set("profile_test", true)
    
    _, err := agent.Run(ctx, state)
    
    // Record end time and memory
    duration := time.Since(start)
    runtime.ReadMemStats(&m2)
    
    result := ProfileResult{
        AgentName:  agent.Name(),
        Duration:   duration,
        MemoryUsed: m2.Alloc - m1.Alloc,
        CPUUsage:   calculateCPUUsage(duration),
        Timestamp:  start,
    }
    
    if err != nil {
        core.Logger().Error().
            Str("agent", agent.Name()).
            Err(err).
            Msg("Agent execution failed during profiling")
    }
    
    // Store result
    pp.mu.Lock()
    pp.results[agent.Name()] = append(pp.results[agent.Name()], result)
    pp.mu.Unlock()
    
    return result
}

func (pp *PerformanceProfiler) GetSummary(agentName string) PerformanceSummary {
    pp.mu.Lock()
    results := pp.results[agentName]
    pp.mu.Unlock()
    
    if len(results) == 0 {
        return PerformanceSummary{AgentName: agentName}
    }
    
    summary := PerformanceSummary{
        AgentName:   agentName,
        MinDuration: results[0].Duration,
        MaxDuration: results[0].Duration,
        RunCount:    len(results),
    }
    
    var totalDuration time.Duration
    var totalMemory uint64
    
    for _, result := range results {
        totalDuration += result.Duration
        totalMemory += result.MemoryUsed
        
        if result.Duration < summary.MinDuration {
            summary.MinDuration = result.Duration
        }
        if result.Duration > summary.MaxDuration {
            summary.MaxDuration = result.Duration
        }
        if result.MemoryUsed > summary.PeakMemory {
            summary.PeakMemory = result.MemoryUsed
        }
    }
    
    summary.AvgDuration = totalDuration / time.Duration(len(results))
    summary.AvgMemory = totalMemory / uint64(len(results))
    
    return summary
}

func (pp *PerformanceProfiler) DetectIssues(agentName string) []string {
    summary := pp.GetSummary(agentName)
    var issues []string
    
    // Check for slow execution
    if summary.AvgDuration > 200*time.Millisecond {
        issues = append(issues, fmt.Sprintf("Slow execution: avg %v", summary.AvgDuration))
    }
    
    // Check for high memory usage
    if summary.PeakMemory > 1024*1024 { // 1MB
        issues = append(issues, fmt.Sprintf("High memory usage: peak %d KB", summary.PeakMemory/1024))
    }
    
    // Check for inconsistent performance
    if summary.MaxDuration > 2*summary.MinDuration {
        issues = append(issues, "Inconsistent performance: high variance in execution time")
    }
    
    return issues
}

func (pp *PerformanceProfiler) GenerateReport() {
    pp.mu.Lock()
    defer pp.mu.Unlock()
    
    fmt.Printf("Agent Performance Comparison:\n")
    fmt.Printf("%-20s %-15s %-15s %-15s\n", "Agent", "Avg Duration", "Peak Memory", "Issues")
    fmt.Printf("%-20s %-15s %-15s %-15s\n", "-----", "------------", "-----------", "------")
    
    for agentName := range pp.results {
        summary := pp.GetSummary(agentName)
        issues := pp.DetectIssues(agentName)
        
        issueCount := len(issues)
        issueStr := fmt.Sprintf("%d issues", issueCount)
        if issueCount == 0 {
            issueStr = "None"
        }
        
        fmt.Printf("%-20s %-15v %-15s %-15s\n",
            agentName,
            summary.AvgDuration,
            fmt.Sprintf("%d KB", summary.PeakMemory/1024),
            issueStr)
    }
}

func calculateCPUUsage(duration time.Duration) float64 {
    // Simplified CPU usage calculation
    // In a real implementation, you'd use proper CPU monitoring
    return float64(duration.Nanoseconds()) / float64(time.Second.Nanoseconds()) * 100
}
```

:::

### Key Learning Points

::: warning Performance Monitoring
- **Memory Profiling**: Use `runtime.MemStats` to track memory allocation
- **Timing Analysis**: Measure execution duration for performance baselines
- **Issue Detection**: Implement automated detection of performance problems
- **Comparative Analysis**: Compare performance across different agents
:::

### Running the Example

```bash
# Run the performance profiling example
go run performance_example.go profiler.go

# Enable memory profiling for detailed analysis
GODEBUG=gctrace=1 go run performance_example.go profiler.go
```

## Example 3: Distributed Tracing and Error Correlation

### Scenario
You have a complex multi-agent workflow where errors in one agent affect others, and you need to trace the error propagation through the system.

### Complete Implementation::
: code-group

```go [tracing_example.go]
package main

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "time"
    "github.com/google/uuid"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create a workflow with multiple agents
    workflow := NewWorkflow()
    
    // Add agents to the workflow
    workflow.AddAgent(&DataProcessorAgent{name: "processor-1"})
    workflow.AddAgent(&DataValidatorAgent{name: "validator-1"})
    workflow.AddAgent(&DataStorageAgent{name: "storage-1"})
    
    // Create tracer
    tracer := NewDistributedTracer()
    workflow.SetTracer(tracer)
    
    // Run multiple workflow executions
    ctx := context.Background()
    for i := 0; i < 5; i++ {
        sessionID := uuid.New().String()
        fmt.Printf("\n--- Executing Workflow Session %s ---\n", sessionID[:8])
        
        // Create initial state with correlation ID
        state := core.NewState()
        state.Set("data", fmt.Sprintf("input-data-%d", i))
        state.SetMeta("session_id", sessionID)
        state.SetMeta("correlation_id", uuid.New().String())
        
        // Execute workflow
        result, err := workflow.Execute(ctx, state)
        
        if err != nil {
            fmt.Printf("❌ Workflow failed: %v\n", err)
            
            // Analyze error trace
            errorTrace := tracer.GetErrorTrace(sessionID)
            if errorTrace != nil {
                fmt.Printf("🔍 Error Analysis:\n")
                fmt.Printf("  Root Cause: %s\n", errorTrace.RootCause)
                fmt.Printf("  Error Chain: %s\n", errorTrace.ErrorChain)
                fmt.Printf("  Affected Agents: %v\n", errorTrace.AffectedAgents)
            }
        } else {
            fmt.Printf("✅ Workflow completed successfully\n")
            fmt.Printf("  Final State Keys: %v\n", result.Keys())
        }
        
        // Print execution trace
        trace := tracer.GetExecutionTrace(sessionID)
        fmt.Printf("📊 Execution Trace:\n")
        for _, step := range trace.Steps {
            status := "✅"
            if step.Error != nil {
                status = "❌"
            }
            fmt.Printf("  %s %s -> %s (%v)\n", 
                status, step.AgentName, step.Action, step.Duration)
        }
        
        time.Sleep(1 * time.Second)
    }
    
    // Generate trace analysis report
    fmt.Printf("\n--- Trace Analysis Report ---\n")
    tracer.GenerateReport()
}

type Workflow struct {
    agents []core.Agent
    tracer *DistributedTracer
}

func NewWorkflow() *Workflow {
    return &Workflow{
        agents: make([]core.Agent, 0),
    }
}

func (w *Workflow) AddAgent(agent core.Agent) {
    w.agents = append(w.agents, agent)
}

func (w *Workflow) SetTracer(tracer *DistributedTracer) {
    w.tracer = tracer
}

func (w *Workflow) Execute(ctx context.Context, initialState core.State) (core.State, error) {
    sessionID, _ := initialState.GetMeta("session_id")
    
    // Start workflow trace
    w.tracer.StartWorkflow(sessionID, len(w.agents))
    
    currentState := initialState
    
    // Execute agents in sequence
    for _, agent := range w.agents {
        // Start agent trace
        w.tracer.StartAgent(sessionID, agent.Name())
        
        start := time.Now()
        result, err := agent.Run(ctx, currentState)
        duration := time.Since(start)
        
        // Record agent execution
        w.tracer.RecordAgentExecution(sessionID, agent.Name(), duration, err)
        
        if err != nil {
            // Record error and stop workflow
            w.tracer.RecordError(sessionID, agent.Name(), err)
            return currentState, fmt.Errorf("workflow failed at agent %s: %w", agent.Name(), err)
        }
        
        currentState = result
    }
    
    // Complete workflow trace
    w.tracer.CompleteWorkflow(sessionID)
    
    return currentState, nil
}

// Agent implementations with different error patterns
type DataProcessorAgent struct {
    name string
}

func (d *DataProcessorAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Simulate processing with occasional errors
    if rand.Float32() < 0.2 { // 20% error rate
        return state, fmt.Errorf("data processing failed: invalid input format")
    }
    
    time.Sleep(100 * time.Millisecond)
    
    result := state.Clone()
    result.Set("processed", true)
    result.Set("processor", d.name)
    
    return result, nil
}

func (d *DataProcessorAgent) Name() string { return d.name }
func (d *DataProcessorAgent) GetRole() string { return "processor" }
func (d *DataProcessorAgent) GetDescription() string { return "Data processing agent" }
func (d *DataProcessorAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return core.AgentResult{}, nil
}
func (d *DataProcessorAgent) GetCapabilities() []string { return []string{"processing"} }
func (d *DataProcessorAgent) GetSystemPrompt() string { return "" }
func (d *DataProcessorAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (d *DataProcessorAgent) IsEnabled() bool { return true }
func (d *DataProcessorAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (d *DataProcessorAgent) Initialize(ctx context.Context) error { return nil }
func (d *DataProcessorAgent) Shutdown(ctx context.Context) error { return nil }

type DataValidatorAgent struct {
    name string
}

func (d *DataValidatorAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Check if data was processed
    processed, exists := state.Get("processed")
    if !exists || processed != true {
        return state, fmt.Errorf("validation failed: data not processed")
    }
    
    // Simulate validation with occasional errors
    if rand.Float32() < 0.15 { // 15% error rate
        return state, fmt.Errorf("validation failed: data integrity check failed")
    }
    
    time.Sleep(50 * time.Millisecond)
    
    result := state.Clone()
    result.Set("validated", true)
    result.Set("validator", d.name)
    
    return result, nil
}

func (d *DataValidatorAgent) Name() string { return d.name }
func (d *DataValidatorAgent) GetRole() string { return "validator" }
func (d *DataValidatorAgent) GetDescription() string { return "Data validation agent" }
func (d *DataValidatorAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return core.AgentResult{}, nil
}
func (d *DataValidatorAgent) GetCapabilities() []string { return []string{"validation"} }
func (d *DataValidatorAgent) GetSystemPrompt() string { return "" }
func (d *DataValidatorAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (d *DataValidatorAgent) IsEnabled() bool { return true }
func (d *DataValidatorAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (d *DataValidatorAgent) Initialize(ctx context.Context) error { return nil }
func (d *DataValidatorAgent) Shutdown(ctx context.Context) error { return nil }

type DataStorageAgent struct {
    name string
}

func (d *DataStorageAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Check if data was validated
    validated, exists := state.Get("validated")
    if !exists || validated != true {
        return state, fmt.Errorf("storage failed: data not validated")
    }
    
    // Simulate storage with occasional errors
    if rand.Float32() < 0.1 { // 10% error rate
        return state, fmt.Errorf("storage failed: database connection error")
    }
    
    time.Sleep(75 * time.Millisecond)
    
    result := state.Clone()
    result.Set("stored", true)
    result.Set("storage", d.name)
    result.Set("storage_id", uuid.New().String())
    
    return result, nil
}

func (d *DataStorageAgent) Name() string { return d.name }
func (d *DataStorageAgent) GetRole() string { return "storage" }
func (d *DataStorageAgent) GetDescription() string { return "Data storage agent" }
func (d *DataStorageAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return core.AgentResult{}, nil
}
func (d *DataStorageAgent) GetCapabilities() []string { return []string{"storage"} }
func (d *DataStorageAgent) GetSystemPrompt() string { return "" }
func (d *DataStorageAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (d *DataStorageAgent) IsEnabled() bool { return true }
func (d *DataStorageAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (d *DataStorageAgent) Initialize(ctx context.Context) error { return nil }
func (d *DataStorageAgent) Shutdown(ctx context.Context) error { return nil }
```

```go [tracer.go]
package main

import (
    "fmt"
    "sync"
    "time"
)

type DistributedTracer struct {
    workflows map[string]*WorkflowTrace
    mu        sync.Mutex
}

type WorkflowTrace struct {
    SessionID     string
    StartTime     time.Time
    EndTime       time.Time
    TotalAgents   int
    Steps         []ExecutionStep
    Errors        []ErrorRecord
    Status        string
}

type ExecutionStep struct {
    AgentName string
    Action    string
    StartTime time.Time
    Duration  time.Duration
    Error     error
}

type ErrorRecord struct {
    AgentName   string
    Error       error
    Timestamp   time.Time
    Context     map[string]interface{}
}

type ErrorTrace struct {
    SessionID       string
    RootCause       string
    ErrorChain      string
    AffectedAgents  []string
    FailurePoint    string
}

func NewDistributedTracer() *DistributedTracer {
    return &DistributedTracer{
        workflows: make(map[string]*WorkflowTrace),
    }
}

func (dt *DistributedTracer) StartWorkflow(sessionID string, totalAgents int) {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    dt.workflows[sessionID] = &WorkflowTrace{
        SessionID:   sessionID,
        StartTime:   time.Now(),
        TotalAgents: totalAgents,
        Steps:       make([]ExecutionStep, 0),
        Errors:      make([]ErrorRecord, 0),
        Status:      "running",
    }
}

func (dt *DistributedTracer) StartAgent(sessionID, agentName string) {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    if workflow, exists := dt.workflows[sessionID]; exists {
        step := ExecutionStep{
            AgentName: agentName,
            Action:    "executing",
            StartTime: time.Now(),
        }
        workflow.Steps = append(workflow.Steps, step)
    }
}

func (dt *DistributedTracer) RecordAgentExecution(sessionID, agentName string, duration time.Duration, err error) {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    if workflow, exists := dt.workflows[sessionID]; exists {
        // Update the last step
        if len(workflow.Steps) > 0 {
            lastStep := &workflow.Steps[len(workflow.Steps)-1]
            if lastStep.AgentName == agentName {
                lastStep.Duration = duration
                lastStep.Error = err
                if err != nil {
                    lastStep.Action = "failed"
                } else {
                    lastStep.Action = "completed"
                }
            }
        }
    }
}

func (dt *DistributedTracer) RecordError(sessionID, agentName string, err error) {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    if workflow, exists := dt.workflows[sessionID]; exists {
        errorRecord := ErrorRecord{
            AgentName: agentName,
            Error:     err,
            Timestamp: time.Now(),
            Context: map[string]interface{}{
                "session_id": sessionID,
                "step_count": len(workflow.Steps),
            },
        }
        workflow.Errors = append(workflow.Errors, errorRecord)
        workflow.Status = "failed"
    }
}

func (dt *DistributedTracer) CompleteWorkflow(sessionID string) {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    if workflow, exists := dt.workflows[sessionID]; exists {
        workflow.EndTime = time.Now()
        if workflow.Status != "failed" {
            workflow.Status = "completed"
        }
    }
}

func (dt *DistributedTracer) GetExecutionTrace(sessionID string) *WorkflowTrace {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    if workflow, exists := dt.workflows[sessionID]; exists {
        return workflow
    }
    return nil
}

func (dt *DistributedTracer) GetErrorTrace(sessionID string) *ErrorTrace {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    workflow, exists := dt.workflows[sessionID]
    if !exists || len(workflow.Errors) == 0 {
        return nil
    }
    
    // Analyze error chain
    var affectedAgents []string
    var errorChain string
    var rootCause string
    var failurePoint string
    
    for i, errorRecord := range workflow.Errors {
        affectedAgents = append(affectedAgents, errorRecord.AgentName)
        
        if i == 0 {
            rootCause = errorRecord.Error.Error()
            failurePoint = errorRecord.AgentName
        }
        
        if errorChain != "" {
            errorChain += " -> "
        }
        errorChain += fmt.Sprintf("%s: %s", errorRecord.AgentName, errorRecord.Error.Error())
    }
    
    return &ErrorTrace{
        SessionID:      sessionID,
        RootCause:      rootCause,
        ErrorChain:     errorChain,
        AffectedAgents: affectedAgents,
        FailurePoint:   failurePoint,
    }
}

func (dt *DistributedTracer) GenerateReport() {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    totalWorkflows := len(dt.workflows)
    successfulWorkflows := 0
    failedWorkflows := 0
    
    var totalDuration time.Duration
    errorsByAgent := make(map[string]int)
    
    for _, workflow := range dt.workflows {
        if workflow.Status == "completed" {
            successfulWorkflows++
        } else if workflow.Status == "failed" {
            failedWorkflows++
        }
        
        if !workflow.EndTime.IsZero() {
            totalDuration += workflow.EndTime.Sub(workflow.StartTime)
        }
        
        // Count errors by agent
        for _, errorRecord := range workflow.Errors {
            errorsByAgent[errorRecord.AgentName]++
        }
    }
    
    fmt.Printf("Workflow Execution Summary:\n")
    fmt.Printf("  Total Workflows: %d\n", totalWorkflows)
    fmt.Printf("  Successful: %d (%.1f%%)\n", successfulWorkflows, 
        float64(successfulWorkflows)/float64(totalWorkflows)*100)
    fmt.Printf("  Failed: %d (%.1f%%)\n", failedWorkflows,
        float64(failedWorkflows)/float64(totalWorkflows)*100)
    
    if totalWorkflows > 0 {
        avgDuration := totalDuration / time.Duration(totalWorkflows)
        fmt.Printf("  Average Duration: %v\n", avgDuration)
    }
    
    if len(errorsByAgent) > 0 {
        fmt.Printf("\nError Distribution by Agent:\n")
        for agentName, errorCount := range errorsByAgent {
            fmt.Printf("  %s: %d errors\n", agentName, errorCount)
        }
    }
}
```

:::

### Key Learning Points

::: tip Distributed Tracing Patterns
- **Correlation IDs**: Use unique IDs to track requests across agents
- **Error Propagation**: Understand how errors cascade through agent workflows
- **Trace Analysis**: Implement automated analysis of execution patterns
- **Root Cause Analysis**: Identify the original source of failures
:::

### Running the Example

```bash
# Run the distributed tracing example
go run tracing_example.go tracer.go

# Expected output shows workflow executions with success/failure patterns
# and detailed error analysis when failures occur
```

## Example 4: Production Debugging with Circuit Breaker

### Scenario
You need to implement robust debugging for a production system that can handle agent failures gracefully while maintaining system stability.

### Complete Implementation::: code
-group

```go [production_example.go]
package main

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "sync"
    "time"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create production system with circuit breaker
    system := NewProductionSystem()
    
    // Add agents with different reliability characteristics
    system.AddAgent(&ReliableAgent{name: "reliable-agent"})
    system.AddAgent(&UnreliableAgent{name: "unreliable-agent"})
    system.AddAgent(&IntermittentAgent{name: "intermittent-agent"})
    
    // Start monitoring
    monitor := system.StartMonitoring()
    
    // Simulate production load
    ctx := context.Background()
    for i := 0; i < 20; i++ {
        fmt.Printf("\n--- Request %d ---\n", i+1)
        
        // Create request state
        state := core.NewState()
        state.Set("request_id", fmt.Sprintf("req-%d", i+1))
        state.Set("timestamp", time.Now())
        
        // Process request through all agents
        for _, agentName := range []string{"reliable-agent", "unreliable-agent", "intermittent-agent"} {
            result := system.ProcessWithAgent(ctx, agentName, state)
            
            fmt.Printf("  %s: %s", agentName, result.Status)
            if result.Error != nil {
                fmt.Printf(" (Error: %s)", result.Error.Error())
            }
            if result.CircuitOpen {
                fmt.Printf(" [CIRCUIT OPEN]")
            }
            fmt.Printf("\n")
            
            // Update state for next agent if successful
            if result.State != nil {
                state = result.State
            }
        }
        
        time.Sleep(500 * time.Millisecond)
    }
    
    // Stop monitoring and generate report
    monitor.Stop()
    system.GenerateHealthReport()
}

type ProductionSystem struct {
    agents          map[string]core.Agent
    circuitBreakers map[string]*CircuitBreaker
    monitor         *SystemMonitor
    mu              sync.RWMutex
}

type ProcessResult struct {
    Status       string
    State        core.State
    Error        error
    Duration     time.Duration
    CircuitOpen  bool
}

func NewProductionSystem() *ProductionSystem {
    return &ProductionSystem{
        agents:          make(map[string]core.Agent),
        circuitBreakers: make(map[string]*CircuitBreaker),
    }
}

func (ps *ProductionSystem) AddAgent(agent core.Agent) {
    ps.mu.Lock()
    defer ps.mu.Unlock()
    
    ps.agents[agent.Name()] = agent
    ps.circuitBreakers[agent.Name()] = NewCircuitBreaker(agent.Name())
}

func (ps *ProductionSystem) StartMonitoring() *SystemMonitor {
    ps.monitor = NewSystemMonitor(ps)
    ps.monitor.Start()
    return ps.monitor
}

func (ps *ProductionSystem) ProcessWithAgent(ctx context.Context, agentName string, state core.State) ProcessResult {
    ps.mu.RLock()
    agent, exists := ps.agents[agentName]
    circuitBreaker, cbExists := ps.circuitBreakers[agentName]
    ps.mu.RUnlock()
    
    if !exists || !cbExists {
        return ProcessResult{
            Status: "NOT_FOUND",
            Error:  fmt.Errorf("agent %s not found", agentName),
        }
    }
    
    // Check circuit breaker
    if !circuitBreaker.CanExecute() {
        return ProcessResult{
            Status:      "CIRCUIT_OPEN",
            CircuitOpen: true,
            Error:       fmt.Errorf("circuit breaker open for agent %s", agentName),
        }
    }
    
    // Execute agent with timeout and monitoring
    start := time.Now()
    
    // Create execution context with timeout
    execCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    result, err := agent.Run(execCtx, state)
    duration := time.Since(start)
    
    // Record execution result in circuit breaker
    if err != nil {
        circuitBreaker.RecordFailure()
        return ProcessResult{
            Status:   "FAILED",
            Error:    err,
            Duration: duration,
        }
    } else {
        circuitBreaker.RecordSuccess()
        return ProcessResult{
            Status:   "SUCCESS",
            State:    result,
            Duration: duration,
        }
    }
}

func (ps *ProductionSystem) GenerateHealthReport() {
    ps.mu.RLock()
    defer ps.mu.RUnlock()
    
    fmt.Printf("\n--- System Health Report ---\n")
    
    for agentName, cb := range ps.circuitBreakers {
        stats := cb.GetStats()
        fmt.Printf("Agent: %s\n", agentName)
        fmt.Printf("  State: %s\n", cb.GetState())
        fmt.Printf("  Success Rate: %.1f%%\n", stats.SuccessRate*100)
        fmt.Printf("  Total Requests: %d\n", stats.TotalRequests)
        fmt.Printf("  Failed Requests: %d\n", stats.FailedRequests)
        fmt.Printf("  Circuit Opens: %d\n", stats.CircuitOpens)
        fmt.Printf("\n")
    }
}

// Agent implementations with different reliability patterns
type ReliableAgent struct {
    name string
}

func (r *ReliableAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Simulate reliable processing (95% success rate)
    if rand.Float32() < 0.05 {
        return state, fmt.Errorf("rare failure in reliable agent")
    }
    
    time.Sleep(50 * time.Millisecond)
    
    result := state.Clone()
    result.Set("processed_by_reliable", true)
    
    return result, nil
}

func (r *ReliableAgent) Name() string { return r.name }
func (r *ReliableAgent) GetRole() string { return "processor" }
func (r *ReliableAgent) GetDescription() string { return "Reliable processing agent" }
func (r *ReliableAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return core.AgentResult{}, nil
}
func (r *ReliableAgent) GetCapabilities() []string { return []string{"reliable-processing"} }
func (r *ReliableAgent) GetSystemPrompt() string { return "" }
func (r *ReliableAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (r *ReliableAgent) IsEnabled() bool { return true }
func (r *ReliableAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (r *ReliableAgent) Initialize(ctx context.Context) error { return nil }
func (r *ReliableAgent) Shutdown(ctx context.Context) error { return nil }

type UnreliableAgent struct {
    name string
}

func (u *UnreliableAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    // Simulate unreliable processing (60% success rate)
    if rand.Float32() < 0.4 {
        return state, fmt.Errorf("frequent failure in unreliable agent")
    }
    
    time.Sleep(100 * time.Millisecond)
    
    result := state.Clone()
    result.Set("processed_by_unreliable", true)
    
    return result, nil
}

func (u *UnreliableAgent) Name() string { return u.name }
func (u *UnreliableAgent) GetRole() string { return "processor" }
func (u *UnreliableAgent) GetDescription() string { return "Unreliable processing agent" }
func (u *UnreliableAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return core.AgentResult{}, nil
}
func (u *UnreliableAgent) GetCapabilities() []string { return []string{"unreliable-processing"} }
func (u *UnreliableAgent) GetSystemPrompt() string { return "" }
func (u *UnreliableAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (u *UnreliableAgent) IsEnabled() bool { return true }
func (u *UnreliableAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (u *UnreliableAgent) Initialize(ctx context.Context) error { return nil }
func (u *UnreliableAgent) Shutdown(ctx context.Context) error { return nil }

type IntermittentAgent struct {
    name        string
    failureMode bool
    counter     int
}

func (i *IntermittentAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    i.counter++
    
    // Simulate intermittent failures (fail for 3 requests, then succeed for 5)
    if i.counter%8 < 3 {
        return state, fmt.Errorf("intermittent failure mode active")
    }
    
    time.Sleep(75 * time.Millisecond)
    
    result := state.Clone()
    result.Set("processed_by_intermittent", true)
    
    return result, nil
}

func (i *IntermittentAgent) Name() string { return i.name }
func (i *IntermittentAgent) GetRole() string { return "processor" }
func (i *IntermittentAgent) GetDescription() string { return "Intermittent processing agent" }
func (i *IntermittentAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    return core.AgentResult{}, nil
}
func (i *IntermittentAgent) GetCapabilities() []string { return []string{"intermittent-processing"} }
func (i *IntermittentAgent) GetSystemPrompt() string { return "" }
func (i *IntermittentAgent) GetTimeout() time.Duration { return 30 * time.Second }
func (i *IntermittentAgent) IsEnabled() bool { return true }
func (i *IntermittentAgent) GetLLMConfig() *core.ResolvedLLMConfig { return nil }
func (i *IntermittentAgent) Initialize(ctx context.Context) error { return nil }
func (i *IntermittentAgent) Shutdown(ctx context.Context) error { return nil }
```

```go [circuit_breaker.go]
package main

import (
    "sync"
    "time"
)

type CircuitBreaker struct {
    name            string
    state           CircuitState
    failureCount    int
    successCount    int
    totalRequests   int
    failedRequests  int
    circuitOpens    int
    lastFailureTime time.Time
    mu              sync.Mutex
    
    // Configuration
    failureThreshold int
    recoveryTimeout  time.Duration
    successThreshold int
}

type CircuitState int

const (
    CircuitClosed CircuitState = iota
    CircuitOpen
    CircuitHalfOpen
)

func (cs CircuitState) String() string {
    switch cs {
    case CircuitClosed:
        return "CLOSED"
    case CircuitOpen:
        return "OPEN"
    case CircuitHalfOpen:
        return "HALF_OPEN"
    default:
        return "UNKNOWN"
    }
}

type CircuitStats struct {
    State           string
    SuccessRate     float64
    TotalRequests   int
    FailedRequests  int
    CircuitOpens    int
}

func NewCircuitBreaker(name string) *CircuitBreaker {
    return &CircuitBreaker{
        name:             name,
        state:            CircuitClosed,
        failureThreshold: 3,  // Open after 3 consecutive failures
        recoveryTimeout:  5 * time.Second,
        successThreshold: 2,  // Close after 2 consecutive successes in half-open
    }
}

func (cb *CircuitBreaker) CanExecute() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    switch cb.state {
    case CircuitClosed:
        return true
    case CircuitOpen:
        // Check if recovery timeout has passed
        if time.Since(cb.lastFailureTime) > cb.recoveryTimeout {
            cb.state = CircuitHalfOpen
            cb.successCount = 0
            return true
        }
        return false
    case CircuitHalfOpen:
        return true
    default:
        return false
    }
}

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.totalRequests++
    cb.failureCount = 0
    
    switch cb.state {
    case CircuitClosed:
        // Stay closed
    case CircuitHalfOpen:
        cb.successCount++
        if cb.successCount >= cb.successThreshold {
            cb.state = CircuitClosed
        }
    case CircuitOpen:
        // This shouldn't happen, but reset to half-open
        cb.state = CircuitHalfOpen
        cb.successCount = 1
    }
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.totalRequests++
    cb.failedRequests++
    cb.failureCount++
    cb.successCount = 0
    cb.lastFailureTime = time.Now()
    
    switch cb.state {
    case CircuitClosed:
        if cb.failureCount >= cb.failureThreshold {
            cb.state = CircuitOpen
            cb.circuitOpens++
        }
    case CircuitHalfOpen:
        cb.state = CircuitOpen
        cb.circuitOpens++
    case CircuitOpen:
        // Already open, just update failure time
    }
}

func (cb *CircuitBreaker) GetState() string {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    return cb.state.String()
}

func (cb *CircuitBreaker) GetStats() CircuitStats {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    successRate := 0.0
    if cb.totalRequests > 0 {
        successRate = float64(cb.totalRequests-cb.failedRequests) / float64(cb.totalRequests)
    }
    
    return CircuitStats{
        State:          cb.state.String(),
        SuccessRate:    successRate,
        TotalRequests:  cb.totalRequests,
        FailedRequests: cb.failedRequests,
        CircuitOpens:   cb.circuitOpens,
    }
}

type SystemMonitor struct {
    system   *ProductionSystem
    stopChan chan struct{}
    running  bool
}

func NewSystemMonitor(system *ProductionSystem) *SystemMonitor {
    return &SystemMonitor{
        system:   system,
        stopChan: make(chan struct{}),
    }
}

func (sm *SystemMonitor) Start() {
    if sm.running {
        return
    }
    
    sm.running = true
    go sm.monitorLoop()
}

func (sm *SystemMonitor) Stop() {
    if !sm.running {
        return
    }
    
    close(sm.stopChan)
    sm.running = false
}

func (sm *SystemMonitor) monitorLoop() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-sm.stopChan:
            return
        case <-ticker.C:
            sm.logSystemHealth()
        }
    }
}

func (sm *SystemMonitor) logSystemHealth() {
    sm.system.mu.RLock()
    defer sm.system.mu.RUnlock()
    
    fmt.Printf("\n[MONITOR] System Health Check:\n")
    
    for agentName, cb := range sm.system.circuitBreakers {
        stats := cb.GetStats()
        status := "🟢"
        if stats.State == "OPEN" {
            status = "🔴"
        } else if stats.State == "HALF_OPEN" {
            status = "🟡"
        }
        
        fmt.Printf("  %s %s: %s (Success: %.1f%%, Requests: %d)\n",
            status, agentName, stats.State, stats.SuccessRate*100, stats.TotalRequests)
    }
}
```

:::

### Key Learning Points

::: warning Production Debugging Patterns
- **Circuit Breaker Pattern**: Prevent cascade failures by isolating failing components
- **Health Monitoring**: Continuously monitor system health and agent performance
- **Graceful Degradation**: Handle failures without bringing down the entire system
- **Recovery Mechanisms**: Implement automatic recovery when agents become healthy again
:::

### Running the Example

```bash
# Run the production debugging example
go run production_example.go circuit_breaker.go

# The output shows how the circuit breaker protects the system
# from unreliable agents while allowing recovery
```

## Common Debugging Pitfalls and Solutions

### Pitfall 1: Not Using Timeouts

::: danger Common Mistake
```go
// ❌ Bad: No timeout
result, err := agent.Run(context.Background(), state)
```

```go
// ✅ Good: With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
result, err := agent.Run(ctx, state)
```
:::

### Pitfall 2: Ignoring Error Context

::: danger Common Mistake
```go
// ❌ Bad: Generic error
return fmt.Errorf("agent failed")
```

```go
// ✅ Good: Contextual error
return fmt.Errorf("agent %s failed during %s: %w", agentName, operation, err)
```
:::

### Pitfall 3: Not Monitoring Resource Usage

::: danger Common Mistake
```go
// ❌ Bad: No resource monitoring
for {
    agent.Run(ctx, state)
}
```

```go
// ✅ Good: With resource monitoring
for {
    var m1, m2 runtime.MemStats
    runtime.ReadMemStats(&m1)
    
    agent.Run(ctx, state)
    
    runtime.ReadMemStats(&m2)
    if m2.Alloc-m1.Alloc > threshold {
        log.Printf("High memory usage detected: %d bytes", m2.Alloc-m1.Alloc)
    }
}
```
:::

## Performance Optimization Guidelines

### 1. Agent Execution Optimization

- **Use connection pooling** for external services
- **Implement caching** for frequently accessed data
- **Batch operations** when possible
- **Use goroutines** for concurrent processing

### 2. Memory Management

- **Monitor memory allocation** in agent execution
- **Implement proper cleanup** in agent shutdown
- **Use object pooling** for frequently created objects
- **Profile memory usage** regularly

### 3. Debugging Overhead Minimization

- **Use appropriate log levels** in production
- **Implement sampling** for high-volume tracing
- **Cache debugging metadata** to reduce computation
- **Use async logging** for performance-critical paths

## Next Steps

::: info Continue Learning
Explore these related topics to deepen your debugging expertise.
:::

- **[Logging and Tracing](./logging-and-tracing)** - Advanced logging patterns
- **[Debugging Multi-Agent Systems](./debugging-multi-agent-systems)** - Core debugging techniques
- **[Agent Lifecycle](../core-concepts/agent-lifecycle)** - Understanding agent execution
- **[Error Handling](../core-concepts/error-handling)** - Robust error management patterns

## Conclusion

These practical examples demonstrate real-world debugging scenarios and provide complete, runnable code that you can adapt to your specific use cases. Each example builds on the previous ones, showing progressively more sophisticated debugging patterns and techniques for multi-agent systems.

The key to effective debugging is to implement comprehensive monitoring, use appropriate error handling patterns, and build resilient systems that can handle failures gracefully while providing clear visibility into system behavior.