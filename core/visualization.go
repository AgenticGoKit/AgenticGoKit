// Package core provides workflow visualization capabilities for multi-agent systems
package core

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// =============================================================================
// MERMAID DIAGRAM GENERATION
// =============================================================================

// MermaidDiagramType defines the type of Mermaid diagram to generate
type MermaidDiagramType string

const (
	MermaidFlowchart    MermaidDiagramType = "flowchart"
	MermaidSequence     MermaidDiagramType = "sequenceDiagram"
	MermaidStateDiagram MermaidDiagramType = "stateDiagram"
	MermaidTimeline     MermaidDiagramType = "timeline"
)

// MermaidConfig configures diagram generation options
type MermaidConfig struct {
	DiagramType    MermaidDiagramType
	Title          string
	Direction      string // TB (top-bottom), LR (left-right), etc.
	Theme          string // default, dark, forest, etc.
	ShowMetadata   bool   // Include metadata like timeouts, error strategies
	ShowAgentTypes bool   // Show agent type information
	CompactMode    bool   // Generate more compact diagrams
}

// DefaultMermaidConfig returns sensible defaults for Mermaid diagram generation
func DefaultMermaidConfig() MermaidConfig {
	return MermaidConfig{
		DiagramType:    MermaidFlowchart,
		Direction:      "TD", // Top-Down
		Theme:          "default",
		ShowMetadata:   true,
		ShowAgentTypes: true,
		CompactMode:    false,
	}
}

// =============================================================================
// COMPOSITION BUILDER VISUALIZATION
// =============================================================================

// GenerateMermaidDiagram generates a Mermaid diagram for the composition
func (cb *CompositionBuilder) GenerateMermaidDiagram() string {
	return cb.GenerateMermaidDiagramWithConfig(DefaultMermaidConfig())
}

// GenerateMermaidDiagramWithConfig generates a Mermaid diagram with custom configuration
func (cb *CompositionBuilder) GenerateMermaidDiagramWithConfig(config MermaidConfig) string {
	var diagram strings.Builder

	// Add diagram header with properly quoted title
	title := cb.getTitle(config)
	diagram.WriteString(fmt.Sprintf("---\ntitle: \"%s\"\n---\n", title))
	diagram.WriteString(fmt.Sprintf("flowchart %s\n", config.Direction))

	// Generate nodes and connections based on composition mode
	switch cb.mode {
	case "parallel":
		cb.generateParallelDiagram(&diagram, config)
	case "sequential":
		cb.generateSequentialDiagram(&diagram, config)
	case "loop":
		cb.generateLoopDiagram(&diagram, config)
	default:
		cb.generateDefaultDiagram(&diagram, config)
	}

	// Add metadata if enabled
	if config.ShowMetadata {
		cb.addMetadataToMermaid(&diagram)
	}

	// Add styling
	cb.addStylingToMermaid(&diagram, config)

	return diagram.String()
}

// generateParallelDiagram creates a parallel composition diagram
func (cb *CompositionBuilder) generateParallelDiagram(diagram *strings.Builder, config MermaidConfig) {
	diagram.WriteString("    INPUT[\"🎯 Input Event\"]\n")
	diagram.WriteString("    FANOUT[\"📤 Fan-Out\"]\n")
	diagram.WriteString("    MERGE[\"🔄 Merge Results\"]\n")
	diagram.WriteString("    OUTPUT[\"✅ Final Output\"]\n\n")

	// Connect input to fan-out
	diagram.WriteString("    INPUT --> FANOUT\n")

	// Create agent nodes and connect them
	for i, agent := range cb.agents {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		agentName := cb.getAgentDisplayName(agent, i+1)

		if config.ShowAgentTypes {
			diagram.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", agentId, agentName))
		} else {
			diagram.WriteString(fmt.Sprintf("    %s[%s]\n", agentId, agent.Name()))
		}

		diagram.WriteString(fmt.Sprintf("    FANOUT --> %s\n", agentId))
		diagram.WriteString(fmt.Sprintf("    %s --> MERGE\n", agentId))
	}

	// Connect merge to output
	diagram.WriteString("    MERGE --> OUTPUT\n")
}

