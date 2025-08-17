// Package core provides essential configuration types and loading for AgentFlow.
package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field      string      `json:"field"`
	Value      interface{} `json:"value"`
	Message    string      `json:"message"`
	Suggestion string      `json:"suggestion"`
}

func (e ValidationError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%s: %s. Suggestion: %s", e.Field, e.Message, e.Suggestion)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ConfigValidator interface for agent configuration validation
type ConfigValidator interface {
	ValidateAgentConfig(name string, config *AgentConfig) []ValidationError
	ValidateLLMConfig(config *AgentLLMConfig) []ValidationError
	ValidateOrchestrationAgents(orchestration *OrchestrationConfigToml, agents map[string]AgentConfig) []ValidationError
	ValidateCapabilities(capabilities []string) []ValidationError
	ValidateConfig(config *Config) []ValidationError
}

// ConfigResolver interface for configuration resolution with environment overrides
type ConfigResolver interface {
	ResolveAgentConfigWithEnv(agentName string) (*ResolvedAgentConfig, error)
	ApplyEnvironmentOverrides() error
	GetResolvedConfig() *Config
	ResolveAllAgents() (map[string]*ResolvedAgentConfig, error)
	ValidateResolvedConfig() []ValidationError
}

// ConfigReloader interface defines the contract for configuration hot-reloading
type ConfigReloader interface {
	StartWatching(configPath string) error
	StopWatching() error
	ReloadConfig() error
	OnConfigChanged(callback func(*Config, error))
	GetLastReloadTime() time.Time
	IsWatching() bool
}

// AgentManager is defined in agent.go

// AgentLLMConfig represents LLM provider configuration for agents
type AgentLLMConfig struct {
	Provider         string  `toml:"provider"`
	Model            string  `toml:"model"`
	Temperature      float64 `toml:"temperature"`
	MaxTokens        int     `toml:"max_tokens"`
	TimeoutSeconds   int     `toml:"timeout_seconds"`
	TopP             float64 `toml:"top_p,omitempty"`
	FrequencyPenalty float64 `toml:"frequency_penalty,omitempty"`
	PresencePenalty  float64 `toml:"presence_penalty,omitempty"`
}

// AgentConfig represents agent-specific configuration
type AgentConfig struct {
	Role         string            `toml:"role"`
	Description  string            `toml:"description"`
	SystemPrompt string            `toml:"system_prompt"`
	Capabilities []string          `toml:"capabilities"`
	Enabled      bool              `toml:"enabled"`
	LLM          *AgentLLMConfig   `toml:"llm,omitempty"`
	Metadata     map[string]string `toml:"metadata,omitempty"`

	// Advanced configuration
	RetryPolicy *AgentRetryPolicyConfig `toml:"retry_policy,omitempty"`
	RateLimit   *RateLimitConfig        `toml:"rate_limit,omitempty"`
	Timeout     int                     `toml:"timeout_seconds,omitempty"`
}

// AgentRetryPolicyConfig represents retry policy configuration for agents
type AgentRetryPolicyConfig struct {
	MaxRetries    int     `toml:"max_retries"`
	BaseDelayMs   int     `toml:"base_delay_ms"`
	MaxDelayMs    int     `toml:"max_delay_ms"`
	BackoffFactor float64 `toml:"backoff_factor"`
}

// RateLimitConfig represents rate limiting configuration for agents
type RateLimitConfig struct {
	RequestsPerSecond int `toml:"requests_per_second"`
	BurstSize         int `toml:"burst_size"`
}

// ResolvedLLMConfig represents resolved LLM configuration for runtime use
type ResolvedLLMConfig struct {
	Provider         string
	Model            string
	APIKey           string // API key for the provider
	Temperature      float64
	MaxTokens        int
	Timeout          time.Duration
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
}

// ResolvedAgentConfig represents resolved agent configuration for runtime use
type ResolvedAgentConfig struct {
	Name         string
	Role         string
	Description  string
	SystemPrompt string
	Capabilities []string
	Enabled      bool
	LLMConfig    *ResolvedLLMConfig
	RetryPolicy  *AgentRetryPolicyConfig
	RateLimit    *RateLimitConfig
	Timeout      time.Duration
}

// Config represents the essential AgentFlow configuration structure
type Config struct {
	AgentFlow struct {
		Name     string `toml:"name"`
		Version  string `toml:"version"`
		Provider string `toml:"provider"`
	} `toml:"agent_flow"`

	Logging struct {
		Level  string `toml:"level"`
		Format string `toml:"format"`
	} `toml:"logging"`

	Runtime struct {
		MaxConcurrentAgents int `toml:"max_concurrent_agents"`
		TimeoutSeconds      int `toml:"timeout_seconds"`
	} `toml:"runtime"`

	// Global LLM configuration
	LLM AgentLLMConfig `toml:"llm"`

	// Agent-specific configuration
	Agents map[string]AgentConfig `toml:"agents"`

	// Agent memory configuration
	AgentMemory AgentMemoryConfig `toml:"agent_memory"`

	// Error routing configuration
	ErrorRouting struct {
		Enabled              bool                     `toml:"enabled"`
		MaxRetries           int                      `toml:"max_retries"`
		RetryDelayMs         int                      `toml:"retry_delay_ms"`
		EnableCircuitBreaker bool                     `toml:"enable_circuit_breaker"`
		ErrorHandlerName     string                   `toml:"error_handler_name"`
		CategoryHandlers     map[string]string        `toml:"category_handlers"`
		SeverityHandlers     map[string]string        `toml:"severity_handlers"`
		CircuitBreaker       CircuitBreakerConfigToml `toml:"circuit_breaker"`
		Retry                RetryConfigToml          `toml:"retry"`
	} `toml:"error_routing"`

	Providers map[string]map[string]interface{} `toml:"providers"`

	// MCP configuration
	MCP MCPConfigToml `toml:"mcp"`

	// Orchestration configuration
	Orchestration OrchestrationConfigToml `toml:"orchestration"`
}

// CircuitBreakerConfigToml represents circuit breaker configuration in TOML
type CircuitBreakerConfigToml struct {
	FailureThreshold int `toml:"failure_threshold"`
	SuccessThreshold int `toml:"success_threshold"`
	TimeoutMs        int `toml:"timeout_ms"`
	ResetTimeoutMs   int `toml:"reset_timeout_ms"`
	HalfOpenMaxCalls int `toml:"half_open_max_calls"`
}

// RetryConfigToml represents retry configuration in TOML
type RetryConfigToml struct {
	MaxRetries    int     `toml:"max_retries"`
	BaseDelayMs   int     `toml:"base_delay_ms"`
	MaxDelayMs    int     `toml:"max_delay_ms"`
	BackoffFactor float64 `toml:"backoff_factor"`
	EnableJitter  bool    `toml:"enable_jitter"`
}

// MCPConfigToml represents MCP configuration in TOML format
type MCPConfigToml struct {
	Enabled           bool                  `toml:"enabled"`
	EnableDiscovery   bool                  `toml:"enable_discovery"`
	DiscoveryTimeout  int                   `toml:"discovery_timeout_ms"`
	ScanPorts         []int                 `toml:"scan_ports"`
	ConnectionTimeout int                   `toml:"connection_timeout_ms"`
	MaxRetries        int                   `toml:"max_retries"`
	RetryDelay        int                   `toml:"retry_delay_ms"`
	EnableCaching     bool                  `toml:"enable_caching"`
	CacheTimeout      int                   `toml:"cache_timeout_ms"`
	MaxConnections    int                   `toml:"max_connections"`
	Servers           []MCPServerConfigToml `toml:"servers"`
}

// MCPServerConfigToml represents individual MCP server configuration in TOML
type MCPServerConfigToml struct {
	Name    string `toml:"name"`
	Type    string `toml:"type"` // tcp, stdio, docker, websocket
	Host    string `toml:"host,omitempty"`
	Port    int    `toml:"port,omitempty"`
	Command string `toml:"command,omitempty"` // for stdio transport
	Enabled bool   `toml:"enabled"`
}

// OrchestrationConfigToml represents orchestration configuration in TOML format
type OrchestrationConfigToml struct {
	Mode                string   `toml:"mode"`                 // route, collaborative, sequential, loop, mixed
	TimeoutSeconds      int      `toml:"timeout_seconds"`      // Overall timeout for orchestration operations
	MaxIterations       int      `toml:"max_iterations"`       // For loop mode: maximum iterations
	SequentialAgents    []string `toml:"sequential_agents"`    // For sequential mode: ordered list of agent names
	CollaborativeAgents []string `toml:"collaborative_agents"` // For mixed mode: agents that run collaboratively
	LoopAgent           string   `toml:"loop_agent"`           // For loop mode: agent to run in loop
}

// LoadConfig loads configuration from the specified TOML file path
func LoadConfig(path string) (*Config, error) {
	// If path is empty, return default configuration
	if path == "" {
		config := &Config{}
		applyConfigDefaults(config)
		return config, nil
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", path, err)
	}

	// Parse TOML
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML configuration: %w", err)
	}

	// Apply defaults and environment overrides using internal implementation
	// TODO: This will be replaced with internal loader after refactoring is complete
	applyConfigDefaults(&config)

	return &config, nil
}

