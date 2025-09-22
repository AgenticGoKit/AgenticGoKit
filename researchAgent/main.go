package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	
	"net/http"
	"path/filepath"
	"log"
	"encoding/json"
	configh "researchAgent/internal/config"
	httpHandlers "researchAgent/internal/handlers"
	tracingh "researchAgent/internal/tracing"
	

	"github.com/kunalkushwaha/agenticgokit/core"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/logging/zerolog"
	_ "github.com/kunalkushwaha/agenticgokit/plugins/orchestrator/default"
	_ "researchAgent/agents"
)




// Global agent handlers and shared results for WebUI access
var globalAgentHandlers map[string]core.AgentHandler
var globalResults *([]AgentOutput)
var globalResultsMutex *sync.Mutex
// Keep a global orchestrator reference for synchronous dispatch in WebUI
var globalOrchestrator core.Orchestrator



// main is the entry point for the researchAgent multi-agent system.
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
	logger.Info().Str("log_level", config.Logging.Level).Str("log_format", config.Logging.Format).Msg("Starting researchAgent multi-agent system with configured logging")

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

	
	// Initialize MCP (Model Context Protocol) manager for tool integration
	// MCP allows agents to access external tools and services like file systems,
	// databases, APIs, and other integrations defined in your agentflow.toml
	// TODO: Customize MCP initialization for your specific tool requirements
	// You might want to add custom tool validation, authentication, or configuration
	// MCP initialization - debug logging reduced for cleaner output
	mcpInitCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var mcpManager core.MCPManager
	mcpDone := make(chan bool, 1)
	var mcpErr error

	// Initialize MCP in a separate goroutine to handle timeouts gracefully
	// TODO: Add custom MCP initialization logic if needed
	go func() {
		mcpManager, mcpErr = initializeMCP()
		mcpDone <- true
	}()

	select {
	case <-mcpDone:
		if mcpErr != nil {
			logger.Warn().Err(mcpErr).Msg("MCP initialization failed, continuing without MCP")
			mcpManager = nil
		}
	case <-mcpInitCtx.Done():
		logger.Warn().Msg("MCP initialization timed out, continuing without MCP")
		mcpManager = nil
		mcpErr = fmt.Errorf("MCP initialization timeout")
	}

	if mcpManager != nil {
		// MCP manager initialized successfully - debug logging reduced

		// Initialize MCP tool registry with timeout
		registryCtx, registryCancel := context.WithTimeout(ctx, 10*time.Second)
		defer registryCancel()

		registryDone := make(chan error, 1)
		go func() {
			registryDone <- core.InitializeMCPToolRegistry()
		}()

		select {
		case err := <-registryDone:
			if err != nil {
				logger.Warn().Err(err).Msg("Failed to initialize MCP tool registry")
			} else {
				// MCP tool registry initialized successfully - debug logging reduced
			}
		case <-registryCtx.Done():
			logger.Warn().Msg("MCP tool registry initialization timed out")
		}

		// Register MCP tools with the registry with timeout
		toolsCtx, toolsCancel := context.WithTimeout(ctx, 10*time.Second)
		defer toolsCancel()

		toolsDone := make(chan error, 1)
		go func() {
			toolsDone <- core.RegisterMCPToolsWithRegistry(toolsCtx)
		}()

		select {
		case err := <-toolsDone:
			if err != nil {
				logger.Warn().Err(err).Msg("Failed to register MCP tools with registry")
			} else {
				// MCP tools registered with registry successfully - debug logging reduced
			}
		case <-toolsCtx.Done():
			logger.Warn().Msg("MCP tools registration timed out")
		}
	}

	
	

	

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




