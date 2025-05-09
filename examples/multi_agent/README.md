# Multi-Agent Research Workflow Example (Factory Version)

This example demonstrates how to implement a multi-agent research workflow using the AgentFlow framework and the factory-based runner setup. It showcases a pipeline where specialized agents collaborate to process a user request, perform research, summarize findings, and present the final output.

## Architecture

- **Planner Agent**: Receives the initial user request and generates a research plan using Azure OpenAI.
- **Researcher Agent**: Takes the plan, simulates research (can be extended to use real tools), and produces research results.
- **Summarizer Agent**: Summarizes the research results using Azure OpenAI.
- **Final Output Agent**: Collects the final state and signals workflow completion.

## Features
- Modular agent design with clear responsibilities
- Uses Azure OpenAI for planning and summarization
- Tool registry for extensible research tools (e.g., web search)
- Tracing and state propagation between agents
- Factory function for concise runner setup

## Prerequisites
- Go 1.21 or later
- Azure OpenAI Service credentials:
  - Endpoint URL
  - API Key
  - Chat Deployment name
  - Embedding Deployment name (optional)

## Setup
Set the following environment variables with your Azure OpenAI credentials:

```sh
export AZURE_OPENAI_ENDPOINT="https://your-resource-name.openai.azure.com"
export AZURE_OPENAI_API_KEY="your-api-key"
export AZURE_OPENAI_CHAT_DEPLOYMENT="gpt-35-turbo" # or your deployment name
export AZURE_OPENAI_EMBEDDING_DEPLOYMENT="text-embedding-ada-002" # optional
```

## Running the Example
From the project root directory:

```sh
go run ./examples/multi_agent/
```

The application will:
- Initialize the agents and tool registry
- Set up the research workflow using the factory runner
- Process a sample research request
- Save the execution trace to the `traces/` directory
- Output the final summary to the console

## Trace Visualization
You can inspect the execution trace using the CLI:

```sh
agentcli trace --flow-only <session-id>
```

Replace `<session-id>` with the actual session ID printed in the console or found in the `traces/` directory.

## Extending the Example
- Add more tools to the tool registry for richer research capabilities
- Integrate real web search or data APIs in the Researcher agent
- Customize prompts and agent logic for your use case

---
For more details, see the [AgentFlow documentation](../../docs/DevGuide.md).
