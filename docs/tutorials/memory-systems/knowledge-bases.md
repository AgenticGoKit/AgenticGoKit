---
title: Knowledge Bases
description: Learn how to build comprehensive knowledge bases using AgenticGoKit's current SearchKnowledge API and advanced search patterns.
---

# Knowledge Bases in AgenticGoKit

## Overview

Knowledge bases are structured repositories of information that enable agents to access, search, and reason over large collections of documents and data. This tutorial covers building comprehensive knowledge bases with AgenticGoKit using the current SearchKnowledge API, advanced search patterns, and production deployment strategies.

Knowledge bases transform raw information into accessible, searchable knowledge that agents can use to provide accurate and contextual responses.

## Prerequisites

- Understanding of [RAG Implementation](rag-implementation.md)
- Familiarity with [Document Ingestion](document-ingestion.md)
- Knowledge of [Vector Databases](vector-databases.md)
- Basic understanding of information retrieval concepts

## Knowledge Base Architecture

### Current API Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Documents     │───▶│ IngestDocument/  │───▶│   Vector        │
│   (PDF, MD,     │    │ IngestDocuments  │    │   Storage       │
│    HTML, etc.)  │    └──────────────────┘    └─────────────────┘
└─────────────────┘                                      │
                                                         ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ SearchKnowledge │◀───│   Search with    │◀───│   Embedding &   │
│ with Options    │    │   Filters        │    │   Indexing      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Knowledge Base Layers

1. **Ingestion Layer**: Document processing with `IngestDocument`/`IngestDocuments`
2. **Storage Layer**: Vector database with metadata and tags
3. **Search Layer**: `SearchKnowledge` with `SearchOption` functions
4. **Retrieval Layer**: Advanced filtering and ranking
5. **Management Layer**: Updates, versioning, and maintenance

> Note: While the documentation references multiple document types (PDF, Markdown, HTML), native PDF processing is coming soon. The current CLI/document processors natively support text and Markdown; PDF processor integration will be added in a future release.

### Current PDF reader — limitations and recommended fallbacks

The repository includes a minimal PDF extraction implementation intended as an MVP to enable basic PDF ingestion. Please be aware of these limitations when including PDFs in your knowledge base today:

- No OCR support: The current reader extracts only embedded text. Scanned documents (image-only pages) will not yield usable text unless OCR is applied prior to ingestion.
- Layout fidelity: Multi-column documents, tables, and complex layouts may produce reordered or jumbled text when converted to plain text. This affects retrieval quality for content heavily reliant on structure.
- Basic metadata only: The MVP provides page counts and simple file stats; it does not extract detailed embedded metadata or images/figures.
- Encrypted PDFs: Password-protected PDFs are not currently supported and will fail with an error.

Recommended approaches if you rely on PDFs:

- Preprocess scanned PDFs with an OCR step (e.g., Tesseract) and ingest the resulting text files.
- Use `pdftotext` (Poppler) for more robust text extraction when available on your platform.
- For production-quality extraction (tables, images, layout), consider using a commercial parser such as UniDoc or an external preprocessing pipeline that extracts and normalizes content into plain text or per-page files before ingestion.

We plan to enhance the PDF processor with optional OCR, page-aware chunking, and image extraction in future releases; these changes will be documented in the migration notes when ready.

## Advanced Search Patterns

