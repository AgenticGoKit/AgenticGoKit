# Basic Memory Operations in AgenticGoKit

## Overview

This tutorial covers the fundamentals of memory operations in AgenticGoKit, including storing information, retrieving data, and managing conversation history. We'll start with simple in-memory storage and progress to more advanced concepts.

Basic memory operations form the foundation for all memory-enabled agents, providing the essential building blocks for more sophisticated memory systems.

## Prerequisites

- Understanding of [Core Concepts](../core-concepts/README.md)
- Basic knowledge of Go programming
- Familiarity with key-value storage concepts

## Setting Up Basic Memory

### 1. In-Memory Storage

The simplest memory provider stores data in RAM:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create in-memory storage
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider:   "memory",
        Connection: "memory",
        MaxResults: 10,
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    
    ctx := context.Background()
    
    // Store some information with metadata
    err = memory.Store(ctx, "Paris is the capital of France", map[string]interface{}{
        "type":     "fact",
        "category": "geography",
        "source":   "general_knowledge",
    })
    if err != nil {
        log.Fatalf("Failed to store: %v", err)
    }
    
    // Search for information
    results, err := memory.Search(ctx, "capital France", core.WithLimit(5))
    if err != nil {
        log.Fatalf("Failed to search: %v", err)
    }
    
    for _, result := range results {
        fmt.Printf("Found: %s (Score: %.3f)\n", result.Content, result.Score)
        if result.Metadata != nil {
            fmt.Printf("  Type: %v, Category: %v\n", 
                result.Metadata["type"], result.Metadata["category"])
        }
    }
}
```

### 2. Memory Configuration Options

```go
// Basic configuration
config := core.AgentMemoryConfig{
    Provider:   "memory",
    Connection: "memory",
    MaxResults: 10,
    Dimensions: 1536,
    AutoEmbed:  true,
    
    // Enable knowledge base features
    EnableKnowledgeBase:     true,
    KnowledgeMaxResults:     20,
    KnowledgeScoreThreshold: 0.7,
    
    // Document processing settings
    Documents: core.DocumentConfig{
        AutoChunk:                true,
        SupportedTypes:           []string{"txt", "md", "pdf"},
        MaxFileSize:              "10MB",
        EnableMetadataExtraction: true,
    },
    
    // Embedding configuration
    Embedding: core.EmbeddingConfig{
        Provider:        "dummy", // Use "openai" for production
        Model:           "text-embedding-3-small",
        CacheEmbeddings: true,
        MaxBatchSize:    100,
        TimeoutSeconds:  30,
    },
}

memory, err := core.NewMemory(config)
```

## Basic Storage Operations

### 1. Storing Simple Information

```go
func storeBasicInformation(memory core.Memory) error {
    ctx := context.Background()
    
    // Store facts
    facts := []struct {
        content     string
        contentType string
    }{
        {"The Earth orbits the Sun", "scientific-fact"},
        {"Water boils at 100°C at sea level", "scientific-fact"},
        {"Shakespeare wrote Hamlet", "literary-fact"},
        {"The Great Wall of China is visible from space", "myth"}, // Actually false!
    }
    
    for _, fact := range facts {
        err := memory.Store(ctx, fact.content, fact.contentType)
        if err != nil {
            return fmt.Errorf("failed to store fact: %w", err)
        }
    }
    
    fmt.Println("Stored", len(facts), "facts")
    return nil
}
```

### 2. Storing with Metadata

```go
func storeWithMetadata(memory core.Memory) error {
    ctx := context.Background()
    
    // Store with additional metadata
    err := memory.Store(ctx,
        "The user prefers detailed technical explanations",
        "user-preference",
        core.WithMetadata(map[string]string{
            "user_id":   "user-123",
            "category":  "communication-style",
            "priority":  "high",
            "source":    "conversation-analysis",
        }),
        core.WithSession("session-456"),
        core.WithTimestamp(time.Now()),
    )
    
    if err != nil {
        return fmt.Errorf("failed to store with metadata: %w", err)
    }
    
    return nil
}
```

### 3. Storing Structured Data

```go
type UserProfile struct {
    Name        string   `json:"name"`
    Interests   []string `json:"interests"`
    Preferences struct {
        Language string `json:"language"`
        Style    string `json:"style"`
    } `json:"preferences"`
}

