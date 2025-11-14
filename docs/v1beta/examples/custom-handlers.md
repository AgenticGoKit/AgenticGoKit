# Custom Handlers Example

Implement custom business logic with full control over agent execution.

---

## Overview

Custom handlers allow you to:
- Implement custom business logic
- Access LLM, tools, and memory directly
- Build complex processing pipelines
- Override default agent behavior

---

## Complete Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create agent with custom handler
    agent, err := v1beta.NewBuilder("CustomAgent").
        WithLLM("openai", "gpt-4").
        WithHandler(customHandler).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Execute
    result, err := agent.Run(context.Background(), "What is Go?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Content)
}

// Custom handler with business logic
func customHandler(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    // Pre-processing
    input = strings.TrimSpace(input)
    input = strings.ToLower(input)

    // Check cache first
    if cached, found := checkCache(input); found {
        return cached, nil
    }

    // Call LLM with custom system prompt
    response, err := capabilities.LLM(
        "You are an expert Go programmer. Be concise and technical.",
        input,
    )
    if err != nil {
        return "", err
    }

    // Post-processing
    response = formatResponse(response)

    // Store in cache
    storeCache(input, response)

    return response, nil
}

func checkCache(key string) (string, bool) {
    // Implementation
    return "", false
}

func storeCache(key, value string) {
    // Implementation
}

func formatResponse(text string) string {
    // Add formatting
    return "**Response:**\n" + text
}
```

---

## Handler Signature

```go
type HandlerFunc func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error)
```

### Capabilities Object

```go
type Capabilities struct {
    // Call LLM with custom prompts
    LLM func(system, user string) (string, error)
    
    // Access registered tools
    Tools ToolManager
    
    // Access memory
    Memory Memory
}
```

---

## Handler Patterns

### Simple Handler

```go
func simpleHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Direct LLM call
    return caps.LLM("You are helpful", input)
}
```

### Multi-Step Handler

```go
func multiStepHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Step 1: Analyze input
    analysis, err := caps.LLM("Analyze the query", input)
    if err != nil {
        return "", err
    }

    // Step 2: Generate response based on analysis
    response, err := caps.LLM(
        "Generate a response based on this analysis: "+analysis,
        input,
    )
    if err != nil {
        return "", err
    }

    return response, nil
}
```

### Tool-Using Handler

```go
func toolHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Determine if tools are needed
    needsTool, err := caps.LLM("Does this need external data? Reply YES or NO", input)
    if err != nil {
        return "", err
    }

    if strings.Contains(strings.ToUpper(needsTool), "YES") {
        // Call tool
        toolResult, err := caps.Tools.Call("web_search", map[string]interface{}{
            "query": input,
        })
        if err != nil {
            return "", err
        }

        // Generate response with tool result
        return caps.LLM(
            "Answer based on this data: "+toolResult.Content,
            input,
        )
    }

    // No tool needed
    return caps.LLM("You are helpful", input)
}
```

### Memory-Aware Handler

```go
func memoryHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Search memory for relevant context
    memories, err := caps.Memory.Search(ctx, input, 3)
    if err != nil {
        return "", err
    }

    // Build context from memories
    context := "Relevant information:\n"
    for _, mem := range memories {
        context += fmt.Sprintf("- %s\n", mem.Content)
    }

    // Generate response with context
    response, err := caps.LLM(context, input)
    if err != nil {
        return "", err
    }

    // Store new memory
    caps.Memory.Store(ctx, input+": "+response)

    return response, nil
}
```

---

## Real-World Examples

### Validation Handler

```go
func validationHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Validate input
    if len(input) < 10 {
        return "", fmt.Errorf("input too short")
    }

    if containsOffensiveContent(input) {
        return "I cannot process offensive content", nil
    }

    // Process valid input
    return caps.LLM("You are helpful", input)
}
```

### Rate-Limited Handler

```go
var (
    rateLimiter = time.NewTicker(time.Second)
    requestChan = make(chan request, 100)
)

type request struct {
    input   string
    caps    *v1beta.Capabilities
    respCh  chan response
}

type response struct {
    content string
    err     error
}

func init() {
    go func() {
        for {
            <-rateLimiter.C
            req := <-requestChan
            resp, err := req.caps.LLM("You are helpful", req.input)
            req.respCh <- response{resp, err}
        }
    }()
}

func rateLimitedHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    respCh := make(chan response)
    requestChan <- request{input, caps, respCh}
    
    resp := <-respCh
    return resp.content, resp.err
}
```

### Fallback Handler

```go
func fallbackHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Try primary LLM
    response, err := caps.LLM("You are helpful", input)
    if err == nil {
        return response, nil
    }

    // Log failure
    log.Printf("Primary LLM failed: %v, trying fallback", err)

    // Try fallback (simpler response)
    return generateFallbackResponse(input), nil
}

func generateFallbackResponse(input string) string {
    return "I'm having trouble processing your request. Please try again later."
}
```

### Logging Handler

```go
func loggingHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    start := time.Now()
    
    // Log request
    log.Printf("Request: %s", input)

    // Process
    response, err := caps.LLM("You are helpful", input)
    
    duration := time.Since(start)
    
    // Log response
    if err != nil {
        log.Printf("Error after %v: %v", duration, err)
        return "", err
    }
    
    log.Printf("Success in %v, response length: %d", duration, len(response))
    
    return response, nil
}
```

---

## Advanced Patterns

### Chained Handlers

```go
func chainHandlers(handlers ...v1beta.HandlerFunc) v1beta.HandlerFunc {
    return func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
        currentInput := input
        
        for _, handler := range handlers {
            output, err := handler(ctx, currentInput, caps)
            if err != nil {
                return "", err
            }
            currentInput = output
        }
        
        return currentInput, nil
    }
}

// Usage
agent, _ := v1beta.NewBuilder("Agent").
    WithLLM("openai", "gpt-4").
    WithHandler(chainHandlers(
        validationHandler,
        preprocessHandler,
        mainHandler,
        postprocessHandler,
    )).
    Build()
```

### Conditional Handler

```go
func conditionalHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Route based on input type
    if isQuestion(input) {
        return handleQuestion(ctx, input, caps)
    } else if isCommand(input) {
        return handleCommand(ctx, input, caps)
    } else {
        return handleStatement(ctx, input, caps)
    }
}
```

---

## Error Handling

```go
func robustHandler(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Input validation
    if input == "" {
        return "", v1beta.NewAgentError(
            v1beta.ErrCodeInvalidInput,
            "empty input",
            nil,
        )
    }

    // Try with retries
    var lastErr error
    for i := 0; i < 3; i++ {
        response, err := caps.LLM("You are helpful", input)
        if err == nil {
            return response, nil
        }
        
        lastErr = err
        time.Sleep(time.Second * time.Duration(i+1))
    }

    return "", fmt.Errorf("handler failed after retries: %w", lastErr)
}
```

---

## Running the Example

```bash
go get github.com/agenticgokit/agenticgokit/v1beta
export OPENAI_API_KEY="sk-..."
go run main.go
```

---

## Next Steps

- **[Streaming Agent](./streaming-agent.md)** - Add streaming to custom handlers
- **[Memory & RAG](./memory-rag.md)** - Use memory in custom handlers
- **[Workflows](./workflow-sequential.md)** - Use custom handlers in workflows

---

## Related Documentation

- [Custom Handlers Guide](../custom-handlers.md) - Complete handler documentation
- [Core Concepts](../core-concepts.md) - Understanding handlers
- [Error Handling](../error-handling.md) - Error patterns in handlers
