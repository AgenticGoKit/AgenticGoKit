package main

import (
	"context" // Import context
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/orchestrator"
	"kunalkushwaha/agentflow/internal/tools"
)

const (
	PlannerAgentName    = "planner"
	ResearcherAgentName = "researcher"
	SummarizerAgentName = "summarizer"
	FinalOutputAgent    = "final_output" // A conceptual sink or final step
)

func NewFinalOutputHandler(wg *sync.WaitGroup) *FinalOutputHandler {
	return &FinalOutputHandler{
		wg:           wg,
		finalResults: make(map[string]agentflow.State),
	}
}

// Add debugging in the FinalOutputHandler to check routes
// (Implementation in handler.go)

func (h *FinalOutputHandler) GetFinalState(sessionID string) (agentflow.State, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	state, found := h.finalResults[sessionID]
	return state, found
}

func main() {
	log.Println("Starting Multi-Agent Team Example...")

	// --- Configuration ---
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")
	if embeddingDeployment == "" {
		embeddingDeployment = "not-used"
	}
	if endpoint == "" || apiKey == "" || chatDeployment == "" {
		log.Fatal("Error: Set AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, and AZURE_OPENAI_CHAT_DEPLOYMENT.")
	}
	// TODO: Add env var for WebSearchTool API Key if needed

	// --- Setup ---
	// 1. LLM Adapter
	adapterOpts := llm.AzureOpenAIAdapterOptions{
		Endpoint:            endpoint,
		APIKey:              apiKey,
		ChatDeployment:      chatDeployment,
		EmbeddingDeployment: embeddingDeployment,
		HTTPClient:          &http.Client{Timeout: 90 * time.Second},
	}
	azureAdapter, err := llm.NewAzureOpenAIAdapter(adapterOpts)
	if err != nil {
		log.Fatalf("Error creating Azure OpenAI Adapter: %v", err)
	}
	log.Println("Azure OpenAI Adapter created.")

	// Add these lines after creating azureAdapter
	// Create an adapter that implements the LLMAdapter interface
	llmAdapter := llm.NewModelProviderAdapter(azureAdapter)

	// 2. Tool Registry
	toolRegistry := tools.NewToolRegistry()
	searchTool := &tools.WebSearchTool{} // Assumes WebSearchTool struct exists
	if err := toolRegistry.Register(searchTool); err != nil {
		log.Fatalf("Failed to register web search tool: %v", err)
	}
	log.Println("Tool Registry created and WebSearchTool registered.")

	// 3. Create Core Components
	callbackRegistry := agentflow.NewCallbackRegistry() // Create registry
	// TODO: Register callbacks if needed (e.g., for logging)
	// callbackRegistry.Register(agentflow.HookAfterAgentRun, "logAgent", logAgentCallback)

	// Add this event deduplication callback before your existing callbacks
	callbackRegistry.Register(agentflow.HookBeforeAgentRun, "eventDeduplicator",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			if args.Event == nil {
				return args.State, nil
			}

			// Get processed events from context
			var processedEvents map[string]bool
			if pe, ok := ctx.Value("processed_events").(map[string]bool); ok {
				processedEvents = pe
			} else {
				// Initialize if not present
				processedEvents = make(map[string]bool)
				ctx = context.WithValue(ctx, "processed_events", processedEvents)
			}

			// Create a unique event signature
			eventID := args.Event.GetID()
			agentID := args.AgentID
			eventKey := fmt.Sprintf("%s-%s", eventID, agentID)

			// Check if this exact event-agent combo was already processed
			if processedEvents[eventKey] {
				log.Printf("⚠️ DEDUPLICATION: Skipping duplicate event %s for agent %s", eventID, agentID)

				// Set skip flag in state to prevent processing
				args.State.SetMeta("skip_processing", "true")
				return args.State, nil
			}

			// Mark as processed
			processedEvents[eventKey] = true
			log.Printf("✅ DEDUPLICATION: First processing of event %s for agent %s", eventID, agentID)

			return args.State, nil
		})

	// Add after creating callbackRegistry
	// Register a route tracing callback
	callbackRegistry.Register(agentflow.HookAll, "routeTracer",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			// Only trace BeforeAgentRun and AfterAgentRun hooks for cleaner logs
			if args.Hook == agentflow.HookBeforeAgentRun || args.Hook == agentflow.HookAfterAgentRun {
				if route, ok := args.State.GetMeta(agentflow.RouteMetadataKey); ok {
					log.Printf("ROUTE TRACE [%s]: Agent=%s, Route=%s",
						string(args.Hook), args.AgentID, route)
				} else {
					log.Printf("ROUTE TRACE [%s]: Agent=%s, Route=<none>",
						string(args.Hook), args.AgentID)
				}
			}
			return args.State, nil
		})
	log.Println("Route tracing callback registered.")

	// Add after creating callbackRegistry
	callbackRegistry.Register(agentflow.HookBeforeEventHandling, "routeDebugger",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			if args.Event != nil {
				meta := args.Event.GetMetadata()
				route := ""
				if meta != nil {
					route = meta[agentflow.RouteMetadataKey]
				}
				log.Printf("ROUTE DEBUG: Event %s being routed to agent: %s",
					args.Event.GetID(), route)
			}
			return args.State, nil
		})

	// Add after other callbacks, around line ~107
	callbackRegistry.Register(agentflow.HookBeforeAgentRun, "transitionCounter",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			// Get transition count from state
			transitionCount := 0
			if countVal, ok := args.State.Get("_transition_count"); ok {
				if count, ok := countVal.(int); ok {
					transitionCount = count
				}
			}

			// Increment count
			transitionCount++
			args.State.Set("_transition_count", transitionCount)

			// Break loops if too many transitions
			if transitionCount > 10 {
				log.Printf("WARNING: Detected potential infinite loop - exceeded 10 transitions")
				args.State.SetMeta(agentflow.RouteMetadataKey, "")
				args.State.Set("error", "Potential infinite loop detected and stopped")
				args.State.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
			}

			return args.State, nil
		})
	log.Println("Transition counter callback registered for loop prevention")

	// Add this progress tracker callback
	callbackRegistry.Register(agentflow.HookAfterAgentRun, "progressTracker",
		func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
			// Get session ID
			sessionID := "unknown"
			if args.Event != nil {
				sessionID = args.Event.GetSessionID()
			}

			// Get the visit count map from context
			var visitCounts map[string]int
			if vc, ok := ctx.Value("agent_visit_counts").(map[string]int); ok {
				visitCounts = vc
			} else {
				visitCounts = make(map[string]int)
				ctx = context.WithValue(ctx, "agent_visit_counts", visitCounts)
			}

			// Create a visit key with session and agent
			visitKey := fmt.Sprintf("%s-%s", sessionID, args.AgentID)

			// Increment visit count
			visitCounts[visitKey]++

			// Get current visit count
			visitCount := visitCounts[visitKey]

			// Circuit breaker pattern - if an agent has been visited too many times, force to final output
			if visitCount > 3 {
				log.Printf("🔄 CIRCUIT BREAKER: Agent %s has been visited %d times for session %s - forcing to final output",
					args.AgentID, visitCount, sessionID)

				// Force route to final output to break potential loops
				args.State.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
				args.State.Set("circuit_breaker_activated", true)
				args.State.Set("error", fmt.Sprintf("Circuit breaker triggered after %d visits to %s",
					visitCount, args.AgentID))
			}

			return args.State, nil
		})

	orchestratorImpl := orchestrator.NewRouteOrchestrator(callbackRegistry) // Pass registry
	concurrency := runtime.NumCPU()
	runner := agentflow.NewRunner(concurrency)   // Pass queue size
	runner.SetOrchestrator(orchestratorImpl)     // Set orchestrator
	runner.SetCallbackRegistry(callbackRegistry) // Set registry
	log.Printf("Runner & Route Orchestrator created (concurrency %d).", concurrency)

	// Set the runner as the emitter for the orchestrator
	// This is critical for proper routing between agents
	orchestratorImpl.SetEmitter(runner)
	log.Println("RouteOrchestrator configured with runner as emitter")

	// Setup tracing
	traceLogger := agentflow.NewInMemoryTraceLogger()
	runner.SetTraceLogger(traceLogger)
	// log.Println("Trace logging enabled for observability")

	// // Register trace callback for all hooks
	// callbackRegistry.Register(agentflow.HookAll, "traceLogger",
	// 	func(ctx context.Context, args agentflow.CallbackArgs) (agentflow.State, error) {
	// 		// Get session ID from state or event
	// 		var sessionID string

	// 		// Try to get from state metadata first
	// 		if sid, ok := args.State.GetMeta(agentflow.SessionIDKey); ok && sid != "" {
	// 			sessionID = sid
	// 		} else if args.Event != nil {
	// 			// Fall back to event metadata
	// 			if meta := args.Event.GetMetadata(); meta != nil {
	// 				if sid, ok := meta[agentflow.SessionIDKey]; ok && sid != "" {
	// 					sessionID = sid
	// 					// Add to state metadata for future use
	// 					args.State.SetMeta(agentflow.SessionIDKey, sid)
	// 				}
	// 			}
	// 		}

	// 		// Default to event ID if still no session ID
	// 		if sessionID == "" && args.Event != nil {
	// 			sessionID = "session-" + args.Event.GetID()
	// 		}

	// 		entry := agentflow.TraceEntry{
	// 			Hook:      args.Hook,
	// 			Timestamp: time.Now(),
	// 			State:     args.State,
	// 			SessionID: sessionID,
	// 		}

	// 		if args.Event != nil {
	// 			entry.EventID = args.Event.GetID()
	// 		}

	// 		if args.AgentID != "" {
	// 			entry.AgentID = args.AgentID
	// 		}

	// 		if args.Output.OutputState != nil {
	// 			entry.AgentResult = &args.Output
	// 		}

	// 		if args.Output.Error != "" {
	// 			entry.Error = args.Output.Error
	// 		}

	// 		if err := traceLogger.Log(entry); err != nil {
	// 			log.Printf("Error logging trace entry: %v", err)
	// 		}
	// 		return args.State, nil
	// 	})
	// log.Println("Trace callback registered for all hooks")

	// Add trace context to the initialContext
	ctx := context.Background()
	ctx = context.WithValue(ctx, "traceLogger", traceLogger)
	ctx = context.WithValue(ctx, "processed_events", make(map[string]bool))
	// Initialize context with visit counts map
	ctx = context.WithValue(ctx, "agent_visit_counts", make(map[string]int))
	log.Println("Added traceLogger to context for agent observability")

	// 4. Create Agent Handlers (Defined in handler.go, implementing agentflow.AgentHandler)
	var wg sync.WaitGroup // WaitGroup to track completion of the *flow*
	plannerHandler := NewPlannerHandler(azureAdapter)
	// Fix: Use the interface (not the pointer) and provide the required LLMAdapter
	researcherHandler := NewResearcherHandler(*toolRegistry, llmAdapter)
	// Fix: Provide both required arguments
	summarizerHandler := NewSummarizerHandler(azureAdapter, llmAdapter)
	finalOutputHandler := NewFinalOutputHandler(&wg) // Handler to receive the final result

	// 5. Register Handlers with Orchestrator
	log.Println("Registering agent handlers...")
	requireNoError(orchestratorImpl.RegisterAgent(PlannerAgentName, plannerHandler), "Register planner")
	requireNoError(orchestratorImpl.RegisterAgent(ResearcherAgentName, researcherHandler), "Register researcher")
	requireNoError(orchestratorImpl.RegisterAgent(SummarizerAgentName, summarizerHandler), "Register summarizer")
	requireNoError(orchestratorImpl.RegisterAgent(FinalOutputAgent, finalOutputHandler), "Register final output")
	log.Println("All agent handlers registered successfully.")

	// Add this after the call to orchestratorImpl.RegisterAgent()

	// ❶ Create the in‑memory logger
	//traceLogger = agentflow.NewInMemoryTraceLogger()

	// ❷ Wire the logger to every lifecycle hook
	agentflow.RegisterTraceHooks(callbackRegistry, traceLogger)

	// ❸ (optional) keep a direct reference on the runner – useful for DumpTrace()
	//runner.SetTraceLogger(traceLogger)

	// 6. Start the Runner
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	runner.Start(ctx)                  // Pass context
	time.Sleep(100 * time.Millisecond) // Allow runner to start
	log.Println("Runner started.")

	// --- Execution ---
	// 1. Prepare the initial event
	initialEventID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	userRequest := "Research the recent developments in AI-powered code generation and summarize the key findings."
	eventData := agentflow.EventData{
		"user_request": userRequest,
	}
	eventMeta := map[string]string{
		agentflow.RouteMetadataKey: PlannerAgentName, // Start with the Planner
		agentflow.SessionIDKey:     initialEventID,   // Use SessionIDKey for tracking
	}
	// Use agentflow.NewEvent
	event := agentflow.NewEvent(PlannerAgentName, eventData, eventMeta)
	event.SetID(initialEventID) // Set specific ID if needed, though SessionID is preferred for tracking flow
	log.Printf("Initial event prepared: %s (Session: %s)", event.GetID(), event.GetSessionID())

	// 2. Emit the initial event
	log.Println("Emitting initial event to start the flow...")
	// wg.Add(1) // Add wait group count *if* FinalOutputHandler uses it reliably
	wg.Add(1)                // Add this if using WaitGroup pattern
	err = runner.Emit(event) // Use event instead of initialEvent
	if err != nil {
		log.Fatalf("Failed to emit initial event: %v", err)
	}
	log.Println("Initial event emitted.")

	// 3. Wait for the final result by checking the FinalOutputHandler
	log.Println("Waiting for final result...")
	var finalState agentflow.State
	var found bool
	timeout := time.After(180 * time.Second) // Timeout for the whole flow
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

