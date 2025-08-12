package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigReloader_Basic(t *testing.T) {
	// Create basic components
	validator := NewDefaultConfigValidator()
	config := &Config{}
	factory := NewConfigurableAgentFactory(config)
	agentManager := NewAgentManager(factory)
	
	// Create reloader
	reloader := NewConfigReloader(validator, agentManager)
	
	// Test basic properties
	assert.NotNil(t, reloader)
	assert.False(t, reloader.IsWatching())
	assert.Equal(t, 500*time.Millisecond, reloader.debouncePeriod)
	assert.Zero(t, reloader.GetLastReloadTime())
	
	// Test debounce period setting
	newPeriod := 1 * time.Second
	reloader.SetDebouncePeriod(newPeriod)
	assert.Equal(t, newPeriod, reloader.debouncePeriod)
	
	// Test callback registration
	var callbackCalled bool
	reloader.OnConfigChanged(func(config *Config, err error) {
		callbackCalled = true
	})
	
	// Simulate callback
	reloader.notifyCallbacks(nil, nil)
	time.Sleep(100 * time.Millisecond) // Give callback time to execute
	assert.True(t, callbackCalled)
}

func TestAgentManager_Basic(t *testing.T) {
	// Create basic components
	config := &Config{}
	factory := NewConfigurableAgentFactory(config)
	manager := NewAgentManager(factory)
	
	// Test basic properties
	assert.NotNil(t, manager)
	assert.Equal(t, 0, manager.GetAgentCount())
	assert.Equal(t, 0, manager.GetEnabledAgentCount())
	assert.Empty(t, manager.ListAgents())
	assert.Empty(t, manager.ListEnabledAgents())
	
	// Test getting non-existent agent
	_, exists := manager.GetAgent("non_existent")
	assert.False(t, exists)
	
	_, exists = manager.GetAgentConfig("non_existent")
	assert.False(t, exists)
}