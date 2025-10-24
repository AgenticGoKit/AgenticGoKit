package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

func TestParallelWorkflowStreaming() {
	fmt.Println("🔬 Testing Parallel Workflow Streaming")
	fmt.Println("=====================================")

	// Create test agents with extended timeouts
	agent1, _ := vnext.QuickChatAgentWithConfig("Agent1", &vnext.Config{
		Name:         "agent1",
		SystemPrompt: "You are Agent 1. Be brief.",
		Timeout:      800 * time.Second, // Extended timeout for thorough testing
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
	})

	agent2, _ := vnext.QuickChatAgentWithConfig("Agent2", &vnext.Config{
		Name:         "agent2",
		SystemPrompt: "You are Agent 2. Be brief.",
		Timeout:      800 * time.Second, // Extended timeout for thorough testing
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
	})

	// Test Parallel Workflow with extended timeout
	workflow, err := vnext.NewParallelWorkflow(&vnext.WorkflowConfig{
		Mode:    vnext.Parallel,
		Timeout: 800 * time.Second, // Extended timeout for thorough testing
	})
	if err != nil {
		log.Fatalf("Failed to create parallel workflow: %v", err)
	}

	workflow.AddStep(vnext.WorkflowStep{Name: "task1", Agent: agent1})
	workflow.AddStep(vnext.WorkflowStep{Name: "task2", Agent: agent2})

	ctx := context.Background()
	workflow.Initialize(ctx)

	fmt.Println("Running parallel workflow with streaming...")
	startTime := time.Now()

	stream, err := workflow.RunStream(ctx, "Hello from parallel workflow")
	if err != nil {
		log.Fatalf("Parallel workflow streaming failed: %v", err)
	}

	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			fmt.Printf("❌ Error: %v\n", chunk.Error)
			break
		}

		switch chunk.Type {
		case vnext.ChunkTypeMetadata:
			if stepName, ok := chunk.Metadata["step_name"].(string); ok {
				fmt.Printf("[%s] %s\n", stepName, chunk.Content)
			} else {
				fmt.Printf("[WORKFLOW] %s\n", chunk.Content)
			}
		case vnext.ChunkTypeDelta:
			fmt.Print(chunk.Delta)
		case vnext.ChunkTypeDone:
			fmt.Println("\n✅ Step completed!")
		}
	}

	result, err := stream.Wait()
	duration := time.Since(startTime)

	if err != nil {
		log.Fatalf("Parallel workflow failed: %v", err)
	}

	fmt.Printf("✅ Parallel workflow completed: Success=%t, Duration=%.2fs\n",
		result.Success, duration.Seconds())

	workflow.Shutdown(ctx)
}

func TestSequentialWorkflowComparison() {
	fmt.Println("\n🔬 Testing Sequential vs Parallel Performance")
	fmt.Println("===========================================")

	// Create test agents
	agent1, _ := vnext.QuickChatAgentWithConfig("SeqAgent1", &vnext.Config{
		Name:         "seq_agent1",
		SystemPrompt: "You are Agent 1. Respond with exactly 20 words.",
		Timeout:      800 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
	})

	agent2, _ := vnext.QuickChatAgentWithConfig("SeqAgent2", &vnext.Config{
		Name:         "seq_agent2",
		SystemPrompt: "You are Agent 2. Respond with exactly 20 words.",
		Timeout:      800 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
	})

	// Test Sequential Workflow
	seqWorkflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Mode:    vnext.Sequential,
		Timeout: 800 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create sequential workflow: %v", err)
	}

	seqWorkflow.AddStep(vnext.WorkflowStep{Name: "seq_task1", Agent: agent1})
	seqWorkflow.AddStep(vnext.WorkflowStep{Name: "seq_task2", Agent: agent2})

	ctx := context.Background()
	seqWorkflow.Initialize(ctx)

	fmt.Println("Running sequential workflow...")
	seqStart := time.Now()

	seqStream, err := seqWorkflow.RunStream(ctx, "Sequential test prompt")
	if err != nil {
		log.Fatalf("Sequential workflow streaming failed: %v", err)
	}

	// Process sequential stream
	for chunk := range seqStream.Chunks() {
		if chunk.Error != nil {
			fmt.Printf("❌ Sequential Error: %v\n", chunk.Error)
			break
		}
		// Just consume chunks for timing
	}

	_, err = seqStream.Wait()
	seqDuration := time.Since(seqStart)

	if err != nil {
		log.Fatalf("Sequential workflow failed: %v", err)
	}

	fmt.Printf("✅ Sequential completed: Duration=%.2fs\n", seqDuration.Seconds())
	seqWorkflow.Shutdown(ctx)

	// Compare with parallel (would need separate agents)
	fmt.Printf("📊 Performance Comparison:\n")
	fmt.Printf("   Sequential: %.2fs\n", seqDuration.Seconds())
	fmt.Printf("   (Parallel timing from previous test)\n")
}

func main() {
	fmt.Println("🧪 Extended Workflow Streaming Integration Tests")
	fmt.Println("===============================================")

	// Quick connection check with extended timeout
	testAgent, _ := vnext.QuickChatAgentWithConfig("Test", &vnext.Config{
		Name:    "test",
		Timeout: 60 * time.Second, // Extended connection test timeout
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := testAgent.Run(ctx, "Hello")
	if err != nil {
		log.Fatalf("Connection test failed: %v", err)
	}

	fmt.Println("✅ Connection successful\n")

	// Run integration tests
	TestParallelWorkflowStreaming()
	TestSequentialWorkflowComparison()

	fmt.Println("\n🎉 All integration tests complete!")
}
