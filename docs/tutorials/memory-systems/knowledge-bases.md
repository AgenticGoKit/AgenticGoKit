# Knowledge Bases in AgenticGoKit

## Overview

Knowledge bases are structured repositories of information that enable agents to access, search, and reason over large collections of documents and data. This tutorial covers building comprehensive knowledge bases with AgenticGoKit, including document ingestion, chunking strategies, metadata management, and search optimization.

Knowledge bases transform raw information into accessible, searchable knowledge that agents can use to provide accurate and contextual responses.

## Prerequisites

- Understanding of [RAG Implementation](rag-implementation.md)
- Familiarity with [Vector Databases](vector-databases.md)
- Knowledge of document processing and text extraction
- Basic understanding of information retrieval concepts

## Knowledge Base Architecture

### Components Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Documents     │───▶│   Ingestion      │───▶│   Processing    │
│   (PDF, MD,     │    │   Pipeline       │    │   Pipeline      │
│    HTML, etc.)  │    └──────────────────┘    └─────────────────┘
└─────────────────┘                                      │
                                                         ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Search &      │◀───│   Vector Store   │◀───│   Chunking &    │
│   Retrieval     │    │   (pgvector/     │    │   Embedding     │
└─────────────────┘    │    Weaviate)     │    └─────────────────┘
                       └──────────────────┘
```

### Knowledge Base Layers

1. **Storage Layer**: Vector database with metadata
2. **Processing Layer**: Document parsing and chunking
3. **Embedding Layer**: Vector representation generation
4. **Retrieval Layer**: Search and ranking algorithms
5. **Management Layer**: Updates, versioning, and maintenance

## Document Ingestion Pipeline

### 1. Basic Document Processor

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

type DocumentProcessor struct {
    memory     core.Memory
    parsers    map[string]DocumentParser
    chunker    *DocumentChunker
    config     ProcessingConfig
}

type ProcessingConfig struct {
    ChunkSize        int
    ChunkOverlap     int
    MaxFileSize      int64
    SupportedFormats []string
    BatchSize        int
}

type DocumentParser interface {
    Parse(filePath string) (*Document, error)
    SupportedExtensions() []string
}

type Document struct {
    ID          string
    Title       string
    Content     string
    Metadata    map[string]string
    Source      string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func NewDocumentProcessor(memory core.Memory) *DocumentProcessor {
    dp := &DocumentProcessor{
        memory:  memory,
        parsers: make(map[string]DocumentParser),
        chunker: NewDocumentChunker(ChunkingConfig{
            ChunkSize:    1000,
            ChunkOverlap: 200,
            Strategy:     "semantic",
        }),
        config: ProcessingConfig{
            ChunkSize:        1000,
            ChunkOverlap:     200,
            MaxFileSize:      10 * 1024 * 1024, // 10MB
            SupportedFormats: []string{".txt", ".md", ".pdf", ".html"},
            BatchSize:        10,
        },
    }
    
    // Register parsers
    dp.registerParsers()
    return dp
}

func (dp *DocumentProcessor) registerParsers() {
    dp.parsers[".txt"] = &TextParser{}
    dp.parsers[".md"] = &MarkdownParser{}
    dp.parsers[".pdf"] = &PDFParser{}
    dp.parsers[".html"] = &HTMLParser{}
}

func (dp *DocumentProcessor) ProcessFile(ctx context.Context, filePath string) error {
    // Check file size
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return fmt.Errorf("failed to stat file: %w", err)
    }
    
    if fileInfo.Size() > dp.config.MaxFileSize {
        return fmt.Errorf("file too large: %d bytes", fileInfo.Size())
    }
    
    // Get parser for file extension
    ext := strings.ToLower(filepath.Ext(filePath))
    parser, exists := dp.parsers[ext]
    if !exists {
        return fmt.Errorf("unsupported file format: %s", ext)
    }
    
    // Parse document
    doc, err := parser.Parse(filePath)
    if err != nil {
        return fmt.Errorf("failed to parse document: %w", err)
    }
    
    // Add file metadata
    doc.Metadata["file_path"] = filePath
    doc.Metadata["file_size"] = fmt.Sprintf("%d", fileInfo.Size())
    doc.Metadata["processed_at"] = time.Now().Format(time.RFC3339)
    
    // Process document
    return dp.ProcessDocument(ctx, doc)
}

func (dp *DocumentProcessor) ProcessDocument(ctx context.Context, doc *Document) error {
    // Chunk the document
    chunks, err := dp.chunker.ChunkDocument(doc)
    if err != nil {
        return fmt.Errorf("failed to chunk document: %w", err)
    }
    
    log.Printf("Processing document '%s' with %d chunks", doc.Title, len(chunks))
    
    // Store chunks in batches
    for i := 0; i < len(chunks); i += dp.config.BatchSize {
        end := i + dp.config.BatchSize
        if end > len(chunks) {
            end = len(chunks)
        }
        
        batch := chunks[i:end]
        err := dp.storeBatch(ctx, batch)
        if err != nil {
            return fmt.Errorf("failed to store batch: %w", err)
        }
    }
    
    log.Printf("Successfully processed document '%s'", doc.Title)
    return nil
}

func (dp *DocumentProcessor) storeBatch(ctx context.Context, chunks []*DocumentChunk) error {
    for _, chunk := range chunks {
        err := dp.memory.Store(ctx, chunk.Content, "document-chunk",
            core.WithMetadata(chunk.Metadata),
            core.WithTimestamp(chunk.CreatedAt),
        )
        if err != nil {
            return fmt.Errorf("failed to store chunk %s: %w", chunk.ID, err)
        }
    }
    return nil
}

// Simple text parser
type TextParser struct{}

func (tp *TextParser) Parse(filePath string) (*Document, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    
    return &Document{
        ID:        generateDocumentID(filePath),
        Title:     filepath.Base(filePath),
        Content:   string(content),
        Source:    filePath,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Metadata: map[string]string{
            "format": "text",
            "parser": "text",
        },
    }, nil
}

func (tp *TextParser) SupportedExtensions() []string {
    return []string{".txt"}
}

// Markdown parser
type MarkdownParser struct{}

func (mp *MarkdownParser) Parse(filePath string) (*Document, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    
    // Extract title from first heading
    lines := strings.Split(string(content), "\n")
    title := filepath.Base(filePath)
    
    for _, line := range lines {
        if strings.HasPrefix(line, "# ") {
            title = strings.TrimPrefix(line, "# ")
            break
        }
    }
    
    return &Document{
        ID:        generateDocumentID(filePath),
        Title:     title,
        Content:   string(content),
        Source:    filePath,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Metadata: map[string]string{
            "format": "markdown",
            "parser": "markdown",
        },
    }, nil
}

func (mp *MarkdownParser) SupportedExtensions() []string {
    return []string{".md", ".markdown"}
}

func generateDocumentID(filePath string) string {
    // Generate unique ID based on file path
    hash := sha256.Sum256([]byte(filePath + time.Now().String()))
    return fmt.Sprintf("%x", hash)[:16]
}
```

