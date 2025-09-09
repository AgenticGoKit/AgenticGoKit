package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kunalkushwaha/agenticgokit/core"
)

func TestWebSocketConnection(t *testing.T) {
	// Create test server
	config := &core.Config{}

	sessionManager := NewSessionManager(config)
	connManager := NewConnectionManager(sessionManager)
	connManager.Start()
	defer connManager.Stop()

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(connManager.HandleWebSocket))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Test should pass if connection succeeds
	t.Log("WebSocket connection established successfully")
}

func TestWebSocketProtocol(t *testing.T) {
	// Create test server
	config := &core.Config{}

	sessionManager := NewSessionManager(config)
	connManager := NewConnectionManager(sessionManager)
	connManager.Start()
	defer connManager.Stop()

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(connManager.HandleWebSocket))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Test session creation
	createMsg := &WebSocketMessage{
		Type:      MsgTypeSessionCreate,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"user_agent": "test-agent",
		},
	}

	err = conn.WriteJSON(createMsg)
	if err != nil {
		t.Fatalf("Failed to send session create message: %v", err)
	}

	// Read welcome message first
	var welcomeMsg WebSocketMessage
	err = conn.ReadJSON(&welcomeMsg)
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	// Read session status response
	var response WebSocketMessage
	err = conn.ReadJSON(&response)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if response.Type != MsgTypeSessionStatus {
		t.Errorf("Expected session status message, got %s", response.Type)
	}

	t.Log("Session creation protocol test passed")
}

func TestWebSocketChatFlow(t *testing.T) {
	// Create test server
	config := &core.Config{}

	sessionManager := NewSessionManager(config)
	connManager := NewConnectionManager(sessionManager)
	connManager.Start()
	defer connManager.Stop()

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(connManager.HandleWebSocket))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Create session first
	createMsg := &WebSocketMessage{
		Type:      MsgTypeSessionCreate,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data:      map[string]interface{}{},
	}

	err = conn.WriteJSON(createMsg)
	if err != nil {
		t.Fatalf("Failed to send session create message: %v", err)
	}

	// Read welcome message first
	var welcomeMsg WebSocketMessage
	err = conn.ReadJSON(&welcomeMsg)
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	// Read session status response
	var sessionResponse WebSocketMessage
	err = conn.ReadJSON(&sessionResponse)
	if err != nil {
		t.Fatalf("Failed to read session response: %v", err)
	}

	sessionID := sessionResponse.SessionID

	// Send chat message
	chatMsg := &WebSocketMessage{
		Type:      MsgTypeChatMessage,
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"content":      "Hello, agent!",
			"message_type": "text",
		},
	}

	err = conn.WriteJSON(chatMsg)
	if err != nil {
		t.Fatalf("Failed to send chat message: %v", err)
	}

	// Read agent response
	var agentResponse WebSocketMessage
	err = conn.ReadJSON(&agentResponse)
	if err != nil {
		t.Fatalf("Failed to read agent response: %v", err)
	}

	if agentResponse.Type != MsgTypeAgentResponse {
		t.Errorf("Expected agent response message, got %s", agentResponse.Type)
	}

	if agentResponse.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, agentResponse.SessionID)
	}

	t.Log("Chat flow test passed")
}

func TestWebSocketPingPong(t *testing.T) {
	// Create test server
	config := &core.Config{}

	sessionManager := NewSessionManager(config)
	connManager := NewConnectionManager(sessionManager)
	connManager.Start()
	defer connManager.Stop()

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(connManager.HandleWebSocket))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Send ping
	pingMsg := &WebSocketMessage{
		Type:      MsgTypePing,
		MessageID: generateMessageID(),
		Timestamp: time.Now(),
	}

	err = conn.WriteJSON(pingMsg)
	if err != nil {
		t.Fatalf("Failed to send ping message: %v", err)
	}

	// Read pong response
	var pongResponse WebSocketMessage
	err = conn.ReadJSON(&pongResponse)
	if err != nil {
		t.Fatalf("Failed to read pong response: %v", err)
	}

	if pongResponse.Type != MsgTypePong {
		t.Errorf("Expected pong message, got %s", pongResponse.Type)
	}

	t.Log("Ping-pong test passed")
}

