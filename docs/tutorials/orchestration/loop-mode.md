# Loop Orchestration Mode

## Overview

Loop orchestration enables a single agent to process repeatedly until a condition is met or a maximum number of iterations is reached. This pattern is perfect for iterative refinement, quality improvement, and self-correcting processes.

Think of loop orchestration like a craftsperson perfecting their work - they keep refining and improving until they achieve the desired quality or reach their time limit.

## Prerequisites

- Understanding of [Message Passing and Event Flow](../core-concepts/message-passing.md)
- Familiarity with [Route Orchestration](routing-mode.md)
- Basic knowledge of [State Management](../core-concepts/state-management.md)
- Understanding of the [Orchestration Overview](README.md)

## How Loop Orchestration Works

### Basic Flow

```
┌─────────┐     ┌──────────┐     ┌─────────┐
│ Client  │────▶│  Loop    │────▶│ Agent A │
└─────────┘     │ Orchestr.│     └─────────┘
                └──────────┘         │
                     ▲               │
                     └───────────────┘
                    [until condition met]
```

1. **Initial Processing**: Agent processes the initial event
2. **Condition Check**: Loop condition evaluates the result
3. **Iteration Decision**: If condition not met, agent processes its own output
4. **Termination**: Process repeats until condition met or max iterations reached

### Loop Processing Flow

```go
// Simplified loop orchestration logic
func (o *LoopOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
    currentState := core.NewState()
    
    // Initialize state with event data
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult AgentResult
    iteration := 0
    
    for iteration < o.maxIterations {
        // Check condition
        if o.condition(currentState) {
            fmt.Printf("Loop condition met after %d iterations\n", iteration)
            break
        }
        
        // Run the agent
        result, err := o.handler.Run(ctx, event, currentState)
        if err != nil {
            return AgentResult{}, fmt.Errorf("iteration %d failed: %w", iteration, err)
        }
        
        // Update state for next iteration
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        
        finalResult = result
        iteration++
    }
    
    return finalResult, nil
}
```

## When to Use Loop Orchestration

Loop orchestration is ideal for:

- **Iterative Refinement**: Continuously improving content or solutions
- **Quality Improvement**: Processing until quality threshold is met
- **Self-Correction**: Fixing errors through repeated attempts
- **Optimization**: Finding better solutions through iteration
- **Convergence**: Reaching a stable state through repeated processing
- **Learning**: Improving performance through practice

### Use Case Examples

1. **Content Refinement**: Improve writing quality through multiple revisions
2. **Code Optimization**: Optimize code performance through iterative improvements
3. **Problem Solving**: Refine solutions until they meet requirements
4. **Data Cleaning**: Clean data until error rate is acceptable
5. **Model Training**: Train models until accuracy threshold is reached

## Implementation Examples

### Basic Loop Orchestration

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
    // Create a refining agent
    refiner, err := createContentRefinerAgent()
    if err != nil {
        log.Fatal(err)
    }
    
    // Define loop condition
    condition := func(state core.State) bool {
        // Check if quality score exceeds threshold
        if score, ok := state.Get("quality_score"); ok {
            return score.(float64) >= 0.9
        }
        return false
    }
    
    // Create loop orchestrator
    runner := core.NewRunner(100)
    orchestrator := createLoopOrchestrator(
        runner.GetCallbackRegistry(),
        "refiner",
        condition,
        5, // max iterations
    )
    runner.SetOrchestrator(orchestrator)
    
    // Register the agent
    runner.RegisterAgent("refiner", refiner)
    
    // Set up loop monitoring
    setupLoopMonitoring(runner)
    
    // Start runner
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Create event
    event := core.NewEvent(
        "refiner",
        core.EventData{
            "content": "Initial draft content that needs improvement...",
            "quality_score": 0.3,
            "target_quality": 0.9,
        },
        map[string]string{
            "session_id": "loop-session-123",
            "loop_type": "content_refinement",
        },
    )
    
    // Emit the event
    runner.Emit(event)
    
    // Wait for loop completion
    time.Sleep(60 * time.Second)
}

// Loop orchestrator implementation
type LoopOrchestrator struct {
    agentName     string
    handler       core.AgentHandler
    condition     func(core.State) bool
    maxIterations int
    registry      *core.CallbackRegistry
    mu            sync.RWMutex
}

