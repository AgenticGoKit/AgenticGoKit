# OpenRouter LLM Driver for AgenticGoKit

This package provides OpenRouter integration for AgenticGoKit, enabling access to multiple AI models from various providers through a unified OpenAI-compatible API.

## Features

- **Multi-Provider Access**: Access models from OpenAI, Anthropic, Google, Meta, and many others
- **OpenAI-Compatible API**: Easy migration from OpenAI-based implementations
- **Streaming Support**: Full support for streaming responses (SSE)
- **Site Tracking**: Optional headers for OpenRouter rankings and analytics
- **Model Routing**: Automatic model selection and fallback support
- **Cost Optimization**: Access to cost-efficient model alternatives

## Installation

```bash
go get github.com/kunalkushwaha/agenticgokit
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kunalkushwaha/agenticgokit/core"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/openrouter"
)

func main() {
    config := core.LLMProviderConfig{
        Type:        "openrouter",
        APIKey:      "sk-or-v1-...",
        Model:       "openai/gpt-3.5-turbo",
        MaxTokens:   500,
        Temperature: 0.7,
    }

    provider, err := core.NewModelProviderFromConfig(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    response, err := provider.Call(ctx, core.Prompt{
        System: "You are a helpful assistant.",
        User:   "What is OpenRouter?",
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response.Content)
}
```

### Using Environment Variables

Set the `OPENROUTER_API_KEY` environment variable:

```bash
export OPENROUTER_API_KEY=sk-or-v1-...
```

Then use `AgentLLMConfig`:

```go
config := core.AgentLLMConfig{
    Provider:    "openrouter",
    Model:       "anthropic/claude-3-sonnet",
    MaxTokens:   2000,
    Temperature: 0.7,
}

provider, err := core.NewLLMProvider(config)
```

### Streaming Responses

```go
tokenChan, err := provider.Stream(ctx, core.Prompt{
    User: "Tell me a story.",
})

if err != nil {
    log.Fatal(err)
}

for token := range tokenChan {
    if token.Error != nil {
        log.Fatal(token.Error)
    }
    fmt.Print(token.Content)
}
```

## Configuration

### Configuration Options

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Type` | string | Provider type (must be "openrouter") | Required |
| `APIKey` | string | OpenRouter API key | Required (or env var) |
| `Model` | string | Model identifier (e.g., "openai/gpt-4") | "openai/gpt-3.5-turbo" |
| `MaxTokens` | int | Maximum tokens to generate | 2000 |
| `Temperature` | float64 | Sampling temperature (0.0-2.0) | 0.7 |
| `BaseURL` | string | API base URL | "https://openrouter.ai/api/v1" |
| `SiteURL` | string | Your site URL (for rankings) | Optional |
| `SiteName` | string | Your site name (for rankings) | Optional |

### Environment Variables

- `OPENROUTER_API_KEY` - Your OpenRouter API key
- `OPENROUTER_BASE_URL` - Custom API base URL (optional)
- `OPENROUTER_SITE_URL` - Site URL for tracking (optional)
- `OPENROUTER_SITE_NAME` - Site name for tracking (optional)

## Available Models

OpenRouter provides access to numerous models across providers:

### OpenAI Models
- `openai/gpt-4`
- `openai/gpt-4-turbo`
- `openai/gpt-3.5-turbo`

### Anthropic Models
- `anthropic/claude-3-opus`
- `anthropic/claude-3-sonnet`
- `anthropic/claude-3-haiku`

### Google Models
- `google/gemini-pro`
- `google/gemini-pro-vision`

### Meta Models
- `meta-llama/llama-3-70b-instruct`
- `meta-llama/llama-3-8b-instruct`

### And Many More!
Visit [OpenRouter Models](https://openrouter.ai/models) for the full list.

## Advanced Usage

### Site Tracking

Enable site tracking to appear on OpenRouter rankings:

```go
config := core.LLMProviderConfig{
    Type:        "openrouter",
    APIKey:      "sk-or-v1-...",
    Model:       "anthropic/claude-3-sonnet",
    SiteURL:     "https://myapp.com",
    SiteName:    "My Awesome App",
}
```

This sets the `HTTP-Referer` and `X-Title` headers for proper attribution.

### Parameter Overrides

Override default parameters per request:

```go
maxTokens := int32(100)
temperature := float32(0.3)

response, err := provider.Call(ctx, core.Prompt{
    User: "Brief answer please.",
    Parameters: core.ModelParameters{
        MaxTokens:   &maxTokens,
        Temperature: &temperature,
    },
})
```

### Multiple Models

Easily switch between models:

```go
models := []string{
    "openai/gpt-4",
    "anthropic/claude-3-sonnet",
    "google/gemini-pro",
}

for _, model := range models {
    config.Model = model
    provider, _ := core.NewModelProviderFromConfig(config)
    response, _ := provider.Call(ctx, prompt)
    fmt.Printf("%s: %s\n", model, response.Content)
}
```

## Error Handling

The adapter provides detailed error messages:

```go
response, err := provider.Call(ctx, prompt)
if err != nil {
    if strings.Contains(err.Error(), "API key") {
        log.Fatal("Invalid API key")
    } else if strings.Contains(err.Error(), "rate limit") {
        log.Fatal("Rate limit exceeded")
    } else {
        log.Fatalf("API error: %v", err)
    }
}
```

## Testing

Run unit tests:

```bash
go test ./internal/llm/openrouter_adapter_test.go
```

Run integration tests (requires API key):

```bash
export OPENROUTER_API_KEY=sk-or-v1-...
go test ./plugins/llm/openrouter/...
```

## Limitations

- **Embeddings**: Not currently supported by this adapter. Use a dedicated embedding provider.
- **Function Calling**: Not yet implemented (planned for future release).
- **Vision Models**: Supported by OpenRouter but requires additional image handling (coming soon).

## Cost Management

OpenRouter provides competitive pricing across models. Check the [OpenRouter Pricing](https://openrouter.ai/docs#models) page for current rates.

Tips:
- Use cheaper models for simple tasks (e.g., `openai/gpt-3.5-turbo`)
- Use more powerful models for complex reasoning (e.g., `anthropic/claude-3-opus`)
- Monitor usage through OpenRouter dashboard

## Troubleshooting

### Invalid API Key

```
Error: OpenRouter API error [invalid_api_key]: Invalid API key provided
```

**Solution**: Check your API key or set `OPENROUTER_API_KEY` environment variable.

### Model Not Found

```
Error: OpenRouter API error [model_not_found]: Model not found
```

**Solution**: Verify model name at [OpenRouter Models](https://openrouter.ai/models).

### Rate Limit Exceeded

```
Error: OpenRouter API error [rate_limit_exceeded]: Rate limit exceeded
```

**Solution**: Implement retry logic with exponential backoff or upgrade your plan.

## Examples

See the [examples/openrouter_usage](../../examples/openrouter_usage/) directory for complete examples.

## Contributing

Contributions are welcome! Please see the [Contributing Guide](../../../docs/contributors/ContributorGuide.md).

## License

This project is licensed under the same license as AgenticGoKit. See [LICENSE](../../../LICENSE) for details.

## Resources

- [OpenRouter Documentation](https://openrouter.ai/docs)
- [OpenRouter Models](https://openrouter.ai/models)
- [OpenRouter API](https://openrouter.ai/docs#api)
- [AgenticGoKit Documentation](../../../docs/)

## Support

- GitHub Issues: [agenticgokit/agenticgokit](https://github.com/kunalkushwaha/agenticgokit/issues)
- OpenRouter Discord: [Join Here](https://discord.gg/openrouter)
