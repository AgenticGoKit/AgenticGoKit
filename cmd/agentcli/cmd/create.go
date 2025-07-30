package cmd

import (
	"fmt"
	"strings"

	"github.com/kunalkushwaha/agenticgokit/internal/scaffold"
	"github.com/spf13/cobra"
)

// New consolidated flags
var consolidatedFlags ConsolidatedCreateFlags

// createNewCmd represents the new simplified create command
var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new AgenticGoKit project with multi-agent workflows",
	Long: `Create a new AgenticGoKit project with simplified, intuitive configuration.

This command generates a complete project structure with intelligent defaults
based on the features you enable. Use templates for common project types
or customize with individual flags.

BASIC USAGE:
  # Create a basic project
  agentcli create my-project

  # Create from template
  agentcli create my-project --template research-assistant

  # Interactive mode (recommended for beginners)
  agentcli create --interactive

TEMPLATES:
  basic              Simple multi-agent system (2 agents, sequential)
  research-assistant Multi-agent research with web search and analysis
  rag-system         Document Q&A with vector search and RAG
  data-pipeline      Sequential data processing workflow
  chat-system        Conversational agents with memory

FEATURE FLAGS:
  --memory [provider]        Enable memory system (memory, pgvector, weaviate)
  --embedding [provider]     Embedding provider (openai, ollama:model, dummy)
  --mcp [level]             MCP integration (basic, production, full)
  --rag [chunk-size]        Enable RAG with optional chunk size

EXAMPLES:
  # Research assistant with web search
  agentcli create research-bot --template research-assistant

  # RAG system with PostgreSQL
  agentcli create knowledge-base --memory pgvector --embedding openai --rag

  # Data pipeline with visualization
  agentcli create data-flow --template data-pipeline --visualize

  # Custom configuration
  agentcli create custom-bot --agents 3 --memory pgvector --mcp production

  # Chat system with session memory
  agentcli create chat-bot --template chat-system --memory pgvector

For detailed template information, use: agentcli create --help-templates`,
	Args: func(cmd *cobra.Command, args []string) error {
		if consolidatedFlags.Interactive {
			return nil // Interactive mode doesn't need project name upfront
		}
		if len(args) != 1 {
			return fmt.Errorf("project name is required (or use --interactive)")
		}
		return nil
	},
	RunE: runCreateCommand,
}

// Help command for templates
var createHelpTemplatesCmd = &cobra.Command{
	Use:   "help-templates",
	Short: "Show detailed information about available project templates",
	Long:  "Display comprehensive information about all available project templates including features and use cases.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(GetTemplateHelp())
		
		fmt.Println("TEMPLATE USAGE EXAMPLES:")
		fmt.Println("  agentcli create my-research --template research-assistant")
		fmt.Println("  agentcli create my-kb --template rag-system --memory pgvector")
		fmt.Println("  agentcli create my-pipeline --template data-pipeline --visualize")
		fmt.Println("  agentcli create my-chat --template chat-system")
		fmt.Println()
		fmt.Println("You can override template defaults with additional flags:")
		fmt.Println("  agentcli create my-project --template basic --agents 4 --mcp basic")
	},
}

