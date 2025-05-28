package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateAgentProject creates a new AgentFlow project scaffold.
func CreateAgentProject(agentName string, numAgents int, responsibleAI bool, errorHandler bool, provider string) error {
	// Create the main project directory
	if err := os.Mkdir(agentName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory %s: %w", agentName, err)
	}
	fmt.Printf("Created directory: %s\n", agentName)

	// Create go.mod file
	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire (\n\tgithub.com/kunalkushwaha/agentflow v0.1.0\n)\n", agentName)
	goModPath := filepath.Join(agentName, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}
	fmt.Printf("Created file: %s\n", goModPath)

	// Create README.md file
	readmeContent := createReadmeContent(agentName, numAgents, responsibleAI, errorHandler, provider)
	readmePath := filepath.Join(agentName, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	fmt.Printf("Created file: %s\n", readmePath)

	// Create main.go file with provider-specific configuration
	mainGoContent := createMainGoContent(agentName, provider, numAgents, responsibleAI, errorHandler)
	mainGoPath := filepath.Join(agentName, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}
	fmt.Printf("Created file: %s\n", mainGoPath)

	// Create agent files
	if numAgents == 1 {
		if err := createAgentFile(agentName, "agent.go", 1); err != nil {
			return err
		}
	} else {
		// For multiple agents, create all agents in separate files in the main directory
		for i := 1; i <= numAgents; i++ {
			filename := fmt.Sprintf("agent%d.go", i)
			if err := createAgentFile(agentName, filename, i); err != nil {
				return err
			}
		}
	}

	// Create error handler agent if requested
	if errorHandler {
		if err := createErrorHandlerAgent(agentName); err != nil {
			return err
		}
	}

	// Create responsible AI agent if requested
	if responsibleAI {
		if err := createResponsibleAIAgent(agentName); err != nil {
			return err
		}
	}

	// Create workflows directory
	workflowsDir := filepath.Join(agentName, "workflows")
	if err := os.Mkdir(workflowsDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}
	fmt.Printf("Created directory: %s\n", workflowsDir)

	// Create workflow file
	workflowContent := createWorkflowContent(numAgents, responsibleAI, errorHandler)
	workflowPath := filepath.Join(workflowsDir, "main.workflow.md")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		return fmt.Errorf("failed to create workflow file: %w", err)
	}
	fmt.Printf("Created file: %s\n", workflowPath)

	// Create agentflow.toml config file
	configContent := createConfigContent(provider)
	configPath := filepath.Join(agentName, "agentflow.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	fmt.Printf("Created file: %s\n", configPath)

	return nil
}

func createAgentFile(dir, filename string, agentNum int) error {
	content := fmt.Sprintf(`package main

import (
	"context"
	"fmt"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// Agent%dHandler represents the %d agent handler
type Agent%dHandler struct {
	llm agentflow.ModelProvider
}

// NewAgent%d creates a new Agent%d instance
func NewAgent%d(llmProvider agentflow.ModelProvider) *Agent%dHandler {
	return &Agent%dHandler{llm: llmProvider}
}

// Run implements the agentflow.AgentHandler interface
func (a *Agent%dHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	fmt.Printf("Agent%d processing event: %%s\n", event.GetID())
	
	// Get input from event payload or state
	var message interface{}
	eventData := event.GetData()
	if msg, ok := eventData["message"]; ok {
		message = msg
	} else if stateMessage, exists := state.Get("message"); exists {
		message = stateMessage
	} else {
		message = "No message provided"
	}
	
	fmt.Printf("Agent%d received message: %%v\n", message)
	
	// Create LLM prompt
	prompt := agentflow.Prompt{
		System: "You are Agent%d. Process the given input and provide a helpful response.",
		User:   fmt.Sprintf("Input: %%v", message),
	}
	
	// Call LLM
	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return agentflow.AgentResult{}, fmt.Errorf("Agent%d LLM call failed: %%w", err)
	}
	
	fmt.Printf("Agent%d LLM response: %%s\n", response.Content)
	
	// Create output state
	outputState := agentflow.NewState()
	outputState.Set("agent%d_response", response.Content)
	outputState.Set("processed_by", "agent%d")
	
	// Copy existing state data
	for _, key := range state.Keys() {
		if value, exists := state.Get(key); exists {
			outputState.Set(key, value)
		}
	}
	
	fmt.Printf("Agent%d completed processing\n")
	
	return agentflow.AgentResult{OutputState: outputState}, nil
}
`, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum, agentNum) // 17 arguments

	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", filename, err)
	}
	fmt.Printf("Created file: %s\n", filePath)
	return nil
}

