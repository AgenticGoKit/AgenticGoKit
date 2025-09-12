package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/websocket"

	_ "testBot/agents"

	"github.com/kunalkushwaha/agenticgokit/core"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/logging/zerolog"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/orchestrator/default"
)

// Global agent handlers and shared results for WebUI access
var globalAgentHandlers map[string]core.AgentHandler
var globalResults *([]AgentOutput)
var globalResultsMutex *sync.Mutex

// Keep a global orchestrator reference for synchronous dispatch in WebUI
var globalOrchestrator core.Orchestrator

// --- Simple in-memory flow trace to drive a Mermaid debug view ---
type FlowEdge struct {
	From      string
	To        string
	Message   string
	Timestamp time.Time
}

var flowTraceMu sync.Mutex
var flowTrace = map[string][]FlowEdge{} // sessionID -> edges

func recordFlowEdge(sessionID, from, to, message string) {
	if sessionID == "" {
		sessionID = "webui-session"
	}
	edge := FlowEdge{From: from, To: to, Message: message, Timestamp: time.Now()}
	flowTraceMu.Lock()
	flowTrace[sessionID] = append(flowTrace[sessionID], edge)
	flowTraceMu.Unlock()
}

func extractMessage(event core.Event, state core.State) string {
	// Prefer state keys, then event data
	if state != nil {
		// Common input/output keys in priority order
		for _, key := range []string{"message", "user_input", "response", "output", "result", "content"} {
			if v, ok := state.Get(key); ok {
				if s, ok2 := v.(string); ok2 && s != "" {
					return s
				}
			}
		}
		// Back-compat explicit checks
		if v, ok := state.Get("message"); ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				return s
			}
		}
		if v, ok := state.Get("user_input"); ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				return s
			}
		}
	}
	if event != nil && event.GetData() != nil {
		// Check same set of keys in event data
		for _, key := range []string{"message", "user_input", "response", "output", "result", "content"} {
			if v, ok := event.GetData()[key]; ok {
				if s, ok2 := v.(string); ok2 && s != "" {
					return s
				}
			}
		}
		// Back-compat explicit checks
		if v, ok := event.GetData()["message"]; ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				return s
			}
		}
		if v, ok := event.GetData()["user_input"]; ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				return s
			}
		}
	}
	return ""
}

