package agentflow

import (
	"context"
	"reflect"
	"testing"
)

func TestMemorySessionStore(t *testing.T) {
	store := NewMemorySessionStore()
	ctx := context.Background()
	sessionID := "test-session-1"

	// 1. Test Get non-existent session
	state, found, err := store.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetSession failed for non-existent session: %v", err)
	}
	if found {
		t.Errorf("Expected found=false for non-existent session, got true")
	}
	if state != nil {
		t.Errorf("Expected nil state for non-existent session, got %v", state)
	}

	// 2. Test SaveSession (Create)
	newState := NewState() // Use State interface, implemented by *SimpleState
	newState.Set("count", 1)
	newState.SetMeta("agent", "tester")

	err = store.SaveSession(ctx, sessionID, newState) // Use SaveSession
	if err != nil {
		t.Fatalf("SaveSession (create) failed: %v", err)
	}

	// 3. Test Get existing session
	retrievedState, found, err := store.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetSession failed for existing session: %v", err)
	}
	if !found {
		t.Errorf("Expected found=true for existing session, got false")
	}
	if retrievedState == nil {
		t.Fatal("Expected non-nil state for existing session, got nil")
	}

	// Compare retrieved state with original (use Keys/MetaKeys for safety)
	if !reflect.DeepEqual(retrievedState.Keys(), newState.Keys()) {
		t.Errorf("Retrieved state data keys mismatch: got %v, want %v", retrievedState.Keys(), newState.Keys())
	}
	if !reflect.DeepEqual(retrievedState.MetaKeys(), newState.MetaKeys()) {
		t.Errorf("Retrieved state meta keys mismatch: got %v, want %v", retrievedState.MetaKeys(), newState.MetaKeys())
	}
	// Check a specific value
	countVal, _ := retrievedState.Get("count")
	if countVal != 1 {
		t.Errorf("Retrieved state 'count' mismatch: got %v, want 1", countVal)
	}

	// 4. Test SaveSession (Update)
	updatedState := retrievedState.Clone() // Clone before modifying
	updatedState.Set("count", 2)
	updatedState.Set("status", "updated")

	err = store.SaveSession(ctx, sessionID, updatedState) // Use SaveSession
	if err != nil {
		t.Fatalf("SaveSession (update) failed: %v", err)
	}

	// 5. Test Get updated session
	finalState, found, err := store.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetSession failed for updated session: %v", err)
	}
	if !found {
		t.Errorf("Expected found=true for updated session, got false")
	}
	updatedCount, _ := finalState.Get("count")
	if updatedCount != 2 {
		t.Errorf("Updated state 'count' mismatch: got %v, want 2", updatedCount)
	}
	statusVal, _ := finalState.Get("status")
	if statusVal != "updated" {
		t.Errorf("Updated state 'status' mismatch: got %v, want 'updated'", statusVal)
	}

	// 6. Test ListSessions (Optional, if implemented)
	// sessions, err := store.ListSessions(ctx, 0, 0, "") // Example signature
	// if err != nil { t.Fatalf("ListSessions failed: %v", err) }
	// Check if sessionID is in the list

	// 7. Test concurrency (Optional but recommended)
	// Add a new concurrency test using GetSession/SaveSession if needed
}

// Add tests for edge cases like nil state, empty session ID, context cancellation etc.
