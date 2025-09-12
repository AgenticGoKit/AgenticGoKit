package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kunalkushwaha/agenticgokit/core"
)

// ConnectionManager manages all WebSocket connections and their associated sessions
type ConnectionManager struct {
	// WebSocket upgrader
	upgrader websocket.Upgrader

	// Connected clients mapping sessionID -> connection
	connections map[string]*ClientConnection

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Message channels
	register   chan *ClientConnection
	unregister chan *ClientConnection
	broadcast  chan []byte

	// Session manager reference - using enhanced session manager
	sessionManager *EnhancedSessionManager

	// Logger
	logger core.CoreLogger

	// Agent system integration (optional)
	bridge *AgentBridge

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Metrics
	totalConnections  int64
	activeConnections int64
	messagesSent      int64
	messagesReceived  int64
}

// ClientConnection represents a WebSocket connection with metadata
type ClientConnection struct {
	// WebSocket connection
	conn *websocket.Conn

	// Session information
	sessionID string
	userAgent string

	// Connection metadata
	connectedAt time.Time
	lastPong    time.Time

	// Message channels
	send chan []byte

	// Connection manager reference
	manager *ConnectionManager

	// Context for connection lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

// NewConnectionManager creates a new WebSocket connection manager
func NewConnectionManager(sessionManager *EnhancedSessionManager, logger core.CoreLogger) *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for now - should be configurable in production
			return true
		},
	}

	return &ConnectionManager{
		upgrader:       upgrader,
		connections:    make(map[string]*ClientConnection),
		register:       make(chan *ClientConnection),
		unregister:     make(chan *ClientConnection),
		broadcast:      make(chan []byte),
		sessionManager: sessionManager,
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// AttachBridge connects the agent bridge to the connection manager (optional)
func (cm *ConnectionManager) AttachBridge(bridge *AgentBridge) {
	cm.bridge = bridge
}

// Start begins the connection manager's main loop
func (cm *ConnectionManager) Start() {
	go cm.run()
}

// Stop gracefully shuts down the connection manager
func (cm *ConnectionManager) Stop() {
	cm.cancel()
}

// run is the main event loop for the connection manager
func (cm *ConnectionManager) run() {
	ticker := time.NewTicker(30 * time.Second) // Ping interval
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return

		case conn := <-cm.register:
			cm.handleRegister(conn)

		case conn := <-cm.unregister:
			cm.handleUnregister(conn)

		case message := <-cm.broadcast:
			cm.handleBroadcast(message)

		case <-ticker.C:
			cm.handlePing()
		}
	}
}

// HandleWebSocket handles WebSocket upgrade and connection setup
func (cm *ConnectionManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := cm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client connection
	ctx, cancel := context.WithCancel(cm.ctx)
	client := &ClientConnection{
		conn:        conn,
		userAgent:   r.Header.Get("User-Agent"),
		connectedAt: time.Now(),
		lastPong:    time.Now(),
		send:        make(chan []byte, 256),
		manager:     cm,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Register the connection
	cm.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// SendToSession sends a message to a specific session
func (cm *ConnectionManager) SendToSession(sessionID string, message *WebSocketMessage) error {
	cm.mu.RLock()
	conn, exists := cm.connections[sessionID]
	cm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not connected", sessionID)
	}

	// Marshal message to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Send to client
	select {
	case conn.send <- data:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send timeout for session %s", sessionID)
	}
}

// BroadcastToAll sends a message to all connected clients
func (cm *ConnectionManager) BroadcastToAll(message *WebSocketMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	select {
	case cm.broadcast <- data:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("broadcast timeout")
	}
}

// GetConnectionCount returns the number of active connections
func (cm *ConnectionManager) GetConnectionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.connections)
}

// GetConnectedSessions returns a list of connected session IDs
func (cm *ConnectionManager) GetConnectedSessions() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	sessions := make([]string, 0, len(cm.connections))
	for sessionID := range cm.connections {
		if sessionID != "" {
			sessions = append(sessions, sessionID)
		}
	}
	return sessions
}

// Internal handler methods

func (cm *ConnectionManager) handleRegister(conn *ClientConnection) {
	log.Printf("[WebUI] transport=websocket event=connected user_agent=%s", conn.userAgent)

	// Send welcome message
	welcome := NewSystemMessage("", "info", "Connected to AgenticGoKit WebUI")
	cm.sendToConnection(conn, welcome)
}

