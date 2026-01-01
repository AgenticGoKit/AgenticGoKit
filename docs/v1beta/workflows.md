# Workflows Guide

Connect multiple agents in different execution patterns to orchestrate complex tasks. This guide covers sequential, parallel, DAG, loop, and subworkflow composition.

---

## What You Get

- Sequential, parallel, DAG, and loop execution modes
- Automatic dependency management and execution ordering
- Stream workflow progress in real-time
- Reusable subworkflow composition
- **Shared memory across agents with automatic context querying**
- Error isolation and partial result tracking

---

## Execution Modes

AgenticGoKit provides 4 workflow execution patterns plus subworkflow composition:

### Sequential
Steps execute one after another, with each step receiving previous results. Use for data pipelines and multi-stage processing.

### Parallel
All steps run concurrently and results are collected. Use for independent tasks, parallel analysis, or when speed matters.

### DAG (Directed Acyclic Graph)
Steps execute based on explicit dependencies. Use for complex workflows with both parallel and sequential phases.

### Loop
Steps repeat until a condition is met. Use for iterative refinement, retry logic, or convergence-based processing.

### Subworkflow Composition
Nest workflows within workflows for modular, reusable design. Use for hierarchical task decomposition and workflow libraries.

---

## Sequential Workflow

Execute steps one after another, passing results forward. Perfect for data pipelines and multi-stage processing.

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create agents
    extract, _ := v1beta.NewChatAgent("Extractor", v1beta.WithLLM("openai", "gpt-4"))
    transform, _ := v1beta.NewChatAgent("Transformer", v1beta.WithLLM("openai", "gpt-4"))
    load, _ := v1beta.NewChatAgent("Loader", v1beta.WithLLM("openai", "gpt-4"))
    
    // Create workflow
    config := &v1beta.WorkflowConfig{
        Mode:    v1beta.Sequential,
        Timeout: 60 * time.Second,
    }
    
    workflow, err := v1beta.NewSequentialWorkflow(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Add steps in order
    workflow.AddStep(v1beta.WorkflowStep{Name: "extract", Agent: extract})
    workflow.AddStep(v1beta.WorkflowStep{Name: "transform", Agent: transform})
    workflow.AddStep(v1beta.WorkflowStep{Name: "load", Agent: load})
    
    // Execute
    result, err := workflow.Run(context.Background(), "Process data")
    if err != nil {
        log.Fatal(err)
    }
    
    // Access results
    for _, stepResult := range result.StepResults {
        if !stepResult.Skipped {
            log.Printf("%s: %s", stepResult.StepName, stepResult.Output)
        }
    }
}
```

---

## Parallel Workflow

Run multiple agents concurrently. Perfect for independent tasks and gathering multiple perspectives.

```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 90 * time.Second,
}

workflow, _ := v1beta.NewParallelWorkflow(config)

// Add steps - all run concurrently
tech, _ := v1beta.NewChatAgent("Tech", v1beta.WithLLM("openai", "gpt-4"))
biz, _ := v1beta.NewChatAgent("Biz", v1beta.WithLLM("openai", "gpt-4"))
legal, _ := v1beta.NewChatAgent("Legal", v1beta.WithLLM("openai", "gpt-4"))

workflow.AddStep(v1beta.WorkflowStep{Name: "technical", Agent: tech})
workflow.AddStep(v1beta.WorkflowStep{Name: "business", Agent: biz})
workflow.AddStep(v1beta.WorkflowStep{Name: "legal", Agent: legal})

result, _ := workflow.Run(context.Background(), "Analyze the product launch")

// All results available
for _, stepResult := range result.StepResults {
    log.Printf("%s: %s", stepResult.StepName, stepResult.Output)
}
```

### With Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 30 * time.Second,
}

workflow, _ := v1beta.NewParallelWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "task1", Agent: agent1})
workflow.AddStep(v1beta.WorkflowStep{Name: "task2", Agent: agent2})

result, err := workflow.Run(ctx, "Process tasks")
if err == context.DeadlineExceeded {
    log.Printf("Timeout: %d tasks completed", len(result.StepResults))
}
```

