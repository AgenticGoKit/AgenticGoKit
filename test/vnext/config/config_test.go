package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core/vnext"
)

// TestConfigStructure tests the Config structure
func TestConfigStructure(t *testing.T) {
	config := &vnext.Config{
		Name:         "test-agent",
		SystemPrompt: "You are a helpful assistant",
		Timeout:      30 * time.Second,
		DebugMode:    true,
		LLM: vnext.LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
	}

	if config.Name != "test-agent" {
		t.Errorf("Name = %s, want test-agent", config.Name)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", config.Timeout)
	}

	if !config.DebugMode {
		t.Error("DebugMode should be true")
	}
}

// TestLLMConfig tests LLM configuration
func TestLLMConfig(t *testing.T) {
	tests := []struct {
		name   string
		config vnext.LLMConfig
	}{
		{
			name: "openai_config",
			config: vnext.LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   2000,
			},
		},
		{
			name: "ollama_config",
			config: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "llama2",
				Temperature: 0.5,
				MaxTokens:   1000,
				BaseURL:     "http://localhost:11434",
			},
		},
		{
			name: "azure_config",
			config: vnext.LLMConfig{
				Provider:    "azure",
				Model:       "gpt-35-turbo",
				Temperature: 0.8,
				MaxTokens:   1500,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Provider == "" {
				t.Error("Provider should not be empty")
			}
			if tt.config.Model == "" {
				t.Error("Model should not be empty")
			}
		})
	}
}

// TestMemoryConfig tests memory configuration
func TestMemoryConfig(t *testing.T) {
	memConfig := &vnext.MemoryConfig{
		Provider:   "memory",
		Connection: "local",
		RAG: &vnext.RAGConfig{
			MaxTokens:       1000,
			PersonalWeight:  0.7,
			KnowledgeWeight: 0.3,
			HistoryLimit:    10,
		},
		Options: map[string]string{
			"key": "value",
		},
	}

	if memConfig.Provider != "memory" {
		t.Errorf("Provider = %s, want memory", memConfig.Provider)
	}

	if memConfig.RAG == nil {
		t.Fatal("RAG config is nil")
	}

	if memConfig.RAG.MaxTokens != 1000 {
		t.Errorf("MaxTokens = %d, want 1000", memConfig.RAG.MaxTokens)
	}
}

// TestRAGConfig tests RAG configuration
func TestRAGConfig(t *testing.T) {
	ragConfig := &vnext.RAGConfig{
		MaxTokens:       2000,
		PersonalWeight:  0.6,
		KnowledgeWeight: 0.4,
		HistoryLimit:    20,
	}

	if ragConfig.MaxTokens != 2000 {
		t.Errorf("MaxTokens = %d, want 2000", ragConfig.MaxTokens)
	}

	if ragConfig.PersonalWeight != 0.6 {
		t.Errorf("PersonalWeight = %f, want 0.6", ragConfig.PersonalWeight)
	}

	if ragConfig.KnowledgeWeight != 0.4 {
		t.Errorf("KnowledgeWeight = %f, want 0.4", ragConfig.KnowledgeWeight)
	}

	if ragConfig.HistoryLimit != 20 {
		t.Errorf("HistoryLimit = %d, want 20", ragConfig.HistoryLimit)
	}
}

// TestToolsConfig tests tools configuration
func TestToolsConfig(t *testing.T) {
	toolsConfig := &vnext.ToolsConfig{
		Enabled:       true,
		MaxRetries:    3,
		Timeout:       30 * time.Second,
		RateLimit:     10,
		MaxConcurrent: 5,
	}

	if !toolsConfig.Enabled {
		t.Error("Tools should be enabled")
	}

	if toolsConfig.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", toolsConfig.MaxRetries)
	}

	if toolsConfig.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", toolsConfig.Timeout)
	}
}

// TestMCPConfig tests MCP configuration
func TestMCPConfig(t *testing.T) {
	mcpConfig := &vnext.MCPConfig{
		Enabled:           true,
		Discovery:         true,
		ConnectionTimeout: 10 * time.Second,
		MaxRetries:        3,
		RetryDelay:        1 * time.Second,
		DiscoveryTimeout:  5 * time.Second,
		Servers: []vnext.MCPServer{
			{
				Name:    "test-server",
				Type:    "tcp",
				Address: "localhost",
				Port:    8080,
				Enabled: true,
			},
		},
	}

	if !mcpConfig.Enabled {
		t.Error("MCP should be enabled")
	}

	if len(mcpConfig.Servers) != 1 {
		t.Errorf("Servers length = %d, want 1", len(mcpConfig.Servers))
	}

	if mcpConfig.Servers[0].Name != "test-server" {
		t.Errorf("Server name = %s, want test-server", mcpConfig.Servers[0].Name)
	}
}

