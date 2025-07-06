package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/kunalkushwaha/agentflow/internal/scaffold"
	"github.com/kunalkushwaha/agentflow/internal/scaffold/templates"
	"github.com/kunalkushwaha/agentflow/internal/scaffold/utils"
)

// ProjectGenerator handles the generation of project structure and main files
type ProjectGenerator struct {
	mainTemplate *template.Template
}

// NewProjectGenerator creates a new project generator
func NewProjectGenerator() (*ProjectGenerator, error) {
	mainTmpl, err := template.New("main").Parse(templates.MainTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse main template: %w", err)
	}

	return &ProjectGenerator{mainTemplate: mainTmpl}, nil
}

// GenerateProjectStructure creates the basic project structure
func (g *ProjectGenerator) GenerateProjectStructure(config scaffold.ProjectConfig) error {
	// Create the main project directory
	if err := os.Mkdir(config.Name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory %s: %w", config.Name, err)
	}
	fmt.Printf("Created directory: %s\n", config.Name)

	// Create go.mod file
	if err := g.generateGoMod(config); err != nil {
		return err
	}

	// Create README.md file
	if err := g.generateReadme(config); err != nil {
		return err
	}

	// Create main.go file
	if err := g.generateMainFile(config); err != nil {
		return err
	}

	// Generate workflow diagrams if requested
	if config.Visualize {
		if err := g.generateWorkflowDiagrams(config); err != nil {
			return err
		}
	}

	return nil
}

// generateGoMod creates the go.mod file
func (g *ProjectGenerator) generateGoMod(config scaffold.ProjectConfig) error {
	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire github.com/kunalkushwaha/agentflow %s\n", config.Name, scaffold.AgentFlowVersion)
	goModPath := filepath.Join(config.Name, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}
	fmt.Printf("Created file: %s\n", goModPath)
	return nil
}

// generateReadme creates the README.md file
func (g *ProjectGenerator) generateReadme(config scaffold.ProjectConfig) error {
	readmeContent := g.createReadmeContent(config)
	readmePath := filepath.Join(config.Name, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	fmt.Printf("Created file: %s\n", readmePath)
	return nil
}

// generateMainFile creates the main.go file using templates
func (g *ProjectGenerator) generateMainFile(config scaffold.ProjectConfig) error {
	// Convert to utils.ProjectConfig
	utilsConfig := utils.ProjectConfig{
		Name:                 config.Name,
		NumAgents:            config.NumAgents,
		Provider:             config.Provider,
		ResponsibleAI:        config.ResponsibleAI,
		ErrorHandler:         config.ErrorHandler,
		MCPEnabled:           config.MCPEnabled,
		MCPProduction:        config.MCPProduction,
		WithCache:            config.WithCache,
		WithMetrics:          config.WithMetrics,
		MCPTools:             config.MCPTools,
		MCPServers:           config.MCPServers,
		CacheBackend:         config.CacheBackend,
		MetricsPort:          config.MetricsPort,
		WithLoadBalancer:     config.WithLoadBalancer,
		ConnectionPoolSize:   config.ConnectionPoolSize,
		RetryPolicy:          config.RetryPolicy,
		OrchestrationMode:    config.OrchestrationMode,
		CollaborativeAgents:  config.CollaborativeAgents,
		SequentialAgents:     config.SequentialAgents,
		LoopAgent:            config.LoopAgent,
		MaxIterations:        config.MaxIterations,
		OrchestrationTimeout: config.OrchestrationTimeout,
		FailureThreshold:     config.FailureThreshold,
		MaxConcurrency:       config.MaxConcurrency,
		Visualize:            config.Visualize,
		VisualizeOutputDir:   config.VisualizeOutputDir,
	}

	agents := utils.ResolveAgentNames(utilsConfig)

	// Create template data
	templateData := struct {
		Config               scaffold.ProjectConfig
		Agents               []utils.AgentInfo
		ProviderInitFunction string
		MCPInitFunction      string
		CacheInitFunction    string
	}{
		Config:               config,
		Agents:               agents,
		ProviderInitFunction: g.createProviderInitFunction(config.Provider),
		MCPInitFunction:      g.createMCPInitFunction(config),
		CacheInitFunction:    g.createCacheInitFunction(config),
	}

	// Generate file using template
	mainGoPath := filepath.Join(config.Name, "main.go")
	file, err := os.Create(mainGoPath)
	if err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}
	defer file.Close()

	if err := g.mainTemplate.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute main template: %w", err)
	}

	fmt.Printf("Created file: %s\n", mainGoPath)
	return nil
}

