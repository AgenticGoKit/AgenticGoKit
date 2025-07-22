# Document Ingestion and Knowledge Base Management

## Overview

Document ingestion is a critical component of building comprehensive knowledge bases in AgenticGoKit. This tutorial covers the complete pipeline from raw documents to searchable knowledge, including document processing, chunking strategies, metadata extraction, and optimization techniques.

Effective document ingestion enables agents to access and reason over large collections of structured and unstructured data.

## Prerequisites

- Understanding of [Memory Systems Overview](README.md)
- Familiarity with [Vector Databases](vector-databases.md)
- Knowledge of document formats (PDF, Markdown, HTML, etc.)
- Basic understanding of text processing and NLP concepts

## Document Ingestion Pipeline

### Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Raw           │───▶│   Document       │───▶│   Text          │
│   Documents     │    │   Parser         │    │   Extraction    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Vector        │◀───│   Embedding      │◀───│   Text          │
│   Storage       │    │   Generation     │    │   Chunking      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
                                                ┌─────────────────┐
                                                │   Metadata      │
                                                │   Extraction    │
                                                └─────────────────┘
```

## Document Types and Processing

### 1. Supported Document Types

```go
// Document types supported by AgenticGoKit
const (
    DocumentTypePDF      DocumentType = "pdf"
    DocumentTypeText     DocumentType = "txt"
    DocumentTypeMarkdown DocumentType = "md"
    DocumentTypeWeb      DocumentType = "web"
    DocumentTypeCode     DocumentType = "code"
    DocumentTypeJSON     DocumentType = "json"
)

// Document structure for ingestion
type Document struct {
    ID         string         `json:"id"`
    Title      string         `json:"title,omitempty"`
    Content    string         `json:"content"`
    Source     string         `json:"source,omitempty"` // URL, file path, etc.
    Type       DocumentType   `json:"type,omitempty"`
    Metadata   map[string]any `json:"metadata,omitempty"`
    Tags       []string       `json:"tags,omitempty"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at,omitempty"`
    ChunkIndex int            `json:"chunk_index,omitempty"` // For chunked documents
    ChunkTotal int            `json:"chunk_total,omitempty"`
}
```

### 2. Basic Document Ingestion

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func ingestBasicDocument(memory core.Memory) error {
    ctx := context.Background()
    
    // Create a document
    doc := core.Document{
        ID:      "doc-001",
        Title:   "Introduction to Machine Learning",
        Content: `Machine learning is a subset of artificial intelligence that enables computers to learn and make decisions from data without being explicitly programmed for every task. It involves algorithms that can identify patterns, make predictions, and improve their performance over time.`,
        Source:  "textbook-chapter-1.pdf",
        Type:    core.DocumentTypePDF,
        Metadata: map[string]any{
            "author":     "Dr. Jane Smith",
            "chapter":    1,
            "subject":    "machine-learning",
            "difficulty": "beginner",
            "language":   "english",
        },
        Tags:      []string{"ml", "ai", "introduction"},
        CreatedAt: time.Now(),
    }
    
    // Ingest the document
    err := memory.IngestDocument(ctx, doc)
    if err != nil {
        return fmt.Errorf("failed to ingest document: %w", err)
    }
    
    fmt.Printf("Successfully ingested document: %s\n", doc.Title)
    return nil
}
```

### 3. Batch Document Ingestion

```go
func ingestMultipleDocuments(memory core.Memory) error {
    ctx := context.Background()
    
    // Prepare multiple documents
    documents := []core.Document{
        {
            ID:      "doc-002",
            Title:   "Neural Networks Fundamentals",
            Content: "Neural networks are computing systems inspired by biological neural networks...",
            Source:  "textbook-chapter-2.pdf",
            Type:    core.DocumentTypePDF,
            Metadata: map[string]any{
                "author":     "Dr. Jane Smith",
                "chapter":    2,
                "subject":    "neural-networks",
                "difficulty": "intermediate",
            },
            Tags: []string{"neural-networks", "deep-learning"},
        },
        {
            ID:      "doc-003",
            Title:   "Data Preprocessing Techniques",
            Content: "Data preprocessing is a crucial step in machine learning pipelines...",
            Source:  "textbook-chapter-3.pdf",
            Type:    core.DocumentTypePDF,
            Metadata: map[string]any{
                "author":     "Dr. Jane Smith",
                "chapter":    3,
                "subject":    "data-preprocessing",
                "difficulty": "beginner",
            },
            Tags: []string{"data-science", "preprocessing"},
        },
    }
    
    // Batch ingest documents
    err := memory.IngestDocuments(ctx, documents)
    if err != nil {
        return fmt.Errorf("failed to ingest documents: %w", err)
    }
    
    fmt.Printf("Successfully ingested %d documents\n", len(documents))
    return nil
}
```

