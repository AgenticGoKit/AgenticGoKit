# Performance Optimization

Learn how to optimize your v1beta agents for maximum performance, efficiency, and scalability.

---

## 🎯 Overview

Optimization strategies for AgenticGoKit v1beta:

- **Streaming** - Reduce latency and memory usage
- **Caching** - Minimize redundant operations  
- **Concurrency** - Parallelize independent work
- **Memory Management** - Control resource usage
- **LLM Optimization** - Choose appropriate models and settings
- **Tool Execution** - Optimize external integrations
- **Workflow Efficiency** - Maximize parallel execution

---

## 🚀 Quick Wins

### 1. Use Streaming for Long Responses

```go
// ❌ Slow: Wait for full response
result, _ := agent.Run(ctx, longQuery)
fmt.Println(result.Content) // All at once after waiting

// ✅ Fast: Stream as it generates
stream, _ := agent.RunStream(ctx, longQuery)
for chunk := range stream.Chunks() {
    if chunk.Type == "text" {
        fmt.Print(chunk.Delta) // Show immediately
    }
}
```

**Performance gain:** 70% memory reduction, perceived latency improvement

### 2. Enable Tool Caching

```toml
[tools.cache]
enabled = true
ttl = "15m"
max_size = 100  # MB
```

**Performance gain:** 90%+ speedup for repeated tool calls

### 3. Use Parallel Workflows

```go
// ❌ Sequential: 6 seconds total (2s + 2s + 2s)
workflow, _ := v1beta.NewSequentialWorkflow("tasks",
    v1beta.Step("step1", agent1, "task1"),
    v1beta.Step("step2", agent2, "task2"),
    v1beta.Step("step3", agent3, "task3"),
)

// ✅ Parallel: 2 seconds total (all at once)
workflow, _ := v1beta.NewParallelWorkflow("tasks",
    v1beta.Step("step1", agent1, "task1"),
    v1beta.Step("step2", agent2, "task2"),
    v1beta.Step("step3", agent3, "task3"),
)
```

**Performance gain:** 3× speedup for independent tasks

---

## 📡 Streaming Optimization

### Buffer Sizing

Choose buffer size based on use case:

```go
// Real-time chat (low latency)
stream, _ := agent.RunStream(ctx, query,
    v1beta.WithBufferSize(50),
)

// Balanced (recommended default)
stream, _ := agent.RunStream(ctx, query,
    v1beta.WithBufferSize(100),
)

// Batch processing (high throughput)
stream, _ := agent.RunStream(ctx, query,
    v1beta.WithBufferSize(500),
)
```

**Guidelines:**
- Real-time UI: 25-50
- Interactive chat: 50-100  
- Data processing: 200-500
- Batch operations: 500-1000

### Flush Intervals

Control update frequency:

```go
// Immediate updates (more CPU)
v1beta.WithFlushInterval(10 * time.Millisecond)

// Balanced (recommended)
v1beta.WithFlushInterval(100 * time.Millisecond)

// Batched (less CPU)
v1beta.WithFlushInterval(500 * time.Millisecond)
```

**Impact:**
- Shorter: Lower latency, higher CPU usage
- Longer: Higher latency, lower CPU, better throughput

### Text-Only Mode

Skip unnecessary metadata:

```go
stream, _ := agent.RunStream(ctx, query,
    v1beta.WithTextOnly(true), // Skip thoughts, tools, metadata
)
```

**Performance gain:** ~30% reduction in chunk processing overhead

### Stream Processing Patterns

```go
// Fastest: Direct chunk processing
for chunk := range stream.Chunks() {
    processChunk(chunk)
}

// Fast: Collect to string
text, _ := v1beta.CollectStream(stream)

// Moderate: Stream to channel
textChan := v1beta.StreamToChannel(stream)

// Slower: AsReader (adds buffering layer)
reader := stream.AsReader()
```

---

## 💾 Memory Management

### Context Size Limits

Reduce memory footprint:

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

result, _ := agent.RunWithOptions(ctx, input, &v1beta.RunOptions{
    MaxTokens:    1000,  // Limit output size
    HistoryLimit: 10,    // Keep last 10 messages
})
```

**Memory savings:** Up to 80% for long conversations

### Streaming vs Buffering

```go
// ❌ High memory: Buffer full response
result, _ := agent.Run(ctx, input)
fullText := result.Content // Entire response in memory

// ✅ Low memory: Stream and process
stream, _ := agent.RunStream(ctx, input)
for chunk := range stream.Chunks() {
    sendToClient(chunk.Delta) // Process and discard
}
```

**Memory reduction:** 70% for large responses

### Short-Lived Agents

Create agents per request:

```go
// ✅ Memory efficient
func handleRequest(query string) (*v1beta.Result, error) {
    agent, _ := v1beta.NewBuilder("agent").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        Build()
    defer agent.Cleanup(context.Background())
    return agent.Run(context.Background(), query)
}

