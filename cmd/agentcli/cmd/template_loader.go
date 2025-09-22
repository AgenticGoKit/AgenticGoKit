package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kunalkushwaha/agenticgokit/internal/scaffold"
	"gopkg.in/yaml.v3"
)

// ExternalProjectTemplate represents a template that can be loaded from external files
type ExternalProjectTemplate struct {
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description" yaml:"description"`
	Features    []string               `json:"features" yaml:"features"`
	Config      ExternalTemplateConfig `json:"config" yaml:"config"`
}

// ExternalTemplateConfig represents the configuration part of a template
type ExternalTemplateConfig struct {
	NumAgents           int      `json:"numAgents" yaml:"numAgents"`
	Provider            string   `json:"provider" yaml:"provider"`
	OrchestrationMode   string   `json:"orchestrationMode" yaml:"orchestrationMode"`
	CollaborativeAgents []string `json:"collaborativeAgents,omitempty" yaml:"collaborativeAgents,omitempty"`
	SequentialAgents    []string `json:"sequentialAgents,omitempty" yaml:"sequentialAgents,omitempty"`
	LoopAgent           string   `json:"loopAgent,omitempty" yaml:"loopAgent,omitempty"`

	// Memory configuration
	MemoryEnabled     bool   `json:"memoryEnabled,omitempty" yaml:"memoryEnabled,omitempty"`
	MemoryProvider    string `json:"memoryProvider,omitempty" yaml:"memoryProvider,omitempty"`
	EmbeddingProvider string `json:"embeddingProvider,omitempty" yaml:"embeddingProvider,omitempty"`
	EmbeddingModel    string `json:"embeddingModel,omitempty" yaml:"embeddingModel,omitempty"`

	// RAG configuration
	RAGEnabled        bool    `json:"ragEnabled,omitempty" yaml:"ragEnabled,omitempty"`
	RAGChunkSize      int     `json:"ragChunkSize,omitempty" yaml:"ragChunkSize,omitempty"`
	RAGOverlap        int     `json:"ragOverlap,omitempty" yaml:"ragOverlap,omitempty"`
	RAGTopK           int     `json:"ragTopK,omitempty" yaml:"ragTopK,omitempty"`
	RAGScoreThreshold float64 `json:"ragScoreThreshold,omitempty" yaml:"ragScoreThreshold,omitempty"`
	HybridSearch      bool    `json:"hybridSearch,omitempty" yaml:"hybridSearch,omitempty"`
	SessionMemory     bool    `json:"sessionMemory,omitempty" yaml:"sessionMemory,omitempty"`

	// MCP configuration
	MCPEnabled    bool     `json:"mcpEnabled,omitempty" yaml:"mcpEnabled,omitempty"`
	MCPProduction bool     `json:"mcpProduction,omitempty" yaml:"mcpProduction,omitempty"`
	MCPTools      []string `json:"mcpTools,omitempty" yaml:"mcpTools,omitempty"`
	WithCache     bool     `json:"withCache,omitempty" yaml:"withCache,omitempty"`
	WithMetrics   bool     `json:"withMetrics,omitempty" yaml:"withMetrics,omitempty"`

	// Other options
	ResponsibleAI bool `json:"responsibleAI,omitempty" yaml:"responsibleAI,omitempty"`
	ErrorHandler  bool `json:"errorHandler,omitempty" yaml:"errorHandler,omitempty"`
	Visualize     bool `json:"visualize,omitempty" yaml:"visualize,omitempty"`
}

// TemplateLoader handles loading templates from various sources
type TemplateLoader struct {
	builtinTemplates  map[string]ProjectTemplate
	externalTemplates map[string]ProjectTemplate
	templatePaths     []string
}

// NewTemplateLoader creates a new template loader
func NewTemplateLoader() *TemplateLoader {
	loader := &TemplateLoader{
		builtinTemplates:  getBuiltinTemplates(),
		externalTemplates: make(map[string]ProjectTemplate),
		templatePaths:     getTemplatePaths(),
	}

	// Load external templates
	loader.loadExternalTemplates()

	return loader
}

// getTemplatePaths returns the paths where templates should be searched
func getTemplatePaths() []string {
	paths := []string{}

	// 1. Current directory .agenticgokit/templates/
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".agenticgokit", "templates"))
		// Also add examples/templates from current directory (for development)
		paths = append(paths, filepath.Join(cwd, "examples", "templates"))
	}

	// 2. User home directory ~/.agenticgokit/templates/
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".agenticgokit", "templates"))
	}

	// 3. System-wide /etc/agenticgokit/templates/ (Unix) or %PROGRAMDATA%/AgenticGoKit/templates/ (Windows)
	if systemPath := getSystemTemplatePath(); systemPath != "" {
		paths = append(paths, systemPath)
	}

	return paths
}

