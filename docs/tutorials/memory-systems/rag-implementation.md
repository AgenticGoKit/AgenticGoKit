# RAG Implementation in AgenticGoKit

## Overview

Retrieval-Augmented Generation (RAG) combines the power of large language models with external knowledge retrieval to provide more accurate, up-to-date, and contextually relevant responses. This tutorial covers implementing RAG systems in AgenticGoKit, from basic retrieval to advanced techniques.

RAG enables agents to access vast amounts of information while maintaining the conversational abilities of language models, making them more knowledgeable and reliable.

## Prerequisites

- Understanding of [Vector Databases](vector-databases.md)
- Familiarity with [Basic Memory Operations](basic-memory.md)
- Knowledge of language model APIs
- Basic understanding of information retrieval concepts

## RAG Architecture

### Basic RAG Flow

```
┌─────────────┐    ┌──────────────┐    ┌─────────────────┐
│ User Query  │───▶│   Retrieval  │───▶│ Context + Query │
└─────────────┘    │   System     │    └─────────────────┘
                   └──────────────┘             │
                           │                    ▼
                           ▼            ┌─────────────────┐
                   ┌──────────────┐    │ Language Model  │
                   │ Vector Store │    │   Generation    │
                   └──────────────┘    └─────────────────┘
                                               │
                                               ▼
                                       ┌─────────────────┐
                                       │ Enhanced        │
                                       │ Response        │
                                       └─────────────────┘
```

### Advanced RAG Components

1. **Query Processing**: Understanding and reformulating user queries
2. **Retrieval**: Finding relevant information from knowledge base
3. **Reranking**: Improving relevance of retrieved results
4. **Context Building**: Constructing effective prompts
5. **Generation**: Producing responses with retrieved context
6. **Response Enhancement**: Post-processing and validation

## Basic RAG Implementation

### 1. Simple RAG Agent

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

type BasicRAGAgent struct {
    name   string
    memory core.Memory
    llm    core.ModelProvider
    config RAGConfig
}

type RAGConfig struct {
    MaxRetrievalResults int
    ScoreThreshold      float32
    MaxContextLength    int
    ContextTemplate     string
}

func NewBasicRAGAgent(name string, memory core.Memory, llm core.ModelProvider) *BasicRAGAgent {
    return &BasicRAGAgent{
        name:   name,
        memory: memory,
        llm:    llm,
        config: RAGConfig{
            MaxRetrievalResults: 5,
            ScoreThreshold:      0.7,
            MaxContextLength:    2000,
            ContextTemplate: `Based on the following information:

%s

Please answer the question: %s`,
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
    
    // Retrieve relevant context
    contextStr, sources, err := r.retrieveContext(ctx, queryStr)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("retrieval failed: %w", err)
    }
    
    // Generate response with context
    response, err := r.generateResponse(ctx, queryStr, contextStr)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("generation failed: %w", err)
    }
    
    // Store the interaction
    sessionID, _ := event.GetMetadataValue(core.SessionIDKey)
    err = r.storeInteraction(ctx, sessionID, queryStr, response, sources)
    if err != nil {
        log.Printf("Failed to store interaction: %v", err)
    }
    
    // Return result
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("sources", sources)
    outputState.Set("context_used", len(sources) > 0)
    
    return core.AgentResult{OutputState: outputState}, nil
}

func (r *BasicRAGAgent) retrieveContext(ctx context.Context, query string) (string, []string, error) {
    // Search for relevant information
    results, err := r.memory.Search(ctx, query,
        core.WithLimit(r.config.MaxRetrievalResults),
        core.WithScoreThreshold(r.config.ScoreThreshold),
    )
    if err != nil {
        return "", nil, fmt.Errorf("search failed: %w", err)
    }
    
    if len(results) == 0 {
        return "", nil, nil // No relevant context found
    }
    
    // Build context string
    var contextBuilder strings.Builder
    sources := make([]string, 0, len(results))
    
    for i, result := range results {
        // Add numbered context item
        contextBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Content))
        
        // Track sources
        if source, ok := result.Metadata["source"]; ok {
            sources = append(sources, source)
        } else {
            sources = append(sources, fmt.Sprintf("Document %s", result.ID))
        }
    }
    
    context := contextBuilder.String()
    
    // Truncate if too long
    if len(context) > r.config.MaxContextLength {
        context = context[:r.config.MaxContextLength] + "..."
    }
    
    return context, sources, nil
}