func (cm *ConnectionManager) handleUnregister(conn *ClientConnection) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn.sessionID != "" {
		delete(cm.connections, conn.sessionID)
		log.Printf("[WebUI] transport=websocket event=disconnected session_id=%s", conn.sessionID)
	}

	close(conn.send)
	conn.cancel()
}

func (cm *ConnectionManager) handleBroadcast(message []byte) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, conn := range cm.connections {
		select {
		case conn.send <- message:
		default:
			// Connection is blocked, remove it
			go func(c *ClientConnection) {
				cm.unregister <- c
			}(conn)
		}
	}
}

func (cm *ConnectionManager) handlePing() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, conn := range cm.connections {
		// Check if connection is still alive
		if time.Since(conn.lastPong) > 60*time.Second {
			log.Printf("Connection timeout for session %s", conn.sessionID)
			go func(c *ClientConnection) {
				cm.unregister <- c
			}(conn)
			continue
		}

		// Send ping
		ping := NewPong(conn.sessionID)
		cm.sendToConnection(conn, ping)
	}
}

func (cm *ConnectionManager) sendToConnection(conn *ClientConnection, message *WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case conn.send <- data:
	default:
		// Connection is blocked
		go func() {
			cm.unregister <- conn
		}()
	}
}

// ClientConnection methods

// readPump handles incoming messages from the WebSocket connection
func (c *ClientConnection) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	// Set read limits and timeouts
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.lastPong = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// Read message
		_, messageData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse and handle message
		c.handleMessage(messageData)
	}
}

// writePump handles outgoing messages to the WebSocket connection
func (c *ClientConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return

		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *ClientConnection) handleMessage(data []byte) {
	var msg WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to parse WebSocket message: %v", err)
		c.sendError("INVALID_JSON", "Failed to parse message", err.Error())
		return
	}

	// Validate message
	if err := msg.Validate(); err != nil {
		log.Printf("Invalid WebSocket message: %v", err)
		c.sendError("INVALID_MESSAGE", "Message validation failed", err.Error())
		return
	}

	// Handle different message types
	switch msg.Type {
	case MsgTypeSessionCreate:
		c.handleSessionCreate(&msg)
	case MsgTypeSessionJoin:
		c.handleSessionJoin(&msg)
	case MsgTypeChatMessage:
		c.handleChatMessage(&msg)
	case MsgTypePing:
		c.handlePing(&msg)
	case MsgTypeTyping:
		c.handleTyping(&msg)
	default:
		c.sendError("UNKNOWN_MESSAGE_TYPE", "Unknown message type", string(msg.Type))
	}
}

// Message handlers

func (c *ClientConnection) handleSessionCreate(msg *WebSocketMessage) {
	_, err := msg.ParseSessionCreate()
	if err != nil {
		c.sendError("INVALID_SESSION_CREATE", "Failed to parse session create", err.Error())
		return
	}

	// Create new session with enhanced session manager
	session, err := c.manager.sessionManager.CreateSession(c.userAgent, "127.0.0.1") // TODO: get real IP
	if err != nil {
		c.sendError("SESSION_CREATE_FAILED", "Failed to create session", err.Error())
		return
	}

	c.sessionID = session.ID

	// Store connection in manager
	c.manager.mu.Lock()
	c.manager.connections[c.sessionID] = c
	c.manager.mu.Unlock()

	// Send session status
	statusMsg := NewSessionStatus(
		session.ID,
		"active",
		len(session.Messages),
		session.CreatedAt,
		session.LastUsed,
	)
	c.manager.sendToConnection(c, statusMsg)

	log.Printf("Created new session %s", session.ID)
}

