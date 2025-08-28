---
title: RAG Implementation
description: Learn how to implement Retrieval-Augmented Generation (RAG) systems using AgenticGoKit's current RAG methods and APIs.
---

# RAG Implementation in AgenticGoKit

## Overview

Retrieval-Augmented Generation (RAG) combines the power of large language models with external knowledge retrieval to provide more accurate, up-to-date, and contextually relevant responses. This tutorial covers implementing RAG systems in AgenticGoKit using the current RAG-specific methods and APIs.

RAG enables agents to access vast amounts of information while maintaining the conversational abilities of language models, making them more knowledgeable and reliable.

## Prerequisites

- Understanding of [Vector Databases](vector-databases.md)
- Familiarity with [Document Ingestion](document-ingestion.md)
- Knowledge of language model APIs
- Basic understanding of information retrieval concepts

## RAG Architecture

### Current RAG Flow in AgenticGoKit

```
┌─────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ User Query  │───▶│ SearchKnowledge  │───▶│ Knowledge       │
└─────────────┘    │                  │    │ Results         │
                   └──────────────────┘    └─────────────────┘
                           │                         │
                           ▼                         ▼
                   ┌──────────────────┐    ┌─────────────────┐
                   │ SearchAll        │───▶│ Hybrid Results  │
                   │ (Hybrid Search)  │    │ (Personal +     │
                   └──────────────────┘    │ Knowledge)      │
                                          └─────────────────┘
                                                   │
                                                   ▼
                                          ┌─────────────────┐
                                          │ BuildContext    │
                                          │ (RAG Context)   │
                                          └─────────────────┘
                                                   │
                                                   ▼
                                          ┌─────────────────┐
                                          │ LLM Generation  │
                                          │ with Context    │
                                          └─────────────────┘
```

### Key RAG Components in Current API

1. **SearchKnowledge**: Search the knowledge base with advanced options
2. **SearchAll**: Hybrid search combining personal memory and knowledge
3. **BuildContext**: Assemble RAG context for LLM prompts
4. **IngestDocument/IngestDocuments**: Add knowledge to the system
5. **SearchOption/ContextOption**: Configure search and context building

## Basic RAG Implementation

### 1. Simple RAG Agent with Current API

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    "sync"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

type BasicRAGAgent struct {
    name   string
    memory core.Memory
    llm    core.LLMProvider
    config RAGConfig
}

type RAGConfig struct {
    MaxKnowledgeResults int
    ScoreThreshold      float32
    MaxContextTokens    int
    PersonalWeight      float32
    KnowledgeWeight     float32
}

func NewBasicRAGAgent(name string, memory core.Memory, llm core.LLMProvider) *BasicRAGAgent {
    return &BasicRAGAgent{
        name:   name,
        memory: memory,
        llm:    llm,
        config: RAGConfig{
            MaxKnowledgeResults: 5,
            ScoreThreshold:      0.7,
            MaxContextTokens:    3000,
            PersonalWeight:      0.3,
            KnowledgeWeight:     0.7,
        },
    }
}

func (r *BasicRAGAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Extract user query
    query, ok := state.Get("message")
    if !ok {
        return core.AgentResult{}, fmt.Errorf("no message in state")
    }
    queryStr := query.(string)
    
    // Set session context
    sessionID := event.GetSessionID()
    ctx = r.memory.SetSession(ctx, sessionID)
    
    // Store user message in chat history
    err := r.memory.AddMessage(ctx, "user", queryStr)
    if err != nil {
        log.Printf("Failed to store user message: %v", err)
    }
    
    // Build RAG context using current API
    ragContext, err := r.buildRAGContext(ctx, queryStr)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to build RAG context: %w", err)
    }
    
    // Generate response with RAG context
    response, err := r.generateResponse(ctx, queryStr, ragContext)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("generation failed: %w", err)
    }
    
    // Store assistant response
    err = r.memory.AddMessage(ctx, "assistant", response)
    if err != nil {
        log.Printf("Failed to store assistant response: %v", err)
    }
    
    // Return result
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("sources", ragContext.Sources)
    outputState.Set("context_used", len(ragContext.Knowledge) > 0 || len(ragContext.PersonalMemory) > 0)
    outputState.Set("token_count", ragContext.TokenCount)
    
    return core.AgentResult{OutputState: outputState}, nil
}

