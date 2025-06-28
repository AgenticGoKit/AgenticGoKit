# Core Package API Reference

**Complete reference for the AgentFlow core package**

The `core` package provides the complete public API for AgentFlow. All user code should import only from this package.

```go
import agentflow "github.com/kunalkushwaha/agentflow/core"
```

## Interfaces

### AgentHandler

Primary interface for implementing agent logic.

```go
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}
```

**Description**: Processes events and manages agent state. This is the main interface for building agents.

**Parameters**:
- `ctx` - Context for cancellation and timeouts
- `event` - Input event containing user data and metadata  
- `state` - Thread-safe state storage for the agent workflow

**Returns**:
- `AgentResult` - The agent's response and updated state
- `error` - Any error that occurred during processing

**Example**:
```go
type MyAgent struct {
    llm agentflow.ModelProvider
}

func (a *MyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    message := event.GetData()["message"]
    response, err := a.llm.Generate(ctx, fmt.Sprintf("User: %s", message))
    if err != nil {
        return agentflow.AgentResult{}, err
    }
    
    state.Set("response", response)
    return agentflow.AgentResult{Result: response, State: state}, nil
}
```

### ModelProvider

Unified interface for LLM providers.

```go
type ModelProvider interface {
    Generate(ctx context.Context, prompt string) (string, error)
    GenerateWithHistory(ctx context.Context, messages []Message) (string, error)
    Name() string
}
```

**Methods**:

#### Generate
```go
Generate(ctx context.Context, prompt string) (string, error)
```
Generate a response from a single prompt.

**Parameters**:
- `ctx` - Context for cancellation
- `prompt` - The input prompt string

**Returns**:
- Response string from the LLM
- Error if generation fails

#### GenerateWithHistory
```go
GenerateWithHistory(ctx context.Context, messages []Message) (string, error)
```
Generate a response with conversation history.

**Parameters**:
- `ctx` - Context for cancellation
- `messages` - Array of conversation messages

**Returns**:
- Response string from the LLM
- Error if generation fails

#### Name
```go
Name() string
```
Returns the provider name (e.g., "azure", "openai", "ollama").

### MCPManager

Interface for MCP (Model Context Protocol) integration.

```go
type MCPManager interface {
    // Connection Management
    Connect(ctx context.Context, serverName string) error
    Disconnect(serverName string) error
    DisconnectAll() error

    // Server Discovery
    DiscoverServers(ctx context.Context) ([]MCPServerInfo, error)
    ListConnectedServers() []string
    GetServerInfo(serverName string) (*MCPServerInfo, error)

    // Tool Management
    RefreshTools(ctx context.Context) error
    GetAvailableTools() []MCPToolInfo
    GetToolsFromServer(serverName string) []MCPToolInfo

    // Tool Execution
    ListTools(ctx context.Context) ([]ToolSchema, error)
    CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
    GetToolSchema(toolName string) (*ToolSchema, error)

    // Cache Management
    GetCacheStats() CacheStats
    ClearCache() error

    // Lifecycle
    Close() error
}
```

**Key Methods**:

#### ListTools
```go
ListTools(ctx context.Context) ([]ToolSchema, error)
```
Discover and return all available tools from connected MCP servers.

#### CallTool
```go
CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
```
Execute a specific tool with the given arguments.

**Example**:
```go
// List available tools
tools, err := mcpManager.ListTools(ctx)
for _, tool := range tools {
    fmt.Printf("Tool: %s - %s\n", tool.Name, tool.Description)
}

// Execute a tool
result, err := mcpManager.CallTool(ctx, "search", map[string]interface{}{
    "query": "latest Go tutorials",
})
```

### Event

Interface for event data that flows between agents.

```go
type Event interface {
    GetID() string
    GetData() EventData
    GetMetadata() map[string]string
    GetMetadataValue(key string) (string, bool)
    GetSourceAgentID() string
    GetTargetAgentID() string
}
```

**Methods**:

#### GetData
```go
GetData() EventData
```
Returns the event payload as key-value pairs.

**Example**:
```go
data := event.GetData()
message, ok := data["message"]
if ok {
    fmt.Printf("User message: %s\n", message)
}
```

### State

Thread-safe state management interface.

