package config

import (
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// mockValidator is a simple mock for testing
type mockValidator struct{}

func (m *mockValidator) ValidateAgentConfig(name string, config *core.AgentConfig) []core.ValidationError {
	return []core.ValidationError{}
}

func (m *mockValidator) ValidateLLMConfig(config *core.AgentLLMConfig) []core.ValidationError {
	return []core.ValidationError{}
}

func (m *mockValidator) ValidateOrchestrationAgents(orchestration *core.OrchestrationConfigToml, agents map[string]core.AgentConfig) []core.ValidationError {
	return []core.ValidationError{}
}

func (m *mockValidator) ValidateCapabilities(capabilities []string) []core.ValidationError {
	return []core.ValidationError{}
}

func (m *mockValidator) ValidateConfig(config *core.Config) []core.ValidationError {
	return []core.ValidationError{}
}

// mockAgentManager is a simple mock for testing
type mockAgentManager struct{}

func (m *mockAgentManager) UpdateAgentConfigurations(config *core.Config) error {
	return nil
}

func (m *mockAgentManager) GetCurrentAgents() map[string]core.Agent {
	return make(map[string]core.Agent)
}

func (m *mockAgentManager) CreateAgent(name string, config *core.ResolvedAgentConfig) (core.Agent, error) {
	return nil, nil
}

func (m *mockAgentManager) DisableAgent(name string) error {
	return nil
}

func TestConfigReloader_Basic(t *testing.T) {
	// Create basic components
	validator := &mockValidator{}
	agentManager := &mockAgentManager{}
	
	// Create reloader
	reloader := NewConfigReloader(validator, agentManager)
	
	// Test basic properties
	if reloader == nil {
		t.Fatal("Expected reloader to be created")
	}
	
	if reloader.IsWatching() {
		t.Error("Expected reloader to not be watching initially")
	}
	
	if !reloader.GetLastReloadTime().IsZero() {
		t.Error("Expected last reload time to be zero initially")
	}
	
	// Test debounce period setting
	newPeriod := 1 * time.Second
	reloader.SetDebouncePeriod(newPeriod)
	
	// Test callback registration
	var callbackCalled bool
	reloader.OnConfigChanged(func(config *core.Config, err error) {
		callbackCalled = true
	})
	
	// Simulate callback
	reloader.notifyCallbacks(nil, nil)
	time.Sleep(100 * time.Millisecond) // Give callback time to execute
	
	if !callbackCalled {
		t.Error("Expected callback to be called")
	}
}

func TestConfigReloader_StopWatchingWhenNotWatching(t *testing.T) {
	validator := &mockValidator{}
	agentManager := &mockAgentManager{}
	reloader := NewConfigReloader(validator, agentManager)
	
	// Should not error when stopping while not watching
	err := reloader.StopWatching()
	if err != nil {
		t.Errorf("Expected no error when stopping while not watching, got: %v", err)
	}
}

func TestConfigReloader_ReloadConfigWithoutWatching(t *testing.T) {
	validator := &mockValidator{}
	agentManager := &mockAgentManager{}
	reloader := NewConfigReloader(validator, agentManager)
	
	// Should error when trying to reload without watching
	err := reloader.ReloadConfig()
	if err == nil {
		t.Error("Expected error when reloading without watching")
	}
}

func TestConfigReloader_GetCurrentConfig(t *testing.T) {
	validator := &mockValidator{}
	agentManager := &mockAgentManager{}
	reloader := NewConfigReloader(validator, agentManager)
	
	// Should return nil initially
	config := reloader.GetCurrentConfig()
	if config != nil {
		t.Error("Expected current config to be nil initially")
	}
}