// generateSequentialDiagram creates a sequential composition diagram
func (cb *CompositionBuilder) generateSequentialDiagram(diagram *strings.Builder, config MermaidConfig) {
	diagram.WriteString("    INPUT[\"🎯 Input Event\"]\n")

	var prevNode = "INPUT"
	for i, agent := range cb.agents {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		agentName := cb.getAgentDisplayName(agent, i+1)

		if config.ShowAgentTypes {
			diagram.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", agentId, agentName))
		} else {
			diagram.WriteString(fmt.Sprintf("    %s[%s]\n", agentId, agent.Name()))
		}

		diagram.WriteString(fmt.Sprintf("    %s --> %s\n", prevNode, agentId))
		prevNode = agentId
	}

	diagram.WriteString("    OUTPUT[\"✅ Final Output\"]\n")
	diagram.WriteString(fmt.Sprintf("    %s --> OUTPUT\n", prevNode))
}

// generateLoopDiagram creates a loop composition diagram
func (cb *CompositionBuilder) generateLoopDiagram(diagram *strings.Builder, config MermaidConfig) {
	if len(cb.agents) == 0 {
		return
	}

	agent := cb.agents[0]
	agentName := cb.getAgentDisplayName(agent, 1)

	diagram.WriteString("    INPUT[\"🎯 Input Event\"]\n")
	diagram.WriteString("    CONDITION{\"🔍 Check Condition\"}\n")
	diagram.WriteString(fmt.Sprintf("    AGENT[%s]\n", agentName))
	diagram.WriteString("    OUTPUT[\"✅ Final Output\"]\n")
	diagram.WriteString("    MAXITER[\"⚠️ Max Iterations\"]\n\n")

	// Create the loop flow
	diagram.WriteString("    INPUT --> CONDITION\n")
	diagram.WriteString("    CONDITION -->|\"Continue\"| AGENT\n")
	diagram.WriteString("    CONDITION -->|\"Stop\"| OUTPUT\n")
	diagram.WriteString("    AGENT --> CONDITION\n")
	diagram.WriteString("    CONDITION -->|\"Max Reached\"| MAXITER\n")
	diagram.WriteString("    MAXITER --> OUTPUT\n")
}

// generateDefaultDiagram creates a default diagram when mode is not set
func (cb *CompositionBuilder) generateDefaultDiagram(diagram *strings.Builder, config MermaidConfig) {
	diagram.WriteString("    INPUT[\"🎯 Input Event\"]\n")
	diagram.WriteString("    CONFIG[\"⚙️ Configuration Needed\"]\n")
	diagram.WriteString("    OUTPUT[\"❓ Unknown Output\"]\n\n")

	diagram.WriteString("    INPUT --> CONFIG\n")
	diagram.WriteString("    CONFIG --> OUTPUT\n")

	// Add note about configuration
	diagram.WriteString("    CONFIG -.->|\"Use .AsParallel(), .AsSequential(), or .AsLoop()\"| OUTPUT\n")
}

// =============================================================================
// ORCHESTRATION BUILDER VISUALIZATION
// =============================================================================

// GenerateMermaidDiagram generates a Mermaid diagram for the orchestration
func (ob *OrchestrationBuilder) GenerateMermaidDiagram() string {
	return ob.GenerateMermaidDiagramWithConfig(DefaultMermaidConfig())
}

// GenerateMermaidDiagramWithConfig generates a Mermaid diagram with custom configuration
func (ob *OrchestrationBuilder) GenerateMermaidDiagramWithConfig(config MermaidConfig) string {
	var diagram strings.Builder

	// Add diagram header
	title := fmt.Sprintf("%s Orchestration", strings.Title(string(ob.mode)))
	if config.Title != "" {
		title = config.Title
	}

	diagram.WriteString(fmt.Sprintf("---\ntitle: \"%s\"\n---\n", title))
	diagram.WriteString(fmt.Sprintf("flowchart %s\n", config.Direction))

	// Generate diagram based on orchestration mode
	switch ob.mode {
	case OrchestrationCollaborate:
		ob.generateCollaborativeDiagram(&diagram, config)
	case OrchestrationRoute:
		ob.generateRoutingDiagram(&diagram, config)
	case OrchestrationSequential:
		ob.generateSequentialOrchestrationDiagram(&diagram, config)
	case OrchestrationParallel:
		ob.generateParallelOrchestrationDiagram(&diagram, config)
	case OrchestrationLoop:
		ob.generateLoopOrchestrationDiagram(&diagram, config)
	default:
		ob.generateDefaultOrchestrationDiagram(&diagram, config)
	}

	// Add orchestration metadata
	if config.ShowMetadata {
		ob.addOrchestrationMetadata(&diagram)
	}

	// Add styling
	ob.addOrchestrationStyling(&diagram, config)

	return diagram.String()
}