// main is the entry point for the testBot multi-agent system.
//
// This function orchestrates the entire workflow by:
// 1. Loading configuration from agentflow.toml
// 2. Initializing the LLM provider (OpenAI, Azure, Ollama, etc.)
// 3. Setting up optional components (MCP tools, memory system)
// 4. Creating and registering agent handlers from the agents package
// 5. Starting the workflow orchestrator and processing user input
// 6. Collecting and displaying results from all agents
//
// TODO: Customize this main function for your specific use case.
// Key customization points:
// - Add additional initialization steps for your services
// - Modify input processing and validation
// - Add custom middleware or interceptors
// - Implement custom result processing and output formatting
// - Add monitoring, metrics, or logging integrations
func main() {
	ctx := context.Background()

	// Load configuration first to get logging settings
	config, err := core.LoadConfig("agentflow.toml")
	if err != nil {
		// Use default logging for early errors
		core.SetLogLevel(core.INFO)
		logger := core.Logger()
		logger.Error().Err(err).Msg("Failed to load configuration")
		fmt.Printf("Failed to load configuration: %v\n", err)
		fmt.Printf("Hint: Make sure agentflow.toml exists and is properly formatted\n")
		os.Exit(1)
	}

	// Apply logging configuration from agentflow.toml
	config.ApplyLoggingConfig()

	logger := core.Logger()
	logger.Info().Str("log_level", config.Logging.Level).Str("log_format", config.Logging.Format).Msg("Starting testBot multi-agent system with configured logging")

	// TODO: Add any custom initialization logic here
	// Examples:
	// - Initialize database connections
	// - Set up monitoring or metrics collection
	// - Load additional configuration files
	// - Initialize external service clients

	// TODO: Customize command-line arguments for your application
	// You can add additional flags for configuration, debugging, or feature toggles
	messageFlag := flag.String("m", "", "Message to process")

	webuiFlag := flag.Bool("webui", false, "Start web interface mode")

	// TODO: Add your custom flags here
	// Examples:
	// debugFlag := flag.Bool("debug", false, "Enable debug mode")
	// configFlag := flag.String("config", "agentflow.toml", "Configuration file path")
	// outputFlag := flag.String("output", "", "Output file path")
	flag.Parse()

	// TODO: Add custom flag validation here
	// Example: if *messageFlag == "" { fmt.Println("Message is required"); os.Exit(1) }

	// Configuration already loaded above for logging setup
	// This file contains all the settings for your multi-agent system including
	// LLM provider configuration, orchestration settings, and feature toggles
	// TODO: Customize configuration loading if you need multiple config files
	// or environment-specific configurations

	// Initialize the LLM provider based on configuration
	// This creates the connection to your chosen AI service (OpenAI, Azure, Ollama, etc.)
	// TODO: Add custom provider initialization logic if needed
	// You might want to add connection pooling, rate limiting, or custom authentication
	llmProvider, err := initializeProvider(config.LLM.Provider)
	if err != nil {
		fmt.Printf("ERROR: Failed to initialize LLM provider '%s': %v\n", config.LLM.Provider, err)
		fmt.Printf("\nHint: Make sure you have set the appropriate environment variables:\n")
		switch config.LLM.Provider {
		case "azure":
			fmt.Printf("  AZURE_OPENAI_API_KEY=your-api-key\n")
			fmt.Printf("  AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/\n")
			fmt.Printf("  AZURE_OPENAI_DEPLOYMENT=your-deployment-name\n")
		case "openai":
			fmt.Printf("  OPENAI_API_KEY=your-api-key\n")
		case "ollama":
			fmt.Printf("  Ollama should be running on localhost:11434\n")
		default:
			fmt.Printf("  Check the documentation for provider '%s'\n", config.LLM.Provider)
		}
		// TODO: Add custom error handling or fallback providers here
		// Example: Try a fallback provider or provide offline mode
		os.Exit(1)
	}
	// Currently not used directly; agents created via configuration
	_ = llmProvider

	// TODO: Add LLM provider validation or health checks here
	// Example: Test the connection with a simple query
	// LLM provider initialized - debug logging reduced for cleaner output

	// Create configuration-driven agent factory
	// This factory creates agents based on the configuration in agentflow.toml
	// instead of hardcoded agent constructors, providing much more flexibility
	// TODO: Customize agent factory initialization if needed
	// You might want to add custom agent types or initialization logic
	// Configuration-driven agent factory initialization - debug logging reduced

	factory := core.NewConfigurableAgentFactory(config)
	if factory == nil {
		logger.Error().Msg("Failed to create agent factory")
		fmt.Printf("ERROR: Error creating agent factory\n")
		os.Exit(1)
	}

	// Create agent manager for centralized agent lifecycle management
	// The agent manager handles agent creation, initialization, and state management
	// TODO: Customize agent manager configuration if needed
	agentManager := core.NewAgentManager(config)
	if err := agentManager.InitializeAgents(); err != nil {
		logger.Error().Err(err).Msg("Failed to initialize agents from configuration")
		fmt.Printf("ERROR: Error initializing agents: %v\n", err)
		fmt.Printf("Hint: Check your agentflow.toml [agents] configuration\n")
		os.Exit(1)
	}

	// Get all active agents from the manager
	// This automatically excludes disabled agents and handles configuration-based filtering
	activeAgents := agentManager.GetActiveAgents()
	// Active agents loaded from configuration - debug logging reduced

	// Create agent handlers map for the workflow orchestrator
	// We wrap each agent with result collection for output tracking
	agentHandlers := make(map[string]core.AgentHandler)
	results := make([]AgentOutput, 0)
	var resultsMutex sync.Mutex

	// Register all active agents with result collection wrappers
	for _, agent := range activeAgents {
		agentName := agent.GetRole()

		// Adapt core.Agent to core.AgentHandler by delegating to HandleEvent
		baseHandler := core.AgentHandlerFunc(func(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
			return agent.HandleEvent(ctx, event, state)
		})

		// Wrap the adapted handler with result collection for output tracking
		wrappedAgent := &ResultCollectorHandler{
			originalHandler: baseHandler,
			agentName:       agentName,
			outputs:         &results,
			mutex:           &resultsMutex,
		}

		agentHandlers[agentName] = wrappedAgent
		// Agent registered with result collection - debug logging reduced
	}

	// Validate that we have at least one active agent
	if len(agentHandlers) == 0 {
		logger.Error().Msg("No active agents found in configuration")
		fmt.Printf("ERROR: No active agents configured\n")
		fmt.Printf("Hint: Check your agentflow.toml [agents] section and ensure at least one agent is enabled\n")
		fmt.Printf("Example agent configuration:\n")
		fmt.Printf("   [agents.my_agent]\n")
		fmt.Printf("   role = \"processor\"\n")
		fmt.Printf("   description = \"My processing agent\"\n")
		fmt.Printf("   system_prompt = \"You are a helpful assistant.\"\n")
		fmt.Printf("   capabilities = [\"processing\"]\n")
		fmt.Printf("   enabled = true\n")
		os.Exit(1)
	}

	// TODO: Add custom agent registration logic here
	// Example: Register agents conditionally based on configuration or environment

	// Create basic error handlers to prevent routing errors
	// These use the first active agent as a fallback handler for simplicity
	// TODO: Customize error handling agents for different error types
	var firstAgent core.AgentHandler
	for _, handler := range agentHandlers {
		firstAgent = handler
		break // Get the first agent
	}

	if firstAgent != nil {
		// Register error handlers using the first active agent as fallback
		// TODO: Create dedicated error handling agents for better error management
		agentHandlers["error-handler"] = firstAgent
		agentHandlers["validation-error-handler"] = firstAgent
		agentHandlers["timeout-error-handler"] = firstAgent
		agentHandlers["critical-error-handler"] = firstAgent
		agentHandlers["high-priority-error-handler"] = firstAgent
		agentHandlers["network-error-handler"] = firstAgent
		agentHandlers["llm-error-handler"] = firstAgent
		agentHandlers["auth-error-handler"] = firstAgent

		// Error handlers registered using first active agent as fallback - debug logging reduced
	}

	// Create the workflow orchestrator (runner) from configuration
	// This honors the orchestration mode from agentflow.toml and uses the plugin-registered orchestrator
	runner, err := core.NewRunnerFromConfig("agentflow.toml")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create runner from configuration")
		fmt.Printf("ERROR: Failed to create runner from configuration: %v\n", err)
		os.Exit(1)
	}

	// Create and configure the orchestrator based on the configuration
	orchestratorConfig := core.OrchestratorConfig{
		Type: config.Orchestration.Mode,
	}

	// Set up agent names based on orchestration mode
	switch config.Orchestration.Mode {
	case "sequential":
		orchestratorConfig.SequentialAgentNames = config.Orchestration.SequentialAgents
		orchestratorConfig.AgentNames = config.Orchestration.SequentialAgents
	case "collaborative", "parallel":
		orchestratorConfig.CollaborativeAgentNames = config.Orchestration.CollaborativeAgents
		orchestratorConfig.AgentNames = config.Orchestration.CollaborativeAgents
	case "loop":
		orchestratorConfig.AgentNames = []string{config.Orchestration.LoopAgent}
		orchestratorConfig.MaxIterations = config.Orchestration.MaxIterations
	case "mixed":
		orchestratorConfig.CollaborativeAgentNames = config.Orchestration.CollaborativeAgents
		orchestratorConfig.SequentialAgentNames = config.Orchestration.SequentialAgents
	}

	// Create the orchestrator using the registered factory
	orchestrator, err := core.NewOrchestrator(orchestratorConfig, runner.GetCallbackRegistry())
	if err != nil {
		logger.Error().Err(err).Str("mode", config.Orchestration.Mode).Msg("Failed to create orchestrator")
		fmt.Printf("ERROR: Failed to create orchestrator for mode '%s': %v\n", config.Orchestration.Mode, err)
		fmt.Printf("Hint: Make sure the orchestration mode is valid (sequential, collaborative, parallel, loop, mixed, route)\n")
		os.Exit(1)
	}

	// Attach the orchestrator to the runner (type assert to concrete implementation)
	if runnerImpl, ok := runner.(*core.RunnerImpl); ok {
		runnerImpl.SetOrchestrator(orchestrator)
		// Orchestrator configured successfully - debug logging reduced
	} else {
		logger.Error().Msg("Failed to cast runner to RunnerImpl for orchestrator setup")
		fmt.Printf("ERROR: Failed to configure orchestrator - runner type assertion failed\n")
		os.Exit(1)
	}

	// Expose orchestrator globally for WebUI synchronous dispatch
	globalOrchestrator = orchestrator

	// --- Hooks & Callbacks (Observability/Policies) ---
	// You can register before/after hooks and error hooks for traceability, metrics, or policy checks.
	// Hook points include:
	// - core.HookBeforeEventHandling / core.HookAfterEventHandling
	// - core.HookBeforeAgentRun / core.HookAgentError
	// Example:
	// cb := func(hookCtx context.Context, args core.CallbackArgs) (core.State, error) {
	//   // Inspect args.Event, args.State, args.AgentID, args.Error
	//   // Return a new state to replace args.State, or nil to keep it unchanged
	//   return nil, nil
	// }
	// if reg := runner.GetCallbackRegistry(); reg != nil {
	//   _ = reg.Register(core.HookBeforeAgentRun, "example-before-agent", cb)
	//   _ = reg.Register(core.HookAfterEventHandling, "example-after-event", cb)
	// }

	// Register all agents with the workflow orchestrator
	// This makes the agents available for routing and execution
	// TODO: Add custom agent registration logic if needed
	// Examples: conditional registration, agent prioritization, or custom routing
	for name, handler := range agentHandlers {
		if err := runner.RegisterAgent(name, handler); err != nil {
			logger.Error().Err(err).Str("agent", name).Msg("Failed to register agent")
			fmt.Printf("ERROR: Error registering agent %s: %v\n", name, err)
			os.Exit(1)
		}
		// Agent registered successfully - debug logging reduced
	}

	// All agents registered with orchestrator - debug logging reduced

	// Expose handlers/results to WebUI (share the same underlying data)
	globalAgentHandlers = agentHandlers
	globalResults = &results
	globalResultsMutex = &resultsMutex

	// Check if WebUI mode is requested
	if *webuiFlag {
		logger.Info().Msg("Starting WebUI mode")
		startWebUI(ctx, runner, config)
		return
	}

	// Process user input from command line or interactive prompt
	// TODO: Customize input processing for your application needs
	// You might want to add:
	// - Input validation and sanitization
	// - Support for different input formats (JSON, XML, etc.)
	// - File input processing
	// - Batch processing capabilities
	// - Input preprocessing or transformation
	var message string
	if *messageFlag != "" {
		message = *messageFlag
		// Using message from command line flag - debug logging reduced
	} else {
		// Interactive mode: prompt user for input
		// TODO: Enhance interactive input with better UX
		// Examples: multi-line input, input history, auto-completion
		fmt.Print("Enter your message: ")
		fmt.Scanln(&message)
	}

	// Provide default message if none specified
	// TODO: Customize the default message for your use case
	if message == "" {
		message = "Hello! Please provide information about current topics."
		// Using default message - debug logging reduced
	}

	// TODO: Add input validation and preprocessing here
	// Examples:
	// - Validate input length and format
	// - Sanitize input for security
	// - Transform input to expected format
	// - Add metadata or context to input

	logger.Info().Str("message", message).Msg("Processing user message")

	// Start the workflow orchestrator (non-blocking)
	// This initializes the event processing system and prepares agents for execution
	// TODO: Add custom pre-execution setup here
	// Examples: warm-up agents, pre-load data, initialize monitoring
	runner.Start(ctx)

	// Create and emit the initial event to start the workflow
	// The event contains the user message and routing information
	// TODO: Customize event creation for your workflow needs
	// You might want to add:
	// - Additional metadata or context
	// - Custom event types for different workflows
	// - Event validation or preprocessing
	// - Multiple initial events for parallel processing

	// Determine the first agent to route to based on orchestration configuration
	var firstAgentName string
	if config.Orchestration.Mode == "sequential" && len(config.Orchestration.SequentialAgents) > 0 {
		firstAgentName = config.Orchestration.SequentialAgents[0]
	} else if config.Orchestration.Mode == "collaborative" && len(config.Orchestration.CollaborativeAgents) > 0 {
		firstAgentName = config.Orchestration.CollaborativeAgents[0]
	} else {
		// Use the first active agent as fallback
		for agentName := range agentHandlers {
			if agentName != "error-handler" && !strings.Contains(agentName, "-error-handler") {
				firstAgentName = agentName
				break
			}
		}
	}

	// Fallback if no agent found
	if firstAgentName == "" {
		firstAgentName = "user_request"
	}

	event := core.NewEvent(firstAgentName, core.EventData{
		"message": message,
		// TODO: Add custom event data here
		// Examples:
		// "timestamp": time.Now(),
		// "user_id": userID,
		// "session_id": sessionID,
		// "priority": "normal",
	}, map[string]string{
		"route": firstAgentName,
		// TODO: Add custom routing metadata here
		// Examples:
		// "workflow_type": "standard",
		// "execution_mode": "async",
	})

	logger.Debug().Str("first_agent", firstAgentName).Msg("Initial event created for workflow start")

	// Emit the event to start workflow execution
	// TODO: Add custom event emission logic if needed
	// Examples: event batching, priority queuing, or conditional routing
	if err := runner.Emit(event); err != nil {
		logger.Error().Err(err).Msg("Workflow execution failed")
		fmt.Printf("Workflow execution error: %v\n", err)
		// TODO: Add custom error recovery or fallback logic here
		os.Exit(1)
	}

	logger.Info().Msg("Workflow execution started")

	// Wait for processing to complete BEFORE printing results.
	// We call runner.Stop() explicitly here (instead of using defer runner.Stop()).
	// A deferred call would execute only when main() returns—after the result-printing
	// code below—so we could print an empty "Agent Responses" section while the
	// agents are still working.  Calling Stop() now closes the queue and blocks
	// until the event-processing goroutine has finished handling all queued events,
	// guaranteeing the results slice is fully populated.
	logger.Debug().Msg("Waiting for agents to complete processing...")
	runner.Stop()

	// Process and display the collected results from all agents
	// TODO: Customize result processing and output formatting
	// You might want to add:
	// - Custom output formats (JSON, XML, CSV, etc.)
	// - Result filtering or aggregation
	// - Export to files or external systems
	// - Result validation or post-processing
	// - Custom visualization or reporting
	fmt.Printf("\n=== Agent Responses ===\n")
	resultsMutex.Lock()
	if len(results) > 0 {
		// TODO: Customize result display format
		// Examples: structured output, color coding, progress indicators
		for i, result := range results {
			fmt.Printf("\nAgent %s (Step %d):\n", result.AgentName, i+1)
			fmt.Printf("%s\n", result.Content)
			fmt.Printf("Completed at: %s\n", result.Timestamp.Format("15:04:05"))

			// TODO: Add custom result metadata display here
			// Examples: processing time, confidence scores, source information
		}

		// TODO: Add result summary or analytics here
		// Examples: total processing time, success rate, performance metrics
		fmt.Printf("\nSummary: %d agents completed successfully\n", len(results))
	} else {
		fmt.Printf("ERROR: No agent responses captured. This might indicate:\n")
		fmt.Printf("   - LLM provider credentials are not configured\n")
		fmt.Printf("   - An agent encountered an error during processing\n")
		fmt.Printf("   - The LLM provider is not responding\n")
		fmt.Printf("   - Network connectivity issues\n")

		// TODO: Add custom troubleshooting guidance for your specific setup
		fmt.Printf("\nTroubleshooting tips:\n")
		fmt.Printf("   - Check your environment variables\n")
		fmt.Printf("   - Verify agentflow.toml configuration\n")
		fmt.Printf("   - Review the logs above for detailed error information\n")
	}
	resultsMutex.Unlock()

	// TODO: Add custom post-processing here
	// Examples:
	// - Save results to database or file
	// - Send notifications or webhooks
	// - Update metrics or analytics
	// - Trigger follow-up workflows

	fmt.Printf("\n=== Workflow Completed ===\n")
	fmt.Printf("Check the logs above for detailed agent execution results.\n")

	// TODO: Add custom completion logic here
	// Examples: cleanup resources, send completion notifications, update status

	logger.Info().Msg("Workflow completed successfully")
}

