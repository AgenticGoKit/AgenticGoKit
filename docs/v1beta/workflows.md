# Workflows Guide

Multi-agent workflows enable complex task orchestration by connecting multiple agents in different execution patterns. This guide covers all workflow types and composition strategies.

---

## 🎯 Why Workflows?

Workflows solve several key challenges:

- **Complex tasks** - Break down large problems into specialized steps
- **Parallel processing** - Run multiple agents simultaneously
- **Conditional logic** - Execute different paths based on results
- **Dependency management** - Handle step dependencies automatically
- **Error handling** - Isolate failures and enable recovery
- **Progress tracking** - Monitor execution through streaming
- **Reusability** - Compose workflows as building blocks

---

## 📊 Workflow Types

AgenticGoKit provides **4 workflow execution patterns** plus **subworkflow composition**:

### 1. Sequential Workflow
Execute agents one after another, passing results forward.

```
Agent 1 → Agent 2 → Agent 3 → Result
```

**Use when:**
- Steps must execute in order
- Each step depends on previous results
- Building data pipelines
- Multi-stage processing

### 2. Parallel Workflow
Execute multiple agents simultaneously, collect all results.

```
       → Agent 1 →
Start → Agent 2 → Collect → Result
       → Agent 3 →
```

**Use when:**
- Independent tasks can run concurrently
- Gathering multiple perspectives
- Parallel data processing
- Speed is critical

### 3. DAG Workflow (Directed Acyclic Graph)
Execute agents based on explicit dependencies.

```
    → Agent 2 →
Agent 1 →         → Agent 4 → Result
    → Agent 3 →
```

**Use when:**
- Complex dependency relationships
- Some parallel, some sequential steps
- Conditional execution paths
- Advanced orchestration

### 4. Loop Workflow
Execute agents iteratively until a condition is met.

```
Agent 1 → Agent 2 → Check Condition
    ↑                      ↓
    └──────── Continue ────┘
                ↓
            Complete
```

**Use when:**
- Iterative refinement needed
- Quality thresholds must be met
- Retry logic required
- Convergence-based tasks

### 5. Subworkflow Composition
Nest workflows within workflows for modular design.

```
Main Workflow:
  Step 1 → Subworkflow → Step 3
              ↓
          [Sequential:
           Sub-A → Sub-B]
```

**Use when:**
- Reusing workflow logic
- Hierarchical task decomposition
- Encapsulating complex steps
- Building workflow libraries

---

## 🔄 Sequential Workflow

### Basic Sequential Workflow

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create specialized agents
    researchAgent, _ := v1beta.NewBuilder("Researcher").
        WithPreset(v1beta.ResearchAgent).
        Build()
    
    analyzerAgent, _ := v1beta.NewBuilder("Analyzer").
        WithPreset(v1beta.ChatAgent).
        WithSystemPrompt("You analyze and summarize research data").
        Build()
    
    writerAgent, _ := v1beta.NewBuilder("Writer").
        WithPreset(v1beta.ChatAgent).
        WithSystemPrompt("You write clear, concise summaries").
        Build()
    
    // Create sequential workflow
    config := &v1beta.WorkflowConfig{
        Mode:    v1beta.Sequential,
        Timeout: 60 * time.Second,
    }
    
    workflow, err := v1beta.NewSequentialWorkflow(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Add steps
    workflow.AddStep(v1beta.WorkflowStep{
        Name:  "research",
        Agent: researchAgent,
    })
    workflow.AddStep(v1beta.WorkflowStep{
        Name:  "analyze",
        Agent: analyzerAgent,
    })
    workflow.AddStep(v1beta.WorkflowStep{
        Name:  "write",
        Agent: writerAgent,
    })
    
    // Execute workflow
    result, err := workflow.Run(context.Background(), "Research quantum computing")
    if err != nil {
        log.Fatal(err)
    }
    
    // Access final result
    fmt.Println("Final Output:", result.FinalOutput)
    fmt.Println("Duration:", result.Duration)
}
```

### Sequential Workflow with Step Results

Access results from previous steps:

```go
// Step results are available in WorkflowResult
result, _ := workflow.Run(context.Background(), "Initial input")

for _, stepResult := range result.StepResults {
    fmt.Printf("Step %s: %s\n", stepResult.StepName, stepResult.Output)
}
```

### Sequential with Multiple Runs

```go
// Build workflow once
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 60 * time.Second,
}