// generateCollaborativeDiagram creates a collaborative orchestration diagram
func (ob *OrchestrationBuilder) generateCollaborativeDiagram(diagram *strings.Builder, config MermaidConfig) {
	diagram.WriteString("    EVENT[\"📨 Event\"]\n")
	diagram.WriteString("    BROADCAST[\"📡 Broadcast to All\"]\n")
	diagram.WriteString("    COLLECT[\"📊 Collect Results\"]\n")
	diagram.WriteString("    RESULT[\"🎯 Combined Result\"]\n\n")

	diagram.WriteString("    EVENT --> BROADCAST\n")

	// Add agent handlers
	agentCount := 0
	for name := range ob.agents {
		agentCount++
		agentId := fmt.Sprintf("AGENT%d", agentCount)
		diagram.WriteString(fmt.Sprintf("    %s[\"%s Handler\"]\n", agentId, name))
		diagram.WriteString(fmt.Sprintf("    BROADCAST --> %s\n", agentId))
		diagram.WriteString(fmt.Sprintf("    %s --> COLLECT\n", agentId))
	}

	diagram.WriteString("    COLLECT --> RESULT\n")
}

// generateRoutingDiagram creates a routing orchestration diagram
func (ob *OrchestrationBuilder) generateRoutingDiagram(diagram *strings.Builder, config MermaidConfig) {
	diagram.WriteString("    EVENT[\"📨 Event\"]\n")
	diagram.WriteString("    ROUTER{\"🎯 Route Decision\"}\n")
	diagram.WriteString("    RESULT[\"📤 Result\"]\n\n")

	diagram.WriteString("    EVENT --> ROUTER\n")

	// Add agent handlers with routing
	for name := range ob.agents {
		agentId := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
		diagram.WriteString(fmt.Sprintf("    %s[\"%s Handler\"]\n", agentId, name))
		diagram.WriteString(fmt.Sprintf("    ROUTER -->|\"route=%s\"| %s\n", name, agentId))
		diagram.WriteString(fmt.Sprintf("    %s --> RESULT\n", agentId))
	}
}

// generateSequentialOrchestrationDiagram creates a sequential orchestration diagram
func (ob *OrchestrationBuilder) generateSequentialOrchestrationDiagram(diagram *strings.Builder, config MermaidConfig) {
	diagram.WriteString("    EVENT[\"📨 Input Event\"]\n")

	var prevNode = "EVENT"
	agentCount := 0
	for name := range ob.agents {
		agentCount++
		agentId := fmt.Sprintf("AGENT%d", agentCount)
		diagram.WriteString(fmt.Sprintf("    %s[\"%s Handler\"]\n", agentId, name))
		diagram.WriteString(fmt.Sprintf("    %s --> %s\n", prevNode, agentId))
		prevNode = agentId
	}

	diagram.WriteString("    RESULT[\"🎯 Final Result\"]\n")
	diagram.WriteString(fmt.Sprintf("    %s --> RESULT\n", prevNode))
}

// generateParallelOrchestrationDiagram creates a parallel orchestration diagram
func (ob *OrchestrationBuilder) generateParallelOrchestrationDiagram(diagram *strings.Builder, config MermaidConfig) {
	// Similar to collaborative but with different semantics
	ob.generateCollaborativeDiagram(diagram, config)
}

