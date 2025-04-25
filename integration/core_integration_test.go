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
	t.Helper()
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	// Embedding deployment not strictly needed for this test, but constructor requires it
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")
	if embeddingDeployment == "" {
		embeddingDeployment = "not-used-in-this-test"
	}

	if endpoint == "" || apiKey == "" || chatDeployment == "" {
		return llm.AzureOpenAIAdapterOptions{}, false // Skip if vars not set
	}

	return llm.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment, // Required by constructor
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
	if h.wg != nil {
		defer h.wg.Done() // Ensure WaitGroup is decremented
	}
	log.Printf("ComputeHandler: Handling event %s", event.GetID())

	// Use GetData() instead of GetPayload()
	payload, ok := event.GetData()["tool_args"].(map[string]any)
	if !ok {
		err := fmt.Errorf("missing or invalid 'tool_args' in event data for ComputeHandler")
		log.Printf("ERROR: %v - Event Data: %v", err, event.GetData())
		return agentflow.AgentResult{}, err
	}

	result, err := h.registry.CallTool(ctx, h.toolName, payload)
	if err != nil {
		log.Printf("ERROR: ComputeHandler tool call failed: %v", err)
		return agentflow.AgentResult{}, fmt.Errorf("tool call failed: %w", err)
	}

	log.Printf("ComputeHandler: Tool '%s' executed successfully. Result: %v", h.toolName, result)

	// Prepare data for the next step
	nextData := agentflow.EventData{
		"tool_name":   h.toolName,
		"tool_args":   payload, // Pass original args along
		"tool_result": result,
	}

	// Create the next event, routing to the next agent via metadata
	nextEvent := agentflow.NewEvent(
		"", // No specific target ID
		nextData,
		map[string]string{
			orchestrator.RouteMetadataKey: h.nextAgent,   // Set routing hint
			"original_event_id":           event.GetID(), // Pass original ID if needed by next step
		},
	)
	nextEvent.SetSourceAgentID("computer") // Identify the source

	// Use CallbackRegistry to signal completion/emit next event
	if h.cbRegistry != nil {
		if h.wg != nil {
			h.wg.Add(1) // Add to waitgroup for the next handler we are triggering
		}
		// FIX: Use CallbackArgs struct for Invoke
		args := agentflow.CallbackArgs{
			Ctx:   ctx,                         // Pass the current context
			Hook:  agentflow.HookAfterAgentRun, // Use an appropriate hook, e.g., after agent run
			Event: nextEvent,                   // Pass the event to be potentially handled by callbacks
			// State: state, // Optionally pass the current state if needed by callbacks
		}
		h.cbRegistry.Invoke(args)
		log.Printf("ComputeHandler: Invoked hook '%s' with event %s for agent %s", args.Hook, nextEvent.GetID(), h.nextAgent)
	} else {
		log.Printf("ComputeHandler: Warning - CallbackRegistry is nil, cannot trigger next event.")
	}

	// Return an empty result for now, assuming state isn't modified here
	return agentflow.AgentResult{}, nil
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
	if h.wg != nil {
		defer h.wg.Done() // Ensure WaitGroup is decremented
	}
	log.Printf("SummarizeHandler: Handling event %s", event.GetID())

	// Use GetData() instead of GetPayload()
	data := event.GetData()
	toolName, _ := data["tool_name"].(string)
	toolResult, _ := data["tool_result"] // Keep as any for prompt flexibility

	promptText := fmt.Sprintf("The tool '%s' was called and returned the following result: %v. Briefly describe what happened.", toolName, toolResult)

	prompt := llm.Prompt{
		System: "You are an assistant that summarizes tool execution.",
		User:   promptText,
	}

	// Use context passed into Run
	resp, err := h.provider.Call(ctx, prompt)
	if err != nil {
		log.Printf("ERROR: SummarizeHandler LLM call failed: %v", err)
		return agentflow.AgentResult{}, fmt.Errorf("llm call failed: %w", err)
	}

	log.Printf("SummarizeHandler: LLM call successful. Summary: %s", resp.Content)

	// Store the result for assertion
	h.mu.Lock()
	if h.results == nil {
		h.results = make(map[string]string)
	}
	// Use original event ID from metadata if available, otherwise current event ID
	originalEventID := event.GetID()
	if metaID, ok := event.GetMetadata()["original_event_id"]; ok && metaID != "" {
		originalEventID = metaID
	} else {
		log.Printf("SummarizeHandler: Warning - could not find original_event_id in metadata for event %s", event.GetID())
	}
	h.results[originalEventID] = resp.Content
	h.mu.Unlock()

	// Return an empty result
	return agentflow.AgentResult{}, nil
}

