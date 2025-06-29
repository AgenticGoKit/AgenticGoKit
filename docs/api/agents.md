# Agent Interface API Reference

This document provides comprehensive API reference for AgentFlow's agent interfaces, types, and related functionality.

## 🏗️ Core Agent Interfaces

### `Agent`

The foundational interface for any component that can process state.

```go
type Agent interface {
    // Run processes the input State and returns an output State or an error.
    // The context can be used for cancellation or deadlines.
    Run(ctx context.Context, inputState State) (State, error)
    
    // Name returns the unique identifier name of the agent.
    Name() string
}
```

**Usage Example:**
```go
type SimpleAgent struct {
    name string
}

func (a *SimpleAgent) Name() string {
    return a.name
}

func (a *SimpleAgent) Run(ctx context.Context, inputState core.State) (core.State, error) {
    // Process the state
    query := inputState.GetString("query")
    result := fmt.Sprintf("Processed: %s", query)
    
    // Create output state
    outputState := inputState.Clone()
    outputState.Set("result", result)
    
    return outputState, nil
}
```

### `AgentHandler`

The primary interface for implementing event-driven agent logic.

```go
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}
```

**Key Components:**
- **Context**: For cancellation, timeouts, and request tracing
- **Event**: Contains the input data and metadata
- **State**: Represents the current conversation/session state
- **AgentResult**: The response data and any state changes

**Usage Example:**
```go
type ChatAgent struct {
    llm LLMProvider
}

func (a *ChatAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Extract query from event
    data := event.GetData()
    query, ok := data["query"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("missing query in event data")
    }
    
    // Get conversation history from state
    history := state.GetStringSlice("history")
    
    // Build prompt with context
    prompt := buildPrompt(query, history)
    
    // Call LLM
    response, err := a.llm.Complete(ctx, prompt)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("LLM error: %w", err)
    }
    
    // Update state with new exchange
    updatedHistory := append(history, query, response)
    state.Set("history", updatedHistory)
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "response": response,
            "query":    query,
        },
        State: state,
    }, nil
}
```

### `AgentHandlerFunc`

Function adapter for implementing `AgentHandler` with a simple function.

```go
type AgentHandlerFunc func(ctx context.Context, event Event, state State) (AgentResult, error)

func (f AgentHandlerFunc) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    return f(ctx, event, state)
}
```

**Usage Example:**
```go
// Simple function-based agent
echoHandler := core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    data := event.GetData()
    return core.AgentResult{
        Data: map[string]interface{}{
            "echo": data,
            "timestamp": time.Now(),
        },
    }, nil
})

// Register with runner
runner.RegisterAgent("echo", echoHandler)
```

## 📊 Agent Result Types

### `AgentResult`

The response structure returned by agent handlers.

```go
type AgentResult struct {
    // OutputState contains the updated state after processing
    OutputState State `json:"output_state"`
    
    // Error contains any error message that occurred during processing
    Error string `json:"error,omitempty"`
    
    // StartTime indicates when the agent processing began
    StartTime time.Time `json:"start_time"`
    
    // EndTime indicates when the agent processing completed
    EndTime time.Time `json:"end_time"`
    
    // Duration is the total processing time
    Duration time.Duration
}
```

**Field Details:**

#### `OutputState`
The updated state after processing:
```go
// Agent updates state and returns it
outputState := inputState.Clone()
outputState.Set("result", "Processed successfully")
outputState.Set("last_query", query)

result := core.AgentResult{
    OutputState: outputState,
    StartTime:   start,
    EndTime:     time.Now(),
    Duration:    time.Since(start),
}
```

#### `Error`
String representation of any error that occurred:
```go
if err != nil {
    return core.AgentResult{
        OutputState: inputState,
        Error:       err.Error(),
        StartTime:   start,
        EndTime:     time.Now(),
        Duration:    time.Since(start),
    }
}
```

#### Timing Fields
Automatic timing information for monitoring and debugging:
```go
start := time.Now()
// ... processing ...
result := core.AgentResult{
    OutputState: processedState,
    StartTime:   start,
    EndTime:     time.Now(),
    Duration:    time.Since(start),
}
```

