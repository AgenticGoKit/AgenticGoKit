# Parallel Workflow Example

Execute multiple agents concurrently for faster processing.

---

## Overview

This example demonstrates:
- Running agents in parallel
- Collecting results from concurrent execution
- Handling errors in parallel workflows
- Performance benefits of concurrency

---

## Complete Code

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
    // Create specialized agents for different analysis tasks
    sentimentAgent, _ := v1beta.NewBuilder("SentimentAnalyzer").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-3.5-turbo"},
        }).
        Build()

    topicsAgent, _ := v1beta.NewBuilder("TopicExtractor").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-3.5-turbo"},
        }).
        Build()

    summaryAgent, _ := v1beta.NewBuilder("Summarizer").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
        }).
        Build()

    keywordsAgent, _ := v1beta.NewBuilder("KeywordExtractor").
        WithConfig(&v1beta.Config{
            LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-3.5-turbo"},
        }).
        Build()

    // Create parallel workflow
    workflow, err := v1beta.NewParallelWorkflow(
        "TextAnalysis",
        v1beta.Step("sentiment", sentimentAgent, "Analyze sentiment"),
        v1beta.Step("topics", topicsAgent, "Extract topics"),
        v1beta.Step("summary", summaryAgent, "Generate summary"),
        v1beta.Step("keywords", keywordsAgent, "Extract keywords"),
    )
    if err != nil {
        log.Fatalf("Failed to create workflow: %v", err)
    }

    // Sample text to analyze
    text := `
    Go 1.23 brings exciting new features and improvements. 
    The performance enhancements are remarkable, with significant 
    improvements in the garbage collector. The standard library 
    additions make development even more enjoyable.
    `

    // Execute all agents in parallel
    start := time.Now()
    results, err := workflow.Run(context.Background(), text)
    if err != nil {
        log.Fatalf("Workflow failed: %v", err)
    }
    duration := time.Since(start)

    // Display results
    fmt.Println("=== Analysis Results ===\n")
    
    fmt.Println("Sentiment:")
    fmt.Println(results["sentiment"].Content)
    fmt.Println()
    
    fmt.Println("Topics:")
    fmt.Println(results["topics"].Content)
    fmt.Println()
    
    fmt.Println("Summary:")
    fmt.Println(results["summary"].Content)
    fmt.Println()
    
    fmt.Println("Keywords:")
    fmt.Println(results["keywords"].Content)
    fmt.Println()
    
    fmt.Printf("Total execution time: %v\n", duration)
}
```

---

## Step-by-Step Breakdown

### 1. Create Multiple Agents

```go
sentimentAgent, _ := v1beta.NewBuilder("SentimentAnalyzer").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-3.5-turbo"},
    }).
    Build()

topicsAgent, _ := v1beta.NewBuilder("TopicExtractor").
    WithConfig(&v1beta.Config{
        LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-3.5-turbo"},
    }).
    Build()

// ... more agents
```

Each agent can use different models and configurations based on their specific task.

### 2. Create Parallel Workflow

```go
workflow, err := v1beta.NewParallelWorkflow(
    "TextAnalysis",
    v1beta.Step("sentiment", sentimentAgent, "Analyze sentiment"),
    v1beta.Step("topics", topicsAgent, "Extract topics"),
    v1beta.Step("summary", summaryAgent, "Generate summary"),
    v1beta.Step("keywords", keywordsAgent, "Extract keywords"),
)
```

All steps receive the **same input** and execute concurrently.

### 3. Execute and Collect Results

```go
results, err := workflow.Run(context.Background(), text)

// Access individual results
sentimentResult := results["sentiment"]
topicsResult := results["topics"]
```

Results are returned as a map keyed by step ID.

---

## Execution Model

### Parallel vs Sequential

```
Sequential (4 steps × 2 seconds each = 8 seconds):
Step 1 ──► Step 2 ──► Step 3 ──► Step 4
 (2s)      (2s)       (2s)       (2s)

