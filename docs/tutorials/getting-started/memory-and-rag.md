# Memory and RAG Tutorial (15 minutes)

## Overview

Learn how to add persistent memory and knowledge systems to your agents using RAG (Retrieval-Augmented Generation). You'll set up vector databases, implement document ingestion, and create knowledge-aware agents.

## Prerequisites

- Complete the [Multi-Agent Collaboration](multi-agent-collaboration.md) tutorial
- Docker installed (for database setup)
- Basic understanding of vector databases

## Learning Objectives

By the end of this tutorial, you'll understand:
- How to set up vector databases for agent memory
- Document ingestion and chunking strategies
- RAG implementation for knowledge-aware responses
- Hybrid search combining semantic and keyword matching

## What You'll Build

A knowledge-aware agent system that can:
1. **Ingest documents** into a vector database
2. **Search knowledge** using semantic similarity
3. **Generate responses** enhanced with retrieved context
4. **Remember conversations** across sessions

---

## Part 1: Basic Memory Setup (5 minutes)

Start with in-memory storage to understand the concepts.

### Create a Memory-Enabled Project

```bash
# Create project with basic memory
agentcli create knowledge-agent --memory memory --agents 2
cd knowledge-agent
```

### Understanding Memory Configuration

The generated `agentflow.toml` includes memory settings:

```toml
[agent_memory]
provider = "memory"           # In-memory storage (temporary)
auto_embed = true            # Automatically create embeddings
max_results = 10             # Maximum search results
dimensions = 1536            # Embedding dimensions

[agent_memory.embedding]
provider = "ollama"          # Local embeddings (recommended)
model = "nomic-embed-text:latest"
```

### Test Basic Memory

```bash
# Make sure Ollama is running with the embedding model
ollama pull nomic-embed-text:latest

# Set your API key and run
export OPENAI_API_KEY=your-api-key-here
go run main.go
```

The agents now have basic memory capabilities, but data is lost when the program stops.

---

## Part 2: Persistent Memory with PostgreSQL (5 minutes)

Set up persistent memory using PostgreSQL with pgvector extension.

### Create a Persistent Memory Project

```bash
# Create project with PostgreSQL memory
agentcli create persistent-agent --memory pgvector --rag default --agents 2
cd persistent-agent
```

### Start the Database

The project includes a `docker-compose.yml` file:

```bash
# Start PostgreSQL with pgvector extension
docker-compose up -d

# Wait for database to be ready (about 30 seconds)
docker-compose logs -f postgres
```

### Understanding Persistent Configuration

```toml
[agent_memory]
provider = "pgvector"        # PostgreSQL with vector extension
connection = "postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"
enable_knowledge_base = true # Enable document storage
enable_rag = true           # Enable RAG functionality

[agent_memory.documents]
supported_types = ["pdf", "txt", "md", "web"]
auto_chunk = true           # Automatically chunk documents
chunk_size = 1000          # Tokens per chunk
chunk_overlap = 200        # Overlap between chunks

[agent_memory.search]
hybrid_search = true       # Combine semantic + keyword search
keyword_weight = 0.3       # 30% keyword, 70% semantic
semantic_weight = 0.7
```

### Test Persistent Memory

```bash
export OPENAI_API_KEY=your-api-key-here
go run main.go
```

Now your agents have persistent memory that survives restarts!

---

## Part 3: RAG Implementation (5 minutes)

Implement full RAG (Retrieval-Augmented Generation) with document ingestion.

### Create a RAG-Enabled System

```bash
# Create comprehensive RAG system
agentcli create rag-system --template rag-system
cd rag-system
```

### Start the Enhanced Database

```bash
docker-compose up -d
```

### Understanding RAG Configuration

```toml
[agent_memory]
provider = "pgvector"
enable_rag = true
rag_max_context_tokens = 4000    # Max context for RAG
rag_personal_weight = 0.3        # Weight for personal memory
rag_knowledge_weight = 0.7       # Weight for knowledge base

[agent_memory.documents]
enable_metadata_extraction = true # Extract document metadata
enable_url_scraping = true       # Support web URLs
max_file_size = "10MB"          # Maximum document size

[agent_memory.search]
hybrid_search = true            # Semantic + keyword search
top_k = 5                      # Top results to retrieve
score_threshold = 0.7          # Minimum similarity score
```

### Add Documents to the Knowledge Base

Create a sample document:

```bash
# Create a sample knowledge document
cat > knowledge.md << 'EOF'
# AgenticGoKit Knowledge Base

## Multi-Agent Systems
AgenticGoKit supports multiple orchestration patterns:
- Collaborative: Agents work in parallel
- Sequential: Agents work in pipeline
- Mixed: Combination of both patterns

## Memory Systems
AgenticGoKit provides several memory providers:
- In-memory: Fast but temporary
- PostgreSQL: Persistent with pgvector
- Weaviate: Dedicated vector database

## Tool Integration
Agents can use external tools through MCP:
- Web search capabilities
- File operations
- Custom API integrations
EOF
```

### Test RAG System

```bash
export OPENAI_API_KEY=your-api-key-here
go run main.go
```

The system will:
1. **Ingest** the knowledge document
2. **Chunk** it into searchable pieces
3. **Embed** chunks using vector embeddings
4. **Retrieve** relevant context for queries
5. **Generate** enhanced responses using RAG

### Query the Knowledge Base

The agents can now answer questions using the ingested knowledge:

- "What orchestration patterns does AgenticGoKit support?"
- "How do I set up persistent memory?"
- "What tools can agents use?"

---

## Memory Providers Comparison

| Provider | Persistence | Performance | Use Case |
|----------|-------------|-------------|----------|
| **memory** | ❌ Temporary | ⚡ Fastest | Development, testing |
| **pgvector** | ✅ Persistent | 🚀 Fast | Production, SQL integration |
| **weaviate** | ✅ Persistent | 🚀 Fast | Advanced vector operations |

## RAG Configuration Options

### Document Processing

```bash
# Customize document processing
agentcli create doc-system --memory pgvector --rag 512
```

### Search Configuration

```bash
# Fine-tune search behavior
agentcli create search-system --memory pgvector --rag default
```

### Embedding Models

```bash
# Use OpenAI embeddings (requires API key)
agentcli create openai-system --memory pgvector --embedding openai

# Use local Ollama embeddings (recommended)
agentcli create local-system --memory pgvector --embedding ollama:nomic-embed-text
```

## Advanced Memory Features

### Session Memory

```bash
# Enable session-based memory isolation
agentcli create session-system --template chat-system
```

Session memory keeps conversations separate for different users or contexts.

### Hybrid Search

```bash
# Configure hybrid search weights
agentcli create hybrid-system --memory pgvector --rag default
```

Hybrid search combines:
- **Semantic search**: Understanding meaning and context
- **Keyword search**: Exact term matching

## Troubleshooting

### Common Issues

**Database connection failed:**
```bash
# Check if PostgreSQL is running
docker-compose ps

# Check logs
docker-compose logs postgres

# Restart if needed
docker-compose restart postgres
```

**Embedding model not found:**
```bash
# For Ollama embeddings
ollama pull nomic-embed-text:latest
ollama list  # Verify model is installed

# Check Ollama is running
curl http://localhost:11434/api/tags
```

**RAG not working:**
```bash
# Verify documents are ingested
# Check agentflow.toml configuration
# Ensure embedding provider is working
```

### Performance Issues

**Slow search:**
- Reduce `rag_top_k` value
- Increase `score_threshold`
- Use smaller embedding models

**High memory usage:**
- Reduce `chunk_size`
- Limit `max_results`
- Use pgvector instead of in-memory

## Memory System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Documents     │───▶│   Chunking      │───▶│   Embeddings    │
│   (PDF, MD,     │    │   (1000 tokens) │    │   (Vector DB)   │
│    TXT, Web)    │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Agent         │◀───│   RAG Context   │◀───│   Similarity    │
│   Response      │    │   Injection     │    │   Search        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Next Steps

Now that your agents have memory and knowledge capabilities:

1. **Add Tools**: Learn [Tool Integration](tool-integration.md) to connect external services
2. **Go Production**: Check [Production Deployment](production-deployment.md) for scaling
3. **Advanced Memory**: Explore [Memory System Tutorials](../memory-systems/) for deep dives

## Key Takeaways

- **Memory Providers**: Choose based on persistence and performance needs
- **RAG**: Combines retrieval and generation for knowledge-aware responses
- **Document Processing**: Automatic chunking and embedding for searchability
- **Hybrid Search**: Best results combining semantic and keyword matching
- **Session Memory**: Isolate conversations for multi-user scenarios

## Further Reading

- [Memory Systems Deep Dive](../memory-systems/README.md) - Advanced memory concepts
- [Vector Databases Guide](../memory-systems/vector-databases.md) - Database comparison
- [RAG Implementation](../memory-systems/rag-implementation.md) - Advanced RAG patterns