package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

func main() {
	fmt.Println("🚀 AgenticGoKit vNext Streaming Demo")
	fmt.Println("=====================================")
	fmt.Println()

	// Show menu
	showMenu()

	// Get user choice
	var choice string
	fmt.Print("Enter your choice (1-4): ")
	fmt.Scanln(&choice)
	fmt.Println()

	switch choice {
	case "1":
		demoBasicStreaming()
	case "2":
		demoStreamingWithOptions()
	case "3":
		demoMultipleProvidersStreaming()
	case "4":
		demoInteractiveStreaming()
	default:
		fmt.Println("❌ Invalid choice. Running basic streaming demo...")
		demoBasicStreaming()
	}
}

func showMenu() {
	fmt.Println("Choose a streaming demo:")
	fmt.Println("1. Basic Streaming - See tokens arrive in real-time")
	fmt.Println("2. Streaming with Options - Advanced streaming configuration")
	fmt.Println("3. Multiple Providers - Compare Ollama, OpenAI, Azure streaming")
	fmt.Println("4. Interactive Streaming - Real-time conversation")
	fmt.Println()
}

// Demo 1: Basic streaming demonstration
func demoBasicStreaming() {
	fmt.Println("🔥 Demo 1: Basic Streaming")
	fmt.Println("=========================")
	fmt.Println("This demo shows how streaming works with real-time token delivery.")
	fmt.Println()

	// Create agent with Ollama (most accessible for local testing)
	agent, err := createOllamaAgent()
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start streaming
	prompt := "Write a short story about a robot learning to paint. Keep it under 200 words."
	fmt.Printf("🎨 Prompt: %s\n\n", prompt)
	fmt.Println("📡 Streaming response:")
	fmt.Println("─────────────────────")

	stream, err := agent.RunStream(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to start streaming: %v", err)
	}

	// Process stream chunks in real-time
	var fullResponse string
	chunkCount := 0
	startTime := time.Now()

	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			fmt.Printf("\n❌ Stream error: %v\n", chunk.Error)
			break
		}

		switch chunk.Type {
		case vnext.ChunkTypeDelta:
			// Print each token as it arrives
			fmt.Print(chunk.Delta)
			fullResponse += chunk.Delta
			chunkCount++
		case vnext.ChunkTypeDone:
			fmt.Println("\n\n✅ Stream completed!")
		case vnext.ChunkTypeMetadata:
			fmt.Printf("\n📊 Metadata: %v\n", chunk.Metadata)
		}
	}

	// Show streaming statistics
	duration := time.Since(startTime)
	fmt.Println("\n📊 Streaming Statistics:")
	fmt.Println("─────────────────────")
	fmt.Printf("• Total chunks: %d\n", chunkCount)
	fmt.Printf("• Duration: %v\n", duration)
	fmt.Printf("• Characters: %d\n", len(fullResponse))
	if chunkCount > 0 {
		fmt.Printf("• Avg chunk size: %.1f chars\n", float64(len(fullResponse))/float64(chunkCount))
		fmt.Printf("• Tokens per second: %.1f\n", float64(chunkCount)/duration.Seconds())
	}
}

