# State Management and Data Flow in AgenticGoKit

## Overview

State management is the backbone of data flow in AgenticGoKit. It determines how information is stored, passed between agents, and persisted across interactions. This tutorial explores how State objects work, how data flows through agent systems, and best practices for managing complex data transformations.

Understanding state management is crucial because it's how agents share information, maintain context, and build upon each other's work.

## Prerequisites

- Understanding of [Message Passing and Event Flow](message-passing.md)
- Basic knowledge of Go interfaces and concurrency
- Familiarity with AgenticGoKit's core concepts

## Core Concepts

### State: The Agent's Working Memory

State represents the current context and data an agent is working with. It's a thread-safe container that holds both data and metadata:

```go
type State interface {
    // Data operations
    Get(key string) (any, bool)     // Retrieve a value by key
    Set(key string, value any)      // Store a value by key
    Keys() []string                 // Get all data keys
    
    // Metadata operations
    GetMeta(key string) (string, bool)  // Retrieve metadata by key
    SetMeta(key string, value string)   // Store metadata by key
    MetaKeys() []string                 // Get all metadata keys
    
    // State operations
    Clone() State                   // Create a deep copy
    Merge(source State)             // Merge another state into this one
}
```

### Data vs Metadata

**Data** contains the actual information agents work with:
- User messages
- Processing results
- Intermediate calculations
- Business logic data

**Metadata** contains information about the data:
- Processing instructions
- Routing information
- Quality scores
- Timestamps and tracking info

```go
// Example state with data and metadata
state := core.NewState()

// Data - what the agent works with
state.Set("user_message", "What's the weather in Paris?")
state.Set("location", "Paris")
state.Set("temperature", 22.5)

// Metadata - information about the data
state.SetMeta("confidence", "0.95")
state.SetMeta("source", "weather-api")
state.SetMeta("timestamp", time.Now().Format(time.RFC3339))
```

## State Lifecycle

### 1. State Creation

States are typically created at the beginning of agent processing:

```go
// Create empty state
state := core.NewState()

// Create state with initial data
initialData := map[string]any{
    "user_id": "user-123",
    "session_id": "session-456",
    "query": "Tell me about AI",
}
state := core.NewStateWithData(initialData)

// Alternative creation method
state := core.NewSimpleState(initialData)
```

### 2. State Transformation

Agents receive state as input and produce modified state as output:

```go
func (a *MyAgent) Run(ctx context.Context, event Event, inputState State) (AgentResult, error) {
    // Read from input state
    query, ok := inputState.Get("query")
    if !ok {
        return AgentResult{}, errors.New("no query in state")
    }
    
    // Process the query
    response := a.processQuery(query.(string))
    
    // Create output state with new data
    outputState := inputState.Clone() // Start with input state
    outputState.Set("response", response)
    outputState.Set("processed_at", time.Now())
    outputState.SetMeta("agent", a.name)
    
    return AgentResult{
        OutputState: outputState,
    }, nil
}
```

### 3. State Propagation

State flows between agents through the orchestration system:

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│ Initial │───▶│ Agent A │───▶│ Agent B │───▶│ Final   │
│ State   │    │ State   │    │ State   │    │ State   │
└─────────┘    └─────────┘    └─────────┘    └─────────┘
```

## Data Flow Patterns

### 1. Linear Data Flow

Data flows sequentially through agents, with each agent adding or modifying information:

```go
// Agent 1: Data Collection
func (a *CollectorAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Collect raw data
    rawData := a.collectData()
    
    outputState := state.Clone()
    outputState.Set("raw_data", rawData)
    outputState.SetMeta("stage", "collection")
    
    return AgentResult{OutputState: outputState}, nil
}

// Agent 2: Data Processing
func (a *ProcessorAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Get raw data from previous agent
    rawData, ok := state.Get("raw_data")
    if !ok {
        return AgentResult{}, errors.New("no raw data to process")
    }
    
    // Process the data
    processedData := a.processData(rawData)
    
    outputState := state.Clone()
    outputState.Set("processed_data", processedData)
    outputState.SetMeta("stage", "processing")
    
    return AgentResult{OutputState: outputState}, nil
}

