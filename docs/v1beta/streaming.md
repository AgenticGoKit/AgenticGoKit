# Streaming

Receive LLM responses token-by-token in real time with v1beta's streaming API. Track tool calls, reasoning steps, and metadata as execution progresses.

---

## What you get

- Token-by-token deltas for live UI updates
- Tool execution and thought visibility
- Cancellation support via context
- Final aggregated result after streaming completes
- Chunk types for content, metadata, errors, and completion

---

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    agent, err := v1beta.NewBuilder("StreamingAgent").
        WithPreset(v1beta.ChatAgent).
        WithLLM("openai", "gpt-4o-mini").
        Build()
    if err != nil {
        log.Fatal(err)
    }

    stream, err := agent.RunStream(context.Background(), "Tell me a short story")
    if err != nil {
        log.Fatal(err)
    }

    for chunk := range stream.Chunks() {
        switch chunk.Type {
        case v1beta.ChunkTypeDelta:
            fmt.Print(chunk.Delta)
        case v1beta.ChunkTypeDone:
            fmt.Println("\n✓ Complete")
        case v1beta.ChunkTypeError:
            fmt.Println("\nError:", chunk.Error)
        }
    }

    result, _ := stream.Wait()
    fmt.Println("Tokens:", result.TokensUsed)
}
```

---

## Chunk types

### ChunkTypeDelta

Incremental text tokens.

```go
if chunk.Type == v1beta.ChunkTypeDelta {
    fmt.Print(chunk.Delta)  // "This", " is", " streaming"
}
```

### ChunkTypeText

Complete text blocks (paragraphs or full messages).

```go
if chunk.Type == v1beta.ChunkTypeText {
    fmt.Println(chunk.Content)  // Full paragraph
}
```

### ChunkTypeThought

Agent reasoning or internal process.

```go
if chunk.Type == v1beta.ChunkTypeThought {
    fmt.Printf("[Thinking: %s]\n", chunk.Content)
}
```

### ChunkTypeToolCall

Tool execution request with name and arguments.

```go
if chunk.Type == v1beta.ChunkTypeToolCall {
    toolName := chunk.Metadata["tool_name"].(string)
    fmt.Printf("Calling: %s\n", toolName)
}
```

### ChunkTypeToolRes

Tool execution result.

```go
if chunk.Type == v1beta.ChunkTypeToolRes {
    result := chunk.Metadata["result"]
    fmt.Printf("Result: %v\n", result)
}
```

### ChunkTypeMetadata

Execution metadata (tokens, timestamps, etc.).

```go
if chunk.Type == v1beta.ChunkTypeMetadata {
    if tokens, ok := chunk.Metadata["tokens"].(int); ok {
        fmt.Printf("Tokens: %d\n", tokens)
    }
}
```

### ChunkTypeError

Streaming errors.

```go
if chunk.Type == v1beta.ChunkTypeError {
    log.Println("Error:", chunk.Error)
}
```

### ChunkTypeDone

Stream completion marker.

```go
if chunk.Type == v1beta.ChunkTypeDone {
    fmt.Println("Stream complete")
}
```

### ChunkTypeAgentStart

Workflow/step execution begins (for multi-agent workflows).

```go
if chunk.Type == v1beta.ChunkTypeAgentStart {
    stepName := chunk.Metadata["agent_name"].(string)
    fmt.Printf("→ Starting: %s\n", stepName)
}
```

### ChunkTypeAgentComplete

Workflow/step execution completes (for multi-agent workflows).

```go
if chunk.Type == v1beta.ChunkTypeAgentComplete {
    stepName := chunk.Metadata["agent_name"].(string)
    fmt.Printf("✓ Completed: %s\n", stepName)
}
```

### Multimodal chunk types

For agents generating images, audio, or video:

- `ChunkTypeImage` - Image content (url or base64)
- `ChunkTypeAudio` - Audio content (url or base64)
- `ChunkTypeVideo` - Video content (url or base64)

```go
if chunk.Type == v1beta.ChunkTypeImage {
    imageData := chunk.Metadata["image"]
    // Use imageData.URL or imageData.Base64
}
```

---

## Stream interface

```go
type Stream interface {
    Chunks() <-chan *StreamChunk         // Channel of stream chunks
    Wait() (*Result, error)               // Block until complete, return final result
    Cancel()                              // Cancel the stream
    Metadata() *StreamMetadata            // Stream metadata
    AsReader() io.Reader                  // io.Reader for text streaming
}
```

Access metadata during streaming:

```go
stream, _ := agent.RunStream(ctx, query)
metadata := stream.Metadata()
fmt.Println("Agent:", metadata.AgentName)
fmt.Println("Trace ID:", metadata.TraceID)
```

Get final aggregated result:

```go
for chunk := range stream.Chunks() {
    fmt.Print(chunk.Delta)
}
result, _ := stream.Wait()
fmt.Println("Final:", result.FinalOutput)
```

Use as io.Reader:

```go
import "io"

