# Error Handling

Learn how to handle errors effectively in v1beta, with structured error types, retry strategies, and graceful degradation patterns.

---

## 🎯 Overview

AgenticGoKit v1beta provides comprehensive error handling:

- **Structured Errors** - Typed error codes with contextual details
- **Error Helpers** - Component-specific error creators with suggestions
- **Error Checking** - Type-safe error classification
- **Retry Logic** - Built-in retry strategies for transient failures
- **Circuit Breaker** - Fault tolerance for external dependencies
- **Graceful Degradation** - Fallback strategies when components fail

---

## 🏗️ Error Structure

### AgentError Type

All v1beta errors use the `AgentError` type:

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

// AgentError structure
type AgentError struct {
    Code       v1beta.ErrorCode           // Standardized error code
    Message    string                     // Human-readable message
    InnerError error                      // Wrapped underlying error
    Details    map[string]interface{}     // Additional context
}
```

### Error Codes

```go
// Configuration errors
v1beta.ErrConfigInvalid
v1beta.ErrConfigMissing
v1beta.ErrCodeConfigNotFound
v1beta.ErrCodeConfigParse
v1beta.ErrCodeConfigValidation

// LLM errors
v1beta.ErrCodeLLMNotConfigured
v1beta.ErrCodeLLMCallFailed
v1beta.ErrCodeLLMTimeout
v1beta.ErrCodeLLMRateLimited
v1beta.ErrCodeLLMAuth
v1beta.ErrCodeLLMQuotaExceeded
v1beta.ErrCodeLLMInvalidModel
v1beta.ErrCodeLLMConnection

// Tool errors
v1beta.ErrCodeToolNotFound
v1beta.ErrCodeToolExecute
v1beta.ErrCodeToolTimeout
v1beta.ErrCodeToolInvalidArgs
v1beta.ErrCodeToolNotAvailable

// Memory errors
v1beta.ErrCodeMemoryNotConfigured
v1beta.ErrCodeMemoryStore
v1beta.ErrCodeMemoryQuery
v1beta.ErrCodeMemoryConnection
v1beta.ErrCodeMemoryInvalidBackend

// MCP errors
v1beta.ErrCodeMCPServerNotFound
v1beta.ErrCodeMCPConnection
v1beta.ErrCodeMCPTimeout
v1beta.ErrCodeMCPInvalidResponse
v1beta.ErrCodeMCPServerUnhealthy

// Workflow errors
v1beta.ErrCodeWorkflowInvalid
v1beta.ErrCodeWorkflowStepFailed
v1beta.ErrCodeWorkflowTimeout
v1beta.ErrCodeWorkflowCycleDetected
v1beta.ErrCodeWorkflowMaxIterations

// Handler errors
v1beta.ErrCodeHandlerFailed
v1beta.ErrCodeHandlerTimeout
v1beta.ErrCodeHandlerPanic

// Runtime errors
v1beta.ErrCodeTimeout
v1beta.ErrCodeCancelled
v1beta.ErrCodeInternal
```

---

## 🚀 Quick Start

### Basic Error Handling

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    agent, err := v1beta.NewBuilder("MyAgent").
        WithLLM("openai", "gpt-4").
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    result, err := agent.Run(context.Background(), "Hello")
    if err != nil {
        // Check error type
        if v1beta.IsLLMError(err) {
            fmt.Println("LLM error:", err)
            fmt.Println("Suggestion:", v1beta.GetErrorSuggestion(err))
        } else {
            fmt.Println("Error:", err)
        }
        return
    }
    
    fmt.Println(result.Content)
}
```

---

## 🔍 Error Checking

### Type Checking Functions

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

// Check error component
if v1beta.IsLLMError(err) {
    // Handle LLM-specific errors
}

if v1beta.IsToolError(err) {
    // Handle tool-specific errors
}

if v1beta.IsMemoryError(err) {
    // Handle memory-specific errors
}

if v1beta.IsMCPError(err) {
    // Handle MCP-specific errors
}

if v1beta.IsWorkflowError(err) {
    // Handle workflow-specific errors
}

if v1beta.IsHandlerError(err) {
    // Handle handler-specific errors
}

if v1beta.IsConfigError(err) {
    // Handle configuration errors
}

