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
## License
This project is licensed under the MIT License