## Text Chunking Strategies

### 1. Fixed-Size Chunking

```go
type FixedSizeChunker struct {
    ChunkSize    int
    ChunkOverlap int
}

func NewFixedSizeChunker(chunkSize, overlap int) *FixedSizeChunker {
    return &FixedSizeChunker{
        ChunkSize:    chunkSize,
        ChunkOverlap: overlap,
    }
}

func (c *FixedSizeChunker) ChunkText(text string) []string {
    if len(text) <= c.ChunkSize {
        return []string{text}
    }
    
    var chunks []string
    start := 0
    
    for start < len(text) {
        end := start + c.ChunkSize
        if end > len(text) {
            end = len(text)
        }
        
        chunk := text[start:end]
        chunks = append(chunks, chunk)
        
        // Move start position considering overlap
        start += c.ChunkSize - c.ChunkOverlap
        if start >= len(text) {
            break
        }
    }
    
    return chunks
}

// Example usage
func chunkLargeDocument(memory core.Memory, largeText string) error {
    ctx := context.Background()
    chunker := NewFixedSizeChunker(1000, 200)
    
    chunks := chunker.ChunkText(largeText)
    
    for i, chunk := range chunks {
        doc := core.Document{
            ID:         fmt.Sprintf("large-doc-chunk-%d", i),
            Title:      fmt.Sprintf("Large Document - Chunk %d", i+1),
            Content:    chunk,
            Source:     "large-document.pdf",
            Type:       core.DocumentTypePDF,
            ChunkIndex: i,
            ChunkTotal: len(chunks),
            Metadata: map[string]any{
                "chunk_method": "fixed-size",
                "chunk_size":   1000,
                "chunk_overlap": 200,
            },
            CreatedAt: time.Now(),
        }
        
        err := memory.IngestDocument(ctx, doc)
        if err != nil {
            return fmt.Errorf("failed to ingest chunk %d: %w", i, err)
        }
    }
    
    return nil
}
```

### 2. Semantic Chunking

```go
type SemanticChunker struct {
    MaxChunkSize int
    MinChunkSize int
}

func NewSemanticChunker(minSize, maxSize int) *SemanticChunker {
    return &SemanticChunker{
        MinChunkSize: minSize,
        MaxChunkSize: maxSize,
    }
}

func (c *SemanticChunker) ChunkText(text string) []string {
    // Split by paragraphs first
    paragraphs := strings.Split(text, "\n\n")
    
    var chunks []string
    var currentChunk strings.Builder
    
    for _, paragraph := range paragraphs {
        paragraph = strings.TrimSpace(paragraph)
        if paragraph == "" {
            continue
        }
        
        // Check if adding this paragraph would exceed max size
        if currentChunk.Len() > 0 && 
           currentChunk.Len()+len(paragraph) > c.MaxChunkSize {
            
            // Finalize current chunk if it meets minimum size
            if currentChunk.Len() >= c.MinChunkSize {
                chunks = append(chunks, currentChunk.String())
                currentChunk.Reset()
            }
        }
        
        // Add paragraph to current chunk
        if currentChunk.Len() > 0 {
            currentChunk.WriteString("\n\n")
        }
        currentChunk.WriteString(paragraph)
    }
    
    // Add final chunk if it has content
    if currentChunk.Len() > 0 {
        chunks = append(chunks, currentChunk.String())
    }
    
    return chunks
}
```

### 3. Sentence-Based Chunking

