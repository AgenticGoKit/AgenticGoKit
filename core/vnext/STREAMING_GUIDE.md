# Streaming Guide

## Overview

AgenticGoKit vNext provides a powerful streaming system that allows you to receive real-time chunks of data as your agents process requests. This is especially useful for:

- **Live UI updates**: Display responses as they're generated
- **Long-running operations**: Show progress without waiting for completion
- **Token-by-token output**: See LLM responses in real-time
- **Tool execution monitoring**: Track tool calls as they happen
- **Thought process visibility**: Observe agent reasoning in real-time

## Quick Start

### Basic Streaming

```go
package main

import (
    "context"
    "fmt"
    "log"

    vnext "github.com/kunalkushwaha/agenticgokit/core/vnext"
)

func main() {
    // Build an agent
    agent, err := vnext.PresetChatAgentBuilder().
        WithName("StreamingAgent").
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Run with streaming
    ctx := context.Background()
    stream, err := agent.RunStream(ctx, "Tell me a story")
    if err != nil {
        log.Fatal(err)
    }

    // Process chunks as they arrive
    for chunk := range stream.Chunks() {
        switch chunk.Type {
        case vnext.ChunkTypeText:
            fmt.Print(chunk.Content)
        case vnext.ChunkTypeDelta:
            fmt.Print(chunk.Delta)
        case vnext.ChunkTypeThought:
            fmt.Printf("\n[Thinking: %s]\n", chunk.Content)
        case vnext.ChunkTypeDone:
            fmt.Println("\n✓ Complete")
        }
    }

    // Get final result
    result, err := stream.Wait()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("\nFinal result: %s\n", result.Content)
}
```

## Stream Interface

The `Stream` interface provides several methods for working with streaming data:

```go
type Stream interface {
    // Chunks returns a channel of streaming chunks
    Chunks() <-chan *StreamChunk
    
    // Wait blocks until streaming completes and returns the final result
    Wait() (*Result, error)
    
    // Cancel stops the stream
    Cancel()
    
    // Metadata returns information about the stream
    Metadata() *StreamMetadata
    
    // AsReader returns an io.Reader for the stream
    AsReader() io.Reader
}
```

## Chunk Types

Streaming supports 8 different chunk types:

| Chunk Type | Description | Use Case |
|------------|-------------|----------|
| `ChunkTypeText` | Complete text content | Full messages or paragraphs |
| `ChunkTypeDelta` | Incremental text changes | Token-by-token streaming |
| `ChunkTypeThought` | Agent reasoning process | Show thinking/planning |
| `ChunkTypeToolCall` | Tool execution request | Display tool usage |
| `ChunkTypeToolResult` | Tool execution result | Show tool output |
| `ChunkTypeMetadata` | Additional information | Timestamps, agent info |
| `ChunkTypeError` | Error information | Handle failures gracefully |
| `ChunkTypeDone` | Stream completion | Signal end of stream |

## Advanced Usage

### Streaming with Options

```go
// Configure streaming behavior
stream, err := agent.RunStream(ctx, "Complex task",
    vnext.WithBufferSize(200),              // Larger buffer for high throughput
    vnext.WithThoughts(true),               // Include thought process
    vnext.WithToolCalls(true),              // Include tool executions
    vnext.WithMetadata(true),               // Include metadata chunks
    vnext.WithFlushInterval(50*time.Millisecond), // Flush every 50ms
    vnext.WithTimeout(5*time.Minute),       // 5 minute timeout
)
```

### Streaming with Callback Handler

For more control, use a callback handler:

```go
handler := func(chunk *vnext.StreamChunk) error {
    switch chunk.Type {
    case vnext.ChunkTypeDelta:
        // Send to websocket, update UI, etc.
        sendToWebSocket(chunk.Delta)
    case vnext.ChunkTypeToolCall:
        log.Printf("Calling tool: %s(%v)", chunk.ToolName, chunk.ToolArgs)
    case vnext.ChunkTypeError:
        log.Printf("Error: %v", chunk.Error)
        return chunk.Error // Stop streaming on error
    }
    return nil
}

// Use the handler
result, err := agent.Run(ctx, "Your query", vnext.WithStreamHandler(handler))
```