func (r *BasicRAGAgent) generateResponse(ctx context.Context, query, context string) (string, error) {
    var prompt string
    
    if context != "" {
        // Use context template
        prompt = fmt.Sprintf(r.config.ContextTemplate, context, query)
    } else {
        // Fallback to direct query
        prompt = fmt.Sprintf("Please answer the following question: %s", query)
    }
    
    // Generate response
    response, err := r.llm.Generate(ctx, prompt)
    if err != nil {
        return "", fmt.Errorf("LLM generation failed: %w", err)
    }
    
    return response, nil
}

func (r *BasicRAGAgent) storeInteraction(ctx context.Context, sessionID, query, response string, sources []string) error {
    // Store user query
    err := r.memory.Store(ctx, query, "user-query",
        core.WithSession(sessionID),
        core.WithTimestamp(time.Now()),
        core.WithMetadata(map[string]string{
            "interaction_type": "rag-query",
        }),
    )
    if err != nil {
        return err
    }
    
    // Store agent response with sources
    sourcesStr := strings.Join(sources, ", ")
    err = r.memory.Store(ctx, response, "agent-response",
        core.WithSession(sessionID),
        core.WithTimestamp(time.Now()),
        core.WithMetadata(map[string]string{
            "interaction_type": "rag-response",
            "sources_used":     sourcesStr,
            "sources_count":    fmt.Sprintf("%d", len(sources)),
        }),
    )
    
    return err
}

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
    
    // Setup LLM provider
    llm, err := core.NewOpenAIAdapter(
        os.Getenv("OPENAI_API_KEY"),
        "gpt-3.5-turbo",
        2000,
        0.7,
    )
    if err != nil {
        log.Fatalf("Failed to create LLM: %v", err)
    }
    
    // Create RAG agent
    ragAgent := NewBasicRAGAgent("rag-assistant", memory, llm)
    
    // Test the agent
    ctx := context.Background()
    
    // First, populate some knowledge
    knowledge := []string{
        "AgenticGoKit is a Go framework for building multi-agent systems with support for LLM integration, memory systems, and orchestration patterns.",
        "Vector databases like pgvector and Weaviate are used in AgenticGoKit for semantic search and RAG implementations.",
        "The framework supports multiple orchestration patterns including route, collaborative, sequential, and loop modes.",
    }
    
    for _, info := range knowledge {
        memory.Store(ctx, info, "knowledge",
            core.WithMetadata(map[string]string{
                "source": "documentation",
                "topic":  "agenticgokit",
            }),
        )
    }
    
    // Test query
    event := core.NewEvent(
        "rag-assistant",
        core.EventData{"message": "What is AgenticGoKit and what databases does it support?"},
        map[string]string{"session_id": "test-session"},
    )
    
    state := core.NewState()
    state.Set("message", "What is AgenticGoKit and what databases does it support?")
    
    result, err := ragAgent.Run(ctx, event, state)
    if err != nil {
        log.Fatalf("RAG agent failed: %v", err)
    }
    
    response, _ := result.OutputState.Get("response")
    sources, _ := result.OutputState.Get("sources")
    
    fmt.Printf("Response: %s\n", response)
    fmt.Printf("Sources: %v\n", sources)
}
```

## Advanced RAG Techniques

### 1. Query Enhancement

```go
type QueryEnhancer struct {
    llm core.LLMProvider
}

func NewQueryEnhancer(llm core.LLMProvider) *QueryEnhancer {
    return &QueryEnhancer{llm: llm}
}

func (qe *QueryEnhancer) EnhanceQuery(ctx context.Context, originalQuery string, conversationHistory []core.Message) (string, error) {
    // Build context from conversation history
    var historyBuilder strings.Builder
    for _, msg := range conversationHistory {
        historyBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
    }
    
    // Create enhancement prompt
    prompt := fmt.Sprintf(`Given the conversation history:
%s

The user's current query is: "%s"

Please rewrite this query to be more specific and searchable, incorporating relevant context from the conversation history. The enhanced query should be optimized for semantic search.

Enhanced query:`, historyBuilder.String(), originalQuery)
    
    enhancedQuery, err := qe.llm.Generate(ctx, prompt)
    if err != nil {
        // Fallback to original query
        return originalQuery, nil
    }
    
    return strings.TrimSpace(enhancedQuery), nil
}

