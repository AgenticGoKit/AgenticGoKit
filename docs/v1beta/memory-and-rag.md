# Memory and RAG Integration

Learn how to add persistent memory and Retrieval-Augmented Generation (RAG) capabilities to your agents for context-aware conversations and knowledge-based responses.

---

## 🎯 Overview

AgenticGoKit v1beta provides flexible memory integration that enables:

- **Conversation History** - Remember past interactions
- **Knowledge Retrieval** - Access external knowledge bases
- **Context Augmentation** - Enhance prompts with relevant information
- **Vector Search** - Semantic similarity matching
- **Production-Ready Backends** - In-memory and PostgreSQL with pgvector

---

## 🧠 Memory Hierarchy & Preferences

AgenticGoKit uses a layered approach to memory. When multiple memory types are enabled, they are prioritized and mixed to provide the best context for the LLM.

### 1. Default Memory Behavior ("Batteries Included")

Starting from `v1beta`, memory is **enabled by default** using the `chromem` provider (an embedded vector database). This provides an "out-of-the-box" experience where agents can remember recent conversation history and facts without any extra configuration.

| Scenario | Configuration | Resulting Memory |
| :--- | :--- | :--- |
| **Implicit (Default)** | Omitted `Memory` config | **`chromem` enabled** |
| **Explicit (Custom)** | `Memory: { Provider: "pgvector" }` | `pgvector` enabled |
| **Opting Out** | `Memory: { Enabled: false }` | `nil` (No memory) |

> [!NOTE]
> If you want to completely disable memory for an agent (making it purely ephemeral), you must explicitly set `Enabled: false` in your `MemoryConfig`.

### 2. The Three Tiers of Memory

| Memory Type | Scope | How it Works | Primary Purpose |
| :--- | :--- | :--- | :--- |
| **Chat History** | Session | Recent messages in the current conversation. | Maintaining flow and following pronouns (e.g., "What did I just say?"). |
| **Personal Memory** | User/Global | Past interactions across sessions, retrieved by semantic similarity. | Recalling user preferences or facts mentioned long ago (e.g., "favorite color"). |
| **Knowledge Base** | Global | Static documents or data ingested via `IngestDocuments`. | Providing factual, external context the agent wasn't trained on (e.g., "Product Specs"). |

### 2. Priority & Ordering

In the final prompt sent to the LLM, the information is typically ordered from most recent to most relevant:

1.  **Chat History**: Placed at the top to establish the immediate conversation state.
2.  **RAG Context**: Mixed results from Personal Memory and Knowledge Base.
3.  **User Query**: The actual question at the very bottom.

### 3. Tuning Preferences with Weights

When using RAG, you can control the "trust level" between your Knowledge Base and the User's Personal Memory using weights:

```go
v1beta.WithRAG(
    2000, // maxTokens
    0.3,  // personalWeight (User facts)
    0.7,  // knowledgeWeight (Official docs)
)
```

- **Higher Knowledge Weight**: Best for "Expert" assistants where official data is the source of truth.
- **Higher Personal Weight**: Best for "Personal" assistants that should adapt to the user's history and style.

---


## 🚀 Quick Start

### Basic In-Memory Storage

```go
package main

import (
    "context"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem" // Register provider
)

func main() {
    // Create agent with memory
    agent, err := v1beta.NewBuilder("MemoryAgent").
        WithPreset(v1beta.ChatAgent).
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{
                Provider: "openai",
                Model:    "gpt-4",
            },
        }).
        WithMemory(
            v1beta.WithMemoryProvider("chromem"),
        ).
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    // First interaction - agent has no context
    result1, _ := agent.Run(context.Background(), "My name is Alice")
    // Output: "Nice to meet you, Alice!"
    
    // Second interaction - agent remembers
    result2, _ := agent.Run(context.Background(), "What's my name?")
    // Output: "Your name is Alice."
}
```

---

## 💾 Memory Providers

AgenticGoKit v1beta includes production-ready memory providers with vector search capabilities:

### 1. Chromem (Default - Embedded Vector Database)

The default provider using chromem-go for embedded vector search:

```go
import (
    "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem"
)

agent, _ := v1beta.NewBuilder("DevAgent").
    WithPreset(v1beta.ChatAgent).
    // Memory defaults to chromem - no explicit configuration needed
    Build()
```

**Pros:**
- ✅ No external dependencies
- ✅ Fast and simple
- ✅ True vector embeddings for semantic search
- ✅ Optional persistence to disk
- ✅ Good for development and production

**Cons:**
- ⚠️ Single-instance only (not distributed)
- ⚠️ Limited to in-process memory

**Use Cases:**
- Local development
- Single-instance deployments
- Embedded applications
- Prototyping and demos

### 2. PostgreSQL with pgvector (Production)

Enterprise-grade storage with vector search capabilities:

```go
import (
    "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/pgvector"
)

agent, _ := v1beta.NewBuilder("ProductionAgent").
    WithPreset(v1beta.ResearchAgent).
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
    ).
    Build()
```

**Pros:**
- ✅ Persistent storage
- ✅ Horizontally scalable
- ✅ Vector similarity search
- ✅ ACID compliance
- ✅ Production-ready with connection pooling

**Database Setup:**

```sql
-- Install pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Tables are automatically created by the provider:
-- - personal_memory (conversation context)
-- - key_value_store (structured data)
-- - chat_history (message history)
-- - documents (document metadata)
-- - knowledge_base (RAG embeddings)
```

**Configuration via Environment:**

```bash
# Set database connection
export AGENTFLOW_MEMORY_CONNECTION="postgresql://user:password@localhost:5432/agentdb"
export AGENTFLOW_MEMORY_PROVIDER="pgvector"
export AGENTFLOW_MEMORY_DIMENSIONS=1536  # OpenAI ada-002
```

**Configuration via TOML:**

```toml
[memory]
provider = "pgvector"
connection = "postgresql://user:password@localhost:5432/agentdb"
max_results = 10
dimensions = 1536
auto_embed = true

[memory.embedding]
provider = "openai"
model = "text-embedding-ada-002"
cache_embeddings = true
max_batch_size = 100
timeout_seconds = 30
```

---

## 🔍 RAG (Retrieval-Augmented Generation)

RAG enhances agent responses by retrieving relevant context from memory before generating responses.

### Basic RAG Configuration

```go
import (
    "context"
    "github.com/agenticgokit/agenticgokit/v1beta"
    "github.com/agenticgokit/agenticgokit/core"
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem"
)

func main() {
    agent, _ := v1beta.NewBuilder("RAGAgent").
        WithPreset(v1beta.ResearchAgent).
        WithMemory(
            v1beta.WithMemoryProvider("chromem"),
            v1beta.WithRAG(2000, 0.3, 0.7), // maxTokens, personalWeight, knowledgeWeight
        ).
        Build()
    
    // Get memory provider to ingest documents
    memProvider := core.GetMemoryProvider("memory")
    
    // Ingest knowledge
    ctx := context.Background()
    doc := core.Document{
        ID:      "geo-001",
        Title:   "World Capitals",
        Content: "The capital of France is Paris. The capital of Germany is Berlin.",
        Source:  "geography-textbook",
        Type:    core.DocumentTypeText,
        Tags:    []string{"geography", "capitals"},
    }
    
    memProvider.IngestDocument(ctx, doc)
    
    // Agent automatically retrieves relevant context
    result, _ := agent.Run(ctx, "What is the capital of France?")
    // Agent uses retrieved knowledge to answer accurately
}
```

### RAG Configuration Options

```go
// Configure RAG weights and limits
agent, _ := v1beta.NewBuilder("SmartAgent").
    WithPreset(v1beta.ResearchAgent).
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithRAG(
            4000, // maxTokens - maximum context size
            0.3,  // personalWeight - weight for conversation history
            0.7,  // knowledgeWeight - weight for knowledge base
        ),
    ).
    Build()
```

**RAG Parameters:**
- **maxTokens**: Maximum tokens for RAG context (2000-8000 recommended)
- **personalWeight**: Weight for personal memory/conversation history (0.0-1.0)
- **knowledgeWeight**: Weight for knowledge base documents (0.0-1.0)

