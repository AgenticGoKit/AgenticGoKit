package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// helpCmd provides enhanced help with examples and workflows
var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Help about any command with examples",
	Long: `Get detailed help about AgentCLI commands with practical examples and workflows.

This enhanced help system provides:
- Command usage examples
- Common workflows
- Configuration guidance
- Troubleshooting tips`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			showMainHelp()
		} else {
			showCommandHelp(args[0])
		}
	},
}

func showMainHelp() {
	fmt.Println("AgentCLI - AgentFlow Configuration Management")
	fmt.Println("=" + strings.Repeat("=", 48))
	fmt.Println()

	fmt.Println("QUICK START:")
	fmt.Println("  1. Create a new project:")
	fmt.Println("     agentcli create my-project --template simple-workflow")
	fmt.Println()
	fmt.Println("  2. Generate configuration only:")
	fmt.Println("     agentcli config generate simple-workflow my-config")
	fmt.Println()
	fmt.Println("  3. Validate configuration:")
	fmt.Println("     agentcli validate agentflow.toml")
	fmt.Println()

	fmt.Println("MAIN COMMANDS:")
	fmt.Println("  create      Create new AgentFlow projects")
	fmt.Println("  config      Configuration management utilities")
	fmt.Println("  template    Template management and validation")
	fmt.Println("  validate    Validate configuration files")
	fmt.Println("  version     Show version information")
	fmt.Println()

	fmt.Println("CONFIGURATION WORKFLOW:")
	fmt.Println("  1. List available templates:")
	fmt.Println("     agentcli template list --detailed")
	fmt.Println()
	fmt.Println("  2. Get template information:")
	fmt.Println("     agentcli template info research-assistant")
	fmt.Println()
	fmt.Println("  3. Generate configuration:")
	fmt.Println("     agentcli config generate research-assistant my-research-bot")
	fmt.Println()
	fmt.Println("  4. Validate configuration:")
	fmt.Println("     agentcli validate --detailed agentflow.toml")
	fmt.Println()
	fmt.Println("  5. Create full project:")
	fmt.Println("     agentcli create my-project --template research-assistant")
	fmt.Println()

	fmt.Println("COMMON USE CASES:")
	fmt.Println("  - Quick prototype:")
	fmt.Println("    agentcli create demo --template simple-workflow")
	fmt.Println()
	fmt.Println("  - Research system:")
	fmt.Println("    agentcli create research-bot --template research-assistant")
	fmt.Println()
	fmt.Println("  - Content creation:")
	fmt.Println("    agentcli create content-pipeline --template content-creation")
	fmt.Println()
	fmt.Println("  - Customer support:")
	fmt.Println("    agentcli create support-system --template customer-support")
	fmt.Println()

	fmt.Println("TIPS:")
	fmt.Println("  - Use --help with any command for detailed options")
	fmt.Println("  - Use --format json/yaml for structured output")
	fmt.Println("  - Use --detailed for comprehensive information")
	fmt.Println("  - Configuration files support TOML, YAML, and JSON formats")
	fmt.Println()

	fmt.Println("MORE HELP:")
	fmt.Println("  agentcli help create     - Project creation help")
	fmt.Println("  agentcli help config     - Configuration management help")
	fmt.Println("  agentcli help template   - Template management help")
	fmt.Println("  agentcli help validate   - Validation help")
}

func showCommandHelp(command string) {
	switch command {
	case "create":
		showCreateHelp()
	case "config":
		showConfigHelp()
	case "template":
		showTemplateHelp()
	case "validate":
		showValidateHelp()
	default:
		fmt.Printf("No detailed help available for command: %s\n", command)
		fmt.Println("Available commands: create, config, template, validate")
	}
}

