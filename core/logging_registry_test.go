package core

import (
	"testing"
	"time"
)

// test logger implements CoreLogger/LogEvent with counters
type testLogger struct{ called bool }

func (t *testLogger) Debug() LogEvent { t.called = true; return &testEvent{} }
func (t *testLogger) Info() LogEvent  { t.called = true; return &testEvent{} }
func (t *testLogger) Warn() LogEvent  { t.called = true; return &testEvent{} }
func (t *testLogger) Error() LogEvent { t.called = true; return &testEvent{} }
func (t *testLogger) With() LogEvent  { t.called = true; return &testEvent{} }

type testEvent struct{}

func (e *testEvent) Str(string, string) LogEvent            { return e }
func (e *testEvent) Strs(string, []string) LogEvent         { return e }
func (e *testEvent) Int(string, int) LogEvent               { return e }
func (e *testEvent) Bool(string, bool) LogEvent             { return e }
func (e *testEvent) Float64(string, float64) LogEvent       { return e }
func (e *testEvent) Dur(string, time.Duration) LogEvent     { return e }
func (e *testEvent) Time(string, time.Time) LogEvent        { return e }
func (e *testEvent) Interface(string, interface{}) LogEvent { return e }
func (e *testEvent) Err(error) LogEvent                     { return e }
func (e *testEvent) Msg(string)                             {}
func (e *testEvent) Msgf(string, ...interface{})            {}
func (e *testEvent) Debug() LogEvent                        { return e }
func (e *testEvent) Info() LogEvent                         { return e }
func (e *testEvent) Warn() LogEvent                         { return e }
func (e *testEvent) Error() LogEvent                        { return e }
func (e *testEvent) Logger() CoreLogger                     { return &testLogger{} }

func TestLoggingProviderRegistry(t *testing.T) {
	// Register a test provider
	RegisterLoggingProvider("test-logger", LoggingProvider{
		New:      func() CoreLogger { return &testLogger{} },
		SetLevel: func(l LogLevel) {},
		GetLevel: func() LogLevel { return INFO },
	})

	// Switch to it explicitly
	if !UseLoggingProvider("test-logger") {
		t.Fatalf("failed to select test-logger provider")
	}

	// Verify Logger() returns a non-nil instance and supports chaining
	l := Logger()
	if l == nil {
		t.Fatalf("Logger() returned nil")
	}
	l.Info().Str("k", "v").Msg("test")

	// Level getters/setters should not panic
	SetLogLevel(DEBUG)
	_ = GetLogLevel()
}
