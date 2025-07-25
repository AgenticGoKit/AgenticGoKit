# Advanced Tool Patterns in AgenticGoKit

## Overview

This tutorial explores advanced patterns for tool usage in AgenticGoKit, including tool composition, conditional execution, parallel tool usage, and sophisticated error handling strategies. These patterns enable you to build complex, production-ready agent systems that can handle sophisticated workflows and edge cases.

## Prerequisites

- Understanding of [MCP Overview](README.md)
- Completion of [Tool Development](tool-development.md)
- Familiarity with [Tool Integration](tool-integration.md)
- Knowledge of [Orchestration Patterns](../orchestration/orchestration-patterns.md)

## Tool Composition Patterns

### 1. Sequential Tool Chains

```go
package patterns

import (
    "context"
    "fmt"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

// ToolChain executes tools in sequence, passing results between them
type ToolChain struct {
    name  string
    steps []ToolStep
}

type ToolStep struct {
    ToolName    string
    ParamMapper func(previousResult interface{}, initialParams map[string]interface{}) map[string]interface{}
}

func NewToolChain(name string) *ToolChain {
    return &ToolChain{
        name:  name,
        steps: make([]ToolStep, 0),
    }
}

func (tc *ToolChain) AddStep(toolName string, paramMapper func(interface{}, map[string]interface{}) map[string]interface{}) *ToolChain {
    tc.steps = append(tc.steps, ToolStep{
        ToolName:    toolName,
        ParamMapper: paramMapper,
    })
    return tc
}

func (tc *ToolChain) Name() string {
    return tc.name
}

func (tc *ToolChain) Description() string {
    return fmt.Sprintf("Sequential tool chain with %d steps", len(tc.steps))
}

func (tc *ToolChain) ParameterSchema() map[string]core.ParameterDefinition {
    return map[string]core.ParameterDefinition{
        "initial_params": {
            Type:        "object",
            Description: "Initial parameters for the tool chain",
            Required:    true,
        },
        "mcp_manager": {
            Type:        "object",
            Description: "MCP manager for tool execution",
            Required:    true,
        },
    }
}

func (tc *ToolChain) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    initialParams, ok := params["initial_params"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("initial_params must be an object")
    }
    
    mcpManager, ok := params["mcp_manager"].(*core.MCPManager)
    if !ok {
        return nil, fmt.Errorf("mcp_manager must be provided")
    }
    
    var currentResult interface{} = initialParams
    results := make([]interface{}, 0, len(tc.steps))
    
    for i, step := range tc.steps {
        // Get tool
        tool, err := mcpManager.GetTool(step.ToolName)
        if err != nil {
            return nil, fmt.Errorf("step %d: failed to get tool %s: %w", i+1, step.ToolName, err)
        }
        
        // Map parameters
        var toolParams map[string]interface{}
        if step.ParamMapper != nil {
            toolParams = step.ParamMapper(currentResult, initialParams)
        } else {
            toolParams = initialParams
        }
        
        // Execute tool
        result, err := tool.Execute(ctx, toolParams)
        if err != nil {
            return nil, fmt.Errorf("step %d: tool %s execution failed: %w", i+1, step.ToolName, err)
        }
        
        currentResult = result
        results = append(results, result)
    }
    
    return map[string]interface{}{
        "final_result": currentResult,
        "all_results":  results,
        "steps":        len(tc.steps),
    }, nil
}

// Example: Research and Analysis Chain
func CreateResearchChain() *ToolChain {
    return NewToolChain("research_analysis").
        AddStep("search", func(prev interface{}, initial map[string]interface{}) map[string]interface{} {
            query, _ := initial["query"].(string)
            return map[string]interface{}{
                "query": query,
                "limit": 5,
            }
        }).
        AddStep("summarizer", func(prev interface{}, initial map[string]interface{}) map[string]interface{} {
            searchResults := prev.(map[string]interface{})
            return map[string]interface{}{
                "content": searchResults["results"],
                "max_length": 500,
            }
        }).
        AddStep("analyzer", func(prev interface{}, initial map[string]interface{}) map[string]interface{} {
            summary := prev.(map[string]interface{})
            return map[string]interface{}{
                "text": summary["summary"],
                "analysis_type": "sentiment_and_topics",
            }
        })
}
```### 2.
 Parallel Tool Execution

