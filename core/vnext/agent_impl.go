package vnext

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/internal/llm"
)

// realAgent is the concrete implementation of the Agent interface.
// It integrates with real LLM providers, memory systems, and tools to provide
// full agent functionality. This replaces the mock streamlinedAgent implementation.
//
// The realAgent follows the same pattern as core.SimpleAgent's agentImpl,
// keeping implementation alongside interfaces in the same package.
type realAgent struct {
	// Configuration
	config  *Config
	handler HandlerFunc

	// Core dependencies - directly using internal implementations
	llmProvider    llm.ModelProvider // LLM provider (Ollama, OpenAI, Azure, etc.)
	memoryProvider core.Memory       // Memory provider (optional, for context/RAG)
	tools          []Tool            // Tools available to the agent (optional)

	// Runtime state
	initialized bool
	sessions    map[string]*sessionState
	metrics     *agentMetrics
}

// sessionState tracks per-session information for the agent
type sessionState struct {
	id        string
	startTime time.Time
	messages  []sessionMessage
	metadata  map[string]interface{}
}

// sessionMessage represents a single message in a session
type sessionMessage struct {
	role      string
	content   string
	timestamp time.Time
}

// agentMetrics tracks runtime metrics for the agent
type agentMetrics struct {
	totalRuns       int64
	totalErrors     int64
	totalDuration   time.Duration
	averageDuration time.Duration
	lastRunTime     time.Time
}

// newRealAgent creates a new agent instance from configuration.
// This is the constructor called by builder.Build() to create a real,
// functional agent that makes actual LLM calls.
//
// Parameters:
//   - config: Agent configuration including LLM, memory, and tool settings
//   - handler: Optional custom handler function for advanced use cases
//
// Returns:
//   - Agent implementation or error if initialization fails
func newRealAgent(config *Config, handler HandlerFunc) (Agent, error) {
	// Validate configuration
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if config.Name == "" {
		return nil, fmt.Errorf("agent name cannot be empty")
	}
	if config.LLM.Provider == "" {
		return nil, fmt.Errorf("LLM provider must be specified")
	}
	if config.LLM.Model == "" {
		return nil, fmt.Errorf("LLM model must be specified")
	}

	// Initialize agent struct
	agent := &realAgent{
		config:      config,
		handler:     handler,
		initialized: false,
		sessions:    make(map[string]*sessionState),
		metrics: &agentMetrics{
			totalRuns:     0,
			totalErrors:   0,
			totalDuration: 0,
		},
	}

	// Initialize LLM provider from config
	llmProvider, err := createLLMProvider(config.LLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}
	agent.llmProvider = llmProvider

	// Initialize memory provider if configured
	if config.Memory != nil {
		memoryProvider, err := createMemoryProvider(config.Memory)
		if err != nil {
			return nil, fmt.Errorf("failed to create memory provider: %w", err)
		}
		agent.memoryProvider = memoryProvider
	}

	// Initialize tools if configured
	if config.Tools != nil && config.Tools.Enabled {
		tools, err := createTools(config.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to create tools: %w", err)
		}
		agent.tools = tools
	}

	// Mark as initialized
	agent.initialized = true

	return agent, nil
}

