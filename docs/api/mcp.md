# MCP Integration API

**Tool integration via Model Context Protocol**

This document covers AgenticGoKit's MCP (Model Context Protocol) integration API, which enables agents to discover, connect to, and use external tools and services. MCP provides a standardized way to integrate with various tools, from web search to database operations.

## 📋 Core Concepts

### MCP Overview

MCP (Model Context Protocol) is a protocol for connecting AI agents with external tools and services. AgenticGoKit provides comprehensive MCP integration with three levels of complexity:

- **Basic MCP**: Simple tool discovery and execution
- **Enhanced MCP**: Caching and performance optimization
- **Production MCP**: Enterprise-grade features with monitoring and scaling

### Core Interfaces

```go
// MCPManager provides the main interface for managing MCP connections and tools
type MCPManager interface {
    // Connection Management
    Connect(ctx context.Context, serverName string) error
    Disconnect(serverName string) error
    DisconnectAll() error

    // Server Discovery and Management
    DiscoverServers(ctx context.Context) ([]MCPServerInfo, error)
    ListConnectedServers() []string
    GetServerInfo(serverName string) (*MCPServerInfo, error)

    // Tool Management
    RefreshTools(ctx context.Context) error
    GetAvailableTools() []MCPToolInfo
    GetToolsFromServer(serverName string) []MCPToolInfo

    // Health and Monitoring
    HealthCheck(ctx context.Context) map[string]MCPHealthStatus
    GetMetrics() MCPMetrics
}

// MCPAgent represents an agent that can utilize MCP tools
type MCPAgent interface {
    Agent
    // MCP-specific methods
    SelectTools(ctx context.Context, query string, stateContext State) ([]string, error)
    ExecuteTools(ctx context.Context, tools []MCPToolExecution) ([]MCPToolResult, error)
    GetAvailableMCPTools() []MCPToolInfo
}
```

## 🚀 Basic Usage

### Quick Start with MCP

```go
package main

import (
    "context"
    "fmt"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func quickStartExample() {
    // Initialize MCP with automatic discovery
    err := core.QuickStartMCP()
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize MCP: %v", err))
    }
    
    // Create an MCP-aware agent
    llmProvider, err := core.NewOpenAIProvider()
    if err != nil {
        panic(fmt.Sprintf("Failed to create LLM provider: %v", err))
    }
    
    agent, err := core.NewMCPAgent("assistant", llmProvider)
    if err != nil {
        panic(fmt.Sprintf("Failed to create MCP agent: %v", err))
    }
    
    // Create an event that might require tool usage
    event := core.NewEvent("query", map[string]interface{}{
        "question": "What's the weather like in San Francisco?",
    })
    
    // Process the event - agent will automatically discover and use appropriate tools
    result, err := agent.Run(context.Background(), event, core.NewState())
    if err != nil {
        panic(fmt.Sprintf("Agent execution failed: %v", err))
    }
    
    fmt.Printf("Response: %s\n", result.Data["response"])
    if tools, ok := result.Data["tools_used"]; ok {
        fmt.Printf("Tools used: %v\n", tools)
    }
}
```

### Manual Tool Execution

```go
func manualToolExample() {
    // Initialize MCP
    config := core.DefaultMCPConfig()
    err := core.InitializeMCP(config)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize MCP: %v", err))
    }
    
    // Execute a specific tool directly
    result, err := core.ExecuteMCPTool(context.Background(), "web_search", map[string]interface{}{
        "query": "AgenticGoKit framework",
        "limit": 5,
    })
    
    if err != nil {
        panic(fmt.Sprintf("Tool execution failed: %v", err))
    }
    
    fmt.Printf("Search results: %+v\n", result.Content)
    fmt.Printf("Execution time: %v\n", result.Duration)
}
```

### Server Discovery and Connection

