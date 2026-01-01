# Streaming Agent Example

Real-time streaming responses with chunk-by-chunk output.

---

## Overview

This example demonstrates:
- Streaming agent responses in real-time
- Processing different chunk types
- Handling streaming errors
- Channel-based streaming pattern

---

## Complete Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create agent
    agent, err := v1beta.NewBuilder("StreamingAgent").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        Build()
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Start streaming
    stream, err := agent.RunStream(context.Background(), "Explain Go concurrency in detail")
    if err != nil {
        log.Fatalf("Failed to start stream: %v", err)
    }

    fmt.Println("Streaming response:")
    fmt.Println("---")

    // Process stream chunks via channel
    for chunk := range stream.Chunks() {
        // Handle different chunk types
        switch chunk.Type {
        case v1beta.ChunkTypeText:
            fmt.Print(chunk.Content)
        case v1beta.ChunkTypeDelta:
            fmt.Print(chunk.Content)
        case v1beta.ChunkTypeDone:
            fmt.Println("\n---")
            fmt.Println("Stream complete")
        case v1beta.ChunkTypeError:
            log.Printf("Chunk error: %s", chunk.Content)
        }
    }

    // Wait for stream completion and get final result
    result, err := stream.Wait()
    if err != nil {
        log.Printf("Stream error: %v", err)
    } else {
        fmt.Printf("Tokens used: %d\n", result.TokensUsed)
    }
}
```

---

## Step-by-Step Breakdown

### 1. Create Streaming Agent

```go
agent, err := v1beta.NewBuilder("StreamingAgent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
    }).
    Build()
```

Same builder pattern as basic agents - streaming is available by default.

### 2. Start Stream

```go
stream, err := agent.RunStream(context.Background(), query)
if err != nil {
    log.Fatalf("Failed to start stream: %v", err)
}
```

**Key Points:**
- `RunStream()` returns a `Stream` interface
- Use `stream.Chunks()` channel to receive chunks
- Use `stream.Wait()` to get the final result
- Use `stream.Cancel()` to stop the stream early

### 3. Process Chunks

```go
// Channel-based streaming - iterate over chunks
for chunk := range stream.Chunks() {
    // Process each chunk as it arrives
    fmt.Print(chunk.Content)
}

// Wait for completion and get final result
result, err := stream.Wait()
if err != nil {
    log.Printf("Stream error: %v", err)
}
```

---

## Chunk Types

The v1beta API provides several chunk types:

```go
const (
    ChunkTypeText       ChunkType = "text"       // Complete text chunk
    ChunkTypeDelta      ChunkType = "delta"      // Incremental text update
    ChunkTypeThought    ChunkType = "thought"    // Agent reasoning
    ChunkTypeToolCall   ChunkType = "tool_call"  // Tool invocation
    ChunkTypeToolResult ChunkType = "tool_result" // Tool response
    ChunkTypeMetadata   ChunkType = "metadata"   // Execution metadata
    ChunkTypeError      ChunkType = "error"      // Error information
    ChunkTypeDone       ChunkType = "done"       // Stream completion
)
```

### Handling All Chunk Types

```go
for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeText:
        fmt.Print(chunk.Content)
        
    case v1beta.ChunkTypeDelta:
        fmt.Print(chunk.Content)
        
    case v1beta.ChunkTypeThought:
        fmt.Printf("[Thinking: %s]\n", chunk.Content)
        
    case v1beta.ChunkTypeToolCall:
        fmt.Printf("[Calling tool: %s]\n", chunk.Metadata["tool_name"])
        
    case v1beta.ChunkTypeToolResult:
        fmt.Printf("[Tool result: %s]\n", chunk.Content)
        
    case v1beta.ChunkTypeMetadata:
        // Process metadata (tokens, timing, etc.)
        if tokens, ok := chunk.Metadata["tokens"].(int); ok {
            fmt.Printf("[Tokens used: %d]\n", tokens)
        }
        
    case v1beta.ChunkTypeError:
        fmt.Printf("[Error: %s]\n", chunk.Content)
        
    case v1beta.ChunkTypeDone:
        fmt.Println("\n[Stream complete]")
    }
}