### Text-Only Streaming

If you only need text output without thoughts or tool calls:

```go
stream, err := agent.RunStream(ctx, "Simple query",
    vnext.WithTextOnly(true), // Only receive text chunks
)

for chunk := range stream.Chunks() {
    // Only ChunkTypeText and ChunkTypeDelta will arrive
    fmt.Print(chunk.Delta)
}
```

### Stream to io.Reader

Convert a stream to an `io.Reader` for integration with existing code:

```go
stream, err := agent.RunStream(ctx, "Query")
if err != nil {
    log.Fatal(err)
}

reader := stream.AsReader()

// Use with any function expecting io.Reader
io.Copy(os.Stdout, reader)
```

## Workflow Streaming

Workflows also support streaming, providing visibility into multi-step execution:

```go
// Create a workflow
workflow, err := vnext.NewSequentialWorkflow("DataPipeline",
    vnext.Step("extract", extractAgent, "Extract data from source"),
    vnext.Step("transform", transformAgent, "Transform the data"),
    vnext.Step("load", loadAgent, "Load data to destination"),
)
if err != nil {
    log.Fatal(err)
}

// Run with streaming
stream, err := workflow.RunStream(ctx, "Process dataset.csv")
if err != nil {
    log.Fatal(err)
}

// Track progress through the workflow
for chunk := range stream.Chunks() {
    if chunk.Type == vnext.ChunkTypeMetadata {
        if stepName, ok := chunk.Metadata["step_name"].(string); ok {
            fmt.Printf("Executing step: %s\n", stepName)
        }
    }
}
```

## Stream Utilities

### Collect Stream to String

```go
stream, err := agent.RunStream(ctx, "Query")
if err != nil {
    log.Fatal(err)
}

// Collect all text into a single string
fullText, err := vnext.CollectStream(stream)
if err != nil {
    log.Fatal(err)
}
fmt.Println(fullText)
```

### Print Stream to Console

```go
stream, err := agent.RunStream(ctx, "Query")
if err != nil {
    log.Fatal(err)
}

// Print all chunks to stdout
vnext.PrintStream(stream)
```

### Convert Stream to Channel

```go
stream, err := agent.RunStream(ctx, "Query")
if err != nil {
    log.Fatal(err)
}

// Get a simple text channel
textChannel := vnext.StreamToChannel(stream)

for text := range textChannel {
    fmt.Print(text)
}
```

## Stream Builder

For custom streaming scenarios, use the StreamBuilder:

```go
// Create a custom stream
stream := vnext.NewStreamBuilder().
    WithBufferSize(100).
    WithMetadata(&vnext.StreamMetadata{
        AgentName: "CustomAgent",
        SessionID: "session-123",
    }).
    Build()

// Get the writer to emit chunks
writer := stream.(vnext.StreamWriter)

// Emit chunks
writer.Write(&vnext.StreamChunk{
    Type:    vnext.ChunkTypeDelta,
    Delta:   "Hello",
    Timestamp: time.Now(),
})

writer.Write(&vnext.StreamChunk{
    Type:    vnext.ChunkTypeDelta,
    Delta:   " world!",
    Timestamp: time.Now(),
})

// Close when done
writer.Close()
```

## Configuration

### In Code

```go
opts := []vnext.StreamOption{
    vnext.WithBufferSize(200),
    vnext.WithThoughts(true),
    vnext.WithToolCalls(true),
    vnext.WithTimeout(5 * time.Minute),
}

stream, err := agent.RunStream(ctx, "Query", opts...)
```

### In Configuration File (TOML)