func init() {
	// Add the consolidated create command
	rootCmd.AddCommand(createCmd)
	
	// Add help subcommand
	createCmd.AddCommand(createHelpTemplatesCmd)

	// Basic flags
	createCmd.Flags().IntVarP(&consolidatedFlags.Agents, "agents", "a", 0, 
		"Number of agents to create (0 = use template default)")
	createCmd.Flags().StringVarP(&consolidatedFlags.Provider, "provider", "p", "", 
		"LLM provider (openai, azure, ollama, mock)")
	createCmd.Flags().StringVarP(&consolidatedFlags.Template, "template", "t", "", 
		"Project template (basic, research-assistant, rag-system, data-pipeline, chat-system)")
	createCmd.Flags().BoolVarP(&consolidatedFlags.Interactive, "interactive", "i", false, 
		"Interactive mode for guided setup")

	// Feature flags
	createCmd.Flags().StringVar(&consolidatedFlags.Memory, "memory", "", 
		"Enable memory system with provider (memory, pgvector, weaviate)")
	createCmd.Flags().StringVar(&consolidatedFlags.Embedding, "embedding", "", 
		"Embedding provider and model (openai, ollama:nomic-embed-text, dummy)")
	createCmd.Flags().StringVar(&consolidatedFlags.MCP, "mcp", "", 
		"MCP integration level (basic, production, full) or tool list")
	createCmd.Flags().StringVar(&consolidatedFlags.RAG, "rag", "", 
		"Enable RAG with optional chunk size (default, 1000, 2000)")

	// Orchestration flags
	createCmd.Flags().StringVar(&consolidatedFlags.Orchestration, "orchestration", "", 
		"Orchestration mode (sequential, collaborative, loop, route)")
	createCmd.Flags().StringVar(&consolidatedFlags.AgentsConfig, "agents-config", "", 
		"Agent configuration (comma-separated names or JSON)")

	// Output flags
	createCmd.Flags().BoolVar(&consolidatedFlags.Visualize, "visualize", false, 
		"Generate Mermaid workflow diagrams")
	createCmd.Flags().StringVar(&consolidatedFlags.OutputDir, "output-dir", "", 
		"Output directory for generated files")

	// Add flag groups for better help organization
	createCmd.Flags().SortFlags = false
	
	// Add completion functions
	createCmd.RegisterFlagCompletionFunc("template", completeTemplateNames)
	createCmd.RegisterFlagCompletionFunc("provider", completeProviderNames)
	createCmd.RegisterFlagCompletionFunc("memory", completeMemoryProviders)
	createCmd.RegisterFlagCompletionFunc("orchestration", completeOrchestrationModes)
	createCmd.RegisterFlagCompletionFunc("mcp", completeMCPLevels)
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	// Handle interactive mode
	if consolidatedFlags.Interactive {
		return runInteractiveCreateMode()
	}

	projectName := args[0]

	// Validate consolidated flags
	if err := consolidatedFlags.Validate(); err != nil {
		return fmt.Errorf("flag validation failed: %w", err)
	}

	// Convert consolidated flags to project config
	config, err := consolidatedFlags.ToProjectConfig(projectName)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Show what we're creating
	fmt.Printf("Creating AgenticGoKit project '%s'...\n", projectName)
	
	if consolidatedFlags.Template != "" {
		template := ProjectTemplates[consolidatedFlags.Template]
		fmt.Printf("[INFO] Using template: %s\n", template.Name)
		fmt.Printf("[INFO] Features: %s\n", strings.Join(template.Features, ", "))
	}

	if config.MemoryEnabled {
		fmt.Printf("[INFO] Memory system enabled (%s)\n", config.MemoryProvider)
		if config.RAGEnabled {
			fmt.Printf("[INFO] RAG enabled (chunk size: %d)\n", config.RAGChunkSize)
		}
	}

	if config.MCPEnabled {
		level := "basic"
		if config.MCPProduction {
			level = "production"
		}
		fmt.Printf("[INFO] MCP integration enabled (%s)\n", level)
		if len(config.MCPTools) > 0 {
			fmt.Printf("[INFO] MCP tools: %s\n", strings.Join(config.MCPTools, ", "))
		}
	}

	if config.Visualize {
		fmt.Printf("[INFO] Workflow visualization enabled\n")
	}

	// Create the project
	return scaffold.CreateAgentProjectModular(config)
}

