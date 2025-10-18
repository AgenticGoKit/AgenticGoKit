# Researcher & Reporter - Sequential Workflow Example

> **Perfect for blog posts!** This example shows how simple it is to create a multi-agent workflow with vNext APIs.

Build a **research-to-report pipeline** in just **60 lines of code**:
1. **Researcher Agent** - Gathers information
2. **Reporter Agent** - Creates structured reports

## 🎯 Why This Example?

This demonstrates how **easy** it is to:
- ✅ Create multiple specialized agents
- ✅ Chain them in a sequential workflow
- ✅ Transform data between steps
- ✅ Get professional results

## ⚡ Quick Start

```bash
# 1. Install Ollama and pull the model
ollama pull gemma3:1b

# 2. Run the example
cd examples/vnext/researcher-reporter
go run main.go
```

## 🏗️ How It Works (4 Simple Steps)

```go
// Step 1: Create specialized agents
researcher := createAgent("researcher", "Gather key facts...", 0.7)
reporter := createAgent("reporter", "Create a report...", 0.5)

// Step 2: Create workflow
workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{...})

// Step 3: Add agents as steps
workflow.AddStep(vnext.WorkflowStep{Name: "research", Agent: researcher})
workflow.AddStep(vnext.WorkflowStep{Name: "report", Agent: reporter})

// Step 4: Run!
result, _ := workflow.Run(ctx, "What is Kubernetes?")
```

That's it! **4 steps, ~60 lines of code.**

## 📖 The Complete Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core/vnext"
    _ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"
)

func main() {
    ctx := context.Background()
    
    // Step 1: Create two agents
    researcher := createAgent("researcher", "Gather key facts...", 0.7)
    reporter := createAgent("reporter", "Create a report...", 0.5)
    
    // Step 2: Create workflow
    workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
        Mode:    vnext.Sequential,
        Timeout: 120 * time.Second,
    })
    
    // Step 3: Add steps
    workflow.AddStep(vnext.WorkflowStep{Name: "research", Agent: researcher})
    workflow.AddStep(vnext.WorkflowStep{
        Name: "report", 
        Agent: reporter,
        Transform: func(input string) string {
            return "Create a report from:\n" + input
        },
    })
    
    // Step 4: Run!
    workflow.Initialize(ctx)
    defer workflow.Shutdown(ctx)
    
    result, _ := workflow.Run(ctx, "What is Kubernetes?")
    fmt.Println(result.StepResults[1].Output) // Final report
}

func createAgent(name, prompt string, temp float64) vnext.Agent {
    agent, _ := vnext.NewBuilder(name).WithConfig(&vnext.Config{
        Name:         name,
        SystemPrompt: prompt,
        LLM:          vnext.LLMConfig{Provider: "ollama", Model: "gemma3:1b", Temperature: temp},
        Timeout:      60 * time.Second,
    }).Build()
    agent.Initialize(context.Background())
    return agent
}
```

**That's the entire example!** Just 60 lines.

## 📊 Sample Output

```
📋 Topic: What is Kubernetes and why is it important?

🔍 RESEARCH FINDINGS:
Kubernetes is an open-source container orchestration platform...
[detailed research findings]

📄 FINAL REPORT:
Summary:
Kubernetes is a powerful container orchestration system that 
automates deployment, scaling, and management of applications.

Key Points:
• Automates container deployment and scaling
• Provides self-healing capabilities
• Enables microservices architecture
• Industry standard for cloud-native apps

Conclusion:
Kubernetes has become essential for modern cloud infrastructure,
enabling teams to build scalable, resilient applications.

✅ Completed in 4.2s
```

## 🔑 Key Features

| Feature | Description |
|---------|-------------|
| **Sequential Workflow** | Steps execute in order, output flows from one to next |
| **Transform Function** | Modify data between steps for better prompts |
| **Specialized Agents** | Each agent has its own role and temperature |
| **Simple API** | Just 4 steps to create a working pipeline |

## 💡 Why This Matters

Creating multi-agent workflows used to require complex orchestration. With vNext:

```go
// Old way: Complex orchestration, manual state management
// 100+ lines of code, error-prone

// New way: Simple, declarative
workflow.AddStep(vnext.WorkflowStep{Name: "step1", Agent: agent1})
workflow.AddStep(vnext.WorkflowStep{Name: "step2", Agent: agent2})
result := workflow.Run(ctx, input)
```

**Result**: Professional multi-agent systems in minutes, not hours.

## 🎨 Extend It (Optional)

Want to make it even better? Add more steps:

```go
// Add an editor agent
editor := createAgent("editor", "Proofread and improve the report", 0.3)
workflow.AddStep(vnext.WorkflowStep{Name: "edit", Agent: editor})

// Add memory to agents
researcher := vnext.NewBuilder("researcher").
    WithConfig(config).
    WithMemory().  // Remembers past queries!
    Build()

// Use different models per agent
researcher.LLM.Model = "llama3.1:8b"  // Larger model for research
reporter.LLM.Model = "gemma3:1b"      // Faster model for reporting
```

## 🎯 Real-World Use Cases

This simple pattern powers:
- 📝 **Content Pipelines**: Research → Write → Edit → Publish
- 📊 **Data Analysis**: Collect → Analyze → Report → Visualize  
- 🔍 **Due Diligence**: Gather → Verify → Summarize → Recommend
- 🎨 **Creative Work**: Ideate → Draft → Refine → Finalize

## 🐛 Quick Troubleshooting

| Problem | Solution |
|---------|----------|
| `connection refused` | Start Ollama: `ollama serve` |
| `model not found` | Pull model: `ollama pull gemma3:1b` |
| Workflow timeout | Increase `Timeout` in config |

## � Learn More

- [Other vNext Examples](../)
- [Full Workflow Documentation](../../../docs/guides/vnext-workflows.md)
- [Multi-Agent Patterns](../../../docs/guides/multi-agent-patterns.md)

---

**Ready to build your own workflows?** This example shows how easy it is! 🚀
