package v1beta

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// STREAMLINED BUILDER PATTERN (8 CORE METHODS)
// =============================================================================

// Builder defines the streamlined interface for building agents
// Reduced from 30+ methods to 8 core methods using functional options
type Builder interface {
	// Core configuration (4 methods)
	WithConfig(config *Config) Builder
	WithPreset(preset PresetType) Builder
	WithHandler(handler HandlerFunc) Builder
	Build() (Agent, error)

	// Convenience methods (5 methods)
	WithMemory(opts ...MemoryOption) Builder
	WithTools(opts ...ToolOption) Builder
	WithWorkflow(opts ...WorkflowOption) Builder
	WithSubWorkflow(opts ...BuilderSubWorkflowOption) Builder
	Clone() Builder
}

// PresetType defines common agent patterns for quick creation
type PresetType string

const (
	ChatAgent     PresetType = "chat"
	ResearchAgent PresetType = "research"
	DataAgent     PresetType = "data"
	WorkflowAgent PresetType = "workflow"
	// Note: SubWorkflow agents are created via WithSubWorkflow(), not as a preset
)

// HandlerFunc defines a simplified custom handler signature
type HandlerFunc func(ctx context.Context, input string, capabilities *Capabilities) (string, error)

// Capabilities provides access to LLM, Tools, and Memory for custom handlers
type Capabilities struct {
	LLM    func(system, user string) (string, error)
	Tools  ToolManager
	Memory Memory
}

// =============================================================================
// FUNCTIONAL OPTIONS FOR CONFIGURATION
// =============================================================================

// Option defines functional options for agent configuration
type Option func(*Config)

// WithLLM configures the LLM provider and model
func WithLLM(provider, model string) Option {
	return func(c *Config) {
		c.LLM.Provider = provider
		c.LLM.Model = model
	}
}

// WithLLMConfig configures the complete LLM settings
func WithLLMConfig(provider, model string, temperature float64, maxTokens int) Option {
	return func(c *Config) {
		c.LLM = LLMConfig{
			Provider:    provider,
			Model:       model,
			Temperature: float32(temperature),
			MaxTokens:   maxTokens,
		}
	}
}

// WithSystemPrompt sets the system prompt
func WithSystemPrompt(prompt string) Option {
	return func(c *Config) {
		c.SystemPrompt = prompt
	}
}

// WithAgentTimeout sets the execution timeout for the agent
func WithAgentTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithDebugMode enables or disables debug mode
func WithDebugMode(enabled bool) Option {
	return func(c *Config) {
		c.DebugMode = enabled
	}
}

// MemoryOption defines functional options for memory configuration
type MemoryOption func(*MemoryConfig)

// WithMemoryProvider sets the memory provider
func WithMemoryProvider(provider string) MemoryOption {
	return func(mc *MemoryConfig) {
		mc.Provider = provider
	}
}

// WithRAG enables RAG with specified configuration
func WithRAG(maxTokens int, personalWeight, knowledgeWeight float32) MemoryOption {
	return func(mc *MemoryConfig) {
		mc.RAG = &RAGConfig{
			MaxTokens:       maxTokens,
			PersonalWeight:  personalWeight,
			KnowledgeWeight: knowledgeWeight,
			HistoryLimit:    10,
		}
	}
}

// WithSessionScoped enables session-scoped memory
func WithSessionScoped() MemoryOption {
	return func(mc *MemoryConfig) {
		if mc.Options == nil {
			mc.Options = make(map[string]string)
		}
		mc.Options["session_scoped"] = "true"
	}
}

// WithContextAware enables context-aware memory
func WithContextAware() MemoryOption {
	return func(mc *MemoryConfig) {
		if mc.Options == nil {
			mc.Options = make(map[string]string)
		}
		mc.Options["context_aware"] = "true"
	}
}

// ToolOption defines functional options for tool configuration
type ToolOption func(*ToolsConfig)

// WithMCP enables MCP with specified servers
func WithMCP(servers ...MCPServer) ToolOption {
	return func(tc *ToolsConfig) {
		tc.Enabled = true

		// Initialize MCP config if not exists
		if tc.MCP == nil {
			tc.MCP = &MCPConfig{
				Enabled:           true,
				Discovery:         false, // Explicit servers only
				ConnectionTimeout: 30 * time.Second,
				MaxRetries:        3,
				RetryDelay:        1 * time.Second,
			}
		} else {
			tc.MCP.Enabled = true
		}

		// Add servers to MCP configuration
		tc.MCP.Servers = append(tc.MCP.Servers, servers...)
	}
}

