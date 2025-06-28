# MCP Integration API Reference

This document provides comprehensive API reference for AgentFlow's Model Context Protocol (MCP) integration, including tool discovery, execution, and server management.

## 🏗️ Core MCP Interfaces

### `MCPManager`

The primary interface for managing MCP servers and tool execution.

```go
type MCPManager interface {
    // ListTools returns all available tools from all connected MCP servers
    ListTools(ctx context.Context) ([]ToolSchema, error)
    
    // GetTool returns the schema for a specific tool
    GetTool(ctx context.Context, toolName string) (*ToolSchema, error)
    
    // ExecuteTool executes a tool with the given parameters
    ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*ToolResult, error)
    
    // ListServers returns information about all configured MCP servers
    ListServers(ctx context.Context) ([]ServerInfo, error)
    
    // GetServerStatus returns the health status of a specific server
    GetServerStatus(ctx context.Context, serverName string) (*ServerStatus, error)
    
    // RefreshTools re-discovers tools from all servers
    RefreshTools(ctx context.Context) error
    
    // Close shuts down all MCP server connections
    Close() error
}
```

**Usage Example:**
```go
func (a *MCPAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    mcpManager := a.runner.GetMCPManager()
    
    // List available tools
    tools, err := mcpManager.ListTools(ctx)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to list tools: %w", err)
    }
    
    // Select appropriate tool based on query
    tool := a.selectTool(event.GetData()["query"].(string), tools)
    
    // Execute tool
    result, err := mcpManager.ExecuteTool(ctx, tool.Name, a.buildToolParams(event))
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("tool execution failed: %w", err)
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "tool_result": result,
            "tool_used":   tool.Name,
        },
    }, nil
}
```

### `MCPAgent`

Interface for agents that can interact with MCP tools.

```go
type MCPAgent interface {
    Agent
    
    // GetToolRegistry returns the agent's tool registry
    GetToolRegistry() ToolRegistry
    
    // ExecuteToolCall executes a single tool call
    ExecuteToolCall(ctx context.Context, toolCall ToolCall) (*ToolResult, error)
    
    // ExecuteToolCalls executes multiple tool calls in parallel
    ExecuteToolCalls(ctx context.Context, toolCalls []ToolCall) ([]ToolResult, error)
    
    // FormatToolsForPrompt formats available tools for LLM prompt inclusion
    FormatToolsForPrompt(ctx context.Context) (string, error)
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
    
    // Stats returns cache statistics
    Stats() CacheStats
}
```

## 🛠️ Tool Types and Schemas

### `ToolSchema`

Describes the structure and parameters of an MCP tool.

```go
type ToolSchema struct {
    // Name is the unique identifier for the tool
    Name string `json:"name"`
    
    // Description explains what the tool does
    Description string `json:"description"`
    
    // Parameters define the input schema using JSON Schema
    Parameters Parameters `json:"parameters"`
    
    // ServerName identifies which MCP server provides this tool
    ServerName string `json:"server_name"`
    
    // Version of the tool schema
    Version string `json:"version,omitempty"`
    
    // Category for tool organization
    Category string `json:"category,omitempty"`
    
    // Tags for tool discovery
    Tags []string `json:"tags,omitempty"`
    
    // Examples of tool usage
    Examples []ToolExample `json:"examples,omitempty"`
}
```

### `Parameters`

JSON Schema definition for tool parameters.

```go
type Parameters struct {
    Type        string                `json:"type"`
    Properties  map[string]Property   `json:"properties"`
    Required    []string              `json:"required,omitempty"`
    Description string                `json:"description,omitempty"`
    Examples    []interface{}         `json:"examples,omitempty"`
}

type Property struct {
    Type        string        `json:"type"`
    Description string        `json:"description,omitempty"`
    Enum        []interface{} `json:"enum,omitempty"`
    Default     interface{}   `json:"default,omitempty"`
    Examples    []interface{} `json:"examples,omitempty"`
    Format      string        `json:"format,omitempty"`
    Pattern     string        `json:"pattern,omitempty"`
    Minimum     *float64      `json:"minimum,omitempty"`
    Maximum     *float64      `json:"maximum,omitempty"`
    MinLength   *int          `json:"minLength,omitempty"`
    MaxLength   *int          `json:"maxLength,omitempty"`
}
```

