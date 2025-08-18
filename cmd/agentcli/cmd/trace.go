/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort" // Add sort import
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	agenticgokit "github.com/kunalkushwaha/agenticgokit/core" // Import core types

	"github.com/spf13/cobra"
)

// JSON structs for trace deserialization - moved to package level for reuse
// This follows Azure best practices for modular code organization
type JSONState struct {
	Data map[string]interface{} `json:"data,omitempty"` // Add omitempty
	Meta map[string]string      `json:"meta,omitempty"` // Add omitempty
}

type JSONAgentResult struct {
	OutputState *JSONState `json:"output_state,omitempty"` // ← matches trace JSON
	Error       string     `json:"error,omitempty"`
}

type JSONTraceEntry struct {
	Timestamp     time.Time        `json:"timestamp"`
	Type          string           `json:"type"`
	EventID       string           `json:"event_id"`
	SessionID     string           `json:"session_id"`
	AgentID       string           `json:"agent_id"`
	State         *JSONState       `json:"state"`
	AgentResult   *JSONAgentResult `json:"agent_result,omitempty"`
	Hook          string           `json:"hook"`
	Error         string           `json:"error,omitempty"`
	TargetAgentID string           `json:"target_agent_id,omitempty"`
	SourceAgentID string           `json:"source_agent_id,omitempty"`
}

// AgentFlowEntry represents a single hop recorded in the flow-only view.
type AgentFlowEntry struct {
	Timestamp time.Time
	Agent     string
	NextAgent string
	EventID   string
	Hook      string
	Kind      string // "", "requeue", "retry"
}

var filterAgent string // Variable to hold the filter flag value
var flowOnlyFlag bool  // Flag to show only agent flow without state details
var verboseFlag bool   // Flag to show full state details
var debugFlag bool     // Flag to show raw JSON structure for debugging

// traceCmd represents the trace command
var traceCmd = &cobra.Command{
	Use:   "trace <sessionID>",
	Short: "Display the execution trace for a specific session",
	Long: `Reads the trace data for the given session ID (expected as a JSON file 
named <sessionID>.trace.json in the current directory) and displays it.`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument: sessionID
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		if flowOnlyFlag {
			displayAgentFlow(sessionID, filterAgent)
		} else {
			displayTrace(sessionID, filterAgent)
		}
	},
}

func init() {
	rootCmd.AddCommand(traceCmd)

	// Add --filter flag (only supporting agent=<name> for now)
	traceCmd.Flags().StringVar(&filterAgent, "filter", "", "Filter trace entries (e.g., agent=<name>)")

	// Add --flow-only flag to show only the flow between agents
	traceCmd.Flags().BoolVar(&flowOnlyFlag, "flow-only", false, "Show only the flow of requests between agents")

	// Add --verbose flag to show full state details
	traceCmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Show full state details (warning: can produce very large output)")

	// Add debug flag for showing raw JSON structure
	traceCmd.Flags().BoolVar(&debugFlag, "debug", false, "Show raw JSON structure for debugging")
}

// Update the displayTrace function with better table formatting

