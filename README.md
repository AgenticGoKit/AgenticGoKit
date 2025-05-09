# Agentflow

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Agentflow is a Go framework for building AI agent systems. It provides core abstractions for event-based workflows, agent coordination, and tracing capabilities, enabling the creation of sophisticated multi-agent applications.

## Features

- **Event-driven Architecture**: Process events through configurable orchestration patterns
- **Multi-modal Orchestration**: Choose between route (single-agent) or collaborate (multi-agent) execution modes
- **Deterministic Workflow Agents**: Build pipelines with SequentialAgent, ParallelAgent, and LoopAgent
- **LLM Integration**: Abstract any LLM backend via unified ModelProvider interface (Azure OpenAI, OpenAI, Ollama)
- **Tool Ecosystem**: Extend agent capabilities with function tool registry
- **Observability**: Comprehensive tracing and callback hooks at key lifecycle points
- **Memory Management**: Both short-term session storage and long-term vector-based memory
- **CLI Support**: Built-in command-line tools for trace inspection and debugging

## Installation

```bash
go get github.com/kunalkushwaha/agentflow@latest
```

## Quick Start (with Factory)

```go
// See more in [DevGuide.md](docs/DevGuide.md#quick-start-with-factory)
import (
    "context"
    "kunalkushwaha/agentflow/internal/core"
    "kunalkushwaha/agentflow/internal/factory"
)

func main() {
    agents := map[string]agentflow.AgentHandler{
        "echo": agentflow.AgentHandlerFunc(
            func(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
                state.Set("echo", event.GetData()["message"])
                return agentflow.AgentResult{OutputState: state}, nil
            }),
    }
    runner := factory.NewRunnerWithConfig(factory.RunnerConfig{
        Agents: agents,
    })
    runner.Start(context.Background())
    runner.Emit(core.NewEvent("echo", core.EventData{"message": "hello"}, nil))
    // ...wait, then runner.Stop()
}
```

## When to Use Factory Functions

- Use the factory for most production and prototype workflows.
- Use manual setup only if you need custom callback wiring, advanced orchestrator logic, or deep integration with external systems.

## Adding Custom Callbacks with Factory

You can register custom callbacks after creating the runner:

```go
runner := factory.NewRunnerWithConfig(factory.RunnerConfig{Agents: agents})
registry := runner.GetCallbackRegistry()
registry.Register(agentflow.HookAfterAgentRun, "myCustomLogger", myCallbackFunc)
```

## Error Handler Agent

The factory will register a default "error-handler" agent if not provided. To override, add your own to the `Agents` map with the key `"error-handler"`.

## Troubleshooting / FAQ

- **My agent isn't called?**
  - Check that your event's metadata includes the correct `RouteMetadataKey`.
- **How do I see the trace?**
  - Use `runner.DumpTrace(sessionID)` or the `agentcli trace` command. See [Tracing Guide](docs/TracingGuide.md).
- **How do I add more tools/LLMs?**
  - Use `factory.NewDefaultToolRegistry()` and `factory.NewDefaultLLMAdapter()` or register your own.

## See Also
- [Multi-Agent Example](examples/multi_agent/)
- [Clean Multi-Agent Example](examples/clean_multi_agent/)
- [Factory Implementation](internal/factory/agent_factory.go)
- [Developer Guide](docs/DevGuide.md)
- [Tracing Guide](docs/TracingGuide.md)
- [Architecture Overview](docs/Architecture.md)
- [Project Roadmap](docs/ROADMAP.md)

## Project Structure

```
agentflow/
├── cmd/
│   └── agentcli/           # CLI tools for trace inspection
├── internal/               # Core framework code
│   ├── core/               # Core abstractions (Event, State, Runner)
│   ├── orchestrator/       # Orchestration strategies
│   ├── agents/             # Workflow agent implementations
│   ├── llm/                # LLM adapters and interfaces
│   ├── tools/              # Tool registry and implementations
│   └── memory/             # Memory and session services
├── examples/               # Example implementations
├── docs/                   # Documentation
└── benchmarks/             # Performance benchmarks
```

## Documentation
- [Developer Guide](docs/DevGuide.md) - Comprehensive guide to using the framework
- [Tracing Guide](docs/TracingGuide.md) - Details on the tracing system
- [Architecture Overview](docs/Architecture.md) - High-level architecture overview
- [Project Roadmap](docs/ROADMAP.md) - Development timeline and upcoming features

## LLM Integration
Agentflow provides a unified interface for different LLM backends:

```go
// Create Azure OpenAI adapter
options := llm.AzureOpenAIAdapterOptions{
    Endpoint:            "https://your-resource-name.openai.azure.com",
    APIKey:              os.Getenv("AZURE_OPENAI_API_KEY"),
    ChatDeployment:      "gpt-4-turbo",
    EmbeddingDeployment: "text-embedding-3",
}

azureLLM, err := llm.NewAzureOpenAIAdapter(options)
if err != nil {
    log.Fatalf("Failed to create Azure OpenAI adapter: %v", err)
}
```
## Tool Integration
Register and use tools to extend agent capabilities:

```go
// Create a tool registry
registry := tools.NewToolRegistry()

// Register tools
registry.Register(&tools.WebSearchTool{})
registry.Register(&tools.ComputeMetricTool{})

// Use the registry in an agent
agent := &ToolAgent{registry: registry}
```

## AgentCLI - Command Line Tooling

Agentflow includes a powerful command-line interface tool (`agentcli`) that provides diagnostic capabilities for monitoring and debugging agent workflows.

### Installation

```bash
# Build and install the CLI tool
go install ./cmd/agentcli
```
Key Commands

