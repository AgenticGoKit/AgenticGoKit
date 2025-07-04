package scaffold

// AgentInfo represents information about an agent including its name and purpose
type AgentInfo struct {
	Name        string // User-defined name like "analyzer", "processor"
	FileName    string // File name like "analyzer.go"
	DisplayName string // Capitalized name like "Analyzer"
	Purpose     string // Brief description of the agent's purpose
	Role        string // Agent role like "collaborative", "sequential", "loop"
}

// ProjectConfig represents the configuration for creating a new AgentFlow project
type ProjectConfig struct {
	Name          string
	NumAgents     int
	Provider      string
	ResponsibleAI bool
	ErrorHandler  bool

	// MCP configuration
	MCPEnabled         bool
	MCPProduction      bool
	WithCache          bool
	WithMetrics        bool
	MCPTools           []string
	MCPServers         []string
	CacheBackend       string
	MetricsPort        int
	WithLoadBalancer   bool
	ConnectionPoolSize int
	RetryPolicy        string

	// Multi-agent orchestration configuration
	OrchestrationMode    string
	CollaborativeAgents  []string
	SequentialAgents     []string
	LoopAgent            string
	MaxIterations        int
	OrchestrationTimeout int
	FailureThreshold     float64
	MaxConcurrency       int

	// Visualization configuration
	Visualize          bool
	VisualizeOutputDir string
}

// TemplateData represents the data structure passed to templates
type TemplateData struct {
	Config         ProjectConfig
	Agent          AgentInfo
	Agents         []AgentInfo
	AgentIndex     int
	TotalAgents    int
	NextAgent      string
	PrevAgent      string
	IsFirstAgent   bool
	IsLastAgent    bool
	SystemPrompt   string
	RoutingComment string
}
