package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

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
	return &ResearchAgent{
		registry: registry,
	}
}

// Research executes the plan using available tools (e.g., web search).
func (a *ResearchAgent) Research(ctx context.Context, plan string, userRequestContext string) (string, error) {
	log.Println("ResearchAgent: Executing research plan...")

	// Create a context with timeout specifically for Azure services
	// Azure best practice: Use appropriate timeouts for service calls
	// - Remove unused searchCtx variable by using the parent ctx directly
	// - Keep the cancel function to ensure proper resource cleanup
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Extract search queries from the plan
	queries := extractSearchQueriesFromPlan(plan)
	if len(queries) == 0 {
		queries = []string{plan} // Fallback to using the entire plan as a query
	}

	var results strings.Builder
	results.WriteString("# Research Results\n\n")
	results.WriteString(fmt.Sprintf("Plan: %s\n\n", plan))

	// Azure best practice: Add telemetry for monitoring
	startTime := time.Now()

	// Execute web searches for each query with Azure best practices
	for i, query := range queries {
		// Azure best practice: Check for context cancellation between operations
		if ctx.Err() != nil {
			return "", fmt.Errorf("research cancelled: %w", ctx.Err())
		}

		// Azure best practice: Rate limiting between calls
		if i > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		log.Printf("ResearchAgent: Simulating search for query %d: %s", i+1, query)

		// Simulate processing time
		select {
		case <-time.After(100 * time.Millisecond):
			// Simulation completed normally
		case <-ctx.Done():
			// Azure best practice: Handle context cancellation gracefully
			log.Printf("ResearchAgent: Search for query %d was cancelled", i+1)
			return "", fmt.Errorf("research operation cancelled: %w", ctx.Err())
		}

		// Generate mock results instead of using the actual tool
		mockResults := fmt.Sprintf("Mock search results for: %s\n\n"+
			"- Result 1: Information about %s\n"+
			"- Result 2: Additional details related to this topic\n"+
			"- Result 3: Sample data point for demonstration\n",
			query, query)

		results.WriteString(fmt.Sprintf("## Search Results for: %s\n\n", query))
		results.WriteString(mockResults)
		results.WriteString("\n\n")
	}

	// Azure best practice: Log performance metrics
	elapsed := time.Since(startTime)
	log.Printf("ResearchAgent: Completed %d searches in %.2f seconds", len(queries), elapsed.Seconds())

	return results.String(), nil
}

// Helper to extract search queries from a plan
func extractSearchQueriesFromPlan(plan string) []string {
	// Simple implementation - look for lines that seem to be search queries
	lines := strings.Split(plan, "\n")
	var queries []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines that start with numbers, bullets, or keywords
		if matched, _ := regexp.MatchString(`^(\d+\.|\-|\*|Search|Query|Find|Look up|Research)`, line); matched {
			// Extract the actual query by removing prefixes
			query := regexp.MustCompile(`^(\d+\.|\-|\*|Search|Query|Find|Look up|Research)(\s+for)?:?\s*`).
				ReplaceAllString(line, "")
			if query != "" {
				queries = append(queries, query)
			}
		}
	}

	return queries
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
