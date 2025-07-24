# State & Event API Reference

**Complete reference for data flow and communication between agents**

This document provides comprehensive API reference for AgenticGoKit's state management and event system, which enables data flow and communication between agents in multi-agent systems.

## 🏗️ Core Interfaces

### Event Interface

The Event interface represents data that flows between agents in the system.

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

#### Methods

##### GetID
```go
GetID() string
```
Returns the unique identifier for this event.

##### GetData
```go
GetData() EventData
```
Returns the event payload as key-value pairs.

**Example**:
```go
data := event.GetData()
message, ok := data[\"message\"]
if ok {
    fmt.Printf(\"User message: %s\\n\", message)
}
```

##### GetMetadata
```go
GetMetadata() map[string]string
```
Returns all metadata associated with the event.

##### GetMetadataValue
```go
GetMetadataValue(key string) (string, bool)
```
Returns a specific metadata value by key.

**Example**:
```go
sessionID, exists := event.GetMetadataValue(\"session_id\")
if exists {
    fmt.Printf(\"Session: %s\\n\", sessionID)
}
```

##### GetSourceAgentID / GetTargetAgentID
```go
GetSourceAgentID() string
GetTargetAgentID() string
```
Returns the source and target agent identifiers for event routing.

### State Interface

Thread-safe state management interface for maintaining data across agent interactions.

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

#### Data Methods

##### Get/Set
```go
Set(key string, value any)
Get(key string) (any, bool)
```
Store and retrieve typed values in the state.

**Example**:
```go
// Store different types of data
state.Set(\"user_preferences\", []string{\"technical\", \"detailed\"})
state.Set(\"conversation_turn\", 3)
state.Set(\"user_context\", map[string]interface{}{
    \"name\": \"John\",
    \"role\": \"developer\",
})

// Retrieve data with type assertion
prefs, exists := state.Get(\"user_preferences\")
if exists {
    preferences := prefs.([]string)
    fmt.Printf(\"User preferences: %v\\n\", preferences)
}

turn, exists := state.Get(\"conversation_turn\")
if exists {
    turnNumber := turn.(int)
    fmt.Printf(\"Turn: %d\\n\", turnNumber)
}
```

##### Keys
```go
Keys() []string
```
Returns all data keys currently stored in the state.

**Example**:
```go
keys := state.Keys()
fmt.Printf(\"State contains: %v\\n\", keys)
```

#### Metadata Methods

##### GetMeta/SetMeta
```go
SetMeta(key, value string)
GetMeta(key string) (string, bool)
```
Store and retrieve string metadata about the state.

**Example**:
```go
// Store metadata
state.SetMeta(\"processed_by\", \"agent1\")
state.SetMeta(\"timestamp\", time.Now().Format(time.RFC3339))
state.SetMeta(\"session_id\", \"abc123\")

// Retrieve metadata
agent, exists := state.GetMeta(\"processed_by\")
if exists {
    fmt.Printf(\"Processed by: %s\\n\", agent)
}
```

##### MetaKeys
```go
MetaKeys() []string
```
Returns all metadata keys currently stored in the state.

#### Utility Methods

##### Clone
```go
Clone() State
```
Creates a deep copy of the state for safe concurrent access.

**Example**:
```go
// Create a copy for modification
stateCopy := state.Clone()
stateCopy.Set(\"modified\", true)

// Original state is unchanged
original, _ := state.Get(\"modified\")
fmt.Printf(\"Original modified: %v\\n\", original) // nil
```

## 📊 Type Definitions

### EventData

Type alias for event payload data.

```go
type EventData map[string]interface{}
```

**Usage**:
```go
eventData := core.EventData{
    \"message\":    \"Hello, world!\",
    \"user_id\":    \"12345\",
    \"timestamp\":  time.Now(),
    \"context\":    map[string]string{\"session\": \"abc123\"},
}

event := core.NewEvent(\"user-message\", eventData, nil)
```

### AgentResult

Result returned by agent execution, containing response data and updated state.

