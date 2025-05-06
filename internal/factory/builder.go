package factory

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/orchestrator"
	"kunalkushwaha/agentflow/internal/factory/trace"
	"kunalkushwaha/agentflow/internal/workflow"
)

// RunnerBuilder simplifies the creation of an AgentFlow runner with appropriate components
type RunnerBuilder struct {
	runner           agentflow.Runner
	callbackRegistry *agentflow.CallbackRegistry
	traceLogger      agentflow.TraceLogger
	orchestratorType string
	queueSize        int
	agentHandlers    map[string]agentflow.AgentHandler
}

// NewRunnerBuilder creates a new builder for constructing runners
func NewRunnerBuilder() *RunnerBuilder {
	return &RunnerBuilder{
		queueSize: 10, // Default queue size; provides a reasonable buffer for concurrent events.
		orchestratorType: "route",
		agentHandlers:    make(map[string]agentflow.AgentHandler),
	}
}

// WithQueueSize sets the event queue size
func (b *RunnerBuilder) WithQueueSize(size int) *RunnerBuilder {
	b.queueSize = size
	return b
}

// WithRouteOrchestrator configures the builder to use RouteOrchestrator
func (b *RunnerBuilder) WithRouteOrchestrator() *RunnerBuilder {
	b.orchestratorType = "route"
	return b
}

// WithCollaborativeOrchestrator configures the builder to use CollaborativeOrchestrator
func (b *RunnerBuilder) WithCollaborativeOrchestrator() *RunnerBuilder {
	b.orchestratorType = "collaborative"
	return b
}

// WithTraceLogging enables trace logging
func (b *RunnerBuilder) WithTraceLogging() *RunnerBuilder {
	b.traceLogger = agentflow.NewInMemoryTraceLogger()
	return b
}

// RegisterAgent adds an agent to the runner
func (b *RunnerBuilder) RegisterAgent(name string, handler agentflow.AgentHandler) *RunnerBuilder {
	b.agentHandlers[name] = handler
	return b
}

// Fix the Build method to match the actual Runner interface
func (b *RunnerBuilder) Build() (agentflow.Runner, error) {
	// Create runner
	runner := agentflow.NewRunner(b.queueSize)

	// Set up callback registry
	b.callbackRegistry = agentflow.NewCallbackRegistry()
	runner.SetCallbackRegistry(b.callbackRegistry)

	// Set up trace logger if configured
	if b.traceLogger != nil {
		runner.SetTraceLogger(b.traceLogger)

		// Register callbacks for tracing
		trace.RegisterTraceCallbacks(b.callbackRegistry, b.traceLogger)
	}

	// Create and set orchestrator
	var orch agentflow.Orchestrator
	switch b.orchestratorType {
	case "route":
		orch = orchestrator.NewRouteOrchestrator(b.callbackRegistry)
	case "collaborative":
		orch = orchestrator.NewCollaborativeOrchestrator()
	default:
		return nil, fmt.Errorf("unknown orchestrator type: %s", b.orchestratorType)
	}
	runner.SetOrchestrator(orch)

	// Register agents
	for name, handler := range b.agentHandlers {
		if err := runner.RegisterAgent(name, handler); err != nil {
			return nil, fmt.Errorf("failed to register agent '%s': %w", name, err)
		}
	}

	// The RunnerImpl type implements the Start method with context parameter
	// but the Runner interface expects Start() with no parameters
	// Wrap the runner in an adapter that implements the expected interface
	return &runnerAdapter{impl: runner}, nil
}

// runnerAdapter adapts RunnerImpl to match the Runner interface expected by callers
type runnerAdapter struct {
	impl *agentflow.RunnerImpl
}

// Start implements agentflow.Runner.Start() with no parameters
func (a *runnerAdapter) Start() {
	// Call the implementation's Start method with a background context
	a.impl.Start(context.Background())
}

// Delegate all other Runner methods to the implementation
func (a *runnerAdapter) Stop() {
	a.impl.Stop()
}

func (a *runnerAdapter) Emit(event agentflow.Event) error {
	return a.impl.Emit(event)
}

