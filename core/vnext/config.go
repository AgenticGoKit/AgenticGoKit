package vnext

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// =============================================================================
// UNIFIED CONFIGURATION SYSTEM
// =============================================================================

// Config defines the unified configuration for an agent
// This consolidates AgentConfig, ProjectConfig, and related types
type Config struct {
	// Core settings
	Name         string        `toml:"name"`
	SystemPrompt string        `toml:"system_prompt"`
	Timeout      time.Duration `toml:"timeout"`
	DebugMode    bool          `toml:"debug_mode"`

	// LLM configuration
	LLM LLMConfig `toml:"llm"`

	// Feature configurations
	Memory    *MemoryConfig    `toml:"memory,omitempty"`
	Tools     *ToolsConfig     `toml:"tools,omitempty"`
	Workflow  *WorkflowConfig  `toml:"workflow,omitempty"`
	Tracing   *TracingConfig   `toml:"tracing,omitempty"`
	Streaming *StreamingConfig `toml:"streaming,omitempty"`
}

// LLMConfig contains LLM provider configuration
type LLMConfig struct {
	Provider    string  `toml:"provider"`           // openai, ollama, azure, anthropic
	Model       string  `toml:"model"`              // Model name
	Temperature float32 `toml:"temperature"`        // 0.0 to 2.0
	MaxTokens   int     `toml:"max_tokens"`         // Maximum tokens to generate
	BaseURL     string  `toml:"base_url,omitempty"` // Custom base URL
	APIKey      string  `toml:"api_key,omitempty"`  // API key (prefer env vars)
}

// MemoryConfig contains memory and RAG configuration
type MemoryConfig struct {
	Provider   string            `toml:"provider"`   // "memory", "pgvector", "weaviate"
	Connection string            `toml:"connection"` // Connection string
	RAG        *RAGConfig        `toml:"rag,omitempty"`
	Options    map[string]string `toml:"options,omitempty"`
}

// RAGConfig contains RAG-specific settings
type RAGConfig struct {
	MaxTokens       int     `toml:"max_tokens"`
	PersonalWeight  float32 `toml:"personal_weight"`
	KnowledgeWeight float32 `toml:"knowledge_weight"`
	HistoryLimit    int     `toml:"history_limit"`
}

// ToolsConfig contains tool management configuration
type ToolsConfig struct {
	Enabled        bool                  `toml:"enabled"`
	MaxRetries     int                   `toml:"max_retries"`
	Timeout        time.Duration         `toml:"timeout"`
	RateLimit      int                   `toml:"rate_limit"`     // requests per second
	MaxConcurrent  int                   `toml:"max_concurrent"` // max parallel executions
	MCP            *MCPConfig            `toml:"mcp,omitempty"`
	Cache          *CacheConfig          `toml:"cache,omitempty"`
	CircuitBreaker *CircuitBreakerConfig `toml:"circuit_breaker,omitempty"`
}

// MCPConfig contains MCP server configuration
type MCPConfig struct {
	Enabled           bool          `toml:"enabled"`
	Discovery         bool          `toml:"discovery"`
	Servers           []MCPServer   `toml:"servers"`
	Cache             *CacheConfig  `toml:"cache,omitempty"`
	ConnectionTimeout time.Duration `toml:"connection_timeout"`
	MaxRetries        int           `toml:"max_retries"`
	RetryDelay        time.Duration `toml:"retry_delay"`
	DiscoveryTimeout  time.Duration `toml:"discovery_timeout"`
	ScanPorts         []int         `toml:"scan_ports,omitempty"`
}

// MCPServer defines an individual MCP server
type MCPServer struct {
	Name    string `toml:"name" json:"name"`
	Type    string `toml:"type" json:"type"`       // "tcp", "stdio", "websocket", "http_sse", "http_streaming"
	Address string `toml:"address" json:"address"` // Host or connection string
	Port    int    `toml:"port,omitempty" json:"port,omitempty"`
	Command string `toml:"command,omitempty" json:"command,omitempty"` // For stdio
	Enabled bool   `toml:"enabled" json:"enabled"`
}

