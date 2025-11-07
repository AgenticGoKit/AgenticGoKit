# OpenRouter QuickStart - vNext API Example

This example demonstrates how to use the AgenticGoKit vNext API with OpenRouter as the LLM provider.

## Overview

OpenRouter provides unified access to multiple LLM providers through a single API. This example shows various patterns for creating and using agents with OpenRouter, including:

1. **Basic Agent Creation** - Using Config struct
2. **Preset Chat Agent** - Using factory functions with options
3. **Streaming Responses** - Real-time token streaming
4. **Multiple Models** - Testing different OpenRouter models
5. **RunOptions** - Detailed execution results
6. **Site Tracking** - OpenRouter rankings and analytics
7. **Custom Handlers** - Custom logic with LLM fallback
8. **Simple Setup** - Using WithLLM option for quick configuration

## Prerequisites

1. **OpenRouter API Key**: Sign up at [OpenRouter.ai](https://openrouter.ai/) to get your API key
2. **Set Environment Variable**:
   ```powershell
   # PowerShell
   $env:OPENROUTER_API_KEY = "sk-or-v1-..."
   ```
   
   ```bash
   # Bash/Linux
   export OPENROUTER_API_KEY="sk-or-v1-..."
   ```

## Running the Example

```powershell
# Navigate to the example directory
cd examples/vnext/openrouter-quickstart

# Run the example
go run main.go
```

## Example Breakdown

### 1. Basic Agent with Config

Creates an agent using a complete `Config` struct with API key from environment:

```go
// Read API key from environment
apiKey := os.Getenv("OPENROUTER_API_KEY")

config := &vnext.Config{
    Name:         "openrouter-assistant",
    SystemPrompt: "You are a helpful assistant.",
    Timeout:      30 * time.Second,
    LLM: vnext.LLMConfig{
        Provider:    "openrouter",
        Model:       "openai/gpt-3.5-turbo",
        APIKey:      apiKey, // Pass API key explicitly
        Temperature: 0.7,
        MaxTokens:   500,
    },
}

agent, err := vnext.NewBuilder("openrouter-assistant").
    WithConfig(config).
    Build()
```

**Important**: The vNext API requires you to explicitly pass the API key in the config. It does not automatically read from environment variables like the core API does.

### 2. Preset Chat Agent

Uses a complete `Config` struct with the builder pattern:

```go
chatConfig := &vnext.Config{
    Name:         "chat-bot",
    SystemPrompt: "You are a conversational assistant.",
    Timeout:      30 * time.Second,
    LLM: vnext.LLMConfig{
        Provider:    "openrouter",
        Model:       "anthropic/claude-3-haiku",
        APIKey:      apiKey, // Must include API key
        Temperature: 0.7,
        MaxTokens:   100,
    },
}

agent, err := vnext.NewBuilder("chat-bot").
    WithConfig(chatConfig).
    Build()
```

### 3. Streaming Responses

Demonstrates real-time streaming with chunk processing:

```go
stream, err := agent.RunStream(ctx, "Write a haiku about coding.",
    vnext.WithBufferSize(10),
    vnext.WithThoughts(),
)

for chunk := range stream.Chunks() {
    if chunk.Type == vnext.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
}
```

### 4. Multiple Models

Tests different LLM providers available through OpenRouter:

- OpenAI GPT-3.5 Turbo
- Anthropic Claude 3 Haiku
- Google Gemini 2.0 Flash (free tier)
- Meta Llama 3.1 8B Instruct

### 5. RunOptions for Detailed Results

Uses `RunOptions` to get comprehensive execution information:

```go
opts := vnext.RunWithDetailedResult().
    SetTimeout(30 * time.Second).
    AddContext("request_id", "demo-123")

result, err := agent.RunWithOptions(ctx, query, opts)
```

### 6. Site Tracking

Enables OpenRouter rankings and analytics by providing site information:

```go
LLM: vnext.LLMConfig{
    Provider:    "openrouter",
    Model:       "openai/gpt-3.5-turbo",
    SiteURL:     "https://myapp.com",     // Your app URL
    SiteName:    "My Awesome App",        // Your app name
}
```

This sets the `HTTP-Referer` and `X-Title` headers in requests to OpenRouter.

### 7. Custom Handler

Implements custom logic with LLM fallback:

```go
customHandler := func(ctx context.Context, input string, caps *vnext.Capabilities) (string, error) {
    if len(input) < 10 {
        // Short query: expand it
        return caps.LLM(
            "You are a helpful assistant that expands short queries.",
            fmt.Sprintf("Expand this query: %s", input),
        )
    }
    // Normal processing
    return caps.LLM("You are a helpful assistant.", input)
}
```

### 8. Preset with Custom Prompt

Uses a preset configuration with custom system prompt:

```go
config8 := &vnext.Config{
    Name:         "simple-agent",
    SystemPrompt: "You are a helpful assistant specialized in programming languages.",
    Timeout:      30 * time.Second,
    LLM: vnext.LLMConfig{
        Provider:    "openrouter",
        Model:       "openai/gpt-3.5-turbo",
        APIKey:      apiKey, // API key required
        Temperature: 0.7,
        MaxTokens:   200,
    },
}

agent, err := vnext.NewBuilder("simple-agent").
    WithConfig(config8).
    Build()
```

**Note**: While the vNext API has `WithLLM()` and `WithLLMConfig()` options, they don't support setting the API key. Always use the full `Config` struct approach for OpenRouter.

## OpenRouter Models

OpenRouter provides access to models from various providers. Some free-tier options include:

- `openai/gpt-3.5-turbo` - Fast and efficient
- `anthropic/claude-3-haiku` - Balanced performance
- `google/gemini-2.0-flash-exp:free` - Google's latest
- `meta-llama/llama-3.1-8b-instruct` - Open source

For the full list of available models and pricing, visit [OpenRouter Models](https://openrouter.ai/models).

## Configuration Options

### LLMConfig Fields

- `Provider`: Set to `"openrouter"`
- `Model`: Model identifier (e.g., `"openai/gpt-3.5-turbo"`)
- `APIKey`: **Required** - Your OpenRouter API key (read from environment)
- `Temperature`: 0.0-2.0 (controls randomness)
- `MaxTokens`: Maximum tokens to generate
- `SiteURL`: (Optional) Your application URL for rankings
- `SiteName`: (Optional) Your application name for analytics

### Agent Configuration

- `Name`: Agent identifier
- `SystemPrompt`: System-level instructions
- `Timeout`: Execution timeout duration

### API Key Configuration

**Important**: The vNext framework requires explicit API key configuration. Unlike the core API which automatically reads environment variables, you must:

1. Read the environment variable yourself:
   ```go
   apiKey := os.Getenv("OPENROUTER_API_KEY")
   ```

2. Pass it explicitly in the config:
   ```go
   LLM: vnext.LLMConfig{
       APIKey: apiKey, // Required!
       // ... other settings
   }
   ```

## Best Practices

1. **API Key Security**: Never hardcode API keys - always read from environment variables
   ```go
   apiKey := os.Getenv("OPENROUTER_API_KEY")
   if apiKey == "" {
       log.Fatal("OPENROUTER_API_KEY not set")
   }
   ```

2. **Config Pattern**: Always use the full `Config` struct with API key:
   ```go
   config := &vnext.Config{
       LLM: vnext.LLMConfig{
           APIKey: apiKey, // Required for vNext API
           // ... other settings
       },
   }
   ```

3. **Model Selection**: Choose models based on your use case:
   - GPT-3.5 for general tasks
   - Claude for reasoning and analysis
   - Gemini for multimodal tasks
   - Llama for cost-sensitive applications

4. **Timeout Configuration**: Set appropriate timeouts based on expected response time

5. **Site Tracking**: Always set SiteURL and SiteName for production apps

6. **Error Handling**: Always check for errors and handle them appropriately

7. **Context Management**: Use `defer agent.Cleanup(ctx)` to ensure proper cleanup

## Troubleshooting

### "OPENROUTER_API_KEY environment variable not set"

Set the environment variable:
```powershell
# PowerShell
$env:OPENROUTER_API_KEY = "sk-or-v1-your-api-key-here"
```

```bash
# Bash/Linux
export OPENROUTER_API_KEY="sk-or-v1-your-api-key-here"
```

### "API key is required for OpenRouter provider"

This means the API key wasn't passed in the config. Make sure you're setting it:

```go
config := &vnext.Config{
    LLM: vnext.LLMConfig{
        Provider: "openrouter",
        APIKey:   apiKey, // ← Must be set!
        Model:    "openai/gpt-3.5-turbo",
    },
}
```

**Do not use**: `WithLLM()` or `WithLLMConfig()` options alone - they don't support API keys.

### "404 - No endpoints found" or "400 - not a valid model ID"

The model name is incorrect or unavailable. Check [OpenRouter Models](https://openrouter.ai/models) for valid model IDs.

### Connection Timeout

Increase the timeout in the config:
```go
Timeout: 60 * time.Second,  // Increase from 30s to 60s
```

## Related Documentation

- [vNext Framework README](../../../core/vnext/README.md)
- [OpenRouter Integration Design](../../../docs/design/OpenRouterIntegration.md)
- [Streaming Guide](../../../core/vnext/STREAMING_GUIDE.md)
- [Migration Guide](../../../core/vnext/MIGRATION_GUIDE.md)

## Support

For issues specific to:
- **AgenticGoKit**: Open an issue on GitHub
- **OpenRouter API**: Visit [OpenRouter Support](https://openrouter.ai/docs)