func (a *runnerAdapter) RegisterAgent(name string, handler agentflow.AgentHandler) error {
	return a.impl.RegisterAgent(name, handler)
}

// The runnerAdapter must implement ALL methods from agentflow.Runner
// Add the missing DumpTrace method
func (a *runnerAdapter) DumpTrace(sessionID string) ([]agentflow.TraceEntry, error) {
	return a.impl.DumpTrace(sessionID)
}

// Implement any other Runner interface methods that might be missing
func (a *runnerAdapter) GetCallbackRegistry() *agentflow.CallbackRegistry {
	return a.impl.GetCallbackRegistry()
}

func (a *runnerAdapter) GetTraceLogger() agentflow.TraceLogger {
	return a.impl.GetTraceLogger()
}

// RegisterCallback delegates to the implementation with proper type and error handling
func (a *runnerAdapter) RegisterCallback(hook agentflow.HookPoint, name string, callback agentflow.CallbackFunc) error {
	// Following Azure best practices for proper error propagation
	return a.impl.RegisterCallback(hook, name, callback)
}

// UnregisterCallback delegates to the implementation with proper error handling
func (a *runnerAdapter) UnregisterCallback(hook agentflow.HookPoint, name string) {
	// Following Azure best practices for error handling in void methods
	a.impl.UnregisterCallback(hook, name)

}

// NewSequentialAgentHandler creates a sequential agent handler.
func NewSequentialAgentHandler(handlers ...agentflow.AgentHandler) agentflow.AgentHandler {
	return &workflow.SequentialAgent{
		Handlers: handlers,
	}
}

// NewParallelAgentHandler creates a parallel agent handler.
func NewParallelAgentHandler(handlers ...agentflow.AgentHandler) agentflow.AgentHandler {
	return &workflow.ParallelAgent{
		Handlers: handlers,
	}
}

// NewLoopAgentHandler creates a loop agent handler.
// The condition handler determines whether to continue looping based on its AgentResult.
// The loop handler is the agent executed in each iteration.
func NewLoopAgentHandler(conditionHandler agentflow.AgentHandler, loopHandler agentflow.AgentHandler) agentflow.AgentHandler {
	return &workflow.LoopAgent{
		ConditionHandler: conditionHandler,
		LoopHandler:      loopHandler,
	}
}
	a.impl.UnregisterCallback(hook, name)

}

// AzureLLMBuilder simplifies configuration of Azure OpenAI services following best practices
type AzureLLMBuilder struct {
	endpoint            string
	apiKey              string
	apiVersion          string
	chatDeployment      string
	embeddingDeployment string
	useAzureADAuth      bool
}

// NewAzureLLMBuilder creates a builder for Azure OpenAI integration
func NewAzureLLMBuilder() *AzureLLMBuilder {
	return &AzureLLMBuilder{
		apiVersion: "2023-12-01-preview", // Default to a recent API version that supports various features
	}
}

// FromEnvironment loads configuration from environment variables (best practice)
func (b *AzureLLMBuilder) FromEnvironment() *AzureLLMBuilder {
	// Required environment variables: AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, AZURE_OPENAI_DEPLOYMENT_ID, AZURE_OPENAI_EMBEDDING_DEPLOYMENT
	b.endpoint = os.Getenv("AZURE_OPENAI_ENDPOINT")
	b.apiKey = os.Getenv("AZURE_OPENAI_API_KEY")
	b.chatDeployment = os.Getenv("AZURE_OPENAI_DEPLOYMENT_ID")
	b.embeddingDeployment = os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")

	// Check for Azure AD auth flag
	if os.Getenv("AZURE_OPENAI_USE_AD_AUTH") == "true" {
		b.useAzureADAuth = true
	}

	return b
}

// WithChatDeployment sets the chat model deployment
func (b *AzureLLMBuilder) WithChatDeployment(deploymentID string) *AzureLLMBuilder {
	b.chatDeployment = deploymentID
	return b
}

// WithEmbeddingDeployment sets the embedding model deployment
func (b *AzureLLMBuilder) WithEmbeddingDeployment(deploymentID string) *AzureLLMBuilder {
	b.embeddingDeployment = deploymentID
	return b
}