```go
type State interface {
    Get(key string) (any, bool)
    Set(key string, value any)
    Keys() []string

    GetMeta(key string) (string, bool)
    SetMeta(key, value string)
    MetaKeys() []string

    Clone() State
}
```

**Key Methods**:

#### Get/Set
```go
Set(key string, value any)
Get(key string) (any, bool)
```
Store and retrieve typed values.

#### GetMeta/SetMeta
```go
SetMeta(key, value string)
GetMeta(key string) (string, bool)
```
Store and retrieve string metadata.

**Example**:
```go
// Store data
state.Set("user_preferences", []string{"technical", "detailed"})
state.Set("conversation_turn", 3)

// Store metadata
state.SetMeta("processed_by", "agent1")
state.SetMeta("timestamp", time.Now().Format(time.RFC3339))

// Retrieve data
prefs, exists := state.Get("user_preferences")
if exists {
    preferences := prefs.([]string)
    fmt.Printf("User preferences: %v\n", preferences)
}

// Retrieve metadata
agent, exists := state.GetMeta("processed_by")
if exists {
    fmt.Printf("Processed by: %s\n", agent)
}
```

## Types

### Message

Represents a conversation message for LLM providers.

```go
type Message struct {
    Role    string // "system", "user", "assistant"
    Content string
}
```

**Example**:
```go
messages := []agentflow.Message{
    {Role: "system", Content: "You are a helpful assistant."},
    {Role: "user", Content: "Explain Go interfaces."},
    {Role: "assistant", Content: "Go interfaces define method signatures..."},
    {Role: "user", Content: "Can you give an example?"},
}

response, err := provider.GenerateWithHistory(ctx, messages)
```

### AgentResult

Result returned by agent execution.

```go
type AgentResult struct {
    Result string
    State  State
    Error  error
}
```

**Fields**:
- `Result` - The agent's response text
- `State` - Updated state after processing
- `Error` - Any error that occurred (optional)

### ToolSchema

Describes an available MCP tool.

```go
type ToolSchema struct {
    Name        string
    Description string
    Parameters  map[string]interface{}
}
```

**Example**:
```go
tool := agentflow.ToolSchema{
    Name:        "search",
    Description: "Search the web for information",
    Parameters: map[string]interface{}{
        "query": map[string]interface{}{
            "type":        "string",
            "description": "Search query",
            "required":    true,
        },
    },
}
```

### EventData

Type alias for event payload data.

```go
type EventData map[string]interface{}
```

**Usage**:
```go
eventData := agentflow.EventData{
    "message":    "Hello, world!",
    "user_id":    "12345",
    "timestamp":  time.Now(),
    "context":    map[string]string{"session": "abc123"},
}

event := agentflow.NewEvent("user-message", eventData, nil)
```

## Factory Functions

### Agent Factories

#### NewMCPAgent
```go
func NewMCPAgent(name string, llm ModelProvider, mcp MCPManager) AgentHandler
```
Creates an agent with MCP tool integration.

**Parameters**:
- `name` - Unique agent identifier
- `llm` - Model provider for generating responses
- `mcp` - MCP manager for tool discovery and execution

**Example**:
```go
provider, _ := agentflow.NewAzureProvider(azureConfig)
mcpManager, _ := agentflow.InitializeProductionMCP(ctx, mcpConfig)
agent := agentflow.NewMCPAgent("research-agent", provider, mcpManager)
```

### Provider Factories

#### NewAzureProvider
```go
func NewAzureProvider(config AzureConfig) (ModelProvider, error)
```
Creates an Azure OpenAI provider.

**Example**:
```go
config := agentflow.AzureConfig{
    APIKey:      "your-api-key",
    Endpoint:    "https://your-resource.openai.azure.com",
    Deployment:  "gpt-4",
    APIVersion:  "2024-02-15-preview",
    MaxTokens:   2000,
    Temperature: 0.7,
}

provider, err := agentflow.NewAzureProvider(config)
```

#### NewOpenAIProvider
```go
func NewOpenAIProvider(config OpenAIConfig) (ModelProvider, error)
```
Creates an OpenAI provider.

#### NewOllamaProvider
```go
func NewOllamaProvider(config OllamaConfig) (ModelProvider, error)
```
Creates an Ollama provider for local models.