// Run executes the agent with the given input and returns the result.
// This method makes actual LLM API calls and integrates with memory and tools.
//
// Implementation follows this flow:
//  1. Build prompt (system + user input + memory context if enabled)
//  2. Call LLM provider
//  3. Parse response and check for tool calls
//  4. Execute tools if needed
//  5. Store interaction in memory if enabled
//  6. Return result with content, timing, and metadata
func (a *realAgent) Run(ctx context.Context, input string) (*Result, error) {
	startTime := time.Now()

	// Validate that agent is properly initialized
	if a.llmProvider == nil {
		return nil, fmt.Errorf("agent not properly initialized: LLM provider is nil")
	}

	// Step 1: Build the prompt with system context and user input
	prompt := llm.Prompt{
		System: a.config.SystemPrompt,
		User:   input,
	}

	// Step 1.5: Add tool descriptions to system prompt if tools are available
	if len(a.tools) > 0 {
		toolDescriptions := FormatToolsForPrompt(a.tools)
		prompt.System = prompt.System + toolDescriptions

		core.Logger().Debug().
			Int("tool_count", len(a.tools)).
			Msg("Added tool descriptions to system prompt")
	}

	// Add model parameters from config if specified
	if a.config.LLM.Temperature > 0 {
		temp := float32(a.config.LLM.Temperature)
		prompt.Parameters.Temperature = &temp
	}
	if a.config.LLM.MaxTokens > 0 {
		maxTok := int32(a.config.LLM.MaxTokens)
		prompt.Parameters.MaxTokens = &maxTok
	}

	// Step 2: Enhance prompt with memory context if memory is enabled
	// Use the new BuildEnrichedPrompt utility for proper RAG integration
	if a.memoryProvider != nil && a.config.Memory != nil {
		// Convert llm.Prompt to core.Prompt for enrichment
		corePrompt := BuildEnrichedPrompt(ctx, prompt.System, prompt.User, a.memoryProvider, a.config.Memory)

		// Update the LLM prompt with enriched content
		prompt.System = corePrompt.System
		prompt.User = corePrompt.User

		core.Logger().Debug().
			Str("original_input", input).
			Int("enriched_length", len(prompt.User)).
			Msg("Enhanced prompt with memory context")
	}

	// Step 3: Call the LLM provider
	response, err := a.llmProvider.Call(ctx, prompt)
	if err != nil {
		// Update metrics
		a.updateMetrics(startTime, true)
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Step 3.5: Execute tool calls if any are detected in the response
	// This implements an agentic loop where the LLM can use tools and continue reasoning
	var toolCalls []ToolCall
	finalResponse := response.Content
	if len(a.tools) > 0 {
		const maxToolIterations = 5 // Prevent infinite tool-calling loops
		var toolErr error
		finalResponse, toolCalls, toolErr = a.executeToolsAndContinue(ctx, response.Content, prompt, maxToolIterations)
		if toolErr != nil {
			// Log warning but don't fail - use the last valid response
			core.Logger().Warn().Err(toolErr).Msg("Tool execution encountered error, using last valid response")
		}
	}

	// Step 4: Store the interaction in memory if enabled
	if a.memoryProvider != nil {
		if err := a.storeInMemory(ctx, input, finalResponse); err != nil {
			// Log warning but don't fail - memory storage is non-critical
			core.Logger().Warn().Err(err).Msg("Failed to store interaction in memory")
		}
	}

	// Step 5: Call custom handler if configured
	// Handler can modify or replace the response
	if a.handler != nil {
		capabilities := &Capabilities{
			LLM: func(system, user string) (string, error) {
				prompt := llm.Prompt{System: system, User: user}
				resp, err := a.llmProvider.Call(ctx, prompt)
				if err != nil {
					return "", err
				}
				return resp.Content, nil
			},
			// Tools and Memory would be provided here if available
			// Tools:  a.tools,
			// Memory: a.memoryProvider,
		}

		handlerResponse, err := a.handler(ctx, finalResponse, capabilities)
		if err != nil {
			core.Logger().Warn().Err(err).Msg("Custom handler returned error")
		} else if handlerResponse != "" {
			// Handler can override the response
			finalResponse = handlerResponse
		}
	}

	// Step 6: Update metrics
	a.updateMetrics(startTime, false)

	// Step 7: Build and return the result
	duration := time.Since(startTime)
	result := &Result{
		Success:    true,
		Content:    finalResponse,
		Duration:   duration,
		TokensUsed: response.Usage.TotalTokens,
		MemoryUsed: a.memoryProvider != nil,
		ToolCalls:  toolCalls, // Include tool execution details
		StartTime:  startTime,
		EndTime:    time.Now(),
		Metadata: map[string]interface{}{
			"model":         a.config.LLM.Model,
			"provider":      a.config.LLM.Provider,
			"finish_reason": response.FinishReason,
		},
	}

	// Add tool names to ToolsCalled list for convenience
	if len(toolCalls) > 0 {
		var toolNames []string
		for _, tc := range toolCalls {
			toolNames = append(toolNames, tc.Name)
		}
		result.ToolsCalled = toolNames
	}

	// Add LLM interaction details
	result.LLMInteractions = []LLMInteraction{
		{
			Provider:       a.config.LLM.Provider,
			Model:          a.config.LLM.Model,
			PromptTokens:   response.Usage.PromptTokens,
			ResponseTokens: response.Usage.CompletionTokens,
			Duration:       duration,
			Success:        true,
		},
	}

	return result, nil
}

// buildMemoryContext retrieves relevant context from memory for the given input.
// Deprecated: This method is kept for backward compatibility. Use BuildEnrichedPrompt instead.
func (a *realAgent) buildMemoryContext(ctx context.Context, input string) (string, error) {
	if a.memoryProvider == nil || a.config.Memory == nil {
		return "", nil
	}

	// Use the new utility function for enrichment
	enrichedInput := EnrichWithMemory(ctx, a.memoryProvider, input, a.config.Memory)

	// Return only the added context (not the original input)
	if enrichedInput == input {
		return "", nil // No context was added
	}

	return enrichedInput, nil
}

// storeInMemory stores the current interaction in memory for future reference.
// This includes both the personal memory (for RAG) and chat history (for context).
func (a *realAgent) storeInMemory(ctx context.Context, input, output string) error {
	if a.memoryProvider == nil {
		return nil
	}

	// Store as personal memory for RAG retrieval
	// This allows the agent to recall facts and context from past conversations
	if err := a.memoryProvider.Store(ctx, input, "user_message", "conversation"); err != nil {
		core.Logger().Warn().Err(err).Msg("Failed to store user message in memory")
		// Don't return error - continue with chat history storage
	}

	if err := a.memoryProvider.Store(ctx, output, "agent_response", "conversation"); err != nil {
		core.Logger().Warn().Err(err).Msg("Failed to store agent response in memory")
		// Don't return error - continue with chat history storage
	}

	// Store as chat messages for conversation history
	// This enables the agent to maintain context across multiple turns
	if err := a.memoryProvider.AddMessage(ctx, "user", input); err != nil {
		return fmt.Errorf("failed to add user message to chat history: %w", err)
	}

	if err := a.memoryProvider.AddMessage(ctx, "assistant", output); err != nil {
		return fmt.Errorf("failed to add assistant message to chat history: %w", err)
	}

	core.Logger().Debug().
		Int("input_length", len(input)).
		Int("output_length", len(output)).
		Msg("Stored interaction in memory and chat history")

	return nil
}

// updateMetrics updates the agent's runtime metrics after an execution.
func (a *realAgent) updateMetrics(startTime time.Time, hadError bool) {
	if a.metrics == nil {
		return
	}

	duration := time.Since(startTime)
	a.metrics.totalRuns++
	if hadError {
		a.metrics.totalErrors++
	}
	a.metrics.totalDuration += duration
	a.metrics.averageDuration = time.Duration(int64(a.metrics.totalDuration) / a.metrics.totalRuns)
	a.metrics.lastRunTime = startTime
}

// RunWithOptions executes the agent with additional options.
// Options allow fine-grained control over execution behavior including
// timeouts, memory settings, tool configuration, and result detail level.
// RunWithOptions executes the agent with custom runtime options.
//
// This method allows fine-grained control over execution parameters without
// modifying the agent's base configuration. Options can override:
//   - Timeout: Execution deadline via context
//   - Memory: Session ID, memory provider settings
//   - Tools: Tool selection and mode
//   - LLM: Temperature, max tokens
//   - Result: Detailed metadata, trace data, source attributions
//
// The method applies options by:
//  1. Creating a derived context with timeout if specified
//  2. Temporarily overriding agent configuration fields
//  3. Calling Run() with the modified configuration
//  4. Restoring original configuration
//  5. Enhancing result with additional metadata if DetailedResult is true
//
// Example:
//
//	opts := &RunOptions{
//	    Timeout: 30 * time.Second,
//	    SessionID: "user-session-123",
//	    DetailedResult: true,
//	    Temperature: &temperature,
//	}
//	result, err := agent.RunWithOptions(ctx, "analyze this data", opts)
func (a *realAgent) RunWithOptions(ctx context.Context, input string, opts *RunOptions) (*Result, error) {
	if opts == nil {
		// No options provided, delegate to standard Run()
		return a.Run(ctx, input)
	}

	// Step 1: Apply timeout to context if specified
	runCtx := ctx
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Step 2: Save original configuration to restore later
	originalTools := a.tools
	originalTemperature := a.config.LLM.Temperature
	originalMaxTokens := a.config.LLM.MaxTokens

	// Restore configuration after execution
	defer func() {
		a.tools = originalTools
		a.config.LLM.Temperature = originalTemperature
		a.config.LLM.MaxTokens = originalMaxTokens
	}()

	// Step 3: Apply tool options
	if opts.ToolMode == "none" {
		// Disable all tools for this run
		a.tools = nil
	} else if len(opts.Tools) > 0 && opts.ToolMode == "specific" {
		// Filter to only specified tools
		filteredTools := []Tool{}
		for _, toolName := range opts.Tools {
			for _, tool := range a.tools {
				if tool.Name() == toolName {
					filteredTools = append(filteredTools, tool)
					break
				}
			}
		}
		a.tools = filteredTools
	}
	// If ToolMode is "auto" or unspecified, use all configured tools (no change)

	// Step 4: Apply memory options
	if opts.Memory != nil {
		// Override memory configuration for this run
		if !opts.Memory.Enabled && a.memoryProvider != nil {
			// Temporarily disable memory by setting provider to nil
			originalMemory := a.memoryProvider
			a.memoryProvider = nil
			defer func() { a.memoryProvider = originalMemory }()
		}
		// Note: Changing memory provider at runtime is complex and not supported here.
		// The Memory.Provider field would require recreating the provider.
	}

	// Apply session ID if specified
	if opts.SessionID != "" {
		// Store session ID in context for memory operations
		runCtx = context.WithValue(runCtx, "session_id", opts.SessionID)
	}

	// Step 5: Apply LLM parameter overrides
	if opts.Temperature != nil {
		a.config.LLM.Temperature = float32(*opts.Temperature)
	}
	if opts.MaxTokens > 0 {
		a.config.LLM.MaxTokens = opts.MaxTokens
	}

	// Step 6: Execute the run with applied options
	result, err := a.Run(runCtx, input)
	if err != nil {
		return nil, err
	}

	// Step 7: Enhance result if detailed information is requested
	if opts.DetailedResult {
		// Add session information
		if opts.SessionID != "" {
			result.SessionID = opts.SessionID
		}

		// Add tool execution details if not already present
		if len(result.ToolCalls) > 0 && len(result.ToolExecutions) == 0 {
			for _, tc := range result.ToolCalls {
				result.ToolExecutions = append(result.ToolExecutions, ToolExecution{
					Name:      tc.Name,
					Duration:  tc.Duration,
					Success:   tc.Success,
					Error:     tc.Error,
					InputSize: len(fmt.Sprintf("%v", tc.Arguments)),
				})
			}
		}

		// Add configuration metadata
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["timeout"] = opts.Timeout.String()
		result.Metadata["tool_mode"] = opts.ToolMode
		result.Metadata["max_retries"] = opts.MaxRetries
		if opts.Temperature != nil {
			result.Metadata["temperature_override"] = *opts.Temperature
		}
		if opts.MaxTokens > 0 {
			result.Metadata["max_tokens_override"] = opts.MaxTokens
		}
	}

	// Step 8: Add trace data if requested
	if opts.IncludeTrace && result.TraceID != "" {
		// Trace data would be fetched from tracing system
		// For now, just flag that trace is available
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["trace_available"] = true
		result.Metadata["trace_id"] = result.TraceID
	}

	return result, nil
}

// RunStream executes the agent with streaming output.
// Tokens are streamed in real-time as they are generated by the LLM.
//
// The returned Stream can be consumed via:
//
//	for chunk := range stream.Chunks() {
//	    fmt.Print(chunk.Delta)
//	}
func (a *realAgent) RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error) {
	// TODO: Implementation in Task 2.5
	return nil, fmt.Errorf("not implemented yet")
}