```go
type SentenceChunker struct {
    MaxSentences int
    Overlap      int
}

func NewSentenceChunker(maxSentences, overlap int) *SentenceChunker {
    return &SentenceChunker{
        MaxSentences: maxSentences,
        Overlap:      overlap,
    }
}

func (c *SentenceChunker) ChunkText(text string) []string {
    sentences := c.splitIntoSentences(text)
    
    if len(sentences) <= c.MaxSentences {
        return []string{text}
    }
    
    var chunks []string
    start := 0
    
    for start < len(sentences) {
        end := start + c.MaxSentences
        if end > len(sentences) {
            end = len(sentences)
        }
        
        chunk := strings.Join(sentences[start:end], " ")
        chunks = append(chunks, chunk)
        
        start += c.MaxSentences - c.Overlap
        if start >= len(sentences) {
            break
        }
    }
    
    return chunks
}

func (c *SentenceChunker) splitIntoSentences(text string) []string {
    // Simple sentence splitting (in production, use a proper NLP library)
    sentences := strings.FieldsFunc(text, func(r rune) bool {
        return r == '.' || r == '!' || r == '?'
    })
    
    // Clean up sentences
    var cleanSentences []string
    for _, sentence := range sentences {
        sentence = strings.TrimSpace(sentence)
        if len(sentence) > 10 { // Filter out very short fragments
            cleanSentences = append(cleanSentences, sentence)
        }
    }
    
    return cleanSentences
}
```

## Advanced Document Processing

### 1. Document Processor with Multiple Strategies

```go
type DocumentProcessor struct {
    memory      core.Memory
    chunkers    map[string]TextChunker
    extractors  map[core.DocumentType]MetadataExtractor
    config      ProcessorConfig
}

type TextChunker interface {
    ChunkText(text string) []string
}

type MetadataExtractor interface {
    ExtractMetadata(doc core.Document) (map[string]any, error)
}

type ProcessorConfig struct {
    DefaultChunkStrategy string
    MaxConcurrentDocs    int
    EnableMetadataExtraction bool
    EnableContentCleaning    bool
}

func NewDocumentProcessor(memory core.Memory, config ProcessorConfig) *DocumentProcessor {
    dp := &DocumentProcessor{
        memory:     memory,
        chunkers:   make(map[string]TextChunker),
        extractors: make(map[core.DocumentType]MetadataExtractor),
        config:     config,
    }
    
    // Register default chunkers
    dp.chunkers["fixed"] = NewFixedSizeChunker(1000, 200)
    dp.chunkers["semantic"] = NewSemanticChunker(500, 1500)
    dp.chunkers["sentence"] = NewSentenceChunker(10, 2)
    
    // Register metadata extractors
    dp.extractors[core.DocumentTypePDF] = &PDFMetadataExtractor{}
    dp.extractors[core.DocumentTypeMarkdown] = &MarkdownMetadataExtractor{}
    dp.extractors[core.DocumentTypeCode] = &CodeMetadataExtractor{}
    
    return dp
}

func (dp *DocumentProcessor) ProcessDocument(ctx context.Context, doc core.Document, chunkStrategy string) error {
    // Clean content if enabled
    if dp.config.EnableContentCleaning {
        doc.Content = dp.cleanContent(doc.Content)
    }
    
    // Extract metadata if enabled
    if dp.config.EnableMetadataExtraction {
        if extractor, exists := dp.extractors[doc.Type]; exists {
            metadata, err := extractor.ExtractMetadata(doc)
            if err == nil {
                // Merge extracted metadata with existing
                if doc.Metadata == nil {
                    doc.Metadata = make(map[string]any)
                }
                for k, v := range metadata {
                    doc.Metadata[k] = v
                }
            }
        }
    }
    
    // Choose chunking strategy
    if chunkStrategy == "" {
        chunkStrategy = dp.config.DefaultChunkStrategy
    }
    
    chunker, exists := dp.chunkers[chunkStrategy]
    if !exists {
        return fmt.Errorf("unknown chunking strategy: %s", chunkStrategy)
    }
    
    // Chunk the document
    chunks := chunker.ChunkText(doc.Content)
    
    // Process chunks
    if len(chunks) == 1 {
        // Single chunk - ingest as-is
        return dp.memory.IngestDocument(ctx, doc)
    }
    
    // Multiple chunks - create separate documents
    var documents []core.Document
    for i, chunk := range chunks {
        chunkDoc := doc // Copy original document
        chunkDoc.ID = fmt.Sprintf("%s-chunk-%d", doc.ID, i)
        chunkDoc.Content = chunk
        chunkDoc.ChunkIndex = i
        chunkDoc.ChunkTotal = len(chunks)
        
        // Add chunking metadata
        if chunkDoc.Metadata == nil {
            chunkDoc.Metadata = make(map[string]any)
        }
        chunkDoc.Metadata["chunk_strategy"] = chunkStrategy
        chunkDoc.Metadata["original_doc_id"] = doc.ID
        
        documents = append(documents, chunkDoc)
    }
    
    return dp.memory.IngestDocuments(ctx, documents)
}

func (dp *DocumentProcessor) cleanContent(content string) string {
    // Remove excessive whitespace
    content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
    
    // Remove special characters that might interfere with processing
    content = regexp.MustCompile(`[^\w\s\.,!?;:()\-"']`).ReplaceAllString(content, "")
    
    // Trim whitespace
    content = strings.TrimSpace(content)
    
    return content
}
```

### 2. Metadata Extractors

```go
// PDF Metadata Extractor
type PDFMetadataExtractor struct{}

