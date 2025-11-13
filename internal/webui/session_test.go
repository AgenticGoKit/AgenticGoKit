package webui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
)

func TestInMemoryStorage(t *testing.T) {
	storage := NewInMemoryStorage()

	// Test session creation and storage
	session := &ChatSession{
		ID:        "test-session-1",
		Messages:  []ChatMessage{},
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		Active:    true,
	}

	// Test Save
	err := storage.Save(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Test Load
	loadedSession, err := storage.Load("test-session-1")
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if loadedSession.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, loadedSession.ID)
	}

	// Test List
	sessionIDs, err := storage.List()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessionIDs) != 1 || sessionIDs[0] != "test-session-1" {
		t.Errorf("Expected 1 session with ID test-session-1, got %v", sessionIDs)
	}

	// Test Delete
	err = storage.Delete("test-session-1")
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify deletion
	_, err = storage.Load("test-session-1")
	if err == nil {
		t.Error("Expected error when loading deleted session")
	}

	// Test Clear
	storage.Save(session)
	err = storage.Clear()
	if err != nil {
		t.Fatalf("Failed to clear storage: %v", err)
	}

	sessionIDs, _ = storage.List()
	if len(sessionIDs) != 0 {
		t.Errorf("Expected 0 sessions after clear, got %d", len(sessionIDs))
	}

	t.Log("InMemoryStorage test passed")
}

func TestFileStorage(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "webui_test_sessions")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file storage: %v", err)
	}

	// Test session creation and storage
	session := &ChatSession{
		ID:        "test-session-file",
		Messages:  []ChatMessage{{ID: "msg1", Content: "Hello", Role: "user"}},
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		UserAgent: "test-agent",
		Active:    true,
	}

	// Test Save
	err = storage.Save(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tempDir, "test-session-file.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Session file was not created")
	}

	// Test Load
	loadedSession, err := storage.Load("test-session-file")
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if loadedSession.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, loadedSession.ID)
	}

	if len(loadedSession.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(loadedSession.Messages))
	}

	// Test List
	sessionIDs, err := storage.List()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if len(sessionIDs) != 1 || sessionIDs[0] != "test-session-file" {
		t.Errorf("Expected 1 session with ID test-session-file, got %v", sessionIDs)
	}

	// Test Delete
	err = storage.Delete("test-session-file")
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Session file was not deleted")
	}

	t.Log("FileStorage test passed")
}

func TestEnhancedSessionManager(t *testing.T) {
	config := &core.Config{}
	sessionConfig := DefaultSessionConfig()
	sessionConfig.MaxSessions = 5
	sessionConfig.MaxMessages = 10
	sessionConfig.SessionTimeout = 1 * time.Second // Short timeout for testing

	manager, err := NewEnhancedSessionManager(config, sessionConfig)
	if err != nil {
		t.Fatalf("Failed to create enhanced session manager: %v", err)
	}
	defer manager.Stop()

	// Test session creation
	session, err := manager.CreateSession("test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}

	if session.UserAgent != "test-agent" {
		t.Errorf("Expected user agent 'test-agent', got '%s'", session.UserAgent)
	}

	// Test session retrieval
	retrievedSession, err := manager.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrievedSession.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, retrievedSession.ID)
	}

	// Test adding messages
	message := ChatMessage{
		ID:      "test-msg-1",
		Role:    "user",
		Content: "Hello, world!",
	}

	err = manager.AddMessage(session.ID, message)
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Verify message was added
	updatedSession, err := manager.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	if len(updatedSession.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(updatedSession.Messages))
	}

	// Test session listing
	sessions, total, err := manager.ListSessions(0, 10)
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 total session, got %d", total)
	}

	if len(sessions) != 1 {
		t.Errorf("Expected 1 session in list, got %d", len(sessions))
	}

	// Test metrics
	metrics, err := manager.GetMetrics()
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if metrics.TotalSessions != 1 {
		t.Errorf("Expected 1 total session in metrics, got %d", metrics.TotalSessions)
	}

	if metrics.ActiveSessions != 1 {
		t.Errorf("Expected 1 active session in metrics, got %d", metrics.ActiveSessions)
	}

	if metrics.TotalMessages != 1 {
		t.Errorf("Expected 1 total message in metrics, got %d", metrics.TotalMessages)
	}

	t.Log("EnhancedSessionManager basic test passed")
}

func TestEnhancedSessionExpiration(t *testing.T) {
	config := &core.Config{}
	sessionConfig := DefaultSessionConfig()
	sessionConfig.SessionTimeout = 100 * time.Millisecond // Very short timeout
	sessionConfig.CleanupInterval = 50 * time.Millisecond

	manager, err := NewEnhancedSessionManager(config, sessionConfig)
	if err != nil {
		t.Fatalf("Failed to create enhanced session manager: %v", err)
	}
	defer manager.Stop()

	// Create a session
	session, err := manager.CreateSession("test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Wait for session to expire
	time.Sleep(200 * time.Millisecond)

	// Try to get expired session
	_, err = manager.GetSession(session.ID)
	if err == nil {
		t.Error("Expected error when getting expired session")
	}

	// Wait for cleanup routine
	time.Sleep(100 * time.Millisecond)

	// Verify session was cleaned up
	sessions, total, err := manager.ListSessions(0, 10)
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if total != 0 {
		t.Errorf("Expected 0 sessions after cleanup, got %d", total)
	}

	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions in list after cleanup, got %d", len(sessions))
	}

	t.Log("Session expiration test passed")
}

