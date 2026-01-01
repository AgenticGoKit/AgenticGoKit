# Performance Optimization

Optimize v1beta agents for maximum performance, efficiency, and scalability.

---

## Quick Wins

### Streaming vs Non-Streaming

```go
// Non-streaming: Wait for full response
result, _ := agent.Run(ctx, query)
fmt.Println(result.Content)

// Streaming: Show results immediately (70% less memory)
stream, _ := agent.RunStream(ctx, query)
for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
}
```

### Parallel Workflows

```go
// Sequential: 6 seconds (2s + 2s + 2s)
workflow, _ := v1beta.NewSequentialWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "s1", Agent: agent1})
workflow.AddStep(v1beta.WorkflowStep{Name: "s2", Agent: agent2})
workflow.AddStep(v1beta.WorkflowStep{Name: "s3", Agent: agent3})

// Parallel: 2 seconds (3× faster)
workflow, _ := v1beta.NewParallelWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "s1", Agent: agent1})
workflow.AddStep(v1beta.WorkflowStep{Name: "s2", Agent: agent2})
workflow.AddStep(v1beta.WorkflowStep{Name: "s3", Agent: agent3})
```

### Token Limits

```go
// Shorter responses = faster + cheaper
result, _ := agent.RunWithOptions(ctx, query, &v1beta.RunOptions{
    MaxTokens: 100,  // Brief: ~0.5s
})

result, _ := agent.RunWithOptions(ctx, query, &v1beta.RunOptions{
    MaxTokens: 2000, // Detailed: ~8s
})
```

---

## Streaming Optimization

---

## Streaming Optimization

### Buffer Sizing

```go
// Real-time UI (low latency, high CPU)
stream, _ := agent.RunStream(ctx, query, v1beta.WithBufferSize(50))

// Interactive (balanced, recommended)
stream, _ := agent.RunStream(ctx, query, v1beta.WithBufferSize(100))

// Batch processing (high throughput, low CPU)
stream, _ := agent.RunStream(ctx, query, v1beta.WithBufferSize(500))
```

Guidelines: Real-time 25-50 | Interactive 50-100 | Batch 200-500

### Flush Intervals

```go
// Immediate updates (more CPU)
stream, _ := agent.RunStream(ctx, query, v1beta.WithFlushInterval(10*time.Millisecond))

// Balanced (recommended)
stream, _ := agent.RunStream(ctx, query, v1beta.WithFlushInterval(100*time.Millisecond))

// Batched (less CPU)
stream, _ := agent.RunStream(ctx, query, v1beta.WithFlushInterval(500*time.Millisecond))
```

### Text-Only Mode

Skip unnecessary metadata:

```go
stream, _ := agent.RunStream(ctx, query, v1beta.WithTextOnly())
```

Saves ~30% chunk processing overhead.

---

## Memory Management

### Token Limits

Reduce memory footprint:

```go
result, _ := agent.RunWithOptions(ctx, input, &v1beta.RunOptions{
    MaxTokens:    1000,  // Limit output size
    HistoryLimit: 10,    // Keep last 10 messages
})
```

Saves up to 80% for long conversations.

### Streaming vs Buffering

```go
// High memory: Full response in memory
result, _ := agent.Run(ctx, input)
fullText := result.Content

// Low memory: Process chunks as they arrive (70% reduction)
stream, _ := agent.RunStream(ctx, input)
for chunk := range stream.Chunks() {
    process(chunk.Delta)
}
```

### Agent Cleanup

```go
agent, _ := v1beta.NewBuilder("agent").WithLLM("openai", "gpt-4").Build()
defer agent.Cleanup(context.Background())

result, _ := agent.Run(ctx, query)
```

---

## Concurrent Execution

### Parallel Workflows

```go
// Sequential: 3 seconds (1s + 1s + 1s)
workflow, _ := v1beta.NewSequentialWorkflow(config)

// Parallel: 1 second (3× faster)
workflow, _ := v1beta.NewParallelWorkflow(config)

// DAG: 2 seconds (steps 2,3 parallel after step 1)
workflow, _ := v1beta.NewDAGWorkflow(config)
```

### Concurrent Agents

```go
var wg sync.WaitGroup
results := make(chan *v1beta.Result, len(queries))

for _, query := range queries {
    wg.Add(1)
    go func(q string) {
        defer wg.Done()
        agent, _ := v1beta.NewBuilder("agent").
            WithLLM("openai", "gpt-4").
            Build()
        defer agent.Cleanup(context.Background())
        
        result, _ := agent.Run(context.Background(), q)
        results <- result
    }(query)
}

wg.Wait()
close(results)
```

### Worker Pools

```go
type WorkerPool struct {
    workers int
    jobs    chan string
    agent   v1beta.Agent
}

func (p *WorkerPool) worker() {
    for job := range p.jobs {
        p.agent.Run(context.Background(), job)
    }
}

// Usage
pool := &WorkerPool{workers: 10, jobs: make(chan string)}
for i := 0; i < pool.workers; i++ {
    go pool.worker()
}

for _, query := range queries {
    pool.jobs <- query
}
```

---

## LLM Optimization

### Model Selection

```go
// Fast, cheap (simple tasks)
v1beta.LLMConfig{Provider: "openai", Model: "gpt-3.5-turbo"}

// Balanced (most cases) - RECOMMENDED
v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"}

// Powerful (complex reasoning)
v1beta.LLMConfig{Provider: "openai", Model: "gpt-4-turbo"}
```

