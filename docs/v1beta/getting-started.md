# Getting Started

Welcome to AgenticGoKit! This guide will help you build your first AI agent in just a few minutes.

---

## 📋 Prerequisites

Before you begin, make sure you have:

- **Go 1.21 or later** installed
- **An LLM API key** (OpenAI, Azure AI, Ollama, HuggingFace, or OpenRouter)
- **Basic Go knowledge** (functions, structs, error handling)

---

## 📦 Installation

Install AgenticGoKit:

```bash
go get github.com/agenticgokit/agenticgokit/v1beta
```

Initialize your Go module (if you haven't already):

```bash
go mod init myagent
go mod tidy
```

---

## 🚀 Your First Agent (5 Minutes)

### Step 1: Create a Basic Agent

Create a file named `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Set your OpenAI API key
    os.Setenv("OPENAI_API_KEY", "your-api-key-here")
    
    // Create an agent with preset configuration
    agent, err := v1beta.NewBuilder("Assistant").
        WithPreset(v1beta.ChatAgent).
        WithLLM("openai", "gpt-4").
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    // Run a simple query
    result, err := agent.Run(context.Background(), "What is Go programming language?")
    if err != nil {
        log.Fatal(err)
    }
    
    // Print the response
    fmt.Printf("Response: %s\n", result.Content)
    fmt.Printf("Success: %t\n", result.Success)
}
```

### Step 2: Run Your Agent

```bash
go run main.go
```

**Output:**
```
Response: Go is a statically typed, compiled programming language designed at Google...
Success: true
```

Congratulations! You've built your first AgenticGoKit agent! 🎉

---

## 🎯 Preset Builders

AgenticGoKit provides preset configurations for common agent types:

### Chat Agent Preset
For general-purpose chat agents:

```go
agent, err := v1beta.NewBuilder("chat-assistant").
    WithPreset(v1beta.ChatAgent).
    WithName("ChatBot").
    WithModel("openai", "gpt-4").
    WithTemperature(0.7).
    Build()
```

### Research Agent Preset
For research and analysis tasks:

```go
agent, err := v1beta.NewBuilder("Researcher").
    WithPreset(v1beta.ResearchAgent).
    WithLLM("openai", "gpt-4").
    Build()
```

### QuickChatAgent
For rapid prototyping:

```go
agent, _ := v1beta.QuickChatAgent("gpt-4")
result, err := agent.Run(context.Background(), "Hello!")
```

---

## 🔧 Builder Pattern

AgenticGoKit uses a fluent builder pattern for configuration:

```go
agent, err := v1beta.NewBuilder("MyAgent").
    WithLLM("openai", "gpt-4").
    WithSystemPrompt("You are a helpful assistant").
    WithAgentTimeout(30 * time.Second).
    Build()
```

### Common Builder Methods

| Method | Description | Example |
|--------|-------------|---------|
| `WithLLM(provider, model string)` | Set LLM provider and model | `.WithLLM("openai", "gpt-4")` |
| `WithSystemPrompt(string)` | Set system prompt | `.WithSystemPrompt("You are helpful")` |
| `WithAgentTimeout(duration)` | Set timeout | `.WithAgentTimeout(30*time.Second)` |
| `WithTools(opts ...ToolOption)` | Configure tools | `.WithTools(v1beta.WithMCP(servers...))` |
| `WithMemory(opts ...MemoryOption)` | Configure memory | `.WithMemory(v1beta.WithMemoryProvider("memory"))` |

---

## 📝 Running Agents

### Simple Run
Execute a single query and get the complete response:

```go
result, err := agent.Run(context.Background(), "Explain quantum computing")
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Content)
```

### Run with Options
Pass additional options at runtime:

```go
opts := &v1beta.RunOptions{
    MaxTokens:   1000,
    Temperature: func(t float64) *float64 { return &t }(0.5),
}

result, err := agent.RunWithOptions(
    context.Background(),
    "Explain quantum computing",
    opts,
)
```

### Context with Cancellation
Use context for timeouts and cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := agent.Run(ctx, "Long computation task")
if err == context.DeadlineExceeded {
    fmt.Println("Request timed out!")
}
```

---

## ⚡ Streaming Responses

Get real-time streaming responses:

### Channel-Based Streaming

```go
stream, err := agent.RunStream(context.Background(), "Write a story")
if err != nil {
    log.Fatal(err)
}

for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeText, v1beta.ChunkTypeDelta:
        if chunk.Delta != "" {
            fmt.Print(chunk.Delta)
        } else {
            fmt.Print(chunk.Content)
        }
    case v1beta.ChunkTypeDone:
        fmt.Println("\n✓ Done!")
    case v1beta.ChunkTypeError:
        fmt.Println("Error:", chunk.Error)
    }
}

