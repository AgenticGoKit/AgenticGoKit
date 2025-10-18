# Migration Guide: vNext API Updates

This guide helps you migrate from older AgenticGoKit APIs to the consolidated vNext APIs.

## Table of Contents

- [Overview](#overview)
- [Agent Interface Changes](#agent-interface-changes)
- [RunOptions Consolidation](#runoptions-consolidation)
- [Streaming API Changes](#streaming-api-changes)
- [Configuration Changes](#configuration-changes)
- [Builder Pattern Changes](#builder-pattern-changes)
- [Workflow Changes](#workflow-changes)
- [Error Handling Changes](#error-handling-changes)
- [Quick Migration Checklist](#quick-migration-checklist)

## Overview

The vNext API consolidation focused on:

1. **Simplified interfaces**: Reduced methods from 30+ to 8 core methods
2. **Unified RunOptions**: One options struct for all execution modes
3. **Enhanced streaming**: New Stream interface with better control
4. **Cleaner configuration**: Consolidated config structs with functional options
5. **Better error handling**: Structured errors with helpful messages

## Agent Interface Changes

### Old API

```go
// Multiple run methods
result, err := agent.Run(ctx, input)
result, err := agent.RunWithOptions(ctx, input, options)
result, err := agent.RunWithStreaming(ctx, input, streamHandler)
result, err := agent.RunWithConfig(ctx, input, config)
```

### New API

```go
// Single Run method with options
result, err := agent.Run(ctx, input, opts...)

// Or use RunOptions explicitly
opts := &vnext.RunOptions{
    MaxTokens:   1000,
    Temperature: 0.7,
}
result, err := agent.RunWithOptions(ctx, input, opts)
```

### Migration Steps

1. Replace all `RunWithOptions` calls with `Run` + functional options
2. Replace `RunWithStreaming` with `RunStream` (see Streaming section)
3. Replace `RunWithConfig` with `Run` + config-based options

**Before:**
```go
result, err := agent.RunWithOptions(ctx, input, &Options{
    MaxTokens: 1000,
    Temperature: 0.7,
})
```

**After:**
```go
result, err := agent.Run(ctx, input,
    vnext.WithMaxTokens(1000),
    vnext.WithTemperature(0.7),
)
```

## RunOptions Consolidation

### Old API

```go
// Multiple separate option structs
type ExecutionOptions struct {
    MaxTokens int
    Temperature float32
}

type StreamingOptions struct {
    BufferSize int
    Handler StreamHandler
}

type MemoryOptions struct {
    SessionID string
    RAGEnabled bool
}
```

### New API

```go
// Unified RunOptions with functional options
type RunOptions struct {
    // Execution options
    MaxTokens   int
    Temperature float32
    
    // Streaming options
    StreamOptions *StreamOptions
    StreamHandler StreamHandler
    
    // Memory options
    SessionID  string
    RAGEnabled bool
    
    // Tool options
    ToolTimeout time.Duration
    MaxToolRetries int
}

// Functional options pattern
result, err := agent.Run(ctx, input,
    vnext.WithMaxTokens(1000),
    vnext.WithTemperature(0.7),
    vnext.WithStream(vnext.WithBufferSize(100)),
    vnext.WithSessionID("session-123"),
)
```

### Migration Steps

1. Combine multiple option structs into single `RunOptions`
2. Use functional options for cleaner API
3. Remove deprecated option fields

**Before:**
```go
execOpts := &ExecutionOptions{MaxTokens: 1000}
streamOpts := &StreamingOptions{BufferSize: 100}
memoryOpts := &MemoryOptions{SessionID: "s1"}

// Multiple calls with different options
result1, _ := agent.RunWithExecOptions(ctx, input, execOpts)
result2, _ := agent.RunWithStreamOptions(ctx, input, streamOpts)
result3, _ := agent.RunWithMemoryOptions(ctx, input, memoryOpts)
```

**After:**
```go
result, err := agent.Run(ctx, input,
    vnext.WithMaxTokens(1000),
    vnext.WithSessionID("s1"),
)
```

## Streaming API Changes

### Old API

```go
// Old boolean-based streaming
result, err := agent.Run(ctx, input, &Options{
    Streaming: true,
})

// Or with handler
err := agent.RunWithStreaming(ctx, input, func(text string) {
    fmt.Print(text)
})
```

### New API

```go
// New Stream interface
stream, err := agent.RunStream(ctx, input, opts...)
if err != nil {
    log.Fatal(err)
}

// Process chunks
for chunk := range stream.Chunks() {
    switch chunk.Type {
    case vnext.ChunkTypeDelta:
        fmt.Print(chunk.Delta)
    case vnext.ChunkTypeThought:
        log.Printf("Thinking: %s", chunk.Content)
    }
}

// Get final result
result, err := stream.Wait()
```

### Streaming with Callback Handler

**Before:**
```go
err := agent.RunWithStreaming(ctx, input, func(text string) {
    fmt.Print(text)
})
```

**After:**
```go
handler := func(chunk *vnext.StreamChunk) error {
    if chunk.Type == vnext.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
    return nil
}

result, err := agent.Run(ctx, input, vnext.WithStreamHandler(handler))
```

### Migration Steps

1. Replace `Streaming: true` with `RunStream()` method
2. Replace simple string handlers with `StreamChunk` handlers
3. Use chunk types for better control
4. Always call `stream.Wait()` to get final result

**Complete Before/After Example:**

**Before:**
```go
func processWithStreaming(agent Agent, query string) error {
    var fullResponse string
    
    err := agent.RunWithStreaming(context.Background(), query, 
        func(text string) {
            fmt.Print(text)
            fullResponse += text
        })
    
    if err != nil {
        return err
    }
    
    log.Printf("Complete response: %s", fullResponse)
    return nil
}
```

**After:**
```go
func processWithStreaming(agent vnext.Agent, query string) error {
    stream, err := agent.RunStream(context.Background(), query,
        vnext.WithTextOnly(true),
    )
    if err != nil {
        return err
    }
    
    for chunk := range stream.Chunks() {
        fmt.Print(chunk.Delta)
    }
    
    result, err := stream.Wait()
    if err != nil {
        return err
    }
    
    log.Printf("Complete response: %s", result.Content)
    return nil
}
```

## Configuration Changes

### Old API

```go
// Multiple config files
config, err := LoadConfig("config.toml")
projectConfig, err := LoadProjectConfig("project.toml")
validationConfig, err := LoadValidationConfig("validation.toml")

// Separate config structs
llmConfig := &LLMConfig{...}
memoryConfig := &MemoryConfig{...}
toolsConfig := &ToolsConfig{...}
```

### New API

```go
// Single unified config
config, err := vnext.LoadConfig("config.toml")

// All configurations in one struct
type Config struct {
    Name         string
    SystemPrompt string
    LLM          LLMConfig
    Memory       *MemoryConfig
    Tools        *ToolsConfig
    Workflow     *WorkflowConfig
    Tracing      *TracingConfig
    Streaming    *StreamingConfig
}
```

### TOML Configuration

**Before:**
```toml
# llm.toml
provider = "openai"
model = "gpt-4"

# memory.toml
provider = "memory"
connection = "inmemory"

# tools.toml
enabled = true
max_retries = 3
```

**After:**
```toml
# config.toml - Everything in one file
name = "MyAgent"
system_prompt = "You are a helpful assistant"

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7

[memory]
provider = "memory"
connection = "inmemory"

[tools]
enabled = true
max_retries = 3

[streaming]
enabled = true
buffer_size = 100
include_thoughts = true
```

### Migration Steps

1. Merge multiple config files into single `config.toml`
2. Update struct references to use unified `Config`
3. Use nested config sections in TOML

## Builder Pattern Changes

### Old API

```go
// Many builder methods
builder := NewAgentBuilder().
    SetName("Agent").
    SetSystemPrompt("Prompt").
    SetLLMProvider("openai").
    SetLLMModel("gpt-4").
    SetLLMTemperature(0.7).
    SetLLMMaxTokens(1000).
    SetMemoryProvider("memory").
    SetMemoryConnection("inmemory").
    EnableTools(true).
    SetToolMaxRetries(3).
    SetToolTimeout(30 * time.Second).
    EnableStreaming(true).
    SetStreamingBufferSize(100).
    EnableTracing(true).
    SetTracingLevel("debug")
    
agent, err := builder.Build()
```

### New API

```go
// Simplified builder with preset
agent, err := vnext.PresetChatAgentBuilder().
    WithName("Agent").
    WithSystemPrompt("Prompt").
    WithLLM("openai", "gpt-4").
    WithMemory("memory", "inmemory").
    WithTools(true).
    Build()

// Or use functional options
agent, err := vnext.PresetChatAgentBuilder().
    WithName("Agent").
    WithOptions(
        vnext.WithTemperature(0.7),
        vnext.WithMaxTokens(1000),
    ).
    Build()
```

### Preset Builders

New preset builders for common scenarios:

```go
// Chat agent
vnext.PresetChatAgentBuilder()

// Research agent with memory and tools
vnext.PresetResearchAgentBuilder()

// Data processing agent
vnext.PresetDataAgentBuilder()

// Workflow orchestration agent
vnext.PresetWorkflowAgentBuilder()
```

### Migration Steps

1. Replace `NewAgentBuilder()` with appropriate preset
2. Consolidate setter methods into `WithOptions()`
3. Use preset defaults instead of setting everything

**Before:**
```go
agent, err := NewAgentBuilder().
    SetName("ChatBot").
    SetSystemPrompt("You are a helpful assistant").
    SetLLMProvider("openai").
    SetLLMModel("gpt-4").
    SetLLMTemperature(0.7).
    SetLLMMaxTokens(1000).
    EnableStreaming(true).
    Build()
```

**After:**
```go
agent, err := vnext.PresetChatAgentBuilder().
    WithName("ChatBot").
    WithSystemPrompt("You are a helpful assistant").
    WithLLM("openai", "gpt-4").
    Build()
// Temperature, MaxTokens, and Streaming use preset defaults
```

## Workflow Changes

### Old API

```go
// Complex workflow creation
workflow := NewWorkflow("Pipeline")
workflow.AddStep("step1", agent1, "Do task 1")
workflow.AddStep("step2", agent2, "Do task 2")
workflow.SetMode(Sequential)
workflow.SetErrorHandling(ContinueOnError)

result, err := workflow.Execute(ctx, input)
```

### New API

```go
// Simplified workflow creation
workflow, err := vnext.NewSequentialWorkflow("Pipeline",
    vnext.Step("step1", agent1, "Do task 1"),
    vnext.Step("step2", agent2, "Do task 2"),
)

// Run workflow
result, err := workflow.Run(ctx, input)

// Or with streaming
stream, err := workflow.RunStream(ctx, input)
```

### Workflow Modes

**Before:**
```go
workflow := NewWorkflow("Pipeline")
workflow.SetMode(Sequential)  // or Parallel, DAG, Loop

result, err := workflow.Execute(ctx, input)
```

**After:**
```go
// Mode is explicit in constructor
sequential, err := vnext.NewSequentialWorkflow(name, steps...)
parallel, err := vnext.NewParallelWorkflow(name, steps...)
dag, err := vnext.NewDAGWorkflow(name, steps...)
loop, err := vnext.NewLoopWorkflow(name, steps...)

result, err := workflow.Run(ctx, input)
```

### Migration Steps

1. Replace `NewWorkflow()` with mode-specific constructor
2. Replace `AddStep()` with `Step()` helpers
3. Replace `Execute()` with `Run()` or `RunStream()`

## Error Handling Changes

### Old API

```go
// Generic errors
result, err := agent.Run(ctx, input)
if err != nil {
    log.Printf("Error: %v", err)
    return err
}
```

### New API

```go
// Structured errors with codes
result, err := agent.Run(ctx, input)
if err != nil {
    if agentErr, ok := err.(*vnext.Error); ok {
        switch agentErr.Code {
        case vnext.ErrCodeLLM:
            log.Printf("LLM error: %v", agentErr)
            // Retry with different LLM
        case vnext.ErrCodeTimeout:
            log.Printf("Timeout: %v", agentErr)
            // Increase timeout
        case vnext.ErrCodeTool:
            log.Printf("Tool error: %v", agentErr)
            // Disable tool and retry
        default:
            return agentErr
        }
    }
}
```

### Error Codes

Available error codes:

- `ErrCodeLLM`: LLM provider errors
- `ErrCodeTool`: Tool execution errors
- `ErrCodeMemory`: Memory/RAG errors
- `ErrCodeConfig`: Configuration errors
- `ErrCodeTimeout`: Timeout errors
- `ErrCodeValidation`: Input validation errors
- `ErrCodeStream`: Streaming errors

### Migration Steps

1. Add error type checking for `*vnext.Error`
2. Handle specific error codes
3. Use error context for debugging

**Before:**
```go
result, err := agent.Run(ctx, input)
if err != nil {
    if strings.Contains(err.Error(), "timeout") {
        // Handle timeout
    } else if strings.Contains(err.Error(), "llm") {
        // Handle LLM error
    }
}
```

**After:**
```go
result, err := agent.Run(ctx, input)
if err != nil {
    if agentErr, ok := err.(*vnext.Error); ok {
        switch agentErr.Code {
        case vnext.ErrCodeTimeout:
            // Handle timeout
        case vnext.ErrCodeLLM:
            // Handle LLM error
        }
        
        // Get suggestions
        log.Printf("Error: %s\nSuggestion: %s", 
            agentErr.Message, agentErr.Suggestion)
    }
}
```

## Quick Migration Checklist

### Imports

- [ ] Update import: `vnext "github.com/kunalkushwaha/agenticgokit/core/vnext"`

### Agent Creation

- [ ] Replace `NewAgentBuilder()` with `PresetXXXAgentBuilder()`
- [ ] Consolidate setter methods into fewer builder calls
- [ ] Use `WithOptions()` for execution options

### Agent Execution

- [ ] Replace `RunWithOptions()` with `Run()` + functional options
- [ ] Replace `RunWithStreaming()` with `RunStream()`
- [ ] Replace `RunWithConfig()` with `Run()` + config options

### Streaming

- [ ] Replace `Streaming: true` with `RunStream()` method
- [ ] Update handlers to accept `*StreamChunk` instead of `string`
- [ ] Add `stream.Wait()` calls to get final results
- [ ] Use chunk types for filtering

### Configuration

- [ ] Merge multiple config files into single `config.toml`
- [ ] Update config struct references to `vnext.Config`
- [ ] Use nested config sections in TOML

### Workflows

- [ ] Replace `NewWorkflow()` with mode-specific constructors
- [ ] Replace `AddStep()` with `Step()` helpers
- [ ] Replace `Execute()` with `Run()` or `RunStream()`

### Error Handling

- [ ] Add type checking for `*vnext.Error`
- [ ] Handle specific error codes
- [ ] Use error suggestions for user feedback

### Testing

- [ ] Update test cases to use new APIs
- [ ] Test streaming functionality
- [ ] Verify error handling

## Common Migration Patterns

### Pattern 1: Simple Agent

**Before:**
```go
builder := NewAgentBuilder().
    SetName("Assistant").
    SetLLMProvider("openai").
    SetLLMModel("gpt-4")
    
agent, err := builder.Build()
result, err := agent.Run(ctx, "Hello")
```

**After:**
```go
agent, err := vnext.PresetChatAgentBuilder().
    WithName("Assistant").
    Build()
    
result, err := agent.Run(ctx, "Hello")
```

### Pattern 2: Agent with Streaming

**Before:**
```go
agent, _ := NewAgentBuilder().Build()
agent.RunWithStreaming(ctx, "Hello", func(text string) {
    fmt.Print(text)
})
```

**After:**
```go
agent, _ := vnext.PresetChatAgentBuilder().Build()
stream, _ := agent.RunStream(ctx, "Hello")
for chunk := range stream.Chunks() {
    if chunk.Type == vnext.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
}
stream.Wait()
```

### Pattern 3: Agent with Memory

**Before:**
```go
agent, _ := NewAgentBuilder().
    SetMemoryProvider("memory").
    SetMemoryConnection("inmemory").
    EnableRAG(true).
    Build()
```

**After:**
```go
agent, _ := vnext.PresetResearchAgentBuilder().
    WithMemory("memory", "inmemory").
    Build()
```

### Pattern 4: Multi-Step Workflow

**Before:**
```go
workflow := NewWorkflow("Pipeline")
workflow.AddStep("s1", agent1, "Step 1")
workflow.AddStep("s2", agent2, "Step 2")
workflow.SetMode(Sequential)
result, _ := workflow.Execute(ctx, input)
```

**After:**
```go
workflow, _ := vnext.NewSequentialWorkflow("Pipeline",
    vnext.Step("s1", agent1, "Step 1"),
    vnext.Step("s2", agent2, "Step 2"),
)
result, _ := workflow.Run(ctx, input)
```

## Getting Help

- **Documentation**: See [README.md](README.md) and [STREAMING_GUIDE.md](STREAMING_GUIDE.md)
- **Examples**: Check `examples/` directory for complete examples
- **Issues**: Report issues at https://github.com/kunalkushwaha/agenticgokit/issues

## Version Compatibility

- **vNext**: Consolidated APIs (recommended)
- **Legacy**: Old APIs still available but deprecated
- **Migration period**: Both APIs supported for 2 releases

### Deprecation Notices

The following will be removed in future versions:

- `RunWithOptions()` - Use `Run()` with functional options
- `RunWithStreaming()` - Use `RunStream()`
- `Streaming` bool field - Use `StreamOptions`
- Multiple config files - Use single unified config
- Long builder chains - Use presets and functional options

## Summary

The vNext API consolidation provides:

✅ **Fewer methods**: 30+ reduced to 8 core methods  
✅ **Cleaner streaming**: Dedicated Stream interface  
✅ **Better options**: Functional options pattern  
✅ **Unified config**: Single config file  
✅ **Preset builders**: Quick agent creation  
✅ **Better errors**: Structured errors with suggestions  

Start your migration today for a cleaner, more maintainable codebase!