## 🔧 Agent Builder and Factory

### `AgentBuilder`

Struct for creating configured agents with capabilities using a fluent interface.

```go
type AgentBuilder struct {
    name         string
    capabilities []AgentCapability
    errors       []error
    config       AgentBuilderConfig
}
```

### Builder Configuration

```go
type AgentBuilderConfig struct {
    ValidateCapabilities bool // Whether to validate capability combinations
    SortByPriority       bool // Whether to sort capabilities by priority
    StrictMode           bool // Whether to fail on any capability error
}
```

### Builder Methods

#### Constructor
```go
// NewAgent creates a new agent builder with the specified name
func NewAgent(name string) *AgentBuilder

// NewAgentWithConfig creates a new agent builder with custom configuration
func NewAgentWithConfig(name string, config AgentBuilderConfig) *AgentBuilder
```

#### Capability Methods
```go
// WithLLM adds LLM capability to the agent
func (b *AgentBuilder) WithLLM(provider ModelProvider) *AgentBuilder

// WithLLMAndConfig adds LLM capability with custom configuration
func (b *AgentBuilder) WithLLMAndConfig(provider ModelProvider, config LLMConfig) *AgentBuilder

// WithMCP adds MCP capability to the agent
func (b *AgentBuilder) WithMCP(manager MCPManager) *AgentBuilder

// WithMCPAndConfig adds MCP capability with custom configuration
func (b *AgentBuilder) WithMCPAndConfig(manager MCPManager, config MCPAgentConfig) *AgentBuilder

// WithMCPAndCache adds MCP capability with caching
func (b *AgentBuilder) WithMCPAndCache(manager MCPManager, cacheManager MCPCacheManager) *AgentBuilder

// WithCache adds cache capability to the agent
func (b *AgentBuilder) WithCache(manager interface{}, config interface{}) *AgentBuilder

// WithMetrics adds metrics capability to the agent
func (b *AgentBuilder) WithMetrics(config MetricsConfig) *AgentBuilder

// WithDefaultMetrics adds metrics capability with default configuration
func (b *AgentBuilder) WithDefaultMetrics() *AgentBuilder

// WithCapability adds a custom capability to the agent
func (b *AgentBuilder) WithCapability(capability AgentCapability) *AgentBuilder

// WithCapabilities adds multiple capabilities to the agent
func (b *AgentBuilder) WithCapabilities(capabilities ...AgentCapability) *AgentBuilder
```

#### Build Methods
```go
// Build creates the final agent with all configured capabilities
func (b *AgentBuilder) Build() (Agent, error)

// BuildOrPanic builds the agent and panics if there are any errors
func (b *AgentBuilder) BuildOrPanic() Agent
```

**Usage Example:**
```go
agent, err := core.NewAgent("research-assistant").
    WithLLM(azureLLM).
    WithMCP(mcpManager).
    WithMetrics(metricsConfig).
    Build()

if err != nil {
    log.Fatal("Failed to build agent:", err)
}
```

### Convenience Factory Functions

#### `NewMCPEnabledAgent`
Creates an agent with MCP and LLM capabilities:
```go
func NewMCPEnabledAgent(name string, mcpManager MCPManager, llmProvider ModelProvider) (Agent, error)

// Usage
agent, err := core.NewMCPEnabledAgent("assistant", mcpManager, llmProvider)
```

#### `NewAgentWithCapabilities`
Creates an agent with specific capabilities:
```go
func NewAgentWithCapabilities(name string, capabilities ...AgentCapability) (Agent, error)

// Usage
agent, err := core.NewAgentWithCapabilities(
    "multi-capability-agent",
    llmCapability,
    mcpCapability,
    metricsCapability,
)
```

#### `NewAgentFromConfig`
Creates an agent from configuration (placeholder implementation):
```go
func NewAgentFromConfig(name string, config SimpleAgentConfig) (Agent, error)

// Usage with TOML config
type SimpleAgentConfig struct {
    Name           string         `toml:"name"`
    LLM            *LLMConfig     `toml:"llm"`
    MCP            *MCPConfig     `toml:"mcp"`
    Metrics        *MetricsConfig `toml:"metrics"`
    LLMEnabled     bool           `toml:"llm_enabled"`
    MCPEnabled     bool           `toml:"mcp_enabled"`
    MetricsEnabled bool           `toml:"metrics_enabled"`
}
```

