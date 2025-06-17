// Package core provides the MCP-aware agent implementation.
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// MCPAwareAgent is an intelligent agent that leverages MCP tools for task execution.
// It uses LLM integration to select appropriate tools and execute them in the right sequence.
type MCPAwareAgent struct {
	name        string
	llmProvider LLMProvider
	mcpManager  MCPManager
	config      MCPAgentConfig
	logger      *zerolog.Logger
}

// MCPAgentConfig holds configuration for MCP-aware agents.
type MCPAgentConfig struct {
	// Tool selection settings
	MaxToolsPerExecution int           `toml:"max_tools_per_execution"`
	ToolSelectionTimeout time.Duration `toml:"tool_selection_timeout"`

	// Execution settings
	ParallelExecution bool          `toml:"parallel_execution"`
	ExecutionTimeout  time.Duration `toml:"execution_timeout"`
	RetryFailedTools  bool          `toml:"retry_failed_tools"`
	MaxRetries        int           `toml:"max_retries"`

	// LLM integration settings
	UseToolDescriptions  bool   `toml:"use_tool_descriptions"`
	ToolSelectionPrompt  string `toml:"tool_selection_prompt"`
	ResultInterpretation bool   `toml:"result_interpretation"`
}

// DefaultMCPAgentConfig returns a default configuration for MCP agents.
func DefaultMCPAgentConfig() MCPAgentConfig {
	return MCPAgentConfig{
		MaxToolsPerExecution: 5,
		ToolSelectionTimeout: 30 * time.Second,
		ParallelExecution:    false,
		ExecutionTimeout:     2 * time.Minute,
		RetryFailedTools:     true,
		MaxRetries:           3,
		UseToolDescriptions:  true,
		ToolSelectionPrompt:  DefaultToolSelectionPrompt,
		ResultInterpretation: true,
	}
}

// DefaultToolSelectionPrompt is the default prompt for tool selection.
const DefaultToolSelectionPrompt = `You are an intelligent agent that needs to select the most appropriate tools to accomplish a task.

Available MCP tools:
{{.Tools}}

User request: {{.Query}}
Current context: {{.Context}}

Please analyze the request and context, then respond with a JSON array of tool names that should be executed to accomplish the task. Consider:
1. Which tools are most relevant to the user's request
2. The logical order of tool execution
3. Dependencies between tools
4. The current context and state

Respond with only a JSON array of tool names, for example: ["search", "fetch_content", "summarize"]`

// NewMCPAwareAgent creates a new MCP-aware agent.
func NewMCPAwareAgent(name string, llmProvider LLMProvider, mcpManager MCPManager, config MCPAgentConfig) *MCPAwareAgent {
	logger := GetLogger().With().Str("component", "mcp_agent").Str("name", name).Logger()

	return &MCPAwareAgent{
		name:        name,
		llmProvider: llmProvider,
		mcpManager:  mcpManager,
		config:      config,
		logger:      &logger,
	}
}

// Name returns the agent's name.
func (a *MCPAwareAgent) Name() string {
	return a.name
}

// Run implements the Agent interface. It intelligently selects and executes MCP tools
// based on the input state and LLM guidance.
func (a *MCPAwareAgent) Run(ctx context.Context, inputState State) (State, error) {
	a.logger.Info().Msg("MCPAwareAgent starting execution")

	// Extract the query/task from state
	query, err := a.extractQuery(inputState)
	if err != nil {
		return inputState, fmt.Errorf("failed to extract query from state: %w", err)
	}

	// Select appropriate tools using LLM
	selectedTools, err := a.SelectTools(ctx, query, inputState)
	if err != nil {
		return inputState, fmt.Errorf("failed to select tools: %w", err)
	}

	if len(selectedTools) == 0 {
		a.logger.Info().Msg("No tools selected for execution")
		return inputState, nil
	}

	// Prepare tool executions
	executions, err := a.prepareToolExecutions(ctx, selectedTools, inputState)
	if err != nil {
		return inputState, fmt.Errorf("failed to prepare tool executions: %w", err)
	}

	// Execute tools
	results, err := a.ExecuteTools(ctx, executions)
	if err != nil {
		return inputState, fmt.Errorf("failed to execute tools: %w", err)
	}

	// Update state with results
	outputState := a.updateStateWithResults(inputState, results)

	a.logger.Info().
		Int("tools_executed", len(results)).
		Msg("MCPAwareAgent execution completed")

	return outputState, nil
}

