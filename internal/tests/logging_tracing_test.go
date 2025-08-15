package tests

import (
	"testing"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

func TestLoggingAndTracingConsolidation(t *testing.T) {
	// Test logging functionality
	t.Run("LogLevel", func(t *testing.T) {
		// Test setting and getting log level
		core.SetLogLevel(core.DEBUG)
		level := core.GetLogLevel()
		if level != core.DEBUG {
			t.Errorf("Expected log level DEBUG, got %v", level)
		}

		core.SetLogLevel(core.INFO)
		level = core.GetLogLevel()
		if level != core.INFO {
			t.Errorf("Expected log level INFO, got %v", level)
		}
	})

	t.Run("Logger", func(t *testing.T) {
		// Test logger creation and method chaining
		logger := core.Logger()
		if logger == nil {
			t.Error("Logger should not be nil")
		}

		// Test that all logging methods can be called without panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Logger methods should not panic: %v", r)
			}
		}()

		logger.Debug().Str("test", "value").Msg("debug message")
		logger.Info().Int("count", 42).Msg("info message")
		logger.Warn().Bool("flag", true).Msg("warn message")
		logger.Error().Err(nil).Msg("error message")

		// Test With() method chaining
		logger.With().Str("component", "test").Info().Msg("with chaining")
	})

	t.Run("TraceLogger", func(t *testing.T) {
		// Test trace logger functionality
		traceLogger := core.NewInMemoryTraceLogger()
		if traceLogger == nil {
			t.Error("TraceLogger should not be nil")
		}

		// Test logging trace entries
		entry := core.TraceEntry{
			Timestamp: time.Now(),
			Type:      "test_event",
			EventID:   "event_123",
			SessionID: "session_456",
			AgentID:   "agent_789",
		}

		err := traceLogger.Log(entry)
		if err != nil {
			t.Errorf("Logging trace entry should not fail: %v", err)
		}

		// Test retrieving trace entries
		traces, err := traceLogger.GetTrace("session_456")
		if err != nil {
			t.Errorf("Getting trace should not fail: %v", err)
		}

		if len(traces) != 1 {
			t.Errorf("Expected 1 trace entry, got %d", len(traces))
		}

		if traces[0].EventID != "event_123" {
			t.Errorf("Expected EventID 'event_123', got '%s'", traces[0].EventID)
		}
	})

	t.Run("TraceHooksRegistration", func(t *testing.T) {
		// Test that trace hooks registration works
		registry := core.NewCallbackRegistry()
		traceLogger := core.NewInMemoryTraceLogger()

		err := core.RegisterTraceHooks(registry, traceLogger)
		if err != nil {
			t.Errorf("Registering trace hooks should not fail: %v", err)
		}

		// Test error cases
		err = core.RegisterTraceHooks(nil, traceLogger)
		if err == nil {
			t.Error("Expected error when registry is nil")
		}

		err = core.RegisterTraceHooks(registry, nil)
		if err == nil {
			t.Error("Expected error when logger is nil")
		}
	})
}