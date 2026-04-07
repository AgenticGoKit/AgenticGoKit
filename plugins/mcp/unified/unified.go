package mcp_unified

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

// unifiedMCPManager supports multiple transport types: TCP, HTTP SSE, HTTP Streaming, WebSocket, STDIO
type unifiedMCPManager struct {
	config           core.MCPConfig
	connectedServers map[string]bool
	tools            []core.MCPToolInfo
	mu               sync.RWMutex
}

// authStreamingHTTPTransport wraps StreamingHTTPTransport with Authorization header support
type authStreamingHTTPTransport struct {
	baseURL      string
	endpoint     string
	client       *http.Client
	sessionID    string
	connected    bool
	mu           sync.RWMutex
	lastResponse *mcp.Message
	authToken    string
}

// authSSETransport implements SSE transport with Authorization header support
type authSSETransport struct {
	baseURL       string
	endpoint      string
	client        *http.Client
	sessionURL    string
	authToken     string
	connected     bool
	mu            sync.RWMutex
	lastResponse  *mcp.Message
	sseConnection *http.Response
}

func newAuthStreamingHTTPTransport(baseURL, endpoint, token string) *authStreamingHTTPTransport {
	return &authStreamingHTTPTransport{
		baseURL:   baseURL,
		endpoint:  endpoint,
		authToken: token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *authStreamingHTTPTransport) Connect(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.connected {
		return nil
	}
	h.connected = true
	return nil
}

func (h *authStreamingHTTPTransport) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connected = false
	h.sessionID = ""
	return nil
}

func (h *authStreamingHTTPTransport) Send(message *mcp.Message) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.connected {
		return fmt.Errorf("transport not connected")
	}

	log.Printf("[Streaming] Sending message: method=%s, id=%v", message.Method, message.ID)

	// Check if this is a notification (no ID field) - notifications don't get responses
	isNotification := message.ID == nil
	if isNotification {
		log.Printf("[Streaming] Message is a notification (no ID) - no response expected")
	}

	url := h.baseURL + h.endpoint
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	log.Printf("[Streaming] POST %s", url)
	log.Printf("[Streaming] Request JSON: %s", string(data))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Content-Type", "application/json")

	// Add authorization header if token is present
	if h.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.authToken)
		log.Printf("[Streaming] Using auth token (length: %d)", len(h.authToken))
	}

	// Add session ID to subsequent requests
	if h.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", h.sessionID)
		log.Printf("[Streaming] Using session ID: %s", h.sessionID)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[Streaming] Response status: %d", resp.StatusCode)

	// Extract session ID from response headers
	sessionID := resp.Header.Get("Mcp-Session-Id")
	if sessionID != "" {
		h.sessionID = sessionID
		log.Printf("[Streaming] Got session ID: %s", sessionID)
	}

	// Read response body and store for Receive()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[Streaming] Response body: %s", string(body))

	// Handle notifications - no response expected
	if isNotification {
		log.Printf("[Streaming] Notification sent, not waiting for response")
		h.lastResponse = nil
		return nil
	}

	// Handle empty responses
	if len(body) == 0 {
		log.Printf("[Streaming] Empty response body")
		h.lastResponse = nil
		return nil
	}

	// Check if response is in SSE format (starts with "event:" or "data:")
	bodyStr := string(body)
	if strings.HasPrefix(bodyStr, "event:") || strings.HasPrefix(bodyStr, "data:") {
		log.Printf("[Streaming] Response is in SSE format, parsing...")

		// Parse SSE format to extract JSON data
		scanner := bufio.NewScanner(bytes.NewReader(body))
		var currentEvent string
		var messageData string

		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event:") {
				currentEvent = strings.TrimSpace(line[6:])
			} else if strings.HasPrefix(line, "data:") {
				data := strings.TrimSpace(line[5:])
				if currentEvent == "message" || currentEvent == "" {
					messageData = data
					break
				}
			}
		}

		if messageData == "" {
			log.Printf("[Streaming] No message data found in SSE response")
			h.lastResponse = nil
			return nil
		}

		var response mcp.Message
		if err := json.Unmarshal([]byte(messageData), &response); err != nil {
			return fmt.Errorf("failed to unmarshal SSE message data: %w", err)
		}

		log.Printf("[Streaming] Parsed SSE response: method=%s, result present=%v", response.Method, response.Result != nil)
		h.lastResponse = &response
		return nil
	}

	// Handle regular JSON response
	var response mcp.Message
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.Printf("[Streaming] Parsed response: method=%s, result present=%v", response.Method, response.Result != nil)

	h.lastResponse = &response
	return nil
}

