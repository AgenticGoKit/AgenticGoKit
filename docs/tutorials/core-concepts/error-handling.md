# Error Handling and Recovery in AgenticGoKit

## Overview

Robust error handling is critical for building reliable multi-agent systems. This tutorial explores AgenticGoKit's error handling mechanisms, recovery strategies, and best practices for building fault-tolerant agent workflows.

Understanding error handling is essential because agent systems involve multiple components, network calls, and complex interactions where failures can occur at any point.

## Prerequisites

- Understanding of [Message Passing and Event Flow](message-passing.md)
- Knowledge of [State Management](state-management.md)
- Basic understanding of Go error handling patterns

## Core Error Handling Concepts

### Error Types in AgenticGoKit

AgenticGoKit handles several types of errors:

1. **Agent Execution Errors**: Failures during agent processing
2. **Orchestration Errors**: Issues with agent coordination
3. **Communication Errors**: Problems with event routing
4. **Resource Errors**: Database, API, or external service failures
5. **Validation Errors**: Invalid input or state validation failures

### Error Flow Architecture

```
┌─────────┐    ┌──────────┐    ┌─────────────┐    ┌─────────────┐
│  Event  │───▶│  Runner  │───▶│ Orchestrator│───▶│    Agent    │
└─────────┘    └──────────┘    └─────────────┘    └─────────────┘
                     │                │                   │
                     ▼                ▼                   ▼
               ┌──────────┐    ┌─────────────┐    ┌─────────────┐
               │Error Hook│    │Error Router │    │Error Result │
               └──────────┘    └─────────────┘    └─────────────┘
```

## Error Handling Mechanisms

### 1. Agent-Level Error Handling

Agents can return errors through the `AgentResult`:

```go
func (a *MyAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Validate input
    query, ok := state.Get("query")
    if !ok {
        return AgentResult{
            Error: "missing required field: query",
        }, errors.New("query not found in state")
    }
    
    // Process with error handling
    result, err := a.processQuery(query.(string))
    if err != nil {
        return AgentResult{
            Error: fmt.Sprintf("processing failed: %v", err),
            OutputState: state, // Return original state
        }, err
    }
    
    // Success case
    outputState := state.Clone()
    outputState.Set("response", result)
    
    return AgentResult{
        OutputState: outputState,
    }, nil
}
```

### 2. Error Hooks and Callbacks

AgenticGoKit provides hooks for intercepting and handling errors:

```go
// Register error handling callback
runner.RegisterCallback(core.HookAgentError, "error-handler",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        fmt.Printf("Agent %s failed: %v\n", args.AgentID, args.Error)
        
        // Log error details
        logError(args.AgentID, args.Error, args.Event)
        
        // Optionally modify state or trigger recovery
        if isRecoverableError(args.Error) {
            return handleRecoverableError(args)
        }
        
        return args.State, nil
    },
)
```

### 3. Enhanced Error Routing

AgenticGoKit includes sophisticated error routing capabilities:

```go
// Configure error routing
errorConfig := &core.ErrorRouterConfig{
    MaxRetries: 3,
    RetryDelay: time.Second * 2,
    FallbackAgent: "error-recovery-agent",
    ErrorClassification: map[string]core.ErrorAction{
        "timeout":     core.ErrorActionRetry,
        "rate_limit":  core.ErrorActionDelay,
        "auth_error":  core.ErrorActionFallback,
        "fatal_error": core.ErrorActionFail,
    },
}

runner.SetErrorRouterConfig(errorConfig)
```

## Error Recovery Strategies

### 1. Retry Mechanisms

Implement automatic retry for transient failures:

```go
type RetryableAgent struct {
    baseAgent Agent
    maxRetries int
    retryDelay time.Duration
}

func (r *RetryableAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    var lastErr error
    
    for attempt := 0; attempt <= r.maxRetries; attempt++ {
        if attempt > 0 {
            // Wait before retry
            select {
            case <-time.After(r.retryDelay):
            case <-ctx.Done():
                return AgentResult{}, ctx.Err()
            }
        }
        
        result, err := r.baseAgent.Run(ctx, event, state)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !isRetryableError(err) {
            break
        }
        
        fmt.Printf("Attempt %d failed, retrying: %v\n", attempt+1, err)
    }
    
    return AgentResult{
        Error: fmt.Sprintf("failed after %d attempts: %v", r.maxRetries+1, lastErr),
    }, lastErr
}

func isRetryableError(err error) bool {
    // Define which errors are worth retrying
    return strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "connection") ||
           strings.Contains(err.Error(), "rate limit")
}
```

