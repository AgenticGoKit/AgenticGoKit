# Memory Basics

Goal: Enable agents to remember information across conversations and searches.

## Prerequisites
Complete [orchestration-basics.md](orchestration-basics.md) to understand multi-agent systems.

## What is Memory?
Memory allows agents to:
- Remember previous conversations
- Store and search knowledge documents
- Perform RAG (Retrieval-Augmented Generation)
- Share context between agents

## 1) Start with in-memory provider
```pwsh
agentcli create memory-demo --memory memory --agents 2
Set-Location memory-demo
```

This creates a project with memory enabled. Check the generated `agentflow.toml`:
```toml
[agent_memory]
provider = "memory"          # In-memory storage (non-persistent)
auto_embed = true           # Automatically create embeddings
max_results = 10           # Maximum search results
enable_knowledge_base = true
```

Validate and run:
```pwsh
agentcli validate
go run . -m "Remember that I like pizza"
```

Expected: Agent stores the information and can reference it in responses.

## 2) Test memory persistence
Run another query:
```pwsh
go run . -m "What food do I like?"
```

Expected: Agent recalls your pizza preference from the previous interaction.

## 3) Enable RAG (Retrieval-Augmented Generation)
Edit `agentflow.toml` to add RAG:
```toml
[agent_memory]
provider = "memory"
enable_rag = true
chunk_size = 1000
chunk_overlap = 100
rag_max_context_tokens = 4000
```

This allows agents to break down large documents and find relevant sections.

## 4) Going persistent (PostgreSQL + pgvector)
For production use, create a persistent memory system:
```pwsh
Set-Location ..
agentcli create persistent-demo --memory pgvector --agents 2
Set-Location persistent-demo
```

Check the generated configuration:
```toml
[agent_memory]
provider = "pgvector"
connection = "postgres://user:password@localhost:15432/agentflow?sslmode=disable"
dimensions = 1536
enable_rag = true
```

⚠️ **Note**: You need a running PostgreSQL database with pgvector extension. See the [Memory Provider Setup Guide](../../guides/MemoryProviderSetup.md) for database setup instructions.

## 5) Alternative: Weaviate vector database
```pwsh
agentcli create weaviate-demo --memory weaviate --agents 2
```

Configuration:
```toml
[agent_memory]
provider = "weaviate"
connection = "http://localhost:8080"
```

## Memory Provider Comparison
| Provider | Persistence | Setup | Best For |
|----------|-------------|--------|----------|
| `memory` | No | None | Development, testing |
| `pgvector` | Yes | PostgreSQL + extension | Production, SQL ecosystem |
| `weaviate` | Yes | Docker container | Production, advanced vector features |

## Next Steps
- Add external tools: [tools-basics.md](tools-basics.md)
- Learn deployment: [deploy-basics.md](deploy-basics.md)

## Verification checklist
- [ ] In-memory project created and validated
- [ ] Memory persistence demonstrated across runs
- [ ] RAG configuration understood
- [ ] Persistent provider option explored
