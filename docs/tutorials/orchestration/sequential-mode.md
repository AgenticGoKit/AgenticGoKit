# Sequential Orchestration Mode

## Overview

Sequential orchestration enables agents to process in a defined sequence, where each agent's output becomes the next agent's input. This creates powerful processing pipelines where agents build upon each other's work to solve complex, multi-step problems.

Think of sequential orchestration like an assembly line where each worker (agent) performs a specific task and passes the work to the next worker in line, with each step adding value to the final product.

## Prerequisites

- Understanding of [Message Passing and Event Flow](../core-concepts/message-passing.md)
- Familiarity with [Route Orchestration](routing-mode.md)
- Basic knowledge of [State Management](../core-concepts/state-management.md)
- Understanding of the [Orchestration Overview](README.md)

## How Sequential Orchestration Works

### Basic Flow

```
┌─────────┐     ┌──────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│ Client  │────▶│ Sequence │────▶│ Agent A │────▶│ Agent B │────▶│ Agent C │
└─────────┘     │ Orchestr.│     └─────────┘     └─────────┘     └─────────┘
                └──────────┘                                           │
                     ▲                                                │
                     └────────────────────────────────────────────────┘
```

1. **Initial Event**: Client sends event to the first agent in the sequence
2. **Sequential Processing**: Each agent processes and passes its output to the next agent
3. **State Propagation**: The state object carries data through the entire pipeline
4. **Final Result**: The last agent's output becomes the sequence result

### State Flow Through the Pipeline

```go
// Simplified sequential processing
func (o *SequentialOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
    currentState := core.NewState()
    
    // Initialize state with event data
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    // Process through each agent in sequence
    for i, agentName := range o.sequence {
        handler := o.handlers[agentName]
        
        // Create event for this stage
        stageEvent := core.NewEvent(agentName, currentState.GetAll(), event.GetMetadata())
        
        // Run the agent
        result, err := handler.Run(ctx, stageEvent, currentState)
        if err != nil {
            return AgentResult{}, fmt.Errorf("stage %d (%s) failed: %w", i+1, agentName, err)
        }
        
        // Update state for next agent
        if result.OutputState != nil {
            currentState = result.OutputState
        }
    }
    
    return AgentResult{OutputState: currentState}, nil
}
```

## When to Use Sequential Orchestration

Sequential orchestration is ideal for:

- **Pipeline Processing**: Multi-step workflows where each step depends on the previous
- **Data Transformation**: Converting data through multiple processing stages
- **Workflow Automation**: Business processes with defined steps
- **Content Processing**: Document processing, analysis, and formatting pipelines
- **Decision Trees**: Sequential decision-making processes
- **Quality Assurance**: Multi-stage validation and refinement processes

### Use Case Examples

1. **Research Pipeline**: Collect → Analyze → Summarize → Format
2. **Content Creation**: Research → Write → Edit → Publish
3. **Data Processing**: Extract → Transform → Load → Validate
4. **Code Review**: Analyze → Test → Review → Approve
5. **Customer Support**: Classify → Route → Process → Follow-up

## Implementation Examples

### Basic Sequential Pipeline

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create pipeline agents
    collector, err := createDataCollectorAgent()
    if err != nil {
        log.Fatal(err)
    }
    
    analyzer, err := createDataAnalyzerAgent()
    if err != nil {
        log.Fatal(err)
    }
    
    formatter, err := createReportFormatterAgent()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create sequential orchestrator
    runner := core.NewRunner(100)
    orchestrator := createSequentialOrchestrator(
        runner.GetCallbackRegistry(),
        []string{"collector", "analyzer", "formatter"}, // Execution order
    )
    runner.SetOrchestrator(orchestrator)
    
    // Register agents in order
    runner.RegisterAgent("collector", collector)
    runner.RegisterAgent("analyzer", analyzer)
    runner.RegisterAgent("formatter", formatter)
    
    // Set up pipeline monitoring
    setupPipelineMonitoring(runner)
    
    // Start the runner
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Create initial event (goes to first agent in sequence)
    pipelineEvent := core.NewEvent(
        "collector",  // Start with first agent
        core.EventData{
            "source": "https://api.example.com/data",
            "format": "json",
            "query": "sales data for Q4 2023",
        },
        map[string]string{
            "session_id": "pipeline-session-123",
            "pipeline_id": "sales-report-pipeline",
        },
    )
    
    // Emit the event
    runner.Emit(pipelineEvent)
    
    // Wait for pipeline completion
    time.Sleep(30 * time.Second)
}