func (r *BasicRAGAgent) buildRAGContext(ctx context.Context, query string) (*core.RAGContext, error) {
    // Use current BuildContext API with options
    ragContext, err := r.memory.BuildContext(ctx, query,
        core.WithMaxTokens(r.config.MaxContextTokens),
        core.WithPersonalWeight(r.config.PersonalWeight),
        core.WithKnowledgeWeight(r.config.KnowledgeWeight),
        core.WithHistoryLimit(5), // Include recent chat history
        core.WithIncludeSources(true), // Include source attribution
    )
    if err != nil {
        return nil, fmt.Errorf("BuildContext failed: %w", err)
    }
    
    return ragContext, nil
}

func (r *BasicRAGAgent) generateResponse(ctx context.Context, query string, ragContext *core.RAGContext) (string, error) {
    // Use the pre-formatted context text from RAGContext
    prompt := fmt.Sprintf(`You are a helpful assistant. Use the provided context to answer the user's question accurately.

%s

Please provide a helpful and accurate response based on the context above.`, ragContext.ContextText)
    
    // Generate response using LLM
    response, err := r.llm.Generate(ctx, prompt)
    if err != nil {
        return "", fmt.Errorf("LLM generation failed: %w", err)
    }
    
    // Add source attribution if sources are available
    if len(ragContext.Sources) > 0 {
        response += fmt.Sprintf("\n\nSources: %s", strings.Join(ragContext.Sources, ", "))
    }
    
    return response, nil
}

func main() {
    // Create memory with RAG configuration
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "memory", // Use "pgvector" or "weaviate" for production
        Connection: "memory",
        MaxResults: 10,
        Dimensions: 1536,
        AutoEmbed:  true,
        
        // RAG-enhanced settings
        EnableRAG:               true,
        EnableKnowledgeBase:     true,
        KnowledgeMaxResults:     20,
        KnowledgeScoreThreshold: 0.7,
        ChunkSize:               1000,
        ChunkOverlap:            200,
        
        // RAG context assembly settings
        RAGMaxContextTokens: 4000,
        RAGPersonalWeight:   0.3,
        RAGKnowledgeWeight:  0.7,
        RAGIncludeSources:   true,
        
        Embedding: core.EmbeddingConfig{
            Provider: "dummy", // Use "openai" for production
            Model:    "dummy-model",
        },
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    defer memory.Close()
    
    // Create mock LLM for demonstration
    llm := &MockLLM{}
    
    // Create RAG agent
    ragAgent := NewBasicRAGAgent("rag-assistant", memory, llm)
    
    // Populate knowledge base with sample documents
    err = populateKnowledgeBase(memory)
    if err != nil {
        log.Fatalf("Failed to populate knowledge base: %v", err)
    }
    
    // Test the RAG agent
    ctx := context.Background()
    
    // Create test event and state
    event := core.NewEvent(
        "rag-assistant",
        core.EventData{"message": "What is AgenticGoKit and how does it support RAG?"},
        map[string]string{"session_id": "test-session"},
    )
    
    state := core.NewState()
    state.Set("message", "What is AgenticGoKit and how does it support RAG?")
    
    // Run the agent
    result, err := ragAgent.Run(ctx, event, state)
    if err != nil {
        log.Fatalf("RAG agent failed: %v", err)
    }
    
    // Display results
    response, _ := result.OutputState.Get("response")
    sources, _ := result.OutputState.Get("sources")
    tokenCount, _ := result.OutputState.Get("token_count")
    
    fmt.Printf("Response: %s\n", response)
    fmt.Printf("Sources: %v\n", sources)
    fmt.Printf("Token Count: %v\n", tokenCount)
}

// Mock LLM for demonstration
type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, prompt string) (string, error) {
    // Simple mock response based on prompt content
    if strings.Contains(strings.ToLower(prompt), "agenticgokit") {
        return "AgenticGoKit is a Go framework for building multi-agent systems with comprehensive RAG support, including knowledge base management, vector search, and context assembly for LLM integration.", nil
    }
    return "I can help answer questions based on the provided context.", nil
}