// createReadmeContent generates README content
func (g *ProjectGenerator) createReadmeContent(config scaffold.ProjectConfig) string {
	content := fmt.Sprintf("# %s\n\n", config.Name)
	content += "This AgentFlow project was generated using the AgentFlow CLI.\n\n"

	content += "## Project Configuration\n\n"
	content += fmt.Sprintf("- **Orchestration Mode**: %s\n", config.OrchestrationMode)
	content += fmt.Sprintf("- **LLM Provider**: %s\n", config.Provider)

	if config.MCPEnabled {
		content += "- **MCP Integration**: Enabled\n"
		if config.MCPProduction {
			content += "- **MCP Production Features**: Enabled\n"
		}
	}

	content += "\n## Agents\n\n"
	// Convert to utils.ProjectConfig
	utilsConfig := utils.ProjectConfig{
		Name:                 config.Name,
		NumAgents:            config.NumAgents,
		Provider:             config.Provider,
		ResponsibleAI:        config.ResponsibleAI,
		ErrorHandler:         config.ErrorHandler,
		MCPEnabled:           config.MCPEnabled,
		MCPProduction:        config.MCPProduction,
		WithCache:            config.WithCache,
		WithMetrics:          config.WithMetrics,
		MCPTools:             config.MCPTools,
		MCPServers:           config.MCPServers,
		CacheBackend:         config.CacheBackend,
		MetricsPort:          config.MetricsPort,
		WithLoadBalancer:     config.WithLoadBalancer,
		ConnectionPoolSize:   config.ConnectionPoolSize,
		RetryPolicy:          config.RetryPolicy,
		OrchestrationMode:    config.OrchestrationMode,
		CollaborativeAgents:  config.CollaborativeAgents,
		SequentialAgents:     config.SequentialAgents,
		LoopAgent:            config.LoopAgent,
		MaxIterations:        config.MaxIterations,
		OrchestrationTimeout: config.OrchestrationTimeout,
		FailureThreshold:     config.FailureThreshold,
		MaxConcurrency:       config.MaxConcurrency,
		Visualize:            config.Visualize,
		VisualizeOutputDir:   config.VisualizeOutputDir,
	}
	agents := utils.ResolveAgentNames(utilsConfig)
	for _, agent := range agents {
		content += fmt.Sprintf("- **%s** (%s): %s\n", agent.DisplayName, agent.Name, agent.Purpose)
	}

	content += "\n## Usage\n\n"
	content += "```bash\n"
	content += "# Install dependencies\n"
	content += "go mod tidy\n\n"
	content += "# Run the application\n"
	content += "go run . -m \"Your message here\"\n"
	content += "```\n\n"

	if config.Visualize {
		content += "## Workflow Diagrams\n\n"
		content += fmt.Sprintf("Visual workflow diagrams have been generated in the `%s` directory. These diagrams show the orchestration pattern and agent interactions for this project.\n\n", config.VisualizeOutputDir)
	}

	content += "## Configuration\n\n"
	content += "Set the following environment variables:\n\n"

	switch config.Provider {
	case "azure":
		content += "```bash\n"
		content += "export AZURE_OPENAI_API_KEY=\"your-api-key\"\n"
		content += "export AZURE_OPENAI_ENDPOINT=\"your-endpoint\"\n"
		content += "export AZURE_OPENAI_DEPLOYMENT=\"your-deployment\"\n"
		content += "```\n"
	case "openai":
		content += "```bash\n"
		content += "export OPENAI_API_KEY=\"your-api-key\"\n"
		content += "```\n"
	case "ollama":
		content += "```bash\n"
		content += "export OLLAMA_HOST=\"http://localhost:11434\"\n"
		content += "```\n"
	}

	return content
}