```go
type AgentResult struct {
    Data map[string]interface{} `json:\"data\"`
    State State `json:\"state,omitempty\"`
    Metadata map[string]interface{} `json:\"metadata,omitempty\"`
    Errors []error `json:\"errors,omitempty\"`
    Success bool `json:\"success\"`
}
```

#### Fields

##### Data
The primary response data that will be returned to the caller:
```go
result := core.AgentResult{
    Data: map[string]interface{}{
        \"answer\":     \"Paris is the capital of France\",
        \"confidence\": 0.95,
        \"sources\":    []string{\"wikipedia\", \"britannica\"},
    },
}
```

##### State
Updated state to persist for the session:
```go
// Update conversation state
state.Set(\"last_query\", query)
state.Set(\"query_count\", state.GetInt(\"query_count\")+1)

result := core.AgentResult{
    Data:  responseData,
    State: state,
}
```

##### Metadata
Additional execution information:
```go
result := core.AgentResult{
    Data: responseData,
    Metadata: map[string]interface{}{
        \"execution_time\": time.Since(start),
        \"tokens_used\":    tokenCount,
        \"model\":          \"gpt-4o\",
        \"tools_called\":   []string{\"search\", \"calculator\"},
    },
}
```

##### Errors
Non-fatal errors that occurred during processing:
```go
result := core.AgentResult{
    Data: partialData,
    Errors: []error{
        fmt.Errorf(\"tool 'advanced_search' failed: %w\", searchErr),
        fmt.Errorf(\"cache miss for query: %s\", query),
    },
    Success: true, // Still successful despite errors
}
```

## 🏭 Factory Functions

### Event Creation

#### NewEvent
```go
func NewEvent(eventType string, data EventData, metadata map[string]string) Event
```
Creates a new event with the specified data and metadata.

**Parameters**:
- `eventType` - Type identifier for the event
- `data` - Event payload data
- `metadata` - Optional metadata map

**Example**:
```go
eventData := core.EventData{\"message\": \"Hello, world!\"}
metadata := map[string]string{\"session_id\": \"123\", \"user_id\": \"456\"}
event := core.NewEvent(\"user-message\", eventData, metadata)
```

### State Creation

#### NewState
```go
func NewState() State
```
Creates a new empty state instance.

**Example**:
```go
state := core.NewState()
state.Set(\"conversation_history\", []string{})
state.SetMeta(\"session_id\", \"abc123\")
```

## 🔄 State Management Patterns

### Conversation State Pattern

Managing conversation history and context:

```go
type ConversationAgent struct {
    llm core.ModelProvider
}

func (a *ConversationAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get current conversation history
    history, exists := state.Get(\"conversation_history\")
    var messages []core.Message
    if exists {
        messages = history.([]core.Message)
    }
    
    // Extract user message from event
    data := event.GetData()
    userMessage := data[\"message\"].(string)
    
    // Add user message to history
    messages = append(messages, core.Message{
        Role:    \"user\",
        Content: userMessage,
    })
    
    // Generate response
    response, err := a.llm.GenerateWithHistory(ctx, messages)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Add assistant response to history
    messages = append(messages, core.Message{
        Role:    \"assistant\",
        Content: response,
    })
    
    // Update state with new history
    state.Set(\"conversation_history\", messages)
    state.Set(\"last_response\", response)
    state.SetMeta(\"last_updated\", time.Now().Format(time.RFC3339))
    
    return core.AgentResult{
        Data: map[string]interface{}{
            \"response\": response,
        },
        State: state,
        Success: true,
    }, nil
}
```

### Session State Pattern

Managing user session data across multiple interactions:

```go
type SessionAgent struct {
    sessionStore SessionStore
}

func (a *SessionAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get session ID from event metadata
    sessionID, exists := event.GetMetadataValue(\"session_id\")
    if !exists {
        return core.AgentResult{}, fmt.Errorf(\"missing session_id in event metadata\")
    }
    
    // Load session data if not in state
    if _, exists := state.Get(\"session_data\"); !exists {
        sessionData, err := a.sessionStore.Load(sessionID)
        if err != nil {
            // Create new session
            sessionData = map[string]interface{}{
                \"created_at\": time.Now(),
                \"user_id\":    event.GetMetadataValue(\"user_id\"),
            }
        }
        state.Set(\"session_data\", sessionData)
    }
    
    // Process the event
    data := event.GetData()
    query := data[\"query\"].(string)
    
    // Update session with query count
    sessionData := state.Get(\"session_data\").(map[string]interface{})
    queryCount, _ := sessionData[\"query_count\"].(int)
    sessionData[\"query_count\"] = queryCount + 1
    sessionData[\"last_query\"] = query
    sessionData[\"last_activity\"] = time.Now()
    
    state.Set(\"session_data\", sessionData)
    
    // Save session data
    if err := a.sessionStore.Save(sessionID, sessionData); err != nil {
        // Log error but don't fail the request
        log.Printf(\"Failed to save session data: %v\", err)
    }
    
    return core.AgentResult{
        Data: map[string]interface{}{
            \"processed\": true,
            \"query_count\": queryCount + 1,
        },
        State: state,
        Success: true,
    }, nil
}
```

### Multi-Agent State Sharing Pattern

Sharing state between multiple agents in a workflow:

```go
type AnalysisAgent struct {
    name string
}

func (a *AnalysisAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get shared analysis context
    analysisContext, exists := state.Get(\"analysis_context\")
    if !exists {
        analysisContext = map[string]interface{}{
            \"started_at\": time.Now(),
            \"agents_completed\": []string{},
            \"findings\": map[string]interface{}{},
        }
    }
    
    context := analysisContext.(map[string]interface{})
    
    // Perform analysis
    data := event.GetData()
    text := data[\"text\"].(string)
    
    analysis := a.performAnalysis(text)
    
    // Add findings to shared context
    findings := context[\"findings\"].(map[string]interface{})
    findings[a.name] = analysis
    
    // Mark this agent as completed
    completed := context[\"agents_completed\"].([]string)
    completed = append(completed, a.name)
    context[\"agents_completed\"] = completed
    
    // Update shared state
    state.Set(\"analysis_context\", context)
    
    return core.AgentResult{
        Data: map[string]interface{}{
            \"analysis\": analysis,
            \"agent\": a.name,
        },
        State: state,
        Success: true,
    }, nil
}

func (a *AnalysisAgent) performAnalysis(text string) map[string]interface{} {
    // Implement specific analysis logic
    return map[string]interface{}{
        \"word_count\": len(strings.Fields(text)),
        \"sentiment\": \"positive\",
        \"topics\": []string{\"technology\", \"programming\"},
    }
}
```

## 🔍 Event Routing Patterns

### Route-Based Event Handling

Using event metadata for routing decisions:

```go
type RouterAgent struct {
    handlers map[string]core.AgentHandler
}

func (r *RouterAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get route from event metadata
    route, exists := event.GetMetadataValue(\"route\")
    if !exists {
        // Default routing based on event data
        data := event.GetData()
        if _, hasQuery := data[\"query\"]; hasQuery {
            route = \"query_handler\"
        } else if _, hasCommand := data[\"command\"]; hasCommand {
            route = \"command_handler\"
        } else {
            route = \"default_handler\"
        }
    }
    
    // Find appropriate handler
    handler, exists := r.handlers[route]
    if !exists {
        return core.AgentResult{}, fmt.Errorf(\"no handler found for route: %s\", route)
    }
    
    // Delegate to specific handler
    return handler.Run(ctx, event, state)
}
```

### Event Transformation Pattern

Transforming events between different formats:

```go
type TransformAgent struct {
    transformer EventTransformer
}

func (t *TransformAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Transform incoming event
    transformedData, err := t.transformer.Transform(event.GetData())
    if err != nil {
        return core.AgentResult{}, fmt.Errorf(\"event transformation failed: %w\", err)
    }
    
    // Create new event with transformed data
    newEvent := core.NewEvent(\"transformed\", transformedData, event.GetMetadata())
    
    // Store transformation history in state
    history, exists := state.Get(\"transformation_history\")
    if !exists {
        history = []string{}
    }
    
    historyList := history.([]string)
    historyList = append(historyList, fmt.Sprintf(\"Transformed at %s\", time.Now().Format(time.RFC3339)))
    state.Set(\"transformation_history\", historyList)
    
    return core.AgentResult{
        Data: map[string]interface{}{
            \"transformed_event\": newEvent,
            \"original_event\": event,
        },
        State: state,
        Success: true,
    }, nil
}
```

## 🧪 Testing State and Events

### State Testing Utilities

```go
func TestStateOperations(t *testing.T) {
    state := core.NewState()
    
    // Test basic operations
    state.Set(\"key1\", \"value1\")
    state.Set(\"key2\", 42)
    
    value1, exists := state.Get(\"key1\")
    assert.True(t, exists)
    assert.Equal(t, \"value1\", value1.(string))
    
    value2, exists := state.Get(\"key2\")
    assert.True(t, exists)
    assert.Equal(t, 42, value2.(int))
    
    // Test metadata
    state.SetMeta(\"meta1\", \"metavalue1\")
    metaValue, exists := state.GetMeta(\"meta1\")
    assert.True(t, exists)
    assert.Equal(t, \"metavalue1\", metaValue)
    
    // Test cloning
    clonedState := state.Clone()
    clonedState.Set(\"key3\", \"value3\")
    
    _, exists = state.Get(\"key3\")
    assert.False(t, exists) // Original state unchanged
    
    _, exists = clonedState.Get(\"key3\")
    assert.True(t, exists) // Cloned state has new value
}
```

### Event Testing Utilities

```go
func TestEventCreation(t *testing.T) {
    eventData := core.EventData{
        \"message\": \"test message\",
        \"user_id\": \"123\",
    }
    
    metadata := map[string]string{
        \"session_id\": \"session123\",
        \"route\": \"test_route\",
    }
    
    event := core.NewEvent(\"test_event\", eventData, metadata)
    
    // Test event data
    data := event.GetData()
    assert.Equal(t, \"test message\", data[\"message\"])
    assert.Equal(t, \"123\", data[\"user_id\"])
    
    // Test metadata
    sessionID, exists := event.GetMetadataValue(\"session_id\")
    assert.True(t, exists)
    assert.Equal(t, \"session123\", sessionID)
    
    route, exists := event.GetMetadataValue(\"route\")
    assert.True(t, exists)
    assert.Equal(t, \"test_route\", route)
}
```

### Agent Result Testing

```go
func TestAgentResult(t *testing.T) {
    state := core.NewState()
    state.Set(\"processed\", true)
    
    result := core.AgentResult{
        Data: map[string]interface{}{
            \"response\": \"test response\",
            \"confidence\": 0.95,
        },
        State: state,
        Metadata: map[string]interface{}{
            \"execution_time\": \"100ms\",
        },
        Success: true,
    }
    
    // Test result data
    assert.Equal(t, \"test response\", result.Data[\"response\"])
    assert.Equal(t, 0.95, result.Data[\"confidence\"])
    
    // Test state
    processed, exists := result.State.Get(\"processed\")
    assert.True(t, exists)
    assert.True(t, processed.(bool))
    
    // Test metadata
    assert.Equal(t, \"100ms\", result.Metadata[\"execution_time\"])
    
    // Test success
    assert.True(t, result.Success)
}
```

## 🔧 Best Practices

### State Management

1. **Use appropriate data types**: Store data in the most appropriate Go type
2. **Namespace keys**: Use prefixes to avoid key collisions (e.g., \"user.preferences\", \"session.data\")
3. **Clean up state**: Remove unnecessary data to prevent memory leaks
4. **Use metadata for system info**: Store system-level information in metadata

### Event Design

1. **Consistent event types**: Use consistent naming for event types
2. **Include context**: Always include necessary context in event data
3. **Use metadata for routing**: Use metadata for system-level routing information
4. **Validate event data**: Always validate event data before processing

### Error Handling

1. **Graceful degradation**: Handle missing state gracefully
2. **Preserve partial results**: Return partial results when possible
3. **Log state changes**: Log important state changes for debugging
4. **Use non-fatal errors**: Use the Errors field for non-fatal issues

This comprehensive reference covers all aspects of state and event management in AgenticGoKit, providing the foundation for building robust multi-agent systems.