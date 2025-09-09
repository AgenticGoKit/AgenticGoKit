package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// Server represents the WebUI HTTP server with WebSocket support
type Server struct {
	port              string
	server            *http.Server
	config            *core.Config
	agentManager      core.AgentManager
	sessions          map[string]*ChatSession
	sessionMutex      sync.RWMutex
	sessionManager    *SessionManager
	connectionManager *ConnectionManager
	logger            core.CoreLogger

	// Server state
	started bool
	mutex   sync.Mutex
}

// ServerConfig contains configuration for the WebUI server
type ServerConfig struct {
	Port         string
	StaticDir    string
	Config       *core.Config
	AgentManager core.AgentManager
}

// NewServer creates a new WebUI server instance with WebSocket support
func NewServer(config ServerConfig) *Server {
	if config.Port == "" {
		config.Port = "8080"
	}

	if config.StaticDir == "" {
		config.StaticDir = "./internal/webui/static"
	}

	// Create session manager
	sessionManager := NewSessionManager(config.Config)

	// Create connection manager
	connectionManager := NewConnectionManager(sessionManager)

	server := &Server{
		port:              config.Port,
		config:            config.Config,
		agentManager:      config.AgentManager,
		sessions:          make(map[string]*ChatSession),
		sessionManager:    sessionManager,
		connectionManager: connectionManager,
		logger:            core.Logger(),
	}

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Static file serving
	staticHandler := http.FileServer(http.Dir(config.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))

	// Root handler serves the main chat interface
	mux.HandleFunc("/", server.handleRoot)

	// API endpoints
	mux.HandleFunc("/api/health", server.handleHealth)
	mux.HandleFunc("/api/config", server.handleConfig)
	mux.HandleFunc("/api/agents", server.handleAgents)

	// Session management endpoints
	mux.HandleFunc("/api/sessions", server.handleSessions)
	mux.HandleFunc("/api/sessions/", server.handleSessionDetails)

	// WebSocket endpoint
	mux.HandleFunc("/ws", server.connectionManager.HandleWebSocket)

	// Create HTTP server with middleware
	handler := server.withMiddleware(mux)

	server.server = &http.Server{
		Addr:         ":" + config.Port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return server
}

// Start starts the WebUI server
func (s *Server) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.started {
		return fmt.Errorf("server already started")
	}

	s.logger.Info().
		Str("port", s.port).
		Str("addr", "http://localhost:"+s.port).
		Msg("Starting WebUI server")

	// Start connection manager
	s.connectionManager.Start()

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Start session cleanup routine
	go s.startSessionCleanup(ctx)

	s.started = true

	// Handle context cancellation and graceful shutdown
	go func() {
		select {
		case <-ctx.Done():
			s.logger.Info().Msg("Shutting down WebUI server due to context cancellation")
			s.Stop()
		case err := <-errChan:
			s.logger.Error().Err(err).Msg("Server error")
		}
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	s.logger.Info().
		Str("url", "http://localhost:"+s.port).
		Msg("WebUI server started successfully")

	return nil
}

// Stop gracefully stops the WebUI server
func (s *Server) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.started {
		return nil
	}

	s.logger.Info().Msg("Stopping WebUI server")

	// Stop connection manager
	s.connectionManager.Stop()

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Graceful shutdown
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Error during server shutdown")
		return err
	}

	s.started = false
	s.logger.Info().Msg("WebUI server stopped successfully")
	return nil
}

// IsStarted returns whether the server is currently running
func (s *Server) IsStarted() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.started
}

// GetURL returns the server URL
func (s *Server) GetURL() string {
	return "http://localhost:" + s.port
}