### 2. Circuit Breaker Pattern

Prevent cascading failures with circuit breakers:

```go
type CircuitBreakerAgent struct {
    baseAgent Agent
    breaker   *CircuitBreaker
}

type CircuitBreaker struct {
    maxFailures int
    resetTimeout time.Duration
    state       CircuitState
    failures    int
    lastFailure time.Time
    mu          sync.RWMutex
}

type CircuitState int

const (
    CircuitClosed CircuitState = iota
    CircuitOpen
    CircuitHalfOpen
)

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    // Check if circuit should reset
    if cb.state == CircuitOpen && time.Since(cb.lastFailure) > cb.resetTimeout {
        cb.state = CircuitHalfOpen
        cb.failures = 0
    }
    
    // Reject if circuit is open
    if cb.state == CircuitOpen {
        return errors.New("circuit breaker is open")
    }
    
    // Execute function
    err := fn()
    
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = CircuitOpen
        }
        return err
    }
    
    // Success - reset circuit
    cb.failures = 0
    cb.state = CircuitClosed
    return nil
}

func (cba *CircuitBreakerAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    var result AgentResult
    var err error
    
    breakerErr := cba.breaker.Call(func() error {
        result, err = cba.baseAgent.Run(ctx, event, state)
        return err
    })
    
    if breakerErr != nil {
        return AgentResult{
            Error: fmt.Sprintf("circuit breaker: %v", breakerErr),
        }, breakerErr
    }
    
    return result, err
}
```

### 3. Fallback Agents

Implement fallback mechanisms for critical failures:

```go
type FallbackAgent struct {
    primaryAgent   Agent
    fallbackAgent  Agent
    fallbackTrigger func(error) bool
}

func (f *FallbackAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Try primary agent first
    result, err := f.primaryAgent.Run(ctx, event, state)
    
    // If primary succeeds, return result
    if err == nil {
        return result, nil
    }
    
    // Check if we should use fallback
    if !f.fallbackTrigger(err) {
        return result, err
    }
    
    fmt.Printf("Primary agent failed, using fallback: %v\n", err)
    
    // Try fallback agent
    fallbackResult, fallbackErr := f.fallbackAgent.Run(ctx, event, state)
    
    if fallbackErr != nil {
        // Both failed - return combined error
        return AgentResult{
            Error: fmt.Sprintf("primary failed: %v, fallback failed: %v", err, fallbackErr),
        }, fmt.Errorf("both primary and fallback failed: %v, %v", err, fallbackErr)
    }
    
    // Mark result as from fallback
    if fallbackResult.OutputState != nil {
        fallbackResult.OutputState.SetMeta("fallback_used", "true")
        fallbackResult.OutputState.SetMeta("primary_error", err.Error())
    }
    
    return fallbackResult, nil
}
```

## Error Handling in Different Orchestration Patterns

### 1. Route Orchestration Error Handling

```go
// Route orchestrator with error handling
func (o *RouteOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
    targetName, ok := event.GetMetadataValue(RouteMetadataKey)
    if !ok {
        return o.handleRoutingError(event, errors.New("missing route metadata"))
    }
    
    handler, exists := o.handlers[targetName]
    if !exists {
        return o.handleRoutingError(event, fmt.Errorf("agent not found: %s", targetName))
    }
    
    // Execute with timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    result, err := handler.Run(ctx, event, core.NewState())
    if err != nil {
        return o.handleAgentError(event, targetName, err)
    }
    
    return result, nil
}

func (o *RouteOrchestrator) handleRoutingError(event Event, err error) (AgentResult, error) {
    // Create error event
    errorEvent := core.NewEvent(
        "error-handler",
        core.EventData{
            "error_type": "routing_error",
            "error_message": err.Error(),
            "original_event": event,
        },
        map[string]string{
            "route": "error-handler",
            "session_id": event.GetSessionID(),
        },
    )
    
    // Emit error event if emitter is available
    if o.emitter != nil {
        o.emitter.Emit(errorEvent)
    }
    
    return AgentResult{
        Error: err.Error(),
    }, err
}
```

### 2. Collaborative Orchestration Error Handling