// Sequential orchestrator implementation
type SequentialOrchestrator struct {
    handlers map[string]core.AgentHandler
    sequence []string
    registry *core.CallbackRegistry
    mu       sync.RWMutex
}

func createSequentialOrchestrator(registry *core.CallbackRegistry, sequence []string) *SequentialOrchestrator {
    return &SequentialOrchestrator{
        handlers: make(map[string]core.AgentHandler),
        sequence: sequence,
        registry: registry,
    }
}

func (o *SequentialOrchestrator) RegisterAgent(name string, handler core.AgentHandler) error {
    o.mu.Lock()
    defer o.mu.Unlock()
    o.handlers[name] = handler
    return nil
}

func (o *SequentialOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    o.mu.RLock()
    sequence := make([]string, len(o.sequence))
    copy(sequence, o.sequence)
    handlers := make(map[string]core.AgentHandler)
    for name, handler := range o.handlers {
        handlers[name] = handler
    }
    o.mu.RUnlock()
    
    currentState := core.NewState()
    
    // Initialize state with event data
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    // Copy metadata to state
    for key, value := range event.GetMetadata() {
        currentState.SetMeta(key, value)
    }
    
    var finalResult core.AgentResult
    
    // Process through each agent in sequence
    for i, agentName := range sequence {
        handler, exists := handlers[agentName]
        if !exists {
            return core.AgentResult{}, fmt.Errorf("agent '%s' not found in sequence", agentName)
        }
        
        // Create event for this stage
        stageEvent := core.NewEvent(
            agentName,
            currentState.GetAll(),
            event.GetMetadata(),
        )
        
        // Add stage information to metadata
        stageEvent.GetMetadata()["stage_number"] = fmt.Sprintf("%d", i+1)
        stageEvent.GetMetadata()["total_stages"] = fmt.Sprintf("%d", len(sequence))
        stageEvent.GetMetadata()["stage_name"] = agentName
        
        // Execute pre-stage callbacks
        if o.registry != nil {
            callbackArgs := core.CallbackArgs{
                AgentID:     agentName,
                Event:       stageEvent,
                State:       currentState,
                StageIndex:  i,
                TotalStages: len(sequence),
            }
            
            _, err := o.registry.ExecuteCallbacks(ctx, core.HookBeforeAgentRun, callbackArgs)
            if err != nil {
                log.Printf("Pre-stage callback error: %v", err)
            }
        }
        
        // Run the agent
        startTime := time.Now()
        result, err := handler.Run(ctx, stageEvent, currentState)
        duration := time.Since(startTime)
        
        if err != nil {
            // Execute error callbacks
            if o.registry != nil {
                callbackArgs := core.CallbackArgs{
                    AgentID:     agentName,
                    Event:       stageEvent,
                    Error:       err,
                    State:       currentState,
                    StageIndex:  i,
                    TotalStages: len(sequence),
                }
                
                o.registry.ExecuteCallbacks(ctx, core.HookAgentError, callbackArgs)
            }
            
            return core.AgentResult{}, fmt.Errorf("stage %d (%s) failed: %w", i+1, agentName, err)
        }
        
        // Execute post-stage callbacks
        if o.registry != nil {
            callbackArgs := core.CallbackArgs{
                AgentID:     agentName,
                Event:       stageEvent,
                AgentResult: result,
                State:       result.OutputState,
                StageIndex:  i,
                TotalStages: len(sequence),
                Duration:    duration,
            }
            
            _, err := o.registry.ExecuteCallbacks(ctx, core.HookAfterAgentRun, callbackArgs)
            if err != nil {
                log.Printf("Post-stage callback error: %v", err)
            }
        }
        
        // Update state for next agent
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        
        // Keep track of final result
        finalResult = result
        
        fmt.Printf("Stage %d (%s) completed in %v\\n", i+1, agentName, duration)
    }
    
    // Add pipeline completion metadata
    finalResult.OutputState.SetMeta("pipeline_completed", "true")
    finalResult.OutputState.SetMeta("total_stages", fmt.Sprintf("%d", len(sequence)))
    finalResult.OutputState.SetMeta("pipeline_id", event.GetMetadata()["pipeline_id"])
    
    return finalResult, nil
}

