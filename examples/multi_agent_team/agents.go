package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/tools"
)

// --- Planner Agent ---
type PlannerAgent struct{}

func NewPlannerAgent() *PlannerAgent {
	return &PlannerAgent{}
}

func (a *PlannerAgent) Run(ctx context.Context, in agentflow.State) (agentflow.State, error) {
	log.Printf("[%s] Running...", PlannerAgentName)
	requestVal, ok := in.Get("user_request")
	if !ok {
		return in, fmt.Errorf("[%s] 'user_request' not found", PlannerAgentName)
	}
	request, _ := requestVal.(string)
	log.Printf("[%s] Received request: %q", PlannerAgentName, request)

	// Simple rule: Extract topic between "Research" and "and summarize"
	topic := ""
	if strings.Contains(strings.ToLower(request), "research") && strings.Contains(strings.ToLower(request), "summarize") {
		startIdx := strings.Index(strings.ToLower(request), "research") + len("research")
		endIdx := strings.Index(strings.ToLower(request), "and summarize")
		if startIdx < endIdx {
			topic = strings.TrimSpace(request[startIdx:endIdx])
		}
	}

	if topic == "" {
		log.Printf("[%s] Could not determine research topic from request.", PlannerAgentName)
		out := in.Clone()
		out.Set("agent_error", "Could not determine research topic.")
		// Ensure original request is passed along even on error
		if req, ok := in.Get("user_request"); ok {
			out.Set("user_request", req)
		}
		metadata := in.GetMetadata()
		if originalID, ok := metadata["original_event_id"]; ok {
			out.SetMeta("original_event_id", originalID)
		}
		return out, nil
	}

	log.Printf("[%s] Determined research topic: %q", PlannerAgentName, topic)

	out := agentflow.NewState()
	out.Set("research_topic", topic)
	if req, ok := in.Get("user_request"); ok {
		out.Set("user_request", req)
	}
	metadata := in.GetMetadata()
	if originalID, ok := metadata["original_event_id"]; ok {
		out.SetMeta("original_event_id", originalID)
	}
	log.Printf("[%s] Planning complete. State prepared.", PlannerAgentName)
	return out, nil
}

// --- Research Agent ---
type ResearchAgent struct {
	registry *tools.ToolRegistry
}

func NewResearchAgent(registry *tools.ToolRegistry) *ResearchAgent {
	if registry == nil {
		log.Fatalf("[%s] Requires a non-nil ToolRegistry", ResearcherAgentName)
	}
	return &ResearchAgent{registry: registry}
}

func (a *ResearchAgent) Run(ctx context.Context, in agentflow.State) (agentflow.State, error) {
	log.Printf("[%s] Running...", ResearcherAgentName)
	topicVal, ok := in.Get("research_topic")
	if !ok {
		return in, fmt.Errorf("[%s] 'research_topic' not found", ResearcherAgentName)
	}
	topic, _ := topicVal.(string)
	log.Printf("[%s] Received topic: %q", ResearcherAgentName, topic)

	toolName := "web_search"
	toolArgs := map[string]any{"query": topic}

	log.Printf("[%s] Calling tool '%s'...", ResearcherAgentName, toolName)
	toolCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	toolResult, err := a.registry.CallTool(toolCtx, toolName, toolArgs)
	if err != nil {
		log.Printf("[%s] Tool call failed: %v", ResearcherAgentName, err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("Tool call failed: %v", err))
		if req, ok := in.Get("user_request"); ok {
			out.Set("user_request", req)
		}
		metadata := in.GetMetadata()
		if originalID, ok := metadata["original_event_id"]; ok {
			out.SetMeta("original_event_id", originalID)
		}
		return out, nil
	}

	rawResults := "No results found."
	if resultsVal, ok := toolResult["results"]; ok {
		if resultsSlice, ok := resultsVal.([]string); ok {
			if len(resultsSlice) > 0 {
				rawResults = strings.Join(resultsSlice, "\n\n")
			}
		} else if resultsStr, ok := resultsVal.(string); ok {
			rawResults = resultsStr
		} else {
			log.Printf("[%s] Tool result for 'results' key is neither []string nor string, it's a %T", ResearcherAgentName, resultsVal)
		}
	}
	log.Printf("[%s] Research complete. Found %d bytes of results.", ResearcherAgentName, len(rawResults))
	log.Printf("[%s] Raw results from tool:\n---\n%s\n---", ResearcherAgentName, rawResults)

	out := agentflow.NewState()
	out.Set("raw_results", rawResults)
	if req, ok := in.Get("user_request"); ok {
		out.Set("user_request", req)
	}
	metadata := in.GetMetadata()
	if originalID, ok := metadata["original_event_id"]; ok {
		out.SetMeta("original_event_id", originalID)
	}
	log.Printf("[%s] Research complete. State prepared.", ResearcherAgentName)
	return out, nil
}

// --- Summarize Agent ---
type SummarizeAgent struct {
	provider llm.ModelProvider
}

func NewSummarizeAgent(provider llm.ModelProvider) *SummarizeAgent {
	if provider == nil {
		log.Fatalf("[%s] Requires a non-nil ModelProvider", SummarizerAgentName)
	}
	return &SummarizeAgent{provider: provider}
}

func (a *SummarizeAgent) Run(ctx context.Context, in agentflow.State) (agentflow.State, error) {
	log.Printf("[%s] Running...", SummarizerAgentName)
	resultsVal, ok := in.Get("raw_results")
	if !ok {
		return in, fmt.Errorf("[%s] 'raw_results' not found", SummarizerAgentName)
	}
	rawResults, _ := resultsVal.(string)
	log.Printf("[%s] Received %d bytes of raw results.", SummarizerAgentName, len(rawResults))

	maxLen := 8000
	if len(rawResults) > maxLen {
		rawResults = rawResults[:maxLen] + "\n... [truncated]"
	}

	prompt := llm.Prompt{
		System: "You are an expert summarizer. Summarize the following text concisely.",
		User:   rawResults,
	}

	log.Printf("[%s] Calling LLM for summarization...", SummarizerAgentName)
	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	resp, err := a.provider.Call(llmCtx, prompt)
	if err != nil {
		log.Printf("[%s] LLM call failed: %v", SummarizerAgentName, err)
		out := in.Clone()
		out.Set("agent_error", fmt.Sprintf("LLM call failed: %v", err))
		if req, ok := in.Get("user_request"); ok {
			out.Set("user_request", req)
		}
		metadata := in.GetMetadata()
		if originalID, ok := metadata["original_event_id"]; ok {
			out.SetMeta("original_event_id", originalID)
		}
		return out, nil
	}

	log.Printf("[%s] Summarization complete.", SummarizerAgentName)

	out := agentflow.NewState()
	out.Set("summary", resp.Content)
	if req, ok := in.Get("user_request"); ok {
		out.Set("user_request", req)
	}
	metadata := in.GetMetadata()
	if originalID, ok := metadata["original_event_id"]; ok {
		out.SetMeta("original_event_id", originalID)
	}
	return out, nil
}
