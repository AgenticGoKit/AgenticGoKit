# Troubleshooting

Common issues and solutions when working with AgenticGoKit v1beta.

---

## 🎯 Quick Diagnostics

### Check Configuration

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

agent, err := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    Build()

if err != nil {
    fmt.Printf("Error: %v\n", err)
    fmt.Printf("Error code: %v\n", v1beta.GetErrorCode(err))
    fmt.Printf("Suggestion: %v\n", v1beta.GetErrorSuggestion(err))
}

// Check agent capabilities
caps := agent.Capabilities()
fmt.Printf("Capabilities: %v\n", caps)

// Check configuration
config := agent.Config()
fmt.Printf("LLM: %s/%s\n", config.LLM.Provider, config.LLM.Model)
```

### Enable Debug Mode

```go
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    WithConfig(&v1beta.Config{
        DebugMode: true, // Enable verbose logging
    }).
    Build()
```

---

## ⚙️ Configuration Issues

### Issue: Config File Not Found

**Error:**
```
CONFIG_NOT_FOUND: configuration file not found
```

**Solutions:**

1. Use absolute path:
```go
config, err := v1beta.LoadConfig("/absolute/path/to/config.toml")
```

2. Verify working directory:
```go
wd, _ := os.Getwd()
fmt.Printf("Working directory: %s\n", wd)
```

3. Use code-based configuration:
```go
config := &v1beta.Config{
    Name: "MyAgent",
    LLM: v1beta.LLMConfig{
        Provider: "openai",
        Model:    "gpt-4",
    },
}

agent, _ := v1beta.NewBuilder("agent").
    WithConfig(config).
    Build()
