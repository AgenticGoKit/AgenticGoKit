package agentflow

import (
	"context" // <<< Add context import
	"log"
	"sync"
	"sync/atomic" // <<< Add atomic import
	"testing"
)

// Helper function to create a simple callback for testing
// This one has the correct signature (event Event)
func createTestCallback(id string, hook HookPoint, counter *int) CallbackFunc {
	return func(ctx context.Context, currentState State, event Event) (State, error) {
		*counter++
		log.Printf("TestCallback %s executed for hook %s (count: %d)", id, hook, *counter)
		// Optionally return a modified state or error for more complex tests
		return currentState, nil
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

	// FIX: Change e *Event to e Event (cb1 is already correct)
	cb1 := func(ctx context.Context, s State, e Event) (State, error) {
		callback1Called = true
		if s != nil {
			t.Log("cb1 received state")
		}
		if e != nil {
			t.Logf("cb1 received event ID: %s", e.GetID())
		} else {
			t.Log("cb1 received nil event")
		}
		return nil, nil
	}
	// FIX: Change e *Event to e Event (cb2 is already correct)
	cb2 := func(ctx context.Context, s State, e Event) (State, error) {
		callback2Called = true
		return &SimpleState{data: map[string]interface{}{"cb2_processed": true}}, nil
	}
	// FIX: Change e *Event to e Event for cbAll
	cbAll := func(ctx context.Context, s State, e Event) (State, error) {
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
	var counterSpecific, counterAll, counterOther int // Use int for the remaining helper

	// Use the correct helper function
	cbSpecific := createTestCallback("specific", HookBeforeAgentRun, &counterSpecific)
	cbAll := createTestCallback("all", HookAll, &counterAll)
	cbOther := createTestCallback("other", HookAfterAgentRun, &counterOther) // Should not be called

	registry.Register(HookBeforeAgentRun, "cbSpecific", cbSpecific)
	registry.Register(HookAll, "cbAll", cbAll)
	registry.Register(HookAfterAgentRun, "cbOther", cbOther) // Different hook

	args := CallbackArgs{Ctx: context.Background(), Hook: HookBeforeAgentRun} // Invoke for BeforeAgentRun

	registry.Invoke(args)

	if counterSpecific != 1 {
		t.Errorf("Expected specific callback counter to be 1, got %d", counterSpecific)
	}
	if counterAll != 1 {
		t.Errorf("Expected HookAll callback counter to be 1, got %d", counterAll)
	}
	if counterOther != 0 {
		t.Errorf("Expected other hook callback counter to be 0, got %d", counterOther)
	}

	// Test invoking a hook with no registered callbacks
	counterSpecific = 0 // Reset counters
	counterAll = 0
	argsNoCallbacks := CallbackArgs{Ctx: context.Background(), Hook: HookAfterEventHandling}
	registry.Invoke(argsNoCallbacks) // Should not panic and only call HookAll

	if counterSpecific != 0 {
		t.Errorf("Expected specific callback counter to be 0 after invoking empty hook, got %d", counterSpecific)
	}
	if counterAll != 1 { // HookAll should be called again
		t.Errorf("Expected HookAll callback counter to be 1 after invoking empty hook, got %d", counterAll)
	}
}

func TestCallbackRegistry_Invoke_Multiple(t *testing.T) {
	registry := NewCallbackRegistry()
	var counter int // Use int for the remaining helper

	cb1 := createTestCallback("cb1", HookBeforeAgentRun, &counter)
	cb2 := createTestCallback("cb2", HookBeforeAgentRun, &counter)
	cbAll := createTestCallback("cbAll", HookAll, &counter)

	registry.Register(HookBeforeAgentRun, "cb1", cb1)
	registry.Register(HookBeforeAgentRun, "cb2", cb2)
	registry.Register(HookAll, "cbAll", cbAll)

	args := CallbackArgs{Ctx: context.Background(), Hook: HookBeforeAgentRun}
	registry.Invoke(args)

	// Expect 3 calls: cb1, cb2, cbAll
	if counter != 3 {
		t.Errorf("Expected counter to be 3 after invoking hook with multiple specific and one HookAll callback, got %d", counter)
	}
}

func TestCallbackRegistry_Concurrency(t *testing.T) {
	registry := NewCallbackRegistry()
	// FIX: Change counter type to int32 or int64 for atomic operations
	var counter int64
	var wg sync.WaitGroup
	numGoroutines := 50
	numInvokesPerRoutine := 100

	// Pre-register one callback using the correct helper
	// Note: The helper doesn't handle WaitGroup, adjust if needed or use inline func
	cb := func(ctx context.Context, s State, e Event) (State, error) {
		// FIX: Use atomic increment
		atomic.AddInt64(&counter, 1) // Use AddInt64 for int64 counter
		wg.Done()
		return s, nil
	}
	registry.Register(HookBeforeAgentRun, "cb_concurrent", cb)

	wg.Add(numGoroutines * numInvokesPerRoutine)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			// Simulate concurrent invokes
			for j := 0; j < numInvokesPerRoutine; j++ {
				registry.Invoke(CallbackArgs{Ctx: context.Background(), Hook: HookBeforeAgentRun})
			}
		}(i)
	}

	wg.Wait()

	expectedCount := int64(numGoroutines * numInvokesPerRoutine) // Cast expected count to int64
	// FIX: Use atomic load to safely read the counter
	finalCount := atomic.LoadInt64(&counter)
	if finalCount != expectedCount {
		// Use %d for int64
		t.Errorf("Expected counter to be %d after concurrent invokes, got %d", expectedCount, finalCount)
	}
}
