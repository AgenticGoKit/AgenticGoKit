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

// getFloat is a helper function to safely extract a float64 from the arguments map.
// It handles both float64 and int types, promoting ints to float64.
func getFloat(args map[string]any, key string) (float64, error) {
	val, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("missing required argument '%s'", key)
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("argument '%s' must be a number (float64 or int)", key)
	}
}

// Call performs the calculation based on the 'operation' argument.
// Expects "operation" (string: "add", "subtract", "multiply", "divide") and "a", "b" (float64 or int) in args.
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

	a, err := getFloat(args, "a")
	if err != nil {
		return nil, err
	}

	b, err := getFloat(args, "b")
	if err != nil {
		return nil, err
	}

	log.Printf("ComputeMetricTool: Performing operation '%s' on a=%.2f, b=%.2f", operation, a, b)

	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero is not allowed")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unsupported operation '%s'", operation)
	}

	return map[string]any{
		"result": result,
	}, nil
}
