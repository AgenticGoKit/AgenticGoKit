package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
)

// determineNextAgentViaLLM asks the LLM to choose the next agent based on the state.
func determineNextAgentViaLLM(ctx context.Context, provider llm.ModelProvider, currentState agentflow.State, currentAgentName string, availableAgentNames []string) (string, error) {
	if provider == nil {
		return "", fmt.Errorf("LLM provider is nil, cannot determine next agent via LLM")
	}

	// --- Prepare data for the prompt ---
	originalRequest, _ := currentState.Get("user_request")
	if originalRequest == nil {
		metadata := currentState.GetMetadata()
		originalRequest = metadata["original_event_id"]
	}

	stateSummary := fmt.Sprintf("%+v", currentState.GetData())
	maxLen := 500
	if len(stateSummary) > maxLen {
		stateSummary = stateSummary[:maxLen] + "...}"
	}

	nextAgentChoices := []string{}
	agentDescriptions := map[string]string{
		PlannerAgentName:    "Analyzes the initial request and creates a plan.",
		ResearcherAgentName: "Uses tools (like web search) to find information based on a topic.",
		SummarizerAgentName: "Summarizes provided text using an LLM.",
	}
	descriptionList := ""
	for _, name := range availableAgentNames {
		if name != currentAgentName {
			nextAgentChoices = append(nextAgentChoices, name)
			descriptionList += fmt.Sprintf("- %s: %s\n", name, agentDescriptions[name])
		}
	}
	if len(nextAgentChoices) == 0 {
		log.Printf("[Router LLM] No other agents available after %s.", currentAgentName)
		return "DONE", nil
	}

	// --- Construct the prompt ---
	systemPrompt := fmt.Sprintf(`You are a workflow routing assistant. Your goal is to decide the next logical step in a process based on the original request and the current state.

Available agents for the next step:
%s
Analyze the original request and the current state data. Choose the single best agent from the list above to run next.
If the original request seems fulfilled by the current state, or if no available agent is suitable for the next step, respond with the exact word "DONE".
Respond ONLY with the chosen agent's name (e.g., "%s") or the word "DONE".`, descriptionList, nextAgentChoices[0])

	userPrompt := fmt.Sprintf(`Original Request: %v
Current State Data: %s

Which agent should run next, or is the process DONE?`, originalRequest, stateSummary)

	prompt := llm.Prompt{
		System: systemPrompt,
		User:   userPrompt,
	}

	// --- Call LLM ---
	log.Printf("[%s Handler] Asking LLM router which agent is next...", currentAgentName) // Log still references handler context
	llmCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := provider.Call(llmCtx, prompt)
	if err != nil {
		return "", fmt.Errorf("LLM routing call failed: %w", err)
	}

	decision := strings.TrimSpace(resp.Content)
	log.Printf("[%s Handler] LLM router decision: %q", currentAgentName, decision) // Log still references handler context

	// --- Validate decision ---
	if decision == "DONE" {
		return "DONE", nil
	}
	for _, validName := range nextAgentChoices {
		if decision == validName {
			return decision, nil
		}
	}

	log.Printf("[%s Handler] LLM router returned an invalid choice: %q. Treating as DONE.", currentAgentName, decision) // Log still references handler context
	return "DONE", nil
}
