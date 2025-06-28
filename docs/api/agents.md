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
    // Data contains the response data to be returned
    Data map[string]interface{} `json:"data"`
    
    // State contains any state changes (optional)
    State State `json:"state,omitempty"`
    
    // Metadata contains additional information about the execution
    Metadata map[string]interface{} `json:"metadata,omitempty"`
    
    // Errors contains any non-fatal errors that occurred
    Errors []error `json:"errors,omitempty"`
    
    // Success indicates if the operation was successful
    Success bool `json:"success"`
}
```

**Field Details:**

#### `Data`
The primary response data that will be returned to the caller:
```go
result := core.AgentResult{
    Data: map[string]interface{}{
        "answer":     "Paris is the capital of France",
        "confidence": 0.95,
        "sources":    []string{"wikipedia", "britannica"},
    },
}
```

#### `State`
Updated state to persist for the session:
```go
// Update conversation state
state.Set("last_query", query)
state.Set("query_count", state.GetInt("query_count")+1)

result := core.AgentResult{
    Data:  responseData,
    State: state,
}
```

#### `Metadata`
Additional execution information:
```go
result := core.AgentResult{
    Data: responseData,
    Metadata: map[string]interface{}{
        "execution_time": time.Since(start),
        "tokens_used":    tokenCount,
        "model":          "gpt-4o",
        "tools_called":   []string{"search", "calculator"},
    },
}
```

#### `Errors`
Non-fatal errors that occurred during processing:
```go
result := core.AgentResult{
    Data: partialData,
    Errors: []error{
        fmt.Errorf("tool 'advanced_search' failed: %w", searchErr),
        fmt.Errorf("cache miss for query: %s", query),
    },
    Success: true, // Still successful despite errors
}
```

## 🔧 Agent Builder and Factory

### `AgentBuilder`

Factory for creating configured agents with capabilities.

```go
type AgentBuilder interface {
    // WithName sets the agent name
    WithName(name string) AgentBuilder
    
    // WithLLM configures the LLM provider
    WithLLM(provider LLMProvider) AgentBuilder
    
    // WithMCP enables MCP tool integration
    WithMCP(config MCPConfig) AgentBuilder
    
    // WithCapabilities adds specific capabilities
    WithCapabilities(caps ...Capability) AgentBuilder
    
    // WithMiddleware adds middleware functions
    WithMiddleware(middleware ...MiddlewareFunc) AgentBuilder
    
    // Build creates the configured agent
    Build() (Agent, error)
}
```

**Usage Example:**
```go
agent, err := core.NewAgentBuilder().
    WithName("research-assistant").
    WithLLM(azureLLM).
    WithMCP(mcpConfig).
    WithCapabilities(
        core.SearchCapability,
        core.CalculationCapability,
        core.MemoryCapability,
    ).
    WithMiddleware(
        LoggingMiddleware,
        AuthenticationMiddleware,
        RateLimitMiddleware,
    ).
    Build()

if err != nil {
    log.Fatal("Failed to build agent:", err)
}
```

### `NewAgent`

Factory function for creating simple agents from handlers.

```go
func NewAgent(name string, handler AgentHandler) Agent
```

**Usage Example:**
```go
handler := &MyCustomHandler{}
agent := core.NewAgent("custom-agent", handler)

// Register with runner
runner.RegisterAgent(agent.Name(), handler)
```

## 🎭 Agent Capabilities

### `Capability`

Interface for extending agent functionality.

```go
type Capability interface {
    // Name returns the capability identifier
    Name() string
    
    // Configure applies the capability to an agent
    Configure(agent Agent) error
    
    // Dependencies returns required capabilities
    Dependencies() []string
}
```

### Built-in Capabilities

#### `SearchCapability`
Enables web search functionality:
```go
agent := core.NewAgentBuilder().
    WithCapabilities(core.SearchCapability).
    Build()
```

#### `CalculationCapability`
Adds mathematical computation tools:
```go
agent := core.NewAgentBuilder().
    WithCapabilities(core.CalculationCapability).
    Build()
```

#### `MemoryCapability`
Provides persistent memory across sessions:
```go
agent := core.NewAgentBuilder().
    WithCapabilities(core.MemoryCapability).
    Build()
```

#### `FileCapability`
Enables file system operations:
```go
agent := core.NewAgentBuilder().
    WithCapabilities(core.FileCapability).
    Build()
