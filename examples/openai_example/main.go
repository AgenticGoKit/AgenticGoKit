package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// OpenAIAgent implements agentflow.AgentHandler
// Uses the public core.ModelProvider interface
// No internal/llm references

type OpenAIAgent struct {
	provider agentflow.ModelProvider
}

func (a *OpenAIAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	systemPrompt := "You are a helpful assistant."
	userPrompt, ok := event.GetData()["user_prompt"].(string)
	if !ok || userPrompt == "" {
		return agentflow.AgentResult{}, errors.New("user_prompt is missing or invalid in the event payload")
	}
	prompt := agentflow.Prompt{
		System: systemPrompt,
		User:   userPrompt,
		Parameters: agentflow.ModelParameters{
			Temperature: agentflow.FloatPtr(0.7),
			MaxTokens:   agentflow.Int32Ptr(100),
		},
	}
	response, err := a.provider.Call(ctx, prompt)
	if err != nil {
		return agentflow.AgentResult{}, err
	}
	agentflow.Logger().Info().Msgf("OpenAI Response: %s", response.Content)
	outputState := state.Clone()
	outputState.Set("llm_response", response.Content)
	return agentflow.AgentResult{OutputState: outputState}, nil
}

func main() {
	agentflow.SetLogLevel(agentflow.INFO)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}
	provider, err := agentflow.NewOpenAIAdapter(apiKey, "gpt-4o-mini", 100, 0.7)
	if err != nil {
		log.Fatalf("Failed to create OpenAIAdapter: %v", err)
	}

	agents := map[string]agentflow.AgentHandler{
		"openai": &OpenAIAgent{provider: provider},
	}
	runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
		Agents: agents,
	})

	runner.Start(context.Background())
	defer runner.Stop()

	runner.Emit(agentflow.NewEvent(
		"question",
		map[string]interface{}{
			"user_prompt": "What is the capital of France?",
		},
		map[string]string{agentflow.RouteMetadataKey: "openai"},
	))

	time.Sleep(500 * time.Millisecond)
}
