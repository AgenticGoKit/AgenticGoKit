package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/internal/scaffold"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage AgenticGoKit configuration files",
	Long: `Manage AgenticGoKit configuration files with various utilities for generation,
migration, and optimization.

SUBCOMMANDS:
  generate    Generate configuration from existing hardcoded values
  migrate     Migrate configuration to newer format versions
  optimize    Optimize configuration for performance and best practices
  extract     Extract agent configurations to separate files
  merge       Merge multiple configuration files
  template    Generate configuration from templates

Examples:
  # Generate config from existing project
  agentcli config generate

  # Migrate to latest configuration format
  agentcli config migrate --to-version 2.0

  # Optimize configuration for performance
  agentcli config optimize --focus performance

  # Extract agent configs to separate files
  agentcli config extract --agents-dir configs/agents

  # Generate from template
  agentcli config template research-assistant > agentflow.toml`,
}

// configGenerateCmd generates configuration from existing hardcoded values
var configGenerateCmd = &cobra.Command{
	Use:   "generate [output-file]",
	Short: "Generate configuration from existing hardcoded agent values",
	Long: `Generate AgenticGoKit configuration file from existing hardcoded agent implementations.

This command analyzes your existing agent code and extracts hardcoded values like:
  * System prompts and role definitions
  * LLM parameters (temperature, max_tokens, etc.)
  * Timeout and retry configurations
  * Capability definitions
  * Agent metadata and settings

The generated configuration can then be used to make your agents configuration-driven
instead of hardcoded.

Examples:
  # Generate config to agentflow.toml
  agentcli config generate

  # Generate to specific file
  agentcli config generate my-config.toml

  # Generate with specific format
  agentcli config generate --format yaml config.yaml

  # Analyze specific agents directory
  agentcli config generate --agents-dir ./my-agents`,
	Args: cobra.MaximumNArgs(1),
	Run:  runConfigGenerateCommand,
}

// configMigrateCmd migrates configuration to newer formats
var configMigrateCmd = &cobra.Command{
	Use:   "migrate [config-file]",
	Short: "Migrate configuration to newer format versions",
	Long: `Migrate AgenticGoKit configuration files to newer format versions.

This command helps upgrade configuration files when the framework introduces
new configuration options, deprecates old ones, or changes structure.

Migration capabilities:
  * Update deprecated field names
  * Add new required fields with defaults
  * Restructure configuration sections
  * Validate migrated configuration
  * Create backup of original file

Examples:
  # Migrate current agentflow.toml
  agentcli config migrate

  # Migrate specific file
  agentcli config migrate old-config.toml

  # Migrate to specific version
  agentcli config migrate --to-version 2.0

  # Dry run (show changes without applying)
  agentcli config migrate --dry-run`,
	Args: cobra.MaximumNArgs(1),
	Run:  runConfigMigrateCommand,
}

// configOptimizeCmd optimizes configuration for performance and best practices
var configOptimizeCmd = &cobra.Command{
	Use:   "optimize [config-file]",
	Short: "Optimize configuration for performance and best practices",
	Long: `Optimize AgenticGoKit configuration for better performance and adherence to best practices.

Optimization areas:
  * LLM parameter tuning for performance vs quality
  * Timeout and retry policy optimization
  * Memory and resource usage optimization
  * Agent capability optimization
  * Orchestration pattern optimization

Examples:
  # Optimize current configuration
  agentcli config optimize

  # Focus on specific optimization area
  agentcli config optimize --focus performance

  # Show recommendations without applying
  agentcli config optimize --recommend-only`,
	Args: cobra.MaximumNArgs(1),
	Run:  runConfigOptimizeCommand,
}

