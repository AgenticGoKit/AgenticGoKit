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

### Creating a Struct-Based Agent

```go
type GreetingAgent struct {
    name     string
    language string
}

func NewGreetingAgent(name, language string) *GreetingAgent {
    return &GreetingAgent{
        name:     name,
        language: language,
    }
}

func (g *GreetingAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    message, ok := event.Data["message"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("missing message in event")
    }
    
    var greeting string
    switch g.language {
    case "spanish":
        greeting = fmt.Sprintf("¡Hola, %s!", message)
    case "french":
        greeting = fmt.Sprintf("Bonjour, %s!", message)
    default:
        greeting = fmt.Sprintf("Hello, %s!", message)
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "response": greeting,
            "language": g.language,
            "agent":    g.name,
        },
    }, nil
}

// Usage
func main() {
    agent := NewGreetingAgent("greeter", "spanish")
    
    event := core.NewEvent("greeting", map[string]interface{}{
        "message": "María",
    })
    
    result, _ := agent.Run(context.Background(), event, core.NewState())
    fmt.Printf("Response: %s\n", result.Data["response"]) // "¡Hola, María!"
}
```

## 🔄 Agent Patterns

### Stateful Agent

```go
type CounterAgent struct {
    name  string
    count int
    mutex sync.Mutex
}

func NewCounterAgent(name string) *CounterAgent {
    return &CounterAgent{
        name: name,
    }
}

func (c *CounterAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    // Increment counter
    c.count++
    
    // Get increment value from event (default to 1)
    increment := 1
    if inc, ok := event.Data["increment"].(int); ok {
        increment = inc
        c.count += increment - 1 // Adjust since we already incremented by 1
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "count":     c.count,
            "increment": increment,
            "agent":     c.name,
        },
    }, nil
}

// Usage
func statefulAgentExample() {
    agent := NewCounterAgent("counter")
    
    // First call
    event1 := core.NewEvent("count", map[string]interface{}{
        "increment": 5,
    })
    result1, _ := agent.Run(context.Background(), event1, core.NewState())
    fmt.Printf("Count: %d\n", result1.Data["count"]) // Count: 5
    
    // Second call - state is maintained
    event2 := core.NewEvent("count", map[string]interface{}{
        "increment": 3,
    })
    result2, _ := agent.Run(context.Background(), event2, core.NewState())
    fmt.Printf("Count: %d\n", result2.Data["count"]) // Count: 8
}
```

### Configurable Agent

```go
type ProcessorAgentConfig struct {
    MaxLength    int           `json:"max_length"`
    Timeout      time.Duration `json:"timeout"`
    EnableCache  bool          `json:"enable_cache"`
    OutputFormat string        `json:"output_format"`
}

type ProcessorAgent struct {
    name   string
    config ProcessorAgentConfig
    cache  map[string]interface{}
    mutex  sync.RWMutex
}

func NewProcessorAgent(name string, config ProcessorAgentConfig) *ProcessorAgent {
    return &ProcessorAgent{
        name:   name,
        config: config,
        cache:  make(map[string]interface{}),
    }
}

func (p *ProcessorAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Apply timeout from config
    ctx, cancel := context.WithTimeout(ctx, p.config.Timeout)
    defer cancel()
    
    input, ok := event.Data["input"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("missing input in event")
    }
    
    // Check cache if enabled
    if p.config.EnableCache {
        p.mutex.RLock()
        if cached, exists := p.cache[input]; exists {
            p.mutex.RUnlock()
            return core.AgentResult{
                Data: map[string]interface{}{
                    "output": cached,
                    "cached": true,
                },
            }, nil
        }
        p.mutex.RUnlock()
    }
    
    // Apply max length limit
    if len(input) > p.config.MaxLength {
        input = input[:p.config.MaxLength]
    }
    
    // Process input (simulate work)
    processed := strings.ToUpper(input)
    
    // Format output based on config
    var output interface{}
    switch p.config.OutputFormat {
    case "json":
        output = map[string]string{"processed": processed}
    case "xml":
        output = fmt.Sprintf("<result>%s</result>", processed)
    default:
        output = processed
    }
    
    // Cache result if enabled
    if p.config.EnableCache {
        p.mutex.Lock()
        p.cache[input] = output
        p.mutex.Unlock()
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "output": output,
            "cached": false,
        },
    }, nil
}

// Usage
func configurableAgentExample() {
    config := ProcessorAgentConfig{
        MaxLength:    100,
        Timeout:      5 * time.Second,
        EnableCache:  true,
        OutputFormat: "json",
    }
    
    agent := NewProcessorAgent("processor", config)
    
    event := core.NewEvent("process", map[string]interface{}{
        "input": "hello world",
    })
    
    result, _ := agent.Run(context.Background(), event, core.NewState())
    fmt.Printf("Output: %+v\n", result.Data["output"])
}
```

## 🧪 Testing Agents

### Unit Testing

```go
func TestGreetingAgent(t *testing.T) {
    agent := NewGreetingAgent("test-greeter", "english")
    
    tests := []struct {
        name     string
        event    core.Event
        expected string
    }{
        {
            name: "simple greeting",
            event: core.NewEvent("greeting", map[string]interface{}{
                "message": "World",
            }),
            expected: "Hello, World!",
        },
        {
            name: "empty message",
            event: core.NewEvent("greeting", map[string]interface{}{
                "message": "",
            }),
            expected: "Hello, !",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := agent.Run(context.Background(), tt.event, core.NewState())
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result.Data["response"])
            assert.Equal(t, "english", result.Data["language"])
        })
    }
}
```

## 📚 Best Practices

### 1. Agent Design Principles

- **Single Responsibility**: Each agent should have one clear purpose
- **Stateless When Possible**: Prefer stateless agents for better scalability
- **Error Handling**: Always handle errors gracefully and provide meaningful messages
- **Context Awareness**: Respect context cancellation and timeouts

### 2. Performance Considerations

- **Use Context**: Always respect context cancellation
- **Avoid Blocking**: Don't block on long-running operations without context
- **Resource Management**: Clean up resources properly
- **Concurrent Safety**: Use proper synchronization for shared state

### 3. Error Handling Patterns

```go
// Good: Structured error handling
func (a *Agent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Validate inputs
    input, ok := event.Data["input"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("invalid input type: expected string, got %T", event.Data["input"])
    }
    
    if len(input) == 0 {
        return core.AgentResult{}, fmt.Errorf("input cannot be empty")
    }
    
    // Process with error handling
    result, err := processInput(input)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("processing failed: %w", err)
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            "result": result,
        },
    }, nil
}
```

## 🔗 Related APIs

- **[Orchestration API](orchestration.md)** - Multi-agent coordination
- **[State & Event API](state-event.md)** - Data flow between agents
- **[Memory API](memory.md)** - Persistent agent memory
- **[Configuration API](configuration.md)** - Agent configuration

---

*This documentation covers the current Agent API in AgenticGoKit. The framework is actively developed, so some interfaces may evolve.*