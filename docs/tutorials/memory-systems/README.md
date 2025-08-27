---
title: Memory Systems Overview
description: Comprehensive guide to AgenticGoKit's memory systems, including personal memory, chat history, knowledge bases, and RAG implementations.
---

# Memory Systems in AgenticGoKit

::: tip Navigation
[Documentation Home](../../README.md) → [Tutorials](../README.md) → **Memory Systems**
:::

## Overview

Memory systems are fundamental to building intelligent agents that can learn, remember, and build upon previous interactions. AgenticGoKit provides a comprehensive memory architecture that supports everything from simple in-memory storage to advanced RAG (Retrieval-Augmented Generation) systems with vector databases.

The memory system enables agents to maintain conversational context, store and retrieve knowledge, implement learning patterns, and provide contextually relevant responses through sophisticated retrieval mechanisms.

## Prerequisites

- Understanding of [Core Concepts](../core-concepts/README.md)
- Basic knowledge of Go programming and interfaces
- Familiarity with vector embeddings and similarity search
- Understanding of database concepts and storage systems

## Memory System Architecture

AgenticGoKit's memory system is built on a unified Memory interface that supports multiple storage backends, embedding providers, and retrieval strategies:

```mermaid
graph TB
    A[Agent] --> B[Memory Interface]
    B --> C[Storage Backend]
    B --> D[Embedding Service]
    B --> E[RAG Engine]
    
    C --> F[In-Memory]
    C --> G[pgvector]
    C --> H[Weaviate]
    
    D --> I[OpenAI]
    D --> J[Ollama]
    D --> K[Dummy]
    
    E --> L[Document Ingestion]
    E --> M[Knowledge Search]
    E --> N[Context Assembly]
    
    subgraph "Memory Operations"
        O[Personal Memory]
        P[Chat History]
        Q[Knowledge Base]
        R[Hybrid Search]
    end
    
    B --> O
    B --> P
    B --> Q
    B --> R
```

## Memory Capabilities

### 1. Personal Memory Operations
Store and retrieve personal information and preferences:
- Key-value storage with `Remember` and `Recall`
- Content storage with tags using `Store` and `Query`
- Flexible metadata and tagging system
- Session-based organization

### 2. Conversational Memory
Manage chat history and conversation context:
- Message storage with `AddMessage`
- History retrieval with `GetHistory`
- Session management with `NewSession` and `SetSession`
- Conversation context preservation

### 3. Knowledge Base Operations
Advanced document storage and retrieval:
- Document ingestion with `IngestDocument` and `IngestDocuments`
- Semantic search with `SearchKnowledge`
- Metadata-rich document management
- Chunking and processing pipeline

### 4. RAG (Retrieval-Augmented Generation)
Sophisticated context assembly for LLM enhancement:
- Hybrid search with `SearchAll`
- Context building with `BuildContext`
- Source attribution and citation
- Token-aware context management

### 5. Hybrid Search Capabilities
Combine multiple memory types for comprehensive retrieval:
- Personal memory + knowledge base search
- Configurable weighting and scoring
- Multi-modal content support
- Advanced filtering and ranking

## Memory Interface

The unified Memory interface provides a comprehensive API for all memory operations, from basic storage to advanced RAG implementations:

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

### Key Result Types

```go
// Basic memory result
type Result struct {
    Content   string    `json:"content"`
    Score     float32   `json:"score"`
    Tags      []string  `json:"tags,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

// Knowledge base result with rich metadata
type KnowledgeResult struct {
    Content    string         `json:"content"`
    Score      float32        `json:"score"`
    Source     string         `json:"source"`
    Title      string         `json:"title,omitempty"`
    DocumentID string         `json:"document_id"`
    Metadata   map[string]any `json:"metadata,omitempty"`
    Tags       []string       `json:"tags,omitempty"`
    CreatedAt  time.Time      `json:"created_at"`
    ChunkIndex int            `json:"chunk_index,omitempty"`
}

// Hybrid search combining personal memory and knowledge
type HybridResult struct {
    PersonalMemory []Result          `json:"personal_memory"`
    Knowledge      []KnowledgeResult `json:"knowledge"`
    Query          string            `json:"query"`
    TotalResults   int               `json:"total_results"`
    SearchTime     time.Duration     `json:"search_time"`
}

