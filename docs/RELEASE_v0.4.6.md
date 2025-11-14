---
title: "AgenticGoKit ❤️ HuggingFace and OpenRouter"
date: 2025-11-08
description: "AgenticGoKit v0.4.6 introduces Hugging Face and OpenRouter support with the new router API, enabling Go developers to build agentic workflows with a wide variety of LLMs including Llama, Claude, and GPT models."
tags: ["Go", "AI", "LLMs", "Agents", "HuggingFace", "OpenRouter", "Release", "Llama"]
author: "AgenticGoKit Team"
draft: false
---

# AgenticGoKit ❤️ HuggingFace and OpenRouter

We're excited to announce the release of **AgenticGoKit v0.4.6** — a major update that expands how Go developers can build and experiment with agentic workflows.

With this release, **AgenticGoKit** now supports:

- **[Hugging Face](https://huggingface.co/)** — Access models through the new router API with OpenAI-compatible format
- **[OpenRouter](https://openrouter.ai/)** — Use a wide variety of LLMs (GPT, Claude, Gemini, Llama) through a unified API

These integrations empower developers to build richer, more flexible **AI agents** without switching languages or managing complex SDKs.

---

## What's New

### Hugging Face Integration with New Router API

AgenticGoKit now supports **Hugging Face's latest router-based architecture** (`router.huggingface.co`), which provides:

- **OpenAI-compatible format** — Seamless migration from other providers
- **Multiple API types** — Inference API, Chat API, dedicated Endpoints, and self-hosted TGI
- **Llama 3.2 models** — Access to the latest open-source models
- **Streaming support** — Real-time token-by-token responses
- **Advanced parameters** — Fine-tune with temperature, top-p, top-k, and more

#### Quick Example

```go
import (
    "context"
    "github.com/kunalkushwaha/agenticgokit/core/vnext"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/huggingface"
)

config := &vnext.Config{
    Name:         "hf-assistant",
    SystemPrompt: "You are a helpful AI assistant.",
    LLM: vnext.LLMConfig{
        Provider:    "huggingface",
        Model:       "meta-llama/Llama-3.2-1B-Instruct",
        APIKey:      os.Getenv("HUGGINGFACE_API_KEY"),
        Temperature: 0.7,
        MaxTokens:   500,
    },
}

agent, _ := vnext.NewBuilder("hf-assistant").WithConfig(config).Build()
result, _ := agent.Run(context.Background(), "Explain machine learning")
fmt.Println(result.Content)
```

**Available Models:**
- `meta-llama/Llama-3.2-1B-Instruct` — Fast, efficient
- `meta-llama/Llama-3.2-3B-Instruct` — Better quality
- `deepseek-ai/DeepSeek-R1` — Advanced reasoning
- Model routing with `:fastest` and `:cheapest` suffixes

[HuggingFace Examples](https://github.com/AgenticGoKit/AgenticGoKit/tree/master/examples/huggingface_usage) | [HuggingFace vNext Quickstart](https://github.com/AgenticGoKit/AgenticGoKit/tree/master/examples/vnext/huggingface-quickstart)

---

### OpenRouter Integration

OpenRouter provides a single interface to **dozens of leading LLMs** from OpenAI, Anthropic, Google, Meta, and more. With AgenticGoKit v0.4.6, you can:

- Access GPT-4, Claude 3, Gemini 2.0, Llama 3.1, and many others
- Compare models side-by-side for your use case
- Switch between providers with a single line of code
- Track usage with site analytics

#### Quick Example

```go
config := &vnext.Config{
    Name:         "openrouter-agent",
    SystemPrompt: "You are a helpful assistant.",
    LLM: vnext.LLMConfig{
        Provider:    "openrouter",
        Model:       "anthropic/claude-3-haiku",
        APIKey:      os.Getenv("OPENROUTER_API_KEY"),
        Temperature: 0.7,
        MaxTokens:   500,
    },
}

agent, _ := vnext.NewBuilder("openrouter-agent").WithConfig(config).Build()
result, _ := agent.Run(ctx, "What's the difference between REST and GraphQL?")
```

**Popular Models Available:**
- `openai/gpt-4-turbo` — Most capable GPT model
- `anthropic/claude-3-haiku` — Fast, cost-effective
- `google/gemini-2.0-flash-exp:free` — Free Google model
- `meta-llama/llama-3.1-8b-instruct` — Open-source Llama

[OpenRouter Examples](https://github.com/AgenticGoKit/AgenticGoKit/tree/master/examples/openrouter_usage) | [OpenRouter vNext Quickstart](https://github.com/AgenticGoKit/AgenticGoKit/tree/master/examples/vnext/openrouter-quickstart)

---

## Why This Matters

Building **agentic systems** in Go is now easier and more powerful than ever. By connecting to Hugging Face and OpenRouter, AgenticGoKit developers can:

- **Rapidly prototype** with different LLMs and compare performance
- **Optimize costs** by testing various model/provider combinations
- **Build multi-agent systems** using diverse model capabilities
- **Access the latest models** including Llama 3.2, Claude 3, Gemini 2.0
- **Keep the speed and reliability** of Go with type safety
- **Switch providers easily** without rewriting application logic

---

## Key Features in v0.4.6

### Multiple API Types for HuggingFace

1. **Inference API** — New router with OpenAI-compatible format
2. **Chat API** — Conversational models with legacy support
3. **Inference Endpoints** — Dedicated hosting for production
4. **Text Generation Inference (TGI)** — Self-hosted optimized inference

### Streaming Support

Both HuggingFace and OpenRouter now support real-time streaming:

```go
stream, _ := agent.RunStream(ctx, "Write a story about AI")
for chunk := range stream.Chunks() {
    if chunk.Type == vnext.ChunkTypeDelta {
        fmt.Print(chunk.Delta)
    }
}
```

### Advanced Configuration

Fine-tune model behavior with provider-specific parameters:

```go
LLM: vnext.LLMConfig{
    Provider:          "huggingface",
    Model:            "meta-llama/Llama-3.2-1B-Instruct",
    Temperature:      0.8,
    MaxTokens:        1000,
    TopP:             0.9,
    TopK:             50,
    RepetitionPenalty: 1.2,
}
```

### Comprehensive Examples

- **9 HuggingFace examples** covering all API types
- **8 OpenRouter examples** demonstrating various models
- **Simple quickstart examples** for both providers
- **Migration guides** for the new HuggingFace router API
- **Complete documentation** with troubleshooting tips

---

## The Story: Breaking Free from LLM Lock-In

### The Challenge

You're building an AI agent in Go. You start with OpenAI's GPT-4 because it's the best. But then:

- **Costs add up** — Your prototype's API bills are growing faster than your user base
- **You're locked in** — What if you want to try Anthropic's Claude? Or use open-source Llama models?
- **Experimentation is hard** — Testing different models means rewriting integration code
- **Enterprise needs** — Your client wants Azure OpenAI for compliance, but your code is OpenAI-specific

**You're stuck.** Switching providers means days of refactoring.

### The Solution

**AgenticGoKit v0.4.6 changes the game.** With HuggingFace and OpenRouter support, you can now:

```go
// Start with HuggingFace for experimentation
config.LLM.Provider = "huggingface"
config.LLM.Model = "meta-llama/Llama-3.2-1B-Instruct"

// Switch to OpenRouter to test Claude
config.LLM.Provider = "openrouter"
config.LLM.Model = "anthropic/claude-3-haiku"

// Move to production with OpenAI
config.LLM.Provider = "openai"
config.LLM.Model = "gpt-4-turbo"
```

**Same code. Different providers. One line change.**

### The Impact

With **5 supported providers** (OpenAI, Azure OpenAI, HuggingFace, OpenRouter, Ollama), AgenticGoKit now gives you:

| Provider | What You Get | Why It Matters |
|----------|--------------|----------------|
| **HuggingFace** (NEW) | Open-source models (Llama 3.2) + Router API | Free tier, experimentation, self-hosting options |
| **OpenRouter** (NEW) | 50+ models (GPT, Claude, Gemini, Llama) | Single API, compare models, optimize costs |
| **OpenAI** | GPT-4, GPT-3.5 | Industry standard, highest quality |
| **Azure OpenAI** | Enterprise GPT models | Compliance, SLAs, enterprise support |
| **Ollama** | Local models | Privacy, offline, zero API costs |

**Build once. Deploy anywhere. Switch anytime.**

---

## Take Action: Try It Now

### Step 1: Install (30 seconds)

```bash
go get github.com/kunalkushwaha/agenticgokit@latest
```

### Step 2: Get an API Key (2 minutes)

Pick **one** provider to start:

- **HuggingFace** (free tier): https://huggingface.co/settings/tokens
- **OpenRouter** (flexible): https://openrouter.ai/keys

### Step 3: Run Your First Agent (1 minute)

```bash
# HuggingFace - Free tier, open models
export HUGGINGFACE_API_KEY="hf_..."
cd examples/vnext/huggingface-quickstart
go run simple.go

# OR OpenRouter - 50+ models to choose from
export OPENROUTER_API_KEY="sk-or-..."
cd examples/vnext/openrouter-quickstart
go run main.go
```

**That's it.** You're running AI agents in Go.

---

## What You Can Build Today

### 1. Cost Optimizer Agent
```go
// Start cheap with HuggingFace Llama
config.LLM.Provider = "huggingface"
config.LLM.Model = "meta-llama/Llama-3.2-1B-Instruct"

// Upgrade to GPT-4 for complex queries
if complexQuery {
    config.LLM.Provider = "openai"
    config.LLM.Model = "gpt-4-turbo"
}
```

### 2. Multi-Model Research Assistant
```go
// Use OpenRouter to query multiple models
models := []string{
    "openai/gpt-4-turbo",
    "anthropic/claude-3-haiku",
    "google/gemini-2.0-flash-exp:free",
}

for _, model := range models {
    config.LLM.Model = model
    // Compare responses from different models
}
```

### 3. Privacy-First Local Agent
```go
// Start with Ollama for development
config.LLM.Provider = "ollama"
config.LLM.Model = "llama3.2"

// Switch to HuggingFace TGI for production
config.LLM.Provider = "huggingface"
config.LLM.BaseURL = "http://your-tgi-server:8080"
```

---

## Learn More

**Quick Links:**
- [HuggingFace Quickstart](https://github.com/AgenticGoKit/AgenticGoKit/tree/master/examples/vnext/huggingface-quickstart) — 9 examples
- [OpenRouter Quickstart](https://github.com/AgenticGoKit/AgenticGoKit/tree/master/examples/vnext/openrouter-quickstart) — 8 examples
- [HuggingFace Migration Guide](https://github.com/AgenticGoKit/AgenticGoKit/blob/master/examples/huggingface_usage/MIGRATION_NOTES.md) — New router API
- [All Examples](https://github.com/AgenticGoKit/AgenticGoKit/tree/master/examples) — 20+ working examples
- [GitHub Discussions](https://github.com/AgenticGoKit/AgenticGoKit/discussions) — Get help

---

## Join the Community

AgenticGoKit is open source and growing. We'd love your feedback:

- **Found a bug?** [Open an issue](https://github.com/AgenticGoKit/AgenticGoKit/issues)
- **Have an idea?** [Start a discussion](https://github.com/AgenticGoKit/AgenticGoKit/discussions)
- **Want to contribute?** [Check out our contributor guide](https://github.com/AgenticGoKit/AgenticGoKit/blob/master/docs/contributors/ContributorGuide.md)
- **Like what you see?** [Star us on GitHub](https://github.com/AgenticGoKit/AgenticGoKit)

---

## The Bottom Line

**AgenticGoKit v0.4.6 gives you freedom.**

- Start with **HuggingFace** for free experimentation
- Switch to **OpenRouter** to test 50+ models
- Move to **OpenAI** for production quality
- Deploy to **Azure OpenAI** for enterprise compliance
- Or keep it local with **Ollama**

**One codebase. Five providers. Endless possibilities.**

```bash
# Get started now
go get github.com/kunalkushwaha/agenticgokit@latest
```

**Happy building!**

---

*Questions? Join our [community discussions](https://github.com/AgenticGoKit/AgenticGoKit/discussions) or check out the [documentation](https://github.com/AgenticGoKit/AgenticGoKit/blob/master/docs/README.md).*
