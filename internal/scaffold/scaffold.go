package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/kunalkushwaha/agenticgokit/internal/scaffold/templates"
	"github.com/kunalkushwaha/agenticgokit/internal/scaffold/utils"
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
	// Validate import paths before creating the project
	if err := ValidateImportPaths(config); err != nil {
		return fmt.Errorf("import path validation failed: %w", err)
	}

	// Create the main project directory
	if err := os.Mkdir(config.Name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory %s: %w", config.Name, err)
	}
	fmt.Printf("Created directory: %s\n", config.Name)

	// Create project subdirectories
	if err := createProjectDirectories(config); err != nil {
		return err
	}

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

	// Create agents README documentation
	if err := createAgentsReadme(config); err != nil {
		return err
	}

	// Create customization guide documentation
	if err := createCustomizationGuide(config); err != nil {
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

	// Generate Docker Compose files for database providers
	if config.MemoryEnabled && (config.MemoryProvider == "pgvector" || config.MemoryProvider == "weaviate") {
		dockerGenerator := NewDockerComposeGenerator(config)
		if err := dockerGenerator.GenerateDockerCompose(); err != nil {
			return fmt.Errorf("failed to generate Docker Compose files: %w", err)
		}
		if err := dockerGenerator.GenerateSetupScript(); err != nil {
			return fmt.Errorf("failed to generate setup scripts: %w", err)
		}
	}

	// Generate workflow diagrams if requested
	if config.Visualize {
		if err := generateWorkflowDiagrams(config); err != nil {
			return err
		}
	}

	fmt.Printf("\n✅ Project '%s' created successfully using modular templates!\n", config.Name)
	fmt.Printf("📁 Directory: %s\n", config.Name)
	fmt.Printf("🚀 Run: cd %s && go mod tidy && go run . -m \"Your message\"\n", config.Name)

	return nil
}

// createGoMod creates the go.mod file
func createGoMod(config ProjectConfig) error {
	// Create go.mod content with proper module declaration
	goModContent := fmt.Sprintf(`module %s

go 1.21

require github.com/kunalkushwaha/agenticgokit %s
`, config.Name, AgenticGoKitVersion)

	goModPath := filepath.Join(config.Name, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}
	fmt.Printf("Created file: %s\n", goModPath)
	return nil
}

