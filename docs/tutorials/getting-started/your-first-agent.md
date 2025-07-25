# Your First Agent Tutorial (15 minutes)

> **Navigation:** [Documentation Home](../../README.md) → [Tutorials](../README.md) → [Getting Started](README.md) → **Your First Agent**

## Overview

Learn how to build your first AgenticGoKit agent using the powerful `agentcli create` command. This tutorial showcases AgenticGoKit's biggest USP - getting started with production-ready agentic code in seconds.

## Prerequisites

- Go 1.21 or later installed
- OpenAI API key (or other LLM provider)
- Basic understanding of Go programming (helpful but not required)

## Learning Objectives

By the end of this tutorial, you'll understand:
- How to use `agentcli create` to scaffold agent projects instantly
- The generated project structure and its components
- How to customize and extend generated agents
- Basic agent concepts through working code
- How to run and test your agent

## What You'll Build

A fully functional agent system that:
- Is generated instantly with `agentcli create`
- Accepts user input and processes requests using an LLM
- Includes proper error handling and logging
- Comes with tests and documentation
- Is ready for production deployment

---

## Step 1: Install AgenticGoKit CLI (1 minute)

The `agentcli` tool is AgenticGoKit's secret weapon for instant agent creation:

```bash
# Install the CLI
go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest

# Verify installation
agentcli --version
```

Set up your environment:

```bash
export OPENAI_API_KEY=your-api-key-here
```

## Step 2: Generate Your Agent Project (30 seconds)

This is where AgenticGoKit shines! Instead of writing boilerplate code, let's generate a complete agent project:

```bash
# Create a single agent project
agentcli create my-first-agent

# Or create with specific options
agentcli create my-first-agent --provider openai --agents 1 --template basic

cd my-first-agent
```

**That's it!** You now have a complete, working agent project. Let's see what was generated:

### Generated Project Structure

```
my-first-agent/
├── main.go                 # Main application entry point with multi-agent orchestration
├── agent1.go              # First agent implementation
├── agent2.go              # Second agent implementation  
├── agentflow.toml         # Configuration file for providers and settings
├── go.mod                 # Go module file
├── README.md              # Project documentation with usage instructions
```

### Let's Look at the Generated Code

**`main.go`** - The application entry point:
```go
package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "sync"
    "time"

    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    ctx := context.Background()
    core.SetLogLevel(core.INFO)
    logger := core.Logger()
    logger.Info().Msg("Starting my-first-agent multi-agent system...")

    messageFlag := flag.String("m", "", "Message to process")
    flag.Parse()

    // Read provider from config
    config, err := core.LoadConfig("agentflow.toml")
    if err != nil {
        fmt.Printf("Failed to load configuration: %v\n", err)
        os.Exit(1)
    }

    llmProvider, err := initializeProvider(config.AgentFlow.Provider)
    if err != nil {
        fmt.Printf("Failed to initialize LLM provider '%s': %v\n", config.AgentFlow.Provider, err)
        fmt.Printf("Make sure you have set the appropriate environment variables:\n")
        switch config.AgentFlow.Provider {
        case "openai":
            fmt.Printf("  OPENAI_API_KEY=your-api-key\n")
        case "azure":
            fmt.Printf("  AZURE_OPENAI_API_KEY=your-api-key\n")
            fmt.Printf("  AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/\n")
        }
        os.Exit(1)
    }

    // Create agents with result collection
    agents := make(map[string]core.AgentHandler)
    results := make([]AgentOutput, 0)
    var resultsMutex sync.Mutex

    // Wrap agents with result collectors
    agent1 := NewAgent1(llmProvider)
    wrappedAgent1 := &ResultCollectorHandler{
        originalHandler: agent1,
        agentName:       "agent1",
        outputs:         &results,
        mutex:           &resultsMutex,
    }
    agents["agent1"] = wrappedAgent1

    agent2 := NewAgent2(llmProvider)
    wrappedAgent2 := &ResultCollectorHandler{
        originalHandler: agent2,
        agentName:       "agent2",
        outputs:         &results,
        mutex:           &resultsMutex,
    }
    agents["agent2"] = wrappedAgent2

    // Create collaborative runner
    runner := core.CreateCollaborativeRunner(agents, 30*time.Second)

    var message string
    if *messageFlag != "" {
        message = *messageFlag
    } else {
        fmt.Print("Enter your message: ")
        fmt.Scanln(&message)
    }

    if message == "" {
        message = "Hello! Please provide information about current topics."
    }

    // Start the runner and process
    runner.Start(ctx)
    defer runner.Stop()

    event := core.NewEvent("agent1", core.EventData{
        "message": message,
    }, map[string]string{
        "route": "agent1",
    })

    if err := runner.Emit(event); err != nil {
        logger.Error().Err(err).Msg("Workflow execution failed")
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    // Wait for processing and display results
    time.Sleep(5 * time.Second)
    
    fmt.Printf("\n=== Agent Responses ===\n")
    resultsMutex.Lock()
    for _, result := range results {
        fmt.Printf("\n🤖 %s:\n%s\n", result.AgentName, result.Content)
        fmt.Printf("⏰ %s\n", result.Timestamp.Format("15:04:05"))
    }
    resultsMutex.Unlock()

    fmt.Printf("\n=== Workflow Completed ===\n")
}

func initializeProvider(providerType string) (core.ModelProvider, error) {
    return core.NewProviderFromWorkingDir()
}
```

