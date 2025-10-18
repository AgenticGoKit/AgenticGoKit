# Troubleshooting Guide

This guide helps you diagnose and fix common issues when working with AgenticGoKit vNext.

## Table of Contents

- [Configuration Issues](#configuration-issues)
- [Streaming Problems](#streaming-problems)
- [Memory and Performance](#memory-and-performance)
- [LLM Provider Issues](#llm-provider-issues)
- [Tool Execution Errors](#tool-execution-errors)
- [Workflow Issues](#workflow-issues)
- [Build and Compilation](#build-and-compilation)
- [Common Error Messages](#common-error-messages)

## Configuration Issues

### Problem: Config file not found

**Error:**
```
error loading config: open config.toml: no such file or directory
```

**Solutions:**

1. **Check file path:**
   ```go
   // Use absolute path
   config, err := vnext.LoadConfig("/path/to/config.toml")
   
   // Or relative to working directory
   config, err := vnext.LoadConfig("./config.toml")
   ```

2. **Verify working directory:**
   ```go
   wd, _ := os.Getwd()
   fmt.Printf("Working directory: %s\n", wd)
   ```

3. **Use embedded config:**
   ```go
   config := &vnext.Config{
       Name: "MyAgent",
       LLM: vnext.LLMConfig{
           Provider: "openai",
           Model:    "gpt-4",
       },
   }
   agent, err := vnext.NewAgentFromConfig(config)
   ```

### Problem: Invalid configuration values

**Error:**
```
validation error: temperature must be between 0.0 and 2.0
```

**Solutions:**

Check your configuration values:

```toml
[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7  # Must be 0.0-2.0
max_tokens = 1000  # Must be positive

[streaming]
buffer_size = 100  # Must be positive
timeout_ms = 30000 # Must be positive
```

Valid ranges:
- **Temperature**: 0.0 - 2.0
- **MaxTokens**: 1 - model limit
- **BufferSize**: 1 - 10000 (recommended: 50-500)
- **Timeout**: > 0 (milliseconds)

### Problem: Missing API keys

**Error:**
```
LLM error: API key not found
```

**Solutions:**

1. **Use environment variables:**
   ```bash
   export OPENAI_API_KEY="sk-..."
   export ANTHROPIC_API_KEY="..."
   ```

2. **Set in code:**
   ```go
   agent, err := vnext.PresetChatAgentBuilder().
       WithLLM("openai", "gpt-4").
       WithAPIKey(os.Getenv("OPENAI_API_KEY")).
       Build()
   ```

3. **Use config file (not recommended for production):**
   ```toml
   [llm]
   provider = "openai"
   model = "gpt-4"
   api_key = "sk-..."  # Better: use env vars
   ```

## Streaming Problems

### Problem: Stream hangs and never completes

**Symptoms:**
- `stream.Wait()` never returns
- No chunks received
- Program appears frozen

**Solutions:**

1. **Always use context with timeout:**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   
   stream, err := agent.RunStream(ctx, query)
   ```

2. **Consume all chunks:**
   ```go
   for chunk := range stream.Chunks() {
       // Process chunk
   }
   // Channel must be fully consumed
   result, err := stream.Wait()
   ```

3. **Check for errors in chunks:**
   ```go
   for chunk := range stream.Chunks() {
       if chunk.Type == vnext.ChunkTypeError {
           log.Printf("Stream error: %v", chunk.Error)
           break
       }
   }
   ```

### Problem: Memory leak with streaming

**Symptoms:**
- Memory usage grows over time
- Goroutine count increases
- Application becomes slow

**Solutions:**

1. **Always cancel unused streams:**
   ```go
   stream, err := agent.RunStream(ctx, query)
   if err != nil {
       return err
   }
   
   // Cancel if not consuming
   if !needsStreaming {
       defer stream.Cancel()
       return nil
   }
   ```

2. **Fully consume or cancel:**
   ```go
   stream, err := agent.RunStream(ctx, query)
   defer stream.Cancel() // Safety net
   
   for chunk := range stream.Chunks() {
       processChunk(chunk)
   }
   ```

3. **Use context cancellation:**
   ```go
   ctx, cancel := context.WithCancel(context.Background())
   defer cancel() // Cleanup
   
   stream, err := agent.RunStream(ctx, query)
   ```

### Problem: Slow streaming performance

**Symptoms:**
- Chunks arrive slowly
- High latency between chunks
- UI feels sluggish

**Solutions:**

1. **Increase buffer size:**
   ```go
   stream, err := agent.RunStream(ctx, query,
       vnext.WithBufferSize(500), // Larger buffer
   )
   ```

2. **Adjust flush interval:**
   ```go
   stream, err := agent.RunStream(ctx, query,
       vnext.WithFlushInterval(50*time.Millisecond), // Faster flushing
   )
   ```

3. **Use text-only mode:**
   ```go
   stream, err := agent.RunStream(ctx, query,
       vnext.WithTextOnly(true), // Skip thoughts, tools, metadata
   )
   ```

### Problem: Missing chunks or incomplete data

**Symptoms:**
- Some chunks don't arrive
- Incomplete text output
- Missing tool calls

**Solutions:**

1. **Check buffer overflow:**
   ```go
   // Increase buffer size
   stream, err := agent.RunStream(ctx, query,
       vnext.WithBufferSize(1000),
   )
   ```

2. **Process chunks quickly:**
   ```go
   for chunk := range stream.Chunks() {
       // Don't block here - send to channel or process async
       go processChunkAsync(chunk)
   }
   ```

3. **Check for errors:**
   ```go
   for chunk := range stream.Chunks() {
       if chunk.Type == vnext.ChunkTypeError {
           log.Printf("Error: %v", chunk.Error)
       }
   }
   
   result, err := stream.Wait()
   if err != nil {
       log.Printf("Stream ended with error: %v", err)
   }
   ```

## Memory and Performance

### Problem: High memory usage

**Solutions:**

1. **Limit context size:**
   ```go
   result, err := agent.Run(ctx, input,
       vnext.WithMaxTokens(1000), // Limit output
       vnext.WithHistoryLimit(10), // Limit conversation history
   )
   ```

2. **Clear memory periodically:**
   ```go
   // Clear old sessions
   memory.Clear(sessionID)
   
   // Or use short-lived agents
   agent, _ := vnext.PresetChatAgentBuilder().Build()
   result, _ := agent.Run(ctx, input)
   // Agent GC'd after use
   ```

3. **Use streaming to reduce buffering:**
   ```go
   // Instead of buffering full response
   stream, err := agent.RunStream(ctx, query)
   for chunk := range stream.Chunks() {
       // Process and discard immediately
       sendToClient(chunk.Delta)
   }
   ```

### Problem: Slow performance

**Solutions:**

1. **Enable caching:**
   ```toml
   [tools.cache]
   enabled = true
   ttl = "5m"
   max_size = 1000
   ```

2. **Use parallel workflows:**
   ```go
   // Instead of sequential
   workflow, _ := vnext.NewParallelWorkflow("Pipeline",
       vnext.Step("s1", agent1, "Task 1"),
       vnext.Step("s2", agent2, "Task 2"),
   )
   ```

3. **Optimize LLM calls:**
   ```go
   result, err := agent.Run(ctx, input,
       vnext.WithTemperature(0.0), // Faster, deterministic
       vnext.WithMaxTokens(500),   // Shorter responses
   )
   ```

### Problem: Goroutine leaks

**Symptoms:**
- Goroutine count grows
- Application slows over time
- Cannot create new goroutines

**Solutions:**

1. **Use context cancellation:**
   ```go
   ctx, cancel := context.WithCancel(context.Background())
   defer cancel()
   
   result, err := agent.Run(ctx, input)
   ```

2. **Cancel streams:**
   ```go
   stream, err := agent.RunStream(ctx, query)
   defer stream.Cancel()
   ```

3. **Wait for workflows:**
   ```go
   workflow, _ := vnext.NewParallelWorkflow(...)
   result, err := workflow.Run(ctx, input)
   // Workflow waits for all agents to complete
   ```

## LLM Provider Issues

### Problem: OpenAI API errors

**Error:**
```
LLM error: status 429: rate limit exceeded
```

**Solutions:**

1. **Add retry logic:**
   ```go
   result, err := agent.Run(ctx, input,
       vnext.WithMaxRetries(3),
       vnext.WithRetryDelay(time.Second),
   )
   ```

2. **Implement backoff:**
   ```go
   for i := 0; i < 3; i++ {
       result, err := agent.Run(ctx, input)
       if err == nil {
           break
       }
       if strings.Contains(err.Error(), "429") {
           time.Sleep(time.Second * time.Duration(i+1))
           continue
       }
       return err
   }
   ```

3. **Use rate limiting:**
   ```toml
   [tools]
   rate_limit = 10  # requests per second
   max_concurrent = 5
   ```

### Problem: Model not found

**Error:**
```
LLM error: model 'gpt-5' not found
```

**Solutions:**

1. **Check model name:**
   ```go
   // OpenAI
   model := "gpt-4"        // or "gpt-3.5-turbo"
   
   // Anthropic
   model := "claude-3-opus-20240229"
   
   // Ollama
   model := "llama2"
   ```

2. **List available models:**
   ```bash
   # OpenAI
   curl https://api.openai.com/v1/models \
     -H "Authorization: Bearer $OPENAI_API_KEY"
   
   # Ollama
   ollama list
   ```

### Problem: Context length exceeded

**Error:**
```
LLM error: maximum context length exceeded
```

**Solutions:**

1. **Reduce input size:**
   ```go
   result, err := agent.Run(ctx, input,
       vnext.WithMaxTokens(500),    // Limit output
       vnext.WithHistoryLimit(5),   // Limit history
   )
   ```

2. **Use summarization:**
   ```go
   if len(input) > 4000 {
       summary, _ := summarizeAgent.Run(ctx, input)
       input = summary.Content
   }
   result, err := agent.Run(ctx, input)
   ```

3. **Switch to larger context model:**
   ```go
   agent, _ := vnext.PresetChatAgentBuilder().
       WithLLM("openai", "gpt-4-32k"). // Larger context
       Build()
   ```

## Tool Execution Errors

### Problem: Tool not found

**Error:**
```
tool error: tool 'calculator' not found
```

**Solutions:**

1. **Register tool:**
   ```go
   tools := []vnext.Tool{
       {
           Name:        "calculator",
           Description: "Performs calculations",
           Handler:     calculatorHandler,
       },
   }
   
   agent, _ := vnext.PresetChatAgentBuilder().
       WithTools(tools).
       Build()
   ```

2. **Check tool name:**
   ```go
   // Names must match exactly
   tool := vnext.Tool{
       Name: "calculator", // Not "calc" or "Calculator"
   }
   ```

### Problem: Tool timeout

**Error:**
```
tool error: execution timeout after 30s
```

**Solutions:**

1. **Increase timeout:**
   ```go
   result, err := agent.Run(ctx, input,
       vnext.WithToolTimeout(60*time.Second),
   )
   ```

2. **Configure in TOML:**
   ```toml
   [tools]
   timeout = "60s"
   max_retries = 3
   ```

3. **Optimize tool handler:**
   ```go
   func toolHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
       // Use context timeout
       ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
       defer cancel()
       
       // Make operation cancellable
       select {
       case result := <-doWork(ctx, args):
           return result, nil
       case <-ctx.Done():
           return nil, ctx.Err()
       }
   }
   ```

### Problem: Tool execution fails

**Error:**
```
tool error: invalid arguments
```

**Solutions:**

1. **Validate arguments:**
   ```go
   func toolHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
       x, ok := args["x"].(float64)
       if !ok {
           return nil, fmt.Errorf("x must be a number")
       }
       
       y, ok := args["y"].(float64)
       if !ok {
           return nil, fmt.Errorf("y must be a number")
       }
       
       return x + y, nil
   }
   ```

2. **Provide clear schema:**
   ```go
   tool := vnext.Tool{
       Name:        "calculator",
       Description: "Adds two numbers. Args: x (number), y (number)",
       Schema: map[string]interface{}{
           "type": "object",
           "properties": map[string]interface{}{
               "x": map[string]string{"type": "number"},
               "y": map[string]string{"type": "number"},
           },
           "required": []string{"x", "y"},
       },
       Handler: calculatorHandler,
   }
   ```

## Workflow Issues

### Problem: Workflow steps execute out of order

**Symptoms:**
- Sequential workflow runs in wrong order
- DAG dependencies not respected

**Solutions:**

1. **Use correct workflow type:**
   ```go
   // For strict ordering
   workflow, _ := vnext.NewSequentialWorkflow("Pipeline",
       vnext.Step("step1", agent1, "First"),
       vnext.Step("step2", agent2, "Second"),
   )
   
   // For parallel execution
   workflow, _ := vnext.NewParallelWorkflow("Pipeline",
       vnext.Step("task1", agent1, "Independent 1"),
       vnext.Step("task2", agent2, "Independent 2"),
   )
   ```

2. **Define DAG dependencies:**
   ```go
   workflow, _ := vnext.NewDAGWorkflow("Pipeline",
       vnext.Step("step1", agent1, "First"),
       vnext.Step("step2", agent2, "Second", "step1"), // Depends on step1
       vnext.Step("step3", agent3, "Third", "step1", "step2"),
   )
   ```

### Problem: Workflow doesn't pass data between steps

**Solutions:**

1. **Use workflow context:**
   ```go
   agent1 := vnext.PresetChatAgentBuilder().
       WithCustomHandler(func(ctx context.Context, input string, llmCall func(string, string) (string, error)) (string, error) {
           result := "processed: " + input
           // Store in context for next step
           return result, nil
       }).
       Build()
   ```

2. **Access previous step results:**
   ```go
   workflow, _ := vnext.NewSequentialWorkflow("Pipeline",
       vnext.Step("extract", extractAgent, "Extract data"),
       vnext.Step("transform", transformAgent, "Transform using previous result"),
   )
   ```

### Problem: Workflow fails on error

**Solutions:**

1. **Add error handling:**
   ```go
   workflow, _ := vnext.NewSequentialWorkflow("Pipeline",
       vnext.Step("step1", agent1, "May fail"),
       vnext.Step("step2", agent2, "Continue anyway"),
   )
   
   result, err := workflow.Run(ctx, input)
   if err != nil {
       // Check which step failed
       log.Printf("Workflow error: %v", err)
   }
   ```

2. **Use try-catch pattern:**
   ```go
   result, err := workflow.Run(ctx, input)
   if err != nil {
       // Fallback workflow
       fallbackWorkflow, _ := vnext.NewSequentialWorkflow("Fallback", ...)
       result, err = fallbackWorkflow.Run(ctx, input)
   }
   ```

## Build and Compilation

### Problem: Import errors

**Error:**
```
package github.com/kunalkushwaha/agenticgokit/core/vnext: no matching versions
```

**Solutions:**

1. **Update dependencies:**
   ```bash
   go get github.com/kunalkushwaha/agenticgokit/core/vnext@latest
   go mod tidy
   ```

2. **Check import path:**
   ```go
   import (
       vnext "github.com/kunalkushwaha/agenticgokit/core/vnext"
   )
   ```

### Problem: Type errors

**Error:**
```
cannot use agent (type *Agent) as type vnext.Agent
```

**Solutions:**

1. **Update to new interfaces:**
   ```go
   // Old
   var agent *Agent
   
   // New
   var agent vnext.Agent
   ```

2. **Check type assertions:**
   ```go
   if streamAgent, ok := agent.(vnext.StreamingAgent); ok {
       stream, _ := streamAgent.RunStream(ctx, input)
   }
   ```

## Common Error Messages

### "context deadline exceeded"

**Cause:** Operation took longer than timeout

**Solution:**
```go
// Increase timeout
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

result, err := agent.Run(ctx, input)
```

### "rate limit exceeded"

**Cause:** Too many API requests

**Solution:**
```go
// Add rate limiting
result, err := agent.Run(ctx, input,
    vnext.WithMaxRetries(3),
    vnext.WithRetryDelay(2*time.Second),
)
```

### "invalid configuration"

**Cause:** Config validation failed

**Solution:**
```go
// Check config values
config := &vnext.Config{
    Name: "Agent", // Required
    LLM: vnext.LLMConfig{
        Provider: "openai", // Required
        Model:    "gpt-4",  // Required
    },
}
```

### "stream already closed"

**Cause:** Attempting to use closed stream

**Solution:**
```go
stream, err := agent.RunStream(ctx, query)
if err != nil {
    return err
}

// Consume stream only once
for chunk := range stream.Chunks() {
    processChunk(chunk)
}

// Don't reuse stream
result, err := stream.Wait() // OK
result, err = stream.Wait()  // ERROR: stream closed
```

### "nil pointer dereference"

**Cause:** Using nil agent or config

**Solution:**
```go
// Always check build errors
agent, err := vnext.PresetChatAgentBuilder().Build()
if err != nil {
    log.Fatal(err) // Don't ignore
}

// Check before use
if agent == nil {
    log.Fatal("agent is nil")
}

result, err := agent.Run(ctx, input)
```

## Debugging Tips

### Enable Debug Logging

```go
agent, _ := vnext.PresetChatAgentBuilder().
    WithDebugMode(true).
    Build()
```

Or in config:
```toml
debug_mode = true

[tracing]
enabled = true
level = "debug"
```

### Check Agent State

```go
// Log configuration
log.Printf("Agent config: %+v", agent.Config())

// Check streaming support
if _, ok := agent.(vnext.StreamingAgent); ok {
    log.Println("Streaming supported")
}
```

### Monitor Performance

```go
start := time.Now()
result, err := agent.Run(ctx, input)
duration := time.Since(start)

log.Printf("Execution took: %v", duration)
log.Printf("Token count: %d", result.TokenCount)
```

### Inspect Chunks

```go
stream, _ := agent.RunStream(ctx, input)
for chunk := range stream.Chunks() {
    log.Printf("Chunk: type=%s, timestamp=%v, index=%d",
        chunk.Type, chunk.Timestamp, chunk.Index)
}
```

## Getting More Help

1. **Check documentation:**
   - [README.md](README.md)
   - [STREAMING_GUIDE.md](STREAMING_GUIDE.md)
   - [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)

2. **Search examples:**
   - `examples/` directory
   - `test/` directory

3. **Report issues:**
   - GitHub: https://github.com/kunalkushwaha/agenticgokit/issues
   - Include: error message, minimal reproduction code, Go version

4. **Community:**
   - Discussions: https://github.com/kunalkushwaha/agenticgokit/discussions
   - Discord: [Link if available]

## Quick Fixes Checklist

When something goes wrong, check these first:

- [ ] Using context with timeout
- [ ] Consuming or canceling streams
- [ ] API keys set in environment
- [ ] Config file exists and is valid
- [ ] Model names are correct
- [ ] Tools are registered
- [ ] Error checking after every operation
- [ ] Dependencies are up to date (`go mod tidy`)
- [ ] Using correct import path
- [ ] Not reusing closed streams
