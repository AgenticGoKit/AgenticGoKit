package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateAgentProject creates a new AgentFlow project scaffold.
func CreateAgentProject(agentName string, numAgents int, responsibleAI bool, errorHandler bool) error {
	// Create the main project directory
	if err := os.Mkdir(agentName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory %s: %w", agentName, err)
	}

	fmt.Printf("Created directory: %s\n", agentName)

	// Create the main.go file
	mainGoContent := fmt.Sprintf(`package main

import "fmt"

func main() {
	fmt.Println("Hello from %s!")
}
`, agentName)

	mainGoPath := filepath.Join(agentName, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go in %s: %w", agentName, err)
	}

	fmt.Printf("Created file: %s\n", mainGoPath)

	// Helper function to generate agent.go content
	getAgentGoContent := func(agentIndex int) string {
		// Corrected format specifiers: use %%s for literal %s in Printf, and %d for agentIndex
		return fmt.Sprintf(`package main // Or potentially a specific package for this agent

import (
	"fmt"
	"github.com/kunalkushwaha/agentflow/core" // Assuming this is a relevant import
)

// Define a struct for your agent
type Agent%d struct {
	Name string
}

// Implement the core.Agent interface (or a relevant interface for AgentFlow)
func (a *Agent%d) Execute(ctx core.Context, event core.Event) core.Result {
	fmt.Printf("Agent%d (Name: %%s) received event: %%v\n", %d, a.Name, event)
	// Placeholder: process the event and return a result
	return core.NewResult(event.ID(), nil) // Example result
}

// A constructor function for the agent
func NewAgent%d(name string) *Agent%d {
	return &Agent%d{Name: name}
}
`, agentIndex, agentIndex, agentIndex, agentIndex, agentIndex, agentIndex, agentIndex)
	}

	if numAgents == 1 {
		agentGoContent := getAgentGoContent(1)
		agentGoPath := filepath.Join(agentName, "agent.go")
		if err := os.WriteFile(agentGoPath, []byte(agentGoContent), 0644); err != nil {
			return fmt.Errorf("failed to create agent.go in %s: %w", agentName, err)
		}
		fmt.Printf("Created file: %s\n", agentGoPath)
	} else if numAgents > 1 {
		for i := 1; i <= numAgents; i++ {
			agentDirName := fmt.Sprintf("agent%d", i)
			agentDirPath := filepath.Join(agentName, agentDirName)

			if err := os.Mkdir(agentDirPath, 0755); err != nil {
				return fmt.Errorf("failed to create agent directory %s: %w", agentDirPath, err)
			}
			fmt.Printf("Created directory: %s\n", agentDirPath)

			agentGoContent := getAgentGoContent(i)
			agentGoPath := filepath.Join(agentDirPath, "agent.go")
			if err := os.WriteFile(agentGoPath, []byte(agentGoContent), 0644); err != nil {
				return fmt.Errorf("failed to create agent.go in %s: %w", agentDirPath, err)
			}
			fmt.Printf("Created file: %s\n", agentGoPath)
		}
	}

	if responsibleAI {
		raiDir := filepath.Join(agentName, "responsible_ai")
		if err := os.Mkdir(raiDir, 0755); err != nil {
			return fmt.Errorf("failed to create responsible_ai directory %s: %w", raiDir, err)
		}
		fmt.Printf("Created directory: %s\n", raiDir)

		raiContent := `package responsible_ai

import (
	"fmt"
	"github.com/kunalkushwaha/agentflow/core"
)

type ResponsibleAIAgent struct {
	Name string
}

func NewResponsibleAIAgent(name string) *ResponsibleAIAgent {
	return &ResponsibleAIAgent{Name: name}
}

func (a *ResponsibleAIAgent) Execute(ctx core.Context, event core.Event) core.Result {
	fmt.Printf("Responsible AI Agent %s checking event: %v\n", a.Name, event)
	// Placeholder: Implement Responsible AI checks (e.g., bias detection, safety guidelines)
	// For now, assume the event is okay and pass it through or return a specific result.
	fmt.Println("Responsible AI: Event conforms to guidelines.")
	return core.NewResult(event.ID(), nil) // Or modify event/result as needed
}
`
		raiPath := filepath.Join(raiDir, "agent.go")
		if err := os.WriteFile(raiPath, []byte(raiContent), 0644); err != nil {
			return fmt.Errorf("failed to create responsible_ai/agent.go: %w", err)
		}
		fmt.Printf("Created file: %s\n", raiPath)
	}

	if errorHandler {
		ehDir := filepath.Join(agentName, "error_handler")
		if err := os.Mkdir(ehDir, 0755); err != nil {
			return fmt.Errorf("failed to create error_handler directory %s: %w", ehDir, err)
		}
		fmt.Printf("Created directory: %s\n", ehDir)

		ehContent := `package error_handler

import (
	"fmt"
	"github.com/kunalkushwaha/agentflow/core"
)

type ErrorHandlerAgent struct {
	Name string
}

func NewErrorHandlerAgent(name string) *ErrorHandlerAgent {
	return &ErrorHandlerAgent{Name: name}
}

// This agent might be triggered differently, perhaps on errors from other agents.
// The Execute method here is a placeholder for its logic.
func (a *ErrorHandlerAgent) Execute(ctx core.Context, event core.Event) core.Result {
	fmt.Printf("Error Handler Agent %s received event: %v\n", a.Name, event)
	if event.Error() != nil {
		fmt.Printf("Error Handler: Detected error - %s\n", event.Error().Error())
		// Placeholder: Implement error handling logic (e.g., logging, retries, notifications)
		// For now, just acknowledge the error.
		// It might return a specific result indicating the error was handled or needs escalation.
		return core.NewResult(event.ID(), fmt.Errorf("error handled: %w", event.Error()))
	}
	fmt.Println("Error Handler: No error detected in this event.")
	return core.NewResult(event.ID(), nil)
}
`
		ehPath := filepath.Join(ehDir, "agent.go")
		if err := os.WriteFile(ehPath, []byte(ehContent), 0644); err != nil {
			return fmt.Errorf("failed to create error_handler/agent.go: %w", err)
		}
		fmt.Printf("Created file: %s\n", ehPath)
	}

	return nil
}
