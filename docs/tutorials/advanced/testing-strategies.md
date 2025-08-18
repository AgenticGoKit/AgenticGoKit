# Testing Strategies for Multi-Agent Systems

**Comprehensive testing approaches for building reliable and maintainable agent systems**

Testing multi-agent systems presents unique challenges due to their distributed nature, asynchronous communication, and external dependencies. This guide covers testing strategies, patterns, and tools specifically designed for AgenticGoKit applications.

## 🧪 Testing Pyramid for Agent Systems

```mermaid
flowchart TD
    subgraph "End-to-End Tests (10%)"
        E2E1[Full System Tests]
        E2E2[User Journey Tests]
        E2E3[Performance Tests]
    end
    
    subgraph "Integration Tests (30%)"
        INT1[Agent Interaction Tests]
        INT2[External Service Tests]
        INT3[Orchestration Tests]
        INT4[Memory System Tests]
        INT5[MCP Tool Tests]
    end
    
    subgraph "Unit Tests (60%)"
        UNIT1[Individual Agent Tests]
        UNIT2[Business Logic Tests]
        UNIT3[Utility Function Tests]
        UNIT4[Error Handling Tests]
        UNIT5[State Management Tests]
    end
    
    E2E1 --> INT1
    E2E2 --> INT2
    E2E3 --> INT3
    INT1 --> UNIT1
    INT2 --> UNIT2
    INT3 --> UNIT3
    INT4 --> UNIT4
    INT5 --> UNIT5
    
    classDef e2e fill:#ffebee,stroke:#c62828,stroke-width:2px
    classDef integration fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef unit fill:#e3f2fd,stroke:#1565c0,stroke-width:2px
    
    class E2E1,E2E2,E2E3 e2e
    class INT1,INT2,INT3,INT4,INT5 integration
    class UNIT1,UNIT2,UNIT3,UNIT4,UNIT5 unit
```

## 🔧 Unit Testing Strategies

### **1. Testing Individual Agents**

```go
package agents

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/kunalkushwaha/agenticgokit/core"
)

// Mock LLM Provider for testing
type MockLLMProvider struct {
    mock.Mock
}

func (m *MockLLMProvider) GenerateResponse(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
    args := m.Called(ctx, prompt, options)
    return args.String(0), args.Error(1)
}

func TestAnalyzerAgent_Execute(t *testing.T) {
    tests := []struct {
        name           string
        event          core.Event
        state          *core.State
        mockResponse   string
        mockError      error
        expectedResult map[string]interface{}
        expectedError  string
    }{
        {
            name: "successful analysis",
            event: core.NewEvent("analyze", map[string]interface{}{
                "data": "Sample data to analyze",
            }),
            state: &core.State{
                SessionID: "test-session",
                Data: map[string]interface{}{
                    "context": "test context",
                },
            },
            mockResponse: "Analysis complete: The data shows positive trends",
            expectedResult: map[string]interface{}{
                "analysis": "Analysis complete: The data shows positive trends",
                "agent":    "analyzer",
                "status":   "completed",
            },
        },
        {
            name: "empty data handling",
            event: core.NewEvent("analyze", map[string]interface{}{
                "data": "",
            }),
            state: &core.State{SessionID: "test-session"},
            expectedError: "no data provided for analysis",
        },
        {
            name: "LLM provider error",
            event: core.NewEvent("analyze", map[string]interface{}{
                "data": "Sample data",
            }),
            state: &core.State{SessionID: "test-session"},
            mockError: errors.New("LLM service unavailable"),
            expectedError: "LLM request failed: LLM service unavailable",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock
            mockProvider := new(MockLLMProvider)
            if tt.mockResponse != "" || tt.mockError != nil {
                mockProvider.On("GenerateResponse", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
                    Return(tt.mockResponse, tt.mockError)
            }
            
            // Create agent with mock
            agent := NewAnalyzerAgent("analyzer", mockProvider)
            
            // Execute test
            result, err := agent.Execute(context.Background(), tt.event, tt.state)
            
            // Assertions
            if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
                for key, expectedValue := range tt.expectedResult {
                    assert.Equal(t, expectedValue, result.Data[key])
                }
            }
            
            // Verify mock expectations
            mockProvider.AssertExpectations(t)
        })
    }
}
```

### **2. Testing State Management**