```go
func serverDiscoveryExample() {
    // Initialize MCP with discovery enabled
    config := core.MCPConfig{
        EnableDiscovery:   true,
        DiscoveryTimeout:  10 * time.Second,
        ScanPorts:         []int{8080, 8081, 8090, 3000},
        ConnectionTimeout: 30 * time.Second,
        MaxRetries:        3,
        RetryDelay:        1 * time.Second,
    }
    
    err := core.InitializeMCP(config)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize MCP: %v", err))
    }
    
    // Get the MCP manager
    manager := core.GetMCPManager()
    if manager == nil {
        panic("MCP manager not available")
    }
    
    // Discover available servers
    ctx := context.Background()
    servers, err := manager.DiscoverServers(ctx)
    if err != nil {
        fmt.Printf("Discovery failed: %v\n", err)
        return
    }
    
    fmt.Printf("Discovered %d MCP servers:\n", len(servers))
    for _, server := range servers {
        fmt.Printf("- %s (%s) at %s:%d\n", 
            server.Name, server.Type, server.Address, server.Port)
        
        // Connect to the server
        err := manager.Connect(ctx, server.Name)
        if err != nil {
            fmt.Printf("  Failed to connect: %v\n", err)
            continue
        }
        
        // Get available tools from this server
        tools := manager.GetToolsFromServer(server.Name)
        fmt.Printf("  Available tools: %d\n", len(tools))
        for _, tool := range tools {
            fmt.Printf("    - %s: %s\n", tool.Name, tool.Description)
        }
    }
    
    // Check health of all connections
    healthStatus := manager.HealthCheck(ctx)
    fmt.Printf("\nHealth Status:\n")
    for serverName, status := range healthStatus {
        fmt.Printf("- %s: %s (response time: %v)\n", 
            serverName, status.Status, status.ResponseTime)
    }
}
```

## 🔧 Enhanced MCP with Caching

### Caching Configuration

```go
func cachingExample() {
    // Configure MCP with caching
    mcpConfig := core.DefaultMCPConfig()
    mcpConfig.EnableCaching = true
    
    cacheConfig := core.MCPCacheConfig{
        Enabled:         true,
        DefaultTTL:      15 * time.Minute,
        MaxSize:         100, // 100 MB
        MaxKeys:         10000,
        EvictionPolicy:  "lru",
        CleanupInterval: 5 * time.Minute,
        Backend:         "memory",
        ToolTTLs: map[string]time.Duration{
            "web_search":     5 * time.Minute,  // Search results change frequently
            "content_fetch":  30 * time.Minute, // Content is more stable
            "weather":        10 * time.Minute, // Weather updates regularly
        },
    }
    
    // Initialize MCP with caching
    err := core.InitializeMCPWithCache(mcpConfig, cacheConfig)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize MCP with cache: %v", err))
    }
    
    // Create cache-aware agent
    llmProvider, _ := core.NewOpenAIProvider()
    agent, err := core.NewMCPAgentWithCache("cached-assistant", llmProvider)
    if err != nil {
        panic(fmt.Sprintf("Failed to create cached MCP agent: %v", err))
    }
    
    // First execution - will cache results
    event1 := core.NewEvent("query", map[string]interface{}{
        "question": "What's the current weather in New York?",
    })
    
    start := time.Now()
    result1, _ := agent.Run(context.Background(), event1, core.NewState())
    duration1 := time.Since(start)
    
    fmt.Printf("First execution: %v\n", duration1)
    fmt.Printf("Response: %s\n", result1.Data["response"])
    
    // Second execution - should use cache
    start = time.Now()
    result2, _ := agent.Run(context.Background(), event1, core.NewState())
    duration2 := time.Since(start)
    
    fmt.Printf("Second execution: %v (cached)\n", duration2)
    fmt.Printf("Speed improvement: %.2fx\n", float64(duration1)/float64(duration2))
    
    // Get cache statistics
    cacheManager := core.GetMCPCacheManager()
    if cacheManager != nil {
        stats, _ := cacheManager.GetGlobalStats(context.Background())
        fmt.Printf("Cache stats: Hit rate: %.2f%%, Total keys: %d\n", 
            stats.HitRate*100, stats.TotalKeys)
    }
}
```

