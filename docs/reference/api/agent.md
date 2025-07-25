# Agent API

**Building individual agents and agent handlers**

This document covers the Agent API in AgenticGoKit, which provides the foundation for creating individual agents that can process events, maintain state, and participate in multi-agent orchestrations.

## 📋 Core Interfaces

### Agent Interface

The basic Agent interface for simple state transformations:

```go
type Agent interface {
    // Run processes the input State and returns an output State or an error
    Run(ctx context.Context, inputState State) (State, error)
    // Name returns the unique identifier name of the agent
    Name() string
}
```

### AgentHandler Interface

The enhanced AgentHandler interface for event-driven processing:

```go
type AgentHandler interface {
    Run(ctx context.Context, event Event, state State) (AgentResult, error)
}
```

### AgentHandlerFunc

A function type that implements AgentHandler:

```go
type AgentHandlerFunc func(ctx context.Context, event Event, state State) (AgentResult, error)

func (f AgentHandlerFunc) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    return f(ctx, event, state)
}
```

## 🚀 Basic Usage

### Creating a Simple Agent

```go
package main

import (
    "context"
    "fmt"
    "github.com/kunalkushwaha/agenticgokit/core"
)

// Method 1: Using AgentHandlerFunc
func main() {
    agent := core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
        // Get data from event
        message, ok := event.Data["message"].(string)
        if !ok {
            return core.AgentResult{}, fmt.Errorf("missing message in event")
        }
        
        // Process the message
        response := fmt.Sprintf("Hello, %s!", message)
        
        // Return result
        return core.AgentResult{
            Data: map[string]interface{}{
                "response": response,
                "processed_at": time.Now().Unix(),
            },
        }, nil
    })
    
    // Test the agent
    event := core.NewEvent("greeting", map[string]interface{}{
        "message": "World",
    })
    
    state := core.NewState()
    result, err := agent.Run(context.Background(), event, state)
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Response: %s\n", result.Data["response"])
}
```

## 🏗️ Agent Implementation Patterns

### Struct-Based Agent

```go
type ChatAgent struct {
    llm core.LLMProvider
    name string
}

func NewChatAgent(name string, llm core.LLMProvider) *ChatAgent {
    return &ChatAgent{
        name: name,
        llm:  llm,
    }
}

func (a *ChatAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Extract query from event
    query, ok := event.Data["query"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("missing query in event data")
    }
    
    // Get conversation history from state
    history, _ := state.Get("history")
    var messages []string
    if history != nil {
        messages = history.([]string)
    }
    
    // Build prompt with context
    prompt := buildPrompt(query, messages)
    
    // Call LLM
    response, err := a.llm.Complete(ctx, prompt)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("LLM error: %w", err)
    }
    
    // Update conversation history
    updatedHistory := append(messages, query, response)
    state.Set("history", updatedHistory)
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "response": response,
            "query":    query,
        },
    }, nil
}

func buildPrompt(query string, history []string) string {
    var prompt strings.Builder
    prompt.WriteString("You are a helpful assistant.\n\n")
    
    // Add conversation history
    for i := 0; i < len(history); i += 2 {
        if i+1 < len(history) {
            prompt.WriteString(fmt.Sprintf("User: %s\n", history[i]))
            prompt.WriteString(fmt.Sprintf("Assistant: %s\n\n", history[i+1]))
        }
    }
    
    prompt.WriteString(fmt.Sprintf("User: %s\n", query))
    return prompt.String()
}
```

### Stateless Processing Agent

```go
type DataProcessorAgent struct{}

func (a *DataProcessorAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Extract data from event
    data, ok := event.Data["data"].([]interface{})
    if !ok {
        return core.AgentResult{}, fmt.Errorf("missing data array in event")
    }
    
    // Process data (example: calculate sum)
    var sum float64
    var count int
    
    for _, item := range data {
        if num, ok := item.(float64); ok {
            sum += num
            count++
        }
    }
    
    var average float64
    if count > 0 {
        average = sum / float64(count)
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "sum":     sum,
            "count":   count,
            "average": average,
        },
    }, nil
}
```

