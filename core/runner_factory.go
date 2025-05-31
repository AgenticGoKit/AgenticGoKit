// Package core provides public factory functions for creating runners and tool registries in AgentFlow.
package core

import (
	"context"
	"log"
)

// RunnerConfig allows customization but provides sensible defaults.
type RunnerConfig struct {
	QueueSize    int
	Orchestrator Orchestrator
	Agents       map[string]AgentHandler
	TraceLogger  TraceLogger // Optional trace logger
	ConfigPath   string      // Path to agentflow.toml config file
	Config       *Config     // Pre-loaded configuration (optional)
}

// NewRunnerWithConfig wires up everything, registers agents, and returns a ready-to-use runner.
func NewRunnerWithConfig(cfg RunnerConfig) Runner {
	// Load configuration if specified
	var config *Config
	if cfg.Config != nil {
		config = cfg.Config
	} else if cfg.ConfigPath != "" {
		var err error
		config, err = LoadConfig(cfg.ConfigPath)
		if err != nil {
			log.Printf("Warning: Failed to load config from %s: %v", cfg.ConfigPath, err)
		}
	} else {
		// Try to load from working directory
		var err error
		config, err = LoadConfigFromWorkingDir()
		if err != nil {
			log.Printf("Info: No agentflow.toml found in working directory: %v", err)
		}
	}

	// Apply configuration settings
	if config != nil {
		config.ApplyLoggingConfig()

		// Use configuration values if not specified in RunnerConfig
		if cfg.QueueSize <= 0 && config.Runtime.MaxConcurrentAgents > 0 {
			cfg.QueueSize = config.Runtime.MaxConcurrentAgents
		}

		Logger().Info().
			Str("config_name", config.AgentFlow.Name).
			Str("config_version", config.AgentFlow.Version).
			Str("config_provider", config.AgentFlow.Provider).
			Str("log_level", config.Logging.Level).
			Msg("Loaded AgentFlow configuration")
	}

	queueSize := cfg.QueueSize
	if queueSize <= 0 {
		queueSize = 10
	}
	runner := NewRunner(queueSize)

	// Callbacks and tracing
	callbackRegistry := NewCallbackRegistry()

	// Use provided trace logger or create default
	var traceLogger TraceLogger
	if cfg.TraceLogger != nil {
		traceLogger = cfg.TraceLogger
	} else {
		traceLogger = NewInMemoryTraceLogger()
	}

	runner.SetCallbackRegistry(callbackRegistry)
	runner.SetTraceLogger(traceLogger)
	RegisterTraceHooks(callbackRegistry, traceLogger)

	// Orchestrator
	var orch Orchestrator
	if cfg.Orchestrator != nil {
		orch = cfg.Orchestrator
	} else {
		orch = NewRouteOrchestrator(callbackRegistry)
	}
	runner.SetOrchestrator(orch)

	// Register agents
	for name, agent := range cfg.Agents {
		if err := runner.RegisterAgent(name, agent); err != nil {
			log.Fatalf("Failed to register agent %s: %v", name, err)
		}
	}
	// Register a default no-op error handler if not present
	if _, ok := cfg.Agents["error-handler"]; !ok {
		runner.RegisterAgent("error-handler", AgentHandlerFunc(
			func(ctx context.Context, event Event, state State) (AgentResult, error) {
				state.SetMeta(RouteMetadataKey, "")
				return AgentResult{OutputState: state}, nil
			},
		))
	}

	// Automatically configure error routing based on available error handlers
	configureErrorRouting(runner, cfg.Agents)

	return runner
}

// NewRunnerWithConfigFile creates a runner by loading configuration from the specified file
func NewRunnerWithConfigFile(configPath string, agents map[string]AgentHandler) Runner {
	return NewRunnerWithConfig(RunnerConfig{
		ConfigPath: configPath,
		Agents:     agents,
	})
}

// NewRunnerFromWorkingDir creates a runner by loading agentflow.toml from the working directory
func NewRunnerFromWorkingDir(agents map[string]AgentHandler) Runner {
	return NewRunnerWithConfig(RunnerConfig{
		Agents: agents,
	})
}

// NewProviderFromConfig creates a ModelProvider from the loaded configuration
func NewProviderFromConfig(configPath string) (ModelProvider, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return config.InitializeProvider()
}

// NewProviderFromWorkingDir creates a ModelProvider from agentflow.toml in the working directory
func NewProviderFromWorkingDir() (ModelProvider, error) {
	config, err := LoadConfigFromWorkingDir()
	if err != nil {
		return nil, err
	}
	return config.InitializeProvider()
}

// ...other factory functions (e.g., for tool registry, LLM adapter) can be added here...

// configureErrorRouting automatically configures error routing based on available error handlers
func configureErrorRouting(runner Runner, agents map[string]AgentHandler) {
	if runner == nil || agents == nil {
		return
	}

	// Check which specialized error handlers are available
	errorConfig := DefaultErrorRouterConfig()

	// Update category handlers based on available agents
	if _, exists := agents["validation-error-handler"]; exists {
		errorConfig.CategoryHandlers[ErrorCodeValidation] = "validation-error-handler"
	}
	if _, exists := agents["timeout-error-handler"]; exists {
		errorConfig.CategoryHandlers[ErrorCodeTimeout] = "timeout-error-handler"
	}
	if _, exists := agents["critical-error-handler"]; exists {
		errorConfig.SeverityHandlers[SeverityCritical] = "critical-error-handler"
	}
	if _, exists := agents["network-error-handler"]; exists {
		errorConfig.CategoryHandlers[ErrorCodeNetwork] = "network-error-handler"
	}
	if _, exists := agents["llm-error-handler"]; exists {
		errorConfig.CategoryHandlers[ErrorCodeLLM] = "llm-error-handler"
	}
	if _, exists := agents["auth-error-handler"]; exists {
		errorConfig.CategoryHandlers[ErrorCodeAuth] = "auth-error-handler"
	}

	// Set the default error handler if available
	if _, exists := agents["error-handler"]; exists {
		errorConfig.ErrorHandlerName = "error-handler"
	} else if _, exists := agents["error_handler"]; exists {
		// Support underscore naming convention as well
		errorConfig.ErrorHandlerName = "error_handler"
	}
	// Apply the configuration to the runner
	if runnerImpl, ok := runner.(*RunnerImpl); ok {
		runnerImpl.SetErrorRouterConfig(errorConfig)
	}
}
