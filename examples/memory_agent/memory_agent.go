package main

import (
	"context"
	"encoding/json" // Import encoding/json for trace dumping
	"fmt"
	"io/ioutil" // Import io/ioutil for writing trace to file
	"log"
	"sync"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/orchestrator"
)

const CounterAgentName = "counter"

// --- Counter Agent ---

type CounterAgent struct {
	sessionStore agentflow.SessionStore
}

func NewCounterAgent(store agentflow.SessionStore) *CounterAgent {
	if store == nil {
		log.Fatal("CounterAgent requires a non-nil SessionStore")
	}
	return &CounterAgent{sessionStore: store}
}

// Run increments a counter stored in the session state.
func (a *CounterAgent) Run(ctx context.Context, in agentflow.State) (agentflow.State, error) {
	log.Printf("[%s] Running...", CounterAgentName)

	// 1. Get Session ID from metadata
	sessionID, ok := in.GetMeta(agentflow.SessionIDKey) // <<< Use exported constant
	if !ok || sessionID == "" {
		log.Printf("[%s] Error: %s not found in input state metadata", CounterAgentName, agentflow.SessionIDKey) // <<< Use exported constant
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("%s not found in metadata", agentflow.SessionIDKey)) // <<< Use exported constant
		return out, nil
	}
	log.Printf("[%s] Found SessionID: %s", CounterAgentName, sessionID)

	// 2. Get or Create Session State
	var sessionState agentflow.State
	var isNew bool
	var err error

	// Try to get the existing session
	sessionState, found, err := a.sessionStore.GetSession(ctx, sessionID) // <<< Use GetSession
	if err != nil {
		// Handle error during retrieval (e.g., database connection issue)
		log.Printf("[%s] Error getting session %s: %v", CounterAgentName, sessionID, err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("session store get error: %v", err))
		return out, nil
	}

	if !found {
		// Session not found, create a new one
		log.Printf("[%s] Session %s not found, creating new state.", CounterAgentName, sessionID)
		sessionState = agentflow.NewState() // Create a new empty state
		isNew = true
	} else {
		// Session found
		log.Printf("[%s] Session %s retrieved.", CounterAgentName, sessionID)
		isNew = false
	}
	// Ensure sessionState is not nil (should be handled by NewState if not found)
	if sessionState == nil {
		log.Printf("[%s] Error: sessionState is nil after Get/Create for session %s", CounterAgentName, sessionID)
		out := in.Clone()
		out.Set("agent_error", "internal session state error")
		return out, nil
	}

	// 3. Increment Counter
	currentCount := 0
	if countVal, found := sessionState.Get("count"); found {
		if countInt, ok := countVal.(int); ok {
			currentCount = countInt
		} else {
			log.Printf("[%s] Warning: 'count' in session %s is not an int, resetting.", CounterAgentName, sessionID)
		}
	}
	currentCount++
	sessionState.Set("count", currentCount)

	// 4. Save Session State
	err = a.sessionStore.SaveSession(ctx, sessionID, sessionState) // <<< Use SaveSession
	if err != nil {
		log.Printf("[%s] Error saving session %s: %v", CounterAgentName, sessionID, err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("session store save error: %v", err))
		return out, nil
	}
	log.Printf("[%s] Session %s saved with count %d", CounterAgentName, sessionID, currentCount)

	// 5. Prepare output state
	out := in.Clone()
	out.Set("current_count", currentCount)
	out.Set("is_new_session", isNew)
	if req, ok := in.Get("user_request"); ok {
		out.Set("user_request", req)
	}

	log.Printf("[%s] Finished run for session %s", CounterAgentName, sessionID)
	return out, nil
}

// --- Agent Handler (Simplified for direct call) ---

type SimpleHandler struct {
	agent      agentflow.Agent
	agentName  string
	resultChan chan agentflow.State // Channel to send results back to main goroutine
	wg         *sync.WaitGroup      // WaitGroup to signal completion
}

func NewSimpleHandler(name string, agent agentflow.Agent, resultChan chan agentflow.State, wg *sync.WaitGroup) *SimpleHandler {
	if agent == nil {
		panic("SimpleHandler requires a non-nil agent")
	}
	if resultChan == nil {
		panic("SimpleHandler requires a non-nil result channel")
	}
	if wg == nil {
		panic("SimpleHandler requires a non-nil WaitGroup")
	}
	return &SimpleHandler{
		agentName:  name,
		agent:      agent,
		resultChan: resultChan,
		wg:         wg,
	}
}