// RAG context ready for LLM consumption
type RAGContext struct {
    Query          string            `json:"query"`
    PersonalMemory []Result          `json:"personal_memory"`
    Knowledge      []KnowledgeResult `json:"knowledge"`
    ChatHistory    []Message         `json:"chat_history"`
    ContextText    string            `json:"context_text"` // Formatted for LLM
    Sources        []string          `json:"sources"`      // Source attribution
    TokenCount     int               `json:"token_count"`  // Estimated tokens
    Timestamp      time.Time         `json:"timestamp"`
}
```

## Memory Providers

AgenticGoKit supports multiple memory providers, each optimized for different use cases and deployment scenarios.

### 1. In-Memory Provider
Fast, non-persistent storage ideal for development and testing:

```go
// Create in-memory storage
memory, err := core.NewMemory(core.AgentMemoryConfig{
    Provider:   "memory",
    Connection: "memory",
    MaxResults: 10,
    Dimensions: 1536,
    AutoEmbed:  true,
})
if err != nil {
    log.Fatalf("Failed to create memory: %v", err)
}
```

**Characteristics:**
- Zero setup required
- High-speed access
- No persistence (data lost on restart)
- Perfect for development and testing

**Use Cases:**
- Development and prototyping
- Unit testing
- Temporary storage needs
- Stateless applications

### 2. PostgreSQL with pgvector
Production-ready vector storage with full SQL capabilities:

```go
// Create pgvector storage with full RAG support
memory, err := core.NewMemory(core.AgentMemoryConfig{
    Provider:                "pgvector",
    Connection:              "postgres://user:pass@localhost:5432/agentdb",
    EnableRAG:               true,
    EnableKnowledgeBase:     true,
    Dimensions:              1536,
    KnowledgeMaxResults:     20,
    KnowledgeScoreThreshold: 0.7,
    ChunkSize:               1000,
    ChunkOverlap:            200,
    
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
    },
})
```

**Characteristics:**
- ACID compliance and transactions
- SQL query capabilities
- Excellent performance with proper indexing
- Mature ecosystem and tooling

**Use Cases:**
- Production applications
- Complex relational queries
- High-availability requirements
- Enterprise deployments

### 3. Weaviate Vector Database
Specialized vector database with advanced ML capabilities:

```go
// Create Weaviate storage with advanced features
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
        CacheEmbeddings: true,
        MaxBatchSize:    100,
        TimeoutSeconds:  30,
    },
    
    Search: core.SearchConfigToml{
        HybridSearch:    true,
        KeywordWeight:   0.3,
        SemanticWeight:  0.7,
        EnableReranking: true,
    },
})
```

**Characteristics:**
- Native vector operations
- Built-in ML capabilities
- GraphQL API
- Multi-modal support

**Use Cases:**
- Large-scale vector search
- Multi-modal data (text, images, audio)
- Advanced filtering and faceting
- Real-time vector operations

### Provider Comparison

| Feature | In-Memory | pgvector | Weaviate |
|---------|-----------|----------|----------|
| **Persistence** | No | Yes | Yes |
| **Setup Complexity** | None | Medium | Medium |
| **Performance** | Highest | High | High |
| **Scalability** | Limited | High | Very High |
| **SQL Support** | No | Yes | No |
| **Multi-modal** | No | Limited | Yes |
| **Production Ready** | No | Yes | Yes |
| **Cost** | Free | Low | Medium |

## Quick Start Example

Here's a complete example showing the key memory operations:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create memory instance
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "memory", // Use "pgvector" or "weaviate" for production
        Connection: "memory",
        MaxResults: 10,
        Dimensions: 1536,
        EnableRAG:  true,
        
        Embedding: core.EmbeddingConfig{
            Provider: "dummy", // Use "openai" for production
            Model:    "text-embedding-3-small",
        },
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    defer memory.Close()
    
    ctx := context.Background()
    
    // 1. Personal Memory Operations
    err = memory.Remember(ctx, "user_name", "Alice")
    err = memory.Store(ctx, "User prefers detailed explanations", "preference")
    
    name, _ := memory.Recall(ctx, "user_name")
    fmt.Printf("User name: %v\n", name)
    
    results, _ := memory.Query(ctx, "detailed explanations", 5)
    fmt.Printf("Found %d preferences\n", len(results))
    
    // 2. Chat History Management
    sessionID := memory.NewSession()
    ctx = memory.SetSession(ctx, sessionID)
    
    memory.AddMessage(ctx, "user", "What is AgenticGoKit?")
    memory.AddMessage(ctx, "assistant", "AgenticGoKit is a Go framework for building multi-agent systems...")
    
    history, _ := memory.GetHistory(ctx, 10)
    fmt.Printf("Chat history: %d messages\n", len(history))
    
    // 3. Knowledge Base Operations
    doc := core.Document{
        ID:      "doc-1",
        Title:   "AgenticGoKit Overview",
        Content: "AgenticGoKit provides memory systems, orchestration, and LLM integration...",
        Source:  "documentation",
        Type:    core.DocumentTypeMarkdown,
        Tags:    []string{"framework", "go", "agents"},
    }
    
    err = memory.IngestDocument(ctx, doc)
    if err != nil {
        log.Printf("Failed to ingest document: %v", err)
    }
    
    // 4. RAG Operations
    knowledge, err := memory.SearchKnowledge(ctx, "AgenticGoKit framework",
        core.WithLimit(3),
        core.WithScoreThreshold(0.7),
    )
    if err == nil {
        fmt.Printf("Found %d knowledge results\n", len(knowledge))
    }
    
    // 5. Hybrid Search
    hybridResult, err := memory.SearchAll(ctx, "framework capabilities",
        core.WithLimit(5),
        core.WithIncludePersonal(true),
        core.WithIncludeKnowledge(true),
    )
    if err == nil {
        fmt.Printf("Hybrid search: %d personal + %d knowledge results\n", 
            len(hybridResult.PersonalMemory), len(hybridResult.Knowledge))
    }
    
    // 6. RAG Context Assembly
    ragContext, err := memory.BuildContext(ctx, "How do I use memory systems?",
        core.WithMaxTokens(2000),
        core.WithPersonalWeight(0.3),
        core.WithKnowledgeWeight(0.7),
        core.WithHistoryLimit(5),
        core.WithIncludeSources(true),
    )
    if err == nil {
        fmt.Printf("RAG context: %d tokens, %d sources\n", 
            ragContext.TokenCount, len(ragContext.Sources))
        fmt.Printf("Context text: %s\n", ragContext.ContextText[:100] + "...")
    }
}
```

