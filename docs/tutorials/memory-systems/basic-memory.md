---
title: Basic Memory Operations
description: Learn the fundamentals of AgenticGoKit's Memory interface, including personal memory operations, chat history management, and session handling.
---

# Basic Memory Operations in AgenticGoKit

## Overview

This tutorial covers the fundamentals of AgenticGoKit's Memory interface, including personal memory operations, chat history management, and session handling. You'll learn how to store and retrieve information, manage conversations, and build memory-enabled agents using the unified Memory interface.

Basic memory operations provide the foundation for all memory-enabled agents and are essential building blocks for advanced features like RAG and knowledge bases.

## Prerequisites

- Understanding of [Core Concepts](../core-concepts/README.md)
- Basic knowledge of Go programming and interfaces
- Familiarity with context management in Go
- Understanding of key-value storage concepts

## Memory Interface Overview

AgenticGoKit provides a unified Memory interface that handles all memory operations:

```go
type Memory interface {
    // Personal memory operations
    Store(ctx context.Context, content string, tags ...string) error
    Query(ctx context.Context, query string, limit ...int) ([]Result, error)
    Remember(ctx context.Context, key string, value any) error
    Recall(ctx context.Context, key string) (any, error)

    // Chat history management
    AddMessage(ctx context.Context, role, content string) error
    GetHistory(ctx context.Context, limit ...int) ([]Message, error)

    // Session management
    NewSession() string
    SetSession(ctx context.Context, sessionID string) context.Context
    ClearSession(ctx context.Context) error
    Close() error

    // RAG-Enhanced Knowledge Base Operations
    IngestDocument(ctx context.Context, doc Document) error
    IngestDocuments(ctx context.Context, docs []Document) error
    SearchKnowledge(ctx context.Context, query string, options ...SearchOption) ([]KnowledgeResult, error)

    // Hybrid Search (Personal Memory + Knowledge Base)
    SearchAll(ctx context.Context, query string, options ...SearchOption) (*HybridResult, error)

    // RAG Context Assembly for LLM Prompts
    BuildContext(ctx context.Context, query string, options ...ContextOption) (*RAGContext, error)
}
```

## Setting Up Basic Memory

### 1. Creating a Memory Instance

The simplest way to get started is with the in-memory provider:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create in-memory storage
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "memory",
        Connection: "memory",
        MaxResults: 10,
        Dimensions: 1536,
        AutoEmbed:  true,
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    defer memory.Close()
    
    ctx := context.Background()
    
    // Test basic operations
    err = demonstrateBasicOperations(ctx, memory)
    if err != nil {
        log.Fatalf("Demo failed: %v", err)
    }
}

func demonstrateBasicOperations(ctx context.Context, memory core.Memory) error {
    // Store some facts with tags
    err := memory.Store(ctx, "Paris is the capital of France", "fact", "geography")
    if err != nil {
        return fmt.Errorf("failed to store fact: %w", err)
    }
    
    // Query for information
    results, err := memory.Query(ctx, "capital France", 5)
    if err != nil {
        return fmt.Errorf("failed to query: %w", err)
    }
    
    fmt.Printf("Found %d results:\n", len(results))
    for _, result := range results {
        fmt.Printf("- %s (Score: %.3f, Tags: %v)\n", 
            result.Content, result.Score, result.Tags)
    }
    
    return nil
}
```

### 2. Memory Configuration Options

::: code-group

```go [Development]
// Development configuration - fast iteration
devConfig := core.AgentMemoryConfig{
    Provider:   "memory",
    Connection: "memory",
    MaxResults: 10,
    Dimensions: 1536,
    AutoEmbed:  true,
    
    // Enable RAG features
    EnableRAG:           true,
    EnableKnowledgeBase: true,
    
    // Use dummy embeddings for development
    Embedding: core.EmbeddingConfig{
        Provider: "dummy",
        Model:    "dummy-model",
    },
}

memory, err := core.NewMemory(devConfig)
if err != nil {
    log.Fatalf("Failed to create memory: %v", err)
}
defer memory.Close()
```

```go [Production]
// Production configuration - optimized for scale
prodConfig := core.AgentMemoryConfig{
    Provider:   "pgvector",
    Connection: "postgres://user:pass@localhost:5432/agentdb",
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
        Provider:        "openai",
        Model:           "text-embedding-3-small",
        APIKey:          os.Getenv("OPENAI_API_KEY"),
        CacheEmbeddings: true,
        MaxBatchSize:    100,
        TimeoutSeconds:  30,
    },
    
    Documents: core.DocumentConfig{
        AutoChunk:                true,
        SupportedTypes:           []string{"pdf", "txt", "md", "web", "code"},
        MaxFileSize:              "10MB",
        EnableMetadataExtraction: true,
        EnableURLScraping:        true,
    },
    
    Search: core.SearchConfigToml{
        HybridSearch:         true,
        KeywordWeight:        0.3,
        SemanticWeight:       0.7,
        EnableReranking:      false,
        EnableQueryExpansion: false,
    },
}

memory, err := core.NewMemory(prodConfig)
if err != nil {
    log.Fatalf("Failed to create memory: %v", err)
}
defer memory.Close()
```

```toml [Configuration File]
# config.toml - External configuration
[memory]
provider = "pgvector"
connection = "postgres://user:pass@localhost:5432/agentdb"
max_results = 10
dimensions = 1536
auto_embed = true

# RAG settings
enable_rag = true
enable_knowledge_base = true
knowledge_max_results = 20
knowledge_score_threshold = 0.7
chunk_size = 1000
chunk_overlap = 200

[memory.embedding]
provider = "openai"
model = "text-embedding-3-small"
cache_embeddings = true
max_batch_size = 100
timeout_seconds = 30

