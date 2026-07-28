package v1beta

import (
	"context"
	"strings"
	"testing"
)

type fakeCalcTool struct{ calls int }

func (t *fakeCalcTool) Name() string        { return "calculator" }
func (t *fakeCalcTool) Description() string { return "adds two numbers" }
func (t *fakeCalcTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	t.calls++
	return &ToolResult{Success: true, Content: 4}, nil
}

// Reproduces docs/v1beta/custom-handlers.md's own "Tool-Augmented" example
// contract: a HandlerFunc calling capabilities.Tools.Execute(ctx, name, args)
// directly. Before this fix, Capabilities.Tools was always nil at both
// construction sites in agent_impl.go (commented out) — any handler written
// exactly as the docs show would nil-pointer-panic on capabilities.Tools.Execute.
func TestSliceToolManager_ExecuteRunsTheRealTool(t *testing.T) {
	tool := &fakeCalcTool{}
	tm := newSliceToolManager([]Tool{tool})

	result, err := tm.Execute(context.Background(), "calculator", map[string]interface{}{"a": 2, "b": 2})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !result.Success || result.Content != 4 {
		t.Errorf("Execute result = %+v, want Success=true Content=4", result)
	}
	if tool.calls != 1 {
		t.Errorf("underlying tool called %d times, want 1", tool.calls)
	}
}

func TestSliceToolManager_ExecuteUnknownToolIsAnHonestError(t *testing.T) {
	tm := newSliceToolManager([]Tool{&fakeCalcTool{}})

	result, err := tm.Execute(context.Background(), "web_search", nil)
	if err == nil {
		t.Fatal("expected an error for a tool not in this agent's configured list")
	}
	if result.Success {
		t.Errorf("result.Success = true, want false: %+v", result)
	}
	if !strings.Contains(result.Error, "web_search") {
		t.Errorf("result.Error = %q, want it to name the missing tool", result.Error)
	}
}

func TestSliceToolManager_ListAvailableIsAvailable(t *testing.T) {
	tm := newSliceToolManager([]Tool{&fakeCalcTool{}})

	if !tm.IsAvailable("calculator") {
		t.Error("IsAvailable(calculator) = false, want true")
	}
	if tm.IsAvailable("web_search") {
		t.Error("IsAvailable(web_search) = true, want false")
	}
	avail := tm.Available()
	if len(avail) != 1 || avail[0] != "calculator" {
		t.Errorf("Available() = %v, want [calculator]", avail)
	}
	list := tm.List()
	if len(list) != 1 || list[0].Name != "calculator" || list[0].Description != "adds two numbers" {
		t.Errorf("List() = %+v, want one ToolInfo for calculator", list)
	}
}

// ConnectMCP/DisconnectMCP/DiscoverMCP must fail loudly, not silently
// no-op-succeed — this adapter wraps a fixed snapshot of tools, and a fake
// success here would repeat the swallowed-error shape that already caused a
// real production incident in this framework (stdio transport argv bug
// surfacing as "0 tools" instead of a build error).
func TestSliceToolManager_MCPOperationsFailLoudlyNotSilently(t *testing.T) {
	tm := newSliceToolManager([]Tool{&fakeCalcTool{}})
	ctx := context.Background()

	if err := tm.ConnectMCP(ctx); err == nil {
		t.Error("ConnectMCP: want an explicit error, got nil (silent no-op)")
	}
	if err := tm.DisconnectMCP("any"); err == nil {
		t.Error("DisconnectMCP: want an explicit error, got nil (silent no-op)")
	}
	if _, err := tm.DiscoverMCP(ctx); err == nil {
		t.Error("DiscoverMCP: want an explicit error, got nil (silent no-op)")
	}
}
