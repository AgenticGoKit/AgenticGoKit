package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/tools"
)

// --- Planner Handler ---
type PlannerHandler struct {
	llm llm.ModelProvider
}

func NewPlannerHandler(provider llm.ModelProvider) *PlannerHandler {
	return &PlannerHandler{llm: provider}
}

// Run implements the agentflow.AgentHandler interface
func (h *PlannerHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	userRequest, ok := state.Get("user_request")
	if !ok || userRequest == nil {
		return agentflow.AgentResult{}, fmt.Errorf("missing user_request in state")
	}

	requestStr := fmt.Sprintf("%v", userRequest)

	// Create proper prompt following Azure AI best practices
	prompt := llm.Prompt{
		System: "You are a planning assistant. Given a user request, create a detailed plan for research.",
		User:   fmt.Sprintf("User Request: %s\n\nResearch Plan:", requestStr),
	}

	// Use the Call method which is defined on ModelProvider
	resp, err := h.llm.Call(ctx, prompt)
	if err != nil {
		return agentflow.AgentResult{}, fmt.Errorf("planner LLM call failed: %w", err)
	}

	// Create a new state with the planning output
	outputState := agentflow.NewState()
	outputState.Set("plan", resp.Content)
	outputState.Set("user_request", requestStr) // Pass through the original request

	// Route to the researcher agent next
	outputState.SetMeta(agentflow.RouteMetadataKey, ResearcherAgentName)

	return agentflow.AgentResult{OutputState: outputState}, nil
}

// --- Researcher Handler ---
type ResearcherHandler struct {
	tools tools.ToolRegistry
	llm   llm.LLMAdapter
}

func NewResearcherHandler(tools tools.ToolRegistry, llm llm.LLMAdapter) *ResearcherHandler {
	return &ResearcherHandler{
		tools: tools,
		llm:   llm,
	}
}

// Run implements the agentflow.AgentHandler interface
func (h *ResearcherHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	plan, ok := state.Get("plan")
	if !ok || plan == nil {
		return agentflow.AgentResult{}, fmt.Errorf("missing plan in state")
	}

	// For simplicity, simulate research result
	// In a real implementation, would use the web search tool
	researchResult := fmt.Sprintf("Research results based on plan: %v\n\n"+
		"1. Recent developments in AI code generation include GitHub Copilot X and improvements in multi-language support.\n"+
		"2. New approaches to generating test suites automatically have shown promise in industry studies.\n"+
		"3. Several academic papers have demonstrated improved code quality with LLM assistance.",
		plan)

	// Create new state with research results
	outputState := agentflow.NewState()

	// Copy existing data using Keys() and Get() instead of GetData()
	for _, k := range state.Keys() {
		if v, ok := state.Get(k); ok {
			outputState.Set(k, v)
		}
	}

	outputState.Set("research_result", researchResult)
	outputState.Set("research_completed", true)

	// Route to the summarizer next
	outputState.SetMeta(agentflow.RouteMetadataKey, SummarizerAgentName)

	return agentflow.AgentResult{OutputState: outputState}, nil
}

// --- Summarizer Handler ---
type SummarizerHandler struct {
	azure llm.ModelProvider
	llm   llm.LLMAdapter
}

func NewSummarizerHandler(azure llm.ModelProvider, llm llm.LLMAdapter) *SummarizerHandler {
	return &SummarizerHandler{
		azure: azure,
		llm:   llm,
	}
}

// Run implements the agentflow.AgentHandler interface
func (h *SummarizerHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	researchResult, ok := state.Get("research_result")
	if !ok || researchResult == nil {
		return agentflow.AgentResult{}, fmt.Errorf("missing research_result in state")
	}

	userRequest, _ := state.Get("user_request")
	userRequestStr := ""
	if userRequest != nil {
		userRequestStr = fmt.Sprintf("%v", userRequest)
	}

	// Properly structure prompt following Azure AI best practices
	prompt := llm.Prompt{
		System: "You are a summarization assistant. Based on the research results, create a comprehensive summary.",
		User:   fmt.Sprintf("Original Request: %s\n\nResearch Results:\n%v\n\nSummary:", userRequestStr, researchResult),
	}

	// Use the Call method which is defined on ModelProvider
	resp, err := h.azure.Call(ctx, prompt)
	if err != nil {
		return agentflow.AgentResult{}, fmt.Errorf("summarizer LLM call failed: %w", err)
	}

	// Create new state with summary
	outputState := agentflow.NewState()

	// Copy existing data using Keys() and Get()
	for _, k := range state.Keys() {
		if v, ok := state.Get(k); ok {
			outputState.Set(k, v)
		}
	}

	outputState.Set("summary", resp.Content)

	// Route to final output
	outputState.SetMeta(agentflow.RouteMetadataKey, FinalOutputAgent)

	return agentflow.AgentResult{OutputState: outputState}, nil
}

// --- Final Output Handler ---
type FinalOutputHandler struct {
	wg           *sync.WaitGroup
	mu           sync.Mutex
	finalResults map[string]agentflow.State
}

func NewFinalOutputHandler(wg *sync.WaitGroup) *FinalOutputHandler {
	return &FinalOutputHandler{
		wg:           wg,
		finalResults: make(map[string]agentflow.State),
	}
}

// Run implements the agentflow.AgentHandler interface
func (h *FinalOutputHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	sessionID := event.GetSessionID()

	log.Printf("Final output processing for session: %s", sessionID)

	// Store the final state
	h.mu.Lock()
	h.finalResults[sessionID] = state
	h.mu.Unlock()

	// Signal that we're done with this flow
	h.wg.Done()

	// Return an empty result
	return agentflow.AgentResult{OutputState: state}, nil
}

func (h *FinalOutputHandler) GetFinalState(sessionID string) (agentflow.State, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	state, found := h.finalResults[sessionID]
	return state, found
}
