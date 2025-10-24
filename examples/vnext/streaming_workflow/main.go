package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

// CreateResearcherAgent creates a research agent
func CreateResearcherAgent() (vnext.Agent, error) {
	return vnext.QuickChatAgentWithConfig("Researcher", &vnext.Config{
		Name:         "researcher",
		SystemPrompt: "You are a Research Agent. Provide detailed information about the given topic. Be thorough and informative.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.2,
			MaxTokens:   400,
			BaseURL:     "http://localhost:11434",
		},
	})
}

// CreateSummarizerAgent creates a summarizer agent
func CreateSummarizerAgent() (vnext.Agent, error) {
	return vnext.QuickChatAgentWithConfig("Summarizer", &vnext.Config{
		Name:         "summarizer",
		SystemPrompt: "You are a Summarizer Agent. Create concise summaries of the given content. Focus on key points and main takeaways.",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.3,
			MaxTokens:   200,
			BaseURL:     "http://localhost:11434",
		},
	})
}

// streamAgentResponse handles streaming output from an agent
func streamAgentResponse(agent vnext.Agent, prompt string, stepName string) (string, error) {
	fmt.Printf("\n🔄 STEP: %s\n", stepName)
	fmt.Printf("📝 Prompt: %s\n", prompt)
	fmt.Println("💬 Streaming Response:")
	fmt.Println("─────────────────────")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := agent.RunStream(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to start streaming: %w", err)
	}

	var fullResponse string
	tokenCount := 0
	startTime := time.Now()

	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			return "", fmt.Errorf("streaming error: %v", chunk.Error)
		}

		switch chunk.Type {
		case vnext.ChunkTypeDelta:
			// Print each token immediately
			fmt.Print(chunk.Delta)
			fullResponse += chunk.Delta
			tokenCount++
		case vnext.ChunkTypeDone:
			duration := time.Since(startTime)
			fmt.Printf("\n✅ %s completed!\n", stepName)
			fmt.Printf("📊 Stats: %d tokens in %.1fs (%.1f tokens/sec)\n",
				tokenCount, duration.Seconds(), float64(tokenCount)/duration.Seconds())
		}
	}

	return fullResponse, nil
}

// RunSequentialWorkflowWithStreaming demonstrates a sequential workflow with streaming
func RunSequentialWorkflowWithStreaming() {
	fmt.Println("🌟 Sequential Workflow with Streaming")
	fmt.Println("=====================================")
	fmt.Println("Two agents working in sequence: Researcher → Summarizer")
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

	// Input topic
	topic := "Benefits of streaming in AI applications"
	fmt.Printf("🎯 Topic: %s\n", topic)

	// Step 1: Research
	researchPrompt := fmt.Sprintf("Research the topic: %s. Provide key information, benefits, and current applications.", topic)
	researchResult, err := streamAgentResponse(researcher, researchPrompt, "RESEARCH")
	if err != nil {
		log.Fatalf("Research step failed: %v", err)
	}

	// Step 2: Summarize (using research result as input)
	summaryPrompt := fmt.Sprintf("Please summarize this research into key points:\n\n%s", researchResult)
	summaryResult, err := streamAgentResponse(summarizer, summaryPrompt, "SUMMARIZE")
	if err != nil {
		log.Fatalf("Summary step failed: %v", err)
	}

	// Final results
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🎉 SEQUENTIAL WORKFLOW COMPLETED!")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("📊 Research Output: %d characters\n", len(researchResult))
	fmt.Printf("📊 Summary Output: %d characters\n", len(summaryResult))
	fmt.Println("✅ Both agents executed successfully with streaming")
	fmt.Println("🔄 Data flowed: Topic → Research → Summary")
}

func main() {
	fmt.Println("🚀 Simple Sequential Workflow with Streaming")
	fmt.Println("============================================")

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

	// Run the sequential workflow
	RunSequentialWorkflowWithStreaming()

	fmt.Println("\n🎉 Demo Complete!")
	fmt.Println("\n💡 What we demonstrated:")
	fmt.Println("• 🔄 Sequential execution: Research → Summarize")
	fmt.Println("• ⚡ Real-time streaming from each agent")
	fmt.Println("• 🤖 Agent specialization with different roles")
	fmt.Println("• 🛤️  Data flow between agents (research output → summary input)")
	fmt.Println("• 📊 Performance metrics for each step")
}
