package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/agenticgokit/agenticgokit/internal/scaffold/templates"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// TemplateConfig represents a template configuration loaded from YAML/JSON
type TemplateConfig struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Features    []string `yaml:"features" json:"features"`
	Config      struct {
		NumAgents           int      `yaml:"numAgents" json:"numAgents"`
		Provider            string   `yaml:"provider" json:"provider"`
		OrchestrationMode   string   `yaml:"orchestrationMode" json:"orchestrationMode"`
		CollaborativeAgents []string `yaml:"collaborativeAgents" json:"collaborativeAgents"`
		SequentialAgents    []string `yaml:"sequentialAgents" json:"sequentialAgents"`
		LoopAgent           string   `yaml:"loopAgent" json:"loopAgent"`
		MaxIterations       int      `yaml:"maxIterations" json:"maxIterations"`

		// Memory configuration
		MemoryEnabled     bool    `yaml:"memoryEnabled" json:"memoryEnabled"`
		MemoryProvider    string  `yaml:"memoryProvider" json:"memoryProvider"`
		EmbeddingProvider string  `yaml:"embeddingProvider" json:"embeddingProvider"`
		EmbeddingModel    string  `yaml:"embeddingModel" json:"embeddingModel"`
		RAGEnabled        bool    `yaml:"ragEnabled" json:"ragEnabled"`
		RAGChunkSize      int     `yaml:"ragChunkSize" json:"ragChunkSize"`
		RAGOverlap        int     `yaml:"ragOverlap" json:"ragOverlap"`
		RAGTopK           int     `yaml:"ragTopK" json:"ragTopK"`
		RAGScoreThreshold float64 `yaml:"ragScoreThreshold" json:"ragScoreThreshold"`
		HybridSearch      bool    `yaml:"hybridSearch" json:"hybridSearch"`
		SessionMemory     bool    `yaml:"sessionMemory" json:"sessionMemory"`

		// MCP configuration
		MCPEnabled    bool     `yaml:"mcpEnabled" json:"mcpEnabled"`
		MCPProduction bool     `yaml:"mcpProduction" json:"mcpProduction"`
		MCPTransport  string   `yaml:"mcpTransport" json:"mcpTransport"`
		MCPTools      []string `yaml:"mcpTools" json:"mcpTools"`

		// Other features
		ResponsibleAI bool `yaml:"responsibleAI" json:"responsibleAI"`
		ErrorHandler  bool `yaml:"errorHandler" json:"errorHandler"`
		WithCache     bool `yaml:"withCache" json:"withCache"`
		WithMetrics   bool `yaml:"withMetrics" json:"withMetrics"`
		Visualize     bool `yaml:"visualize" json:"visualize"`

		// WebUI configuration
		WebUIEnabled      bool   `yaml:"webUIEnabled" json:"webUIEnabled"`
		WebUIPort         int    `yaml:"webUIPort" json:"webUIPort"`
		WebUIStaticDir    string `yaml:"webUIStaticDir" json:"webUIStaticDir"`
		WebUITemplatesDir string `yaml:"webUITemplatesDir" json:"webUITemplatesDir"`

		// Performance settings
		MaxConcurrency       int `yaml:"maxConcurrency" json:"maxConcurrency"`
		OrchestrationTimeout int `yaml:"orchestrationTimeout" json:"orchestrationTimeout"`
	} `yaml:"config" json:"config"`

	// Agent-specific configurations
	Agents map[string]AgentTemplateConfig `yaml:"agents" json:"agents"`

	// MCP server configurations
	MCPServers []MCPServerConfig `yaml:"mcpServers" json:"mcpServers"`
}

// AgentTemplateConfig represents agent-specific configuration in templates
type AgentTemplateConfig struct {
	Role         string             `yaml:"role" json:"role"`
	Description  string             `yaml:"description" json:"description"`
	Capabilities []string           `yaml:"capabilities" json:"capabilities"`
	SystemPrompt string             `yaml:"systemPrompt" json:"systemPrompt"`
	Enabled      bool               `yaml:"enabled" json:"enabled"`
	Timeout      int                `yaml:"timeout" json:"timeout"`
	LLM          *LLMTemplateConfig `yaml:"llm" json:"llm"`
	RetryPolicy  *RetryPolicyConfig `yaml:"retryPolicy" json:"retryPolicy"`
	RateLimit    *RateLimitConfig   `yaml:"rateLimit" json:"rateLimit"`
	Metadata     map[string]string  `yaml:"metadata" json:"metadata"`
}

