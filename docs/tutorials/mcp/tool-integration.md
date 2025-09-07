# Tool Integration in AgenticGoKit

## Overview

Integrating tools with agents is a key capability in AgenticGoKit that enables agents to perform actions beyond text generation. This tutorial covers how to connect tools to agents, configure tool access, handle tool results, and implement tool-aware prompting.

Effective tool integration allows agents to interact with external systems, access specialized information, and perform complex tasks while maintaining a natural conversation flow.

> **Quick Start:** For the fastest way to set up MCP tools, see the [MCP CLI Guide](../../guides/MCP-CLI-Guide.md) which shows how to create projects with `agentcli create --enable-mcp`.

## Prerequisites

- Understanding of [MCP Overview](README.md)
- Familiarity with [Tool Development](tool-development.md)
- Knowledge of [Agent Lifecycle](../core-concepts/agent-lifecycle.md)
- Basic understanding of [State Management](../core-concepts/state-management.md)

## MCP Integration Architecture

```
┌─────────────┐     ┌───────────────┐     ┌─────────────┐
│             │     │               │     │             │
│    Agent    │────▶│  MCP Manager  │────▶│    Tool     │
│             │     │               │     │             │
└─────────────┘     └───────────────┘     └─────────────┘
       ▲                     │                   │
       │                     ▼                   ▼
       │              ┌───────────────┐    ┌─────────────┐
       └──────────────│  Tool Result  │◀───│   External  │
                      │   Processor   │    │   Service   │
                      └───────────────┘    └─────────────┘
```

## Basic Tool Integration

### 1. Setting Up MCP Manager

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
    "github.com/kunalkushwaha/agenticgokit/tools"
)

