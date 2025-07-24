# Agent Lifecycle in AgenticGoKit

## Overview

Understanding the agent lifecycle is fundamental to building effective multi-agent systems. This tutorial explores how agents are created, initialized, executed, and cleaned up in AgenticGoKit, along with best practices for managing agent resources and state.

The agent lifecycle encompasses everything from agent creation and configuration to execution patterns and resource cleanup.

## Prerequisites

- Basic understanding of Go programming
- Familiarity with [Message Passing and Event Flow](message-passing.md)
- Knowledge of [State Management](state-management.md)

## Agent Lifecycle Phases

### 1. Creation and Configuration

Agents go through several phases during their lifecycle:

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌─────────────┐
│   Creation  │───▶│Configuration │───▶│Initialization│───▶│   Ready     │
└─────────────┘    └──────────────┘    └─────────────┘    └─────────────┘
                                                                  │
┌─────────────┐    ┌──────────────┐    ┌─────────────┐           │
│   Cleanup   │◀───│   Shutdown   │◀───│  Execution  │◀──────────┘
└─────────────┘    └──────────────┘    └─────────────┘
```

## Agent Creation Patterns

### 1. Builder Pattern

AgenticGoKit uses the builder pattern for agent creation:

```go
// Create an agent using the builder pattern
agent, err := core.NewAgent("my-agent").
    WithLLMAndConfig(provider, core.LLMConfig{
        SystemPrompt: "You are a helpful assistant.",
        Temperature:  0.7,
        MaxTokens:    1000,
    }).
    WithMemory(memorySystem).
    WithMCP(mcpManager).
    WithMetrics().
    Build()

if err != nil {
    log.Fatalf("Failed to create agent: %v", err)
}
```

### 2. Factory Pattern

For complex agent creation scenarios:

```go
type AgentFactory struct {
    defaultConfig AgentConfig
    providers     map[string]LLMProvider
    memory        Memory
}

func (f *AgentFactory) CreateAgent(agentType string, config AgentConfig) (Agent, error) {
    switch agentType {
    case "research":
        return f.createResearchAgent(config)
    case "analysis":
        return f.createAnalysisAgent(config)
    case "writing":
        return f.createWritingAgent(config)
    default:
        return nil, fmt.Errorf("unknown agent type: %s", agentType)
    }
}

func (f *AgentFactory) createResearchAgent(config AgentConfig) (Agent, error) {
    return core.NewAgent("research-agent").
        WithLLMAndConfig(f.providers["research"], core.LLMConfig{
            SystemPrompt: "You are a research specialist...",
            Temperature:  0.3,
        }).
        WithMemory(f.memory).
        WithMCP(f.getMCPForResearch()).
        Build()
}
```

### 3. Configuration-Based Creation

Create agents from configuration files:

```go
type AgentSpec struct {
    Name     string            `yaml:"name"`
    Type     string            `yaml:"type"`
    LLM      LLMConfig         `yaml:"llm"`
    Memory   MemoryConfig      `yaml:"memory"`
    MCP      MCPConfig         `yaml:"mcp"`
    Metadata map[string]string `yaml:"metadata"`
}

func CreateAgentFromSpec(spec AgentSpec) (Agent, error) {
    builder := core.NewAgent(spec.Name)
    
    // Configure LLM
    if spec.LLM.Provider != "" {
        provider, err := createLLMProvider(spec.LLM)
        if err != nil {
            return nil, err
        }
        builder = builder.WithLLMAndConfig(provider, spec.LLM)
    }
    
    // Configure memory
    if spec.Memory.Enabled {
        memory, err := createMemorySystem(spec.Memory)
        if err != nil {
            return nil, err
        }
        builder = builder.WithMemory(memory)
    }
    
    // Configure MCP
    if len(spec.MCP.Tools) > 0 {
        mcpManager, err := createMCPManager(spec.MCP)
        if err != nil {
            return nil, err
        }
        builder = builder.WithMCP(mcpManager)
    }
    
    return builder.Build()
}
```

## Agent Initialization

### 1. Initialization Hooks

Agents can implement initialization logic:

```go
type InitializableAgent interface {
    Agent
    Initialize(ctx context.Context) error
    IsInitialized() bool
}