// ResultCollectorHandler wraps an agent handler to capture its outputs for display.
//
// This middleware pattern allows us to collect results from all agents without
// modifying the agent implementations themselves. It intercepts the agent's
// output and stores it for later display while passing through the original
// result to the workflow orchestrator.
//
// TODO: Customize this handler for your specific result collection needs
// You might want to add:
// - Result filtering or transformation
// - Custom metadata collection
// - Result persistence to database or file
// - Real-time result streaming
// - Result validation or quality checks
type ResultCollectorHandler struct {
	originalHandler core.AgentHandler
	agentName       string
	outputs         *[]AgentOutput
	mutex           *sync.Mutex
}

// AgentOutput holds the output from an agent along with metadata.
//
// This structure captures the essential information about each agent's
// execution result for display and analysis purposes.
//
// TODO: Extend this structure for your specific needs
// You might want to add:
// - Processing duration
// - Confidence scores
// - Error details
// - Source information
// - Custom metadata fields
type AgentOutput struct {
	AgentName string    // Name of the agent that produced this output
	Content   string    // The actual output content from the agent
	Timestamp time.Time // When the agent completed processing

	// TODO: Add custom fields here
	// Examples:
	Duration time.Duration
	// Confidence   float64
	// ErrorDetails string
	// Metadata     map[string]interface{}
}