---

## 📚 Document Ingestion

Load documents into the knowledge base for RAG:

### Ingesting Text Documents

```go
import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "github.com/agenticgokit/agenticgokit/core"
)

func ingestTextFile(memProvider core.Memory, filePath string) error {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return err
    }
    
    // Split into chunks for better retrieval
    chunks := chunkText(string(content), 1000, 200) // size, overlap
    
    ctx := context.Background()
    for i, chunk := range chunks {
        doc := core.Document{
            ID:         fmt.Sprintf("%s-chunk-%d", filePath, i),
            Title:      filepath.Base(filePath),
            Content:    chunk,
            Source:     filePath,
            Type:       core.DocumentTypeText,
            ChunkIndex: i,
            ChunkTotal: len(chunks),
        }
        
        if err := memProvider.IngestDocument(ctx, doc); err != nil {
            return err
        }
    }
    
    return nil
}

func chunkText(text string, chunkSize, overlap int) []string {
    var chunks []string
    start := 0
    
    for start < len(text) {
        end := start + chunkSize
        if end > len(text) {
            end = len(text)
        }
        
        chunks = append(chunks, text[start:end])
        start += chunkSize - overlap
    }
    
    return chunks
}
```

### Ingesting PDF Documents

```go
import (
    "github.com/ledongthuc/pdf"
    "github.com/agenticgokit/agenticgokit/core"
)

func ingestPDF(memProvider core.Memory, filePath string) error {
    f, r, err := pdf.Open(filePath)
    if err != nil {
        return err
    }
    defer f.Close()
    
    ctx := context.Background()
    totalPages := r.NumPage()
    
    for pageNum := 1; pageNum <= totalPages; pageNum++ {
        p := r.Page(pageNum)
        if p.V.IsNull() {
            continue
        }
        
        text, _ := p.GetPlainText(nil)
        
        doc := core.Document{
            ID:      fmt.Sprintf("%s-page-%d", filePath, pageNum),
            Title:   filepath.Base(filePath),
            Content: text,
            Source:  filePath,
            Type:    core.DocumentTypePDF,
            Metadata: map[string]interface{}{
                "page":        pageNum,
                "total_pages": totalPages,
            },
        }
        
        if err := memProvider.IngestDocument(ctx, doc); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Ingesting Web Pages

```go
import (
    "github.com/gocolly/colly"
    "strings"
)

func ingestWebPage(memProvider core.Memory, url string) error {
    c := colly.NewCollector()
    
    var contentBuilder strings.Builder
    var title string
    
    c.OnHTML("title", func(e *colly.HTMLElement) {
        title = e.Text
    })
    
    c.OnHTML("article, main, .content", func(e *colly.HTMLElement) {
        contentBuilder.WriteString(e.Text)
    })
    
    if err := c.Visit(url); err != nil {
        return err
    }
    
    ctx := context.Background()
    doc := core.Document{
        ID:      url,
        Title:   title,
        Content: contentBuilder.String(),
        Source:  url,
        Type:    core.DocumentTypeWeb,
        Tags:    []string{"webpage"},
    }
    
    return memProvider.IngestDocument(ctx, doc)
}
```

### Batch Ingestion

For efficient ingestion of multiple documents:

```go
func batchIngest(memProvider core.Memory, filePaths []string) error {
    docs := make([]core.Document, 0, len(filePaths))
    
    for i, path := range filePaths {
        content, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        
        docs = append(docs, core.Document{
            ID:      fmt.Sprintf("doc-%d", i),
            Title:   filepath.Base(path),
            Content: string(content),
            Source:  path,
            Type:    core.DocumentTypeText,
        })
    }
    
    ctx := context.Background()
    return memProvider.IngestDocuments(ctx, docs)
}
```

---

## 🔄 Session Management

Manage separate conversation sessions:

```go
import (
    "github.com/agenticgokit/agenticgokit/v1beta"
    "github.com/agenticgokit/agenticgokit/core"
)