// WithMCPDiscovery enables automatic MCP server discovery
func WithMCPDiscovery(scanPorts ...int) ToolOption {
	return func(tc *ToolsConfig) {
		tc.Enabled = true

		// Initialize MCP config if not exists
		if tc.MCP == nil {
			tc.MCP = &MCPConfig{
				Enabled:           true,
				Discovery:         true,
				ConnectionTimeout: 30 * time.Second,
				MaxRetries:        3,
				RetryDelay:        1 * time.Second,
				DiscoveryTimeout:  10 * time.Second,
			}
		} else {
			tc.MCP.Enabled = true
			tc.MCP.Discovery = true
			if tc.MCP.DiscoveryTimeout == 0 {
				tc.MCP.DiscoveryTimeout = 10 * time.Second
			}
		}

		// Set scan ports if provided, otherwise use defaults
		if len(scanPorts) > 0 {
			tc.MCP.ScanPorts = scanPorts
		} else if len(tc.MCP.ScanPorts) == 0 {
			// Default MCP ports
			tc.MCP.ScanPorts = []int{8080, 8081, 8090, 8100, 3000, 3001}
		}
	}
}

// WithToolTimeout sets the tool execution timeout
func WithToolTimeout(timeout time.Duration) ToolOption {
	return func(tc *ToolsConfig) {
		tc.Timeout = timeout
	}
}

// WithMaxConcurrentTools sets the maximum concurrent tool executions
func WithMaxConcurrentTools(max int) ToolOption {
	return func(tc *ToolsConfig) {
		tc.MaxConcurrent = max
	}
}

// WithToolCaching enables tool result caching
func WithToolCaching(ttl time.Duration) ToolOption {
	return func(tc *ToolsConfig) {
		if tc.Cache == nil {
			tc.Cache = &CacheConfig{
				Enabled: true,
				TTL:     ttl,
			}
		} else {
			tc.Cache.Enabled = true
			tc.Cache.TTL = ttl
		}
	}
}

// WorkflowOption defines functional options for workflow configuration
type WorkflowOption func(*WorkflowConfig)

// WithWorkflowMode sets the workflow execution mode
func WithWorkflowMode(mode string) WorkflowOption {
	return func(wc *WorkflowConfig) {
		wc.Mode = WorkflowMode(mode)
	}
}

// BuilderSubWorkflowOption defines functional options for SubWorkflow configuration in builder
type BuilderSubWorkflowOption func(*streamlinedBuilder)

// WithWorkflowInstance sets the workflow instance to wrap as an agent
func WithWorkflowInstance(workflow Workflow) BuilderSubWorkflowOption {
	return func(b *streamlinedBuilder) {
		b.workflowInstance = workflow
	}
}

// WithSubWorkflowMaxDepthBuilder sets the maximum nesting depth for SubWorkflow
func WithSubWorkflowMaxDepthBuilder(depth int) BuilderSubWorkflowOption {
	return func(b *streamlinedBuilder) {
		b.subworkflowMaxDepth = depth
	}
}

// WithSubWorkflowDescriptionBuilder sets a custom description for the SubWorkflow agent
func WithSubWorkflowDescriptionBuilder(description string) BuilderSubWorkflowOption {
	return func(b *streamlinedBuilder) {
		b.subworkflowDesc = description
	}
}

// WithWorkflowAgents sets the agents for workflow execution
func WithWorkflowAgents(agents ...string) WorkflowOption {
	return func(wc *WorkflowConfig) {
		wc.Agents = agents
	}
}

// WithMaxIterations sets the maximum workflow iterations
func WithMaxIterations(max int) WorkflowOption {
	return func(wc *WorkflowConfig) {
		wc.MaxIterations = max
	}
}

// =============================================================================
// STREAMLINED BUILDER IMPLEMENTATION
// =============================================================================