### 1. Comprehensive Search with All Options

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "sync"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func demonstrateAdvancedSearch(memory core.Memory) error {
    ctx := context.Background()
    
    // Advanced search using all available SearchOption functions
    results, err := memory.SearchKnowledge(ctx, "machine learning algorithms",
        // Basic search options
        core.WithLimit(20),                    // Get more results
        core.WithScoreThreshold(0.75),         // Higher quality threshold
        
        // Content filtering
        core.WithTags([]string{"ml", "algorithms", "supervised"}), // Filter by tags
        core.WithDocumentTypes([]core.DocumentType{
            core.DocumentTypePDF,
            core.DocumentTypeMarkdown,
            core.DocumentTypeText,
        }), // Filter by document types
        core.WithSources([]string{
            "ml-textbook.pdf",
            "algorithms-guide.md",
            "research-papers",
        }), // Filter by specific sources
    )
    if err != nil {
        return fmt.Errorf("advanced search failed: %w", err)
    }
    
    fmt.Printf("Advanced search found %d results:\n", len(results))
    for i, result := range results {
        fmt.Printf("%d. %s (Score: %.3f)\n", i+1, result.Title, result.Score)
        fmt.Printf("   Source: %s, Document ID: %s\n", result.Source, result.DocumentID)
        fmt.Printf("   Tags: %v\n", result.Tags)
        if result.ChunkIndex > 0 {
            fmt.Printf("   Chunk: %d/%d\n", result.ChunkIndex+1, result.ChunkTotal)
        }
        fmt.Printf("   Content: %s...\n", truncateString(result.Content, 100))
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

### 2. Multi-Modal Knowledge Base Search

```go
type MultiModalKnowledgeBase struct {
    memory core.Memory
}

func NewMultiModalKnowledgeBase(memory core.Memory) *MultiModalKnowledgeBase {
    return &MultiModalKnowledgeBase{memory: memory}
}

func (mk *MultiModalKnowledgeBase) SearchByContentType(ctx context.Context, query string, contentTypes []core.DocumentType) ([]core.KnowledgeResult, error) {
    // Search specific content types
    results, err := mk.memory.SearchKnowledge(ctx, query,
        core.WithLimit(15),
        core.WithScoreThreshold(0.6),
        core.WithDocumentTypes(contentTypes),
    )
    if err != nil {
        return nil, fmt.Errorf("content type search failed: %w", err)
    }
    
    return results, nil
}

func (mk *MultiModalKnowledgeBase) SearchCodeDocuments(ctx context.Context, query string, programmingLanguage string) ([]core.KnowledgeResult, error) {
    // Search specifically in code documents
    results, err := mk.memory.SearchKnowledge(ctx, query,
        core.WithLimit(10),
        core.WithDocumentTypes([]core.DocumentType{core.DocumentTypeCode}),
        core.WithTags([]string{programmingLanguage, "code", "implementation"}),
        core.WithScoreThreshold(0.7),
    )
    if err != nil {
        return nil, fmt.Errorf("code search failed: %w", err)
    }
    
    return results, nil
}

func (mk *MultiModalKnowledgeBase) SearchResearchPapers(ctx context.Context, query string, researchArea string) ([]core.KnowledgeResult, error) {
    // Search in research papers with academic focus
    results, err := mk.memory.SearchKnowledge(ctx, query,
        core.WithLimit(8),
        core.WithDocumentTypes([]core.DocumentType{core.DocumentTypePDF}),
        core.WithTags([]string{"research", "academic", researchArea}),
        core.WithScoreThreshold(0.8), // Higher threshold for research
        core.WithSources([]string{"arxiv", "ieee", "acm", "research-papers"}),
    )
    if err != nil {
        return nil, fmt.Errorf("research paper search failed: %w", err)
    }
    
    return results, nil
}

func (mk *MultiModalKnowledgeBase) SearchTutorials(ctx context.Context, query string, difficulty string) ([]core.KnowledgeResult, error) {
    // Search for tutorial content
    results, err := mk.memory.SearchKnowledge(ctx, query,
        core.WithLimit(12),
        core.WithDocumentTypes([]core.DocumentType{
            core.DocumentTypeMarkdown,
            core.DocumentTypeWeb,
        }),
        core.WithTags([]string{"tutorial", "guide", difficulty}),
        core.WithScoreThreshold(0.65),
    )
    if err != nil {
        return nil, fmt.Errorf("tutorial search failed: %w", err)
    }
    
    return results, nil
}

// Example usage of multi-modal search
func demonstrateMultiModalSearch(mk *MultiModalKnowledgeBase) error {
    ctx := context.Background()
    
    // Search for Python code examples
    codeResults, err := mk.SearchCodeDocuments(ctx, "neural network implementation", "python")
    if err != nil {
        return err
    }
    fmt.Printf("Found %d Python code examples\n", len(codeResults))
    
    // Search for machine learning research papers
    researchResults, err := mk.SearchResearchPapers(ctx, "transformer architecture", "deep-learning")
    if err != nil {
        return err
    }
    fmt.Printf("Found %d research papers\n", len(researchResults))
    
    // Search for beginner tutorials
    tutorialResults, err := mk.SearchTutorials(ctx, "getting started with machine learning", "beginner")
    if err != nil {
        return err
    }
    fmt.Printf("Found %d beginner tutorials\n", len(tutorialResults))
    
    return nil
}
```

### 3. Contextual and Semantic Search Patterns

```go
type ContextualSearchEngine struct {
    memory core.Memory
}

func NewContextualSearchEngine(memory core.Memory) *ContextualSearchEngine {
    return &ContextualSearchEngine{memory: memory}
}

func (cs *ContextualSearchEngine) SearchWithContext(ctx context.Context, query string, contextTags []string, userLevel string) ([]core.KnowledgeResult, error) {
    // Build search options based on context
    searchOptions := []core.SearchOption{
        core.WithLimit(15),
        core.WithScoreThreshold(0.7),
    }
    
    // Add context-based tags
    if len(contextTags) > 0 {
        searchOptions = append(searchOptions, core.WithTags(contextTags))
    }
    
    // Adjust search based on user level
    switch userLevel {
    case "beginner":
        searchOptions = append(searchOptions, 
            core.WithTags([]string{"beginner", "introduction", "basics"}),
            core.WithScoreThreshold(0.6), // Lower threshold for more results
        )
    case "intermediate":
        searchOptions = append(searchOptions,
            core.WithTags([]string{"intermediate", "practical", "examples"}),
        )
    case "advanced":
        searchOptions = append(searchOptions,
            core.WithTags([]string{"advanced", "research", "technical"}),
            core.WithScoreThreshold(0.8), // Higher threshold for quality
        )
    }
    
    results, err := cs.memory.SearchKnowledge(ctx, query, searchOptions...)
    if err != nil {
        return nil, fmt.Errorf("contextual search failed: %w", err)
    }
    
    return results, nil
}

func (cs *ContextualSearchEngine) SearchSimilarConcepts(ctx context.Context, baseQuery string, conceptArea string) ([]core.KnowledgeResult, error) {
    // Search for related concepts in the same area
    results, err := cs.memory.SearchKnowledge(ctx, baseQuery,
        core.WithLimit(20),
        core.WithScoreThreshold(0.6), // Lower threshold to find related concepts
        core.WithTags([]string{conceptArea, "related", "concepts"}),
    )
    if err != nil {
        return nil, fmt.Errorf("similar concepts search failed: %w", err)
    }
    
    return results, nil
}

func (cs *ContextualSearchEngine) SearchByTopicHierarchy(ctx context.Context, query string, topicPath []string) ([]core.KnowledgeResult, error) {
    // Search within a specific topic hierarchy
    // topicPath example: ["computer-science", "machine-learning", "neural-networks"]
    
    searchOptions := []core.SearchOption{
        core.WithLimit(12),
        core.WithScoreThreshold(0.7),
        core.WithTags(topicPath), // Use topic path as tags
    }
    
    results, err := cs.memory.SearchKnowledge(ctx, query, searchOptions...)
    if err != nil {
        return nil, fmt.Errorf("topic hierarchy search failed: %w", err)
    }
    
    return results, nil
}
```

### 4. Production Knowledge Base Management

```go
type ProductionKnowledgeBase struct {
    memory      core.Memory
    config      ProductionConfig
    metrics     *KnowledgeBaseMetrics
    cache       *SearchCache
}

type ProductionConfig struct {
    DefaultSearchLimit      int
    DefaultScoreThreshold   float32
    CacheEnabled           bool
    CacheTTL               time.Duration
    MetricsEnabled         bool
    MaxConcurrentSearches  int
}

type KnowledgeBaseMetrics struct {
    TotalSearches       int64         `json:"total_searches"`
    AverageResponseTime time.Duration `json:"average_response_time"`
    CacheHitRate        float64       `json:"cache_hit_rate"`
    PopularQueries      []string      `json:"popular_queries"`
    mu                  sync.RWMutex
}

type SearchCache struct {
    cache      map[string][]core.KnowledgeResult
    timestamps map[string]time.Time
    ttl        time.Duration
    mu         sync.RWMutex
}

func NewProductionKnowledgeBase(memory core.Memory, config ProductionConfig) *ProductionKnowledgeBase {
    pkb := &ProductionKnowledgeBase{
        memory:  memory,
        config:  config,
        metrics: &KnowledgeBaseMetrics{},
    }
    
    if config.CacheEnabled {
        pkb.cache = &SearchCache{
            cache:      make(map[string][]core.KnowledgeResult),
            timestamps: make(map[string]time.Time),
            ttl:        config.CacheTTL,
        }
        go pkb.cache.cleanup()
    }
    
    return pkb
}

func (pkb *ProductionKnowledgeBase) Search(ctx context.Context, query string, options ...core.SearchOption) ([]core.KnowledgeResult, error) {
    start := time.Now()
    
    // Check cache first
    cacheKey := pkb.buildCacheKey(query, options)
    if pkb.config.CacheEnabled {
        if cached := pkb.cache.Get(cacheKey); cached != nil {
            pkb.updateMetrics(start, true, query)
            return cached, nil
        }
    }
    
    // Apply default options if not specified
    searchOptions := pkb.applyDefaults(options)
    
    // Perform search
    results, err := pkb.memory.SearchKnowledge(ctx, query, searchOptions...)
    if err != nil {
        return nil, fmt.Errorf("knowledge base search failed: %w", err)
    }
    
    // Cache results
    if pkb.config.CacheEnabled {
        pkb.cache.Set(cacheKey, results)
    }
    
    // Update metrics
    if pkb.config.MetricsEnabled {
        pkb.updateMetrics(start, false, query)
    }
    
    return results, nil
}

func (pkb *ProductionKnowledgeBase) applyDefaults(options []core.SearchOption) []core.SearchOption {
    // Check if limit is already specified
    hasLimit := false
    hasThreshold := false
    
    for _, option := range options {
        // This is a simplified check - in practice, you'd need to inspect the options
        // For now, we'll assume defaults are needed
    }
    
    if !hasLimit {
        options = append(options, core.WithLimit(pkb.config.DefaultSearchLimit))
    }
    
    if !hasThreshold {
        options = append(options, core.WithScoreThreshold(pkb.config.DefaultScoreThreshold))
    }
    
    return options
}

func (pkb *ProductionKnowledgeBase) buildCacheKey(query string, options []core.SearchOption) string {
    // Simple cache key generation - in production, use more sophisticated hashing
    return fmt.Sprintf("%s_%d", query, len(options))
}

func (pkb *ProductionKnowledgeBase) updateMetrics(start time.Time, cacheHit bool, query string) {
    pkb.metrics.mu.Lock()
    defer pkb.metrics.mu.Unlock()
    
    pkb.metrics.TotalSearches++
    
    // Update average response time
    responseTime := time.Since(start)
    if pkb.metrics.AverageResponseTime == 0 {
        pkb.metrics.AverageResponseTime = responseTime
    } else {
        pkb.metrics.AverageResponseTime = (pkb.metrics.AverageResponseTime + responseTime) / 2
    }
    
    // Update cache hit rate
    if cacheHit {
        pkb.metrics.CacheHitRate = (pkb.metrics.CacheHitRate*float64(pkb.metrics.TotalSearches-1) + 1.0) / float64(pkb.metrics.TotalSearches)
    } else {
        pkb.metrics.CacheHitRate = (pkb.metrics.CacheHitRate * float64(pkb.metrics.TotalSearches-1)) / float64(pkb.metrics.TotalSearches)
    }
    
    // Track popular queries (simplified)
    if len(pkb.metrics.PopularQueries) < 10 {
        pkb.metrics.PopularQueries = append(pkb.metrics.PopularQueries, query)
    }
}

func (pkb *ProductionKnowledgeBase) GetMetrics() *KnowledgeBaseMetrics {
    pkb.metrics.mu.RLock()
    defer pkb.metrics.mu.RUnlock()
    
    // Return a copy
    return &KnowledgeBaseMetrics{
        TotalSearches:       pkb.metrics.TotalSearches,
        AverageResponseTime: pkb.metrics.AverageResponseTime,
        CacheHitRate:        pkb.metrics.CacheHitRate,
        PopularQueries:      append([]string{}, pkb.metrics.PopularQueries...),
    }
}

// Cache implementation
func (sc *SearchCache) Get(key string) []core.KnowledgeResult {
    sc.mu.RLock()
    defer sc.mu.RUnlock()
    
    if results, exists := sc.cache[key]; exists {
        if timestamp, ok := sc.timestamps[key]; ok {
            if time.Since(timestamp) < sc.ttl {
                return results
            }
        }
    }
    return nil
}

func (sc *SearchCache) Set(key string, results []core.KnowledgeResult) {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    
    sc.cache[key] = results
    sc.timestamps[key] = time.Now()
}

func (sc *SearchCache) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        sc.mu.Lock()
        now := time.Now()
        for key, timestamp := range sc.timestamps {
            if now.Sub(timestamp) > sc.ttl {
                delete(sc.cache, key)
                delete(sc.timestamps, key)
            }
        }
        sc.mu.Unlock()
    }
}
```

### 5. Knowledge Base Analytics and Insights

```go
type KnowledgeBaseAnalytics struct {
    memory core.Memory
}

func NewKnowledgeBaseAnalytics(memory core.Memory) *KnowledgeBaseAnalytics {
    return &KnowledgeBaseAnalytics{memory: memory}
}

func (kba *KnowledgeBaseAnalytics) AnalyzeContentDistribution(ctx context.Context) (*ContentDistribution, error) {
    // Analyze content by document types
    distribution := &ContentDistribution{
        ByDocumentType: make(map[core.DocumentType]int),
        BySource:       make(map[string]int),
        ByTags:         make(map[string]int),
    }
    
    // Search for all content with different document types
    documentTypes := []core.DocumentType{
        core.DocumentTypePDF,
        core.DocumentTypeMarkdown,
        core.DocumentTypeText,
        core.DocumentTypeCode,
        core.DocumentTypeWeb,
        core.DocumentTypeJSON,
    }
    
    for _, docType := range documentTypes {
        results, err := kba.memory.SearchKnowledge(ctx, "",
            core.WithLimit(1000), // Get many results for analysis
            core.WithScoreThreshold(0.0), // Include all results
            core.WithDocumentTypes([]core.DocumentType{docType}),
        )
        if err != nil {
            log.Printf("Failed to analyze document type %s: %v", docType, err)
            continue
        }
        
        distribution.ByDocumentType[docType] = len(results)
        
        // Analyze sources and tags
        for _, result := range results {
            if result.Source != "" {
                distribution.BySource[result.Source]++
            }
            for _, tag := range result.Tags {
                distribution.ByTags[tag]++
            }
        }
    }
    
    return distribution, nil
}

type ContentDistribution struct {
    ByDocumentType map[core.DocumentType]int `json:"by_document_type"`
    BySource       map[string]int            `json:"by_source"`
    ByTags         map[string]int            `json:"by_tags"`
}

func (kba *KnowledgeBaseAnalytics) FindContentGaps(ctx context.Context, expectedTopics []string) ([]string, error) {
    var gaps []string
    
    for _, topic := range expectedTopics {
        results, err := kba.memory.SearchKnowledge(ctx, topic,
            core.WithLimit(5),
            core.WithScoreThreshold(0.7),
            core.WithTags([]string{topic}),
        )
        if err != nil {
            log.Printf("Failed to search for topic %s: %v", topic, err)
            continue
        }
        
        if len(results) == 0 {
            gaps = append(gaps, topic)
        }
    }
    
    return gaps, nil
}

func (kba *KnowledgeBaseAnalytics) GetTopSources(ctx context.Context, limit int) ([]SourceInfo, error) {
    // This is a simplified implementation
    // In practice, you'd need to aggregate across all documents
    
    results, err := kba.memory.SearchKnowledge(ctx, "",
        core.WithLimit(1000),
        core.WithScoreThreshold(0.0),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to get all documents: %w", err)
    }
    
    sourceCounts := make(map[string]int)
    for _, result := range results {
        if result.Source != "" {
            sourceCounts[result.Source]++
        }
    }
    
    // Convert to sorted slice
    var sources []SourceInfo
    for source, count := range sourceCounts {
        sources = append(sources, SourceInfo{
            Source: source,
            Count:  count,
        })
    }
    
    // Sort by count (simplified - in practice, use proper sorting)
    // Return top sources up to limit
    if len(sources) > limit {
        sources = sources[:limit]
    }
    
    return sources, nil
}

type SourceInfo struct {
    Source string `json:"source"`
    Count  int    `json:"count"`
}
```

## Complete Knowledge Base Implementation

### 1. Enterprise Knowledge Base System

```go
func main() {
    // Create memory with production configuration
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "pgvector", // Use production vector database
        Connection: "postgres://user:pass@localhost:5432/knowledge_db",
        MaxResults: 10,
        Dimensions: 1536,
        AutoEmbed:  true,
        
        // RAG-enhanced settings for knowledge base
        EnableRAG:               true,
        EnableKnowledgeBase:     true,
        KnowledgeMaxResults:     50,  // Higher for comprehensive search
        KnowledgeScoreThreshold: 0.7,
        ChunkSize:               1500, // Larger chunks for better context
        ChunkOverlap:            300,
        
        // RAG context assembly settings
        RAGMaxContextTokens: 6000, // Larger context for knowledge base
        RAGPersonalWeight:   0.1,  // Focus on knowledge base
        RAGKnowledgeWeight:  0.9,
        RAGIncludeSources:   true,
        
        Embedding: core.EmbeddingConfig{
            Provider:        "openai",
            Model:           "text-embedding-3-small",
            APIKey:          os.Getenv("OPENAI_API_KEY"),
            CacheEmbeddings: true,
            MaxBatchSize:    200, // Larger batches for efficiency
            TimeoutSeconds:  60,
        },
        
        Documents: core.DocumentConfig{
            AutoChunk:                true,
            SupportedTypes:           []string{"pdf", "txt", "md", "web", "code", "json"},
            MaxFileSize:              "50MB", // Larger files for knowledge base
            EnableMetadataExtraction: true,
            EnableURLScraping:        true,
        },
        
        Search: core.SearchConfigToml{
            HybridSearch:         true,
            KeywordWeight:        0.2, // Favor semantic search
            SemanticWeight:       0.8,
            EnableReranking:      false,
            EnableQueryExpansion: false,
        },
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    defer memory.Close()
    
    // Create production knowledge base
    pkb := NewProductionKnowledgeBase(memory, ProductionConfig{
        DefaultSearchLimit:     15,
        DefaultScoreThreshold:  0.7,
        CacheEnabled:          true,
        CacheTTL:              10 * time.Minute,
        MetricsEnabled:        true,
        MaxConcurrentSearches: 10,
    })
    
    // Create analytics engine
    analytics := NewKnowledgeBaseAnalytics(memory)
    
    // Create multi-modal search engine
    multiModal := NewMultiModalKnowledgeBase(memory)
    
    // Create contextual search engine
    contextual := NewContextualSearchEngine(memory)
    
    ctx := context.Background()
    
    // Populate knowledge base with sample documents
    err = populateKnowledgeBase(memory)
    if err != nil {
        log.Fatalf("Failed to populate knowledge base: %v", err)
    }
    
    // Demonstrate various search patterns
    fmt.Println("=== Production Knowledge Base Demo ===\n")
    
    // 1. Basic production search
    results, err := pkb.Search(ctx, "machine learning algorithms",
        core.WithLimit(10),
        core.WithTags([]string{"ml", "algorithms"}),
    )
    if err != nil {
        log.Printf("Production search failed: %v", err)
    } else {
        fmt.Printf("Production search found %d results\n", len(results))
    }
    
    // 2. Multi-modal search
    codeResults, err := multiModal.SearchCodeDocuments(ctx, "neural network implementation", "python")
    if err != nil {
        log.Printf("Code search failed: %v", err)
    } else {
        fmt.Printf("Code search found %d Python examples\n", len(codeResults))
    }
    
    // 3. Contextual search
    contextResults, err := contextual.SearchWithContext(ctx, "introduction to AI", []string{"ai", "introduction"}, "beginner")
    if err != nil {
        log.Printf("Contextual search failed: %v", err)
    } else {
        fmt.Printf("Contextual search found %d beginner-friendly results\n", len(contextResults))
    }
    
    // 4. Analytics
    distribution, err := analytics.AnalyzeContentDistribution(ctx)
    if err != nil {
        log.Printf("Analytics failed: %v", err)
    } else {
        fmt.Printf("Content distribution: %+v\n", distribution.ByDocumentType)
    }
    
    // 5. Performance metrics
    metrics := pkb.GetMetrics()
    fmt.Printf("Knowledge base metrics: %+v\n", metrics)
    
    // Start monitoring
    go monitorKnowledgeBase(pkb)
    
    // Keep running for demonstration
    time.Sleep(5 * time.Second)
}

func monitorKnowledgeBase(pkb *ProductionKnowledgeBase) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        metrics := pkb.GetMetrics()
        
        fmt.Printf("=== Knowledge Base Metrics ===\n")
        fmt.Printf("Total Searches: %d\n", metrics.TotalSearches)
        fmt.Printf("Average Response Time: %v\n", metrics.AverageResponseTime)

func populateKnowledgeBase(memory core.Memory) error {
    ctx := context.Background()
    
    // Sample knowledge base documents
    documents := []core.Document{
        {
            ID:      "ml-fundamentals",
            Title:   "Machine Learning Fundamentals",
            Content: `Machine learning is a subset of artificial intelligence that enables computers to learn and make decisions from data without being explicitly programmed. Key concepts include supervised learning, unsupervised learning, and reinforcement learning.

Supervised learning uses labeled data to train models that can make predictions on new, unseen data. Common algorithms include linear regression, decision trees, and neural networks.

Unsupervised learning finds patterns in data without labeled examples. Techniques include clustering, dimensionality reduction, and association rule learning.

Reinforcement learning involves agents learning to make decisions through interaction with an environment, receiving rewards or penalties for their actions.`,
            Source:  "ml-textbook.pdf",
            Type:    core.DocumentTypePDF,
            Metadata: map[string]any{
                "author":     "Dr. Sarah Johnson",
                "category":   "textbook",
                "difficulty": "beginner",
                "chapter":    1,
            },
            Tags:      []string{"ml", "fundamentals", "supervised", "unsupervised", "reinforcement"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "python-ml-code",
            Title:   "Python Machine Learning Implementation",
            Content: `# Machine Learning with Python and Scikit-learn

import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.linear_model import LinearRegression
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score, mean_squared_error

# Load and prepare data
def load_data(filename):
    """Load dataset from CSV file"""
    return pd.read_csv(filename)

def preprocess_data(data):
    """Clean and preprocess the dataset"""
    # Handle missing values
    data = data.dropna()
    
    # Feature scaling
    from sklearn.preprocessing import StandardScaler
    scaler = StandardScaler()
    
    return data, scaler

# Train models
def train_regression_model(X_train, y_train):
    """Train a linear regression model"""
    model = LinearRegression()
    model.fit(X_train, y_train)
    return model

def train_classification_model(X_train, y_train):
    """Train a random forest classifier"""
    model = RandomForestClassifier(n_estimators=100, random_state=42)
    model.fit(X_train, y_train)
    return model

# Example usage
if __name__ == "__main__":
    # Load data
    data = load_data("dataset.csv")
    processed_data, scaler = preprocess_data(data)
    
    # Split data
    X = processed_data.drop('target', axis=1)
    y = processed_data['target']
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)
    
    # Train and evaluate
    model = train_regression_model(X_train, y_train)
    predictions = model.predict(X_test)
    mse = mean_squared_error(y_test, predictions)
    
    print(f"Model MSE: {mse}")`,
            Source:  "ml_examples.py",
            Type:    core.DocumentTypeCode,
            Metadata: map[string]any{
                "programming_language": "python",
                "framework":           "scikit-learn",
                "difficulty":          "intermediate",
                "lines_of_code":       65,
            },
            Tags:      []string{"python", "scikit-learn", "code", "implementation", "ml"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "ai-research-trends",
            Title:   "Current Trends in AI Research",
            Content: `Artificial Intelligence research is rapidly evolving with several key trends shaping the field:

1. **Large Language Models (LLMs)**: Models like GPT, BERT, and T5 have revolutionized natural language processing with their ability to understand and generate human-like text.

2. **Multimodal AI**: Integration of text, image, audio, and video processing in single models, enabling more comprehensive understanding of complex data.

3. **Federated Learning**: Training models across distributed datasets without centralizing data, addressing privacy and security concerns.

4. **Explainable AI (XAI)**: Development of techniques to make AI decision-making processes more transparent and interpretable.

5. **AI Ethics and Fairness**: Increasing focus on developing AI systems that are fair, unbiased, and aligned with human values.

6. **Edge AI**: Deployment of AI models on edge devices for real-time processing with reduced latency and improved privacy.

7. **Neuromorphic Computing**: Hardware architectures inspired by biological neural networks for more efficient AI computation.

8. **AI for Science**: Application of AI to accelerate scientific discovery in fields like drug discovery, climate modeling, and materials science.

These trends are driving innovation across industries and opening new possibilities for AI applications in healthcare, finance, transportation, and beyond.`,
            Source:  "ai-research-2024.md",
            Type:    core.DocumentTypeMarkdown,
            Metadata: map[string]any{
                "category":     "research",
                "topic":        "ai-trends",
                "year":         2024,
                "content_type": "review",
                "difficulty":   "intermediate",
            },
            Tags:      []string{"ai", "research", "trends", "llm", "multimodal", "ethics"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "deep-learning-tutorial",
            Title:   "Deep Learning for Beginners",
            Content: `Deep learning is a subset of machine learning that uses artificial neural networks with multiple layers to model and understand complex patterns in data.

## What is Deep Learning?

Deep learning mimics the way the human brain processes information through interconnected nodes (neurons) organized in layers. The "deep" in deep learning refers to the multiple hidden layers between the input and output layers.

## Key Concepts:

**Neural Networks**: The foundation of deep learning, consisting of interconnected nodes that process and transmit information.

**Layers**: 
- Input Layer: Receives the raw data
- Hidden Layers: Process the data through weighted connections
- Output Layer: Produces the final result

**Activation Functions**: Mathematical functions that determine whether a neuron should be activated, introducing non-linearity to the network.

**Backpropagation**: The learning algorithm that adjusts weights based on the error between predicted and actual outputs.

## Common Applications:

- Image Recognition: Identifying objects, faces, and scenes in photos
- Natural Language Processing: Understanding and generating human language
- Speech Recognition: Converting spoken words to text
- Recommendation Systems: Suggesting products, movies, or content
- Autonomous Vehicles: Processing sensor data for navigation

## Getting Started:

1. Learn the fundamentals of linear algebra and calculus
2. Understand basic machine learning concepts
3. Practice with frameworks like TensorFlow or PyTorch
4. Start with simple projects like image classification
5. Gradually work on more complex problems

Deep learning has revolutionized AI and continues to drive breakthroughs in various fields.`,
            Source:  "deep-learning-guide.md",
            Type:    core.DocumentTypeMarkdown,
            Metadata: map[string]any{
                "category":       "tutorial",
                "topic":          "deep-learning",
                "difficulty":     "beginner",
                "reading_time":   "10 minutes",
                "target_audience": "beginners",
            },
            Tags:      []string{"deep-learning", "tutorial", "beginner", "neural-networks", "guide"},
            CreatedAt: time.Now(),
        } Large Language Models (LLMs): Models like GPT, BERT, and T5 have revolutionized natural language processing, enabling applications in text generation, translation, and question answering.

2. Multimodal AI: Integration of text, image, and audio processing in single models, enabling more comprehensive understanding and generation capabilities.

3. Federated Learning: Training models across distributed data sources while preserving privacy and reducing data centralization requirements.

4. Explainable AI (XAI): Development of methods to make AI decisions more interpretable and transparent, crucial for high-stakes applications.

5. Edge AI: Deployment of AI models on edge devices for real-time processing with reduced latency and improved privacy.

6. AI Ethics and Fairness: Increasing focus on developing fair, unbiased AI systems and addressing ethical concerns in AI deployment.

These trends are driving innovation across industries and creating new opportunities for AI applications in healthcare, finance, autonomous systems, and more. preserving privacy, crucial for healthcare and financial applications.

4. Explainable AI (XAI): Developing methods to make AI decisions more interpretable and transparent, essential for high-stakes applications.

5. AI Ethics and Fairness: Addressing bias, fairness, and ethical considerations in AI systems to ensure responsible deployment.

6. Edge AI: Optimizing AI models for deployment on resource-constrained devices, enabling real-time processing without cloud connectivity.

7. Reinforcement Learning from Human Feedback (RLHF): Improving AI systems by incorporating human preferences and feedback into the training process.

These trends are driving innovation across industries and creating new opportunities for AI applications.`,
            Source:  "ai-research-2024.md",
            Type:    core.DocumentTypeMarkdown,
            Metadata: map[string]any{
                "publication_year": 2024,
                "category":        "research",
                "topic":           "trends",
                "author":          "AI Research Lab",
            },
            Tags:      []string{"ai", "research", "trends", "llm", "multimodal", "ethics"},
            CreatedAt: time.Now(),
        },
    }
    
    // Ingest documents
    err := memory.IngestDocuments(ctx, documents)
    if err != nil {
        return fmt.Errorf("failed to ingest documents: %w", err)
    }
    
    fmt.Printf("Successfully populated knowledge base with %d documents\n", len(documents))
    return nil
}