// Agent 3: Data Formatting
func (a *FormatterAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Get processed data
    processedData, ok := state.Get("processed_data")
    if !ok {
        return AgentResult{}, errors.New("no processed data to format")
    }
    
    // Format the data
    formattedData := a.formatData(processedData)
    
    outputState := state.Clone()
    outputState.Set("final_result", formattedData)
    outputState.SetMeta("stage", "formatting")
    
    return AgentResult{OutputState: outputState}, nil
}
```

### 2. Branching Data Flow

Data flows to multiple agents, each processing different aspects:

```go
// Main agent creates branches
func (a *MainAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    query, _ := state.Get("query")
    
    outputState := state.Clone()
    
    // Create different processing branches
    outputState.Set("text_analysis_query", query)
    outputState.Set("sentiment_analysis_query", query)
    outputState.Set("entity_extraction_query", query)
    
    // Set routing for different agents
    outputState.SetMeta("next_agents", "text_analyzer,sentiment_analyzer,entity_extractor")
    
    return AgentResult{OutputState: outputState}, nil
}
```

### 3. Merging Data Flow

Multiple agents contribute to a shared state that gets combined:

```go
// Collaborative agents contribute to shared state
func (a *ResearchAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    query, _ := state.Get("research_query")
    
    // Perform research
    findings := a.research(query.(string))
    
    outputState := state.Clone()
    
    // Add findings to shared research data
    existingFindings, _ := outputState.Get("research_findings")
    if existingFindings == nil {
        existingFindings = make([]ResearchFinding, 0)
    }
    
    allFindings := append(existingFindings.([]ResearchFinding), findings...)
    outputState.Set("research_findings", allFindings)
    outputState.SetMeta("contributor", a.name)
    
    return AgentResult{OutputState: outputState}, nil
}
```

## Advanced State Management Patterns

### 1. Namespaced State Keys

Use namespaces to organize state data and avoid conflicts:

```go
// Namespaced keys prevent conflicts
state.Set("user.profile.name", "Alice")
state.Set("user.profile.preferences", []string{"history", "science"})
state.Set("user.session.start_time", time.Now())

state.Set("system.version", "1.0.0")
state.Set("system.environment", "production")

state.Set("processing.stage", "analysis")
state.Set("processing.confidence", 0.95)
```

### 2. Structured State Data

Store complex objects as structured data:

```go
// Define structured data types
type UserProfile struct {
    Name        string    `json:"name"`
    Email       string    `json:"email"`
    Preferences []string  `json:"preferences"`
    LastSeen    time.Time `json:"last_seen"`
}

type ProcessingContext struct {
    Stage       string            `json:"stage"`
    Confidence  float64           `json:"confidence"`
    Metadata    map[string]string `json:"metadata"`
}

// Store structured data in state
profile := UserProfile{
    Name:        "Alice",
    Email:       "alice@example.com",
    Preferences: []string{"AI", "Technology"},
    LastSeen:    time.Now(),
}

context := ProcessingContext{
    Stage:      "analysis",
    Confidence: 0.95,
    Metadata:   map[string]string{"source": "llm"},
}

state.Set("user_profile", profile)
state.Set("processing_context", context)
```

### 3. State Validation

Implement validation to ensure state integrity:

```go
// State validator interface
type StateValidator interface {
    Validate(state State) error
}

// Example validator
type UserQueryValidator struct{}

func (v *UserQueryValidator) Validate(state State) error {
    // Check required fields
    if _, ok := state.Get("user_query"); !ok {
        return errors.New("user_query is required")
    }
    
    if _, ok := state.Get("user_id"); !ok {
        return errors.New("user_id is required")
    }
    
    // Validate data types
    if query, ok := state.Get("user_query"); ok {
        if _, isString := query.(string); !isString {
            return errors.New("user_query must be a string")
        }
    }
    
    return nil
}

// Use validator in agent
func (a *MyAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    validator := &UserQueryValidator{}
    if err := validator.Validate(state); err != nil {
        return AgentResult{}, fmt.Errorf("state validation failed: %w", err)
    }
    
    // Continue with processing...
}
```

### 4. State Transformation Helpers

Create helper functions for common state transformations:

```go
// State transformation helpers
func AddUserContext(state State, userID, sessionID string) State {
    newState := state.Clone()
    newState.Set("user_id", userID)
    newState.Set("session_id", sessionID)
    newState.SetMeta("context_added", time.Now().Format(time.RFC3339))
    return newState
}