// streamlinedBuilder implements the Builder interface
type streamlinedBuilder struct {
	config  *Config
	handler HandlerFunc
	built   bool

	// SubWorkflow fields
	workflowInstance    Workflow
	subworkflowMaxDepth int
	subworkflowDesc     string
	isSubWorkflow       bool
}

// NewBuilder creates a new streamlined agent builder
func NewBuilder(name string) Builder {
	return &streamlinedBuilder{
		config: &Config{
			Name:         name,
			SystemPrompt: "You are a helpful assistant",
			Timeout:      30 * time.Second,
			LLM: LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   2048,
			},
		},
		built: false,
	}
}

// WithConfig sets the complete configuration
func (b *streamlinedBuilder) WithConfig(config *Config) Builder {
	if b.built {
		panic("Cannot modify frozen builder. Use Clone() first.")
	}
	b.config = config
	return b
}

// WithPreset applies a preset configuration
func (b *streamlinedBuilder) WithPreset(preset PresetType) Builder {
	if b.built {
		panic("Cannot modify frozen builder. Use Clone() first.")
	}

	switch preset {
	case ChatAgent:
		b.config.SystemPrompt = "You are a conversational assistant focused on providing helpful and friendly responses"
		b.config.LLM.Temperature = 0.8
		b.config.LLM.MaxTokens = 1024
		b.config.Memory = &MemoryConfig{
			Enabled:  true,
			Provider: "chromem",
			Options: map[string]string{
				"context_aware":  "true",
				"session_scoped": "true",
			},
		}

	case ResearchAgent:
		b.config.SystemPrompt = "You are a research assistant specialized in finding and summarizing information"
		b.config.LLM.Temperature = 0.3
		b.config.LLM.MaxTokens = 4096
		b.config.Timeout = 60 * time.Second
		b.config.Memory = &MemoryConfig{
			Enabled:  true,
			Provider: "chromem",
			Options: map[string]string{
				"context_aware":  "true",
				"session_scoped": "true",
			},
		}
		b.config.Tools = &ToolsConfig{
			Enabled:       true,
			MaxConcurrent: 10,
			Timeout:       60 * time.Second,
		}

	case DataAgent:
		b.config.SystemPrompt = "You are a data analysis assistant specialized in processing and analyzing data"
		b.config.LLM.Temperature = 0.1
		b.config.LLM.MaxTokens = 2048
		b.config.Timeout = 45 * time.Second
		b.config.Memory = &MemoryConfig{
			Enabled:  true,
			Provider: "chromem",
			Options: map[string]string{
				"context_aware": "false",
			},
		}
		b.config.Tools = &ToolsConfig{
			Enabled:       true,
			MaxConcurrent: 3,
			Timeout:       45 * time.Second,
		}

	case WorkflowAgent:
		b.config.SystemPrompt = "You are a workflow orchestration assistant"
		b.config.LLM.Temperature = 0.5
		b.config.LLM.MaxTokens = 2048
		b.config.Workflow = &WorkflowConfig{
			Mode:          WorkflowMode("sequential"),
			Timeout:       120 * time.Second,
			MaxIterations: 10,
		}
		b.config.Memory = &MemoryConfig{
			Enabled:  true,
			Provider: "chromem",
			Options: map[string]string{
				"context_aware":  "true",
				"session_scoped": "false",
			},
		}
	}

	return b
}

// WithHandler sets a custom handler function
func (b *streamlinedBuilder) WithHandler(handler HandlerFunc) Builder {
	if b.built {
		panic("Cannot modify frozen builder. Use Clone() first.")
	}
	b.handler = handler
	return b
}

