/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context" // <<< Add context import
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	// Remove the memory_agent import
	// memagent "kunalkushwaha/agentflow/examples/memory_agent"

	"github.com/google/uuid" // For generating event IDs
	"github.com/spf13/cobra"
)

var initialData string // Flag to hold initial data (as JSON string)

// --- Copied from examples/memory_agent ---

const CounterAgentName = "counter" // Define locally

// CounterAgent increments a counter stored in session state.
type CounterAgent struct {
	sessionStore agentflow.SessionStore
}

// NewCounterAgent creates a new CounterAgent.
func NewCounterAgent(store agentflow.SessionStore) *CounterAgent {
	if store == nil {
		panic("CounterAgent requires a non-nil session store")
	}
	return &CounterAgent{sessionStore: store}
}

// Run implements the agentflow.Agent interface.
func (a *CounterAgent) Run(ctx context.Context, in agentflow.State) (agentflow.State, error) {
	// Simplified Run logic for CLI context (can refine later)
	log.Printf("[%s] Running...", CounterAgentName)

	sessionID, ok := in.GetMeta(agentflow.SessionIDKey)
	if !ok || sessionID == "" {
		log.Printf("[%s] Error: %s not found in input state metadata", CounterAgentName, agentflow.SessionIDKey)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("%s not found in metadata", agentflow.SessionIDKey))
		return out, nil
	}

	sessionState, found, err := a.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		log.Printf("[%s] Error getting session %s: %v", CounterAgentName, sessionID, err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("session store get error: %v", err))
		return out, nil
	}

	isNew := !found
	if !found {
		sessionState = agentflow.NewState()
	}
	if sessionState == nil { // Should not happen with NewState
		sessionState = agentflow.NewState()
	}

	currentCount := 0
	if countVal, found := sessionState.Get("count"); found {
		// Handle potential different number types from JSON unmarshal
		switch v := countVal.(type) {
		case float64: // JSON numbers often unmarshal as float64
			currentCount = int(v)
		case int:
			currentCount = v
		case int64:
			currentCount = int(v)
		default:
			log.Printf("[%s] Warning: 'count' in session %s is not a recognized number type (%T), resetting.", CounterAgentName, sessionID, countVal)
		}
	}
	currentCount++
	sessionState.Set("count", currentCount)

	err = a.sessionStore.SaveSession(ctx, sessionID, sessionState)
	if err != nil {
		log.Printf("[%s] Error saving session %s: %v", CounterAgentName, sessionID, err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("session store save error: %v", err))
		return out, nil
	}

	out := in.Clone()
	out.Set("current_count", currentCount)
	out.Set("is_new_session", isNew)
	if req, ok := in.Get("user_request"); ok {
		out.Set("user_request", req)
	}

	log.Printf("[%s] Finished run for session %s", CounterAgentName, sessionID)
	return out, nil
}

// SimpleHandler invokes an agent (copied for CLI use).
type SimpleHandler struct {
	agent      agentflow.Agent
	agentName  string
	resultChan chan agentflow.State
	wg         *sync.WaitGroup
}

// NewSimpleHandler creates a new SimpleHandler (copied for CLI use).
func NewSimpleHandler(name string, agent agentflow.Agent, resultChan chan agentflow.State, wg *sync.WaitGroup) *SimpleHandler {
	// Panics removed for brevity, assume valid inputs in this context
	return &SimpleHandler{
		agentName:  name,
		agent:      agent,
		resultChan: resultChan,
		wg:         wg,
	}
}

