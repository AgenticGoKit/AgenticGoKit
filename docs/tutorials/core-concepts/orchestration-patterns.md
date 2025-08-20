# Multi-Agent Orchestration Patterns

## Overview

AgenticGoKit provides powerful multi-agent orchestration patterns that enable you to build complex workflows with multiple agents working together. The system supports various orchestration modes and includes workflow visualization via the CLI using Mermaid diagrams.

## Prerequisites

- Understanding of [Agent Lifecycle](agent-lifecycle.md)
- Familiarity with [State Management](state-management.md)
- Basic knowledge of Go concurrency

## Learning Objectives

By the end of this tutorial, you'll understand:
- Different orchestration patterns and their use cases
- How to configure each orchestration mode using CLI and configuration files
- Performance characteristics of each pattern
- When to use which pattern
- How to visualize your orchestration workflows

## CLI Quick Start

The fastest way to create multi-agent workflows is using the AgentFlow CLI:

```bash
# Collaborative workflow - all agents process events in parallel
agentcli create research-system \
  --orchestration-mode collaborative \
  --agents 3 \
  --orchestration-timeout 60 \
  --mcp-enabled

# Sequential pipeline - agents process one after another
agentcli create data-pipeline \
  --orchestration-mode sequential \
  --sequential-agents "collector,processor,formatter" \
  --orchestration-timeout 45 \
  

# Loop-based workflow - single agent repeats with conditions
agentcli create quality-loop \
  --orchestration-mode loop \
  --loop-agent "quality-checker" \
  --max-iterations 5 \
  --orchestration-timeout 120

# Mixed orchestration - combine collaborative and sequential
agentcli create complex-workflow \
  --orchestration-mode mixed \
  --collaborative-agents "analyzer,validator" \
  --sequential-agents "processor,reporter" \
  --orchestration-timeout 90
```

All generated projects use **configuration-based orchestration** via `agentflow.toml`, making it easy to modify orchestration patterns without changing code.

## Configuration-Based Orchestration

AgentFlow supports configuration-driven orchestration through `agentflow.toml` files. This approach allows you to change orchestration patterns without modifying code.

### agentflow.toml Configuration

```toml
[orchestration]
mode = "sequential"                    # route, collaborative, sequential, loop, mixed
timeout_seconds = 30                   # Timeout for orchestration operations
max_iterations = 5                     # Maximum iterations for loop mode

# Sequential mode: agents process in order
sequential_agents = ["agent1", "agent2", "agent3"]

# Collaborative mode: agents process in parallel  
collaborative_agents = ["analyzer", "validator", "processor"]

# Loop mode: single agent repeats
loop_agent = "processor"

# Mixed mode: combine collaborative and sequential
# collaborative_agents = ["analyzer", "validator"]
# sequential_agents = ["processor", "reporter"]
```

### Using Configuration-Based Runners

Generated projects use configuration-based orchestration:

```go
// Create runner and orchestrator from agentflow.toml (route/collab/seq/loop/mixed)
runner, err := core.NewRunnerFromConfig("agentflow.toml")
if err != nil { log.Fatal(err) }

// Register your agents by name
_ = runner.RegisterAgent("agent1", agent1)
_ = runner.RegisterAgent("agent2", agent2)
_ = runner.RegisterAgent("agent3", agent3)

// Start/Emit/Stop lifecycle
ctx := context.Background()
_ = runner.Start(ctx)
defer runner.Stop()
_ = runner.Emit(core.NewEvent("agent1", core.EventData{"message": "hello"}, map[string]string{"route":"agent1"}))
```

### Benefits of Configuration-Based Approach

- **No Code Changes**: Switch orchestration modes by editing TOML files
- **Environment-Specific**: Different configs for dev/staging/production
- **Runtime Flexibility**: Change orchestration without rebuilding
- **Validation**: Built-in validation of orchestration parameters
- **Consistency**: Same configuration format across all projects

## Orchestration Patterns

### 1. Route Pattern (Default)

**Use Case**: Simple request-response scenarios where each event goes to a specific agent.

```go
// Route orchestration (default)
runner, _ := core.NewRunnerFromConfig("agentflow.toml")
_ = runner.RegisterAgent("research", researchAgent)
_ = runner.RegisterAgent("analysis", analysisAgent)

// Events are routed to specific agents based on routing metadata
evt := core.NewEvent("research", data, map[string]string{"route":"research"})
_ = runner.Start(context.Background())
defer runner.Stop()
_ = runner.Emit(evt)
```