[memory.documents]
auto_chunk = true
supported_types = ["pdf", "txt", "md", "web", "code"]
max_file_size = "10MB"
enable_metadata_extraction = true

[memory.search]
hybrid_search = true
keyword_weight = 0.3
semantic_weight = 0.7
```

:::

::: tip Configuration Strategy
- **Development**: Use in-memory storage with dummy embeddings for fast iteration
- **Testing**: Use pgvector with dummy embeddings to test database integration
- **Production**: Use pgvector/Weaviate with real embedding services for full functionality
:::



## Personal Memory Operations

The Memory interface provides two types of personal memory operations: content storage with tags and key-value storage.

### 1. Content Storage with Store and Query

Use `Store` and `Query` for content that you want to search semantically:

```go
func demonstrateContentStorage(ctx context.Context, memory core.Memory) error {
    // Store facts with tags
    facts := []struct {
        content string
        tags    []string
    }{
        {"The Earth orbits the Sun", []string{"science", "astronomy", "fact"}},
        {"Water boils at 100°C at sea level", []string{"science", "physics", "fact"}},
        {"Shakespeare wrote Hamlet", []string{"literature", "history", "fact"}},
        {"Go is a programming language developed by Google", []string{"programming", "go", "technology"}},
    }
    
    for _, fact := range facts {
        err := memory.Store(ctx, fact.content, fact.tags...)
        if err != nil {
            return fmt.Errorf("failed to store fact: %w", err)
        }
    }
    
    fmt.Printf("Stored %d facts\n", len(facts))
    
    // Query for information
    results, err := memory.Query(ctx, "programming language", 3)
    if err != nil {
        return fmt.Errorf("failed to query: %w", err)
    }
    
    fmt.Printf("Found %d results for 'programming language':\n", len(results))
    for i, result := range results {
        fmt.Printf("%d. %s (Score: %.3f, Tags: %v)\n", 
            i+1, result.Content, result.Score, result.Tags)
    }
    
    return nil
}
```

### 2. Key-Value Storage with Remember and Recall

Use `Remember` and `Recall` for structured data and preferences:

```go
func demonstrateKeyValueStorage(ctx context.Context, memory core.Memory) error {
    // Store user preferences and settings
    preferences := map[string]any{
        "user_name":          "Alice Johnson",
        "communication_style": "technical",
        "preferred_language":  "Go",
        "experience_level":    "intermediate",
        "notification_settings": map[string]bool{
            "email":  true,
            "push":   false,
            "sms":    false,
        },
    }
    
    // Store each preference
    for key, value := range preferences {
        err := memory.Remember(ctx, key, value)
        if err != nil {
            return fmt.Errorf("failed to remember %s: %w", key, err)
        }
    }
    
    fmt.Println("Stored user preferences")
    
    // Retrieve specific preferences
    userName, err := memory.Recall(ctx, "user_name")
    if err != nil {
        return fmt.Errorf("failed to recall user_name: %w", err)
    }
    
    commStyle, err := memory.Recall(ctx, "communication_style")
    if err != nil {
        return fmt.Errorf("failed to recall communication_style: %w", err)
    }
    
    fmt.Printf("User: %v, Communication Style: %v\n", userName, commStyle)
    
    // Handle missing keys
    theme, err := memory.Recall(ctx, "theme_preference")
    if err != nil {
        fmt.Printf("Theme preference not set, using default\n")
    } else {
        fmt.Printf("Theme: %v\n", theme)
    }
    
    return nil
}
```

### 3. Combining Both Approaches

```go
func demonstrateCombinedUsage(ctx context.Context, memory core.Memory) error {
    // Store user profile using Remember
    err := memory.Remember(ctx, "user_id", "user-123")
    if err != nil {
        return err
    }
    
    err = memory.Remember(ctx, "user_preferences", map[string]string{
        "response_style": "detailed",
        "expertise_level": "intermediate",
    })
    if err != nil {
        return err
    }
    
    // Store user's questions and interests using Store
    interests := []string{
        "How do I implement concurrency in Go?",
        "What are the best practices for error handling?",
        "How do I optimize Go applications for performance?",
    }
    
    for _, interest := range interests {
        err := memory.Store(ctx, interest, "user-interest", "question")
        if err != nil {
            return err
        }
    }
    
    // Query for related interests
    results, err := memory.Query(ctx, "Go performance optimization", 2)
    if err != nil {
        return err
    }
    
    // Retrieve user preferences for personalized response
    prefs, err := memory.Recall(ctx, "user_preferences")
    if err != nil {
        return err
    }
    
    fmt.Printf("Found %d related interests for user with preferences: %v\n", 
        len(results), prefs)
    
    for _, result := range results {
        fmt.Printf("- %s (Score: %.3f)\n", result.Content, result.Score)
    }
    
    return nil
}
```

## Query Operations and Search Patterns

### 1. Basic Query Operations

```go
func demonstrateQueryOperations(ctx context.Context, memory core.Memory) error {
    // Store some test data first
    testData := []struct {
        content string
        tags    []string
    }{
        {"Go is excellent for concurrent programming", []string{"go", "programming", "concurrency"}},
        {"Python is great for data science and AI", []string{"python", "programming", "ai", "data-science"}},
        {"JavaScript runs in browsers and servers", []string{"javascript", "programming", "web"}},
        {"Rust provides memory safety without garbage collection", []string{"rust", "programming", "memory-safety"}},
    }
    
    for _, item := range testData {
        err := memory.Store(ctx, item.content, item.tags...)
        if err != nil {
            return err
        }
    }
    
    // Basic query - finds semantically similar content
    results, err := memory.Query(ctx, "concurrent programming languages")
    if err != nil {
        return fmt.Errorf("query failed: %w", err)
    }
    
    fmt.Printf("Query: 'concurrent programming languages'\n")
    fmt.Printf("Found %d results:\n", len(results))
    for i, result := range results {
        fmt.Printf("%d. %s (Score: %.3f)\n", 
            i+1, result.Content, result.Score)
        fmt.Printf("   Tags: %v\n", result.Tags)
    }
    
    return nil
}
```

### 2. Query with Limits

```go
func demonstrateQueryLimits(ctx context.Context, memory core.Memory) error {
    // Query with different limits
    queries := []struct {
        query string
        limit int
    }{
        {"programming", 2},
        {"memory", 1},
        {"web development", 3},
    }
    
    for _, q := range queries {
        results, err := memory.Query(ctx, q.query, q.limit)
        if err != nil {
            return fmt.Errorf("query failed for '%s': %w", q.query, err)
        }
        
        fmt.Printf("Query: '%s' (limit: %d)\n", q.query, q.limit)
        for i, result := range results {
            fmt.Printf("  %d. %s (%.3f)\n", i+1, result.Content, result.Score)
        }
        fmt.Println()
    }
    
    return nil
}
```

### 3. Understanding Query Results

```go
func analyzeQueryResults(ctx context.Context, memory core.Memory) error {
    // Store content with different relevance levels
    content := []string{
        "Go programming language features",
        "Go is used for backend development",
        "Golang has excellent concurrency support",
        "Python programming basics",
        "Java enterprise applications",
    }
    
    for _, item := range content {
        memory.Store(ctx, item, "programming")
    }
    
    // Query and analyze results
    results, err := memory.Query(ctx, "Go language features", 5)
    if err != nil {
        return err
    }
    
    fmt.Println("Query: 'Go language features'")
    fmt.Println("Results ordered by relevance:")
    
    for i, result := range results {
        relevance := "Low"
        if result.Score > 0.8 {
            relevance = "High"
        } else if result.Score > 0.5 {
            relevance = "Medium"
        }
        
        fmt.Printf("%d. %s\n", i+1, result.Content)
        fmt.Printf("   Score: %.3f (%s relevance)\n", result.Score, relevance)
        fmt.Printf("   Created: %s\n", result.CreatedAt.Format("2006-01-02 15:04:05"))
        fmt.Println()
    }
    
    return nil
}
```

## Chat History Management

AgenticGoKit provides dedicated methods for managing conversation history with proper session handling.

### 1. Session Management

```go
func demonstrateSessionManagement(ctx context.Context, memory core.Memory) error {
    // Create a new session
    sessionID := memory.NewSession()
    fmt.Printf("Created new session: %s\n", sessionID)
    
    // Set session context
    ctx = memory.SetSession(ctx, sessionID)
    
    // All subsequent operations will use this session
    err := memory.AddMessage(ctx, "user", "Hello, I'm starting a new conversation")
    if err != nil {
        return fmt.Errorf("failed to add message: %w", err)
    }
    
    err = memory.AddMessage(ctx, "assistant", "Hello! I'm here to help. What can I do for you?")
    if err != nil {
        return fmt.Errorf("failed to add message: %w", err)
    }
    
    // Get history for this session
    messages, err := memory.GetHistory(ctx, 10)
    if err != nil {
        return fmt.Errorf("failed to get history: %w", err)
    }
    
    fmt.Printf("Session %s has %d messages\n", sessionID, len(messages))
    return nil
}
```

### 2. Adding and Retrieving Messages

```go
func demonstrateChatHistory(ctx context.Context, memory core.Memory) error {
    // Create and set session
    sessionID := memory.NewSession()
    ctx = memory.SetSession(ctx, sessionID)
    
    // Simulate a conversation about Go programming
    conversation := []struct {
        role    string
        content string
    }{
        {"user", "What is Go programming language?"},
        {"assistant", "Go is a statically typed, compiled programming language developed by Google. It's known for its simplicity, efficiency, and excellent concurrency support."},
        {"user", "What makes Go good for concurrent programming?"},
        {"assistant", "Go has built-in concurrency primitives like goroutines and channels. Goroutines are lightweight threads, and channels provide a way to communicate between them safely."},
        {"user", "Can you show me a simple example?"},
        {"assistant", "Sure! Here's a basic example:\n\ngo func() {\n    fmt.Println(\"Hello from goroutine\")\n}()\n\nThis creates a goroutine that runs concurrently."},
    }
    
    // Add messages to the conversation
    for _, msg := range conversation {
        err := memory.AddMessage(ctx, msg.role, msg.content)
        if err != nil {
            return fmt.Errorf("failed to add %s message: %w", msg.role, err)
        }
        
        // Small delay to simulate real conversation timing
        time.Sleep(100 * time.Millisecond)
    }
    
    // Retrieve conversation history
    messages, err := memory.GetHistory(ctx, 10)
    if err != nil {
        return fmt.Errorf("failed to get history: %w", err)
    }
    
    fmt.Printf("Conversation History (%d messages):\n", len(messages))
    fmt.Println(strings.Repeat("-", 50))
    
    for _, msg := range messages {
        timestamp := msg.CreatedAt.Format("15:04:05")
        fmt.Printf("[%s] %s: %s\n", timestamp, strings.Title(msg.Role), msg.Content)
        fmt.Println()
    }
    
    return nil
}
```

### 3. Managing Multiple Sessions

```go
func demonstrateMultipleSessions(ctx context.Context, memory core.Memory) error {
    // Create multiple sessions for different users/conversations
    sessions := make(map[string]string)
    
    // User 1 session
    sessions["user1"] = memory.NewSession()
    ctx1 := memory.SetSession(ctx, sessions["user1"])
    
    memory.AddMessage(ctx1, "user", "I'm interested in learning Go")
    memory.AddMessage(ctx1, "assistant", "Great! Go is an excellent language to learn.")
    
    // User 2 session
    sessions["user2"] = memory.NewSession()
    ctx2 := memory.SetSession(ctx, sessions["user2"])
    
    memory.AddMessage(ctx2, "user", "I need help with Python")
    memory.AddMessage(ctx2, "assistant", "I'd be happy to help with Python!")
    
    // Retrieve history for each session
    for user, sessionID := range sessions {
        sessionCtx := memory.SetSession(ctx, sessionID)
        messages, err := memory.GetHistory(sessionCtx, 10)
        if err != nil {
            return fmt.Errorf("failed to get history for %s: %w", user, err)
        }
        
        fmt.Printf("%s's conversation (%s):\n", user, sessionID)
        for _, msg := range messages {
            fmt.Printf("  %s: %s\n", msg.Role, msg.Content)
        }
        fmt.Println()
    }
    
    return nil
}
```

### 4. Session Cleanup

```go
func demonstrateSessionCleanup(ctx context.Context, memory core.Memory) error {
    // Create a session and add some messages
    sessionID := memory.NewSession()
    ctx = memory.SetSession(ctx, sessionID)
    
    memory.AddMessage(ctx, "user", "This is a test message")
    memory.AddMessage(ctx, "assistant", "This is a test response")
    
    // Check history before cleanup
    messages, err := memory.GetHistory(ctx, 10)
    if err != nil {
        return err
    }
    fmt.Printf("Messages before cleanup: %d\n", len(messages))
    
    // Clear the session
    err = memory.ClearSession(ctx)
    if err != nil {
        return fmt.Errorf("failed to clear session: %w", err)
    }
    
    // Check history after cleanup
    messages, err = memory.GetHistory(ctx, 10)
    if err != nil {
        return err
    }
    fmt.Printf("Messages after cleanup: %d\n", len(messages))
    
    return nil
}
```

## Building Memory-Enabled Agents

### 1. Simple Memory Agent

```go
type BasicMemoryAgent struct {
    name   string
    memory core.Memory
    llm    core.LLMProvider
}