// Run implements the AgentHandler interface and captures the output.
//
// This method acts as a middleware that:
// 1. Calls the original agent handler
// 2. Extracts meaningful content from the result
// 3. Stores the output for later display
// 4. Returns the original result unchanged
//
// TODO: Customize result extraction and storage logic
// You might want to modify how content is extracted, add custom
// metadata collection, or implement different storage strategies.
func (r *ResultCollectorHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// TODO: Add pre-execution logic here if needed
	// Examples: start timing, log execution start, validate input
	startTime := time.Now()

	// Call the original agent handler to get the actual result
	result, err := r.originalHandler.Run(ctx, event, state)

	// TODO: Add post-execution logic here if needed
	// Examples: calculate processing time, validate output, add metrics
	processingTime := time.Since(startTime)

	// Record flow edge for debug visualization (User -> Agent input)
	sessionID, _ := event.GetMetadataValue("session_id")
	from := event.GetSourceAgentID()
	if from == "" {
		from = "User"
	}
	to := r.agentName
	recordFlowEdge(sessionID, from, to, extractMessage(event, state))

	// Extract meaningful content from the result for display
	// This tries multiple common output keys to find the agent's response
	// TODO: Customize content extraction for your specific agent output format
	var content string
	if err != nil {
		content = fmt.Sprintf("Error: %v", err)
	} else if result.Error != "" {
		content = fmt.Sprintf("Agent Error: %s", result.Error)
	} else {
		// Try to extract content from the result's output state
		// TODO: Add custom content extraction logic here
		if result.OutputState != nil {
			// Try common output keys in order of preference
			keys := []string{"response", "output", "message", "result", "content"}
			for _, key := range keys {
				if data, exists := result.OutputState.Get(key); exists {
					if str, ok := data.(string); ok && str != "" {
						content = str
						break
					}
				}
			}
		}
	}

	// Provide fallback content if nothing was extracted
	// TODO: Customize fallback content for your use case
	if content == "" {
		content = fmt.Sprintf("Agent %s completed processing successfully", r.agentName)
	}

	// Record agent -> next route (or -> User) with the agent's response for trace visualization
	{
		nextTo := "User"
		if result.OutputState != nil {
			if route, ok := result.OutputState.GetMeta(core.RouteMetadataKey); ok && route != "" {
				nextTo = fmt.Sprintf("%v", route)
			}
		}
		// Prefer the content we extracted; if empty, try common keys from OutputState
		outMsg := content
		if outMsg == "" && result.OutputState != nil {
			for _, key := range []string{"response", "output", "result", "content"} {
				if v, ok := result.OutputState.Get(key); ok {
					if s, ok2 := v.(string); ok2 && s != "" {
						outMsg = s
						break
					}
				}
			}
		}
		recordFlowEdge(sessionID, r.agentName, nextTo, outMsg)
	}

	// Store the output in a thread-safe manner
	// TODO: Customize output storage (e.g., add to database, send to external system)
	r.mutex.Lock()
	*r.outputs = append(*r.outputs, AgentOutput{
		AgentName: r.agentName,
		Content:   content,
		Timestamp: time.Now(),
		// TODO: Add custom fields here
		// Examples:
		Duration: processingTime,
		// Confidence: extractConfidence(result),
		// Metadata:   extractMetadata(result),
	})
	r.mutex.Unlock()

	// Return the original result unchanged so the workflow continues normally
	return result, err
}

