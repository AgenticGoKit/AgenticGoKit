# Error Handling

Structured error handling with type checking, retries, and graceful degradation patterns.

---

## Error Structure

All v1beta errors use `AgentError` type:

```go
type AgentError struct {
    Code       ErrorCode              // Standardized error code
    Message    string                 // Human-readable message
    InnerError error                  // Wrapped underlying error
    Details    map[string]interface{} // Additional context
}
```


---

## Error Codes

Key error codes by component:

```go
// Configuration
ErrCodeConfigNotFound, ErrCodeConfigParse, ErrCodeConfigValidation

// LLM
ErrCodeLLMNotConfigured, ErrCodeLLMCallFailed, ErrCodeLLMTimeout,
ErrCodeLLMRateLimited, ErrCodeLLMAuth, ErrCodeLLMQuotaExceeded

// Tools
ErrCodeToolNotFound, ErrCodeToolExecute, ErrCodeToolTimeout, ErrCodeToolInvalidArgs

// Memory
ErrCodeMemoryNotConfigured, ErrCodeMemoryStore, ErrCodeMemoryQuery, ErrCodeMemoryConnection

// Workflows
ErrCodeWorkflowInvalid, ErrCodeWorkflowStepFailed, ErrCodeWorkflowTimeout, ErrCodeWorkflowCycleDetected

// MCP
ErrCodeMCPServerNotFound, ErrCodeMCPConnection, ErrCodeMCPTimeout

// Runtime
ErrCodeTimeout, ErrCodeCancelled, ErrCodeInternal
```

---

## Quick Start

### Basic Error Handling

```go
result, err := agent.Run(context.Background(), "Hello")
if err != nil {
    if v1beta.IsLLMError(err) {
        log.Printf("LLM error: %s", v1beta.GetErrorSuggestion(err))
        return
    }
    log.Fatal(err)
}
```

### Check Error Type

```go
// Check component
if v1beta.IsLLMError(err) { }
if v1beta.IsToolError(err) { }
if v1beta.IsMemoryError(err) { }
if v1beta.IsWorkflowError(err) { }
if v1beta.IsMCPError(err) { }

// Check recoverability
if v1beta.IsRetryable(err) { }     // Can be retried
if v1beta.IsFatal(err) { }         // Do not retry

// Check specific code
if v1beta.IsErrorCode(err, v1beta.ErrCodeLLMRateLimited) { }
```

### Extract Error Details

```go
code := v1beta.GetErrorCode(err)
details := v1beta.GetErrorDetails(err)
suggestion := v1beta.GetErrorSuggestion(err)

var agErr *v1beta.AgentError
if errors.As(err, &agErr) {
    log.Printf("Code: %s, Message: %s, Details: %+v", 
        agErr.Code, agErr.Message, agErr.Details)
}
```

---

## Creating Errors

### Component-Specific Error Creators

```go
// LLM errors
err := v1beta.LLMError(v1beta.ErrCodeLLMCallFailed, "OpenAI call failed", originalErr)

// Tool errors
err := v1beta.ToolError(v1beta.ErrCodeToolNotFound, "calculator", "Tool not registered", nil)

// Memory errors
err := v1beta.MemoryError(v1beta.ErrCodeMemoryStore, "Failed to store", originalErr)

// Workflow errors
err := v1beta.WorkflowError(v1beta.ErrCodeWorkflowStepFailed, "Step failed", originalErr)

// Config errors
err := v1beta.ConfigError(v1beta.ErrCodeConfigNotFound, "Config file missing", nil)

// MCP errors
err := v1beta.MCPError(v1beta.ErrCodeMCPConnection, "filesystem", "Connection failed", originalErr)

// Handler errors
err := v1beta.HandlerError(v1beta.ErrCodeHandlerFailed, "Custom handler error", originalErr)

// Validation errors
err := v1beta.NewValidationError("name", "Agent name is required")
```

---

## Retry Patterns

### Basic Retry with Backoff