func createResponsibleAIAgent(projectDir string) error {
	content := `package main

import (
	"context"
	"fmt"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// ResponsibleAIHandler handles AI safety and compliance checks
type ResponsibleAIHandler struct {
	llm agentflow.ModelProvider
}

// NewResponsibleAIHandler creates a new ResponsibleAIHandler
func NewResponsibleAIHandler(llmProvider agentflow.ModelProvider) *ResponsibleAIHandler {
	return &ResponsibleAIHandler{llm: llmProvider}
}

// Run implements the agentflow.AgentHandler interface
func (a *ResponsibleAIHandler) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	fmt.Printf("ResponsibleAI Agent processing event: %s\n", event.GetID())
	
	// Get content to check from event or state
	var content interface{}
	eventData := event.GetData()
	if data, ok := eventData["content"]; ok {
		content = data
	} else if stateContent, exists := state.Get("content"); exists {
		content = stateContent
	} else {
		content = "No content to check"
	}
	
	fmt.Printf("ResponsibleAI checking content: %%v\n", content)
	
	// Create LLM prompt for responsible AI checking
	prompt := agentflow.Prompt{
		System: "You are a responsible AI assistant. Check the given content for safety, bias, and compliance with ethical AI guidelines. Respond with 'SAFE' if content is appropriate, or 'UNSAFE: reason' if not.",
		User:   fmt.Sprintf("Content to check: %%v", content),
	}
	
	// Call LLM
	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return agentflow.AgentResult{}, fmt.Errorf("ResponsibleAI LLM call failed: %%w", err)
	}
	
	fmt.Printf("ResponsibleAI result: %%s\n", response.Content)
	
	// Create output state
	outputState := agentflow.NewState()
	outputState.Set("rai_check_result", response.Content)
	outputState.Set("processed_by", "responsible_ai")
	
	// Copy existing state data
	for _, key := range state.Keys() {
		if value, exists := state.Get(key); exists {
			outputState.Set(key, value)
		}
	}
	
	fmt.Printf("ResponsibleAI check completed\n")
	
	return agentflow.AgentResult{OutputState: outputState}, nil
}
`

	filePath := filepath.Join(projectDir, "responsible_ai.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create responsible_ai.go: %w", err)
	}
	fmt.Printf("Created file: %s\n", filePath)
	return nil
}

func createErrorHandlerAgent(projectDir string) error {
	content := `package main

import (
	"context"
	"fmt"

	agentflow "github.com/kunalkushwaha/agentflow/core"
)

// ErrorHandlerAgent handles errors and fallback logic
type ErrorHandlerAgent struct {
	llm agentflow.ModelProvider
}

// NewErrorHandler creates a new ErrorHandlerAgent
func NewErrorHandler(llmProvider agentflow.ModelProvider) *ErrorHandlerAgent {
	return &ErrorHandlerAgent{llm: llmProvider}
}

// Run implements the agentflow.AgentHandler interface
func (a *ErrorHandlerAgent) Run(ctx context.Context, event agentflow.Event, state agentflow.State) (agentflow.AgentResult, error) {
	fmt.Printf("Error Handler processing event: %s\n", event.GetID())
	
	// Get error information from event or state
	var errorInfo interface{}
	eventData := event.GetData()
	if err, ok := eventData["error"]; ok {
		errorInfo = err
	} else if stateError, exists := state.Get("error"); exists {
		errorInfo = stateError
	} else {
		errorInfo = "No error information available"
	}
	
	fmt.Printf("Error Handler analyzing: %%v\n", errorInfo)
	
	// Create LLM prompt for error handling
	prompt := agentflow.Prompt{
		System: "You are an error handling assistant. Analyze the given error and provide helpful suggestions for resolution.",
		User:   fmt.Sprintf("Error to analyze: %%v", errorInfo),
	}
	
	// Call LLM
	response, err := a.llm.Call(ctx, prompt)
	if err != nil {
		return agentflow.AgentResult{}, fmt.Errorf("Error Handler LLM call failed: %%w", err)
	}
	
	fmt.Printf("Error Handler analysis: %%s\n", response.Content)
	
	// Create output state
	outputState := agentflow.NewState()
	outputState.Set("error_analysis", response.Content)
	outputState.Set("processed_by", "error_handler")
	
	// Copy existing state data
	for _, key := range state.Keys() {
		if value, exists := state.Get(key); exists {
			outputState.Set(key, value)
		}
	}
	
	fmt.Printf("Error handling completed\n")
	
	return agentflow.AgentResult{OutputState: outputState}, nil
}
`

	filePath := filepath.Join(projectDir, "error_handler.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create error_handler.go: %w", err)
	}
	fmt.Printf("Created file: %s\n", filePath)
	return nil
}

