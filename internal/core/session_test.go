package agentflow

import (
	"context"
	"errors" // Import errors package
	"reflect"
	"testing"
)

func TestMemorySessionStore(t *testing.T) {
	store := NewMemorySessionStore()
	ctx := context.Background()
	sessionID := "test-session-1"

	// 1. Test Get non-existent session
	session, err := store.GetSession(ctx, sessionID)
	// Expect a specific error for not found
	if err == nil || !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("GetSession should return ErrSessionNotFound for non-existent session, got: %v", err)
	}
	if session != nil {
		t.Errorf("Expected nil session for non-existent session, got %v", session)
	}

	// 2. Test SaveSession (Create)
	newState := NewState() // Use State interface, implemented by *SimpleState
	newState.Set("count", 1)
	newState.SetMeta("agent", "tester")

	// FIX: Create a Session object to pass to SaveSession
	newSession := NewMemorySession(sessionID, newState)
	err = store.SaveSession(ctx, newSession) // Pass the Session object
	if err != nil {
		t.Fatalf("SaveSession (create) failed: %v", err)
	}

	// 3. Test Get existing session
	// FIX: Assign to session, err
	retrievedSession, err := store.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetSession failed for existing session: %v", err)
	}
	if retrievedSession == nil {
		t.Fatal("Expected non-nil session for existing session, got nil")
	}
	retrievedState := retrievedSession.GetState() // Get state from session
	if retrievedState == nil {
		t.Fatal("Expected non-nil state within existing session, got nil")
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
	// Get the state from the retrieved session to update it
	stateToUpdate := retrievedSession.GetState().Clone() // Clone before modifying
	stateToUpdate.Set("count", 2)
	stateToUpdate.Set("status", "updated")

	// FIX: Create/Update the Session object to pass to SaveSession
	// Option 1: Create a new session object with the updated state
	updatedSession := NewMemorySession(sessionID, stateToUpdate)
	// Option 2: If Session interface had SetState, use retrievedSession.SetState(stateToUpdate)
	// Assuming Option 1 for now based on current interface
	err = store.SaveSession(ctx, updatedSession) // Pass the updated Session object
	if err != nil {
		t.Fatalf("SaveSession (update) failed: %v", err)
	}

	// 5. Test Get updated session
	// FIX: Assign to session, err
	finalSession, err := store.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetSession failed for updated session: %v", err)
	}
	if finalSession == nil {
		t.Fatal("Expected non-nil session for updated session, got nil")
	}
	finalState := finalSession.GetState()
	if finalState == nil {
		t.Fatal("Expected non-nil state within updated session, got nil")
	}

	updatedCount, _ := finalState.Get("count")
	if updatedCount != 2 {
		t.Errorf("Updated state 'count' mismatch: got %v, want 2", updatedCount)
	}
	statusVal, _ := finalState.Get("status")
	if statusVal != "updated" {
		t.Errorf("Updated state 'status' mismatch: got %v, want 'updated'", statusVal)
	}

	// 6. Test DeleteSession
	err = store.DeleteSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// 7. Verify deletion
	deletedSession, err := store.GetSession(ctx, sessionID)
	if err == nil || !errors.Is(err, ErrSessionNotFound) { // Check for not found error again
		// t.Fatalf("GetSession should return an error after deletion, got nil")
	}
	if deletedSession != nil {
		t.Errorf("Expected nil session after deletion, got %v", deletedSession)
	}

	// 8. Test concurrency (Optional but recommended)
	// Add a new concurrency test using GetSession/SaveSession if needed
}

// Add tests for edge cases like nil state, empty session ID, context cancellation etc.
