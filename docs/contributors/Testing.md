# Testing Strategy

This document outlines the comprehensive testing approach for AgenticGoKit, covering unit tests, integration tests, benchmarks, and quality assurance practices.

## 🎯 Testing Philosophy

AgenticGoKit follows a multi-layered testing strategy:

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test component interactions and workflows
3. **End-to-End Tests**: Test complete user scenarios
4. **Performance Tests**: Validate performance characteristics
5. **Chaos Tests**: Test resilience under failure conditions

## 🏗️ Test Organization

### Directory Structure

```
agenticgokit/
├── core/                           # Public API tests
│   ├── agent_test.go
│   ├── runner_test.go
│   ├── mcp_test.go
│   └── *_test.go
├── internal/                       # Implementation tests
│   ├── agents/
│   │   └── *_test.go
│   ├── mcp/
│   │   └── *_test.go
│   └── */
│       └── *_test.go
├── integration/                    # Integration tests
│   ├── mcp_integration_test.go
│   ├── workflow_integration_test.go
│   └── *_integration_test.go
├── benchmarks/                     # Performance benchmarks
│   ├── agent_benchmark_test.go
│   ├── mcp_benchmark_test.go
│   └── *_benchmark_test.go
└── testdata/                       # Test fixtures and data
    ├── configs/
    ├── fixtures/
    └── mocks/
```

### Test File Naming Conventions

| Pattern | Purpose | Example |
|---------|---------|---------|
| `*_test.go` | Unit tests | `agent_test.go` |
| `*_integration_test.go` | Integration tests | `mcp_integration_test.go` |
| `*_benchmark_test.go` | Benchmarks | `runner_benchmark_test.go` |
| `mock_*.go` | Mock implementations | `mock_llm_provider.go` |
| `test_*.go` | Test utilities | `test_helpers.go` |

## 🧪 Unit Testing

### Test Structure

Follow the AAA (Arrange, Act, Assert) pattern:

```go
func TestAgentRun(t *testing.T) {
    // Arrange
    agent := NewTestAgent("test-agent")
    event := core.NewEvent("test", map[string]interface{}{
        "query": "Hello world",
    })
    state := core.NewState()
    
    // Act
    result, err := agent.Run(context.Background(), event, state)
    
    // Assert
    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "Hello world", result.Data["processed_query"])
}
```

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestLLMProviderComplete(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "Simple query",
            input:    "What is 2+2?",
            expected: "4",
            wantErr:  false,
        },
        {
            name:     "Empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
        {
            name:     "Complex query",
            input:    "Explain quantum computing",
            expected: "Quantum computing is...",
            wantErr:  false,
        },
    }
    
    provider := NewMockLLMProvider()
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := provider.Complete(context.Background(), tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Contains(t, result, tt.expected)
        })
    }
}
```

### Mock Usage

Create focused mocks for external dependencies:

```go
type MockLLMProvider struct {
    responses map[string]string
    errors    map[string]error
    callCount int
    mu        sync.Mutex
}

func NewMockLLMProvider() *MockLLMProvider {
    return &MockLLMProvider{
        responses: make(map[string]string),
        errors:    make(map[string]error),
    }
}

func (m *MockLLMProvider) SetResponse(input, output string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.responses[input] = output
}

func (m *MockLLMProvider) SetError(input string, err error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.errors[input] = err
}

func (m *MockLLMProvider) Complete(ctx context.Context, prompt string) (string, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.callCount++
    
    if err, exists := m.errors[prompt]; exists {
        return "", err
    }
    
    if response, exists := m.responses[prompt]; exists {
        return response, nil
    }
    
    return "default response", nil
}

func (m *MockLLMProvider) GetCallCount() int {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.callCount
}
```

### Test Utilities

Create reusable test utilities:

```go
// test_helpers.go
package core

import (
    "context"
    "testing"
    "time"
)

