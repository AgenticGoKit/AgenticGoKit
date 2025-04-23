package agentflow

import (
	"sync"
)

// EventHandler processes events. (Renamed from Agent)
type EventHandler interface {
	Handle(Event) error
}

// Runner accepts events and uses an Orchestrator to dispatch them.
type Runner struct {
	orchestrator Orchestrator // Orchestrator now works with EventHandler
	queue        chan Event
	wgServe      sync.WaitGroup // waits for loop to finish
	wgEmit       sync.WaitGroup // waits for all Emit calls
}

// NewRunner creates a Runner with a specific Orchestrator and queue size.
func NewRunner(orchestrator Orchestrator, buffer int) *Runner {
	if orchestrator == nil {
		// Default or panic, depending on design choice. Let's panic for now.
		panic("Runner requires a non-nil Orchestrator")
	}
	r := &Runner{
		orchestrator: orchestrator,
		queue:        make(chan Event, buffer),
	}
	r.wgServe.Add(1)
	go r.loop()
	return r
}

func (r *Runner) loop() {
	defer r.wgServe.Done()
	for ev := range r.queue {
		r.orchestrator.Dispatch(ev) // Delegate dispatch to the orchestrator
	}
}

// RegisterAgent registers an EventHandler with the underlying Orchestrator.
func (r *Runner) RegisterAgent(name string, agent EventHandler) {
	r.orchestrator.RegisterAgent(name, agent) // Delegate registration
}

// Emit enqueues an Event for delivery.
func (r *Runner) Emit(event Event) {
	r.wgEmit.Add(1)
	defer r.wgEmit.Done()
	// Add safety check in case Stop is called concurrently
	// This requires a more complex shutdown signal, or we accept potential panic
	// For simplicity now, we assume Emit won't be called after Stop starts.
	r.queue <- event
}

// Stop waits for emits, closes queue, waits for loop, and stops orchestrator.
func (r *Runner) Stop() {
	r.wgEmit.Wait()
	close(r.queue)
	r.wgServe.Wait()
	r.orchestrator.Stop() // Allow orchestrator cleanup
}

// Note: You will also need to update the Orchestrator interface and its
// implementations (RouteOrchestrator, CollaborateOrchestrator, and mocks)
// to accept and use EventHandler instead of the old Agent type.
