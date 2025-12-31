# AgenticGoKit v1beta API

The v1beta API is an advanced, comprehensive agent framework that provides flexible and powerful capabilities for building custom AI agents. It offers streamlined APIs, real-time streaming, multi-agent workflows, and comprehensive tooling support.

## 🚀 Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create an agent with preset configuration
    agent, err := v1beta.PresetChatAgentBuilder().
        WithName("Assistant").
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Run a simple query
    result, err := agent.Run(context.Background(), "Hello, world!")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", result.Content)
}
```

## 📋 Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Core Concepts](#core-concepts)
- [Streaming](#streaming)
- [Workflows](#workflows)
- [Configuration](#configuration)
- [Documentation](#documentation)
- [Examples](#examples)
- [Performance](#performance)
- [Testing](#testing)

## ✨ Features

### 🎯 Streamlined API
- **8 core methods** (reduced from 30+)
- **Unified RunOptions** for all execution modes
- **Preset builders** for common agent types
- **Functional options** pattern for clean configuration

### ⚡ Real-Time Streaming
- **8 chunk types**: Text, Delta, Thought, ToolCall, ToolResult, Metadata, Error, Done
- **Multiple patterns**: Channel-based, callback-based, io.Reader
- **Configurable buffering** and flush intervals
- **Full lifecycle control** with cancellation support

### 🔄 Multi-Agent Workflows
- **4 workflow modes**: Sequential, Parallel, DAG, Loop
- **Step-by-step streaming** with progress tracking
- **Context sharing** between agents
- **Error handling** and recovery

### 🛠️ Comprehensive Tooling
- **Tool registration** and discovery
- **MCP integration** for Model Context Protocol
- **Caching** and rate limiting
- **Timeout** and retry handling

### 💾 Memory & RAG
- **Multiple backends**: In-memory, PostgreSQL (pgvector), Weaviate
- **RAG support** with configurable weights
- **Session management** and history tracking
- **Context augmentation** for handlers

### 🎛️ Flexible Configuration
- **TOML-based** configuration files
- **Environment variables** support
- **Functional options** for programmatic config
- **Validation** and defaults

## 📦 Installation

```bash
go get github.com/agenticgokit/agenticgokit/v1beta
```

## 🎓 Core Concepts

### 1. Custom Handler Functions

The framework provides two types of custom handlers:

#### CustomHandlerFunc
A simple handler function that receives user input and an LLM call function. It's perfect for basic custom logic with LLM fallback.

```go
customHandler := func(ctx context.Context, query string, llmCall func(string, string) (string, error)) (string, error) {
    if strings.Contains(query, "weather") {
        return "I can help with weather queries!", nil
    }
    // Return empty string to fall back to default LLM processing
    return "", nil
}

builder := v1beta.PresetChatAgentBuilder().WithCustomHandler(customHandler)
```

#### EnhancedHandlerFunc
An advanced handler with full access to agent capabilities through HandlerCapabilities, including LLM, tools, and memory systems.

```go
enhancedHandler := func(ctx context.Context, query string, capabilities *v1beta.HandlerCapabilities) (string, error) {
    // Use LLM
    llmResponse, err := capabilities.LLMCall("You are a helpful assistant", query)
    if err != nil {
        return "", err
    }

    // Use tools
    toolResult, err := capabilities.ToolCall("weather_lookup", map[string]interface{}{
        "location": "New York",
    })
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("LLM: %s\nTool: %v", llmResponse, toolResult), nil
}

builder := v1beta.PresetChatAgentBuilder().WithEnhancedHandler(enhancedHandler)
```

### 2. Handler Augmentation Functions

Pre-built handlers that automatically integrate common capabilities:

#### CreateToolAugmentedHandler
Creates a handler that automatically includes tool information in LLM prompts.

#### CreateMemoryAugmentedHandler
Creates a handler that automatically includes relevant memory context.

#### CreateFullAugmentedHandler
Creates a handler with both tool and memory augmentation.

```go
toolHandler := v1beta.CreateToolAugmentedHandler(func(ctx context.Context, query, toolPrompt string, llmCall func(string, string) (string, error)) (string, error) {
    response, err := llmCall("You are a helpful assistant with tool access", query)
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("LLM Response: %s\nTool Context was available", response), nil
})

builder := v1beta.PresetChatAgentBuilder().WithEnhancedHandler(toolHandler)
```

### 3. ToolCallHelper

A simplified interface for custom handlers to execute tools with various argument types:

```go
enhancedHandler := func(ctx context.Context, query string, capabilities *v1beta.HandlerCapabilities) (string, error) {
    toolHelper := v1beta.NewToolCallHelper(capabilities)
    
    // Call tool with map arguments
    result, err := toolHelper.Call("weather_lookup", map[string]interface{}{
        "location": "London",
        "units": "celsius",
    })
    if err != nil {
        return "", err
    }
    
    return result, nil
}
```

### 4. Middleware Support

Flexible middleware system with BeforeRun and AfterRun hooks:

```go
type LoggingMiddleware struct{}

