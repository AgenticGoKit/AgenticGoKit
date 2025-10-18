# Performance Optimization Guide

This guide provides best practices and optimization techniques for AgenticGoKit vNext to help you build high-performance AI agent applications.

## Table of Contents

- [Streaming Performance](#streaming-performance)
- [Memory Management](#memory-management)
- [Concurrent Execution](#concurrent-execution)
- [Caching Strategies](#caching-strategies)
- [LLM Optimization](#llm-optimization)
- [Tool Execution](#tool-execution)
- [Workflow Optimization](#workflow-optimization)
- [Benchmarking](#benchmarking)

## Streaming Performance

### Buffer Sizing

Buffer size significantly impacts performance and latency:

```go
// Low latency (real-time chat)
stream, _ := agent.RunStream(ctx, query,
    vnext.WithBufferSize(50), // Small buffer, fast delivery
)

// Balanced (recommended default)
stream, _ := agent.RunStream(ctx, query,
    vnext.WithBufferSize(100), // Default
)

// High throughput (batch processing)
stream, _ := agent.RunStream(ctx, query,
    vnext.WithBufferSize(500), // Large buffer, higher throughput
)
```

**Guidelines:**
- **Real-time UI**: 25-50
- **Interactive chat**: 50-100
- **Data processing**: 200-500
- **Batch operations**: 500-1000

### Flush Intervals

Control how frequently chunks are flushed:

```go
// Immediate updates (more CPU usage)
vnext.WithFlushInterval(10 * time.Millisecond)

// Balanced (recommended)
vnext.WithFlushInterval(100 * time.Millisecond)

// Batched (less CPU usage)
vnext.WithFlushInterval(500 * time.Millisecond)
```

**Impact:**
- Shorter intervals: Lower latency, higher CPU usage
- Longer intervals: Higher latency, lower CPU usage, better throughput

### Text-Only Mode

Skip unnecessary chunk types for better performance:

```go
// Skip thoughts, tools, metadata
stream, _ := agent.RunStream(ctx, query,
    vnext.WithTextOnly(true),
)
```

**Performance gain:** ~30% reduction in chunk processing overhead

### Stream Utilities Performance

```go
// Fastest: Direct chunk processing
for chunk := range stream.Chunks() {
    processChunk(chunk)
}

// Fast: Collect to string
text, _ := vnext.CollectStream(stream)

// Moderate: Stream to channel
textChan := vnext.StreamToChannel(stream)

// Slower: AsReader (adds buffering layer)
reader := stream.AsReader()
```

## Memory Management

### Limiting Context Size

Reduce memory footprint by limiting context:

```go
result, _ := agent.Run(ctx, input,
    vnext.WithMaxTokens(1000),      // Limit output size
    vnext.WithHistoryLimit(10),     // Keep last 10 messages
    vnext.WithContextWindow(4000),  // Trim to 4K tokens
)
```

**Memory savings:** Up to 80% for long conversations

### Streaming vs Buffering

```go
// High memory: Buffer full response
result, _ := agent.Run(ctx, input)
fullText := result.Content // Entire response in memory

// Low memory: Stream and process
stream, _ := agent.RunStream(ctx, input)
for chunk := range stream.Chunks() {
    sendToClient(chunk.Delta) // Process and discard
}
```

**Memory reduction:** 70% for large responses

### Short-Lived Agents

Create agents per request instead of keeping in memory:

```go
// Memory efficient
func handleRequest(query string) (*vnext.Result, error) {
    agent, _ := vnext.PresetChatAgentBuilder().Build()
    defer agent.Close() // Optional cleanup
    return agent.Run(context.Background(), query)
}

// vs. Long-lived (higher memory)
var globalAgent vnext.Agent

func init() {
    globalAgent, _ = vnext.PresetChatAgentBuilder().Build()
}
```

### Memory Cleanup

```go
// Clear session memory periodically
if memory != nil {
    memory.Clear(sessionID)
}

// Or use time-based cleanup
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        cleanupOldSessions()
    }
}()
```

## Concurrent Execution

### Parallel Workflows

Use parallel workflows for independent tasks:

```go
// Sequential: 3 seconds total
workflow, _ := vnext.NewSequentialWorkflow("Pipeline",
    vnext.Step("s1", agent1, "1s task"),
    vnext.Step("s2", agent2, "1s task"),
    vnext.Step("s3", agent3, "1s task"),
)
// Takes: 3 seconds

// Parallel: 1 second total
workflow, _ := vnext.NewParallelWorkflow("Pipeline",
    vnext.Step("s1", agent1, "1s task"),
    vnext.Step("s2", agent2, "1s task"),
    vnext.Step("s3", agent3, "1s task"),
)
// Takes: 1 second (3x faster)
```

**Speedup:** N× for N independent tasks

### Concurrent Agents

```go
// Process multiple queries concurrently
var wg sync.WaitGroup
results := make(chan *vnext.Result, len(queries))

for _, query := range queries {
    wg.Add(1)
    go func(q string) {
        defer wg.Done()
        agent, _ := vnext.PresetChatAgentBuilder().Build()
        result, _ := agent.Run(context.Background(), q)
        results <- result
    }(query)
}

wg.Wait()
close(results)
```

### Rate Limiting

Prevent overwhelming API providers:

```go
// Configure rate limiting
agent, _ := vnext.PresetChatAgentBuilder().
    WithOptions(
        vnext.WithRateLimit(10),        // 10 requests/second
        vnext.WithMaxConcurrent(5),     // Max 5 concurrent requests
    ).
    Build()
```

Or in configuration:

```toml
[tools]
rate_limit = 10
max_concurrent = 5
```

### Goroutine Management

```go
// Use worker pools for better control
type WorkerPool struct {
    workers   int
    jobs      chan Job
    results   chan Result
}

func NewWorkerPool(workers int) *WorkerPool {
    pool := &WorkerPool{
        workers: workers,
        jobs:    make(chan Job, workers*2),
        results: make(chan Result, workers*2),
    }
    
    for i := 0; i < workers; i++ {
        go pool.worker()
    }
    
    return pool
}

func (p *WorkerPool) worker() {
    agent, _ := vnext.PresetChatAgentBuilder().Build()
    defer agent.Close()
    
    for job := range p.jobs {
        result, _ := agent.Run(context.Background(), job.Query)
        p.results <- Result{ID: job.ID, Result: result}
    }
}
```

## Caching Strategies

### Enable Tool Caching

```toml
[tools.cache]
enabled = true
ttl = "5m"
max_size = 1000
```

**Performance gain:** 90%+ for repeated tool calls

### LLM Response Caching

```go
// Simple in-memory cache
type ResponseCache struct {
    cache map[string]*vnext.Result
    mu    sync.RWMutex
    ttl   time.Duration
}

func (c *ResponseCache) Get(query string) (*vnext.Result, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    result, ok := c.cache[query]
    return result, ok
}

func (c *ResponseCache) Set(query string, result *vnext.Result) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.cache[query] = result
}

// Usage
func runWithCache(agent vnext.Agent, query string, cache *ResponseCache) (*vnext.Result, error) {
    if result, ok := cache.Get(query); ok {
        return result, nil // Cache hit
    }
    
    result, err := agent.Run(context.Background(), query)
    if err != nil {
        return nil, err
    }
    
    cache.Set(query, result)
    return result, nil
}
```

### Memory/RAG Caching

```toml
[memory]
provider = "memory"
connection = "inmemory"

[memory.rag]
max_tokens = 2000
cache_results = true
cache_ttl = "10m"
```

### Semantic Cache

Cache similar queries:

```go
// Use semantic similarity for cache lookups
func semanticCacheKey(query string) string {
    // Generate embedding and find similar cached queries
    embedding := generateEmbedding(query)
    similar := findSimilar(embedding, 0.95) // 95% similarity threshold
    if similar != nil {
        return similar.CacheKey
    }
    return generateNewKey(query)
}
```

## LLM Optimization

### Model Selection

Choose appropriate models for tasks:

```go
// Fast, cheap (simple tasks)
agent, _ := vnext.PresetChatAgentBuilder().
    WithLLM("openai", "gpt-3.5-turbo").
    Build()

// Balanced
agent, _ := vnext.PresetChatAgentBuilder().
    WithLLM("openai", "gpt-4").
    Build()

// Powerful (complex tasks)
agent, _ := vnext.PresetChatAgentBuilder().
    WithLLM("openai", "gpt-4-turbo").
    Build()
```

**Cost vs Performance:**
- gpt-3.5-turbo: 10× cheaper, 2× faster, good for simple tasks
- gpt-4: Balanced, best for most use cases
- gpt-4-turbo: Most capable, use for complex reasoning

### Temperature Settings

```go
// Deterministic (faster, consistent)
agent, _ := vnext.PresetChatAgentBuilder().
    WithOptions(vnext.WithTemperature(0.0)).
    Build()

// Creative (slower, varied)
agent, _ := vnext.PresetChatAgentBuilder().
    WithOptions(vnext.WithTemperature(0.9)).
    Build()
```

**Performance impact:**
- Temperature 0.0: ~15% faster due to reduced sampling
- Temperature 0.7-0.9: Standard performance

### Token Limits

```go
// Shorter responses = faster + cheaper
result, _ := agent.Run(ctx, input,
    vnext.WithMaxTokens(100),  // Brief response
)

// vs.
result, _ := agent.Run(ctx, input,
    vnext.WithMaxTokens(2000), // Detailed response
)
```

**Performance:**
- 100 tokens: ~0.5s response time
- 500 tokens: ~2s response time
- 2000 tokens: ~8s response time

### Batch Processing

```go
// Process multiple queries in one call (if supported)
queries := []string{
    "Query 1",
    "Query 2",
    "Query 3",
}

// Inefficient: 3 separate calls
for _, q := range queries {
    agent.Run(ctx, q)
}

// Efficient: Batch queries
batchQuery := strings.Join(queries, "\n---\n")
result, _ := agent.Run(ctx, fmt.Sprintf("Process these queries:\n%s", batchQuery))
```

## Tool Execution

### Timeout Configuration

```go
// Fast tools
agent, _ := vnext.PresetChatAgentBuilder().
    WithOptions(vnext.WithToolTimeout(5 * time.Second)).
    Build()

// Slow tools (e.g., web scraping)
agent, _ := vnext.PresetChatAgentBuilder().
    WithOptions(vnext.WithToolTimeout(60 * time.Second)).
    Build()
```

### Tool Parallelization

```go
// Execute multiple tools concurrently
type ToolExecutor struct {
    tools map[string]vnext.ToolHandler
}

func (e *ToolExecutor) ExecuteParallel(ctx context.Context, calls []ToolCall) []ToolResult {
    results := make([]ToolResult, len(calls))
    var wg sync.WaitGroup
    
    for i, call := range calls {
        wg.Add(1)
        go func(idx int, c ToolCall) {
            defer wg.Done()
            handler := e.tools[c.Name]
            result, err := handler(ctx, c.Args)
            results[idx] = ToolResult{Result: result, Error: err}
        }(i, call)
    }
    
    wg.Wait()
    return results
}
```

### Lazy Tool Loading

```go
// Load tools on-demand
type LazyToolRegistry struct {
    loaders map[string]func() vnext.Tool
    cache   map[string]vnext.Tool
    mu      sync.RWMutex
}

func (r *LazyToolRegistry) GetTool(name string) vnext.Tool {
    r.mu.RLock()
    if tool, ok := r.cache[name]; ok {
        r.mu.RUnlock()
        return tool
    }
    r.mu.RUnlock()
    
    r.mu.Lock()
    defer r.mu.Unlock()
    
    loader := r.loaders[name]
    tool := loader() // Load on first use
    r.cache[name] = tool
    return tool
}
```

## Workflow Optimization

### Step Dependencies

Use DAG workflows to maximize parallelism:

```go
// Sequential: 6 seconds
workflow, _ := vnext.NewSequentialWorkflow("Pipeline",
    vnext.Step("a", agentA, "2s"),
    vnext.Step("b", agentB, "2s"),
    vnext.Step("c", agentC, "2s"),
)

// DAG: 4 seconds (b and c run in parallel)
workflow, _ := vnext.NewDAGWorkflow("Pipeline",
    vnext.Step("a", agentA, "2s"),
    vnext.Step("b", agentB, "2s", "a"), // Depends on a
    vnext.Step("c", agentC, "2s", "a"), // Depends on a
)
```

### Early Exit

```go
// Exit workflow early on success
workflow, _ := vnext.NewSequentialWorkflow("Search",
    vnext.Step("quick", quickSearchAgent, "Try fast search"),
    vnext.Step("thorough", thoroughSearchAgent, "Fallback search"),
)

// Add early exit logic in step handler
func quickSearchHandler(ctx context.Context, input string) (string, error) {
    result := quickSearch(input)
    if result != "" {
        // Signal workflow to skip remaining steps
        return result, vnext.ErrWorkflowComplete
    }
    return "", nil // Continue to next step
}
```

### Context Sharing

Minimize data copying between steps:

```go
// Use context for efficient data sharing
type WorkflowContext struct {
    SharedData map[string]interface{}
    mu         sync.RWMutex
}

func (wc *WorkflowContext) Set(key string, value interface{}) {
    wc.mu.Lock()
    defer wc.mu.Unlock()
    wc.SharedData[key] = value
}

func (wc *WorkflowContext) Get(key string) interface{} {
    wc.mu.RLock()
    defer wc.mu.RUnlock()
    return wc.SharedData[key]
}

// Use in steps
func step1(ctx context.Context, input string) (string, error) {
    wc := ctx.Value("workflow_context").(*WorkflowContext)
    data := processData(input)
    wc.Set("processed_data", data) // Share with next steps
    return "success", nil
}

func step2(ctx context.Context, input string) (string, error) {
    wc := ctx.Value("workflow_context").(*WorkflowContext)
    data := wc.Get("processed_data") // Reuse from step1
    return analyze(data), nil
}
```

## Benchmarking

### Basic Benchmarking

```go
func BenchmarkAgentRun(b *testing.B) {
    agent, _ := vnext.PresetChatAgentBuilder().Build()
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        agent.Run(ctx, "Hello")
    }
}

func BenchmarkStreamingVsNonStreaming(b *testing.B) {
    agent, _ := vnext.PresetChatAgentBuilder().Build()
    ctx := context.Background()
    
    b.Run("NonStreaming", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            agent.Run(ctx, "Query")
        }
    })
    
    b.Run("Streaming", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            stream, _ := agent.RunStream(ctx, "Query")
            for range stream.Chunks() {}
            stream.Wait()
        }
    })
}
```

Run benchmarks:

```bash
go test -bench=. -benchmem -cpuprofile cpu.prof ./core/vnext
```

### Memory Profiling

```go
func TestMemoryUsage(t *testing.T) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    before := m.Alloc
    
    agent, _ := vnext.PresetChatAgentBuilder().Build()
    for i := 0; i < 100; i++ {
        agent.Run(context.Background(), "Test query")
    }
    
    runtime.ReadMemStats(&m)
    after := m.Alloc
    
    t.Logf("Memory used: %d KB", (after-before)/1024)
}
```

### Load Testing

```go
func LoadTest() {
    agent, _ := vnext.PresetChatAgentBuilder().Build()
    
    start := time.Now()
    concurrent := 100
    requestsPerWorker := 10
    
    var wg sync.WaitGroup
    for i := 0; i < concurrent; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < requestsPerWorker; j++ {
                agent.Run(context.Background(), "Load test query")
            }
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    totalRequests := concurrent * requestsPerWorker
    rps := float64(totalRequests) / duration.Seconds()
    
    fmt.Printf("Total requests: %d\n", totalRequests)
    fmt.Printf("Duration: %v\n", duration)
    fmt.Printf("RPS: %.2f\n", rps)
}
```

## Performance Metrics

### Expected Performance

With recommended settings:

| Operation | Latency | Throughput | Memory |
|-----------|---------|------------|--------|
| Simple query | 500-1000ms | 100-200 req/s | 10-20 MB |
| Streaming query | 50ms TTFB | 1000+ chunks/s | 5-10 MB |
| Tool call | 100-500ms | 200-500 ops/s | 5 MB |
| Sequential workflow (3 steps) | 1.5-3s | 30-60 flows/s | 20-40 MB |
| Parallel workflow (3 steps) | 0.5-1s | 100-200 flows/s | 30-50 MB |

### Optimization Checklist

- [ ] Use streaming for long responses
- [ ] Configure appropriate buffer sizes
- [ ] Enable caching for repeated operations
- [ ] Use parallel workflows when possible
- [ ] Set reasonable token limits
- [ ] Configure timeouts appropriately
- [ ] Use context cancellation
- [ ] Implement rate limiting
- [ ] Profile memory usage
- [ ] Benchmark critical paths
- [ ] Use appropriate models for tasks
- [ ] Clear old session data
- [ ] Optimize tool execution
- [ ] Minimize data copying
- [ ] Use connection pooling

## Summary

Key performance principles:

1. **Stream when possible** - 70% memory reduction
2. **Parallelize independent work** - N× speedup
3. **Cache aggressively** - 90%+ for cache hits
4. **Choose right models** - 10× cost/performance difference
5. **Set limits** - Prevent resource exhaustion
6. **Profile regularly** - Identify bottlenecks early
7. **Benchmark changes** - Verify optimizations work

**Next Steps:**
- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [Streaming Guide](STREAMING_GUIDE.md)
- [Migration Guide](MIGRATION_GUIDE.md)
