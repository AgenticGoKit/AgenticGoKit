package defaultorchestrator

import (
	"github.com/kunalkushwaha/agenticgokit/core"
	internal "github.com/kunalkushwaha/agenticgokit/internal/orchestrator"
)

// init registers the internal orchestrator factory with the core package via a public plugin.
// This allows third-party consumers to enable orchestrators using a blank import of this package.
func init() {
	core.RegisterOrchestratorFactory(func(cfg core.OrchestratorConfig, registry *core.CallbackRegistry) (core.Orchestrator, error) {
		// Convert core config to internal config
		icfg := internal.OrchestratorConfig{
			Type:                internal.OrchestratorType(cfg.Type),
			AgentSequence:       cfg.AgentNames,
			MaxIterations:       cfg.MaxIterations,
			CollaborativeAgents: cfg.CollaborativeAgentNames,
			SequentialAgents:    cfg.SequentialAgentNames,
		}
		// Create and delegate to internal factory
		f := internal.NewOrchestratorFactory(registry)
		return f.CreateOrchestrator(icfg)
	})
}