```go
func TestStateTransformations(t *testing.T) {
    tests := []struct {
        name           string
        initialState   *core.State
        transformation func(*core.State) *core.State
        expectedData   map[string]interface{}
    }{
        {
            name: "add analysis result",
            initialState: &core.State{
                SessionID: "test",
                Data: map[string]interface{}{
                    "input": "test data",
                },
            },
            transformation: func(state *core.State) *core.State {
                newState := state.Clone()
                newState.Data["analysis"] = "completed"
                newState.Data["confidence"] = 0.95
                return newState
            },
            expectedData: map[string]interface{}{
                "input":      "test data",
                "analysis":   "completed",
                "confidence": 0.95,
            },
        },
        {
            name: "merge states",
            initialState: &core.State{
                SessionID: "test",
                Data: map[string]interface{}{
                    "step1": "completed",
                },
            },
            transformation: func(state *core.State) *core.State {
                additionalData := map[string]interface{}{
                    "step2": "completed",
                    "final": true,
                }
                return state.Merge(additionalData)
            },
            expectedData: map[string]interface{}{
                "step1": "completed",
                "step2": "completed",
                "final": true,
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.transformation(tt.initialState)
            
            for key, expected := range tt.expectedData {
                assert.Equal(t, expected, result.Data[key])
            }
        })
    }
}
```

### **3. Testing Error Handling**

```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name          string
        setupMock     func(*MockLLMProvider)
        expectedError string
        shouldRetry   bool
    }{
        {
            name: "timeout error",
            setupMock: func(m *MockLLMProvider) {
                m.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).
                    Return("", context.DeadlineExceeded)
            },
            expectedError: "context deadline exceeded",
            shouldRetry:   true,
        },
        {
            name: "rate limit error",
            setupMock: func(m *MockLLMProvider) {
                m.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).
                    Return("", errors.New("rate limit exceeded"))
            },
            expectedError: "rate limit exceeded",
            shouldRetry:   true,
        },
        {
            name: "authentication error",
            setupMock: func(m *MockLLMProvider) {
                m.On("GenerateResponse", mock.Anything, mock.Anything, mock.Anything).
                    Return("", errors.New("invalid API key"))
            },
            expectedError: "invalid API key",
            shouldRetry:   false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockProvider := new(MockLLMProvider)
            tt.setupMock(mockProvider)
            
            agent := NewResilientAgent("test", mockProvider)
            
            event := core.NewEvent("test", "data")
            state := &core.State{SessionID: "test"}
            
            _, err := agent.Execute(context.Background(), event, state)
            
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedError)
            
            // Test retry behavior
            if tt.shouldRetry {
                // Verify that retries were attempted
                assert.True(t, mockProvider.AssertNumberOfCalls(t, "GenerateResponse", 3))
            } else {
                // Verify that no retries were attempted
                assert.True(t, mockProvider.AssertNumberOfCalls(t, "GenerateResponse", 1))
            }
        })
    }
}
```

## 🔗 Integration Testing

### **1. Testing Agent Interactions**

```go
func TestAgentOrchestration(t *testing.T) {
    // Setup test agents
    mockProvider := new(MockLLMProvider)
    mockProvider.On("GenerateResponse", mock.Anything, mock.MatchedBy(func(prompt string) bool {
        return strings.Contains(prompt, "analyze")
    }), mock.Anything).Return("Analysis: Data processed", nil)
    
    mockProvider.On("GenerateResponse", mock.Anything, mock.MatchedBy(func(prompt string) bool {
        return strings.Contains(prompt, "validate")
    }), mock.Anything).Return("Validation: Data is valid", nil)
    
    analyzer := NewAnalyzerAgent("analyzer", mockProvider)
    validator := NewValidatorAgent("validator", mockProvider)
    
    // Create sequential runner
    agents := map[string]core.AgentHandler{
        "analyzer":  analyzer,
        "validator": validator,
    }
    
    config := core.RunnerConfig{
        OrchestrationMode: core.OrchestrationSequential,
        SequentialAgents:  []string{"analyzer", "validator"},
        Timeout:          30 * time.Second,
    }
    
    runner := core.NewRunnerWithConfig(config)
    for name, agent := range agents {
        runner.RegisterAgent(name, agent)
    }
    
    // Test sequential execution
    event := core.NewEvent("process", map[string]interface{}{
        "data": "test data for processing",
    })
    
    _ = runner.Start(context.Background())
    defer runner.Stop()
    err := runner.Emit(event)
    
    assert.NoError(t, err)
    assert.Len(t, results, 2)
    
    // Verify execution order and data flow
    analyzerResult := results[0]
    validatorResult := results[1]
    
    assert.Equal(t, "analyzer", analyzerResult.Agent)
    assert.Contains(t, analyzerResult.Data["analysis"], "Data processed")
    
    assert.Equal(t, "validator", validatorResult.Agent)
    assert.Contains(t, validatorResult.Data["validation"], "Data is valid")
}
```