```

### Issue: Invalid Configuration Values

**Error:**
```
CONFIG_VALIDATION: temperature must be between 0.0 and 2.0
```

**Valid Ranges:**
- **Temperature**: 0.0 - 2.0
- **MaxTokens**: 1 - model limit
- **BufferSize**: 1 - 10000 (recommended: 50-500)
- **Timeout**: > 0

**Solution:**
```toml
[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7       # 0.0-2.0
max_tokens = 1000       # positive

[streaming]
buffer_size = 100       # positive
flush_interval_ms = 100 # positive
```

### Issue: Missing API Keys

**Error:**
```
LLM_AUTH: API key not found or invalid
```

**Solutions:**

1. Environment variables:
```bash
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="..."
export AZURE_OPENAI_KEY="..."
```

2. In configuration:
```toml
[llm]
provider = "openai"
model = "gpt-4"
api_key = "${OPENAI_API_KEY}"  # Use env var
```

3. Programmatically:
```go
os.Setenv("OPENAI_API_KEY", "sk-...")
```

---

## 📡 Streaming Issues

### Issue: Stream Hangs

**Symptoms:**
- `stream.Wait()` never returns
- No chunks received
- Program frozen

**Solutions:**

1. Use context with timeout:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream, err := agent.RunStream(ctx, query)
if err != nil {
    return err
}
```

2. Consume all chunks:
```go
for chunk := range stream.Chunks() {
    // Process each chunk
    if chunk.Type == "error" {
        log.Printf("Error: %v", chunk.Error)
        break
    }
}

result, err := stream.Wait() // Now completes
```

3. Check for errors in stream:
```go
stream, err := agent.RunStream(ctx, query)
if err != nil {
    return err
}

for chunk := range stream.Chunks() {
    if chunk.Error != "" {
        log.Printf("Stream error: %s", chunk.Error)
        stream.Cancel()
        break
    }
}
```

### Issue: Memory Leak with Streams

**Symptoms:**
- Memory grows over time
- Goroutines accumulate
- Application slows down

**Solutions:**

1. Always cancel unused streams:
```go
stream, err := agent.RunStream(ctx, query)
if err != nil {
    return err
}

// Cancel if not consuming
defer stream.Cancel()
```

2. Fully consume or cancel:
```go
stream, err := agent.RunStream(ctx, query)
defer stream.Cancel() // Safety net

for chunk := range stream.Chunks() {
    processChunk(chunk)
}
```

3. Use context cancellation:
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel() // Cleanup all resources

stream, err := agent.RunStream(ctx, query)
```

### Issue: Slow Streaming

**Symptoms:**
- Chunks arrive slowly
- High latency
- UI sluggish

**Solutions:**

1. Reduce buffer size:
```go
stream, _ := agent.RunStream(ctx, query,
    v1beta.WithBufferSize(25), // Smaller = faster first chunk
)
```

2. Reduce flush interval:
```go
stream, _ := agent.RunStream(ctx, query,
    v1beta.WithFlushInterval(50 * time.Millisecond),
)
```

3. Use text-only mode:
```go
stream, _ := agent.RunStream(ctx, query,
    v1beta.WithTextOnly(true), // Skip metadata
)
```

---

## 🤖 LLM Provider Issues

### Issue: OpenAI Rate Limiting

**Error:**
```
LLM_RATE_LIMITED: rate limit exceeded
```

**Solutions:**

1. Implement exponential backoff:
```go
import "github.com/agenticgokit/agenticgokit/v1beta"

func runWithRetry(agent v1beta.Agent, ctx context.Context, input string) (*v1beta.Result, error) {
    for attempt := 1; attempt <= 5; attempt++ {
        result, err := agent.Run(ctx, input)
        if err == nil {
            return result, nil
        }
        
        if v1beta.IsErrorCode(err, v1beta.ErrCodeLLMRateLimited) {
            backoff := time.Duration(attempt*attempt) * time.Second
            time.Sleep(backoff)
            continue
        }
        
        return nil, err
    }
    return nil, fmt.Errorf("max retries exceeded")
}
```

2. Configure rate limiting:
```toml
[llm]
rate_limit = 10  # requests per second
max_concurrent = 3
```

3. Use different tier or model:
```go
// Switch to higher tier model
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-3.5-turbo"). // Fewer restrictions
    Build()
```

### Issue: Azure OpenAI Connection

**Error:**
```
LLM_CONNECTION: failed to connect to Azure OpenAI
```

**Solutions:**

1. Verify endpoint and deployment:
```toml
[llm]
provider = "azure"
model = "your-deployment-name"
endpoint = "https://your-resource.openai.azure.com/"
api_key = "${AZURE_OPENAI_KEY}"
api_version = "2024-02-15-preview"
```

2. Check network access:
```go
// Test connection
resp, err := http.Get("https://your-resource.openai.azure.com/")
if err != nil {
    fmt.Printf("Network error: %v\n", err)
}
```

3. Verify credentials:
```bash
# Test with curl
curl -H "api-key: $AZURE_OPENAI_KEY" \
  https://your-resource.openai.azure.com/openai/deployments/your-deployment/chat/completions?api-version=2024-02-15-preview
```

### Issue: Ollama Model Not Found

**Error:**
```
LLM_CALL_FAILED: model not found
```

**Solutions:**

1. List available models:
```bash
ollama list
```

2. Pull the model:
```bash
ollama pull llama2
ollama pull mistral
```

3. Verify model name:
```go
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("ollama", "llama2"). // Must match exact name
    Build()
```

---

## 🔧 Tool Integration Issues

### Issue: MCP Server Not Connecting

**Error:**
```
MCP_CONNECTION: failed to connect to MCP server
```

**Solutions:**

1. Verify server is running:
```bash
# For stdio servers
which mcp-server-filesystem

# For TCP servers
curl http://localhost:8080/health
```

2. Check server configuration:
```go
mcpServer := v1beta.MCPServer{
    Name:    "filesystem",
    Type:    "stdio",
    Command: "mcp-server-filesystem", // Must be in PATH
    Enabled: true,
}

// Or for TCP
mcpServer := v1beta.MCPServer{
    Name:    "api",
    Type:    "tcp",
    Address: "localhost",
    Port:    8080, // Must be correct
    Enabled: true,
}
```

3. Enable debug logging:
```toml
[tools.mcp]
enabled = true
connection_timeout = "30s"
max_retries = 3
```

### Issue: Tool Not Found

**Error:**
```
TOOL_NOT_FOUND: tool 'calculator' not found
```

**Solutions:**

1. List available tools:
```go
import "github.com/agenticgokit/agenticgokit/v1beta"

handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    tools := capabilities.Tools.List()
    for _, tool := range tools {
        fmt.Printf("Tool: %s - %s\n", tool.Name, tool.Description)
    }
    return capabilities.LLM("You are a helpful assistant.", input)
}
```

2. Check MCP server health:
```go
handler := func(ctx context.Context, input string, capabilities *v1beta.Capabilities) (string, error) {
    health := capabilities.Tools.HealthCheck(ctx)
    for name, status := range health {
        fmt.Printf("Server %s: %s\n", name, status.Status)
        if status.Error != "" {
            fmt.Printf("  Error: %s\n", status.Error)
        }
    }
    return "", nil
}
```

3. Enable tool discovery:
```go
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCPDiscovery(8080, 8081, 8090),
    ).
    Build()
