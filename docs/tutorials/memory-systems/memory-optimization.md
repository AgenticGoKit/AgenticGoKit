# Memory Optimization in AgenticGoKit

## Overview

Memory optimization is crucial for building scalable, high-performance agent systems. This tutorial covers advanced techniques for optimizing memory systems in AgenticGoKit, including performance tuning, caching strategies, resource management, and scaling patterns.

## Prerequisites

- Understanding of [Vector Databases](vector-databases.md)
- Familiarity with [RAG Implementation](rag-implementation.md)
- Knowledge of [Knowledge Bases](knowledge-bases.md)
- Basic understanding of database performance tuning

## Performance Optimization Strategies

### 1. Vector Database Optimization

```go
// Optimized pgvector configuration
func createOptimizedPgvectorMemory() (core.Memory, error) {
    config := core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://user:pass@localhost:5432/agentdb",
        EnableRAG:  true,
        Dimensions: 1536,
        Embedding: core.EmbeddingConfig{
            Provider:   "openai",
            Model:      "text-embedding-3-small",
            APIKey:     os.Getenv("OPENAI_API_KEY"),
            Dimensions: 1536,
            BatchSize:  100,
        },
        Options: map[string]interface{}{
            // Connection pool settings
            "max_connections":     20,
            "max_idle_connections": 5,
            "connection_timeout":  "30s",
            
            // Index optimization
            "index_type":      "hnsw",
            "hnsw_m":         16,
            "hnsw_ef_construction": 64,
            
            // Memory settings
            "work_mem":        "256MB",
            "effective_cache_size": "4GB",
        },
    }
    
    return core.NewMemory(config)
}
```

### 2. Caching Strategies

```go
type OptimizedMemoryCache struct {
    queryCache    *LRUCache
    embeddingCache *LRUCache
    resultCache   *LRUCache
    config        CacheConfig
}

type CacheConfig struct {
    QueryCacheSize     int
    EmbeddingCacheSize int
    ResultCacheSize    int
    TTL               time.Duration
}

func NewOptimizedMemoryCache(config CacheConfig) *OptimizedMemoryCache {
    return &OptimizedMemoryCache{
        queryCache:    NewLRUCache(config.QueryCacheSize, config.TTL),
        embeddingCache: NewLRUCache(config.EmbeddingCacheSize, config.TTL),
        resultCache:   NewLRUCache(config.ResultCacheSize, config.TTL),
        config:        config,
    }
}

func (omc *OptimizedMemoryCache) GetCachedResults(query string) ([]core.MemoryResult, bool) {
    if results, found := omc.queryCache.Get(query); found {
        return results.([]core.MemoryResult), true
    }
    return nil, false
}

func (omc *OptimizedMemoryCache) CacheResults(query string, results []core.MemoryResult) {
    omc.queryCache.Set(query, results)
}

func (omc *OptimizedMemoryCache) GetCachedEmbedding(text string) ([]float32, bool) {
    if embedding, found := omc.embeddingCache.Get(text); found {
        return embedding.([]float32), true
    }
    return nil, false
}

func (omc *OptimizedMemoryCache) CacheEmbedding(text string, embedding []float32) {
    omc.embeddingCache.Set(text, embedding)
}
```

### 3. Batch Processing Optimization

```go
type BatchProcessor struct {
    memory     core.Memory
    batchSize  int
    timeout    time.Duration
    semaphore  chan struct{}
}

func NewBatchProcessor(memory core.Memory, batchSize int, concurrency int) *BatchProcessor {
    return &BatchProcessor{
        memory:    memory,
        batchSize: batchSize,
        timeout:   30 * time.Second,
        semaphore: make(chan struct{}, concurrency),
    }
}

func (bp *BatchProcessor) ProcessBatch(ctx context.Context, items []string) error {
    // Process items in batches
    for i := 0; i < len(items); i += bp.batchSize {
        end := i + bp.batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        
        // Acquire semaphore for concurrency control
        select {
        case bp.semaphore <- struct{}{}:
            go func(b []string) {
                defer func() { <-bp.semaphore }()
                bp.processBatch(ctx, b)
            }(batch)
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return nil
}

func (bp *BatchProcessor) processBatch(ctx context.Context, batch []string) error {
    ctx, cancel := context.WithTimeout(ctx, bp.timeout)
    defer cancel()
    
    for _, item := range batch {
        err := bp.memory.Store(ctx, item, "batch-item")
        if err != nil {
            log.Printf("Failed to store batch item: %v", err)
        }
    }
    
    return nil
}
```

## Resource Management

### 1. Connection Pool Optimization

```go
type OptimizedConnectionManager struct {
    pool    *sql.DB
    metrics *ConnectionMetrics
    config  PoolConfig
}

type PoolConfig struct {
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

func NewOptimizedConnectionManager(dsn string, config PoolConfig) (*OptimizedConnectionManager, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(config.MaxOpenConns)
    db.SetMaxIdleConns(config.MaxIdleConns)
    db.SetConnMaxLifetime(config.ConnMaxLifetime)
    db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
    
    return &OptimizedConnectionManager{
        pool:    db,
        metrics: NewConnectionMetrics(),
        config:  config,
    }, nil
}

func (ocm *OptimizedConnectionManager) GetConnection(ctx context.Context) (*sql.Conn, error) {
    start := time.Now()
    
    conn, err := ocm.pool.Conn(ctx)
    if err != nil {
        ocm.metrics.RecordError()
        return nil, err
    }
    
    ocm.metrics.RecordAcquisition(time.Since(start))
    return conn, nil
}
```