func createReadmeContent(projectName string, numAgents int, responsibleAI bool, errorHandler bool, provider string) string {
	return fmt.Sprintf(`# %s

An AgentFlow multi-agent system with %d agents.

## Configuration

- **Provider**: %s
- **Agents**: %d
- **Responsible AI**: %t
- **Error Handler**: %t

## Setup

1. Install dependencies:
`+"```bash"+`
   go mod tidy
`+"```"+`

2. Configure your LLM provider:
   - For OpenAI: Set OPENAI_API_KEY environment variable
   - For Azure: Set AZURE_OPENAI_API_KEY, AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_DEPLOYMENT
   - For Ollama: Ensure Ollama is running locally

3. Run the application:
`+"```bash"+`
   go run .
`+"```"+`

## Project Structure

- `+"`main.go`"+` - Main application entry point
- `+"`agent*.go`"+` - Individual agent implementations
- `+"`workflows/`"+` - Workflow definitions
- `+"`agentflow.toml`"+` - Configuration file

## Usage

This project implements a multi-agent system using the AgentFlow framework. Each agent can process events and maintain state throughout the workflow execution.

Generated with AgentFlow CLI v0.1.0
`, projectName, numAgents, provider, numAgents, responsibleAI, errorHandler)
}

func createMainGoContent(projectName, provider string, numAgents int, responsibleAI bool, errorHandler bool) string {
	imports := `"context"
	"fmt"
	"log"`

	if provider != "mock" {
		imports += `
	"os"`
	}

	imports += `

	agentflow "github.com/kunalkushwaha/agentflow/core"`

	var providerSetup string
	switch provider {
	case "openai":
		providerSetup = `	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}
	llmProvider, err := agentflow.NewOpenAIAdapter(apiKey, "gpt-3.5-turbo", 1000, 0.7)
	if err != nil {
		log.Fatalf("Failed to create OpenAI adapter: %v", err)
	}`
	case "azure":
		providerSetup = `	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	chatDeployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT")
	if apiKey == "" || endpoint == "" || chatDeployment == "" {
		log.Fatal("AZURE_OPENAI_API_KEY, AZURE_OPENAI_ENDPOINT, and AZURE_OPENAI_CHAT_DEPLOYMENT environment variables are required")
	}
	llmProvider, err := agentflow.NewAzureOpenAIAdapter(agentflow.AzureOpenAIAdapterOptions{
		Endpoint:       endpoint,
		APIKey:         apiKey,
		ChatDeployment: chatDeployment,
	})
	if err != nil {
		log.Fatalf("Failed to create Azure OpenAI adapter: %v", err)
	}`
	case "ollama":
		providerSetup = `	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama2"
	}
	llmProvider, err := agentflow.NewOllamaAdapter(baseURL, model, 1000, 0.7)
	if err != nil {
		log.Fatalf("Failed to create Ollama adapter: %v", err)
	}`
	default: // mock
		providerSetup = `	// Create a simple mock provider for testing
	llmProvider := &MockProvider{}`
	}

	var agentRegistrations string
	if numAgents == 1 {
		agentRegistrations = `	agents["agent1"] = NewAgent1(llmProvider)`
	} else {
		for i := 1; i <= numAgents; i++ {
			agentRegistrations += fmt.Sprintf(`	agents["agent%d"] = NewAgent%d(llmProvider)
`, i, i)
		}
	}

	if responsibleAI {
		agentRegistrations += `	agents["responsible_ai"] = NewResponsibleAIHandler(llmProvider)
`
	}

	if errorHandler {
		agentRegistrations += `	agents["error_handler"] = NewErrorHandler(llmProvider)
`
	}

	var mockProviderCode string
	if provider == "mock" {
		mockProviderCode = `
// MockProvider is a simple mock implementation for testing
type MockProvider struct{}

func (m *MockProvider) Call(ctx context.Context, prompt agentflow.Prompt) (agentflow.Response, error) {
	return agentflow.Response{
		Content: fmt.Sprintf("Mock response to: %s", prompt.User),
		Usage: agentflow.UsageStats{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		FinishReason: "stop",
	}, nil
}

func (m *MockProvider) Stream(ctx context.Context, prompt agentflow.Prompt) (<-chan agentflow.Token, error) {
	ch := make(chan agentflow.Token, 1)
	go func() {
		defer close(ch)
		ch <- agentflow.Token{Content: "Mock response", Error: nil}
	}()
	return ch, nil
}

// Embeddings implements the agentflow.ModelProvider interface for MockProvider.
func (m *MockProvider) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	// Mock implementation for Embeddings
	// Return a slice of nil or empty float64 slices, one for each input text
	embeddings := make([][]float64, len(texts))
	for i := range texts {
		embeddings[i] = []float64{} // Or nil, depending on desired mock behavior
	}
	return embeddings, nil
}
`
	}
	return fmt.Sprintf(`package main

import (
	%s
)
%s
func main() {
	ctx := context.Background()
	
	// Initialize LLM provider
%s

	// Create agents map for runner configuration
	agents := make(map[string]agentflow.AgentHandler)
%s
	
	// Create AgentFlow runner with configuration
	runner := agentflow.NewRunnerWithConfig(agentflow.RunnerConfig{
		QueueSize: 10,
		Agents:    agents,
	})

	// Start the runner
	if err := runner.Start(ctx); err != nil {
		log.Fatalf("Failed to start runner: %%v", err)
	}
	defer runner.Stop()
	
	// Create initial event
	event := agentflow.NewEvent("agent1", agentflow.EventData{
		"message": "Welcome to the multi-agent system",
	}, map[string]string{
		"session_id": "demo-session",
		"route":      "agent1",
	})	// Emit the event to start processing
	if err := runner.Emit(event); err != nil {
		log.Fatalf("Error emitting event: %%v", err)
	}
	
	fmt.Println("Multi-agent system '%s' is running...")
	fmt.Println("Check the logs for agent processing details.")
}
`, imports, mockProviderCode, providerSetup, agentRegistrations, projectName)
}

