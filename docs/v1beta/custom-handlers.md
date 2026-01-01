# Custom Handlers

Create custom agent logic with full control over execution flow while leveraging LLM, tools, and memory.

---

## Handler Basics

### Signature

```go
type HandlerFunc func(ctx context.Context, input string, capabilities *Capabilities) (string, error)
```

### Simple Handler

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return "Echo: " + input, nil
}

agent, _ := v1beta.NewBuilder("SimpleAgent").
    WithLLM("openai", "gpt-4").
    WithHandler(handler).
    Build()
```

---

## Core Patterns

### LLM-Only

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return capabilities.LLM("You are a helpful assistant.", input)
}
```

### Tool-Augmented

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    if strings.Contains(input, "calculate") {
        result, err := capabilities.Tools.Execute(ctx, "calculator", map[string]interface{}{
            "expression": extractExpression(input),
        })
        if err == nil {
            return fmt.Sprintf("Result: %v", result.Content), nil
        }
    }
    return capabilities.LLM("Answer this.", input)
}
```

### Memory-Aware

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    var context string
    if capabilities.Memory != nil {
        memories, _ := capabilities.Memory.Query(ctx, input, 3)
        for _, mem := range memories {
            context += "- " + mem.Content + "\n"
        }
    }
    
    systemPrompt := "You are helpful. Context:\n" + context
    response, _ := capabilities.LLM(systemPrompt, input)
    
    if capabilities.Memory != nil {
        capabilities.Memory.Store(ctx, fmt.Sprintf("Q: %s\nA: %s", input, response))
    }
    
    return response, nil
}
```

### Multi-Step Processing

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    // Step 1: Analyze intent
    intent, err := capabilities.LLM("Classify as: question, command, or statement", input)
    if err != nil {
        return "", err
    }
    
    // Step 2: Route based on intent
    switch strings.TrimSpace(intent) {
    case "question":
        return capabilities.LLM("Answer this question.", input)
    case "command":
        return "Executing: " + input, nil
    default:
        return "Understood: " + input, nil
    }
}
```

---

## Handler Augmentation

Enhance handlers with pre-built functionality:

### WithToolAugmentation

Automatic tool discovery and calling:

```go
baseHandler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return capabilities.LLM("You are helpful.", input)
}

handler := v1beta.WithToolAugmentation(baseHandler)

agent, _ := v1beta.NewBuilder("ToolAgent").
    WithLLM("openai", "gpt-4").
    WithTools(v1beta.WithMCP(servers...)).
    WithHandler(handler).
    Build()
```

### WithMemoryAugmentation

Automatic memory integration:

```go
baseHandler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return capabilities.LLM("You are helpful.", input)
}

handler := v1beta.WithMemoryAugmentation(baseHandler)

agent, _ := v1beta.NewBuilder("MemoryAgent").
    WithLLM("openai", "gpt-4").
    WithMemory(v1beta.WithMemoryProvider("chromem")).
    WithHandler(handler).
    Build()
```

### WithLLMAugmentation

Retry logic for LLM calls:

```go
baseHandler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return capabilities.LLM("You are helpful.", input)
}

// Retry up to 3 times
handler := v1beta.WithLLMAugmentation(baseHandler, 3)
```

### WithRAGAugmentation

Automatic RAG knowledge retrieval:

```go
baseHandler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return capabilities.LLM("Answer based on provided knowledge.", input)
}

// Retrieve top 5 relevant documents
handler := v1beta.WithRAGAugmentation(baseHandler, "knowledge_base", 5)

agent, _ := v1beta.NewBuilder("RAGAgent").
    WithLLM("openai", "gpt-4").
    WithMemory(v1beta.WithMemoryProvider("pgvector")).
    WithHandler(handler).
    Build()
```

---

## Handler Composition

### Chain Handlers

Execute in sequence (output of one is input to next):

```go
preprocess := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return strings.ToLower(input), nil
}

process := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return capabilities.LLM("Help with this.", input)
}

postprocess := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return strings.ToUpper(input), nil
}

handler := v1beta.Chain(preprocess, process, postprocess)

agent, _ := v1beta.NewBuilder("ChainedAgent").
    WithLLM("openai", "gpt-4").
    WithHandler(handler).
    Build()