### **2. Testing Memory System Integration**

```go
func TestMemorySystemIntegration(t *testing.T) {
    // Setup in-memory test database
    memoryProvider := memory.NewInMemoryProvider()
    
    // Create agent with memory
    agent := NewMemoryEnabledAgent("test-agent", mockLLMProvider, memoryProvider)
    
    // Test storing and retrieving memories
    t.Run("store and retrieve memory", func(t *testing.T) {
        event := core.NewEvent("remember", map[string]interface{}{
            "fact": "The sky is blue",
            "category": "general_knowledge",
        })
        
        state := &core.State{
            SessionID: "test-session",
            UserID:    "test-user",
        }
        
        result, err := agent.Execute(context.Background(), event, state)
        assert.NoError(t, err)
        assert.Equal(t, "stored", result.Data["status"])
        
        // Verify memory was stored
        memories, err := memoryProvider.Search(context.Background(), "sky blue", 5)
        assert.NoError(t, err)
        assert.Len(t, memories, 1)
        assert.Contains(t, memories[0].Content, "sky is blue")
    })
    
    t.Run("RAG retrieval", func(t *testing.T) {
        // Setup mock to expect context from memory
        mockProvider.On("GenerateResponse", mock.Anything, mock.MatchedBy(func(prompt string) bool {
            return strings.Contains(prompt, "sky is blue")
        }), mock.Anything).Return("Based on the stored knowledge, the sky appears blue due to light scattering", nil)
        
        event := core.NewEvent("query", map[string]interface{}{
            "question": "What color is the sky?",
        })
        
        state := &core.State{SessionID: "test-session"}
        
        result, err := agent.Execute(context.Background(), event, state)
        assert.NoError(t, err)
        assert.Contains(t, result.Data["response"], "blue due to light scattering")
    })
}
```

### **3. Testing MCP Tool Integration**

```go
func TestMCPToolIntegration(t *testing.T) {
    // Setup mock MCP server
    mockServer := &MockMCPServer{}
    mockServer.On("ExecuteTool", "web_search", map[string]interface{}{
        "query": "AgenticGoKit testing",
    }).Return(map[string]interface{}{
        "results": []string{"AgenticGoKit is a Go framework for building multi-agent systems"},
    }, nil)
    
    // Create agent with MCP tools
    toolManager := mcp.NewToolManager()
    toolManager.RegisterServer("test-server", mockServer)
    
    agent := NewToolEnabledAgent("test-agent", mockLLMProvider, toolManager)
    
    event := core.NewEvent("search", map[string]interface{}{
        "query": "AgenticGoKit testing",
        "tool":  "web_search",
    })
    
    state := &core.State{SessionID: "test-session"}
    
    result, err := agent.Execute(context.Background(), event, state)
    
    assert.NoError(t, err)
    assert.Contains(t, result.Data["search_results"], "Go framework")
    
    mockServer.AssertExpectations(t)
}
```

## 🌐 End-to-End Testing

### **1. Full System Tests**

