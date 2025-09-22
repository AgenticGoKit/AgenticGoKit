package utils

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// AgentInfo represents information about an agent
type AgentInfo struct {
	Name        string
	FileName    string
	DisplayName string
	Purpose     string
	Role        string
}

// ProjectConfig represents the configuration for a scaffold project
type ProjectConfig struct {
	Name               string
	NumAgents          int
	Provider           string
	ResponsibleAI      bool
	ErrorHandler       bool
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

	// Orchestration configuration
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

	// Memory/RAG configuration
	MemoryEnabled       bool
	MemoryProvider      string // inmemory, pgvector, weaviate
	EmbeddingProvider   string // openai, dummy
	EmbeddingModel      string // text-embedding-3-small, etc.
	EmbeddingDimensions int    // Auto-calculated based on embedding model
	RAGEnabled          bool
	RAGChunkSize        int
	RAGOverlap          int
	RAGTopK             int
	RAGScoreThreshold   float64
	HybridSearch        bool
	SessionMemory       bool

	// MCP configuration
	MCPTransport string // tcp, stdio, etc.
}

// ResolveAgentNames determines the final list of agents to create based on the config
func ResolveAgentNames(config ProjectConfig) []AgentInfo {
	var agents []AgentInfo

	switch config.OrchestrationMode {
	case "collaborative":
		if len(config.CollaborativeAgents) > 0 {
			for _, name := range config.CollaborativeAgents {
				agents = append(agents, CreateAgentInfo(name, "collaborative"))
			}
		} else {
			// Fallback to numbered agents
			for i := 1; i <= config.NumAgents; i++ {
				agents = append(agents, CreateAgentInfo(fmt.Sprintf("agent%d", i), "collaborative"))
			}
		}

	case "sequential":
		if len(config.SequentialAgents) > 0 {
			for _, name := range config.SequentialAgents {
				agents = append(agents, CreateAgentInfo(name, "sequential"))
			}
		} else {
			// Fallback to numbered agents
			for i := 1; i <= config.NumAgents; i++ {
				agents = append(agents, CreateAgentInfo(fmt.Sprintf("agent%d", i), "sequential"))
			}
		}

	case "loop":
		if config.LoopAgent != "" {
			agents = append(agents, CreateAgentInfo(config.LoopAgent, "loop"))
		} else {
			agents = append(agents, CreateAgentInfo("agent1", "loop"))
		}

	case "mixed":
		// Add collaborative agents first
		for _, name := range config.CollaborativeAgents {
			agents = append(agents, CreateAgentInfo(name, "collaborative"))
		}
		// Add sequential agents
		for _, name := range config.SequentialAgents {
			agents = append(agents, CreateAgentInfo(name, "sequential"))
		}

	default:
		// Route mode or unknown - create numbered agents
		for i := 1; i <= config.NumAgents; i++ {
			agents = append(agents, CreateAgentInfo(fmt.Sprintf("agent%d", i), config.OrchestrationMode))
		}
	}

	return agents
}

// CreateAgentInfo creates an AgentInfo from a name and role
func CreateAgentInfo(name, role string) AgentInfo {
	// Convert name to display name (capitalize and clean up)
	titleCaser := cases.Title(language.English)
	displayName := titleCaser.String(strings.ReplaceAll(name, "_", " "))
	displayName = strings.ReplaceAll(displayName, "-", " ")

	// Generate filename
	fileName := fmt.Sprintf("%s.go", name)

	// Infer purpose from name
	purpose := InferPurpose(name)

	return AgentInfo{
		Name:        name,
		FileName:    fileName,
		DisplayName: displayName,
		Purpose:     purpose,
		Role:        role,
	}
}

// InferPurpose infers the purpose of an agent based on its name
func InferPurpose(name string) string {
	name = strings.ToLower(name)

	switch {
	case strings.Contains(name, "document") && strings.Contains(name, "ingester"):
		return "Ingests and processes documents for the knowledge base"
	case strings.Contains(name, "query") && strings.Contains(name, "processor"):
		return "Analyzes and optimizes user queries for retrieval"
	case strings.Contains(name, "response") && strings.Contains(name, "generator"):
		return "Generates comprehensive responses using retrieved information"
	case strings.Contains(name, "research"):
		return "Researches topics and gathers comprehensive information"
	case strings.Contains(name, "analyzer") || strings.Contains(name, "analysis"):
		return "Analyzes and processes input data to extract insights"
	case strings.Contains(name, "synthesizer") || strings.Contains(name, "synthesis"):
		return "Synthesizes information and creates comprehensive responses"
	case strings.Contains(name, "writer") || strings.Contains(name, "content"):
		return "Creates and formats written content"
	case strings.Contains(name, "reviewer") || strings.Contains(name, "validator"):
		return "Reviews and validates content for accuracy and quality"
	case strings.Contains(name, "processor"):
		return "Processes and transforms data"
	case strings.Contains(name, "collector"):
		return "Collects and organizes information"
	case strings.Contains(name, "coordinator") || strings.Contains(name, "manager"):
		return "Coordinates workflow and manages tasks"
	case strings.Contains(name, "ingester"):
		return "Ingests and preprocesses data"
	case strings.Contains(name, "outputter"):
		return "Formats and outputs final results"
	case strings.Contains(name, "fact"):
		return "Verifies facts and checks information accuracy"
	default:
		return "Provides general assistance and processing capabilities"
	}
}

// ValidateAgentNames validates agent names for compliance
func ValidateAgentNames(agents []AgentInfo) error {
	if len(agents) == 0 {
		return fmt.Errorf("no agents defined")
	}

	nameMap := make(map[string]bool)
	for _, agent := range agents {
		if agent.Name == "" {
			return fmt.Errorf("agent name cannot be empty")
		}

		if nameMap[agent.Name] {
			return fmt.Errorf("duplicate agent name: %s", agent.Name)
		}
		nameMap[agent.Name] = true

		// Validate naming convention
		if !isValidAgentName(agent.Name) {
			return fmt.Errorf("invalid agent name '%s': must contain only lowercase letters, numbers, underscores, and hyphens", agent.Name)
		}
	}

	return nil
}

// isValidAgentName checks if an agent name follows the naming convention
func isValidAgentName(name string) bool {
	if name == "" {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' || char == '-') {
			return false
		}
	}

	return true
}
