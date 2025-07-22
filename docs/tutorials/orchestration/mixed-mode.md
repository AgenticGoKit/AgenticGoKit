# Mixed-Mode Orchestration in AgenticGoKit

## Overview

Mixed-mode orchestration combines multiple orchestration patterns within a single workflow, allowing you to create sophisticated agent systems that can handle complex, multi-stage processes. This tutorial covers how to design and implement mixed-mode orchestrations that leverage the strengths of different patterns.

Mixed-mode orchestration is ideal for real-world applications where different parts of a workflow require different coordination strategies.

## Prerequisites

- Understanding of [Orchestration Overview](README.md)
- Familiarity with [Route Orchestration](routing-mode.md)
- Knowledge of [Collaborative Orchestration](collaborative-mode.md)
- Understanding of [Sequential Orchestration](sequential-mode.md)
- Basic knowledge of [Loop Orchestration](loop-mode.md)

## Mixed-Mode Concepts

### What is Mixed-Mode Orchestration?

Mixed-mode orchestration allows you to combine different orchestration patterns within a single workflow:

- **Sequential + Collaborative**: Process data sequentially, then collaborate on analysis
- **Route + Loop**: Route to different agents, with some operating in loops
- **Collaborative + Sequential**: Collaborate on planning, then execute sequentially
- **All Patterns**: Complex workflows using all orchestration types

### Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Initial       │───▶│   Sequential     │───▶│  Collaborative  │
│   Routing       │    │   Processing     │    │   Analysis      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Loop          │    │   Final          │    │   Output        │
│   Refinement    │    │   Synthesis      │    │   Generation    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Basic Mixed-Mode Implementation

### 1. Sequential + Collaborative Pattern

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

// SequentialCollaborativeOrchestrator combines sequential and collaborative patterns
type SequentialCollaborativeOrchestrator struct {
    sequentialAgents     []string
    collaborativeAgents  []string
    runner              *core.Runner
    sequentialTimeout   time.Duration
    collaborativeTimeout time.Duration
}

func NewSequentialCollaborativeOrchestrator(
    sequentialAgents, collaborativeAgents []string,
    runner *core.Runner,
) *SequentialCollaborativeOrchestrator {
    return &SequentialCollaborativeOrchestrator{
        sequentialAgents:      sequentialAgents,
        collaborativeAgents:   collaborativeAgents,
        runner:               runner,
        sequentialTimeout:    30 * time.Second,
        collaborativeTimeout: 60 * time.Second,
    }
}

func (sco *SequentialCollaborativeOrchestrator) Execute(ctx context.Context, event core.Event, state core.State) (core.State, error) {
    // Phase 1: Sequential Processing
    fmt.Println("Phase 1: Sequential Processing")
    sequentialState, err := sco.executeSequential(ctx, event, state)
    if err != nil {
        return state, fmt.Errorf("sequential phase failed: %w", err)
    }
    
    // Phase 2: Collaborative Analysis
    fmt.Println("Phase 2: Collaborative Analysis")
    finalState, err := sco.executeCollaborative(ctx, event, sequentialState)
    if err != nil {
        return sequentialState, fmt.Errorf("collaborative phase failed: %w", err)
    }
    
    return finalState, nil
}

func (sco *SequentialCollaborativeOrchestrator) executeSequential(ctx context.Context, event core.Event, state core.State) (core.State, error) {
    currentState := state.Clone()
    
    for i, agentName := range sco.sequentialAgents {
        fmt.Printf("  Sequential Step %d: %s\n", i+1, agentName)
        
        // Create event for this agent
        agentEvent := core.NewEvent(
            agentName,
            event.Data,
            map[string]string{
                "phase":     "sequential",
                "step":      fmt.Sprintf("%d", i+1),
                "session_id": event.GetSessionID(),
            },
        )
        
        // Execute agent with timeout
        ctx, cancel := context.WithTimeout(ctx, sco.sequentialTimeout)
        result, err := sco.executeAgentWithTimeout(ctx, agentEvent, currentState)
        cancel()
        
        if err != nil {
            return currentState, fmt.Errorf("agent %s failed: %w", agentName, err)
        }
        
        // Update state with result
        currentState = result.OutputState
        
        // Add step metadata
        currentState.Set(fmt.Sprintf("sequential_step_%d_result", i+1), result)
        currentState.Set(fmt.Sprintf("sequential_step_%d_agent", i+1), agentName)
    }
    
    currentState.Set("sequential_phase_complete", true)
    return currentState, nil
}