**Characteristics**:
- Single agent processes each event
- Fast and efficient for simple scenarios
- No coordination overhead
- Easy to debug and understand

### 2. Collaborative Pattern

**Use Case**: When you need multiple perspectives or parallel processing of the same input.

Think of collaborative orchestration like a research team where multiple experts examine the same problem from different angles and contribute their unique insights to create a comprehensive solution.

```go
// Collaborative orchestration - all agents process the same event
runner, _ := core.NewRunnerFromConfig("agentflow.toml") // [orchestration].mode = "collaborative"
_ = runner.RegisterAgent("researcher", NewResearchAgent())
_ = runner.RegisterAgent("analyzer", NewAnalysisAgent())
_ = runner.RegisterAgent("validator", NewValidationAgent())

_ = runner.Start(context.Background())
defer runner.Stop()
_ = runner.Emit(core.NewEvent("researcher", data, map[string]string{"route":"researcher"}))
```

**Flow Diagram**:
```
                ┌─────────────┐
                │ Researcher  │
                └─────────────┘
                      ▲
                      │
┌─────────┐     ┌──────────────┐
│ Client  │────▶│ Collaborative│
└─────────┘     │ Orchestrator │
                └──────────────┘
                      │
                      ▼
                ┌─────────────┐
                │  Analyzer   │
                └─────────────┘
                      │
                      ▼
                ┌─────────────┐
                │  Validator  │
                └─────────────┘
```

**Characteristics**:
- All agents process the same event simultaneously
- Results are automatically aggregated
- Higher throughput for complex analysis
- Built-in fault tolerance
- Perfect for ensemble approaches and multiple perspectives

### 3. Sequential Pattern

**Use Case**: Data processing pipelines where each step builds on the previous.

Think of sequential orchestration like an assembly line where each worker (agent) performs a specific task and passes the work to the next worker in line, with each step adding value to the final product.

```go
// Sequential orchestration - agents process in order
runner, _ := core.NewRunnerFromConfig("agentflow.toml") // [orchestration].mode = "sequential"
_ = runner.RegisterAgent("collector", NewCollectorAgent())
_ = runner.RegisterAgent("processor", NewProcessorAgent())
_ = runner.RegisterAgent("formatter", NewFormatterAgent())
```

**Flow Diagram**:
```
┌─────────┐     ┌──────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│ Client  │────▶│ Sequence │────▶│Collector│────▶│Processor│────▶│Formatter│
└─────────┘     │Orchestr. │     └─────────┘     └─────────┘     └─────────┘
                └──────────┘                                           │
                     ▲                                                │
                     └────────────────────────────────────────────────┘
```

**State Flow**: The state object carries data through the entire pipeline, with each agent adding or transforming data for the next agent.

**Characteristics**:
- Agents process events in a specific order
- Output of one agent becomes input to the next
- Perfect for data transformation pipelines
- Easy to reason about data flow
- State propagates through the entire sequence

### 4. Loop Pattern

**Use Case**: Iterative refinement or quality improvement workflows.

```go
// Loop orchestration - single agent repeats until conditions are met
runner, _ := core.NewRunnerFromConfig("agentflow.toml") // [orchestration].mode = "loop"
_ = runner.RegisterAgent("quality-checker", NewQualityCheckerAgent())
```

**Characteristics**:
- Single agent processes the same event multiple times
- Continues until completion condition is met or max iterations reached
- Useful for iterative improvement
- Built-in loop detection and termination

### 5. Mixed Pattern

**Use Case**: Complex workflows that need both parallel and sequential processing.

```go
// Mixed orchestration - combines collaborative and sequential patterns
runner, _ := core.NewRunnerFromConfig("agentflow.toml") // [orchestration].mode = "mixed"
_ = runner.RegisterAgent("analyzer", NewAnalyzerAgent())
_ = runner.RegisterAgent("validator", NewValidatorAgent())
_ = runner.RegisterAgent("processor", NewProcessorAgent())
_ = runner.RegisterAgent("reporter", NewReporterAgent())
```

**Characteristics**:
- Collaborative agents run first in parallel
- Sequential agents run after collaborative phase completes
- Combines benefits of both patterns
- More complex but very powerful

