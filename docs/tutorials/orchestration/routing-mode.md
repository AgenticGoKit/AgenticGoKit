# Route Orchestration Mode

## Overview

Route orchestration is the simplest and most common orchestration pattern in AgenticGoKit. Each event is routed to a specific agent based on its target ID or routing metadata. This is the default orchestration mode and forms the foundation for more complex patterns.

Think of route orchestration like a traditional web router - incoming requests are examined and directed to the appropriate handler based on their characteristics.

## Prerequisites

- Understanding of [Message Passing and Event Flow](../core-concepts/message-passing.md)
- Basic knowledge of AgenticGoKit's core concepts
- Familiarity with the [Orchestration Overview](README.md)

## How Route Orchestration Works

### Basic Flow

```
┌─────────┐     ┌──────────┐     ┌─────────┐
│ Client  │────▶│  Router  │────▶│ Agent A │
└─────────┘     └──────────┘     └─────────┘
                     │
                     │ (based on routing metadata)
                     ▼
                ┌─────────┐
                │ Agent B │
                └─────────┘
```

1. **Event Creation**: Client creates an event with routing information
2. **Route Resolution**: Router examines event metadata to determine target agent
3. **Agent Dispatch**: Event is delivered to the specific agent
4. **Result Return**: Agent processes event and returns result

### Routing Mechanism

The route orchestrator uses metadata-based routing:

```go
// Events must include routing metadata
event := core.NewEvent(
    "target-agent-id",  // Target agent identifier
    core.EventData{"message": "Hello, world!"},
    map[string]string{
        "route": "target-agent-id", // Required for routing
        "session_id": "user-123",
    },
)
```

The orchestrator looks for the `"route"` key in the event metadata to determine which agent should handle the event.

## When to Use Route Orchestration

Route orchestration is ideal for:

- **Simple workflows** with clear routing logic
- **Traditional request/response** patterns
- **Microservice-style** agent architectures
- **Deterministic routing** requirements
- **Single-responsibility** agent designs
- **API-like interfaces** where each endpoint has a specific handler

### Use Case Examples

1. **Chat Bot Router**: Route messages to different agents based on intent
2. **API Gateway**: Route requests to specialized service agents
3. **Command Dispatcher**: Route commands to appropriate action handlers
4. **Content Router**: Route content to different processing agents based on type

## Implementation Examples

### Basic Route Orchestration

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create specialized agents
    weatherAgent, err := createWeatherAgent()
    if err != nil {
        log.Fatal(err)
    }
    
    newsAgent, err := createNewsAgent()
    if err != nil {
        log.Fatal(err)
    }
    
    calculatorAgent, err := createCalculatorAgent()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create runner with route orchestrator (default)
    runner := core.NewRunner(100)
    // Route orchestrator is the default, but you can be explicit:
    orchestrator := core.NewRouteOrchestrator(runner.GetCallbackRegistry())
    runner.SetOrchestrator(orchestrator)
    
    // Register agents with their routing keys
    runner.RegisterAgent("weather", weatherAgent)
    runner.RegisterAgent("news", newsAgent)
    runner.RegisterAgent("calculator", calculatorAgent)
    
    // Start the runner
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Example 1: Route to weather agent
    weatherEvent := core.NewEvent(
        "weather",  // Target agent ID
        core.EventData{
            "location": "New York",
            "query": "What's the weather like?",
        },
        map[string]string{
            "route": "weather",
            "session_id": "user-123",
        },
    )
    
    runner.Emit(weatherEvent)
    
    // Example 2: Route to calculator agent
    calcEvent := core.NewEvent(
        "calculator",
        core.EventData{
            "expression": "2 + 2 * 3",
            "operation": "evaluate",
        },
        map[string]string{
            "route": "calculator",
            "session_id": "user-123",
        },
    )
    
    runner.Emit(calcEvent)
    
    // Example 3: Route to news agent
    newsEvent := core.NewEvent(
        "news",
        core.EventData{
            "topic": "artificial intelligence",
            "timeframe": "last 24 hours",
        },
        map[string]string{
            "route": "news",
            "session_id": "user-123",
        },
    )
    
    runner.Emit(newsEvent)
}