// Build creates the final agent
func (b *streamlinedBuilder) Build() (Agent, error) {
	if b.built {
		return nil, fmt.Errorf("builder is frozen - use Clone() to create a new builder")
	}

	// Mark as built to prevent further modifications
	b.built = true

	// Check if this is a SubWorkflow agent
	if b.isSubWorkflow {
		if b.workflowInstance == nil {
			return nil, fmt.Errorf("SubWorkflow agent requires workflow instance - use WithWorkflowInstance()")
		}

		// Build SubWorkflow agent options
		var swOpts []SubWorkflowOption

		if b.subworkflowMaxDepth > 0 {
			swOpts = append(swOpts, WithSubWorkflowMaxDepth(b.subworkflowMaxDepth))
		}

		if b.subworkflowDesc != "" {
			swOpts = append(swOpts, WithSubWorkflowDescription(b.subworkflowDesc))
		}

		return NewSubWorkflowAgent(b.config.Name, b.workflowInstance, swOpts...), nil
	}

	// Validate configuration for non-SubWorkflow agents
	if err := b.validateConfig(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create and return the agent with REAL LLM integration
	// OLD: return newStreamlinedAgent(b.config, b.handler), nil
	return newRealAgent(b.config, b.handler)
}

// WithMemory configures memory with functional options
func (b *streamlinedBuilder) WithMemory(opts ...MemoryOption) Builder {
	if b.built {
		panic("Cannot modify frozen builder. Use Clone() first.")
	}

	// Initialize memory config if not exists
	if b.config.Memory == nil {
		b.config.Memory = &MemoryConfig{
			Enabled:  true,
			Provider: "chromem",
			Options: map[string]string{
				"context_aware":  "true",
				"session_scoped": "false",
			},
		}
	}

	// Apply functional options
	for _, opt := range opts {
		opt(b.config.Memory)
	}

	return b
}

// WithTools configures tools with functional options
func (b *streamlinedBuilder) WithTools(opts ...ToolOption) Builder {
	if b.built {
		panic("Cannot modify frozen builder. Use Clone() first.")
	}

	// Initialize tools config if not exists
	if b.config.Tools == nil {
		b.config.Tools = &ToolsConfig{
			Enabled:       true,
			MaxRetries:    3,
			Timeout:       30 * time.Second,
			MaxConcurrent: 5,
		}
	}

	// Apply functional options
	for _, opt := range opts {
		opt(b.config.Tools)
	}

	return b
}

// WithWorkflow configures workflow with functional options
func (b *streamlinedBuilder) WithWorkflow(opts ...WorkflowOption) Builder {
	if b.built {
		panic("Cannot modify frozen builder. Use Clone() first.")
	}

	// Initialize workflow config if not exists
	if b.config.Workflow == nil {
		b.config.Workflow = &WorkflowConfig{
			Mode:          "sequential",
			Timeout:       60 * time.Second,
			MaxIterations: 5,
		}
	}

	// Apply functional options
	for _, opt := range opts {
		opt(b.config.Workflow)
	}

	return b
}

// WithSubWorkflow configures a SubWorkflow agent (workflow wrapped as agent)
// This allows you to wrap any Workflow as an Agent for hierarchical composition.
//
// Example:
//
//	loop := vnext.NewLoopWorkflowWithCondition("revisions", agents, 3, vnext.OutputContains("APPROVED"))
//	agent, err := vnext.NewBuilder("revision-agent").
//	    WithSubWorkflow(
//	        vnext.WithWorkflowInstance(loop),
//	        vnext.WithSubWorkflowMaxDepthBuilder(5),
//	        vnext.WithSubWorkflowDescriptionBuilder("Writer-Editor revision loop"),
//	    ).
//	    Build()
func (b *streamlinedBuilder) WithSubWorkflow(opts ...BuilderSubWorkflowOption) Builder {
	if b.built {
		panic("Cannot modify frozen builder. Use Clone() first.")
	}

	// Mark as SubWorkflow agent
	b.isSubWorkflow = true
	b.subworkflowMaxDepth = 10 // Default

	// Apply functional options
	for _, opt := range opts {
		opt(b)
	}

	return b
}

// Clone creates a new builder with the same configuration
func (b *streamlinedBuilder) Clone() Builder {
	// Deep copy the configuration
	newConfig := &Config{
		Name:         b.config.Name,
		SystemPrompt: b.config.SystemPrompt,
		Timeout:      b.config.Timeout,
		DebugMode:    b.config.DebugMode,
		LLM:          b.config.LLM,
	}

	// Deep copy memory config if exists
	if b.config.Memory != nil {
		memConfig := *b.config.Memory
		if b.config.Memory.RAG != nil {
			ragConfig := *b.config.Memory.RAG
			memConfig.RAG = &ragConfig
		}
		newConfig.Memory = &memConfig
	}

	// Deep copy tools config if exists
	if b.config.Tools != nil {
		toolsConfig := *b.config.Tools
		newConfig.Tools = &toolsConfig
	}

	// Deep copy workflow config if exists
	if b.config.Workflow != nil {
		workflowConfig := *b.config.Workflow
		if b.config.Workflow.Agents != nil {
			workflowConfig.Agents = make([]string, len(b.config.Workflow.Agents))
			copy(workflowConfig.Agents, b.config.Workflow.Agents)
		}
		newConfig.Workflow = &workflowConfig
	}

	// Deep copy tracing config if exists
	if b.config.Tracing != nil {
		tracingConfig := *b.config.Tracing
		newConfig.Tracing = &tracingConfig
	}

	return &streamlinedBuilder{
		config:              newConfig,
		handler:             b.handler,
		built:               false, // New builder is not built yet
		workflowInstance:    b.workflowInstance,
		subworkflowMaxDepth: b.subworkflowMaxDepth,
		subworkflowDesc:     b.subworkflowDesc,
		isSubWorkflow:       b.isSubWorkflow,
	}
}

// validateConfig validates the builder configuration
func (b *streamlinedBuilder) validateConfig() error {
	if b.config.Name == "" {
		return fmt.Errorf("agent name is required")
	}

	if b.config.LLM.Provider == "" {
		return fmt.Errorf("LLM provider is required")
	}

	if b.config.LLM.Model == "" {
		return fmt.Errorf("LLM model is required")
	}

	if b.config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// =============================================================================
// FACTORY FUNCTIONS FOR QUICK AGENT CREATION
// =============================================================================

// NewChatAgent creates a chat agent with sensible defaults
func NewChatAgent(name string, opts ...Option) (Agent, error) {
	builder := NewBuilder(name).WithPreset(ChatAgent)

	// Apply additional options
	config := builder.(*streamlinedBuilder).config
	for _, opt := range opts {
		opt(config)
	}

	return builder.Build()
}

// NewResearchAgent creates a research agent with sensible defaults
func NewResearchAgent(name string, opts ...Option) (Agent, error) {
	builder := NewBuilder(name).WithPreset(ResearchAgent)

	// Apply additional options
	config := builder.(*streamlinedBuilder).config
	for _, opt := range opts {
		opt(config)
	}

	return builder.Build()
}

// NewDataAgent creates a data agent with sensible defaults
func NewDataAgent(name string, opts ...Option) (Agent, error) {
	builder := NewBuilder(name).WithPreset(DataAgent)

	// Apply additional options
	config := builder.(*streamlinedBuilder).config
	for _, opt := range opts {
		opt(config)
	}

	return builder.Build()
}

// NewWorkflowAgent creates a workflow agent with sensible defaults
func NewWorkflowAgent(name string, opts ...Option) (Agent, error) {
	builder := NewBuilder(name).WithPreset(WorkflowAgent)

	// Apply additional options
	config := builder.(*streamlinedBuilder).config
	for _, opt := range opts {
		opt(config)
	}

	return builder.Build()
}

// =============================================================================
// STREAMLINED AGENT IMPLEMENTATION
// =============================================================================

// streamlinedAgent implements the Agent interface using the streamlined configuration
type streamlinedAgent struct {
	config  *Config
	handler HandlerFunc
}

// newStreamlinedAgent creates a new streamlined agent
func newStreamlinedAgent(config *Config, handler HandlerFunc) Agent {
	return &streamlinedAgent{
		config:  config,
		handler: handler,
	}
}

// Name returns the agent name
func (a *streamlinedAgent) Name() string {
	return a.config.Name
}

// Run executes the agent with the given input
func (a *streamlinedAgent) Run(ctx context.Context, input string) (*Result, error) {
	startTime := time.Now()

	// If custom handler is provided, use it
	if a.handler != nil {
		capabilities := &Capabilities{
			LLM: func(system, user string) (string, error) {
				// This would call the actual LLM provider
				return fmt.Sprintf("LLM response to: %s", user), nil
			},
			// Tools and Memory would be initialized based on config
		}

		content, err := a.handler(ctx, input, capabilities)
		if err != nil {
			return &Result{
				Success:  false,
				Content:  "",
				Duration: time.Since(startTime),
				Error:    err.Error(),
			}, err
		}

		return &Result{
			Success:  true,
			Content:  content,
			Duration: time.Since(startTime),
		}, nil
	}

	// Default LLM execution
	content := fmt.Sprintf("Agent '%s' processed: %s", a.config.Name, input)

	return &Result{
		Success:  true,
		Content:  content,
		Duration: time.Since(startTime),
	}, nil
}

// RunWithOptions executes the agent with the given input and options
func (a *streamlinedAgent) RunWithOptions(ctx context.Context, input string, opts *RunOptions) (*Result, error) {
	// Apply timeout from options if provided
	if opts != nil && opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// For now, delegate to basic Run method
	// In a full implementation, this would handle all the options
	result, err := a.Run(ctx, input)

	// Apply options to result if needed
	if opts != nil && opts.DetailedResult && result != nil {
		// Add detailed information when requested
		result.Metadata = map[string]interface{}{
			"options_applied": true,
			"streaming":       opts.Streaming,
			"tools_requested": opts.Tools,
		}
	}

	return result, err
}

// Config returns the agent configuration
func (a *streamlinedAgent) Config() *Config {
	return a.config
}

// Capabilities returns the agent capabilities
func (a *streamlinedAgent) Capabilities() []string {
	capabilities := []string{"llm"}

	if a.config.Memory != nil {
		capabilities = append(capabilities, "memory")
		if a.config.Memory.RAG != nil {
			capabilities = append(capabilities, "rag")
		}
	}

	if a.config.Tools != nil && a.config.Tools.Enabled {
		capabilities = append(capabilities, "tools")
	}

	if a.config.Workflow != nil {
		capabilities = append(capabilities, "workflow")
	}

	if a.handler != nil {
		capabilities = append(capabilities, "custom_handler")
	}

	return capabilities
}

// Memory returns the memory provider (nil for streamlined agent)
func (a *streamlinedAgent) Memory() Memory {
	return nil
}

// RunStream executes the agent with streaming output
func (a *streamlinedAgent) RunStream(ctx context.Context, input string, opts ...StreamOption) (Stream, error) {
	return a.RunStreamWithOptions(ctx, input, nil, opts...)
}

// RunStreamWithOptions executes the agent with streaming output and run options
func (a *streamlinedAgent) RunStreamWithOptions(ctx context.Context, input string, runOpts *RunOptions, streamOpts ...StreamOption) (Stream, error) {
	// Apply timeout from run options if provided
	if runOpts != nil && runOpts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, runOpts.Timeout)
		defer cancel()
	}

	// Create stream
	metadata := &StreamMetadata{
		AgentName: a.config.Name,
		StartTime: time.Now(),
		Model:     a.config.LLM.Model,
		Extra:     make(map[string]interface{}),
	}

	if runOpts != nil {
		metadata.SessionID = runOpts.SessionID
		// Merge stream options from run options if provided
		if runOpts.StreamOptions != nil {
			opts := []StreamOption{
				WithBufferSize(runOpts.StreamOptions.BufferSize),
			}
			if runOpts.StreamOptions.IncludeThoughts {
				opts = append(opts, WithThoughts())
			}
			if runOpts.StreamOptions.IncludeToolCalls {
				opts = append(opts, WithToolCalls())
			}
			streamOpts = append(opts, streamOpts...)
		}
		if runOpts.StreamHandler != nil {
			streamOpts = append(streamOpts, WithStreamHandler(runOpts.StreamHandler))
		}
	}

	stream, writer := NewStream(ctx, metadata, streamOpts...)

	// Start streaming execution in goroutine
	go func() {
		defer writer.Close()
		startTime := time.Now()

		// Emit thinking chunk
		writer.Write(&StreamChunk{
			Type:    ChunkTypeThought,
			Content: fmt.Sprintf("Processing input: %s", input),
		})

		// Simulate LLM streaming response
		// In real implementation, this would stream from the actual LLM
		response := fmt.Sprintf("Agent '%s' processed: %s", a.config.Name, input)

		// Stream response word by word
		words := []rune(response)
		for i := 0; i < len(words); i += 5 {
			end := i + 5
			if end > len(words) {
				end = len(words)
			}

			chunk := &StreamChunk{
				Type:  ChunkTypeDelta,
				Delta: string(words[i:end]),
			}

			if err := writer.Write(chunk); err != nil {
				writer.CloseWithError(err)
				return
			}

			// Simulate typing delay
			time.Sleep(20 * time.Millisecond)
		}

		// Emit done chunk
		writer.Write(&StreamChunk{
			Type: ChunkTypeDone,
		})

		// Set final result
		result := &Result{
			Success:  true,
			Content:  response,
			Duration: time.Since(startTime),
			Metadata: map[string]interface{}{
				"streamed": true,
			},
		}

		if s, ok := stream.(*basicStream); ok {
			s.SetResult(result)
		}
	}()

	return stream, nil
}