## Configuration-Based Orchestration

You can configure orchestration patterns using `agentflow.toml`:

```toml
[llm]
provider = "ollama"
model = "gemma3:1b"

[orchestration]
mode = "collaborative"  # route, collaborative, sequential, loop, mixed
timeout_seconds = 30
max_iterations = 5      # For loop mode

# For sequential mode
sequential_agents = ["collector", "processor", "formatter"]

# For mixed mode
collaborative_agents = ["analyzer", "validator"]
sequential_agents = ["processor", "reporter"]

# For loop mode
loop_agent = "quality-checker"
```

Then load the configuration:

```go
runner, _ := core.NewRunnerFromConfig("agentflow.toml")
_ = runner.RegisterAgent("collector", collectorAgent)
_ = runner.RegisterAgent("processor", processorAgent)
// ... register other agents
_ = runner.Start(context.Background())
defer runner.Stop()
_ = runner.Emit(core.NewEvent("collector", data, nil))
```

## Pattern Selection Guide

| Pattern | Best For | Latency | Throughput | Complexity |
|---------|----------|---------|------------|------------|
| **Route** | Simple requests | Low | Medium | Low |
| **Collaborative** | Analysis, multiple perspectives | Medium | High | Medium |
| **Sequential** | Data pipelines | High | Low | Low |
| **Loop** | Iterative refinement | High | Low | Medium |
| **Mixed** | Complex workflows | High | Medium | High |

## Performance Considerations

### Collaborative Pattern
- **Pros**: High throughput, fault tolerance, multiple perspectives
- **Cons**: Higher resource usage, result aggregation complexity
- **Best for**: Analysis tasks, validation workflows

### Sequential Pattern
- **Pros**: Clear data flow, easy debugging, resource efficient
- **Cons**: Higher latency, single point of failure
- **Best for**: Data transformation, step-by-step processing

### Loop Pattern
- **Pros**: Iterative improvement, quality assurance
- **Cons**: Potentially long execution time, complexity in termination
- **Best for**: Quality checking, iterative refinement

## Error Handling in Orchestration

Different patterns handle errors differently:

```go
// Configure failure thresholds, retries, and timeouts in agentflow.toml.
// Build runner from config and use Start/Emit/Stop lifecycle.
runner, _ := core.NewRunnerFromConfig("agentflow.toml")
```

## Monitoring and Observability

All orchestration patterns support built-in monitoring:

```go
// Enable tracing for orchestration
runner.RegisterCallback(core.HookBeforeOrchestration, func(ctx context.Context, event core.Event) {
    core.Logger().Info().Str("event_id", event.GetID()).Msg("Starting orchestration")
})

runner.RegisterCallback(core.HookAfterOrchestration, func(ctx context.Context, result core.AgentResult) {
    core.Logger().Info().
        Dur("duration", result.Duration).
        Bool("success", result.Error == "").
        Msg("Orchestration completed")
})
```

## Best Practices

1. **Start Simple**: Begin with Route pattern and add complexity as needed
2. **Consider Latency**: Sequential patterns have higher latency than parallel ones
3. **Resource Management**: Collaborative patterns use more resources
4. **Error Handling**: Design appropriate error handling for your pattern
5. **Monitoring**: Always add monitoring for production orchestration
6. **Testing**: Test each pattern thoroughly with your specific use case

## Common Patterns

### Research and Analysis Workflow
```go
// Collaborative for research, sequential for reporting
runner, _ := core.NewRunnerFromConfig("agentflow.toml")
_ = runner.RegisterAgent("web-researcher", webResearchAgent)
_ = runner.RegisterAgent("doc-analyzer", docAnalysisAgent)
_ = runner.RegisterAgent("fact-checker", factCheckAgent)
_ = runner.RegisterAgent("synthesizer", synthesizer)
_ = runner.RegisterAgent("formatter", formatter)
```

### Data Processing Pipeline
```go
// Sequential processing with error handling
runner, _ := core.NewRunnerFromConfig("agentflow.toml")
_ = runner.RegisterAgent("validator", dataValidatorAgent)
_ = runner.RegisterAgent("transformer", dataTransformerAgent)
_ = runner.RegisterAgent("enricher", dataEnricherAgent)
_ = runner.RegisterAgent("publisher", dataPublisherAgent)
```

