// Package validation provides internal validation functionality for AgenticGoKit.
package validation

import (
	"fmt"
	"strings"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// init registers the WorkflowValidator implementation with the core package
func init() {
	core.RegisterWorkflowValidatorFactory(func() core.WorkflowValidator {
		return &WorkflowValidatorImpl{}
	})
}

// WorkflowValidatorImpl implements the core.WorkflowValidator interface
type WorkflowValidatorImpl struct{}

// ValidateComposition validates a composition of agents
func (wv *WorkflowValidatorImpl) ValidateComposition(agents []core.Agent, mode string) core.WorkflowValidationResult {
	result := core.WorkflowValidationResult{
		IsValid:  true,
		Errors:   []core.WorkflowValidationError{},
		Warnings: []core.WorkflowValidationWarning{},
	}

	// Check for empty agent list
	if len(agents) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, core.WorkflowValidationError{
			Type:      string(core.ValidationErrorNoEntryPoint),
			Message:   "No agents defined in composition",
			Severity:  "error",
			Component: "composition",
			Suggestions: []string{
				"Add at least one agent to the composition",
				"Verify that agent configuration is correct",
			},
		})
		return result
	}

	// Check for duplicate agent names
	seenNames := make(map[string]bool)
	for _, agent := range agents {
		name := agent.Name()
		if seenNames[name] {
			result.IsValid = false
			result.Errors = append(result.Errors, core.WorkflowValidationError{
				Type:      string(core.ValidationErrorInvalidRouting),
				Message:   fmt.Sprintf("Duplicate agent name: %s", name),
				Severity:  "error",
				Component: name,
				Suggestions: []string{
					"Use unique names for all agents",
					"Check agent configuration for duplicates",
				},
			})
		}
		seenNames[name] = true
	}

	// Check for invalid agent names
	for _, agent := range agents {
		name := agent.Name()
		if name == "" {
			result.IsValid = false
			result.Errors = append(result.Errors, core.WorkflowValidationError{
				Type:      string(core.ValidationErrorMissingAgent),
				Message:   "Agent has empty name",
				Severity:  "error",
				Component: "unnamed_agent",
				Suggestions: []string{
					"Provide a valid name for each agent",
					"Check agent initialization code",
				},
			})
		} else if strings.Contains(name, " ") || strings.Contains(name, "\t") {
			result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
				Type:      "NAMING_CONVENTION",
				Message:   fmt.Sprintf("Agent name '%s' contains whitespace", name),
				Component: name,
			})
		}
	}

	// Mode-specific validation
	switch strings.ToLower(mode) {
	case "collaborative", "parallel":
		wv.validateCollaborativeMode(agents, &result)
	case "sequential", "pipeline":
		wv.validateSequentialMode(agents, &result)
	case "route", "routing":
		wv.validateRoutingMode(agents, &result)
	default:
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "UNKNOWN_MODE",
			Message:   fmt.Sprintf("Unknown composition mode: %s", mode),
			Component: "composition",
		})
	}

	return result
}

// ValidateOrchestration validates an orchestration configuration
func (wv *WorkflowValidatorImpl) ValidateOrchestration(agents map[string]core.AgentHandler, mode core.OrchestrationMode) core.WorkflowValidationResult {
	result := core.WorkflowValidationResult{
		IsValid:  true,
		Errors:   []core.WorkflowValidationError{},
		Warnings: []core.WorkflowValidationWarning{},
	}

	// Check for empty agent map
	if len(agents) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, core.WorkflowValidationError{
			Type:      string(core.ValidationErrorNoEntryPoint),
			Message:   "No agents registered in orchestration",
			Severity:  "error",
			Component: "orchestration",
			Suggestions: []string{
				"Register at least one agent",
				"Check agent registration logic",
			},
		})
		return result
	}

	// Validate agent names and handlers
	for name, handler := range agents {
		if name == "" {
			result.IsValid = false
			result.Errors = append(result.Errors, core.WorkflowValidationError{
				Type:      string(core.ValidationErrorMissingAgent),
				Message:   "Agent registered with empty name",
				Severity:  "error",
				Component: "orchestration",
				Suggestions: []string{
					"Use non-empty names for agent registration",
				},
			})
		}

		if handler == nil {
			result.IsValid = false
			result.Errors = append(result.Errors, core.WorkflowValidationError{
				Type:      string(core.ValidationErrorMissingAgent),
				Message:   fmt.Sprintf("Agent '%s' has nil handler", name),
				Severity:  "error",
				Component: name,
				Suggestions: []string{
					"Provide a valid handler for each agent",
					"Check agent initialization",
				},
			})
		}
	}

	// Mode-specific orchestration validation
	switch mode {
	case core.OrchestrationRoute:
		wv.validateRouteOrchestration(agents, &result)
	case core.OrchestrationCollaborate:
		wv.validateCollaborativeOrchestration(agents, &result)
	case core.OrchestrationSequential:
		wv.validateSequentialOrchestration(agents, &result)
	case core.OrchestrationLoop:
		wv.validateLoopOrchestration(agents, &result)
	case core.OrchestrationMixed:
		wv.validateMixedOrchestration(agents, &result)
	default:
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "UNKNOWN_ORCHESTRATION_MODE",
			Message:   fmt.Sprintf("Unknown orchestration mode: %s", mode),
			Component: "orchestration",
		})
	}

	return result
}