### Custom Cache Backends

```go
func customCacheExample() {
    // Configure Redis cache backend
    cacheConfig := core.MCPCacheConfig{
        Enabled:    true,
        DefaultTTL: 20 * time.Minute,
        Backend:    "redis",
        BackendConfig: map[string]string{
            "address":  "localhost:6379",
            "password": "",
            "database": "0",
        },
    }
    
    mcpConfig := core.DefaultMCPConfig()
    
    err := core.InitializeMCPWithCache(mcpConfig, cacheConfig)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize MCP with Redis cache: %v", err))
    }
    
    // Use the cache-enabled system
    result, err := core.ExecuteMCPTool(context.Background(), "expensive_computation", map[string]interface{}{
        "input": "complex data",
    })
    
    if err != nil {
        panic(fmt.Sprintf("Tool execution failed: %v", err))
    }
    
    fmt.Printf("Result: %+v\n", result)
}
```

## 🏭 Production MCP

### Production Configuration

```go
func productionMCPExample() {
    // Configure production-grade MCP
    productionConfig := core.ProductionConfig{
        ConnectionPool: core.ConnectionPoolConfig{
            MinConnections:       5,
            MaxConnections:       50,
            MaxIdleTime:          30 * time.Minute,
            HealthCheckInterval:  1 * time.Minute,
            HealthCheckTimeout:   10 * time.Second,
            ReconnectBackoff:     1 * time.Second,
            MaxReconnectBackoff:  30 * time.Second,
            MaxReconnectAttempts: 10,
            ConnectionTimeout:    30 * time.Second,
            MaxConnectionAge:     1 * time.Hour,
        },
        
        RetryPolicy: core.RetryPolicyConfig{
            Strategy:    "exponential",
            BaseDelay:   100 * time.Millisecond,
            MaxDelay:    10 * time.Second,
            MaxAttempts: 5,
            Multiplier:  2.0,
            Jitter:      0.1,
            RetryableErrors: []string{
                "timeout",
                "connection_error",
                "temporary_failure",
            },
            ToolSpecificPolicies: map[string]core.ToolRetryConfig{
                "web_search": {
                    Strategy:    "linear",
                    BaseDelay:   500 * time.Millisecond,
                    MaxDelay:    5 * time.Second,
                    MaxAttempts: 3,
                },
            },
        },
        
        LoadBalancer: core.LoadBalancerConfig{
            Strategy:              "round_robin",
            HealthCheckInterval:   30 * time.Second,
            HealthCheckTimeout:    5 * time.Second,
            UnhealthyThreshold:    3,
            HealthyThreshold:      2,
            FailoverEnabled:       true,
            CircuitBreakerEnabled: true,
        },
        
        Metrics: core.MetricsConfig{
            Enabled:           true,
            Port:              9090,
            Path:              "/metrics",
            UpdateInterval:    10 * time.Second,
            PrometheusEnabled: true,
        },
        
        HealthCheck: core.HealthCheckConfig{
            Enabled:        true,
            Port:           8080,
            Path:           "/health",
            Interval:       30 * time.Second,
            Timeout:        5 * time.Second,
            ChecksRequired: 3,
        },
        
        Cache: core.CacheConfig{
            Type:               "redis",
            TTL:                15 * time.Minute,
            MaxSize:            1000, // 1GB
            BackgroundCleanup:  true,
            CleanupInterval:    5 * time.Minute,
            CompressionEnabled: true,
            PersistenceEnabled: true,
            Redis: core.RedisConfig{
                Enabled:    true,
                Address:    "redis-cluster:6379",
                Password:   "secure-password",
                Database:   0,
                PoolSize:   20,
                Timeout:    5 * time.Second,
                MaxRetries: 3,
            },
        },
        
        CircuitBreaker: core.ProductionCircuitBreakerConfig{
            FailureThreshold:    10,
            SuccessThreshold:    5,
            Timeout:             60 * time.Second,
            HalfOpenMaxCalls:    3,
            OpenStateTimeout:    30 * time.Second,
            MetricsEnabled:      true,
            NotificationEnabled: true,
        },
    }
    
    // Initialize production MCP
    ctx := context.Background()
    err := core.InitializeProductionMCP(ctx, productionConfig)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize production MCP: %v", err))
    }
    
    // Create production-ready agent
    llmProvider, _ := core.NewOpenAIProvider()
    agent, err := core.NewProductionMCPAgent("production-assistant", llmProvider, productionConfig)
    if err != nil {
        panic(fmt.Sprintf("Failed to create production MCP agent: %v", err))
    }
    
    // Use the production agent
    event := core.NewEvent("complex_query", map[string]interface{}{
        "question": "Analyze market trends and provide investment recommendations",
        "context":  "financial services",
    })
    
    result, err := agent.Run(context.Background(), event, core.NewState())
    if err != nil {
        panic(fmt.Sprintf("Production agent execution failed: %v", err))
    }
    
    fmt.Printf("Production response: %s\n", result.Data["response"])
    
    // Get production metrics
    manager := core.GetMCPManager()
    metrics := manager.GetMetrics()
    
    fmt.Printf("Production Metrics:\n")
    fmt.Printf("- Connected servers: %d\n", metrics.ConnectedServers)
    fmt.Printf("- Total tools: %d\n", metrics.TotalTools)
    fmt.Printf("- Tool executions: %d\n", metrics.ToolExecutions)
    fmt.Printf("- Average latency: %v\n", metrics.AverageLatency)
    fmt.Printf("- Error rate: %.2f%%\n", metrics.ErrorRate*100)
}
```

