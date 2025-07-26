package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/kunalkushwaha/agenticgokit/internal/scaffold/templates"
	"github.com/kunalkushwaha/agenticgokit/internal/scaffold/utils"
)

// Use types from the utils package for consistency
type AgentInfo = utils.AgentInfo
type ProjectConfig = utils.ProjectConfig

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

// AgentGenerator handles the generation of agent files
type AgentGenerator struct {
	template *template.Template
}

// NewAgentGenerator creates a new agent file generator
func NewAgentGenerator() (*AgentGenerator, error) {
	tmpl, err := template.New("agent").Parse(templates.AgentTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent template: %w", err)
	}

	return &AgentGenerator{template: tmpl}, nil
}

// GenerateAgentFile creates a single agent file based on the template
func (g *AgentGenerator) GenerateAgentFile(config ProjectConfig, agent AgentInfo, agentIndex int, agents []AgentInfo) error {
	// Determine next agent in the workflow chain
	var nextAgent string
	var routingComment string

	if agentIndex < len(agents)-1 {
		// Route to next agent in the list
		nextAgent = agents[agentIndex+1].Name
		routingComment = fmt.Sprintf("Route to the next agent (%s) in the workflow", agents[agentIndex+1].DisplayName)
	} else if config.ResponsibleAI {
		// Last agent routes to responsible AI
		nextAgent = "responsible_ai"
		routingComment = "Route to Responsible AI for final content check"
	} else {
		// Route to workflow finalizer to complete the workflow
		nextAgent = "workflow_finalizer"
		routingComment = "Route to workflow finalizer to complete the workflow"
	}

	// Create system prompt for this agent
	systemPrompt := utils.CreateSystemPrompt(agent, agentIndex, len(agents), config.OrchestrationMode)

	// Create template data
	templateData := TemplateData{
		Config:         config,
		Agent:          agent,
		Agents:         agents,
		AgentIndex:     agentIndex,
		TotalAgents:    len(agents),
		NextAgent:      nextAgent,
		IsFirstAgent:   agentIndex == 0,
		IsLastAgent:    agentIndex == len(agents)-1,
		SystemPrompt:   systemPrompt,
		RoutingComment: routingComment,
	}

	// Generate file content using template
	filePath := filepath.Join(config.Name, agent.FileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	if err := g.template.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute template for %s: %w", agent.FileName, err)
	}

	fmt.Printf("Created file: %s (%s agent)\n", filePath, agent.DisplayName)
	return nil
}

// GenerateAllAgentFiles creates all agent files for the project
func (g *AgentGenerator) GenerateAllAgentFiles(config ProjectConfig) error {
	// Resolve agent names based on config
	agents := utils.ResolveAgentNames(config)

	if len(agents) == 0 {
		// Fallback - create at least one agent
		agents = append(agents, utils.CreateAgentInfo("agent1", config.OrchestrationMode))
	}

	// Validate agent names
	if err := utils.ValidateAgentNames(agents); err != nil {
		return fmt.Errorf("agent name validation failed: %w", err)
	}

	// Generate each agent file
	for i, agent := range agents {
		if err := g.GenerateAgentFile(config, agent, i, agents); err != nil {
			return fmt.Errorf("failed to generate agent file for %s: %w", agent.Name, err)
		}
	}

	return nil
}