---

## DAG Workflow

Execute steps based on explicit dependencies. Steps with no dependencies run in parallel; dependent steps wait for prerequisites.

```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.DAG,
    Timeout: 120 * time.Second,
}

workflow, _ := v1beta.NewDAGWorkflow(config)

// Create agents
data, _ := v1beta.NewChatAgent("DataFetcher", v1beta.WithLLM("openai", "gpt-4"))
proc1, _ := v1beta.NewChatAgent("Processor1", v1beta.WithLLM("openai", "gpt-4"))
proc2, _ := v1beta.NewChatAgent("Processor2", v1beta.WithLLM("openai", "gpt-4"))
agg, _ := v1beta.NewChatAgent("Aggregator", v1beta.WithLLM("openai", "gpt-4"))

// Step 1: collect data (no dependencies)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "collect",
    Agent:        data,
    Dependencies: nil,
})

// Steps 2 & 3: process in parallel (depend on collect)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "process1",
    Agent:        proc1,
    Dependencies: []string{"collect"},
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "process2",
    Agent:        proc2,
    Dependencies: []string{"collect"},
})

// Step 4: aggregate (depends on both processors)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "aggregate",
    Agent:        agg,
    Dependencies: []string{"process1", "process2"},
})

result, _ := workflow.Run(context.Background(), "Collect and process data")
```

### Complex Example: E-commerce Order Processing

```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.DAG,
    Timeout: 180 * time.Second,
}

workflow, _ := v1beta.NewDAGWorkflow(config)

// Initial validation (no deps)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "validate",
    Agent:        validateAgent,
    Dependencies: nil,
})

// Parallel checks (depend on validation)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "check_inventory",
    Agent:        inventoryAgent,
    Dependencies: []string{"validate"},
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "check_payment",
    Agent:        paymentAgent,
    Dependencies: []string{"validate"},
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "check_fraud",
    Agent:        fraudAgent,
    Dependencies: []string{"validate"},
})

// Authorization (needs all checks)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "authorize",
    Agent:        authAgent,
    Dependencies: []string{"check_inventory", "check_payment", "check_fraud"},
})

// Parallel fulfillment (after authorization)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "reserve",
    Agent:        reserveAgent,
    Dependencies: []string{"authorize"},
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "notify",
    Agent:        notifyAgent,
    Dependencies: []string{"authorize"},
})

// Final confirmation (needs both fulfillment steps)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "confirm",
    Agent:        confirmAgent,
    Dependencies: []string{"reserve", "notify"},
})

result, _ := workflow.Run(context.Background(), "Process order")
```

---

## Loop Workflow

Execute steps iteratively until a condition is met. Perfect for iterative refinement and convergence-based processing.

```go
// Define loop condition
shouldContinue := func(ctx context.Context, iteration int, lastResult *v1beta.WorkflowResult) (bool, error) {
    // Stop after 5 iterations
    if iteration >= 5 {
        return false, nil
    }
    
    if lastResult == nil {
        return true, nil // Continue on first iteration
    }
    
    // Check quality score from result metadata
    if score, ok := lastResult.Metadata["quality_score"].(float64); ok {
        return score < 0.8, nil // Continue if quality below threshold
    }
    
    return true, nil
}

config := &v1beta.WorkflowConfig{
    Mode:          v1beta.Loop,
    Timeout:       300 * time.Second,
    MaxIterations: 5,
}

workflow, _ := v1beta.NewLoopWorkflowWithCondition(config, shouldContinue)

draft, _ := v1beta.NewChatAgent("Drafter", v1beta.WithLLM("openai", "gpt-4"))
critic, _ := v1beta.NewChatAgent("Critic", v1beta.WithLLM("openai", "gpt-4"))

workflow.AddStep(v1beta.WorkflowStep{Name: "draft", Agent: draft})
workflow.AddStep(v1beta.WorkflowStep{Name: "critique", Agent: critic})
workflow.AddStep(v1beta.WorkflowStep{Name: "refine", Agent: draft})

result, _ := workflow.Run(context.Background(), "Write essay on artificial intelligence")

// Get final result
log.Printf("Final output: %s", result.FinalOutput)
if result.IterationInfo != nil {
    log.Printf("Iterations: %d, Exit reason: %s", 
        result.IterationInfo.TotalIterations,
        result.IterationInfo.ExitReason)
}
```