// TestConfig returns a configuration suitable for testing
func TestConfig() *Config {
    return &Config{
        LLM: LLMConfig{
            Provider: "mock",
        },
        MCP: MCPConfig{
            Enabled: false, // Disable for unit tests
        },
        Runner: RunnerConfig{
            MaxConcurrentEvents: 1,
            EventTimeout:        time.Second * 5,
        },
    }
}

// WithTimeout creates a context with timeout for tests
func WithTimeout(t *testing.T, timeout time.Duration) context.Context {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    t.Cleanup(cancel)
    return ctx
}

// AssertEventually retries assertion until it passes or times out
func AssertEventually(t *testing.T, assertion func() bool, timeout time.Duration, interval time.Duration) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        if assertion() {
            return
        }
        
        select {
        case <-ctx.Done():
            t.Fatal("Assertion timed out")
        case <-ticker.C:
            continue
        }
    }
}
```

## 🔗 Integration Testing

### MCP Integration Tests

Test MCP server interactions:

```go
func TestMCPIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Start test MCP server
    server := startTestMCPServer(t)
    defer server.Stop()
    
    // Configure AgenticGoKit with test server
    config := &core.Config{
        MCP: core.MCPConfig{
            Enabled: true,
            Servers: []core.MCPServerConfig{
                {
                    Name:    "test-server",
                    Address: server.Address(),
                },
            },
        },
    }
    
    runner, err := core.NewRunner(config)
    require.NoError(t, err)
    defer runner.Stop()
    
    // Test tool discovery
    tools, err := runner.GetMCPManager().ListTools(context.Background())
    require.NoError(t, err)
    assert.NotEmpty(t, tools)
    
    // Test tool execution
    result, err := runner.GetMCPManager().ExecuteTool(context.Background(), "test_tool", map[string]interface{}{
        "input": "test data",
    })
    require.NoError(t, err)
    assert.True(t, result.Success)
}
```

### Multi-Agent Workflow Tests

Test complex agent interactions:

```go
func TestMultiAgentWorkflow(t *testing.T) {
    // Setup agents
    searchAgent := &SearchAgent{}
    analysisAgent := &AnalysisAgent{}
    summaryAgent := &SummaryAgent{}
    
    // Create orchestrator
    orchestrator := core.NewOrchestrator(core.OrchestrationModeCollaborate)
    orchestrator.RegisterAgent("search", searchAgent)
    orchestrator.RegisterAgent("analysis", analysisAgent)
    orchestrator.RegisterAgent("summary", summaryAgent)
    
    // Define workflow
    workflow := &core.Workflow{
        Steps: []core.WorkflowStep{
            {AgentName: "search", Dependencies: []string{}},
            {AgentName: "analysis", Dependencies: []string{"search"}},
            {AgentName: "summary", Dependencies: []string{"analysis"}},
        },
    }
    
    // Execute workflow
    event := core.NewEvent("research", map[string]interface{}{
        "topic": "AI advancements in 2024",
    })
    
    result, err := orchestrator.ExecuteWorkflow(context.Background(), workflow, event)
    require.NoError(t, err)
    
    // Verify workflow execution
    assert.Contains(t, result.Data, "search_results")
    assert.Contains(t, result.Data, "analysis")
    assert.Contains(t, result.Data, "summary")
}
```

### Database Integration Tests

Test persistent storage:

```go
func TestStatePeristenceIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create session service with DB
    sessionService := memory.NewDatabaseSessionService(db)
    
    // Test session creation and retrieval
    session := core.NewSession("test-user", "test-session")
    session.GetState().Set("key", "value")
    
    err := sessionService.SaveSession(context.Background(), session)
    require.NoError(t, err)
    
    retrieved, err := sessionService.GetSession(context.Background(), "test-session")
    require.NoError(t, err)
    assert.Equal(t, "value", retrieved.GetState().GetString("key"))
}
```

## 🚀 Performance Testing

### Benchmarks

Create comprehensive benchmarks:

```go
func BenchmarkAgentExecution(b *testing.B) {
    agent := &TestAgent{}
    event := core.NewEvent("benchmark", map[string]interface{}{
        "query": "test query",
    })
    state := core.NewState()
    ctx := context.Background()
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := agent.Run(ctx, event, state)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkConcurrentAgentExecution(b *testing.B) {
    agent := &TestAgent{}
    event := core.NewEvent("benchmark", map[string]interface{}{
        "query": "test query",
    })
    state := core.NewState()
    
    b.ResetTimer()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := agent.Run(context.Background(), event, state.Clone())
            if err != nil {
                b.Error(err)
            }
        }
    })
}