// handleRoot serves the main chat interface
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// For now, serve a simple HTML page
	// This will be replaced with a proper template in Task 5
	html := `<!DOCTYPE html>
<html>
<head>
    <title>AgenticGoKit Chat Interface</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 20px; }
        .status { padding: 10px; background: #e8f5e8; border-radius: 4px; margin-bottom: 20px; }
        .info { margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 AgenticGoKit Web Interface</h1>
            <p>WebUI Server is Running Successfully</p>
        </div>
        
        <div class="status">
            <strong>✅ Server Status: Running</strong>
        </div>
        
        <div class="info">
            <h3>Server Information</h3>
            <ul>
                <li><strong>Port:</strong> ` + s.port + `</li>
                <li><strong>URL:</strong> <a href="` + s.GetURL() + `">` + s.GetURL() + `</a></li>
                <li><strong>API Health:</strong> <a href="/api/health">/api/health</a></li>
                <li><strong>Config:</strong> <a href="/api/config">/api/config</a></li>
                <li><strong>Agents:</strong> <a href="/api/agents">/api/agents</a></li>
            </ul>
        </div>
        
        <div class="info">
            <h3>Next Steps</h3>
            <p>This is a basic server setup for Task 1. The chat interface will be implemented in Task 5.</p>
            <p>WebSocket communication will be added in Task 2.</p>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// handleHealth provides a health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"server":    "webui",
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now().Add(-time.Hour)).String(), // Placeholder
	}

	s.writeJSON(w, health)
}

// handleConfig returns server configuration information
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := map[string]interface{}{
		"server": map[string]interface{}{
			"port": s.port,
			"url":  s.GetURL(),
		},
		"features": map[string]interface{}{
			"websocket": false, // Will be true after Task 2
			"sessions":  true,
			"agents":    s.agentManager != nil,
		},
	}

	if s.config != nil {
		config["agentflow"] = map[string]interface{}{
			"name":     s.config.AgentFlow.Name,
			"version":  s.config.AgentFlow.Version,
			"provider": s.config.AgentFlow.Provider,
		}
	}

	s.writeJSON(w, config)
}

// handleAgents returns available agents information
func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agents := map[string]interface{}{
		"available": []string{}, // Will be populated when agent manager is integrated
		"total":     0,
		"status":    "not_integrated", // Will change in Task 4
	}

	// TODO: Integrate with actual agent manager in Task 4
	if s.agentManager != nil {
		agents["status"] = "integrated"
		// agents["available"] = s.agentManager.GetActiveAgents()
	}

	s.writeJSON(w, agents)
}

// handleSessions handles session management endpoints
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetSessions(w, r)
	case http.MethodPost:
		s.handleCreateSession(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSessionDetails handles individual session operations
func (s *Server) handleSessionDetails(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from URL path
	sessionID := filepath.Base(r.URL.Path)
	if sessionID == "sessions" || sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetSession(w, r, sessionID)
	case http.MethodDelete:
		s.handleDeleteSession(w, r, sessionID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Helper methods for session management (stubs for now, will be implemented in Task 3)

func (s *Server) handleGetSessions(w http.ResponseWriter, r *http.Request) {
	s.sessionMutex.RLock()
	defer s.sessionMutex.RUnlock()

	sessions := make([]map[string]interface{}, 0, len(s.sessions))
	for id, session := range s.sessions {
		sessions = append(sessions, map[string]interface{}{
			"id":         id,
			"created_at": session.CreatedAt,
			"last_used":  session.LastUsed,
			"messages":   len(session.Messages),
		})
	}

	response := map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	}

	s.writeJSON(w, response)
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	// Stub implementation - will be completed in Task 3
	sessionID := generateSessionID()

	session := &ChatSession{
		ID:        sessionID,
		Messages:  make([]ChatMessage, 0),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	s.sessionMutex.Lock()
	s.sessions[sessionID] = session
	s.sessionMutex.Unlock()

	response := map[string]interface{}{
		"id":         sessionID,
		"created_at": session.CreatedAt,
		"status":     "created",
	}

	s.writeJSON(w, response)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	s.sessionMutex.RLock()
	session, exists := s.sessions[sessionID]
	s.sessionMutex.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"id":         session.ID,
		"created_at": session.CreatedAt,
		"last_used":  session.LastUsed,
		"messages":   session.Messages,
	}

	s.writeJSON(w, response)
}

func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	delete(s.sessions, sessionID)

	response := map[string]interface{}{
		"id":     sessionID,
		"status": "deleted",
	}

	s.writeJSON(w, response)
}

// startSessionCleanup starts a background routine to clean up expired sessions
func (s *Server) startSessionCleanup(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute) // Clean up every 30 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanupExpiredSessions()
		}
	}
}

// cleanupExpiredSessions removes sessions that haven't been used for more than 24 hours
func (s *Server) cleanupExpiredSessions() {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()

	expiration := time.Now().Add(-24 * time.Hour)
	var expired []string

	for id, session := range s.sessions {
		if session.LastUsed.Before(expiration) {
			expired = append(expired, id)
		}
	}

	for _, id := range expired {
		delete(s.sessions, id)
	}

	if len(expired) > 0 {
		s.logger.Info().
			Int("count", len(expired)).
			Msg("Cleaned up expired sessions")
	}
}

// writeJSON writes a JSON response
func (s *Server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status": "success",
		"data":   data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
