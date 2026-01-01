# Memory and RAG

Add conversational memory and Retrieval-Augmented Generation (RAG) to v1beta agents with the built-in chromem default or a production pgvector backend.

---

## What you get

- Memory on by default via the embedded chromem provider (vector search included)
- Functional options to swap providers, enable RAG, and scope memory per session
- Direct access to memory inside handlers through `caps.Memory` or via `agent.Memory()`
- Config-file support through `LoadConfigFromTOML` when you want to drive everything from TOML

---

## Defaults and options

- Default: chromem is enabled when no memory config is provided. Imports register providers; no extra code needed.
- Disable memory: set `memory.enabled = false` in TOML or skip `WithMemory` and provide a `Config` with memory disabled.
- MemoryOption helpers:
  - `WithMemoryProvider(provider string)` – "chromem" (embedded) or "pgvector" (PostgreSQL + pgvector).
  - `WithRAG(maxTokens int, personalWeight, knowledgeWeight float32)` – sets weights and a 10-message history window.
  - `WithSessionScoped()` – isolates memory per session ID.
  - `WithContextAware()` – hints providers to include richer chat context.

Do not chain `WithPreset(...)` and `WithConfig(...)` together; pick one source of truth.

---

## Quick start (chromem)

```go
package main

import (
    "context"
    "log"

    "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem" // registers chromem
)

func main() {
    agent, err := v1beta.NewBuilder("MemoryAgent").
        WithPreset(v1beta.ChatAgent).
        WithLLM("openai", "gpt-4o-mini").
        WithMemory( // optional; chromem is already the default
            v1beta.WithMemoryProvider("chromem"),
        ).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    _, _ = agent.Run(context.Background(), "My name is Alice")
    _, _ = agent.Run(context.Background(), "What is my name?") // remembers Alice
}
```

---

## Providers

### Chromem (embedded default)

- Zero external dependencies, vector search included.
- Great for local dev and single-instance deployments.

```go
_ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem"
agent, _ := v1beta.NewBuilder("Dev").
    WithPreset(v1beta.ChatAgent).
    WithLLM("openai", "gpt-4o-mini").
    Build()
```

### Pgvector (PostgreSQL)

- Persistent, multi-instance ready, vector search with pgvector extension.

```go
_ "github.com/agenticgokit/agenticgokit/plugins/memory/pgvector"

agent, _ := v1beta.NewBuilder("Prod").
    WithPreset(v1beta.ResearchAgent).
    WithLLM("openai", "gpt-4o-mini").
    WithMemory(v1beta.WithMemoryProvider("pgvector")).
    Build()
```

PostgreSQL prep:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

Minimal TOML for pgvector:

```toml
[memory]
enabled = true
provider = "pgvector"
connection = "postgresql://user:pass@localhost:5432/agentdb"

[memory.rag]
max_tokens = 3000
personal_weight = 0.4
knowledge_weight = 0.6
history_limit = 10
```

Load it when you want config-driven setup:

```go
import (
    "context"

    "github.com/agenticgokit/agenticgokit/v1beta"
)

cfg, _ := v1beta.LoadConfigFromTOML("agent.toml")

agent, _ := v1beta.NewBuilder(cfg.Name).
    WithConfig(cfg). // do not combine WithPreset here
    WithHandler(func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
        return caps.LLM("You are helpful", input)
    }).
    Build()
```

---

## RAG basics

```go
import (
    "context"

    "github.com/agenticgokit/agenticgokit/core"
    "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem"
)

agent, _ := v1beta.NewBuilder("RAG").
    WithPreset(v1beta.ResearchAgent).
    WithLLM("openai", "gpt-4o-mini").
    WithMemory(
        v1beta.WithRAG(3000, 0.5, 0.5),
    ).
    Build()

mem := agent.Memory()
_ = mem.IngestDocument(context.Background(), core.Document{
    ID:      "kb-1",
    Title:   "World Capitals",
    Content: "The capital of France is Paris. The capital of Germany is Berlin.",
    Type:    core.DocumentTypeText,
})

_, _ = agent.Run(context.Background(), "What is the capital of France?")
```

Weights: raise `personalWeight` for personal assistants, raise `knowledgeWeight` for doc-grounded agents. `history_limit` controls how many past turns are considered when assembling RAG context.

---

## Ingesting knowledge

Use the memory interface; it already handles embeddings and storage.