// createProviderInitFunction creates the provider initialization function
func (g *ProjectGenerator) createProviderInitFunction(provider string) string {
	return fmt.Sprintf(`
// initializeProvider creates and configures the LLM provider
func initializeProvider(providerType string) (core.ModelProvider, error) {
	switch providerType {
	case "azure":
		return core.NewAzureOpenAIProvider()
	case "openai":
		return core.NewOpenAIProvider()
	case "ollama":
		return core.NewOllamaProvider()
	case "mock":
		return core.NewMockProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %%s", providerType)
	}
}`)
}

// createMCPInitFunction creates the MCP initialization function
func (g *ProjectGenerator) createMCPInitFunction(config scaffold.ProjectConfig) string {
	if !config.MCPEnabled {
		return ""
	}

	content := `
// initializeMCP initializes the MCP manager with configured tools and servers
func initializeMCP() (*core.MCPManager, error) {
	mcpManager := core.NewMCPManager()
	
	// Configure MCP servers`

	if len(config.MCPServers) > 0 {
		content += "\n\tservers := []string{"
		for i, server := range config.MCPServers {
			if i > 0 {
				content += ", "
			}
			content += fmt.Sprintf("\"%s\"", server)
		}
		content += "}\n"
		content += "\tfor _, server := range servers {\n"
		content += "\t\tif err := mcpManager.AddServer(server); err != nil {\n"
		content += "\t\t\treturn nil, fmt.Errorf(\"failed to add MCP server %s: %w\", server, err)\n"
		content += "\t\t}\n"
		content += "\t}\n"
	}

	if config.MCPProduction {
		content += "\n\t// Configure production features\n"
		content += fmt.Sprintf("\tmcpManager.SetConnectionPoolSize(%d)\n", config.ConnectionPoolSize)
		content += fmt.Sprintf("\tmcpManager.SetRetryPolicy(\"%s\")\n", config.RetryPolicy)

		if config.WithLoadBalancer {
			content += "\tmcpManager.EnableLoadBalancing(true)\n"
		}
	}

	content += "\n\t return mcpManager, nil\n}"

	return content
}

// createCacheInitFunction creates the cache initialization function
func (g *ProjectGenerator) createCacheInitFunction(config scaffold.ProjectConfig) string {
	if !config.WithCache {
		return ""
	}

	return fmt.Sprintf(`
// initializeCache sets up the caching system
func initializeCache() error {
	switch "%s" {
	case "redis":
		return core.InitializeRedisCache()
	case "memory":
		return core.InitializeMemoryCache()
	default:
		return fmt.Errorf("unsupported cache backend: %s")
	}
}`, config.CacheBackend, config.CacheBackend)
}

// generateWorkflowDiagrams creates Mermaid workflow diagrams for the project
func (g *ProjectGenerator) generateWorkflowDiagrams(config scaffold.ProjectConfig) error {
	// Create the output directory for diagrams
	diagramsDir := filepath.Join(config.Name, config.VisualizeOutputDir)
	if err := os.MkdirAll(diagramsDir, 0755); err != nil {
		return fmt.Errorf("failed to create diagrams directory: %w", err)
	}
	fmt.Printf("Created directory: %s\n", diagramsDir)

	// Generate diagram based on orchestration mode
	var diagram string
	var title string

	switch config.OrchestrationMode {
	case "collaborative":
		diagram, title = g.generateCollaborativeDiagram(config)
	case "sequential":
		diagram, title = g.generateSequentialDiagram(config)
	case "loop":
		diagram, title = g.generateLoopDiagram(config)
	case "mixed":
		diagram, title = g.generateMixedDiagram(config)
	default:
		diagram, title = g.generateRouteDiagram(config)
	}

	// Create the diagram file
	diagramPath := filepath.Join(diagramsDir, "workflow.md")
	content := fmt.Sprintf(`# %s Workflow

## Overview
This diagram shows the %s orchestration pattern used in this project.

## Workflow Diagram

%s

## Configuration
- **Orchestration Mode**: %s
- **Number of Agents**: %d
- **Timeout**: %d seconds
- **Max Concurrency**: %d
- **Failure Threshold**: %.2f

## Agent Details
%s
`, title, config.OrchestrationMode, diagram, config.OrchestrationMode, config.NumAgents, config.OrchestrationTimeout, config.MaxConcurrency, config.FailureThreshold, g.generateAgentDetails(config))

	if err := os.WriteFile(diagramPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create workflow diagram: %w", err)
	}
	fmt.Printf("Created workflow diagram: %s\n", diagramPath)

	return nil
}