result, err := stream.Wait()
```

### Callback-Based Streaming

```go
handler := func(chunk *v1beta.StreamChunk) bool {
    if chunk.Type == v1beta.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
    return true // continue streaming
}

stream, err := agent.RunStream(
    context.Background(),
    "Explain AI",
    v1beta.WithStreamHandler(handler),
)
result, _ := stream.Wait()
```

**Learn more**: [Streaming Guide](./streaming.md)

---

## 🔄 Multi-Agent Workflows

Create workflows with multiple agents:

### Sequential Workflow

```go
// Create agents
agent1, _ := v1beta.QuickChatAgent("gpt-4")
agent2, _ := v1beta.QuickChatAgent("gpt-4")

// Create sequential workflow
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 60 * time.Second,
}

workflow, err := v1beta.NewSequentialWorkflow(config)
if err != nil {
    log.Fatal(err)
}

workflow.AddStep(v1beta.WorkflowStep{
    Name:  "research",
    Agent: agent1,
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:  "write",
    Agent: agent2,
})

// Execute workflow
result, err := workflow.Run(context.Background(), "Research quantum computing")
fmt.Println(result.FinalOutput)
```

### Parallel Workflow

```go
// Create parallel workflow
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 90 * time.Second,
}

workflow, err := v1beta.NewParallelWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "tech", Agent: techAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "business", Agent: bizAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "legal", Agent: legalAgent})

result, err := workflow.Run(context.Background(), "Analyze the product")
```

**Learn more**: [Workflows Guide](./workflows.md)

---

## 🛠️ Adding Tools

Extend agent capabilities with tools:

```go
// Define a custom tool
weatherTool := v1beta.Tool{
    Name:        "get_weather",
    Description: "Get current weather for a location",
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type":        "string",
                "description": "City name",
            },
        },
        "required": []string{"location"},
    },
    Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
        location := args["location"].(string)
        // Call weather API...
        return fmt.Sprintf("Weather in %s: Sunny, 72°F", location), nil
    },
}

// Create agent with tools
// Note: Tools are typically integrated via MCP servers or tool managers
// For custom tools, implement via CustomHandlerFunc
agent, err := v1beta.NewBuilder("WeatherBot").
    WithPreset(v1beta.ChatAgent).
    Build()

// Agent can now use the weather tool
result, err := agent.Run(context.Background(), "What's the weather in San Francisco?")
```

**Learn more**: [Tool Integration Guide](./tool-integration.md)

---

## 💾 Adding Memory

Give your agent memory capabilities:

```go
// Create agent with memory
agent, err := v1beta.NewBuilder("MemoryBot").
    WithPreset(v1beta.ChatAgent).
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
        v1beta.WithSessionScoped(),
        v1beta.WithContextAware(),
    ).
    Build()

// Agent remembers context across calls
result1, _ := agent.Run(context.Background(), "My name is Alice")
result2, _ := agent.Run(context.Background(), "What is my name?")
// Response: "Your name is Alice"
```

**Learn more**: [Memory & RAG Guide](./memory-and-rag.md)

---

## 🎨 Custom Handlers

Implement custom logic with handlers:

### CustomHandlerFunc
Simple handler with LLM fallback:

```go
customHandler := func(ctx context.Context, query string, llmCall func(string, string) (string, error)) (string, error) {
    // Custom logic
    if strings.Contains(query, "time") {
        return fmt.Sprintf("Current time: %s", time.Now().Format(time.RFC3339)), nil
    }
    
    // Fallback to LLM (return empty string)
    return "", nil
}

