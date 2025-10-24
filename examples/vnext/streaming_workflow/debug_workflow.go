package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

// TestWorkflowStreaming tests the actual vnext workflow streaming to identify the bug
func TestWorkflowStreaming() {
	fmt.Println("🔍 Testing vnext.Workflow Streaming Bug")
	fmt.Println("======================================")

	// Create a simple agent
	agent, err := vnext.QuickChatAgentWithConfig("TestAgent", &vnext.Config{
		Name:         "test",
		SystemPrompt: "You are a helpful assistant. Be brief.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.5,
			MaxTokens:   100,
			BaseURL:     "http://localhost:11434",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Test 1: Direct agent streaming (should work)
	fmt.Println("\n✅ Test 1: Direct Agent Streaming")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel1()

	stream1, err := agent.RunStream(ctx1, "What is streaming?")
	if err != nil {
		log.Fatalf("Direct agent streaming failed: %v", err)
	}

	fmt.Print("Response: ")
	for chunk := range stream1.Chunks() {
		if chunk.Error != nil {
			fmt.Printf("Error: %v\n", chunk.Error)
			break
		}
		if chunk.Type == vnext.ChunkTypeDelta {
			fmt.Print(chunk.Delta)
		}
		if chunk.Type == vnext.ChunkTypeDone {
			fmt.Println("\n✅ Direct streaming works!")
		}
	}

	// Test 2: Workflow streaming (likely to fail)
	fmt.Println("\n❓ Test 2: Workflow Streaming")

	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Mode:    vnext.Sequential,
		Timeout: 60 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create workflow: %v", err)
	}

	err = workflow.AddStep(vnext.WorkflowStep{
		Name:  "test_step",
		Agent: agent,
	})
	if err != nil {
		log.Fatalf("Failed to add step: %v", err)
	}

	ctx2 := context.Background()
	if err := workflow.Initialize(ctx2); err != nil {
		fmt.Printf("⚠️  Workflow initialization warning: %v\n", err)
	}

	fmt.Println("Starting workflow streaming...")
	stream2, err := workflow.RunStream(ctx2, "What is streaming?")
	if err != nil {
		fmt.Printf("❌ Workflow streaming failed immediately: %v\n", err)
		return
	}

	fmt.Print("Workflow Response: ")
	chunkCount := 0
	for chunk := range stream2.Chunks() {
		chunkCount++
		fmt.Printf("[Chunk %d: %s] ", chunkCount, chunk.Type)

		if chunk.Error != nil {
			fmt.Printf("❌ Workflow streaming error: %v\n", chunk.Error)
			break
		}

		switch chunk.Type {
		case vnext.ChunkTypeMetadata:
			fmt.Printf("Meta: %s ", chunk.Content)
		case vnext.ChunkTypeText:
			fmt.Print(chunk.Content)
		case vnext.ChunkTypeDelta:
			fmt.Print(chunk.Delta)
		case vnext.ChunkTypeDone:
			fmt.Println("\n✅ Workflow streaming completed!")
		}
	}

	result, err := stream2.Wait()
	if err != nil {
		fmt.Printf("❌ Workflow streaming wait failed: %v\n", err)
	} else {
		fmt.Printf("✅ Workflow result: Success=%t, Duration=%.2fs\n",
			result.Success, result.Duration.Seconds())
	}

	workflow.Shutdown(ctx2)
}

func main() {
	TestWorkflowStreaming()
}