### With Convergence Detection

```go
var previousOutput string

shouldContinue := func(ctx context.Context, iteration int, lastResult *v1beta.WorkflowResult) (bool, error) {
    if iteration >= 10 {
        return false, nil
    }
    
    if lastResult == nil {
        return true, nil
    }
    
    // Stop if output unchanged (converged)
    currentOutput := lastResult.FinalOutput
    if previousOutput != "" && currentOutput == previousOutput {
        return false, nil
    }
    
    previousOutput = currentOutput
    return true, nil
}

config := &v1beta.WorkflowConfig{
    Mode:          v1beta.Loop,
    Timeout:       600 * time.Second,
    MaxIterations: 10,
}

workflow, _ := v1beta.NewLoopWorkflowWithCondition(config, shouldContinue)
// ... add steps
result, _ := workflow.Run(context.Background(), "Optimize the algorithm")
```

### With Error Recovery

```go
shouldContinue := func(ctx context.Context, iteration int, lastResult *v1beta.WorkflowResult) (bool, error) {
    // Retry up to 3 times on failure
    if iteration >= 3 {
        return false, nil
    }
    
    if lastResult == nil {
        return true, nil
    }
    
    // Retry if last attempt failed
    if !lastResult.Success {
        return true, nil
    }
    
    return false, nil
}

config := &v1beta.WorkflowConfig{
    Mode:          v1beta.Loop,
    Timeout:       180 * time.Second,
    MaxIterations: 3,
}

workflow, _ := v1beta.NewLoopWorkflowWithCondition(config, shouldContinue)
// ... add steps
result, _ := workflow.Run(context.Background(), "Execute complex task")
```

---

## Shared Memory

Enable agents in a workflow to access the same memory store. Agents automatically query workflow memory for relevant context, enabling true multi-agent collaboration.

### How It Works

1. Workflow stores step inputs/outputs in shared memory
2. Memory is passed to agents via context
3. Agents automatically query shared memory when processing input
4. Relevant context is injected into prompts

### Basic Shared Memory

```go
package main

import (
    "context"
    "log"
    "time"
    
    _ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create shared memory
    sharedMemory, _ := v1beta.NewMemory(&v1beta.MemoryConfig{
        Enabled:  true,
        Provider: "chromem",
    })
    
    // Create agents
    learner, _ := v1beta.NewChatAgent("Learner", v1beta.WithLLM("openai", "gpt-4"))
    answerer, _ := v1beta.NewChatAgent("Answerer", v1beta.WithLLM("openai", "gpt-4"))
    
    // Create workflow
    workflow, _ := v1beta.NewSequentialWorkflow(&v1beta.WorkflowConfig{
        Mode:    v1beta.Sequential,
        Timeout: 60 * time.Second,
    })
    
    // Attach shared memory to workflow
    workflow.SetMemory(sharedMemory)
    
    // Add steps
    workflow.AddStep(v1beta.WorkflowStep{Name: "learn", Agent: learner})
    workflow.AddStep(v1beta.WorkflowStep{Name: "answer", Agent: answerer})
    
    // Agent 1 learns facts, Agent 2 automatically has access via shared memory
    result, _ := workflow.Run(context.Background(), "Company data...")
    
    log.Printf("Result: %s", result.FinalOutput)
}
```

### With Custom Question Transform

