package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	ia "github.com/kunalkushwaha/agenticgokit/internal/agents"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/logging/zerolog"
)

// Minimal smoke test to verify LLM flow with Ollama using the UnifiedAgent builder
func main() {
	core.SetLogLevel(core.INFO)
	logger := core.Logger()

	baseURL := getenv("OLLAMA_BASE_URL", "http://localhost:11434")
	model := getenv("OLLAMA_MODEL", "gemma3:latest")

	// Create provider directly (bypasses plugin registry)
	provider, err := core.NewOllamaAdapter(baseURL, model, 1024, 0.7)
	if err != nil {
		fmt.Printf("Failed to create Ollama provider: %v\n", err)
		os.Exit(1)
	}
	logger.Info().Str("base_url", baseURL).Str("model", model).Msg("Ollama provider ready")

	// Build a UnifiedAgent with LLM capability
	agent, err := ia.NewAgent("ollama_smoke").
		WithLLMAndConfig(provider, core.LLMConfig{Model: model, Temperature: 0.7, MaxTokens: 512, TimeoutSeconds: 30}).
		WithDefaultMetrics().
		Build()
	if err != nil {
		fmt.Printf("Failed to build agent: %v\n", err)
		os.Exit(1)
	}

	// Run with an input message
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	state := core.NewState(map[string]interface{}{
		"message":       "Say a short hello from Ollama.",
		"system_prompt": "You are a concise assistant.",
	})

	out, err := agent.Run(ctx, state)
	if err != nil {
		fmt.Printf("Agent run error: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	if v, ok := out.Get("response"); ok {
		if s, ok := v.(string); ok {
			fmt.Printf("LLM Response: %s\n", s)
			return
		}
	}
	if v, ok := out.Get("message"); ok {
		if s, ok := v.(string); ok {
			fmt.Printf("Message: %s\n", s)
			return
		}
	}
	fmt.Println("No response found in state; keys:", out.Keys())
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