// Multi-query generation for better retrieval
func (qe *QueryEnhancer) GenerateMultipleQueries(ctx context.Context, originalQuery string) ([]string, error) {
    prompt := fmt.Sprintf(`Given the query: "%s"

Generate 3 different ways to ask the same question that would help find relevant information:

1.
2.
3.`, originalQuery)
    
    response, err := qe.llm.Generate(ctx, prompt)
    if err != nil {
        return []string{originalQuery}, nil
    }
    
    // Parse the numbered list
    lines := strings.Split(response, "\n")
    queries := make([]string, 0, 3)
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") || strings.HasPrefix(line, "3.") {
            query := strings.TrimSpace(line[2:])
            if query != "" {
                queries = append(queries, query)
            }
        }
    }
    
    // Always include original query
    if len(queries) == 0 {
        queries = append(queries, originalQuery)
    }
    
    return queries, nil
}
```

### 2. Advanced Retrieval with Reranking

```go
type AdvancedRetriever struct {
    memory   core.Memory
    reranker *Reranker
    config   RetrievalConfig
}

type RetrievalConfig struct {
    InitialRetrievalLimit int
    FinalResultLimit      int
    ScoreThreshold        float64
    RerankingEnabled      bool
    DiversityThreshold    float64
}

type Reranker struct {
    llm core.LLMProvider
}

func NewReranker(llm core.LLMProvider) *Reranker {
    return &Reranker{llm: llm}
}

func (r *Reranker) Rerank(ctx context.Context, query string, results []core.MemoryResult) ([]core.MemoryResult, error) {
    if len(results) <= 1 {
        return results, nil
    }
    
    // Create reranking prompt
    var resultsBuilder strings.Builder
    resultsBuilder.WriteString("Rank the following passages by relevance to the query:\n\n")
    resultsBuilder.WriteString(fmt.Sprintf("Query: %s\n\n", query))
    
    for i, result := range results {
        resultsBuilder.WriteString(fmt.Sprintf("Passage %d: %s\n\n", i+1, result.Content))
    }
    
    resultsBuilder.WriteString("Please rank these passages from most relevant (1) to least relevant, providing only the numbers separated by commas (e.g., 3,1,4,2):")
    
    response, err := r.llm.Generate(ctx, resultsBuilder.String())
    if err != nil {
        // Fallback to original order
        return results, nil
    }
    
    // Parse ranking
    ranking := r.parseRanking(response, len(results))
    
    // Reorder results based on ranking
    rerankedResults := make([]core.MemoryResult, 0, len(results))
    for _, idx := range ranking {
        if idx >= 0 && idx < len(results) {
            rerankedResults = append(rerankedResults, results[idx])
        }
    }
    
    // Add any missing results
    used := make(map[int]bool)
    for _, idx := range ranking {
        used[idx] = true
    }
    
    for i, result := range results {
        if !used[i] {
            rerankedResults = append(rerankedResults, result)
        }
    }
    
    return rerankedResults, nil
}

func (r *Reranker) parseRanking(response string, maxItems int) []int {
    // Clean and split the response
    response = strings.TrimSpace(response)
    parts := strings.Split(response, ",")
    
    ranking := make([]int, 0, len(parts))
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if num, err := strconv.Atoi(part); err == nil && num >= 1 && num <= maxItems {
            ranking = append(ranking, num-1) // Convert to 0-based index
        }
    }
    
    return ranking
}

func (ar *AdvancedRetriever) Retrieve(ctx context.Context, query string) ([]core.MemoryResult, error) {
    // Initial retrieval with higher limit
    results, err := ar.memory.Search(ctx, query,
        core.WithLimit(ar.config.InitialRetrievalLimit),
        core.WithScoreThreshold(ar.config.ScoreThreshold*0.8), // Lower threshold initially
    )
    if err != nil {
        return nil, fmt.Errorf("initial retrieval failed: %w", err)
    }
    
    if len(results) == 0 {
        return results, nil
    }
    
    // Apply reranking if enabled
    if ar.config.RerankingEnabled && ar.reranker != nil {
        results, err = ar.reranker.Rerank(ctx, query, results)
        if err != nil {
            log.Printf("Reranking failed, using original order: %v", err)
        }
    }
    
    // Apply diversity filtering
    if ar.config.DiversityThreshold > 0 {
        results = ar.applyDiversityFilter(results)
    }
    
    // Limit final results
    if len(results) > ar.config.FinalResultLimit {
        results = results[:ar.config.FinalResultLimit]
    }
    
    return results, nil
}