## Document Chunking Strategies

### 1. Document Chunker

```go
type DocumentChunker struct {
    config ChunkingConfig
}

type ChunkingConfig struct {
    ChunkSize    int
    ChunkOverlap int
    Strategy     string // "fixed", "semantic", "sentence", "paragraph"
    MinChunkSize int
    MaxChunkSize int
}

type DocumentChunk struct {
    ID          string
    DocumentID  string
    Content     string
    ChunkIndex  int
    StartOffset int
    EndOffset   int
    Metadata    map[string]string
    CreatedAt   time.Time
}

func NewDocumentChunker(config ChunkingConfig) *DocumentChunker {
    return &DocumentChunker{config: config}
}

func (dc *DocumentChunker) ChunkDocument(doc *Document) ([]*DocumentChunk, error) {
    switch dc.config.Strategy {
    case "fixed":
        return dc.fixedSizeChunking(doc)
    case "semantic":
        return dc.semanticChunking(doc)
    case "sentence":
        return dc.sentenceChunking(doc)
    case "paragraph":
        return dc.paragraphChunking(doc)
    default:
        return dc.fixedSizeChunking(doc)
    }
}

func (dc *DocumentChunker) fixedSizeChunking(doc *Document) ([]*DocumentChunk, error) {
    content := doc.Content
    chunks := make([]*DocumentChunk, 0)
    
    for i := 0; i < len(content); i += dc.config.ChunkSize - dc.config.ChunkOverlap {
        end := i + dc.config.ChunkSize
        if end > len(content) {
            end = len(content)
        }
        
        chunkContent := content[i:end]
        
        // Skip chunks that are too small
        if len(chunkContent) < dc.config.MinChunkSize {
            continue
        }
        
        chunk := &DocumentChunk{
            ID:          fmt.Sprintf("%s_chunk_%d", doc.ID, len(chunks)),
            DocumentID:  doc.ID,
            Content:     chunkContent,
            ChunkIndex:  len(chunks),
            StartOffset: i,
            EndOffset:   end,
            CreatedAt:   time.Now(),
            Metadata: map[string]string{
                "document_id":    doc.ID,
                "document_title": doc.Title,
                "document_source": doc.Source,
                "chunk_strategy": "fixed",
                "chunk_index":    fmt.Sprintf("%d", len(chunks)),
            },
        }
        
        chunks = append(chunks, chunk)
        
        if end >= len(content) {
            break
        }
    }
    
    return chunks, nil
}

func (dc *DocumentChunker) semanticChunking(doc *Document) ([]*DocumentChunk, error) {
    content := doc.Content
    
    // Split by double newlines (paragraphs)
    paragraphs := strings.Split(content, "\n\n")
    
    chunks := make([]*DocumentChunk, 0)
    currentChunk := ""
    startOffset := 0
    
    for _, paragraph := range paragraphs {
        paragraph = strings.TrimSpace(paragraph)
        if paragraph == "" {
            continue
        }
        
        // Check if adding this paragraph would exceed chunk size
        if len(currentChunk)+len(paragraph)+2 > dc.config.ChunkSize && currentChunk != "" {
            // Create chunk from current content
            chunk := dc.createChunk(doc, currentChunk, len(chunks), startOffset, startOffset+len(currentChunk))
            chunks = append(chunks, chunk)
            
            // Start new chunk with overlap
            overlapSize := min(dc.config.ChunkOverlap, len(currentChunk))
            currentChunk = currentChunk[len(currentChunk)-overlapSize:] + "\n\n" + paragraph
            startOffset = startOffset + len(currentChunk) - overlapSize
        } else {
            // Add paragraph to current chunk
            if currentChunk != "" {
                currentChunk += "\n\n"
            }
            currentChunk += paragraph
        }
    }
    
    // Add final chunk if there's content
    if currentChunk != "" && len(currentChunk) >= dc.config.MinChunkSize {
        chunk := dc.createChunk(doc, currentChunk, len(chunks), startOffset, startOffset+len(currentChunk))
        chunks = append(chunks, chunk)
    }
    
    return chunks, nil
}

func (dc *DocumentChunker) createChunk(doc *Document, content string, index, startOffset, endOffset int) *DocumentChunk {
    return &DocumentChunk{
        ID:          fmt.Sprintf("%s_chunk_%d", doc.ID, index),
        DocumentID:  doc.ID,
        Content:     content,
        ChunkIndex:  index,
        StartOffset: startOffset,
        EndOffset:   endOffset,
        CreatedAt:   time.Now(),
        Metadata: map[string]string{
            "document_id":    doc.ID,
            "document_title": doc.Title,
            "document_source": doc.Source,
            "chunk_strategy": dc.config.Strategy,
            "chunk_index":    fmt.Sprintf("%d", index),
        },
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

## Knowledge Base Management

### 1. Knowledge Base Manager

```go
type KnowledgeBaseManager struct {
    memory     core.Memory
    processor  *DocumentProcessor
    config     ManagerConfig
}

