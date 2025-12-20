package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Message types for WebSocket communication
const (
	MsgTypeUserMessage    = "user_message"
	MsgTypeWorkflowStart  = "workflow_start"
	MsgTypeAgentStart     = "agent_start"
	MsgTypeAgentProgress  = "agent_progress"
	MsgTypeAgentComplete  = "agent_complete"
	MsgTypeWorkflowDone   = "workflow_done"
	MsgTypeError          = "error"
	MsgTypeChatHistory    = "chat_history"
	MsgTypeSessionCreated = "session_created"
	MsgTypeAgentConfig    = "agent_config" // New: sends agent configuration to frontend
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string                 `json:"type"`
	Content   string                 `json:"content,omitempty"`
	Agent     string                 `json:"agent,omitempty"`
	Step      string                 `json:"step,omitempty"`
	Progress  int                    `json:"progress,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Timestamp float64                `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessageSender is a function type for sending WebSocket messages
type MessageSender func(WSMessage)

// WebSocketServer provides reusable WebSocket server infrastructure
type WebSocketServer struct {
	workflow       WorkflowExecutor
	sessionManager *SessionManager
	upgrader       websocket.Upgrader
	port           string
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(port string, workflow WorkflowExecutor) *WebSocketServer {
	return &WebSocketServer{
		workflow:       workflow,
		sessionManager: NewSessionManager(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		port: port,
	}
}

// HandleWebSocket handles WebSocket connections (reusable)
func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Create new session
	session := s.sessionManager.CreateSession()

	// Send session created message
	s.sendWSMessage(conn, WSMessage{
		Type:      MsgTypeSessionCreated,
		SessionID: session.ID,
		Content:   s.workflow.WelcomeMessage(),
		Timestamp: float64(time.Now().Unix()),
	})

	// Send agent configuration
	s.sendWSMessage(conn, WSMessage{
		Type:      MsgTypeAgentConfig,
		SessionID: session.ID,
		Timestamp: float64(time.Now().Unix()),
		Metadata: map[string]interface{}{
			"workflow_name": s.workflow.Name(),
			"agents":        s.workflow.GetAgents(),
		},
	})

	// Send chat history if available
	if len(session.Messages) > 0 {
		historyMsg := WSMessage{
			Type:      MsgTypeChatHistory,
			SessionID: session.ID,
			Timestamp: float64(time.Now().Unix()),
			Metadata: map[string]interface{}{
				"messages": session.Messages,
			},
		}
		s.sendWSMessage(conn, historyMsg)
	}

	// Handle incoming messages
	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected WebSocket close: %v", err)
			}
			break
		}

		if msg.Type == MsgTypeUserMessage {
			// Store user message
			s.sessionManager.AddMessage(session.ID, ChatMessage{
				Role:      "user",
				Content:   msg.Content,
				Timestamp: time.Now(),
			})

			// Create message sender for this session
			messageSender := func(wsMsg WSMessage) {
				wsMsg.SessionID = session.ID
				s.sendWSMessage(conn, wsMsg)

				// Store agent messages
				if wsMsg.Type == MsgTypeAgentComplete && wsMsg.Agent != "" {
					s.sessionManager.AddMessage(session.ID, ChatMessage{
						Role:      wsMsg.Agent,
						Content:   wsMsg.Content,
						Agent:     wsMsg.Agent,
						Timestamp: time.Now(),
					})
				}
			}

			// Execute workflow
			ctx := context.Background()
			if err := s.workflow.Execute(ctx, msg.Content, messageSender); err != nil {
				log.Printf("Workflow error: %v", err)
				messageSender(WSMessage{
					Type:      MsgTypeError,
					Content:   err.Error(),
					Timestamp: float64(time.Now().Unix()),
				})
			}
		}
	}
}

func (s *WebSocketServer) sendWSMessage(conn *websocket.Conn, msg WSMessage) {
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send WebSocket message: %v", err)
	}
}

// HandleHealth handles health check endpoint (reusable)
func (s *WebSocketServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   s.workflow.Name(),
	})
}

// HandleSessions lists all sessions (reusable)
func (s *WebSocketServer) HandleSessions(w http.ResponseWriter, r *http.Request) {
	sessions := s.sessionManager.ListSessions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// Start starts the server (reusable)
func (s *WebSocketServer) Start() error {
	log.Println("🚀 Starting WebSocket Server...")
	log.Printf("📡 Service: %s", s.workflow.Name())
	log.Println("📡 WebSocket API: ws://localhost:" + s.port + "/ws")
	log.Println("💡 React Frontend:")
	log.Println("   1. cd frontend")
	log.Println("   2. npm install (first time only)")
	log.Println("   3. npm run dev")
	log.Println("   4. Open http://localhost:5173")

	// Setup routes
	http.HandleFunc("/ws", s.HandleWebSocket)
	http.HandleFunc("/api/health", s.HandleHealth)
	http.HandleFunc("/api/sessions", s.HandleSessions)

	// Serve info page at root
	http.HandleFunc("/", s.serveInfoPage)

	addr := ":" + s.port
	log.Printf("✅ Server running on http://localhost%s", addr)

	return http.ListenAndServe(addr, nil)
}

func (s *WebSocketServer) serveInfoPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html>
<head>
    <title>` + s.workflow.Name() + ` - API Server</title>
</head>
<body style="font-family: system-ui; max-width: 800px; margin: 50px auto; padding: 20px;">
    <h1>📝 ` + s.workflow.Name() + ` - API Server</h1>
    <p>This is the backend API server. The frontend is separate.</p>
    
    <h2>🚀 Getting Started</h2>
    <ol>
        <li>Navigate to the frontend directory: <code>cd frontend</code></li>
        <li>Install dependencies (first time): <code>npm install</code></li>
        <li>Start the development server: <code>npm run dev</code></li>
        <li>Open <a href="http://localhost:5173">http://localhost:5173</a> in your browser</li>
    </ol>
    
    <h2>📡 API Endpoints</h2>
    <ul>
        <li><strong>WebSocket:</strong> ws://localhost:` + s.port + `/ws</li>
        <li><strong>Health Check:</strong> <a href="/api/health">/api/health</a></li>
        <li><strong>Sessions:</strong> <a href="/api/sessions">/api/sessions</a></li>
    </ul>
</body>
</html>`
	w.Write([]byte(html))
}
