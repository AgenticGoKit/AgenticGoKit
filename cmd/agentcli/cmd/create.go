package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kunalkushwaha/agentflow/internal/scaffold"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new AgentFlow project with optional MCP integration",
	Long: `Create a new AgentFlow project with customizable multi-agent workflows.

This command generates a complete project structure including:
  * Multi-agent workflow implementation
  * Configuration files (agentflow.toml)
  * Error handling and responsible AI agents
  * Optional MCP (Model Context Protocol) integration
  * Production-ready features (caching, metrics, load balancing)

Examples:
  # Basic project with 2 agents
  agentcli create myproject

  # Project with specific provider and agent count
  agentcli create myproject --agents 3 --provider azure

  # MCP-enabled project with basic tools
  agentcli create myproject --mcp-enabled

  # Production MCP project with caching and metrics
  agentcli create myproject --mcp-production --with-cache --with-metrics

  # Interactive mode for guided setup
  agentcli create --interactive

  # MCP project with specific tools and servers
  agentcli create myproject --mcp-enabled --mcp-tools "web_search,summarize,translate" --mcp-servers "docker,web-service"`,
	Args: func(cmd *cobra.Command, args []string) error {
		interactive, _ := cmd.Flags().GetBool("interactive")
		if !interactive && len(args) != 1 {
			return fmt.Errorf("project name is required (or use --interactive)")
		}
		return nil
	},
	RunE: runCreateCommand,
}

// Command flags
var (
	// Basic project flags
	numAgents     int
	provider      string
	responsibleAI bool
	errorHandler  bool
	interactive   bool

	// MCP flags
	mcpEnabled         bool
	mcpProduction      bool
	withCache          bool
	withMetrics        bool
	mcpTools           string
	mcpServers         string
	cacheBackend       string
	metricsPort        int
	withLoadBalancer   bool
	connectionPoolSize int
	retryPolicy        string
)

func init() {
	rootCmd.AddCommand(createCmd)

	// Basic project flags
	createCmd.Flags().IntVarP(&numAgents, "agents", "a", 2, "Number of agents to create")
	createCmd.Flags().StringVarP(&provider, "provider", "p", "openai", "LLM provider (openai, azure, ollama, mock)")
	createCmd.Flags().BoolVar(&responsibleAI, "responsible-ai", true, "Include responsible AI agent")
	createCmd.Flags().BoolVar(&errorHandler, "error-handler", true, "Include error handling agents")
	createCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode for guided setup")

	// MCP integration flags
	createCmd.Flags().BoolVar(&mcpEnabled, "mcp-enabled", false, "Enable MCP tool integration")
	createCmd.Flags().BoolVar(&mcpProduction, "mcp-production", false, "Include production MCP features (pooling, retry, metrics)")
	createCmd.Flags().BoolVar(&withCache, "with-cache", false, "Enable MCP result caching")
	createCmd.Flags().BoolVar(&withMetrics, "with-metrics", false, "Enable Prometheus metrics")
	createCmd.Flags().StringVar(&mcpTools, "mcp-tools", "web_search,summarize", "Comma-separated list of MCP tools")
	createCmd.Flags().StringVar(&mcpServers, "mcp-servers", "docker", "Comma-separated list of MCP server names")
	createCmd.Flags().StringVar(&cacheBackend, "cache-backend", "memory", "Cache backend (memory, redis)")
	createCmd.Flags().IntVar(&metricsPort, "metrics-port", 8080, "Metrics server port")
	createCmd.Flags().BoolVar(&withLoadBalancer, "with-load-balancer", false, "Enable MCP load balancing")
	createCmd.Flags().IntVar(&connectionPoolSize, "connection-pool-size", 5, "MCP connection pool size")
	createCmd.Flags().StringVar(&retryPolicy, "retry-policy", "exponential", "Retry policy (exponential, linear, fixed)")

	// Mark MCP production dependencies
	createCmd.MarkFlagsMutuallyExclusive("mcp-production", "mcp-enabled")
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	var projectName string

	if interactive {
		config, err := interactiveSetup()
		if err != nil {
			return fmt.Errorf("interactive setup failed: %w", err)
		}
		return scaffold.CreateAgentProject(config)
	}

	// Non-interactive mode
	projectName = args[0]

	// Validate MCP flag combinations
	if err := validateMCPFlags(); err != nil {
		return err
	}

	// Parse tool and server lists
	toolList := parseCommaSeparatedList(mcpTools)
	serverList := parseCommaSeparatedList(mcpServers)

	// Create project configuration
	config := scaffold.ProjectConfig{
		Name:          projectName,
		NumAgents:     numAgents,
		Provider:      provider,
		ResponsibleAI: responsibleAI,
		ErrorHandler:  errorHandler,

		// MCP configuration
		MCPEnabled:         mcpEnabled || mcpProduction,
		MCPProduction:      mcpProduction,
		WithCache:          withCache,
		WithMetrics:        withMetrics,
		MCPTools:           toolList,
		MCPServers:         serverList,
		CacheBackend:       cacheBackend,
		MetricsPort:        metricsPort,
		WithLoadBalancer:   withLoadBalancer,
		ConnectionPoolSize: connectionPoolSize,
		RetryPolicy:        retryPolicy,
	}

	// Create the project
	fmt.Printf("Creating AgentFlow project '%s'...\n", projectName)
	if config.MCPEnabled {
		fmt.Printf("✓ MCP integration enabled\n")
		if config.MCPProduction {
			fmt.Printf("✓ Production MCP features enabled\n")
		}
		if config.WithCache {
			fmt.Printf("✓ MCP caching enabled (%s)\n", config.CacheBackend)
		}
		if config.WithMetrics {
			fmt.Printf("✓ Metrics enabled on port %d\n", config.MetricsPort)
		}
	}

	return scaffold.CreateAgentProject(config)
}

