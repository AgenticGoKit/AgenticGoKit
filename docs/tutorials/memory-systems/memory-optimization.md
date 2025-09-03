---
title: Memory Optimization
description: Learn advanced techniques for optimizing memory systems in AgenticGoKit, including performance tuning, caching strategies, and scaling patterns using current APIs.
---

# Memory Optimization in AgenticGoKit

## Overview

Memory optimization is crucial for building scalable, high-performance agent systems. This tutorial covers advanced techniques for optimizing memory systems in AgenticGoKit using the current AgentMemoryConfig structure, embedding configurations, and performance patterns.

Proper optimization ensures your agents can handle production workloads efficiently and cost-effectively while maintaining fast response times and high accuracy.

## Prerequisites

- Understanding of [Vector Databases](vector-databases.md)
- Familiarity with [RAG Implementation](rag-implementation.md)
- Knowledge of [Knowledge Bases](knowledge-bases.md)
- Basic understanding of database performance tuning

## Current Memory Configuration Optimization

### 1. Optimized AgentMemoryConfig Structure

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math"
    "os"
    "runtime"
    "sync"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

// Production-optimized memory configuration
func createOptimizedMemoryConfig() core.AgentMemoryConfig {
    return core.AgentMemoryConfig{
        // Core memory settings - optimized for performance
        Provider:   "pgvector", // Use production vector database
        Connection: "postgres://user:pass@localhost:5432/agentdb?pool_max_conns=50&pool_min_conns=10",
        MaxResults: 20,  // Higher for better context
        Dimensions: 1536,
        AutoEmbed:  true,
        
        // RAG-enhanced settings - optimized for quality and performance
        EnableRAG:               true,
        EnableKnowledgeBase:     true,
        KnowledgeMaxResults:     50,  // Higher for comprehensive results
        KnowledgeScoreThreshold: 0.75, // Higher threshold for quality
        ChunkSize:               1500, // Larger chunks for better context
        ChunkOverlap:            300,  // More overlap for continuity
        
        // RAG context assembly settings - optimized for LLM performance
        RAGMaxContextTokens: 8000, // Larger context window
        RAGPersonalWeight:   0.2,  // Focus more on knowledge base
        RAGKnowledgeWeight:  0.8,
        RAGIncludeSources:   true,
        
        // Embedding configuration - optimized for throughput and caching
        Embedding: core.EmbeddingConfig{
            Provider:        "openai",
            Model:           "text-embedding-3-small", // Good balance of speed/quality
            APIKey:          os.Getenv("OPENAI_API_KEY"),
            CacheEmbeddings: true, // Critical for performance
            MaxBatchSize:    200,  // Larger batches for efficiency
            TimeoutSeconds:  60,   // Longer timeout for large batches
        },
        
        // Document processing - optimized for large-scale ingestion
        Documents: core.DocumentConfig{
            AutoChunk:                true,
            SupportedTypes:           []string{"pdf", "txt", "md", "web", "code", "json"},
            MaxFileSize:              "100MB", // Larger files for enterprise
            EnableMetadataExtraction: true,
            EnableURLScraping:        true,
        },
        
        // Search configuration - optimized for semantic search
        Search: core.SearchConfigToml{
            HybridSearch:         true,
            KeywordWeight:        0.2, // Favor semantic search
            SemanticWeight:       0.8,
            EnableReranking:      false, // Disable for performance unless needed
            EnableQueryExpansion: false, // Disable for performance unless needed
        },
    }
}
```### 2. Env
ironment-Specific Configurations

```go
// Development configuration - optimized for fast iteration
func createDevelopmentConfig() core.AgentMemoryConfig {
    return core.AgentMemoryConfig{
        Provider:   "memory", // In-memory for fast development
        Connection: "memory",
        MaxResults: 10,
        Dimensions: 1536,
        AutoEmbed:  true,
        
        EnableRAG:           true,
        EnableKnowledgeBase: true,
        
        Embedding: core.EmbeddingConfig{
            Provider:        "dummy", // No API calls during development
            Model:           "dummy-model",
            CacheEmbeddings: false,
            MaxBatchSize:    50,
            TimeoutSeconds:  10,
        },
    }
}

