# Memory Systems in AgenticGoKit

> **Navigation:** [Documentation Home](../../README.md) → [Tutorials](../README.md) → **Memory Systems**

## Overview

Memory systems are crucial for building intelligent agents that can learn, remember, and build upon previous interactions. This tutorial series explores AgenticGoKit's memory capabilities, from basic in-memory storage to advanced RAG (Retrieval-Augmented Generation) systems with vector databases.

Memory systems enable agents to maintain context across conversations, store knowledge, and retrieve relevant information to enhance their responses.

## Prerequisites

- Understanding of [Core Concepts](../core-concepts/README.md)
- Basic knowledge of databases and data storage
- Familiarity with vector embeddings and similarity search

## Memory System Architecture

AgenticGoKit's memory system is built on a flexible architecture that supports multiple storage backends and retrieval strategies:

```
┌─────────────┐    ┌──────────────┐    ┌─────────────────┐
│   Agent     │───▶│    Memory    │───▶│ Storage Backend │
│             │    │   Interface  │    │                 │
└─────────────┘    └──────────────┘    └─────────────────┘
                           │
                           ▼
                   ┌──────────────┐
                   │   Embedding  │
                   │   Provider   │
                   └──────────────┘
```

## Memory Types

### 1. Conversational Memory
Stores chat history and conversation context:
- Message history
- User preferences
- Session context
- Conversation metadata

### 2. Knowledge Memory
Stores factual information and documents:
- Document chunks
- Factual knowledge
- Reference materials
- Structured data

### 3. Episodic Memory
Stores experiences and events:
- Past interactions
- Learning experiences
- Feedback and corrections
- Temporal sequences

### 4. Working Memory
Temporary storage during processing:
- Intermediate results
- Processing context
- Temporary variables
- Computation state

## Memory Interface

The core memory interface provides a unified API for all memory operations including RAG:

```go
type Memory interface {
    // Personal memory operations
    Store(ctx context.Context, content string, tags ...string) error
    Query(ctx context.Context, query string, limit ...int) ([]Result, error)
    Remember(ctx context.Context, key string, value any) error
    Recall(ctx context.Context, key string) (any, error)

    // Chat history management
    AddMessage(ctx context.Context, role, content string) error
    GetHistory(ctx context.Context, limit ...int) ([]Message, error)

    // Session management
    NewSession() string
    SetSession(ctx context.Context, sessionID string) context.Context
    ClearSession(ctx context.Context) error
    Close() error

    // RAG-Enhanced Knowledge Base Operations
    IngestDocument(ctx context.Context, doc Document) error
    IngestDocuments(ctx context.Context, docs []Document) error
    SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error)

    // Hybrid Search (Personal Memory + Knowledge Base)
    SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error)

    // RAG Context Assembly for LLM Prompts
    BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error)
}
```

## Memory Providers

### 1. In-Memory Provider
Fast, non-persistent storage for development and testing:

```go
// Create in-memory storage
memory, err := core.NewMemory(core.AgentMemoryConfig{
    Provider:   "memory",
    Connection: "memory",
    MaxResults: 10,
    Dimensions: 1536,
    AutoEmbed:  true,
})
```

**Use Cases:**
- Development and testing
- Temporary storage
- High-speed access
- Stateless applications

### 2. PostgreSQL with pgvector
Production-ready vector storage with SQL capabilities:

```go
// Create pgvector storage
memory, err := core.NewMemory(core.AgentMemoryConfig{
    Provider:                "pgvector",
    Connection:              "postgres://user:pass@localhost:5432/agentdb",
    EnableRAG:               true,
    EnableKnowledgeBase:     true,
    Dimensions:              1536, // OpenAI embedding dimensions
    KnowledgeMaxResults:     20,
    KnowledgeScoreThreshold: 0.7,
    Embedding: core.EmbeddingConfig{
        Provider:        "openai",
        Model:           "text-embedding-3-small",
        APIKey:          os.Getenv("OPENAI_API_KEY"),
        MaxBatchSize:    100,
        TimeoutSeconds:  30,
        CacheEmbeddings: true,
    },
})
```

**Use Cases:**
- Production applications
- Complex queries
- ACID compliance
- Structured and unstructured data

### 3. Weaviate Vector Database
Specialized vector database with advanced features:

```go
// Create Weaviate storage
memory, err := core.NewMemory(core.AgentMemoryConfig{
    Provider:                "weaviate",
    Connection:              "http://localhost:8080",
    EnableRAG:               true,
    EnableKnowledgeBase:     true,
    Dimensions:              1536,
    KnowledgeMaxResults:     20,
    KnowledgeScoreThreshold: 0.7,
    Embedding: core.EmbeddingConfig{
        Provider:        "openai",
        Model:           "text-embedding-3-small",
        APIKey:          os.Getenv("OPENAI_API_KEY"),
        MaxBatchSize:    100,
        TimeoutSeconds:  30,
        CacheEmbeddings: true,
    },
})
```

**Use Cases:**
- Large-scale vector search
- Multi-modal data
- Advanced filtering
- Real-time updates

## Basic Memory Operations

### 1. Storing Information

```go
// Store simple text
err := memory.Store(ctx, "Paris is the capital of France", "fact")

// Store with metadata
err = memory.Store(ctx, 
    "The user prefers technical explanations", 
    "preference",
    core.WithSession("user-123"),
    core.WithMetadata(map[string]string{
        "category": "user-preference",
        "priority": "high",
    }),
)

// Store conversation message
err = memory.Store(ctx,
    "How do I implement a binary search?",
    "user-message",
    core.WithSession("user-123"),
    core.WithTimestamp(time.Now()),
)
```

### 2. Searching Memory

```go
// Basic search
results, err := memory.Search(ctx, "capital of France")

// Search with options
results, err = memory.Search(ctx, 
    "user preferences",
    core.WithLimit(5),
    core.WithScoreThreshold(0.7),
    core.WithSession("user-123"),
    core.WithContentType("preference"),
)

// Process results
for _, result := range results {
    fmt.Printf("Content: %s (Score: %.3f)\n", result.Content, result.Score)
}
```

### 3. Conversation History

```go
// Get recent conversation history
messages, err := memory.GetHistory(ctx, 10, 
    core.WithSession("user-123"),
    core.WithTimeRange(time.Now().Add(-24*time.Hour), time.Now()),
)

// Process conversation history
for _, msg := range messages {
    fmt.Printf("[%s] %s: %s\n", 
        msg.Timestamp.Format("15:04"), 
        msg.Role, 
        msg.Content)
}
```

## RAG (Retrieval-Augmented Generation)

RAG enhances agent responses by retrieving relevant information from memory:

### 1. Basic RAG Implementation

```go
type RAGAgent struct {
    name   string
    llm    LLMProvider
    memory Memory
}

func (r *RAGAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    query, _ := state.Get("message")
    queryStr := query.(string)
    
    // Retrieve relevant context
    results, err := r.memory.Search(ctx, queryStr,
        core.WithLimit(3),
        core.WithScoreThreshold(0.7),
    )
    if err != nil {
        return AgentResult{}, err
    }
    
    // Build context from retrieved results
    var context strings.Builder
    for _, result := range results {
        context.WriteString(fmt.Sprintf("- %s\n", result.Content))
    }
    
    // Create enhanced prompt
    prompt := fmt.Sprintf(`Context:
%s

Question: %s

Please answer the question using the provided context.`, 
        context.String(), queryStr)
    
    // Generate response with context
    response, err := r.llm.Generate(ctx, prompt)
    if err != nil {
        return AgentResult{}, err
    }
    
    // Store the interaction
    r.memory.Store(ctx, queryStr, "user-message")
    r.memory.Store(ctx, response, "assistant-response")
    
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("context_used", len(results))
    
    return AgentResult{OutputState: outputState}, nil
}
```

### 2. Advanced RAG with Reranking

```go
type AdvancedRAGAgent struct {
    name     string
    llm      LLMProvider
    memory   Memory
    reranker Reranker
}

func (r *AdvancedRAGAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    query, _ := state.Get("message")
    queryStr := query.(string)
    
    // Initial retrieval with higher limit
    results, err := r.memory.Search(ctx, queryStr,
        core.WithLimit(10),
        core.WithScoreThreshold(0.5),
    )
    if err != nil {
        return AgentResult{}, err
    }
    
    // Rerank results for better relevance
    rerankedResults, err := r.reranker.Rerank(ctx, queryStr, results)
    if err != nil {
        return AgentResult{}, err
    }
    
    // Take top results after reranking
    topResults := rerankedResults[:min(3, len(rerankedResults))]
    
    // Build enhanced context
    context := r.buildContext(topResults)
    
    // Generate response
    response, err := r.generateWithContext(ctx, queryStr, context)
    if err != nil {
        return AgentResult{}, err
    }
    
    // Store interaction with metadata
    r.storeInteraction(ctx, queryStr, response, topResults)
    
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("sources", r.extractSources(topResults))
    
    return AgentResult{OutputState: outputState}, nil
}
```