**Example Tool Schema:**
```go
searchToolSchema := &ToolSchema{
    Name:        "web_search",
    Description: "Search the web for information on a given topic",
    Parameters: Parameters{
        Type: "object",
        Properties: map[string]Property{
            "query": {
                Type:        "string",
                Description: "The search query",
                Examples:    []interface{}{"climate change", "AI developments 2024"},
            },
            "max_results": {
                Type:        "integer",
                Description: "Maximum number of results to return",
                Default:     10,
                Minimum:     func() *float64 { v := 1.0; return &v }(),
                Maximum:     func() *float64 { v := 100.0; return &v }(),
            },
            "language": {
                Type:        "string",
                Description: "Language for search results",
                Enum:        []interface{}{"en", "es", "fr", "de", "zh"},
                Default:     "en",
            },
        },
        Required: []string{"query"},
    },
    Category: "search",
    Tags:     []string{"web", "information", "research"},
}
```

### `ToolCall`

Represents a request to execute a specific tool.

```go
type ToolCall struct {
    // ID is a unique identifier for this tool call
    ID string `json:"id"`
    
    // Name is the name of the tool to execute
    Name string `json:"name"`
    
    // Parameters contains the arguments for the tool
    Parameters map[string]interface{} `json:"parameters"`
    
    // ServerName specifies which MCP server to use (optional)
    ServerName string `json:"server_name,omitempty"`
    
    // Timeout for tool execution
    Timeout time.Duration `json:"timeout,omitempty"`
    
    // Metadata for tracking and debugging
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

### `ToolResult`

The response from executing an MCP tool.

```go
type ToolResult struct {
    // ID matches the ID from the corresponding ToolCall
    ID string `json:"id"`
    
    // Success indicates if the tool execution was successful
    Success bool `json:"success"`
    
    // Content contains the tool output
    Content []Content `json:"content,omitempty"`
    
    // Error contains error information if Success is false
    Error string `json:"error,omitempty"`
    
    // Metadata contains additional information about the execution
    Metadata map[string]interface{} `json:"metadata,omitempty"`
    
    // ExecutionTime tracks how long the tool took to execute
    ExecutionTime time.Duration `json:"execution_time,omitempty"`
    
    // ServerName identifies which server executed the tool
    ServerName string `json:"server_name,omitempty"`
}

type Content struct {
    Type string      `json:"type"`
    Text string      `json:"text,omitempty"`
    Data interface{} `json:"data,omitempty"`
}
```

## 🚀 Tool Execution Functions

### `ParseAndExecuteToolCalls`

Parses LLM-generated tool calls and executes them.

```go
func ParseAndExecuteToolCalls(
    ctx context.Context,
    mcpManager MCPManager,
    llmResponse string,
) ([]ToolResult, error)
```

**Usage Example:**
```go
func (a *ToolAgent) processLLMResponse(ctx context.Context, llmResponse string) ([]ToolResult, error) {
    // Parse and execute tool calls from LLM response
    results, err := core.ParseAndExecuteToolCalls(ctx, a.mcpManager, llmResponse)
    if err != nil {
        return nil, fmt.Errorf("failed to execute tool calls: %w", err)
    }
    
    // Filter out failed results
    var successfulResults []ToolResult
    for _, result := range results {
        if result.Success {
            successfulResults = append(successfulResults, result)
        } else {
            log.Printf("Tool call failed: %s - %s", result.ID, result.Error)
        }
    }
    
    return successfulResults, nil
}
```

### `FormatToolsForPrompt`

Formats available tools for inclusion in LLM prompts.

```go
func FormatToolsForPrompt(
    ctx context.Context,
    mcpManager MCPManager,
    options ...ToolFormattingOption,
) (string, error)
```

**Tool Formatting Options:**
```go
type ToolFormattingOption func(*ToolFormattingConfig)

func WithToolCategories(categories ...string) ToolFormattingOption
func WithMaxTools(max int) ToolFormattingOption
func WithIncludeExamples(include bool) ToolFormattingOption
func WithFormat(format ToolFormat) ToolFormattingOption

type ToolFormat string

