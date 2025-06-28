# Performance Tuning Guide

This guide covers optimization techniques, benchmarking, and performance best practices for AgentFlow applications. Learn how to build high-throughput, low-latency agent systems.

## 🎯 Performance Overview

AgentFlow is designed for high performance with:
- **Go's Concurrency**: Native goroutines and channels for parallel processing
- **Event-Driven Architecture**: Non-blocking event processing
- **Connection Pooling**: Efficient resource utilization
- **Streaming Support**: Low-memory processing of large datasets
- **Intelligent Caching**: Reduced redundant computations

## 📊 Benchmarking Basics

### **Built-in Benchmarking**

AgentFlow includes benchmarking tools for measuring performance:

```bash
# Run standard benchmarks
agentcli benchmark --duration 60s --concurrent-users 10

# Benchmark specific agents
agentcli benchmark --agent myagent --requests 1000

# Memory profiling
agentcli benchmark --profile memory --output profile.mem

# CPU profiling
agentcli benchmark --profile cpu --output profile.cpu
```

### **Custom Benchmark Setup**

```go
func BenchmarkAgentExecution(b *testing.B) {
    // Setup
    config := &core.Config{
        LLM: core.LLMConfig{
            Provider: "azure",
            Azure: core.AzureConfig{
                Endpoint:   os.Getenv("AZURE_OPENAI_ENDPOINT"),
                APIKey:     os.Getenv("AZURE_OPENAI_API_KEY"),
                Deployment: "gpt-4o",
            },
        },
    }
    
    runner, err := core.NewRunner(config)
    require.NoError(b, err)
    
    agent := &TestAgent{}
    runner.RegisterAgent("test", agent)
    
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Create test event
    event := core.NewEvent("test_query", map[string]interface{}{
        "query": "What is the capital of France?",
    })
    
    // Reset timer before benchmarking
    b.ResetTimer()
    
    // Run benchmark
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            err := runner.Emit(event)
            if err != nil {
                b.Error(err)
            }
        }
    })
}
```

### **Performance Metrics Collection**

```go
type PerformanceMetrics struct {
    RequestsPerSecond    float64
    AverageLatency      time.Duration
    P95Latency          time.Duration
    P99Latency          time.Duration
    ErrorRate           float64
    MemoryUsage         uint64
    GCPauses            []time.Duration
    ActiveGoroutines    int
}

func CollectMetrics(duration time.Duration) *PerformanceMetrics {
    var metrics PerformanceMetrics
    
    // Collect runtime stats
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    metrics.MemoryUsage = m.Alloc
    metrics.ActiveGoroutines = runtime.NumGoroutine()
    
    // Collect latency percentiles
    latencies := collectLatencies(duration)
    sort.Slice(latencies, func(i, j int) bool {
        return latencies[i] < latencies[j]
    })
    
    if len(latencies) > 0 {
        metrics.P95Latency = latencies[int(float64(len(latencies))*0.95)]
        metrics.P99Latency = latencies[int(float64(len(latencies))*0.99)]
    }
    
    return &metrics
}
```

## ⚡ Agent Performance Optimization

### **1. Efficient State Management**

Minimize state copying and mutations:

```go
type OptimizedAgent struct {
    cache  *sync.Map // Thread-safe cache
    config AgentConfig
}

func (a *OptimizedAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Use read-only state access when possible
    query := state.GetString("query")
    
    // Check cache before expensive operations
    if cached, ok := a.cache.Load(query); ok {
        return cached.(core.AgentResult), nil
    }
    
    // Only clone state when necessary
    workingState := state.CloneIfNeeded()
    
    result := a.processQuery(ctx, query, workingState)
    
    // Cache successful results
    if result.Success {
        a.cache.Store(query, result)
    }
    
    return result, nil
}

// Implement efficient state cloning
func (s *State) CloneIfNeeded() *State {
    if s.IsReadOnly() {
        return s // No need to clone read-only state
    }
    return s.Clone()
}
```

### **2. Parallel Processing Patterns**

Leverage goroutines for concurrent operations:

