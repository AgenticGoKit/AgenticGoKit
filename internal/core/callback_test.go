package agentflow

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Helper function to create a simple callback for testing
// Signature now matches CallbackFunc
func createTestCallback(id string, counter *int32) CallbackFunc {
	return func(ctx context.Context, args CallbackArgs) (State, error) {
		atomic.AddInt32(counter, 1) // Use atomic for safety if used concurrently
		// log.Printf("TestCallback %s executed for hook %s (count: %d), EventID: %s", id, args.Hook, atomic.LoadInt32(counter), args.Event.GetID())
		// Optionally return a modified state or error for more complex tests
		return args.State, nil // Return existing state by default
	}
}

func TestNewCallbackRegistry(t *testing.T) {
	registry := NewCallbackRegistry()
	if registry == nil {
		t.Fatal("NewCallbackRegistry returned nil")
	}
	if registry.callbacks == nil {
		t.Fatal("NewCallbackRegistry did not initialize callbacks map")
	}
}

// Test registration, unregistration, and invocation order.
func TestCallbackRegistry_Register_Unregister(t *testing.T) {
	registry := NewCallbackRegistry()
	ctx := context.Background()
	var order []string
	var mu sync.Mutex // Protect access to order slice

	// Callback 1 (Specific Hook)
	cb1 := func(ctx context.Context, args CallbackArgs) (State, error) {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, "cb1")
		log.Printf("Callback 1 executed for hook %s", args.Hook)

		// FIX: Use Clone() and Set() instead of GetData()
		newState := args.State.Clone() // Clone to modify safely
		newState.Set("cb1_called", true)
		return newState, nil
	}

	// Callback 2 (Specific Hook)
	cb2 := func(ctx context.Context, args CallbackArgs) (State, error) {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, "cb2")
		log.Printf("Callback 2 executed for hook %s", args.Hook)

		// FIX: Use Clone() and Set() instead of GetData()
		newState := args.State.Clone() // Clone to modify safely
		if called, ok := args.State.Get("cb1_called"); ok && called.(bool) {
			newState.Set("cb2_saw_cb1", true)
		}
		newState.Set("cb2_called", true)
		return newState, nil
	}

	// Callback 3 (HookAll)
	cbAll := func(ctx context.Context, args CallbackArgs) (State, error) {
		mu.Lock()
		defer mu.Unlock()
		// FIX: Use the actual hook name string, not the constant name
		order = append(order, fmt.Sprintf("cbAll_hook_%s", string(args.Hook)))
		log.Printf("Callback All executed for hook %s", args.Hook)

		// FIX: Use Clone() and Set() instead of GetData()
		newState := args.State.Clone() // Clone to modify safely
		if called, ok := args.State.Get("cb2_called"); ok && called.(bool) {
			newState.Set("cbAll_saw_cb2", true)
		}
		newState.Set("cbAll_called_on_"+string(args.Hook), true)
		return newState, nil
	}

	// Register callbacks
	// FIX: Pass HookPoint constants directly
	registry.Register(HookBeforeEventHandling, "cb1", cb1)
	registry.Register(HookBeforeEventHandling, "cb2", cb2)
	registry.Register(HookAll, "cbAll", cbAll)

	// --- Invoke HookBeforeEventHandling ---
	initialState := NewState() // Use NewState() which returns State interface
	initialState.Set("initial", "value")
	args := CallbackArgs{
		Hook:  HookBeforeEventHandling,
		Event: NewEvent("test", nil, nil),
		State: initialState, // Pass initial state
	}
	log.Println("Invoking HookBeforeEventHandling...")
	// Capture the final state returned by Invoke
	finalState, err := registry.Invoke(ctx, args) // Pass ctx
	if err != nil {
		t.Fatalf("Invoke failed for HookBeforeEventHandling: %v", err)
	}

	// Assertions for HookBeforeEventHandling
	mu.Lock() // Lock for reading order
	if len(order) != 3 {
		t.Errorf("Expected 3 callbacks to run for HookBeforeEventHandling, got %d: %v", len(order), order)
	} else {
		// Check presence, order might vary based on map iteration + sorting
		hasCb1 := false
		hasCb2 := false
		hasCbAll := false
		for _, name := range order {
			switch name {
			case "cb1":
				hasCb1 = true
			case "cb2":
				hasCb2 = true
			// FIX: Match the actual generated string
			case "cbAll_hook_BeforeEventHandling":
				hasCbAll = true
			}
		}
		if !hasCb1 || !hasCb2 || !hasCbAll {
			// FIX: Update expected string
			t.Errorf("Expected cb1, cb2, and cbAll_hook_BeforeEventHandling to run, got order: %v", order)
		}
	}
	mu.Unlock() // Unlock after reading order

	// Assert state modifications using the finalState returned by Invoke
	if finalState == nil {
		t.Fatalf("Invoke returned nil state, expected modified state")
	}

	// Check state modifications
	if cb1Called, ok := finalState.Get("cb1_called"); !ok || !cb1Called.(bool) {
		t.Errorf("State check failed: expected 'cb1_called' to be true in final state.")
	}
	if cb2SawCb1, ok := finalState.Get("cb2_saw_cb1"); !ok || !cb2SawCb1.(bool) {
		t.Errorf("State check failed: expected 'cb2_saw_cb1' to be true (cb2 should see cb1's change).")
	}
	if cb2Called, ok := finalState.Get("cb2_called"); !ok || !cb2Called.(bool) {
		t.Errorf("State check failed: expected 'cb2_called' to be true.")
	}
	if cbAllSawCb2, ok := finalState.Get("cbAll_saw_cb2"); !ok || !cbAllSawCb2.(bool) {
		t.Errorf("State check failed: expected 'cbAll_saw_cb2' to be true (cbAll should see cb2's change).")
	}
	// FIX: Update expected state key
	if cbAllCalled, ok := finalState.Get("cbAll_called_on_BeforeEventHandling"); !ok || !cbAllCalled.(bool) {
		// FIX: Update expected state key string in error message
		t.Errorf("State check failed: expected 'cbAll_called_on_BeforeEventHandling' to be true.")
	}
	if initialVal, ok := finalState.Get("initial"); !ok || initialVal.(string) != "value" {
		t.Errorf("State check failed: expected 'initial' value 'value' to be preserved.")
	}

	// Reset order for next hook
	mu.Lock()
	order = []string{}
	mu.Unlock()

	// --- Unregister cb1 ---
	// FIX: Pass HookPoint constant directly
	registry.Unregister(HookBeforeEventHandling, "cb1")
	log.Println("Unregistered cb1 for HookBeforeEventHandling.")

	// --- Invoke HookBeforeEventHandling again ---
	secondInitialState := NewState()
	secondInitialState.Set("second", "run")
	args.State = secondInitialState // Update state in args for the second run
	log.Println("Invoking HookBeforeEventHandling again after unregistering cb1...")
	secondFinalState, err := registry.Invoke(ctx, args) // Pass ctx
	if err != nil {
		t.Fatalf("Second Invoke failed for HookBeforeEventHandling: %v", err)
	}

	// Assertions for second HookBeforeEventHandling invocation
	mu.Lock()
	if len(order) != 2 { // cb2 and cbAll should run
		t.Errorf("Expected 2 callbacks after unregistering cb1, got %d: %v", len(order), order)
	} else {
		// Check presence, order might vary
		hasCb2 := false
		hasCbAll := false
		for _, name := range order {
			switch name {
			case "cb2":
				hasCb2 = true
			// FIX: Match the actual generated string
			case "cbAll_hook_BeforeEventHandling":
				hasCbAll = true
			}
		}
		if !hasCb2 || !hasCbAll {
			// FIX: Update expected string
			t.Errorf("Expected cb2 and cbAll_hook_BeforeEventHandling to run, got order: %v", order)
		}
	}
	mu.Unlock()

	// Assert state for the second run using secondFinalState
	if secondFinalState == nil {
		t.Fatalf("Second Invoke returned nil state")
	}

	if _, exists := secondFinalState.Get("cb1_called"); exists {
		t.Errorf("State check failed: 'cb1_called' should not exist in state after cb1 was unregistered.")
	}
	if _, exists := secondFinalState.Get("cb2_saw_cb1"); exists {
		t.Errorf("State check failed: 'cb2_saw_cb1' should not exist as cb1 did not run.")
	}
	if cb2Called, ok := secondFinalState.Get("cb2_called"); !ok || !cb2Called.(bool) {
		t.Errorf("State check failed: expected 'cb2_called' to be true in second run.")
	}
	if cbAllSawCb2, ok := secondFinalState.Get("cbAll_saw_cb2"); !ok || !cbAllSawCb2.(bool) {
		t.Errorf("State check failed: expected 'cbAll_saw_cb2' to be true in second run.")
	}
	// FIX: Update expected state key
	if cbAllCalled, ok := secondFinalState.Get("cbAll_called_on_BeforeEventHandling"); !ok || !cbAllCalled.(bool) {
		// FIX: Update expected state key string in error message
		t.Errorf("State check failed: expected 'cbAll_called_on_BeforeEventHandling' to be true in second run.")
	}
	if secondVal, ok := secondFinalState.Get("second"); !ok || secondVal.(string) != "run" {
		t.Errorf("State check failed: expected 'second' value 'run' to be present.")
	}

	// Reset order
	mu.Lock()
	order = []string{}
	mu.Unlock()

	// --- Invoke HookAfterEventHandling (only cbAll should run) ---
	afterArgs := CallbackArgs{
		Hook:  HookAfterEventHandling,
		Event: NewEvent("test", nil, nil),
		State: NewState(), // Fresh state
	}
	afterArgs.State.Set("after", "hook")
	log.Println("Invoking HookAfterEventHandling...")
	afterFinalState, err := registry.Invoke(ctx, afterArgs) // Pass ctx
	if err != nil {
		t.Fatalf("Invoke failed for HookAfterEventHandling: %v", err)
	}

	// Assertions for HookAfterEventHandling
	mu.Lock()
	// FIX: Match the actual generated string
	if len(order) != 1 || order[0] != "cbAll_hook_AfterEventHandling" {
		// FIX: Update expected string
		t.Errorf("Expected 1 callback ('cbAll_hook_AfterEventHandling') for HookAfterEventHandling, got %v", order)
	}
	mu.Unlock()

	// Assert state for HookAfterEventHandling
	if afterFinalState == nil {
		t.Fatalf("Invoke for HookAfterEventHandling returned nil state")
	}

	if _, exists := afterFinalState.Get("cb1_called"); exists {
		t.Errorf("State check failed: 'cb1_called' should not exist for HookAfterEventHandling run.")
	}
	if _, exists := afterFinalState.Get("cb2_called"); exists {
		t.Errorf("State check failed: 'cb2_called' should not exist for HookAfterEventHandling run.")
	}
	// FIX: Update expected state key
	if cbAllCalled, ok := afterFinalState.Get("cbAll_called_on_AfterEventHandling"); !ok || !cbAllCalled.(bool) {
		// FIX: Update expected state key string in error message
		t.Errorf("State check failed: expected 'cbAll_called_on_AfterEventHandling' to be true.")
	}
	if afterVal, ok := afterFinalState.Get("after"); !ok || afterVal.(string) != "hook" {
		t.Errorf("State check failed: expected 'after' value 'hook' to be present.")
	}

	// --- Unregister cbAll ---
	// FIX: Pass HookPoint constant directly
	registry.Unregister(HookAll, "cbAll")
	log.Println("Unregistered cbAll for HookAll.")

	// --- Invoke HookBeforeEventHandling again (no callbacks should run) ---
	finalArgs := CallbackArgs{
		Hook:  HookBeforeEventHandling,
		Event: NewEvent("test", nil, nil),
		State: NewState(), // Fresh state
	}
	finalArgs.State.Set("final", "check")
	log.Println("Invoking HookBeforeEventHandling after unregistering cbAll...")
	// FIX: Reset order before final invoke
	mu.Lock()
	order = []string{}
	mu.Unlock()
	veryFinalState, err := registry.Invoke(ctx, finalArgs) // Pass ctx
	if err != nil {
		t.Fatalf("Final Invoke failed for HookBeforeEventHandling: %v", err)
	}

	// Assertions for final HookBeforeEventHandling invocation
	mu.Lock()
	// FIX: Expect cb2 to run, as it was never unregistered for this hook
	if len(order) != 1 || order[0] != "cb2" {
		// FIX: Update expected callback list
		t.Errorf("Expected 1 callback ('cb2') after unregistering cbAll, got %d: %v", len(order), order)
	}
	mu.Unlock()

	// Assert final state
	// FIX: Expect 2 keys: 'final' (set before invoke) and 'cb2_called' (set by cb2)
	if len(veryFinalState.Keys()) != 2 {
		// FIX: Update expected key count and list
		t.Errorf("Expected final state to have 2 keys ('final', 'cb2_called'), got %d keys: %v", len(veryFinalState.Keys()), veryFinalState.Keys())
	}
	if finalVal, ok := veryFinalState.Get("final"); !ok || finalVal.(string) != "check" {
		t.Errorf("Expected final state key 'final' to have value 'check', got %v (State: %v)", finalVal, veryFinalState)
	}
	// FIX: Verify cb2's state key is present
	if _, ok := veryFinalState.Get("cb2_called"); !ok {
		t.Errorf("Expected final state key 'cb2_called' to be present, but it was missing (State: %v)", veryFinalState)
	}
}