const (
    ToolFormatJSON     ToolFormat = "json"
    ToolFormatMarkdown ToolFormat = "markdown"
    ToolFormatPlain    ToolFormat = "plain"
)
```

**Usage Example:**
```go
func (a *ToolAgent) buildPrompt(ctx context.Context, query string) (string, error) {
    // Format tools for prompt
    toolsText, err := core.FormatToolsForPrompt(ctx, a.mcpManager,
        core.WithToolCategories("search", "calculation"),
        core.WithMaxTools(10),
        core.WithIncludeExamples(true),
        core.WithFormat(core.ToolFormatJSON),
    )
    if err != nil {
        return "", fmt.Errorf("failed to format tools: %w", err)
    }
    
    prompt := fmt.Sprintf(`
You are an AI assistant with access to the following tools:

%s

User Query: %s

Please analyze the query and use appropriate tools to provide a comprehensive answer.
When calling tools, use the exact JSON format specified above.
`, toolsText, query)
    
    return prompt, nil
}
```

## 🖥️ Server Management

### `ServerInfo`

Information about an MCP server.

```go
type ServerInfo struct {
    Name        string            `json:"name"`
    Command     string            `json:"command"`
    Args        []string          `json:"args"`
    Env         map[string]string `json:"env"`
    WorkingDir  string            `json:"working_dir"`
    Status      ServerStatus      `json:"status"`
    Tools       []string          `json:"tools"`
    Resources   []string          `json:"resources"`
    Prompts     []string          `json:"prompts"`
    StartedAt   time.Time         `json:"started_at"`
    LastPing    time.Time         `json:"last_ping"`
}
```

### `ServerStatus`

Status information for an MCP server.

```go
type ServerStatus struct {
    State       ServerState `json:"state"`
    Healthy     bool        `json:"healthy"`
    LastError   string      `json:"last_error,omitempty"`
    Uptime      time.Duration `json:"uptime"`
    RequestCount int64       `json:"request_count"`
    ErrorCount   int64       `json:"error_count"`
    AvgLatency   time.Duration `json:"avg_latency"`
}

type ServerState string

const (
    ServerStateStarting ServerState = "starting"
    ServerStateRunning  ServerState = "running"
    ServerStateStopping ServerState = "stopping"
    ServerStateStopped  ServerState = "stopped"
    ServerStateError    ServerState = "error"
)
```

### Server Management Functions

```go
// StartMCPServer starts a new MCP server
func StartMCPServer(ctx context.Context, config ServerConfig) (*MCPServer, error)

// StopMCPServer gracefully stops an MCP server
func StopMCPServer(ctx context.Context, server *MCPServer) error

// RestartMCPServer restarts an MCP server
func RestartMCPServer(ctx context.Context, server *MCPServer) error

// HealthCheckMCPServer checks if an MCP server is healthy
func HealthCheckMCPServer(ctx context.Context, server *MCPServer) (*ServerStatus, error)
```

## 🔧 Configuration

### `MCPConfig`

Configuration for MCP integration.

```go
type MCPConfig struct {
    Enabled bool `toml:"enabled"`
    
    // Global settings
    Timeout         time.Duration `toml:"timeout"`
    MaxConnections  int           `toml:"max_connections"`
    RetryAttempts   int           `toml:"retry_attempts"`
    RetryDelay      time.Duration `toml:"retry_delay"`
    
    // Cache settings
    CacheEnabled    bool          `toml:"cache_enabled"`
    CacheTTL        time.Duration `toml:"cache_ttl"`
    CacheMaxSize    int           `toml:"cache_max_size"`
    
    // Server configurations
    Servers []ServerConfig `toml:"servers"`
}

type ServerConfig struct {
    Name       string            `toml:"name"`
    Command    string            `toml:"command"`
    Args       []string          `toml:"args"`
    Env        map[string]string `toml:"env"`
    WorkingDir string            `toml:"working_dir"`
    
    // Connection settings
    Address     string        `toml:"address"`
    Timeout     time.Duration `toml:"timeout"`
    HealthCheck time.Duration `toml:"health_check"`
    
    // Tool filtering
    IncludeTools []string `toml:"include_tools"`
    ExcludeTools []string `toml:"exclude_tools"`
}
```

**Configuration Example:**
```toml
[mcp]
enabled = true
timeout = "30s"
max_connections = 10
retry_attempts = 3
retry_delay = "1s"
cache_enabled = true
cache_ttl = "1h"
cache_max_size = 1000