## RAG (Retrieval-Augmented Generation)

AgenticGoKit provides built-in RAG capabilities through the Memory interface, making it easy to build knowledge-enhanced agents:

### RAG Architecture Flow

```mermaid
sequenceDiagram
    participant A as Agent
    participant M as Memory
    participant K as Knowledge Base
    participant E as Embedding Service
    participant L as LLM
    
    A->>M: SearchKnowledge(query)
    M->>E: Generate embedding
    M->>K: Vector search
    K-->>M: Relevant documents
    M-->>A: KnowledgeResult[]
    
    A->>M: BuildContext(query, options)
    M->>M: Assemble context
    M-->>A: RAGContext
    
    A->>L: Generate(context + query)
    L-->>A: Enhanced response
```

### Simple RAG Agent Implementation

```go
type RAGAgent struct {
    name   string
    memory core.Memory
    llm    core.LLMProvider
}

func (r *RAGAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    query, _ := state.Get("message")
    queryStr := query.(string)
    
    // Use built-in RAG context assembly
    ragContext, err := r.memory.BuildContext(ctx, queryStr,
        core.WithMaxTokens(3000),
        core.WithPersonalWeight(0.2),
        core.WithKnowledgeWeight(0.8),
        core.WithHistoryLimit(5),
        core.WithIncludeSources(true),
    )
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to build context: %w", err)
    }
    
    // Generate response with assembled context
    response, err := r.llm.Generate(ctx, ragContext.ContextText)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to generate response: %w", err)
    }
    
    // Store the interaction
    sessionID := event.GetSessionID()
    ctx = r.memory.SetSession(ctx, sessionID)
    r.memory.AddMessage(ctx, "user", queryStr)
    r.memory.AddMessage(ctx, "assistant", response)
    
    // Return result with metadata
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("sources", ragContext.Sources)
    outputState.Set("token_count", ragContext.TokenCount)
    outputState.Set("knowledge_used", len(ragContext.Knowledge))
    
    return core.AgentResult{OutputState: outputState}, nil
}
```

### Advanced RAG with Custom Search

```go
type AdvancedRAGAgent struct {
    name   string
    memory core.Memory
    llm    core.LLMProvider
}

func (a *AdvancedRAGAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    query, _ := state.Get("message")
    queryStr := query.(string)
    sessionID := event.GetSessionID()
    ctx = a.memory.SetSession(ctx, sessionID)
    
    // Perform hybrid search for comprehensive results
    hybridResult, err := a.memory.SearchAll(ctx, queryStr,
        core.WithLimit(10),
        core.WithScoreThreshold(0.7),
        core.WithIncludePersonal(true),
        core.WithIncludeKnowledge(true),
    )
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("hybrid search failed: %w", err)
    }
    
    // Build custom context with both personal and knowledge results
    context := a.buildCustomContext(queryStr, hybridResult)
    
    // Generate response
    response, err := a.llm.Generate(ctx, context)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("generation failed: %w", err)
    }
    
    // Store interaction
    a.memory.AddMessage(ctx, "user", queryStr)
    a.memory.AddMessage(ctx, "assistant", response)
    
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("personal_results", len(hybridResult.PersonalMemory))
    outputState.Set("knowledge_results", len(hybridResult.Knowledge))
    outputState.Set("search_time", hybridResult.SearchTime.String())
    
    return core.AgentResult{OutputState: outputState}, nil
}

func (a *AdvancedRAGAgent) buildCustomContext(query string, result *core.HybridResult) string {
    var builder strings.Builder
    
    builder.WriteString("Based on the following information:\n\n")
    
    // Add personal memory context
    if len(result.PersonalMemory) > 0 {
        builder.WriteString("Personal Context:\n")
        for i, item := range result.PersonalMemory {
            builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Content))
        }
        builder.WriteString("\n")
    }
    
    // Add knowledge base context
    if len(result.Knowledge) > 0 {
        builder.WriteString("Knowledge Base:\n")
        for i, item := range result.Knowledge {
            builder.WriteString(fmt.Sprintf("%d. %s", i+1, item.Content))
            if item.Source != "" {
                builder.WriteString(fmt.Sprintf(" (Source: %s)", item.Source))
            }
            builder.WriteString("\n")
        }
        builder.WriteString("\n")
    }
    
    builder.WriteString(fmt.Sprintf("Question: %s\n\n", query))
    builder.WriteString("Please provide a comprehensive answer based on the context above.")
    
    return builder.String()
}
```

