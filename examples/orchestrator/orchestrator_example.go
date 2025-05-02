package main

import (
	"context" // Import errors package
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
)

// Define constants matching those in the runner
const (
	callbackStateKeyAgentResult = "__agentResult"
	callbackStateKeyAgentError  = "__agentError"
)

// --- Callback Function to Print Response ---
// FIX: Update signature to match agentflow.CallbackFunc
func printLLMResponseCallback(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
	// FIX: Extract result and error from args.AgentResult and args.Error
	agentResult := args.AgentResult
	agentErr := args.Error // The error from the agent run is directly in args.Error for AfterAgentRun/AgentError hooks

	// Use the extracted values (original logic)
	// Check if the hook *should* have been AfterAgentRun conceptually
	// Note: The hook type isn't passed here, so we rely on registration context.
	if args.Hook == agentflow.HookAfterAgentRun { // Check the hook type from args
		if agentErr != nil {
			log.Printf("Callback: Agent '%s' finished with error: %v", args.AgentID, agentErr)
		} else {
			// Assuming the LLM response is in the OutputState of the AgentResult
			// You might need to adjust the key based on how ChatAgent structures its output
			llmResponse := "N/A"
			if agentResult.OutputState != nil {
				if respVal, ok := agentResult.OutputState.Get("llm_response"); ok {
					llmResponse, _ = respVal.(string)
				}
			}
			log.Printf("Callback: Agent '%s' finished successfully. Response snippet: %.50s...", args.AgentID, llmResponse)
			log.Printf("Callback: Agent '%s' took %s", args.AgentID, agentResult.Duration)
		}
	} else {
		log.Printf("Callback: Hook '%s' triggered for agent '%s' (Error: %v)", args.Hook, args.AgentID, agentErr)
	}

	// This callback doesn't modify state, just observes.
	// Return the *original* state received (args.State), or nil to indicate no change.
	return nil, nil // Indicate no state change intended by this observer callback
}

// --- Chat Agent (implements agentflow.AgentHandler) ---
type ChatAgent struct {
	provider llm.ModelProvider
	wg       *sync.WaitGroup // Optional: For tracking completion
}

func NewChatAgent(provider llm.ModelProvider, wg *sync.WaitGroup) *ChatAgent {
	if provider == nil {
		log.Fatal("ChatAgent requires a non-nil ModelProvider")
	}
	return &ChatAgent{provider: provider, wg: wg}
}

// Run implements the AgentHandler interface.
func (a *ChatAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	startTime := time.Now()
	if a.wg != nil {
		defer a.wg.Done() // Signal completion if WaitGroup is used
	}
	log.Printf("ChatAgent: Running for event %s...", event.GetID())

	eventData := event.GetData()
	userInputVal, ok := eventData["user_prompt"]
	if !ok {
		err := fmt.Errorf("missing 'user_prompt' in event data")
		return agentflow.AgentResult{
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   time.Now(),
			Duration:  time.Since(startTime),
		}, err
	}
	userInput, ok := userInputVal.(string)
	if !ok || userInput == "" {
		err := fmt.Errorf("'user_prompt' must be a non-empty string")
		return agentflow.AgentResult{
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   time.Now(),
			Duration:  time.Since(startTime),
		}, err
	}

	prompt := llm.Prompt{
		System: "You are a helpful assistant.",
		User:   userInput,
	}

	// FIX: Log the prompt being sent
	log.Printf("ChatAgent: Sending prompt to LLM: System='%s', User='%s'", prompt.System, prompt.User)
	log.Println("ChatAgent: Calling LLM...")
	resp, err := a.provider.Call(ctx, prompt)
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	if err != nil {
		log.Printf("ChatAgent: LLM call failed: %v", err)
		return agentflow.AgentResult{
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  duration,
		}, err
	}

	// FIX: Log the full response received
	log.Printf("ChatAgent: LLM response received: FinishReason=%s", resp.FinishReason)
	log.Printf("ChatAgent: LLM Full Response: %s", resp.Content) // Log the full content

	// Create a new state for the output
	outState := agentflow.NewState() // Returns *SimpleState
	outState.Set("llm_response", resp.Content)
	// Copy metadata from input state if needed
	// for _, key := range state.MetaKeys() {
	// 	if val, ok := state.GetMeta(key); ok {
	// 		outState.SetMeta(key, val)
	// 	}
	// }

	// Return the state directly (concrete *SimpleState satisfies State interface)
	return agentflow.AgentResult{
		OutputState: outState,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    duration,
	}, nil
}

// --- Main Program ---