// Check error properties
if v1beta.IsRetryable(err) {
    // Error can be retried
}

if v1beta.IsFatal(err) {
    // Error is fatal, don't retry
}
```

### Specific Error Codes

```go
// Check for specific error code
if v1beta.IsErrorCode(err, v1beta.ErrCodeLLMRateLimited) {
    fmt.Println("Rate limited! Waiting...")
    time.Sleep(5 * time.Second)
}

// Get error code
code := v1beta.GetErrorCode(err)
if code == v1beta.ErrCodeToolNotFound {
    fmt.Println("Tool missing")
}
```

### Extract Error Details

```go
// Get all details
details := v1beta.GetErrorDetails(err)
if details != nil {
    fmt.Printf("Component: %v\n", details["component"])
    fmt.Printf("Tool Name: %v\n", details["tool_name"])
}

// Get suggestion directly
suggestion := v1beta.GetErrorSuggestion(err)
if suggestion != "" {
    fmt.Println("Suggestion:", suggestion)
}

// Access AgentError directly
var agErr *v1beta.AgentError
if errors.As(err, &agErr) {
    fmt.Printf("Code: %s\n", agErr.Code)
    fmt.Printf("Message: %s\n", agErr.Message)
    fmt.Printf("Details: %+v\n", agErr.Details)
}
```

---

## 🔧 Creating Errors

### Component-Specific Errors

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

// LLM errors with suggestions
err := v1beta.LLMError(
    v1beta.ErrCodeLLMCallFailed,
    "Failed to call OpenAI API",
    originalErr,
)

// Tool errors with tool name
err := v1beta.ToolError(
    v1beta.ErrCodeToolNotFound,
    "calculator",
    "Tool not registered",
    nil,
)

// Memory errors
err := v1beta.MemoryError(
    v1beta.ErrCodeMemoryStore,
    "Failed to store conversation",
    originalErr,
)

// Workflow errors
err := v1beta.WorkflowError(
    v1beta.ErrCodeWorkflowStepFailed,
    "Step 2 execution failed",
    originalErr,
)

// Config errors
err := v1beta.ConfigError(
    v1beta.ErrCodeConfigNotFound,
    "Configuration file not found",
    nil,
)

// MCP errors with server name
err := v1beta.MCPError(
    v1beta.ErrCodeMCPConnection,
    "filesystem",
    "Failed to connect to MCP server",
    originalErr,
)

// Handler errors
err := v1beta.HandlerError(
    v1beta.ErrCodeHandlerFailed,
    "Custom handler returned error",
    originalErr,
)

// Validation errors with field context
err := v1beta.NewValidationError(
    "name",
    "Agent name is required",
)
```

### Generic Errors

```go
// Create basic AgentError
err := v1beta.NewAgentError(
    v1beta.ErrCodeInternal,
    "Something went wrong",
)

// With wrapped error
err := v1beta.NewAgentErrorWithError(
    v1beta.ErrCodeTimeout,
    "Operation timed out",
    originalErr,
)

// Add custom details
err.AddDetail("request_id", "req-123")
err.AddDetail("retry_count", 3)
```

---

## 🔄 Retry Strategies

### Basic Retry Pattern

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

func runWithRetry(agent v1beta.Agent, ctx context.Context, input string, maxRetries int) (*v1beta.Result, error) {
    var lastErr error
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        result, err := agent.Run(ctx, input)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // Don't retry if error is not retryable
        if !v1beta.IsRetryable(err) {
            return nil, err
        }
        
        // Don't retry if error is fatal
        if v1beta.IsFatal(err) {
            return nil, err
        }
        
        if attempt < maxRetries {
            // Exponential backoff
            backoff := time.Duration(attempt*attempt) * time.Second
            log.Printf("Attempt %d failed, retrying in %v...", attempt, backoff)
            time.Sleep(backoff)
        }
    }
    
    return nil, fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, lastErr)
}
```

### Configurable Retry

```go
type RetryConfig struct {
    MaxRetries  int
    InitialWait time.Duration
    MaxWait     time.Duration
    Multiplier  float64
}

