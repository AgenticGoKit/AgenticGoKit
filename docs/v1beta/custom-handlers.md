# Custom Handlers

Learn how to create custom agent logic with handlers that give you full control over agent behavior while still leveraging LLM, tools, and memory capabilities.

---

## 🎯 Overview

Custom handlers in v1beta allow you to:

- **Control Execution Flow** - Define exactly how your agent processes input
- **Access Capabilities** - Use LLM, tools, and memory within your logic
- **Compose Handlers** - Chain, parallelize, and combine handler logic
- **Add Middleware** - Apply cross-cutting concerns like logging and retries

---

## 🚀 Quick Start

### Basic Handler

```go
package main

import (
    "context"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Define custom handler
    handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
        return "You said: " + input, nil
    }
    
    // Create agent with custom handler
    agent, _ := v1beta.NewBuilder("CustomAgent").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        WithHandler(handler).
        Build()
    
    result, _ := agent.Run(context.Background(), "Hello")
    // Output: "You said: Hello"
}
```

---

## 📋 Handler Signature

### HandlerFunc Type

```go
type HandlerFunc func(ctx context.Context, input string, capabilities *Capabilities) (string, error)
```

**Parameters:**
- `ctx` - Context for cancellation and deadlines
- `input` - User input string
- `capabilities` - Access to LLM, tools, and memory

**Returns:**
- `string` - Agent response
- `error` - Error if processing fails

### Capabilities Structure

```go
type Capabilities struct {
    LLM    func(system, user string) (string, error)
    Tools  ToolManager
    Memory Memory
}
```

**Fields:**
- `LLM` - Function to call the configured language model
- `Tools` - Interface to discover and execute tools
- `Memory` - Interface to store and query memory

---

## 🎨 Handler Patterns

### Pattern 1: LLM-Only Handler

Direct LLM calls with custom prompts:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    systemPrompt := "You are a helpful coding assistant specializing in Go."
    return capabilities.LLM(systemPrompt, input)
}

agent, _ := v1beta.NewBuilder("CodingAssistant").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithHandler(handler).
    Build()
```

### Pattern 2: Tool-Augmented Handler

Execute tools based on input analysis:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Check if input requires calculation
    if strings.Contains(strings.ToLower(input), "calculate") {
        // Execute calculator tool
        result, err := capabilities.Tools.Execute(ctx, "calculator", map[string]interface{}{
            "expression": extractExpression(input),
        })
        if err != nil {
            return "", err
        }
        
        return fmt.Sprintf("The result is: %v", result.Content), nil
    }
    
    // Otherwise use LLM
    return capabilities.LLM("You are a helpful assistant.", input)
}
```

### Pattern 3: Memory-Aware Handler

Store and retrieve context from memory:

```go
import "github.com/agenticgokit/agenticgokit/core"

handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    if capabilities.Memory == nil {
        return capabilities.LLM("You are a helpful assistant.", input)
    }
    
    // Query relevant memories
    memories, err := capabilities.Memory.Query(ctx, input, 5)
    if err != nil {
        return "", err
    }
    
    // Build context from memories
    var context strings.Builder
    for _, mem := range memories {
        context.WriteString(fmt.Sprintf("- %s\n", mem.Content))
    }
    
    // Call LLM with memory context
    systemPrompt := fmt.Sprintf("You are a helpful assistant. Context:\n%s", context.String())
    response, err := capabilities.LLM(systemPrompt, input)
    if err != nil {
        return "", err
    }
    
    // Store interaction in memory
    interaction := fmt.Sprintf("User: %s\nAssistant: %s", input, response)
    capabilities.Memory.Store(ctx, interaction)
    
    return response, nil
}
```

### Pattern 4: Multi-Step Processing