// Weather agent implementation
func createWeatherAgent() (core.AgentHandler, error) {
    return core.NewLLMAgent("weather-agent", core.LLMConfig{
        SystemPrompt: `You are a weather assistant. Provide current weather information 
                      for the requested location. Use the location from the event data.`,
        Temperature: 0.3,
        MaxTokens: 200,
    }, core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
}

// News agent implementation
func createNewsAgent() (core.AgentHandler, error) {
    return core.NewLLMAgent("news-agent", core.LLMConfig{
        SystemPrompt: `You are a news assistant. Provide recent news about the requested topic.
                      Focus on factual, recent information.`,
        Temperature: 0.4,
        MaxTokens: 300,
    }, core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
}

// Calculator agent implementation
func createCalculatorAgent() (core.AgentHandler, error) {
    return core.NewLLMAgent("calculator-agent", core.LLMConfig{
        SystemPrompt: `You are a calculator assistant. Evaluate mathematical expressions
                      and provide accurate numerical results.`,
        Temperature: 0.0, // Deterministic for math
        MaxTokens: 100,
    }, core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
}
```

### Dynamic Routing Based on Content

```go
// Create a smart router that examines content to determine routing
func createSmartRouter() core.AgentHandler {
    return &SmartRouterAgent{}
}

type SmartRouterAgent struct{}

func (a *SmartRouterAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Extract message from event
    message, ok := event.GetData()["message"]
    if !ok {
        return core.AgentResult{}, fmt.Errorf("no message in event data")
    }
    
    messageStr := message.(string)
    
    // Determine routing based on content analysis
    var targetAgent string
    
    if containsWeatherKeywords(messageStr) {
        targetAgent = "weather"
    } else if containsNewsKeywords(messageStr) {
        targetAgent = "news"
    } else if containsMathKeywords(messageStr) {
        targetAgent = "calculator"
    } else {
        targetAgent = "general-assistant"
    }
    
    // Create new event with determined routing
    routedEvent := core.NewEvent(
        targetAgent,
        event.GetData(),
        map[string]string{
            "route": targetAgent,
            "session_id": event.GetSessionID(),
            "original_route": event.GetMetadata()["route"],
        },
    )
    
    // Emit the routed event (this would typically be done through the runner)
    // For this example, we'll return the routing decision
    outputState := state.Clone()
    outputState.Set("routed_to", targetAgent)
    outputState.Set("routing_reason", fmt.Sprintf("Content analysis determined %s", targetAgent))
    
    return core.AgentResult{
        OutputState: outputState,
    }, nil
}

func containsWeatherKeywords(message string) bool {
    keywords := []string{"weather", "temperature", "rain", "sunny", "cloudy", "forecast"}
    messageLower := strings.ToLower(message)
    for _, keyword := range keywords {
        if strings.Contains(messageLower, keyword) {
            return true
        }
    }
    return false
}

func containsNewsKeywords(message string) bool {
    keywords := []string{"news", "latest", "breaking", "headlines", "current events"}
    messageLower := strings.ToLower(message)
    for _, keyword := range keywords {
        if strings.Contains(messageLower, keyword) {
            return true
        }
    }
    return false
}

func containsMathKeywords(message string) bool {
    keywords := []string{"calculate", "math", "+", "-", "*", "/", "=", "equation"}
    messageLower := strings.ToLower(message)
    for _, keyword := range keywords {
        if strings.Contains(messageLower, keyword) {
            return true
        }
    }
    return false
}
```

### Conditional Routing with Callbacks

```go
// Set up conditional routing using callbacks
func setupConditionalRouting(runner *core.Runner) {
    runner.RegisterCallback(core.HookAfterAgentRun, "conditional-router",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            // Check if this was a classification agent
            if args.AgentID == "classifier" && args.Error == nil {
                // Get classification result
                if classification, ok := args.AgentResult.OutputState.Get("classification"); ok {
                    // Route to different agents based on classification
                    var nextAgent string
                    switch classification.(string) {
                    case "question":
                        nextAgent = "qa-agent"
                    case "command":
                        nextAgent = "action-agent"
                    case "conversation":
                        nextAgent = "chat-agent"
                    default:
                        nextAgent = "general-agent"
                    }
                    
                    // Create and emit next event
                    nextEvent := core.NewEvent(
                        nextAgent,
                        args.AgentResult.OutputState.GetAll(),
                        map[string]string{
                            "route": nextAgent,
                            "session_id": args.Event.GetSessionID(),
                            "previous_agent": args.AgentID,
                        },
                    )
                    
                    // Emit the routed event
                    runner.Emit(nextEvent)
                }
            }
            return args.State, nil
        },
    )
}
```

## Configuration-Based Routing

You can configure route orchestration through TOML configuration:

```toml
# agentflow.toml
[orchestration]
mode = "route"
timeout = "30s"

# Define routing rules
[orchestration.routing]
default_agent = "general-assistant"

# Route mapping based on metadata
[orchestration.routing.rules]
weather = "weather-agent"
news = "news-agent"
calculate = "calculator-agent"
chat = "conversation-agent"

# Content-based routing patterns
[[orchestration.routing.patterns]]
keywords = ["weather", "temperature", "forecast"]
agent = "weather-agent"

[[orchestration.routing.patterns]]
keywords = ["news", "headlines", "breaking"]
agent = "news-agent"

[[orchestration.routing.patterns]]
keywords = ["calculate", "math", "+", "-", "*", "/"]
agent = "calculator-agent"
```

Load and use the configuration:

```go
// Load configuration
config, err := core.LoadConfig("agentflow.toml")
if err != nil {
    log.Fatal(err)
}

// Create runner from configuration
runner, err := core.NewRunnerFromConfig(config, agents)
if err != nil {
    log.Fatal(err)
}
```

## Advanced Routing Patterns

### 1. Load-Balanced Routing

Route to multiple instances of the same agent type for load balancing:

```go
type LoadBalancedRouter struct {
    agents map[string][]core.AgentHandler
    roundRobin map[string]int
    mu sync.RWMutex
}

func NewLoadBalancedRouter() *LoadBalancedRouter {
    return &LoadBalancedRouter{
        agents: make(map[string][]core.AgentHandler),
        roundRobin: make(map[string]int),
    }
}

func (r *LoadBalancedRouter) RegisterAgent(agentType string, handler core.AgentHandler) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.agents[agentType]; !exists {
        r.agents[agentType] = make([]core.AgentHandler, 0)
        r.roundRobin[agentType] = 0
    }
    
    r.agents[agentType] = append(r.agents[agentType], handler)
}

func (r *LoadBalancedRouter) GetAgent(agentType string) (core.AgentHandler, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    agents, exists := r.agents[agentType]
    if !exists || len(agents) == 0 {
        return nil, fmt.Errorf("no agents registered for type: %s", agentType)
    }
    
    // Round-robin selection
    index := r.roundRobin[agentType]
    agent := agents[index]
    
    // Update round-robin counter
    r.roundRobin[agentType] = (index + 1) % len(agents)
    
    return agent, nil
}
```

### 2. Priority-Based Routing

Route based on priority levels:

```go
type PriorityRouter struct {
    highPriorityAgents map[string]core.AgentHandler
    normalAgents       map[string]core.AgentHandler
}

func (r *PriorityRouter) Route(event core.Event) (core.AgentHandler, error) {
    agentType := event.GetMetadata()["route"]
    priority := event.GetMetadata()["priority"]
    
    if priority == "high" {
        if agent, exists := r.highPriorityAgents[agentType]; exists {
            return agent, nil
        }
    }
    
    // Fall back to normal priority agents
    if agent, exists := r.normalAgents[agentType]; exists {
        return agent, nil
    }
    
    return nil, fmt.Errorf("no agent found for type: %s", agentType)
}
```

### 3. Fallback Routing

Implement fallback agents when primary agents are unavailable:

```go
type FallbackRouter struct {
    primaryAgents  map[string]core.AgentHandler
    fallbackAgents map[string]core.AgentHandler
    defaultAgent   core.AgentHandler
}

func (r *FallbackRouter) Route(event core.Event) (core.AgentHandler, error) {
    agentType := event.GetMetadata()["route"]
    
    // Try primary agent first
    if agent, exists := r.primaryAgents[agentType]; exists {
        return agent, nil
    }
    
    // Try fallback agent
    if agent, exists := r.fallbackAgents[agentType]; exists {
        return agent, nil
    }
    
    // Use default agent as last resort
    if r.defaultAgent != nil {
        return r.defaultAgent, nil
    }
    
    return nil, fmt.Errorf("no agent available for type: %s", agentType)
}
```

## Error Handling in Route Orchestration

### 1. Agent Not Found Errors

```go
runner.RegisterCallback(core.HookAgentError, "route-error-handler",
    func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
        if strings.Contains(args.Error.Error(), "agent not found") {
            // Route to default agent
            defaultEvent := core.NewEvent(
                "default-agent",
                args.Event.GetData(),
                map[string]string{
                    "route": "default-agent",
                    "session_id": args.Event.GetSessionID(),
                    "fallback_reason": "original agent not found",
                },
            )
            
            runner.Emit(defaultEvent)
        }
        return args.State, nil
    },
)
```

### 2. Routing Validation

```go
func validateRouting(event core.Event, availableAgents map[string]core.AgentHandler) error {
    route := event.GetMetadata()["route"]
    if route == "" {
        return fmt.Errorf("no route specified in event metadata")
    }
    
    if _, exists := availableAgents[route]; !exists {
        return fmt.Errorf("agent '%s' not found", route)
    }
    
    return nil
}
```

## Performance Considerations

### 1. Routing Table Optimization

```go
// Use efficient data structures for large routing tables
type OptimizedRouter struct {
    agentMap sync.Map // Concurrent map for better performance
    metrics  *RoutingMetrics
}

