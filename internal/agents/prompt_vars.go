package agents

import (
	agenticgokit "github.com/agenticgokit/agenticgokit/internal/core"
)

// internalStateVars snapshots an internal/core.State's data map into a plain
// map[string]any for core.FormatPromptString. Mirrors core.StateVars, which
// can't be reused here because internal/core.State and core.State are
// distinct interfaces (differing Clone() return types).
func internalStateVars(state agenticgokit.State) map[string]any {
	keys := state.Keys()
	vs := make(map[string]any, len(keys))
	for _, k := range keys {
		if v, ok := state.Get(k); ok {
			vs[k] = v
		}
	}
	return vs
}