// Initialize initializes the agent
func (a *streamlinedAgent) Initialize(ctx context.Context) error {
	// Initialize components based on configuration
	// This would set up LLM, memory, tools, etc.
	return nil
}

// Cleanup cleans up agent resources
func (a *streamlinedAgent) Cleanup(ctx context.Context) error {
	// Clean up resources
	return nil
}

// =============================================================================
// EXAMPLE USAGE AND DOCUMENTATION
// =============================================================================

/*
Example usage of the streamlined builder pattern:

Basic builder usage:
	agent, err := NewBuilder("my-agent").
		WithPreset(ChatAgent).
		Build()
	if err != nil {
		log.Fatal(err)
	}

Custom LLM configuration:
	agent, err := NewBuilder("research-agent").
		WithConfig(&Config{
			SystemPrompt: "You are a research assistant",
			LLM: LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.3,
				MaxTokens:   4096,
			},
		}).
		Build()

Using functional options:
	agent, err := NewBuilder("my-agent").
		WithLLM("ollama", "llama2").
		WithSystemPrompt("You are a helpful coding assistant").
		WithAgentTimeout(60 * time.Second).
		Build()

Memory-enabled agent:
	agent, err := NewBuilder("memory-agent").
		WithMemory(
			WithMemoryProvider("memory"),
			WithSessionScoped(),
			WithContextAware(),
			WithRAG(2048, 0.3, 0.7),
		).
		Build()

Tools-enabled agent:
	agent, err := NewBuilder("tool-agent").
		WithTools(
			WithToolTimeout(30 * time.Second),
			WithMaxConcurrentTools(5),
			WithToolCaching(5 * time.Minute),
		).
		Build()

MCP server integration:
	mcpServers := []MCPServer{
		{Name: "filesystem", Type: "stdio", Command: "mcp-server-fs"},
		{Name: "web", Type: "http_sse", Address: "localhost", Port: 8080},
	}

	agent, err := NewBuilder("mcp-agent").
		WithTools(WithMCP(mcpServers...)).
		Build()

Workflow agent:
	agent, err := NewBuilder("workflow-agent").
		WithWorkflow(
			WithWorkflowMode("sequential"),
			WithWorkflowAgents("agent1", "agent2", "agent3"),
			WithMaxIterations(10),
		).
		Build()

Using presets:
	// Chat agent preset
	chatAgent, err := NewChatAgent("chat-bot",
		WithLLM("openai", "gpt-4"),
		WithDebugMode(true),
	)

	// Research agent preset
	researchAgent, err := NewResearchAgent("researcher",
		WithLLM("openai", "gpt-4"),
		WithAgentTimeout(120 * time.Second),
	)

	// Data agent preset
	dataAgent, err := NewDataAgent("analyzer",
		WithLLM("openai", "gpt-4"),
	)

	// Workflow agent preset
	workflowAgent, err := NewWorkflowAgent("orchestrator",
		WithLLM("openai", "gpt-4"),
	)

Custom handler agent:
	handler := func(ctx context.Context, input string, caps *Capabilities) (string, error) {
		// Custom logic here
		if strings.Contains(input, "weather") {
			// Use tools
			result, err := caps.Tools.Execute(ctx, "get_weather", map[string]interface{}{
				"location": "New York",
			})
			if err != nil {
				return "", err
			}
			return result.Content.(string), nil
		}

		// Fall back to LLM
		return caps.LLM("You are a helpful assistant", input)
	}

	agent, err := NewBuilder("custom-agent").
		WithHandler(handler).
		Build()

Complete configuration example:
	agent, err := NewBuilder("advanced-agent").
		WithConfig(&Config{
			Name:         "advanced-agent",
			SystemPrompt: "You are an advanced AI assistant",
			Timeout:      60 * time.Second,
			DebugMode:    true,
			LLM: LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4-turbo",
				Temperature: 0.7,
				MaxTokens:   4096,
			},
			Memory: &MemoryConfig{
				Provider: "memory",
				RAG: &RAGConfig{
					MaxTokens:       2048,
					PersonalWeight:  0.3,
					KnowledgeWeight: 0.7,
					HistoryLimit:    10,
				},
			},
			Tools: &ToolsConfig{
				Enabled:       true,
				MaxRetries:    3,
				Timeout:       30 * time.Second,
				MaxConcurrent: 5,
				Cache: &CacheConfig{
					Enabled: true,
					TTL:     5 * time.Minute,
				},
			},
			Tracing: &TracingConfig{
				Enabled: true,
				Level:   "enhanced",
				WebUI:   true,
			},
		}).
		Build()

Builder cloning for variations:
	baseBuilder := NewBuilder("base-agent").
		WithLLM("openai", "gpt-4").
		WithMemory(WithMemoryProvider("memory"))

	// Create variations from the base
	agent1, _ := baseBuilder.Clone().
		WithSystemPrompt("You are agent 1").
		Build()

	agent2, _ := baseBuilder.Clone().
		WithSystemPrompt("You are agent 2").
		Build()

Chaining options:
	agent, err := NewBuilder("chained-agent").
		WithLLM("openai", "gpt-4").
		WithSystemPrompt("You are a helpful assistant").
		WithAgentTimeout(30 * time.Second).
		WithDebugMode(true).
		WithMemory(
			WithMemoryProvider("memory"),
			WithSessionScoped(),
		).
		WithTools(
			WithToolTimeout(20 * time.Second),
			WithMaxConcurrentTools(3),
		).
		Build()

SubWorkflow agent - wrapping workflows as agents:
	// Create individual agents
	writer, _ := NewChatAgent("writer", WithLLM("openai", "gpt-4"))
	editor, _ := NewChatAgent("editor", WithLLM("openai", "gpt-4"))
	publisher, _ := NewChatAgent("publisher", WithLLM("openai", "gpt-4"))

	// Create a conditional loop workflow
	revisionLoop := NewLoopWorkflowWithCondition(
		"revision_loop",
		[]Agent{writer, editor},
		3, // max iterations
		OutputContains("APPROVED"), // exit when editor approves
	)

	// Wrap loop as an agent using builder
	loopAgent, err := NewBuilder("revision-agent").
		WithSubWorkflow(
			WithWorkflowInstance(revisionLoop),
			WithSubWorkflowMaxDepthBuilder(5),
			WithSubWorkflowDescriptionBuilder("Writer-Editor revision loop"),
		).
		Build()

	// Use in a larger workflow
	mainPipeline := NewSequentialWorkflow("story_pipeline", []Agent{
		loopAgent,   // SubWorkflow agent!
		publisher,
	})

SubWorkflow with complex nesting:
	// Level 1: Parallel analysis
	analysisWorkflow := NewParallelWorkflow("analysis", []Agent{
		sentimentAgent,
		summaryAgent,
		keywordsAgent,
	})

	// Level 2: Wrap as agent
	analysisAgent, _ := NewBuilder("analysis-agent").
		WithSubWorkflow(WithWorkflowInstance(analysisWorkflow)).
		Build()

	// Level 3: Sequential with loop
	reviewLoop := NewLoopWorkflowWithCondition(
		"review_loop",
		[]Agent{analysisAgent, reviewAgent},
		5,
		Convergence(0.95),
	)

	// Level 4: Final pipeline
	mainWorkflow := NewSequentialWorkflow("main", []Agent{
		preprocessAgent,
		NewSubWorkflowAgent("review", reviewLoop),
		outputAgent,
	})

SubWorkflow quick creation (without builder):
	// Direct creation for simple cases
	loop := NewLoopWorkflow("task_loop", []Agent{agent1, agent2}, 3)
	loopAgent := NewSubWorkflowAgent("loop-agent", loop)

	// Or using builder for more control
	loopAgent2, _ := NewBuilder("loop-agent").
		WithSubWorkflow(
			WithWorkflowInstance(loop),
			WithSubWorkflowMaxDepthBuilder(10),
		).
		Build()
*/