func NewBasicMemoryAgent(name string, memory core.Memory, llm core.LLMProvider) *BasicMemoryAgent {
    return &BasicMemoryAgent{
        name:   name,
        memory: memory,
        llm:    llm,
    }
}

func (m *BasicMemoryAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Extract message from state
    message, ok := state.Get("message")
    if !ok {
        return core.AgentResult{}, fmt.Errorf("no message in state")
    }
    messageStr := message.(string)
    
    // Set session context
    sessionID := event.GetSessionID()
    ctx = m.memory.SetSession(ctx, sessionID)
    
    // Store the user message in chat history
    err := m.memory.AddMessage(ctx, "user", messageStr)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to store user message: %w", err)
    }
    
    // Get conversation context
    context, err := m.buildContext(ctx, messageStr)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to build context: %w", err)
    }
    
    // Generate response with context
    response, err := m.llm.Generate(ctx, context)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to generate response: %w", err)
    }
    
    // Store the assistant response
    err = m.memory.AddMessage(ctx, "assistant", response)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to store assistant response: %w", err)
    }
    
    // Return result
    outputState := state.Clone()
    outputState.Set("response", response)
    
    return core.AgentResult{OutputState: outputState}, nil
}

func (m *BasicMemoryAgent) buildContext(ctx context.Context, currentMessage string) (string, error) {
    // Get recent conversation history
    history, err := m.memory.GetHistory(ctx, 5)
    if err != nil {
        return "", fmt.Errorf("failed to get history: %w", err)
    }
    
    // Search for relevant stored information
    relevant, err := m.memory.Query(ctx, currentMessage, 3)
    if err != nil {
        return "", fmt.Errorf("failed to query memory: %w", err)
    }
    
    // Build context
    var contextBuilder strings.Builder
    
    contextBuilder.WriteString("You are a helpful assistant with access to conversation history and stored information.\n\n")
    
    // Add relevant stored information
    if len(relevant) > 0 {
        contextBuilder.WriteString("Relevant stored information:\n")
        for _, item := range relevant {
            contextBuilder.WriteString(fmt.Sprintf("- %s\n", item.Content))
        }
        contextBuilder.WriteString("\n")
    }
    
    // Add conversation history (excluding the current message)
    if len(history) > 1 { // More than just the current message
        contextBuilder.WriteString("Recent conversation:\n")
        for _, msg := range history[:len(history)-1] { // Exclude the last (current) message
            contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", 
                strings.Title(msg.Role), msg.Content))
        }
        contextBuilder.WriteString("\n")
    }
    
    contextBuilder.WriteString(fmt.Sprintf("Current message: %s\n\n", currentMessage))
    contextBuilder.WriteString("Please provide a helpful response based on the context above.")
    
    return contextBuilder.String(), nil
}
```

### 2. Personalized Memory Agent

```go
type PersonalizedAgent struct {
    name   string
    memory core.Memory
    llm    core.LLMProvider
}

