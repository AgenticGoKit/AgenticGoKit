# Memory API

**Persistent storage, RAG, and knowledge management**

This document covers AgenticGoKit's Memory API, which enables agents to store and retrieve information persistently. The Memory system is essential for building agents with long-term memory, knowledge bases, and RAG (Retrieval-Augmented Generation) capabilities.

## 📋 Core Concepts

### Memory Interface

The core interface for agent memory systems:

```go
type Memory interface {
    // Store stores a memory with optional metadata
    Store(ctx context.Context, content string, metadata map[string]interface{}) (string, error)
    
    // Search finds memories similar to the query
    Search(ctx context.Context, query string, limit int, minScore float64) ([]MemorySearchResult, error)
    
    // Get retrieves a specific memory by ID
    Get(ctx context.Context, id string) (*MemorySearchResult, error)
    
    // Delete removes a memory by ID
    Delete(ctx context.Context, id string) error
}
```

## 🚀 Basic Usage

### Creating Memory System

```go
// Create memory configuration
config := agentflow.AgentMemoryConfig{
    Provider:   "pgvector",
    Connection: "postgres://user:password@localhost:5432/agentflow",
    Dimensions: 1536,
    Embedding: agentflow.EmbeddingConfig{
        Provider: "openai",
        APIKey:   "your-api-key",
        Model:    "text-embedding-3-small",
    },
}

// Initialize memory
memory, err := agentflow.NewMemory(config)
if err != nil {
    log.Fatal(err)
}
defer memory.Close()
```

For complete documentation including RAG operations, document ingestion, and all memory providers, see the [Agent API reference](agent.md#memory).