[[mcp.servers]]
name = "web-search"
command = "mcp-web-search"
args = ["--port", "8080"]
env = { SEARCH_API_KEY = "${SEARCH_API_KEY}" }
timeout = "30s"
health_check = "10s"

[[mcp.servers]]
name = "file-tools"
command = "docker"
args = ["run", "-v", "/data:/data", "mcp-file-tools"]
working_dir = "/app"
include_tools = ["read_file", "write_file", "list_files"]

[[mcp.servers]]
name = "database"
address = "tcp://localhost:9090"
env = { DB_CONNECTION = "${DATABASE_URL}" }
exclude_tools = ["drop_table", "delete_database"]
```

## 📊 Tool Discovery and Registry

### `ToolRegistry`

Registry for managing discovered tools.

```go
type ToolRegistry interface {
    // RegisterTool adds a tool to the registry
    RegisterTool(tool *ToolSchema) error
    
    // UnregisterTool removes a tool from the registry
    UnregisterTool(toolName string) error
    
    // GetTool retrieves a tool by name
    GetTool(toolName string) (*ToolSchema, bool)
    
    // ListTools returns all registered tools
    ListTools() []*ToolSchema
    
    // SearchTools finds tools matching criteria
    SearchTools(criteria SearchCriteria) []*ToolSchema
    
    // RefreshFromServers re-discovers tools from all servers
    RefreshFromServers(ctx context.Context) error
}

type SearchCriteria struct {
    Categories []string
    Tags       []string
    Pattern    string
    ServerName string
}
```

### Tool Discovery Functions

```go
// DiscoverTools discovers all tools from configured MCP servers
func DiscoverTools(ctx context.Context, config MCPConfig) ([]ToolSchema, error)

// DiscoverToolsFromServer discovers tools from a specific server
func DiscoverToolsFromServer(ctx context.Context, server ServerConfig) ([]ToolSchema, error)

// FilterTools filters tools based on criteria
func FilterTools(tools []ToolSchema, criteria SearchCriteria) []ToolSchema
```

## 🧪 Testing MCP Integration

### Mock MCP Manager

```go
type MockMCPManager struct {
    tools   map[string]*ToolSchema
    results map[string]*ToolResult
    errors  map[string]error
}

func NewMockMCPManager() *MockMCPManager {
    return &MockMCPManager{
        tools:   make(map[string]*ToolSchema),
        results: make(map[string]*ToolResult),
        errors:  make(map[string]error),
    }
}

func (m *MockMCPManager) AddTool(tool *ToolSchema) {
    m.tools[tool.Name] = tool
}

func (m *MockMCPManager) SetToolResult(toolName string, result *ToolResult) {
    m.results[toolName] = result
}

func (m *MockMCPManager) SetToolError(toolName string, err error) {
    m.errors[toolName] = err
}

func (m *MockMCPManager) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*ToolResult, error) {
    if err, exists := m.errors[toolName]; exists {
        return nil, err
    }
    
    if result, exists := m.results[toolName]; exists {
        return result, nil
    }
    
    return &ToolResult{
        Success: true,
        Content: []Content{{Type: "text", Text: "Mock result"}},
    }, nil
}
```

### Testing Example

```go
func TestMCPAgent(t *testing.T) {
    // Create mock MCP manager
    mockMCP := NewMockMCPManager()
    
    // Add test tool
    mockMCP.AddTool(&ToolSchema{
        Name:        "test_tool",
        Description: "A test tool",
        Parameters: Parameters{
            Type:       "object",
            Properties: map[string]Property{
                "input": {Type: "string", Description: "Test input"},
            },
            Required: []string{"input"},
        },
    })
    
    // Set expected result
    mockMCP.SetToolResult("test_tool", &ToolResult{
        Success: true,
        Content: []Content{{Type: "text", Text: "Test result"}},
    })
    
    // Create agent with mock MCP manager
    agent := &MCPAgent{mcpManager: mockMCP}
    
    // Test tool execution
    event := core.NewEvent("test", map[string]interface{}{
        "query": "use test_tool with input 'hello'",
    })
    
    result, err := agent.Run(context.Background(), event, core.NewState())
    
    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.Contains(t, result.Data, "tool_result")
}
```

This MCP integration API reference provides comprehensive coverage of AgentFlow's MCP capabilities, from basic tool execution to advanced server management and testing utilities.
