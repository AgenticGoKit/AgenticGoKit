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
    // Create memory provider
    memory, err := v1beta.NewMemoryBuilder().
        WithProvider("memory").  // Use "pgvector" for production
        WithRAGConfig(&v1beta.RAGConfig{
            MaxTokens:       4000,
            PersonalWeight:  0.3,
            KnowledgeWeight: 0.7,
        }).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Create agent with memory and RAG
    agent, err := v1beta.NewBuilder("QAAgent").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{
                Provider: "openai",
                Model:    "gpt-4",
            },
        }).
        WithMemory(
            v1beta.WithMemoryProvider("memory"),
            v1beta.WithSessionScoped(),
            v1beta.WithRAG(4000, 0.3, 0.7), // contextSize, personalWeight, knowledgeWeight
        ).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Store knowledge using the Memory interface
    fmt.Println("Storing knowledge...")
    documents := []string{
        "Go is a statically typed, compiled language designed for building reliable and efficient software.",
        "Goroutines are lightweight threads managed by the Go runtime, making concurrent programming easy.",
        "Channels in Go provide a way for goroutines to communicate with each other and synchronize execution.",
    }

    for _, doc := range documents {
        if err := memory.Store(ctx, doc); err != nil {
            log.Printf("Failed to store: %v", err)
        }
    }

    // Query with RAG - agent uses the memory automatically when configured
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
// Create memory with pgvector (connection via environment variable)
memory, _ := v1beta.NewMemoryBuilder().
    WithProvider("pgvector").
    Build()

agent, _ := v1beta.NewBuilder("Agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
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
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
    ).
    Build()
```

### Weaviate

```go
// Create memory with Weaviate (connection via environment variable)
memory, _ := v1beta.NewMemoryBuilder().
    WithProvider("weaviate").
    Build()

agent, _ := v1beta.NewBuilder("Agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithMemory(
        v1beta.WithMemoryProvider("weaviate"),
    ).
    Build()
```

**Environment Variables:**
```bash
export WEAVIATE_URL="http://localhost:8080"
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
// Advanced RAG is configured via RAGConfig
memory, _ := v1beta.NewMemoryBuilder().
    WithProvider("pgvector").
    WithRAGConfig(&v1beta.RAGConfig{
        MaxTokens:       8000,       // Larger context
        PersonalWeight:  0.5,        // Balanced weights
        KnowledgeWeight: 0.5,
        HistoryLimit:    10,         // Limit conversation history
    }).
    Build()

// Use with agent builder
WithMemory(
    v1beta.WithMemoryProvider("pgvector"),
    v1beta.WithRAG(8000, 0.5, 0.5), // Larger context, balanced weights
)
```

---

## Memory Operations

### Store Documents

Use the `Memory` interface to store documents:

```go
// Create memory instance
memory, _ := v1beta.NewMemoryBuilder().
    WithProvider("memory").
    Build()

ctx := context.Background()

// Single document
memory.Store(ctx, "Go was created at Google in 2007")

// Batch storage
documents := []string{
    "Document 1",
    "Document 2",
    "Document 3",
}
for _, doc := range documents {
    memory.Store(ctx, doc)
}
```

### Query with Context

```go
// Agent automatically retrieves relevant documents when configured with memory
result, _ := agent.Run(ctx, "Tell me about Go's history")

// The LLM receives both the query and retrieved documents
```

### Search Memory Directly

```go
// Direct memory search using Query method
results, err := memory.Query(ctx, "goroutines", v1beta.WithLimit(5))
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
// Create memory instance
memory, _ := v1beta.NewMemoryBuilder().
    WithProvider("memory").
    Build()

agent, _ := v1beta.NewBuilder("Agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
        v1beta.WithSessionScoped(), // Each session has isolated memory
    ).
    Build()

// Create session context using Memory interface
sessionCtx := memory.SetSession(ctx, "user-123")