func createLoopOrchestrator(registry *core.CallbackRegistry, agentName string, condition func(core.State) bool, maxIterations int) *LoopOrchestrator {
    return &LoopOrchestrator{
        agentName:     agentName,
        condition:     condition,
        maxIterations: maxIterations,
        registry:      registry,
    }
}

func (o *LoopOrchestrator) RegisterAgent(name string, handler core.AgentHandler) error {
    o.mu.Lock()
    defer o.mu.Unlock()
    
    if name == o.agentName {
        o.handler = handler
        return nil
    }
    
    return fmt.Errorf("loop orchestrator only accepts agent '%s', got '%s'", o.agentName, name)
}

func (o *LoopOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    if o.handler == nil {
        return core.AgentResult{}, fmt.Errorf("loop agent '%s' not registered", o.agentName)
    }
    
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
    iteration := 0
    
    for iteration < o.maxIterations {
        // Check condition before processing (except first iteration)
        if iteration > 0 && o.condition(currentState) {
            fmt.Printf("Loop condition met after %d iterations\n", iteration)
            break
        }
        
        // Create event for this iteration
        iterationEvent := core.NewEvent(
            o.agentName,
            currentState.GetAll(),
            event.GetMetadata(),
        )
        iterationEvent.GetMetadata()["loop_iteration"] = fmt.Sprintf("%d", iteration)
        iterationEvent.GetMetadata()["max_iterations"] = fmt.Sprintf("%d", o.maxIterations)
        
        // Execute pre-iteration callbacks
        if o.registry != nil {
            callbackArgs := core.CallbackArgs{
                AgentID:     o.agentName,
                Event:       iterationEvent,
                State:       currentState,
                Iteration:   iteration,
                MaxIterations: o.maxIterations,
            }
            
            _, err := o.registry.ExecuteCallbacks(ctx, core.HookBeforeAgentRun, callbackArgs)
            if err != nil {
                log.Printf("Pre-iteration callback error: %v", err)
            }
        }
        
        // Run the agent
        startTime := time.Now()
        result, err := o.handler.Run(ctx, iterationEvent, currentState)
        duration := time.Since(startTime)
        
        if err != nil {
            // Execute error callbacks
            if o.registry != nil {
                callbackArgs := core.CallbackArgs{
                    AgentID:       o.agentName,
                    Event:         iterationEvent,
                    Error:         err,
                    State:         currentState,
                    Iteration:     iteration,
                    MaxIterations: o.maxIterations,
                }
                
                o.registry.ExecuteCallbacks(ctx, core.HookAgentError, callbackArgs)
            }
            
            return core.AgentResult{}, fmt.Errorf("loop iteration %d failed: %w", iteration, err)
        }
        
        // Execute post-iteration callbacks
        if o.registry != nil {
            callbackArgs := core.CallbackArgs{
                AgentID:       o.agentName,
                Event:         iterationEvent,
                AgentResult:   result,
                State:         result.OutputState,
                Iteration:     iteration,
                MaxIterations: o.maxIterations,
                Duration:      duration,
            }
            
            _, err := o.registry.ExecuteCallbacks(ctx, core.HookAfterAgentRun, callbackArgs)
            if err != nil {
                log.Printf("Post-iteration callback error: %v", err)
            }
        }
        
        // Update state for next iteration
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        
        finalResult = result
        iteration++
        
        fmt.Printf("Loop iteration %d completed\n", iteration)
    }
    
    // Check final condition
    conditionMet := o.condition(currentState)
    
    if iteration >= o.maxIterations && !conditionMet {
        fmt.Printf("Loop terminated after max iterations (%d) without meeting condition\n", o.maxIterations)
    }
    
    // Add loop completion metadata
    finalResult.OutputState.SetMeta("loop_completed", "true")
    finalResult.OutputState.SetMeta("total_iterations", fmt.Sprintf("%d", iteration))
    finalResult.OutputState.SetMeta("condition_met", fmt.Sprintf("%t", conditionMet))
    finalResult.OutputState.SetMeta("max_iterations_reached", fmt.Sprintf("%t", iteration >= o.maxIterations))
    
    return finalResult, nil
}