```toml
[streaming]
enabled = true
buffer_size = 200
flush_interval_ms = 100
timeout_ms = 300000
include_thoughts = true
include_tool_calls = true
include_metadata = true
text_only = false
```

Load and use:

```go
config, err := vnext.LoadConfig("config.toml")
if err != nil {
    log.Fatal(err)
}

// Configuration is automatically applied when building agents
agent, err := vnext.NewAgentFromConfig(config)
```

## Performance Considerations

### Buffer Sizing

- **Small buffers (10-50)**: Lower latency, more overhead
- **Medium buffers (50-200)**: Balanced performance (recommended)
- **Large buffers (200+)**: Higher throughput, slightly higher latency

```go
// For real-time chat UI
stream, err := agent.RunStream(ctx, query, vnext.WithBufferSize(50))

// For batch processing
stream, err := agent.RunStream(ctx, query, vnext.WithBufferSize(500))
```

### Flush Intervals

Control how often chunks are sent:

```go
// Immediate updates (more CPU usage)
vnext.WithFlushInterval(10 * time.Millisecond)

// Balanced (recommended)
vnext.WithFlushInterval(100 * time.Millisecond)

// Batched updates (less CPU usage)
vnext.WithFlushInterval(500 * time.Millisecond)
```

### Memory Management

Streaming is designed to be memory-efficient:

- Chunks are processed and released immediately
- No buffering of full responses in memory
- Goroutines are cleaned up automatically
- Use `stream.Cancel()` for early termination

```go
stream, err := agent.RunStream(ctx, query)
if err != nil {
    log.Fatal(err)
}

// Cancel if needed
go func() {
    time.Sleep(5 * time.Second)
    stream.Cancel()
}()
```

## Error Handling

### Handling Errors in Chunks

```go
for chunk := range stream.Chunks() {
    if chunk.Type == vnext.ChunkTypeError {
        log.Printf("Stream error: %v", chunk.Error)
        // Stream will close after error chunk
        break
    }
}

// Always check final result
result, err := stream.Wait()
if err != nil {
    log.Printf("Final error: %v", err)
}
```

### Graceful Degradation

```go
stream, err := agent.RunStream(ctx, query)
if err != nil {
    // Fall back to non-streaming
    result, err := agent.Run(ctx, query)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.Content)
    return
}

// Process stream normally
for chunk := range stream.Chunks() {
    fmt.Print(chunk.Delta)
}
```

## Best Practices

### 1. Always Check Final Result

```go
stream, err := agent.RunStream(ctx, query)
if err != nil {
    log.Fatal(err)
}

// Process chunks
for chunk := range stream.Chunks() {
    // Handle chunks...
}

// IMPORTANT: Check final result for errors
result, err := stream.Wait()
if err != nil {
    log.Printf("Stream completed with error: %v", err)
}
```

### 2. Use Context for Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream, err := agent.RunStream(ctx, query)
// Stream will automatically cancel when context times out
```

### 3. Filter Chunks Based on Needs

```go
// For UI display, filter what users see
for chunk := range stream.Chunks() {
    switch chunk.Type {
    case vnext.ChunkTypeDelta:
        // Show to user
        displayInUI(chunk.Delta)
    case vnext.ChunkTypeThought:
        // Log for debugging, don't show to user
        log.Printf("Internal thought: %s", chunk.Content)
    case vnext.ChunkTypeToolCall:
        // Show loading indicator
        showLoadingIndicator(chunk.ToolName)
    }
}
```

### 4. Use Appropriate Buffer Sizes

```go
// Real-time chat (low latency)
vnext.WithBufferSize(50)

// Data processing (high throughput)
vnext.WithBufferSize(500)

// Default (balanced)
// Don't specify - uses default of 100
```

### 5. Handle Backpressure

```go
stream, err := agent.RunStream(ctx, query, vnext.WithBufferSize(100))
if err != nil {
    log.Fatal(err)
}

// Process chunks with rate limiting
rateLimiter := time.NewTicker(10 * time.Millisecond)
defer rateLimiter.Stop()

