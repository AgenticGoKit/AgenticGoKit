---
title: Vector Databases
description: Learn how to set up and use vector databases with AgenticGoKit, including pgvector and Weaviate integration for production-ready memory systems.
---

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
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"
    "sync"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
    // Note: In production, you would use a cron library like:
    // "github.com/robfig/cron/v3"
)

func main() {
    // Configure pgvector memory with current API
    config := core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://agent_user:agent_pass@localhost:5432/agentdb?sslmode=disable",
        MaxResults: 10,
        Dimensions: 1536,
        AutoEmbed:  true,
        
        // RAG-enhanced settings
        EnableRAG:               true,
        EnableKnowledgeBase:     true,
        KnowledgeMaxResults:     20,
        KnowledgeScoreThreshold: 0.7,
        ChunkSize:               1000,
        ChunkOverlap:            200,
        
        // RAG context assembly settings
        RAGMaxContextTokens: 4000,
        RAGPersonalWeight:   0.3,
        RAGKnowledgeWeight:  0.7,
        RAGIncludeSources:   true,
        
        Embedding: core.EmbeddingConfig{
            Provider:        "openai",
            Model:           "text-embedding-3-small",
            APIKey:          os.Getenv("OPENAI_API_KEY"),
            CacheEmbeddings: true,
            MaxBatchSize:    100,
            TimeoutSeconds:  30,
        },
        
        Documents: core.DocumentConfig{
            AutoChunk:                true,
            SupportedTypes:           []string{"pdf", "txt", "md", "web", "code"},
            MaxFileSize:              "10MB",
            EnableMetadataExtraction: true,
            EnableURLScraping:        true,
        },
        
        Search: core.SearchConfigToml{
            HybridSearch:         true,
            KeywordWeight:        0.3,
            SemanticWeight:       0.7,
            EnableReranking:      false,
            EnableQueryExpansion: false,
        },
    }
    
    // Create memory instance
    memory, err := core.NewMemory(config)
    if err != nil {
        log.Fatalf("Failed to create pgvector memory: %v", err)
    }
    defer memory.Close()
    
    // Test basic operations
    ctx := context.Background()
    
    // Store some test data
    err = memory.Store(ctx, "pgvector is a PostgreSQL extension for vector similarity search", "database", "vector")
    if err != nil {
        log.Fatalf("Failed to store data: %v", err)
    }
    
    // Query the data
    results, err := memory.Query(ctx, "PostgreSQL vector search", 5)
    if err != nil {
        log.Fatalf("Failed to query: %v", err)
    }
    
    fmt.Printf("Found %d results:\n", len(results))
    for _, result := range results {
        fmt.Printf("- %s (Score: %.3f)\n", result.Content, result.Score)
    }
}
```

### 3. Advanced pgvector Operations

```go
func demonstratePgvectorFeatures(memory core.Memory) error {
    ctx := context.Background()
    
    // Store documents using current Document API
    documents := []core.Document{
        {
            ID:      "ml-def-1",
            Title:   "Machine Learning Definition",
            Content: "Machine learning is a subset of artificial intelligence that enables computers to learn from data.",
            Source:  "textbook",
            Type:    core.DocumentTypeText,
            Metadata: map[string]any{
                "topic":      "machine-learning",
                "difficulty": "beginner",
                "category":   "definition",
            },
            Tags:      []string{"ai", "ml", "definition"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "nn-def-1",
            Title:   "Neural Networks Definition",
            Content: "Neural networks are computing systems inspired by biological neural networks.",
            Source:  "research-paper",
            Type:    core.DocumentTypeText,
            Metadata: map[string]any{
                "topic":      "neural-networks",
                "difficulty": "intermediate",
                "category":   "definition",
            },
            Tags:      []string{"ai", "neural-networks", "definition"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "dl-def-1",
            Title:   "Deep Learning Definition",
            Content: "Deep learning uses neural networks with multiple layers to model complex patterns.",
            Source:  "textbook",
            Type:    core.DocumentTypeText,
            Metadata: map[string]any{
                "topic":      "deep-learning",
                "difficulty": "advanced",
                "category":   "definition",
            },
            Tags:      []string{"ai", "deep-learning", "definition"},
            CreatedAt: time.Now(),
        },
    }
    
    // Ingest documents using current API
    err := memory.IngestDocuments(ctx, documents)
    if err != nil {
        return fmt.Errorf("failed to ingest documents: %w", err)
    }
    
    // Perform knowledge search with current API
    results, err := memory.SearchKnowledge(ctx, "What is AI and how does it learn?",
        core.WithLimit(3),
        core.WithScoreThreshold(0.7),
        core.WithTags([]string{"ai", "definition"}),
    )
    if err != nil {
        return fmt.Errorf("knowledge search failed: %w", err)
    }
    
    fmt.Printf("Found %d relevant documents:\n", len(results))
    for _, result := range results {
        fmt.Printf("- %s (Score: %.3f)\n", result.Content, result.Score)
        fmt.Printf("  Source: %s, Document ID: %s\n", result.Source, result.DocumentID)
        if len(result.Tags) > 0 {
            fmt.Printf("  Tags: %v\n", result.Tags)
        }
    }
    
    // Demonstrate hybrid search (personal memory + knowledge base)
    hybridResults, err := memory.SearchAll(ctx, "machine learning concepts",
        core.WithLimit(5),
        core.WithIncludePersonal(true),
        core.WithIncludeKnowledge(true),
    )
    if err != nil {
        return fmt.Errorf("hybrid search failed: %w", err)
    }
    
    fmt.Printf("\nHybrid search results:\n")
    fmt.Printf("Personal Memory: %d results\n", len(hybridResults.PersonalMemory))
    fmt.Printf("Knowledge Base: %d results\n", len(hybridResults.Knowledge))
    fmt.Printf("Total Results: %d\n", hybridResults.TotalResults)
    
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
        MaxResults: 10,
        Dimensions: 1536,
        AutoEmbed:  true,
        
        // RAG-enhanced settings
        EnableRAG:               true,
        EnableKnowledgeBase:     true,
        KnowledgeMaxResults:     20,
        KnowledgeScoreThreshold: 0.7,
        ChunkSize:               1000,
        ChunkOverlap:            200,
        
        // RAG context assembly settings
        RAGMaxContextTokens: 4000,
        RAGPersonalWeight:   0.3,
        RAGKnowledgeWeight:  0.7,
        RAGIncludeSources:   true,
        
        Embedding: core.EmbeddingConfig{
            Provider:        "openai",
            Model:           "text-embedding-3-small",
            APIKey:          os.Getenv("OPENAI_API_KEY"),
            CacheEmbeddings: true,
            MaxBatchSize:    100,
            TimeoutSeconds:  30,
        },
        
        Documents: core.DocumentConfig{
            AutoChunk:                true,
            SupportedTypes:           []string{"pdf", "txt", "md", "web", "code"},
            MaxFileSize:              "10MB",
            EnableMetadataExtraction: true,
            EnableURLScraping:        true,
        },
        
        Search: core.SearchConfigToml{
            HybridSearch:         true,
            KeywordWeight:        0.3,
            SemanticWeight:       0.7,
            EnableReranking:      false,
            EnableQueryExpansion: false,
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
    
    // Create multi-modal documents using current Document API
    documents := []core.Document{
        {
            ID:      "car-red-mountain",
            Title:   "Red Sports Car",
            Content: "A red sports car driving on a mountain road",
            Source:  "image-description",
            Type:    core.DocumentTypeText,
            Metadata: map[string]any{
                "category": "automotive",
                "color":    "red",
                "setting":  "mountain",
                "vehicle_type": "sports-car",
            },
            Tags:      []string{"automotive", "red", "mountain", "sports-car"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "car-blue-urban",
            Title:   "Blue Sedan",
            Content: "A blue sedan parked in a city street",
            Source:  "image-description",
            Type:    core.DocumentTypeText,
            Metadata: map[string]any{
                "category": "automotive",
                "color":    "blue",
                "setting":  "urban",
                "vehicle_type": "sedan",
            },
            Tags:      []string{"automotive", "blue", "urban", "sedan"},
            CreatedAt: time.Now(),
        },
    }
    
    // Ingest documents using current API
    err := memory.IngestDocuments(ctx, documents)
    if err != nil {
        return fmt.Errorf("failed to ingest documents: %w", err)
    }
    
    // Perform complex search with current SearchKnowledge API
    results, err := memory.SearchKnowledge(ctx, "vehicle in urban environment",
        core.WithLimit(5),
        core.WithScoreThreshold(0.5),
        core.WithTags([]string{"automotive", "urban"}),
    )
    if err != nil {
        return fmt.Errorf("knowledge search failed: %w", err)
    }
    
    fmt.Printf("Found %d matching items:\n", len(results))
    for _, result := range results {
        fmt.Printf("- %s (Score: %.3f)\n", result.Content, result.Score)
        fmt.Printf("  Document ID: %s, Source: %s\n", result.DocumentID, result.Source)
        if len(result.Tags) > 0 {
            fmt.Printf("  Tags: %v\n", result.Tags)
        }
    }
    
    // Demonstrate RAG context building
    ragContext, err := memory.BuildContext(ctx, "Tell me about cars in different environments",
        core.WithMaxTokens(2000),
        core.WithKnowledgeWeight(0.8),
        core.WithPersonalWeight(0.2),
        core.WithIncludeSources(true),
    )
    if err != nil {
        return fmt.Errorf("failed to build RAG context: %w", err)
    }
    
    fmt.Printf("\nRAG Context built:\n")
    fmt.Printf("Query: %s\n", ragContext.Query)
    fmt.Printf("Knowledge results: %d\n", len(ragContext.Knowledge))
    fmt.Printf("Personal memory results: %d\n", len(ragContext.PersonalMemory))
    fmt.Printf("Token count: %d\n", ragContext.TokenCount)
    fmt.Printf("Sources: %v\n", ragContext.Sources)
    
    return nil
}
```

## Embedding Strategies

### 1. OpenAI Embeddings

```go
func setupOpenAIEmbeddings() core.EmbeddingConfig {
    return core.EmbeddingConfig{
        Provider:        "openai",
        Model:           "text-embedding-3-small", // or text-embedding-3-large
        APIKey:          os.Getenv("OPENAI_API_KEY"),
        CacheEmbeddings: true,
        MaxBatchSize:    100,
        TimeoutSeconds:  30,
    }
}
```

### 2. Ollama Embeddings

```go
func setupOllamaEmbeddings() core.EmbeddingConfig {
    return core.EmbeddingConfig{
        Provider:        "ollama",
        Model:           "mxbai-embed-large", // or other Ollama embedding models
        BaseURL:         "http://localhost:11434", // Ollama server URL
        CacheEmbeddings: true,
        MaxBatchSize:    32,
        TimeoutSeconds:  60, // Ollama might be slower
    }
}
```

### 3. Dummy Embeddings (for testing)

```go
func setupDummyEmbeddings() core.EmbeddingConfig {
    return core.EmbeddingConfig{
        Provider:        "dummy",
        Model:           "dummy-model",
        CacheEmbeddings: false, // No need to cache dummy embeddings
        MaxBatchSize:    100,
        TimeoutSeconds:  5,
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
    
    // Prepare batch data using current Document structure
    documents := make([]core.Document, 100) // Smaller batch for example
    for i := 0; i < 100; i++ {
        documents[i] = core.Document{
            ID:      fmt.Sprintf("batch-doc-%d", i),
            Title:   fmt.Sprintf("Batch Document %d", i),
            Content: fmt.Sprintf("This is the content of document %d in the batch operation", i),
            Source:  "batch-operation",
            Type:    core.DocumentTypeText,
            Metadata: map[string]any{
                "batch_id": "batch-001",
                "index":    i,
                "category": "batch-document",
            },
            Tags:      []string{"batch", "document", fmt.Sprintf("doc-%d", i)},
            CreatedAt: time.Now(),
        }
    }
    
    // Batch ingest operation using current API
    err := memory.IngestDocuments(ctx, documents)
    if err != nil {
        return fmt.Errorf("batch ingest failed: %w", err)
    }
    
    fmt.Printf("Successfully ingested %d documents in batch\n", len(documents))
    
    // Test batch search
    results, err := memory.SearchKnowledge(ctx, "batch document content",
        core.WithLimit(10),
        core.WithTags([]string{"batch"}),
    )
    if err != nil {
        return fmt.Errorf("batch search failed: %w", err)
    }
    
    fmt.Printf("Found %d documents from batch search\n", len(results))
    return nil
}
```

### 3. Optimized Memory Configuration

```go
func createOptimizedMemory() (core.Memory, error) {
    config := core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://user:pass@localhost:5432/agentdb?pool_max_conns=20&pool_min_conns=5",
        MaxResults: 10,
        Dimensions: 1536,
        AutoEmbed:  true,
        
        // Optimized RAG settings
        EnableRAG:               true,
        EnableKnowledgeBase:     true,
        KnowledgeMaxResults:     50, // Higher for better context
        KnowledgeScoreThreshold: 0.75, // Higher threshold for quality
        ChunkSize:               1500, // Larger chunks for better context
        ChunkOverlap:            300,  // More overlap for continuity
        
        // Optimized RAG context settings
        RAGMaxContextTokens: 8000, // Larger context window
        RAGPersonalWeight:   0.2,  // Focus more on knowledge
        RAGKnowledgeWeight:  0.8,
        RAGIncludeSources:   true,
        
        Embedding: core.EmbeddingConfig{
            Provider:        "openai",
            Model:           "text-embedding-3-small",
            APIKey:          os.Getenv("OPENAI_API_KEY"),
            CacheEmbeddings: true, // Important for performance
            MaxBatchSize:    200,  // Larger batches
            TimeoutSeconds:  60,   // Longer timeout for large batches
        },
        
        Documents: core.DocumentConfig{
            AutoChunk:                true,
            SupportedTypes:           []string{"pdf", "txt", "md", "web", "code", "json"},
            MaxFileSize:              "50MB", // Larger files
            EnableMetadataExtraction: true,
            EnableURLScraping:        true,
        },
        
        Search: core.SearchConfigToml{
            HybridSearch:         true,
            KeywordWeight:        0.2, // Favor semantic search
            SemanticWeight:       0.8,
            EnableReranking:      false, // Can be enabled if available
            EnableQueryExpansion: false,
        },
    }
    
    return core.NewMemory(config)
}
```

## Memory Monitoring and Metrics

### 1. Performance Metrics

```go
type VectorDBMetrics struct {
    SearchLatency       []time.Duration
    KnowledgeSearches   int64
    PersonalSearches    int64
    HybridSearches      int64
    ContextBuilds       int64
    DocumentIngestions  int64
    mu                  sync.RWMutex
}

func (m *VectorDBMetrics) RecordSearch(searchType string, duration time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.SearchLatency = append(m.SearchLatency, duration)
    
    switch searchType {
    case "knowledge":
        m.KnowledgeSearches++
    case "personal":
        m.PersonalSearches++
    case "hybrid":
        m.HybridSearches++
    case "context":
        m.ContextBuilds++
    case "ingest":
        m.DocumentIngestions++
    }
    
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

func (m *VectorDBMetrics) GetStats() map[string]interface{} {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    return map[string]interface{}{
        "average_latency":      m.GetAverageLatency(),
        "knowledge_searches":   m.KnowledgeSearches,
        "personal_searches":    m.PersonalSearches,
        "hybrid_searches":      m.HybridSearches,
        "context_builds":       m.ContextBuilds,
        "document_ingestions":  m.DocumentIngestions,
        "total_operations":     m.KnowledgeSearches + m.PersonalSearches + m.HybridSearches + m.ContextBuilds + m.DocumentIngestions,
    }
}
```

### 2. Health Monitoring

```go
func monitorVectorDBHealth(memory core.Memory, metrics *VectorDBMetrics) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        
        // Check basic connectivity with personal memory
        start := time.Now()
        _, err := memory.Query(ctx, "health check query", 1)
        personalLatency := time.Since(start)
        
        if err != nil {
            log.Printf("Personal memory health check failed: %v", err)
        } else {
            metrics.RecordSearch("personal", personalLatency)
            if personalLatency > 1*time.Second {
                log.Printf("Personal memory latency high: %v", personalLatency)
            }
        }
        
        // Check knowledge base connectivity
        start = time.Now()
        _, err = memory.SearchKnowledge(ctx, "health check query", core.WithLimit(1))
        knowledgeLatency := time.Since(start)
        
        if err != nil {
            log.Printf("Knowledge base health check failed: %v", err)
        } else {
            metrics.RecordSearch("knowledge", knowledgeLatency)
            if knowledgeLatency > 2*time.Second {
                log.Printf("Knowledge search latency high: %v", knowledgeLatency)
            }
        }
        
        // Check hybrid search
        start = time.Now()
        hybridResult, err := memory.SearchAll(ctx, "health check query",
            core.WithLimit(1),
            core.WithIncludePersonal(true),
            core.WithIncludeKnowledge(true),
        )
        hybridLatency := time.Since(start)
        
        if err != nil {
            log.Printf("Hybrid search health check failed: %v", err)
        } else {
            metrics.RecordSearch("hybrid", hybridLatency)
            log.Printf("Vector DB Health - Personal: %dms, Knowledge: %dms, Hybrid: %dms (Total: %d results)", 
                personalLatency.Milliseconds(),
                knowledgeLatency.Milliseconds(),
                hybridLatency.Milliseconds(),
                hybridResult.TotalResults)
        }
        
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
    // Example backup strategy using Go's time.Ticker
    // In production, use a proper cron library like github.com/robfig/cron/v3
    
    // Daily full backup ticker
    dailyTicker := time.NewTicker(24 * time.Hour)
    go func() {
        for range dailyTicker.C {
            err := performFullBackup(memory)
            if err != nil {
                log.Printf("Full backup failed: %v", err)
            }
        }
    }()
    
    // Hourly incremental backup ticker
    hourlyTicker := time.NewTicker(1 * time.Hour)
    go func() {
        for range hourlyTicker.C {
            err := performIncrementalBackup(memory)
            if err != nil {
                log.Printf("Incremental backup failed: %v", err)
            }
        }
    }()
    
    log.Printf("Backup strategy initialized - daily full backups and hourly incremental backups")
    return nil
}

func performFullBackup(memory core.Memory) error {
    ctx := context.Background()
    
    backupFile := fmt.Sprintf("backup-full-%s.json", 
        time.Now().Format("2006-01-02-15-04-05"))
    
    // Since there's no ExportAll method in current API, we'll demonstrate
    // a backup strategy using available methods
    
    // Create backup structure
    backup := struct {
        Timestamp        time.Time                `json:"timestamp"`
        PersonalMemory   []core.Result           `json:"personal_memory"`
        KnowledgeBase    []core.KnowledgeResult  `json:"knowledge_base"`
        ChatHistory      []core.Message          `json:"chat_history"`
    }{
        Timestamp: time.Now(),
    }
    
    // Backup personal memory (get recent entries)
    personalResults, err := memory.Query(ctx, "", 1000) // Get up to 1000 entries
    if err != nil {
        log.Printf("Warning: Failed to backup personal memory: %v", err)
    } else {
        backup.PersonalMemory = personalResults
    }
    
    // Backup knowledge base (search with broad query)
    knowledgeResults, err := memory.SearchKnowledge(ctx, "",
        core.WithLimit(1000),
        core.WithScoreThreshold(0.0), // Include all results
    )
    if err != nil {
        log.Printf("Warning: Failed to backup knowledge base: %v", err)
    } else {
        backup.KnowledgeBase = knowledgeResults
    }
    
    // Backup chat history
    chatHistory, err := memory.GetHistory(ctx, 1000)
    if err != nil {
        log.Printf("Warning: Failed to backup chat history: %v", err)
    } else {
        backup.ChatHistory = chatHistory
    }
    
    // Serialize to JSON
    data, err := json.MarshalIndent(backup, "", "  ")
    if err != nil {
        return fmt.Errorf("JSON marshal failed: %w", err)
    }
    
    // Write to file
    err = os.WriteFile(backupFile, data, 0644)
    if err != nil {
        return fmt.Errorf("write backup failed: %w", err)
    }
    
    log.Printf("Full backup completed: %s (%d personal, %d knowledge, %d messages)", 
        backupFile, len(backup.PersonalMemory), len(backup.KnowledgeBase), len(backup.ChatHistory))
    return nil
}

func performIncrementalBackup(memory core.Memory) error {
    ctx := context.Background()
    
    // Get timestamp for incremental backup (last hour)
    since := time.Now().Add(-1 * time.Hour)
    backupFile := fmt.Sprintf("backup-incremental-%s.json", 
        time.Now().Format("2006-01-02-15-04-05"))
    
    // Create incremental backup structure
    backup := struct {
        Timestamp     time.Time               `json:"timestamp"`
        Since         time.Time               `json:"since"`
        RecentMemory  []core.Result          `json:"recent_memory"`
        RecentHistory []core.Message         `json:"recent_history"`
    }{
        Timestamp: time.Now(),
        Since:     since,
    }
    
    // Get recent personal memory entries
    recentMemory, err := memory.Query(ctx, "", 100) // Get recent entries
    if err != nil {
        log.Printf("Warning: Failed to backup recent memory: %v", err)
    } else {
        // Filter by timestamp (simplified - in production, use proper filtering)
        var filtered []core.Result
        for _, result := range recentMemory {
            if result.CreatedAt.After(since) {
                filtered = append(filtered, result)
            }
        }
        backup.RecentMemory = filtered
    }
    
    // Get recent chat history
    recentHistory, err := memory.GetHistory(ctx, 100)
    if err != nil {
        log.Printf("Warning: Failed to backup recent history: %v", err)
    } else {
        // Filter by timestamp
        var filtered []core.Message
        for _, msg := range recentHistory {
            if msg.CreatedAt.After(since) {
                filtered = append(filtered, msg)
            }
        }
        backup.RecentHistory = filtered
    }
    
    // Serialize to JSON
    data, err := json.MarshalIndent(backup, "", "  ")
    if err != nil {
        return fmt.Errorf("JSON marshal failed: %w", err)
    }
    
    // Write to file
    err = os.WriteFile(backupFile, data, 0644)
    if err != nil {
        return fmt.Errorf("write incremental backup failed: %w", err)
    }
    
    log.Printf("Incremental backup completed: %s (%d memory, %d messages)", 
        backupFile, len(backup.RecentMemory), len(backup.RecentHistory))
    return nilemory) error {
    ctx := context.Background()
    
    backupFile := fmt.Sprintf("backup-incremental-%s.json", 
        time.Now().Format("2006-01-02-15-04-05"))
    
    // For incremental backup, we'll get recent data (last hour)
    // This is a simplified approach - in production, you'd track changes
    
    backup := struct {
        Timestamp        time.Time                `json:"timestamp"`
        PersonalMemory   []core.Result           `json:"personal_memory"`
        KnowledgeBase    []core.KnowledgeResult  `json:"knowledge_base"`
        ChatHistory      []core.Message          `json:"chat_history"`
    }{
        Timestamp: time.Now(),
    }
    
    // Get recent personal memory entries
    personalResults, err := memory.Query(ctx, "", 100) // Smaller batch for incremental
    if err != nil {
        log.Printf("Warning: Failed to backup recent personal memory: %v", err)
    } else {
        backup.PersonalMemory = personalResults
    }
    
    // Get recent knowledge base entries
    knowledgeResults, err := memory.SearchKnowledge(ctx, "",
        core.WithLimit(100),
        core.WithScoreThreshold(0.0),
    )
    if err != nil {
        log.Printf("Warning: Failed to backup recent knowledge base: %v", err)
    } else {
        backup.KnowledgeBase = knowledgeResults
    }
    
    // Get recent chat history
    chatHistory, err := memory.GetHistory(ctx, 100)
    if err != nil {
        log.Printf("Warning: Failed to backup recent chat history: %v", err)
    } else {
        backup.ChatHistory = chatHistory
    }
    
    // Serialize to JSON
    data, err := json.MarshalIndent(backup, "", "  ")
    if err != nil {
        return fmt.Errorf("JSON marshal failed: %w", err)
    }
    
    // Write to file
    err = os.WriteFile(backupFile, data, 0644)
    if err != nil {
        return fmt.Errorf("write incremental backup failed: %w", err)
    }
    
    log.Printf("Incremental backup completed: %s (%d personal, %d knowledge, %d messages)", 
        backupFile, len(backup.PersonalMemory), len(backup.KnowledgeBase), len(backup.ChatHistory))
    return nil
}
```

## Troubleshooting Common Issues

### 1. Performance Issues

```go
func diagnosePerformanceIssues(memory core.Memory) {
    ctx := context.Background()
    
    log.Printf("=== Vector Database Performance Diagnosis ===")
    
    // Test personal memory performance
    queries := []string{
        "machine learning algorithms",
        "neural network architecture", 
        "data processing pipeline",
        "artificial intelligence concepts",
    }
    
    log.Printf("Testing Personal Memory Performance:")
    for _, query := range queries {
        start := time.Now()
        results, err := memory.Query(ctx, query, 5)
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("  Query '%s' failed: %v", query, err)
        } else {
            log.Printf("  Query '%s': %dms (%d results)", 
                query, duration.Milliseconds(), len(results))
        }
    }
    
    log.Printf("Testing Knowledge Base Performance:")
    for _, query := range queries {
        start := time.Now()
        results, err := memory.SearchKnowledge(ctx, query, core.WithLimit(5))
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("  Knowledge search '%s' failed: %v", query, err)
        } else {
            log.Printf("  Knowledge search '%s': %dms (%d results)", 
                query, duration.Milliseconds(), len(results))
        }
    }
    
    log.Printf("Testing Hybrid Search Performance:")
    for _, query := range queries {
        start := time.Now()
        hybridResult, err := memory.SearchAll(ctx, query,
            core.WithLimit(10),
            core.WithIncludePersonal(true),
            core.WithIncludeKnowledge(true),
        )
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("  Hybrid search '%s' failed: %v", query, err)
        } else {
            log.Printf("  Hybrid search '%s': %dms (%d total results)", 
                query, duration.Milliseconds(), hybridResult.TotalResults)
        }
    }
}or _, query := range queries {
        start := time.Now()
        results, err := memory.Query(ctx, query, 10)
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("  ❌ Personal query failed: %s - %v", query, err)
        } else {
            log.Printf("  ✅ Personal query: %s - %d results in %v", 
                query, len(results), duration)
        }
    }
    
    // Test knowledge base performance
    log.Printf("\nTesting Knowledge Base Performance:")
    for _, query := range queries {
        start := time.Now()
        results, err := memory.SearchKnowledge(ctx, query,
            core.WithLimit(10),
            core.WithScoreThreshold(0.5),
        )
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("  ❌ Knowledge query failed: %s - %v", query, err)
        } else {
            log.Printf("  ✅ Knowledge query: %s - %d results in %v", 
                query, len(results), duration)
        }
    }
    
    // Test hybrid search performance
    log.Printf("\nTesting Hybrid Search Performance:")
    for _, query := range queries {
        start := time.Now()
        results, err := memory.SearchAll(ctx, query,
            core.WithLimit(10),
            core.WithIncludePersonal(true),
            core.WithIncludeKnowledge(true),
        )
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("  ❌ Hybrid query failed: %s - %v", query, err)
        } else {
            log.Printf("  ✅ Hybrid query: %s - %d total results (%d personal, %d knowledge) in %v", 
                query, results.TotalResults, len(results.PersonalMemory), 
                len(results.Knowledge), duration)
        }
    }
    
    // Test RAG context building performance
    log.Printf("\nTesting RAG Context Building Performance:")
    start := time.Now()
    ragContext, err := memory.BuildContext(ctx, "Explain machine learning concepts",
        core.WithMaxTokens(2000),
        core.WithPersonalWeight(0.3),
        core.WithKnowledgeWeight(0.7),
        core.WithIncludeSources(true),
    )
    duration := time.Since(start)
    
    if err != nil {
        log.Printf("  ❌ RAG context building failed: %v", err)
    } else {
        log.Printf("  ✅ RAG context built in %v - %d tokens, %d sources", 
            duration, ragContext.TokenCount, len(ragContext.Sources))
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
    CONSTRAINT valid_content_type CHECK (content_type ~ '^[a-z][a-z0-9_-]*$')$')
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
        Provider:        "openai",
        Model:           "text-embedding-3-small", // Good balance of cost/performance
        CacheEmbeddings: true, // Important for performance and cost
        MaxBatchSize:    100,  // Optimize for API rate limits
        TimeoutSeconds:  30,   // Reasonable timeout
    }
}