// Agent implementations
func createDataCollectorAgent() (core.AgentHandler, error) {
    return core.NewLLMAgent("data-collector", core.LLMConfig{
        SystemPrompt: `You are a data collection specialist. Your job is to gather data from the specified source.
                      Extract relevant information and prepare it for analysis.
                      Return the collected data in a structured format.`,
        Temperature: 0.3,
        MaxTokens: 800,
    }, core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
}

func createDataAnalyzerAgent() (core.AgentHandler, error) {
    return core.NewLLMAgent("data-analyzer", core.LLMConfig{
        SystemPrompt: `You are a data analysis specialist. Analyze the collected data from the previous stage.
                      Identify patterns, trends, and key insights.
                      Prepare analytical findings for report formatting.`,
        Temperature: 0.4,
        MaxTokens: 1000,
    }, core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
}

func createReportFormatterAgent() (core.AgentHandler, error) {
    return core.NewLLMAgent("report-formatter", core.LLMConfig{
        SystemPrompt: `You are a report formatting specialist. Take the analysis from the previous stage
                      and format it into a professional, well-structured report.
                      Include executive summary, key findings, and recommendations.`,
        Temperature: 0.2,
        MaxTokens: 1200,
    }, core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
}

// Pipeline monitoring setup
func setupPipelineMonitoring(runner *core.Runner) {
    runner.RegisterCallback(core.HookBeforeAgentRun, "pipeline-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            stageNum := args.Event.GetMetadata()["stage_number"]
            totalStages := args.Event.GetMetadata()["total_stages"]
            
            fmt.Printf("[PIPELINE] Starting stage %s/%s: %s\\n", 
                stageNum, totalStages, args.AgentID)
            
            return args.State, nil
        },
    )
    
    runner.RegisterCallback(core.HookAfterAgentRun, "pipeline-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            stageNum := args.Event.GetMetadata()["stage_number"]
            totalStages := args.Event.GetMetadata()["total_stages"]
            
            fmt.Printf("[PIPELINE] Completed stage %s/%s: %s (Duration: %v)\\n", 
                stageNum, totalStages, args.AgentID, args.Duration)
            
            return args.State, nil
        },
    )
    
    runner.RegisterCallback(core.HookAgentError, "pipeline-error-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            stageNum := args.Event.GetMetadata()["stage_number"]
            
            fmt.Printf("[PIPELINE ERROR] Stage %s failed: %s - %v\\n", 
                stageNum, args.AgentID, args.Error)
            
            return args.State, nil
        },
    )
}
```

### Advanced Sequential Patterns

#### 1. Conditional Branching Pipeline

```go
type ConditionalSequentialOrchestrator struct {
    *SequentialOrchestrator
    branchingRules map[string]func(core.State) string
}

func (o *ConditionalSequentialOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    currentState := core.NewState()
    
    // Initialize state with event data
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    currentAgent := o.sequence[0] // Start with first agent
    
    for {
        handler, exists := o.handlers[currentAgent]
        if !exists {
            return core.AgentResult{}, fmt.Errorf("agent '%s' not found", currentAgent)
        }
        
        // Create event for current agent
        stageEvent := core.NewEvent(currentAgent, currentState.GetAll(), event.GetMetadata())
        
        // Run the agent
        result, err := handler.Run(ctx, stageEvent, currentState)
        if err != nil {
            return core.AgentResult{}, fmt.Errorf("agent %s failed: %w", currentAgent, err)
        }
        
        // Update state
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        finalResult = result
        
        // Determine next agent based on branching rules
        if branchingRule, exists := o.branchingRules[currentAgent]; exists {
            nextAgent := branchingRule(currentState)
            if nextAgent == "" {
                // End of pipeline
                break
            }
            currentAgent = nextAgent
        } else {
            // No branching rule, end pipeline
            break
        }
    }
    
    return finalResult, nil
}

// Example branching rule
func createContentProcessingPipeline() *ConditionalSequentialOrchestrator {
    orchestrator := &ConditionalSequentialOrchestrator{
        SequentialOrchestrator: createSequentialOrchestrator(nil, []string{"classifier"}),
        branchingRules: make(map[string]func(core.State) string),
    }
    
    // Branching rule for classifier
    orchestrator.branchingRules["classifier"] = func(state core.State) string {
        if contentType, ok := state.Get("content_type"); ok {
            switch contentType.(string) {
            case "article":
                return "article-processor"
            case "video":
                return "video-processor"
            case "image":
                return "image-processor"
            default:
                return "generic-processor"
            }
        }
        return ""
    }
    
    return orchestrator
}
```

#### 2. Parallel Sub-Pipelines

```go
type ParallelSubPipelineOrchestrator struct {
    *SequentialOrchestrator
    subPipelines map[string][]string
}

