package v1beta

import (
	"context"
	"fmt"
	"time"
)

// Agent represents a unified vNext agent with enhanced capabilities
// This consolidates all agent interfaces into a single, clean interface
type Agent interface {
	// Core execution methods
	Name() string
	Run(ctx context.Context, input string) (*Result, error)
	RunWithOptions(ctx context.Context, input string, opts *RunOptions) (*Result, error)

	// Streaming execution methods
	RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error)
	RunStreamWithOptions(ctx context.Context, input string, runOpts *RunOptions, streamOpts ...StreamOption) (Stream, error)

	// Configuration access
	Config() *Config
	Capabilities() []string

	// Lifecycle methods
	Initialize(ctx context.Context) error
	Cleanup(ctx context.Context) error
}

// NOTE: Core configuration types (Config, LLMConfig, TracingConfig, MemoryConfig,
// ToolsConfig, WorkflowConfig, RAGConfig) are now defined in config.go.
// This provides a centralized location for all configuration types.

// AgentMemoryConfig defines the basic agent memory configuration
type AgentMemoryConfig struct {
	Enabled       bool   `toml:"enabled"`
	Provider      string `toml:"provider"`
	ContextAware  bool   `toml:"context_aware"`
	SessionScoped bool   `toml:"session_scoped"`
}

// Result represents a unified result from agent execution
// This consolidates AgentResult, DetailedResult, and related types into a single type
type Result struct {
	// Core result data
	Success   bool                   `json:"success"`
	Content   string                 `json:"content"`
	Duration  time.Duration          `json:"duration"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	// Execution details (consolidated from DetailedResult)
	TokensUsed    int      `json:"tokens_used,omitempty"`
	ToolsCalled   []string `json:"tools_called,omitempty"`
	MemoryUsed    bool     `json:"memory_used,omitempty"`
	MemoryQueries int      `json:"memory_queries,omitempty"`

	// New fields for multimodal output
	Images      []ImageData  `json:"images,omitempty"`
	Audio       []AudioData  `json:"audio,omitempty"`
	Video       []VideoData  `json:"video,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`

	// Advanced execution details (from DetailedResult)
	ToolExecutions  []ToolExecution  `json:"tool_executions,omitempty"`
	LLMInteractions []LLMInteraction `json:"llm_interactions,omitempty"`

	// Legacy compatibility fields (from core.AgentResult)
	StartTime time.Time           `json:"start_time,omitempty"`
	EndTime   time.Time           `json:"end_time,omitempty"`
	Error     string              `json:"error,omitempty"`
	ToolCalls []ToolCall          `json:"tool_calls,omitempty"`
	Sources   []SourceAttribution `json:"sources,omitempty"`
}

// ToolExecution represents detailed information about a tool execution
type ToolExecution struct {
	Name       string        `json:"name"`
	Duration   time.Duration `json:"duration"`
	Success    bool          `json:"success"`
	InputSize  int           `json:"input_size"`
	OutputSize int           `json:"output_size"`
	Error      string        `json:"error,omitempty"`
}

