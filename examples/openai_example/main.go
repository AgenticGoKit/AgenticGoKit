package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/factory"
	"kunalkushwaha/agentflow/internal/llm"
)

// OpenAIAgent implements agentflow.AgentHandler
type OpenAIAgent struct {
	adapter *llm.OpenAIAdapter
}

func (a *OpenAIAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	systemPrompt := "You are a helpful assistant."

	// Fetch the user prompt from the event payload
	userPrompt, ok := event.GetData()["user_prompt"].(string)
	if !ok || userPrompt == "" {
		return agentflow.AgentResult{}, errors.New("user_prompt is missing or invalid in the event payload")
	}

	response, err := a.adapter.Call(ctx, systemPrompt+"\n"+userPrompt)
	if err != nil {
		return agentflow.AgentResult{}, err
	}

	agentflow.Logger().Info().Msgf("OpenAI Response: %s", response)
	return agentflow.AgentResult{OutputState: state}, nil
}

func main() {
	agentflow.SetLogLevel(agentflow.INFO)
	// Fetch the OpenAI API key from the environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	// Create a new OpenAIAdapter instance
	adapter, err := llm.NewOpenAIAdapter(apiKey, "gpt-4o-mini", 100, 0.7)
	if err != nil {
		log.Fatalf("Failed to create OpenAIAdapter: %v", err)
	}

	agents := map[string]agentflow.AgentHandler{
		"openai": &OpenAIAgent{adapter: adapter},
	}

	runner := factory.NewRunnerWithConfig(factory.RunnerConfig{
		Agents: agents,
	})

	runner.Start(context.Background())
	defer runner.Stop()

	// Emit an event routed to the "openai" agent
	runner.Emit(agentflow.NewEvent(
		"question",
		map[string]interface{}{
			"user_prompt": "What is the capital of France?",
		},
		map[string]string{agentflow.RouteMetadataKey: "openai"},
	))

	// Give the agent time to run
	time.Sleep(500 * time.Millisecond)
}
