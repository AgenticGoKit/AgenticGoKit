package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"

	"kunalkushwaha/agentflow/internal/factory"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/tools"
)

const (
	PlannerAgentName    = "planner"
	ResearcherAgentName = "researcher"
	SummarizerAgentName = "summarizer"
	FinalOutputAgent    = "final_output"
)

func main() {
	agentflow.SetLogLevel(agentflow.INFO)
	log.Println("Starting Multi-Agent Workflow Example (Factory Version)")

	// --- Setup Azure OpenAI ---
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")
	if endpoint == "" || apiKey == "" || chatDeployment == "" {
		log.Fatal("Error: Set AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, and AZURE_OPENAI_CHAT_DEPLOYMENT.")
	}
	azureAdapter, err := llm.NewAzureOpenAIAdapter(llm.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment,
	})
	if err != nil {
		log.Fatalf("Error creating Azure OpenAI Adapter: %v", err)
	}
	llmAdapter := llm.NewModelProviderAdapter(azureAdapter)

	// --- Tool Registry ---
	toolRegistry := tools.NewToolRegistry()
	if err := toolRegistry.Register(&tools.WebSearchTool{}); err != nil {
		log.Fatalf("Failed to register search tool: %v", err)
	}

	// --- WaitGroup for completion ---
	var wg sync.WaitGroup

	// --- Agent Handlers ---
	plannerHandler := NewPlannerHandler(azureAdapter)
	researcherHandler := NewResearcherHandler(*toolRegistry, llmAdapter)
	summarizerHandler := NewSummarizerHandler(azureAdapter, llmAdapter)
	finalOutputHandler := NewFinalOutputHandler(&wg)

	// --- Register all agents in a map ---
	agents := map[string]agentflow.AgentHandler{
		PlannerAgentName:    plannerHandler,
		ResearcherAgentName: researcherHandler,
		SummarizerAgentName: summarizerHandler,
		FinalOutputAgent:    finalOutputHandler,
	}

	// --- Create runner using factory ---
	runner := factory.NewRunnerWithConfig(factory.RunnerConfig{
		QueueSize: 2,
		Agents:    agents,
	})

	// --- Start the runner ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner.Start(ctx)

	// --- Prepare and Emit Initial Event ---
	initialEventID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	userRequest := "Research the recent developments in AI-powered code generation and summarize the key findings."
	event := agentflow.NewEvent(
		PlannerAgentName,
		agentflow.EventData{"user_request": userRequest},
		map[string]string{
			agentflow.RouteMetadataKey: PlannerAgentName,
			agentflow.SessionIDKey:     initialEventID,
		},
	)
	event.SetID(initialEventID)

	wg.Add(1)
	if err := runner.Emit(event); err != nil {
		log.Fatalf("Failed to emit initial event: %v", err)
	}

	// --- Wait for Completion ---
	wg.Wait()
	cancel()
	runner.Stop()

	// --- Save Trace ---
	traceDir := "traces"
	os.MkdirAll(traceDir, 0755)
	traceEntries, err := runner.DumpTrace(initialEventID)
	if err != nil {
		log.Printf("Error dumping trace: %v", err)
	} else {
		traceFile := filepath.Join(traceDir, initialEventID+".trace.json")
		saveTraceToFile(traceEntries, traceFile)
		fmt.Printf("\nTrace saved to: %s\n", traceFile)
		fmt.Println("View the trace with: agentcli trace --flow-only", initialEventID)
	}

	// --- Display Final Result ---
	finalState, found := finalOutputHandler.GetFinalState(initialEventID)
	if found {
		if summary, ok := finalState.Get("summary"); ok && summary != nil {
			fmt.Println("\n--- Final Summary ---")
			fmt.Printf("%v\n", summary)
			fmt.Println("---------------------")
		} else if errMsg, ok := finalState.Get("error"); ok && errMsg != nil {
			fmt.Printf("\nFlow completed with error: %v\n", errMsg)
		}
	}
}

func saveTraceToFile(entries []agentflow.TraceEntry, filePath string) {
	f, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create trace file: %v", err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		log.Printf("Failed to write trace data: %v", err)
	}
}
