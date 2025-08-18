# Error Handling Guide

AgentFlow provides comprehensive error handling mechanisms to build resilient agent systems. This guide covers error patterns, recovery strategies, and best practices for production-ready error handling.

## 🎯 Error Handling Philosophy

AgentFlow follows these principles for error handling:
- **Graceful Degradation**: Agents should continue functioning with reduced capabilities when possible
- **Context Preservation**: Error information should include enough context for debugging
- **Recovery Strategies**: Automatic retry and fallback mechanisms for transient failures
- **User-Friendly Messages**: End users should receive helpful, non-technical error messages

## 🔧 Core Error Types

### **Agent Execution Errors**
Errors that occur during agent processing:

```go
type AgentError struct {
    AgentName string
    Operation string
    Err       error
    Context   map[string]interface{}
}


func (e *AgentError) Error() string {
    return fmt.Sprintf("agent %s failed in %s: %v", e.AgentName, e.Operation, e.Err)
}
```

### **Tool Execution Errors**
Errors from MCP tool calls:

```go
type ToolError struct {
    ToolName string
    ServerName string
    Method   string
    Args     map[string]interface{}
    Err      error
}

func (e *ToolError) Error() string {
    return fmt.Sprintf("tool %s/%s failed: %v", e.ServerName, e.ToolName, e.Err)
}
```

### **Provider Errors**
LLM provider specific errors:

```go
type ProviderError struct {
    Provider string
    Type     ProviderErrorType
    Err      error
    Retryable bool
}

type ProviderErrorType string

const (
    ProviderRateLimit   ProviderErrorType = "rate_limit"
    ProviderTimeout     ProviderErrorType = "timeout"
    ProviderUnauthorized ProviderErrorType = "unauthorized"
    ProviderQuotaExceeded ProviderErrorType = "quota_exceeded"
    ProviderServiceError ProviderErrorType = "service_error"
)
```

## 🛡️ Error Handling Strategies

### **1. Graceful Error Recovery**

Implement agents that can continue with partial failures:

```go
func (h *ResilientHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    result := core.AgentResult{
        Data: make(map[string]interface{}),
        Errors: []error{},
    }
    
    // Try primary tool first
    primaryResult, err := h.tryPrimaryTool(ctx, event, state)
    if err != nil {
        result.Errors = append(result.Errors, fmt.Errorf("primary tool failed: %w", err))
        
        // Fall back to secondary tool
        fallbackResult, fallbackErr := h.tryFallbackTool(ctx, event, state)
        if fallbackErr != nil {
            result.Errors = append(result.Errors, fmt.Errorf("fallback tool failed: %w", fallbackErr))
            return result, fmt.Errorf("all tools failed")
        }
        
        result.Data["source"] = "fallback"
        result.Data["result"] = fallbackResult
        result.Data["warnings"] = []string{"Primary tool unavailable, used fallback"}
    } else {
        result.Data["source"] = "primary"
        result.Data["result"] = primaryResult
    }
    
    return result, nil
}
```

### **2. Retry Logic with Exponential Backoff**

AgentFlow provides built-in retry mechanisms:

```go
func (h *RetryableHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    retryConfig := &core.RetryConfig{
        MaxAttempts:     3,
        InitialInterval: time.Second,
        MaxInterval:     10 * time.Second,
        Multiplier:      2.0,
        RetryableErrors: []core.RetryableErrorChecker{
            &core.ProviderTimeoutChecker{},
            &core.RateLimitChecker{},
            &core.NetworkErrorChecker{},
        },
    }
            // Pseudocode for fault-tolerant orchestration handling; use Runner Start/Emit/Stop in real code
            // ... implementation ...
            return nil
        return h.executeWithPossibleFailure(ctx, event, state)
    })
}
```

### **3. Circuit Breaker Pattern**

Prevent cascading failures with circuit breakers:

```go
func NewCircuitBreakerHandler(handler core.AgentHandler) *CircuitBreakerHandler {
    return &CircuitBreakerHandler{
        handler: handler,
        breaker: core.NewCircuitBreaker(&core.CircuitBreakerConfig{
            FailureThreshold:   5,
            RecoveryTimeout:    30 * time.Second,
            SuccessThreshold:   3,
            Timeout:           10 * time.Second,
        }),
    }
}

func (h *CircuitBreakerHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    result, err := h.breaker.Execute(ctx, func(ctx context.Context) (interface{}, error) {
        return h.handler.Run(ctx, event, state)
    })
    
    if err != nil {
        if core.IsCircuitBreakerOpen(err) {
            return core.AgentResult{
                Data: map[string]interface{}{
                    "error": "Service temporarily unavailable",
                    "retry_after": 30,
                },
            }, nil // Return graceful degradation instead of error
        }
        return core.AgentResult{}, err
    }
    
    return result.(core.AgentResult), nil
}
```

### **4. Input Validation and Sanitization**

Validate inputs before processing:

