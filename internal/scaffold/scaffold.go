package scaffold

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/kunalkushwaha/agenticgokit/internal/scaffold/templates"
	"github.com/kunalkushwaha/agenticgokit/internal/scaffold/utils"
)

func CreateAgentProject(config ProjectConfig) error {
	return CreateAgentProjectModular(config)
}

// CreateAgentProjectFromTemplate creates a new AgentFlow project from a template
func CreateAgentProjectFromTemplate(templateName, projectName string) error {
	templatePath := filepath.Join("examples/templates", templateName+".yaml")

	generator, err := NewTemplateGenerator(templatePath)
	if err != nil {
		return fmt.Errorf("failed to create template generator: %w", err)
	}

	return generator.GenerateProject(projectName)
}

// CreateAgentProjectFromConfig creates a new AgentFlow project using ProjectConfig (alias)
func CreateAgentProjectFromConfig(config ProjectConfig) error {
	return CreateAgentProjectModular(config)
}

// CreateAgentProjectModular creates a new AgentFlow project using the modular template system
func CreateAgentProjectModular(config ProjectConfig) error {
	// Apply intelligent defaults for embedding models and configurations
	ApplyIntelligentDefaults(&config)

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

	// Create plugins bundle (blank imports) to activate selected providers
	if err := createPluginsBundle(config); err != nil {
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

	fmt.Printf("\nProject '%s' created successfully using modular templates.\n", config.Name)
	fmt.Printf("Directory: %s\n", config.Name)
	fmt.Printf("Run: cd %s && go mod tidy && go run . -m \"Your message\"\n", config.Name)

	return nil
}

// createGoMod creates the go.mod file
func createGoMod(config ProjectConfig) error {
	// Create go.mod content with proper module declaration
	goModContent := fmt.Sprintf(`module %s

go 1.21

require github.com/kunalkushwaha/agenticgokit %s
`, config.Name, AgenticGoKitVersion)

	// If running from a local source checkout, add a replace to the repo root so
	// generated projects can resolve in-repo packages (e.g., plugins/*) during tests/dev.
	if repoRoot := findLocalRepoRoot(); repoRoot != "" {
		goModContent += fmt.Sprintf("\nreplace github.com/kunalkushwaha/agenticgokit => %s\n", repoRoot)
	}

	goModPath := filepath.Join(config.Name, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}
	fmt.Printf("Created file: %s\n", goModPath)
	return nil
}

// findLocalRepoRoot walks up from this file to locate the repo root containing go.mod
// that declares module github.com/kunalkushwaha/agenticgokit. Returns "" if not found.
func findLocalRepoRoot() string {
	// This file lives under internal/scaffold. Start from there.
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	dir := filepath.Dir(file)
	const modDecl = "module github.com/kunalkushwaha/agenticgokit"

	for {
		gm := filepath.Join(dir, "go.mod")
		if f, err := os.Open(gm); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if strings.HasPrefix(strings.TrimSpace(scanner.Text()), modDecl) {
					f.Close()
					return dir
				}
			}
			f.Close()
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
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
	// Ensure embedding dimensions are available for templates
	if config.EmbeddingDimensions == 0 {
		config.EmbeddingDimensions = GetModelDimensions(config.EmbeddingProvider, config.EmbeddingModel)
	}

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
		ProviderInitFunction: "", // Remove generated function - template contains full implementation
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
description = "Configuration-driven multi-agent system"

# Global LLM configuration - can be overridden per agent
[llm]
provider = "%s"
model = "%s"
temperature = 0.7
max_tokens = 2000
timeout_seconds = 30

[logging]
level = "info"
format = "console"

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
`, config.Name, config.Provider, getDefaultModelForProvider(config.Provider))

	// Add MCP configuration if enabled
	if config.MCPEnabled {
		// Determine transport string (default to tcp)
		transport := config.MCPTransport
		if transport == "" {
			transport = "tcp"
		}

		mcpConfig := `
[mcp]
enabled = true
transport = "` + transport + `"
enable_discovery = true
connection_timeout = 5000
max_retries = 3
retry_delay = 1000
enable_caching = true
cache_timeout = 300000
max_connections = 10

# MCP Server Examples - Uncomment and configure as needed
#
# For Docker AI Gateway (provides web search, content fetching, etc.)
# Start with: docker run -p 8812:8812 -p 8813:8813 your-docker-image

[[mcp.servers]]
name = "docker-http-sse"
type = "http_sse"
host = "localhost"
port = 8812
enabled = false

[[mcp.servers]]
name = "docker-http-streaming"
type = "http_streaming"
host = "localhost"
port = 8813
enabled = false

# For file system access
# Install with: npm install -g @modelcontextprotocol/server-filesystem
# [[mcp.servers]]
# name = "filesystem"
# type = "stdio"
# command = "npx @modelcontextprotocol/server-filesystem /path/to/allowed/files"
# enabled = true

# For web search capabilities
# Install with: npm install -g @modelcontextprotocol/server-brave-search
# [[mcp.servers]]
# name = "brave-search"
# type = "stdio"
# command = "npx @modelcontextprotocol/server-brave-search"
# enabled = true

# For SQLite database access
# Install with: npm install -g @modelcontextprotocol/server-sqlite
# [[mcp.servers]]
# name = "sqlite"
# type = "stdio"
# command = "npx @modelcontextprotocol/server-sqlite /path/to/database.db"
# enabled = true

# For GitHub integration
# Install with: npm install -g @modelcontextprotocol/server-github
# Set GITHUB_PERSONAL_ACCESS_TOKEN environment variable
# [[mcp.servers]]
# name = "github"
# type = "stdio"
# command = "npx @modelcontextprotocol/server-github"
# enabled = true

# For custom TCP server
# [[mcp.servers]]
# name = "custom-tcp"
# type = "tcp"
# host = "localhost"
# port = 8811
# enabled = false
`
		configContent += mcpConfig

		// Include cache configuration section if caching is requested
		if config.WithCache {
			mcpCache := `
[mcp.cache]
enabled = true
# Default TTL for cached tool results (milliseconds)
default_ttl_ms = 900000
# Maximum cache size (MB) and max number of keys
max_size_mb = 100
max_keys = 10000
# Eviction policy: lru | lfu | ttl
eviction_policy = "lru"
# Cleanup interval for expired entries (milliseconds)
cleanup_interval_ms = 300000
# Backend: memory | redis | file
backend = "memory"

# Backend-specific configuration (keys depend on selected backend)
[mcp.cache.backend_config]
redis_addr = "localhost:6379"
redis_password = ""
redis_db = "0"
file_path = "./cache"

# Optional per-tool TTL overrides (milliseconds)
[mcp.cache.tool_ttls_ms]
# web_search = 300000
# content_fetch = 1800000
`
			configContent += mcpCache
		}
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

	// Add agent definitions based on project configuration
	agentConfig := generateAgentConfig(config)
	configContent += agentConfig

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
	diagram += "    ORCHESTRATOR[\"Collaborative Orchestrator\"]\n"
	diagram += "    AGGREGATOR[\"Result Aggregator\"]\n"
	diagram += "    RESULT[\"📤 Final Result\"]\n\n"
	diagram += "    EVENT --> ORCHESTRATOR\n"

	for i, agent := range agents {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		diagram += fmt.Sprintf("    %s[\"%s\"]\n", agentId, agent)
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
		diagram += fmt.Sprintf("    %s[\"%s\"]\n", agentId, agent)
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

// generateAgentConfig generates agent definitions for the configuration file
func generateAgentConfig(config ProjectConfig) string {
	utilsConfig := convertToUtilsConfig(config)
	agents := utils.ResolveAgentNames(utilsConfig)

	if len(agents) == 0 {
		agents = append(agents, utils.CreateAgentInfo("agent1", config.OrchestrationMode))
	}

	agentConfig := "\n# Agent Definitions\n# Each agent has its own configuration including role, capabilities, and LLM settings\n"

	for i, agent := range agents {
		agentConfig += fmt.Sprintf(`
[agents.%s]
role = "%s"
description = "%s"
system_prompt = """%s"""
capabilities = %s
enabled = true
auto_llm = true
`, agent.Name, agent.Name, agent.Purpose,
			generateSystemPromptForConfig(agent, i, len(agents), config.OrchestrationMode),
			formatCapabilitiesArray(generateCapabilitiesForAgent(agent.Name)))

		// Add agent-specific LLM configuration with variations
		agentConfig += fmt.Sprintf(`
# Agent-specific LLM settings (overrides global settings)
[agents.%s.llm]`, agent.Name)

		// Vary temperature based on agent role for better results
		temperature := getTemperatureForAgent(agent.Name, i)
		agentConfig += fmt.Sprintf(`
temperature = %.1f`, temperature)

		// Vary max_tokens based on agent purpose
		maxTokens := getMaxTokensForAgent(agent.Name, i)
		agentConfig += fmt.Sprintf(`
max_tokens = %d`, maxTokens)

		// Add retry policy for production agents
		if config.Provider != "mock" {
			agentConfig += fmt.Sprintf(`

# Retry policy for %s
[agents.%s.retry_policy]
max_retries = 3
base_delay_ms = 1000
max_delay_ms = 10000
backoff_factor = 2.0`, agent.DisplayName, agent.Name)
		}

		agentConfig += "\n"
	}

	return agentConfig
}

// generateSystemPromptForConfig creates a comprehensive system prompt suitable for configuration files
func generateSystemPromptForConfig(agent utils.AgentInfo, index, total int, orchestrationMode string) string {
	// Create detailed, comprehensive system prompts based on orchestration mode
	switch orchestrationMode {
	case "sequential":
		return generateSequentialConfigPrompt(agent, index, total)
	case "collaborative":
		return generateCollaborativeConfigPrompt(agent, index, total)
	case "loop":
		return generateLoopConfigPrompt(agent, index, total)
	default:
		return generateGenericConfigPrompt(agent, index, total)
	}
}

// generateSequentialConfigPrompt creates detailed sequential workflow prompts
func generateSequentialConfigPrompt(agent utils.AgentInfo, index, total int) string {
	var prompt strings.Builder

	// Check for research-specific agents and provide specialized prompts
	switch agent.Name {
	case "researcher":
		return generateResearcherPrompt()
	case "analyzer":
		return generateAnalyzerPrompt()
	case "synthesizer":
		return generateSynthesizerPrompt()
	}

	// Default sequential prompts for non-research agents
	if index == 0 {
		prompt.WriteString(fmt.Sprintf("You are %s, the first agent in a sequential multi-agent system. Your primary role is to analyze and process the initial user query comprehensively. You should:\n\n", agent.DisplayName))
		prompt.WriteString("1. Thoroughly understand the user's request and identify key requirements\n")
		prompt.WriteString("2. Gather relevant facts, data, and insights using available tools\n")
		prompt.WriteString("3. Analyze the information critically and identify important details\n")
		prompt.WriteString("4. Present your findings in a structured, factual manner\n")
		prompt.WriteString("5. Include any relevant context, examples, or supporting evidence\n")
		prompt.WriteString("6. Flag any uncertainties or areas that need further clarification\n")
		prompt.WriteString("7. Provide substantial content that fully addresses the user's query\n\n")
		prompt.WriteString("Your output will be passed to the next agent for further processing. Focus on accuracy, completeness, and providing a strong foundation for subsequent analysis.")
	} else if index == total-1 {
		prompt.WriteString(fmt.Sprintf("You are %s, the final agent in a sequential multi-agent system. Your role is to synthesize and present the final response to the user. You should:\n\n", agent.DisplayName))
		prompt.WriteString("1. Review and integrate all analysis from previous agents\n")
		prompt.WriteString("2. Organize the information in a logical, easy-to-follow structure\n")
		prompt.WriteString("3. Use clear, simple language that is accessible to the user\n")
		prompt.WriteString("4. Create proper headings, bullet points, and formatting for readability\n")
		prompt.WriteString("5. Eliminate jargon and explain technical terms when necessary\n")
		prompt.WriteString("6. Ensure the response flows naturally and is engaging\n")
		prompt.WriteString("7. Highlight key takeaways and important points\n")
		prompt.WriteString("8. Provide a concise summary or conclusion when appropriate\n\n")
		prompt.WriteString("Your goal is to make the information as clear and digestible as possible while maintaining accuracy and completeness. Write in a friendly, professional tone that helps users easily understand complex topics.")
	} else {
		prompt.WriteString(fmt.Sprintf("You are %s, agent %d of %d in a sequential multi-agent system. Your role is to enhance and expand upon the work of previous agents. You should:\n\n", agent.DisplayName, index+1, total))
		prompt.WriteString("1. Carefully analyze the output from previous agents\n")
		prompt.WriteString("2. Identify areas that need additional detail or clarification\n")
		prompt.WriteString("3. Add specialized insights based on your capabilities\n")
		prompt.WriteString("4. Use available tools to gather additional relevant information\n")
		prompt.WriteString("5. Validate and cross-check information for accuracy\n")
		prompt.WriteString("6. Enhance the analysis with deeper context and understanding\n")
		prompt.WriteString("7. Prepare the information for the next agent in the sequence\n\n")
		prompt.WriteString("Focus on adding genuine value while maintaining the coherence and flow of the overall analysis.")
	}

	return prompt.String()
}

// generateResearcherPrompt creates specialized prompt for research agents
func generateResearcherPrompt() string {
	return `You are Researcher, the first agent in a sequential research workflow. Your specialized role is to conduct comprehensive research and information gathering. You should:

**Primary Research Objectives:**
1. Thoroughly understand the research query and identify key information needs
2. Use available web search tools to find current, authoritative sources
3. Gather comprehensive facts, data, statistics, and expert opinions
4. Identify multiple perspectives on controversial or complex topics
5. Note publication dates, source credibility, and potential biases
6. Collect specific examples, case studies, and supporting evidence
7. Flag areas requiring deeper investigation or specialized knowledge

**Research Quality Standards:**
- Prioritize recent, authoritative sources over outdated information
- Cross-reference claims across multiple reliable sources
- Include both primary sources (research papers, official reports) and secondary analyses
- Note any conflicting information or ongoing debates
- Gather quantitative data (statistics, trends) when relevant
- Document source URLs and publication information for verification

**Information Organization:**
- Structure findings by topic/theme for easy analysis
- Separate facts from opinions and interpretations
- Highlight key insights and surprising findings
- Note information gaps that need further investigation

**Output for Next Agent:**
Provide a comprehensive research brief with:
- Summary of key findings organized by topic
- Important facts, statistics, and quotes with sources
- Multiple perspectives on the topic
- Areas of uncertainty or conflicting information
- Recommendations for further analysis

Your thorough research will enable the next agent (Analyzer) to perform deep analysis on solid factual foundations.`
}

// generateAnalyzerPrompt creates specialized prompt for analysis agents
func generateAnalyzerPrompt() string {
	return `You are Analyzer, the second agent in a sequential research workflow. You receive comprehensive research from the Researcher agent and your role is to perform deep analysis and identify insights. You should:

**Analysis Objectives:**
1. Critically evaluate the research findings from the previous agent
2. Identify patterns, trends, and relationships in the data
3. Analyze cause-and-effect relationships and underlying factors
4. Compare different perspectives and evaluate their strengths/weaknesses
5. Identify gaps in logic, missing information, or potential biases
6. Generate insights that aren't immediately obvious from the raw data
7. Assess the reliability and significance of different findings

**Analytical Framework:**
- **Pattern Recognition**: Identify recurring themes, trends, or behaviors
- **Comparative Analysis**: Compare different approaches, solutions, or viewpoints
- **Root Cause Analysis**: Dig deeper into underlying causes and contributing factors
- **Impact Assessment**: Evaluate potential consequences and implications
- **Credibility Assessment**: Evaluate source reliability and potential biases
- **Gap Analysis**: Identify missing information or logical inconsistencies

**Critical Thinking Applications:**
- Challenge assumptions and conventional wisdom
- Look for contradictions or conflicting evidence
- Consider alternative explanations for phenomena
- Evaluate the strength of evidence supporting different claims
- Identify potential confounding variables or hidden factors

**Analysis Output Structure:**
- **Key Insights**: Most important discoveries and implications
- **Pattern Analysis**: Identified trends, relationships, and recurring themes
- **Comparative Assessment**: Strengths/weaknesses of different approaches
- **Risk/Opportunity Analysis**: Potential positive and negative outcomes
- **Evidence Evaluation**: Reliability assessment of key claims
- **Knowledge Gaps**: Areas needing additional research or clarification
- **Analytical Conclusions**: Evidence-based conclusions with confidence levels

**Output for Next Agent:**
Provide a comprehensive analytical report that the Synthesizer can use to create a well-structured, insightful final response. Include both your analytical findings and the supporting evidence from the research phase.`
}

// generateSynthesizerPrompt creates specialized prompt for synthesis agents
func generateSynthesizerPrompt() string {
	return `You are Synthesizer, the final agent in a sequential research workflow. You receive both comprehensive research data and deep analysis from previous agents. Your role is to create a coherent, well-structured final response that effectively communicates insights to the user. You should:

**Synthesis Objectives:**
1. Integrate research findings and analytical insights into a coherent narrative
2. Organize complex information into a logical, easy-to-follow structure
3. Translate technical or complex concepts into accessible language
4. Create compelling, engaging content that maintains accuracy
5. Provide balanced perspectives while highlighting key insights
6. Structure the response for the intended audience and purpose
7. Include actionable recommendations when appropriate

**Content Organization Strategy:**
- **Executive Summary**: Key findings and insights (for long responses)
- **Main Content**: Organized by themes, importance, or logical flow
- **Supporting Evidence**: Specific examples, data, and expert opinions
- **Multiple Perspectives**: Balanced presentation of different viewpoints
- **Practical Implications**: Real-world applications and consequences
- **Conclusions**: Clear synthesis of key insights and recommendations

**Communication Excellence:**
- Use clear, engaging language appropriate for the target audience
- Create logical flow with smooth transitions between ideas
- Employ headings, bullet points, and formatting for readability
- Include specific examples and concrete details to illustrate points
- Balance comprehensive coverage with focused insights
- Maintain objectivity while making the content engaging

**Quality Assurance:**
- Ensure all major research findings are incorporated
- Verify that analytical insights are clearly communicated
- Check for logical consistency throughout the response
- Confirm that conclusions are supported by evidence
- Eliminate jargon and explain technical terms when necessary
- Provide source attributions for key claims

**Response Structure Guidelines:**
- **Introduction**: Brief overview of the topic and approach
- **Key Findings**: Most important discoveries (3-5 main points)
- **Detailed Analysis**: In-depth exploration of findings with evidence
- **Implications**: What these findings mean for different stakeholders
- **Recommendations**: Actionable next steps when appropriate
- **Conclusion**: Summary of key insights and final thoughts

**Final Output:**
Create a comprehensive, well-structured response that:
- Fully addresses the original user query
- Integrates all valuable research and analysis
- Provides clear insights and actionable information
- Is engaging, readable, and appropriately detailed
- Demonstrates the value of the multi-agent research process

Your synthesis should showcase the depth and quality of the collaborative research effort while being accessible and valuable to the user.`
}

// generateCollaborativeConfigPrompt creates detailed collaborative workflow prompts
func generateCollaborativeConfigPrompt(agent utils.AgentInfo, index, total int) string {
	// Use the comprehensive prompt generation from utils package which includes RAG-specific prompts
	return utils.CreateSystemPrompt(agent, index, total, "collaborative")
}

// generateLoopConfigPrompt creates detailed loop workflow prompts
func generateLoopConfigPrompt(agent utils.AgentInfo, index, total int) string {
	return fmt.Sprintf("You are %s, operating in an iterative loop mode where you may process the same query multiple times. Your goal is to improve and refine your response with each iteration. Analyze the current state and any previous iterations, identify areas for improvement, use tools to gather new information, and build upon previous iterations while addressing gaps. Each iteration should add meaningful value and aim for convergence toward an optimal response.", agent.DisplayName)
}

// generateGenericConfigPrompt creates detailed generic workflow prompts
func generateGenericConfigPrompt(agent utils.AgentInfo, index, total int) string {
	return fmt.Sprintf("You are %s, an intelligent agent designed to provide comprehensive assistance. Your role is to process user queries effectively and provide helpful responses. Analyze user queries thoroughly, use available tools to gather current and accurate information, provide comprehensive responses that fully address user needs, and maintain high standards for accuracy and relevance. Always be professional, thorough, and focused on delivering value to the user.", agent.DisplayName)
}

// generateCapabilitiesForAgent creates appropriate capabilities based on agent name
func generateCapabilitiesForAgent(agentName string) []string {
	// Generate capabilities based on agent name patterns using known capabilities
	switch {
	case strings.Contains(agentName, "research"):
		return []string{"research", "information_gathering", "fact_checking", "source_identification"}
	case strings.Contains(agentName, "writer") || strings.Contains(agentName, "content"):
		return []string{"content_creation", "writing", "editing", "documentation"}
	case strings.Contains(agentName, "review") || strings.Contains(agentName, "validator"):
		return []string{"fact_checking", "editing", "analysis", "testing"}
	case strings.Contains(agentName, "analyst") || strings.Contains(agentName, "analyzer"):
		return []string{"data_analysis", "pattern_recognition", "insight_generation", "trend_analysis"}
	case strings.Contains(agentName, "processor"):
		return []string{"data_processing", "text_analysis", "pattern_recognition", "analysis"}
	case strings.Contains(agentName, "summary") || strings.Contains(agentName, "summarizer"):
		return []string{"summarization", "text_analysis", "content_creation", "editing"}
	case strings.Contains(agentName, "creative"):
		return []string{"content_creation", "writing", "editing", "analysis"}
	case strings.Contains(agentName, "coordinator") || strings.Contains(agentName, "manager"):
		return []string{"analysis", "data_processing", "documentation", "research"}
	case strings.Contains(agentName, "collector"):
		return []string{"information_gathering", "data_processing", "source_identification", "research"}
	case strings.Contains(agentName, "synthesizer"):
		return []string{"analysis", "summarization", "insight_generation", "content_creation"}
	default:
		// Default capabilities for generic agents using known capabilities
		return []string{"analysis", "text_analysis", "data_processing"}
	}
}

// formatCapabilitiesArray formats capabilities as a TOML array
func formatCapabilitiesArray(capabilities []string) string {
	if len(capabilities) == 0 {
		return `["general_assistance"]`
	}

	quotedCapabilities := make([]string, len(capabilities))
	for i, cap := range capabilities {
		quotedCapabilities[i] = fmt.Sprintf(`"%s"`, cap)
	}

	return fmt.Sprintf("[%s]", strings.Join(quotedCapabilities, ", "))
}

// getTemperatureForAgent returns appropriate temperature based on agent role
func getTemperatureForAgent(agentName string, index int) float64 {
	// Vary temperature based on agent name/role for better results
	switch {
	case strings.Contains(agentName, "research") || strings.Contains(agentName, "fact"):
		return 0.3 // Lower temperature for factual tasks
	case strings.Contains(agentName, "creative") || strings.Contains(agentName, "writer"):
		return 0.8 // Higher temperature for creative tasks
	case strings.Contains(agentName, "review") || strings.Contains(agentName, "validator"):
		return 0.2 // Very low temperature for review tasks
	case strings.Contains(agentName, "analyst") || strings.Contains(agentName, "processor"):
		return 0.5 // Medium temperature for analytical tasks
	default:
		// Vary by position: first agent more creative, last agent more precise
		if index == 0 {
			return 0.7
		}
		return 0.6 - float64(index)*0.1 // Gradually decrease temperature
	}
}

// getMaxTokensForAgent returns appropriate max_tokens based on agent role
func getMaxTokensForAgent(agentName string, index int) int {
	switch {
	case strings.Contains(agentName, "writer") || strings.Contains(agentName, "content"):
		return 3000 // More tokens for content creation
	case strings.Contains(agentName, "research") || strings.Contains(agentName, "analyst"):
		return 2500 // More tokens for detailed analysis
	case strings.Contains(agentName, "review") || strings.Contains(agentName, "validator"):
		return 1500 // Fewer tokens for review tasks
	case strings.Contains(agentName, "summary") || strings.Contains(agentName, "brief"):
		return 1000 // Fewer tokens for summaries
	default:
		return 2000 // Default from global config
	}
}

// getDefaultModelForProvider returns the default model for each provider
func getDefaultModelForProvider(provider string) string {
	switch provider {
	case "openai":
		return "gpt-4"
	case "azure":
		return "gpt-4"
	case "ollama":
		return "llama2"
	case "mock":
		return "mock-model"
	default:
		return "gpt-4"
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
	case "collaborative":
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
	diagram += "    AGENT[\"" + agentName + "\"]\n"
	diagram += "    CONDITION{\"Continue Loop?\"}\n"
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
			diagram += fmt.Sprintf("    %s[\"%s\"]\n", agentId, agent)
			diagram += fmt.Sprintf("    PHASE1 --> %s\n", agentId)
			diagram += fmt.Sprintf("    %s --> PHASE2\n", agentId)
		}
	}

	// Add sequential agents
	if len(config.SequentialAgents) > 0 {
		var prevNode = "PHASE2"
		for i, agent := range config.SequentialAgents {
			agentId := fmt.Sprintf("SEQ%d", i+1)
			diagram += fmt.Sprintf("    %s[\"%s\"]\n", agentId, agent)
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
	diagram += "    ROUTER[\"Event Router\"]\n"
	diagram += "    OUTPUT[\"📤 Result\"]\n\n"
	diagram += "    INPUT --> ROUTER\n"

	for i := 0; i < config.NumAgents; i++ {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		agentName := fmt.Sprintf("agent%d", i+1)
		diagram += fmt.Sprintf("    %s[\"%s\"]\n", agentId, agentName)
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
	// Initialize MCP cache manager if MCP caching is enabled
	cfg, err := core.LoadConfig("agentflow.toml")
	if err != nil {
		return fmt.Errorf("failed to load config for cache: %w", err)
	}
	if !cfg.MCP.Enabled {
		return nil
	}

	// Only proceed if caching is enabled globally or via [mcp.cache]
	if !(cfg.MCP.EnableCaching || cfg.MCP.Cache.Enabled) {
		return nil
	}

	// Start from defaults and override using TOML values
	cacheCfg := core.DefaultMCPCacheConfig()

	// Global toggle
	cacheCfg.Enabled = cfg.MCP.EnableCaching || cfg.MCP.Cache.Enabled

	// TTL and cleanup
	if cfg.MCP.Cache.DefaultTTLMS > 0 {
		cacheCfg.DefaultTTL = time.Duration(cfg.MCP.Cache.DefaultTTLMS) * time.Millisecond
	} else if cfg.MCP.CacheTimeout > 0 {
		// Back-compat global cache_timeout_ms
		cacheCfg.DefaultTTL = time.Duration(cfg.MCP.CacheTimeout) * time.Millisecond
	}
	if cfg.MCP.Cache.CleanupIntervalMS > 0 {
		cacheCfg.CleanupInterval = time.Duration(cfg.MCP.Cache.CleanupIntervalMS) * time.Millisecond
	}

	// Size & keys
	if cfg.MCP.Cache.MaxSizeMB > 0 {
		cacheCfg.MaxSize = cfg.MCP.Cache.MaxSizeMB
	}
	if cfg.MCP.Cache.MaxKeys > 0 {
		cacheCfg.MaxKeys = cfg.MCP.Cache.MaxKeys
	}

	// Policy
	if cfg.MCP.Cache.EvictionPolicy != "" {
		cacheCfg.EvictionPolicy = cfg.MCP.Cache.EvictionPolicy
	}

	// Backend
	if cfg.MCP.Cache.Backend != "" {
		cacheCfg.Backend = cfg.MCP.Cache.Backend
	}
	if cfg.MCP.Cache.BackendConfig != nil {
		// Ensure map exists then copy entries
		if cacheCfg.BackendConfig == nil {
			cacheCfg.BackendConfig = map[string]string{}
		}
		for k, v := range cfg.MCP.Cache.BackendConfig {
			cacheCfg.BackendConfig[k] = v
		}
	}

	// Per-tool TTLs
	if len(cfg.MCP.Cache.ToolTTLsMS) > 0 {
		cacheCfg.ToolTTLs = map[string]time.Duration{}
		for tool, ms := range cfg.MCP.Cache.ToolTTLsMS {
			if ms > 0 {
				cacheCfg.ToolTTLs[tool] = time.Duration(ms) * time.Millisecond
			}
		}
	}

	if err := core.InitializeMCPCacheManager(cacheCfg); err != nil {
		return fmt.Errorf("failed to initialize MCP cache manager: %w", err)
	}
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

	// Create WebUI directory if enabled
	if config.WebUIEnabled {
		if err := createWebUIDirectory(config); err != nil {
			return err
		}
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

// createWebUIDirectory creates the WebUI subdirectory and static files
func createWebUIDirectory(config ProjectConfig) error {
	webuiDir := filepath.Join(config.Name, "internal", "webui")
	staticDir := filepath.Join(webuiDir, "static")

	// Create WebUI directories
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("failed to create WebUI directory %s: %w", staticDir, err)
	}
	fmt.Printf("Created directory: %s\n", staticDir)

	// Create WebUI files
	if err := createWebUIFiles(config, staticDir); err != nil {
		return err
	}

	return nil
}

// createWebUIFiles creates the static WebUI files
func createWebUIFiles(config ProjectConfig, staticDir string) error {
	// Create index.html
	indexPath := filepath.Join(staticDir, "index.html")
	indexTemplate, err := template.New("webui_index").Parse(templates.WebUIIndexTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse WebUI index template: %w", err)
	}

	var indexBuffer bytes.Buffer
	indexTemplateData := struct {
		Config ProjectConfig
	}{
		Config: config,
	}
	if err := indexTemplate.Execute(&indexBuffer, indexTemplateData); err != nil {
		return fmt.Errorf("failed to execute WebUI index template: %w", err)
	}

	if err := os.WriteFile(indexPath, indexBuffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write WebUI index file: %w", err)
	}
	fmt.Printf("Created file: %s\n", indexPath)

	// Create style.css
	cssPath := filepath.Join(staticDir, "style.css")
	cssTemplate, err := template.New("webui_css").Parse(templates.WebUICSSTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse WebUI CSS template: %w", err)
	}

	var cssBuffer bytes.Buffer
	cssTemplateData := struct {
		Config ProjectConfig
	}{
		Config: config,
	}
	if err := cssTemplate.Execute(&cssBuffer, cssTemplateData); err != nil {
		return fmt.Errorf("failed to execute WebUI CSS template: %w", err)
	}

	if err := os.WriteFile(cssPath, cssBuffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write WebUI CSS file: %w", err)
	}
	fmt.Printf("Created file: %s\n", cssPath)

	// Create app.js
	jsPath := filepath.Join(staticDir, "app.js")
	// Temporarily disable JS template to test build
	jsContent := "// WebUI JavaScript will be added here"
	if err := os.WriteFile(jsPath, []byte(jsContent), 0644); err != nil {
		return fmt.Errorf("failed to write WebUI JS file: %w", err)
	}
	fmt.Printf("Created file: %s\n", jsPath)

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

// createPluginsBundle generates a plugins.go file that blank-imports selected plugins
func createPluginsBundle(config ProjectConfig) error {
	tmpl, err := template.New("plugins").Parse(templates.PluginsBundleTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse plugins bundle template: %w", err)
	}

	pluginsPath := filepath.Join(config.Name, "plugins.go")
	f, err := os.Create(pluginsPath)
	if err != nil {
		return fmt.Errorf("failed to create plugins.go: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, struct{ Config ProjectConfig }{Config: config}); err != nil {
		return fmt.Errorf("failed to execute plugins bundle template: %w", err)
	}

	fmt.Printf("Created file: %s\n", pluginsPath)
	return nil
}
