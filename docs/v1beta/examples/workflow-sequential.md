# Sequential Workflow Example

Execute agents in a step-by-step sequence, passing results between steps.

---

## Overview

This example demonstrates:
- Creating a sequential workflow
- Passing data between workflow steps
- Error handling in workflows
- Accessing step results

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
    // Create specialized agents
    researcher, err := v1beta.NewBuilder("Researcher").
        WithLLM("openai", "gpt-4").
        Build()
    if err != nil {
        log.Fatalf("Failed to create researcher: %v", err)
    }

    summarizer, err := v1beta.NewBuilder("Summarizer").
        WithLLM("openai", "gpt-3.5-turbo").
        Build()
    if err != nil {
        log.Fatalf("Failed to create summarizer: %v", err)
    }

    writer, err := v1beta.NewBuilder("Writer").
        WithLLM("openai", "gpt-4").
        Build()
    if err != nil {
        log.Fatalf("Failed to create writer: %v", err)
    }

    // Create sequential workflow
    workflow, err := v1beta.NewSequentialWorkflow(
        "ResearchPipeline",
        v1beta.Step("research", researcher, "Research the topic"),
        v1beta.Step("summarize", summarizer, "Summarize findings"),
        v1beta.Step("write", writer, "Write final report"),
    )
    if err != nil {
        log.Fatalf("Failed to create workflow: %v", err)
    }

    // Execute workflow
    results, err := workflow.Run(context.Background(), "Latest developments in Go 1.23")
    if err != nil {
        log.Fatalf("Workflow failed: %v", err)
    }

    // Access results from each step
    fmt.Println("=== Research Step ===")
    fmt.Println(results["research"].Content)
    
    fmt.Println("\n=== Summary Step ===")
    fmt.Println(results["summarize"].Content)
    
    fmt.Println("\n=== Final Report ===")
    fmt.Println(results["write"].Content)
}
```

---

## Step-by-Step Breakdown

### 1. Create Individual Agents

```go
researcher, err := v1beta.NewBuilder("Researcher").
    WithLLM("openai", "gpt-4").
    Build()

summarizer, err := v1beta.NewBuilder("Summarizer").
    WithLLM("openai", "gpt-3.5-turbo").
    Build()

writer, err := v1beta.NewBuilder("Writer").
    WithLLM("openai", "gpt-4").
    Build()
```

Each agent can have different configurations, models, or providers.

### 2. Define Workflow Steps

```go
workflow, err := v1beta.NewSequentialWorkflow(
    "ResearchPipeline",
    v1beta.Step("research", researcher, "Research the topic"),
    v1beta.Step("summarize", summarizer, "Summarize findings"),
    v1beta.Step("write", writer, "Write final report"),
)
```

**Step Parameters:**
- `id` - Unique identifier for the step
- `agent` - Agent to execute
- `description` - Human-readable description

### 3. Execute Workflow

```go
results, err := workflow.Run(context.Background(), "Latest developments in Go 1.23")
```

The input is passed to the first step. Each subsequent step receives the output of the previous step.

### 4. Access Results

```go
// Results is a map[string]*Result
fmt.Println(results["research"].Content)
fmt.Println(results["summarize"].Content)
fmt.Println(results["write"].Content)
```

---

## Data Flow Pattern

### How Data Flows Between Steps

```
Input: "Latest developments in Go 1.23"
   ↓
Step 1 (research): Receives input, produces research findings
   ↓
Step 2 (summarize): Receives research findings, produces summary
   ↓
Step 3 (write): Receives summary, produces final report
   ↓
Output: Map of all step results
```

### Example Flow

```go
// Step 1 receives: "Latest developments in Go 1.23"
// Step 1 outputs: "Go 1.23 introduces improvements to..."

// Step 2 receives: "Go 1.23 introduces improvements to..."
// Step 2 outputs: "Key points: 1) Performance, 2) Standard library..."

// Step 3 receives: "Key points: 1) Performance, 2) Standard library..."
// Step 3 outputs: "# Go 1.23 Report\n\n## Overview..."
```

---

## Advanced Patterns

### With Custom Step Prompts

```go
// Create workflow with custom prompts per step
workflow, err := v1beta.NewSequentialWorkflow(
    "DataPipeline",
    v1beta.StepWithPrompt("extract", extractor, 
        "Extract key data points from the following text",
        "Focus on numerical data and dates"),
    v1beta.StepWithPrompt("transform", transformer,
        "Transform the extracted data into JSON format",
        "Use snake_case for field names"),
    v1beta.StepWithPrompt("validate", validator,
        "Validate the JSON structure",
        "Check for required fields and data types"),
)
```

### With Error Recovery

```go
results, err := workflow.Run(ctx, input)
if err != nil {
    // Check which step failed
    if workflowErr, ok := err.(*v1beta.WorkflowError); ok {
        fmt.Printf("Failed at step: %s\n", workflowErr.StepID)
        fmt.Printf("Error: %v\n", workflowErr.Err)
        
        // Access partial results
        for stepID, result := range workflowErr.PartialResults {
            fmt.Printf("Step %s succeeded with: %s\n", stepID, result.Content)
        }
    }
}
```

### With Step Conditions

```go
// Custom workflow with conditional execution
type ConditionalWorkflow struct {
    steps []v1beta.WorkflowStep
}

