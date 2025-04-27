package main

import (
	"context" // Import context
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
	FinalOutputAgent    = "final_output" // A conceptual sink or final step
)

// Simple final output handler to log the result
type FinalOutputHandler struct {
	wg *sync.WaitGroup
	mu sync.Mutex
	// Store final results keyed by session ID
	finalResults map[string]agentflow.State
}

func NewFinalOutputHandler(wg *sync.WaitGroup) *FinalOutputHandler {
	return &FinalOutputHandler{
		wg:           wg,
		finalResults: make(map[string]agentflow.State),
	}
}

func (h *FinalOutputHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	sessionID := event.GetSessionID()
	log.Printf("--- Final Output Handler ---")
	log.Printf("Received final state for session %s:", sessionID)

	// FIX: Use state.Get(key) instead of GetMust
	summaryVal, _ := state.Get("summary")
	errorVal, _ := state.Get("error")
	log.Printf("State Data (Summary): %v", summaryVal)
	log.Printf("State Data (Error): %v", errorVal)

	// FIX: Use state.MetaKeys() and state.GetMeta(key) instead of GetAllMeta
	log.Println("State Meta:")
	for _, key := range state.MetaKeys() {
		metaVal, _ := state.GetMeta(key)
		log.Printf("  %s: %v", key, metaVal)
	}
	// log.Printf("State Meta: %+v", state.GetAllMeta()) // REMOVE THIS

	h.finalResults[sessionID] = state // Store the final state
	if h.wg != nil {
		// This might be tricky if multiple events reach here.
		// The WaitGroup should ideally track the completion of the *entire flow*
		// rather than just individual handler runs.
		// For simplicity, we might remove WG from this handler or rethink its usage.
		// h.wg.Done() // Decrementing here assumes this is the absolute last step.
	}
	return agentflow.AgentResult{OutputState: state}, nil // Return the state, no further routing
}

func (h *FinalOutputHandler) GetFinalState(sessionID string) (agentflow.State, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	state, found := h.finalResults[sessionID]
	return state, found
}

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

	// 3. Create Core Components
	callbackRegistry := agentflow.NewCallbackRegistry() // Create registry
	// TODO: Register callbacks if needed (e.g., for logging)
	// callbackRegistry.Register(agentflow.HookAfterAgentRun, "logAgent", logAgentCallback)

	orchestratorImpl := orchestrator.NewRouteOrchestrator(callbackRegistry) // Pass registry
	concurrency := runtime.NumCPU()
	runner := agentflow.NewRunner(concurrency)   // Pass queue size
	runner.SetOrchestrator(orchestratorImpl)     // Set orchestrator
	runner.SetCallbackRegistry(callbackRegistry) // Set registry
	log.Printf("Runner & Route Orchestrator created (concurrency %d).", concurrency)

	// 4. Create Agent Handlers (Defined in handler.go, implementing agentflow.AgentHandler)
	var wg sync.WaitGroup // WaitGroup to track completion of the *flow*
	plannerHandler := NewPlannerHandler(azureAdapter)
	researcherHandler := NewResearcherHandler(toolRegistry)
	summarizerHandler := NewSummarizerHandler(azureAdapter)
	finalOutputHandler := NewFinalOutputHandler(&wg) // Handler to receive the final result

	// 5. Register Handlers with Orchestrator
	err = orchestratorImpl.RegisterAgent(PlannerAgentName, plannerHandler)
	requireNoError(err, "Register Planner")
	err = orchestratorImpl.RegisterAgent(ResearcherAgentName, researcherHandler)
	requireNoError(err, "Register Researcher")
	err = orchestratorImpl.RegisterAgent(SummarizerAgentName, summarizerHandler)
	requireNoError(err, "Register Summarizer")
	err = orchestratorImpl.RegisterAgent(FinalOutputAgent, finalOutputHandler) // Register final step
	requireNoError(err, "Register FinalOutput")
	log.Println("Agent Handlers created and registered.")

	// 6. Start the Runner
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner.Start(ctx)                  // Pass context
	time.Sleep(100 * time.Millisecond) // Allow runner to start
	log.Println("Runner started.")

	// --- Execution ---
	// 1. Prepare the initial event
	initialEventID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	userRequest := "Research the recent developments in AI-powered code generation and summarize the key findings."
	eventData := agentflow.EventData{
		"user_request": userRequest,
	}
	eventMeta := map[string]string{
		agentflow.RouteMetadataKey: PlannerAgentName, // Start with the Planner
		agentflow.SessionIDKey:     initialEventID,   // Use SessionIDKey for tracking
	}
	// Use agentflow.NewEvent
	event := agentflow.NewEvent(PlannerAgentName, eventData, eventMeta)
	event.SetID(initialEventID) // Set specific ID if needed, though SessionID is preferred for tracking flow
	log.Printf("Initial event prepared: %s (Session: %s)", event.GetID(), event.GetSessionID())

	// 2. Emit the initial event
	log.Println("Emitting initial event to start the flow...")
	// wg.Add(1) // Add wait group count *if* FinalOutputHandler uses it reliably
	err = runner.Emit(event)
	requireNoError(err, "Emit initial event")
	log.Println("Initial event emitted.")

	// 3. Wait for the final result by checking the FinalOutputHandler
	log.Println("Waiting for final result...")
	var finalState agentflow.State
	var found bool
	timeout := time.After(180 * time.Second) // Timeout for the whole flow
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

waitLoop:
	for {
		select {
		case <-timeout:
			log.Fatal("Timeout waiting for the multi-agent flow to complete.")
			break waitLoop // Should not be reached due to log.Fatal
		case <-ticker.C:
			finalState, found = finalOutputHandler.GetFinalState(initialEventID)
			if found {
				log.Println("Final state found.")
				break waitLoop
			}
			log.Println("Still waiting for final state...") // Optional progress log
		case <-ctx.Done():
			log.Printf("Context cancelled while waiting: %v", ctx.Err())
			break waitLoop
		}
	}

	// 4. Stop the runner
	log.Println("Stopping runner...")
	cancel()      // Cancel context first
	runner.Stop() // Then stop runner
	log.Println("Runner stopped.")

	// --- Output ---
	if !found || finalState == nil {
		log.Fatal("Failed to get final state.")
	}

	// Check for errors first in the output state
	// FIX: Use finalState.Get(key) instead of GetMust
	if errMsg, ok := finalState.Get("error"); ok && errMsg != nil {
		log.Printf("Flow ended with error: %v", errMsg)
	} else if summary, ok := finalState.Get("summary"); ok && summary != nil {
		log.Println("Multi-agent flow successful.")
		fmt.Println("\n--- Final Summary ---")
		fmt.Printf("%v\n", summary) // Use %v for safety
		fmt.Println("---------------------")
	} else {
		log.Println("Flow ended, but no summary or known error found in final state.")
	}

	// FIX: Use finalState.Get(key) for logging
	finalSummary, _ := finalState.Get("summary")
	finalError, _ := finalState.Get("error")
	log.Printf("Final state summary: %v", finalSummary)
	log.Printf("Final state error: %v", finalError)
	// log.Printf("Final state data: %+v", finalState.GetData()) // Avoid this
}

// Helper for checking registration errors
func requireNoError(err error, step string) {
	if err != nil {
		log.Fatalf("Error during setup - %s: %v", step, err)
	}
}
