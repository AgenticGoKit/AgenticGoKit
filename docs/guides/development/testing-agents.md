# Testing Agents

**Comprehensive testing strategies for multi-agent systems**

This guide covers testing approaches for AgenticGoKit applications, from unit testing individual agents to integration testing complex multi-agent workflows.

## Testing Philosophy

AgenticGoKit testing follows these principles:

- **Test at multiple levels**: Unit, integration, and end-to-end
- **Mock external dependencies**: LLM providers, databases, external APIs
- **Test orchestration patterns**: Verify agent interactions work correctly
- **Performance testing**: Ensure agents meet performance requirements
- **Deterministic testing**: Use mocks to ensure reproducible results

## Quick Start (10 minutes)

### 1. Basic Agent Unit Test

```go
package main

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/kunalkushwaha/agenticgokit/core"
)

func TestGreetingAgent(t *testing.T) {
    // Create mock LLM provider
    mockLLM := &MockLLMProvider{
        response: "Hello, World! Nice to meet you.",
    }
    
    // Create agent
    agent := NewGreetingAgent("greeter", mockLLM)
    
    // Create test event
    event := core.NewEvent("greeting", map[string]interface{}{
        "name": "World",
    })
    
    // Execute agent
    result, err := agent.Run(context.Background(), event, core.NewState())
    
    // Verify results
    require.NoError(t, err)
    assert.Contains(t, result.Data["response"], "Hello")
    assert.Contains(t, result.Data["response"], "World")
}

// Mock LLM Provider for testing
type MockLLMProvider struct {
    response string
    err      error
}

func (m *MockLLMProvider) Generate(ctx context.Context, prompt string) (string, error) {
    if m.err != nil {
        return "", m.err
    }
    return m.response, nil
}

func (m *MockLLMProvider) Name() string {
    return "mock"
}
```

### 2. Multi-Agent Integration Test

```go
func TestMultiAgentWorkflow(t *testing.T) {
    // Create Ollama provider for testing
    ollamaProvider := &OllamaProvider{
        config: OllamaConfig{
            BaseURL: "http://localhost:11434",
            Model:   "gemma3:1b",
        },
    }
    
    // Create agents
    agents := map[string]core.AgentHandler{
        "analyzer": NewAnalyzerAgent(mockLLM),
        "processor": NewProcessorAgent(mockLLM),
    }
    
    // Create sequential runner
    runner := core.CreateSequentialRunner(agents, []string{"analyzer", "processor"}, 30*time.Second)
    
    // Test workflow
    event := core.NewEvent("analyze", map[string]interface{}{
        "data": "test data",
    })
    
    _ = runner.Start(context.Background())
    defer runner.Stop()
    err := runner.Emit(event)
    
    require.NoError(t, err)
    assert.Len(t, results, 2)
    assert.Contains(t, results, "analyzer")
    assert.Contains(t, results, "processor")
}
```

## Unit Testing Patterns

### Testing Agent Logic

```go
func TestDataProcessingAgent(t *testing.T) {
    tests := []struct {
        name     string
        input    map[string]interface{}
        expected map[string]interface{}
        wantErr  bool
    }{
        {
            name: "valid data processing",
            input: map[string]interface{}{
                "numbers": []int{1, 2, 3, 4, 5},
            },
            expected: map[string]interface{}{
                "sum":     15,
                "average": 3.0,
                "count":   5,
            },
            wantErr: false,
        },
        {
            name: "empty data",
            input: map[string]interface{}{
                "numbers": []int{},
            },
            expected: nil,
            wantErr:  true,
        },
        {
            name: "invalid input type",
            input: map[string]interface{}{
                "numbers": "not a slice",
            },
            expected: nil,
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            agent := NewDataProcessingAgent()
            event := core.NewEvent("process", tt.input)
            
            result, err := agent.Run(context.Background(), event, core.NewState())
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            for key, expectedValue := range tt.expected {
                assert.Equal(t, expectedValue, result.Data[key])
            }
        })
    }
}
```

