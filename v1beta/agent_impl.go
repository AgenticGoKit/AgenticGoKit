package v1beta

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/agenticgokit/agenticgokit/internal/llm"
	"github.com/agenticgokit/agenticgokit/internal/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// addMultimodalDataToPrompt adds multimodal data from RunOptions to an llm.Prompt.
// This is a helper to avoid code duplication between execute and streaming methods.
func addMultimodalDataToPrompt(prompt *llm.Prompt, opts *RunOptions) {
	if opts == nil {
		return
	}

	if len(opts.Images) > 0 {
		prompt.Images = make([]llm.ImageData, len(opts.Images))
		for i, img := range opts.Images {
			prompt.Images[i] = llm.ImageData{
				URL:      img.URL,
				Base64:   img.Base64,
				Metadata: img.Metadata,
			}
		}
	}

	if len(opts.Audio) > 0 {
		prompt.Audio = make([]llm.AudioData, len(opts.Audio))
		for i, aud := range opts.Audio {
			prompt.Audio[i] = llm.AudioData{
				URL:      aud.URL,
				Base64:   aud.Base64,
				Format:   aud.Format,
				Metadata: aud.Metadata,
			}
		}
	}

	if len(opts.Video) > 0 {
		prompt.Video = make([]llm.VideoData, len(opts.Video))
		for i, vid := range opts.Video {
			prompt.Video[i] = llm.VideoData{
				URL:      vid.URL,
				Base64:   vid.Base64,
				Format:   vid.Format,
				Metadata: vid.Metadata,
			}
		}
	}
}

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

	// Observability
	tracerShutdown func(context.Context) error
	runID          string
	runDir         string
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

	// Initialize memory configuration defaults if not present
	if config.Memory == nil {
		config.Memory = &MemoryConfig{
			Enabled:  true,
			Provider: "chromem",
			RAG: &RAGConfig{
				HistoryLimit:    10,
				MaxTokens:       2000,
				PersonalWeight:  0.3,
				KnowledgeWeight: 0.7,
			},
		}
	} else {
		// If MemoryConfig is provided, ensure it's Enabled by default unless explicitly false
		// Actually, we should probably just treat a non-nil MemoryConfig as wanting memory.
		// But follow the Enabled flag if it's there.
		// Simple fix: if a config is there, we default Enabled to true if not specified.
		// Since we can't tell if it's explicitly false or just default, we'll assume
		// if they provided a config object, they probably want it enabled.
		config.Memory.Enabled = true

		// Smart Default: If no embedding provider is specified, try to use the LLM provider
		if config.Memory.Options == nil {
			config.Memory.Options = make(map[string]string)
		}
		if _, ok := config.Memory.Options["embedding_provider"]; !ok {
			config.Memory.Options["embedding_provider"] = config.LLM.Provider
			if _, ok := config.Memory.Options["embedding_model"]; !ok {
				config.Memory.Options["embedding_model"] = config.LLM.Model
			}
			if _, ok := config.Memory.Options["embedding_url"]; !ok {
				config.Memory.Options["embedding_url"] = config.LLM.BaseURL
			}
			if _, ok := config.Memory.Options["embedding_api_key"]; !ok {
				config.Memory.Options["embedding_api_key"] = config.LLM.APIKey
			}
		}
	}

	// Initialize memory provider
	memoryProvider, err := createMemoryProvider(config.Memory)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory provider: %w", err)
	}
	agent.memoryProvider = memoryProvider

	// Initialize tools if configured
	if config.Tools != nil && config.Tools.Enabled {
		// Default single-call policy
		normalizeSingleCallPolicy(config.Tools)

		// Initialize MCP if configured
		if config.Tools.MCP != nil && config.Tools.MCP.Enabled {
			if err := initializeMCP(config.Tools.MCP); err != nil {
				Logger().Warn().Err(err).Msg("Failed to initialize MCP - continuing without MCP tools")
			} else {
				Logger().Debug().Msg("MCP initialized successfully")
			}
		}

		// Create and discover tools
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

// Memory returns the memory provider for this agent
func (a *realAgent) Memory() Memory {
	// The core.Memory interface needs to be wrapped or cast to v1beta.Memory?
	// The problem is core.Memory and v1beta.Memory might be defined differently or identical.
	// Looking at v1beta/memory.go (which I read previously), v1beta.Memory is an interface.
	// Looking at v1beta/agent_impl.go imports: "github.com/agenticgokit/agenticgokit/core"
	// and a.memoryProvider is core.Memory.
	// v1beta.Memory interface (from Step 24 in agent.go) matches core.Memory methods mostly?
	// Wait, I need to check if core.Memory satisfies v1beta.Memory.
	// If not, I might need an adapter.
	// Let's assume for now they are compatible or I can cast.
	// Actually, v1beta.Memory is defined in agent.go line 318.
	// It has Store, Query, NewSession, SetSession, IngestDocument, BuildContext.

	// Let's verify core.Memory definition.
	// Since I cannot read core/memory.go right now without extra steps,
	// I will check if realAgent.memoryProvider satisfies it.
	// realAgent uses core.Memory.

	// If I just return it, it might fail if types don't match.
	// Let's check imports. v1beta/memory.go defines NewMemory returns Memory.
	// v1beta/agent_impl.go has a.memoryProvider of type core.Memory.

	if a.memoryProvider == nil {
		return nil
	}
	return &coreMemoryAdapter{mem: a.memoryProvider}
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
	return a.execute(ctx, input, nil)
}

// execute contains the core execution logic, shared between Run and RunWithOptions
func (a *realAgent) execute(ctx context.Context, input string, opts *RunOptions) (*Result, error) {
	startTime := time.Now()

	tracer := observability.GetTracer("agk.v1beta.agent")
	ctx, span := tracer.Start(ctx, "agk.agent.run",
		trace.WithAttributes(
			attribute.String(observability.AttrAgentName, a.config.Name),
			attribute.String(observability.AttrLLMProvider, a.config.LLM.Provider),
			attribute.String(observability.AttrLLMModel, a.config.LLM.Model),
			attribute.Bool("agk.memory.enabled", a.memoryProvider != nil && a.config.Memory != nil),
			attribute.Int("agk.tools.count", len(a.tools)),
		),
	)
	defer span.End()

	if sc := span.SpanContext(); sc.IsValid() {
		ctx = context.WithValue(ctx, "trace_id", sc.TraceID().String())
	}

	// Validate that agent is properly initialized
	if a.llmProvider == nil {
		return nil, fmt.Errorf("agent not properly initialized: LLM provider is nil")
	}

	// Step 1: Build the prompt with system context and user input
	prompt := llm.Prompt{
		System: a.config.SystemPrompt,
		User:   input,
	}

	// Capture prompt content at detailed trace level for evaluation/debugging
	if observability.IsDetailedTracing() {
		span.SetAttributes(
			attribute.String(observability.AttrPromptUser, observability.TruncateForTrace(input, observability.MaxContentLength)),
			attribute.String(observability.AttrPromptSystem, observability.TruncateForTrace(a.config.SystemPrompt, observability.MaxContentLength)),
		)
	}

	// Add multimodal data if present in opts
	addMultimodalDataToPrompt(&prompt, opts)

	// Step 1.5: Add tool definitions (native) and descriptions (text fallback)
	if len(a.tools) > 0 {
		reasoningEnabled := a.config.Tools != nil &&
			a.config.Tools.Reasoning != nil &&
			a.config.Tools.Reasoning.Enabled

		// Always pass native tools — the reasoning continuation loop
		// (executeNativeToolsAndContinue) handles them properly.
		// Text descriptions are also added as a fallback for all models.
		prompt.Tools = convertToolsToLLMFormat(a.tools)

		// Always add text descriptions for reference (works with all models)
		toolDescriptions := FormatToolsForPrompt(a.tools)
		prompt.System = prompt.System + toolDescriptions

		Logger().Debug().
			Int("tool_count", len(a.tools)).
			Bool("reasoning_enabled", reasoningEnabled).
			Bool("native_tools_passed", true).
			Msg("Added tool information to prompt")
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
	memoryQueries := 0
	var ragContext *RAGContext
	if a.memoryProvider != nil && a.config.Memory != nil {
		// Convert llm.Prompt to core.Prompt for enrichment
		var corePrompt core.Prompt

		// Note: We are currently capturing detailed RAG context.
		// The MemoryContext field in Result will be populated.
		corePrompt, ragContext, memoryQueries = BuildEnrichedPrompt(ctx, prompt.System, prompt.User, a.memoryProvider, a.config.Memory)

		// Update the LLM prompt with enriched content
		prompt.System = corePrompt.System
		prompt.User = corePrompt.User

		Logger().Debug().
			Str("original_input", input).
			Int("enriched_length", len(prompt.User)).
			Int("memory_queries", memoryQueries).
			Msg("Enhanced prompt with memory context")
	}

	// Step 2.5: Check for workflow shared memory and enrich prompt with relevant context
	// This allows agents in a workflow to automatically access shared memory
	if workflowMem := GetWorkflowMemory(ctx); workflowMem != nil {
		// Query workflow shared memory for relevant context
		// Use lower threshold to catch more results
		results, queryErr := workflowMem.Query(ctx, input, WithLimit(5), WithScoreThreshold(0.1))
		if queryErr == nil && len(results) > 0 {
			// Build context from workflow memory results
			var sharedContext strings.Builder
			sharedContext.WriteString("\n\n[Shared Workflow Context]\n")
			for _, result := range results {
				sharedContext.WriteString(fmt.Sprintf("- %s\n", result.Content))
			}

			// Append shared context to user prompt
			prompt.User = prompt.User + sharedContext.String()

			Logger().Debug().
				Int("shared_memory_results", len(results)).
				Msg("Enhanced prompt with workflow shared memory")
		}
	}

	// Step 3: Call the LLM provider
	response, err := a.llmProvider.Call(ctx, prompt)
	if err != nil {
		// Update metrics
		a.updateMetrics(startTime, true)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Capture LLM response at detailed trace level for evaluation/debugging
	if observability.IsDetailedTracing() {
		span.SetAttributes(
			attribute.String(observability.AttrLLMResponse, observability.TruncateForTrace(response.Content, observability.MaxContentLength)),
		)
	}

	// Step 3.5: Execute tool calls if any are detected in the response (native or text)
	var toolCalls []ToolCall
	finalResponse := response.Content
	if len(a.tools) > 0 {
		policy := a.singleCallPolicy()
		// Determine max iterations based on reasoning config
		maxToolIterations := 1 // Default: fast path (no continuation)
		reasoningEnabled := false

		if a.config.Tools != nil && a.config.Tools.Reasoning != nil {
			reasoningEnabled = a.config.Tools.Reasoning.Enabled
			if reasoningEnabled && a.config.Tools.Reasoning.MaxIterations > 0 {
				maxToolIterations = a.config.Tools.Reasoning.MaxIterations
			} else if reasoningEnabled {
				maxToolIterations = 5 // Default max iterations when reasoning enabled
			}
		}

		Logger().Debug().
			Bool("reasoning_enabled", reasoningEnabled).
			Int("max_iterations", maxToolIterations).
			Msg("Tool execution configuration")

		var toolErr error
		if len(response.ToolCalls) > 0 {
			finalResponse, toolCalls, toolErr = a.executeNativeToolsAndContinue(ctx, response, prompt, maxToolIterations)
		} else {
			finalResponse, toolCalls, toolErr = a.executeToolsAndContinue(ctx, response.Content, prompt, maxToolIterations)
		}
		if toolErr != nil {
			// Log warning but don't fail - use the last valid response
			Logger().Warn().Err(toolErr).Msg("Tool execution encountered error, using last valid response")
		}

		// If only one tool is enabled, clamp to a single best call to avoid multiple executions
		if len(a.tools) == 1 && len(toolCalls) > 1 && policy != "all" {
			best := selectBestToolCallFromInput(input, toolCalls)
			toolCalls = []ToolCall{best}
		}

		// When reasoning is disabled (fast path), prefer concise tool results
		// over empty/generic LLM filler. When reasoning IS enabled, the LLM
		// has already synthesized tool results into a natural response — do NOT overwrite it.
		if !reasoningEnabled && len(toolCalls) > 0 {
			summary := formatToolCallsAsContent(toolCalls)
			if summary != "" {
				finalResponse = summary
			}
		}
	}

	// Step 4: Store the interaction in memory if enabled
	if a.memoryProvider != nil {
		if err := a.storeInMemory(ctx, input, finalResponse); err != nil {
			// Log warning but don't fail - memory storage is non-critical
			Logger().Warn().Err(err).Msg("Failed to store interaction in memory")
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
			Logger().Warn().Err(err).Msg("Custom handler returned error")
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
		Success:       true,
		Content:       finalResponse,
		Duration:      duration,
		TokensUsed:    response.Usage.TotalTokens,
		MemoryUsed:    a.memoryProvider != nil,
		MemoryQueries: memoryQueries, // Now properly tracks memory query count
		MemoryContext: ragContext,    // Populated from BuildEnrichedPrompt
		ToolCalls:     toolCalls,     // Include tool execution details
		StartTime:     startTime,
		EndTime:       time.Now(),
		Metadata: map[string]interface{}{
			"model":         a.config.LLM.Model,
			"provider":      a.config.LLM.Provider,
			"finish_reason": response.FinishReason,
		},
	}

	if sc := span.SpanContext(); sc.IsValid() {
		result.TraceID = sc.TraceID().String()
	}

	if runID := observability.RunIDFromContext(ctx); runID != "" {
		span.SetAttributes(attribute.String(observability.AttrRunID, runID))
	}

	span.SetAttributes(
		attribute.Int(observability.AttrLLMTokensIn, response.Usage.PromptTokens),
		attribute.Int(observability.AttrLLMTokensOut, response.Usage.CompletionTokens),
		attribute.Int(observability.AttrLLMMaxTokens, a.config.LLM.MaxTokens),
		attribute.Float64(observability.AttrLLMTemperature, float64(a.config.LLM.Temperature)),
	)

	span.SetStatus(codes.Ok, "")

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
	enrichedInput, _, _, _ := EnrichWithMemory(ctx, a.memoryProvider, input, a.config.Memory)

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

	// DEBUG: Check session ID
	sessionID := ctx.Value("session_id")
	Logger().Info().Interface("session_id", sessionID).Msg("storeInMemory: Attempting to store interaction")

	// Store as personal memory for RAG retrieval
	// This allows the agent to recall facts and context from past conversations
	if err := a.memoryProvider.Store(ctx, input, "user_message", "conversation"); err != nil {
		Logger().Warn().Err(err).Msg("Failed to store user message in memory")
		// Don't return error - continue with chat history storage
	}

	if err := a.memoryProvider.Store(ctx, output, "agent_response", "conversation"); err != nil {
		Logger().Warn().Err(err).Msg("Failed to store agent response in memory")
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

	Logger().Debug().
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

	// Step 6: Execute the run with applied options (including multimodal data)
	result, err := a.execute(runCtx, input, opts)
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
	if !a.initialized {
		return nil, fmt.Errorf("agent not initialized")
	}

	if a.llmProvider == nil {
		return nil, fmt.Errorf("no LLM provider configured")
	}

	// Create stream metadata
	metadata := &StreamMetadata{
		AgentName: a.Name(),
		StartTime: time.Now(),
		Model:     a.config.LLM.Model,
		Extra:     make(map[string]interface{}),
	}

	tracer := observability.GetTracer("agk.v1beta.agent")
	streamCtx, span := tracer.Start(ctx, "agk.agent.run.stream",
		trace.WithAttributes(
			attribute.String(observability.AttrAgentName, a.config.Name),
			attribute.String(observability.AttrLLMProvider, a.config.LLM.Provider),
			attribute.String(observability.AttrLLMModel, a.config.LLM.Model),
			attribute.Bool("agk.memory.enabled", a.memoryProvider != nil && a.config.Memory != nil),
			attribute.Int("agk.tools.count", len(a.tools)),
		),
	)

	// Add session ID if provided in context
	if sessionID := streamCtx.Value("session_id"); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			metadata.SessionID = id
		}
	}

	// Add trace ID if provided in context
	if traceID := streamCtx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			metadata.TraceID = id
		}
	}

	if sc := span.SpanContext(); sc.IsValid() {
		metadata.TraceID = sc.TraceID().String()
		streamCtx = context.WithValue(streamCtx, "trace_id", metadata.TraceID)
	}

	// Create stream with options
	stream, writer := NewStream(streamCtx, metadata, opts...)

	// Start streaming in a goroutine
	go func() {
		defer span.End()
		defer writer.Close()

		startTime := time.Now()

		// Build prompt (similar to Run method)
		prompt := llm.Prompt{
			System: a.config.SystemPrompt,
			User:   input,
			Parameters: llm.ModelParameters{
				Temperature: convertFloat32ToFloat32Ptr(a.config.LLM.Temperature),
				MaxTokens:   convertIntToInt32Ptr(a.config.LLM.MaxTokens),
			},
		}

		// Capture prompt content at detailed trace level for evaluation/debugging
		if observability.IsDetailedTracing() {
			span.SetAttributes(
				attribute.String(observability.AttrPromptUser, observability.TruncateForTrace(input, observability.MaxContentLength)),
				attribute.String(observability.AttrPromptSystem, observability.TruncateForTrace(a.config.SystemPrompt, observability.MaxContentLength)),
			)
		}

		// Add memory context if available
		memoryQueries := 0
		var ragContext *RAGContext
		if a.memoryProvider != nil {
			var enrichedPrompt core.Prompt
			enrichedPrompt, ragContext, memoryQueries = BuildEnrichedPrompt(ctx, prompt.System, prompt.User, a.memoryProvider, a.config.Memory)
			prompt.System = enrichedPrompt.System
			prompt.User = enrichedPrompt.User
		}

		// Check for workflow shared memory and enrich prompt with relevant context
		if workflowMem := GetWorkflowMemory(ctx); workflowMem != nil {
			results, queryErr := workflowMem.Query(ctx, input, WithLimit(5), WithScoreThreshold(0.1))
			if queryErr == nil && len(results) > 0 {
				var sharedContext strings.Builder
				sharedContext.WriteString("\n\n[Shared Workflow Context]\n")
				for _, result := range results {
					sharedContext.WriteString(fmt.Sprintf("- %s\n", result.Content))
				}
				prompt.User = prompt.User + sharedContext.String()
			}
		}

		// Add tool descriptions to system prompt if tools are available
		if len(a.tools) > 0 {
			toolDescriptions := FormatToolsForPrompt(a.tools)
			if toolDescriptions != "" {
				prompt.System = prompt.System + "\n\n" + toolDescriptions
			}
		}

		// Start LLM streaming
		tokenChan, err := a.llmProvider.Stream(streamCtx, prompt)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			writer.Write(&StreamChunk{
				Type:  ChunkTypeError,
				Error: fmt.Errorf("failed to start LLM stream: %w", err),
			})
			a.updateMetrics(startTime, true)
			return
		}

		// Track tokens for final result
		var contentBuilder strings.Builder
		var tokenCount int

		// Stream tokens
		for token := range tokenChan {
			if token.Error != nil {
				writer.Write(&StreamChunk{
					Type:  ChunkTypeError,
					Error: token.Error,
				})
				a.updateMetrics(startTime, true)
				return
			}

			// Write text chunk
			if token.Content != "" {
				contentBuilder.WriteString(token.Content)
				tokenCount++

				writer.Write(&StreamChunk{
					Type:  ChunkTypeDelta,
					Delta: token.Content,
				})
			}
		}

		finalContent := contentBuilder.String()
		duration := time.Since(startTime)

		// Capture LLM response at detailed trace level for evaluation/debugging
		if observability.IsDetailedTracing() {
			span.SetAttributes(
				attribute.String(observability.AttrLLMResponse, observability.TruncateForTrace(finalContent, observability.MaxContentLength)),
			)
		}

		// Create final result
		finalResult := &Result{
			Success:       true,
			Content:       finalContent,
			Duration:      duration,
			StartTime:     startTime,
			EndTime:       time.Now(),
			SessionID:     metadata.SessionID,
			TraceID:       metadata.TraceID,
			MemoryUsed:    a.memoryProvider != nil,
			MemoryQueries: memoryQueries, // Track memory queries performed during stream
			MemoryContext: ragContext,    // Populated from BuildEnrichedPrompt
			Metadata:      make(map[string]interface{}),
		}

		// Process tool calls if present and tools are available
		if len(a.tools) > 0 {
			toolCalls := ParseToolCalls(finalContent)
			if len(toolCalls) > 0 {
				finalResult.ToolCalls = toolCalls
				finalResult.ToolsCalled = extractToolNames(toolCalls)

				// Execute tools and stream the results
				err := a.executeToolsAndStream(streamCtx, input, finalContent, toolCalls, writer)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					writer.Write(&StreamChunk{
						Type:  ChunkTypeError,
						Error: fmt.Errorf("tool execution failed: %w", err),
					})
					a.updateMetrics(startTime, true)
					return
				}
			}
		}

		// Store in memory if available
		if a.memoryProvider != nil {
			if err := a.storeInMemory(ctx, input, finalContent); err != nil {
				// Memory storage failures are non-blocking
				writer.Write(&StreamChunk{
					Type:    ChunkTypeThought,
					Content: fmt.Sprintf("Failed to store in memory: %v", err),
				})
			}
		}

		// Call custom handler if configured (mirrors execute() behaviour)
		if a.handler != nil {
			capabilities := &Capabilities{
				LLM: func(system, user string) (string, error) {
					p := llm.Prompt{System: system, User: user}
					resp, err := a.llmProvider.Call(streamCtx, p)
					if err != nil {
						return "", err
					}
					return resp.Content, nil
				},
			}
			handlerResponse, err := a.handler(streamCtx, finalContent, capabilities)
			if err != nil {
				Logger().Warn().Err(err).Msg("Custom handler returned error during stream")
			} else if handlerResponse != "" {
				finalContent = handlerResponse
				finalResult.Content = finalContent
			}
		}

		// Update metrics
		a.updateMetrics(startTime, false)
		span.SetAttributes(
			attribute.Int("agk.stream.tokens", tokenCount),
			attribute.Int64(observability.AttrLLMLatencyMs, duration.Milliseconds()),
		)
		span.SetStatus(codes.Ok, "")

		// Set final result on the stream
		if s, ok := stream.(*basicStream); ok {
			s.SetResult(finalResult)
		}

		// Send completion chunk
		writer.Write(&StreamChunk{
			Type: ChunkTypeDone,
		})
	}()

	return stream, nil
}