```go
// Collaborative orchestrator with partial failure handling
func (o *CollaborativeOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
    var wg sync.WaitGroup
    results := make([]AgentResult, 0)
    errors := make([]error, 0)
    mu := &sync.Mutex{}
    
    // Execute all agents
    for name, handler := range o.handlers {
        wg.Add(1)
        go func(agentName string, agentHandler AgentHandler) {
            defer wg.Done()
            
            result, err := agentHandler.Run(ctx, event, core.NewState())
            
            mu.Lock()
            if err != nil {
                errors = append(errors, fmt.Errorf("agent %s: %w", agentName, err))
            } else {
                results = append(results, result)
            }
            mu.Unlock()
        }(name, handler)
    }
    
    wg.Wait()
    
    // Handle partial failures
    totalAgents := len(o.handlers)
    successCount := len(results)
    failureCount := len(errors)
    
    // Check if we have enough successes
    successThreshold := 0.5 // At least 50% must succeed
    if float64(successCount)/float64(totalAgents) < successThreshold {
        return AgentResult{
            Error: fmt.Sprintf("insufficient successes: %d/%d agents failed", failureCount, totalAgents),
        }, fmt.Errorf("collaborative orchestration failed: %v", errors)
    }
    
    // Combine successful results
    combinedResult := o.combineResults(results)
    
    // Add failure information to metadata
    if len(errors) > 0 {
        if combinedResult.OutputState != nil {
            combinedResult.OutputState.SetMeta("partial_failures", fmt.Sprintf("%d", failureCount))
            combinedResult.OutputState.SetMeta("success_rate", fmt.Sprintf("%.2f", float64(successCount)/float64(totalAgents)))
        }
    }
    
    return combinedResult, nil
}
```

### 3. Sequential Orchestration Error Handling

```go
// Sequential orchestrator with rollback capability
func (o *SequentialOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
    currentState := core.NewState()
    completedStages := make([]string, 0)
    
    // Merge event data
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    // Process through sequence
    for i, agentName := range o.sequence {
        handler, exists := o.handlers[agentName]
        if !exists {
            return o.rollback(completedStages, fmt.Errorf("agent %s not found", agentName))
        }
        
        // Create stage event
        stageEvent := core.NewEvent(agentName, currentState.GetAll(), event.GetMetadata())
        
        // Execute with timeout
        stageCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
        result, err := handler.Run(stageCtx, stageEvent, currentState)
        cancel()
        
        if err != nil {
            return o.rollback(completedStages, fmt.Errorf("stage %d (%s) failed: %w", i+1, agentName, err))
        }
        
        // Update state and track completion
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        completedStages = append(completedStages, agentName)
        
        fmt.Printf("Stage %d (%s) completed successfully\n", i+1, agentName)
    }
    
    return AgentResult{OutputState: currentState}, nil
}

func (o *SequentialOrchestrator) rollback(completedStages []string, err error) (AgentResult, error) {
    fmt.Printf("Rolling back %d completed stages due to error: %v\n", len(completedStages), err)
    
    // Execute rollback in reverse order
    for i := len(completedStages) - 1; i >= 0; i-- {
        stageName := completedStages[i]
        if rollbackHandler, exists := o.rollbackHandlers[stageName]; exists {
            rollbackCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            rollbackHandler.Rollback(rollbackCtx, stageName)
            cancel()
        }
    }
    
    return AgentResult{
        Error: fmt.Sprintf("sequential processing failed: %v", err),
    }, err
}
```

## Error Monitoring and Observability

### 1. Error Metrics Collection

```go
type ErrorMetrics struct {
    errorCounts    map[string]int64
    errorRates     map[string]float64
    lastErrors     map[string]time.Time
    mu             sync.RWMutex
}

func (em *ErrorMetrics) RecordError(agentID string, errorType string) {
    em.mu.Lock()
    defer em.mu.Unlock()
    
    key := fmt.Sprintf("%s:%s", agentID, errorType)
    em.errorCounts[key]++
    em.lastErrors[key] = time.Now()
    
    // Calculate error rate (errors per minute)
    em.calculateErrorRate(key)
}

func (em *ErrorMetrics) calculateErrorRate(key string) {
    // Implementation for calculating error rates
    // This would typically involve time windows and moving averages
}

// Register error metrics callback
runner.RegisterCallback(core.HookAgentError, "metrics-collector",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        errorType := classifyError(args.Error)
        errorMetrics.RecordError(args.AgentID, errorType)
        return args.State, nil
    },
)
```