// Production configuration - optimized for scale and reliability
func createProductionConfig() core.AgentMemoryConfig {
    return core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: buildOptimizedConnectionString(),
        MaxResults: 15,
        Dimensions: 1536,
        AutoEmbed:  true,
        
        // Production RAG settings
        EnableRAG:               true,
        EnableKnowledgeBase:     true,
        KnowledgeMaxResults:     100, // Higher for comprehensive search
        KnowledgeScoreThreshold: 0.8, // Higher quality threshold
        ChunkSize:               2000, // Larger chunks for better context
        ChunkOverlap:            400,
        
        RAGMaxContextTokens: 12000, // Large context for complex queries
        RAGPersonalWeight:   0.15,
        RAGKnowledgeWeight:  0.85,
        RAGIncludeSources:   true,
        
        Embedding: core.EmbeddingConfig{
            Provider:        "openai",
            Model:           "text-embedding-3-large", // Higher quality for production
            APIKey:          os.Getenv("OPENAI_API_KEY"),
            CacheEmbeddings: true,
            MaxBatchSize:    300, // Larger batches for efficiency
            TimeoutSeconds:  120, // Longer timeout for reliability
        },
        
        Documents: core.DocumentConfig{
            AutoChunk:                true,
            SupportedTypes:           []string{"pdf", "txt", "md", "web", "code", "json"},
            MaxFileSize:              "500MB", // Large files for enterprise
            EnableMetadataExtraction: true,
            EnableURLScraping:        true,
        },
        
        Search: core.SearchConfigToml{
            HybridSearch:         true,
            KeywordWeight:        0.15,
            SemanticWeight:       0.85,
            EnableReranking:      true,  // Enable for production quality
            EnableQueryExpansion: false, // Keep disabled for performance
        },
    }
}

func buildOptimizedConnectionString() string {
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?"+
            "pool_max_conns=100&"+
            "pool_min_conns=20&"+
            "pool_max_conn_lifetime=1h&"+
            "pool_max_conn_idle_time=30m&"+
            "sslmode=require",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
    )
}
```#
# Advanced Caching Strategies

### 1. Multi-Level Caching System

```go
type OptimizedMemorySystem struct {
    memory      core.Memory
    l1Cache     *InMemoryCache     // Fast, small cache
    l2Cache     *RedisCache        // Larger, persistent cache
    l3Cache     *DatabaseCache     // Largest, slowest cache
    metrics     *CacheMetrics
    config      CacheConfig
}

type CacheConfig struct {
    L1Size          int
    L1TTL           time.Duration
    L2Size          int
    L2TTL           time.Duration
    L3TTL           time.Duration
    EnableL1        bool
    EnableL2        bool
    EnableL3        bool
}

type CacheMetrics struct {
    L1Hits, L1Misses int64
    L2Hits, L2Misses int64
    L3Hits, L3Misses int64
    mu               sync.RWMutex
}

func NewOptimizedMemorySystem(memory core.Memory, config CacheConfig) *OptimizedMemorySystem {
    oms := &OptimizedMemorySystem{
        memory:  memory,
        metrics: &CacheMetrics{},
        config:  config,
    }
    
    if config.EnableL1 {
        oms.l1Cache = NewInMemoryCache(config.L1Size, config.L1TTL)
    }
    if config.EnableL2 {
        oms.l2Cache = NewRedisCache(config.L2Size, config.L2TTL)
    }
    if config.EnableL3 {
        oms.l3Cache = NewDatabaseCache(config.L3TTL)
    }
    
    return oms
}

func (oms *OptimizedMemorySystem) SearchKnowledge(ctx context.Context, query string, options ...core.SearchOption) ([]core.KnowledgeResult, error) {
    cacheKey := oms.buildCacheKey(query, options)
    
    // Try L1 cache first (fastest)
    if oms.config.EnableL1 {
        if results := oms.l1Cache.Get(cacheKey); results != nil {
            oms.metrics.recordL1Hit()
            return results.([]core.KnowledgeResult), nil
        }
        oms.metrics.recordL1Miss()
    }
    
    // Try L2 cache (Redis)
    if oms.config.EnableL2 {
        if results := oms.l2Cache.Get(ctx, cacheKey); results != nil {
            oms.metrics.recordL2Hit()
            // Populate L1 cache
            if oms.config.EnableL1 {
                oms.l1Cache.Set(cacheKey, results)
            }
            return results.([]core.KnowledgeResult), nil
        }
        oms.metrics.recordL2Miss()
    }
    
    // Try L3 cache (Database)
    if oms.config.EnableL3 {
        if results := oms.l3Cache.Get(ctx, cacheKey); results != nil {
            oms.metrics.recordL3Hit()
            // Populate upper caches
            if oms.config.EnableL2 {
                oms.l2Cache.Set(ctx, cacheKey, results)
            }
            if oms.config.EnableL1 {
                oms.l1Cache.Set(cacheKey, results)
            }
            return results.([]core.KnowledgeResult), nil
        }
        oms.metrics.recordL3Miss()
    }
    
    // Cache miss - query the actual memory system
    results, err := oms.memory.SearchKnowledge(ctx, query, options...)
    if err != nil {
        return nil, err
    }
    
    // Populate all cache levels
    if oms.config.EnableL3 {
        oms.l3Cache.Set(ctx, cacheKey, results)
    }
    if oms.config.EnableL2 {
        oms.l2Cache.Set(ctx, cacheKey, results)
    }
    if oms.config.EnableL1 {
        oms.l1Cache.Set(cacheKey, results)
    }
    
    return results, nil
}

