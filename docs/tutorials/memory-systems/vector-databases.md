# Vector Databases in AgenticGoKit

## Overview

Vector databases enable sophisticated similarity search and retrieval-augmented generation (RAG) by storing and querying high-dimensional embeddings. This tutorial covers setting up and using vector databases with AgenticGoKit, including pgvector and Weaviate integration.

Vector databases are essential for production-ready memory systems that need to handle large amounts of data with fast, semantically-aware search capabilities.

## Prerequisites

- Understanding of [Basic Memory Operations](basic-memory.md)
- Knowledge of vector embeddings and similarity search
- Basic database administration skills
- Familiarity with Docker (for setup examples)

## Vector Database Concepts

### What are Vector Embeddings?

Vector embeddings are numerical representations of text, images, or other data that capture semantic meaning in high-dimensional space:

```
"The cat sat on the mat" → [0.1, -0.3, 0.8, ..., 0.2] (1536 dimensions)
"A feline rested on the rug" → [0.2, -0.2, 0.7, ..., 0.3] (similar vector)
```

### Similarity Search

Vector databases use distance metrics to find similar content:

- **Cosine Similarity**: Measures angle between vectors
- **Euclidean Distance**: Measures straight-line distance
- **Dot Product**: Measures vector alignment

## pgvector Setup and Configuration

### 1. Database Setup with Docker

```bash
# Create docker-compose.yml for pgvector
version: '3.8'
services:
  postgres:
    image: pgvector/pgvector:pg15
    environment:
      POSTGRES_DB: agentdb
      POSTGRES_USER: agent_user
      POSTGRES_PASSWORD: agent_pass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql

volumes:
  postgres_data:
```

```sql
-- init.sql
CREATE EXTENSION IF NOT EXISTS vector;

-- Create memory table
CREATE TABLE agent_memory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    embedding vector(1536), -- OpenAI embedding dimensions
    metadata JSONB,
    session_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_agent_memory_content_type ON agent_memory(content_type);
CREATE INDEX idx_agent_memory_session_id ON agent_memory(session_id);
CREATE INDEX idx_agent_memory_created_at ON agent_memory(created_at);
CREATE INDEX idx_agent_memory_embedding ON agent_memory USING ivfflat (embedding vector_cosine_ops);

-- Create conversation history table
CREATE TABLE conversation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_conversation_session_id ON conversation_history(session_id);
CREATE INDEX idx_conversation_timestamp ON conversation_history(timestamp);
```

### 2. pgvector Configuration

```go
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Configure pgvector memory
    config := core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://agent_user:agent_pass@localhost:5432/agentdb?sslmode=disable",
        EnableRAG:  true,
        Dimensions: 1536, // OpenAI embedding dimensions
        ChunkSize:  1000,
        ChunkOverlap: 200,
        Embedding: core.EmbeddingConfig{
            Provider:   "openai",
            Model:      "text-embedding-3-small",
            APIKey:     os.Getenv("OPENAI_API_KEY"),
            Dimensions: 1536,
            BatchSize:  100,
        },
        Options: map[string]interface{}{
            "max_connections": 10,
            "timeout":         "30s",
            "retry_attempts":  3,
            "index_type":      "ivfflat", // or "hnsw"
            "index_lists":     100,       // for ivfflat
        },
    }
    
    // Create memory instance
    memory, err := core.NewMemory(config)
    if err != nil {
        log.Fatalf("Failed to create pgvector memory: %v", err)
    }
    
    // Test the connection
    ctx := context.Background()
    stats, err := memory.GetStats(ctx)
    if err != nil {
        log.Fatalf("Failed to get stats: %v", err)
    }
    
    log.Printf("Connected to pgvector: %d items stored", stats.ItemCount)
}
```

### 3. Advanced pgvector Operations

