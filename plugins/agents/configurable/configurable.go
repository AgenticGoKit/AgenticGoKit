// Package configurable exposes the internal configurable agent factory via a public plugin import.
//
// Importing this package (via a blank import) ensures the internal agents factory
// is registered with the core registry for use by generated projects.
package configurable

import (
	// Trigger init() in internal/agents to register the ConfigurableAgentFactory.
	_ "github.com/kunalkushwaha/agenticgokit/internal/agents"
)