func (oms *OptimizedMemorySystem) buildCacheKey(query string, options []core.SearchOption) string {
    // Simple cache key generation - in production, use proper hashing
    return fmt.Sprintf("search:%s:%d", query, len(options))
}
```### 2. In
telligent Cache Management

```go
type IntelligentCacheManager struct {
    cache           map[string]*CacheEntry
    accessFrequency map[string]int64
    lastAccess      map[string]time.Time
    mu              sync.RWMutex
    maxSize         int
    ttl             time.Duration
}

type CacheEntry struct {
    Data      interface{}
    CreatedAt time.Time
    AccessCount int64
    Size      int64
}

func NewIntelligentCacheManager(maxSize int, ttl time.Duration) *IntelligentCacheManager {
    icm := &IntelligentCacheManager{
        cache:           make(map[string]*CacheEntry),
        accessFrequency: make(map[string]int64),
        lastAccess:      make(map[string]time.Time),
        maxSize:         maxSize,
        ttl:             ttl,
    }
    
    // Start cleanup goroutine
    go icm.cleanup()
    
    return icm
}

func (icm *IntelligentCacheManager) Get(key string) (interface{}, bool) {
    icm.mu.Lock()
    defer icm.mu.Unlock()
    
    entry, exists := icm.cache[key]
    if !exists {
        return nil, false
    }
    
    // Check TTL
    if time.Since(entry.CreatedAt) > icm.ttl {
        delete(icm.cache, key)
        delete(icm.accessFrequency, key)
        delete(icm.lastAccess, key)
        return nil, false
    }
    
    // Update access statistics
    entry.AccessCount++
    icm.accessFrequency[key]++
    icm.lastAccess[key] = time.Now()
    
    return entry.Data, true
}

func (icm *IntelligentCacheManager) Set(key string, data interface{}) {
    icm.mu.Lock()
    defer icm.mu.Unlock()
    
    // Calculate approximate size
    size := icm.calculateSize(data)
    
    // Check if we need to evict entries
    if len(icm.cache) >= icm.maxSize {
        icm.evictLeastUseful()
    }
    
    icm.cache[key] = &CacheEntry{
        Data:        data,
        CreatedAt:   time.Now(),
        AccessCount: 1,
        Size:        size,
    }
    icm.accessFrequency[key] = 1
    icm.lastAccess[key] = time.Now()
}

func (icm *IntelligentCacheManager) evictLeastUseful() {
    // Find the least useful entry based on frequency and recency
    var leastUsefulKey string
    var lowestScore float64 = math.MaxFloat64
    
    now := time.Now()
    for key, entry := range icm.cache {
        // Calculate usefulness score (higher = more useful)
        timeSinceAccess := now.Sub(icm.lastAccess[key]).Hours()
        frequency := float64(icm.accessFrequency[key])
        
        // Score combines frequency and recency
        score := frequency / (1 + timeSinceAccess)
        
        if score < lowestScore {
            lowestScore = score
            leastUsefulKey = key
        }
    }
    
    if leastUsefulKey != "" {
        delete(icm.cache, leastUsefulKey)
        delete(icm.accessFrequency, leastUsefulKey)
        delete(icm.lastAccess, leastUsefulKey)
    }
}

func (icm *IntelligentCacheManager) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        icm.mu.Lock()
        now := time.Now()
        
        for key, entry := range icm.cache {
            if now.Sub(entry.CreatedAt) > icm.ttl {
                delete(icm.cache, key)
                delete(icm.accessFrequency, key)
                delete(icm.lastAccess, key)
            }
        }
        icm.mu.Unlock()
    }
}

func (icm *IntelligentCacheManager) calculateSize(data interface{}) int64 {
    // Simplified size calculation
    switch v := data.(type) {
    case []core.KnowledgeResult:
        size := int64(0)
        for _, result := range v {
            size += int64(len(result.Content) + len(result.Title) + len(result.Source))
        }
        return size
    case string:
        return int64(len(v))
    default:
        return 1024 // Default size estimate
    }
}
```## P
erformance Monitoring and Metrics

### 1. Comprehensive Performance Monitoring

```go
type MemoryPerformanceMonitor struct {
    metrics     *PerformanceMetrics
    alerts      chan Alert
    thresholds  PerformanceThresholds
    interval    time.Duration
}

type PerformanceMetrics struct {
    // Search performance
    SearchLatency       []time.Duration
    SearchThroughput    int64
    SearchErrors        int64
    
    // Cache performance
    CacheHitRate        float64
    CacheSize           int64
    CacheEvictions      int64
    
    // Memory usage
    HeapSize            int64
    GCPauses            []time.Duration
    
    // Database performance
    ConnectionPoolUsage float64
    QueryLatency        []time.Duration
    
    mu sync.RWMutex
}

type PerformanceThresholds struct {
    MaxSearchLatency    time.Duration
    MinCacheHitRate     float64
    MaxMemoryUsage      int64
    MaxErrorRate        float64
}