func createWorkflowContent(numAgents int, responsibleAI bool, errorHandler bool) string {
	workflow := `# Multi-Agent Workflow

This workflow demonstrates the interaction between multiple agents in the system.

## Workflow Diagram

`

	// Create a simple mermaid diagram
	workflow += "```mermaid\n"
	workflow += "graph TD\n"
	workflow += "    Start([Start Event]) --> A1[Agent 1]\n"

	if numAgents > 1 {
		for i := 2; i <= numAgents; i++ {
			workflow += fmt.Sprintf("    A%d[Agent %d] --> A%d[Agent %d]\n", i-1, i-1, i, i)
		}
	}

	if responsibleAI {
		workflow += fmt.Sprintf("    A%d --> RAI[Responsible AI Check]\n", numAgents)
	}

	if errorHandler {
		workflow += "    RAI --> EH[Error Handler]\n"
		workflow += "    EH --> End([End])\n"
	} else if responsibleAI {
		workflow += "    RAI --> End([End])\n"
	} else {
		workflow += fmt.Sprintf("    A%d --> End([End])\n", numAgents)
	}

	workflow += "```\n\n"

	workflow += "## Agent Descriptions\n\n"
	for i := 1; i <= numAgents; i++ {
		workflow += fmt.Sprintf("- **Agent %d**: Processes input and generates responses\n", i)
	}

	if responsibleAI {
		workflow += "- **Responsible AI**: Checks content for safety and compliance\n"
	}

	if errorHandler {
		workflow += "- **Error Handler**: Manages errors and provides fallback logic\n"
	}

	return workflow
}

func createConfigContent(provider string) string {
	return fmt.Sprintf(`[agent_flow]
name = "Multi-Agent System"
version = "1.0.0"
provider = "%s"

[logging]
level = "info"
format = "json"

[runtime]
max_concurrent_agents = 10
timeout_seconds = 30

[providers.%s]
# Provider-specific configuration
# Add your configuration here based on the selected provider
`, provider, provider)
}
