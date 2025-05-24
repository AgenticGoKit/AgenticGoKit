package main

import (
	"context"
	"encoding/json" // Added for parsing LLM response
	"fmt"
	"log"
	"net/http" // Added for LLM client
	"os"       // Added for env vars
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	agentflow "github.com/kunalkushwaha/agentflow/core"
	"github.com/kunalkushwaha/agentflow/internal/llm" // Added LLM import
	"github.com/kunalkushwaha/agentflow/internal/tools"
)

// --- Tool Using Agent (implements agentflow.Agent) ---
type ToolAgent struct {
	registry *tools.ToolRegistry
	provider llm.ModelProvider // LLM provider to decide which tool/args
}

// Name allows ToolAgent to satisfy agentflow.Agent.
func (a *ToolAgent) Name() string { return "tool_agent" }

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
	startTime := time.Now()
	requestID := fmt.Sprintf("azure-req-%d", time.Now().UnixNano())

	// Add request ID to all logs for correlation
	log.Printf("ToolAgent: Starting request %s", requestID)

	// Standard Azure execution telemetry pattern
	defer func() {
		duration := time.Since(startTime)
		log.Printf("ToolAgent: Request %s completed in %v", requestID, duration)
	}()

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
	// 1. Define available tools for the LLM prompt by querying registry
	availableTools := a.registry.List()
	toolDefinitions := make([]string, 0, len(availableTools))

	// Convert registered tools to OpenAI-compatible function definitions
	for _, toolName := range availableTools {
		_, exists := a.registry.Get(toolName) // we don’t need the value itself
		if !exists {
			continue
		}

		// For compute_metric tool
		if toolName == "compute_metric" {
			toolDefinitions = append(toolDefinitions, `{
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
  }`)
		}
		// Add other tool definitions here as needed
	}

	availableToolsPrompt := fmt.Sprintf(`
Available Tools:
[
  %s
]
`, strings.Join(toolDefinitions, ",\n  "))

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
	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Add Azure-style retry logic with exponential backoff
	var llmResp llm.Response
	var err error
	retryMax := 3
	retryDelay := time.Second

	for attempt := 0; attempt < retryMax; attempt++ {
		if attempt > 0 {
			log.Printf("ToolAgent: Retrying LLM call (attempt %d/%d) after error: %v",
				attempt+1, retryMax, err)
			// Exponential backoff following Azure best practices
			time.Sleep(retryDelay * time.Duration(1<<uint(attempt)))
		}

		llmResp, err = a.provider.Call(llmCtx, llmPrompt)
		if err == nil {
			break
		}

		// Check for permanent errors that shouldn't be retried
		if strings.Contains(err.Error(), "authentication") ||
			strings.Contains(err.Error(), "invalid_request") {
			log.Printf("ToolAgent: Permanent Azure OpenAI error, not retrying: %v", err)
			break
		}
	}

	if err != nil {
		log.Printf("ToolAgent: Azure OpenAI call failed after %d attempts: %v", retryMax, err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("Azure OpenAI call failed: %v", err))
		return out, nil
	}

	log.Printf("ToolAgent: LLM response received: %q", llmResp.Content)

	// 3. Parse the LLM's JSON response with Azure-optimized error handling
	var decision LLMResponseFormat

	// Azure-optimized JSON response cleaning
	cleanedContent := strings.TrimSpace(llmResp.Content)

	// Remove JSON code fences if present
	jsonStartPos := strings.Index(cleanedContent, "{")
	jsonEndPos := strings.LastIndex(cleanedContent, "}")

	if jsonStartPos >= 0 && jsonEndPos > jsonStartPos {
		// Extract only the JSON object
		cleanedContent = cleanedContent[jsonStartPos : jsonEndPos+1]
	} else {
		// Try markdown code block extraction if JSON object not found
		cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
		cleanedContent = strings.TrimPrefix(cleanedContent, "```")
		cleanedContent = strings.TrimSuffix(cleanedContent, "```")
		cleanedContent = strings.TrimSpace(cleanedContent)
	}

	// Azure-compliant telemetry
	log.Printf("ToolAgent: Cleaned JSON for parsing: %s", cleanedContent)

	if err := json.Unmarshal([]byte(cleanedContent), &decision); err != nil {
		// Azure-style structured error logging
		log.Printf("ToolAgent: Failed to parse Azure OpenAI response: %v", err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("Failed to parse Azure OpenAI response: %v", err))
		out.Set("llm_raw_response", llmResp.Content)
		return out, nil
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

	// Azure best practice: Add explicit validation before tool call
	if toolName == "" {
		log.Printf("ToolAgent: Tool name is empty, cannot proceed with call")
		out := in.Clone()
		out.Set("agent_error", "Invalid tool name: empty string")
		return out, nil
	}

	// Azure best practice: Validate tool exists before attempting call
	if _, exists := a.registry.Get(toolName); !exists {
		log.Printf("ToolAgent: Tool '%s' not found in registry", toolName)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("Tool '%s' not found", toolName))
		return out, nil
	}

	// Azure best practice: Structured logging before call
	log.Printf("ToolAgent: Calling Azure-registered tool '%s' with parameters: %+v", toolName, toolArgs)

	toolResult, err := a.registry.CallTool(toolCtx, toolName, toolArgs)
	if err != nil {
		// Azure-style detailed error categorization
		errorType := "tool_execution_error"
		if strings.Contains(err.Error(), "not found") {
			errorType = "tool_not_found"
		} else if strings.Contains(err.Error(), "permission") {
			errorType = "tool_permission_denied"
		}

		log.Printf("ToolAgent: Tool call failed [%s]: %v", errorType, err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("Tool call failed: %v", err))
		out.Set("error_type", errorType)
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
type AgentHandler struct {
	agent agentflow.Agent
	wg    *sync.WaitGroup
}

// Ensure at compile‑time that AgentHandler implements agentflow.AgentHandler.
var _ agentflow.AgentHandler = (*AgentHandler)(nil)

func NewAgentHandler(agent agentflow.Agent, wg *sync.WaitGroup) *AgentHandler {
	return &AgentHandler{agent: agent, wg: wg}
}

// Run adapts the inner agent to the AgentHandler signature.
func (h *AgentHandler) Run(
	ctx context.Context,
	event agentflow.Event,
	state agentflow.State,
) (agentflow.AgentResult, error) {

	defer h.wg.Done()

	log.Printf("AgentHandler: Handling event %s", event.GetID())

	// Ensure we have a mutable state value
	if state == nil {
		state = agentflow.NewState()
	}

	// Copy payload → state
	pMap := event.GetData() // EventData == map[string]any
	if v, ok := pMap["user_request"].(string); ok {
		state.Set("user_request", v)
	}
	for k, v := range event.GetMetadata() {
		state.SetMeta(k, v)
	}

	_, err := h.agent.Run(ctx, state)   // ignore returned state
	return agentflow.AgentResult{}, err // keep signature happy
}

// --- Main Program ---

func main() {
	// Azure best practice: Configure context with proper cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure all resources are cleaned up

	// Azure best practice: Set up proper signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received termination signal, cleaning up Azure resources...")
		cancel()
		os.Exit(0)
	}()

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
	if err := toolRegistry.Register(&tools.ComputeMetricTool{}); err != nil {
		log.Fatalf("tool register error: %v", err)
	}
	// 3. Create the ToolAgent
	toolAgent := NewToolAgent(toolRegistry, azureAdapter)

	// 4. Wrap with AgentHandler
	var wg sync.WaitGroup
	agentHandler := NewAgentHandler(toolAgent, &wg)
	// 5. Orchestrator / runner
	cbReg := agentflow.NewCallbackRegistry()
	orch := agentflow.NewRouteOrchestrator(cbReg)

	concurrency := runtime.NumCPU()
	runner := agentflow.NewRunner(concurrency) // NewRunner wants only int
	runner.SetOrchestrator(orch)

	// 7. Register handler
	const handlerName = "tool_agent_handler"
	if err := runner.RegisterAgent(handlerName, agentHandler); err != nil {
		log.Fatalf("register agent: %v", err)
	}

	// 8. Start
	if err := runner.Start(ctx); err != nil {
		log.Fatalf("runner start: %v", err)
	}

	// Prepare the event
	event := &agentflow.SimpleEvent{
		ID: fmt.Sprintf("tool-evt-%d", time.Now().UnixNano()),
		Data: map[string]any{
			"user_request": "What is 15 plus 27.5?",
		}, Metadata: map[string]string{
			agentflow.RouteMetadataKey: handlerName,
		},
	}

	// Emit
	wg.Add(1)
	if err := runner.Emit(event); err != nil {
		log.Fatalf("emit: %v", err)
	}
	wg.Wait()

	// Stop (no arguments, no return)
	runner.Stop()
}