func TestCallbackRegistry_Invoke(t *testing.T) {
	registry := NewCallbackRegistry()
	var countSpecific, countAll, countOther int32 // Use int32 for atomic ops
	ctx := context.Background()                   // Define context

	// Use createTestCallback helper
	cbSpecific := createTestCallback("cbSpecific", &countSpecific)
	cbAll := createTestCallback("cbAll", &countAll)
	cbOther := createTestCallback("cbOther", &countOther)

	// FIX: Pass HookPoint constants directly
	registry.Register(HookBeforeAgentRun, "cbSpecific", cbSpecific)
	registry.Register(HookAll, "cbAll", cbAll)
	registry.Register(HookAfterAgentRun, "cbOther", cbOther)

	testEvent := NewEvent("agent1", EventData{"key": "value"}, nil)
	initialState := NewState() // Assuming NewState() returns a State interface value

	// Invoke for BeforeAgentRun
	argsBefore := CallbackArgs{
		Hook:  HookBeforeAgentRun,
		Event: testEvent,
		State: initialState,
	}
	_, err := registry.Invoke(ctx, argsBefore) // Pass ctx
	if err != nil {
		t.Fatalf("Invoke(BeforeAgentRun) failed: %v", err)
	}

	if atomic.LoadInt32(&countSpecific) != 1 {
		t.Errorf("Expected countSpecific=1 for BeforeAgentRun, got %d", countSpecific)
	}
	if atomic.LoadInt32(&countAll) != 1 { // HookAll runs for BeforeAgentRun
		t.Errorf("Expected countAll=1 for BeforeAgentRun, got %d", countAll)
	}
	if atomic.LoadInt32(&countOther) != 0 {
		t.Errorf("Expected countOther=0 for BeforeAgentRun, got %d", countOther)
	}

	// Reset counts
	atomic.StoreInt32(&countSpecific, 0)
	atomic.StoreInt32(&countAll, 0)
	atomic.StoreInt32(&countOther, 0)

	// Invoke for AfterAgentRun
	argsAfter := CallbackArgs{
		Hook:  HookAfterAgentRun,
		Event: testEvent,
		State: initialState, // Or potentially the state returned from Before hooks
	}
	_, err = registry.Invoke(ctx, argsAfter) // Pass ctx
	if err != nil {
		t.Fatalf("Invoke(AfterAgentRun) failed: %v", err)
	}

	if atomic.LoadInt32(&countSpecific) != 0 {
		t.Errorf("Expected countSpecific=0 for AfterAgentRun, got %d", countSpecific)
	}
	if atomic.LoadInt32(&countAll) != 1 { // HookAll runs for AfterAgentRun
		t.Errorf("Expected countAll=1 for AfterAgentRun, got %d", countAll)
	}
	if atomic.LoadInt32(&countOther) != 1 {
		t.Errorf("Expected countOther=1 for AfterAgentRun, got %d", countOther)
	}
}