### 2. Memory Usage Monitoring

```go
type MemoryMonitor struct {
    thresholds MemoryThresholds
    alerts     chan Alert
    interval   time.Duration
}

type MemoryThresholds struct {
    WarningMB  int64
    CriticalMB int64
}

type Alert struct {
    Level   string
    Message string
    Time    time.Time
}

func NewMemoryMonitor(thresholds MemoryThresholds) *MemoryMonitor {
    return &MemoryMonitor{
        thresholds: thresholds,
        alerts:     make(chan Alert, 100),
        interval:   30 * time.Second,
    }
}

func (mm *MemoryMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(mm.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            mm.checkMemoryUsage()
        }
    }
}

func (mm *MemoryMonitor) checkMemoryUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    heapMB := int64(m.HeapAlloc / 1024 / 1024)
    
    if heapMB > mm.thresholds.CriticalMB {
        mm.alerts <- Alert{
            Level:   "CRITICAL",
            Message: fmt.Sprintf("Memory usage: %d MB", heapMB),
            Time:    time.Now(),
        }
        runtime.GC() // Force garbage collection
    } else if heapMB > mm.thresholds.WarningMB {
        mm.alerts <- Alert{
            Level:   "WARNING",
            Message: fmt.Sprintf("Memory usage: %d MB", heapMB),
            Time:    time.Now(),
        }
    }
}

func (mm *MemoryMonitor) GetAlerts() <-chan Alert {
    return mm.alerts
}
```

## Performance Monitoring

### 1. Metrics Collection

```go
type PerformanceMetrics struct {
    searchLatency    []time.Duration
    cacheHitRate     float64
    throughput       int64
    errorCount       int64
    mu               sync.RWMutex
}

func NewPerformanceMetrics() *PerformanceMetrics {
    return &PerformanceMetrics{
        searchLatency: make([]time.Duration, 0, 1000),
    }
}

func (pm *PerformanceMetrics) RecordSearch(duration time.Duration, cached bool) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    pm.searchLatency = append(pm.searchLatency, duration)
    pm.throughput++
    
    // Keep only recent measurements
    if len(pm.searchLatency) > 1000 {
        pm.searchLatency = pm.searchLatency[len(pm.searchLatency)-1000:]
    }
    
    // Update cache hit rate
    if cached {
        pm.cacheHitRate = (pm.cacheHitRate + 1.0) / 2.0
    } else {
        pm.cacheHitRate = pm.cacheHitRate / 2.0
    }
}

func (pm *PerformanceMetrics) GetAverageLatency() time.Duration {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    
    if len(pm.searchLatency) == 0 {
        return 0
    }
    
    var total time.Duration
    for _, latency := range pm.searchLatency {
        total += latency
    }
    
    return total / time.Duration(len(pm.searchLatency))
}

func (pm *PerformanceMetrics) GetReport() map[string]interface{} {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    
    return map[string]interface{}{
        "avg_latency_ms": pm.GetAverageLatency().Milliseconds(),
        "cache_hit_rate": pm.cacheHitRate,
        "throughput":     pm.throughput,
        "error_count":    pm.errorCount,
    }
}
```

## Best Practices

### 1. Configuration Optimization

- **Connection Pooling**: Configure appropriate pool sizes based on workload
- **Index Selection**: Choose HNSW for better query performance, IVFFlat for faster indexing
- **Memory Settings**: Tune PostgreSQL memory settings for vector operations
- **Batch Sizes**: Optimize batch sizes for embedding API calls
- **Cache Sizes**: Size caches based on available memory and hit rate requirements

### 2. Performance Monitoring

- **Latency Tracking**: Monitor search and indexing latencies
- **Resource Usage**: Track memory, CPU, and disk usage
- **Cache Performance**: Monitor cache hit rates and effectiveness
- **Error Rates**: Track and alert on error rates
- **Throughput**: Monitor requests per second and concurrent users

### 3. Scaling Strategies

- **Vertical Scaling**: Increase memory and CPU for single-node performance
- **Read Replicas**: Use read replicas for read-heavy workloads
- **Sharding**: Distribute data across multiple nodes for large datasets
- **Caching Layers**: Implement multiple caching layers for frequently accessed data
- **Load Balancing**: Distribute requests across multiple instances

## Conclusion

Memory optimization is essential for production-ready agent systems. Key takeaways:

- Implement comprehensive caching strategies at multiple levels
- Monitor resource usage and performance metrics continuously
- Optimize database configurations for vector operations
- Use appropriate scaling patterns based on workload characteristics
- Plan for growth and implement monitoring from the beginning

Proper optimization ensures your agents can handle production workloads efficiently and cost-effectively.

## Next Steps

- [Production Deployment](../deployment/README.md) - Deploy optimized systems
- [Monitoring and Observability](../debugging-monitoring/README.md) - Advanced monitoring
- [Advanced Patterns](../advanced-patterns/README.md) - Learn scaling techniques

## Further Reading

- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Vector Database Benchmarks](https://github.com/erikbern/ann-benchmarks)
- [Go Performance Optimization](https://github.com/dgryski/go-perfbook)