package integration_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/orchestrator"
	"kunalkushwaha/agentflow/internal/tools"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Setup ---

// Helper to get Azure Adapter configuration from environment variables
func getAzureTestConfig(t *testing.T) (llm.AzureOpenAIAdapterOptions, bool) {
	t.Helper() // Mark as test helper
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeploy := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	embedDeploy := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT") // Required by constructor

	if endpoint == "" || apiKey == "" || chatDeploy == "" || embedDeploy == "" {
		log.Println("Azure environment variables not fully set (AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, AZURE_OPENAI_CHAT_DEPLOYMENT, AZURE_OPENAI_EMBEDDING_DEPLOYMENT)")
		return llm.AzureOpenAIAdapterOptions{}, false
	}

	return llm.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeploy,
		EmbeddingDeployment: embedDeploy, // Pass embedding deployment
		HTTPClient:          &http.Client{Timeout: 90 * time.Second},
	}, true
}

// --- Test Handlers ---

// ComputeHandler calls a specific tool from the registry.
type ComputeHandler struct {
	registry   *tools.ToolRegistry
	toolName   string
	nextAgent  string                      // Agent/Handler to route to next
	cbRegistry *agentflow.CallbackRegistry // Use CallbackRegistry for signaling/emitting
	wg         *sync.WaitGroup
}

// Implement the correct Run method signature
func (h *ComputeHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	defer h.wg.Done()
	log.Printf("ComputeHandler: Running for event %s", event.GetID())

	// Use GetData() instead of GetPayload()
	args, ok := event.GetData()["tool_args"].(map[string]any)
	if !ok {
		err := fmt.Errorf("ComputeHandler: missing or invalid 'tool_args' in event data for event %s", event.GetID())
		log.Printf("%v", err)
		return agentflow.AgentResult{Error: err.Error()}, err
	}

	log.Printf("ComputeHandler: Calling tool '%s' with args: %v", h.toolName, args)
	toolResult, err := h.registry.CallTool(ctx, h.toolName, args) // Use context from Run
	if err != nil {
		err = fmt.Errorf("ComputeHandler: tool '%s' failed for event %s: %w", h.toolName, event.GetID(), err)
		log.Printf("%v", err)
		return agentflow.AgentResult{Error: err.Error()}, err
	}
	log.Printf("ComputeHandler: Tool '%s' successful. Result: %v", h.toolName, toolResult)

	// Prepare data for the next step
	nextStateData := agentflow.EventData{
		"tool_result": toolResult,
		// Pass original request if needed
		"original_request": event.GetData()["original_request"],
	}

	// Create output state containing the result and routing metadata
	outputState := agentflow.NewStateWithData(nextStateData)
	// Set metadata to indicate routing decision
	outputState.SetMeta(agentflow.RouteMetadataKey, h.nextAgent)
	outputState.SetMeta(agentflow.SessionIDKey, event.GetSessionID()) // Preserve session ID

	log.Printf("ComputeHandler: Finished. Returning state with routing key '%s'.", h.nextAgent)
	return agentflow.AgentResult{OutputState: outputState}, nil
}

// SummarizeHandler calls an LLM to summarize the tool result.
type SummarizeHandler struct {
	provider llm.ModelProvider
	wg       *sync.WaitGroup
	// Use a mutex and map to store results per event ID for assertion
	mu      sync.Mutex
	results map[string]string // Map event ID -> summary
}

// Implement the correct Run method signature
func (h *SummarizeHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	defer h.wg.Done()
	log.Printf("SummarizeHandler: Running for event %s", event.GetID())

	// Use GetData() instead of GetPayload()
	toolResultVal, ok := state.Get("tool_result") // Get from incoming state
	if !ok {
		err := fmt.Errorf("SummarizeHandler: missing 'tool_result' in state for event %s", event.GetID())
		log.Printf("%v", err)
		return agentflow.AgentResult{Error: err.Error()}, err
	}
	toolResult := fmt.Sprintf("%v", toolResultVal) // Keep as any for prompt flexibility

	originalRequestVal, _ := state.Get("original_request")
	originalRequest := fmt.Sprintf("%v", originalRequestVal)

	prompt := llm.Prompt{
		System: "You are an expert summarizer. Summarize the result of the calculation based on the original request.",
		User:   fmt.Sprintf("Original Request: %s\nCalculation Result: %s", originalRequest, toolResult),
	}

	log.Printf("SummarizeHandler: Calling LLM...")
	resp, err := h.provider.Call(ctx, prompt) // Use context passed into Run
	if err != nil {
		err = fmt.Errorf("SummarizeHandler: LLM call failed for event %s: %w", event.GetID(), err)
		log.Printf("%v", err)
		return agentflow.AgentResult{Error: err.Error()}, err
	}
	log.Printf("SummarizeHandler: LLM call successful. Summary: %s", resp.Content)

	// Store the result for assertion
	// Use original event ID from metadata if available, otherwise current event ID
	sessionID := event.GetSessionID()
	if sessionID == "" {
		sessionID = event.GetID() // Fallback if session ID isn't propagated
		log.Printf("SummarizeHandler: Warning - SessionID not found in metadata, using current event ID %s for result storage.", sessionID)
	}

	h.mu.Lock()
	if h.results == nil {
		h.results = make(map[string]string)
	}
	h.results[sessionID] = resp.Content
	h.mu.Unlock()

	// Return an empty result (no further state change needed)
	log.Printf("SummarizeHandler: Finished for event %s.", event.GetID())
	return agentflow.AgentResult{}, nil
}

