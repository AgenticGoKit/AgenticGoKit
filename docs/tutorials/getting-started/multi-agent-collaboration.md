# Multi-Agent Collaboration Tutorial (15 minutes)

## Overview

Learn how to orchestrate multiple agents working together using different patterns. You'll explore collaborative, sequential, and mixed orchestration modes to build sophisticated multi-agent workflows.

## Prerequisites

- Complete the [5-Minute Quickstart](quickstart.md)
- Basic understanding of AgenticGoKit concepts

## Learning Objectives

By the end of this tutorial, you'll understand:
- Different orchestration patterns and when to use them
- How to configure collaborative vs sequential processing
- How to build mixed orchestration workflows
- Performance characteristics of each pattern

## What You'll Build

Three different multi-agent systems:
1. **Collaborative System**: Agents work in parallel for faster processing
2. **Sequential Pipeline**: Agents work in sequence for data transformation
3. **Mixed Workflow**: Combination of parallel and sequential processing

---

## Part 1: Collaborative Orchestration (5 minutes)

Collaborative orchestration runs all agents in parallel, combining their outputs.

### Create a Collaborative Analysis System

```bash
# Create a collaborative analysis project
agentcli create analysis-system --template research-assistant
cd analysis-system
```

### Understanding the Generated Code

The generated `main.go` creates three agents that work together:

```go
// Generated agents work in parallel
agents := map[string]core.AgentHandler{
    "researcher": researcherHandler,
    "analyst":    analystHandler, 
    "validator":  validatorHandler,
}

// Collaborative runner sends events to ALL agents simultaneously
runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
```

### Test Collaborative Processing

```bash
# Set your API key
export OPENAI_API_KEY=your-api-key-here

# Run the collaborative system
go run main.go
```

You'll see all three agents process the same input simultaneously, then their outputs are combined.

### View the Collaboration

```bash
# Check the trace to see parallel execution
agentcli trace --flow-only <session-id>
```

**Expected Flow:**
```
14:32:15.123  researcher    (processing)
14:32:15.124  analyst       (processing)  ← All start simultaneously
14:32:15.125  validator     (processing)
14:32:18.456  researcher    (completed)
14:32:18.789  analyst       (completed)   ← Results combined
14:32:19.012  validator     (completed)
```

---

## Part 2: Sequential Pipeline (5 minutes)

Sequential orchestration processes agents one after another, passing data through a pipeline.

### Create a Data Processing Pipeline

```bash
# Create a sequential pipeline project
agentcli create data-pipeline --template data-pipeline
cd data-pipeline
```

### Understanding Sequential Processing

```go
// Generated sequential configuration
runner := core.NewRunnerWithOrchestration(core.EnhancedRunnerConfig{
    OrchestrationMode: core.OrchestrationSequential,
    SequentialAgents:  []string{"extractor", "transformer", "enricher", "formatter"},
    // ... other config
})
```

### Test Sequential Processing

```bash
export OPENAI_API_KEY=your-api-key-here
go run main.go
```

Each agent processes the output from the previous agent in sequence.

### View the Pipeline

```bash
agentcli trace --flow-only <session-id>
```

**Expected Flow:**
```
14:32:15.123  extractor     transformer
14:32:16.456  transformer   enricher      ← Sequential processing
14:32:17.789  enricher      formatter
14:32:19.012  formatter     (end)
```

---

## Part 3: Mixed Orchestration (5 minutes)

Mixed orchestration combines parallel and sequential processing for complex workflows.

### Create a Mixed Workflow System

```bash
# Create a mixed orchestration project
agentcli create content-system --orchestration collaborative --agents 5
cd content-system
```

### Understanding Mixed Processing

```go
// Mixed orchestration configuration
runner := core.NewRunnerWithOrchestration(core.EnhancedRunnerConfig{
    OrchestrationMode:   core.OrchestrationMixed,
    CollaborativeAgents: []string{"researcher", "fact-checker"}, // Run in parallel
    SequentialAgents:    []string{"writer", "editor", "publisher"}, // Run in sequence
})
```