func showCreateHelp() {
	fmt.Println("PROJECT CREATION HELP")
	fmt.Println("=" + strings.Repeat("=", 24))
	fmt.Println()

	fmt.Println("BASIC USAGE:")
	fmt.Println("  agentcli create <project-name> [flags]")
	fmt.Println()

	fmt.Println("EXAMPLES:")
	fmt.Println("  # Create with template")
	fmt.Println("  agentcli create my-bot --template simple-workflow")
	fmt.Println()
	fmt.Println("  # Create in specific directory")
	fmt.Println("  agentcli create my-bot --output-dir ./projects")
	fmt.Println()
	fmt.Println("  # Create with custom configuration")
	fmt.Println("  agentcli create my-bot --config ./custom-config.toml")
	fmt.Println()

	fmt.Println("AVAILABLE TEMPLATES:")
	fmt.Println("  - simple-workflow     - Basic three-agent sequential workflow")
	fmt.Println("  - research-assistant  - Advanced research system with RAG")
	fmt.Println("  - content-creation    - Content creation pipeline with SEO")
	fmt.Println("  - customer-support    - Collaborative support system")
	fmt.Println("  - custom-rag          - Enhanced RAG system")
	fmt.Println()

	fmt.Println("FLAGS:")
	fmt.Println("  --template, -t        Template to use")
	fmt.Println("  --output-dir, -o      Output directory")
	fmt.Println("  --config, -c          Custom configuration file")
	fmt.Println("  --force, -f           Overwrite existing files")
	fmt.Println()

	fmt.Println("WORKFLOW:")
	fmt.Println("  1. Choose a template: agentcli template list")
	fmt.Println("  2. Review template: agentcli template info <name>")
	fmt.Println("  3. Create project: agentcli create <name> --template <template>")
	fmt.Println("  4. Validate result: agentcli validate <project>/agentflow.toml")
}

func showConfigHelp() {
	fmt.Println("CONFIGURATION MANAGEMENT HELP")
	fmt.Println("=" + strings.Repeat("=", 32))
	fmt.Println()

	fmt.Println("SUBCOMMANDS:")
	fmt.Println("  generate    Generate configuration from templates")
	fmt.Println("  migrate     Migrate existing configurations")
	fmt.Println("  schema      Schema management operations")
	fmt.Println()

	fmt.Println("GENERATE EXAMPLES:")
	fmt.Println("  # Generate TOML configuration")
	fmt.Println("  agentcli config generate simple-workflow my-project")
	fmt.Println()
	fmt.Println("  # Generate in specific directory")
	fmt.Println("  agentcli config generate research-assistant bot --output-dir ./configs")
	fmt.Println()
	fmt.Println("  # Generate as YAML")
	fmt.Println("  agentcli config generate content-creation blog --format yaml")
	fmt.Println()
	fmt.Println("  # List available templates")
	fmt.Println("  agentcli config generate --list")
	fmt.Println()

	fmt.Println("MIGRATE EXAMPLES:")
	fmt.Println("  # Dry run migration")
	fmt.Println("  agentcli config migrate --dry-run agentflow.toml")
	fmt.Println()
	fmt.Println("  # Migrate with backup")
	fmt.Println("  agentcli config migrate --backup-dir ./backups agentflow.toml")
	fmt.Println()

	fmt.Println("SCHEMA EXAMPLES:")
	fmt.Println("  # Generate JSON schema")
	fmt.Println("  agentcli config schema --generate")
	fmt.Println()
	fmt.Println("  # Export documentation")
	fmt.Println("  agentcli config schema --export-docs")
	fmt.Println()

	fmt.Println("SUPPORTED FORMATS:")
	fmt.Println("  - TOML (default)      - Human-readable configuration")
	fmt.Println("  - YAML               - Alternative human-readable format")
	fmt.Println("  - JSON               - Machine-readable format")
}

