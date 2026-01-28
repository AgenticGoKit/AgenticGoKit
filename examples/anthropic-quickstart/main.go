package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/agenticgokit/agenticgokit/plugins/llm/anthropic"
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  Anthropic QuickStart")
	fmt.Println("===========================================")
	fmt.Println()

	// Check for API key in environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable not set. Please set it with your Anthropic API key.")
	}

	ctx := context.Background()

	// Example 1: Basic Usage with Config
	fmt.Println("Example 1: Basic Agent with Config")
	fmt.Println("====================================")

	config1 := &vnext.Config{
		Name:         "claude-assistant",
		SystemPrompt: "You are a helpful assistant.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "anthropic",
			Model:       "claude-sonnet-4-20250514",
			APIKey:      apiKey,
			Temperature: 0.7,
			MaxTokens:   500,
		},
	}

	agent1, err := vnext.NewBuilder("claude-assistant").
		WithConfig(config1).
		Build()

	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	if err := agent1.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}
	defer agent1.Cleanup(ctx)

	result1, err := agent1.Run(ctx, "What are the three laws of robotics?")
	if err != nil {
		log.Fatalf("Run failed: %v", err)
	}

	fmt.Printf("Response: %s\n", result1.Content)
	fmt.Printf("Duration: %v | Tokens: %d\n\n", result1.Duration, result1.TokensUsed)

	// Example 2: Streaming Responses
	fmt.Println("Example 2: Streaming Responses")
	fmt.Println("================================")

	streamAgent, err := vnext.NewBuilder("streaming-agent").
		WithConfig(config1).
		Build()

	if err != nil {
		log.Fatalf("Failed to create streaming agent: %v", err)
	}

	if err := streamAgent.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize streaming agent: %v", err)
	}
	defer streamAgent.Cleanup(ctx)

	fmt.Print("Streaming response: ")
	stream, err := streamAgent.RunStream(ctx, "Write a haiku about clouds.",
		vnext.WithBufferSize(10),
	)

	if err != nil {
		log.Fatalf("Stream failed: %v", err)
	}

	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			log.Fatalf("Stream error: %v", chunk.Error)
		}
		if chunk.Type == vnext.ChunkTypeDelta {
			fmt.Print(chunk.Delta)
		}
	}

	streamResult, err := stream.Wait()
	if err != nil {
		log.Fatalf("Stream wait failed: %v", err)
	}

	fmt.Printf("\n\nDuration: %v | Success: %v\n\n", streamResult.Duration, streamResult.Success)

	fmt.Println("===========================================")
	fmt.Println("  Anthropic examples completed!")
	fmt.Println("===========================================")
}