// TestMCPServer tests MCP server configuration
func TestMCPServer(t *testing.T) {
	tests := []struct {
		name   string
		server vnext.MCPServer
	}{
		{
			name: "tcp_server",
			server: vnext.MCPServer{
				Name:    "tcp-server",
				Type:    "tcp",
				Address: "localhost",
				Port:    8080,
				Enabled: true,
			},
		},
		{
			name: "stdio_server",
			server: vnext.MCPServer{
				Name:    "stdio-server",
				Type:    "stdio",
				Command: "/path/to/command",
				Enabled: true,
			},
		},
		{
			name: "websocket_server",
			server: vnext.MCPServer{
				Name:    "ws-server",
				Type:    "websocket",
				Address: "ws://localhost:9090",
				Enabled: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.server.Name == "" {
				t.Error("Server name should not be empty")
			}
			if tt.server.Type == "" {
				t.Error("Server type should not be empty")
			}
		})
	}
}

// TestStreamingConfig tests streaming configuration
func TestStreamingConfig(t *testing.T) {
	streamingConfig := &vnext.StreamingConfig{
		Enabled:          true,
		BufferSize:       100,
		FlushInterval:    100,
		Timeout:          30000,
		IncludeThoughts:  true,
		IncludeToolCalls: true,
	}

	if !streamingConfig.Enabled {
		t.Error("Streaming should be enabled")
	}

	if streamingConfig.BufferSize != 100 {
		t.Errorf("BufferSize = %d, want 100", streamingConfig.BufferSize)
	}

	if !streamingConfig.IncludeThoughts {
		t.Error("IncludeThoughts should be true")
	}
}

// TestTracingConfig tests tracing configuration
func TestTracingConfig(t *testing.T) {
	tracingConfig := &vnext.TracingConfig{
		Enabled: true,
		Level:   "debug",
	}

	if !tracingConfig.Enabled {
		t.Error("Tracing should be enabled")
	}

	if tracingConfig.Level != "debug" {
		t.Errorf("Level = %s, want debug", tracingConfig.Level)
	}
}

// TestWorkflowConfigStructure tests workflow configuration structure
func TestWorkflowConfigStructure(t *testing.T) {
	workflowConfig := &vnext.WorkflowConfig{
		Mode:          vnext.Sequential,
		Agents:        []string{"agent1", "agent2"},
		Timeout:       60 * time.Second,
		MaxIterations: 5,
	}

	if workflowConfig.Mode != vnext.Sequential {
		t.Errorf("Mode = %s, want %s", workflowConfig.Mode, vnext.Sequential)
	}

	if len(workflowConfig.Agents) != 2 {
		t.Errorf("Agents length = %d, want 2", len(workflowConfig.Agents))
	}
}

// TestDefaultConfig tests default configuration creation
func TestDefaultConfig(t *testing.T) {
	config := vnext.DefaultConfig("test-agent")

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if config.Name != "test-agent" {
		t.Errorf("Name = %s, want test-agent", config.Name)
	}

	if config.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}

	if config.LLM.Provider == "" {
		t.Error("LLM provider should be set")
	}
}

// TestNewConfig tests configuration creation with options
func TestNewConfig(t *testing.T) {
	config := vnext.NewConfig("test-agent")

	if config == nil {
		t.Fatal("NewConfig returned nil")
	}

	if config.Name != "test-agent" {
		t.Errorf("Name = %s, want test-agent", config.Name)
	}
}