func populateKnowledgeBase(memory core.Memory) error {
    ctx := context.Background()
    
    // Sample documents for the knowledge base
    documents := []core.Document{
        {
            ID:      "agenticgokit-overview",
            Title:   "AgenticGoKit Framework Overview",
            Content: `AgenticGoKit is a comprehensive Go framework designed for building sophisticated multi-agent systems. The framework provides built-in support for Retrieval-Augmented Generation (RAG), enabling agents to access and reason over large knowledge bases.

Key RAG features include:
- Knowledge base management with document ingestion
- Vector-based semantic search using pgvector and Weaviate
- Hybrid search combining personal memory and knowledge base
- Context assembly for LLM prompts with source attribution
- Advanced search options with filtering and scoring

The framework supports multiple orchestration patterns and provides seamless integration with various LLM providers, making it ideal for building intelligent, knowledge-aware agents.`,
            Source:  "framework-documentation.md",
            Type:    core.DocumentTypeMarkdown,
            Metadata: map[string]any{
                "category": "framework",
                "topic":    "overview",
                "version":  "1.0",
            },
            Tags:      []string{"agenticgokit", "framework", "rag", "multi-agent"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "rag-implementation-guide",
            Title:   "RAG Implementation in AgenticGoKit",
            Content: `Implementing RAG in AgenticGoKit involves several key components:

1. Document Ingestion: Use IngestDocument() and IngestDocuments() to add knowledge to the system
2. Knowledge Search: Use SearchKnowledge() with SearchOption functions for targeted retrieval
3. Hybrid Search: Use SearchAll() to combine personal memory and knowledge base results
4. Context Building: Use BuildContext() with ContextOption functions to assemble RAG context
5. LLM Integration: Use the assembled context with your preferred LLM provider

The framework handles embedding generation, vector storage, and similarity search automatically, allowing developers to focus on building intelligent agent behaviors.

Advanced features include query enhancement, result reranking, and source attribution for transparent and reliable responses.`,
            Source:  "rag-guide.md",
            Type:    core.DocumentTypeMarkdown,
            Metadata: map[string]any{
                "category": "tutorial",
                "topic":    "rag",
                "difficulty": "intermediate",
            },
            Tags:      []string{"rag", "implementation", "tutorial", "search"},
            CreatedAt: time.Now(),
        },
        {
            ID:      "vector-databases-support",
            Title:   "Vector Database Support",
            Content: `AgenticGoKit supports multiple vector database backends for RAG implementations:

pgvector: PostgreSQL extension for vector similarity search
- Excellent for existing PostgreSQL infrastructure
- ACID compliance and mature ecosystem
- Support for HNSW and IVFFlat indexes
- Hybrid search capabilities

Weaviate: Cloud-native vector database
- Built-in vectorization and search capabilities
- Multi-modal support (text, images, etc.)
- GraphQL API and advanced filtering
- Horizontal scaling support

In-Memory: For development and testing
- Fast setup and testing
- No external dependencies
- Suitable for prototyping and small datasets

All backends support the same Memory interface, making it easy to switch between different storage solutions based on your requirements.`,
            Source:  "vector-databases.md",
            Type:    core.DocumentTypeMarkdown,
            Metadata: map[string]any{
                "category": "infrastructure",
                "topic":    "databases",
                "coverage": "comprehensive",
            },
            Tags:      []string{"vector-database", "pgvector", "weaviate", "storage"},
            CreatedAt: time.Now(),
        },
    }
    
    // Ingest documents into knowledge base
    err := memory.IngestDocuments(ctx, documents)
    if err != nil {
        return fmt.Errorf("failed to ingest documents: %w", err)
    }
    
    fmt.Printf("Successfully populated knowledge base with %d documents\n", len(documents))
    return nil
}
```

## Advanced RAG Techniques

### 1. Multi-Query RAG with SearchKnowledge

```go
type MultiQueryRAGAgent struct {
    memory core.Memory
    llm    core.LLMProvider
}

func NewMultiQueryRAGAgent(memory core.Memory, llm core.LLMProvider) *MultiQueryRAGAgent {
    return &MultiQueryRAGAgent{
        memory: memory,
        llm:    llm,
    }
}

func (m *MultiQueryRAGAgent) ProcessQuery(ctx context.Context, query string) (*core.RAGContext, error) {
    // Generate multiple query variations for better retrieval
    queryVariations, err := m.generateQueryVariations(ctx, query)
    if err != nil {
        log.Printf("Failed to generate query variations: %v", err)
        queryVariations = []string{query} // Fallback to original query
    }
    
    // Search with each query variation
    allResults := make([]core.KnowledgeResult, 0)
    for _, queryVar := range queryVariations {
        results, err := m.memory.SearchKnowledge(ctx, queryVar,
            core.WithLimit(5),
            core.WithScoreThreshold(0.6),
        )
        if err != nil {
            log.Printf("Search failed for query '%s': %v", queryVar, err)
            continue
        }
        allResults = append(allResults, results...)
    }
    
    // Deduplicate and rank results
    uniqueResults := m.deduplicateResults(allResults)
    
    // Build comprehensive RAG context
    ragContext, err := m.memory.BuildContext(ctx, query,
        core.WithMaxTokens(4000),
        core.WithKnowledgeWeight(0.8),
        core.WithPersonalWeight(0.2),
        core.WithIncludeSources(true),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to build context: %w", err)
    }
    
    return ragContext, nil
}

func (m *MultiQueryRAGAgent) generateQueryVariations(ctx context.Context, originalQuery string) ([]string, error) {
    prompt := fmt.Sprintf(`Given the query: "%s"

Generate 3 different ways to ask the same question that would help find relevant information. Make each variation focus on different aspects or use different terminology.

1.
2.
3.`, originalQuery)
    
    response, err := m.llm.Generate(ctx, prompt)
    if err != nil {
        return nil, err
    }
    
    // Parse the numbered list
    variations := []string{originalQuery} // Always include original
    lines := strings.Split(response, "\n")
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") || strings.HasPrefix(line, "3.") {
            variation := strings.TrimSpace(line[2:])
            if variation != "" && variation != originalQuery {
                variations = append(variations, variation)
            }
        }
    }
    
    return variations, nil
}