// SelectTools implements the MCPAgent interface. It uses LLM to intelligently
// select the most appropriate tools for the given query and context.
func (a *MCPAwareAgent) SelectTools(ctx context.Context, query string, stateContext State) ([]string, error) {
	// Get available tools
	availableTools := a.mcpManager.GetAvailableTools()
	if len(availableTools) == 0 {
		a.logger.Warn().Msg("No MCP tools available")
		return []string{}, nil
	}

	// Prepare tools description for LLM
	toolsDesc := a.formatToolsForLLM(availableTools)

	// Create prompt for tool selection
	prompt, err := a.createToolSelectionPrompt(toolsDesc, query, stateContext)
	if err != nil {
		return nil, fmt.Errorf("failed to create tool selection prompt: %w", err)
	}

	// Use LLM to select tools
	ctxWithTimeout, cancel := context.WithTimeout(ctx, a.config.ToolSelectionTimeout)
	defer cancel()

	// Create LLM prompt
	llmPrompt := Prompt{
		System: "You are an intelligent agent that selects the most appropriate tools for a given task.",
		User:   prompt,
	}

	response, err := a.llmProvider.Call(ctxWithTimeout, llmPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM tool selection failed: %w", err)
	}

	// Parse LLM response to extract tool names
	selectedTools, err := a.parseToolSelection(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tool selection: %w", err)
	}

	// Validate selected tools exist
	validTools := a.validateSelectedTools(selectedTools, availableTools)

	a.logger.Info().
		Strs("selected_tools", validTools).
		Int("total_available", len(availableTools)).
		Msg("Tools selected for execution")

	return validTools, nil
}

// ExecuteTools implements the MCPAgent interface. It executes the specified tools
// either sequentially or in parallel based on configuration.
func (a *MCPAwareAgent) ExecuteTools(ctx context.Context, tools []MCPToolExecution) ([]MCPToolResult, error) {
	if len(tools) == 0 {
		return []MCPToolResult{}, nil
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, a.config.ExecutionTimeout)
	defer cancel()

	if a.config.ParallelExecution {
		return a.executeToolsParallel(ctxWithTimeout, tools)
	}
	return a.executeToolsSequential(ctxWithTimeout, tools)
}

// GetAvailableMCPTools implements the MCPAgent interface.
func (a *MCPAwareAgent) GetAvailableMCPTools() []MCPToolInfo {
	return a.mcpManager.GetAvailableTools()
}

