package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kunalkushwaha/agenticgokit/internal/scaffold"
)

// ConsolidatedCreateFlags represents the new simplified flag structure
type ConsolidatedCreateFlags struct {
	// Basic flags
	Agents      int
	Provider    string
	Template    string
	Interactive bool

	// Feature flags
	Memory    string // "", "memory", "pgvector", "weaviate"
	Embedding string // "provider:model" format
	MCP       string // "", "basic", "production", "full"
	RAG       string // "", "default", or chunk size

	// Orchestration flags
	Orchestration string // "sequential", "collaborative", "loop", "route"
	AgentsConfig  string // JSON or comma-separated config

	// Output flags
	Visualize bool
	OutputDir string
}

// ProjectTemplate represents a predefined project configuration
type ProjectTemplate struct {
	Name        string
	Description string
	Config      scaffold.ProjectConfig
	Features    []string
}

// Global template loader instance
var templateLoader *TemplateLoader

// Initialize template loader
func init() {
	templateLoader = NewTemplateLoader()
}

// normalize maps user-provided flag values to canonical plugin keys and modes.
func (f *ConsolidatedCreateFlags) normalize() {
	// Provider synonyms
	switch strings.ToLower(f.Provider) {
	case "azureopenai", "aoai", "msazure", "azure-openai", "azure_oai":
		f.Provider = "azure"
	case "oai":
		f.Provider = "openai"
	case "local":
		f.Provider = "ollama"
	}

	// Memory provider synonyms
	switch strings.ToLower(f.Memory) {
	case "pg", "postgres", "postgresql", "pg_vec", "pg-vector":
		f.Memory = "pgvector"
	case "mem", "inmemory", "in-memory":
		f.Memory = "memory"
	}

	// Embedding provider synonyms (provider[:model])
	if f.Embedding != "" {
		parts := strings.Split(f.Embedding, ":")
		prov := strings.ToLower(parts[0])
		switch prov {
		case "oai":
			prov = "openai"
		case "local":
			prov = "ollama"
		}
		if len(parts) > 1 {
			f.Embedding = prov + ":" + parts[1]
		} else {
			f.Embedding = prov
		}
	}

	// Orchestration mode synonyms
	switch strings.ToLower(f.Orchestration) {
	case "router", "routing":
		f.Orchestration = "route"
	case "collab", "collaborate":
		f.Orchestration = "collaborative"
	case "seq", "sequential-pipeline":
		f.Orchestration = "sequential"
	case "hybrid":
		f.Orchestration = "mixed"
	}

	// MCP level synonyms
	switch strings.ToLower(f.MCP) {
	case "prod":
		f.MCP = "production"
	}
}