### Tool-Using Agent

```go
type ResearchAgent struct {
    llm core.LLMProvider
    mcp core.MCPManager
}

func NewResearchAgent(llm core.LLMProvider, mcp core.MCPManager) *ResearchAgent {
    return &ResearchAgent{
        llm: llm,
        mcp: mcp,
    }
}

func (a *ResearchAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    query, ok := event.GetData()["query"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("missing query in event data")
    }
    
    // Get available tools
    tools, err := a.mcp.ListTools(ctx)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to list tools: %w", err)
    }
    
    // Build prompt with tool information
    toolPrompt := core.FormatToolsForPrompt(ctx, a.mcp)
    prompt := fmt.Sprintf(`You are a research assistant with access to tools.
    
Available tools:
%s

Research query: %s

Use the appropriate tools to research this query and provide a comprehensive answer.`, toolPrompt, query)
    
    // Generate response
    response, err := a.llm.Complete(ctx, prompt)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("LLM generation failed: %w", err)
    }
    
    // Execute any tool calls found in the response
    toolResults := core.ParseAndExecuteToolCalls(ctx, a.mcp, response)
    
    // If tools were used, synthesize final response
    var finalResponse string
    if len(toolResults) > 0 {
        synthesisPrompt := fmt.Sprintf(`Based on the research findings below, provide a comprehensive answer to: %s

Research findings:
%v

Please synthesize this information into a clear, well-structured response.`, query, toolResults)
        
        finalResponse, err = a.llm.Complete(ctx, synthesisPrompt)
        if err != nil {
            finalResponse = response // Fallback to original response
        }
    } else {
        finalResponse = response
    }
    
    // Store results in state
    state.Set("response", finalResponse)
    state.Set("tools_used", len(toolResults) > 0)
    state.Set("tool_count", len(toolResults))
    
    return core.AgentResult{
        OutputState: state,
    }, nil
}
```

## 📊 Agent Result Structure

### AgentResult Type

```go
type AgentResult struct {
    // OutputState contains the updated state after agent execution
    OutputState State `json:"output_state"`
    
    // Error contains any error message (empty string if successful)
    Error string `json:"error,omitempty"`
    
    // StartTime and EndTime track execution timing
    StartTime time.Time `json:"start_time"`
    EndTime   time.Time `json:"end_time"`
    
    // Duration tracks how long the agent took to execute
    Duration time.Duration `json:"duration"`
}
```

### Result Examples

#### Simple Response
```go
return core.AgentResult{
    Data: map[string]interface{}{
        "answer": "Paris is the capital of France",
        "confidence": 0.95,
    },
    Success: true,
}, nil
```

#### Response with State Updates
```go
// Update conversation state
state.Set("last_query", query)
state.Set("query_count", state.GetInt("query_count")+1)

return core.AgentResult{
    Data: map[string]interface{}{
        "response": answer,
    },
    State: state,
    Success: true,
}, nil
```

#### Response with Metadata
```go
return core.AgentResult{
    Data: map[string]interface{}{
        "response": answer,
    },
    Metadata: map[string]interface{}{
        "execution_time": time.Since(start),
        "tokens_used":    tokenCount,
        "model":          "gpt-4",
        "tools_called":   []string{"search", "calculator"},
    },
    Success: true,
}, nil
```

#### Partial Success with Errors
```go
return core.AgentResult{
    Data: map[string]interface{}{
        "response": partialAnswer,
    },
    Errors: []error{
        fmt.Errorf("tool 'advanced_search' failed: %w", searchErr),
        fmt.Errorf("cache miss for query: %s", query),
    },
    Success: true, // Still successful despite errors
}, nil
```

## 🔧 Agent Builder Pattern

### AgentBuilder Interface

```go
type AgentBuilder interface {
    WithName(name string) AgentBuilder
    WithLLM(provider LLMProvider) AgentBuilder
    WithMCP(config MCPConfig) AgentBuilder
    WithCapabilities(caps ...Capability) AgentBuilder
    WithMiddleware(middleware ...MiddlewareFunc) AgentBuilder
    Build() (Agent, error)
}
```