func (sco *SequentialCollaborativeOrchestrator) executeCollaborative(ctx context.Context, event core.Event, state core.State) (core.State, error) {
    // Create collaborative orchestrator
    collaborativeOrchestrator := core.NewCollaborativeOrchestrator(
        sco.runner.GetCallbackRegistry(),
        sco.collaborativeTimeout,
    )
    
    // Set up collaborative agents
    for _, agentName := range sco.collaborativeAgents {
        collaborativeOrchestrator.AddAgent(agentName)
    }
    
    // Create collaborative event
    collaborativeEvent := core.NewEvent(
        "collaborative",
        event.Data,
        map[string]string{
            "phase":      "collaborative",
            "session_id": event.GetSessionID(),
        },
    )
    
    // Execute collaborative phase
    return collaborativeOrchestrator.Execute(ctx, collaborativeEvent, state)
}

func (sco *SequentialCollaborativeOrchestrator) executeAgentWithTimeout(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    // Get agent from runner
    agent, exists := sco.runner.GetAgent(event.AgentName)
    if !exists {
        return core.AgentResult{}, fmt.Errorf("agent not found: %s", event.AgentName)
    }
    
    // Execute agent
    return agent.Run(ctx, event, state)
}
```#
## 2. Route + Loop Pattern

```go
// RouteLoopOrchestrator routes to different agents, some operating in loops
type RouteLoopOrchestrator struct {
    routingRules map[string]string  // condition -> agent
    loopAgents   map[string]LoopConfig
    runner       *core.Runner
}

type LoopConfig struct {
    MaxIterations int
    Timeout       time.Duration
    BreakCondition func(core.State) bool
}

func NewRouteLoopOrchestrator(runner *core.Runner) *RouteLoopOrchestrator {
    return &RouteLoopOrchestrator{
        routingRules: make(map[string]string),
        loopAgents:   make(map[string]LoopConfig),
        runner:       runner,
    }
}

func (rlo *RouteLoopOrchestrator) AddRoutingRule(condition, agentName string) {
    rlo.routingRules[condition] = agentName
}

func (rlo *RouteLoopOrchestrator) AddLoopAgent(agentName string, config LoopConfig) {
    rlo.loopAgents[agentName] = config
}

func (rlo *RouteLoopOrchestrator) Execute(ctx context.Context, event core.Event, state core.State) (core.State, error) {
    // Phase 1: Route to appropriate agent
    targetAgent, err := rlo.routeToAgent(ctx, event, state)
    if err != nil {
        return state, fmt.Errorf("routing failed: %w", err)
    }
    
    // Phase 2: Execute agent (with loop if configured)
    if loopConfig, isLoopAgent := rlo.loopAgents[targetAgent]; isLoopAgent {
        return rlo.executeWithLoop(ctx, event, state, targetAgent, loopConfig)
    } else {
        return rlo.executeSingle(ctx, event, state, targetAgent)
    }
}

func (rlo *RouteLoopOrchestrator) routeToAgent(ctx context.Context, event core.Event, state core.State) (string, error) {
    // Simple routing based on state content
    message, ok := state.Get("message")
    if !ok {
        return "", fmt.Errorf("no message in state for routing")
    }
    
    messageStr := fmt.Sprintf("%v", message)
    
    // Check routing rules
    for condition, agentName := range rlo.routingRules {
        if strings.Contains(strings.ToLower(messageStr), strings.ToLower(condition)) {
            fmt.Printf("Routed to %s based on condition: %s\n", agentName, condition)
            return agentName, nil
        }
    }
    
    return "", fmt.Errorf("no routing rule matched for message: %s", messageStr)
}

