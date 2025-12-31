// package v1beta provides the next-generation Agent API for AgenticGoKit
package v1beta

import (
	"fmt"
	"sync"
)

// =============================================================================
// FACTORY.GO - Unified Factory and Initialization System
// =============================================================================
//
// This file provides:
// - Agent preset registry and factories
// - QuickStart functions for common scenarios
// - Global initialization and configuration helpers
//
// Factory patterns:
// - Preset registry for pre-configured agent types
// - QuickStart functions using NewBuilder() for minimal setup
// - Integration with existing factory systems (Memory, Tools, Workflow)
//
// =============================================================================

// =============================================================================
// AGENT PRESET REGISTRY
// =============================================================================

// AgentPreset represents a pre-configured agent type
//
// Presets provide named configurations for common agent patterns.
// Each preset includes a builder function that creates a fully configured agent.
type AgentPreset struct {
	Name        string                       // Preset name (e.g., "chat", "research")
	Description string                       // Human-readable description
	Builder     func(*Config) (Agent, error) // Function to build the agent
}

var (
	agentPresets = make(map[string]*AgentPreset)
	presetMutex  sync.RWMutex
)

// RegisterAgentPreset registers a named agent preset
//
// Presets allow you to define reusable agent configurations with custom names.
//
// Example:
//
//	RegisterAgentPreset(&AgentPreset{
//	    Name: "my-agent",
//	    Description: "Custom agent",
//	    Builder: func(cfg *Config) (Agent, error) {
//	        return NewBuilder("my-agent").WithConfig(cfg).Build()
//	    },
//	})
func RegisterAgentPreset(preset *AgentPreset) error {
	if preset == nil {
		err := ConfigError("preset", "preset", nil)
		err.Details["message"] = "Agent preset cannot be nil"
		return err
	}
	if preset.Name == "" {
		err := ConfigError("preset", "name", nil)
		err.Details["message"] = "Agent preset name is required"
		return err
	}
	if preset.Builder == nil {
		err := ConfigError("preset", "builder", nil)
		err.Details["message"] = "Agent preset builder function is required"
		return err
	}

	presetMutex.Lock()
	defer presetMutex.Unlock()
	agentPresets[preset.Name] = preset
	return nil
}

// GetAgentPreset retrieves a registered agent preset by name
//
// Returns nil if the preset is not found.
func GetAgentPreset(name string) *AgentPreset {
	presetMutex.RLock()
	defer presetMutex.RUnlock()
	return agentPresets[name]
}

// ListAgentPresets returns all registered preset names
func ListAgentPresets() []string {
	presetMutex.RLock()
	defer presetMutex.RUnlock()

	names := make([]string, 0, len(agentPresets))
	for name := range agentPresets {
		names = append(names, name)
	}
	return names
}

// =============================================================================
// INITIALIZATION AND SETUP
// =============================================================================

var (
	initOnce      sync.Once
	isInitialized bool
	initError     error
	initMutex     sync.RWMutex
)

// InitializeDefaults initializes the vNext API with built-in presets
//
// This should be called once at application startup.
// Calling multiple times is safe - it only initializes once.
//
// Example:
//
//	func main() {
//	    if err := vnext.InitializeDefaults(); err != nil {
//	        log.Fatal(err)
//	    }
//	    agent, _ := vnext.QuickChatAgent("gpt-4")
//	}
func InitializeDefaults() error {
	initOnce.Do(func() {
		// Register built-in presets
		if err := registerBuiltinPresets(); err != nil {
			initError = err
			return
		}

		// Mark as initialized
		initMutex.Lock()
		isInitialized = true
		initMutex.Unlock()
	})

	return initError
}

// IsInitialized returns whether InitializeDefaults has been called
func IsInitialized() bool {
	initMutex.RLock()
	defer initMutex.RUnlock()
	return isInitialized
}

// registerBuiltinPresets registers the standard agent presets
func registerBuiltinPresets() error {
	// Register chat preset
	chatPreset := &AgentPreset{
		Name:        "chat",
		Description: "Conversational agent optimized for interactive chat",
		Builder: func(cfg *Config) (Agent, error) {
			builder := NewBuilder("chat-agent").WithPreset(ChatAgent)
			if cfg != nil {
				builder = builder.WithConfig(cfg)
			}
			return builder.Build()
		},
	}
	if err := RegisterAgentPreset(chatPreset); err != nil {
		return err
	}

	// Register research preset
	researchPreset := &AgentPreset{
		Name:        "research",
		Description: "Research agent optimized for information gathering",
		Builder: func(cfg *Config) (Agent, error) {
			builder := NewBuilder("research-agent").WithPreset(ResearchAgent)
			if cfg != nil {
				builder = builder.WithConfig(cfg)
			}
			return builder.Build()
		},
	}
	if err := RegisterAgentPreset(researchPreset); err != nil {
		return err
	}

	return nil
}

// SetupLogging configures the logging level
//
// Valid levels: "debug", "info", "warn", "error"
func SetupLogging(level string) error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[level] {
		err := ConfigError("logging", "level", nil)
		err.Details["message"] = fmt.Sprintf("Invalid log level '%s'. Valid levels: debug, info, warn, error", level)
		return err
	}

	// TODO: Implement actual logging configuration
	return nil
}