func (o *ParallelSubPipelineOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    currentState := core.NewState()
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    // Process main sequence
    for _, agentName := range o.sequence {
        if subPipeline, isSubPipeline := o.subPipelines[agentName]; isSubPipeline {
            // Execute sub-pipeline in parallel
            result, err := o.executeSubPipeline(ctx, event, currentState, subPipeline)
            if err != nil {
                return core.AgentResult{}, err
            }
            currentState = result.OutputState
        } else {
            // Execute single agent
            handler := o.handlers[agentName]
            stageEvent := core.NewEvent(agentName, currentState.GetAll(), event.GetMetadata())
            
            result, err := handler.Run(ctx, stageEvent, currentState)
            if err != nil {
                return core.AgentResult{}, fmt.Errorf("agent %s failed: %w", agentName, err)
            }
            
            if result.OutputState != nil {
                currentState = result.OutputState
            }
        }
    }
    
    return core.AgentResult{OutputState: currentState}, nil
}

func (o *ParallelSubPipelineOrchestrator) executeSubPipeline(ctx context.Context, event core.Event, state core.State, pipeline []string) (core.AgentResult, error) {
    var wg sync.WaitGroup
    results := make([]core.AgentResult, len(pipeline))
    errors := make([]error, len(pipeline))
    
    // Execute pipeline agents in parallel
    for i, agentName := range pipeline {
        wg.Add(1)
        go func(index int, name string) {
            defer wg.Done()
            
            handler := o.handlers[name]
            stageEvent := core.NewEvent(name, state.GetAll(), event.GetMetadata())
            
            result, err := handler.Run(ctx, stageEvent, state)
            results[index] = result
            errors[index] = err
        }(i, agentName)
    }
    
    wg.Wait()
    
    // Combine results
    combinedState := core.NewState()
    for i, result := range results {
        if errors[i] == nil && result.OutputState != nil {
            // Merge result into combined state
            for key, value := range result.OutputState.GetAll() {
                combinedState.Set(fmt.Sprintf("%s_%s", pipeline[i], key), value)
            }
        }
    }
    
    return core.AgentResult{OutputState: combinedState}, nil
}
```

#### 3. Retry and Recovery Pipeline

```go
type RetryableSequentialOrchestrator struct {
    *SequentialOrchestrator
    maxRetries map[string]int
    retryDelay time.Duration
}

func (o *RetryableSequentialOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    currentState := core.NewState()
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    
    // Process through each agent with retry logic
    for i, agentName := range o.sequence {
        handler := o.handlers[agentName]
        maxRetries := o.maxRetries[agentName]
        if maxRetries == 0 {
            maxRetries = 1 // Default: no retries
        }
        
        var result core.AgentResult
        var err error
        
        // Retry loop
        for attempt := 0; attempt < maxRetries; attempt++ {
            stageEvent := core.NewEvent(agentName, currentState.GetAll(), event.GetMetadata())
            stageEvent.GetMetadata()["attempt"] = fmt.Sprintf("%d", attempt+1)
            
            result, err = handler.Run(ctx, stageEvent, currentState)
            if err == nil {
                break // Success, no need to retry
            }
            
            // Log retry attempt
            fmt.Printf("Stage %d (%s) attempt %d failed: %v\\n", i+1, agentName, attempt+1, err)
            
            // Wait before retry (except for last attempt)
            if attempt < maxRetries-1 {
                time.Sleep(o.retryDelay)
            }
        }
        
        if err != nil {
            return core.AgentResult{}, fmt.Errorf("stage %d (%s) failed after %d attempts: %w", 
                i+1, agentName, maxRetries, err)
        }
        
        // Update state for next agent
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        finalResult = result
    }
    
    return finalResult, nil
}
```

## Configuration-Based Sequential Orchestration

Configure sequential orchestration through TOML:

```toml
# agentflow.toml
[orchestration]
mode = "sequential"
timeout = "120s"

[orchestration.sequential]
# Define the sequence of agents
agents = ["collector", "analyzer", "formatter"]

# Per-agent configuration
[orchestration.sequential.agents.collector]
timeout = "30s"
retries = 2
retry_delay = "5s"

[orchestration.sequential.agents.analyzer]
timeout = "60s"
retries = 1
retry_delay = "10s"

[orchestration.sequential.agents.formatter]
timeout = "30s"
retries = 0

# Pipeline-wide settings
[orchestration.sequential.pipeline]
fail_fast = true  # Stop on first failure
save_intermediate_results = true
checkpoint_frequency = 2  # Save state every 2 stages