func (ar *AdvancedRetriever) applyDiversityFilter(results []core.MemoryResult) []core.MemoryResult {
    if len(results) <= 1 {
        return results
    }
    
    filtered := []core.MemoryResult{results[0]} // Always include the top result
    
    for _, candidate := range results[1:] {
        isDiverse := true
        
        for _, selected := range filtered {
            similarity := ar.calculateSimilarity(candidate.Content, selected.Content)
            if similarity > ar.config.DiversityThreshold {
                isDiverse = false
                break
            }
        }
        
        if isDiverse {
            filtered = append(filtered, candidate)
        }
    }
    
    return filtered
}

func (ar *AdvancedRetriever) calculateSimilarity(text1, text2 string) float64 {
    // Simple similarity calculation (in production, use proper similarity metrics)
    words1 := strings.Fields(strings.ToLower(text1))
    words2 := strings.Fields(strings.ToLower(text2))
    
    wordSet1 := make(map[string]bool)
    for _, word := range words1 {
        wordSet1[word] = true
    }
    
    common := 0
    for _, word := range words2 {
        if wordSet1[word] {
            common++
        }
    }
    
    if len(words1) == 0 || len(words2) == 0 {
        return 0
    }
    
    return float64(common) / float64(len(words1)+len(words2)-common) // Jaccard similarity
}
```

### 3. Context-Aware Response Generation

```go
type ContextAwareGenerator struct {
    llm    core.LLMProvider
    config GenerationConfig
}

type GenerationConfig struct {
    MaxContextLength    int
    ResponseMaxLength   int
    IncludeSources      bool
    FactCheckingEnabled bool
    TemperatureAdjustment float32
}

func NewContextAwareGenerator(llm core.LLMProvider, config GenerationConfig) *ContextAwareGenerator {
    return &ContextAwareGenerator{
        llm:    llm,
        config: config,
    }
}

func (cag *ContextAwareGenerator) Generate(ctx context.Context, query string, retrievedContext []core.MemoryResult, conversationHistory []core.Message) (string, error) {
    // Build comprehensive context
    context := cag.buildContext(query, retrievedContext, conversationHistory)
    
    // Create generation prompt
    prompt := cag.createPrompt(query, context, retrievedContext)
    
    // Generate response
    response, err := cag.llm.Generate(ctx, prompt)
    if err != nil {
        return "", fmt.Errorf("generation failed: %w", err)
    }
    
    // Post-process response
    response = cag.postProcessResponse(response, retrievedContext)
    
    // Fact-check if enabled
    if cag.config.FactCheckingEnabled {
        response, err = cag.factCheck(ctx, response, retrievedContext)
        if err != nil {
            log.Printf("Fact-checking failed: %v", err)
        }
    }
    
    return response, nil
}

func (cag *ContextAwareGenerator) buildContext(query string, retrievedContext []core.MemoryResult, history []core.Message) string {
    var contextBuilder strings.Builder
    
    // Add conversation history if relevant
    if len(history) > 0 {
        contextBuilder.WriteString("Recent conversation:\n")
        for _, msg := range history {
            contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
        }
        contextBuilder.WriteString("\n")
    }
    
    // Add retrieved context
    if len(retrievedContext) > 0 {
        contextBuilder.WriteString("Relevant information:\n")
        for i, result := range retrievedContext {
            contextBuilder.WriteString(fmt.Sprintf("%d. %s", i+1, result.Content))
            
            // Add source information if available
            if source, ok := result.Metadata["source"]; ok {
                contextBuilder.WriteString(fmt.Sprintf(" (Source: %s)", source))
            }
            contextBuilder.WriteString("\n")
        }
    }
    
    context := contextBuilder.String()
    
    // Truncate if too long
    if len(context) > cag.config.MaxContextLength {
        context = context[:cag.config.MaxContextLength] + "...\n[Context truncated]"
    }
    
    return context
}

