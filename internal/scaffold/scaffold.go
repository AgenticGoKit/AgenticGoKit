package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/kunalkushwaha/agentflow/internal/scaffold/templates"
	"github.com/kunalkushwaha/agentflow/internal/scaffold/utils"
)

// CreateAgentProject creates a new AgentFlow project (alias for CreateAgentProjectModular)
func CreateAgentProject(config ProjectConfig) error {
	return CreateAgentProjectModular(config)
}

// CreateAgentProjectFromConfig creates a new AgentFlow project using ProjectConfig (alias)
func CreateAgentProjectFromConfig(config ProjectConfig) error {
	return CreateAgentProjectModular(config)
}

// CreateAgentProjectModular creates a new AgentFlow project using the modular template system
func CreateAgentProjectModular(config ProjectConfig) error {
	// Create the main project directory
	if err := os.Mkdir(config.Name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory %s: %w", config.Name, err)
	}
	fmt.Printf("Created directory: %s\n", config.Name)

	// Create go.mod file
	if err := createGoMod(config); err != nil {
		return err
	}

	// Create README.md file
	if err := createReadme(config); err != nil {
		return err
	}

	// Create agent files using the new template-based approach
	if err := createAgentFilesWithTemplates(config); err != nil {
		return err
	}

	// Create main.go file
	if err := createMainGoWithTemplate(config); err != nil {
		return err
	}

	// Create agentflow.toml configuration file
	if err := createConfig(config); err != nil {
		return err
	}

	fmt.Printf("\n✅ Project '%s' created successfully using modular templates!\n", config.Name)
	fmt.Printf("📁 Directory: %s\n", config.Name)
	fmt.Printf("🚀 Run: cd %s && go mod tidy && go run . -m \"Your message\"\n", config.Name)

	return nil
}

// createGoMod creates the go.mod file
func createGoMod(config ProjectConfig) error {
	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire github.com/kunalkushwaha/agentflow v0.2.0\n", config.Name)
	goModPath := filepath.Join(config.Name, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}
	fmt.Printf("Created file: %s\n", goModPath)
	return nil
}

// createReadme creates the README.md file
func createReadme(config ProjectConfig) error {
	content := fmt.Sprintf("# %s\n\n", config.Name)
	content += "This AgentFlow project was generated using the AgentFlow CLI with modular templates.\n\n"

	content += "## Project Configuration\n\n"
	content += fmt.Sprintf("- **Orchestration Mode**: %s\n", config.OrchestrationMode)
	content += fmt.Sprintf("- **LLM Provider**: %s\n", config.Provider)

	if config.MCPEnabled {
		content += "- **MCP Integration**: Enabled\n"
	}

	content += "\n## Agents\n\n"
	agents := utils.ResolveAgentNames(convertToUtilsConfig(config))
	for _, agent := range agents {
		content += fmt.Sprintf("- **%s** (%s): %s\n", agent.DisplayName, agent.Name, agent.Purpose)
	}

	content += "\n## Usage\n\n"
	content += "```bash\n"
	content += "go mod tidy\n"
	content += "go run . -m \"Your message here\"\n"
	content += "```\n"

	readmePath := filepath.Join(config.Name, "README.md")
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	fmt.Printf("Created file: %s\n", readmePath)
	return nil
}

// createAgentFilesWithTemplates creates agent files using the template system
func createAgentFilesWithTemplates(config ProjectConfig) error {
	utilsConfig := convertToUtilsConfig(config)
	agents := utils.ResolveAgentNames(utilsConfig)

	if len(agents) == 0 {
		agents = append(agents, utils.CreateAgentInfo("agent1", config.OrchestrationMode))
	}

	// Validate agent names
	if err := utils.ValidateAgentNames(agents); err != nil {
		return fmt.Errorf("agent name validation failed: %w", err)
	}

	// Use the comprehensive template from templates package
	tmpl, err := template.New("agent").Parse(templates.AgentTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse agent template: %w", err)
	}

	// Generate each agent file
	for i, agent := range agents {
		var nextAgent string
		var routingComment string

		if i < len(agents)-1 {
			// Route to next agent in the list
			nextAgent = agents[i+1].Name
			routingComment = fmt.Sprintf("Route to the next agent (%s) in the workflow", agents[i+1].DisplayName)
		} else if config.ResponsibleAI {
			// Last agent routes to responsible AI
			nextAgent = "responsible_ai"
			routingComment = "Route to Responsible AI for final content check"
		} else {
			// Route to workflow finalizer to complete the workflow
			nextAgent = "workflow_finalizer"
			routingComment = "Route to workflow finalizer to complete the workflow"
		}

		// Create system prompt for this agent
		systemPrompt := utils.CreateSystemPrompt(agent, i, len(agents), config.OrchestrationMode)

		// Create template data structure that matches the comprehensive template
		templateData := struct {
			Agent          utils.AgentInfo
			Agents         []utils.AgentInfo
			AgentIndex     int
			TotalAgents    int
			NextAgent      string
			PrevAgent      string
			IsFirstAgent   bool
			IsLastAgent    bool
			SystemPrompt   string
			RoutingComment string
		}{
			Agent:          agent,
			Agents:         agents,
			AgentIndex:     i,
			TotalAgents:    len(agents),
			NextAgent:      nextAgent,
			IsFirstAgent:   i == 0,
			IsLastAgent:    i == len(agents)-1,
			SystemPrompt:   systemPrompt,
			RoutingComment: routingComment,
		}

		filePath := filepath.Join(config.Name, agent.FileName)
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}

		if err := tmpl.Execute(file, templateData); err != nil {
			file.Close()
			return fmt.Errorf("failed to execute template for %s: %w", agent.FileName, err)
		}
		file.Close()

		fmt.Printf("Created file: %s (%s agent)\n", filePath, agent.DisplayName)
	}

	return nil
}

