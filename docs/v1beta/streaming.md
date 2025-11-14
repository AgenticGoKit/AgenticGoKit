# Streaming Guide

Real-time streaming is a core feature of AgenticGoKit, allowing you to receive responses as they're generated. This guide covers streaming patterns, chunk types, and best practices.

---

## 🎯 Why Streaming?

Streaming provides several advantages:

- **Live UI updates** - Display responses as they're generated
- **Better UX** - Show progress without waiting for completion
- **Token-by-token output** - See LLM responses in real-time
- **Tool execution visibility** - Track tool calls as they happen
- **Thought process transparency** - Observe agent reasoning
- **Early cancellation** - Stop processing when needed

---

## 🚀 Quick Start

### Basic Stream Interface

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
        WithPreset(v1beta.ChatAgent).
        Build()
    if err != nil {
        log.Fatal(err)
    }
    
    // Start streaming
    stream, err := agent.RunStream(context.Background(), "Tell me a story")
    if err != nil {
        log.Fatal(err)
    }
    
    // Process chunks as they arrive
    for chunk := range stream.Chunks() {
        switch chunk.Type {
        case v1beta.ChunkTypeDelta:
            fmt.Print(chunk.Delta)
        case v1beta.ChunkTypeContent:
            fmt.Print(chunk.Content)
        case v1beta.ChunkTypeThought:
            fmt.Printf("\n[Thinking: %s]\n", chunk.Content)
        case v1beta.ChunkTypeDone:
            fmt.Println("\n✓ Complete")
        case v1beta.ChunkTypeError:
            fmt.Println("\n✗ Error:", chunk.Error)
        }
    }
    
    // Get final result
    result, err := stream.Wait()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("\nFinal output:", result.FinalOutput)
}
```

### Stream with Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream, err := agent.RunStream(ctx, "Explain quantum computing")
if err != nil {
    log.Fatal(err)
}

for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
}

result, err := stream.Wait()
if err == context.DeadlineExceeded {
    fmt.Println("\nStream timed out")
}
```

---

## 📦 Chunk Types

AgenticGoKit provides **8 chunk types** for different streaming scenarios:

### ChunkTypeContent
Complete text content (paragraphs or full messages).

```go
if chunk.Type == v1beta.ChunkTypeContent {
    fmt.Println(chunk.Content)  // "This is a complete paragraph."
}
```

**Use when:**
- You want full messages
- Displaying complete thoughts
- Non-incremental updates

### ChunkTypeDelta
Incremental text changes (token-by-token).

```go
if chunk.Type == v1beta.ChunkTypeDelta {
    fmt.Print(chunk.Delta)  // "This", " is", " incremental"
}
```

**Use when:**
- Typewriter effect in UI
- Real-time token streaming
- Maximum responsiveness

### ChunkTypeThought
Agent's internal reasoning process.

```go
if chunk.Type == v1beta.ChunkTypeThought {
    fmt.Printf("[Thinking: %s]\n", chunk.Content)
    // "Analyzing the question to determine the best approach..."
}
```

**Use when:**
- Debugging agent logic
- Showing "thinking" indicators
- Understanding agent decisions

### ChunkTypeToolCall
Tool execution request.

```go
if chunk.Type == v1beta.ChunkTypeToolCall {
    toolName := chunk.Metadata["tool_name"].(string)
    toolArgs := chunk.Metadata["tool_args"].(map[string]interface{})
    fmt.Printf("Calling tool: %s(%v)\n", toolName, toolArgs)
}
```

**Use when:**
- Tracking tool usage
- Showing "searching..." indicators
- Debugging tool calls

### ChunkTypeToolResult
Tool execution result.

```go
if chunk.Type == v1beta.ChunkTypeToolResult {
    result := chunk.Metadata["result"]
    fmt.Printf("Tool result: %v\n", result)
}
```

**Use when:**
- Displaying tool outputs
- Logging tool execution
- Debugging tool responses

### ChunkTypeMetadata
Additional information (timestamps, token counts, etc.).