// CacheConfig contains caching configuration
type CacheConfig struct {
	Enabled         bool                     `toml:"enabled"`
	TTL             time.Duration            `toml:"ttl"`
	MaxSize         int64                    `toml:"max_size_mb"` // Max size in MB
	MaxKeys         int                      `toml:"max_keys"`
	EvictionPolicy  string                   `toml:"eviction_policy"` // "lru", "lfu", "ttl"
	CleanupInterval time.Duration            `toml:"cleanup_interval"`
	ToolTTLs        map[string]time.Duration `toml:"tool_ttls,omitempty"` // Per-tool TTL overrides
	Backend         string                   `toml:"backend"`             // "memory", "redis", "file"
	BackendConfig   map[string]string        `toml:"backend_config,omitempty"`
}

// CircuitBreakerConfig defines circuit breaker settings for tool execution
type CircuitBreakerConfig struct {
	Enabled          bool          `toml:"enabled"`
	FailureThreshold int           `toml:"failure_threshold"` // Failures before opening
	SuccessThreshold int           `toml:"success_threshold"` // Successes before closing
	Timeout          time.Duration `toml:"timeout"`           // How long circuit stays open
	HalfOpenMaxCalls int           `toml:"half_open_max_calls"`
}

// WorkflowConfig contains workflow orchestration settings
type WorkflowConfig struct {
	Mode          WorkflowMode  `toml:"mode"`
	Agents        []string      `toml:"agents"`
	Timeout       time.Duration `toml:"timeout"`
	MaxIterations int           `toml:"max_iterations"`
	Memory        *MemoryConfig `toml:"memory,omitempty"`
}

// WorkflowMode defines workflow execution modes
type WorkflowMode string

const (
	Sequential WorkflowMode = "sequential"
	Parallel   WorkflowMode = "parallel"
	DAG        WorkflowMode = "dag"
	Loop       WorkflowMode = "loop"
)

// TracingConfig contains tracing configuration
type TracingConfig struct {
	Enabled bool   `toml:"enabled"`
	Level   string `toml:"level"` // none, basic, enhanced, debug
}

// StreamingConfig contains streaming configuration
type StreamingConfig struct {
	Enabled          bool `toml:"enabled"`
	BufferSize       int  `toml:"buffer_size"`        // Size of the streaming buffer (default: 100)
	FlushInterval    int  `toml:"flush_interval_ms"`  // Milliseconds between flushes (default: 100)
	Timeout          int  `toml:"timeout_ms"`         // Milliseconds before timeout (default: 30000)
	IncludeThoughts  bool `toml:"include_thoughts"`   // Include thought process chunks
	IncludeToolCalls bool `toml:"include_tool_calls"` // Include tool call chunks
	IncludeMetadata  bool `toml:"include_metadata"`   // Include metadata chunks
	TextOnly         bool `toml:"text_only"`          // Only stream text chunks (excludes thoughts, tools, metadata)
}

// ProjectConfig represents a complete multi-agent project configuration
type ProjectConfig struct {
	Project     ProjectInfo                   `toml:"project"`
	Logging     LoggingConfig                 `toml:"logging"`
	LLM         LLMConfig                     `toml:"llm"` // Global LLM fallback
	Providers   ProvidersConfig               `toml:"providers"`
	Memory      MemoryConfig                  `toml:"memory"` // Global memory fallback
	MCP         MCPConfig                     `toml:"mcp"`
	Workflow    WorkflowConfig                `toml:"workflow"`
	Agents      map[string]ProjectAgentConfig `toml:"agents"` // Multiple agents
	Middleware  MiddlewareConfig              `toml:"middleware"`
	Development DevelopmentConfig             `toml:"development"`
}