func (m *MultiQueryRAGAgent) deduplicateResults(results []core.KnowledgeResult) []core.KnowledgeResult {
    seen := make(map[string]bool)
    unique := make([]core.KnowledgeResult, 0)
    
    for _, result := range results {
        // Use document ID and chunk index as unique identifier
        key := fmt.Sprintf("%s-%d", result.DocumentID, result.ChunkIndex)
        if !seen[key] {
            seen[key] = true
            unique = append(unique, result)
        }
    }
    
    return unique
}
```

### 2. Hybrid RAG with Personal Memory Integration

```go
type HybridRAGAgent struct {
    memory core.Memory
    llm    core.LLMProvider
}

func NewHybridRAGAgent(memory core.Memory, llm core.LLMProvider) *HybridRAGAgent {
    return &HybridRAGAgent{
        memory: memory,
        llm:    llm,
    }
}

func (h *HybridRAGAgent) ProcessQuery(ctx context.Context, sessionID, query string) (string, error) {
    // Set session context
    ctx = h.memory.SetSession(ctx, sessionID)
    
    // Store user message
    err := h.memory.AddMessage(ctx, "user", query)
    if err != nil {
        log.Printf("Failed to store user message: %v", err)
    }
    
    // Perform hybrid search using SearchAll
    hybridResult, err := h.memory.SearchAll(ctx, query,
        core.WithLimit(10),
        core.WithIncludePersonal(true),
        core.WithIncludeKnowledge(true),
        core.WithScoreThreshold(0.5),
    )
    if err != nil {
        return "", fmt.Errorf("hybrid search failed: %w", err)
    }
    
    // Build comprehensive context using BuildContext
    ragContext, err := h.memory.BuildContext(ctx, query,
        core.WithMaxTokens(3500),
        core.WithPersonalWeight(0.4), // Higher weight for personal memory
        core.WithKnowledgeWeight(0.6),
        core.WithHistoryLimit(3),
        core.WithIncludeSources(true),
    )
    if err != nil {
        return "", fmt.Errorf("failed to build RAG context: %w", err)
    }
    
    // Generate personalized response
    response, err := h.generateHybridResponse(ctx, query, ragContext, hybridResult)
    if err != nil {
        return "", fmt.Errorf("response generation failed: %w", err)
    }
    
    // Store assistant response
    err = h.memory.AddMessage(ctx, "assistant", response)
    if err != nil {
        log.Printf("Failed to store assistant response: %v", err)
    }
    
    return response, nil
}

func (h *HybridRAGAgent) generateHybridResponse(ctx context.Context, query string, ragContext *core.RAGContext, hybridResult *core.HybridResult) (string, error) {
    // Create enhanced prompt with hybrid information
    prompt := fmt.Sprintf(`You are a personalized assistant with access to both general knowledge and personal information about the user.

Context Information:
%s

Hybrid Search Results:
- Personal Memory Results: %d
- Knowledge Base Results: %d
- Total Search Time: %v

Please provide a personalized response that:
1. Uses both general knowledge and personal context
2. Acknowledges the user's personal information when relevant
3. Provides accurate information from the knowledge base
4. Maintains a conversational and helpful tone

User Query: %s

Response:`, 
        ragContext.ContextText,
        len(hybridResult.PersonalMemory),
        len(hybridResult.Knowledge),
        hybridResult.SearchTime,
        query)
    
    response, err := h.llm.Generate(ctx, prompt)
    if err != nil {
        return "", err
    }
    
    // Add source attribution
    if len(ragContext.Sources) > 0 {
        response += fmt.Sprintf("\n\nSources: %s", strings.Join(ragContext.Sources, ", "))
    }
    
    return response, nil
}
```

### 3. Advanced Context Assembly with Custom Templates

```go
type AdvancedContextBuilder struct {
    memory core.Memory
}

func NewAdvancedContextBuilder(memory core.Memory) *AdvancedContextBuilder {
    return &AdvancedContextBuilder{memory: memory}
}