// LoadConfigFromWorkingDir loads agentflow.toml from the current working directory
func LoadConfigFromWorkingDir() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	configPath := filepath.Join(wd, "agentflow.toml")
	return LoadConfig(configPath)
}

// ResolveAgentConfig resolves agent configuration with environment variable support
func (c *Config) ResolveAgentConfig(agentName string) (*ResolvedAgentConfig, error) {
	// Use the resolver for environment variable support
	resolver := NewConfigResolver(c)
	return resolver.ResolveAgentConfigWithEnv(agentName)
}

// Factory functions for creating configuration components
func NewDefaultConfigValidator() ConfigValidator {
	// TODO: This will be replaced with internal validator after refactoring is complete
	return &noOpValidator{}
}

func NewConfigResolver(config *Config) ConfigResolver {
	// TODO: This will be replaced with internal resolver after refactoring is complete
	return &noOpResolver{config: config}
}

func NewConfigReloader(validator ConfigValidator, agentManager AgentManager) ConfigReloader {
	// TODO: This will be replaced with internal reloader after refactoring is complete
	return &noOpReloader{validator: validator, agentManager: agentManager}
}

// Temporary no-op implementations during refactoring
type noOpValidator struct{}

func (v *noOpValidator) ValidateAgentConfig(name string, config *AgentConfig) []ValidationError {
	return []ValidationError{}
}

