package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	_ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem" // Register chromem provider
	v1beta "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
	fmt.Println("🤖 Personal Assistant with Memory (Conversation History)")
	fmt.Println("=========================================================\n")

	ctx := context.Background()

	// Step 1: Create agent with memory integration enabled
	// The agent will automatically manage conversation history
	agent, err := v1beta.NewBuilder("personal-assistant").
		WithConfig(&v1beta.Config{
			Name: "personal-assistant",
			SystemPrompt: `You are a helpful personal assistant. 
Remember information from our conversation and provide personalized responses.`,
			LLM: v1beta.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: 0.7,
				MaxTokens:   80, // Short responses for faster demo
			},
			Memory: &v1beta.MemoryConfig{
				Enabled: true,
				// Provider defaults to "chromem" - embedded vector database
				RAG: &v1beta.RAGConfig{
					MaxTokens:       500,
					PersonalWeight:  0.6,
					KnowledgeWeight: 0.4,
					HistoryLimit:    10,
				},
			},
			Timeout: 90 * time.Second, // Generous timeout for Ollama
		}).
		Build()

	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Step 2: Initialize agent
	if err := agent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	defer agent.Cleanup(ctx)

	// Step 3: Demo conversations showing memory working
	conversations := []struct {
		input string
		note  string // what this conversation demonstrates
	}{
		{
			input: "My name is Kunal and I'm a Go developer.",
			note:  "Storing user information in memory",
		},
		{
			input: "I work on microservices using Docker and Kubernetes.",
			note:  "Adding more context to memory",
		},
		{
			input: "What is my name?",
			note:  "Testing memory recall - should remember 'Kunal'",
		},
		{
			input: "What kind of developer am I?",
			note:  "Should recall Go developer",
		},
		{
			input: "What tools do I use?",
			note:  "Should remember Docker and Kubernetes",
		},
	}

	fmt.Println("💬 Demonstration: Memory-Powered Personalized Responses\n")
	fmt.Println("======================================================================")
	fmt.Println()

	for i, conv := range conversations {
		fmt.Printf("👤 User [%d]: %s\n", i+1, conv.input)
		fmt.Printf("   💡 %s\n\n", conv.note)

		result, err := agent.Run(ctx, conv.input)
		if err != nil {
			log.Printf("❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("🤖 Assistant: %s\n", result.Content)

		// Show memory usage
		if result.MemoryUsed {
			fmt.Printf("   � Memory: Used (queries=%d)\n", result.MemoryQueries)
		} else {
			fmt.Printf("   ⚠️  Memory: Not used\n")
		}

		fmt.Printf("   ⏱️  Duration: %v\n", result.Duration)
		fmt.Println()
		fmt.Println("----------------------------------------------------------------------")
		fmt.Println()

		// Small delay between requests
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("======================================================================")
	fmt.Println("✅ Demo Complete!")
	fmt.Println("\n📊 Key Features Demonstrated:")
	fmt.Println("   ✅ Memory persistence (user profile & preferences)")
	fmt.Println("   ✅ Context-aware responses (RAG retrieves relevant context)")
	fmt.Println("   ✅ Personalization (responses tailored to user)")
	fmt.Println("   ✅ Memory integration with LLM (enriched prompts)")
	fmt.Println("\n💡 Try modifying the stored preferences and see how responses change!")
}