// Agent implementation
func createContentRefinerAgent() (core.AgentHandler, error) {
    return core.NewLLMAgent("content-refiner", core.LLMConfig{
        SystemPrompt: `You are a content refinement specialist. Your job is to improve the quality, 
                      clarity, and coherence of the given content. 
                      
                      For each iteration:
                      1. Analyze the current content quality
                      2. Identify specific areas for improvement
                      3. Make targeted improvements
                      4. Assign a quality score from 0.0 to 1.0
                      
                      Focus on:
                      - Clarity and readability
                      - Logical flow and structure
                      - Grammar and style
                      - Completeness and accuracy
                      
                      Always include a "quality_score" in your response.`,
        Temperature: 0.3,
        MaxTokens: 1000,
    }, core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
}

// Loop monitoring setup
func setupLoopMonitoring(runner *core.Runner) {
    runner.RegisterCallback(core.HookBeforeAgentRun, "loop-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            iteration := args.Event.GetMetadata()["loop_iteration"]
            maxIterations := args.Event.GetMetadata()["max_iterations"]
            
            fmt.Printf("[LOOP] Starting iteration %s/%s for agent %s\n", 
                iteration, maxIterations, args.AgentID)
            
            return args.State, nil
        },
    )
    
    runner.RegisterCallback(core.HookAfterAgentRun, "loop-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            iteration := args.Event.GetMetadata()["loop_iteration"]
            
            // Extract quality score if available
            qualityScore := "unknown"
            if score, ok := args.AgentResult.OutputState.Get("quality_score"); ok {
                qualityScore = fmt.Sprintf("%.2f", score.(float64))
            }
            
            fmt.Printf("[LOOP] Completed iteration %s for agent %s (Quality: %s, Duration: %v)\n", 
                iteration, args.AgentID, qualityScore, args.Duration)
            
            return args.State, nil
        },
    )
    
    runner.RegisterCallback(core.HookAgentError, "loop-error-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            iteration := args.Event.GetMetadata()["loop_iteration"]
            
            fmt.Printf("[LOOP ERROR] Iteration %s failed for agent %s: %v\n", 
                iteration, args.AgentID, args.Error)
            
            return args.State, nil
        },
    )
}
```

### Advanced Loop Patterns

#### 1. Adaptive Loop Conditions

```go
// Create adaptive conditions that change based on context
func createAdaptiveCondition(targetQuality float64) func(core.State) bool {
    return func(state core.State) bool {
        // Get current quality score
        currentQuality, ok := state.Get("quality_score")
        if !ok {
            return false
        }
        
        // Get iteration count
        iteration, _ := state.GetMeta("current_iteration")
        iterationNum, _ := strconv.Atoi(iteration)
        
        // Adaptive threshold - lower requirements for later iterations
        adaptiveThreshold := targetQuality
        if iterationNum > 3 {
            adaptiveThreshold = targetQuality * 0.9 // 90% of target after 3 iterations
        }
        
        return currentQuality.(float64) >= adaptiveThreshold
    }
}
```

#### 2. Multi-Criteria Loop Conditions

```go
// Create complex conditions with multiple criteria
func createMultiCriteriaCondition() func(core.State) bool {
    return func(state core.State) bool {
        // Check quality score
        qualityScore, hasQuality := state.Get("quality_score")
        if !hasQuality || qualityScore.(float64) < 0.8 {
            return false
        }
        
        // Check completeness
        completeness, hasCompleteness := state.Get("completeness_score")
        if !hasCompleteness || completeness.(float64) < 0.9 {
            return false
        }
        
        // Check error count
        errorCount, hasErrors := state.Get("error_count")
        if !hasErrors || errorCount.(int) > 0 {
            return false
        }
        
        // All criteria met
        return true
    }
}
```

#### 3. Time-Based Loop Termination

```go
type TimeBoundLoopOrchestrator struct {
    *LoopOrchestrator
    maxDuration time.Duration
    startTime   time.Time
}

func (o *TimeBoundLoopOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    o.startTime = time.Now()
    
    // Create context with timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, o.maxDuration)
    defer cancel()
    
    currentState := core.NewState()
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    iteration := 0
    
    for iteration < o.maxIterations {
        // Check time limit
        if time.Since(o.startTime) >= o.maxDuration {
            fmt.Printf("Loop terminated due to time limit after %d iterations\n", iteration)
            break
        }
        
        // Check condition
        if iteration > 0 && o.condition(currentState) {
            fmt.Printf("Loop condition met after %d iterations\n", iteration)
            break
        }
        
        // Create iteration event
        iterationEvent := core.NewEvent(
            o.agentName,
            currentState.GetAll(),
            event.GetMetadata(),
        )
        
        // Run agent with timeout context
        result, err := o.handler.Run(timeoutCtx, iterationEvent, currentState)
        if err != nil {
            if timeoutCtx.Err() == context.DeadlineExceeded {
                return core.AgentResult{}, fmt.Errorf("loop iteration %d timed out", iteration)
            }
            return core.AgentResult{}, fmt.Errorf("loop iteration %d failed: %w", iteration, err)
        }
        
        // Update state
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        finalResult = result
        iteration++
    }
    
    // Add timing metadata
    finalResult.OutputState.SetMeta("total_duration", time.Since(o.startTime).String())
    finalResult.OutputState.SetMeta("time_limit_reached", fmt.Sprintf("%t", time.Since(o.startTime) >= o.maxDuration))
    
    return finalResult, nil
}
```

## Configuration-Based Loop Orchestration

Configure loop orchestration through TOML:

```toml
# agentflow.toml
[orchestration]
mode = "loop"
timeout = "300s"