func (m *LoggingMiddleware) BeforeRun(ctx context.Context, input string) (context.Context, string, error) {
    fmt.Printf("Processing input: %s\n", input)
    return ctx, input, nil
}

func (m *LoggingMiddleware) AfterRun(ctx context.Context, input string, result *v1beta.AgentResult, err error) (*v1beta.AgentResult, error) {
    fmt.Printf("Result success: %t\n", result.Success)
    return result, err
}

builder := v1beta.PresetChatAgentBuilder().
    WithCustomHandler(customHandler).
    WithMiddlewares([]v1beta.AgentMiddleware{&LoggingMiddleware{}})
```

## 🌊 Streaming

Real-time streaming for responsive UIs and long-running operations:

### Basic Streaming

```go
stream, err := agent.RunStream(ctx, "Tell me a story")
if err != nil {
    log.Fatal(err)
}

// Process chunks as they arrive
for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeDelta:
        fmt.Print(chunk.Delta)
    case v1beta.ChunkTypeThought:
        log.Printf("Thinking: %s", chunk.Content)
    case v1beta.ChunkTypeToolCall:
        log.Printf("Using tool: %s", chunk.ToolName)
    }
}

result, err := stream.Wait()
```

### Streaming with Options

```go
stream, err := agent.RunStream(ctx, query,
    v1beta.WithBufferSize(200),
    v1beta.WithThoughts(true),
    v1beta.WithToolCalls(true),
    v1beta.WithTimeout(5*time.Minute),
)
```

### Callback Handler

```go
handler := func(chunk *v1beta.StreamChunk) error {
    if chunk.Type == v1beta.ChunkTypeDelta {
        sendToWebSocket(chunk.Delta)
    }
    return nil
}

result, err := agent.Run(ctx, query, v1beta.WithStreamHandler(handler))
```

**[📖 Complete Streaming Guide →](STREAMING_GUIDE.md)**

## 🔄 Workflows

Build multi-agent systems with different execution patterns:

### Sequential Workflow

```go
workflow, err := v1beta.NewSequentialWorkflow("DataPipeline",
    v1beta.Step("extract", extractAgent, "Extract data"),
    v1beta.Step("transform", transformAgent, "Transform data"),
    v1beta.Step("load", loadAgent, "Load data"),
)

result, err := workflow.Run(ctx, "Process dataset.csv")
```

### Parallel Workflow

```go
workflow, err := v1beta.NewParallelWorkflow("Analysis",
    v1beta.Step("sentiment", sentimentAgent, "Analyze sentiment"),
    v1beta.Step("summary", summaryAgent, "Summarize content"),
    v1beta.Step("keywords", keywordAgent, "Extract keywords"),
)

result, err := workflow.Run(ctx, "Analyze this article")
```

### SubWorkflows (Workflow Composition)

**Workflows can be used as agents within other workflows**, enabling powerful composition patterns:

```go
// Create a parallel analysis subworkflow
analysisWorkflow, _ := v1beta.NewParallelWorkflow(&v1beta.WorkflowConfig{
    Name: "Analysis",
})
analysisWorkflow.AddStep(v1beta.WorkflowStep{Name: "sentiment", Agent: sentimentAgent})
analysisWorkflow.AddStep(v1beta.WorkflowStep{Name: "keywords", Agent: keywordAgent})

// Wrap as an agent using the builder
subAgent, _ := v1beta.NewBuilder("sub-agent").
    WithSubWorkflow(
        v1beta.WithWorkflowInstance(analysisWorkflow),
        v1beta.WithSubWorkflowMaxDepthBuilder(5),
    ).
    Build()

// Use in parent workflow
mainWorkflow, _ := v1beta.NewSequentialWorkflow(&v1beta.WorkflowConfig{
    Name: "ContentPipeline",
})
mainWorkflow.AddStep(v1beta.WorkflowStep{Name: "fetch", Agent: fetchAgent})
mainWorkflow.AddStep(v1beta.WorkflowStep{Name: "analyze", Agent: subAgent}) // SubWorkflow!
mainWorkflow.AddStep(v1beta.WorkflowStep{Name: "report", Agent: reportAgent})
```

**Alternative: Direct SubWorkflow Creation**

```go
// Direct creation without builder
subAgent := v1beta.NewSubWorkflowAgent("analysis", analysisWorkflow,
    v1beta.WithSubWorkflowMaxDepth(5),
    v1beta.WithSubWorkflowDescription("Multi-faceted analysis"),
)
```

**Benefits:**
- **Modularity**: Break complex workflows into reusable components
- **Clarity**: Each workflow focuses on a specific task
- **Testability**: Test subworkflows independently
- **Reusability**: Use same subworkflow in multiple parent workflows

**Example:** See `examples/story-writer-chat-v2/` for a complete multi-character story generation system using SubWorkflows.

### Workflow Streaming

```go
stream, err := workflow.RunStream(ctx, input)

