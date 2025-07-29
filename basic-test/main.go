package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"basic-test/agents"
)



// main is the entry point for the basic-test multi-agent system.
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
	logger.Info().Msg("Starting basic-test multi-agent system...")

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

	

	

	// Create agent handlers from the agents package
	// Each agent is responsible for a specific part of your workflow
	// TODO: Customize agent creation and configuration here
	// You might want to add:
	// - Custom initialization parameters for each agent
	// - Agent-specific configuration or dependencies
	// - Custom middleware or decorators for agents
	// - Agent health checks or validation
	agentHandlers := make(map[string]core.AgentHandler)
	results := make([]AgentOutput, 0)
	var resultsMutex sync.Mutex

	
	// Create Agent1 handler with result collection
	// TODO: Customize Agent1 initialization if needed
	// You can pass additional dependencies or configuration to the agent constructor
	
	agent1 := agents.NewAgent1(llmProvider)
	
	
	// Wrap the agent with result collection for output tracking
	// TODO: Add custom agent middleware here if needed
	// Examples: logging, metrics, rate limiting, caching
	wrappedAgent1 := &ResultCollectorHandler{
		originalHandler: agent1,
		agentName:       "agent1",
		outputs:         &results,
		mutex:           &resultsMutex,
	}
	agentHandlers["agent1"] = wrappedAgent1
	
	// Create Agent2 handler with result collection
	// TODO: Customize Agent2 initialization if needed
	// You can pass additional dependencies or configuration to the agent constructor
	
	agent2 := agents.NewAgent2(llmProvider)
	
	
	// Wrap the agent with result collection for output tracking
	// TODO: Add custom agent middleware here if needed
	// Examples: logging, metrics, rate limiting, caching
	wrappedAgent2 := &ResultCollectorHandler{
		originalHandler: agent2,
		agentName:       "agent2",
		outputs:         &results,
		mutex:           &resultsMutex,
	}
	agentHandlers["agent2"] = wrappedAgent2
	

	// TODO: Add custom agent registration logic here
	// Example: Register agents conditionally based on configuration or environment

	// Create basic error handlers to prevent routing errors
	// These use the first agent as a fallback handler for simplicity
	
	firstAgent := agentHandlers["agent1"]
	if firstAgent != nil {
		agentHandlers["error-handler"] = firstAgent
		agentHandlers["validation-error-handler"] = firstAgent
		agentHandlers["timeout-error-handler"] = firstAgent
		agentHandlers["critical-error-handler"] = firstAgent
		agentHandlers["high-priority-error-handler"] = firstAgent
		agentHandlers["network-error-handler"] = firstAgent
		agentHandlers["llm-error-handler"] = firstAgent
		agentHandlers["auth-error-handler"] = firstAgent
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
	
	event := core.NewEvent("agent1", core.EventData{
		"message": message,
		// TODO: Add custom event data here
		// Examples:
		// "timestamp": time.Now(),
		// "user_id": userID,
		// "session_id": sessionID,
		// "priority": "normal",
	}, map[string]string{
		"route": "agent1",
		// TODO: Add custom routing metadata here
		// Examples:
		// "workflow_type": "standard",
		// "execution_mode": "async",
	})
	

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

func initializeProvider(providerType string) (core.ModelProvider, error) {
	// Use the config-based provider initialization
	return core.NewProviderFromWorkingDir()
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

	// Return the original result unchanged so the workflow continues normally
	return result, err
}