### Monitoring and Observability

```go
func monitoringExample() {
    // Initialize production MCP with monitoring
    config := core.ProductionConfig{
        Metrics: core.MetricsConfig{
            Enabled:           true,
            Port:              9090,
            Path:              "/metrics",
            UpdateInterval:    5 * time.Second,
            PrometheusEnabled: true,
            HistogramBuckets:  []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
        },
        HealthCheck: core.HealthCheckConfig{
            Enabled:        true,
            Port:           8080,
            Path:           "/health",
            Interval:       10 * time.Second,
            Timeout:        3 * time.Second,
            ChecksRequired: 2,
        },
    }
    
    ctx := context.Background()
    err := core.InitializeProductionMCP(ctx, config)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize production MCP: %v", err))
    }
    
    // Start monitoring goroutine
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                manager := core.GetMCPManager()
                if manager == nil {
                    continue
                }
                
                // Get health status
                healthStatus := manager.HealthCheck(context.Background())
                fmt.Printf("Health Check Results:\n")
                for serverName, status := range healthStatus {
                    fmt.Printf("- %s: %s (tools: %d, latency: %v)\n",
                        serverName, status.Status, status.ToolCount, status.ResponseTime)
                    
                    if status.Error != "" {
                        fmt.Printf("  Error: %s\n", status.Error)
                    }
                }
                
                // Get detailed metrics
                metrics := manager.GetMetrics()
                fmt.Printf("System Metrics:\n")
                fmt.Printf("- Total executions: %d\n", metrics.ToolExecutions)
                fmt.Printf("- Average latency: %v\n", metrics.AverageLatency)
                fmt.Printf("- Error rate: %.2f%%\n", metrics.ErrorRate*100)
                
                // Server-specific metrics
                for serverName, serverMetrics := range metrics.ServerMetrics {
                    fmt.Printf("- %s: %d executions, %.2f%% success rate\n",
                        serverName, 
                        serverMetrics.Executions,
                        float64(serverMetrics.SuccessfulCalls)/float64(serverMetrics.Executions)*100)
                }
                
                fmt.Println("---")
            }
        }
    }()
    
    // Simulate some tool usage
    for i := 0; i < 10; i++ {
        _, err := core.ExecuteMCPTool(context.Background(), "test_tool", map[string]interface{}{
            "iteration": i,
        })
        if err != nil {
            fmt.Printf("Tool execution %d failed: %v\n", i, err)
        }
        time.Sleep(2 * time.Second)
    }
}
```

