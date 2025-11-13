package memoryregistry

import (
	"context"
	"fmt"
	"sync"

	"github.com/agenticgokit/agenticgokit/core"
)

// Plugin: In-memory FunctionToolRegistry extracted from core
// Provides a thread-safe registry for function tools and registers via factory hook.

type inMemoryFunctionToolRegistry struct {
	tools map[string]core.FunctionTool
	mu    sync.RWMutex
}

func newInMemoryFunctionToolRegistry() core.FunctionToolRegistry {
	return &inMemoryFunctionToolRegistry{
		tools: make(map[string]core.FunctionTool),
	}
}

func (r *inMemoryFunctionToolRegistry) Register(tool core.FunctionTool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	r.tools[name] = tool
	core.Logger().Info().Str("tool", name).Msg("Registered function tool")
	return nil
}

func (r *inMemoryFunctionToolRegistry) Get(name string) (core.FunctionTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	return tool, exists
}

func (r *inMemoryFunctionToolRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}

	return names
}

func (r *inMemoryFunctionToolRegistry) CallTool(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	r.mu.RLock()
	tool, exists := r.tools[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	result, err := tool.Call(ctx, args)
	if err != nil {
		core.Logger().Error().
			Str("tool", name).
			Err(err).
			Msg("Tool execution failed")
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}

// Register factory on import
func init() {
	core.SetFunctionToolRegistryFactory(func() (core.FunctionToolRegistry, error) {
		return newInMemoryFunctionToolRegistry(), nil
	})
}