// ValidateWorkflowGraph validates a workflow graph structure
func (wv *WorkflowValidatorImpl) ValidateWorkflowGraph(graph core.WorkflowGraph) core.WorkflowValidationResult {
	result := core.WorkflowValidationResult{
		IsValid:  true,
		Errors:   []core.WorkflowValidationError{},
		Warnings: []core.WorkflowValidationWarning{},
	}

	// Check for empty graph
	if len(graph.Nodes) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, core.WorkflowValidationError{
			Type:      string(core.ValidationErrorNoEntryPoint),
			Message:   "Workflow graph has no nodes",
			Severity:  "error",
			Component: "graph",
			Suggestions: []string{
				"Add nodes to the workflow graph",
			},
		})
		return result
	}

	// Validate nodes
	entryPoints := 0
	endpoints := 0
	for name, node := range graph.Nodes {
		if node.Name != name {
			result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
				Type:      "INCONSISTENT_NAMING",
				Message:   fmt.Sprintf("Node key '%s' doesn't match node name '%s'", name, node.Name),
				Component: name,
			})
		}

		if node.IsEntryPoint {
			entryPoints++
		}
		if node.IsEndpoint {
			endpoints++
		}
	}

	// Check for entry points
	if entryPoints == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, core.WorkflowValidationError{
			Type:      string(core.ValidationErrorNoEntryPoint),
			Message:   "No entry point defined in workflow graph",
			Severity:  "error",
			Component: "graph",
			Suggestions: []string{
				"Mark at least one node as an entry point",
			},
		})
	}

	// Check for endpoints
	if endpoints == 0 {
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "NO_ENDPOINT",
			Message:   "No explicit endpoint defined in workflow graph",
			Component: "graph",
		})
	}

	// Validate edges
	wv.validateGraphEdges(graph, &result)

	// Check for circular dependencies
	wv.detectCircularDependencies(graph, &result)

	// Check for orphaned nodes
	wv.detectOrphanedNodes(graph, &result)

	return result
}

// Helper methods for mode-specific validation

func (wv *WorkflowValidatorImpl) validateCollaborativeMode(agents []core.Agent, result *core.WorkflowValidationResult) {
	// In collaborative mode, all agents should be able to process the same input
	if len(agents) > 10 {
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "PERFORMANCE_WARNING",
			Message:   fmt.Sprintf("Large number of agents (%d) in collaborative mode may impact performance", len(agents)),
			Component: "composition",
		})
	}
}

func (wv *WorkflowValidatorImpl) validateSequentialMode(agents []core.Agent, result *core.WorkflowValidationResult) {
	// In sequential mode, agent order matters
	if len(agents) > 20 {
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "PERFORMANCE_WARNING",
			Message:   fmt.Sprintf("Long sequential chain (%d agents) may have high latency", len(agents)),
			Component: "composition",
		})
	}
}

func (wv *WorkflowValidatorImpl) validateRoutingMode(agents []core.Agent, result *core.WorkflowValidationResult) {
	// In routing mode, we need to ensure routing logic exists
	if len(agents) == 1 {
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "DESIGN_WARNING",
			Message:   "Only one agent in routing mode - consider direct execution",
			Component: "composition",
		})
	}
}

func (wv *WorkflowValidatorImpl) validateRouteOrchestration(agents map[string]core.AgentHandler, result *core.WorkflowValidationResult) {
	// Check for common routing agent names
	hasDefaultAgent := false
	hasErrorHandler := false

	for name := range agents {
		if name == "default" || name == "fallback" {
			hasDefaultAgent = true
		}
		if name == "error-handler" || name == "error" {
			hasErrorHandler = true
		}
	}

	if !hasDefaultAgent {
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "ROUTING_WARNING",
			Message:   "No default/fallback agent found for route orchestration",
			Component: "orchestration",
		})
	}

	if !hasErrorHandler {
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "ERROR_HANDLING_WARNING",
			Message:   "No error handler agent found",
			Component: "orchestration",
		})
	}
}

func (wv *WorkflowValidatorImpl) validateCollaborativeOrchestration(agents map[string]core.AgentHandler, result *core.WorkflowValidationResult) {
	if len(agents) > 15 {
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "PERFORMANCE_WARNING",
			Message:   fmt.Sprintf("Large number of collaborative agents (%d) may impact performance", len(agents)),
			Component: "orchestration",
		})
	}
}