for chunk := range stream.Chunks() {
    <-rateLimiter.C // Wait for rate limiter
    processChunk(chunk)
}
```

## Examples

### Example 1: Interactive Chat UI

```go
func handleChatMessage(w http.ResponseWriter, r *http.Request) {
    // Set up SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    query := r.URL.Query().Get("message")
    
    stream, err := agent.RunStream(r.Context(), query,
        vnext.WithTextOnly(true),
        vnext.WithBufferSize(50),
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    flusher := w.(http.Flusher)
    
    for chunk := range stream.Chunks() {
        if chunk.Type == vnext.ChunkTypeDelta {
            fmt.Fprintf(w, "data: %s\n\n", chunk.Delta)
            flusher.Flush()
        }
    }
}
```

### Example 2: Progress Tracking

```go
func processLargeDataset(ctx context.Context, agent vnext.Agent) error {
    stream, err := agent.RunStream(ctx, "Analyze dataset.csv",
        vnext.WithThoughts(true),
        vnext.WithToolCalls(true),
    )
    if err != nil {
        return err
    }

    progressBar := createProgressBar()
    
    for chunk := range stream.Chunks() {
        switch chunk.Type {
        case vnext.ChunkTypeThought:
            progressBar.SetStatus(chunk.Content)
        case vnext.ChunkTypeToolCall:
            progressBar.IncrementStep()
            log.Printf("Executing: %s", chunk.ToolName)
        case vnext.ChunkTypeToolResult:
            log.Printf("Tool result: %v", chunk.Content)
        }
    }
    
    result, err := stream.Wait()
    if err != nil {
        return err
    }
    
    progressBar.Complete()
    log.Printf("Analysis complete: %s", result.Content)
    return nil
}
```

### Example 3: Logging and Monitoring

```go
func runWithMonitoring(ctx context.Context, agent vnext.Agent, query string) (*vnext.Result, error) {
    metrics := &StreamMetrics{
        StartTime: time.Now(),
    }
    
    stream, err := agent.RunStream(ctx, query,
        vnext.WithMetadata(true),
    )
    if err != nil {
        return nil, err
    }

    for chunk := range stream.Chunks() {
        metrics.ChunkCount++
        
        switch chunk.Type {
        case vnext.ChunkTypeDelta:
            metrics.TokenCount += len(chunk.Delta)
        case vnext.ChunkTypeToolCall:
            metrics.ToolCalls++
            logToolCall(chunk.ToolName, chunk.ToolArgs)
        case vnext.ChunkTypeError:
            metrics.Errors++
            alertOnError(chunk.Error)
        case vnext.ChunkTypeMetadata:
            updateDashboard(chunk.Metadata)
        }
    }
    
    result, err := stream.Wait()
    metrics.Duration = time.Since(metrics.StartTime)
    
    sendMetrics(metrics)
    return result, err
}
```

## Troubleshooting

### Stream Hangs or Doesn't Complete

```go
// Always use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream, err := agent.RunStream(ctx, query)
```

### Memory Leaks

```go
// Always consume or cancel streams
stream, err := agent.RunStream(ctx, query)
if err != nil {
    log.Fatal(err)
}

// If not consuming all chunks, must cancel
defer stream.Cancel()
```

### Slow Streaming Performance

```go
// Increase buffer size
stream, err := agent.RunStream(ctx, query, vnext.WithBufferSize(500))

// Increase flush interval
stream, err := agent.RunStream(ctx, query, 
    vnext.WithFlushInterval(200*time.Millisecond))
```

### Missing Chunks

```go
// Ensure you're consuming all chunks
for chunk := range stream.Chunks() {
    // Process each chunk
    processChunk(chunk)
}

// Don't forget to wait
result, err := stream.Wait()
```

## See Also

- [README.md](README.md) - Main documentation
- [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) - Migrating from old APIs
- [examples/streaming/](../../examples/streaming/) - Complete examples
