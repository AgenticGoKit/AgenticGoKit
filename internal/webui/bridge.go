package webui

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// AgentBridge provides the interface between WebUI and the core agent system
type AgentBridge struct {
	agentManager   core.AgentManager
	sessionManager *EnhancedSessionManager
	logger         core.CoreLogger

	// Event routing and streaming
	eventHandlers   map[string]AgentEventHandler
	responseStreams map[string]chan *AgentResponse
	streamMutex     sync.RWMutex

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Configuration
	config *BridgeConfig
}

// BridgeConfig contains configuration options for the agent bridge
type BridgeConfig struct {
	// Timeouts and limits
	AgentTimeout       time.Duration `json:"agent_timeout"`
	ResponseBufferSize int           `json:"response_buffer_size"`
	MaxConcurrentTasks int           `json:"max_concurrent_tasks"`

	// Error handling
	RetryAttempts int           `json:"retry_attempts"`
	RetryDelay    time.Duration `json:"retry_delay"`

	// Streaming options
	StreamingEnabled bool `json:"streaming_enabled"`
	ChunkSize        int  `json:"chunk_size"`
}

// DefaultBridgeConfig returns sensible default configuration
func DefaultBridgeConfig() *BridgeConfig {
	return &BridgeConfig{
		AgentTimeout:       30 * time.Second,
		ResponseBufferSize: 100,
		MaxConcurrentTasks: 10,
		RetryAttempts:      3,
		RetryDelay:         1 * time.Second,
		StreamingEnabled:   true,
		ChunkSize:          1024,
	}
}

// AgentEventHandler defines the interface for handling agent events
type AgentEventHandler interface {
	HandleEvent(ctx context.Context, event *WebUIEvent) (*AgentResponse, error)
}

// WebUIEvent represents an event from the WebUI that needs to be processed by an agent
type WebUIEvent struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"session_id"`
	Type        string                 `json:"type"`
	Message     string                 `json:"message"`
	AgentName   string                 `json:"agent_name,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	UserContext map[string]string      `json:"user_context"`
}

