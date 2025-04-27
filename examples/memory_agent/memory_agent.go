package main

import (
	"context"
	"encoding/json" // Import encoding/json for trace dumping
	"fmt"
	"io/ioutil" // Import io/ioutil for writing trace to file
	"log"
	"os"        // Import os for signal handling
	"os/signal" // Import os/signal
	"runtime"   // Import runtime
	"sync"
	"syscall" // Import syscall for signal handling
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/orchestrator"
)

// MemoryAgent demonstrates a simple agent that stores received data in memory.
type MemoryAgent struct {
	id         string
	mu         sync.RWMutex
	memory     map[string]agentflow.EventData // Store data per event ID
	cbRegistry *agentflow.CallbackRegistry    // For potential future event emission
	wg         *sync.WaitGroup                // Optional: For tracking completion
}

// NewMemoryAgent creates a new MemoryAgent.
func NewMemoryAgent(id string, registry *agentflow.CallbackRegistry, wg *sync.WaitGroup) *MemoryAgent {
	return &MemoryAgent{
		id:         id,
		memory:     make(map[string]agentflow.EventData),
		cbRegistry: registry, // Store registry
		wg:         wg,
	}
}

// Run implements the agentflow.AgentHandler interface.
func (a *MemoryAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	if a.wg != nil {
		defer a.wg.Done() // Signal completion if WaitGroup is used
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	eventID := event.GetID()
	// FIX: Use GetData() instead of GetPayload()
	eventData := event.GetData()

	log.Printf("MemoryAgent [%s]: Received event %s with data: %v", a.id, eventID, eventData)

	// Store the data associated with the event ID
	a.memory[eventID] = eventData

	// Example: Optionally trigger another event via callback
	// if someCondition {
	//     nextEvent := agentflow.NewEvent(...)
	//     // FIX: Use correct CallbackArgs fields
	//     args := agentflow.CallbackArgs{ Hook: agentflow.HookAfterAgentRun, Event: nextEvent, Ctx: ctx, State: state }
	//     a.cbRegistry.Invoke(args)
	// }

	// Return an empty result, indicating success without state modification
	return agentflow.AgentResult{}, nil
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
	runner := agentflow.NewRunner(runtime.NumCPU()) // Use multiple workers

	// Create and set TraceLogger
	traceLogger := agentflow.NewInMemoryTraceLogger()
	// FIX: SetTraceLogger does not return an error to check here
	runner.SetTraceLogger(traceLogger)
	log.Println("InMemoryTraceLogger created and set.")

	// Set the Orchestrator for the Runner
	// FIX: SetOrchestrator does not return an error to check here
	runner.SetOrchestrator(orchestratorImpl)
	log.Println("Runner, Logger & RouteOrchestrator created and linked.")

	// --- Register Callbacks ---
	// Create the trace logging callback
	traceCallback := agentflow.CreateTraceCallback(traceLogger)

	// Register the trace callback for relevant hooks
	runner.RegisterCallback(agentflow.HookBeforeEventHandling, agentflow.TraceCallbackName, traceCallback)
	runner.RegisterCallback(agentflow.HookAfterEventHandling, agentflow.TraceCallbackName, traceCallback)
	runner.RegisterCallback(agentflow.HookBeforeAgentRun, agentflow.TraceCallbackName, traceCallback)
	runner.RegisterCallback(agentflow.HookAfterAgentRun, agentflow.TraceCallbackName, traceCallback)
	log.Println("Trace callback registered for hooks.")
	// --------------------------

	// 2. Setup Agent & Register
	const memoryAgentName = "memory-agent-1"
	var wg sync.WaitGroup // WaitGroup for handler completion

	memoryAgent := NewMemoryAgent(memoryAgentName, callbackRegistry, &wg)

	// Register the agent with the orchestrator
	err := orchestratorImpl.RegisterAgent(memoryAgentName, memoryAgent)
	if err != nil {
		log.Fatalf("Failed to register agent '%s': %v", memoryAgentName, err)
	}
	log.Printf("Agent '%s' registered.", memoryAgentName)

	// 3. Start the Runner's processing loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is cancelled on exit

	wg.Add(1) // Add for the runner goroutine itself
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

	eventIDs := []string{} // Store IDs for later retrieval

	for i := 1; i <= numEvents; i++ {
		eventID := fmt.Sprintf("event-%d-for-%s", i, sessionID)
		eventIDs = append(eventIDs, eventID) // Store the ID

		eventData := agentflow.EventData{
			"message": fmt.Sprintf("Data for event %d", i),
			"index":   i,
		}
		// Route event to the memory agent using metadata
		eventMetadata := map[string]string{
			orchestrator.RouteMetadataKey: memoryAgentName,
			agentflow.SessionIDKey:        sessionID, // Include session ID for tracing
		}
		event := agentflow.NewEvent(
			"", // No target ID needed when using RouteMetadataKey
			eventData,
			eventMetadata,
		)
		event.SetID(eventID) // Override generated ID for predictability

		log.Printf("Dispatching event %s...", event.GetID())
		wg.Add(1) // Increment WaitGroup *before* dispatching for the agent's Run method
		dispatchErr := orchestratorImpl.Dispatch(event)
		if dispatchErr != nil {
			log.Printf("Error dispatching event %s: %v", event.GetID(), dispatchErr)
			wg.Done() // Decrement if dispatch fails immediately
		}
		time.Sleep(50 * time.Millisecond) // Small delay for clearer logs
	}

	// 5. Wait for graceful shutdown signal
	log.Println("Application running. Press Ctrl+C to exit.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal or context cancellation (e.g., if runner fails)
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Shutting down...", sig)
	case <-ctx.Done():
		log.Printf("Context cancelled. Shutting down...")
	}

	// 6. Initiate shutdown
	cancel() // Signal runner and other components to stop

	// Wait for handlers and runner to finish
	log.Println("Waiting for active handlers and runner to complete...")
	waitTimeout := 10 * time.Second // Timeout for waiting
	done := make(chan struct{})
	go func() {
		wg.Wait() // Wait for runner and agent Run calls
		close(done)
	}()

	select {
	case <-done:
		log.Println("All components shut down gracefully.")
	case <-time.After(waitTimeout):
		log.Println("Warning: Timeout waiting for components to shut down.")
	}

	// 7. Verify stored data (optional)
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
	log.Println("Retrieving trace via runner for session:", sessionID)
	finalTrace, traceErr := runner.DumpTrace(sessionID)
	if traceErr != nil {
		log.Printf("Error retrieving trace via runner: %v", traceErr)
	} else {
		// --- Write Trace to File ---
		traceFilename := fmt.Sprintf("%s.trace.json", sessionID)
		traceJSON, jsonErr := json.MarshalIndent(finalTrace, "", "  ")
		if jsonErr != nil {
			log.Printf("Error marshaling trace to JSON: %v", jsonErr)
		} else {
			writeErr := ioutil.WriteFile(traceFilename, traceJSON, 0644) // Use io/ioutil or os.WriteFile
			if writeErr != nil {
				log.Printf("Error writing trace file '%s': %v", traceFilename, writeErr)
			} else {
				log.Printf("Trace written to %s", traceFilename)
			}
		}
		// --- End Write Trace to File ---

		// Still print to console as well
		fmt.Println("\n--- Trace Dump (Console) ---")
		if jsonErr != nil {
			// Fallback print if JSON failed
			for _, entry := range finalTrace {
				fmt.Printf("%+v\n", entry)
			}
		} else {
			fmt.Println(string(traceJSON))
		}
		fmt.Println("--- End Trace Dump (Console) ---")
	}
	// ------------------

	log.Println("Memory Agent Example finished.")
}