# Conditional branching (optional)
[[orchestration.sequential.branches]]
from_agent = "classifier"
condition = "content_type == 'article'"
to_agent = "article-processor"

[[orchestration.sequential.branches]]
from_agent = "classifier"
condition = "content_type == 'video'"
to_agent = "video-processor"
```

Load and use the configuration:

```go
config, err := core.LoadConfig("agentflow.toml")
if err != nil {
    log.Fatal(err)
}

runner, err := core.NewRunnerFromConfig(config, agents)
if err != nil {
    log.Fatal(err)
}
```

## Error Handling and Recovery

### 1. Checkpoint and Resume

```go
type CheckpointSequentialOrchestrator struct {
    *SequentialOrchestrator
    checkpointStore CheckpointStore
    checkpointFreq  int
}

type Checkpoint struct {
    PipelineID   string                 `json:"pipeline_id"`
    StageIndex   int                    `json:"stage_index"`
    State        map[string]interface{} `json:"state"`
    Timestamp    time.Time              `json:"timestamp"`
}

func (o *CheckpointSequentialOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    pipelineID := event.GetMetadata()["pipeline_id"]
    
    // Try to resume from checkpoint
    checkpoint, err := o.checkpointStore.Load(pipelineID)
    if err == nil && checkpoint != nil {
        fmt.Printf("Resuming pipeline from stage %d\\n", checkpoint.StageIndex)
        return o.resumeFromCheckpoint(ctx, event, checkpoint)
    }
    
    // Start fresh pipeline
    return o.executeWithCheckpoints(ctx, event)
}

func (o *CheckpointSequentialOrchestrator) executeWithCheckpoints(ctx context.Context, event core.Event) (core.AgentResult, error) {
    currentState := core.NewState()
    pipelineID := event.GetMetadata()["pipeline_id"]
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    
    for i, agentName := range o.sequence {
        handler := o.handlers[agentName]
        stageEvent := core.NewEvent(agentName, currentState.GetAll(), event.GetMetadata())
        
        result, err := handler.Run(ctx, stageEvent, currentState)
        if err != nil {
            // Save checkpoint before failing
            checkpoint := &Checkpoint{
                PipelineID: pipelineID,
                StageIndex: i,
                State:      currentState.GetAll(),
                Timestamp:  time.Now(),
            }
            o.checkpointStore.Save(checkpoint)
            
            return core.AgentResult{}, fmt.Errorf("stage %d (%s) failed: %w", i+1, agentName, err)
        }
        
        // Update state
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        finalResult = result
        
        // Save checkpoint if needed
        if (i+1)%o.checkpointFreq == 0 {
            checkpoint := &Checkpoint{
                PipelineID: pipelineID,
                StageIndex: i + 1,
                State:      currentState.GetAll(),
                Timestamp:  time.Now(),
            }
            o.checkpointStore.Save(checkpoint)
            fmt.Printf("Checkpoint saved at stage %d\\n", i+1)
        }
    }
    
    // Clean up checkpoint on successful completion
    o.checkpointStore.Delete(pipelineID)
    
    return finalResult, nil
}
```

### 2. Rollback Mechanism

```go
type RollbackSequentialOrchestrator struct {
    *SequentialOrchestrator
    rollbackHandlers map[string]func(core.State) error
}

func (o *RollbackSequentialOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    currentState := core.NewState()
    completedStages := make([]string, 0)
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    
    for i, agentName := range o.sequence {
        handler := o.handlers[agentName]
        stageEvent := core.NewEvent(agentName, currentState.GetAll(), event.GetMetadata())
        
        result, err := handler.Run(ctx, stageEvent, currentState)
        if err != nil {
            // Rollback completed stages
            rollbackErr := o.rollbackStages(completedStages, currentState)
            if rollbackErr != nil {
                return core.AgentResult{}, fmt.Errorf("stage %d (%s) failed and rollback failed: %w (original: %v)", 
                    i+1, agentName, rollbackErr, err)
            }
            
            return core.AgentResult{}, fmt.Errorf("stage %d (%s) failed: %w", i+1, agentName, err)
        }
        
        // Update state and track completion
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        finalResult = result
        completedStages = append(completedStages, agentName)
    }
    
    return finalResult, nil
}

