// Package agenttest provides LLM-free test doubles so agents, tools, and
// workflows can be tested in ordinary `go test` runs with no network, no API
// keys, and no cost.
package agenttest

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/agenticgokit/agentkit"
)

// Script is a Provider that replays a fixed sequence of responses and records
// every request it received, so tests can assert on what the agent actually
// sent to the model.
type Script struct {
	mu    sync.Mutex
	steps []agentkit.Response
	next  int
	reqs  []agentkit.Request

	// Caps is returned by Capabilities. The zero value reports a fully
	// capable model; set fields to exercise degraded paths.
	Caps *agentkit.Capabilities
}

// NewScript returns a provider that returns the given responses in order.
func NewScript(steps ...agentkit.Response) *Script {
	return &Script{steps: steps}
}

// Text is a response containing only assistant text.
func Text(s string) agentkit.Response {
	return agentkit.Response{
		Message:    agentkit.Assistant(s),
		StopReason: agentkit.StopEndTurn,
		Usage:      agentkit.Usage{InputTokens: 10, OutputTokens: 5},
	}
}

// Calls is a response in which the model requests one or more tool calls.
func Calls(calls ...agentkit.ToolCall) agentkit.Response {
	parts := make([]agentkit.Part, 0, len(calls))
	for _, c := range calls {
		parts = append(parts, c)
	}
	return agentkit.Response{
		Message:    agentkit.Message{Role: agentkit.RoleAssistant, Parts: parts},
		StopReason: agentkit.StopToolUse,
		Usage:      agentkit.Usage{InputTokens: 10, OutputTokens: 5},
	}
}

// Call builds a tool call whose arguments are the JSON encoding of args.
func Call(id, name string, args any) agentkit.ToolCall {
	raw, err := json.Marshal(args)
	if err != nil {
		panic("agenttest: cannot marshal tool args: " + err.Error())
	}
	return agentkit.ToolCall{ID: id, Name: name, Input: raw}
}

// Name implements agentkit.Provider.
func (s *Script) Name() string { return "agenttest.Script" }

// Generate returns the next scripted response.
func (s *Script) Generate(ctx context.Context, req agentkit.Request) (*agentkit.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reqs = append(s.reqs, req)
	if s.next >= len(s.steps) {
		return nil, fmt.Errorf("agenttest: script exhausted after %d responses; the agent made more model calls than the test scripted", len(s.steps))
	}
	resp := s.steps[s.next]
	s.next++
	return &resp, nil
}

// Stream replays the next scripted response as a stream of events.
func (s *Script) Stream(ctx context.Context, req agentkit.Request) (agentkit.Stream, error) {
	resp, err := s.Generate(ctx, req)
	if err != nil {
		return nil, err
	}

	var events []agentkit.Event
	for _, p := range resp.Message.Parts {
		switch v := p.(type) {
		case agentkit.Text:
			events = append(events, agentkit.TextDelta{Text: v.Text})
		case agentkit.Reasoning:
			events = append(events, agentkit.ReasoningDelta{Text: v.Text})
		case agentkit.ToolCall:
			events = append(events,
				agentkit.ToolCallStart{ID: v.ID, Name: v.Name},
				agentkit.ToolCallDelta{ID: v.ID, InputJSON: string(v.Input)},
				agentkit.ToolCallDone{ID: v.ID, Input: v.Input},
			)
		}
	}
	events = append(events, agentkit.Done{StopReason: resp.StopReason, Usage: resp.Usage})
	return &sliceStream{events: events}, nil
}

// Capabilities reports a fully capable model unless Caps is set.
func (s *Script) Capabilities(model string) agentkit.Capabilities {
	if s.Caps != nil {
		return *s.Caps
	}
	return agentkit.Capabilities{
		NativeTools:     true,
		ParallelTools:   true,
		NativeJSON:      true,
		Streaming:       true,
		Input:           []agentkit.Modality{agentkit.ModalityText},
		ContextWindow:   128000,
		MaxOutputTokens: 4096,
	}
}

// Requests returns every request the agent sent, in order.
func (s *Script) Requests() []agentkit.Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]agentkit.Request, len(s.reqs))
	copy(out, s.reqs)
	return out
}

// LastRequest returns the most recent request, or false if none was made.
func (s *Script) LastRequest() (agentkit.Request, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.reqs) == 0 {
		return agentkit.Request{}, false
	}
	return s.reqs[len(s.reqs)-1], true
}

// sliceStream replays a fixed event slice.
type sliceStream struct {
	events []agentkit.Event
	i      int
	cur    agentkit.Event
	closed bool
}

func (s *sliceStream) Next() bool {
	if s.closed || s.i >= len(s.events) {
		return false
	}
	s.cur = s.events[s.i]
	s.i++
	return true
}

func (s *sliceStream) Event() agentkit.Event { return s.cur }
func (s *sliceStream) Err() error            { return nil }
func (s *sliceStream) Close() error          { s.closed = true; return nil }
