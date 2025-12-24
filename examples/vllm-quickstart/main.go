// Package main demonstrates vLLM integration with AgenticGoKit.
// This example shows how to use vLLM as an LLM provider for AI agents.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/agenticgokit/agenticgokit/core"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/vllm" // Import vLLM plugin
)

func main() {
	// Get vLLM configuration from environment variables
	baseURL := os.Getenv("VLLM_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000" // Default vLLM URL
	}

	model := os.Getenv("VLLM_MODEL")
	if model == "" {
		model = "meta-llama/Llama-2-7b-chat-hf" // Default model
	}

	// Create vLLM provider
	provider, err := core.NewModelProviderFromConfig(core.LLMProviderConfig{
		Type:        "vllm",
		BaseURL:     baseURL,
		Model:       model,
		MaxTokens:   2048,
		Temperature: 0.7,
		// vLLM-specific options (optional)
		VLLMTopP: 0.9,
		VLLMTopK: 50,
	})
	if err != nil {
		log.Fatalf("Failed to create vLLM provider: %v", err)
	}

	ctx := context.Background()

	// Example 1: Simple completion
	fmt.Println("=== Example 1: Simple Completion ===")
	resp, err := provider.Call(ctx, core.Prompt{
		System: "You are a helpful assistant that provides concise answers.",
		User:   "What is Go programming language and why is it popular?",
	})
	if err != nil {
		log.Fatalf("Call failed: %v", err)
	}
	fmt.Printf("Response: %s\n", resp.Content)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n\n",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)

	// Example 2: Streaming response
	fmt.Println("=== Example 2: Streaming Response ===")
	tokenChan, err := provider.Stream(ctx, core.Prompt{
		System: "You are a creative writer.",
		User:   "Write a haiku about programming.",
	})
	if err != nil {
		log.Fatalf("Stream failed: %v", err)
	}

	fmt.Print("Streaming: ")
	for token := range tokenChan {
		if token.Error != nil {
			log.Fatalf("Stream error: %v", token.Error)
		}
		fmt.Print(token.Content)
	}
	fmt.Println("\n")

	// Example 3: Code generation
	fmt.Println("=== Example 3: Code Generation ===")
	codeResp, err := provider.Call(ctx, core.Prompt{
		System: "You are an expert Go programmer. Provide clean, idiomatic Go code.",
		User:   "Write a function to reverse a string in Go.",
	})
	if err != nil {
		log.Fatalf("Code generation failed: %v", err)
	}
	fmt.Printf("Generated Code:\n%s\n", codeResp.Content)

	fmt.Println("\n=== vLLM Quick Start Complete ===")
}
