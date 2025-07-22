package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

func main() {
	// 1. Get the user's question from command line
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go \"Your question here\"")
		fmt.Println("Example: go run main.go \"What is the capital of France?\"")
		os.Exit(1)
	}
	question := os.Args[1]

	// 2. Check for API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ OPENAI_API_KEY environment variable not set")
		fmt.Println("💡 Get your API key from: https://platform.openai.com/api-keys")
		fmt.Println("💡 Then run: export OPENAI_API_KEY=your-key-here")
		os.Exit(1)
	}

	// 3. Create an LLM provider (OpenAI in this case)
	provider := core.OpenAIProvider{
		APIKey:      apiKey,
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7, // Balanced creativity
		MaxTokens:   500,  // Reasonable response length
	}

	// 4. Create a simple agent with a helpful personality
	agent := core.NewLLMAgent("helpful-assistant", provider).
		WithSystemPrompt(`You are a helpful, knowledgeable assistant. 
		Provide clear, accurate, and concise answers to questions. 
		If you're not sure about something, say so honestly.
		Keep responses informative but not overly long.`)

	// 5. Process the question
	fmt.Printf("🤖 Simple Agent Starting...\n")
	fmt.Printf("📝 Processing: \"%s\"\n\n", question)

	startTime := time.Now()

	// Create input state with the user's question
	inputState := core.NewState().Set("message", question)

	// Run the agent with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := agent.Run(ctx, inputState)

	duration := time.Since(startTime)

	// 6. Handle the response
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("❌ Request timed out (30s limit)")
			fmt.Println("💡 Try a simpler question or check your internet connection")
		} else {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("💡 Check your API key and internet connection")
		}
		os.Exit(1)
	}

	// Extract the response from the result state
	response, ok := result.Get("response")
	if !ok {
		fmt.Println("❌ No response received from agent")
		fmt.Println("💡 This might be a configuration issue")
		os.Exit(1)
	}

	// 7. Display the results
	fmt.Printf("✅ Agent Response:\n")
	fmt.Printf("%s\n\n", response)

	fmt.Printf("📊 Stats:\n")
	fmt.Printf("   • Response time: %v\n", duration)
	fmt.Printf("   • Model used: %s\n", provider.Model)
	fmt.Printf("   • Success: true\n")

	// 8. Helpful next steps
	fmt.Printf("\n🚀 Next Steps:\n")
	fmt.Printf("   • Try another question: go run main.go \"Your next question\"\n")
	fmt.Printf("   • Explore multi-agent examples: cd ../02-multi-agent-collab\n")
	fmt.Printf("   • Read the tutorial: https://agenticgokit.dev/tutorials\n")
}