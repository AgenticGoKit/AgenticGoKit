package agentkit

import (
	"encoding/json"
	"strings"
)

// Role identifies who produced a message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Part is one piece of message content. The closed set of implementations is
// the provider-neutral intermediate representation every adapter translates
// to and from; adapters must never require callers to speak a provider's
// wire format.
type Part interface{ isPart() }

// Text is plain text content.
type Text struct {
	Text string
}

// Image is image content, referenced by URL or carried inline.
type Image struct {
	URL  string
	Data []byte
	MIME string
}

// Audio is audio content, referenced by URL or carried inline.
type Audio struct {
	URL  string
	Data []byte
	MIME string
}

// Reasoning is a reasoning/thinking block from a reasoning model. Signature
// carries provider-specific verification data that must be echoed back on
// subsequent turns when the provider requires it.
type Reasoning struct {
	Text      string
	Redacted  bool
	Signature string
}

// ToolCall is a request from the model to invoke a tool. Input is the raw
// JSON arguments so no fidelity is lost before validation.
type ToolCall struct {
	ID    string
	Name  string
	Input json.RawMessage
}

// ToolOutput is the result of a tool invocation, correlated to a ToolCall by
// CallID. IsError distinguishes a tool that failed from one that returned a
// negative result — providers render the two differently.
type ToolOutput struct {
	CallID  string
	Name    string
	Content []Part
	IsError bool
}

func (Text) isPart()       {}
func (Image) isPart()      {}
func (Audio) isPart()      {}
func (Reasoning) isPart()  {}
func (ToolCall) isPart()   {}
func (ToolOutput) isPart() {}

// Message is one turn in a conversation.
type Message struct {
	Role  Role
	Parts []Part
}

// System returns a system message carrying text.
func System(text string) Message {
	return Message{Role: RoleSystem, Parts: []Part{Text{Text: text}}}
}

// User returns a user message carrying text.
func User(text string) Message {
	return Message{Role: RoleUser, Parts: []Part{Text{Text: text}}}
}

// Assistant returns an assistant message carrying text.
func Assistant(text string) Message {
	return Message{Role: RoleAssistant, Parts: []Part{Text{Text: text}}}
}

// UserParts returns a user message with arbitrary content parts (multimodal).
func UserParts(parts ...Part) Message {
	return Message{Role: RoleUser, Parts: parts}
}

// ToolResult returns a tool-role message carrying the output of one call.
func ToolResult(callID, name, content string, isError bool) Message {
	return Message{Role: RoleTool, Parts: []Part{ToolOutput{
		CallID:  callID,
		Name:    name,
		Content: []Part{Text{Text: content}},
		IsError: isError,
	}}}
}

// Text concatenates every text part in the message. Reasoning parts are not
// included: they are model scratch space, not the answer.
func (m Message) Text() string {
	var b strings.Builder
	for _, p := range m.Parts {
		if t, ok := p.(Text); ok {
			b.WriteString(t.Text)
		}
	}
	return b.String()
}

// ToolCalls returns every tool call requested in this message.
func (m Message) ToolCalls() []ToolCall {
	var calls []ToolCall
	for _, p := range m.Parts {
		if c, ok := p.(ToolCall); ok {
			calls = append(calls, c)
		}
	}
	return calls
}