// LLMInteraction represents detailed information about an LLM interaction
type LLMInteraction struct {
	Provider       string        `json:"provider"`
	Model          string        `json:"model"`
	PromptTokens   int           `json:"prompt_tokens"`
	ResponseTokens int           `json:"response_tokens"`
	Duration       time.Duration `json:"duration"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
}

// ToolCall represents a tool call made during execution (legacy compatibility)
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Result    interface{}            `json:"result,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Success   bool                   `json:"success,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// SourceAttribution represents a source of information used in the response (legacy compatibility)
type SourceAttribution struct {
	Source    string    `json:"source"`
	Content   string    `json:"content"`
	Score     float64   `json:"score,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// Text extracts the text content from the result, providing compatibility with core.AgentResult
func (r *Result) Text() string {
	return r.Content
}

// IsSuccess returns true if the execution was successful
func (r *Result) IsSuccess() bool {
	return r.Success && r.Error == ""
}

// GetDuration returns the execution duration
func (r *Result) GetDuration() time.Duration {
	return r.Duration
}

// =============================================================================
// RUNOPTIONS FACTORY FUNCTIONS
// =============================================================================

// NewRunOptions creates a new RunOptions with default values
func NewRunOptions() *RunOptions {
	return &RunOptions{
		ToolMode:       "auto",
		MaxRetries:     3,
		DetailedResult: false,
		Context:        make(map[string]interface{}),
	}
}

// RunWithTools creates RunOptions with specific tools enabled
func RunWithTools(tools ...string) *RunOptions {
	return &RunOptions{
		Tools:          tools,
		ToolMode:       "specific",
		MaxRetries:     3,
		DetailedResult: false,
		Context:        make(map[string]interface{}),
	}
}

// RunWithMemory creates RunOptions with memory configuration
func RunWithMemory(sessionID string, opts *MemoryOptions) *RunOptions {
	return &RunOptions{
		Memory:         opts,
		SessionID:      sessionID,
		ToolMode:       "auto",
		MaxRetries:     3,
		DetailedResult: false,
		Context:        make(map[string]interface{}),
	}
}

// RunWithStreaming creates RunOptions with streaming enabled
func RunWithStreaming() *RunOptions {
	return &RunOptions{
		Streaming:      true,
		ToolMode:       "auto",
		MaxRetries:     3,
		DetailedResult: false,
		Context:        make(map[string]interface{}),
	}
}

// RunWithDetailedResult creates RunOptions that returns detailed execution information
func RunWithDetailedResult() *RunOptions {
	return &RunOptions{
		DetailedResult: true,
		IncludeTrace:   true,
		IncludeSources: true,
		ToolMode:       "auto",
		MaxRetries:     3,
		Context:        make(map[string]interface{}),
	}
}

// RunWithTimeout creates RunOptions with a specific timeout
func RunWithTimeout(timeout time.Duration) *RunOptions {
	return &RunOptions{
		Timeout:        timeout,
		ToolMode:       "auto",
		MaxRetries:     3,
		DetailedResult: false,
		Context:        make(map[string]interface{}),
	}
}

// Chainable methods for RunOptions

// SetTools sets the tools to use for execution
func (opts *RunOptions) SetTools(tools ...string) *RunOptions {
	opts.Tools = tools
	if len(tools) > 0 {
		opts.ToolMode = "specific"
	}
	return opts
}

// SetMemory sets the memory configuration
func (opts *RunOptions) SetMemory(sessionID string, memOpts *MemoryOptions) *RunOptions {
	opts.Memory = memOpts
	opts.SessionID = sessionID
	return opts
}

// SetStreaming enables or disables streaming (deprecated: use RunStream)
func (opts *RunOptions) SetStreaming(enabled bool) *RunOptions {
	opts.Streaming = enabled
	return opts
}

// WithStream adds streaming configuration to RunOptions
func (opts *RunOptions) WithStream(streamOpts ...StreamOption) *RunOptions {
	if opts.StreamOptions == nil {
		opts.StreamOptions = DefaultStreamOptions()
	}
	for _, opt := range streamOpts {
		opt(opts.StreamOptions)
	}
	opts.Streaming = true
	return opts
}

// WithStreamHandler sets a streaming callback handler
func (opts *RunOptions) WithStreamHandler(handler StreamHandler) *RunOptions {
	opts.StreamHandler = handler
	opts.Streaming = true
	return opts
}

// SetTimeout sets the execution timeout
func (opts *RunOptions) SetTimeout(timeout time.Duration) *RunOptions {
	opts.Timeout = timeout
	return opts
}

// SetDetailedResult enables or disables detailed result information
func (opts *RunOptions) SetDetailedResult(enabled bool) *RunOptions {
	opts.DetailedResult = enabled
	return opts
}

// SetTracing configures tracing for this execution
func (opts *RunOptions) SetTracing(enabled bool, level string) *RunOptions {
	opts.TraceEnabled = enabled
	opts.TraceLevel = level
	return opts
}

// AddContext adds a key-value pair to the execution context
func (opts *RunOptions) AddContext(key string, value interface{}) *RunOptions {
	if opts.Context == nil {
		opts.Context = make(map[string]interface{})
	}
	opts.Context[key] = value
	return opts
}

// =============================================================================
// TOOL AND MEMORY INTERFACES (FOR STREAMLINED BUILDER)
// =============================================================================

// ToolManager defines the interface for tool management and execution
// This will be consolidated from MCP and other tool systems
// Extended types (ToolMetrics, MCPServerInfo, etc.) are defined in tools.go
type ToolManager interface {
	// Tool operations
	Execute(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error)
	List() []ToolInfo

	// Tool discovery and management
	Available() []string
	IsAvailable(name string) bool

	// MCP integration (implemented in tools.go)
	ConnectMCP(ctx context.Context, servers ...MCPServer) error
	DisconnectMCP(serverName string) error
	DiscoverMCP(ctx context.Context) ([]MCPServerInfo, error)

	// Health and monitoring (types defined in tools.go)
	HealthCheck(ctx context.Context) map[string]MCPHealthStatus
	GetMetrics() ToolMetrics

	// Lifecycle
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// ToolInfo provides information about an available tool
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Category    string                 `json:"category"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success bool        `json:"success"`
	Content interface{} `json:"content"`
	Error   string      `json:"error,omitempty"`
}

// Memory defines the interface for memory operations
// This will be consolidated from various memory providers
type Memory interface {
	// Basic operations
	Store(ctx context.Context, content string, opts ...StoreOption) error
	Query(ctx context.Context, query string, opts ...QueryOption) ([]MemoryResult, error)

	// Session management
	NewSession() string
	SetSession(ctx context.Context, sessionID string) context.Context

	// RAG operations (if RAG is configured)
	IngestDocument(ctx context.Context, doc Document) error
	BuildContext(ctx context.Context, query string, opts ...ContextOption) (*RAGContext, error)
}

// StoreOption defines options for storing content in memory
type StoreOption func(*StoreConfig)

// QueryOption defines options for querying memory
type QueryOption func(*QueryConfig)

// ContextOption defines options for building RAG context
type ContextOption func(*ContextConfig)

// StoreConfig defines configuration for memory storage operations
type StoreConfig struct {
	ContentType string
	Source      string
	Metadata    map[string]interface{}
}

// QueryConfig defines configuration for memory query operations
type QueryConfig struct {
	Limit           int
	ScoreThreshold  float32
	IncludeMetadata bool
}

// ContextConfig defines configuration for RAG context building
type ContextConfig struct {
	MaxTokens       int
	PersonalWeight  float32
	KnowledgeWeight float32
}

// MemoryResult represents a result from memory query
type MemoryResult struct {
	Content   string                 `json:"content"`
	Score     float32                `json:"score"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// Document represents a document for RAG ingestion
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Title    string                 `json:"title"`
	Source   string                 `json:"source"`
	Metadata map[string]interface{} `json:"metadata"`
}

// RAGContext represents the context built for RAG operations
type RAGContext struct {
	PersonalMemory    []MemoryResult `json:"personal_memory"`
	KnowledgeBase     []MemoryResult `json:"knowledge_base"`
	ChatHistory       []string       `json:"chat_history"`
	TotalTokens       int            `json:"total_tokens"`
	SourceAttribution []string       `json:"source_attribution"`
}

// NOTE: MCPServer is now defined in config.go

// MemoryEntry represents an entry in memory (referenced by HandlerCapabilities)
type MemoryEntry struct {
	Content   string                 `json:"content"`
	Score     float32                `json:"score"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// Tool represents a tool interface (referenced by HandlerCapabilities)
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

// RunOptions defines flexible options for agent execution
// This provides comprehensive configuration for different execution scenarios
type RunOptions struct {
	// Tool configuration
	Tools    []string `json:"tools,omitempty"`     // Specific tools to use
	ToolMode string   `json:"tool_mode,omitempty"` // "auto", "specific", "none"

	// Memory configuration
	Memory    *MemoryOptions `json:"memory,omitempty"`     // Memory settings for this run
	SessionID string         `json:"session_id,omitempty"` // Session identifier

	// Execution configuration
	Streaming     bool                   `json:"streaming,omitempty"`      // Enable streaming response (deprecated: use RunStream)
	StreamOptions *StreamOptions         `json:"stream_options,omitempty"` // Streaming configuration
	StreamHandler StreamHandler          `json:"-"`                        // Stream handler callback
	Timeout       time.Duration          `json:"timeout,omitempty"`        // Execution timeout
	Context       map[string]interface{} `json:"context,omitempty"`        // Additional context

	// Retry and error handling
	MaxRetries int `json:"max_retries,omitempty"` // Maximum retry attempts

	// Tracing configuration
	TraceEnabled bool   `json:"trace_enabled,omitempty"` // Enable tracing for this run
	TraceLevel   string `json:"trace_level,omitempty"`   // Tracing detail level

	// Result configuration
	DetailedResult bool `json:"detailed_result,omitempty"` // Return detailed execution information
	IncludeTrace   bool `json:"include_trace,omitempty"`   // Include trace data in result
	IncludeSources bool `json:"include_sources,omitempty"` // Include source attributions

	// Performance configuration
	MaxTokens   int      `json:"max_tokens,omitempty"`  // Override max tokens for this run
	Temperature *float64 `json:"temperature,omitempty"` // Override temperature for this run

	// Multimodal input
	Images []ImageData `json:"images,omitempty"` // Images to include in the prompt
	Audio  []AudioData `json:"audio,omitempty"`  // Audio to include in the prompt
	Video  []VideoData `json:"video,omitempty"`  // Video to include in the prompt

	// Middleware configuration
	SkipMiddleware []string `json:"skip_middleware,omitempty"` // Skip specific middleware by name
}

// SessionInfo holds information about an agent session
type SessionInfo struct {
	SessionID   string                 `json:"session_id"`
	StartTime   time.Time              `json:"start_time"`
	RunCount    int                    `json:"run_count"`
	Context     map[string]interface{} `json:"context"`
	ActiveTools []string               `json:"active_tools"`
}

// AgentMetrics holds runtime metrics for an agent
type AgentMetrics struct {
	TotalRuns        int64         `json:"total_runs"`
	SuccessfulRuns   int64         `json:"successful_runs"`
	FailedRuns       int64         `json:"failed_runs"`
	AverageRunTime   time.Duration `json:"average_run_time"`
	TotalTokensUsed  int64         `json:"total_tokens_used"`
	ToolsInvoked     int64         `json:"tools_invoked"`
	MemoryOperations int64         `json:"memory_operations"`
	LastRunTime      time.Time     `json:"last_run_time"`
	SessionsActive   int           `json:"sessions_active"`
}

// AgentMiddleware defines the interface for agent middleware
// Middleware can intercept and modify the agent execution flow
type AgentMiddleware interface {
	// BeforeRun is called before the agent execution
	// It can modify the context and input, or cancel the execution
	BeforeRun(ctx context.Context, input string) (context.Context, string, error)

	// AfterRun is called after the agent execution
	// It can modify the result or handle errors
	AfterRun(ctx context.Context, input string, result *Result, err error) (*Result, error)
}

// AgentError represents a structured error with additional context for vNext agents
type AgentError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	InnerError error                  `json:"inner_error,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Error implements the error interface for AgentError
func (e *AgentError) Error() string {
	if e.InnerError != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.InnerError)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAgentError creates a new AgentError with the specified code and message
func NewAgentError(code ErrorCode, message string) *AgentError {
	return &AgentError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}
}

// NewAgentErrorWithDetails creates a new AgentError with the specified code, message, and details
func NewAgentErrorWithDetails(code ErrorCode, message string, details map[string]interface{}) *AgentError {
	return &AgentError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// NewAgentErrorWithError creates a new AgentError with the specified code, message, and inner error
func NewAgentErrorWithError(code ErrorCode, message string, err error) *AgentError {
	return &AgentError{
		Code:       code,
		Message:    message,
		InnerError: err,
		Timestamp:  time.Now(),
		Details:    make(map[string]interface{}),
	}
}

// AddDetail adds a key-value pair to the error details
func (e *AgentError) AddDetail(key string, value interface{}) *AgentError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithStackTrace adds a stack trace to the error
func (e *AgentError) WithStackTrace(stackTrace string) *AgentError {
	e.StackTrace = stackTrace
	return e
}

// IsErrorCode checks if the error has the specified error code
func (e *AgentError) IsErrorCode(code ErrorCode) bool {
	return e.Code == code
}

// Unwrap returns the inner error for error unwrapping
func (e *AgentError) Unwrap() error {
	return e.InnerError
}

// ErrorCode represents a standardized error code for vNext agents
type ErrorCode string

// Common error codes for different failure scenarios
const (
	// Configuration errors
	ErrConfigInvalid ErrorCode = "CONFIG_INVALID"
	ErrConfigMissing ErrorCode = "CONFIG_MISSING"

	// Agent lifecycle errors
	ErrAgentInitFailed    ErrorCode = "AGENT_INIT_FAILED"
	ErrAgentCleanupFailed ErrorCode = "AGENT_CLEANUP_FAILED"

	// LLM provider errors
	ErrLLMProviderNotFound ErrorCode = "LLM_PROVIDER_NOT_FOUND"
	ErrLLMCallFailed       ErrorCode = "LLM_CALL_FAILED"
	ErrLLMTimeout          ErrorCode = "LLM_TIMEOUT"

	// Memory errors
	ErrMemoryStoreFailed    ErrorCode = "MEMORY_STORE_FAILED"
	ErrMemoryRetrieveFailed ErrorCode = "MEMORY_RETRIEVE_FAILED"
	ErrMemoryClearFailed    ErrorCode = "MEMORY_CLEAR_FAILED"

	// Tool errors
	ErrToolNotFound         ErrorCode = "TOOL_NOT_FOUND"
	ErrToolExecutionFailed  ErrorCode = "TOOL_EXECUTION_FAILED"
	ErrToolValidationFailed ErrorCode = "TOOL_VALIDATION_FAILED"

	// Middleware errors
	ErrMiddlewareBeforeRun ErrorCode = "MIDDLEWARE_BEFORE_RUN_FAILED"
	ErrMiddlewareAfterRun  ErrorCode = "MIDDLEWARE_AFTER_RUN_FAILED"

	// Workflow errors
	ErrWorkflowInvalid       ErrorCode = "WORKFLOW_INVALID"
	ErrWorkflowNodeNotFound  ErrorCode = "WORKFLOW_NODE_NOT_FOUND"
	ErrWorkflowCycleDetected ErrorCode = "WORKFLOW_CYCLE_DETECTED"

	// Context errors
	ErrContextCancelled ErrorCode = "CONTEXT_CANCELLED"
	ErrContextTimeout   ErrorCode = "CONTEXT_TIMEOUT"

	// General errors
	ErrInternal       ErrorCode = "INTERNAL_ERROR"
	ErrNotImplemented ErrorCode = "NOT_IMPLEMENTED"
)

// Grouped Configuration Options

// AgentConfigOptions defines grouped configuration options for agents
type AgentConfigOptions struct {
	Name         string
	SystemPrompt string
	Timeout      time.Duration
	Model        *LLMOptions
	Memory       *MemoryOptions
	Tools        *ToolOptions
	Tracing      *TracingOptions
}

// MemoryOptions defines memory-related configuration with RAG capabilities
type MemoryOptions struct {
	Enabled       bool
	Provider      string
	ContextAware  bool
	SessionScoped bool
	RAGConfig     *RAGConfig
}

// NOTE: RAGConfig is now defined in config.go

// ToolOptions defines tool-related configuration with orchestration capabilities
type ToolOptions struct {
	Enabled             bool
	MaxRetries          int
	Timeout             time.Duration
	RateLimit           int // requests per second
	CacheEnabled        bool
	CacheTTL            time.Duration
	MaxConcurrent       int
	OrchestrationConfig *ToolOrchestrationConfig
}

// ToolOrchestrationConfig defines configuration for tool orchestration
type ToolOrchestrationConfig struct {
	AutoDiscovery      bool
	MaxRetries         int
	Timeout            time.Duration
	ParallelExecution  bool
	DependencyTracking bool
}

// LLMOptions defines LLM-related configuration
type LLMOptions struct {
	Provider    string
	Model       string
	Temperature float64
	MaxTokens   int
	Timeout     time.Duration
}

// TracingOptions defines tracing-related configuration
type TracingOptions struct {
	Enabled     bool
	Level       string // "none", "basic", "enhanced", "debug"
	WebUI       bool
	Performance bool
	MemoryTrace bool
	ToolTrace   bool
}

// =============================================================================
// CUSTOM HANDLER SUPPORT
// =============================================================================

// CustomHandlerFunc defines a function that can process queries with custom logic
// It receives the original query and a helper function to call LLM if needed
// This is useful for simple custom logic with LLM fallback
type CustomHandlerFunc func(ctx context.Context, query string, llmCall func(systemPrompt, userPrompt string) (string, error)) (string, error)

// EnhancedHandlerFunc defines an enhanced function that can process queries with full access to agent capabilities
// This provides access to LLM, tools, and memory through capability bridges
// This is useful for complex workflows that require multiple capabilities
type EnhancedHandlerFunc func(ctx context.Context, query string, capabilities *HandlerCapabilities) (string, error)

// HandlerCapabilities provides access to all agent capabilities for enhanced handlers
type HandlerCapabilities struct {
	// LLM access - call the configured LLM provider
	LLMCall func(systemPrompt, userPrompt string) (string, error)

	// Tool access - execute tools by name
	ToolCall func(toolName string, args map[string]interface{}) (*ToolResult, error)

	// Memory access - store and query memory
	MemoryStore func(content string, contentType, source string) error
	MemoryQuery func(query string, limit int) ([]*MemoryEntry, error)

	// Tool discovery - get available tools
	GetAvailableTools func() ([]Tool, error)

	// Helper for formatting tools for LLM prompts
	FormatToolsForPrompt func() string
}

// =============================================================================
// EXAMPLE USAGE AND DOCUMENTATION
// =============================================================================

/*
Example usage of the unified agent system:

Basic agent execution:
	agent := NewAgent("my-agent", config)
	result, err := agent.Run(ctx, "Hello, world!")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Content)

Agent execution with options:
	opts := NewRunOptions().
		SetTools("calculator", "web_search").
		SetTimeout(30 * time.Second).
		SetDetailedResult(true)

	result, err := agent.RunWithOptions(ctx, "Calculate 2+2 and search for weather", opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Result: %s\n", result.Content)
	fmt.Printf("Tools used: %v\n", result.ToolsCalled)
	fmt.Printf("Duration: %v\n", result.Duration)

Memory-enabled execution:
	memOpts := &MemoryOptions{
		Enabled:       true,
		Provider:      "memory",
		ContextAware:  true,
		SessionScoped: true,
	}

	opts := WithMemory("session-123", memOpts).
		SetDetailedResult(true)

	result, err := agent.RunWithOptions(ctx, "Remember my name is John", opts)

Streaming execution:
	opts := RunWithStreaming()
	result, err := agent.RunWithOptions(ctx, "Tell me a story", opts)

Tool-specific execution:
	opts := RunWithTools("calculator", "web_search").
		SetTimeout(60 * time.Second)

	result, err := agent.RunWithOptions(ctx, "Calculate the square root of 144 and find current news", opts)
*/