// createReadme creates the README.md file using the enhanced template
func createReadme(config ProjectConfig) error {
	utilsConfig := convertToUtilsConfig(config)
	agents := utils.ResolveAgentNames(utilsConfig)

	if len(agents) == 0 {
		agents = append(agents, utils.CreateAgentInfo("agent1", config.OrchestrationMode))
	}

	// Use the comprehensive template from templates package
	tmpl, err := template.New("readme").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).Parse(templates.ProjectReadmeTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse project README template: %w", err)
	}

	// Create template data structure with enhanced agent info
	type EnhancedAgentInfo struct {
		utils.AgentInfo
		IsFirstAgent bool
		IsLastAgent  bool
		Index        int
	}

	var enhancedAgents []EnhancedAgentInfo
	for i, agent := range agents {
		enhancedAgents = append(enhancedAgents, EnhancedAgentInfo{
			AgentInfo:    agent,
			IsFirstAgent: i == 0,
			IsLastAgent:  i == len(agents)-1,
			Index:        i,
		})
	}

	templateData := struct {
		Config              ProjectConfig
		Agents              []EnhancedAgentInfo
		ProjectStructure    ProjectStructureInfo
		CustomizationPoints []CustomizationPoint
		ImportPaths         ImportPathInfo
		Features            []string
		FrameworkVersion    string
	}{
		Config:              config,
		Agents:              enhancedAgents,
		ProjectStructure:    CreateProjectStructureInfo(config),
		CustomizationPoints: CreateCustomizationPoints(config),
		ImportPaths:         CreateImportPathInfo(config),
		Features:            GetEnabledFeatures(config),
		FrameworkVersion:    AgenticGoKitVersion,
	}

	readmePath := filepath.Join(config.Name, "README.md")
	file, err := os.Create(readmePath)
	if err != nil {
		return fmt.Errorf("failed to create README.md file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute README template: %w", err)
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
			// Last agent - workflow completion (no routing)
			nextAgent = ""
			routingComment = "Workflow completion"
		}

		// Create system prompt for this agent
		systemPrompt := utils.CreateSystemPrompt(agent, i, len(agents), config.OrchestrationMode)

		// Create enhanced template data structure
		var prevAgent string
		if i > 0 {
			prevAgent = agents[i-1].Name
		}

		templateData := struct {
			Config              ProjectConfig
			Agent               utils.AgentInfo
			Agents              []utils.AgentInfo
			AgentIndex          int
			TotalAgents         int
			NextAgent           string
			PrevAgent           string
			IsFirstAgent        bool
			IsLastAgent         bool
			SystemPrompt        string
			RoutingComment      string
			ProjectStructure    ProjectStructureInfo
			CustomizationPoints []CustomizationPoint
			ImportPaths         ImportPathInfo
			Features            []string
			FrameworkVersion    string
		}{
			Config:              config,
			Agent:               agent,
			Agents:              agents,
			AgentIndex:          i,
			TotalAgents:         len(agents),
			NextAgent:           nextAgent,
			PrevAgent:           prevAgent,
			IsFirstAgent:        i == 0,
			IsLastAgent:         i == len(agents)-1,
			SystemPrompt:        systemPrompt,
			RoutingComment:      routingComment,
			ProjectStructure:    CreateProjectStructureInfo(config),
			CustomizationPoints: CreateCustomizationPoints(config),
			ImportPaths:         CreateImportPathInfo(config),
			Features:            GetEnabledFeatures(config),
			FrameworkVersion:    AgenticGoKitVersion,
		}

		filePath := filepath.Join(config.Name, "agents", agent.FileName)
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

	// Use the comprehensive template from templates package
	tmpl, err := template.New("main").Parse(templates.MainTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse main template: %w", err)
	}

	// Create enhanced template data structure
	templateData := struct {
		Config               ProjectConfig
		Agents               []utils.AgentInfo
		ProviderInitFunction string
		MCPInitFunction      string
		CacheInitFunction    string
		ProjectStructure     ProjectStructureInfo
		CustomizationPoints  []CustomizationPoint
		ImportPaths          ImportPathInfo
		Features             []string
		FrameworkVersion     string
	}{
		Config:               config,
		Agents:               agents,
		ProviderInitFunction: generateProviderInitFunction(config),
		MCPInitFunction:      generateMCPInitFunction(config),
		CacheInitFunction:    generateCacheInitFunction(config),
		ProjectStructure:     CreateProjectStructureInfo(config),
		CustomizationPoints:  CreateCustomizationPoints(config),
		ImportPaths:          CreateImportPathInfo(config),
		Features:             GetEnabledFeatures(config),
		FrameworkVersion:     AgenticGoKitVersion,
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
base_url = "http://localhost:11434"
model = "llama2"

[providers.mock]
# Mock provider for testing - no configuration needed
`, config.Name, config.Provider)

	// Add MCP configuration if enabled
	if config.MCPEnabled {
		mcpConfig := `
[mcp]
enabled = true
enable_discovery = true
connection_timeout = 5000
max_retries = 3
retry_delay = 1000
enable_caching = true
cache_timeout = 300000
max_connections = 10

# Example MCP servers - configure as needed
[[mcp.servers]]
name = "docker"
type = "tcp"
host = "localhost"
port = 8811
enabled = false

[[mcp.servers]]
name = "filesystem"
type = "stdio"
command = "npx @modelcontextprotocol/server-filesystem /path/to/allowed/files"
enabled = false

[[mcp.servers]]
name = "brave-search"
type = "stdio"
command = "npx @modelcontextprotocol/server-brave-search"
enabled = false
`
		configContent += mcpConfig
	}

	// Add memory configuration if enabled
	if config.MemoryEnabled {
		// Get embedding dimensions from intelligence system
		dimensions := GetModelDimensions(config.EmbeddingProvider, config.EmbeddingModel)
		
		memoryConfig := fmt.Sprintf(`
[agent_memory]
provider = "%s"
connection = "%s"
max_results = %d
dimensions = %d
auto_embed = true
enable_knowledge_base = true
knowledge_max_results = %d
knowledge_score_threshold = %.1f
chunk_size = %d
chunk_overlap = %d
enable_rag = %t
rag_max_context_tokens = 4000
rag_personal_weight = 0.3
rag_knowledge_weight = 0.7
rag_include_sources = true

[agent_memory.embedding]
provider = "%s"
model = "%s"`,
			config.MemoryProvider,
			getConnectionString(config.MemoryProvider),
			config.RAGTopK,
			dimensions,
			config.RAGTopK,
			config.RAGScoreThreshold,
			config.RAGChunkSize,
			config.RAGOverlap,
			config.RAGEnabled,
			config.EmbeddingProvider,
			config.EmbeddingModel)

		// Add provider-specific embedding configuration
		if config.EmbeddingProvider == "ollama" {
			memoryConfig += `
base_url = "http://localhost:11434"`
		}
		
		memoryConfig += `
cache_embeddings = true
max_batch_size = 100
timeout_seconds = 30

[agent_memory.documents]
auto_chunk = true
supported_types = ["pdf", "txt", "md", "web", "code"]
max_file_size = "10MB"
enable_metadata_extraction = true
enable_url_scraping = true

[agent_memory.search]
hybrid_search = ` + fmt.Sprintf("%t", config.HybridSearch) + `
keyword_weight = 0.3
semantic_weight = 0.7
enable_reranking = false
enable_query_expansion = false
`

		configContent += memoryConfig
	}

	// Add orchestration configuration
	orchestrationConfig := generateOrchestrationConfig(config)
	configContent += orchestrationConfig

	configPath := filepath.Join(config.Name, "agentflow.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create agentflow.toml: %w", err)
	}
	fmt.Printf("Created file: %s\n", configPath)
	return nil
}

// generateWorkflowDiagrams creates Mermaid workflow diagrams for the project
func generateWorkflowDiagrams(config ProjectConfig) error {
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
		diagram, title = generateCollaborativeDiagram(config)
	case "sequential":
		diagram, title = generateSequentialDiagram(config)
	case "loop":
		diagram, title = generateLoopDiagram(config)
	case "mixed":
		diagram, title = generateMixedDiagram(config)
	default:
		diagram, title = generateRouteDiagram(config)
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
`, title, config.OrchestrationMode, diagram, config.OrchestrationMode, config.NumAgents, config.OrchestrationTimeout, config.MaxConcurrency, config.FailureThreshold, generateAgentDetails(config))

	if err := os.WriteFile(diagramPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create workflow diagram: %w", err)
	}
	fmt.Printf("Created workflow diagram: %s\n", diagramPath)

	return nil
}