func displayTrace(sessionID, filter string) {
	// Extract actual session ID from path if provided
	actualSessionID := filepath.Base(sessionID)
	if strings.HasSuffix(actualSessionID, ".trace.json") {
		actualSessionID = strings.TrimSuffix(actualSessionID, ".trace.json")
	}

	// Construct filename - check if sessionID is already a path
	var filePath string
	if strings.Contains(sessionID, string(filepath.Separator)) || strings.Contains(sessionID, "/") {
		// sessionID contains path, use it directly
		if strings.HasSuffix(sessionID, ".trace.json") {
			filePath = sessionID
		} else {
			filePath = sessionID + ".trace.json"
		}
	} else {
		// sessionID is just the ID, construct path
		filename := fmt.Sprintf("%s.trace.json", sessionID)
		filePath = filepath.Join(".", filename)
	}

	// Read the trace file
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading trace file: %v\n", err)
		os.Exit(1)
	}

	// Unmarshal JSON data into intermediate struct
	var jsonEntries []JSONTraceEntry
	err = json.Unmarshal(data, &jsonEntries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing trace data: %v\n", err)
		os.Exit(1)
	}

	// Convert to agenticgokit.TraceEntry
	var traceEntries []agenticgokit.TraceEntry
	for _, je := range jsonEntries {
		entry := agenticgokit.TraceEntry{
			Timestamp: je.Timestamp,
			EventID:   je.EventID,
			SessionID: je.SessionID,
			AgentID:   je.AgentID,
			Hook:      agenticgokit.HookPoint(je.Hook),
			Error:     je.Error,
		}

		// Extract data from direct state (rare in this trace format but handle it)
		if je.State != nil && je.State.Data != nil {
			state := agenticgokit.NewState()
			for k, v := range je.State.Data {
				if v != nil {
					state.Set(k, v)
				}
			}
			if je.State.Meta != nil {
				for k, v := range je.State.Meta {
					if v != "" {
						state.SetMeta(k, v)
					}
				}
			}
			entry.State = state
		}

		// IMPORTANT: Also extract from agent_result.output_state
		// This is where most of the real state data lives!
		if je.AgentResult != nil && je.AgentResult.OutputState != nil {
			// Create new state if needed, or use existing
			state := entry.State
			if state == nil {
				state = agenticgokit.NewState()
			}

			// Add data from output_state
			if je.AgentResult.OutputState.Data != nil {
				for k, v := range je.AgentResult.OutputState.Data {
					if v != nil {
						state.Set(k, v)
					}
				}
			}

			// Add meta from output_state
			if je.AgentResult.OutputState.Meta != nil {
				for k, v := range je.AgentResult.OutputState.Meta {
					if v != "" {
						state.SetMeta(k, v)
					}
				}
			}

			entry.State = state
		}

		// Convert AgentResult if present
		if je.AgentResult != nil {
			agentResult := &agenticgokit.AgentResult{
				Error: je.AgentResult.Error,
			}

			if je.AgentResult.OutputState != nil {
				outputState := agenticgokit.NewState()
				for k, v := range je.AgentResult.OutputState.Data {
					outputState.Set(k, v)
				}
				for k, v := range je.AgentResult.OutputState.Meta {
					outputState.SetMeta(k, v)
				}
				agentResult.OutputState = outputState
			}

			entry.AgentResult = agentResult
		}

		traceEntries = append(traceEntries, entry)
	}

	// Apply filters
	var filteredEntries []agenticgokit.TraceEntry
	filterAgentName := ""
	if strings.HasPrefix(filter, "agent=") {
		filterAgentName = strings.TrimPrefix(filter, "agent=")
	}

	if filterAgentName != "" {
		for _, entry := range traceEntries {
			// Update: Use AgentID instead of AgentName
			if entry.AgentID == filterAgentName {
				filteredEntries = append(filteredEntries, entry)
			}
		}
	} else {
		filteredEntries = traceEntries // No agent filter applied
	}

	if len(filteredEntries) == 0 {
		var agentFilterMsg string
		if filterAgentName != "" {
			agentFilterMsg = " for agent '" + filterAgentName + "'"
		}
		fmt.Println("No trace entries found" + agentFilterMsg + " in session " + actualSessionID)
		return
	}
	// Print trace table with improved formatting
	fmt.Printf("Trace for session %s:\n\n", actualSessionID)

	// Use box-drawing characters for table borders
	if !verboseFlag {
		printFormattedTable(filteredEntries)
	} else {
		printVerboseTable(filteredEntries)
	}

	// Add debug output in displayTrace if needed
	if debugFlag && len(jsonEntries) > 0 {
		fmt.Println("\nSample entry JSON structure:")
		sampleBytes, _ := json.MarshalIndent(jsonEntries[0], "", "  ")
		fmt.Println(string(sampleBytes))
	}

	// Add this after converting JSONTraceEntry to TraceEntry in displayTrace

	// Debug state extraction
	if debugFlag {
		fmt.Println("\nState Data Debug:")
		for i, je := range jsonEntries[:3] { // Just the first 3 for brevity
			fmt.Printf("Entry %d (%s - %s):\n", i, je.Hook, je.AgentID)

			// Check direct state
			if je.State != nil && je.State.Data != nil {
				fmt.Printf("  Direct state keys: %v\n", maps.Keys(je.State.Data))
			} else {
				fmt.Println("  No direct state data")
			}

			// Check agent result state
			if je.AgentResult != nil && je.AgentResult.OutputState != nil {
				fmt.Printf("  Result state keys: %v\n", maps.Keys(je.AgentResult.OutputState.Data))
			} else {
				fmt.Println("  No result state data")
			}
		}
	}
}

