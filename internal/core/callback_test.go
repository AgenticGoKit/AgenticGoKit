package agentflow

import (
	"context" // <<< Add context import
	"sync"
	"sync/atomic"
	"testing"
)

// Helper function for testing callbacks
// FIX: Match CallbackFunc signature (accept *Event)
func createTestCallback(counter *int32, wg *sync.WaitGroup) CallbackFunc {
	return func(ctx context.Context, currentState State, event *Event) (State, error) { // Accept *Event
		if counter != nil {
			atomic.AddInt32(counter, 1)
		}
		if wg != nil {
			wg.Done()
		}
		// Add nil check if accessing event
		// if event != nil { log.Printf("Test callback received event: %s", (*event).GetID()) }
		return nil, nil
	}
}

func TestNewCallbackRegistry(t *testing.T) {
	registry := NewCallbackRegistry()
	if registry == nil {
		t.Fatal("NewCallbackRegistry returned nil")
	}
}

func TestCallbackRegistry_Register_Unregister(t *testing.T) {
	registry := NewCallbackRegistry()
	// Use boolean flags to track calls
	var callback1Called, callback2Called, allHooksCalled bool

	// Reset flags before each invoke section
	resetFlags := func() {
		callback1Called, callback2Called, allHooksCalled = false, false, false
	}

	cb1 := func(ctx context.Context, s State, e *Event) (State, error) {
		callback1Called = true
		if s != nil {
			t.Log("cb1 received state")
		}
		if e != nil {
			t.Logf("cb1 received event ID: %s", (*e).GetID())
		} else {
			t.Log("cb1 received nil event")
		}
		return nil, nil
	}
	cb2 := func(ctx context.Context, s State, e *Event) (State, error) {
		callback2Called = true
		return &SimpleState{data: map[string]interface{}{"cb2_processed": true}}, nil
	}
	// FIX: Change allHooksCalled back to bool
	cbAll := func(ctx context.Context, s State, e *Event) (State, error) {
		allHooksCalled = true
		return nil, nil
	}

	// --- Test Registration ---
	registry.Register(HookBeforeAgentRun, "cb1", cb1)
	registry.Register(HookBeforeAgentRun, "cb2", cb2)
	registry.Register(HookAll, "cbAll", cbAll)

	// Invoke to check registration
	resetFlags() // Reset flags before invoke
	invokeArgs := CallbackArgs{Ctx: context.Background(), Hook: HookBeforeAgentRun}
	registry.Invoke(invokeArgs)

	// Check results
	if !callback1Called {
		t.Errorf("After initial registration, expected callback1 to be called")
	}
	if !callback2Called {
		t.Errorf("After initial registration, expected callback2 to be called")
	}
	if !allHooksCalled { // Check boolean flag
		t.Errorf("After initial registration, expected allHooksCalled to be true")
	}

	// --- Test Unregister Specific ---
	registry.Unregister(HookBeforeAgentRun, "cb1")

	// Invoke again to check effect of unregister
	resetFlags()                // Reset flags before invoke
	registry.Invoke(invokeArgs) // Use same args

	// Check results
	if callback1Called { // Should NOT be called
		t.Errorf("After unregistering 'cb1', expected callback1 NOT to be called")
	}
	if !callback2Called { // Should still be called
		t.Errorf("After unregistering 'cb1', expected callback2 to be called")
	}
	if !allHooksCalled { // HookAll should still be called
		t.Errorf("After unregistering 'cb1', expected allHooksCalled to be true")
	}

	// --- Test Unregister All ---
	registry.Unregister(HookAll, "cbAll")

	// Invoke again
	resetFlags()                // Reset flags before invoke
	registry.Invoke(invokeArgs) // Use same args

	// Check results
	if callback1Called { // Should NOT be called
		t.Errorf("After unregistering 'cbAll', expected callback1 NOT to be called")
	}
	if !callback2Called { // Should still be called
		t.Errorf("After unregistering 'cbAll', expected callback2 to be called")
	}
	if allHooksCalled { // HookAll should NOT be called
		t.Errorf("After unregistering 'cbAll', expected allHooksCalled to be false")
	}

	// --- Test Unregister Non-existent ---
	registry.Unregister(HookAfterAgentRun, "cb_nonexistent")  // Non-existent hook point
	registry.Unregister(HookBeforeAgentRun, "cb_nonexistent") // Non-existent name

	// Invoke one last time to ensure cb2 is still there
	resetFlags()                // Reset flags before invoke
	registry.Invoke(invokeArgs) // Use same args

	// Check results
	if callback1Called { // Should NOT be called
		t.Errorf("After unregistering non-existent, expected callback1 NOT to be called")
	}
	if !callback2Called { // Should still be called
		t.Errorf("After unregistering non-existent, expected callback2 to be called")
	}
	if allHooksCalled { // HookAll should NOT be called
		t.Errorf("After unregistering non-existent, expected allHooksCalled to be false")
	}
}