func initializeProvider(providerType string) (core.ModelProvider, error) {
	// Load configuration to get provider settings
	config, err := core.LoadConfig("agentflow.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Use the global LLM configuration from agentflow.toml
	llmConfig := config.LLM
	if llmConfig.Provider == "" {
		llmConfig.Provider = providerType // Use parameter as fallback
	}

	// Create provider configuration with environment variable resolution
	providerConfig := core.LLMProviderConfig{
		Type:        llmConfig.Provider,
		Model:       llmConfig.Model,
		Temperature: llmConfig.Temperature,
		MaxTokens:   llmConfig.MaxTokens,
		HTTPTimeout: core.TimeoutFromSeconds(llmConfig.TimeoutSeconds),
	}

	// Read API key from environment variables based on provider type
	switch llmConfig.Provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required for OpenAI provider")
		}
		providerConfig.APIKey = apiKey
	case "azure", "azureopenai":
		apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
		endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
		deployment := os.Getenv("AZURE_OPENAI_DEPLOYMENT")
		if apiKey == "" {
			return nil, fmt.Errorf("AZURE_OPENAI_API_KEY environment variable is required for Azure provider")
		}
		if endpoint == "" {
			return nil, fmt.Errorf("AZURE_OPENAI_ENDPOINT environment variable is required for Azure provider")
		}
		if deployment == "" {
			return nil, fmt.Errorf("AZURE_OPENAI_DEPLOYMENT environment variable is required for Azure provider")
		}
		providerConfig.APIKey = apiKey
		providerConfig.Endpoint = endpoint
		providerConfig.ChatDeployment = deployment
		providerConfig.EmbeddingDeployment = deployment
	case "ollama":
		// Ollama doesn't require API key, use base URL from config or default
		baseURL := "http://localhost:11434"
		if ollamaConfig, exists := config.Providers["ollama"]; exists {
			if url, ok := ollamaConfig["base_url"].(string); ok && url != "" {
				baseURL = url
			}
		}
		providerConfig.BaseURL = baseURL
	}

	// Create provider from configuration
	provider, err := core.NewModelProviderFromConfig(providerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider '%s': %w", llmConfig.Provider, err)
	}

	return provider, nil
}

