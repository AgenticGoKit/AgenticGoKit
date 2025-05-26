package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to check if a path exists and is a directory or file
func checkExists(t *testing.T, path string, isDir bool) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("Expected %s to exist, but it does not", path)
		return
	}
	if err != nil {
		t.Errorf("Error stating %s: %v", path, err)
		return
	}
	if isDir && !info.IsDir() {
		t.Errorf("Expected %s to be a directory, but it is a file", path)
	}
	if !isDir && info.IsDir() {
		t.Errorf("Expected %s to be a file, but it is a directory", path)
	}
}

// Helper function to clean up created directories and files
func cleanup(t *testing.T, path string) {
	t.Helper()
	err := os.RemoveAll(path)
	if err != nil {
		t.Logf("Warning: failed to clean up %s: %v", path, err) // Log instead of Error for cleanup
	}
}

// Helper function to check if a file contains specific snippets
func checkFileContains(t *testing.T, path string, snippets []string) {
	t.Helper()
	checkExists(t, path, false) // Ensure file exists before reading

	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Failed to read file %s: %v", path, err)
		return
	}
	strContent := string(content)
	for _, snippet := range snippets {
		if !strings.Contains(strContent, snippet) {
			t.Errorf("Expected file %s to contain snippet:\n%s\n\nActual content:\n%s", path, snippet, strContent)
		}
	}
}

func TestCreateAgentProject_Basic(t *testing.T) {
	agentName := "test_basic_agent"
	defer cleanup(t, agentName)

	err := CreateAgentProject(agentName, 1, false, false)
	if err != nil {
		t.Fatalf("CreateAgentProject failed: %v", err)
	}

	// Verify main agent directory
	checkExists(t, agentName, true)

	// Verify main.go
	mainGoPath := filepath.Join(agentName, "main.go")
	checkExists(t, mainGoPath, false)
	checkFileContains(t, mainGoPath, []string{fmt.Sprintf("Hello from %s!", agentName)})

	// Verify agent.go (with Agent1 struct)
	agentGoPath := filepath.Join(agentName, "agent.go")
	checkExists(t, agentGoPath, false)
	checkFileContains(t, agentGoPath, []string{"package main", "type Agent1 struct", "func NewAgent1("})

	// Verify responsible_ai and error_handler directories are NOT created
	raiDir := filepath.Join(agentName, "responsible_ai")
	if _, err := os.Stat(raiDir); !os.IsNotExist(err) {
		t.Errorf("Expected directory %s to not exist, but it does", raiDir)
	}

	ehDir := filepath.Join(agentName, "error_handler")
	if _, err := os.Stat(ehDir); !os.IsNotExist(err) {
		t.Errorf("Expected directory %s to not exist, but it does", ehDir)
	}
}

// Placeholder for TestCreateAgentProject_MultipleAgents
func TestCreateAgentProject_MultipleAgents(t *testing.T) {
	agentName := "test_multi_agent"
	defer cleanup(t, agentName)

	err := CreateAgentProject(agentName, 3, false, false)
	if err != nil {
		t.Fatalf("CreateAgentProject for multiple agents failed: %v", err)
	}

	checkExists(t, agentName, true)
	mainGoPath := filepath.Join(agentName, "main.go")
	checkExists(t, mainGoPath, false)
	checkFileContains(t, mainGoPath, []string{fmt.Sprintf("Hello from %s!", agentName)})

	for i := 1; i <= 3; i++ {
		agentSubDir := filepath.Join(agentName, fmt.Sprintf("agent%d", i))
		checkExists(t, agentSubDir, true)
		agentGoPath := filepath.Join(agentSubDir, "agent.go")
		checkExists(t, agentGoPath, false)
		checkFileContains(t, agentGoPath, []string{
			"package main",
			fmt.Sprintf("type Agent%d struct", i),
			fmt.Sprintf("func NewAgent%d(", i),
			fmt.Sprintf("Agent%d (Name: %%s) received event: %%v", i), // Check for the Printf string
		})
	}
}