```go
func (a *ParallelAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    query := event.GetData()["query"].(string)
    
    // Parallel tool execution
    toolCalls := a.identifyToolCalls(query)
    results := make(chan ToolResult, len(toolCalls))
    
    // Launch goroutines for each tool call
    var wg sync.WaitGroup
    for _, toolCall := range toolCalls {
        wg.Add(1)
        go func(tc ToolCall) {
            defer wg.Done()
            result := a.executeToolCall(ctx, tc)
            results <- result
        }(toolCall)
    }
    
    // Close channel when all goroutines complete
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    var toolResults []ToolResult
    for result := range results {
        toolResults = append(toolResults, result)
    }
    
    return a.synthesizeResults(toolResults), nil
}
```

### **3. Memory Pool Usage**

Reduce garbage collection pressure:

```go
type Agent struct {
    bufferPool sync.Pool
    eventPool  sync.Pool
}

func NewOptimizedAgent() *Agent {
    return &Agent{
        bufferPool: sync.Pool{
            New: func() interface{} {
                return make([]byte, 0, 4096) // Pre-allocate 4KB
            },
        },
        eventPool: sync.Pool{
            New: func() interface{} {
                return &ProcessingEvent{}
            },
        },
    }
}

func (a *Agent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get buffer from pool
    buffer := a.bufferPool.Get().([]byte)
    defer a.bufferPool.Put(buffer[:0]) // Reset and return to pool
    
    // Get event object from pool
    procEvent := a.eventPool.Get().(*ProcessingEvent)
    defer func() {
        procEvent.Reset()
        a.eventPool.Put(procEvent)
    }()
    
    // Use pooled objects for processing
    result := a.processWithBuffer(ctx, event, state, buffer, procEvent)
    return result, nil
}
```

## 🚀 LLM Provider Optimization

### **1. Connection Pooling**

Configure optimal connection pools:

```toml
[llm.azure]
max_connections = 50
min_connections = 5
connection_timeout = "10s"
idle_timeout = "300s"
max_connection_lifetime = "3600s"

[llm.openai]
max_connections = 30
request_timeout = "30s"
retry_max_attempts = 3
retry_initial_interval = "1s"
```

```go
type OptimizedLLMClient struct {
    client *http.Client
    pool   *ConnectionPool
}

func NewOptimizedLLMClient(config *LLMConfig) *OptimizedLLMClient {
    transport := &http.Transport{
        MaxIdleConns:        config.MaxConnections,
        MaxIdleConnsPerHost: config.MaxConnectionsPerHost,
        IdleConnTimeout:     config.IdleTimeout,
        DisableKeepAlives:   false,
        MaxConnsPerHost:     config.MaxConnectionsPerHost,
    }
    
    client := &http.Client{
        Transport: transport,
        Timeout:   config.RequestTimeout,
    }
    
    return &OptimizedLLMClient{
        client: client,
        pool:   NewConnectionPool(config),
    }
}
```

### **2. Request Batching**

Batch multiple requests when supported:

```go
type BatchingLLMClient struct {
    client    LLMClient
    batcher   *RequestBatcher
    maxBatch  int
    batchTime time.Duration
}

func (c *BatchingLLMClient) ProcessRequests(requests []*LLMRequest) ([]*LLMResponse, error) {
    if len(requests) == 1 {
        // Single request - process immediately
        response, err := c.client.SendRequest(requests[0])
        return []*LLMResponse{response}, err
    }
    
    // Batch multiple requests
    batches := c.createBatches(requests, c.maxBatch)
    responses := make([]*LLMResponse, 0, len(requests))
    
    for _, batch := range batches {
        batchResponses, err := c.client.SendBatchRequest(batch)
        if err != nil {
            return nil, err
        }
        responses = append(responses, batchResponses...)
    }
    
    return responses, nil
}
```

### **3. Smart Caching**

Implement intelligent LLM response caching:

```go
type CachedLLMClient struct {
    client LLMClient
    cache  Cache
    hasher ContentHasher
}

func (c *CachedLLMClient) SendRequest(req *LLMRequest) (*LLMResponse, error) {
    // Generate cache key from request content
    key := c.hasher.Hash(req)
    
    // Check cache first
    if cached, found := c.cache.Get(key); found {
        return cached.(*LLMResponse), nil
    }
    
    // Send request to provider
    response, err := c.client.SendRequest(req)
    if err != nil {
        return nil, err
    }
    
    // Cache successful responses
    if response.IsSuccessful() {
        c.cache.Set(key, response, c.getTTL(req))
    }
    
    return response, nil
}

func (c *CachedLLMClient) getTTL(req *LLMRequest) time.Duration {
    // Dynamic TTL based on request type
    if req.IsFactual() {
        return 24 * time.Hour // Cache factual queries longer
    }
    if req.IsCreative() {
        return 5 * time.Minute // Cache creative requests briefly
    }
    return time.Hour // Default TTL
}
```