func (o *RollbackSequentialOrchestrator) rollbackStages(stages []string, state core.State) error {
    // Rollback in reverse order
    for i := len(stages) - 1; i >= 0; i-- {
        stageName := stages[i]
        if rollbackHandler, exists := o.rollbackHandlers[stageName]; exists {
            if err := rollbackHandler(state); err != nil {
                return fmt.Errorf("rollback failed for stage %s: %w", stageName, err)
            }
            fmt.Printf("Rolled back stage: %s\\n", stageName)
        }
    }
    return nil
}
```

### 3. Circuit Breaker for Stages

```go
type CircuitBreakerSequentialOrchestrator struct {
    *SequentialOrchestrator
    circuitBreakers map[string]*StageCircuitBreaker
}

type StageCircuitBreaker struct {
    failures    int
    maxFailures int
    resetTime   time.Time
    state       string // "closed", "open", "half-open"
    mu          sync.Mutex
}

func (cb *StageCircuitBreaker) CanExecute() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if cb.state == "open" {
        if time.Now().After(cb.resetTime) {
            cb.state = "half-open"
            return true
        }
        return false
    }
    
    return true
}

func (o *CircuitBreakerSequentialOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    currentState := core.NewState()
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    
    for i, agentName := range o.sequence {
        // Check circuit breaker
        if cb, exists := o.circuitBreakers[agentName]; exists && !cb.CanExecute() {
            return core.AgentResult{}, fmt.Errorf("stage %d (%s) circuit breaker is open", i+1, agentName)
        }
        
        handler := o.handlers[agentName]
        stageEvent := core.NewEvent(agentName, currentState.GetAll(), event.GetMetadata())
        
        result, err := handler.Run(ctx, stageEvent, currentState)
        
        // Update circuit breaker
        if cb, exists := o.circuitBreakers[agentName]; exists {
            if err != nil {
                cb.RecordFailure()
            } else {
                cb.RecordSuccess()
            }
        }
        
        if err != nil {
            return core.AgentResult{}, fmt.Errorf("stage %d (%s) failed: %w", i+1, agentName, err)
        }
        
        // Update state
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        finalResult = result
    }
    
    return finalResult, nil
}
```

## Performance Optimization

### 1. Parallel Stage Execution (Where Possible)

```go
type OptimizedSequentialOrchestrator struct {
    *SequentialOrchestrator
    parallelStages map[int][]string // stage index -> parallel agents
}

func (o *OptimizedSequentialOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    currentState := core.NewState()
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    stageIndex := 0
    
    for stageIndex < len(o.sequence) {
        if parallelAgents, isParallel := o.parallelStages[stageIndex]; isParallel {
            // Execute parallel stages
            result, err := o.executeParallelStages(ctx, event, currentState, parallelAgents)
            if err != nil {
                return core.AgentResult{}, err
            }
            currentState = result.OutputState
            finalResult = result
            stageIndex += len(parallelAgents)
        } else {
            // Execute single stage
            agentName := o.sequence[stageIndex]
            handler := o.handlers[agentName]
            stageEvent := core.NewEvent(agentName, currentState.GetAll(), event.GetMetadata())
            
            result, err := handler.Run(ctx, stageEvent, currentState)
            if err != nil {
                return core.AgentResult{}, fmt.Errorf("stage %d (%s) failed: %w", stageIndex+1, agentName, err)
            }
            
            if result.OutputState != nil {
                currentState = result.OutputState
            }
            finalResult = result
            stageIndex++
        }
    }
    
    return finalResult, nil
}
```

### 2. State Compression

```go
type CompressedStateOrchestrator struct {
    *SequentialOrchestrator
    compressor StateCompressor
}

type StateCompressor interface {
    Compress(state core.State) ([]byte, error)
    Decompress(data []byte) (core.State, error)
}

func (o *CompressedStateOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    currentState := core.NewState()
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    
    for i, agentName := range o.sequence {
        // Compress state before passing to agent (if large)
        if o.shouldCompress(currentState) {
            compressed, err := o.compressor.Compress(currentState)
            if err != nil {
                return core.AgentResult{}, fmt.Errorf("state compression failed: %w", err)
            }
            
            // Store compressed state and pass reference
            compressedState := core.NewState()
            compressedState.Set("_compressed_state", compressed)
            currentState = compressedState
        }
        
        handler := o.handlers[agentName]
        stageEvent := core.NewEvent(agentName, currentState.GetAll(), event.GetMetadata())
        
        result, err := handler.Run(ctx, stageEvent, currentState)
        if err != nil {
            return core.AgentResult{}, fmt.Errorf("stage %d (%s) failed: %w", i+1, agentName, err)
        }
        
        // Decompress if needed
        if compressed, ok := result.OutputState.Get("_compressed_state"); ok {
            decompressed, err := o.compressor.Decompress(compressed.([]byte))
            if err != nil {
                return core.AgentResult{}, fmt.Errorf("state decompression failed: %w", err)
            }
            currentState = decompressed
        } else if result.OutputState != nil {
            currentState = result.OutputState
        }
        
        finalResult = result
    }
    
    return finalResult, nil
}

