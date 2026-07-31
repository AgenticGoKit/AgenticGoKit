package agentkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// outputToolName is the synthetic tool through which structured output is
// requested when the model has no native JSON mode.
const outputToolName = "final_output"

// Agent is a typed agent. D is the dependency type threaded through a run and
// available to tools; O is the output type. Both are checked at compile time,
// so a refactor of either surfaces as a build error rather than a runtime
// surprise.
//
//	agent, err := agentkit.New[Deps, Verdict](provider, "gpt-4o",
//	    agentkit.WithSystem("You review code."),
//	    agentkit.WithTools(searchTool),
//	)
//	res, err := agent.Run(ctx, "Review PR 42", deps)
//	fmt.Println(res.Output.Score) // Verdict, no casting
type Agent[D any, O any] struct {
	provider Provider
	model    string

	system   string
	tools    *toolSet
	params   Params
	maxSteps int
	limits   UsageLimits

	structured   bool
	outputFormat OutputFormat
	outputSchema *Schema

	diagnostics []Diagnostic
}

// Result is the outcome of a run.
type Result[O any] struct {
	// Output is the typed result of the run.
	Output O
	// Messages is the full transcript, including tool calls and results.
	Messages []Message
	// Usage is the accumulated token usage across every model call.
	Usage Usage
	// Steps is the number of model calls made.
	Steps int
}

// UsageLimits bounds a run so a misbehaving loop cannot spend without limit.
type UsageLimits struct {
	MaxTotalTokens  int
	MaxOutputTokens int
}

type options struct {
	system   string
	tools    []Tool
	params   Params
	maxSteps int
	limits   UsageLimits
	format   OutputFormat
	onDiag   DiagnosticHandler
}

// Option configures an agent.
type Option func(*options)

// WithSystem sets the system instruction.
func WithSystem(s string) Option { return func(o *options) { o.system = s } }

// WithTools registers tools available to the model.
func WithTools(ts ...Tool) Option {
	return func(o *options) { o.tools = append(o.tools, ts...) }
}

// WithParams sets sampling parameters. Unset fields use the provider default;
// an explicit zero is honored.
func WithParams(p Params) Option { return func(o *options) { o.params = p } }

// WithMaxSteps bounds how many model calls one run may make. Default is 10.
func WithMaxSteps(n int) Option { return func(o *options) { o.maxSteps = n } }

// WithUsageLimits bounds token spend for a run.
func WithUsageLimits(l UsageLimits) Option { return func(o *options) { o.limits = l } }

// WithOutputFormat selects how structured output is obtained. Default is
// OutputTool, which works on nearly every tool-capable model.
func WithOutputFormat(f OutputFormat) Option { return func(o *options) { o.format = f } }

// WithDiagnosticHandler receives non-fatal findings produced while building
// the agent. Diagnostics are also readable afterwards via Diagnostics.
func WithDiagnosticHandler(h DiagnosticHandler) Option {
	return func(o *options) { o.onDiag = h }
}

// New builds a typed agent.
func New[D any, O any](p Provider, model string, opts ...Option) (*Agent[D, O], error) {
	if p == nil {
		return nil, fmt.Errorf("%w: provider is nil", ErrInvalidConfig)
	}
	if model == "" {
		return nil, fmt.Errorf("%w: model is empty", ErrInvalidConfig)
	}

	o := options{maxSteps: 10, format: OutputTool}
	for _, opt := range opts {
		opt(&o)
	}
	if o.maxSteps <= 0 {
		return nil, fmt.Errorf("%w: max steps must be positive, got %d", ErrInvalidConfig, o.maxSteps)
	}

	a := &Agent[D, O]{
		provider:     p,
		model:        model,
		system:       o.system,
		tools:        newToolSet(),
		params:       o.params,
		maxSteps:     o.maxSteps,
		limits:       o.limits,
		structured:   !isStringOutput[O](),
		outputFormat: o.format,
	}
	for _, t := range o.tools {
		if err := a.tools.add(t); err != nil {
			return nil, err
		}
	}
	if a.structured {
		a.outputSchema = SchemaOf[O]()
	}

	a.checkCapabilities(o.onDiag)
	return a, nil
}

// checkCapabilities compares what the agent needs against what the model
// supports, reporting degraded modes instead of discovering them at runtime.
func (a *Agent[D, O]) checkCapabilities(onDiag DiagnosticHandler) {
	caps := a.provider.Capabilities(a.model)

	if len(a.tools.order) > 0 && !caps.NativeTools {
		a.report(onDiag, Diagnostic{
			Severity: SevWarning,
			Code:     DiagNoNativeTools,
			Message: "model does not support native tool calling; tool results depend on the provider adapter's fallback strategy. " +
				"Choose a tool-capable model, or remove the tools.",
			Details: map[string]string{"provider": a.provider.Name(), "model": a.model},
		})
	}

	if a.structured && a.outputFormat == OutputNative && !caps.NativeJSON {
		a.outputFormat = OutputTool
		a.report(onDiag, Diagnostic{
			Severity: SevInfo,
			Code:     DiagNoNativeJSON,
			Message:  "model has no native JSON mode; falling back to tool-based structured output.",
			Details:  map[string]string{"provider": a.provider.Name(), "model": a.model},
		})
	}
}

func (a *Agent[D, O]) report(onDiag DiagnosticHandler, d Diagnostic) {
	a.diagnostics = append(a.diagnostics, d)
	if onDiag != nil {
		onDiag(d)
	}
}