func (cag *ContextAwareGenerator) createPrompt(query, context string, retrievedContext []core.MemoryResult) string {
    var promptBuilder strings.Builder
    
    promptBuilder.WriteString("You are a knowledgeable assistant. Use the provided context to answer the user's question accurately and helpfully.\n\n")
    
    if context != "" {
        promptBuilder.WriteString("Context:\n")
        promptBuilder.WriteString(context)
        promptBuilder.WriteString("\n")
    }
    
    promptBuilder.WriteString(fmt.Sprintf("Question: %s\n\n", query))
    
    promptBuilder.WriteString("Instructions:\n")
    promptBuilder.WriteString("- Answer based on the provided context\n")
    promptBuilder.WriteString("- If the context doesn't contain enough information, say so\n")
    promptBuilder.WriteString("- Be accurate and cite sources when possible\n")
    
    if cag.config.IncludeSources && len(retrievedContext) > 0 {
        promptBuilder.WriteString("- Include source references in your response\n")
    }
    
    promptBuilder.WriteString("\nAnswer:")
    
    return promptBuilder.String()
}

func (cag *ContextAwareGenerator) postProcessResponse(response string, context []core.MemoryResult) string {
    // Clean up response
    response = strings.TrimSpace(response)
    
    // Add source citations if configured
    if cag.config.IncludeSources && len(context) > 0 {
        response = cag.addSourceCitations(response, context)
    }
    
    // Truncate if too long
    if len(response) > cag.config.ResponseMaxLength {
        response = response[:cag.config.ResponseMaxLength] + "..."
    }
    
    return response
}

func (cag *ContextAwareGenerator) addSourceCitations(response string, context []core.MemoryResult) string {
    if len(context) == 0 {
        return response
    }
    
    var sourcesBuilder strings.Builder
    sourcesBuilder.WriteString("\n\nSources:\n")
    
    for i, result := range context {
        if source, ok := result.Metadata["source"]; ok {
            sourcesBuilder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, source))
        } else {
            sourcesBuilder.WriteString(fmt.Sprintf("[%d] Internal knowledge base\n", i+1))
        }
    }
    
    return response + sourcesBuilder.String()
}

func (cag *ContextAwareGenerator) factCheck(ctx context.Context, response string, context []core.MemoryResult) (string, error) {
    // Create fact-checking prompt
    var contextBuilder strings.Builder
    for _, result := range context {
        contextBuilder.WriteString(fmt.Sprintf("- %s\n", result.Content))
    }
    
    prompt := fmt.Sprintf(`Please fact-check the following response against the provided context:

Context:
%s

Response to check:
%s

Is the response factually accurate based on the context? If there are any inaccuracies, please provide a corrected version.

Fact-check result:`, contextBuilder.String(), response)
    
    factCheckResult, err := cag.llm.Generate(ctx, prompt)
    if err != nil {
        return response, err // Return original response if fact-checking fails
    }
    
    // Simple heuristic: if the fact-check suggests corrections, use them
    if strings.Contains(strings.ToLower(factCheckResult), "corrected version") ||
       strings.Contains(strings.ToLower(factCheckResult), "inaccurate") {
        
        // Extract corrected version (this is a simplified approach)
        lines := strings.Split(factCheckResult, "\n")
        for i, line := range lines {
            if strings.Contains(strings.ToLower(line), "corrected") && i+1 < len(lines) {
                return strings.TrimSpace(lines[i+1]), nil
            }
        }
    }
    
    return response, nil
}
```

## RAG Agent Integration

### 1. Complete RAG Agent

```go
type ComprehensiveRAGAgent struct {
    name              string
    memory            core.Memory
    llm               core.LLMProvider
    queryEnhancer     *QueryEnhancer
    retriever         *AdvancedRetriever
    generator         *ContextAwareGenerator
    conversationMgr   *ConversationManager
}

type ConversationManager struct {
    memory core.Memory
}

func NewConversationManager(memory core.Memory) *ConversationManager {
    return &ConversationManager{memory: memory}
}

func (cm *ConversationManager) GetRecentHistory(ctx context.Context, sessionID string, limit int) ([]core.Message, error) {
    return cm.memory.GetHistory(ctx, limit,
        core.WithSession(sessionID),
        core.WithTimeRange(time.Now().Add(-24*time.Hour), time.Now()),
    )
}