// Demo 2: Streaming with advanced options
func demoStreamingWithOptions() {
	fmt.Println("⚙️ Demo 2: Streaming with Options")
	fmt.Println("=================================")
	fmt.Println("This demo shows advanced streaming configuration and options.")
	fmt.Println()

	agent, err := createOllamaAgent()
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Create run options
	// runOpts := &vnext.RunOptions{
	// 	Timeout: 30 * time.Second,
	// }

	// Create streaming options as individual options
	// streamOpts := []vnext.StreamOption{
	// 	vnext.WithBufferSize(100),
	// 	vnext.WithThoughts(),
	// 	vnext.WithToolCalls(),
	// }

	prompt := "Explain quantum computing in simple terms. Think step by step about how to explain this complex topic."
	fmt.Printf("🔬 Prompt: %s\n\n", prompt)
	fmt.Println("📡 Streaming with advanced options:")
	fmt.Println("─────────────────────────────────")

	// Note: Using basic RunStream since RunStreamWithOptions needs investigation
	fmt.Println("📡 Streaming with advanced options (using basic streaming for now):")
	fmt.Println("─────────────────────────────────────────────────────────────")

	stream, err := agent.RunStream(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to start streaming: %v", err)
	}

	// Process different chunk types
	var textContent, thoughts string
	chunkCounts := make(map[vnext.ChunkType]int)

	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			fmt.Printf("\n❌ Stream error: %v\n", chunk.Error)
			break
		}

		chunkCounts[chunk.Type]++

		switch chunk.Type {
		case vnext.ChunkTypeDelta:
			fmt.Print(chunk.Delta)
			textContent += chunk.Delta
		case vnext.ChunkTypeThought:
			fmt.Printf("\n💭 Thought: %s\n", chunk.Delta)
			thoughts += chunk.Delta
		case vnext.ChunkTypeToolCall:
			fmt.Printf("\n🔧 Tool Call: %s\n", chunk.Delta)
		case vnext.ChunkTypeMetadata:
			fmt.Printf("\n📊 Metadata: %v\n", chunk.Metadata)
		case vnext.ChunkTypeDone:
			fmt.Println("\n\n✅ Stream completed!")
		}
	} // Show chunk type statistics
	fmt.Println("\n📊 Chunk Type Statistics:")
	fmt.Println("─────────────────────────")
	for chunkType, count := range chunkCounts {
		fmt.Printf("• %s: %d chunks\n", chunkType, count)
	}
}

// Demo 3: Multiple providers streaming comparison
func demoMultipleProvidersStreaming() {
	fmt.Println("🌐 Demo 3: Multiple Providers Streaming")
	fmt.Println("=======================================")
	fmt.Println("This demo compares streaming across different LLM providers.")
	fmt.Println("Note: OpenAI and Azure require valid API keys in environment variables.")
	fmt.Println()

	prompt := "List 5 benefits of renewable energy in bullet points."
	fmt.Printf("⚡ Prompt: %s\n\n", prompt)

	// Try Ollama (most accessible)
	if agent, err := createOllamaAgent(); err == nil {
		fmt.Println("🦙 Ollama Streaming:")
		fmt.Println("──────────────────")
		streamWithProvider(agent, prompt, "Ollama")
		fmt.Println()
	} else {
		fmt.Printf("⚠️ Ollama not available: %v\n\n", err)
	}

	// Try OpenAI (if API key available)
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		if agent, err := createOpenAIAgent(); err == nil {
			fmt.Println("🤖 OpenAI Streaming:")
			fmt.Println("───────────────────")
			streamWithProvider(agent, prompt, "OpenAI")
			fmt.Println()
		} else {
			fmt.Printf("⚠️ OpenAI not available: %v\n\n", err)
		}
	} else {
		fmt.Println("⚠️ OpenAI not available: OPENAI_API_KEY not set\n")
	}

	// Try Azure OpenAI (if API key available)
	if azureKey := os.Getenv("AZURE_OPENAI_API_KEY"); azureKey != "" {
		if agent, err := createAzureAgent(); err == nil {
			fmt.Println("☁️ Azure OpenAI Streaming:")
			fmt.Println("─────────────────────────")
			streamWithProvider(agent, prompt, "Azure")
			fmt.Println()
		} else {
			fmt.Printf("⚠️ Azure not available: %v\n\n", err)
		}
	} else {
		fmt.Println("⚠️ Azure not available: AZURE_OPENAI_API_KEY not set\n")
	}
}