func (w *ConditionalWorkflow) Run(ctx context.Context, input string) (map[string]*v1beta.Result, error) {
    results := make(map[string]*v1beta.Result)
    currentInput := input
    
    for _, step := range w.steps {
        // Execute step
        result, err := step.Agent.Run(ctx, currentInput)
        if err != nil {
            return results, err
        }
        results[step.ID] = result
        
        // Check condition before proceeding
        if shouldSkipNextStep(result) {
            break
        }
        
        currentInput = result.Content
    }
    
    return results, nil
}
```

---

## Real-World Use Cases

### ETL Pipeline

```go
extractor, _ := v1beta.NewBuilder("Extractor").
    WithLLM("openai", "gpt-4").
    Build()

transformer, _ := v1beta.NewBuilder("Transformer").
    WithLLM("openai", "gpt-4").
    Build()

loader, _ := v1beta.NewBuilder("Loader").
    WithLLM("openai", "gpt-3.5-turbo").
    Build()

workflow, _ := v1beta.NewSequentialWorkflow(
    "ETL",
    v1beta.Step("extract", extractor, "Extract data from source"),
    v1beta.Step("transform", transformer, "Transform to target schema"),
    v1beta.Step("load", loader, "Generate load commands"),
)

results, err := workflow.Run(ctx, sourceData)
```

### Content Creation Pipeline

```go
// Research → Outline → Draft → Edit → Format
workflow, _ := v1beta.NewSequentialWorkflow(
    "ContentPipeline",
    v1beta.Step("research", researchAgent, "Research topic"),
    v1beta.Step("outline", outlineAgent, "Create outline"),
    v1beta.Step("draft", draftAgent, "Write first draft"),
    v1beta.Step("edit", editAgent, "Edit and refine"),
    v1beta.Step("format", formatAgent, "Format for publication"),
)
```

### Customer Support Pipeline

```go
// Classify → Route → Respond → Follow-up
workflow, _ := v1beta.NewSequentialWorkflow(
    "SupportPipeline",
    v1beta.Step("classify", classifyAgent, "Classify issue type"),
    v1beta.Step("route", routeAgent, "Determine responsible team"),
    v1beta.Step("respond", respondAgent, "Generate response"),
    v1beta.Step("followup", followupAgent, "Create follow-up tasks"),
)
```

---

## Running the Example

### Prerequisites

```bash
go get github.com/agenticgokit/agenticgokit/v1beta
export OPENAI_API_KEY="sk-..."
```

### Execute

```bash
go run main.go
```

---

## Performance Considerations

### Step-Level Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

results, err := workflow.Run(ctx, input)
```

### Logging Step Progress

```go
for i, step := range workflow.Steps() {
    fmt.Printf("Executing step %d/%d: %s\n", i+1, len(workflow.Steps()), step.Description)
}

results, err := workflow.Run(ctx, input)
```

---

## Error Handling

### Graceful Failure Handling

```go
results, err := workflow.Run(ctx, input)
if err != nil {
    // Log the error
    log.Printf("Workflow failed: %v", err)
    
    // Check if we have partial results
    if workflowErr, ok := err.(*v1beta.WorkflowError); ok {
        log.Printf("Partial results available from %d steps", len(workflowErr.PartialResults))
        
        // Use partial results if possible
        return workflowErr.PartialResults, err
    }
    
    return nil, err
}
```

### Retry Failed Steps

```go
func runWithRetry(workflow v1beta.Workflow, ctx context.Context, input string, maxRetries int) (map[string]*v1beta.Result, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        results, err := workflow.Run(ctx, input)
        if err == nil {
            return results, nil
        }
        
        lastErr = err
        log.Printf("Attempt %d failed: %v", i+1, err)
        time.Sleep(time.Second * time.Duration(i+1))
    }
    
    return nil, fmt.Errorf("workflow failed after %d retries: %w", maxRetries, lastErr)
}
```

---

## Next Steps

- **[Parallel Workflow](./workflow-parallel.md)** - Execute agents concurrently
- **[DAG Workflow](./workflow-dag.md)** - Complex dependencies
- **[Loop Workflow](./workflow-loop.md)** - Iterative processing
- **[Subworkflows](./subworkflow-composition.md)** - Nest workflows

---

## Related Documentation

- [Workflows Guide](../workflows.md) - Complete workflow documentation
- [Core Concepts](../core-concepts.md) - Understanding workflows
- [Error Handling](../error-handling.md) - Workflow error patterns
- [Performance](../performance.md) - Workflow optimization
