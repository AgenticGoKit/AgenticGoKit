package tests

import (
	"os"
	"testing"

	"github.com/kunalkushwaha/agenticgokit/core"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := core.LoadConfig("") // Should load defaults if no file
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig returned nil config")
	}
	if cfg.AgentFlow.Name == "" {
		t.Error("AgentFlow.Name should have a default value")
	}
	if cfg.Logging.Level == "" {
		t.Error("Logging.Level should have a default value")
	}
}

func TestLoadConfig_File(t *testing.T) {
	// Create a minimal config file
	file := "test_agentflow.toml"
	content := `[agent_flow]
name = "TestAgent"
version = "0.1"
provider = "mock"
[logging]
level = "debug"
format = "text"
`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	defer os.Remove(file)

	cfg, err := core.LoadConfig(file)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.AgentFlow.Name != "TestAgent" {
		t.Errorf("AgentFlow.Name mismatch: got %q", cfg.AgentFlow.Name)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level mismatch: got %q", cfg.Logging.Level)
	}
}

func TestInitializeProvider_MissingProvider(t *testing.T) {
	cfg := &core.Config{}
	cfg.AgentFlow.Provider = ""
	_, err := cfg.InitializeProvider()
	if err == nil {
		t.Error("Expected error for missing provider, got nil")
	}
}

func TestInitializeProvider_UnknownProvider(t *testing.T) {
	cfg := &core.Config{}
	cfg.AgentFlow.Provider = "unknown"
	_, err := cfg.InitializeProvider()
	if err == nil {
		t.Error("Expected error for unknown provider, got nil")
	}
}