## 🎭 Agent Capabilities

### `AgentCapability`

Interface for extending agent functionality through composable capabilities.

```go
type AgentCapability interface {
    // Name returns a unique identifier for this capability
    Name() string
    
    // Configure applies this capability to an agent during creation
    Configure(agent CapabilityConfigurable) error
    
    // Validate checks if this capability can be applied with others
    Validate(others []AgentCapability) error
    
    // Priority returns the initialization priority for this capability
    Priority() int
}
```

### `CapabilityConfigurable`

Interface that agents must implement to support capabilities.

```go
type CapabilityConfigurable interface {
    // SetLLMProvider sets the LLM provider for the agent
    SetLLMProvider(provider ModelProvider, config LLMConfig)
    
    // SetCacheManager sets the cache manager for the agent
    SetCacheManager(manager interface{}, config interface{})
    
    // SetMetricsConfig sets the metrics configuration for the agent
    SetMetricsConfig(config MetricsConfig)
    
    // GetLogger returns the agent's logger for capability configuration
    GetLogger() *zerolog.Logger
}
```

### Built-in Capability Types

#### `CapabilityType`
Defines the type of capability for validation and ordering:
```go
type CapabilityType string

const (
    CapabilityTypeLLM     CapabilityType = "llm"
    CapabilityTypeMCP     CapabilityType = "mcp"
    CapabilityTypeCache   CapabilityType = "cache"
    CapabilityTypeMetrics CapabilityType = "metrics"
)
```

### Creating Capabilities

#### LLM Capability
```go
// NewLLMCapability creates a new LLM capability
func NewLLMCapability(provider ModelProvider, config LLMConfig) *LLMCapability

// Usage in agent builder
agent := core.NewAgent("llm-agent").
    WithLLM(azureProvider).
    Build()
```

#### MCP Capability
```go
// NewMCPCapability creates a new MCP capability
func NewMCPCapability(manager MCPManager, config MCPAgentConfig) *MCPCapability

// NewMCPCapabilityWithCache creates MCP capability with caching
func NewMCPCapabilityWithCache(manager MCPManager, config MCPAgentConfig, cacheManager MCPCacheManager) *MCPCapability

// Usage in agent builder
agent := core.NewAgent("mcp-agent").
    WithMCP(mcpManager).
    Build()
```

#### Cache Capability
```go
// NewCacheCapability creates a new cache capability
func NewCacheCapability(manager interface{}, config interface{}) *CacheCapability

// Usage in agent builder
agent := core.NewAgent("cached-agent").
    WithCache(cacheManager, cacheConfig).
    Build()
```

#### Metrics Capability
```go
// NewMetricsCapability creates a new metrics capability
func NewMetricsCapability(config MetricsConfig) *MetricsCapability

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() MetricsConfig

// Usage in agent builder
agent := core.NewAgent("monitored-agent").
    WithMetrics(metricsConfig).
    // Or use defaults
    WithDefaultMetrics().
    Build()
```

### Custom Capabilities

#### Creating Custom Capabilities
```go
type CustomCapability struct {
    name     string
    priority int
}

func (c *CustomCapability) Name() string {
    return c.name
}

func (c *CustomCapability) Priority() int {
    return c.priority
}

func (c *CustomCapability) Configure(agent CapabilityConfigurable) error {
    // Configure the agent with this capability
    logger := agent.GetLogger()
    logger.Info().Str("capability", c.name).Msg("Configuring custom capability")
    return nil
}

func (c *CustomCapability) Validate(others []AgentCapability) error {
    // Check compatibility with other capabilities
    return nil
}

// Usage
customCap := &CustomCapability{name: "custom", priority: 100}
agent := core.NewAgent("custom-agent").
    WithCapability(customCap).
    Build()
```

## 🔄 Agent Middleware

**Note:** Middleware functionality is not currently implemented in the core AgentBuilder. The following represents planned functionality that may be available through the Runner system.