### 2. Error Alerting

```go
type ErrorAlerter struct {
    thresholds map[string]AlertThreshold
    notifier   Notifier
}

type AlertThreshold struct {
    ErrorRate    float64       // Errors per minute
    TimeWindow   time.Duration // Time window for rate calculation
    Cooldown     time.Duration // Minimum time between alerts
}

func (ea *ErrorAlerter) CheckThresholds(agentID string, errorRate float64) {
    threshold, exists := ea.thresholds[agentID]
    if !exists {
        return
    }
    
    if errorRate > threshold.ErrorRate {
        alert := Alert{
            AgentID:   agentID,
            ErrorRate: errorRate,
            Threshold: threshold.ErrorRate,
            Timestamp: time.Now(),
        }
        
        ea.notifier.SendAlert(alert)
    }
}
```

### 3. Error Logging and Tracing

```go
type ErrorLogger struct {
    logger *log.Logger
    tracer trace.Tracer
}

func (el *ErrorLogger) LogError(ctx context.Context, agentID string, err error, event Event) {
    // Create span for error
    ctx, span := el.tracer.Start(ctx, "agent_error")
    defer span.End()
    
    // Add error attributes
    span.SetAttributes(
        attribute.String("agent.id", agentID),
        attribute.String("error.message", err.Error()),
        attribute.String("event.id", event.GetID()),
    )
    
    // Log structured error
    el.logger.Printf("AGENT_ERROR agent=%s event=%s error=%v", 
        agentID, event.GetID(), err)
    
    // Record error in span
    span.RecordError(err)
}
```

## Testing Error Scenarios

### 1. Error Injection for Testing

```go
type ErrorInjectingAgent struct {
    baseAgent     Agent
    errorRate     float64 // 0.0 to 1.0
    errorTypes    []error
    random        *rand.Rand
}

func (eia *ErrorInjectingAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Inject errors based on configured rate
    if eia.random.Float64() < eia.errorRate {
        errorIndex := eia.random.Intn(len(eia.errorTypes))
        injectedError := eia.errorTypes[errorIndex]
        
        return AgentResult{
            Error: fmt.Sprintf("injected error: %v", injectedError),
        }, injectedError
    }
    
    // Normal execution
    return eia.baseAgent.Run(ctx, event, state)
}
```

### 2. Error Scenario Testing

```go
func TestErrorRecovery(t *testing.T) {
    // Create agent that fails first time, succeeds second time
    failingAgent := &FailingAgent{failCount: 1}
    retryAgent := &RetryableAgent{
        baseAgent:  failingAgent,
        maxRetries: 2,
        retryDelay: time.Millisecond * 100,
    }
    
    // Test retry mechanism
    event := core.NewEvent("test", core.EventData{"test": "data"}, nil)
    result, err := retryAgent.Run(context.Background(), event, core.NewState())
    
    assert.NoError(t, err)
    assert.NotEmpty(t, result.OutputState)
}

func TestCircuitBreaker(t *testing.T) {
    // Create agent that always fails
    alwaysFailingAgent := &AlwaysFailingAgent{}
    circuitAgent := &CircuitBreakerAgent{
        baseAgent: alwaysFailingAgent,
        breaker: &CircuitBreaker{
            maxFailures:  3,
            resetTimeout: time.Second,
        },
    }
    
    // Test that circuit opens after max failures
    for i := 0; i < 5; i++ {
        _, err := circuitAgent.Run(context.Background(), event, core.NewState())
        if i < 3 {
            assert.Contains(t, err.Error(), "always fails")
        } else {
            assert.Contains(t, err.Error(), "circuit breaker is open")
        }
    }
}
```

## Best Practices for Error Handling

### 1. Error Classification

```go
type ErrorClass int

const (
    ErrorClassTransient ErrorClass = iota // Temporary, retry-able
    ErrorClassPermanent                   // Permanent, don't retry
    ErrorClassResource                    // Resource exhaustion
    ErrorClassValidation                  // Input validation
    ErrorClassSecurity                    // Security/auth issues
)

func ClassifyError(err error) ErrorClass {
    errStr := strings.ToLower(err.Error())
    
    switch {
    case strings.Contains(errStr, "timeout"):
        return ErrorClassTransient
    case strings.Contains(errStr, "connection"):
        return ErrorClassTransient
    case strings.Contains(errStr, "rate limit"):
        return ErrorClassResource
    case strings.Contains(errStr, "validation"):
        return ErrorClassValidation
    case strings.Contains(errStr, "unauthorized"):
        return ErrorClassSecurity
    default:
        return ErrorClassPermanent
    }
}
```

