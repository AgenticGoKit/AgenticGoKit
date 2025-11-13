package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// CreateResearcherAgent creates a research agent
func CreateResearcherAgent() (vnext.Agent, error) {
	return vnext.QuickChatAgentWithConfig("Researcher", &vnext.Config{
		Name:         "researcher",
		SystemPrompt: "You are a Research Agent. Provide detailed information about the given topic. Be thorough and informative.",
		Timeout:      60 * time.Second, // Increased timeout for workflow
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.2,
			MaxTokens:   300,
			BaseURL:     "http://localhost:11434",
		},
	})
}

// CreateSummarizerAgent creates a summarizer agent
func CreateSummarizerAgent() (vnext.Agent, error) {
	return vnext.QuickChatAgentWithConfig("Summarizer", &vnext.Config{
		Name:         "summarizer",
		SystemPrompt: "You are a Summarizer Agent. Create concise summaries of the given content. Focus on key points and main takeaways.",
		Timeout:      60 * time.Second, // Increased timeout for workflow
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.3,
			MaxTokens:   150,
			BaseURL:     "http://localhost:11434",
		},
	})
}

// RunSequentialWorkflowWithVNextStreaming demonstrates the FIXED vnext.Workflow streaming
func RunSequentialWorkflowWithVNextStreaming() {
	fmt.Println("🌟 FIXED vnext.Workflow Sequential Streaming")
	fmt.Println("===========================================")
	fmt.Println("Using real vnext.Workflow with streaming support!")
	fmt.Println()

	// Create agents
	researcher, err := CreateResearcherAgent()
	if err != nil {
		log.Fatalf("Failed to create researcher: %v", err)
	}

	summarizer, err := CreateSummarizerAgent()
	if err != nil {
		log.Fatalf("Failed to create summarizer: %v", err)
	}

	// Create workflow
	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Mode:    vnext.Sequential,
		Timeout: 180 * time.Second, // 3 minutes for the whole workflow
	})
	if err != nil {
		log.Fatalf("Failed to create workflow: %v", err)
	}

	// Add workflow steps
	err = workflow.AddStep(vnext.WorkflowStep{
		Name:  "research",
		Agent: researcher,
		Transform: func(input string) string {
			return fmt.Sprintf("Research the topic: %s. Provide key information, benefits, and current applications.", input)
		},
	})
	if err != nil {
		log.Fatalf("Failed to add research step: %v", err)
	}

	err = workflow.AddStep(vnext.WorkflowStep{
		Name:  "summarize",
		Agent: summarizer,
		Transform: func(input string) string {
			return fmt.Sprintf("Please summarize this research into key points:\n\n%s", input)
		},
	})
	if err != nil {
		log.Fatalf("Failed to add summarize step: %v", err)
	}

	// Input topic
	topic := "Benefits of streaming in AI applications"
	fmt.Printf("🎯 Topic: %s\n", topic)
	fmt.Printf("🔄 Processing through workflow...\n\n")

	// Initialize workflow
	ctx := context.Background()
	if err := workflow.Initialize(ctx); err != nil {
		fmt.Printf("⚠️ Workflow initialization warning: %v\n", err)
	}

	// Run workflow with streaming
	startTime := time.Now()
	stream, err := workflow.RunStream(ctx, topic)
	if err != nil {
		log.Fatalf("Workflow streaming failed: %v", err)
	}

	var finalOutput string
	stepOutputs := make(map[string]string)
	chunkCount := 0

	fmt.Println("💬 Real-time Workflow Streaming:")
	fmt.Println("─────────────────────────────────")

	for chunk := range stream.Chunks() {
		chunkCount++

		if chunk.Error != nil {
			fmt.Printf("❌ Error in chunk %d: %v\n", chunkCount, chunk.Error)
			break
		}

		switch chunk.Type {
		case vnext.ChunkTypeMetadata:
			if stepName, ok := chunk.Metadata["step_name"].(string); ok {
				fmt.Printf("\n🔄 [STEP: %s] %s\n", strings.ToUpper(stepName), chunk.Content)
				fmt.Println("─────────────────────")
			} else {
				fmt.Printf("\n📋 [WORKFLOW] %s\n", chunk.Content)
			}
		case vnext.ChunkTypeText:
			fmt.Print(chunk.Content)
			finalOutput += chunk.Content
		case vnext.ChunkTypeDelta:
			fmt.Print(chunk.Delta)
			finalOutput += chunk.Delta
			// Track step outputs
			if stepName, ok := chunk.Metadata["step_name"].(string); ok {
				stepOutputs[stepName] += chunk.Delta
			}
		case vnext.ChunkTypeDone:
			fmt.Printf("\n✅ Workflow step completed!")
		}
	}

	// Get final result
	result, err := stream.Wait()
	duration := time.Since(startTime)

	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	// Display results
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 vnext.WORKFLOW STREAMING COMPLETED!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("✅ Success: %t\n", result.Success)
	fmt.Printf("⏱️ Duration: %.2f seconds\n", duration.Seconds())
	fmt.Printf("📊 Total Chunks: %d\n", chunkCount)
	fmt.Printf("📄 Final Output Length: %d characters\n", len(finalOutput))

	// Show step breakdown
	fmt.Println("\n📋 Step Breakdown:")
	for stepName, output := range stepOutputs {
		fmt.Printf("  🔸 %s: %d chars\n", strings.Title(stepName), len(output))
	}

	workflow.Shutdown(ctx)
}

func main() {
	fmt.Println("🚀 vnext.Workflow Streaming Showcase")
	fmt.Println("====================================")
	fmt.Println("Demonstrating the FIXED vnext.Workflow streaming!")
	fmt.Println()

	// Quick connection test
	fmt.Println("🔍 Testing Ollama connection...")
	testAgent, err := vnext.QuickChatAgentWithConfig("Test", &vnext.Config{
		Name:    "test",
		Timeout: 10 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
			BaseURL:  "http://localhost:11434",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create test agent: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = testAgent.Run(ctx, "Hello")
	if err != nil {
		log.Fatalf("Ollama connection test failed: %v", err)
	}

	fmt.Println("✅ Ollama connection successful")
	fmt.Println()

	// Run the FIXED vnext.Workflow streaming
	RunSequentialWorkflowWithVNextStreaming()

	fmt.Println("\n🎉 Demo Complete!")
	fmt.Println("• ✅ Real-time streaming from workflow")
	fmt.Println("• 🔄 Automatic data flow between steps")
	fmt.Println("• 🛡️ Robust error handling and recovery")
	fmt.Println("• 📊 Built-in progress tracking and metadata")
	fmt.Println("• 🎯 Cleaner, more maintainable code")
	fmt.Println("• 🚀 Better performance and reliability")
}



