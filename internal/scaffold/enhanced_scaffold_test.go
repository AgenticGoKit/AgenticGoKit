package scaffold

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnhancedScaffoldGeneration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "enhanced_scaffold_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test enhanced scaffold generation with error handling
	projectName := "enhanced_test_project"
	config := ProjectConfig{
		Name:          projectName,
		NumAgents:     2,
		Provider:      "openai",
		ResponsibleAI: true,
		ErrorHandler:  true,
	}
	err = CreateAgentProject(config)
	if err != nil {
		t.Fatalf("Failed to create enhanced project: %v", err)
	}

	projectDir := filepath.Join(tempDir, projectName)

	// Verify enhanced configuration file includes error routing
	configPath := filepath.Join(projectDir, "agentflow.toml")
	configContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	configStr := string(configContent)

	// Verify error routing configuration is present
	expectedConfigSections := []string{
		"[error_routing]",
		"enabled = true",
		"auto_detect_handlers = true",
		"[error_routing.circuit_breaker]",
		"failure_threshold = 3",
		"success_threshold = 2",
		"timeout = \"30s\"",
		"reset_timeout = \"60s\"",
		"half_open_max_calls = 2",
		"[error_routing.retry]",
		"max_retries = 2",
		"base_delay = \"1s\"",
		"max_delay = \"10s\"",
		"backoff_factor = 2.0",
		"enable_jitter = true",
		"[error_routing.handlers]",
		"[error_routing.handlers.category]",
		"validation = \"validation_error_handler\"",
		"timeout = \"timeout_error_handler\"",
		"critical = \"critical_error_handler\"",
		"[error_routing.handlers.severity]",
		"high = \"critical_error_handler\"",
		"medium = \"timeout_error_handler\"",
		"low = \"validation_error_handler\"",
		"default_handler = \"error_handler\"",
	}

	for _, expected := range expectedConfigSections {
		if !strings.Contains(configStr, expected) {
			t.Errorf("Config file missing expected section: %s", expected)
		}
	}

	// Verify enhanced error handlers exist
	enhancedHandlers := []string{
		"validation_error_handler.go",
		"timeout_error_handler.go",
		"critical_error_handler.go",
	}

	for _, handler := range enhancedHandlers {
		handlerPath := filepath.Join(projectDir, handler)
		if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
			t.Errorf("Enhanced error handler file missing: %s", handler)
		}
	}

	// Verify validation error handler has circuit breaker integration
	validationPath := filepath.Join(projectDir, "validation_error_handler.go")
	validationContent, err := ioutil.ReadFile(validationPath)
	if err != nil {
		t.Fatalf("Failed to read validation error handler: %v", err)
	}
	validationStr := string(validationContent)
	expectedValidationFeatures := []string{
		"ValidationErrorHandler struct",
		"llm agentflow.ModelProvider",
		"maxRetries int",
		"retryDelay time.Duration",
		"NewValidationErrorHandler",
		"simple retry logic",
		"validation_fix_suggestions",
		"recovery_action",
		"fallback_used",
		"retry_with_corrections",
	}

	for _, expected := range expectedValidationFeatures {
		if !strings.Contains(validationStr, expected) {
			t.Errorf("Validation error handler missing expected feature: %s", expected)
		}
	}

	// Verify timeout error handler has circuit breaker integration
	timeoutPath := filepath.Join(projectDir, "timeout_error_handler.go")
	timeoutContent, err := ioutil.ReadFile(timeoutPath)
	if err != nil {
		t.Fatalf("Failed to read timeout error handler: %v", err)
	}
	timeoutStr := string(timeoutContent)
	expectedTimeoutFeatures := []string{
		"TimeoutErrorHandler struct",
		"llm agentflow.ModelProvider",
		"maxRetries int",
		"baseDelay time.Duration",
		"NewTimeoutErrorHandler",
		"exponential backoff",
		"timeout_optimization_suggestions",
		"retry_delay",
		"timeout_strategy",
	}

	for _, expected := range expectedTimeoutFeatures {
		if !strings.Contains(timeoutStr, expected) {
			t.Errorf("Timeout error handler missing expected feature: %s", expected)
		}
	}

	// Verify critical error handler has circuit breaker integration
	criticalPath := filepath.Join(projectDir, "critical_error_handler.go")
	criticalContent, err := ioutil.ReadFile(criticalPath)
	if err != nil {
		t.Fatalf("Failed to read critical error handler: %v", err)
	}
	criticalStr := string(criticalContent)
	expectedCriticalFeatures := []string{
		"CriticalErrorHandler struct",
		"llm agentflow.ModelProvider",
		"llmTimeout time.Duration",
		"NewCriticalErrorHandler",
		"immediate fallback",
		"emergency_timestamp",
		"error_analysis",
		"recommended_action",
		"CRITICAL ERROR",
	}

	for _, expected := range expectedCriticalFeatures {
		if !strings.Contains(criticalStr, expected) {
			t.Errorf("Critical error handler missing expected feature: %s", expected)
		}
	}

	// Verify main.go includes all specialized error handler registrations
	mainPath := filepath.Join(projectDir, "main.go")
	mainContent, err := ioutil.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}
	mainStr := string(mainContent)
	expectedMainFeatures := []string{
		"NewValidationErrorHandler",
		"NewTimeoutErrorHandler",
		"NewCriticalErrorHandler",
		"\"validation-error-handler\": NewValidationErrorHandler",
		"\"timeout-error-handler\": NewTimeoutErrorHandler",
		"\"critical-error-handler\": NewCriticalErrorHandler",
		"core.NewRunnerFromWorkingDir",
		"agents := map[string]core.AgentHandler",
	}

	for _, expected := range expectedMainFeatures {
		if !strings.Contains(mainStr, expected) {
			t.Errorf("main.go missing expected feature: %s", expected)
		}
	}

	t.Logf("Enhanced scaffold generation test passed successfully")
}