func (r *OptimizedRouter) Route(event core.Event) (core.AgentHandler, error) {
    start := time.Now()
    defer func() {
        r.metrics.RecordRoutingTime(time.Since(start))
    }()
    
    route := event.GetMetadata()["route"]
    if agent, ok := r.agentMap.Load(route); ok {
        return agent.(core.AgentHandler), nil
    }
    
    return nil, fmt.Errorf("agent not found: %s", route)
}
```

### 2. Caching Routing Decisions

```go
type CachedRouter struct {
    router core.Orchestrator
    cache  map[string]string // content hash -> agent ID
    mu     sync.RWMutex
}

func (r *CachedRouter) Route(event core.Event) (string, error) {
    // Create content hash
    contentHash := r.hashEventContent(event)
    
    // Check cache first
    r.mu.RLock()
    if agentID, exists := r.cache[contentHash]; exists {
        r.mu.RUnlock()
        return agentID, nil
    }
    r.mu.RUnlock()
    
    // Perform routing logic
    agentID := r.determineAgent(event)
    
    // Cache the result
    r.mu.Lock()
    r.cache[contentHash] = agentID
    r.mu.Unlock()
    
    return agentID, nil
}
```

## Monitoring and Debugging

### 1. Routing Metrics

```go
type RoutingMetrics struct {
    routingCounts map[string]int64
    routingTimes  map[string][]time.Duration
    mu            sync.RWMutex
}