func (v *noOpValidator) ValidateLLMConfig(config *AgentLLMConfig) []ValidationError {
	return []ValidationError{}
}

func (v *noOpValidator) ValidateOrchestrationAgents(orchestration *OrchestrationConfigToml, agents map[string]AgentConfig) []ValidationError {
	return []ValidationError{}
}

func (v *noOpValidator) ValidateCapabilities(capabilities []string) []ValidationError {
	return []ValidationError{}
}

func (v *noOpValidator) ValidateConfig(config *Config) []ValidationError {
	return []ValidationError{}
}

type noOpResolver struct {
	config *Config
}

func (r *noOpResolver) ResolveAgentConfigWithEnv(agentName string) (*ResolvedAgentConfig, error) {
	agent, exists := r.config.Agents[agentName]
	if !exists {
		return nil, fmt.Errorf("agent '%s' not found in configuration", agentName)
	}

	// Simple resolution without environment overrides for now
	return &ResolvedAgentConfig{
		Name:         agentName,
		Role:         agent.Role,
		Description:  agent.Description,
		SystemPrompt: agent.SystemPrompt,
		Capabilities: agent.Capabilities,
		Enabled:      agent.Enabled,
		LLMConfig:    r.resolveLLMConfig(&agent),
		RetryPolicy:  agent.RetryPolicy,
		RateLimit:    agent.RateLimit,
		Timeout:      time.Duration(agent.Timeout) * time.Second,
	}, nil
}