func (h *authStreamingHTTPTransport) Receive() (*mcp.Message, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	// Return the stored response from the last Send() call
	if h.lastResponse != nil {
		response := h.lastResponse
		h.lastResponse = nil // Clear after returning
		return response, nil
	}

	return nil, fmt.Errorf("no response available")
}

func (h *authStreamingHTTPTransport) GetReader() io.Reader {
	return nil
}

func (h *authStreamingHTTPTransport) GetWriter() io.Writer {
	return nil
}

func (h *authStreamingHTTPTransport) IsConnected() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.connected
}

// SSE Transport implementation
func newAuthSSETransport(baseURL, endpoint, token string) *authSSETransport {
	return &authSSETransport{
		baseURL:   baseURL,
		endpoint:  endpoint,
		authToken: token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *authSSETransport) Connect(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.connected {
		return nil
	}
	h.connected = true
	return nil
}

func (h *authSSETransport) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sseConnection != nil {
		h.sseConnection.Body.Close()
		h.sseConnection = nil
	}
	h.connected = false
	h.sessionURL = ""
	return nil
}

func (h *authSSETransport) Send(message *mcp.Message) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.connected {
		return fmt.Errorf("transport not connected")
	}

	// Special handling for initialize request
	if message.Method == "initialize" && h.sessionURL == "" {
		return h.sendInitializeRequest(message)
	}

	// For other requests, use session-based request
	return h.sendSessionRequest(message)
}

func (h *authSSETransport) sendInitializeRequest(message *mcp.Message) error {
	// First, establish SSE connection to get session endpoint
	sseURL := h.baseURL + h.endpoint

	log.Printf("[SSE] Connecting to %s", sseURL)

	req, err := http.NewRequest("GET", sseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	if h.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.authToken)
		log.Printf("[SSE] Using auth token (length: %d)", len(h.authToken))
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to establish SSE connection: %w", err)
	}

	log.Printf("[SSE] Connection established, status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("SSE connection failed with status %d", resp.StatusCode)
	}

	// Parse SSE stream to get session endpoint
	// SSE format:
	// event: endpoint
	// data: <session-url>
	scanner := bufio.NewScanner(resp.Body)
	var currentEvent string
	var sessionEndpoint string

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[SSE] Received line: %q", line)

		// Parse SSE event format
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimSpace(line[7:]) // Remove "event: " prefix
			log.Printf("[SSE] Event type: %s", currentEvent)
		} else if strings.HasPrefix(line, "data: ") {
			data := strings.TrimSpace(line[6:]) // Remove "data: " prefix
			log.Printf("[SSE] Event data: %s", data)

			// Only process data for "endpoint" event
			if currentEvent == "endpoint" {
				sessionEndpoint = data
				break // Found the endpoint, exit loop
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		resp.Body.Close()
		return fmt.Errorf("error reading SSE stream: %w", err)
	}

	// Validate session endpoint
	if sessionEndpoint == "" {
		resp.Body.Close()
		return fmt.Errorf("failed to get session endpoint from SSE stream")
	}

	log.Printf("[SSE] Extracted session endpoint: %s", sessionEndpoint)

	// Build session URL
	// Check if sessionEndpoint is already a full URL or just a path
	if strings.HasPrefix(sessionEndpoint, "http://") || strings.HasPrefix(sessionEndpoint, "https://") {
		h.sessionURL = sessionEndpoint
	} else {
		// It's a relative path, prepend base URL
		h.sessionURL = h.baseURL + sessionEndpoint
	}

	log.Printf("[SSE] Session URL: %s", h.sessionURL)

	h.sseConnection = resp

	// Now send the initialize request to the session endpoint
	log.Printf("[SSE] Sending initialize request to session")
	return h.sendMessageToSession(message)
}

