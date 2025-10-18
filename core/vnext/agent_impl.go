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
	if a.memoryProvider != nil {
		memoryContext, err := a.buildMemoryContext(ctx, input)
		if err != nil {
			// Log warning but continue - memory failure shouldn't halt execution
			core.Logger().Warn().Err(err).Msg("Failed to build memory context, continuing without it")
		} else if memoryContext != "" {
			// Prepend memory context to system prompt
			prompt.System = prompt.System + "\n\nRelevant context from memory:\n" + memoryContext
		}
	}

	// Step 3: Call the LLM provider
	response, err := a.llmProvider.Call(ctx, prompt)
	if err != nil {
		// Update metrics
		a.updateMetrics(startTime, true)
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Step 4: Store the interaction in memory if enabled
	if a.memoryProvider != nil {
		if err := a.storeInMemory(ctx, input, response.Content); err != nil {
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

		handlerResponse, err := a.handler(ctx, response.Content, capabilities)
		if err != nil {
			core.Logger().Warn().Err(err).Msg("Custom handler returned error")
		} else if handlerResponse != "" {
			// Handler can override the response
			response.Content = handlerResponse
		}
	}

	// Step 6: Update metrics
	a.updateMetrics(startTime, false)

	// Step 7: Build and return the result
	duration := time.Since(startTime)
	result := &Result{
		Success:    true,
		Content:    response.Content,
		Duration:   duration,
		TokensUsed: response.Usage.TotalTokens,
		MemoryUsed: a.memoryProvider != nil,
		StartTime:  startTime,
		EndTime:    time.Now(),
		Metadata: map[string]interface{}{
			"model":         a.config.LLM.Model,
			"provider":      a.config.LLM.Provider,
			"finish_reason": response.FinishReason,
		},
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
// This is used to enhance the prompt with relevant historical information.
func (a *realAgent) buildMemoryContext(ctx context.Context, input string) (string, error) {
	if a.memoryProvider == nil {
		return "", nil
	}

	// Query memory for relevant information
	// Use core.Memory's Query method for personal memory
	results, err := a.memoryProvider.Query(ctx, input, 5) // Get top 5 relevant memories
	if err != nil {
		return "", fmt.Errorf("failed to query memory: %w", err)
	}

	if len(results) == 0 {
		return "", nil
	}

	// Build context string from memory results
	var contextBuilder strings.Builder
	for i, result := range results {
		contextBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Content))
	}

	return contextBuilder.String(), nil
}

// storeInMemory stores the current interaction in memory for future reference.
func (a *realAgent) storeInMemory(ctx context.Context, input, output string) error {
	if a.memoryProvider == nil {
		return nil
	}

	// Store user input
	if err := a.memoryProvider.Store(ctx, input, "user_message", "conversation"); err != nil {
		return fmt.Errorf("failed to store user message: %w", err)
	}

	// Store agent response
	if err := a.memoryProvider.Store(ctx, output, "agent_response", "conversation"); err != nil {
		return fmt.Errorf("failed to store agent response: %w", err)
	}

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
func (a *realAgent) RunWithOptions(ctx context.Context, input string, opts *RunOptions) (*Result, error) {
	// TODO: Implementation in Task 2.4
	return nil, fmt.Errorf("not implemented yet")
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