```go
if chunk.Type == v1beta.ChunkTypeMetadata {
    if tokens, ok := chunk.Metadata["tokens"].(int); ok {
        fmt.Printf("Tokens used: %d\n", tokens)
    }
    if timestamp, ok := chunk.Metadata["timestamp"].(time.Time); ok {
        fmt.Printf("Time: %s\n", timestamp.Format(time.RFC3339))
    }
}
```

**Use when:**
- Tracking resource usage
- Performance monitoring
- Analytics and logging

### ChunkTypeError
Error information during streaming.

```go
if chunk.Type == v1beta.ChunkTypeError {
    fmt.Println("Error:", chunk.Error)
    // Handle error gracefully
}
```

**Use when:**
- Graceful error handling
- Showing error messages to users
- Logging failures

### ChunkTypeDone
Stream completion marker.

```go
if chunk.Type == v1beta.ChunkTypeDone {
    fmt.Println("✓ Stream complete")
    // Chunk may contain final metadata
    if result, ok := chunk.Metadata["final_result"]; ok {
        fmt.Println("Final:", result)
    }
}
```

**Use when:**
- Cleanup after streaming
- Displaying completion status
- Collecting final statistics

---

## 🎨 Streaming Patterns

### Pattern 1: Stream Interface (Recommended)

Most flexible pattern with full control:

```go
// Start streaming
stream, err := agent.RunStream(ctx, query)
if err != nil {
    log.Fatal(err)
}

// Process chunks
for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeContent, v1beta.ChunkTypeDelta:
        fmt.Print(chunk.Delta)
    case v1beta.ChunkTypeToolCall:
        handleToolCall(chunk)
    case v1beta.ChunkTypeDone:
        fmt.Println("\nDone!")
    }
}

// Get final result
result, err := stream.Wait()
if err != nil {
    log.Println("Error:", err)
}
```

**Pros:**
- Clean and simple API
- Automatic channel management
- Built-in result aggregation
- Context cancellation support

**Cons:**
- Less control over buffering
- Fixed chunk types

### Pattern 2: Filtered Streaming

Process only specific chunk types:

```go
stream, err := agent.RunStream(ctx, query)
if err != nil {
    log.Fatal(err)
}

// Filter for text content only
var response strings.Builder
for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeDelta {
        response.WriteString(chunk.Delta)
    }
}

fmt.Println(response.String())
```

**Pros:**
- Simple filtering
- Focused processing
- Clean text accumulation

**Cons:**
- May miss important metadata
- No tool/thought visibility

### Pattern 3: Stream Cancellation

Cancel stream based on conditions:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

stream, _ := agent.RunStream(ctx, query)

// Simple text accumulation with cancellation
var response strings.Builder
for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeDelta {
        response.WriteString(chunk.Delta)
        
        // Cancel if response too long
        if response.Len() > 1000 {
            stream.Cancel()
            break
        }
    }
}

fmt.Println(response.String())
```

**Pros:**
- Early termination support
- Resource control
- User-initiated stop

**Cons:**
- May lose partial results
- Needs cleanup handling

## ⚙️ Stream Interface Methods

### Stream Methods

```go
type Stream interface {
    // Chunks returns channel of stream chunks
    Chunks() <-chan *StreamChunk
    
    // Wait blocks until stream completes and returns final result
    Wait() (*Result, error)
    
    // Cancel cancels the stream
    Cancel()
    
    // Metadata returns stream metadata
    Metadata() map[string]interface{}
    
    // AsReader returns io.Reader interface for text streaming
    AsReader() io.Reader
}
```

### Using Stream Methods

```go
stream, _ := agent.RunStream(ctx, "Query")

// Get metadata during streaming
metadata := stream.Metadata()
fmt.Println("Stream ID:", metadata["stream_id"])

// Process chunks
for chunk := range stream.Chunks() {
    fmt.Print(chunk.Delta)
}

// Get final aggregated result
result, err := stream.Wait()
fmt.Println("Final:", result.FinalOutput)
```

### AsReader() for io.Reader Interface

```go
stream, _ := agent.RunStream(ctx, "Query")