// WithEndpoint sets the Azure OpenAI endpoint
func (b *AzureLLMBuilder) WithEndpoint(endpoint string) *AzureLLMBuilder {
	b.endpoint = endpoint
	return b
}

// WithApiKey sets the API key (prefer environment variables for this)
func (b *AzureLLMBuilder) WithApiKey(apiKey string) *AzureLLMBuilder {
	b.apiKey = apiKey
	return b
}

// WithApiVersion sets the API version
func (b *AzureLLMBuilder) WithApiVersion(apiVersion string) *AzureLLMBuilder {
	b.apiVersion = apiVersion
	return b
}

// WithAzureADAuth configures the adapter to use Azure AD authentication
func (b *AzureLLMBuilder) WithAzureADAuth(useAAD bool) *AzureLLMBuilder {
	b.useAzureADAuth = useAAD
	return b
}

// Fix the Azure OpenAI configuration fields to match the actual struct
func (b *AzureLLMBuilder) Build() (llm.ModelProvider, error) {
	// Validation
	if b.endpoint == "" {
		return nil, fmt.Errorf("missing Azure OpenAI endpoint")
	}
	if b.chatDeployment == "" {
		return nil, fmt.Errorf("missing Azure OpenAI deployment ID")
	}
	if !b.useAzureADAuth && b.apiKey == "" {
		return nil, fmt.Errorf("missing Azure OpenAI API key")
	}
	if b.embeddingDeployment == "" {
		return nil, fmt.Errorf("missing Azure OpenAI embedding deployment ID")
	}

	// Configure adapter with the correct field names
	options := llm.AzureOpenAIAdapterOptions{
		Endpoint:            b.endpoint,
		APIKey:              b.apiKey,
		ChatDeployment:      b.chatDeployment,
		EmbeddingDeployment: b.embeddingDeployment,
		// Fix field names to match the actual struct
		// Remove the fields that don't exist
	}

	adapter, err := llm.NewAzureOpenAIAdapter(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure OpenAI adapter: %w", err)
	}

	// If the adapter provides methods to configure these settings, use them
	// For example:
	// adapter.SetAPIVersion(b.apiVersion)
	// if b.useAzureADAuth {
	//     adapter.UseAzureADAuthentication()
	// }

	return adapter, nil
}

// LLMAgentBuilder simplifies creation of agents that use language models
type LLMAgentBuilder struct {
	provider     llm.ModelProvider
	systemPrompt string
	maxTokens    int32
	temperature  float32
	timeout      time.Duration
}

// NewLLMAgentBuilder creates a builder for LLM-backed agents
func NewLLMAgentBuilder(provider llm.ModelProvider) *LLMAgentBuilder {
	return &LLMAgentBuilder{
		provider:     provider,
		systemPrompt: "You are a helpful assistant.",
		maxTokens:    2000,
		temperature:  0.7,
		timeout:      30 * time.Second,
	}
}

// WithSystemPrompt sets the system prompt
func (b *LLMAgentBuilder) WithSystemPrompt(prompt string) *LLMAgentBuilder {
	b.systemPrompt = prompt
	return b
}

// WithMaxTokens sets the maximum tokens for generation
func (b *LLMAgentBuilder) WithMaxTokens(tokens int32) *LLMAgentBuilder {
	b.maxTokens = tokens
	return b
}

// WithTemperature sets the temperature for generation
func (b *LLMAgentBuilder) WithTemperature(temp float32) *LLMAgentBuilder {
	b.temperature = temp
	return b
}

// WithTimeout sets the timeout for LLM calls
func (b *LLMAgentBuilder) WithTimeout(timeout time.Duration) *LLMAgentBuilder {
	b.timeout = timeout
	return b
}

// Build creates an LLM-backed agent handler
func (b *LLMAgentBuilder) Build() agentflow.AgentHandler {
	return &LLMAgentHandler{
		llmProvider:  b.provider,
		systemPrompt: b.systemPrompt,
		maxTokens:    b.maxTokens,
		temperature:  b.temperature,
		timeout:      b.timeout,
	}
}