func (rlo *RouteLoopOrchestrator) executeWithLoop(ctx context.Context, event core.Event, state core.State, agentName string, config LoopConfig) (core.State, error) {
    fmt.Printf("Executing %s with loop (max %d iterations)\n", agentName, config.MaxIterations)
    
    currentState := state.Clone()
    iteration := 0
    
    for iteration < config.MaxIterations {
        iteration++
        fmt.Printf("  Loop iteration %d for %s\n", iteration, agentName)
        
        // Create event for this iteration
        loopEvent := core.NewEvent(
            agentName,
            event.Data,
            map[string]string{
                "iteration":  fmt.Sprintf("%d", iteration),
                "session_id": event.GetSessionID(),
            },
        )
        
        // Execute agent with timeout
        ctx, cancel := context.WithTimeout(ctx, config.Timeout)
        result, err := rlo.executeAgentWithTimeout(ctx, loopEvent, currentState)
        cancel()
        
        if err != nil {
            return currentState, fmt.Errorf("loop iteration %d failed: %w", iteration, err)
        }
        
        // Update state
        currentState = result.OutputState
        currentState.Set("loop_iteration", iteration)
        
        // Check break condition
        if config.BreakCondition != nil && config.BreakCondition(currentState) {
            fmt.Printf("  Loop break condition met at iteration %d\n", iteration)
            break
        }
    }
    
    currentState.Set("loop_completed", true)
    currentState.Set("total_iterations", iteration)
    
    return currentState, nil
}

func (rlo *RouteLoopOrchestrator) executeSingle(ctx context.Context, event core.Event, state core.State, agentName string) (core.State, error) {
    fmt.Printf("Executing %s (single execution)\n", agentName)
    
    singleEvent := core.NewEvent(
        agentName,
        event.Data,
        map[string]string{
            "execution_type": "single",
            "session_id":     event.GetSessionID(),
        },
    )
    
    result, err := rlo.executeAgentWithTimeout(ctx, singleEvent, state)
    if err != nil {
        return state, err
    }
    
    return result.OutputState, nil
}

func (rlo *RouteLoopOrchestrator) executeAgentWithTimeout(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
    agent, exists := rlo.runner.GetAgent(event.AgentName)
    if !exists {
        return core.AgentResult{}, fmt.Errorf("agent not found: %s", event.AgentName)
    }
    
    return agent.Run(ctx, event, state)
}
```

### 3. Complex Multi-Stage Workflow

```go
// MultiStageOrchestrator implements a complex workflow with multiple orchestration patterns
type MultiStageOrchestrator struct {
    stages []WorkflowStage
    runner *core.Runner
}

type WorkflowStage struct {
    Name           string
    Pattern        OrchestrationPattern
    Agents         []string
    Timeout        time.Duration
    Prerequisites  []string // Required state keys from previous stages
    SuccessCondition func(core.State) bool
}

type OrchestrationPattern string

const (
    PatternRoute        OrchestrationPattern = "route"
    PatternSequential   OrchestrationPattern = "sequential"
    PatternCollaborative OrchestrationPattern = "collaborative"
    PatternLoop         OrchestrationPattern = "loop"
)

func NewMultiStageOrchestrator(runner *core.Runner) *MultiStageOrchestrator {
    return &MultiStageOrchestrator{
        stages: make([]WorkflowStage, 0),
        runner: runner,
    }
}

func (mso *MultiStageOrchestrator) AddStage(stage WorkflowStage) {
    mso.stages = append(mso.stages, stage)
}

func (mso *MultiStageOrchestrator) Execute(ctx context.Context, event core.Event, state core.State) (core.State, error) {
    currentState := state.Clone()
    
    fmt.Printf("Starting multi-stage workflow with %d stages\n", len(mso.stages))
    
    for i, stage := range mso.stages {
        fmt.Printf("Stage %d: %s (%s pattern)\n", i+1, stage.Name, stage.Pattern)
        
        // Check prerequisites
        if err := mso.checkPrerequisites(stage, currentState); err != nil {
            return currentState, fmt.Errorf("stage %s prerequisites not met: %w", stage.Name, err)
        }
        
        // Execute stage
        stageResult, err := mso.executeStage(ctx, event, currentState, stage)
        if err != nil {
            return currentState, fmt.Errorf("stage %s failed: %w", stage.Name, err)
        }
        
        // Update state
        currentState = stageResult
        currentState.Set(fmt.Sprintf("stage_%d_complete", i+1), true)
        currentState.Set(fmt.Sprintf("stage_%d_name", i+1), stage.Name)
        
        // Check success condition
        if stage.SuccessCondition != nil && !stage.SuccessCondition(currentState) {
            return currentState, fmt.Errorf("stage %s success condition not met", stage.Name)
        }
        
        fmt.Printf("Stage %d completed successfully\n", i+1)
    }
    
    currentState.Set("workflow_complete", true)
    currentState.Set("total_stages", len(mso.stages))
    
    return currentState, nil
}

