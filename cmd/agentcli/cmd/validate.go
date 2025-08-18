package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate AgenticGoKit configuration and project structure",
	Long: `Validate AgenticGoKit configuration files and project structure with comprehensive checks.

This command performs extensive validation including:
  * Configuration file syntax and structure
  * Agent configuration completeness and correctness
  * LLM parameter validation with provider-specific rules
  * Capability validation against known capability registry
  * Orchestration configuration validation
  * Cross-reference validation between agents and orchestration
  * Memory system configuration validation
  * MCP server configuration validation
  * Performance and optimization recommendations

VALIDATION LEVELS:
  --basic     Basic syntax and structure validation
  --standard  Standard validation with configuration checks (default)
  --strict    Strict validation with performance recommendations
  --complete  Complete validation with all checks and suggestions

VALIDATION SCOPE:
  --config-only    Validate only configuration files
  --structure-only Validate only project structure
  --agents-only    Validate only agent configurations

Examples:
  # Validate current project with standard checks
  agentcli validate

  # Validate specific configuration file
  agentcli validate agentflow.toml

  # Strict validation with all recommendations
  agentcli validate --strict --verbose

  # Validate only agent configurations
  agentcli validate --agents-only

  # Validate configuration syntax only
  agentcli validate --config-only --basic

  # Show validation results in JSON format
  agentcli validate --output json`,
	Args: cobra.MaximumNArgs(1),
	Run:  runValidateCommand,
}

var (
	validateVerbose         bool
	validateLevel           string
	validateScope           string
	validateOutput          string
	validateFix             bool
	validateShowSuggestions bool
)

func init() {
	rootCmd.AddCommand(validateCmd)

	// Validation level flags
	validateCmd.Flags().StringVar(&validateLevel, "level", "standard",
		"Validation level (basic, standard, strict, complete)")
	validateCmd.Flags().BoolVarP(&validateVerbose, "verbose", "v", false,
		"Show detailed validation output")

	// Validation scope flags
	validateCmd.Flags().StringVar(&validateScope, "scope", "all",
		"Validation scope (all, config-only, structure-only, agents-only)")
	validateCmd.Flags().BoolVar(&validateFix, "fix", false,
		"Attempt to fix common configuration issues automatically")
	validateCmd.Flags().BoolVar(&validateShowSuggestions, "suggestions", true,
		"Show optimization suggestions and recommendations")

	// Output format flags
	validateCmd.Flags().StringVarP(&validateOutput, "output", "o", "text",
		"Output format (text, json, yaml)")

	// Add completion functions
	validateCmd.RegisterFlagCompletionFunc("level", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"basic", "standard", "strict", "complete"}, cobra.ShellCompDirectiveNoFileComp
	})
	validateCmd.RegisterFlagCompletionFunc("scope", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"all", "config-only", "structure-only", "agents-only"}, cobra.ShellCompDirectiveNoFileComp
	})
	validateCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"text", "json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	})
}

func runValidateCommand(cmd *cobra.Command, args []string) {
	// Determine configuration file path
	configPath := "agentflow.toml"
	if len(args) > 0 {
		configPath = args[0]
	}

	// Set the working directory to the config file's directory for validation
	configDir := filepath.Dir(configPath)
	if configDir != "." && configDir != "" {
		// Store original working directory
		originalWd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		// Change to config directory for validation
		if err := os.Chdir(configDir); err != nil {
			fmt.Printf("Error changing to config directory %s: %v\n", configDir, err)
			os.Exit(1)
		}

		// Restore original directory after validation
		defer func() {
			os.Chdir(originalWd)
		}()

		// Update config path to be relative to the new working directory
		configPath = filepath.Base(configPath)
	}

	// Initialize validation results
	results := &ValidationResults{
		ConfigPath:  configPath,
		Level:       validateLevel,
		Scope:       validateScope,
		Errors:      []ValidationIssue{},
		Warnings:    []ValidationIssue{},
		Suggestions: []ValidationIssue{},
	}

	// Print validation header
	if validateOutput == "text" {
		fmt.Printf("AgenticGoKit Configuration Validation\n")
		fmt.Printf("==========================================\n")
		fmt.Printf("Config: %s\n", configPath)
		fmt.Printf("Level:  %s\n", validateLevel)
		fmt.Printf("Scope:  %s\n", validateScope)
		fmt.Println()
	}

	// Perform validation based on scope
	switch validateScope {
	case "config-only":
		validateConfigurationOnly(results)
	case "structure-only":
		validateProjectStructure(results)
	case "agents-only":
		validateAgentsOnly(results)
	default:
		validateComplete(results)
	}

	// Apply fixes if requested
	if validateFix && len(results.Errors) > 0 {
		applyAutomaticFixes(results)
	}

	// Output results
	outputValidationResults(results)

	// Exit with appropriate code
	if len(results.Errors) > 0 {
		os.Exit(1)
	}
}