func (m *RoutingMetrics) RecordRouting(agentID string, duration time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.routingCounts[agentID]++
    m.routingTimes[agentID] = append(m.routingTimes[agentID], duration)
}

func (m *RoutingMetrics) GetStats() map[string]interface{} {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    stats := make(map[string]interface{})
    for agentID, count := range m.routingCounts {
        times := m.routingTimes[agentID]
        avgTime := calculateAverage(times)
        
        stats[agentID] = map[string]interface{}{
            "count": count,
            "avg_time_ms": avgTime.Milliseconds(),
        }
    }
    
    return stats
}
```

### 2. Routing Visualization

```go
// Generate routing diagram
func GenerateRoutingDiagram(agents map[string]core.AgentHandler) string {
    var diagram strings.Builder
    
    diagram.WriteString("graph TD\\n")
    diagram.WriteString("    Client[Client] --> Router[Route Orchestrator]\\n")
    
    for agentID := range agents {
        diagram.WriteString(fmt.Sprintf("    Router --> %s[%s]\\n", agentID, agentID))
    }
    
    return diagram.String()
}
```

## Best Practices

### 1. Routing Design Principles

- **Single Responsibility**: Each agent should have a clear, focused purpose
- **Explicit Routing**: Always specify routing information clearly
- **Fallback Strategy**: Implement fallback agents for robustness
- **Validation**: Validate routing information before processing
- **Monitoring**: Track routing patterns and performance

### 2. Common Patterns

```go
// Pattern 1: Intent-based routing
func routeByIntent(message string) string {
    intent := classifyIntent(message)
    switch intent {
    case "weather":
        return "weather-agent"
    case "news":
        return "news-agent"
    default:
        return "general-agent"
    }
}