func TestCallbackRegistry_Invoke_Multiple(t *testing.T) {
	registry := NewCallbackRegistry()
	var count1, count2, countAll int32 // Use atomic for potential concurrency
	ctx := context.Background()

	// Use createTestCallback helper
	cb1 := createTestCallback("cb1", &count1)
	cb2 := createTestCallback("cb2", &count2)
	cbAll := createTestCallback("cbAll", &countAll)

	// FIX: Pass HookPoint constants directly
	registry.Register(HookBeforeAgentRun, "cb1", cb1)
	registry.Register(HookBeforeAgentRun, "cb2", cb2)
	registry.Register(HookAll, "cbAll", cbAll)

	testEvent := NewEvent("agent1", EventData{"key": "value"}, nil)
	initialState := NewState()
	args := CallbackArgs{
		Hook:  HookBeforeAgentRun,
		Event: testEvent,
		State: initialState,
	}
	_, err := registry.Invoke(ctx, args) // Pass ctx
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if atomic.LoadInt32(&count1) != 1 {
		t.Errorf("Expected count1=1, got %d", count1)
	}
	if atomic.LoadInt32(&count2) != 1 {
		t.Errorf("Expected count2=1, got %d", count2)
	}
	if atomic.LoadInt32(&countAll) != 1 { // HookAll runs too
		t.Errorf("Expected countAll=1, got %d", countAll)
	}
}

