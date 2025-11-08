# HuggingFace QuickStart - vNext API

This example demonstrates how to use HuggingFace's LLM services with the AgenticGoKit vNext API.

## Overview

The vNext API provides a modern, streamlined way to create and use AI agents with HuggingFace models. This example showcases:

- ✅ Basic agent creation with the new router API
- ✅ Using different Llama models (1B and 3B variants)
- ✅ Streaming responses for real-time output
- ✅ Conversational agents with context
- ✅ Temperature comparison for creativity control
- ✅ Custom Inference Endpoints
- ✅ Detailed results with RunOptions
- ✅ Custom handlers for advanced logic
- ✅ Multiple queries with reusable agents

## Prerequisites

### 1. HuggingFace API Key

Get your API key from [HuggingFace Settings](https://huggingface.co/settings/tokens):

```bash
export HUGGINGFACE_API_KEY="your-api-key-here"
```

### 2. Optional: Custom Inference Endpoint

For the dedicated endpoint example (Example 6), set:

```bash
export HUGGINGFACE_ENDPOINT_URL="https://your-endpoint.endpoints.huggingface.cloud"
```

## Running the Example

### Full Example (main.go)
Run all 9 examples demonstrating different features:

```bash
cd examples/vnext/huggingface-quickstart
go run main.go
```

### Simple Example (simple.go)
Run a minimal example with just the basics:

```bash
cd examples/vnext/huggingface-quickstart
go run simple.go
```

The simple example shows the minimum code needed to use HuggingFace with vNext API.

## What This Example Demonstrates

### Example 1: Basic Agent with New Router API

Shows how to create a simple agent using the new HuggingFace router (`router.huggingface.co`) with OpenAI-compatible format:

```go
config := &vnext.Config{
    Name:         "hf-assistant",
    SystemPrompt: "You are a helpful assistant.",
    Timeout:      30 * time.Second,
    LLM: vnext.LLMConfig{
        Provider:    "huggingface",
        Model:       "meta-llama/Llama-3.2-1B-Instruct",
        APIKey:      apiKey,
        Temperature: 0.7,
        MaxTokens:   500,
    },
}

agent, _ := vnext.NewBuilder("hf-assistant").
    WithConfig(config).
    Build()
```

**Key Points:**
- Uses router-compatible models (Llama-3.2 series)
- New router API is OpenAI-compatible
- Automatic endpoint management

### Example 2: Using Larger Models

Demonstrates using the 3B parameter Llama model for better quality responses:

```go
Model: "meta-llama/Llama-3.2-3B-Instruct"
```

**Trade-offs:**
- **1B Model**: Faster, lower resource usage
- **3B Model**: Better quality, more accurate responses

### Example 3: Streaming Responses

Real-time token-by-token streaming for responsive applications:

```go
stream, _ := agent.RunStream(ctx, "Write a haiku about programming.")

for chunk := range stream.Chunks() {
    if chunk.Type == vnext.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
}
```

**Benefits:**
- Better user experience
- Lower perceived latency
- Ability to process tokens as they arrive

### Example 4: Conversational Agent

Demonstrates maintaining context across multiple turns:

```go
conversation := []string{
    "My favorite color is blue.",
    "What's my favorite color?",
}

for _, msg := range conversation {
    result, _ := agent.Run(ctx, msg)
    fmt.Println(result.Content)
}
```

**Use Cases:**
- Chat applications
- Interactive assistants
- Context-aware responses

### Example 5: Temperature Comparison

Shows how temperature affects response creativity:

```go
temperatures := []float32{0.3, 0.7, 1.0}
```

**Temperature Guide:**
- **0.0-0.3**: Factual, deterministic (good for Q&A, documentation)
- **0.5-0.7**: Balanced (general purpose)
- **0.8-1.0**: Creative, varied (good for brainstorming, creative writing)

### Example 6: Custom Inference Endpoints

Using dedicated HuggingFace Inference Endpoints for production:

```go
LLM: vnext.LLMConfig{
    Provider: "huggingface",
    Model:    "custom-model",
    APIKey:   apiKey,
    BaseURL:  endpointURL,  // Your dedicated endpoint
}
```

**Benefits:**
- Guaranteed uptime and availability
- Custom model hosting
- Predictable performance
- No cold starts

### Example 7: Detailed Results

Using RunOptions for comprehensive execution information:

```go
opts := vnext.RunWithDetailedResult().
    SetTimeout(30*time.Second).
    AddContext("request_id", "hf-demo-123")

result, _ := agent.RunWithOptions(ctx, "What is deep learning?", opts)
```

**Provides:**
- Execution duration
- Token usage statistics
- Success/failure status
- Custom metadata

### Example 8: Custom Handler

Implementing custom logic before LLM processing:

```go
customHandler := func(ctx context.Context, input string, caps *vnext.Capabilities) (string, error) {
    if containsWord(input, "how") || containsWord(input, "what") {
        enhancedPrompt := fmt.Sprintf("Please provide a clear answer to: %s", input)
        return caps.LLM("You are a technical expert.", enhancedPrompt)
    }
    return caps.LLM("You are a helpful assistant.", input)
}

agent, _ := vnext.NewBuilder("custom-agent").
    WithConfig(config).
    WithHandler(customHandler).
    Build()
```

**Use Cases:**
- Input validation
- Prompt enhancement
- Routing logic
- Pre/post-processing

### Example 9: Multiple Queries

Efficiently reusing a single agent for multiple queries:

```go
queries := []string{
    "What is REST API?",
    "What is GraphQL?",
    "Difference between SQL and NoSQL?",
}

for _, query := range queries {
    result, _ := agent.Run(ctx, query)
    fmt.Println(result.Content)
}
```

**Benefits:**
- Better resource utilization
- Consistent configuration
- Faster than creating new agents

## Available Models on New Router

### Llama 3.2 Series (Recommended)
- `meta-llama/Llama-3.2-1B-Instruct` - Fast, efficient for most tasks
- `meta-llama/Llama-3.2-3B-Instruct` - Better quality, slightly slower

### Provider Routing
- Add `:fastest` suffix: `meta-llama/Llama-3.2-1B-Instruct:fastest`
- Add `:cheapest` suffix: `meta-llama/Llama-3.2-1B-Instruct:cheapest`

### Other Models
- `deepseek-ai/DeepSeek-R1` - Advanced reasoning capabilities
- See [HuggingFace Inference Providers](https://huggingface.co/docs/inference-providers/supported-models) for full list

## Configuration Options

### Basic Configuration
```go
config := &vnext.Config{
    Name:         "my-agent",
    SystemPrompt: "You are a helpful assistant.",
    Timeout:      30 * time.Second,
    LLM: vnext.LLMConfig{
        Provider:    "huggingface",
        Model:       "meta-llama/Llama-3.2-1B-Instruct",
        APIKey:      "your-api-key",
        Temperature: 0.7,
        MaxTokens:   500,
    },
}
```

### Advanced Configuration
```go
config := &vnext.Config{
    Name:         "advanced-agent",
    SystemPrompt: "You are a specialized assistant.",
    Timeout:      60 * time.Second,
    LLM: vnext.LLMConfig{
        Provider:    "huggingface",
        Model:       "meta-llama/Llama-3.2-3B-Instruct",
        APIKey:      apiKey,
        BaseURL:     "https://custom-endpoint.com",  // Optional
        Temperature: 0.8,
        MaxTokens:   1000,
        TopP:        0.9,
    },
}
```

## vNext API Benefits

### 1. Simplified Agent Creation
```go
agent, _ := vnext.NewBuilder("my-agent").
    WithConfig(config).
    Build()
```

### 2. Cleaner Error Handling
```go
result, err := agent.Run(ctx, "Your query")
if err != nil {
    log.Printf("Error: %v", err)
}
```

### 3. Built-in Streaming Support
```go
stream, _ := agent.RunStream(ctx, "Your query")
for chunk := range stream.Chunks() {
    // Process chunks
}
```

### 4. Flexible Configuration
```go
agent, _ := vnext.NewBuilder("agent").
    WithConfig(config).
    WithHandler(customHandler).
    WithTimeout(60 * time.Second).
    Build()
```

## Troubleshooting

### Error: "410 Gone"
The old `api-inference.huggingface.co` endpoint is deprecated. Update to the latest AgenticGoKit version which uses the new router automatically.

### Error: "Model does not exist"
Use router-compatible models like `meta-llama/Llama-3.2-1B-Instruct`. Old HF Hub models (like `gpt2`) are not available on the new router.

### Slow Responses
- Try using the 1B model instead of 3B
- Add `:fastest` suffix to model name
- Consider using dedicated Inference Endpoints

### API Key Issues
```bash
# Check if key is set
echo $HUGGINGFACE_API_KEY

# Set the key
export HUGGINGFACE_API_KEY="your-key-here"
```

## Performance Tips

1. **Model Selection**
   - Use 1B for speed
   - Use 3B for quality
   - Use dedicated endpoints for production

2. **Temperature Tuning**
   - Lower (0.3) for factual content
   - Medium (0.7) for general use
   - Higher (0.9) for creativity

3. **Token Management**
   - Set appropriate MaxTokens to control cost and latency
   - Monitor TokensUsed in results

4. **Agent Reuse**
   - Create once, use multiple times
   - More efficient than creating new agents

## Related Examples

- **Basic HuggingFace**: `examples/huggingface_usage/` - Legacy API examples
- **OpenRouter vNext**: `examples/vnext/openrouter-quickstart/` - Similar patterns with OpenRouter
- **Ollama vNext**: `examples/vnext/ollama-quickstart/` - Local model examples
- **Streaming Demo**: `examples/vnext/streaming-demo/` - Advanced streaming patterns

## Additional Resources

- [HuggingFace Inference Providers Docs](https://huggingface.co/docs/inference-providers/index)
- [vNext API Documentation](../../../core/vnext/README.md)
- [Migration Notes](../../huggingface_usage/MIGRATION_NOTES.md)
- [Model Hub](https://huggingface.co/models)

## Questions?

If you encounter issues:
1. Verify your API key is valid
2. Check you're using router-compatible models
3. Ensure latest AgenticGoKit version
4. Review the [Migration Notes](../../huggingface_usage/MIGRATION_NOTES.md)