func (mso *MultiStageOrchestrator) checkPrerequisites(stage WorkflowStage, state core.State) error {
    for _, prereq := range stage.Prerequisites {
        if !state.Has(prereq) {
            return fmt.Errorf("missing prerequisite: %s", prereq)
        }
    }
    return nil
}

func (mso *MultiStageOrchestrator) executeStage(ctx context.Context, event core.Event, state core.State, stage WorkflowStage) (core.State, error) {
    // Create stage-specific context with timeout
    stageCtx, cancel := context.WithTimeout(ctx, stage.Timeout)
    defer cancel()
    
    // Create stage event
    stageEvent := core.NewEvent(
        "stage",
        event.Data,
        map[string]string{
            "stage_name": stage.Name,
            "pattern":    string(stage.Pattern),
            "session_id": event.GetSessionID(),
        },
    )
    
    switch stage.Pattern {
    case PatternRoute:
        return mso.executeRouteStage(stageCtx, stageEvent, state, stage)
    case PatternSequential:
        return mso.executeSequentialStage(stageCtx, stageEvent, state, stage)
    case PatternCollaborative:
        return mso.executeCollaborativeStage(stageCtx, stageEvent, state, stage)
    case PatternLoop:
        return mso.executeLoopStage(stageCtx, stageEvent, state, stage)
    default:
        return state, fmt.Errorf("unknown orchestration pattern: %s", stage.Pattern)
    }
}

func (mso *MultiStageOrchestrator) executeRouteStage(ctx context.Context, event core.Event, state core.State, stage WorkflowStage) (core.State, error) {
    // Simple routing to first agent in the list
    if len(stage.Agents) == 0 {
        return state, fmt.Errorf("no agents specified for route stage")
    }
    
    agentName := stage.Agents[0]
    routeEvent := core.NewEvent(agentName, event.Data, event.Metadata)
    
    agent, exists := mso.runner.GetAgent(agentName)
    if !exists {
        return state, fmt.Errorf("agent not found: %s", agentName)
    }
    
    result, err := agent.Run(ctx, routeEvent, state)
    if err != nil {
        return state, err
    }
    
    return result.OutputState, nil
}

func (mso *MultiStageOrchestrator) executeSequentialStage(ctx context.Context, event core.Event, state core.State, stage WorkflowStage) (core.State, error) {
    currentState := state.Clone()
    
    for i, agentName := range stage.Agents {
        seqEvent := core.NewEvent(
            agentName,
            event.Data,
            map[string]string{
                "stage_step": fmt.Sprintf("%d", i+1),
                "session_id": event.GetSessionID(),
            },
        )
        
        agent, exists := mso.runner.GetAgent(agentName)
        if !exists {
            return currentState, fmt.Errorf("agent not found: %s", agentName)
        }
        
        result, err := agent.Run(ctx, seqEvent, currentState)
        if err != nil {
            return currentState, fmt.Errorf("sequential step %d (%s) failed: %w", i+1, agentName, err)
        }
        
        currentState = result.OutputState
    }
    
    return currentState, nil
}

func (mso *MultiStageOrchestrator) executeCollaborativeStage(ctx context.Context, event core.Event, state core.State, stage WorkflowStage) (core.State, error) {
    // Create collaborative orchestrator for this stage
    collaborativeOrchestrator := core.NewCollaborativeOrchestrator(
        mso.runner.GetCallbackRegistry(),
        stage.Timeout,
    )
    
    for _, agentName := range stage.Agents {
        collaborativeOrchestrator.AddAgent(agentName)
    }
    
    return collaborativeOrchestrator.Execute(ctx, event, state)
}

