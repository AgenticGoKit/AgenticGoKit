package tools

import (
	"context"
	"fmt"
	"log"
)

// ComputeMetricTool performs simple calculations.
type ComputeMetricTool struct{}

// Name returns the tool's name.
func (t *ComputeMetricTool) Name() string {
	return "compute_metric"
}

// Call performs the calculation based on the 'operation' argument.
// Expects "operation" (string: "add", "subtract") and "a", "b" (float64) in args.
// Returns "result" (float64) in the result map.
func (t *ComputeMetricTool) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
	opVal, ok := args["operation"]
	if !ok {
		return nil, fmt.Errorf("missing required argument 'operation'")
	}
	operation, ok := opVal.(string)
	if !ok {
		return nil, fmt.Errorf("argument 'operation' must be a string")
	}

	aVal, ok := args["a"]
	if !ok {
		return nil, fmt.Errorf("missing required argument 'a'")
	}
	// Use type assertion with check for float64 (JSON numbers often decode to float64)
	a, ok := aVal.(float64)
	if !ok {
		// Allow int promotion to float64
		if aInt, isInt := aVal.(int); isInt {
			a = float64(aInt)
			ok = true
		} else {
			return nil, fmt.Errorf("argument 'a' must be a number (float64 or int)")
		}
	}

	bVal, ok := args["b"]
	if !ok {
		return nil, fmt.Errorf("missing required argument 'b'")
	}
	b, ok := bVal.(float64)
	if !ok {
		if bInt, isInt := bVal.(int); isInt {
			b = float64(bInt)
			ok = true
		} else {
			return nil, fmt.Errorf("argument 'b' must be a number (float64 or int)")
		}
	}

	log.Printf("ComputeMetricTool: Performing operation '%s' on a=%.2f, b=%.2f", operation, a, b)

	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	// TODO: Add more operations like multiply, divide
	default:
		return nil, fmt.Errorf("unsupported operation '%s'", operation)
	}

	return map[string]any{
		"result": result,
	}, nil
}