workflow, _ := v1beta.NewSequentialWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "extract", Agent: extractAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "transform", Agent: transformAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "load", Agent: loadAgent})

// Run with different inputs
result1, _ := workflow.Run(context.Background(), "Process dataset1.csv")
result2, _ := workflow.Run(context.Background(), "Process dataset2.csv")
```

---

## ⚡ Parallel Workflow

### Basic Parallel Workflow

```go
// Create specialized agents
techAgent, _ := v1beta.QuickChatAgent("gpt-4")
bizAgent, _ := v1beta.QuickChatAgent("gpt-4")
legalAgent, _ := v1beta.QuickChatAgent("gpt-4")

// Create parallel workflow
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 90 * time.Second,
}

workflow, err := v1beta.NewParallelWorkflow(config)
if err != nil {
    log.Fatal(err)
}

// Add parallel steps
workflow.AddStep(v1beta.WorkflowStep{Name: "technical", Agent: techAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "business", Agent: bizAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "legal", Agent: legalAgent})

// Execute all agents concurrently
result, err := workflow.Run(context.Background(), "Analyze the product launch")
if err != nil {
    log.Fatal(err)
}

// Access results
for _, stepResult := range result.StepResults {
    fmt.Printf("%s: %s\n", stepResult.StepName, stepResult.Output)
}
```

### Parallel with Aggregation

```go
// Analysis agents
agents := []v1beta.Agent{analyst1, analyst2, analyst3}

// Create parallel workflow
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 120 * time.Second,
}

workflow, _ := v1beta.NewParallelWorkflow(config)

// Add steps
for i, agent := range agents {
    workflow.AddStep(v1beta.WorkflowStep{
        Name:  fmt.Sprintf("analyst_%d", i),
        Agent: agent,
    })
}

result, _ := workflow.Run(context.Background(), "Analyze the product launch strategy")

// Aggregate all analyses
var allAnalyses []string
for _, stepResult := range result.StepResults {
    allAnalyses = append(allAnalyses, stepResult.Output)
}

// Use aggregator agent to synthesize
aggregator, _ := v1beta.QuickChatAgent("gpt-4")
final, _ := aggregator.Run(
    context.Background(),
    fmt.Sprintf("Synthesize these analyses: %v", allAnalyses),
)
```

### Parallel with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 30 * time.Second,
}

workflow, _ := v1beta.NewParallelWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "slow_task", Agent: slowAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "fast_task", Agent: fastAgent})

result, err := workflow.Run(ctx, "Analyze data")
if err == context.DeadlineExceeded {
    // Handle timeout - may have partial results
    fmt.Println("Timeout! Completed:", len(result.StepResults), "tasks")
}
```

---

## 🕸️ DAG Workflow

### Basic DAG Workflow

```go
// Create agents
dataAgent, _ := v1beta.QuickChatAgent("gpt-4")
processor1, _ := v1beta.QuickChatAgent("gpt-4")
processor2, _ := v1beta.QuickChatAgent("gpt-4")
aggregator, _ := v1beta.QuickChatAgent("gpt-4")

// Define DAG structure
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.DAG,
    Timeout: 120 * time.Second,
}

workflow, err := v1beta.NewDAGWorkflow(config)
if err != nil {
    log.Fatal(err)
}

// Step 1: Collect data (no dependencies)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "collect",
    Agent:        dataAgent,
    Dependencies: nil,
})

// Steps 2 & 3: Process in parallel (both depend on collect)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "process1",
    Agent:        processor1,
    Dependencies: []string{"collect"},
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "process2",
    Agent:        processor2,
    Dependencies: []string{"collect"},
})

// Step 4: Aggregate (depends on both processors)
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "aggregate",
    Agent:        aggregator,
    Dependencies: []string{"process1", "process2"},
})

// Execute - automatically handles dependencies
result, err := workflow.Run(context.Background(), "Collect data from sources")
```

### Complex DAG Example

```go
// E-commerce order processing pipeline
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

result, _ := workflow.Run(context.Background(), "Process order #12345")
```

### DAG Execution Flow

```
validate
    ↓
    ├─→ check_inventory ─┐
    ├─→ check_payment ───┼─→ authorize
    └─→ check_fraud ─────┘       ↓
                               ├─→ reserve ─┐
                               └─→ notify ──┼─→ confirm
                                            ┘
```

---

## 🔁 Loop Workflow

