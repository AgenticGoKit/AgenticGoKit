# 🤖 Simple Agent Example

**Your first AgenticGoKit agent - a helpful AI assistant that responds to any question.**

## What This Demonstrates

- **Basic Agent Creation**: How to create and configure a single LLM agent
- **Provider Integration**: Using OpenAI (or any LLM provider) with AgenticGoKit
- **Simple Interaction**: Processing messages and getting responses
- **Error Handling**: Graceful handling of API errors and timeouts

## Quick Start

### 1. Setup
```bash
cd examples/01-simple-agent
cp .env.example .env
# Edit .env with your OpenAI API key
```

### 2. Run
```bash
go run main.go "What is the capital of France?"
```

### 3. Expected Output
```
🤖 Simple Agent Starting...
📝 Processing: "What is the capital of France?"

✅ Agent Response:
The capital of France is Paris. Paris is not only the political capital but also the cultural and economic center of France, known for its iconic landmarks like the Eiffel Tower, Louvre Museum, and Notre-Dame Cathedral.

📊 Stats:
   • Response time: 1.2s
   • Tokens used: ~45
   • Success: true
```

## Code Walkthrough

### main.go - The Complete Agent

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // 1. Get the user's question from command line
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go \"Your question here\"")
        os.Exit(1)
    }
    question := os.Args[1]
    
    // 2. Create an LLM provider (OpenAI in this case)
    provider := core.OpenAIProvider{
        APIKey:      os.Getenv("OPENAI_API_KEY"),
        Model:       "gpt-3.5-turbo",
        Temperature: 0.7, // Balanced creativity
        MaxTokens:   500,  // Reasonable response length
    }
    
    // 3. Create a simple agent with a helpful personality
    agent := core.NewLLMAgent("helpful-assistant", provider).
        WithSystemPrompt(`You are a helpful, knowledgeable assistant. 
        Provide clear, accurate, and concise answers to questions. 
        If you're not sure about something, say so honestly.`)
    
    // 4. Process the question
    fmt.Printf("🤖 Simple Agent Starting...\n")
    fmt.Printf("📝 Processing: \"%s\"\n\n", question)
    
    startTime := time.Now()
    
    // Create input state with the user's question
    inputState := core.NewState().Set("message", question)
    
    // Run the agent
    result, err := agent.Run(context.Background(), inputState)
    
    duration := time.Since(startTime)
    
    // 5. Handle the response
    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)
        os.Exit(1)
    }
    
    // Extract the response from the result state
    response, ok := result.Get("response")
    if !ok {
        fmt.Println("❌ No response received from agent")
        os.Exit(1)
    }
    
    // 6. Display the results
    fmt.Printf("✅ Agent Response:\n")
    fmt.Printf("%s\n\n", response)
    
    fmt.Printf("📊 Stats:\n")
    fmt.Printf("   • Response time: %v\n", duration)
    fmt.Printf("   • Success: true\n")
}
```

### Key Concepts Explained

#### 1. **LLM Provider Setup**
```go
provider := core.OpenAIProvider{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Model:       "gpt-3.5-turbo",
    Temperature: 0.7,
    MaxTokens:   500,
}
```
- **APIKey**: Your OpenAI API key from environment variables
- **Model**: Which LLM model to use (gpt-3.5-turbo is fast and cost-effective)
- **Temperature**: Controls creativity (0.0 = deterministic, 1.0 = very creative)
- **MaxTokens**: Maximum response length

#### 2. **Agent Creation**
```go
agent := core.NewLLMAgent("helpful-assistant", provider).
    WithSystemPrompt("You are a helpful assistant...")
```
- **Name**: Unique identifier for the agent
- **Provider**: The LLM provider to use
- **SystemPrompt**: Instructions that define the agent's personality and behavior

#### 3. **State Management**
```go
inputState := core.NewState().Set("message", question)
result, err := agent.Run(context.Background(), inputState)
response := result.Get("response")
```
- **State**: Key-value store for passing data between agents
- **Input State**: Contains the user's message
- **Output State**: Contains the agent's response

## Configuration Options

### Environment Variables (.env)
```bash
# Required
OPENAI_API_KEY=sk-your-openai-api-key-here

# Optional - Model Configuration
OPENAI_MODEL=gpt-3.5-turbo
OPENAI_TEMPERATURE=0.7
OPENAI_MAX_TOKENS=500
```

### Alternative LLM Providers

#### Azure OpenAI
```go
provider := core.AzureOpenAIProvider{
    APIKey:   os.Getenv("AZURE_OPENAI_API_KEY"),
    Endpoint: os.Getenv("AZURE_OPENAI_ENDPOINT"),
    Model:    "gpt-35-turbo",
}
```

#### Ollama (Local)
```go
provider := core.OllamaProvider{
    BaseURL: "http://localhost:11434",
    Model:   "llama2",
}
```

## Customization Examples

### 1. Specialized Agent Personalities

#### Code Assistant
```go
agent := core.NewLLMAgent("code-assistant", provider).
    WithSystemPrompt(`You are a senior software engineer. 
    Provide code examples, explain programming concepts clearly, 
    and suggest best practices. Always include working code when relevant.`)