Chain multiple processing steps:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Step 1: Analyze intent
    intent, err := capabilities.LLM(
        "Classify the user's intent: question, command, or statement",
        input,
    )
    if err != nil {
        return "", err
    }
    
    // Step 2: Process based on intent
    var response string
    switch strings.TrimSpace(intent) {
    case "question":
        response, err = handleQuestion(ctx, input, capabilities)
    case "command":
        response, err = handleCommand(ctx, input, capabilities)
    default:
        response, err = handleStatement(ctx, input, capabilities)
    }
    
    return response, err
}
```

### Pattern 5: Hybrid Tool + LLM

Intelligently route between tools and LLM:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // First, ask LLM if tools are needed
    decision, err := capabilities.LLM(
        "You are a routing assistant. Reply ONLY 'TOOL:toolname' if a tool is needed, or 'LLM' otherwise.",
        input,
    )
    if err != nil {
        return "", err
    }
    
    // Route based on decision
    if strings.HasPrefix(decision, "TOOL:") {
        toolName := strings.TrimPrefix(decision, "TOOL:")
        result, err := capabilities.Tools.Execute(ctx, strings.TrimSpace(toolName), map[string]interface{}{
            "query": input,
        })
        if err != nil {
            return "", err
        }
        return fmt.Sprintf("%v", result.Content), nil
    }
    
    // Use LLM for general queries
    return capabilities.LLM("You are a helpful assistant.", input)
}
```

---

## 🔧 Handler Augmentation

v1beta provides pre-built augmentations to enhance handlers:

### WithToolAugmentation

Automatically adds tool-calling capability:

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

// Base handler
baseHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("You are a helpful assistant.", input)
}

// Augment with automatic tool discovery and calling
handler := v1beta.WithToolAugmentation(baseHandler)

agent, _ := v1beta.NewBuilder("ToolAgent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithTools(
        v1beta.WithMCP(mcpServers...),
    ).
    WithHandler(handler).
    Build()
```

### WithMemoryAugmentation

Automatically adds memory storage and retrieval:

```go
// Base handler
baseHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("You are a helpful assistant.", input)
}

// Augment with automatic memory integration
handler := v1beta.WithMemoryAugmentation(baseHandler)

agent, _ := v1beta.NewBuilder("MemoryAgent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
    ).
    WithHandler(handler).
    Build()
```

### WithRAGAugmentation

Automatically adds RAG knowledge retrieval:

```go
// Base handler
baseHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("Answer based on the provided knowledge.", input)
}

// Augment with RAG - retrieves top 5 relevant documents
handler := v1beta.WithRAGAugmentation(baseHandler, "knowledge_base", 5)

agent, _ := v1beta.NewBuilder("RAGAgent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithRAG(4000, 0.3, 0.7),
    ).
    WithHandler(handler).
    Build()
```

### WithLLMAugmentation

Adds retry logic and error handling to LLM calls:

```go
baseHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("You are a helpful assistant.", input)
}

// Augment with retry logic (max 3 retries)
handler := v1beta.WithLLMAugmentation(baseHandler, 3)
```

---

## 🔗 Handler Composition

Combine multiple handlers using composition functions:

### Chain

Execute handlers in sequence:

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

preprocessHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return strings.ToLower(input), nil
}

processHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("You are a helpful assistant.", input)
}

postprocessHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return strings.ToUpper(input), nil
}

// Chain handlers: preprocess -> process -> postprocess
handler := v1beta.Chain(preprocessHandler, processHandler, postprocessHandler)

agent, _ := v1beta.NewBuilder("ChainedAgent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithHandler(handler).
    Build()
```

### ParallelHandlers

Run handlers in parallel and combine results:

```go
summaryHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("Summarize this in one sentence.", input)
}

keywordsHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("Extract 5 keywords.", input)
}

sentimentHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("Analyze sentiment: positive, negative, or neutral.", input)
}

// Run all handlers in parallel, combine with separator
handler := v1beta.ParallelHandlers("\n---\n", summaryHandler, keywordsHandler, sentimentHandler)

agent, _ := v1beta.NewBuilder("AnalysisAgent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    WithHandler(handler).
    Build()
```

### Conditional

Execute handler only if condition is met:

```go
isQuestion := func(ctx context.Context, input string) bool {
    return strings.HasSuffix(strings.TrimSpace(input), "?")
}

questionHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("You are an expert at answering questions.", input)
}

// Only run questionHandler if input is a question
handler := v1beta.Conditional(isQuestion, questionHandler)
```

### Fallback

Try primary handler, fall back if it fails:

```go
primaryHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Try using expensive GPT-4
    return capabilities.LLM("You are a helpful assistant.", input)
}

fallbackHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Fall back to cheaper GPT-3.5
    // In practice, you'd need to switch models here
    return "I'm experiencing high load. Here's a basic response: " + input, nil
}

handler := v1beta.Fallback(primaryHandler, fallbackHandler)
```

### Retry

Add retry logic to handlers:

```go
unreliableHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Handler that might fail occasionally
    return capabilities.LLM("You are a helpful assistant.", input)
}

// Retry up to 3 times with exponential backoff
handler := v1beta.Retry(unreliableHandler, 3)
```

### WithTimeout

Add timeout to handler execution:

```go
slowHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("You are a helpful assistant.", input)
}

// Timeout after 30 seconds
handler := v1beta.WithTimeout(slowHandler, 30*time.Second)
```

### WithLogging

Add logging to handlers:

```go
import "log"

businessLogicHandler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    return capabilities.LLM("You are a helpful assistant.", input)
}

// Add logging
handler := v1beta.WithLogging(businessLogicHandler, func(format string, args ...interface{}) {
    log.Printf(format, args...)
})
```

---

## 🎯 Complete Examples

### Example 1: Research Assistant

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "github.com/agenticgokit/agenticgokit/v1beta"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func createResearchAssistant() (v1beta.Agent, error) {
    handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
        // Step 1: Analyze if web search is needed
        decision, err := capabilities.LLM(
            "Determine if this query requires web search. Reply 'YES' or 'NO'.",
            input,
        )
        if err != nil {
            return "", err
        }
        
        var context string
        if strings.TrimSpace(decision) == "YES" {
            // Step 2: Execute web search tool
            searchResult, err := capabilities.Tools.Execute(ctx, "web_search", map[string]interface{}{
                "query": input,
            })
            if err == nil {
                context = fmt.Sprintf("Search results: %v", searchResult.Content)
            }
        }
        
        // Step 3: Generate response with context
        systemPrompt := "You are a research assistant. Use the provided context to answer."
        if context != "" {
            systemPrompt += "\n\nContext: " + context
        }
        
        response, err := capabilities.LLM(systemPrompt, input)
        if err != nil {
            return "", err
        }
        
        // Step 4: Store in memory for future reference
        if capabilities.Memory != nil {
            capabilities.Memory.Store(ctx, fmt.Sprintf("Q: %s\nA: %s", input, response))
        }
        
        return response, nil
    }
    
    return v1beta.NewBuilder("ResearchAssistant").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        WithTools(
            v1beta.WithMCP(/* web search server */),
        ).
        WithMemory(
            v1beta.WithMemoryProvider("memory"),
        ).
        WithHandler(handler).
        Build()
}
```

### Example 2: Code Review Agent

```go
func createCodeReviewAgent() (v1beta.Agent, error) {
    handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
        // Parallel analysis
        analysisHandlers := v1beta.ParallelHandlers("\n\n",
            // Security analysis
            func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
                return capabilities.LLM(
                    "Analyze this code for security vulnerabilities.",
                    input,
                )
            },
            // Performance analysis
            func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
                return capabilities.LLM(
                    "Analyze this code for performance issues.",
                    input,
                )
            },
            // Best practices
            func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
                return capabilities.LLM(
                    "Review this code for best practices and style.",
                    input,
                )
            },
        )
        
        // Run parallel analysis with timeout and retry
        robustHandler := v1beta.WithTimeout(
            v1beta.Retry(analysisHandlers, 2),
            60*time.Second,
        )
        
        return robustHandler(ctx, input, capabilities)
    }
    
    return v1beta.NewBuilder("CodeReviewer").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        WithHandler(handler).
        Build()
}
```

### Example 3: Customer Support Bot

```go
func createSupportBot() (v1beta.Agent, error) {
    handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
        // Check memory for previous interactions
        var customerHistory string
        if capabilities.Memory != nil {
            memories, err := capabilities.Memory.Query(ctx, "customer history", 5)
            if err == nil && len(memories) > 0 {
                customerHistory = "Previous interactions:\n"
                for _, mem := range memories {
                    customerHistory += "- " + mem.Content + "\n"
                }
            }
        }
        
        // Check knowledge base for relevant documentation
        var knowledgeContext string
        if capabilities.Memory != nil {
            // Query knowledge base (assuming documents were ingested)
            docs, err := capabilities.Memory.Query(ctx, input, 3)
            if err == nil && len(docs) > 0 {
                knowledgeContext = "Relevant documentation:\n"
                for _, doc := range docs {
                    knowledgeContext += "- " + doc.Content + "\n"
                }
            }
        }
        
        // Build context-aware prompt
        systemPrompt := "You are a helpful customer support agent."
        if customerHistory != "" {
            systemPrompt += "\n\n" + customerHistory
        }
        if knowledgeContext != "" {
            systemPrompt += "\n\n" + knowledgeContext
        }
        
        // Generate response
        response, err := capabilities.LLM(systemPrompt, input)
        if err != nil {
            return "", err
        }
        
        // Store interaction
        if capabilities.Memory != nil {
            capabilities.Memory.Store(ctx, fmt.Sprintf("User: %s\nAgent: %s", input, response))
        }
        
        return response, nil
    }
    
    // Add memory augmentation for automatic context handling
    augmentedHandler := v1beta.WithMemoryAugmentation(handler)
    
    return v1beta.NewBuilder("SupportBot").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        WithMemory(
            v1beta.WithMemoryProvider("pgvector"),
            v1beta.WithRAG(4000, 0.3, 0.7),
            v1beta.WithSessionScoped(),
        ).
        WithHandler(augmentedHandler).
        Build()
}
```

---

## 🎯 Best Practices

### 1. Error Handling

Always handle errors gracefully:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    response, err := capabilities.LLM("You are a helpful assistant.", input)
    if err != nil {
        // Log error
        log.Printf("LLM error: %v", err)
        
        // Return fallback response
        return "I'm experiencing technical difficulties. Please try again.", nil
    }
    return response, nil
}
```