// Handle now accepts CallbackRegistry and invokes agent hooks.
func (h *SimpleHandler) Handle(event agentflow.Event, registry *agentflow.CallbackRegistry) error {
	// Signal completion when this function returns
	defer h.wg.Done()

	log.Printf("[%s Handler] Handling event %s", h.agentName, event.GetID())

	// Prepare initial state from event payload and metadata
	var initialState agentflow.State // <<< Declare as interface type agentflow.State
	if payloadMap, ok := event.GetPayload().(map[string]any); ok {
		initialState = agentflow.NewStateWithData(payloadMap) // Assign interface value
	} else {
		log.Printf("[%s Handler] Warning: Event %s payload is not map[string]any, starting with empty state.", h.agentName, event.GetID())
		initialState = agentflow.NewState() // NewState() returns *SimpleState which satisfies State
	}
	// Ensure initialState is not nil before setting meta
	if initialState == nil {
		initialState = agentflow.NewState()
	}
	for k, v := range event.GetMetadata() {
		initialState.SetMeta(k, v)
	}

	// Create context for the agent run
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// --- Invoke BeforeAgentRun Hook ---
	agentName := h.agentName // Capture agent name for the callback args
	registry.Invoke(agentflow.CallbackArgs{
		Ctx:       ctx,
		Hook:      agentflow.HookBeforeAgentRun,
		Event:     &event,
		AgentName: &agentName,
		Input:     &initialState, // <<< Pass pointer to interface variable
	})
	// ---------------------------------

	// Run the actual agent logic
	// Declare finalState as interface type agentflow.State
	finalState, agentErr := h.agent.Run(ctx, initialState) // Run returns State interface

	// --- Invoke AfterAgentRun Hook ---
	// This hook runs regardless of whether agent.Run returned an error
	registry.Invoke(agentflow.CallbackArgs{
		Ctx:       ctx,
		Hook:      agentflow.HookAfterAgentRun,
		Event:     &event,
		AgentName: &agentName,
		Input:     &initialState, // <<< Pass pointer to interface variable
		Output:    &finalState,   // <<< Pass pointer to interface variable
		Error:     &agentErr,     // Pass pointer to the error returned by agent.Run
	})
	// --------------------------------

	// Log if the agent returned an actual error (vs. managing errors via State)
	if agentErr != nil {
		log.Printf("[%s Handler] Agent run for event %s returned an error: %v", h.agentName, event.GetID(), agentErr)
		// return agentErr // Option: Propagate error from Handle
	}

	// Ensure finalState is not nil before sending (agent might return nil state on error)
	if finalState == nil {
		log.Printf("[%s Handler] Warning: Agent returned nil final state for event %s. Sending empty state.", h.agentName, event.GetID())
		finalState = agentflow.NewState() // Send an empty state instead of nil
	}

	// Send the final state back to the main goroutine via the channel
	select {
	case h.resultChan <- finalState: // Send interface value
		// Successfully sent result
	case <-ctx.Done(): // Handle potential timeout if channel is blocked
		log.Printf("[%s Handler] Timeout waiting to send result for event %s", h.agentName, event.GetID())
	default:
		// Should not happen with a buffered channel unless buffer is full and main isn't reading
		log.Printf("[%s Handler] Warning: Result channel blocked or closed for event %s.", h.agentName, event.GetID())
	}

	// Return nil assuming the handler's responsibility is just to invoke the agent
	// and manage callbacks. Errors from the agent itself are handled via state or logged.
	return nil
}

// --- Main Execution ---

