// Package core provides workflow validation for ensuring agent chain integrity
package core

import (
	"fmt"
	"strings"
)

// WorkflowValidationError represents validation errors in agent workflows
type WorkflowValidationError struct {
	Type    ValidationErrorType    `json:"type"`
	Message string                 `json:"message"`
	Agent   string                 `json:"agent,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *WorkflowValidationError) Error() string {
	if e.Agent != "" {
		return fmt.Sprintf("workflow validation error for agent '%s': %s", e.Agent, e.Message)
	}
	return fmt.Sprintf("workflow validation error: %s", e.Message)
}

// ValidationErrorType represents different types of validation errors
type ValidationErrorType string

const (
	ValidationErrorOrphanedAgent      ValidationErrorType = "ORPHANED_AGENT"
	ValidationErrorCircularDependency ValidationErrorType = "CIRCULAR_DEPENDENCY"
	ValidationErrorMissingAgent       ValidationErrorType = "MISSING_AGENT"
	ValidationErrorInvalidRouting     ValidationErrorType = "INVALID_ROUTING"
	ValidationErrorNoEntryPoint       ValidationErrorType = "NO_ENTRY_POINT"
	ValidationErrorMultipleEndpoints  ValidationErrorType = "MULTIPLE_ENDPOINTS"
)

// WorkflowValidator provides validation for agent workflows
type WorkflowValidator struct {
	agents      map[string]*AgentNode
	entryPoints []string
	endpoints   []string
	strictMode  bool
}

// AgentNode represents an agent in the workflow graph
type AgentNode struct {
	Name         string                 `json:"name"`
	Type         AgentType              `json:"type"`
	Dependencies []string               `json:"dependencies"`
	Routes       []string               `json:"routes"`
	IsEntryPoint bool                   `json:"is_entry_point"`
	IsEndpoint   bool                   `json:"is_endpoint"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AgentType categorizes different types of agents
type AgentType string

const (
	AgentTypeStandard      AgentType = "STANDARD"
	AgentTypeErrorHandler  AgentType = "ERROR_HANDLER"
	AgentTypeResponsibleAI AgentType = "RESPONSIBLE_AI"
	AgentTypeFinalizer     AgentType = "FINALIZER"
)

// NewWorkflowValidator creates a new workflow validator
func NewWorkflowValidator(strictMode bool) *WorkflowValidator {
	return &WorkflowValidator{
		agents:     make(map[string]*AgentNode),
		strictMode: strictMode,
	}
}

// AddAgent adds an agent to the workflow for validation
func (v *WorkflowValidator) AddAgent(name string, agentType AgentType) *AgentNode {
	node := &AgentNode{
		Name:         name,
		Type:         agentType,
		Dependencies: []string{},
		Routes:       []string{},
		Metadata:     make(map[string]interface{}),
	}
	v.agents[name] = node
	return node
}

// AddRoute adds a routing relationship between agents
func (v *WorkflowValidator) AddRoute(fromAgent, toAgent string) error {
	from, exists := v.agents[fromAgent]
	if !exists {
		return fmt.Errorf("source agent '%s' not found", fromAgent)
	}

	if _, exists := v.agents[toAgent]; !exists {
		return fmt.Errorf("destination agent '%s' not found", toAgent)
	}

	// Add route if not already present
	for _, route := range from.Routes {
		if route == toAgent {
			return nil // Route already exists
		}
	}

	from.Routes = append(from.Routes, toAgent)

	// Add dependency to destination agent
	to := v.agents[toAgent]
	for _, dep := range to.Dependencies {
		if dep == fromAgent {
			return nil // Dependency already exists
		}
	}
	to.Dependencies = append(to.Dependencies, fromAgent)

	return nil
}

// SetEntryPoint marks an agent as an entry point to the workflow
func (v *WorkflowValidator) SetEntryPoint(agentName string) error {
	agent, exists := v.agents[agentName]
	if !exists {
		return fmt.Errorf("agent '%s' not found", agentName)
	}

	agent.IsEntryPoint = true
	v.entryPoints = append(v.entryPoints, agentName)
	return nil
}

// SetEndpoint marks an agent as an endpoint of the workflow
func (v *WorkflowValidator) SetEndpoint(agentName string) error {
	agent, exists := v.agents[agentName]
	if !exists {
		return fmt.Errorf("agent '%s' not found", agentName)
	}

	agent.IsEndpoint = true
	v.endpoints = append(v.endpoints, agentName)
	return nil
}

// ValidateWorkflow performs comprehensive validation of the workflow
func (v *WorkflowValidator) ValidateWorkflow() []*WorkflowValidationError {
	var errors []*WorkflowValidationError

	// Check for entry points
	if len(v.entryPoints) == 0 {
		errors = append(errors, &WorkflowValidationError{
			Type:    ValidationErrorNoEntryPoint,
			Message: "workflow has no entry points defined",
		})
	}

	// Check for orphaned agents
	errors = append(errors, v.checkOrphanedAgents()...)

	// Check for circular dependencies
	errors = append(errors, v.checkCircularDependencies()...)

	// Check for missing agent references
	errors = append(errors, v.checkMissingAgents()...)

	// Check routing validity
	errors = append(errors, v.checkRoutingValidity()...)

	// Check for unreachable agents
	errors = append(errors, v.checkUnreachableAgents()...)

	return errors
}

// checkOrphanedAgents finds agents with no incoming or outgoing connections
func (v *WorkflowValidator) checkOrphanedAgents() []*WorkflowValidationError {
	var errors []*WorkflowValidationError

	for name, agent := range v.agents {
		hasIncoming := len(agent.Dependencies) > 0 || agent.IsEntryPoint
		hasOutgoing := len(agent.Routes) > 0 || agent.IsEndpoint

		if !hasIncoming && !hasOutgoing {
			errors = append(errors, &WorkflowValidationError{
				Type:    ValidationErrorOrphanedAgent,
				Message: "agent has no incoming or outgoing connections",
				Agent:   name,
				Details: map[string]interface{}{
					"is_entry_point": agent.IsEntryPoint,
					"is_endpoint":    agent.IsEndpoint,
				},
			})
		}
	}

	return errors
}

// checkCircularDependencies detects circular dependencies in the workflow
func (v *WorkflowValidator) checkCircularDependencies() []*WorkflowValidationError {
	var errors []*WorkflowValidationError
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for name := range v.agents {
		if !visited[name] {
			if cycle := v.detectCycle(name, visited, recursionStack, []string{}); len(cycle) > 0 {
				errors = append(errors, &WorkflowValidationError{
					Type:    ValidationErrorCircularDependency,
					Message: fmt.Sprintf("circular dependency detected in path: %s", strings.Join(cycle, " -> ")),
					Details: map[string]interface{}{
						"cycle_path": cycle,
					},
				})
			}
		}
	}

	return errors
}

// detectCycle performs DFS to detect cycles
func (v *WorkflowValidator) detectCycle(agent string, visited, recursionStack map[string]bool, path []string) []string {
	visited[agent] = true
	recursionStack[agent] = true
	currentPath := append(path, agent)

	agentNode := v.agents[agent]
	for _, route := range agentNode.Routes {
		if !visited[route] {
			if cycle := v.detectCycle(route, visited, recursionStack, currentPath); len(cycle) > 0 {
				return cycle
			}
		} else if recursionStack[route] {
			// Found cycle
			cycleStart := -1
			for i, p := range currentPath {
				if p == route {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				return append(currentPath[cycleStart:], route)
			}
		}
	}

	recursionStack[agent] = false
	return nil
}

// checkMissingAgents finds references to non-existent agents
func (v *WorkflowValidator) checkMissingAgents() []*WorkflowValidationError {
	var errors []*WorkflowValidationError

	for name, agent := range v.agents {
		for _, route := range agent.Routes {
			if _, exists := v.agents[route]; !exists {
				errors = append(errors, &WorkflowValidationError{
					Type:    ValidationErrorMissingAgent,
					Message: fmt.Sprintf("routes to non-existent agent '%s'", route),
					Agent:   name,
					Details: map[string]interface{}{
						"missing_agent": route,
					},
				})
			}
		}
	}

	return errors
}

// checkRoutingValidity ensures routing logic is valid
func (v *WorkflowValidator) checkRoutingValidity() []*WorkflowValidationError {
	var errors []*WorkflowValidationError

	for name, agent := range v.agents {
		// Check for invalid routing patterns
		if agent.Type == AgentTypeFinalizer && len(agent.Routes) > 0 {
			errors = append(errors, &WorkflowValidationError{
				Type:    ValidationErrorInvalidRouting,
				Message: "finalizer agent should not have outgoing routes",
				Agent:   name,
				Details: map[string]interface{}{
					"routes": agent.Routes,
				},
			})
		}

		// Error handlers should typically be endpoints
		if agent.Type == AgentTypeErrorHandler && len(agent.Routes) > 0 && v.strictMode {
			errors = append(errors, &WorkflowValidationError{
				Type:    ValidationErrorInvalidRouting,
				Message: "error handler should typically be an endpoint (strict mode)",
				Agent:   name,
				Details: map[string]interface{}{
					"routes": agent.Routes,
				},
			})
		}
	}

	return errors
}

// checkUnreachableAgents finds agents that cannot be reached from entry points
func (v *WorkflowValidator) checkUnreachableAgents() []*WorkflowValidationError {
	var errors []*WorkflowValidationError

	if len(v.entryPoints) == 0 {
		return errors // Already reported as no entry points
	}

	reachable := make(map[string]bool)

	// BFS from all entry points
	queue := make([]string, len(v.entryPoints))
	copy(queue, v.entryPoints)

	for _, entry := range v.entryPoints {
		reachable[entry] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		agent := v.agents[current]
		for _, route := range agent.Routes {
			if !reachable[route] {
				reachable[route] = true
				queue = append(queue, route)
			}
		}
	}

	// Check for unreachable agents
	for name := range v.agents {
		if !reachable[name] {
			errors = append(errors, &WorkflowValidationError{
				Type:    ValidationErrorOrphanedAgent,
				Message: "agent is unreachable from entry points",
				Agent:   name,
				Details: map[string]interface{}{
					"entry_points": v.entryPoints,
				},
			})
		}
	}

	return errors
}

// GetWorkflowSummary returns a summary of the workflow structure
func (v *WorkflowValidator) GetWorkflowSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"total_agents":   len(v.agents),
		"entry_points":   v.entryPoints,
		"endpoints":      v.endpoints,
		"strict_mode":    v.strictMode,
		"agents_by_type": make(map[AgentType][]string),
	}

	agentsByType := make(map[AgentType][]string)
	for name, agent := range v.agents {
		agentsByType[agent.Type] = append(agentsByType[agent.Type], name)
	}
	summary["agents_by_type"] = agentsByType

	return summary
}

// GetWorkflowGraph returns the workflow as a graph structure for visualization
func (v *WorkflowValidator) GetWorkflowGraph() map[string]interface{} {
	nodes := []map[string]interface{}{}
	edges := []map[string]interface{}{}

	for name, agent := range v.agents {
		node := map[string]interface{}{
			"id":             name,
			"type":           string(agent.Type),
			"is_entry_point": agent.IsEntryPoint,
			"is_endpoint":    agent.IsEndpoint,
		}
		nodes = append(nodes, node)

		for _, route := range agent.Routes {
			edge := map[string]interface{}{
				"source": name,
				"target": route,
			}
			edges = append(edges, edge)
		}
	}

	return map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	}
}