func (h *authSSETransport) sendMessageToSession(message *mcp.Message) error {
	log.Printf("[SSE] Sending message to session: method=%s, id=%v", message.Method, message.ID)

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	log.Printf("[SSE] Request JSON: %s", string(data))

	req, err := http.NewRequest("POST", h.sessionURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if h.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.authToken)
	}

	log.Printf("[SSE] POST %s", h.sessionURL)

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[SSE] Response status: %d", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[SSE] Response body: %s", string(body))

	// Check if this is a notification (no ID field) - notifications don't get responses
	isNotification := message.ID == nil
	if isNotification {
		log.Printf("[SSE] Message is a notification (no ID) - no response expected")
		h.lastResponse = nil
		return nil
	}

	// Handle SSE protocol: POST returns 202 Accepted, actual response comes via SSE stream
	if resp.StatusCode == http.StatusAccepted || len(body) == 0 {
		log.Printf("[SSE] Status 202/empty body - reading response from SSE stream")

		// Read response from the SSE connection
		if h.sseConnection == nil {
			return fmt.Errorf("SSE connection not established")
		}

		// Read events from SSE stream until we get a message response
		scanner := bufio.NewScanner(h.sseConnection.Body)
		var currentEvent string
		var messageData string

		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("[SSE] Stream line: %q", line)

			if strings.HasPrefix(line, "event: ") {
				currentEvent = strings.TrimSpace(line[7:])
				log.Printf("[SSE] Stream event type: %s", currentEvent)
			} else if strings.HasPrefix(line, "data: ") {
				data := strings.TrimSpace(line[6:])

				// Look for message or response events
				if currentEvent == "message" || currentEvent == "response" {
					messageData = data
					log.Printf("[SSE] Got message data: %s", messageData)
					break
				}
			} else if line == "" {
				// Empty line marks end of event
				if messageData != "" {
					break
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading SSE stream: %w", err)
		}

		if messageData == "" {
			log.Printf("[SSE] No message found in SSE stream")
			h.lastResponse = nil
			return nil
		}

		// Parse the message data
		var response mcp.Message
		if err := json.Unmarshal([]byte(messageData), &response); err != nil {
			return fmt.Errorf("failed to unmarshal SSE message: %w", err)
		}

		log.Printf("[SSE] Parsed SSE response: method=%s, result present=%v", response.Method, response.Result != nil)
		h.lastResponse = &response
		return nil
	}

	// Handle synchronous response (for non-SSE compatible servers)
	var response mcp.Message
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.Printf("[SSE] Parsed response: method=%s, result present=%v", response.Method, response.Result != nil)

	h.lastResponse = &response
	return nil
}

func (h *authSSETransport) sendSessionRequest(message *mcp.Message) error {
	if h.sessionURL == "" {
		return fmt.Errorf("session not established")
	}
	return h.sendMessageToSession(message)
}

func (h *authSSETransport) Receive() (*mcp.Message, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	if h.lastResponse != nil {
		response := h.lastResponse
		h.lastResponse = nil
		return response, nil
	}

	return nil, fmt.Errorf("no response available")
}

func (h *authSSETransport) GetReader() io.Reader {
	return nil
}

func (h *authSSETransport) GetWriter() io.Writer {
	return nil
}

func (h *authSSETransport) IsConnected() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.connected
}

func newUnifiedManager(cfg core.MCPConfig) (core.MCPManager, error) {
	return &unifiedMCPManager{
		config:           cfg,
		connectedServers: make(map[string]bool),
		tools:            []core.MCPToolInfo{},
	}, nil
}

func (m *unifiedMCPManager) Connect(ctx context.Context, serverName string) error {
	// Find server configuration
	var server *core.MCPServerConfig
	for i := range m.config.Servers {
		s := &m.config.Servers[i]
		if s.Name == serverName {
			server = s
			break
		}
	}
	if server == nil {
		return fmt.Errorf("server %s not found in configuration", serverName)
	}
	if !server.Enabled {
		return fmt.Errorf("server %s is disabled", serverName)
	}

	// Mark as connected; actual connectivity is tested during tool operations
	m.mu.Lock()
	m.connectedServers[serverName] = true
	m.mu.Unlock()
	return nil
}