```go
// ParallelToolExecutor runs multiple tools concurrently
type ParallelToolExecutor struct {
    name        string
    tools       []ParallelToolConfig
    aggregator  ResultAggregator
}

type ParallelToolConfig struct {
    ToolName   string
    Parameters map[string]interface{}
    Optional   bool // If true, failure won't fail the entire execution
}

type ResultAggregator func(results map[string]interface{}, errors map[string]error) (interface{}, error)

func NewParallelToolExecutor(name string, aggregator ResultAggregator) *ParallelToolExecutor {
    return &ParallelToolExecutor{
        name:       name,
        tools:      make([]ParallelToolConfig, 0),
        aggregator: aggregator,
    }
}

func (pte *ParallelToolExecutor) AddTool(toolName string, params map[string]interface{}, optional bool) *ParallelToolExecutor {
    pte.tools = append(pte.tools, ParallelToolConfig{
        ToolName:   toolName,
        Parameters: params,
        Optional:   optional,
    })
    return pte
}

func (pte *ParallelToolExecutor) Name() string {
    return pte.name
}

func (pte *ParallelToolExecutor) Description() string {
    return fmt.Sprintf("Parallel execution of %d tools", len(pte.tools))
}

func (pte *ParallelToolExecutor) ParameterSchema() map[string]core.ParameterDefinition {
    return map[string]core.ParameterDefinition{
        "mcp_manager": {
            Type:        "object",
            Description: "MCP manager for tool execution",
            Required:    true,
        },
        "timeout": {
            Type:        "number",
            Description: "Timeout in seconds for parallel execution",
            Required:    false,
            Default:     30,
        },
    }
}

func (pte *ParallelToolExecutor) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    mcpManager, ok := params["mcp_manager"].(*core.MCPManager)
    if !ok {
        return nil, fmt.Errorf("mcp_manager must be provided")
    }
    
    timeout := 30 * time.Second
    if timeoutParam, ok := params["timeout"].(float64); ok {
        timeout = time.Duration(timeoutParam) * time.Second
    }
    
    // Create context with timeout
    execCtx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    // Execute tools in parallel
    results := make(map[string]interface{})
    errors := make(map[string]error)
    
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for _, toolConfig := range pte.tools {
        wg.Add(1)
        go func(config ParallelToolConfig) {
            defer wg.Done()
            
            tool, err := mcpManager.GetTool(config.ToolName)
            if err != nil {
                mu.Lock()
                errors[config.ToolName] = fmt.Errorf("failed to get tool: %w", err)
                mu.Unlock()
                return
            }
            
            result, err := tool.Execute(execCtx, config.Parameters)
            
            mu.Lock()
            if err != nil {
                errors[config.ToolName] = err
            } else {
                results[config.ToolName] = result
            }
            mu.Unlock()
        }(toolConfig)
    }
    
    wg.Wait()
    
    // Check for required tool failures
    for _, toolConfig := range pte.tools {
        if !toolConfig.Optional {
            if err, exists := errors[toolConfig.ToolName]; exists {
                return nil, fmt.Errorf("required tool %s failed: %w", toolConfig.ToolName, err)
            }
        }
    }
    
    // Aggregate results
    return pte.aggregator(results, errors)
}

// Example: Multi-Source Information Gathering
func CreateInfoGatheringTool() *ParallelToolExecutor {
    return NewParallelToolExecutor("info_gathering", func(results map[string]interface{}, errors map[string]error) (interface{}, error) {
        gathered := map[string]interface{}{
            "sources": make(map[string]interface{}),
            "errors":  make(map[string]string),
            "summary": "",
        }
        
        // Collect successful results
        for toolName, result := range results {
            gathered["sources"].(map[string]interface{})[toolName] = result
        }
        
        // Collect errors for optional tools
        for toolName, err := range errors {
            gathered["errors"].(map[string]string)[toolName] = err.Error()
        }
        
        // Create summary
        sourceCount := len(results)
        errorCount := len(errors)
        gathered["summary"] = fmt.Sprintf("Gathered information from %d sources with %d errors", sourceCount, errorCount)
        
        return gathered, nil
    }).
    AddTool("web_search", map[string]interface{}{"query": "latest news"}, false).
    AddTool("weather", map[string]interface{}{"location": "current"}, true).
    AddTool("stock_prices", map[string]interface{}{"symbols": []string{"AAPL", "GOOGL"}}, true).
    AddTool("calendar", map[string]interface{}{"days": 7}, true)
}
```