func NewPersonalizedAgent(name string, memory core.Memory, llm core.LLMProvider) *PersonalizedAgent {
    return &PersonalizedAgent{
        name:   name,
        memory: memory,
        llm:    llm,
    }
}

func (p *PersonalizedAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    message, _ := state.Get("message")
    messageStr := message.(string)
    
    // Set session context
    sessionID := event.GetSessionID()
    ctx = p.memory.SetSession(ctx, sessionID)
    
    // Check for user preferences
    userPrefs, err := p.getUserPreferences(ctx)
    if err != nil {
        log.Printf("Failed to get user preferences: %v", err)
        userPrefs = map[string]any{} // Use empty preferences
    }
    
    // Store user message
    p.memory.AddMessage(ctx, "user", messageStr)
    
    // Build personalized context
    context := p.buildPersonalizedContext(ctx, messageStr, userPrefs)
    
    // Generate response
    response, err := p.llm.Generate(ctx, context)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Store response
    p.memory.AddMessage(ctx, "assistant", response)
    
    // Learn from interaction
    p.learnFromInteraction(ctx, messageStr, response)
    
    outputState := state.Clone()
    outputState.Set("response", response)
    outputState.Set("personalized", true)
    
    return core.AgentResult{OutputState: outputState}, nil
}

func (p *PersonalizedAgent) getUserPreferences(ctx context.Context) (map[string]any, error) {
    prefs := make(map[string]any)
    
    // Try to recall various preferences
    prefKeys := []string{
        "communication_style",
        "expertise_level", 
        "preferred_format",
        "language_preference",
    }
    
    for _, key := range prefKeys {
        if value, err := p.memory.Recall(ctx, key); err == nil {
            prefs[key] = value
        }
    }
    
    return prefs, nil
}