Parallel (max 2 seconds total):
Step 1 ──►
Step 2 ──► All complete at ~2s
Step 3 ──►
Step 4 ──►
```

### How It Works

```go
// All agents receive the same input
input := "Analyze this text..."

// Executed concurrently using goroutines
go func() { results["sentiment"] = sentimentAgent.Run(ctx, input) }()
go func() { results["topics"] = topicsAgent.Run(ctx, input) }()
go func() { results["summary"] = summaryAgent.Run(ctx, input) }()
go func() { results["keywords"] = keywordsAgent.Run(ctx, input) }()

// Workflow waits for all to complete
```

---

## Advanced Patterns

### With Timeout

```go
// Set timeout for entire workflow
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

results, err := workflow.Run(ctx, input)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Workflow timed out")
        // Some agents may have completed
    }
}
```

### With Progress Tracking

```go
type ProgressTracker struct {
    total     int
    completed int
    mu        sync.Mutex
}

func (pt *ProgressTracker) Complete(stepID string) {
    pt.mu.Lock()
    defer pt.mu.Unlock()
    pt.completed++
    fmt.Printf("Progress: %d/%d steps complete\n", pt.completed, pt.total)
}

// Use with custom workflow wrapper
```

### Partial Failure Handling

```go
results, err := workflow.Run(ctx, input)

// Check for partial failures
if err != nil {
    if pErr, ok := err.(*v1beta.ParallelWorkflowError); ok {
        fmt.Printf("Some steps failed: %v\n", pErr.FailedSteps)
        
        // Use successful results
        for stepID, result := range pErr.SuccessfulResults {
            fmt.Printf("Step %s succeeded: %s\n", stepID, result.Content)
        }
    }
}
```

---

## Real-World Use Cases

### Multi-Source Data Aggregation

```go
// Fetch from multiple sources in parallel
newsAgent, _ := v1beta.NewBuilder("NewsAggregator").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()
socialAgent, _ := v1beta.NewBuilder("SocialAggregator").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()
researchAgent, _ := v1beta.NewBuilder("ResearchAggregator").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()

workflow, _ := v1beta.NewParallelWorkflow(
    "DataAggregation",
    v1beta.Step("news", newsAgent, "Fetch news articles"),
    v1beta.Step("social", socialAgent, "Fetch social media posts"),
    v1beta.Step("research", researchAgent, "Fetch research papers"),
)

results, _ := workflow.Run(ctx, "Go programming language")
```

### Multi-Language Translation

```go
// Translate to multiple languages simultaneously
enAgent, _ := v1beta.NewBuilder("EnglishTranslator").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()
esAgent, _ := v1beta.NewBuilder("SpanishTranslator").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()
frAgent, _ := v1beta.NewBuilder("FrenchTranslator").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()
deAgent, _ := v1beta.NewBuilder("GermanTranslator").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()

workflow, _ := v1beta.NewParallelWorkflow(
    "MultiTranslation",
    v1beta.Step("en", enAgent, "Translate to English"),
    v1beta.Step("es", esAgent, "Translate to Spanish"),
    v1beta.Step("fr", frAgent, "Translate to French"),
    v1beta.Step("de", deAgent, "Translate to German"),
)
```

### Content Validation

```go
// Validate content from multiple perspectives
grammarAgent, _ := v1beta.NewBuilder("GrammarChecker").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()
factAgent, _ := v1beta.NewBuilder("FactChecker").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()
toneAgent, _ := v1beta.NewBuilder("ToneAnalyzer").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()
plagiarismAgent, _ := v1beta.NewBuilder("PlagiarismChecker").WithConfig(&v1beta.Config{
    LLM: v1beta.LLMConfig{Provider: "openai", Model: "gpt-4"},
}).Build()

