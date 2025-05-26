# AgentFlow Core Package

This package contains the public API for the AgentFlow framework. Use this package to build agent-based applications and workflows.

## Installation

```bash
go get github.com/kunalkushwaha/agentflow@latest
```

## Usage

```go
import agentflow "github.com/kunalkushwaha/agentflow/core"
```

## Key Types

- **`AgentHandler`**: Interface for implementing custom agents
- **`Runner`**: Orchestrates agent execution and event processing
- **`Event`**: Represents data and metadata flowing between agents
- **`State`**: Manages data state throughout workflow execution
- **`AgentResult`**: Contains the output of agent execution

## Quick Example

```go
package main

import (
    "context"
    "log"
    "time"
    agentflow "github.com/kunalkushwaha/agentflow/core"
)

type MyAgent struct{}

func (a *MyAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
    outputState := agentflow.NewState()
    outputState.Set("processed", true)
    
    return agentflow.AgentResult{
        OutputState: outputState,
        StartTime:   time.Now(),
        EndTime:     time.Now(),
        Duration:    time.Millisecond * 10,
    }, nil
}

func main() {
    agents := map[string]agentflow.AgentHandler{
        "my-agent": &MyAgent{},
    }

    runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
        Agents:    agents,
        QueueSize: 10,
    })

    // Start runner and emit events...
}
```

## Documentation

- [Developer Guide](../docs/DevGuide.md)
- [Library Integration Guide](../docs/agentflow_library_integration.md)
- [Architecture Overview](../docs/Architecture.md)
- [Examples](../examples/README.md)

## LLM Integration

To integrate with LLMs, use the public core API:

```go
import "github.com/kunalkushwaha/agentflow/core"

adapter := core.NewOpenAIAdapter("your-api-key")
model := core.ModelProvider(adapter)

// Use model in your agent...
```

Refer to the [LLM Integration Guide](../docs/LLMIntegration.md) for detailed instructions.