func (p *PersonalizedAgent) buildPersonalizedContext(ctx context.Context, message string, prefs map[string]any) string {
    var builder strings.Builder
    
    builder.WriteString("You are a personalized assistant. Adapt your response based on user preferences.\n\n")
    
    // Add user preferences
    if len(prefs) > 0 {
        builder.WriteString("User Preferences:\n")
        for key, value := range prefs {
            builder.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
        }
        builder.WriteString("\n")
    }
    
    // Get relevant stored information
    relevant, err := p.memory.Query(ctx, message, 3)
    if err == nil && len(relevant) > 0 {
        builder.WriteString("Relevant Information:\n")
        for _, item := range relevant {
            builder.WriteString(fmt.Sprintf("- %s\n", item.Content))
        }
        builder.WriteString("\n")
    }
    
    // Get conversation history
    history, err := p.memory.GetHistory(ctx, 5)
    if err == nil && len(history) > 1 {
        builder.WriteString("Recent Conversation:\n")
        for _, msg := range history[:len(history)-1] {
            builder.WriteString(fmt.Sprintf("%s: %s\n", strings.Title(msg.Role), msg.Content))
        }
        builder.WriteString("\n")
    }
    
    builder.WriteString(fmt.Sprintf("Current Message: %s\n\n", message))
    builder.WriteString("Provide a helpful, personalized response based on the user's preferences and context.")
    
    return builder.String()
}

func (p *PersonalizedAgent) learnFromInteraction(ctx context.Context, userMessage, response string) {
    // Store interaction patterns for learning
    interaction := fmt.Sprintf("User: %s | Assistant: %s", userMessage, response)
    p.memory.Store(ctx, interaction, "interaction", "learning")
    
    // Simple learning heuristics
    if strings.Contains(strings.ToLower(userMessage), "explain") {
        p.memory.Remember(ctx, "prefers_explanations", true)
    }
    
    if strings.Contains(strings.ToLower(userMessage), "example") {
        p.memory.Remember(ctx, "prefers_examples", true)
    }
    
    if strings.Contains(strings.ToLower(userMessage), "simple") {
        p.memory.Remember(ctx, "communication_style", "simple")
    }
}
```

### 3. Complete Example with Agent Integration

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create memory system
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "memory",
        Connection: "memory",
        MaxResults: 10,
        Dimensions: 1536,
        EnableRAG:  true,
        
        Embedding: core.EmbeddingConfig{
            Provider: "dummy", // Use "openai" for production
            Model:    "dummy-model",
        },
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    defer memory.Close()
    
    // Create a mock LLM for demonstration
    llm := &MockLLM{}
    
    // Create memory-enabled agent
    agent := NewPersonalizedAgent("memory-assistant", memory, llm)
    
    // Store some initial knowledge
    ctx := context.Background()
    facts := []string{
        "Go is a programming language developed by Google",
        "AgenticGoKit is a framework for building multi-agent systems in Go",
        "Vector databases are used for similarity search and RAG implementations",
        "Memory systems enable agents to learn and remember information",
    }
    
    for _, fact := range facts {
        memory.Store(ctx, fact, "fact", "knowledge")
    }
    
    // Simulate a conversation
    sessionID := memory.NewSession()
    
    messages := []string{
        "What is Go programming language?",
        "I prefer simple explanations",
        "How is Go related to AgenticGoKit?",
        "Can you give me an example of memory systems?",
    }
    
    fmt.Println("=== Memory-Enabled Agent Conversation ===\n")
    
    for i, msg := range messages {
        fmt.Printf("User: %s\n", msg)
        
        // Create event and state
        event := core.NewEvent(
            "memory-assistant",
            core.EventData{"message": msg},
            map[string]string{"session_id": sessionID},
        )
        
        state := core.NewState()
        state.Set("message", msg)
        
        // Run agent
        result, err := agent.Run(ctx, event, state)
        if err != nil {
            log.Printf("Agent error: %v", err)
            continue
        }
        
        response, _ := result.OutputState.Get("response")
        fmt.Printf("Assistant: %s\n", response)
        fmt.Println(strings.Repeat("-", 50))
        
        // Small delay between messages
        time.Sleep(500 * time.Millisecond)
    }
    
    // Show what the agent learned
    fmt.Println("\n=== Agent Learning Summary ===")
    showAgentLearning(ctx, memory, sessionID)
}

// MockLLM for demonstration purposes
type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, prompt string) (string, error) {
    // Simple mock responses based on prompt content
    if strings.Contains(prompt, "Go programming") {
        return "Go is a statically typed, compiled programming language known for its simplicity and efficiency.", nil
    }
    if strings.Contains(prompt, "AgenticGoKit") {
        return "AgenticGoKit is a Go framework that makes it easy to build multi-agent systems with memory and LLM integration.", nil
    }
    if strings.Contains(prompt, "memory systems") {
        return "Memory systems allow agents to store, retrieve, and learn from information across conversations.", nil
    }
    if strings.Contains(prompt, "simple") {
        return "I'll keep my explanations simple and clear as you prefer.", nil
    }
    
    return "I understand your question and I'm here to help based on our conversation history.", nil
}

func showAgentLearning(ctx context.Context, memory core.Memory, sessionID string) {
    ctx = memory.SetSession(ctx, sessionID)
    
    // Show conversation history
    history, err := memory.GetHistory(ctx, 10)
    if err != nil {
        log.Printf("Failed to get history: %v", err)
        return
    }
    
    fmt.Printf("Conversation History (%d messages):\n", len(history))
    for _, msg := range history {
        fmt.Printf("- %s: %s\n", strings.Title(msg.Role), 
            truncateString(msg.Content, 60))
    }
    
    // Show learned preferences
    fmt.Println("\nLearned Preferences:")
    prefKeys := []string{"communication_style", "prefers_explanations", "prefers_examples"}
    
    for _, key := range prefKeys {
        if value, err := memory.Recall(ctx, key); err == nil {
            fmt.Printf("- %s: %v\n", key, value)
        }
    }
    
    // Show stored interactions
    interactions, err := memory.Query(ctx, "interaction", 5)
    if err == nil && len(interactions) > 0 {
        fmt.Printf("\nStored Interactions: %d\n", len(interactions))
    }
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}
```