## Conditional Tool Execution

### 1. Rule-Based Tool Selection

```go
// ConditionalToolExecutor executes tools based on conditions
type ConditionalToolExecutor struct {
    name  string
    rules []ExecutionRule
}

type ExecutionRule struct {
    Condition   func(context.Context, map[string]interface{}) bool
    ToolName    string
    Parameters  func(map[string]interface{}) map[string]interface{}
    Description string
}

func NewConditionalToolExecutor(name string) *ConditionalToolExecutor {
    return &ConditionalToolExecutor{
        name:  name,
        rules: make([]ExecutionRule, 0),
    }
}

func (cte *ConditionalToolExecutor) AddRule(
    condition func(context.Context, map[string]interface{}) bool,
    toolName string,
    paramMapper func(map[string]interface{}) map[string]interface{},
    description string,
) *ConditionalToolExecutor {
    cte.rules = append(cte.rules, ExecutionRule{
        Condition:   condition,
        ToolName:    toolName,
        Parameters:  paramMapper,
        Description: description,
    })
    return cte
}

func (cte *ConditionalToolExecutor) Name() string {
    return cte.name
}

func (cte *ConditionalToolExecutor) Description() string {
    return fmt.Sprintf("Conditional tool executor with %d rules", len(cte.rules))
}

func (cte *ConditionalToolExecutor) ParameterSchema() map[string]core.ParameterDefinition {
    return map[string]core.ParameterDefinition{
        "input_data": {
            Type:        "object",
            Description: "Input data for condition evaluation",
            Required:    true,
        },
        "mcp_manager": {
            Type:        "object",
            Description: "MCP manager for tool execution",
            Required:    true,
        },
    }
}

func (cte *ConditionalToolExecutor) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    inputData, ok := params["input_data"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("input_data must be an object")
    }
    
    mcpManager, ok := params["mcp_manager"].(*core.MCPManager)
    if !ok {
        return nil, fmt.Errorf("mcp_manager must be provided")
    }
    
    executedRules := make([]map[string]interface{}, 0)
    
    for i, rule := range cte.rules {
        if rule.Condition(ctx, inputData) {
            // Get tool
            tool, err := mcpManager.GetTool(rule.ToolName)
            if err != nil {
                return nil, fmt.Errorf("rule %d: failed to get tool %s: %w", i+1, rule.ToolName, err)
            }
            
            // Prepare parameters
            toolParams := rule.Parameters(inputData)
            
            // Execute tool
            result, err := tool.Execute(ctx, toolParams)
            if err != nil {
                return nil, fmt.Errorf("rule %d: tool %s execution failed: %w", i+1, rule.ToolName, err)
            }
            
            executedRules = append(executedRules, map[string]interface{}{
                "rule_description": rule.Description,
                "tool_name":        rule.ToolName,
                "result":           result,
            })
        }
    }
    
    return map[string]interface{}{
        "executed_rules": executedRules,
        "total_rules":    len(cte.rules),
        "matched_rules":  len(executedRules),
    }, nil
}

// Example: Smart Assistant Tool Selection
func CreateSmartAssistantTool() *ConditionalToolExecutor {
    return NewConditionalToolExecutor("smart_assistant").
        AddRule(
            func(ctx context.Context, data map[string]interface{}) bool {
                message, ok := data["message"].(string)
                return ok && strings.Contains(strings.ToLower(message), "weather")
            },
            "weather",
            func(data map[string]interface{}) map[string]interface{} {
                location := "current"
                if loc, ok := data["location"].(string); ok {
                    location = loc
                }
                return map[string]interface{}{"location": location}
            },
            "Weather information requested",
        ).
        AddRule(
            func(ctx context.Context, data map[string]interface{}) bool {
                message, ok := data["message"].(string)
                if !ok {
                    return false
                }
                mathKeywords := []string{"calculate", "math", "+", "-", "*", "/", "="}
                msgLower := strings.ToLower(message)
                for _, keyword := range mathKeywords {
                    if strings.Contains(msgLower, keyword) {
                        return true
                    }
                }
                return false
            },
            "calculator",
            func(data map[string]interface{}) map[string]interface{} {
                // Simple math expression parser would go here
                return map[string]interface{}{
                    "expression": data["message"],
                }
            },
            "Mathematical calculation requested",
        ).
        AddRule(
            func(ctx context.Context, data map[string]interface{}) bool {
                message, ok := data["message"].(string)
                return ok && strings.Contains(strings.ToLower(message), "search")
            },
            "web_search",
            func(data map[string]interface{}) map[string]interface{} {
                return map[string]interface{}{
                    "query": data["message"],
                    "limit": 5,
                }
            },
            "Web search requested",
        )
}
```## 
Error Handling and Recovery Patterns

