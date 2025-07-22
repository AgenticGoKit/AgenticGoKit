# 🚀 5-Minute Quickstart

Get your first AgenticGoKit multi-agent system running in 5 minutes. No complex setup, no configuration files—just working code.

## What You'll Build

A simple but powerful multi-agent system where:
- 🤖 **Agent 1** processes your request
- 🤖 **Agent 2** enhances the response  
- 🤖 **Agent 3** formats the final output

All working together automatically!

## Prerequisites

- **Go 1.21+** ([install here](https://golang.org/dl/))
- **OpenAI API Key** ([get one here](https://platform.openai.com/api-keys))

*That's it! No Docker, no databases, no complex setup.*

---

## Step 1: Create Your Project (30 seconds)

```bash
mkdir my-agents && cd my-agents
go mod init my-agents
go get github.com/kunalkushwaha/agenticgokit
```

## Step 2: Write Your First Multi-Agent System (2 minutes)

Create `main.go`:

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
    // 🔑 Set up your LLM provider
    provider := core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
        Model:  "gpt-3.5-turbo", // Fast and cost-effective
    }
    
    // 🤖 Create three specialized agents
    agents := map[string]core.AgentHandler{
        "processor": core.NewLLMAgent("processor", provider).
            WithSystemPrompt("You process user requests and extract key information."),
            
        "enhancer": core.NewLLMAgent("enhancer", provider).
            WithSystemPrompt("You enhance and improve responses with additional insights."),
            
        "formatter": core.NewLLMAgent("formatter", provider).
            WithSystemPrompt("You format responses in a clear, professional manner."),
    }
    
    // 🚀 Create a collaborative runner (agents work together)
    runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
    
    // 💬 Process a message - watch the magic happen!
    fmt.Println("🤖 Starting multi-agent collaboration...")
    
    result, err := runner.ProcessMessage(context.Background(), 
        "Explain quantum computing in simple terms")
    
    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)
        return
    }
    
    // 🎉 Show the collaborative result
    fmt.Println("\n✅ Multi-Agent Response:")
    fmt.Println("=" + strings.Repeat("=", 50))
    fmt.Println(result.Content)
    fmt.Println("=" + strings.Repeat("=", 50))
    
    // 📊 Show what happened behind the scenes
    fmt.Printf("\n📊 Execution Stats:\n")
    fmt.Printf("   • Agents involved: %d\n", len(agents))
    fmt.Printf("   • Processing time: %v\n", result.Duration)
    fmt.Printf("   • Success: %t\n", result.Success)
}
```

## Step 3: Run It! (30 seconds)

```bash
export OPENAI_API_KEY=your-api-key-here
go run main.go
```

You should see:
```
🤖 Starting multi-agent collaboration...

✅ Multi-Agent Response:
==================================================
Quantum computing is like having a super-powered computer that can explore multiple solutions simultaneously...

[Enhanced with additional insights about real-world applications...]

[Professionally formatted with clear structure and examples...]
==================================================

📊 Execution Stats:
   • Agents involved: 3
   • Processing time: 2.3s
   • Success: true
```

---

## 🎉 Congratulations!

You just created a **multi-agent system** that:
- ✅ Runs three agents in parallel
- ✅ Combines their outputs intelligently  
- ✅ Handles errors gracefully
- ✅ Provides execution metrics

**And it took less than 5 minutes!**

---

## 🤔 What Just Happened?

### The Magic Behind the Scenes

1. **🏗️ Agent Creation**: Each agent has a specialized role and system prompt
2. **🤝 Collaborative Orchestration**: `CreateCollaborativeRunner` makes agents work together
3. **⚡ Parallel Processing**: All agents process your message simultaneously
4. **🧠 Intelligent Combination**: Results are automatically merged and enhanced
5. **📊 Built-in Monitoring**: You get metrics and error handling for free

### Key Concepts You Just Used

- **`core.AgentHandler`**: The interface for all agents
- **`core.CreateCollaborativeRunner`**: Orchestrates multiple agents in parallel
- **`runner.ProcessMessage`**: Sends messages to all agents and combines results
- **System Prompts**: Give each agent a specialized role and personality

---

## 🚀 Next Steps

Now that you have a working multi-agent system, here's what to explore next:

### 🎓 **15-Minute Tutorials** (Choose Your Path)

<table>
<tr>
<td width="50%">

**🤝 Multi-Agent Patterns**
Learn different orchestration modes:
- Collaborative (parallel)
- Sequential (pipeline)  
- Mixed (hybrid workflows)

[→ Multi-Agent Tutorial](tutorials/multi-agent.md)

</td>
<td width="50%">

**🧠 Memory & RAG**
Add persistent memory and knowledge:
- Vector databases
- Document ingestion
- Semantic search

[→ Memory Tutorial](tutorials/memory-rag.md)

</td>
</tr>
<tr>
<td width="50%">

**🔧 Tool Integration**
Connect to external tools:
- Web search
- File operations
- API integrations
- Custom tools

[→ Tools Tutorial](tutorials/tools.md)

</td>
<td width="50%">

**🏭 Production Ready**
Deploy and scale your agents:
- Docker deployment
- Monitoring setup
- Performance optimization

[→ Production Tutorial](tutorials/production.md)

</td>
</tr>
</table>

### 🎯 **Quick Wins** (5-10 minutes each)

- **[🔄 Try Sequential Processing](how-to/sequential-agents.md)** - Build a data processing pipeline
- **[🌐 Add Web Search](how-to/web-search.md)** - Give your agents internet access
- **[💾 Add Memory](how-to/add-memory.md)** - Make agents remember conversations
- **[📊 Add Monitoring](how-to/monitoring.md)** - See what your agents are doing

### 🏗️ **Build Something Cool**

Ready to build a real application? Try these templates:

```bash
# Research assistant with web search and analysis
agentcli create research-assistant --template research-assistant

# Data processing pipeline with error handling  
agentcli create data-pipeline --template data-processor

# Chat system with persistent memory
agentcli create chat-system --template chatbot --memory-enabled

# Knowledge base with document ingestion
agentcli create knowledge-base --template knowledge-base --rag-enabled
```

---

## 🆘 Need Help?

### Common Issues

**❌ "OpenAI API key not found"**
```bash
# Make sure your API key is set
export OPENAI_API_KEY=sk-your-key-here
echo $OPENAI_API_KEY  # Should show your key
```

**❌ "Module not found"**
```bash
# Make sure you're in the right directory and ran go mod init
go mod tidy
```

**❌ "Context deadline exceeded"**
```bash
# Increase the timeout if agents are taking too long
runner := core.CreateCollaborativeRunner(agents, 60*time.Second) // Increased to 60s
```

### Get Support

- **💬 [Discord Community](https://discord.gg/agenticgokit)** - Real-time help
- **💡 [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)** - Q&A
- **📖 [Troubleshooting Guide](troubleshooting.md)** - Common solutions

---

## 🎯 What's Next?

You've successfully created your first multi-agent system! Here are some paths to continue your AgenticGoKit journey:

<div align="center">

### [🎓 **Take the 15-Minute Tutorial**](tutorials/multi-agent.md)
*Learn advanced orchestration patterns*

### [🏗️ **Build a Real Application**](examples/)
*Explore production-ready examples*

### [📖 **Read the Full Documentation**](../README.md)
*Dive deep into all features*

</div>

---

*⏱️ **Actual time**: Most developers complete this in 3-4 minutes. The extra minute is for reading and understanding!*