func (a *AdvancedContextBuilder) BuildCustomContext(ctx context.Context, query string, template string) (*core.RAGContext, error) {
    // Use BuildContext with custom template
    ragContext, err := a.memory.BuildContext(ctx, query,
        core.WithMaxTokens(4000),
        core.WithPersonalWeight(0.3),
        core.WithKnowledgeWeight(0.7),
        core.WithHistoryLimit(5),
        core.WithIncludeSources(true),
        core.WithFormatTemplate(template),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to build custom context: %w", err)
    }
    
    return ragContext, nil
}

// Example usage with different templates
func (a *AdvancedContextBuilder) BuildTechnicalContext(ctx context.Context, query string) (*core.RAGContext, error) {
    template := `TECHNICAL DOCUMENTATION CONTEXT

Query: {{.Query}}

Knowledge Base Information:
{{range .Knowledge}}
📄 {{.Title}} ({{.Source}})
   {{.Content}}
   Score: {{.Score}}

{{end}}

Personal Notes:
{{range .PersonalMemory}}
📝 {{.Content}} ({{.CreatedAt.Format "2006-01-02"}})

{{end}}

Recent Conversation:
{{range .ChatHistory}}
{{.Role | title}}: {{.Content}}
{{end}}

Please provide a technical response based on the above information.`
    
    return a.BuildCustomContext(ctx, query, template)
}

func (a *AdvancedContextBuilder) BuildConversationalContext(ctx context.Context, query string) (*core.RAGContext, error) {
    template := `Hey! Here's what I found that might help answer your question:

Your question: {{.Query}}

From my knowledge base:
{{range .Knowledge}}
• {{.Content}} (from {{.Source}})
{{end}}

From our previous conversations:
{{range .PersonalMemory}}
• {{.Content}}
{{end}}

Recent chat:
{{range .ChatHistory}}
{{if eq .Role "user"}}You{{else}}Me{{end}}: {{.Content}}
{{end}}

Let me help you with this!`
    
    return a.BuildCustomContext(ctx, query, template)
}
```

## Production RAG Implementation

### 1. Enterprise RAG Agent

```go
type EnterpriseRAGAgent struct {
    memory      core.Memory
    llm         core.LLMProvider
    config      EnterpriseRAGConfig
    metrics     *RAGMetrics
    cache       *RAGCache
}

type EnterpriseRAGConfig struct {
    MaxKnowledgeResults     int
    ScoreThreshold          float32
    MaxContextTokens        int
    PersonalWeight          float32
    KnowledgeWeight         float32
    EnableCaching           bool
    EnableMetrics           bool
    EnableSourceValidation  bool
    ResponseMaxLength       int
}

type RAGMetrics struct {
    QueriesProcessed    int64         `json:"queries_processed"`
    AverageResponseTime time.Duration `json:"average_response_time"`
    CacheHitRate        float64       `json:"cache_hit_rate"`
    SourcesUsedAvg      float64       `json:"sources_used_average"`
    mu                  sync.RWMutex
}

type RAGCache struct {
    contextCache map[string]*core.RAGContext
    mu           sync.RWMutex
    ttl          time.Duration
    timestamps   map[string]time.Time
}

func NewEnterpriseRAGAgent(memory core.Memory, llm core.LLMProvider, config EnterpriseRAGConfig) *EnterpriseRAGAgent {
    agent := &EnterpriseRAGAgent{
        memory:  memory,
        llm:     llm,
        config:  config,
        metrics: &RAGMetrics{},
    }
    
    if config.EnableCaching {
        agent.cache = &RAGCache{
            contextCache: make(map[string]*core.RAGContext),
            timestamps:   make(map[string]time.Time),
            ttl:          5 * time.Minute,
        }
        go agent.cache.cleanup()
    }
    
    return agent
}