### Usage Example

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

## 🎭 Agent Capabilities

### Built-in Capabilities

#### SearchCapability
```go
agent := core.NewAgentBuilder().
    WithCapabilities(core.SearchCapability).
    Build()
```

#### CalculationCapability
```go
agent := core.NewAgentBuilder().
    WithCapabilities(core.CalculationCapability).
    Build()
```

#### MemoryCapability
```go
agent := core.NewAgentBuilder().
    WithCapabilities(core.MemoryCapability).
    Build()
```

#### FileCapability
```go
agent := core.NewAgentBuilder().
    WithCapabilities(core.FileCapability).
    Build()
```

### Custom Capabilities

```go
type CustomCapability struct {
    name string
}

func (c *CustomCapability) Name() string {
    return c.name
}

func (c *CustomCapability) Configure(agent Agent) error {
    // Configure the agent with custom functionality
    return nil
}

func (c *CustomCapability) Dependencies() []string {
    return []string{"SearchCapability"} // Depends on search
}
```

## 🔄 Agent Middleware

### Middleware Function Type

```go
type MiddlewareFunc func(next AgentHandler) AgentHandler
```

### Common Middleware Examples

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

### Test Utilities

```go
func TestChatAgent(t *testing.T) {
    // Create mock LLM
    mockLLM := &MockLLMProvider{
        responses: map[string]string{
            "Hello": "Hi there! How can I help you?",
        },
    }
    
    // Create agent
    agent := NewChatAgent("test-agent", mockLLM)
    
    // Create test event and state
    event := core.NewEvent("chat", map[string]interface{}{
        "query": "Hello",
    })
    state := core.NewState()
    
    // Run agent
    result, err := agent.Run(context.Background(), event, state)
    
    // Assertions
    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "Hi there! How can I help you?", result.Data["response"])
    
    // Verify state was updated
    history, exists := result.State.Get("history")
    assert.True(t, exists)
    assert.Equal(t, []string{"Hello", "Hi there! How can I help you?"}, history)
}
```

### Mock LLM Provider

```go
type MockLLMProvider struct {
    responses map[string]string
}

func (m *MockLLMProvider) Complete(ctx context.Context, prompt string) (string, error) {
    for key, response := range m.responses {
        if strings.Contains(prompt, key) {
            return response, nil
        }
    }
    return "I don't understand", nil
}

func (m *MockLLMProvider) Name() string {
    return "mock"
}
```

## 📚 Best Practices

### Error Handling

```go
func (a *MyAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Validate input
    query, ok := event.Data["query"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("missing or invalid query in event data")
    }
    
    if strings.TrimSpace(query) == "" {
        return core.AgentResult{}, fmt.Errorf("query cannot be empty")
    }
    
    // Process with error handling
    result, err := a.processQuery(ctx, query)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to process query: %w", err)
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "response": result,
        },
        Success: true,
    }, nil
}
```

### Context Handling

```go
func (a *MyAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Check if context is cancelled
    select {
    case <-ctx.Done():
        return core.AgentResult{}, ctx.Err()
    default:
    }
    
    // Pass context to all operations
    result, err := a.llm.Complete(ctx, prompt)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Check context again for long operations
    select {
    case <-ctx.Done():
        return core.AgentResult{}, ctx.Err()
    default:
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "response": result,
        },
        Success: true,
    }, nil
}
```

### State Management

```go
func (a *MyAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Read from state safely
    conversationHistory, _ := state.Get("conversation")
    var history []string
    if conversationHistory != nil {
        if h, ok := conversationHistory.([]string); ok {
            history = h
        }
    }
    
    // Process query
    query := event.Data["query"].(string)
    response, err := a.processWithHistory(ctx, query, history)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Update state
    updatedHistory := append(history, query, response)
    state.Set("conversation", updatedHistory)
    state.Set("last_interaction", time.Now().Unix())
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "response": response,
        },
        Success: true,
    }, nil
}
```

This comprehensive Agent API reference covers all aspects of building and using agents in AgenticGoKit, from basic implementations to advanced patterns and best practices.