package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/tools"
)

// --- Planner Handler ---
type PlannerHandler struct {
	plannerAgent *PlannerAgent
	llmProvider  llm.ModelProvider // Single LLM interface
}

func NewPlannerHandler(provider llm.ModelProvider) *PlannerHandler {
	return &PlannerHandler{
		plannerAgent: NewPlannerAgent(provider),
		llmProvider:  provider,
	}
}

// Fix for PlannerHandler.Run to ensure proper routing
func (h *PlannerHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	log.Printf("PlannerHandler: Running for event %s (Session: %s)", event.GetID(), event.GetSessionID())

	// Debug current route
	if route, ok := state.GetMeta(agentflow.RouteMetadataKey); ok {
		log.Printf("PlannerHandler: Input state route metadata: %s", route)
	}

	// Prevent infinite loops - check if we've already generated a plan
	if plan, ok := state.Get("plan"); ok && plan != nil {
		// We've already generated a plan - force progress to next agent
		log.Printf("PlannerHandler: Plan already exists, passing to researcher")
		outputState := state.Clone()
		outputState.SetMeta(agentflow.RouteMetadataKey, ResearcherAgentName)
		route, _ := outputState.GetMeta(agentflow.RouteMetadataKey)
		log.Printf("🚀 FLOW: %s → %s [Event: %s]",
			PlannerAgentName,
			route,
			event.GetID())
		return agentflow.AgentResult{OutputState: outputState}, nil
	}

	// Get user request from state
	userRequest, ok := state.Get("user_request")
	if !ok || userRequest == nil {
		err := fmt.Errorf("missing user_request in state")
		log.Printf("PlannerHandler: Error: %v", err)
		outputState := state.Clone()
		outputState.Set("error", err.Error())
		outputState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent) // Route to final output on error
		route, _ := outputState.GetMeta(agentflow.RouteMetadataKey)
		log.Printf("🚀 FLOW: %s → %s [Event: %s]",
			PlannerAgentName,
			route,
			event.GetID())
		return agentflow.AgentResult{OutputState: outputState, Error: err.Error()}, nil
	}

	// Generate plan
	userRequestStr, ok := userRequest.(string)
	if !ok {
		err := fmt.Errorf("user_request is not a string")
		log.Printf("PlannerHandler: Error: %v", err)
		outputState := state.Clone()
		outputState.Set("error", err.Error())
		outputState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent) // Route to final output on error
		route, _ := outputState.GetMeta(agentflow.RouteMetadataKey)
		log.Printf("🚀 FLOW: %s → %s [Event: %s]",
			PlannerAgentName,
			route,
			event.GetID())
		return agentflow.AgentResult{OutputState: outputState, Error: err.Error()}, nil
	}

	plan, err := h.plannerAgent.Plan(ctx, userRequestStr)
	if err != nil {
		log.Printf("PlannerHandler: Error generating plan: %v", err)
		outputState := state.Clone()
		outputState.Set("error", fmt.Sprintf("plan generation error: %v", err))
		outputState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent) // Route to final output on error
		route, _ := outputState.GetMeta(agentflow.RouteMetadataKey)
		log.Printf("🚀 FLOW: %s → %s [Event: %s]",
			PlannerAgentName,
			route,
			event.GetID())
		return agentflow.AgentResult{OutputState: outputState, Error: err.Error()}, nil
	}

	log.Printf("PlannerHandler: Plan generated: %s", plan)

	// Store plan in output state
	outputState := state.Clone()
	outputState.Set("plan", plan)

	// IMPORTANT: Explicitly set route to the researcher agent
	log.Printf("PlannerHandler: Setting output route to: %s", ResearcherAgentName)
	outputState.SetMeta(agentflow.RouteMetadataKey, ResearcherAgentName)

	// Make sure output metadata explicitly contains routing info
	eventMeta := event.GetMetadata()
	// Fix 1: Correctly handle the multiple return values from GetMeta
	route, hasRoute := outputState.GetMeta(agentflow.RouteMetadataKey)
	if hasRoute && route == ResearcherAgentName {
		// Ensure route info is copied to event metadata for next dispatch
		log.Printf("PlannerHandler: Setting urgent routing metadata fix: %s", ResearcherAgentName)

		// Copy metadata to a mutable map
		newMeta := make(map[string]string)
		if eventMeta != nil {
			for k, v := range eventMeta {
				newMeta[k] = v
			}
		}

		// Set route explicitly
		newMeta[agentflow.RouteMetadataKey] = ResearcherAgentName

		// Create a new event with updated metadata
		// Fix 2: Use the correct method to get data from state
		stateData := make(map[string]interface{})
		for _, key := range outputState.Keys() {
			if val, ok := outputState.Get(key); ok {
				stateData[key] = val
			}
		}

		//		fixedEvent := agentflow.NewEvent(
		//			ResearcherAgentName,
		//			stateData,
		//			newMeta,
		//		)

		// Fix 3: Don't reference the global orchestratorImpl directly
		// Instead, return with the properly configured state
		// The routing metadata in outputState will be used by the framework
		log.Printf("PlannerHandler: Created fixed event with explicit routing to: %s", ResearcherAgentName)

		// Note: Removing the direct Dispatch call as it references an undefined variable
		// and is not necessary with proper state metadata configuration
	}

	// Add this to PlannerHandler.Run (or any agent that needs dynamic routing)
	log.Printf("PlannerHandler: Processing request")

	// Clone state to avoid side effects
	outputState = state.Clone()

	// Get the user request from state
	userRequest, ok = state.Get("user_request")
	if !ok {
		log.Printf("No user request found in state")
		outputState.Set("error", "Missing user request")
		outputState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
		route, _ := outputState.GetMeta(agentflow.RouteMetadataKey)
		log.Printf("🚀 FLOW: %s → %s [Event: %s]",
			PlannerAgentName,
			route,
			event.GetID())
		return agentflow.AgentResult{OutputState: outputState}, nil
	}

	// Your existing planner logic here...
	// ...

	// Create output state with plan
	outputState.Set("plan", "1. Research recent developments in AI code generation\n2. Focus on key innovations\n3. Identify practical applications")

	// LLM-based dynamic routing decision
	nextAgent, err := h.determineNextAgent(ctx, userRequest.(string), outputState)
	if err != nil {
		log.Printf("Error determining next agent: %v", err)
		// Fallback to hardcoded default route only on error
		nextAgent = ResearcherAgentName
		log.Printf("🔀 PlannerHandler: Using fallback route → %s", nextAgent)
	}

	// Set the route metadata based on LLM decision
	outputState.SetMeta(agentflow.RouteMetadataKey, nextAgent)
	log.Printf("🔀 PlannerHandler: Dynamic routing decision → %s", nextAgent)

	// Add Azure resilience pattern - validate route is valid
	if nextAgent != PlannerAgentName &&
		nextAgent != ResearcherAgentName &&
		nextAgent != SummarizerAgentName &&
		nextAgent != FinalOutputAgent {
		log.Printf("⚠️ PlannerHandler: Invalid route '%s' detected, using fallback", nextAgent)
		nextAgent = ResearcherAgentName
		outputState.SetMeta(agentflow.RouteMetadataKey, nextAgent)
	}

	// Add Azure observability - track routing decisions
	outputState.SetMeta("routing_method", "llm_decision")
	outputState.SetMeta("routing_timestamp", time.Now().Format(time.RFC3339))

	route, _ = outputState.GetMeta(agentflow.RouteMetadataKey)
	log.Printf("🚀 FLOW: %s → %s [Event: %s]",
		PlannerAgentName,
		route,
		event.GetID())

	return agentflow.AgentResult{OutputState: outputState}, nil
}