// generateLoopOrchestrationDiagram creates a loop orchestration diagram
func (ob *OrchestrationBuilder) generateLoopOrchestrationDiagram(diagram *strings.Builder, config MermaidConfig) {
	diagram.WriteString("    EVENT[\"📨 Input Event\"]\n")
	diagram.WriteString("    CONDITION{\"🔍 Continue Loop?\"}\n")
	diagram.WriteString("    RESULT[\"🎯 Final Result\"]\n\n")

	// Add single agent for loop
	if len(ob.agents) > 0 {
		// Get first agent for loop
		var agentName string
		for name := range ob.agents {
			agentName = name
			break
		}
		diagram.WriteString(fmt.Sprintf("    AGENT[\"%s Handler\"]\n", agentName))
		diagram.WriteString("    EVENT --> CONDITION\n")
		diagram.WriteString("    CONDITION -->|\"Yes\"| AGENT\n")
		diagram.WriteString("    CONDITION -->|\"No\"| RESULT\n")
		diagram.WriteString("    AGENT --> CONDITION\n")
	}
}

// generateDefaultOrchestrationDiagram creates a default orchestration diagram
func (ob *OrchestrationBuilder) generateDefaultOrchestrationDiagram(diagram *strings.Builder, config MermaidConfig) {
	diagram.WriteString("    EVENT[\"📨 Event\"]\n")
	diagram.WriteString("    UNKNOWN[\"❓ Unknown Orchestration\"]\n")
	diagram.WriteString("    RESULT[\"📤 Result\"]\n\n")

	diagram.WriteString("    EVENT --> UNKNOWN\n")
	diagram.WriteString("    UNKNOWN --> RESULT\n")
}

// =============================================================================
// HELPER METHODS
// =============================================================================

// getTitle returns the diagram title
func (cb *CompositionBuilder) getTitle(config MermaidConfig) string {
	if config.Title != "" {
		return config.Title
	}

	modeStr := strings.Title(cb.mode)
	if modeStr == "" {
		modeStr = "Composition"
	}

	return fmt.Sprintf("%s: %s (%d agents)", modeStr, cb.name, len(cb.agents))
}

// getAgentDisplayName returns a formatted agent display name
func (cb *CompositionBuilder) getAgentDisplayName(agent Agent, index int) string {
	if agent == nil {
		return fmt.Sprintf("🤖 Agent %d", index)
	}

	name := agent.Name()
	if name == "" {
		name = fmt.Sprintf("Agent%d", index)
	}

	return fmt.Sprintf("🤖 %s", name)
}

// addMetadataToMermaid adds metadata information to the diagram
func (cb *CompositionBuilder) addMetadataToMermaid(diagram *strings.Builder) {
	diagram.WriteString("\n    %% Metadata\n")

	if cb.config.Timeout > 0 {
		diagram.WriteString("    %% Timeout: " + cb.config.Timeout.String() + "\n")
	}

	if cb.config.ErrorStrategy != "" {
		diagram.WriteString("    %% Error Strategy: " + string(cb.config.ErrorStrategy) + "\n")
	}

	if cb.config.MaxConcurrency > 0 {
		diagram.WriteString(fmt.Sprintf("    %%%% Max Concurrency: %d\n", cb.config.MaxConcurrency))
	}
}

// addStylingToMermaid adds CSS styling to the diagram
func (cb *CompositionBuilder) addStylingToMermaid(diagram *strings.Builder, config MermaidConfig) {
	diagram.WriteString("\n    %% Styling\n")
	diagram.WriteString("    classDef inputNode fill:#e1f5fe,stroke:#01579b,stroke-width:2px,color:#000\n")
	diagram.WriteString("    classDef agentNode fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,color:#000\n")
	diagram.WriteString("    classDef outputNode fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px,color:#000\n")
	diagram.WriteString("    classDef processNode fill:#fff3e0,stroke:#e65100,stroke-width:2px,color:#000\n")

	diagram.WriteString("    class INPUT,EVENT inputNode\n")
	diagram.WriteString("    class OUTPUT,RESULT outputNode\n")
	diagram.WriteString("    class FANOUT,MERGE,BROADCAST,COLLECT,ROUTER processNode\n")

	// Style agent nodes
	for i := range cb.agents {
		diagram.WriteString(fmt.Sprintf("    class AGENT%d agentNode\n", i+1))
	}
}