Pass different input to Agent 2 while maintaining shared memory access:

```go
workflow.AddStep(v1beta.WorkflowStep{
    Name:  "learn",
    Agent: learner,
})

workflow.AddStep(v1beta.WorkflowStep{
    Name:  "answer",
    Agent: answerer,
    Transform: func(_ string) string {
        // Agent 2 gets only the question,
        // but automatically accesses Agent 1's learned facts via shared memory
        return "What company was founded in 2020?"
    },
})

result, _ := workflow.Run(ctx, "Company: TechStart Inc\nFounded: 2020\nFocus: AI tools")
// Agent 2 will answer "TechStart Inc" using facts from shared memory
```

### Parallel Workflow with Shared Memory

All agents can access shared knowledge:

```go
sharedMemory, _ := v1beta.NewMemory(&v1beta.MemoryConfig{
    Provider: "chromem",
})

workflow, _ := v1beta.NewParallelWorkflow(&v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 90 * time.Second,
})

workflow.SetMemory(sharedMemory)

// All analysts can access shared research data
workflow.AddStep(v1beta.WorkflowStep{Name: "financial", Agent: financialAnalyst})
workflow.AddStep(v1beta.WorkflowStep{Name: "technical", Agent: technicalAnalyst})
workflow.AddStep(v1beta.WorkflowStep{Name: "market", Agent: marketAnalyst})

result, _ := workflow.Run(ctx, "Analyze TechStart Inc")
// All analysts query shared memory for relevant context
```

### DAG Workflow with Shared Memory

Complex dependencies with shared context:

```go
workflow, _ := v1beta.NewDAGWorkflow(&v1beta.WorkflowConfig{
    Mode:    v1beta.DAG,
    Timeout: 120 * time.Second,
})

workflow.SetMemory(sharedMemory)

// Research agent gathers data
workflow.AddStep(v1beta.WorkflowStep{
    Name:  "research",
    Agent: researcher,
})

// Both analysis agents depend on research and can access shared memory
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "analyze_a",
    Agent:        analyzerA,
    Dependencies: []string{"research"},
})

workflow.AddStep(v1beta.WorkflowStep{
    Name:         "analyze_b",
    Agent:        analyzerB,
    Dependencies: []string{"research"},
})

// Final synthesis accesses all previous context via shared memory
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "synthesize",
    Agent:        synthesizer,
    Dependencies: []string{"analyze_a", "analyze_b"},
})
```

### Advanced: Direct Memory Access

Agents automatically query workflow memory, but you can also access it directly:

```go
// In your code
if workflowMem := v1beta.GetWorkflowMemory(ctx); workflowMem != nil {
    results, _ := workflowMem.Query(ctx, "company founded in 2020", 
        v1beta.WithLimit(5),
        v1beta.WithScoreThreshold(0.3))
    
    for _, result := range results {
        log.Printf("Found: %s (score: %.2f)", result.Content, result.Score)
    }
}
```

### Memory Providers

Choose the right provider for your use case:

**chromem** (default): Embedded vector database
```go
sharedMemory, _ := v1beta.NewMemory(&v1beta.MemoryConfig{
    Provider: "chromem",  // No external dependencies
})
```

**pgvector**: PostgreSQL for production
```go
sharedMemory, _ := v1beta.NewMemory(&v1beta.MemoryConfig{
    Provider:   "pgvector",
    Connection: "postgresql://user:pass@localhost:5432/db",
})
```

### When to Use Shared Memory

**Use shared memory when:**
- Agents need to reference previous steps' outputs
- Building knowledge progressively across steps
- Multiple agents collaborate on the same problem
- Context needs to persist across workflow runs (with persistent providers)

**Skip shared memory when:**
- Steps are completely independent
- Each step has all needed context in input
- Workflow is simple sequential passthrough

---

## Subworkflow Composition

Nest workflows as agents for modular, reusable design.

### Basic Subworkflow