```go
func TestFullSystemWorkflow(t *testing.T) {
    // Setup complete system
    system := setupTestSystem(t)
    defer system.Cleanup()
    
    // Test complete user workflow
    t.Run("research assistant workflow", func(t *testing.T) {
        // Step 1: User asks a question
        request := &api.Request{
            Type: "research_query",
            Data: map[string]interface{}{
                "question": "What are the latest developments in AI agent frameworks?",
                "depth":    "comprehensive",
            },
            UserID:    "test-user",
            SessionID: "test-session",
        }
        
        response, err := system.ProcessRequest(context.Background(), request)
        assert.NoError(t, err)
        
        // Verify response structure
        assert.NotNil(t, response)
        assert.Equal(t, "success", response.Status)
        assert.Contains(t, response.Data, "research_results")
        assert.Contains(t, response.Data, "sources")
        assert.Contains(t, response.Data, "summary")
        
        // Verify agent execution trace
        trace := response.Trace
        assert.NotEmpty(t, trace)
        
        // Should have executed: researcher -> analyzer -> validator -> formatter
        expectedAgents := []string{"researcher", "analyzer", "validator", "formatter"}
        assert.Len(t, trace, len(expectedAgents))
        
        for i, expectedAgent := range expectedAgents {
            assert.Equal(t, expectedAgent, trace[i].Agent)
            assert.Equal(t, "completed", trace[i].Status)
        }
    })
    
    t.Run("error handling workflow", func(t *testing.T) {
        // Simulate external service failure
        system.SimulateServiceFailure("web_search")
        
        request := &api.Request{
            Type: "research_query",
            Data: map[string]interface{}{
                "question": "Test question",
            },
            UserID:    "test-user",
            SessionID: "test-session-2",
        }
        
        response, err := system.ProcessRequest(context.Background(), request)
        
        // Should handle gracefully with fallback
        assert.NoError(t, err)
        assert.Equal(t, "partial_success", response.Status)
        assert.Contains(t, response.Data, "fallback_response")
        assert.Contains(t, response.Warnings, "web search unavailable")
    })
}
```

### **2. Performance Testing**

```go
func TestSystemPerformance(t *testing.T) {
    system := setupTestSystem(t)
    defer system.Cleanup()
    
    t.Run("concurrent request handling", func(t *testing.T) {
        concurrency := 10
        requestsPerWorker := 5
        
        var wg sync.WaitGroup
        results := make(chan time.Duration, concurrency*requestsPerWorker)
        errors := make(chan error, concurrency*requestsPerWorker)
        
        for i := 0; i < concurrency; i++ {
            wg.Add(1)
            go func(workerID int) {
                defer wg.Done()
                
                for j := 0; j < requestsPerWorker; j++ {
                    start := time.Now()
                    
                    request := &api.Request{
                        Type: "simple_query",
                        Data: map[string]interface{}{
                            "question": fmt.Sprintf("Test question %d-%d", workerID, j),
                        },
                        UserID:    fmt.Sprintf("user-%d", workerID),
                        SessionID: fmt.Sprintf("session-%d-%d", workerID, j),
                    }
                    
                    _, err := system.ProcessRequest(context.Background(), request)
                    duration := time.Since(start)
                    
                    if err != nil {
                        errors <- err
                    } else {
                        results <- duration
                    }
                }
            }(i)
        }
        
        wg.Wait()
        close(results)
        close(errors)
        
        // Collect results
        var durations []time.Duration
        for duration := range results {
            durations = append(durations, duration)
        }
        
        var errorList []error
        for err := range errors {
            errorList = append(errorList, err)
        }
        
        // Assertions
        assert.Empty(t, errorList, "No errors should occur during concurrent processing")
        assert.Len(t, durations, concurrency*requestsPerWorker)
        
        // Performance assertions
        avgDuration := calculateAverage(durations)
        maxDuration := calculateMax(durations)
        
        assert.Less(t, avgDuration, 5*time.Second, "Average response time should be under 5 seconds")
        assert.Less(t, maxDuration, 10*time.Second, "Max response time should be under 10 seconds")
        
        t.Logf("Performance results: avg=%v, max=%v, total_requests=%d", 
            avgDuration, maxDuration, len(durations))
    })
}
```

## 🎭 Test Doubles and Mocking

### **1. Mock External Services**

```go
type MockExternalService struct {
    mock.Mock
    responses map[string]interface{}
    delays    map[string]time.Duration
    failures  map[string]error
}

func NewMockExternalService() *MockExternalService {
    return &MockExternalService{
        responses: make(map[string]interface{}),
        delays:    make(map[string]time.Duration),
        failures:  make(map[string]error),
    }
}

func (m *MockExternalService) SetResponse(endpoint string, response interface{}) {
    m.responses[endpoint] = response
}

func (m *MockExternalService) SetDelay(endpoint string, delay time.Duration) {
    m.delays[endpoint] = delay
}

func (m *MockExternalService) SetFailure(endpoint string, err error) {
    m.failures[endpoint] = err
}

func (m *MockExternalService) Call(ctx context.Context, endpoint string, params map[string]interface{}) (interface{}, error) {
    // Simulate delay
    if delay, exists := m.delays[endpoint]; exists {
        select {
        case <-time.After(delay):
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    
    // Simulate failure
    if err, exists := m.failures[endpoint]; exists {
        return nil, err
    }
    
    // Return mock response
    if response, exists := m.responses[endpoint]; exists {
        return response, nil
    }
    
    return nil, errors.New("endpoint not mocked")
}
```

