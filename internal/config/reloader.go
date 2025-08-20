// Package config provides internal configuration reloading functionality.
package config

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kunalkushwaha/agenticgokit/core"
)

// ConfigReloadCallback defines the callback function signature for configuration changes
type ConfigReloadCallback func(*core.Config, error)

// DefaultConfigReloader implements the core.ConfigReloader interface
type DefaultConfigReloader struct {
	configPath      string
	watcher         *fsnotify.Watcher
	validator       core.ConfigValidator
	agentManager    core.AgentManager
	callbacks       []ConfigReloadCallback
	currentConfig   *core.Config
	lastReloadTime  time.Time
	isWatching      bool
	debounceTimer   *time.Timer
	debouncePeriod  time.Duration
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewConfigReloader creates a new configuration reloader
func NewConfigReloader(validator core.ConfigValidator, agentManager core.AgentManager) *DefaultConfigReloader {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &DefaultConfigReloader{
		validator:      validator,
		agentManager:   agentManager,
		callbacks:      make([]ConfigReloadCallback, 0),
		debouncePeriod: 500 * time.Millisecond, // Default debounce period
		ctx:            ctx,
		cancel:         cancel,
	}
}

// StartWatching begins monitoring the configuration file for changes
func (r *DefaultConfigReloader) StartWatching(configPath string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isWatching {
		return fmt.Errorf("already watching configuration file: %s", r.configPath)
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Create file system watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Add the configuration file to the watcher
	err = watcher.Add(absPath)
	if err != nil {
		watcher.Close()
		return fmt.Errorf("failed to watch config file %s: %w", absPath, err)
	}

	r.configPath = absPath
	r.watcher = watcher
	r.isWatching = true

	// Load initial configuration
	err = r.loadInitialConfig()
	if err != nil {
		core.Logger().Warn().
			Err(err).
			Str("config_path", absPath).
			Msg("Failed to load initial configuration, will retry on file changes")
	}

	// Start watching in a goroutine
	go r.watchLoop()

	core.Logger().Debug().
		Str("config_path", absPath).
		Msg("Started watching configuration file for changes")

	return nil
}

// StopWatching stops monitoring the configuration file
func (r *DefaultConfigReloader) StopWatching() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.isWatching {
		return nil
	}

	// Cancel context to stop watch loop
	r.cancel()

	// Stop debounce timer if running
	if r.debounceTimer != nil {
		r.debounceTimer.Stop()
		r.debounceTimer = nil
	}

	// Close file watcher
	if r.watcher != nil {
		err := r.watcher.Close()
		r.watcher = nil
		if err != nil {
			core.Logger().Error().
				Err(err).
				Msg("Error closing file watcher")
		}
	}

	r.isWatching = false
	r.configPath = ""

	core.Logger().Debug().Msg("Stopped watching configuration file")
	return nil
}

// ReloadConfig manually triggers a configuration reload
func (r *DefaultConfigReloader) ReloadConfig() error {
	r.mutex.RLock()
	configPath := r.configPath
	r.mutex.RUnlock()

	if configPath == "" {
		return fmt.Errorf("no configuration file being watched")
	}

	return r.handleConfigChange()
}

// OnConfigChanged registers a callback for configuration change events
func (r *DefaultConfigReloader) OnConfigChanged(callback func(*core.Config, error)) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.callbacks = append(r.callbacks, callback)
}

// GetLastReloadTime returns the timestamp of the last successful configuration reload
func (r *DefaultConfigReloader) GetLastReloadTime() time.Time {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.lastReloadTime
}

// IsWatching returns whether the reloader is currently watching for file changes
func (r *DefaultConfigReloader) IsWatching() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.isWatching
}

// loadInitialConfig loads the configuration file for the first time
func (r *DefaultConfigReloader) loadInitialConfig() error {
	config, err := core.LoadConfig(r.configPath)
	if err != nil {
		return fmt.Errorf("failed to load initial config: %w", err)
	}

	// Validate configuration
	validationErrors := r.validator.ValidateConfig(config)
	if len(validationErrors) > 0 {
		core.Logger().Warn().
			Int("error_count", len(validationErrors)).
			Msg("Initial configuration has validation warnings")
		
		for _, validationError := range validationErrors {
			core.Logger().Warn().
				Str("field", validationError.Field).
				Interface("value", validationError.Value).
				Str("suggestion", validationError.Suggestion).
				Msg("Configuration validation warning")
		}
	}

	// Create resolver with config and apply environment overrides
	resolver := core.NewConfigResolver(config)
	err = resolver.ApplyEnvironmentOverrides()
	if err != nil {
		return fmt.Errorf("failed to resolve initial config: %w", err)
	}
	
	resolvedConfig := resolver.GetResolvedConfig()

	r.currentConfig = resolvedConfig
	r.lastReloadTime = time.Now()

	// Update agent manager with initial configuration
	if r.agentManager != nil {
		err = r.agentManager.UpdateAgentConfigurations(resolvedConfig)
		if err != nil {
			core.Logger().Error().
				Err(err).
				Msg("Failed to update agent configurations with initial config")
			// Don't fail completely, just log the error
		}
	}

	core.Logger().Debug().
		Str("config_path", r.configPath).
		Msg("Initial configuration loaded successfully")

	return nil
}