## 🔧 MCP Tool Optimization

### **1. Tool Connection Management**

Optimize MCP server connections:

```go
type OptimizedMCPManager struct {
    connections map[string]*MCPConnectionPool
    healthCheck *HealthChecker
    loadBalancer *LoadBalancer
}

type MCPConnectionPool struct {
    servers []*MCPConnection
    current int
    mutex   sync.RWMutex
}

func (p *MCPConnectionPool) GetConnection() *MCPConnection {
    p.mutex.RLock()
    defer p.mutex.RUnlock()
    
    // Round-robin load balancing
    conn := p.servers[p.current]
    p.current = (p.current + 1) % len(p.servers)
    
    return conn
}

func (m *OptimizedMCPManager) ExecuteTool(ctx context.Context, tool string, params map[string]interface{}) (*ToolResult, error) {
    pool := m.connections[tool]
    if pool == nil {
        return nil, fmt.Errorf("no connections available for tool %s", tool)
    }
    
    // Get healthy connection
    conn := pool.GetConnection()
    if !m.healthCheck.IsHealthy(conn) {
        conn = m.loadBalancer.GetHealthyConnection(pool)
    }
    
    return conn.ExecuteTool(ctx, tool, params)
}
```

### **2. Tool Result Streaming**

Handle large tool results efficiently:

```go
func (t *StreamingTool) ExecuteStreaming(ctx context.Context, params map[string]interface{}) (<-chan *ToolResult, error) {
    resultChan := make(chan *ToolResult, 100) // Buffered channel
    
    go func() {
        defer close(resultChan)
        
        // Use streaming database query
        rows, err := t.db.QueryContext(ctx, t.buildQuery(params))
        if err != nil {
            resultChan <- &ToolResult{Error: err.Error()}
            return
        }
        defer rows.Close()
        
        batch := make([]map[string]interface{}, 0, 100)
        
        for rows.Next() {
            row := make(map[string]interface{})
            err := rows.MapScan(row)
            if err != nil {
                resultChan <- &ToolResult{Error: err.Error()}
                return
            }
            
            batch = append(batch, row)
            
            // Send batch when full
            if len(batch) >= 100 {
                resultChan <- &ToolResult{
                    Success: true,
                    Data:    batch,
                }
                batch = batch[:0] // Reset slice
            }
        }
        
        // Send remaining items
        if len(batch) > 0 {
            resultChan <- &ToolResult{
                Success: true,
                Data:    batch,
            }
        }
    }()
    
    return resultChan, nil
}
```

## 📈 Concurrency Optimization

### **1. Worker Pool Pattern**

Limit concurrent operations with worker pools:

```go
type WorkerPool struct {
    workers  int
    jobQueue chan Job
    wg       sync.WaitGroup
    ctx      context.Context
    cancel   context.CancelFunc
}

type Job struct {
    ID     string
    Event  core.Event
    State  core.State
    Result chan<- JobResult
}

func NewWorkerPool(workers int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    return &WorkerPool{
        workers:  workers,
        jobQueue: make(chan Job, workers*2),
        ctx:      ctx,
        cancel:   cancel,
    }
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(i)
    }
}

func (p *WorkerPool) worker(id int) {
    defer p.wg.Done()
    
    for {
        select {
        case job := <-p.jobQueue:
            result := p.processJob(job)
            job.Result <- result
        case <-p.ctx.Done():
            return
        }
    }
}

func (p *WorkerPool) Submit(job Job) {
    select {
    case p.jobQueue <- job:
        // Job queued successfully
    case <-p.ctx.Done():
        job.Result <- JobResult{Error: "Worker pool is shutting down"}
    }
}
```

### **2. Circuit Breaker for High Load**

Protect against cascade failures:

```go
type CircuitBreaker struct {
    state         CircuitState
    failureCount  int64
    successCount  int64
    lastFailTime  time.Time
    timeout       time.Duration
    maxFailures   int
    mutex         sync.RWMutex
}

func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
    cb.mutex.Lock()
    state := cb.state
    cb.mutex.Unlock()
    
    switch state {
    case CircuitOpen:
        if time.Since(cb.lastFailTime) > cb.timeout {
            cb.setState(CircuitHalfOpen)
        } else {
            return nil, ErrCircuitOpen
        }
    case CircuitHalfOpen:
        // Allow limited requests through
    }
    
    result, err := fn()
    
    if err != nil {
        cb.onFailure()
        return nil, err
    }
    
    cb.onSuccess()
    return result, nil
}
```

## 🔍 Profiling and Monitoring

### **1. Runtime Profiling**

Enable Go's built-in profiling:

```go
import (
    _ "net/http/pprof"
    "net/http"
    "log"
)

func main() {
    // Start pprof server
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Your AgentFlow application
    runAgentFlow()
}

// Profile CPU usage
// go tool pprof http://localhost:6060/debug/pprof/profile

// Profile memory usage
// go tool pprof http://localhost:6060/debug/pprof/heap

// Profile goroutines
// go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### **2. Custom Metrics**

Track application-specific metrics:

```go
type PerformanceTracker struct {
    requestDuration prometheus.HistogramVec
    requestCount    prometheus.CounterVec
    activeRequests  prometheus.GaugeVec
    errorRate       prometheus.CounterVec
}

func NewPerformanceTracker() *PerformanceTracker {
    return &PerformanceTracker{
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "agent_request_duration_seconds",
                Help: "Agent request duration in seconds",
                Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
            },
            []string{"agent_name", "status"},
        ),
        requestCount: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "agent_requests_total",
                Help: "Total agent requests",
            },
            []string{"agent_name", "status"},
        ),
    }
}

func (pt *PerformanceTracker) TrackRequest(agentName string, duration time.Duration, success bool) {
    status := "success"
    if !success {
        status = "error"
    }
    
    pt.requestDuration.WithLabelValues(agentName, status).Observe(duration.Seconds())
    pt.requestCount.WithLabelValues(agentName, status).Inc()
}
```

## 🎛️ Configuration Tuning

### **1. Runtime Optimization**

Tune Go runtime parameters:

```go
import "runtime"

func init() {
    // Set GOMAXPROCS to number of CPU cores
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Tune GC target percentage
    debug.SetGCPercent(100) // Default is 100
    
    // Set memory limit (Go 1.19+)
    debug.SetMemoryLimit(8 << 30) // 8GB limit
}
```

### **2. Application Configuration**

Optimize AgentFlow settings:

```toml
[runner]
max_concurrent_events = 100
event_buffer_size = 1000
worker_pool_size = 50
event_timeout = "30s"

[memory]
state_cache_size = 10000
state_cache_ttl = "1h"
enable_state_compression = true

[performance]
enable_request_batching = true
batch_size = 10
batch_timeout = "100ms"
enable_connection_pooling = true
pool_size = 50

[monitoring]
enable_metrics = true
metrics_interval = "10s"
enable_profiling = true
profile_port = 6060
```

## 📊 Performance Benchmarks

### **Typical Performance Characteristics**

| Metric | Single Agent | Multi-Agent | Streaming |
|--------|-------------|-------------|-----------|
| **Throughput** | 1000 req/s | 500 req/s | 10k items/s |
| **Latency (P95)** | 50ms | 200ms | 5ms |
| **Memory Usage** | 50MB | 200MB | 30MB |
| **CPU Usage** | 20% | 60% | 15% |

### **Optimization Targets**

- **Latency**: < 100ms for simple queries, < 500ms for complex multi-tool operations
- **Throughput**: > 500 requests/second per instance
- **Memory**: < 512MB per instance under normal load
- **Error Rate**: < 0.1% for well-configured systems

### **Load Testing Example**

```bash
# Install k6 load testing tool
# https://k6.io/docs/getting-started/installation/

# Basic load test
agentcli loadtest --script loadtest.js --duration 5m --vus 100

# Stress test
agentcli loadtest --script stress.js --stages "5m:100,10m:500,5m:100"

# Spike test
agentcli loadtest --script spike.js --stages "2m:10,1m:1000,2m:10"
```

This performance guide provides the tools and techniques needed to build high-performance AgentFlow applications that scale efficiently under load.
