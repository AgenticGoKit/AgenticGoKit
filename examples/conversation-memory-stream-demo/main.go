package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	_ "github.com/agenticgokit/agenticgokit/plugins/memory/memory" // Register memory provider
)

func main() {
	fmt.Println("Interactive Chat Agent with Memory (Streaming)")
	fmt.Println("===============================================")
	fmt.Println()
	fmt.Println("This demo shows how an agent maintains conversation history")
	fmt.Println("with REAL-TIME STREAMING responses for a more interactive experience.")
	fmt.Println()
	fmt.Println("Features demonstrated:")
	fmt.Println("  * Conversation history storage")
	fmt.Println("  * Memory retrieval for context")
	fmt.Println("  * Real-time streaming responses (token-by-token)")
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
				Model:       "qwen3:0.6b",
				Temperature: 0.7,
				MaxTokens:   2000, // Allow detailed responses
			},
			Memory: &vnext.MemoryConfig{
				Provider: "memory", // Use in-memory provider
				RAG: &vnext.RAGConfig{
					MaxTokens:       1000,
					PersonalWeight:  0.8, // Prioritize conversation history
					KnowledgeWeight: 0.2,
					HistoryLimit:    20, // Keep last 20 messages
				},
			},
			Streaming: &vnext.StreamingConfig{
				Enabled:       true,
				BufferSize:    100,
				FlushInterval: 50, // 50ms flush interval for smooth streaming
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

	// Step 3: Start interactive chat loop with streaming
	scanner := bufio.NewScanner(os.Stdin)
	conversationCount := 0

	fmt.Println("Start chatting! Type 'quit' or 'exit' to end the conversation.")
	fmt.Println("Watch the responses stream in real-time, word by word!")
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
		fmt.Printf("\nAssistant (Turn %d): ", conversationCount)

		// Run agent with streaming
		stream, err := agent.RunStream(ctx, userInput)

		if err != nil {
			fmt.Printf("\nError: %v\n\n", err)
			continue
		}

		// Process streaming chunks
		var fullResponse strings.Builder

		for chunk := range stream.Chunks() {
			switch chunk.Type {
			case vnext.ChunkTypeDelta:
				// Stream the actual response text
				fmt.Print(chunk.Delta)
				fullResponse.WriteString(chunk.Delta)

			case vnext.ChunkTypeError:
				// Handle streaming errors
				fmt.Printf("\nStream error: %v\n", chunk.Error)

			case vnext.ChunkTypeDone:
				// Streaming completed
				fmt.Println() // New line after response
			}
		}

		// Get final result with metadata
		result, err := stream.Wait()
		if err != nil {
			fmt.Printf("\nError getting result: %v\n\n", err)
			continue
		}
		if result == nil {
			fmt.Printf("\nNo result available\n\n")
			continue
		}

		// Show memory usage information
		fmt.Println()
		if result.MemoryUsed {
			fmt.Printf("[Memory] Used (%d queries)\n", result.MemoryQueries)
		} else {
			fmt.Printf("[Memory] Not used\n")
		}

		fmt.Printf("[Time] Response time: %v\n", result.Duration)
		if result.TokensUsed > 0 {
			fmt.Printf("[Tokens] Used: %d\n", result.TokensUsed)
		}
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println()
	}

	// Step 4: After conversation ends, show what was stored in memory
	fmt.Println("\nMemory Inspection")
	fmt.Println("====================")

	fmt.Println("Conversation Summary:")
	fmt.Println("  - The agent automatically stored each user message and assistant response")
	fmt.Println("  - Memory is session-scoped, so each conversation maintains its own history")
	fmt.Println("  - Responses were streamed in real-time for better UX")
	fmt.Println("  - Future messages can reference previous context through RAG retrieval")
	fmt.Println("  - Try asking 'What did I say earlier?' or 'Remind me what we talked about'")
	fmt.Println()
	fmt.Println("Run this demo again to start a fresh conversation with new memory!")

	fmt.Println("\n✅ Demo completed! The agent streamed responses while remembering conversation history.")
}