func (o *CompressedStateOrchestrator) shouldCompress(state core.State) bool {
    // Simple heuristic: compress if state is large
    stateSize := len(fmt.Sprintf("%v", state.GetAll()))
    return stateSize > 10000 // 10KB threshold
}
```

## Monitoring and Metrics

### 1. Pipeline Metrics

```go
type PipelineMetrics struct {
    stageLatencies    map[string][]time.Duration
    stageSuccessRates map[string]float64
    pipelineLatencies []time.Duration
    totalPipelines    int64
    mu                sync.RWMutex
}

func (m *PipelineMetrics) RecordStageExecution(stageName string, duration time.Duration, success bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.stageLatencies == nil {
        m.stageLatencies = make(map[string][]time.Duration)
        m.stageSuccessRates = make(map[string]float64)
    }
    
    m.stageLatencies[stageName] = append(m.stageLatencies[stageName], duration)
    
    // Update success rate (simplified)
    if success {
        m.stageSuccessRates[stageName] = (m.stageSuccessRates[stageName] + 1.0) / 2.0
    } else {
        m.stageSuccessRates[stageName] = m.stageSuccessRates[stageName] / 2.0
    }
}

func (m *PipelineMetrics) RecordPipelineExecution(duration time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.pipelineLatencies = append(m.pipelineLatencies, duration)
    m.totalPipelines++
}

func (m *PipelineMetrics) GetAverageStageLatency(stageName string) time.Duration {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    latencies := m.stageLatencies[stageName]
    if len(latencies) == 0 {
        return 0
    }
    
    var total time.Duration
    for _, latency := range latencies {
        total += latency
    }
    
    return total / time.Duration(len(latencies))
}
```

### 2. Pipeline Visualization

```go
func GeneratePipelineDiagram(sequence []string, metrics *PipelineMetrics) string {
    var diagram strings.Builder
    
    diagram.WriteString("graph LR\\n")
    diagram.WriteString("    Start([Start]) --> ")
    
    for i, stage := range sequence {
        avgLatency := metrics.GetAverageStageLatency(stage)
        successRate := metrics.stageSuccessRates[stage]
        
        diagram.WriteString(fmt.Sprintf("%s[%s<br/>Avg: %v<br/>Success: %.1f%%]", 
            stage, stage, avgLatency, successRate*100))
        
        if i < len(sequence)-1 {
            diagram.WriteString(" --> ")
        }
    }
    
    diagram.WriteString(" --> End([End])\\n")
    
    return diagram.String()
}
```

## Best Practices

### 1. Pipeline Design Principles

- **Single Responsibility**: Each stage should have one clear purpose
- **Idempotency**: Stages should be safe to retry
- **State Management**: Keep state minimal and well-structured
- **Error Handling**: Plan for failures at each stage
- **Monitoring**: Instrument each stage for observability

### 2. Common Pipeline Patterns

```go
// Pattern 1: ETL Pipeline
func createETLPipeline() []string {
    return []string{
        "extractor",    // Extract data from source
        "transformer",  // Transform data format
        "validator",    // Validate data quality
        "loader",       // Load into destination
    }
}

// Pattern 2: Content Processing Pipeline
func createContentPipeline() []string {
    return []string{
        "content-fetcher",  // Fetch raw content
        "content-parser",   // Parse and structure
        "content-analyzer", // Analyze and categorize
        "content-enricher", // Add metadata
        "content-publisher", // Publish to destination
    }
}