func (mso *MultiStageOrchestrator) executeLoopStage(ctx context.Context, event core.Event, state core.State, stage WorkflowStage) (core.State, error) {
    if len(stage.Agents) == 0 {
        return state, fmt.Errorf("no agents specified for loop stage")
    }
    
    // Use first agent for loop
    agentName := stage.Agents[0]
    currentState := state.Clone()
    maxIterations := 5 // Default max iterations
    
    for iteration := 1; iteration <= maxIterations; iteration++ {
        loopEvent := core.NewEvent(
            agentName,
            event.Data,
            map[string]string{
                "iteration":  fmt.Sprintf("%d", iteration),
                "session_id": event.GetSessionID(),
            },
        )
        
        agent, exists := mso.runner.GetAgent(agentName)
        if !exists {
            return currentState, fmt.Errorf("agent not found: %s", agentName)
        }
        
        result, err := agent.Run(ctx, loopEvent, currentState)
        if err != nil {
            return currentState, fmt.Errorf("loop iteration %d failed: %w", iteration, err)
        }
        
        currentState = result.OutputState
        currentState.Set("loop_iteration", iteration)
        
        // Check if we should break (simple condition)
        if completed, ok := currentState.Get("loop_complete"); ok && completed.(bool) {
            break
        }
    }
    
    return currentState, nil
}
```## Pra
ctical Examples

### 1. Research and Analysis Workflow

```go
func createResearchWorkflow() *MultiStageOrchestrator {
    runner := core.NewRunner(100)
    orchestrator := NewMultiStageOrchestrator(runner)
    
    // Stage 1: Route to appropriate research agent
    orchestrator.AddStage(WorkflowStage{
        Name:    "Initial Research Routing",
        Pattern: PatternRoute,
        Agents:  []string{"research-router"},
        Timeout: 30 * time.Second,
        SuccessCondition: func(state core.State) bool {
            return state.Has("research_topic") && state.Has("research_scope")
        },
    })
    
    // Stage 2: Sequential data gathering
    orchestrator.AddStage(WorkflowStage{
        Name:    "Data Gathering",
        Pattern: PatternSequential,
        Agents:  []string{"web-searcher", "document-analyzer", "data-extractor"},
        Timeout: 120 * time.Second,
        Prerequisites: []string{"research_topic"},
        SuccessCondition: func(state core.State) bool {
            return state.Has("raw_data") && state.Has("sources")
        },
    })
    
    // Stage 3: Collaborative analysis
    orchestrator.AddStage(WorkflowStage{
        Name:    "Collaborative Analysis",
        Pattern: PatternCollaborative,
        Agents:  []string{"data-analyst", "trend-analyzer", "insight-generator"},
        Timeout: 90 * time.Second,
        Prerequisites: []string{"raw_data"},
        SuccessCondition: func(state core.State) bool {
            return state.Has("analysis_results") && state.Has("insights")
        },
    })
    
    // Stage 4: Iterative refinement
    orchestrator.AddStage(WorkflowStage{
        Name:    "Report Refinement",
        Pattern: PatternLoop,
        Agents:  []string{"report-writer"},
        Timeout: 60 * time.Second,
        Prerequisites: []string{"analysis_results", "insights"},
        SuccessCondition: func(state core.State) bool {
            return state.Has("final_report") && state.Has("quality_score")
        },
    })
    
    return orchestrator
}

// Usage example
func runResearchWorkflow() {
    orchestrator := createResearchWorkflow()
    
    ctx := context.Background()
    event := core.NewEvent(
        "workflow",
        core.EventData{
            "message": "Research the impact of AI on software development",
            "depth":   "comprehensive",
        },
        map[string]string{"session_id": "research-001"},
    )
    
    initialState := core.NewState()
    initialState.Set("user_request", "Research the impact of AI on software development")
    
    finalState, err := orchestrator.Execute(ctx, event, initialState)
    if err != nil {
        log.Printf("Workflow failed: %v", err)
        return
    }
    
    // Process results
    if report, ok := finalState.Get("final_report"); ok {
        fmt.Printf("Research completed successfully:\n%s\n", report)
    }
}
```

### 2. Content Creation Pipeline

```go
func createContentPipeline() *SequentialCollaborativeOrchestrator {
    runner := core.NewRunner(100)
    
    // Sequential agents for content preparation
    sequentialAgents := []string{
        "topic-researcher",
        "outline-creator", 
        "fact-checker",
    }
    
    // Collaborative agents for content creation
    collaborativeAgents := []string{
        "content-writer",
        "editor",
        "seo-optimizer",
    }
    
    return NewSequentialCollaborativeOrchestrator(
        sequentialAgents,
        collaborativeAgents,
        runner,
    )
}

