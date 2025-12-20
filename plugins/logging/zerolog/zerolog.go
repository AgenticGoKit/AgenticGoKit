package zerolog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/agenticgokit/agenticgokit/core"
)

// Register the zerolog provider on import.
func init() {
	// Default to console-friendly output with time.
	// Users can reconfigure zerolog globally if needed.
	zerolog.TimeFieldFormat = time.RFC3339
	core.RegisterLoggingProvider("zerolog", core.LoggingProvider{
		New:       func() core.CoreLogger { return &coreZeroLogger{l: &zlog.Logger} },
		SetLevel:  setLevel,
		GetLevel:  getLevel,
		SetFormat: setFormat,
		SetFile:   setFile,
		SetConfig: setConfig,
	})
}

// Adapter: zerolog -> core.CoreLogger / core.LogEvent
type coreZeroLogger struct {
	l *zerolog.Logger
}

func (c *coreZeroLogger) Debug() core.LogEvent { return &zeroEvent{logger: c.l, evt: c.l.Debug(), level: "debug"} }
func (c *coreZeroLogger) Info() core.LogEvent  { return &zeroEvent{logger: c.l, evt: c.l.Info(), level: "info"} }
func (c *coreZeroLogger) Warn() core.LogEvent  { return &zeroEvent{logger: c.l, evt: c.l.Warn(), level: "warn"} }
func (c *coreZeroLogger) Error() core.LogEvent { return &zeroEvent{logger: c.l, evt: c.l.Error(), level: "error"} }
func (c *coreZeroLogger) With() core.LogEvent {
	return &zeroEvent{logger: c.l, fields: map[string]any{}}
}

type zeroEvent struct {
	logger *zerolog.Logger
	evt    *zerolog.Event
	fields map[string]any
	level  string // Track the intended log level
}

// helpers
func (e *zeroEvent) ensure(level string) *zerolog.Event {
	if e.evt != nil {
		return e.evt
	}
	switch level {
	case "debug":
		e.evt = e.logger.Debug()
	case "info":
		e.evt = e.logger.Info()
	case "warn":
		e.evt = e.logger.Warn()
	case "error":
		e.evt = e.logger.Error()
	default:
		e.evt = e.logger.Info()
	}
	// apply stored fields
	if len(e.fields) > 0 {
		for k, v := range e.fields {
			switch vv := v.(type) {
			case string:
				e.evt = e.evt.Str(k, vv)
			case []string:
				e.evt = e.evt.Strs(k, vv)
			case int:
				e.evt = e.evt.Int(k, vv)
			case bool:
				e.evt = e.evt.Bool(k, vv)
			case float64:
				e.evt = e.evt.Float64(k, vv)
			case time.Duration:
				e.evt = e.evt.Dur(k, vv)
			case time.Time:
				e.evt = e.evt.Time(k, vv)
			default:
				e.evt = e.evt.Interface(k, vv)
			}
		}
		e.fields = nil
	}
	return e.evt
}

// field methods
func (e *zeroEvent) Str(key, val string) core.LogEvent {
	if e.evt != nil {
		e.evt = e.evt.Str(key, val)
		return e
	}
	if e.fields == nil {
		e.fields = map[string]any{}
	}
	e.fields[key] = val
	return e
}

func (e *zeroEvent) Strs(key string, val []string) core.LogEvent {
	if e.evt != nil {
		e.evt = e.evt.Strs(key, val)
		return e
	}
	if e.fields == nil {
		e.fields = map[string]any{}
	}
	e.fields[key] = val
	return e
}

func (e *zeroEvent) Int(key string, val int) core.LogEvent {
	if e.evt != nil {
		e.evt = e.evt.Int(key, val)
		return e
	}
	if e.fields == nil {
		e.fields = map[string]any{}
	}
	e.fields[key] = val
	return e
}

func (e *zeroEvent) Bool(key string, val bool) core.LogEvent {
	if e.evt != nil {
		e.evt = e.evt.Bool(key, val)
		return e
	}
	if e.fields == nil {
		e.fields = map[string]any{}
	}
	e.fields[key] = val
	return e
}

func (e *zeroEvent) Float64(key string, val float64) core.LogEvent {
	if e.evt != nil {
		e.evt = e.evt.Float64(key, val)
		return e
	}
	if e.fields == nil {
		e.fields = map[string]any{}
	}
	e.fields[key] = val
	return e
}