```go
func demonstratePgvectorFeatures(memory core.Memory) error {
    ctx := context.Background()
    
    // Store documents with rich metadata
    documents := []struct {
        content  string
        docType  string
        metadata map[string]string
    }{
        {
            content: "Machine learning is a subset of artificial intelligence that enables computers to learn from data.",
            docType: "definition",
            metadata: map[string]string{
                "topic":      "machine-learning",
                "difficulty": "beginner",
                "source":     "textbook",
            },
        },
        {
            content: "Neural networks are computing systems inspired by biological neural networks.",
            docType: "definition",
            metadata: map[string]string{
                "topic":      "neural-networks",
                "difficulty": "intermediate",
                "source":     "research-paper",
            },
        },
        {
            content: "Deep learning uses neural networks with multiple layers to model complex patterns.",
            docType: "definition",
            metadata: map[string]string{
                "topic":      "deep-learning",
                "difficulty": "advanced",
                "source":     "textbook",
            },
        },
    }
    
    // Store documents
    for _, doc := range documents {
        err := memory.Store(ctx, doc.content, doc.docType,
            core.WithMetadata(doc.metadata),
            core.WithTimestamp(time.Now()),
        )
        if err != nil {
            return fmt.Errorf("failed to store document: %w", err)
        }
    }
    
    // Perform semantic search
    results, err := memory.Search(ctx, "What is AI and how does it learn?",
        core.WithLimit(3),
        core.WithScoreThreshold(0.7),
        core.WithMetadataFilter(map[string]string{
            "difficulty": "beginner",
        }),
    )
    if err != nil {
        return fmt.Errorf("search failed: %w", err)
    }
    
    fmt.Printf("Found %d relevant documents:\n", len(results))
    for _, result := range results {
        fmt.Printf("- %s (Score: %.3f)\n", result.Content, result.Score)
        fmt.Printf("  Topic: %s, Difficulty: %s\n", 
            result.Metadata["topic"], result.Metadata["difficulty"])
    }
    
    return nil
}
```

## Weaviate Integration

### 1. Weaviate Setup with Docker

```bash
# docker-compose.yml for Weaviate
version: '3.4'
services:
  weaviate:
    command:
    - --host
    - 0.0.0.0
    - --port
    - '8080'
    - --scheme
    - http
    image: semitechnologies/weaviate:1.22.4
    ports:
    - 8080:8080
    restart: on-failure:0
    environment:
      QUERY_DEFAULTS_LIMIT: 25
      AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'true'
      PERSISTENCE_DATA_PATH: '/var/lib/weaviate'
      DEFAULT_VECTORIZER_MODULE: 'none'
      ENABLE_MODULES: 'text2vec-openai,generative-openai'
      CLUSTER_HOSTNAME: 'node1'
    volumes:
      - weaviate_data:/var/lib/weaviate

volumes:
  weaviate_data:
```

### 2. Weaviate Configuration

```go
func setupWeaviateMemory() (core.Memory, error) {
    config := core.AgentMemoryConfig{
        Provider:   "weaviate",
        Connection: "http://localhost:8080",
        EnableRAG:  true,
        Embedding: core.EmbeddingConfig{
            Provider:   "openai",
            Model:      "text-embedding-3-small",
            APIKey:     os.Getenv("OPENAI_API_KEY"),
            Dimensions: 1536,
            BatchSize:  100,
        },
        Options: map[string]interface{}{
            "class_name":     "AgentMemory",
            "timeout":        "30s",
            "retry_attempts": 3,
            "batch_size":     100,
        },
    }
    
    memory, err := core.NewMemory(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create Weaviate memory: %w", err)
    }
    
    return memory, nil
}
```

### 3. Advanced Weaviate Features

```go
func demonstrateWeaviateFeatures(memory core.Memory) error {
    ctx := context.Background()
    
    // Store multi-modal content
    content := []struct {
        text     string
        docType  string
        metadata map[string]string
    }{
        {
            text:    "A red sports car driving on a mountain road",
            docType: "image-description",
            metadata: map[string]string{
                "category": "automotive",
                "color":    "red",
                "setting":  "mountain",
            },
        },
        {
            text:    "A blue sedan parked in a city street",
            docType: "image-description",
            metadata: map[string]string{
                "category": "automotive",
                "color":    "blue",
                "setting":  "urban",
            },
        },
    }
    
    // Store content
    for _, item := range content {
        err := memory.Store(ctx, item.text, item.docType,
            core.WithMetadata(item.metadata),
        )
        if err != nil {
            return fmt.Errorf("failed to store: %w", err)
        }
    }
    
    // Perform complex search with filters
    results, err := memory.Search(ctx, "vehicle in urban environment",
        core.WithLimit(5),
        core.WithMetadataFilter(map[string]string{
            "category": "automotive",
            "setting":  "urban",
        }),
    )
    if err != nil {
        return fmt.Errorf("search failed: %w", err)
    }
    
    fmt.Printf("Found %d matching items:\n", len(results))
    for _, result := range results {
        fmt.Printf("- %s (Score: %.3f)\n", result.Content, result.Score)
    }
    
    return nil
}
```

