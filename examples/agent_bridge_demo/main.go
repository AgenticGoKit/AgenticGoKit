package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
	"github.com/kunalkushwaha/agenticgokit/internal/webui"
)

func main() {
	// Use the core logger
	logger := core.Logger()
	logger.Info().Msg("🚀 Starting AgenticGoKit WebUI Demo")

	// Create core config
	config := &core.Config{}

	// Create mock agent manager
	agentManager := NewExampleAgentManager()

	// Debug: Check if agents were created
	activeAgents := agentManager.GetActiveAgents()
	logger.Info().Int("created_agents_count", len(activeAgents)).Msg("Agent manager created with agents")
	for _, agent := range activeAgents {
		logger.Info().Str("agent_name", agent.Name()).Str("agent_role", agent.GetRole()).Msg("Agent available")
	}

	// Create server config
	serverConfig := webui.ServerConfig{
		Port:         "8080",
		Config:       config,
		AgentManager: agentManager,
	}

	// Create WebUI server
	server := webui.NewServer(serverConfig)

	// Create session manager for agent handlers
	sessionConfig := webui.DefaultSessionConfig()
	sessionManager, err := webui.NewEnhancedSessionManager(config, sessionConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create session manager")
		os.Exit(1)
	}

	// Create agent bridge
	bridgeConfig := webui.DefaultBridgeConfig()
	bridge := webui.NewAgentBridge(agentManager, sessionManager, logger, bridgeConfig)

	// Start the bridge
	ctx := context.Background()
	if err := bridge.Start(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to start agent bridge")
		os.Exit(1)
	}

	// Start server in a goroutine
	go func() {
		logger.Info().Msg("🌐 Starting WebUI server on :8080")
		logger.Info().Msg("📱 Open http://localhost:8080 in your browser")

		ctx := context.Background()
		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Failed to start server")
			os.Exit(1)
		}
	}()

	// Print available endpoints
	fmt.Println("\n🔗 Available endpoints:")
	fmt.Println("   📄 GET  http://localhost:8080/ - Chat interface")
	fmt.Println("   💬 POST http://localhost:8080/api/chat - Process chat messages")
	fmt.Println("   🤖 GET  http://localhost:8080/api/agents - List available agents")
	fmt.Println("   📊 GET  http://localhost:8080/api/health - Health check")
	fmt.Println("   🔌 WS   ws://localhost:8080/ws - WebSocket connection")
	fmt.Println("\n🎯 Try the web interface or use curl examples below:")
	fmt.Println("\n# Create a chat session and send a message:")
	fmt.Println("curl -X POST http://localhost:8080/api/chat \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -d '{\"message\": \"Hello, can you help me?\", \"stream\": false}'")
	fmt.Println("\n# Get available agents:")
	fmt.Println("curl http://localhost:8080/api/agents")
	fmt.Println("")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("⏹️  Shutting down server...")

	// Stop the server
	server.Stop()
	logger.Info().Msg("✅ Server shut down gracefully")
}

// ExampleAgentManager is a simple implementation for demo purposes
type ExampleAgentManager struct {
	agents map[string]core.Agent
	mutex  sync.RWMutex
}

// NewExampleAgentManager creates a new example agent manager
func NewExampleAgentManager() core.AgentManager {
	manager := &ExampleAgentManager{
		agents: make(map[string]core.Agent),
	}

	// Add some example agents
	manager.agents["assistant"] = &ExampleAgent{
		name:        "assistant",
		description: "A helpful AI assistant that can answer questions and provide guidance",
		role:        "general-assistant",
	}
	manager.agents["coder"] = &ExampleAgent{
		name:        "coder",
		description: "A specialized coding assistant that helps with programming tasks",
		role:        "code-specialist",
	}
	manager.agents["writer"] = &ExampleAgent{
		name:        "writer",
		description: "A creative writing assistant for content creation and editing",
		role:        "content-creator",
	}

	return manager
}

func (m *ExampleAgentManager) UpdateAgentConfigurations(config *core.Config) error {
	return nil
}

func (m *ExampleAgentManager) GetCurrentAgents() map[string]core.Agent {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]core.Agent)
	for k, v := range m.agents {
		result[k] = v
	}
	return result
}

func (m *ExampleAgentManager) CreateAgent(name string, config *core.ResolvedAgentConfig) (core.Agent, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	agent := &ExampleAgent{
		name:        name,
		description: config.Description,
		role:        config.Role,
	}
	m.agents[name] = agent
	return agent, nil
}

