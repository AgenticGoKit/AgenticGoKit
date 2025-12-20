# DAG Workflow Example

Execute agents with complex dependencies using Directed Acyclic Graph patterns.

---

## Overview

DAG workflows allow you to define complex execution dependencies where steps can run in parallel when their dependencies are met.

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
    // Create agents
    dataFetcher, _ := v1beta.NewBuilder("DataFetcher").WithLLM("openai", "gpt-4").Build()
    validator, _ := v1beta.NewBuilder("Validator").WithLLM("openai", "gpt-3.5-turbo").Build()
    processor, _ := v1beta.NewBuilder("Processor").WithLLM("openai", "gpt-4").Build()
    analyzer, _ := v1beta.NewBuilder("Analyzer").WithLLM("openai", "gpt-4").Build()
    reporter, _ := v1beta.NewBuilder("Reporter").WithLLM("openai", "gpt-4").Build()

    // Create DAG workflow with dependencies
    workflow, err := v1beta.NewDAGWorkflow("DataPipeline",
        v1beta.DAGStep("fetch", dataFetcher, "Fetch data", nil), // No dependencies
        v1beta.DAGStep("validate", validator, "Validate data", []string{"fetch"}),
        v1beta.DAGStep("process", processor, "Process data", []string{"validate"}),
        v1beta.DAGStep("analyze", analyzer, "Analyze data", []string{"validate"}), // Parallel with process
        v1beta.DAGStep("report", reporter, "Generate report", []string{"process", "analyze"}),
    )
    if err != nil {
        log.Fatalf("Failed to create DAG workflow: %v", err)
    }

    // Execute workflow
    results, err := workflow.Run(context.Background(), "customer_data.csv")
    if err != nil {
        log.Fatalf("Workflow failed: %v", err)
    }

    // Access results
    fmt.Println("=== Final Report ===")
    fmt.Println(results["report"].Content)
}
```

---

## Execution Graph

```
          fetch
            │
            ▼
         validate
         ╱      ╲
        ▼        ▼
     process   analyze
        ╲      ╱
         ▼    ▼
        report
```

**Execution Order:**
1. `fetch` runs first (no dependencies)
2. `validate` runs after `fetch` completes
3. `process` and `analyze` run in parallel after `validate`
4. `report` runs after both `process` and `analyze` complete

---

## Key Concepts

### Defining Dependencies

```go
// No dependencies - runs immediately
v1beta.DAGStep("step1", agent1, "Description", nil)

// Depends on one step
v1beta.DAGStep("step2", agent2, "Description", []string{"step1"})

// Depends on multiple steps
v1beta.DAGStep("step3", agent3, "Description", []string{"step1", "step2"})
```

### Data Flow

Each step receives the combined output of all its dependencies:

```go
// Single dependency - receives that step's output
v1beta.DAGStep("step2", agent2, "Process", []string{"step1"})
// step2 receives: results["step1"].Content

// Multiple dependencies - receives concatenated outputs
v1beta.DAGStep("step3", agent3, "Merge", []string{"step1", "step2"})
// step3 receives: results["step1"].Content + "\n\n" + results["step2"].Content
```

---

## Real-World Examples

### Content Generation Pipeline

```go
workflow, _ := v1beta.NewDAGWorkflow("ContentPipeline",
    v1beta.DAGStep("research", researchAgent, "Research topic", nil),
    v1beta.DAGStep("outline", outlineAgent, "Create outline", []string{"research"}),
    v1beta.DAGStep("facts", factAgent, "Gather facts", []string{"research"}),
    v1beta.DAGStep("draft", draftAgent, "Write draft", []string{"outline", "facts"}),
    v1beta.DAGStep("edit", editAgent, "Edit content", []string{"draft"}),
    v1beta.DAGStep("seo", seoAgent, "SEO optimization", []string{"draft"}),
    v1beta.DAGStep("finalize", finalAgent, "Finalize", []string{"edit", "seo"}),
)
```

### Multi-Source Analysis

```go
workflow, _ := v1beta.NewDAGWorkflow("Analysis",
    // Parallel data collection
    v1beta.DAGStep("web", webAgent, "Fetch web data", nil),
    v1beta.DAGStep("db", dbAgent, "Fetch DB data", nil),
    v1beta.DAGStep("api", apiAgent, "Fetch API data", nil),
    
    // Process each source
    v1beta.DAGStep("process_web", processAgent, "Process web", []string{"web"}),
    v1beta.DAGStep("process_db", processAgent, "Process DB", []string{"db"}),
    v1beta.DAGStep("process_api", processAgent, "Process API", []string{"api"}),
    
    // Merge and analyze
    v1beta.DAGStep("merge", mergeAgent, "Merge data", []string{"process_web", "process_db", "process_api"}),
    v1beta.DAGStep("analyze", analyzeAgent, "Analyze", []string{"merge"}),
)
```

---

## Advanced Patterns

### Conditional Execution

```go
// Custom DAG with conditional logic
type ConditionalDAG struct {
    workflow v1beta.Workflow
}