// generateCollaborativeDiagram creates a collaborative orchestration diagram
func (g *ProjectGenerator) generateCollaborativeDiagram(config scaffold.ProjectConfig) (string, string) {
	agents := config.CollaborativeAgents
	if len(agents) == 0 {
		// Use default agent names if none specified
		agents = make([]string, config.NumAgents)
		for i := 0; i < config.NumAgents; i++ {
			agents[i] = fmt.Sprintf("agent%d", i+1)
		}
	}

	diagram := "```mermaid\n---\ntitle: Collaborative Orchestration\n---\nflowchart TD\n"
	diagram += "    EVENT[\"📨 Input Event\"]\n"
	diagram += "    ORCHESTRATOR[\"🎯 Collaborative Orchestrator\"]\n"
	diagram += "    AGGREGATOR[\"📊 Result Aggregator\"]\n"
	diagram += "    RESULT[\"📤 Final Result\"]\n\n"
	diagram += "    EVENT --> ORCHESTRATOR\n"

	for i, agent := range agents {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		diagram += fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agent)
		diagram += fmt.Sprintf("    ORCHESTRATOR --> %s\n", agentId)
		diagram += fmt.Sprintf("    %s --> AGGREGATOR\n", agentId)
	}

	diagram += "    AGGREGATOR --> RESULT\n"
	diagram += "```"

	return diagram, "Collaborative"
}

// generateSequentialDiagram creates a sequential orchestration diagram
func (g *ProjectGenerator) generateSequentialDiagram(config scaffold.ProjectConfig) (string, string) {
	agents := config.SequentialAgents
	if len(agents) == 0 {
		// Use default agent names if none specified
		agents = make([]string, config.NumAgents)
		for i := 0; i < config.NumAgents; i++ {
			agents[i] = fmt.Sprintf("agent%d", i+1)
		}
	}

	diagram := "```mermaid\n---\ntitle: Sequential Pipeline\n---\nflowchart TD\n"
	diagram += "    INPUT[\"📨 Input Event\"]\n"

	var prevNode = "INPUT"
	for i, agent := range agents {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		diagram += fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agent)
		diagram += fmt.Sprintf("    %s --> %s\n", prevNode, agentId)
		prevNode = agentId
	}

	diagram += "    OUTPUT[\"📤 Final Result\"]\n"
	diagram += fmt.Sprintf("    %s --> OUTPUT\n", prevNode)
	diagram += "```"

	return diagram, "Sequential Pipeline"
}

// generateLoopDiagram creates a loop orchestration diagram
func (g *ProjectGenerator) generateLoopDiagram(config scaffold.ProjectConfig) (string, string) {
	agentName := config.LoopAgent
	if agentName == "" {
		agentName = "processor"
	}

	diagram := "```mermaid\n---\ntitle: Loop Processing\n---\nflowchart TD\n"
	diagram += "    INPUT[\"📨 Input Event\"]\n"
	diagram += "    AGENT[\"🤖 " + agentName + "\"]\n"
	diagram += "    CONDITION{\"🔄 Continue Loop?\"}\n"
	diagram += "    OUTPUT[\"📤 Final Result\"]\n\n"
	diagram += "    INPUT --> AGENT\n"
	diagram += "    AGENT --> CONDITION\n"
	diagram += "    CONDITION -->|Yes| AGENT\n"
	diagram += "    CONDITION -->|No| OUTPUT\n"
	diagram += fmt.Sprintf("    CONDITION -.->|Max %d iterations| OUTPUT\n", config.MaxIterations)
	diagram += "```"

	return diagram, "Loop Processing"
}