func runWithConfigurableRetry(agent v1beta.Agent, ctx context.Context, input string, cfg RetryConfig) (*v1beta.Result, error) {
    var lastErr error
    wait := cfg.InitialWait
    
    for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
        result, err := agent.Run(ctx, input)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        if !v1beta.IsRetryable(err) || v1beta.IsFatal(err) {
            return nil, err
        }
        
        if attempt < cfg.MaxRetries {
            log.Printf("Retry attempt %d after %v", attempt, wait)
            
            select {
            case <-time.After(wait):
                // Continue to next attempt
            case <-ctx.Done():
                return nil, ctx.Err()
            }
            
            // Exponential backoff with max wait
            wait = time.Duration(float64(wait) * cfg.Multiplier)
            if wait > cfg.MaxWait {
                wait = cfg.MaxWait
            }
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Usage
config := RetryConfig{
    MaxRetries:  5,
    InitialWait: 1 * time.Second,
    MaxWait:     30 * time.Second,
    Multiplier:  2.0,
}
result, err := runWithConfigurableRetry(agent, ctx, input, config)
```

### Retry with Jitter

```go
import "math/rand"

func runWithJitter(agent v1beta.Agent, ctx context.Context, input string, maxRetries int) (*v1beta.Result, error) {
    var lastErr error
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        result, err := agent.Run(ctx, input)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        if !v1beta.IsRetryable(err) || v1beta.IsFatal(err) {
            return nil, err
        }
        
        if attempt < maxRetries {
            // Exponential backoff with jitter
            baseWait := time.Duration(attempt*attempt) * time.Second
            jitter := time.Duration(rand.Int63n(int64(time.Second)))
            wait := baseWait + jitter
            
            log.Printf("Retry %d after %v", attempt, wait)
            time.Sleep(wait)
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

---

## 🛡️ Circuit Breaker

### Built-in Circuit Breaker

Configure circuit breaker in TOML:

```toml
[tools.circuit_breaker]
enabled = true
failure_threshold = 5        # Open after 5 consecutive failures
success_threshold = 2        # Close after 2 consecutive successes
timeout = "60s"              # Time before attempting to half-open
half_open_max_calls = 3      # Max calls in half-open state
```

### Custom Circuit Breaker Pattern

```go
type CircuitBreaker struct {
    maxFailures     int
    resetTimeout    time.Duration
    failureCount    int
    lastFailureTime time.Time
    state           string // "closed", "open", "half-open"
    mu              sync.Mutex
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:        "closed",
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    
    // Check if we should transition from open to half-open
    if cb.state == "open" && time.Since(cb.lastFailureTime) > cb.resetTimeout {
        cb.state = "half-open"
        cb.failureCount = 0
    }
    
    // Reject calls if circuit is open
    if cb.state == "open" {
        cb.mu.Unlock()
        return fmt.Errorf("circuit breaker is open")
    }
    
    cb.mu.Unlock()
    
    // Execute function
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failureCount++
        cb.lastFailureTime = time.Now()
        
        if cb.failureCount >= cb.maxFailures {
            cb.state = "open"
        }
        
        return err
    }
    
    // Success - reset or close circuit
    if cb.state == "half-open" {
        cb.state = "closed"
    }
    cb.failureCount = 0
    
    return nil
}

// Usage
cb := NewCircuitBreaker(5, 60*time.Second)

for i := 0; i < 10; i++ {
    err := cb.Call(func() error {
        result, err := agent.Run(ctx, input)
        if err != nil {
            return err
        }
        fmt.Println(result.Content)
        return nil
    })
    
    if err != nil {
        fmt.Println("Circuit breaker error:", err)
    }
    
    time.Sleep(time.Second)
}
```

---

## 🎯 Graceful Degradation

### Pattern 1: LLM Fallback

Fall back to simpler model when primary fails:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Try primary LLM
    response, err := capabilities.LLM("You are a helpful assistant.", input)
    if err != nil {
        // Check if retryable
        if v1beta.IsRetryable(err) {
            // Try once more
            response, err = capabilities.LLM("You are a helpful assistant.", input)
            if err == nil {
                return response, nil
            }
        }
        
        // Fall back to simpler template-based response
        log.Printf("LLM failed, using fallback: %v", err)
        return fmt.Sprintf("I received your message: %s. However, I'm experiencing technical difficulties.", input), nil
    }
    
    return response, nil
}
```

### Pattern 2: Tool Fallback Chain

Try multiple tools in order:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Try primary tool
    result, err := capabilities.Tools.Execute(ctx, "web_search", map[string]interface{}{
        "query": input,
    })
    
    if err == nil && result.Success {
        return fmt.Sprintf("%v", result.Content), nil
    }
    
    // Try fallback tool
    log.Printf("Primary tool failed, trying fallback")
    result, err = capabilities.Tools.Execute(ctx, "local_search", map[string]interface{}{
        "query": input,
    })
    
    if err == nil && result.Success {
        return fmt.Sprintf("%v", result.Content), nil
    }
    
    // Use LLM as last resort
    log.Printf("All tools failed, using LLM")
    return capabilities.LLM(
        "Answer based on your general knowledge.",
        input,
    )
}
```

### Pattern 3: Partial Success

Return partial results when some operations fail:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    var results []string
    var errors []error
    
    // Try multiple sources
    sources := []string{"source1", "source2", "source3"}
    
    for _, source := range sources {
        result, err := capabilities.Tools.Execute(ctx, source, map[string]interface{}{
            "query": input,
        })
        
        if err != nil {
            errors = append(errors, err)
            log.Printf("Source %s failed: %v", source, err)
            continue
        }
        
        if result.Success {
            results = append(results, fmt.Sprintf("%v", result.Content))
        }
    }
    
    // Return partial results if at least one succeeded
    if len(results) > 0 {
        combined := strings.Join(results, "\n\n")
        if len(errors) > 0 {
            combined += fmt.Sprintf("\n\nNote: %d of %d sources failed", len(errors), len(sources))
        }
        return combined, nil
    }
    
    // All sources failed
    return "", fmt.Errorf("all sources failed: %d errors", len(errors))
}
```

### Pattern 4: Cached Fallback

Use cached data when live data unavailable:

```go
type CachedHandler struct {
    cache map[string]string
    mu    sync.RWMutex
}

func (h *CachedHandler) Handle(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Try live data
    result, err := capabilities.Tools.Execute(ctx, "live_data", map[string]interface{}{
        "query": input,
    })
    
    if err == nil && result.Success {
        content := fmt.Sprintf("%v", result.Content)
        
        // Update cache
        h.mu.Lock()
        h.cache[input] = content
        h.mu.Unlock()
        
        return content, nil
    }
    
    // Fall back to cache
    log.Printf("Live data failed, checking cache: %v", err)
    h.mu.RLock()
    cached, exists := h.cache[input]
    h.mu.RUnlock()
    
    if exists {
        return cached + " (cached)", nil
    }
    
    // No cache available
    return "", fmt.Errorf("no live or cached data available: %w", err)
}
```

---

## 📊 Error Logging and Monitoring

### Structured Error Logging

```go
import "log/slog"

func logError(err error) {
    var agErr *v1beta.AgentError
    if errors.As(err, &agErr) {
        slog.Error("Agent error occurred",
            "code", agErr.Code,
            "message", agErr.Message,
            "component", agErr.Details["component"],
            "suggestion", agErr.Details["suggestion"],
        )
        
        if agErr.InnerError != nil {
            slog.Error("Caused by", "error", agErr.InnerError)
        }
    } else {
        slog.Error("Unknown error", "error", err)
    }
}
```

### Error Metrics

```go
type ErrorMetrics struct {
    totalErrors    int64
    errorsByCode   map[v1beta.ErrorCode]int64
    errorsByType   map[string]int64
    retryableCount int64
    fatalCount     int64
    mu             sync.RWMutex
}

func (m *ErrorMetrics) RecordError(err error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.totalErrors++
    
    code := v1beta.GetErrorCode(err)
    if code != "" {
        m.errorsByCode[code]++
    }
    
    switch {
    case v1beta.IsLLMError(err):
        m.errorsByType["llm"]++
    case v1beta.IsToolError(err):
        m.errorsByType["tool"]++
    case v1beta.IsMemoryError(err):
        m.errorsByType["memory"]++
    case v1beta.IsWorkflowError(err):
        m.errorsByType["workflow"]++
    }
    
    if v1beta.IsRetryable(err) {
        m.retryableCount++
    }
    
    if v1beta.IsFatal(err) {
        m.fatalCount++
    }
}

func (m *ErrorMetrics) Report() {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    fmt.Printf("Total Errors: %d\n", m.totalErrors)
    fmt.Printf("Retryable: %d (%.1f%%)\n", 
        m.retryableCount, 
        float64(m.retryableCount)/float64(m.totalErrors)*100)
    fmt.Printf("Fatal: %d (%.1f%%)\n", 
        m.fatalCount, 
        float64(m.fatalCount)/float64(m.totalErrors)*100)
    
    fmt.Println("\nErrors by Type:")
    for typ, count := range m.errorsByType {
        fmt.Printf("  %s: %d\n", typ, count)
    }
    
    fmt.Println("\nTop Error Codes:")
    for code, count := range m.errorsByCode {
        fmt.Printf("  %s: %d\n", code, count)
    }
}
```

---

## 🐛 Common Errors and Solutions

### LLM Configuration Error

**Error:** `LLM_NOT_CONFIGURED`

**Cause:** No LLM provider configured

**Solution:**
```go
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").  // Add this
    Build()
```

### Rate Limiting Error

**Error:** `LLM_RATE_LIMITED`

**Cause:** Exceeded API rate limits

**Solution:** Implement retry with backoff
```go
if v1beta.IsErrorCode(err, v1beta.ErrCodeLLMRateLimited) {
    time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
    // Retry operation
}
```

### Tool Not Found Error

**Error:** `TOOL_NOT_FOUND`

**Cause:** Tool not registered

**Solution:**
```go
agent, _ := v1beta.NewBuilder("Agent").
    WithTools(
        v1beta.WithMCP(servers...),  // Add tools
    ).
    Build()
```

### Memory Connection Error

**Error:** `MEMORY_CONNECTION`

**Cause:** Cannot connect to memory backend

**Solution:** Check connection string and backend availability
```toml
[memory]
provider = "pgvector"
connection_string = "postgresql://user:pass@localhost:5432/db"
```

### MCP Server Unhealthy

**Error:** `MCP_SERVER_UNHEALTHY`

**Cause:** MCP server not responding

**Solution:** Check server health
```go
health := capabilities.Tools.HealthCheck(ctx)
for name, status := range health {
    if status.Status != "healthy" {
        log.Printf("Server %s is unhealthy: %s", name, status.Error)
    }
}
```

---

## 🎯 Best Practices

### 1. Always Check Error Types

```go
if err != nil {
    switch {
    case v1beta.IsLLMError(err):
        // Handle LLM errors
    case v1beta.IsToolError(err):
        // Handle tool errors
    case v1beta.IsMemoryError(err):
        // Handle memory errors
    default:
        // Handle unknown errors
    }
}
```

### 2. Use Suggestions

```go
if err != nil {
    log.Printf("Error: %v", err)
    if suggestion := v1beta.GetErrorSuggestion(err); suggestion != "" {
        log.Printf("Suggestion: %s", suggestion)
    }
}
```

### 3. Implement Proper Retry Logic

```go
// Don't retry fatal errors
if v1beta.IsFatal(err) {
    return err
}

// Only retry retryable errors
if v1beta.IsRetryable(err) {
    return retryOperation()
}
```

### 4. Monitor Error Patterns

Track error frequency to identify systemic issues:
- High rate of `LLM_RATE_LIMITED`: Implement better rate limiting
- Frequent `TOOL_TIMEOUT`: Increase timeouts or optimize tools
- Many `MEMORY_CONNECTION`: Check backend health

### 5. Provide Fallbacks

Always have a fallback strategy for critical operations:
- LLM failures → simpler responses or cached data
- Tool failures → alternative tools or manual processes
- Memory failures → stateless operation mode

---

## 📚 Next Steps

- **[Performance](./performance.md)** - Optimization strategies
- **[Troubleshooting](./troubleshooting.md)** - Common issues and solutions
- **[Custom Handlers](./custom-handlers.md)** - Implement custom error handling

---

**Ready to optimize performance?** Continue to [Performance](./performance.md) →