## 🔧 Custom Tool Development

### Creating MCP Function Tools

```go
// Custom tool implementation
type WeatherTool struct {
    apiKey string
    client *http.Client
}

func NewWeatherTool(apiKey string) *WeatherTool {
    return &WeatherTool{
        apiKey: apiKey,
        client: &http.Client{Timeout: 10 * time.Second},
    }
}

func (w *WeatherTool) Name() string {
    return "get_weather"
}

func (w *WeatherTool) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
    location, ok := args["location"].(string)
    if !ok {
        return nil, fmt.Errorf("location parameter is required")
    }
    
    // Make API call to weather service
    url := fmt.Sprintf("https://api.weather.com/v1/current?key=%s&location=%s", 
        w.apiKey, url.QueryEscape(location))
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    resp, err := w.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("weather API request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
    }
    
    var weatherData map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
        return nil, fmt.Errorf("failed to decode weather response: %w", err)
    }
    
    return map[string]any{
        "location":    location,
        "temperature": weatherData["temperature"],
        "conditions":  weatherData["conditions"],
        "humidity":    weatherData["humidity"],
        "timestamp":   time.Now().Unix(),
    }, nil
}

func customToolExample() {
    // Initialize MCP tool registry
    err := core.InitializeMCPToolRegistry()
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize tool registry: %v", err))
    }
    
    // Register custom tool
    registry := core.GetMCPToolRegistry()
    weatherTool := NewWeatherTool("your-api-key")
    
    err = registry.Register(weatherTool)
    if err != nil {
        panic(fmt.Sprintf("Failed to register weather tool: %v", err))
    }
    
    // Use the custom tool
    result, err := registry.CallTool(context.Background(), "get_weather", map[string]any{
        "location": "San Francisco, CA",
    })
    
    if err != nil {
        panic(fmt.Sprintf("Tool execution failed: %v", err))
    }
    
    fmt.Printf("Weather data: %+v\n", result)
}
```

### Tool Registry Management

```go
func toolRegistryExample() {
    // Initialize registry
    err := core.InitializeMCPToolRegistry()
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize tool registry: %v", err))
    }
    
    registry := core.GetMCPToolRegistry()
    
    // Register multiple custom tools
    tools := []core.FunctionTool{
        NewWeatherTool("weather-api-key"),
        NewCalculatorTool(),
        NewTextAnalyzerTool(),
    }
    
    for _, tool := range tools {
        err := registry.Register(tool)
        if err != nil {
            fmt.Printf("Failed to register tool %s: %v\n", tool.Name(), err)
            continue
        }
        fmt.Printf("Registered tool: %s\n", tool.Name())
    }
    
    // List all available tools
    availableTools := registry.List()
    fmt.Printf("Available tools: %v\n", availableTools)
    
    // Execute tools
    ctx := context.Background()
    
    // Weather tool
    weatherResult, err := registry.CallTool(ctx, "get_weather", map[string]any{
        "location": "New York, NY",
    })
    if err == nil {
        fmt.Printf("Weather: %+v\n", weatherResult)
    }
    
    // Calculator tool
    calcResult, err := registry.CallTool(ctx, "calculate", map[string]any{
        "operation": "add",
        "a":         10,
        "b":         5,
    })
    if err == nil {
        fmt.Printf("Calculation: %+v\n", calcResult)
    }
    
    // Text analyzer tool
    textResult, err := registry.CallTool(ctx, "analyze_text", map[string]any{
        "text": "This is a sample text for analysis.",
    })
    if err == nil {
        fmt.Printf("Text analysis: %+v\n", textResult)
    }
}
```