// addOrchestrationMetadata adds orchestration metadata to the diagram
func (ob *OrchestrationBuilder) addOrchestrationMetadata(diagram *strings.Builder) {
	diagram.WriteString("\n    %% Orchestration Metadata\n")

	if ob.config.Timeout > 0 {
		diagram.WriteString("    %% Timeout: " + ob.config.Timeout.String() + "\n")
	}

	if ob.config.MaxConcurrency > 0 {
		diagram.WriteString(fmt.Sprintf("    %%%% Max Concurrency: %d\n", ob.config.MaxConcurrency))
	}

	diagram.WriteString(fmt.Sprintf("    %%%% Agents: %d\n", len(ob.agents)))
}

// addOrchestrationStyling adds CSS styling to orchestration diagrams
func (ob *OrchestrationBuilder) addOrchestrationStyling(diagram *strings.Builder, config MermaidConfig) {
	diagram.WriteString("\n    %% Orchestration Styling\n")
	diagram.WriteString("    classDef eventNode fill:#e3f2fd,stroke:#0277bd,stroke-width:2px,color:#000\n")
	diagram.WriteString("    classDef handlerNode fill:#fce4ec,stroke:#c2185b,stroke-width:2px,color:#000\n")
	diagram.WriteString("    classDef resultNode fill:#e8f5e8,stroke:#388e3c,stroke-width:2px,color:#000\n")
	diagram.WriteString("    classDef processNode fill:#fff8e1,stroke:#f57c00,stroke-width:2px,color:#000\n")

	diagram.WriteString("    class EVENT eventNode\n")
	diagram.WriteString("    class RESULT resultNode\n")
	diagram.WriteString("    class BROADCAST,COLLECT,ROUTER,CONDITION processNode\n")

	// Style agent handler nodes
	agentCount := 0
	for range ob.agents {
		agentCount++
		diagram.WriteString(fmt.Sprintf("    class AGENT%d handlerNode\n", agentCount))
	}
}

// =============================================================================
// CONVENIENCE FUNCTIONS FOR DIRECT AGENT VISUALIZATION
// =============================================================================

// GenerateAgentMermaidDiagram generates a simple Mermaid diagram for any agent
func GenerateAgentMermaidDiagram(agent Agent) string {
	return GenerateAgentMermaidDiagramWithConfig(agent, DefaultMermaidConfig())
}

// GenerateAgentMermaidDiagramWithConfig generates a Mermaid diagram for an agent with custom config
func GenerateAgentMermaidDiagramWithConfig(agent Agent, config MermaidConfig) string {
	var diagram strings.Builder

	title := fmt.Sprintf("Agent: %s", agent.Name())
	if config.Title != "" {
		title = config.Title
	}

	diagram.WriteString(fmt.Sprintf("---\ntitle: %s\n---\n", title))
	diagram.WriteString(fmt.Sprintf("flowchart %s\n", config.Direction))

	diagram.WriteString("    INPUT[\"📨 Input State\"]\n")
	diagram.WriteString(fmt.Sprintf("    AGENT[\"🤖 %s\"]\n", agent.Name()))
	diagram.WriteString("    OUTPUT[\"📤 Output State\"]\n\n")

	diagram.WriteString("    INPUT --> AGENT\n")
	diagram.WriteString("    AGENT --> OUTPUT\n")

	// Add basic styling
	diagram.WriteString("\n    %% Styling\n")
	diagram.WriteString("    classDef inputNode fill:#e1f5fe,stroke:#01579b,stroke-width:2px,color:#000\n")
	diagram.WriteString("    classDef agentNode fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,color:#000\n")
	diagram.WriteString("    classDef outputNode fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px,color:#000\n")

	diagram.WriteString("    class INPUT inputNode\n")
	diagram.WriteString("    class AGENT agentNode\n")
	diagram.WriteString("    class OUTPUT outputNode\n")

	return diagram.String()
}

// =============================================================================
// WORKFLOW PATTERN VISUALIZATION
// =============================================================================

// GenerateWorkflowPatternDiagram creates diagrams for common workflow patterns
func GenerateWorkflowPatternDiagram(patternName string, agents []Agent) string {
	config := DefaultMermaidConfig()
	config.Title = fmt.Sprintf("%s Pattern", strings.Title(patternName))

	var diagram strings.Builder
	diagram.WriteString(fmt.Sprintf("---\ntitle: %s\n---\n", config.Title))
	diagram.WriteString(fmt.Sprintf("flowchart %s\n", config.Direction))

	switch strings.ToLower(patternName) {
	case "map-reduce":
		generateMapReduceDiagram(&diagram, agents)
	case "pipeline":
		generatePipelineDiagram(&diagram, agents)
	case "scatter-gather":
		generateScatterGatherDiagram(&diagram, agents)
	case "fan-out-fan-in":
		generateFanOutFanInDiagram(&diagram, agents)
	default:
		generateGenericPatternDiagram(&diagram, agents, patternName)
	}

	return diagram.String()
}

