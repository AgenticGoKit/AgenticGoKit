package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/huggingface"
)

// Simple HuggingFace vnext example - minimal code to get started
func main() {
	// Get API key
	apiKey := os.Getenv("HUGGINGFACE_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set HUGGINGFACE_API_KEY environment variable")
	}

	// Initialize vNext
	if err := vnext.InitializeDefaults(); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Create agent config using new router
	config := &vnext.Config{
		Name:         "hf-simple",
		SystemPrompt: "You are a helpful assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "huggingface",
			Model:       "meta-llama/Llama-3.2-1B-Instruct", // New router-compatible model
			APIKey:      apiKey,
			Temperature: 0.7,
			MaxTokens:   200,
		},
	}

	// Build agent
	agent, err := vnext.NewBuilder("hf-simple").
		WithConfig(config).
		Build()

	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Initialize and cleanup
	ctx := context.Background()
	if err := agent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	defer agent.Cleanup(ctx)

	// Ask a question
	fmt.Println("Question: What is artificial intelligence?")
	fmt.Println()

	result, err := agent.Run(ctx, "What is artificial intelligence? Answer in 2-3 sentences.")
	if err != nil {
		log.Fatalf("Run failed: %v", err)
	}

	fmt.Println("Answer:")
	fmt.Println(result.Content)
	fmt.Println()
	fmt.Printf("Duration: %v | Tokens: %d\n", result.Duration, result.TokensUsed)
}
