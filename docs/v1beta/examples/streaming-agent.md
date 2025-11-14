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
        WithLLM("openai", "gpt-4").
        Build()
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Start streaming
    stream, err := agent.RunStream(context.Background(), "Explain Go concurrency in detail")
    if err != nil {
        log.Fatalf("Failed to start stream: %v", err)
    }
    defer stream.Close()

    fmt.Println("Streaming response:")
    fmt.Println("---")

    // Process stream chunks
    for {
        chunk, err := stream.Next()
        if err != nil {
            if err == v1beta.ErrStreamDone {
                break
            }
            log.Printf("Stream error: %v", err)
            break
        }

        // Handle different chunk types
        switch chunk.Type {
        case v1beta.ChunkTypeText:
            fmt.Print(chunk.Content)
        case v1beta.ChunkTypeDelta:
            fmt.Print(chunk.Content)
        case v1beta.ChunkTypeDone:
            fmt.Println("\n---")
            fmt.Println("Stream complete")
            return
        case v1beta.ChunkTypeError:
            log.Printf("Chunk error: %s", chunk.Content)
        }
    }
}
```

---

## Step-by-Step Breakdown

### 1. Create Streaming Agent

```go
agent, err := v1beta.NewBuilder("StreamingAgent").
    WithLLM("openai", "gpt-4").
    Build()
```

Same builder pattern as basic agents - streaming is available by default.

### 2. Start Stream

```go
stream, err := agent.RunStream(context.Background(), query)
if err != nil {
    log.Fatalf("Failed to start stream: %v", err)
}
defer stream.Close()
```

**Key Points:**
- `RunStream()` returns a `Stream` interface
- Always `defer stream.Close()` to clean up resources
- Stream starts immediately upon creation

### 3. Process Chunks

```go
for {
    chunk, err := stream.Next()
    if err != nil {
        if err == v1beta.ErrStreamDone {
            break // Normal completion
        }
        log.Printf("Stream error: %v", err)
        break
    }
    
    // Process chunk
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
for {
    chunk, err := stream.Next()
    if err != nil {
        if err == v1beta.ErrStreamDone {
            break
        }
        return err
    }

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
        return nil
    }
}
```

---

## Advanced Patterns

### Channel-Based Streaming

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
        defer stream.Close()

        for {
            chunk, err := stream.Next()
            if err != nil {
                if err != v1beta.ErrStreamDone {
                    errChan <- err
                }
                return
            }

            if chunk.Type == v1beta.ChunkTypeText || chunk.Type == v1beta.ChunkTypeDelta {
                textChan <- chunk.Content
            }
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
defer stream.Close()

// Process chunks with cancellation
go func() {
    time.Sleep(5 * time.Second)
    cancel() // Cancel after 5 seconds
}()

for {
    chunk, err := stream.Next()
    if err != nil {
        if err == context.Canceled {
            fmt.Println("\nStream cancelled")
            return
        }
        if err == v1beta.ErrStreamDone {
            break
        }
        log.Printf("Error: %v", err)
        return
    }
    
    fmt.Print(chunk.Content)
}
```

### Callback-Based Streaming

```go
func streamWithCallback(agent v1beta.Agent, query string, onChunk func(string)) error {
    stream, err := agent.RunStream(context.Background(), query)
    if err != nil {
        return err
    }
    defer stream.Close()

    for {
        chunk, err := stream.Next()
        if err != nil {
            if err == v1beta.ErrStreamDone {
                return nil
            }
            return err
        }

        if chunk.Type == v1beta.ChunkTypeText || chunk.Type == v1beta.ChunkTypeDelta {
            onChunk(chunk.Content)
        }
    }
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
    WithLLM("openai", "gpt-4").
    WithConfig(&v1beta.Config{
        StreamBufferSize: 100,  // Increase for high-throughput streams
        StreamFlushInterval: 50 * time.Millisecond,
    }).
    Build()
```

### Memory Management

```go
// Process large streams without loading entire response
var totalTokens int
for {
    chunk, err := stream.Next()
    if err != nil {
        break
    }
    
    // Process chunk immediately, don't accumulate
    fmt.Print(chunk.Content)
    
    // Track metadata
    if tokens, ok := chunk.Metadata["tokens"].(int); ok {
        totalTokens += tokens
    }
}
fmt.Printf("\nTotal tokens: %d\n", totalTokens)
```

---

## Error Handling

### Robust Stream Processing

```go
func processStream(stream v1beta.Stream) error {
    defer stream.Close()
    
    for {
        chunk, err := stream.Next()
        if err != nil {
            if err == v1beta.ErrStreamDone {
                return nil // Normal completion
            }
            
            // Check if retryable
            if v1beta.IsRetryable(err) {
                log.Printf("Retryable error: %v", err)
                // Implement retry logic
                continue
            }
            
            return fmt.Errorf("stream error: %w", err)
        }
        
        // Handle chunk
        if chunk.Type == v1beta.ChunkTypeError {
            log.Printf("Chunk error: %s", chunk.Content)
            continue
        }
        
        fmt.Print(chunk.Content)
    }
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