```go
// Create reusable research workflow
researchConfig := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 90 * time.Second,
}

researchWorkflow, _ := v1beta.NewSequentialWorkflow(researchConfig)
researchWorkflow.AddStep(v1beta.WorkflowStep{Name: "search", Agent: searchAgent})
researchWorkflow.AddStep(v1beta.WorkflowStep{Name: "analyze", Agent: analyzerAgent})
researchWorkflow.AddStep(v1beta.WorkflowStep{Name: "summarize", Agent: summarizerAgent})

// Wrap workflow as agent
researchAsAgent := v1beta.NewSubWorkflowAgent("research", researchWorkflow)

// Use in main workflow
mainConfig := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 180 * time.Second,
}

mainWorkflow, _ := v1beta.NewSequentialWorkflow(mainConfig)
mainWorkflow.AddStep(v1beta.WorkflowStep{Name: "topic1", Agent: researchAsAgent})
mainWorkflow.AddStep(v1beta.WorkflowStep{Name: "topic2", Agent: researchAsAgent})
mainWorkflow.AddStep(v1beta.WorkflowStep{Name: "compare", Agent: compareAgent})

result, _ := mainWorkflow.Run(context.Background(), "Research AI trends")
```

### Nested Levels

```go
// Level 3: Validation (parallel)
validationWorkflow, _ := v1beta.NewParallelWorkflow(&v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 30 * time.Second,
})
validationWorkflow.AddStep(v1beta.WorkflowStep{Name: "format", Agent: formatValidator})
validationWorkflow.AddStep(v1beta.WorkflowStep{Name: "integrity", Agent: integrityValidator})
validationAsAgent := v1beta.NewSubWorkflowAgent("validation", validationWorkflow)

// Level 2: Processing (sequential, uses validation)
processingWorkflow, _ := v1beta.NewSequentialWorkflow(&v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 60 * time.Second,
})
processingWorkflow.AddStep(v1beta.WorkflowStep{Name: "validate", Agent: validationAsAgent})
processingWorkflow.AddStep(v1beta.WorkflowStep{Name: "transform", Agent: transformAgent})
processingAsAgent := v1beta.NewSubWorkflowAgent("processing", processingWorkflow)

// Level 1: Main ETL (uses processing)
etlWorkflow, _ := v1beta.NewSequentialWorkflow(&v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 120 * time.Second,
})
etlWorkflow.AddStep(v1beta.WorkflowStep{Name: "extract", Agent: extractAgent})
etlWorkflow.AddStep(v1beta.WorkflowStep{Name: "process", Agent: processingAsAgent})
etlWorkflow.AddStep(v1beta.WorkflowStep{Name: "load", Agent: loadAgent})

result, _ := etlWorkflow.Run(context.Background(), "Process dataset")
```

### Mixing Workflow Types

```go
// Parallel analysis subworkflow
analysisWorkflow, _ := v1beta.NewParallelWorkflow(&v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 60 * time.Second,
})
analysisWorkflow.AddStep(v1beta.WorkflowStep{Name: "sentiment", Agent: sentimentAgent})
analysisWorkflow.AddStep(v1beta.WorkflowStep{Name: "entities", Agent: entityAgent})
analysisAsAgent := v1beta.NewSubWorkflowAgent("analysis", analysisWorkflow)

// Main DAG workflow
mainWorkflow, _ := v1beta.NewDAGWorkflow(&v1beta.WorkflowConfig{
    Mode:    v1beta.DAG,
    Timeout: 120 * time.Second,
})
mainWorkflow.AddStep(v1beta.WorkflowStep{
    Name:         "fetch",
    Agent:        fetchAgent,
    Dependencies: nil,
})
mainWorkflow.AddStep(v1beta.WorkflowStep{
    Name:         "analyze",
    Agent:        analysisAsAgent,
    Dependencies: []string{"fetch"},
})
mainWorkflow.AddStep(v1beta.WorkflowStep{
    Name:         "generate",
    Agent:        generateAgent,
    Dependencies: []string{"analyze"},
})

result, _ := mainWorkflow.Run(context.Background(), "Fetch and analyze content")
```