// createMainGoWithTemplate creates main.go using templates
func createMainGoWithTemplate(config ProjectConfig) error {
	utilsConfig := convertToUtilsConfig(config)
	agents := utils.ResolveAgentNames(utilsConfig)

	// Simple main.go template
	mainTemplate := `package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kunalkushwaha/agentflow/core"
)

func main() {
	ctx := context.Background()
	core.SetLogLevel(core.INFO)
	logger := core.Logger()
	logger.Info().Msg("Starting {{.Name}} multi-agent system...")

	messageFlag := flag.String("m", "", "Message to process")
	flag.Parse()

	llmProvider, err := initializeProvider("{{.Provider}}")
	if err != nil {
		fmt.Printf("Failed to initialize LLM provider: %v\n", err)
		os.Exit(1)
	}

	agents := make(map[string]core.AgentHandler)
	{{range .Agents}}
	{{.Name}} := New{{.DisplayName}}(llmProvider)
	agents["{{.Name}}"] = {{.Name}}
	{{end}}

	// Create a simple runner configuration  
	runner := core.NewRunnerWithConfig(core.RunnerConfig{
		Agents: agents,
	})

	var message string
	if *messageFlag != "" {
		message = *messageFlag
	} else {
		fmt.Print("Enter your message: ")
		fmt.Scanln(&message)
	}

	if message == "" {
		message = "Hello! Please provide information about current topics."
	}

	// Start the runner
	runner.Start(ctx)
	defer runner.Stop()

	event := core.NewEvent("{{range $i, $agent := .Agents}}{{if eq $i 0}}{{$agent.Name}}{{end}}{{end}}", core.EventData{
		"message": message,
	}, map[string]string{})

	if err := runner.Emit(event); err != nil {
		logger.Error().Err(err).Msg("Workflow execution failed")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Simple wait for completion (in a real app you'd want better synchronization)
	time.Sleep(5 * time.Second)
	
	fmt.Printf("\n=== Workflow Completed ===\n")
	fmt.Printf("Event processed successfully\n")
}

func initializeProvider(providerType string) (core.ModelProvider, error) {
	// Use the config-based provider initialization
	return core.NewProviderFromWorkingDir()
}
`

	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse main template: %w", err)
	}

	templateData := struct {
		ProjectConfig
		Agents []utils.AgentInfo
	}{
		ProjectConfig: config,
		Agents:        agents,
	}

	mainGoPath := filepath.Join(config.Name, "main.go")
	file, err := os.Create(mainGoPath)
	if err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute main template: %w", err)
	}

	fmt.Printf("Created file: %s\n", mainGoPath)
	return nil
}

// createConfig creates the agentflow.toml configuration file
func createConfig(config ProjectConfig) error {
	configContent := fmt.Sprintf(`# AgentFlow Configuration

[agent_flow]
name = "%s"
version = "1.0.0"
provider = "%s"

[logging]
level = "info"
format = "json"

[runtime]
max_concurrent_agents = 10
timeout_seconds = 30

[providers.azure]
# API key will be read from AZURE_OPENAI_API_KEY environment variable
# Endpoint will be read from AZURE_OPENAI_ENDPOINT environment variable
# Deployment will be read from AZURE_OPENAI_DEPLOYMENT environment variable

[providers.openai]
# API key will be read from OPENAI_API_KEY environment variable

[providers.ollama]
endpoint = "http://localhost:11434"
model = "llama2"

[providers.mock]
# Mock provider for testing - no configuration needed
`, config.Name, config.Provider)

	configPath := filepath.Join(config.Name, "agentflow.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create agentflow.toml: %w", err)
	}
	fmt.Printf("Created file: %s\n", configPath)
	return nil
}

// convertToUtilsConfig converts ProjectConfig to utils.ProjectConfig
func convertToUtilsConfig(config ProjectConfig) utils.ProjectConfig {
	return utils.ProjectConfig{
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
	}
}