func main() {
    // Create MCP manager
    mcpManager := core.NewMCPManager()
    
    // Register tools
    mcpManager.RegisterTool("calculator", tools.NewCalculatorTool())
    mcpManager.RegisterTool("weather", tools.NewWeatherTool(os.Getenv("WEATHER_API_KEY")))
    mcpManager.RegisterTool("counter", tools.NewCounterTool())
    
    // Create LLM provider
    llmProvider, err := core.NewOpenAIProvider(
        os.Getenv("OPENAI_API_KEY"),
        "gpt-4",
        core.WithTemperature(0.2),
        core.WithMaxTokens(1000),
    )
    if err != nil {
        log.Fatalf("Failed to create LLM provider: %v", err)
    }
    
    // Create agent with MCP capability
    agent, err := core.NewAgent("assistant").
        WithLLM(llmProvider).
        WithMCP(mcpManager).
        WithMCPConfig(core.MCPConfig{
            Tools:        []string{"calculator", "weather", "counter"},
            MaxToolCalls: 5,
            ToolTimeout:  10 * time.Second,
        }).
        Build()
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }
    
    // Create runner
    runner := core.NewRunner(100)
    runner.RegisterAgent("assistant", agent)
    
    // Start runner
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Create event with user query
    event := core.NewEvent(
        "assistant",
        core.EventData{"message": "What's 25 * 16 and what's the weather in New York?"},
        map[string]string{"session_id": "test-session"},
    )
    
    // Register callback to handle agent response
    runner.RegisterCallback(core.HookAfterAgentRun, "response-handler",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            if response, ok := args.AgentResult.OutputState.Get("response"); ok {
                fmt.Printf("Agent response: %s\n", response)
            }
            return args.State, nil
        },
    )
    
    // Emit event
    runner.Emit(event)
    
    // Wait for response
    time.Sleep(5 * time.Second)
}
```

### 2. Tool-Aware Prompting

```go
func createToolAwareAgent() (core.AgentHandler, error) {
    // Create MCP manager
    mcpManager := core.NewMCPManager()
    
    // Register tools
    mcpManager.RegisterTool("calculator", tools.NewCalculatorTool())
    mcpManager.RegisterTool("weather", tools.NewWeatherTool(os.Getenv("WEATHER_API_KEY")))
    
    // Create LLM provider
    llmProvider, err := core.NewOpenAIProvider(
        os.Getenv("OPENAI_API_KEY"),
        "gpt-4",
        core.WithTemperature(0.2),
        core.WithMaxTokens(1000),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create LLM provider: %w", err)
    }
    
    // Create tool-aware system prompt
    systemPrompt := `You are a helpful assistant with access to the following tools:

1. calculator: Performs basic arithmetic operations (add, subtract, multiply, divide)
   - Parameters: operation (string), a (number), b (number)

2. weather: Gets current weather information for a specified location
   - Parameters: location (string), units (string, optional)

When you need to use a tool, use the following format:

<tool>calculator</tool>
<parameters>
{
  "operation": "add",
  "a": 5,
  "b": 3
}
</parameters>

Wait for the tool result before continuing. Tool results will be provided in this format:

<tool_result>
{
  "result": 8
}
</tool_result>

Answer user questions directly when you can, and use tools when necessary.`
    
    // Create agent with MCP capability and tool-aware prompt
    return core.NewAgent("assistant").
        WithLLM(llmProvider).
        WithSystemPrompt(systemPrompt).
        WithMCP(mcpManager).
        WithMCPConfig(core.MCPConfig{
            Tools:        []string{"calculator", "weather"},
            MaxToolCalls: 5,
            ToolTimeout:  10 * time.Second,
        }).
        Build()
}
```

### 3. Handling Tool Results

```go
func setupToolResultHandling(runner *core.Runner) {
    // Register callback for tool execution
    runner.RegisterCallback(core.HookBeforeToolExecution, "tool-execution-logger",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            toolName := args.ToolName
            toolParams := args.ToolParams
            
            fmt.Printf("Executing tool: %s with params: %v\n", toolName, toolParams)
            return args.State, nil
        },
    )
    
    // Register callback for tool result
    runner.RegisterCallback(core.HookAfterToolExecution, "tool-result-logger",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            toolName := args.ToolName
            toolResult := args.ToolResult
            toolError := args.Error
            
            if toolError != nil {
                fmt.Printf("Tool %s failed: %v\n", toolName, toolError)
            } else {
                fmt.Printf("Tool %s result: %v\n", toolName, toolResult)
            }
            
            return args.State, nil
        },
    )
    
    // Register callback for tool error
    runner.RegisterCallback(core.HookToolError, "tool-error-handler",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            toolName := args.ToolName
            toolError := args.Error
            
            fmt.Printf("Tool error: %s - %v\n", toolName, toolError)
            
            // Add error information to state
            newState := args.State.Clone()
            newState.Set("tool_error", fmt.Sprintf("The %s tool encountered an error: %v", toolName, toolError))
            
            return newState, nil
        },
    )
}
```

## Advanced Tool Integration

### 1. Dynamic Tool Selection

```go
type DynamicToolSelector struct {
    tools       map[string]core.Tool
    llmProvider core.LLMProvider
}

func NewDynamicToolSelector(llmProvider core.LLMProvider) *DynamicToolSelector {
    return &DynamicToolSelector{
        tools:       make(map[string]core.Tool),
        llmProvider: llmProvider,
    }
}

func (dts *DynamicToolSelector) RegisterTool(name string, tool core.Tool) {
    dts.tools[name] = tool
}

