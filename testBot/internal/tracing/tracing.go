package tracing

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kunalkushwaha/agenticgokit/core"
)

// FlowEdge captures a hop in the agent flow for a given session.
type FlowEdge struct {
	From      string
	To        string
	Message   string
	Timestamp time.Time
	IsError   bool
	Err       string
}

var (
	mu     sync.Mutex
	traces = map[string][]FlowEdge{} // sessionID -> edges
)

// RecordEdge appends a normal (non-error) edge to the session trace.
func RecordEdge(sessionID, from, to, message string) {
	if sessionID == "" {
		sessionID = "webui-session"
	}
	edge := FlowEdge{From: from, To: to, Message: message, Timestamp: time.Now()}
	mu.Lock()
	traces[sessionID] = append(traces[sessionID], edge)
	mu.Unlock()
}

// RecordError appends an error note edge to the session trace.
func RecordError(sessionID, from, to, errMsg string) {
	if sessionID == "" {
		sessionID = "webui-session"
	}
	edge := FlowEdge{From: from, To: to, Timestamp: time.Now(), IsError: true, Err: errMsg}
	mu.Lock()
	traces[sessionID] = append(traces[sessionID], edge)
	mu.Unlock()
}

// GetEdges returns a copy of the edges captured for a given session.
func GetEdges(sessionID string) []FlowEdge {
	mu.Lock()
	defer mu.Unlock()
	out := append([]FlowEdge(nil), traces[sessionID]...)
	return out
}

// RecordAgentTransition inspects an agent result to determine the next route and records
// the corresponding edge and any error notes. If no explicit message is found in the result,
// fallbackContent is used for the edge label mapping.
func RecordAgentTransition(sessionID, agentName string, result core.AgentResult, fallbackContent string) {
	if sessionID == "" {
		sessionID = "webui-session"
	}
	nextTo := "User"
	if result.OutputState != nil {
		if route, ok := result.OutputState.GetMeta(core.RouteMetadataKey); ok && route != "" {
			nextTo = fmt.Sprintf("%v", route)
		}
	}
	if nextTo != "" {
		outMsg := fallbackContent
		if outMsg == "" && result.OutputState != nil {
			for _, key := range []string{"response", "output", "result", "content", "message"} {
				if v, ok := result.OutputState.Get(key); ok {
					if s, ok2 := v.(string); ok2 && s != "" {
						outMsg = s
						break
					}
				}
			}
		}
		RecordEdge(sessionID, agentName, nextTo, outMsg)
	}
	if result.OutputState != nil {
		if v, ok := result.OutputState.Get("error"); ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				RecordError(sessionID, agentName, nextTo, s)
			}
		}
		if v, ok := result.OutputState.Get("error_message"); ok {
			if s, ok2 := v.(string); ok2 && s != "" {
				RecordError(sessionID, agentName, nextTo, s)
			}
		}
	}
}

// LabelInfo provides a mapping from a short label to the full message content for UI tooltips.
type LabelInfo struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
}

// BuildMermaidSequenceLabeled builds a Mermaid sequence diagram from local shim edges and returns label mappings.
func BuildMermaidSequenceLabeled(edges []FlowEdge) (string, []LabelInfo) {
	var b strings.Builder
	b.WriteString("%%{init: {\"theme\": \"base\", \"sequence\": {\"mirrorActors\": false}, \"themeVariables\": {\"primaryColor\": \"#eef2ff\", \"primaryTextColor\": \"#0b1021\", \"actorBorder\": \"#4c6ef5\", \"actorBkg\": \"#eef2ff\", \"noteBkgColor\": \"#fff7ed\", \"noteTextColor\": \"#92400e\"}} }%%\n")
	b.WriteString("sequenceDiagram\n")
	b.WriteString("  autonumber\n")

	// Participants in insertion order
	seen := map[string]bool{}
	order := []string{}
	add := func(name string) {
		if name == "" {
			return
		}
		if !seen[name] {
			seen[name] = true
			order = append(order, name)
		}
	}
	for _, e := range edges {
		add(e.From)
		add(e.To)
	}
	if len(order) == 0 {
		b.WriteString("  participant User\n  participant Agent\n  User->>Agent: (no activity captured)\n")
		return b.String(), nil
	}
	for _, p := range order {
		b.WriteString("  participant " + escapeIdent(p) + "\n")
	}

	labels := make([]LabelInfo, 0, len(edges))
	msgCounter := 0
	for _, e := range edges {
		if e.IsError {
			errTxt := e.Err
			if len(errTxt) > 200 {
				errTxt = errTxt[:197] + "..."
			}
			who := e.From
			if who == "" {
				who = e.To
			}
			if who == "" {
				who = "User"
			}
			b.WriteString("  Note over " + escapeIdent(who) + ": ❌ ERROR: " + escapeText(errTxt) + "\n")
			continue
		}
		msgCounter++
		label := fmt.Sprintf("M%d", msgCounter)
		full := e.Message
		if full == "" {
			full = "(no message)"
		}
		labels = append(labels, LabelInfo{ID: label, From: e.From, To: e.To, Message: full})
		b.WriteString("  " + escapeIdent(e.From) + "->>" + escapeIdent(e.To) + ": " + escapeText(label) + "\n")
	}
	return b.String(), labels
}

