package main

import (
	"context"
	"errors"
	"log"
	"time"

	agentflow "github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/llm"
)

// OllamaAgent implements agentflow.AgentHandler
type OllamaAgent struct {
	adapter *llm.OllamaAdapter
}

func (a *OllamaAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	systemPrompt := "You are a helpful assistant."

	// Fetch the user prompt from the event payload
	userPrompt, ok := event.GetData()["user_prompt"].(string)
	if !ok || userPrompt == "" {
		return agentflow.AgentResult{}, errors.New("user_prompt is missing or invalid in the event payload")
	}

	prompt := llm.Prompt{
		System: systemPrompt,
		User:   userPrompt,
	}

	response, err := a.adapter.Call(ctx, prompt)
	if err != nil {
		return agentflow.AgentResult{}, err
	}

	agentflow.Logger().Info().Msgf("Ollama Response: %s", response.Content)
	return agentflow.AgentResult{OutputState: state}, nil
}

// Refactor to use agentflow with OllamaAdapter
func main() {
	agentflow.SetLogLevel(agentflow.INFO)

	adapter, err := llm.NewOllamaAdapter("test-key", "gemma3:latest", 100, 0.7)
	if err != nil {
		agentflow.Logger().Error().Msgf("Failed to create OllamaAdapter: %v", err)
		return
	}

	agent := &OllamaAgent{adapter: adapter}
	event := &agentflow.SimpleEvent{
		Data: agentflow.EventData{"user_prompt": "What is the capital of France?"},
	}
	state := agentflow.NewState()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := agent.Run(ctx, event, state)
	if err != nil {
		log.Fatalf("Agent execution failed: %v", err)
	}
	agentflow.Logger().Debug().Msgf("Agent execution succeeded: %+v", result)
}