waitLoop:
	for {
		select {
		case <-timeout:
			log.Fatal("Timeout waiting for the multi-agent flow to complete.")
			break waitLoop // Should not be reached due to log.Fatal
		case <-ticker.C:
			finalState, found = finalOutputHandler.GetFinalState(initialEventID)
			if found {
				log.Println("Final state found.")
				break waitLoop
			}
			log.Println("Still waiting for final state...") // Optional progress log
		case <-ctx.Done():
			log.Printf("Context cancelled while waiting: %v", ctx.Err())
			break waitLoop
		}
	}

	// 4. Stop the runner
	log.Println("Stopping runner...")
	cancel()      // Cancel context first
	runner.Stop() // Then stop runner
	log.Println("Runner stopped.")

	// -----------------------------------------------------------------
	// Persist the trace so that `agentcli trace` can show it later.
	// -----------------------------------------------------------------
	traceDir := "traces"
	if err := os.MkdirAll(traceDir, 0o755); err != nil {
		log.Printf("Failed to create trace directory: %v", err)
	} else {
		traceEntries, err := runner.DumpTrace(initialEventID) // sessionID == initialEventID
		if err != nil {
			log.Printf("DumpTrace error: %v", err)
		} else {
			outPath := filepath.Join(traceDir, initialEventID+".trace.json")
			f, err := os.Create(outPath)
			if err != nil {
				log.Printf("Cannot create trace file: %v", err)
			} else {
				enc := json.NewEncoder(f)
				enc.SetIndent("", "  ")
				if err := enc.Encode(traceEntries); err != nil {
					log.Printf("Failed to write trace file: %v", err)
				} else {
					log.Printf("Trace written to %s", outPath)
					// Store the path in state so downstream CLI can print it
					finalState.Set("trace_file", outPath)
				}
				_ = f.Close()
			}
		}
	}

	// --- Output ---
	if !found || finalState == nil {
		log.Fatal("Failed to get final state.")
	}

	// Check for errors first in the output state
	// FIX: Use finalState.Get(key) instead of GetMust
	if errMsg, ok := finalState.Get("error"); ok && errMsg != nil {
		log.Printf("Flow ended with error: %v", errMsg)
	} else if summary, ok := finalState.Get("summary"); ok && summary != nil {
		log.Println("Multi-agent flow successful.")
		fmt.Println("\n--- Final Summary ---")
		fmt.Printf("%v\n", summary) // Use %v for safety
		fmt.Println("---------------------")
	} else {
		log.Println("Flow ended, but no summary or known error found in final state.")
	}

	// FIX: Use finalState.Get(key) for logging
	finalSummary, _ := finalState.Get("summary")
	finalError, _ := finalState.Get("error")
	log.Printf("Final state summary: %v", finalSummary)
	log.Printf("Final state error: %v", finalError)
	// log.Printf("Final state data: %+v", finalState.GetData()) // Avoid this

	// Add this at the end of main() before displaying the final output
	if finalState != nil {
		if traceFile, ok := finalState.Get("trace_file"); ok {
			fmt.Printf("\nComplete trace saved to: %s\n", traceFile)
			fmt.Println("Review this file for detailed execution flow and debugging")
		}
	}
}

// Helper for checking registration errors
func requireNoError(err error, step string) {
	if err != nil {
		log.Fatalf("Error during setup - %s: %v", step, err)
	}
}
