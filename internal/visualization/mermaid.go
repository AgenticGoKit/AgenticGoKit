// Package visualization provides internal visualization functionality for AgenticGoKit.
package visualization

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// init registers the MermaidGenerator implementation with the core package
func init() {
	core.RegisterMermaidGeneratorFactory(func() core.MermaidGenerator {
		return &MermaidGeneratorImpl{}
	})
}

// MermaidGeneratorImpl implements the core.MermaidGenerator interface
type MermaidGeneratorImpl struct{}

// GenerateCompositionDiagram generates a Mermaid diagram for agent composition
func (mg *MermaidGeneratorImpl) GenerateCompositionDiagram(mode, name string, agents []core.Agent, config core.MermaidConfig) string {
	var builder strings.Builder

	// Header with title and config
	if config.Title != "" {
		builder.WriteString(fmt.Sprintf("---\ntitle: \"%s\"\n---\n", config.Title))
	}

	// Diagram type and direction
	builder.WriteString(fmt.Sprintf("%s %s\n", config.DiagramType, config.Direction))

	// Input node
	builder.WriteString("    INPUT[\"🚀 Input\"] --> ORCHESTRATOR{\"Orchestrator\"}\n")

	// Agent nodes
	for i, agent := range agents {
		agentID := fmt.Sprintf("AGENT_%d", i)
		agentLabel := agent.Name()

		if config.ShowAgentTypes {
			agentLabel = fmt.Sprintf("%s\\n[Agent]", agentLabel)
		}

		builder.WriteString(fmt.Sprintf("    ORCHESTRATOR --> %s[\"%s\"]\n", agentID, agentLabel))

		if config.ShowMetadata {
			builder.WriteString(fmt.Sprintf("    %s --> PROCESSING_%d[\"⚙️ Processing\"]\n", agentID, i))
			builder.WriteString(fmt.Sprintf("    PROCESSING_%d --> OUTPUT_%d[\"📤 Output\"]\n", i, i))
		} else {
			builder.WriteString(fmt.Sprintf("    %s --> OUTPUT_%d[\"📤 Output\"]\n", agentID, i))
		}
	}

	// Final output aggregation
	if len(agents) > 1 {
		builder.WriteString("    ")
		for i := range agents {
			if i > 0 {
				builder.WriteString(" & ")
			}
			builder.WriteString(fmt.Sprintf("OUTPUT_%d", i))
		}
		builder.WriteString(" --> FINAL_OUTPUT[\"✅ Final Result\"]\n")
	}

	// Styling
	if config.Theme != "default" {
		builder.WriteString(fmt.Sprintf("    %%{theme: %s}%%\n", config.Theme))
	}

	return builder.String()
}

// GenerateOrchestrationDiagram generates a Mermaid diagram for orchestration patterns
func (mg *MermaidGeneratorImpl) GenerateOrchestrationDiagram(mode core.OrchestrationMode, agents map[string]core.AgentHandler, config core.MermaidConfig) string {
	var builder strings.Builder

	// Header
	if config.Title != "" {
		builder.WriteString(fmt.Sprintf("---\ntitle: \"%s\"\n---\n", config.Title))
	}

	builder.WriteString(fmt.Sprintf("%s %s\n", config.DiagramType, config.Direction))

	// Input
	builder.WriteString("    INPUT[\"🚀 Input Event\"] --> ORCHESTRATOR{\"" + string(mode) + " Orchestrator\"}\n")

	switch mode {
	case core.OrchestrationRoute:
		mg.generateRoutePattern(&builder, agents, config)
	case core.OrchestrationCollaborate:
		mg.generateCollaboratePattern(&builder, agents, config)
	case core.OrchestrationSequential:
		mg.generateSequentialPattern(&builder, agents, config)
	case core.OrchestrationLoop:
		mg.generateLoopPattern(&builder, agents, config)
	case core.OrchestrationMixed:
		mg.generateMixedPattern(&builder, agents, config)
	default:
		mg.generateRoutePattern(&builder, agents, config)
	}

	return builder.String()
}

// GenerateAgentDiagram generates a detailed diagram for a single agent
func (mg *MermaidGeneratorImpl) GenerateAgentDiagram(agent core.Agent, config core.MermaidConfig) string {
	var builder strings.Builder

	if config.Title != "" {
		builder.WriteString(fmt.Sprintf("---\ntitle: \"%s\"\n---\n", config.Title))
	}

	builder.WriteString(fmt.Sprintf("%s %s\n", config.DiagramType, config.Direction))

	agentName := agent.Name()

	// Agent flow
	builder.WriteString(fmt.Sprintf("    INPUT[\"📥 Input\"] --> VALIDATE[\"✅ Validate\"]\n"))
	builder.WriteString(fmt.Sprintf("    VALIDATE --> PROCESS[\"%s\\n🤖 Process\"]\n", agentName))

	if config.ShowMetadata {
		builder.WriteString(fmt.Sprintf("    PROCESS --> METRICS[\"📊 Metrics\"]\n"))
		builder.WriteString(fmt.Sprintf("    METRICS --> OUTPUT[\"📤 Output\"]\n"))
	} else {
		builder.WriteString(fmt.Sprintf("    PROCESS --> OUTPUT[\"📤 Output\"]\n"))
	}

	// Error handling
	builder.WriteString(fmt.Sprintf("    VALIDATE -->|Error| ERROR[\"❌ Error Handler\"]\n"))
	builder.WriteString(fmt.Sprintf("    PROCESS -->|Error| ERROR\n"))
	builder.WriteString(fmt.Sprintf("    ERROR --> OUTPUT\n"))

	return builder.String()
}

