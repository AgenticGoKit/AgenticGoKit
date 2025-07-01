# AgentFlow Workflow Visualization Guide

## Overview

AgentFlow provides comprehensive workflow visualization capabilities using Mermaid diagrams. You can visualize multi-agent compositions, orchestrations, and complex workflow patterns to better understand and document your agent systems.

## Features

### 1. Composition Builder Visualization
- **Sequential workflows**: Step-by-step agent processing pipelines
- **Parallel processing**: Fan-out/fan-in patterns for concurrent execution  
- **Loop patterns**: Retry and iteration logic with conditions

### 2. Orchestration Builder Visualization
- **Collaborative orchestration**: Broadcast events to multiple agents
- **Routing patterns**: Conditional routing based on event types
- **Mixed orchestration modes**: Sequential, parallel, and loop orchestrations

### 3. AgentBuilder Integration
- **Multi-agent composition visualization**: Preview compositions before building
- **Pre-build diagram generation**: Generate diagrams during agent construction
- **Configuration validation**: Visualize settings like timeouts and error strategies

### 4. Advanced Features
- **Custom diagram configuration**: Control direction, themes, and metadata display
- **Pre-built workflow patterns**: Map-reduce, pipeline, and other common patterns
- **File export**: Save diagrams as .mmd files for documentation
- **Styling and metadata**: Professional diagrams with execution details

## Quick Start Examples

### Basic Sequential Workflow

```go
// Create agents
orderAgent := createAgent("OrderProcessor")
paymentAgent := createAgent("PaymentProcessor") 
shippingAgent := createAgent("ShippingService")

// Build sequential composition
pipeline := core.NewComposition("order-pipeline").
    WithAgents(orderAgent, paymentAgent, shippingAgent).
    AsSequential().
    WithTimeout(2 * time.Minute)

// Generate diagram
diagram := pipeline.GenerateMermaidDiagram()
fmt.Println(diagram)
```

### Parallel Processing

```go
// Create analysis agents
sentimentAgent := createAgent("SentimentAnalyzer")
keywordAgent := createAgent("KeywordExtractor")
languageAgent := createAgent("LanguageDetector")

// Build parallel composition
analysis := core.NewComposition("content-analysis").
    WithAgents(sentimentAgent, keywordAgent, languageAgent).
    AsParallel().
    WithTimeout(30 * time.Second).
    WithErrorStrategy(core.ErrorStrategyCollectAll)

// Generate custom diagram (left-to-right)
config := core.MermaidConfig{
    DiagramType:    core.MermaidFlowchart,
    Title:          "Content Analysis System",
    Direction:      "LR",
    ShowMetadata:   true,
    ShowAgentTypes: true,
}

diagram := analysis.GenerateMermaidDiagramWithConfig(config)
```

### AgentBuilder with Visualization

```go
// Build agent with multi-agent composition
builder := core.NewAgent("DataProcessor").
    WithParallelAgents(dataAgent, analyticsAgent, reportAgent).
    WithMultiAgentConfig(core.MultiAgentConfig{
        Timeout:        90 * time.Second,
        MaxConcurrency: 8,
        ErrorStrategy:  core.ErrorStrategyCollectAll,
        StateStrategy:  core.StateStrategyMerge,
    })

// Check if it can be visualized
if builder.CanVisualize() {
    diagram := builder.GenerateMermaidDiagram()
    // Save to file
    os.WriteFile("workflow.mmd", []byte(diagram), 0644)
}

// Build the actual agent
agent, err := builder.Build()
```

### Loop with Conditions

```go
qualityAgent := createAgent("QualityChecker")

// Define stop condition
condition := func(state core.State) bool {
    if score, exists := state.Get("quality_score"); exists {
        if qualityScore, ok := score.(float64); ok {
            return qualityScore >= 0.95 // Stop at 95% quality
        }
    }
    return false
}

// Build loop composition
qualityLoop := core.NewComposition("quality-monitor").
    WithAgents(qualityAgent).
    AsLoop(10, condition). // Max 10 iterations
    WithTimeout(5 * time.Minute)

diagram := qualityLoop.GenerateMermaidDiagram()
```

### Orchestration Patterns

```go
// Collaborative microservices
serviceHandlers := map[string]core.AgentHandler{
    "user-service":    core.ConvertAgentToHandler(userAgent),
    "order-service":   core.ConvertAgentToHandler(orderAgent),
    "payment-service": core.ConvertAgentToHandler(paymentAgent),
}

collaboration := core.NewOrchestrationBuilder(core.OrchestrationCollaborate).
    WithAgents(serviceHandlers).
    WithTimeout(1 * time.Minute).
    WithMaxConcurrency(20)

diagram := collaboration.GenerateMermaidDiagram()

// API routing
routing := core.NewOrchestrationBuilder(core.OrchestrationRoute).
    WithAgents(serviceHandlers).
    WithTimeout(30 * time.Second)

routingDiagram := routing.GenerateMermaidDiagram()
```

## Configuration Options

### MermaidConfig Structure

```go
type MermaidConfig struct {
    DiagramType    MermaidDiagramType // flowchart, sequenceDiagram, etc.
    Title          string             // Custom diagram title  
    Direction      string             // "TD", "LR", "BT", "RL"
    Theme          string             // "default", "dark", "forest"
    ShowMetadata   bool               // Include timeout/error info
    ShowAgentTypes bool               // Show agent type details
    CompactMode    bool               // Generate compact diagrams
}
```

### Available Directions
- `"TD"` or `"TB"`: Top to Bottom (default)
- `"LR"`: Left to Right
- `"BT"`: Bottom to Top  
- `"RL"`: Right to Left

