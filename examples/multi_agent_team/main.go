package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/orchestrator"
	"kunalkushwaha/agentflow/internal/tools"
)

const (
	PlannerAgentName    = "planner"
	ResearcherAgentName = "researcher"
	SummarizerAgentName = "summarizer"
)

func main() {
	log.Println("Starting Multi-Agent Team Example...")

	// --- Configuration ---
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")
	if embeddingDeployment == "" {
		embeddingDeployment = "not-used"
	}
	if endpoint == "" || apiKey == "" || chatDeployment == "" {
		log.Fatal("Error: Set AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, and AZURE_OPENAI_CHAT_DEPLOYMENT.")
	}
	// TODO: Add env var for WebSearchTool API Key if needed

	// --- Setup ---
	// 1. LLM Adapter
	adapterOpts := llm.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment,
		HTTPClient:          &http.Client{Timeout: 90 * time.Second},
	}
	azureAdapter, err := llm.NewAzureOpenAIAdapter(adapterOpts)
	if err != nil {
		log.Fatalf("Error creating Azure OpenAI Adapter: %v", err)
	}
	log.Println("Azure OpenAI Adapter created.")

	// 2. Tool Registry
	toolRegistry := tools.NewToolRegistry()
	searchTool := &tools.WebSearchTool{} // Assumes WebSearchTool struct exists
	if err := toolRegistry.Register(searchTool); err != nil {
		log.Fatalf("Failed to register web search tool: %v", err)
	}
	log.Println("Tool Registry created and WebSearchTool registered.")

	// 3. Create Agents (Defined in agents.go)
	plannerAgent := NewPlannerAgent()
	researchAgent := NewResearchAgent(toolRegistry)
	summarizeAgent := NewSummarizeAgent(azureAdapter)
	log.Println("Agents created (Planner, Researcher, Summarizer).")

	// 4. Create Orchestrator & Runner
	orchestratorImpl := orchestrator.NewRouteOrchestrator()
	concurrency := runtime.NumCPU()
	runner := agentflow.NewRunner(orchestratorImpl, concurrency)
	log.Printf("Runner & Route Orchestrator created (concurrency %d).", concurrency)

	// Create ONE shared result channel BEFORE handlers
	resultChan := make(chan agentflow.State, 1) // Buffer of 1 is important

	// 5. Create Agent Handlers (Defined in handler.go) & Register
	var wg sync.WaitGroup
	agentNames := []string{PlannerAgentName, ResearcherAgentName, SummarizerAgentName}

	// Pass the SAME resultChan to all handlers
	plannerHandler := NewAgentHandler(PlannerAgentName, plannerAgent, runner, azureAdapter, agentNames, resultChan, &wg)
	researcherHandler := NewAgentHandler(ResearcherAgentName, researchAgent, runner, azureAdapter, agentNames, resultChan, &wg)
	summarizerHandler := NewAgentHandler(SummarizerAgentName, summarizeAgent, runner, azureAdapter, agentNames, resultChan, &wg)

	runner.RegisterAgent(PlannerAgentName, plannerHandler)
	runner.RegisterAgent(ResearcherAgentName, researcherHandler)
	runner.RegisterAgent(SummarizerAgentName, summarizerHandler)
	log.Println("Agent Handlers created and registered.")

	// --- Execution ---
	// 1. Prepare the initial event
	initialEventID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	eventPayload := map[string]interface{}{
		"user_request": "Research about France and summarize it", // Example request
	}
	eventMetadata := map[string]string{
		orchestrator.RouteMetadataKey: PlannerAgentName, // Start with the Planner
		"original_event_id":           initialEventID,
		"user_request":                eventPayload["user_request"].(string),
	}
	event := &agentflow.SimpleEvent{
		ID:       initialEventID,
		Payload:  eventPayload,
		Metadata: eventMetadata,
	}
	log.Printf("Initial event prepared: %s", event.GetID())

	// 2. Emit the initial event
	log.Println("Emitting initial event to start the flow...")
	wg.Add(1) // Add 1 for the *first* handler call (Planner)
	runner.Emit(event)
	log.Println("Initial event emitted.")

	// 3. Wait for the final result
	log.Println("Waiting for final result/error...")
	var finalState agentflow.State
	select {
	case finalState = <-resultChan:
		log.Println("Received final state.")
	case <-time.After(180 * time.Second): // Timeout for the whole flow
		log.Fatal("Timeout waiting for the multi-agent flow to complete.")
	}

	// 4. Stop the runner
	log.Println("Stopping runner...")
	runner.Stop()
	log.Println("Runner stopped.")

	// --- Output ---
	if finalState.GetData() == nil {
		log.Fatal("Failed to get final state.")
	}

	// Check for errors first in the output
	if handlerErr, ok := finalState.Get("handler_error"); ok {
		log.Printf("Flow ended with handler error: %v", handlerErr)
	} else if agentErr, ok := finalState.Get("agent_error"); ok {
		log.Printf("Flow ended with agent error: %v", agentErr)
	} else if summary, ok := finalState.Get("summary"); ok {
		log.Println("Multi-agent flow successful.")
		fmt.Println("\n--- Final Summary ---")
		fmt.Println(summary)
		fmt.Println("---------------------")
	} else {
		log.Println("Flow ended, but no summary or known error found in final state.")
	}
	log.Printf("Final state data: %+v", finalState.GetData())
}
