package main

import (
	"context"
	"fmt"
	"log"

	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/tools"
)

// --- Planner Agent ---
type PlannerAgent struct {
	provider llm.ModelProvider
}

func NewPlannerAgent(provider llm.ModelProvider) *PlannerAgent {
	return &PlannerAgent{provider: provider}
}

// Plan generates a research plan based on the user request.
func (a *PlannerAgent) Plan(ctx context.Context, userRequest string) (string, error) {
	log.Println("PlannerAgent: Generating plan...")
	prompt := llm.Prompt{
		System: "You are a planning assistant. Given a user request, create a concise, step-by-step research plan focusing on keywords and questions for a web search.",
		User:   fmt.Sprintf("User Request: %s\n\nResearch Plan:", userRequest),
	}

	// Pass context to LLM call
	resp, err := a.provider.Call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("planner LLM call failed: %w", err)
	}
	log.Println("PlannerAgent: Plan generated successfully.")
	return resp.Content, nil
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

// Research executes the plan using available tools (e.g., web search).
func (a *ResearchAgent) Research(ctx context.Context, plan string) (string, error) {
	log.Println("ResearchAgent: Executing research plan...")
	// In a real scenario, this would parse the plan and make multiple tool calls.
	// For simplicity, we'll use the plan directly as a search query.
	toolName := "web_search" // Assuming the registered tool name
	args := map[string]any{
		"query": plan, // Use the plan (or parts of it) as the query
	}

	log.Printf("ResearchAgent: Calling tool '%s'...", toolName)
	// Pass context to tool call
	result, err := a.registry.CallTool(ctx, toolName, args)
	if err != nil {
		return "", fmt.Errorf("research tool call failed: %w", err)
	}

	log.Println("ResearchAgent: Research tool executed successfully.")
	// Return the result, converting it to string
	return fmt.Sprintf("%v", result), nil
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

// Summarize combines the research results into a final answer.
func (a *SummarizeAgent) Summarize(ctx context.Context, originalRequest, researchResult string) (string, error) {
	log.Println("SummarizeAgent: Generating summary...")
	prompt := llm.Prompt{
		System: "You are a summarization assistant. Based on the original request and the provided research results, create a comprehensive summary.",
		User:   fmt.Sprintf("Original Request: %s\n\nResearch Results:\n%s\n\nSummary:", originalRequest, researchResult),
	}

	// Pass context to LLM call
	resp, err := a.provider.Call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("summarizer LLM call failed: %w", err)
	}
	log.Println("SummarizeAgent: Summary generated successfully.")
	return resp.Content, nil
}
