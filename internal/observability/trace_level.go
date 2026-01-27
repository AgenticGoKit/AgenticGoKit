package observability

import (
	"os"
	"strings"
)

// TraceLevel controls what data is captured in spans
type TraceLevel int

const (
	// TraceLevelMinimal captures only timing and status - for production
	TraceLevelMinimal TraceLevel = iota
	// TraceLevelStandard captures tokens, tool names, latency - default behavior
	TraceLevelStandard
	// TraceLevelDetailed captures full prompts, responses, tool args/outputs - for dev/eval
	TraceLevelDetailed
)

var currentTraceLevel TraceLevel = TraceLevelStandard

// GetTraceLevel returns the current trace level from AGK_TRACE_LEVEL env var
func GetTraceLevel() TraceLevel {
	level := os.Getenv("AGK_TRACE_LEVEL")
	switch strings.ToLower(level) {
	case "detailed", "full", "debug":
		return TraceLevelDetailed
	case "minimal", "min", "prod":
		return TraceLevelMinimal
	default:
		return TraceLevelStandard
	}
}

// IsDetailedTracing returns true if detailed tracing is enabled
func IsDetailedTracing() bool {
	return GetTraceLevel() == TraceLevelDetailed
}

// IsMinimalTracing returns true if minimal tracing is enabled
func IsMinimalTracing() bool {
	return GetTraceLevel() == TraceLevelMinimal
}

// TruncateForTrace truncates content if it exceeds maxLen, adding "..." suffix
func TruncateForTrace(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	if maxLen <= 3 {
		return "..."
	}
	return content[:maxLen-3] + "..."
}

// Content attribute keys (only populated at detailed level)
const (
	AttrPromptUser       = "agk.prompt.user"
	AttrPromptSystem     = "agk.prompt.system"
	AttrLLMResponse      = "agk.llm.response"
	AttrToolArguments    = "agk.tool.arguments"
	AttrToolResult       = "agk.tool.result"
	AttrReasoningThought = "agk.reasoning.thought"
)

// MaxContentLength is the max characters for content attributes (prevents huge spans)
const MaxContentLength = 4000