// Pattern 3: ML Pipeline
func createMLPipeline() []string {
    return []string{
        "data-collector",   // Collect training data
        "data-preprocessor", // Clean and prepare data
        "feature-extractor", // Extract features
        "model-trainer",    // Train model
        "model-evaluator",  // Evaluate performance
        "model-deployer",   // Deploy to production
    }
}
```

### 3. Testing Sequential Orchestration

```go
func TestSequentialOrchestration(t *testing.T) {
    // Create mock agents
    agent1 := &MockAgent{
        processFunc: func(state core.State) (core.State, error) {
            newState := state.Clone()
            newState.Set("stage1_result", "processed by agent1")
            return newState, nil
        },
    }
    
    agent2 := &MockAgent{
        processFunc: func(state core.State) (core.State, error) {
            newState := state.Clone()
            stage1Result, _ := state.Get("stage1_result")
            newState.Set("stage2_result", fmt.Sprintf("processed %s by agent2", stage1Result))
            return newState, nil
        },
    }
    
    // Create orchestrator
    orchestrator := createSequentialOrchestrator(nil, []string{"agent1", "agent2"})
    orchestrator.RegisterAgent("agent1", agent1)
    orchestrator.RegisterAgent("agent2", agent2)
    
    // Create test event
    event := core.NewEvent(
        "agent1",
        core.EventData{"input": "test data"},
        map[string]string{"session_id": "test"},
    )
    
    // Test pipeline
    result, err := orchestrator.Dispatch(context.Background(), event)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    // Verify pipeline execution
    stage1Result, ok := result.OutputState.Get("stage1_result")
    assert.True(t, ok)
    assert.Equal(t, "processed by agent1", stage1Result)
    
    stage2Result, ok := result.OutputState.Get("stage2_result")
    assert.True(t, ok)
    assert.Contains(t, stage2Result.(string), "processed by agent2")
}
```

## Common Pitfalls and Solutions

### 1. State Bloat

**Problem**: State grows too large as it passes through the pipeline.

**Solution**: Implement state pruning and compression:

```go
func pruneState(state core.State, keepKeys []string) core.State {
    prunedState := core.NewState()
    
    for _, key := range keepKeys {
        if value, ok := state.Get(key); ok {
            prunedState.Set(key, value)
        }
    }
    
    // Copy metadata
    for key, value := range state.GetMetadata() {
        prunedState.SetMeta(key, value)
    }
    
    return prunedState
}
```

### 2. Pipeline Deadlocks

**Problem**: Agents waiting for resources or dependencies can cause deadlocks.

**Solution**: Implement timeouts and resource management:

```go
func (o *SequentialOrchestrator) runAgentWithTimeout(ctx context.Context, handler core.AgentHandler, event core.Event, state core.State, timeout time.Duration) (core.AgentResult, error) {
    timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    resultChan := make(chan core.AgentResult, 1)
    errorChan := make(chan error, 1)
    
    go func() {
        result, err := handler.Run(timeoutCtx, event, state)
        if err != nil {
            errorChan <- err
        } else {
            resultChan <- result
        }
    }()
    
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errorChan:
        return core.AgentResult{}, err
    case <-timeoutCtx.Done():
        return core.AgentResult{}, fmt.Errorf("agent execution timeout")
    }
}
```

### 3. Inconsistent State Formats

**Problem**: Different agents expect different state formats.

**Solution**: Implement state adapters:

```go
type StateAdapter interface {
    AdaptInput(state core.State) core.State
    AdaptOutput(state core.State) core.State
}

type AgentWithAdapter struct {
    agent   core.AgentHandler
    adapter StateAdapter
}

func (a *AgentWithAdapter) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Adapt input state
    adaptedState := a.adapter.AdaptInput(state)
    
    // Run agent
    result, err := a.agent.Run(ctx, event, adaptedState)
    if err != nil {
        return result, err
    }
    
    // Adapt output state
    if result.OutputState != nil {
        result.OutputState = a.adapter.AdaptOutput(result.OutputState)
    }
    
    return result, nil
}
```

## Conclusion

Sequential orchestration provides a powerful foundation for building complex, multi-step agent workflows. By understanding the patterns, implementing proper error handling, and following best practices, you can create robust pipeline systems that solve sophisticated problems.

Key takeaways:
- Sequential orchestration excels at multi-step workflows with dependencies
- Proper error handling and recovery mechanisms are essential
- State management is crucial for pipeline performance
- Monitoring and metrics help optimize pipeline performance
- Design stages to be independent and idempotent when possible

## Next Steps

- [Loop Orchestration](loop-mode.md) - Learn iterative processing patterns
- [Mixed Orchestration](mixed-mode.md) - Combine multiple orchestration patterns
- [Error Handling](../core-concepts/error-handling.md) - Master robust error management
- [State Management](../core-concepts/state-management.md) - Optimize data flow through pipelines

## Further Reading

- [API Reference: SequentialOrchestrator](../../api/core.md#sequential-orchestrator)
- [Examples: Sequential Pipelines](../../examples/03-sequential-pipeline/)
- [Production Guide: Pipeline Monitoring](../../production/monitoring.md#pipeline-monitoring)