func (h *SummarizeHandler) GetResult(eventID string) (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	res, ok := h.results[eventID]
	return res, ok
}

// --- Integration Test ---

func TestIntegration_Azure_ComputeAndSummarize(t *testing.T) {
	// --- Test Setup ---
	opts, configured := getAzureTestConfig(t)
	if !configured {
		t.Skip("Skipping Azure integration test: AZURE_OPENAI_* environment variables not set")
	}

	azureAdapter, err := llm.NewAzureOpenAIAdapter(opts)
	require.NoError(t, err, "Failed to create Azure adapter")

	// Create Tool Registry and register tool
	toolRegistry := tools.NewToolRegistry()
	computeTool := &tools.ComputeMetricTool{}
	err = toolRegistry.Register(computeTool)
	require.NoError(t, err, "Failed to register compute tool")

	// Create Core Components
	callbackRegistry := agentflow.NewCallbackRegistry()
	// Pass registry to orchestrator
	orchestratorImpl := orchestrator.NewRouteOrchestrator(callbackRegistry)
	// Pass only worker count to runner
	runner := agentflow.NewRunner(runtime.NumCPU())

	// Create Handlers
	var wg sync.WaitGroup
	summarizeHandler := &SummarizeHandler{
		provider: azureAdapter,
		wg:       &wg,
		results:  make(map[string]string),
	}
	computeHandler := &ComputeHandler{
		registry:   toolRegistry,
		toolName:   computeTool.Name(),
		nextAgent:  "summarizer",     // Route to the summarizer next
		cbRegistry: callbackRegistry, // Pass callback registry
		wg:         &wg,
	}

	// Register Handlers with the Orchestrator
	orchestratorImpl.RegisterAgent("computer", computeHandler)
	orchestratorImpl.RegisterAgent("summarizer", summarizeHandler)

	// --- Test Execution ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the runner - Assuming it takes the registry to know about callbacks
	wg.Add(1) // Add one for the runner goroutine itself
	go func() {
		defer wg.Done()
		log.Println("Runner starting...")
		// FIX: runner.Start only takes context
		if startErr := runner.Start(ctx); startErr != nil && startErr != context.Canceled {
			t.Errorf("Runner failed: %v", startErr) // Use t.Errorf for goroutine
		}
		log.Println("Runner stopped.")
	}()
	time.Sleep(100 * time.Millisecond) // Give runner time to start

	initialEventID := fmt.Sprintf("int-test-%d", time.Now().UnixNano())
	// Use EventData type
	initialData := agentflow.EventData{
		"tool_args": map[string]any{
			"operation": "add",
			"a":         15,
			"b":         27.5,
		},
	}
	initialMetadata := map[string]string{
		orchestrator.RouteMetadataKey: "computer",     // Start with the computer handler
		"original_event_id":           initialEventID, // Store original ID for result lookup
	}
	// Use NewEvent constructor
	initialEvent := agentflow.NewEvent(
		"", // No target ID, use metadata
		initialData,
		initialMetadata,
	)
	initialEvent.SetID(initialEventID) // Override generated ID for test predictability

	log.Printf("Integration Test: Dispatching initial event %s", initialEventID)
	wg.Add(1) // Expect the first handler ('computer') to be called
	// Dispatch via orchestrator
	dispatchErr := orchestratorImpl.Dispatch(initialEvent)
	require.NoError(t, dispatchErr, "Failed to dispatch initial event")

	// Wait for all handlers triggered by the initial event and subsequent events
	waitTimeout := 100 * time.Second // Adjust timeout as needed
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Integration Test: WaitGroup finished.")
	case <-time.After(waitTimeout):
		t.Fatalf("Integration Test: Timeout waiting for handlers to complete")
	}

	// Stop the runner
	cancel() // Signal runner to stop
	// Wait for the runner goroutine to exit (already covered by wg.Wait() in select)

	// --- Assertions ---
	finalSummary, found := summarizeHandler.GetResult(initialEventID)
	require.True(t, found, "Integration Test: Final summary not found for event %s", initialEventID)

	t.Logf("Integration Test: Final Summary: %s", finalSummary)

	// Basic assertion: check if the summary mentions the result (15 + 27.5 = 42.5)
	assert.True(t, contains(finalSummary, "42.5") || contains(finalSummary, "42,5"), "Expected summary to contain '42.5', got: %s", finalSummary)
	assert.True(t, contains(finalSummary, computeTool.Name()), "Expected summary to contain tool name '%s', got: %s", computeTool.Name(), finalSummary)
}

// Helper function for substring check (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// TODO: Remove or update TestAgentWorkflowIntegration if it's redundant or needs fixing
// func TestAgentWorkflowIntegration(t *testing.T) { ... }
