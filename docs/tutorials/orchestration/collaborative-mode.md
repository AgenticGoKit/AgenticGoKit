# Collaborative Orchestration Mode

## Overview

Collaborative orchestration enables multiple agents to process the same input simultaneously, combining their results to produce richer, more comprehensive outputs. This pattern is perfect for tasks that benefit from multiple perspectives, redundancy, or ensemble approaches.

Think of collaborative orchestration like a research team where multiple experts examine the same problem from different angles and contribute their unique insights to create a comprehensive solution.

## Prerequisites

- Understanding of [Message Passing and Event Flow](../core-concepts/message-passing.md)
- Familiarity with [Route Orchestration](routing-mode.md)
- Basic knowledge of Go concurrency patterns
- Understanding of the [Orchestration Overview](README.md)

## How Collaborative Orchestration Works

### Basic Flow

```
                ┌─────────┐
                │ Agent A │
                └─────────┘
                     ▲
                     │
┌─────────┐     ┌──────────┐
│ Client  │────▶│ Collab.  │
└─────────┘     │ Orchestr.│
                └──────────┘
                     │
                     ▼
                ┌─────────┐
                │ Agent B │
                └─────────┘
```

1. **Event Broadcasting**: The same event is sent to all registered agents
2. **Parallel Processing**: Agents process the event simultaneously
3. **Result Collection**: Results are collected as they complete
4. **Result Combination**: Results are merged or synthesized into a final output

### Concurrency Model

Collaborative orchestration uses Go's concurrency features to manage parallel execution:

```go
// Simplified collaborative processing
func (o *CollaborativeOrchestrator) Dispatch(ctx context.Context, event Event) (AgentResult, error) {
    var wg sync.WaitGroup
    results := make([]AgentResult, 0)
    resultsMutex := &sync.Mutex{}
    
    // Launch all agents concurrently
    for name, agent := range o.handlers {
        wg.Add(1)
        go func(agentName string, handler AgentHandler) {
            defer wg.Done()
            
            result, err := handler.Run(ctx, event, core.NewState())
            if err == nil {
                resultsMutex.Lock()
                results = append(results, result)
                resultsMutex.Unlock()
            }
        }(name, agent)
    }
    
    // Wait for completion or timeout
    wg.Wait()
    
    // Combine results
    return o.combineResults(results), nil
}
```

## When to Use Collaborative Orchestration

Collaborative orchestration is ideal for:

- **Multiple Perspectives**: Tasks that benefit from different viewpoints
- **Redundancy**: Critical tasks where you want backup processing
- **Ensemble Methods**: Combining multiple approaches for better results
- **Parallel Research**: Information gathering from multiple sources
- **Quality Improvement**: Using multiple agents to validate or enhance results
- **Brainstorming**: Creative tasks that benefit from diverse inputs

### Use Case Examples

1. **Research Team**: Multiple agents search different sources simultaneously
2. **Code Review**: Multiple agents analyze code from different perspectives
3. **Content Creation**: Multiple writers create variations for A/B testing
4. **Decision Making**: Multiple agents provide recommendations for complex decisions
5. **Quality Assurance**: Multiple agents validate results independently

## Implementation Examples

