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
	logger.Info().Msg("🚀 Starting AgenticGoKit WebUI Chat Demo")

	// Create core config
	config := &core.Config{}

	// Create mock agent manager with enhanced agents
	agentManager := NewEnhancedAgentManager()

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
		StaticDir:    "../../internal/webui/static", // Point to the correct static files directory
	}

	// Create WebUI server
	server := webui.NewServer(serverConfig)

	// Create session manager for agent handlers
	sessionConfig := webui.DefaultSessionConfig()
	sessionConfig.MaxSessions = 100
	sessionConfig.MaxMessages = 50
	sessionConfig.SessionTimeout = 24 * time.Hour // Long session timeout for demo

	sessionManager, err := webui.NewEnhancedSessionManager(config, sessionConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create session manager")
		os.Exit(1)
	}

	// Create agent bridge with enhanced configuration
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
		logger.Info().Msg("🌐 Starting WebUI Chat Demo server on :8080")
		logger.Info().Msg("📱 Open http://localhost:8080 in your browser to start chatting!")

		ctx := context.Background()
		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Failed to start server")
			os.Exit(1)
		}
	}()

	// Print comprehensive endpoint information
	fmt.Println("\n🔗 WebUI Chat Demo - Available Endpoints:")
	fmt.Println("==========================================")
	fmt.Println("   📄 GET  http://localhost:8080/           - Main Chat Interface")
	fmt.Println("   💬 POST http://localhost:8080/api/chat   - Send message to agent")
	fmt.Println("   🤖 GET  http://localhost:8080/api/agents - List available agents")
	fmt.Println("   📊 GET  http://localhost:8080/api/health - Health check")
	fmt.Println("   🔌 WS   ws://localhost:8080/ws          - Real-time WebSocket")
	fmt.Println("   📝 GET  http://localhost:8080/api/sessions - List chat sessions")
	fmt.Println("")

	fmt.Println("🎯 Available Agents:")
	fmt.Println("===================")
	for _, agent := range activeAgents {
		fmt.Printf("   🤖 %-12s - %s\n", agent.Name(), agent.GetDescription())
		fmt.Printf("      Role: %s\n", agent.GetRole())
		fmt.Printf("      Capabilities: %v\n", agent.GetCapabilities())
		fmt.Println("")
	}

	fmt.Println("🚀 Quick Start:")
	fmt.Println("===============")
	fmt.Println("1. Open http://localhost:8080 in your browser")
	fmt.Println("2. Select an agent from the sidebar")
	fmt.Println("3. Start chatting!")
	fmt.Println("")

	fmt.Println("📡 API Testing Examples:")
	fmt.Println("========================")
	fmt.Println("# List agents:")
	fmt.Println("curl http://localhost:8080/api/agents")
	fmt.Println("")
	fmt.Println("# Chat with assistant:")
	fmt.Println("curl -X POST http://localhost:8080/api/chat \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -d '{\"agent_name\": \"assistant\", \"message\": \"Hello!\"}'")
	fmt.Println("")
	fmt.Println("# Chat with coder:")
	fmt.Println("curl -X POST http://localhost:8080/api/chat \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -d '{\"agent_name\": \"coder\", \"message\": \"Help me with Python\"}'")
	fmt.Println("")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("⏹️  Shutting down WebUI Chat Demo...")

	// Stop the bridge gracefully
	if err := bridge.Stop(); err != nil {
		logger.Error().Err(err).Msg("Error stopping bridge")
	}

	// Stop the server
	server.Stop()
	logger.Info().Msg("✅ WebUI Chat Demo shut down gracefully")
}

// EnhancedAgentManager provides a more comprehensive agent manager for the demo
type EnhancedAgentManager struct {
	agents map[string]core.Agent
	mutex  sync.RWMutex
}