// getSystemTemplatePath returns the system-wide template path
func getSystemTemplatePath() string {
	// On Windows, use %PROGRAMDATA%/AgenticGoKit/templates/
	if programData := os.Getenv("PROGRAMDATA"); programData != "" {
		return filepath.Join(programData, "AgenticGoKit", "templates")
	}

	// On Unix-like systems, use /etc/agenticgokit/templates/
	return "/etc/agenticgokit/templates"
}

// loadExternalTemplates loads templates from external files
func (tl *TemplateLoader) loadExternalTemplates() {
	for _, templatePath := range tl.templatePaths {
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			continue
		}

		// Load all .json and .yaml/.yml files in the directory
		files, err := filepath.Glob(filepath.Join(templatePath, "*"))
		if err != nil {
			continue
		}

		for _, file := range files {
			if strings.HasSuffix(file, ".json") || strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") {
				if template, err := tl.loadTemplateFile(file); err == nil {
					templateName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
					tl.externalTemplates[templateName] = template
				}
			}
		}
	}
}

// loadTemplateFile loads a single template file
func (tl *TemplateLoader) loadTemplateFile(filePath string) (ProjectTemplate, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ProjectTemplate{}, fmt.Errorf("failed to read template file %s: %w", filePath, err)
	}

	var externalTemplate ExternalProjectTemplate

	// Determine file format and parse
	if strings.HasSuffix(filePath, ".json") {
		err = json.Unmarshal(data, &externalTemplate)
	} else {
		err = yaml.Unmarshal(data, &externalTemplate)
	}

	if err != nil {
		return ProjectTemplate{}, fmt.Errorf("failed to parse template file %s: %w", filePath, err)
	}

	// Convert to internal ProjectTemplate format
	return tl.convertExternalTemplate(externalTemplate), nil
}

// convertExternalTemplate converts an external template to internal format
func (tl *TemplateLoader) convertExternalTemplate(ext ExternalProjectTemplate) ProjectTemplate {
	config := scaffold.ProjectConfig{
		NumAgents:           ext.Config.NumAgents,
		Provider:            ext.Config.Provider,
		OrchestrationMode:   ext.Config.OrchestrationMode,
		CollaborativeAgents: ext.Config.CollaborativeAgents,
		SequentialAgents:    ext.Config.SequentialAgents,
		LoopAgent:           ext.Config.LoopAgent,

		MemoryEnabled:     ext.Config.MemoryEnabled,
		MemoryProvider:    ext.Config.MemoryProvider,
		EmbeddingProvider: ext.Config.EmbeddingProvider,
		EmbeddingModel:    ext.Config.EmbeddingModel,

		RAGEnabled:        ext.Config.RAGEnabled,
		RAGChunkSize:      ext.Config.RAGChunkSize,
		RAGOverlap:        ext.Config.RAGOverlap,
		RAGTopK:           ext.Config.RAGTopK,
		RAGScoreThreshold: ext.Config.RAGScoreThreshold,
		HybridSearch:      ext.Config.HybridSearch,
		SessionMemory:     ext.Config.SessionMemory,

		MCPEnabled:    ext.Config.MCPEnabled,
		MCPProduction: ext.Config.MCPProduction,
		MCPTools:      ext.Config.MCPTools,
		WithCache:     ext.Config.WithCache,
		WithMetrics:   ext.Config.WithMetrics,

		ResponsibleAI: ext.Config.ResponsibleAI,
		ErrorHandler:  ext.Config.ErrorHandler,
		Visualize:     ext.Config.Visualize,
	}

	// Set defaults if not specified
	if config.ResponsibleAI == false && config.ErrorHandler == false {
		config.ResponsibleAI = true
		config.ErrorHandler = true
	}

	return ProjectTemplate{
		Name:        ext.Name,
		Description: ext.Description,
		Config:      config,
		Features:    ext.Features,
	}
}

// GetAllTemplates returns all available templates (builtin + external)
func (tl *TemplateLoader) GetAllTemplates() map[string]ProjectTemplate {
	allTemplates := make(map[string]ProjectTemplate)

	// Add builtin templates first
	for name, template := range tl.builtinTemplates {
		allTemplates[name] = template
	}

	// Add external templates (they can override builtin ones)
	for name, template := range tl.externalTemplates {
		allTemplates[name] = template
	}

	return allTemplates
}

