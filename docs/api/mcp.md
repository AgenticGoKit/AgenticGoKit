# MCP Integration API Reference

This document provides comprehensive API reference for AgentFlow's Model Context Protocol (MCP) integration, including tool discovery, execution, and server management.

## 🏗️ Core MCP Interfaces

### `MCPManager`

The primary interface for managing MCP servers and tool execution.

```go
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
```

**Usage Example:**
```go
func (a *MCPAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    mcpManager := a.runner.GetMCPManager()
    
    // Get available tools
    tools := mcpManager.GetAvailableTools()
    if len(tools) == 0 {
        return core.AgentResult{}, fmt.Errorf("no MCP tools available")
    }
    
    // Refresh tools from servers
    err := mcpManager.RefreshTools(ctx)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to refresh tools: %w", err)
    }
    
    // Use ExecuteMCPTool helper function for simple tool execution
    result, err := core.ExecuteMCPTool(ctx, "search", map[string]interface{}{
        "query": event.GetData()["query"],
    })
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("tool execution failed: %w", err)
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "tool_result": result,
            "tools_available": len(tools),
        },
    }, nil
}
```

### `MCPAgent`

Interface for agents that can interact with MCP tools.

```go
type MCPAgent interface {
    Agent
    
    // MCP-specific methods
    SelectTools(ctx context.Context, query string, stateContext State) ([]string, error)
    ExecuteTools(ctx context.Context, tools []MCPToolExecution) ([]MCPToolResult, error)
    GetAvailableMCPTools() []MCPToolInfo
}
```

### `MCPCache`

Interface for caching MCP tool results.

```go
type MCPCache interface {
    // Get retrieves a cached tool result
    Get(ctx context.Context, key string) (*ToolResult, bool)
    
    // Set stores a tool result in the cache
    Set(ctx context.Context, key string, result *ToolResult, ttl time.Duration) error
    
    // Delete removes a cached result
    Delete(ctx context.Context, key string) error
    
    // Clear removes all cached results
    Clear(ctx context.Context) error
    
    // Stats returns cache performance statistics
    Stats() CacheStats
}
```

## 🛠️ Data Types and Structures

### `MCPServerInfo`

Represents information about an MCP server.

```go
type MCPServerInfo struct {
    Name         string                 `json:"name"`
    Type         string                 `json:"type"`
    Address      string                 `json:"address"`
    Port         int                    `json:"port"`
    Version      string                 `json:"version"`
    Description  string                 `json:"description"`
    Capabilities map[string]interface{} `json:"capabilities"`
    Status       string                 `json:"status"` // connected, disconnected, error
}
```

### `MCPToolInfo`

Represents metadata about an available MCP tool.

```go
type MCPToolInfo struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Schema      map[string]interface{} `json:"schema"`
    ServerName  string                 `json:"server_name"`
}
```

### `MCPToolExecution`

Represents a tool execution request.

```go
type MCPToolExecution struct {
    ToolName   string                 `json:"tool_name"`
    Arguments  map[string]interface{} `json:"arguments"`
    ServerName string                 `json:"server_name,omitempty"`
}
```

### `MCPToolResult`

Represents the result of an MCP tool execution.

```go
type MCPToolResult struct {
    ToolName   string        `json:"tool_name"`
    ServerName string        `json:"server_name"`
    Success    bool          `json:"success"`
    Content    []MCPContent  `json:"content"`
    Error      string        `json:"error"`
    Duration   time.Duration `json:"duration"`
}
```

### `MCPContent`

Represents content returned by an MCP tool.

```go
type MCPContent struct {
    Type     string `json:"type"`
    Text     string `json:"text"`
    Data     string `json:"data"`
    MimeType string `json:"mime_type"`
}
```

## 🚀 Helper Functions

### `ExecuteMCPTool`

Executes a single MCP tool with a simple interface.

```go
func ExecuteMCPTool(ctx context.Context, toolName string, args map[string]interface{}) (MCPToolResult, error)
```

**Description**: The simplest way to execute an MCP tool without creating an agent. Handles caching automatically if configured.

**Parameters**:
- `ctx` - Context for cancellation and timeouts
- `toolName` - Name of the tool to execute
- `args` - Arguments to pass to the tool

**Returns**:
- `MCPToolResult` - The tool execution result
- `error` - Any error that occurred

**Usage Example:**
```go
func executeTool(ctx context.Context) error {
    result, err := core.ExecuteMCPTool(ctx, "search", map[string]interface{}{
        "query": "latest Go tutorials",
        "limit": 10,
    })
    if err != nil {
        return fmt.Errorf("tool execution failed: %w", err)
    }
    
    if !result.Success {
        return fmt.Errorf("tool returned error: %s", result.Error)
    }
    
    for _, content := range result.Content {
        fmt.Printf("Result: %s\n", content.Text)
    }
    
    return nil
}
```

### `RegisterMCPToolsWithRegistry`

Discovers and registers all available MCP tools with the global registry.

```go
func RegisterMCPToolsWithRegistry(ctx context.Context) error
```

**Description**: Automatically discovers tools from all connected MCP servers and registers them with the FunctionTool registry for use in agents.