### Basic Loop Workflow

```go
// Create agents for iterative refinement
drafterAgent, _ := v1beta.QuickChatAgent("gpt-4")
criticAgent, _ := v1beta.QuickChatAgent("gpt-4")

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

// Create loop workflow
config := &v1beta.WorkflowConfig{
    Mode:          v1beta.Loop,
    Timeout:       300 * time.Second,
    MaxIterations: 5,
}

workflow, err := v1beta.NewLoopWorkflowWithCondition(config, shouldContinue)
if err != nil {
    log.Fatal(err)
}

workflow.AddStep(v1beta.WorkflowStep{Name: "draft", Agent: drafterAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "critique", Agent: criticAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "refine", Agent: drafterAgent})

// Execute loop
result, err := workflow.Run(context.Background(), "Write essay on artificial intelligence")
if err != nil {
    log.Fatal(err)
}

// Get final result
fmt.Println("Final output:", result.FinalOutput)
if result.IterationInfo != nil {
    fmt.Println("Iterations:", result.IterationInfo.TotalIterations)
    fmt.Println("Exit reason:", result.IterationInfo.ExitReason)
}
```

### Loop with Convergence Detection

```go
var previousOutput string

shouldContinue := func(ctx context.Context, iteration int, lastResult *v1beta.WorkflowResult) (bool, error) {
    // Maximum iterations safeguard
    if iteration >= 10 {
        return false, nil
    }
    
    if lastResult == nil {
        return true, nil
    }
    
    // Check if improvement is minimal (convergence)
    currentOutput := lastResult.FinalOutput
    
    if previousOutput != "" {
        // Simple similarity check
        if currentOutput == previousOutput {
            return false, nil // Converged - no change
        }
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
workflow.AddStep(v1beta.WorkflowStep{Name: "analyze", Agent: analyzerAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "optimize", Agent: optimizerAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "process", Agent: processorAgent})

result, _ := workflow.Run(context.Background(), "Optimize the algorithm")
```

### Loop with Error Recovery

```go
shouldContinue := func(ctx context.Context, iteration int, lastResult *v1beta.WorkflowResult) (bool, error) {
    // Retry up to 3 times on failure
    if iteration >= 3 {
        return false, nil
    }
    
    if lastResult == nil {
        return true, nil
    }
    
    // Check if last attempt failed
    if !lastResult.Success {
        return true, nil // Retry on failure
    }
    
    // Check success criteria from metadata
    if validated, ok := lastResult.Metadata["validated"].(bool); ok {
        return !validated, nil // Continue if not validated
    }
    
    return false, nil
}

config := &v1beta.WorkflowConfig{
    Mode:          v1beta.Loop,
    Timeout:       180 * time.Second,
    MaxIterations: 3,
}

workflow, _ := v1beta.NewLoopWorkflowWithCondition(config, shouldContinue)
workflow.AddStep(v1beta.WorkflowStep{Name: "attempt", Agent: taskAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "validate", Agent: validatorAgent})

result, _ := workflow.Run(context.Background(), "Execute complex task")
```

---

## 🎨 Subworkflow Composition

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

// Wrap workflow as an agent using SubWorkflowAgent
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

### Nested Subworkflows

```go
// Level 3: Data validation subworkflow
validationConfig := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 30 * time.Second,
}

validationWorkflow, _ := v1beta.NewParallelWorkflow(validationConfig)
validationWorkflow.AddStep(v1beta.WorkflowStep{Name: "format", Agent: formatValidator})
validationWorkflow.AddStep(v1beta.WorkflowStep{Name: "integrity", Agent: integrityValidator})
validationWorkflow.AddStep(v1beta.WorkflowStep{Name: "completeness", Agent: completenessValidator})

// Wrap as agent
validationAsAgent := v1beta.NewSubWorkflowAgent("validation", validationWorkflow)

// Level 2: Data processing subworkflow
processingConfig := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 60 * time.Second,
}

processingWorkflow, _ := v1beta.NewSequentialWorkflow(processingConfig)
processingWorkflow.AddStep(v1beta.WorkflowStep{Name: "validate", Agent: validationAsAgent})
processingWorkflow.AddStep(v1beta.WorkflowStep{Name: "transform", Agent: transformAgent})
processingWorkflow.AddStep(v1beta.WorkflowStep{Name: "enrich", Agent: enrichAgent})

// Wrap as agent
processingAsAgent := v1beta.NewSubWorkflowAgent("processing", processingWorkflow)

// Level 1: Main ETL workflow
etlConfig := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 120 * time.Second,
}

etlWorkflow, _ := v1beta.NewSequentialWorkflow(etlConfig)
etlWorkflow.AddStep(v1beta.WorkflowStep{Name: "extract", Agent: extractAgent})
etlWorkflow.AddStep(v1beta.WorkflowStep{Name: "process", Agent: processingAsAgent})
etlWorkflow.AddStep(v1beta.WorkflowStep{Name: "load", Agent: loadAgent})

result, _ := etlWorkflow.Run(context.Background(), "Process dataset")
```