## 🧪 Testing MCP Integration

### Unit Testing MCP Tools

```go
func TestWeatherTool(t *testing.T) {
    tool := NewWeatherTool("test-api-key")
    
    // Test valid input
    result, err := tool.Call(context.Background(), map[string]any{
        "location": "San Francisco, CA",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Contains(t, result, "location")
    assert.Contains(t, result, "temperature")
    
    // Test missing location
    _, err = tool.Call(context.Background(), map[string]any{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "location parameter is required")
}

func TestMCPAgentIntegration(t *testing.T) {
    // Initialize test MCP environment
    config := core.DefaultMCPConfig()
    config.EnableDiscovery = false // Disable discovery for testing
    
    err := core.InitializeMCP(config)
    require.NoError(t, err)
    
    // Register test tools
    registry := core.GetMCPToolRegistry()
    testTool := &MockTool{
        name: "test_tool",
        response: map[string]any{
            "result": "test_response",
        },
    }
    
    err = registry.Register(testTool)
    require.NoError(t, err)
    
    // Create MCP agent
    mockLLM := &MockLLMProvider{}
    agent, err := core.NewMCPAgent("test-agent", mockLLM)
    require.NoError(t, err)
    
    // Test agent with tool usage
    event := core.NewEvent("query", map[string]interface{}{
        "question": "Use the test tool",
    })
    
    result, err := agent.Run(context.Background(), event, core.NewState())
    require.NoError(t, err)
    
    // Verify tool was used
    assert.Contains(t, result.Data, "tools_used")
    toolsUsed := result.Data["tools_used"].([]string)
    assert.Contains(t, toolsUsed, "test_tool")
}
```

### Integration Testing

```go
func TestMCPEndToEnd(t *testing.T) {
    // Skip if not running integration tests
    if testing.Short() {
        t.Skip("Skipping MCP integration test")
    }
    
    // Initialize full MCP stack
    mcpConfig := core.DefaultMCPConfig()
    cacheConfig := core.DefaultMCPCacheConfig()
    
    err := core.InitializeMCPWithCache(mcpConfig, cacheConfig)
    require.NoError(t, err)
    
    // Register MCP tools from discovered servers
    ctx := context.Background()
    err = core.RegisterMCPToolsWithRegistry(ctx)
    require.NoError(t, err)
    
    // Create agent and test complete workflow
    llmProvider, err := core.NewOpenAIProvider()
    require.NoError(t, err)
    
    agent, err := core.NewMCPAgentWithCache("integration-test-agent", llmProvider)
    require.NoError(t, err)
    
    // Test complex query that requires multiple tools
    event := core.NewEvent("complex_query", map[string]interface{}{
        "question": "Search for information about Go programming and summarize the results",
    })
    
    result, err := agent.Run(ctx, event, core.NewState())
    require.NoError(t, err)
    
    // Verify response
    assert.Contains(t, result.Data, "response")
    assert.NotEmpty(t, result.Data["response"])
    
    // Verify tools were used
    if toolsUsed, ok := result.Data["tools_used"]; ok {
        tools := toolsUsed.([]string)
        assert.NotEmpty(t, tools)
        t.Logf("Tools used: %v", tools)
    }
    
    // Test caching by running the same query again
    start := time.Now()
    result2, err := agent.Run(ctx, event, core.NewState())
    duration := time.Since(start)
    
    require.NoError(t, err)
    assert.Equal(t, result.Data["response"], result2.Data["response"])
    
    // Second execution should be faster due to caching
    assert.Less(t, duration, 1*time.Second, "Cached execution should be faster")
}
```

