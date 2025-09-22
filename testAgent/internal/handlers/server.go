package handlers

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"
    "time"

    "context"
    tracingh "testAgent/internal/tracing"

    "github.com/gorilla/websocket"
    "github.com/kunalkushwaha/agenticgokit/core"
)

// ResultStore abstracts access to results collected by the app
type ResultStore interface {
    Reset()
    Latest() (string, bool)
}

// Server bundles dependencies for HTTP and WebSocket handlers
type Server struct {
    Ctx           context.Context
    Runner        core.Runner
    Config        *core.Config
    Orchestrator  core.Orchestrator
    AgentHandlers map[string]core.AgentHandler
    Results       ResultStore
}

func NewServer(ctx context.Context, runner core.Runner, cfg *core.Config, orch core.Orchestrator, agents map[string]core.AgentHandler, results ResultStore) *Server {
    return &Server{Ctx: ctx, Runner: runner, Config: cfg, Orchestrator: orch, AgentHandlers: agents, Results: results}
}

// --- HTTP Handlers ---

func (s *Server) HandleGetAgents(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    agents := []map[string]any{}
    for name := range s.AgentHandlers {
        if strings.Contains(name, "error-handler") {
            continue
        }
        agents = append(agents, map[string]any{
            "id": name, "name": name, "description": fmt.Sprintf("Agent %s from configuration", name),
        })
    }
    if len(agents) == 0 {
        agents = []map[string]any{map[string]any{"id": "agent1", "name": "Agent1", "description": "Default agent 1"}, map[string]any{"id": "agent2", "name": "Agent2", "description": "Default agent 2"}}
    }
    _ = json.NewEncoder(w).Encode(agents)
}