func (dts *DynamicToolSelector) SelectTools(ctx context.Context, query string) ([]string, error) {
    // Create tool descriptions
    var toolDescriptions strings.Builder
    for name, tool := range dts.tools {
        toolDescriptions.WriteString(fmt.Sprintf("- %s: %s\n", name, tool.Description()))
    }
    
    // Create prompt for tool selection
    prompt := fmt.Sprintf(`Given the following user query and available tools, select the most appropriate tools to answer the query.

User query: %s

Available tools:
%s

Return only the names of the tools that should be used, separated by commas (e.g., "calculator,weather").
If no tools are needed, return "none".

Selected tools:`, query, toolDescriptions.String())
    
    // Get tool selection from LLM
    response, err := dts.llmProvider.Generate(ctx, prompt)
    if err != nil {
        return nil, fmt.Errorf("tool selection failed: %w", err)
    }
    
    // Parse response
    response = strings.TrimSpace(response)
    if response == "none" {
        return []string{}, nil
    }
    
    // Split by comma and trim spaces
    selectedTools := strings.Split(response, ",")
    for i := range selectedTools {
        selectedTools[i] = strings.TrimSpace(selectedTools[i])
    }
    
    // Validate selected tools
    validTools := make([]string, 0, len(selectedTools))
    for _, name := range selectedTools {
        if _, exists := dts.tools[name]; exists {
            validTools = append(validTools, name)
        }
    }
    
    return validTools, nil
}
```

### 2. Tool Result Processing

```go
type ToolResultProcessor struct {
    llmProvider core.LLMProvider
}

func NewToolResultProcessor(llmProvider core.LLMProvider) *ToolResultProcessor {
    return &ToolResultProcessor{
        llmProvider: llmProvider,
    }
}

func (trp *ToolResultProcessor) ProcessResult(ctx context.Context, toolName string, result interface{}, query string) (string, error) {
    // Convert result to string representation
    resultStr := fmt.Sprintf("%v", result)
    if resultMap, ok := result.(map[string]interface{}); ok {
        resultBytes, err := json.MarshalIndent(resultMap, "", "  ")
        if err == nil {
            resultStr = string(resultBytes)
        }
    }
    
    // Create prompt for result processing
    prompt := fmt.Sprintf(`You are processing the result of a tool execution. Format the result in a clear, human-readable way that answers the user's query.

User query: %s
Tool used: %s
Tool result:
%s

Provide a concise, helpful response based on this result:`, query, toolName, resultStr)
    
    // Get formatted response from LLM
    return trp.llmProvider.Generate(ctx, prompt)
}
```

### 3. Tool Caching and Performance

```go
type CachedTool struct {
    tool      core.Tool
    cache     map[string]CacheEntry
    cacheTTL  time.Duration
    mu        sync.RWMutex
}

type CacheEntry struct {
    Result    interface{}
    Timestamp time.Time
}

func NewCachedTool(tool core.Tool, cacheTTL time.Duration) *CachedTool {
    return &CachedTool{
        tool:     tool,
        cache:    make(map[string]CacheEntry),
        cacheTTL: cacheTTL,
    }
}

func (ct *CachedTool) Name() string {
    return ct.tool.Name()
}

func (ct *CachedTool) Description() string {
    return ct.tool.Description()
}

func (ct *CachedTool) ParameterSchema() map[string]core.ParameterDefinition {
    return ct.tool.ParameterSchema()
}

func (ct *CachedTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Create cache key from parameters
    cacheKey := ct.createCacheKey(params)
    
    // Check cache
    if result, found := ct.getCachedResult(cacheKey); found {
        return result, nil
    }
    
    // Execute tool
    result, err := ct.tool.Execute(ctx, params)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    ct.setCachedResult(cacheKey, result)
    
    return result, nil
}

func (ct *CachedTool) createCacheKey(params map[string]interface{}) string {
    // Create deterministic key from parameters
    keys := make([]string, 0, len(params))
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    var keyBuilder strings.Builder
    for _, k := range keys {
        keyBuilder.WriteString(fmt.Sprintf("%s:%v;", k, params[k]))
    }
    
    // Hash the key for consistent length
    hasher := sha256.New()
    hasher.Write([]byte(keyBuilder.String()))
    return fmt.Sprintf("%x", hasher.Sum(nil))
}

func (ct *CachedTool) getCachedResult(key string) (interface{}, bool) {
    ct.mu.RLock()
    defer ct.mu.RUnlock()
    
    entry, exists := ct.cache[key]
    if !exists {
        return nil, false
    }
    
    // Check if entry is expired
    if time.Since(entry.Timestamp) > ct.cacheTTL {
        return nil, false
    }
    
    return entry.Result, true
}

