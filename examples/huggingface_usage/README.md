# Hugging Face LLM Integration Examples

This directory contains comprehensive examples demonstrating how to use Hugging Face's various APIs with AgenticGoKit.

## Quick Start

The simplest way to get started with HuggingFace:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/kunalkushwaha/agenticgokit/core"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/huggingface"
)

func main() {
    // Get API key from environment
    apiKey := os.Getenv("HUGGINGFACE_API_KEY")
    if apiKey == "" {
        log.Fatal("Please set HUGGINGFACE_API_KEY environment variable")
    }

    // Create a simple configuration using the new router
    config := core.LLMProviderConfig{
        Type:        "huggingface",
        APIKey:      apiKey,
        Model:       "meta-llama/Llama-3.2-1B-Instruct",
        MaxTokens:   200,
        Temperature: 0.7,
        HFAPIType:   "inference",
    }

    // Create provider and make a call
    provider, _ := core.NewModelProviderFromConfig(config)
    response, _ := provider.Call(context.Background(), core.Prompt{
        User: "What is AI?",
    })
    
    fmt.Println(response.Content)
}
```

## Prerequisites

1. **Hugging Face API Key**: Get your API key from [Hugging Face](https://huggingface.co/settings/tokens)
2. **Set Environment Variables**:
   ```bash
   export HUGGINGFACE_API_KEY="your-api-key-here"
   export HUGGINGFACE_API_TYPE="inference"  # Optional, defaults to "inference"
   ```

## ⚠️ API Update Notice

**As of late 2024**, Hugging Face has migrated to a new router-based architecture:
- **Old API** (deprecated): `api-inference.huggingface.co` → Returns 410 Gone
- **New API**: `router.huggingface.co` → Uses OpenAI-compatible format
- The new router supports multiple providers and uses chat completions format
- Model catalog has changed - use provider-based models like `meta-llama/Llama-3.2-1B-Instruct`

## Supported API Types

The Hugging Face integration supports four different API types:

### 1. Inference API (Default) - **Updated for New Router**
- **Type**: `inference`
- **Description**: New router-based API with OpenAI-compatible format
- **Use Case**: Quick prototyping, small-scale applications
- **Base URL**: `https://router.huggingface.co` (automatic)
- **Endpoint**: `/v1/chat/completions`
- **Requires API Key**: Yes
- **Models**: Router-compatible models (e.g., `meta-llama/Llama-3.2-1B-Instruct`, `meta-llama/Llama-3.2-3B-Instruct`)
- **Format**: OpenAI-compatible chat completions

### 2. Chat API (Legacy)
- **Type**: `chat`
- **Description**: Legacy chat API type (now merged with inference type)
- **Use Case**: Backward compatibility
- **Note**: The new router API uses OpenAI-compatible chat format for all models
- **Requires API Key**: Yes
- **Models**: Same as inference type

### 3. Inference Endpoints
- **Type**: `endpoint`
- **Description**: Dedicated inference endpoints with guaranteed uptime
- **Use Case**: Production deployments requiring reliability
- **Base URL**: Your endpoint URL (e.g., `https://xxx.endpoints.huggingface.cloud`)
- **Requires API Key**: Yes
- **Models**: Any model deployed to your endpoint

### 4. Text Generation Inference (TGI)
- **Type**: `tgi`
- **Description**: Self-hosted optimized text generation server
- **Use Case**: On-premise deployments, custom infrastructure
- **Base URL**: Your TGI server URL (e.g., `http://localhost:8080`)
- **Requires API Key**: No (for local deployment)
- **Models**: Any model loaded in your TGI server

## Running the Examples

1. **Install dependencies**:
   ```bash
   cd examples/huggingface_usage
   go mod download
   ```

2. **Set up environment variables**:
   ```bash
   export HUGGINGFACE_API_KEY="your-api-key-here"
   
   # Optional: For Inference Endpoints example
   export HUGGINGFACE_ENDPOINT_URL="https://your-endpoint.endpoints.huggingface.cloud"
   
   # Optional: For TGI example (if running local TGI server)
   export HUGGINGFACE_TGI_URL="http://localhost:8080"
   ```

3. **Run the examples**:
   ```bash
   go run main.go
   ```

## Example Breakdown

The main example file demonstrates:

### Example 1: Basic Inference API Usage (New Router)
Simple text generation using the new router API with Llama models and OpenAI-compatible format.

### Example 2: Using Different Models on the Router
Demonstrating various models available through the new router infrastructure.

