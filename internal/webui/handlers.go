package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// AgentHTTPHandlers provides HTTP handlers for agent-related operations
type AgentHTTPHandlers struct {
	bridge         *AgentBridge
	sessionManager *EnhancedSessionManager
	logger         core.CoreLogger
}

// NewAgentHTTPHandlers creates new agent HTTP handlers
func NewAgentHTTPHandlers(bridge *AgentBridge, sessionManager *EnhancedSessionManager, logger core.CoreLogger) *AgentHTTPHandlers {
	return &AgentHTTPHandlers{
		bridge:         bridge,
		sessionManager: sessionManager,
		logger:         logger,
	}
}

// ChatRequest represents a chat request
type ChatRequest struct {
	SessionID   string                 `json:"session_id"`
	Message     string                 `json:"message"`
	AgentName   string                 `json:"agent_name,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	SessionID   string                 `json:"session_id"`
	AgentName   string                 `json:"agent_name"`
	Content     string                 `json:"content"`
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	IsStreaming bool                   `json:"is_streaming"`
	ChunkIndex  int                    `json:"chunk_index,omitempty"`
	TotalChunks int                    `json:"total_chunks,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   int64                  `json:"timestamp"`
}

// AgentsListResponse represents the response for listing agents
type AgentsListResponse struct {
	Agents    []AgentInfo `json:"agents"`
	Count     int         `json:"count"`
	Timestamp int64       `json:"timestamp"`
}