func (e *PDFMetadataExtractor) ExtractMetadata(doc core.Document) (map[string]any, error) {
    metadata := make(map[string]any)
    
    // Extract basic statistics
    metadata["word_count"] = len(strings.Fields(doc.Content))
    metadata["char_count"] = len(doc.Content)
    metadata["paragraph_count"] = len(strings.Split(doc.Content, "\n\n"))
    
    // Extract potential headings (lines that are short and followed by longer content)
    lines := strings.Split(doc.Content, "\n")
    var headings []string
    for i, line := range lines {
        line = strings.TrimSpace(line)
        if len(line) > 0 && len(line) < 100 && i+1 < len(lines) {
            nextLine := strings.TrimSpace(lines[i+1])
            if len(nextLine) > len(line)*2 {
                headings = append(headings, line)
            }
        }
    }
    metadata["potential_headings"] = headings
    
    // Detect language (simple heuristic)
    metadata["detected_language"] = detectLanguage(doc.Content)
    
    return metadata, nil
}

// Markdown Metadata Extractor
type MarkdownMetadataExtractor struct{}

func (e *MarkdownMetadataExtractor) ExtractMetadata(doc core.Document) (map[string]any, error) {
    metadata := make(map[string]any)
    
    // Extract headings
    headings := extractMarkdownHeadings(doc.Content)
    metadata["headings"] = headings
    metadata["heading_count"] = len(headings)
    
    // Extract links
    links := extractMarkdownLinks(doc.Content)
    metadata["links"] = links
    metadata["link_count"] = len(links)
    
    // Extract code blocks
    codeBlocks := extractMarkdownCodeBlocks(doc.Content)
    metadata["code_blocks"] = len(codeBlocks)
    
    // Extract front matter if present
    frontMatter := extractFrontMatter(doc.Content)
    if frontMatter != nil {
        metadata["front_matter"] = frontMatter
    }
    
    return metadata, nil
}

// Code Metadata Extractor
type CodeMetadataExtractor struct{}

func (e *CodeMetadataExtractor) ExtractMetadata(doc core.Document) (map[string]any, error) {
    metadata := make(map[string]any)
    
    // Detect programming language
    language := detectProgrammingLanguage(doc.Source, doc.Content)
    metadata["programming_language"] = language
    
    // Count lines of code
    lines := strings.Split(doc.Content, "\n")
    metadata["total_lines"] = len(lines)
    
    // Count non-empty lines
    nonEmptyLines := 0
    commentLines := 0
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" {
            nonEmptyLines++
            if isCommentLine(line, language) {
                commentLines++
            }
        }
    }
    metadata["code_lines"] = nonEmptyLines
    metadata["comment_lines"] = commentLines
    
    // Extract functions/methods (basic pattern matching)
    functions := extractFunctions(doc.Content, language)
    metadata["functions"] = functions
    metadata["function_count"] = len(functions)
    
    return metadata, nil
}

// Helper functions for metadata extraction
func detectLanguage(content string) string {
    // Simple language detection based on common words
    englishWords := []string{"the", "and", "is", "in", "to", "of", "a", "that", "it", "with"}
    
    words := strings.Fields(strings.ToLower(content))
    englishCount := 0
    
    for _, word := range words {
        for _, englishWord := range englishWords {
            if word == englishWord {
                englishCount++
                break
            }
        }
    }
    
    if float64(englishCount)/float64(len(words)) > 0.1 {
        return "english"
    }
    
    return "unknown"
}

func extractMarkdownHeadings(content string) []string {
    var headings []string
    lines := strings.Split(content, "\n")
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "#") {
            headings = append(headings, line)
        }
    }
    
    return headings
}