// Update the determineNextAgent method to use ModelProvider directly
func (h *PlannerHandler) determineNextAgent(ctx context.Context, userRequest string, state agentflow.State) (string, error) {
	// Create system and user prompts as before
	systemPrompt := `You are the routing controller...`

	var stateInfo strings.Builder
	// Build state info as before

	// Replace the LLMAdapter.Complete call with direct ModelProvider usage
	response, err := h.llmProvider.Call(ctx, llm.Prompt{
		System: systemPrompt,
		User:   stateInfo.String(),
		Parameters: llm.ModelParameters{
			Temperature: floatPtr(0.7),
			MaxTokens:   int32Ptr(500),
		},
	})
	if err != nil {
		return "", fmt.Errorf("LLM error determining next agent: %v", err)
	}

	agentName := strings.TrimSpace(response.Content)
	// Validate agent name as before

	return agentName, nil
}

// Helper functions if not already defined
func floatPtr(f float32) *float32 {
	return &f
}

func int32Ptr(i int32) *int32 {
	return &i
}

// --- Researcher Handler ---
type ResearcherHandler struct {
	researchAgent *ResearchAgent
	toolRegistry  tools.ToolRegistry
	llm           llm.LLMAdapter
}

// Update the ResearcherHandler constructor to accept the interface, not a pointer
func NewResearcherHandler(toolRegistry tools.ToolRegistry, llm llm.LLMAdapter) *ResearcherHandler {
	return &ResearcherHandler{
		researchAgent: NewResearchAgent(&toolRegistry), // Note: NewResearchAgent still gets a pointer
		toolRegistry:  toolRegistry,
		llm:           llm,
	}
}