### Subworkflow with Different Types

```go
// Parallel analysis subworkflow
analysisConfig := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 60 * time.Second,
}

analysisWorkflow, _ := v1beta.NewParallelWorkflow(analysisConfig)
analysisWorkflow.AddStep(v1beta.WorkflowStep{Name: "sentiment", Agent: sentimentAgent})
analysisWorkflow.AddStep(v1beta.WorkflowStep{Name: "entities", Agent: entityAgent})
analysisWorkflow.AddStep(v1beta.WorkflowStep{Name: "topics", Agent: topicAgent})

// Wrap as agent
analysisAsAgent := v1beta.NewSubWorkflowAgent("analysis", analysisWorkflow)

// Main DAG workflow using subworkflow
mainConfig := &v1beta.WorkflowConfig{
    Mode:    v1beta.DAG,
    Timeout: 120 * time.Second,
}

mainWorkflow, _ := v1beta.NewDAGWorkflow(mainConfig)
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

## 📡 Workflow Streaming

Monitor workflow execution in real-time:

### Sequential Workflow Streaming

```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 120 * time.Second,
}

workflow, _ := v1beta.NewSequentialWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "step1", Agent: agent1})
workflow.AddStep(v1beta.WorkflowStep{Name: "step2", Agent: agent2})
workflow.AddStep(v1beta.WorkflowStep{Name: "step3", Agent: agent3})

stream, err := workflow.RunStream(context.Background(), "First task")
if err != nil {
    log.Fatal(err)
}

for chunk := range stream.Chunks() {
    switch chunk.Type {
    case v1beta.ChunkTypeMetadata:
        if step, ok := chunk.Metadata["step_name"].(string); ok {
            fmt.Printf("→ Executing: %s\n", step)
        }
    case v1beta.ChunkTypeDelta:
        fmt.Print(chunk.Delta)
    case v1beta.ChunkTypeDone:
        fmt.Println("\n✓ Workflow complete")
    }
}
```

### Parallel Workflow Progress

```go
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 90 * time.Second,
}

workflow, _ := v1beta.NewParallelWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "task1", Agent: agent1})
workflow.AddStep(v1beta.WorkflowStep{Name: "task2", Agent: agent2})
workflow.AddStep(v1beta.WorkflowStep{Name: "task3", Agent: agent3})

stream, _ := workflow.RunStream(context.Background(), "Process tasks")
completedTasks := make(map[string]bool)

for chunk := range stream.Chunks() {
    if chunk.Type == v1beta.ChunkTypeMetadata {
        if step, ok := chunk.Metadata["step_name"].(string); ok {
            if status, ok := chunk.Metadata["status"].(string); ok && status == "completed" {
                completedTasks[step] = true
                fmt.Printf("✓ Completed: %s (%d/3)\n", step, len(completedTasks))
            }
        }
    }
}
```

---

## 🎯 Best Practices

### 1. Choose the Right Workflow Type

```go
// ✅ Good - Sequential for dependent steps
config := &v1beta.WorkflowConfig{
    Mode:    v1beta.Sequential,
    Timeout: 60 * time.Second,
}
workflow, _ := v1beta.NewSequentialWorkflow(config)
workflow.AddStep(v1beta.WorkflowStep{Name: "fetch", Agent: fetchAgent})
workflow.AddStep(v1beta.WorkflowStep{Name: "process", Agent: processAgent})

