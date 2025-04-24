package main

import (
	"context"
	"encoding/json" // Added for parsing LLM response
	"fmt"
	"log"
	"net/http" // Added for LLM client
	"os"       // Added for env vars
	"runtime"
	"strings"
	"sync"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm" // Added LLM import
	"kunalkushwaha/agentflow/internal/orchestrator"
	"kunalkushwaha/agentflow/internal/tools"
)

// --- Tool Using Agent (implements agentflow.Agent) ---
type ToolAgent struct {
	registry *tools.ToolRegistry
	provider llm.ModelProvider // LLM provider to decide which tool/args
}

// LLMResponseFormat defines the expected JSON structure from the LLM.
type LLMResponseFormat struct {
	ToolName string         `json:"tool_name"` // Name of the tool to call, or empty/null if none
	Args     map[string]any `json:"args"`      // Arguments for the tool
}

func NewToolAgent(registry *tools.ToolRegistry, provider llm.ModelProvider) *ToolAgent {
	if registry == nil {
		log.Fatal("ToolAgent requires a non-nil ToolRegistry")
	}
	if provider == nil {
		log.Fatal("ToolAgent requires a non-nil ModelProvider")
	}
	return &ToolAgent{
		registry: registry,
		provider: provider,
	}
}

func (a *ToolAgent) Run(ctx context.Context, in agentflow.State) (agentflow.State, error) {
	log.Println("ToolAgent: Running...")
	requestVal, ok := in.Get("user_request")
	if !ok {
		return in, fmt.Errorf("ToolAgent: 'user_request' not found in input state")
	}
	request, ok := requestVal.(string)
	if !ok || request == "" {
		return in, fmt.Errorf("ToolAgent: 'user_request' is not a non-empty string")
	}
	log.Printf("ToolAgent: Received request: %q", request)

	// --- LLM-Based Tool Selection & Argument Extraction ---
	// 1. Define available tools for the LLM prompt
	// TODO: Dynamically generate this from registry and tool schemas
	availableToolsPrompt := `
Available Tools:
[
  {
    "name": "compute_metric",
    "description": "Performs simple arithmetic calculations like addition or subtraction on two numbers.",
    "parameters": {
      "type": "object",
      "properties": {
        "operation": {
          "type": "string",
          "description": "The operation to perform. Must be 'add' or 'subtract'."
        },
        "a": {
          "type": "number",
          "description": "The first number."
        },
        "b": {
          "type": "number",
          "description": "The second number."
        }
      },
      "required": ["operation", "a", "b"]
    }
  }
  // Add other tools here (e.g., web_search) if registered and needed
]
`

	// 2. Construct the prompt for the LLM
	systemPrompt := fmt.Sprintf(`You are an assistant that decides if a function call is needed to answer a user request.
You have access to the following tools:
%s
Analyze the user's request. If one of the tools can help fulfill the request, respond with a JSON object containing the 'tool_name' and the 'args' object for that tool based on its parameters schema.
Ensure argument values match the required types (string, number).
If no tool is suitable or needed for the request, respond with a JSON object where 'tool_name' is null or an empty string.
Respond ONLY with the JSON object and nothing else.`, availableToolsPrompt)

	userPrompt := request

	llmPrompt := llm.Prompt{
		System: systemPrompt,
		User:   userPrompt,
	}

	log.Println("ToolAgent: Asking LLM to decide on tool call...")
	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second) // Timeout for LLM call
	defer cancel()
	llmResp, err := a.provider.Call(llmCtx, llmPrompt)
	if err != nil {
		log.Printf("ToolAgent: LLM call failed: %v", err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("LLM call failed: %v", err))
		return out, nil // Return error in state
	}

	log.Printf("ToolAgent: LLM response received: %q", llmResp.Content)

	// 3. Parse the LLM's JSON response
	var decision LLMResponseFormat
	// Attempt to clean potential markdown code fences if the LLM adds them
	cleanedContent := strings.TrimSpace(llmResp.Content)
	cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
	cleanedContent = strings.TrimPrefix(cleanedContent, "```")
	cleanedContent = strings.TrimSuffix(cleanedContent, "```")
	cleanedContent = strings.TrimSpace(cleanedContent)

	if err := json.Unmarshal([]byte(cleanedContent), &decision); err != nil {
		log.Printf("ToolAgent: Failed to parse LLM JSON response '%s': %v", llmResp.Content, err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("Failed to parse LLM response: %v", err))
		out.Set("llm_raw_response", llmResp.Content) // Include raw response for debugging
		return out, nil                              // Return error in state
	}

	// 4. Check if the LLM decided to call a tool
	if decision.ToolName == "" {
		log.Println("ToolAgent: LLM decided no tool call is needed.")
		out := in.Clone()
		out.Set("agent_message", "No suitable tool found or needed for the request.")
		// Optionally, could make another LLM call here to answer directly
		return out, nil
	}

	toolName := decision.ToolName
	toolArgs := decision.Args // Args should already be in map[string]any format

	// --- End LLM-Based Section ---

	log.Printf("ToolAgent: LLM decided to call tool '%s' with args: %v", toolName, toolArgs)

	// 5. Call the selected tool via the registry
	// Use a separate context for the tool call if needed
	toolCtx, toolCancel := context.WithTimeout(ctx, 30*time.Second)
	defer toolCancel()
	toolResult, err := a.registry.CallTool(toolCtx, toolName, toolArgs)
	if err != nil {
		log.Printf("ToolAgent: Tool call failed: %v", err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("Tool call failed: %v", err))
		return out, nil
	}

	log.Printf("ToolAgent: Tool '%s' executed successfully. Result: %v", toolName, toolResult)

	// 6. Add results to state
	out := in.Clone()
	out.Set("tool_name_called", toolName)
	out.Set("tool_result", toolResult)
	// Optionally: Make a final LLM call here to synthesize the tool result into a user-friendly response
	log.Printf("ToolAgent: Added tool result to state.")
	return out, nil
}

