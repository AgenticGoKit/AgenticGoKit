package templates

const MainTemplate = `package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	{{if or (eq .Config.OrchestrationMode "collaborative") .Config.MemoryEnabled}}
	"strings"
	{{end}}
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"{{.Config.Name}}/agents"
)

{{if .Config.MemoryEnabled}}
// Global memory instance for access by agents
var memory core.Memory
{{end}}

// main is the entry point for the {{.Config.Name}} multi-agent system.
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
	core.SetLogLevel(core.INFO)
	logger := core.Logger()
	logger.Info().Msg("Starting {{.Config.Name}} multi-agent system...")

	// TODO: Add any custom initialization logic here
	// Examples:
	// - Initialize database connections
	// - Set up monitoring or metrics collection
	// - Load additional configuration files
	// - Initialize external service clients

	// TODO: Customize command-line arguments for your application
	// You can add additional flags for configuration, debugging, or feature toggles
	messageFlag := flag.String("m", "", "Message to process")
	// TODO: Add your custom flags here
	// Examples:
	// debugFlag := flag.Bool("debug", false, "Enable debug mode")
	// configFlag := flag.String("config", "agentflow.toml", "Configuration file path")
	// outputFlag := flag.String("output", "", "Output file path")
	flag.Parse()

	// TODO: Add custom flag validation here
	// Example: if *messageFlag == "" { fmt.Println("Message is required"); os.Exit(1) }

	// Load configuration from agentflow.toml
	// This file contains all the settings for your multi-agent system including
	// LLM provider configuration, orchestration settings, and feature toggles
	// TODO: Customize configuration loading if you need multiple config files
	// or environment-specific configurations
	config, err := core.LoadConfig("agentflow.toml")
	if err != nil {
		// TODO: Add custom error handling for configuration loading
		// You might want to provide more specific error messages or fallback configurations
		fmt.Printf("Failed to load configuration: %v\n", err)
		fmt.Printf("💡 Make sure agentflow.toml exists and is properly formatted\n")
		os.Exit(1)
	}

	// Initialize the LLM provider based on configuration
	// This creates the connection to your chosen AI service (OpenAI, Azure, Ollama, etc.)
	// TODO: Add custom provider initialization logic if needed
	// You might want to add connection pooling, rate limiting, or custom authentication
	llmProvider, err := initializeProvider(config.AgentFlow.Provider)
	if err != nil {
		fmt.Printf("❌ Failed to initialize LLM provider '%s': %v\n", config.AgentFlow.Provider, err)
		fmt.Printf("\n💡 Make sure you have set the appropriate environment variables:\n")
		switch config.AgentFlow.Provider {
		case "azure":
			fmt.Printf("  AZURE_OPENAI_API_KEY=your-api-key\n")
			fmt.Printf("  AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/\n")
			fmt.Printf("  AZURE_OPENAI_DEPLOYMENT=your-deployment-name\n")
		case "openai":
			fmt.Printf("  OPENAI_API_KEY=your-api-key\n")
		case "ollama":
			fmt.Printf("  Ollama should be running on localhost:11434\n")
		default:
			fmt.Printf("  Check the documentation for provider '%s'\n", config.AgentFlow.Provider)
		}
		// TODO: Add custom error handling or fallback providers here
		// Example: Try a fallback provider or provide offline mode
		os.Exit(1)
	}
	
	// TODO: Add LLM provider validation or health checks here
	// Example: Test the connection with a simple query
	logger.Info().Str("provider", config.AgentFlow.Provider).Msg("LLM provider initialized successfully")

	{{if .Config.MCPEnabled}}
	// Initialize MCP (Model Context Protocol) manager for tool integration
	// MCP allows agents to access external tools and services like file systems,
	// databases, APIs, and other integrations defined in your agentflow.toml
	// TODO: Customize MCP initialization for your specific tool requirements
	// You might want to add custom tool validation, authentication, or configuration
	logger.Info().Msg("🔧 Initializing MCP with timeout handling...")
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
		logger.Info().Msg("MCP manager initialized successfully - agents can access tools via core.GetMCPManager()")

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
				logger.Info().Msg("MCP tool registry initialized successfully")
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
				logger.Info().Msg("MCP tools registered with registry successfully")
			}
		case <-toolsCtx.Done():
			logger.Warn().Msg("MCP tools registration timed out")
		}
	}
	{{end}}

	{{if .Config.MemoryEnabled}}
	// Initialize the memory system for persistent storage and retrieval
	// This enables agents to remember previous conversations, store knowledge,
	// and perform RAG (Retrieval-Augmented Generation) operations
	// TODO: Customize memory initialization for your specific use case
	// You might want to add custom indexing, data preprocessing, or storage optimization
	fmt.Println("🧠 Initializing memory system...")
	
	// Create memory configuration from agentflow.toml settings
	// This includes database connections, embedding models, and RAG parameters
	// TODO: Add custom memory configuration validation or enhancement here
	memoryConfig := config.AgentMemory
	
	// Validate configuration before initializing memory
	fmt.Println("🔍 Validating memory configuration...")
	if err := validateMemoryConfig(memoryConfig, "{{.Config.EmbeddingModel}}"); err != nil {
		logger.Error().Err(err).Msg("Memory configuration validation failed")
		fmt.Printf("❌ Configuration Error: %v\n", err)
		os.Exit(1)
	}
	
	logger.Info().Msg("Memory configuration validation passed")
	fmt.Println("✅ Configuration validated!")
	
	memory, err := core.NewMemory(memoryConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize memory")
		fmt.Printf("Memory initialization failed: %v\n", err)
		
		// Provide specific troubleshooting based on provider
		switch memoryConfig.Provider {
		case "pgvector":
			fmt.Printf("\n💡 PostgreSQL/PgVector Troubleshooting:\n")
			fmt.Printf("   1. Start database: docker compose up -d\n")
			fmt.Printf("   2. Run setup script: ./setup.sh (or setup.bat on Windows)\n")
			fmt.Printf("   3. Check connection string in agentflow.toml\n")
			fmt.Printf("   4. Verify database exists: psql -h localhost -U user -d agentflow\n")
		case "weaviate":
			fmt.Printf("\n💡 Weaviate Troubleshooting:\n")
			fmt.Printf("   1. Start Weaviate: docker compose up -d\n")
			fmt.Printf("   2. Check Weaviate is running: curl http://localhost:8080/v1/meta\n")
			fmt.Printf("   3. Verify connection string in agentflow.toml\n")
		case "memory":
			fmt.Printf("\n💡 In-Memory Provider Issue:\n")
			fmt.Printf("   This shouldn't fail - check your configuration\n")
		}
		
		// Check embedding provider availability
		if memoryConfig.Embedding.Provider == "ollama" {
			fmt.Printf("\n💡 Ollama Troubleshooting:\n")
			fmt.Printf("   1. Start Ollama: ollama serve\n")
			fmt.Printf("   2. Pull model: ollama pull %s\n", memoryConfig.Embedding.Model)
			fmt.Printf("   3. Test connection: curl http://localhost:11434/api/tags\n")
		} else if memoryConfig.Embedding.Provider == "openai" {
			fmt.Printf("\n💡 OpenAI Troubleshooting:\n")
			fmt.Printf("   1. Set API key: export OPENAI_API_KEY=\"your-key\"\n")
			fmt.Printf("   2. Verify key is valid and has credits\n")
		}
		
		os.Exit(1)
	}
	defer memory.Close()
	
	// Test memory connection
	testContent := fmt.Sprintf("System initialized at %s", time.Now().Format("2006-01-02 15:04:05"))
	if err := memory.Store(ctx, testContent, "system-init"); err != nil {
		logger.Warn().Err(err).Msg("Memory connection test failed, continuing anyway")
		fmt.Printf("⚠️  Memory connection test failed: %v\n", err)
		fmt.Printf("Your agents will still work, but memory features may be limited\n")
	} else {
		logger.Info().Msg("Memory system initialized successfully")
		fmt.Printf("✅ Memory system ready!\n")
	}
	{{end}}

	// Create configuration-driven agent factory
	// This factory creates agents based on the configuration in agentflow.toml
	// instead of hardcoded agent constructors, providing much more flexibility
	// TODO: Customize agent factory initialization if needed
	// You might want to add custom agent types or initialization logic
	logger.Info().Msg("🤖 Initializing configuration-driven agent factory...")
	
	factory := core.NewConfigurableAgentFactory(config)
	if factory == nil {
		logger.Error().Msg("Failed to create agent factory")
		fmt.Printf("❌ Error creating agent factory\n")
		os.Exit(1)
	}

	// Create agent manager for centralized agent lifecycle management
	// The agent manager handles agent creation, initialization, and state management
	// TODO: Customize agent manager configuration if needed
	agentManager := core.NewAgentManager(config)
	if err := agentManager.InitializeAgents(); err != nil {
		logger.Error().Err(err).Msg("Failed to initialize agents from configuration")
		fmt.Printf("❌ Error initializing agents: %v\n", err)
		fmt.Printf("💡 Check your agentflow.toml [agents] configuration\n")
		os.Exit(1)
	}

	// Get all active agents from the manager
	// This automatically excludes disabled agents and handles configuration-based filtering
	activeAgents := agentManager.GetActiveAgents()
	logger.Info().Int("active_agents", len(activeAgents)).Msg("Active agents loaded from configuration")

	// Create agent handlers map for the workflow orchestrator
	// We wrap each agent with result collection for output tracking
	agentHandlers := make(map[string]core.AgentHandler)
	results := make([]AgentOutput, 0)
	var resultsMutex sync.Mutex

	// Register all active agents with result collection wrappers
	for _, agent := range activeAgents {
		agentName := agent.GetRole()
		
		// Wrap the agent with result collection for output tracking
		// TODO: Add custom agent middleware here if needed
		// Examples: logging, metrics, rate limiting, caching, authentication
		wrappedAgent := &ResultCollectorHandler{
			originalHandler: agent,
			agentName:       agentName,
			outputs:         &results,
			mutex:           &resultsMutex,
		}
		
		agentHandlers[agentName] = wrappedAgent
		logger.Debug().Str("agent", agentName).Msg("Agent registered with result collection")
	}

	// Validate that we have at least one active agent
	if len(agentHandlers) == 0 {
		logger.Error().Msg("No active agents found in configuration")
		fmt.Printf("❌ No active agents configured\n")
		fmt.Printf("💡 Check your agentflow.toml [agents] section and ensure at least one agent is enabled\n")
		fmt.Printf("💡 Example agent configuration:\n")
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
		
		logger.Debug().Msg("Error handlers registered using first active agent as fallback")
	}

	// Create the workflow orchestrator (runner) using configuration-based setup
	// The runner manages the execution flow between agents based on your orchestration mode
	// (sequential, collaborative, loop, or mixed)
	// TODO: Customize runner creation for advanced orchestration needs
	// You might want to add custom routing logic, middleware, or execution policies
	runner, err := core.NewRunnerFromConfig("agentflow.toml")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create runner from config")
		fmt.Printf("❌ Error creating runner: %v\n", err)
		fmt.Printf("💡 Check your agentflow.toml orchestration configuration\n")
		os.Exit(1)
	}
	
	// Register all agents with the workflow orchestrator
	// This makes the agents available for routing and execution
	// TODO: Add custom agent registration logic if needed
	// Examples: conditional registration, agent prioritization, or custom routing
	for name, handler := range agentHandlers {
		if err := runner.RegisterAgent(name, handler); err != nil {
			logger.Error().Err(err).Str("agent", name).Msg("Failed to register agent")
			fmt.Printf("❌ Error registering agent %s: %v\n", name, err)
			os.Exit(1)
		}
		logger.Debug().Str("agent", name).Msg("Agent registered successfully")
	}
	
	logger.Info().Int("agent_count", len(agentHandlers)).Msg("All agents registered with orchestrator")



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
		logger.Debug().Msg("Using message from command line flag")
	} else {
		// Interactive mode: prompt user for input
		// TODO: Enhance interactive input with better UX
		// Examples: multi-line input, input history, auto-completion
		fmt.Print("💬 Enter your message: ")
		fmt.Scanln(&message)
	}

	// Provide default message if none specified
	// TODO: Customize the default message for your use case
	if message == "" {
		message = "Hello! Please provide information about current topics."
		logger.Debug().Msg("Using default message")
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
		fmt.Printf("❌ Workflow execution error: %v\n", err)
		// TODO: Add custom error recovery or fallback logic here
		os.Exit(1)
	}

	logger.Info().Msg("🚀 Workflow execution started")

	// Wait for processing to complete BEFORE printing results.
	// We call runner.Stop() explicitly here (instead of using defer runner.Stop()).
	// A deferred call would execute only when main() returns—after the result-printing
	// code below—so we could print an empty "Agent Responses" section while the
	// agents are still working.  Calling Stop() now closes the queue and blocks
	// until the event-processing goroutine has finished handling all queued events,
	// guaranteeing the results slice is fully populated.
	logger.Info().Msg("Waiting for agents to complete processing...")
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
			fmt.Printf("\n🤖 %s (Step %d):\n", result.AgentName, i+1)
			fmt.Printf("%s\n", result.Content)
			fmt.Printf("⏰ Completed at: %s\n", result.Timestamp.Format("15:04:05"))
			
			// TODO: Add custom result metadata display here
			// Examples: processing time, confidence scores, source information
		}
		
		// TODO: Add result summary or analytics here
		// Examples: total processing time, success rate, performance metrics
		fmt.Printf("\n📊 Summary: %d agents completed successfully\n", len(results))
	} else {
		fmt.Printf("❌ No agent responses captured. This might indicate:\n")
		fmt.Printf("   • LLM provider credentials are not configured\n")
		fmt.Printf("   • An agent encountered an error during processing\n")
		fmt.Printf("   • The LLM provider is not responding\n")
		fmt.Printf("   • Network connectivity issues\n")
		
		// TODO: Add custom troubleshooting guidance for your specific setup
		fmt.Printf("\n💡 Troubleshooting tips:\n")
		fmt.Printf("   • Check your environment variables\n")
		fmt.Printf("   • Verify agentflow.toml configuration\n")
		fmt.Printf("   • Review the logs above for detailed error information\n")
	}
	resultsMutex.Unlock()

	// TODO: Add custom post-processing here
	// Examples:
	// - Save results to database or file
	// - Send notifications or webhooks
	// - Update metrics or analytics
	// - Trigger follow-up workflows

	fmt.Printf("\n=== Workflow Completed ===\n")
	fmt.Printf("✅ Check the logs above for detailed agent execution results.\n")
	
	// TODO: Add custom completion logic here
	// Examples: cleanup resources, send completion notifications, update status

	logger.Info().Msg("🎉 Workflow completed successfully")
}

{{.ProviderInitFunction}}

{{if .Config.MCPEnabled}}
{{.MCPInitFunction}}
{{end}}

{{if .Config.WithCache}}
{{.CacheInitFunction}}
{{end}}

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

	// Return the original result unchanged so the workflow continues normally
	return result, err
}

{{if .Config.MemoryEnabled}}
// validateMemoryConfig validates the memory configuration against expected values
func validateMemoryConfig(memoryConfig core.AgentMemoryConfig, expectedModel string) error {
	// Validate embedding dimensions
	expectedDimensions := {{.Config.EmbeddingDimensions}}
	if memoryConfig.Dimensions != expectedDimensions {
		return fmt.Errorf("%s requires %d dimensions, but %d configured in agentflow.toml\n💡 Solution: Update [agent_memory] dimensions = %d", 
			expectedModel, expectedDimensions, memoryConfig.Dimensions, expectedDimensions)
	}
	
	// Validate embedding provider and model
	expectedProvider := "{{.Config.EmbeddingProvider}}"
	expectedModelName := "{{.Config.EmbeddingModel}}"
	
	if memoryConfig.Embedding.Provider != expectedProvider {
		return fmt.Errorf("embedding provider mismatch: expected '%s', got '%s'\n💡 Solution: Update [agent_memory.embedding] provider = \"%s\"", 
			expectedProvider, memoryConfig.Embedding.Provider, expectedProvider)
	}
	
	if memoryConfig.Embedding.Model != expectedModelName {
		return fmt.Errorf("embedding model mismatch: expected '%s', got '%s'\n💡 Solution: Update [agent_memory.embedding] model = \"%s\"", 
			expectedModelName, memoryConfig.Embedding.Model, expectedModelName)
	}
	
	// Validate memory provider configuration
	switch memoryConfig.Provider {
	case "pgvector":
		if memoryConfig.Connection == "" {
			return fmt.Errorf("pgvector provider requires a connection string\n💡 Solution: Set [agent_memory] connection = \"postgres://user:password@localhost:15432/agentflow?sslmode=disable\"")
		}
		if !strings.Contains(memoryConfig.Connection, "postgres://") {
			return fmt.Errorf("pgvector connection string should start with 'postgres://'\n💡 Current: %s", memoryConfig.Connection)
		}
	case "weaviate":
		if memoryConfig.Connection == "" {
			return fmt.Errorf("weaviate provider requires a connection string\n💡 Solution: Set [agent_memory] connection = \"http://localhost:8080\"")
		}
		if !strings.Contains(memoryConfig.Connection, "http") {
			return fmt.Errorf("weaviate connection string should be an HTTP URL\n💡 Current: %s", memoryConfig.Connection)
		}
	case "memory":
		// In-memory provider doesn't need connection validation
	default:
		return fmt.Errorf("unknown memory provider: %s\n💡 Valid options: memory, pgvector, weaviate", memoryConfig.Provider)
	}
	
	// Validate RAG configuration if enabled
	{{if .Config.RAGEnabled}}
	if memoryConfig.EnableRAG {
		if memoryConfig.ChunkSize <= 0 {
			return fmt.Errorf("RAG chunk size must be positive, got %d\n💡 Solution: Set [agent_memory] chunk_size = 1000", memoryConfig.ChunkSize)
		}
		if memoryConfig.ChunkOverlap < 0 || memoryConfig.ChunkOverlap >= memoryConfig.ChunkSize {
			return fmt.Errorf("RAG chunk overlap must be between 0 and chunk_size (%d), got %d\n💡 Solution: Set [agent_memory] chunk_overlap = 100", 
				memoryConfig.ChunkSize, memoryConfig.ChunkOverlap)
		}
		if memoryConfig.KnowledgeScoreThreshold < 0.0 || memoryConfig.KnowledgeScoreThreshold > 1.0 {
			return fmt.Errorf("RAG score threshold must be between 0.0 and 1.0, got %.2f\n💡 Solution: Set [agent_memory] knowledge_score_threshold = 0.7", 
				memoryConfig.KnowledgeScoreThreshold)
		}
	}
	{{end}}
	
	// Validate embedding provider specific settings
	switch memoryConfig.Embedding.Provider {
	case "ollama":
		if memoryConfig.Embedding.BaseURL == "" {
			return fmt.Errorf("ollama embedding provider requires base_url\n💡 Solution: Set [agent_memory.embedding] base_url = \"http://localhost:11434\"")
		}
	case "openai":
		// OpenAI uses environment variables, so we can't validate API key here
		// But we can check if the model name looks reasonable
		if !strings.Contains(memoryConfig.Embedding.Model, "embedding") {
			return fmt.Errorf("OpenAI model '%s' doesn't look like an embedding model\n💡 Recommended: text-embedding-3-small or text-embedding-3-large", 
				memoryConfig.Embedding.Model)
		}
	case "dummy":
		// Dummy provider doesn't need validation
	default:
		return fmt.Errorf("unknown embedding provider: %s\n💡 Valid options: openai, ollama, dummy", memoryConfig.Embedding.Provider)
	}
	
	return nil
}
{{end}}
`