// Use as io.Reader
reader := stream.AsReader()
io.Copy(os.Stdout, reader) // Stream directly to stdout
```

---

## 🎯 Best Practices

**Guidelines:**
- Always handle all chunk types for robustness
- Use context for timeouts and cancellation
- Process chunks quickly to avoid blocking
- Aggregate text for final output when needed
- Handle errors gracefully

### Context Usage

Use context for control:

```go
// Timeout after 30 seconds
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream, err := agent.RunStream(ctx, query)
if err != nil {
    log.Fatal(err)
}

for chunk := range stream.Chunks() {
    // Process chunks
}

result, err := stream.Wait()
if err == context.DeadlineExceeded {
    log.Println("Streaming timed out")
}
```

### User Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Start streaming
stream, _ := agent.RunStream(ctx, query)

// Cancel from another goroutine
go func() {
    <-userCancelSignal
    stream.Cancel() // Stop streaming immediately
}()

for chunk := range stream.Chunks() {
    fmt.Print(chunk.Delta)
}
```
```

---

## 🔄 Workflow Streaming

Workflows support streaming to track multi-agent execution:

```go
// Create workflow
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 120 * time.Second,
}
workflow, _ := v1beta.NewSequentialWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "extract", Agent: extractAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "transform", Agent: transformAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "load", Agent: loadAgent})

// Stream workflow execution
stream, err := workflow.RunStream(context.Background(), "Process data")
if err != nil {
    log.Fatal(err)
}

for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeMetadata:
        if stepName, ok := chunk.Metadata["step_name"].(string); ok {
            fmt.Printf("→ Executing step: %s\n", stepName)
        }
    case v1beta.ChunkTypeDelta:
        fmt.Print(chunk.Delta)
    case v1beta.ChunkTypeDone:
        fmt.Println("\n✓ Workflow complete")
    }
}

result, _ := stream.Wait()
fmt.Println("Final result:", result.FinalOutput)
```

---

## 🌐 Integration Examples

### WebSocket Integration

```go
func streamToWebSocket(ws *websocket.Conn, agent v1beta.Agent, query string) {
    stream, err := agent.RunStream(context.Background(), query)
    if err != nil {
        ws.WriteJSON(map[string]interface{}{"type": "error", "error": err.Error()})
        return
    }
    
    for chunk := range stream.Chunks() {
        if chunk.Type == v1beta.ChunkTypeDelta {
            ws.WriteJSON(map[string]interface{}{
                "type":    "content",
                "content": chunk.Delta,
            })
        } else if chunk.Type == v1beta.ChunkTypeDone {
            ws.WriteJSON(map[string]interface{}{
                "type": "done",
            })
        }
    }
}
```

### Server-Sent Events (SSE)

```go
func streamToSSE(w http.ResponseWriter, agent v1beta.Agent, query string) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    flusher, _ := w.(http.Flusher)
    
    stream, err := agent.RunStream(context.Background(), query)
    if err != nil {
        fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
        flusher.Flush()
        return
    }
    
    for chunk := range stream.Chunks() {
        if chunk.Type == v1beta.ChunkTypeDelta {
            fmt.Fprintf(w, "data: %s\n\n", chunk.Delta)
            flusher.Flush()
        } else if chunk.Type == v1beta.ChunkTypeDone {
            fmt.Fprintf(w, "data: [DONE]\n\n")
            flusher.Flush()
            break
        }
    }
}
```

### CLI Progress Bar

```go
import "github.com/schollz/progressbar/v3"

func streamWithProgress(agent v1beta.Agent, query string) {
    bar := progressbar.NewOptions(-1,
        progressbar.OptionSetDescription("Processing..."),
        progressbar.OptionSpinnerType(14),
    )
    
    stream, err := agent.RunStream(context.Background(), query)
    if err != nil {
        log.Fatal(err)
    }
    
    var response strings.Builder
    for chunk := range stream.Chunks() {
        bar.Add(1)
        if chunk.Type == v1beta.ChunkTypeDelta {
            response.WriteString(chunk.Delta)
        }
    }
    
    bar.Finish()
    fmt.Println("\n", response.String())
}
```
    }()
    
    var response strings.Builder
    for chunk := range chunks {
        bar.Add(1)
        if chunk.Type == v1beta.ChunkTypeDelta {
            response.WriteString(chunk.Content)
        } else if chunk.Type == v1beta.ChunkTypeDone {
            bar.Finish()
            fmt.Printf("\n%s\n", response.String())
        }
    }
}
```