[orchestration.loop]
agent = "content-refiner"
max_iterations = 5
max_duration = "120s"

# Condition configuration
[orchestration.loop.condition]
type = "quality_threshold"
threshold = 0.85
field = "quality_score"

# Alternative: multi-criteria condition
# [orchestration.loop.condition]
# type = "multi_criteria"
# [[orchestration.loop.condition.criteria]]
# field = "quality_score"
# operator = ">="
# value = 0.8
# [[orchestration.loop.condition.criteria]]
# field = "error_count"
# operator = "<="
# value = 0

# Adaptive condition
# [orchestration.loop.condition]
# type = "adaptive"
# base_threshold = 0.9
# decay_rate = 0.1
# decay_after_iteration = 3
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

### 1. Iteration-Level Error Recovery

```go
type RecoverableLoopOrchestrator struct {
    *LoopOrchestrator
    recoveryStrategy func(error, int) (core.AgentResult, error)
    maxFailures      int
}

func (o *RecoverableLoopOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    currentState := core.NewState()
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    iteration := 0
    failures := 0
    
    for iteration < o.maxIterations && failures < o.maxFailures {
        // Check condition
        if iteration > 0 && o.condition(currentState) {
            break
        }
        
        // Create iteration event
        iterationEvent := core.NewEvent(
            o.agentName,
            currentState.GetAll(),
            event.GetMetadata(),
        )
        
        // Run agent
        result, err := o.handler.Run(ctx, iterationEvent, currentState)
        if err != nil {
            failures++
            fmt.Printf("Iteration %d failed (failure %d/%d): %v\n", iteration, failures, o.maxFailures, err)
            
            // Try recovery
            if o.recoveryStrategy != nil {
                recoveredResult, recoveryErr := o.recoveryStrategy(err, iteration)
                if recoveryErr == nil {
                    result = recoveredResult
                    err = nil
                    fmt.Printf("Recovery successful for iteration %d\n", iteration)
                }
            }
            
            if err != nil {
                if failures >= o.maxFailures {
                    return core.AgentResult{}, fmt.Errorf("loop failed after %d failures", failures)
                }
                continue // Skip this iteration
            }
        }
        
        // Update state
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        finalResult = result
        iteration++
    }
    
    return finalResult, nil
}
```

### 2. Checkpoint and Resume

```go
type CheckpointLoopOrchestrator struct {
    *LoopOrchestrator
    checkpointStore CheckpointStore
    checkpointFreq  int
}

type LoopCheckpoint struct {
    LoopID    string                 `json:"loop_id"`
    Iteration int                    `json:"iteration"`
    State     map[string]interface{} `json:"state"`
    Timestamp time.Time              `json:"timestamp"`
}

func (o *CheckpointLoopOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    loopID := event.GetMetadata()["loop_id"]
    
    // Try to resume from checkpoint
    checkpoint, err := o.checkpointStore.LoadLoopCheckpoint(loopID)
    if err == nil && checkpoint != nil {
        return o.resumeFromCheckpoint(ctx, event, checkpoint)
    }
    
    // Start fresh loop with checkpointing
    return o.executeWithCheckpoints(ctx, event)
}

func (o *CheckpointLoopOrchestrator) executeWithCheckpoints(ctx context.Context, event core.Event) (core.AgentResult, error) {
    currentState := core.NewState()
    loopID := event.GetMetadata()["loop_id"]
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    iteration := 0
    
    for iteration < o.maxIterations {
        // Save checkpoint if needed
        if iteration%o.checkpointFreq == 0 {
            checkpoint := &LoopCheckpoint{
                LoopID:    loopID,
                Iteration: iteration,
                State:     currentState.GetAll(),
                Timestamp: time.Now(),
            }
            o.checkpointStore.SaveLoopCheckpoint(checkpoint)
            fmt.Printf("Checkpoint saved at iteration %d\n", iteration)
        }
        
        // Check condition
        if iteration > 0 && o.condition(currentState) {
            break
        }
        
        // Create iteration event
        iterationEvent := core.NewEvent(
            o.agentName,
            currentState.GetAll(),
            event.GetMetadata(),
        )
        
        // Run agent
        result, err := o.handler.Run(ctx, iterationEvent, currentState)
        if err != nil {
            return core.AgentResult{}, fmt.Errorf("loop iteration %d failed: %w", iteration, err)
        }
        
        // Update state
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        finalResult = result
        iteration++
    }
    
    // Clean up checkpoint on success
    o.checkpointStore.DeleteLoopCheckpoint(loopID)
    
    return finalResult, nil
}
```