// ❌ Higher memory (long-lived)
var globalAgent v1beta.Agent

func init() {
    globalAgent, _ = v1beta.NewBuilder("agent").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        Build()
}
```

### Memory Cleanup

```go
// Clear session memory periodically
if memory != nil {
    memory.Clear(sessionID)
}

// Time-based cleanup
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        cleanupOldSessions()
    }
}()
```

---

## ⚡ Concurrent Execution

### Parallel Workflows

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

// Sequential: 3 seconds
workflow, _ := v1beta.NewSequentialWorkflow("pipeline",
    v1beta.Step("s1", agent1, "1s task"),
    v1beta.Step("s2", agent2, "1s task"),
    v1beta.Step("s3", agent3, "1s task"),
)

// Parallel: 1 second (3× faster)
workflow, _ := v1beta.NewParallelWorkflow("pipeline",
    v1beta.Step("s1", agent1, "1s task"),
    v1beta.Step("s2", agent2, "1s task"),
    v1beta.Step("s3", agent3, "1s task"),
)
```

### Concurrent Agents

```go
import "sync"

var wg sync.WaitGroup
results := make(chan *v1beta.Result, len(queries))

for _, query := range queries {
    wg.Add(1)
    go func(q string) {
        defer wg.Done()
        agent, _ := v1beta.NewBuilder("agent").
            WithConfig(&v1beta.Config{
                LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
            }).
            Build()
        result, _ := agent.Run(context.Background(), q)
        results <- result
    }(query)
}

wg.Wait()
close(results)
```

### Rate Limiting

Prevent overwhelming API providers:

```toml
[tools]
rate_limit = 10        # 10 requests/second
max_concurrent = 5     # Max 5 parallel executions
```

Or programmatically:

```go
agent, _ := v1beta.NewBuilder("agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithTools(
        v1beta.WithToolRateLimit(10),
        v1beta.WithMaxConcurrentTools(5),
    ).
    Build()
```

### Worker Pools

Better goroutine management:

```go
type WorkerPool struct {
    workers int
    jobs    chan Job
    results chan Result
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
    agent, _ := v1beta.NewBuilder("worker").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        Build()
    defer agent.Cleanup(context.Background())
    
    for job := range p.jobs {
        result, _ := agent.Run(context.Background(), job.Query)
        p.results <- Result{ID: job.ID, Result: result}
    }
}
```

---

## 💰 Caching Strategies

### Tool Result Caching

```toml
[tools.cache]
enabled = true
ttl = "15m"
max_size = 100        # MB
max_keys = 10000
eviction_policy = "lru"

[tools.cache.tool_ttls]
web_search = "5m"     # Short TTL for dynamic data
content_fetch = "30m" # Medium TTL
static_api = "24h"    # Long TTL for static data
```

**Performance gain:** 90%+ for cache hits

### LLM Response Caching

```go
import "sync"

type ResponseCache struct {
    cache map[string]*v1beta.Result
    mu    sync.RWMutex
    ttl   time.Duration
}

func (c *ResponseCache) Get(query string) (*v1beta.Result, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    result, ok := c.cache[query]
    return result, ok
}

func (c *ResponseCache) Set(query string, result *v1beta.Result) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.cache[query] = result
}

// Usage
func runWithCache(agent v1beta.Agent, query string, cache *ResponseCache) (*v1beta.Result, error) {
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

[memory.rag]
max_tokens = 2000
cache_results = true
cache_ttl = "10m"
```

### Semantic Caching

Cache similar queries:

```go
func semanticCacheKey(query string) string {
    // Generate embedding and find similar cached queries
    embedding := generateEmbedding(query)
    similar := findSimilar(embedding, 0.95) // 95% similarity
    if similar != nil {
        return similar.CacheKey
    }
    return generateNewKey(query)
}
```

---

## 🤖 LLM Optimization

### Model Selection

Choose appropriate models:

```go
// Fast, cheap (simple tasks)
agent, _ := v1beta.NewBuilder("agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-3.5-turbo"},
    }).
    Build()

// Balanced (most use cases)
agent, _ := v1beta.NewBuilder("agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    Build()

// Powerful (complex reasoning)
agent, _ := v1beta.NewBuilder("agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4-turbo"},
    }).
    Build()
```

**Cost vs Performance:**
- gpt-3.5-turbo: 10× cheaper, 2× faster, good for simple tasks
- gpt-4: Balanced, best for most cases
- gpt-4-turbo: Most capable, use for complex reasoning

### Temperature Settings

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

// Deterministic (faster, consistent)
config := &v1beta.Config{
    LLM: v1beta.LLMConfig{
        Provider:    "openai",
        Model:       "gpt-4",
        Temperature: 0.0,
    },
}