func (m *ExampleAgentManager) DisableAgent(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.agents, name)
	return nil
}

func (m *ExampleAgentManager) InitializeAgents() error {
	return nil
}

func (m *ExampleAgentManager) GetActiveAgents() []core.Agent {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	agents := make([]core.Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents
}

// ExampleAgent is a simple agent implementation for demo purposes
type ExampleAgent struct {
	name        string
	description string
	role        string
}

func (a *ExampleAgent) Name() string {
	return a.name
}

func (a *ExampleAgent) GetRole() string {
	return a.role
}

func (a *ExampleAgent) GetDescription() string {
	return a.description
}

func (a *ExampleAgent) Run(ctx context.Context, inputState core.State) (core.State, error) {
	// For demo purposes, return a simple state
	outputState := core.NewSimpleState(map[string]any{
		"agent_response": fmt.Sprintf("Agent %s processed the state", a.name),
		"timestamp":      time.Now().Format(time.RFC3339),
	})
	return outputState, nil
}

func (a *ExampleAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
	// Get message from event data
	eventData := event.GetData()
	message, ok := eventData["message"].(string)
	if !ok {
		// Try to get from state if not in event
		if msgVal, exists := state.Get("message"); exists {
			if msgStr, ok := msgVal.(string); ok {
				message = msgStr
			}
		}
		if message == "" {
			message = "No message provided"
		}
	}

	// Generate different responses based on agent type
	var response string
	switch a.name {
	case "assistant":
		response = fmt.Sprintf("👋 Hello! I'm your AI assistant. You asked: \"%s\"\n\nI'm here to help you with general questions, provide information, and offer guidance on various topics. How can I assist you today?", message)
	case "coder":
		response = fmt.Sprintf("💻 Hey there! I'm the coding specialist. You said: \"%s\"\n\nI can help you with:\n• Programming questions\n• Code review and debugging\n• Best practices and patterns\n• Technology recommendations\n\nWhat coding challenge can I help you solve?", message)
	case "writer":
		response = fmt.Sprintf("✍️ Greetings! I'm your writing assistant. You wrote: \"%s\"\n\nI specialize in:\n• Content creation and editing\n• Creative writing and storytelling\n• Grammar and style improvements\n• Writing strategies and techniques\n\nWhat writing project can I help you with?", message)
	default:
		response = fmt.Sprintf("🤖 Hi! I'm %s. You said: \"%s\"\n\nI'm ready to help you with your request. Let me know what you need!", a.name, message)
	}

	// Create output state
	outputState := core.NewSimpleState(map[string]any{
		"agent_name":    a.name,
		"agent_role":    a.role,
		"response_type": "chat",
		"timestamp":     time.Now().Format(time.RFC3339),
		"message_id":    fmt.Sprintf("msg_%d", time.Now().Unix()),
		"message":       response,
	})

	result := core.AgentResult{
		OutputState: outputState,
		Error:       "",
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		Duration:    time.Millisecond * 100,
	}
	return result, nil
}

func (a *ExampleAgent) GetCapabilities() []string {
	switch a.name {
	case "assistant":
		return []string{"general-assistance", "q-and-a", "information-lookup"}
	case "coder":
		return []string{"code-review", "debugging", "programming-help", "best-practices"}
	case "writer":
		return []string{"content-creation", "editing", "creative-writing", "grammar-check"}
	default:
		return []string{"general-purpose"}
	}
}

func (a *ExampleAgent) GetSystemPrompt() string {
	switch a.name {
	case "assistant":
		return "You are a helpful AI assistant. Provide clear, accurate, and helpful responses to user questions."
	case "coder":
		return "You are a specialized coding assistant. Help users with programming tasks, code review, and technical questions."
	case "writer":
		return "You are a creative writing assistant. Help users with content creation, editing, and writing improvement."
	default:
		return "You are a helpful AI agent. Assist users with their requests to the best of your ability."
	}
}

func (a *ExampleAgent) GetTimeout() time.Duration {
	return 30 * time.Second
}

func (a *ExampleAgent) IsEnabled() bool {
	return true
}

func (a *ExampleAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	return nil
}

func (a *ExampleAgent) Initialize(ctx context.Context) error {
	return nil
}

func (a *ExampleAgent) Shutdown(ctx context.Context) error {
	return nil
}
