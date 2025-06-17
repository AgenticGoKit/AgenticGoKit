// Package main provides a simple test for debugging Ollama hanging issues.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agentflow/internal/llm"
)

func main() {
	fmt.Println("🔍 Testing Ollama Integration - Simple Debug Version")
	fmt.Println("====================================================")

	// Test 1: Basic Ollama adapter
	fmt.Println("\n📡 Test 1: Creating Ollama Adapter")
	adapter, err := llm.NewOllamaAdapter("", "llama3.2:latest", 100, 0.7)
	if err != nil {
		log.Fatalf("❌ Failed to create adapter: %v", err)
	}
	fmt.Println("✅ Adapter created successfully")

	// Test 2: Simple call with timeout
	fmt.Println("\n📡 Test 2: Simple Ollama Call")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	prompt := llm.Prompt{
		System: "You are a helpful assistant. Respond with just one word.",
		User:   "Say hello",
		Parameters: llm.ModelParameters{
			MaxTokens:   func() *int32 { v := int32(10); return &v }(),
			Temperature: func() *float32 { v := float32(0.1); return &v }(),
		},
	}

	fmt.Printf("🧠 Making Ollama call...\n")
	start := time.Now()

	response, err := adapter.Call(ctx, prompt)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Call failed after %v: %v\n", duration, err)
		return
	}

	fmt.Printf("✅ Call succeeded in %v\n", duration)
	fmt.Printf("📝 Response: %s\n", response.Content)

	// Test 3: Multiple calls with delays
	fmt.Println("\n📡 Test 3: Multiple Calls with Delays")

	for i := 1; i <= 3; i++ {
		fmt.Printf("\n🔄 Call %d/3\n", i)

		callCtx, callCancel := context.WithTimeout(context.Background(), 10*time.Second)
		start := time.Now()

		testPrompt := llm.Prompt{
			System: "Respond with a JSON array of tools",
			User:   fmt.Sprintf("Test query %d", i),
			Parameters: llm.ModelParameters{
				MaxTokens:   func() *int32 { v := int32(20); return &v }(),
				Temperature: func() *float32 { v := float32(0.1); return &v }(),
			},
		}

		response, err := adapter.Call(callCtx, testPrompt)
		duration := time.Since(start)
		callCancel()

		if err != nil {
			fmt.Printf("❌ Call %d failed after %v: %v\n", i, duration, err)
		} else {
			fmt.Printf("✅ Call %d succeeded in %v: %s\n", i, duration, response.Content)
		}

		// Add delay between calls
		if i < 3 {
			fmt.Printf("⏳ Waiting 2 seconds...\n")
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Println("\n🎉 Debug test completed!")
}