### **2. Test Fixtures and Builders**

```go
type TestSystemBuilder struct {
    agents      map[string]core.AgentHandler
    mockServices map[string]*MockExternalService
    config      core.RunnerConfig
}

func NewTestSystemBuilder() *TestSystemBuilder {
    return &TestSystemBuilder{
        agents:       make(map[string]core.AgentHandler),
        mockServices: make(map[string]*MockExternalService),
        config: core.RunnerConfig{
            OrchestrationMode: core.OrchestrationSequential,
            Timeout:          30 * time.Second,
        },
    }
}

func (b *TestSystemBuilder) WithAgent(name string, agent core.AgentHandler) *TestSystemBuilder {
    b.agents[name] = agent
    return b
}

func (b *TestSystemBuilder) WithMockService(name string, service *MockExternalService) *TestSystemBuilder {
    b.mockServices[name] = service
    return b
}

func (b *TestSystemBuilder) WithOrchestration(mode core.OrchestrationMode, agents []string) *TestSystemBuilder {
    b.config.OrchestrationMode = mode
    switch mode {
    case core.OrchestrationSequential:
        b.config.SequentialAgents = agents
    case core.OrchestrationCollaborative:
        b.config.CollaborativeAgents = agents
    }
    return b
}

func (b *TestSystemBuilder) Build() *TestSystem {
    runner := core.NewRunnerWithConfig(b.config)
    
    for name, agent := range b.agents {
        runner.RegisterAgent(name, agent)
    }
    
    return &TestSystem{
        runner:       runner,
        mockServices: b.mockServices,
    }
}

type TestSystem struct {
    runner       core.Runner
    mockServices map[string]*MockExternalService
}

func (ts *TestSystem) ProcessRequest(ctx context.Context, request *api.Request) (*api.Response, error) {
    event := core.NewEvent(request.Type, request.Data)
    state := &core.State{
        SessionID: request.SessionID,
        UserID:    request.UserID,
    }
    
    _ = ts.runner.Start(ctx)
    defer ts.runner.Stop()
    err := ts.runner.Emit(event)
    if err != nil {
        return nil, err
    }
    
    return &api.Response{
        Status: "success",
        Data:   results[len(results)-1].Data,
        Trace:  convertToTrace(results),
    }, nil
}

func (ts *TestSystem) SimulateServiceFailure(serviceName string) {
    if service, exists := ts.mockServices[serviceName]; exists {
        service.SetFailure("*", errors.New("simulated service failure"))
    }
}

func (ts *TestSystem) Cleanup() {
    // Cleanup resources
    for _, service := range ts.mockServices {
        service.AssertExpectations(&testing.T{})
    }
}
```

## 📊 Test Data Management

### **1. Test Data Builders**

```go
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

func (b *EventBuilder) WithData(key string, value interface{}) *EventBuilder {
    b.data[key] = value
    return b
}

func (b *EventBuilder) WithMetadata(key string, value interface{}) *EventBuilder {
    b.metadata[key] = value
    return b
}

func (b *EventBuilder) Build() core.Event {
    event := core.NewEvent(b.eventType, b.data)
    for key, value := range b.metadata {
        event.Metadata[key] = value
    }
    return event
}

type StateBuilder struct {
    sessionID string
    userID    string
    data      map[string]interface{}
}

func NewStateBuilder() *StateBuilder {
    return &StateBuilder{
        data: make(map[string]interface{}),
    }
}

func (b *StateBuilder) WithSession(sessionID string) *StateBuilder {
    b.sessionID = sessionID
    return b
}

func (b *StateBuilder) WithUser(userID string) *StateBuilder {
    b.userID = userID
    return b
}

func (b *StateBuilder) WithData(key string, value interface{}) *StateBuilder {
    b.data[key] = value
    return b
}

func (b *StateBuilder) Build() *core.State {
    return &core.State{
        SessionID: b.sessionID,
        UserID:    b.userID,
        Data:      b.data,
    }
}
```

### **2. Test Scenarios**