func TestCallbackRegistry_Invoke(t *testing.T) {
	registry := NewCallbackRegistry()
	var counterSpecific int32
	var counterAll int32
	var counterOther int32

	cbSpecific := createTestCallback(&counterSpecific, nil)
	cbAll := createTestCallback(&counterAll, nil)
	cbOther := createTestCallback(&counterOther, nil) // Should not be called

	registry.Register(HookBeforeAgentRun, "cbSpecific", cbSpecific)
	registry.Register(HookAll, "cbAll", cbAll)
	registry.Register(HookAfterAgentRun, "cbOther", cbOther) // Different hook

	args := CallbackArgs{Hook: HookBeforeAgentRun} // Invoke for BeforeAgentRun

	registry.Invoke(args)

	if atomic.LoadInt32(&counterSpecific) != 1 {
		t.Errorf("Expected specific callback counter to be 1, got %d", counterSpecific)
	}
	if atomic.LoadInt32(&counterAll) != 1 {
		t.Errorf("Expected HookAll callback counter to be 1, got %d", counterAll)
	}
	if atomic.LoadInt32(&counterOther) != 0 {
		t.Errorf("Expected other hook callback counter to be 0, got %d", counterOther)
	}

	// Test invoking a hook with no registered callbacks
	atomic.StoreInt32(&counterSpecific, 0)
	atomic.StoreInt32(&counterAll, 0) // Reset HookAll counter
	argsNoCallbacks := CallbackArgs{Hook: HookAfterEventHandling}
	registry.Invoke(argsNoCallbacks) // Should not panic and only call HookAll

	if atomic.LoadInt32(&counterSpecific) != 0 {
		t.Errorf("Expected specific callback counter to be 0 after invoking empty hook, got %d", counterSpecific)
	}
	if atomic.LoadInt32(&counterAll) != 1 { // HookAll should be called again
		t.Errorf("Expected HookAll callback counter to be 1 after invoking empty hook, got %d", counterAll)
	}
}

func TestCallbackRegistry_Invoke_Multiple(t *testing.T) {
	registry := NewCallbackRegistry()
	var counter int32

	cb1 := createTestCallback(&counter, nil)
	cb2 := createTestCallback(&counter, nil)
	cbAll := createTestCallback(&counter, nil)

	registry.Register(HookBeforeAgentRun, "cb1", cb1)
	registry.Register(HookBeforeAgentRun, "cb2", cb2)
	registry.Register(HookAll, "cbAll", cbAll)

	args := CallbackArgs{Hook: HookBeforeAgentRun}
	registry.Invoke(args)

	// Expect 3 calls: cb1, cb2, cbAll
	if atomic.LoadInt32(&counter) != 3 {
		t.Errorf("Expected counter to be 3 after invoking hook with multiple specific and one HookAll callback, got %d", counter)
	}
}

func TestCallbackRegistry_Concurrency(t *testing.T) {
	registry := NewCallbackRegistry()
	var counter int32
	var wg sync.WaitGroup
	numGoroutines := 50
	numInvokesPerRoutine := 100

	// Pre-register one callback
	cb := createTestCallback(&counter, &wg)
	registry.Register(HookBeforeAgentRun, "cb_concurrent", cb)

	wg.Add(numGoroutines * numInvokesPerRoutine)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			// Simulate concurrent invokes
			for j := 0; j < numInvokesPerRoutine; j++ {
				registry.Invoke(CallbackArgs{Hook: HookBeforeAgentRun})
			}
		}(i)
	}

	wg.Wait()

	expectedCount := int32(numGoroutines * numInvokesPerRoutine)
	if atomic.LoadInt32(&counter) != expectedCount {
		t.Errorf("Expected counter to be %d after concurrent invokes, got %d", expectedCount, counter)
	}
}