### Testing State Management

```go
func TestAgentStateHandling(t *testing.T) {
    agent := NewStatefulAgent()
    
    // Test state initialization
    initialState := core.NewState()
    initialState.Set("counter", 0)
    
    event := core.NewEvent("increment", nil)
    
    // First execution
    result1, err := agent.Run(context.Background(), event, initialState)
    require.NoError(t, err)
    assert.Equal(t, 1, result1.Data["counter"])
    
    // Second execution with updated state
    result2, err := agent.Run(context.Background(), event, result1.State)
    require.NoError(t, err)
    assert.Equal(t, 2, result2.Data["counter"])
    
    // Verify state persistence
    assert.Equal(t, 2, result2.State.Get("counter"))
}
```

### Testing Error Handling

```go
func TestAgentErrorHandling(t *testing.T) {
    // Test with failing LLM provider
    failingLLM := &MockLLMProvider{
        err: errors.New("LLM service unavailable"),
    }
    
    agent := NewResilientAgent(failingLLM)
    event := core.NewEvent("process", map[string]interface{}{
        "input": "test",
    })
    
    result, err := agent.Run(context.Background(), event, core.NewState())
    
    // Agent should handle LLM failure gracefully
    require.NoError(t, err)
    assert.Contains(t, result.Data["response"], "service temporarily unavailable")
    assert.True(t, result.Data["fallback_used"].(bool))
}
```

## Integration Testing

### Testing Orchestration Patterns

```go
func TestSequentialOrchestration(t *testing.T) {
    // Create agents that depend on each other
    agents := map[string]core.AgentHandler{
        "step1": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            input := event.Data["input"].(string)
            return core.AgentResult{
                Data: map[string]interface{}{
                    "step1_output": input + "_processed",
                },
            }, nil
        }),
        
        "step2": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            step1Output := state.Data["step1_output"].(string)
            return core.AgentResult{
                Data: map[string]interface{}{
                    "final_output": step1Output + "_finalized",
                },
            }, nil
        }),
    }
    
    runner := core.CreateSequentialRunner(agents, []string{"step1", "step2"}, 30*time.Second)
    
    event := core.NewEvent("process", map[string]interface{}{
        "input": "test",
    })
    
    _ = runner.Start(context.Background())
    defer runner.Stop()
    err := runner.Emit(event)
    
    require.NoError(t, err)
    assert.Equal(t, "test_processed_finalized", results["step2"].Data["final_output"])
}

func TestCollaborativeOrchestration(t *testing.T) {
    // Create agents that work in parallel
    agents := map[string]core.AgentHandler{
        "analyzer": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            return core.AgentResult{
                Data: map[string]interface{}{
                    "analysis": "positive sentiment",
                },
            }, nil
        }),
        
        "summarizer": core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
            return core.AgentResult{
                Data: map[string]interface{}{
                    "summary": "brief summary",
                },
            }, nil
        }),
    }
    
    runner, _ := core.NewRunnerFromConfig("agentflow.toml")
    
    event := core.NewEvent("analyze", map[string]interface{}{
        "text": "This is great!",
    })
    
    _ = runner.Start(context.Background())
    defer runner.Stop()
    err := runner.Emit(event)
    
    require.NoError(t, err)
    assert.Len(t, results, 2)
    assert.Equal(t, "positive sentiment", results["analyzer"].Data["analysis"])
    assert.Equal(t, "brief summary", results["summarizer"].Data["summary"])
}
```

### Testing Memory Integration

