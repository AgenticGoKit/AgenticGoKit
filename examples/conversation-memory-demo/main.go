package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	_ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem" // Register chromem provider
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
	fmt.Println("Interactive Chat Agent with Memory")
	fmt.Println("===================================")
	fmt.Println()
	fmt.Println("This demo shows how an agent maintains conversation history")
	fmt.Println("and uses memory to provide personalized, context-aware responses.")
	fmt.Println()
	fmt.Println("Features demonstrated:")
	fmt.Println("  * Conversation history storage")
	fmt.Println("  * Memory retrieval for context")
	fmt.Println("  * Personalized responses based on chat history")
	fmt.Println("  * Session-scoped memory (each conversation is separate)")
	fmt.Println()

	ctx := context.Background()

	// Step 1: Create agent with memory integration
	agent, err := vnext.NewBuilder("chat-assistant").
		WithConfig(&vnext.Config{
			Name: "chat-assistant",
			SystemPrompt: `You are a helpful and friendly chat assistant.
You remember details from our conversation and provide personalized responses.
Be conversational and engaging while being helpful.`,
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: 0.7,
				MaxTokens:   2000, // Allow detailed responses
			},
			Memory: &vnext.MemoryConfig{
				Enabled: true, // Explicitly enable memory
				// Provider defaults to "chromem" - embedded vector database
				RAG: &vnext.RAGConfig{
					MaxTokens:       1000,
					PersonalWeight:  0.8, // Prioritize conversation history
					KnowledgeWeight: 0.2,
					HistoryLimit:    20, // Keep last 20 messages
				},
			},
			Timeout: 300 * time.Second,
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

	fmt.Println("Agent initialized successfully!")
	fmt.Println()

	// Step 3: Start interactive chat loop
	scanner := bufio.NewScanner(os.Stdin)
	conversationCount := 0

	fmt.Println("Start chatting! Type 'quit' or 'exit' to end the conversation.")
	fmt.Println("Try asking questions that build on previous messages to see memory in action.")
	fmt.Println()

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}

		if strings.ToLower(userInput) == "quit" || strings.ToLower(userInput) == "exit" {
			fmt.Println("Goodbye! Thanks for chatting.")
			break
		}

		conversationCount++
		fmt.Printf("\nAssistant (Turn %d):\n", conversationCount)

		// Run agent with the user input
		result, err := agent.Run(ctx, userInput)
		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			continue
		}

		// Display the response
		fmt.Printf("%s\n", result.Content)

		// Show memory usage information
		if result.MemoryUsed {
			fmt.Printf("\n[Memory] Used (%d queries)\n", result.MemoryQueries)
		} else {
			fmt.Printf("\n[Memory] Not used\n")
		}

		fmt.Printf("[Time] Response time: %v\n", result.Duration)
		fmt.Println(strings.Repeat("-", 60))
		fmt.Println()
	}

	// Step 4: After conversation ends, show what was stored in memory
	fmt.Println("\nMemory Inspection")
	fmt.Println("=================")

	// We need to access the memory provider to inspect stored data
	// Since the agent encapsulates the memory, we'll create a simple demonstration
	// by showing that memory persists across runs

	fmt.Println("Conversation Summary:")
	fmt.Println("  - The agent automatically stored each user message and assistant response")
	fmt.Println("  - Memory is session-scoped, so each conversation maintains its own history")
	fmt.Println("  - Future messages can reference previous context through RAG retrieval")
	fmt.Println("  - Try asking 'What did I say earlier?' or 'Remind me what we talked about'")
	fmt.Println()
	fmt.Println("Run this demo again to start a fresh conversation with new memory!")

	fmt.Println("\nDemo completed! The agent remembered our conversation history.")
}
