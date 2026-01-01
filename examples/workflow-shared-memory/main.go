package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/agenticgokit/agenticgokit/plugins/memory/chromem" // Register chromem memory provider
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

func main() {
	fmt.Println("\n🤖 WORKFLOW SHARED MEMORY DEMO\n")

	ctx := context.Background()

	// Create shared memory
	sharedMemory, err := vnext.NewMemory(&vnext.MemoryConfig{
		Enabled:  true,
		Provider: "chromem",
	})
	if err != nil {
		log.Fatalf("Failed to create shared memory: %v", err)
	}

	// Create Agent 1 (Information Learner)
	agent1, err := vnext.NewBuilder("info-learner").
		WithConfig(&vnext.Config{
			Name: "info-learner",
			SystemPrompt: `You are an Information Learner. Your job:
Extract key facts from the input. Output ONLY the facts in this format:

Company Name: [name]
- [key fact 1]
- [key fact 2]
- [key fact 3]

Do NOT include any explanations or extra text. Just the facts.`,
			Timeout: 30 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: 0.5,
				MaxTokens:   150,
			},
		}).
		Build()

	if err != nil {
		log.Fatalf("Failed to create agent 1: %v", err)
	}

	if err := agent1.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize agent 1: %v", err)
	}
	defer agent1.Cleanup(ctx)

	// Create Agent 2 (Question Answerer)
	agent2, err := vnext.NewBuilder("question-answerer").
		WithConfig(&vnext.Config{
			Name: "question-answerer",
			SystemPrompt: `You are a Question Answerer. Your job:
Answer the question based ONLY on the learned facts provided.
Output just the answer - nothing else.`,
			Timeout: 30 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: 0.5,
				MaxTokens:   150,
			},
		}).
		Build()

	if err != nil {
		log.Fatalf("Failed to create agent 2: %v", err)
	}

	if err := agent2.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize agent 2: %v", err)
	}
	defer agent2.Cleanup(ctx)

	// Create workflow with shared memory
	workflow, err := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Mode:    vnext.Sequential,
		Timeout: 120 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create workflow: %v", err)
	}

	workflow.SetMemory(sharedMemory)

	// Add agents to workflow
	workflow.AddStep(vnext.WorkflowStep{
		Name:  "learn",
		Agent: agent1,
	})

	// Agent 2 receives ONLY the question (via Transform)
	// But it automatically has access to Agent 1's output through shared memory!
	// The workflow stores Agent 1's output in shared memory,
	// and Agent 2 queries it via GetWorkflowMemory(ctx)
	workflow.AddStep(vnext.WorkflowStep{
		Name:  "answer",
		Agent: agent2,
		Transform: func(_ string) string {
			// Ignore Agent 1's direct output - just pass the question
			// Agent 2 will get context from shared memory automatically
			return "What company was founded in 2020 and focuses on AI tools?"
		},
	})

	// Input data - will be processed by Agent 1
	companyInfo := `Company: TechStart Inc
- Founded in 2020
- Focuses on AI tools
- Has 50 employees
- Located in San Francisco
- Annual revenue: $10 million`

	fmt.Println("INPUT (Agent 1 learns this):")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println(companyInfo)
	fmt.Println("─────────────────────────────────────────────\n")

	fmt.Println("QUESTION (Agent 2 gets this + shared memory context):")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println("What company was founded in 2020 and focuses on AI tools?")
	fmt.Println("─────────────────────────────────────────────\n")

	// Run workflow
	fmt.Println("PROCESSING...")
	startTime := time.Now()

	result, err := workflow.Run(ctx, companyInfo)
	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	duration := time.Since(startTime)

	// Show results
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("RESULTS")
	fmt.Println(strings.Repeat("=", 70))

	for i, stepResult := range result.StepResults {
		if i == 0 {
			fmt.Printf("\n1️⃣  AGENT 1 (Extract Facts):\n")
		} else {
			fmt.Printf("\n2️⃣  AGENT 2 (Answer Question):\n")
		}
		fmt.Println(stepResult.Output)
	}

	fmt.Printf("\n⏱️  Total Time: %.2f seconds\n", duration.Seconds())
	fmt.Println("✅ Status: " + fmt.Sprintf("%v", result.Success))
	fmt.Println()
}
