# Configuration (vNext)

**Unified agent and project configuration via Go structs and TOML**

The vNext configuration stack replaces disparate config types with a single `Config` struct plus helpers for TOML loading, environment interpolation, and validation. Use it directly, feed it to the builder, or load an entire multi-agent project definition.

## 🧾 Core Config Struct

```go
type Config struct {
    Name         string
    SystemPrompt string
    Timeout      time.Duration
    DebugMode    bool

    LLM      LLMConfig
    Memory   *MemoryConfig
    Tools    *ToolsConfig
    Workflow *WorkflowConfig
    Tracing  *TracingConfig
    Streaming *StreamingConfig
}
```

Key nested structures:

- `LLMConfig` → provider, model, temperature, max tokens, optional base URL / API key
- `MemoryConfig` → provider, connection string, optional `RAGConfig`, arbitrary options
- `ToolsConfig` → tool orchestration, MCP config, caching, circuit breaker settings
- `WorkflowConfig` → execution mode (`sequential`, `parallel`, `dag`, `loop`), agent list, timeout, max iterations
- `StreamingConfig` → buffer size, flush interval, metadata/tool chunk toggles

## 📥 Loading from TOML

```go
cfg, err := vnext.LoadConfigFromTOML("./agent.toml")
if err != nil {
    log.Fatal(err)
}

agent, err := vnext.NewBuilder(cfg.Name).
    WithConfig(cfg).
    Build()
```

Environment variables are resolved automatically for strings containing `${NAME}` (with optional default `${NAME:default}` syntax).

### Example TOML

```toml
name = "support-bot"
system_prompt = "You are a helpful assistant"

[llm]
provider = "openai"
model = "gpt-4o"
temperature = 0.6
max_tokens = 2048
api_key = "${OPENAI_API_KEY}"

[memory]
provider = "memory"
connection = "localhost"

[memory.rag]
max_tokens = 4000
personal_weight = 0.3
knowledge_weight = 0.7
history_limit = 10

[tools]
enabled = true
max_retries = 2
timeout = "30s"

[tools.mcp]
enabled = true
servers = [{ name = "fs", type = "stdio", command = "mcp-fs" }]
```

Use standard Go `time.Duration` strings (`"30s"`, `"2m"`) for duration fields.

## 🧪 Validation

`LoadConfigFromTOML` and `ValidateConfig` enforce required fields and ranges:

- `name`, `llm.provider`, `llm.model` must be present
- Supported providers: `openai`, `ollama`, `azure`, `anthropic`
- Temperatures must be 0.0–2.0; max tokens must be positive
- Memory, tools, MCP, workflow configs run through detailed validators with severity levels

Critical issues return an error. Non-critical warnings are collected via `ValidationError` when using the validation helpers directly.

```go
if err := vnext.ValidateConfig(cfg); err != nil {
    log.Fatalf("invalid config: %v", err)
}
```

## 🛠 Manual Construction

```go
cfg := vnext.NewConfig("ops-bot",
    vnext.WithLLM("ollama", "llama3.2"),
    vnext.WithSystemPrompt("You answer operations questions"),
)

cfg.Memory = &vnext.MemoryConfig{Provider: "memory"}
cfg.Tools = vnext.DefaultToolsConfig()
```

`NewConfig` applies reasonable defaults (30s timeout, temp 0.7, max tokens 1000). Combine with the functional options defined in `builder.go`.

## 🧩 Project Configs

For multi-agent orchestration, use `ProjectConfig` and `LoadProjectConfigFromTOML`:

```go
project, err := vnext.LoadProjectConfigFromTOML("project.toml")
if err != nil {
    log.Fatal(err)
}

for name, agentCfg := range project.Agents {
    agent, err := vnext.NewBuilder(name).
        WithConfig(&agentCfg.Config).
        Build()
    // store agent instance...
}
```

`ProjectConfig` adds:

- `ProjectInfo` (name/version/description)
- Provider-specific blocks (`[providers.ollama]`, `[providers.openai]`, etc.)
- Global defaults for LLM, memory, MCP, workflow
- `Agents` map with per-agent overrides (inherits base config fields)

`ValidateProjectConfig` returns a slice of `ValidationError` for all agents; use `HasCriticalErrors` to determine whether to halt.

## 🔄 Helpers

- `MergeConfigs(cfgA, cfgB, ...)` merges non-zero values, later configs override earlier ones
- `CloneConfig(cfg)` deep copies nested pointers (RAG, tools, workflow, tracing) for safe mutation
- `DefaultConfig(name)` / `DefaultProjectConfig(name)` provide fully populated baseline structs

## 🌐 Environment Interpolation

The resolver replaces `${VAR}` placeholders in string fields with `os.LookupEnv`. Supply defaults using `${VAR:foo}`. Unresolved placeholders are left intact so you can detect missing secrets at runtime.

```go
os.Setenv("OPENAI_API_KEY", "sk-...")
cfg, _ := vnext.LoadConfigFromTOML("agent.toml")
fmt.Println(cfg.LLM.APIKey) // => sk-...
```

## 🔗 Related Docs

- [builder.md](builder.md) to feed configs into the streamlined builder
- [tools.md](tools.md) for MCP and tool-specific config fields
- [workflow.md](workflow.md) if you are wiring `WorkflowConfig` for orchestration
