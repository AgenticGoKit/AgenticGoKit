# Core vs Internal Architecture

**Understanding AgentFlow's Package Structure**

AgentFlow uses a clear separation between public API (`core/`) and private implementation (`internal/`) to provide a stable, developer-friendly interface while maintaining implementation flexibility.

## Overview

```
agentflow/
├── core/           # Public API - what users import
│   ├── agent.go    # Agent interfaces and types
│   ├── mcp.go      # MCP integration public API  
│   ├── factory.go  # Factory functions for creating components
│   ├── llm.go      # LLM provider interfaces
│   └── ...         # Other public interfaces
└── internal/       # Private implementation - not importable
    ├── agents/     # Concrete agent implementations
    ├── mcp/        # MCP client and server management
    ├── llm/        # LLM provider implementations
    ├── orchestrator/ # Workflow orchestration logic
    └── ...         # Other implementation packages
```

## Design Principles

### 1. Interface Segregation

**Public interfaces are defined in `core/`:**

```go
// core/agent.go
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

type ModelProvider interface {
    Generate(ctx context.Context, prompt string) (string, error)
    GenerateWithHistory(ctx context.Context, messages []Message) (string, error)
    Name() string
}

type MCPManager interface {
    ListTools(ctx context.Context) ([]ToolSchema, error)
    CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
    // ... more methods
}
```

**Implementations are in `internal/`:**

```go
// internal/agents/mcp_agent.go
type mcpAgent struct {
    name       string
    llm        llm.Provider          // internal interface
    mcpManager mcp.Manager           // internal interface
}

func (a *mcpAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Implementation details hidden from users
}
```

### 2. Factory Pattern

**Factories in `core/` create internal implementations:**

```go
// core/factory.go
func NewMCPAgent(name string, llm ModelProvider, mcp MCPManager) AgentHandler {
    // Create internal implementation
    return agents.NewMCPAgent(name, llm, mcp)
}

func InitializeProductionMCP(ctx context.Context, config MCPConfig) (MCPManager, error) {
    // Create internal MCP manager
    return mcp.NewProductionManager(ctx, config)
}
```

This pattern allows users to work with interfaces while we manage complex implementations internally.

## Core Package Structure

### agent.go - Agent System

```go
// Primary interfaces for agent development
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

type Agent interface {
    Run(ctx context.Context, inputState State) (State, error)
    Name() string
}

// Supporting types
type Event interface {
    GetID() string
    GetData() EventData
    GetMetadata() map[string]string
    // ...
}

type State interface {
    Get(key string) (any, bool)
    Set(key string, value any)
    Clone() State
    // ...
}
```

### mcp.go - MCP Integration

```go
// Complete MCP public API
type MCPManager interface {
    // Server management
    Connect(ctx context.Context, serverName string) error
    Disconnect(serverName string) error
    
    // Tool discovery and execution
    ListTools(ctx context.Context) ([]ToolSchema, error)
    CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
    
    // Cache and configuration
    RefreshTools(ctx context.Context) error
    GetCacheStats() CacheStats
}

// Configuration types
type MCPConfig struct {
    Servers           map[string]MCPServerConfig
    CacheEnabled      bool
    CacheTTL          time.Duration
    ConnectionTimeout time.Duration
    MaxRetries        int
}

// Helper functions for agent development
func FormatToolsForPrompt(ctx context.Context, mgr MCPManager) string
func ParseAndExecuteToolCalls(ctx context.Context, mgr MCPManager, response string) []ToolResult
```

### llm.go - LLM Providers

```go
// Unified LLM interface
type ModelProvider interface {
    Generate(ctx context.Context, prompt string) (string, error)
    GenerateWithHistory(ctx context.Context, messages []Message) (string, error)
    Name() string
}

// Provider-specific configuration types
type AzureConfig struct {
    APIKey      string
    Endpoint    string
    Deployment  string
    APIVersion  string
    MaxTokens   int
    Temperature float64
}

type OpenAIConfig struct {
    APIKey      string
    Model       string
    MaxTokens   int
    Temperature float64
}

// And so on for other providers...
```

