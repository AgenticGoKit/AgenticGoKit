package core

import "fmt"

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

// MermaidGenerator interface for generating Mermaid diagrams
type MermaidGenerator interface {
	GenerateCompositionDiagram(mode, name string, agents []Agent, config MermaidConfig) string
	GenerateOrchestrationDiagram(mode OrchestrationMode, agents map[string]AgentHandler, config MermaidConfig) string
	GenerateAgentDiagram(agent Agent, config MermaidConfig) string
	GenerateWorkflowPatternDiagram(patternName string, agents []Agent, config MermaidConfig) string
	SaveDiagramAsMarkdown(filename, title, diagram string) error
	SaveDiagramWithMetadata(filename, title, description, diagram string, metadata map[string]interface{}) error
}

// MermaidGeneratorFactory is the function signature for creating MermaidGenerator instances
type MermaidGeneratorFactory func() MermaidGenerator

// mermaidGeneratorFactory holds the registered factory function
var mermaidGeneratorFactory MermaidGeneratorFactory

// RegisterMermaidGeneratorFactory registers the MermaidGenerator factory function
func RegisterMermaidGeneratorFactory(factory MermaidGeneratorFactory) {
	mermaidGeneratorFactory = factory
}

// NewMermaidGenerator creates a new MermaidGenerator instance
func NewMermaidGenerator() MermaidGenerator {
	if mermaidGeneratorFactory != nil {
		return mermaidGeneratorFactory()
	}

	// Fallback to simple implementation if internal package not imported
	return &simpleMermaidGenerator{}
}

// Simple implementation for core package
type simpleMermaidGenerator struct{}

func (mg *simpleMermaidGenerator) GenerateCompositionDiagram(mode, name string, agents []Agent, config MermaidConfig) string {
	// Simple basic implementation for now
	return fmt.Sprintf(`---
title: "%s Composition"
---
flowchart TD
    INPUT["Input"] --> OUTPUT["Output"]
    `, mode)
}

func (mg *simpleMermaidGenerator) GenerateOrchestrationDiagram(mode OrchestrationMode, agents map[string]AgentHandler, config MermaidConfig) string {
	return `---
title: "Orchestration"
---
flowchart TD
    INPUT["Input"] --> OUTPUT["Output"]`
}

func (mg *simpleMermaidGenerator) GenerateAgentDiagram(agent Agent, config MermaidConfig) string {
	return fmt.Sprintf(`---
title: "Agent: %s"
---
flowchart TD
    INPUT["Input"] --> AGENT["%s"] --> OUTPUT["Output"]`, agent.Name(), agent.Name())
}

func (mg *simpleMermaidGenerator) GenerateWorkflowPatternDiagram(patternName string, agents []Agent, config MermaidConfig) string {
	return fmt.Sprintf(`---
title: "%s Pattern"
---
flowchart TD
    INPUT["Input"] --> OUTPUT["Output"]`, patternName)
}

func (mg *simpleMermaidGenerator) SaveDiagramAsMarkdown(filename, title, diagram string) error {
	return nil // TODO: Implement
}

func (mg *simpleMermaidGenerator) SaveDiagramWithMetadata(filename, title, description, diagram string, metadata map[string]interface{}) error {
	return nil // TODO: Implement
}
