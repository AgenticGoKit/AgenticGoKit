package agentflow

import (
	"fmt"
	"sync"
	"testing"
)

// MockOrchestrator is a simple mock for testing the Runner.
type MockOrchestrator struct {
	mu       sync.Mutex
	handlers map[string]EventHandler // Use EventHandler, renamed field
	events   []Event
}

func NewMockOrchestrator() *MockOrchestrator {
	return &MockOrchestrator{
		handlers: make(map[string]EventHandler), // Use EventHandler
		events:   make([]Event, 0),
	}
}

// RegisterAgent now accepts EventHandler to match the interface
func (m *MockOrchestrator) RegisterAgent(name string, handler EventHandler) { // Use EventHandler
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[name] = handler // Use renamed field
}

func (m *MockOrchestrator) Dispatch(event Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	// You might want to dispatch to mocked handlers here if needed for specific tests
}

func (m *MockOrchestrator) Stop() {
	// No-op
}

// SpyEventHandler implements EventHandler for testing orchestrators/runner
// Renamed from SpyAgent to avoid confusion with the new Agent interface
type SpyEventHandler struct {
	mu     sync.Mutex
	events []string
	failOn string // Event ID to fail on
}

// Handle implements the EventHandler interface
func (s *SpyEventHandler) Handle(e Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := e.GetID()
	s.events = append(s.events, id)
	if s.failOn == id {
		return fmt.Errorf("simulated handler failure on %s", id)
	}
	return nil
}

func (s *SpyEventHandler) EventCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.events)
}

// TestRunnerStrictFIFO verifies that events emitted sequentially are processed in FIFO order.
func TestRunnerStrictFIFO(t *testing.T) {
	const N = 10
	mockOrchestrator := NewMockOrchestrator()
	r := NewRunner(mockOrchestrator, N) // This should now compile

	// We don't register handlers on the mock for this specific test,
	// as we only care that the runner passes events to the orchestrator's Dispatch.

	for i := 0; i < N; i++ {
		r.Emit(&SimpleEvent{ID: fmt.Sprintf("evt-%d", i)})
	}
	r.Stop()

	// Verification: Check events captured by the mock orchestrator
	mockOrchestrator.mu.Lock()
	defer mockOrchestrator.mu.Unlock()
	if len(mockOrchestrator.events) != N {
		t.Fatalf("want %d events dispatched, got %d", N, len(mockOrchestrator.events))
	}
	for i := 0; i < N; i++ {
		want := fmt.Sprintf("evt-%d", i)
		got := mockOrchestrator.events[i].GetID()
		if got != want {
			t.Errorf("at index %d: want %q, got %q", i, want, got)
		}
	}
}

// TestRunnerConcurrentSafety verifies that the runner can handle many concurrent emits without losing events.
func TestRunnerConcurrentSafety(t *testing.T) {
	mockOrchestrator := NewMockOrchestrator()
	r := NewRunner(mockOrchestrator, 100) // This should now compile

	const M = 1000
	var wg sync.WaitGroup
	wg.Add(M)
	for i := 0; i < M; i++ {
		go func(i int) {
			defer wg.Done()
			r.Emit(&SimpleEvent{ID: fmt.Sprintf("x-%d", i)})
		}(i)
	}
	wg.Wait()
	r.Stop()

	// Verification: Check total events captured by the mock orchestrator
	mockOrchestrator.mu.Lock()
	defer mockOrchestrator.mu.Unlock()
	if len(mockOrchestrator.events) != M {
		t.Errorf("lost events: want %d dispatched, got %d", M, len(mockOrchestrator.events))
	}
}