## Memory Configuration

AgenticGoKit provides comprehensive configuration options for memory systems, embedding services, and RAG capabilities.

### Complete Configuration Structure

```go
type AgentMemoryConfig struct {
    // Core memory settings
    Provider   string `toml:"provider"`    // pgvector, weaviate, memory
    Connection string `toml:"connection"`  // Connection string
    MaxResults int    `toml:"max_results"` // Default: 10
    Dimensions int    `toml:"dimensions"`  // Default: 1536
    AutoEmbed  bool   `toml:"auto_embed"`  // Default: true

    // RAG-enhanced settings
    EnableKnowledgeBase     bool    `toml:"enable_knowledge_base"`     // Default: true
    KnowledgeMaxResults     int     `toml:"knowledge_max_results"`     // Default: 20
    KnowledgeScoreThreshold float32 `toml:"knowledge_score_threshold"` // Default: 0.7
    ChunkSize               int     `toml:"chunk_size"`                // Default: 1000
    ChunkOverlap            int     `toml:"chunk_overlap"`             // Default: 200

    // RAG context assembly settings
    EnableRAG           bool    `toml:"enable_rag"`             // Default: true
    RAGMaxContextTokens int     `toml:"rag_max_context_tokens"` // Default: 4000
    RAGPersonalWeight   float32 `toml:"rag_personal_weight"`    // Default: 0.3
    RAGKnowledgeWeight  float32 `toml:"rag_knowledge_weight"`   // Default: 0.7
    RAGIncludeSources   bool    `toml:"rag_include_sources"`    // Default: true

    // Nested configurations
    Documents core.DocumentConfig     `toml:"documents"`
    Embedding core.EmbeddingConfig    `toml:"embedding"`
    Search    core.SearchConfigToml   `toml:"search"`
}
```

### Embedding Provider Configuration

```go
// OpenAI Embeddings (Production)
embeddingConfig := core.EmbeddingConfig{
    Provider:        "openai",
    Model:           "text-embedding-3-small", // or "text-embedding-3-large"
    APIKey:          os.Getenv("OPENAI_API_KEY"),
    CacheEmbeddings: true,
    MaxBatchSize:    100,
    TimeoutSeconds:  30,
}

// Ollama Embeddings (Local)
embeddingConfig := core.EmbeddingConfig{
    Provider:        "ollama",
    Model:           "mxbai-embed-large",
    BaseURL:         "http://localhost:11434",
    CacheEmbeddings: true,
    MaxBatchSize:    32,
    TimeoutSeconds:  60,
}

// Dummy Embeddings (Development/Testing)
embeddingConfig := core.EmbeddingConfig{
    Provider: "dummy",
    Model:    "dummy-model",
}
```

### Document Processing Configuration

```go
documentConfig := core.DocumentConfig{
    AutoChunk:                true,
    SupportedTypes:           []string{"pdf", "txt", "md", "web", "code", "json"},
    MaxFileSize:              "10MB",
    EnableMetadataExtraction: true,
    EnableURLScraping:        true,
}
```

### Search Configuration

```go
searchConfig := core.SearchConfigToml{
    HybridSearch:         true,
    KeywordWeight:        0.3,
    SemanticWeight:       0.7,
    EnableReranking:      false,
    RerankingModel:       "",
    EnableQueryExpansion: false,
}
```

### Production Configuration Example