// generateCollaborativeDiagram creates a collaborative orchestration diagram
func generateCollaborativeDiagram(config ProjectConfig) (string, string) {
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
func generateSequentialDiagram(config ProjectConfig) (string, string) {
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

// getConnectionString returns the appropriate connection string for a memory provider
func getConnectionString(memoryProvider string) string {
	switch memoryProvider {
	case "pgvector":
		return "postgres://user:password@localhost:15432/agentflow?sslmode=disable"
	case "weaviate":
		return "http://localhost:8080"
	default:
		return "memory"
	}
}

// generateOrchestrationConfig generates the orchestration configuration section for TOML
func generateOrchestrationConfig(config ProjectConfig) string {
	orchestrationConfig := fmt.Sprintf(`
[orchestration]
mode = "%s"
timeout_seconds = %d`, config.OrchestrationMode, config.OrchestrationTimeout)

	// Add mode-specific configuration
	switch config.OrchestrationMode {
	case "sequential":
		if len(config.SequentialAgents) > 0 {
			orchestrationConfig += "\nsequential_agents = ["
			for i, agent := range config.SequentialAgents {
				if i > 0 {
					orchestrationConfig += ", "
				}
				orchestrationConfig += fmt.Sprintf("\"%s\"", agent)
			}
			orchestrationConfig += "]"
		} else {
			// Generate default sequential agents based on NumAgents
			orchestrationConfig += "\nsequential_agents = ["
			for i := 0; i < config.NumAgents; i++ {
				if i > 0 {
					orchestrationConfig += ", "
				}
				orchestrationConfig += fmt.Sprintf("\"agent%d\"", i+1)
			}
			orchestrationConfig += "]"
		}

	case "loop":
		if config.LoopAgent != "" {
			orchestrationConfig += fmt.Sprintf("\nloop_agent = \"%s\"", config.LoopAgent)
		} else {
			orchestrationConfig += "\nloop_agent = \"agent1\""
		}
		orchestrationConfig += fmt.Sprintf("\nmax_iterations = %d", config.MaxIterations)

	case "mixed":
		// Add collaborative agents
		if len(config.CollaborativeAgents) > 0 {
			orchestrationConfig += "\ncollaborative_agents = ["
			for i, agent := range config.CollaborativeAgents {
				if i > 0 {
					orchestrationConfig += ", "
				}
				orchestrationConfig += fmt.Sprintf("\"%s\"", agent)
			}
			orchestrationConfig += "]"
		}

		// Add sequential agents
		if len(config.SequentialAgents) > 0 {
			orchestrationConfig += "\nsequential_agents = ["
			for i, agent := range config.SequentialAgents {
				if i > 0 {
					orchestrationConfig += ", "
				}
				orchestrationConfig += fmt.Sprintf("\"%s\"", agent)
			}
			orchestrationConfig += "]"
		}

		// If no agents specified, create default mixed configuration
		if len(config.CollaborativeAgents) == 0 && len(config.SequentialAgents) == 0 {
			// First agent collaborative, rest sequential
			orchestrationConfig += "\ncollaborative_agents = [\"agent1\"]"
			if config.NumAgents > 1 {
				orchestrationConfig += "\nsequential_agents = ["
				for i := 1; i < config.NumAgents; i++ {
					if i > 1 {
						orchestrationConfig += ", "
					}
					orchestrationConfig += fmt.Sprintf("\"agent%d\"", i+1)
				}
				orchestrationConfig += "]"
			}
		}
	}

	orchestrationConfig += "\n"
	return orchestrationConfig
}

// generateLoopDiagram creates a loop orchestration diagram
func generateLoopDiagram(config ProjectConfig) (string, string) {
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
func generateMixedDiagram(config ProjectConfig) (string, string) {
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
func generateRouteDiagram(config ProjectConfig) (string, string) {
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
func generateAgentDetails(config ProjectConfig) string {
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
		Visualize:            config.Visualize,
		VisualizeOutputDir:   config.VisualizeOutputDir,
	}
}

// generateProviderInitFunction generates the provider initialization function code
func generateProviderInitFunction(config ProjectConfig) string {
	return `func initializeProvider(providerType string) (core.ModelProvider, error) {
	// Use the config-based provider initialization
	return core.NewProviderFromWorkingDir()
}`
}

// generateMCPInitFunction generates the MCP initialization function code
func generateMCPInitFunction(config ProjectConfig) string {
	return `func initializeMCP() (core.MCPManager, error) {
	// Load configuration from agentflow.toml in current directory
	config, err := core.LoadConfigFromWorkingDir()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if MCP is enabled in configuration
	if !config.MCP.Enabled {
		return nil, fmt.Errorf("MCP is not enabled in agentflow.toml")
	}

	// Convert TOML config to MCP config
	mcpConfig := core.MCPConfig{
		EnableDiscovery:   config.MCP.EnableDiscovery,
		ConnectionTimeout: time.Duration(config.MCP.ConnectionTimeout) * time.Millisecond,
		MaxRetries:        config.MCP.MaxRetries,
		RetryDelay:        time.Duration(config.MCP.RetryDelay) * time.Millisecond,
		EnableCaching:     config.MCP.EnableCaching,
		CacheTimeout:      time.Duration(config.MCP.CacheTimeout) * time.Millisecond,
		MaxConnections:    config.MCP.MaxConnections,
		Servers:           make([]core.MCPServerConfig, len(config.MCP.Servers)),
	}

	// Convert server configurations
	for i, server := range config.MCP.Servers {
		mcpConfig.Servers[i] = core.MCPServerConfig{
			Name:    server.Name,
			Type:    server.Type,
			Host:    server.Host,
			Port:    server.Port,
			Command: server.Command,
			Enabled: server.Enabled,
		}
	}

	// Initialize MCP manager with configuration from TOML
	err = core.InitializeMCP(mcpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MCP: %w", err)
	}

	// Get the initialized MCP manager
	manager := core.GetMCPManager()
	if manager == nil {
		return nil, fmt.Errorf("MCP manager not available after initialization")
	}

	return manager, nil
}`
}

// generateCacheInitFunction generates the cache initialization function code
func generateCacheInitFunction(config ProjectConfig) string {
	return `func initializeCache() error {
	// Cache initialization placeholder
	return nil
}`
}
// createProjectDirectories creates the main project subdirectories
func createProjectDirectories(config ProjectConfig) error {
	// Create agents directory
	if err := createAgentsDirectory(config); err != nil {
		return err
	}

	// Create internal directory (optional, for future use)
	if err := createInternalDirectory(config); err != nil {
		return err
	}

	// Create docs directory
	if err := createDocsDirectory(config); err != nil {
		return err
	}

	return nil
}

// createAgentsDirectory creates the agents subdirectory for agent implementations
func createAgentsDirectory(config ProjectConfig) error {
	agentsDir := filepath.Join(config.Name, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory %s: %w", agentsDir, err)
	}
	fmt.Printf("Created directory: %s\n", agentsDir)
	return nil
}

// createInternalDirectory creates the internal subdirectory for internal packages
func createInternalDirectory(config ProjectConfig) error {
	internalDir := filepath.Join(config.Name, "internal")
	if err := os.MkdirAll(internalDir, 0755); err != nil {
		return fmt.Errorf("failed to create internal directory %s: %w", internalDir, err)
	}
	fmt.Printf("Created directory: %s\n", internalDir)
	
	// Create subdirectories within internal
	configDir := filepath.Join(internalDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create internal/config directory %s: %w", configDir, err)
	}
	fmt.Printf("Created directory: %s\n", configDir)

	handlersDir := filepath.Join(internalDir, "handlers")
	if err := os.MkdirAll(handlersDir, 0755); err != nil {
		return fmt.Errorf("failed to create internal/handlers directory %s: %w", handlersDir, err)
	}
	fmt.Printf("Created directory: %s\n", handlersDir)

	return nil
}

// createDocsDirectory creates the docs subdirectory for additional documentation
func createDocsDirectory(config ProjectConfig) error {
	docsDir := filepath.Join(config.Name, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory %s: %w", docsDir, err)
	}
	fmt.Printf("Created directory: %s\n", docsDir)
	return nil
}

// createAgentsReadme creates the agents/README.md file
func createAgentsReadme(config ProjectConfig) error {
	utilsConfig := convertToUtilsConfig(config)
	agents := utils.ResolveAgentNames(utilsConfig)

	if len(agents) == 0 {
		agents = append(agents, utils.CreateAgentInfo("agent1", config.OrchestrationMode))
	}

	// Use the agents README template
	tmpl, err := template.New("agents_readme").Parse(templates.AgentsReadmeTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse agents README template: %w", err)
	}

	// Create template data structure with enhanced agent info
	type EnhancedAgentInfo struct {
		utils.AgentInfo
		IsFirstAgent bool
		IsLastAgent  bool
		Index        int
	}

	var enhancedAgents []EnhancedAgentInfo
	for i, agent := range agents {
		enhancedAgents = append(enhancedAgents, EnhancedAgentInfo{
			AgentInfo:    agent,
			IsFirstAgent: i == 0,
			IsLastAgent:  i == len(agents)-1,
			Index:        i,
		})
	}

	templateData := struct {
		Config              ProjectConfig
		Agents              []EnhancedAgentInfo
		ProjectStructure    ProjectStructureInfo
		CustomizationPoints []CustomizationPoint
		ImportPaths         ImportPathInfo
		Features            []string
		FrameworkVersion    string
	}{
		Config:              config,
		Agents:              enhancedAgents,
		ProjectStructure:    CreateProjectStructureInfo(config),
		CustomizationPoints: CreateCustomizationPoints(config),
		ImportPaths:         CreateImportPathInfo(config),
		Features:            GetEnabledFeatures(config),
		FrameworkVersion:    AgenticGoKitVersion,
	}

	readmePath := filepath.Join(config.Name, "agents", "README.md")
	file, err := os.Create(readmePath)
	if err != nil {
		return fmt.Errorf("failed to create agents/README.md file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute agents README template: %w", err)
	}

	fmt.Printf("Created file: %s\n", readmePath)
	return nil
}

// createCustomizationGuide creates the docs/CUSTOMIZATION.md file
func createCustomizationGuide(config ProjectConfig) error {
	utilsConfig := convertToUtilsConfig(config)
	agents := utils.ResolveAgentNames(utilsConfig)

	if len(agents) == 0 {
		agents = append(agents, utils.CreateAgentInfo("agent1", config.OrchestrationMode))
	}

	// Use the customization guide template
	tmpl, err := template.New("customization_guide").Parse(templates.CustomizationGuideTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse customization guide template: %w", err)
	}

	// Create template data structure with enhanced agent info
	type EnhancedAgentInfo struct {
		utils.AgentInfo
		IsFirstAgent bool
		IsLastAgent  bool
		Index        int
	}

	var enhancedAgents []EnhancedAgentInfo
	for i, agent := range agents {
		enhancedAgents = append(enhancedAgents, EnhancedAgentInfo{
			AgentInfo:    agent,
			IsFirstAgent: i == 0,
			IsLastAgent:  i == len(agents)-1,
			Index:        i,
		})
	}

	templateData := struct {
		Config              ProjectConfig
		Agents              []EnhancedAgentInfo
		ProjectStructure    ProjectStructureInfo
		CustomizationPoints []CustomizationPoint
		ImportPaths         ImportPathInfo
		Features            []string
		FrameworkVersion    string
	}{
		Config:              config,
		Agents:              enhancedAgents,
		ProjectStructure:    CreateProjectStructureInfo(config),
		CustomizationPoints: CreateCustomizationPoints(config),
		ImportPaths:         CreateImportPathInfo(config),
		Features:            GetEnabledFeatures(config),
		FrameworkVersion:    AgenticGoKitVersion,
	}

	guidePath := filepath.Join(config.Name, "docs", "CUSTOMIZATION.md")
	file, err := os.Create(guidePath)
	if err != nil {
		return fmt.Errorf("failed to create docs/CUSTOMIZATION.md file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute customization guide template: %w", err)
	}

	fmt.Printf("Created file: %s\n", guidePath)
	return nil
}