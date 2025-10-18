package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/llm/ollama"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  Ollama TOML Config Agent - vNext API")
	fmt.Println("===========================================\n")

	// Check if config file exists
	configPath := "config.toml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// Load configuration from TOML file
	fmt.Printf("Loading configuration from: %s\n", configPath)
	config, err := vnext.LoadConfigFromTOML(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("✓ Configuration loaded successfully\n")
	fmt.Printf("  Agent Name: %s\n", config.Name)
	fmt.Printf("  LLM Provider: %s\n", config.LLM.Provider)
	fmt.Printf("  LLM Model: %s\n", config.LLM.Model)
	fmt.Printf("  Max Tokens: %d\n\n", config.LLM.MaxTokens)

	// Build agent from configuration (no preset needed - using config directly)
	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		Build()

	if err != nil {
		log.Fatalf("Failed to build agent: %v", err)
	}

	// Initialize
	ctx := context.Background()
	if err := agent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	defer agent.Cleanup(ctx)

	// Example queries
	queries := []string{
		"What is GraphQL?",
		"Explain what Terraform does.",
	}

	for i, query := range queries {
		fmt.Printf("\n[Query %d] %s\n", i+1, query)
		fmt.Println("---")

		queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		result, err := agent.Run(queryCtx, query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			cancel()
			continue
		}

		fmt.Printf("✓ Answer: %s\n", result.Content)
		fmt.Printf("   Duration: %v\n\n", result.Duration)

		cancel()
	}

	fmt.Println("===========================================")
	fmt.Println("  TOML Config demo completed!")
	fmt.Println("===========================================")
}