// watchLoop runs the file watching loop
func (r *DefaultConfigReloader) watchLoop() {
	defer func() {
		if r := recover(); r != nil {
			core.Logger().Error().
				Interface("panic", r).
				Msg("Config watcher panic recovered")
		}
	}()

	for {
		select {
		case <-r.ctx.Done():
			core.Logger().Debug().Msg("Config watcher stopping due to context cancellation")
			return

		case event, ok := <-r.watcher.Events:
			if !ok {
				core.Logger().Debug().Msg("Config watcher events channel closed")
				return
			}

			// Only handle write and create events for our config file
			if event.Name == r.configPath && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				core.Logger().Debug().
					Str("event", event.String()).
					Msg("Configuration file change detected")

				r.debounceConfigChange()
			}

		case err, ok := <-r.watcher.Errors:
			if !ok {
				core.Logger().Debug().Msg("Config watcher errors channel closed")
				return
			}

			core.Logger().Error().
				Err(err).
				Msg("File watcher error")
		}
	}
}

// debounceConfigChange implements debouncing to avoid excessive reloads
func (r *DefaultConfigReloader) debounceConfigChange() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Stop existing timer
	if r.debounceTimer != nil {
		r.debounceTimer.Stop()
	}

	// Start new timer
	r.debounceTimer = time.AfterFunc(r.debouncePeriod, func() {
		err := r.handleConfigChange()
		if err != nil {
			core.Logger().Error().
				Err(err).
				Msg("Failed to handle configuration change")
		}
	})
}

// handleConfigChange processes a configuration file change
func (r *DefaultConfigReloader) handleConfigChange() error {
	core.Logger().Debug().
		Str("config_path", r.configPath).
		Msg("Processing configuration file change")

	// Load new configuration
	newConfig, err := core.LoadConfig(r.configPath)
	if err != nil {
		core.Logger().Error().
			Err(err).
			Str("config_path", r.configPath).
			Msg("Failed to parse new configuration")
		r.notifyCallbacks(nil, fmt.Errorf("failed to parse configuration: %w", err))
		return err
	}

	// Validate new configuration
	validationErrors := r.validator.ValidateConfig(newConfig)
	if len(validationErrors) > 0 {
		// Log warnings but don't fail the reload for warnings
		core.Logger().Warn().
			Int("error_count", len(validationErrors)).
			Msg("New configuration has validation warnings")
		
		for _, validationError := range validationErrors {
			core.Logger().Warn().
				Str("field", validationError.Field).
				Interface("value", validationError.Value).
				Str("suggestion", validationError.Suggestion).
				Msg("Configuration validation warning")
		}
	}

	// Create resolver with new config and apply environment overrides
	resolver := core.NewConfigResolver(newConfig)
	err = resolver.ApplyEnvironmentOverrides()
	if err != nil {
		core.Logger().Error().
			Err(err).
			Msg("Failed to apply environment overrides to new configuration")
		r.notifyCallbacks(nil, fmt.Errorf("failed to resolve configuration: %w", err))
		return err
	}
	
	resolvedConfig := resolver.GetResolvedConfig()

	// Update agent configurations
	if r.agentManager != nil {
		err = r.agentManager.UpdateAgentConfigurations(resolvedConfig)
		if err != nil {
			core.Logger().Error().
				Err(err).
				Msg("Failed to update agent configurations")
			r.notifyCallbacks(nil, fmt.Errorf("failed to update agents: %w", err))
			return err
		}
	}

	// Update current configuration
	r.mutex.Lock()
	r.currentConfig = resolvedConfig
	r.lastReloadTime = time.Now()
	r.mutex.Unlock()

	core.Logger().Debug().
		Str("config_path", r.configPath).
		Time("reload_time", r.lastReloadTime).
		Msg("Configuration reloaded successfully")

	// Notify callbacks of successful reload
	r.notifyCallbacks(resolvedConfig, nil)

	return nil
}

// notifyCallbacks notifies all registered callbacks of configuration changes
func (r *DefaultConfigReloader) notifyCallbacks(config *core.Config, err error) {
	r.mutex.RLock()
	callbacks := make([]ConfigReloadCallback, len(r.callbacks))
	copy(callbacks, r.callbacks)
	r.mutex.RUnlock()

	for _, callback := range callbacks {
		go func(cb ConfigReloadCallback) {
			defer func() {
				if r := recover(); r != nil {
					core.Logger().Error().
						Interface("panic", r).
						Msg("Config reload callback panic recovered")
				}
			}()
			cb(config, err)
		}(callback)
	}
}

// GetCurrentConfig returns the current configuration (for testing/debugging)
func (r *DefaultConfigReloader) GetCurrentConfig() *core.Config {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.currentConfig
}

// SetDebouncePeriod sets the debounce period for file change events
func (r *DefaultConfigReloader) SetDebouncePeriod(period time.Duration) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.debouncePeriod = period
}