```

#### Creative Writer
```go
agent := core.NewLLMAgent("creative-writer", provider).
    WithSystemPrompt(`You are a creative writing assistant. 
    Help with storytelling, character development, and creative ideas. 
    Be imaginative and inspiring while maintaining quality.`)
```

#### Data Analyst
```go
agent := core.NewLLMAgent("data-analyst", provider).
    WithSystemPrompt(`You are a data analyst. 
    Help interpret data, suggest analysis approaches, and explain 
    statistical concepts clearly. Focus on actionable insights.`)
```

### 2. Interactive Mode

Create an interactive chat session:

```go
// interactive.go
func main() {
    agent := createAgent()
    scanner := bufio.NewScanner(os.Stdin)
    
    fmt.Println("🤖 Interactive Agent (type 'quit' to exit)")
    
    for {
        fmt.Print("\n💬 You: ")
        if !scanner.Scan() {
            break
        }
        
        input := strings.TrimSpace(scanner.Text())
        if input == "quit" {
            break
        }
        
        response := processMessage(agent, input)
        fmt.Printf("🤖 Agent: %s\n", response)
    }
}
```

### 3. Batch Processing

Process multiple questions at once:

```go
// batch.go
func main() {
    questions := []string{
        "What is machine learning?",
        "How does blockchain work?",
        "Explain quantum computing",
    }
    
    agent := createAgent()
    
    for i, question := range questions {
        fmt.Printf("\n📝 Question %d: %s\n", i+1, question)
        response := processMessage(agent, question)
        fmt.Printf("✅ Answer: %s\n", response)
    }
}
```

## Error Handling

### Common Issues and Solutions

#### API Key Problems
```go
if os.Getenv("OPENAI_API_KEY") == "" {
    fmt.Println("❌ OPENAI_API_KEY environment variable not set")
    fmt.Println("💡 Get your API key from: https://platform.openai.com/api-keys")
    os.Exit(1)
}
```

#### Rate Limiting
```go
// Add retry logic for rate limits
agent := core.NewLLMAgent("assistant", provider).
    WithRetryPolicy(core.RetryPolicy{
        MaxRetries:    3,
        BackoffFactor: 2.0,
        MaxDelay:      30 * time.Second,
    })
```

#### Timeout Handling
```go
// Set context timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := agent.Run(ctx, inputState)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("❌ Request timed out")
    } else {
        fmt.Printf("❌ Error: %v\n", err)
    }
}
```

## Performance Tips

### 1. Model Selection
```go
// For speed and cost-effectiveness
provider.Model = "gpt-3.5-turbo"

// For higher quality responses
provider.Model = "gpt-4"

// For longer contexts
provider.Model = "gpt-3.5-turbo-16k"
```

### 2. Response Optimization
```go
// Shorter responses for faster processing
provider.MaxTokens = 150

// Lower temperature for more consistent responses
provider.Temperature = 0.3

// Use streaming for real-time responses (if supported)
provider.Stream = true
```

### 3. Caching Responses
```go
// Add simple response caching
cache := make(map[string]string)

func processWithCache(agent core.Agent, question string) string {
    if cached, exists := cache[question]; exists {
        return cached
    }
    
    response := processMessage(agent, question)
    cache[question] = response
    return response
}
```

## Next Steps

Now that you understand basic agents, explore more advanced patterns:

- **[🤝 Multi-Agent Collaboration](../02-multi-agent-collab/)** - Multiple agents working together
- **[🔄 Sequential Processing](../03-sequential-pipeline/)** - Step-by-step data processing
- **[🧠 Memory & RAG](../04-rag-knowledge-base/)** - Persistent memory and knowledge
- **[🏭 Production System](../05-production-system/)** - Full production deployment

## Troubleshooting

### Debug Mode
```bash
# Enable debug logging
export AGENTICGOKIT_LOG_LEVEL=debug
go run main.go "Your question"
```

### Test Your Setup
```bash
# Quick connectivity test
go run -c 'package main; import "fmt"; func main() { fmt.Println("Go setup OK") }'

# Test API key
curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
```

---

**💡 Pro Tip**: This simple agent is the foundation for all AgenticGoKit applications. Master these basics, and you'll be ready to build complex multi-agent systems!