```

### Issue: Tool Timeout

**Error:**
```
TOOL_TIMEOUT: tool execution exceeded timeout
```

**Solutions:**

1. Increase timeout:
```go
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    WithTools(
        v1beta.WithMCP(servers...),
        v1beta.WithToolTimeout(60 * time.Second), // Increase
    ).
    Build()
```

2. Configure per-tool:
```toml
[tools]
timeout = "30s"

[tools.timeouts]
web_scraper = "120s"  # Slow tool
calculator = "5s"     # Fast tool
```

3. Optimize tool implementation:
```go
// Add caching to slow tools
// Use connection pooling
// Implement early returns
```

---

## 💾 Memory Issues

### Issue: Memory Connection Failed

**Error:**
```
MEMORY_CONNECTION: failed to connect to pgvector
```

**Solutions:**

1. Verify connection string:
```toml
[memory]
provider = "pgvector"
connection_string = "postgresql://user:password@localhost:5432/dbname"
```

2. Test connection:
```bash
psql "postgresql://user:password@localhost:5432/dbname" -c "\l"
```

3. Check pgvector extension:
```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

4. Use fallback provider:
```go
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("memory"), // Fallback to in-memory
    ).
    Build()
```

### Issue: Memory Not Persisting

**Symptoms:**
- Agent doesn't remember previous conversations
- Session data lost

**Solutions:**

1. Enable session-scoped memory:
```go
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("memory"),
        v1beta.WithSessionScoped(),
    ).
    Build()
```

2. Use persistent provider:
```go
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    WithMemory(
        v1beta.WithMemoryProvider("pgvector"),
        v1beta.WithSessionScoped(),
    ).
    Build()
```

3. Verify session ID is set:
```go
import "github.com/agenticgokit/agenticgokit/core"

provider := core.GetMemoryProvider("memory")
if provider != nil {
    provider.SetSession("user-123") // Set session ID
}
```

---

## 🔀 Workflow Issues

### Issue: Workflow Cycle Detected

**Error:**
```
WORKFLOW_CYCLE_DETECTED: circular dependency found
```

**Solution:**

Check workflow dependencies:

```go
// ❌ Bad: Circular dependency
workflow, _ := v1beta.NewDAGWorkflow("pipeline",
    v1beta.Step("a", agent1, "task", "b"), // depends on b
    v1beta.Step("b", agent2, "task", "a"), // depends on a
)

// ✅ Good: Linear dependencies
workflow, _ := v1beta.NewDAGWorkflow("pipeline",
    v1beta.Step("a", agent1, "task"),
    v1beta.Step("b", agent2, "task", "a"),
    v1beta.Step("c", agent3, "task", "b"),
)
```

### Issue: Workflow Step Failed

**Error:**
```
WORKFLOW_STEP_FAILED: step 'analysis' failed
```

**Solutions:**

1. Add error handling:
```go
workflow, _ := v1beta.NewSequentialWorkflow("pipeline",
    v1beta.Step("step1", agent1, "task"),
    v1beta.Step("step2_optional", agent2, "task"), // Can fail
    v1beta.Step("step3", agent3, "task"),
)

// Continue on non-critical failures
```