// GenerateWorkflowPatternDiagram generates a diagram for specific workflow patterns
func (mg *MermaidGeneratorImpl) GenerateWorkflowPatternDiagram(patternName string, agents []core.Agent, config core.MermaidConfig) string {
	var builder strings.Builder

	if config.Title != "" {
		builder.WriteString(fmt.Sprintf("---\ntitle: \"%s\"\n---\n", config.Title))
	}

	builder.WriteString(fmt.Sprintf("%s %s\n", config.DiagramType, config.Direction))

	switch strings.ToLower(patternName) {
	case "pipeline":
		mg.generatePipelinePattern(&builder, agents, config)
	case "fork-join":
		mg.generateForkJoinPattern(&builder, agents, config)
	case "map-reduce":
		mg.generateMapReducePattern(&builder, agents, config)
	default:
		mg.generateGenericPattern(&builder, agents, config)
	}

	return builder.String()
}

// SaveDiagramAsMarkdown saves a diagram as a markdown file
func (mg *MermaidGeneratorImpl) SaveDiagramAsMarkdown(filename, title, diagram string) error {
	content := fmt.Sprintf("# %s\n\n```mermaid\n%s\n```\n", title, diagram)
	return os.WriteFile(filename, []byte(content), 0644)
}

// SaveDiagramWithMetadata saves a diagram with additional metadata
func (mg *MermaidGeneratorImpl) SaveDiagramWithMetadata(filename, title, description, diagram string, metadata map[string]interface{}) error {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s\n\n", title))

	if description != "" {
		builder.WriteString(fmt.Sprintf("%s\n\n", description))
	}

	// Metadata section
	if len(metadata) > 0 {
		builder.WriteString("## Metadata\n\n")
		for key, value := range metadata {
			builder.WriteString(fmt.Sprintf("- **%s**: %v\n", key, value))
		}
		builder.WriteString("\n")
	}

	// Timestamp
	builder.WriteString(fmt.Sprintf("*Generated: %s*\n\n", time.Now().Format(time.RFC3339)))

	// Diagram
	builder.WriteString(fmt.Sprintf("```mermaid\n%s\n```\n", diagram))

	return os.WriteFile(filename, []byte(builder.String()), 0644)
}

// Helper methods for different orchestration patterns

func (mg *MermaidGeneratorImpl) generateRoutePattern(builder *strings.Builder, agents map[string]core.AgentHandler, config core.MermaidConfig) {
	for name := range agents {
		agentID := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
		builder.WriteString(fmt.Sprintf("    ORCHESTRATOR -->|route=%s| %s[\"%s\"]\n", name, agentID, name))
		builder.WriteString(fmt.Sprintf("    %s --> OUTPUT[\"📤 Result\"]\n", agentID))
	}
}

func (mg *MermaidGeneratorImpl) generateCollaboratePattern(builder *strings.Builder, agents map[string]core.AgentHandler, config core.MermaidConfig) {
	// All agents run in parallel
	for name := range agents {
		agentID := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
		builder.WriteString(fmt.Sprintf("    ORCHESTRATOR --> %s[\"%s\"]\n", agentID, name))
	}

	// Collect results
	builder.WriteString("    ")
	i := 0
	for name := range agents {
		agentID := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
		if i > 0 {
			builder.WriteString(" & ")
		}
		builder.WriteString(agentID)
		i++
	}
	builder.WriteString(" --> COLLECT[\"🔗 Collect Results\"] --> OUTPUT[\"📤 Final Result\"]\n")
}

func (mg *MermaidGeneratorImpl) generateSequentialPattern(builder *strings.Builder, agents map[string]core.AgentHandler, config core.MermaidConfig) {
	var prevNode = "ORCHESTRATOR"
	i := 0
	for name := range agents {
		agentID := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
		builder.WriteString(fmt.Sprintf("    %s --> %s[\"%s\"]\n", prevNode, agentID, name))
		prevNode = agentID
		i++
	}
	builder.WriteString(fmt.Sprintf("    %s --> OUTPUT[\"📤 Result\"]\n", prevNode))
}