// Placeholder for TestCreateAgentProject_WithResponsibleAI
func TestCreateAgentProject_WithResponsibleAI(t *testing.T) {
	agentName := "test_rai_agent"
	defer cleanup(t, agentName)

	err := CreateAgentProject(agentName, 1, true, false)
	if err != nil {
		t.Fatalf("CreateAgentProject with Responsible AI failed: %v", err)
	}

	checkExists(t, agentName, true)
	raiDir := filepath.Join(agentName, "responsible_ai")
	checkExists(t, raiDir, true)
	raiAgentGoPath := filepath.Join(raiDir, "agent.go")
	checkExists(t, raiAgentGoPath, false)
	checkFileContains(t, raiAgentGoPath, []string{"package responsible_ai", "type ResponsibleAIAgent struct"})
}

// Placeholder for TestCreateAgentProject_WithErrorHandler
func TestCreateAgentProject_WithErrorHandler(t *testing.T) {
	agentName := "test_eh_agent"
	defer cleanup(t, agentName)

	err := CreateAgentProject(agentName, 1, false, true)
	if err != nil {
		t.Fatalf("CreateAgentProject with Error Handler failed: %v", err)
	}

	checkExists(t, agentName, true)
	ehDir := filepath.Join(agentName, "error_handler")
	checkExists(t, ehDir, true)
	ehAgentGoPath := filepath.Join(ehDir, "agent.go")
	checkExists(t, ehAgentGoPath, false)
	checkFileContains(t, ehAgentGoPath, []string{"package error_handler", "type ErrorHandlerAgent struct"})
}

// Placeholder for TestCreateAgentProject_WithAllOptions
func TestCreateAgentProject_WithAllOptions(t *testing.T) {
	agentName := "test_all_options_agent"
	defer cleanup(t, agentName)

	err := CreateAgentProject(agentName, 2, true, true)
	if err != nil {
		t.Fatalf("CreateAgentProject with all options failed: %v", err)
	}

	// Project Root and main.go
	checkExists(t, agentName, true)
	mainGoPath := filepath.Join(agentName, "main.go")
	checkExists(t, mainGoPath, false)
	checkFileContains(t, mainGoPath, []string{fmt.Sprintf("Hello from %s!", agentName)})

	// Numbered Agents (agent1, agent2)
	for i := 1; i <= 2; i++ {
		agentSubDir := filepath.Join(agentName, fmt.Sprintf("agent%d", i))
		checkExists(t, agentSubDir, true)
		agentGoPath := filepath.Join(agentSubDir, "agent.go")
		checkExists(t, agentGoPath, false)
		checkFileContains(t, agentGoPath, []string{
			fmt.Sprintf("type Agent%d struct", i),
			fmt.Sprintf("func NewAgent%d(", i),
		})
	}

	// Responsible AI Agent
	raiDir := filepath.Join(agentName, "responsible_ai")
	checkExists(t, raiDir, true)
	raiAgentGoPath := filepath.Join(raiDir, "agent.go")
	checkExists(t, raiAgentGoPath, false)
	checkFileContains(t, raiAgentGoPath, []string{"package responsible_ai", "type ResponsibleAIAgent struct"})

	// Error Handler Agent
	ehDir := filepath.Join(agentName, "error_handler")
	checkExists(t, ehDir, true)
	ehAgentGoPath := filepath.Join(ehDir, "agent.go")
	checkExists(t, ehAgentGoPath, false)
	checkFileContains(t, ehAgentGoPath, []string{"package error_handler", "type ErrorHandlerAgent struct"})
}

// Optional: Test for error handling when directory already exists
func TestCreateAgentProject_ErrorOnExistingDir(t *testing.T) {
	agentName := "test_existing_dir_agent"
	// Create a directory with the same name beforehand
	err := os.Mkdir(agentName, 0755)
	if err != nil {
		t.Fatalf("Setup: Failed to create pre-existing directory %s: %v", agentName, err)
	}
	// Create a dummy file inside to make it non-empty (optional, os.Mkdir might fail on non-empty)
	// For simplicity, os.Mkdir in CreateAgentProject will fail if agentName dir exists.
	defer cleanup(t, agentName) // Ensure cleanup even if parts of the test fail

	err = CreateAgentProject(agentName, 1, false, false)
	if err == nil {
		t.Errorf("Expected CreateAgentProject to fail when directory '%s' already exists, but it succeeded", agentName)
	}
}
