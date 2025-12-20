package workflow

import (
	"fmt"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	"github.com/agenticgokit/agenticgokit/examples/story-writer-chat-v2/infrastructure"
)

// StreamHandler processes streaming chunks from the workflow execution
type StreamHandler struct {
	sendMessage             infrastructure.MessageSender
	currentAgent            string
	currentAgentDisplay     string
	agentContent            string
	finalContent            string
	isSubWorkflow           bool
	seenSubworkflowStart    bool
	seenSubworkflowComplete bool
	seenAgents              map[string]bool
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(sendMessage infrastructure.MessageSender) *StreamHandler {
	return &StreamHandler{
		sendMessage: sendMessage,
		seenAgents:  make(map[string]bool),
	}
}

// HandleAgentStart processes ChunkTypeAgentStart events
func (h *StreamHandler) HandleAgentStart(chunk *vnext.StreamChunk) {
	stepName := ""
	if name, ok := chunk.Metadata["step_name"].(string); ok {
		stepName = name
	}

	// Check if this is from a nested agent inside a SubWorkflow
	isNestedAgent := false
	if _, hasParent := chunk.Metadata["parent_subworkflow"]; hasParent {
		isNestedAgent = true
	}

	// Skip duplicate SubWorkflow container starts (not nested agents - they run multiple times in loops)
	agentKey := stepName
	if h.seenAgents[agentKey] && stepName == "revisions" && !isNestedAgent {
		return
	}
	// Only mark as seen if it's NOT a nested agent (nested agents can emit multiple starts in loops)
	if !isNestedAgent {
		h.seenAgents[agentKey] = true
	}

	h.currentAgent = stepName
	h.agentContent = ""
	h.isSubWorkflow = false

	// Map internal step names to display names
	displayName, displayIcon, startMsg := h.mapStepToDisplay(stepName)
	h.currentAgentDisplay = displayName

	// For SubWorkflows, send workflow info start message (only once)
	if h.isSubWorkflow && !isNestedAgent {
		if !h.seenSubworkflowStart {
			h.sendMessage(infrastructure.WSMessage{
				Type:      infrastructure.MsgTypeWorkflowInfo,
				Content:   fmt.Sprintf("%s %s", displayIcon, startMsg),
				Timestamp: float64(time.Now().Unix()),
				Metadata: map[string]interface{}{
					"subworkflow": displayName,
					"event":       "start",
				},
			})
			h.seenSubworkflowStart = true
		}
	} else {
		// Regular agents OR nested agents inside SubWorkflows
		h.sendMessage(infrastructure.WSMessage{
			Type:      infrastructure.MsgTypeAgentStart,
			Agent:     displayName,
			Content:   fmt.Sprintf("%s %s", displayIcon, startMsg),
			Timestamp: float64(time.Now().Unix()),
		})
	}
}

// HandleContent processes ChunkTypeText and ChunkTypeDelta events
func (h *StreamHandler) HandleContent(chunk *vnext.StreamChunk) {
	content := chunk.Content
	if chunk.Type == vnext.ChunkTypeDelta {
		content = chunk.Delta
	}

	// Check if this chunk is from a nested agent
	isNestedAgent := false
	if _, hasParent := chunk.Metadata["parent_subworkflow"]; hasParent {
		isNestedAgent = true

		// Update currentAgentDisplay based on step_name for nested agents
		if stepName, ok := chunk.Metadata["step_name"].(string); ok {
			switch stepName {
			case "write":
				h.currentAgentDisplay = "writer"
			case "review":
				h.currentAgentDisplay = "editor"
			}
		}
	}

	// Skip content only for SubWorkflow itself, not nested agents
	if h.isSubWorkflow && !isNestedAgent {
		return
	}

	h.agentContent += content
	h.finalContent += content

	h.sendMessage(infrastructure.WSMessage{
		Type:      infrastructure.MsgTypeAgentProgress,
		Agent:     h.currentAgentDisplay,
		Content:   content,
		Timestamp: float64(time.Now().Unix()),
	})
}

// HandleAgentComplete processes ChunkTypeAgentComplete events
func (h *StreamHandler) HandleAgentComplete(chunk *vnext.StreamChunk) {
	success := true
	if s, ok := chunk.Metadata["success"].(bool); ok {
		success = s
	}

	// Check if this is a nested agent completion
	isNestedAgentComplete := false
	completedAgentName := h.currentAgentDisplay
	if _, hasParent := chunk.Metadata["parent_subworkflow"]; hasParent {
		isNestedAgentComplete = true
		// Update agent name based on step_name
		if stepName, ok := chunk.Metadata["step_name"].(string); ok {
			switch stepName {
			case "write":
				completedAgentName = "writer"
			case "review":
				completedAgentName = "editor"
			}
		}
	}

	// For SubWorkflow container itself (not nested agents), defer the completion message
	if h.isSubWorkflow && !isNestedAgentComplete {
		if !h.seenSubworkflowComplete {
			h.seenSubworkflowComplete = true
		}
		return // Skip sending completion for SubWorkflow container
	}

	// For nested agents or regular agents: send agent_complete
	h.sendMessage(infrastructure.WSMessage{
		Type:      infrastructure.MsgTypeAgentComplete,
		Agent:     completedAgentName,
		Content:   h.agentContent,
		Timestamp: float64(time.Now().Unix()),
		Metadata: map[string]interface{}{
			"success":  success,
			"duration": chunk.Metadata["duration"],
			"tokens":   chunk.Metadata["tokens"],
		},
	})

	// Reset for next agent
	h.agentContent = ""
}

// HandleMetadata processes ChunkTypeMetadata events
func (h *StreamHandler) HandleMetadata(chunk *vnext.StreamChunk) {
	// Loop iteration updates
	if iteration, ok := chunk.Metadata["iteration"].(int); ok && h.currentAgent == "revisions" {
		h.sendMessage(infrastructure.WSMessage{
			Type:      infrastructure.MsgTypeWorkflowInfo,
			Content:   fmt.Sprintf("🔄 Iteration %d starting...", iteration),
			Timestamp: float64(time.Now().Unix()),
			Metadata: map[string]interface{}{
				"subworkflow": "revision_loop",
				"event":       "iteration",
				"iteration":   iteration,
			},
		})
	} else if exitReason, ok := chunk.Metadata["exit_reason"].(string); ok {
		// Loop exit condition met
		h.handleLoopExit(exitReason, chunk.Metadata["total_iterations"])
	}
}

// HandleError processes ChunkTypeError events
func (h *StreamHandler) HandleError(chunk *vnext.StreamChunk) {
	h.sendMessage(infrastructure.WSMessage{
		Type:      infrastructure.MsgTypeError,
		Content:   chunk.Error.Error(),
		Agent:     h.currentAgentDisplay,
		Timestamp: float64(time.Now().Unix()),
	})
}

// GetFinalContent returns the accumulated final content
func (h *StreamHandler) GetFinalContent() string {
	return h.finalContent
}

// mapStepToDisplay maps internal step names to display information
func (h *StreamHandler) mapStepToDisplay(stepName string) (displayName, icon, message string) {
	switch stepName {
	case "write":
		return "writer", "✍️", "Starting writer..."
	case "review":
		return "editor", "✏️", "Starting editor..."
	case "publish":
		return "publisher", "📚", "Starting publisher..."
	case "revisions":
		h.isSubWorkflow = true
		return "revision_loop", "🔄", "Starting revision loop (Writer ↔ Editor until approved)"
	default:
		return stepName, "", fmt.Sprintf("Starting %s...", stepName)
	}
}

// handleLoopExit sends messages when the loop exits
func (h *StreamHandler) handleLoopExit(exitReason string, totalIter interface{}) {
	var exitMsg string
	switch exitReason {
	case "condition_met":
		exitMsg = fmt.Sprintf("✅ Story approved after %v iteration(s)", totalIter)
	case "max_iterations":
		exitMsg = fmt.Sprintf("⚠️ Reached maximum iterations (%v)", totalIter)
	case "convergence":
		exitMsg = fmt.Sprintf("✅ Output converged after %v iteration(s)", totalIter)
	default:
		exitMsg = fmt.Sprintf("✅ Loop completed: %s (%v iteration(s))", exitReason, totalIter)
	}

	h.sendMessage(infrastructure.WSMessage{
		Type:      infrastructure.MsgTypeWorkflowInfo,
		Content:   exitMsg,
		Timestamp: float64(time.Now().Unix()),
		Metadata: map[string]interface{}{
			"subworkflow":      "revision_loop",
			"event":            "exit",
			"exit_reason":      exitReason,
			"total_iterations": totalIter,
		},
	})

	// Send workflow_info for chat
	h.sendMessage(infrastructure.WSMessage{
		Type:      infrastructure.MsgTypeWorkflowInfo,
		Content:   "🔄 Revision loop complete",
		Timestamp: float64(time.Now().Unix()),
		Metadata: map[string]interface{}{
			"subworkflow": "revision_loop",
			"event":       "complete",
		},
	})

	// Send agent_complete for status panel
	h.sendMessage(infrastructure.WSMessage{
		Type:      infrastructure.MsgTypeAgentComplete,
		Agent:     "revision_loop",
		Content:   "",
		Timestamp: float64(time.Now().Unix()),
		Metadata: map[string]interface{}{
			"success": true,
		},
	})
}