// For high-performance scenarios
func optimizeEmbeddingsHighPerformance() core.EmbeddingConfig {
    return core.EmbeddingConfig{
        Provider:        "openai",
        Model:           "text-embedding-3-large", // Higher quality embeddings
        CacheEmbeddings: true,
        MaxBatchSize:    200, // Larger batches for better throughput
        TimeoutSeconds:  60,  // Longer timeout for large batches
    }
}

// For cost-optimized scenarios
func optimizeEmbeddingsCostEffective() core.EmbeddingConfig {
    return core.EmbeddingConfig{
        Provider:        "ollama",
        Model:           "mxbai-embed-large", // Free local embeddings
        BaseURL:         "http://localhost:11434",
        CacheEmbeddings: true,
        MaxBatchSize:    50,  // Smaller batches for local processing
        TimeoutSeconds:  120, // Longer timeout for local processing
    }
}
```

### 3. Query Optimization

```go
func optimizeQueries(memory core.Memory) {
    ctx := context.Background()
    
    // Example query to optimize
    query := "machine learning algorithms"
    
    // 1. Use appropriate limits - don't retrieve more than needed
    results, err := memory.SearchKnowledge(ctx, query,
        core.WithLimit(10), // Reasonable limit
        core.WithScoreThreshold(0.7), // Filter low-relevance results
    )
    if err != nil {
        log.Printf("Optimized knowledge search failed: %v", err)
        return
    }
    
    // 2. Use tag filters to reduce search space
    results, err = memory.SearchKnowledge(ctx, query,
        core.WithLimit(5),
        core.WithTags([]string{"ai", "algorithms"}), // Filter by relevant tags
        core.WithScoreThreshold(0.8), // Higher threshold for quality
    )
    if err != nil {
        log.Printf("Tag-filtered search failed: %v", err)
        return
    }
    
    // 3. Use document type filters for specific content
    results, err = memory.SearchKnowledge(ctx, query,
        core.WithLimit(10),
        core.WithDocumentTypes([]core.DocumentType{
            core.DocumentTypeMarkdown,
            core.DocumentTypePDF,
        }),
    )
    if err != nil {
        log.Printf("Document type filtered search failed: %v", err)
        return
    }
    
    // 4. Optimize hybrid search by controlling weights
    hybridResults, err := memory.SearchAll(ctx, query,
        core.WithLimit(15), // Slightly higher for hybrid
        core.WithIncludePersonal(true),
        core.WithIncludeKnowledge(true),
    )
    if err != nil {
        log.Printf("Hybrid search failed: %v", err)
        return
    }
    
    // 5. Cache frequent queries (simple in-memory cache example)
    var queryCache = make(map[string][]core.KnowledgeResult)
    var cacheMutex sync.RWMutex
    
    cacheKey := fmt.Sprintf("knowledge:%s", query)
    
    // Check cache first
    cacheMutex.RLock()
    if cached, exists := queryCache[cacheKey]; exists {
        cacheMutex.RUnlock()
        log.Printf("Cache hit for query: %s", query)
        results = cached
        return
    }
    cacheMutex.RUnlock()
    
    // Perform search and cache results
    results, err = memory.SearchKnowledge(ctx, query, core.WithLimit(10))
    if err == nil {
        cacheMutex.Lock()
        queryCache[cacheKey] = results
        cacheMutex.Unlock()
        log.Printf("Cached results for query: %s", query)
    }
    
    // 6. Optimize RAG context building
    ragContext, err := memory.BuildContext(ctx, query,
        core.WithMaxTokens(4000), // Appropriate context size
        core.WithPersonalWeight(0.2), // Focus more on knowledge
        core.WithKnowledgeWeight(0.8),
        core.WithHistoryLimit(5), // Limited history for focus
        core.WithIncludeSources(true), // Include sources for transparency
    )
    if err != nil {
        log.Printf("RAG context building failed: %v", err)
        return
    }
    
    log.Printf("Optimized RAG context: %d tokens, %d sources", 
        ragContext.TokenCount, len(ragContext.Sources))
}
```

## Error Handling and Troubleshooting

### 1. Common Connection Issues

```go
func handleConnectionErrors(config core.AgentMemoryConfig) (core.Memory, error) {
    // Implement retry logic for connection failures
    maxRetries := 3
    backoff := time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        memory, err := core.NewMemory(config)
        if err == nil {
            // Test the connection
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            _, testErr := memory.Query(ctx, "connection test", 1)
            cancel()
            
            if testErr == nil {
                log.Printf("Successfully connected to %s on attempt %d", config.Provider, attempt+1)
                return memory, nil
            }
            
            memory.Close() // Clean up failed connection
            err = testErr
        }
        
        if attempt == maxRetries-1 {
            return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
        }
        
        log.Printf("Connection attempt %d failed, retrying in %v: %v", 
            attempt+1, backoff, err)
        time.Sleep(backoff)
        backoff *= 2 // Exponential backoff
    }
    
    return nil, fmt.Errorf("connection failed after all retries")
}
```

### 2. Embedding Service Issues

```go
func handleEmbeddingErrors(memory core.Memory) error {
    ctx := context.Background()
    
    // Test embedding generation with fallback
    testContent := "This is a test for embedding generation"
    
    err := memory.Store(ctx, testContent, "test", "embedding-check")
    if err != nil {
        if strings.Contains(err.Error(), "embedding") {
            log.Printf("Embedding service error: %v", err)
            log.Printf("Check your embedding configuration:")
            log.Printf("- API key is valid")
            log.Printf("- Model name is correct")
            log.Printf("- Service is accessible")
            log.Printf("- Rate limits are not exceeded")
            return fmt.Errorf("embedding service unavailable: %w", err)
        }
        return fmt.Errorf("storage error: %w", err)
    }
    
    // Clean up test data
    memory.ClearSession(ctx)
    return nil
}
```

### 3. Performance Troubleshooting

```go
func diagnosePerformanceIssues(memory core.Memory) {
    ctx := context.Background()
    
    // Test query performance
    queries := []string{
        "machine learning",
        "artificial intelligence",
        "neural networks",
        "deep learning",
    }
    
    for _, query := range queries {
        start := time.Now()
        results, err := memory.SearchKnowledge(ctx, query, core.WithLimit(10))
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("Query failed: %s - %v", query, err)
            continue
        }
        
        log.Printf("Query: '%s' - %d results in %v", query, len(results), duration)
        
        if duration > 2*time.Second {
            log.Printf("WARNING: Slow query detected. Consider:")
            log.Printf("- Optimizing database indexes")
            log.Printf("- Reducing search scope with filters")
            log.Printf("- Checking database resource usage")
            log.Printf("- Reviewing embedding service performance")
        }
    }
}
```

## Conclusion

Vector databases provide the foundation for sophisticated memory systems in AgenticGoKit. Key takeaways:

- Choose the right vector database for your use case (pgvector for SQL integration, Weaviate for specialized features)
- Configure embedding services properly with appropriate models and settings
- Optimize indexes and query patterns for performance
- Implement proper monitoring, backup, and error handling strategies
- Use current AgentMemoryConfig structure with RAG-enhanced settings
- Leverage SearchKnowledge, SearchAll, and BuildContext methods for advanced operations

Vector databases enable semantic search, RAG systems, and intelligent memory that can significantly enhance agent capabilities.

## Next Steps

Now that you have production-ready vector storage, build on this foundation:

### 📄 **Content Processing**
- **[Document Ingestion](document-ingestion.md)** - Process and ingest documents into your vector database
- Learn chunking strategies, metadata extraction, and batch processing

### 🧠 **Intelligent Retrieval**
- **[RAG Implementation](rag-implementation.md)** - Build retrieval-augmented generation systems
- Combine your vector database with LLMs for intelligent responses

### 🏗️ **Advanced Search**
- **[Knowledge Bases](knowledge-bases.md)** - Create comprehensive knowledge systems
- Advanced search patterns and multi-modal content handling

### ⚡ **Performance & Scale**
- **[Memory Optimization](memory-optimization.md)** - Advanced performance tuning
- Database optimization, caching strategies, and scaling patterns

::: info Prerequisites Complete
✅ You now have production-ready vector storage  
🎯 Next: Add content with document ingestion
:::

## Related Topics

- **[Basic Memory Operations](basic-memory.md)** - Review memory fundamentals
- **[Memory Systems Overview](README.md)** - Complete architecture guide

## Further Reading

- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [Weaviate Documentation](https://weaviate.io/developers/weaviate)
- [OpenAI Embeddings Guide](https://platform.openai.com/docs/guides/embeddings)