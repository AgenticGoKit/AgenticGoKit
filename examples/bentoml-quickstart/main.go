// Package main demonstrates BentoML integration with AgenticGoKit.
// This example shows how to use BentoML as an LLM provider for AI agents.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/agenticgokit/agenticgokit/core"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/bentoml" // Import BentoML plugin
)

func main() {
	// Get BentoML configuration from environment variables
	baseURL := os.Getenv("BENTOML_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000" // Default BentoML URL
	}

	model := os.Getenv("BENTOML_MODEL")
	if model == "" {
		model = "llama2-7b-chat" // Default model
	}

	// Create BentoML provider
	provider, err := core.NewModelProviderFromConfig(core.LLMProviderConfig{
		Type:        "bentoml",
		BaseURL:     baseURL,
		Model:       model,
		MaxTokens:   2048,
		Temperature: 0.7,
		// BentoML-specific options (optional)
		BentoMLTopP: 0.9,
		BentoMLTopK: 50,
	})
	if err != nil {
		log.Fatalf("Failed to create BentoML provider: %v", err)
	}

	ctx := context.Background()

	// Example 1: Simple completion
	fmt.Println("=== Example 1: Simple Completion ===")
	resp, err := provider.Call(ctx, core.Prompt{
		System: "You are a helpful assistant that provides concise answers.",
		User:   "What is BentoML and why is it useful for ML deployment?",
	})
	if err != nil {
		log.Fatalf("Call failed: %v", err)
	}
	fmt.Printf("Response: %s\n", resp.Content)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n\n",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)

	// Example 2: Streaming response
	fmt.Println("=== Example 2: Streaming Response ===")
	tokenChan, err := provider.Stream(ctx, core.Prompt{
		System: "You are a creative writer.",
		User:   "Write a haiku about machine learning.",
	})
	if err != nil {
		log.Fatalf("Stream failed: %v", err)
	}

	fmt.Print("Streaming: ")
	for token := range tokenChan {
		if token.Error != nil {
			log.Fatalf("Stream error: %v", token.Error)
		}
		fmt.Print(token.Content)
	}
	fmt.Println("\n")

	// Example 3: Code generation
	fmt.Println("=== Example 3: Code Generation ===")
	codeResp, err := provider.Call(ctx, core.Prompt{
		System: "You are an expert Python programmer. Provide clean, idiomatic Python code.",
		User:   "Write a function to calculate the Fibonacci sequence.",
	})
	if err != nil {
		log.Fatalf("Code generation failed: %v", err)
	}
	fmt.Printf("Generated Code:\n%s\n", codeResp.Content)

	fmt.Println("\n=== BentoML Quick Start Complete ===")
}
