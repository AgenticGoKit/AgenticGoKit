package agentkit

import "context"

// Provider is the only interface a model backend must implement. It is
// deliberately small so third parties can implement it without depending on
// framework internals, and so cross-cutting concerns (retry, rate limiting,
// cost accounting, fallback) can be written once as decorators.
type Provider interface {
	// Name identifies the provider for diagnostics and telemetry
	// (e.g. "openai", "anthropic", "ollama").
	Name() string

	// Generate runs one non-streaming model call.
	Generate(ctx context.Context, req Request) (*Response, error)

	// Stream runs one streaming model call. Implementations must deliver
	// terminal errors through Stream.Err, never by silently ending.
	Stream(ctx context.Context, req Request) (Stream, error)

	// Capabilities reports what the named model supports. Callers branch on
	// this instead of discovering unsupported features by silent degradation.
	Capabilities(model string) Capabilities
}

// Embedder is implemented separately from Provider: many chat providers have
// no embeddings API, and forcing them to stub one produced zero-vector
// embeddings in a previous generation of this framework.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	// Dimensions reports the vector width, so stores can be provisioned and
	// mismatches detected before the first write.
	Dimensions() int
}

// Request is one model call.
type Request struct {
	Model      string
	Messages   []Message
	Tools      []ToolDef
	ToolChoice ToolChoice
	Params     Params

	// Output requests schema-constrained output. Nil means free-form text.
	Output *OutputSpec
}

// Params holds sampling parameters. Every field is a pointer so "unset" is
// distinguishable from a meaningful zero — Temperature 0 (deterministic) is a
// legitimate setting, not an absent one.
type Params struct {
	Temperature   *float32
	TopP          *float32
	MaxTokens     *int
	StopSequences []string
	Seed          *int
}

// ToolChoiceMode controls whether the model may, must, or must not call tools.
type ToolChoiceMode string

const (
	// ToolChoiceAuto lets the model decide (default).
	ToolChoiceAuto ToolChoiceMode = "auto"
	// ToolChoiceNone forbids tool calls for this turn.
	ToolChoiceNone ToolChoiceMode = "none"
	// ToolChoiceRequired forces at least one tool call.
	ToolChoiceRequired ToolChoiceMode = "required"
	// ToolChoiceTool forces the specific tool named in ToolChoice.Name.
	ToolChoiceTool ToolChoiceMode = "tool"
)

// ToolChoice constrains tool selection for one request.
type ToolChoice struct {
	Mode ToolChoiceMode
	Name string
}

// OutputFormat selects how schema-constrained output is obtained. Declaring
// the desired type and choosing the mechanism are separate decisions, because
// the mechanism depends on model capability while the type does not.
type OutputFormat string

const (
	// OutputTool passes the schema as a tool the model calls. Works on nearly
	// every tool-capable model.
	OutputTool OutputFormat = "tool"
	// OutputNative uses the provider's JSON-schema response format.
	OutputNative OutputFormat = "native"
	// OutputPrompted injects the schema into the prompt. Last resort for
	// models with neither capability.
	OutputPrompted OutputFormat = "prompted"
)

// OutputSpec describes schema-constrained output for a request.
type OutputSpec struct {
	Format OutputFormat
	Name   string
	Schema []byte
}

// StopReason explains why generation ended.
type StopReason string

const (
	StopEndTurn   StopReason = "end_turn"
	StopToolUse   StopReason = "tool_use"
	StopMaxTokens StopReason = "max_tokens"
	StopStop      StopReason = "stop_sequence"
	StopFiltered  StopReason = "content_filter"
)

// Usage reports token consumption for one call. Streaming providers must
// populate this on the terminal event; cost accounting depends on it.
type Usage struct {
	InputTokens       int
	OutputTokens      int
	CachedInputTokens int
	ReasoningTokens   int
}

// Add accumulates usage across the calls of a run.
func (u *Usage) Add(o Usage) {
	u.InputTokens += o.InputTokens
	u.OutputTokens += o.OutputTokens
	u.CachedInputTokens += o.CachedInputTokens
	u.ReasoningTokens += o.ReasoningTokens
}

// Total returns all tokens counted for this usage.
func (u Usage) Total() int { return u.InputTokens + u.OutputTokens }

// Response is the result of a non-streaming call.
type Response struct {
	Message    Message
	StopReason StopReason
	Usage      Usage
	// Model is the model that actually served the request, which may differ
	// from the requested alias.
	Model string
}

// Modality is an input or output content type a model supports.
type Modality string

const (
	ModalityText  Modality = "text"
	ModalityImage Modality = "image"
	ModalityAudio Modality = "audio"
	ModalityVideo Modality = "video"
)

// Capabilities describes what a model supports, so the runtime can choose a
// strategy instead of failing silently.
type Capabilities struct {
	NativeTools     bool
	ParallelTools   bool
	NativeJSON      bool
	Streaming       bool
	PromptCaching   bool
	Reasoning       bool
	Input           []Modality
	ContextWindow   int
	MaxOutputTokens int
}

// Supports reports whether the model accepts the given input modality.
func (c Capabilities) Supports(m Modality) bool {
	for _, x := range c.Input {
		if x == m {
			return true
		}
	}
	return false
}

// Event is one item in a stream. The closed set mirrors what modern chat UIs
// need to render incrementally.
type Event interface{ isEvent() }

// TextDelta is an incremental chunk of assistant text.
type TextDelta struct{ Text string }

// ReasoningDelta is an incremental chunk of model reasoning.
type ReasoningDelta struct{ Text string }

// ToolCallStart announces a tool call whose arguments will stream.
type ToolCallStart struct {
	ID   string
	Name string
}

// ToolCallDelta is an incremental chunk of a tool call's JSON arguments.
type ToolCallDelta struct {
	ID        string
	InputJSON string
}

// ToolCallDone marks a tool call's arguments complete.
type ToolCallDone struct {
	ID    string
	Input []byte
}

// Done is the terminal event, carrying the accounting a caller needs.
type Done struct {
	StopReason StopReason
	Usage      Usage
}

func (TextDelta) isEvent()      {}
func (ReasoningDelta) isEvent() {}
func (ToolCallStart) isEvent()  {}
func (ToolCallDelta) isEvent()  {}
func (ToolCallDone) isEvent()   {}
func (Done) isEvent()           {}

// Stream is a cursor over stream events. The cursor shape (rather than a bare
// channel) exists so a failed stream cannot be mistaken for a successful one:
// Next returns false on both completion and failure, and Err is the single
// source of truth for which occurred.
type Stream interface {
	// Next advances to the next event, returning false at end of stream or on
	// error. Callers must consult Err after Next returns false.
	Next() bool
	// Event returns the event Next advanced to.
	Event() Event
	// Err returns the terminal error, or nil if the stream completed.
	Err() error
	// Close releases resources and cancels any in-flight request. It is safe
	// to call more than once.
	Close() error
}
