# Memory API (vNext)

**Unified memory interface with RAG context, sessions, and helper utilities**

The vNext memory layer standardises how agents store and retrieve conversational or knowledge snippets. It exposes the same `Memory` interface across in-process memory, vector stores, and custom providers registered at runtime.

## 🔑 Interface Overview

```go
type Memory interface {
    Store(ctx context.Context, content string, opts ...StoreOption) error
    Query(ctx context.Context, query string, opts ...QueryOption) ([]MemoryResult, error)

    NewSession() string
    SetSession(ctx context.Context, sessionID string) context.Context

    IngestDocument(ctx context.Context, doc Document) error
    BuildContext(ctx context.Context, query string, opts ...ContextOption) (*RAGContext, error)
}
```

### Core Types

- `MemoryResult` → text, score, source, metadata, timestamp
- `Document` → ID, title, content, source, metadata (used for ingestion)
- `RAGContext` → personal memory, knowledge base, chat history, total tokens, source attribution

## 🧱 Configuring Memory

Attach `MemoryConfig` via the builder or configuration files:

```go
builder.WithMemory(
    vnext.WithMemoryProvider("memory"), // swap for pgvector, weaviate, etc.
    vnext.WithSessionScoped(),
    vnext.WithContextAware(),
    vnext.WithRAG(4096, 0.3, 0.7),
)
```

When using TOML:

```toml
[memory]
provider = "pgvector"
connection = "postgres://..."

[memory.rag]
max_tokens = 4096
personal_weight = 0.4
knowledge_weight = 0.6
history_limit = 12
```

`MemoryOptions` on `RunOptions` allow per-call overrides (session ID, enable/disable memory, custom RAG settings).

## 🗃️ Storing and Querying

```go
mem, _ := vnext.NewMemory(&vnext.MemoryConfig{Provider: "memory"})

_ = mem.Store(ctx, "User name is Priya",
    vnext.WithContentType("fact"),
    vnext.WithSource("profile_form"),
    vnext.WithMetadata(map[string]interface{}{"confidence": 0.9}),
)

results, _ := mem.Query(ctx, "What is the user's name?",
    vnext.WithLimit(3),
    vnext.WithScoreThreshold(0.6),
    vnext.WithIncludeMetadata(true),
)
```

`StoreOption`, `QueryOption`, and `ContextOption` mutate the request structs under the hood, letting providers honour content type, source attribution, score thresholds, and token budgets.

## 🧠 RAG Context

```go
ctxBuilder := vnext.BuildSimpleContext(ctx, mem, "Explain latest incidents", 2000)

ragCtx, err := mem.BuildContext(ctx, "Explain latest incidents",
    vnext.WithMaxTokens(3000),
    vnext.WithPersonalWeight(0.2),
    vnext.WithKnowledgeWeight(0.8),
)

formatted := vnext.FormatRAGContext(ragCtx, "")
```

Use helper functions from `memory.go` / `utils.go`:

- `BuildSimpleContext` → quick personal/knowledge blend
- `IngestTextDocument` → wraps `Document` creation
- `FormatRAGContext` → render retrieved memories into a prompt snippet
- `EstimateTokenCount` → rough token approximation for gating

## 🔄 Sessions

```go
sessionID := mem.NewSession()
session := vnext.NewSessionContext(mem, sessionID)

_ = session.Store(ctx, "User prefers concise answers")
memories, _ := session.Query(ctx, "preferences", vnext.WithLimit(5))
```

`Memory.SetSession` embeds a session ID into the context so successive calls stay scoped. Agents automatically use the configured session when `RunOptions.SessionID` is set.

## 🧬 Provider Registry

Register custom backends by providing a factory:

```go
vnext.RegisterMemoryProvider("redis", func(cfg *vnext.MemoryConfig) (vnext.Memory, error) {
    return newRedisMemory(cfg)
})
```

Call `vnext.NewMemory(&vnext.MemoryConfig{Provider: "redis"})` after registration to get instances from plugins or internal packages.

## 📊 Monitoring

Memory implementations can expose metrics by fulfilling the optional `StatsProvider` interface:

```go
stats, err := vnext.GetMemoryStats(ctx, mem)
if err == nil {
    log.Printf("stored documents: %d", stats.TotalDocuments)
}
```

## 🌉 Prompt Helpers

Combine memory with the LLM prompt helpers from `utils.go`:

```go
enriched := vnext.EnrichWithMemory(ctx, legacyCoreMemory, input, cfg.Memory)
prompt := vnext.BuildEnrichedPrompt(ctx, cfg.SystemPrompt, input, legacyCoreMemory, cfg.Memory)
```

These helpers:

- Query memory for relevant facts
- Apply RAG weighting if configured
- Optionally attach chat history (using `RAGConfig.HistoryLimit`)

## 🔗 Related Docs

- [agent.md](agent.md) explains how `RunOptions` activate memory per call
- [tools.md](tools.md) to fetch context for tool plans
- [workflow.md](workflow.md) covers shared memory inside multi-step orchestrations