// LLMTemplateConfig represents LLM configuration in templates
type LLMTemplateConfig struct {
	Provider         string  `yaml:"provider" json:"provider"`
	Model            string  `yaml:"model" json:"model"`
	Temperature      float64 `yaml:"temperature" json:"temperature"`
	MaxTokens        int     `yaml:"maxTokens" json:"maxTokens"`
	TopP             float64 `yaml:"topP" json:"topP"`
	FrequencyPenalty float64 `yaml:"frequencyPenalty" json:"frequencyPenalty"`
	PresencePenalty  float64 `yaml:"presencePenalty" json:"presencePenalty"`
}

// RetryPolicyConfig represents retry policy configuration
type RetryPolicyConfig struct {
	MaxRetries    int     `yaml:"maxRetries" json:"maxRetries"`
	BaseDelayMs   int     `yaml:"baseDelayMs" json:"baseDelayMs"`
	MaxDelayMs    int     `yaml:"maxDelayMs" json:"maxDelayMs"`
	BackoffFactor float64 `yaml:"backoffFactor" json:"backoffFactor"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int `yaml:"requestsPerSecond" json:"requestsPerSecond"`
	BurstSize         int `yaml:"burstSize" json:"burstSize"`
}

// MCPServerConfig represents MCP server configuration
type MCPServerConfig struct {
	Name    string `yaml:"name" json:"name"`
	Type    string `yaml:"type" json:"type"`
	Host    string `yaml:"host" json:"host"`
	Port    int    `yaml:"port" json:"port"`
	Command string `yaml:"command" json:"command"`
	Enabled bool   `yaml:"enabled" json:"enabled"`
}

// TemplateGenerator handles template-based project generation
type TemplateGenerator struct {
	templateConfig TemplateConfig
	projectConfig  ProjectConfig
}

// NewTemplateGenerator creates a new template generator
func NewTemplateGenerator(templatePath string) (*TemplateGenerator, error) {
	// Load template configuration
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var templateConfig TemplateConfig
	if err := yaml.Unmarshal(templateData, &templateConfig); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	// Set defaults
	if templateConfig.Config.MaxConcurrency == 0 {
		templateConfig.Config.MaxConcurrency = 10
	}
	if templateConfig.Config.OrchestrationTimeout == 0 {
		templateConfig.Config.OrchestrationTimeout = 30
	}
	if templateConfig.Config.RAGTopK == 0 {
		templateConfig.Config.RAGTopK = 5
	}
	if templateConfig.Config.RAGScoreThreshold == 0 {
		templateConfig.Config.RAGScoreThreshold = 0.7
	}
	if templateConfig.Config.RAGChunkSize == 0 {
		templateConfig.Config.RAGChunkSize = 1000
	}
	if templateConfig.Config.RAGOverlap == 0 {
		templateConfig.Config.RAGOverlap = 100
	}

	// Enable agents by default
	for name, agent := range templateConfig.Agents {
		if !agent.Enabled {
			agent.Enabled = true
			templateConfig.Agents[name] = agent
		}
	}

	return &TemplateGenerator{
		templateConfig: templateConfig,
	}, nil
}

// GenerateProject creates a project from the template
func (tg *TemplateGenerator) GenerateProject(projectName string) error {
	// Convert template config to project config
	tg.projectConfig = tg.convertToProjectConfig(projectName)

	// Create the project using the standard scaffold system
	if err := CreateAgentProjectModular(tg.projectConfig); err != nil {
		return fmt.Errorf("failed to create project from template: %w", err)
	}

	// Generate enhanced agentflow.toml with agent configurations
	if err := tg.generateEnhancedConfig(projectName); err != nil {
		return fmt.Errorf("failed to generate enhanced configuration: %w", err)
	}

	fmt.Printf("Project '%s' created successfully from template '%s'.\n", projectName, tg.templateConfig.Name)
	fmt.Printf("Template: %s\n", tg.templateConfig.Description)
	fmt.Printf("Features: %s\n", strings.Join(tg.templateConfig.Features, ", "))

	return nil
}

// convertToProjectConfig converts template config to ProjectConfig
func (tg *TemplateGenerator) convertToProjectConfig(projectName string) ProjectConfig {
	tc := tg.templateConfig.Config

	config := ProjectConfig{
		Name:                 projectName,
		NumAgents:            tc.NumAgents,
		Provider:             tc.Provider,
		OrchestrationMode:    tc.OrchestrationMode,
		CollaborativeAgents:  tc.CollaborativeAgents,
		SequentialAgents:     tc.SequentialAgents,
		LoopAgent:            tc.LoopAgent,
		MaxIterations:        tc.MaxIterations,
		MaxConcurrency:       tc.MaxConcurrency,
		OrchestrationTimeout: tc.OrchestrationTimeout,

		// Memory settings
		MemoryEnabled:     tc.MemoryEnabled,
		MemoryProvider:    tc.MemoryProvider,
		EmbeddingProvider: tc.EmbeddingProvider,
		EmbeddingModel:    tc.EmbeddingModel,
		RAGEnabled:        tc.RAGEnabled,
		RAGChunkSize:      tc.RAGChunkSize,
		RAGOverlap:        tc.RAGOverlap,
		RAGTopK:           tc.RAGTopK,
		RAGScoreThreshold: tc.RAGScoreThreshold,
		HybridSearch:      tc.HybridSearch,
		SessionMemory:     tc.SessionMemory,

		// MCP settings
		MCPEnabled:    tc.MCPEnabled,
		MCPProduction: tc.MCPProduction,
		MCPTransport:  tc.MCPTransport,
		MCPTools:      tc.MCPTools,

		// Other features
		ResponsibleAI: tc.ResponsibleAI,
		ErrorHandler:  tc.ErrorHandler,
		WithCache:     tc.WithCache,
		WithMetrics:   tc.WithMetrics,
		Visualize:     tc.Visualize,
	}

	// Set embedding dimensions based on model
	config.EmbeddingDimensions = GetModelDimensions(tc.EmbeddingProvider, tc.EmbeddingModel)

	// Apply intelligent defaults to ensure proper configuration
	ApplyIntelligentDefaults(&config)

	return config
}

// formatTOMLStringArray formats a string slice as a proper TOML array
func formatTOMLStringArray(slice []string) string {
	if len(slice) == 0 {
		return "[]"
	}
	quoted := make([]string, len(slice))
	for i, s := range slice {
		quoted[i] = strconv.Quote(s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// createTemplateFuncMap creates a comprehensive function map for template execution
func createTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		// String formatting functions
		"printf": func(format string, args ...interface{}) string {
			// Handle slice formatting specially for TOML arrays
			if format == "%q" && len(args) == 1 {
				if slice, ok := args[0].([]string); ok {
					return formatTOMLStringArray(slice)
				}
			}
			return fmt.Sprintf(format, args...)
		},
		"sprintf": fmt.Sprintf,

		// String manipulation functions
		"toString": func(v interface{}) string {
			return fmt.Sprintf("%v", v)
		},
		"quote": func(s string) string {
			return strconv.Quote(s)
		},
		"quoteSlice": func(slice []string) string {
			quoted := make([]string, len(slice))
			for i, s := range slice {
				quoted[i] = strconv.Quote(s)
			}
			return "[" + strings.Join(quoted, ", ") + "]"
		},

		// Conditional functions
		"eq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
		"ne": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
		},
		"lt": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
		},
		"le": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) <= fmt.Sprintf("%v", b)
		},
		"gt": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) > fmt.Sprintf("%v", b)
		},
		"ge": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) >= fmt.Sprintf("%v", b)
		},

		// Logical functions
		"and": func(args ...bool) bool {
			for _, arg := range args {
				if !arg {
					return false
				}
			}
			return true
		},
		"or": func(args ...bool) bool {
			for _, arg := range args {
				if arg {
					return true
				}
			}
			return false
		},
		"not": func(b bool) bool {
			return !b
		},

		// Utility functions
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || fmt.Sprintf("%v", value) == "" {
				return defaultValue
			}
			return value
		},
		"empty": func(value interface{}) bool {
			if value == nil {
				return true
			}
			str := fmt.Sprintf("%v", value)
			return str == "" || str == "<nil>"
		},

		// String processing
		"title": func(s string) string {
			titleCaser := cases.Title(language.English)
			return titleCaser.String(s)
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"trim":  strings.TrimSpace,
		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"join": func(sep string, elems []string) string {
			return strings.Join(elems, sep)
		},
	}
}

// generateEnhancedConfig creates an enhanced agentflow.toml with agent configurations
func (tg *TemplateGenerator) generateEnhancedConfig(projectName string) error {
	// Prepare template data
	templateData := struct {
		TemplateName     string
		Description      string
		Config           ProjectConfig
		Agents           []AgentConfigData
		GlobalLLM        LLMTemplateConfig
		MemoryConnection string
		MCPServers       []MCPServerConfig
	}{
		TemplateName:     tg.templateConfig.Name,
		Description:      tg.templateConfig.Description,
		Config:           tg.projectConfig,
		Agents:           tg.convertAgentsToConfigData(),
		GlobalLLM:        tg.getGlobalLLMConfig(),
		MemoryConnection: tg.getMemoryConnection(),
		MCPServers:       tg.templateConfig.MCPServers,
	}

	// Parse and execute the complete config template
	tmpl, err := template.New("agentconfig").Funcs(createTemplateFuncMap()).Parse(templates.CompleteAgentConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}

	// Create agentflow.toml file
	configPath := filepath.Join(tg.projectConfig.Name, "agentflow.toml")
	configFile, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create agentflow.toml: %w", err)
	}
	defer configFile.Close()

	// Execute template
	if err := tmpl.Execute(configFile, templateData); err != nil {
		return fmt.Errorf("failed to execute config template: %w", err)
	}

	fmt.Printf("Generated enhanced configuration: %s\n", configPath)
	return nil
}

// AgentConfigData represents agent data for template generation
type AgentConfigData struct {
	Name         string
	DisplayName  string
	Role         string
	Description  string
	SystemPrompt string
	Capabilities []string
	Enabled      bool
	Timeout      int
	LLM          *LLMTemplateConfig
	RetryPolicy  *RetryPolicyConfig
	RateLimit    *RateLimitConfig
	Metadata     map[string]string
}

// convertAgentsToConfigData converts template agents to config data
func (tg *TemplateGenerator) convertAgentsToConfigData() []AgentConfigData {
	var agents []AgentConfigData

	for name, agent := range tg.templateConfig.Agents {
		titleCaser := cases.Title(language.English)
		displayName := strings.ReplaceAll(titleCaser.String(strings.ReplaceAll(name, "-", " ")), " ", "")

		agents = append(agents, AgentConfigData{
			Name:         name,
			DisplayName:  displayName,
			Role:         agent.Role,
			Description:  agent.Description,
			SystemPrompt: agent.SystemPrompt,
			Capabilities: agent.Capabilities,
			Enabled:      agent.Enabled,
			Timeout:      agent.Timeout,
			LLM:          agent.LLM,
			RetryPolicy:  agent.RetryPolicy,
			RateLimit:    agent.RateLimit,
			Metadata:     agent.Metadata,
		})
	}

	return agents
}

// getGlobalLLMConfig returns global LLM configuration
func (tg *TemplateGenerator) getGlobalLLMConfig() LLMTemplateConfig {
	provider := tg.templateConfig.Config.Provider
	model := "gpt-4"

	// Set appropriate model based on provider
	switch provider {
	case "openai":
		model = "gpt-4"
	case "azure":
		model = "gpt-4"
	case "ollama":
		model = "llama2"
	default:
		model = "gpt-4"
	}

	return LLMTemplateConfig{
		Provider:    provider,
		Model:       model,
		Temperature: 0.7,  // Default global temperature
		MaxTokens:   2000, // Default global max tokens
	}
}

// getMemoryConnection returns the memory connection string
func (tg *TemplateGenerator) getMemoryConnection() string {
	switch tg.templateConfig.Config.MemoryProvider {
	case "pgvector":
		return "postgres://user:password@localhost:15432/agentflow?sslmode=disable"
	case "weaviate":
		return "http://localhost:8080"
	default:
		return "memory"
	}
}

// ListAvailableTemplates returns a list of available templates
func ListAvailableTemplates() ([]string, error) {
	templatesDir := "examples/templates"

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []string
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
			templates = append(templates, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
		}
	}

	return templates, nil
}

// GetTemplateInfo returns information about a specific template
func GetTemplateInfo(templateName string) (*TemplateConfig, error) {
	templatePath := filepath.Join("examples/templates", templateName+".yaml")

	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var templateConfig TemplateConfig
	if err := yaml.Unmarshal(templateData, &templateConfig); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	return &templateConfig, nil
}