### Example 3: Inference Endpoints
Connecting to a dedicated Inference Endpoint for production use.

### Example 4: Text Generation Inference (TGI)
Using a self-hosted TGI server for on-premise deployments.

### Example 5: Streaming Responses
Real-time token-by-token streaming for responsive applications.

### Example 6: Embeddings API
Note about embeddings requiring separate endpoint configuration.

### Example 7: Environment Variable Configuration
Automatic configuration from environment variables using `AgentLLMConfig`.

### Example 8: Advanced Parameters
Fine-tuning generation with Hugging Face-specific parameters:
- `HFWaitForModel`: Wait for model loading
- `HFUseCache`: Control response caching
- `HFDoSample`: Enable/disable sampling
- `HFTopP`: Nucleus sampling parameter
- `HFTopK`: Top-k sampling parameter
- `HFRepetitionPenalty`: Penalize repeated tokens
- `HFStopSequences`: Define custom stop sequences

## Configuration Options

### Basic Configuration (New Router)
```go
config := core.LLMProviderConfig{
    Type:        "huggingface",
    APIKey:      "your-api-key",
    Model:       "meta-llama/Llama-3.2-1B-Instruct", // Router-compatible model
    MaxTokens:   150,
    Temperature: 0.7,
    HFAPIType:   "inference", // Uses new router with OpenAI format
}
```

### Advanced Configuration
```go
config := core.LLMProviderConfig{
    Type:                "huggingface",
    APIKey:              "your-api-key",
    Model:               "meta-llama/Llama-3.2-1B-Instruct", // Router model
    BaseURL:             "", // Auto-detected for inference
    MaxTokens:           150,
    Temperature:         0.7,
    HFAPIType:           "inference",
    HFWaitForModel:      true,
    HFUseCache:          true,
    HFDoSample:          true,
    HFTopP:              0.9,
    HFTopK:              50,
    HFRepetitionPenalty: 1.2,
    HFStopSequences:     []string{"\n\n", "END"},
}
```

## Popular Models (New Router)

### Text Generation (Router-Compatible)
- `meta-llama/Llama-3.2-1B-Instruct` - Small, efficient Llama model
- `meta-llama/Llama-3.2-3B-Instruct` - Larger Llama model
- `deepseek-ai/DeepSeek-R1` - DeepSeek reasoning model
- Use `:fastest` suffix for fastest available provider
- Use `:cheapest` suffix for most cost-effective provider

### Legacy Models (Inference Endpoints/TGI only)
- `gpt2` - Requires custom endpoint
- `facebook/bart-large` - Requires custom endpoint
- `bigscience/bloom-560m` - Requires custom endpoint

## Troubleshooting

### API Deprecation Notice
If you see **410 Gone** errors:
- The old `api-inference.huggingface.co` endpoint is deprecated
- Update to use the new router at `router.huggingface.co`
- This should be automatic with the latest version

### Model Not Found (400/404 Errors)
If you get "model does not exist" errors:
- The new router uses a different model catalog
- Use router-compatible models like `meta-llama/Llama-3.2-1B-Instruct`
- Old HF Hub models (like `gpt2`) are not available on the new router
- For legacy models, use Inference Endpoints or TGI with custom deployment

### Model Loading Errors
If you get "503 Model is loading" errors:
- Set `HFWaitForModel: true` to wait for model to load
- Use the `waitForModel` parameter in your requests
- Consider using Inference Endpoints for production

### Embeddings Not Working
- Embeddings API has a different endpoint structure
- For production embeddings, use dedicated Inference Endpoints
- Consider alternative embedding services

### Rate Limiting
Free tier has rate limits. Consider:
- Using your own Inference Endpoints
- Deploying TGI locally
- Upgrading to a paid plan

### Authentication Errors
- Verify your API key is correct
- Check that `HUGGINGFACE_API_KEY` environment variable is set
- Ensure your API key has appropriate permissions

## Additional Resources

- [Hugging Face Inference Providers Documentation](https://huggingface.co/docs/inference-providers/index) - **New Router API**
- [Hugging Face API Documentation](https://huggingface.co/docs/api-inference) - Legacy docs
- [Text Generation Inference](https://github.com/huggingface/text-generation-inference)
- [Inference Endpoints](https://huggingface.co/inference-endpoints)
- [Model Hub](https://huggingface.co/models)

## See Also

- Main project documentation: `../../docs/`
- Design document: `../../docs/design/HuggingFaceIntegration.md`
- Other LLM examples: `../openrouter_usage/`, `../ollama_smoke_test/`
