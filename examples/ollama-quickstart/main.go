package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	v1beta "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  Ollama QuickStart Agent - v1beta API")
	fmt.Println("===========================================\n")

	// Initialize v1beta with defaults (optional but recommended)
	if err := v1beta.InitializeDefaults(); err != nil {
		log.Fatalf("Failed to initialize v1beta: %v", err)
	}

	// Quick way to create a chat agent with custom configuration
	config := &v1beta.Config{
		Name:         "quick-helper",
		SystemPrompt: "You are a helpful assistant that provides short, concise answers in 2-3 sentences.",
		Timeout:      30 * time.Second,
		LLM: v1beta.LLMConfig{
			Provider:    "ollama",
			Model:       "gpt-oss:20b-cloud",
			Temperature: 0.3,
			MaxTokens:   200,
			BaseURL:     "http://localhost:11434",
		},
	}

	// Create agent using streamlined Builder API
	agent, err := v1beta.NewBuilder(config.Name).
		WithConfig(config).
		WithLLM("ollama", "gpt-oss:20b-cloud").
		Build()
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Initialize
	ctx := context.Background()
	if err := agent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	defer agent.Cleanup(ctx)

	// Interactive loop
	queries := []string{
		"What is REST API?",
		"Explain CI/CD in simple terms.",
		"What is the difference between HTTP and HTTPS?",
	}

	for i, query := range queries {
		fmt.Printf("\n[Question %d] %s\n", i+1, query)

		queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		// Using simple Run method
		result, err := agent.Run(queryCtx, query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			cancel()
			continue
		}

		fmt.Printf("\n📝 Answer:\n%s\n", result.Content)
		fmt.Printf("\n⏱️  Duration: %v | Success: %v\n", result.Duration, result.Success)

		cancel()
	}

	fmt.Println("\n===========================================")
	fmt.Println("  QuickStart demo completed!")
	fmt.Println("===========================================")
}