### 1. Retry with Backoff

```go
// RetryableTool wraps a tool with retry logic
type RetryableTool struct {
    tool         core.Tool
    maxRetries   int
    baseDelay    time.Duration
    maxDelay     time.Duration
    backoffFunc  func(attempt int, baseDelay time.Duration) time.Duration
    retryChecker func(error) bool
}

func NewRetryableTool(tool core.Tool, maxRetries int, baseDelay time.Duration) *RetryableTool {
    return &RetryableTool{
        tool:       tool,
        maxRetries: maxRetries,
        baseDelay:  baseDelay,
        maxDelay:   5 * time.Minute,
        backoffFunc: func(attempt int, baseDelay time.Duration) time.Duration {
            // Exponential backoff with jitter
            delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
            jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
            return delay + jitter
        },
        retryChecker: func(err error) bool {
            // Retry on network errors, timeouts, and rate limits
            errStr := strings.ToLower(err.Error())
            return strings.Contains(errStr, "timeout") ||
                   strings.Contains(errStr, "network") ||
                   strings.Contains(errStr, "rate limit") ||
                   strings.Contains(errStr, "temporary")
        },
    }
}

func (rt *RetryableTool) WithMaxDelay(maxDelay time.Duration) *RetryableTool {
    rt.maxDelay = maxDelay
    return rt
}

func (rt *RetryableTool) WithRetryChecker(checker func(error) bool) *RetryableTool {
    rt.retryChecker = checker
    return rt
}

func (rt *RetryableTool) Name() string {
    return rt.tool.Name()
}

func (rt *RetryableTool) Description() string {
    return rt.tool.Description()
}

func (rt *RetryableTool) ParameterSchema() map[string]core.ParameterDefinition {
    return rt.tool.ParameterSchema()
}

func (rt *RetryableTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    var lastErr error
    
    for attempt := 0; attempt <= rt.maxRetries; attempt++ {
        result, err := rt.tool.Execute(ctx, params)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // Check if we should retry
        if !rt.retryChecker(err) {
            return nil, fmt.Errorf("non-retryable error: %w", err)
        }
        
        // Don't wait after the last attempt
        if attempt == rt.maxRetries {
            break
        }
        
        // Calculate delay
        delay := rt.backoffFunc(attempt, rt.baseDelay)
        if delay > rt.maxDelay {
            delay = rt.maxDelay
        }
        
        // Wait with context cancellation support
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(delay):
            // Continue to next attempt
        }
    }
    
    return nil, fmt.Errorf("max retries (%d) exceeded, last error: %w", rt.maxRetries, lastErr)
}
```