// Add these new helper functions for table display

// printFormattedTable prints a visually appealing boxed table with proper truncation
func printFormattedTable(entries []agenticgokit.TraceEntry) {
	// Define column widths for balanced display
	const (
		timeWidth  = 19 // Fixed width for timestamp (15:04:05.000)
		hookWidth  = 25
		agentWidth = 16
		stateWidth = 40
		errorWidth = 30
	)

	// Print table header with box-drawing characters
	fmt.Println("+" + strings.Repeat("-", timeWidth) + "+" +
		strings.Repeat("-", hookWidth) + "+" +
		strings.Repeat("-", agentWidth) + "+" +
		strings.Repeat("-", stateWidth) + "+" +
		strings.Repeat("-", errorWidth) + "+")

	fmt.Printf("| %-*s | %-*s | %-*s | %-*s | %-*s |\n",
		timeWidth-2, "TIMESTAMP",
		hookWidth-2, "HOOK",
		agentWidth-2, "AGENT",
		stateWidth-2, "STATE",
		errorWidth-2, "ERROR")

	fmt.Println("+" + strings.Repeat("-", timeWidth) + "+" +
		strings.Repeat("-", hookWidth) + "+" +
		strings.Repeat("-", agentWidth) + "+" +
		strings.Repeat("-", stateWidth) + "+" +
		strings.Repeat("-", errorWidth) + "+")

	// Print each row
	for _, entry := range entries {
		timeStr := entry.Timestamp.Format("15:04:05.000")
		hook := truncateString(string(entry.Hook), hookWidth-2)
		agent := truncateString(entry.AgentID, agentWidth-2)

		// Format state summary
		stateSummaryText := truncateString(stateSummary(entry.State), stateWidth-2)

		// Format error message
		var errorText string
		if entry.Error != "" {
			errorText = truncateString(entry.Error, errorWidth-2)
		} else if entry.AgentResult != nil && entry.AgentResult.Error != "" {
			errorText = truncateString(entry.AgentResult.Error, errorWidth-2)
		} else {
			errorText = "-"
		}

		fmt.Printf("| %-*s | %-*s | %-*s | %-*s | %-*s |\n",
			timeWidth-2, timeStr,
			hookWidth-2, hook,
			agentWidth-2, agent,
			stateWidth-2, stateSummaryText,
			errorWidth-2, errorText)
	}

	fmt.Println("+" + strings.Repeat("-", timeWidth) + "+" +
		strings.Repeat("-", hookWidth) + "+" +
		strings.Repeat("-", agentWidth) + "+" +
		strings.Repeat("-", stateWidth) + "+" +
		strings.Repeat("-", errorWidth) + "+")
}

// truncateString shortens a string to fit within width and adds ellipsis if needed
func truncateString(s string, width int) string {
	if len(s) <= width {
		return s
	}

	if width <= 3 {
		return "..."[:width]
	}

	return s[:width-3] + "..."
}