type Alert struct {
    Level     string
    Component string
    Message   string
    Timestamp time.Time
    Value     interface{}
}

func NewMemoryPerformanceMonitor(thresholds PerformanceThresholds) *MemoryPerformanceMonitor {
    return &MemoryPerformanceMonitor{
        metrics:    &PerformanceMetrics{},
        alerts:     make(chan Alert, 1000),
        thresholds: thresholds,
        interval:   30 * time.Second,
    }
}

func (mpm *MemoryPerformanceMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(mpm.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            mpm.collectMetrics()
            mpm.checkThresholds()
        }
    }
}

func (mpm *MemoryPerformanceMonitor) RecordSearch(duration time.Duration, cached bool, err error) {
    mpm.metrics.mu.Lock()
    defer mpm.metrics.mu.Unlock()
    
    mpm.metrics.SearchLatency = append(mpm.metrics.SearchLatency, duration)
    mpm.metrics.SearchThroughput++
    
    if err != nil {
        mpm.metrics.SearchErrors++
    }
    
    // Update cache hit rate
    if cached {
        mpm.metrics.CacheHitRate = (mpm.metrics.CacheHitRate + 1.0) / 2.0
    } else {
        mpm.metrics.CacheHitRate = mpm.metrics.CacheHitRate * 0.99
    }
    
    // Keep only recent measurements
    if len(mpm.metrics.SearchLatency) > 1000 {
        mpm.metrics.SearchLatency = mpm.metrics.SearchLatency[len(mpm.metrics.SearchLatency)-1000:]
    }
}

func (mpm *MemoryPerformanceMonitor) collectMetrics() {
    // Collect Go runtime metrics
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    mpm.metrics.mu.Lock()
    mpm.metrics.HeapSize = int64(m.HeapAlloc)
    mpm.metrics.mu.Unlock()
}

func (mpm *MemoryPerformanceMonitor) checkThresholds() {
    mpm.metrics.mu.RLock()
    defer mpm.metrics.mu.RUnlock()
    
    // Check search latency
    if len(mpm.metrics.SearchLatency) > 0 {
        avgLatency := mpm.calculateAverageLatency()
        if avgLatency > mpm.thresholds.MaxSearchLatency {
            mpm.alerts <- Alert{
                Level:     "WARNING",
                Component: "search",
                Message:   "High search latency detected",
                Timestamp: time.Now(),
                Value:     avgLatency,
            }
        }
    }
    
    // Check cache hit rate
    if mpm.metrics.CacheHitRate < mpm.thresholds.MinCacheHitRate {
        mpm.alerts <- Alert{
            Level:     "WARNING",
            Component: "cache",
            Message:   "Low cache hit rate",
            Timestamp: time.Now(),
            Value:     mpm.metrics.CacheHitRate,
        }
    }
    
    // Check memory usage
    if mpm.metrics.HeapSize > mpm.thresholds.MaxMemoryUsage {
        mpm.alerts <- Alert{
            Level:     "CRITICAL",
            Component: "memory",
            Message:   "High memory usage",
            Timestamp: time.Now(),
            Value:     mpm.metrics.HeapSize,
        }
        runtime.GC() // Force garbage collection
    }
    
    // Check error rate
    if mpm.metrics.SearchThroughput > 0 {
        errorRate := float64(mpm.metrics.SearchErrors) / float64(mpm.metrics.SearchThroughput)
        if errorRate > mpm.thresholds.MaxErrorRate {
            mpm.alerts <- Alert{
                Level:     "CRITICAL",
                Component: "errors",
                Message:   "High error rate",
                Timestamp: time.Now(),
                Value:     errorRate,
            }
        }
    }
}

func (mpm *MemoryPerformanceMonitor) calculateAverageLatency() time.Duration {
    if len(mpm.metrics.SearchLatency) == 0 {
        return 0
    }
    
    var total time.Duration
    for _, latency := range mpm.metrics.SearchLatency {
        total += latency
    }
    
    return total / time.Duration(len(mpm.metrics.SearchLatency))
}

func (mpm *MemoryPerformanceMonitor) GetAlerts() <-chan Alert {
    return mpm.alerts
}

func (mpm *MemoryPerformanceMonitor) GetMetricsReport() map[string]interface{} {
    mpm.metrics.mu.RLock()
    defer mpm.metrics.mu.RUnlock()
    
    return map[string]interface{}{
        "avg_search_latency_ms": mpm.calculateAverageLatency().Milliseconds(),
        "search_throughput":     mpm.metrics.SearchThroughput,
        "search_errors":         mpm.metrics.SearchErrors,
        "cache_hit_rate":        mpm.metrics.CacheHitRate,
        "heap_size_mb":          mpm.metrics.HeapSize / 1024 / 1024,
        "cache_size":            mpm.metrics.CacheSize,
        "cache_evictions":       mpm.metrics.CacheEvictions,
    }
}
```##
# 2. Resource Management and Auto-Scaling

```go
type ResourceManager struct {
    memory          core.Memory
    monitor         *MemoryPerformanceMonitor
    scaler          *AutoScaler
    config          ResourceConfig
}