func BenchmarkMemoryAllocation(b *testing.B) {
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        state := core.NewState()
        state.Set("key", "value")
        _ = state.Clone()
    }
}
```

### Load Testing

Use external tools for load testing:

```javascript
// k6 load test script
import http from 'k6/http';
import { check } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 20 },
        { duration: '1m', target: 20 },
        { duration: '20s', target: 0 },
    ],
};

export default function() {
    let response = http.post('http://localhost:8080/api/chat', JSON.stringify({
        query: 'Hello, how are you?',
        session_id: `session_${__VU}_${__ITER}`,
    }), {
        headers: { 'Content-Type': 'application/json' },
    });
    
    check(response, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
    });
}
```

## 🔥 Chaos Testing

### Failure Injection

Test system resilience:

```go
func TestChaosFailureRecovery(t *testing.T) {
    // Create chaos injector
    chaos := &ChaosInjector{
        FailureRate: 0.3, // 30% failure rate
        FailureTypes: []FailureType{
            NetworkTimeout,
            ServiceUnavailable,
            RateLimitExceeded,
        },
    }
    
    // Wrap agent with chaos injection
    agent := NewChaosAgent(&TestAgent{}, chaos)
    
    // Run multiple iterations
    successCount := 0
    totalRuns := 100
    
    for i := 0; i < totalRuns; i++ {
        result, err := agent.Run(context.Background(), testEvent, testState)
        if err == nil && result.Success {
            successCount++
        }
    }
    
    // Verify system remains functional despite failures
    successRate := float64(successCount) / float64(totalRuns)
    assert.Greater(t, successRate, 0.6, "System should maintain >60% success rate under chaos")
}
```

### Resource Exhaustion Tests

```go
func TestMemoryPressure(t *testing.T) {
    // Create memory pressure
    var memoryHog [][]byte
    defer func() {
        memoryHog = nil
        runtime.GC()
    }()
    
    // Allocate significant memory
    for i := 0; i < 100; i++ {
        memoryHog = append(memoryHog, make([]byte, 1024*1024)) // 1MB chunks
    }
    
    // Test agent behavior under memory pressure
    agent := &TestAgent{}
    result, err := agent.Run(context.Background(), testEvent, testState)
    
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

## 📊 Test Coverage

### Coverage Requirements

- **Minimum Overall Coverage**: 80%
- **Critical Path Coverage**: 95%
- **Public API Coverage**: 90%
- **Error Path Coverage**: 70%

### Measuring Coverage

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Check coverage percentage
go tool cover -func=coverage.out | grep total

# Fail if coverage below threshold
go test -coverprofile=coverage.out ./... && \
go tool cover -func=coverage.out | \
awk '/total:/ {print $3}' | \
sed 's/%//' | \
awk '{if($1 < 80) exit 1}'
```

### Coverage Analysis

```go
//go:build coverage
// +build coverage

package main

import (
    "encoding/json"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "testing"
)

func TestCoverageAnalysis(t *testing.T) {
    // Parse coverage profile
    coverage := parseCoverageProfile("coverage.out")
    
    // Analyze critical functions
    criticalFunctions := []string{
        "Agent.Run",
        "Runner.Emit",
        "MCPManager.ExecuteTool",
    }
    
    for _, fn := range criticalFunctions {
        if coverage[fn] < 95.0 {
            t.Errorf("Critical function %s has insufficient coverage: %.1f%%", fn, coverage[fn])
        }
    }
}
```

## 🚦 Continuous Integration

### GitHub Actions Workflow

```yaml
name: Test Suite
on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests
      run: go test -v -short ./...
    
    - name: Run integration tests
      run: go test -v -tags=integration ./integration/...
      env:
        AZURE_OPENAI_ENDPOINT: ${{ secrets.AZURE_OPENAI_ENDPOINT }}
        AZURE_OPENAI_API_KEY: ${{ secrets.AZURE_OPENAI_API_KEY }}
    
    - name: Run benchmarks
      run: go test -bench=. -benchmem ./benchmarks/...
    
    - name: Generate coverage
      run: go test -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
    
    - name: Check coverage threshold
      run: |
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$COVERAGE < 80" | bc -l) )); then
          echo "Coverage $COVERAGE% is below threshold 80%"
          exit 1
        fi
```

## 🛠️ Test Utilities and Helpers

### Test Server Setup

```go
// testserver.go
type TestServer struct {
    httpServer *httptest.Server
    mcpServer  *TestMCPServer
    cleanup    []func()
}

func NewTestServer(t *testing.T) *TestServer {
    ts := &TestServer{}
    
    // Setup HTTP server
    mux := http.NewServeMux()
    mux.HandleFunc("/health", ts.healthHandler)
    mux.HandleFunc("/api/chat", ts.chatHandler)
    
    ts.httpServer = httptest.NewServer(mux)
    ts.cleanup = append(ts.cleanup, ts.httpServer.Close)
    
    // Setup MCP server
    ts.mcpServer = NewTestMCPServer(t)
    ts.cleanup = append(ts.cleanup, ts.mcpServer.Stop)
    
    t.Cleanup(func() {
        for i := len(ts.cleanup) - 1; i >= 0; i-- {
            ts.cleanup[i]()
        }
    })
    
    return ts
}

func (ts *TestServer) URL() string {
    return ts.httpServer.URL
}
```

### Test Data Management

```go
// testdata.go
type TestDataManager struct {
    baseDir string
}

func NewTestDataManager(t *testing.T) *TestDataManager {
    return &TestDataManager{
        baseDir: filepath.Join("testdata", t.Name()),
    }
}

func (tdm *TestDataManager) LoadJSON(filename string, v interface{}) error {
    data, err := os.ReadFile(filepath.Join(tdm.baseDir, filename))
    if err != nil {
        return err
    }
    return json.Unmarshal(data, v)
}

func (tdm *TestDataManager) SaveJSON(filename string, v interface{}) error {
    data, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return err
    }
    
    dir := filepath.Dir(filepath.Join(tdm.baseDir, filename))
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    
    return os.WriteFile(filepath.Join(tdm.baseDir, filename), data, 0644)
}
```

## 📝 Testing Best Practices

### Do's

- ✅ Write tests before or alongside code (TDD/BDD)
- ✅ Use descriptive test names that explain behavior
- ✅ Test both happy path and error conditions
- ✅ Use table-driven tests for multiple scenarios
- ✅ Mock external dependencies appropriately
- ✅ Keep tests independent and idempotent
- ✅ Use proper setup and teardown
- ✅ Measure and maintain high test coverage

### Don'ts

- ❌ Test implementation details instead of behavior
- ❌ Write tests that depend on external services in unit tests
- ❌ Create tests that depend on execution order
- ❌ Use overly complex test setups
- ❌ Ignore test failures or flaky tests
- ❌ Write tests without assertions
- ❌ Mock everything (test real integrations when appropriate)
- ❌ Skip testing error conditions

This comprehensive testing strategy ensures AgenticGoKit maintains high quality, reliability, and performance across all components and use cases.