func (ct *CachedTool) setCachedResult(key string, result interface{}) {
    ct.mu.Lock()
    defer ct.mu.Unlock()
    
    ct.cache[key] = CacheEntry{
        Result:    result,
        Timestamp: time.Now(),
    }
    
    // Clean up expired entries periodically
    if len(ct.cache)%100 == 0 {
        go ct.cleanupExpiredEntries()
    }
}

func (ct *CachedTool) cleanupExpiredEntries() {
    ct.mu.Lock()
    defer ct.mu.Unlock()
    
    now := time.Now()
    for key, entry := range ct.cache {
        if now.Sub(entry.Timestamp) > ct.cacheTTL {
            delete(ct.cache, key)
        }
    }
}
```

## Tool Configuration and Management

### 1. Tool Configuration

```go
type ToolConfig struct {
    Name        string                 `json:"name"`
    Enabled     bool                   `json:"enabled"`
    MaxCalls    int                    `json:"max_calls"`
    Timeout     time.Duration          `json:"timeout"`
    RateLimit   RateLimitConfig        `json:"rate_limit"`
    Cache       CacheConfig            `json:"cache"`
    Parameters  map[string]interface{} `json:"parameters"`
}

type RateLimitConfig struct {
    MaxRequests int           `json:"max_requests"`
    Interval    time.Duration `json:"interval"`
}

type CacheConfig struct {
    Enabled bool          `json:"enabled"`
    TTL     time.Duration `json:"ttl"`
}

type ToolManager struct {
    tools   map[string]core.Tool
    configs map[string]ToolConfig
    mu      sync.RWMutex
}

func NewToolManager() *ToolManager {
    return &ToolManager{
        tools:   make(map[string]core.Tool),
        configs: make(map[string]ToolConfig),
    }
}

func (tm *ToolManager) RegisterTool(tool core.Tool, config ToolConfig) error {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    name := tool.Name()
    
    // Apply configuration wrappers
    wrappedTool := tool
    
    // Add caching if enabled
    if config.Cache.Enabled {
        wrappedTool = NewCachedTool(wrappedTool, config.Cache.TTL)
    }
    
    // Add rate limiting if configured
    if config.RateLimit.MaxRequests > 0 {
        wrappedTool = NewRateLimitedTool(wrappedTool, config.RateLimit.MaxRequests, config.RateLimit.Interval)
    }
    
    // Add validation
    wrappedTool = NewValidatedTool(wrappedTool, true)
    
    tm.tools[name] = wrappedTool
    tm.configs[name] = config
    
    return nil
}

func (tm *ToolManager) GetTool(name string) (core.Tool, error) {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    
    tool, exists := tm.tools[name]
    if !exists {
        return nil, fmt.Errorf("tool not found: %s", name)
    }
    
    config := tm.configs[name]
    if !config.Enabled {
        return nil, fmt.Errorf("tool disabled: %s", name)
    }
    
    return tool, nil
}

func (tm *ToolManager) ListEnabledTools() []string {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    
    var enabled []string
    for name, config := range tm.configs {
        if config.Enabled {
            enabled = append(enabled, name)
        }
    }
    
    return enabled
}
```

### 2. Tool Discovery and Registration

```go
type ToolDiscovery struct {
    registry map[string]ToolFactory
    mu       sync.RWMutex
}

type ToolFactory func(config map[string]interface{}) (core.Tool, error)

func NewToolDiscovery() *ToolDiscovery {
    td := &ToolDiscovery{
        registry: make(map[string]ToolFactory),
    }
    
    // Register built-in tool factories
    td.RegisterFactory("calculator", func(config map[string]interface{}) (core.Tool, error) {
        return tools.NewCalculatorTool(), nil
    })
    
    td.RegisterFactory("weather", func(config map[string]interface{}) (core.Tool, error) {
        apiKey, ok := config["api_key"].(string)
        if !ok || apiKey == "" {
            return nil, fmt.Errorf("weather tool requires api_key parameter")
        }
        return tools.NewWeatherTool(apiKey), nil
    })
    
    td.RegisterFactory("counter", func(config map[string]interface{}) (core.Tool, error) {
        return tools.NewCounterTool(), nil
    })
    
    return td
}