```go
func TestAgentWithMemory(t *testing.T) {
    // Create in-memory provider for testing
    memoryConfig := core.AgentMemoryConfig{
        Provider:   "memory",
        Connection: "memory",
        Dimensions: 384,
        Embedding: core.EmbeddingConfig{
            Provider: "dummy",
        },
    }
    
    memory, err := core.NewMemory(memoryConfig)
    require.NoError(t, err)
    defer memory.Close()
    
    // Create memory-enabled agent
    agent := NewMemoryEnabledAgent(memory)
    
    ctx := memory.SetSession(context.Background(), "test-session")
    
    // First interaction - store information
    event1 := core.NewEvent("remember", map[string]interface{}{
        "info": "User likes coffee",
    })
    
    result1, err := agent.Run(ctx, event1, core.NewState())
    require.NoError(t, err)
    assert.Contains(t, result1.Data["response"], "remembered")
    
    // Second interaction - recall information
    event2 := core.NewEvent("recall", map[string]interface{}{
        "query": "What does the user like?",
    })
    
    result2, err := agent.Run(ctx, event2, core.NewState())
    require.NoError(t, err)
    assert.Contains(t, result2.Data["response"], "coffee")
}
```

### Testing MCP Tool Integration

```go
func TestAgentWithMCPTools(t *testing.T) {
    // Create mock MCP manager
    mockMCP := &MockMCPManager{
        tools: []core.ToolSchema{
            {
                Name:        "search",
                Description: "Search for information",
            },
        },
        responses: map[string]interface{}{
            "search": map[string]interface{}{
                "results": []string{"Mock search result"},
            },
        },
    }
    
    mockLLM := &MockLLMProvider{
        response: `I'll search for that information.
        
<tool_call>
{"name": "search", "args": {"query": "test query"}}
</tool_call>

Based on the search results, here's what I found...`,
    }
    
    agent := NewToolEnabledAgent(mockLLM, mockMCP)
    
    event := core.NewEvent("query", map[string]interface{}{
        "question": "Search for test information",
    })
    
    result, err := agent.Run(context.Background(), event, core.NewState())
    
    require.NoError(t, err)
    assert.True(t, result.Data["tools_used"].(bool))
    assert.Contains(t, result.Data["tool_results"], "Mock search result")
}

type MockMCPManager struct {
    tools     []core.ToolSchema
    responses map[string]interface{}
}

func (m *MockMCPManager) ListTools(ctx context.Context) ([]core.ToolSchema, error) {
    return m.tools, nil
}

