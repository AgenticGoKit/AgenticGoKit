# AgentFlow Examples

This directory contains a variety of examples demonstrating how to use the `AgentFlow` framework for building intelligent, event-driven workflows. Each example showcases different features and use cases of the framework, ranging from simple single-agent setups to complex multi-agent workflows.

## Examples Overview

### [Minimal Example](./minimal-example/README.md)
- **Description**: A basic example to get started with `AgentFlow`.
- **Highlights**:
  - Demonstrates the minimal setup required to create and run an agent.
  - Ideal for beginners exploring the framework.

### [Memory Agent](./memory_agent/README.md)
- **Description**: Demonstrates a stateful agent that stores event data in memory.
- **Highlights**:
  - Shows how to implement custom agents with memory.
  - Includes tracing and callback hooks.
  - Useful for learning state management in event-driven workflows.

### [Multi-Agent Workflow](./multi_agent/README.md)
- **Description**: A complex example showcasing a multi-agent research workflow.
- **Highlights**:
  - Uses Azure OpenAI for planning and summarization.
  - Demonstrates event routing and state propagation between agents.
  - Includes a factory-based runner setup for modular workflows.

### [Clean Multi-Agent Workflow](./clean_multi_agent/README.md)
- **Description**: A cleaner and more modular version of the multi-agent workflow.
- **Highlights**:
  - Focuses on code organization and reusability.
  - Ideal for developers looking to build scalable workflows.

### [Ollama Example](./ollama_example/README.md)
- **Description**: Demonstrates how to use the `OllamaAdapter` for direct LLM-based applications.
- **Highlights**:
  - Focuses on simple use cases without the need for a runner.
  - Ideal for prototyping and standalone LLM interactions.

### [OpenAI Example](./openai_example/README.md)
- **Description**: Demonstrates how to use the `OpenAIAdapter` with the `AgentFlow` runner.
- **Highlights**:
  - Showcases event routing and agent orchestration.
  - Suitable for workflows requiring multiple agents.

### [Orchestrator Example](./orchestrator/README.md)
- **Description**: Demonstrates the use of the `Orchestrator` for managing agent collaboration.
- **Highlights**:
  - Focuses on advanced routing and collaboration between agents.
  - Useful for building complex workflows with multiple dependencies.

### [Tools Example](./tools/README.md)
- **Description**: Demonstrates how to use and extend tools in the `AgentFlow` framework.
- **Highlights**:
  - Includes examples of tool registration and usage.
  - Ideal for developers looking to integrate external tools into their workflows.

---

## How to Run Examples

1. Navigate to the example directory you want to explore.
2. Follow the instructions in the `README.md` file of the example.
3. Run the example using:
   ```sh
   go run ./examples/<example-folder>/
   ```

## Choosing the Right Example
- **Beginner**: Start with the [Minimal Example](./minimal-example/README.md).
- **Stateful Agents**: Explore the [Memory Agent](./memory_agent/README.md).
- **Complex Workflows**: Check out the [Multi-Agent Workflow](./multi_agent/README.md) or [Clean Multi-Agent Workflow](./clean_multi_agent/README.md).
- **LLM Integration**: Use the [Ollama Example](./ollama_example/README.md) or [OpenAI Example](./openai_example/README.md).
- **Advanced Routing**: Dive into the [Orchestrator Example](./orchestrator/README.md).
- **Tool Integration**: Learn from the [Tools Example](./tools/README.md).

For more details, refer to the [AgentFlow documentation](../docs/DevGuide.md).