func storeStructuredData(memory core.Memory) error {
    ctx := context.Background()
    
    profile := UserProfile{
        Name:      "Alice Johnson",
        Interests: []string{"AI", "Machine Learning", "Go Programming"},
    }
    profile.Preferences.Language = "English"
    profile.Preferences.Style = "Technical"
    
    // Store structured data
    err := memory.StoreStructured(ctx, profile,
        core.WithContentType("user-profile"),
        core.WithSession("user-123"),
        core.WithMetadata(map[string]string{
            "version": "1.0",
            "source":  "onboarding",
        }),
    )
    
    if err != nil {
        return fmt.Errorf("failed to store structured data: %w", err)
    }
    
    return nil
}
```

## Basic Retrieval Operations

### 1. Simple Search

```go
func performBasicSearch(memory core.Memory) error {
    ctx := context.Background()
    
    // Simple text search
    results, err := memory.Search(ctx, "Earth Sun orbit")
    if err != nil {
        return fmt.Errorf("search failed: %w", err)
    }
    
    fmt.Printf("Found %d results:\n", len(results))
    for i, result := range results {
        fmt.Printf("%d. %s (Score: %.3f)\n", 
            i+1, result.Content, result.Score)
    }
    
    return nil
}
```

### 2. Search with Filters

```go
func performFilteredSearch(memory core.Memory) error {
    ctx := context.Background()
    
    // Search with various filters
    results, err := memory.Search(ctx, "technical explanation",
        core.WithLimit(5),                    // Limit results
        core.WithScoreThreshold(0.7),         // Minimum relevance score
        core.WithContentType("user-preference"), // Filter by content type
        core.WithSession("session-456"),      // Filter by session
        core.WithMetadataFilter(map[string]string{
            "priority": "high",
        }),
    )
    
    if err != nil {
        return fmt.Errorf("filtered search failed: %w", err)
    }
    
    fmt.Printf("Filtered search found %d results:\n", len(results))
    for _, result := range results {
        fmt.Printf("- %s\n", result.Content)
        fmt.Printf("  Type: %s, Score: %.3f\n", 
            result.ContentType, result.Score)
        fmt.Printf("  Metadata: %+v\n", result.Metadata)
    }
    
    return nil
}
```

### 3. Retrieving by ID

```go
func retrieveById(memory core.Memory, id string) error {
    ctx := context.Background()
    
    // Get specific item by ID
    result, err := memory.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to retrieve by ID: %w", err)
    }
    
    fmt.Printf("Retrieved item:\n")
    fmt.Printf("ID: %s\n", result.ID)
    fmt.Printf("Content: %s\n", result.Content)
    fmt.Printf("Type: %s\n", result.ContentType)
    fmt.Printf("Created: %s\n", result.CreatedAt.Format(time.RFC3339))
    
    return nil
}
```

## Conversation History Management

### 1. Storing Conversation Messages

```go
func storeConversation(memory core.Memory, sessionID string) error {
    ctx := context.Background()
    
    // Simulate a conversation
    conversation := []struct {
        role    string
        content string
    }{
        {"user", "What is machine learning?"},
        {"assistant", "Machine learning is a subset of AI that enables computers to learn from data..."},
        {"user", "Can you give me an example?"},
        {"assistant", "Sure! A common example is email spam detection..."},
        {"user", "How does it work technically?"},
        {"assistant", "Technically, spam detection uses features like sender reputation, keywords..."},
    }
    
    for i, msg := range conversation {
        err := memory.Store(ctx, msg.content, msg.role+"-message",
            core.WithSession(sessionID),
            core.WithTimestamp(time.Now().Add(time.Duration(i)*time.Minute)),
            core.WithMetadata(map[string]string{
                "role":     msg.role,
                "sequence": fmt.Sprintf("%d", i+1),
            }),
        )
        if err != nil {
            return fmt.Errorf("failed to store message: %w", err)
        }
    }
    
    return nil
}
```

### 2. Retrieving Conversation History

```go
func getConversationHistory(memory core.Memory, sessionID string) error {
    ctx := context.Background()
    
    // Get recent conversation history
    messages, err := memory.GetHistory(ctx, 10,
        core.WithSession(sessionID),
        core.WithTimeRange(
            time.Now().Add(-24*time.Hour), // Last 24 hours
            time.Now(),
        ),
    )
    if err != nil {
        return fmt.Errorf("failed to get history: %w", err)
    }
    
    fmt.Printf("Conversation history (%d messages):\n", len(messages))
    for _, msg := range messages {
        fmt.Printf("[%s] %s: %s\n",
            msg.Timestamp.Format("15:04"),
            msg.Role,
            msg.Content,
        )
    }
    
    return nil
}
```

### 3. Conversation Context Building

```go
func buildConversationContext(memory core.Memory, sessionID string, currentMessage string) (string, error) {
    ctx := context.Background()
    
    // Get recent history for context
    history, err := memory.GetHistory(ctx, 5,
        core.WithSession(sessionID),
    )
    if err != nil {
        return "", fmt.Errorf("failed to get history: %w", err)
    }
    
    // Build context string
    var contextBuilder strings.Builder
    contextBuilder.WriteString("Recent conversation:\n")
    
    for _, msg := range history {
        contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", 
            strings.Title(msg.Role), msg.Content))
    }
    
    contextBuilder.WriteString(fmt.Sprintf("\nCurrent message: %s\n", currentMessage))
    
    return contextBuilder.String(), nil
}
```

## Memory-Enabled Agent Example

### 1. Simple Memory Agent

```go
type MemoryAgent struct {
    name   string
    memory core.Memory
    llm    core.LLMProvider
}