type ResourceConfig struct {
    MaxMemoryMB        int64
    MaxConcurrency     int
    ScaleUpThreshold   float64
    ScaleDownThreshold float64
    CooldownPeriod     time.Duration
}

type AutoScaler struct {
    currentCapacity int
    maxCapacity     int
    minCapacity     int
    lastScaleTime   time.Time
    cooldownPeriod  time.Duration
}

func NewResourceManager(memory core.Memory, config ResourceConfig) *ResourceManager {
    monitor := NewMemoryPerformanceMonitor(PerformanceThresholds{
        MaxSearchLatency: 2 * time.Second,
        MinCacheHitRate:  0.7,
        MaxMemoryUsage:   config.MaxMemoryMB * 1024 * 1024,
        MaxErrorRate:     0.05,
    })
    
    scaler := &AutoScaler{
        currentCapacity: 10,
        maxCapacity:     100,
        minCapacity:     5,
        cooldownPeriod:  config.CooldownPeriod,
    }
    
    return &ResourceManager{
        memory:  memory,
        monitor: monitor,
        scaler:  scaler,
        config:  config,
    }
}

func (rm *ResourceManager) Start(ctx context.Context) {
    // Start performance monitoring
    go rm.monitor.Start(ctx)
    
    // Start auto-scaling
    go rm.autoScale(ctx)
    
    // Start alert handling
    go rm.handleAlerts(ctx)
}

func (rm *ResourceManager) autoScale(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            rm.evaluateScaling()
        }
    }
}

func (rm *ResourceManager) evaluateScaling() {
    // Check if we're in cooldown period
    if time.Since(rm.scaler.lastScaleTime) < rm.scaler.cooldownPeriod {
        return
    }
    
    metrics := rm.monitor.GetMetricsReport()
    
    // Calculate load metrics
    avgLatency := metrics["avg_search_latency_ms"].(int64)
    cacheHitRate := metrics["cache_hit_rate"].(float64)
    heapSizeMB := metrics["heap_size_mb"].(int64)
    
    // Determine if scaling is needed
    loadScore := rm.calculateLoadScore(avgLatency, cacheHitRate, heapSizeMB)
    
    if loadScore > rm.config.ScaleUpThreshold && rm.scaler.currentCapacity < rm.scaler.maxCapacity {
        rm.scaleUp()
    } else if loadScore < rm.config.ScaleDownThreshold && rm.scaler.currentCapacity > rm.scaler.minCapacity {
        rm.scaleDown()
    }
}

func (rm *ResourceManager) calculateLoadScore(latency int64, cacheHitRate float64, heapSizeMB int64) float64 {
    // Normalize metrics to 0-1 scale and combine
    latencyScore := float64(latency) / 2000.0  // Normalize to 2 second max
    cacheScore := 1.0 - cacheHitRate           // Lower hit rate = higher load
    memoryScore := float64(heapSizeMB) / float64(rm.config.MaxMemoryMB)
    
    // Weighted combination
    return (latencyScore*0.4 + cacheScore*0.3 + memoryScore*0.3)
}

func (rm *ResourceManager) scaleUp() {
    newCapacity := int(float64(rm.scaler.currentCapacity) * 1.5)
    if newCapacity > rm.scaler.maxCapacity {
        newCapacity = rm.scaler.maxCapacity
    }
    
    log.Printf("Scaling up from %d to %d", rm.scaler.currentCapacity, newCapacity)
    rm.scaler.currentCapacity = newCapacity
    rm.scaler.lastScaleTime = time.Now()
    
    // Trigger garbage collection to free memory
    runtime.GC()
}

func (rm *ResourceManager) scaleDown() {
    newCapacity := int(float64(rm.scaler.currentCapacity) * 0.8)
    if newCapacity < rm.scaler.minCapacity {
        newCapacity = rm.scaler.minCapacity
    }
    
    log.Printf("Scaling down from %d to %d", rm.scaler.currentCapacity, newCapacity)
    rm.scaler.currentCapacity = newCapacity
    rm.scaler.lastScaleTime = time.Now()
}

func (rm *ResourceManager) handleAlerts(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case alert := <-rm.monitor.GetAlerts():
            rm.processAlert(alert)
        }
    }
}