// GetTemplate returns a specific template by name
func (tl *TemplateLoader) GetTemplate(name string) (ProjectTemplate, bool) {
	// Check external templates first (they have priority)
	if template, exists := tl.externalTemplates[name]; exists {
		return template, true
	}

	// Check builtin templates
	if template, exists := tl.builtinTemplates[name]; exists {
		return template, true
	}

	return ProjectTemplate{}, false
}

// ListTemplatePaths returns the paths where templates are searched
func (tl *TemplateLoader) ListTemplatePaths() []string {
	return tl.templatePaths
}

// CreateTemplateExample creates an example template file
func (tl *TemplateLoader) CreateTemplateExample(filePath string, format string) error {
	example := ExternalProjectTemplate{
		Name:        "Custom Template",
		Description: "A custom project template example",
		Features:    []string{"custom-feature", "example"},
		Config: ExternalTemplateConfig{
			NumAgents:         2,
			Provider:          "openai",
			OrchestrationMode: "sequential",
			MemoryEnabled:     true,
			MemoryProvider:    "pgvector",
			EmbeddingProvider: "openai",
			RAGEnabled:        true,
			RAGChunkSize:      1000,
			MCPEnabled:        true,
			MCPTools:          []string{"web_search"},
			ResponsibleAI:     true,
			ErrorHandler:      true,
		},
	}

	var data []byte
	var err error

	if format == "json" {
		data, err = json.MarshalIndent(example, "", "  ")
	} else {
		data, err = yaml.Marshal(example)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal example template: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// getBuiltinTemplates returns the hardcoded builtin templates
func getBuiltinTemplates() map[string]ProjectTemplate {
	return map[string]ProjectTemplate{
		"basic": {
			Name:        "Basic Multi-Agent System",
			Description: "Simple multi-agent project with 2 agents and sequential orchestration",
			Config: scaffold.ProjectConfig{
				NumAgents:         2,
				Provider:          "openai",
				OrchestrationMode: "sequential",
				ResponsibleAI:     true,
				ErrorHandler:      true,
			},
			Features: []string{"sequential-orchestration", "error-handling"},
		},
		"research-assistant": {
			Name:        "Research Assistant",
			Description: "Multi-agent research system with sequential information gathering, analysis, and synthesis",
			Config: scaffold.ProjectConfig{
				NumAgents:         3,
				Provider:          "openai",
				OrchestrationMode: "sequential",
				SequentialAgents:  []string{"researcher", "analyzer", "synthesizer"},
				MCPEnabled:        true,
				MCPTools:          []string{"web_search", "summarize"},
				ResponsibleAI:     true,
				ErrorHandler:      true,
			},
			Features: []string{"sequential-research", "mcp-tools", "web-search", "specialized-agents"},
		},
		"rag-system": {
			Name:        "RAG Knowledge Base",
			Description: "Document ingestion and Q&A system with vector search",
			Config: scaffold.ProjectConfig{
				NumAgents:           3,
				Provider:            "openai",
				OrchestrationMode:   "collaborative",
				CollaborativeAgents: []string{"document-ingester", "query-processor", "response-generator"},
				MemoryEnabled:       true,
				MemoryProvider:      "pgvector",
				EmbeddingProvider:   "openai",
				RAGEnabled:          true,
				RAGChunkSize:        1000,
				RAGOverlap:          100,
				RAGTopK:             5,
				ResponsibleAI:       true,
				ErrorHandler:        true,
			},
			Features: []string{"memory", "rag", "vector-search", "collaborative-agents", "specialized-rag-agents"},
		},
		"data-pipeline": {
			Name:        "Data Processing Pipeline",
			Description: "Sequential data processing workflow with error handling",
			Config: scaffold.ProjectConfig{
				NumAgents:         4,
				Provider:          "openai",
				OrchestrationMode: "sequential",
				SequentialAgents:  []string{"ingester", "processor", "validator", "outputter"},
				ResponsibleAI:     true,
				ErrorHandler:      true,
				Visualize:         true,
			},
			Features: []string{"sequential-pipeline", "error-handling", "visualization"},
		},
		"chat-system": {
			Name:        "Conversational System",
			Description: "Chat agents with persistent memory and context",
			Config: scaffold.ProjectConfig{
				NumAgents:         2,
				Provider:          "openai",
				OrchestrationMode: "route",
				MemoryEnabled:     true,
				MemoryProvider:    "memory",
				SessionMemory:     true,
				ResponsibleAI:     true,
				ErrorHandler:      true,
			},
			Features: []string{"memory", "session-memory", "route-orchestration"},
		},
	}
}