// printVerboseTable displays full details with potential line wrapping
func printVerboseTable(entries []agenticgokit.TraceEntry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tHOOK\tAGENT\tSTATE\tERROR")
	fmt.Fprintln(w, "---------\t----\t-----\t-----\t-----")

	for _, entry := range entries {
		ts := entry.Timestamp.Format("15:04:05.000")
		hook := string(entry.Hook)
		agent := entry.AgentID

		// Full state data
		stateText := verboseStateSummary(entry.State)

		// Full error
		errorText := "-"
		if entry.Error != "" {
			errorText = entry.Error
		} else if entry.AgentResult != nil && entry.AgentResult.Error != "" {
			errorText = entry.AgentResult.Error
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			ts, hook, agent, stateText, errorText)
	}
	w.Flush()
}

// verboseStateSummary provides a complete state dump for verbose mode
func verboseStateSummary(s agenticgokit.State) string {
	if s == nil {
		return "-"
	}

	keys := s.Keys()
	if len(keys) == 0 {
		return "{}"
	}

	sort.Strings(keys)
	var parts []string

	for _, key := range keys {
		if val, ok := s.Get(key); ok {
			parts = append(parts, fmt.Sprintf("%s:%v", key, val))
		}
	}

	return "{" + strings.Join(parts, ", ") + "}"
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
func stateSummary(s agenticgokit.State) string {
	if s == nil {
		return "-" // Return dash for nil state
	}

	// Use the Keys() method defined in the State interface
	keys := s.Keys()
	if len(keys) == 0 {
		return "{}" // Return empty braces for empty state
	}

	// Sort keys for consistent output
	sort.Strings(keys)

	// Show first few keys for brevity
	maxKeys := 3
	if len(keys) > maxKeys {
		keys = keys[:maxKeys]
	}

	var parts []string
	for _, key := range keys {
		if val, ok := s.Get(key); ok {
			// Format key-value pairs with truncation for large values
			parts = append(parts, fmt.Sprintf("%s:%s", key, formatStateValue(val)))
		}
	}

	summary := strings.Join(parts, ", ")
	if len(s.Keys()) > maxKeys {
		summary += ", ..."
	}

	return "{" + summary + "}"
}

// formatStateValue formats a value for display in the CLI with length constraints
func formatStateValue(val interface{}) string {
	const maxLen = 30 // Maximum length for displayed strings

	if val == nil {
		return "null"
	}

	switch v := val.(type) {
	case string:
		if len(v) > maxLen {
			return fmt.Sprintf("\"%s\"... (%d chars)", v[:maxLen], len(v))
		}
		return fmt.Sprintf("\"%s\"", v)
	case map[string]interface{}:
		return fmt.Sprintf("{...} (%d keys)", len(v))
	case []interface{}:
		return fmt.Sprintf("[...] (%d items)", len(v))
	default:
		// For other types, use standard formatting but limit length
		s := fmt.Sprintf("%v", v)
		if len(s) > maxLen {
			return fmt.Sprintf("%s... (%d chars)", s[:maxLen], len(s))
		}
		return s
	}
}

// Enhance displayAgentFlow to better show next routes by analyzing agent state metadata
func displayAgentFlow(sessionID, filter string) {
	// Extract actual session ID from path if provided
	actualSessionID := filepath.Base(sessionID)
	if strings.HasSuffix(actualSessionID, ".trace.json") {
		actualSessionID = strings.TrimSuffix(actualSessionID, ".trace.json")
	}

	// Construct filename - check if sessionID is already a path
	var filePath string
	if strings.Contains(sessionID, string(filepath.Separator)) || strings.Contains(sessionID, "/") {
		// sessionID contains path, use it directly
		if strings.HasSuffix(sessionID, ".trace.json") {
			filePath = sessionID
		} else {
			filePath = sessionID + ".trace.json"
		}
	} else {
		// sessionID is just the ID, construct path
		filename := fmt.Sprintf("%s.trace.json", sessionID)
		filePath = filepath.Join(".", filename)
	}

	// collect hop entries here
	var flowEntries []AgentFlowEntry

	// Read and parse the trace file
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading trace file: %v\n", err)
		os.Exit(1)
	}

	// Unmarshal JSON data - now JSONTraceEntry is available at package level
	var jsonEntries []JSONTraceEntry
	err = json.Unmarshal(data, &jsonEntries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing trace data: %v\n", err)
		os.Exit(1)
	}

	// Using the same intermediate struct from displayTrace
	// ... (same JSON struct definitions) ...
	// Process entries to extract flow information
	// Update extracting next route from trace entries
	for _, je := range jsonEntries {
		// Only focus on AfterEventHandling and AfterAgentRun hooks as they contain the routing decision
		if je.Hook != "AfterEventHandling" && je.Hook != "AfterAgentRun" {
			continue
		}

		// Extract route metadata from trace entry's state
		nextAgent := "-"

		// 2a. try state / agent-result meta (unchanged)
		if je.State != nil && je.State.Meta != nil {
			if route, ok := je.State.Meta["route"]; ok && route != "" {
				nextAgent = route
			}
		}
		if nextAgent == "-" && je.AgentResult != nil && je.AgentResult.OutputState != nil &&
			je.AgentResult.OutputState.Meta != nil {
			if route, ok := je.AgentResult.OutputState.Meta["route"]; ok && route != "" {
				nextAgent = route
			}
		}

		// 2b. if still unknown, fall back to the explicit JSON field
		if nextAgent == "-" && je.TargetAgentID != "" {
			nextAgent = je.TargetAgentID
		}

		kind := ""
		// Check for two types of requeues:
		// 1. Direct self-dispatch in this hop
		if nextAgent != "-" && nextAgent == je.AgentID {
			if je.AgentResult != nil && je.AgentResult.Error != "" {
				kind = "retry"
			} else {
				kind = "requeue"
			}
			// 2. Event was created by the same agent - indicates a requeue
		} else if je.SourceAgentID != "" && je.SourceAgentID == je.AgentID {
			if je.AgentResult != nil && je.AgentResult.Error != "" {
				kind = "retry"
			} else {
				kind = "requeue"
			}
		}

		flowEntries = append(flowEntries, AgentFlowEntry{
			Timestamp: je.Timestamp,
			Agent:     je.AgentID,
			NextAgent: nextAgent,
			EventID:   je.EventID,
			Hook:      je.Hook,
			Kind:      kind,
		})
	}
	if len(flowEntries) == 0 {
		fmt.Println("No agent flow data found for session " + actualSessionID)
		return
	}

	// Sort by timestamp to ensure chronological order
	sort.Slice(flowEntries, func(i, j int) bool {
		return flowEntries[i].Timestamp.Before(flowEntries[j].Timestamp)
	})

	// Print the flow
	fmt.Printf("Agent request flow for session %s:\n\n", actualSessionID)

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 1, ' ', 0)
	fmt.Fprintf(w, "TIME\tAGENT\tNEXT\tHOOK\tEVENT ID\n") // added HOOK column

	for _, entry := range flowEntries {
		timeStr := entry.Timestamp.Format("15:04:05.000")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			timeStr,
			entry.Agent,
			entry.NextAgent, // unchanged
			entry.Hook,      // new column value
			shortenID(entry.EventID))
	}
	w.Flush()

	// Print a basic sequence diagram
	fmt.Println("\nSequence diagram:")
	fmt.Println("----------------")

	// Keep track of unique agents to create columns
	agents := make(map[string]bool)
	for _, e := range flowEntries {
		agents[e.Agent] = true
		if e.NextAgent != "-" {
			agents[e.NextAgent] = true
		}
	}

	// Create a list of unique agents
	var agentList []string
	for a := range agents {
		agentList = append(agentList, a)
	}
	sort.Strings(agentList)

	// Print the sequence
	for i, entry := range flowEntries {
		fmt.Printf("%d. %s -> ", i+1, entry.Agent)
		if entry.NextAgent != "-" {
			fmt.Printf("%s\n", entry.NextAgent)
		} else {
			fmt.Println("(end)")
		}
	}

	// ---------------------------------------------------------------
	// Condensed route (single linear chain, duplicates removed)
	// ---------------------------------------------------------------
	condensed := buildRouteChain(flowEntries)
	if len(condensed) > 1 {
		fmt.Println("\nCondensed route:")
		fmt.Println(strings.Join(condensed, " -> "))
	}

	// ---------------------------------------------------------------
	// Per-event sequence diagrams
	// ---------------------------------------------------------------
	// 1. bucket entries by EventID
	eventBuckets := make(map[string][]AgentFlowEntry)
	firstSeen := make(map[string]time.Time)
	agentNames := make(map[string]struct{})

	for _, e := range flowEntries {
		base := normalizeEventID(e.EventID)

		eventBuckets[base] = append(eventBuckets[base], e)
		if _, ok := firstSeen[base]; !ok {
			firstSeen[base] = e.Timestamp
		}
		agentNames[e.Agent] = struct{}{}
	}

	// 2. order event IDs chronologically
	type idTime struct {
		id string
		t  time.Time
	}
	var ordered []idTime
	for id, t := range firstSeen {
		ordered = append(ordered, idTime{id, t})
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].t.Before(ordered[j].t) })

	fmt.Println("\nPer-event sequence diagrams:")
	fmt.Println("-----------------------------")

	seenLabel := make(map[string]int) // to avoid duplicate "[req-1746...]" labels

	for _, it := range ordered {
		if _, isAgent := agentNames[it.id]; isAgent {
			continue // skip synthetic ids that equal agent names
		}

		entries := eventBuckets[it.id] // show even 1-hop events

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.Before(entries[j].Timestamp)
		})

		// create a unique, compact label
		label := compactID(it.id)
		if n := seenLabel[label]; n > 0 {
			label = fmt.Sprintf("%s~%d", label, n) // req-1746...~1, req-1746...~2, ...
		}
		seenLabel[label]++

		fmt.Printf("\n[%s]\n", label)
		for _, e := range entries {
			var tail string
			if e.NextAgent == "-" {
				tail = "(end)"
			} else {
				tail = e.NextAgent // show even self-dispatch explicitly
			}
			line := fmt.Sprintf("%s -> %s", e.Agent, tail)
			if e.Kind != "" {
				line += "  [" + e.Kind + "]"
			}
			fmt.Println(line)
		}
	}
}