func main() {
	log.Println("Starting Orchestrator Example...")

	// --- Configuration ---
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT") // Required by adapter
	if embeddingDeployment == "" {
		embeddingDeployment = "dummy-embed" // Provide a dummy if not set
		log.Println("Warning: AZURE_OPENAI_EMBEDDING_DEPLOYMENT not set, using dummy value.")
	}
	if endpoint == "" || apiKey == "" || chatDeployment == "" {
		log.Fatal("Error: AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, and AZURE_OPENAI_CHAT_DEPLOYMENT environment variables must be set.")
	}

	// --- Setup ---
	// 1. Create LLM Adapter
	adapterOpts := llm.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment,
		HTTPClient:          &http.Client{Timeout: 90 * time.Second},
	}
	azureAdapter, err := llm.NewAzureOpenAIAdapter(adapterOpts)
	if err != nil {
		log.Fatalf("Failed to create Azure OpenAI Adapter: %v", err)
	}
	log.Println("Azure OpenAI Adapter created.")

	// 2. Create Core Components
	callbackRegistry := agentflow.NewCallbackRegistry()
	// Register with the standard signature
	// FIX: Ensure the hook string matches where the Runner invokes it
	err = callbackRegistry.Register(agentflow.HookAfterAgentRun, "printResponse", printLLMResponseCallback) // Register for AfterAgentRun
	if err != nil {
		log.Fatalf("Failed to register callback: %v", err)
	}
	log.Println("Callback registered.")

	// Use RouteOrchestrator which requires the registry
	orchestratorImpl := orchestrator.NewRouteOrchestrator(callbackRegistry)
	concurrency := runtime.NumCPU()
	// Create runner and link orchestrator
	runner := agentflow.NewRunner(concurrency)
	runner.SetOrchestrator(orchestratorImpl)     // Set orchestrator after creation
	runner.SetCallbackRegistry(callbackRegistry) // Also set registry on runner
	log.Printf("Runner & Route Orchestrator created (concurrency %d).", concurrency)

	// 3. Create Chat Agent (which is the AgentHandler)
	var wg sync.WaitGroup
	chatAgentHandler := NewChatAgent(azureAdapter, &wg)
	log.Println("Chat Agent Handler created.")

	// 4. Register the ChatAgent handler with the orchestrator
	// RouteOrchestrator uses RegisterAgent
	err = orchestratorImpl.RegisterAgent("chat", chatAgentHandler)
	if err != nil {
		log.Fatalf("Failed to register agent handler: %v", err)
	}
	log.Println("Chat Agent Handler registered with orchestrator.")

	// 5. Start the Runner
	ctx, cancel := context.WithCancel(context.Background()) // Ensure context is cancelled on main exit
	defer cancel()

	// FIX: Pass ctx to Start
	runner.Start(ctx) // Start the runner's loop
	log.Println("Runner started.")
	time.Sleep(100 * time.Millisecond) // Small pause to ensure runner loop starts listening

	// --- Execution ---
	// 1. Prepare the event
	// FIX: Remove unused eventID
	// eventID := fmt.Sprintf("chat-%d", time.Now().UnixNano())
	eventData := agentflow.EventData{"user_prompt": "What is the capital of France?"}
	eventMeta := map[string]string{agentflow.RouteMetadataKey: "chat"} // Target the 'chat' handler
	event := agentflow.NewEvent(eventMeta[agentflow.RouteMetadataKey], eventData, eventMeta)
	log.Printf("Prepared event %s targeting agent 'chat'.", event.GetID())

	// 2. Emit the event via the Runner
	wg.Add(1) // Expect one call to ChatAgent.Run for this event
	err = runner.Emit(event)
	if err != nil {
		log.Fatalf("Failed to emit event: %v", err)
	}
	log.Printf("Event %s emitted.", event.GetID())

	// 3. Wait for the handler (ChatAgent.Run) and the runner to complete processing
	waitTimeout := 100 * time.Second // Add a timeout for waiting
	waitChan := make(chan struct{})
	go func() {
		wg.Wait() // Wait for the handler to finish
		close(waitChan)
	}()

	select {
	case <-waitChan:
		log.Println("Chat agent handler completed.")
	case <-time.After(waitTimeout):
		log.Printf("Timeout waiting for agent handler after %v.", waitTimeout)
		cancel() // Cancel context on timeout
	case <-ctx.Done():
		log.Printf("Context cancelled while waiting for agent handler: %v", ctx.Err())
	}

	// 4. Stop the runner gracefully (already initiated by context cancellation if timeout didn't occur)
	log.Println("Stopping runner...")
	cancel() // Ensure context is cancelled
	runner.Stop()
	log.Println("Runner stopped.")

	// Final wait to ensure everything cleans up, especially if timeout happened
	// time.Sleep(500 * time.Millisecond)

	// --- Output ---
	// The response is now printed by the callback function when the agent finishes.
	log.Println("Execution finished. Check logs above for details.")
}
