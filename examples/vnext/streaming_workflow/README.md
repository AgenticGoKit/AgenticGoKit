# Multi-Agent Streaming Workflow Example

This example demonstrates how **streaming** enhances user experience in **multi-agent workflows**. Instead of waiting for each agent to complete entirely, users see real-time progress as specialized agents collaborate and hand off results to each other.

## 🎯 What This Example Shows

### Multi-Agent Collaboration Workflow
1. **🔍 Researcher Agent ("Dr. Research")** - Conducts comprehensive research on topics
2. **📝 Summarizer Agent ("Ms. Summarizer")** - Distills research findings into key insights

### Streaming Benefits Demonstrated
- **Real-time Agent Work**: See each agent's tokens arriving as they're generated
- **Agent Specialization**: Different roles with distinct system prompts and parameters
- **Agent Handoffs**: Watch results flow seamlessly between specialized agents
- **Progress Tracking**: Token counts, timing, and completion status per agent
- **Better Engagement**: Users stay engaged watching collaborative AI process
- **Workflow Transparency**: Clear visibility into multi-agent collaboration

## 🚀 Running the Example

### Prerequisites
1. **Ollama** installed and running:
   ```bash
   # Install Ollama (if not already installed)
   # Visit: https://ollama.ai/download
   
   # Pull required model  
   ollama pull gemma3:1b
   
   # Start Ollama (if not running)
   ollama serve
   ```

2. **Go environment** set up in the AgenticGoKit project

### Running the Multi-Agent Workflow
```bash
cd examples/vnext/streaming_workflow
go run main.go
```

### What You'll See
- **Phase 1**: Researcher agent streams comprehensive analysis
- **Phase 2**: Summarizer agent processes and condenses results  
- **Real-time tokens**: Watch each agent's thought process unfold
- **Performance metrics**: Token counts and timing for each agent

## 📋 Example Output

```
🌟 Multi-Agent Streaming Workflow: Benefits of Streaming in AI Applications
============================================================
🤖 Workflow: Researcher → Summarizer
⚡ Watch real-time streaming from each agent!

📅 PHASE 1: RESEARCH
————————————————————

🤖 🔍 Dr. Research Starting...
   🔍 Streaming response...
   Okay, let's dive into the benefits of streaming in AI applications! It's a rapidly growing field...
   🔍 [Dr. Research] 25 tokens, 2.5s...
   with huge potential, and it's moving beyond just simple video streaming to become a foundational...
   🔍 [Dr. Research] 50 tokens, 3.6s...
   component of many AI systems. Here's a comprehensive breakdown:

   **1. Key Background Information:**
   � [Dr. Research] 75 tokens, 4.7s...
   ...
   ✅ Dr. Research completed in 55.71s (1010 tokens)

� PHASE 2: SUMMARIZATION
—————————————————————————

🤖 📝 Ms. Summarizer Starting...
   📝 Streaming response...
   Okay, let's focus on understanding the core insights and resources for you...
   📝 [Ms. Summarizer] 25 tokens, 25.4s...
   **Core Insights (Condensed):**
   
   Streaming is a foundational shift in AI, moving beyond static datasets...
   📝 [Ms. Summarizer] 50 tokens, 27.0s...
   ...
   ✅ Ms. Summarizer completed in 58.10s (536 tokens)

============================================================
🎉 MULTI-AGENT WORKFLOW COMPLETED!
============================================================
📊 Research Output: 5075 characters
📊 Summary Output: 2000 characters
🤖 Agents Used: 2 specialized agents
⚡ Streaming: Real-time progress visibility
🔄 Collaboration: Research results fed into summarizer
```

## 🔍 Code Architecture

### Key Components

1. **SpecializedAgent**: Individual agents with specific roles and expertise
2. **MultiAgentWorkflow**: Orchestrates multiple specialized agents
3. **Real-time Streaming**: Uses `vnext.Stream.Chunks()` to process tokens as they arrive
4. **Agent Collaboration**: Results from one agent feed into the next

### Multi-Agent Architecture
```go
// Create specialized agents with different roles
type SpecializedAgent struct {
    agent vnext.Agent
    name  string
    icon  string
}

// Researcher Agent - specialized for information gathering
researcher := CreateResearcherAgent() // Temperature: 0.2, MaxTokens: 800

// Summarizer Agent - specialized for distillation
summarizer := CreateSummarizerAgent() // Temperature: 0.3, MaxTokens: 400
```

### Streaming Flow
```go
// Agent executes with streaming visualization
func (sa *SpecializedAgent) Execute(ctx context.Context, prompt string) (string, error) {
    stream, err := sa.agent.RunStream(ctx, prompt)
    
    for chunk := range stream.Chunks() {
        switch chunk.Type {
        case vnext.ChunkTypeDelta:
            // Handle individual tokens with role-specific icons
            fmt.Print(chunk.Delta)
            result += chunk.Delta
            
        case vnext.ChunkTypeDone:
            // Show completion with agent-specific stats
            fmt.Printf("✅ %s completed in %.2fs (%d tokens)", sa.name, duration, tokenCount)
        }
    }
}
```