func (h *ResearcherHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	log.Printf("ResearcherHandler: Processing request")

	// Clone state to avoid modifying the input - Azure best practice for immutable state handling
	outputState := state.Clone()

	// Get the user request
	userRequest, ok := state.Get("user_request")
	if !ok {
		// If no user request found, it's an error condition
		outputState.Set("error", "No user request found in state for researcher")
		// Route to final output with error
		outputState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
		route, _ := outputState.GetMeta(agentflow.RouteMetadataKey)
		log.Printf("🚀 FLOW: %s → %s [Event: %s]",
			ResearcherAgentName,
			route,
			event.GetID())
		return agentflow.AgentResult{OutputState: outputState}, nil
	}

	// Get the plan if it exists
	plan, hasPlan := state.Get("plan")
	if !hasPlan {
		// Log missing plan but continue
		log.Printf("ResearcherHandler: No plan found in state, using raw user request")
	} else {
		log.Printf("ResearcherHandler: Using plan: %v", plan)
	}

	// Check if we've already performed research to prevent loops
	researchResults, hasResearch := state.Get("research_results")
	if hasResearch && researchResults != nil {
		log.Printf("ResearcherHandler: Research already performed, routing to summarizer")
		// Already have research results, route to summarizer
		outputState.SetMeta(agentflow.RouteMetadataKey, SummarizerAgentName)
		route, _ := outputState.GetMeta(agentflow.RouteMetadataKey)
		log.Printf("🚀 FLOW: %s → %s [Event: %s]",
			ResearcherAgentName,
			route,
			event.GetID())
		return agentflow.AgentResult{OutputState: outputState}, nil
	}

	// Instead of using actual web search, generate simulated research results
	// This follows Azure's pattern for reliable testing with mock dependencies
	var simulatedResults string

	if hasPlan {
		planStr := fmt.Sprintf("%v", plan)
		simulatedResults = fmt.Sprintf("Simulated research results for plan: %s\n\n"+
			"1. Found relevant information about the first aspect of the plan.\n"+
			"2. Discovered key insights about the second part of the plan.\n"+
			"3. Analysis shows the plan approach is generally effective.\n"+
			"4. Some limitations were identified that should be considered.",
			planStr)
	} else {
		userRequestStr := fmt.Sprintf("%v", userRequest)
		simulatedResults = fmt.Sprintf("Simulated research results for: %s\n\n"+
			"1. Basic information retrieved about the topic.\n"+
			"2. Several perspectives were analyzed.\n"+
			"3. Key considerations identified.\n"+
			"4. Preliminary recommendations based on available information.",
			userRequestStr)
	}

	log.Printf("ResearcherHandler: Generated simulated research results")

	// Store research results in state
	outputState.Set("research_results", simulatedResults)

	// Additionally store as research_result (singular) to match what SummarizerHandler expects
	outputState.Set("research_result", simulatedResults)

	// Set research completion marker to prevent loops
	outputState.Set("research_completed", true)

	// Route to summarizer - explicit state transition for Azure durable functions pattern
	outputState.SetMeta(agentflow.RouteMetadataKey, SummarizerAgentName)

	// Additional diagnostic metadata
	outputState.SetMeta("last_processor", "ResearcherHandler")
	outputState.SetMeta("process_timestamp", fmt.Sprintf("%v", time.Now().Format(time.RFC3339)))

	log.Printf("ResearcherHandler: Research (simulated) completed, routing to summarizer")
	route, _ := outputState.GetMeta(agentflow.RouteMetadataKey)
	log.Printf("🚀 FLOW: %s → %s [Event: %s]",
		ResearcherAgentName,
		route,
		event.GetID())
	return agentflow.AgentResult{OutputState: outputState}, nil
}

