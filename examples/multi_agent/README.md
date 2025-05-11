# Multi-Agent Research Workflow Example (Factory Version)

This example demonstrates a multi-agent research workflow using the `AgentFlow` framework and its factory-based runner setup. The workflow involves specialized agents collaborating to process a user request, perform research, summarize findings, and present the final output. The example highlights the use of the `Runner` for routing events between agents and managing the workflow lifecycle.

## Architecture

- **Planner Agent**: Receives the initial user request and generates a research plan using Azure OpenAI.
- **Researcher Agent**: Takes the plan, simulates research (can be extended to use real tools), and produces research results.
- **Summarizer Agent**: Summarizes the research results using Azure OpenAI.
- **Final Output Agent**: Collects the final state and signals workflow completion.

## Features

- Modular agent design with clear responsibilities.
- Uses Azure OpenAI for planning and summarization.
- Tool registry for extensible research tools (e.g., web search).
- Event routing and state propagation between agents.
- Factory function for concise runner setup.

## How It Works

### Key Components

1. **Runner**:
   - The `Runner` is responsible for managing agents and routing events between them.
   - It provides methods like `Start`, `Stop`, and `Emit` to control the flow of events.
   - Events are routed based on metadata, allowing seamless transitions between agents.

2. **Agents**:
   - Each agent implements the `agentflow.AgentHandler` interface.
   - Agents process events, update the state, and route the event to the next agent.

3. **Event Routing**:
   - Events are emitted with metadata specifying the target agent.
   - The `Runner` uses this metadata to determine the next agent in the workflow.

### Workflow Steps

1. **Initialize the Runner**:
   - The `Runner` is created using the factory function and configured with a map of agents.

2. **Emit Initial Event**:
   - The workflow starts by emitting an event to the `Planner` agent with the user request.

3. **Agent Collaboration**:
   - The `Planner` generates a research plan and routes the event to the `Researcher`.
   - The `Researcher` simulates research and routes the event to the `Summarizer`.
   - The `Summarizer` creates a summary and routes the event to the `Final Output` agent.

4. **Workflow Completion**:
   - The `Final Output` agent collects the final state and signals the workflow's completion.

## Running the Example

1. Set the following environment variables with your Azure OpenAI credentials:

   ```sh
   export AZURE_OPENAI_ENDPOINT="https://your-resource-name.openai.azure.com"
   export AZURE_OPENAI_API_KEY="your-api-key"
   export AZURE_OPENAI_CHAT_DEPLOYMENT="gpt-35-turbo" # or your deployment name
   export AZURE_OPENAI_EMBEDDING_DEPLOYMENT="text-embedding-ada-002" # optional
   ```

2. Run the example:

   ```sh
   go run ./examples/multi_agent/
   ```

3. The application will:
   - Initialize the agents and tool registry.
   - Set up the research workflow using the factory runner.
   - Process a sample research request.
   - Save the execution trace to the `traces/` directory.
   - Output the final summary to the console.

## Trace Visualization

You can inspect the execution trace using the CLI:

```sh
agentcli trace --flow-only <session-id>
```

Replace `<session-id>` with the actual session ID printed in the console or found in the `traces/` directory.

## Extending the Example

- Add more tools to the tool registry for richer research capabilities.
- Integrate real web search or data APIs in the `Researcher` agent.
- Customize prompts and agent logic for your use case.

---
For more details, see the [AgentFlow documentation](../../docs/DevGuide.md).