func (rm *ResourceManager) processAlert(alert Alert) {
    log.Printf("ALERT [%s] %s: %s (Value: %v)", 
        alert.Level, alert.Component, alert.Message, alert.Value)
    
    switch alert.Component {
    case "memory":
        if alert.Level == "CRITICAL" {
            // Force garbage collection
            runtime.GC()
            // Consider scaling up
            rm.scaleUp()
        }
    case "search":
        if alert.Level == "WARNING" {
            // Consider increasing cache size or scaling up
            log.Printf("Consider optimizing search performance")
        }
    case "cache":
        if alert.Level == "WARNING" {
            // Consider increasing cache size or TTL
            log.Printf("Consider optimizing cache configuration")
        }
    }
}
```## Pro
duction Optimization Example

### Complete Optimized Memory System

```go
func main() {
    // Create optimized memory configuration
    config := createProductionConfig()
    
    // Create memory instance
    memory, err := core.NewMemory(config)
    if err != nil {
        log.Fatalf("Failed to create optimized memory: %v", err)
    }
    defer memory.Close()
    
    // Create optimized memory system with caching
    cacheConfig := CacheConfig{
        L1Size:   1000,
        L1TTL:    5 * time.Minute,
        L2Size:   10000,
        L2TTL:    30 * time.Minute,
        L3TTL:    2 * time.Hour,
        EnableL1: true,
        EnableL2: true,
        EnableL3: true,
    }
    
    optimizedSystem := NewOptimizedMemorySystem(memory, cacheConfig)
    
    // Create resource manager
    resourceConfig := ResourceConfig{
        MaxMemoryMB:        4096, // 4GB
        MaxConcurrency:     50,
        ScaleUpThreshold:   0.7,
        ScaleDownThreshold: 0.3,
        CooldownPeriod:     5 * time.Minute,
    }
    
    resourceManager := NewResourceManager(memory, resourceConfig)
    
    // Start monitoring and resource management
    ctx := context.Background()
    resourceManager.Start(ctx)
    
    // Demonstrate optimized operations
    err = demonstrateOptimizedOperations(optimizedSystem)
    if err != nil {
        log.Fatalf("Demonstration failed: %v", err)
    }
    
    // Monitor performance
    go monitorPerformance(resourceManager)
    
    // Keep running
    select {}
}

func demonstrateOptimizedOperations(system *OptimizedMemorySystem) error {
    ctx := context.Background()
    
    // Populate with sample data
    documents := []core.Document{
        {
            ID:      "opt-doc-1",
            Title:   "Optimization Techniques",
            Content: "Advanced optimization techniques for memory systems include caching, indexing, and resource management...",
            Source:  "optimization-guide.md",
            Type:    core.DocumentTypeMarkdown,
            Tags:    []string{"optimization", "performance", "caching"},
            CreatedAt: time.Now(),
        },
        // Add more documents...
    }
    
    // Ingest documents
    for _, doc := range documents {
        err := system.memory.IngestDocument(ctx, doc)
        if err != nil {
            log.Printf("Failed to ingest document: %v", err)
        }
    }
    
    // Perform optimized searches
    queries := []string{
        "optimization techniques",
        "caching strategies",
        "performance monitoring",
        "resource management",
    }
    
    for _, query := range queries {
        start := time.Now()
        
        results, err := system.SearchKnowledge(ctx, query,
            core.WithLimit(10),
            core.WithScoreThreshold(0.7),
            core.WithTags([]string{"optimization"}),
        )
        
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("Search failed for '%s': %v", query, err)
            continue
        }
        
        fmt.Printf("Query: '%s' - %d results in %v\n", 
            query, len(results), duration)
    }
    
    return nil
}

func monitorPerformance(rm *ResourceManager) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        metrics := rm.monitor.GetMetricsReport()
        
        fmt.Printf("=== Performance Metrics ===\n")
        fmt.Printf("Average Search Latency: %d ms\n", metrics["avg_search_latency_ms"])
        fmt.Printf("Search Throughput: %d\n", metrics["search_throughput"])
        fmt.Printf("Cache Hit Rate: %.2f%%\n", metrics["cache_hit_rate"].(float64)*100)
        fmt.Printf("Heap Size: %d MB\n", metrics["heap_size_mb"])
        fmt.Printf("Current Capacity: %d\n", rm.scaler.currentCapacity)
        fmt.Println()
    }
}
```

## Best Practices and Recommendations

### 1. Configuration Optimization

- **Embedding Caching**: Always enable `CacheEmbeddings: true` for production
- **Batch Sizes**: Use larger `MaxBatchSize` (200-300) for better API efficiency
- **Chunk Sizes**: Use larger chunks (1500-2000) for better context
- **Score Thresholds**: Use higher thresholds (0.75-0.8) for better quality
- **Connection Pooling**: Configure appropriate pool sizes in connection string

### 2. Embedding Service Performance Tuning

```go
type EmbeddingServiceOptimizer struct {
    config      core.EmbeddingConfig
    rateLimiter *RateLimiter
    batcher     *EmbeddingBatcher
    metrics     *EmbeddingMetrics
}

type EmbeddingMetrics struct {
    RequestCount    int64
    ErrorCount      int64
    AverageLatency  time.Duration
    CacheHitRate    float64
    BatchEfficiency float64
    mu              sync.RWMutex
}