func (td *ToolDiscovery) RegisterFactory(name string, factory ToolFactory) {
    td.mu.Lock()
    defer td.mu.Unlock()
    
    td.registry[name] = factory
}

func (td *ToolDiscovery) CreateTool(name string, config map[string]interface{}) (core.Tool, error) {
    td.mu.RLock()
    factory, exists := td.registry[name]
    td.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("unknown tool type: %s", name)
    }
    
    return factory(config)
}

func (td *ToolDiscovery) ListAvailableTools() []string {
    td.mu.RLock()
    defer td.mu.RUnlock()
    
    tools := make([]string, 0, len(td.registry))
    for name := range td.registry {
        tools = append(tools, name)
    }
    
    return tools
}
```

## Production Integration Patterns

### 1. Tool Health Monitoring

```go
type ToolHealthMonitor struct {
    tools   map[string]core.Tool
    metrics map[string]*ToolMetrics
    mu      sync.RWMutex
}

type ToolMetrics struct {
    TotalCalls    int64         `json:"total_calls"`
    SuccessCalls  int64         `json:"success_calls"`
    ErrorCalls    int64         `json:"error_calls"`
    AverageTime   time.Duration `json:"average_time"`
    LastError     string        `json:"last_error"`
    LastErrorTime time.Time     `json:"last_error_time"`
}

func NewToolHealthMonitor() *ToolHealthMonitor {
    return &ToolHealthMonitor{
        tools:   make(map[string]core.Tool),
        metrics: make(map[string]*ToolMetrics),
    }
}

func (thm *ToolHealthMonitor) WrapTool(tool core.Tool) core.Tool {
    thm.mu.Lock()
    defer thm.mu.Unlock()
    
    name := tool.Name()
    thm.tools[name] = tool
    thm.metrics[name] = &ToolMetrics{}
    
    return &MonitoredTool{
        tool:    tool,
        monitor: thm,
    }
}

type MonitoredTool struct {
    tool    core.Tool
    monitor *ToolHealthMonitor
}

func (mt *MonitoredTool) Name() string {
    return mt.tool.Name()
}

func (mt *MonitoredTool) Description() string {
    return mt.tool.Description()
}

func (mt *MonitoredTool) ParameterSchema() map[string]core.ParameterDefinition {
    return mt.tool.ParameterSchema()
}

func (mt *MonitoredTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    start := time.Now()
    name := mt.tool.Name()
    
    // Execute tool
    result, err := mt.tool.Execute(ctx, params)
    
    // Record metrics
    duration := time.Since(start)
    mt.monitor.recordExecution(name, duration, err)
    
    return result, err
}

func (thm *ToolHealthMonitor) recordExecution(toolName string, duration time.Duration, err error) {
    thm.mu.Lock()
    defer thm.mu.Unlock()
    
    metrics := thm.metrics[toolName]
    metrics.TotalCalls++
    
    if err != nil {
        metrics.ErrorCalls++
        metrics.LastError = err.Error()
        metrics.LastErrorTime = time.Now()
    } else {
        metrics.SuccessCalls++
    }
    
    // Update average time (simple moving average)
    if metrics.TotalCalls == 1 {
        metrics.AverageTime = duration
    } else {
        metrics.AverageTime = time.Duration(
            (int64(metrics.AverageTime)*metrics.TotalCalls + int64(duration)) / (metrics.TotalCalls + 1),
        )
    }
}