type MyAgent struct {
    name        string
    llm         LLMProvider
    initialized bool
    resources   []Resource
}

func (a *MyAgent) Initialize(ctx context.Context) error {
    if a.initialized {
        return nil
    }
    
    // Initialize resources
    for _, resource := range a.resources {
        if err := resource.Initialize(ctx); err != nil {
            return fmt.Errorf("failed to initialize resource: %w", err)
        }
    }
    
    // Perform any setup tasks
    if err := a.setupTasks(ctx); err != nil {
        return fmt.Errorf("setup tasks failed: %w", err)
    }
    
    a.initialized = true
    return nil
}

func (a *MyAgent) IsInitialized() bool {
    return a.initialized
}
```

### 2. Lazy Initialization

Initialize resources only when needed:

```go
type LazyAgent struct {
    name         string
    config       AgentConfig
    llm          LLMProvider
    initOnce     sync.Once
    initError    error
}

func (a *LazyAgent) ensureInitialized(ctx context.Context) error {
    a.initOnce.Do(func() {
        a.initError = a.initialize(ctx)
    })
    return a.initError
}

func (a *LazyAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    if err := a.ensureInitialized(ctx); err != nil {
        return AgentResult{}, fmt.Errorf("initialization failed: %w", err)
    }
    
    // Normal execution
    return a.execute(ctx, event, state)
}
```

## Agent Execution Lifecycle

### 1. Pre-Execution Phase

Before each agent execution:

```go
func (a *MyAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Pre-execution setup
    executionID := generateExecutionID()
    startTime := time.Now()
    
    // Log execution start
    a.logger.Info("Agent execution started", 
        "agent", a.name,
        "execution_id", executionID,
        "event_id", event.GetID())
    
    // Validate inputs
    if err := a.validateInputs(event, state); err != nil {
        return AgentResult{}, fmt.Errorf("input validation failed: %w", err)
    }
    
    // Setup execution context
    execCtx := a.setupExecutionContext(ctx, executionID)
    
    // Execute main logic
    result, err := a.execute(execCtx, event, state)
    
    // Post-execution cleanup
    duration := time.Since(startTime)
    a.recordMetrics(executionID, duration, err)
    
    return result, err
}
```

### 2. Execution Context Management

Manage execution-specific context:

```go
type ExecutionContext struct {
    ExecutionID string
    StartTime   time.Time
    Metadata    map[string]interface{}
    Resources   map[string]interface{}
    Cleanup     []func()
}

func (a *MyAgent) setupExecutionContext(ctx context.Context, executionID string) context.Context {
    execCtx := &ExecutionContext{
        ExecutionID: executionID,
        StartTime:   time.Now(),
        Metadata:    make(map[string]interface{}),
        Resources:   make(map[string]interface{}),
        Cleanup:     make([]func(), 0),
    }
    
    return context.WithValue(ctx, "execution_context", execCtx)
}

func (a *MyAgent) execute(ctx context.Context, event Event, state State) (AgentResult, error) {
    execCtx := ctx.Value("execution_context").(*ExecutionContext)
    defer a.cleanup(execCtx)
    
    // Main execution logic
    return a.processEvent(ctx, event, state)
}

func (a *MyAgent) cleanup(execCtx *ExecutionContext) {
    for _, cleanupFunc := range execCtx.Cleanup {
        cleanupFunc()
    }
}
```

### 3. Resource Management During Execution

Manage resources throughout execution:

```go
type ResourceManager struct {
    resources map[string]Resource
    mu        sync.RWMutex
}

func (rm *ResourceManager) AcquireResource(ctx context.Context, name string) (Resource, error) {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    resource, exists := rm.resources[name]
    if !exists {
        return nil, fmt.Errorf("resource not found: %s", name)
    }
    
    if err := resource.Acquire(ctx); err != nil {
        return nil, fmt.Errorf("failed to acquire resource %s: %w", name, err)
    }
    
    return resource, nil
}