## Embedding Strategies

### 1. OpenAI Embeddings

```go
func setupOpenAIEmbeddings() core.EmbeddingConfig {
    return core.EmbeddingConfig{
        Provider:   "openai",
        Model:      "text-embedding-3-small", // or text-embedding-3-large
        APIKey:     os.Getenv("OPENAI_API_KEY"),
        Dimensions: 1536, // 3072 for large model
        BatchSize:  100,
        Options: map[string]string{
            "user": "agenticgokit-user", // for usage tracking
        },
    }
}
```

### 2. Hugging Face Embeddings

```go
func setupHuggingFaceEmbeddings() core.EmbeddingConfig {
    return core.EmbeddingConfig{
        Provider:   "huggingface",
        Model:      "sentence-transformers/all-MiniLM-L6-v2",
        APIKey:     os.Getenv("HUGGINGFACE_API_KEY"), // optional for hosted inference
        Dimensions: 384,
        BatchSize:  32,
        Options: map[string]string{
            "normalize_embeddings": "true",
            "pooling_mode":        "mean",
        },
    }
}
```

### 3. Local Embeddings

```go
func setupLocalEmbeddings() core.EmbeddingConfig {
    return core.EmbeddingConfig{
        Provider:   "local",
        Model:      "all-MiniLM-L6-v2",
        Dimensions: 384,
        BatchSize:  16,
        Options: map[string]string{
            "model_path":    "./models/sentence-transformer",
            "device":        "cpu", // or "cuda"
            "max_seq_length": "512",
        },
    }
}
```

## Performance Optimization

### 1. Index Configuration

```go
// pgvector index optimization
func optimizePgvectorIndexes(db *sql.DB) error {
    // Create HNSW index for better performance
    _, err := db.Exec(`
        CREATE INDEX CONCURRENTLY idx_agent_memory_embedding_hnsw 
        ON agent_memory USING hnsw (embedding vector_cosine_ops)
        WITH (m = 16, ef_construction = 64);
    `)
    if err != nil {
        return fmt.Errorf("failed to create HNSW index: %w", err)
    }
    
    // Update table statistics
    _, err = db.Exec("ANALYZE agent_memory;")
    if err != nil {
        return fmt.Errorf("failed to analyze table: %w", err)
    }
    
    return nil
}
```

### 2. Batch Operations

```go
func performBatchOperations(memory core.Memory) error {
    ctx := context.Background()
    
    // Prepare batch data
    documents := make([]core.Document, 1000)
    for i := 0; i < 1000; i++ {
        documents[i] = core.Document{
            Content:     fmt.Sprintf("Document %d content", i),
            ContentType: "batch-document",
            Metadata: map[string]string{
                "batch_id": "batch-001",
                "index":    fmt.Sprintf("%d", i),
            },
        }
    }
    
    // Batch store operation
    err := memory.StoreBatch(ctx, documents,
        core.WithBatchSize(100),
        core.WithConcurrency(4),
    )
    if err != nil {
        return fmt.Errorf("batch store failed: %w", err)
    }
    
    fmt.Printf("Successfully stored %d documents in batch\n", len(documents))
    return nil
}
```

### 3. Connection Pooling

```go
type OptimizedMemoryConfig struct {
    core.AgentMemoryConfig
    ConnectionPool struct {
        MaxConnections     int           `yaml:"max_connections"`
        MaxIdleConnections int           `yaml:"max_idle_connections"`
        ConnectionTimeout  time.Duration `yaml:"connection_timeout"`
        IdleTimeout        time.Duration `yaml:"idle_timeout"`
    } `yaml:"connection_pool"`
}

func createOptimizedMemory() (core.Memory, error) {
    config := OptimizedMemoryConfig{
        AgentMemoryConfig: core.AgentMemoryConfig{
            Provider:   "pgvector",
            Connection: "postgres://user:pass@localhost:5432/agentdb",
            EnableRAG:  true,
            Dimensions: 1536,
        },
    }
    
    // Configure connection pool
    config.ConnectionPool.MaxConnections = 20
    config.ConnectionPool.MaxIdleConnections = 5
    config.ConnectionPool.ConnectionTimeout = 30 * time.Second
    config.ConnectionPool.IdleTimeout = 5 * time.Minute
    
    return core.NewMemoryWithConfig(config)
}
```