## 📚 Best Practices

### 1. MCP Configuration

```go
// Good: Appropriate timeouts and retry policies
config := core.MCPConfig{
    EnableDiscovery:   true,
    DiscoveryTimeout:  10 * time.Second,  // Reasonable discovery timeout
    ConnectionTimeout: 30 * time.Second,  // Allow time for connection
    MaxRetries:        3,                 // Reasonable retry attempts
    RetryDelay:        1 * time.Second,   // Progressive backoff
    EnableCaching:     true,              // Enable performance optimization
    CacheTimeout:      15 * time.Minute, // Balance freshness vs performance
}

// Bad: Unrealistic or missing configuration
config := core.MCPConfig{
    EnableDiscovery:   true,
    DiscoveryTimeout:  1 * time.Second,   // Too short
    ConnectionTimeout: 5 * time.Minute,   // Too long
    MaxRetries:        20,                // Too many retries
    RetryDelay:        100 * time.Millisecond, // Too aggressive
    EnableCaching:     false,             // Missing optimization
}
```

### 2. Error Handling

```go
// Good: Comprehensive error handling
func robustMCPUsage() {
    manager := core.GetMCPManager()
    if manager == nil {
        log.Fatal("MCP manager not initialized")
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    result, err := core.ExecuteMCPTool(ctx, "web_search", map[string]interface{}{
        "query": "search term",
    })
    
    if err != nil {
        // Handle different types of errors
        if strings.Contains(err.Error(), "timeout") {
            log.Printf("Tool execution timed out: %v", err)
            // Implement fallback or retry logic
        } else if strings.Contains(err.Error(), "not found") {
            log.Printf("Tool not available: %v", err)
            // Use alternative tool or method
        } else {
            log.Printf("Tool execution failed: %v", err)
            // General error handling
        }
        return
    }
    
    // Process successful result
    fmt.Printf("Tool result: %+v\n", result)
}
```

### 3. Performance Optimization

```go
// Good: Use caching and connection pooling
func optimizedMCPUsage() {
    // Configure with performance optimizations
    productionConfig := core.ProductionConfig{
        ConnectionPool: core.ConnectionPoolConfig{
            MinConnections: 5,
            MaxConnections: 20,
            MaxIdleTime:    30 * time.Minute,
        },
        Cache: core.CacheConfig{
            Type:               "redis",
            TTL:                15 * time.Minute,
            CompressionEnabled: true,
        },
    }
    
    ctx := context.Background()
    err := core.InitializeProductionMCP(ctx, productionConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the optimized system
    llmProvider, _ := core.NewOpenAIProvider()
    agent, _ := core.NewProductionMCPAgent("optimized-agent", llmProvider, productionConfig)
    
    // Agent will automatically use connection pooling and caching
}
```

### 4. Security Considerations

```go
// Good: Secure MCP configuration
func secureMCPSetup() {
    config := core.MCPConfig{
        // Only connect to trusted servers
        Servers: []core.MCPServerConfig{
            {
                Name:    "trusted-search",
                Type:    "tcp",
                Host:    "internal-search.company.com",
                Port:    8080,
                Enabled: true,
            },
        },
        EnableDiscovery: false, // Disable auto-discovery in production
        MaxConnections:  10,    // Limit resource usage
    }
    
    // Validate server certificates and use secure connections
    err := core.InitializeMCP(config)
    if err != nil {
        log.Fatal(err)
    }
}
```

## 🔗 Related APIs

- **[Agent API](agent.md)** - Building individual agents
- **[Orchestration API](orchestration.md)** - Multi-agent coordination
- **[State & Event API](state-event.md)** - Data flow and communication
- **[Configuration API](configuration.md)** - System configuration

---

*This documentation covers the current MCP Integration API in AgenticGoKit. The framework is actively developed, so some interfaces may evolve.*