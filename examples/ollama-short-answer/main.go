package main

import (
	"context"
	"fmt"
	"log"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  Ollama Short Answer Agent - vNext API")
	fmt.Println("===========================================\n")

	// Create a simple chat agent using Ollama with short, concise responses
	agent, err := createShortAnswerAgent()
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Initialize the agent
	ctx := context.Background()
	if err := agent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	defer agent.Cleanup(ctx)

	// Run example queries
	queries := []string{
		"What is 2+29?",
		"Explain what Docker is.",
	}

	for i, query := range queries {
		fmt.Printf("[Query %d] %s\n", i+1, query)
		fmt.Println("---")

		// Create context with timeout
		queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		// Run the agent
		result, err := agent.Run(queryCtx, query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			cancel()
			continue
		}

		// Display the result
		fmt.Printf("✓ Answer: %s\n", result.Content)
		fmt.Printf("   Duration: %v\n", result.Duration)
		fmt.Printf("   Success: %v\n\n", result.Success)

		cancel()
	}

	fmt.Println("===========================================")
	fmt.Println("  Demo completed successfully!")
	fmt.Println("===========================================")
}

// createShortAnswerAgent creates an Ollama-based agent that provides short, concise answers
func createShortAnswerAgent() (vnext.Agent, error) {
	// Define the system prompt for short answers
	systemPrompt := `You are a helpful AI assistant that provides short, concise answers.
Keep your responses to 2-3 sentences maximum.
Be direct and to the point.
Do not provide long explanations unless specifically asked.`

	// Create agent configuration
	config := &vnext.Config{
		Name:         "short-answer-agent",
		SystemPrompt: systemPrompt,
		Timeout:      30 * time.Second,
		DebugMode:    false,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",              // Using Llama 3.2 model
			Temperature: 0.3,                      // Lower temperature for more focused answers
			MaxTokens:   200,                      // Limit tokens to keep answers short
			BaseURL:     "http://localhost:11434", // Default Ollama URL
		},
	}

	// Build the agent using the Builder pattern with ChatAgent preset
	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		WithPreset(vnext.ChatAgent).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build agent: %w", err)
	}

	return agent, nil
}