for chunk := range stream.Chunks() {
    if stepName, ok := chunk.Metadata["step_name"].(string); ok {
        fmt.Printf("Executing: %s\n", stepName)
    }
}
```

## ⚙️ Configuration

### Programmatic Configuration

```go
agent, err := v1beta.PresetChatAgentBuilder().
    WithName("Assistant").
    WithSystemPrompt("You are a helpful assistant").
    WithLLM("openai", "gpt-4").
    WithMemory("memory", "inmemory").
    WithTools(myTools).
    Build()
```

### TOML Configuration

```toml
name = "MyAgent"
system_prompt = "You are a helpful assistant"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 1000

[memory]
provider = "memory"
connection = "inmemory"

[streaming]
enabled = true
buffer_size = 100
include_thoughts = true
include_tool_calls = true

[tools]
enabled = true
max_retries = 3
timeout = "30s"
```

Load configuration:

```go
config, err := v1beta.LoadConfig("config.toml")
agent, err := v1beta.NewAgentFromConfig(config)
```

## 📚 Documentation

- **[Streaming Guide](STREAMING_GUIDE.md)** - Complete streaming documentation with examples
- **[Migration Guide](MIGRATION_GUIDE.md)** - Migrating from older APIs
- **[Troubleshooting Guide](TROUBLESHOOTING.md)** - Common issues and solutions
- **[API Reference](https://pkg.go.dev/github.com/agenticgokit/agenticgokit/v1beta)** - Go package documentation

## 💡 Examples

Complete working examples in the `examples/` directory:

### Basic Examples
- `examples/ollama-quickstart/` - Simple agent usage with Ollama
- `examples/openrouter-quickstart/` - Using OpenRouter API
- `examples/huggingface-quickstart/` - HuggingFace integration

### Advanced Examples
- `examples/streaming-demo/` - Real-time streaming implementations
- `examples/sequential-workflow-demo/` - Multi-agent workflows
- `examples/story-writer-chat-v2/` - Complex workflow with SubWorkflows
- `examples/researcher-reporter/` - Research and reporting workflow

### Integration Examples
- `examples/memory-and-tools/` - Memory and tool integration
- `examples/mcp-integration/` - MCP protocol integration
- `examples/conversation-memory-demo/` - Memory management

## 🎯 Core API Reference

### Agent Interface

```go
type Agent interface {
    // Basic execution
    Run(ctx context.Context, input string, opts ...RunOption) (*Result, error)
    RunWithOptions(ctx context.Context, input string, opts *RunOptions) (*Result, error)
    
    // Streaming execution
    RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error)
    RunStreamWithOptions(ctx context.Context, input string, runOpts *RunOptions, streamOpts ...StreamOption) (Stream, error)
    
    // Metadata
    Name() string
    Config() *Config
}
```

### Workflow Interface

```go
type Workflow interface {
    // Basic execution
    Run(ctx context.Context, input string) (*WorkflowResult, error)
    
    // Streaming execution
    RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error)
    
    // Metadata
    Name() string
    Steps() []WorkflowStep
}
```

### Stream Interface

```go
type Stream interface {
    Chunks() <-chan *StreamChunk
    Wait() (*Result, error)
    Cancel()
    Metadata() *StreamMetadata
    AsReader() io.Reader
}
```

### Preset Builders

```go
// Chat-focused agent
v1beta.PresetChatAgentBuilder()

// Research agent with memory and tools
v1beta.PresetResearchAgentBuilder()

// Data processing agent
v1beta.PresetDataAgentBuilder()

// Workflow orchestration agent
v1beta.PresetWorkflowAgentBuilder()
```

## 🚀 Performance

### Benchmarks

- **Memory efficient**: Streaming reduces memory usage by 70%
- **Low latency**: Chunks delivered in <50ms
- **High throughput**: Handles 1000+ concurrent streams
- **Optimized**: Zero-allocation hot paths

### Best Practices

1. **Use streaming for long-running operations**
   ```go
   stream, _ := agent.RunStream(ctx, query)
   ```

2. **Configure buffer sizes appropriately**
   ```go
   // Real-time: small buffer (50)
   v1beta.WithBufferSize(50)
   
   // Batch: large buffer (500)
   v1beta.WithBufferSize(500)
   ```

3. **Use parallel workflows when possible**
   ```go
   workflow, _ := v1beta.NewParallelWorkflow(...)
   ```

4. **Enable caching for repeated operations**
   ```toml
   [tools.cache]
   enabled = true
   ttl = "5m"
   ```

5. **Always use context timeouts**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

**[📖 Performance Optimization Guide →](STREAMING_GUIDE.md#performance-considerations)**

## 🧪 Testing

Run the test suite:

```bash
# All tests
go test ./test/v1beta/...

# Specific test
go test ./test/v1beta/streaming -run TestStreamingAgent

# With coverage
go test ./test/v1beta/... -cover

# Verbose output
go test ./test/v1beta/... -v

# Run benchmarks
go test ./test/v1beta/benchmarks -bench=.
```