func TestScaffoldWithoutErrorHandling(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "scaffold_no_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test scaffold generation without error handling
	projectName := "no_error_test_project"
	config := ProjectConfig{
		Name:          projectName,
		NumAgents:     1,
		Provider:      "openai",
		ResponsibleAI: false,
		ErrorHandler:  false,
	}
	err = CreateAgentProject(config)
	if err != nil {
		t.Fatalf("Failed to create project without error handling: %v", err)
	}

	projectDir := filepath.Join(tempDir, projectName)

	// Verify configuration file does NOT include error routing
	configPath := filepath.Join(projectDir, "agentflow.toml")
	configContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	configStr := string(configContent)

	// Verify error routing configuration is NOT present
	unexpectedConfigSections := []string{
		"[error_routing]",
		"circuit_breaker",
		"retry",
	}

	for _, unexpected := range unexpectedConfigSections {
		if strings.Contains(configStr, unexpected) {
			t.Errorf("Config file should not contain error routing section: %s", unexpected)
		}
	}

	// Verify specialized error handlers do NOT exist
	enhancedHandlers := []string{
		"validation_error_handler.go",
		"timeout_error_handler.go",
		"critical_error_handler.go",
	}

	for _, handler := range enhancedHandlers {
		handlerPath := filepath.Join(projectDir, handler)
		if _, err := os.Stat(handlerPath); !os.IsNotExist(err) {
			t.Errorf("Enhanced error handler file should not exist: %s", handler)
		}
	}

	t.Logf("Scaffold without error handling test passed successfully")
}

