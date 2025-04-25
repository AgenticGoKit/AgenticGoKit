/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil" // Use io/ioutil for simplicity, replace with io and os if preferred
	"os"
	"path/filepath"
	"sort" // Add sort import
	"strings"
	"text/tabwriter"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core" // Import core types

	"github.com/spf13/cobra"
)

var filterAgent string // Variable to hold the filter flag value

// traceCmd represents the trace command
var traceCmd = &cobra.Command{
	Use:   "trace <sessionID>",
	Short: "Display the execution trace for a specific session",
	Long: `Reads the trace data for the given session ID (expected as a JSON file 
named <sessionID>.trace.json in the current directory) and displays it.`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument: sessionID
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		displayTrace(sessionID, filterAgent)
	},
}

func init() {
	rootCmd.AddCommand(traceCmd)

	// Add --filter flag (only supporting agent=<name> for now)
	traceCmd.Flags().StringVar(&filterAgent, "filter", "", "Filter trace entries (e.g., agent=<name>)")
}

func displayTrace(sessionID, filter string) {
	// Construct filename
	filename := fmt.Sprintf("%s.trace.json", sessionID)
	filePath := filepath.Join(".", filename) // Look in current directory

	// Read the trace file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading trace file '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	// Unmarshal JSON data
	var traceEntries []agentflow.TraceEntry
	err = json.Unmarshal(data, &traceEntries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing trace file '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	// Apply filters
	var filteredEntries []agentflow.TraceEntry
	filterAgentName := ""
	if strings.HasPrefix(filter, "agent=") {
		filterAgentName = strings.TrimPrefix(filter, "agent=")
	}

	if filterAgentName != "" {
		for _, entry := range traceEntries {
			// Include if AgentName matches or if AgentName is nil (e.g., event hooks)
			if entry.AgentName != nil && *entry.AgentName == filterAgentName {
				filteredEntries = append(filteredEntries, entry)
			} else if entry.AgentName == nil && (entry.Hook == agentflow.HookBeforeEventHandling || entry.Hook == agentflow.HookAfterEventHandling) {
				// Optionally include event-level hooks even when filtering by agent?
				// Let's exclude them for now for a stricter filter.
				// filteredEntries = append(filteredEntries, entry)
			}
		}
	} else {
		filteredEntries = traceEntries // No agent filter applied
	}

	if len(filteredEntries) == 0 {
		fmt.Println("No trace entries found", func() string {
			if filterAgentName != "" {
				return fmt.Sprintf("for agent '%s'.", filterAgentName)
			}
			return "."
		}())

		return
	}

	// Print trace table
	fmt.Printf("Trace for Session: %s\n", sessionID)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tHOOK/ACTION\tAGENT\tINPUT\tOUTPUT\tERROR")
	fmt.Fprintln(w, "---------\t-----------\t-----\t-----\t------\t-----")

	for _, entry := range filteredEntries {
		ts := entry.Timestamp.Format(time.RFC3339)
		hook := string(entry.Hook)
		agent := safeAgentNameCLI(entry.AgentName)

		// Assign Input and Output pointers directly to the interface variables
		var inputState agentflow.State
		if entry.Input != nil { // entry.Input is *SimpleState
			inputState = entry.Input // <<< Assign the pointer directly (*SimpleState implements State)
		}
		var outputState agentflow.State
		if entry.Output != nil { // entry.Output is *SimpleState
			outputState = entry.Output // <<< Assign the pointer directly (*SimpleState implements State)
		}

		input := stateSummary(inputState)   // Pass the interface value (which holds *SimpleState or nil)
		output := stateSummary(outputState) // Pass the interface value (which holds *SimpleState or nil)
		errorMsg := safeErrorMsgCLI(entry.Error)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", ts, hook, agent, input, output, errorMsg)
	}
	w.Flush()
}

// Helper functions for CLI display
func safeAgentNameCLI(name *string) string {
	if name != nil {
		return *name
	}
	return "-" // Use dash for missing agent name
}

func safeErrorMsgCLI(errMsg *string) string {
	if errMsg != nil {
		return *errMsg
	}
	return "-" // Use dash for no error
}

// stateSummary provides a brief summary of state for table view
// Change parameter type from *agentflow.State to agentflow.State
func stateSummary(s agentflow.State) string {
	if s == nil { // Check if the interface itself is nil
		return "-"
	}
	// Use the Keys() method defined in the State interface
	keys := s.Keys() // <<< Use Keys() method from the interface
	if len(keys) == 0 {
		return "{}"
	}

	// Sort keys for consistent output (optional but nice)
	sort.Strings(keys)
	// Show first few keys for brevity
	const maxKeysToShow = 3
	displayKeys := keys
	ellipsis := ""
	if len(keys) > maxKeysToShow {
		displayKeys = keys[:maxKeysToShow]
		ellipsis = ",..."
	}

	return fmt.Sprintf("{%s%s}", strings.Join(displayKeys, ","), ellipsis)
}