type EmbeddingBatcher struct {
    batch     []string
    batchSize int
    timeout   time.Duration
    mu        sync.Mutex
}

func NewEmbeddingServiceOptimizer(config core.EmbeddingConfig) *EmbeddingServiceOptimizer {
    return &EmbeddingServiceOptimizer{
        config:      config,
        rateLimiter: NewRateLimiter(config.MaxBatchSize, time.Minute),
        batcher:     NewEmbeddingBatcher(config.MaxBatchSize, 100*time.Millisecond),
        metrics:     &EmbeddingMetrics{},
    }
}

func (eso *EmbeddingServiceOptimizer) OptimizeEmbeddingConfig() core.EmbeddingConfig {
    optimized := eso.config
    
    // Optimize based on provider
    switch eso.config.Provider {
    case "openai":
        optimized.MaxBatchSize = 200      // OpenAI supports large batches
        optimized.TimeoutSeconds = 60     // Longer timeout for large batches
        optimized.CacheEmbeddings = true  // Always cache for OpenAI
        
    case "ollama":
        optimized.MaxBatchSize = 32       // Smaller batches for local processing
        optimized.TimeoutSeconds = 120    // Longer timeout for local processing
        optimized.CacheEmbeddings = true  // Cache to avoid recomputation
        
    case "dummy":
        optimized.MaxBatchSize = 1000     // No API limits for dummy
        optimized.TimeoutSeconds = 5      // Fast for testing
        optimized.CacheEmbeddings = false // No need to cache dummy embeddings
    }
    
    return optimized
}

func (eso *EmbeddingServiceOptimizer) GetOptimizedConfig(workloadType string) core.EmbeddingConfig {
    base := eso.OptimizeEmbeddingConfig()
    
    switch workloadType {
    case "high-throughput":
        base.MaxBatchSize = 300
        base.TimeoutSeconds = 90
        base.CacheEmbeddings = true
        
    case "low-latency":
        base.MaxBatchSize = 50
        base.TimeoutSeconds = 30
        base.CacheEmbeddings = true
        
    case "cost-optimized":
        base.Provider = "ollama"
        base.Model = "mxbai-embed-large"
        base.MaxBatchSize = 64
        base.TimeoutSeconds = 180
        base.CacheEmbeddings = true
        
    case "quality-focused":
        base.Provider = "openai"
        base.Model = "text-embedding-3-large"
        base.MaxBatchSize = 100
        base.TimeoutSeconds = 120
        base.CacheEmbeddings = true
    }
    
    return base
}
```

### 3. Advanced Caching Patterns

```go
type AdvancedCacheManager struct {
    embeddingCache  *LRUCache
    searchCache     *TTLCache
    contextCache    *AdaptiveCache
    precomputeCache *PrecomputeCache
}

type PrecomputeCache struct {
    commonQueries map[string][]core.KnowledgeResult
    mu            sync.RWMutex
    updateTicker  *time.Ticker
}

func NewPrecomputeCache(memory core.Memory) *PrecomputeCache {
    pc := &PrecomputeCache{
        commonQueries: make(map[string][]core.KnowledgeResult),
        updateTicker:  time.NewTicker(1 * time.Hour),
    }
    
    // Precompute common queries
    go pc.precomputeCommonQueries(memory)
    
    return pc
}

func (pc *PrecomputeCache) precomputeCommonQueries(memory core.Memory) {
    commonQueries := []string{
        "machine learning",
        "artificial intelligence",
        "deep learning",
        "neural networks",
        "data science",
        "python programming",
        "algorithms",
        "optimization",
    }
    
    for range pc.updateTicker.C {
        ctx := context.Background()
        
        for _, query := range commonQueries {
            results, err := memory.SearchKnowledge(ctx, query,
                core.WithLimit(20),
                core.WithScoreThreshold(0.7),
            )
            if err != nil {
                log.Printf("Failed to precompute query '%s': %v", query, err)
                continue
            }
            
            pc.mu.Lock()
            pc.commonQueries[query] = results
            pc.mu.Unlock()
        }
    }
}

func (pc *PrecomputeCache) Get(query string) ([]core.KnowledgeResult, bool) {
    pc.mu.RLock()
    defer pc.mu.RUnlock()
    
    results, exists := pc.commonQueries[query]
    return results, exists
}
```

### 4. Performance Monitoring

- **Latency Tracking**: Monitor search and embedding generation latencies
- **Cache Performance**: Track hit rates and optimize cache sizes
- **Resource Usage**: Monitor memory, CPU, and database connections
- **Error Rates**: Track and alert on error rates and failures
- **Throughput**: Monitor requests per second and concurrent operations
- **Embedding Efficiency**: Track batch utilization and API costs

### 5. Database Optimization Strategies

```go
type DatabaseOptimizer struct {
    connectionPool *ConnectionPool
    indexManager   *IndexManager
    queryOptimizer *QueryOptimizer
}

