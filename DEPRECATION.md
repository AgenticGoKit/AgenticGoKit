# Deprecation Notice

## Overview

The `core` and `core/vnext` packages are **deprecated** and will be removed in **v1.0.0**.

**Migration Timeline:**
- **Now (v0.x)**: `core` and `core/vnext` are deprecated with notices
- **v1.0.0**: Only `v1beta` APIs will remain (becoming the primary `v1` package)

## Why the Change?

We're consolidating our APIs for better stability and maintainability:

- **Before**: Multiple API surfaces (`core`, `core/vnext`, `v1beta`)
- **After v1.0**: Single, stable `v1` API (currently `v1beta`)

## Migration Guide

### Quick Migration

**Old Code (core/vnext):**
```go
import "github.com/agenticgokit/agenticgokit/core/vnext"

agent, err := vnext.NewBuilder("assistant").
    WithConfig(config).
    Build()
```

**New Code (v1beta):**
```go
import "github.com/agenticgokit/agenticgokit/v1beta"

agent, err := v1beta.NewBuilder().
    WithName("assistant").
    WithLLM("ollama", "gemma3:1b").
    Build()
```

### Complete Migration Path

#### 1. Update Imports

```diff
- import "github.com/agenticgokit/agenticgokit/core"
- import "github.com/agenticgokit/agenticgokit/core/vnext"
+ import "github.com/agenticgokit/agenticgokit/v1beta"
```

#### 2. Update Agent Creation

**vnext → v1beta:**
```diff
- agent, err := vnext.NewBuilder("myagent").
-     WithConfig(&vnext.Config{...}).
+ agent, err := v1beta.NewBuilder().
+     WithName("myagent").
+     WithLLM("openai", "gpt-4").
      Build()
```

**core → v1beta:**
```diff
- agent, err := core.NewAgent(config)
+ agent, err := v1beta.NewBuilder().
+     WithLLM("openai", "gpt-4").
+     Build()
```

#### 3. Update Configuration

**vnext.Config → v1beta.Config:**
```diff
- config := &vnext.Config{
-     Name: "assistant",
-     LLM: vnext.LLMConfig{
-         Provider: "openai",
-         Model: "gpt-4",
-     },
- }
+ agent, err := v1beta.NewBuilder().
+     WithName("assistant").
+     WithLLM("openai", "gpt-4").
+     WithAPIKey(os.Getenv("OPENAI_API_KEY")).
+     Build()
```

#### 4. Update Streaming

**vnext streaming → v1beta streaming:**
```diff
- stream, err := agent.RunStream(ctx, "Hello")
+ stream, err := agent.RunStream(ctx, "Hello")
  for chunk := range stream.Chunks() {
-     if chunk.Type == vnext.ChunkTypeDelta {
+     if chunk.Type == v1beta.ChunkTypeDelta {
          fmt.Print(chunk.Delta)
      }
  }
```

The streaming API is very similar, just update the package import!

#### 5. Update Workflows

**vnext workflows → v1beta workflows:**
```diff
- workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
+ workflow, err := v1beta.NewSequentialWorkflow(&v1beta.WorkflowConfig{
      Name: "pipeline",
      Timeout: 300 * time.Second,
  })
- workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
+ workflow.AddStep(v1beta.WorkflowStep{Name: "step1", Agent: agent1})
```

## What's Not Changing

The v1beta API provides the same functionality as vnext:

- ✅ Streaming execution
- ✅ Multi-agent workflows (Sequential, Parallel, DAG, Loop)
- ✅ Subworkflows
- ✅ Memory and RAG
- ✅ Tool integration (MCP)
- ✅ Multiple LLM providers

## Breaking Changes in v1.0

### Removed Packages
- ❌ `core` package (entire package removed)
- ❌ `core/vnext` package (entire package removed)

### What Stays
- ✅ `v1beta` becomes primary `v1` package
- ✅ All v1beta functionality preserved
- ✅ Import path becomes: `github.com/agenticgokit/agenticgokit/v1`

## Migration Checklist

- [ ] Identify all uses of `core` or `core/vnext` imports
- [ ] Update imports to `v1beta`
- [ ] Update agent creation to use `v1beta.NewBuilder()`
- [ ] Update config structs to v1beta types
- [ ] Update streaming code (minimal changes needed)
- [ ] Update workflow code (minimal changes needed)
- [ ] Test your application
- [ ] Remove old import statements

## Need Help?

- **Documentation**: See [v1beta/README.md](v1beta/README.md) for complete API reference
- **Examples**: Check [examples/](examples/) for working code samples
- **Issues**: [GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)

## API Comparison Table

| Feature | core/vnext | v1beta | Notes |
|---------|-----------|--------|-------|
| Package | `core/vnext` | `v1beta` | Import path change |
| Agent Creation | `NewBuilder(name)` | `NewBuilder().WithName(name)` | Builder pattern |
| Config | `WithConfig(&Config{...})` | Fluent methods | More ergonomic |
| Streaming | `RunStream()` | `RunStream()` | Same API |
| Workflows | `NewSequentialWorkflow()` | `NewSequentialWorkflow()` | Same API |
| Subworkflows | `NewSubWorkflowAgent()` | `NewSubWorkflowAgent()` | Same API |
| LLM Providers | OpenAI, Ollama, Azure | OpenAI, Ollama, Azure, HuggingFace | More providers |

## Support Policy

- **v0.x (Current)**: All packages available, deprecation warnings
- **v1.0 (Future)**: Only v1 (formerly v1beta) available
- **After v1.0**: No support for `core` or `core/vnext`

We recommend migrating to v1beta **immediately** to ensure a smooth transition to v1.0.