func extractMarkdownLinks(content string) []string {
    // Simple regex for markdown links [text](url)
    linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
    matches := linkRegex.FindAllStringSubmatch(content, -1)
    
    var links []string
    for _, match := range matches {
        if len(match) >= 3 {
            links = append(links, match[2]) // URL part
        }
    }
    
    return links
}

func extractMarkdownCodeBlocks(content string) []string {
    // Simple extraction of code blocks
    codeBlockRegex := regexp.MustCompile("```[\\s\\S]*?```")
    matches := codeBlockRegex.FindAllString(content, -1)
    return matches
}

func extractFrontMatter(content string) map[string]any {
    // Extract YAML front matter
    if !strings.HasPrefix(content, "---") {
        return nil
    }
    
    parts := strings.SplitN(content, "---", 3)
    if len(parts) < 3 {
        return nil
    }
    
    // Simple key-value extraction (in production, use a YAML parser)
    frontMatter := make(map[string]any)
    lines := strings.Split(parts[1], "\n")
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.Contains(line, ":") {
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                key := strings.TrimSpace(parts[0])
                value := strings.TrimSpace(parts[1])
                frontMatter[key] = value
            }
        }
    }
    
    return frontMatter
}

func detectProgrammingLanguage(filename, content string) string {
    // Detect by file extension
    ext := strings.ToLower(filepath.Ext(filename))
    switch ext {
    case ".go":
        return "go"
    case ".py":
        return "python"
    case ".js":
        return "javascript"
    case ".ts":
        return "typescript"
    case ".java":
        return "java"
    case ".cpp", ".cc", ".cxx":
        return "cpp"
    case ".c":
        return "c"
    case ".rs":
        return "rust"
    }
    
    // Detect by content patterns
    if strings.Contains(content, "package main") || strings.Contains(content, "func ") {
        return "go"
    }
    if strings.Contains(content, "def ") || strings.Contains(content, "import ") {
        return "python"
    }
    
    return "unknown"
}

func isCommentLine(line, language string) bool {
    switch language {
    case "go", "javascript", "typescript", "java", "cpp", "c", "rust":
        return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*")
    case "python":
        return strings.HasPrefix(line, "#")
    }
    return false
}