agent, err := v1beta.NewBuilder("custom-agent").
    WithHandler(customHandler).
    Build()
```

### EnhancedHandlerFunc
Advanced handler with full capabilities:

```go
enhancedHandler := func(ctx context.Context, query string, capabilities *v1beta.Capabilities) (string, error) {
    // Access LLM
    llmResponse, err := capabilities.LLM("system prompt", query)
    if err != nil {
        return "", err
    }
    
    // Access tools if available
    if capabilities.Tools != nil {
        toolResult, _ := capabilities.Tools.Execute(ctx, "search", map[string]interface{}{
            "query": query,
        })
        return fmt.Sprintf("LLM: %s\nTool: %v", llmResponse, toolResult.Content), nil
    }
    
    return llmResponse, nil
}

agent, err := v1beta.NewBuilder("enhanced-agent").
    WithHandler(enhancedHandler).
    Build()
```

**Learn more**: [Custom Handlers Guide](./custom-handlers.md)

---

## 📊 Error Handling

AgenticGoKit provides clear error types:

```go
result, err := agent.Run(context.Background(), "query")
if err != nil {
    // Check for structured error
    if agentErr, ok := err.(*v1beta.AgentError); ok {
        switch agentErr.Code {
        case v1beta.ErrConfigInvalid:
            log.Println("Configuration error")
        case v1beta.ErrLLMCallFailed:
            log.Println("LLM provider error")
        case v1beta.ErrContextTimeout:
            log.Println("Request timeout")
        default:
            log.Println("Error:", agentErr.Message)
        }
    } else {
        log.Println("Unknown error:", err)
    }
}

// Check result status
if !result.Success {
    log.Printf("Agent failed: %v", result.Error)
}
```

**Learn more**: [Error Handling Guide](./error-handling.md)

---

## 🎯 Best Practices

### 1. Always Handle Errors
```go
agent, err := v1beta.NewBuilder("my-agent").
    WithPreset(v1beta.ChatAgent).
    Build()
if err != nil {
    log.Fatal(err)
}
```

### 2. Use Context for Cancellation
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := agent.Run(ctx, query)
```

### 3. Set Appropriate Timeouts
```go
agent, err := v1beta.NewBuilder("Agent").
    WithTimeout(15 * time.Second).
    Build()
```

### 4. Use Preset Builders for Common Cases
```go
// Instead of manual configuration
agent, err := v1beta.PresetChatAgentBuilder().
    WithName("Assistant").
    Build()
```

### 5. Check Result Success
```go
result, err := agent.Run(ctx, query)
if err != nil || !result.Success {
    // Handle failure
}
```

---

## 📚 Next Steps

Now that you've built your first agent, explore more features:

1. **[Core Concepts](./core-concepts.md)** - Deep dive into architecture
2. **[Streaming Guide](./streaming.md)** - Real-time responses
3. **[Workflows](./workflows.md)** - Multi-agent orchestration
4. **[Tool Integration](./tool-integration.md)** - Extend agent capabilities
5. **[Memory & RAG](./memory-and-rag.md)** - Add memory and knowledge
6. **[Examples](./examples/)** - Complete code examples

---

## 🆘 Need Help?

- **[Troubleshooting](./troubleshooting.md)** - Common issues
- **[API Reference](./api-reference.md)** - Complete API docs
- **[GitHub Issues](https://github.com/agenticgokit/agenticgokit/issues)** - Report bugs
- **[Discussions](https://github.com/agenticgokit/agenticgokit/discussions)** - Ask questions

---

## 🔄 Migrating from core/vnext?

If you're upgrading from the deprecated `core` or `core/vnext` packages, see the **[Migration Guide](./migration-from-core.md)** for step-by-step instructions.

---

**Ready for more?** Continue to [Core Concepts](./core-concepts.md) →