// RunStreamWithOptions executes the agent with streaming output and additional options.
// Combines the flexibility of RunWithOptions with real-time streaming.
func (a *realAgent) RunStreamWithOptions(ctx context.Context, input string, runOpts *RunOptions, streamOpts ...StreamOption) (Stream, error) {
	// TODO: Implementation in Task 2.6
	return nil, fmt.Errorf("not implemented yet")
}

// Name returns the agent's name from configuration.
func (a *realAgent) Name() string {
	if a.config == nil {
		return ""
	}
	return a.config.Name
}

// Config returns the agent's configuration.
// This allows inspection of the agent's settings at runtime.
func (a *realAgent) Config() *Config {
	return a.config
}

// Capabilities returns a list of agent capabilities based on configuration.
// Capabilities may include: "llm", "memory", "rag", "tools", "streaming", "custom_handler"
func (a *realAgent) Capabilities() []string {
	capabilities := []string{}

	// LLM capability - always present if we have an LLM provider
	if a.llmProvider != nil {
		capabilities = append(capabilities, "llm")
	}

	// Memory capability
	if a.memoryProvider != nil {
		capabilities = append(capabilities, "memory")

		// RAG capability if memory has RAG enabled
		if a.config != nil && a.config.Memory != nil && a.config.Memory.RAG != nil {
			capabilities = append(capabilities, "rag")
		}
	}

	// Tools capability
	if len(a.tools) > 0 {
		capabilities = append(capabilities, "tools")
	}

	// Streaming capability (if LLM provider supports it)
	if a.llmProvider != nil {
		capabilities = append(capabilities, "streaming")
	}

	// Custom handler capability
	if a.handler != nil {
		capabilities = append(capabilities, "custom_handler")
	}

	return capabilities
}