### factory.go - Creation Functions

```go
// Agent factories
func NewMCPAgent(name string, llm ModelProvider, mcp MCPManager) AgentHandler
func NewBasicAgent(name string, llm ModelProvider) AgentHandler

// Provider factories  
func NewAzureProvider(config AzureConfig) (ModelProvider, error)
func NewOpenAIProvider(config OpenAIConfig) (ModelProvider, error)
func NewOllamaProvider(config OllamaConfig) (ModelProvider, error)

// Configuration-driven factories
func NewProviderFromWorkingDir() (ModelProvider, error)
func NewProviderFromConfig(configPath string) (ModelProvider, error)

// MCP factories
func InitializeProductionMCP(ctx context.Context, config MCPConfig) (MCPManager, error)
func QuickStartMCP() (MCPManager, error)
```

## Internal Package Structure

### internal/agents/ - Agent Implementations

```go
// internal/agents/mcp_agent.go
package agents

type mcpAgent struct {
    name       string
    llm        llm.Provider     // Internal LLM interface
    mcpManager mcp.Manager      // Internal MCP interface
    logger     zerolog.Logger
}

func NewMCPAgent(name string, llm llm.Provider, mcp mcp.Manager) core.AgentHandler {
    return &mcpAgent{
        name:       name,
        llm:        llm,
        mcpManager: mcp,
        logger:     log.With().Str("agent", name).Logger(),
    }
}

func (a *mcpAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Complex implementation details
    // Error handling, logging, tool execution, etc.
}
```

### internal/mcp/ - MCP Implementation

```go
// internal/mcp/manager.go
package mcp

type Manager struct {
    clients    map[string]*client.MCPClient
    config     core.MCPConfig
    cache      *toolCache
    metrics    *mcpMetrics
    logger     zerolog.Logger
}

func NewProductionManager(ctx context.Context, config core.MCPConfig) core.MCPManager {
    mgr := &Manager{
        clients: make(map[string]*client.MCPClient),
        config:  config,
        cache:   newToolCache(config.CacheTTL),
        metrics: newMCPMetrics(),
        logger:  log.With().Str("component", "mcp-manager").Logger(),
    }
    
    // Initialize servers, set up monitoring, etc.
    return mgr
}

func (m *Manager) ListTools(ctx context.Context) ([]core.ToolSchema, error) {
    // Implementation with caching, error handling, metrics
}
```

### internal/llm/ - LLM Implementations

```go
// internal/llm/azure.go
package llm

type azureProvider struct {
    client      *openai.Client
    config      core.AzureConfig
    metrics     *llmMetrics
    rateLimiter *rate.Limiter
}

func NewAzureProvider(config core.AzureConfig) core.ModelProvider {
    return &azureProvider{
        client:      createAzureClient(config),
        config:      config,
        metrics:     newLLMMetrics("azure"),
        rateLimiter: rate.NewLimiter(rate.Limit(config.RequestsPerSecond), 1),
    }
}

func (p *azureProvider) Generate(ctx context.Context, prompt string) (string, error) {
    // Rate limiting, retries, error handling, metrics
}
```

## Benefits of This Architecture

### 1. Stable Public API

Users import only from `core/`:
```go
import agentflow "github.com/kunalkushwaha/agentflow/core"

// All user code works with interfaces
var agent agentflow.AgentHandler = agentflow.NewMCPAgent("my-agent", llm, mcp)
var provider agentflow.ModelProvider = agentflow.NewAzureProvider(config)
```

We can refactor `internal/` without breaking user code.

### 2. Implementation Flexibility

We can:
- Optimize internal algorithms
- Change data structures
- Add new features to implementations
- Fix bugs in internal logic