func AddProcessingMetadata(state State, agentName string, confidence float64) State {
    newState := state.Clone()
    newState.SetMeta("processed_by", agentName)
    newState.SetMeta("confidence", fmt.Sprintf("%.2f", confidence))
    newState.SetMeta("processed_at", time.Now().Format(time.RFC3339))
    return newState
}

func ExtractUserQuery(state State) (string, error) {
    query, ok := state.Get("user_query")
    if !ok {
        return "", errors.New("no user query in state")
    }
    
    queryStr, ok := query.(string)
    if !ok {
        return "", errors.New("user query is not a string")
    }
    
    return queryStr, nil
}

// Usage in agents
func (a *MyAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Add context
    state = AddUserContext(state, "user-123", "session-456")
    
    // Extract query
    query, err := ExtractUserQuery(state)
    if err != nil {
        return AgentResult{}, err
    }
    
    // Process query
    response := a.processQuery(query)
    
    // Add processing metadata
    outputState := AddProcessingMetadata(state, a.name, 0.95)
    outputState.Set("response", response)
    
    return AgentResult{OutputState: outputState}, nil
}
```

## State Persistence and Serialization

### 1. JSON Serialization

State objects can be serialized to JSON for storage or transmission:

```go
// Serialize state to JSON
func SerializeState(state State) ([]byte, error) {
    return json.Marshal(state)
}

// Deserialize state from JSON
func DeserializeState(data []byte) (State, error) {
    var state core.SimpleState
    err := json.Unmarshal(data, &state)
    if err != nil {
        return nil, err
    }
    return &state, nil
}

// Example usage
func saveStateToFile(state State, filename string) error {
    data, err := SerializeState(state)
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, data, 0644)
}

func loadStateFromFile(filename string) (State, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    return DeserializeState(data)
}
```

### 2. State Snapshots

Capture state at specific points for debugging or rollback:

```go
// State snapshot manager
type StateSnapshot struct {
    Timestamp time.Time `json:"timestamp"`
    AgentID   string    `json:"agent_id"`
    State     State     `json:"state"`
}

type StateSnapshotManager struct {
    snapshots []StateSnapshot
    mu        sync.RWMutex
}

func (sm *StateSnapshotManager) TakeSnapshot(agentID string, state State) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    snapshot := StateSnapshot{
        Timestamp: time.Now(),
        AgentID:   agentID,
        State:     state.Clone(),
    }
    
    sm.snapshots = append(sm.snapshots, snapshot)
}

func (sm *StateSnapshotManager) GetSnapshots() []StateSnapshot {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    
    // Return a copy to avoid race conditions
    snapshots := make([]StateSnapshot, len(sm.snapshots))
    copy(snapshots, sm.snapshots)
    return snapshots
}

// Usage in agent processing
var snapshotManager = &StateSnapshotManager{}

func (a *MyAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Take snapshot before processing
    snapshotManager.TakeSnapshot(a.name+"-input", state)
    
    // Process...
    outputState := state.Clone()
    outputState.Set("response", "processed")
    
    // Take snapshot after processing
    snapshotManager.TakeSnapshot(a.name+"-output", outputState)
    
    return AgentResult{OutputState: outputState}, nil
}
```

## State in Different Orchestration Patterns

### 1. Route Orchestration State Flow

```go
// Simple state passing between specific agents
event := core.NewEvent(
    "agent-a",
    core.EventData{"input": "data"},
    map[string]string{"route": "agent-a"},
)

// Agent A processes and routes to Agent B
func (a *AgentA) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Process input
    input, _ := state.Get("input")
    result := a.process(input)
    
    outputState := state.Clone()
    outputState.Set("intermediate_result", result)
    outputState.SetMeta("route", "agent-b") // Route to next agent
    
    return AgentResult{OutputState: outputState}, nil
}
```

### 2. Collaborative Orchestration State Merging

```go
// Multiple agents contribute to shared state
type CollaborativeStateManager struct {
    contributions map[string]State
    mu           sync.RWMutex
}

func (csm *CollaborativeStateManager) AddContribution(agentID string, state State) {
    csm.mu.Lock()
    defer csm.mu.Unlock()
    
    if csm.contributions == nil {
        csm.contributions = make(map[string]State)
    }
    
    csm.contributions[agentID] = state.Clone()
}

