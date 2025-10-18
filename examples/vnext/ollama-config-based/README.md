# Ollama Config-Based Agent Example

> ⚠️ **IMPORTANT**: This demonstrates **configuration patterns** for vNext. The agent returns mock responses currently. For working LLM integration, use `core.SimpleAgent` API. See [IMPLEMENTATION_STATUS.md](../IMPLEMENTATION_STATUS.md).

This example demonstrates **TOML-based configuration** for AgenticGoKit vNext agents.

## Features

- ✅ Configuration loaded from TOML file
- ✅ Separation of code and configuration
- ✅ Environment-specific configs (dev, prod)
- ✅ Easy to modify without recompiling
- ✅ Demonstrates production config patterns

## Quick Start

```bash
# Ensure Ollama is running
ollama pull llama3.2

# Run with default config
cd examples/vnext/ollama-config-based
go run main.go

# Run with custom config
go run main.go my-config.toml
```

## Configuration File

The `config.toml` file contains all agent settings:

```toml
name = "ollama-helper"
system_prompt = "You are a helpful assistant..."
timeout = "30s"

[llm]
provider = "ollama"
model = "llama3.2"
temperature = 0.3
max_tokens = 200
```

## Code Highlights

### Loading Configuration

```go
// Load from file
config, err := vnext.LoadConfigFromTOML("config.toml")

// Build agent from config
agent, err := vnext.NewBuilder(config.Name).
    WithConfig(config).
    WithPreset(vnext.ChatAgent).
    Build()
```

## Benefits of TOML Configuration

- 📝 **Easy to Read**: Human-friendly format
- 🔧 **Easy to Modify**: Change settings without recompiling
- 🌍 **Environment-Specific**: Different configs for dev/staging/prod
- ✅ **Validation**: Built-in config validation
- 🔐 **Environment Variables**: Support for `${ENV_VAR}` substitution

## Creating Multiple Configs

```bash
# Development config
config.dev.toml

# Production config
config.prod.toml

# Run with specific config
go run main.go config.prod.toml
```

## Next Steps

- Add environment variables: `api_key = "${OLLAMA_API_KEY}"`
- Add memory configuration: `[memory]` section
- Add tools configuration: `[tools]` section
- Try [Builder Pattern Example](../ollama-short-answer/)