func runInteractiveCreateMode() error {
	fmt.Println("AgenticGoKit Project Setup")
	fmt.Println("==========================")
	fmt.Println()

	// Project name
	var projectName string
	fmt.Print("Project name: ")
	fmt.Scanln(&projectName)

	// Template selection
	fmt.Println("\nSelect a project template:")
	templates := []string{"basic", "research-assistant", "rag-system", "data-pipeline", "chat-system"}
	for i, template := range templates {
		tmpl := ProjectTemplates[template]
		fmt.Printf("%d. %s - %s\n", i+1, tmpl.Name, tmpl.Description)
	}
	fmt.Print("Choice (1-5, or 0 for custom): ")
	
	var choice int
	fmt.Scanln(&choice)
	
	if choice > 0 && choice <= len(templates) {
		consolidatedFlags.Template = templates[choice-1]
	}

	// Provider selection
	fmt.Println("\nSelect LLM provider:")
	fmt.Println("1. OpenAI (default)")
	fmt.Println("2. Azure OpenAI")
	fmt.Println("3. Ollama (local)")
	fmt.Println("4. Mock (testing)")
	fmt.Print("Choice (1-4): ")
	
	var providerChoice int
	fmt.Scanln(&providerChoice)
	
	providers := []string{"openai", "azure", "ollama", "mock"}
	if providerChoice > 0 && providerChoice <= len(providers) {
		consolidatedFlags.Provider = providers[providerChoice-1]
	}

	// Memory system
	fmt.Print("\nEnable memory system? (y/N): ")
	var memoryChoice string
	fmt.Scanln(&memoryChoice)
	
	if strings.ToLower(memoryChoice) == "y" || strings.ToLower(memoryChoice) == "yes" {
		fmt.Println("Select memory provider:")
		fmt.Println("1. In-memory (development)")
		fmt.Println("2. PostgreSQL with pgvector (production)")
		fmt.Println("3. Weaviate (vector database)")
		fmt.Print("Choice (1-3): ")
		
		var memProviderChoice int
		fmt.Scanln(&memProviderChoice)
		
		memProviders := []string{"memory", "pgvector", "weaviate"}
		if memProviderChoice > 0 && memProviderChoice <= len(memProviders) {
			consolidatedFlags.Memory = memProviders[memProviderChoice-1]
		}

		// RAG
		fmt.Print("Enable RAG (Retrieval-Augmented Generation)? (y/N): ")
		var ragChoice string
		fmt.Scanln(&ragChoice)
		
		if strings.ToLower(ragChoice) == "y" || strings.ToLower(ragChoice) == "yes" {
			consolidatedFlags.RAG = "default"
		}
	}

	// MCP integration
	fmt.Print("\nEnable MCP tool integration? (y/N): ")
	var mcpChoice string
	fmt.Scanln(&mcpChoice)
	
	if strings.ToLower(mcpChoice) == "y" || strings.ToLower(mcpChoice) == "yes" {
		fmt.Println("Select MCP level:")
		fmt.Println("1. Basic (web search)")
		fmt.Println("2. Production (caching, metrics)")
		fmt.Println("3. Full (load balancing, all features)")
		fmt.Print("Choice (1-3): ")
		
		var mcpLevelChoice int
		fmt.Scanln(&mcpLevelChoice)
		
		mcpLevels := []string{"basic", "production", "full"}
		if mcpLevelChoice > 0 && mcpLevelChoice <= len(mcpLevels) {
			consolidatedFlags.MCP = mcpLevels[mcpLevelChoice-1]
		}
	}

	// Visualization
	fmt.Print("\nGenerate workflow diagrams? (y/N): ")
	var vizChoice string
	fmt.Scanln(&vizChoice)
	
	if strings.ToLower(vizChoice) == "y" || strings.ToLower(vizChoice) == "yes" {
		consolidatedFlags.Visualize = true
	}

	// Convert to config and create project
	config, err := consolidatedFlags.ToProjectConfig(projectName)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	fmt.Printf("\nCreating project '%s' with selected configuration...\n", projectName)
	return scaffold.CreateAgentProjectModular(config)
}

// ShowFlagComparison shows the difference between old and new flag structures
func ShowFlagComparison() {
	fmt.Println("FLAG CONSOLIDATION COMPARISON")
	fmt.Println("=============================")
	fmt.Println()
	
	fmt.Println("OLD FLAGS (32 flags):")
	fmt.Println("  --memory-enabled --memory-provider pgvector --embedding-provider openai --rag-enabled --rag-chunk-size 1000")
	fmt.Println("  --mcp-enabled --with-cache --with-metrics --mcp-tools web_search,summarize")
	fmt.Println()
	
	fmt.Println("NEW FLAGS (4 flags for same functionality):")
	fmt.Println("  --memory pgvector --embedding openai --rag 1000 --mcp production")
	fmt.Println()
	
	fmt.Println("TEMPLATE APPROACH:")
	fmt.Println("  --template rag-system  (automatically configures memory, embedding, and RAG)")
	fmt.Println()
}

// Completion functions for intelligent shell completion

// completeTemplateNames provides completion for template names
func completeTemplateNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	templates := getTemplateNames()
	return templates, cobra.ShellCompDirectiveNoFileComp
}

// completeProviderNames provides completion for LLM provider names
func completeProviderNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	providers := []string{"openai", "azure", "ollama", "mock"}
	return providers, cobra.ShellCompDirectiveNoFileComp
}

// completeMemoryProviders provides completion for memory provider names
func completeMemoryProviders(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	providers := []string{"memory", "pgvector", "weaviate"}
	return providers, cobra.ShellCompDirectiveNoFileComp
}

// completeOrchestrationModes provides completion for orchestration modes
func completeOrchestrationModes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	modes := []string{"sequential", "collaborative", "loop", "route"}
	return modes, cobra.ShellCompDirectiveNoFileComp
}

// completeMCPLevels provides completion for MCP integration levels
func completeMCPLevels(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	levels := []string{"basic", "production", "full"}
	return levels, cobra.ShellCompDirectiveNoFileComp
}