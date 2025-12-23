# vLLM Quick Start

This example demonstrates how to use vLLM as an LLM provider with AgenticGoKit.

## Prerequisites

1. **vLLM Server**: You need a running vLLM server. Start one with:
   ```bash
   # Using pip
   pip install vllm
   python -m vllm.entrypoints.openai.api_server \
       --model meta-llama/Llama-2-7b-chat-hf \
       --port 8000

   # Or using Docker
   docker run --runtime nvidia --gpus all \
       -p 8000:8000 \
       --ipc=host \
       vllm/vllm-openai:latest \
       --model meta-llama/Llama-2-7b-chat-hf
   ```

2. **Environment Variables** (optional):
   ```bash
   export VLLM_BASE_URL="http://localhost:8000"
   export VLLM_MODEL="meta-llama/Llama-2-7b-chat-hf"
   ```

## Running the Example

```bash
cd examples/vllm-quickstart
go run main.go
```

## Configuration Options

The vLLM provider supports these configuration options:

| Option | Description | Default |
|--------|-------------|---------|
| `BaseURL` | vLLM server URL | `http://localhost:8000` |
| `Model` | Model name/path | Required |
| `MaxTokens` | Maximum tokens to generate | `2048` |
| `Temperature` | Sampling temperature | `0.7` |
| `VLLMTopP` | Nucleus sampling | - |
| `VLLMTopK` | Top-k sampling | - |
| `VLLMPresencePenalty` | Presence penalty | - |
| `VLLMFrequencyPenalty` | Frequency penalty | - |
| `VLLMRepetitionPenalty` | Repetition penalty | - |
| `VLLMBestOf` | Generate N completions, return best | - |
| `VLLMUseBeamSearch` | Use beam search | `false` |
| `VLLMStopTokenIds` | Token IDs to stop generation | - |
| `VLLMStop` | Stop sequences | - |

## Example Code

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/agenticgokit/agenticgokit/core"
    _ "github.com/agenticgokit/agenticgokit/plugins/llm/vllm"
)

func main() {
    provider, err := core.NewModelProviderFromConfig(core.LLMProviderConfig{
        Type:        "vllm",
        BaseURL:     "http://localhost:8000",
        Model:       "meta-llama/Llama-2-7b-chat-hf",
        MaxTokens:   2048,
        Temperature: 0.7,
        VLLMTopP:    0.9,
    })
    if err != nil {
        log.Fatal(err)
    }

    resp, err := provider.Call(context.Background(), core.Prompt{
        System: "You are a helpful assistant.",
        User:   "Explain quantum computing in simple terms.",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Content)
}
```

## vLLM Features

vLLM offers several advantages:

- **High Throughput**: Up to 24x higher throughput than HuggingFace Transformers
- **PagedAttention**: Memory-efficient attention mechanism
- **Continuous Batching**: Dynamic batching for optimal throughput
- **OpenAI-Compatible API**: Drop-in replacement for OpenAI endpoints
- **Quantization Support**: AWQ, GPTQ, INT8, FP8 quantization

## Troubleshooting

1. **Connection refused**: Ensure vLLM server is running on the specified port
2. **Model not found**: Verify the model name matches what's loaded in vLLM
3. **Out of memory**: Try using a smaller model or enabling quantization
4. **Slow responses**: Check GPU utilization and consider using tensor parallelism
