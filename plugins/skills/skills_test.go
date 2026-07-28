package skills

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1beta "github.com/agenticgokit/agenticgokit/v1beta"
)

func writeSkill(t *testing.T, root, name, frontmatter, body string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\n" + frontmatter + "\n---\n" + body
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadCatalog_IndexesFrontmatterOnly(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "kantar-segmentation-q-query",
		"title: Kantar Shopping Mission Segmentation\ntrigger: User asks for Kantar segmentation\ntags:\n  - retail\n  - segmentation",
		"## Concept\nLong body that should NOT be in the catalog render.")

	cat, err := LoadCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Entries) != 1 {
		t.Fatalf("Entries = %d, want 1", len(cat.Entries))
	}
	e := cat.Entries[0]
	if e.Name != "kantar-segmentation-q-query" {
		t.Errorf("Name = %q", e.Name)
	}
	if e.Trigger != "User asks for Kantar segmentation" {
		t.Errorf("Trigger = %q", e.Trigger)
	}
	if len(e.Tags) != 2 || e.Tags[0] != "retail" {
		t.Errorf("Tags = %v", e.Tags)
	}

	render := cat.Render()
	if !strings.Contains(render, "kantar-segmentation-q-query") {
		t.Errorf("Render() missing skill name: %q", render)
	}
	if strings.Contains(render, "Long body") {
		t.Errorf("Render() leaked the body, want frontmatter-only: %q", render)
	}
}

func TestCatalog_LoadBody_RejectsPathTraversal(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "safe-skill", "title: Safe", "body text")

	cat, err := LoadCatalog(root)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := cat.LoadBody("../../etc/passwd"); err == nil {
		t.Error("LoadBody accepted a path-traversal name, want rejection")
	}
	if _, err := cat.LoadBody("does-not-exist"); err == nil {
		t.Error("LoadBody accepted an unknown skill name, want rejection")
	}
	body, err := cat.LoadBody("safe-skill")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "body text") {
		t.Errorf("LoadBody = %q, want the real body", body)
	}
}

// Reproduces the real 2026-07-16 production incident this whole plugin
// generalizes: Planning loaded a skill, then re-requested the SAME skill
// two iterations later because a hand-rolled loop's currentInput had
// silently dropped the body. As a REAL v1beta.Tool, load_skill dispatches
// through v1beta.ExecuteToolByName like any other registered tool — an app
// no longer needs its own bespoke interception/dedup machinery just to call
// it twice safely; both calls simply return the real body, deterministically,
// straight from disk.
func TestInstall_RegistersARealToolReachableThroughExecuteToolByName(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "kantar-segmentation-q-query",
		"trigger: Kantar segmentation questions",
		"Look for a dimension matching shopping mission / trip purpose.")

	render, err := Install(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(render, "kantar-segmentation-q-query") {
		t.Fatalf("Install's rendered catalog missing the skill: %q", render)
	}

	tools, err := v1beta.DiscoverInternalTools()
	if err != nil {
		t.Fatal(err)
	}
	var tool v1beta.Tool
	for _, tl := range tools {
		if tl.Name() == "load_skill" {
			tool = tl
			break
		}
	}
	if tool == nil {
		t.Fatal("load_skill not found among DiscoverInternalTools() results")
	}

	// Call it twice, exactly the sequence that bit the hand-rolled version —
	// both must return the real body since there's no ephemeral currentInput
	// string in between to lose it from.
	for i := 0; i < 2; i++ {
		result, err := tool.Execute(context.Background(), map[string]interface{}{"name": "kantar-segmentation-q-query"})
		if err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
		body, _ := result.Content.(string)
		if !strings.Contains(body, "shopping mission") {
			t.Errorf("call %d: body = %q, want the real skill content", i, body)
		}
	}

	if ws, ok := tool.(v1beta.ToolWithSchema); ok {
		schema := ws.JSONSchema()
		if schema["type"] != "object" {
			t.Errorf("JSONSchema() type = %v, want object", schema["type"])
		}
	} else {
		t.Error("load_skill tool does not implement ToolWithSchema — native tool-calling models get no schema")
	}
}

func TestLoadCatalog_EmptyDirIsNotAnError(t *testing.T) {
	cat, err := LoadCatalog("")
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Entries) != 0 {
		t.Errorf("Entries = %v, want empty", cat.Entries)
	}
	if cat.Render() != "" {
		t.Errorf("Render() = %q, want empty for an empty catalog", cat.Render())
	}
}

// Confirms this package's init() self-registers the declarative path
// (v1beta.SetSkillsProviderFactory) purely by being imported — a caller
// only has to set Config.Tools.Skills.Dir (TOML-loadable) and blank-import
// this package; no Install() call needed. v1beta.GetSkillsProviderFactory
// is the only external signal for "did init() actually run" short of a
// live Build() call — see its doc comment.
func TestInit_SelfRegistersDeclarativeFactory(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "declarative-skill", "trigger: test", "declarative body")

	factory := v1beta.GetSkillsProviderFactory()
	if factory == nil {
		t.Fatal("v1beta.GetSkillsProviderFactory() = nil — this package's init() did not self-register")
	}

	provider, err := factory(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(provider.Catalog(), "declarative-skill") {
		t.Errorf("Catalog() = %q, want it to list declarative-skill", provider.Catalog())
	}
	result, err := provider.Tool().Execute(context.Background(), map[string]interface{}{"name": "declarative-skill"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Content.(string), "declarative body") {
		t.Errorf("Tool().Execute content = %v, want the real body", result.Content)
	}
}