## Memory Configuration

### 1. Embedding Configuration

```go
type EmbeddingConfig struct {
    Provider   string            `yaml:"provider"`   // "openai", "huggingface", "local"
    Model      string            `yaml:"model"`      // Model name
    APIKey     string            `yaml:"api_key"`    // API key if needed
    Dimensions int               `yaml:"dimensions"` // Embedding dimensions
    BatchSize  int               `yaml:"batch_size"` // Batch size for processing
    Options    map[string]string `yaml:"options"`    // Provider-specific options
}

// OpenAI embeddings
embeddingConfig := core.EmbeddingConfig{
    Provider:   "openai",
    Model:      "text-embedding-3-small",
    APIKey:     os.Getenv("OPENAI_API_KEY"),
    Dimensions: 1536,
    BatchSize:  100,
}

// Hugging Face embeddings
embeddingConfig := core.EmbeddingConfig{
    Provider:   "huggingface",
    Model:      "sentence-transformers/all-MiniLM-L6-v2",
    Dimensions: 384,
    BatchSize:  32,
}
```

### 2. Memory Configuration

```go
type AgentMemoryConfig struct {
    // Core memory settings
    Provider   string `toml:"provider"`    // pgvector, weaviate, memory
    Connection string `toml:"connection"`  // postgres://..., http://..., or "memory"
    MaxResults int    `toml:"max_results"` // default: 10
    Dimensions int    `toml:"dimensions"`  // default: 1536
    AutoEmbed  bool   `toml:"auto_embed"`  // default: true

    // RAG-enhanced settings
    EnableKnowledgeBase     bool    `toml:"enable_knowledge_base"`     // default: true
    KnowledgeMaxResults     int     `toml:"knowledge_max_results"`     // default: 20
    KnowledgeScoreThreshold float32 `toml:"knowledge_score_threshold"` // default: 0.7
    ChunkSize               int     `toml:"chunk_size"`                // default: 1000
    ChunkOverlap            int     `toml:"chunk_overlap"`             // default: 200

    // RAG context assembly settings
    EnableRAG           bool    `toml:"enable_rag"`             // default: true
    RAGMaxContextTokens int     `toml:"rag_max_context_tokens"` // default: 4000
    RAGPersonalWeight   float32 `toml:"rag_personal_weight"`    // default: 0.3
    RAGKnowledgeWeight  float32 `toml:"rag_knowledge_weight"`   // default: 0.7
    RAGIncludeSources   bool    `toml:"rag_include_sources"`    // default: true

    // Document processing settings
    Documents DocumentConfig `toml:"documents"`

    // Embedding service settings
    Embedding EmbeddingConfig `toml:"embedding"`

    // Search settings
    Search SearchConfigToml `toml:"search"`
}

// Complete memory configuration
config := core.AgentMemoryConfig{
    Provider:                "pgvector",
    Connection:              "postgres://user:pass@localhost:5432/agentdb",
    EnableRAG:               true,
    EnableKnowledgeBase:     true,
    ChunkSize:               1000,
    ChunkOverlap:            200,
    Dimensions:              1536,
    KnowledgeMaxResults:     20,
    KnowledgeScoreThreshold: 0.7,
    RAGMaxContextTokens:     4000,
    RAGPersonalWeight:       0.3,
    RAGKnowledgeWeight:      0.7,
    RAGIncludeSources:       true,
    Embedding: core.EmbeddingConfig{
        Provider:        "openai",
        Model:           "text-embedding-3-small",
        APIKey:          os.Getenv("OPENAI_API_KEY"),
        MaxBatchSize:    100,
        TimeoutSeconds:  30,
        CacheEmbeddings: true,
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
```

## Memory Integration with Agents

### 1. Agent Builder Integration

```go
// Create agent with memory
agent, err := core.NewAgent("knowledge-agent").
    WithLLMAndConfig(llmProvider, llmConfig).
    WithMemory(memory).
    WithMCPAndConfig(mcpManager, mcpConfig).
    Build()
```

### 2. Custom Memory Integration