// Initialize initializes the agent and its dependencies.
// This is called automatically by the builder but can be called manually
// if needed for lazy initialization patterns.
func (a *realAgent) Initialize(ctx context.Context) error {
	// Check if already initialized
	if a.initialized {
		return nil
	}

	// Validate required components
	if a.config == nil {
		return fmt.Errorf("agent configuration is nil")
	}
	if a.llmProvider == nil {
		return fmt.Errorf("LLM provider is nil")
	}

	// Initialize memory provider if present
	if a.memoryProvider != nil {
		// Memory providers don't have an explicit Initialize method
		// They're initialized when created via core.NewMemory()
		core.Logger().Debug().Msg("Memory provider initialized")
	}

	// Initialize sessions map if not already done
	if a.sessions == nil {
		a.sessions = make(map[string]*sessionState)
	}

	// Initialize metrics if not already done
	if a.metrics == nil {
		a.metrics = &agentMetrics{}
	}

	// Mark as initialized
	a.initialized = true
	core.Logger().Info().
		Str("agent", a.config.Name).
		Str("provider", a.config.LLM.Provider).
		Str("model", a.config.LLM.Model).
		Msg("Agent initialized successfully")

	return nil
}

// Cleanup releases agent resources and closes connections.
// Should be called when the agent is no longer needed, typically via defer:
//
//	agent, _ := builder.Build()
//	defer agent.Cleanup(context.Background())
func (a *realAgent) Cleanup(ctx context.Context) error {
	core.Logger().Debug().Str("agent", a.config.Name).Msg("Cleaning up agent resources")

	// Close memory provider if present
	if a.memoryProvider != nil {
		if err := a.memoryProvider.Close(); err != nil {
			core.Logger().Warn().Err(err).Msg("Error closing memory provider")
			// Don't return error, continue cleanup
		}
	}

	// Clear sessions
	a.sessions = nil

	// Mark as not initialized
	a.initialized = false

	core.Logger().Info().Str("agent", a.config.Name).Msg("Agent cleanup completed")
	return nil
}