func NewComprehensiveRAGAgent(name string, memory core.Memory, llm core.LLMProvider) *ComprehensiveRAGAgent {
    return &ComprehensiveRAGAgent{
        name:            name,
        memory:          memory,
        llm:             llm,
        queryEnhancer:   NewQueryEnhancer(llm),
        retriever: &AdvancedRetriever{
            memory:   memory,
            reranker: NewReranker(llm),
            config: RetrievalConfig{
                InitialRetrievalLimit: 10,
                FinalResultLimit:      5,
                ScoreThreshold:        0.7,
                RerankingEnabled:      true,
                DiversityThreshold:    0.8,
            },
        },
        generator: NewContextAwareGenerator(llm, GenerationConfig{
            MaxContextLength:      3000,
            ResponseMaxLength:     1500,
            IncludeSources:        true,
            FactCheckingEnabled:   true,
            TemperatureAdjustment: 0.1,
        }),
        conversationMgr: NewConversationManager(memory),
    }
}

func (cra *ComprehensiveRAGAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Extract query
    query, ok := state.Get("message")
    if !ok {
        return core.AgentResult{}, fmt.Errorf("no message in state")
    }
    queryStr := query.(string)
    sessionID := event.GetSessionID()
    
    // Get conversation history
    history, err := cra.conversationMgr.GetRecentHistory(ctx, sessionID, 5)
    if err != nil {
        log.Printf("Failed to get conversation history: %v", err)
        history = []core.Message{} // Continue without history
    }
    
    // Enhance query with conversation context
    enhancedQuery, err := cra.queryEnhancer.EnhanceQuery(ctx, queryStr, history)
    if err != nil {
        log.Printf("Query enhancement failed: %v", err)
        enhancedQuery = queryStr // Fallback to original
    }
    
    // Retrieve relevant context
    retrievedContext, err := cra.retriever.Retrieve(ctx, enhancedQuery)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("retrieval failed: %w", err)
    }
    
    // Generate response with context
    response, err := cra.generator.Generate(ctx, queryStr, retrievedContext, history)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("generation failed: %w", err)
    }
    
    // Store interaction
    err = cra.storeInteraction(ctx, sessionID, queryStr, response, retrievedContext)
    if err != nil {
        log.Printf("Failed to store interaction: %v", err)
    }
    
    // Prepare result
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("enhanced_query", enhancedQuery)
    outputState.Set("sources_count", len(retrievedContext))
    outputState.Set("context_used", len(retrievedContext) > 0)
    
    // Add source information
    sources := make([]string, 0, len(retrievedContext))
    for _, result := range retrievedContext {
        if source, ok := result.Metadata["source"]; ok {
            sources = append(sources, source)
        }
    }
    outputState.Set("sources", sources)
    
    return core.AgentResult{OutputState: outputState}, nil
}

func (cra *ComprehensiveRAGAgent) storeInteraction(ctx context.Context, sessionID, query, response string, context []core.MemoryResult) error {
    // Store user query
    err := cra.memory.Store(ctx, query, "user-message",
        core.WithSession(sessionID),
        core.WithTimestamp(time.Now()),
        core.WithMetadata(map[string]string{
            "agent_type": "comprehensive-rag",
        }),
    )
    if err != nil {
        return err
    }
    
    // Store agent response with context metadata
    contextSources := make([]string, 0, len(context))
    for _, result := range context {
        if source, ok := result.Metadata["source"]; ok {
            contextSources = append(contextSources, source)
        }
    }
    
    err = cra.memory.Store(ctx, response, "assistant-message",
        core.WithSession(sessionID),
        core.WithTimestamp(time.Now()),
        core.WithMetadata(map[string]string{
            "agent_type":     "comprehensive-rag",
            "sources_used":   strings.Join(contextSources, ", "),
            "sources_count":  fmt.Sprintf("%d", len(context)),
            "context_length": fmt.Sprintf("%d", len(strings.Join(contextSources, " "))),
        }),
    )
    
    return err
}
```

## RAG Performance Optimization

### 1. Caching Strategies

```go
type RAGCache struct {
    retrievalCache map[string][]core.MemoryResult
    responseCache  map[string]string
    mu             sync.RWMutex
    ttl            time.Duration
    timestamps     map[string]time.Time
}

func NewRAGCache(ttl time.Duration) *RAGCache {
    cache := &RAGCache{
        retrievalCache: make(map[string][]core.MemoryResult),
        responseCache:  make(map[string]string),
        timestamps:     make(map[string]time.Time),
        ttl:            ttl,
    }
    
    // Start cleanup goroutine
    go cache.cleanup()
    
    return cache
}