// Get final result after stream completes
result, err := stream.Wait()
if err != nil {
    return err
}
fmt.Printf("Total tokens: %d\n", result.TokensUsed)
```

---

## Advanced Patterns

### Channel-Based Streaming

The Stream interface already provides a channel via `Chunks()`. Here's how to forward to your own channel:

```go
func streamToChannel(agent v1beta.Agent, query string) (<-chan string, <-chan error) {
    textChan := make(chan string, 10)
    errChan := make(chan error, 1)

    go func() {
        defer close(textChan)
        defer close(errChan)

        stream, err := agent.RunStream(context.Background(), query)
        if err != nil {
            errChan <- err
            return
        }

        // Read from the stream's built-in channel
        for chunk := range stream.Chunks() {
            if chunk.Type == v1beta.ChunkTypeText || chunk.Type == v1beta.ChunkTypeDelta {
                textChan <- chunk.Content
            }
        }

        // Check for errors on completion
        if _, err := stream.Wait(); err != nil {
            errChan <- err
        }
    }()

    return textChan, errChan
}

// Usage
textChan, errChan := streamToChannel(agent, "Tell me about Go")
for text := range textChan {
    fmt.Print(text)
}
if err := <-errChan; err != nil {
    log.Printf("Error: %v", err)
}
```

### Streaming with Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Start stream
stream, err := agent.RunStream(ctx, "Long running query...")
if err != nil {
    log.Fatal(err)
}

// Cancel stream after 5 seconds
go func() {
    time.Sleep(5 * time.Second)
    stream.Cancel() // Use stream's Cancel method
}()

// Process chunks - channel closes when cancelled or complete
for chunk := range stream.Chunks() {
    fmt.Print(chunk.Content)
}

// Check completion status
result, err := stream.Wait()
if err != nil {
    if err == context.Canceled {
        fmt.Println("\nStream cancelled")
    } else {
        log.Printf("Error: %v", err)
    }
}
```

### Callback-Based Streaming

```go
func streamWithCallback(agent v1beta.Agent, query string, onChunk func(string)) error {
    stream, err := agent.RunStream(context.Background(), query)
    if err != nil {
        return err
    }

    // Process chunks via channel with callback
    for chunk := range stream.Chunks() {
        if chunk.Type == v1beta.ChunkTypeText || chunk.Type == v1beta.ChunkTypeDelta {
            onChunk(chunk.Content)
        }
    }

    // Return any error from stream completion
    _, err = stream.Wait()
    return err
}

// Usage
err := streamWithCallback(agent, "Explain goroutines", func(text string) {
    fmt.Print(text)
})
```

---

## Running the Example

### Prerequisites

```bash
go get github.com/agenticgokit/agenticgokit/v1beta
export OPENAI_API_KEY="sk-..."
```

### Execute

```bash
go run main.go
```

---

## Performance Tips

### Buffer Size Configuration

```go
agent, err := v1beta.NewBuilder("Agent").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        StreamBufferSize: 100,  // Increase for high-throughput streams
        StreamFlushInterval: 50 * time.Millisecond,
    }).
    Build()
```

### Memory Management

```go
// Process large streams without loading entire response
var totalTokens int
for chunk := range stream.Chunks() {
    // Process chunk immediately, don't accumulate
    fmt.Print(chunk.Content)
    
    // Track metadata
    if tokens, ok := chunk.Metadata["tokens"].(int); ok {
        totalTokens += tokens
    }
}

// Get final token count from result
result, _ := stream.Wait()
fmt.Printf("\nTotal tokens: %d\n", result.TokensUsed)
```

---

## Error Handling

### Robust Stream Processing

```go
func processStream(stream v1beta.Stream) error {
    // Process all chunks from the channel
    for chunk := range stream.Chunks() {
        // Handle error chunks
        if chunk.Type == v1beta.ChunkTypeError {
            log.Printf("Chunk error: %s", chunk.Content)
            continue
        }
        
        fmt.Print(chunk.Content)
    }
    
    // Check final result for errors
    _, err := stream.Wait()
    if err != nil {
        return fmt.Errorf("stream error: %w", err)
    }
    
    return nil
}
```

---

## Next Steps

- **[Sequential Workflow](./workflow-sequential.md)** - Chain streaming agents
- **[Custom Handlers](./custom-handlers.md)** - Add streaming to custom handlers
- **[Performance](../performance.md)** - Optimize streaming performance

---

## Related Documentation

- [Streaming Guide](../streaming.md) - Complete streaming documentation
- [Core Concepts](../core-concepts.md) - Understanding the Stream interface
- [Performance](../performance.md) - Streaming optimization strategies
- [Troubleshooting](../troubleshooting.md) - Common streaming issues