func (mg *MermaidGeneratorImpl) generateLoopPattern(builder *strings.Builder, agents map[string]core.AgentHandler, config core.MermaidConfig) {
	// Pick first agent for loop (in real implementation, this would be configurable)
	var firstAgent string
	for name := range agents {
		firstAgent = name
		break
	}

	if firstAgent != "" {
		agentID := strings.ToUpper(strings.ReplaceAll(firstAgent, "-", "_"))
		builder.WriteString(fmt.Sprintf("    ORCHESTRATOR --> %s[\"%s\"]\n", agentID, firstAgent))
		builder.WriteString(fmt.Sprintf("    %s --> CONDITION{\"Continue?\"}\n", agentID))
		builder.WriteString(fmt.Sprintf("    CONDITION -->|Yes| %s\n", agentID))
		builder.WriteString(fmt.Sprintf("    CONDITION -->|No| OUTPUT[\"📤 Result\"]\n"))
	}
}

func (mg *MermaidGeneratorImpl) generateMixedPattern(builder *strings.Builder, agents map[string]core.AgentHandler, config core.MermaidConfig) {
	// Simplified mixed pattern - collaborative then sequential
	builder.WriteString("    ORCHESTRATOR --> COLLAB_PHASE[\"🤝 Collaborative Phase\"]\n")
	builder.WriteString("    COLLAB_PHASE --> SEQ_PHASE[\"⏭️ Sequential Phase\"]\n")
	builder.WriteString("    SEQ_PHASE --> OUTPUT[\"📤 Result\"]\n")
}

// Helper methods for workflow patterns

func (mg *MermaidGeneratorImpl) generatePipelinePattern(builder *strings.Builder, agents []core.Agent, config core.MermaidConfig) {
	builder.WriteString("    INPUT[\"📥 Input\"]")
	for i, agent := range agents {
		agentID := fmt.Sprintf("STAGE_%d", i+1)
		builder.WriteString(fmt.Sprintf(" --> %s[\"%s\"]\n", agentID, agent.Name()))
		if i < len(agents)-1 {
			builder.WriteString(fmt.Sprintf("    %s", agentID))
		}
	}
	builder.WriteString(fmt.Sprintf(" --> OUTPUT[\"📤 Output\"]\n"))
}

func (mg *MermaidGeneratorImpl) generateForkJoinPattern(builder *strings.Builder, agents []core.Agent, config core.MermaidConfig) {
	builder.WriteString("    INPUT[\"📥 Input\"] --> FORK{\"🍴 Fork\"}\n")

	for i, agent := range agents {
		agentID := fmt.Sprintf("BRANCH_%d", i+1)
		builder.WriteString(fmt.Sprintf("    FORK --> %s[\"%s\"]\n", agentID, agent.Name()))
	}

	builder.WriteString("    ")
	for i := range agents {
		if i > 0 {
			builder.WriteString(" & ")
		}
		builder.WriteString(fmt.Sprintf("BRANCH_%d", i+1))
	}
	builder.WriteString(" --> JOIN{\"🔗 Join\"} --> OUTPUT[\"📤 Output\"]\n")
}

func (mg *MermaidGeneratorImpl) generateMapReducePattern(builder *strings.Builder, agents []core.Agent, config core.MermaidConfig) {
	if len(agents) >= 2 {
		mapAgent := agents[0]
		reduceAgent := agents[len(agents)-1]

		builder.WriteString(fmt.Sprintf("    INPUT[\"📥 Input\"] --> MAP[\"%s\\n(Map)\"]\n", mapAgent.Name()))
		builder.WriteString("    MAP --> PARALLEL{\"🔀 Parallel Processing\"}\n")

		// Middle agents as parallel processors
		for i := 1; i < len(agents)-1; i++ {
			agentID := fmt.Sprintf("PROC_%d", i)
			builder.WriteString(fmt.Sprintf("    PARALLEL --> %s[\"%s\"]\n", agentID, agents[i].Name()))
		}

		builder.WriteString("    ")
		for i := 1; i < len(agents)-1; i++ {
			if i > 1 {
				builder.WriteString(" & ")
			}
			builder.WriteString(fmt.Sprintf("PROC_%d", i))
		}
		builder.WriteString(fmt.Sprintf(" --> REDUCE[\"%s\\n(Reduce)\"] --> OUTPUT[\"📤 Output\"]\n", reduceAgent.Name()))
	} else {
		mg.generateGenericPattern(builder, agents, config)
	}
}

func (mg *MermaidGeneratorImpl) generateGenericPattern(builder *strings.Builder, agents []core.Agent, config core.MermaidConfig) {
	builder.WriteString("    INPUT[\"📥 Input\"]")
	for i, agent := range agents {
		agentID := fmt.Sprintf("AGENT_%d", i+1)
		builder.WriteString(fmt.Sprintf(" --> %s[\"%s\"]\n", agentID, agent.Name()))
		if i < len(agents)-1 {
			builder.WriteString(fmt.Sprintf("    %s", agentID))
		}
	}
	builder.WriteString(" --> OUTPUT[\"📤 Output\"]\n")
}
