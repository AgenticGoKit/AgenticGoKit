# Memory Agent Example

This example demonstrates a simple agent that stores received event data in memory using the AgentFlow framework. It shows how to:

- Implement a custom agent that persists data in memory
- Register the agent with a route orchestrator
- Emit and process multiple events
- Use callback hooks and trace logging
- Gracefully handle shutdown and verify stored data

## How it works
- The `MemoryAgent` receives events, stores their data by event ID, and updates the state.
- The runner and orchestrator are set up with tracing enabled.
- Events are emitted to the agent, which processes and stores them.
- On shutdown, the example prints all stored data and writes the trace to a JSON file.

## Running the Example
From the project root directory:

```sh
go run ./examples/memory_agent/
```

## Output
- Log messages showing event processing and memory storage
- A summary of all stored data for each event
- A trace file (e.g., `session-<timestamp>.trace.json`) with the full execution trace

## Use Case
Use this example as a template for building stateful agents or for learning how to:
- Store and retrieve data in custom agents
- Integrate tracing and callback hooks
- Manage graceful shutdown in event-driven workflows

---
For more details, see the [AgentFlow documentation](../../docs/DevGuide.md).