// --- Agent Handler (Adapter: EventHandler -> Agent) ---
// (AgentHandler struct and NewAgentHandler function remain the same)
type AgentHandler struct {
	agent   agentflow.Agent
	results chan agentflow.State
	wg      *sync.WaitGroup
}

func NewAgentHandler(agent agentflow.Agent, wg *sync.WaitGroup) *AgentHandler {
	return &AgentHandler{
		agent:   agent,
		results: make(chan agentflow.State, 1),
		wg:      wg,
	}
}

// Handle implements the agentflow.EventHandler interface.
// (Handle function remains the same)
func (h *AgentHandler) Handle(event agentflow.Event) error {
	defer h.wg.Done() // Signal completion

	log.Printf("AgentHandler: Handling event %s", event.GetID())

	initialState := agentflow.NewState()
	payload, ok := event.GetPayload().(map[string]interface{})
	if !ok {
		log.Printf("AgentHandler: Error - Event payload is not map[string]interface{}")
		return fmt.Errorf("invalid event payload type")
	}
	// Expecting "user_request" in the event payload
	if request, ok := payload["user_request"].(string); ok {
		initialState.Set("user_request", request)
	} else {
		log.Printf("AgentHandler: Error - 'user_request' not found or not string in event payload")
		return fmt.Errorf("'user_request' missing or invalid in event payload")
	}
	for k, v := range event.GetMetadata() {
		initialState.SetMeta(k, v)
	}

	// Increase timeout slightly to accommodate LLM call within agent
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	finalState, err := h.agent.Run(ctx, initialState)
	if err != nil {
		log.Printf("AgentHandler: Agent run failed for event %s: %v", event.GetID(), err)
		// Don't send to results channel on error, just return the error
		return fmt.Errorf("agent run failed: %w", err)
	}

	log.Printf("AgentHandler: Agent run successful for event %s.", event.GetID())
	h.results <- finalState
	return nil
}

// --- Main Program ---

