# Troubleshooting

Common issues and solutions when working with AgenticGoKit v1beta.

---

## Configuration Issues

### Config File Not Found

**Error:** `CONFIG_NOT_FOUND: configuration file not found`

**Solutions:**
```go
// Use absolute path
config, err := v1beta.LoadConfigFromTOML("/absolute/path/to/config.toml")

// Or use code-based configuration
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    Build()
```

### Invalid Configuration Values

**Error:** `CONFIG_VALIDATION: temperature must be between 0.0 and 2.0`

**Valid ranges:**
- Temperature: 0.0 - 2.0
- MaxTokens: 1 - model limit

**Fix in TOML:**
```toml
[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7      # Range: 0.0-2.0
max_tokens = 1000
```

### Missing API Keys

**Error:** `LLM_AUTH: API key not found or invalid`

**Solutions:**
```bash
# Set environment variables
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="..."

# Or in TOML
[llm]
api_key = "${OPENAI_API_KEY}"
```

---

## LLM Issues

### OpenAI Rate Limiting

**Error:** `LLM_RATE_LIMITED: rate limit exceeded`

**Solution: Implement retry with backoff**
```go
func runWithRetry(agent v1beta.Agent, ctx context.Context, input string) (*v1beta.Result, error) {
    for attempt := 1; attempt <= 5; attempt++ {
        result, err := agent.Run(ctx, input)
        if err == nil {
            return result, nil
        }
        
        if strings.Contains(err.Error(), "RATE_LIMITED") {
            backoff := time.Duration(attempt*attempt) * time.Second
            time.Sleep(backoff)
            continue
        }
        
        return nil, err
    }
    return nil, fmt.Errorf("max retries exceeded")
}
```

### Ollama Model Not Found

**Error:** `LLM_CALL_FAILED: model not found`

**Solutions:**
```bash
ollama list
ollama pull llama2
```

**Verify in code:**
```go
agent, _ := v1beta.NewChatAgent("agent",
    v1beta.WithLLM("ollama", "llama2"),
)
```

### Connection Timeout

**Error:** `LLM_TIMEOUT: request timeout`

**Solution:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

result, err := agent.Run(ctx, query)
```

---

## Streaming Issues

### Stream Hangs

**Problem:** `stream.Wait()` never returns, no chunks received.

**Solutions:**
```go
// Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream, err := agent.RunStream(ctx, query)
if err != nil {
    return err
}

// Always consume all chunks or cancel
defer stream.Cancel()

for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeError {
        stream.Cancel()
        break
    }
}

result, err := stream.Wait()
```

### Memory Leak with Streams

**Problem:** Memory grows over time, goroutines accumulate.

**Solution:**
```go
// Always defer Cancel on unused streams
stream, err := agent.RunStream(ctx, query)
if err != nil {
    return err
}

defer stream.Cancel() // Safety net

for chunk := range stream.Chunks() {
    // Process chunk
}
```

### Slow Streaming

**Problem:** Chunks arrive slowly, high latency.

**Solutions:**
```go
// Reduce buffer size (faster first chunk)
stream, _ := agent.RunStream(ctx, query,
    v1beta.WithBufferSize(25),
)

// Use context cancellation for cleanup
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

---

## Tool Issues

### MCP Server Not Connecting

**Error:** `MCP_CONNECTION: failed to connect to MCP server`

**Solutions:**
```go
// Check server is running and properly configured
agent, _ := v1beta.NewBuilder("agent").
    WithTools(
        v1beta.WithMCP(&v1beta.MCPServer{
            Name:    "filesystem",
            Type:    "stdio",
            Command: "mcp-server-filesystem",
            Enabled: true,
        }),
    ).
    Build()
```

### Tool Not Found

**Error:** `TOOL_NOT_FOUND: tool 'calculator' not found`

**Solution:**
```go
// Register tools via builder
agent, _ := v1beta.NewBuilder("agent").
    WithTools(
        v1beta.WithMCP(servers...),
    ).
    Build()
```

### Tool Timeout

**Error:** `TOOL_TIMEOUT: tool execution exceeded timeout`

**Solution:**
```go
agent, _ := v1beta.NewBuilder("agent").
    WithTools(
        v1beta.WithToolTimeout(60 * time.Second),
    ).
    Build()
```