func initializeMCP() (core.MCPManager, error) {
	// Load configuration from agentflow.toml in current directory
	config, err := core.LoadConfigFromWorkingDir()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if MCP is enabled in configuration
	if !config.MCP.Enabled {
		return nil, fmt.Errorf("MCP is not enabled in agentflow.toml")
	}

	// Convert TOML config to MCP config
	mcpConfig := core.MCPConfig{
		EnableDiscovery:   config.MCP.EnableDiscovery,
		ConnectionTimeout: time.Duration(config.MCP.ConnectionTimeout) * time.Millisecond,
		MaxRetries:        config.MCP.MaxRetries,
		RetryDelay:        time.Duration(config.MCP.RetryDelay) * time.Millisecond,
		EnableCaching:     config.MCP.EnableCaching,
		CacheTimeout:      time.Duration(config.MCP.CacheTimeout) * time.Millisecond,
		MaxConnections:    config.MCP.MaxConnections,
		Servers:           make([]core.MCPServerConfig, len(config.MCP.Servers)),
	}

	// Convert server configurations
	for i, server := range config.MCP.Servers {
		mcpConfig.Servers[i] = core.MCPServerConfig{
			Name:    server.Name,
			Type:    server.Type,
			Host:    server.Host,
			Port:    server.Port,
			Command: server.Command,
			Enabled: server.Enabled,
		}
	}

	// Initialize MCP manager with configuration from TOML
	err = core.InitializeMCP(mcpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MCP: %w", err)
	}

	// Get the initialized MCP manager
	manager := core.GetMCPManager()
	if manager == nil {
		return nil, fmt.Errorf("MCP manager not available after initialization")
	}

	return manager, nil
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
	Duration     time.Duration
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

	// Store the output in a thread-safe manner
	// TODO: Customize output storage (e.g., add to database, send to external system)
	r.mutex.Lock()
	*r.outputs = append(*r.outputs, AgentOutput{
		AgentName: r.agentName,
		Content:   content,
		Timestamp: time.Now(),
		// TODO: Add custom fields here
		// Examples:
		Duration:   processingTime,
		// Confidence: extractConfidence(result),
		// Metadata:   extractMetadata(result),
	})
	r.mutex.Unlock()

	
	// Record an edge for debugging visualization (shim), including full message content
	tracingh.RecordAgentTransition("webui-session", r.agentName, result, content)
	

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
	// Wire framework-native tracing hooks before starting the runner
	if reg := runner.GetCallbackRegistry(); reg != nil {
		if tl := runner.GetTraceLogger(); tl != nil {
			if err := core.RegisterTraceHooks(reg, tl); err != nil {
				log.Printf("WARN: RegisterTraceHooks failed: %v", err)
			}
		} else {
			log.Printf("WARN: TraceLogger not configured; trace view may be empty")
		}
	}

	// Start the runner for orchestration support in WebUI mode
	log.Printf("Starting workflow orchestrator for WebUI mode")
	runner.Start(ctx)

	staticDir := "internal/webui/static"
	if staticDir == "" {
		staticDir = "internal/webui/static"
	}
	port := 8080
	if port == 0 { port = 8080 }

	// Setup HTTP routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" { http.ServeFile(w, r, filepath.Join(staticDir, "index.html")); return }
		http.FileServer(http.Dir(staticDir)).ServeHTTP(w, r)
	})

	// Handlers server bundling orchestrator/agents/results
	srv := httpHandlers.NewServer(ctx, runner, config, globalOrchestrator, globalAgentHandlers, &resultsProxy{res: globalResults, mu: globalResultsMutex})
	http.HandleFunc("/api/agents", srv.HandleGetAgents)
	http.HandleFunc("/api/chat", srv.HandleChat)
	// Raw config endpoints
	http.HandleFunc("/api/config/raw", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { configh.HandleGetRaw(w, r); return }
		if r.Method == http.MethodPut { configh.HandlePutRaw(w, r); return }
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	// Visualization endpoints
	http.HandleFunc("/api/visualization/composition", srv.HandleCompositionDiagram)
	http.HandleFunc("/api/visualization/trace", srv.HandleTraceDiagram)
	// WebSocket
	http.HandleFunc("/ws", srv.HandleWebSocket)
	// Config info
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed); return }
		w.Header().Set("Content-Type", "application/json")
		resp := configh.BuildInfo(config, port)
		_ = json.NewEncoder(w).Encode(resp)
	})

	log.Printf("Starting WebUI server on port %d", port)
	log.Printf("Visit http://localhost:%d to access the chat interface", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// resultsProxy adapts the shared results slice to the handlers.ResultStore interface
type resultsProxy struct { res *[]AgentOutput; mu *sync.Mutex }
func (rp *resultsProxy) Reset() { rp.mu.Lock(); defer rp.mu.Unlock(); *rp.res = []AgentOutput{} }
func (rp *resultsProxy) Latest() (string, bool) {
	rp.mu.Lock(); defer rp.mu.Unlock()
	if len(*rp.res) == 0 { return "", false }
	latest := (*rp.res)[len(*rp.res)-1]
	return latest.Content, true
}