func (m *unifiedMCPManager) Disconnect(serverName string) error {
	m.mu.Lock()
	delete(m.connectedServers, serverName)
	m.mu.Unlock()
	return nil
}

func (m *unifiedMCPManager) DisconnectAll() error {
	m.mu.Lock()
	m.connectedServers = make(map[string]bool)
	m.mu.Unlock()
	return nil
}

func (m *unifiedMCPManager) DiscoverServers(ctx context.Context) ([]core.MCPServerInfo, error) {
	servers := make([]core.MCPServerInfo, 0, len(m.config.Servers))
	for _, s := range m.config.Servers {
		if !s.Enabled {
			continue
		}
		status := "discovered"
		m.mu.RLock()
		if m.connectedServers[s.Name] {
			status = "connected"
		}
		m.mu.RUnlock()

		address := s.Host
		if s.Endpoint != "" {
			address = s.Endpoint
		}

		servers = append(servers, core.MCPServerInfo{
			Name:    s.Name,
			Type:    s.Type,
			Address: address,
			Port:    s.Port,
			Status:  status,
			Version: "",
		})
	}
	return servers, nil
}

func (m *unifiedMCPManager) ListConnectedServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []string
	for name := range m.connectedServers {
		out = append(out, name)
	}
	return out
}

func (m *unifiedMCPManager) GetServerInfo(serverName string) (*core.MCPServerInfo, error) {
	for _, s := range m.config.Servers {
		if s.Name == serverName {
			status := "disconnected"
			m.mu.RLock()
			if m.connectedServers[serverName] {
				status = "connected"
			}
			m.mu.RUnlock()

			address := s.Host
			if s.Endpoint != "" {
				address = s.Endpoint
			}

			info := &core.MCPServerInfo{
				Name:    s.Name,
				Type:    s.Type,
				Address: address,
				Port:    s.Port,
				Status:  status,
				Version: "",
			}
			return info, nil
		}
	}
	return nil, fmt.Errorf("server %s not found", serverName)
}

func (m *unifiedMCPManager) RefreshTools(ctx context.Context) error {
	// For each enabled server, connect and list tools
	var all []core.MCPToolInfo
	for _, s := range m.config.Servers {
		if !s.Enabled {
			continue
		}
		tools, err := m.discoverToolsFromServer(ctx, s.Name)
		if err != nil {
			core.Logger().Warn().
				Str("server_name", s.Name).
				Err(err).
				Msg("Failed to discover tools from server")
			continue
		}
		all = append(all, tools...)
	}
	m.mu.Lock()
	m.tools = all
	m.mu.Unlock()
	return nil
}

func (m *unifiedMCPManager) GetAvailableTools() []core.MCPToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]core.MCPToolInfo(nil), m.tools...)
}

func (m *unifiedMCPManager) GetToolsFromServer(serverName string) []core.MCPToolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []core.MCPToolInfo
	for _, t := range m.tools {
		if t.ServerName == serverName {
			out = append(out, t)
		}
	}
	return out
}

func (m *unifiedMCPManager) HealthCheck(ctx context.Context) map[string]core.MCPHealthStatus {
	health := make(map[string]core.MCPHealthStatus)
	for _, s := range m.config.Servers {
		if !s.Enabled {
			continue
		}
		status := core.MCPHealthStatus{Status: "unknown", LastCheck: time.Now()}

		// Try to create a client and connect briefly for health check
		client, err := m.createClientForServer(&s)
		if err != nil {
			status.Status = "unhealthy"
			status.Error = fmt.Sprintf("Failed to create client: %v", err)
		} else {
			start := time.Now()
			if err := client.Connect(ctx); err != nil {
				status.Status = "unhealthy"
				status.Error = fmt.Sprintf("Connection failed: %v", err)
			} else {
				status.Status = "healthy"
				status.ResponseTime = time.Since(start)
				client.Disconnect()
			}
		}
		health[s.Name] = status
	}
	return health
}

