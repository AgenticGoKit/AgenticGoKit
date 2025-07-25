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

## Choose Your Approach

### 🚀 **Option A: CLI Approach** (Fastest - 2 minutes)

Perfect for getting started quickly with scaffolded projects.

#### Step 1: Install CLI and Create Project
```bash
# Install the AgenticGoKit CLI
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest

# Create a collaborative multi-agent project
agentcli create my-agents --agents 3 --orchestration-mode collaborative
cd my-agents
```

#### Step 2: Configure and Run
```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your-api-key-here

# Run your multi-agent system
go run main.go
```

### 💻 **Option B: Code-First Approach** (Learn by doing - 3 minutes)

Perfect for understanding how AgenticGoKit works under the hood.

#### Step 1: Create Your Project
```bash
mkdir my-agents && cd my-agents
go mod init my-agents
go get github.com/kunalkushwaha/agenticgokit
```

#### Step 2: Create Configuration

Create `agentflow.toml`:

```toml
[agent_flow]
name = "my-agents"
version = "1.0.0"
provider = "openai"

[logging]
level = "info"
format = "json"

[providers.openai]
# API key will be read from OPENAI_API_KEY environment variable
```

#### Step 3: Write Your Multi-Agent System

Create `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // 🔑 Set up your LLM provider from environment
    provider, err := core.NewProviderFromWorkingDir()
    if err != nil {
        log.Fatalf("Failed to create LLM provider: %v", err)
    }
    
    // 🤖 Create three specialized agents
    agents := map[string]core.AgentHandler{
        "processor": &ProcessorAgent{llm: provider},
        "enhancer":  &EnhancerAgent{llm: provider},
        "formatter": &FormatterAgent{llm: provider},
    }
    
    // 🚀 Create a collaborative runner (agents work together)
    runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
    
    // 💬 Process a message - watch the magic happen!
    fmt.Println("🤖 Starting multi-agent collaboration...")
    
    // Start the runner
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Create an event for processing
    event := core.NewEvent("processor", core.EventData{
        "input": "Explain quantum computing in simple terms",
    }, map[string]string{
        "route": "processor",
    })
    
    // Emit the event to the runner
    if err := runner.Emit(event); err != nil {
        log.Fatalf("Failed to emit event: %v", err)
    }
    
    // Wait for processing to complete
    time.Sleep(5 * time.Second)
    
    fmt.Println("\n✅ Multi-Agent Processing Complete!")
    fmt.Println("=" + strings.Repeat("=", 50))
    fmt.Printf("📊 Execution Stats:\n")
    fmt.Printf("   • Agents involved: %d\n", len(agents))
    fmt.Printf("   • Event ID: %s\n", event.GetID())
}

// ProcessorAgent handles initial processing
type ProcessorAgent struct {
    llm core.ModelProvider
}

func (a *ProcessorAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get user input from event data
    input, ok := event.GetData()["input"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("no input provided")
    }
    
    // Process with LLM
    prompt := core.Prompt{
        System: "You are a processor agent. Extract and organize key information from user requests.",
        User:   fmt.Sprintf("Process this request and extract key information: %s", input),
    }
    
    response, err := a.llm.Call(ctx, prompt)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Update state with processed result
    outputState := core.NewState()
    outputState.Set("processed", response.Content)
    outputState.Set("message", response.Content)
    
    // Route to enhancer
    outputState.SetMeta(core.RouteMetadataKey, "enhancer")
    
    return core.AgentResult{OutputState: outputState}, nil
}

// EnhancerAgent enhances the processed information
type EnhancerAgent struct {
    llm core.ModelProvider
}

func (a *EnhancerAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get processed result from state
    var processed interface{}
    if processedData, exists := state.Get("processed"); exists {
        processed = processedData
    } else if msg, exists := state.Get("message"); exists {
        processed = msg
    } else {
        return core.AgentResult{}, fmt.Errorf("no processed data found")
    }
    
    // Enhance with LLM
    prompt := core.Prompt{
        System: "You are an enhancer agent. Add insights, context, and additional valuable information.",
        User:   fmt.Sprintf("Enhance this response with additional insights: %v", processed),
    }
    
    response, err := a.llm.Call(ctx, prompt)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Update state with enhanced result
    outputState := core.NewState()
    outputState.Set("enhanced", response.Content)
    outputState.Set("message", response.Content)
    
    // Route to formatter
    outputState.SetMeta(core.RouteMetadataKey, "formatter")
    
    return core.AgentResult{OutputState: outputState}, nil
}

// FormatterAgent formats the final response
type FormatterAgent struct {
    llm core.ModelProvider
}

func (a *FormatterAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get enhanced result from state
    var enhanced interface{}
    if enhancedData, exists := state.Get("enhanced"); exists {
        enhanced = enhancedData
    } else if msg, exists := state.Get("message"); exists {
        enhanced = msg
    } else {
        return core.AgentResult{}, fmt.Errorf("no enhanced data found")
    }
    
    // Format with LLM
    prompt := core.Prompt{
        System: "You are a formatter agent. Present information in a clear, professional, and well-structured manner.",
        User:   fmt.Sprintf("Format this response in a clear, professional manner: %v", enhanced),
    }
    
    response, err := a.llm.Call(ctx, prompt)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Update state with final result
    outputState := core.NewState()
    outputState.Set("final_response", response.Content)
    outputState.Set("message", response.Content)
    
    // Print the final result
    fmt.Printf("\n📝 Final Response:\n%s\n", response.Content)
    
    return core.AgentResult{OutputState: outputState}, nil
}
```