#### NewMockProvider
```go
func NewMockProvider(config MockConfig) ModelProvider
```
Creates a mock provider for testing.

### MCP Factories

#### InitializeProductionMCP
```go
func InitializeProductionMCP(ctx context.Context, config MCPConfig) (MCPManager, error)
```
Creates a production-ready MCP manager with caching and error handling.

#### QuickStartMCP
```go
func QuickStartMCP() (MCPManager, error)
```
Creates an MCP manager with default settings for quick prototyping.

### Configuration Factories

#### NewProviderFromWorkingDir
```go
func NewProviderFromWorkingDir() (ModelProvider, error)
```
Automatically creates a provider from `agentflow.toml` in the working directory.

#### NewRunnerFromWorkingDir
```go
func NewRunnerFromWorkingDir() (Runner, error)
```
Creates a complete runner setup from configuration.

## Helper Functions

### MCP Integration Helpers

#### FormatToolsForPrompt
```go
func FormatToolsForPrompt(ctx context.Context, mgr MCPManager) string
```
Formats available tools for inclusion in LLM prompts.

**Returns**: Formatted string describing available tools that can be included in prompts.

**Example**:
```go
toolPrompt := agentflow.FormatToolsForPrompt(ctx, mcpManager)
fullPrompt := fmt.Sprintf("You are a helpful assistant.\n%s\nUser: %s", toolPrompt, userMessage)
```

#### ParseAndExecuteToolCalls
```go
func ParseAndExecuteToolCalls(ctx context.Context, mgr MCPManager, response string) []ToolResult
```
Parses LLM responses for tool calls and executes them.

**Parameters**:
- `ctx` - Context for cancellation
- `mgr` - MCP manager for tool execution
- `response` - LLM response that may contain tool calls

**Returns**: Array of tool execution results.

**Example**:
```go
response, _ := llm.Generate(ctx, prompt)
toolResults := agentflow.ParseAndExecuteToolCalls(ctx, mcpManager, response)

if len(toolResults) > 0 {
    // Tools were executed, synthesize results
    synthesisPrompt := fmt.Sprintf("Response: %s\nTool results: %v\nFinal answer:", response, toolResults)
    finalResponse, _ := llm.Generate(ctx, synthesisPrompt)
}
```

### Event Creation

#### NewEvent
```go
func NewEvent(eventType string, data EventData, metadata map[string]string) Event
```
Creates a new event with the specified data and metadata.

**Example**:
```go
eventData := agentflow.EventData{"message": "Hello, world!"}
metadata := map[string]string{"session_id": "123", "user_id": "456"}
event := agentflow.NewEvent("user-message", eventData, metadata)
```

### State Creation

#### NewState
```go
func NewState() State
```
Creates a new empty state instance.

**Example**:
```go
state := agentflow.NewState()
state.Set("conversation_history", []string{})
state.SetMeta("session_id", "abc123")
```

## Constants

### Hook Points
```go
const (
    HookBeforeEventHandling HookPoint = "BeforeEventHandling"
    HookAfterEventHandling  HookPoint = "AfterEventHandling"
    HookBeforeAgentRun      HookPoint = "BeforeAgentRun"
    HookAfterAgentRun       HookPoint = "AfterAgentRun"
    HookAgentError          HookPoint = "AgentError"
)
```

### Metadata Keys
```go
const (
    RouteMetadataKey = "route"
    SessionIDKey     = "session_id"
)
```

## Configuration Types

### AzureConfig
```go
type AzureConfig struct {
    APIKey      string
    Endpoint    string
    Deployment  string
    APIVersion  string
    Model       string
    MaxTokens   int
    Temperature float64
    Timeout     time.Duration
}
```

### OpenAIConfig
```go
type OpenAIConfig struct {
    APIKey      string
    Model       string
    MaxTokens   int
    Temperature float64
    Organization string
    Timeout     time.Duration
}
```

### OllamaConfig
```go
type OllamaConfig struct {
    Host          string
    Model         string
    Temperature   float64
    ContextWindow int
    Timeout       time.Duration
}
```

### MCPConfig
```go
type MCPConfig struct {
    Enabled           bool
    Servers           map[string]MCPServerConfig
    CacheEnabled      bool
    CacheTTL          time.Duration
    ConnectionTimeout time.Duration
    MaxRetries        int
}
```