func NewMemoryAgent(name string, memory core.Memory, llm core.LLMProvider) *MemoryAgent {
    return &MemoryAgent{
        name:   name,
        memory: memory,
        llm:    llm,
    }
}

func (m *MemoryAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Extract message from state
    message, ok := state.Get("message")
    if !ok {
        return core.AgentResult{}, errors.New("no message in state")
    }
    messageStr := message.(string)
    
    sessionID := event.GetSessionID()
    
    // Store the user message
    err := m.memory.Store(ctx, messageStr, "user-message",
        core.WithSession(sessionID),
        core.WithTimestamp(time.Now()),
    )
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to store message: %w", err)
    }
    
    // Get conversation context
    context, err := m.buildContext(ctx, sessionID, messageStr)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to build context: %w", err)
    }
    
    // Generate response with context
    response, err := m.llm.Generate(ctx, context)
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to generate response: %w", err)
    }
    
    // Store the assistant response
    err = m.memory.Store(ctx, response, "assistant-message",
        core.WithSession(sessionID),
        core.WithTimestamp(time.Now()),
    )
    if err != nil {
        return core.AgentResult{}, fmt.Errorf("failed to store response: %w", err)
    }
    
    // Return result
    outputState := state.Clone()
    outputState.Set("response", response)
    
    return core.AgentResult{OutputState: outputState}, nil
}

