package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

func main() {
	fmt.Println("🧪 Simple Workflow Streaming Test")
	fmt.Println("=================================")

	// Create agent with longer timeout
	agent, err := vnext.QuickChatAgentWithConfig("TestAgent", &vnext.Config{
		Name:         "test",
		SystemPrompt: "You are a helpful assistant. Be very brief.",
		Timeout:      120 * time.Second, // Longer timeout
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.5,
			MaxTokens:   50, // Shorter response
			BaseURL:     "http://localhost:11434",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create workflow with longer timeout
	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Mode:    vnext.Sequential,
		Timeout: 300 * time.Second, // Much longer timeout
	})
	if err != nil {
		log.Fatalf("Failed to create workflow: %v", err)
	}

	err = workflow.AddStep(vnext.WorkflowStep{
		Name:  "simple_step",
		Agent: agent,
	})
	if err != nil {
		log.Fatalf("Failed to add step: %v", err)
	}

	// Test with minimal context (no timeout)
	ctx := context.Background()
	if err := workflow.Initialize(ctx); err != nil {
		fmt.Printf("⚠️  Workflow initialization warning: %v\n", err)
	}

	fmt.Println("Starting simple workflow streaming test...")
	stream, err := workflow.RunStream(ctx, "Say hello")
	if err != nil {
		fmt.Printf("❌ Workflow streaming failed immediately: %v\n", err)
		return
	}

	fmt.Print("Response: ")
	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			fmt.Printf("❌ Error: %v\n", chunk.Error)
			break
		}

		switch chunk.Type {
		case vnext.ChunkTypeMetadata:
			fmt.Printf("[%s] ", chunk.Content)
		case vnext.ChunkTypeText:
			fmt.Print(chunk.Content)
		case vnext.ChunkTypeDelta:
			fmt.Print(chunk.Delta)
		case vnext.ChunkTypeDone:
			fmt.Println("\n✅ Success!")
		}
	}

	result, err := stream.Wait()
	if err != nil {
		fmt.Printf("❌ Wait failed: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %t, Duration: %.2fs\n", result.Success, result.Duration.Seconds())
	}

	workflow.Shutdown(ctx)
	fmt.Println("Test complete!")
}