// Available project templates (now loaded dynamically)
var ProjectTemplates = map[string]ProjectTemplate{
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
		Description: "Multi-agent research system with web search, analysis, and synthesis",
		Config: scaffold.ProjectConfig{
			NumAgents:           3,
			Provider:            "openai",
			OrchestrationMode:   "collaborative",
			CollaborativeAgents: []string{"researcher", "analyzer", "synthesizer"},
			MCPEnabled:          true,
			MCPTools:            []string{"web_search", "summarize"},
			ResponsibleAI:       true,
			ErrorHandler:        true,
		},
		Features: []string{"collaborative-agents", "mcp-tools", "web-search"},
	},
	"rag-system": {
		Name:        "RAG Knowledge Base",
		Description: "Document ingestion and Q&A system with vector search",
		Config: scaffold.ProjectConfig{
			NumAgents:         3,
			Provider:          "openai",
			OrchestrationMode: "collaborative",
			MemoryEnabled:     true,
			MemoryProvider:    "pgvector",
			EmbeddingProvider: "openai",
			RAGEnabled:        true,
			RAGChunkSize:      1000,
			RAGOverlap:        100,
			RAGTopK:           5,
			ResponsibleAI:     true,
			ErrorHandler:      true,
		},
		Features: []string{"memory", "rag", "vector-search", "collaborative-agents"},
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

// ParseConsolidatedFlags converts consolidated flags to ProjectConfig
func (f *ConsolidatedCreateFlags) ToProjectConfig(projectName string) (scaffold.ProjectConfig, error) {
	// Normalize inputs first
	f.normalize()
	config := scaffold.ProjectConfig{
		Name:          projectName,
		NumAgents:     f.Agents,
		Provider:      f.Provider,
		ResponsibleAI: true,
		ErrorHandler:  true,
		Visualize:     f.Visualize,
	}

	// Apply template if specified
	if f.Template != "" {
		template, exists := templateLoader.GetTemplate(f.Template)
		if !exists {
			return config, fmt.Errorf("unknown template: %s. Available templates: %s",
				f.Template, strings.Join(getTemplateNames(), ", "))
		}

		// Start with template config
		config = template.Config
		config.Name = projectName // Override name

		// Override with explicit flags
		if f.Agents > 0 {
			config.NumAgents = f.Agents
		}
		if f.Provider != "" {
			config.Provider = f.Provider
		}
		// Always override visualization setting from flag
		config.Visualize = f.Visualize
	}

	// Parse memory flag
	if f.Memory != "" {
		config.MemoryEnabled = true
		config.MemoryProvider = f.Memory

		// Set intelligent defaults based on provider
		switch f.Memory {
		case "pgvector":
			if f.Embedding == "" {
				config.EmbeddingProvider = "openai"
			}
		case "weaviate":
			if f.Embedding == "" {
				config.EmbeddingProvider = "openai"
			}
		case "memory":
			if f.Embedding == "" {
				config.EmbeddingProvider = "dummy"
			}
		}
	}

	// Parse embedding flag (provider:model format)
	if f.Embedding != "" {
		parts := strings.Split(f.Embedding, ":")
		config.EmbeddingProvider = parts[0]
		if len(parts) > 1 {
			config.EmbeddingModel = parts[1]
		} else {
			// Set default model based on provider
			switch parts[0] {
			case "openai":
				config.EmbeddingModel = "text-embedding-ada-002"
			case "ollama":
				config.EmbeddingModel = "nomic-embed-text:latest"
			case "dummy":
				config.EmbeddingModel = "dummy"
			}
		}

		// Auto-enable memory if embedding is specified
		if !config.MemoryEnabled {
			config.MemoryEnabled = true
			config.MemoryProvider = "memory" // Default to in-memory
		}
	}

	// Parse MCP flag
	if f.MCP != "" {
		config.MCPEnabled = true
		switch f.MCP {
		case "basic":
			config.MCPTools = []string{"web_search"}
		case "production":
			config.MCPProduction = true
			config.WithCache = true
			config.WithMetrics = true
			config.MCPTools = []string{"web_search", "summarize"}
		case "full":
			config.MCPProduction = true
			config.WithCache = true
			config.WithMetrics = true
			config.WithLoadBalancer = true
			config.MCPTools = []string{"web_search", "summarize", "translate"}
		default:
			// Treat as custom tool list
			config.MCPTools = strings.Split(f.MCP, ",")
		}
	}

	// Parse RAG flag
	if f.RAG != "" {
		config.RAGEnabled = true

		// Auto-enable memory if not already enabled
		if !config.MemoryEnabled {
			config.MemoryEnabled = true
			config.MemoryProvider = "pgvector" // Default to persistent storage for RAG
			if config.EmbeddingProvider == "" {
				config.EmbeddingProvider = "openai"
			}
		}

		if f.RAG == "default" || f.RAG == "true" {
			config.RAGChunkSize = 1000
			config.RAGOverlap = 100
			config.RAGTopK = 5
			config.RAGScoreThreshold = 0.7
		} else {
			// Parse as chunk size
			if chunkSize, err := strconv.Atoi(f.RAG); err == nil {
				config.RAGChunkSize = chunkSize
				config.RAGOverlap = chunkSize / 10 // 10% overlap
				config.RAGTopK = 5
				config.RAGScoreThreshold = 0.7
			}
		}
	}

	// Parse orchestration flag
	if f.Orchestration != "" {
		config.OrchestrationMode = f.Orchestration

		// Set intelligent defaults based on mode
		switch f.Orchestration {
		case "collaborative":
			if len(config.CollaborativeAgents) == 0 {
				// Generate default agent names
				agents := make([]string, config.NumAgents)
				for i := 0; i < config.NumAgents; i++ {
					agents[i] = fmt.Sprintf("agent%d", i+1)
				}
				config.CollaborativeAgents = agents
			}
		case "sequential":
			if len(config.SequentialAgents) == 0 {
				// Generate default agent names
				agents := make([]string, config.NumAgents)
				for i := 0; i < config.NumAgents; i++ {
					agents[i] = fmt.Sprintf("agent%d", i+1)
				}
				config.SequentialAgents = agents
			}
		case "loop":
			if config.LoopAgent == "" {
				config.LoopAgent = "processor"
			}
			config.NumAgents = 1 // Loop mode uses single agent
		}
	}

	// Set output directory
	if f.OutputDir != "" {
		config.VisualizeOutputDir = f.OutputDir
	} else if config.Visualize {
		// Set default visualization output directory when visualize is enabled
		config.VisualizeOutputDir = "docs/workflows"
	}

	return config, nil
}

// ValidateConsolidatedFlags validates the consolidated flag structure
func (f *ConsolidatedCreateFlags) Validate() error {
	// Validate template
	if f.Template != "" {
		if _, exists := templateLoader.GetTemplate(f.Template); !exists {
			return fmt.Errorf("unknown template: %s. Available: %s",
				f.Template, strings.Join(getTemplateNames(), ", "))
		}
	}

	// Validate provider
	validProviders := []string{"openai", "azure", "ollama", "mock"}
	if f.Provider != "" && !containsString(validProviders, f.Provider) {
		return fmt.Errorf("invalid provider: %s. Valid options: %s",
			f.Provider, strings.Join(validProviders, ", "))
	}

	// Validate memory provider
	if f.Memory != "" {
		validMemoryProviders := []string{"memory", "pgvector", "weaviate"}
		if !containsString(validMemoryProviders, f.Memory) {
			return fmt.Errorf("invalid memory provider: %s. Valid options: %s",
				f.Memory, strings.Join(validMemoryProviders, ", "))
		}
	}

	// Validate embedding provider
	if f.Embedding != "" {
		parts := strings.Split(f.Embedding, ":")
		validEmbeddingProviders := []string{"openai", "ollama", "dummy"}
		if !containsString(validEmbeddingProviders, parts[0]) {
			return fmt.Errorf("invalid embedding provider: %s. Valid options: %s",
				parts[0], strings.Join(validEmbeddingProviders, ", "))
		}
	}

	// Validate MCP level
	if f.MCP != "" {
		validMCPLevels := []string{"basic", "production", "full"}
		if !containsString(validMCPLevels, f.MCP) && !strings.Contains(f.MCP, ",") {
			return fmt.Errorf("invalid MCP level: %s. Valid options: %s or comma-separated tool list",
				f.MCP, strings.Join(validMCPLevels, ", "))
		}
	}

	// Validate orchestration mode
	if f.Orchestration != "" {
		validModes := []string{"sequential", "collaborative", "loop", "route"}
		if !containsString(validModes, f.Orchestration) {
			return fmt.Errorf("invalid orchestration mode: %s. Valid options: %s",
				f.Orchestration, strings.Join(validModes, ", "))
		}
	}

	return nil
}

// Helper functions
func getTemplateNames() []string {
	allTemplates := templateLoader.GetAllTemplates()
	names := make([]string, 0, len(allTemplates))
	for name := range allTemplates {
		names = append(names, name)
	}
	return names
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetTemplateHelp returns help text for available templates
func GetTemplateHelp() string {
	var help strings.Builder
	help.WriteString("Available project templates:\n")

	allTemplates := templateLoader.GetAllTemplates()
	for name, template := range allTemplates {
		help.WriteString(fmt.Sprintf("  %-20s %s\n", name, template.Description))
		help.WriteString(fmt.Sprintf("  %-20s Features: %s\n", "", strings.Join(template.Features, ", ")))
		help.WriteString("\n")
	}

	// Show template search paths
	help.WriteString("\nTemplate search paths:\n")
	for _, path := range templateLoader.ListTemplatePaths() {
		help.WriteString(fmt.Sprintf("  %s\n", path))
	}

	return help.String()
}
