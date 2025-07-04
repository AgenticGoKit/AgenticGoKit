package utils

import (
	"fmt"
	"strings"
)

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
			// Fallback
			agents = append(agents, CreateAgentInfo("agent1", "loop"))
		}

	case "mixed":
		// Combine collaborative and sequential agents
		for _, name := range config.CollaborativeAgents {
			agents = append(agents, CreateAgentInfo(name, "collaborative"))
		}
		for _, name := range config.SequentialAgents {
			agents = append(agents, CreateAgentInfo(name, "sequential"))
		}
		// If no named agents, fallback to numbered
		if len(agents) == 0 {
			for i := 1; i <= config.NumAgents; i++ {
				agents = append(agents, CreateAgentInfo(fmt.Sprintf("agent%d", i), "mixed"))
			}
		}

	default: // "route" mode
		if len(config.CollaborativeAgents) > 0 {
			for _, name := range config.CollaborativeAgents {
				agents = append(agents, CreateAgentInfo(name, "route"))
			}
		} else if len(config.SequentialAgents) > 0 {
			for _, name := range config.SequentialAgents {
				agents = append(agents, CreateAgentInfo(name, "route"))
			}
		} else if config.LoopAgent != "" {
			agents = append(agents, CreateAgentInfo(config.LoopAgent, "route"))
		} else {
			// Fallback to numbered agents
			for i := 1; i <= config.NumAgents; i++ {
				agents = append(agents, CreateAgentInfo(fmt.Sprintf("agent%d", i), "route"))
			}
		}
	}

	return agents
}

// CreateAgentInfo creates AgentInfo from a name and mode
func CreateAgentInfo(name, mode string) AgentInfo {
	cleanName := strings.TrimSpace(strings.ToLower(name))

	// Sanitize name for Go identifiers
	sanitizedName := SanitizeIdentifier(cleanName)

	return AgentInfo{
		Name:        sanitizedName,
		FileName:    sanitizedName + ".go",
		DisplayName: CapitalizeFirst(sanitizedName),
		Purpose:     InferPurpose(sanitizedName, mode),
		Role:        mode,
	}
}

// SanitizeIdentifier ensures the name is a valid Go identifier
func SanitizeIdentifier(name string) string {
	// Replace invalid characters with underscores
	result := ""
	for i, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (i > 0 && r >= '0' && r <= '9') || r == '_' {
			result += string(r)
		} else {
			result += "_"
		}
	}

	// Ensure it starts with a letter
	if len(result) == 0 || (result[0] >= '0' && result[0] <= '9') {
		result = "agent_" + result
	}

	return result
}

// CapitalizeFirst capitalizes the first letter of a string
func CapitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// InferPurpose creates a purpose description based on the agent name and mode
func InferPurpose(name, mode string) string {
	// Common purposes based on name patterns
	purposeMap := map[string]string{
		"analyzer":         "Analyzes and processes input data to extract insights",
		"processor":        "Processes and transforms data through various operations",
		"validator":        "Validates results and ensures quality and accuracy",
		"transformer":      "Transforms data from one format to another",
		"enricher":         "Enriches data with additional context and information",
		"filter":           "Filters and selects relevant information from input",
		"summarizer":       "Summarizes and condenses information into key points",
		"router":           "Routes requests to appropriate handlers or systems",
		"monitor":          "Monitors system state and performance metrics",
		"collector":        "Collects data from various sources and systems",
		"detector":         "Detects patterns, anomalies, or specific conditions",
		"optimizer":        "Optimizes processes and improves efficiency",
		"coordinator":      "Coordinates activities between different components",
		"scheduler":        "Schedules and manages task execution timing",
		"integrator":       "Integrates data and functionality from multiple sources",
		"researcher":       "Researches topics and gathers comprehensive information",
		"data_collector":   "Collects and aggregates data from multiple sources",
		"report_generator": "Generates comprehensive reports and summaries",
	}

	if purpose, exists := purposeMap[name]; exists {
		return purpose
	}

	// Fallback purposes based on mode
	switch mode {
	case "collaborative":
		return "Collaborates with other agents to process tasks in parallel"
	case "sequential":
		return "Processes tasks in sequence as part of a processing pipeline"
	case "loop":
		return "Iteratively processes tasks until completion criteria are met"
	default:
		return "Handles specialized task processing within the workflow"
	}
}

// ValidateAgentNames checks for naming conflicts and invalid names
func ValidateAgentNames(agents []AgentInfo) error {
	nameSet := make(map[string]bool)

	for _, agent := range agents {
		if nameSet[agent.Name] {
			return fmt.Errorf("duplicate agent name: %s", agent.Name)
		}
		nameSet[agent.Name] = true

		// Check for reserved names
		if isReservedName(agent.Name) {
			return fmt.Errorf("agent name '%s' is reserved", agent.Name)
		}
	}

	return nil
}

// isReservedName checks if a name is reserved by the system
func isReservedName(name string) bool {
	reserved := []string{
		"main", "config", "error", "handler", "responsible_ai",
		"error_handler", "workflow_finalizer", "state", "event",
	}

	for _, r := range reserved {
		if name == r {
			return true
		}
	}

	return false
}