### 2. Context Timeout

Respect context deadlines:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Check if context is already cancelled
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }
    
    // Perform work...
    return capabilities.LLM("You are a helpful assistant.", input)
}
```

### 3. Capability Checks

Verify capabilities before using them:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    if capabilities.Memory == nil {
        return capabilities.LLM("You are a helpful assistant.", input)
    }
    
    // Use memory...
    memories, _ := capabilities.Memory.Query(ctx, input, 3)
    // ...
}
```

### 4. Structured Output

Return well-formatted responses:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    response, err := capabilities.LLM("You are a helpful assistant.", input)
    if err != nil {
        return "", err
    }
    
    // Format output consistently
    return fmt.Sprintf("Response: %s\nTimestamp: %s", response, time.Now().Format(time.RFC3339)), nil
}
```

### 5. Logging and Observability

Add logging for debugging:

```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    log.Printf("Processing input: %s", input)
    
    start := time.Now()
    response, err := capabilities.LLM("You are a helpful assistant.", input)
    duration := time.Since(start)
    
    if err != nil {
        log.Printf("LLM call failed after %v: %v", duration, err)
        return "", err
    }
    
    log.Printf("LLM call completed in %v", duration)
    return response, nil
}
```

---

## 📚 Next Steps

- **[Tool Integration](./tool-integration.md)** - Add tools to your handlers
- **[Memory and RAG](./memory-and-rag.md)** - Use memory in handlers
- **[Workflows](./workflows.md)** - Combine agents with custom handlers
- **[Error Handling](./error-handling.md)** - Advanced error patterns

---

**Ready to integrate tools?** Continue to [Tool Integration](./tool-integration.md) →
