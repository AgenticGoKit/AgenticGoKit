// Package agents contains the agent implementations for this project.
// 
// This package is where you'll implement your custom agent logic. Each agent
// in this directory represents a specific processing step in your workflow.
//
// TODO: Customize the agents in this package to match your specific use case.
// Each agent can be modified to handle different types of input processing,
// data transformation, or business logic as needed for your application.
package agents

import (
	"context"
	"fmt"
	"strings"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/core"
)

// Agent2Handler represents the agent2 agent handler.
// 
// Purpose: Processes tasks in sequence as part of a processing pipeline
//
// TODO: Modify this agent to implement your specific business logic.
// You can customize the processing logic, add additional fields to the struct,
// or integrate with external services as needed.
//
// Example customizations:
// - Add database connections or API clients as struct fields
// - Implement domain-specific validation logic
// - Add custom error handling for your use case
// - Integrate with external services or APIs
type Agent2Handler struct {
	llm    agenticgokit.ModelProvider
	
	
	// TODO: Add your custom fields here
	// Examples:
	// database    *sql.DB
	// apiClient   *http.Client
	// config      *YourConfig
}

// NewAgent2 creates a new Agent2 instance.
//
// This constructor initializes the agent with the required dependencies.
// 
// TODO: Customize this constructor to accept additional dependencies
// your agent needs, such as database connections, API clients, or
// configuration objects.
//
// Example:
// func NewAgent2(llmProvider agenticgokit.ModelProvider, db *sql.DB, config *YourConfig) *Agent2Handler
func NewAgent2(llmProvider agenticgokit.ModelProvider) *Agent2Handler {
	return &Agent2Handler{
		llm: llmProvider,
		
		
		// TODO: Initialize your custom fields here
		// Examples:
		// database:  db,
		// apiClient: &http.Client{Timeout: 30 * time.Second},
		// config:    config,
	}
}