func showTemplateHelp() {
	fmt.Println("TEMPLATE MANAGEMENT HELP")
	fmt.Println("=" + strings.Repeat("=", 27))
	fmt.Println()

	fmt.Println("SUBCOMMANDS:")
	fmt.Println("  list        List available templates")
	fmt.Println("  info        Show detailed template information")
	fmt.Println("  create      Create new custom template")
	fmt.Println("  validate    Validate template configuration")
	fmt.Println("  paths       Show template search paths")
	fmt.Println()

	fmt.Println("LIST EXAMPLES:")
	fmt.Println("  # Basic list")
	fmt.Println("  agentcli template list")
	fmt.Println()
	fmt.Println("  # Detailed information")
	fmt.Println("  agentcli template list --detailed")
	fmt.Println()
	fmt.Println("  # JSON output")
	fmt.Println("  agentcli template list --format json")
	fmt.Println()

	fmt.Println("INFO EXAMPLES:")
	fmt.Println("  # Show template details")
	fmt.Println("  agentcli template info research-assistant")
	fmt.Println()
	fmt.Println("  # YAML output")
	fmt.Println("  agentcli template info content-creation --format yaml")
	fmt.Println()

	fmt.Println("CREATE EXAMPLES:")
	fmt.Println("  # Create YAML template")
	fmt.Println("  agentcli template create my-custom-template")
	fmt.Println()
	fmt.Println("  # Create JSON template")
	fmt.Println("  agentcli template create my-template --format json")
	fmt.Println()
	fmt.Println("  # Custom output location")
	fmt.Println("  agentcli template create my-template --output ./templates/my-template.yaml")
	fmt.Println()

	fmt.Println("VALIDATE EXAMPLES:")
	fmt.Println("  # Validate specific template")
	fmt.Println("  agentcli template validate my-template.yaml")
	fmt.Println()
	fmt.Println("  # Validate all templates")
	fmt.Println("  agentcli template validate")
	fmt.Println()

	fmt.Println("TEMPLATE LOCATIONS:")
	fmt.Println("  - Current directory: .agenticgokit/templates/")
	fmt.Println("  - User home: ~/.agenticgokit/templates/")
	fmt.Println("  - System-wide: /etc/agenticgokit/templates/")
}

func showValidateHelp() {
	fmt.Println("VALIDATION HELP")
	fmt.Println("=" + strings.Repeat("=", 17))
	fmt.Println()

	fmt.Println("BASIC USAGE:")
	fmt.Println("  agentcli validate [config-file] [flags]")
	fmt.Println()

	fmt.Println("EXAMPLES:")
	fmt.Println("  # Basic validation")
	fmt.Println("  agentcli validate agentflow.toml")
	fmt.Println()
	fmt.Println("  # Detailed validation report")
	fmt.Println("  agentcli validate --detailed agentflow.toml")
	fmt.Println()
	fmt.Println("  # JSON output")
	fmt.Println("  agentcli validate --format json agentflow.toml")
	fmt.Println()
	fmt.Println("  # Quiet mode (errors only)")
	fmt.Println("  agentcli validate --quiet agentflow.toml")
	fmt.Println()

	fmt.Println("VALIDATION CHECKS:")
	fmt.Println("  ✓ TOML/YAML/JSON syntax and structure")
	fmt.Println("  ✓ Required fields and sections")
	fmt.Println("  ✓ Agent configuration validity")
	fmt.Println("  ✓ LLM provider settings")
	fmt.Println("  ✓ Orchestration configuration")
	fmt.Println("  ✓ Cross-reference validation")
	fmt.Println("  ✓ Performance optimization suggestions")
	fmt.Println("  ✓ Security and compliance checks")
	fmt.Println()

	fmt.Println("OUTPUT FORMATS:")
	fmt.Println("  - text (default)     - Human-readable validation report")
	fmt.Println("  - json              - Machine-readable JSON output")
	fmt.Println("  - yaml              - YAML structured output")
	fmt.Println()

	fmt.Println("FLAGS:")
	fmt.Println("  --detailed, -d       Show detailed validation report")
	fmt.Println("  --format, -f         Output format: text, json, yaml")
	fmt.Println("  --quiet, -q          Suppress non-error output")
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
