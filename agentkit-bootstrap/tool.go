package agentkit

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolDef is the provider-neutral description of a callable tool. Adapters
// translate it to each provider's wire shape; callers never write provider
// JSON by hand.
type ToolDef struct {
	Name         string
	Description  string
	InputSchema  *Schema
	OutputSchema *Schema
	Annotations  ToolAnnotations
}

// ToolAnnotations carry behavioral hints used for approval gating and UI
// affordances. They mirror MCP's tool annotations so MCP-sourced tools map
// across without loss.
type ToolAnnotations struct {
	ReadOnly    bool
	Destructive bool
	Idempotent  bool
	OpenWorld   bool
}

// Tool is an executable capability offered to a model.
type Tool interface {
	Def() ToolDef
	// Invoke runs the tool with raw JSON arguments and returns raw JSON
	// output. Implementations validate their own input.
	Invoke(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
}

// typedTool adapts a typed Go function to the Tool interface.
type typedTool[In, Out any] struct {
	def ToolDef
	fn  func(ctx context.Context, in In) (Out, error)
}

// NewTool builds a Tool from a typed Go function. Input and output schemas are
// derived from In and Out, so the schema the model sees cannot drift from the
// function that runs.
//
//	type SearchArgs struct {
//	    Query string `json:"query" jsonschema:"description=Search terms"`
//	    Limit int    `json:"limit,omitempty"`
//	}
//
//	tool := agentkit.NewTool("search", "Search the web",
//	    func(ctx context.Context, a SearchArgs) ([]Hit, error) { ... })
func NewTool[In, Out any](name, description string, fn func(ctx context.Context, in In) (Out, error), opts ...ToolOption) Tool {
	def := ToolDef{
		Name:         name,
		Description:  description,
		InputSchema:  SchemaOf[In](),
		OutputSchema: SchemaOf[Out](),
	}
	for _, o := range opts {
		o(&def)
	}
	return &typedTool[In, Out]{def: def, fn: fn}
}

// ToolOption customizes a tool definition.
type ToolOption func(*ToolDef)

// ReadOnly marks a tool as having no side effects, which approval policies and
// caching layers can rely on.
func ReadOnly() ToolOption {
	return func(d *ToolDef) { d.Annotations.ReadOnly = true }
}

// Destructive marks a tool as performing irreversible actions, so approval
// policies can require confirmation.
func Destructive() ToolOption {
	return func(d *ToolDef) { d.Annotations.Destructive = true }
}

func (t *typedTool[In, Out]) Def() ToolDef { return t.def }

func (t *typedTool[In, Out]) Invoke(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var in In
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			// Argument errors are returned to the model for self-correction
			// rather than aborting the run.
			return nil, &RetryError{Message: fmt.Sprintf("invalid arguments for %q: %v", t.def.Name, err)}
		}
	}
	out, err := t.fn(ctx, in)
	if err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

// toolSet indexes tools by name and rejects duplicates, so a shadowed tool is
// a build-time error rather than a silent first-wins surprise.
type toolSet struct {
	order  []Tool
	byName map[string]Tool
}

func newToolSet() *toolSet { return &toolSet{byName: map[string]Tool{}} }

func (s *toolSet) add(t Tool) error {
	name := t.Def().Name
	if name == "" {
		return fmt.Errorf("%w: tool has no name", ErrInvalidConfig)
	}
	if _, dup := s.byName[name]; dup {
		return fmt.Errorf("%w: duplicate tool %q", ErrInvalidConfig, name)
	}
	s.byName[name] = t
	s.order = append(s.order, t)
	return nil
}

func (s *toolSet) defs() []ToolDef {
	defs := make([]ToolDef, 0, len(s.order))
	for _, t := range s.order {
		defs = append(defs, t.Def())
	}
	return defs
}

func (s *toolSet) get(name string) (Tool, bool) {
	t, ok := s.byName[name]
	return t, ok
}
