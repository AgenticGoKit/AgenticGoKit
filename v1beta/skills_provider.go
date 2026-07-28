package v1beta

import "sync"

// SkillsProvider is implemented by a skills-loading plugin (e.g.
// plugins/skills) and returned by a registered SkillsProviderFactory. Kept
// deliberately narrow — just enough for createTools() to wire the load_skill
// tool in and splice the rendered catalog into the agent's SystemPrompt —
// not a general plugin API.
type SkillsProvider interface {
	// Tool returns the load_skill Tool itself, ready to append to the
	// agent's discovered []Tool slice.
	Tool() Tool
	// Catalog returns the boot-time frontmatter index rendered as markdown,
	// for splicing into Config.SystemPrompt. Bodies are never included here
	// — see the plugin's own doc comment for why (flat memory regardless of
	// corpus size; only Tool().Execute reads a body, on demand).
	Catalog() string
}

// SkillsProviderFactory builds a SkillsProvider for the skills directory
// named in ToolsConfig.Skills.Dir.
type SkillsProviderFactory func(dir string) (SkillsProvider, error)

// SetSkillsProviderFactory registers the factory a skills plugin provides.
// Same registration shape as SetToolManagerFactory (tools.go) and
// RegisterLoggingProvider (core) — v1beta defines the hook, the plugin
// self-registers via its own init() on blank import, so v1beta never needs
// to depend on the plugin (which itself imports v1beta for Tool/ToolResult
// — the dependency can only run one direction).
func SetSkillsProviderFactory(factory SkillsProviderFactory) {
	skillsFactoryMutex.Lock()
	defer skillsFactoryMutex.Unlock()
	skillsProviderFactory = factory
}

var (
	skillsProviderFactory SkillsProviderFactory
	skillsFactoryMutex    sync.RWMutex
)

// GetSkillsProviderFactory returns the currently registered factory, or nil
// if no skills plugin has been blank-imported. Exported mainly so a plugin's
// own tests can confirm its init() actually ran (there's no other external
// signal for that short of a live Build() call), but also useful for a
// caller that wants to detect/log a missing blank-import before relying on
// ToolsConfig.Skills.Dir.
func GetSkillsProviderFactory() SkillsProviderFactory {
	skillsFactoryMutex.RLock()
	defer skillsFactoryMutex.RUnlock()
	return skillsProviderFactory
}