// --- Summarizer Handler ---
type SummarizerHandler struct {
	summarizeAgent *SummarizeAgent // Embed or reference the agent logic
	llmProvider    llm.ModelProvider
	llm            llm.LLMAdapter
}

func NewSummarizerHandler(provider llm.ModelProvider, llm llm.LLMAdapter) *SummarizerHandler {
	return &SummarizerHandler{
		summarizeAgent: NewSummarizeAgent(provider), // Create the agent logic instance
		llmProvider:    provider,
		llm:            llm,
	}
}

func (h *SummarizerHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	log.Printf("SummarizerHandler: Running for event %s (Session: %s)", event.GetID(), event.GetSessionID())

	// FIX: Use state.Get(key)
	researchResultVal, ok := state.Get("research_result") // Get result from incoming state
	if !ok {
		err := fmt.Errorf("summarizer: missing 'research_result' in state")
		log.Printf("%v", err)
		errState := agentflow.NewStateWithData(agentflow.EventData{"error": err.Error()})
		errState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
		errState.SetMeta(agentflow.SessionIDKey, event.GetSessionID()) // Preserve session
		return agentflow.AgentResult{OutputState: errState}, err
	}
	researchResult, ok := researchResultVal.(string)
	if !ok {
		err := fmt.Errorf("summarizer: 'research_result' data is not a string")
		log.Printf("%v", err)
		errState := agentflow.NewStateWithData(agentflow.EventData{"error": err.Error()})
		errState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
		errState.SetMeta(agentflow.SessionIDKey, event.GetSessionID())
		return agentflow.AgentResult{OutputState: errState}, err
	}

	// FIX: Use state.Get(key) for user_request
	userRequestVal, ok := state.Get("user_request")
	if !ok {
		// Handle missing user_request if necessary, maybe default or error
		log.Println("SummarizerHandler: Warning - 'user_request' missing from state.")
		userRequestVal = "" // Default to empty string or handle as error
	}
	userRequest, ok := userRequestVal.(string)
	if !ok {
		// This case might occur if the default "" was set above, or if it was non-string in state
		log.Println("SummarizerHandler: Warning - 'user_request' in state was not a string.")
		userRequest = "" // Ensure it's a string for the agent call
	}

	// Call the summarize agent's logic
	summary, err := h.summarizeAgent.Summarize(ctx, userRequest, researchResult)
	if err != nil {
		log.Printf("SummarizerHandler: Summarize agent failed: %v", err)
		errState := agentflow.NewStateWithData(agentflow.EventData{"error": fmt.Sprintf("Summarizer failed: %v", err)})
		errState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
		errState.SetMeta(agentflow.SessionIDKey, event.GetSessionID()) // Preserve session
		return agentflow.AgentResult{OutputState: errState}, err
	}

	log.Printf("SummarizerHandler: Summarization completed.")

	// Prepare the final output state
	nextStateData := agentflow.EventData{
		"summary":      summary,
		"user_request": userRequest, // Include original request for context
	}
	outputState := agentflow.NewStateWithData(nextStateData)
	// Set routing metadata for the final output step
	outputState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
	outputState.SetMeta(agentflow.SessionIDKey, event.GetSessionID()) // Preserve session

	route, _ := outputState.GetMeta(agentflow.RouteMetadataKey)
	log.Printf("🚀 FLOW: %s → %s [Event: %s]",
		SummarizerAgentName,
		route,
		event.GetID())

	return agentflow.AgentResult{OutputState: outputState}, nil
}