func (c *ClientConnection) handleSessionJoin(msg *WebSocketMessage) {
	data, err := msg.ParseSessionJoin()
	if err != nil {
		c.sendError("INVALID_SESSION_JOIN", "Failed to parse session join", err.Error())
		return
	}

	// Validate session exists
	session, err := c.manager.sessionManager.GetSession(data.SessionID)
	if err != nil {
		c.sendError("SESSION_NOT_FOUND", "Session not found", data.SessionID)
		return
	}

	// Join session
	c.sessionID = data.SessionID

	// Store connection in manager
	c.manager.mu.Lock()
	c.manager.connections[c.sessionID] = c
	c.manager.mu.Unlock()

	// Send session status and history
	statusMsg := NewSessionStatus(
		session.ID,
		"active",
		len(session.Messages),
		session.CreatedAt,
		session.LastUsed,
	)
	c.manager.sendToConnection(c, statusMsg)

	// Send message history
	for _, chatMsg := range session.Messages {
		wsMsg := &WebSocketMessage{
			Type:      MsgTypeAgentResponse,
			SessionID: session.ID,
			MessageID: generateMessageID(),
			Timestamp: chatMsg.Timestamp,
			Data: map[string]interface{}{
				"agent_name": chatMsg.Role, // Use Role field
				"content":    chatMsg.Content,
				"status":     "complete",
			},
		}
		c.manager.sendToConnection(c, wsMsg)
	}

	log.Printf("Client joined session %s", session.ID)
}

func (c *ClientConnection) handleChatMessage(msg *WebSocketMessage) {
	if c.sessionID == "" {
		c.sendError("NO_SESSION", "No active session", "Create or join a session first")
		return
	}

	data, err := msg.ParseChatMessage()
	if err != nil {
		c.sendError("INVALID_CHAT_MESSAGE", "Failed to parse chat message", err.Error())
		return
	}

	// Create chat message with enhanced structure
	chatMsg := ChatMessage{
		ID:        generateMessageID(),
		Role:      "user",
		Content:   data.Content,
		Timestamp: time.Now(),
	}

	// Add message to session using enhanced session manager
	err = c.manager.sessionManager.AddMessage(c.sessionID, chatMsg)
	if err != nil {
		c.sendError("MESSAGE_ADD_FAILED", "Failed to add message to session", err.Error())
		return
	}

	if c.manager.bridge == nil {
		// Fallback mock response
		response := NewAgentResponse(
			c.sessionID,
			"MockAgent",
			fmt.Sprintf("Received: %s", data.Content),
			"complete",
		)
		c.manager.sendToConnection(c, response)
		log.Printf("[WebUI] transport=websocket event=chat session_id=%s mode=mock", c.sessionID)
		return
	}

	// With bridge: trigger processing and forward stream to WS
	// Start processing via bridge (HTTP handlers already use bridge for sessions)
	go func(sessionID, content string) {
		// Ensure a session response channel exists
		stream := c.manager.bridge.GetResponseStream(sessionID)

		// Kick off processing
		_ = c.manager.bridge.ProcessChatMessage(c.ctx, sessionID, content, map[string]interface{}{})

		// Forward responses as WS messages
		for resp := range stream {
			switch resp.Status {
			case "processing":
				// Map to progress message (simple form)
				progress := NewAgentProgress(sessionID, []AgentStatus{{
					Name:     resp.AgentName,
					Status:   "processing",
					Progress: 0.0,
					Message:  resp.Content,
				}}, 0.0)
				c.manager.sendToConnection(c, progress)
			case "partial":
				chunk := NewAgentChunk(sessionID, resp.AgentName, resp.Content, resp.ChunkIndex, resp.TotalChunks, resp.Metadata)
				c.manager.sendToConnection(c, chunk)
			case "complete":
				complete := NewAgentComplete(sessionID, resp.AgentName, resp.Content, map[string]interface{}{}, resp.Metadata)
				c.manager.sendToConnection(c, complete)
			case "error":
				errMsg := NewAgentError(sessionID, resp.AgentName, "AGENT_ERROR", resp.Error, resp.Metadata)
				c.manager.sendToConnection(c, errMsg)
			default:
				// Fallback to generic agent response
				generic := NewAgentResponse(sessionID, resp.AgentName, resp.Content, resp.Status)
				c.manager.sendToConnection(c, generic)
			}
		}
	}(c.sessionID, data.Content)
}

func (c *ClientConnection) handlePing(msg *WebSocketMessage) {
	pong := NewPong(c.sessionID)
	c.manager.sendToConnection(c, pong)
}

func (c *ClientConnection) handleTyping(msg *WebSocketMessage) {
	// Typing indicators could be broadcast to other clients in the future
	log.Printf("Typing indicator received from session %s", c.sessionID)
}

func (c *ClientConnection) sendError(code, message, details string) {
	errorMsg := NewWSErrorMessage(c.sessionID, code, message, details)
	c.manager.sendToConnection(c, errorMsg)
}