// TestValidateConfig tests configuration validation
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *vnext.Config
		wantErr bool
	}{
		{
			name: "valid_config",
			config: &vnext.Config{
				Name:    "valid-agent",
				Timeout: 30 * time.Second,
				LLM: vnext.LLMConfig{
					Provider: "openai",
					Model:    "gpt-4",
				},
			},
			wantErr: false,
		},
		{
			name: "missing_name",
			config: &vnext.Config{
				Name:    "",
				Timeout: 30 * time.Second,
				LLM: vnext.LLMConfig{
					Provider: "openai",
					Model:    "gpt-4",
				},
			},
			wantErr: true,
		},
		{
			name: "zero_timeout",
			config: &vnext.Config{
				Name:    "test-agent",
				Timeout: 0,
				LLM: vnext.LLMConfig{
					Provider: "openai",
					Model:    "gpt-4",
				},
			},
			wantErr: false, // Zero timeout is allowed (will use default)
		},
		{
			name: "missing_llm_provider",
			config: &vnext.Config{
				Name:    "test-agent",
				Timeout: 30 * time.Second,
				LLM: vnext.LLMConfig{
					Provider: "",
					Model:    "gpt-4",
				},
			},
			wantErr: true,
		},
		{
			name: "missing_llm_model",
			config: &vnext.Config{
				Name:    "test-agent",
				Timeout: 30 * time.Second,
				LLM: vnext.LLMConfig{
					Provider: "openai",
					Model:    "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vnext.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCloneConfig tests configuration cloning
func TestCloneConfig(t *testing.T) {
	original := &vnext.Config{
		Name:         "original-agent",
		SystemPrompt: "Original prompt",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
		},
	}

	cloned := vnext.CloneConfig(original)

	if cloned == nil {
		t.Fatal("CloneConfig returned nil")
	}

	if cloned.Name != original.Name {
		t.Errorf("Cloned name = %s, want %s", cloned.Name, original.Name)
	}

	if cloned.Timeout != original.Timeout {
		t.Errorf("Cloned timeout = %v, want %v", cloned.Timeout, original.Timeout)
	}

	// Modify cloned config and verify original is unchanged
	cloned.Name = "modified-agent"
	if original.Name == "modified-agent" {
		t.Error("Modifying cloned config affected original")
	}
}

// TestMergeConfigs tests configuration merging
func TestMergeConfigs(t *testing.T) {
	config1 := &vnext.Config{
		Name:    "agent1",
		Timeout: 30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "openai",
			Model:    "gpt-4",
		},
	}

	config2 := &vnext.Config{
		SystemPrompt: "Custom prompt",
		Timeout:      60 * time.Second,
	}

	merged := vnext.MergeConfigs(config1, config2)

	if merged == nil {
		t.Fatal("MergeConfigs returned nil")
	}

	// Should have values from both configs
	if merged.Name != "agent1" {
		t.Errorf("Name = %s, want agent1", merged.Name)
	}

	if merged.SystemPrompt != "Custom prompt" {
		t.Errorf("SystemPrompt = %s, want 'Custom prompt'", merged.SystemPrompt)
	}

	// config2's timeout should override config1
	if merged.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", merged.Timeout)
	}
} // TestLoadConfigFromTOML tests loading configuration from TOML file
func TestLoadConfigFromTOML(t *testing.T) {
	// Create a temporary TOML file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")

	tomlContent := `
name = "test-agent"
system_prompt = "You are a test assistant"
timeout = "30s"
debug_mode = true

[llm]
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 2000

[streaming]
enabled = true
buffer_size = 100
flush_interval_ms = 100
`

	err := os.WriteFile(configFile, []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	config, err := vnext.LoadConfigFromTOML(configFile)
	if err != nil {
		t.Fatalf("LoadConfigFromTOML() error = %v", err)
	}

	if config.Name != "test-agent" {
		t.Errorf("Name = %s, want test-agent", config.Name)
	}

	if config.LLM.Provider != "openai" {
		t.Errorf("LLM Provider = %s, want openai", config.LLM.Provider)
	}

	if config.LLM.Model != "gpt-4" {
		t.Errorf("LLM Model = %s, want gpt-4", config.LLM.Model)
	}

	if config.Streaming == nil {
		t.Fatal("Streaming config is nil")
	}

	if !config.Streaming.Enabled {
		t.Error("Streaming should be enabled")
	}
}

// TestLoadConfigFromTOMLWithMemory tests loading config with memory settings
func TestLoadConfigFromTOMLWithMemory(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config_with_memory.toml")

	tomlContent := `
name = "memory-agent"
timeout = "30s"

[llm]
provider = "openai"
model = "gpt-4"

[memory]
provider = "memory"
connection = "local"

[memory.rag]
max_tokens = 1000
personal_weight = 0.7
knowledge_weight = 0.3
history_limit = 10
`

	err := os.WriteFile(configFile, []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	config, err := vnext.LoadConfigFromTOML(configFile)
	if err != nil {
		t.Fatalf("LoadConfigFromTOML() error = %v", err)
	}

	if config.Memory == nil {
		t.Fatal("Memory config is nil")
	}

	if config.Memory.Provider != "memory" {
		t.Errorf("Memory Provider = %s, want memory", config.Memory.Provider)
	}

	if config.Memory.RAG == nil {
		t.Fatal("RAG config is nil")
	}

	if config.Memory.RAG.MaxTokens != 1000 {
		t.Errorf("RAG MaxTokens = %d, want 1000", config.Memory.RAG.MaxTokens)
	}
}

// TestLoadConfigFromTOMLWithWorkflow tests loading config with workflow settings
func TestLoadConfigFromTOMLWithWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config_with_workflow.toml")

	tomlContent := `
name = "workflow-agent"
timeout = "30s"

[llm]
provider = "openai"
model = "gpt-4"

[workflow]
mode = "sequential"
agents = ["agent1", "agent2"]
timeout = "60s"
max_iterations = 5
`

	err := os.WriteFile(configFile, []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	config, err := vnext.LoadConfigFromTOML(configFile)
	if err != nil {
		t.Fatalf("LoadConfigFromTOML() error = %v", err)
	}

	if config.Workflow == nil {
		t.Fatal("Workflow config is nil")
	}

	if config.Workflow.Mode != vnext.Sequential {
		t.Errorf("Workflow Mode = %s, want %s", config.Workflow.Mode, vnext.Sequential)
	}

	if len(config.Workflow.Agents) != 2 {
		t.Errorf("Workflow Agents length = %d, want 2", len(config.Workflow.Agents))
	}
}

// TestLoadConfigFromNonexistentFile tests error handling for missing file
func TestLoadConfigFromNonexistentFile(t *testing.T) {
	_, err := vnext.LoadConfigFromTOML("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("Expected error when loading from nonexistent file")
	}
}

// TestLoadConfigFromInvalidTOML tests error handling for invalid TOML
func TestLoadConfigFromInvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.toml")

	invalidContent := `
name = "test-agent"
this is not valid TOML syntax
[llm
`

	err := os.WriteFile(configFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	_, err = vnext.LoadConfigFromTOML(configFile)
	if err == nil {
		t.Error("Expected error when loading invalid TOML")
	}
}

// TestConfigResolver tests configuration resolution with environment variables
func TestConfigResolver(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_API_KEY", "test-key-123")
	defer os.Unsetenv("TEST_API_KEY")

	config := &vnext.Config{
		Name:    "test-agent",
		Timeout: 30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "openai",
			Model:    "gpt-4",
			APIKey:   "${TEST_API_KEY}",
		},
	}

	resolver := vnext.NewConfigResolver(config)
	resolved := resolver.ResolveConfig()

	if resolved.LLM.APIKey != "test-key-123" {
		t.Errorf("API key = %s, want test-key-123", resolved.LLM.APIKey)
	}
}

// TestConfigWithAllFeatures tests configuration with all features enabled
func TestConfigWithAllFeatures(t *testing.T) {
	config := &vnext.Config{
		Name:         "full-featured-agent",
		SystemPrompt: "Complete assistant",
		Timeout:      60 * time.Second,
		DebugMode:    true,
		LLM: vnext.LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
		Memory: &vnext.MemoryConfig{
			Provider: "memory",
			RAG: &vnext.RAGConfig{
				MaxTokens:       1000,
				PersonalWeight:  0.7,
				KnowledgeWeight: 0.3,
			},
		},
		Tools: &vnext.ToolsConfig{
			Enabled:    true,
			MaxRetries: 3,
			Timeout:    30 * time.Second,
		},
		Workflow: &vnext.WorkflowConfig{
			Mode:    vnext.Sequential,
			Timeout: 120 * time.Second,
		},
		Tracing: &vnext.TracingConfig{
			Enabled: true,
			Level:   "debug",
		},
		Streaming: &vnext.StreamingConfig{
			Enabled:    true,
			BufferSize: 100,
		},
	}

	// Validate the complete configuration
	err := vnext.ValidateConfig(config)
	if err != nil {
		t.Errorf("ValidateConfig() error = %v", err)
	}

	// Verify all features are configured
	if config.Memory == nil {
		t.Error("Memory should be configured")
	}
	if config.Tools == nil {
		t.Error("Tools should be configured")
	}
	if config.Workflow == nil {
		t.Error("Workflow should be configured")
	}
	if config.Tracing == nil {
		t.Error("Tracing should be configured")
	}
	if config.Streaming == nil {
		t.Error("Streaming should be configured")
	}
}