```go
// Production-ready configuration with pgvector
config := core.AgentMemoryConfig{
    // Core settings
    Provider:   "pgvector",
    Connection: "postgres://agent_user:secure_pass@localhost:5432/agentdb?sslmode=require",
    MaxResults: 10,
    Dimensions: 1536,
    AutoEmbed:  true,

    // RAG settings
    EnableRAG:               true,
    EnableKnowledgeBase:     true,
    KnowledgeMaxResults:     20,
    KnowledgeScoreThreshold: 0.75,
    ChunkSize:               1000,
    ChunkOverlap:            200,
    RAGMaxContextTokens:     4000,
    RAGPersonalWeight:       0.3,
    RAGKnowledgeWeight:      0.7,
    RAGIncludeSources:       true,

    // Embedding configuration
    Embedding: core.EmbeddingConfig{
        Provider:        "openai",
        Model:           "text-embedding-3-small",
        APIKey:          os.Getenv("OPENAI_API_KEY"),
        CacheEmbeddings: true,
        MaxBatchSize:    100,
        TimeoutSeconds:  30,
    },

    // Document processing
    Documents: core.DocumentConfig{
        AutoChunk:                true,
        SupportedTypes:           []string{"pdf", "txt", "md", "web", "code"},
        MaxFileSize:              "50MB",
        EnableMetadataExtraction: true,
        EnableURLScraping:        true,
    },

    // Search optimization
    Search: core.SearchConfigToml{
        HybridSearch:         true,
        KeywordWeight:        0.3,
        SemanticWeight:       0.7,
        EnableReranking:      true,
        RerankingModel:       "cross-encoder/ms-marco-MiniLM-L-6-v2",
        EnableQueryExpansion: false,
    },
}

memory, err := core.NewMemory(config)
if err != nil {
    log.Fatalf("Failed to create memory: %v", err)
}
```

### Development Configuration Example

```go
// Simple development configuration
config := core.AgentMemoryConfig{
    Provider:   "memory",
    Connection: "memory",
    MaxResults: 10,
    Dimensions: 1536,
    EnableRAG:  true,
    
    Embedding: core.EmbeddingConfig{
        Provider: "dummy",
        Model:    "dummy-model",
    },
    
    Documents: core.DocumentConfig{
        AutoChunk:      true,
        SupportedTypes: []string{"txt", "md"},
        MaxFileSize:    "1MB",
    },
}

memory, err := core.NewMemory(config)
```

## Memory Integration Patterns

### 1. Memory-Enabled Agent Implementation

```go
type MemoryEnabledAgent struct {
    name   string
    memory core.Memory
    llm    core.LLMProvider
}

func NewMemoryEnabledAgent(name string, memory core.Memory, llm core.LLMProvider) *MemoryEnabledAgent {
    return &MemoryEnabledAgent{
        name:   name,
        memory: memory,
        llm:    llm,
    }
}

func (m *MemoryEnabledAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    message, _ := state.Get("message")
    messageStr := message.(string)
    sessionID := event.GetSessionID()
    
    // Set session context
    ctx = m.memory.SetSession(ctx, sessionID)
    
    // Use built-in RAG context assembly
    ragContext, err := m.memory.BuildContext(ctx, messageStr,
        core.WithMaxTokens(3000),
        core.WithPersonalWeight(0.3),
        core.WithKnowledgeWeight(0.7),
        core.WithHistoryLimit(5),
        core.WithIncludeSources(true),
    )
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to build context: %w", err)
    }
    
    // Generate response with context
    response, err := m.llm.Generate(ctx, ragContext.ContextText)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to generate response: %w", err)
    }
    
    // Store the interaction
    m.memory.AddMessage(ctx, "user", messageStr)
    m.memory.AddMessage(ctx, "assistant", response)
    
    // Return enhanced result
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("sources", ragContext.Sources)
    outputState.Set("context_tokens", ragContext.TokenCount)
    outputState.Set("knowledge_count", len(ragContext.Knowledge))
    outputState.Set("personal_count", len(ragContext.PersonalMemory))
    
    return core.AgentResult{OutputState: outputState}, nil
}
```

### 2. Learning Agent with Memory