**Usage Example:**
```go
func initializeTools(ctx context.Context) error {
    // Register all MCP tools
    if err := core.RegisterMCPToolsWithRegistry(ctx); err != nil {
        return fmt.Errorf("failed to register MCP tools: %w", err)
    }
    
    // Get tool registry
    registry := core.GetMCPToolRegistry()
    tools := registry.List()
    
    log.Printf("Registered %d MCP tools", len(tools))
    return nil
}
```

## 🔧 Configuration

### `MCPConfig`

Configuration for MCP manager and servers.

```go
type MCPConfig struct {
    Servers             []MCPServerConfig `toml:"servers"`
    DefaultTimeout      time.Duration     `toml:"default_timeout"`
    RetryAttempts       int               `toml:"retry_attempts"`
    HealthCheckInterval time.Duration     `toml:"health_check_interval"`
    ToolDiscoveryMode   string            `toml:"tool_discovery_mode"`
}
```

### `MCPServerConfig`

Configuration for a single MCP server.

```go
type MCPServerConfig struct {
    Name        string            `toml:"name"`
    Type        string            `toml:"type"`
    Host        string            `toml:"host"`
    Port        int               `toml:"port"`
    Enabled     bool              `toml:"enabled"`
    Metadata    map[string]string `toml:"metadata"`
}
```

### `MCPAgentConfig`

Configuration for MCP-aware agents.

```go
type MCPAgentConfig struct {
    MaxToolsPerExecution   int           `toml:"max_tools_per_execution"`
    ToolSelectionTimeout   time.Duration `toml:"tool_selection_timeout"`
    ExecutionTimeout       time.Duration `toml:"execution_timeout"`
    EnableCaching          bool          `toml:"enable_caching"`
    ParallelExecution      bool          `toml:"parallel_execution"`
    RetryFailedTools       bool          `toml:"retry_failed_tools"`
}
```

## 🧪 Testing MCP Integration

### Mock MCP Manager

```go
type MockMCPManager struct {
    tools   map[string]*MCPToolInfo
    results map[string]*MCPToolResult
    errors  map[string]error
}

func NewMockMCPManager() *MockMCPManager {
    return &MockMCPManager{
        tools:   make(map[string]*MCPToolInfo),
        results: make(map[string]*MCPToolResult),
        errors:  make(map[string]error),
    }
}

func (m *MockMCPManager) AddTool(tool *MCPToolInfo) {
    m.tools[tool.Name] = tool
}

func (m *MockMCPManager) SetToolResult(toolName string, result *MCPToolResult) {
    m.results[toolName] = result
}

func (m *MockMCPManager) SetToolError(toolName string, err error) {
    m.errors[toolName] = err
}

func (m *MockMCPManager) GetAvailableTools() []MCPToolInfo {
    tools := make([]MCPToolInfo, 0, len(m.tools))
    for _, tool := range m.tools {
        tools = append(tools, *tool)
    }
    return tools
}
```

### Integration Tests

```go
func TestMCPIntegration(t *testing.T) {
    // Create mock MCP manager
    mockManager := NewMockMCPManager()
    
    // Add test tools
    mockManager.AddTool(&MCPToolInfo{
        Name:        "test_tool",
        Description: "A test tool",
        ServerName:  "test_server",
    })
    
    // Set expected result
    mockManager.SetToolResult("test_tool", &MCPToolResult{
        ToolName: "test_tool",
        Success:  true,
        Content: []MCPContent{
            {Type: "text", Text: "test result"},
        },
    })
    
    // Create agent with mock manager
    agent := &MCPAwareAgent{
        mcpManager: mockManager,
        config: MCPAgentConfig{
            MaxToolsPerExecution: 5,
            ExecutionTimeout:     30 * time.Second,
        },
    }
    
    // Test tool execution
    result, err := core.ExecuteMCPTool(context.Background(), "test_tool", map[string]interface{}{
        "param": "value",
    })
    
    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "test result", result.Content[0].Text)
}
```

## 📚 Integration Patterns

### Simple Tool Execution

```go
func simpleToolExample(ctx context.Context) {
    // Execute a single tool
    result, err := core.ExecuteMCPTool(ctx, "search", map[string]interface{}{
        "query": "AgentFlow documentation",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    for _, content := range result.Content {
        fmt.Println(content.Text)
    }
}
```

### Agent with MCP Tools

```go
type MyMCPAgent struct {
    mcpManager core.MCPManager
    llm        core.ModelProvider
}

func (a *MyMCPAgent) Run(ctx context.Context, state core.State) (core.State, error) {
    query := state.GetString("query")
    
    // Get available tools
    tools := a.mcpManager.GetAvailableTools()
    
    // Use LLM to select appropriate tools
    selectedTools, err := a.selectTools(ctx, query, tools)
    if err != nil {
        return state, err
    }
    
    // Execute tools
    var results []core.MCPToolResult
    for _, toolName := range selectedTools {
        result, err := core.ExecuteMCPTool(ctx, toolName, map[string]interface{}{
            "query": query,
        })
        if err != nil {
            continue // Skip failed tools
        }
        results = append(results, result)
    }
    
    // Update state with results
    state.Set("tool_results", results)
    return state, nil
}
```

This corrected MCP API reference accurately reflects the actual implementation in the AgentFlow codebase.