// Pattern 2: User-based routing
func routeByUser(userID string) string {
    userType := getUserType(userID)
    switch userType {
    case "premium":
        return "premium-agent"
    case "enterprise":
        return "enterprise-agent"
    default:
        return "standard-agent"
    }
}

// Pattern 3: Load-based routing
func routeByLoad(agentType string) string {
    agents := getAgentsOfType(agentType)
    return selectLeastLoadedAgent(agents)
}
```

### 3. Testing Route Orchestration

```go
func TestRouteOrchestration(t *testing.T) {
    // Create test agents
    testAgent := &MockAgent{}
    
    // Create orchestrator
    orchestrator := core.NewRouteOrchestrator(core.NewCallbackRegistry())
    orchestrator.RegisterAgent("test-agent", testAgent)
    
    // Create test event
    event := core.NewEvent(
        "test-agent",
        core.EventData{"message": "test"},
        map[string]string{"route": "test-agent"},
    )
    
    // Test routing
    result, err := orchestrator.Dispatch(context.Background(), event)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.True(t, testAgent.WasCalled())
}
```

## Common Pitfalls and Solutions

### 1. Missing Route Information

**Problem**: Events without routing metadata cause errors.

**Solution**: Always validate and provide default routing:

```go
func ensureRouting(event core.Event) core.Event {
    if event.GetMetadata()["route"] == "" {
        // Add default routing
        metadata := event.GetMetadata()
        metadata["route"] = "default-agent"
        
        return core.NewEvent(
            event.GetTargetAgentID(),
            event.GetData(),
            metadata,
        )
    }
    return event
}
```

### 2. Agent Registration Issues

**Problem**: Agents not properly registered with orchestrator.

**Solution**: Implement registration validation:

```go
func validateAgentRegistration(orchestrator core.Orchestrator, requiredAgents []string) error {
    for _, agentID := range requiredAgents {
        if !orchestrator.HasAgent(agentID) {
            return fmt.Errorf("required agent '%s' not registered", agentID)
        }
    }
    return nil
}
```

### 3. Circular Routing

**Problem**: Agents routing events back to themselves or creating loops.

**Solution**: Track routing history:

```go
func preventCircularRouting(event core.Event, maxHops int) error {
    hops := event.GetMetadata()["routing_hops"]
    if hops == "" {
        hops = "0"
    }
    
    hopCount, _ := strconv.Atoi(hops)
    if hopCount >= maxHops {
        return fmt.Errorf("maximum routing hops exceeded: %d", hopCount)
    }
    
    return nil
}
```

## Conclusion

Route orchestration provides a solid foundation for building agent systems with clear separation of concerns. It's the perfect starting point for most applications and can be extended with more complex patterns as your needs grow.

Key takeaways:
- Route orchestration is simple, predictable, and efficient
- Use metadata-based routing for flexibility
- Implement fallback strategies for robustness
- Monitor routing patterns for optimization opportunities
- Start simple and evolve to more complex patterns as needed

## Next Steps

- [Collaborative Orchestration](collaborative-mode.md) - Learn parallel agent processing
- [Sequential Orchestration](sequential-mode.md) - Build processing pipelines
- [Mixed Orchestration](mixed-mode.md) - Combine multiple patterns
- [State Management](../core-concepts/state-management.md) - Understand data flow between agents

## Further Reading

- [API Reference: RouteOrchestrator](../../api/core.md#route-orchestrator)
- [Examples: Route Orchestration](../../examples/01-simple-agent/)
- [Configuration Guide: Routing](../../configuration/orchestration.md#routing)