func extractFunctions(content, language string) []string {
    var functions []string
    
    switch language {
    case "go":
        funcRegex := regexp.MustCompile(`func\s+(\w+)\s*\(`)
        matches := funcRegex.FindAllStringSubmatch(content, -1)
        for _, match := range matches {
            if len(match) >= 2 {
                functions = append(functions, match[1])
            }
        }
    case "python":
        funcRegex := regexp.MustCompile(`def\s+(\w+)\s*\(`)
        matches := funcRegex.FindAllStringSubmatch(content, -1)
        for _, match := range matches {
            if len(match) >= 2 {
                functions = append(functions, match[1])
            }
        }
    case "javascript", "typescript":
        funcRegex := regexp.MustCompile(`function\s+(\w+)\s*\(`)
        matches := funcRegex.FindAllStringSubmatch(content, -1)
        for _, match := range matches {
            if len(match) >= 2 {
                functions = append(functions, match[1])
            }
        }
    }
    
    return functions
}
```

## Knowledge Base Search and Retrieval

### 1. Advanced Search with Filters

```go
func performAdvancedKnowledgeSearch(memory core.Memory) error {
    ctx := context.Background()
    
    // Search with multiple filters
    results, err := memory.SearchKnowledge(ctx, "machine learning algorithms",
        core.WithLimit(10),
        core.WithScoreThreshold(0.7),
        core.WithSources([]string{"textbook-chapter-1.pdf", "textbook-chapter-2.pdf"}),
        core.WithDocumentTypes([]core.DocumentType{core.DocumentTypePDF}),
        core.WithTags([]string{"ml", "algorithms"}),
        core.WithDateRange(&core.DateRange{
            Start: time.Now().Add(-30 * 24 * time.Hour),
            End:   time.Now(),
        }),
    )
    if err != nil {
        return fmt.Errorf("knowledge search failed: %w", err)
    }
    
    fmt.Printf("Found %d relevant knowledge items:\n", len(results))
    for _, result := range results {
        fmt.Printf("- %s (Score: %.3f)\n", result.Title, result.Score)
        fmt.Printf("  Source: %s\n", result.Source)
        fmt.Printf("  Content: %s...\n", truncateString(result.Content, 100))
        
        if result.ChunkIndex > 0 {
            fmt.Printf("  Chunk: %d/%d\n", result.ChunkIndex+1, result.ChunkTotal)
        }
        
        fmt.Println()
    }
    
    return nil
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}
```

### 2. Hybrid Search (Personal + Knowledge)

```go
func performHybridSearch(memory core.Memory) error {
    ctx := context.Context()
    
    // Perform hybrid search combining personal memory and knowledge base
    result, err := memory.SearchAll(ctx, "neural network implementation",
        core.WithLimit(15),
        core.WithScoreThreshold(0.6),
        core.WithIncludePersonal(true),
        core.WithIncludeKnowledge(true),
        core.WithHybridWeight(0.7), // 70% semantic, 30% keyword
    )
    if err != nil {
        return fmt.Errorf("hybrid search failed: %w", err)
    }
    
    fmt.Printf("Hybrid Search Results for: %s\n", result.Query)
    fmt.Printf("Total Results: %d (Search Time: %v)\n\n", result.TotalResults, result.SearchTime)
    
    // Display personal memory results
    if len(result.PersonalMemory) > 0 {
        fmt.Println("Personal Memory Results:")
        for _, item := range result.PersonalMemory {
            fmt.Printf("- %s (Score: %.3f)\n", truncateString(item.Content, 80), item.Score)
        }
        fmt.Println()
    }
    
    // Display knowledge base results
    if len(result.Knowledge) > 0 {
        fmt.Println("Knowledge Base Results:")
        for _, item := range result.Knowledge {
            fmt.Printf("- %s (Score: %.3f)\n", item.Title, item.Score)
            fmt.Printf("  Source: %s\n", item.Source)
        }
    }
    
    return nil
}
```

### 3. RAG Context Building

```go
func buildRAGContext(memory core.Memory, query string) error {
    ctx := context.Background()
    
    // Build comprehensive RAG context
    ragContext, err := memory.BuildContext(ctx, query,
        core.WithMaxTokens(4000),
        core.WithPersonalWeight(0.3),
        core.WithKnowledgeWeight(0.7),
        core.WithHistoryLimit(5),
        core.WithIncludeSources(true),
        core.WithFormatTemplate(`Context Information:

Personal Memory:
{{range .PersonalMemory}}
- {{.Content}}
{{end}}

Knowledge Base:
{{range .Knowledge}}
- {{.Content}} (Source: {{.Source}})
{{end}}

Recent Conversation:
{{range .ChatHistory}}
{{.Role}}: {{.Content}}
{{end}}

Query: {{.Query}}`),
    )
    if err != nil {
        return fmt.Errorf("failed to build RAG context: %w", err)
    }
    
    fmt.Printf("RAG Context for: %s\n", ragContext.Query)
    fmt.Printf("Token Count: %d\n", ragContext.TokenCount)
    fmt.Printf("Sources: %v\n", ragContext.Sources)
    fmt.Printf("Context Text:\n%s\n", ragContext.ContextText)
    
    return nil
}
```

## Production Optimization

### 1. Batch Processing Pipeline

```go
type BatchProcessor struct {
    memory       core.Memory
    processor    *DocumentProcessor
    concurrency  int
    batchSize    int
}

func NewBatchProcessor(memory core.Memory, concurrency, batchSize int) *BatchProcessor {
    return &BatchProcessor{
        memory:      memory,
        processor:   NewDocumentProcessor(memory, ProcessorConfig{
            DefaultChunkStrategy:     "semantic",
            MaxConcurrentDocs:        concurrency,
            EnableMetadataExtraction: true,
            EnableContentCleaning:    true,
        }),
        concurrency: concurrency,
        batchSize:   batchSize,
    }
}

func (bp *BatchProcessor) ProcessDocuments(ctx context.Context, documents []core.Document) error {
    // Process documents in batches
    for i := 0; i < len(documents); i += bp.batchSize {
        end := i + bp.batchSize
        if end > len(documents) {
            end = len(documents)
        }
        
        batch := documents[i:end]
        err := bp.processBatch(ctx, batch)
        if err != nil {
            return fmt.Errorf("failed to process batch %d-%d: %w", i, end-1, err)
        }
        
        fmt.Printf("Processed batch %d-%d (%d documents)\n", i, end-1, len(batch))
    }
    
    return nil
}

