package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	agentflow "kunalkushwaha/agentflow/internal/core"
	"kunalkushwaha/agentflow/internal/llm"
	"kunalkushwaha/agentflow/internal/orchestrator"
)

// --- Agent Handler Adapter ---
type AgentHandler struct {
	agentName  string
	agent      agentflow.Agent
	runner     *agentflow.Runner
	provider   llm.ModelProvider
	agentNames []string
	resultChan chan agentflow.State
	wg         *sync.WaitGroup
}

func NewAgentHandler(name string, agent agentflow.Agent, runner *agentflow.Runner, provider llm.ModelProvider, allAgentNames []string, resultChan chan agentflow.State, wg *sync.WaitGroup) *AgentHandler {
	if provider == nil && name != SummarizerAgentName {
		log.Printf("[%s Handler] Warning: LLM Provider is nil. Routing decisions will not use LLM.", name)
	}
	if resultChan == nil {
		log.Fatalf("[%s Handler] Error: Result channel cannot be nil.", name)
	}
	return &AgentHandler{
		agentName:  name,
		agent:      agent,
		runner:     runner,
		provider:   provider,
		agentNames: allAgentNames,
		resultChan: resultChan,
		wg:         wg,
	}
}

func (h *AgentHandler) Handle(event agentflow.Event) error {
	defer h.wg.Done()

	log.Printf("[%s Handler] Handling event %s", h.agentName, event.GetID())

	// --- Convert event payload/metadata to initial agent state ---
	initialState := agentflow.NewState()
	originalEventID := ""
	eventMeta := event.GetMetadata()
	if id, ok := eventMeta["original_event_id"]; ok {
		originalEventID = id
	}

	if payload, ok := event.GetPayload().(map[string]interface{}); ok {
		for k, v := range payload {
			initialState.Set(k, v)
		}
	} else if state, ok := event.GetPayload().(agentflow.State); ok {
		initialState = state.Clone()
		stateMeta := initialState.GetMetadata()
		if id, ok := stateMeta["original_event_id"]; ok {
			originalEventID = id
		}
	} else {
		log.Printf("[%s Handler] Warning: Unexpected payload type %T for event %s", h.agentName, event.GetPayload(), event.GetID())
		initialState.Set("payload", event.GetPayload())
	}

	for k, v := range eventMeta {
		initialState.SetMeta(k, v)
	}

	if originalEventID == "" {
		originalEventID = event.GetID()
	}
	initialState.SetMeta("original_event_id", originalEventID)

	// --- Run the Agent ---
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	finalState, err := h.agent.Run(ctx, initialState)
	if err != nil {
		log.Printf("[%s Handler] Agent run returned an error for event %s: %v", h.agentName, event.GetID(), err)
		errorState := agentflow.NewStateWithData(map[string]any{"handler_error": fmt.Sprintf("Agent %s failed: %v", h.agentName, err)})
		if req, ok := initialState.Get("user_request"); ok {
			errorState.Set("user_request", req)
		}
		h.sendResult(originalEventID, errorState)
		return fmt.Errorf("agent run failed: %w", err)
	}

	log.Printf("[%s Handler] Agent run successful for event %s.", h.agentName, event.GetID())

	// Check if the agent *internally* recorded an error in the state
	if agentErr, ok := finalState.Get("agent_error"); ok {
		log.Printf("[%s Handler] Agent %s recorded an error in its state: %v. Ending flow.", h.agentName, h.agentName, agentErr)
		h.sendResult(originalEventID, finalState)
		return nil
	}

	// --- LLM-Based Routing Decision ---
	if _, ok := finalState.Get("user_request"); !ok {
		if req, ok := initialState.Get("user_request"); ok {
			finalState.Set("user_request", req)
		}
	}

	// Call the router function (defined in router.go)
	nextAgent, err := determineNextAgentViaLLM(ctx, h.provider, finalState, h.agentName, h.agentNames)
	if err != nil {
		log.Printf("[%s Handler] Error determining next agent via LLM: %v. Ending flow.", h.agentName, err)
		h.sendResult(originalEventID, finalState)
		return nil
	}

	if nextAgent != "" && nextAgent != "DONE" {
		// --- Emit Next Event ---
		log.Printf("[%s Handler] LLM chose '%s' as the next agent.", h.agentName, nextAgent)
		finalState.SetMeta(orchestrator.RouteMetadataKey, nextAgent)

		nextEvent := &agentflow.SimpleEvent{
			ID:       fmt.Sprintf("%s-next-%s", event.GetID(), nextAgent),
			Payload:  finalState,
			Metadata: finalState.GetMetadata(),
		}
		log.Printf("[%s Handler] Emitting next event %s for agent %s", h.agentName, nextEvent.GetID(), nextAgent)
		h.wg.Add(1)
		h.runner.Emit(nextEvent)
	} else {
		// --- Final Step Reached ---
		log.Printf("[%s Handler] LLM decided flow is DONE after this agent.", h.agentName)
		h.sendResult(originalEventID, finalState)
	}

	return nil
}

// Helper function to send results to the SHARED channel
func (h *AgentHandler) sendResult(originalEventID string, state agentflow.State) {
	log.Printf("[%s Handler] Sending final state to shared results channel for %s.", h.agentName, originalEventID)
	select {
	case h.resultChan <- state:
	// OK
	default:
		log.Printf("[%s Handler] Warning: Shared results channel was blocked or closed when trying to send state for %s.", h.agentName, originalEventID)
	}
}
