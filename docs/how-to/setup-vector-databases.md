# How to Set Up Vector Databases

**Configure persistent vector storage for RAG and memory systems**

This guide walks you through setting up vector databases with AgenticGoKit for persistent memory and RAG (Retrieval-Augmented Generation) capabilities. You'll learn to configure PostgreSQL with pgvector and Weaviate for different use cases.

## Prerequisites

- Docker installed on your system
- Basic understanding of AgenticGoKit memory systems
- Command line familiarity

## What You'll Build

A vector database setup that supports:
- Document storage and retrieval
- Semantic search capabilities
- Persistent agent memory
- RAG-powered question answering

## Quick Start (10 minutes)

### 1. Create Memory-Enabled Project

```bash
# Create project with pgvector memory
agentcli create vector-db-demo --memory-enabled --memory-provider pgvector \
  --rag-enabled --embedding-provider ollama
cd vector-db-demo
```

### 2. Start Database Services

The project includes a `docker-compose.yml` file:

```bash
# Start PostgreSQL with pgvector
docker compose up -d

# Verify the database is running
docker compose ps
```

### 3. Initialize Database

```bash
# Run the setup script (generated with your project)
./setup.sh  # On Linux/Mac
# or
setup.bat   # On Windows
```

### 4. Test the Setup

```bash
# Install dependencies
go mod tidy

# Set up Ollama (if using local embeddings)
ollama pull nomic-embed-text:latest

# Test the system
go run . -m "Tell me about vector databases"
```

## Database Options Comparison

| Feature | PostgreSQL + pgvector | Weaviate | In-Memory |
|---------|----------------------|----------|-----------|
| **Persistence** | ✅ Full | ✅ Full | ❌ Temporary |
| **Scalability** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| **Setup Complexity** | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Query Performance** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Resource Usage** | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Best For** | Production apps | Large scale | Development |

## PostgreSQL + pgvector Setup

### Detailed Configuration

The generated `docker-compose.yml` includes:

```yaml
version: '3.8'
services:
  postgres:
    image: pgvector/pgvector:pg15
    environment:
      POSTGRES_DB: agentflow
      POSTGRES_USER: agentflow
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U agentflow"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:
```

### Database Initialization

The `init-db.sql` file sets up the required extensions and tables:

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create embeddings table
CREATE TABLE IF NOT EXISTS embeddings (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    embedding vector(384),  -- Adjust dimensions based on your model
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for vector similarity search
CREATE INDEX IF NOT EXISTS embeddings_embedding_idx 
ON embeddings USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);

-- Create memory table for agent conversations
CREATE TABLE IF NOT EXISTS agent_memory (
    id SERIAL PRIMARY KEY,
    agent_name VARCHAR(255) NOT NULL,
    session_id VARCHAR(255),
    content TEXT NOT NULL,
    embedding vector(384),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS agent_memory_embedding_idx 
ON agent_memory USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);
```

### Configuration in AgenticGoKit

The generated `agentflow.toml` includes:

```toml
[agent_memory]
provider = "pgvector"
connection = "postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"
max_results = 5
dimensions = 384  # Matches your embedding model
auto_embed = true
enable_knowledge_base = true

[agent_memory.embedding]
provider = "ollama"
model = "nomic-embed-text:latest"
base_url = "http://localhost:11434"
cache_embeddings = true
max_batch_size = 100
timeout_seconds = 30
```

### Testing PostgreSQL Setup

```bash
# Test database connection
psql -h localhost -U agentflow -d agentflow -c "SELECT version();"

# Check pgvector extension
psql -h localhost -U agentflow -d agentflow -c "SELECT * FROM pg_extension WHERE extname = 'vector';"

# Test vector operations
psql -h localhost -U agentflow -d agentflow -c "SELECT '[1,2,3]'::vector <-> '[4,5,6]'::vector;"
```

## Weaviate Setup

### Create Weaviate Project

```bash
# Create project with Weaviate
agentcli create weaviate-demo --memory-enabled --memory-provider weaviate \
  --rag-enabled --embedding-provider openai
cd weaviate-demo
```

### Weaviate Docker Compose

```yaml
version: '3.8'
services:
  weaviate:
    image: semitechnologies/weaviate:latest
    ports:
      - "8080:8080"
    environment:
      QUERY_DEFAULTS_LIMIT: 25
      AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'true'
      PERSISTENCE_DATA_PATH: '/var/lib/weaviate'
      DEFAULT_VECTORIZER_MODULE: 'none'
      ENABLE_MODULES: 'text2vec-openai,text2vec-cohere,text2vec-huggingface'
      CLUSTER_HOSTNAME: 'node1'
    volumes:
      - weaviate_data:/var/lib/weaviate

volumes:
  weaviate_data:
```

### Weaviate Configuration

```toml
[agent_memory]
provider = "weaviate"
connection = "http://localhost:8080"
max_results = 5
dimensions = 1536  # OpenAI ada-002 dimensions
auto_embed = true

[agent_memory.embedding]
provider = "openai"
model = "text-embedding-ada-002"
cache_embeddings = true
```

### Testing Weaviate Setup

```bash
# Start Weaviate
docker compose up -d

