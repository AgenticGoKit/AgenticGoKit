package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	sessionManager    *EnhancedSessionManager
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

	// Create enhanced session manager
	sessionConfig := DefaultSessionConfig()
	sessionManager, err := NewEnhancedSessionManager(config.Config, sessionConfig)
	if err != nil {
		// Fallback to basic configuration if enhanced setup fails
		sessionManager, _ = NewEnhancedSessionManager(&core.Config{}, DefaultSessionConfig())
	}

	// Create logger
	logger := core.Logger()

	// Create connection manager with enhanced session manager
	connectionManager := NewConnectionManager(sessionManager, logger)

	server := &Server{
		port:              config.Port,
		config:            config.Config,
		agentManager:      config.AgentManager,
		sessionManager:    sessionManager,
		connectionManager: connectionManager,
		logger:            logger,
	}

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Static file serving
	log.Printf("DEBUG: StaticDir configured as: %s", config.StaticDir)
	absStaticDir, _ := filepath.Abs(config.StaticDir)
	log.Printf("DEBUG: StaticDir absolute path: %s", absStaticDir)

	staticHandler := http.FileServer(http.Dir(config.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))

	// Root handler serves the main chat interface by redirecting to static/index.html
	mux.HandleFunc("/", server.handleRoot)
	log.Printf("DEBUG: Root handler registered for path '/'")

	// API endpoints
	mux.HandleFunc("/api/health", server.handleHealth)
	mux.HandleFunc("/api/config", server.handleConfig)
	mux.HandleFunc("/api/agents", server.handleAgents)
	mux.HandleFunc("/api/chat", server.handleChat)

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
	log.Printf("DEBUG: handleRoot called - Method: %s, Path: %s", r.Method, r.URL.Path)

	if r.URL.Path != "/" {
		log.Printf("DEBUG: Path is not '/', returning 404 for path: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	log.Printf("DEBUG: handleRoot called for path: %s", r.URL.Path)

	// Get current working directory for debugging
	cwd, _ := os.Getwd()
	log.Printf("DEBUG: Current working directory: %s", cwd)

	// Try multiple paths to find index.html
	possiblePaths := []string{
		"../../internal/webui/static/index.html",    // Relative to examples directory
		"internal/webui/static/index.html",          // From project root
		"./internal/webui/static/index.html",        // With explicit current dir
		"../../../internal/webui/static/index.html", // In case we're deeper
	}

	var indexPath string
	var found bool

	for _, path := range possiblePaths {
		absPath, _ := filepath.Abs(path)
		log.Printf("DEBUG: Trying path: %s (absolute: %s)", path, absPath)

		if _, err := os.Stat(path); err == nil {
			indexPath = path
			found = true
			log.Printf("DEBUG: SUCCESS - Found file at: %s", path)
			break
		} else {
			log.Printf("DEBUG: File not found at %s: %v", path, err)
		}
	}

	if !found {
		log.Printf("ERROR: Could not find index.html in any of the tried paths")
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}

	log.Printf("DEBUG: Serving file: %s", indexPath)
	http.ServeFile(w, r, indexPath)
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
		"available": []string{},
		"total":     0,
		"status":    "not_integrated",
	}

	// Debug: log agent manager status
	if s.logger != nil {
		s.logger.Debug().
			Bool("agent_manager_nil", s.agentManager == nil).
			Msg("handleAgents: checking agent manager")
	}

	// Integrate with actual agent manager
	if s.agentManager != nil {
		agents["status"] = "integrated"

		// Get active agents from agent manager
		activeAgents := s.agentManager.GetActiveAgents()

		// Debug: log agent count
		if s.logger != nil {
			s.logger.Debug().
				Int("active_agents_count", len(activeAgents)).
				Msg("handleAgents: got active agents")
		}

		agentList := make([]map[string]interface{}, 0, len(activeAgents))

		for _, agent := range activeAgents {
			agentInfo := map[string]interface{}{
				"name":         agent.Name(),
				"description":  agent.GetDescription(),
				"role":         agent.GetRole(),
				"capabilities": agent.GetCapabilities(),
				"enabled":      agent.IsEnabled(),
			}
			agentList = append(agentList, agentInfo)
		}

		agents["available"] = agentList
		agents["total"] = len(activeAgents)
	}

	s.writeJSON(w, agents)
}