type ManagerConfig struct {
    AutoIndexing     bool
    IndexingInterval time.Duration
    BackupEnabled    bool
    BackupInterval   time.Duration
}

func NewKnowledgeBaseManager(memory core.Memory) *KnowledgeBaseManager {
    processor := NewDocumentProcessor(memory)
    
    return &KnowledgeBaseManager{
        memory:    memory,
        processor: processor,
        config: ManagerConfig{
            AutoIndexing:     true,
            IndexingInterval: 1 * time.Hour,
            BackupEnabled:    true,
            BackupInterval:   24 * time.Hour,
        },
    }
}

func (kbm *KnowledgeBaseManager) AddDocument(ctx context.Context, filePath string) error {
    return kbm.processor.ProcessFile(ctx, filePath)
}

func (kbm *KnowledgeBaseManager) AddDocumentFromContent(ctx context.Context, title, content string, metadata map[string]string) error {
    doc := &Document{
        ID:        generateDocumentID(title + content),
        Title:     title,
        Content:   content,
        Source:    "direct-input",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Metadata:  metadata,
    }
    
    return kbm.processor.ProcessDocument(ctx, doc)
}

func (kbm *KnowledgeBaseManager) Search(ctx context.Context, query string, options ...core.SearchOption) ([]core.MemoryResult, error) {
    return kbm.memory.Search(ctx, query, options...)
}

func (kbm *KnowledgeBaseManager) GetStats(ctx context.Context) (*KnowledgeBaseStats, error) {
    memoryStats, err := kbm.memory.GetStats(ctx)
    if err != nil {
        return nil, err
    }
    
    return &KnowledgeBaseStats{
        TotalDocuments: memoryStats.ItemCount,
        TotalChunks:    memoryStats.ItemCount,
        IndexSize:      memoryStats.SizeBytes,
        LastUpdated:    time.Now(),
    }, nil
}