### 2. Graceful Degradation

```go
type GracefulAgent struct {
    primaryAgent   Agent
    degradedMode   Agent
    healthChecker  HealthChecker
}

func (ga *GracefulAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Check if we should use degraded mode
    if !ga.healthChecker.IsHealthy() {
        fmt.Println("Using degraded mode due to health issues")
        result, err := ga.degradedMode.Run(ctx, event, state)
        if err == nil && result.OutputState != nil {
            result.OutputState.SetMeta("degraded_mode", "true")
        }
        return result, err
    }
    
    // Try primary agent
    result, err := ga.primaryAgent.Run(ctx, event, state)
    if err != nil {
        // Mark as unhealthy and try degraded mode
        ga.healthChecker.MarkUnhealthy()
        return ga.degradedMode.Run(ctx, event, state)
    }
    
    return result, nil
}
```

### 3. Error Context Preservation

```go
type ContextualError struct {
    Err       error
    AgentID   string
    EventID   string
    SessionID string
    Timestamp time.Time
    Context   map[string]interface{}
}

func (ce *ContextualError) Error() string {
    return fmt.Sprintf("agent=%s event=%s session=%s: %v", 
        ce.AgentID, ce.EventID, ce.SessionID, ce.Err)
}

func WrapError(err error, agentID string, event Event) error {
    return &ContextualError{
        Err:       err,
        AgentID:   agentID,
        EventID:   event.GetID(),
        SessionID: event.GetSessionID(),
        Timestamp: time.Now(),
        Context:   make(map[string]interface{}),
    }
}
```

## Common Error Patterns and Solutions

### 1. Timeout Handling

```go
func WithTimeout(agent Agent, timeout time.Duration) Agent {
    return AgentFunc(func(ctx context.Context, event Event, state State) (AgentResult, error) {
        ctx, cancel := context.WithTimeout(ctx, timeout)
        defer cancel()
        
        done := make(chan struct{})
        var result AgentResult
        var err error
        
        go func() {
            result, err = agent.Run(ctx, event, state)
            close(done)
        }()
        
        select {
        case <-done:
            return result, err
        case <-ctx.Done():
            return AgentResult{
                Error: fmt.Sprintf("agent timeout after %v", timeout),
            }, ctx.Err()
        }
    })
}
```

### 2. Resource Exhaustion Handling

```go
type ResourceLimitedAgent struct {
    baseAgent Agent
    semaphore chan struct{}
}

func NewResourceLimitedAgent(agent Agent, maxConcurrency int) *ResourceLimitedAgent {
    return &ResourceLimitedAgent{
        baseAgent: agent,
        semaphore: make(chan struct{}, maxConcurrency),
    }
}

func (rla *ResourceLimitedAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Acquire resource
    select {
    case rla.semaphore <- struct{}{}:
        defer func() { <-rla.semaphore }()
    case <-ctx.Done():
        return AgentResult{
            Error: "resource acquisition timeout",
        }, ctx.Err()
    }
    
    return rla.baseAgent.Run(ctx, event, state)
}
```

## Conclusion

Effective error handling is crucial for building robust multi-agent systems. AgenticGoKit provides comprehensive error handling mechanisms including hooks, routing, recovery strategies, and monitoring capabilities.

Key takeaways:
- **Classify errors** appropriately for proper handling
- **Implement retry mechanisms** for transient failures
- **Use circuit breakers** to prevent cascading failures
- **Provide fallback options** for critical functionality
- **Monitor and alert** on error patterns
- **Test error scenarios** thoroughly
- **Preserve error context** for debugging

## Next Steps

- [Memory Systems](../memory-systems/README.md) - Learn about persistent storage and RAG
- [Debugging Guide](../debugging/README.md) - Advanced debugging techniques
- [Performance Optimization](../advanced/README.md) - Optimize agent performance
- [Production Deployment](../../guides/deployment/README.md) - Deploy with proper error handling

## Further Reading

- [API Reference: Error Handling](../../reference/api/agent.md#error-handling)
- [Examples: Error Recovery Patterns](../../examples/)
- [Configuration Guide: Error Settings](../../reference/api/configuration.md)