func (rc *RAGCache) GetRetrievalResults(query string) ([]core.MemoryResult, bool) {
    rc.mu.RLock()
    defer rc.mu.RUnlock()
    
    if timestamp, exists := rc.timestamps[query]; exists {
        if time.Since(timestamp) < rc.ttl {
            if results, exists := rc.retrievalCache[query]; exists {
                return results, true
            }
        }
    }
    
    return nil, false
}

func (rc *RAGCache) SetRetrievalResults(query string, results []core.MemoryResult) {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    
    rc.retrievalCache[query] = results
    rc.timestamps[query] = time.Now()
}

func (rc *RAGCache) cleanup() {
    ticker := time.NewTicker(rc.ttl / 2)
    defer ticker.Stop()
    
    for range ticker.C {
        rc.mu.Lock()
        now := time.Now()
        
        for query, timestamp := range rc.timestamps {
            if now.Sub(timestamp) > rc.ttl {
                delete(rc.retrievalCache, query)
                delete(rc.responseCache, query)
                delete(rc.timestamps, query)
            }
        }
        
        rc.mu.Unlock()
    }
}
```

### 2. Batch Processing

```go
type BatchRAGProcessor struct {
    agent     *ComprehensiveRAGAgent
    batchSize int
    timeout   time.Duration
}

func NewBatchRAGProcessor(agent *ComprehensiveRAGAgent, batchSize int, timeout time.Duration) *BatchRAGProcessor {
    return &BatchRAGProcessor{
        agent:     agent,
        batchSize: batchSize,
        timeout:   timeout,
    }
}

func (brp *BatchRAGProcessor) ProcessBatch(ctx context.Context, queries []string, sessionID string) ([]string, error) {
    responses := make([]string, len(queries))
    
    // Process in batches
    for i := 0; i < len(queries); i += brp.batchSize {
        end := i + brp.batchSize
        if end > len(queries) {
            end = len(queries)
        }
        
        batch := queries[i:end]
        batchResponses, err := brp.processBatch(ctx, batch, sessionID)
        if err != nil {
            return nil, fmt.Errorf("batch processing failed: %w", err)
        }
        
        copy(responses[i:], batchResponses)
    }
    
    return responses, nil
}

func (brp *BatchRAGProcessor) processBatch(ctx context.Context, queries []string, sessionID string) ([]string, error) {
    ctx, cancel := context.WithTimeout(ctx, brp.timeout)
    defer cancel()
    
    responses := make([]string, len(queries))
    var wg sync.WaitGroup
    var mu sync.Mutex
    var firstError error
    
    for i, query := range queries {
        wg.Add(1)
        go func(index int, q string) {
            defer wg.Done()
            
            event := core.NewEvent(
                brp.agent.name,
                core.EventData{"message": q},
                map[string]string{"session_id": sessionID},
            )
            
            state := core.NewState()
            state.Set("message", q)
            
            result, err := brp.agent.Run(ctx, event, state)
            
            mu.Lock()
            defer mu.Unlock()
            
            if err != nil && firstError == nil {
                firstError = err
            } else if err == nil {
                if response, ok := result.OutputState.Get("response"); ok {
                    responses[index] = response.(string)
                }
            }
        }(i, query)
    }
    
    wg.Wait()
    
    if firstError != nil {
        return nil, firstError
    }
    
    return responses, nil
}
```

## RAG Evaluation and Monitoring

### 1. RAG Metrics

```go
type RAGMetrics struct {
    RetrievalLatency    []time.Duration
    GenerationLatency   []time.Duration
    RetrievalAccuracy   float64
    ResponseQuality     float64
    SourceUtilization   map[string]int
    mu                  sync.RWMutex
}

func NewRAGMetrics() *RAGMetrics {
    return &RAGMetrics{
        SourceUtilization: make(map[string]int),
    }
}

func (rm *RAGMetrics) RecordRetrieval(latency time.Duration, resultsCount int, accuracy float64) {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    rm.RetrievalLatency = append(rm.RetrievalLatency, latency)
    rm.RetrievalAccuracy = (rm.RetrievalAccuracy + accuracy) / 2 // Simple moving average
}