## Trace Inspection
The trace command visualizes execution traces for debugging agent workflows:
```bash
# View complete trace with all details
agentcli trace <sessionID>

# View only the flow between agents (simplified view)
agentcli trace --flow-only <sessionID>

# Filter trace entries for a specific agent
agentcli trace --filter agent=researcher <sessionID>

# Show detailed state information (verbose mode)
agentcli trace --verbose <sessionID>

# Debug raw JSON structure
agentcli trace --debug <sessionID>
```
### Features
- Table Visualization: Formats trace data in clear, structured tables
- Flow Analysis: Shows the sequence of agent interactions
- Filtering: Focus on specific agents or events
- State Inspection: Examine state transitions between agents
- Error Analysis: Quickly identify and diagnose failures

### Example Output
```bash
┌───────────────────┬────────────────┬────────────────┬──────────────────────────────┬────────────────────────────────────────┐
│ TIMESTAMP         │ HOOK           │ AGENT          │ STATE                        │ ERROR                                  │
├───────────────────┼────────────────┼────────────────┼──────────────────────────────┼────────────────────────────────────────┤
│ 13:20:37.976      │ BeforeEvent... │                │ {user_request: "Research...  │ -                                      │
│ 13:20:37.977      │ BeforeAgentRun │ planner        │ {user_request: "Research...  │ -                                      │
│ 13:20:43.867      │ AfterAgentRun  │ planner        │ {plan: "1. Research recen... │ -                                      │
│ 13:20:43.869      │ AfterEventH... │ planner        │ {plan: "1. Research recen... │ -                                      │
└───────────────────┴────────────────┴────────────────┴──────────────────────────────┴────────────────────────────────────────┘
```

for flow only
```bash
TIME         AGENT        NEXT         HOOK          EVENT ID
16:20:49.670 planner      researcher   AfterAgentRun req-1746...
16:20:49.671 researcher   summarizer   AfterAgentRun req-1746...
16:20:59.130 planner      researcher   AfterAgentRun 5b0b5e18...
16:21:03.875 summarizer   final_output AfterAgentRun req-1746...
16:21:04.879 researcher   summarizer   AfterAgentRun 5b0b5e18...
16:21:05.883 final_output final_output AfterAgentRun req-1746...

Sequence diagram:
----------------
1. planner → researcher
2. researcher → summarizer
3. planner → researcher
4. summarizer → final_output
5. researcher → summarizer
6. final_output → final_output

Condensed route:
planner → researcher → summarizer → final_output

Per‑event sequence diagrams:
-----------------------------

[req-1746…]
planner → researcher
researcher → summarizer
summarizer → final_output
final_output → final_output  [requeue]

[5b0b5e18]
planner → researcher  [requeue]
researcher → summarizer
```

## Project Status

AgentFlow is under active development. Below is the current status of features based on our project roadmap.

### ✅ Completed Features

#### Core Infrastructure
- **Event System**: Uniform event interface with ID, payload, and metadata
- **Runner Service**: Core event processing with agent registration and event emission
- **Orchestration Modes**: 
  - Route Orchestrator for single-agent routing
  - Collaborate Orchestrator for parallel processing

#### Observability
- **Tracing System**: Comprehensive trace logging for all agent interactions
- **CLI Tool**: Command-line interface for trace inspection and visualization
- **Callback Hooks**: Pre and post execution hooks for all lifecycle events

#### LLM Integration
- **ModelProvider Interface**: Abstraction for different LLM backends
- **Azure OpenAI Adapter**: Integration with Azure OpenAI Service


#### Agent Workflows
```mermaid
graph TD
    subgraph "Sequential Agent"
        SeqAgent[SequentialAgent] --> Seq1[Agent 1]
        Seq1 --> Seq2[Agent 2]
        Seq2 --> Seq3[Agent 3]
    end
    
    subgraph "Parallel Agent"
        ParAgent[ParallelAgent] --> Par1[Agent 1]
        ParAgent --> Par2[Agent 2]
        ParAgent --> Par3[Agent 3]
        Par1 -->|Merge Results| ParResult[Aggregated Result]
        Par2 -->|Merge Results| ParResult
        Par3 -->|Merge Results| ParResult
    end
    
    subgraph "Loop Agent"
        LoopAgent --> SubAgent[Agent]
        SubAgent -->|Condition False| LoopAgent
        SubAgent -->|Condition True| LoopResult[Final Result]
        SubAgent -->|Max Iterations| LoopResult
    end
    
    style SeqAgent fill:#f9f,stroke:#333
    style ParAgent fill:#bbf,stroke:#333
    style LoopAgent fill:#bfb,stroke:#333
  ```
- **Deterministic Workflow Agents**: Implementation of workflow patterns
  - **Sequential Agent**: Ordered execution with state propagation
  - **Parallel Agent**: Concurrent execution with result aggregation
  - **Loop Agent**: Condition-based iteration with safety limits

#### Tool Integration
- **Tool Registry**: Framework for registering and invoking agent tools

### 🚧 In Progress

#### Tool Integration
- **Built-in Tools**: Implementation of common tools like web search

### 📋 Upcoming Features

#### Memory Systems
- **Session Storage**: In-memory conversation state management
- **Vector Memory**: Long-term storage with embedding-based retrieval
- **Artifact Management**: File storage service for agent outputs

#### API & Deployment
- **REST API**: HTTP endpoints for event submission and state retrieval
- **Developer UI**: Web dashboard for trace visualization
- **Containerization**: Docker and Helm support for Kubernetes deployment

See our [detailed roadmap](docs/ROADMAP.md) for more information on the development timeline.

## License
This project is licensed under the MIT License
