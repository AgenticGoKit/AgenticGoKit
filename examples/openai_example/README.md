# OpenAI Example

This example demonstrates how to use the `OpenAIAdapter` in conjunction with the `agentflow` framework's `Runner`. It is designed for more complex use cases where you need to manage multiple agents, route events, and orchestrate workflows.

## Background

The `agentflow` framework provides a powerful abstraction for managing agents and routing events. The `Runner` is a key component of this framework, enabling you to:

1. Register multiple agents.
2. Emit events that are routed to the appropriate agent based on metadata.
3. Start and stop the framework, managing the lifecycle of agents and events.

This example showcases how to:

1. Initialize the `OpenAIAdapter`.
2. Register the adapter as an agent in the `Runner`.
3. Emit events to interact with the agent and process responses.

## How It Works

### Key Components

1. **Runner**:
   - The `Runner` is responsible for managing agents and routing events to them.
   - It provides methods like `Start`, `Stop`, and `Emit` to control the flow of events.

2. **Agent**:
   - The `OpenAIAgent` implements the `agentflow.AgentHandler` interface.
   - It uses the `OpenAIAdapter` to process prompts and fetch responses from OpenAI's API.

3. **Event Routing**:
   - Events are emitted with metadata that specifies the target agent.
   - The `Runner` routes the event to the appropriate agent based on this metadata.

### Code Walkthrough

#### Main Function

The `main` function demonstrates the following steps:

1. **Set Log Level**:
   ```go
   agentflow.SetLogLevel(agentflow.INFO)
   ```

2. **Initialize the Adapter**:
   ```go
   adapter, err := llm.NewOpenAIAdapter(apiKey, "gpt-4o-mini", 100, 0.7)
   if err != nil {
       log.Fatalf("Failed to create OpenAIAdapter: %v", err)
   }
   ```

3. **Register the Agent**:
   ```go
   agents := map[string]agentflow.AgentHandler{
       "openai": &OpenAIAgent{adapter: adapter},
   }
   ```

4. **Create and Start the Runner**:
   ```go
   runner := factory.NewRunnerWithConfig(factory.RunnerConfig{
       Agents: agents,
   })

   runner.Start(context.Background())
   defer runner.Stop()
   ```

5. **Emit an Event**:
   ```go
   runner.Emit(agentflow.NewEvent(
       "question",
       map[string]interface{}{
           "user_prompt": "What is the capital of France?",
       },
       map[string]string{agentflow.RouteMetadataKey: "openai"},
   ))
   ```

6. **Wait for the Agent to Process the Event**:
   ```go
   time.Sleep(500 * time.Millisecond)
   ```

## When to Use This Example

This example is ideal for:
- Applications requiring multiple agents and complex workflows.
- Scenarios where event routing and orchestration are necessary.
- Learning how to integrate the `OpenAIAdapter` with the `agentflow` framework.

For simpler use cases, refer to the `ollama_example`, which demonstrates direct interaction with an LLM without using the `Runner`.

## Running the Example

1. Ensure you have the required dependencies installed.
2. Set the `OPENAI_API_KEY` environment variable with your API key.
3. Run the example:
   ```bash
   go run main.go
   ```

## Output

The application will log the response from the OpenAI API and indicate whether the execution succeeded or failed.