// =============================================================================
// TOOL EXECUTION HELPERS
// =============================================================================

// executeTool looks up a tool by name and executes it with the given arguments.
// Returns the tool result, including success status and any errors.
func (a *realAgent) executeTool(ctx context.Context, toolCall ToolCall) ToolCall {
	startTime := time.Now()

	// Initialize result fields
	toolCall.Success = false
	toolCall.Duration = 0

	// Find the tool by name
	var tool Tool
	for _, t := range a.tools {
		if t.Name() == toolCall.Name {
			tool = t
			break
		}
	}

	if tool == nil {
		toolCall.Error = fmt.Sprintf("tool '%s' not found", toolCall.Name)
		toolCall.Duration = time.Since(startTime)
		return toolCall
	}

	// Execute the tool
	result, err := tool.Execute(ctx, toolCall.Arguments)
	toolCall.Duration = time.Since(startTime)

	if err != nil {
		toolCall.Error = err.Error()
		toolCall.Success = false
		if result != nil {
			toolCall.Result = result.Content
		}
		return toolCall
	}

	if result == nil {
		toolCall.Error = "tool returned nil result"
		toolCall.Success = false
		return toolCall
	}

	// Tool executed successfully
	toolCall.Success = result.Success
	toolCall.Result = result.Content
	if !result.Success {
		toolCall.Error = result.Error
	}

	return toolCall
}