func (m *MemoryAgent) buildContext(ctx context.Context, sessionID, currentMessage string) (string, error) {
    // Get recent conversation history
    history, err := m.memory.GetHistory(ctx, 5,
        core.WithSession(sessionID),
    )
    if err != nil {
        return "", err
    }
    
    // Search for relevant information
    relevant, err := m.memory.Search(ctx, currentMessage,
        core.WithLimit(3),
        core.WithScoreThreshold(0.7),
        core.WithContentType("fact"),
    )
    if err != nil {
        return "", err
    }
    
    // Build enhanced context
    var contextBuilder strings.Builder
    
    // Add relevant facts
    if len(relevant) > 0 {
        contextBuilder.WriteString("Relevant information:\n")
        for _, item := range relevant {
            contextBuilder.WriteString(fmt.Sprintf("- %s\n", item.Content))
        }
        contextBuilder.WriteString("\n")
    }
    
    // Add conversation history
    if len(history) > 0 {
        contextBuilder.WriteString("Recent conversation:\n")
        for _, msg := range history {
            contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", 
                strings.Title(msg.Role), msg.Content))
        }
        contextBuilder.WriteString("\n")
    }
    
    contextBuilder.WriteString(fmt.Sprintf("Current question: %s\n", currentMessage))
    contextBuilder.WriteString("Please provide a helpful response based on the context above.")
    
    return contextBuilder.String(), nil
}
```

### 2. Using the Memory Agent

```go
func main() {
    // Create memory system
    memory, err := core.NewMemory(core.AgentMemoryConfig{
        Provider: "memory",
        MaxSize:  1000,
    })
    if err != nil {
        log.Fatalf("Failed to create memory: %v", err)
    }
    
    // Create LLM provider
    llm, err := core.NewOpenAIAdapter(
        os.Getenv("OPENAI_API_KEY"),
        "gpt-3.5-turbo",
        1000,
        0.7,
    )
    if err != nil {
        log.Fatalf("Failed to create LLM: %v", err)
    }
    
    // Create memory-enabled agent
    agent := NewMemoryAgent("memory-assistant", memory, llm)
    
    // Store some initial facts
    ctx := context.Background()
    facts := []string{
        "Go is a programming language developed by Google",
        "AgenticGoKit is a framework for building multi-agent systems in Go",
        "Vector databases are used for similarity search",
    }
    
    for _, fact := range facts {
        memory.Store(ctx, fact, "fact")
    }
    
    // Create runner and register agent
    runner := core.NewRunner(100)
    orchestrator := core.NewRouteOrchestrator(runner.GetCallbackRegistry())
    runner.SetOrchestrator(orchestrator)
    
    agentHandler := core.ConvertAgentToHandler(agent)
    runner.RegisterAgent("memory-assistant", agentHandler)
    
    // Start runner
    runner.Start(ctx)
    defer runner.Stop()
    
    // Simulate conversation
    sessionID := "user-session-123"
    
    messages := []string{
        "What is Go?",
        "How is it related to AgenticGoKit?",
        "What are vector databases used for?",
    }
    
    for _, msg := range messages {
        event := core.NewEvent(
            "memory-assistant",
            core.EventData{"message": msg},
            map[string]string{
                "session_id": sessionID,
                "route":      "memory-assistant",
            },
        )
        
        runner.Emit(event)
        time.Sleep(2 * time.Second) // Wait for processing
    }
}
```

## Memory Management

### 1. Updating Stored Information

```go
func updateMemoryContent(memory core.Memory) error {
    ctx := context.Background()
    
    // First, find the item to update
    results, err := memory.Search(ctx, "Great Wall China visible space")
    if err != nil {
        return err
    }
    
    if len(results) > 0 {
        // Update the incorrect information
        err = memory.Update(ctx, results[0].ID,
            "The Great Wall of China is NOT visible from space with the naked eye",
            core.WithMetadata(map[string]string{
                "corrected": "true",
                "updated_at": time.Now().Format(time.RFC3339),
            }),
        )
        if err != nil {
            return fmt.Errorf("failed to update: %w", err)
        }
        
        fmt.Println("Corrected misinformation about the Great Wall")
    }
    
    return nil
}
```

### 2. Deleting Information

```go
func cleanupOldMemories(memory core.Memory) error {
    ctx := context.Background()
    
    // Get memory statistics
    stats, err := memory.GetStats(ctx)
    if err != nil {
        return err
    }
    
    fmt.Printf("Memory stats: %d items, %d MB used\n", 
        stats.ItemCount, stats.SizeBytes/1024/1024)
    
    // Delete specific items
    results, err := memory.Search(ctx, "temporary data",
        core.WithContentType("temporary"),
    )
    if err != nil {
        return err
    }
    
    for _, result := range results {
        err = memory.Delete(ctx, result.ID)
        if err != nil {
            fmt.Printf("Failed to delete %s: %v\n", result.ID, err)
        } else {
            fmt.Printf("Deleted temporary item: %s\n", result.ID)
        }
    }
    
    return nil
}
```

### 3. Memory Cleanup Strategies

```go
type MemoryManager struct {
    memory     core.Memory
    maxAge     time.Duration
    maxItems   int
    cleanupInterval time.Duration
}

func NewMemoryManager(memory core.Memory) *MemoryManager {
    return &MemoryManager{
        memory:          memory,
        maxAge:          24 * time.Hour,
        maxItems:        1000,
        cleanupInterval: time.Hour,
    }
}

func (mm *MemoryManager) StartCleanup(ctx context.Context) {
    ticker := time.NewTicker(mm.cleanupInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            mm.performCleanup(ctx)
        }
    }
}

func (mm *MemoryManager) performCleanup(ctx context.Context) {
    // Get memory statistics
    stats, err := mm.memory.GetStats(ctx)
    if err != nil {
        fmt.Printf("Failed to get memory stats: %v\n", err)
        return
    }
    
    // Clean up old items if over limit
    if stats.ItemCount > mm.maxItems {
        mm.cleanupOldItems(ctx, stats.ItemCount-mm.maxItems)
    }
    
    // Clean up expired items
    mm.cleanupExpiredItems(ctx)
}

func (mm *MemoryManager) cleanupOldItems(ctx context.Context, itemsToRemove int) {
    // Implementation would search for oldest items and remove them
    fmt.Printf("Cleaning up %d old items\n", itemsToRemove)
}