// generateMapReduceDiagram creates a map-reduce pattern diagram
func generateMapReduceDiagram(diagram *strings.Builder, agents []Agent) {
	diagram.WriteString("    INPUT[\"📊 Input Data\"]\n")
	diagram.WriteString("    SPLIT[\"🔄 Split Data\"]\n")
	diagram.WriteString("    REDUCE[\"🔄 Reduce Results\"]\n")
	diagram.WriteString("    OUTPUT[\"📈 Final Result\"]\n\n")

	diagram.WriteString("    INPUT --> SPLIT\n")

	// Map phase
	mapperCount := len(agents) - 1 // Last agent is reducer
	if mapperCount < 1 {
		mapperCount = 1
	}

	for i := 0; i < mapperCount; i++ {
		mapperId := fmt.Sprintf("MAP%d", i+1)
		agentName := "Mapper"
		if i < len(agents) {
			agentName = agents[i].Name()
		}
		diagram.WriteString(fmt.Sprintf("    %s[\"🗺️ %s\"]\n", mapperId, agentName))
		diagram.WriteString(fmt.Sprintf("    SPLIT --> %s\n", mapperId))
		diagram.WriteString(fmt.Sprintf("    %s --> REDUCE\n", mapperId))
	}

	diagram.WriteString("    REDUCE --> OUTPUT\n")
}

// generatePipelineDiagram creates a pipeline pattern diagram
func generatePipelineDiagram(diagram *strings.Builder, agents []Agent) {
	diagram.WriteString("    INPUT[\"📨 Input\"]\n")

	var prevNode = "INPUT"
	for i, agent := range agents {
		stageId := fmt.Sprintf("STAGE%d", i+1)
		diagram.WriteString(fmt.Sprintf("    %s[\"⚙️ %s\"]\n", stageId, agent.Name()))
		diagram.WriteString(fmt.Sprintf("    %s --> %s\n", prevNode, stageId))
		prevNode = stageId
	}

	diagram.WriteString("    OUTPUT[\"📤 Output\"]\n")
	diagram.WriteString(fmt.Sprintf("    %s --> OUTPUT\n", prevNode))
}

// generateScatterGatherDiagram creates a scatter-gather pattern diagram
func generateScatterGatherDiagram(diagram *strings.Builder, agents []Agent) {
	diagram.WriteString("    INPUT[\"📨 Input\"]\n")
	diagram.WriteString("    SCATTER[\"📤 Scatter\"]\n")
	diagram.WriteString("    GATHER[\"📥 Gather\"]\n")
	diagram.WriteString("    OUTPUT[\"📊 Output\"]\n\n")

	diagram.WriteString("    INPUT --> SCATTER\n")

	for i, agent := range agents {
		agentId := fmt.Sprintf("WORKER%d", i+1)
		diagram.WriteString(fmt.Sprintf("    %s[\"👷 %s\"]\n", agentId, agent.Name()))
		diagram.WriteString(fmt.Sprintf("    SCATTER --> %s\n", agentId))
		diagram.WriteString(fmt.Sprintf("    %s --> GATHER\n", agentId))
	}

	diagram.WriteString("    GATHER --> OUTPUT\n")
}

// generateFanOutFanInDiagram creates a fan-out/fan-in pattern diagram
func generateFanOutFanInDiagram(diagram *strings.Builder, agents []Agent) {
	diagram.WriteString("    INPUT[\"📨 Input\"]\n")
	diagram.WriteString("    FANOUT[\"📤 Fan-Out\"]\n")
	diagram.WriteString("    FANIN[\"📥 Fan-In\"]\n")
	diagram.WriteString("    OUTPUT[\"📊 Output\"]\n\n")

	diagram.WriteString("    INPUT --> FANOUT\n")

	for i, agent := range agents {
		processorId := fmt.Sprintf("PROC%d", i+1)
		diagram.WriteString(fmt.Sprintf("    %s[\"⚡ %s\"]\n", processorId, agent.Name()))
		diagram.WriteString(fmt.Sprintf("    FANOUT --> %s\n", processorId))
		diagram.WriteString(fmt.Sprintf("    %s --> FANIN\n", processorId))
	}

	diagram.WriteString("    FANIN --> OUTPUT\n")
}