Without affecting users.

### 3. Testing Boundaries

**Public API Testing** (what users care about):
```go
func TestMCPAgent_PublicBehavior(t *testing.T) {
    // Test through public interfaces only
    agent := core.NewMCPAgent("test", mockLLM, mockMCP)
    result, err := agent.Run(ctx, event, state)
    
    // Assert on public behavior
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Result)
}
```

**Internal Testing** (implementation details):
```go
func TestMCPManager_CacheLogic(t *testing.T) {
    // Test internal implementation details
    mgr := mcp.NewManager(config)
    
    // Test caching behavior, error handling, etc.
}
```

### 4. Clear Documentation Focus

**User Documentation**: Focus on `core/` package
**Developer Documentation**: Cover `internal/` architecture

## Guidelines for Development

### When to Add to `core/`

Add to `core/` when:
- Users need to interact with the functionality
- It's part of the public contract
- It needs to be stable across versions
- It's a configuration type or interface

```go
// Good: User-facing interface
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}

// Good: Configuration type
type MCPConfig struct {
    Servers      map[string]MCPServerConfig
    CacheEnabled bool
    CacheTTL     time.Duration
}

// Good: Factory function
func NewMCPAgent(name string, llm ModelProvider, mcp MCPManager) AgentHandler
```

### When to Keep in `internal/`

Keep in `internal/` when:
- It's implementation detail
- It might change frequently
- Users don't need direct access
- It's complex business logic

```go
// Good: Implementation detail
type mcpClient struct {
    conn       net.Conn
    encoder    *json.Encoder
    decoder    *json.Decoder
    msgID      int64
    pending    map[int64]chan<- mcpResponse
}

// Good: Internal algorithm
func (c *mcpClient) sendRequest(req mcpRequest) (mcpResponse, error) {
    // Complex protocol handling
}
```

### Interface Design Patterns

**Do:**
```go
// Small, focused interfaces
type ToolExecutor interface {
    ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
}

// Composition of interfaces
type MCPManager interface {
    ToolDiscoverer
    ToolExecutor
    ConnectionManager
}
```

**Don't:**
```go
// Large, monolithic interfaces
type EverythingManager interface {
    // 20+ methods that do different things
}
```

## Migration and Versioning

### Adding New Features

1. **Add interface to `core/`**
2. **Implement in `internal/`**
3. **Add factory in `core/`**
4. **Maintain backward compatibility**

```go
// core/agent.go - Add new interface
type StreamingAgent interface {
    AgentHandler
    RunStreaming(ctx context.Context, event Event, state State) (<-chan AgentResult, error)
}

// internal/agents/streaming_agent.go - Implement
type streamingAgent struct {
    // Implementation
}

// core/factory.go - Add factory
func NewStreamingAgent(name string, llm ModelProvider) StreamingAgent {
    return agents.NewStreamingAgent(name, llm)
}
```

### Deprecating Features

1. **Mark as deprecated in `core/`**
2. **Keep implementation working**
3. **Provide migration path**
4. **Remove in next major version**

```go
// Deprecated: Use NewMCPAgent instead.
func NewBasicAgent(name string) AgentHandler {
    return NewMCPAgent(name, nil, nil)
}
```

## Performance Considerations

### Interface Overhead

Go interfaces have minimal overhead:
- Method calls through interfaces are fast
- Interface conversions are optimized
- The separation doesn't impact performance

### Memory Management

- Interfaces don't increase memory usage significantly
- Internal implementations can be optimized independently
- Factory functions don't add overhead

### Compilation Benefits

- Users only compile against `core/` interfaces
- `internal/` changes don't trigger user recompilation
- Faster development iteration

## Next Steps

- **[Contributor Guide](ContributorGuide.md)** - Get started with development
- **[Adding Features](AddingFeatures.md)** - Learn how to extend AgentFlow
- **[Testing Strategy](Testing.md)** - Understand our testing approach