// GetResult safely retrieves the summary for a given event ID.
func (h *SummarizeHandler) GetResult(sessionID string) (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.results == nil {
		return "", false
	}
	summary, found := h.results[sessionID]
	return summary, found
}

// --- Integration Test ---

// TestAgentWorkflowIntegration runs a simple 2-step workflow: Compute -> Summarize
func TestAgentWorkflowIntegration(t *testing.T) {
	// --- Test Setup ---
	azureOpts, configured := getAzureTestConfig(t)
	if !configured {
		t.Skip("Skipping integration test: Azure environment variables not set")
	}

	// 1. LLM Adapter
	azureAdapter, err := llm.NewAzureOpenAIAdapter(azureOpts)
	require.NoError(t, err, "Failed to create Azure adapter")

	// 2. Tool Registry
	toolRegistry := tools.NewToolRegistry()
	computeTool := &tools.ComputeMetricTool{}
	err = toolRegistry.Register(computeTool)
	require.NoError(t, err, "Failed to register compute tool")

	// 3. Orchestrator & Runner
	callbackRegistry := agentflow.NewCallbackRegistry()
	orchestratorImpl := orchestrator.NewRouteOrchestrator(callbackRegistry) // Use Route Orchestrator
	runner := agentflow.NewRunner(runtime.NumCPU())                         // Use NewRunner with queue size
	runner.SetOrchestrator(orchestratorImpl)
	runner.SetCallbackRegistry(callbackRegistry)
	// Optional: Add TraceLogger
	traceLogger := agentflow.NewInMemoryTraceLogger()
	runner.SetTraceLogger(traceLogger)

	// 4. Handlers
	var wg sync.WaitGroup
	computeHandler := &ComputeHandler{
		registry:   toolRegistry,
		toolName:   computeTool.Name(),
		nextAgent:  "summarizer", // Route to summarizer next
		cbRegistry: callbackRegistry,
		wg:         &wg,
	}
	summarizeHandler := &SummarizeHandler{
		provider: azureAdapter,
		wg:       &wg,
		results:  make(map[string]string),
	}

	// 5. Register Handlers with Orchestrator
	err = orchestratorImpl.RegisterAgent("compute", computeHandler) // Use orchestrator's RegisterAgent
	require.NoError(t, err)
	err = orchestratorImpl.RegisterAgent("summarizer", summarizeHandler)
	require.NoError(t, err)

	// 6. Start Runner
	ctx, cancel := context.WithCancel(context.Background()) // FIX: Use this ctx
	defer cancel()
	runner.Start(ctx)                 // FIX: Pass ctx to Start
	time.Sleep(50 * time.Millisecond) // Give runner time to start listening

	// --- Test Execution ---
	initialEventID := fmt.Sprintf("integ-test-%d", time.Now().UnixNano())
	initialEventData := agentflow.EventData{
		"tool_args": map[string]any{
			"operation": "add",
			"a":         15.5,
			"b":         4.5,
		},
		"original_request": "Calculate 15.5 + 4.5 and summarize the result.",
	}
	initialMetadata := map[string]string{
		agentflow.RouteMetadataKey: "compute", // Start with the compute handler
		agentflow.SessionIDKey:     initialEventID,
	}
	initialEvent := agentflow.NewEvent("compute", initialEventData, initialMetadata) // Target is compute
	initialEvent.SetID(initialEventID)                                               // Set specific ID for tracking

	// Add wait group count for the expected number of handler runs (compute + summarize)
	wg.Add(2)

	// Emit the initial event
	err = runner.Emit(initialEvent)
	require.NoError(t, err, "Failed to emit initial event")

	// Wait for handlers to complete or timeout
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		log.Println("All handlers completed.")
	case <-time.After(90 * time.Second): // Increased timeout
		t.Fatal("Timeout waiting for handlers to complete")
	}

	// --- Assertions ---
	// Check SummarizeHandler result
	summary, found := summarizeHandler.GetResult(initialEventID)
	require.True(t, found, "Summary result not found for session ID %s", initialEventID)
	assert.NotEmpty(t, summary, "Summary should not be empty")
	// Check if the summary contains the expected calculation result (20)
	assert.True(t, containsIgnoreCase(summary, "20"), "Summary '%s' does not contain the expected result '20'", summary)
	log.Printf("Final Summary: %s", summary)

	// Optional: Check trace logs
	traceEntries, traceErr := runner.DumpTrace(initialEventID)
	require.NoError(t, traceErr, "Failed to dump trace")
	// Expect more entries now due to hooks: EventStart, BeforeAgent(C), AfterAgent(C), BeforeAgent(S), AfterAgent(S), EventEnd
	require.GreaterOrEqual(t, len(traceEntries), 6, "Expected at least 6 trace entries")
	log.Printf("Trace contains %d entries.", len(traceEntries))
	// Add more specific trace assertions if needed

	// --- Cleanup ---
	runner.Stop() // Gracefully stop the runner
}

// Helper function for substring check (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// TODO: Remove or update TestAgentWorkflowIntegration if it's redundant or needs fixing
// func TestAgentWorkflowIntegration(t *testing.T) { ... }