// Add this helper function to shorten event IDs for display
func shortenID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8] + "..."
}

// compactID returns the first 8 characters followed by an ellipsis.
// Examples:
//
//	req-17461593abcd  -> req-1746...
//	1f78e6e6          -> 1f78e6e6   (already ≤8)
func compactID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8] + "..."
}

// buildRouteChain derives a linear chain without duplicates
func buildRouteChain(entries []AgentFlowEntry) []string {
	if len(entries) == 0 {
		return nil
	}
	chain := []string{entries[0].Agent}
	seen := map[string]bool{entries[0].Agent: true}
	cur := entries[0].Agent

	for _, e := range entries {
		if e.Agent == cur && e.NextAgent != "-" && !seen[e.NextAgent] {
			chain = append(chain, e.NextAgent)
			seen[e.NextAgent] = true
			cur = e.NextAgent
		}
	}
	return chain
}

// normalizeEventID collapses variants like
//
//	req-17461596-planner / req-17461596-summarizer -> req-17461596
//	1f78e6e6-beb2...      / 1f78e6e6-c2fc...          -> 1f78e6e6
//
// Other ids are returned unchanged.
func normalizeEventID(id string) string {
	// Case 1: user-request ids
	if strings.HasPrefix(id, "req-") {
		i := 4
		for i < len(id) && id[i] >= '0' && id[i] <= '9' {
			i++
		}
		return id[:i] // e.g. req-17461596
	}

	// Case 2: GUID-like ids where the first 8 chars are hex and the 9th is '-'
	if len(id) > 8 && id[8] == '-' {
		if _, err := strconv.ParseUint(id[:8], 16, 32); err == nil {
			return id[:8] // e.g. 1f78e6e6
		}
	}
	return id
}