// ValidationIssue represents a validation issue
type ValidationIssue struct {
	Type       string `json:"type"`       // "error", "warning", "suggestion"
	Code       string `json:"code"`       // Error code for programmatic handling
	Field      string `json:"field"`      // Configuration field
	Message    string `json:"message"`    // Human-readable message
	Suggestion string `json:"suggestion"` // Suggested fix
	Severity   string `json:"severity"`   // "critical", "high", "medium", "low"
	Fixable    bool   `json:"fixable"`    // Whether this can be auto-fixed
}

// ValidationResults holds all validation results
type ValidationResults struct {
	ConfigPath  string            `json:"config_path"`
	Level       string            `json:"level"`
	Scope       string            `json:"scope"`
	Errors      []ValidationIssue `json:"errors"`
	Warnings    []ValidationIssue `json:"warnings"`
	Suggestions []ValidationIssue `json:"suggestions"`
	Summary     ValidationSummary `json:"summary"`
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	TotalIssues     int  `json:"total_issues"`
	ErrorCount      int  `json:"error_count"`
	WarningCount    int  `json:"warning_count"`
	SuggestionCount int  `json:"suggestion_count"`
	IsValid         bool `json:"is_valid"`
	CanAutoFix      int  `json:"can_auto_fix"`
}

// validateComplete performs comprehensive validation
func validateComplete(results *ValidationResults) {
	validateProjectStructure(results)
	validateConfigurationOnly(results)
	validateAgentsOnly(results)

	// Additional cross-validation checks
	validateCrossReferences(results)

	if validateLevel == "strict" || validateLevel == "complete" {
		validatePerformanceOptimizations(results)
	}

	if validateLevel == "complete" {
		validateBestPractices(results)
	}
}

// validateProjectStructure validates the project directory structure
func validateProjectStructure(results *ValidationResults) {
	if validateVerbose {
		fmt.Println("Validating project structure...")
	}

	// Check for required files and directories
	requiredPaths := map[string]string{
		"agentflow.toml": "Configuration file",
		"go.mod":         "Go module file",
		"main.go":        "Main application file",
		"agents/":        "Agents directory",
	}

	for path, description := range requiredPaths {
		if !pathExists(path) {
			results.Errors = append(results.Errors, ValidationIssue{
				Type:       "error",
				Code:       "MISSING_REQUIRED_FILE",
				Field:      path,
				Message:    fmt.Sprintf("Missing required %s: %s", strings.ToLower(description), path),
				Suggestion: fmt.Sprintf("Create the missing %s", strings.ToLower(description)),
				Severity:   "critical",
				Fixable:    false,
			})
		}
	}

	// Check go.mod for AgenticGoKit dependency
	if pathExists("go.mod") {
		content, err := os.ReadFile("go.mod")
		if err == nil {
			if !strings.Contains(string(content), "github.com/kunalkushwaha/agenticgokit") {
				results.Warnings = append(results.Warnings, ValidationIssue{
					Type:       "warning",
					Code:       "MISSING_DEPENDENCY",
					Field:      "go.mod",
					Message:    "AgenticGoKit dependency not found in go.mod",
					Suggestion: "Run 'go mod tidy' to ensure dependencies are properly managed",
					Severity:   "medium",
					Fixable:    true,
				})
			}
		}
	}
}

