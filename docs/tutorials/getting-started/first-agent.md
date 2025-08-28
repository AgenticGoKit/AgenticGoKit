---
title: "Your First Agent"
description: "Create and run your first intelligent agent with AgenticGoKit"
prev:
  text: "Understanding Agents"
  link: "./understanding-agents"
next:
  text: "Agent Configuration"
  link: "./agent-configuration"
---

# Your First Agent

Now that you understand what agents are, let's create your first intelligent agent! This hands-on experience will show you how AgenticGoKit makes it easy to go from idea to working agent.

## Learning Objectives

By the end of this section, you'll have:
- Created a working agent project using the CLI
- Understood the project structure and key files
- Successfully run your first agent and seen it respond
- Made your first customizations to agent behavior
- Gained confidence in the AgenticGoKit development process

## Prerequisites

Before starting, make sure you've completed:
- ✅ [Installation](./installation.md) - AgenticGoKit CLI installed and working
- ✅ LLM provider configured (OpenAI, Azure OpenAI, or Ollama)

## Step 1: Create Your First Agent Project

Let's create a simple but functional agent that can help with general questions:

```bash
# Create a new agent project
agentcli create my-first-agent --template basic

# Navigate into the project directory
cd my-first-agent
```

::: tip What Just Happened?
The CLI created a complete Go project with everything needed to run an intelligent agent. The `basic` template gives you a single agent that's perfect for learning.
:::

## Step 2: Explore the Project Structure

Let's look at what was created:

```bash
# List all files in the project
ls -la
```

You should see:

```
my-first-agent/
├── main.go              # Application entry point
├── agentflow.toml       # Agent configuration
├── go.mod               # Go module definition
├── go.sum               # Go dependencies
├── agents/              # Agent implementations
│   └── basic_agent.go   # Your agent's code
└── README.md            # Project documentation
```

### Understanding Each File

**`main.go`** - The heart of your application:
```go
package main

import (
    "context"
    "log"
    
    "github.com/kunalkushwaha/agenticgokit/core"
    // Plugin imports for LLM providers, orchestration, etc.
)

func main() {
    // Creates a runner from your configuration
    runner, err := core.NewRunnerFromConfig("agentflow.toml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Starts the agent system
    ctx := context.Background()
    err = runner.Start(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer runner.Stop()
    
    // Your agent is now ready to handle requests!
}
```

**`agentflow.toml`** - Your agent's configuration:
```toml
[agent_flow]
name = "my-first-agent"
version = "1.0.0"

[llm]
provider = "openai"  # or "azure" or "ollama"
model = "gpt-4"
temperature = 0.7

[agents.assistant]
role = "helpful_assistant"
description = "A friendly and knowledgeable assistant"
system_prompt = """
You are a helpful assistant that provides clear, accurate information.
You're friendly, professional, and always try to be as helpful as possible.
"""
enabled = true
```

**`agents/basic_agent.go`** - Your agent's implementation:
```go
package agents

import (
    "context"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

type BasicAgent struct {
    name        string
    role        string
    description string
    systemPrompt string
}

func (a *BasicAgent) Run(ctx context.Context, inputState core.State) (core.State, error) {
    // This is where your agent processes requests and generates responses
    // The framework handles LLM communication based on your configuration
    return inputState, nil
}
```

## Step 3: Validate Your Configuration

Before running the agent, let's make sure everything is configured correctly:

```bash
agentcli validate
```

You should see:
```
✅ Status: VALID
Configuration is correct and ready to use.
```

If you see any errors, check:
- Your LLM provider environment variables are set
- The `agentflow.toml` file syntax is correct
- All required fields are present

## Step 4: Run Your First Agent

Now for the exciting part - let's run your agent!

```bash
# Run your agent with a simple question
go run . -m "Hello! Can you tell me what you can help me with?"
```

**What you should see:**

1. **Startup messages** - The agent system initializing
2. **Agent response** - Your agent introducing itself and explaining its capabilities
3. **Clean shutdown** - The system shutting down gracefully

**Example output:**
```
2024/01/15 10:30:15 Starting AgenticGoKit runner...
2024/01/15 10:30:15 Agent 'assistant' initialized successfully
2024/01/15 10:30:16 Processing message: "Hello! Can you tell me what you can help me with?"

Hello! I'm your helpful assistant. I can help you with a wide variety of tasks including:

- Answering questions on many topics
- Helping with problem-solving and analysis
- Providing explanations and tutorials
- Assisting with writing and communication
- Offering suggestions and recommendations

What would you like help with today?

2024/01/15 10:30:17 Agent processing completed successfully
2024/01/15 10:30:17 Shutting down runner...
```

::: tip Success!
If you see a response from your agent, congratulations! You've successfully created and run your first AI agent with AgenticGoKit.
:::

## Step 5: Experiment with Different Questions

Try asking your agent different types of questions to see how it responds:

```bash
# Ask for information
go run . -m "What are the benefits of using Go for backend development?"

# Ask for help with a task
go run . -m "Can you help me understand how to structure a Go project?"

# Ask for creative input
go run . -m "Give me 5 ideas for a simple Go CLI application"
```

Notice how your agent adapts its responses based on the type of question you ask.