func TestSessionLimits(t *testing.T) {
	config := &core.Config{}
	sessionConfig := DefaultSessionConfig()
	sessionConfig.MaxSessions = 2
	sessionConfig.MaxMessages = 3

	manager, err := NewEnhancedSessionManager(config, sessionConfig)
	if err != nil {
		t.Fatalf("Failed to create enhanced session manager: %v", err)
	}
	defer manager.Stop()

	// Test max sessions limit
	session1, err := manager.CreateSession("agent1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Failed to create first session: %v", err)
	}

	_, err = manager.CreateSession("agent2", "127.0.0.1")
	if err != nil {
		t.Fatalf("Failed to create second session: %v", err)
	}

	// Try to create third session (should fail)
	_, err = manager.CreateSession("agent3", "127.0.0.1")
	if err == nil {
		t.Error("Expected error when exceeding max sessions")
	}

	// Test max messages limit
	for i := 0; i < 5; i++ {
		message := ChatMessage{
			ID:      fmt.Sprintf("msg-%d", i),
			Role:    "user",
			Content: fmt.Sprintf("Message %d", i),
		}

		err = manager.AddMessage(session1.ID, message)
		if err != nil {
			t.Fatalf("Failed to add message %d: %v", i, err)
		}
	}

	// Check that old messages were removed
	updatedSession, err := manager.GetSession(session1.ID)
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	if len(updatedSession.Messages) != 3 {
		t.Errorf("Expected 3 messages (max limit), got %d", len(updatedSession.Messages))
	}

	// Verify that the latest messages are kept
	if updatedSession.Messages[0].Content != "Message 2" {
		t.Errorf("Expected oldest remaining message to be 'Message 2', got '%s'", updatedSession.Messages[0].Content)
	}

	t.Log("Session limits test passed")
}

func TestSessionCallbacks(t *testing.T) {
	config := &core.Config{}
	sessionConfig := DefaultSessionConfig()

	manager, err := NewEnhancedSessionManager(config, sessionConfig)
	if err != nil {
		t.Fatalf("Failed to create enhanced session manager: %v", err)
	}
	defer manager.Stop()

	// Track callback invocations
	var createdSession *ChatSession
	var updatedSession *ChatSession
	var deletedSessionID string

	manager.SetCallbacks(
		func(session *ChatSession) {
			createdSession = session
		},
		func(session *ChatSession) {
			updatedSession = session
		},
		func(sessionID string) {
			deletedSessionID = sessionID
		},
		func(count int) {
			// Log cleanup count but don't store in unused variable
			t.Logf("Cleanup called with count: %d", count)
		},
	) // Test create callback
	session, err := manager.CreateSession("test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if createdSession == nil || createdSession.ID != session.ID {
		t.Error("Create callback was not invoked correctly")
	}

	// Test update callback
	err = manager.UpdateSession(session)
	if err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	if updatedSession == nil || updatedSession.ID != session.ID {
		t.Error("Update callback was not invoked correctly")
	}

	// Test delete callback
	err = manager.DeleteSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	if deletedSessionID != session.ID {
		t.Error("Delete callback was not invoked correctly")
	}

	t.Log("Session callbacks test passed")
}

func TestSessionPagination(t *testing.T) {
	config := &core.Config{}
	sessionConfig := DefaultSessionConfig()
	sessionConfig.MaxSessions = 100

	manager, err := NewEnhancedSessionManager(config, sessionConfig)
	if err != nil {
		t.Fatalf("Failed to create enhanced session manager: %v", err)
	}
	defer manager.Stop()

	// Create multiple sessions
	numSessions := 10
	for i := 0; i < numSessions; i++ {
		_, err := manager.CreateSession(fmt.Sprintf("agent-%d", i), "127.0.0.1")
		if err != nil {
			t.Fatalf("Failed to create session %d: %v", i, err)
		}
	}

	// Test pagination
	sessions, total, err := manager.ListSessions(0, 5)
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	if total != numSessions {
		t.Errorf("Expected %d total sessions, got %d", numSessions, total)
	}

	if len(sessions) != 5 {
		t.Errorf("Expected 5 sessions in first page, got %d", len(sessions))
	}

	// Test second page
	sessions, total, err = manager.ListSessions(5, 5)
	if err != nil {
		t.Fatalf("Failed to list sessions (page 2): %v", err)
	}

	if total != numSessions {
		t.Errorf("Expected %d total sessions, got %d", numSessions, total)
	}

	if len(sessions) != 5 {
		t.Errorf("Expected 5 sessions in second page, got %d", len(sessions))
	}

	// Test beyond end
	sessions, total, err = manager.ListSessions(15, 5)
	if err != nil {
		t.Fatalf("Failed to list sessions (beyond end): %v", err)
	}

	if total != numSessions {
		t.Errorf("Expected %d total sessions, got %d", numSessions, total)
	}

	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions beyond end, got %d", len(sessions))
	}

	t.Log("Session pagination test passed")
}

