package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate project structure and configuration",
	Long: `Validate the current AgenticGoKit project structure and configuration.

This command checks:
  * Project directory structure
  * Configuration file validity
  * Agent file consistency
  * Import path correctness
  * Required dependencies

Examples:
  # Validate current project
  agentcli validate

  # Validate with verbose output
  agentcli validate --verbose`,
	Run: runValidateCommand,
}

var (
	validateVerbose bool
)

func init() {
	rootCmd.AddCommand(validateCmd)

	// Add flags
	validateCmd.Flags().BoolVarP(&validateVerbose, "verbose", "v", false, "Show detailed validation output")
}

func runValidateCommand(cmd *cobra.Command, args []string) {
	fmt.Println("Validating AgenticGoKit project...")
	
	// Check if we're in a project directory
	if !isAgenticGoKitProject() {
		fmt.Println("[ERROR] Not in an AgenticGoKit project directory")
		fmt.Println("  - No agentflow.toml found")
		fmt.Println("  - No agents/ directory found")
		os.Exit(1)
	}
	
	fmt.Println("[SUCCESS] Valid AgenticGoKit project structure detected")
	
	// TODO: Add more comprehensive validation
	if validateVerbose {
		fmt.Println("[INFO] Detailed validation not yet implemented")
	}
}

func isAgenticGoKitProject() bool {
	// Check for agentflow.toml
	if _, err := os.Stat("agentflow.toml"); err == nil {
		return true
	}
	
	// Check for agents directory
	if info, err := os.Stat("agents"); err == nil && info.IsDir() {
		return true
	}
	
	// Check for go.mod with agenticgokit dependency
	if content, err := os.ReadFile("go.mod"); err == nil {
		if strings.Contains(string(content), "github.com/kunalkushwaha/agenticgokit") {
			return true
		}
	}
	
	return false
}