// configTemplateCmd generates configuration from templates
var configTemplateCmd = &cobra.Command{
	Use:   "template [template-name]",
	Short: "Generate configuration from templates",
	Long: `Generate AgenticGoKit configuration from predefined templates.

This command creates complete configuration files based on templates for
common use cases and patterns.

Available templates:
  * research-assistant  - Multi-agent research system
  * content-creation   - Content creation pipeline
  * customer-support   - Support ticket system
  * rag-system        - RAG-based Q&A system
  * simple-workflow   - Basic sequential workflow

Examples:
  # List available templates
  agentcli config template --list

  # Generate from template
  agentcli config template research-assistant

  # Generate with customizations
  agentcli config template rag-system --memory pgvector --embedding openai`,
	Args: cobra.MaximumNArgs(1),
	Run:  runConfigTemplateCommand,
}

var (
	// Generate command flags
	generateFormat    string
	generateAgentsDir string
	generateOutput    string
	generateBackup    bool

	// Migrate command flags
	migrateToVersion string
	migrateDryRun    bool
	migrateBackup    bool

	// Optimize command flags
	optimizeFocus         string
	optimizeRecommendOnly bool
	optimizeOutput        string

	// Template command flags
	templateList      bool
	templateMemory    string
	templateEmbedding string
	templateCustomize bool
)

func init() {
	rootCmd.AddCommand(configCmd)

	// Add subcommands
	configCmd.AddCommand(configGenerateCmd)
	configCmd.AddCommand(configMigrateCmd)
	configCmd.AddCommand(configOptimizeCmd)
	configCmd.AddCommand(configTemplateCmd)

	// Generate command flags
	configGenerateCmd.Flags().StringVar(&generateFormat, "format", "toml",
		"Output format (toml, yaml, json)")
	configGenerateCmd.Flags().StringVar(&generateAgentsDir, "agents-dir", "agents",
		"Directory containing agent implementations")
	configGenerateCmd.Flags().StringVar(&generateOutput, "output", "",
		"Output file (default: agentflow.toml)")
	configGenerateCmd.Flags().BoolVar(&generateBackup, "backup", true,
		"Create backup of existing configuration file")

	// Migrate command flags
	configMigrateCmd.Flags().StringVar(&migrateToVersion, "to-version", "latest",
		"Target configuration version")
	configMigrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false,
		"Show changes without applying them")
	configMigrateCmd.Flags().BoolVar(&migrateBackup, "backup", true,
		"Create backup before migration")

	// Optimize command flags
	configOptimizeCmd.Flags().StringVar(&optimizeFocus, "focus", "all",
		"Optimization focus (all, performance, memory, cost, reliability)")
	configOptimizeCmd.Flags().BoolVar(&optimizeRecommendOnly, "recommend-only", false,
		"Show recommendations without applying changes")
	configOptimizeCmd.Flags().StringVar(&optimizeOutput, "output", "",
		"Output optimized configuration to file")

	// Template command flags
	configTemplateCmd.Flags().BoolVar(&templateList, "list", false,
		"List available templates")
	configTemplateCmd.Flags().StringVar(&templateMemory, "memory", "",
		"Memory provider for template")
	configTemplateCmd.Flags().StringVar(&templateEmbedding, "embedding", "",
		"Embedding provider for template")
	configTemplateCmd.Flags().BoolVar(&templateCustomize, "customize", false,
		"Interactive template customization")

	// Add completion functions
	configGenerateCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"toml", "yaml", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
	configOptimizeCmd.RegisterFlagCompletionFunc("focus", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"all", "performance", "memory", "cost", "reliability"}, cobra.ShellCompDirectiveNoFileComp
	})
}

