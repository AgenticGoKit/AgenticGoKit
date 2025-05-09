# Minimal AgentFlow Example

This example demonstrates the simplest possible usage of the AgentFlow framework. It shows how to:

- Define a minimal agent
- Register the agent with the runner
- Emit a single event
- Observe agent execution and event processing

## How it works
- A `MinimalAgent` is registered with the runner.
- The runner is started, and a single event is emitted.
- The agent processes the event and prints a log message.
- The runner is stopped after a short delay.

## Running the Example
From the project root directory:

```sh
go run ./examples/minimal-example/
```

## Output
You should see a log message from the minimal agent and basic runner logs.

## Use Case
Use this example as a starting point for building more complex agent workflows with AgentFlow.

---
For more details, see the [AgentFlow documentation](../../docs/DevGuide.md).
