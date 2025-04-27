package main

import (
	"context"
	"fmt"
	"log"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/tools"
)

// --- Planner Handler ---
type PlannerHandler struct {
	plannerAgent *PlannerAgent // Embed or reference the agent logic
	llmProvider  llm.ModelProvider
}

func NewPlannerHandler(provider llm.ModelProvider) *PlannerHandler {
	return &PlannerHandler{
		plannerAgent: NewPlannerAgent(provider), // Create the agent logic instance
		llmProvider:  provider,
	}
}

func (h *PlannerHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	log.Printf("PlannerHandler: Running for event %s (Session: %s)", event.GetID(), event.GetSessionID())

	// FIX: Use event.GetData() and access the key
	eventData := event.GetData()
	userRequestVal, ok := eventData["user_request"]
	if !ok {
		err := fmt.Errorf("planner: missing 'user_request' in event data")
		log.Printf("%v", err)
		// Return error state
		errState := agentflow.NewStateWithData(agentflow.EventData{"error": err.Error()})
		errState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent) // Route error to final output
		errState.SetMeta(agentflow.SessionIDKey, event.GetSessionID()) // Preserve session
		return agentflow.AgentResult{OutputState: errState}, err
	}
	userRequest, ok := userRequestVal.(string) // Assume string
	if !ok {
		err := fmt.Errorf("planner: 'user_request' data is not a string")
		log.Printf("%v", err)
		errState := agentflow.NewStateWithData(agentflow.EventData{"error": err.Error()})
		errState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
		errState.SetMeta(agentflow.SessionIDKey, event.GetSessionID())
		return agentflow.AgentResult{OutputState: errState}, err
	}

	// Call the planner agent's logic
	plan, err := h.plannerAgent.Plan(ctx, userRequest)
	if err != nil {
		log.Printf("PlannerHandler: Planner agent failed: %v", err)
		errState := agentflow.NewStateWithData(agentflow.EventData{"error": fmt.Sprintf("Planner failed: %v", err)})
		errState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
		errState.SetMeta(agentflow.SessionIDKey, event.GetSessionID()) // Preserve session
		return agentflow.AgentResult{OutputState: errState}, err
	}

	log.Printf("PlannerHandler: Plan generated: %s", plan)

	// Prepare the next state for the researcher
	nextStateData := agentflow.EventData{
		"plan":         plan,
		"user_request": userRequest, // Pass along original request
	}
	outputState := agentflow.NewStateWithData(nextStateData)
	// Set routing metadata for the next step
	outputState.SetMeta(agentflow.RouteMetadataKey, ResearcherAgentName)
	outputState.SetMeta(agentflow.SessionIDKey, event.GetSessionID()) // Preserve session

	return agentflow.AgentResult{OutputState: outputState}, nil
}

// --- Researcher Handler ---
type ResearcherHandler struct {
	researchAgent *ResearchAgent // Embed or reference the agent logic
	toolRegistry  *tools.ToolRegistry
}

func NewResearcherHandler(registry *tools.ToolRegistry) *ResearcherHandler {
	return &ResearcherHandler{
		researchAgent: NewResearchAgent(registry), // Create the agent logic instance
		toolRegistry:  registry,
	}
}

func (h *ResearcherHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	log.Printf("ResearcherHandler: Running for event %s (Session: %s)", event.GetID(), event.GetSessionID())

	// FIX: Use state.Get(key)
	planVal, ok := state.Get("plan") // Get plan from incoming state
	if !ok {
		err := fmt.Errorf("researcher: missing 'plan' in state")
		log.Printf("%v", err)
		errState := agentflow.NewStateWithData(agentflow.EventData{"error": err.Error()})
		errState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
		errState.SetMeta(agentflow.SessionIDKey, event.GetSessionID()) // Preserve session
		return agentflow.AgentResult{OutputState: errState}, err
	}
	plan, ok := planVal.(string)
	if !ok {
		err := fmt.Errorf("researcher: 'plan' data is not a string")
		log.Printf("%v", err)
		errState := agentflow.NewStateWithData(agentflow.EventData{"error": err.Error()})
		errState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
		errState.SetMeta(agentflow.SessionIDKey, event.GetSessionID())
		return agentflow.AgentResult{OutputState: errState}, err
	}

	// Call the research agent's logic
	researchResult, err := h.researchAgent.Research(ctx, plan)
	if err != nil {
		log.Printf("ResearcherHandler: Research agent failed: %v", err)
		errState := agentflow.NewStateWithData(agentflow.EventData{"error": fmt.Sprintf("Researcher failed: %v", err)})
		errState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)
		errState.SetMeta(agentflow.SessionIDKey, event.GetSessionID()) // Preserve session
		return agentflow.AgentResult{OutputState: errState}, err
	}

	log.Printf("ResearcherHandler: Research completed.") // Avoid logging potentially large result here

	// Prepare the next state for the summarizer
	// FIX: Use state.Get(key) for user_request
	userRequestVal, ok := state.Get("user_request")
	if !ok {
		// Handle missing user_request if necessary, maybe default or error
		log.Println("ResearcherHandler: Warning - 'user_request' missing from state.")
		userRequestVal = "" // Default to empty string or handle as error? For now, empty.
	}

	nextStateData := agentflow.EventData{
		"research_result": researchResult,
		"user_request":    userRequestVal, // Pass along original request from state
	}
	outputState := agentflow.NewStateWithData(nextStateData)
	// Set routing metadata for the next step
	outputState.SetMeta(agentflow.RouteMetadataKey, SummarizerAgentName)
	outputState.SetMeta(agentflow.SessionIDKey, event.GetSessionID()) // Preserve session

	return agentflow.AgentResult{OutputState: outputState}, nil
}

// --- Summarizer Handler ---
type SummarizerHandler struct {
	summarizeAgent *SummarizeAgent // Embed or reference the agent logic
	llmProvider    llm.ModelProvider
}

func NewSummarizerHandler(provider llm.ModelProvider) *SummarizerHandler {
	return &SummarizerHandler{
		summarizeAgent: NewSummarizeAgent(provider), // Create the agent logic instance
		llmProvider:    provider,
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

	return agentflow.AgentResult{OutputState: outputState}, nil
}
