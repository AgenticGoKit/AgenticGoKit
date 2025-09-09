package webui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	config := ServerConfig{
		Port:      "8080",
		StaticDir: "./static",
	}

	server := NewServer(config)

	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.port != "8080" {
		t.Errorf("Expected port 8080, got %s", server.port)
	}

	if server.GetURL() != "http://localhost:8080" {
		t.Errorf("Expected URL http://localhost:8080, got %s", server.GetURL())
	}
}

func TestHealthEndpoint(t *testing.T) {
	config := ServerConfig{
		Port:      "8080",
		StaticDir: "./static",
	}

	server := NewServer(config)

	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	// Check that response contains expected content
	body := rr.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}

	// Check content type
	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, contentType)
	}
}

func TestConfigEndpoint(t *testing.T) {
	config := ServerConfig{
		Port:      "8080",
		StaticDir: "./static",
	}

	server := NewServer(config)

	req, err := http.NewRequest("GET", "/api/config", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	body := rr.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestSessionManagement(t *testing.T) {
	config := ServerConfig{
		Port:      "8080",
		StaticDir: "./static",
	}

	server := NewServer(config)

	// Test creating a session
	req, err := http.NewRequest("POST", "/api/sessions", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	// Test getting sessions
	req, err = http.NewRequest("GET", "/api/sessions", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
}

func TestCORSHeaders(t *testing.T) {
	config := ServerConfig{
		Port:      "8080",
		StaticDir: "./static",
	}

	server := NewServer(config)

	req, err := http.NewRequest("OPTIONS", "/api/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, status)
	}

	// Check CORS headers
	allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin == "" {
		t.Error("Expected Access-Control-Allow-Origin header to be set")
	}

	allowMethods := rr.Header().Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Error("Expected Access-Control-Allow-Methods header to be set")
	}
}

func TestSecurityHeaders(t *testing.T) {
	config := ServerConfig{
		Port:      "8080",
		StaticDir: "./static",
	}

	server := NewServer(config)

	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rr, req)

	// Check security headers
	headers := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Content-Security-Policy",
	}

	for _, header := range headers {
		if value := rr.Header().Get(header); value == "" {
			t.Errorf("Expected %s header to be set", header)
		}
	}
}

func TestServerStartStop(t *testing.T) {
	config := ServerConfig{
		Port:      "8081", // Use different port to avoid conflicts
		StaticDir: "./static",
	}

	server := NewServer(config)

	if server.IsStarted() {
		t.Error("Expected server to not be started initially")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	if !server.IsStarted() {
		t.Error("Expected server to be started")
	}

	// Stop server
	err = server.Stop()
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	if server.IsStarted() {
		t.Error("Expected server to be stopped")
	}
}

func TestSessionLifecycle(t *testing.T) {
	session := NewChatSession()

	if session.ID == "" {
		t.Error("Expected session ID to be generated")
	}

	if session.GetMessageCount() != 0 {
		t.Error("Expected new session to have 0 messages")
	}

	// Add a user message
	userMsg := NewUserMessage("Hello, world!")
	session.AddMessage(userMsg)

	if session.GetMessageCount() != 1 {
		t.Error("Expected session to have 1 message after adding")
	}

	lastMsg := session.GetLastMessage()
	if lastMsg == nil || lastMsg.Content != "Hello, world!" {
		t.Error("Expected last message to match added message")
	}

	// Add an agent message
	agentMsg := NewAgentMessage("test-agent", "Hello back!")
	session.AddMessage(agentMsg)

	if session.GetMessageCount() != 2 {
		t.Error("Expected session to have 2 messages")
	}

	// Test filtering by role
	userMsgs := session.GetMessagesByRole("user")
	if len(userMsgs) != 1 {
		t.Error("Expected 1 user message")
	}

	agentMsgs := session.GetMessagesByRole("agent")
	if len(agentMsgs) != 1 {
		t.Error("Expected 1 agent message")
	}
}

func TestSessionExpiration(t *testing.T) {
	session := NewChatSession()

	// New session should not be expired
	if session.IsExpired() {
		t.Error("Expected new session to not be expired")
	}

	// Manually set last used to past
	session.LastUsed = time.Now().Add(-25 * time.Hour)

	if !session.IsExpired() {
		t.Error("Expected old session to be expired")
	}

	// Touch should update last used
	session.Touch()

	if session.IsExpired() {
		t.Error("Expected touched session to not be expired")
	}
}

func TestMessageGeneration(t *testing.T) {
	userMsg := NewUserMessage("Test content")
	if userMsg.Role != "user" {
		t.Error("Expected user message role to be 'user'")
	}
	if userMsg.Content != "Test content" {
		t.Error("Expected message content to match")
	}
	if userMsg.Status != "complete" {
		t.Error("Expected user message status to be 'complete'")
	}

	agentMsg := NewAgentMessage("test-agent", "Agent response")
	if agentMsg.Role != "agent" {
		t.Error("Expected agent message role to be 'agent'")
	}
	if agentMsg.AgentName != "test-agent" {
		t.Error("Expected agent name to match")
	}

	processingMsg := NewProcessingMessage("test-agent")
	if processingMsg.Status != "processing" {
		t.Error("Expected processing message status to be 'processing'")
	}

	errorMsg := NewErrorMessage("test-agent", "Test error")
	if errorMsg.Status != "error" {
		t.Error("Expected error message status to be 'error'")
	}
	if errorMsg.Error != "Test error" {
		t.Error("Expected error text to match")
	}
}
