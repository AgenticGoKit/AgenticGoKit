package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync" // Added for WaitGroup
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/orchestrator" // Import concrete orchestrators
)

// --- Chat Agent (implements agentflow.Agent) ---
type ChatAgent struct {
	provider llm.ModelProvider
}

func NewChatAgent(provider llm.ModelProvider) *ChatAgent {
	if provider == nil {
		log.Fatal("ChatAgent requires a non-nil ModelProvider")
	}
	return &ChatAgent{provider: provider}
}

func (a *ChatAgent) Run(ctx context.Context, in agentflow.State) (agentflow.State, error) {
	log.Println("ChatAgent: Running...")
	userInputVal, ok := in.Get("user_prompt")
	if !ok {
		return in, fmt.Errorf("ChatAgent: 'user_prompt' not found in input state")
	}
	userInput, ok := userInputVal.(string)
	if !ok || userInput == "" {
		return in, fmt.Errorf("ChatAgent: 'user_prompt' is not a non-empty string")
	}
	log.Printf("ChatAgent: Received prompt: %q", userInput)

	prompt := llm.Prompt{
		System: "You are a concise assistant.",
		User:   userInput,
	}

	log.Println("ChatAgent: Calling LLM...")
	// Use a timeout specific to the LLM call if needed, derived from the main context
	callCtx, cancel := context.WithTimeout(ctx, 60*time.Second) // Example: 60s timeout for LLM
	defer cancel()
	resp, err := a.provider.Call(callCtx, prompt)
	if err != nil {
		log.Printf("ChatAgent: LLM call failed: %v", err)
		return in, fmt.Errorf("ChatAgent: LLM call failed: %w", err)
	}
	log.Printf("ChatAgent: LLM response received: FinishReason=%s", resp.FinishReason)

	out := in.Clone()
	out.Set("llm_response", resp.Content)
	log.Printf("ChatAgent: Added response to state.")
	return out, nil
}

// --- Agent Handler (Adapter: EventHandler -> Agent) ---
// This adapter allows an Agent (like ChatAgent) to be used where an EventHandler is expected.
type AgentHandler struct {
	agent agentflow.Agent
	// Add a channel or callback to signal completion and pass results if needed
	results chan agentflow.State // Simple channel to get the final state back
	wg      *sync.WaitGroup      // To signal when Handle is done
}

func NewAgentHandler(agent agentflow.Agent, wg *sync.WaitGroup) *AgentHandler {
	return &AgentHandler{
		agent:   agent,
		results: make(chan agentflow.State, 1), // Buffered channel
		wg:      wg,
	}
}

// Handle implements the agentflow.EventHandler interface.
func (h *AgentHandler) Handle(event agentflow.Event) error {
	defer h.wg.Done() // Signal completion when Handle finishes

	log.Printf("AgentHandler: Handling event %s", event.GetID())

	// 1. Extract necessary data from the event payload/metadata to create initial State
	initialState := agentflow.NewState()
	payload, ok := event.GetPayload().(map[string]interface{})
	if !ok {
		log.Printf("AgentHandler: Error - Event payload is not map[string]interface{}")
		return fmt.Errorf("invalid event payload type")
	}
	// Example: Assume payload contains "user_prompt"
	if prompt, ok := payload["user_prompt"].(string); ok {
		initialState.Set("user_prompt", prompt)
	} else {
		log.Printf("AgentHandler: Error - 'user_prompt' not found or not string in event payload")
		return fmt.Errorf("'user_prompt' missing or invalid in event payload")
	}
	// Copy metadata if needed
	for k, v := range event.GetMetadata() {
		initialState.SetMeta(k, v)
	}

	// 2. Run the underlying agent
	// Use a background context for the agent run, potentially with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second) // Timeout for the agent's Run
	defer cancel()

	finalState, err := h.agent.Run(ctx, initialState)
	if err != nil {
		log.Printf("AgentHandler: Agent run failed for event %s: %v", event.GetID(), err)
		// Don't send to results channel on error, just return the error
		return fmt.Errorf("agent run failed: %w", err)
	}

	log.Printf("AgentHandler: Agent run successful for event %s.", event.GetID())
	// 3. Send the final state back via the channel
	h.results <- finalState
	return nil
}

// --- Main Program ---

