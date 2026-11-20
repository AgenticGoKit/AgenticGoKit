package main

import (
	"context"
	"fmt"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	"github.com/agenticgokit/agenticgokit/v1beta"
	"log"
	"time"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  Testing REAL LLM Integration - v1beta")
	fmt.Print("===========================================\n\n")

	// Create a simple agent configuration using an available model
	fmt.Println("Building agent with real LLM integration...")
	agent, err := v1beta.NewBuilder("test-agent").
		WithConfig(&v1beta.Config{
			Name:         "test-agent",
			SystemPrompt: "You are a helpful assistant. Keep your answers very short and concise.",
			Timeout:      30 * time.Second,
			LLM: v1beta.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b", // Using available model
				Temperature: 0.7,
				MaxTokens:   100,
				BaseURL:     "http://localhost:11434",
			},
		}).
		Build()

	if err != nil {
		log.Fatalf("❌ Failed to build agent: %v", err)
	}
	fmt.Println("✅ Agent built successfully!")

	// Initialize the agent
	ctx := context.Background()
	fmt.Println("\nInitializing agent...")
	if err := agent.Initialize(ctx); err != nil {
		log.Fatalf("❌ Failed to initialize agent: %v", err)
	}
	fmt.Println("✅ Agent initialized!")
	defer agent.Cleanup(ctx)

	// Check capabilities
	fmt.Println("\nAgent Capabilities:")
	capabilities := agent.Capabilities()
	for _, cap := range capabilities {
		fmt.Printf("  - %s\n", cap)
	}

	// Test with a simple query
	query := "What is 2+2? Answer in one sentence."
	fmt.Printf("\n🔍 Testing Query: %s\n", query)
	fmt.Println("---")

	queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	startTime := time.Now()
	result, err := agent.Run(queryCtx, query)
	duration := time.Since(startTime)

	if err != nil {
		log.Fatalf("❌ Failed to run agent: %v", err)
	}

	// Display results
	fmt.Println("\n✅ SUCCESS! Got real LLM response!")
	fmt.Println("===========================================")
	fmt.Printf("📝 Response:\n%s\n\n", result.Content)
	fmt.Printf("⏱️  Duration: %v\n", duration)
	fmt.Printf("🎯 Success: %v\n", result.Success)
	fmt.Printf("🔢 Tokens Used: %d\n", result.TokensUsed)
	fmt.Printf("💾 Memory Used: %v\n", result.MemoryUsed)

	if len(result.LLMInteractions) > 0 {
		llm := result.LLMInteractions[0]
		fmt.Printf("\n📊 LLM Interaction Details:\n")
		fmt.Printf("   Provider: %s\n", llm.Provider)
		fmt.Printf("   Model: %s\n", llm.Model)
		fmt.Printf("   Prompt Tokens: %d\n", llm.PromptTokens)
		fmt.Printf("   Response Tokens: %d\n", llm.ResponseTokens)
	}

	fmt.Println("\n===========================================")
	fmt.Println("  ✨ Real LLM Integration Test PASSED! ✨")
	fmt.Println("===========================================")
}