## Memory Management and Cleanup

### 1. Session Management Best Practices

```go
func demonstrateSessionBestPractices(ctx context.Context, memory core.Memory) error {
    // Create session with meaningful ID
    userID := "user-123"
    timestamp := time.Now().Unix()
    sessionID := fmt.Sprintf("%s-%d", userID, timestamp)
    
    // Alternative: use the built-in session generator
    // sessionID := memory.NewSession()
    
    ctx = memory.SetSession(ctx, sessionID)
    
    // Store session metadata
    memory.Remember(ctx, "session_start", time.Now())
    memory.Remember(ctx, "user_id", userID)
    
    // Use session for conversation
    memory.AddMessage(ctx, "user", "Hello, I'm starting a new session")
    memory.AddMessage(ctx, "assistant", "Welcome! I'm ready to help.")
    
    // Later, when session ends
    memory.Remember(ctx, "session_end", time.Now())
    
    fmt.Printf("Session %s managed successfully\n", sessionID)
    return nil
}
```

### 2. Memory Cleanup Strategies

```go
func demonstrateMemoryCleanup(ctx context.Context, memory core.Memory) error {
    // Store some temporary data
    tempData := []string{
        "Temporary calculation result: 42",
        "Cache entry for user session",
        "Intermediate processing data",
    }
    
    for _, data := range tempData {
        memory.Store(ctx, data, "temporary", "cache")
    }
    
    fmt.Printf("Stored %d temporary items\n", len(tempData))
    
    // Query all temporary items
    tempItems, err := memory.Query(ctx, "temporary", 10)
    if err != nil {
        return err
    }
    
    fmt.Printf("Found %d temporary items before cleanup\n", len(tempItems))
    
    // Clear session to clean up session-specific data
    err = memory.ClearSession(ctx)
    if err != nil {
        return fmt.Errorf("failed to clear session: %w", err)
    }
    
    fmt.Println("Session cleared successfully")
    
    return nil
}
```

### 3. Memory Usage Monitoring

```go
type MemoryMonitor struct {
    memory core.Memory
}

func NewMemoryMonitor(memory core.Memory) *MemoryMonitor {
    return &MemoryMonitor{memory: memory}
}

func (m *MemoryMonitor) MonitorUsage(ctx context.Context) error {
    // Monitor different types of stored data
    categories := []string{"fact", "preference", "temporary", "interaction"}
    
    fmt.Println("Memory Usage Report:")
    fmt.Println(strings.Repeat("=", 40))
    
    totalItems := 0
    for _, category := range categories {
        items, err := m.memory.Query(ctx, category, 100) // Get up to 100 items
        if err != nil {
            log.Printf("Failed to query %s: %v", category, err)
            continue
        }
        
        fmt.Printf("%-12s: %d items\n", strings.Title(category), len(items))
        totalItems += len(items)
    }
    
    fmt.Println(strings.Repeat("-", 40))
    fmt.Printf("%-12s: %d items\n", "Total", totalItems)
    
    return nil
}

func (m *MemoryMonitor) ShowRecentActivity(ctx context.Context, sessionID string) error {
    ctx = m.memory.SetSession(ctx, sessionID)
    
    // Get recent messages
    messages, err := m.memory.GetHistory(ctx, 10)
    if err != nil {
        return err
    }
    
    fmt.Printf("\nRecent Activity (Session: %s):\n", sessionID)
    fmt.Println(strings.Repeat("=", 50))
    
    if len(messages) == 0 {
        fmt.Println("No recent activity")
        return nil
    }
    
    for _, msg := range messages {
        timestamp := msg.CreatedAt.Format("15:04:05")
        fmt.Printf("[%s] %s: %s\n", 
            timestamp, 
            strings.Title(msg.Role), 
            truncateString(msg.Content, 50))
    }
    
    return nil
}
```

### 4. Resource Management

