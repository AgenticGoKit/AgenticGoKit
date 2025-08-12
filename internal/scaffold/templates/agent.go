package templates

const AgentTemplate = `// Package agents contains the agent implementations for this project.
// 
// This package provides configuration-aware agent implementations that are
// automatically created and managed by the AgentFlow configuration system.
// 
// IMPORTANT: This package is now primarily for reference and custom agent types.
// The main application uses the ConfigurableAgentFactory to create agents
// directly from the agentflow.toml configuration file.
//
// TODO: Use this package to implement custom agent types that extend the
// standard configuration-driven agents with specialized business logic.
package agents

import (
	"context"
	"fmt"
	"strings"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/core"
)

// {{.Agent.DisplayName}}Handler represents a custom {{.Agent.Name}} agent handler.
// 
// Purpose: {{.Agent.Purpose}}
//
// NOTE: This is a reference implementation. The main application now uses
// configuration-driven agents created by ConfigurableAgentFactory.
// 
// This custom implementation is provided for:
// - Reference and learning purposes
// - Custom agent types that need specialized logic
// - Migration from hardcoded to configuration-driven agents
//
// TODO: If you need custom logic beyond what the configuration system provides,
// you can modify this agent and register it as a custom agent type.
type {{.Agent.DisplayName}}Handler struct {
	config agenticgokit.ResolvedAgentConfig
	llm    agenticgokit.ModelProvider
	{{if .Config.MemoryEnabled}}memory agenticgokit.Memory{{end}}
	
	// TODO: Add your custom fields here
	// Examples:
	// database    *sql.DB
	// apiClient   *http.Client
	// customConfig *YourConfig
}

// New{{.Agent.DisplayName}} creates a new {{.Agent.DisplayName}} instance.
//
// NOTE: This constructor is for reference. The main application uses
// ConfigurableAgentFactory.CreateAgent() to create agents from configuration.
//
// This constructor shows how to create configuration-aware agents manually
// if you need custom initialization logic.
func New{{.Agent.DisplayName}}FromConfig(config agenticgokit.ResolvedAgentConfig) (*{{.Agent.DisplayName}}Handler, error) {
	// Initialize LLM provider from resolved configuration
	llm, err := agenticgokit.NewLLMProvider(config.LLM)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM provider: %w", err)
	}

	{{if .Config.MemoryEnabled}}
	// Get memory instance from global context
	memory := agenticgokit.GetGlobalMemory()
	if memory == nil {
		return nil, fmt.Errorf("memory system not initialized")
	}
	{{end}}

	return &{{.Agent.DisplayName}}Handler{
		config: config,
		llm:    llm,
		{{if .Config.MemoryEnabled}}memory: memory,{{end}}
		
		// TODO: Initialize your custom fields here
		// Examples:
		// database:  initDatabase(),
		// apiClient: &http.Client{Timeout: 30 * time.Second},
		// customConfig: loadCustomConfig(),
	}, nil
}

// GetRole returns the agent's role from configuration
func (a *{{.Agent.DisplayName}}Handler) GetRole() string {
	return a.config.Role
}

// GetCapabilities returns the agent's capabilities from configuration
func (a *{{.Agent.DisplayName}}Handler) GetCapabilities() []string {
	return a.config.Capabilities
}

// IsEnabled returns whether the agent is enabled from configuration
func (a *{{.Agent.DisplayName}}Handler) IsEnabled() bool {
	return a.config.Enabled
}

// Run implements the agenticgokit.AgentHandler interface.
//
// This method demonstrates how to implement a configuration-aware agent.
// It uses the resolved configuration for system prompts, LLM settings, and capabilities.
//
// NOTE: This is a reference implementation. The main application uses
// configuration-driven agents that are automatically created and don't need
// this custom implementation unless you have specialized business logic.
//
// TODO: Customize this method only if you need logic beyond what the
// configuration system provides.
func (a *{{.Agent.DisplayName}}Handler) Run(ctx context.Context, event agenticgokit.Event, state agenticgokit.State) (agenticgokit.AgentResult, error) {
	// Get logger for debug output
	logger := agenticgokit.Logger()
	logger.Debug().
		Str("agent", a.config.Role).
		Str("event_id", event.GetID()).
		Bool("enabled", a.config.Enabled).
		Msg("Configuration-aware agent processing started")
	
	// Check if agent is enabled in configuration
	if !a.config.Enabled {
		logger.Debug().Str("agent", a.config.Role).Msg("Agent is disabled in configuration, skipping")
		return agenticgokit.AgentResult{
			OutputState: agenticgokit.NewState(map[string]interface{}{
				"response": fmt.Sprintf("Agent %s is disabled in configuration", a.config.Role),
				"skipped":  true,
			}),
		}, nil
	}
	
	// TODO: Add custom capability checks here
	// Example: Validate that the agent has required capabilities for this task
	// requiredCapability := "processing"
	// if !contains(a.config.Capabilities, requiredCapability) {
	//     return agenticgokit.AgentResult{}, fmt.Errorf("agent lacks required capability: %s", requiredCapability)
	// }
	
	// Use system prompt from configuration instead of hardcoded prompt
	systemPrompt := a.config.SystemPrompt
	if systemPrompt == "" {
		// Fallback system prompt if not configured
		systemPrompt = fmt.Sprintf("You are %s, a helpful AI assistant.", a.config.Role)
	}
	
	// TODO: Customize input processing logic
	// This section determines what input the agent will process.
	var inputToProcess interface{}
	
	{{if .IsFirstAgent}}
	// {{.Agent.DisplayName}} processes the original input message
	// TODO: Customize how the first agent extracts input from events
	// You might want to validate the input format or extract specific fields
	eventData := event.GetData()
	if msg, ok := eventData["message"]; ok {
		inputToProcess = msg
		// TODO: Add input validation here
		// Example: if !isValidInput(msg) { return error }
	} else if stateMessage, exists := state.Get("message"); exists {
		inputToProcess = stateMessage
	} else {
		inputToProcess = "No message provided"
		// TODO: Consider if this default is appropriate for your use case
	}
	
	// TODO: Customize the system prompt for your specific use case
	// This prompt defines the agent's role and behavior
	systemPrompt = ` + "`{{.SystemPrompt}}`" + `
	logger.Debug().Str("agent", "{{.Agent.Name}}").Interface("input", inputToProcess).Msg("Processing original message")
	{{else}}
	// Sequential processing: Use previous agent's output, with fallback chain
	// TODO: Customize how this agent processes input from previous agents
	// You might want to:
	// - Validate the format of previous agent outputs
	// - Combine outputs from multiple previous agents
	// - Transform the input before processing
	// - Add error handling for malformed previous outputs
	found := false
	agents := []string{{"{"}}{{range $i, $agent := .Agents}}{{if gt $i 0}}, {{end}}"{{$agent.Name}}"{{end}}{{"}"}}
	
	for i := {{.AgentIndex}} - 1; i >= 0; i-- {
		if i < len(agents) {
			prevAgentName := agents[i]
			if agentResponse, exists := state.Get(fmt.Sprintf("%s_response", prevAgentName)); exists {
				inputToProcess = agentResponse
				// TODO: Add validation or transformation of previous agent output
				// Example: inputToProcess = validateAndTransform(agentResponse)
				logger.Debug().Str("agent", "{{.Agent.Name}}").Str("source_agent", prevAgentName).Interface("input", agentResponse).Msg("Processing previous agent's output")
				found = true
				break
			}
		}
	}
	
	if !found {
		// Final fallback to original message
		// TODO: Consider if this fallback behavior is appropriate for your use case
		eventData := event.GetData()
		if msg, ok := eventData["message"]; ok {
			inputToProcess = msg
		} else if stateMessage, exists := state.Get("message"); exists {
			inputToProcess = stateMessage
		} else {
			inputToProcess = "No message provided"
		}
		logger.Debug().Str("agent", "{{.Agent.Name}}").Interface("input", inputToProcess).Msg("Processing original message (final fallback)")
	}
	
	// TODO: Customize the system prompt for your specific use case
	systemPrompt = ` + "`{{.SystemPrompt}}`" + `
	{{end}}
	
	// Get available MCP tools to include in prompt
	var toolsPrompt string
	mcpManager := agenticgokit.GetMCPManager()
	if mcpManager != nil {
		availableTools := mcpManager.GetAvailableTools()
		logger.Debug().Str("agent", "{{.Agent.Name}}").Int("tool_count", len(availableTools)).Msg("MCP Tools discovered")
		toolsPrompt = agenticgokit.FormatToolsPromptForLLM(availableTools)
	} else {
		logger.Warn().Str("agent", "{{.Agent.Name}}").Msg("MCP Manager is not available")
	}
	
	{{if .Config.MemoryEnabled}}
	// Memory system integration with error handling
	var memoryContext string
	if a.memory != nil {
		logger.Debug().Str("agent", "{{.Agent.Name}}").Msg("Building memory context")
		
		{{if .Config.SessionMemory}}
		// Create or get session context with validation
		sessionID := a.memory.NewSession()
		if sessionID == "" {
			logger.Warn().Str("agent", "{{.Agent.Name}}").Msg("Failed to create session ID, continuing without session context")
		} else {
			ctx = a.memory.SetSession(ctx, sessionID)
			logger.Debug().Str("agent", "{{.Agent.Name}}").Str("session_id", sessionID).Msg("Session context created")
		}
		{{end}}
		
		{{if .Config.RAGEnabled}}
		// Build RAG context from knowledge base with error handling
		ragContext, err := a.memory.BuildContext(ctx, fmt.Sprintf("%v", inputToProcess),
			agenticgokit.WithMaxTokens({{.Config.RAGChunkSize}}),
			agenticgokit.WithIncludeSources(true))
		if err != nil {
			logger.Warn().Str("agent", "{{.Agent.Name}}").Err(err).Msg("Failed to build RAG context - continuing without knowledge base context")
		} else if ragContext != nil && ragContext.ContextText != "" {
			memoryContext = fmt.Sprintf("\n\nRelevant Context from Knowledge Base:\n%s", ragContext.ContextText)
			logger.Debug().Str("agent", "{{.Agent.Name}}").Int("context_tokens", ragContext.TokenCount).Msg("RAG context built successfully")
		} else {
			logger.Debug().Str("agent", "{{.Agent.Name}}").Msg("No relevant knowledge base context found")
		}
		{{end}}
		
		// Query relevant memories with error handling
		memoryResults, err := a.memory.Query(ctx, fmt.Sprintf("%v", inputToProcess), {{.Config.RAGTopK}})
		if err != nil {
			logger.Warn().Str("agent", "{{.Agent.Name}}").Err(err).Msg("Failed to query memories - continuing without memory context")
		} else if len(memoryResults) > 0 {
			memoryContext += "\n\nRelevant Memories:\n"
			for i, result := range memoryResults {
				if result.Score >= {{.Config.RAGScoreThreshold}} {
					memoryContext += fmt.Sprintf("%d. %s (score: %.3f)\n", i+1, result.Content, result.Score)
				}
			}
			logger.Debug().Str("agent", "{{.Agent.Name}}").Int("memory_count", len(memoryResults)).Msg("Memory context retrieved")
		} else {
			logger.Debug().Str("agent", "{{.Agent.Name}}").Msg("No relevant memories found")
		}
		
		// Get chat history with error handling
		chatHistory, err := a.memory.GetHistory(ctx, 3)
		if err != nil {
			logger.Warn().Str("agent", "{{.Agent.Name}}").Err(err).Msg("Failed to get chat history - continuing without history context")
		} else if len(chatHistory) > 0 {
			memoryContext += "\n\nRecent Chat History:\n"
			for _, msg := range chatHistory {
				memoryContext += fmt.Sprintf("[%s] %s\n", msg.Role, msg.Content)
			}
			logger.Debug().Str("agent", "{{.Agent.Name}}").Int("history_count", len(chatHistory)).Msg("Chat history retrieved")
		} else {
			logger.Debug().Str("agent", "{{.Agent.Name}}").Msg("No chat history available")
		}
	} else {
		logger.Warn().Str("agent", "{{.Agent.Name}}").Msg("Memory system not available - continuing without memory context")
	}
	{{end}}
	
	// TODO: Customize prompt construction
	// This is where you can modify how the prompt is built for the LLM.
	// You might want to:
	// - Add domain-specific context or instructions
	// - Format the input in a specific way
	// - Include additional metadata or constraints
	// - Add examples or few-shot learning prompts
	userPrompt := fmt.Sprintf("User query: %v", inputToProcess)
	userPrompt += toolsPrompt
	{{if .Config.MemoryEnabled}}
	userPrompt += memoryContext
	{{end}}
	
	// TODO: Add your custom prompt enhancements here
	// Examples:
	// userPrompt += "\n\nAdditional context: " + yourCustomContext
	// userPrompt += "\n\nConstraints: " + yourConstraints
	// userPrompt += "\n\nExpected output format: " + yourFormat
	
	prompt := agenticgokit.Prompt{
		System: systemPrompt,
		User:   userPrompt,
	}
	
	// Debug: Log the full prompt being sent to LLM
	logger.Debug().Str("agent", "{{.Agent.Name}}").Str("system_prompt", systemPrompt).Str("user_prompt", userPrompt).Msg("Full LLM prompt")
	
	// TODO: Add pre-processing logic here if needed
	// Example: Validate prompt length, add rate limiting, etc.
	
	// Call LLM to get initial response and potential tool calls
	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		// TODO: Customize error handling for LLM failures
		// You might want to add retry logic, fallback responses, or
		// specific error categorization for your use case
		return agenticgokit.AgentResult{}, fmt.Errorf("{{.Agent.DisplayName}} LLM call failed: %w", err)
	}
	
	logger.Debug().Str("agent", "{{.Agent.Name}}").Str("response", response.Content).Msg("Initial LLM response received")
	
	// Parse LLM response for tool calls using core function
	toolCalls := agenticgokit.ParseLLMToolCalls(response.Content)
	var mcpResults []string
	
	// Debug: Log the LLM response to see tool call format
	logger.Debug().Str("agent", "{{.Agent.Name}}").Str("llm_response", response.Content).Msg("LLM response for tool call analysis")
	logger.Debug().Str("agent", "{{.Agent.Name}}").Interface("parsed_tool_calls", toolCalls).Msg("Parsed tool calls from LLM response")
	
	// Execute any requested tools
	if len(toolCalls) > 0 && mcpManager != nil {
		logger.Info().Str("agent", "{{.Agent.Name}}").Int("tool_calls", len(toolCalls)).Msg("Executing LLM-requested tools")
		
		for _, toolCall := range toolCalls {
			if toolName, ok := toolCall["name"].(string); ok {
				var args map[string]interface{}
				if toolArgs, exists := toolCall["args"]; exists {
					if argsMap, ok := toolArgs.(map[string]interface{}); ok {
						args = argsMap
					} else {
						args = make(map[string]interface{})
					}
				} else {
					args = make(map[string]interface{})
				}
				
				logger.Info().Str("agent", "{{.Agent.Name}}").Str("tool_name", toolName).Interface("args", args).Msg("Executing tool as requested by LLM")
				
				// Execute tool using the global ExecuteMCPTool function
				result, err := agenticgokit.ExecuteMCPTool(ctx, toolName, args)
				if err != nil {
					logger.Error().Str("agent", "{{.Agent.Name}}").Str("tool_name", toolName).Err(err).Msg("Tool execution failed")
					mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' failed: %v", toolName, err))
				} else {
					if result.Success {
						logger.Info().Str("agent", "{{.Agent.Name}}").Str("tool_name", toolName).Msg("Tool execution successful")
						
						// Format the result content
						var resultContent string
						if len(result.Content) > 0 {
							resultContent = result.Content[0].Text
						} else {
							resultContent = "Tool executed successfully but returned no content"
						}
						
						mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' result: %s", toolName, resultContent))
					} else {
						logger.Error().Str("agent", "{{.Agent.Name}}").Str("tool_name", toolName).Msg("Tool execution was not successful")
						mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' was not successful", toolName))
					}
				}
			}
		}
	} else {
		logger.Debug().Str("agent", "{{.Agent.Name}}").Msg("No tool calls requested or MCP manager not available")
	}
	
	// Generate final response if tools were used
	var finalResponse string
	if len(mcpResults) > 0 {
		// Create enhanced prompt with tool results
		enhancedPrompt := agenticgokit.Prompt{
			System: systemPrompt,
			User:   fmt.Sprintf("Original query: %v\n\nTool results:\n%s\n\nPlease provide a comprehensive response incorporating these tool results:", inputToProcess, strings.Join(mcpResults, "\n")),
		}
		
		// Get final response from LLM
		finalLLMResponse, err := a.llm.Call(ctx, enhancedPrompt)
		if err != nil {
			return agenticgokit.AgentResult{}, fmt.Errorf("{{.Agent.DisplayName}} final LLM call failed: %w", err)
		}
		finalResponse = finalLLMResponse.Content
		logger.Info().Str("agent", "{{.Agent.Name}}").Str("final_response", finalResponse).Msg("Final response generated with tool results")
	} else {
		finalResponse = response.Content
		logger.Debug().Str("agent", "{{.Agent.Name}}").Msg("Using initial LLM response (no tools used)")
	}
	
	// TODO: Customize response processing and output formatting
	// This is where you can modify the final response before returning it.
	// You might want to:
	// - Format the response in a specific structure
	// - Add metadata or timestamps
	// - Validate the response format
	// - Transform the response for downstream agents
	// - Add custom logging or metrics
	
	// TODO: Add your custom response processing here
	// Examples:
	// finalResponse = formatResponse(finalResponse)
	// finalResponse = addMetadata(finalResponse, metadata)
	// if !isValidResponse(finalResponse) { return error }
	
	// Store agent response in state for potential use by subsequent agents
	outputState := agenticgokit.NewState()
	outputState.Set("{{.Agent.Name}}_response", finalResponse)
	outputState.Set("message", finalResponse)
	
	// TODO: Add custom state data here if needed
	// Examples:
	// outputState.Set("{{.Agent.Name}}_metadata", yourMetadata)
	// outputState.Set("{{.Agent.Name}}_confidence", confidenceScore)
	// outputState.Set("processing_time", processingTime)
	
	{{if .Config.MemoryEnabled}}
	// Store interaction in memory
	if a.memory != nil {
		// Store the user query
		if err := a.memory.Store(ctx, fmt.Sprintf("%v", inputToProcess), "user-query", "{{.Agent.Name}}"); err != nil {
			logger.Warn().Str("agent", "{{.Agent.Name}}").Err(err).Msg("Failed to store user query in memory")
		}
		
		// Store the agent response
		if err := a.memory.Store(ctx, finalResponse, "agent-response", "{{.Agent.Name}}"); err != nil {
			logger.Warn().Str("agent", "{{.Agent.Name}}").Err(err).Msg("Failed to store agent response in memory")
		}
		
		// Add to chat history
		if err := a.memory.AddMessage(ctx, "user", fmt.Sprintf("%v", inputToProcess)); err != nil {
			logger.Warn().Str("agent", "{{.Agent.Name}}").Err(err).Msg("Failed to add user message to chat history")
		}
		if err := a.memory.AddMessage(ctx, "assistant", finalResponse); err != nil {
			logger.Warn().Str("agent", "{{.Agent.Name}}").Err(err).Msg("Failed to add assistant message to chat history")
		}
		
		logger.Debug().Str("agent", "{{.Agent.Name}}").Msg("Interaction stored in memory")
	}
	{{end}}
	
	{{if .NextAgent}}
	// {{.RoutingComment}}
	outputState.SetMeta(agenticgokit.RouteMetadataKey, "{{.NextAgent}}")
	{{else}}
	// Workflow completion
	{{end}}
	
	// TODO: Add final processing steps here
	// This is the last chance to modify the result before returning.
	// You might want to:
	// - Add final validation
	// - Update metrics or analytics
	// - Trigger notifications or webhooks
	// - Clean up resources
	// - Log completion status
	
	logger.Info().Str("agent", "{{.Agent.Name}}").Msg("Agent processing completed successfully")
	
	// TODO: Customize the AgentResult if needed
	// You can add additional fields or modify the output state
	return agenticgokit.AgentResult{
		OutputState: outputState,
		// TODO: Add custom result fields here if needed
		// Examples:
		// Error: customErrorMessage,
		// Metadata: map[string]interface{}{"processing_time": time.Since(startTime)},
	}, nil
}

// TODO: Add your custom helper functions here
// These functions can be used to organize your agent logic and make it more maintainable.
//
// Example helper functions you might want to implement:

// validateInput validates the input format and content
// TODO: Implement input validation logic for your specific use case
// func (a *{{.Agent.DisplayName}}Handler) validateInput(input interface{}) error {
//     // Add your validation logic here
//     return nil
// }

// preprocessInput transforms the input before sending to LLM
// TODO: Implement input preprocessing for your specific use case
// func (a *{{.Agent.DisplayName}}Handler) preprocessInput(input interface{}) (interface{}, error) {
//     // Add your preprocessing logic here
//     return input, nil
// }

// postprocessOutput transforms the LLM output before returning
// TODO: Implement output postprocessing for your specific use case
// func (a *{{.Agent.DisplayName}}Handler) postprocessOutput(output string) (string, error) {
//     // Add your postprocessing logic here
//     return output, nil
// }

// handleCustomError provides custom error handling for your use case
// TODO: Implement custom error handling logic
// func (a *{{.Agent.DisplayName}}Handler) handleCustomError(err error) error {
//     // Add your error handling logic here
//     return err
// }

// Example of integrating with external services:
// TODO: Uncomment and customize if you need external service integration
// func (a *{{.Agent.DisplayName}}Handler) callExternalAPI(ctx context.Context, data interface{}) (interface{}, error) {
//     // Example API call implementation
//     // client := &http.Client{Timeout: 30 * time.Second}
//     // resp, err := client.Post("https://api.example.com/process", "application/json", bytes.NewBuffer(jsonData))
//     // Handle response and return processed data
//     return nil, nil
// }

// Example of database integration:
// TODO: Uncomment and customize if you need database integration
// func (a *{{.Agent.DisplayName}}Handler) saveToDatabase(ctx context.Context, data interface{}) error {
//     // Example database save implementation
//     // query := "INSERT INTO your_table (data) VALUES ($1)"
//     // _, err := a.database.ExecContext(ctx, query, data)
//     // return err
//     return nil
// }
`