```go
type LearningAgent struct {
    name   string
    memory core.Memory
    llm    core.LLMProvider
}

func (l *LearningAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    message, _ := state.Get("message")
    messageStr := message.(string)
    sessionID := event.GetSessionID()
    ctx = l.memory.SetSession(ctx, sessionID)
    
    // Check for user feedback on previous responses
    if feedback, ok := state.Get("feedback"); ok {
        l.processFeedback(ctx, feedback.(string))
    }
    
    // Get personalized context based on user history
    personalResults, err := l.memory.Query(ctx, messageStr, 3)
    if err != nil {
        log.Printf("Failed to query personal memory: %v", err)
    }
    
    // Search knowledge base
    knowledgeResults, err := l.memory.SearchKnowledge(ctx, messageStr,
        core.WithLimit(5),
        core.WithScoreThreshold(0.7),
    )
    if err != nil {
        log.Printf("Failed to search knowledge: %v", err)
    }
    
    // Build personalized prompt
    prompt := l.buildPersonalizedPrompt(messageStr, personalResults, knowledgeResults)
    
    // Generate response
    response, err := l.llm.Generate(ctx, prompt)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Store interaction with learning metadata
    l.memory.Store(ctx, messageStr, "user-query", "learning")
    l.memory.Store(ctx, response, "agent-response", "learning")
    l.memory.AddMessage(ctx, "user", messageStr)
    l.memory.AddMessage(ctx, "assistant", response)
    
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("learning_active", true)
    
    return core.AgentResult{OutputState: outputState}, nil
}

func (l *LearningAgent) processFeedback(ctx context.Context, feedback string) {
    // Store feedback for learning
    l.memory.Store(ctx, feedback, "user-feedback", "learning")
    
    // Analyze feedback and adjust preferences
    if strings.Contains(strings.ToLower(feedback), "too technical") {
        l.memory.Remember(ctx, "communication_style", "simple")
    } else if strings.Contains(strings.ToLower(feedback), "more detail") {
        l.memory.Remember(ctx, "communication_style", "detailed")
    }
}

func (l *LearningAgent) buildPersonalizedPrompt(query string, personal []core.Result, knowledge []core.KnowledgeResult) string {
    var builder strings.Builder
    
    builder.WriteString("You are a helpful assistant that learns from user interactions.\n\n")
    
    // Add personal context
    if len(personal) > 0 {
        builder.WriteString("User Context:\n")
        for _, item := range personal {
            builder.WriteString(fmt.Sprintf("- %s\n", item.Content))
        }
        builder.WriteString("\n")
    }
    
    // Add knowledge context
    if len(knowledge) > 0 {
        builder.WriteString("Relevant Knowledge:\n")
        for _, item := range knowledge {
            builder.WriteString(fmt.Sprintf("- %s\n", item.Content))
        }
        builder.WriteString("\n")
    }
    
    builder.WriteString(fmt.Sprintf("User Question: %s\n\n", query))
    builder.WriteString("Please provide a helpful response based on the context above.")
    
    return builder.String()
}
```

### 3. Multi-Agent Memory Sharing

```go
type SharedMemorySystem struct {
    sharedMemory core.Memory
    agents       map[string]*MemoryAgent
}

type MemoryAgent struct {
    name         string
    personalMem  core.Memory
    sharedMem    core.Memory
    llm          core.LLMProvider
}

func NewSharedMemorySystem(sharedMemory core.Memory) *SharedMemorySystem {
    return &SharedMemorySystem{
        sharedMemory: sharedMemory,
        agents:       make(map[string]*MemoryAgent),
    }
}

func (s *SharedMemorySystem) AddAgent(name string, personalMemory core.Memory, llm core.LLMProvider) {
    agent := &MemoryAgent{
        name:        name,
        personalMem: personalMemory,
        sharedMem:   s.sharedMemory,
        llm:         llm,
    }
    s.agents[name] = agent
}

func (m *MemoryAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    message, _ := state.Get("message")
    messageStr := message.(string)
    
    // Search personal memory first
    personalResults, err := m.personalMem.Query(ctx, messageStr, 3)
    if err != nil {
        log.Printf("Personal memory search failed: %v", err)
    }
    
    // Search shared knowledge base
    sharedResults, err := m.sharedMem.SearchKnowledge(ctx, messageStr,
        core.WithLimit(5),
        core.WithScoreThreshold(0.7),
    )
    if err != nil {
        log.Printf("Shared memory search failed: %v", err)
    }
    
    // Build context from both sources
    context := m.buildHybridContext(messageStr, personalResults, sharedResults)
    
    // Generate response
    response, err := m.llm.Generate(ctx, context)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Store in personal memory
    sessionID := event.GetSessionID()
    ctx = m.personalMem.SetSession(ctx, sessionID)
    m.personalMem.AddMessage(ctx, "user", messageStr)
    m.personalMem.AddMessage(ctx, "assistant", response)
    
    // Share valuable insights to shared memory
    if m.isValuableInsight(response) {
        doc := core.Document{
            ID:      fmt.Sprintf("%s-%d", m.name, time.Now().Unix()),
            Content: response,
            Source:  m.name,
            Type:    core.DocumentTypeText,
            Metadata: map[string]any{
                "agent":     m.name,
                "query":     messageStr,
                "timestamp": time.Now(),
            },
            Tags: []string{"agent-insight", m.name},
        }
        m.sharedMem.IngestDocument(ctx, doc)
    }
    
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("personal_results", len(personalResults))
    outputState.Set("shared_results", len(sharedResults))
    
    return core.AgentResult{OutputState: outputState}, nil
}

func (m *MemoryAgent) buildHybridContext(query string, personal []core.Result, shared []core.KnowledgeResult) string {
    var builder strings.Builder
    
    builder.WriteString(fmt.Sprintf("Agent: %s\n\n", m.name))
    
    if len(personal) > 0 {
        builder.WriteString("Personal Context:\n")
        for _, item := range personal {
            builder.WriteString(fmt.Sprintf("- %s\n", item.Content))
        }
        builder.WriteString("\n")
    }
    
    if len(shared) > 0 {
        builder.WriteString("Shared Knowledge:\n")
        for _, item := range shared {
            builder.WriteString(fmt.Sprintf("- %s (from %s)\n", item.Content, item.Source))
        }
        builder.WriteString("\n")
    }
    
    builder.WriteString(fmt.Sprintf("Query: %s\n\n", query))
    builder.WriteString("Provide a helpful response based on the available context.")
    
    return builder.String()
}

func (m *MemoryAgent) isValuableInsight(response string) bool {
    // Simple heuristic - in practice, use more sophisticated analysis
    return len(response) > 100 && 
           (strings.Contains(response, "solution") || 
            strings.Contains(response, "approach") ||
            strings.Contains(response, "pattern"))
}
```