### MCPServerConfig
```go
type MCPServerConfig struct {
    Command   string
    Args      []string
    Transport string
    Env       map[string]string
}
```

## Error Types

AgentFlow defines custom error types for better error handling:

```go
type ConfigurationError struct {
    Field   string
    Message string
}

type ProviderError struct {
    Provider string
    Type     string  // "authentication", "rate_limit", "timeout", etc.
    Message  string
}

type MCPError struct {
    Server  string
    Tool    string
    Message string
}
```

## Usage Examples

### Complete Agent Implementation

```go
package main

import (
    "context"
    "fmt"
    "log"
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

type ResearchAgent struct {
    llm        agentflow.ModelProvider
    mcpManager agentflow.MCPManager
    name       string
}

func NewResearchAgent(name string, llm agentflow.ModelProvider, mcp agentflow.MCPManager) *ResearchAgent {
    return &ResearchAgent{
        name:       name,
        llm:        llm,
        mcpManager: mcp,
    }
}

func (a *ResearchAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    // Extract user query
    data := event.GetData()
    query, ok := data["message"]
    if !ok {
        return agentflow.AgentResult{}, fmt.Errorf("missing message in event data")
    }

    // Build research-focused prompt
    systemPrompt := `You are a research agent. For any query:
1. Use search tools to find current information
2. Use fetch_content for specific URLs
3. Provide comprehensive, well-sourced answers`

    // Include available tools
    toolPrompt := agentflow.FormatToolsForPrompt(ctx, a.mcpManager)
    fullPrompt := fmt.Sprintf("%s\n\n%s\n\nResearch query: %s", systemPrompt, toolPrompt, query)

    // Generate response
    response, err := a.llm.Generate(ctx, fullPrompt)
    if err != nil {
        return agentflow.AgentResult{}, fmt.Errorf("LLM generation failed: %w", err)
    }

    // Execute any tool calls
    toolResults := agentflow.ParseAndExecuteToolCalls(ctx, a.mcpManager, response)
    
    // Synthesize final response if tools were used
    var finalResponse string
    if len(toolResults) > 0 {
        synthesisPrompt := fmt.Sprintf(`Research findings: %v
        
Please compile this into a comprehensive research report with:
1. Executive summary
2. Key findings with sources  
3. Detailed analysis
4. Implications and recommendations`, toolResults)
        
        finalResponse, err = a.llm.Generate(ctx, synthesisPrompt)
        if err != nil {
            finalResponse = response // Fallback to original
        }
    } else {
        finalResponse = response
    }

    // Update state
    state.Set("research_report", finalResponse)
    state.Set("tools_used", len(toolResults) > 0)
    state.SetMeta("processed_by", a.name)
    
    return agentflow.AgentResult{
        Result: finalResponse,
        State:  state,
    }, nil
}

func main() {
    ctx := context.Background()
    
    // Create provider from config
    provider, err := agentflow.NewProviderFromWorkingDir()
    if err != nil {
        log.Fatal("Failed to create provider:", err)
    }
    
    // Initialize MCP
    mcpManager, err := agentflow.InitializeProductionMCP(ctx, agentflow.MCPConfig{
        Enabled:      true,
        CacheEnabled: true,
        CacheTTL:     5 * time.Minute,
    })
    if err != nil {
        log.Fatal("Failed to initialize MCP:", err)
    }
    defer mcpManager.Close()
    
    // Create agent
    agent := NewResearchAgent("research-agent", provider, mcpManager)
    
    // Process query
    eventData := agentflow.EventData{"message": "Latest developments in Go programming language"}
    event := agentflow.NewEvent("research-query", eventData, nil)
    state := agentflow.NewState()
    
    result, err := agent.Run(ctx, event, state)
    if err != nil {
        log.Fatal("Agent execution failed:", err)
    }
    
    fmt.Printf("Research Report:\n%s\n", result.Result)
}
```

## Next Steps

- **[Agent Interface](agents.md)** - Detailed agent interface documentation
- **[MCP Integration](mcp.md)** - Complete MCP integration reference
- **[CLI Commands](cli.md)** - AgentFlow CLI reference