### `MiddlewareFunc`

Function type for implementing agent middleware (planned functionality).

```go
type MiddlewareFunc func(next AgentHandler) AgentHandler
```

### Current State

The AgentBuilder does not currently support middleware through `WithMiddleware()`. Middleware functionality may be available through:

1. **Runner-level middleware** (see Runner documentation)
2. **Custom capability implementations** that wrap agent functionality
3. **Future AgentBuilder enhancements**

### Alternative: Capability-Based Middleware

You can achieve middleware-like functionality using custom capabilities:

```go
type LoggingCapability struct {
    name string
}

func (lc *LoggingCapability) Name() string {
    return lc.name
}

func (lc *LoggingCapability) Configure(agent CapabilityConfigurable) error {
    logger := agent.GetLogger()
    logger.Info().Msg("Logging capability configured")
    return nil
}

func (lc *LoggingCapability) Validate(others []AgentCapability) error {
    return nil
}

func (lc *LoggingCapability) Priority() int {
    return 1000 // High priority to configure early
}

// Usage
agent := core.NewAgent("logged-agent").
    WithCapability(&LoggingCapability{name: "logging"}).
    WithLLM(provider).
    Build()
```

### Future Middleware Support

When middleware support is added to AgentBuilder, it will likely follow this pattern:

```go
// Future implementation (not currently available)
agent := core.NewAgent("middleware-agent").
    WithLLM(provider).
    WithMiddleware(
        LoggingMiddleware,
        AuthenticationMiddleware,
        RateLimitMiddleware,
    ).
    Build()
```

## 🧪 Testing Agents

### Testing with AgentBuilder

The AgentBuilder provides several methods for testing and validation:

```go
func TestAgentBuilder(t *testing.T) {
    // Create a builder
    builder := core.NewAgent("test-agent").
        WithLLM(mockLLMProvider).
        WithDefaultMetrics()
    
    // Test validation
    err := builder.Validate()
    assert.NoError(t, err)
    
    // Test capability inspection
    assert.True(t, builder.HasCapability(core.CapabilityTypeLLM))
    assert.Equal(t, 2, builder.CapabilityCount())
    
    // Build the agent
    agent, err := builder.Build()
    assert.NoError(t, err)
    assert.NotNil(t, agent)
}
```

### Testing Agent Execution

```go
func TestAgentExecution(t *testing.T) {
    // Create mock provider
    mockLLM := &MockLLMProvider{
        responses: map[string]string{
            "Hello": "Hi there! How can I help you?",
        },
    }
    
    // Create agent
    agent, err := core.NewAgent("test-agent").
        WithLLM(mockLLM).
        Build()
    require.NoError(t, err)
    
    // Create test state
    inputState := core.NewState()
    inputState.Set("query", "Hello")
    
    // Run agent
    outputState, err := agent.Run(context.Background(), inputState)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, outputState)
    assert.Equal(t, "test-agent", outputState.GetString("processed_by"))
}
```

### Mock Implementations

#### Mock LLM Provider
```go
type MockLLMProvider struct {
    responses map[string]string
    callCount int
}

func (m *MockLLMProvider) Complete(ctx context.Context, prompt string) (string, error) {
    m.callCount++
    if response, exists := m.responses[prompt]; exists {
        return response, nil
    }
    return "Mock response", nil
}

func (m *MockLLMProvider) GetCallCount() int {
    return m.callCount
}
```

#### Mock MCP Manager
```go
type MockMCPManager struct {
    tools []string
}

func (m *MockMCPManager) ListTools() []string {
    return m.tools
}

func (m *MockMCPManager) ExecuteTool(name string, args map[string]interface{}) (interface{}, error) {
    return map[string]interface{}{
        "tool":   name,
        "args":   args,
        "result": "mock result",
    }, nil
}
```

### Integration Testing