// Conversation memory is scoped to user-123
agent.Run(sessionCtx, "Remember my name is Alice")
agent.Run(sessionCtx, "What is my name?") // Remembers "Alice"
```

### Shared Memory Across Sessions

```go
agent, _ := v1beta.NewBuilder("Agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
        // Without WithSessionScoped(), memory is shared
    ).
    Build()

// Memory shared by all users
agent.Run(ctx, "What is Go?")
```

---

## Real-World Examples

### Knowledge Base Q&A

```go
func buildKnowledgeBase() (v1beta.Agent, v1beta.Memory, error) {
    // Create memory for storing knowledge
    memory, err := v1beta.NewMemoryBuilder().
        WithProvider("pgvector").
        WithRAGConfig(&v1beta.RAGConfig{
            MaxTokens:       6000,
            PersonalWeight:  0.2,
            KnowledgeWeight: 0.8, // High relevance weight
        }).
        Build()
    if err != nil {
        return nil, nil, err
    }

    // Create agent with memory
    agent, err := v1beta.NewBuilder("KnowledgeBase").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        WithMemory(
            v1beta.WithMemoryProvider("pgvector"),
            v1beta.WithRAG(6000, 0.2, 0.8),
        ).
        Build()
    if err != nil {
        return nil, nil, err
    }

    // Load knowledge base using Memory interface
    docs := loadDocumentsFromFiles("./knowledge/")
    for _, doc := range docs {
        memory.Store(context.Background(), doc)
    }

    return agent, memory, nil
}
```

### Conversational Agent with History

```go
// Create memory for session management
memory, _ := v1beta.NewMemoryBuilder().
    WithProvider("memory").
    Build()

agent, _ := v1beta.NewBuilder("ChatBot").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
        v1beta.WithSessionScoped(),
        v1beta.WithRAG(4000, 0.5, 0.5), // Include conversation history
    ).
    Build()

sessionCtx := memory.SetSession(ctx, userID)

// Maintains conversation context
agent.Run(sessionCtx, "My name is John")
agent.Run(sessionCtx, "What's my name?") // Remembers "John"
```

### Document Analysis

```go
// Create memory for document storage
memory, _ := v1beta.NewMemoryBuilder().
    WithProvider("weaviate").
    WithRAGConfig(&v1beta.RAGConfig{
        MaxTokens:       8000,
        PersonalWeight:  0.3,
        KnowledgeWeight: 0.7,
    }).
    Build()

analyzer, _ := v1beta.NewBuilder("Analyzer").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithMemory(
        v1beta.WithMemoryProvider("weaviate"),
        v1beta.WithRAG(8000, 0.3, 0.7),
    ).
    Build()

ctx := context.Background()

// Index documents using Memory interface
for _, doc := range documents {
    memory.Store(ctx, doc)
}

// Analyze with RAG - agent retrieves relevant context automatically
result, _ := analyzer.Run(ctx, "Summarize the main themes")
result, _ = analyzer.Run(ctx, "What are the key findings?")
result, _ = analyzer.Run(ctx, "Compare sections 2 and 3")
```

---

## Performance Tips

### Batch Indexing

```go
// Store multiple documents efficiently
func batchStore(memory v1beta.Memory, docs []string) error {
    ctx := context.Background()
    
    for _, doc := range docs {
        if err := memory.Store(ctx, doc); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Optimize Retrieval

```go
// Configure memory with RAG options for optimized retrieval
memory, _ := v1beta.NewMemoryBuilder().
    WithProvider("pgvector").
    WithRAGConfig(&v1beta.RAGConfig{
        MaxTokens:       4000,
        PersonalWeight:  0.3,
        KnowledgeWeight: 0.7,
        HistoryLimit:    5, // Limit retrieved history
    }).
    Build()

// Use with agent
agent, _ := v1beta.NewBuilder("Agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithRAG(4000, 0.3, 0.7),
    ).
    Build()
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