// AgentResponse represents a response from an agent back to the WebUI
type AgentResponse struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"session_id"`
	AgentName   string                 `json:"agent_name"`
	Content     string                 `json:"content"`
	Status      string                 `json:"status"` // "processing", "partial", "complete", "error"
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	IsStreaming bool                   `json:"is_streaming"`
	ChunkIndex  int                    `json:"chunk_index,omitempty"`
	TotalChunks int                    `json:"total_chunks,omitempty"`
}

// NewAgentBridge creates a new agent bridge instance
func NewAgentBridge(
	agentManager core.AgentManager,
	sessionManager *EnhancedSessionManager,
	logger core.CoreLogger,
	config *BridgeConfig,
) *AgentBridge {
	if config == nil {
		config = DefaultBridgeConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &AgentBridge{
		agentManager:    agentManager,
		sessionManager:  sessionManager,
		logger:          logger,
		eventHandlers:   make(map[string]AgentEventHandler),
		responseStreams: make(map[string]chan *AgentResponse),
		ctx:             ctx,
		cancel:          cancel,
		config:          config,
	}
}

// Start initializes the agent bridge and starts background routines
func (ab *AgentBridge) Start(ctx context.Context) error {
	ab.logger.Info().Msg("Starting Agent Bridge")

	// Initialize agent manager
	if err := ab.agentManager.InitializeAgents(); err != nil {
		return fmt.Errorf("failed to initialize agents: %w", err)
	}

	// Start background routines
	go ab.cleanupExpiredStreams(ctx)

	ab.logger.Info().Msg("Agent Bridge started successfully")
	return nil
}

// Stop gracefully shuts down the agent bridge
func (ab *AgentBridge) Stop() error {
	ab.logger.Info().Msg("Stopping Agent Bridge")

	ab.cancel()

	// Close all response streams
	ab.streamMutex.Lock()
	for sessionID, stream := range ab.responseStreams {
		close(stream)
		delete(ab.responseStreams, sessionID)
	}
	ab.streamMutex.Unlock()

	ab.logger.Info().Msg("Agent Bridge stopped")
	return nil
}

// ProcessChatMessage processes a chat message from the WebUI and routes it to appropriate agents
func (ab *AgentBridge) ProcessChatMessage(ctx context.Context, sessionID, message string, metadata map[string]interface{}) error {
	// Create WebUI event
	event := &WebUIEvent{
		ID:          generateEventID(),
		SessionID:   sessionID,
		Type:        "chat_message",
		Message:     message,
		Metadata:    metadata,
		Timestamp:   time.Now(),
		UserContext: make(map[string]string),
	}

	// Get session context
	session, err := ab.sessionManager.GetSession(sessionID)
	if err != nil {
		ab.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get session")
		return fmt.Errorf("session not found: %w", err)
	}

	// Add session context to event
	event.UserContext["user_agent"] = session.UserAgent
	event.UserContext["ip_address"] = session.IPAddress

	// Process with agents (async)
	go ab.processEventAsync(ctx, event)

	return nil
}

// processEventAsync processes an event asynchronously with the agent system
func (ab *AgentBridge) processEventAsync(ctx context.Context, event *WebUIEvent) {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, ab.config.AgentTimeout)
	defer cancel()

	// Convert WebUI event to core event
	coreEvent, err := ab.convertWebUIEventToCoreEvent(event)
	if err != nil {
		ab.logger.Error().Err(err).Str("event_id", event.ID).Msg("Failed to convert WebUI event to core event")
		ab.sendErrorResponse(event.SessionID, "EVENT_CONVERSION_ERROR", err.Error())
		return
	}

	// Get available agents
	agents := ab.agentManager.GetActiveAgents()
	if len(agents) == 0 {
		ab.logger.Warn().Str("session_id", event.SessionID).Msg("No active agents available")
		ab.sendErrorResponse(event.SessionID, "NO_AGENTS_AVAILABLE", "No agents are currently available to process your request")
		return
	}

	// For now, use the first available agent
	// TODO: Implement agent selection logic based on capabilities and routing
	agent := agents[0]

	ab.logger.Info().
		Str("session_id", event.SessionID).
		Str("agent_name", agent.Name()).
		Str("message", event.Message).
		Msg("Processing chat message with agent")

	// Send processing status
	ab.sendProcessingResponse(event.SessionID, agent.Name(), "Processing your request...")

	// Create state for agent execution
	stateData := map[string]any{
		"message":      event.Message,
		"session_id":   event.SessionID,
		"user_context": event.UserContext,
	}
	state := core.NewSimpleState(stateData)

	// Execute agent
	result, err := agent.HandleEvent(timeoutCtx, coreEvent, state)
	if err != nil {
		ab.logger.Error().
			Err(err).
			Str("session_id", event.SessionID).
			Str("agent_name", agent.Name()).
			Msg("Agent execution failed")
		ab.sendErrorResponse(event.SessionID, "AGENT_EXECUTION_ERROR", err.Error())
		return
	}

	// Process agent result and send response
	ab.processAgentResult(event.SessionID, agent.Name(), result)
}

// convertWebUIEventToCoreEvent converts a WebUI event to a core event
func (ab *AgentBridge) convertWebUIEventToCoreEvent(event *WebUIEvent) (core.Event, error) {
	// Create event data
	eventData := core.EventData{
		"message":      event.Message,
		"type":         event.Type,
		"session_id":   event.SessionID,
		"timestamp":    event.Timestamp,
		"user_context": event.UserContext,
	}

	// Add metadata
	for key, value := range event.Metadata {
		eventData[key] = value
	}

	// Create metadata map
	metadata := map[string]string{
		"source":     "webui",
		"session_id": event.SessionID,
		"event_id":   event.ID,
	}

	// Create core event
	coreEvent := core.NewEvent("", eventData, metadata)
	coreEvent.SetID(event.ID)

	return coreEvent, nil
}

// processAgentResult processes the result from an agent and sends appropriate responses
func (ab *AgentBridge) processAgentResult(sessionID, agentName string, result core.AgentResult) {
	if result.Error != "" {
		ab.sendErrorResponse(sessionID, "AGENT_RESULT_ERROR", result.Error)
		return
	}

	// Extract response content from state
	content := "I've processed your request."
	if result.OutputState != nil {
		if responseData, ok := result.OutputState.Get("response"); ok {
			if responseStr, ok := responseData.(string); ok {
				content = responseStr
			}
		}
	}

	// Send successful response
	response := &AgentResponse{
		ID:        generateResponseID(),
		SessionID: sessionID,
		AgentName: agentName,
		Content:   content,
		Status:    "complete",
		Metadata: map[string]interface{}{
			"execution_duration_ms": result.Duration.Milliseconds(),
			"start_time":            result.StartTime,
			"end_time":              result.EndTime,
		},
		Timestamp:   time.Now(),
		IsStreaming: false,
	}

	ab.sendResponse(response)
}

// sendProcessingResponse sends a processing status response
func (ab *AgentBridge) sendProcessingResponse(sessionID, agentName, message string) {
	response := &AgentResponse{
		ID:          generateResponseID(),
		SessionID:   sessionID,
		AgentName:   agentName,
		Content:     message,
		Status:      "processing",
		Metadata:    make(map[string]interface{}),
		Timestamp:   time.Now(),
		IsStreaming: false,
	}

	ab.sendResponse(response)
}

// sendErrorResponse sends an error response
func (ab *AgentBridge) sendErrorResponse(sessionID, errorCode, errorMessage string) {
	response := &AgentResponse{
		ID:        generateResponseID(),
		SessionID: sessionID,
		AgentName: "system",
		Content:   "An error occurred while processing your request.",
		Status:    "error",
		Error:     fmt.Sprintf("%s: %s", errorCode, errorMessage),
		Metadata: map[string]interface{}{
			"error_code": errorCode,
		},
		Timestamp:   time.Now(),
		IsStreaming: false,
	}

	ab.sendResponse(response)
}

// sendResponse sends a response through the appropriate channel
func (ab *AgentBridge) sendResponse(response *AgentResponse) {
	ab.streamMutex.RLock()
	stream, exists := ab.responseStreams[response.SessionID]
	ab.streamMutex.RUnlock()

	if !exists {
		ab.logger.Warn().
			Str("session_id", response.SessionID).
			Msg("No response stream found for session")
		return
	}

	select {
	case stream <- response:
		ab.logger.Debug().
			Str("session_id", response.SessionID).
			Str("agent_name", response.AgentName).
			Str("status", response.Status).
			Msg("Response sent to stream")
	case <-time.After(5 * time.Second):
		ab.logger.Warn().
			Str("session_id", response.SessionID).
			Msg("Timeout sending response to stream")
	}
}

// GetResponseStream gets or creates a response stream for a session
func (ab *AgentBridge) GetResponseStream(sessionID string) <-chan *AgentResponse {
	ab.streamMutex.Lock()
	defer ab.streamMutex.Unlock()

	if stream, exists := ab.responseStreams[sessionID]; exists {
		return stream
	}

	stream := make(chan *AgentResponse, ab.config.ResponseBufferSize)
	ab.responseStreams[sessionID] = stream

	ab.logger.Debug().
		Str("session_id", sessionID).
		Msg("Created new response stream")

	return stream
}

// CloseResponseStream closes the response stream for a session
func (ab *AgentBridge) CloseResponseStream(sessionID string) {
	ab.streamMutex.Lock()
	defer ab.streamMutex.Unlock()

	if stream, exists := ab.responseStreams[sessionID]; exists {
		close(stream)
		delete(ab.responseStreams, sessionID)

		ab.logger.Debug().
			Str("session_id", sessionID).
			Msg("Closed response stream")
	}
}

// cleanupExpiredStreams periodically cleans up expired response streams
func (ab *AgentBridge) cleanupExpiredStreams(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ab.performStreamCleanup()
		}
	}
}

// performStreamCleanup removes streams for expired sessions
func (ab *AgentBridge) performStreamCleanup() {
	ab.streamMutex.Lock()
	defer ab.streamMutex.Unlock()

	// Get all active sessions
	sessions, _, err := ab.sessionManager.ListSessions(0, 1000) // Get all sessions
	if err != nil {
		ab.logger.Error().Err(err).Msg("Failed to list sessions for stream cleanup")
		return
	}

	// Create set of active session IDs
	activeSessions := make(map[string]bool)
	for _, session := range sessions {
		activeSessions[session.ID] = true
	}

	// Remove streams for inactive sessions
	var cleanedStreams []string
	for sessionID, stream := range ab.responseStreams {
		if !activeSessions[sessionID] {
			close(stream)
			delete(ab.responseStreams, sessionID)
			cleanedStreams = append(cleanedStreams, sessionID)
		}
	}

	if len(cleanedStreams) > 0 {
		ab.logger.Info().
			Int("count", len(cleanedStreams)).
			Msg("Cleaned up expired response streams")
	}
}

// GetAvailableAgents returns information about available agents
func (ab *AgentBridge) GetAvailableAgents() []AgentInfo {
	agents := ab.agentManager.GetActiveAgents()
	agentInfos := make([]AgentInfo, len(agents))

	for i, agent := range agents {
		agentInfos[i] = AgentInfo{
			Name:         agent.Name(),
			Role:         agent.GetRole(),
			Description:  agent.GetDescription(),
			Capabilities: agent.GetCapabilities(),
			IsEnabled:    agent.IsEnabled(),
			Timeout:      agent.GetTimeout(),
		}
	}

	return agentInfos
}

// AgentInfo provides information about an available agent
type AgentInfo struct {
	Name         string        `json:"name"`
	Role         string        `json:"role"`
	Description  string        `json:"description"`
	Capabilities []string      `json:"capabilities"`
	IsEnabled    bool          `json:"is_enabled"`
	Timeout      time.Duration `json:"timeout"`
}

// Utility functions for ID generation
func generateEventID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("evt_%x", bytes)
}

func generateResponseID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("resp_%x", bytes)
}