```go
func (h *ValidatedHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Validate event data
    if err := h.validateEvent(event); err != nil {
        return core.AgentResult{}, &core.ValidationError{
            Field:   "event",
            Message: "Invalid event data",
            Err:     err,
        }
    }
    
    // Validate state
    if err := h.validateState(state); err != nil {
        return core.AgentResult{}, &core.ValidationError{
            Field:   "state", 
            Message: "Invalid state data",
            Err:     err,
        }
    }
    
    // Sanitize input
    sanitizedEvent := h.sanitizeEvent(event)
    
    return h.processValidatedEvent(ctx, sanitizedEvent, state)
}

func (h *ValidatedHandler) validateEvent(event core.Event) error {
    data := event.GetData()
    
    // Check required fields
    if query, ok := data["query"].(string); !ok || strings.TrimSpace(query) == "" {
        return fmt.Errorf("query field is required and cannot be empty")
    }
    
    // Check data size
    if len(fmt.Sprintf("%v", data)) > h.maxEventSize {
        return fmt.Errorf("event data exceeds maximum size of %d bytes", h.maxEventSize)
    }
    
    // Content safety check
    if !h.isContentSafe(data) {
        return fmt.Errorf("event contains unsafe content")
    }
    
    return nil
}
```

## 🚨 Error Routing and Handling

### **Automatic Error Routing**

AgentFlow can automatically route errors to specialized handlers:

```go
func setupErrorRouting(runner *core.Runner) error {
    // Register error handlers for different error types
    runner.RegisterErrorHandler(core.ValidationErrorType, &ValidationErrorHandler{})
    runner.RegisterErrorHandler(core.TimeoutErrorType, &TimeoutErrorHandler{})
    runner.RegisterErrorHandler(core.RateLimitErrorType, &RateLimitErrorHandler{})
    runner.RegisterErrorHandler(core.CriticalErrorType, &CriticalErrorHandler{})
    
    return nil
}

type ValidationErrorHandler struct{}

func (h *ValidationErrorHandler) HandleError(ctx context.Context, err error, event core.Event) core.AgentResult {
    validationErr, ok := err.(*core.ValidationError)
    if !ok {
        return core.AgentResult{
            Data: map[string]interface{}{
                "error": "Internal validation error",
            },
        }
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "error": "Invalid input",
            "field": validationErr.Field,
            "message": validationErr.GetUserFriendlyMessage(),
            "suggestions": validationErr.GetSuggestions(),
        },
    }
}
```

### **Custom Error Middleware**

Create middleware for consistent error handling:

```go
func ErrorHandlingMiddleware(next core.AgentHandler) core.AgentHandler {
    return core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Agent panic recovered: %v\nStack: %s", r, debug.Stack())
                // Could emit error event or alert here
            }
        }()
        
        // Add request ID for tracing
        ctx = core.WithRequestID(ctx, core.GenerateRequestID())
        
        // Execute handler with timeout
        timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
        
        result, err := next.Run(timeoutCtx, event, state)
        
        if err != nil {
            // Log error with context
            log.Printf("Agent error [%s]: %v", core.GetRequestID(ctx), err)
            
            // Transform error for user
            if userErr := transformErrorForUser(err); userErr != nil {
                return core.AgentResult{
                    Data: map[string]interface{}{
                        "error": userErr.Error(),
                        "request_id": core.GetRequestID(ctx),
                    },
                }, nil
            }
            
            return result, err
        }
        
        return result, nil
    })
}
```

## 📊 Error Monitoring and Observability

### **Error Metrics**

Track error patterns and frequencies:

```go
func setupErrorMetrics() {
    // Prometheus metrics for error tracking
    errorCounter := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agent_errors_total",
            Help: "Total number of agent errors by type and agent",
        },
        []string{"agent_name", "error_type", "severity"},
    )
    
    errorDuration := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "agent_error_recovery_duration_seconds",
            Help: "Time taken to recover from errors",
            Buckets: prometheus.DefBuckets,
        },
        []string{"agent_name", "recovery_strategy"},
    )
    
    prometheus.MustRegister(errorCounter, errorDuration)
}

func recordError(agentName string, err error) {
    errorType := classifyError(err)
    severity := determineSeverity(err)
    
    errorCounter.WithLabelValues(agentName, errorType, severity).Inc()
    
    // Also log structured error data
    log.WithFields(log.Fields{
        "agent_name": agentName,
        "error_type": errorType,
        "severity":   severity,
        "error":      err.Error(),
        "timestamp":  time.Now(),
    }).Error("Agent execution error")
}
```

### **Error Context Collection**

Gather comprehensive context for debugging:

```go
func enrichErrorContext(ctx context.Context, err error, event core.Event, state core.State) *ErrorReport {
    return &ErrorReport{
        Error:       err,
        RequestID:   core.GetRequestID(ctx),
        Timestamp:   time.Now(),
        AgentName:   core.GetAgentName(ctx),
        SessionID:   core.GetSessionID(ctx),
        Event: ErrorEventContext{
            Type:     event.GetType(),
            Data:     sanitizeForLogging(event.GetData()),
            UserID:   event.GetUserID(),
        },
        State: ErrorStateContext{
            Keys:        state.Keys(),
            Size:        len(state.String()),
            LastUpdate:  state.GetLastModified(),
        },
        Environment: ErrorEnvironmentContext{
            Version:     core.Version(),
            GoVersion:   runtime.Version(),
            Platform:    runtime.GOOS + "/" + runtime.GOARCH,
            MemStats:    getMemoryStats(),
        },
        Stacktrace: string(debug.Stack()),
    }
}
```

