// Package skills implements progressive-disclosure access to a directory of
// SKILL.md packages, as a real v1beta.Tool instead of an app hand-rolling its
// own pseudo-tool interception (the pattern this generalizes: a production
// AGK app built exactly this from scratch — SkillCatalog rendered into the
// system prompt at boot, bodies fetched on demand via a load_skill tool call
// — because no such primitive existed anywhere in the framework).
//
// Two ways to wire this in, both end up registering the same load_skill Tool:
//
//  1. Declarative: blank-import this package (`_ "…/plugins/skills"`, its
//     init() self-registers via v1beta.SetSkillsProviderFactory — same
//     convention as plugins/logging/zerolog) and set
//     Config.Tools.Skills.Dir (TOML-loadable: [tools.skills] dir = "..."),
//     matching how MCP servers are configured. Builder.Build() then loads
//     the catalog, appends the tool, and splices the rendered catalog into
//     SystemPrompt itself — nothing else to call.
//  2. Imperative: call Install(dir) directly at startup, before Build() —
//     for callers who need the rendered catalog string themselves (e.g. to
//     fold into a hand-built SystemPrompt via core.FormatPromptString,
//     before ever constructing a Config).
//
// A skill directory is a runtime value either way — there's no fixed name
// to register at import time the way a logging provider has a static
// string key, which is why (1) still needs an explicit Dir/Install call
// even though the registration itself is init()-based.
//
// Bodies are intentionally NOT pre-loaded into the catalog or the system
// prompt — only frontmatter (name/trigger/description/tags) is read at
// load time. A skill's full body is read from disk fresh on every
// load_skill call, keeping memory flat regardless of corpus size.
package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	v1beta "github.com/agenticgokit/agenticgokit/v1beta"
)

// Entry is one skill's boot-time index record. Body is deliberately not
// held here — see package doc.
type Entry struct {
	Name        string   // directory basename, or frontmatter "name" if set — canonical id passed to load_skill
	Title       string   // frontmatter title, falls back to Name
	Trigger     string   // frontmatter trigger, falls back to Description
	Description string   // frontmatter description
	Tags        []string // frontmatter tags
	Path        string   // absolute path to SKILL.md
}

// Catalog is the boot-time index of on-disk skills under one directory.
type Catalog struct {
	Dir     string
	Entries []Entry
	paths   map[string]string // name -> SKILL.md path, built once by LoadCatalog
}

// LoadCatalog walks dir for SKILL.md files (one per skill subdirectory) and
// indexes their frontmatter. Returns an empty catalog (nil error) for an
// empty dir; error only for real FS failures — matching WithMCP's own
// fail-open shape (a bad/missing skills dir shouldn't crash agent Build()).
func LoadCatalog(dir string) (Catalog, error) {
	cat := Catalog{Dir: dir}
	if dir == "" {
		return cat, nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		return cat, fmt.Errorf("skills: dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return cat, fmt.Errorf("skills: %q is not a directory", dir)
	}

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToUpper(filepath.Base(path)) != "SKILL.MD" {
			return nil
		}
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		fm, _ := splitFrontmatter(string(raw))
		name := filepath.Base(filepath.Dir(path))
		e := Entry{
			Name:        name,
			Title:       firstNonEmpty(yamlString(fm, "title"), name),
			Trigger:     yamlString(fm, "trigger"),
			Description: yamlString(fm, "description"),
			Tags:        yamlList(fm, "tags"),
			Path:        path,
		}
		if fmName := yamlString(fm, "name"); fmName != "" {
			e.Name = fmName
		}
		cat.Entries = append(cat.Entries, e)
		return nil
	})
	sort.Slice(cat.Entries, func(i, j int) bool { return cat.Entries[i].Name < cat.Entries[j].Name })
	cat.paths = make(map[string]string, len(cat.Entries))
	for _, e := range cat.Entries {
		cat.paths[e.Name] = e.Path
	}
	return cat, err
}

// Render returns compact markdown for injection into a system prompt: one
// line per skill (name, trigger, tags). Bodies excluded by design.
func (c Catalog) Render() string {
	if len(c.Entries) == 0 {
		return ""
	}
	var b strings.Builder
	for _, e := range c.Entries {
		hint := firstNonEmpty(e.Trigger, e.Description, e.Title)
		fmt.Fprintf(&b, "- %s — %s", e.Name, singleLine(hint, 240))
		if len(e.Tags) > 0 {
			fmt.Fprintf(&b, " [tags: %s]", strings.Join(e.Tags, ","))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// Names returns the name -> SKILL.md path index.
func (c Catalog) Names() map[string]string {
	if c.paths != nil {
		return c.paths
	}
	out := make(map[string]string, len(c.Entries))
	for _, e := range c.Entries {
		out[e.Name] = e.Path
	}
	return out
}

// LoadBody reads one skill's full SKILL.md body from disk on demand. Path
// traversal is rejected by requiring name to match a catalog entry — dot
// segments/absolute paths never resolve, since only base names are indexed.
func (c Catalog) LoadBody(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("skills: empty skill name")
	}
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		return "", fmt.Errorf("skills: invalid skill name %q", name)
	}
	path, ok := c.Names()[name]
	if !ok {
		return "", fmt.Errorf("skills: %q not in catalog", name)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("skills: read %q: %w", name, err)
	}
	return string(raw), nil
}

// loadSkillTool adapts a Catalog into a v1beta.Tool + v1beta.ToolWithSchema,
// so it dispatches through every real path a registered internal tool
// reaches: capabilities.Tools.Execute (via sliceToolManager), step 3.5's
// native/LLM-decided auto-tool-exec, and the package-level
// v1beta.ExecuteToolByName — instead of an app hand-rolling its own
// interception in its execution-loop code.
type loadSkillTool struct{ cat Catalog }

func (t *loadSkillTool) Name() string { return "load_skill" }

func (t *loadSkillTool) Description() string {
	return "Load the full body of one on-demand skill package by name. Only names listed in the skill catalog above are valid."
}

func (t *loadSkillTool) JSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Exact skill name from the catalog.",
			},
		},
		"required": []string{"name"},
	}
}