### Agent Collaboration Flow
```go
// Phase 1: Research
researchResult, err := workflow.researcher.Execute(ctx, researchPrompt)

// Phase 2: Summarization (using research results)
summaryPrompt := fmt.Sprintf(`Based on this research: %s...`, researchResult)
summaryResult, err := workflow.summarizer.Execute(ctx, summaryPrompt)
```

## 💡 Why Streaming Matters for Multi-Agent Workflows

### Without Streaming (Traditional Multi-Agent)
```
User: "Research AI streaming benefits"
System: [Agent 1 working... 45 seconds of silence]
System: [Agent 2 working... 30 seconds of silence]
System: [All results appear at once after 75 seconds]
```

### With Streaming (This Example)
```
User: "Research AI streaming benefits"
System: "🤖 � Dr. Research Starting..."
System: "Okay, let's dive into the benefits..." [tokens stream live]
System: "🔍 [Dr. Research] 25 tokens, 2.5s..."
System: "✅ Dr. Research completed in 55.71s (1010 tokens)"
System: "🤖 � Ms. Summarizer Starting..."
System: "Based on the research findings..." [tokens stream live]
System: "✅ Ms. Summarizer completed in 58.10s (536 tokens)"
```

### Multi-Agent Streaming Benefits
1. **Agent Transparency**: See each agent's thinking process in real-time
2. **Workflow Visibility**: Clear indication of which agent is working
3. **Collaboration Tracking**: Watch results flow between specialized agents
4. **Perceived Performance**: Feels engaging even during long agent operations
5. **Early Quality Assessment**: Evaluate agent output as it's generated
6. **Interruptible Workflows**: Can stop or redirect mid-execution
7. **Role-Based Progress**: Different visualization for different agent types

## 🛠 Customization

### Different LLM Providers
```go
// OpenAI
LLM: vnext.LLMConfig{
    Provider: "openai",
    Model:    "gpt-4",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
}

// Azure OpenAI
LLM: vnext.LLMConfig{
    Provider: "azure",
    Model:    "gpt-4",
    BaseURL:  "https://your-resource.openai.azure.com/",
    APIKey:   os.Getenv("AZURE_OPENAI_KEY"),
}
```

### Adding New Agent Types
```go
// Create a new specialized agent
func CreateAnalystAgent() (*SpecializedAgent, error) {
    agent, err := vnext.QuickChatAgentWithConfig("Prof. Analyst", &vnext.Config{
        SystemPrompt: `You are an Analysis Agent specialized in:
- Data synthesis and pattern recognition
- Critical thinking and evaluation
- Drawing insights from complex information`,
        LLM: vnext.LLMConfig{
            Provider:    "ollama",
            Model:       "gemma3:1b", 
            Temperature: 0.4, // Balanced for analysis
            MaxTokens:   600,
        },
    })
    
    return &SpecializedAgent{
        agent: agent,
        name:  "Prof. Analyst",
        icon:  "�",
    }, nil
}
```

### Custom Agent Workflows
```go
// Create a 3-agent workflow
type ExtendedWorkflow struct {
    researcher *SpecializedAgent
    analyst    *SpecializedAgent
    summarizer *SpecializedAgent
}

// Execute with 3 phases
researchResult, _ := workflow.researcher.Execute(ctx, researchPrompt)
analysisResult, _ := workflow.analyst.Execute(ctx, analysisPrompt, researchResult)
summaryResult, _ := workflow.summarizer.Execute(ctx, summaryPrompt, researchResult, analysisResult)
```

### Progress Customization
```go
// Custom progress indicators per agent type
if tokenCount%15 == 0 {  // More frequent updates for summarizer
    fmt.Printf("%s [%s] %d tokens, %.1fs...\n", sa.icon, sa.name, tokenCount, elapsed)
}
```

## 🎭 Use Cases

This streaming workflow pattern is ideal for:

- **Research & Analysis**: Multi-step research like this example
- **Content Creation**: Planning → Drafting → Editing → Finalizing
- **Data Processing**: Ingestion → Analysis → Reporting → Summary
- **Decision Making**: Information gathering → Analysis → Recommendations
- **Creative Workflows**: Brainstorming → Concept development → Refinement

## 🚀 Next Steps

1. **Try different topics**: Modify the research topics in `main()`
2. **Add more steps**: Extend the workflow with additional research phases
3. **Custom progress**: Implement domain-specific progress indicators
4. **Error handling**: Add retry logic and graceful degradation
5. **Persistence**: Save intermediate results between steps