## 🔄 Recovery Patterns

### **Stateful Recovery**

For agents that maintain state across calls:

```go
func (h *StatefulHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Create checkpoint before risky operation
    checkpoint := state.CreateCheckpoint()
    
    defer func() {
        if r := recover(); r != nil {
            // Restore from checkpoint on panic
            state.RestoreFromCheckpoint(checkpoint)
            log.Printf("Restored state from checkpoint due to panic: %v", r)
        }
    }()
    
    result, err := h.riskyOperation(ctx, event, state)
    if err != nil {
        // Restore state on error
        state.RestoreFromCheckpoint(checkpoint)
        
        // Try alternative approach with clean state
        return h.fallbackOperation(ctx, event, state)
    }
    
    return result, nil
}
```

### **Multi-Agent Error Recovery**

Coordinate error handling across multiple agents (pseudocode example):

```go
func (o *FaultTolerantOrchestrator) Handle(ctx context.Context, event core.Event) error {
    agents := o.getActiveAgents()
    results := make(chan agentResult, len(agents))
    
    // Run agents in parallel
    for _, agent := range agents {
        go func(a core.Agent) {
            result, err := a.Run(ctx, event.Clone(), o.state.Clone())
            results <- agentResult{agent: a.Name(), result: result, err: err}
        }(agent)
    }
    
    // Collect results and handle failures
    var successful []agentResult
    var failed []agentResult
    
    for i := 0; i < len(agents); i++ {
        result := <-results
        if result.err != nil {
            failed = append(failed, result)
        } else {
            successful = append(successful, result)
        }
    }
    
    // Determine if we can proceed with partial success
    if len(successful) >= o.minSuccessfulAgents {
        return o.consolidateResults(successful, failed)
    }
    
    // All critical agents failed - trigger recovery
    return o.initiateRecovery(ctx, event, failed)
}
```

## 🧪 Testing Error Scenarios

### **Error Injection for Testing**

Test error handling with controlled failures:

```go
func TestAgentErrorHandling(t *testing.T) {
    tests := []struct {
        name          string
        injectedError error
        expectedResult string
        shouldRecover bool
    }{
        {
            name:          "Network timeout",
            injectedError: &core.TimeoutError{Operation: "llm_call", Duration: time.Second * 30},
            expectedResult: "retry_scheduled",
            shouldRecover: true,
        },
        {
            name:          "Rate limit exceeded",
            injectedError: &core.RateLimitError{Provider: "azure", RetryAfter: time.Minute},
            expectedResult: "rate_limited",
            shouldRecover: true,
        },
        {
            name:          "Invalid input",
            injectedError: &core.ValidationError{Field: "query", Message: "empty query"},
            expectedResult: "validation_failed",
            shouldRecover: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create agent with error injection
            handler := &TestHandler{
                injectedError: tt.injectedError,
            }
            
            agent := core.NewAgent("test-agent", handler)
            
            // Run and verify error handling
            result, err := agent.Run(context.Background(), createTestEvent(), createTestState())
            
            if tt.shouldRecover {
                assert.NoError(t, err)
                assert.Contains(t, result.Data["status"], tt.expectedResult)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

### **Chaos Engineering**

Implement chaos testing for production resilience:

```go
func ChaosTestingMiddleware(chaosConfig *ChaosConfig) func(core.AgentHandler) core.AgentHandler {
    return func(next core.AgentHandler) core.AgentHandler {
        return core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            // Randomly inject failures based on configuration
            if chaosConfig.ShouldInjectFailure() {
                return core.AgentResult{}, chaosConfig.GetRandomError()
            }
            
            // Introduce artificial latency
            if delay := chaosConfig.GetRandomDelay(); delay > 0 {
                time.Sleep(delay)
            }
            
            return next.Run(ctx, event, state)
        })
    }
}
```

## 📋 Error Handling Best Practices

### **1. Error Classification**
- **Transient**: Network timeouts, rate limits, temporary service outages
- **Permanent**: Authentication failures, malformed requests, invalid configurations
- **Critical**: System panics, out-of-memory, disk full

### **2. User Experience**
- Provide clear, actionable error messages
- Avoid exposing internal implementation details
- Offer alternative actions when possible
- Include request IDs for support purposes

### **3. Logging and Monitoring**
- Log errors with sufficient context for debugging
- Use structured logging for easy parsing
- Set up alerts for error rate thresholds
- Track error patterns and trends

### **4. Recovery Strategies**
- Implement graceful degradation for non-critical failures
- Use circuit breakers to prevent cascade failures
- Provide fallback mechanisms for critical operations
- Cache successful responses for replay during failures

This comprehensive error handling approach ensures your AgentFlow applications remain resilient and user-friendly even when things go wrong.
