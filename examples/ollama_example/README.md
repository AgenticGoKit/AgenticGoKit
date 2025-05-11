# Ollama Example

This example demonstrates how to use the `OllamaAdapter` to build a simple LLM-based application without relying on the `agentflow` framework's `Runner`. It is designed for straightforward use cases where you want to directly interact with the LLM and process responses without the additional complexity of event routing or agent orchestration.

## Background

The `agentflow` framework provides powerful tools for managing agents, routing events, and orchestrating workflows. However, for simple applications where you only need to interact with a single LLM, the overhead of setting up a `Runner` and managing events may not be necessary. This example focuses on such a scenario, showcasing how to:

1. Initialize the `OllamaAdapter`.
2. Use the adapter to process a user prompt and fetch a response.
3. Handle the response directly in your application logic.

## How It Works

1. **Adapter Initialization**:
   - The `OllamaAdapter` is initialized with the required API key, model name, maximum tokens, and temperature.

2. **Direct Interaction**:
   - The `OllamaAgent` implements the `agentflow.AgentHandler` interface.
   - The `Run` method combines a system prompt and a user prompt into a single request and sends it to the `OllamaAdapter` using the `Call` method.

3. **Response Handling**:
   - The response from the LLM is logged and returned as part of the agent's result.

## Code Walkthrough

### Main Function

The `main` function demonstrates the following steps:

1. **Set Log Level**:
   ```go
   agentflow.SetLogLevel(agentflow.INFO)
   ```

2. **Initialize the Adapter**:
   ```go
   adapter, err := llm.NewOllamaAdapter("test-key", "gemma3:latest", 100, 0.7)
   if err != nil {
       agentflow.Logger().Error().Msgf("Failed to create OllamaAdapter: %v", err)
       return
   }
   ```

3. **Create an Agent and Event**:
   ```go
   agent := &OllamaAgent{adapter: adapter}
   event := &agentflow.SimpleEvent{
       Data: agentflow.EventData{"user_prompt": "What is the capital of France?"},
   }
   state := agentflow.NewState()
   ```

4. **Run the Agent**:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   defer cancel()

   result, err := agent.Run(ctx, event, state)
   if err != nil {
       log.Fatalf("Agent execution failed: %v", err)
   }
   agentflow.Logger().Debug().Msgf("Agent execution succeeded: %+v", result)
   ```

## When to Use This Example

This example is ideal for:
- Prototyping simple LLM-based applications.
- Scenarios where you don't need complex event routing or multiple agents.
- Learning how to use the `OllamaAdapter` in isolation.

For more advanced use cases involving multiple agents, event routing, and orchestration, refer to the `openai_example` or other examples in the `agentflow` repository.

## Running the Example

1. Ensure you have the required dependencies installed.
2. Set the `OLLAMA_API_KEY` environment variable with your API key.
3. Run the example:
   ```bash
   go run main.go
   ```

## Output

The application will log the response from the LLM and indicate whether the execution succeeded or failed.