func (bp *BatchProcessor) processBatch(ctx context.Context, documents []core.Document) error {
    // Use worker pool for concurrent processing
    jobs := make(chan core.Document, len(documents))
    results := make(chan error, len(documents))
    
    // Start workers
    for w := 0; w < bp.concurrency; w++ {
        go bp.worker(ctx, jobs, results)
    }
    
    // Send jobs
    for _, doc := range documents {
        jobs <- doc
    }
    close(jobs)
    
    // Collect results
    var errors []error
    for i := 0; i < len(documents); i++ {
        if err := <-results; err != nil {
            errors = append(errors, err)
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("batch processing failed with %d errors: %v", len(errors), errors[0])
    }
    
    return nil
}

func (bp *BatchProcessor) worker(ctx context.Context, jobs <-chan core.Document, results chan<- error) {
    for doc := range jobs {
        err := bp.processor.ProcessDocument(ctx, doc, "")
        results <- err
    }
}
```

### 2. Performance Monitoring

```go
type IngestionMetrics struct {
    DocumentsProcessed int64         `json:"documents_processed"`
    ChunksCreated     int64         `json:"chunks_created"`
    ProcessingTime    time.Duration `json:"processing_time"`
    ErrorCount        int64         `json:"error_count"`
    AverageChunkSize  float64       `json:"average_chunk_size"`
    mu                sync.RWMutex
}

func (m *IngestionMetrics) RecordDocument(chunkCount int, processingTime time.Duration, chunkSizes []int) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.DocumentsProcessed++
    m.ChunksCreated += int64(chunkCount)
    m.ProcessingTime += processingTime
    
    // Update average chunk size
    if len(chunkSizes) > 0 {
        totalSize := 0
        for _, size := range chunkSizes {
            totalSize += size
        }
        avgSize := float64(totalSize) / float64(len(chunkSizes))
        
        // Running average
        totalChunks := float64(m.ChunksCreated)
        m.AverageChunkSize = (m.AverageChunkSize*(totalChunks-float64(chunkCount)) + avgSize*float64(chunkCount)) / totalChunks
    }
}

func (m *IngestionMetrics) RecordError() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.ErrorCount++
}

func (m *IngestionMetrics) GetStats() IngestionMetrics {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    return IngestionMetrics{
        DocumentsProcessed: m.DocumentsProcessed,
        ChunksCreated:     m.ChunksCreated,
        ProcessingTime:    m.ProcessingTime,
        ErrorCount:        m.ErrorCount,
        AverageChunkSize:  m.AverageChunkSize,
    }
}
```

## Best Practices

### 1. Document Ingestion Guidelines

- **Chunk Size**: Balance between context preservation and retrieval precision
- **Overlap**: Use 10-20% overlap to maintain context continuity
- **Metadata**: Extract and store rich metadata for better filtering
- **Batch Processing**: Process documents in batches for better performance
- **Error Handling**: Implement robust error handling and retry mechanisms

### 2. Performance Optimization

- **Concurrent Processing**: Use worker pools for parallel document processing
- **Embedding Caching**: Cache embeddings to avoid recomputation
- **Index Optimization**: Optimize vector database indexes for your query patterns
- **Memory Management**: Monitor memory usage during large batch operations

### 3. Quality Assurance

- **Content Validation**: Validate document content before ingestion
- **Duplicate Detection**: Implement deduplication to avoid redundant storage
- **Quality Metrics**: Track ingestion quality and search relevance
- **Regular Maintenance**: Periodically clean up and optimize the knowledge base

## Conclusion

Document ingestion and knowledge base management are critical for building effective RAG systems. By implementing proper chunking strategies, metadata extraction, and optimization techniques, you can create knowledge bases that provide accurate and relevant information to your agents.

Key takeaways:
- Choose appropriate chunking strategies based on your content type
- Extract rich metadata to enable better filtering and search
- Implement batch processing for handling large document collections
- Monitor performance and optimize based on usage patterns
- Follow best practices for quality and maintenance

## Next Steps

- [RAG Implementation](rag-implementation.md) - Build complete RAG systems
- [Memory Optimization](memory-optimization.md) - Optimize performance and scaling
- [Vector Databases](vector-databases.md) - Advanced database configuration

## Further Reading

- [API Reference: Document Interface](../../api/core.md#document)
- [Examples: Document Processing](../../examples/)
- [Configuration Guide: Memory Settings](../../guides/Configuration.md)