func (m *MockMCPManager) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
    if response, exists := m.responses[name]; exists {
        return response, nil
    }
    return nil, fmt.Errorf("tool not found: %s", name)
}
```

## Performance Testing

### Load Testing

```go
func TestAgentPerformance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }
    
    agent := NewPerformantAgent()
    
    // Test concurrent execution
    concurrency := 10
    iterations := 100
    
    var wg sync.WaitGroup
    results := make(chan time.Duration, concurrency*iterations)
    
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for j := 0; j < iterations; j++ {
                start := time.Now()
                
                event := core.NewEvent("process", map[string]interface{}{
                    "data": fmt.Sprintf("test_%d_%d", i, j),
                })
                
                _, err := agent.Run(context.Background(), event, core.NewState())
                duration := time.Since(start)
                
                require.NoError(t, err)
                results <- duration
            }
        }()
    }
    
    wg.Wait()
    close(results)
    
    // Analyze performance
    var total time.Duration
    var max time.Duration
    count := 0
    
    for duration := range results {
        total += duration
        if duration > max {
            max = duration
        }
        count++
    }
    
    average := total / time.Duration(count)
    
    t.Logf("Performance Results:")
    t.Logf("  Total requests: %d", count)
    t.Logf("  Average duration: %v", average)
    t.Logf("  Max duration: %v", max)
    t.Logf("  Requests per second: %.2f", float64(count)/total.Seconds())
    
    // Assert performance requirements
    assert.Less(t, average, 100*time.Millisecond, "Average response time should be under 100ms")
    assert.Less(t, max, 500*time.Millisecond, "Max response time should be under 500ms")
}
```

### Memory Usage Testing

```go
func TestAgentMemoryUsage(t *testing.T) {
    agent := NewMemoryEfficientAgent()
    
    // Measure initial memory
    var m1 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Run many iterations
    for i := 0; i < 1000; i++ {
        event := core.NewEvent("process", map[string]interface{}{
            "data": strings.Repeat("x", 1000), // 1KB of data
        })
        
        _, err := agent.Run(context.Background(), event, core.NewState())
        require.NoError(t, err)
    }
    
    // Force garbage collection and measure final memory
    runtime.GC()
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)
    
    // Calculate memory increase
    memoryIncrease := m2.Alloc - m1.Alloc
    
    t.Logf("Memory usage:")
    t.Logf("  Initial: %d KB", m1.Alloc/1024)
    t.Logf("  Final: %d KB", m2.Alloc/1024)
    t.Logf("  Increase: %d KB", memoryIncrease/1024)
    
    // Assert reasonable memory usage (adjust threshold as needed)
    assert.Less(t, memoryIncrease, uint64(10*1024*1024), "Memory increase should be less than 10MB")
}
```

## End-to-End Testing

### Complete Workflow Testing

```go
func TestCompleteWorkflow(t *testing.T) {
    // Set up complete system
    memoryConfig := core.AgentMemoryConfig{
        Provider:   "memory",
        Connection: "memory",
        Dimensions: 384,
        Embedding: core.EmbeddingConfig{
            Provider: "dummy",
        },
    }
    
    memory, err := core.NewMemory(memoryConfig)
    require.NoError(t, err)
    defer memory.Close()
    
    mockLLM := &MockLLMProvider{
        response: "I understand your request and will process it accordingly.",
    }
    
    // Create complete agent system
    agents := map[string]core.AgentHandler{
        "intake":    NewIntakeAgent(mockLLM),
        "processor": NewProcessorAgent(mockLLM, memory),
        "responder": NewResponderAgent(mockLLM),
    }
    
    runner := core.CreateSequentialRunner(agents, []string{"intake", "processor", "responder"}, 60*time.Second)
    
    // Test complete user interaction
    ctx := memory.SetSession(context.Background(), "e2e-test-session")
    
    event := core.NewEvent("user_request", map[string]interface{}{
        "message": "I need help with my project",
        "user_id": "test_user",
    })
    
    _ = runner.Start(ctx)
    defer runner.Stop()
    err := runner.Emit(event)
    
    require.NoError(t, err)
    assert.Len(t, results, 3)
    
    // Verify each stage completed successfully
    assert.Contains(t, results["intake"].Data, "processed")
    assert.Contains(t, results["processor"].Data, "analyzed")
    assert.Contains(t, results["responder"].Data, "response")
    
    // Verify final response quality
    finalResponse := results["responder"].Data["response"].(string)
    assert.NotEmpty(t, finalResponse)
    assert.Greater(t, len(finalResponse), 10)
}
```

## Test Utilities and Helpers

### Test Data Builders

```go
// Event builder for consistent test data
type EventBuilder struct {
    eventType string
    data      map[string]interface{}
    metadata  map[string]interface{}
}

func NewEventBuilder(eventType string) *EventBuilder {
    return &EventBuilder{
        eventType: eventType,
        data:      make(map[string]interface{}),
        metadata:  make(map[string]interface{}),
    }
}

func (eb *EventBuilder) WithData(key string, value interface{}) *EventBuilder {
    eb.data[key] = value
    return eb
}

func (eb *EventBuilder) WithMetadata(key string, value interface{}) *EventBuilder {
    eb.metadata[key] = value
    return eb
}

func (eb *EventBuilder) Build() core.Event {
    event := core.NewEvent(eb.eventType, eb.data)
    for k, v := range eb.metadata {
        event.Metadata[k] = v
    }
    return event
}

// Usage in tests
func TestWithEventBuilder(t *testing.T) {
    event := NewEventBuilder("process").
        WithData("input", "test data").
        WithData("priority", "high").
        WithMetadata("user_id", "123").
        Build()
    
    // Use event in test...
}
```

### Agent Test Harness

```go
type AgentTestHarness struct {
    agent   core.AgentHandler
    mockLLM *MockLLMProvider
    memory  core.Memory
    ctx     context.Context
}

