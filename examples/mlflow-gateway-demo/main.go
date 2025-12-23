// Package main demonstrates MLFlow AI Gateway integration with AgenticGoKit.
// This example shows how to use MLFlow AI Gateway as a unified LLM provider.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/agenticgokit/agenticgokit/core"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/mlflow" // Import MLFlow plugin
)

func main() {
	// Get MLFlow Gateway configuration from environment variables
	baseURL := os.Getenv("MLFLOW_GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5001" // Default MLFlow Gateway URL
	}

	chatRoute := os.Getenv("MLFLOW_CHAT_ROUTE")
	if chatRoute == "" {
		chatRoute = "chat" // Default route name
	}

	embeddingsRoute := os.Getenv("MLFLOW_EMBEDDINGS_ROUTE")
	if embeddingsRoute == "" {
		embeddingsRoute = "embeddings"
	}

	// Create MLFlow Gateway provider
	provider, err := core.NewModelProviderFromConfig(core.LLMProviderConfig{
		Type:                  "mlflow",
		BaseURL:               baseURL,
		MLFlowChatRoute:       chatRoute,
		MLFlowEmbeddingsRoute: embeddingsRoute,
		MaxTokens:             2048,
		Temperature:           0.7,
		// Retry configuration (optional)
		MLFlowMaxRetries: 3,
	})
	if err != nil {
		log.Fatalf("Failed to create MLFlow Gateway provider: %v", err)
	}

	ctx := context.Background()

	// Example 1: Simple completion
	fmt.Println("=== Example 1: Simple Completion ===")
	resp, err := provider.Call(ctx, core.Prompt{
		System: "You are a helpful assistant that provides concise answers.",
		User:   "What is MLFlow and why is it useful for ML operations?",
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
		System: "You are an expert in AI/ML systems.",
		User:   "Explain the benefits of using an AI Gateway for LLM access.",
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

	// Example 3: Generate embeddings (if embeddings route is configured)
	fmt.Println("=== Example 3: Generate Embeddings ===")
	texts := []string{
		"Machine learning is a subset of artificial intelligence.",
		"Deep learning uses neural networks with multiple layers.",
	}
	embeddings, err := provider.Embeddings(ctx, texts)
	if err != nil {
		fmt.Printf("Embeddings failed (this may be expected if embeddings route is not configured): %v\n", err)
	} else {
		fmt.Printf("Generated %d embeddings\n", len(embeddings))
		for i, emb := range embeddings {
			fmt.Printf("  Text %d: %d dimensions\n", i+1, len(emb))
		}
	}

	fmt.Println("\n=== MLFlow Gateway Demo Complete ===")
}