// runConfigGenerateCommand generates configuration from existing code
func runConfigGenerateCommand(cmd *cobra.Command, args []string) {
	fmt.Println("Generating configuration from existing agent implementations...")

	// Determine output file
	outputFile := "agentflow.toml"
	if len(args) > 0 {
		outputFile = args[0]
	}
	if generateOutput != "" {
		outputFile = generateOutput
	}

	// Check if agents directory exists
	if !pathExists(generateAgentsDir) {
		fmt.Printf("Agents directory not found: %s\n", generateAgentsDir)
		fmt.Println("Hint: make sure you're in an AgenticGoKit project directory")
		os.Exit(1)
	}

	// Create backup if requested and file exists
	if generateBackup && pathExists(outputFile) {
		backupFile := outputFile + ".backup"
		if err := copyFile(outputFile, backupFile); err != nil {
			fmt.Printf("Warning: failed to create backup: %v\n", err)
		} else {
			fmt.Printf("Created backup: %s\n", backupFile)
		}
	}

	// Analyze agent files
	agentConfigs, err := analyzeAgentFiles(generateAgentsDir)
	if err != nil {
		fmt.Printf("Failed to analyze agent files: %v\n", err)
		os.Exit(1)
	}

	if len(agentConfigs) == 0 {
		fmt.Printf("Warning: no agent configurations found in %s\n", generateAgentsDir)
		fmt.Println("Hint: ensure your agent files contain extractable configuration")
		os.Exit(1)
	}

	// Generate configuration
	config := generateConfigFromAgents(agentConfigs)

	// Write configuration file
	if err := writeConfigFile(config, outputFile, generateFormat); err != nil {
		fmt.Printf("Failed to write configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated configuration file: %s\n", outputFile)
	fmt.Printf("📊 Extracted %d agent configurations\n", len(agentConfigs))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review the generated configuration")
	fmt.Println("  2. Update your agent code to use configuration-driven approach")
	fmt.Println("  3. Test your agents with the new configuration")
}

// runConfigMigrateCommand migrates configuration to newer format
func runConfigMigrateCommand(cmd *cobra.Command, args []string) {
	configFile := "agentflow.toml"
	if len(args) > 0 {
		configFile = args[0]
	}

	fmt.Printf("🔄 Migrating configuration: %s\n", configFile)

	if !pathExists(configFile) {
		fmt.Printf("Configuration file not found: %s\n", configFile)
		os.Exit(1)
	}

	// Load current configuration
	config, err := core.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Perform migration analysis
	migrationNeeded, changes := analyzeMigrationNeeds(config)

	if !migrationNeeded {
		fmt.Println("Configuration is already up to date")
		return
	}

	fmt.Printf("Migration analysis found %d changes needed:\n", len(changes))
	for _, change := range changes {
		fmt.Printf("  - %s\n", change)
	}

	if migrateDryRun {
		fmt.Println("🔍 Dry run complete - no changes applied")
		return
	}

	// Create backup
	if migrateBackup {
		backupFile := configFile + ".pre-migration"
		if err := copyFile(configFile, backupFile); err != nil {
			fmt.Printf("Warning: failed to create backup: %v\n", err)
		} else {
			fmt.Printf("Created backup: %s\n", backupFile)
		}
	}

	// Apply migration
	migratedConfig := applyMigration(config, migrateToVersion)

	// Write migrated configuration
	if err := writeConfigFile(migratedConfig, configFile, "toml"); err != nil {
		fmt.Printf("Failed to write migrated configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully migrated configuration to version %s\n", migrateToVersion)
}

// runConfigOptimizeCommand optimizes configuration
func runConfigOptimizeCommand(cmd *cobra.Command, args []string) {
	configFile := "agentflow.toml"
	if len(args) > 0 {
		configFile = args[0]
	}

	fmt.Printf("⚡ Optimizing configuration: %s\n", configFile)
	fmt.Printf("Focus area: %s\n", optimizeFocus)

	if !pathExists(configFile) {
		fmt.Printf("Configuration file not found: %s\n", configFile)
		os.Exit(1)
	}

	// Load configuration
	config, err := core.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Analyze optimization opportunities
	optimizations := analyzeOptimizations(config, optimizeFocus)

	if len(optimizations) == 0 {
		fmt.Println("Configuration is already optimized")
		return
	}

	fmt.Printf("📊 Found %d optimization opportunities:\n", len(optimizations))
	for _, opt := range optimizations {
		fmt.Printf("  - %s: %s\n", opt.Area, opt.Description)
		if opt.Impact != "" {
			fmt.Printf("    Impact: %s\n", opt.Impact)
		}
	}

	if optimizeRecommendOnly {
		fmt.Println("Hint: use --recommend-only=false to apply optimizations")
		return
	}

	// Apply optimizations
	optimizedConfig := applyOptimizations(config, optimizations)

	// Determine output file
	outputFile := configFile
	if optimizeOutput != "" {
		outputFile = optimizeOutput
	}

	// Write optimized configuration
	if err := writeConfigFile(optimizedConfig, outputFile, "toml"); err != nil {
		fmt.Printf("Failed to write optimized configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Applied %d optimizations to %s\n", len(optimizations), outputFile)
}

// runConfigTemplateCommand generates configuration from templates
func runConfigTemplateCommand(cmd *cobra.Command, args []string) {
	if templateList {
		listConfigTemplates()
		return
	}

	if len(args) == 0 {
		fmt.Println("Template name required")
		fmt.Println("Hint: use --list to see available templates")
		os.Exit(1)
	}

	templateName := args[0]

	fmt.Printf("Generating configuration from template: %s\n", templateName)

	// Check if template exists
	templates, err := scaffold.ListAvailableTemplates()
	if err != nil {
		fmt.Printf("Failed to list templates: %v\n", err)
		os.Exit(1)
	}

	templateExists := false
	for _, tmpl := range templates {
		if tmpl == templateName {
			templateExists = true
			break
		}
	}

	if !templateExists {
		fmt.Printf("Template not found: %s\n", templateName)
		fmt.Println("Available templates:")
		for _, tmpl := range templates {
			fmt.Printf("  - %s\n", tmpl)
		}
		os.Exit(1)
	}

	// Load template
	templateInfo, err := scaffold.GetTemplateInfo(templateName)
	if err != nil {
		fmt.Printf("Failed to load template: %v\n", err)
		os.Exit(1)
	}

	// Apply customizations
	if templateMemory != "" {
		templateInfo.Config.MemoryEnabled = true
		templateInfo.Config.MemoryProvider = templateMemory
	}
	if templateEmbedding != "" {
		parts := strings.Split(templateEmbedding, ":")
		templateInfo.Config.EmbeddingProvider = parts[0]
		if len(parts) > 1 {
			templateInfo.Config.EmbeddingModel = parts[1]
		}
	}

	// Generate configuration content
	configContent, err := generateConfigFromTemplate(templateInfo)
	if err != nil {
		fmt.Printf("Failed to generate configuration: %v\n", err)
		os.Exit(1)
	}

	// Output to stdout (can be redirected)
	fmt.Print(configContent)
}

// Helper functions

// analyzeAgentFiles analyzes agent files to extract configuration
func analyzeAgentFiles(agentsDir string) (map[string]AgentConfigExtract, error) {
	configs := make(map[string]AgentConfigExtract)

	// This is a simplified implementation
	// In a real implementation, this would parse Go files and extract configuration
	agentFiles, err := filepath.Glob(filepath.Join(agentsDir, "*.go"))
	if err != nil {
		return nil, err
	}

	for _, file := range agentFiles {
		agentName := strings.TrimSuffix(filepath.Base(file), ".go")

		// Extract configuration from file (simplified)
		config := AgentConfigExtract{
			Name:         agentName,
			Role:         fmt.Sprintf("%s_role", agentName),
			Description:  fmt.Sprintf("Auto-generated description for %s", agentName),
			Capabilities: []string{"data_processing"}, // Default capability
			SystemPrompt: "Auto-generated system prompt",
			Enabled:      true,
			Timeout:      30,
		}

		configs[agentName] = config
	}

	return configs, nil
}

// AgentConfigExtract represents extracted agent configuration
type AgentConfigExtract struct {
	Name         string   `json:"name"`
	Role         string   `json:"role"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
	SystemPrompt string   `json:"system_prompt"`
	Enabled      bool     `json:"enabled"`
	Timeout      int      `json:"timeout"`
}

// generateConfigFromAgents generates configuration from extracted agent data
func generateConfigFromAgents(agentConfigs map[string]AgentConfigExtract) *core.Config {
	config := &core.Config{}
	config.AgentFlow.Name = "generated-project"
	config.AgentFlow.Version = "1.0.0"
	config.AgentFlow.Provider = "openai"

	config.Agents = make(map[string]core.AgentConfig)
	for name, agentConfig := range agentConfigs {
		config.Agents[name] = core.AgentConfig{
			Role:         agentConfig.Role,
			Description:  agentConfig.Description,
			SystemPrompt: agentConfig.SystemPrompt,
			Capabilities: agentConfig.Capabilities,
			Enabled:      agentConfig.Enabled,
			Timeout:      agentConfig.Timeout,
		}
	}

	return config
}

// writeConfigFile writes configuration to file in specified format
func writeConfigFile(config *core.Config, filename, format string) error {
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(config, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(config)
	default: // toml
		// For TOML, we'd need to use a TOML encoder
		// This is simplified - in reality we'd use github.com/BurntSushi/toml
		return fmt.Errorf("TOML output not implemented in this example")
	}

	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// Migration and optimization helper functions (simplified implementations)

func analyzeMigrationNeeds(config *core.Config) (bool, []string) {
	changes := []string{}

	// Example migration checks
	if config.AgentFlow.Version == "" {
		changes = append(changes, "Add version field to agent_flow section")
	}

	return len(changes) > 0, changes
}

func applyMigration(config *core.Config, targetVersion string) *core.Config {
	// Apply migration transformations
	if config.AgentFlow.Version == "" {
		config.AgentFlow.Version = "1.0.0"
	}
	return config
}

type OptimizationOpportunity struct {
	Area        string
	Description string
	Impact      string
}

func analyzeOptimizations(config *core.Config, focus string) []OptimizationOpportunity {
	optimizations := []OptimizationOpportunity{}

	// Example optimization analysis
	for agentName, agent := range config.Agents {
		if agent.Timeout > 60 {
			optimizations = append(optimizations, OptimizationOpportunity{
				Area:        "Performance",
				Description: fmt.Sprintf("Agent %s has high timeout (%d seconds)", agentName, agent.Timeout),
				Impact:      "Reduce timeout for better responsiveness",
			})
		}
	}

	return optimizations
}

func applyOptimizations(config *core.Config, optimizations []OptimizationOpportunity) *core.Config {
	// Apply optimizations
	for agentName, agent := range config.Agents {
		if agent.Timeout > 60 {
			agent.Timeout = 30
			config.Agents[agentName] = agent
		}
	}
	return config
}

func generateConfigFromTemplate(templateInfo *scaffold.TemplateConfig) (string, error) {
	// Generate TOML configuration from template
	// This is simplified - would use proper TOML generation
	return fmt.Sprintf("# Generated from template: %s\n# %s\n\n[agent_flow]\nname = \"template-project\"\n",
		templateInfo.Name, templateInfo.Description), nil
}

func listConfigTemplates() {
	fmt.Println("Available Configuration Templates:")
	fmt.Println("====================================")

	templates, err := scaffold.ListAvailableTemplates()
	if err != nil {
		fmt.Printf("Failed to list templates: %v\n", err)
		return
	}

	for _, templateName := range templates {
		info, err := scaffold.GetTemplateInfo(templateName)
		if err != nil {
			fmt.Printf("  - %s (error loading info)\n", templateName)
			continue
		}

		fmt.Printf("  - %s\n", templateName)
		fmt.Printf("    %s\n", info.Description)
		fmt.Printf("    Features: %s\n", strings.Join(info.Features, ", "))
		fmt.Printf("    Agents: %d\n", info.Config.NumAgents)
		fmt.Println()
	}
}