func (rm *ResourceManager) ReleaseResource(name string) error {
    rm.mu.RLock()
    resource, exists := rm.resources[name]
    rm.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("resource not found: %s", name)
    }
    
    return resource.Release()
}
```

## Agent State Management

### 1. Agent Internal State

Manage agent's internal state across executions:

```go
type StatefulAgent struct {
    name          string
    internalState map[string]interface{}
    stateMutex    sync.RWMutex
    persistence   StatePersistence
}

func (a *StatefulAgent) GetInternalState(key string) (interface{}, bool) {
    a.stateMutex.RLock()
    defer a.stateMutex.RUnlock()
    
    value, exists := a.internalState[key]
    return value, exists
}

func (a *StatefulAgent) SetInternalState(key string, value interface{}) {
    a.stateMutex.Lock()
    defer a.stateMutex.Unlock()
    
    a.internalState[key] = value
    
    // Persist state if configured
    if a.persistence != nil {
        a.persistence.SaveState(a.name, a.internalState)
    }
}

func (a *StatefulAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Load persisted state on first run
    if len(a.internalState) == 0 && a.persistence != nil {
        if persistedState, err := a.persistence.LoadState(a.name); err == nil {
            a.internalState = persistedState
        }
    }
    
    // Use internal state in processing
    return a.processWithInternalState(ctx, event, state)
}
```

### 2. State Persistence

Persist agent state across restarts:

```go
type StatePersistence interface {
    SaveState(agentName string, state map[string]interface{}) error
    LoadState(agentName string) (map[string]interface{}, error)
    DeleteState(agentName string) error
}

type FileStatePersistence struct {
    baseDir string
}

func (fsp *FileStatePersistence) SaveState(agentName string, state map[string]interface{}) error {
    filename := filepath.Join(fsp.baseDir, agentName+".json")
    
    data, err := json.Marshal(state)
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, data, 0644)
}

func (fsp *FileStatePersistence) LoadState(agentName string) (map[string]interface{}, error) {
    filename := filepath.Join(fsp.baseDir, agentName+".json")
    
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var state map[string]interface{}
    err = json.Unmarshal(data, &state)
    return state, err
}
```

## Agent Health and Monitoring

### 1. Health Checks

Implement health monitoring for agents:

```go
type HealthStatus int

const (
    HealthStatusHealthy HealthStatus = iota
    HealthStatusDegraded
    HealthStatusUnhealthy
)

type HealthCheckable interface {
    HealthCheck(ctx context.Context) HealthStatus
    GetHealthDetails() map[string]interface{}
}

type MyAgent struct {
    name           string
    lastExecution  time.Time
    errorCount     int64
    successCount   int64
    resources      []HealthCheckable
}

func (a *MyAgent) HealthCheck(ctx context.Context) HealthStatus {
    // Check if agent has been executing recently
    if time.Since(a.lastExecution) > 5*time.Minute {
        return HealthStatusDegraded
    }
    
    // Check error rate
    totalExecutions := a.errorCount + a.successCount
    if totalExecutions > 0 {
        errorRate := float64(a.errorCount) / float64(totalExecutions)
        if errorRate > 0.5 {
            return HealthStatusUnhealthy
        } else if errorRate > 0.2 {
            return HealthStatusDegraded
        }
    }
    
    // Check resource health
    for _, resource := range a.resources {
        if resource.HealthCheck(ctx) == HealthStatusUnhealthy {
            return HealthStatusDegraded
        }
    }
    
    return HealthStatusHealthy
}

func (a *MyAgent) GetHealthDetails() map[string]interface{} {
    return map[string]interface{}{
        "name":            a.name,
        "last_execution":  a.lastExecution,
        "error_count":     a.errorCount,
        "success_count":   a.successCount,
        "error_rate":      float64(a.errorCount) / float64(a.errorCount + a.successCount),
    }
}
```

### 2. Performance Monitoring

Monitor agent performance metrics:

```go
type PerformanceMonitor struct {
    agentName       string
    executionTimes  []time.Duration
    memoryUsage     []int64
    maxHistorySize  int
    mu              sync.RWMutex
}

func (pm *PerformanceMonitor) RecordExecution(duration time.Duration, memoryUsed int64) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    pm.executionTimes = append(pm.executionTimes, duration)
    pm.memoryUsage = append(pm.memoryUsage, memoryUsed)
    
    // Keep only recent history
    if len(pm.executionTimes) > pm.maxHistorySize {
        pm.executionTimes = pm.executionTimes[1:]
        pm.memoryUsage = pm.memoryUsage[1:]
    }
}

