# MLFlow AI Gateway Demo

This example demonstrates how to use MLFlow AI Gateway as an LLM provider with AgenticGoKit.

## Prerequisites

1. **MLFlow AI Gateway**: You need a running MLFlow AI Gateway. Set it up with:
   ```bash
   pip install mlflow[gateway]

   # Start the gateway
   mlflow gateway start --config-path gateway-config.yaml --port 5001
   ```

2. **Gateway Configuration**: Create a `gateway-config.yaml` file (see example below)

3. **Environment Variables** (optional):
   ```bash
   export MLFLOW_GATEWAY_URL="http://localhost:5001"
   export MLFLOW_CHAT_ROUTE="chat"
   export MLFLOW_EMBEDDINGS_ROUTE="embeddings"
   ```

## Gateway Configuration Example

Create a `gateway-config.yaml` file. Note: MLFlow Gateway >= 2.9 uses `endpoints` with `endpoint_type`:

```yaml
# For MLFlow >= 2.9
endpoints:
  - name: chat
    endpoint_type: llm/v1/chat
    model:
      provider: openai
      name: gpt-4
      config:
        openai_api_key: $OPENAI_API_KEY

  - name: embeddings
    endpoint_type: llm/v1/embeddings
    model:
      provider: openai
      name: text-embedding-3-small
      config:
        openai_api_key: $OPENAI_API_KEY

  - name: anthropic-chat
    endpoint_type: llm/v1/chat
    model:
      provider: anthropic
      name: claude-3-sonnet-20240229
      config:
        anthropic_api_key: $ANTHROPIC_API_KEY
```

For older versions of MLFlow (< 2.9), use `routes` with `route_type`:

```yaml
# For MLFlow < 2.9
routes:
  - name: chat
    route_type: llm/v1/chat
    model:
      provider: openai
      name: gpt-4
      config:
        openai_api_key: $OPENAI_API_KEY
```

## Running the Example

```bash
cd examples/mlflow-gateway-demo
go run main.go
```

## Configuration Options

The MLFlow Gateway provider supports these configuration options:

| Option | Description | Default |
|--------|-------------|---------|
| `BaseURL` | MLFlow Gateway URL | `http://localhost:5001` |
| `MLFlowChatRoute` | Route name for chat completions | Required |
| `MLFlowEmbeddingsRoute` | Route name for embeddings | - |
| `MLFlowCompletionsRoute` | Route name for completions | - |
| `Model` | Override the route's default model | - |
| `APIKey` | Gateway API key (if configured) | - |
| `MaxTokens` | Maximum tokens to generate | `2048` |
| `Temperature` | Sampling temperature | `0.7` |
| `MLFlowMaxRetries` | Maximum retry attempts | `3` |
| `MLFlowRetryDelay` | Delay between retries | `1s` |
| `MLFlowExtraHeaders` | Additional HTTP headers | - |
| `MLFlowTopP` | Nucleus sampling | - |
| `MLFlowStop` | Stop sequences | - |

## Example Code

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/agenticgokit/agenticgokit/core"
    _ "github.com/agenticgokit/agenticgokit/plugins/llm/mlflow"
)

func main() {
    provider, err := core.NewModelProviderFromConfig(core.LLMProviderConfig{
        Type:            "mlflow",
        BaseURL:         "http://localhost:5001",
        MLFlowChatRoute: "chat",
        MaxTokens:       2048,
        Temperature:     0.7,
    })
    if err != nil {
        log.Fatal(err)
    }

    resp, err := provider.Call(context.Background(), core.Prompt{
        System: "You are a helpful assistant.",
        User:   "What are the benefits of using an AI Gateway?",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Content)
}
```

## Switching Routes Dynamically

You can create multiple providers pointing to different routes:

```go
// GPT-4 via OpenAI route
gpt4Provider, _ := core.NewModelProviderFromConfig(core.LLMProviderConfig{
    Type:            "mlflow",
    BaseURL:         "http://localhost:5001",
    MLFlowChatRoute: "chat",
})

// Claude via Anthropic route
claudeProvider, _ := core.NewModelProviderFromConfig(core.LLMProviderConfig{
    Type:            "mlflow",
    BaseURL:         "http://localhost:5001",
    MLFlowChatRoute: "anthropic-chat",
})
```

## MLFlow AI Gateway Benefits

- **Unified Interface**: Single API for multiple LLM providers
- **Provider Abstraction**: Seamlessly switch between OpenAI, Anthropic, Cohere, etc.
- **Rate Limiting**: Built-in rate limiting and quota management
- **Credential Management**: Centralized API key management
- **Request/Response Logging**: Enterprise-grade observability
- **Route-based Routing**: Define custom routes to different models

## Troubleshooting

1. **Connection refused**: Ensure MLFlow Gateway is running on the specified port
2. **Route not found**: Verify the route name matches your gateway configuration
3. **Authentication error**: Check that API keys are properly configured in the gateway
4. **Rate limited**: Adjust `MLFlowMaxRetries` and `MLFlowRetryDelay` settings
