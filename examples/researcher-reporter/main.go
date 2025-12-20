package main

import (
	"context"
	"fmt"
	"log"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
)

func main() {
	ctx := context.Background()

	// Step 1: Create two agents with different roles
	researcher := createAgent("researcher", "Gather key facts and information about the topic.", 0.7)
	reporter := createAgent("reporter", "Create a concise report with Summary, Key Points, and Conclusion.", 0.5)

	// Step 2: Create a sequential workflow
	workflow, _ := vnext.NewSequentialWorkflow(&vnext.WorkflowConfig{
		Mode:    vnext.Sequential,
		Timeout: 600 * time.Second,
	})

	// Step 3: Add agents as workflow steps
	workflow.AddStep(vnext.WorkflowStep{Name: "research", Agent: researcher})
	workflow.AddStep(vnext.WorkflowStep{
		Name:  "report",
		Agent: reporter,
		Transform: func(input string) string {
			return "Based on this research, create a professional report:\n\n" + input
		},
	})

	// Step 4: Run the workflow!
	workflow.Initialize(ctx)
	defer workflow.Shutdown(ctx)

	topic := "What is 2(3+6)"
	fmt.Printf("📋 Topic: %s\n\n", topic)

	result, err := workflow.Run(ctx, topic)
	if err != nil {
		log.Fatal(err)
	}

	// Display results
	fmt.Println("� RESEARCH FINDINGS:")
	fmt.Println(result.StepResults[0].Output)

	fmt.Println("\n📄 FINAL REPORT:")
	fmt.Println(result.StepResults[1].Output)

	fmt.Printf("\n✅ Completed in %v\n", result.Duration)
}

// Helper function to create an agent
func createAgent(name, prompt string, temperature float32) vnext.Agent {
	agent, err := vnext.NewBuilder(name).
		WithConfig(&vnext.Config{
			Name:         name,
			SystemPrompt: prompt,
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: temperature,
				MaxTokens:   400,
			},
			Timeout: 600 * time.Second,
		}).
		Build()

	if err != nil {
		log.Fatalf("Failed to create %s: %v", name, err)
	}

	agent.Initialize(context.Background())
	return agent
}