func validateMCPFlags() error {
	// MCP production implies MCP enabled
	if mcpProduction {
		mcpEnabled = true
	}

	// Cache and metrics require MCP
	if (withCache || withMetrics) && !mcpEnabled && !mcpProduction {
		return fmt.Errorf("--with-cache and --with-metrics require --mcp-enabled or --mcp-production")
	}

	// Load balancer requires production
	if withLoadBalancer && !mcpProduction {
		return fmt.Errorf("--with-load-balancer requires --mcp-production")
	}

	// Validate provider
	validProviders := []string{"openai", "azure", "ollama", "mock"}
	if !contains(validProviders, provider) {
		return fmt.Errorf("invalid provider: %s. Valid options: %s", provider, strings.Join(validProviders, ", "))
	}

	// Validate cache backend
	validBackends := []string{"memory", "redis"}
	if !contains(validBackends, cacheBackend) {
		return fmt.Errorf("invalid cache backend: %s. Valid options: %s", cacheBackend, strings.Join(validBackends, ", "))
	}

	// Validate retry policy
	validPolicies := []string{"exponential", "linear", "fixed"}
	if !contains(validPolicies, retryPolicy) {
		return fmt.Errorf("invalid retry policy: %s. Valid options: %s", retryPolicy, strings.Join(validPolicies, ", "))
	}

	return nil
}

func parseCommaSeparatedList(input string) []string {
	if input == "" {
		return []string{}
	}
	items := strings.Split(input, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func interactiveSetup() (scaffold.ProjectConfig, error) {
	config := scaffold.ProjectConfig{}

	fmt.Println("🚀 AgentFlow Project Setup")
	fmt.Println("==========================")

	// Project name
	fmt.Print("Project name: ")
	fmt.Scanln(&config.Name)

	// Basic configuration
	fmt.Print("Number of agents (default 2): ")
	var agentsInput string
	fmt.Scanln(&agentsInput)
	if agentsInput != "" {
		if parsed, err := strconv.Atoi(agentsInput); err == nil {
			config.NumAgents = parsed
		} else {
			config.NumAgents = 2
		}
	} else {
		config.NumAgents = 2
	}

	// Provider selection
	fmt.Println("Select LLM provider:")
	fmt.Println("1. OpenAI (default)")
	fmt.Println("2. Azure OpenAI")
	fmt.Println("3. Ollama (local)")
	fmt.Println("4. Mock (testing)")
	fmt.Print("Choice (1-4): ")
	var providerChoice string
	fmt.Scanln(&providerChoice)

	providers := map[string]string{
		"1": "openai", "2": "azure", "3": "ollama", "4": "mock",
		"": "openai", // default
	}
	if p, exists := providers[providerChoice]; exists {
		config.Provider = p
	} else {
		config.Provider = "openai"
	}

	// MCP integration
	fmt.Print("Enable MCP integration? (y/N): ")
	var mcpChoice string
	fmt.Scanln(&mcpChoice)
	config.MCPEnabled = strings.ToLower(mcpChoice) == "y" || strings.ToLower(mcpChoice) == "yes"

	if config.MCPEnabled {
		// MCP feature selection
		fmt.Print("Enable production MCP features? (y/N): ")
		var prodChoice string
		fmt.Scanln(&prodChoice)
		config.MCPProduction = strings.ToLower(prodChoice) == "y" || strings.ToLower(prodChoice) == "yes"

		fmt.Print("Enable MCP caching? (y/N): ")
		var cacheChoice string
		fmt.Scanln(&cacheChoice)
		config.WithCache = strings.ToLower(cacheChoice) == "y" || strings.ToLower(cacheChoice) == "yes"

		fmt.Print("Enable metrics? (y/N): ")
		var metricsChoice string
		fmt.Scanln(&metricsChoice)
		config.WithMetrics = strings.ToLower(metricsChoice) == "y" || strings.ToLower(metricsChoice) == "yes"

		// MCP tools
		fmt.Print("MCP tools (comma-separated, default: web_search,summarize): ")
		var toolsInput string
		fmt.Scanln(&toolsInput)
		if toolsInput != "" {
			config.MCPTools = parseCommaSeparatedList(toolsInput)
		} else {
			config.MCPTools = []string{"web_search", "summarize"}
		}

		// MCP servers
		fmt.Print("MCP servers (comma-separated, default: docker): ")
		var serversInput string
		fmt.Scanln(&serversInput)
		if serversInput != "" {
			config.MCPServers = parseCommaSeparatedList(serversInput)
		} else {
			config.MCPServers = []string{"docker"}
		}

		// Set defaults
		config.CacheBackend = "memory"
		config.MetricsPort = 8080
		config.ConnectionPoolSize = 5
		config.RetryPolicy = "exponential"
	}

	// Responsible AI and error handling (defaults to true)
	config.ResponsibleAI = true
	config.ErrorHandler = true

	fmt.Println("\n✓ Configuration complete!")
	return config, nil
}
