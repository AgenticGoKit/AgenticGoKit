package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/agenticgokit/agenticgokit/cmd/agentcli/version"
	"github.com/spf13/cobra"
)

var (
	versionOutput string
	versionShort  bool
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display version information for agentcli including build details.

This command shows comprehensive version information including:
  * Version number or development build info
  * Git commit hash and branch
  * Build date and time
  * Go version used for compilation
  * Target platform and compiler

Examples:
  # Show basic version information
  agentcli version

  # Show detailed version information
  agentcli version --output detailed

  # Show version in JSON format
  agentcli version --output json

  # Show short version only
  agentcli version --short`,
	Run: runVersionCommand,
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Add flags for different output formats
	versionCmd.Flags().StringVarP(&versionOutput, "output", "o", "default", 
		"Output format (default, detailed, json)")
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, 
		"Show only the version number")
}

func runVersionCommand(cmd *cobra.Command, args []string) {
	if versionShort {
		info := version.GetVersionInfo()
		fmt.Println(info.Version)
		return
	}

	switch versionOutput {
	case "json":
		info := version.GetVersionInfo()
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(info); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding version info: %v\n", err)
			os.Exit(1)
		}
	case "detailed":
		fmt.Println(version.GetDetailedVersionString())
	case "default":
		fallthrough
	default:
		fmt.Println(version.GetVersionString())
	}
}