// generateGenericPatternDiagram creates a generic pattern diagram
func generateGenericPatternDiagram(diagram *strings.Builder, agents []Agent, patternName string) {
	diagram.WriteString("    INPUT[\"📨 Input\"]\n")
	diagram.WriteString(fmt.Sprintf("    PATTERN[\"%s Pattern\"]\n", patternName))
	diagram.WriteString("    OUTPUT[\"📤 Output\"]\n\n")

	diagram.WriteString("    INPUT --> PATTERN\n")

	for i, agent := range agents {
		agentId := fmt.Sprintf("AGENT%d", i+1)
		diagram.WriteString(fmt.Sprintf("    %s[\"🤖 %s\"]\n", agentId, agent.Name()))
		diagram.WriteString(fmt.Sprintf("    PATTERN --> %s\n", agentId))
		diagram.WriteString(fmt.Sprintf("    %s --> PATTERN\n", agentId))
	}

	diagram.WriteString("    PATTERN --> OUTPUT\n")
}

// =============================================================================
// MARKDOWN FILE EXPORT UTILITIES
// =============================================================================

// SaveDiagramAsMarkdown saves a Mermaid diagram as a Markdown file with proper formatting
func SaveDiagramAsMarkdown(filename, title, diagram string) error {
	content := fmt.Sprintf(`# %s

%s

## How to View

1. Open this file in any Markdown viewer (VS Code, GitHub, GitLab, etc.)
2. Copy the Mermaid code below to [Mermaid Live Editor](https://mermaid.live)
3. Export as PNG/SVG for presentations

## Mermaid Diagram

`+"```mermaid\n%s\n```", title, getGeneratedTimestamp(), diagram)

	return os.WriteFile(filename, []byte(content), 0644)
}

// SaveDiagramWithMetadata saves a diagram with additional metadata as Markdown
func SaveDiagramWithMetadata(filename, title, description, diagram string, metadata map[string]interface{}) error {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s\n\n", title))

	if description != "" {
		content.WriteString(fmt.Sprintf("%s\n\n", description))
	}

	content.WriteString(getGeneratedTimestamp() + "\n\n")

	// Add metadata if provided
	if len(metadata) > 0 {
		content.WriteString("## Configuration\n\n")
		for key, value := range metadata {
			content.WriteString(fmt.Sprintf("- **%s**: %v\n", key, value))
		}
		content.WriteString("\n")
	}

	content.WriteString("## Workflow Diagram\n\n")
	content.WriteString("```mermaid\n")
	content.WriteString(diagram)
	content.WriteString("\n```\n\n")

	content.WriteString("## Usage\n\n")
	content.WriteString("1. View this diagram in any Markdown-compatible viewer\n")
	content.WriteString("2. Copy the Mermaid code to [Mermaid Live Editor](https://mermaid.live) for interactive viewing\n")
	content.WriteString("3. Export as PNG/SVG for documentation and presentations\n")

	return os.WriteFile(filename, []byte(content.String()), 0644)
}

// getGeneratedTimestamp returns a formatted timestamp for documentation
func getGeneratedTimestamp() string {
	return fmt.Sprintf("*Generated on %s*", time.Now().Format("January 2, 2006 at 3:04 PM"))
}

// ConvertMmdToMarkdown converts existing .mmd files to .md format
func ConvertMmdToMarkdown(mmdFile, title string) error {
	// Read the existing .mmd file
	content, err := os.ReadFile(mmdFile)
	if err != nil {
		return fmt.Errorf("failed to read .mmd file: %w", err)
	}

	// Generate markdown filename
	mdFile := strings.TrimSuffix(mmdFile, ".mmd") + ".md"

	// Save as markdown
	return SaveDiagramAsMarkdown(mdFile, title, string(content))
}