### Themes
- `"default"`: Standard Mermaid theme
- `"dark"`: Dark theme for presentations
- `"forest"`: Green theme
- `"base"`: Minimal theme

## Workflow Patterns

### Pre-built Patterns

```go
// Map-Reduce pattern
agents := []core.Agent{dataAgent, processor1, processor2, reducer}
mapReduceDiagram := core.GenerateWorkflowPatternDiagram("map-reduce", agents)

// Pipeline pattern  
pipelineAgents := []core.Agent{input, transform, validate, output}
pipelineDiagram := core.GenerateWorkflowPatternDiagram("pipeline", pipelineAgents)
```

## File Export and Integration

### Save Diagrams to Files

```go
// Create output directory
outputDir := "workflow_diagrams"
os.MkdirAll(outputDir, 0755)

// Save diagram as Markdown file
filename := filepath.Join(outputDir, "my_workflow.md")
err := core.SaveDiagramAsMarkdown(filename, "My Workflow", diagram)

// Save with metadata
metadata := map[string]interface{}{
    "Pattern":      "Sequential Processing", 
    "Agents":       4,
    "Timeout":      "2 minutes",
    "Error Strategy": "Fail Fast",
}
err = core.SaveDiagramWithMetadata(filename, "My Workflow", 
    "Description of the workflow", diagram, metadata)
```

### View Diagrams

1. **VS Code/GitHub/GitLab**: Open `.md` files directly - Mermaid diagrams render automatically
2. **Online**: Copy Mermaid code to [Mermaid Live Editor](https://mermaid.live)  
3. **Documentation**: Include `.md` files in project documentation
4. **Presentations**: Export from Mermaid Live as PNG/SVG

### Integration in Documentation

```markdown
# My Workflow

Here's how our order processing works:

\`\`\`mermaid
---
title: Order Processing Pipeline
---
flowchart TD
    INPUT["🎯 Order Request"]
    AGENT1["🤖 InventoryChecker"]
    INPUT --> AGENT1
    AGENT2["🤖 PaymentProcessor"] 
    AGENT1 --> AGENT2
    AGENT3["🤖 ShippingService"]
    AGENT2 --> AGENT3
    OUTPUT["✅ Order Complete"]
    AGENT3 --> OUTPUT
\`\`\`
```

## Best Practices

### 1. Use Descriptive Agent Names
- ✅ `"OrderProcessor"`, `"PaymentGateway"`
- ❌ `"Agent1"`, `"Worker"`

### 2. Choose Appropriate Directions
- **Sequential workflows**: Top-to-bottom (`"TD"`)
- **Data flows**: Left-to-right (`"LR"`)
- **Process hierarchies**: Top-to-bottom (`"TD"`)

### 3. Include Metadata for Complex Systems
```go
config := core.MermaidConfig{
    ShowMetadata:   true,  // Show timeouts, error strategies
    ShowAgentTypes: true,  // Show agent capabilities
}
```

### Save Diagrams During Development
```go
// Always save diagrams for documentation
if builder.CanVisualize() {
    diagram := builder.GenerateMermaidDiagram()
    filename := fmt.Sprintf("docs/%s.md", builder.Name())
    core.SaveDiagramAsMarkdown(filename, builder.Name(), diagram)
}
```

### 5. Use Custom Titles for Clarity
```go
config := core.MermaidConfig{
    Title: "E-commerce Order Processing Pipeline v2.1",
}
```

## Troubleshooting

### Common Issues

1. **Empty Diagram**: Check if composition has agents and a mode set
   ```go
   if builder.CanVisualize() {
       // Safe to generate diagram
   }
   ```

2. **Missing Styling**: Ensure metadata is enabled
   ```go
   config.ShowMetadata = true
   config.ShowAgentTypes = true
   ```

3. **Complex Diagrams**: Use compact mode for large compositions
   ```go
   config.CompactMode = true
   ```

### Validation

```go
// Validate before visualization
if err := builder.Validate(); err != nil {
    fmt.Printf("Builder error: %v\n", err)
    return
}

// Check visualization capability
if !builder.CanVisualize() {
    fmt.Println("No multi-agent composition to visualize")
    return
}
```

## API Reference

### Composition Builder
- `GenerateMermaidDiagram() string`
- `GenerateMermaidDiagramWithConfig(config MermaidConfig) string`

### Orchestration Builder  
- `GenerateMermaidDiagram() string`
- `GenerateMermaidDiagramWithConfig(config MermaidConfig) string`

### Agent Builder
- `CanVisualize() bool`
- `GenerateMermaidDiagram() string`
- `GenerateMermaidDiagramWithConfig(config MermaidConfig) string`

### Workflow Patterns
- `GenerateWorkflowPatternDiagram(pattern string, agents []Agent) string`

### File Export Utilities
- `SaveDiagramAsMarkdown(filename, title, diagram string) error`
- `SaveDiagramWithMetadata(filename, title, description, diagram string, metadata map[string]interface{}) error`
- `ConvertMmdToMarkdown(mmdFile, title string) error`

Supported patterns: `"map-reduce"`, `"pipeline"`, `"scatter-gather"`

## Examples Repository

See the `examples/visualization/` directory for complete working examples:
- `demo.go`: Basic visualization examples
- `comprehensive_demo.go`: Advanced patterns and configurations
- Generated `.mmd` files for reference

## Next Steps

1. **Explore Examples**: Run the demo programs to see all features
2. **Create Custom Patterns**: Build your own workflow visualizations  
3. **Integrate Documentation**: Add diagrams to your project documentation
4. **Share Diagrams**: Export and share workflow visualizations with your team