## Step 6: Make Your First Customization

Let's personalize your agent by modifying its personality and capabilities.

### Customize the Agent's Personality

Open `agentflow.toml` and modify the system prompt:

```toml
[agents.assistant]
role = "coding_mentor"
description = "A friendly Go programming mentor and guide"
system_prompt = """
You are an experienced Go developer and mentor who loves helping others learn.
You provide clear, practical advice with code examples when helpful.
You're encouraging, patient, and always explain the 'why' behind your suggestions.
When discussing Go, you emphasize best practices, idiomatic code, and the Go philosophy.
"""
```

### Test Your Customization

```bash
go run . -m "I'm new to Go. What should I learn first?"
```

Notice how the agent's response now reflects its new role as a Go programming mentor!

### Try Different Roles

Experiment with different agent personalities:

**Creative Writer:**
```toml
system_prompt = """
You are a creative writing assistant who helps with storytelling, 
character development, and narrative structure. You're imaginative, 
encouraging, and always ready to help brainstorm ideas.
"""
```

**Data Analyst:**
```toml
system_prompt = """
You are a data analyst who helps interpret data, create visualizations, 
and explain statistical concepts. You're methodical, precise, and 
excellent at making complex data insights accessible.
"""
```

**Business Consultant:**
```toml
system_prompt = """
You are a business consultant who helps with strategy, planning, 
and problem-solving. You ask insightful questions and provide 
structured, actionable advice.
"""
```

## Step 7: Understanding What Happened

Let's break down what occurred when you ran your agent:

1. **Configuration Loading**: AgenticGoKit read your `agentflow.toml` file
2. **Agent Creation**: The framework created your agent with the specified role and prompt
3. **LLM Connection**: Connected to your configured LLM provider (OpenAI, Azure, or Ollama)
4. **Message Processing**: Your input was processed by the agent
5. **Response Generation**: The LLM generated a response based on your agent's system prompt
6. **Output**: The response was displayed and the system shut down cleanly

## Step 8: Hands-On Exercises

Try these exercises to deepen your understanding:

### Exercise 1: Create a Specialized Agent
Create an agent specialized for a specific domain you're interested in (cooking, fitness, technology, etc.).

### Exercise 2: Test Edge Cases
Try asking your agent:
- Very long questions
- Questions outside its expertise
- Questions in different languages
- Follow-up questions

### Exercise 3: Modify LLM Settings
In `agentflow.toml`, try adjusting:
- `temperature` (0.1 for focused, 0.9 for creative)
- `model` (if you have access to different models)
- `max_tokens` (to control response length)

## Troubleshooting Common Issues

### Agent Doesn't Respond
**Check:**
- Environment variables are set correctly
- Internet connection (for cloud providers)
- Ollama is running (for local setup)

### "Provider not registered" Error
**Solution:** The generated project includes necessary imports, but ensure your `main.go` has:
```go
import (
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/openai"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"
    // ... other necessary plugins
)
```

### Configuration Validation Fails
**Check:**
- TOML syntax is correct (no missing quotes, brackets)
- All required fields are present
- Environment variables match your provider choice

## What You've Learned

✅ **Created your first agent project** using the AgenticGoKit CLI  
✅ **Understood the project structure** and key files  
✅ **Successfully ran an agent** and saw it respond intelligently  
✅ **Customized agent behavior** through configuration  
✅ **Experimented with different** agent personalities and roles  
✅ **Learned the agent execution flow** from input to output  
✅ **Gained hands-on experience** with AgenticGoKit development  

## Understanding Check

Before moving on, make sure you can:
- [ ] Create a new agent project with the CLI
- [ ] Modify the agent's system prompt and see the changes
- [ ] Run the agent and get responses to different questions
- [ ] Explain what each file in the project does
- [ ] Troubleshoot basic configuration issues

## Next Steps

You've successfully created and customized your first agent! Now let's dive deeper into agent configuration to unlock more powerful capabilities.

**[→ Continue to Agent Configuration](./agent-configuration.md)**

---

::: details Quick Navigation

**Previous:** [Understanding Agents](./understanding-agents.md) - Core concepts and mental models  
**Next:** [Agent Configuration](./agent-configuration.md) - Customize your agent's behavior  
**Jump to:** [Multi-Agent Basics](./multi-agent-basics.md) - Skip to multi-agent systems  

:::

::: details What's Next?

**Ready to customize your agent?**
- [Agent Configuration](./agent-configuration.md) - Master the configuration system
- [Examples](https://github.com/kunalkushwaha/agenticgokit/tree/main/examples/01-basic-agent) - Study the basic agent example

**Want to see more capabilities?**
- [Multi-Agent Basics](./multi-agent-basics.md) - Multiple agents working together
- [Adding Memory](./adding-memory.md) - Persistent memory and knowledge bases
- [Tool Integration](./tool-integration.md) - Connect to external services

**Need help or want to share?**
- [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions) - Share your first agent!
- [Troubleshooting](./troubleshooting.md) - If you encountered any issues

:::

::: tip Congratulations!
You've just built your first AI agent with AgenticGoKit! This is the foundation for everything else you'll learn. The same patterns you used here - configuration, validation, and execution - apply to much more complex agent systems.
:::