## Performance Optimization

### 1. Early Termination Strategies

```go
// Implement early termination based on convergence
func createConvergenceCondition(tolerance float64) func(core.State) bool {
    var previousScore float64
    var stableIterations int
    
    return func(state core.State) bool {
        currentScore, ok := state.Get("quality_score")
        if !ok {
            return false
        }
        
        score := currentScore.(float64)
        
        // Check if score has stabilized
        if math.Abs(score-previousScore) < tolerance {
            stableIterations++
        } else {
            stableIterations = 0
        }
        
        previousScore = score
        
        // Terminate if stable for 2 iterations
        return stableIterations >= 2
    }
}
```

### 2. Resource Management

```go
type ResourceManagedLoopOrchestrator struct {
    *LoopOrchestrator
    resourcePool *ResourcePool
    maxMemory    int64
}

func (o *ResourceManagedLoopOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    // Monitor resource usage
    var memStats runtime.MemStats
    
    currentState := core.NewState()
    
    // Initialize state
    for key, value := range event.GetData() {
        currentState.Set(key, value)
    }
    
    var finalResult core.AgentResult
    iteration := 0
    
    for iteration < o.maxIterations {
        // Check memory usage
        runtime.ReadMemStats(&memStats)
        if memStats.Alloc > uint64(o.maxMemory) {
            fmt.Printf("Memory limit exceeded, terminating loop at iteration %d\n", iteration)
            break
        }
        
        // Check condition
        if iteration > 0 && o.condition(currentState) {
            break
        }
        
        // Acquire resources
        resources, err := o.resourcePool.Acquire()
        if err != nil {
            return core.AgentResult{}, fmt.Errorf("failed to acquire resources: %w", err)
        }
        
        // Create iteration event
        iterationEvent := core.NewEvent(
            o.agentName,
            currentState.GetAll(),
            event.GetMetadata(),
        )
        
        // Run agent
        result, err := o.handler.Run(ctx, iterationEvent, currentState)
        
        // Release resources
        o.resourcePool.Release(resources)
        
        if err != nil {
            return core.AgentResult{}, fmt.Errorf("loop iteration %d failed: %w", iteration, err)
        }
        
        // Update state
        if result.OutputState != nil {
            currentState = result.OutputState
        }
        finalResult = result
        iteration++
        
        // Force garbage collection periodically
        if iteration%5 == 0 {
            runtime.GC()
        }
    }
    
    return finalResult, nil
}
```

## Monitoring and Metrics

### 1. Loop Performance Metrics

```go
type LoopMetrics struct {
    totalLoops        int64
    totalIterations   int64
    averageIterations float64
    conditionMetRate  float64
    timeoutRate       float64
    mu                sync.RWMutex
}

func (m *LoopMetrics) RecordLoop(iterations int, conditionMet bool, timedOut bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.totalLoops++
    m.totalIterations += int64(iterations)
    m.averageIterations = float64(m.totalIterations) / float64(m.totalLoops)
    
    if conditionMet {
        m.conditionMetRate = (m.conditionMetRate*float64(m.totalLoops-1) + 1) / float64(m.totalLoops)
    } else {
        m.conditionMetRate = (m.conditionMetRate * float64(m.totalLoops-1)) / float64(m.totalLoops)
    }
    
    if timedOut {
        m.timeoutRate = (m.timeoutRate*float64(m.totalLoops-1) + 1) / float64(m.totalLoops)
    } else {
        m.timeoutRate = (m.timeoutRate * float64(m.totalLoops-1)) / float64(m.totalLoops)
    }
}

func (m *LoopMetrics) GetStats() map[string]interface{} {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    return map[string]interface{}{
        "total_loops":         m.totalLoops,
        "total_iterations":    m.totalIterations,
        "average_iterations":  m.averageIterations,
        "condition_met_rate":  m.conditionMetRate,
        "timeout_rate":        m.timeoutRate,
    }
}
```