func TestEnhancedScaffoldDifferentProviders(t *testing.T) {
	providers := []string{"openai", "azure", "ollama"}

	for _, provider := range providers {
		t.Run("provider_"+provider, func(t *testing.T) {
			// Create a temporary directory for testing
			tempDir, err := ioutil.TempDir("", "enhanced_scaffold_"+provider+"_test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Change to temp directory for the test
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}
			defer os.Chdir(originalDir)

			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}

			// Test enhanced scaffold generation with different providers
			projectName := "enhanced_" + provider + "_project"
			config := ProjectConfig{
				Name:          projectName,
				NumAgents:     1,
				Provider:      provider,
				ResponsibleAI: true,
				ErrorHandler:  true,
			}
			err = CreateAgentProject(config)
			if err != nil {
				t.Fatalf("Failed to create enhanced project with %s provider: %v", provider, err)
			}

			projectDir := filepath.Join(tempDir, projectName)

			// Verify configuration file includes both provider config and error routing
			configPath := filepath.Join(projectDir, "agentflow.toml")
			configContent, err := ioutil.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read config file: %v", err)
			}

			configStr := string(configContent)

			// Verify provider-specific configuration
			expectedProviderSection := "[providers." + provider + "]"
			if !strings.Contains(configStr, expectedProviderSection) {
				t.Errorf("Config file missing provider section: %s", expectedProviderSection)
			}

			// Verify error routing configuration is present
			if !strings.Contains(configStr, "[error_routing]") {
				t.Errorf("Config file missing error routing configuration")
			}

			t.Logf("Enhanced scaffold generation test with %s provider passed", provider)
		})
	}
}
// TestCreateProjectDirectories tests the directory structure creation
func TestCreateProjectDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "project_directories_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test configuration
	projectName := "test_project_structure"
	config := ProjectConfig{
		Name:      projectName,
		NumAgents: 2,
		Provider:  "openai",
	}

	// Create the main project directory first
	if err := os.Mkdir(config.Name, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Test directory creation
	err = createProjectDirectories(config)
	if err != nil {
		t.Fatalf("Failed to create project directories: %v", err)
	}

	projectDir := filepath.Join(tempDir, projectName)

	// Verify agents directory was created
	agentsDir := filepath.Join(projectDir, "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Errorf("Agents directory was not created: %s", agentsDir)
	}

	// Verify internal directory was created
	internalDir := filepath.Join(projectDir, "internal")
	if _, err := os.Stat(internalDir); os.IsNotExist(err) {
		t.Errorf("Internal directory was not created: %s", internalDir)
	}

	// Verify internal/config directory was created
	configDir := filepath.Join(internalDir, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Internal/config directory was not created: %s", configDir)
	}

	// Verify internal/handlers directory was created
	handlersDir := filepath.Join(internalDir, "handlers")
	if _, err := os.Stat(handlersDir); os.IsNotExist(err) {
		t.Errorf("Internal/handlers directory was not created: %s", handlersDir)
	}

	// Verify docs directory was created
	docsDir := filepath.Join(projectDir, "docs")
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		t.Errorf("Docs directory was not created: %s", docsDir)
	}
}

// TestCreateAgentsDirectory tests the agents directory creation specifically
func TestCreateAgentsDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "agents_directory_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test configuration
	projectName := "test_agents_dir"
	config := ProjectConfig{
		Name:      projectName,
		NumAgents: 1,
		Provider:  "openai",
	}

	// Create the main project directory first
	if err := os.Mkdir(config.Name, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Test agents directory creation
	err = createAgentsDirectory(config)
	if err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Verify agents directory was created
	agentsDir := filepath.Join(tempDir, projectName, "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Errorf("Agents directory was not created: %s", agentsDir)
	}

	// Verify directory has correct permissions
	info, err := os.Stat(agentsDir)
	if err != nil {
		t.Fatalf("Failed to get directory info: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Agents path is not a directory: %s", agentsDir)
	}
}

// TestCreateInternalDirectory tests the internal directory creation specifically
func TestCreateInternalDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "internal_directory_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test configuration
	projectName := "test_internal_dir"
	config := ProjectConfig{
		Name:      projectName,
		NumAgents: 1,
		Provider:  "openai",
	}

	// Create the main project directory first
	if err := os.Mkdir(config.Name, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Test internal directory creation
	err = createInternalDirectory(config)
	if err != nil {
		t.Fatalf("Failed to create internal directory: %v", err)
	}

	projectDir := filepath.Join(tempDir, projectName)

	// Verify internal directory was created
	internalDir := filepath.Join(projectDir, "internal")
	if _, err := os.Stat(internalDir); os.IsNotExist(err) {
		t.Errorf("Internal directory was not created: %s", internalDir)
	}

	// Verify internal/config subdirectory was created
	configDir := filepath.Join(internalDir, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Internal/config directory was not created: %s", configDir)
	}

	// Verify internal/handlers subdirectory was created
	handlersDir := filepath.Join(internalDir, "handlers")
	if _, err := os.Stat(handlersDir); os.IsNotExist(err) {
		t.Errorf("Internal/handlers directory was not created: %s", handlersDir)
	}
}

// TestCreateDocsDirectory tests the docs directory creation specifically
func TestCreateDocsDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "docs_directory_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test configuration
	projectName := "test_docs_dir"
	config := ProjectConfig{
		Name:      projectName,
		NumAgents: 1,
		Provider:  "openai",
	}

	// Create the main project directory first
	if err := os.Mkdir(config.Name, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Test docs directory creation
	err = createDocsDirectory(config)
	if err != nil {
		t.Fatalf("Failed to create docs directory: %v", err)
	}

	// Verify docs directory was created
	docsDir := filepath.Join(tempDir, projectName, "docs")
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		t.Errorf("Docs directory was not created: %s", docsDir)
	}

	// Verify directory has correct permissions
	info, err := os.Stat(docsDir)
	if err != nil {
		t.Fatalf("Failed to get directory info: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Docs path is not a directory: %s", docsDir)
	}
}