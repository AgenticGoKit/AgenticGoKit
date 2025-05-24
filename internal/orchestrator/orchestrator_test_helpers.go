package orchestrator

import (
	"context" // <<< Add context import
	"fmt"
	"sync"
	"time"

	agentflow "github.com/kunalkushwaha/agentflow/internal/core"
)

// MockOrchestrator for testing handler interactions
type MockOrchestrator struct {
	eventHandlers map[string]agentflow.EventHandler // Renamed for clarity
	agentHandlers map[string]agentflow.AgentHandler // Added for AgentHandler registration
	registry      *agentflow.CallbackRegistry       // Add registry if needed
	mu            sync.Mutex
}

// NewMockOrchestrator creates a mock orchestrator
func NewMockOrchestrator(registry *agentflow.CallbackRegistry) *MockOrchestrator {
	return &MockOrchestrator{
		eventHandlers: make(map[string]agentflow.EventHandler), // Initialize map
		agentHandlers: make(map[string]agentflow.AgentHandler), // Initialize map
		registry:      registry,
	}
}

// RegisterAgent adds a handler - Now handles both types based on interface assertion
func (m *MockOrchestrator) RegisterAgent(agentID string, handler interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if it's an AgentHandler
	if ah, ok := handler.(agentflow.AgentHandler); ok {
		m.agentHandlers[agentID] = ah
		return nil
	}
	// Check if it's an EventHandler
	if eh, ok := handler.(agentflow.EventHandler); ok {
		m.eventHandlers[agentID] = eh
		return nil
	}

	return fmt.Errorf("handler type not supported by mock orchestrator: %T", handler)
}

// Dispatch simulates routing based on TargetAgentID (like RouteOrchestrator)
func (m *MockOrchestrator) Dispatch(event agentflow.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if event == nil {
		return fmt.Errorf("mock orchestrator cannot handle nil event")
	}
	targetID := event.GetTargetAgentID()
	if targetID == "" {
		metadata := event.GetMetadata()
		routeKey := "route"
		targetID = metadata[routeKey]
		if targetID == "" {
			return fmt.Errorf("event has no target agent ID or route metadata")
		}
	}

	// Check AgentHandlers first
	if handler, ok := m.agentHandlers[targetID]; ok {
		if m.registry == nil {
			return fmt.Errorf("mock orchestrator registry is nil, cannot call AgentHandler")
		}
		// FIX: Call Run with correct signature and handle return values
		// Pass background context and new empty state for the mock call
		ctx := context.Background()
		state := agentflow.NewState()
		_, err := handler.Run(ctx, event, state) // Call Run, ignore result for mock Dispatch
		return err                               // Return only the error, matching mock Dispatch signature
	}

	// Check EventHandlers if no AgentHandler found
	if handler, ok := m.eventHandlers[targetID]; ok {
		return handler.Handle(event) // EventHandler signature
	}

	return fmt.Errorf("no handler registered for agent %s", targetID)
}

// Stop is a placeholder
func (m *MockOrchestrator) Stop() {}

// GetCallbackRegistry returns the stored registry
func (m *MockOrchestrator) GetCallbackRegistry() *agentflow.CallbackRegistry {
	return m.registry
}

// SpyEventHandler is a mock EventHandler for testing orchestrators.
type SpyEventHandler struct {
	AgentName    string
	HandleCalled bool
	LastEvent    agentflow.Event
	ReturnError  error
	mu           sync.Mutex
	eventCount   int
	events       []string
	failOn       string
}

// Handle implements the agentflow.EventHandler interface.
func (h *SpyEventHandler) Handle(event agentflow.Event) error { // Correct signature for EventHandler
	h.mu.Lock()
	defer h.mu.Unlock()

	h.HandleCalled = true
	h.LastEvent = event
	h.eventCount++
	if event != nil {
		// FIX: Use GetID() method
		eventID := event.GetID()
		h.events = append(h.events, eventID)
		if h.failOn != "" && eventID == h.failOn {
			err := h.ReturnError
			if err == nil {
				err = fmt.Errorf("handler '%s' failed deliberately for event '%s'", h.AgentName, eventID)
			}
			return err
		}
	}
	return h.ReturnError
}

// EventCount returns the number of events handled (thread-safe).
func (h *SpyEventHandler) EventCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.eventCount
}

// GetEvents returns a copy of the handled event IDs (thread-safe).
func (h *SpyEventHandler) GetEvents() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Return a copy to prevent external modification
	evts := make([]string, len(h.events))
	copy(evts, h.events)
	return evts
}

// SlowSpyEventHandler simulates a handler that takes time to process.
type SlowSpyEventHandler struct {
	SpyEventHandler // Embed SpyEventHandler
	delay           time.Duration
}

func (s *SlowSpyEventHandler) Handle(e agentflow.Event) error {
	time.Sleep(s.delay)
	return s.SpyEventHandler.Handle(e) // Call embedded Handle
}

// --- SpyAgentHandler (implements AgentHandler) ---
type SpyAgentHandler struct {
	AgentName    string
	RunCalled    bool // Renamed from HandleCalled
	LastEvent    agentflow.Event
	LastRegistry *agentflow.CallbackRegistry
	LastState    agentflow.State // Store the received state
	LastContext  context.Context // Store the received context
	ReturnError  error
	ReturnResult agentflow.AgentResult // Define what result to return
	mu           sync.Mutex
	eventCount   int
	events       []string // Store event IDs
	failOn       string   // Event ID to deliberately fail on
}

// FIX: Implement the Run method required by the AgentHandler interface.
// Run implements the agentflow.AgentHandler interface.
func (h *SpyAgentHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.RunCalled = true // Use RunCalled
	h.LastEvent = event
	h.LastState = state // Store state
	h.LastContext = ctx // Store context
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
			// Return default/empty result along with the error
			return agentflow.AgentResult{}, err
		}
	}
	// Return the configured result and error
	return h.ReturnResult, h.ReturnError
}

// EventCount remains the same
func (h *SpyAgentHandler) EventCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.eventCount
}

// GetEvents remains the same
func (h *SpyAgentHandler) GetEvents() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	evts := make([]string, len(h.events))
	copy(evts, h.events)
	return evts
}