// Diagnostics returns non-fatal findings collected while building this agent.
func (a *Agent[D, O]) Diagnostics() []Diagnostic { return a.diagnostics }

// Run executes the agent against input, with deps available to tools via
// DepsFrom.
func (a *Agent[D, O]) Run(ctx context.Context, input string, deps D) (*Result[O], error) {
	ctx = withDeps(ctx, deps)

	msgs := make([]Message, 0, 4)
	if a.system != "" {
		msgs = append(msgs, System(a.system))
	}
	msgs = append(msgs, User(input))

	var usage Usage

	for step := 1; step <= a.maxSteps; step++ {
		if err := ctx.Err(); err != nil {
			return nil, &StepError{Step: step, Phase: "generate", Err: err}
		}

		resp, err := a.provider.Generate(ctx, a.request(msgs))
		if err != nil {
			return nil, &StepError{Step: step, Phase: "generate", Err: fmt.Errorf("%w: %v", ErrProvider, err)}
		}
		usage.Add(resp.Usage)
		if err := a.limits.check(usage); err != nil {
			return nil, &StepError{Step: step, Phase: "limits", Err: err}
		}
		msgs = append(msgs, resp.Message)

		calls := resp.Message.ToolCalls()
		if len(calls) == 0 {
			if !a.structured {
				return &Result[O]{
					Output:   stringOutput[O](resp.Message.Text()),
					Messages: msgs,
					Usage:    usage,
					Steps:    step,
				}, nil
			}
			// Structured output was requested but the model answered in prose.
			// Ask once more, explicitly, rather than failing the run.
			msgs = append(msgs, User(fmt.Sprintf(
				"Respond by calling the %q tool with the required fields.", outputToolName)))
			continue
		}

		for _, c := range calls {
			if a.structured && c.Name == outputToolName {
				var out O
				if err := json.Unmarshal(c.Input, &out); err != nil {
					msgs = append(msgs, ToolResult(c.ID, c.Name,
						fmt.Sprintf("output did not match the schema: %v", err), true))
					continue
				}
				return &Result[O]{Output: out, Messages: msgs, Usage: usage, Steps: step}, nil
			}

			tool, ok := a.tools.get(c.Name)
			if !ok {
				// A hallucinated tool name is the model's mistake to correct,
				// not a reason to abort the run.
				msgs = append(msgs, ToolResult(c.ID, c.Name,
					fmt.Sprintf("%v: %q", ErrToolNotFound, c.Name), true))
				continue
			}

			out, err := tool.Invoke(ctx, c.Input)
			if err != nil {
				var retry *RetryError
				if errors.As(err, &retry) {
					msgs = append(msgs, ToolResult(c.ID, c.Name, retry.Message, true))
					continue
				}
				return nil, &StepError{Step: step, Phase: "tool", Err: &ToolError{Tool: c.Name, Err: err}}
			}
			msgs = append(msgs, ToolResult(c.ID, c.Name, string(out), false))
		}
	}

	return nil, fmt.Errorf("%w: %d steps without a final answer", ErrMaxSteps, a.maxSteps)
}

func (a *Agent[D, O]) request(msgs []Message) Request {
	req := Request{
		Model:    a.model,
		Messages: msgs,
		Tools:    a.tools.defs(),
		Params:   a.params,
	}
	if a.structured {
		switch a.outputFormat {
		case OutputNative:
			req.Output = &OutputSpec{
				Format: OutputNative,
				Name:   outputToolName,
				Schema: mustJSON(a.outputSchema),
			}
		default:
			req.Tools = append(req.Tools, ToolDef{
				Name:        outputToolName,
				Description: "Return the final answer in the required structure. Call this when you are done.",
				InputSchema: a.outputSchema,
			})
		}
	}
	return req
}

func (l UsageLimits) check(u Usage) error {
	if l.MaxTotalTokens > 0 && u.Total() > l.MaxTotalTokens {
		return fmt.Errorf("%w: %d total tokens exceeds limit %d", ErrUsageLimit, u.Total(), l.MaxTotalTokens)
	}
	if l.MaxOutputTokens > 0 && u.OutputTokens > l.MaxOutputTokens {
		return fmt.Errorf("%w: %d output tokens exceeds limit %d", ErrUsageLimit, u.OutputTokens, l.MaxOutputTokens)
	}
	return nil
}

// depsKey is unexported so only this package can put deps in a context.
type depsKey struct{}

func withDeps[D any](ctx context.Context, deps D) context.Context {
	return context.WithValue(ctx, depsKey{}, deps)
}

// DepsFrom retrieves the dependencies for the current run. Tools use it to
// reach the database handles, clients, and configuration a run was given,
// without globals.
func DepsFrom[D any](ctx context.Context) (D, bool) {
	d, ok := ctx.Value(depsKey{}).(D)
	return d, ok
}

// isStringOutput reports whether O is string, in which case the model's text
// is the answer and no output schema is needed.
func isStringOutput[O any]() bool {
	var zero O
	_, ok := any(zero).(string)
	return ok
}

// stringOutput converts text to O when O is string; otherwise the zero value.
func stringOutput[O any](text string) O {
	var out O
	if p, ok := any(&out).(*string); ok {
		*p = text
	}
	return out
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic("agentkit: marshal failed: " + err.Error())
	}
	return b
}