### 2. Real-time Loop Visualization

```go
func GenerateLoopProgressVisualization(currentIteration, maxIterations int, qualityScore float64) string {
    progress := float64(currentIteration) / float64(maxIterations)
    progressBar := strings.Repeat("█", int(progress*20)) + strings.Repeat("░", 20-int(progress*20))
    
    return fmt.Sprintf(`
Loop Progress:
[%s] %d/%d iterations (%.1f%%)
Quality Score: %.2f
`, progressBar, currentIteration, maxIterations, progress*100, qualityScore)
}
```

## Best Practices

### 1. Loop Design Principles

- **Clear Termination Conditions**: Define specific, measurable conditions
- **Reasonable Iteration Limits**: Set appropriate maximum iterations
- **Progress Monitoring**: Track progress and quality improvements
- **Resource Management**: Monitor and limit resource usage
- **Error Handling**: Implement robust error recovery

### 2. Common Loop Patterns

```go
// Pattern 1: Quality Improvement Loop
func CreateQualityImprovementLoop(targetQuality float64) func(core.State) bool {
    return func(state core.State) bool {
        if quality, ok := state.Get("quality_score"); ok {
            return quality.(float64) >= targetQuality
        }
        return false
    }
}

// Pattern 2: Error Reduction Loop
func CreateErrorReductionLoop(maxErrors int) func(core.State) bool {
    return func(state core.State) bool {
        if errorCount, ok := state.Get("error_count"); ok {
            return errorCount.(int) <= maxErrors
        }
        return false
    }
}

// Pattern 3: Convergence Loop
func CreateConvergenceLoop(tolerance float64) func(core.State) bool {
    var previousValue float64
    var initialized bool
    
    return func(state core.State) bool {
        if value, ok := state.Get("convergence_value"); ok {
            currentValue := value.(float64)
            
            if !initialized {
                previousValue = currentValue
                initialized = true
                return false
            }
            
            converged := math.Abs(currentValue-previousValue) < tolerance
            previousValue = currentValue
            return converged
        }
        return false
    }
}
```

### 3. Testing Loop Orchestration

```go
func TestLoopOrchestration(t *testing.T) {
    // Create test condition
    condition := func(state core.State) bool {
        if count, ok := state.Get("count"); ok {
            return count.(int) >= 3
        }
        return false
    }
    
    // Create test agent
    testAgent := &MockLoopAgent{}
    
    // Create orchestrator
    orchestrator := createLoopOrchestrator(nil, "test-agent", condition, 5)
    orchestrator.RegisterAgent("test-agent", testAgent)
    
    // Create test event
    event := core.NewEvent(
        "test-agent",
        core.EventData{"count": 0},
        map[string]string{"loop_id": "test-123"},
    )
    
    // Execute loop
    result, err := orchestrator.Dispatch(context.Background(), event)
    
    // Verify results
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "true", result.OutputState.GetMeta("condition_met"))
    assert.Equal(t, 3, testAgent.GetCallCount())
}
```

## Conclusion

Loop orchestration provides powerful capabilities for iterative improvement and refinement processes. By understanding how to design effective loop conditions and manage iterations, you can create agents that continuously improve their output quality.

Key takeaways:
- Loop orchestration is perfect for iterative refinement tasks
- Design clear, measurable termination conditions
- Implement appropriate error handling and recovery
- Monitor resource usage and performance
- Use checkpointing for long-running loops

Loop orchestration enables sophisticated self-improving agent behaviors that can achieve high-quality results through iteration and refinement.

## Next Steps

- [Mixed Orchestration](mixed-mode.md) - Combine loop with other patterns
- [State Management](../core-concepts/state-management.md) - Master data flow in loops
- [Error Handling](../core-concepts/error-handling.md) - Implement robust loop error management
- [Performance Optimization](../advanced-patterns/load-balancing.md) - Optimize loop performance

## Further Reading

- [API Reference: Loop Orchestrator](../../api/core.md#loop-orchestrator)
- [Examples: Loop Orchestration](../../examples/03-sequential-pipeline/)
- [Configuration Guide: Loop Configuration](../../configuration/orchestration.md#loop)