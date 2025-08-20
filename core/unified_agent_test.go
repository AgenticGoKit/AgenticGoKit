package core

import (
	"context"
	"testing"
	"time"
)

// dummyCapability implements AgentCapability for testing capability listing
type dummyCapability struct{ name string }

func (d *dummyCapability) Name() string                                 { return d.name }
func (d *dummyCapability) Configure(agent CapabilityConfigurable) error { return nil }
func (d *dummyCapability) Validate(_ []AgentCapability) error           { return nil }
func (d *dummyCapability) Priority() int                                { return 10 }

// handler that marks state and returns it
type testHandler struct{}

func (h testHandler) Run(ctx context.Context, event Event, state State) (AgentResult, error) {
	state.Set("ran_with_handler", true)
	return AgentResult{OutputState: state}, nil
}

func TestUnifiedAgent_Run_Default(t *testing.T) {
	ua := NewUnifiedAgent("ua", nil, nil)
	if ua == nil {
		t.Fatalf("expected non-nil UnifiedAgent")
	}

	if ua.GetTimeout() <= 0 {
		t.Fatalf("expected default timeout > 0, got %v", ua.GetTimeout())
	}
	if !ua.IsEnabled() {
		t.Fatalf("expected agent to be enabled by default")
	}

	st := NewState()
	out, err := ua.Run(context.Background(), st)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if out == nil {
		t.Fatalf("expected non-nil output state")
	}
	if v, ok := out.Get("processed_by"); !ok || v != "ua" {
		t.Fatalf("expected processed_by to be 'ua', got %v (ok=%v)", v, ok)
	}
	if v, ok := out.Get("agent_type"); !ok || v != "unified" {
		t.Fatalf("expected agent_type to be 'unified', got %v (ok=%v)", v, ok)
	}
}

func TestUnifiedAgent_HandleEvent_WithHandler(t *testing.T) {
	ua := NewUnifiedAgent("ua", nil, testHandler{})

	st := NewState()
	res, err := ua.HandleEvent(context.Background(), NewEvent("ua", nil, nil), st)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	if res.OutputState == nil {
		t.Fatalf("expected non-nil OutputState")
	}
	if v, ok := res.OutputState.Get("ran_with_handler"); !ok || v != true {
		t.Fatalf("expected ran_with_handler=true, got %v (ok=%v)", v, ok)
	}
	if res.Duration < 0 || res.EndTime.Before(res.StartTime) {
		t.Fatalf("invalid timing in AgentResult: start=%v end=%v dur=%v", res.StartTime, res.EndTime, res.Duration)
	}
}

func TestUnifiedAgent_Capabilities_And_LLMConfig(t *testing.T) {
	caps := map[CapabilityType]AgentCapability{
		CapabilityTypeLLM: &dummyCapability{name: "llm"},
		CapabilityTypeMCP: &dummyCapability{name: "mcp"},
	}
	ua := NewUnifiedAgent("ua", caps, nil)

	// Capabilities list contains our keys (string values)
	gotCaps := ua.GetCapabilities()
	if len(gotCaps) != 2 {
		t.Fatalf("expected 2 capabilities, got %d: %v", len(gotCaps), gotCaps)
	}

	// Map LLM config via SetLLMProvider
	cfg := AgentLLMConfig{
		Provider:       "openai",
		Model:          "gpt-4o-mini",
		Temperature:    0.5,
		MaxTokens:      1024,
		TimeoutSeconds: 15,
		TopP:           0.9,
	}
	ua.SetLLMProvider(nil, cfg)
	r := ua.GetLLMConfig()
	if r == nil {
		t.Fatalf("expected resolved LLM config, got nil")
	}
	if r.Provider != cfg.Provider || r.Model != cfg.Model {
		t.Fatalf("provider/model mismatch: %+v vs %+v", r, cfg)
	}
	if r.Temperature != cfg.Temperature || r.MaxTokens != cfg.MaxTokens {
		t.Fatalf("temperature/tokens mismatch: %+v vs %+v", r, cfg)
	}
	expectedTimeout := TimeoutFromSeconds(cfg.TimeoutSeconds)
	if r.Timeout != expectedTimeout {
		t.Fatalf("timeout mismatch: got %v want %v", r.Timeout, expectedTimeout)
	}
}

func TestUnifiedAgentBuilder_DefaultsAndCustomizations(t *testing.T) {
	// Defaults when only name is set
	a, err := NewUnifiedAgentBuilder("builder-agent").Build()
	if err != nil {
		t.Fatalf("unexpected error building agent: %v", err)
	}
	ua, ok := a.(*UnifiedAgent)
	if !ok {
		t.Fatalf("expected UnifiedAgent type")
	}
	if ua.Name() != "builder-agent" {
		t.Fatalf("name mismatch: %s", ua.Name())
	}
	if ua.GetTimeout() <= 0 || !ua.IsEnabled() {
		t.Fatalf("expected sensible defaults for timeout and enabled")
	}

	// Customizations
	cfg := AgentLLMConfig{Provider: "openai", Model: "gpt-4o-mini", MaxTokens: 777, Temperature: 0.2, TimeoutSeconds: 5}
	enabled := false
	a2, err := NewUnifiedAgentBuilder("custom").
		WithRole("research").
		WithDescription("desc").
		WithSystemPrompt("sys").
		WithTimeout(42 * time.Second).
		Enabled(enabled).
		WithLLMConfig(cfg).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ua2 := a2.(*UnifiedAgent)
	if ua2.GetRole() != "research" || ua2.GetDescription() != "desc" || ua2.GetSystemPrompt() != "sys" {
		t.Fatalf("role/description/systemPrompt not applied")
	}
	if ua2.GetTimeout() != 42*time.Second || ua2.IsEnabled() != false {
		t.Fatalf("timeout/enabled not applied: %v %v", ua2.GetTimeout(), ua2.IsEnabled())
	}
	if llm := ua2.GetLLMConfig(); llm == nil || llm.Model != "gpt-4o-mini" || llm.MaxTokens != 777 {
		t.Fatalf("llm config not applied: %+v", llm)
	}
}
