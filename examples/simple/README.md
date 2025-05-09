# Simple Chat Agent Example (No Factory)

This example demonstrates how to implement a basic LLM-powered agent using the AgentFlow framework **without using a factory function**. It shows how to:

- Manually configure and instantiate an Azure OpenAI adapter
- Implement a simple agent that calls the LLM and processes user prompts
- Prepare and pass state to the agent
- Run the agent and print the LLM's response

## How it works
- The `ChatAgent` receives a user prompt from the state
- It calls the Azure OpenAI LLM using the provided credentials
- The LLM's response is added to the state and printed to the console
- All setup and execution is done manually, without a factory or workflow runner

## Running the Example
From the project root directory:

```sh
go run ./examples/simple/agent.go
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
- The LLM's response to the hardcoded user prompt is printed to the console
- Log messages show the agent's execution steps

## Use Case
Use this example as a starting point for direct LLM integration, or as a reference for manual agent setup without the AgentFlow factory or runner abstractions.

---
For more details, see the [AgentFlow documentation](../../docs/DevGuide.md).