## Tutorial Series Structure

This comprehensive tutorial series guides you through AgenticGoKit's memory systems from basic concepts to advanced production deployments:

### 1. [Basic Memory Operations](basic-memory.md)
Learn the fundamentals of the Memory interface and core operations:
- Memory interface methods (Store, Query, Remember, Recall)
- Session management and chat history
- In-memory provider setup and usage
- Basic memory-enabled agent patterns

### 2. [Vector Databases](vector-databases.md)
Set up production-ready vector storage with pgvector and Weaviate:
- pgvector configuration and optimization
- Weaviate setup and advanced features
- Embedding service integration (OpenAI, Ollama)
- Performance tuning and scaling strategies

### 3. [Document Ingestion](document-ingestion.md)
Master document processing and knowledge base creation:
- Document structure and processing pipeline
- Chunking strategies and metadata extraction
- Batch ingestion and processing optimization
- Multi-format document support

### 4. [RAG Implementation](rag-implementation.md)
Build sophisticated retrieval-augmented generation systems:
- SearchKnowledge and BuildContext methods
- Hybrid search with SearchAll
- Advanced RAG patterns and optimization
- Context assembly and prompt engineering

### 5. [Knowledge Bases](knowledge-bases.md)
Create and manage comprehensive knowledge systems:
- Knowledge base architecture and design
- Advanced search patterns and filtering
- Multi-modal content management
- Production deployment strategies

### 6. [Memory Optimization](memory-optimization.md)
Optimize performance and scale memory systems:
- Configuration optimization and tuning
- Caching strategies and resource management
- Monitoring and observability patterns
- Scaling and production considerations

## Best Practices

### 1. Memory System Design

**Separation of Concerns**
- Use personal memory for user-specific data
- Use knowledge base for shared, factual information
- Implement proper session management
- Design clear data lifecycle policies

**Performance Optimization**
- Choose appropriate embedding dimensions
- Configure proper chunk sizes and overlap
- Implement caching for frequent queries
- Monitor memory usage and search performance

**Data Quality**
- Validate document quality before ingestion
- Use consistent metadata schemas
- Implement proper error handling
- Regular cleanup of outdated information

### 2. RAG Implementation Best Practices

**Context Management**
- Balance context size with relevance
- Use appropriate token limits
- Implement source attribution
- Handle context overflow gracefully

**Search Optimization**
- Use hybrid search for better results
- Implement proper score thresholds
- Consider query enhancement techniques
- Use metadata filtering effectively

**Response Quality**
- Validate retrieved context relevance
- Implement fallback strategies
- Monitor response quality metrics
- Collect user feedback for improvement

### 3. Production Deployment

**Scalability**
- Use production-ready providers (pgvector, Weaviate)
- Implement horizontal scaling strategies
- Monitor resource usage and costs
- Plan for data growth and retention

**Reliability**
- Implement proper error handling and retries
- Use connection pooling and timeouts
- Monitor system health and performance
- Implement backup and recovery procedures

**Security**
- Secure database connections
- Implement proper access controls
- Protect API keys and credentials
- Consider data privacy and compliance requirements

## Common Use Cases

### 1. Conversational AI Systems
Build intelligent chatbots and virtual assistants:
- **Chat History**: Maintain conversation context across sessions
- **User Preferences**: Learn and remember user communication styles
- **Personalization**: Adapt responses based on user history
- **Context Awareness**: Provide relevant responses using conversation context

```go
// Example: Conversational AI with memory
memory, _ := core.NewMemory(core.AgentMemoryConfig{
    Provider: "pgvector",
    Connection: "postgres://...",
    EnableRAG: true,
})

// Store user preferences
memory.Remember(ctx, "communication_style", "technical")
memory.Store(ctx, "User prefers code examples", "preference")

// Use in conversation
ragContext, _ := memory.BuildContext(ctx, userQuery,
    core.WithPersonalWeight(0.4), // Higher weight for personalization
    core.WithKnowledgeWeight(0.6),
)
```

### 2. Knowledge Management Systems
Create intelligent document Q&A and research assistants:
- **Document Ingestion**: Process and index large document collections
- **Semantic Search**: Find relevant information across documents
- **Source Attribution**: Track and cite information sources
- **Multi-format Support**: Handle PDFs, markdown, web content, and code