func (wv *WorkflowValidatorImpl) validateSequentialOrchestration(agents map[string]core.AgentHandler, result *core.WorkflowValidationResult) {
	if len(agents) > 25 {
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "PERFORMANCE_WARNING",
			Message:   fmt.Sprintf("Long sequential orchestration (%d agents) may have high latency", len(agents)),
			Component: "orchestration",
		})
	}
}

func (wv *WorkflowValidatorImpl) validateLoopOrchestration(agents map[string]core.AgentHandler, result *core.WorkflowValidationResult) {
	if len(agents) != 1 {
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "DESIGN_WARNING",
			Message:   fmt.Sprintf("Loop orchestration typically uses one agent, but %d are registered", len(agents)),
			Component: "orchestration",
		})
	}
}

func (wv *WorkflowValidatorImpl) validateMixedOrchestration(agents map[string]core.AgentHandler, result *core.WorkflowValidationResult) {
	if len(agents) < 2 {
		result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
			Type:      "DESIGN_WARNING",
			Message:   "Mixed orchestration typically requires multiple agents for effective use",
			Component: "orchestration",
		})
	}
}

// Graph validation helpers

func (wv *WorkflowValidatorImpl) validateGraphEdges(graph core.WorkflowGraph, result *core.WorkflowValidationResult) {
	for _, edge := range graph.Edges {
		// Check if referenced nodes exist
		if _, exists := graph.Nodes[edge.From]; !exists {
			result.IsValid = false
			result.Errors = append(result.Errors, core.WorkflowValidationError{
				Type:      string(core.ValidationErrorMissingAgent),
				Message:   fmt.Sprintf("Edge references non-existent 'from' node: %s", edge.From),
				Severity:  "error",
				Component: "graph",
				Details: map[string]interface{}{
					"edge_from": edge.From,
					"edge_to":   edge.To,
				},
			})
		}

		if _, exists := graph.Nodes[edge.To]; !exists {
			result.IsValid = false
			result.Errors = append(result.Errors, core.WorkflowValidationError{
				Type:      string(core.ValidationErrorMissingAgent),
				Message:   fmt.Sprintf("Edge references non-existent 'to' node: %s", edge.To),
				Severity:  "error",
				Component: "graph",
				Details: map[string]interface{}{
					"edge_from": edge.From,
					"edge_to":   edge.To,
				},
			})
		}
	}
}

func (wv *WorkflowValidatorImpl) detectCircularDependencies(graph core.WorkflowGraph, result *core.WorkflowValidationResult) {
	// Simple cycle detection using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		// Check all outgoing edges
		for _, edge := range graph.Edges {
			if edge.From == node {
				if !visited[edge.To] {
					if dfs(edge.To) {
						return true
					}
				} else if recStack[edge.To] {
					result.IsValid = false
					result.Errors = append(result.Errors, core.WorkflowValidationError{
						Type:      string(core.ValidationErrorCircularDependency),
						Message:   fmt.Sprintf("Circular dependency detected involving nodes: %s -> %s", node, edge.To),
						Severity:  "error",
						Component: "graph",
						Details: map[string]interface{}{
							"cycle_nodes": []string{node, edge.To},
						},
						Suggestions: []string{
							"Remove or restructure edges to eliminate cycles",
							"Consider using conditional logic instead of cycles",
						},
					})
					return true
				}
			}
		}

		recStack[node] = false
		return false
	}

	for nodeName := range graph.Nodes {
		if !visited[nodeName] {
			dfs(nodeName)
		}
	}
}

func (wv *WorkflowValidatorImpl) detectOrphanedNodes(graph core.WorkflowGraph, result *core.WorkflowValidationResult) {
	reachableNodes := make(map[string]bool)

	// Find all nodes reachable from entry points
	entryPoints := []string{}
	for name, node := range graph.Nodes {
		if node.IsEntryPoint {
			entryPoints = append(entryPoints, name)
		}
	}

	var markReachable func(node string)
	markReachable = func(node string) {
		if reachableNodes[node] {
			return
		}
		reachableNodes[node] = true

		// Mark all nodes reachable from this node
		for _, edge := range graph.Edges {
			if edge.From == node {
				markReachable(edge.To)
			}
		}
	}

	for _, entryPoint := range entryPoints {
		markReachable(entryPoint)
	}

	// Check for orphaned nodes
	for nodeName := range graph.Nodes {
		if !reachableNodes[nodeName] {
			result.Warnings = append(result.Warnings, core.WorkflowValidationWarning{
				Type:      string(core.ValidationErrorOrphanedAgent),
				Message:   fmt.Sprintf("Node '%s' is not reachable from any entry point", nodeName),
				Component: nodeName,
			})
		}
	}
}