func (thm *ToolHealthMonitor) GetMetrics(toolName string) (*ToolMetrics, error) {
    thm.mu.RLock()
    defer thm.mu.RUnlock()
    
    metrics, exists := thm.metrics[toolName]
    if !exists {
        return nil, fmt.Errorf("tool not found: %s", toolName)
    }
    
    // Return a copy to avoid race conditions
    return &ToolMetrics{
        TotalCalls:    metrics.TotalCalls,
        SuccessCalls:  metrics.SuccessCalls,
        ErrorCalls:    metrics.ErrorCalls,
        AverageTime:   metrics.AverageTime,
        LastError:     metrics.LastError,
        LastErrorTime: metrics.LastErrorTime,
    }, nil
}

func (thm *ToolHealthMonitor) GetHealthStatus() map[string]string {
    thm.mu.RLock()
    defer thm.mu.RUnlock()
    
    status := make(map[string]string)
    
    for name, metrics := range thm.metrics {
        if metrics.TotalCalls == 0 {
            status[name] = "unknown"
            continue
        }
        
        errorRate := float64(metrics.ErrorCalls) / float64(metrics.TotalCalls)
        
        switch {
        case errorRate == 0:
            status[name] = "healthy"
        case errorRate < 0.1:
            status[name] = "warning"
        default:
            status[name] = "unhealthy"
        }
    }
    
    return status
}
```

### 2. Complete Integration Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
    "github.com/kunalkushwaha/agenticgokit/tools"
)

func main() {
    // Create tool discovery and manager
    discovery := NewToolDiscovery()
    manager := NewToolManager()
    monitor := NewToolHealthMonitor()
    
    // Configure tools
    toolConfigs := []struct {
        name   string
        config ToolConfig
        params map[string]interface{}
    }{
        {
            name: "calculator",
            config: ToolConfig{
                Name:     "calculator",
                Enabled:  true,
                MaxCalls: 100,
                Timeout:  5 * time.Second,
                Cache: CacheConfig{
                    Enabled: true,
                    TTL:     1 * time.Hour,
                },
            },
        },
        {
            name: "weather",
            config: ToolConfig{
                Name:     "weather",
                Enabled:  true,
                MaxCalls: 50,
                Timeout:  10 * time.Second,
                RateLimit: RateLimitConfig{
                    MaxRequests: 10,
                    Interval:    1 * time.Minute,
                },
                Cache: CacheConfig{
                    Enabled: true,
                    TTL:     15 * time.Minute,
                },
            },
            params: map[string]interface{}{
                "api_key": os.Getenv("WEATHER_API_KEY"),
            },
        },
    }
    
    // Create and register tools
    for _, tc := range toolConfigs {
        tool, err := discovery.CreateTool(tc.name, tc.params)
        if err != nil {
            log.Fatalf("Failed to create tool %s: %v", tc.name, err)
        }
        
        // Wrap with monitoring
        monitoredTool := monitor.WrapTool(tool)
        
        // Register with manager
        if err := manager.RegisterTool(monitoredTool, tc.config); err != nil {
            log.Fatalf("Failed to register tool %s: %v", tc.name, err)
        }
    }
    
    // Create MCP manager and register tools
    mcpManager := core.NewMCPManager()
    for _, toolName := range manager.ListEnabledTools() {
        tool, err := manager.GetTool(toolName)
        if err != nil {
            log.Printf("Failed to get tool %s: %v", toolName, err)
            continue
        }
        mcpManager.RegisterTool(toolName, tool)
    }
    
    // Create LLM provider
    llmProvider, err := core.NewOpenAIProvider(
        os.Getenv("OPENAI_API_KEY"),
        "gpt-4",
        core.WithTemperature(0.2),
        core.WithMaxTokens(1000),
    )
    if err != nil {
        log.Fatalf("Failed to create LLM provider: %v", err)
    }
    
    // Create agent
    agent, err := core.NewAgent("assistant").
        WithLLM(llmProvider).
        WithMCP(mcpManager).
        WithMCPConfig(core.MCPConfig{
            Tools:        manager.ListEnabledTools(),
            MaxToolCalls: 5,
            ToolTimeout:  10 * time.Second,
        }).
        Build()
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }
    
    // Create runner with tool monitoring
    runner := core.NewRunner(100)
    runner.RegisterAgent("assistant", agent)
    
    // Setup tool result handling
    setupToolResultHandling(runner)
    
    // Start runner
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Example interaction
    event := core.NewEvent(
        "assistant",
        core.EventData{"message": "Calculate 15 * 23 and tell me the weather in London"},
        map[string]string{"session_id": "demo-session"},
    )
    
    runner.Emit(event)
    
    // Wait and show metrics
    time.Sleep(10 * time.Second)
    
    // Display tool health metrics
    fmt.Println("\nTool Health Status:")
    healthStatus := monitor.GetHealthStatus()
    for tool, status := range healthStatus {
        fmt.Printf("- %s: %s\n", tool, status)
        
        if metrics, err := monitor.GetMetrics(tool); err == nil {
            fmt.Printf("  Calls: %d (Success: %d, Errors: %d)\n", 
                metrics.TotalCalls, metrics.SuccessCalls, metrics.ErrorCalls)
            fmt.Printf("  Average Time: %v\n", metrics.AverageTime)
            if metrics.LastError != "" {
                fmt.Printf("  Last Error: %s (%v)\n", metrics.LastError, metrics.LastErrorTime)
            }
        }
    }
}
```