func (r *noOpResolver) ApplyEnvironmentOverrides() error {
	return nil
}

func (r *noOpResolver) GetResolvedConfig() *Config {
	return r.config
}

func (r *noOpResolver) ResolveAllAgents() (map[string]*ResolvedAgentConfig, error) {
	resolved := make(map[string]*ResolvedAgentConfig)
	for agentName := range r.config.Agents {
		agentConfig, err := r.ResolveAgentConfigWithEnv(agentName)
		if err != nil {
			return nil, err
		}
		resolved[agentName] = agentConfig
	}
	return resolved, nil
}

func (r *noOpResolver) ValidateResolvedConfig() []ValidationError {
	return []ValidationError{}
}

func (r *noOpResolver) resolveLLMConfig(agent *AgentConfig) *ResolvedLLMConfig {
	// Start with global LLM config
	resolved := &ResolvedLLMConfig{
		Provider:         r.config.LLM.Provider,
		Model:            r.config.LLM.Model,
		Temperature:      r.config.LLM.Temperature,
		MaxTokens:        r.config.LLM.MaxTokens,
		Timeout:          time.Duration(r.config.LLM.TimeoutSeconds) * time.Second,
		TopP:             r.config.LLM.TopP,
		FrequencyPenalty: r.config.LLM.FrequencyPenalty,
		PresencePenalty:  r.config.LLM.PresencePenalty,
	}

	// Override with agent-specific LLM config if provided
	if agent.LLM != nil {
		if agent.LLM.Provider != "" {
			resolved.Provider = agent.LLM.Provider
		}
		if agent.LLM.Model != "" {
			resolved.Model = agent.LLM.Model
		}
		if agent.LLM.Temperature != 0 {
			resolved.Temperature = agent.LLM.Temperature
		}
		if agent.LLM.MaxTokens != 0 {
			resolved.MaxTokens = agent.LLM.MaxTokens
		}
		if agent.LLM.TimeoutSeconds != 0 {
			resolved.Timeout = time.Duration(agent.LLM.TimeoutSeconds) * time.Second
		}
		if agent.LLM.TopP != 0 {
			resolved.TopP = agent.LLM.TopP
		}
		if agent.LLM.FrequencyPenalty != 0 {
			resolved.FrequencyPenalty = agent.LLM.FrequencyPenalty
		}
		if agent.LLM.PresencePenalty != 0 {
			resolved.PresencePenalty = agent.LLM.PresencePenalty
		}
	}

	return resolved
}

type noOpReloader struct {
	validator    ConfigValidator
	agentManager AgentManager
	isWatching   bool
	lastReload   time.Time
}

func (r *noOpReloader) StartWatching(configPath string) error {
	r.isWatching = true
	r.lastReload = time.Now()
	return nil
}

func (r *noOpReloader) StopWatching() error {
	r.isWatching = false
	return nil
}

func (r *noOpReloader) ReloadConfig() error {
	r.lastReload = time.Now()
	return nil
}

func (r *noOpReloader) OnConfigChanged(callback func(*Config, error)) {
	// No-op for now
}

func (r *noOpReloader) GetLastReloadTime() time.Time {
	return r.lastReload
}

func (r *noOpReloader) IsWatching() bool {
	return r.isWatching
}

// GetAgentCapabilities returns the capabilities for a specific agent
func (c *Config) GetAgentCapabilities(name string) []string {
	if agent, exists := c.Agents[name]; exists {
		return agent.Capabilities
	}
	return []string{}
}

// IsAgentEnabled checks if an agent is enabled
func (c *Config) IsAgentEnabled(name string) bool {
	if agent, exists := c.Agents[name]; exists {
		return agent.Enabled
	}
	return false
}

// GetEnabledAgents returns a list of enabled agent names
func (c *Config) GetEnabledAgents() []string {
	var enabled []string
	for name, agent := range c.Agents {
		if agent.Enabled {
			enabled = append(enabled, name)
		}
	}
	return enabled
}