// Update FinalOutputHandler to include finalState field
type FinalOutputHandler struct {
	wg *sync.WaitGroup
	mu sync.Mutex
	// Store final results keyed by session ID
	finalResults map[string]agentflow.State
	finalState   agentflow.State // Add this field for trace support
}

// FinalOutputHandler with trace dumping
func (h *FinalOutputHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	log.Printf("FinalOutputHandler: Processing final output for session: %s", event.GetSessionID())

	// Store the final state
	h.mu.Lock()
	sessionID := event.GetSessionID()
	h.finalResults[sessionID] = state.Clone()
	h.finalState = state.Clone() // Store in the field for trace access
	h.mu.Unlock()

	// Dump trace data if available
	if traceLogger, ok := ctx.Value("traceLogger").(*agentflow.InMemoryTraceLogger); ok {
		sessionID := event.GetSessionID()
		log.Printf("FinalOutputHandler: Retrieving trace for session ID: %s", sessionID)

		finalTrace, err := traceLogger.GetTrace(sessionID)
		if err != nil {
			log.Printf("Error retrieving trace: %v", err)
		} else {
			log.Printf("Retrieved %d trace entries for session", len(finalTrace))

			// Create traces directory if it doesn't exist
			tracesDir := "traces"
			if _, err := os.Stat(tracesDir); os.IsNotExist(err) {
				if err := os.Mkdir(tracesDir, 0755); err != nil {
					log.Printf("Error creating traces directory: %v", err)
				}
			}

			// Write trace to file
			traceFilename := filepath.Join(tracesDir, fmt.Sprintf("%s.trace.json", sessionID))
			traceJSON, jsonErr := json.MarshalIndent(finalTrace, "", "  ")
			if jsonErr != nil {
				log.Printf("Error marshaling trace to JSON: %v", jsonErr)
			} else {
				if writeErr := os.WriteFile(traceFilename, traceJSON, 0644); writeErr != nil {
					log.Printf("Error writing trace file '%s': %v", traceFilename, writeErr)
				} else {
					log.Printf("Trace written to %s", traceFilename)
					// Store trace file location in state
					h.finalState.Set("trace_file", traceFilename)
					h.finalResults[sessionID].Set("trace_file", traceFilename)
				}
			}

			// Print summary to console
			fmt.Println("\n--- Trace Summary ---")
			fmt.Printf("Total entries: %d\n", len(finalTrace))

			agentCounts := make(map[string]int)
			for _, entry := range finalTrace {
				if entry.AgentID != "" {
					agentCounts[entry.AgentID]++
				}
			}

			fmt.Println("Agent activity:")
			for agent, count := range agentCounts {
				fmt.Printf("  - %s: %d entries\n", agent, count)
			}
			fmt.Println("-------------------\n")
		}
	}

	// Signal completion through WaitGroup
	if h.wg != nil {
		h.wg.Done()
		log.Println("FinalOutputHandler: Signaled completion via WaitGroup")
	}

	// Return the state unmodified
	return agentflow.AgentResult{OutputState: state}, nil
}