// Demo 4: Interactive streaming conversation
func demoInteractiveStreaming() {
	fmt.Println("💬 Demo 4: Interactive Streaming")
	fmt.Println("================================")
	fmt.Println("This demo shows interactive streaming conversation.")
	fmt.Println("Type 'quit' to exit, 'clear' to clear screen.")
	fmt.Println("Tip: You can ask full questions with multiple words!")
	fmt.Println()

	agent, err := createOllamaAgent()
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create a scanner to read full lines including spaces
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Get user input
		fmt.Print("🧑 You: ")

		// Read the full line including spaces
		if !scanner.Scan() {
			break // EOF or error
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "quit" {
			fmt.Println("👋 Goodbye!")
			break
		}

		if input == "clear" {
			fmt.Print("\033[H\033[2J") // Clear screen
			continue
		}

		if input == "" {
			continue
		}

		// Start streaming response
		fmt.Print("🤖 Agent: ")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		stream, err := agent.RunStream(ctx, input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			cancel()
			continue
		}

		// Stream response in real-time
		for chunk := range stream.Chunks() {
			if chunk.Error != nil {
				fmt.Printf("\n❌ Error: %v\n", chunk.Error)
				break
			}

			if chunk.Type == vnext.ChunkTypeDelta {
				fmt.Print(chunk.Delta)
			} else if chunk.Type == vnext.ChunkTypeDone {
				fmt.Println("\n")
				break
			}
		}

		cancel()
	}
} // Helper function to stream with any provider
func streamWithProvider(agent vnext.Agent, prompt, providerName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	startTime := time.Now()
	stream, err := agent.RunStream(ctx, prompt)
	if err != nil {
		fmt.Printf("❌ Error starting stream: %v\n", err)
		return
	}

	var response string
	chunkCount := 0

	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			fmt.Printf("\n❌ Stream error: %v\n", chunk.Error)
			break
		}

		if chunk.Type == vnext.ChunkTypeDelta {
			fmt.Print(chunk.Delta)
			response += chunk.Delta
			chunkCount++
		} else if chunk.Type == vnext.ChunkTypeDone {
			break
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\n📊 %s: %d chunks in %v (%.1f chunks/sec)\n",
		providerName, chunkCount, duration, float64(chunkCount)/duration.Seconds())
}

// Agent creation helpers
func createOllamaAgent() (vnext.Agent, error) {
	config := &vnext.Config{
		Name:         "streaming-demo-ollama",
		SystemPrompt: "You are a helpful assistant. Provide clear, concise responses.",
		Timeout:      60 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b", // Fast small model for demo
			Temperature: 0.7,
			MaxTokens:   500,
			BaseURL:     "http://localhost:11434",
		},
	}

	return vnext.NewBuilder("streaming-demo").
		WithConfig(config).
		WithPreset(vnext.ChatAgent).
		Build()
}

func createOpenAIAgent() (vnext.Agent, error) {
	config := &vnext.Config{
		Name:         "streaming-demo-openai",
		SystemPrompt: "You are a helpful assistant. Provide clear, concise responses.",
		Timeout:      60 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o-mini",
			Temperature: 0.7,
			MaxTokens:   500,
			APIKey:      os.Getenv("OPENAI_API_KEY"),
		},
	}

	return vnext.NewBuilder("streaming-demo").
		WithConfig(config).
		WithPreset(vnext.ChatAgent).
		Build()
}

func createAzureAgent() (vnext.Agent, error) {
	config := &vnext.Config{
		Name:         "streaming-demo-azure",
		SystemPrompt: "You are a helpful assistant. Provide clear, concise responses.",
		Timeout:      60 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "azure",
			Model:       "gpt-4o-mini",
			Temperature: 0.7,
			MaxTokens:   500,
			APIKey:      os.Getenv("AZURE_OPENAI_API_KEY"),
			BaseURL:     os.Getenv("AZURE_OPENAI_ENDPOINT"),
		},
	}

	return vnext.NewBuilder("streaming-demo").
		WithConfig(config).
		WithPreset(vnext.ChatAgent).
		Build()
}