### 2. Circuit Breaker Pattern

```go
// CircuitBreakerTool implements circuit breaker pattern for tools
type CircuitBreakerTool struct {
    tool            core.Tool
    failureThreshold int
    resetTimeout     time.Duration
    state           CircuitState
    failures        int
    lastFailureTime time.Time
    mu              sync.RWMutex
}

type CircuitState int

const (
    CircuitClosed CircuitState = iota
    CircuitOpen
    CircuitHalfOpen
)

func NewCircuitBreakerTool(tool core.Tool, failureThreshold int, resetTimeout time.Duration) *CircuitBreakerTool {
    return &CircuitBreakerTool{
        tool:             tool,
        failureThreshold: failureThreshold,
        resetTimeout:     resetTimeout,
        state:            CircuitClosed,
    }
}

func (cbt *CircuitBreakerTool) Name() string {
    return cbt.tool.Name()
}

func (cbt *CircuitBreakerTool) Description() string {
    return cbt.tool.Description()
}

func (cbt *CircuitBreakerTool) ParameterSchema() map[string]core.ParameterDefinition {
    return cbt.tool.ParameterSchema()
}

func (cbt *CircuitBreakerTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    cbt.mu.Lock()
    
    // Check if circuit should be reset
    if cbt.state == CircuitOpen && time.Since(cbt.lastFailureTime) > cbt.resetTimeout {
        cbt.state = CircuitHalfOpen
        cbt.failures = 0
    }
    
    // Fail fast if circuit is open
    if cbt.state == CircuitOpen {
        cbt.mu.Unlock()
        return nil, fmt.Errorf("circuit breaker is open for tool %s", cbt.tool.Name())
    }
    
    cbt.mu.Unlock()
    
    // Execute tool
    result, err := cbt.tool.Execute(ctx, params)
    
    cbt.mu.Lock()
    defer cbt.mu.Unlock()
    
    if err != nil {
        cbt.failures++
        cbt.lastFailureTime = time.Now()
        
        // Open circuit if threshold exceeded
        if cbt.failures >= cbt.failureThreshold {
            cbt.state = CircuitOpen
        }
        
        return nil, err
    }
    
    // Success - reset circuit if it was half-open
    if cbt.state == CircuitHalfOpen {
        cbt.state = CircuitClosed
        cbt.failures = 0
    }
    
    return result, nil
}

func (cbt *CircuitBreakerTool) GetState() (CircuitState, int, time.Time) {
    cbt.mu.RLock()
    defer cbt.mu.RUnlock()
    
    return cbt.state, cbt.failures, cbt.lastFailureTime
}
```

### 3. Fallback Tool Pattern