// executeToolsAndContinue executes any tool calls found in the LLM response,
// then calls the LLM again with the tool results. Returns the final response
// after all tool executions or when no more tool calls are detected.
//
// This implements an agentic loop where the LLM can request tool usage,
// get results, and continue reasoning with those results.
func (a *realAgent) executeToolsAndContinue(
	ctx context.Context,
	initialResponse string,
	originalPrompt llm.Prompt,
	maxIterations int,
) (string, []ToolCall, error) {
	if len(a.tools) == 0 {
		// No tools available, return response as-is
		return initialResponse, nil, nil
	}

	var allToolCalls []ToolCall
	currentResponse := initialResponse
	iteration := 0

	for iteration < maxIterations {
		// Parse tool calls from the current response
		toolCalls := ParseToolCalls(currentResponse)

		if len(toolCalls) == 0 {
			// No tool calls found, we're done
			break
		}

		core.Logger().Debug().
			Int("iteration", iteration+1).
			Int("tool_calls", len(toolCalls)).
			Msg("Executing tool calls")

		// Execute all tool calls
		var executedCalls []ToolCall
		var toolResults strings.Builder

		for _, call := range toolCalls {
			executedCall := a.executeTool(ctx, call)
			executedCalls = append(executedCalls, executedCall)
			allToolCalls = append(allToolCalls, executedCall)

			// Format tool result for next LLM call
			toolResults.WriteString(FormatToolResult(executedCall.Name, &ToolResult{
				Success: executedCall.Success,
				Content: executedCall.Result,
				Error:   executedCall.Error,
			}))
		}

		// Build a new prompt with tool results
		continuationPrompt := llm.Prompt{
			System: originalPrompt.System,
			User: fmt.Sprintf(
				"Previous response:\n%s\n\nTool execution results:\n%s\n\nPlease continue with your response based on the tool results.",
				currentResponse,
				toolResults.String(),
			),
			Parameters: originalPrompt.Parameters,
		}

		// Call LLM again with tool results
		response, err := a.llmProvider.Call(ctx, continuationPrompt)
		if err != nil {
			// Return what we have so far with the error
			return currentResponse, allToolCalls, fmt.Errorf("LLM call failed after tool execution: %w", err)
		}

		currentResponse = response.Content
		iteration++

		// Check if we've reached max iterations
		if iteration >= maxIterations {
			core.Logger().Warn().
				Int("max_iterations", maxIterations).
				Msg("Reached maximum tool execution iterations")
			break
		}
	}

	return currentResponse, allToolCalls, nil
}
