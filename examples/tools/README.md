# Tool Usage Example (LLM-Driven)

This example demonstrates how to use the AgentFlow framework to build an agent that leverages both LLMs and a tool registry. The agent uses an LLM (Azure OpenAI) to decide which tool to call and with what arguments, then executes the tool and returns the result.

## Features
- LLM-driven tool selection and argument extraction
- Tool registry for extensible tool integration (e.g., arithmetic, web search)
- Manual runner and orchestrator setup (no factory abstraction)
- Graceful shutdown and signal handling
- Tracing and callback hooks can be added as needed

## How it works
- The `ToolAgent` receives a user request
- It prompts the LLM to decide if a tool should be called and with what arguments
- If a tool is selected, the agent calls it via the registry and adds the result to the state
- The result is printed to the console

## Running the Example
From the project root directory:

```sh
go run ./examples/tools/tool_usage_example.go
```

## Prerequisites
- Go 1.21 or later
- Azure OpenAI Service credentials:
  - Endpoint URL
  - API Key
  - Chat Deployment name
  - Embedding Deployment name (can be a dummy value for this example)

Set the following environment variables:

```sh
export AZURE_OPENAI_ENDPOINT="https://your-resource-name.openai.azure.com"
export AZURE_OPENAI_API_KEY="your-api-key"
export AZURE_OPENAI_CHAT_DEPLOYMENT="gpt-35-turbo" # or your deployment name
export AZURE_OPENAI_EMBEDDING_DEPLOYMENT="dummy-value" # required by constructor
```

## Output
- The result of the tool call (e.g., arithmetic calculation) is printed to the console
- Log messages show the agent's execution and LLM/tool decisions

## Use Case
Use this example as a template for building LLM-augmented tool agents, or as a reference for integrating tool registries and LLMs in AgentFlow.

---
For more details, see the [AgentFlow documentation](../../docs/DevGuide.md).