```go
// FallbackTool tries primary tool first, then fallbacks
type FallbackTool struct {
    name         string
    description  string
    primaryTool  core.Tool
    fallbackTools []FallbackConfig
}

type FallbackConfig struct {
    Tool      core.Tool
    Condition func(error) bool
    ParamMapper func(map[string]interface{}) map[string]interface{}
}

func NewFallbackTool(name, description string, primaryTool core.Tool) *FallbackTool {
    return &FallbackTool{
        name:          name,
        description:   description,
        primaryTool:   primaryTool,
        fallbackTools: make([]FallbackConfig, 0),
    }
}

func (ft *FallbackTool) AddFallback(
    tool core.Tool,
    condition func(error) bool,
    paramMapper func(map[string]interface{}) map[string]interface{},
) *FallbackTool {
    ft.fallbackTools = append(ft.fallbackTools, FallbackConfig{
        Tool:        tool,
        Condition:   condition,
        ParamMapper: paramMapper,
    })
    return ft
}

func (ft *FallbackTool) Name() string {
    return ft.name
}

func (ft *FallbackTool) Description() string {
    return ft.description
}

func (ft *FallbackTool) ParameterSchema() map[string]core.ParameterDefinition {
    return ft.primaryTool.ParameterSchema()
}

func (ft *FallbackTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Try primary tool first
    result, err := ft.primaryTool.Execute(ctx, params)
    if err == nil {
        return map[string]interface{}{
            "result":    result,
            "tool_used": ft.primaryTool.Name(),
            "fallback":  false,
        }, nil
    }
    
    primaryError := err
    
    // Try fallback tools
    for i, fallback := range ft.fallbackTools {
        if fallback.Condition(err) {
            // Map parameters if needed
            fallbackParams := params
            if fallback.ParamMapper != nil {
                fallbackParams = fallback.ParamMapper(params)
            }
            
            result, err := fallback.Tool.Execute(ctx, fallbackParams)
            if err == nil {
                return map[string]interface{}{
                    "result":        result,
                    "tool_used":     fallback.Tool.Name(),
                    "fallback":      true,
                    "fallback_index": i,
                    "primary_error": primaryError.Error(),
                }, nil
            }
        }
    }
    
    return nil, fmt.Errorf("primary tool and all fallbacks failed, primary error: %w", primaryError)
}

// Example: Weather with Multiple Sources
func CreateWeatherWithFallbacks() *FallbackTool {
    primaryWeather := tools.NewWeatherTool(os.Getenv("OPENWEATHER_API_KEY"))
    
    return NewFallbackTool("weather_with_fallbacks", "Weather information with multiple sources", primaryWeather).
        AddFallback(
            tools.NewWeatherAPITool(os.Getenv("WEATHERAPI_KEY")),
            func(err error) bool {
                return strings.Contains(err.Error(), "api key") || strings.Contains(err.Error(), "rate limit")
            },
            func(params map[string]interface{}) map[string]interface{} {
                // Convert location format if needed
                return params
            },
        ).
        AddFallback(
            tools.NewMockWeatherTool(), // Returns mock data
            func(err error) bool {
                return true // Always try mock as last resort
            },
            nil,
        )
}
```

## Best Practices and Guidelines

### 1. Tool Design Principles

- **Composability**: Design tools that can be easily combined
- **Idempotency**: Ensure tools can be safely retried
- **Observability**: Include comprehensive monitoring and logging
- **Graceful Degradation**: Implement fallback mechanisms
- **Configuration**: Make tools configurable for different environments

### 2. Performance Optimization

- **Caching**: Cache expensive operations and API calls
- **Parallel Execution**: Use parallel execution where appropriate
- **Connection Pooling**: Reuse connections for external services
- **Batch Operations**: Combine multiple operations when possible
- **Resource Management**: Properly manage resources and connections

### 3. Error Handling Strategy

- **Categorize Errors**: Distinguish between retryable and non-retryable errors
- **Circuit Breakers**: Protect against cascading failures
- **Fallback Mechanisms**: Provide alternative execution paths
- **Monitoring**: Track error patterns and rates
- **User Experience**: Provide meaningful error messages

## Conclusion

Advanced tool patterns enable you to build sophisticated, production-ready agent systems that can handle complex workflows, recover from failures, and scale effectively. By combining these patterns, you can create robust tool ecosystems that enhance your agents' capabilities while maintaining reliability and performance.

Key takeaways:
- Use composition patterns for complex workflows
- Implement proper error handling and recovery mechanisms
- Monitor tool performance and health comprehensively
- Design for scalability and maintainability
- Follow production deployment best practices

## Next Steps

- [Tool Development](tool-development.md) - Learn the basics of tool creation
- [Tool Integration](tool-integration.md) - Understand tool integration patterns
- [Orchestration Patterns](../orchestration/orchestration-patterns.md) - Explore agent orchestration
- [Production Deployment](../deployment/README.md) - Deploy your systems

## Further Reading

- [API Reference: MCP](../../api/core.md#mcp)
- [Examples: Advanced Patterns](../../examples/)
- [Monitoring and Observability](../debugging-monitoring/README.md)