func (s *Server) HandleChat(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req struct {
        Message          string `json:"message"`
        Agent            string `json:"agent"`
        UseOrchestration bool   `json:"useOrchestration,omitempty"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("ERROR: Failed to decode JSON request: %v", err)
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    log.Printf("INFO: Received chat request - Agent: %s, UseOrchestration: %v", req.Agent, req.UseOrchestration)
    agentHandler, exists := s.AgentHandlers[req.Agent]
    if !exists {
        http.Error(w, fmt.Sprintf("Agent '%s' not found", req.Agent), http.StatusNotFound)
        return
    }

    var agentResponse, status string
    if req.UseOrchestration {
        if s.Orchestrator == nil {
            http.Error(w, "Orchestrator not available", http.StatusInternalServerError)
            return
        }
        s.Results.Reset()
        tracingh.RecordEdge("webui-session", "User", req.Agent, req.Message)
        event := core.NewEvent(req.Agent, core.EventData{"message": req.Message, "user_input": req.Message}, map[string]string{"route": req.Agent, "session_id": "webui-session"})
        if _, err := s.Orchestrator.Dispatch(s.Ctx, event); err != nil {
            http.Error(w, fmt.Sprintf("Orchestration error: %v", err), http.StatusInternalServerError)
            return
        }
        if content, ok := s.Results.Latest(); ok {
            agentResponse, status = content, "completed"
        } else {
            agentResponse, status = "Orchestration completed, but no agent responses captured.", "no_output"
        }
    } else {
        // direct
        state := core.NewState()
        state.Set("message", req.Message)
        state.Set("user_input", req.Message)
        tracingh.RecordEdge("webui-session", "User", req.Agent, req.Message)
        event := core.NewEvent(req.Agent, core.EventData{"message": req.Message}, map[string]string{"route": req.Agent, "session_id": "webui-session"})
        result, err := agentHandler.Run(s.Ctx, event, state)
        if err != nil {
            http.Error(w, fmt.Sprintf("Agent processing error: %v", err), http.StatusInternalServerError)
            return
        }
        if result.OutputState != nil {
            if v, ok := result.OutputState.Get("result"); ok {
                agentResponse = fmt.Sprintf("%v", v)
            } else if v, ok := result.OutputState.Get("response"); ok {
                agentResponse = fmt.Sprintf("%v", v)
            } else if v, ok := result.OutputState.Get("output"); ok {
                agentResponse = fmt.Sprintf("%v", v)
            } else {
                agentResponse = "Agent processed your request successfully"
            }
        } else {
            agentResponse = "Agent processed your request"
        }
        status = "completed"
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{"response": agentResponse, "agent": req.Agent, "status": status})
}

func (s *Server) HandleCompositionDiagram(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    name := "agentflow"
    mode := "composition"
    if s.Config != nil {
        if s.Config.AgentFlow.Name != "" {
            name = s.Config.AgentFlow.Name
        }
        if s.Config.Orchestration.Mode != "" {
            mode = s.Config.Orchestration.Mode
        }
    }
    var agents []core.Agent
    gen := core.NewMermaidGenerator()
    mcfg := core.DefaultMermaidConfig()
    diagram := gen.GenerateCompositionDiagram(mode, name, agents, mcfg)
    _ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": map[string]any{"title": name + " (" + mode + ")", "diagram": diagram}})
}

func (s *Server) HandleTraceDiagram(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    session := r.URL.Query().Get("session")
    if session == "" {
        session = "webui-session"
    }
    linear := true
    if v := r.URL.Query().Get("linear"); v == "false" || v == "0" || strings.EqualFold(v, "no") {
        linear = false
    }
    theme := r.URL.Query().Get("theme")
    var diagram string
    var labels []tracingh.LabelInfo
    if theme != "" {
        diagram, labels = tracingh.BuildTraceDataThemed(s.Runner, session, linear, theme)
    } else {
        diagram, labels = tracingh.BuildTraceData(s.Runner, session, linear)
    }
    _ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": map[string]any{"session": session, "diagram": diagram, "labels": labels, "edges": len(labels)}})
}

// --- WebSocket ---

var wsUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin:     func(r *http.Request) bool { return true },
}

type wsInbound struct {
    Type             string
    Agent            string
    Message          string
    UseOrchestration bool
}

type wsOutbound struct {
    Type      string                 `json:"type"`
    Agent     string                 `json:"agent,omitempty"`
    Content   string                 `json:"content,omitempty"`
    Status    string                 `json:"status,omitempty"`
    Chunk     int                    `json:"chunk_index,omitempty"`
    Total     int                    `json:"total_chunks,omitempty"`
    Timestamp int64                  `json:"timestamp"`
    Data      map[string]interface{} `json:"data,omitempty"`
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := wsUpgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WS upgrade failed: %v", err)
        return
    }
    defer conn.Close()
    _ = conn.WriteJSON(wsOutbound{Type: "welcome", Timestamp: time.Now().Unix()})
    for {
        var msg wsInbound
        if err := conn.ReadJSON(&msg); err != nil {
            log.Printf("WS read error: %v", err)
            return
        }
        switch msg.Type {
        case "chat":
            if msg.Agent == "" || msg.Message == "" {
                _ = conn.WriteJSON(wsOutbound{Type: "error", Status: "bad_request", Content: "agent and message required", Timestamp: time.Now().Unix()})
                continue
            }
            _ = conn.WriteJSON(wsOutbound{Type: "agent_progress", Agent: msg.Agent, Status: "processing", Content: "Processing...", Timestamp: time.Now().Unix()})
            go func(m wsInbound) {
                tracingh.RecordEdge("webui-session", "User", m.Agent, m.Message)
                content, status := s.processRequest(m.Agent, m.Message, m.UseOrchestration)
                chunks := chunkString(content, 180)
                total := len(chunks)
                for i, c := range chunks {
                    _ = conn.WriteJSON(wsOutbound{Type: "agent_chunk", Agent: m.Agent, Content: c, Chunk: i, Total: total, Timestamp: time.Now().Unix()})
                    time.Sleep(50 * time.Millisecond)
                }
                _ = conn.WriteJSON(wsOutbound{Type: "agent_complete", Agent: m.Agent, Status: status, Content: content, Timestamp: time.Now().Unix()})
            }(msg)
        default:
            _ = conn.WriteJSON(wsOutbound{Type: "error", Status: "unknown_type", Content: "Unsupported message type", Timestamp: time.Now().Unix()})
        }
    }
}

func (s *Server) processRequest(agent string, message string, useOrch bool) (string, string) {
    if useOrch {
        if s.Orchestrator == nil {
            return "Orchestrator not available", "error"
        }
        s.Results.Reset()
        event := core.NewEvent(agent, core.EventData{"message": message, "user_input": message}, map[string]string{"route": agent, "session_id": "webui-session"})
        if _, err := s.Orchestrator.Dispatch(s.Ctx, event); err != nil {
            return fmt.Sprintf("Orchestration error: %v", err), "error"
        }
        if content, ok := s.Results.Latest(); ok {
            return content, "completed"
        }
        return "Orchestration completed, but no agent responses captured.", "no_output"
    }
    // direct
    agentHandler, exists := s.AgentHandlers[agent]
    if !exists {
        return fmt.Sprintf("Agent '%s' not found", agent), "not_found"
    }
    state := core.NewState()
    state.Set("message", message)
    state.Set("user_input", message)
    event := core.NewEvent(agent, core.EventData{"message": message}, map[string]string{"route": agent, "session_id": "webui-session"})
    result, err := agentHandler.Run(s.Ctx, event, state)
    if err != nil {
        return fmt.Sprintf("Agent processing error: %v", err), "error"
    }
    if result.OutputState != nil {
        if v, ok := result.OutputState.Get("result"); ok {
            return fmt.Sprintf("%v", v), "completed"
        } else if v, ok := result.OutputState.Get("response"); ok {
            return fmt.Sprintf("%v", v), "completed"
        } else if v, ok := result.OutputState.Get("output"); ok {
            return fmt.Sprintf("%v", v), "completed"
        }
    }
    return "Agent processed your request", "completed"
}

func chunkString(s string, size int) []string {
    if size <= 0 || len(s) == 0 {
        return []string{s}
    }
    var chunks []string
    for start := 0; start < len(s); start += size {
        end := start + size
        if end > len(s) {
            end = len(s)
        }
        chunks = append(chunks, s[start:end])
    }
    return chunks
}