// executeToolsSequential executes tools one after another.
func (a *MCPAwareAgent) executeToolsSequential(ctx context.Context, tools []MCPToolExecution) ([]MCPToolResult, error) {
	results := make([]MCPToolResult, 0, len(tools))

	for i, tool := range tools {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		a.logger.Debug().
			Str("tool", tool.ToolName).
			Int("step", i+1).
			Int("total", len(tools)).
			Msg("Executing tool")

		result, err := a.executeSingleTool(ctx, tool)
		if err != nil {
			if a.config.RetryFailedTools {
				result, err = a.retryToolExecution(ctx, tool, err)
			}
			if err != nil {
				// Return partial results with error
				result = MCPToolResult{
					ToolName: tool.ToolName,
					Success:  false,
					Error:    err.Error(),
				}
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// executeToolsParallel executes tools in parallel (future enhancement).
func (a *MCPAwareAgent) executeToolsParallel(ctx context.Context, tools []MCPToolExecution) ([]MCPToolResult, error) {
	// For now, fall back to sequential execution
	// This can be enhanced with goroutines and proper synchronization
	a.logger.Debug().Msg("Parallel execution not fully implemented, falling back to sequential")
	return a.executeToolsSequential(ctx, tools)
}

// executeSingleTool executes a single MCP tool.
func (a *MCPAwareAgent) executeSingleTool(ctx context.Context, tool MCPToolExecution) (MCPToolResult, error) {
	// This would call the actual MCP tool execution
	// For now, we'll create a placeholder implementation
	// The actual implementation would involve calling the MCP manager's tool execution

	a.logger.Debug().
		Str("tool", tool.ToolName).
		Interface("args", tool.Arguments).
		Msg("Executing MCP tool")
	// Simulate tool execution (replace with actual MCP tool call)
	result := MCPToolResult{
		ToolName: tool.ToolName,
		Success:  true,
		Content: []MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Result from %s with args %v", tool.ToolName, tool.Arguments),
			},
		},
	}

	return result, nil
}

// retryToolExecution retries a failed tool execution.
func (a *MCPAwareAgent) retryToolExecution(ctx context.Context, tool MCPToolExecution, originalErr error) (MCPToolResult, error) {
	a.logger.Warn().
		Str("tool", tool.ToolName).
		Err(originalErr).
		Msg("Tool execution failed, retrying")

	for attempt := 1; attempt <= a.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return MCPToolResult{}, ctx.Err()
		default:
		}

		a.logger.Debug().
			Str("tool", tool.ToolName).
			Int("attempt", attempt).
			Msg("Retrying tool execution")

		result, err := a.executeSingleTool(ctx, tool)
		if err == nil {
			a.logger.Info().
				Str("tool", tool.ToolName).
				Int("attempt", attempt).
				Msg("Tool execution succeeded on retry")
			return result, nil
		}

		// Wait before next retry
		if attempt < a.config.MaxRetries {
			select {
			case <-ctx.Done():
				return MCPToolResult{}, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
	}

	return MCPToolResult{}, fmt.Errorf("tool execution failed after %d retries: %w", a.config.MaxRetries, originalErr)
}

// Helper methods

// extractQuery extracts the main query/task from the input state.
func (a *MCPAwareAgent) extractQuery(state State) (string, error) {
	// Try different common keys for the query
	queryKeys := []string{"query", "task", "request", "message", "prompt", "input"}

	for _, key := range queryKeys {
		if value, exists := state.Get(key); exists {
			if str, ok := value.(string); ok && str != "" {
				return str, nil
			}
		}
	}

	// If no specific query found, try to find any string values
	keys := state.Keys()
	if len(keys) == 0 {
		return "", fmt.Errorf("no query or task found in state")
	}

	// Try to find any string values
	for _, key := range keys {
		if value, exists := state.Get(key); exists {
			if str, ok := value.(string); ok && str != "" {
				return str, nil
			}
		}
	}

	return fmt.Sprintf("Process state with %d fields", len(keys)), nil
}

// formatToolsForLLM creates a formatted description of available tools for the LLM.
func (a *MCPAwareAgent) formatToolsForLLM(tools []MCPToolInfo) string {
	var sb strings.Builder

	for i, tool := range tools {
		if i > 0 {
			sb.WriteString("\n")
		}

		sb.WriteString(fmt.Sprintf("- %s: %s", tool.Name, tool.Description))
		if tool.ServerName != "" {
			sb.WriteString(fmt.Sprintf(" (from %s)", tool.ServerName))
		}
	}

	return sb.String()
}

// createToolSelectionPrompt creates the prompt for LLM tool selection.
func (a *MCPAwareAgent) createToolSelectionPrompt(toolsDesc, query string, stateContext State) (string, error) {
	prompt := a.config.ToolSelectionPrompt

	// Replace placeholders
	prompt = strings.ReplaceAll(prompt, "{{.Tools}}", toolsDesc)
	prompt = strings.ReplaceAll(prompt, "{{.Query}}", query)

	// Format context
	contextStr := ""
	keys := stateContext.Keys()
	if len(keys) > 0 {
		contextMap := make(map[string]interface{})
		for _, key := range keys {
			if value, exists := stateContext.Get(key); exists {
				contextMap[key] = value
			}
		}
		contextData, _ := json.Marshal(contextMap)
		contextStr = string(contextData)
	}
	prompt = strings.ReplaceAll(prompt, "{{.Context}}", contextStr)

	return prompt, nil
}

// parseToolSelection parses the LLM response to extract selected tool names.
func (a *MCPAwareAgent) parseToolSelection(response string) ([]string, error) {
	// Clean up the response
	response = strings.TrimSpace(response)

	// Try to find JSON array in the response
	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")

	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no valid JSON array found in response: %s", response)
	}

	jsonStr := response[start : end+1]

	var tools []string
	if err := json.Unmarshal([]byte(jsonStr), &tools); err != nil {
		return nil, fmt.Errorf("failed to parse tool selection JSON: %w", err)
	}

	return tools, nil
}

// validateSelectedTools ensures that selected tools actually exist.
func (a *MCPAwareAgent) validateSelectedTools(selected []string, available []MCPToolInfo) []string {
	if len(selected) == 0 {
		return []string{}
	}

	// Create a map for quick lookup
	availableMap := make(map[string]bool)
	for _, tool := range available {
		availableMap[tool.Name] = true
	}

	// Filter valid tools
	validTools := make([]string, 0, len(selected))
	for _, toolName := range selected {
		if availableMap[toolName] {
			validTools = append(validTools, toolName)
		} else {
			a.logger.Warn().
				Str("tool", toolName).
				Msg("Selected tool not available, skipping")
		}
	}

	// Limit to max tools per execution
	if len(validTools) > a.config.MaxToolsPerExecution {
		a.logger.Info().
			Int("selected", len(validTools)).
			Int("max_allowed", a.config.MaxToolsPerExecution).
			Msg("Limiting number of tools to execute")
		validTools = validTools[:a.config.MaxToolsPerExecution]
	}

	return validTools
}

// prepareToolExecutions prepares tool execution requests from tool names and state.
func (a *MCPAwareAgent) prepareToolExecutions(ctx context.Context, toolNames []string, state State) ([]MCPToolExecution, error) {
	executions := make([]MCPToolExecution, 0, len(toolNames))

	for _, toolName := range toolNames {
		// Get tool info to understand its schema
		toolInfo := a.findToolInfo(toolName)
		if toolInfo == nil {
			a.logger.Warn().
				Str("tool", toolName).
				Msg("Tool info not found, creating basic execution")
		}

		// Prepare arguments based on state and tool schema
		args := a.prepareToolArguments(toolName, toolInfo, state)

		execution := MCPToolExecution{
			ToolName:  toolName,
			Arguments: args,
		}

		executions = append(executions, execution)
	}

	return executions, nil
}

// findToolInfo finds information about a specific tool.
func (a *MCPAwareAgent) findToolInfo(toolName string) *MCPToolInfo {
	tools := a.mcpManager.GetAvailableTools()
	for _, tool := range tools {
		if tool.Name == toolName {
			return &tool
		}
	}
	return nil
}

// prepareToolArguments prepares arguments for a tool based on state and schema.
func (a *MCPAwareAgent) prepareToolArguments(toolName string, toolInfo *MCPToolInfo, state State) map[string]interface{} {
	args := make(map[string]interface{})

	// Basic argument mapping based on common patterns
	keys := state.Keys()
	stateMap := make(map[string]interface{})
	for _, key := range keys {
		if value, exists := state.Get(key); exists {
			stateMap[key] = value
		}
	}

	// Common argument mappings
	if query, exists := stateMap["query"]; exists {
		args["query"] = query
	}
	if text, exists := stateMap["text"]; exists {
		args["text"] = text
	}
	if url, exists := stateMap["url"]; exists {
		args["url"] = url
	}

	// Tool-specific argument mapping
	switch toolName {
	case "search":
		if q, exists := stateMap["query"]; exists {
			args["q"] = q
		}
	case "fetch_content":
		if url, exists := stateMap["url"]; exists {
			args["url"] = url
		}
	}

	return args
}

// updateStateWithResults updates the state with tool execution results.
func (a *MCPAwareAgent) updateStateWithResults(inputState State, results []MCPToolResult) State {
	// Create a copy of the input state
	outputState := NewState()

	// Copy all existing data
	keys := inputState.Keys()
	for _, key := range keys {
		if value, exists := inputState.Get(key); exists {
			outputState.Set(key, value)
		}
	}

	// Add results to state
	resultData := make([]map[string]interface{}, 0, len(results))
	for _, result := range results {
		resultMap := map[string]interface{}{
			"tool_name": result.ToolName,
			"success":   result.Success,
			"content":   result.Content,
			"duration":  result.Duration,
		}
		if result.Error != "" {
			resultMap["error"] = result.Error
		}
		resultData = append(resultData, resultMap)
	}

	outputState.Set("mcp_results", resultData)

	// Extract and set specific content from results
	for _, result := range results {
		if result.Success && len(result.Content) > 0 {
			// Get the first text content
			for _, content := range result.Content {
				if content.Type == "text" && content.Text != "" {
					key := fmt.Sprintf("mcp_%s_result", result.ToolName)
					outputState.Set(key, content.Text)
					break
				}
			}
		}
	}

	return outputState
}
