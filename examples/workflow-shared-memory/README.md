# Workflow Shared Memory Demo - Beginner Friendly

This is the **simplest way to learn** how workflow-level shared memory works in v1beta.

Perfect for:
- ✅ Beginners new to agentic AI
- ✅ Learning memory patterns in multi-agent systems
- ✅ Understanding agent collaboration

## The Simple Idea

Imagine two people working together:
1. **Person A** (Researcher): Reads information and takes notes
2. **Person B** (Assistant): Uses Person A's notes to answer questions

They share the same notebook! That's workflow-level memory.

```
Information
    ↓
Agent 1: Learns & stores facts in shared memory
    ↓
Agent 2: Reads facts from memory & answers questions
```

## What This Example Does

1. **Agent 1 (Information Learner)** - Receives company information
   - Extracts key facts
   - Stores them in shared memory

2. **Agent 2 (Question Answerer)** - Answers questions
   - Reads facts from shared memory (learned by Agent 1)
   - Uses that knowledge to answer

**The Magic:** Agent 2 knows things it never saw directly, because it can access what Agent 1 learned!

## Running the Demo

### Prerequisites
```bash
# Install Ollama from ollama.com
ollama pull gemma3:1b

# Then navigate to this directory
cd examples/workflow-shared-memory
```

### Run It
```bash
go run main.go
```

### What You'll See
```
📊 Input: Company information
    ↓
Agent 1: Learns and extracts key facts
    ↓
[Facts stored in shared memory]
    ↓
Agent 2: Answers questions using those facts
    ↓
📊 Output: Answer based on learned facts
```

## How It Works (Step by Step)

### 1. Create Shared Memory
```go
sharedMemory, err := vnext.NewMemory(&vnext.MemoryConfig{
    Enabled:  true,
    Provider: "chromem",  // Embedded vector database
})
```

### 2. Create Agents
```go
agent1 := vnext.NewBuilder("agent1").WithConfig(...).Build()
agent2 := vnext.NewBuilder("agent2").WithConfig(...).Build()
```

### 3. Create Workflow and Attach Memory
```go
workflow, _ := vnext.NewSequentialWorkflow(config)
workflow.SetMemory(sharedMemory)  // ← This is the KEY line!
```

### 4. Add Agents to Workflow
```go
workflow.AddStep(WorkflowStep{Name: "step1", Agent: agent1})
workflow.AddStep(WorkflowStep{Name: "step2", Agent: agent2})
```

### 5. Run
```go
result, _ := workflow.Run(ctx, inputData)
```

**That's it!** Memory sharing is automatic.

## Key Learning Points

| Concept | What It Means |
|---------|---------------|
| **Shared Memory** | One memory store that all agents can access |
| **Workflow Memory** | Persists across all steps in a workflow |
| **Automatic Storage** | Step outputs automatically saved to memory |
| **chromem** | Embedded vector database - great for demos |
| **SetMemory()** | The line that enables memory sharing |

## Without Memory vs With Memory

### WITHOUT Shared Memory
```
Agent 1 learns: "Company has 50 employees"
↓
Agent 2 runs with: "Answer question about the company"
↓
Agent 2: "I don't know how many employees..." ❌
```

### WITH Shared Memory
```
Agent 1 learns: "Company has 50 employees" → Stores in memory
↓
Agent 2 runs with: "Answer question about the company"
↓
Agent 2 reads memory: "Company has 50 employees" → Uses it!
↓
Agent 2: "The company has 50 employees" ✅
```

## Try Modifying

1. **Change Agent 1's job:**
   - Extract different facts
   - Summarize instead of extract

2. **Change Agent 2's job:**
   - Ask different questions
   - Use facts differently

3. **Add more agents:**
   - Agent 3 that validates facts
   - Agent 4 that creates a summary

4. **Add more steps:**
   ```go
   workflow.AddStep(WorkflowStep{Name: "step3", Agent: agent3})
   workflow.AddStep(WorkflowStep{Name: "step4", Agent: agent4})
   ```

## Real-World Examples

| Scenario | Agent 1 | Memory | Agent 2 |
|----------|---------|--------|---------|
| **Research & Writing** | Researcher finds facts | Stores findings | Writer creates article |
| **Customer Service** | FAQ Lookup | Stores answers | Response Generator |
| **Data Processing** | Data Extractor | Stores tables | Data Analyzer |
| **Content Creation** | Outline Creator | Stores structure | Section Writer |

## Memory Providers

### chromem (Used in this demo)
- ✅ Embedded (no setup needed)
- ✅ Perfect for learning
- ✅ Works on any machine
- ❌ Resets when program ends

### pgvector (Production)
- ✅ Persistent (saves to database)
- ✅ Scalable
- ❌ Needs PostgreSQL setup

## Troubleshooting

**Q: Agent 2 doesn't know what Agent 1 learned?**
```
A: Make sure you called: workflow.SetMemory(sharedMemory)
```

**Q: How do I use PostgreSQL instead of chromem?**
```go
sharedMemory, _ := vnext.NewMemory(&vnext.MemoryConfig{
    Provider:   "pgvector",
    Connection: "postgresql://user:pass@localhost/dbname",
})
```

**Q: Can I have more than 2 agents?**
```
A: Yes! Add more with: workflow.AddStep(...)
```

## Next Steps

After mastering this:
1. ✅ Read: [Core Concepts](../../docs/v1beta/core-concepts.md)
2. ✅ Try: [Sequential Workflow Demo](../sequential-workflow-demo/)
3. ✅ Learn: [Memory & RAG](../../docs/v1beta/memory-and-rag.md)
4. ✅ Build: Your own multi-agent system!

---

**Happy Learning! 🚀**
