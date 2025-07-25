# RAG Knowledge Base Example

**Build a knowledge-aware agent system with document ingestion and semantic search**

This example demonstrates how to build a production-ready RAG (Retrieval-Augmented Generation) system using AgenticGoKit. The system can ingest documents, store them in a vector database, and answer questions using retrieved context.

## What This Example Shows

- Document ingestion and chunking
- Vector embeddings and storage
- Semantic search and retrieval
- RAG-powered question answering
- Multi-agent collaboration for knowledge processing

## Architecture

```
Documents → Ingestion Agent → Vector Storage → Query Agent → Response
     ↓           ↓                ↓             ↓           ↓
   PDF/Text   Chunking        PostgreSQL    Retrieval   Enhanced
   Files      Embedding       + pgvector    + Context   Answers
```

## Prerequisites

- Docker and Docker Compose
- Go 1.21 or later
- OpenAI API key (or Ollama for local embeddings)

## Quick Start

### 1. Clone and Setup

```bash
# Navigate to this example
cd examples/04-rag-knowledge-base

# Install dependencies
go mod tidy

# Start PostgreSQL with pgvector
docker compose up -d

# Wait for database to be ready
sleep 10

# Initialize database
./setup.sh
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your API keys
# For OpenAI embeddings:
export OPENAI_API_KEY="your-openai-api-key"

# Or for local Ollama embeddings:
export EMBEDDING_PROVIDER="ollama"
ollama pull nomic-embed-text:latest
```

### 3. Ingest Documents

```bash
# Add some sample documents
mkdir -p documents
echo "AgenticGoKit is a Go framework for building multi-agent systems. It provides orchestration, memory, and tool integration capabilities." > documents/agenticgokit.txt
echo "Vector databases store high-dimensional vectors and enable semantic search. Popular options include pgvector, Weaviate, and Pinecone." > documents/vectors.txt

# Ingest documents
go run . --mode ingest --path documents/
```

### 4. Query the Knowledge Base

```bash
# Ask questions about your documents
go run . --mode query --question "What is AgenticGoKit?"
go run . --mode query --question "Tell me about vector databases"
go run . --mode query --question "How do I build multi-agent systems?"
```

## Code Structure

```
04-rag-knowledge-base/
├── main.go                 # Application entry point
├── agents/
│   ├── ingestion.go       # Document ingestion agent
│   ├── retrieval.go       # Information retrieval agent
│   └── synthesis.go       # Answer synthesis agent
├── services/
│   ├── embeddings.go      # Embedding service
│   ├── vectorstore.go     # Vector storage interface
│   └── chunking.go        # Document chunking
├── documents/             # Sample documents directory
├── docker-compose.yml     # PostgreSQL + pgvector setup
├── init-db.sql           # Database initialization
├── setup.sh              # Setup script
├── .env.example          # Environment template
└── README.md             # This file
```

## Key Components

### Document Ingestion Agent

Processes documents and stores them in the vector database:

```go
type IngestionAgent struct {
    name         string
    embeddings   *services.EmbeddingService
    vectorStore  *services.VectorStore
    chunker      *services.DocumentChunker
}

func (a *IngestionAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    filePath := event.Data["file_path"].(string)
    
    // Read document
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    // Chunk document
    chunks := a.chunker.ChunkText(string(content), 1000, 100)
    
    // Generate embeddings and store
    var storedChunks []string
    for _, chunk := range chunks {
        embedding, err := a.embeddings.GenerateEmbedding(ctx, chunk)
        if err != nil {
            log.Printf("Failed to generate embedding for chunk: %v", err)
            continue
        }
        
        id, err := a.vectorStore.Store(ctx, chunk, embedding, map[string]interface{}{
            "source": filePath,
            "type":   "document_chunk",
        })
        if err != nil {
            log.Printf("Failed to store chunk: %v", err)
            continue
        }
        
        storedChunks = append(storedChunks, id)
    }
    
    return &core.AgentResult{
        Data: map[string]interface{}{
            "file_processed": filePath,
            "chunks_stored":  len(storedChunks),
            "chunk_ids":      storedChunks,
        },
    }, nil
}
```

### Retrieval Agent

Searches for relevant information based on queries:

```go
type RetrievalAgent struct {
    name         string
    embeddings   *services.EmbeddingService
    vectorStore  *services.VectorStore
}

func (a *RetrievalAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    query := event.Data["query"].(string)
    
    // Generate query embedding
    queryEmbedding, err := a.embeddings.GenerateEmbedding(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to generate query embedding: %w", err)
    }
    
    // Search for similar documents
    results, err := a.vectorStore.Search(ctx, queryEmbedding, 5, 0.7)
    if err != nil {
        return nil, fmt.Errorf("failed to search vector store: %w", err)
    }
    
    // Extract relevant context
    var contexts []string
    var sources []string
    for _, result := range results {
        contexts = append(contexts, result.Content)
        if source, ok := result.Metadata["source"].(string); ok {
            sources = append(sources, source)
        }
    }
    
    return &core.AgentResult{
        Data: map[string]interface{}{
            "query":           query,
            "retrieved_contexts": contexts,
            "sources":         sources,
            "relevance_scores": extractScores(results),
        },
    }, nil
}
```

### Synthesis Agent

Combines retrieved context with the query to generate comprehensive answers:

```go
type SynthesisAgent struct {
    name        string
    llmProvider core.ModelProvider
}

func (a *SynthesisAgent) Execute(ctx context.Context, event core.Event, state *core.State) (*core.AgentResult, error) {
    query := event.Data["query"].(string)
    contexts := event.Data["retrieved_contexts"].([]string)
    sources := event.Data["sources"].([]string)
    
    // Build RAG prompt
    contextText := strings.Join(contexts, "\n\n")
    prompt := fmt.Sprintf(`Based on the following context, answer the question comprehensively and accurately.

Context:
%s

Question: %s

Instructions:
- Use only information from the provided context
- If the context doesn't contain enough information, say so
- Cite sources when possible
- Provide a clear, well-structured answer

Answer:`, contextText, query)
    
    // Generate response
    response, err := a.llmProvider.GenerateResponse(ctx, prompt, map[string]interface{}{
        "max_tokens":   1000,
        "temperature": 0.1, // Low temperature for factual responses
    })
    if err != nil {
        return nil, fmt.Errorf("failed to generate response: %w", err)
    }
    
    return &core.AgentResult{
        Data: map[string]interface{}{
            "question":     query,
            "answer":       response,
            "sources":      sources,
            "context_used": len(contexts),
        },
    }, nil
}
```

## Advanced Features

### Hybrid Search

Combine semantic and keyword search for better results:

```go
func (vs *VectorStore) HybridSearch(ctx context.Context, query string, embedding []float32, limit int, threshold float64) ([]SearchResult, error) {
    // Semantic search
    semanticResults, err := vs.Search(ctx, embedding, limit, threshold)
    if err != nil {
        return nil, err
    }
    
    // Keyword search using PostgreSQL full-text search
    keywordResults, err := vs.KeywordSearch(ctx, query, limit)
    if err != nil {
        return nil, err
    }
    
    // Combine and rank results
    combined := combineResults(semanticResults, keywordResults, 0.7, 0.3) // 70% semantic, 30% keyword
    
    return combined[:min(limit, len(combined))], nil
}
```

### Document Metadata Extraction

Extract and store document metadata for better filtering:

```go
type DocumentProcessor struct {
    chunker *DocumentChunker
}

func (dp *DocumentProcessor) ProcessDocument(filePath string) (*Document, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    
    // Extract metadata
    metadata := map[string]interface{}{
        "file_path":    filePath,
        "file_size":    len(content),
        "processed_at": time.Now(),
        "file_type":    filepath.Ext(filePath),
    }
    
    // Add content-based metadata
    if title := extractTitle(string(content)); title != "" {
        metadata["title"] = title
    }
    
    if summary := generateSummary(string(content)); summary != "" {
        metadata["summary"] = summary
    }
    
    return &Document{
        Content:  string(content),
        Metadata: metadata,
        Chunks:   dp.chunker.ChunkText(string(content), 1000, 100),
    }, nil
}
```

### Query Enhancement

Improve queries before retrieval:

```go
type QueryEnhancer struct {
    llmProvider core.ModelProvider
}

func (qe *QueryEnhancer) EnhanceQuery(ctx context.Context, originalQuery string) (string, error) {
    prompt := fmt.Sprintf(`Improve this search query to be more specific and likely to find relevant information:

Original query: "%s"

Enhanced query (return only the improved query, no explanation):`, originalQuery)
    
    enhanced, err := qe.llmProvider.GenerateResponse(ctx, prompt, map[string]interface{}{
        "max_tokens":   100,
        "temperature": 0.3,
    })
    if err != nil {
        return originalQuery, err // Fallback to original
    }
    
    return strings.TrimSpace(enhanced), nil
}
```

## Performance Optimization

### Batch Processing

Process multiple documents efficiently:

```bash
# Batch ingest multiple documents
go run . --mode batch-ingest --path documents/ --batch-size 10
```

### Caching

Cache embeddings and search results:

```go
type CachedEmbeddingService struct {
    service *EmbeddingService
    cache   map[string][]float32
    mutex   sync.RWMutex
}

func (ces *CachedEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
    // Check cache first
    ces.mutex.RLock()
    if cached, exists := ces.cache[text]; exists {
        ces.mutex.RUnlock()
        return cached, nil
    }
    ces.mutex.RUnlock()
    
    // Generate and cache
    embedding, err := ces.service.GenerateEmbedding(ctx, text)
    if err == nil {
        ces.mutex.Lock()
        ces.cache[text] = embedding
        ces.mutex.Unlock()
    }
    
    return embedding, err
}
```

## Usage Examples

### Basic Q&A

```bash
# Simple question answering
go run . --mode query --question "What is machine learning?"
```

### Document-Specific Queries

```bash
# Query specific document types
go run . --mode query --question "What does the API documentation say about authentication?" --filter "type:api_doc"
```

### Batch Processing

```bash
# Process multiple documents
go run . --mode batch-ingest --path ./documents --pattern "*.pdf,*.txt,*.md"
```

### Interactive Mode

```bash
# Start interactive session
go run . --mode interactive
> What is AgenticGoKit?
> How do I set up vector databases?
> exit
```

## Monitoring and Metrics

The example includes basic monitoring:

```go
// Track ingestion metrics
type IngestionMetrics struct {
    DocumentsProcessed int64
    ChunksStored      int64
    ProcessingTime    time.Duration
    Errors            int64
}

// Track query metrics
type QueryMetrics struct {
    QueriesProcessed  int64
    AverageLatency    time.Duration
    CacheHitRate      float64
    RetrievalAccuracy float64
}
```

## Testing

Run the included tests:

```bash
# Unit tests
go test ./...

# Integration tests (requires running database)
go test -tags=integration ./...

# Benchmark tests
go test -bench=. ./...
```

## Troubleshooting

### Common Issues

**Database connection failed:**
```bash
# Check if PostgreSQL is running
docker compose ps
docker compose logs postgres

# Restart services
docker compose down && docker compose up -d
```

**Embeddings not generating:**
```bash
# Check API key
echo $OPENAI_API_KEY

# Test Ollama connection
curl http://localhost:11434/api/tags
```

**Poor search results:**
- Adjust similarity threshold (lower = more results)
- Try different chunk sizes
- Use hybrid search for better coverage
- Check if documents were properly ingested

### Performance Tuning

**For large document collections:**
- Increase batch sizes for ingestion
- Use connection pooling for database
- Implement result caching
- Consider using HNSW indexes for faster search

**For high query volume:**
- Enable query result caching
- Use read replicas for the database
- Implement query batching
- Add rate limiting

## Next Steps

This example provides a foundation for building production RAG systems. Consider:

1. **Adding more document types** (PDF, Word, web scraping)
2. **Implementing user sessions** for personalized results
3. **Adding evaluation metrics** for answer quality
4. **Scaling with multiple databases** or sharding
5. **Adding real-time document updates** with change detection

## Related Examples

- [Simple Agent](../01-simple-agent/) - Basic agent concepts
- [Multi-Agent Collaboration](../02-multi-agent-collab/) - Agent orchestration
- [Production System](../05-production-system/) - Deployment and monitoring

---

*This example demonstrates current AgenticGoKit capabilities for building knowledge systems. The framework is actively developed, so features may evolve.*