func (e *EnterpriseRAGAgent) ProcessQuery(ctx context.Context, sessionID, query string) (string, *RAGMetrics, error) {
    start := time.Now()
    
    // Set session context
    ctx = e.memory.SetSession(ctx, sessionID)
    
    // Check cache if enabled
    var ragContext *core.RAGContext
    var err error
    cacheHit := false
    
    if e.config.EnableCaching {
        if cached := e.cache.Get(query); cached != nil {
            ragContext = cached
            cacheHit = true
        }
    }
    
    // Build context if not cached
    if ragContext == nil {
        ragContext, err = e.memory.BuildContext(ctx, query,
            core.WithMaxTokens(e.config.MaxContextTokens),
            core.WithPersonalWeight(e.config.PersonalWeight),
            core.WithKnowledgeWeight(e.config.KnowledgeWeight),
            core.WithHistoryLimit(5),
            core.WithIncludeSources(true),
        )
        if err != nil {
            return "", nil, fmt.Errorf("failed to build RAG context: %w", err)
        }
        
        // Cache the result
        if e.config.EnableCaching {
            e.cache.Set(query, ragContext)
        }
    }
    
    // Validate sources if enabled
    if e.config.EnableSourceValidation {
        ragContext = e.validateSources(ragContext)
    }
    
    // Generate response
    response, err := e.generateEnterpriseResponse(ctx, query, ragContext)
    if err != nil {
        return "", nil, fmt.Errorf("response generation failed: %w", err)
    }
    
    // Update metrics
    if e.config.EnableMetrics {
        e.updateMetrics(start, cacheHit, len(ragContext.Sources))
    }
    
    // Store interaction
    e.memory.AddMessage(ctx, "user", query)
    e.memory.AddMessage(ctx, "assistant", response)
    
    return response, e.getMetrics(), nil
}

func (e *EnterpriseRAGAgent) generateEnterpriseResponse(ctx context.Context, query string, ragContext *core.RAGContext) (string, error) {
    prompt := fmt.Sprintf(`You are an enterprise AI assistant. Provide accurate, professional responses based on the provided context.

Context Information:
%s

Guidelines:
- Be accurate and cite sources when possible
- Maintain a professional tone
- If information is insufficient, clearly state limitations
- Provide actionable insights when appropriate

User Query: %s

Response:`, ragContext.ContextText, query)
    
    response, err := e.llm.Generate(ctx, prompt)
    if err != nil {
        return "", err
    }
    
    // Truncate if too long
    if len(response) > e.config.ResponseMaxLength {
        response = response[:e.config.ResponseMaxLength] + "..."
    }
    
    // Add source attribution
    if len(ragContext.Sources) > 0 {
        response += fmt.Sprintf("\n\nSources: %s", strings.Join(ragContext.Sources, ", "))
    }
    
    return response, nil
}

func (e *EnterpriseRAGAgent) validateSources(ragContext *core.RAGContext) *core.RAGContext {
    // Simple source validation - in production, implement more sophisticated validation
    validSources := make([]string, 0, len(ragContext.Sources))
    for _, source := range ragContext.Sources {
        if source != "" && len(source) > 3 { // Basic validation
            validSources = append(validSources, source)
        }
    }
    ragContext.Sources = validSources
    return ragContext
}

func (e *EnterpriseRAGAgent) updateMetrics(start time.Time, cacheHit bool, sourcesUsed int) {
    e.metrics.mu.Lock()
    defer e.metrics.mu.Unlock()
    
    e.metrics.QueriesProcessed++
    
    // Update average response time
    responseTime := time.Since(start)
    if e.metrics.AverageResponseTime == 0 {
        e.metrics.AverageResponseTime = responseTime
    } else {
        e.metrics.AverageResponseTime = (e.metrics.AverageResponseTime + responseTime) / 2
    }
    
    // Update cache hit rate
    if cacheHit {
        e.metrics.CacheHitRate = (e.metrics.CacheHitRate*float64(e.metrics.QueriesProcessed-1) + 1.0) / float64(e.metrics.QueriesProcessed)
    } else {
        e.metrics.CacheHitRate = (e.metrics.CacheHitRate * float64(e.metrics.QueriesProcessed-1)) / float64(e.metrics.QueriesProcessed)
    }
    
    // Update average sources used
    e.metrics.SourcesUsedAvg = (e.metrics.SourcesUsedAvg*float64(e.metrics.QueriesProcessed-1) + float64(sourcesUsed)) / float64(e.metrics.QueriesProcessed)
}

func (e *EnterpriseRAGAgent) getMetrics() *RAGMetrics {
    e.metrics.mu.RLock()
    defer e.metrics.mu.RUnlock()
    
    // Return a copy of metrics
    return &RAGMetrics{
        QueriesProcessed:    e.metrics.QueriesProcessed,
        AverageResponseTime: e.metrics.AverageResponseTime,
        CacheHitRate:        e.metrics.CacheHitRate,
        SourcesUsedAvg:      e.metrics.SourcesUsedAvg,
    }
}

// Cache implementation
func (c *RAGCache) Get(key string) *core.RAGContext {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if context, exists := c.contextCache[key]; exists {
        if timestamp, ok := c.timestamps[key]; ok {
            if time.Since(timestamp) < c.ttl {
                return context
            }
        }
    }
    
    return nil
}

func (c *RAGCache) Set(key string, context *core.RAGContext) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.contextCache[key] = context
    c.timestamps[key] = time.Now()
}