// Usage example
func runContentPipeline() {
    pipeline := createContentPipeline()
    
    ctx := context.Background()
    event := core.NewEvent(
        "content-pipeline",
        core.EventData{
            "topic": "Best practices for microservices architecture",
            "target_audience": "software developers",
            "content_type": "blog post",
        },
        map[string]string{"session_id": "content-001"},
    )
    
    initialState := core.NewState()
    initialState.Set("content_brief", "Write a comprehensive blog post about microservices best practices")
    
    finalState, err := pipeline.Execute(ctx, event, initialState)
    if err != nil {
        log.Printf("Content pipeline failed: %v", err)
        return
    }
    
    // Extract results
    if content, ok := finalState.Get("final_content"); ok {
        fmt.Printf("Content created successfully:\n%s\n", content)
    }
}
```

### 3. Customer Support Workflow

```go
func createSupportWorkflow() *RouteLoopOrchestrator {
    runner := core.NewRunner(100)
    orchestrator := NewRouteLoopOrchestrator(runner)
    
    // Add routing rules
    orchestrator.AddRoutingRule("technical", "tech-support")
    orchestrator.AddRoutingRule("billing", "billing-support")
    orchestrator.AddRoutingRule("general", "general-support")
    
    // Configure loop agents for complex issues
    orchestrator.AddLoopAgent("tech-support", LoopConfig{
        MaxIterations: 3,
        Timeout:       45 * time.Second,
        BreakCondition: func(state core.State) bool {
            if resolved, ok := state.Get("issue_resolved"); ok {
                return resolved.(bool)
            }
            return false
        },
    })
    
    orchestrator.AddLoopAgent("billing-support", LoopConfig{
        MaxIterations: 2,
        Timeout:       30 * time.Second,
        BreakCondition: func(state core.State) bool {
            if escalated, ok := state.Get("escalate_to_human"); ok {
                return escalated.(bool)
            }
            return false
        },
    })
    
    return orchestrator
}

// Usage example
func runSupportWorkflow() {
    workflow := createSupportWorkflow()
    
    ctx := context.Background()
    event := core.NewEvent(
        "support",
        core.EventData{
            "message": "I'm having technical issues with the API integration",
            "priority": "high",
        },
        map[string]string{"session_id": "support-001"},
    )
    
    initialState := core.NewState()
    initialState.Set("customer_issue", "API integration problems")
    initialState.Set("customer_tier", "premium")
    
    finalState, err := workflow.Execute(ctx, event, initialState)
    if err != nil {
        log.Printf("Support workflow failed: %v", err)
        return
    }
    
    // Process support results
    if resolution, ok := finalState.Get("resolution"); ok {
        fmt.Printf("Support case resolved: %s\n", resolution)
    }
}
```

## Advanced Mixed-Mode Patterns

### 1. Conditional Branching

```go
type ConditionalOrchestrator struct {
    conditions map[string]func(core.State) bool
    branches   map[string]Orchestrator
    fallback   Orchestrator
}

type Orchestrator interface {
    Execute(ctx context.Context, event core.Event, state core.State) (core.State, error)
}