type ConnectionPool struct {
    maxConnections     int
    minConnections     int
    connectionLifetime time.Duration
    idleTimeout        time.Duration
}

func OptimizeDatabaseConnection() string {
    // Optimized connection string for production
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?"+
            "pool_max_conns=100&"+           // High connection limit
            "pool_min_conns=20&"+            // Maintain minimum connections
            "pool_max_conn_lifetime=1h&"+    // Rotate connections
            "pool_max_conn_idle_time=30m&"+  // Close idle connections
            "pool_health_check_period=1m&"+  // Regular health checks
            "sslmode=require&"+              // Security
            "application_name=agenticgokit", // Identify application
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
    )
}

type IndexOptimizer struct {
    memory core.Memory
}

func (io *IndexOptimizer) OptimizeIndexes(ctx context.Context) error {
    // This would be implemented based on the specific vector database
    // For pgvector, you might optimize HNSW parameters
    // For Weaviate, you might optimize shard configuration
    
    log.Printf("Optimizing vector database indexes...")
    
    // Example optimization strategies:
    // 1. Analyze query patterns
    // 2. Optimize index parameters (m, ef_construction for HNSW)
    // 3. Partition data by usage patterns
    // 4. Implement proper maintenance schedules
    
    return nil
}

func (io *IndexOptimizer) AnalyzeQueryPatterns(ctx context.Context) (*QueryAnalysis, error) {
    // Analyze common query patterns to optimize indexes
    analysis := &QueryAnalysis{
        CommonQueries:    []string{},
        AverageQuerySize: 0,
        PopularTags:      []string{},
        QueryFrequency:   make(map[string]int),
    }
    
    // This would analyze actual query logs
    // For demonstration, we'll use sample patterns
    
    return analysis, nil
}

type QueryAnalysis struct {
    CommonQueries    []string
    AverageQuerySize int
    PopularTags      []string
    QueryFrequency   map[string]int
}
```

### 6. Scaling Strategies

- **Vertical Scaling**: Increase memory and CPU for single-node performance
- **Horizontal Scaling**: Use multiple instances with load balancing
- **Database Scaling**: Use read replicas and connection pooling
- **Cache Scaling**: Implement distributed caching with Redis
- **Auto-Scaling**: Implement automatic scaling based on load metrics
- **Embedding Service Scaling**: Use multiple API keys and load balancing
- **Index Optimization**: Regularly optimize vector database indexes

## Conclusion

Memory optimization is essential for production-ready agent systems. Key takeaways:

- Use optimized `AgentMemoryConfig` with production-appropriate settings
- Implement multi-level caching for better performance
- Monitor performance metrics and implement alerting
- Use resource management and auto-scaling for reliability
- Optimize embedding and search configurations for your use case

Proper optimization ensures your agents can handle production workloads efficiently while maintaining fast response times and high accuracy.

## Next Steps

🎉 **Congratulations!** You've completed the Memory Systems tutorial series.

### 🚀 **Production Deployment**
- **[Production Deployment](../../guides/production-deployment.md)** - Deploy optimized systems
- Take your optimized memory system to production with confidence

### 📊 **Advanced Operations**
- **[Monitoring and Observability](../../guides/monitoring.md)** - Advanced monitoring patterns
- **[Scaling Patterns](../../guides/scaling.md)** - Learn advanced scaling techniques

::: success Memory Systems Mastery
🏆 **Tutorial Series Complete!**  
✅ **Foundation**: Basic memory operations  
✅ **Storage**: Production vector databases  
✅ **Content**: Document ingestion pipeline  
✅ **Intelligence**: RAG implementation  
✅ **Scale**: Enterprise knowledge bases  
✅ **Performance**: Optimization and monitoring  

**You're now ready to build production-grade memory systems!**
:::

## Complete Learning Journey

You've mastered the entire memory systems stack:

1. **[Basic Memory Operations](basic-memory.md)** - ✅ Memory interface fundamentals
2. **[Vector Databases](vector-databases.md)** - ✅ Production storage backends
3. **[Document Ingestion](document-ingestion.md)** - ✅ Content processing pipeline
4. **[RAG Implementation](rag-implementation.md)** - ✅ Intelligent retrieval systems
5. **[Knowledge Bases](knowledge-bases.md)** - ✅ Enterprise search patterns
6. **[Memory Optimization](memory-optimization.md)** - ✅ **Complete!**

## What's Next?

- **Build Production Systems**: Apply your knowledge to real-world projects
- **Contribute**: Share your experience with the AgenticGoKit community
- **Explore**: Investigate advanced patterns and integrations
- **Scale**: Deploy enterprise-grade memory systems

## Further Reading

- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Go Performance Optimization](https://github.com/dgryski/go-perfbook)
- [Vector Database Benchmarks](https://github.com/erikbern/ann-benchmarks)