```

### ParallelHandlers

Run multiple handlers and combine results:

```go
summary := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return capabilities.LLM("Summarize in one sentence.", input)
}

keywords := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return capabilities.LLM("Extract 5 keywords.", input)
}

sentiment := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    return capabilities.LLM("Analyze sentiment.", input)
}

// Combine with separator
handler := v1beta.ParallelHandlers("\n---\n", summary, keywords, sentiment)

agent, _ := v1beta.NewBuilder("AnalysisAgent").
    WithLLM("openai", "gpt-4").
    WithHandler(handler).
    Build()
```

---

## Complete Examples

### Research Assistant

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    // Check if web search needed
    decision, _ := capabilities.LLM("Determine if web search needed. Reply YES or NO.", input)
    
    var context string
    if strings.TrimSpace(decision) == "YES" {
        result, err := capabilities.Tools.Execute(ctx, "web_search", map[string]interface{}{
            "query": input,
        })
        if err == nil {
            context = fmt.Sprintf("Search results:\n%v", result.Content)
        }
    }
    
    systemPrompt := "You are a researcher."
    if context != "" {
        systemPrompt += "\n\nContext:\n" + context
    }
    
    response, _ := capabilities.LLM(systemPrompt, input)
    
    if capabilities.Memory != nil {
        capabilities.Memory.Store(ctx, fmt.Sprintf("Q: %s\nA: %s", input, response))
    }
    
    return response, nil
}

agent, _ := v1beta.NewBuilder("ResearchAssistant").
    WithLLM("openai", "gpt-4").
    WithTools(v1beta.WithMCP(servers...)).
    WithMemory(v1beta.WithMemoryProvider("chromem")).
    WithHandler(handler).
    Build()
```

### Code Review Agent

```go
handler := v1beta.ParallelHandlers("\n\n",
    func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
        return capabilities.LLM("Analyze for security issues.", input)
    },
    func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
        return capabilities.LLM("Analyze for performance issues.", input)
    },
    func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
        return capabilities.LLM("Review for best practices.", input)
    },
)

agent, _ := v1beta.NewBuilder("CodeReviewer").
    WithLLM("openai", "gpt-4").
    WithHandler(handler).
    Build()
```

### Customer Support

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    var context string
    if capabilities.Memory != nil {
        // Get customer history
        memories, _ := capabilities.Memory.Query(ctx, "customer history", 5)
        for _, mem := range memories {
            context += "- " + mem.Content + "\n"
        }
        
        // Get KB articles
        docs, _ := capabilities.Memory.Query(ctx, input, 3)
        for _, doc := range docs {
            context += "- " + doc.Content + "\n"
        }
    }
    
    systemPrompt := "You are a support agent."
    if context != "" {
        systemPrompt += "\n\nKnowledge:\n" + context
    }
    
    response, _ := capabilities.LLM(systemPrompt, input)
    
    if capabilities.Memory != nil {
        capabilities.Memory.Store(ctx, fmt.Sprintf("User: %s\nAgent: %s", input, response))
    }
    
    return response, nil
}

agent, _ := v1beta.NewBuilder("SupportBot").
    WithLLM("openai", "gpt-4").
    WithMemory(v1beta.WithMemoryProvider("pgvector")).
    WithHandler(v1beta.WithMemoryAugmentation(handler)).
    Build()
```

---

## Best Practices

### Error Handling

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    response, err := capabilities.LLM("Be helpful.", input)
    if err != nil {
        log.Printf("LLM error: %v", err)
        return "Technical difficulties. Please try again.", nil
    }
    return response, nil
}
```

### Context Timeout

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
    }
    
    return capabilities.LLM("Be helpful.", input)
}
```

### Capability Checks

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    if capabilities.Memory == nil {
        return capabilities.LLM("Be helpful.", input)
    }
    
    // Use memory...
    memories, _ := capabilities.Memory.Query(ctx, input, 3)
    // ...
}
```

### Logging

```go
handler := func(ctx context.Context, input string, capabilities *Capabilities) (string, error) {
    log.Printf("Processing: %s", input)
    start := time.Now()
    
    response, err := capabilities.LLM("Be helpful.", input)
    
    log.Printf("Completed in %v", time.Since(start))
    return response, err
}
```

---

**Next:** [Error Handling](./error-handling.md) → [Memory and RAG](./memory-and-rag.md)
