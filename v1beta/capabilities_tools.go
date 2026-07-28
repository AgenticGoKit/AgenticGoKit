package v1beta

import (
	"context"
	"fmt"
)

// sliceToolManager adapts the already-discovered []Tool slice (populated by
// createTools during Build(), consumed today only by step 3.5's
// executeToolsAndContinue/executeNativeToolsAndContinue) into the ToolManager
// interface, so a custom HandlerFunc's Capabilities.Tools is real instead of
// always nil.
//
// Why this exists, not NewToolManager/basicToolManager: ToolManager's own
// constructor (tools.go:159) falls back to basicToolManager whenever no
// ToolManagerFactory is registered via SetToolManagerFactory — and nothing in
// this module or its plugins/* ever calls that (confirmed repo-wide grep).
// basicToolManager.Execute unconditionally returns "no tool plugin
// registered" — every ToolManager built via NewToolManager in this codebase
// is that non-functional stub today. Wiring capabilities.Tools to it would
// compile and look wired without ever actually running a tool, silently
// worse than the current nil (docs/v1beta/custom-handlers.md's own
// "Tool-Augmented"/"Research Assistant" examples call capabilities.Tools.Execute
// unconditionally and would get this error text back instead of a result).
// sliceToolManager instead executes against the SAME []Tool the framework's
// own auto-tool-exec step already resolved and connected (MCP included),
// so a handler sees exactly the tools WithTools(...) configured.
//
// Built fresh per call (agent_impl.go's two Capabilities{} construction
// sites), not cached on realAgent — a.tools is intentionally mutated
// per-call by RunWithOptions' ToolMode="specific"/"none" filtering
// (agent_impl.go:776-796), and a stale cached manager would silently ignore
// that per-run restriction.
type sliceToolManager struct {
	tools []Tool
}

func newSliceToolManager(tools []Tool) ToolManager {
	return &sliceToolManager{tools: tools}
}

func (m *sliceToolManager) Execute(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	for _, t := range m.tools {
		if t.Name() == name {
			return t.Execute(ctx, args)
		}
	}
	return &ToolResult{Success: false, Error: fmt.Sprintf("tool %q not found among this agent's configured tools", name)},
		fmt.Errorf("tool %q not found", name)
}

func (m *sliceToolManager) List() []ToolInfo {
	out := make([]ToolInfo, 0, len(m.tools))
	for _, t := range m.tools {
		info := ToolInfo{Name: t.Name(), Description: t.Description()}
		if ws, ok := t.(ToolWithSchema); ok {
			info.Parameters = ws.JSONSchema()
		}
		out = append(out, info)
	}
	return out
}

func (m *sliceToolManager) Available() []string {
	out := make([]string, 0, len(m.tools))
	for _, t := range m.tools {
		out = append(out, t.Name())
	}
	return out
}

func (m *sliceToolManager) IsAvailable(name string) bool {
	for _, t := range m.tools {
		if t.Name() == name {
			return true
		}
	}
	return false
}

// MCP connection management, health, and metrics are intentionally NOT
// no-op-succeed here: this adapter wraps a fixed snapshot of tools resolved
// once at Build()/RunWithOptions time, and pretending an MCP reconnect or a
// health check ran against it (returning a fake nil/empty success) is the
// exact swallowed-error shape that already caused one real production
// incident in this framework (the stdio transport argv-passing bug
// surfacing only as "0 tools, suspiciously fast" instead of a build error —
// see go-implementation-gotchas). A clear error instead tells the caller
// what's actually true: reconfigure tools via WithTools(WithMCP(...)) and
// rebuild/Clone() the agent, there is no dynamic reconnect at this layer.
func (m *sliceToolManager) ConnectMCP(ctx context.Context, servers ...MCPServer) error {
	return fmt.Errorf("sliceToolManager: dynamic MCP connection not supported; configure servers via WithTools(WithMCP(...)) and rebuild the agent")
}

func (m *sliceToolManager) DisconnectMCP(serverName string) error {
	return fmt.Errorf("sliceToolManager: dynamic MCP disconnection not supported; tools are fixed at Build() time")
}

func (m *sliceToolManager) DiscoverMCP(ctx context.Context) ([]MCPServerInfo, error) {
	return nil, fmt.Errorf("sliceToolManager: MCP server discovery not supported; use v1beta.DiscoverTools() or inspect config.Tools.MCP instead")
}

func (m *sliceToolManager) HealthCheck(ctx context.Context) map[string]MCPHealthStatus {
	return map[string]MCPHealthStatus{}
}

func (m *sliceToolManager) GetMetrics() ToolMetrics {
	return ToolMetrics{ToolMetrics: make(map[string]ToolSpecificMetrics)}
}

func (m *sliceToolManager) Initialize(ctx context.Context) error { return nil }
func (m *sliceToolManager) Shutdown(ctx context.Context) error   { return nil }
