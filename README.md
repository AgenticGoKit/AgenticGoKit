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
## Quick Start
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    agentflow "kunalkushwaha/agentflow/internal/core"
    "kunalkushwaha/agentflow/internal/orchestrator"
)

func main() {
    // Create an orchestrator and runner
    orch := orchestrator.NewRouteOrchestrator()
    runner := agentflow.NewRunner(orch, 10) // 10-event queue

    // Create a simple agent
    agent := &SimpleAgent{}
    
    // Register the agent
    if err := runner.RegisterAgent("simple_agent", agent); err != nil {
        log.Fatalf("Failed to register agent: %v", err)
    }
    
    // Set up tracing
    traceLogger := agentflow.NewInMemoryTraceLogger()
    runner.SetTraceLogger(traceLogger)
    agentflow.RegisterTraceHooks(runner.CallbackRegistry(), traceLogger)
    
    // Start the runner
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    runner.Start()
    
    // Create and emit an event
    sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
    event := agentflow.NewEvent(
        "my_event",
        map[string]interface{}{"message": "Hello Agentflow!"},
        map[string]string{
            "route": "simple_agent",
            "session_id": sessionID,
        },
    )
    
    runner.Emit(event)
    
    // Wait for processing
    time.Sleep(1 * time.Second)
    
    // Dump the trace
    trace, _ := runner.DumpTrace(sessionID)
    fmt.Printf("Trace: %+v\n", trace)
}

// SimpleAgent implements the Agent interface
type SimpleAgent struct{}

func (a *SimpleAgent) Name() string {
    return "simple_agent"
}

func (a *SimpleAgent) Run(ctx context.Context, state agentflow.State) (agentflow.State, error) {
    // Get input from state
    input, _ := state.Get("message")
    
    // Process input
    output := fmt.Sprintf("Processed: %v", input)
    
    // Create new state with output
    newState := state.Clone()
    newState.Set("response", output)
    
    return newState, nil
}
```
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
- Developer Guide - Comprehensive guide to using the framework
- Tracing Guide - Details on the tracing system
- Architecture - High-level architecture overview

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