func (c *RAGCache) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for key, timestamp := range c.timestamps {
            if now.Sub(timestamp) > c.ttl {
                delete(c.contextCache, key)
                delete(c.timestamps, key)
            }
        }
        c.mu.Unlock()
    }
}
```

## RAG Performance Optimization

### 1. Query Optimization Strategies

```go
type OptimizedRAGAgent struct {
    memory core.Memory
    llm    core.LLMProvider
}

func (o *OptimizedRAGAgent) OptimizedQuery(ctx context.Context, query string) (*core.RAGContext, error) {
    // Strategy 1: Use targeted search with filters
    results, err := o.memory.SearchKnowledge(ctx, query,
        core.WithLimit(8), // Reasonable limit
        core.WithScoreThreshold(0.75), // Higher threshold for quality
        // Add filters based on query analysis
    )
    if err != nil {
        return nil, fmt.Errorf("knowledge search failed: %w", err)
    }
    
    // Strategy 2: If few results, expand search
    if len(results) < 3 {
        expandedResults, err := o.memory.SearchKnowledge(ctx, query,
            core.WithLimit(10),
            core.WithScoreThreshold(0.6), // Lower threshold
        )
        if err == nil && len(expandedResults) > len(results) {
            results = expandedResults
        }
    }
    
    // Strategy 3: Build optimized context
    ragContext, err := o.memory.BuildContext(ctx, query,
        core.WithMaxTokens(3500), // Optimal context size
        core.WithKnowledgeWeight(0.8), // Focus on knowledge
        core.WithPersonalWeight(0.2),
        core.WithHistoryLimit(3), // Limited history for focus
        core.WithIncludeSources(true),
    )
    if err != nil {
        return nil, fmt.Errorf("context building failed: %w", err)
    }
    
    return ragContext, nil
}

func (o *OptimizedRAGAgent) AnalyzeQuery(query string) map[string]any {
    analysis := make(map[string]any)
    
    // Simple query analysis
    words := strings.Fields(strings.ToLower(query))
    analysis["word_count"] = len(words)
    analysis["has_question_words"] = containsAny(words, []string{"what", "how", "why", "when", "where", "who"})
    analysis["is_technical"] = containsAny(words, []string{"api", "code", "function", "implementation", "algorithm"})
    analysis["is_conceptual"] = containsAny(words, []string{"concept", "theory", "principle", "idea", "philosophy"})
    
    return analysis
}

func containsAny(slice []string, items []string) bool {
    for _, item := range items {
        for _, s := range slice {
            if s == item {
                return true
            }
        }
    }
    return false
}f timestamp, timestampExists := c.timestamps[key]; timestampExists {
            if time.Since(timestamp) < c.ttl {
                return context
            }
            // Expired, remove from cache
            delete(c.contextCache, key)
            delete(c.timestamps, key)
        }
    }
    
    return nil
}

func (c *RAGCache) Set(key string, context *core.RAGContext) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.contextCache[key] = context
    c.timestamps[key] = time.Now()
}

func (c *RAGCache) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mu.Lock()
        now := time.Now() ok := c.timestamps[key]; ok {
            if time.Since(timestamp) < c.ttl {
                return context
            }
        }
    }
    return nil
}

func (c *RAGCache) Set(key string, context *core.RAGContext) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.contextCache[key] = context
    c.timestamps[key] = time.Now()
}

func (c *RAGCache) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for key, timestamp := range c.timestamps {
            if now.Sub(timestamp) > c.ttl {
                delete(c.contextCache, key)
                delete(c.timestamps, key)
            }
        }
        c.mu.Unlock()
    }
}
```

### 2. RAG Performance Monitoring

```go
func monitorRAGPerformance(agent *EnterpriseRAGAgent) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        metrics := agent.getMetrics()
        
        fmt.Printf("=== RAG Performance Metrics ===\n")
        fmt.Printf("Queries Processed: %d\n", metrics.QueriesProcessed)
        fmt.Printf("Average Response Time: %v\n", metrics.AverageResponseTime)
        fmt.Printf("Cache Hit Rate: %.2f%%\n", metrics.CacheHitRate*100)
        fmt.Printf("Average Sources Used: %.2f\n", metrics.SourcesUsedAvg)
        fmt.Println()
    }
}
```

## Best Practices and Optimization

### 1. Query Optimization

```go
func optimizeRAGQueries(memory core.Memory) {
    ctx := context.Background()
    
    // Use specific search options for better results
    results, err := memory.SearchKnowledge(ctx, "machine learning algorithms",
        core.WithLimit(10),                    // Reasonable limit
        core.WithScoreThreshold(0.75),         // Higher threshold for quality
        core.WithTags([]string{"ml", "algorithms"}), // Filter by relevant tags
        core.WithDocumentTypes([]core.DocumentType{
            core.DocumentTypePDF,
            core.DocumentTypeMarkdown,
        }), // Filter by document types
    )
    if err != nil {
        log.Printf("Search failed: %v", err)
        return
    }
    
    // Build optimized context
    ragContext, err := memory.BuildContext(ctx, "machine learning algorithms",
        core.WithMaxTokens(3000),      // Appropriate context size
        core.WithPersonalWeight(0.2),  // Focus more on knowledge
        core.WithKnowledgeWeight(0.8),
        core.WithHistoryLimit(3),      // Limited history for focus
        core.WithIncludeSources(true), // Always include sources
    )
    if err != nil {
        log.Printf("Context building failed: %v", err)
        return
    }
    
    fmt.Printf("Optimized RAG context: %d tokens, %d sources\n", 
        ragContext.TokenCount, len(ragContext.Sources))
}
```

### 2. Error Handling and Fallbacks

```go
type RobustRAGAgent struct {
    memory      core.Memory
    llm         core.LLMProvider
    fallbackLLM core.LLMProvider // Backup LLM provider
}

