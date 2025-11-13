package infrastructure

import "context"

// AgentInfo represents metadata about an agent
type AgentInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// WorkflowExecutor defines the interface that any workflow must implement
// This allows different workflows to be plugged into the WebSocket server
type WorkflowExecutor interface {
	// Name returns the workflow name (e.g., "Story Writer", "Data Analyzer")
	Name() string

	// WelcomeMessage returns the welcome message shown to users
	WelcomeMessage() string

	// GetAgents returns the list of agents in this workflow for UI display
	GetAgents() []AgentInfo

	// Execute runs the workflow with the given user input
	// sendMessage is used to stream updates back to the client
	Execute(ctx context.Context, userInput string, sendMessage MessageSender) error

	// Cleanup performs any necessary cleanup (optional)
	Cleanup(ctx context.Context) error
}