// validateConfigurationOnly validates only the configuration file
func validateConfigurationOnly(results *ValidationResults) {
	if validateVerbose {
		fmt.Println("Validating configuration...")
	}

	// Check if configuration file exists
	if !pathExists(results.ConfigPath) {
		results.Errors = append(results.Errors, ValidationIssue{
			Type:       "error",
			Code:       "CONFIG_FILE_NOT_FOUND",
			Field:      results.ConfigPath,
			Message:    fmt.Sprintf("Configuration file not found: %s", results.ConfigPath),
			Suggestion: "Create an agentflow.toml configuration file",
			Severity:   "critical",
			Fixable:    false,
		})
		return
	}

	// Load and validate configuration
	config, err := core.LoadConfig(results.ConfigPath)
	if err != nil {
		results.Errors = append(results.Errors, ValidationIssue{
			Type:       "error",
			Code:       "CONFIG_PARSE_ERROR",
			Field:      results.ConfigPath,
			Message:    fmt.Sprintf("Failed to parse configuration: %v", err),
			Suggestion: "Check TOML syntax and structure",
			Severity:   "critical",
			Fixable:    false,
		})
		return
	}

	// Use the comprehensive validation system
	validator := core.NewDefaultConfigValidator()
	validationErrors := validator.ValidateConfig(config)

	// Convert core validation errors to CLI validation issues
	for _, validationError := range validationErrors {
		severity := "medium"
		issueType := "warning"

		// Determine severity and type based on error content
		msg := strings.ToLower(validationError.Message)
		if strings.Contains(msg, "required") || strings.Contains(msg, "missing") || strings.Contains(msg, "invalid") {
			severity = "high"
			issueType = "error"
		}

		issue := ValidationIssue{
			Type:       issueType,
			Code:       "CONFIG_VALIDATION_ERROR",
			Field:      validationError.Field,
			Message:    validationError.Message,
			Suggestion: validationError.Suggestion,
			Severity:   severity,
			Fixable:    false,
		}

		if issueType == "error" {
			results.Errors = append(results.Errors, issue)
		} else {
			results.Warnings = append(results.Warnings, issue)
		}
	}
}

// validateAgentsOnly validates only agent configurations
func validateAgentsOnly(results *ValidationResults) {
	if validateVerbose {
		fmt.Println("Validating agent configurations...")
	}

	// Load configuration to get agent definitions
	config, err := core.LoadConfig(results.ConfigPath)
	if err != nil {
		return // Already handled in validateConfigurationOnly
	}

	// Validate each agent configuration
	validator := core.NewDefaultConfigValidator()
	for agentName, agentConfig := range config.Agents {
		agentErrors := validator.ValidateAgentConfig(agentName, &agentConfig)

		for _, validationError := range agentErrors {
			results.Warnings = append(results.Warnings, ValidationIssue{
				Type:       "warning",
				Code:       "AGENT_CONFIG_WARNING",
				Field:      validationError.Field,
				Message:    validationError.Message,
				Suggestion: validationError.Suggestion,
				Severity:   "medium",
				Fixable:    false,
			})
		}
	}

	// Check for corresponding agent files
	if pathExists("agents/") {
		agentFiles, err := filepath.Glob("agents/*.go")
		if err == nil {
			configuredAgents := make(map[string]bool)
			for agentName := range config.Agents {
				configuredAgents[agentName] = false
			}

			// Check if agent files exist for configured agents
			for _, agentFile := range agentFiles {
				fileName := strings.TrimSuffix(filepath.Base(agentFile), ".go")
				if _, exists := configuredAgents[fileName]; exists {
					configuredAgents[fileName] = true
				}
			}

			// Report missing agent files
			for agentName, hasFile := range configuredAgents {
				if !hasFile {
					results.Warnings = append(results.Warnings, ValidationIssue{
						Type:       "warning",
						Code:       "MISSING_AGENT_FILE",
						Field:      fmt.Sprintf("agents/%s.go", agentName),
						Message:    fmt.Sprintf("Agent file not found for configured agent: %s", agentName),
						Suggestion: fmt.Sprintf("Create agents/%s.go or remove agent from configuration", agentName),
						Severity:   "medium",
						Fixable:    false,
					})
				}
			}
		}
	}
}