func TestWebSocketErrorHandling(t *testing.T) {
	// Create test server
	config := &core.Config{}

	sessionManager := NewSessionManager(config)
	connManager := NewConnectionManager(sessionManager)
	connManager.Start()
	defer connManager.Stop()

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(connManager.HandleWebSocket))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Send invalid JSON
	err = conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
	if err != nil {
		t.Fatalf("Failed to send invalid message: %v", err)
	}

	// Read welcome message first
	var welcomeMsg WebSocketMessage
	err = conn.ReadJSON(&welcomeMsg)
	if err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	// Read error response
	var errorResponse WebSocketMessage
	err = conn.ReadJSON(&errorResponse)
	if err != nil {
		t.Fatalf("Failed to read error response: %v", err)
	}

	if errorResponse.Type != MsgTypeError {
		t.Errorf("Expected error message, got %s", errorResponse.Type)
	}

	t.Log("Error handling test passed")
}

func TestConnectionManager(t *testing.T) {
	config := &core.Config{}

	sessionManager := NewSessionManager(config)
	connManager := NewConnectionManager(sessionManager)

	// Test initial state
	if connManager.GetConnectionCount() != 0 {
		t.Errorf("Expected 0 connections, got %d", connManager.GetConnectionCount())
	}

	// Test session list
	sessions := connManager.GetConnectedSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}

	// Start and stop manager
	connManager.Start()
	connManager.Stop()

	t.Log("Connection manager test passed")
}

func TestMessageCreationHelpers(t *testing.T) {
	sessionID := "test-session"

	// Test chat message
	chatMsg := NewChatMessage(sessionID, "Hello")
	if chatMsg.Type != MsgTypeChatMessage {
		t.Errorf("Expected chat message type, got %s", chatMsg.Type)
	}
	if chatMsg.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, chatMsg.SessionID)
	}

	// Test agent response
	agentMsg := NewAgentResponse(sessionID, "TestAgent", "Hi there", "complete")
	if agentMsg.Type != MsgTypeAgentResponse {
		t.Errorf("Expected agent response type, got %s", agentMsg.Type)
	}

	// Test error message
	errorMsg := NewWSErrorMessage(sessionID, "TEST_ERROR", "Test error", "Details")
	if errorMsg.Type != MsgTypeError {
		t.Errorf("Expected error message type, got %s", errorMsg.Type)
	}

	// Test system message
	systemMsg := NewSystemMessage(sessionID, "info", "System message")
	if systemMsg.Type != MsgTypeSystemMessage {
		t.Errorf("Expected system message type, got %s", systemMsg.Type)
	}

	// Test pong message
	pongMsg := NewPong(sessionID)
	if pongMsg.Type != MsgTypePong {
		t.Errorf("Expected pong message type, got %s", pongMsg.Type)
	}

	t.Log("Message creation helpers test passed")
}

func TestMessageParsing(t *testing.T) {
	// Test chat message parsing
	chatMsg := &WebSocketMessage{
		Type: MsgTypeChatMessage,
		Data: map[string]interface{}{
			"content":      "Hello",
			"message_type": "text",
		},
	}

	chatData, err := chatMsg.ParseChatMessage()
	if err != nil {
		t.Fatalf("Failed to parse chat message: %v", err)
	}

	if chatData.Content != "Hello" {
		t.Errorf("Expected content 'Hello', got '%s'", chatData.Content)
	}

	// Test session create parsing
	sessionMsg := &WebSocketMessage{
		Type: MsgTypeSessionCreate,
		Data: map[string]interface{}{
			"user_agent": "test-agent",
		},
	}

	sessionData, err := sessionMsg.ParseSessionCreate()
	if err != nil {
		t.Fatalf("Failed to parse session create message: %v", err)
	}

	if sessionData.UserAgent != "test-agent" {
		t.Errorf("Expected user agent 'test-agent', got '%s'", sessionData.UserAgent)
	}

	t.Log("Message parsing test passed")
}

func TestMessageValidation(t *testing.T) {
	// Test valid message
	validMsg := &WebSocketMessage{
		Type:      MsgTypeChatMessage,
		MessageID: "test-id",
		Timestamp: time.Now(),
	}

	err := validMsg.Validate()
	if err != nil {
		t.Errorf("Valid message failed validation: %v", err)
	}

	// Test invalid message (no type)
	invalidMsg := &WebSocketMessage{}
	err = invalidMsg.Validate()
	if err == nil {
		t.Error("Expected validation error for message without type")
	}

	// Test message with auto-generated fields
	autoMsg := &WebSocketMessage{
		Type: MsgTypePing,
	}

	err = autoMsg.Validate()
	if err != nil {
		t.Errorf("Auto-generation failed: %v", err)
	}

	if autoMsg.MessageID == "" {
		t.Error("Expected auto-generated message ID")
	}

	if autoMsg.Timestamp.IsZero() {
		t.Error("Expected auto-generated timestamp")
	}

	t.Log("Message validation test passed")
}
