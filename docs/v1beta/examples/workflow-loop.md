# Loop Workflow Example

Iterative agent execution with convergence conditions.

---

## Overview

Loop workflows execute an agent repeatedly until a condition is met or a maximum iteration count is reached.

---

## Complete Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
    // Create refining agent
    refiner, err := v1beta.NewBuilder("Refiner").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Convergence function - stops when quality threshold met
    converged := func(iteration int, result *v1beta.Result) bool {
        // Stop after 5 iterations max
        if iteration >= 5 {
            return true
        }
        
        // Stop if result contains "FINAL"
        return strings.Contains(result.Content, "FINAL")
    }

    // Create loop workflow
    workflow, err := v1beta.NewLoopWorkflow(
        "RefinementLoop",
        refiner,
        converged,
        v1beta.WithMaxIterations(5),
        v1beta.WithLoopPrompt("Refine the following text. Add 'FINAL' when perfect"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Execute loop
    results, err := workflow.Run(context.Background(), "Write a short Go tutorial")
    if err != nil {
        log.Fatal(err)
    }

    // Display iterations
    fmt.Printf("Converged after %d iterations\n\n", len(results))
    for i, result := range results {
        fmt.Printf("=== Iteration %d ===\n", i+1)
        fmt.Println(result.Content)
        fmt.Println()
    }
}
```

---

## Key Concepts

### Convergence Function

```go
type ConvergenceFunc func(iteration int, result *v1beta.Result) bool

// Simple iteration limit
converged := func(iteration int, result *v1beta.Result) bool {
    return iteration >= 10
}

// Content-based convergence
converged := func(iteration int, result *v1beta.Result) bool {
    return strings.Contains(result.Content, "COMPLETE")
}

// Metadata-based convergence
converged := func(iteration int, result *v1beta.Result) bool {
    if score, ok := result.Metadata["quality_score"].(float64); ok {
        return score > 0.95
    }
    return false
}
```

### Loop Options

```go
workflow, _ := v1beta.NewLoopWorkflow(
    "Loop",
    agent,
    converged,
    v1beta.WithMaxIterations(10),              // Max iterations
    v1beta.WithLoopPrompt("Improve the text"), // Prompt for each iteration
    v1beta.WithLoopTimeout(5*time.Minute),     // Total timeout
)
```

---

## Real-World Examples

### Code Review Loop

```go
reviewer, _ := v1beta.NewBuilder("CodeReviewer").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()

converged := func(iteration int, result *v1beta.Result) bool {
    // Stop if no more issues found
    return !strings.Contains(result.Content, "ISSUE:")
}

workflow, _ := v1beta.NewLoopWorkflow(
    "CodeReview",
    reviewer,
    converged,
    v1beta.WithMaxIterations(3),
    v1beta.WithLoopPrompt("Review code and suggest improvements"),
)
```

### Iterative Research

```go
researcher, _ := v1beta.NewBuilder("Researcher").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()

converged := func(iteration int, result *v1beta.Result) bool {
    // Stop when comprehensive
    wordCount := len(strings.Fields(result.Content))
    return wordCount > 500 && strings.Contains(result.Content, "Conclusion")
}

workflow, _ := v1beta.NewLoopWorkflow(
    "Research",
    researcher,
    converged,
    v1beta.WithMaxIterations(5),
)
```

### Content Refinement

```go
editor, _ := v1beta.NewBuilder("Editor").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()

converged := func(iteration int, result *v1beta.Result) bool {
    // Stop when polished
    hasIntro := strings.Contains(result.Content, "Introduction")
    hasConclusion := strings.Contains(result.Content, "Conclusion")
    return iteration >= 3 && hasIntro && hasConclusion
}

workflow, _ := v1beta.NewLoopWorkflow(
    "ContentRefinement",
    editor,
    converged,
    v1beta.WithLoopPrompt("Improve structure and clarity"),
)
```

---

## Advanced Patterns

### With State Tracking

```go
type LoopState struct {
    improvements []string
    scores       []float64
}

state := &LoopState{}

converged := func(iteration int, result *v1beta.Result) bool {
    // Track improvements
    state.improvements = append(state.improvements, result.Content)
    
    // Calculate quality score (simplified)
    score := float64(len(result.Content)) / 1000.0
    state.scores = append(state.scores, score)
    
    // Converge if score plateaus
    if len(state.scores) >= 3 {
        recent := state.scores[len(state.scores)-3:]
        variance := calculateVariance(recent)
        return variance < 0.01 // Low variance = plateau
    }
    
    return false
}
```

### With Feedback Accumulation

```go
var previousFeedback []string

converged := func(iteration int, result *v1beta.Result) bool {
    // Accumulate feedback
    previousFeedback = append(previousFeedback, result.Content)
    
    // Inject previous iterations into next prompt
    if iteration < 5 {
        // Modify agent prompt to include history
        return false
    }
    
    return true
}
```

---

## Error Handling

```go
results, err := workflow.Run(ctx, input)
if err != nil {
    if loopErr, ok := err.(*v1beta.LoopWorkflowError); ok {
        fmt.Printf("Failed at iteration %d: %v\n", loopErr.Iteration, loopErr.Err)
        
        // Access results from successful iterations
        fmt.Printf("Completed %d iterations successfully\n", len(loopErr.PartialResults))
        for i, result := range loopErr.PartialResults {
            fmt.Printf("Iteration %d: %s\n", i+1, result.Content)
        }
    }
}
```

---

## Performance Tips

### Set Reasonable Limits

```go
// Prevent infinite loops
workflow, _ := v1beta.NewLoopWorkflow(
    "Loop",
    agent,
    converged,
    v1beta.WithMaxIterations(10),                    // Hard limit
    v1beta.WithLoopTimeout(5*time.Minute),           // Time limit
    v1beta.WithIterationTimeout(30*time.Second),     // Per-iteration limit
)
```

### Early Termination

```go
converged := func(iteration int, result *v1beta.Result) bool {
    // Immediate termination conditions
    if strings.Contains(result.Content, "ERROR") {
        return true
    }
    
    if strings.Contains(result.Content, "CANNOT_IMPROVE") {
        return true
    }
    
    // Normal convergence check
    return iteration >= 10
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

- **[Subworkflows](./subworkflow-composition.md)** - Nested loop workflows
- **[DAG Workflow](./workflow-dag.md)** - Complex dependencies
- **[Custom Handlers](./custom-handlers.md)** - Custom loop logic

---

## Related Documentation

- [Workflows Guide](../workflows.md) - Complete workflow documentation
- [Performance](../performance.md) - Loop optimization strategies
- [Error Handling](../error-handling.md) - Handling loop errors