// NewEnhancedAgentManager creates a new enhanced agent manager with more realistic agents
func NewEnhancedAgentManager() core.AgentManager {
	manager := &EnhancedAgentManager{
		agents: make(map[string]core.Agent),
	}

	// Add enhanced example agents with more detailed responses
	manager.agents["assistant"] = &EnhancedAgent{
		name:        "assistant",
		description: "A helpful AI assistant that can answer questions, provide information, and help with general tasks",
		role:        "general-assistant",
		personality: "friendly and knowledgeable",
	}

	manager.agents["coder"] = &EnhancedAgent{
		name:        "coder",
		description: "A specialized coding assistant that helps with programming, debugging, and software development",
		role:        "code-specialist",
		personality: "technical and precise",
	}

	manager.agents["writer"] = &EnhancedAgent{
		name:        "writer",
		description: "A creative writing assistant for content creation, editing, and storytelling",
		role:        "content-creator",
		personality: "creative and articulate",
	}

	manager.agents["analyst"] = &EnhancedAgent{
		name:        "analyst",
		description: "A data analyst that helps with data interpretation, analysis, and insights",
		role:        "data-specialist",
		personality: "analytical and detail-oriented",
	}

	return manager
}

func (m *EnhancedAgentManager) UpdateAgentConfigurations(config *core.Config) error {
	return nil
}

func (m *EnhancedAgentManager) GetCurrentAgents() map[string]core.Agent {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]core.Agent)
	for k, v := range m.agents {
		result[k] = v
	}
	return result
}

func (m *EnhancedAgentManager) CreateAgent(name string, config *core.ResolvedAgentConfig) (core.Agent, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	agent := &EnhancedAgent{
		name:        name,
		description: config.Description,
		role:        config.Role,
		personality: "custom",
	}
	m.agents[name] = agent
	return agent, nil
}

func (m *EnhancedAgentManager) DisableAgent(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.agents, name)
	return nil
}

func (m *EnhancedAgentManager) InitializeAgents() error {
	return nil
}