## Memory Monitoring and Metrics

### 1. Performance Metrics

```go
type VectorDBMetrics struct {
    SearchLatency    []time.Duration
    IndexSize        int64
    QueryThroughput  float64
    CacheHitRate     float64
    mu               sync.RWMutex
}

func (m *VectorDBMetrics) RecordSearch(duration time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.SearchLatency = append(m.SearchLatency, duration)
    
    // Keep only recent measurements
    if len(m.SearchLatency) > 1000 {
        m.SearchLatency = m.SearchLatency[len(m.SearchLatency)-1000:]
    }
}

func (m *VectorDBMetrics) GetAverageLatency() time.Duration {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if len(m.SearchLatency) == 0 {
        return 0
    }
    
    var total time.Duration
    for _, latency := range m.SearchLatency {
        total += latency
    }
    
    return total / time.Duration(len(m.SearchLatency))
}
```

### 2. Health Monitoring

```go
func monitorVectorDBHealth(memory core.Memory) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        
        // Check basic connectivity
        stats, err := memory.GetStats(ctx)
        if err != nil {
            log.Printf("Health check failed: %v", err)
            cancel()
            continue
        }
        
        // Check search performance
        start := time.Now()
        _, err = memory.Search(ctx, "health check query", core.WithLimit(1))
        searchLatency := time.Since(start)
        
        if err != nil {
            log.Printf("Search health check failed: %v", err)
        } else if searchLatency > 1*time.Second {
            log.Printf("Search latency high: %v", searchLatency)
        }
        
        // Log health status
        log.Printf("Vector DB Health: %d items, search latency: %v", 
            stats.ItemCount, searchLatency)
        
        cancel()
    }
}
```

## Production Deployment

### 1. High Availability Setup

```yaml
# docker-compose.prod.yml for pgvector HA
version: '3.8'
services:
  postgres-primary:
    image: pgvector/pgvector:pg15
    environment:
      POSTGRES_DB: agentdb
      POSTGRES_USER: agent_user
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_REPLICATION_USER: replicator
      POSTGRES_REPLICATION_PASSWORD: ${REPLICATION_PASSWORD}
    volumes:
      - postgres_primary_data:/var/lib/postgresql/data
      - ./postgresql.conf:/etc/postgresql/postgresql.conf
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
    
  postgres-replica:
    image: pgvector/pgvector:pg15
    environment:
      PGUSER: agent_user
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_MASTER_SERVICE: postgres-primary
      POSTGRES_REPLICATION_USER: replicator
      POSTGRES_REPLICATION_PASSWORD: ${REPLICATION_PASSWORD}
    volumes:
      - postgres_replica_data:/var/lib/postgresql/data
    depends_on:
      - postgres-primary

volumes:
  postgres_primary_data:
  postgres_replica_data:
```

### 2. Backup and Recovery

```go
func setupBackupStrategy(memory core.Memory) error {
    // Create backup scheduler
    scheduler := cron.New()
    
    // Daily full backup
    scheduler.AddFunc("0 2 * * *", func() {
        err := performFullBackup(memory)
        if err != nil {
            log.Printf("Full backup failed: %v", err)
        }
    })
    
    // Hourly incremental backup
    scheduler.AddFunc("0 * * * *", func() {
        err := performIncrementalBackup(memory)
        if err != nil {
            log.Printf("Incremental backup failed: %v", err)
        }
    })
    
    scheduler.Start()
    return nil
}

func performFullBackup(memory core.Memory) error {
    ctx := context.Background()
    
    // Export all data
    backupFile := fmt.Sprintf("backup-full-%s.json", 
        time.Now().Format("2006-01-02-15-04-05"))
    
    data, err := memory.ExportAll(ctx)
    if err != nil {
        return fmt.Errorf("export failed: %w", err)
    }
    
    // Write to file
    err = ioutil.WriteFile(backupFile, data, 0644)
    if err != nil {
        return fmt.Errorf("write backup failed: %w", err)
    }
    
    log.Printf("Full backup completed: %s", backupFile)
    return nil
}
```

## Troubleshooting Common Issues

### 1. Performance Issues