func (r *RobustRAGAgent) ProcessQueryWithFallbacks(ctx context.Context, query string) (string, error) {
    // Try to build RAG context
    ragContext, err := r.memory.BuildContext(ctx, query,
        core.WithMaxTokens(3000),
        core.WithPersonalWeight(0.3),
        core.WithKnowledgeWeight(0.7),
        core.WithIncludeSources(true),
    )
    
    var response string
    
    if err != nil {
        log.Printf("RAG context building failed: %v", err)
        // Fallback to simple query without RAG
        response, err = r.llm.Generate(ctx, fmt.Sprintf("Please answer: %s", query))
        if err != nil && r.fallbackLLM != nil {
            // Try fallback LLM
            response, err = r.fallbackLLM.Generate(ctx, fmt.Sprintf("Please answer: %s", query))
        }
    } else {
        // Use RAG context
        prompt := fmt.Sprintf("Based on the following context:\n%s\n\nPlease answer: %s", 
            ragContext.ContextText, query)
        
        response, err = r.llm.Generate(ctx, prompt)
        if err != nil && r.fallbackLLM != nil {
            // Try fallback LLM with context
            response, err = r.fallbackLLM.Generate(ctx, prompt)
        }
    }
    
    if err != nil {
        return "I apologize, but I'm unable to process your request at the moment. Please try again later.", nil
    }
    
    return response, nil
}
```

## Conclusion

RAG implementation in AgenticGoKit leverages the current API to provide powerful retrieval-augmented generation capabilities. Key takeaways:

- Use `SearchKnowledge` for targeted knowledge base searches
- Use `SearchAll` for hybrid searches combining personal and knowledge data
- Use `BuildContext` for comprehensive RAG context assembly
- Leverage `SearchOption` and `ContextOption` functions for fine-tuned control
- Implement caching and metrics for production deployments
- Use proper error handling and fallback strategies

The current RAG API provides a robust foundation for building intelligent, knowledge-aware agents that can access and reason over large amounts of information while maintaining conversational capabilities.

## Next Steps

Take your RAG implementation to the next level:

### 🏗️ **Enterprise Knowledge**
- **[Knowledge Bases](knowledge-bases.md)** - Advanced knowledge management patterns
- Scale your RAG system with sophisticated search and filtering capabilities

### ⚡ **Performance & Scale**
- **[Memory Optimization](memory-optimization.md)** - Performance tuning and scaling
- Optimize RAG performance with caching, monitoring, and resource management

::: success RAG Implementation Complete
🎉 **Congratulations!** You now have a working RAG system  
🚀 **Next Level**: Scale with advanced knowledge base patterns  
📊 **Monitor**: Implement performance optimization for production
:::

## Foundation Requirements

Ensure these components are properly configured:

- **[Vector Databases](vector-databases.md)** - Production storage backends
- **[Document Ingestion](document-ingestion.md)** - Quality content pipeline
- **[Basic Memory Operations](basic-memory.md)** - Core memory understanding

## Advanced Patterns

- **Multi-Agent RAG**: Coordinate RAG across multiple agents
- **Streaming RAG**: Implement real-time response streaming
- **Hybrid Search**: Combine multiple retrieval strategies

## Further Reading

- [RAG Best Practices](../../reference/best-practices/rag.md)
- [API Reference: RAG Methods](../../reference/api/memory.md#rag-methods)
- [Examples: RAG Implementations](../../examples/rag/)