// WebUI Server Functions
func startWebUI(ctx context.Context, runner core.Runner, config *core.Config) {
	// Start the runner for orchestration support in WebUI mode
	log.Printf("Starting workflow orchestrator for WebUI mode")
	runner.Start(ctx)

	staticDir := "internal/webui/static"
	if staticDir == "" {
		staticDir = "internal/webui/static"
	}

	port := 8080
	if port == 0 {
		port = 8080
	}

	// Setup HTTP routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}
		http.FileServer(http.Dir(staticDir)).ServeHTTP(w, r)
	})

	http.HandleFunc("/api/agents", handleGetAgents)
	http.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		handleChat(w, r, ctx, runner, config)
	})
	// Raw config endpoints (read/update agentflow.toml)
	http.HandleFunc("/api/config/raw", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetRawConfig(w, r)
			return
		}
		if r.Method == http.MethodPut {
			handlePutRawConfig(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	// Simple Mermaid composition diagram endpoint
	http.HandleFunc("/api/visualization/composition", func(w http.ResponseWriter, r *http.Request) {
		handleCompositionDiagram(w, r, config)
	})
	// Debug trace diagram endpoint (sequence view per session)
	http.HandleFunc("/api/visualization/trace", func(w http.ResponseWriter, r *http.Request) {
		handleTraceDiagram(w, r)
	})
	// WebSocket streaming endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, ctx, runner, config)
	})
	// Config endpoint to inform frontend of features and defaults
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Determine recommended default for orchestration toggle
		defaultOrch := false
		mode := config.Orchestration.Mode
		if mode == "sequential" || mode == "collaborative" || mode == "parallel" || mode == "loop" || mode == "mixed" || mode == "route" {
			defaultOrch = true
		}
		resp := map[string]any{
			"server": map[string]any{
				"name": config.AgentFlow.Name,
				"port": port,
				"url":  fmt.Sprintf("http://localhost:%d", port),
			},
			"features": map[string]any{
				"websocket": true,
				"streaming": true,
			},
			"orchestration": map[string]any{
				"mode":                 mode,
				"default_enabled":      defaultOrch,
				"sequential_agents":    config.Orchestration.SequentialAgents,
				"collaborative_agents": config.Orchestration.CollaborativeAgents,
				"loop_agent":           config.Orchestration.LoopAgent,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	log.Printf("Starting WebUI server on port %d", port)
	log.Printf("Visit http://localhost:%d to access the chat interface", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func handleGetAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Return actual agents from globalAgentHandlers, skipping error-handlers
	agents := []map[string]interface{}{}
	for agentName := range globalAgentHandlers {
		if strings.Contains(agentName, "error-handler") {
			continue
		}
		agents = append(agents, map[string]interface{}{
			"id":          agentName,
			"name":        agentName,
			"description": fmt.Sprintf("Agent %s from configuration", agentName),
		})
	}

	if len(agents) == 0 {
		// Fallback default agents if none registered
		agents = []map[string]interface{}{
			{"id": "agent1", "name": "Agent1", "description": "Default agent 1"},
			{"id": "agent2", "name": "Agent2", "description": "Default agent 2"},
		}
	}

	_ = json.NewEncoder(w).Encode(agents)
}

func handleChat(w http.ResponseWriter, r *http.Request, ctx context.Context, runner core.Runner, config *core.Config) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request
	var req struct {
		Message          string `json:"message"`
		Agent            string `json:"agent"`
		UseOrchestration bool   `json:"useOrchestration,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to decode JSON request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: Received chat request - Agent: %s, UseOrchestration: %v", req.Agent, req.UseOrchestration)

	// Verify agent exists
	agentHandler, exists := globalAgentHandlers[req.Agent]
	if !exists {
		log.Printf("ERROR: Agent '%s' not found. Available: %v", req.Agent, getAgentNames())
		http.Error(w, fmt.Sprintf("Agent '%s' not found", req.Agent), http.StatusNotFound)
		return
	}

	var agentResponse string
	var status string

	if req.UseOrchestration {
		// Orchestrated multi-agent flow: dispatch synchronously via orchestrator
		log.Printf("INFO: Using orchestration mode for agent %s", req.Agent)
		if globalOrchestrator == nil {
			log.Printf("ERROR: Orchestrator not available for WebUI dispatch")
			http.Error(w, "Orchestrator not available", http.StatusInternalServerError)
			return
		}

		// Clear previous results and run synchronously
		globalResultsMutex.Lock()
		*globalResults = []AgentOutput{}
		globalResultsMutex.Unlock()

		event := core.NewEvent(req.Agent, core.EventData{
			"message":    req.Message,
			"user_input": req.Message,
		}, map[string]string{
			"route":      req.Agent,
			"session_id": "webui-session",
		})

		if _, err := globalOrchestrator.Dispatch(ctx, event); err != nil {
			log.Printf("ERROR: Orchestration dispatch failed: %v", err)
			http.Error(w, fmt.Sprintf("Orchestration error: %v", err), http.StatusInternalServerError)
			return
		}

		// Extract the last collected agent output
		globalResultsMutex.Lock()
		if len(*globalResults) > 0 {
			latest := (*globalResults)[len(*globalResults)-1]
			agentResponse = latest.Content
			status = "completed"
		} else {
			agentResponse = "Orchestration completed, but no agent responses captured."
			status = "no_output"
		}
		globalResultsMutex.Unlock()
	} else {
		// Direct single agent invocation
		log.Printf("INFO: Using direct agent mode for agent %s", req.Agent)
		state := core.NewState()
		state.Set("message", req.Message)
		state.Set("user_input", req.Message)

		event := core.NewEvent(req.Agent, core.EventData{
			"message": req.Message,
		}, map[string]string{
			"route":      req.Agent,
			"session_id": "webui-session",
		})

		result, err := agentHandler.Run(ctx, event, state)
		if err != nil {
			log.Printf("ERROR: Agent processing failed: %v", err)
			http.Error(w, fmt.Sprintf("Agent processing error: %v", err), http.StatusInternalServerError)
			return
		}

		if result.OutputState != nil {
			if v, ok := result.OutputState.Get("result"); ok {
				agentResponse = fmt.Sprintf("%v", v)
			} else if v, ok := result.OutputState.Get("response"); ok {
				agentResponse = fmt.Sprintf("%v", v)
			} else if v, ok := result.OutputState.Get("output"); ok {
				agentResponse = fmt.Sprintf("%v", v)
			} else {
				agentResponse = "Agent processed your request successfully"
			}
		} else {
			agentResponse = "Agent processed your request"
		}
		status = "completed"
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"response": agentResponse,
		"agent":    req.Agent,
		"status":   status,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// Helper: list available agent names
func getAgentNames() []string {
	names := make([]string, 0, len(globalAgentHandlers))
	for name := range globalAgentHandlers {
		names = append(names, name)
	}
	return names
}

// WebSocket support
var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type wsInbound struct {
	Type             string
	Agent            string
	Message          string
	UseOrchestration bool
}

type wsOutbound struct {
	Type      string                 `json:"type"`
	Agent     string                 `json:"agent,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Chunk     int                    `json:"chunk_index,omitempty"`
	Total     int                    `json:"total_chunks,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, ctx context.Context, runner core.Runner, config *core.Config) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Send welcome
	_ = conn.WriteJSON(wsOutbound{Type: "welcome", Timestamp: time.Now().Unix()})

	for {
		var msg wsInbound
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("WS read error: %v", err)
			return
		}

		switch msg.Type {
		case "chat":
			if msg.Agent == "" || msg.Message == "" {
				_ = conn.WriteJSON(wsOutbound{Type: "error", Status: "bad_request", Content: "agent and message required", Timestamp: time.Now().Unix()})
				continue
			}

			// Progress notification
			_ = conn.WriteJSON(wsOutbound{Type: "agent_progress", Agent: msg.Agent, Status: "processing", Content: "Processing...", Timestamp: time.Now().Unix()})

			// Process in a goroutine and stream chunks
			go func(m wsInbound) {
				content, status := processRequest(ctx, m.Agent, m.Message, m.UseOrchestration)
				// Stream chunked content
				chunks := chunkString(content, 180)
				total := len(chunks)
				for i, c := range chunks {
					_ = conn.WriteJSON(wsOutbound{Type: "agent_chunk", Agent: m.Agent, Content: c, Chunk: i, Total: total, Timestamp: time.Now().Unix()})
					time.Sleep(50 * time.Millisecond)
				}
				_ = conn.WriteJSON(wsOutbound{Type: "agent_complete", Agent: m.Agent, Status: status, Content: content, Timestamp: time.Now().Unix()})
			}(msg)
		default:
			_ = conn.WriteJSON(wsOutbound{Type: "error", Status: "unknown_type", Content: "Unsupported message type", Timestamp: time.Now().Unix()})
		}
	}
}

// processRequest runs either orchestrated or direct agent call and returns final content and status
func processRequest(ctx context.Context, agent string, message string, useOrch bool) (string, string) {
	if useOrch {
		if globalOrchestrator == nil {
			return "Orchestrator not available", "error"
		}
		globalResultsMutex.Lock()
		*globalResults = []AgentOutput{}
		globalResultsMutex.Unlock()

		event := core.NewEvent(agent, core.EventData{
			"message":    message,
			"user_input": message,
		}, map[string]string{
			"route":      agent,
			"session_id": "webui-session",
		})

		if _, err := globalOrchestrator.Dispatch(ctx, event); err != nil {
			return fmt.Sprintf("Orchestration error: %v", err), "error"
		}
		globalResultsMutex.Lock()
		defer globalResultsMutex.Unlock()
		if len(*globalResults) > 0 {
			latest := (*globalResults)[len(*globalResults)-1]
			return latest.Content, "completed"
		}
		return "Orchestration completed, but no agent responses captured.", "no_output"
	}

	// Direct
	agentHandler, exists := globalAgentHandlers[agent]
	if !exists {
		return fmt.Sprintf("Agent '%s' not found", agent), "not_found"
	}
	state := core.NewState()
	state.Set("message", message)
	state.Set("user_input", message)

	event := core.NewEvent(agent, core.EventData{
		"message": message,
	}, map[string]string{
		"route":      agent,
		"session_id": "webui-session",
	})

	result, err := agentHandler.Run(ctx, event, state)
	if err != nil {
		return fmt.Sprintf("Agent processing error: %v", err), "error"
	}
	if result.OutputState != nil {
		if v, ok := result.OutputState.Get("result"); ok {
			return fmt.Sprintf("%v", v), "completed"
		} else if v, ok := result.OutputState.Get("response"); ok {
			return fmt.Sprintf("%v", v), "completed"
		} else if v, ok := result.OutputState.Get("output"); ok {
			return fmt.Sprintf("%v", v), "completed"
		}
	}
	return "Agent processed your request", "completed"
}

func chunkString(s string, size int) []string {
	if size <= 0 || len(s) == 0 {
		return []string{s}
	}
	var chunks []string
	for start := 0; start < len(s); start += size {
		end := start + size
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[start:end])
	}
	return chunks
}

// --- Config and Visualization Helpers ---

// getConfigPath returns env override path or default CWD agentflow.toml
func getConfigPath() string {
	if env := os.Getenv("AGENTFLOW_CONFIG_PATH"); env != "" {
		return env
	}
	if wd, err := os.Getwd(); err == nil {
		return filepath.Join(wd, "agentflow.toml")
	}
	return "agentflow.toml"
}

// atomicWriteFile writes data atomically to path
func atomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// handleGetRawConfig returns the contents of agentflow.toml
func handleGetRawConfig(w http.ResponseWriter, r *http.Request) {
	path := getConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "agentflow.toml not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to read config", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"data":   map[string]any{"path": path, "size": len(data), "content": string(data)},
	})
}

// handlePutRawConfig updates agentflow.toml after basic validation
func handlePutRawConfig(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	var body struct {
		Toml string `json:"toml"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		if err == io.EOF {
			http.Error(w, "Empty body", http.StatusBadRequest)
			return
		}
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Toml) == "" {
		http.Error(w, "Missing 'toml'", http.StatusBadRequest)
		return
	}

	// Parse TOML using core loader path for compatibility
	tmpFile := filepath.Join(os.TempDir(), "agentflow-validate-"+fmt.Sprint(time.Now().UnixNano())+".toml")
	_ = os.WriteFile(tmpFile, []byte(body.Toml), 0644)
	parsed, err := core.LoadConfig(tmpFile)
	_ = os.Remove(tmpFile)
	if err != nil {
		http.Error(w, "TOML parse error: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := parsed.ValidateOrchestrationConfig(); err != nil {
		http.Error(w, "Config validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Write file atomically
	path := getConfigPath()
	if err := atomicWriteFile(path, []byte(body.Toml)); err != nil {
		http.Error(w, "Failed to write config", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": map[string]any{"path": path, "updated": true}})
}

// handleCompositionDiagram returns a Mermaid composition diagram
func handleCompositionDiagram(w http.ResponseWriter, r *http.Request, cfg *core.Config) {
	w.Header().Set("Content-Type", "application/json")
	name := "agentflow"
	mode := "composition"
	if cfg != nil {
		if cfg.AgentFlow.Name != "" {
			name = cfg.AgentFlow.Name
		}
		if cfg.Orchestration.Mode != "" {
			mode = cfg.Orchestration.Mode
		}
	}
	// derive agents list from globalAgentHandlers keys
	// For simplicity here, we just pass zero agents if none available
	var agents []core.Agent
	gen := core.NewMermaidGenerator()
	mcfg := core.DefaultMermaidConfig()
	diagram := gen.GenerateCompositionDiagram(mode, name, agents, mcfg)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"data":   map[string]any{"title": name + " (" + mode + ")", "diagram": diagram},
	})
}

// handleTraceDiagram returns a Mermaid sequence diagram for the captured flow trace
func handleTraceDiagram(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// session query param, default to webui-session
	session := r.URL.Query().Get("session")
	if session == "" {
		session = "webui-session"
	}
	flowTraceMu.Lock()
	edges := append([]FlowEdge(nil), flowTrace[session]...)
	flowTraceMu.Unlock()

	diagram := buildMermaidSequence(edges)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"data": map[string]any{
			"session": session,
			"diagram": diagram,
			"edges":   len(edges),
		},
	})
}

func buildMermaidSequence(edges []FlowEdge) string {
	// Mermaid sequence diagram header
	var b strings.Builder
	b.WriteString("sequenceDiagram\n")
	// Collect participants in insertion order
	seen := map[string]bool{}
	order := []string{}
	add := func(name string) {
		if name == "" {
			return
		}
		if !seen[name] {
			seen[name] = true
			order = append(order, name)
		}
	}
	for _, e := range edges {
		add(e.From)
		add(e.To)
	}
	if len(order) == 0 {
		// Provide a stub
		b.WriteString("  participant User\n  participant Agent\n  User->>Agent: (no activity captured)\n")
		return b.String()
	}
	for _, p := range order {
		b.WriteString("  participant " + escapeMermaidIdent(p) + "\n")
	}

	for _, e := range edges {
		msg := e.Message
		if msg == "" {
			msg = "(no message)"
		}
		// Limit message to reasonable length
		if len(msg) > 180 {
			msg = msg[:177] + "..."
		}
		b.WriteString("  " + escapeMermaidIdent(e.From) + "->>" + escapeMermaidIdent(e.To) + ": " + escapeMermaidText(msg) + "\n")
	}
	return b.String()
}

func escapeMermaidIdent(s string) string {
	// Very basic: replace spaces and special chars with underscores
	repl := s
	repl = strings.ReplaceAll(repl, " ", "_")
	repl = strings.ReplaceAll(repl, "-", "_")
	repl = strings.ReplaceAll(repl, ".", "_")
	return repl
}

func escapeMermaidText(s string) string {
	// Escape quotes and backticks minimally
	repl := strings.ReplaceAll(s, "\n", " ")
	repl = strings.ReplaceAll(repl, "\r", " ")
	repl = strings.ReplaceAll(repl, "\"", "\\\"")
	repl = strings.ReplaceAll(repl, "`", "'")
	return repl
}