func main() {
	log.Println("Starting Memory Agent Example...")

	// 1. Setup Session Store
	sessionStore := agentflow.NewMemorySessionStore()
	log.Println("MemorySessionStore created.")

	// 2. Create Agent
	counterAgent := NewCounterAgent(sessionStore)
	log.Println("CounterAgent created.")

	// 3. Setup Runner, TraceLogger & Orchestrator
	runner := agentflow.NewRunner(4) // Buffer size 4

	// Create and set TraceLogger
	traceLogger := agentflow.NewInMemoryTraceLogger()
	err := runner.SetTraceLogger(traceLogger) // Set the trace logger
	if err != nil {
		log.Fatalf("Failed to set trace logger: %v", err)
	}
	log.Println("InMemoryTraceLogger created and set.")

	// Create Orchestrator, passing the Runner's registry
	orchestratorImpl := orchestrator.NewRouteOrchestrator(runner.GetCallbackRegistry())

	// Set the Orchestrator for the Runner
	err = runner.SetOrchestrator(orchestratorImpl)
	if err != nil {
		log.Fatalf("Failed to set orchestrator: %v", err)
	}
	log.Println("Runner, Logger & RouteOrchestrator created and linked.")

	// --- Register Callbacks ---
	// Create the trace logging callback
	traceCallback := agentflow.CreateTraceCallback(traceLogger)

	// Register the trace callback for relevant hooks
	// Using the constant name defined in trace.go
	runner.RegisterCallback(agentflow.HookBeforeEventHandling, agentflow.TraceCallbackName, traceCallback)
	runner.RegisterCallback(agentflow.HookAfterEventHandling, agentflow.TraceCallbackName, traceCallback)
	runner.RegisterCallback(agentflow.HookBeforeAgentRun, agentflow.TraceCallbackName, traceCallback)
	runner.RegisterCallback(agentflow.HookAfterAgentRun, agentflow.TraceCallbackName, traceCallback)
	log.Println("Trace callback registered for hooks.")

	// Optional: Keep the simple logger too? Or remove it. Let's remove it for clarity.
	// runner.RegisterCallback(agentflow.HookBeforeEventHandling, "Logger", loggingCallback)
	// ... register simple logger for other hooks ...
	// --------------------------

	// 4. Setup Handler & Register Agent
	const numEvents = 2
	resultChan := make(chan agentflow.State, numEvents) // Buffered channel for results
	var wg sync.WaitGroup                               // WaitGroup for handler completion

	handler := NewSimpleHandler(CounterAgentName, counterAgent, resultChan, &wg)

	// Register the agent with the runner (which delegates to the orchestrator)
	err = runner.RegisterAgent(CounterAgentName, handler)
	if err != nil {
		log.Fatalf("Failed to register agent: %v", err)
	}
	log.Println("SimpleHandler created and agent registered.")

	// 5. Start the Runner's processing loop
	runner.Start()

	// 6. Emit Events for the SAME session
	mySessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	log.Printf("Using Session ID for this run: %s", mySessionID)

	for i := 1; i <= numEvents; i++ {
		eventID := fmt.Sprintf("event-%d-for-%s", i, mySessionID)
		eventPayload := map[string]interface{}{
			"user_request": fmt.Sprintf("Request number %d", i),
		}
		eventMetadata := map[string]string{
			orchestrator.RouteMetadataKey: CounterAgentName, // Route directly to counter
			agentflow.SessionIDKey:        mySessionID,      // *** Use the same Session ID ***
		}
		event := &agentflow.SimpleEvent{
			ID:       eventID,
			Payload:  eventPayload,
			Metadata: eventMetadata,
		}

		log.Printf("Emitting event %s...", event.GetID())
		wg.Add(1) // Increment WaitGroup *before* emitting
		emitErr := runner.Emit(event)
		if emitErr != nil {
			log.Printf("Error emitting event %s: %v", event.GetID(), emitErr)
			wg.Done() // Decrement if emit fails immediately
		}
		time.Sleep(50 * time.Millisecond) // Small delay for clearer logs
	}

	// 7. Wait for Handlers to Finish
	log.Println("Waiting for handlers to finish...")
	wg.Wait() // Wait for all Handle calls to return (wg.Done() is called in Handle)
	log.Println("Handlers finished.")

	// 8. Stop the Runner Gracefully
	// This signals the loop to stop, waits for emits, closes queue, waits for loop exit.
	runner.Stop()
	log.Println("Runner stopped.")

	// 9. Collect and Print Results (Close channel *after* runner stop and wg.Wait)
	close(resultChan) // Close channel now that all handlers are done and runner is stopped
	results := make([]agentflow.State, 0, numEvents)
	for state := range resultChan {
		results = append(results, state)
	}

	fmt.Println("\n--- Results ---")
	for i, state := range results {
		fmt.Printf("Result %d:\n", i+1)
		// Check for errors reported within the state
		if errVal, ok := state.Get("agent_error"); ok {
			fmt.Printf("  Agent Error: %v\n", errVal)
		} else if errVal, ok := state.Get("handler_error"); ok { // Example if handler added errors
			fmt.Printf("  Handler Error: %v\n", errVal)
		} else {
			// Print normal results
			count, _ := state.Get("current_count")
			isNew, _ := state.Get("is_new_session")
			req, _ := state.Get("user_request")
			fmt.Printf("  Request: %v\n", req)
			fmt.Printf("  Session ID: %v\n", mySessionID)
			fmt.Printf("  Is New Session: %v\n", isNew)
			fmt.Printf("  Current Count: %v\n", count)
		}
		fmt.Println("---------------")
	}

	// --- Dump Trace ---
	log.Println("Retrieving trace via runner for session:", mySessionID)
	finalTrace, traceErr := runner.DumpTrace(mySessionID)
	if traceErr != nil {
		log.Printf("Error retrieving trace via runner: %v", traceErr)
	} else {
		// --- Write Trace to File ---
		traceFilename := fmt.Sprintf("%s.trace.json", mySessionID)
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

// Helper functions for safe logging in callbacks
func safeEventID(ev *agentflow.Event) string {
	if ev != nil && *ev != nil {
		return (*ev).GetID()
	}
	return "<nil>"
}

func safeAgentName(name *string) string {
	if name != nil {
		return *name
	}
	return "<nil>"
}