// Handle invokes the agent and manages callbacks (copied for CLI use).
func (h *SimpleHandler) Handle(event agentflow.Event, registry *agentflow.CallbackRegistry) error {
	defer h.wg.Done()
	log.Printf("[%s Handler] Handling event %s", h.agentName, event.GetID())

	var initialState agentflow.State
	if payloadMap, ok := event.GetPayload().(map[string]any); ok {
		initialState = agentflow.NewStateWithData(payloadMap)
	} else {
		initialState = agentflow.NewState()
	}
	if initialState == nil {
		initialState = agentflow.NewState()
	} // Ensure not nil
	for k, v := range event.GetMetadata() {
		initialState.SetMeta(k, v)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	agentName := h.agentName
	registry.Invoke(agentflow.CallbackArgs{Ctx: ctx, Hook: agentflow.HookBeforeAgentRun, Event: &event, AgentName: &agentName, Input: &initialState})

	finalState, agentErr := h.agent.Run(ctx, initialState)

	registry.Invoke(agentflow.CallbackArgs{Ctx: ctx, Hook: agentflow.HookAfterAgentRun, Event: &event, AgentName: &agentName, Input: &initialState, Output: &finalState, Error: &agentErr})

	if agentErr != nil {
		log.Printf("[%s Handler] Agent run for event %s returned an error: %v", h.agentName, event.GetID(), agentErr)
	}
	if finalState == nil {
		finalState = agentflow.NewState()
	}

	select {
	case h.resultChan <- finalState:
	case <-ctx.Done():
		log.Printf("[%s Handler] Timeout waiting to send result for event %s", h.agentName, event.GetID())
	}
	return nil
}

// --- End Copied ---

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <agent_name>", // Initially, only "counter" will work
	Short: "Run a specific agent (currently supports 'counter')",
	Long: `Runs the specified agent with optional initial data.
Currently, only the 'counter' agent from the memory_agent example is supported.
Example:
  agentcli run counter --data '{"user_request":"increment from cli"}'`,
	Args: cobra.ExactArgs(1), // Requires agent name
	Run: func(cmd *cobra.Command, args []string) {
		agentName := args[0]
		runAgent(agentName, initialData)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Add --data flag for initial input
	runCmd.Flags().StringVarP(&initialData, "data", "d", "{}", "Initial data for the agent run (as a JSON string)")
}

func runAgent(agentName string, dataJSON string) {
	// --- Basic Setup ---
	// For now, hardcode components similar to memory_agent main.go
	sessionStore := agentflow.NewMemorySessionStore()
	traceLogger := agentflow.NewInMemoryTraceLogger()
	registry := agentflow.NewCallbackRegistry()

	// Register the trace callback
	traceCallback := agentflow.CreateTraceCallback(traceLogger)
	registry.Register(agentflow.HookAll, "traceLogger", traceCallback)

	// --- Agent Selection (Hardcoded for now) ---
	var agent agentflow.Agent
	if agentName == CounterAgentName { // <<< Use local constant
		agent = NewCounterAgent(sessionStore) // <<< Use local constructor
		log.Printf("Selected agent: %s", agentName)
	} else {
		fmt.Fprintf(os.Stderr, "Error: Agent '%s' is not supported by this command yet.\n", agentName)
		os.Exit(1)
	}

	// --- Event Creation ---
	sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
	eventID := uuid.New().String()
	metadata := map[string]string{
		agentflow.SessionIDKey: sessionID, // Use the constant from core
		"cli_trigger":          "true",
	}

	// Parse initial data JSON
	var payload map[string]any
	err := json.Unmarshal([]byte(dataJSON), &payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing --data JSON: %v\n", err)
		os.Exit(1)
	}

	event := agentflow.NewEvent(eventID, "cli.run", payload, metadata)
	log.Printf("Created Event ID: %s for Session ID: %s", eventID, sessionID)

	// --- Handler Setup & Execution ---
	var wg sync.WaitGroup
	resultChan := make(chan agentflow.State, 1) // Buffered channel for the result

	handler := NewSimpleHandler(agentName, agent, resultChan, &wg) // <<< Use local constructor

	// Invoke BeforeEventHandling hook
	registry.Invoke(agentflow.CallbackArgs{Hook: agentflow.HookBeforeEventHandling, Event: &event})

	wg.Add(1)
	go handler.Handle(event, registry) // Run handler in a goroutine

	// Wait for the handler to finish and get the result
	finalState := <-resultChan
	wg.Wait() // Ensure Handle goroutine completes fully

	// Invoke AfterEventHandling hook
	registry.Invoke(agentflow.CallbackArgs{Hook: agentflow.HookAfterEventHandling, Event: &event, Output: &finalState})

	// --- Output & Trace ---
	fmt.Println("\n--- Agent Run Complete ---")
	fmt.Printf("Session ID: %s\n", sessionID)
	fmt.Println("Final State:")
	// Basic print of final state (could be improved with formatting)
	if finalState != nil {
		keys := finalState.Keys()
		if len(keys) > 0 {
			fmt.Println("{")
			for _, k := range keys {
				v, _ := finalState.Get(k)
				// Attempt to pretty-print JSON if value is map/slice
				if subMap, ok := v.(map[string]any); ok {
					subJSON, _ := json.MarshalIndent(subMap, "    ", "  ")
					fmt.Printf("  %s: %s\n", k, string(subJSON))
				} else if subSlice, ok := v.([]any); ok {
					subJSON, _ := json.MarshalIndent(subSlice, "    ", "  ")
					fmt.Printf("  %s: %s\n", k, string(subJSON))
				} else {
					fmt.Printf("  %s: %v\n", k, v)
				}
			}
			fmt.Println("}")
		} else {
			fmt.Println("{}")
		}
	} else {
		fmt.Println("<nil>")
	}

	// Save trace to file (similar to memory_agent main.go)
	traceFileName := fmt.Sprintf("%s.trace.json", sessionID)
	traceEntries, _ := traceLogger.GetTrace(sessionID) // Ignore error for simplicity here
	if len(traceEntries) > 0 {
		traceData, err := json.MarshalIndent(traceEntries, "", "  ")
		if err != nil {
			log.Printf("Error marshaling trace data: %v", err)
		} else {
			err = ioutil.WriteFile(traceFileName, traceData, 0644)
			if err != nil {
				log.Printf("Error writing trace file %s: %v", traceFileName, err)
			} else {
				fmt.Printf("\nTrace written to %s\n", traceFileName)
			}
		}
	} else {
		fmt.Println("\nNo trace entries generated for this run.")
	}
}