// ✅ Good - Parallel for independent tasks
config2 := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 60 * time.Second,
}
workflow2, _ := v1beta.NewParallelWorkflow(config2)
workflow2.AddStep(v1beta.WorkflowStep{Name: "tech", Agent: techAgent})
workflow2.AddStep(v1beta.WorkflowStep{Name: "biz", Agent: bizAgent})
```

### 2. Handle Errors at Workflow Level

```go
// ✅ Good - Error handling
result, err := workflow.Run(ctx, "Task input")
if err != nil {
    log.Printf("Workflow failed: %v", err)
    // Check partial results
    if result != nil && result.StepResults != nil {
        for step, stepResult := range result.StepResults {
            if stepResult.Success {
                log.Printf("Step %s succeeded", step)
            }
        }
    }
}
```

### 3. Set Appropriate Timeouts

```go
// ✅ Good - Reasonable timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := workflow.Run(ctx, "Task")
```

### 4. Use Subworkflows for Reusability

```go
// ✅ Good - Reusable validation logic
validationConfig := &v1beta.WorkflowConfig{
    Mode:    v1beta.Parallel,
    Timeout: 30 * time.Second,
}
validationWorkflow, _ := v1beta.NewParallelWorkflow(validationConfig)
validationWorkflow.AddStep(v1beta.WorkflowStep{Name: "format", Agent: formatAgent})
validationWorkflow.AddStep(v1beta.WorkflowStep{Name: "data", Agent: dataAgent})

// Wrap and use in multiple places
validationAsAgent := v1beta.NewSubWorkflowAgent("validation", validationWorkflow)

config1 := &v1beta.WorkflowConfig{Mode: v1beta.Sequential, Timeout: 60 * time.Second}
pipeline1, _ := v1beta.NewSequentialWorkflow(config1)
pipeline1.AddStep(v1beta.WorkflowStep{Name: "validate", Agent: validationAsAgent})
pipeline1.AddStep(v1beta.WorkflowStep{Name: "process", Agent: processAgent})

config2 := &v1beta.WorkflowConfig{Mode: v1beta.Sequential, Timeout: 60 * time.Second}
pipeline2, _ := v1beta.NewSequentialWorkflow(config2)
pipeline2.AddStep(v1beta.WorkflowStep{Name: "validate", Agent: validationAsAgent})
pipeline2.AddStep(v1beta.WorkflowStep{Name: "transform", Agent: transformAgent})
```
)
```

### 6. Track Loop Iterations

```go
// ✅ Good - Safeguard against infinite loops
shouldContinue := func(results map[string]*v1beta.AgentResult, iteration int) bool {
    if iteration >= 10 {
        return false // Maximum iterations
    }
    // ... other conditions
}
```

---

## 🔍 Troubleshooting

### Issue: DAG Circular Dependency

**Symptoms**: Workflow fails to start with dependency error

**Solution**: Check dependencies don't form cycles

```go
// ❌ Bad - Circular dependency
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "a",
    Agent:        agentA,
    Dependencies: []string{"b"}, // Circular!
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "b",
    Agent:        agentB,
    Dependencies: []string{"a"},
})

// ✅ Good - Acyclic dependencies
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "a",
    Agent:        agentA,
    Dependencies: nil,
})
workflow.AddStep(v1beta.WorkflowStep{
    Name:         "b",
    Agent:        agentB,
    Dependencies: []string{"a"},
})
```

### Issue: Step Result Access

**Symptoms**: Need to access previous step results

**Solution**: Results available in WorkflowResult.StepResults map

```go
result, _ := workflow.Run(ctx, "Input")
for stepName, stepResult := range result.StepResults {
    fmt.Printf("Step %s: %s\n", stepName, stepResult.FinalOutput)
}
```

### Issue: Parallel Workflow Hangs

**Symptoms**: Some tasks complete but workflow doesn't finish

**Solution**: Check for agent deadlocks or infinite waits

```go
// ✅ Good - Set timeouts
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

workflow.Run(ctx)
```

---

## 📚 Complete Examples

See full workflow examples:
- [Sequential Workflow Example](./examples/workflow-sequential.md)
- [Parallel Workflow Example](./examples/workflow-parallel.md)
- [DAG Workflow Example](./examples/workflow-dag.md)
- [Loop Workflow Example](./examples/workflow-loop.md)
- [Subworkflow Composition](./examples/subworkflow-composition.md)

---

## 🔗 Related Topics

- **[Core Concepts](./core-concepts.md)** - Understanding agents
- **[Streaming](./streaming.md)** - Workflow streaming patterns
- **[Custom Handlers](./custom-handlers.md)** - Advanced agent logic

---

**Next steps?** Continue to [Memory & RAG Guide](./memory-and-rag.md) →