func monitorKnowledgeBase(pkb *ProductionKnowledgeBase) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        metrics := pkb.GetMetrics()
        
        fmt.Printf("=== Knowledge Base Metrics ===\n")
        fmt.Printf("Total Searches: %d\n", metrics.TotalSearches)
        fmt.Printf("Average Response Time: %v\n", metrics.AverageResponseTime)
        fmt.Printf("Cache Hit Rate: %.2f%%\n", metrics.CacheHitRate*100)
        fmt.Printf("Popular Queries: %v\n", metrics.PopularQueries)
        fmt.Println()
    }
}
```

## Best Practices and Optimization

### 1. Search Optimization Strategies

```go
func optimizeKnowledgeBaseSearch(memory core.Memory) {
    ctx := context.Background()
    
    // Strategy 1: Use specific tags for better filtering
    results, err := memory.SearchKnowledge(ctx, "neural networks",
        core.WithLimit(10),
        core.WithTags([]string{"neural-networks", "deep-learning"}), // Specific tags
        core.WithScoreThreshold(0.8), // Higher threshold for quality
    )
    if err != nil {
        log.Printf("Tagged search failed: %v", err)
        return
    }
    
    // Strategy 2: Filter by document types for targeted search
    codeResults, err := memory.SearchKnowledge(ctx, "implementation examples",
        core.WithDocumentTypes([]core.DocumentType{core.DocumentTypeCode}),
        core.WithLimit(5),
    )
    if err != nil {
        log.Printf("Code-specific search failed: %v", err)
        return
    }
    
    // Strategy 3: Use source filtering for authoritative content
    authoritativeResults, err := memory.SearchKnowledge(ctx, "machine learning theory",
        core.WithSources([]string{"academic-papers", "textbooks", "research-journals"}),
        core.WithScoreThreshold(0.85), // High threshold for authoritative content
        core.WithLimit(8),
    )
    if err != nil {
        log.Printf("Authoritative search failed: %v", err)
        return
    }
    
    fmt.Printf("Optimized searches completed: %d general, %d code, %d authoritative\n",
        len(results), len(codeResults), len(authoritativeResults))
}
```

### 2. Error Handling and Fallbacks

```go
func robustKnowledgeBaseSearch(memory core.Memory, query string) ([]core.KnowledgeResult, error) {
    ctx := context.Background()
    
    // Primary search with strict criteria
    results, err := memory.SearchKnowledge(ctx, query,
        core.WithLimit(10),
        core.WithScoreThreshold(0.8),
    )
    
    if err != nil {
        log.Printf("Primary search failed: %v", err)
        return nil, err
    }
    
    // If no results, try with relaxed criteria
    if len(results) == 0 {
        log.Printf("No results with strict criteria, trying relaxed search")
        results, err = memory.SearchKnowledge(ctx, query,
            core.WithLimit(15),
            core.WithScoreThreshold(0.6), // Lower threshold
        )
        if err != nil {
            return nil, fmt.Errorf("relaxed search also failed: %w", err)
        }
    }
    
    // If still no results, try without score threshold
    if len(results) == 0 {
        log.Printf("No results with relaxed criteria, trying broad search")
        results, err = memory.SearchKnowledge(ctx, query,
            core.WithLimit(20),
            core.WithScoreThreshold(0.0), // No threshold
        )
        if err != nil {
            return nil, fmt.Errorf("broad search failed: %w", err)
        }
    }
    
    return results, nil
}
```

## Error Handling and Troubleshooting

### 1. Robust Search with Fallbacks

```go
func robustKnowledgeSearch(memory core.Memory, query string) ([]core.KnowledgeResult, error) {
    ctx := context.Background()
    
    // Try primary search with high quality threshold
    results, err := memory.SearchKnowledge(ctx, query,
        core.WithLimit(10),
        core.WithScoreThreshold(0.8),
    )
    if err != nil {
        return nil, fmt.Errorf("primary search failed: %w", err)
    }
    
    // If insufficient results, try with relaxed criteria
    if len(results) < 3 {
        log.Printf("Primary search returned %d results, trying relaxed search", len(results))
        results, err = memory.SearchKnowledge(ctx, query,
            core.WithLimit(15),
            core.WithScoreThreshold(0.6), // Lower threshold
        )
        if err != nil {
            return nil, fmt.Errorf("relaxed search failed: %w", err)
        }
    }
    
    // If still no results, try without score threshold
    if len(results) == 0 {
        log.Printf("No results with relaxed criteria, trying broad search")
        results, err = memory.SearchKnowledge(ctx, query,
            core.WithLimit(20),
            core.WithScoreThreshold(0.0), // No threshold
        )
        if err != nil {
            return nil, fmt.Errorf("broad search failed: %w", err)
        }
    }
    
    return results, nil
}