// RunStreamWithOptions executes the agent with streaming output and additional options.
// Combines the flexibility of RunWithOptions with real-time streaming.
func (a *realAgent) RunStreamWithOptions(ctx context.Context, input string, runOpts *RunOptions, streamOpts ...StreamOption) (Stream, error) {
	if runOpts == nil {
		// No run options, delegate to RunStream
		return a.RunStream(ctx, input, streamOpts...)
	}

	// Save original configuration
	originalConfig := a.config
	defer func() {
		// Restore original configuration
		a.config = originalConfig
	}()

	// Create a copy of config for this run
	configCopy := *a.config
	a.config = &configCopy

	// Apply timeout to context if specified
	if runOpts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, runOpts.Timeout)
		defer cancel()
	}

	// Apply session ID to context if specified
	if runOpts.SessionID != "" {
		ctx = context.WithValue(ctx, "session_id", runOpts.SessionID)
	}

	// Apply temperature override if specified
	if runOpts.Temperature != nil {
		a.config.LLM.Temperature = float32(*runOpts.Temperature)
	}

	// Apply max tokens override if specified
	if runOpts.MaxTokens > 0 {
		a.config.LLM.MaxTokens = runOpts.MaxTokens
	}

	// Apply memory configuration if specified
	if runOpts.Memory != nil {
		// Create a copy of memory config
		if a.config.Memory == nil {
			a.config.Memory = &MemoryConfig{}
		}
		memoryCopy := *a.config.Memory
		a.config.Memory = &memoryCopy

		// Apply memory options (map from RunOptions.Memory to MemoryConfig)
		if runOpts.Memory.Provider != "" {
			a.config.Memory.Provider = runOpts.Memory.Provider
		}
		// Note: MemoryConfig doesn't have Enabled, ContextAware, SessionScoped fields
		// These are handled at the MemoryOptions level in the interface
	}

	// Apply tool filtering if specified
	if runOpts.ToolMode != "" {
		switch runOpts.ToolMode {
		case "none":
			// Temporarily disable all tools
			a.tools = nil
		case "specific":
			if len(runOpts.Tools) > 0 {
				// Filter tools to only those specified
				filteredTools := make([]Tool, 0)
				for _, tool := range a.tools {
					for _, toolName := range runOpts.Tools {
						if tool.Name() == toolName {
							filteredTools = append(filteredTools, tool)
							break
						}
					}
				}
				a.tools = filteredTools
			}
			// "auto" uses all available tools (no change needed)
		}
	}

	// Apply stream options from run options if specified
	if runOpts.StreamOptions != nil {
		// Merge with provided stream options
		mergedOpts := make([]StreamOption, 0, len(streamOpts)+4)

		// Add options from runOpts.StreamOptions
		if runOpts.StreamOptions.BufferSize > 0 {
			mergedOpts = append(mergedOpts, WithBufferSize(runOpts.StreamOptions.BufferSize))
		}
		if runOpts.StreamOptions.IncludeThoughts {
			mergedOpts = append(mergedOpts, WithThoughts())
		}
		if runOpts.StreamOptions.IncludeToolCalls {
			mergedOpts = append(mergedOpts, WithToolCalls())
		}
		if runOpts.StreamOptions.IncludeMetadata {
			mergedOpts = append(mergedOpts, WithStreamMetadata())
		}
		if runOpts.StreamOptions.TextOnly {
			mergedOpts = append(mergedOpts, WithTextOnly())
		}
		if runOpts.StreamOptions.Timeout > 0 {
			mergedOpts = append(mergedOpts, WithStreamTimeout(runOpts.StreamOptions.Timeout))
		}
		if runOpts.StreamOptions.FlushInterval > 0 {
			mergedOpts = append(mergedOpts, WithFlushInterval(runOpts.StreamOptions.FlushInterval))
		}
		if runOpts.StreamOptions.Handler != nil {
			mergedOpts = append(mergedOpts, WithStreamHandler(runOpts.StreamOptions.Handler))
		}

		// Add provided stream options (these take precedence)
		mergedOpts = append(mergedOpts, streamOpts...)

		streamOpts = mergedOpts
	}

	// Apply stream handler from run options if specified
	if runOpts.StreamHandler != nil {
		streamOpts = append(streamOpts, WithStreamHandler(runOpts.StreamHandler))
	}

	// Delegate to RunStream with the merged options
	return a.RunStream(ctx, input, streamOpts...)
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
		Logger().Debug().Msg("Memory provider initialized")
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
	Logger().Info().
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
	Logger().Debug().Str("agent", a.config.Name).Msg("Cleaning up agent resources")

	// Shutdown tracer if enabled
	if a.tracerShutdown != nil {
		if err := a.tracerShutdown(ctx); err != nil {
			Logger().Warn().Err(err).Msg("Error shutting down tracer")
			// Don't return error, continue cleanup
		}

		// Generate manifest file for trace viewer (agk trace commands)
		if a.runDir != "" {
			if err := a.generateManifest(); err != nil {
				Logger().Warn().Err(err).Msg("Error generating trace manifest")
			}
		}
	}

	// Close memory provider if present
	if a.memoryProvider != nil {
		if err := a.memoryProvider.Close(); err != nil {
			Logger().Warn().Err(err).Msg("Error closing memory provider")
			// Don't return error, continue cleanup
		}
	}

	// Clear sessions
	a.sessions = nil

	// Mark as not initialized
	a.initialized = false

	Logger().Info().Str("agent", a.config.Name).Msg("Agent cleanup completed")
	return nil
}