func NewAgentTestHarness(agent core.AgentHandler) *AgentTestHarness {
    mockLLM := &MockLLMProvider{}
    
    memoryConfig := core.AgentMemoryConfig{
        Provider:   "memory",
        Connection: "memory",
        Dimensions: 384,
        Embedding: core.EmbeddingConfig{
            Provider: "dummy",
        },
    }
    
    memory, _ := core.NewMemory(memoryConfig)
    ctx := memory.SetSession(context.Background(), "test-session")
    
    return &AgentTestHarness{
        agent:   agent,
        mockLLM: mockLLM,
        memory:  memory,
        ctx:     ctx,
    }
}

func (h *AgentTestHarness) SetLLMResponse(response string) {
    h.mockLLM.response = response
}

func (h *AgentTestHarness) SetLLMError(err error) {
    h.mockLLM.err = err
}

func (h *AgentTestHarness) Execute(event core.Event) (core.AgentResult, error) {
    return h.agent.Run(h.ctx, event, core.NewState())
}

func (h *AgentTestHarness) Cleanup() {
    if h.memory != nil {
        h.memory.Close()
    }
}

// Usage in tests
func TestWithHarness(t *testing.T) {
    harness := NewAgentTestHarness(NewMyAgent())
    defer harness.Cleanup()
    
    harness.SetLLMResponse("Expected response")
    
    event := core.NewEvent("test", map[string]interface{}{
        "input": "test data",
    })
    
    result, err := harness.Execute(event)
    
    require.NoError(t, err)
    assert.Equal(t, "Expected response", result.Data["response"])
}
```

## Continuous Integration

### GitHub Actions Test Configuration

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: pgvector/pgvector:pg15
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_DB: agentflow_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests
      run: go test -v -short ./...
    
    - name: Run integration tests
      run: go test -v ./...
      env:
        TEST_DB_URL: postgres://postgres:password@localhost:5432/agentflow_test?sslmode=disable
    
    - name: Run performance tests
      run: go test -v -run TestPerformance ./...
    
    - name: Generate coverage report
      run: go test -coverprofile=coverage.out ./...
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

## Best Practices

### Test Organization

```go
// Organize tests by functionality
func TestAgentCore(t *testing.T) {
    t.Run("BasicExecution", testBasicExecution)
    t.Run("ErrorHandling", testErrorHandling)
    t.Run("StateManagement", testStateManagement)
}

func TestAgentIntegration(t *testing.T) {
    t.Run("WithMemory", testWithMemory)
    t.Run("WithTools", testWithTools)
    t.Run("WithOrchestration", testWithOrchestration)
}

func TestAgentPerformance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance tests in short mode")
    }
    
    t.Run("LoadTest", testLoadTest)
    t.Run("MemoryUsage", testMemoryUsage)
    t.Run("Concurrency", testConcurrency)
}
```

### Test Data Management

```go
// Use table-driven tests for multiple scenarios
func TestDataProcessing(t *testing.T) {
    testCases := []struct {
        name     string
        input    interface{}
        expected interface{}
        wantErr  bool
    }{
        // Test cases here...
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation...
        })
    }
}

// Use test fixtures for complex data
func loadTestFixture(t *testing.T, filename string) map[string]interface{} {
    data, err := os.ReadFile(filepath.Join("testdata", filename))
    require.NoError(t, err)
    
    var result map[string]interface{}
    err = json.Unmarshal(data, &result)
    require.NoError(t, err)
    
    return result
}
```

## Next Steps

- **[Debugging](debugging.md)** - Debug agent interactions effectively
- **[Best Practices](best-practices.md)** - Development best practices
- **[Production Deployment](../deployment/README.md)** - Production deployment