func sessionExample() {
    agent, _ := v1beta.NewBuilder("SessionAgent").
        WithPreset(v1beta.ChatAgent).
        WithMemory(
            v1beta.WithMemoryProvider("chromem"),
            v1beta.WithSessionScoped(),
        ).
        Build()
    
    // Get memory provider
    memProvider := core.GetMemoryProvider("memory")
    
    // Session 1
    ctx1 := memProvider.SetSession(context.Background(), "user-123")
    agent.Run(ctx1, "My favorite color is blue")
    
    // Session 2 (different user)
    ctx2 := memProvider.SetSession(context.Background(), "user-456")
    agent.Run(ctx2, "My favorite color is red")
    
    // Query session 1 context
    agent.Run(ctx1, "What's my favorite color?")
    // Output: "Your favorite color is blue."
    
    // Clear a session
    memProvider.ClearSession(ctx1)
}
```

---

## 🎨 Memory Patterns

### Pattern 1: Context-Aware Conversations

Enable context awareness for better conversations:

```go
agent, _ := v1beta.NewBuilder("ChatAgent").
    WithPreset(v1beta.ChatAgent).
    WithMemory(
        v1beta.WithMemoryProvider("chromem"),
        v1beta.WithContextAware(),
    ).
    Build()
```

### Pattern 2: Session-Scoped Memory

Isolate memory per user session:

```go
agent, _ := v1beta.NewBuilder("MultiUserAgent").
    WithPreset(v1beta.ChatAgent).
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithSessionScoped(),
    ).
    Build()
```

### Pattern 3: RAG with Both Personal and Knowledge

Combine conversation history with knowledge base:

```go
agent, _ := v1beta.NewBuilder("HybridAgent").
    WithPreset(v1beta.ResearchAgent).
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithContextAware(),
        v1beta.WithSessionScoped(),
        v1beta.WithRAG(4000, 0.3, 0.7),
    ).
    Build()
```

---

## 🔧 Programmatic Access

Access and manipulate memory directly through the `Agent` and `Workflow` interfaces:

### Accessing Memory Provider

```go
// Get memory provider from agent
memory := agent.Memory()

// Get memory provider from workflow
memory := workflow.Memory()

if memory != nil {
    // Perform direct memory operations
    err := memory.Store(ctx, "Direct memory storage")
}
```

### Inspecting Execution Context

The execution result includes detailed information about how memory was used:

```go
result, _ := agent.Run(ctx, "query")

if result.MemoryUsed {
    fmt.Printf("Memory queries performed: %d\n", result.MemoryQueries)
    
    // Access detailed RAG context
    if result.MemoryContext != nil {
        fmt.Printf("Total tokens: %d\n", result.MemoryContext.TotalTokens)
        
        // Inspect knowledge base matches
        for _, match := range result.MemoryContext.KnowledgeBase {
            fmt.Printf("Match: %s (Score: %.2f)\n", match.Content, match.Score)
        }
        
        // Check source attribution
        for _, source := range result.MemoryContext.SourceAttribution {
            fmt.Printf("Source: %s\n", source)
        }
    }
}
```

---

## 🎯 Best Practices

### 1. Choose the Right Provider

```go
// Development & Testing
v1beta.WithMemoryProvider("chromem") // Fast, no setup

// Production
v1beta.WithMemoryProvider("pgvector") // Persistent, scalable, vector search
```

### 2. Optimize Chunk Size

```go
// Too small - too many chunks, higher retrieval overhead
chunkSize := 100 // ❌

// Too large - less relevant retrieval
chunkSize := 10000 // ❌

// Optimal - balanced granularity
chunkSize := 1000
chunkOverlap := 200 // ✅
```

### 3. Configure RAG Weights

```go
// For personal assistant (prioritize conversation history)
v1beta.WithRAG(2000, 0.7, 0.3) // High personal weight

// For knowledge assistant (prioritize knowledge base)
v1beta.WithRAG(4000, 0.2, 0.8) // High knowledge weight