### Quality Assurance Loop
```go
// Loop until quality threshold is met
runner, _ := core.NewRunnerFromConfig("agentflow.toml")
_ = runner.RegisterAgent("quality-checker", qualityAgent)
```

## Workflow Visualization

Use the CLI to generate and preview Mermaid diagrams of your orchestration flows.

## API Reference

### CLI Configuration Options

All CLI flags for multi-agent orchestration:

```bash
# Orchestration mode flags
--orchestration-mode string          # collaborative, sequential, loop, mixed
--collaborative-agents string        # Comma-separated list of agents
--sequential-agents string           # Comma-separated list of agents
--loop-agent string                  # Single agent name for loop mode
--max-iterations int                 # Maximum loop iterations (default: 10)

# Configuration flags
--orchestration-timeout int          # Timeout in seconds (default: 60)
--failure-threshold float           # Failure threshold 0.0-1.0 (default: 0.5)
--max-concurrency int              # Maximum concurrent agents (default: 5)

# Visualization flags
--visualize                         # Generate Mermaid diagrams
--visualize-output string           # Custom output directory for diagrams
```

### Orchestration Modes

```go
type OrchestrationMode string

const (
    OrchestrationRoute       OrchestrationMode = "route"       // Route to single agent
    OrchestrationCollaborate OrchestrationMode = "collaborate" // Send to all agents
    OrchestrationSequential  OrchestrationMode = "sequential"  // Process in sequence
    OrchestrationLoop        OrchestrationMode = "loop"        // Loop single agent
    OrchestrationMixed       OrchestrationMode = "mixed"       // Combine patterns
)
```

### Orchestrator Interface

```go
type Orchestrator interface {
    Dispatch(ctx context.Context, event Event) (AgentResult, error)
    RegisterAgent(name string, handler AgentHandler) error
    GetCallbackRegistry() *CallbackRegistry
    Stop()
}
```

## Advanced Configuration Options

### OrchestrationConfig

```go
type OrchestrationConfig struct {
    Timeout          time.Duration  // Overall orchestration timeout
    MaxConcurrency   int           // Maximum concurrent agents
    FailureThreshold float64       // Failure threshold (0.0-1.0)
    RetryPolicy      *RetryPolicy  // Retry configuration
}
```

### RetryPolicy

```go
type RetryPolicy struct {
    MaxRetries      int           // Maximum retry attempts
    InitialDelay    time.Duration // Initial delay before first retry
    MaxDelay        time.Duration // Maximum delay between retries  
    BackoffFactor   float64       // Exponential backoff multiplier
    Jitter          bool          // Add random jitter to delays
    RetryableErrors []string      // List of retryable error codes
}
```

## Convenience Functions

Some convenience helpers may exist, but the recommended and supported public path is to construct runners from configuration using `core.NewRunnerFromConfig` and control behavior via the `[orchestration]` section.

## Migration from Internal APIs

If you were previously using internal orchestrator packages:

1. Replace `internal/orchestrator` imports with `core`
2. Use `core.NewCollaborativeOrchestrator()` instead of internal constructors
3. Update agent handlers to use public `core.AgentHandler` interface
4. Use public `core.Event` and `core.State` types

## Performance Considerations

- Collaborative orchestration runs agents concurrently for better performance
- Configure `MaxConcurrency` to limit resource usage
- Use timeouts to prevent resource leaks
- Monitor agent execution times and optimize slow agents
- Consider using `FailureThreshold` to fail fast when many agents are failing

## Next Steps

- **[Error Handling](error-handling.md)** - Learn advanced error handling patterns
- **[State Management](state-management.md)** - Understand state flow in orchestration
- **[Visualization Guide](../../guides/development/visualization.md)** - Learn workflow visualization
- **[Performance Optimization](../../guides/development/best-practices.md)** - Scale your orchestration

## Troubleshooting

**Common Issues:**

1. **Deadlocks in Mixed Mode**: Ensure proper state management between phases
2. **Memory Leaks in Loop Mode**: Implement proper termination conditions
3. **Timeout Issues**: Adjust timeout values based on your agent complexity
4. **State Corruption**: Use proper state isolation in collaborative mode
5. **Agents Not Registered**: Ensure all agents are registered before dispatching events
6. **Configuration Errors**: Validate your `agentflow.toml` configuration

For more help, see the [Troubleshooting Guide](../../guides/troubleshooting.md).