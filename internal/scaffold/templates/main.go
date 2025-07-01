package templates

const MainTemplate = `package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
)

func main() {
	ctx := context.Background()
	core.SetLogLevel(core.INFO)
	logger := core.Logger()
	logger.Info().Msg("Starting {{.Config.Name}} multi-agent system...")

	messageFlag := flag.String("m", "", "Message to process")
	flag.Parse()

	llmProvider, err := initializeProvider("{{.Config.Provider}}")
	if err != nil {
		fmt.Printf("Failed to initialize LLM provider: %v\n", err)
		os.Exit(1)
	}

	{{if .Config.MCPEnabled}}
	// Initialize MCP manager for tool integration with timeout handling
	logger.Info().Msg("Initializing MCP with timeout handling...")
	mcpInitCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var mcpManager core.MCPManager
	mcpDone := make(chan bool, 1)
	var mcpErr error

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

	agents := make(map[string]core.AgentHandler)

	{{range .Agents}}
	// Create {{.DisplayName}} handler
	{{.Name}} := New{{.DisplayName}}(llmProvider)
	agents["{{.Name}}"] = {{.Name}}
	{{end}}

	// Create basic error handlers to prevent routing errors
	// These use the first agent as a fallback handler for simplicity
	{{if .Agents}}
	firstAgent := agents["{{(index .Agents 0).Name}}"]
	if firstAgent != nil {
		agents["error-handler"] = firstAgent
		agents["validation-error-handler"] = firstAgent
		agents["timeout-error-handler"] = firstAgent
		agents["critical-error-handler"] = firstAgent
		agents["high-priority-error-handler"] = firstAgent
		agents["network-error-handler"] = firstAgent
		agents["llm-error-handler"] = firstAgent
		agents["auth-error-handler"] = firstAgent
	}
	{{end}}

	// Create orchestrated runner
	{{if eq .Config.OrchestrationMode "collaborative"}}
	runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
	{{else if eq .Config.OrchestrationMode "sequential"}}
	runner, err := core.NewRunnerWithOrchestration(
		core.SequentialOrchestrator{
			AgentSequence: []string{
				{{- range $i, $agent := .Agents}}
				{{- if $i}}, {{end}}"{{$agent.Name}}"
				{{- end}}
			},
		},
		agents,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create sequential runner")
	}
	{{else if eq .Config.OrchestrationMode "loop"}}
	runner, err := core.NewRunnerWithOrchestration(
		core.LoopOrchestrator{
			Agent:         "{{(index .Agents 0).Name}}",
			MaxIterations: {{.Config.MaxIterations}},
		},
		agents,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create loop runner")
	}
	{{else if eq .Config.OrchestrationMode "mixed"}}
	// TODO: Implement true mixed mode orchestration
	runner, err := core.NewRunnerWithOrchestration(
		core.SequentialOrchestrator{
			AgentSequence: []string{
				{{- range $i, $agent := .Agents}}
				{{- if $i}}, {{end}}"{{$agent.Name}}"
				{{- end}}
			},
		},
		agents,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create mixed runner")
	}
	{{else}}
	// Default collaborative mode
	runner := core.CreateCollaborativeRunner(agents, 30*time.Second)
	{{end}}

	// Result collection system
	var agentOutputs []AgentOutput
	var outputMutex sync.Mutex

	{{if eq .Config.OrchestrationMode "collaborative"}}
	// Create a result collector by wrapping the existing agents (only for collaborative mode)
	for name, handler := range agents {
		if strings.Contains(name, "error-handler") {
			continue // Skip error handlers for result collection
		}

		// Wrap the original handler to capture outputs
		originalHandler := handler
		wrappedHandler := &ResultCollectorHandler{
			originalHandler: originalHandler,
			agentName:       name,
			outputs:         &agentOutputs,
			mutex:           &outputMutex,
		}
		agents[name] = wrappedHandler
	}

	// Recreate runner with wrapped agents
	runner = core.CreateCollaborativeRunner(agents, 30*time.Second)
	{{end}}

	var message string
	if *messageFlag != "" {
		message = *messageFlag
	} else {
		fmt.Print("Enter your message: ")
		fmt.Scanln(&message)
	}

	if message == "" {
		message = "Hello! Please provide information about current topics."
	}

	logger.Info().Str("message", message).Msg("Processing user message")

	{{if eq .Config.OrchestrationMode "collaborative"}}
	// Start the collaborative runner
	runner.Start(ctx)
	defer runner.Stop()

	{{if .Agents}}
	event := core.NewEvent("{{(index .Agents 0).Name}}", core.EventData{
		"message": message,
	}, map[string]string{
		"route": "{{(index .Agents 0).Name}}",
	})
	{{else}}
	event := core.NewEvent("user_request", core.EventData{
		"message": message,
	}, map[string]string{
		"route": "user_request",
	})
	{{end}}

	if err := runner.Emit(event); err != nil {
		logger.Error().Err(err).Msg("Workflow execution failed")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Extended wait to allow all agents to complete their work
	logger.Info().Msg("Waiting for agents to complete processing...")
	time.Sleep(10 * time.Second)

	// Display collected agent outputs
	outputMutex.Lock()
	if len(agentOutputs) > 0 {
		fmt.Printf("\n=== Agent Results ===\n")
		for _, output := range agentOutputs {
			fmt.Printf("\n[%s] %s:\n", output.Timestamp.Format("15:04:05"), output.AgentName)
			fmt.Printf("%s\n", output.Content)
			fmt.Printf("%s\n", strings.Repeat("-", 50))
		}
	} else {
		logger.Debug().Msg("No agent outputs captured")
	}
	outputMutex.Unlock()

	logger.Info().Msg("Workflow completed successfully")
	{{else}}
	// Create event and run workflow for traditional orchestration
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
	{{end}}

	logger.Info().Msg("Workflow completed successfully")
}

{{.ProviderInitFunction}}

{{if .Config.MCPEnabled}}
{{.MCPInitFunction}}
{{end}}

{{if .Config.WithCache}}
{{.CacheInitFunction}}
{{end}}

// ResultCollectorHandler wraps an agent handler to capture its outputs
type ResultCollectorHandler struct {
	originalHandler core.AgentHandler
	agentName       string
	outputs         *[]AgentOutput
	mutex           *sync.Mutex
}

// AgentOutput holds the output from an agent
type AgentOutput struct {
	AgentName string
	Content   string
	Timestamp time.Time
}

// Run implements the AgentHandler interface and captures the output
func (r *ResultCollectorHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// Call the original handler
	result, err := r.originalHandler.Run(ctx, event, state)

	// Extract meaningful content from the result
	var content string
	if err != nil {
		content = fmt.Sprintf("Error: %v", err)
	} else if result.Error != "" {
		content = fmt.Sprintf("Agent Error: %s", result.Error)
	} else {
		// Try to extract content from the result's output state
		if result.OutputState != nil {
			if responseData, exists := result.OutputState.Get("response"); exists {
				if responseStr, ok := responseData.(string); ok {
					content = responseStr
				}
			}
			if content == "" {
				if outputData, exists := result.OutputState.Get("output"); exists {
					if outputStr, ok := outputData.(string); ok {
						content = outputStr
					}
				}
			}
			if content == "" {
				if messageData, exists := result.OutputState.Get("message"); exists {
					if messageStr, ok := messageData.(string); ok {
						content = messageStr
					}
				}
			}
		}
	}

	// If we still don't have content, create a summary
	if content == "" {
		content = fmt.Sprintf("Agent %s completed processing successfully", r.agentName)
	}

	// Store the output
	r.mutex.Lock()
	*r.outputs = append(*r.outputs, AgentOutput{
		AgentName: r.agentName,
		Content:   content,
		Timestamp: time.Now(),
	})
	r.mutex.Unlock()

	return result, err
}
`