# Check Weaviate health
curl http://localhost:8080/v1/meta

# Test with your application
export OPENAI_API_KEY=your-key-here
go run . -m "What can you remember?"
```

## Advanced Configuration

### Optimizing PostgreSQL Performance

```sql
-- Tune pgvector settings for better performance
ALTER SYSTEM SET shared_preload_libraries = 'vector';
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';

-- Restart PostgreSQL after changes
```

### Custom Embedding Dimensions

If you're using a different embedding model, update the dimensions:

```sql
-- For different embedding models
-- OpenAI ada-002: 1536 dimensions
-- Sentence Transformers: 384 or 768 dimensions
-- Custom models: varies

ALTER TABLE embeddings ALTER COLUMN embedding TYPE vector(1536);
ALTER TABLE agent_memory ALTER COLUMN embedding TYPE vector(1536);

-- Recreate indexes with new dimensions
DROP INDEX IF EXISTS embeddings_embedding_idx;
CREATE INDEX embeddings_embedding_idx 
ON embeddings USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);
```

### Connection Pooling

For production use, configure connection pooling:

```toml
[agent_memory]
provider = "pgvector"
connection = "postgres://agentflow:password@localhost:5432/agentflow?sslmode=disable"
max_connections = 20
idle_connections = 5
connection_lifetime = "1h"
```

## Production Considerations

### Security

```bash
# Use environment variables for credentials
export DB_PASSWORD=$(openssl rand -base64 32)
export DB_CONNECTION="postgres://agentflow:${DB_PASSWORD}@localhost:5432/agentflow?sslmode=require"
```

```toml
[agent_memory]
provider = "pgvector"
connection = "${DB_CONNECTION}"  # Uses environment variable
```

### Backup Strategy

```bash
# PostgreSQL backup
pg_dump -h localhost -U agentflow agentflow > backup.sql

# Restore
psql -h localhost -U agentflow agentflow < backup.sql

# Automated backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -h localhost -U agentflow agentflow | gzip > "backup_${DATE}.sql.gz"
```

### Monitoring

```sql
-- Monitor vector index usage
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes 
WHERE indexname LIKE '%embedding%';

-- Check table sizes
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE tablename IN ('embeddings', 'agent_memory');
```

## Troubleshooting

### Common Issues

**Database connection failed:**
```bash
# Check if database is running
docker compose ps

# Check logs
docker compose logs postgres

# Test connection manually
psql -h localhost -U agentflow -d agentflow -c "SELECT 1;"
```

**pgvector extension not found:**
```bash
# Ensure you're using the pgvector image
docker compose down
docker compose pull
docker compose up -d
```

**Slow vector queries:**
```sql
-- Check if indexes exist
\d+ embeddings

-- Recreate index with more lists for larger datasets
DROP INDEX embeddings_embedding_idx;
CREATE INDEX embeddings_embedding_idx 
ON embeddings USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 1000);  -- Increase for larger datasets
```

**Memory usage too high:**
```toml
# Reduce batch sizes and connection limits
[agent_memory.embedding]
max_batch_size = 50  # Reduce from 100

[agent_memory]
max_connections = 10  # Reduce connection pool
```

### Performance Tuning

**For large datasets (>1M vectors):**
```sql
-- Use HNSW index for better performance (PostgreSQL 14+)
CREATE INDEX embeddings_embedding_hnsw_idx 
ON embeddings USING hnsw (embedding vector_cosine_ops);
```

**For high-throughput applications:**
```toml
[agent_memory]
connection_timeout = "5s"
query_timeout = "10s"
max_connections = 50
idle_connections = 10
```

## Migration Between Providers

### From In-Memory to PostgreSQL

```bash
# Create new project with PostgreSQL
agentcli create migrated-project --memory-enabled --memory-provider pgvector

# Export data from old system (if applicable)
# Import into new PostgreSQL setup
```

### From PostgreSQL to Weaviate

```bash
# Export vectors from PostgreSQL
psql -h localhost -U agentflow -d agentflow -c "COPY (SELECT content, embedding, metadata FROM embeddings) TO '/tmp/vectors.csv' WITH CSV HEADER;"

# Import into Weaviate (requires custom script)
# See Weaviate documentation for bulk import
```

## Next Steps

With your vector database set up:

1. **Implement Document Ingestion**: [Document Ingestion Guide](implement-document-ingestion.md)
2. **Optimize RAG Performance**: [RAG Optimization Guide](optimize-rag-performance.md)
3. **Build Knowledge Systems**: [Research Assistant Guide](build-research-assistant.md)
4. **Monitor Performance**: [Performance Monitoring Guide](monitor-performance.md)

## Related Resources

- [Memory Systems Tutorial](../tutorials/15-minute-series/memory-and-rag.md)
- [RAG Implementation Guide](../tutorials/memory-systems/rag-implementation.md)
- [Vector Database Comparison](../tutorials/memory-systems/vector-databases.md)

---

*Vector database setup is a foundational step for building knowledge-aware agents. Choose the provider that best fits your scale and requirements.*