**`agent1.go`** - Your first agent implementation:
```go
package main

import (
    "context"
    "fmt"
    "strings"

    agentflow "github.com/kunalkushwaha/agenticgokit/core"
)

// Agent1Handler represents the agent1 agent handler
// Purpose: Handles specialized task processing within the workflow
type Agent1Handler struct {
    llm agentflow.ModelProvider
}

// NewAgent1 creates a new Agent1 instance
func NewAgent1(llmProvider agentflow.ModelProvider) *Agent1Handler {
    return &Agent1Handler{llm: llmProvider}
}

// Run implements the agentflow.AgentHandler interface
func (a *Agent1Handler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    logger := agentflow.Logger()
    logger.Debug().Str("agent", "agent1").Str("event_id", event.GetID()).Msg("Agent processing started")
    
    // Get message from event or state
    var inputToProcess interface{}
    eventData := event.GetData()
    if msg, ok := eventData["message"]; ok {
        inputToProcess = msg
    } else if stateMessage, exists := state.Get("message"); exists {
        inputToProcess = stateMessage
    } else {
        inputToProcess = "No message provided"
    }
    
    // System prompt with detailed responsibilities
    systemPrompt := `You are Agent1, handles specialized task processing within the workflow.

Core Responsibilities:
- Process the original user request and provide initial analysis
- Use available MCP tools to gather current and accurate information
- For financial queries: get current prices, market data, and trends
- Provide concrete answers with actual data rather than generic advice
- Set a strong foundation for subsequent agents

Tool Usage Strategy:
- For stock prices/financial data: Use search tools to find current information
- For current events/news: Use search tools for latest updates
- For specific web content: Use fetch_content tool with URLs
- Always prefer real data over general advice
- Document tool usage and results clearly

Response Quality:
- Provide specific, data-driven answers when possible
- Extract and present key information clearly
- Be conversational but professional
- Integrate tool results naturally into responses