```go
import (
    "context"

    "github.com/agenticgokit/agenticgokit/core"
)

mem := agent.Memory()
ctx := context.Background()

docs := []core.Document{
    {ID: "doc-1", Title: "Product Guide", Content: "Feature A, B, C", Type: core.DocumentTypeMarkdown},
    {ID: "doc-2", Title: "FAQ", Content: "Common questions", Type: core.DocumentTypeText},
}

_ = mem.IngestDocuments(ctx, docs)
```

Add tags or metadata to improve retrieval:

```go
core.Document{Content: "Scaling guide", Tags: []string{"architecture", "production"}}
```

---

## Session scoping

`WithSessionScoped` keeps memory isolated per session ID.

```go
import "context"

agent, _ := v1beta.NewBuilder("Sessions").
    WithPreset(v1beta.ChatAgent).
    WithLLM("openai", "gpt-4o-mini").
    WithMemory(v1beta.WithSessionScoped()).
    Build()

mem := agent.Memory()
ctxA := mem.SetSession(context.Background(), "user-a")
ctxB := mem.SetSession(context.Background(), "user-b")

_, _ = agent.Run(ctxA, "Remember I like blue")
_, _ = agent.Run(ctxB, "Remember I like red")
_, _ = agent.Run(ctxA, "What color do I like?") // blue
```

---

## Using memory inside handlers

`caps.Memory` is available in the handler; you can augment prompts manually or run hybrid searches.

```go
import (
    "context"
    "fmt"

    "github.com/agenticgokit/agenticgokit/v1beta"
)

handler := func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    if caps.Memory != nil {
        if res, err := caps.Memory.SearchAll(ctx, input); err == nil && res != nil {
            return caps.LLM("Answer using this context", fmt.Sprint(res.PersonalMemory, res.Knowledge))
        }
    }
    return caps.LLM("You are helpful", input)
}
```

---

## Workflow Shared Memory

Enable multiple agents in a workflow to share the same memory store. Agents automatically query workflow memory for relevant context.

### Automatic Context Querying

When a workflow has shared memory, agents automatically receive relevant context:

```go
import (
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

// Create shared memory
sharedMemory, _ := v1beta.NewMemory(&v1beta.MemoryConfig{
    Provider: "chromem",
})

// Create workflow
workflow, _ := v1beta.NewSequentialWorkflow(&v1beta.WorkflowConfig{
    Mode: v1beta.Sequential,
})

// Attach shared memory - agents will automatically query it
workflow.SetMemory(sharedMemory)

workflow.AddStep(v1beta.WorkflowStep{Name: "learn", Agent: learner})
workflow.AddStep(v1beta.WorkflowStep{Name: "answer", Agent: answerer})

// Agent 1 stores facts → Agent 2 automatically queries them
result, _ := workflow.Run(ctx, "Company data...")
```

### How It Works

1. Workflow stores step inputs/outputs in shared memory via `SetMemory()`
2. Memory reference is passed to agents through context
3. Agents call `GetWorkflowMemory(ctx)` internally
4. Relevant content is queried and added to prompts automatically

### Direct Access

Access workflow memory directly from custom handlers:

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

handler := func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Check for workflow shared memory
    if workflowMem := v1beta.GetWorkflowMemory(ctx); workflowMem != nil {
        results, _ := workflowMem.Query(ctx, input, 
            v1beta.WithLimit(5),
            v1beta.WithScoreThreshold(0.3))
        
        // Use results to enrich prompt
        for _, result := range results {
            // Process shared context
        }
    }
    
    // Continue with LLM call
    return caps.LLM("system", input)
}
```

### Use Cases

**Multi-agent collaboration:**
- Research agent gathers data → Analysis agents query shared knowledge
- Data extraction → Multiple processors access extracted entities
- Progressive knowledge building across workflow steps

**See [Workflows Guide](./workflows.md#shared-memory) for complete examples.**

---

## Troubleshooting

- Memory is nil: ensure the provider import is present (`plugins/memory/chromem` or `plugins/memory/pgvector`) and memory is not disabled in config.
- Pgvector connection issues: verify `memory.connection` is a valid PostgreSQL URI and the `vector` extension is installed.
- RAG not firing: confirm `WithRAG` is set or `memory.rag` exists in TOML, and that documents have been ingested.
- Avoid mixing config sources: use either `WithPreset` plus options or `WithConfig` loaded from TOML, not both.

---

## Next steps

- [core-concepts](core-concepts.md) for the builder overview
- [tool-integration](tool-integration.md) for memory-aware tool flows
- [configuration](configuration.md) for the complete TOML schema