func (rm *RAGMetrics) RecordGeneration(latency time.Duration, quality float64) {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    rm.GenerationLatency = append(rm.GenerationLatency, latency)
    rm.ResponseQuality = (rm.ResponseQuality + quality) / 2
}

func (rm *RAGMetrics) RecordSourceUsage(sources []string) {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    for _, source := range sources {
        rm.SourceUtilization[source]++
    }
}

func (rm *RAGMetrics) GetAverageRetrievalLatency() time.Duration {
    rm.mu.RLock()
    defer rm.mu.RUnlock()
    
    if len(rm.RetrievalLatency) == 0 {
        return 0
    }
    
    var total time.Duration
    for _, latency := range rm.RetrievalLatency {
        total += latency
    }
    
    return total / time.Duration(len(rm.RetrievalLatency))
}
```

### 2. Quality Assessment

```go
type RAGQualityAssessor struct {
    llm core.LLMProvider
}

func NewRAGQualityAssessor(llm core.LLMProvider) *RAGQualityAssessor {
    return &RAGQualityAssessor{llm: llm}
}

func (rqa *RAGQualityAssessor) AssessResponse(ctx context.Context, query, response string, sources []core.MemoryResult) (float64, error) {
    // Build assessment prompt
    var sourcesBuilder strings.Builder
    for i, source := range sources {
        sourcesBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, source.Content))
    }
    
    prompt := fmt.Sprintf(`Please assess the quality of this RAG response on a scale of 0.0 to 1.0:

Query: %s

Sources used:
%s

Response: %s

Assessment criteria:
- Accuracy: Is the response factually correct based on the sources?
- Relevance: Does the response directly address the query?
- Completeness: Does the response provide sufficient information?
- Coherence: Is the response well-structured and clear?

Please provide only a numerical score between 0.0 and 1.0:`, query, sourcesBuilder.String(), response)
    
    scoreStr, err := rqa.llm.Generate(ctx, prompt)
    if err != nil {
        return 0.0, fmt.Errorf("quality assessment failed: %w", err)
    }
    
    // Parse score
    scoreStr = strings.TrimSpace(scoreStr)
    score, err := strconv.ParseFloat(scoreStr, 64)
    if err != nil {
        return 0.0, fmt.Errorf("failed to parse quality score: %w", err)
    }
    
    // Clamp score to valid range
    if score < 0.0 {
        score = 0.0
    } else if score > 1.0 {
        score = 1.0
    }
    
    return score, nil
}
```

## Best Practices

### 1. RAG System Design

- **Chunk Size Optimization**: Balance context and specificity
- **Embedding Quality**: Use appropriate embedding models for your domain
- **Retrieval Tuning**: Optimize similarity thresholds and result limits
- **Context Management**: Manage context length to avoid token limits
- **Source Attribution**: Always track and cite information sources

### 2. Performance Optimization

- **Caching**: Cache frequent queries and embeddings
- **Batch Processing**: Process multiple queries efficiently
- **Index Optimization**: Use appropriate vector database indexes
- **Async Processing**: Use asynchronous operations where possible

### 3. Quality Assurance

- **Evaluation Metrics**: Implement comprehensive evaluation
- **Human Feedback**: Collect and incorporate user feedback
- **Continuous Monitoring**: Monitor system performance and quality
- **A/B Testing**: Test different RAG configurations

## Conclusion

RAG implementation in AgenticGoKit enables agents to provide accurate, contextual, and up-to-date responses by combining retrieval and generation. Key takeaways:

- Start with basic RAG and gradually add advanced features
- Optimize retrieval quality through query enhancement and reranking
- Implement proper caching and performance monitoring
- Continuously evaluate and improve system quality

RAG transforms agents from static responders to dynamic, knowledgeable assistants that can access and utilize vast amounts of information effectively.

## Next Steps

- [Knowledge Bases](knowledge-bases.md) - Build comprehensive knowledge systems
- [Memory Optimization](memory-optimization.md) - Advanced performance tuning
- [Production Deployment](../../guides/deployment/README.md) - Deploy RAG systems at scale

## Further Reading

- [RAG Research Papers](https://arxiv.org/abs/2005.11401)
- [Vector Database Comparison](https://github.com/pgvector/pgvector)
- [Embedding Model Evaluation](https://huggingface.co/spaces/mteb/leaderboard)