func (pm *PerformanceMonitor) GetAverageExecutionTime() time.Duration {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    
    if len(pm.executionTimes) == 0 {
        return 0
    }
    
    var total time.Duration
    for _, duration := range pm.executionTimes {
        total += duration
    }
    
    return total / time.Duration(len(pm.executionTimes))
}
```

## Agent Cleanup and Shutdown

### 1. Graceful Shutdown

Implement graceful shutdown for agents:

```go
type GracefulAgent interface {
    Agent
    Shutdown(ctx context.Context) error
}

type MyAgent struct {
    name      string
    resources []Resource
    shutdown  chan struct{}
    wg        sync.WaitGroup
}

func (a *MyAgent) Shutdown(ctx context.Context) error {
    // Signal shutdown
    close(a.shutdown)
    
    // Wait for ongoing operations with timeout
    done := make(chan struct{})
    go func() {
        a.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        // All operations completed
    case <-ctx.Done():
        return ctx.Err()
    }
    
    // Cleanup resources
    var errors []error
    for _, resource := range a.resources {
        if err := resource.Close(); err != nil {
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("cleanup errors: %v", errors)
    }
    
    return nil
}

func (a *MyAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Check if shutting down
    select {
    case <-a.shutdown:
        return AgentResult{}, errors.New("agent is shutting down")
    default:
    }
    
    // Track ongoing operation
    a.wg.Add(1)
    defer a.wg.Done()
    
    // Execute normally
    return a.execute(ctx, event, state)
}
```

### 2. Resource Cleanup

Ensure proper resource cleanup:

```go
type ResourceCleanup struct {
    resources []CleanupFunc
    mu        sync.Mutex
}

type CleanupFunc func() error

func (rc *ResourceCleanup) AddCleanup(cleanup CleanupFunc) {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    rc.resources = append(rc.resources, cleanup)
}

func (rc *ResourceCleanup) Cleanup() error {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    
    var errors []error
    
    // Cleanup in reverse order
    for i := len(rc.resources) - 1; i >= 0; i-- {
        if err := rc.resources[i](); err != nil {
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("cleanup errors: %v", errors)
    }
    
    return nil
}
```

## Agent Lifecycle Management

### 1. Agent Manager

Centralized agent lifecycle management:

```go
type AgentManager struct {
    agents    map[string]Agent
    lifecycle map[string]*AgentLifecycle
    mu        sync.RWMutex
}

type AgentLifecycle struct {
    Agent       Agent
    Status      AgentStatus
    CreatedAt   time.Time
    StartedAt   *time.Time
    StoppedAt   *time.Time
    HealthCheck HealthCheckable
    Monitor     *PerformanceMonitor
}

type AgentStatus int

const (
    AgentStatusCreated AgentStatus = iota
    AgentStatusStarted
    AgentStatusStopped
    AgentStatusError
)

func (am *AgentManager) RegisterAgent(name string, agent Agent) error {
    am.mu.Lock()
    defer am.mu.Unlock()
    
    if _, exists := am.agents[name]; exists {
        return fmt.Errorf("agent already registered: %s", name)
    }
    
    am.agents[name] = agent
    am.lifecycle[name] = &AgentLifecycle{
        Agent:     agent,
        Status:    AgentStatusCreated,
        CreatedAt: time.Now(),
        Monitor:   &PerformanceMonitor{agentName: name, maxHistorySize: 100},
    }
    
    return nil
}

func (am *AgentManager) StartAgent(ctx context.Context, name string) error {
    am.mu.Lock()
    lifecycle, exists := am.lifecycle[name]
    am.mu.Unlock()
    
    if !exists {
        return fmt.Errorf("agent not found: %s", name)
    }
    
    // Initialize if needed
    if initializable, ok := lifecycle.Agent.(InitializableAgent); ok {
        if err := initializable.Initialize(ctx); err != nil {
            lifecycle.Status = AgentStatusError
            return fmt.Errorf("agent initialization failed: %w", err)
        }
    }
    
    now := time.Now()
    lifecycle.StartedAt = &now
    lifecycle.Status = AgentStatusStarted
    
    return nil
}

func (am *AgentManager) StopAgent(ctx context.Context, name string) error {
    am.mu.Lock()
    lifecycle, exists := am.lifecycle[name]
    am.mu.Unlock()
    
    if !exists {
        return fmt.Errorf("agent not found: %s", name)
    }
    
    // Graceful shutdown if supported
    if graceful, ok := lifecycle.Agent.(GracefulAgent); ok {
        if err := graceful.Shutdown(ctx); err != nil {
            return fmt.Errorf("graceful shutdown failed: %w", err)
        }
    }
    
    now := time.Now()
    lifecycle.StoppedAt = &now
    lifecycle.Status = AgentStatusStopped
    
    return nil
}
```

### 2. Lifecycle Events

Emit events during lifecycle transitions:

```go
type LifecycleEvent struct {
    AgentName string
    Event     LifecycleEventType
    Timestamp time.Time
    Data      map[string]interface{}
}

type LifecycleEventType string

const (
    LifecycleEventCreated     LifecycleEventType = "created"
    LifecycleEventInitialized LifecycleEventType = "initialized"
    LifecycleEventStarted     LifecycleEventType = "started"
    LifecycleEventStopped     LifecycleEventType = "stopped"
    LifecycleEventError       LifecycleEventType = "error"
)

type LifecycleEventEmitter struct {
    listeners []LifecycleEventListener
    mu        sync.RWMutex
}

type LifecycleEventListener func(event LifecycleEvent)

func (lee *LifecycleEventEmitter) AddListener(listener LifecycleEventListener) {
    lee.mu.Lock()
    defer lee.mu.Unlock()
    lee.listeners = append(lee.listeners, listener)
}

func (lee *LifecycleEventEmitter) EmitEvent(event LifecycleEvent) {
    lee.mu.RLock()
    listeners := make([]LifecycleEventListener, len(lee.listeners))
    copy(listeners, lee.listeners)
    lee.mu.RUnlock()
    
    for _, listener := range listeners {
        go listener(event)
    }
}
```

## Best Practices

### 1. Resource Management

```go
// Always use defer for cleanup
func (a *MyAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    resource, err := a.acquireResource(ctx)
    if err != nil {
        return AgentResult{}, err
    }
    defer resource.Release()
    
    // Use resource...
    return a.processWithResource(ctx, event, state, resource)
}
```

### 2. Error Handling in Lifecycle

```go
// Handle initialization errors gracefully
func (a *MyAgent) Initialize(ctx context.Context) error {
    var errors []error
    
    for _, initializer := range a.initializers {
        if err := initializer.Initialize(ctx); err != nil {
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        // Cleanup any successful initializations
        a.cleanup()
        return fmt.Errorf("initialization failed: %v", errors)
    }
    
    return nil
}
```

### 3. Monitoring Integration

```go
// Integrate monitoring throughout lifecycle
func (a *MyAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        a.monitor.RecordExecution(duration, getCurrentMemoryUsage())
    }()
    
    return a.execute(ctx, event, state)
}
```

## Conclusion

Understanding the agent lifecycle is crucial for building robust multi-agent systems. Proper lifecycle management ensures agents are created correctly, execute reliably, and clean up resources appropriately.

Key takeaways:
- **Use builder patterns** for flexible agent creation
- **Implement proper initialization** and cleanup
- **Monitor agent health** and performance
- **Handle lifecycle transitions** gracefully
- **Manage resources** carefully throughout the lifecycle
- **Emit lifecycle events** for observability

## Next Steps

- [Error Handling](error-handling.md) - Learn robust error management
- [Memory Systems](../memory-systems/README.md) - Understand persistent storage
- [Debugging Guide](../debugging/README.md) - Debug agent issues
- [Production Deployment](../../guides/deployment/README.md) - Deploy agents in production

## Further Reading

- [API Reference: Agent Interfaces](../../reference/api/agent.md#agents)
- [Examples: Agent Lifecycle Patterns](../../examples/)
- [Configuration Guide: Agent Settings](../../reference/api/configuration.md)