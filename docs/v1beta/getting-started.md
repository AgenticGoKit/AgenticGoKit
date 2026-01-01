# Getting Started

Welcome to AgenticGoKit! You'll build and run your first AI agent in just 5 minutes.

AgenticGoKit is a production-ready framework for building AI agents in Go. An **agent** is a program that takes user input, thinks about it (optionally with tools and memory), and returns a response from an LLM.

---

## Installation

```bash
# Install the library
go get github.com/agenticgokit/agenticgokit/v1beta

# Set your LLM provider API key
export OPENAI_API_KEY="sk-..."          # OpenAI
export OLLAMA_HOST="http://localhost:11434"  # Ollama (local)
```

For Ollama, install from [ollama.com](https://ollama.com) and pull a model:
```bash
ollama pull llama2
```

---

## Your First Agent

### Step 1: Create Your Agent

Create a file called `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Set your LLM API key (or use environment variables)
    os.Setenv("OPENAI_API_KEY", "your-api-key-here")

    // Create an agent
    agent, err := v1beta.NewChatAgent("Assistant",
        v1beta.WithLLM("openai", "gpt-4"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Run the agent
    result, err := agent.Run(context.Background(), "What is Go?")
    if err != nil {
        log.Fatal(err)
    }

    // Print the result
    fmt.Println("Response:", result.Content)
    fmt.Println("Success:", result.Success)
}
```

### Step 2: Run It

```bash
go run main.go
```

**Output:**

```
Response: Go is a statically typed, compiled programming language developed at Google...
Success: true
```

**Congratulations!** You just built your first AgenticGoKit agent! 🎉

---

## 🎯 What Just Happened?

### Agent
An **Agent** is your interface to an LLM. You create it once with configuration (which model, what to remember, which tools to use), then call it repeatedly.

```go
agent, err := v1beta.NewChatAgent("Assistant", v1beta.WithLLM("openai", "gpt-4"))
// Result: An Agent that uses OpenAI's GPT-4 model
```

### Run
**Run** executes the agent with user input and waits for the complete response.

```go
result, err := agent.Run(context.Background(), "Your question here")
// Returns: Complete response + metadata (tokens used, execution time, etc.)
```

### Result
The **Result** contains:
- `Content` - The LLM's response text
- `Success` - Whether execution succeeded
- `Duration` - How long it took
- `TokensUsed` - LLM tokens consumed
- `Memory` - Whether memory was used
- Other metadata

```go
if result.Success {
    fmt.Println(result.Content)      // The response
    fmt.Println(result.TokensUsed)   // Cost indicator
}
```

### Memory
**By default, agents remember the conversation.** Each agent automatically stores interactions using an embedded memory provider (`chromem`). Call the same agent multiple times and it remembers what you said.

```go
result1, _ := agent.Run(ctx, "My name is Alice")
result2, _ := agent.Run(ctx, "What is my name?")
// Result2 will answer: "Your name is Alice" ← From memory!
```

If you prefer stateless agents (no memory), you can disable it—see [Memory & RAG](./memory-and-rag.md).

---

## 🔧 Common Customizations

### Use a Different Model

```go
// Use GPT-3.5-turbo instead
agent, err := v1beta.NewChatAgent("Assistant",
    v1beta.WithLLM("openai", "gpt-3.5-turbo"),
)

// Or use Ollama locally
agent, err := v1beta.NewChatAgent("Assistant",
    v1beta.WithLLM("ollama", "llama2"),
)
```

### Add a Custom System Prompt

```go
agent, err := v1beta.NewBuilder("Assistant").
    WithPreset(v1beta.ChatAgent).
    WithConfig(&v1beta.Config{
        SystemPrompt: "You are a friendly pirate",
    }).
    Build()

result, _ := agent.Run(ctx, "Hello!")
// Response: "Ahoy, matey! What be bringin' ye to these waters?"
```

### Set a Timeout

```go
agent, err := v1beta.NewChatAgent("Assistant",
    v1beta.WithLLM("openai", "gpt-4"),
    v1beta.WithAgentTimeout(15 * time.Second), // Max 15 seconds
)
```

**Want more control?** See the [Configuration Guide](./configuration.md) for all builder options, presets, and advanced settings.

---

## ⚡ What's Next?

The agent you just built is functional but basic. Here's what you can add depending on your needs:

### 🌊 Real-Time Streaming
Display responses token-by-token as they're generated (instead of waiting for the complete response).

**→ [Streaming Guide](./streaming.md)**

### 🔄 Multi-Agent Workflows
Build pipelines where multiple agents work together (one researches, another writes, another reviews).

**→ [Workflows Guide](./workflows.md)**

### 🛠️ Tool Integration
Give agents the ability to call external APIs, databases, file systems, etc. via the Model Context Protocol (MCP).

**→ [Tool Integration](./tool-integration.md)**

### 💾 Memory & RAG
Enable agents to store and retrieve custom knowledge, documents, and long-term facts.

**→ [Memory & RAG](./memory-and-rag.md)**

### 🎨 Custom Logic
Write handlers to control exactly how agents process input (bypass LLM for certain queries, apply custom rules, etc.).

**→ [Custom Handlers](./custom-handlers.md)**

### 📊 Deeper Configuration
Fine-tune temperature, max tokens, timeouts, caching, and more.

**→ [Configuration Guide](./configuration.md)**

---

## 📚 Suggested Learning Path

We recommend exploring in this order:

1. **✅ You are here**: Getting Started (5 min) - Build and run an agent
2. **[Core Concepts](./core-concepts.md)** (15 min) - Understand agents, handlers, tools, and memory
3. **Pick your path:**
   - Want real-time responses? → [Streaming Guide](./streaming.md)
   - Multiple agents? → [Workflows Guide](./workflows.md)
   - External APIs? → [Tool Integration](./tool-integration.md)
   - Knowledge base? → [Memory & RAG](./memory-and-rag.md)
   - Custom behavior? → [Custom Handlers](./custom-handlers.md)
4. **Explore examples** - See complete projects in [examples/](../examples/)
5. **Troubleshoot** - Visit [Troubleshooting](./troubleshooting.md) if something breaks

---

## 🆘 Need Help?

- **Problems?** → [Troubleshooting Guide](./troubleshooting.md)
- **See it in action** → [Examples & Tutorials](../examples/)
- **Questions?** → [GitHub Discussions](https://github.com/agenticgokit/agenticgokit/discussions)
- **Found a bug?** → [GitHub Issues](https://github.com/agenticgokit/agenticgokit/issues)

---

**Ready to learn more?** Continue to [Core Concepts](./core-concepts.md) →