// generateMixedDiagram creates a mixed orchestration diagram
func (g *ProjectGenerator) generateMixedDiagram(config scaffold.ProjectConfig) (string, string) {
	diagram := "```mermaid\n---\ntitle: Mixed Orchestration\n---\nflowchart TD\n"
	diagram += "    INPUT[\"📨 Input Event\"]\n"
	diagram += "    PHASE1[\"🤝 Collaborative Phase\"]\n"
	diagram += "    PHASE2[\"🎭 Sequential Phase\"]\n"
	diagram += "    OUTPUT[\"📤 Final Result\"]\n\n"
	diagram += "    INPUT --> PHASE1\n"

	// Add collaborative agents
	if len(config.CollaborativeAgents) > 0 {
		for i, agent := range config.CollaborativeAgents {
			agentId := fmt.Sprintf("COLLAB%d", i+1)
			diagram += fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agent)
			diagram += fmt.Sprintf("    PHASE1 --> %s\n", agentId)
			diagram += fmt.Sprintf("    %s --> PHASE2\n", agentId)
		}
	}

	// Add sequential agents
	if len(config.SequentialAgents) > 0 {
		var prevNode = "PHASE2"
		for i, agent := range config.SequentialAgents {
			agentId := fmt.Sprintf("SEQ%d", i+1)
			diagram += fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agent)
			diagram += fmt.Sprintf("    %s --> %s\n", prevNode, agentId)
			prevNode = agentId
		}
		diagram += fmt.Sprintf("    %s --> OUTPUT\n", prevNode)
	} else {
		diagram += "    PHASE2 --> OUTPUT\n"
	}

	diagram += "```"

	return diagram, "Mixed Orchestration"
}

// generateRouteDiagram creates a route orchestration diagram
func (g *ProjectGenerator) generateRouteDiagram(config scaffold.ProjectConfig) (string, string) {
	diagram := "```mermaid\n---\ntitle: Route Orchestration\n---\nflowchart TD\n"
	diagram += "    INPUT[\"📨 Input Event\"]\n"
	diagram += "    ROUTER[\"🎯 Event Router\"]\n"
	diagram += "    OUTPUT[\"📤 Result\"]\n\n"
	diagram += "    INPUT --> ROUTER\n"

	for i := 0; i < config.NumAgents; i++ {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		agentName := fmt.Sprintf("agent%d", i+1)
		diagram += fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agentName)
		diagram += fmt.Sprintf("    ROUTER -.->|Route| %s\n", agentId)
		diagram += fmt.Sprintf("    %s --> OUTPUT\n", agentId)
	}

	diagram += "```"

	return diagram, "Route Orchestration"
}

// generateAgentDetails creates detailed agent information
func (g *ProjectGenerator) generateAgentDetails(config scaffold.ProjectConfig) string {
	details := ""

	switch config.OrchestrationMode {
	case "collaborative":
		if len(config.CollaborativeAgents) > 0 {
			details += "### Collaborative Agents\n"
			for i, agent := range config.CollaborativeAgents {
				details += fmt.Sprintf("%d. **%s**: Processes events in parallel with other agents\n", i+1, agent)
			}
		}
	case "sequential":
		if len(config.SequentialAgents) > 0 {
			details += "### Sequential Agents\n"
			for i, agent := range config.SequentialAgents {
				details += fmt.Sprintf("%d. **%s**: Processes events in pipeline order\n", i+1, agent)
			}
		}
	case "loop":
		if config.LoopAgent != "" {
			details += "### Loop Agent\n"
			details += fmt.Sprintf("1. **%s**: Processes events iteratively up to %d times\n", config.LoopAgent, config.MaxIterations)
		}
	case "mixed":
		if len(config.CollaborativeAgents) > 0 {
			details += "### Collaborative Agents (Phase 1)\n"
			for i, agent := range config.CollaborativeAgents {
				details += fmt.Sprintf("%d. **%s**: Processes events in parallel\n", i+1, agent)
			}
		}
		if len(config.SequentialAgents) > 0 {
			details += "\n### Sequential Agents (Phase 2)\n"
			for i, agent := range config.SequentialAgents {
				details += fmt.Sprintf("%d. **%s**: Processes events in pipeline order\n", i+1, agent)
			}
		}
	default:
		details += "### Route Agents\n"
		for i := 0; i < config.NumAgents; i++ {
			details += fmt.Sprintf("%d. **agent%d**: Processes events based on routing logic\n", i+1, i+1)
		}
	}

	return details
}
