// Package validation provides internal workflow validation implementation.
package validation

import (
	"fmt"
	"strings"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// WorkflowValidatorImplementation provides the concrete implementation of workflow validation
type WorkflowValidatorImplementation struct{}

// NewWorkflowValidatorImplementation creates a new workflow validator implementation
func NewWorkflowValidatorImplementation() core.WorkflowValidator {
	return &WorkflowValidatorImplementation{}
}

// ValidateComposition validates a composition workflow
func (wv *WorkflowValidatorImplementation) ValidateComposition(agents []core.Agent, mode string) core.WorkflowValidationResult {
	var errors []core.WorkflowValidationError
	var warnings []core.WorkflowValidationWarning

	// Basic validation
	if len(agents) == 0 {
		errors = append(errors, core.WorkflowValidationError{
			Type:        "EmptyComposition",
			Message:     "Composition must contain at least one agent",
			Severity:    "Critical",
			Component:   "Composition",
			Suggestions: []string{"Add at least one agent to the composition"},
		})
		return core.WorkflowValidationResult{
			IsValid:  false,
			Errors:   errors,
			Warnings: warnings,
		}
	}

	// Mode-specific validation
	switch mode {
	case "parallel":
		errors = append(errors, wv.validateParallelComposition(agents)...)
		warnings = append(warnings, wv.warnParallelComposition(agents)...)
	case "sequential":
		errors = append(errors, wv.validateSequentialComposition(agents)...)
		warnings = append(warnings, wv.warnSequentialComposition(agents)...)
	case "loop":
		errors = append(errors, wv.validateLoopComposition(agents)...)
		warnings = append(warnings, wv.warnLoopComposition(agents)...)
	default:
		warnings = append(warnings, core.WorkflowValidationWarning{
			Type:      "UnknownMode",
			Message:   fmt.Sprintf("Unknown composition mode '%s', assuming sequential", mode),
			Component: "Composition",
		})
	}

	// Agent-specific validation
	for i, agent := range agents {
		agentErrors := wv.validateAgent(agent, fmt.Sprintf("Agent[%d]", i))
		errors = append(errors, agentErrors...)
	}

	return core.WorkflowValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// ValidateOrchestration validates an orchestration workflow
func (wv *WorkflowValidatorImplementation) ValidateOrchestration(agents map[string]core.AgentHandler, mode core.OrchestrationMode) core.WorkflowValidationResult {
	var errors []core.WorkflowValidationError
	var warnings []core.WorkflowValidationWarning

	// Basic validation
	if len(agents) == 0 {
		errors = append(errors, core.WorkflowValidationError{
			Type:        "EmptyOrchestration",
			Message:     "Orchestration must contain at least one agent handler",
			Severity:    "Critical",
			Component:   "Orchestration",
			Suggestions: []string{"Add at least one agent handler to the orchestration"},
		})
		return core.WorkflowValidationResult{
			IsValid:  false,
			Errors:   errors,
			Warnings: warnings,
		}
	}

	// Mode-specific validation
	switch mode {
	case core.OrchestrationCollaborate:
		errors = append(errors, wv.validateCollaborativeOrchestration(agents)...)
		warnings = append(warnings, wv.warnCollaborativeOrchestration(agents)...)
	case core.OrchestrationRoute:
		errors = append(errors, wv.validateRoutingOrchestration(agents)...)
		warnings = append(warnings, wv.warnRoutingOrchestration(agents)...)
	case core.OrchestrationSequential:
		errors = append(errors, wv.validateSequentialOrchestration(agents)...)
		warnings = append(warnings, wv.warnSequentialOrchestration(agents)...)
	case core.OrchestrationParallel:
		errors = append(errors, wv.validateParallelOrchestration(agents)...)
		warnings = append(warnings, wv.warnParallelOrchestration(agents)...)
	case core.OrchestrationLoop:
		errors = append(errors, wv.validateLoopOrchestration(agents)...)
		warnings = append(warnings, wv.warnLoopOrchestration(agents)...)
	default:
		errors = append(errors, core.WorkflowValidationError{
			Type:        "InvalidOrchestrationMode",
			Message:     fmt.Sprintf("Invalid orchestration mode: %s", mode),
			Severity:    "Critical",
			Component:   "Orchestration",
			Suggestions: []string{"Use a valid orchestration mode (collaborate, route, sequential, parallel, loop)"},
		})
	}

	// Agent handler validation
	for name, handler := range agents {
		handlerErrors := wv.validateAgentHandler(handler, name)
		errors = append(errors, handlerErrors...)
	}

	return core.WorkflowValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// ValidateWorkflowGraph validates a complex workflow graph
func (wv *WorkflowValidatorImplementation) ValidateWorkflowGraph(graph core.WorkflowGraph) core.WorkflowValidationResult {
	var errors []core.WorkflowValidationError
	var warnings []core.WorkflowValidationWarning

	// Check for cycles in the graph
	if hasCycles := wv.detectCycles(graph); hasCycles {
		errors = append(errors, core.WorkflowValidationError{
			Type:        "CyclicDependency",
			Message:     "Workflow graph contains cyclic dependencies",
			Severity:    "Critical",
			Component:   "WorkflowGraph",
			Suggestions: []string{"Remove circular dependencies between workflow steps"},
		})
	}

	// Check for unreachable nodes
	unreachableNodes := wv.findUnreachableNodes(graph)
	if len(unreachableNodes) > 0 {
		warnings = append(warnings, core.WorkflowValidationWarning{
			Type:      "UnreachableNodes",
			Message:   fmt.Sprintf("Found %d unreachable nodes: %s", len(unreachableNodes), strings.Join(unreachableNodes, ", ")),
			Component: "WorkflowGraph",
		})
	}

	// Check for dead-end nodes (nodes with no outgoing edges)
	deadEndNodes := wv.findDeadEndNodes(graph)
	if len(deadEndNodes) > 1 { // More than one dead-end is suspicious
		warnings = append(warnings, core.WorkflowValidationWarning{
			Type:      "MultipleDeadEnds",
			Message:   fmt.Sprintf("Found multiple dead-end nodes: %s", strings.Join(deadEndNodes, ", ")),
			Component: "WorkflowGraph",
		})
	}

	// Validate graph connectivity
	if !wv.isConnected(graph) {
		errors = append(errors, core.WorkflowValidationError{
			Type:        "DisconnectedGraph",
			Message:     "Workflow graph is not connected",
			Severity:    "High",
			Component:   "WorkflowGraph",
			Suggestions: []string{"Ensure all workflow steps are connected through dependencies"},
		})
	}

	return core.WorkflowValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// =============================================================================
// COMPOSITION VALIDATION METHODS
// =============================================================================

// validateParallelComposition validates parallel composition specific rules
func (wv *WorkflowValidatorImplementation) validateParallelComposition(agents []core.Agent) []core.WorkflowValidationError {
	var errors []core.WorkflowValidationError

	// Check if all agents can handle parallel execution
	for i, agent := range agents {
		if !wv.supportsParallelExecution(agent) {
			errors = append(errors, core.WorkflowValidationError{
				Type:        "ParallelExecutionNotSupported",
				Message:     fmt.Sprintf("Agent %d (%s) may not support parallel execution", i, agent.Name()),
				Severity:    "Medium",
				Component:   fmt.Sprintf("Agent[%d]", i),
				Suggestions: []string{"Verify agent thread-safety", "Consider using sequential composition instead"},
			})
		}
	}

	return errors
}

// warnParallelComposition generates warnings for parallel compositions
func (wv *WorkflowValidatorImplementation) warnParallelComposition(agents []core.Agent) []core.WorkflowValidationWarning {
	var warnings []core.WorkflowValidationWarning

	if len(agents) > 10 {
		warnings = append(warnings, core.WorkflowValidationWarning{
			Type:      "HighParallelism",
			Message:   fmt.Sprintf("High number of parallel agents (%d) may impact performance", len(agents)),
			Component: "Composition",
		})
	}

	return warnings
}

// validateSequentialComposition validates sequential composition specific rules
func (wv *WorkflowValidatorImplementation) validateSequentialComposition(agents []core.Agent) []core.WorkflowValidationError {
	var errors []core.WorkflowValidationError

	// Check output/input compatibility between sequential agents
	for i := 0; i < len(agents)-1; i++ {
		currentAgent := agents[i]
		nextAgent := agents[i+1]

		if !wv.areCompatible(currentAgent, nextAgent) {
			errors = append(errors, core.WorkflowValidationError{
				Type:        "IncompatibleSequentialAgents",
				Message:     fmt.Sprintf("Agent %d output may not be compatible with Agent %d input", i, i+1),
				Severity:    "High",
				Component:   fmt.Sprintf("Agent[%d]->Agent[%d]", i, i+1),
				Suggestions: []string{"Add data transformation between agents", "Verify agent input/output types"},
			})
		}
	}

	return errors
}

// warnSequentialComposition generates warnings for sequential compositions
func (wv *WorkflowValidatorImplementation) warnSequentialComposition(agents []core.Agent) []core.WorkflowValidationWarning {
	var warnings []core.WorkflowValidationWarning

	if len(agents) > 5 {
		warnings = append(warnings, core.WorkflowValidationWarning{
			Type:      "LongSequentialChain",
			Message:   fmt.Sprintf("Long sequential chain (%d agents) may have high latency", len(agents)),
			Component: "Composition",
		})
	}

	return warnings
}

// validateLoopComposition validates loop composition specific rules
func (wv *WorkflowValidatorImplementation) validateLoopComposition(agents []core.Agent) []core.WorkflowValidationError {
	var errors []core.WorkflowValidationError

	if len(agents) != 1 {
		errors = append(errors, core.WorkflowValidationError{
			Type:        "InvalidLoopAgentCount",
			Message:     fmt.Sprintf("Loop composition must have exactly 1 agent, found %d", len(agents)),
			Severity:    "Critical",
			Component:   "Composition",
			Suggestions: []string{"Use exactly one agent for loop composition"},
		})
	}

	if len(agents) == 1 {
		agent := agents[0]
		if !wv.supportsLooping(agent) {
			errors = append(errors, core.WorkflowValidationError{
				Type:        "LoopingNotSupported",
				Message:     fmt.Sprintf("Agent %s may not support loop execution", agent.Name()),
				Severity:    "High",
				Component:   "Agent[0]",
				Suggestions: []string{"Verify agent supports iterative execution", "Implement proper state management"},
			})
		}
	}

	return errors
}

// warnLoopComposition generates warnings for loop compositions
func (wv *WorkflowValidatorImplementation) warnLoopComposition(agents []core.Agent) []core.WorkflowValidationWarning {
	var warnings []core.WorkflowValidationWarning

	warnings = append(warnings, core.WorkflowValidationWarning{
		Type:      "InfiniteLoopRisk",
		Message:   "Loop composition may run indefinitely without proper termination conditions",
		Component: "Composition",
	})

	return warnings
}

// =============================================================================
// ORCHESTRATION VALIDATION METHODS
// =============================================================================

// validateCollaborativeOrchestration validates collaborative orchestration
func (wv *WorkflowValidatorImplementation) validateCollaborativeOrchestration(agents map[string]core.AgentHandler) []core.WorkflowValidationError {
	var errors []core.WorkflowValidationError

	// Check if all agents can collaborate
	for name, handler := range agents {
		if !wv.supportsCollaboration(handler) {
			errors = append(errors, core.WorkflowValidationError{
				Type:        "CollaborationNotSupported",
				Message:     fmt.Sprintf("Agent handler '%s' may not support collaborative execution", name),
				Severity:    "Medium",
				Component:   name,
				Suggestions: []string{"Implement collaboration interfaces", "Use routing orchestration instead"},
			})
		}
	}

	return errors
}

// warnCollaborativeOrchestration generates warnings for collaborative orchestration
func (wv *WorkflowValidatorImplementation) warnCollaborativeOrchestration(agents map[string]core.AgentHandler) []core.WorkflowValidationWarning {
	var warnings []core.WorkflowValidationWarning

	if len(agents) > 5 {
		warnings = append(warnings, core.WorkflowValidationWarning{
			Type:      "HighCollaboratorCount",
			Message:   fmt.Sprintf("High number of collaborators (%d) may cause coordination overhead", len(agents)),
			Component: "Orchestration",
		})
	}

	return warnings
}

// validateRoutingOrchestration validates routing orchestration
func (wv *WorkflowValidatorImplementation) validateRoutingOrchestration(agents map[string]core.AgentHandler) []core.WorkflowValidationError {
	var errors []core.WorkflowValidationError

	// Check for unique routing keys
	routingKeys := make(map[string]bool)
	for name := range agents {
		if routingKeys[name] {
			errors = append(errors, core.WorkflowValidationError{
				Type:        "DuplicateRoutingKey",
				Message:     fmt.Sprintf("Duplicate routing key found: %s", name),
				Severity:    "Critical",
				Component:   "Orchestration",
				Suggestions: []string{"Ensure all routing keys are unique"},
			})
		}
		routingKeys[name] = true
	}

	return errors
}

// warnRoutingOrchestration generates warnings for routing orchestration
func (wv *WorkflowValidatorImplementation) warnRoutingOrchestration(agents map[string]core.AgentHandler) []core.WorkflowValidationWarning {
	var warnings []core.WorkflowValidationWarning

	if len(agents) == 1 {
		warnings = append(warnings, core.WorkflowValidationWarning{
			Type:      "SingleRouteTarget",
			Message:   "Routing orchestration with single target is inefficient",
			Component: "Orchestration",
		})
	}

	return warnings
}

// validateSequentialOrchestration validates sequential orchestration
func (wv *WorkflowValidatorImplementation) validateSequentialOrchestration(agents map[string]core.AgentHandler) []core.WorkflowValidationError {
	return []core.WorkflowValidationError{} // Sequential orchestration is generally safe
}

// warnSequentialOrchestration generates warnings for sequential orchestration
func (wv *WorkflowValidatorImplementation) warnSequentialOrchestration(agents map[string]core.AgentHandler) []core.WorkflowValidationWarning {
	var warnings []core.WorkflowValidationWarning

	if len(agents) > 10 {
		warnings = append(warnings, core.WorkflowValidationWarning{
			Type:      "LongSequentialOrchestration",
			Message:   fmt.Sprintf("Long sequential orchestration (%d agents) may have high latency", len(agents)),
			Component: "Orchestration",
		})
	}

	return warnings
}

// validateParallelOrchestration validates parallel orchestration
func (wv *WorkflowValidatorImplementation) validateParallelOrchestration(agents map[string]core.AgentHandler) []core.WorkflowValidationError {
	return []core.WorkflowValidationError{} // Parallel orchestration is generally safe
}

// warnParallelOrchestration generates warnings for parallel orchestration
func (wv *WorkflowValidatorImplementation) warnParallelOrchestration(agents map[string]core.AgentHandler) []core.WorkflowValidationWarning {
	var warnings []core.WorkflowValidationWarning

	if len(agents) > 20 {
		warnings = append(warnings, core.WorkflowValidationWarning{
			Type:      "HighParallelOrchestration",
			Message:   fmt.Sprintf("High number of parallel agents (%d) may impact performance", len(agents)),
			Component: "Orchestration",
		})
	}

	return warnings
}

// validateLoopOrchestration validates loop orchestration
func (wv *WorkflowValidatorImplementation) validateLoopOrchestration(agents map[string]core.AgentHandler) []core.WorkflowValidationError {
	var errors []core.WorkflowValidationError

	if len(agents) == 0 {
		errors = append(errors, core.WorkflowValidationError{
			Type:        "EmptyLoopOrchestration",
			Message:     "Loop orchestration requires at least one agent handler",
			Severity:    "Critical",
			Component:   "Orchestration",
			Suggestions: []string{"Add at least one agent handler for loop execution"},
		})
	}

	return errors
}

// warnLoopOrchestration generates warnings for loop orchestration
func (wv *WorkflowValidatorImplementation) warnLoopOrchestration(agents map[string]core.AgentHandler) []core.WorkflowValidationWarning {
	var warnings []core.WorkflowValidationWarning

	warnings = append(warnings, core.WorkflowValidationWarning{
		Type:      "InfiniteLoopRisk",
		Message:   "Loop orchestration may run indefinitely without proper termination conditions",
		Component: "Orchestration",
	})

	return warnings
}

// =============================================================================
// GRAPH VALIDATION METHODS
// =============================================================================

// detectCycles uses DFS to detect cycles in the workflow graph
func (wv *WorkflowValidatorImplementation) detectCycles(graph core.WorkflowGraph) bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range graph.Nodes {
		if !visited[nodeID] {
			if wv.dfsHasCycle(graph, nodeID, visited, recStack) {
				return true
			}
		}
	}

	return false
}

// dfsHasCycle performs DFS to detect cycles
func (wv *WorkflowValidatorImplementation) dfsHasCycle(graph core.WorkflowGraph, nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	// Check all adjacent nodes
	for _, edge := range graph.Edges {
		if edge.From == nodeID {
			neighbor := edge.To
			if !visited[neighbor] {
				if wv.dfsHasCycle(graph, neighbor, visited, recStack) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}
	}

	recStack[nodeID] = false
	return false
}

// findUnreachableNodes finds nodes that cannot be reached from any source
func (wv *WorkflowValidatorImplementation) findUnreachableNodes(graph core.WorkflowGraph) []string {
	reachable := make(map[string]bool)

	// Find all source nodes (nodes with no incoming edges)
	hasIncoming := make(map[string]bool)
	for _, edge := range graph.Edges {
		hasIncoming[edge.To] = true
	}

	// Start DFS from all source nodes
	for nodeID := range graph.Nodes {
		if !hasIncoming[nodeID] {
			wv.dfsMarkReachable(graph, nodeID, reachable)
		}
	}

	// Find unreachable nodes
	var unreachable []string
	for nodeID := range graph.Nodes {
		if !reachable[nodeID] {
			unreachable = append(unreachable, nodeID)
		}
	}

	return unreachable
}

// dfsMarkReachable marks all reachable nodes from a given node
func (wv *WorkflowValidatorImplementation) dfsMarkReachable(graph core.WorkflowGraph, nodeID string, reachable map[string]bool) {
	reachable[nodeID] = true

	for _, edge := range graph.Edges {
		if edge.From == nodeID && !reachable[edge.To] {
			wv.dfsMarkReachable(graph, edge.To, reachable)
		}
	}
}

// findDeadEndNodes finds nodes with no outgoing edges
func (wv *WorkflowValidatorImplementation) findDeadEndNodes(graph core.WorkflowGraph) []string {
	hasOutgoing := make(map[string]bool)

	for _, edge := range graph.Edges {
		hasOutgoing[edge.From] = true
	}

	var deadEnds []string
	for nodeID := range graph.Nodes {
		if !hasOutgoing[nodeID] {
			deadEnds = append(deadEnds, nodeID)
		}
	}

	return deadEnds
}

// isConnected checks if the graph is connected (ignoring edge direction)
func (wv *WorkflowValidatorImplementation) isConnected(graph core.WorkflowGraph) bool {
	if len(graph.Nodes) == 0 {
		return true
	}

	// Build adjacency list (undirected)
	adj := make(map[string][]string)
	for nodeID := range graph.Nodes {
		adj[nodeID] = []string{}
	}

	for _, edge := range graph.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
		adj[edge.To] = append(adj[edge.To], edge.From)
	}

	// DFS from first node
	visited := make(map[string]bool)
	var startNode string
	for nodeID := range graph.Nodes {
		startNode = nodeID
		break
	}

	wv.dfsVisit(adj, startNode, visited)

	// Check if all nodes were visited
	return len(visited) == len(graph.Nodes)
}

// dfsVisit performs DFS for connectivity check
func (wv *WorkflowValidatorImplementation) dfsVisit(adj map[string][]string, nodeID string, visited map[string]bool) {
	visited[nodeID] = true

	for _, neighbor := range adj[nodeID] {
		if !visited[neighbor] {
			wv.dfsVisit(adj, neighbor, visited)
		}
	}
}

// =============================================================================
// AGENT VALIDATION HELPERS
// =============================================================================

// validateAgent validates an individual agent
func (wv *WorkflowValidatorImplementation) validateAgent(agent core.Agent, component string) []core.WorkflowValidationError {
	var errors []core.WorkflowValidationError

	if agent == nil {
		errors = append(errors, core.WorkflowValidationError{
			Type:        "NilAgent",
			Message:     "Agent is nil",
			Severity:    "Critical",
			Component:   component,
			Suggestions: []string{"Ensure agent is properly initialized"},
		})
		return errors
	}

	if agent.Name() == "" {
		errors = append(errors, core.WorkflowValidationError{
			Type:        "EmptyAgentName",
			Message:     "Agent name is empty",
			Severity:    "Medium",
			Component:   component,
			Suggestions: []string{"Provide a meaningful agent name"},
		})
	}

	return errors
}

// validateAgentHandler validates an agent handler
func (wv *WorkflowValidatorImplementation) validateAgentHandler(handler core.AgentHandler, name string) []core.WorkflowValidationError {
	var errors []core.WorkflowValidationError

	if handler == nil {
		errors = append(errors, core.WorkflowValidationError{
			Type:        "NilAgentHandler",
			Message:     fmt.Sprintf("Agent handler '%s' is nil", name),
			Severity:    "Critical",
			Component:   name,
			Suggestions: []string{"Ensure agent handler is properly initialized"},
		})
	}

	return errors
}

// =============================================================================
// AGENT CAPABILITY CHECKS
// =============================================================================

// supportsParallelExecution checks if an agent supports parallel execution
func (wv *WorkflowValidatorImplementation) supportsParallelExecution(agent core.Agent) bool {
	// This is a heuristic check - in practice, you'd implement proper interfaces
	// For now, we assume all agents support parallel execution unless proven otherwise
	return true
}

// areCompatible checks if two agents are compatible in sequence
func (wv *WorkflowValidatorImplementation) areCompatible(current, next core.Agent) bool {
	// This is a heuristic check - in practice, you'd check input/output types
	// For now, we assume all agents are compatible unless proven otherwise
	return true
}

// supportsLooping checks if an agent supports loop execution
func (wv *WorkflowValidatorImplementation) supportsLooping(agent core.Agent) bool {
	// This is a heuristic check - in practice, you'd check for statelessness
	// For now, we assume all agents support looping unless proven otherwise
	return true
}

// supportsCollaboration checks if an agent handler supports collaboration
func (wv *WorkflowValidatorImplementation) supportsCollaboration(handler core.AgentHandler) bool {
	// This is a heuristic check - in practice, you'd check for collaboration interfaces
	// For now, we assume all handlers support collaboration unless proven otherwise
	return true
}