// Balanced
v1beta.WithRAG(3000, 0.5, 0.5) // Equal weights
```

### 4. Use Metadata for Better Retrieval

```go
doc := core.Document{
    Content: "Important information",
    Metadata: map[string]interface{}{
        "category":   "technical",
        "importance": "high",
        "timestamp":  time.Now(),
        "author":     "expert",
    },
    Tags: []string{"documentation", "api", "v1beta"},
}
```

### 5. Connection Pooling (PostgreSQL)

The pgvector provider automatically configures connection pooling:

```go
// Automatic configuration:
// - MaxConns: 25
// - MinConns: 5
// - MaxConnLifetime: 1 hour
// - HealthCheckPeriod: 1 minute
```

---

## 🐛 Troubleshooting

### Issue: Memory Not Persisting

**Cause**: Using in-memory provider

**Solution**: Use pgvector for persistence
```go
// ❌ Not persistent
v1beta.WithMemoryProvider("chromem")

// ✅ Persistent
v1beta.WithMemoryProvider("pgvector")
```

### Issue: Slow Vector Search

**Cause**: Missing pgvector indexes

**Solution**: Ensure indexes are created (automatic with pgvector provider)
```sql
-- Automatically created by pgvector provider:
CREATE INDEX idx_personal_memory_embedding 
    ON personal_memory USING ivfflat (embedding vector_cosine_ops);

CREATE INDEX idx_knowledge_embedding 
    ON knowledge_base USING ivfflat (embedding vector_cosine_ops);
```

### Issue: Connection Errors

**Cause**: Invalid connection string or missing environment variables

**Solution**: Set proper configuration
```bash
export AGENTFLOW_MEMORY_CONNECTION="postgresql://user:pass@localhost:5432/db"
export AGENTFLOW_MEMORY_PROVIDER="pgvector"
```

### Issue: Embeddings Not Generated

**Cause**: Missing embedding provider configuration

**Solution**: Configure embedding service
```toml
[memory.embedding]
provider = "openai"
model = "text-embedding-ada-002"
api_key = "${OPENAI_API_KEY}"  # Uses environment variable
```

---

## 📋 Complete Example

Here's a complete example combining all features:

```go
package main

import (
    "context"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
    "github.com/agenticgokit/agenticgokit/core"
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/pgvector"
)

func main() {
    // Create agent with full RAG capabilities
    agent, err := v1beta.NewBuilder("RAGAssistant").
        WithPreset(v1beta.ResearchAgent).
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{
                Provider: "openai",
                Model:    "gpt-4",
            },
        }).
        WithMemory(
            v1beta.WithMemoryProvider("pgvector"),
            v1beta.WithContextAware(),
            v1beta.WithSessionScoped(),
            v1beta.WithRAG(4000, 0.3, 0.7),
        ).
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    // Get memory provider
    memProvider := core.GetMemoryProvider("pgvector")
    
    // Ingest knowledge documents
    docs := []core.Document{
        {
            ID:      "doc-1",
            Title:   "Product Documentation",
            Content: "Our product features X, Y, and Z...",
            Source:  "docs/product.md",
            Type:    core.DocumentTypeMarkdown,
            Tags:    []string{"product", "documentation"},
        },
        {
            ID:      "doc-2",
            Title:   "API Reference",
            Content: "API endpoints include /api/v1/users...",
            Source:  "docs/api.md",
            Type:    core.DocumentTypeMarkdown,
            Tags:    []string{"api", "reference"},
        },
    }
    
    ctx := context.Background()
    if err := memProvider.IngestDocuments(ctx, docs); err != nil {
        log.Fatal(err)
    }
    
    // Create session for user
    sessionCtx := memProvider.SetSession(ctx, "user-123")
    
    // Run queries - agent will use RAG to retrieve relevant context
    result, err := agent.Run(sessionCtx, "What are the product features?")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %s", result.Content)
}
```

---

## 📚 Next Steps

- **[Configuration](./configuration.md)** - Advanced memory configuration
- **[Custom Handlers](./custom-handlers.md)** - Custom memory integration
- **[Tool Integration](./tool-integration.md)** - Memory-aware tools
- **[Performance](./performance.md)** - Memory optimization tips

---

**Ready for custom behavior?** Continue to [Custom Handlers](./custom-handlers.md) →
