// Package logging provides internal logging implementations for AgentFlow.
package logging

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	logger   zerolog.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	logLevel LogLevel       = INFO
	mu       sync.RWMutex
)

func SetLogLevel(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	logLevel = level
	zerolog.SetGlobalLevel(mapLogLevel(level))
}

func GetLogLevel() LogLevel {
	mu.RLock()
	defer mu.RUnlock()
	return logLevel
}

func GetLogger() *zerolog.Logger {
	return &logger
}

func mapLogLevel(level LogLevel) zerolog.Level {
	switch level {
	case DEBUG:
		return zerolog.DebugLevel
	case INFO:
		return zerolog.InfoLevel
	case WARN:
		return zerolog.WarnLevel
	case ERROR:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
