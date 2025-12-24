# BentoML Quick Start

This example demonstrates how to use BentoML as an LLM provider with AgenticGoKit.

## What is BentoML?

BentoML is an open-source platform for building, shipping, and scaling AI applications. It provides:

- **Easy Model Packaging**: Package ML models with dependencies into standardized "Bentos"
- **OpenAI-Compatible API**: Expose models with an OpenAI-compatible REST API
- **Production-Ready**: Built-in observability, batching, and scaling features
- **Multi-Framework Support**: Works with PyTorch, TensorFlow, Transformers, etc.

## Prerequisites

1. **BentoML Service**: You need a running BentoML service with OpenAI-compatible API. 

   **Option A**: Using a pre-built Bento with OpenLLM:
   ```bash
   pip install openllm
   openllm start llama2 --backend vllm
   ```

   **Option B**: Create a custom BentoML service:
   ```python
   # service.py
   import bentoml
   from transformers import AutoModelForCausalLM, AutoTokenizer

   @bentoml.service(
       resources={"gpu": 1},
       traffic={"timeout": 300},
   )
   class LLMService:
       def __init__(self):
           self.model = AutoModelForCausalLM.from_pretrained("meta-llama/Llama-2-7b-chat-hf")
           self.tokenizer = AutoTokenizer.from_pretrained("meta-llama/Llama-2-7b-chat-hf")

       @bentoml.api
       def generate(self, prompt: str) -> str:
           inputs = self.tokenizer(prompt, return_tensors="pt")
           outputs = self.model.generate(**inputs, max_new_tokens=512)
           return self.tokenizer.decode(outputs[0], skip_special_tokens=True)
   ```

   Then serve it:
   ```bash
   bentoml serve service:LLMService --port 3000
   ```

   **Option C**: Using BentoCloud:
   ```bash
   bentoml deploy . --name my-llm-service
   ```

2. **Environment Variables** (optional):
   ```bash
   export BENTOML_BASE_URL="http://localhost:3000"
   export BENTOML_MODEL="llama2-7b-chat"
   ```

## Running the Example

```bash
cd examples/bentoml-quickstart
go run main.go
```

## Configuration Options

The BentoML provider supports these configuration options:

| Option | Description | Default |
|--------|-------------|---------|
| `BaseURL` | BentoML server URL | `http://localhost:3000` |
| `Model` | Model name | Required |
| `MaxTokens` | Maximum tokens to generate | `2048` |
| `Temperature` | Sampling temperature | `0.7` |
| `BentoMLTopP` | Nucleus sampling | - |
| `BentoMLTopK` | Top-k sampling | - |
| `BentoMLPresencePenalty` | Presence penalty | - |
| `BentoMLFrequencyPenalty` | Frequency penalty | - |
| `BentoMLStop` | Stop sequences | - |
| `BentoMLServiceName` | BentoML service name | - |
| `BentoMLRunners` | Specific runners to use | - |
| `BentoMLExtraHeaders` | Additional HTTP headers | - |
| `BentoMLMaxRetries` | Maximum retry attempts | `3` |
| `BentoMLRetryDelay` | Delay between retries | `1s` |

## Example Code

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/agenticgokit/agenticgokit/core"
    _ "github.com/agenticgokit/agenticgokit/plugins/llm/bentoml"
)

func main() {
    provider, err := core.NewModelProviderFromConfig(core.LLMProviderConfig{
        Type:        "bentoml",
        BaseURL:     "http://localhost:3000",
        Model:       "llama2-7b-chat",
        MaxTokens:   2048,
        Temperature: 0.7,
    })
    if err != nil {
        log.Fatal(err)
    }

    resp, err := provider.Call(context.Background(), core.Prompt{
        System: "You are a helpful assistant.",
        User:   "What is BentoML?",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Content)
}
```

## Using with OpenLLM

[OpenLLM](https://github.com/bentoml/OpenLLM) is BentoML's project for running open-source LLMs in production. It provides an OpenAI-compatible API out of the box:

```bash
# Install OpenLLM
pip install openllm

# Start a model (automatically downloads if needed)
openllm start mistral --backend vllm

# Or with a specific model
openllm start meta-llama/Llama-2-7b-chat-hf
```

Then configure AgenticGoKit:

```go
provider, _ := core.NewModelProviderFromConfig(core.LLMProviderConfig{
    Type:    "bentoml",
    BaseURL: "http://localhost:3000",
    Model:   "mistral",
})
```

## BentoML vs vLLM

| Feature | BentoML | vLLM |
|---------|---------|------|
| **Primary Use** | Full ML deployment platform | High-performance LLM inference |
| **Model Packaging** | Yes (Bentos) | No |
| **Multi-Framework** | Yes | LLMs only |
| **Batching** | Built-in | Continuous batching |
| **OpenAI API** | Via OpenLLM | Native |
| **Cloud Deploy** | BentoCloud | Various |

## Troubleshooting

1. **Connection refused**: Ensure BentoML service is running on the specified port
2. **Model not found**: Verify the model name matches your BentoML service configuration
3. **Timeout errors**: Increase `HTTPTimeout` for large models or slow hardware
4. **Authentication error**: Set API key if your BentoML service requires authentication

## References

- [BentoML Documentation](https://docs.bentoml.com/)
- [OpenLLM GitHub](https://github.com/bentoml/OpenLLM)
- [BentoCloud](https://www.bentoml.com/cloud)