// ExecuteTool implements core.MCPToolExecutor for unified transport support
func (m *unifiedMCPManager) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (core.MCPToolResult, error) {
	// Find server containing this tool
	var target string
	m.mu.RLock()
	for _, t := range m.tools {
		if t.Name == toolName {
			target = t.ServerName
			break
		}
	}
	m.mu.RUnlock()

	// If tool not found in cache, try first enabled server
	if target == "" {
		for _, s := range m.config.Servers {
			if s.Enabled {
				target = s.Name
				break
			}
		}
	}
	if target == "" {
		return core.MCPToolResult{}, fmt.Errorf("no enabled MCP server found for tool %s", toolName)
	}

	// Find server config
	var server *core.MCPServerConfig
	for i := range m.config.Servers {
		if m.config.Servers[i].Name == target {
			server = &m.config.Servers[i]
			break
		}
	}
	if server == nil {
		return core.MCPToolResult{}, fmt.Errorf("server config for %s not found", target)
	}

	// Create client for this server
	client, err := m.createClientForServer(server)
	if err != nil {
		return core.MCPToolResult{}, fmt.Errorf("failed to create client: %w", err)
	}

	start := time.Now()
	if err := client.Connect(ctx); err != nil {
		return core.MCPToolResult{}, fmt.Errorf("failed to connect to MCP server %s: %w", target, err)
	}
	defer client.Disconnect()

	if err := client.Initialize(ctx, mcp.ClientInfo{Name: "agentflow-mcp-client", Version: "1.0.0"}); err != nil {
		return core.MCPToolResult{}, fmt.Errorf("failed to initialize MCP session: %w", err)
	}

	res, err := client.CallTool(ctx, toolName, args)
	if err != nil {
		return core.MCPToolResult{}, fmt.Errorf("tool execution failed: %w", err)
	}

	out := core.MCPToolResult{
		ToolName:   toolName,
		ServerName: target,
		Success:    !res.IsError,
		Duration:   time.Since(start),
	}
	for _, content := range res.Content {
		out.Content = append(out.Content, core.MCPContent{
			Type:     content.Type,
			Text:     content.Text,
			Data:     content.Data,
			MimeType: content.MimeType,
		})
	}
	if res.IsError {
		out.Error = "Tool execution returned error"
		if len(res.Content) > 0 && res.Content[0].Text != "" {
			out.Error = res.Content[0].Text
		}
	}
	return out, nil
}

func (m *unifiedMCPManager) discoverToolsFromServer(ctx context.Context, serverName string) ([]core.MCPToolInfo, error) {
	log.Printf("[MCP] Starting tool discovery for server: %s", serverName)

	// Find server config
	var server *core.MCPServerConfig
	for i := range m.config.Servers {
		if m.config.Servers[i].Name == serverName {
			server = &m.config.Servers[i]
			break
		}
	}
	if server == nil {
		return nil, fmt.Errorf("server %s not found", serverName)
	}

	log.Printf("[MCP] Server config: type=%s, host=%s, port=%d", server.Type, server.Host, server.Port)

	// Create client for this server
	client, err := m.createClientForServer(server)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for %s: %w", serverName, err)
	}

	log.Printf("[MCP] Client created, connecting...")

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", serverName, err)
	}
	defer client.Disconnect()

	log.Printf("[MCP] Connected successfully, initializing MCP protocol...")

	if err := client.Initialize(ctx, mcp.ClientInfo{Name: "agentflow-mcp-client", Version: "1.0.0"}); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP session with %s: %w", serverName, err)
	}

	log.Printf("[MCP] Protocol initialized, listing tools...")

	tools, err := client.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools from %s: %w", serverName, err)
	}

	log.Printf("[MCP] Received %d tools from %s", len(tools), serverName)

	var out []core.MCPToolInfo
	for _, t := range tools {
		log.Printf("[MCP] Tool: %s - %s", t.Name, t.Description)
		out = append(out, core.MCPToolInfo{
			Name:        t.Name,
			Description: t.Description,
			Schema:      t.InputSchema,
			ServerName:  serverName,
		})
	}
	return out, nil
}