func (d *ConditionalDAG) Run(ctx context.Context, input string) (map[string]*v1beta.Result, error) {
    results, err := d.workflow.Run(ctx, input)
    if err != nil {
        return nil, err
    }
    
    // Check condition and potentially skip steps
    if needsReview(results["analyze"]) {
        reviewAgent, _ := v1beta.NewBuilder("Reviewer").WithLLM("openai", "gpt-4").Build()
        reviewResult, _ := reviewAgent.Run(ctx, results["analyze"].Content)
        results["review"] = reviewResult
    }
    
    return results, nil
}
```

### Dynamic DAG Generation

```go
func buildDAG(steps []StepConfig) (v1beta.Workflow, error) {
    dagSteps := make([]v1beta.DAGStepConfig, len(steps))
    
    for i, step := range steps {
        agent, _ := v1beta.NewBuilder(step.Name).
            WithLLM(step.Provider, step.Model).
            Build()
        
        dagSteps[i] = v1beta.DAGStep(
            step.ID,
            agent,
            step.Description,
            step.Dependencies,
        )
    }
    
    return v1beta.NewDAGWorkflow("DynamicPipeline", dagSteps...)
}
```

---

## Error Handling

### Dependency Failure Propagation

```go
results, err := workflow.Run(ctx, input)
if err != nil {
    if dagErr, ok := err.(*v1beta.DAGWorkflowError); ok {
        fmt.Printf("Failed step: %s\n", dagErr.FailedStep)
        fmt.Printf("Error: %v\n", dagErr.Err)
        
        // Steps that completed before failure
        fmt.Println("Completed steps:")
        for stepID := range dagErr.CompletedSteps {
            fmt.Printf("  - %s\n", stepID)
        }
        
        // Steps that were skipped due to dependency failure
        fmt.Println("Skipped steps:")
        for stepID := range dagErr.SkippedSteps {
            fmt.Printf("  - %s\n", stepID)
        }
    }
}
```

---

## Performance Tips

### Optimize Parallelism

```go
// Bad: Sequential dependencies
v1beta.DAGStep("step2", agent, "...", []string{"step1"}),
v1beta.DAGStep("step3", agent, "...", []string{"step2"}),
v1beta.DAGStep("step4", agent, "...", []string{"step3"}),

// Good: Parallel execution where possible
v1beta.DAGStep("step2", agent, "...", []string{"step1"}),
v1beta.DAGStep("step3", agent, "...", []string{"step1"}), // Parallel with step2
v1beta.DAGStep("step4", agent, "...", []string{"step2", "step3"}),
```

### Visualize DAG

```go
func visualizeDAG(workflow v1beta.DAGWorkflow) string {
    var mermaid strings.Builder
    mermaid.WriteString("```mermaid\ngraph TD\n")
    
    for _, step := range workflow.Steps() {
        if len(step.Dependencies) == 0 {
            mermaid.WriteString(fmt.Sprintf("  %s[%s]\n", step.ID, step.Description))
        } else {
            for _, dep := range step.Dependencies {
                mermaid.WriteString(fmt.Sprintf("  %s --> %s\n", dep, step.ID))
            }
        }
    }
    
    mermaid.WriteString("```\n")
    return mermaid.String()
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

- **[Loop Workflow](./workflow-loop.md)** - Iterative execution patterns
- **[Subworkflows](./subworkflow-composition.md)** - Nested DAG workflows
- **[Sequential Workflow](./workflow-sequential.md)** - Simpler linear flows
- **[Parallel Workflow](./workflow-parallel.md)** - Independent parallel execution

---

## Related Documentation

- [Workflows Guide](../workflows.md) - Complete workflow documentation
- [Performance](../performance.md) - DAG optimization strategies
- [Error Handling](../error-handling.md) - Handling DAG execution errors
