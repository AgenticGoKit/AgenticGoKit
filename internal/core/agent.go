package agentflow

import (
	"context"
)

// Agent represents a unit of work within a workflow.
// It receives an input State, performs an operation potentially using context,
// and returns an output State or an error.
//
// Example Usage:
//
//	type MyAgent struct { /* ... fields ... */ }
//
//	func (a *MyAgent) Run(ctx context.Context, in State) (State, error) {
//	    // Access input data: data := in.GetData()
//	    // Perform work...
//	    // Check context cancellation: if ctx.Err() != nil { return in, ctx.Err() }
//	    // Create output state: out := in.Clone()
//	    // Modify output data: out.SetData("result", "some value")
//	    return out, nil
//	}
type Agent interface {
	Run(ctx context.Context, in State) (out State, err error)
}

// Note: The previous Agent interface (with Handle(Event)) might need to be
// renamed or refactored depending on how event handling and workflow execution
// will coexist or be integrated. For now, we define the new one as requested.

// NewStateWithData creates a new SimpleState initialized with the provided data map.
func NewStateWithData(data map[string]any) State {
	s := NewState()
	if data != nil {
		s.data = make(map[string]any, len(data))
		for k, v := range data {
			// TODO: Implement proper deep copy for complex value types if needed
			s.data[k] = v
		}
	}
	return s
}