workflow, _ := v1beta.NewParallelWorkflow(
    "ContentValidation",
    v1beta.Step("grammar", grammarAgent, "Check grammar"),
    v1beta.Step("facts", factAgent, "Verify facts"),
    v1beta.Step("tone", toneAgent, "Analyze tone"),
    v1beta.Step("plagiarism", plagiarismAgent, "Check originality"),
)
```

---

## Performance Optimization

### Rate Limiting

```go
// Control concurrency to avoid rate limits
type RateLimitedWorkflow struct {
    workflow  v1beta.Workflow
    semaphore chan struct{}
}

func NewRateLimitedWorkflow(workflow v1beta.Workflow, maxConcurrent int) *RateLimitedWorkflow {
    return &RateLimitedWorkflow{
        workflow:  workflow,
        semaphore: make(chan struct{}, maxConcurrent),
    }
}

func (rlw *RateLimitedWorkflow) Run(ctx context.Context, input string) (map[string]*v1beta.Result, error) {
    // Acquire semaphore before execution
    rlw.semaphore <- struct{}{}
    defer func() { <-rlw.semaphore }()
    
    return rlw.workflow.Run(ctx, input)
}
```

### Batch Processing

```go
// Process multiple inputs with parallel workflows
func batchProcess(workflow v1beta.Workflow, inputs []string) ([]map[string]*v1beta.Result, error) {
    results := make([]map[string]*v1beta.Result, len(inputs))
    errors := make([]error, len(inputs))
    
    var wg sync.WaitGroup
    for i, input := range inputs {
        wg.Add(1)
        go func(idx int, inp string) {
            defer wg.Done()
            res, err := workflow.Run(context.Background(), inp)
            results[idx] = res
            errors[idx] = err
        }(i, input)
    }
    
    wg.Wait()
    
    // Check for errors
    for _, err := range errors {
        if err != nil {
            return results, err
        }
    }
    
    return results, nil
}
```

---

## Error Handling

### Collecting All Errors

```go
results, err := workflow.Run(ctx, input)
if err != nil {
    if pErr, ok := err.(*v1beta.ParallelWorkflowError); ok {
        // Log all failures
        for stepID, stepErr := range pErr.Errors {
            log.Printf("Step %s failed: %v", stepID, stepErr)
        }
        
        // Use partial results if acceptable
        if len(pErr.SuccessfulResults) >= 2 {
            log.Println("Using partial results...")
            results = pErr.SuccessfulResults
        } else {
            return fmt.Errorf("too many failures: %w", err)
        }
    }
}
```

### Retry Failed Steps

```go
func retryFailedSteps(workflow v1beta.Workflow, ctx context.Context, input string, maxRetries int) (map[string]*v1beta.Result, error) {
    results, err := workflow.Run(ctx, input)
    if err == nil {
        return results, nil
    }
    
    pErr, ok := err.(*v1beta.ParallelWorkflowError)
    if !ok {
        return nil, err
    }
    
    // Retry failed steps
    for i := 0; i < maxRetries; i++ {
        if len(pErr.FailedSteps) == 0 {
            break
        }
        
        // Retry only failed steps
        retryResults, retryErr := retrySteps(pErr.FailedSteps, ctx, input)
        if retryErr == nil {
            // Merge with successful results
            for k, v := range retryResults {
                pErr.SuccessfulResults[k] = v
            }
            return pErr.SuccessfulResults, nil
        }
    }
    
    return pErr.SuccessfulResults, fmt.Errorf("some steps failed after retries")
}
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

## Next Steps

- **[DAG Workflow](./workflow-dag.md)** - Complex dependencies between agents
- **[Sequential Workflow](./workflow-sequential.md)** - Step-by-step processing
- **[Loop Workflow](./workflow-loop.md)** - Iterative execution
- **[Subworkflows](./subworkflow-composition.md)** - Nested workflow patterns

---

## Related Documentation

- [Workflows Guide](../workflows.md) - Complete workflow documentation
- [Performance](../performance.md) - Optimization strategies for parallel workflows
- [Error Handling](../error-handling.md) - Handling parallel execution errors
- [Core Concepts](../core-concepts.md) - Understanding parallel execution