func (m *EnhancedAgentManager) GetActiveAgents() []core.Agent {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	agents := make([]core.Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents
}

// EnhancedAgent provides more sophisticated agent responses
type EnhancedAgent struct {
	name        string
	description string
	role        string
	personality string
}

func (a *EnhancedAgent) Name() string {
	return a.name
}

func (a *EnhancedAgent) GetRole() string {
	return a.role
}

func (a *EnhancedAgent) GetDescription() string {
	return a.description
}

func (a *EnhancedAgent) Run(ctx context.Context, inputState core.State) (core.State, error) {
	// Enhanced state processing
	outputState := core.NewSimpleState(map[string]any{
		"agent_response": fmt.Sprintf("Enhanced agent %s processed the state with %s approach", a.name, a.personality),
		"timestamp":      time.Now().Format(time.RFC3339),
		"agent_type":     "enhanced",
	})
	return outputState, nil
}

func (a *EnhancedAgent) HandleEvent(ctx context.Context, event core.Event, state core.State) (core.AgentResult, error) {
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

	// Generate enhanced responses based on agent type and personality
	var response string
	switch a.name {
	case "assistant":
		response = a.generateAssistantResponse(message)
	case "coder":
		response = a.generateCoderResponse(message)
	case "writer":
		response = a.generateWriterResponse(message)
	case "analyst":
		response = a.generateAnalystResponse(message)
	default:
		response = fmt.Sprintf("🤖 Hello! I'm %s, your %s agent. You said: \"%s\"\n\nI'm here to help you with my specialized knowledge and %s approach. What would you like to work on together?", a.name, a.role, message, a.personality)
	}

	// Create enhanced output state
	outputState := core.NewSimpleState(map[string]any{
		"agent_name":        a.name,
		"agent_role":        a.role,
		"agent_personality": a.personality,
		"response_type":     "enhanced_chat",
		"timestamp":         time.Now().Format(time.RFC3339),
		"message_id":        fmt.Sprintf("msg_%d_%s", time.Now().Unix(), a.name),
		"message":           response,
		"capabilities":      a.GetCapabilities(),
	})

	result := core.AgentResult{
		OutputState: outputState,
		Error:       "",
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		Duration:    time.Millisecond * 150, // Slightly longer for more realistic feel
	}
	return result, nil
}

func (a *EnhancedAgent) generateAssistantResponse(message string) string {
	responses := []string{
		fmt.Sprintf("👋 Hello! I'm your friendly AI assistant. You asked: \"%s\"\n\n✨ I'm here to help you with:\n• General questions and information\n• Problem-solving guidance\n• Learning support\n• Task planning and organization\n\nWhat specific topic would you like assistance with? I'm ready to provide detailed, helpful responses!", message),
		fmt.Sprintf("🌟 Hi there! Thanks for reaching out with: \"%s\"\n\n🎯 As your AI assistant, I can help you:\n• Find accurate information\n• Break down complex problems\n• Provide step-by-step guidance\n• Offer different perspectives\n\nI'm designed to be helpful, harmless, and honest. How can I best support you today?", message),
		fmt.Sprintf("💡 Great question! You mentioned: \"%s\"\n\n🔍 Let me help you explore this topic:\n• I can provide comprehensive explanations\n• Offer practical examples\n• Suggest additional resources\n• Help you think through solutions\n\nWhat specific aspect would you like me to focus on first?", message),
	}
	return responses[time.Now().Unix()%int64(len(responses))]
}

func (a *EnhancedAgent) generateCoderResponse(message string) string {
	responses := []string{
		fmt.Sprintf("💻 Hey developer! You mentioned: \"%s\"\n\n🛠️ I'm your coding specialist, ready to help with:\n• **Programming Languages**: Python, JavaScript, Go, Java, C++, and more\n• **Debugging**: Finding and fixing code issues\n• **Best Practices**: Clean code, design patterns, optimization\n• **Architecture**: System design and code organization\n\n```\n# Let's solve this together!\nprint(\"Ready to code!\")\n```\n\nWhat coding challenge are you working on?", message),
		fmt.Sprintf("🚀 Welcome to the coding zone! Your query: \"%s\"\n\n⚡ Technical expertise at your service:\n• **Code Review**: I'll analyze your code for improvements\n• **Algorithm Help**: Optimize performance and logic\n• **Framework Guidance**: React, Node.js, Django, and more\n• **DevOps Support**: CI/CD, Docker, cloud deployment\n\n📊 **Current Stack Recommendation**: Focus on maintainable, scalable solutions\n\nShare your code or describe your technical challenge!", message),
		fmt.Sprintf("🔧 Code master here! Processing: \"%s\"\n\n💡 Ready to tackle:\n• **Bug Fixes**: Let's debug that tricky issue\n• **Feature Development**: Build it right from the start\n• **Performance Tuning**: Make it faster and more efficient\n• **Testing Strategies**: Unit tests, integration tests\n\n```python\n# Example approach\ndef solve_problem(challenge):\n    return \"Let's break it down step by step!\"\n```\n\nWhat's your current development challenge?", message),
	}
	return responses[time.Now().Unix()%int64(len(responses))]
}

func (a *EnhancedAgent) generateWriterResponse(message string) string {
	responses := []string{
		fmt.Sprintf("✍️ Greetings, fellow wordsmith! Your message: \"%s\"\n\n📚 **Creative Writing Studio Open**\n• **Content Creation**: Blog posts, articles, stories\n• **Editing & Revision**: Polish your prose to perfection\n• **Style Enhancement**: Voice, tone, and narrative flow\n• **Structure & Organization**: Compelling beginnings, satisfying endings\n\n🎨 *\"The first draft is just you telling yourself the story.\"*\n\nWhat writing project shall we craft together today?", message),
		fmt.Sprintf("📝 Hello, creative mind! You shared: \"%s\"\n\n🌟 **Your Writing Companion Ready**\n• **Storytelling**: Plot development, character arcs, world-building\n• **Professional Writing**: Reports, proposals, presentations\n• **Creative Expression**: Poetry, fiction, creative non-fiction\n• **Grammar & Style**: Clarity, concision, impact\n\n✨ Every great writer was once a beginner. Every pro was once an amateur.\n\nWhat story are you ready to tell?", message),
		fmt.Sprintf("🖋️ Welcome to the writer's sanctuary! Your words: \"%s\"\n\n📖 **Literary Services Available**\n• **Content Strategy**: Audience engagement, messaging\n• **Narrative Techniques**: Show vs. tell, dialogue, pacing\n• **Research & Fact-checking**: Credible, well-sourced content\n• **Publication Prep**: Formatting, submission guidelines\n\n🎭 *\"Writing is thinking on paper.\"*\n\nLet's transform your ideas into compelling content!", message),
	}
	return responses[time.Now().Unix()%int64(len(responses))]
}

func (a *EnhancedAgent) generateAnalystResponse(message string) string {
	responses := []string{
		fmt.Sprintf("📊 Data Analyst here! Analyzing: \"%s\"\n\n🔍 **Analytics Dashboard Activated**\n• **Data Interpretation**: Pattern recognition, trend analysis\n• **Statistical Analysis**: Correlation, regression, significance testing\n• **Visualization**: Charts, graphs, dashboards\n• **Insights Generation**: Actionable recommendations\n\n📈 Current Focus: Turning data into decisions\n📉 Key Metrics: Accuracy, relevance, actionability\n\nWhat data story shall we uncover together?", message),
		fmt.Sprintf("📈 Analytical mind engaged! Processing: \"%s\"\n\n🎯 **Research & Analysis Services**\n• **Market Research**: Competitive analysis, trend forecasting\n• **Performance Metrics**: KPIs, benchmarking, optimization\n• **Risk Assessment**: Scenario planning, impact analysis\n• **Reporting**: Executive summaries, detailed findings\n\n💡 \"In God we trust. All others must bring data.\" - W. Edwards Deming\n\nWhat analytical challenge can I help you solve?", message),
		fmt.Sprintf("🔬 Senior Analyst reporting! Your query: \"%s\"\n\n📋 **Analytical Toolkit Ready**\n• **Quantitative Analysis**: Numbers, models, forecasts\n• **Qualitative Research**: Interviews, surveys, observations\n• **Business Intelligence**: Strategic insights, growth opportunities\n• **Process Optimization**: Efficiency improvements, cost reduction\n\n🎲 Data-driven decisions lead to measurable success.\n\nWhat would you like to analyze and understand better?", message),
	}
	return responses[time.Now().Unix()%int64(len(responses))]
}

func (a *EnhancedAgent) GetCapabilities() []string {
	switch a.name {
	case "assistant":
		return []string{"general-assistance", "q-and-a", "information-lookup", "problem-solving", "learning-support"}
	case "coder":
		return []string{"code-review", "debugging", "programming-help", "best-practices", "architecture-design", "performance-optimization"}
	case "writer":
		return []string{"content-creation", "editing", "creative-writing", "grammar-check", "storytelling", "copywriting"}
	case "analyst":
		return []string{"data-analysis", "statistical-modeling", "visualization", "research", "reporting", "insights-generation"}
	default:
		return []string{"general-purpose", "enhanced-responses"}
	}
}

func (a *EnhancedAgent) GetSystemPrompt() string {
	switch a.name {
	case "assistant":
		return "You are a helpful, friendly AI assistant with a knowledgeable personality. Provide comprehensive, accurate, and engaging responses to user questions."
	case "coder":
		return "You are a technical and precise coding assistant. Help users with programming tasks, provide clean code examples, and follow best practices."
	case "writer":
		return "You are a creative and articulate writing assistant. Help users craft compelling content, improve their writing, and express ideas effectively."
	case "analyst":
		return "You are an analytical and detail-oriented data specialist. Help users understand data, generate insights, and make data-driven decisions."
	default:
		return fmt.Sprintf("You are a helpful AI agent with a %s personality. Assist users with their requests professionally and effectively.", a.personality)
	}
}

func (a *EnhancedAgent) GetTimeout() time.Duration {
	return 30 * time.Second
}

func (a *EnhancedAgent) IsEnabled() bool {
	return true
}

func (a *EnhancedAgent) GetLLMConfig() *core.ResolvedLLMConfig {
	return nil
}

func (a *EnhancedAgent) Initialize(ctx context.Context) error {
	return nil
}

func (a *EnhancedAgent) Shutdown(ctx context.Context) error {
	return nil
}