Route Mode: Process tasks and route to appropriate next steps in the workflow.`
    
    // Get available MCP tools to include in prompt
    var toolsPrompt string
    mcpManager := agentflow.GetMCPManager()
    if mcpManager != nil {
        availableTools := mcpManager.GetAvailableTools()
        logger.Debug().Str("agent", "agent1").Int("tool_count", len(availableTools)).Msg("MCP Tools discovered")
        toolsPrompt = agentflow.FormatToolsPromptForLLM(availableTools)
    } else {
        logger.Warn().Str("agent", "agent1").Msg("MCP Manager is not available")
    }
    
    // Create initial LLM prompt with available tools information
    userPrompt := fmt.Sprintf("User query: %v", inputToProcess)
    userPrompt += toolsPrompt
    
    prompt := agentflow.Prompt{
        System: systemPrompt,
        User:   userPrompt,
    }
    
    // Call LLM to get initial response and potential tool calls
    response, err := a.llm.Call(ctx, prompt)
    if err != nil {
        return agentflow.AgentResult{}, fmt.Errorf("Agent1 LLM call failed: %w", err)
    }
    
    // Parse LLM response for tool calls and execute them
    toolCalls := agentflow.ParseLLMToolCalls(response.Content)
    var mcpResults []string
    
    // Execute any requested tools
    if len(toolCalls) > 0 && mcpManager != nil {
        logger.Info().Str("agent", "agent1").Int("tool_calls", len(toolCalls)).Msg("Executing LLM-requested tools")
        
        for _, toolCall := range toolCalls {
            if toolName, ok := toolCall["name"].(string); ok {
                var args map[string]interface{}
                if toolArgs, exists := toolCall["args"]; exists {
                    if argsMap, ok := toolArgs.(map[string]interface{}); ok {
                        args = argsMap
                    } else {
                        args = make(map[string]interface{})
                    }
                } else {
                    args = make(map[string]interface{})
                }
                
                // Execute tool using the global ExecuteMCPTool function
                result, err := agentflow.ExecuteMCPTool(ctx, toolName, args)
                if err != nil {
                    logger.Error().Str("agent", "agent1").Str("tool_name", toolName).Err(err).Msg("Tool execution failed")
                    mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' failed: %v", toolName, err))
                } else if result.Success {
                    logger.Info().Str("agent", "agent1").Str("tool_name", toolName).Msg("Tool execution successful")
                    
                    var resultContent string
                    if len(result.Content) > 0 {
                        resultContent = result.Content[0].Text
                    } else {
                        resultContent = "Tool executed successfully but returned no content"
                    }
                    
                    mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' result: %s", toolName, resultContent))
                }
            }
        }
    }
    
    // Generate final response if tools were used
    var finalResponse string
    if len(mcpResults) > 0 {
        // Create enhanced prompt with tool results
        enhancedPrompt := agentflow.Prompt{
            System: systemPrompt,
            User:   fmt.Sprintf("Original query: %v\n\nTool results:\n%s\n\nPlease provide a comprehensive response incorporating these tool results:", inputToProcess, strings.Join(mcpResults, "\n")),
        }
        
        // Get final response from LLM
        finalLLMResponse, err := a.llm.Call(ctx, enhancedPrompt)
        if err != nil {
            return agentflow.AgentResult{}, fmt.Errorf("Agent1 final LLM call failed: %w", err)
        }
        finalResponse = finalLLMResponse.Content
    } else {
        finalResponse = response.Content
    }
    
    // Store agent response in state for potential use by subsequent agents
    outputState := agentflow.NewState()
    outputState.Set("agent1_response", finalResponse)
    outputState.Set("message", finalResponse)
    
    // Route to the next agent (Agent2) in the workflow
    outputState.SetMeta(agentflow.RouteMetadataKey, "agent2")
    
    logger.Info().Str("agent", "agent1").Msg("Agent processing completed successfully")
    
    return agentflow.AgentResult{
        OutputState: outputState,
    }, nil
}
```

**`agentflow.toml`** - Configuration file:
```toml
# AgenticGoKit Configuration

[agent_flow]
name = "my-first-agent"
version = "1.0.0"
provider = "openai"

[logging]
level = "info"
format = "json"

[runtime]
max_concurrent_agents = 10
timeout_seconds = 30

[providers.azure]
# API key will be read from AZURE_OPENAI_API_KEY environment variable
# Endpoint will be read from AZURE_OPENAI_ENDPOINT environment variable
# Deployment will be read from AZURE_OPENAI_DEPLOYMENT environment variable

[providers.openai]
# API key will be read from OPENAI_API_KEY environment variable

[providers.ollama]
endpoint = "http://localhost:11434"
model = "llama2"

[providers.mock]
# Mock provider for testing - no configuration needed
```

The generated code also includes a **ResultCollectorHandler** that wraps your agents to capture and display their outputs:

```go
// ResultCollectorHandler wraps an agent handler to capture its outputs
type ResultCollectorHandler struct {
    originalHandler core.AgentHandler
    agentName       string
    outputs         *[]AgentOutput
    mutex           *sync.Mutex
}

// AgentOutput holds the output from an agent
type AgentOutput struct {
    AgentName string
    Content   string
    Timestamp time.Time
}

// Run implements the AgentHandler interface and captures the output
func (r *ResultCollectorHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Call the original handler
    result, err := r.originalHandler.Run(ctx, event, state)

    // Extract meaningful content from the result
    var content string
    if err != nil {
        content = fmt.Sprintf("Error: %v", err)
    } else if result.Error != "" {
        content = fmt.Sprintf("Agent Error: %s", result.Error)
    } else {
        // Try to extract content from the result's output state
        if result.OutputState != nil {
            if responseData, exists := result.OutputState.Get("response"); exists {
                if responseStr, ok := responseData.(string); ok {
                    content = responseStr
                }
            }
            // Additional fallbacks for different state keys...
        }
    }

    // Store the output with timestamp
    r.mutex.Lock()
    *r.outputs = append(*r.outputs, AgentOutput{
        AgentName: r.agentName,
        Content:   content,
        Timestamp: time.Now(),
    })
    r.mutex.Unlock()

    return result, err
}
```

