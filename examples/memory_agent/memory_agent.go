package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/orchestrator"
)

// MemoryAgent demonstrates a simple agent that stores received data in memory.
type MemoryAgent struct {
	id         string
	mu         sync.RWMutex
	memory     map[string]agentflow.EventData
	cbRegistry *agentflow.CallbackRegistry
	wg         *sync.WaitGroup
}

// NewMemoryAgent creates a new MemoryAgent.
func NewMemoryAgent(id string, registry *agentflow.CallbackRegistry, wg *sync.WaitGroup) *MemoryAgent {
	return &MemoryAgent{
		id:         id,
		memory:     make(map[string]agentflow.EventData),
		cbRegistry: registry,
		wg:         wg,
	}
}

// Run implements the agentflow.AgentHandler interface.
func (a *MemoryAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	if a.wg != nil {
		defer a.wg.Done()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	eventID := event.GetID()
	eventData := event.GetData()

	log.Printf("MemoryAgent [%s]: Received event %s with data: %v", a.id, eventID, eventData)

	// Store the data associated with the event ID
	a.memory[eventID] = eventData

	// Process the state (in a real agent this could be more complex)
	outputState := state.Clone()
	outputState.Set("processed_by", a.id)
	outputState.Set("processed_at", time.Now().Format(time.RFC3339))

	return agentflow.AgentResult{
		OutputState: outputState,
	}, nil
}

// GetData retrieves stored data for a specific event ID.
func (a *MemoryAgent) GetData(eventID string) (agentflow.EventData, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	data, found := a.memory[eventID]
	return data, found
}

// --- Main Execution ---

func main() {
	log.Println("Starting Memory Agent Example...")

	// 1. Setup Core Components
	callbackRegistry := agentflow.NewCallbackRegistry()
	orchestratorImpl := orchestrator.NewRouteOrchestrator(callbackRegistry)
	runner := agentflow.NewRunner(runtime.NumCPU())

	// Create and set TraceLogger
	traceLogger := agentflow.NewInMemoryTraceLogger()
	runner.SetTraceLogger(traceLogger)
	log.Println("InMemoryTraceLogger created and set.")

	// Set the Orchestrator for the Runner
	runner.SetOrchestrator(orchestratorImpl)
	log.Println("Runner, Logger & RouteOrchestrator created and linked.")

	// --- Register Callbacks ---
	// Register trace callback for all hooks
	callbackRegistry.Register(agentflow.HookAll, "traceLogger",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			// Try to extract session ID from multiple sources
			var sessionID string

			// 1. First try state metadata
			if sid, ok := args.State.GetMeta(agentflow.SessionIDKey); ok && sid != "" {
				sessionID = sid
				log.Printf("Trace callback: Found session ID in state metadata: %s", sessionID)
			} else if args.Event != nil {
				// 2. Fall back to event metadata if available
				if meta := args.Event.GetMetadata(); meta != nil {
					if sid, ok := meta[agentflow.SessionIDKey]; ok && sid != "" {
						sessionID = sid
						log.Printf("Trace callback: Found session ID in event metadata: %s", sessionID)

						// Add to state metadata for future use
						args.State.SetMeta(agentflow.SessionIDKey, sid)
					}
				}
			}

			// If still no session ID, generate a warning
			if sessionID == "" {
				log.Printf("Trace callback: No session ID found in metadata")
				sessionID = "session-" + args.Event.GetID() // Use event ID as fallback
			}

			entry := agentflow.TraceEntry{
				Hook:      args.Hook,
				Timestamp: time.Now(),
				State:     args.State,
				SessionID: sessionID, // This will now be correctly set
			}

			if args.AgentID != "" {
				entry.AgentID = args.AgentID
			}

			if args.Output.OutputState != nil {
				entry.AgentResult = &args.Output
			}

			if args.Output.Error != "" {
				errMsg := args.Output.Error
				entry.Error = errMsg
			}

			log.Printf("Logging trace entry with SessionID: %s, Hook: %s, AgentID: %s",
				sessionID, string(args.Hook), args.AgentID)

			// Log the entry with the corrected session ID
			if err := traceLogger.Log(entry); err != nil {
				log.Printf("Error logging trace entry: %v", err)
			}
			return args.State, nil
		})
	log.Println("Trace callback registered for all hooks.")

	// After registering the callback
	log.Printf("Verifying trace logger and callback registration...")
	testEntry := agentflow.TraceEntry{
		Hook:      agentflow.HookBeforeAgentRun,
		Timestamp: time.Now(),
		SessionID: "test-session",
	}
	if err := traceLogger.Log(testEntry); err != nil {
		log.Printf("Warning: Trace logger test failed: %v", err)
	} else {
		log.Printf("Trace logger test succeeded")
		testTrace, _ := traceLogger.GetTrace("test-session")
		log.Printf("Retrieved %d test entries", len(testTrace))
	}

	// --------------------------

	// 2. Setup Agent & Register
	const memoryAgentName = "memory-agent-1"
	var wg sync.WaitGroup

	memoryAgent := NewMemoryAgent(memoryAgentName, callbackRegistry, &wg)

	// Register the agent with the orchestrator
	err := orchestratorImpl.RegisterAgent(memoryAgentName, memoryAgent)
	if err != nil {
		log.Fatalf("Failed to register agent '%s': %v", memoryAgentName, err)
	}
	log.Printf("Agent '%s' registered.", memoryAgentName)

	// 3. Start the Runner's processing loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Runner starting...")
		if startErr := runner.Start(ctx); startErr != nil && startErr != context.Canceled {
			log.Printf("Runner failed: %v", startErr)
		}
		log.Println("Runner stopped.")
	}()

	// Allow runner to initialize
	time.Sleep(100 * time.Millisecond)

	// 4. Emit Events
	const numEvents = 2
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	log.Printf("Using Session ID for this run: %s", sessionID)

	eventIDs := []string{}

	for i := 1; i <= numEvents; i++ {
		eventID := fmt.Sprintf("event-%d-for-%s", i, sessionID)
		eventIDs = append(eventIDs, eventID)

		eventData := agentflow.EventData{
			"message": fmt.Sprintf("Data for event %d", i),
			"index":   i,
		}

		eventMetadata := map[string]string{
			orchestrator.RouteMetadataKey: memoryAgentName,
			agentflow.SessionIDKey:        sessionID,
		}

		event := agentflow.NewEvent(
			"",
			eventData,
			eventMetadata,
		)
		event.SetID(eventID)

		log.Printf("Event %s has session ID %s in metadata",
			eventID, eventMetadata[agentflow.SessionIDKey])
		log.Printf("Dispatching event %s...", event.GetID())
		wg.Add(1)
		result, dispatchErr := orchestratorImpl.Dispatch(ctx, event)
		if dispatchErr != nil {
			log.Printf("Error dispatching event %s: %v", event.GetID(), dispatchErr)
			wg.Done()
		}

		_ = result

		time.Sleep(50 * time.Millisecond)
	}

	// 5. Wait for graceful shutdown signal
	log.Println("Application running. Press Ctrl+C to exit.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Shutting down...", sig)
	case <-ctx.Done():
		log.Printf("Context cancelled. Shutting down...")
	}

	// 6. Initiate shutdown
	cancel()

	// Wait for handlers and runner to finish
	log.Println("Waiting for active handlers and runner to complete...")
	waitTimeout := 10 * time.Second
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All components shut down gracefully.")
	case <-time.After(waitTimeout):
		log.Println("Warning: Timeout waiting for components to shut down.")
	}

	// 7. Verify stored data
	fmt.Println("\n--- Stored Data Verification ---")
	for _, eventID := range eventIDs {
		data, found := memoryAgent.GetData(eventID)
		if found {
			log.Printf("Retrieved data for %s: %v", eventID, data)
		} else {
			log.Printf("Data NOT found for %s", eventID)
		}
	}
	fmt.Println("------------------------------")

	// --- Dump Trace ---
	log.Println("Retrieving trace via tracer for session:", sessionID)
	log.Printf("About to retrieve trace for session ID: %s", sessionID)
	finalTrace, traceErr := traceLogger.GetTrace(sessionID)
	log.Printf("Retrieved %d trace entries for session ID: %s",
		len(finalTrace), sessionID)
	if traceErr != nil {
		log.Printf("Error retrieving trace: %v", traceErr)
	} else {
		// --- Write Trace to File ---
		traceFilename := fmt.Sprintf("%s.trace.json", sessionID)
		traceJSON, jsonErr := json.MarshalIndent(finalTrace, "", "  ")
		if jsonErr != nil {
			log.Printf("Error marshaling trace to JSON: %v", jsonErr)
		} else {
			writeErr := os.WriteFile(traceFilename, traceJSON, 0644)
			if writeErr != nil {
				log.Printf("Error writing trace file '%s': %v", traceFilename, writeErr)
			} else {
				log.Printf("Trace written to %s", traceFilename)
			}
		}

		// Print to console
		fmt.Println("\n--- Trace Dump (Console) ---")
		if jsonErr != nil {
			for _, entry := range finalTrace {
				fmt.Printf("%+v\n", entry)
			}
		} else {
			fmt.Println(string(traceJSON))
		}
		fmt.Println("--- End Trace Dump (Console) ---")
	}

	log.Println("Memory Agent Example finished.")
}
