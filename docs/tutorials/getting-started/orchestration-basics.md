# Orchestration Basics (Config-first)

Goal: Learn how to switch between orchestration modes to control agent interactions.

## Prerequisites
Complete [first-agent.md](first-agent.md) to understand basic project structure.

## What is Orchestration?
Orchestration controls how multiple agents work together:
- **route**: Single agent handles each request
- **collaborative**: Agents work in parallel and combine results  
- **sequential**: Agents work in a specific order
- **loop**: Single agent iterates until condition is met
- **mixed**: Combination of collaborative and sequential

## 1) Create a multi-agent project
```pwsh
agentcli create orchestration-demo --template research-assistant
Set-Location orchestration-demo
```

This creates a project with multiple agents: researcher, analyzer, synthesizer.

## 2) Try collaborative mode (default)
Check your `agentflow.toml`:
```toml
[orchestration]
mode = "collaborative"
timeout_seconds = 30
collaborative_agents = ["researcher", "analyzer", "synthesizer"]
```

Run the project:
```pwsh
agentcli validate
go run . -m "Research the benefits of electric vehicles"
```

Expected: All agents process the query in parallel and combine their results.

## 3) Switch to sequential mode
Edit `agentflow.toml`:
```toml
[orchestration]
mode = "sequential"
sequential_agents = ["researcher", "analyzer", "synthesizer"]
```

Run again:
```pwsh
agentcli validate
go run . -m "Research the benefits of electric vehicles"
```

Expected: Researcher runs first, then analyzer processes researcher's output, then synthesizer creates final result.

## 4) Try route mode
Edit `agentflow.toml`:
```toml
[orchestration]
mode = "route"
```

Run again:
```pwsh
agentcli validate
go run . -m "Research the benefits of electric vehicles"
```

Expected: System routes to the most appropriate single agent based on the request.

## 5) Inspect execution (optional)
If available in your CLI version:
```pwsh
agentcli trace --help
```

## Key Differences
- **Collaborative**: Fast, parallel processing, best for diverse perspectives
- **Sequential**: Structured pipeline, each agent builds on previous results
- **Route**: Single expert agent, fastest execution, good for specialized tasks
- **Loop**: Iterative refinement, good for improving quality over multiple passes

## Next Steps
- Add memory to agents: [memory-basics.md](memory-basics.md)
- Enable tools: [tools-basics.md](tools-basics.md)

## Verification checklist
- [ ] agentcli create succeeded
- [ ] Tried at least 2 different orchestration modes
- [ ] agentcli validate passed for each mode
- [ ] Observed different behavior between modes