```

## 🔄 Agent Middleware

### `MiddlewareFunc`

Function type for implementing agent middleware.

```go
type MiddlewareFunc func(next AgentHandler) AgentHandler
```

### Common Middleware Patterns

#### Authentication Middleware
```go
func AuthenticationMiddleware(next core.AgentHandler) core.AgentHandler {
    return core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        // Check authentication
        userID := event.GetUserID()
        if !isAuthenticated(userID) {
            return core.AgentResult{}, fmt.Errorf("unauthorized")
        }
        
        return next.Run(ctx, event, state)
    })
}
```

#### Logging Middleware
```go
func LoggingMiddleware(next core.AgentHandler) core.AgentHandler {
    return core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        start := time.Now()
        
        log.Printf("Agent request started: %s", event.GetType())
        
        result, err := next.Run(ctx, event, state)
        
        duration := time.Since(start)
        if err != nil {
            log.Printf("Agent request failed: %s (duration: %v, error: %v)", event.GetType(), duration, err)
        } else {
            log.Printf("Agent request completed: %s (duration: %v)", event.GetType(), duration)
        }
        
        return result, err
    })
}
```

#### Rate Limiting Middleware
```go
func RateLimitMiddleware(limiter *rate.Limiter) func(core.AgentHandler) core.AgentHandler {
    return func(next core.AgentHandler) core.AgentHandler {
        return core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            if !limiter.Allow() {
                return core.AgentResult{}, fmt.Errorf("rate limit exceeded")
            }
            
            return next.Run(ctx, event, state)
        })
    }
}
```

#### Timeout Middleware
```go
func TimeoutMiddleware(timeout time.Duration) func(core.AgentHandler) core.AgentHandler {
    return func(next core.AgentHandler) core.AgentHandler {
        return core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
            defer cancel()
            
            type result struct {
                res core.AgentResult
                err error
            }
            
            resultChan := make(chan result, 1)
            
            go func() {
                res, err := next.Run(timeoutCtx, event, state)
                resultChan <- result{res, err}
            }()
            
            select {
            case r := <-resultChan:
                return r.res, r.err
            case <-timeoutCtx.Done():
                return core.AgentResult{}, fmt.Errorf("agent timeout after %v", timeout)
            }
        })
    }
}
```

## 🧪 Testing Agents

### `TestAgent`

Utility for testing agent implementations.

```go
type TestAgent struct {
    handler AgentHandler
    events  []Event
    states  []State
    results []AgentResult
}

func NewTestAgent(handler AgentHandler) *TestAgent {
    return &TestAgent{
        handler: handler,
        events:  make([]Event, 0),
        states:  make([]State, 0),
        results: make([]AgentResult, 0),
    }
}

func (ta *TestAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    ta.events = append(ta.events, event)
    ta.states = append(ta.states, state)
    
    result, err := ta.handler.Run(ctx, event, state)
    ta.results = append(ta.results, result)
    
    return result, err
}

func (ta *TestAgent) GetLastEvent() Event {
    if len(ta.events) == 0 {
        return nil
    }
    return ta.events[len(ta.events)-1]
}

func (ta *TestAgent) GetLastResult() AgentResult {
    if len(ta.results) == 0 {
        return AgentResult{}
    }
    return ta.results[len(ta.results)-1]
}
```

### Testing Example

```go
func TestChatAgent(t *testing.T) {
    // Create mock LLM
    mockLLM := &MockLLMProvider{
        responses: map[string]string{
            "Hello": "Hi there! How can I help you?",
        },
    }
    
    // Create agent
    agent := &ChatAgent{llm: mockLLM}
    testAgent := NewTestAgent(agent)
    
    // Create test event and state
    event := core.NewEvent("chat", map[string]interface{}{
        "query": "Hello",
    })
    state := core.NewState()
    
    // Run agent
    result, err := testAgent.Run(context.Background(), event, state)
    
    // Assertions
    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "Hi there! How can I help you?", result.Data["response"])
    
    // Verify state was updated
    history := result.State.GetStringSlice("history")
    assert.Equal(t, []string{"Hello", "Hi there! How can I help you?"}, history)
}
```

## 📚 Agent Patterns

### Stateless Agent Pattern
For simple, stateless operations:
```go
type StatelessAgent struct{}

func (a *StatelessAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Process event data without modifying state
    data := event.GetData()
    result := processData(data)
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "result": result,
        },
    }, nil
}
```

### Stateful Agent Pattern
For maintaining conversation context:
```go
type StatefulAgent struct{}

func (a *StatefulAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Read current state
    conversationHistory := state.GetStringSlice("conversation")
    
    // Process with context
    query := event.GetData()["query"].(string)
    response := processWithHistory(query, conversationHistory)
    
    // Update state
    updatedHistory := append(conversationHistory, query, response)
    state.Set("conversation", updatedHistory)
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "response": response,
        },
        State: state,
    }, nil
}
```

### Tool-Using Agent Pattern
For agents that use multiple tools:
```go
type ToolUsingAgent struct {
    toolRegistry ToolRegistry
}

func (a *ToolUsingAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    query := event.GetData()["query"].(string)
    
    // Determine which tools to use
    tools := a.selectTools(query)
    
    // Execute tools in parallel
    results := make(chan ToolResult, len(tools))
    for _, tool := range tools {
        go func(t Tool) {
            result := a.executeTool(ctx, t, query)
            results <- result
        }(tool)
    }
    
    // Collect results
    var toolResults []ToolResult
    for i := 0; i < len(tools); i++ {
        toolResults = append(toolResults, <-results)
    }
    
    // Synthesize final response
    response := a.synthesizeResults(toolResults)
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "response": response,
            "tools_used": extractToolNames(tools),
        },
    }, nil
}
```

This agent interface API reference provides comprehensive coverage of AgentFlow's agent system, from basic interfaces to advanced patterns and testing utilities.