type KnowledgeBaseStats struct {
    TotalDocuments int64     `json:"total_documents"`
    TotalChunks    int64     `json:"total_chunks"`
    IndexSize      int64     `json:"index_size_bytes"`
    LastUpdated    time.Time `json:"last_updated"`
}
```

## Usage Example

### Complete Knowledge Base Example

```go
func main() {
    // Setup memory with vector database
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "pgvector",
        Connection: "postgres://user:pass@localhost:5432/agentdb",
        EnableRAG:  true,
        Dimensions: 1536,
        Embedding: core.EmbeddingConfig{
            Provider:   "openai",
            Model:      "text-embedding-3-small",
            APIKey:     os.Getenv("OPENAI_API_KEY"),
            Dimensions: 1536,
        },
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    
    // Create knowledge base manager
    kbManager := NewKnowledgeBaseManager(memory)
    
    ctx := context.Background()
    
    // Add documents to knowledge base
    documents := []string{
        "./docs/tutorial1.md",
        "./docs/tutorial2.md",
        "./docs/api-reference.md",
    }
    
    for _, docPath := range documents {
        err := kbManager.AddDocument(ctx, docPath)
        if err != nil {
            log.Printf("Failed to add document %s: %v", docPath, err)
        } else {
            log.Printf("Successfully added document: %s", docPath)
        }
    }
    
    // Add content directly
    err = kbManager.AddDocumentFromContent(ctx,
        "AgenticGoKit Overview",
        "AgenticGoKit is a Go framework for building multi-agent systems...",
        map[string]string{
            "category": "overview",
            "author":   "AgenticGoKit Team",
        },
    )
    if err != nil {
        log.Printf("Failed to add content: %v", err)
    }
    
    // Search the knowledge base
    results, err := kbManager.Search(ctx, "How to build multi-agent systems?",
        core.WithLimit(5),
        core.WithScoreThreshold(0.7),
    )
    if err != nil {
        log.Printf("Search failed: %v", err)
    } else {
        fmt.Printf("Found %d results:\n", len(results))
        for i, result := range results {
            fmt.Printf("%d. %s (Score: %.3f)\n", i+1, result.Content[:100]+"...", result.Score)
        }
    }
    
    // Get knowledge base statistics
    stats, err := kbManager.GetStats(ctx)
    if err != nil {
        log.Printf("Failed to get stats: %v", err)
    } else {
        fmt.Printf("Knowledge Base Stats:\n")
        fmt.Printf("  Documents: %d\n", stats.TotalDocuments)
        fmt.Printf("  Chunks: %d\n", stats.TotalChunks)
        fmt.Printf("  Index Size: %d MB\n", stats.IndexSize/1024/1024)
    }
}
```

## Best Practices

### 1. Document Processing

- **Format Support**: Implement parsers for all relevant document formats
- **Error Handling**: Gracefully handle parsing errors and corrupted files
- **Batch Processing**: Process multiple documents efficiently
- **Progress Tracking**: Provide feedback on processing progress
- **Validation**: Validate documents before processing

### 2. Chunking Strategy

- **Content-Aware**: Use semantic chunking for better context preservation
- **Overlap Management**: Balance overlap size with storage efficiency
- **Size Optimization**: Optimize chunk size for your embedding model
- **Metadata Preservation**: Maintain document context in chunks
- **Quality Control**: Validate chunk quality and coherence

### 3. Search Optimization

- **Index Tuning**: Optimize vector database indexes
- **Query Enhancement**: Improve query understanding
- **Result Ranking**: Implement effective ranking algorithms
- **Caching**: Cache frequent searches
- **Performance Monitoring**: Track search performance metrics

## Conclusion

Knowledge bases in AgenticGoKit provide the foundation for intelligent information retrieval and RAG systems. Key takeaways:

- Design comprehensive document processing pipelines
- Implement appropriate chunking strategies for your content
- Use rich metadata to enhance search and filtering
- Optimize search performance through indexing and caching
- Monitor and maintain knowledge base quality over time

Well-designed knowledge bases enable agents to access and utilize vast amounts of information effectively, making them more knowledgeable and helpful.

## Next Steps

- [Memory Optimization](memory-optimization.md) - Advanced performance tuning
- [Production Deployment](../README.md) - Deploy knowledge bases at scale
- [Monitoring and Observability](../README.md) - Monitor knowledge base performance

## Further Reading

- [Information Retrieval Fundamentals](https://nlp.stanford.edu/IR-book/)
- [Vector Database Comparison](https://github.com/pgvector/pgvector)
- [Document Processing Libraries](https://github.com/unidoc/unipdf)