---

## 🎯 Best Practices

### 1. Always Use Context

```go
// ✅ Good - with timeout
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

stream, _ := agent.RunStream(ctx, query)
```

### 2. Handle All Chunk Types

```go
// ✅ Good - handle all types
stream, _ := agent.RunStream(ctx, query)

for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeContent, v1beta.ChunkTypeDelta:
        handleText(chunk)
    case v1beta.ChunkTypeError:
        handleError(chunk)
    case v1beta.ChunkTypeDone:
        handleDone(chunk)
    default:
        // Log or ignore unknown types
    }
}
```

### 3. Use Stream.Wait() for Final Result

```go
// ✅ Good - get aggregated result
stream, _ := agent.RunStream(ctx, query)

for chunk := range stream.Chunks() {
    processChunk(chunk)
}

result, err := stream.Wait()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Final:", result.FinalOutput)
```

### 4. Set Appropriate Timeouts

```go
// ✅ Good - reasonable timeout
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

stream, _ := agent.RunStream(ctx, query)
```

### 5. Handle Errors Gracefully

```go
// ✅ Good - error handling
stream, _ := agent.RunStream(ctx, query)

for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeError {
        log.Printf("Stream error: %v", chunk.Error)
        // Show user-friendly message
        // Attempt recovery or cleanup
        break
    }
}

result, err := stream.Wait()
if err != nil {
    // Handle final error
}
```

---

## 🐛 Troubleshooting

### Issue: Stream Never Completes

**Cause**: Not calling stream.Wait() or not draining Chunks()

**Solution**: Always consume all chunks and call Wait()

```go
stream, _ := agent.RunStream(ctx, query)

// Drain chunks
for chunk := range stream.Chunks() {
    processChunk(chunk)
}

// Get final result
result, err := stream.Wait()
```

### Issue: Context Canceled Error

**Cause**: Context timeout or cancellation

**Solution**: Check context deadline and increase if needed

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

stream, err := agent.RunStream(ctx, query)
if err == context.DeadlineExceeded {
    fmt.Println("Increase timeout or optimize query")
}
```

### Issue: Missing Chunks

**Cause**: Not processing all chunk types

**Solution**: Handle all ChunkType values

```go
for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeDelta:
        // Handle delta
    case v1beta.ChunkTypeContent:
        // Handle content
    case v1beta.ChunkTypeMetadata:
        // Handle metadata
    case v1beta.ChunkTypeDone:
        // Handle completion
    }
}
```

**Solution**: Always check for Done chunk

```go
for chunk := range chunks {
    if chunk.Type == v1beta.ChunkTypeDone {
        break // Exit loop
    }
}
```

### Issue: Memory Leak

**Cause**: Goroutine leak or unclosed channels

**Solution**: Ensure proper cleanup

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel() // Always cleanup

chunks := make(chan v1beta.StreamChunk, 100)

go func() {
    defer close(chunks) // Close when done
    agent.RunStream(ctx, query, chunks)
}()
```

---

## 📚 Examples

See complete streaming examples:
- [Streaming Agent Example](./examples/streaming-agent.md)
- [WebSocket Streaming](./examples/websocket-streaming.md)
- [CLI Progress Display](./examples/cli-streaming.md)

---

## 🔗 Related Topics

- **[Core Concepts](./core-concepts.md)** - Understanding agents and handlers
- **[Workflows](./workflows.md)** - Multi-agent streaming
- **[Performance Guide](./performance.md)** - Optimize streaming performance

---

**Ready for workflows?** Continue to [Workflows Guide](./workflows.md) →