---

## Workflow Streaming

Monitor workflow execution in real-time with streaming chunks.

```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 120 * time.Second,
}

workflow, _ := v1beta.NewSequentialWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "step1", Agent: agent1})
workflow.AddStep(v1beta.WorkflowStep{Name: "step2", Agent: agent2})

stream, err := workflow.RunStream(context.Background(), "First task")
if err != nil {
    log.Fatal(err)
}

for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeMetadata:
        if step, ok := chunk.Metadata["step_name"].(string); ok {
            log.Printf("→ Executing: %s", step)
        }
    case v1beta.ChunkTypeDelta:
        fmt.Print(chunk.Delta)
    case v1beta.ChunkTypeDone:
        log.Println("✓ Workflow complete")
    }
}
```

---

## Best Practices

**Choose the right mode:**
- Sequential: dependent steps, data pipelines
- Parallel: independent tasks, speed important
- DAG: complex dependencies
- Loop: iterative refinement, error recovery
- Subworkflow: reusable logic, hierarchical design

**Error handling:**
```go
result, err := workflow.Run(ctx, "input")
if err != nil {
    // Check partial results
    if result != nil {
        for _, stepResult := range result.StepResults {
            if !stepResult.Success {
                log.Printf("Failed step: %s - %s", stepResult.StepName, stepResult.Error)
            }
        }
    }
}
```

**Set appropriate timeouts:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := workflow.Run(ctx, "Task")
```

**Reuse subworkflows:**
```go
// Define validation once, use in multiple workflows
validationConfig := &v1beta.WorkflowConfig{Mode: v1beta.Parallel, Timeout: 30*time.Second}
validationWorkflow, _ := v1beta.NewParallelWorkflow(validationConfig)
// ... add steps
validationAsAgent := v1beta.NewSubWorkflowAgent("validation", validationWorkflow)

// Use in multiple pipelines
pipeline1, _ := v1beta.NewSequentialWorkflow(config1)
pipeline1.AddStep(v1beta.WorkflowStep{Name: "validate", Agent: validationAsAgent})

pipeline2, _ := v1beta.NewSequentialWorkflow(config2)
pipeline2.AddStep(v1beta.WorkflowStep{Name: "validate", Agent: validationAsAgent})
```

---

## Troubleshooting

**DAG circular dependency**: Check dependencies don't form cycles
```go
// Bad: A depends on B, B depends on A
// Good: Ensure acyclic relationships (A → B → C)
```

**Step result access**: Results available in WorkflowResult.StepResults
```go
result, _ := workflow.Run(ctx, "Input")
for _, stepResult := range result.StepResults {
    log.Printf("Step %s: %s", stepResult.StepName, stepResult.Output)
}
```

**Parallel workflow hangs**: Set context timeouts to prevent deadlocks
```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()
workflow.Run(ctx, "input")
```

**Loop infinite execution**: Always set MaxIterations safeguard
```go
config := &v1beta.WorkflowConfig{
    Mode:          v1beta.Loop,
    MaxIterations: 10,  // Required
}
```

---

## Examples & Further Reading

See full workflow examples:
- [Sequential Workflow Example](./examples/workflow-sequential.md)
- [Parallel Workflow Example](./examples/workflow-parallel.md)
- [DAG Workflow Example](./examples/workflow-dag.md)
- [Loop Workflow Example](./examples/workflow-loop.md)
- [Subworkflow Composition](./examples/subworkflow-composition.md)

---

## Related Topics

- [Core Concepts](./core-concepts.md) - Understanding agents
- [Streaming](./streaming.md) - Real-time workflow streaming
- [Custom Handlers](./custom-handlers.md) - Advanced agent logic

---

**Next steps?** Continue to [Memory & RAG Guide](./memory-and-rag.md) →