func (csm *CollaborativeStateManager) MergeContributions() State {
    csm.mu.RLock()
    defer csm.mu.RUnlock()
    
    mergedState := core.NewState()
    
    for agentID, contribution := range csm.contributions {
        // Merge each contribution
        mergedState.Merge(contribution)
        
        // Add contributor metadata
        mergedState.SetMeta("contributor_"+agentID, "true")
    }
    
    return mergedState
}
```

### 3. Sequential Orchestration State Pipeline

```go
// State flows through pipeline stages
type PipelineStage struct {
    Name      string
    Transform func(State) (State, error)
}

func ProcessPipeline(initialState State, stages []PipelineStage) (State, error) {
    currentState := initialState.Clone()
    
    for i, stage := range stages {
        // Add stage metadata
        currentState.SetMeta("current_stage", stage.Name)
        currentState.SetMeta("stage_number", fmt.Sprintf("%d", i+1))
        
        // Transform state
        newState, err := stage.Transform(currentState)
        if err != nil {
            return currentState, fmt.Errorf("stage %s failed: %w", stage.Name, err)
        }
        
        currentState = newState
        
        // Add completion metadata
        currentState.SetMeta("completed_stage_"+stage.Name, time.Now().Format(time.RFC3339))
    }
    
    return currentState, nil
}
```

## Performance Considerations

### 1. State Cloning Optimization

```go
// Efficient state cloning for large states
type OptimizedState struct {
    *core.SimpleState
    copyOnWrite bool
}

func (os *OptimizedState) Clone() State {
    if !os.copyOnWrite {
        // Shallow copy for read-only scenarios
        return &OptimizedState{
            SimpleState: os.SimpleState,
            copyOnWrite: true,
        }
    }
    
    // Deep copy when modifications are needed
    return &OptimizedState{
        SimpleState: os.SimpleState.Clone().(*core.SimpleState),
        copyOnWrite: false,
    }
}
```

### 2. State Size Management

```go
// Monitor and limit state size
func CheckStateSize(state State) error {
    data, err := json.Marshal(state)
    if err != nil {
        return err
    }
    
    const maxStateSize = 1024 * 1024 // 1MB
    if len(data) > maxStateSize {
        return fmt.Errorf("state size %d exceeds maximum %d bytes", len(data), maxStateSize)
    }
    
    return nil
}

// Compress large state data
func CompressStateData(state State, key string) error {
    value, ok := state.Get(key)
    if !ok {
        return nil
    }
    
    // Serialize and compress large data
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    if len(data) > 10240 { // 10KB threshold
        compressed := compress(data) // Your compression function
        state.Set(key+"_compressed", compressed)
        state.SetMeta(key+"_compressed", "true")
        
        // Remove original large data
        state.Set(key, nil)
    }
    
    return nil
}
```

### 3. State Caching

```go
// State cache for expensive computations
type StateCache struct {
    cache map[string]State
    mu    sync.RWMutex
    ttl   time.Duration
}

func NewStateCache(ttl time.Duration) *StateCache {
    return &StateCache{
        cache: make(map[string]State),
        ttl:   ttl,
    }
}

func (sc *StateCache) Get(key string) (State, bool) {
    sc.mu.RLock()
    defer sc.mu.RUnlock()
    
    state, exists := sc.cache[key]
    if !exists {
        return nil, false
    }
    
    // Check TTL
    if timestamp, ok := state.GetMeta("cached_at"); ok {
        if cachedAt, err := time.Parse(time.RFC3339, timestamp); err == nil {
            if time.Since(cachedAt) > sc.ttl {
                delete(sc.cache, key)
                return nil, false
            }
        }
    }
    
    return state.Clone(), true
}

func (sc *StateCache) Set(key string, state State) {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    
    cachedState := state.Clone()
    cachedState.SetMeta("cached_at", time.Now().Format(time.RFC3339))
    sc.cache[key] = cachedState
}
```

## Debugging State Flow

### 1. State Tracing

```go
// State tracer for debugging
type StateTracer struct {
    traces []StateTrace
    mu     sync.RWMutex
}

type StateTrace struct {
    Timestamp time.Time `json:"timestamp"`
    AgentID   string    `json:"agent_id"`
    Operation string    `json:"operation"`
    Key       string    `json:"key,omitempty"`
    Value     any       `json:"value,omitempty"`
    StateSize int       `json:"state_size"`
}