func (t *loadSkillTool) Execute(ctx context.Context, args map[string]interface{}) (*v1beta.ToolResult, error) {
	name, _ := args["name"].(string)
	body, err := t.cat.LoadBody(name)
	if err != nil {
		return &v1beta.ToolResult{Success: false, Error: err.Error()}, err
	}
	return &v1beta.ToolResult{Success: true, Content: body}, nil
}

// Install loads the skill catalog at dir, registers it as the "load_skill"
// internal tool (v1beta.RegisterInternalTool — the same hook
// DiscoverInternalTools/createTools already calls unconditionally as Step 1,
// no MCP required), and returns the catalog's rendered index for the caller
// to splice into its own SystemPrompt.
//
// Call once at startup, before Builder.Build() — v1beta.RegisterInternalTool
// writes into a package-level registry with no per-agent scoping, so
// Installing two different directories in the same process under the same
// tool name last-write-wins. One skills directory per process is the
// supported shape; namespace the tool name yourself (a second Install call
// with a different Name()) if you genuinely need more than one.
//
// The caller MUST also set ToolsConfig.Enabled = true (even with no MCP
// servers configured) — createTools() returns early otherwise and Step 1
// (DiscoverInternalTools) never runs. Confirmed via provider_factory.go;
// documented here since it's the one non-obvious prerequisite.
func Install(dir string) (catalogRender string, err error) {
	cat, err := LoadCatalog(dir)
	if err != nil {
		return "", err
	}
	v1beta.RegisterInternalTool("load_skill", func() v1beta.Tool {
		return &loadSkillTool{cat: cat}
	})
	return cat.Render(), nil
}

// catalogProvider adapts a Catalog into v1beta.SkillsProvider — the
// declarative path (ToolsConfig.Skills.Dir, TOML-loadable) below, as
// opposed to Install's direct/imperative path above. Both end up building
// the same loadSkillTool; this one also hands back the rendered catalog so
// Builder.Build() can splice it into SystemPrompt itself, since a TOML
// config has no Go call site to receive that string the way Install's
// caller does.
type catalogProvider struct{ cat Catalog }

func (p *catalogProvider) Tool() v1beta.Tool { return &loadSkillTool{cat: p.cat} }
func (p *catalogProvider) Catalog() string   { return p.cat.Render() }

// init registers this package as the skills provider for
// ToolsConfig.Skills.Dir (v1beta/skills_provider.go's SetSkillsProviderFactory
// hook) — same self-registration convention as this repo's other plugins
// (e.g. plugins/logging/zerolog's init()). A blank import
// (`_ "github.com/agenticgokit/agenticgokit/plugins/skills"`) plus setting
// Config.Tools.Skills.Dir in TOML or Go is then enough; no Install() call
// needed for this path. Install() itself still works unchanged for callers
// who prefer the direct/imperative style (e.g. when the catalog string is
// needed before Build() runs, to fold into a hand-built SystemPrompt).
func init() {
	v1beta.SetSkillsProviderFactory(func(dir string) (v1beta.SkillsProvider, error) {
		cat, err := LoadCatalog(dir)
		if err != nil {
			return nil, err
		}
		return &catalogProvider{cat: cat}, nil
	})
}

// --- minimal YAML frontmatter parser: enough for the SKILL.md dialect ---

func splitFrontmatter(s string) (string, string) {
	if !strings.HasPrefix(s, "---") {
		return "", s
	}
	rest := strings.TrimPrefix(s, "---")
	rest = strings.TrimLeft(rest, "\r\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", s
	}
	fm := rest[:end]
	body := rest[end+len("\n---"):]
	body = strings.TrimLeft(body, "\r\n")
	return fm, body
}

func yamlString(fm, key string) string {
	prefix := key + ":"
	for _, line := range strings.Split(fm, "\n") {
		t := strings.TrimSpace(line)
		if !strings.HasPrefix(t, prefix) {
			continue
		}
		if line != t {
			continue // indented (nested) key
		}
		return strings.Trim(strings.TrimSpace(strings.TrimPrefix(t, prefix)), `"'`)
	}
	return ""
}

// yamlList returns the top-level list scalar values for key. Handles the
// hyphenated form:
//
//	tags:
//	  - cubejs
//	  - analytics
func yamlList(fm, key string) []string {
	prefix := key + ":"
	lines := strings.Split(fm, "\n")
	var out []string
	inList := false
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, prefix) && line == t {
			val := strings.TrimSpace(strings.TrimPrefix(t, prefix))
			if val != "" && val != "[]" {
				val = strings.Trim(val, "[]")
				for _, p := range strings.Split(val, ",") {
					if s := strings.Trim(strings.TrimSpace(p), `"'`); s != "" {
						out = append(out, s)
					}
				}
				return out
			}
			inList = true
			continue
		}
		if inList {
			if t == "" || strings.HasPrefix(t, "#") {
				continue
			}
			if !strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "\t") {
				return out
			}
			t = strings.TrimSpace(strings.TrimPrefix(t, "-"))
			if t != "" {
				out = append(out, strings.Trim(t, `"'`))
			}
		}
	}
	return out
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func singleLine(s string, max int) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > max {
		s = s[:max] + "…"
	}
	return s
}
