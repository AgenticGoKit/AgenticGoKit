package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  Ollama QuickStart Agent - vNext API")
	fmt.Println("===========================================\n")

	// Initialize vNext with defaults (optional but recommended)
	if err := vnext.InitializeDefaults(); err != nil {
		log.Fatalf("Failed to initialize vNext: %v", err)
	}

	// Quick way to create a chat agent with custom configuration
	config := &vnext.Config{
		Name:         "quick-helper",
		SystemPrompt: "You are a helpful assistant that provides short, concise answers in 2-3 sentences.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gpt-oss:20b-cloud",
			Temperature: 0.3,
			MaxTokens:   200,
			BaseURL:     "http://localhost:11434",
		},
	}

	// Create agent using QuickStart API
	agent, err := vnext.QuickChatAgentWithConfig("llama3.2", config)
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