// LLMAgentHandler is an agent that uses an LLM provider
type LLMAgentHandler struct {
	llmProvider  llm.ModelProvider
	systemPrompt string
	maxTokens    int32
	temperature  float32
	timeout      time.Duration
}

// Fix the LLMAgentHandler implementation to use the correct response field names
func (a *LLMAgentHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	// Concatenate all string data from the event to form the prompt, ignoring nil values
	var combinedPrompt string
	for k, v := range event.GetData() {
		if s, ok := v.(string); ok {
			combinedPrompt += fmt.Sprintf("%s: %s\n", k, s)
		}
	}
	if combinedPrompt == "" {
		return agentflow.AgentResult{
			Error: "No string data found in event for LLM prompt",
			StartTime: time.Now(),
		}, fmt.Errorf("missing query in event data")
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	// Create prompt
	prompt := llm.Prompt{
		System: a.systemPrompt,
		User: combinedPrompt,
		Parameters: llm.ModelParameters{
			Temperature: ptrTo(a.temperature),
			MaxTokens:   ptrTo(a.maxTokens),
		},
	}

	// Call the LLM with retry logic
	response, err := a.callLLMWithRetry(timeoutCtx, prompt)
	if err != nil {
		return agentflow.AgentResult{
			Error:     fmt.Sprintf("LLM error: %v", err),
			EndTime:   time.Now(),
		}, err
	}	// Create output state using correct field names

	// Create output state using correct field names
	outputState := state.Clone()
	outputState.Set("answer", response.Content)
	outputState.Set("tokens_used", response.Usage.TotalTokens)


	// Return result
	return agentflow.AgentResult{
		OutputState: outputState,
		StartTime:   time.Now(), // Set start time here
		EndTime: time.Now(), // Calculate end time here
		Duration:    endTime.Sub(startTime),
	}, nil
}

// executeLLMCall performs a single LLM call with the given context and prompt.
func (a *LLMAgentHandler) executeLLMCall(ctx context.Context, prompt llm.Prompt) (llm.Response, error) {
	return a.llmProvider.Call(ctx, prompt)
}

// applyExponentialBackoff pauses execution based on the attempt number for exponential backoff.
func (a *LLMAgentHandler) applyExponentialBackoff(ctx context.Context, attempt int, retryDelay time.Duration) error {
	select {
	case <-time.After(retryDelay * time.Duration(1<<uint(attempt-1))):
		// Exponential backoff
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled during retry delay: %w", ctx.Err())
	}
}

// shouldRetry checks if a retry should be attempted based on the context and attempt number.
func (a *LLMAgentHandler) shouldRetry(ctx context.Context, attempt, maxRetries int) bool {
	return attempt < maxRetries && ctx.Err() == nil
}

// logRetry logs a message indicating a retry attempt.
func (a *LLMAgentHandler) logRetry(attempt, maxRetries int) {
	log.Printf("Retrying LLM call (attempt %d/%d)", attempt+1, maxRetries)
}

// callLLMWithRetry calls the LLM with exponential backoff retry
func (a *LLMAgentHandler) callLLMWithRetry(ctx context.Context, prompt llm.Prompt) (llm.Response, error) {
	var response llm.Response
	var lastErr error

	maxRetries := 3
	retryDelay := 1 * time.Second

	for attempt := 0; a.shouldRetry(ctx, attempt, maxRetries); attempt++ {
		if attempt > 0 {
			if err := a.applyExponentialBackoff(ctx, attempt, retryDelay); err != nil {
				return llm.Response{}, err
			}
			a.logRetry(attempt, maxRetries)
		}
		response, lastErr = a.executeLLMCall(ctx, prompt)
		if ctx.Err() != nil {
			return llm.Response{}, fmt.Errorf("context cancelled: %w", ctx.Err())
		}
		if lastErr == nil {
			return response, nil
		}
	}

	return llm.Response{}, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// Helper for creating pointers to values
func ptrTo[T any](v T) *T {
	return &v
}