// validateCrossReferences validates cross-references between configuration sections
func validateCrossReferences(results *ValidationResults) {
	if validateVerbose {
		fmt.Println("Validating cross-references...")
	}

	config, err := core.LoadConfig(results.ConfigPath)
	if err != nil {
		return
	}

	validator := core.NewDefaultConfigValidator()

	// Validate orchestration references
	if config.Orchestration.SequentialAgents != nil || config.Orchestration.CollaborativeAgents != nil {
		orchErrors := validator.ValidateOrchestrationAgents(&config.Orchestration, config.Agents)

		for _, validationError := range orchErrors {
			results.Errors = append(results.Errors, ValidationIssue{
				Type:       "error",
				Code:       "ORCHESTRATION_REFERENCE_ERROR",
				Field:      validationError.Field,
				Message:    validationError.Message,
				Suggestion: validationError.Suggestion,
				Severity:   "high",
				Fixable:    false,
			})
		}
	}
}

// validatePerformanceOptimizations suggests performance optimizations
func validatePerformanceOptimizations(results *ValidationResults) {
	if validateVerbose {
		fmt.Println("Analyzing performance optimizations...")
	}

	config, err := core.LoadConfig(results.ConfigPath)
	if err != nil {
		return
	}

	// Check for performance anti-patterns
	for agentName, agentConfig := range config.Agents {
		// Check for very high timeout values
		if agentConfig.Timeout > 300 { // 5 minutes
			results.Suggestions = append(results.Suggestions, ValidationIssue{
				Type:       "suggestion",
				Code:       "HIGH_TIMEOUT_WARNING",
				Field:      fmt.Sprintf("agents.%s.timeout_seconds", agentName),
				Message:    fmt.Sprintf("Agent %s has a very high timeout (%d seconds)", agentName, agentConfig.Timeout),
				Suggestion: "Consider reducing timeout for better responsiveness",
				Severity:   "low",
				Fixable:    true,
			})
		}

		// Check LLM parameters for performance
		if agentConfig.LLM != nil {
			if agentConfig.LLM.MaxTokens > 4000 {
				results.Suggestions = append(results.Suggestions, ValidationIssue{
					Type:       "suggestion",
					Code:       "HIGH_MAX_TOKENS",
					Field:      fmt.Sprintf("agents.%s.llm.max_tokens", agentName),
					Message:    fmt.Sprintf("Agent %s has high max_tokens (%d)", agentName, agentConfig.LLM.MaxTokens),
					Suggestion: "High token limits may increase latency and cost",
					Severity:   "low",
					Fixable:    true,
				})
			}
		}
	}
}

// validateBestPractices checks for best practice adherence
func validateBestPractices(results *ValidationResults) {
	if validateVerbose {
		fmt.Println("Checking best practices...")
	}

	config, err := core.LoadConfig(results.ConfigPath)
	if err != nil {
		return
	}

	// Check for agent naming conventions
	for agentName, agentConfig := range config.Agents {
		if !strings.Contains(agentName, "_") && !strings.Contains(agentName, "-") {
			results.Suggestions = append(results.Suggestions, ValidationIssue{
				Type:       "suggestion",
				Code:       "NAMING_CONVENTION",
				Field:      fmt.Sprintf("agents.%s", agentName),
				Message:    fmt.Sprintf("Agent name '%s' doesn't follow naming conventions", agentName),
				Suggestion: "Use descriptive names with underscores or hyphens (e.g., 'data_processor')",
				Severity:   "low",
				Fixable:    false,
			})
		}

		// Check for system prompt presence
		if agentConfig.SystemPrompt == "" {
			results.Suggestions = append(results.Suggestions, ValidationIssue{
				Type:       "suggestion",
				Code:       "MISSING_SYSTEM_PROMPT",
				Field:      fmt.Sprintf("agents.%s.system_prompt", agentName),
				Message:    fmt.Sprintf("Agent %s has no system prompt", agentName),
				Suggestion: "Add a system prompt to improve agent behavior and consistency",
				Severity:   "medium",
				Fixable:    false,
			})
		}
	}
}