#### Step 4: Run It!
```bash
export OPENAI_API_KEY=your-api-key-here
go mod tidy
go run main.go
```

You should see:
```
🤖 Starting multi-agent collaboration...

📝 Final Response:
Quantum computing is a revolutionary technology that uses quantum mechanics 
principles to process information in fundamentally different ways than 
classical computers. Instead of using traditional bits that can only be 
0 or 1, quantum computers use quantum bits (qubits) that can exist in 
multiple states simultaneously through a property called superposition...

✅ Multi-Agent Processing Complete!
==================================================
📊 Execution Stats:
   • Agents involved: 3
   • Event ID: evt_abc123
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

- **`core.AgentHandler`**: The interface for all agents with `Run()` method
- **`core.ModelProvider`**: Interface for LLM providers with `Call()` method
- **`core.CreateCollaborativeRunner`**: Orchestrates multiple agents in parallel
- **`runner.Start()` and `runner.Emit()`**: Start runner and emit events for processing
- **`core.NewEvent()`**: Creates events with data and metadata
- **`core.State`**: Thread-safe state management between agents
- **`core.AgentResult`**: Result structure with output state and error handling
- **`core.Prompt`**: Structured prompt with system and user messages

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

[→ Multi-Agent Tutorial](../core-concepts/orchestration-patterns.md)

</td>
<td width="50%">

**🧠 Memory & RAG**
Add persistent memory and knowledge:
- Vector databases
- Document ingestion
- Semantic search

[→ Memory Tutorial](../memory-systems/)

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

[→ Tools Tutorial](../../reference/api/mcp.md)

</td>
<td width="50%">

**🏭 Production Ready**
Deploy and scale your agents:
- Docker deployment
- Monitoring setup
- Performance optimization

[→ Production Tutorial](../../guides/deployment/README.md)

</td>
</tr>
</table>

### 🎯 **Quick Wins** (5-10 minutes each)

- **[🔄 Try Sequential Processing](../../guides/development/best-practices.md)** - Build a data processing pipeline
- **[🌐 Add Web Search](../../guides/development/web-search-integration.md)** - Give your agents internet access
- **[💾 Add Memory](../memory-systems/basic-memory.md)** - Make agents remember conversations
- **[📊 Add Monitoring](../../guides/development/debugging.md)** - See what your agents are doing

### 🏗️ **Build Something Cool**

Ready to build a real application? Try these examples:

```bash
# Research assistant with web search and analysis
agentcli create research-assistant --mcp-enabled --mcp-tools "web_search,summarize"

# Data processing pipeline with error handling  
agentcli create data-pipeline --orchestration-mode sequential --agents 4

# Chat system with persistent memory
agentcli create chat-system --memory-enabled --memory-provider pgvector

# Knowledge base with document ingestion and RAG
agentcli create knowledge-base --memory-enabled --memory-provider pgvector --rag-enabled
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
- **📖 [Troubleshooting Guide](../../guides/troubleshooting.md)** - Common solutions

---

## 🎯 What's Next?

You've successfully created your first multi-agent system! Here are some paths to continue your AgenticGoKit journey:

<div align="center">

### [🎓 **Take the 15-Minute Tutorial**](../core-concepts/orchestration-patterns.md)
*Learn advanced orchestration patterns*

### [🏗️ **Build a Real Application**](../../../examples/)
*Explore production-ready examples*

### [📖 **Read the Full Documentation**](../../README.md)
*Dive deep into all features*

</div>

---

*⏱️ **Actual time**: Most developers complete this in 3-4 minutes. The extra minute is for reading and understanding!*