2. Use fallback agents:
```go
// Try primary, fallback to secondary
step := v1beta.Step("analysis", primaryAgent, "task")
// If fails, use secondaryAgent
```

3. Add retry logic:
```toml
[workflow]
max_retries = 3
retry_delay = "1s"
```

---

## 🏗️ Build and Runtime Issues

### Issue: Import Errors

**Error:**
```
cannot find package "github.com/agenticgokit/agenticgokit/v1beta"
```

**Solutions:**

1. Update dependencies:
```bash
go get -u github.com/agenticgokit/agenticgokit/v1beta
go mod tidy
```

2. Verify import path:
```go
import "github.com/agenticgokit/agenticgokit/v1beta"

// Not:
// import "github.com/agenticgokit/agenticgokit/core/vnext"
```

3. Check go.mod:
```
module your-app

go 1.21

require github.com/agenticgokit/agenticgokit v0.5.0
```

### Issue: Plugin Not Loading

**Error:**
```
plugin not found: memory/pgvector
```

**Solutions:**

1. Import plugin:
```go
import (
    "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/pgvector" // Register
)
```

2. Verify plugin is installed:
```bash
go list -m github.com/agenticgokit/agenticgokit/plugins/memory/pgvector
```

3. Use blank import for side effects:
```go
import (
    _ "github.com/agenticgokit/agenticgokit/plugins/llm/openai"
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/memory"
)
```

---

## 🐛 Common Error Messages

### "context deadline exceeded"

**Cause:** Operation took too long

**Solution:**
```go
// Increase timeout
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
// Verify model name
agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4"). // Check spelling
    Build()
```

### "rate limit exceeded"

**Cause:** Too many requests

**Solution:** Implement retry with backoff (see LLM Provider Issues)

### "connection refused"

**Cause:** Service not running or wrong address

**Solution:**
```bash
# Check if service is running
netstat -an | grep 8080

# Verify address
curl http://localhost:8080
```

---

## 🔍 Debugging Tips

### Enable Verbose Logging

```go
config := &v1beta.Config{
    DebugMode: true,
    LLM: v1beta.LLMConfig{
        Provider: "openai",
        Model:    "gpt-4",
    },
}
```

### Inspect Agent State

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

agent, _ := v1beta.NewBuilder("agent").
    WithLLM("openai", "gpt-4").
    Build()

// Check configuration
config := agent.Config()
fmt.Printf("Config: %+v\n", config)

// Check capabilities
caps := agent.Capabilities()
fmt.Printf("Capabilities: %v\n", caps)
```

### Test Components Individually

```go
// Test LLM connection
agent, err := v1beta.NewBuilder("test").
    WithLLM("openai", "gpt-4").
    Build()

result, err := agent.Run(context.Background(), "Hello")
if err != nil {
    fmt.Printf("LLM test failed: %v\n", err)
}

// Test tools separately
// Test memory separately
```

### Check Error Details

```go
import "github.com/agenticgokit/agenticgokit/v1beta"

result, err := agent.Run(ctx, input)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    fmt.Printf("Code: %v\n", v1beta.GetErrorCode(err))
    fmt.Printf("Suggestion: %v\n", v1beta.GetErrorSuggestion(err))
    
    details := v1beta.GetErrorDetails(err)
    fmt.Printf("Details: %+v\n", details)
}
```

---

## 📚 Getting Help

### Check Documentation

- [Getting Started](./getting-started.md) - Basic setup
- [Error Handling](./error-handling.md) - Error patterns
- [Performance](./performance.md) - Optimization tips
- [Configuration](./configuration.md) - All settings

### Enable Debug Mode

```toml
debug_mode = true

[llm]
provider = "openai"
model = "gpt-4"
debug = true

[tools]
enabled = true
debug = true

[memory]
provider = "memory"
debug = true
```

### Report Issues

When reporting issues, include:
- Error message and code
- Minimal reproduction code
- Configuration (redact sensitive info)
- Go version and OS
- AgenticGoKit version

---

**Still stuck?** Check [GitHub Issues](https://github.com/agenticgokit/agenticgokit/issues) or join our community chat.