---

## Memory Issues

### Memory Connection Failed

**Error:** `MEMORY_CONNECTION: failed to connect to pgvector`

**Solutions:**
```go
// Use fallback provider
agent, _ := v1beta.NewBuilder("agent").
    WithMemory(
        v1beta.WithMemoryProvider("chromem"),
    ).
    Build()

// Or configure PostgreSQL
// psql "postgresql://user:password@localhost:5432/dbname"
// CREATE EXTENSION IF NOT EXISTS vector;
```

### Memory Not Persisting

**Problem:** Agent doesn't remember previous conversations.

**Solution:**
```go
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithSessionScoped(),
    ).
    Build()
```

---

## Workflow Issues

### Circular Dependency

**Error:** `WORKFLOW_CYCLE_DETECTED: circular dependency found`

**Solution:** Ensure dependencies form a DAG (no cycles):
```go
// Good: Linear chain A → B → C
workflow, _ := v1beta.NewDAGWorkflow(&v1beta.WorkflowConfig{
    Mode: v1beta.DAG,
})

workflow.AddStep(v1beta.WorkflowStep{
    Name:         "a",
    Agent:        agentA,
    Dependencies: nil,
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "b",
    Agent:        agentB,
    Dependencies: []string{"a"},
})
```

### Workflow Step Failed

**Error:** `WORKFLOW_STEP_FAILED: step 'analysis' failed`

**Solution:**
```go
result, err := workflow.Run(ctx, input)
if err != nil {
    for _, stepResult := range result.StepResults {
        if !stepResult.Success {
            log.Printf("Step %s failed: %s", stepResult.StepName, stepResult.Error)
        }
    }
}
```

---

## Build and Runtime Issues

### Import Errors

**Error:**
```
cannot find package "github.com/agenticgokit/agenticgokit/v1beta"
```

**Solutions:**
```bash
go get -u github.com/agenticgokit/agenticgokit/v1beta
go mod tidy
```

### Plugin Not Loading

**Error:**
```
plugin not found: memory/pgvector
```

**Solution:**
```go
import (
    "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/pgvector"
    _ "github.com/agenticgokit/agenticgokit/plugins/llm/openai"
)
```

---

## Common Error Messages

### "context deadline exceeded"

**Cause:** Operation took too long

**Solution:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

result, err := agent.Run(ctx, query)
```

### "API key not found"

**Cause:** Missing or incorrect API key

**Solution:**
```bash
export OPENAI_API_KEY="sk-..."
```

### "model not found"

**Cause:** Invalid model name

**Solution:**
```go
agent, _ := v1beta.NewChatAgent("agent",
    v1beta.WithLLM("openai", "gpt-4"),
)
```

### "rate limit exceeded"

**Cause:** Too many requests

**Solution:** Implement exponential backoff (see LLM Issues)

### "connection refused"

**Cause:** Service not running or wrong address

**Solution:**
```bash
netstat -an | grep 8080
curl http://localhost:8080
```

---

## Debugging Tips

### Check Agent Configuration

```go
agent, _ := v1beta.NewChatAgent("agent",
    v1beta.WithLLM("openai", "gpt-4"),
)

config := agent.Config()
log.Printf("LLM: %s/%s", config.LLM.Provider, config.LLM.Model)

caps := agent.Capabilities()
log.Printf("Capabilities: %v", caps)
```

### Test Components Individually

```go
// Test LLM connection
agent, err := v1beta.NewBuilder("test").
    WithLLM("openai", "gpt-4").
    Build()

result, err := agent.Run(context.Background(), "Hello")
if err != nil {
    log.Printf("LLM test failed: %v", err)
}
```

---

## Getting Help

### Check Documentation

- [Getting Started](./getting-started.md) - Basic setup
- [Core Concepts](./core-concepts.md) - Architecture
- [Streaming](./streaming.md) - Real-time patterns
- [Configuration](./configuration.md) - All settings

### Report Issues

When reporting, include:
- Error message and code
- Minimal reproduction code
- Configuration (redact sensitive info)
- Go version and OS

---

**Still stuck?** Check [GitHub Issues](https://github.com/agenticgokit/agenticgokit/issues).