```go
func TestCommonScenarios(t *testing.T) {
    scenarios := []struct {
        name     string
        setup    func() (*TestSystem, core.Event, *core.State)
        validate func(*testing.T, *core.AgentResult, error)
    }{
        {
            name: "successful_analysis",
            setup: func() (*TestSystem, core.Event, *core.State) {
                system := NewTestSystemBuilder().
                    WithAgent("analyzer", NewMockAnalyzer("success")).
                    Build()
                
                event := NewEventBuilder("analyze").
                    WithData("content", "test data").
                    Build()
                
                state := NewStateBuilder().
                    WithSession("test-session").
                    Build()
                
                return system, event, state
            },
            validate: func(t *testing.T, result *core.AgentResult, err error) {
                assert.NoError(t, err)
                assert.Equal(t, "completed", result.Data["status"])
            },
        },
        {
            name: "timeout_handling",
            setup: func() (*TestSystem, core.Event, *core.State) {
                system := NewTestSystemBuilder().
                    WithAgent("slow_agent", NewSlowMockAgent(10*time.Second)).
                    Build()
                
                event := NewEventBuilder("process").
                    WithData("timeout", "5s").
                    Build()
                
                state := NewStateBuilder().
                    WithSession("timeout-test").
                    Build()
                
                return system, event, state
            },
            validate: func(t *testing.T, result *core.AgentResult, err error) {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), "timeout")
            },
        },
    }
    
    for _, scenario := range scenarios {
        t.Run(scenario.name, func(t *testing.T) {
            system, event, state := scenario.setup()
            defer system.Cleanup()
            
            result, err := system.ProcessSingleEvent(context.Background(), event, state)
            scenario.validate(t, result, err)
        })
    }
}
```

## 🎯 Best Practices

### **1. Test Organization**

```go
// Package structure for tests
// agents/
//   analyzer_test.go          - Unit tests for analyzer agent
//   validator_test.go         - Unit tests for validator agent
//   integration_test.go       - Integration tests between agents
// orchestration/
//   sequential_test.go        - Sequential orchestration tests
//   collaborative_test.go     - Collaborative orchestration tests
// system/
//   e2e_test.go              - End-to-end system tests
//   performance_test.go       - Performance and load tests
// testutils/
//   builders.go              - Test data builders
//   mocks.go                 - Mock implementations
//   fixtures.go              - Test fixtures and helpers
```

### **2. Test Configuration**

```go
// Test configuration management
type TestConfig struct {
    DatabaseURL    string
    RedisURL       string
    MockServices   bool
    LogLevel       string
    TestTimeout    time.Duration
}

func LoadTestConfig() *TestConfig {
    return &TestConfig{
        DatabaseURL:  getEnvOrDefault("TEST_DATABASE_URL", "postgres://localhost/agentflow_test"),
        RedisURL:     getEnvOrDefault("TEST_REDIS_URL", "redis://localhost:6379/1"),
        MockServices: getEnvOrDefault("MOCK_SERVICES", "true") == "true",
        LogLevel:     getEnvOrDefault("TEST_LOG_LEVEL", "error"),
        TestTimeout:  30 * time.Second,
    }
}
```

### **3. Continuous Integration**

```yaml
# .github/workflows/test.yml
name: Test Suite

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      
      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: pgvector/pgvector:pg15
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: agentflow_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      
      - name: Run integration tests
        run: go test -v -tags=integration ./...
        env:
          TEST_DATABASE_URL: postgres://postgres:test@localhost:5432/agentflow_test
          TEST_REDIS_URL: redis://localhost:6379/1

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      
      - name: Start test environment
        run: docker-compose -f docker-compose.test.yml up -d
      
      - name: Wait for services
        run: ./scripts/wait-for-services.sh
      
      - name: Run E2E tests
        run: go test -v -tags=e2e ./...
      
      - name: Cleanup
        run: docker-compose -f docker-compose.test.yml down
```

Testing multi-agent systems requires a comprehensive approach that covers unit, integration, and end-to-end scenarios. By following these strategies and patterns, you can build reliable and maintainable agent systems with confidence.

## 🚀 Next Steps

- **[Circuit Breaker Patterns](circuit-breaker-patterns.md)** - Add fault tolerance to your tested systems
- **[Load Balancing and Scaling](load-balancing-scaling.md)** - Test scaled deployments
- **[Production Monitoring](../debugging/performance-monitoring.md)** - Monitor your tested systems in production
- **[Debugging Guide](../debugging/debugging-multi-agent-systems.md)** - Debug issues found during testing