// BuildMermaidFromFrameworkTrace converts core.Runner trace entries into a Mermaid diagram and labels.
func BuildMermaidFromFrameworkTrace(entries []core.TraceEntry, linear bool) (string, []LabelInfo) {
	var b strings.Builder
	b.WriteString("%%{init: {\"theme\": \"base\", \"sequence\": {\"mirrorActors\": false}, \"themeVariables\": {\"primaryColor\": \"#eef2ff\", \"primaryTextColor\": \"#0b1021\", \"actorBorder\": \"#4c6ef5\", \"actorBkg\": \"#eef2ff\", \"noteBkgColor\": \"#fff7ed\", \"noteTextColor\": \"#92400e\"}} }%%\n")
	b.WriteString("sequenceDiagram\n")
	b.WriteString("  autonumber\n")

	seen := map[string]bool{"User": true}
	order := []string{"User"}
	firstAgent := ""
	for _, e := range entries {
		if e.Type == "agent_start" && e.AgentID != "" {
			if !seen[e.AgentID] {
				seen[e.AgentID] = true
				order = append(order, e.AgentID)
			}
			if firstAgent == "" {
				firstAgent = e.AgentID
			}
		}
		if e.Type == "agent_end" && e.AgentID != "" {
			if !seen[e.AgentID] {
				seen[e.AgentID] = true
				order = append(order, e.AgentID)
			}
			if !linear && e.AgentResult != nil && e.AgentResult.OutputState != nil {
				if rt, ok := e.AgentResult.OutputState.GetMeta(core.RouteMetadataKey); ok && rt != "" {
					if !seen[rt] {
						seen[rt] = true
						order = append(order, rt)
					}
				}
			}
		}
	}
	if len(order) == 1 {
		b.WriteString("  participant User\n  participant Agent\n  User->>Agent: (no activity captured)\n")
		return b.String(), nil
	}
	for _, p := range order {
		b.WriteString("  participant " + escapeIdent(p) + "\n")
	}

	labels := []LabelInfo{}
	counter := 0
	if firstAgent != "" {
		var msg string
		for _, e := range entries {
			if e.Type == "agent_start" && e.AgentID == firstAgent && e.State != nil {
				for _, k := range []string{"message", "user_input", "response", "output", "result", "content"} {
					if v, ok := e.State.Get(k); ok {
						if s, ok2 := v.(string); ok2 && s != "" {
							msg = s
							break
						}
					}
				}
				if msg == "" {
					msg = "(no message)"
				}
				break
			}
		}
		counter++
		lid := fmt.Sprintf("M%d", counter)
		labels = append(labels, LabelInfo{ID: lid, From: "User", To: firstAgent, Message: msg})
		b.WriteString("  User->>" + escapeIdent(firstAgent) + ": " + escapeText(lid) + "\n")
	}

	for _, e := range entries {
		if e.Type != "agent_end" || e.AgentID == "" {
			continue
		}
		from := e.AgentID
		var outMsg string
		if e.AgentResult != nil && e.AgentResult.OutputState != nil {
			for _, k := range []string{"response", "output", "message", "result", "content"} {
				if v, ok := e.AgentResult.OutputState.Get(k); ok {
					if s, ok2 := v.(string); ok2 && s != "" {
						outMsg = s
						break
					}
				}
			}
		}
		if outMsg == "" {
			outMsg = fmt.Sprintf("Agent %s completed", from)
		}
		nextTo := "User"
		if e.AgentResult != nil && e.AgentResult.OutputState != nil {
			if rt, ok := e.AgentResult.OutputState.GetMeta(core.RouteMetadataKey); ok && rt != "" {
				nextTo = fmt.Sprintf("%v", rt)
			}
		}
		if linear {
			if from != firstAgent {
				nextTo = "User"
			}
		} else {
			if nextTo == from || nextTo == "" {
				continue
			}
		}
		counter++
		lid := fmt.Sprintf("M%d", counter)
		labels = append(labels, LabelInfo{ID: lid, From: from, To: nextTo, Message: outMsg})
		b.WriteString("  " + escapeIdent(from) + "->>" + escapeIdent(nextTo) + ": " + escapeText(lid) + "\n")
	}
	return b.String(), labels
}

// BuildTraceData chooses between framework trace and local shim to produce diagram and labels.
func BuildTraceData(runner core.Runner, session string, linear bool) (string, []LabelInfo) {
	if runner != nil {
		if entries, err := runner.DumpTrace(session); err == nil {
			if len(entries) > 0 {
				return BuildMermaidFromFrameworkTrace(entries, linear)
			}
			edges := GetEdges(session)
			return BuildMermaidSequenceLabeled(edges)
		}
	}
	edges := GetEdges(session)
	return BuildMermaidSequenceLabeled(edges)
}

func escapeIdent(s string) string {
	repl := s
	repl = strings.ReplaceAll(repl, " ", "_")
	repl = strings.ReplaceAll(repl, "-", "_")
	repl = strings.ReplaceAll(repl, ".", "_")
	return repl
}

func escapeText(s string) string {
	repl := strings.ReplaceAll(s, "\n", " ")
	repl = strings.ReplaceAll(repl, "\r", " ")
	repl = strings.ReplaceAll(repl, "\"", "\\\"")
	repl = strings.ReplaceAll(repl, "`", "'")
	return repl
}