func (e *zeroEvent) Dur(key string, val time.Duration) core.LogEvent {
	if e.evt != nil {
		e.evt = e.evt.Dur(key, val)
		return e
	}
	if e.fields == nil {
		e.fields = map[string]any{}
	}
	e.fields[key] = val
	return e
}

func (e *zeroEvent) Time(key string, val time.Time) core.LogEvent {
	if e.evt != nil {
		e.evt = e.evt.Time(key, val)
		return e
	}
	if e.fields == nil {
		e.fields = map[string]any{}
	}
	e.fields[key] = val
	return e
}

func (e *zeroEvent) Interface(key string, val interface{}) core.LogEvent {
	if e.evt != nil {
		e.evt = e.evt.Interface(key, val)
		return e
	}
	if e.fields == nil {
		e.fields = map[string]any{}
	}
	e.fields[key] = val
	return e
}

func (e *zeroEvent) Err(err error) core.LogEvent {
	if e.evt != nil {
		e.evt = e.evt.Err(err)
		return e
	}
	if e.fields == nil {
		e.fields = map[string]any{}
	}
	e.fields["error"] = err
	return e
}

func (e *zeroEvent) Msg(msg string)                          { e.ensure(e.level).Msg(msg) }
func (e *zeroEvent) Msgf(format string, args ...interface{}) { e.ensure(e.level).Msgf(format, args...) }

// chaining to choose level after With()
func (e *zeroEvent) Debug() core.LogEvent { e.level = "debug"; e.ensure("debug"); return e }
func (e *zeroEvent) Info() core.LogEvent  { e.level = "info"; e.ensure("info"); return e }
func (e *zeroEvent) Warn() core.LogEvent  { e.level = "warn"; e.ensure("warn"); return e }
func (e *zeroEvent) Error() core.LogEvent { e.level = "error"; e.ensure("error"); return e }

func (e *zeroEvent) Logger() core.CoreLogger { return &coreZeroLogger{l: e.logger} }

// level mapping helpers
func setLevel(l core.LogLevel) {
	switch l {
	case core.DEBUG:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case core.INFO:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case core.WARN:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case core.ERROR:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func getLevel() core.LogLevel {
	lvl := zerolog.GlobalLevel()
	switch lvl {
	case zerolog.DebugLevel:
		return core.DEBUG
	case zerolog.InfoLevel:
		return core.INFO
	case zerolog.WarnLevel:
		return core.WARN
	case zerolog.ErrorLevel:
		return core.ERROR
	default:
		return core.INFO
	}
}

var (
	currentConfig core.LoggingConfig
)

func setFormat(format string) {
	currentConfig.Format = format
	updateLogger()
}

func setFile(filePath string) {
	currentConfig.File = filePath
	updateLogger()
}

func setConfig(config core.LoggingConfig) {
	currentConfig = config
	updateLogger()
}

func updateLogger() {
	var writers []io.Writer
	
	// Add console output unless file_only is true
	if !currentConfig.FileOnly {
		switch currentConfig.Format {
		case "json":
			writers = append(writers, os.Stderr)
		default: // "console" or any other value defaults to console format
			writers = append(writers, zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
			})
		}
	}
	
	// Add file output if specified
	if currentConfig.File != "" {
		// Create directory if it doesn't exist
		if dir := filepath.Dir(currentConfig.File); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not create log directory %s: %v\n", dir, err)
				return
			}
		}
		
		var fileWriter io.Writer
		
		// Use lumberjack for log rotation if any rotation settings are specified
		if currentConfig.MaxSize > 0 || currentConfig.MaxBackups > 0 || currentConfig.MaxAge > 0 {
			fileWriter = &lumberjack.Logger{
				Filename:   currentConfig.File,
				MaxSize:    currentConfig.MaxSize,    // megabytes
				MaxBackups: currentConfig.MaxBackups, // number of backups
				MaxAge:     currentConfig.MaxAge,     // days
				Compress:   currentConfig.Compress,   // compress rotated files
			}
		} else {
			// Simple file logging without rotation
			file, err := os.OpenFile(currentConfig.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not open log file %s: %v\n", currentConfig.File, err)
				return
			}
			fileWriter = file
		}
		
		// File output is always JSON format for structured logging
		writers = append(writers, fileWriter)
	}
	
	// Create writer
	var writer io.Writer
	if len(writers) == 0 {
		// Fallback to stderr if no writers configured
		writer = os.Stderr
	} else if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = zerolog.MultiLevelWriter(writers...)
	}
	
	zlog.Logger = zerolog.New(writer).With().Timestamp().Logger()
}