### Basic Collaborative Research System

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "sync"
    "time"
    
    "github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
    // Create specialized research agents
    webSearcher, err := createWebSearchAgent()
    if err != nil {
        log.Fatal(err)
    }
    
    academicSearcher, err := createAcademicSearchAgent()
    if err != nil {
        log.Fatal(err)
    }
    
    newsAnalyzer, err := createNewsAnalyzerAgent()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create collaborative orchestrator
    runner := core.NewRunner(100)
    orchestrator := createCollaborativeOrchestrator(runner.GetCallbackRegistry(), 60*time.Second)
    runner.SetOrchestrator(orchestrator)
    
    // Register all research agents
    runner.RegisterAgent("web-searcher", webSearcher)
    runner.RegisterAgent("academic-searcher", academicSearcher)
    runner.RegisterAgent("news-analyzer", newsAnalyzer)
    
    // Set up result collection
    results := make([]core.AgentResult, 0)
    var resultsMutex sync.Mutex
    
    runner.RegisterCallback(core.HookAfterAgentRun, "result-collector",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            if args.Error == nil {
                resultsMutex.Lock()
                results = append(results, args.AgentResult)
                resultsMutex.Unlock()
                
                fmt.Printf("Collected result from %s\\n", args.AgentID)
            } else {
                fmt.Printf("Agent %s failed: %v\\n", args.AgentID, args.Error)
            }
            return args.State, nil
        },
    )
    
    // Start the runner
    ctx := context.Background()
    runner.Start(ctx)
    defer runner.Stop()
    
    // Create research event (no specific target - broadcasts to all)
    researchEvent := core.NewEvent(
        "",  // Empty target means broadcast to all agents
        core.EventData{
            "query": "Latest developments in quantum computing",
            "depth": "comprehensive",
            "timeframe": "last 6 months",
        },
        map[string]string{
            "session_id": "research-session-123",
            "priority": "high",
        },
    )
    
    // Emit the event
    runner.Emit(researchEvent)
    
    // Wait for all results (in production, use proper synchronization)
    time.Sleep(30 * time.Second)
    
    // Process and combine results
    fmt.Printf("\\n=== Research Results ===\\n")
    fmt.Printf("Collected %d results from research agents\\n\\n", len(results))
    
    combinedInsights := combineResearchResults(results)
    fmt.Printf("Combined Research Report:\\n%s\\n", combinedInsights)
}

// Create collaborative orchestrator
func createCollaborativeOrchestrator(registry *core.CallbackRegistry, timeout time.Duration) core.Orchestrator {
    return &CollaborativeOrchestrator{
        handlers: make(map[string]core.AgentHandler),
        registry: registry,
        timeout:  timeout,
    }
}

// Collaborative orchestrator implementation
type CollaborativeOrchestrator struct {
    handlers map[string]core.AgentHandler
    registry *core.CallbackRegistry
    timeout  time.Duration
    mu       sync.RWMutex
}

func (o *CollaborativeOrchestrator) RegisterAgent(name string, handler core.AgentHandler) error {
    o.mu.Lock()
    defer o.mu.Unlock()
    o.handlers[name] = handler
    return nil
}

func (o *CollaborativeOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    o.mu.RLock()
    handlers := make(map[string]core.AgentHandler)
    for name, handler := range o.handlers {
        handlers[name] = handler
    }
    o.mu.RUnlock()
    
    // Create context with timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, o.timeout)
    defer cancel()
    
    var wg sync.WaitGroup
    results := make([]core.AgentResult, 0)
    errors := make([]error, 0)
    resultsMutex := &sync.Mutex{}
    
    // Launch all agents concurrently
    for name, handler := range handlers {
        wg.Add(1)
        go func(agentName string, agentHandler core.AgentHandler) {
            defer wg.Done()
            
            // Run agent with timeout context
            result, err := agentHandler.Run(timeoutCtx, event, core.NewState())
            
            resultsMutex.Lock()
            defer resultsMutex.Unlock()
            
            if err != nil {
                errors = append(errors, fmt.Errorf("agent %s failed: %w", agentName, err))
                
                // Trigger error callback
                if o.registry != nil {
                    callbackArgs := core.CallbackArgs{
                        AgentID: agentName,
                        Event:   event,
                        Error:   err,
                        State:   core.NewState(),
                    }
                    o.registry.ExecuteCallbacks(ctx, core.HookAgentError, callbackArgs)
                }
            } else {
                results = append(results, result)
                
                // Trigger success callback
                if o.registry != nil {
                    callbackArgs := core.CallbackArgs{
                        AgentID:     agentName,
                        Event:       event,
                        AgentResult: result,
                        State:       result.OutputState,
                    }
                    o.registry.ExecuteCallbacks(ctx, core.HookAfterAgentRun, callbackArgs)
                }
            }
        }(name, handler)
    }
    
    // Wait for all agents to complete
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        // All agents completed
    case <-timeoutCtx.Done():
        return core.AgentResult{}, fmt.Errorf("collaborative orchestration timeout after %v", o.timeout)
    }
    
    // Check if we have any results
    if len(results) == 0 {
        return core.AgentResult{}, fmt.Errorf("all agents failed: %v", errors)
    }
    
    // Combine results
    combinedResult := o.combineResults(results)
    
    // Add metadata about the collaboration
    combinedResult.OutputState.SetMeta("collaboration_stats", map[string]interface{}{
        "successful_agents": len(results),
        "failed_agents":     len(errors),
        "total_agents":      len(handlers),
        "success_rate":      float64(len(results)) / float64(len(handlers)),
    })
    
    return combinedResult, nil
}