// Run implements the agenticgokit.AgentHandler interface.
//
// This is the main processing method for the Agent2 agent.
// It receives events and state from the workflow orchestrator and returns
// the processing results.
//
// TODO: Customize this method to implement your specific agent logic.
// The basic structure handles input processing, LLM interaction, and
// result formatting, but you can modify any part of this flow.
//
// Key customization points:
// 1. Input processing: Modify how the agent extracts and validates input
// 2. Business logic: Add your domain-specific processing before/after LLM calls
// 3. LLM prompts: Customize the system and user prompts for your use case
// 4. Output formatting: Change how results are structured and returned
// 5. Error handling: Add custom error handling for your specific scenarios
func (a *Agent2Handler) Run(ctx context.Context, event agenticgokit.Event, state agenticgokit.State) (agenticgokit.AgentResult, error) {
	// Get logger for debug output
	logger := agenticgokit.Logger()
	logger.Debug().Str("agent", "agent2").Str("event_id", event.GetID()).Msg("Agent processing started")
	
	// TODO: Customize input processing logic
	// This section determines what input the agent will process.
	// You can modify this to:
	// - Validate input format and structure
	// - Transform input data before processing
	// - Add input sanitization or filtering
	// - Extract specific fields from complex input objects
	var inputToProcess interface{}
	var systemPrompt string
	
	
	// Sequential processing: Use previous agent's output, with fallback chain
	// TODO: Customize how this agent processes input from previous agents
	// You might want to:
	// - Validate the format of previous agent outputs
	// - Combine outputs from multiple previous agents
	// - Transform the input before processing
	// - Add error handling for malformed previous outputs
	found := false
	agents := []string{"agent1", "agent2"}
	
	for i := 1 - 1; i >= 0; i-- {
		if i < len(agents) {
			prevAgentName := agents[i]
			if agentResponse, exists := state.Get(fmt.Sprintf("%s_response", prevAgentName)); exists {
				inputToProcess = agentResponse
				// TODO: Add validation or transformation of previous agent output
				// Example: inputToProcess = validateAndTransform(agentResponse)
				logger.Debug().Str("agent", "agent2").Str("source_agent", prevAgentName).Interface("input", agentResponse).Msg("Processing previous agent's output")
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
		logger.Debug().Str("agent", "agent2").Interface("input", inputToProcess).Msg("Processing original message (final fallback)")
	}
	
	// TODO: Customize the system prompt for your specific use case
	systemPrompt = `You are Agent2, processes tasks in sequence as part of a processing pipeline.

Core Responsibilities:
- Provide the final, comprehensive response to the user
- Synthesize insights from all previous agents
- Present information in a clear, organized, and authoritative manner
- Use MCP tools only if critical information is still missing
- Ensure the response fully addresses the user's original question

Tool Usage Strategy:
- For stock prices/financial data: Use search tools to find current information
- For current events/news: Use search tools for latest updates
- For specific web content: Use fetch_content tool with URLs
- Always prefer real data over general advice
- Document tool usage and results clearly

Response Quality:
- Provide specific, data-driven answers when possible
- Extract and present key information clearly
- Be conversational but professional
- Integrate tool results naturally into responses

Sequential Mode: You process tasks in sequence. Build upon previous agents' work effectively.`
	
	
	// Get available MCP tools to include in prompt
	var toolsPrompt string
	mcpManager := agenticgokit.GetMCPManager()
	if mcpManager != nil {
		availableTools := mcpManager.GetAvailableTools()
		logger.Debug().Str("agent", "agent2").Int("tool_count", len(availableTools)).Msg("MCP Tools discovered")
		toolsPrompt = agenticgokit.FormatToolsPromptForLLM(availableTools)
	} else {
		logger.Warn().Str("agent", "agent2").Msg("MCP Manager is not available")
	}
	
	
	
	// TODO: Customize prompt construction
	// This is where you can modify how the prompt is built for the LLM.
	// You might want to:
	// - Add domain-specific context or instructions
	// - Format the input in a specific way
	// - Include additional metadata or constraints
	// - Add examples or few-shot learning prompts
	userPrompt := fmt.Sprintf("User query: %v", inputToProcess)
	userPrompt += toolsPrompt
	
	
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
	logger.Debug().Str("agent", "agent2").Str("system_prompt", systemPrompt).Str("user_prompt", userPrompt).Msg("Full LLM prompt")
	
	// TODO: Add pre-processing logic here if needed
	// Example: Validate prompt length, add rate limiting, etc.
	
	// Call LLM to get initial response and potential tool calls
	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		// TODO: Customize error handling for LLM failures
		// You might want to add retry logic, fallback responses, or
		// specific error categorization for your use case
		return agenticgokit.AgentResult{}, fmt.Errorf("Agent2 LLM call failed: %w", err)
	}
	
	logger.Debug().Str("agent", "agent2").Str("response", response.Content).Msg("Initial LLM response received")
	
	// Parse LLM response for tool calls using core function
	toolCalls := agenticgokit.ParseLLMToolCalls(response.Content)
	var mcpResults []string
	
	// Debug: Log the LLM response to see tool call format
	logger.Debug().Str("agent", "agent2").Str("llm_response", response.Content).Msg("LLM response for tool call analysis")
	logger.Debug().Str("agent", "agent2").Interface("parsed_tool_calls", toolCalls).Msg("Parsed tool calls from LLM response")
	
	// Execute any requested tools
	if len(toolCalls) > 0 && mcpManager != nil {
		logger.Info().Str("agent", "agent2").Int("tool_calls", len(toolCalls)).Msg("Executing LLM-requested tools")
		
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
				
				logger.Info().Str("agent", "agent2").Str("tool_name", toolName).Interface("args", args).Msg("Executing tool as requested by LLM")
				
				// Execute tool using the global ExecuteMCPTool function
				result, err := agenticgokit.ExecuteMCPTool(ctx, toolName, args)
				if err != nil {
					logger.Error().Str("agent", "agent2").Str("tool_name", toolName).Err(err).Msg("Tool execution failed")
					mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' failed: %v", toolName, err))
				} else {
					if result.Success {
						logger.Info().Str("agent", "agent2").Str("tool_name", toolName).Msg("Tool execution successful")
						
						// Format the result content
						var resultContent string
						if len(result.Content) > 0 {
							resultContent = result.Content[0].Text
						} else {
							resultContent = "Tool executed successfully but returned no content"
						}
						
						mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' result: %s", toolName, resultContent))
					} else {
						logger.Error().Str("agent", "agent2").Str("tool_name", toolName).Msg("Tool execution was not successful")
						mcpResults = append(mcpResults, fmt.Sprintf("Tool '%s' was not successful", toolName))
					}
				}
			}
		}
	} else {
		logger.Debug().Str("agent", "agent2").Msg("No tool calls requested or MCP manager not available")
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
			return agenticgokit.AgentResult{}, fmt.Errorf("Agent2 final LLM call failed: %w", err)
		}
		finalResponse = finalLLMResponse.Content
		logger.Info().Str("agent", "agent2").Str("final_response", finalResponse).Msg("Final response generated with tool results")
	} else {
		finalResponse = response.Content
		logger.Debug().Str("agent", "agent2").Msg("Using initial LLM response (no tools used)")
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
	outputState.Set("agent2_response", finalResponse)
	outputState.Set("message", finalResponse)
	
	// TODO: Add custom state data here if needed
	// Examples:
	// outputState.Set("agent2_metadata", yourMetadata)
	// outputState.Set("agent2_confidence", confidenceScore)
	// outputState.Set("processing_time", processingTime)
	
	
	
	
	// Route to Responsible AI for final content check
	outputState.SetMeta(agenticgokit.RouteMetadataKey, "responsible_ai")
	
	
	// TODO: Add final processing steps here
	// This is the last chance to modify the result before returning.
	// You might want to:
	// - Add final validation
	// - Update metrics or analytics
	// - Trigger notifications or webhooks
	// - Clean up resources
	// - Log completion status
	
	logger.Info().Str("agent", "agent2").Msg("Agent processing completed successfully")
	
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
// func (a *Agent2Handler) validateInput(input interface{}) error {
//     // Add your validation logic here
//     return nil
// }

// preprocessInput transforms the input before sending to LLM
// TODO: Implement input preprocessing for your specific use case
// func (a *Agent2Handler) preprocessInput(input interface{}) (interface{}, error) {
//     // Add your preprocessing logic here
//     return input, nil
// }

// postprocessOutput transforms the LLM output before returning
// TODO: Implement output postprocessing for your specific use case
// func (a *Agent2Handler) postprocessOutput(output string) (string, error) {
//     // Add your postprocessing logic here
//     return output, nil
// }

// handleCustomError provides custom error handling for your use case
// TODO: Implement custom error handling logic
// func (a *Agent2Handler) handleCustomError(err error) error {
//     // Add your error handling logic here
//     return err
// }

// Example of integrating with external services:
// TODO: Uncomment and customize if you need external service integration
// func (a *Agent2Handler) callExternalAPI(ctx context.Context, data interface{}) (interface{}, error) {
//     // Example API call implementation
//     // client := &http.Client{Timeout: 30 * time.Second}
//     // resp, err := client.Post("https://api.example.com/process", "application/json", bytes.NewBuffer(jsonData))
//     // Handle response and return processed data
//     return nil, nil
// }

// Example of database integration:
// TODO: Uncomment and customize if you need database integration
// func (a *Agent2Handler) saveToDatabase(ctx context.Context, data interface{}) error {
//     // Example database save implementation
//     // query := "INSERT INTO your_table (data) VALUES ($1)"
//     // _, err := a.database.ExecContext(ctx, query, data)
//     // return err
//     return nil
// }