func (mm *MemoryManager) cleanupExpiredItems(ctx context.Context) {
    // Implementation would find and remove expired items
    cutoff := time.Now().Add(-mm.maxAge)
    fmt.Printf("Cleaning up items older than %s\n", cutoff.Format(time.RFC3339))
}
```

## Best Practices for Basic Memory

### 1. Content Organization

```go
// Use consistent content types
const (
    ContentTypeFact        = "fact"
    ContentTypePreference  = "user-preference"
    ContentTypeMessage     = "message"
    ContentTypeKnowledge   = "knowledge"
    ContentTypeTemporary   = "temporary"
)

// Use structured metadata
func storeWithConsistentMetadata(memory core.Memory, content, contentType string) error {
    return memory.Store(context.Background(), content, contentType,
        core.WithMetadata(map[string]string{
            "version":    "1.0",
            "source":     "user-input",
            "confidence": "high",
            "language":   "en",
        }),
        core.WithTimestamp(time.Now()),
    )
}
```

### 2. Session Management

```go
func manageUserSessions(memory core.Memory) {
    // Use consistent session IDs
    sessionID := fmt.Sprintf("user-%s-%d", userID, time.Now().Unix())
    
    // Store session metadata
    memory.Store(context.Background(),
        "Session started",
        "session-event",
        core.WithSession(sessionID),
        core.WithMetadata(map[string]string{
            "event_type": "session_start",
            "user_id":    userID,
            "ip_address": clientIP,
        }),
    )
}
```

### 3. Error Handling

```go
func robustMemoryOperations(memory core.Memory) {
    ctx := context.Background()
    
    // Store with retry logic
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        err := memory.Store(ctx, "important data", "critical")
        if err == nil {
            break
        }
        
        if i == maxRetries-1 {
            log.Printf("Failed to store after %d attempts: %v", maxRetries, err)
            // Handle permanent failure
        } else {
            log.Printf("Store attempt %d failed, retrying: %v", i+1, err)
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }
    
    // Search with fallback
    results, err := memory.Search(ctx, "query")
    if err != nil {
        log.Printf("Search failed, using fallback: %v", err)
        // Use cached results or default response
        results = getFallbackResults()
    }
}
```

## Common Patterns

### 1. Context-Aware Responses

```go
func generateContextAwareResponse(memory core.Memory, sessionID, message string) (string, error) {
    ctx := context.Background()
    
    // Get user preferences
    preferences, err := memory.Search(ctx, "user preference",
        core.WithSession(sessionID),
        core.WithContentType("user-preference"),
    )
    if err != nil {
        return "", err
    }
    
    // Get relevant facts
    facts, err := memory.Search(ctx, message,
        core.WithContentType("fact"),
        core.WithLimit(3),
    )
    if err != nil {
        return "", err
    }
    
    // Build context-aware prompt
    prompt := buildPromptWithContext(message, preferences, facts)
    
    // Generate response (would use LLM here)
    return generateResponse(prompt), nil
}
```

### 2. Learning from Interactions

```go
func learnFromInteraction(memory core.Memory, sessionID, userMessage, response string, feedback string) error {
    ctx := context.Background()
    
    // Store the interaction
    interaction := fmt.Sprintf("Q: %s\nA: %s\nFeedback: %s", 
        userMessage, response, feedback)
    
    err := memory.Store(ctx, interaction, "interaction",
        core.WithSession(sessionID),
        core.WithMetadata(map[string]string{
            "feedback_type": classifyFeedback(feedback),
            "quality":      scoreFeedback(feedback),
        }),
    )
    
    if err != nil {
        return err
    }
    
    // Extract learnings if feedback is positive
    if feedback == "helpful" || feedback == "correct" {
        // Store successful patterns
        pattern := extractPattern(userMessage, response)
        memory.Store(ctx, pattern, "successful-pattern",
            core.WithMetadata(map[string]string{
                "pattern_type": "response-pattern",
                "success_rate": "high",
            }),
        )
    }
    
    return nil
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

- [Vector Databases](vector-databases.md) - Learn about production-ready storage
- [RAG Implementation](rag-implementation.md) - Build retrieval-augmented systems
- [Knowledge Bases](knowledge-bases.md) - Create comprehensive knowledge systems
- [Memory Optimization](memory-optimization.md) - Optimize performance and scaling

## Further Reading

- [API Reference: Memory Interface](../../reference/api/agent.md#memory)
- [Examples: Basic Memory Usage](../../examples/)
- [Configuration Guide: Memory Settings](../../reference/api/configuration.md)