```go
func demonstrateResourceManagement() {
    // Create memory with proper resource management
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "memory",
        Connection: "memory",
        MaxResults: 10,
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    
    // Always close memory when done
    defer func() {
        if err := memory.Close(); err != nil {
            log.Printf("Error closing memory: %v", err)
        } else {
            fmt.Println("Memory closed successfully")
        }
    }()
    
    ctx := context.Background()
    
    // Use memory operations
    memory.Store(ctx, "Test data", "test")
    
    // Memory will be properly closed when function exits
    fmt.Println("Memory operations completed")
}

// Example of graceful shutdown with context
func demonstrateGracefulShutdown() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "memory",
        Connection: "memory",
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    
    // Use context for all operations
    sessionID := memory.NewSession()
    ctx = memory.SetSession(ctx, sessionID)
    
    // Perform operations with context
    err = memory.AddMessage(ctx, "user", "Hello")
    if err != nil {
        log.Printf("Failed to add message: %v", err)
    }
    
    // Clean shutdown
    err = memory.ClearSession(ctx)
    if err != nil {
        log.Printf("Failed to clear session: %v", err)
    }
    
    err = memory.Close()
    if err != nil {
        log.Printf("Failed to close memory: %v", err)
    }
    
    fmt.Println("Graceful shutdown completed")
}
```

## Best Practices

### 1. Content Organization

```go
// Use consistent tagging strategies
const (
    TagFact        = "fact"
    TagPreference  = "preference"
    TagLearning    = "learning"
    TagTemporary   = "temporary"
    TagInteraction = "interaction"
)

func storeWithConsistentTags(ctx context.Context, memory core.Memory) error {
    // Store facts with consistent tags
    err := memory.Store(ctx, "Go supports concurrent programming", TagFact, "programming", "go")
    if err != nil {
        return err
    }
    
    // Store preferences with clear categorization
    err = memory.Store(ctx, "User prefers detailed code examples", TagPreference, "communication")
    if err != nil {
        return err
    }
    
    return nil
}

// Use meaningful key names for Remember/Recall
func useConsistentKeys(ctx context.Context, memory core.Memory) error {
    // Use namespaced keys for organization
    keys := map[string]any{
        "user.name":                "Alice",
        "user.expertise_level":     "intermediate",
        "preferences.style":        "detailed",
        "preferences.format":       "code_examples",
        "session.start_time":       time.Now(),
        "session.interaction_count": 0,
    }
    
    for key, value := range keys {
        err := memory.Remember(ctx, key, value)
        if err != nil {
            return fmt.Errorf("failed to remember %s: %w", key, err)
        }
    }
    
    return nil
}
```

### 2. Session Management Best Practices

```go
func implementSessionBestPractices(memory core.Memory) error {
    // Create meaningful session IDs
    userID := "user-123"
    timestamp := time.Now().Unix()
    sessionID := fmt.Sprintf("%s-%d", userID, timestamp)
    
    ctx := context.Background()
    ctx = memory.SetSession(ctx, sessionID)
    
    // Store session metadata
    sessionData := map[string]any{
        "session.id":         sessionID,
        "session.user_id":    userID,
        "session.start_time": time.Now(),
        "session.platform":   "web",
    }
    
    for key, value := range sessionData {
        memory.Remember(ctx, key, value)
    }
    
    // Use session consistently throughout conversation
    memory.AddMessage(ctx, "system", "Session started")
    
    return nil
}

func cleanupExpiredSessions(memory core.Memory, maxAge time.Duration) error {
    // This would typically be implemented by the memory provider
    // For demonstration, we show the pattern
    
    ctx := context.Background()
    
    // Get session start time
    startTime, err := memory.Recall(ctx, "session.start_time")
    if err != nil {
        return nil // No session data
    }
    
    if start, ok := startTime.(time.Time); ok {
        if time.Since(start) > maxAge {
            // Clear expired session
            err = memory.ClearSession(ctx)
            if err != nil {
                return fmt.Errorf("failed to clear expired session: %w", err)
            }
            fmt.Println("Cleared expired session")
        }
    }
    
    return nil
}
```

### 3. Error Handling and Resilience

```go
func robustMemoryOperations(ctx context.Context, memory core.Memory) error {
    // Implement retry logic for critical operations
    maxRetries := 3
    backoff := time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := memory.Store(ctx, "critical data", "important")
        if err == nil {
            return nil // Success
        }
        
        if attempt == maxRetries-1 {
            return fmt.Errorf("failed to store after %d attempts: %w", maxRetries, err)
        }
        
        log.Printf("Store attempt %d failed, retrying in %v: %v", 
            attempt+1, backoff, err)
        time.Sleep(backoff)
        backoff *= 2 // Exponential backoff
    }
    
    return nil
}

func safeMemoryQuery(ctx context.Context, memory core.Memory, query string) ([]core.Result, error) {
    // Query with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    results, err := memory.Query(ctx, query, 10)
    if err != nil {
        // Log error but don't fail completely
        log.Printf("Memory query failed: %v", err)
        return []core.Result{}, nil // Return empty results instead of error
    }
    
    return results, nil
}

func gracefulMemoryShutdown(memory core.Memory) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Clear any temporary data
    memory.ClearSession(ctx)
    
    // Close memory connection
    if err := memory.Close(); err != nil {
        log.Printf("Error closing memory: %v", err)
    } else {
        log.Println("Memory closed gracefully")
    }
}
```

### 4. Performance Optimization

