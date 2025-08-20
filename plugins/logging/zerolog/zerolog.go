package zerolog

import (
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// Register the zerolog provider on import.
func init() {
	// Default to console-friendly output with time.
	// Users can reconfigure zerolog globally if needed.
	zerolog.TimeFieldFormat = time.RFC3339
	core.RegisterLoggingProvider("zerolog", core.LoggingProvider{
		New:      func() core.CoreLogger { return &coreZeroLogger{l: &zlog.Logger} },
		SetLevel: setLevel,
		GetLevel: getLevel,
	})
}

// Adapter: zerolog -> core.CoreLogger / core.LogEvent
type coreZeroLogger struct {
	l *zerolog.Logger
}

func (c *coreZeroLogger) Debug() core.LogEvent { return &zeroEvent{logger: c.l, evt: c.l.Debug()} }
func (c *coreZeroLogger) Info() core.LogEvent  { return &zeroEvent{logger: c.l, evt: c.l.Info()} }
func (c *coreZeroLogger) Warn() core.LogEvent  { return &zeroEvent{logger: c.l, evt: c.l.Warn()} }
func (c *coreZeroLogger) Error() core.LogEvent { return &zeroEvent{logger: c.l, evt: c.l.Error()} }
func (c *coreZeroLogger) With() core.LogEvent {
	return &zeroEvent{logger: c.l, fields: map[string]any{}}
}

type zeroEvent struct {
	logger *zerolog.Logger
	evt    *zerolog.Event
	fields map[string]any
}

// helpers
func (e *zeroEvent) ensure(level string) *zerolog.Event {
	if e.evt != nil {
		return e.evt
	}
	switch level {
	case "debug":
		e.evt = e.logger.Debug()
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

func (e *zeroEvent) Msg(msg string)                          { e.ensure("info").Msg(msg) }
func (e *zeroEvent) Msgf(format string, args ...interface{}) { e.ensure("info").Msgf(format, args...) }

// chaining to choose level after With()
func (e *zeroEvent) Debug() core.LogEvent { e.ensure("debug"); return e }
func (e *zeroEvent) Info() core.LogEvent  { e.ensure("info"); return e }
func (e *zeroEvent) Warn() core.LogEvent  { e.ensure("warn"); return e }
func (e *zeroEvent) Error() core.LogEvent { e.ensure("error"); return e }

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
