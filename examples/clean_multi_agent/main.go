package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	agentflow "github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/llm"
	"github.com/kunalkushwaha/agentflow/internal/tools"
)

const (
	PlannerAgentName    = "planner"
	ResearcherAgentName = "researcher"
	SummarizerAgentName = "summarizer"
	FinalOutputAgent    = "final_output"
)

func main() {
	log.Println("Starting Multi-Agent Workflow Example")

	// --- Setup Azure OpenAI ---
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")

	if endpoint == "" || apiKey == "" || chatDeployment == "" {
		log.Fatal("Error: Set AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, and AZURE_OPENAI_CHAT_DEPLOYMENT.")
	}

	// --- 1. Create Azure OpenAI Adapter ---
	azureAdapter, err := llm.NewAzureOpenAIAdapter(llm.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment,
		HTTPClient:          &http.Client{Timeout: 60 * time.Second},
	})
	if err != nil {
		log.Fatalf("Error creating Azure OpenAI Adapter: %v", err)
	}

	// Create LLM adapter
	llmAdapter := llm.NewModelProviderAdapter(azureAdapter)

	// --- 2. Setup Tool Registry ---
	toolRegistry := tools.NewToolRegistry()
	searchTool := &tools.WebSearchTool{}
	if err := toolRegistry.Register(searchTool); err != nil {
		log.Fatalf("Failed to register search tool: %v", err)
	}

	// --- 3. Core Component Setup ---
	callbackRegistry := agentflow.NewCallbackRegistry()

	// Add a SINGLE safety circuit breaker (combines transition counter and agent visit tracking)
	// callbackRegistry.Register(agentflow.HookBeforeAgentRun, "safetyCircuitBreaker",
	// 	func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
	// 		// Get transition count from context
	// 		transitionKey := fmt.Sprintf("transitions-%s", args.Event.GetID())
	// 		visitKey := fmt.Sprintf("visits-%s-%s", args.Event.GetID(), args.AgentID)

	// 		// Get or initialize the counts from context
	// 		counts := getOrInitializeCounts(ctx)

	// 		// Increment counts
	// 		counts[transitionKey]++
	// 		counts[visitKey]++

	// 		// Global circuit breaker (max 10 transitions)
	// 		if counts[transitionKey] > 10 {
	// 			log.Printf("Circuit breaker: Max transitions reached")
	// 			args.State.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
	// 			args.State.Set("error", "Max transitions limit reached")
	// 			return args.State, nil
	// 		}

	// 		// Per-agent circuit breaker (max 3 visits to same agent)
	// 		if counts[visitKey] > 3 {
	// 			log.Printf("Circuit breaker: Agent %s visited too many times", args.AgentID)
	// 			args.State.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
	// 			args.State.Set("error", fmt.Sprintf("Agent %s visited too many times", args.AgentID))
	// 			return args.State, nil
	// 		}

	// 		return args.State, nil
	// 	})	// --- 4. Create Orchestrator and Runner ---
	orchestratorImpl := agentflow.NewRouteOrchestrator(callbackRegistry)
	runner := agentflow.NewRunner(2) // Using 2 for concurrency - simple example
	runner.SetOrchestrator(orchestratorImpl)
	runner.SetCallbackRegistry(callbackRegistry)
	orchestratorImpl.SetEmitter(runner)

	// --- 5. Setup Tracing (SINGLE tracing approach) ---
	traceLogger := agentflow.NewInMemoryTraceLogger()
	runner.SetTraceLogger(traceLogger)
	// Register trace hooks - this is the ONLY trace registration needed
	agentflow.RegisterTraceHooks(callbackRegistry, traceLogger)

	// --- 6. Create Agent Handlers ---
	var wg sync.WaitGroup
	plannerHandler := NewPlannerHandler(azureAdapter)
	researcherHandler := NewResearcherHandler(toolRegistry, llmAdapter)
	summarizerHandler := NewSummarizerHandler(azureAdapter, llmAdapter)
	finalOutputHandler := NewFinalOutputHandler(&wg)

	// --- 7. Register Handlers ---
	mustRegisterAgent(orchestratorImpl, PlannerAgentName, plannerHandler)
	mustRegisterAgent(orchestratorImpl, ResearcherAgentName, researcherHandler)
	mustRegisterAgent(orchestratorImpl, SummarizerAgentName, summarizerHandler)
	mustRegisterAgent(orchestratorImpl, FinalOutputAgent, finalOutputHandler)

	// --- 8. Start the Runner ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner.Start(ctx)

	// --- 9. Prepare and Emit Initial Event ---
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

	// Start the flow
	wg.Add(1)
	if err := runner.Emit(event); err != nil {
		log.Fatalf("Failed to emit initial event: %v", err)
	}

	// --- 10. Wait for Completion ---
	wg.Wait() // Simple wait pattern - finalOutputHandler calls wg.Done()

	// --- 11. Shutdown ---
	cancel()
	runner.Stop()

	// --- 12. Save Trace ---
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

	// --- 13. Display Final Result ---
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

// Helper functions : For Circuit Breaker
// func getOrInitializeCounts(ctx context.Context) map[string]int {
// 	countsKey := "transition_counts"
// 	counts, ok := ctx.Value(countsKey).(map[string]int)
// 	if !ok {
// 		counts = make(map[string]int)
// 		ctx = context.WithValue(ctx, countsKey, counts)
// 	}
// 	return counts
// }

func mustRegisterAgent(orchestrator *agentflow.RouteOrchestrator, name string, handler agentflow.AgentHandler) {
	if err := orchestrator.RegisterAgent(name, handler); err != nil {
		log.Fatalf("Failed to register %s agent: %v", name, err)
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
