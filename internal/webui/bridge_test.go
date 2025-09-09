package webui

import (
	"context"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// Simple test for basic functionality without external dependencies

func TestWebUIEvent_Validation(t *testing.T) {
	tests := []struct {
		name    string
		event   *WebUIEvent
		wantErr bool
	}{
		{
			name: "valid event",
			event: &WebUIEvent{
				ID:        "test-id",
				SessionID: "test-session",
				Type:      "chat_message",
				Message:   "Hello",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty session ID",
			event: &WebUIEvent{
				ID:        "test-id",
				SessionID: "",
				Type:      "chat_message",
				Message:   "Hello",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty message",
			event: &WebUIEvent{
				ID:        "test-id",
				SessionID: "test-session",
				Type:      "chat_message",
				Message:   "",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebUIEvent(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWebUIEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentResponse_Creation(t *testing.T) {
	response := &AgentResponse{
		ID:        "test-response",
		SessionID: "test-session",
		AgentName: "test-agent",
		Content:   "Test response",
		Status:    "complete",
		Timestamp: time.Now(),
	}

	if response.ID != "test-response" {
		t.Errorf("Expected ID 'test-response', got %s", response.ID)
	}

	if response.SessionID != "test-session" {
		t.Errorf("Expected SessionID 'test-session', got %s", response.SessionID)
	}

	if response.Status != "complete" {
		t.Errorf("Expected Status 'complete', got %s", response.Status)
	}
}

func TestAgentInfo_Creation(t *testing.T) {
	info := AgentInfo{
		Name:         "test-agent",
		Description:  "A test agent",
		Capabilities: []string{"chat", "text"},
		IsEnabled:    true,
		Role:         "assistant",
	}

	if info.Name != "test-agent" {
		t.Errorf("Expected Name 'test-agent', got %s", info.Name)
	}

	if len(info.Capabilities) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(info.Capabilities))
	}

	if !info.IsEnabled {
		t.Error("Expected IsEnabled to be true")
	}
}

func TestGenerateEventID(t *testing.T) {
	id1 := generateEventID()
	id2 := generateEventID()

	if id1 == id2 {
		t.Error("Expected different event IDs")
	}

	if len(id1) == 0 {
		t.Error("Expected non-empty event ID")
	}
}

func TestGenerateResponseID(t *testing.T) {
	id1 := generateResponseID()
	id2 := generateResponseID()

	if id1 == id2 {
		t.Error("Expected different response IDs")
	}

	if len(id1) == 0 {
		t.Error("Expected non-empty response ID")
	}
}

// Simple validation function for testing
func validateWebUIEvent(event *WebUIEvent) error {
	if event == nil {
		return &core.ValidationError{Field: "event", Message: "event cannot be nil"}
	}

	if event.SessionID == "" {
		return &core.ValidationError{Field: "session_id", Message: "session ID cannot be empty"}
	}

	if event.Message == "" {
		return &core.ValidationError{Field: "message", Message: "message cannot be empty"}
	}

	return nil
}

func TestBridgeConfig_Defaults(t *testing.T) {
	config := DefaultBridgeConfig()

	if config.ResponseBufferSize <= 0 {
		t.Error("Expected positive ResponseBufferSize")
	}

	if config.AgentTimeout <= 0 {
		t.Error("Expected positive AgentTimeout")
	}

	if config.MaxConcurrentTasks <= 0 {
		t.Error("Expected positive MaxConcurrentTasks")
	}

	if config.RetryAttempts <= 0 {
		t.Error("Expected positive RetryAttempts")
	}

	if config.RetryDelay <= 0 {
		t.Error("Expected positive RetryDelay")
	}
}

func TestWebUIEvent_WithMetadata(t *testing.T) {
	event := &WebUIEvent{
		ID:        "test-id",
		SessionID: "test-session",
		Type:      "chat_message",
		Message:   "Hello",
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	event.Metadata["user_agent"] = "test-browser"
	event.Metadata["ip_address"] = "127.0.0.1"

	if event.Metadata["user_agent"] != "test-browser" {
		t.Error("Expected user_agent metadata to be set")
	}

	if event.Metadata["ip_address"] != "127.0.0.1" {
		t.Error("Expected ip_address metadata to be set")
	}
}

func TestAgentResponse_WithStreaming(t *testing.T) {
	response := &AgentResponse{
		ID:          "test-response",
		SessionID:   "test-session",
		AgentName:   "test-agent",
		Content:     "Partial content",
		Status:      "partial",
		IsStreaming: true,
		ChunkIndex:  1,
		TotalChunks: 3,
		Timestamp:   time.Now(),
	}

	if !response.IsStreaming {
		t.Error("Expected IsStreaming to be true")
	}

	if response.ChunkIndex != 1 {
		t.Errorf("Expected ChunkIndex 1, got %d", response.ChunkIndex)
	}

	if response.TotalChunks != 3 {
		t.Errorf("Expected TotalChunks 3, got %d", response.TotalChunks)
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context
	cancel()

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Expected context to be cancelled")
	}
}

func TestTimeouts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Wait for timeout
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(20 * time.Millisecond):
		t.Error("Expected context to timeout")
	}
}