```go
func runWithRetry(agent v1beta.Agent, ctx context.Context, input string, maxRetries int) (*v1beta.Result, error) {
    var lastErr error
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        result, err := agent.Run(ctx, input)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // Don't retry fatal errors
        if v1beta.IsFatal(err) {
            return nil, err
        }
        
        // Only retry if retryable
        if !v1beta.IsRetryable(err) {
            return nil, err
        }
        
        if attempt < maxRetries {
            backoff := time.Duration(attempt*attempt) * time.Second
            log.Printf("Attempt %d failed, retrying in %v...", attempt, backoff)
            time.Sleep(backoff)
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

### Retry with Jitter

```go
func runWithJitter(agent v1beta.Agent, ctx context.Context, input string, maxRetries int) (*v1beta.Result, error) {
    for attempt := 1; attempt <= maxRetries; attempt++ {
        result, err := agent.Run(ctx, input)
        if err == nil {
            return result, nil
        }
        
        if v1beta.IsFatal(err) || !v1beta.IsRetryable(err) {
            return nil, err
        }
        
        if attempt < maxRetries {
            baseWait := time.Duration(attempt*attempt) * time.Second
            jitter := time.Duration(rand.Int63n(int64(time.Second)))
            time.Sleep(baseWait + jitter)
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded")
}
```

### Rate Limit Specific Handling

```go
if v1beta.IsErrorCode(err, v1beta.ErrCodeLLMRateLimited) {
    // Exponential backoff for rate limits
    waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
    log.Printf("Rate limited, waiting %v", waitTime)
    time.Sleep(waitTime)
    continue
}
```


---

## Graceful Degradation

### LLM Fallback

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    response, err := capabilities.LLM("You are helpful.", input)
    if err == nil {
        return response, nil
    }
    
    // Fallback to template
    log.Printf("LLM failed, using fallback: %v", err)
    return fmt.Sprintf("I received: %s. Technical difficulties encountered.", input), nil
}
```

### Tool Fallback Chain

```go
result, err := capabilities.Tools.Execute(ctx, "primary_tool", args)
if err != nil {
    log.Printf("Primary tool failed, trying fallback")
    result, err = capabilities.Tools.Execute(ctx, "fallback_tool", args)
    if err != nil {
        return capabilities.LLM("Answer from knowledge.", input)
    }
}
```

### Partial Success

```go
var results []string
var errors []error

for _, source := range sources {
    result, err := capabilities.Tools.Execute(ctx, source, args)
    if err != nil {
        errors = append(errors, err)
        continue
    }
    results = append(results, fmt.Sprintf("%v", result.Content))
}

if len(results) > 0 {
    output := strings.Join(results, "\n")
    if len(errors) > 0 {
        output += fmt.Sprintf("\nNote: %d sources failed", len(errors))
    }
    return output, nil
}
```

---

## Logging and Monitoring

### Structured Logging

```go
import "log/slog"

var agErr *v1beta.AgentError
if errors.As(err, &agErr) {
    slog.Error("Agent error",
        "code", agErr.Code,
        "message", agErr.Message,
        "suggestion", v1beta.GetErrorSuggestion(err),
    )
}
```

### Error Tracking

```go
type ErrorMetrics struct {
    total          int64
    retryable      int64
    fatal          int64
    byCode         map[string]int64
    mu             sync.RWMutex
}

func (m *ErrorMetrics) Record(err error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.total++
    
    if v1beta.IsRetryable(err) {
        m.retryable++
    } else if v1beta.IsFatal(err) {
        m.fatal++
    }
    
    code := v1beta.GetErrorCode(err)
    if code != "" {
        m.byCode[string(code)]++
    }
}
```

---

## Common Errors and Solutions

| Error | Cause | Solution |
|-------|-------|----------|
| `LLM_NOT_CONFIGURED` | No LLM provider | Use `WithLLM("openai", "gpt-4")` |
| `LLM_RATE_LIMITED` | Rate limit exceeded | Implement exponential backoff retry |
| `TOOL_NOT_FOUND` | Tool not registered | Use `WithTools(WithMCP(servers...))` |
| `MEMORY_CONNECTION` | Backend unavailable | Check connection string and backend health |
| `MCP_SERVER_UNHEALTHY` | MCP server down | Verify server is running and responsive |
| `WORKFLOW_CYCLE_DETECTED` | Circular dependencies | Ensure workflow steps form a DAG |

---

## Best Practices

1. **Always check error types** - Use `IsLLMError()`, `IsToolError()` etc.
2. **Use suggestions** - Call `GetErrorSuggestion(err)` for debugging
3. **Only retry retryable errors** - Check `IsRetryable()` and `IsFatal()`
4. **Monitor patterns** - Track error frequency to identify systemic issues
5. **Implement fallbacks** - Have strategies for critical failures
6. **Add context** - Use `GetErrorDetails()` for rich error information

---

**Next:** [Custom Handlers](./custom-handlers.md) → [Troubleshooting](./troubleshooting.md)