// SessionInfoResponse represents session information
type SessionInfoResponse struct {
	SessionID    string                 `json:"session_id"`
	CreatedAt    time.Time              `json:"created_at"`
	LastActivity time.Time              `json:"last_activity"`
	MessageCount int                    `json:"message_count"`
	IsActive     bool                   `json:"is_active"`
	UserID       string                 `json:"user_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// HandleChatRequest handles chat requests via HTTP
func (h *AgentHTTPHandlers) HandleChatRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode chat request")
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON request body")
		return
	}

	// Validate request
	if req.Message == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "EMPTY_MESSAGE", "Message cannot be empty")
		return
	}

	if req.SessionID == "" {
		req.SessionID = generateSessionID()
	}

	if req.UserID == "" {
		req.UserID = "anonymous"
	}

	h.logger.Info().
		Str("session_id", req.SessionID).
		Str("user_id", req.UserID).
		Str("agent_name", req.AgentName).
		Bool("stream", req.Stream).
		Msg("Processing chat request")

	// Add user message to session
	err := h.sessionManager.AddMessage(req.SessionID, ChatMessage{
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
		Metadata:  req.Metadata,
	})
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", req.SessionID).Msg("Failed to add user message to session")
	}

	// Process with agent bridge
	err = h.bridge.ProcessChatMessage(r.Context(), req.SessionID, req.Message, req.Metadata)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to process chat message with agent bridge")
		h.writeErrorResponse(w, http.StatusInternalServerError, "PROCESSING_ERROR", "Failed to process message with agent")
		return
	}

	// Get response stream for this session
	responseChan := h.bridge.GetResponseStream(req.SessionID)

	// Handle streaming vs non-streaming response
	if req.Stream {
		h.handleStreamingResponse(w, r, responseChan)
	} else {
		h.handleNonStreamingResponse(w, r, responseChan)
	}
}

// handleStreamingResponse handles streaming chat responses
func (h *AgentHTTPHandlers) handleStreamingResponse(w http.ResponseWriter, r *http.Request, responseChan <-chan *AgentResponse) {
	// Set headers for streaming
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeErrorResponse(w, http.StatusInternalServerError, "STREAMING_ERROR", "Streaming not supported")
		return
	}

	// Send responses as they come
	for {
		select {
		case <-r.Context().Done():
			return
		case response, ok := <-responseChan:
			if !ok {
				// Channel closed
				return
			}

			// Convert to chat response
			chatResponse := ChatResponse{
				SessionID:   response.SessionID,
				AgentName:   response.AgentName,
				Content:     response.Content,
				Status:      response.Status,
				Error:       response.Error,
				IsStreaming: response.IsStreaming,
				ChunkIndex:  response.ChunkIndex,
				TotalChunks: response.TotalChunks,
				Metadata:    response.Metadata,
				Timestamp:   response.Timestamp.Unix(),
			}

			// Write JSON response
			responseBytes, err := json.Marshal(chatResponse)
			if err != nil {
				h.logger.Error().Err(err).Msg("Failed to marshal streaming response")
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", responseBytes)
			flusher.Flush()

			// Add assistant message to session for successful responses
			if response.Status == "complete" && response.Error == "" {
				err := h.sessionManager.AddMessage(response.SessionID, ChatMessage{
					Role:      "assistant",
					Content:   response.Content,
					Timestamp: response.Timestamp,
					Metadata:  response.Metadata,
				})
				if err != nil {
					h.logger.Error().Err(err).Str("session_id", response.SessionID).Msg("Failed to add assistant message to session")
				}
			}
		}
	}
}

// handleNonStreamingResponse handles non-streaming chat responses
func (h *AgentHTTPHandlers) handleNonStreamingResponse(w http.ResponseWriter, r *http.Request, responseChan <-chan *AgentResponse) {
	var finalResponse *AgentResponse

	// Collect all responses
	for {
		select {
		case <-r.Context().Done():
			h.writeErrorResponse(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request cancelled")
			return
		case response, ok := <-responseChan:
			if !ok {
				// Channel closed
				break
			}

			// Keep the last response
			finalResponse = response

			// Break on final response
			if response.Status == "complete" || response.Status == "error" {
				break
			}
		}
	}

	if finalResponse == nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "NO_RESPONSE", "No response received from agent")
		return
	}

	// Convert to chat response
	chatResponse := ChatResponse{
		SessionID:   finalResponse.SessionID,
		AgentName:   finalResponse.AgentName,
		Content:     finalResponse.Content,
		Status:      finalResponse.Status,
		Error:       finalResponse.Error,
		IsStreaming: false,
		Metadata:    finalResponse.Metadata,
		Timestamp:   finalResponse.Timestamp.Unix(),
	}

	// Add assistant message to session for successful responses
	if finalResponse.Status == "complete" && finalResponse.Error == "" {
		err := h.sessionManager.AddMessage(finalResponse.SessionID, ChatMessage{
			Role:      "assistant",
			Content:   finalResponse.Content,
			Timestamp: finalResponse.Timestamp,
			Metadata:  finalResponse.Metadata,
		})
		if err != nil {
			h.logger.Error().Err(err).Str("session_id", finalResponse.SessionID).Msg("Failed to add assistant message to session")
		}
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(chatResponse); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode chat response")
		h.writeErrorResponse(w, http.StatusInternalServerError, "ENCODING_ERROR", "Failed to encode response")
	}
}

// HandleGetAgents handles requests for available agents
func (h *AgentHTTPHandlers) HandleGetAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	agents := h.bridge.GetAvailableAgents()

	response := AgentsListResponse{
		Agents:    agents,
		Count:     len(agents),
		Timestamp: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode agents list response")
		h.writeErrorResponse(w, http.StatusInternalServerError, "ENCODING_ERROR", "Failed to encode response")
	}
}

// HandleGetSessionInfo handles requests for session information
func (h *AgentHTTPHandlers) HandleGetSessionInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_SESSION_ID", "Session ID is required")
		return
	}

	session, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get session")
		h.writeErrorResponse(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found")
		return
	}

	response := SessionInfoResponse{
		SessionID:    session.ID,
		CreatedAt:    session.CreatedAt,
		LastActivity: session.LastUsed,
		MessageCount: len(session.Messages),
		IsActive:     time.Since(session.LastUsed) < 30*time.Minute,
		UserID:       session.UserAgent,            // Use UserAgent as UserID
		Metadata:     make(map[string]interface{}), // Empty metadata for now
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode session info response")
		h.writeErrorResponse(w, http.StatusInternalServerError, "ENCODING_ERROR", "Failed to encode response")
	}
}

// HandleDeleteSession handles session deletion requests
func (h *AgentHTTPHandlers) HandleDeleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only DELETE method is allowed")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_SESSION_ID", "Session ID is required")
		return
	}

	err := h.sessionManager.DeleteSession(sessionID)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to delete session")
		h.writeErrorResponse(w, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete session")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Session deleted successfully",
		"timestamp": time.Now().Unix(),
	})
}

// HandleGetSessionMessages handles requests for session messages
func (h *AgentHTTPHandlers) HandleGetSessionMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_SESSION_ID", "Session ID is required")
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	session, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get session messages")
		h.writeErrorResponse(w, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch session messages")
		return
	}

	// Get messages with pagination
	messages := session.Messages
	totalMessages := len(messages)

	// Apply offset
	if offset >= totalMessages {
		messages = []ChatMessage{}
	} else {
		messages = messages[offset:]
	}

	// Apply limit
	if limit > 0 && len(messages) > limit {
		messages = messages[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"messages":  messages,
		"count":     len(messages),
		"limit":     limit,
		"offset":    offset,
		"timestamp": time.Now().Unix(),
	}); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode session messages response")
		h.writeErrorResponse(w, http.StatusInternalServerError, "ENCODING_ERROR", "Failed to encode response")
	}
}

// HandleHealthCheck handles health check requests
func (h *AgentHTTPHandlers) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Check bridge health
	bridgeHealthy := h.bridge != nil

	// Check session manager health
	sessionHealthy := h.sessionManager != nil

	status := "healthy"
	statusCode := http.StatusOK

	if !bridgeHealthy || !sessionHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().Unix(),
		"components": map[string]bool{
			"bridge":          bridgeHealthy,
			"session_manager": sessionHealthy,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse writes an error response
func (h *AgentHTTPHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	response := ErrorResponse{
		Error:     http.StatusText(statusCode),
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// RegisterHandlers registers all agent HTTP handlers with a mux
func (h *AgentHTTPHandlers) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/chat", h.HandleChatRequest)
	mux.HandleFunc("/api/agents", h.HandleGetAgents)
	mux.HandleFunc("/api/session/info", h.HandleGetSessionInfo)
	mux.HandleFunc("/api/session/delete", h.HandleDeleteSession)
	mux.HandleFunc("/api/session/messages", h.HandleGetSessionMessages)
	mux.HandleFunc("/api/health", h.HandleHealthCheck)
	// Note: WebSocket handler will be registered separately through ConnectionManager
}