func (st *StateTracer) TraceGet(agentID, key string, value any, stateSize int) {
    st.mu.Lock()
    defer st.mu.Unlock()
    
    st.traces = append(st.traces, StateTrace{
        Timestamp: time.Now(),
        AgentID:   agentID,
        Operation: "GET",
        Key:       key,
        Value:     value,
        StateSize: stateSize,
    })
}

func (st *StateTracer) TraceSet(agentID, key string, value any, stateSize int) {
    st.mu.Lock()
    defer st.mu.Unlock()
    
    st.traces = append(st.traces, StateTrace{
        Timestamp: time.Now(),
        AgentID:   agentID,
        Operation: "SET",
        Key:       key,
        Value:     value,
        StateSize: stateSize,
    })
}

// Traced state wrapper
type TracedState struct {
    State
    tracer  *StateTracer
    agentID string
}

func (ts *TracedState) Get(key string) (any, bool) {
    value, ok := ts.State.Get(key)
    if ts.tracer != nil {
        ts.tracer.TraceGet(ts.agentID, key, value, len(ts.State.Keys()))
    }
    return value, ok
}

func (ts *TracedState) Set(key string, value any) {
    ts.State.Set(key, value)
    if ts.tracer != nil {
        ts.tracer.TraceSet(ts.agentID, key, value, len(ts.State.Keys()))
    }
}
```

### 2. State Visualization

```go
// Generate state visualization
func VisualizeState(state State) string {
    var builder strings.Builder
    
    builder.WriteString("State Visualization\n")
    builder.WriteString("==================\n\n")
    
    // Data section
    builder.WriteString("Data:\n")
    for _, key := range state.Keys() {
        if value, ok := state.Get(key); ok {
            builder.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
        }
    }
    
    // Metadata section
    builder.WriteString("\nMetadata:\n")
    for _, key := range state.MetaKeys() {
        if value, ok := state.GetMeta(key); ok {
            builder.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
        }
    }
    
    return builder.String()
}

// Generate state diff
func DiffStates(before, after State) string {
    var builder strings.Builder
    
    builder.WriteString("State Diff\n")
    builder.WriteString("==========\n\n")
    
    // Check for added/modified data
    for _, key := range after.Keys() {
        afterValue, _ := after.Get(key)
        beforeValue, existed := before.Get(key)
        
        if !existed {
            builder.WriteString(fmt.Sprintf("+ %s: %v\n", key, afterValue))
        } else if !reflect.DeepEqual(beforeValue, afterValue) {
            builder.WriteString(fmt.Sprintf("~ %s: %v -> %v\n", key, beforeValue, afterValue))
        }
    }
    
    // Check for removed data
    for _, key := range before.Keys() {
        if _, exists := after.Get(key); !exists {
            beforeValue, _ := before.Get(key)
            builder.WriteString(fmt.Sprintf("- %s: %v\n", key, beforeValue))
        }
    }
    
    return builder.String()
}
```

## Best Practices

### 1. State Design Principles

```go
// Good: Clear, descriptive keys
state.Set("user_query", "What's the weather?")
state.Set("weather_data", weatherInfo)
state.Set("response_confidence", 0.95)

// Bad: Unclear, abbreviated keys
state.Set("q", "What's the weather?")
state.Set("wd", weatherInfo)
state.Set("conf", 0.95)

// Good: Consistent naming conventions
state.Set("user_profile", profile)
state.Set("user_preferences", preferences)
state.Set("user_history", history)

// Bad: Inconsistent naming
state.Set("userProfile", profile)
state.Set("user_prefs", preferences)
state.Set("UserHistory", history)
```

### 2. Error Handling

```go
// Always check if values exist
func SafeGetString(state State, key string) (string, error) {
    value, ok := state.Get(key)
    if !ok {
        return "", fmt.Errorf("key %s not found in state", key)
    }
    
    str, ok := value.(string)
    if !ok {
        return "", fmt.Errorf("key %s is not a string, got %T", key, value)
    }
    
    return str, nil
}

// Use type-safe getters
func GetUserID(state State) (string, error) {
    return SafeGetString(state, "user_id")
}