// Creative (slower, varied)
config := &v1beta.Config{
    LLM: v1beta.LLMConfig{
        Provider:    "openai",
        Model:       "gpt-4",
        Temperature: 0.9,
    },
}
```

**Performance impact:**
- Temperature 0.0: ~15% faster due to reduced sampling
- Temperature 0.7-0.9: Standard performance

### Token Limits

```go
// Shorter responses = faster + cheaper
result, _ := agent.RunWithOptions(ctx, input, &v1beta.RunOptions{
    MaxTokens: 100, // Brief response
})

// vs.
result, _ := agent.RunWithOptions(ctx, input, &v1beta.RunOptions{
    MaxTokens: 2000, // Detailed response
})
```

**Performance:**
- 100 tokens: ~0.5s response time
- 500 tokens: ~2s response time
- 2000 tokens: ~8s response time

### Batch Processing

```go
import "strings"

// ❌ Inefficient: 3 separate calls
for _, q := range queries {
    agent.Run(ctx, q)
}

// ✅ Efficient: Batch queries
batchQuery := strings.Join(queries, "\n---\n")
result, _ := agent.Run(ctx, fmt.Sprintf("Process these queries:\n%s", batchQuery))
```

---

## 🔧 Tool Execution

### Timeout Configuration

```go
agent, _ := v1beta.NewBuilder("agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithTools(
        v1beta.WithMCP(servers...),
        v1beta.WithToolTimeout(30 * time.Second), // Adjust based on tools
    ).
    Build()
```

**Guidelines:**
- Fast tools (calculators): 5s
- Standard tools (APIs): 30s
- Slow tools (web scraping): 60s+

### Tool Parallelization

Execute multiple tools concurrently:

```go
import "sync"

type ToolExecutor struct {
    tools map[string]func(context.Context, map[string]interface{}) (interface{}, error)
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

### Lazy Loading

Load tools on-demand:

```go
type LazyToolRegistry struct {
    loaders map[string]func() Tool
    cache   map[string]Tool
    mu      sync.RWMutex
}

func (r *LazyToolRegistry) GetTool(name string) Tool {
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

---

## 🔀 Workflow Optimization

### DAG Workflows

Maximize parallelism:

```go
// ❌ Sequential: 6 seconds
workflow, _ := v1beta.NewSequentialWorkflow("pipeline",
    v1beta.Step("a", agentA, "2s"),
    v1beta.Step("b", agentB, "2s"),
    v1beta.Step("c", agentC, "2s"),
)

// ✅ DAG: 4 seconds (b and c parallel)
workflow, _ := v1beta.NewDAGWorkflow("pipeline",
    v1beta.Step("a", agentA, "2s"),
    v1beta.Step("b", agentB, "2s", "a"), // Depends on a
    v1beta.Step("c", agentC, "2s", "a"), // Depends on a
)
```

### Early Exit

```go
// Exit workflow early on success
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    result := quickSearch(input)
    if result != "" {
        return result, nil // Skip remaining steps
    }
    return "", nil // Continue
}
```

### Context Sharing

Minimize data copying:

```go
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

// Use in workflow steps
ctx = context.WithValue(ctx, "workflow_context", wc)
```

---

## 📊 Benchmarking

### Basic Benchmarks

```go
func BenchmarkAgentRun(b *testing.B) {
    agent, _ := v1beta.NewBuilder("agent").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        Build()
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        agent.Run(ctx, "Hello")
    }
}

func BenchmarkStreamingVsNonStreaming(b *testing.B) {
    agent, _ := v1beta.NewBuilder("agent").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        Build()
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
go test -bench=. -benchmem ./...
```

### Memory Profiling

```go
func TestMemoryUsage(t *testing.T) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    before := m.Alloc
    
    agent, _ := v1beta.NewBuilder("agent").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        Build()
    
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
    agent, _ := v1beta.NewBuilder("agent").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        Build()
    
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

---

## 📈 Performance Metrics

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

---

## 🎯 Key Takeaways

1. **Stream when possible** - 70% memory reduction
2. **Parallelize independent work** - N× speedup
3. **Cache aggressively** - 90%+ for cache hits
4. **Choose right models** - 10× cost/performance difference
5. **Set limits** - Prevent resource exhaustion
6. **Profile regularly** - Identify bottlenecks early
7. **Benchmark changes** - Verify optimizations work

---

## 📚 Next Steps

- **[Troubleshooting](./troubleshooting.md)** - Common performance issues
- **[Streaming Guide](./streaming.md)** - Advanced streaming patterns
- **[Tool Integration](./tool-integration.md)** - Optimize tool usage
- **[Configuration](./configuration.md)** - Performance-related settings

---

**Ready to troubleshoot?** Continue to [Troubleshooting](./troubleshooting.md) →
