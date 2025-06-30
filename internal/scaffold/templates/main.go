package templates

const MainTemplate = `package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
)

func main() {
	ctx := context.Background()

	// Configure AgentFlow logging level
	core.SetLogLevel(core.INFO)
	logger := core.Logger()
	logger.Info().Msg("Starting unified multi-agent system...")

	// Parse command line flags
	messageFlag := flag.String("m", "", "Message to process")
	flag.Parse()

	// Initialize LLM provider
	llmProvider, err := initializeProvider("{{.Config.Provider}}")
	if err != nil {
		fmt.Printf("Failed to initialize LLM provider: %v\n", err)
		os.Exit(1)
	}

	{{if .Config.MCPEnabled}}
	// Initialize MCP manager for tool integration using production APIs
	// Note: In this orchestrator pattern, agents use the global MCP instance
	// The mcpManager variable demonstrates MCP availability but agents access it via core.GetMCPManager()
	mcpManager, err := initializeMCP()
	if err != nil {
		logger.Warn().Err(err).Msg("MCP initialization failed, continuing without MCP")
		mcpManager = nil
	}
	if mcpManager != nil {
		logger.Info().Msg("MCP manager initialized successfully - agents can access tools via core.GetMCPManager()")

		// Initialize MCP tool registry
		if err := core.InitializeMCPToolRegistry(); err != nil {
			logger.Warn().Err(err).Msg("Failed to initialize MCP tool registry")
		} else {
			logger.Info().Msg("MCP tool registry initialized successfully")
		}

		// Register MCP tools with the registry so agents can use them
		if err := core.RegisterMCPToolsWithRegistry(ctx); err != nil {
			logger.Warn().Err(err).Msg("Failed to register MCP tools with registry")
		} else {
			logger.Info().Msg("MCP tools registered with registry successfully")
		}
	}
	{{end}}

	// Create agents using agent handlers (not unified builder for orchestrator use)
	var wg sync.WaitGroup
	agents := make(map[string]core.AgentHandler)

	{{range .Agents}}
	// Create {{.DisplayName}} handler
	{{.Name}} := New{{.DisplayName}}(llmProvider)
	agents["{{.Name}}"] = {{.Name}}
	{{end}}

	{{if .Config.ResponsibleAI}}
	// Create Responsible AI handler
	responsibleAI := NewResponsibleAIHandler(llmProvider)
	agents["responsible_ai"] = responsibleAI
	{{end}}

	{{if .Config.ErrorHandler}}
	// Create Error Handler
	errorHandler := NewErrorHandler(llmProvider)
	agents["error_handler"] = errorHandler
	agents["error-handler"] = errorHandler // Alias for hyphen-separated routing
	{{end}}

	// Create orchestrated runner
	{{if eq .Config.OrchestrationMode "collaborative"}}
	runner, err := core.NewRunnerWithOrchestration(
		core.CollaborativeOrchestrator{
			Agents:           []string{{{range $i, $agent := .Agents}}{{if $i}}, {{end}}"{{$agent.Name}}"{{end}}},
			FailureThreshold: {{.Config.FailureThreshold}},
			MaxConcurrency:   {{.Config.MaxConcurrency}},
		},
		agents,
	)
	{{else if eq .Config.OrchestrationMode "sequential"}}
	runner, err := core.NewRunnerWithOrchestration(
		core.SequentialOrchestrator{
			AgentSequence: []string{{{range $i, $agent := .Agents}}{{if $i}}, {{end}}"{{$agent.Name}}"{{end}}},
		},
		agents,
	)
	{{else if eq .Config.OrchestrationMode "loop"}}
	runner, err := core.NewRunnerWithOrchestration(
		core.LoopOrchestrator{
			Agent:         "{{(index .Agents 0).Name}}",
			MaxIterations: {{.Config.MaxIterations}},
		},
		agents,
	)
	{{else if eq .Config.OrchestrationMode "mixed"}}
	// TODO: Implement true mixed mode orchestration
	runner, err := core.NewRunnerWithOrchestration(
		core.SequentialOrchestrator{
			AgentSequence: []string{{{range $i, $agent := .Agents}}{{if $i}}, {{end}}"{{$agent.Name}}"{{end}}},
		},
		agents,
	)
	{{else}}
	// Default route mode - use sequential for simplicity
	runner, err := core.NewRunnerWithOrchestration(
		core.SequentialOrchestrator{
			AgentSequence: []string{{{range $i, $agent := .Agents}}{{if $i}}, {{end}}"{{$agent.Name}}"{{end}}},
		},
		agents,
	)
	{{end}}

	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create orchestrated runner")
	}

	// Get user input
	var message string
	if *messageFlag != "" {
		message = *messageFlag
	} else {
		fmt.Print("Enter your message: ")
		fmt.Scanln(&message)
	}

	if message == "" {
		message = "Hello! Please provide information about the current stock market trends."
	}

	logger.Info().Str("message", message).Msg("Processing user message")

	// Create event and run workflow
	event := core.NewEvent("user_request", map[string]interface{}{
		"message":   message,
		"timestamp": time.Now(),
	})

	result, err := runner.Run(ctx, event)
	if err != nil {
		logger.Error().Err(err).Msg("Workflow execution failed")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Display results
	fmt.Printf("\n=== Workflow Results ===\n")
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Content: %s\n", result.Content)

	if len(result.Metrics) > 0 {
		fmt.Printf("\n=== Metrics ===\n")
		for key, value := range result.Metrics {
			fmt.Printf("%s: %v\n", key, value)
		}
	}

	logger.Info().Msg("Workflow completed successfully")
}

{{.ProviderInitFunction}}

{{if .Config.MCPEnabled}}
{{.MCPInitFunction}}
{{end}}

{{if .Config.WithCache}}
{{.CacheInitFunction}}
{{end}}
`
