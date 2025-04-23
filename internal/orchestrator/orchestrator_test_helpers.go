package orchestrator

import (
	"errors"
	agentflow "kunalkushwaha/agentflow/internal/core"
	"sync"
	"time"
)

// SpyEventHandler tracks handled events per handler.
// Used across orchestrator tests.
type SpyEventHandler struct {
	mu     sync.Mutex
	events []string
	failOn string // Event ID to fail on
}

// Implement the agentflow.EventHandler interface
func (s *SpyEventHandler) Handle(e agentflow.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := e.GetID()
	s.events = append(s.events, id)
	if s.failOn == id {
		return errors.New("simulated handler failure")
	}
	return nil
}

func (s *SpyEventHandler) EventCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.events)
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
