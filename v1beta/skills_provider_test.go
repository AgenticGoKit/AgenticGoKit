package v1beta

import (
	"context"
	"strings"
	"testing"
)

type fakeSkillsTool struct{ name string }

func (t *fakeSkillsTool) Name() string        { return t.name }
func (t *fakeSkillsTool) Description() string { return "test skill loader" }
func (t *fakeSkillsTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	return &ToolResult{Success: true, Content: "skill body"}, nil
}

type fakeSkillsProvider struct{ dir string }

func (p *fakeSkillsProvider) Tool() Tool      { return &fakeSkillsTool{name: "load_skill"} }
func (p *fakeSkillsProvider) Catalog() string { return "- fake-skill — a test skill\n" }

// Verifies the declarative path end to end: ToolsConfig.Skills.Dir (the
// TOML-loadable field) plus a registered SkillsProviderFactory (what a
// blank-imported skills plugin, e.g. plugins/skills, does in its own
// init()) is enough for Build() to both append the load_skill tool to
// a.tools AND splice the rendered catalog into a.config.SystemPrompt —
// no Install()-style Go call needed by the app at all.
func TestBuild_WiresSkillsProviderFromToolsConfig(t *testing.T) {
	SetSkillsProviderFactory(func(dir string) (SkillsProvider, error) {
		if dir != "/fake/skills/dir" {
			t.Errorf("factory called with dir = %q, want /fake/skills/dir", dir)
		}
		return &fakeSkillsProvider{dir: dir}, nil
	})
	defer SetSkillsProviderFactory(nil)

	cfg := &Config{
		Name:         "skills-probe",
		LLM:          LLMConfig{Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:1"},
		SystemPrompt: "Base prompt.",
		Tools: &ToolsConfig{
			Enabled: true,
			Skills:  &SkillsConfig{Dir: "/fake/skills/dir"},
		},
	}

	agentIface, err := newRealAgent(cfg, nil)
	if err != nil {
		t.Fatalf("newRealAgent error: %v", err)
	}
	agent := agentIface.(*realAgent)

	var found bool
	for _, tl := range agent.tools {
		if tl.Name() == "load_skill" {
			found = true
		}
	}
	if !found {
		t.Errorf("agent.tools = %v, want load_skill among them", agent.tools)
	}

	if !strings.Contains(agent.config.SystemPrompt, "fake-skill") {
		t.Errorf("SystemPrompt was not spliced with the rendered catalog: %q", agent.config.SystemPrompt)
	}
	if !strings.HasPrefix(agent.config.SystemPrompt, "Base prompt.") {
		t.Errorf("SystemPrompt lost its original content: %q", agent.config.SystemPrompt)
	}
}

// No factory registered: Dir set but nothing to load it — must not error
// Build() (a missing blank-import shouldn't be fatal, just inert, same
// fail-open shape as a bad/unreachable MCP server).
func TestBuild_SkillsDirSetWithNoFactoryIsNotFatal(t *testing.T) {
	SetSkillsProviderFactory(nil)

	cfg := &Config{
		Name:         "skills-probe-no-factory",
		LLM:          LLMConfig{Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:1"},
		SystemPrompt: "Base prompt.",
		Tools: &ToolsConfig{
			Enabled: true,
			Skills:  &SkillsConfig{Dir: "/fake/skills/dir"},
		},
	}

	agentIface, err := newRealAgent(cfg, nil)
	if err != nil {
		t.Fatalf("newRealAgent error: %v, want no error (fail-open)", err)
	}
	agent := agentIface.(*realAgent)
	if agent.config.SystemPrompt != "Base prompt." {
		t.Errorf("SystemPrompt = %q, want unchanged when no factory is registered", agent.config.SystemPrompt)
	}
}

// No Skills config at all: the common case, must be a complete no-op.
func TestBuild_NoSkillsConfigIsANoOp(t *testing.T) {
	SetSkillsProviderFactory(func(dir string) (SkillsProvider, error) {
		t.Fatal("factory should never be called when ToolsConfig.Skills is nil")
		return nil, nil
	})
	defer SetSkillsProviderFactory(nil)

	cfg := &Config{
		Name:         "skills-probe-no-config",
		LLM:          LLMConfig{Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:1"},
		SystemPrompt: "Base prompt.",
		Tools:        &ToolsConfig{Enabled: true},
	}

	agentIface, err := newRealAgent(cfg, nil)
	if err != nil {
		t.Fatalf("newRealAgent error: %v", err)
	}
	agent := agentIface.(*realAgent)
	if agent.config.SystemPrompt != "Base prompt." {
		t.Errorf("SystemPrompt = %q, want unchanged", agent.config.SystemPrompt)
	}
}