func main() {
	log.Println("Starting Orchestrator Example...")

	// --- Configuration ---
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT") // Still needed for adapter creation
	if embeddingDeployment == "" {
		embeddingDeployment = "not-used-in-this-example"
		log.Println("Warning: AZURE_OPENAI_EMBEDDING_DEPLOYMENT not set, using dummy value.")
	}
	if endpoint == "" || apiKey == "" || chatDeployment == "" {
		log.Fatal("Error: Set AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, and AZURE_OPENAI_CHAT_DEPLOYMENT environment variables.")
	}

	// --- Setup ---
	// 1. Create LLM Adapter
	adapterOpts := llm.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment,
		HTTPClient:          &http.Client{Timeout: 120 * time.Second}, // Longer timeout for the client overall
	}
	azureAdapter, err := llm.NewAzureOpenAIAdapter(adapterOpts)
	if err != nil {
		log.Fatalf("Error creating Azure OpenAI Adapter: %v", err)
	}
	log.Println("Azure OpenAI Adapter created.")

	// 2. Create Chat Agent
	chatAgent := NewChatAgent(azureAdapter)
	log.Println("Chat Agent created.")

	// 3. Create the AgentHandler adapter
	var wg sync.WaitGroup // WaitGroup to wait for the handler to finish
	agentHandler := NewAgentHandler(chatAgent, &wg)
	log.Println("Agent Handler created.")

	// 4. Create a CONCRETE Orchestrator instance (e.g., RouteOrchestrator)
	// orchestratorImpl := orchestrator.NewCollaborateOrchestrator() // Or Collaborate
	orchestratorImpl := orchestrator.NewRouteOrchestrator()
	log.Println("Route Orchestrator created.")

	// 5. Create Runner, passing the concrete orchestrator
	concurrency := runtime.NumCPU()
	runner := agentflow.NewRunner(orchestratorImpl, concurrency) // Pass the concrete implementation
	log.Printf("Runner created with concurrency %d.", concurrency)

	// 6. Register the AgentHandler with the Runner (which delegates to the orchestrator)
	runner.RegisterAgent("chat_handler", agentHandler)
	log.Println("Agent Handler registered with Runner.")

	// --- Execution ---
	// 1. Prepare the event
	eventPayload := map[string]interface{}{
		"user_prompt": "What is the capital of France?",
	}
	eventMetadata := map[string]string{
		"correlation_id": fmt.Sprintf("req-%d", time.Now().UnixNano()),
	}
	event := &agentflow.SimpleEvent{
		ID:       fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Payload:  eventPayload,
		Metadata: eventMetadata,
	}
	log.Printf("Event prepared: %s", event.GetID())

	// 2. Emit the event via the Runner
	log.Println("Emitting event via Runner...")
	wg.Add(1) // Expect one call to Handle for this event
	runner.Emit(event)
	log.Println("Event emitted.")

	// 3. Wait for the handler to complete processing
	log.Println("Waiting for handler to complete...")
	// Wait for Handle to finish AND result to be available
	done := make(chan struct{})
	var finalState agentflow.State
	var handlerErr error // To capture potential error from Handle itself

	go func() {
		// This assumes Handle will always call wg.Done()
		wg.Wait() // Wait for Handle goroutine(s) to finish
		// Try to receive the result, but don't block forever if Handle errored
		select {
		case finalState = <-agentHandler.results:
			log.Println("Received final state from handler.")
		case <-time.After(1 * time.Second): // Timeout if no result (maybe Handle errored)
			log.Println("Timed out waiting for final state from handler (Handle might have errored).")
			// Check if Handle returned an error (this requires modifying AgentHandler.Handle to store the error)
			// For now, we assume timeout means error or no result sent.
			handlerErr = fmt.Errorf("handler did not produce a result")
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("Handler processing finished.")
	case <-time.After(110 * time.Second): // Overall timeout slightly longer than agent timeout
		log.Fatal("Timeout waiting for agent handler to complete.")
	}

	// 4. Stop the runner gracefully
	log.Println("Stopping runner...")
	runner.Stop() // Waits for queue to empty and loop to finish
	log.Println("Runner stopped.")

	// --- Output ---
	if handlerErr != nil {
		log.Fatalf("Handler execution failed: %v", handlerErr)
	}
	if finalState.GetData() == nil { // Check if state was actually received
		log.Fatal("Failed to get final state from handler.")
	}

	log.Println("Orchestrator execution successful (via Runner/EventHandler).")
	llmResponse, ok := finalState.Get("llm_response")
	if !ok {
		log.Println("LLM response not found in final state.")
	} else {
		fmt.Println("\n--- LLM Response ---")
		fmt.Printf("%v\n", llmResponse)
		fmt.Println("--------------------")
	}
	log.Printf("Final state data: %+v", finalState.GetData())
	log.Printf("Final state metadata: %+v", finalState.GetMetadata())
}