// handleChat handles chat interactions with agents
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var chatReq struct {
		AgentName string `json:"agent_name"`
		Message   string `json:"message"`
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if chatReq.AgentName == "" || chatReq.Message == "" {
		http.Error(w, "agent_name and message are required", http.StatusBadRequest)
		return
	}

	// Handle chat interaction
	var agentResponse string = "Hello! This is a mock response from " + chatReq.AgentName + ". The actual agent integration is in development."

	// If agent manager is available, try to use it
	if s.agentManager != nil {
		log.Printf("DEBUG: Chat request - Agent: %s, Message: %s", chatReq.AgentName, chatReq.Message)

		// Get all agents to find the requested one
		agents := s.agentManager.GetActiveAgents()
		var targetAgent core.Agent

		for _, agent := range agents {
			if agent.Name() == chatReq.AgentName {
				targetAgent = agent
				break
			}
		}

		if targetAgent != nil {
			log.Printf("DEBUG: Found agent %s, calling HandleEvent", targetAgent.Name())

			// Create event for the agent
			eventData := map[string]any{
				"message":    chatReq.Message,
				"session_id": chatReq.SessionID,
				"timestamp":  time.Now().Format(time.RFC3339),
			}
			metadata := map[string]string{
				"type": "chat_message",
			}
			event := core.NewEvent(targetAgent.Name(), eventData, metadata)

			// Create input state
			state := core.NewSimpleState(map[string]any{
				"message":    chatReq.Message,
				"session_id": chatReq.SessionID,
			})

			// Call the agent
			ctx := context.Background()
			result, err := targetAgent.HandleEvent(ctx, event, state)
			if err != nil {
				log.Printf("ERROR: Agent HandleEvent failed: %v", err)
				agentResponse = fmt.Sprintf("Sorry, I encountered an error while processing your message: %v", err)
			} else {
				// Extract response from agent result
				if responseMsg, exists := result.OutputState.Get("message"); exists {
					if msgStr, ok := responseMsg.(string); ok {
						agentResponse = msgStr
						log.Printf("DEBUG: Agent response: %s", agentResponse)
					}
				}
			}
		} else {
			log.Printf("DEBUG: Agent %s not found", chatReq.AgentName)
			agentResponse = fmt.Sprintf("Sorry, I couldn't find agent '%s'. Available agents: %v", chatReq.AgentName, func() []string {
				names := make([]string, len(agents))
				for i, a := range agents {
					names[i] = a.Name()
				}
				return names
			}())
		}
	}

	responseData := map[string]interface{}{
		"agent":     chatReq.AgentName,
		"message":   chatReq.Message,
		"response":  agentResponse,
		"timestamp": time.Now().Unix(),
	}

	log.Printf("DEBUG: Sending response - agent: %s, response length: %d",
		chatReq.AgentName, len(agentResponse))
	log.Printf("DEBUG: Response content preview: %.100s...", agentResponse)

	s.writeJSON(w, responseData)
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
	// Parse pagination parameters
	offset := 0
	limit := 50 // Default limit

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Get sessions from enhanced session manager
	sessions, total, err := s.sessionManager.ListSessions(offset, limit)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list sessions")
		http.Error(w, "Failed to list sessions", http.StatusInternalServerError)
		return
	}

	sessionData := make([]map[string]interface{}, len(sessions))
	for i, session := range sessions {
		sessionData[i] = map[string]interface{}{
			"id":         session.ID,
			"created_at": session.CreatedAt,
			"last_used":  session.LastUsed,
			"messages":   len(session.Messages),
			"active":     session.Active,
			"user_agent": session.UserAgent,
		}
	}

	response := map[string]interface{}{
		"sessions": sessionData,
		"total":    total,
		"offset":   offset,
		"limit":    limit,
	}

	s.writeJSON(w, response)
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	// Get user agent from request
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		userAgent = "Unknown"
	}

	// Get client IP
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Real-IP")
	}
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	// Create session using enhanced session manager
	session, err := s.sessionManager.CreateSession(userAgent, clientIP)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create session")
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":         session.ID,
		"created_at": session.CreatedAt,
		"status":     "created",
	}

	s.writeJSON(w, response)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	// Get session from enhanced session manager
	session, err := s.sessionManager.GetSession(sessionID)
	if err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Session not found")
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"id":         session.ID,
		"created_at": session.CreatedAt,
		"last_used":  session.LastUsed,
		"messages":   session.Messages,
		"active":     session.Active,
		"user_agent": session.UserAgent,
		"ip_address": session.IPAddress,
	}

	s.writeJSON(w, response)
}

func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	// Delete session using enhanced session manager
	err := s.sessionManager.DeleteSession(sessionID)
	if err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to delete session")
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"id":     sessionID,
		"status": "deleted",
	}

	s.writeJSON(w, response)
}

// startSessionCleanup starts a background routine to clean up expired sessions
func (s *Server) startSessionCleanup(ctx context.Context) {
	// Enhanced session manager handles its own cleanup routines
	// This method is kept for compatibility but delegates to the session manager
	s.logger.Info().Msg("Session cleanup routine delegated to enhanced session manager")
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