// createClientForServer creates appropriate client based on server type
func (m *unifiedMCPManager) createClientForServer(server *core.MCPServerConfig) (*client.Client, error) {
	newClient := func(tr transport.Transport) *client.Client {
		logger := log.New(io.Discard, "", 0)
		if v := os.Getenv("MCP_NAVIGATOR_DEBUG"); v != "" && v != "0" {
			logger = log.Default()
		}

		cfg := client.ClientConfig{
			Name:    "agentflow-mcp-client",
			Version: "1.0.0",
			Timeout: 30 * time.Second,
			Logger:  logger,
		}

		return client.NewClient(tr, cfg)
	}

	switch server.Type {
	case "tcp":
		tr := transport.NewTCPTransport(server.Host, server.Port)
		return newClient(tr), nil

	case "http_sse":
		endpoint := server.Endpoint
		var baseURL string
		var ssePath string

		if endpoint == "" {
			// No endpoint provided, build from host:port
			if server.Host != "" && server.Port > 0 {
				baseURL = fmt.Sprintf("http://%s:%d", server.Host, server.Port)
				ssePath = "/sse"
			} else {
				return nil, fmt.Errorf("http_sse server %s requires either endpoint or host:port configuration", server.Name)
			}
		} else {
			// Endpoint provided - use it as-is (no additional path)
			baseURL = endpoint
			ssePath = "" // Don't append /sse, endpoint already contains the full path
		}

		// Check for auth token in environment
		authToken := os.Getenv("MCP_GATEWAY_AUTH_TOKEN")
		if authToken == "" {
			authToken = os.Getenv("MCP_AUTH_TOKEN")
		}

		log.Printf("[Transport] Creating http_sse client: baseURL=%s, path=%s, hasAuth=%v", baseURL, ssePath, authToken != "")

		if authToken != "" {
			// Use custom SSE transport with auth support
			sseTransport := newAuthSSETransport(baseURL, ssePath, authToken)
			return newClient(sseTransport), nil
		}
		// Fallback to standard transport if no auth token
		sseTransport := transport.NewSSETransport(baseURL, ssePath)
		return newClient(sseTransport), nil

	case "http_streaming":
		endpoint := server.Endpoint
		var baseURL string
		var streamPath string

		if endpoint == "" {
			// No endpoint provided, build from host:port
			if server.Host != "" && server.Port > 0 {
				baseURL = fmt.Sprintf("http://%s:%d", server.Host, server.Port)
				streamPath = "/stream"
			} else {
				return nil, fmt.Errorf("http_streaming server %s requires either endpoint or host:port configuration", server.Name)
			}
		} else {
			// Endpoint provided - use it as-is (no additional path)
			baseURL = endpoint
			streamPath = "" // Don't append /stream, endpoint already contains the full path
		}

		// Check for auth token in environment
		authToken := os.Getenv("MCP_GATEWAY_AUTH_TOKEN")
		if authToken == "" {
			authToken = os.Getenv("MCP_AUTH_TOKEN")
		}

		log.Printf("[Transport] Creating http_streaming client: baseURL=%s, path=%s, hasAuth=%v", baseURL, streamPath, authToken != "")

		if authToken != "" {
			// Use custom transport with auth support
			streamingTransport := newAuthStreamingHTTPTransport(baseURL, streamPath, authToken)
			return newClient(streamingTransport), nil
		}
		// Fallback to standard transport if no auth token
		streamingTransport := transport.NewStreamingHTTPTransport(baseURL, streamPath)
		return newClient(streamingTransport), nil

	case "websocket":
		url := fmt.Sprintf("ws://%s:%d", server.Host, server.Port)
		tr := transport.NewWebSocketTransport(url)
		return newClient(tr), nil

	case "stdio":
		tr := transport.NewStdioTransport(server.Command, []string{})
		return newClient(tr), nil

	default:
		return nil, fmt.Errorf("unsupported transport type: %s", server.Type)
	}
}

func (m *unifiedMCPManager) GetMetrics() core.MCPMetrics {
	m.mu.RLock()
	connected := len(m.connectedServers)
	tools := len(m.tools)
	m.mu.RUnlock()
	return core.MCPMetrics{
		ConnectedServers: connected,
		TotalTools:       tools,
		ServerMetrics:    map[string]core.MCPServerMetrics{},
	}
}

// Register the unified manager factory - this replaces other transport plugins
func init() {
	core.SetMCPManagerFactory(func(cfg core.MCPConfig) (core.MCPManager, error) {
		return newUnifiedManager(cfg)
	})
}
