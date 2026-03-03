# Foundry Local Quickstart

A minimal example that shows how to use **Azure AI Foundry Local** as the LLM provider in AgenticGoKit.

## Prerequisites

1. Install [Azure AI Foundry Local](https://learn.microsoft.com/azure/ai-foundry/foundry-local/get-started):

   ```powershell
   winget install Microsoft.FoundryLocal
   ```

2. Start the service and load a model:

   ```powershell
   foundry model run qwen2.5-0.5b-instruct-generic-gpu:4
   ```

   The service listens on `http://localhost:5272/v1` by default.

3. Verify the service is running:

   ```powershell
   curl http://localhost:5272/v1/models
   ```

4. Check the model alias with:

   ```powershell
   foundry model list
   ```

   Update the `foundryModel` constant in `main.go` if your loaded model uses a different alias.

## Run

```powershell
cd examples/foundrylocal-quickstart
go mod tidy
go run main.go
```

## What the demo does

| Part | Description |
|------|-------------|
| **Basic Chat** | Sends three questions via `agent.Run()` and prints answers with token/latency stats. |
| **Streaming Chat** | Streams a short poem token-by-token via `agent.RunStream()`. |

## Configuration

| Constant | Default | Description |
|----------|---------|-------------|
| `foundryBaseURL` | `http://localhost:5272/v1` | Foundry Local API endpoint |
| `foundryModel` | `qwen2.5-0.5b-instruct-generic-gpu:4` | Model alias to use |

Both constants are at the top of `main.go` — edit them to match your setup.

## How it works

The `foundrylocal` plugin is registered via a blank import:

```go
import _ "github.com/agenticgokit/agenticgokit/plugins/llm/foundrylocal"
```

This registers the `"foundrylocal"` provider in the core factory. The adapter reuses
the OpenAI-compatible endpoint exposed by Foundry Local but **omits** the
`Authorization` header (Foundry Local rejects Bearer tokens).
