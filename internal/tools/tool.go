package tools

import (
	"context"
	"fmt"
	"sync"
)

// FunctionTool defines the interface for a callable tool that agents can use.
// Arguments and return values are expected to be JSON-serializable maps.
type FunctionTool interface {
	// Name returns the unique identifier for the tool.
	Name() string
	// Call executes the tool's logic with the given arguments.
	// It returns a map containing the results or an error if the call fails.
	Call(ctx context.Context, args map[string]any) (map[string]any, error)
	// TODO: Add a Schema() method to describe parameters and return values for LLMs?
}

// ToolRegistry holds a collection of available FunctionTools.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]FunctionTool
}

// NewToolRegistry creates an empty tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]FunctionTool),
	}
}

// Register adds a tool to the registry.
// It returns an error if a tool with the same name is already registered.
func (r *ToolRegistry) Register(tool FunctionTool) error {
	if tool == nil {
		return fmt.Errorf("cannot register a nil tool")
	}
	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool '%s' is already registered", name)
	}
	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by its name.
// It returns the tool and true if found, otherwise nil and false.
func (r *ToolRegistry) Get(name string) (FunctionTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, exists := r.tools[name]
	return tool, exists
}

// List returns the names of all registered tools.
func (r *ToolRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// CallTool looks up a tool by name and executes its Call method.
func (r *ToolRegistry) CallTool(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	tool, exists := r.Get(name)
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found in registry", name)
	}
	return tool.Call(ctx, args)
}