```go
func diagnosePerfomanceIssues(memory core.Memory) {
    ctx := context.Background()
    
    // Check index usage
    stats, err := memory.GetStats(ctx)
    if err != nil {
        log.Printf("Failed to get stats: %v", err)
        return
    }
    
    log.Printf("Memory Stats:")
    log.Printf("  Items: %d", stats.ItemCount)
    log.Printf("  Size: %d MB", stats.SizeBytes/1024/1024)
    log.Printf("  Index Size: %d MB", stats.IndexSizeBytes/1024/1024)
    
    // Test search performance
    queries := []string{
        "machine learning algorithms",
        "neural network architecture",
        "data processing pipeline",
    }
    
    for _, query := range queries {
        start := time.Now()
        results, err := memory.Search(ctx, query, core.WithLimit(10))
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("Query failed: %s - %v", query, err)
        } else {
            log.Printf("Query: %s - %d results in %v", 
                query, len(results), duration)
        }
    }
}
```

### 2. Connection Issues

```go
func handleConnectionIssues(config core.AgentMemoryConfig) core.Memory {
    var memory core.Memory
    var err error
    
    maxRetries := 5
    baseDelay := 1 * time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        memory, err = core.NewMemory(config)
        if err == nil {
            log.Printf("Connected to vector database on attempt %d", attempt+1)
            return memory
        }
        
        delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
        log.Printf("Connection attempt %d failed: %v, retrying in %v", 
            attempt+1, err, delay)
        time.Sleep(delay)
    }
    
    log.Fatalf("Failed to connect after %d attempts: %v", maxRetries, err)
    return nil
}
```

## Best Practices

### 1. Schema Design

```sql
-- Optimized schema for pgvector
CREATE TABLE agent_memory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    embedding vector(1536),
    metadata JSONB,
    session_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Add constraints
    CONSTRAINT content_not_empty CHECK (length(content) > 0),
    CONSTRAINT valid_content_type CHECK (content_type ~ '^[a-z][a-z0-9_-]*$')
);

-- Optimized indexes
CREATE INDEX CONCURRENTLY idx_agent_memory_embedding_hnsw 
ON agent_memory USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

CREATE INDEX idx_agent_memory_content_type_session 
ON agent_memory(content_type, session_id);

CREATE INDEX idx_agent_memory_metadata_gin 
ON agent_memory USING gin (metadata);
```

### 2. Embedding Optimization

```go
func optimizeEmbeddings() core.EmbeddingConfig {
    return core.EmbeddingConfig{
        Provider:   "openai",
        Model:      "text-embedding-3-small", // Good balance of cost/performance
        Dimensions: 1536,
        BatchSize:  100, // Optimize for API rate limits
        Options: map[string]string{
            "max_retries":    "3",
            "retry_delay":    "1s",
            "timeout":        "30s",
            "normalize":      "true", // Normalize embeddings for cosine similarity
        },
    }
}
```

### 3. Query Optimization

```go
func optimizeQueries(memory core.Memory) {
    ctx := context.Background()
    
    // Use appropriate limits
    results, err := memory.Search(ctx, "query",
        core.WithLimit(10), // Don't retrieve more than needed
        core.WithScoreThreshold(0.7), // Filter low-relevance results
    )
    
    // Use metadata filters to reduce search space
    results, err = memory.Search(ctx, "query",
        core.WithContentType("specific-type"),
        core.WithMetadataFilter(map[string]string{
            "category": "relevant-category",
        }),
    )
    
    // Cache frequent queries
    cacheKey := fmt.Sprintf("search:%s", query)
    if cached := getFromCache(cacheKey); cached != nil {
        return cached
    }
    
    results, err = memory.Search(ctx, query)
    if err == nil {
        setCache(cacheKey, results, 5*time.Minute)
    }
}
```

## Conclusion

Vector databases provide the foundation for sophisticated memory systems in AgenticGoKit. Key takeaways:

- Choose the right vector database for your use case (pgvector for SQL integration, Weaviate for specialized features)
- Optimize embeddings and indexes for performance
- Implement proper monitoring and backup strategies
- Use appropriate query patterns and caching

Vector databases enable semantic search, RAG systems, and intelligent memory that can significantly enhance agent capabilities.

## Next Steps

- [RAG Implementation](rag-implementation.md) - Build retrieval-augmented generation systems
- [Knowledge Bases](knowledge-bases.md) - Create comprehensive knowledge systems
- [Memory Optimization](memory-optimization.md) - Advanced performance tuning

## Further Reading

- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [Weaviate Documentation](https://weaviate.io/developers/weaviate)
- [OpenAI Embeddings Guide](https://platform.openai.com/docs/guides/embeddings)