## Step 3: Run Your Agent (1 minute)

The generated project is ready to run immediately:

```bash
# Install dependencies (if needed)
go mod tidy

# Run your agent
go run main.go
```

You'll see an interactive interface where you can interact with your agents:

```bash
# Run with a message flag
go run . -m "What's the weather like today?"

# Or run interactively
go run .
Enter your message: Hello! What can you do?
```

The output will show both agents processing your request:
```
=== Agent Responses ===

🤖 agent1:
Hello! I'm Agent1, specialized in task processing within the workflow. I can help you with:

- Processing and analyzing your requests
- Gathering current information using available tools
- Providing data-driven answers when possible
- Setting up a foundation for more detailed analysis

I work collaboratively with Agent2 to provide comprehensive responses. What would you like help with today?
⏰ 14:23:15

🤖 agent2:
Based on Agent1's initial analysis, I can provide you with comprehensive final responses. I specialize in:

- Synthesizing insights from previous agents
- Presenting information in a clear, organized manner
- Ensuring responses fully address your questions
- Using additional tools if critical information is missing

Together, we form a collaborative system designed to give you thorough, well-researched answers. How can we assist you today?
⏰ 14:23:17

=== Workflow Completed ===
```

**That's it!** You have a fully functional agent running in under 2 minutes!

## Step 4: Explore Advanced CLI Options (3 minutes)

The `agentcli create` command has many powerful options. Here are some examples of what you can do:

### Multi-Agent Systems

Create a system with multiple agents working together:

```bash
# Create a multi-agent system
agentcli create research-team --agents 3

cd research-team
go run main.go
```

### With Different Orchestration Modes

Specify how your agents should collaborate:

```bash
# Create a system with sequential orchestration
agentcli create workflow-agent --orchestration-mode sequential

cd workflow-agent
go run main.go
```

### With Memory Integration

Create an agent with persistent memory capabilities:

```bash
# Create agent with memory
agentcli create memory-agent --memory-enabled

cd memory-agent
go run main.go
```

### With Tool Integration

Create an agent that can use external tools:

```bash
# Create agent with tool integration
agentcli create tool-agent --mcp-enabled

cd tool-agent
go run main.go
```

### Get Help with All Options

To see all available options:

```bash
# View all options
agentcli create --help

# Or use interactive mode
agentcli create --interactive
```

## Step 5: Customize Your Generated Agent (5 minutes)

The generated code is fully customizable. Let's enhance your agent:

### 1. Customize the Agent Behavior

Find the agent implementation file in your generated project and customize the system prompt to give your agent personality:

```go
// Find the Run method in your agent implementation
func (a *YourAgent) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    message, ok := event.GetData()["message"].(string)
    if !ok {
        return core.AgentResult{}, fmt.Errorf("no message found in event data")
    }
    
    // Custom system prompt with personality
    systemPrompt := `You are Alex, a friendly and knowledgeable AI assistant with expertise in:
    - Software development and programming
    - Technology trends and best practices
    - Problem-solving and debugging
    
    You respond with enthusiasm and provide practical, actionable advice.
    Always ask follow-up questions to better understand the user's needs.`
    
    // Generate response using your custom prompt
    // The exact implementation will depend on your generated code
    // but will follow this general pattern
    response, err := a.generateResponse(ctx, systemPrompt, message)
    if err != nil {
        return core.AgentResult{}, err
    }
    
    // Update state with response
    state.Set("response", response)
    
    return core.AgentResult{OutputState: state}, nil
}
```

### 2. Add Error Handling and Fallbacks

Enhance your agent with better error handling:

```go
// Add a retry function to your agent
func (a *YourAgent) generateWithRetry(ctx context.Context, prompt string, maxRetries int) (string, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        response, err := a.llmProvider.Generate(ctx, prompt)
        if err == nil {
            return response, nil
        }
        
        lastErr = err
        
        // Exponential backoff
        if i < maxRetries-1 {
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }
    
    return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// Use it in your Run method
response, err := a.generateWithRetry(ctx, fullPrompt, 3)
if err != nil {
    // Provide fallback response
    fallback := "I'm having trouble processing your request right now. Could you please try again?"
    state.Set("response", fallback)
    state.Set("error_occurred", true)
    
    // Return success with fallback instead of error
    return core.AgentResult{OutputState: state}, nil
}
```

### 3. Test Your Customizations

Test your customized agent to ensure it works as expected:

```bash
# Run your modified agent
go run main.go

```

## 🎉 Congratulations!

You've experienced AgenticGoKit's biggest USP - **instant agent creation**! In just a few minutes, you:

- ✅ **Generated a complete agent project** with `agentcli create`
- ✅ **Got production-ready code** with proper structure and configuration
- ✅ **Ran a functional agent** immediately without writing boilerplate
- ✅ **Learned advanced CLI options** for different use cases
- ✅ **Customized the generated code** to fit your needs

## Why `agentcli create` is Game-Changing

### Traditional Approach (Hours)
```bash
mkdir my-agent
cd my-agent
go mod init my-agent
# Write main.go (100+ lines)
# Write agent.go (50+ lines)  
# Write config.toml
# Write tests
# Write README
# Debug imports and interfaces
# Handle errors and edge cases
```

### AgenticGoKit Approach (Seconds)
```bash
agentcli create my-agent
cd my-agent
go run main.go  # It just works!
```

## Key Features You Get for Free

- **🏗️ Project Structure**: Organized, professional layout
- **⚙️ Configuration**: TOML-based config with sensible defaults
- **🧪 Tests**: Generated test files with examples
- **📚 Documentation**: README with usage instructions
- **🐳 Docker**: Optional containerization setup
- **📊 Monitoring**: Optional metrics and logging
- **🔧 Tools**: MCP integration ready to go
- **💾 Memory**: Vector database integration available

## CLI Command Reference

```bash
# Basic agent
agentcli create my-agent

# Multi-agent system
agentcli create team --agents 3

# With memory
agentcli create smart-agent --memory-enabled

# With tools
agentcli create tool-agent --mcp-enabled

# Get help
agentcli create --help

# Interactive mode
agentcli create --interactive
```

## Next Steps

Now that you've seen the power of `agentcli create`, explore:

### 🚀 **Immediate Next Steps**
- **[Multi-Agent Orchestration](../core-concepts/orchestration-patterns.md)** - Create collaborative agent teams
- **[Memory Systems](../memory-systems/basic-memory.md)** - Add persistent memory and RAG
- **[MCP Tools](../mcp/tool-integration.md)** - Connect external tools and APIs

### 🎓 **Learning Path**
- **[Core Concepts](../core-concepts/README.md)** - Understand agent fundamentals
- **[Memory Systems](../memory-systems/README.md)** - Build knowledge-aware agents
- **[Advanced Patterns](../advanced/README.md)** - Complex orchestration patterns

### 🏭 **Production Ready**
- **[Best Practices](../../guides/development/best-practices.md)** - Code quality and patterns
- **[Production Deployment](../../guides/deployment/README.md)** - Deploy and scale your agents

## Troubleshooting

**Common Issues:**

1. **"agentcli: command not found"**: Run `go install github.com/kunalkushwaha/agenticgokit/cmd/agentcli@latest`
2. **"OpenAI API key not found"**: Set `export OPENAI_API_KEY=your-key`
3. **"Module not found"**: Run `go mod tidy` in the generated project
4. **"Permission denied"**: Make sure you have write permissions in the directory

**Pro Tips:**
- Use `agentcli create --interactive` for guided project creation
- Check `agentcli create --help` for all available options
- Generated projects include comprehensive README files
- All generated code is fully customizable and production-ready

**Need Help?**
- [CLI Reference](../../reference/cli.md) - Complete command documentation
- [Troubleshooting Guide](../../guides/troubleshooting.md) - Common issues and solutions
- [Discord Community](https://discord.gg/agenticgokit) - Get help from the community
- [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions) - Technical discussions

---

**🚀 Ready to build something amazing?** Try `agentcli create --interactive` for your next project!