```go
func TestAgentIntegration(t *testing.T) {
    // Create real providers for integration testing
    llmProvider := createTestLLMProvider()
    mcpManager := createTestMCPManager()
    
    // Create agent with real capabilities
    agent, err := core.NewMCPEnabledAgent(
        "integration-test-agent",
        mcpManager,
        llmProvider,
    )
    require.NoError(t, err)
    
    // Test with real data
    inputState := core.NewState()
    inputState.Set("query", "What is the weather like?")
    
    outputState, err := agent.Run(context.Background(), inputState)
    
    assert.NoError(t, err)
    assert.NotNil(t, outputState)
    // Add more specific assertions based on expected behavior
}
```

### Builder Validation Testing

```go
func TestBuilderValidation(t *testing.T) {
    tests := []struct {
        name        string
        setupFunc   func() *core.AgentBuilder
        expectError bool
    }{
        {
            name: "valid configuration",
            setupFunc: func() *core.AgentBuilder {
                return core.NewAgent("valid").WithLLM(mockProvider)
            },
            expectError: false,
        },
        {
            name: "nil LLM provider",
            setupFunc: func() *core.AgentBuilder {
                return core.NewAgent("invalid").WithLLM(nil)
            },
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            builder := tt.setupFunc()
            _, err := builder.Build()
            
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 📚 Agent Patterns

### Simple Agent Pattern
Using AgentBuilder for basic functionality:
```go
type SimpleAgent struct {
    name string
}

// Create using AgentBuilder
agent, err := core.NewAgent("simple-agent").
    WithLLM(llmProvider).
    Build()

// The agent implements core.Agent interface with Run() method
outputState, err := agent.Run(ctx, inputState)
```

### Capability-Rich Agent Pattern
For agents with multiple capabilities:
```go
// Create agent with multiple capabilities
agent, err := core.NewAgent("rich-agent").
    WithLLM(llmProvider).
    WithMCP(mcpManager).
    WithMetrics(metricsConfig).
    WithCache(cacheManager, cacheConfig).
    Build()

if err != nil {
    log.Fatal("Failed to build agent:", err)
}

// Use the agent
inputState := core.NewState()
inputState.Set("query", "Process this request")

outputState, err := agent.Run(context.Background(), inputState)
if err != nil {
    log.Fatal("Agent execution failed:", err)
}

// Access results from output state
result := outputState.GetString("result")
```

### MCP-Enabled Agent Pattern
For agents that use MCP tools:
```go
// Create MCP manager
mcpManager := createMCPManager()

// Create LLM provider
llmProvider := createLLMProvider()

// Create agent with MCP capabilities
agent, err := core.NewMCPEnabledAgent("mcp-agent", mcpManager, llmProvider)
if err != nil {
    log.Fatal("Failed to create MCP agent:", err)
}

// Use the agent
inputState := core.NewState()
inputState.Set("query", "Search for information about Go programming")

outputState, err := agent.Run(context.Background(), inputState)
if err != nil {
    log.Fatal("MCP agent execution failed:", err)
}
```

### Production Agent Pattern
For production deployments with monitoring:
```go
// Create production agent with all features
agent, err := core.NewAgent("production-agent").
    WithLLM(llmProvider).
    WithMCP(mcpManager).
    WithDefaultMetrics().
    WithCache(cacheManager, cacheConfig).
    WithValidation(true).
    WithStrictMode(true).
    Build()

if err != nil {
    log.Fatal("Failed to build production agent:", err)
}

// Monitor agent capabilities
builder := core.NewAgent("monitor").WithDefaultMetrics()
fmt.Printf("Agent has %d capabilities\n", builder.CapabilityCount())
fmt.Printf("Capabilities: %v\n", builder.ListCapabilities())
```

### Configuration-Driven Agent Pattern
For agents created from configuration files:
```go
// Load configuration from TOML
config := SimpleAgentConfig{
    Name:           "config-agent",
    LLMEnabled:     true,
    MCPEnabled:     true,
    MetricsEnabled: true,
}

// Create agent from configuration
agent, err := core.NewAgentFromConfig("config-agent", config)
if err != nil {
    log.Fatal("Failed to create agent from config:", err)
}

// Note: This is a placeholder implementation
// Full implementation requires provider creation from config
```

This agent interface API reference provides comprehensive coverage of AgentFlow's agent system, from basic interfaces to advanced patterns and testing utilities.