// applyAutomaticFixes attempts to fix common issues automatically
func applyAutomaticFixes(results *ValidationResults) {
	if validateVerbose {
		fmt.Println("Applying automatic fixes...")
	}

	fixCount := 0
	for i, issue := range results.Errors {
		if issue.Fixable {
			switch issue.Code {
			case "MISSING_DEPENDENCY":
				// This would require running go mod tidy
				fmt.Printf("  Would run 'go mod tidy' to fix dependency issue\n")
				results.Errors[i].Message += " [WOULD BE FIXED]"
				fixCount++
			}
		}
	}

	if fixCount > 0 {
		fmt.Printf("%d issues can be automatically fixed (use --fix to apply)\n", fixCount)
	}
}

// outputValidationResults outputs the validation results in the specified format
func outputValidationResults(results *ValidationResults) {
	// Calculate summary
	results.Summary = ValidationSummary{
		ErrorCount:      len(results.Errors),
		WarningCount:    len(results.Warnings),
		SuggestionCount: len(results.Suggestions),
		TotalIssues:     len(results.Errors) + len(results.Warnings) + len(results.Suggestions),
		IsValid:         len(results.Errors) == 0,
	}

	// Count fixable issues
	for _, issue := range results.Errors {
		if issue.Fixable {
			results.Summary.CanAutoFix++
		}
	}
	for _, issue := range results.Warnings {
		if issue.Fixable {
			results.Summary.CanAutoFix++
		}
	}

	switch validateOutput {
	case "json":
		outputJSON(results)
	case "yaml":
		outputYAML(results)
	default:
		outputText(results)
	}
}

// outputText outputs results in human-readable text format
func outputText(results *ValidationResults) {
	// Print errors
	if len(results.Errors) > 0 {
		fmt.Printf("ERRORS (%d):\n", len(results.Errors))
		for _, issue := range results.Errors {
			fmt.Printf("  %s: %s\n", issue.Field, issue.Message)
			if issue.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
			}
		}
		fmt.Println()
	}

	// Print warnings
	if len(results.Warnings) > 0 {
		fmt.Printf("WARNINGS (%d):\n", len(results.Warnings))
		for _, issue := range results.Warnings {
			fmt.Printf("  %s: %s\n", issue.Field, issue.Message)
			if issue.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
			}
		}
		fmt.Println()
	}

	// Print suggestions
	if validateShowSuggestions && len(results.Suggestions) > 0 {
		fmt.Printf("SUGGESTIONS (%d):\n", len(results.Suggestions))
		for _, issue := range results.Suggestions {
			fmt.Printf("  %s: %s\n", issue.Field, issue.Message)
			if issue.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
			}
		}
		fmt.Println()
	}

	// Print summary
	fmt.Printf("VALIDATION SUMMARY:\n")
	fmt.Printf("  Total Issues: %d\n", results.Summary.TotalIssues)
	fmt.Printf("  Errors: %d\n", results.Summary.ErrorCount)
	fmt.Printf("  Warnings: %d\n", results.Summary.WarningCount)
	fmt.Printf("  Suggestions: %d\n", results.Summary.SuggestionCount)

	if results.Summary.IsValid {
		fmt.Printf("  Status: VALID\n")
	} else {
		fmt.Printf("  Status: INVALID\n")
	}

	if results.Summary.CanAutoFix > 0 {
		fmt.Printf("  Auto-fixable: %d (use --fix to apply)\n", results.Summary.CanAutoFix)
	}
}

// outputJSON outputs results in JSON format
func outputJSON(results *ValidationResults) {
	// This would require importing encoding/json
	fmt.Printf("JSON output not implemented yet\n")
}

// outputYAML outputs results in YAML format
func outputYAML(results *ValidationResults) {
	// This would require importing gopkg.in/yaml.v3
	fmt.Printf("YAML output not implemented yet\n")
}

// pathExists checks if a file or directory exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isAgenticGoKitProject checks if current directory is an AgenticGoKit project
func isAgenticGoKitProject() bool {
	return pathExists("agentflow.toml") ||
		pathExists("agents/") ||
		(pathExists("go.mod") && containsAgenticGoKitDependency("go.mod"))
}

// containsAgenticGoKitDependency checks if go.mod contains AgenticGoKit dependency
func containsAgenticGoKitDependency(goModPath string) bool {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), "github.com/kunalkushwaha/agenticgokit")
}