```go
// Example: Document Q&A system
docs := []core.Document{
    {
        Title: "API Documentation",
        Content: "AgenticGoKit provides...",
        Source: "docs/api.md",
        Type: core.DocumentTypeMarkdown,
    },
}

memory.IngestDocuments(ctx, docs)

// Search with source filtering
results, _ := memory.SearchKnowledge(ctx, "How to create agents?",
    core.WithSources([]string{"docs/"}),
    core.WithLimit(5),
)
```

### 3. Learning and Adaptive Systems
Build agents that improve through experience:
- **Feedback Processing**: Learn from user corrections and feedback
- **Pattern Recognition**: Identify successful interaction patterns
- **Continuous Improvement**: Adapt behavior based on outcomes
- **Experience Replay**: Use past experiences to improve future responses

```go
// Example: Learning agent
type LearningAgent struct {
    memory core.Memory
    llm    core.LLMProvider
}

func (a *LearningAgent) processFeedback(ctx context.Context, feedback string) {
    // Store feedback for learning
    a.memory.Store(ctx, feedback, "feedback", "learning")
    
    // Adjust behavior based on feedback
    if strings.Contains(feedback, "too verbose") {
        a.memory.Remember(ctx, "response_style", "concise")
    }
}
```

### 4. Multi-Agent Collaboration
Enable agents to share knowledge and coordinate:
- **Shared Knowledge**: Common knowledge base across agents
- **Inter-agent Communication**: Share insights and learnings
- **Collaborative Learning**: Agents learn from each other's experiences
- **Distributed Memory**: Scale memory across multiple agents

```go
// Example: Multi-agent knowledge sharing
sharedMemory, _ := core.NewMemory(core.AgentMemoryConfig{
    Provider: "weaviate",
    Connection: "http://weaviate:8080",
})

// Agents share valuable insights
if isValuableInsight(response) {
    doc := core.Document{
        Content: response,
        Source: agentName,
        Metadata: map[string]any{
            "agent": agentName,
            "confidence": 0.9,
        },
    }
    sharedMemory.IngestDocument(ctx, doc)
}
```

## Getting Started

### Quick Setup for Development

```go
// 1. Create a simple memory system
memory, err := core.NewMemory(core.AgentMemoryConfig{
    Provider: "memory",
    Connection: "memory",
    EnableRAG: true,
    Embedding: core.EmbeddingConfig{
        Provider: "dummy",
    },
})

// 2. Add some knowledge
doc := core.Document{
    Content: "AgenticGoKit is a Go framework for multi-agent systems",
    Type: core.DocumentTypeText,
}
memory.IngestDocument(ctx, doc)

// 3. Use RAG in your agent
ragContext, _ := memory.BuildContext(ctx, "What is AgenticGoKit?")
response, _ := llm.Generate(ctx, ragContext.ContextText)
```

### Production Setup

```go
// Production configuration with pgvector
memory, err := core.NewMemory(core.AgentMemoryConfig{
    Provider: "pgvector",
    Connection: "postgres://user:pass@localhost:5432/agentdb",
    EnableRAG: true,
    Dimensions: 1536,
    
    Embedding: core.EmbeddingConfig{
        Provider: "openai",
        Model: "text-embedding-3-small",
        APIKey: os.Getenv("OPENAI_API_KEY"),
    },
    
    Documents: core.DocumentConfig{
        AutoChunk: true,
        SupportedTypes: []string{"pdf", "txt", "md", "web"},
        MaxFileSize: "10MB",
    },
})
```

## Next Steps

Ready to dive deeper? Follow our progressive tutorial series:

### 🚀 **Start Here**: [Basic Memory Operations](basic-memory.md)
Learn the Memory interface fundamentals and core operations

### 🗄️ **Production Storage**: [Vector Databases](vector-databases.md)
Set up pgvector or Weaviate for production deployments

### 📄 **Knowledge Processing**: [Document Ingestion](document-ingestion.md)
Master document processing and knowledge base creation

### 🧠 **Advanced AI**: [RAG Implementation](rag-implementation.md)
Build sophisticated retrieval-augmented generation systems

### 🏗️ **Enterprise Scale**: [Knowledge Bases](knowledge-bases.md)
Create and manage large-scale knowledge systems

### ⚡ **Performance**: [Memory Optimization](memory-optimization.md)
Optimize and scale your memory systems

## Additional Resources

- **[Core Concepts](../core-concepts/README.md)** - Understanding AgenticGoKit fundamentals
- **[API Reference](../../reference/api/)** - Complete API documentation
- **[Examples Repository](../../examples/)** - Working code examples
- **[Configuration Guide](../../reference/configuration.md)** - Detailed configuration options

---

*Memory systems are the foundation of intelligent agents. Start with the basics and progressively build more sophisticated systems as your needs grow.*
