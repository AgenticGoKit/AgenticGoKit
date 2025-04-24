package integration_test // Changed from core_test

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
	registry  *tools.ToolRegistry
	toolName  string
	nextAgent string            // Agent/Handler to route to next
	runner    *agentflow.Runner // Needed to emit the next event
	wg        *sync.WaitGroup
}

func (h *ComputeHandler) Handle(event agentflow.Event) error {
	defer h.wg.Done()
	log.Printf("ComputeHandler: Handling event %s", event.GetID())

	payload, ok := event.GetPayload().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid payload type for ComputeHandler")
	}

	// Assume payload contains tool arguments under "tool_args" key
	toolArgs, ok := payload["tool_args"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing or invalid 'tool_args' in payload")
	}

	result, err := h.registry.CallTool(context.Background(), h.toolName, toolArgs)
	if err != nil {
		return fmt.Errorf("tool call failed: %w", err)
	}

	log.Printf("ComputeHandler: Tool '%s' executed successfully. Result: %v", h.toolName, result)

	// Prepare payload for the next step
	nextPayload := map[string]any{
		"tool_name":   h.toolName,
		"tool_args":   toolArgs, // Pass original args along
		"tool_result": result,
	}

	// Create and emit the next event, routing to the next agent
	nextEvent := &agentflow.SimpleEvent{
		ID:       fmt.Sprintf("%s-next", event.GetID()),
		Payload:  nextPayload,
		Metadata: event.GetMetadata(), // Preserve metadata
	}
	// Set routing hint for RouteOrchestrator
	nextEvent.Metadata[orchestrator.RouteMetadataKey] = h.nextAgent

	h.wg.Add(1) // Add to waitgroup for the next handler we are triggering
	h.runner.Emit(nextEvent)
	log.Printf("ComputeHandler: Emitted next event %s for agent %s", nextEvent.GetID(), h.nextAgent)

	return nil
}

// SummarizeHandler calls an LLM to summarize the tool result.
type SummarizeHandler struct {
	provider llm.ModelProvider
	wg       *sync.WaitGroup
	// Use a mutex and map to store results per event ID for assertion
	mu      sync.Mutex
	results map[string]string // Map event ID -> summary
}

func (h *SummarizeHandler) Handle(event agentflow.Event) error {
	defer h.wg.Done()
	log.Printf("SummarizeHandler: Handling event %s", event.GetID())

	payload, ok := event.GetPayload().(map[string]any)
	if !ok {
		return fmt.Errorf("invalid payload type for SummarizeHandler")
	}

	// Extract info needed for the summary prompt
	toolName, _ := payload["tool_name"].(string)
	toolResult, _ := payload["tool_result"].(map[string]any)

	promptText := fmt.Sprintf("The tool '%s' was called and returned the following result: %v. Briefly describe what happened.", toolName, toolResult)

	prompt := llm.Prompt{
		System: "You are an assistant that summarizes tool execution.",
		User:   promptText,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := h.provider.Call(ctx, prompt)
	if err != nil {
		return fmt.Errorf("llm call failed: %w", err)
	}

	log.Printf("SummarizeHandler: LLM call successful. Summary: %s", resp.Content)

	// Store the result for assertion
	h.mu.Lock()
	if h.results == nil {
		h.results = make(map[string]string)
	}
	// Use original event ID from metadata if available, otherwise current event ID
	originalEventID := event.GetID()
	if metaID, ok := event.GetMetadata()["original_event_id"]; ok {
		originalEventID = metaID
	} else {
		// If we came directly here, store under current ID
		log.Printf("SummarizeHandler: Warning - could not find original_event_id in metadata for event %s", event.GetID())
	}
	h.results[originalEventID] = resp.Content
	h.mu.Unlock()

	return nil
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
	if err != nil {
		t.Fatalf("Failed to create Azure adapter: %v", err)
	}

	// Create Tool Registry and register tool
	registry := tools.NewToolRegistry()
	computeTool := &tools.ComputeMetricTool{}
	if err := registry.Register(computeTool); err != nil {
		t.Fatalf("Failed to register compute tool: %v", err)
	}

	// Create Orchestrator and Runner
	orchestratorImpl := orchestrator.NewRouteOrchestrator()
	runner := agentflow.NewRunner(orchestratorImpl, runtime.NumCPU())

	// Create Handlers
	var wg sync.WaitGroup
	summarizeHandler := &SummarizeHandler{
		provider: azureAdapter,
		wg:       &wg,
		results:  make(map[string]string),
	}
	computeHandler := &ComputeHandler{
		registry:  registry,
		toolName:  computeTool.Name(),
		nextAgent: "summarizer", // Route to the summarizer next
		runner:    runner,       // Pass runner to emit next event
		wg:        &wg,
	}

	// Register Handlers
	runner.RegisterAgent("computer", computeHandler)
	runner.RegisterAgent("summarizer", summarizeHandler)

	// --- Test Execution ---
	initialEventID := fmt.Sprintf("int-test-%d", time.Now().UnixNano())
	initialPayload := map[string]any{
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
	initialEvent := &agentflow.SimpleEvent{
		ID:       initialEventID,
		Payload:  initialPayload,
		Metadata: initialMetadata,
	}

	log.Printf("Integration Test: Emitting initial event %s", initialEventID)
	wg.Add(1) // Expect the first handler ('computer') to be called
	runner.Emit(initialEvent)

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
	runner.Stop()

	// --- Assertions ---
	finalSummary, found := summarizeHandler.GetResult(initialEventID)
	if !found {
		t.Fatalf("Integration Test: Final summary not found for event %s", initialEventID)
	}

	t.Logf("Integration Test: Final Summary: %s", finalSummary)

	// Basic assertion: check if the summary mentions the result (15 + 27.5 = 42.5)
	// This is brittle, LLM output varies. A better test might check for keywords.
	if !contains(finalSummary, "42.5") && !contains(finalSummary, "42,5") { // Handle potential locale differences
		t.Errorf("Expected summary to contain the result '42.5', but got: %s", finalSummary)
	}
	if !contains(finalSummary, computeTool.Name()) {
		t.Errorf("Expected summary to contain the tool name '%s', but got: %s", computeTool.Name(), finalSummary)
	}
}

// Helper function for substring check (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// TODO: Add similar tests for OpenAIAdapter and OllamaAdapter if desired,
// potentially refactoring setup into helper functions.
