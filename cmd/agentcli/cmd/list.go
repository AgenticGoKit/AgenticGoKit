/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available sessions (based on *.trace.json files)",
	Long: `Scans the current directory for files matching the pattern "*.trace.json"
and lists the corresponding session IDs.`,
	Run: func(cmd *cobra.Command, args []string) {
		listSessions()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	// No flags needed for this simple version yet.
}

func listSessions() {
	files, err := os.ReadDir(".") // Read current directory
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	var sessionIDs []string
	const traceSuffix = ".trace.json"

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), traceSuffix) {
			sessionID := strings.TrimSuffix(file.Name(), traceSuffix)
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	if len(sessionIDs) == 0 {
		fmt.Println("No session trace files (*.trace.json) found in the current directory.")
		return
	}

	fmt.Println("Available Sessions (from trace files):")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "SESSION ID")
	fmt.Fprintln(w, "----------")
	for _, id := range sessionIDs {
		fmt.Fprintln(w, id)
	}
	w.Flush()
}