// ApplyLoggingConfig applies logging configuration (no-op for now)
func (c *Config) ApplyLoggingConfig() {
	// TODO: This will be replaced with internal implementation after refactoring is complete
}

// ValidateOrchestrationConfig validates orchestration configuration
func (c *Config) ValidateOrchestrationConfig() error {
	// Basic validation to make tests pass during refactoring
	if c.Orchestration.Mode == "" {
		return fmt.Errorf("orchestration mode is required")
	}

	validModes := []string{"route", "collaborative", "sequential", "loop", "mixed"}
	isValidMode := false
	for _, mode := range validModes {
		if c.Orchestration.Mode == mode {
			isValidMode = true
			break
		}
	}
	if !isValidMode {
		return fmt.Errorf("invalid orchestration mode: %s", c.Orchestration.Mode)
	}

	if c.Orchestration.TimeoutSeconds <= 0 {
		return fmt.Errorf("orchestration timeout_seconds must be positive")
	}

	switch c.Orchestration.Mode {
	case "sequential":
		if len(c.Orchestration.SequentialAgents) == 0 {
			return fmt.Errorf("sequential mode requires at least one agent")
		}
	case "loop":
		if c.Orchestration.LoopAgent == "" {
			return fmt.Errorf("loop mode requires a loop agent")
		}
		if c.Orchestration.MaxIterations <= 0 {
			return fmt.Errorf("orchestration max_iterations must be positive for loop mode")
		}
	case "mixed":
		if len(c.Orchestration.SequentialAgents) == 0 && len(c.Orchestration.CollaborativeAgents) == 0 {
			return fmt.Errorf("mixed mode requires at least one sequential or collaborative agent")
		}
	}

	return nil
}

// InitializeProvider initializes the configured provider
func (c *Config) InitializeProvider() (ModelProvider, error) {
	// TODO: This will be replaced with internal implementation after refactoring is complete
	// For now, return a basic provider based on the configuration
	switch c.LLM.Provider {
	case "openai":
		return NewOpenAIAdapter("", c.LLM.Model, c.LLM.MaxTokens, float32(c.LLM.Temperature))
	case "azure":
		return NewAzureOpenAIAdapter(AzureOpenAIAdapterOptions{
			Endpoint:            "",
			APIKey:              "",
			ChatDeployment:      c.LLM.Model,
			EmbeddingDeployment: "",
		})
	case "ollama":
		return NewOllamaAdapter("http://localhost:11434", c.LLM.Model, c.LLM.MaxTokens, float32(c.LLM.Temperature))
	default:
		return NewOpenAIAdapter("", c.LLM.Model, c.LLM.MaxTokens, float32(c.LLM.Temperature))
	}
}

// applyConfigDefaults applies default values to configuration
func applyConfigDefaults(config *Config) {
	// Set defaults if not specified
	if config.AgentFlow.Name == "" {
		config.AgentFlow.Name = "default-agent"
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Runtime.MaxConcurrentAgents == 0 {
		config.Runtime.MaxConcurrentAgents = 10
	}
	if config.Runtime.TimeoutSeconds == 0 {
		config.Runtime.TimeoutSeconds = 30
	}

	// Set global LLM defaults if not specified
	if config.LLM.Provider == "" {
		config.LLM.Provider = config.AgentFlow.Provider // Use the main provider as default
	}
	if config.LLM.Temperature == 0 {
		config.LLM.Temperature = 0.7
	}
	if config.LLM.MaxTokens == 0 {
		config.LLM.MaxTokens = 800
	}
	if config.LLM.TimeoutSeconds == 0 {
		config.LLM.TimeoutSeconds = 30
	}

	// Set agent defaults if not specified
	for name, agent := range config.Agents {
		// Set default role if not specified
		if agent.Role == "" {
			agent.Role = name + "_agent"
		}

		// Set default description if not specified
		if agent.Description == "" {
			agent.Description = "Agent for " + name
		}

		// Set default timeout if not specified
		if agent.Timeout == 0 {
			agent.Timeout = 30
		}

		// Update the agent in the map
		config.Agents[name] = agent
	}
}