func main() {
	log.Println("Starting Tool Usage Example (LLM Driven)...")

	// --- Configuration (Azure OpenAI) ---
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT") // Still needed for adapter
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
		HTTPClient:          &http.Client{Timeout: 90 * time.Second},
	}
	azureAdapter, err := llm.NewAzureOpenAIAdapter(adapterOpts)
	if err != nil {
		log.Fatalf("Error creating Azure OpenAI Adapter: %v", err)
	}
	log.Println("Azure OpenAI Adapter created.")

	// 2. Create Tool Registry and Register Tools
	toolRegistry := tools.NewToolRegistry()
	computeTool := &tools.ComputeMetricTool{}
	if err := toolRegistry.Register(computeTool); err != nil {
		log.Fatalf("Failed to register compute tool: %v", err)
	}
	log.Println("Tool Registry created and tools registered.")

	// 3. Create Tool Agent, passing the LLM provider
	toolAgent := NewToolAgent(toolRegistry, azureAdapter)
	log.Println("Tool Agent created.")

	// 4. Create AgentHandler adapter
	var wg sync.WaitGroup
	agentHandler := NewAgentHandler(toolAgent, &wg)
	log.Println("Agent Handler created.")

	// 5. Create Orchestrator
	orchestratorImpl := orchestrator.NewRouteOrchestrator()
	log.Println("Route Orchestrator created.")

	// 6. Create Runner
	concurrency := runtime.NumCPU()
	runner := agentflow.NewRunner(orchestratorImpl, concurrency)
	log.Printf("Runner created with concurrency %d.", concurrency)

	// 7. Register the AgentHandler
	handlerName := "tool_agent_handler"
	runner.RegisterAgent(handlerName, agentHandler)
	log.Printf("Agent Handler registered with Runner as '%s'.", handlerName)

	// --- Execution ---
	// 1. Prepare the event
	eventPayload := map[string]interface{}{
		// Try different requests:
		"user_request": "What is 15 plus 27.5?", // Should trigger compute_metric
		// "user_request": "Can you subtract 10 from 50.5?",
		// "user_request": "What is the capital of Canada?",
	}
	eventMetadata := map[string]string{
		orchestrator.RouteMetadataKey: handlerName,
	}
	event := &agentflow.SimpleEvent{
		ID:       fmt.Sprintf("tool-evt-%d", time.Now().UnixNano()),
		Payload:  eventPayload,
		Metadata: eventMetadata,
	}
	log.Printf("Event prepared: %s", event.GetID())

	// 2. Emit the event
	log.Println("Emitting event via Runner...")
	wg.Add(1)
	runner.Emit(event)
	log.Println("Event emitted.")

	// 3. Wait for completion
	log.Println("Waiting for handler to complete...")
	done := make(chan struct{})
	var finalState agentflow.State
	var handlerErr error

	go func() {
		wg.Wait()
		select {
		case finalState = <-agentHandler.results:
			log.Println("Received final state from handler.")
		case <-time.After(2 * time.Second): // Increased timeout slightly
			log.Println("Timed out waiting for final state from handler.")
			handlerErr = fmt.Errorf("handler did not produce a result")
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("Handler processing finished.")
	case <-time.After(130 * time.Second): // Increased overall timeout
		log.Fatal("Timeout waiting for agent handler to complete.")
	}

	// 4. Stop runner
	log.Println("Stopping runner...")
	runner.Stop()
	log.Println("Runner stopped.")

	// --- Output ---
	if handlerErr != nil {
		log.Fatalf("Handler execution failed: %v", handlerErr)
	}
	if finalState.GetData() == nil {
		log.Fatal("Failed to get final state from handler.")
	}

	log.Println("Tool Agent execution successful.")
	if agentErr, ok := finalState.Get("agent_error"); ok {
		log.Printf("Agent reported an error: %v", agentErr)
		if rawLLMResp, ok := finalState.Get("llm_raw_response"); ok {
			log.Printf("LLM Raw Response was: %s", rawLLMResp)
		}
	} else if agentMsg, ok := finalState.Get("agent_message"); ok {
		log.Printf("Agent message: %v", agentMsg) // e.g., No tool needed
	} else {
		toolName, _ := finalState.Get("tool_name_called")
		toolResult, _ := finalState.Get("tool_result")
		fmt.Println("\n--- Agent Result ---")
		fmt.Printf("Tool Called: %v\n", toolName)
		fmt.Printf("Tool Result: %+v\n", toolResult)
		fmt.Println("--------------------")
	}
	log.Printf("Final state data: %+v", finalState.GetData())
}