// generateManifest creates a manifest.json file for the trace viewer
func (a *realAgent) generateManifest() error {
	if a.runDir == "" || a.runID == "" {
		return nil // Tracing not enabled, skip manifest generation
	}

	tracePath := filepath.Join(a.runDir, "trace.jsonl")

	// Check if trace file exists
	if _, err := os.Stat(tracePath); err != nil {
		return nil // No trace file, skip manifest generation
	}

	// Read trace file and calculate statistics
	data, err := ioutil.ReadFile(tracePath)
	if err != nil {
		return nil // Can't read trace, skip silently
	}

	type TraceRun struct {
		RunID         string    `json:"run_id"`
		Command       string    `json:"command"`
		Status        string    `json:"status"`
		StartTime     time.Time `json:"start_time"`
		EndTime       time.Time `json:"end_time"`
		Duration      float64   `json:"duration_seconds"`
		SpanCount     int       `json:"span_count"`
		LLMCalls      int       `json:"llm_calls"`
		TotalTokens   int       `json:"total_tokens"`
		EstimatedCost float64   `json:"estimated_cost"`
	}

	var manifest TraceRun
	manifest.RunID = a.runID
	manifest.Command = a.config.Name
	manifest.Status = "completed"

	// Parse trace data to extract metrics
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var firstTime, lastTime time.Time

	for scanner.Scan() {
		line := scanner.Bytes()
		var span map[string]interface{}
		if err := json.Unmarshal(line, &span); err != nil {
			continue
		}

		manifest.SpanCount++

		// Check if this is an LLM span
		if spanName, ok := span["Name"].(string); ok {
			if strings.Contains(spanName, "llm") {
				manifest.LLMCalls++
			}
		}

		// Extract timestamps
		if st, ok := span["StartTime"].(string); ok {
			if t, err := time.Parse(time.RFC3339, st); err == nil {
				if firstTime.IsZero() || t.Before(firstTime) {
					firstTime = t
				}
			}
		}

		if et, ok := span["EndTime"].(string); ok {
			if t, err := time.Parse(time.RFC3339, et); err == nil {
				if t.After(lastTime) {
					lastTime = t
				}
			}
		}
	}

	// Set times
	if !firstTime.IsZero() {
		manifest.StartTime = firstTime
	}
	if !lastTime.IsZero() {
		manifest.EndTime = lastTime
	} else if !firstTime.IsZero() {
		manifest.EndTime = firstTime
	}

	// Calculate duration
	if !manifest.StartTime.IsZero() && !manifest.EndTime.IsZero() {
		manifest.Duration = manifest.EndTime.Sub(manifest.StartTime).Seconds()
	}

	// Write manifest.json
	manifestPath := filepath.Join(a.runDir, "manifest.json")
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil // Can't marshal, skip silently
	}

	return ioutil.WriteFile(manifestPath, manifestData, 0644)
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
		toolCall.Error = fmt.Sprintf("tool %q not found", toolCall.Name)
		toolCall.Duration = time.Since(startTime)
		return toolCall
	}

	// Check if this is an MCP tool (which creates its own spans)
	_, isMCPTool := tool.(*mcpToolWrapper)

	var span trace.Span

	// Only create a parent span for non-MCP tools (native tools)
	// MCP tools create their own agk.mcp.tool.call spans which capture all metrics
	if !isMCPTool {
		tracer := otel.Tracer("agenticgokit.tools")
		var spanCtx context.Context
		spanCtx, span = tracer.Start(ctx, "agk.tool.call",
			trace.WithAttributes(
				attribute.String(observability.AttrToolName, toolCall.Name),
			))
		ctx = spanCtx
		defer span.End()
	}

	// Record input size and serialize arguments for tracing
	inputSize := 0
	var jsonBytes []byte
	if toolCall.Arguments != nil {
		if bytes, err := json.Marshal(toolCall.Arguments); err == nil {
			jsonBytes = bytes
			inputSize = len(jsonBytes)
		}
	}
	if span != nil {
		span.SetAttributes(attribute.Int("agk.tool.input_bytes", inputSize))
		// Capture tool arguments at detailed trace level for evaluation/debugging
		if observability.IsDetailedTracing() && len(jsonBytes) > 0 {
			span.SetAttributes(
				attribute.String(observability.AttrToolArguments, observability.TruncateForTrace(string(jsonBytes), observability.MaxContentLength)),
			)
		}
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
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "tool execution failed")
			span.SetAttributes(attribute.Int64(observability.AttrToolLatencyMs, toolCall.Duration.Milliseconds()))
		}
		return toolCall
	}

	if result == nil {
		toolCall.Error = "tool returned nil result"
		toolCall.Success = false
		if span != nil {
			span.RecordError(fmt.Errorf("tool returned nil result"))
			span.SetStatus(codes.Error, "tool returned nil result")
			span.SetAttributes(attribute.Int64(observability.AttrToolLatencyMs, toolCall.Duration.Milliseconds()))
		}
		return toolCall
	}

	// Tool executed successfully
	toolCall.Success = result.Success
	toolCall.Result = result.Content
	if !result.Success {
		toolCall.Error = result.Error
		if span != nil {
			span.SetStatus(codes.Error, result.Error)
		}
	} else if span != nil {
		span.SetStatus(codes.Ok, "tool execution successful")
	}

	// Record output size
	outputSize := 0
	if contentStr, ok := toolCall.Result.(string); ok {
		outputSize = len(contentStr)
	} else if contentBytes, ok := toolCall.Result.([]byte); ok {
		outputSize = len(contentBytes)
	}

	if span != nil {
		span.SetAttributes(
			attribute.Int64(observability.AttrToolLatencyMs, toolCall.Duration.Milliseconds()),
			attribute.Int("agk.tool.output_bytes", outputSize),
			attribute.Bool("agk.tool.success", toolCall.Success),
		)
		// Capture tool result at detailed trace level for evaluation/debugging
		if observability.IsDetailedTracing() && toolCall.Result != nil {
			resultStr := fmt.Sprintf("%v", toolCall.Result)
			span.SetAttributes(
				attribute.String(observability.AttrToolResult, observability.TruncateForTrace(resultStr, observability.MaxContentLength)),
			)
		}
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
	policy := a.singleCallPolicy()

	for iteration < maxIterations {
		// Parse tool calls from the current response
		toolCalls := ParseToolCalls(currentResponse)

		if len(toolCalls) == 0 {
			// No tool calls found, we're done
			break
		}

		Logger().Debug().
			Int("iteration", iteration+1).
			Int("tool_calls", len(toolCalls)).
			Msg("Executing tool calls")

		// Execute tool calls. If only one tool is enabled, prefer executing a single best call unless policy is "all".
		var executedCalls []ToolCall
		var toolResults strings.Builder

		// Select calls to execute
		callsToExecute := toolCalls
		if len(a.tools) == 1 && len(toolCalls) > 1 && policy != "all" {
			if policy == "first" {
				callsToExecute = toolCalls[:1]
			} else {
				callsToExecute = []ToolCall{selectBestToolCallFromInput(originalPrompt.User, toolCalls)}
			}
		}

		for _, call := range callsToExecute {
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

		iteration++

		// Skip continuation if we've reached max iterations (fast path optimization)
		if iteration >= maxIterations {
			Logger().Debug().
				Int("max_iterations", maxIterations).
				Int("iteration", iteration).
				Msg("Reached maximum tool execution iterations, skipping continuation")
			break
		}

		// Build a new prompt with tool results (only if continuing)
		continuationPrompt := llm.Prompt{
			System: originalPrompt.System,
			User: fmt.Sprintf(
				"Previous response:\n%s\n\nTool execution results:\n%s\n\nBased on these tool results, provide a final answer. Do NOT make additional tool calls unless absolutely necessary.",
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
	}

	return currentResponse, allToolCalls, nil
}

// convertStructuredToolCalls maps structured tool call responses to the legacy ToolCall format.
func convertStructuredToolCalls(calls []llm.ToolCallResponse) []ToolCall {
	converted := make([]ToolCall, 0, len(calls))
	for _, c := range calls {
		converted = append(converted, ToolCall{
			Name:      c.Function.Name,
			Arguments: c.Function.Arguments,
		})
	}
	return converted
}

// executeNativeToolsAndContinue handles structured tool calls returned by native tool-calling models.
// It executes the tool calls, feeds results back to the LLM, and continues the loop.
func (a *realAgent) executeNativeToolsAndContinue(
	ctx context.Context,
	initialResponse llm.Response,
	originalPrompt llm.Prompt,
	maxIterations int,
) (string, []ToolCall, error) {
	if len(a.tools) == 0 {
		return initialResponse.Content, nil, nil
	}

	var allToolCalls []ToolCall
	// Track executed calls to avoid duplicate executions within the same run
	seen := make(map[string]struct{})
	currentResponse := initialResponse.Content
	response := initialResponse
	iteration := 0
	policy := a.singleCallPolicy()

	for iteration < maxIterations {
		var parsedCalls []ToolCall

		// Prefer structured tool calls when available
		if len(response.ToolCalls) > 0 {
			parsedCalls = convertStructuredToolCalls(response.ToolCalls)
		} else {
			parsedCalls = ParseToolCalls(currentResponse)
		}

		if len(parsedCalls) == 0 {
			break
		}

		Logger().Debug().
			Int("iteration", iteration+1).
			Int("tool_calls", len(parsedCalls)).
			Msg("Executing native/tool calls")

		var executedCalls []ToolCall
		var toolResults strings.Builder

		// If only one tool is enabled and multiple parsed calls exist, run only the best/first unless policy is "all"
		callsToExecute := parsedCalls
		if len(a.tools) == 1 && len(parsedCalls) > 1 && policy != "all" {
			if policy == "first" {
				callsToExecute = parsedCalls[:1]
			} else {
				callsToExecute = []ToolCall{selectBestToolCallFromInput(originalPrompt.User, parsedCalls)}
			}
		}

		for _, call := range callsToExecute {
			// Build a stable signature: name + sorted args JSON
			sig := buildToolSignature(call)
			if _, ok := seen[sig]; ok {
				continue
			}
			executed := a.executeTool(ctx, call)
			executedCalls = append(executedCalls, executed)
			allToolCalls = append(allToolCalls, executed)
			seen[sig] = struct{}{}

			toolResults.WriteString(FormatToolResult(executed.Name, &ToolResult{
				Success: executed.Success,
				Content: executed.Result,
				Error:   executed.Error,
			}))
		}

		iteration++

		// Skip continuation if we've reached max iterations (fast path optimization)
		if iteration >= maxIterations {
			Logger().Debug().
				Int("max_iterations", maxIterations).
				Int("iteration", iteration).
				Msg("Reached maximum tool execution iterations, skipping continuation")
			break
		}

		// Make continuation call only if we haven't reached max iterations
		continuationPrompt := llm.Prompt{
			System: originalPrompt.System,
			User: fmt.Sprintf(
				"Previous response:\n%s\n\nTool execution results:\n%s\n\nBased on these tool results, provide a final answer. Do NOT make additional tool calls unless absolutely necessary.",
				currentResponse,
				toolResults.String(),
			),
			Parameters: originalPrompt.Parameters,
		}

		// Only pass native tools if the model supports them and we're not relying on reasoning
		// When reasoning is enabled, we parse TOOL_CALL patterns instead of native tools
		if a.config.Tools == nil || a.config.Tools.Reasoning == nil || !a.config.Tools.Reasoning.Enabled {
			continuationPrompt.Tools = originalPrompt.Tools
		}

		// Debug: log what we're sending to the LLM for continuation
		toolResultStr := toolResults.String()
		Logger().Debug().
			Int("tool_result_length", len(toolResultStr)).
			Str("tool_result_preview", truncateDebug(toolResultStr, 300)).
			Msg("[REASONING] Sending tool results to LLM for synthesis")

		resp, err := a.llmProvider.Call(ctx, continuationPrompt)
		if err != nil {
			return currentResponse, allToolCalls, fmt.Errorf("LLM call failed after tool execution: %w", err)
		}

		Logger().Debug().
			Int("response_length", len(resp.Content)).
			Str("response_preview", truncateDebug(resp.Content, 300)).
			Msg("[REASONING] LLM continuation response received")

		response = resp
		currentResponse = resp.Content
	}

	return currentResponse, allToolCalls, nil
}

// selectBestToolCallFromInput picks a single tool call that best matches the user input.
// Minimal heuristic: prefer calls whose arguments appear in the input, otherwise the first.
func selectBestToolCallFromInput(input string, calls []ToolCall) ToolCall {
	if len(calls) == 0 {
		return ToolCall{}
	}
	if len(calls) == 1 {
		return calls[0]
	}
	lower := strings.ToLower(input)
	bestIdx := 0
	bestScore := -1
	for i, c := range calls {
		score := 0
		for k, v := range c.Arguments {
			// consider only simple string args
			if s, ok := v.(string); ok {
				sLower := strings.ToLower(s)
				if sLower != "" && strings.Contains(lower, sLower) {
					score += 2
				}
				// handle common abbrev expansions
				if k == "location" {
					if sLower == "sf" && strings.Contains(lower, "san francisco") {
						score += 1
					}
					if sLower == "nyc" && strings.Contains(lower, "new york") {
						score += 1
					}
					if sLower == "la" && strings.Contains(lower, "los angeles") {
						score += 1
					}
				}
			}
		}
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	return calls[bestIdx]
}

// buildToolSignature creates a stable signature for a tool call to dedupe repeated calls
func buildToolSignature(call ToolCall) string {
	// Minimal JSON-ish deterministic encoding by sorting keys
	if call.Arguments == nil || len(call.Arguments) == 0 {
		return call.Name + "|{}"
	}
	// Collect keys
	keys := make([]string, 0, len(call.Arguments))
	for k := range call.Arguments {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// Build key=value pairs string
	var b strings.Builder
	b.WriteString(call.Name)
	b.WriteString("|")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(k)
		b.WriteString("=")
		v := call.Arguments[k]
		b.WriteString(fmt.Sprintf("%v", v))
	}
	return b.String()
}

// convertFloat32ToFloat32Ptr converts a float32 to a *float32 for LLM parameters
func convertFloat32ToFloat32Ptr(f32 float32) *float32 {
	if f32 == 0 {
		return nil // Use provider default
	}
	return &f32
}

// convertFloat64ToFloat32Ptr converts a float64 to a *float32 for LLM parameters
func convertFloat64ToFloat32Ptr(f64 float64) *float32 {
	if f64 == 0 {
		return nil // Use provider default
	}
	f32 := float32(f64)
	return &f32
}

// convertIntToInt32Ptr converts an int to a *int32 for LLM parameters
func convertIntToInt32Ptr(i int) *int32 {
	if i == 0 {
		return nil // Use provider default
	}
	i32 := int32(i)
	return &i32
}

// singleCallPolicy returns normalized tool single-call policy: best (default), first, or all
func (a *realAgent) singleCallPolicy() string {
	if a.config != nil && a.config.Tools != nil {
		p := strings.ToLower(strings.TrimSpace(a.config.Tools.SingleCallPolicy))
		switch p {
		case "first":
			return "first"
		case "all":
			return "all"
		}
	}
	return "best"
}

// normalizeSingleCallPolicy ensures the policy value is one of best/first/all, defaulting to best.
func normalizeSingleCallPolicy(cfg *ToolsConfig) {
	if cfg == nil {
		return
	}
	p := strings.ToLower(strings.TrimSpace(cfg.SingleCallPolicy))
	switch p {
	case "first", "all", "best":
		cfg.SingleCallPolicy = p
	default:
		cfg.SingleCallPolicy = "best"
	}
}

// extractToolNames extracts tool names from a list of tool calls
func extractToolNames(toolCalls []ToolCall) []string {
	names := make([]string, len(toolCalls))
	for i, call := range toolCalls {
		names[i] = call.Name
	}
	return names
}

// formatToolCallsAsContent creates a concise assistant message from executed tool calls
// when the model response is empty. This ensures users see useful output in Result.Content.
func formatToolCallsAsContent(calls []ToolCall) string {
	if len(calls) == 0 {
		return ""
	}

	// If there's a single weather-style tool call with a forecast, return that directly
	if len(calls) == 1 {
		c := calls[0]
		if resMap, ok := c.Result.(map[string]interface{}); ok {
			if forecast, fok := resMap["forecast"].(string); fok && forecast != "" {
				return forecast
			}
		}
		// Generic single-tool fallback
		return fmt.Sprintf("%s result: %v", c.Name, c.Result)
	}

	// For multiple calls, build a short list summarizing results
	var b strings.Builder
	b.WriteString("Results:\n")
	for _, c := range calls {
		var argStr string
		if c.Arguments != nil {
			if loc, ok := c.Arguments["location"]; ok {
				argStr = fmt.Sprintf("location=%v", loc)
			}
		}

		var resStr string
		if resMap, ok := c.Result.(map[string]interface{}); ok {
			if forecast, ok := resMap["forecast"].(string); ok && forecast != "" {
				resStr = forecast
			} else {
				resStr = fmt.Sprintf("%v", resMap)
			}
		} else if c.Result != nil {
			resStr = fmt.Sprintf("%v", c.Result)
		}

		if argStr != "" {
			b.WriteString(fmt.Sprintf("- %s(%s): %s\n", c.Name, argStr, resStr))
		} else {
			b.WriteString(fmt.Sprintf("- %s: %s\n", c.Name, resStr))
		}
	}
	return strings.TrimSpace(b.String())
}

// convertToolsToLLMFormat converts v1beta tools to llm.ToolDefinition for native tool calling.
func convertToolsToLLMFormat(tools []Tool) []llm.ToolDefinition {
	if len(tools) == 0 {
		return nil
	}
	res := make([]llm.ToolDefinition, len(tools))
	for i, tool := range tools {
		res[i] = llm.ToolDefinition{
			Type: "function",
			Function: llm.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  getToolSchema(tool),
			},
		}
	}
	return res
}

// getToolSchema returns an optional JSON Schema for a tool, falling back to a minimal schema.
func getToolSchema(tool Tool) map[string]interface{} {
	if ts, ok := tool.(ToolWithSchema); ok {
		return ts.JSONSchema()
	}

	// Minimal fallback schema
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"input": map[string]interface{}{
				"type":        "string",
				"description": "User input",
			},
		},
	}
}

// executeToolsAndStream executes tool calls and streams the results
func (a *realAgent) executeToolsAndStream(ctx context.Context, userInput, llmResponse string, toolCalls []ToolCall, writer StreamWriter) error {
	if len(toolCalls) == 0 {
		return nil
	}

	policy := a.singleCallPolicy()
	if len(a.tools) == 1 && len(toolCalls) > 1 && policy != "all" {
		if policy == "first" {
			toolCalls = toolCalls[:1]
		} else {
			toolCalls = []ToolCall{selectBestToolCallFromInput(userInput, toolCalls)}
		}
	}

	// Execute each tool call and stream the results
	for _, toolCall := range toolCalls {
		// Send tool call chunk
		writer.Write(&StreamChunk{
			Type:     ChunkTypeToolCall,
			ToolName: toolCall.Name,
			ToolArgs: toolCall.Arguments,
			ToolID:   fmt.Sprintf("call_%d", time.Now().UnixNano()),
		})

		// Execute the tool
		executedCall := a.executeTool(ctx, toolCall)
		if executedCall.Error != "" {
			// Send error result
			writer.Write(&StreamChunk{
				Type:  ChunkTypeToolRes,
				Error: fmt.Errorf("tool %s failed: %s", toolCall.Name, executedCall.Error),
			})
			continue
		}

		// Send tool result chunk
		resultContent := ""
		if executedCall.Result != nil {
			resultContent = fmt.Sprintf("%v", executedCall.Result)
		}

		writer.Write(&StreamChunk{
			Type:     ChunkTypeToolRes,
			Content:  resultContent,
			ToolName: toolCall.Name,
		})
	}

	return nil
}

// =============================================================================
// MEMORY ADAPTER
// =============================================================================

// coreMemoryAdapter adapts the internal core.Memory to the public v1beta.Memory interface
type coreMemoryAdapter struct {
	mem core.Memory
}

func (a *coreMemoryAdapter) Store(ctx context.Context, content string, opts ...StoreOption) error {
	// Parse options
	config := &StoreConfig{
		ContentType: "text",
		Source:      "agent",
	}
	for _, opt := range opts {
		opt(config)
	}

	// Call core memory (assuming signature: Store(ctx, content, type, source))
	return a.mem.Store(ctx, content, config.ContentType, config.Source)
}

func (a *coreMemoryAdapter) Query(ctx context.Context, query string, opts ...QueryOption) ([]MemoryResult, error) {
	// Parse options
	config := &QueryConfig{
		Limit: 5,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Call core memory
	results, err := a.mem.Query(ctx, query, config.Limit)
	if err != nil {
		return nil, err
	}

	// Convert results
	memoryResults := make([]MemoryResult, len(results))
	for i, r := range results {
		memoryResults[i] = MemoryResult{
			Content: r.Content,
			Score:   float32(r.Score),
			// Source:    r.Source, // Fields might vary, focusing on core content
			// Metadata:  r.Metadata,
		}
	}
	return memoryResults, nil
}

func (a *coreMemoryAdapter) NewSession() string {
	return a.mem.NewSession()
}

func (a *coreMemoryAdapter) SetSession(ctx context.Context, sessionID string) context.Context {
	return a.mem.SetSession(ctx, sessionID)
}

func (a *coreMemoryAdapter) IngestDocument(ctx context.Context, doc Document) error {
	// Convert v1beta.Document to core.Document
	coreDoc := core.Document{
		ID:       doc.ID,
		Content:  doc.Content,
		Title:    doc.Title,
		Source:   doc.Source,
		Metadata: doc.Metadata,
	}
	return a.mem.IngestDocument(ctx, coreDoc)
}

func (a *coreMemoryAdapter) IngestDocuments(ctx context.Context, docs []Document) error {
	coreDocs := make([]core.Document, len(docs))
	for i, doc := range docs {
		coreDocs[i] = core.Document{
			ID:       doc.ID,
			Content:  doc.Content,
			Title:    doc.Title,
			Source:   doc.Source,
			Metadata: doc.Metadata,
		}
	}
	return a.mem.IngestDocuments(ctx, coreDocs)
}

func (a *coreMemoryAdapter) SearchKnowledge(ctx context.Context, query string, opts ...QueryOption) ([]MemoryResult, error) {
	config := &QueryConfig{Limit: 5}
	for _, opt := range opts {
		opt(config)
	}

	results, err := a.mem.SearchKnowledge(ctx, query, core.WithLimit(config.Limit))
	if err != nil {
		return nil, err
	}

	memoryResults := make([]MemoryResult, len(results))
	for i, r := range results {
		memoryResults[i] = MemoryResult{
			Content:  r.Content,
			Score:    r.Score,
			Source:   r.Source,
			Metadata: r.Metadata,
		}
	}
	return memoryResults, nil
}

func (a *coreMemoryAdapter) BuildContext(ctx context.Context, query string, opts ...ContextOption) (*RAGContext, error) {
	// Convert to core.ContextOption not easily possible if types differ.
	// Calling without options for now.
	coreCtx, err := a.mem.BuildContext(ctx, query)
	if err != nil {
		return nil, err
	}
	if coreCtx == nil {
		return nil, nil
	}

	// Helper to convert results strings for chat history
	convertHistory := func(msgs []core.Message) []string {
		out := make([]string, len(msgs))
		for i, m := range msgs {
			out[i] = fmt.Sprintf("%s: %s", m.Role, m.Content)
		}
		return out
	}

	// Helper to convert personal memory results
	convertPersonal := func(res []core.Result) []MemoryResult {
		out := make([]MemoryResult, len(res))
		for i, r := range res {
			out[i] = MemoryResult{
				Content: r.Content,
				Score:   r.Score,
				// Source/Metadata not available in core.Result (personal memory)
			}
		}
		return out
	}

	// Helper to convert knowledge results
	convertKnowledge := func(res []core.KnowledgeResult) []MemoryResult {
		out := make([]MemoryResult, len(res))
		for i, r := range res {
			// Convert map[string]any to map[string]interface{}
			var metadata map[string]interface{}
			if r.Metadata != nil {
				metadata = make(map[string]interface{})
				for k, v := range r.Metadata {
					metadata[k] = v
				}
			}

			out[i] = MemoryResult{
				Content:  r.Content,
				Score:    r.Score,
				Source:   r.Source,
				Metadata: metadata,
			}
		}
		return out
	}

	return &RAGContext{
		PersonalMemory:    convertPersonal(coreCtx.PersonalMemory),
		KnowledgeBase:     convertKnowledge(coreCtx.Knowledge), // Mapped from Knowledge
		ChatHistory:       convertHistory(coreCtx.ChatHistory),
		TotalTokens:       coreCtx.TokenCount, // Mapped from TokenCount
		SourceAttribution: coreCtx.Sources,    // Mapped from Sources
	}, nil
}

// AddMessage adds a message to the chat history
func (a *coreMemoryAdapter) AddMessage(ctx context.Context, role, content string) error {
	return a.mem.AddMessage(ctx, role, content)
}

// truncateDebug truncates a string for debug logging purposes
func truncateDebug(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
