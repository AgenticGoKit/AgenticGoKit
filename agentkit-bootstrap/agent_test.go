package agentkit_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/agenticgokit/agentkit"
	"github.com/agenticgokit/agentkit/agenttest"
)

type deps struct {
	Caller string
}

type searchArgs struct {
	Query string `json:"query" jsonschema:"description=Search terms"`
	Limit int    `json:"limit,omitempty"`
}

type hit struct {
	Title string `json:"title"`
}

type verdict struct {
	Score  int    `json:"score" jsonschema:"description=0 to 100"`
	Reason string `json:"reason"`
}

func TestRunTextOutput(t *testing.T) {
	p := agenttest.NewScript(agenttest.Text("hello there"))

	agent, err := agentkit.New[deps, string](p, "test-model",
		agentkit.WithSystem("be brief"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := agent.Run(context.Background(), "hi", deps{Caller: "alice"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Output != "hello there" {
		t.Errorf("Output = %q, want %q", res.Output, "hello there")
	}
	if res.Steps != 1 {
		t.Errorf("Steps = %d, want 1", res.Steps)
	}
	if res.Usage.Total() == 0 {
		t.Error("Usage not accumulated")
	}

	req, _ := p.LastRequest()
	if len(req.Messages) != 2 || req.Messages[0].Role != agentkit.RoleSystem {
		t.Errorf("expected system+user messages, got %d", len(req.Messages))
	}
}

func TestRunExecutesToolAndFeedsResultBack(t *testing.T) {
	var gotArgs searchArgs
	var gotCaller string

	tool := agentkit.NewTool("search", "Search the web",
		func(ctx context.Context, in searchArgs) ([]hit, error) {
			gotArgs = in
			if d, ok := agentkit.DepsFrom[deps](ctx); ok {
				gotCaller = d.Caller
			}
			return []hit{{Title: "result one"}}, nil
		}, agentkit.ReadOnly())

	p := agenttest.NewScript(
		agenttest.Calls(agenttest.Call("c1", "search", searchArgs{Query: "go generics", Limit: 5})),
		agenttest.Text("found it"),
	)

	agent, err := agentkit.New[deps, string](p, "test-model", agentkit.WithTools(tool))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := agent.Run(context.Background(), "search for me", deps{Caller: "alice"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if gotArgs.Query != "go generics" || gotArgs.Limit != 5 {
		t.Errorf("tool received %+v, want query/limit from the call", gotArgs)
	}
	// Deps must reach tools without globals.
	if gotCaller != "alice" {
		t.Errorf("DepsFrom in tool = %q, want %q", gotCaller, "alice")
	}
	if res.Output != "found it" {
		t.Errorf("Output = %q", res.Output)
	}
	if res.Steps != 2 {
		t.Errorf("Steps = %d, want 2", res.Steps)
	}

	// The second request must carry the tool result back to the model.
	reqs := p.Requests()
	if len(reqs) != 2 {
		t.Fatalf("expected 2 model calls, got %d", len(reqs))
	}
	var sawToolResult bool
	for _, m := range reqs[1].Messages {
		if m.Role == agentkit.RoleTool {
			sawToolResult = true
		}
	}
	if !sawToolResult {
		t.Error("second request did not include the tool result")
	}
}

func TestStructuredOutput(t *testing.T) {
	p := agenttest.NewScript(
		agenttest.Calls(agenttest.Call("c1", "final_output", verdict{Score: 87, Reason: "solid"})),
	)

	agent, err := agentkit.New[deps, verdict](p, "test-model")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := agent.Run(context.Background(), "review this", deps{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Output.Score != 87 || res.Output.Reason != "solid" {
		t.Errorf("Output = %+v, want the scripted verdict", res.Output)
	}

	// The output schema must be advertised to the model.
	req, _ := p.LastRequest()
	var found bool
	for _, d := range req.Tools {
		if d.Name == "final_output" {
			found = true
			if d.InputSchema == nil || d.InputSchema.Properties["score"] == nil {
				t.Error("final_output schema missing derived properties")
			}
		}
	}
	if !found {
		t.Error("final_output tool was not offered to the model")
	}
}

func TestInvalidToolArgsAreReturnedToModelForCorrection(t *testing.T) {
	calls := 0
	tool := agentkit.NewTool("search", "Search",
		func(ctx context.Context, in searchArgs) ([]hit, error) {
			calls++
			return nil, nil
		})

	// First call sends a string where an object is required.
	bad := agentkit.ToolCall{ID: "c1", Name: "search", Input: json.RawMessage(`"not an object"`)}
	p := agenttest.NewScript(
		agenttest.Calls(bad),
		agenttest.Text("recovered"),
	)

	agent, _ := agentkit.New[deps, string](p, "test-model", agentkit.WithTools(tool))
	res, err := agent.Run(context.Background(), "go", deps{})
	if err != nil {
		t.Fatalf("Run should recover from bad args, got: %v", err)
	}
	if calls != 0 {
		t.Errorf("tool body ran %d times with invalid args, want 0", calls)
	}
	if res.Output != "recovered" {
		t.Errorf("Output = %q", res.Output)
	}

	// The correction must have been sent back as an error tool result.
	reqs := p.Requests()
	var sawError bool
	for _, m := range reqs[len(reqs)-1].Messages {
		for _, part := range m.Parts {
			if to, ok := part.(agentkit.ToolOutput); ok && to.IsError {
				sawError = true
			}
		}
	}
	if !sawError {
		t.Error("validation failure was not returned to the model")
	}
}

func TestRetryErrorFromToolIsSelfCorrecting(t *testing.T) {
	tool := agentkit.NewTool("lookup", "Look up a person",
		func(ctx context.Context, in searchArgs) (string, error) {
			if in.Query == "bob" {
				return "", agentkit.Retry("use a full name, not %q", in.Query)
			}
			return "Bob Smith, engineer", nil
		})

	p := agenttest.NewScript(
		agenttest.Calls(agenttest.Call("c1", "lookup", searchArgs{Query: "bob"})),
		agenttest.Calls(agenttest.Call("c2", "lookup", searchArgs{Query: "bob smith"})),
		agenttest.Text("done"),
	)

	agent, _ := agentkit.New[deps, string](p, "test-model", agentkit.WithTools(tool))
	res, err := agent.Run(context.Background(), "look up bob", deps{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Steps != 3 {
		t.Errorf("Steps = %d, want 3 (retry loop)", res.Steps)
	}
}

func TestUnknownToolDoesNotAbortRun(t *testing.T) {
	p := agenttest.NewScript(
		agenttest.Calls(agenttest.Call("c1", "nonexistent", map[string]string{})),
		agenttest.Text("ok"),
	)
	agent, _ := agentkit.New[deps, string](p, "test-model")

	res, err := agent.Run(context.Background(), "go", deps{})
	if err != nil {
		t.Fatalf("hallucinated tool name should be correctable, got: %v", err)
	}
	if res.Output != "ok" {
		t.Errorf("Output = %q", res.Output)
	}
}

func TestMaxStepsIsEnforced(t *testing.T) {
	tool := agentkit.NewTool("loop", "Loops forever",
		func(ctx context.Context, in searchArgs) (string, error) { return "again", nil })

	steps := make([]agentkit.Response, 10)
	for i := range steps {
		steps[i] = agenttest.Calls(agenttest.Call("c", "loop", searchArgs{Query: "x"}))
	}
	p := agenttest.NewScript(steps...)

	agent, _ := agentkit.New[deps, string](p, "test-model",
		agentkit.WithTools(tool), agentkit.WithMaxSteps(3))

	_, err := agent.Run(context.Background(), "go", deps{})
	if !errors.Is(err, agentkit.ErrMaxSteps) {
		t.Errorf("err = %v, want ErrMaxSteps", err)
	}
}

func TestUsageLimitStopsRun(t *testing.T) {
	p := agenttest.NewScript(agenttest.Text("hi"))
	agent, _ := agentkit.New[deps, string](p, "test-model",
		agentkit.WithUsageLimits(agentkit.UsageLimits{MaxTotalTokens: 1}))

	_, err := agent.Run(context.Background(), "go", deps{})
	if !errors.Is(err, agentkit.ErrUsageLimit) {
		t.Errorf("err = %v, want ErrUsageLimit", err)
	}
}

func TestProviderErrorsWrapSentinelAndStep(t *testing.T) {
	p := agenttest.NewScript() // exhausted immediately
	agent, _ := agentkit.New[deps, string](p, "test-model")

	_, err := agent.Run(context.Background(), "go", deps{})
	if !errors.Is(err, agentkit.ErrProvider) {
		t.Errorf("err = %v, want it to wrap ErrProvider", err)
	}
	var se *agentkit.StepError
	if !errors.As(err, &se) {
		t.Errorf("err = %v, want a *StepError identifying the failing step", err)
	}
}

func TestParamsCarryExplicitZeroTemperature(t *testing.T) {
	p := agenttest.NewScript(agenttest.Text("ok"))
	zero := float32(0)
	agent, _ := agentkit.New[deps, string](p, "test-model",
		agentkit.WithParams(agentkit.Params{Temperature: &zero}))

	if _, err := agent.Run(context.Background(), "go", deps{}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	req, _ := p.LastRequest()
	if req.Params.Temperature == nil {
		t.Fatal("Temperature was dropped; explicit 0 must reach the provider")
	}
	if *req.Params.Temperature != 0 {
		t.Errorf("Temperature = %v, want explicit 0", *req.Params.Temperature)
	}
}

func TestCapabilityGapProducesDiagnostic(t *testing.T) {
	p := agenttest.NewScript(agenttest.Text("ok"))
	p.Caps = &agentkit.Capabilities{NativeTools: false, Streaming: true}

	tool := agentkit.NewTool("search", "Search",
		func(ctx context.Context, in searchArgs) (string, error) { return "", nil })

	var handled []agentkit.Diagnostic
	agent, err := agentkit.New[deps, string](p, "weak-model",
		agentkit.WithTools(tool),
		agentkit.WithDiagnosticHandler(func(d agentkit.Diagnostic) { handled = append(handled, d) }))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if len(agent.Diagnostics()) == 0 {
		t.Fatal("tools on a non-tool-calling model produced no diagnostic")
	}
	if len(handled) == 0 {
		t.Error("diagnostic handler was not called")
	}
	if agent.Diagnostics()[0].Code != agentkit.DiagNoNativeTools {
		t.Errorf("code = %s, want %s", agent.Diagnostics()[0].Code, agentkit.DiagNoNativeTools)
	}
}

func TestCleanConfigProducesNoDiagnostics(t *testing.T) {
	p := agenttest.NewScript(agenttest.Text("ok"))
	agent, _ := agentkit.New[deps, string](p, "good-model")
	if d := agent.Diagnostics(); len(d) != 0 {
		t.Errorf("clean config produced diagnostics: %+v", d)
	}
}

func TestDuplicateToolIsRejectedAtBuild(t *testing.T) {
	mk := func() agentkit.Tool {
		return agentkit.NewTool("search", "Search",
			func(ctx context.Context, in searchArgs) (string, error) { return "", nil })
	}
	p := agenttest.NewScript()
	_, err := agentkit.New[deps, string](p, "m", agentkit.WithTools(mk(), mk()))
	if !errors.Is(err, agentkit.ErrInvalidConfig) {
		t.Errorf("err = %v, want ErrInvalidConfig for a duplicate tool name", err)
	}
	if err != nil && !strings.Contains(err.Error(), "search") {
		t.Errorf("error should name the duplicated tool, got: %v", err)
	}
}

func TestStreamReplaysEvents(t *testing.T) {
	p := agenttest.NewScript(agenttest.Text("streamed"))
	s, err := p.Stream(context.Background(), agentkit.Request{Model: "m"})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer s.Close()

	var text strings.Builder
	var done bool
	for s.Next() {
		switch e := s.Event().(type) {
		case agentkit.TextDelta:
			text.WriteString(e.Text)
		case agentkit.Done:
			done = true
		}
	}
	if err := s.Err(); err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if text.String() != "streamed" {
		t.Errorf("streamed text = %q", text.String())
	}
	if !done {
		t.Error("stream ended without a Done event")
	}
}
