# HuggingFace API Migration Notes

## Overview

As of **late 2024**, HuggingFace has migrated from the legacy `api-inference.huggingface.co` endpoint to a new router-based architecture at `router.huggingface.co`. This document explains the changes and how to migrate your code.

## What Changed?

### Old API (Deprecated)
- **Endpoint**: `https://api-inference.huggingface.co/models/{model}`
- **Format**: Custom inference format
- **Status**: Returns **410 Gone** (permanently removed)
- **Models**: Any HuggingFace Hub model (e.g., `gpt2`, `facebook/bart-large`)

### New API (Current)
- **Endpoint**: `https://router.huggingface.co/v1/chat/completions`
- **Format**: OpenAI-compatible chat completions
- **Status**: Active and maintained
- **Models**: Provider-based catalog (e.g., `meta-llama/Llama-3.2-1B-Instruct`, `deepseek-ai/DeepSeek-R1`)

## Migration Steps

### 1. Update Your Models

**Before** (Old API):
```go
config := core.LLMProviderConfig{
    Type:      "huggingface",
    APIKey:    apiKey,
    Model:     "gpt2", // ❌ Not available on new router
    HFAPIType: "inference",
}
```

**After** (New API):
```go
config := core.LLMProviderConfig{
    Type:      "huggingface",
    APIKey:    apiKey,
    Model:     "meta-llama/Llama-3.2-1B-Instruct", // ✅ Router-compatible
    HFAPIType: "inference",
}
```

### 2. No Code Changes Required for Basic Usage

The AgenticGoKit library handles the API migration automatically. If you're using the latest version, your code will automatically use the new router API.

### 3. Update Environment Variables (If Any)

No changes needed - `HUGGINGFACE_API_KEY` remains the same.

## Available Models on New Router

### Small Models (Fast & Efficient)
- `meta-llama/Llama-3.2-1B-Instruct` - 1B parameters, instruction-tuned
- `meta-llama/Llama-3.2-3B-Instruct` - 3B parameters, better quality

### Larger Models
- `deepseek-ai/DeepSeek-R1` - Advanced reasoning model
- More models available at [HuggingFace Inference Providers](https://huggingface.co/docs/inference-providers/index)

### Provider Routing
- Add `:fastest` suffix for fastest provider: `meta-llama/Llama-3.2-1B-Instruct:fastest`
- Add `:cheapest` suffix for cheapest provider: `meta-llama/Llama-3.2-1B-Instruct:cheapest`

## What If I Need Legacy Models?

If you need to use legacy HuggingFace Hub models like `gpt2`, you have two options:

### Option 1: Use Inference Endpoints
Deploy your model to a dedicated Inference Endpoint:

```go
config := core.LLMProviderConfig{
    Type:      "huggingface",
    APIKey:    apiKey,
    Model:     "gpt2",
    BaseURL:   "https://your-endpoint.endpoints.huggingface.cloud",
    HFAPIType: "endpoint",
}
```

### Option 2: Use Text Generation Inference (TGI)
Run a local TGI server:

```go
config := core.LLMProviderConfig{
    Type:      "huggingface",
    Model:     "gpt2",
    BaseURL:   "http://localhost:8080",
    HFAPIType: "tgi",
}
```

## API Format Differences

### Old Format (Custom Inference)
```json
{
  "inputs": "Your prompt here",
  "parameters": {
    "max_new_tokens": 100,
    "temperature": 0.7
  }
}
```

### New Format (OpenAI-Compatible)
```json
{
  "model": "meta-llama/Llama-3.2-1B-Instruct",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Your prompt here"}
  ],
  "max_tokens": 100,
  "temperature": 0.7,
  "stream": false
}
```

**Note**: AgenticGoKit handles this conversion automatically.

## Error Messages

If you encounter these errors, you need to migrate:

### 410 Gone
```
Hugging Face API error (410): https://api-inference.huggingface.co is no longer supported.
Please use https://router.huggingface.co/hf-inference instead.
```
**Solution**: Update to latest AgenticGoKit version and use router-compatible models.

### 400 Bad Request - Model Not Found
```
Hugging Face API error (400): The requested model 'gpt2' does not exist
```
**Solution**: Use a router-compatible model like `meta-llama/Llama-3.2-1B-Instruct`.

### 404 Not Found
```
Hugging Face API error (404): Not Found
```
**Solution**: Check that you're using the correct API type and model name.

## Testing Your Migration

Run the example to verify everything works:

```bash
export HUGGINGFACE_API_KEY="your-key-here"
cd examples/huggingface_usage
go run main.go
```

You should see successful responses from the new router API.

## Embeddings Note

The embeddings API has a different endpoint structure and is not yet fully migrated. For production embeddings, consider:

1. Using HuggingFace Inference Endpoints
2. Using dedicated embedding services
3. Deploying your own embedding models with TGI

## Resources

- [New Router Documentation](https://huggingface.co/docs/inference-providers/index)
- [AgenticGoKit HuggingFace Integration](./README.md)
- [Available Models](https://huggingface.co/docs/inference-providers/supported-models)

## Questions?

If you encounter issues during migration, check:
1. You're using the latest AgenticGoKit version
2. Your API key is valid
3. You're using router-compatible models
4. Your API key has appropriate permissions

For more help, see the main documentation or open an issue on GitHub.
