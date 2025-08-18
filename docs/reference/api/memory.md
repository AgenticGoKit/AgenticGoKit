# Memory API

**Persistent storage, RAG, and knowledge management**

This document covers AgenticGoKit's Memory API, which enables agents to store and retrieve information persistently. The Memory system is essential for building agents with long-term memory, knowledge bases, and RAG (Retrieval-Augmented Generation) capabilities.

## 📋 Core Concepts

### Memory Interface

The core interface for agent memory systems:

```go
type Memory interface {
    // Personal memory operations
    Store(ctx context.Context, content string, tags ...string) error
    Query(ctx context.Context, query string, limit ...int) ([]core.Result, error)

    // Key-value storage
    Remember(ctx context.Context, key string, value any) error
    Recall(ctx context.Context, key string) (any, error)

    // Chat history
    AddMessage(ctx context.Context, role, content string) error
    GetHistory(ctx context.Context, limit ...int) ([]core.Message, error)

    // Knowledge base (RAG)
    IngestDocument(ctx context.Context, doc core.Document) error
    IngestDocuments(ctx context.Context, docs []core.Document) error
    SearchKnowledge(ctx context.Context, query string, options ...core.SearchOption) ([]core.KnowledgeResult, error)
    SearchAll(ctx context.Context, query string, options ...core.SearchOption) (*core.HybridResult, error)
    BuildContext(ctx context.Context, query string, options ...core.ContextOption) (*core.RAGContext, error)

    // Session management
    NewSession() string
    SetSession(ctx context.Context, sessionID string) context.Context
    ClearSession(ctx context.Context) error
    Close() error
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
