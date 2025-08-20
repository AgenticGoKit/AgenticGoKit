package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// MockOrchestrator for testing handler interactions
type MockOrchestrator struct {
	agentHandlers map[string]core.AgentHandler // AgentHandler registration
	registry      *core.CallbackRegistry       // Add registry if needed
	mu            sync.Mutex
}

// NewMockOrchestrator creates a mock orchestrator
func NewMockOrchestrator(registry *core.CallbackRegistry) *MockOrchestrator {
	return &MockOrchestrator{
		agentHandlers: make(map[string]core.AgentHandler), // Initialize map
		registry:      registry,
	}
}

// RegisterAgent adds a handler
func (m *MockOrchestrator) RegisterAgent(agentID string, handler core.AgentHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	m.agentHandlers[agentID] = handler
	return nil
}

// Dispatch simulates routing based on TargetAgentID (like RouteOrchestrator)
func (m *MockOrchestrator) Dispatch(ctx context.Context, event core.Event) (core.AgentResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if event == nil {
		return core.AgentResult{}, fmt.Errorf("mock orchestrator cannot handle nil event")
	}
	
	targetID, ok := event.GetMetadataValue(core.RouteMetadataKey)
	if !ok {
		return core.AgentResult{}, fmt.Errorf("event has no route metadata")
	}

	// Check AgentHandlers
	if handler, ok := m.agentHandlers[targetID]; ok {
		// Call Run with correct signature and handle return values
		state := core.NewState()
		result, err := handler.Run(ctx, event, state)
		return result, err
	}

	return core.AgentResult{}, fmt.Errorf("no handler registered for agent %s", targetID)
}

// Stop is a placeholder
func (m *MockOrchestrator) Stop() {}

// GetCallbackRegistry returns the stored registry
func (m *MockOrchestrator) GetCallbackRegistry() *core.CallbackRegistry {
	return m.registry
}



// SpyAgentHandler implements AgentHandler for testing
type SpyAgentHandler struct {
	AgentName    string
	RunCalled    bool
	LastEvent    core.Event
	LastRegistry *core.CallbackRegistry
	LastState    core.State
	LastContext  context.Context
	ReturnError  error
	ReturnResult core.AgentResult
	mu           sync.Mutex
	eventCount   int
	events       []string
	failOn       string
}

// Run implements the core.AgentHandler interface.
func (h *SpyAgentHandler) Run(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.RunCalled = true
	h.LastEvent = event
	h.LastState = state
	h.LastContext = ctx
	h.eventCount++
	var eventID string
	if event != nil {
		eventID = event.GetID()
		h.events = append(h.events, eventID)
		if h.failOn != "" && eventID == h.failOn {
			err := h.ReturnError
			if err == nil {
				err = fmt.Errorf("handler '%s' failed deliberately for event '%s'", h.AgentName, eventID)
			}
			return core.AgentResult{OutputState: state}, err
		}
	}
	
	// Return the configured result and error, ensuring OutputState is not nil
	result := h.ReturnResult
	if result.OutputState == nil {
		result.OutputState = state
	}
	return result, h.ReturnError
}

// EventCount returns the number of events handled (thread-safe).
func (h *SpyAgentHandler) EventCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.eventCount
}

// GetEvents returns a copy of the handled event IDs (thread-safe).
func (h *SpyAgentHandler) GetEvents() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	evts := make([]string, len(h.events))
	copy(evts, h.events)
	return evts
}