func GetConfidence(state State) (float64, error) {
    value, ok := state.Get("confidence")
    if !ok {
        return 0, errors.New("confidence not found in state")
    }
    
    confidence, ok := value.(float64)
    if !ok {
        return 0, fmt.Errorf("confidence is not a float64, got %T", value)
    }
    
    return confidence, nil
}
```

### 3. State Documentation

```go
// Document expected state structure
type ExpectedState struct {
    // Required fields
    UserQuery string `json:"user_query" required:"true" description:"The user's input query"`
    UserID    string `json:"user_id" required:"true" description:"Unique user identifier"`
    
    // Optional fields
    Context    string  `json:"context,omitempty" description:"Additional context for the query"`
    Confidence float64 `json:"confidence,omitempty" description:"Confidence score (0.0-1.0)"`
    
    // Metadata
    ProcessedBy string `json:"processed_by,omitempty" metadata:"true" description:"Agent that processed this state"`
    Timestamp   string `json:"timestamp,omitempty" metadata:"true" description:"Processing timestamp"`
}

// Validate state against expected structure
func ValidateExpectedState(state State) error {
    // Check required fields
    if _, ok := state.Get("user_query"); !ok {
        return errors.New("user_query is required")
    }
    
    if _, ok := state.Get("user_id"); !ok {
        return errors.New("user_id is required")
    }
    
    return nil
}
```

## Common Pitfalls and Solutions

### 1. State Mutation Issues

**Problem**: Modifying shared state without proper cloning.

```go
// Bad: Modifying shared state
func (a *BadAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    state.Set("modified", true) // Modifies input state!
    return AgentResult{OutputState: state}, nil
}

// Good: Clone before modifying
func (a *GoodAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    outputState := state.Clone()
    outputState.Set("modified", true)
    return AgentResult{OutputState: outputState}, nil
}
```

### 2. Memory Leaks

**Problem**: Accumulating large amounts of data in state without cleanup.

```go
// Bad: Accumulating data indefinitely
func (a *BadAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    history, _ := state.Get("processing_history")
    if history == nil {
        history = make([]string, 0)
    }
    
    // This grows indefinitely!
    newHistory := append(history.([]string), "processed by "+a.name)
    
    outputState := state.Clone()
    outputState.Set("processing_history", newHistory)
    return AgentResult{OutputState: outputState}, nil
}

// Good: Limit data accumulation
func (a *GoodAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    history, _ := state.Get("processing_history")
    if history == nil {
        history = make([]string, 0)
    }
    
    newHistory := append(history.([]string), "processed by "+a.name)
    
    // Keep only last 10 entries
    const maxHistory = 10
    if len(newHistory) > maxHistory {
        newHistory = newHistory[len(newHistory)-maxHistory:]
    }
    
    outputState := state.Clone()
    outputState.Set("processing_history", newHistory)
    return AgentResult{OutputState: outputState}, nil
}
```

### 3. Type Safety Issues

**Problem**: Assuming data types without checking.

```go
// Bad: Assuming types
func (a *BadAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    query := state.Get("user_query").(string) // Panic if not string!
    // ...
}

// Good: Type checking
func (a *GoodAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    queryValue, ok := state.Get("user_query")
    if !ok {
        return AgentResult{}, errors.New("user_query not found")
    }
    
    query, ok := queryValue.(string)
    if !ok {
        return AgentResult{}, fmt.Errorf("user_query is not a string, got %T", queryValue)
    }
    
    // Safe to use query
    // ...
}
```

## Conclusion

State management is the foundation of data flow in AgenticGoKit. By understanding how State objects work, how data flows between agents, and following best practices, you can build robust multi-agent systems that handle complex data transformations reliably.

Key takeaways:
- **Always clone state** before modification
- **Use clear, consistent naming** for state keys
- **Implement proper type checking** and error handling
- **Monitor state size** and prevent memory leaks
- **Document expected state structure** for maintainability
- **Use namespaced keys** to avoid conflicts
- **Implement validation** for critical state data

## Next Steps

- [Memory Systems](../memory-systems/README.md) - Learn about persistent state storage
- [Error Handling](error-handling.md) - Master robust error management with state
- [Debugging Guide](../debugging/README.md) - Learn to trace state flow
- [Advanced Patterns](../advanced/README.md) - Explore advanced state management patterns

## Further Reading

- [API Reference: State Interface](../../reference/api/state-event.md#state)
- [Examples: State Management Patterns](../../examples/)
- [Configuration Guide: State Settings](../../reference/api/configuration.md)