### Test Mixed Processing

```bash
export OPENAI_API_KEY=your-api-key-here
go run main.go
```

The system first runs collaborative agents in parallel, then processes sequential agents with the combined results.

### View the Mixed Flow

```bash
agentcli trace --flow-only <session-id>
```

**Expected Flow:**
```
Phase 1 - Collaborative:
14:32:15.123  researcher    (processing)
14:32:15.124  fact-checker  (processing)  ← Parallel phase
14:32:17.456  researcher    (completed)
14:32:17.789  fact-checker  (completed)

Phase 2 - Sequential:
14:32:18.012  writer        editor        ← Sequential phase
14:32:19.345  editor        publisher
14:32:20.678  publisher     (end)
```

---

## Orchestration Patterns Comparison

| Pattern | Use Case | Pros | Cons |
|---------|----------|------|------|
| **Collaborative** | Analysis, validation, multiple perspectives | Fast (parallel), diverse outputs | Higher resource usage |
| **Sequential** | Data pipelines, step-by-step processing | Efficient, clear flow | Slower (serial) |
| **Mixed** | Complex workflows, content creation | Best of both worlds | More complex setup |

## Performance Characteristics

### Collaborative Orchestration
- **Speed**: Fastest for independent tasks
- **Resources**: Higher CPU/memory usage
- **Use When**: Tasks can be done independently

### Sequential Orchestration  
- **Speed**: Slower but predictable
- **Resources**: Lower resource usage
- **Use When**: Each step depends on the previous

### Mixed Orchestration
- **Speed**: Optimized for complex workflows
- **Resources**: Balanced usage
- **Use When**: Some tasks are independent, others dependent

## Advanced Configuration

### Timeout and Concurrency Settings

```bash
# Create with custom settings
agentcli create advanced-system --orchestration collaborative --agents 3
```

### Error Handling Configuration

```bash
# Create with fault tolerance
agentcli create resilient-system --orchestration collaborative --agents 4
```

## Troubleshooting

### Common Issues

**Agents not running in parallel:**
```bash
# Check orchestration mode in agentflow.toml
[orchestration]
mode = "collaborative"  # Should be collaborative, not route
```

**Sequential agents running out of order:**
```bash
# Verify agent sequence in configuration
sequential_agents = ["step1", "step2", "step3"]  # Order matters
```

**Mixed orchestration not working:**
```bash
# Ensure both agent lists are specified
collaborative_agents = ["agent1", "agent2"]
sequential_agents = ["agent3", "agent4"]
```

### Performance Issues

**Collaborative too slow:**
- Reduce number of agents
- Increase timeout settings
- Check agent complexity

**Sequential bottlenecks:**
- Identify slow agents with `agentcli trace --verbose`
- Optimize agent prompts
- Consider parallel alternatives

## Next Steps

Now that you understand orchestration patterns, you can:

1. **Add Memory**: Learn [Memory and RAG](memory-and-rag.md) to give agents persistent knowledge
2. **Add Tools**: Explore [Tool Integration](tool-integration.md) to connect external services
3. **Go Production**: Check [Production Deployment](production-deployment.md) for scaling

## Key Takeaways

- **Collaborative**: Use for independent tasks that benefit from multiple perspectives
- **Sequential**: Use for data pipelines where each step builds on the previous
- **Mixed**: Use for complex workflows that need both parallel and sequential processing
- **Configuration**: AgenticGoKit makes it easy to switch between patterns
- **Monitoring**: Always use `agentcli trace` to understand agent interactions

## Further Reading

- [Orchestration Patterns](../core-concepts/orchestration-patterns.md) - Deep dive into patterns
- [Performance Optimization](../advanced/README.md) - Advanced performance tuning
- [State Management](../core-concepts/state-management.md) - How data flows between agents
