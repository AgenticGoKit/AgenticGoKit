# Memory & RAG Example

Build knowledge-powered agents with memory integration and retrieval-augmented generation.

---

## Overview

This example demonstrates:
- Integrating memory providers
- Building RAG-powered Q&A systems
- Session-scoped memory
- Vector search and retrieval

---

## Complete Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create agent with memory and RAG
    agent, err := v1beta.NewBuilder("QAAgent").
        WithLLM("openai", "gpt-4").
        WithMemory(
            v1beta.WithMemoryProvider("pgvector"),
            v1beta.WithSessionScoped(),
            v1beta.WithRAG(4000, 0.3, 0.7), // contextSize, diversityWeight, relevanceWeight
        ).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Store knowledge
    fmt.Println("Storing knowledge...")
    documents := []string{
        "Go is a statically typed, compiled language designed for building reliable and efficient software.",
        "Goroutines are lightweight threads managed by the Go runtime, making concurrent programming easy.",
        "Channels in Go provide a way for goroutines to communicate with each other and synchronize execution.",
    }

    for _, doc := range documents {
        if err := agent.StoreMemory(ctx, doc); err != nil {
            log.Printf("Failed to store: %v", err)
        }
    }

    // Query with RAG
    fmt.Println("\nQuerying with RAG:")
    result, err := agent.Run(ctx, "What are goroutines?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Content)

    // Follow-up question with context
    result, err = agent.Run(ctx, "How do they communicate?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Content)
}
```

---

## Memory Providers

### PostgreSQL (pgvector)

```go
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithConnectionString("postgres://user:pass@localhost/db"),
    ).
    Build()
```

**Environment Variables:**
```bash
export PG_CONNECTION_STRING="postgres://user:pass@localhost/db"
```

### In-Memory (Development)

```go
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
    ).
    Build()
```

### Weaviate

```go
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("weaviate"),
        v1beta.WithConnectionString("http://localhost:8080"),
    ).
    Build()
```

---

## RAG Configuration

### Basic RAG

```go
WithMemory(
    v1beta.WithMemoryProvider("pgvector"),
    v1beta.WithRAG(4000, 0.3, 0.7), // context size, diversity, relevance
)
```

### Advanced RAG

```go
WithMemory(
    v1beta.WithMemoryProvider("pgvector"),
    v1beta.WithRAG(8000, 0.5, 0.5), // Larger context, balanced weights
    v1beta.WithSimilarityThreshold(0.75), // Minimum similarity score
    v1beta.WithMaxRetrieval(10), // Max documents to retrieve
)
```

---

## Memory Operations

### Store Documents

```go
// Single document
agent.StoreMemory(ctx, "Go was created at Google in 2007")

// Batch storage
documents := []string{
    "Document 1",
    "Document 2",
    "Document 3",
}
for _, doc := range documents {
    agent.StoreMemory(ctx, doc)
}
```

### Query with Context

```go
// Agent automatically retrieves relevant documents
result, _ := agent.Run(ctx, "Tell me about Go's history")

// The LLM receives both the query and retrieved documents
```

### Search Memory

```go
// Direct memory search
results, err := agent.SearchMemory(ctx, "goroutines", 5)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("Score: %.2f - %s\n", result.Score, result.Content)
}
```

---

## Session Management

### Session-Scoped Memory

```go
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithSessionScoped(), // Each session has isolated memory
    ).
    Build()

// Create session context
sessionCtx := v1beta.WithSession(ctx, "user-123")

// This memory is scoped to user-123
agent.Run(sessionCtx, "Remember my name is Alice")
```

### Agent-Scoped Memory

```go
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithAgentScoped(), // Shared across all sessions
    ).
    Build()

// Memory shared by all users
agent.Run(ctx, "What is Go?")
```

---

## Real-World Examples

### Knowledge Base Q&A

```go
func buildKnowledgeBase() (v1beta.Agent, error) {
    agent, err := v1beta.NewBuilder("KnowledgeBase").
        WithLLM("openai", "gpt-4").
        WithMemory(
            v1beta.WithMemoryProvider("pgvector"),
            v1beta.WithRAG(6000, 0.2, 0.8), // High relevance weight
            v1beta.WithSimilarityThreshold(0.7),
        ).
        Build()
    if err != nil {
        return nil, err
    }

    // Load knowledge base
    docs := loadDocumentsFromFiles("./knowledge/")
    for _, doc := range docs {
        agent.StoreMemory(context.Background(), doc)
    }

    return agent, nil
}
```

### Conversational Agent with History

```go
agent, _ := v1beta.NewBuilder("ChatBot").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
        v1beta.WithSessionScoped(),
        v1beta.WithConversationHistory(10), // Keep last 10 turns
    ).
    Build()

sessionCtx := v1beta.WithSession(ctx, userID)

// Maintains conversation context
agent.Run(sessionCtx, "My name is John")
agent.Run(sessionCtx, "What's my name?") // Remembers "John"
```

### Document Analysis

```go
analyzer, _ := v1beta.NewBuilder("Analyzer").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("weaviate"),
        v1beta.WithRAG(8000, 0.3, 0.7),
    ).
    Build()

// Index documents
for _, doc := range documents {
    analyzer.StoreMemory(ctx, doc)
}

// Analyze with RAG
result, _ := analyzer.Run(ctx, "Summarize the main themes")
result, _ = analyzer.Run(ctx, "What are the key findings?")
result, _ = analyzer.Run(ctx, "Compare sections 2 and 3")
```

---

## Performance Tips

### Batch Indexing

```go
// More efficient than individual stores
func batchStore(agent v1beta.Agent, docs []string) error {
    batch := agent.BeginBatch()
    defer batch.Commit()
    
    for _, doc := range docs {
        if err := batch.StoreMemory(context.Background(), doc); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Optimize Retrieval

```go
WithMemory(
    v1beta.WithMemoryProvider("pgvector"),
    v1beta.WithRAG(4000, 0.3, 0.7),
    v1beta.WithMaxRetrieval(5), // Limit to top 5 results
    v1beta.WithCaching(true), // Enable result caching
)
```

---

## Running the Example

### Prerequisites

```bash
# Install v1beta
go get github.com/agenticgokit/agenticgokit/v1beta

# Setup PostgreSQL with pgvector
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password ankane/pgvector

# Set environment
export OPENAI_API_KEY="sk-..."
export PG_CONNECTION_STRING="postgres://postgres:password@localhost/postgres"
```

### Execute

```bash
go run main.go
```

---

## Next Steps

- **[Custom Handlers](./custom-handlers.md)** - Add custom memory logic
- **[Sequential Workflow](./workflow-sequential.md)** - Chain RAG-powered agents
- **[Basic Agent](./basic-agent.md)** - Start with simple agents

---

## Related Documentation

- [Memory & RAG Guide](../memory-and-rag.md) - Complete memory documentation
- [Configuration](../configuration.md) - Memory configuration options
- [Performance](../performance.md) - Memory optimization strategies