func (o *CollaborativeOrchestrator) combineResults(results []core.AgentResult) core.AgentResult {
    if len(results) == 0 {
        return core.AgentResult{OutputState: core.NewState()}
    }
    
    if len(results) == 1 {
        return results[0]
    }
    
    // Combine all results into a comprehensive output
    combinedState := core.NewState()
    allResponses := make([]string, 0)
    allSources := make([]string, 0)
    
    for i, result := range results {
        // Collect responses
        if response, ok := result.OutputState.Get("response"); ok {
            allResponses = append(allResponses, response.(string))
        }
        
        // Collect sources
        if sources, ok := result.OutputState.Get("sources"); ok {
            if sourceList, ok := sources.([]string); ok {
                allSources = append(allSources, sourceList...)
            }
        }
        
        // Store individual results
        combinedState.Set(fmt.Sprintf("result_%d", i), result.OutputState.GetAll())
    }
    
    // Create combined response
    combinedResponse := synthesizeResponses(allResponses)
    combinedState.Set("response", combinedResponse)
    combinedState.Set("individual_responses", allResponses)
    combinedState.Set("sources", removeDuplicates(allSources))
    combinedState.Set("agent_count", len(results))
    
    return core.AgentResult{
        OutputState: combinedState,
    }
}

// Agent implementations
func createWebSearchAgent() (core.AgentHandler, error) {
    return core.NewLLMAgent("web-search-agent", core.LLMConfig{
        SystemPrompt: `You are a web search specialist. Search for recent information about the given query.
                      Focus on current developments, trends, and practical applications.
                      Provide sources and links where possible.`,
        Temperature: 0.4,
        MaxTokens: 500,
    }, core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
}

func createAcademicSearchAgent() (core.AgentHandler, error) {
    return core.NewLLMAgent("academic-search-agent", core.LLMConfig{
        SystemPrompt: `You are an academic research specialist. Focus on peer-reviewed research,
                      scientific papers, and academic perspectives on the given query.
                      Emphasize theoretical foundations and research methodologies.`,
        Temperature: 0.3,
        MaxTokens: 500,
    }, core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
}

func createNewsAnalyzerAgent() (core.AgentHandler, error) {
    return core.NewLLMAgent("news-analyzer-agent", core.LLMConfig{
        SystemPrompt: `You are a news analysis specialist. Focus on recent news, industry reports,
                      and market developments related to the given query.
                      Emphasize business impact and real-world applications.`,
        Temperature: 0.4,
        MaxTokens: 500,
    }, core.OpenAIProvider{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
}

// Helper functions
func combineResearchResults(results []core.AgentResult) string {
    if len(results) == 0 {
        return "No research results available."
    }
    
    var combined strings.Builder
    combined.WriteString("# Comprehensive Research Report\\n\\n")
    
    for i, result := range results {
        if response, ok := result.OutputState.Get("response"); ok {
            combined.WriteString(fmt.Sprintf("## Source %d\\n", i+1))
            combined.WriteString(response.(string))
            combined.WriteString("\\n\\n")
        }
    }
    
    combined.WriteString("## Summary\\n")
    combined.WriteString("This report combines insights from multiple specialized research agents ")
    combined.WriteString("to provide a comprehensive view of the topic.\\n")
    
    return combined.String()
}

func synthesizeResponses(responses []string) string {
    if len(responses) == 0 {
        return "No responses to synthesize."
    }
    
    if len(responses) == 1 {
        return responses[0]
    }
    
    // Simple synthesis - in production, you might use an LLM for this
    var synthesis strings.Builder
    synthesis.WriteString("Based on multiple agent perspectives:\\n\\n")
    
    for i, response := range responses {
        synthesis.WriteString(fmt.Sprintf("**Perspective %d:** %s\\n\\n", i+1, response))
    }
    
    return synthesis.String()
}

func removeDuplicates(slice []string) []string {
    keys := make(map[string]bool)
    result := make([]string, 0)
    
    for _, item := range slice {
        if !keys[item] {
            keys[item] = true
            result = append(result, item)
        }
    }
    
    return result
}
```

### Advanced Collaborative Patterns

#### 1. Weighted Collaboration

Give different agents different weights based on their expertise:

```go
type WeightedCollaborativeOrchestrator struct {
    *CollaborativeOrchestrator
    weights map[string]float64
}

func (o *WeightedCollaborativeOrchestrator) combineResults(results []core.AgentResult) core.AgentResult {
    if len(results) == 0 {
        return core.AgentResult{OutputState: core.NewState()}
    }
    
    combinedState := core.NewState()
    weightedResponses := make(map[string]float64)
    totalWeight := 0.0
    
    for i, result := range results {
        agentID := fmt.Sprintf("agent_%d", i) // In practice, get from result metadata
        weight := o.weights[agentID]
        if weight == 0 {
            weight = 1.0 // Default weight
        }
        
        if response, ok := result.OutputState.Get("response"); ok {
            weightedResponses[response.(string)] = weight
            totalWeight += weight
        }
    }
    
    // Create weighted synthesis
    synthesis := o.createWeightedSynthesis(weightedResponses, totalWeight)
    combinedState.Set("response", synthesis)
    combinedState.Set("total_weight", totalWeight)
    
    return core.AgentResult{OutputState: combinedState}
}
```

#### 2. Consensus-Based Collaboration

Require agreement between agents before accepting results:

```go
type ConsensusOrchestrator struct {
    *CollaborativeOrchestrator
    consensusThreshold float64 // 0.6 = 60% agreement required
}

func (o *ConsensusOrchestrator) combineResults(results []core.AgentResult) core.AgentResult {
    if len(results) == 0 {
        return core.AgentResult{OutputState: core.NewState()}
    }
    
    // Analyze responses for consensus
    responseGroups := o.groupSimilarResponses(results)
    
    combinedState := core.NewState()
    
    for groupKey, group := range responseGroups {
        consensusLevel := float64(len(group)) / float64(len(results))
        
        if consensusLevel >= o.consensusThreshold {
            // This response has consensus
            combinedState.Set("consensus_response", groupKey)
            combinedState.Set("consensus_level", consensusLevel)
            combinedState.Set("supporting_agents", len(group))
            break
        }
    }
    
    // If no consensus, provide all perspectives
    if _, hasConsensus := combinedState.Get("consensus_response"); !hasConsensus {
        allResponses := make([]string, 0)
        for _, result := range results {
            if response, ok := result.OutputState.Get("response"); ok {
                allResponses = append(allResponses, response.(string))
            }
        }
        combinedState.Set("all_perspectives", allResponses)
        combinedState.Set("consensus_achieved", false)
    } else {
        combinedState.Set("consensus_achieved", true)
    }
    
    return core.AgentResult{OutputState: combinedState}
}
```

#### 3. Competitive Collaboration

Agents compete and the best result is selected:

```go
type CompetitiveOrchestrator struct {
    *CollaborativeOrchestrator
    evaluator func([]core.AgentResult) core.AgentResult
}

func (o *CompetitiveOrchestrator) combineResults(results []core.AgentResult) core.AgentResult {
    if len(results) == 0 {
        return core.AgentResult{OutputState: core.NewState()}
    }
    
    if len(results) == 1 {
        return results[0]
    }
    
    // Use evaluator to select best result
    if o.evaluator != nil {
        return o.evaluator(results)
    }
    
    // Default: select result with highest confidence score
    bestResult := results[0]
    bestScore := 0.0
    
    for _, result := range results {
        if score, ok := result.OutputState.Get("confidence_score"); ok {
            if scoreFloat, ok := score.(float64); ok && scoreFloat > bestScore {
                bestScore = scoreFloat
                bestResult = result
            }
        }
    }
    
    // Add competition metadata
    bestResult.OutputState.SetMeta("competition_winner", true)
    bestResult.OutputState.SetMeta("competing_agents", len(results))
    bestResult.OutputState.SetMeta("winning_score", bestScore)
    
    return bestResult
}
```

## Configuration-Based Collaborative Orchestration

Configure collaborative orchestration through TOML:

```toml
# agentflow.toml
[orchestration]
mode = "collaborative"
timeout = "60s"
max_concurrency = 5

[orchestration.collaborative]
# List of agents that will collaborate
agents = ["web-searcher", "academic-searcher", "news-analyzer"]

# Collaboration strategy
strategy = "consensus"  # "weighted", "competitive", "synthesis"
consensus_threshold = 0.6
failure_threshold = 0.3  # Continue if 70% succeed

# Agent weights (for weighted strategy)
[orchestration.collaborative.weights]
web-searcher = 1.0
academic-searcher = 1.5  # Higher weight for academic sources
news-analyzer = 0.8

# Timeout per agent
agent_timeout = "30s"

# Result combination settings
[orchestration.collaborative.combination]
max_response_length = 1000
include_individual_results = true
synthesize_responses = true
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

## Error Handling and Fault Tolerance

### 1. Partial Failure Handling

```go
type FaultTolerantCollaborativeOrchestrator struct {
    *CollaborativeOrchestrator
    failureThreshold float64 // 0.5 = continue if at least 50% succeed
    retryFailedAgents bool
    maxRetries       int
}

func (o *FaultTolerantCollaborativeOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
    results, errors := o.dispatchToAll(ctx, event)
    
    successRate := float64(len(results)) / float64(len(results) + len(errors))
    
    if successRate >= o.failureThreshold {
        // Enough agents succeeded
        combinedResult := o.combineResults(results)
        
        // Add failure information to metadata
        combinedResult.OutputState.SetMeta("partial_failure", len(errors) > 0)
        combinedResult.OutputState.SetMeta("success_rate", successRate)
        combinedResult.OutputState.SetMeta("failed_agents", len(errors))
        
        return combinedResult, nil
    }
    
    // Too many failures
    if o.retryFailedAgents && o.maxRetries > 0 {
        return o.retryFailedAgents(ctx, event, errors)
    }
    
    return core.AgentResult{}, fmt.Errorf("collaboration failed: %d/%d agents failed", 
        len(errors), len(results)+len(errors))
}
```

### 2. Timeout Management

```go
func (o *CollaborativeOrchestrator) dispatchWithTimeouts(ctx context.Context, event core.Event) ([]core.AgentResult, []error) {
    var wg sync.WaitGroup
    results := make([]core.AgentResult, 0)
    errors := make([]error, 0)
    resultsMutex := &sync.Mutex{}
    
    for name, handler := range o.handlers {
        wg.Add(1)
        go func(agentName string, agentHandler core.AgentHandler) {
            defer wg.Done()
            
            // Create per-agent timeout
            agentCtx, cancel := context.WithTimeout(ctx, o.agentTimeout)
            defer cancel()
            
            result, err := agentHandler.Run(agentCtx, event, core.NewState())
            
            resultsMutex.Lock()
            defer resultsMutex.Unlock()
            
            if err != nil {
                if agentCtx.Err() == context.DeadlineExceeded {
                    errors = append(errors, fmt.Errorf("agent %s timed out", agentName))
                } else {
                    errors = append(errors, fmt.Errorf("agent %s failed: %w", agentName, err))
                }
            } else {
                results = append(results, result)
            }
        }(name, handler)
    }
    
    // Wait with overall timeout
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        // All completed
    case <-ctx.Done():
        // Overall timeout
        errors = append(errors, fmt.Errorf("overall collaboration timeout"))
    }
    
    return results, errors
}
```

### 3. Circuit Breaker Pattern

```go
type CircuitBreakerCollaborativeOrchestrator struct {
    *CollaborativeOrchestrator
    circuitBreakers map[string]*CircuitBreaker
}

type CircuitBreaker struct {
    failures    int
    maxFailures int
    resetTime   time.Time
    state       string // "closed", "open", "half-open"
    mu          sync.Mutex
}

func (cb *CircuitBreaker) CanExecute() bool {
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

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.failures = 0
    cb.state = "closed"
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.failures++
    if cb.failures >= cb.maxFailures {
        cb.state = "open"
        cb.resetTime = time.Now().Add(30 * time.Second)
    }
}
```

## Performance Optimization

### 1. Concurrency Control

```go
type ConcurrencyLimitedOrchestrator struct {
    *CollaborativeOrchestrator
    semaphore chan struct{}
}

func NewConcurrencyLimitedOrchestrator(maxConcurrency int) *ConcurrencyLimitedOrchestrator {
    return &ConcurrencyLimitedOrchestrator{
        CollaborativeOrchestrator: NewCollaborativeOrchestrator(),
        semaphore: make(chan struct{}, maxConcurrency),
    }
}

func (o *ConcurrencyLimitedOrchestrator) runAgent(ctx context.Context, handler core.AgentHandler, event core.Event) (core.AgentResult, error) {
    // Acquire semaphore
    select {
    case o.semaphore <- struct{}{}:
        defer func() { <-o.semaphore }()
    case <-ctx.Done():
        return core.AgentResult{}, ctx.Err()
    }
    
    return handler.Run(ctx, event, core.NewState())
}
```

### 2. Result Streaming

```go
type StreamingCollaborativeOrchestrator struct {
    *CollaborativeOrchestrator
    resultChannel chan core.AgentResult
}

func (o *StreamingCollaborativeOrchestrator) DispatchStreaming(ctx context.Context, event core.Event) <-chan core.AgentResult {
    resultChan := make(chan core.AgentResult, len(o.handlers))
    
    var wg sync.WaitGroup
    
    for name, handler := range o.handlers {
        wg.Add(1)
        go func(agentName string, agentHandler core.AgentHandler) {
            defer wg.Done()
            
            result, err := agentHandler.Run(ctx, event, core.NewState())
            if err == nil {
                select {
                case resultChan <- result:
                case <-ctx.Done():
                }
            }
        }(name, handler)
    }
    
    // Close channel when all agents complete
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    return resultChan
}
```

## Monitoring and Metrics

### 1. Collaboration Metrics

```go
type CollaborationMetrics struct {
    totalCollaborations int64
    successfulAgents    map[string]int64
    failedAgents        map[string]int64
    averageLatency      map[string]time.Duration
    mu                  sync.RWMutex
}

func (m *CollaborationMetrics) RecordAgentResult(agentID string, duration time.Duration, success bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if success {
        m.successfulAgents[agentID]++
    } else {
        m.failedAgents[agentID]++
    }
    
    // Update average latency (simplified)
    m.averageLatency[agentID] = duration
}

func (m *CollaborationMetrics) GetSuccessRate(agentID string) float64 {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    successful := m.successfulAgents[agentID]
    failed := m.failedAgents[agentID]
    total := successful + failed
    
    if total == 0 {
        return 0.0
    }
    
    return float64(successful) / float64(total)
}
```

### 2. Real-time Monitoring

```go
func setupCollaborationMonitoring(orchestrator *CollaborativeOrchestrator) {
    orchestrator.RegisterCallback(core.HookBeforeAgentRun, "collaboration-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            fmt.Printf("[%s] Starting collaboration with agent %s\\n", 
                time.Now().Format(time.RFC3339), args.AgentID)
            return args.State, nil
        },
    )
    
    orchestrator.RegisterCallback(core.HookAfterAgentRun, "collaboration-monitor",
        func(ctx context.Context, args core.CallbackArgs) (core.State, error) {
            duration := time.Since(args.StartTime)
            fmt.Printf("[%s] Agent %s completed in %v\\n", 
                time.Now().Format(time.RFC3339), args.AgentID, duration)
            return args.State, nil
        },
    )
}
```

## Best Practices

### 1. Agent Design for Collaboration

- **Complementary Expertise**: Design agents with different but complementary skills
- **Consistent Interfaces**: Ensure all agents return compatible result formats
- **Independent Processing**: Agents should not depend on each other's results
- **Timeout Awareness**: Design agents to handle timeout gracefully
- **Error Reporting**: Provide clear error messages for debugging

### 2. Result Combination Strategies

```go
// Strategy 1: Simple concatenation
func concatenateResults(results []core.AgentResult) string {
    var combined strings.Builder
    for i, result := range results {
        if response, ok := result.OutputState.Get("response"); ok {
            combined.WriteString(fmt.Sprintf("Agent %d: %s\\n", i+1, response))
        }
    }
    return combined.String()
}

// Strategy 2: Confidence-weighted combination
func weightedCombination(results []core.AgentResult) string {
    type weightedResponse struct {
        response   string
        confidence float64
    }
    
    responses := make([]weightedResponse, 0)
    for _, result := range results {
        if response, ok := result.OutputState.Get("response"); ok {
            confidence := 1.0 // Default confidence
            if conf, ok := result.OutputState.Get("confidence"); ok {
                confidence = conf.(float64)
            }
            responses = append(responses, weightedResponse{
                response:   response.(string),
                confidence: confidence,
            })
        }
    }
    
    // Sort by confidence and combine
    sort.Slice(responses, func(i, j int) bool {
        return responses[i].confidence > responses[j].confidence
    })
    
    var combined strings.Builder
    for _, resp := range responses {
        combined.WriteString(fmt.Sprintf("(Confidence: %.2f) %s\\n", resp.confidence, resp.response))
    }
    
    return combined.String()
}

// Strategy 3: LLM-based synthesis
func llmSynthesis(results []core.AgentResult, llmProvider core.LLMProvider) (string, error) {
    responses := make([]string, 0)
    for _, result := range results {
        if response, ok := result.OutputState.Get("response"); ok {
            responses = append(responses, response.(string))
        }
    }
    
    synthesisPrompt := fmt.Sprintf(`
        Synthesize the following responses into a comprehensive, coherent answer:
        
        %s
        
        Provide a unified response that incorporates the best insights from each source.
    `, strings.Join(responses, "\\n\\n---\\n\\n"))
    
    request := core.LLMRequest{
        Messages: []core.Message{
            {Role: "user", Content: synthesisPrompt},
        },
        Temperature: 0.3,
        MaxTokens:   800,
    }
    
    response, err := llmProvider.Call(context.Background(), request)
    if err != nil {
        return "", err
    }
    
    return response.Content, nil
}
```

### 3. Testing Collaborative Orchestration

```go
func TestCollaborativeOrchestration(t *testing.T) {
    // Create mock agents
    agent1 := &MockAgent{response: "Response from agent 1"}
    agent2 := &MockAgent{response: "Response from agent 2"}
    agent3 := &MockAgent{response: "Response from agent 3"}
    
    // Create orchestrator
    orchestrator := NewCollaborativeOrchestrator()
    orchestrator.RegisterAgent("agent1", agent1)
    orchestrator.RegisterAgent("agent2", agent2)
    orchestrator.RegisterAgent("agent3", agent3)
    
    // Create test event
    event := core.NewEvent(
        "",  // Broadcast to all
        core.EventData{"query": "test query"},
        map[string]string{"session_id": "test"},
    )
    
    // Test collaboration
    result, err := orchestrator.Dispatch(context.Background(), event)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    // Verify all agents were called
    assert.True(t, agent1.WasCalled())
    assert.True(t, agent2.WasCalled())
    assert.True(t, agent3.WasCalled())
    
    // Verify result combination
    response, ok := result.OutputState.Get("response")
    assert.True(t, ok)
    assert.Contains(t, response.(string), "agent 1")
    assert.Contains(t, response.(string), "agent 2")
    assert.Contains(t, response.(string), "agent 3")
}
```

## Common Pitfalls and Solutions

### 1. Resource Exhaustion

**Problem**: Too many agents running simultaneously can exhaust system resources.

**Solution**: Implement concurrency limits and resource monitoring:

```go
// Limit concurrent agents
orchestrator := NewConcurrencyLimitedOrchestrator(5) // Max 5 concurrent agents

// Monitor resource usage
go func() {
    for {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        if m.Alloc > maxMemoryThreshold {
            log.Printf("High memory usage: %d MB", m.Alloc/1024/1024)
        }
        time.Sleep(10 * time.Second)
    }
}()
```

### 2. Result Quality Degradation

**Problem**: Combining poor results with good ones can degrade overall quality.

**Solution**: Implement quality filtering and validation:

```go
func filterQualityResults(results []core.AgentResult, minQuality float64) []core.AgentResult {
    filtered := make([]core.AgentResult, 0)
    
    for _, result := range results {
        if quality, ok := result.OutputState.Get("quality_score"); ok {
            if qualityFloat, ok := quality.(float64); ok && qualityFloat >= minQuality {
                filtered = append(filtered, result)
            }
        } else {
            // Include results without quality scores
            filtered = append(filtered, result)
        }
    }
    
    return filtered
}
```

### 3. Inconsistent Result Formats

**Problem**: Different agents return results in incompatible formats.

**Solution**: Implement result normalization:

```go
func normalizeResults(results []core.AgentResult) []core.AgentResult {
    normalized := make([]core.AgentResult, 0)
    
    for _, result := range results {
        normalizedResult := normalizeResult(result)
        normalized = append(normalized, normalizedResult)
    }
    
    return normalized
}

func normalizeResult(result core.AgentResult) core.AgentResult {
    normalizedState := core.NewState()
    
    // Ensure standard fields exist
    if response, ok := result.OutputState.Get("response"); ok {
        normalizedState.Set("response", response)
    } else {
        normalizedState.Set("response", "No response provided")
    }
    
    if confidence, ok := result.OutputState.Get("confidence"); ok {
        normalizedState.Set("confidence", confidence)
    } else {
        normalizedState.Set("confidence", 0.5) // Default confidence
    }
    
    // Copy other fields
    for key, value := range result.OutputState.GetAll() {
        if key != "response" && key != "confidence" {
            normalizedState.Set(key, value)
        }
    }
    
    return core.AgentResult{OutputState: normalizedState}
}
```

## Conclusion

Collaborative orchestration enables powerful multi-agent systems that leverage the strengths of multiple agents working in parallel. By understanding the patterns, implementing proper error handling, and following best practices, you can build robust collaborative agent systems.

Key takeaways:
- Collaborative orchestration excels at tasks requiring multiple perspectives
- Proper timeout and error handling are crucial for reliability
- Result combination strategies significantly impact output quality
- Monitor performance and resource usage to prevent system overload
- Design agents with collaboration in mind from the start

## Next Steps

- [Sequential Orchestration](sequential-mode.md) - Learn pipeline processing patterns
- [Mixed Orchestration](mixed-mode.md) - Combine multiple orchestration patterns
- [State Management](../core-concepts/state-management.md) - Understand data flow in collaborative systems
- [Error Handling](../core-concepts/error-handling.md) - Master robust error management

## Further Reading

- [API Reference: CollaborativeOrchestrator](../../api/core.md#collaborative-orchestrator)
- [Examples: Collaborative Agents](../../examples/02-multi-agent-collab/)
- [Performance Guide: Scaling Collaborative Systems](../../production/scaling.md#collaborative-scaling)