func handleSearchErrors(memory core.Memory, query string) ([]core.KnowledgeResult, error) {
    maxRetries := 3
    backoff := time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        results, err := memory.SearchKnowledge(context.Background(), query,
            core.WithLimit(10),
            core.WithScoreThreshold(0.7),
        )
        
        if err == nil {
            return results, nil
        }
        
        if attempt == maxRetries-1 {
            return nil, fmt.Errorf("search failed after %d attempts: %w", maxRetries, err)
        }
        
        log.Printf("Search attempt %d failed, retrying in %v: %v", attempt+1, backoff, err)
        time.Sleep(backoff)
        backoff *= 2 // Exponential backoff
    }
    
    return nil, fmt.Errorf("search failed after all retries")
}
```

### 2. Performance Monitoring and Optimization

```go
func monitorSearchPerformance(memory core.Memory) {
    queries := []string{
        "machine learning algorithms",
        "deep learning neural networks",
        "artificial intelligence applications",
        "data science techniques",
    }
    
    for _, query := range queries {
        start := time.Now()
        results, err := memory.SearchKnowledge(context.Background(), query,
            core.WithLimit(10),
            core.WithScoreThreshold(0.7),
        )
        duration := time.Since(start)
        
        if err != nil {
            log.Printf("Query failed: %s - %v", query, err)
            continue
        }
        
        log.Printf("Query: '%s' - %d results in %v", query, len(results), duration)
        
        if duration > 2*time.Second {
            log.Printf("WARNING: Slow query detected. Consider:")
            log.Printf("- Optimizing search parameters")
            log.Printf("- Checking database performance")
            log.Printf("- Implementing caching")
            log.Printf("- Reducing search scope")
        }
    }
}
```

## Conclusion

Knowledge bases in AgenticGoKit provide powerful capabilities for building intelligent information systems. Key takeaways:

- Use `SearchKnowledge` with `SearchOption` functions for precise control
- Implement multi-modal search patterns for different content types
- Use contextual search to adapt to user needs and expertise levels
- Implement production-ready features like caching and metrics
- Use analytics to understand and optimize knowledge base performance
- Apply proper error handling and fallback strategies

Advanced search patterns enable agents to find relevant information efficiently and provide accurate, contextual responses based on comprehensive knowledge bases.

## Next Steps

Perfect your enterprise knowledge system:

### ⚡ **Performance & Scale**
- **[Memory Optimization](memory-optimization.md)** - Advanced performance tuning and scaling
- Optimize your knowledge base for production workloads and high availability

::: tip Enterprise Ready
🏆 **Achievement Unlocked**: Enterprise Knowledge Base  
📈 **Scale**: Your system can now handle production workloads  
🔍 **Search**: Advanced patterns enable sophisticated information retrieval
:::

## Complete Learning Path

Review the full memory systems journey:

1. **[Basic Memory Operations](basic-memory.md)** - ✅ Foundation concepts
2. **[Vector Databases](vector-databases.md)** - ✅ Production storage  
3. **[Document Ingestion](document-ingestion.md)** - ✅ Content processing
4. **[RAG Implementation](rag-implementation.md)** - ✅ Intelligent retrieval
5. **[Knowledge Bases](knowledge-bases.md)** - ✅ **You are here**
6. **[Memory Optimization](memory-optimization.md)** - 🎯 **Final step**

## Advanced Topics

- **Multi-Tenant Knowledge Bases**: Isolate knowledge by user/organization
- **Knowledge Graph Integration**: Combine vector search with graph relationships  
- **Real-Time Updates**: Implement live knowledge base synchronization
- **Cross-Modal Search**: Search across text, images, and other media types

## Further Reading

- [Information Retrieval Best Practices](../../reference/best-practices/information-retrieval.md)
- [API Reference: SearchKnowledge](../../reference/api/memory.md#searchknowledge)
- [Examples: Knowledge Base Implementations](../../examples/knowledge-bases/)