Cost: gpt-3.5-turbo (1×) | gpt-4 (10×) | gpt-4-turbo (20×)

### Temperature Settings

```go
// Deterministic (faster, consistent) - 15% faster
Temperature: 0.0

// Creative (standard performance)
Temperature: 0.7
```

### Batch Processing

```go
// Inefficient: 3 separate calls
for _, q := range queries {
    agent.Run(ctx, q)
}

// Efficient: Process together
batch := "Process these:\n" + strings.Join(queries, "\n---\n")
agent.Run(ctx, batch)
```

---

## Tool Optimization

### Timeout Configuration

```go
agent, _ := v1beta.NewBuilder("agent").
    WithTools(
        v1beta.WithMCP(servers...),
        v1beta.WithToolTimeout(30 * time.Second),
    ).
    Build()
```

Guidelines: Fast tools 5s | Standard 30s | Slow 60s+

### Lazy Loading

Load tools on-demand:

```go
type ToolRegistry struct {
    loaders map[string]func() Tool
    cache   map[string]Tool
    mu      sync.RWMutex
}

func (r *ToolRegistry) GetTool(name string) Tool {
    r.mu.RLock()
    if tool, ok := r.cache[name]; ok {
        r.mu.RUnlock()
        return tool
    }
    r.mu.RUnlock()
    
    r.mu.Lock()
    defer r.mu.Unlock()
    tool := r.loaders[name]()
    r.cache[name] = tool
    return tool
}
```

---

## Workflow Optimization

### DAG Workflows

Maximize parallelism:

```go
// Sequential: 6 seconds
workflow, _ := v1beta.NewSequentialWorkflow(config)

// DAG: 4 seconds (steps 2,3 parallel after step 1)
workflow, _ := v1beta.NewDAGWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "a", Agent: agentA})
workflow.AddStep(v1beta.WorkflowStep{Name: "b", Agent: agentB, Dependencies: []string{"a"}})
workflow.AddStep(v1beta.WorkflowStep{Name: "c", Agent: agentC, Dependencies: []string{"a"}})
```

### Early Exit

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    result := quickSearch(input)
    if result != "" {
        return result, nil // Skip remaining steps
    }
    return "", nil
}
```

---

## Benchmarking

---

## Benchmarking

### Basic Benchmarks

```go
func BenchmarkAgentRun(b *testing.B) {
    agent, _ := v1beta.NewBuilder("agent").WithLLM("openai", "gpt-4").Build()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        agent.Run(context.Background(), "Test")
    }
}

func BenchmarkStreaming(b *testing.B) {
    agent, _ := v1beta.NewBuilder("agent").WithLLM("openai", "gpt-4").Build()
    
    b.Run("NonStreaming", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            agent.Run(context.Background(), "Query")
        }
    })
    
    b.Run("Streaming", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            stream, _ := agent.RunStream(context.Background(), "Query")
            for range stream.Chunks() {}
            stream.Wait()
        }
    })
}
```

Run with: `go test -bench=. -benchmem ./...`

### Memory Profiling

```go
func TestMemory(t *testing.T) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    before := m.Alloc
    
    agent, _ := v1beta.NewBuilder("agent").WithLLM("openai", "gpt-4").Build()
    
    for i := 0; i < 100; i++ {
        agent.Run(context.Background(), "Test")
    }
    
    runtime.ReadMemStats(&m)
    after := m.Alloc
    
    t.Logf("Memory used: %d KB", (after-before)/1024)
}
```

### Load Testing

```go
func LoadTest() {
    agent, _ := v1beta.NewBuilder("agent").WithLLM("openai", "gpt-4").Build()
    
    start := time.Now()
    concurrent := 100
    requestsPerWorker := 10
    
    var wg sync.WaitGroup
    for i := 0; i < concurrent; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < requestsPerWorker; j++ {
                agent.Run(context.Background(), "Test")
            }
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalRequests := concurrent * requestsPerWorker
    rps := float64(totalRequests) / duration.Seconds()
    
    fmt.Printf("Total: %d, RPS: %.2f\n", totalRequests, rps)
}
```

---

## Performance Expectations

| Operation | Latency | Throughput | Memory |
|-----------|---------|------------|--------|
| Simple query | 500-1000ms | 100-200 req/s | 10-20 MB |
| Streaming | 50ms TTFB | 1000+ chunks/s | 5-10 MB |
| Tool call | 100-500ms | 200-500 ops/s | 5 MB |
| Sequential workflow (3 steps) | 1.5-3s | 30-60 flows/s | 20-40 MB |
| Parallel workflow (3 steps) | 0.5-1s | 100-200 flows/s | 30-50 MB |

---

## Checklist

- [ ] Use streaming for long responses
- [ ] Configure buffer sizes for your use case
- [ ] Enable text-only mode when possible
- [ ] Set appropriate token limits
- [ ] Use parallel workflows for independent tasks
- [ ] Choose right models for tasks (gpt-3.5 vs gpt-4)
- [ ] Implement agent cleanup
- [ ] Use concurrent execution
- [ ] Cache tool results
- [ ] Set tool timeouts
- [ ] Profile memory usage
- [ ] Benchmark critical paths
- [ ] Monitor error rates

---

**Next:** [Streaming](./streaming.md) → [Troubleshooting](./troubleshooting.md)