// ProjectInfo contains basic project metadata
type ProjectInfo struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
	Author      string `toml:"author"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `toml:"level"`       // debug, info, warn, error
	Format     string `toml:"format"`      // text, json
	FileOutput bool   `toml:"file_output"` // Enable file logging
}

// ProjectAgentConfig extends Config for project-level agent definitions
type ProjectAgentConfig struct {
	Config                // Embed unified Config
	Role           string `toml:"role"`            // Agent role in project
	Description    string `toml:"description"`     // Agent description
	Enabled        bool   `toml:"enabled"`         // Enable/disable agent
	TimeoutSeconds int    `toml:"timeout_seconds"` // Timeout in seconds for TOML parsing
}

// ProvidersConfig contains provider-specific configurations
type ProvidersConfig struct {
	Ollama    OllamaConfig    `toml:"ollama"`
	OpenAI    OpenAIConfig    `toml:"openai"`
	Azure     AzureConfig     `toml:"azure"`
	Anthropic AnthropicConfig `toml:"anthropic"`
}

// OllamaConfig contains Ollama-specific configuration
type OllamaConfig struct {
	BaseURL         string   `toml:"base_url"`
	ConnectTimeout  int      `toml:"connect_timeout"`
	ReadTimeout     int      `toml:"read_timeout"`
	AvailableModels []string `toml:"available_models"`
}

// OpenAIConfig contains OpenAI-specific configuration
type OpenAIConfig struct {
	APIKeyEnv       string   `toml:"api_key_env"`
	BaseURL         string   `toml:"base_url"`
	AvailableModels []string `toml:"available_models"`
}

// AzureConfig contains Azure OpenAI-specific configuration
type AzureConfig struct {
	APIKeyEnv     string `toml:"api_key_env"`
	EndpointEnv   string `toml:"endpoint_env"`
	DeploymentEnv string `toml:"deployment_env"`
	APIVersion    string `toml:"api_version"`
}

// AnthropicConfig contains Anthropic-specific configuration
type AnthropicConfig struct {
	APIKeyEnv       string   `toml:"api_key_env"`
	BaseURL         string   `toml:"base_url"`
	AvailableModels []string `toml:"available_models"`
}

// MiddlewareConfig contains middleware configuration
type MiddlewareConfig struct {
	Logging LoggingMiddleware `toml:"logging"`
	Metrics MetricsMiddleware `toml:"metrics"`
	Tracing TracingMiddleware `toml:"tracing"`
}

// LoggingMiddleware contains logging middleware configuration
type LoggingMiddleware struct {
	Enabled bool   `toml:"enabled"`
	Level   string `toml:"level"`
}

// MetricsMiddleware contains metrics middleware configuration
type MetricsMiddleware struct {
	Enabled bool `toml:"enabled"`
}

// TracingMiddleware contains tracing middleware configuration
type TracingMiddleware struct {
	Enabled bool `toml:"enabled"`
}

// DevelopmentConfig contains development and debugging configuration
type DevelopmentConfig struct {
	DebugMode bool `toml:"debug_mode"` // Enable debug mode for testing
	HotReload bool `toml:"hot_reload"` // Enable hot reload
}

// =============================================================================
// FUNCTIONAL OPTIONS PATTERN
// =============================================================================

// NOTE: Functional options (Option, MemoryOption, ToolOption, WorkflowOption)
// and their constructor functions (WithLLM, WithMemory, WithTools, etc.) are
// defined in builder.go as part of the streamlined builder pattern.
// This separation keeps configuration loading/validation in config.go and
// builder functionality in builder.go.

// =============================================================================
// CONFIGURATION LOADING AND VALIDATION
// =============================================================================

// LoadConfigFromTOML loads configuration from a TOML file
func LoadConfigFromTOML(filePath string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", filePath)
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", filePath, err)
	}

	// Parse TOML
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML configuration: %w", err)
	}

	// Apply environment variable resolution
	resolver := NewConfigResolver(&config)
	resolvedConfig := resolver.ResolveConfig()

	// Validate configuration
	if err := ValidateConfig(resolvedConfig); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return resolvedConfig, nil
}

// LoadProjectConfigFromTOML loads a complete project configuration from TOML
func LoadProjectConfigFromTOML(filePath string) (*ProjectConfig, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", filePath)
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", filePath, err)
	}

	// Parse TOML
	var config ProjectConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML configuration: %w", err)
	}

	// Apply environment variable resolution
	resolver := NewProjectConfigResolver(&config)
	resolvedConfig := resolver.ResolveConfig()

	// Validate configuration
	if errors := ValidateProjectConfig(resolvedConfig); len(errors) > 0 {
		if HasCriticalErrors(errors) {
			return nil, fmt.Errorf("critical configuration errors: %s", FormatValidationErrors(errors))
		}
	}

	return resolvedConfig, nil
}

// NewConfig creates a new configuration with defaults
func NewConfig(name string, opts ...Option) *Config {
	config := &Config{
		Name:    name,
		Timeout: 30 * time.Second, // Default timeout
		LLM: LLMConfig{
			Temperature: 0.7,  // Default temperature
			MaxTokens:   1000, // Default max tokens
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	return config
}

// =============================================================================
// CONFIGURATION RESOLUTION (Environment Variables)
// =============================================================================

// ConfigResolver resolves configuration with environment variable substitution
type ConfigResolver struct {
	config *Config
}

// NewConfigResolver creates a new configuration resolver
func NewConfigResolver(config *Config) *ConfigResolver {
	return &ConfigResolver{
		config: config,
	}
}

// ResolveConfig resolves configuration values with environment variables
func (cr *ConfigResolver) ResolveConfig() *Config {
	// Create a copy of the config to avoid modifying the original
	resolvedConfig := *cr.config

	// Resolve string fields that might contain environment variables
	resolvedConfig.Name = cr.resolveString(resolvedConfig.Name)
	resolvedConfig.SystemPrompt = cr.resolveString(resolvedConfig.SystemPrompt)
	resolvedConfig.LLM.Provider = cr.resolveString(resolvedConfig.LLM.Provider)
	resolvedConfig.LLM.Model = cr.resolveString(resolvedConfig.LLM.Model)
	resolvedConfig.LLM.BaseURL = cr.resolveString(resolvedConfig.LLM.BaseURL)
	resolvedConfig.LLM.APIKey = cr.resolveString(resolvedConfig.LLM.APIKey)

	// Resolve memory configuration if present
	if resolvedConfig.Memory != nil {
		resolvedConfig.Memory.Provider = cr.resolveString(resolvedConfig.Memory.Provider)
		resolvedConfig.Memory.Connection = cr.resolveString(resolvedConfig.Memory.Connection)
	}

	return &resolvedConfig
}

// ProjectConfigResolver resolves project configuration with environment variable substitution
type ProjectConfigResolver struct {
	config *ProjectConfig
}

// NewProjectConfigResolver creates a new project configuration resolver
func NewProjectConfigResolver(config *ProjectConfig) *ProjectConfigResolver {
	return &ProjectConfigResolver{
		config: config,
	}
}

// ResolveConfig resolves project configuration values with environment variables
func (pcr *ProjectConfigResolver) ResolveConfig() *ProjectConfig {
	// Create a copy of the config to avoid modifying the original
	resolvedConfig := *pcr.config

	// Resolve project info
	resolvedConfig.Project.Name = pcr.resolveString(resolvedConfig.Project.Name)
	resolvedConfig.Project.Version = pcr.resolveString(resolvedConfig.Project.Version)
	resolvedConfig.Project.Description = pcr.resolveString(resolvedConfig.Project.Description)
	resolvedConfig.Project.Author = pcr.resolveString(resolvedConfig.Project.Author)

	// Resolve global LLM configuration
	resolvedConfig.LLM.Provider = pcr.resolveString(resolvedConfig.LLM.Provider)
	resolvedConfig.LLM.Model = pcr.resolveString(resolvedConfig.LLM.Model)
	resolvedConfig.LLM.BaseURL = pcr.resolveString(resolvedConfig.LLM.BaseURL)
	resolvedConfig.LLM.APIKey = pcr.resolveString(resolvedConfig.LLM.APIKey)

	// Resolve global memory configuration
	resolvedConfig.Memory.Provider = pcr.resolveString(resolvedConfig.Memory.Provider)
	resolvedConfig.Memory.Connection = pcr.resolveString(resolvedConfig.Memory.Connection)

	// Resolve agent configurations
	for name, agent := range resolvedConfig.Agents {
		agent.Name = pcr.resolveString(agent.Name)
		agent.SystemPrompt = pcr.resolveString(agent.SystemPrompt)
		agent.Role = pcr.resolveString(agent.Role)
		agent.Description = pcr.resolveString(agent.Description)

		// Resolve agent-specific LLM config if present
		agent.LLM.Provider = pcr.resolveString(agent.LLM.Provider)
		agent.LLM.Model = pcr.resolveString(agent.LLM.Model)
		agent.LLM.BaseURL = pcr.resolveString(agent.LLM.BaseURL)
		agent.LLM.APIKey = pcr.resolveString(agent.LLM.APIKey)

		// Resolve agent-specific memory config if present
		if agent.Memory != nil {
			agent.Memory.Provider = pcr.resolveString(agent.Memory.Provider)
			agent.Memory.Connection = pcr.resolveString(agent.Memory.Connection)
		}

		resolvedConfig.Agents[name] = agent
	}

	return &resolvedConfig
}

// resolveString resolves environment variables in a string
func (cr *ConfigResolver) resolveString(value string) string {
	return resolveEnvironmentVariables(value)
}

// resolveString resolves environment variables in a string for project config
func (pcr *ProjectConfigResolver) resolveString(value string) string {
	return resolveEnvironmentVariables(value)
}

// resolveEnvironmentVariables is a shared function for environment variable resolution
func resolveEnvironmentVariables(value string) string {
	// Match ${VAR_NAME} or ${VAR_NAME:default_value} patterns
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	return re.ReplaceAllStringFunc(value, func(match string) string {
		// Extract the variable name and default value
		varName := match[2 : len(match)-1] // Remove ${ and }
		defaultValue := ""

		// Check if there's a default value
		if parts := strings.Split(varName, ":"); len(parts) > 1 {
			varName = parts[0]
			defaultValue = strings.Join(parts[1:], ":") // In case the default value contains ":"
		}

		// Get the environment variable value
		if envValue, exists := os.LookupEnv(varName); exists {
			return envValue
		}

		// Return default value if provided, otherwise return the original match
		if defaultValue != "" {
			return defaultValue
		}

		return match // Return original if no env var and no default
	})
}

// =============================================================================
// CONFIGURATION VALIDATION
// =============================================================================

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
	Severity   string `json:"severity"` // critical, high, medium, low
}

// ValidateConfig validates a configuration with detailed errors
func ValidateConfig(config *Config) error {
	var errors []ValidationError

	// Validate core settings
	if config.Name == "" {
		errors = append(errors, ValidationError{
			Field:      "name",
			Message:    "Agent name is required",
			Suggestion: "Add a name field to your configuration",
			Severity:   "critical",
		})
	}

	// Validate LLM configuration
	if config.LLM.Provider == "" {
		errors = append(errors, ValidationError{
			Field:      "llm.provider",
			Message:    "LLM provider is required",
			Suggestion: "Set llm.provider to one of: openai, ollama, azure, anthropic",
			Severity:   "critical",
		})
	}

	if config.LLM.Model == "" {
		errors = append(errors, ValidationError{
			Field:      "llm.model",
			Message:    "LLM model is required",
			Suggestion: "Set llm.model to a valid model name for your provider",
			Severity:   "critical",
		})
	}

	// Validate LLM provider-specific settings
	if err := validateLLMProvider(config.LLM); err != nil {
		errors = append(errors, *err)
	}

	// Validate memory configuration if present
	if config.Memory != nil {
		if errs := validateMemoryConfig(config.Memory); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	// Validate tools configuration if present
	if config.Tools != nil {
		if errs := validateToolsConfig(config.Tools); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	// Validate workflow configuration if present
	if config.Workflow != nil {
		if errs := validateWorkflowConfig(config.Workflow); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	// Return error if critical issues found
	if HasCriticalErrors(errors) {
		return fmt.Errorf("critical configuration errors: %s", FormatValidationErrors(errors))
	}

	return nil
}

// validateLLMProvider validates provider-specific LLM configuration
func validateLLMProvider(llm LLMConfig) *ValidationError {
	validProviders := []string{"openai", "ollama", "azure", "anthropic"}
	isValid := false
	for _, provider := range validProviders {
		if llm.Provider == provider {
			isValid = true
			break
		}
	}

	if !isValid {
		return &ValidationError{
			Field:      "llm.provider",
			Message:    fmt.Sprintf("Invalid LLM provider: %s", llm.Provider),
			Suggestion: fmt.Sprintf("Use one of: %s", strings.Join(validProviders, ", ")),
			Severity:   "critical",
		}
	}

	// Validate temperature range
	if llm.Temperature < 0.0 || llm.Temperature > 2.0 {
		return &ValidationError{
			Field:      "llm.temperature",
			Message:    "Temperature must be between 0.0 and 2.0",
			Suggestion: "Set temperature to a value between 0.0 (deterministic) and 2.0 (creative)",
			Severity:   "high",
		}
	}

	// Validate max tokens
	if llm.MaxTokens <= 0 {
		return &ValidationError{
			Field:      "llm.max_tokens",
			Message:    "Max tokens must be greater than 0",
			Suggestion: "Set max_tokens to a positive integer (e.g., 1000, 4000)",
			Severity:   "high",
		}
	}

	return nil
}

// validateMemoryConfig validates memory configuration
func validateMemoryConfig(memory *MemoryConfig) []ValidationError {
	var errors []ValidationError

	if memory.Provider == "" {
		errors = append(errors, ValidationError{
			Field:      "memory.provider",
			Message:    "Memory provider is required",
			Suggestion: "Set memory.provider to one of: memory, pgvector, weaviate",
			Severity:   "critical",
		})
	}

	// Validate RAG configuration if present
	if memory.RAG != nil {
		if memory.RAG.MaxTokens <= 0 {
			errors = append(errors, ValidationError{
				Field:      "memory.rag.max_tokens",
				Message:    "RAG max tokens must be greater than 0",
				Suggestion: "Set a positive value for max_tokens (e.g., 2000, 4000)",
				Severity:   "high",
			})
		}

		if memory.RAG.PersonalWeight < 0 || memory.RAG.PersonalWeight > 1 {
			errors = append(errors, ValidationError{
				Field:      "memory.rag.personal_weight",
				Message:    "Personal weight must be between 0 and 1",
				Suggestion: "Set personal_weight to a value between 0.0 and 1.0",
				Severity:   "medium",
			})
		}

		if memory.RAG.KnowledgeWeight < 0 || memory.RAG.KnowledgeWeight > 1 {
			errors = append(errors, ValidationError{
				Field:      "memory.rag.knowledge_weight",
				Message:    "Knowledge weight must be between 0 and 1",
				Suggestion: "Set knowledge_weight to a value between 0.0 and 1.0",
				Severity:   "medium",
			})
		}
	}

	return errors
}

// validateToolsConfig validates tools configuration
func validateToolsConfig(tools *ToolsConfig) []ValidationError {
	var errors []ValidationError

	if tools.MaxRetries < 0 {
		errors = append(errors, ValidationError{
			Field:      "tools.max_retries",
			Message:    "Max retries cannot be negative",
			Suggestion: "Set max_retries to 0 or a positive integer",
			Severity:   "medium",
		})
	}

	// Validate MCP configuration if present
	if tools.MCP != nil {
		if errs := validateMCPConfig(tools.MCP); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	return errors
}

// validateMCPConfig validates MCP configuration
func validateMCPConfig(mcp *MCPConfig) []ValidationError {
	var errors []ValidationError

	for i, server := range mcp.Servers {
		if server.Name == "" {
			errors = append(errors, ValidationError{
				Field:      fmt.Sprintf("tools.mcp.servers[%d].name", i),
				Message:    "MCP server name is required",
				Suggestion: "Provide a unique name for each MCP server",
				Severity:   "critical",
			})
		}

		if server.Type == "" {
			errors = append(errors, ValidationError{
				Field:      fmt.Sprintf("tools.mcp.servers[%d].type", i),
				Message:    "MCP server type is required",
				Suggestion: "Set type to one of: tcp, stdio, websocket",
				Severity:   "critical",
			})
		}

		validTypes := []string{"tcp", "stdio", "websocket", "http_sse", "http_streaming"}
		isValidType := false
		for _, validType := range validTypes {
			if server.Type == validType {
				isValidType = true
				break
			}
		}

		if !isValidType {
			errors = append(errors, ValidationError{
				Field:      fmt.Sprintf("tools.mcp.servers[%d].type", i),
				Message:    fmt.Sprintf("Invalid MCP server type: %s", server.Type),
				Suggestion: fmt.Sprintf("Use one of: %s", strings.Join(validTypes, ", ")),
				Severity:   "critical",
			})
		}

		// Validate type-specific requirements
		switch server.Type {
		case "tcp", "websocket":
			if server.Address == "" {
				errors = append(errors, ValidationError{
					Field:      fmt.Sprintf("tools.mcp.servers[%d].address", i),
					Message:    fmt.Sprintf("%s server requires an address", server.Type),
					Suggestion: "Provide a valid host address (e.g., localhost, 192.168.1.1)",
					Severity:   "critical",
				})
			}
			if server.Port <= 0 || server.Port > 65535 {
				errors = append(errors, ValidationError{
					Field:      fmt.Sprintf("tools.mcp.servers[%d].port", i),
					Message:    fmt.Sprintf("%s server requires a valid port (1-65535)", server.Type),
					Suggestion: "Provide a valid port number",
					Severity:   "critical",
				})
			}

		case "stdio":
			if server.Command == "" {
				errors = append(errors, ValidationError{
					Field:      fmt.Sprintf("tools.mcp.servers[%d].command", i),
					Message:    "STDIO server requires a command",
					Suggestion: "Provide the command to execute (e.g., 'python mcp_server.py')",
					Severity:   "critical",
				})
			}

		case "http_sse", "http_streaming":
			if server.Address == "" {
				errors = append(errors, ValidationError{
					Field:      fmt.Sprintf("tools.mcp.servers[%d].address", i),
					Message:    fmt.Sprintf("%s server requires an address/endpoint", server.Type),
					Suggestion: "Provide full URL (e.g., 'http://localhost:8080/mcp')",
					Severity:   "critical",
				})
			}
		}
	}

	return errors
}

// validateWorkflowConfig validates workflow configuration
func validateWorkflowConfig(workflow *WorkflowConfig) []ValidationError {
	var errors []ValidationError

	validModes := []WorkflowMode{Sequential, Parallel, DAG, Loop}
	isValidMode := false
	for _, mode := range validModes {
		if workflow.Mode == mode {
			isValidMode = true
			break
		}
	}

	if !isValidMode {
		errors = append(errors, ValidationError{
			Field:      "workflow.mode",
			Message:    fmt.Sprintf("Invalid workflow mode: %s", workflow.Mode),
			Suggestion: "Use one of: sequential, parallel, dag, loop",
			Severity:   "critical",
		})
	}

	if workflow.MaxIterations < 0 {
		errors = append(errors, ValidationError{
			Field:      "workflow.max_iterations",
			Message:    "Max iterations cannot be negative",
			Suggestion: "Set max_iterations to 0 (unlimited) or a positive integer",
			Severity:   "medium",
		})
	}

	return errors
}

// ValidateProjectConfig validates a project configuration
func ValidateProjectConfig(config *ProjectConfig) []ValidationError {
	var errors []ValidationError

	// Validate project info
	if config.Project.Name == "" {
		errors = append(errors, ValidationError{
			Field:      "project.name",
			Message:    "Project name is required",
			Suggestion: "Add a name field to your project configuration",
			Severity:   "critical",
		})
	}

	// Validate each agent configuration
	for name, agent := range config.Agents {
		if agent.Name == "" {
			agent.Name = name // Use key as default name
		}

		// Validate agent configuration
		if err := ValidateConfig(&agent.Config); err != nil {
			errors = append(errors, ValidationError{
				Field:      fmt.Sprintf("agents.%s", name),
				Message:    fmt.Sprintf("Agent configuration invalid: %v", err),
				Suggestion: "Fix the agent configuration errors",
				Severity:   "critical",
			})
		}
	}

	return errors
}

// HasCriticalErrors checks if any validation errors are critical
func HasCriticalErrors(errors []ValidationError) bool {
	for _, err := range errors {
		if err.Severity == "critical" {
			return true
		}
	}
	return false
}

// FormatValidationErrors formats validation errors for display
func FormatValidationErrors(errors []ValidationError) string {
	var formatted []string
	for _, err := range errors {
		formatted = append(formatted, fmt.Sprintf("[%s] %s: %s (%s)",
			err.Severity, err.Field, err.Message, err.Suggestion))
	}
	return strings.Join(formatted, "; ")
}

// =============================================================================
// DEFAULT CONFIGURATIONS
// =============================================================================

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig(name string) *Config {
	return &Config{
		Name:      name,
		Timeout:   30 * time.Second,
		DebugMode: false,
		LLM: LLMConfig{
			Provider:    "ollama",
			Model:       "llama3.2",
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	}
}

// DefaultProjectConfig returns a project configuration with sensible defaults
func DefaultProjectConfig(name string) *ProjectConfig {
	return &ProjectConfig{
		Project: ProjectInfo{
			Name:        name,
			Version:     "1.0.0",
			Description: "AgenticGoKit project",
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			FileOutput: false,
		},
		LLM: LLMConfig{
			Provider:    "ollama",
			Model:       "llama3.2",
			Temperature: 0.7,
			MaxTokens:   1000,
		},
		Memory: MemoryConfig{
			Provider: "memory",
		},
		Workflow: WorkflowConfig{
			Mode:          Sequential,
			Timeout:       5 * time.Minute,
			MaxIterations: 10,
		},
		Agents: make(map[string]ProjectAgentConfig),
		Development: DevelopmentConfig{
			DebugMode: false,
			HotReload: false,
		},
	}
}

// =============================================================================
// CONFIGURATION HELPERS
// =============================================================================

// MergeConfigs merges multiple configurations, with later configs taking precedence
func MergeConfigs(configs ...*Config) *Config {
	if len(configs) == 0 {
		return DefaultConfig("merged")
	}

	result := *configs[0] // Copy first config

	for i := 1; i < len(configs); i++ {
		config := configs[i]

		// Merge non-zero values
		if config.Name != "" {
			result.Name = config.Name
		}
		if config.SystemPrompt != "" {
			result.SystemPrompt = config.SystemPrompt
		}
		if config.Timeout != 0 {
			result.Timeout = config.Timeout
		}

		// Merge LLM config
		if config.LLM.Provider != "" {
			result.LLM.Provider = config.LLM.Provider
		}
		if config.LLM.Model != "" {
			result.LLM.Model = config.LLM.Model
		}
		if config.LLM.Temperature != 0 {
			result.LLM.Temperature = config.LLM.Temperature
		}
		if config.LLM.MaxTokens != 0 {
			result.LLM.MaxTokens = config.LLM.MaxTokens
		}

		// Replace feature configs if present
		if config.Memory != nil {
			result.Memory = config.Memory
		}
		if config.Tools != nil {
			result.Tools = config.Tools
		}
		if config.Workflow != nil {
			result.Workflow = config.Workflow
		}
		if config.Tracing != nil {
			result.Tracing = config.Tracing
		}
	}

	return &result
}

// CloneConfig creates a deep copy of a configuration
func CloneConfig(config *Config) *Config {
	if config == nil {
		return nil
	}

	clone := *config

	// Deep copy memory config
	if config.Memory != nil {
		memClone := *config.Memory
		if config.Memory.RAG != nil {
			ragClone := *config.Memory.RAG
			memClone.RAG = &ragClone
		}
		if config.Memory.Options != nil {
			memClone.Options = make(map[string]string)
			for k, v := range config.Memory.Options {
				memClone.Options[k] = v
			}
		}
		clone.Memory = &memClone
	}

	// Deep copy tools config
	if config.Tools != nil {
		toolsClone := *config.Tools
		if config.Tools.MCP != nil {
			mcpClone := *config.Tools.MCP
			if config.Tools.MCP.Servers != nil {
				mcpClone.Servers = make([]MCPServer, len(config.Tools.MCP.Servers))
				copy(mcpClone.Servers, config.Tools.MCP.Servers)
			}
			if config.Tools.MCP.Cache != nil {
				cacheClone := *config.Tools.MCP.Cache
				mcpClone.Cache = &cacheClone
			}
			toolsClone.MCP = &mcpClone
		}
		clone.Tools = &toolsClone
	}

	// Deep copy workflow config
	if config.Workflow != nil {
		workflowClone := *config.Workflow
		if config.Workflow.Agents != nil {
			workflowClone.Agents = make([]string, len(config.Workflow.Agents))
			copy(workflowClone.Agents, config.Workflow.Agents)
		}
		if config.Workflow.Memory != nil {
			memClone := *config.Workflow.Memory
			workflowClone.Memory = &memClone
		}
		clone.Workflow = &workflowClone
	}

	// Deep copy tracing config
	if config.Tracing != nil {
		tracingClone := *config.Tracing
		clone.Tracing = &tracingClone
	}

	return &clone
}
