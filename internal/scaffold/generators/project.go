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

	return nil
}

// generateGoMod creates the go.mod file
func (g *ProjectGenerator) generateGoMod(config scaffold.ProjectConfig) error {
	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire github.com/kunalkushwaha/agentflow v0.2.0\n", config.Name)
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
	agents := utils.ResolveAgentNames(config)

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
	agents := utils.ResolveAgentNames(config)
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

	content += "\n\treturn mcpManager, nil\n}"

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