stream, _ := agent.RunStream(ctx, query)
io.Copy(os.Stdout, stream.AsReader())
```

---

## Context and cancellation

### Timeout

```go
import (
    "context"
    "time"
)

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream, err := agent.RunStream(ctx, "Long query")
if err != nil {
    log.Fatal(err)
}

for chunk := range stream.Chunks() {
    fmt.Print(chunk.Delta)
}

if _, err := stream.Wait(); err == context.DeadlineExceeded {
    fmt.Println("Timed out")
}
```

### Manual cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

stream, _ := agent.RunStream(ctx, query)

go func() {
    <-userStopSignal
    stream.Cancel()
}()

for chunk := range stream.Chunks() {
    fmt.Print(chunk.Delta)
}
```

### Conditional stop

```go
import "strings"

stream, _ := agent.RunStream(ctx, query)
var response strings.Builder

for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeDelta {
        response.WriteString(chunk.Delta)
        if response.Len() > 1000 {
            stream.Cancel()
            break
        }
    }
}
```

---

## Workflow streaming

Track multi-agent execution:

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/agenticgokit/agenticgokit/v1beta"
)

config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 120 * time.Second,
}
workflow, _ := v1beta.NewSequentialWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "extract", Agent: extractAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "transform", Agent: transformAgent})

stream, _ := workflow.RunStream(context.Background(), "Process data")

for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeMetadata:
        if step, ok := chunk.Metadata["step_name"].(string); ok {
            fmt.Printf("→ Step: %s\n", step)
        }
    case v1beta.ChunkTypeDelta:
        fmt.Print(chunk.Delta)
    case v1beta.ChunkTypeDone:
        fmt.Println("\nWorkflow complete")
    }
}
```

---

## Integration patterns

### HTTP Server-Sent Events

```go
import (
    "fmt"
    "net/http"
)

func handleSSE(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, _ := w.(http.Flusher)

    stream, _ := agent.RunStream(r.Context(), r.URL.Query().Get("q"))

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

### WebSocket

```go
import "github.com/gorilla/websocket"

func streamToWebSocket(ws *websocket.Conn, agent v1beta.Agent, query string) {
    stream, err := agent.RunStream(context.Background(), query)
    if err != nil {
        ws.WriteJSON(map[string]interface{}{"type": "error", "error": err.Error()})
        return
    }

    for chunk := range stream.Chunks() {
        switch chunk.Type {
        case v1beta.ChunkTypeDelta:
            ws.WriteJSON(map[string]interface{}{"type": "content", "content": chunk.Delta})
        case v1beta.ChunkTypeDone:
            ws.WriteJSON(map[string]interface{}{"type": "done"})
        }
    }
}
```

---

## Common patterns

### Filter by chunk type

```go
import "strings"

stream, _ := agent.RunStream(ctx, query)
var response strings.Builder

for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeDelta {
        response.WriteString(chunk.Delta)
    }
}

fmt.Println(response.String())
```

### Track tools and thoughts

```go
stream, _ := agent.RunStream(ctx, query)

for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeToolCall:
        fmt.Println("Tool:", chunk.Metadata["tool_name"])
    case v1beta.ChunkTypeThought:
        fmt.Println("Thinking:", chunk.Content)
    case v1beta.ChunkTypeDelta:
        fmt.Print(chunk.Delta)
    }
}
```

---

## Troubleshooting

- Stream never completes: ensure you drain `stream.Chunks()` and call `stream.Wait()`.
- Context canceled: increase timeout or check for early cancellation.
- Missing chunks: handle all ChunkType cases in your switch.
- Memory issues: always use `defer cancel()` when creating contexts.

---

## Next steps

- [workflows](workflows.md) for multi-agent streaming
- [tool-integration](tool-integration.md) for tool call visibility
- [core-concepts](core-concepts.md) for agent fundamentals