## Best Practices

### 1. Tool Integration Guidelines

- **Selective Tool Access**: Only provide agents with tools they actually need
- **Tool Validation**: Always validate tool parameters and results
- **Error Handling**: Implement robust error handling and recovery
- **Performance Monitoring**: Monitor tool performance and health
- **Security**: Implement proper authentication and authorization
- **Caching**: Use caching for expensive or frequently called tools
- **Rate Limiting**: Protect external services with rate limiting

### 2. Prompting Best Practices

- **Clear Tool Descriptions**: Provide clear, concise tool descriptions
- **Parameter Documentation**: Document all parameters with examples
- **Usage Examples**: Include examples of when and how to use tools
- **Error Guidance**: Explain how to handle tool errors
- **Tool Selection**: Guide the agent on tool selection criteria

### 3. Production Considerations

- **Monitoring**: Implement comprehensive tool monitoring
- **Logging**: Log all tool executions for debugging
- **Metrics**: Track tool usage and performance metrics
- **Health Checks**: Implement tool health checking
- **Fallbacks**: Provide fallback mechanisms for tool failures
- **Configuration**: Make tool configuration externally manageable

## Troubleshooting

### Common Issues

1. **Tool Not Found**: Ensure tools are properly registered with the MCP manager
2. **Parameter Validation Errors**: Check parameter schemas and validation rules
3. **Timeout Errors**: Adjust tool timeout settings for slow operations
4. **Rate Limit Exceeded**: Configure appropriate rate limits for external services
5. **Authentication Failures**: Verify API keys and authentication credentials

### Debugging Tips

- Enable detailed logging for tool executions
- Use tool health monitoring to identify problematic tools
- Test tools independently before integrating with agents
- Monitor tool performance metrics regularly
- Implement proper error handling and recovery mechanisms

## Conclusion

Tool integration is a powerful feature that extends agent capabilities beyond text generation. By following the patterns and best practices in this tutorial, you can create robust, scalable tool integrations that enhance your agent systems.

Key takeaways:
- Use the MCP manager to register and manage tools
- Implement proper tool configuration and monitoring
- Handle tool results and errors gracefully
- Follow security and performance best practices
- Monitor tool health and performance in production

## Next Steps

- [Advanced Tool Patterns](advanced-tool-patterns.md) - Explore complex tool usage patterns
- [Tool Development](tool-development.md) - Learn how to create custom tools
- [State Management](../core-concepts/state-management.md) - Understand state flow with tools

## Further Reading

- [API Reference: MCP](../../reference/api/agent.md#mcp)
- [Examples: Tool Usage](../../examples/)
- [Production Deployment](../../guides/deployment/README.md)