```go
func optimizeMemoryUsage(ctx context.Context, memory core.Memory) error {
    // Batch operations when possible
    facts := []string{
        "Go was created at Google",
        "Go compiles to native code",
        "Go has garbage collection",
    }
    
    // Instead of individual Store calls, batch if possible
    for _, fact := range facts {
        // Use consistent tags for better organization
        err := memory.Store(ctx, fact, "fact", "go", "programming")
        if err != nil {
            log.Printf("Failed to store fact: %v", err)
            continue // Continue with other facts
        }
    }
    
    // Use appropriate limits for queries
    results, err := memory.Query(ctx, "Go programming", 5) // Limit to 5 results
    if err != nil {
        return err
    }
    
    // Process results efficiently
    for _, result := range results {
        if result.Score > 0.7 { // Only process high-confidence results
            fmt.Printf("High confidence result: %s\n", result.Content)
        }
    }
    
    return nil
}
```

## Common Patterns

### 1. Context-Aware Responses

```go
func generateContextAwareResponse(memory core.Memory, sessionID, message string) (string, error) {
    ctx := context.Background()
    
    // Set session context
    ctx = memory.SetSession(ctx, sessionID)
    
    // Get user preferences using key-value storage
    commStyle, err := memory.Recall(ctx, "communication_style")
    if err != nil {
        // Default if not found
        commStyle = "standard"
    }
    
    expertiseLevel, err := memory.Recall(ctx, "expertise_level")
    if err != nil {
        expertiseLevel = "beginner"
    }
    
    // Get relevant facts using semantic search
    facts, err := memory.Query(ctx, message, 3)
    if err != nil {
        return "", fmt.Errorf("failed to query facts: %w", err)
    }
    
    // Build context-aware prompt
    prompt := buildPromptWithContext(message, commStyle, expertiseLevel, facts)
    
    // Generate response (would use LLM here)
    return generateResponse(prompt), nil
}

func buildPromptWithContext(message string, commStyle, expertiseLevel any, facts []core.Result) string {
    var builder strings.Builder
    
    builder.WriteString(fmt.Sprintf("Communication style: %v\n", commStyle))
    builder.WriteString(fmt.Sprintf("User expertise: %v\n", expertiseLevel))
    
    if len(facts) > 0 {
        builder.WriteString("\nRelevant information:\n")
        for _, fact := range facts {
            builder.WriteString(fmt.Sprintf("- %s\n", fact.Content))
        }
    }
    
    builder.WriteString(fmt.Sprintf("\nUser message: %s\n", message))
    builder.WriteString("Please provide a helpful response based on the context above.")
    
    return builder.String()
}

func generateResponse(prompt string) string {
    // Mock response generation - in real implementation, use LLM
    return "Generated response based on context"
}
```

### 2. Learning from Interactions

```go
func learnFromInteraction(memory core.Memory, sessionID, userMessage, response string, feedback string) error {
    ctx := context.Background()
    
    // Set session context
    ctx = memory.SetSession(ctx, sessionID)
    
    // Store the interaction in chat history (already done by agent)
    // But we can store additional learning data
    
    interaction := fmt.Sprintf("User asked: %s | Response quality: %s", 
        userMessage, feedback)
    
    err := memory.Store(ctx, interaction, "interaction", "learning", "feedback")
    if err != nil {
        return fmt.Errorf("failed to store interaction: %w", err)
    }
    
    // Extract learnings if feedback is positive
    if feedback == "helpful" || feedback == "correct" {
        // Store successful patterns
        pattern := extractPattern(userMessage, response)
        err = memory.Store(ctx, pattern, "successful-pattern", "learning")
        if err != nil {
            return fmt.Errorf("failed to store pattern: %w", err)
        }
        
        // Update user preferences based on successful interactions
        if strings.Contains(strings.ToLower(userMessage), "explain") {
            memory.Remember(ctx, "prefers_explanations", true)
        }
        
        if strings.Contains(strings.ToLower(userMessage), "example") {
            memory.Remember(ctx, "prefers_examples", true)
        }
    }
    
    return nil
}

func extractPattern(userMessage, response string) string {
    // Simple pattern extraction - in real implementation, use more sophisticated analysis
    return fmt.Sprintf("Pattern: %s -> %s", 
        strings.ToLower(userMessage[:min(50, len(userMessage))]), 
        strings.ToLower(response[:min(50, len(response))]))
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

## Conclusion

Basic memory operations provide the foundation for building intelligent agents that can remember, learn, and improve over time. The key concepts covered include:

- Setting up in-memory storage
- Storing and retrieving information
- Managing conversation history
- Building context-aware responses
- Memory management and cleanup

These fundamentals prepare you for more advanced memory systems including vector databases and RAG implementations.

## Next Steps

::: tip Learning Path
Follow the progressive tutorial series to master memory systems:
:::

### 🗄️ **Next**: [Vector Databases](vector-databases.md)
Set up production-ready storage with pgvector or Weaviate

### 📄 **Then**: [Document Ingestion](document-ingestion.md)
Learn document processing and knowledge base creation

### 🧠 **Advanced**: [RAG Implementation](rag-implementation.md)
Build retrieval-augmented generation systems

### 🏗️ **Scale**: [Knowledge Bases](knowledge-bases.md)
Create enterprise-scale knowledge systems

### ⚡ **Optimize**: [Memory Optimization](memory-optimization.md)
Performance tuning and scaling strategies

## Related Resources

::: info Additional Learning
- **[Memory Systems Overview](README.md)** - Complete architecture guide
- **[Core Concepts](../core-concepts/README.md)** - AgenticGoKit fundamentals
- **[API Reference](../../reference/api/memory.md)** - Complete Memory interface documentation
- **[Examples](../../examples/memory/)** - Working code examples
:::

## Further Reading

- [Memory Interface API Reference](../../reference/api/memory.md#memory-interface)
- [Configuration Guide](../../reference/configuration.md#memory-configuration)
- [Best Practices](../../reference/best-practices/memory.md)