func (co *ConditionalOrchestrator) Execute(ctx context.Context, event core.Event, state core.State) (core.State, error) {
    // Evaluate conditions and choose branch
    for conditionName, condition := range co.conditions {
        if condition(state) {
            fmt.Printf("Condition '%s' met, executing branch\n", conditionName)
            if branch, exists := co.branches[conditionName]; exists {
                return branch.Execute(ctx, event, state)
            }
        }
    }
    
    // Execute fallback if no conditions met
    fmt.Println("No conditions met, executing fallback")
    return co.fallback.Execute(ctx, event, state)
}
```

### 2. Parallel Mixed Execution

```go
type ParallelMixedOrchestrator struct {
    parallelBranches []ParallelBranch
    merger          ResultMerger
    timeout         time.Duration
}

type ParallelBranch struct {
    Name         string
    Orchestrator Orchestrator
    Required     bool // If true, failure fails entire execution
}

type ResultMerger interface {
    MergeResults(results map[string]core.State) (core.State, error)
}

func (pmo *ParallelMixedOrchestrator) Execute(ctx context.Context, event core.Event, state core.State) (core.State, error) {
    // Execute branches in parallel
    results := make(map[string]core.State)
    errors := make(map[string]error)
    
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for _, branch := range pmo.parallelBranches {
        wg.Add(1)
        go func(b ParallelBranch) {
            defer wg.Done()
            
            branchCtx, cancel := context.WithTimeout(ctx, pmo.timeout)
            defer cancel()
            
            result, err := b.Orchestrator.Execute(branchCtx, event, state)
            
            mu.Lock()
            if err != nil {
                errors[b.Name] = err
            } else {
                results[b.Name] = result
            }
            mu.Unlock()
        }(branch)
    }
    
    wg.Wait()
    
    // Check for required branch failures
    for _, branch := range pmo.parallelBranches {
        if branch.Required {
            if err, hasError := errors[branch.Name]; hasError {
                return state, fmt.Errorf("required branch %s failed: %w", branch.Name, err)
            }
        }
    }
    
    // Merge results
    return pmo.merger.MergeResults(results)
}
```

## Best Practices

### 1. Design Principles

- **Clear Stage Boundaries**: Define clear inputs and outputs for each stage
- **Failure Handling**: Implement proper error handling and recovery mechanisms
- **State Management**: Maintain clean state transitions between stages
- **Timeout Management**: Set appropriate timeouts for each orchestration pattern
- **Monitoring**: Add comprehensive logging and monitoring

### 2. Performance Optimization

- **Parallel Execution**: Use parallel patterns where possible
- **Resource Management**: Monitor and manage resource usage
- **Caching**: Cache intermediate results when appropriate
- **Load Balancing**: Distribute work across available agents

### 3. Testing Strategies

- **Unit Testing**: Test each orchestration pattern independently
- **Integration Testing**: Test complete workflows end-to-end
- **Failure Testing**: Test failure scenarios and recovery mechanisms
- **Performance Testing**: Test under various load conditions

## Common Pitfalls

### 1. State Pollution
- **Problem**: State becomes cluttered with intermediate results
- **Solution**: Use namespaced keys and clean up unnecessary data

### 2. Timeout Cascading
- **Problem**: Short timeouts cause cascading failures
- **Solution**: Set realistic timeouts based on actual execution times

### 3. Complex Dependencies
- **Problem**: Complex prerequisite chains become hard to manage
- **Solution**: Keep dependencies simple and well-documented

### 4. Error Propagation
- **Problem**: Errors from one stage affect unrelated stages
- **Solution**: Implement proper error isolation and handling

## Conclusion

Mixed-mode orchestration enables you to build sophisticated agent workflows that leverage the strengths of different orchestration patterns. By combining sequential, collaborative, routing, and loop patterns, you can create powerful systems that handle complex real-world scenarios.

Key takeaways:
- Choose the right orchestration pattern for each stage of your workflow
- Implement proper error handling and recovery mechanisms
- Monitor and optimize performance across all stages
- Test thoroughly with realistic scenarios
- Keep state management clean and well-organized

## Next Steps

- [Orchestration Patterns](orchestration-patterns.md) - Learn common workflow patterns
- [Error Handling](../core-concepts/error-handling.md) - Implement robust error handling
- [State Management](../core-concepts/state-management.md) - Master state flow patterns

## Further Reading

- [API Reference: Orchestration](../../api/core.md#orchestration)
- [Examples: Complex Workflows](../../examples/)
- [Production Deployment](../deployment/README.md)