```go
type MemoryEnabledAgent struct {
    name   string
    llm    LLMProvider
    memory Memory
    config MemoryConfig
}

func (m *MemoryEnabledAgent) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
    // Extract user message
    message, _ := state.Get("message")
    messageStr := message.(string)
    
    // Get conversation history for context
    history, err := m.memory.GetHistory(ctx, 5,
        core.WithSession(event.GetSessionID()),
    )
    if err != nil {
        return AgentResult{}, err
    }
    
    // Search for relevant knowledge
    knowledge, err := m.memory.Search(ctx, messageStr,
        core.WithLimit(3),
        core.WithScoreThreshold(0.7),
        core.WithContentType("knowledge"),
    )
    if err != nil {
        return AgentResult{}, err
    }
    
    // Build enhanced prompt with history and knowledge
    prompt := m.buildEnhancedPrompt(messageStr, history, knowledge)
    
    // Generate response
    response, err := m.llm.Generate(ctx, prompt)
    if err != nil {
        return AgentResult{}, err
    }
    
    // Store the interaction
    m.storeInteraction(ctx, event.GetSessionID(), messageStr, response)
    
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("knowledge_used", len(knowledge))
    outputState.Set("history_length", len(history))
    
    return AgentResult{OutputState: outputState}, nil
}
```

## Tutorial Series Structure

This memory systems tutorial series covers:

### 1. [Basic Memory](basic-memory.md)
- In-memory storage
- Simple operations
- Session management
- Basic retrieval

### 2. [Vector Databases](vector-databases.md)
- pgvector setup and usage
- Weaviate integration
- Embedding strategies
- Performance optimization

### 3. [Document Ingestion](document-ingestion.md)
- Document processing pipeline
- Text chunking strategies
- Metadata extraction
- Batch processing optimization

### 4. [RAG Implementation](rag-implementation.md)
- Retrieval-Augmented Generation
- Context building
- Prompt engineering
- Response enhancement

### 5. [Knowledge Bases](knowledge-bases.md)
- Knowledge base architecture
- Advanced search patterns
- Multi-modal content
- Production deployment

### 6. [Memory Optimization](memory-optimization.md)
- Performance tuning
- Scaling strategies
- Caching mechanisms
- Resource management

## Best Practices

### 1. Memory Design Principles

- **Separate concerns**: Use different memory types for different purposes
- **Optimize for retrieval**: Design storage for efficient search
- **Manage lifecycle**: Clean up old or irrelevant memories
- **Monitor performance**: Track memory usage and search performance

### 2. RAG Best Practices

- **Chunk appropriately**: Balance context size and relevance
- **Use metadata**: Enhance retrieval with structured metadata
- **Implement reranking**: Improve relevance with secondary ranking
- **Handle failures**: Gracefully degrade when memory is unavailable

### 3. Production Considerations

- **Scale horizontally**: Use distributed storage for large datasets
- **Implement caching**: Cache frequent queries and embeddings
- **Monitor costs**: Track embedding API usage and storage costs
- **Backup data**: Ensure memory data is backed up and recoverable

## Common Use Cases

### 1. Conversational AI
- Chat history maintenance
- User preference learning
- Context-aware responses
- Personalization

### 2. Knowledge Management
- Document Q&A systems
- Technical support bots
- Research assistants
- Information retrieval

### 3. Learning Systems
- Adaptive agents
- Feedback incorporation
- Experience replay
- Continuous improvement

### 4. Multi-Agent Coordination
- Shared knowledge bases
- Inter-agent communication
- Collaborative learning
- Distributed memory

## Conclusion

Memory systems are fundamental to building intelligent agents that can learn, remember, and improve over time. AgenticGoKit provides flexible memory abstractions that support various storage backends and retrieval strategies.

The key to effective memory systems is choosing the right combination of storage backend, embedding strategy, and retrieval approach for your specific use case.

## Next Steps

- [Basic Memory](basic-memory.md) - Start with simple memory operations
- [Vector Databases](vector-databases.md) - Learn about production storage
- [RAG Implementation](rag-implementation.md) - Build retrieval-augmented systems
- [Knowledge Bases](knowledge-bases.md) - Create comprehensive knowledge systems

## Further Reading

- [API Reference: Memory Interface](../../reference/api/agent.md#memory)
- [Examples: Memory Systems](../../examples/)
- [Configuration Guide: Memory Settings](../../reference/api/configuration.md)
