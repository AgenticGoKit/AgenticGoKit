# Subworkflow Composition Example

Nest workflows within workflows for complex multi-level agent systems.

---

## Overview

Subworkflows allow you to compose complex agent systems by nesting workflows within other workflows.

---

## Complete Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create analysis subworkflow (parallel)
    sentimentAgent, _ := v1beta.NewBuilder("Sentiment").WithLLM("openai", "gpt-3.5-turbo").Build()
    topicsAgent, _ := v1beta.NewBuilder("Topics").WithLLM("openai", "gpt-3.5-turbo").Build()
    
    analysisWorkflow, _ := v1beta.NewParallelWorkflow(
        "Analysis",
        v1beta.Step("sentiment", sentimentAgent, "Analyze sentiment"),
        v1beta.Step("topics", topicsAgent, "Extract topics"),
    )

    // Create processing subworkflow (sequential)
    cleaner, _ := v1beta.NewBuilder("Cleaner").WithLLM("openai", "gpt-3.5-turbo").Build()
    normalizer, _ := v1beta.NewBuilder("Normalizer").WithLLM("openai", "gpt-3.5-turbo").Build()
    
    processingWorkflow, _ := v1beta.NewSequentialWorkflow(
        "Processing",
        v1beta.Step("clean", cleaner, "Clean text"),
        v1beta.Step("normalize", normalizer, "Normalize format"),
    )

    // Create main workflow with subworkflows
    mainWorkflow, err := v1beta.NewSequentialWorkflow(
        "MainPipeline",
        v1beta.SubWorkflowStep("process", processingWorkflow, "Process input"),
        v1beta.SubWorkflowStep("analyze", analysisWorkflow, "Analyze processed data"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Execute
    results, err := mainWorkflow.Run(context.Background(), "Sample customer feedback text...")
    if err != nil {
        log.Fatal(err)
    }

    // Access nested results
    processingResults := results["process"].(map[string]*v1beta.Result)
    analysisResults := results["analyze"].(map[string]*v1beta.Result)

    fmt.Println("=== Processed Text ===")
    fmt.Println(processingResults["normalize"].Content)
    
    fmt.Println("\n=== Analysis ===")
    fmt.Println("Sentiment:", analysisResults["sentiment"].Content)
    fmt.Println("Topics:", analysisResults["topics"].Content)
}
```

---

## Key Concepts

### Subworkflow Step

```go
// Create subworkflow
subWorkflow, _ := v1beta.NewParallelWorkflow(
    "Sub",
    v1beta.Step("step1", agent1, "Description"),
    v1beta.Step("step2", agent2, "Description"),
)

// Use in parent workflow
mainWorkflow, _ := v1beta.NewSequentialWorkflow(
    "Main",
    v1beta.SubWorkflowStep("sub", subWorkflow, "Execute subworkflow"),
)
```

### Nested Results

```go
results, _ := mainWorkflow.Run(ctx, input)

// Access subworkflow results
subResults := results["sub"].(map[string]*v1beta.Result)
step1Result := subResults["step1"]
step2Result := subResults["step2"]
```

---

## Real-World Examples

### Multi-Stage Content Pipeline

```go
// Research subworkflow
researchWorkflow, _ := v1beta.NewParallelWorkflow("Research",
    v1beta.Step("web", webAgent, "Web research"),
    v1beta.Step("papers", paperAgent, "Academic papers"),
)

// Writing subworkflow
writingWorkflow, _ := v1beta.NewSequentialWorkflow("Writing",
    v1beta.Step("outline", outlineAgent, "Create outline"),
    v1beta.Step("draft", draftAgent, "Write draft"),
)

// Review subworkflow
reviewWorkflow, _ := v1beta.NewParallelWorkflow("Review",
    v1beta.Step("grammar", grammarAgent, "Check grammar"),
    v1beta.Step("facts", factAgent, "Verify facts"),
)

// Main pipeline
pipeline, _ := v1beta.NewSequentialWorkflow("ContentPipeline",
    v1beta.SubWorkflowStep("research", researchWorkflow, "Research phase"),
    v1beta.SubWorkflowStep("writing", writingWorkflow, "Writing phase"),
    v1beta.SubWorkflowStep("review", reviewWorkflow, "Review phase"),
)
```

### Customer Support System

```go
// Classification subworkflow
classificationWorkflow, _ := v1beta.NewParallelWorkflow("Classification",
    v1beta.Step("category", categoryAgent, "Categorize issue"),
    v1beta.Step("priority", priorityAgent, "Determine priority"),
    v1beta.Step("sentiment", sentimentAgent, "Analyze sentiment"),
)

// Resolution subworkflow  
resolutionWorkflow, _ := v1beta.NewSequentialWorkflow("Resolution",
    v1beta.Step("search", searchAgent, "Search knowledge base"),
    v1beta.Step("generate", generateAgent, "Generate response"),
    v1beta.Step("validate", validateAgent, "Validate response"),
)

// Main support workflow
supportWorkflow, _ := v1beta.NewSequentialWorkflow("Support",
    v1beta.SubWorkflowStep("classify", classificationWorkflow, "Classify request"),
    v1beta.SubWorkflowStep("resolve", resolutionWorkflow, "Resolve issue"),
)
```

---

## Advanced Patterns

### Conditional Subworkflows

```go
func createDynamicWorkflow(needsReview bool) (v1beta.Workflow, error) {
    steps := []v1beta.WorkflowStep{
        v1beta.SubWorkflowStep("process", processingWorkflow, "Process"),
    }
    
    if needsReview {
        steps = append(steps, 
            v1beta.SubWorkflowStep("review", reviewWorkflow, "Review"),
        )
    }
    
    steps = append(steps,
        v1beta.SubWorkflowStep("finalize", finalizeWorkflow, "Finalize"),
    )
    
    return v1beta.NewSequentialWorkflow("Dynamic", steps...)
}
```

### Recursive Subworkflows

```go
func createRecursiveWorkflow(depth int) (v1beta.Workflow, error) {
    if depth == 0 {
        // Base case: simple agent
        agent, _ := v1beta.NewBuilder("Base").WithLLM("openai", "gpt-4").Build()
        return v1beta.NewSequentialWorkflow("Base",
            v1beta.Step("process", agent, "Process"),
        )
    }
    
    // Recursive case: nest subworkflows
    subWorkflow, _ := createRecursiveWorkflow(depth - 1)
    agent, _ := v1beta.NewBuilder(fmt.Sprintf("Level%d", depth)).
        WithLLM("openai", "gpt-4").
        Build()
    
    return v1beta.NewSequentialWorkflow(
        fmt.Sprintf("Level%d", depth),
        v1beta.Step("process", agent, "Process at this level"),
        v1beta.SubWorkflowStep("sub", subWorkflow, "Nested processing"),
    )
}
```

---

## Error Handling

```go
results, err := mainWorkflow.Run(ctx, input)
if err != nil {
    if subErr, ok := err.(*v1beta.SubWorkflowError); ok {
        fmt.Printf("Subworkflow '%s' failed: %v\n", subErr.SubWorkflowID, subErr.Err)
        
        // Access results from successful subworkflows
        for id, result := range subErr.CompletedSubWorkflows {
            fmt.Printf("Subworkflow %s completed successfully\n", id)
        }
    }
}
```

---

## Running the Example

```bash
go get github.com/agenticgokit/agenticgokit/v1beta
export OPENAI_API_KEY="sk-..."
go run main.go
```

---

## Next Steps

- **[Sequential Workflow](./workflow-sequential.md)** - Basic workflow patterns
- **[Parallel Workflow](./workflow-parallel.md)** - Concurrent execution
- **[DAG Workflow](./workflow-dag.md)** - Complex dependencies
- **[Custom Handlers](./custom-handlers.md)** - Custom workflow logic

---

## Related Documentation

- [Workflows Guide](../workflows.md) - Complete workflow documentation
- [Core Concepts](../core-concepts.md) - Understanding subworkflows
- [Performance](../performance.md) - Optimizing nested workflows