func TestCallbackRegistry_Concurrency(t *testing.T) {
	registry := NewCallbackRegistry()
	var counter int32 // Use atomic counter for concurrency
	ctx := context.Background()

	cb := func(ctx context.Context, args CallbackArgs) (State, error) {
		atomic.AddInt32(&counter, 1)
		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Millisecond)
		return args.State, nil
	}

	// Register the callback *once* before starting goroutines
	// FIX: Pass HookPoint constant directly
	err := registry.Register(HookBeforeAgentRun, "cb_concurrent", cb)
	if err != nil {
		t.Fatalf("Failed to register callback: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 50
	numInvokesPerGoroutine := 10

	testEvent := NewEvent("agent1", EventData{"key": "value"}, nil)
	initialState := NewState()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numInvokesPerGoroutine; j++ {
				args := CallbackArgs{
					Hook:  HookBeforeAgentRun,
					Event: testEvent,
					State: initialState,
				}
				// FIX: Pass context to Invoke
				_, invokeErr := registry.Invoke(ctx, args)
				if invokeErr != nil {
					// Log error from Invoke if it happens, but don't fail test immediately
					// Use t.Logf to avoid failing the test from within goroutine
					t.Logf("Error during concurrent Invoke: %v", invokeErr)
				}
			}
		}()
	}

	wg.Wait()

	expectedCount := int32(numGoroutines * numInvokesPerGoroutine)
	finalCount := atomic.LoadInt32(&counter)
	t.Logf("Final callback count: %d (Expected: %d)", finalCount, expectedCount)

	if finalCount != expectedCount {
		t.Errorf("Expected callback count %d, but got %d", expectedCount, finalCount)
	}
	if finalCount == 0 {
		t.Errorf("Callback was never invoked")
	}
}
