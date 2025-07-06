package tools

import (
	"context"
	"reflect"
	"testing"
)

func TestComputeMetricTool(t *testing.T) {
	tool := &ComputeMetricTool{}
	ctx := context.Background()

	if tool.Name() != "compute_metric" {
		t.Errorf("Name() mismatch: got %s, want compute_metric", tool.Name())
	}

	testCases := []struct {
		name         string
		args         map[string]any
		expectResult map[string]any
		expectErr    string
	}{
		{
			name:         "Add floats",
			args:         map[string]any{"operation": "add", "a": 10.5, "b": 5.5},
			expectResult: map[string]any{"result": 16.0},
			expectErr:    "",
		},
		{
			name:         "Add ints",
			args:         map[string]any{"operation": "add", "a": 10, "b": 5},
			expectResult: map[string]any{"result": 15.0}, // Result is float64
			expectErr:    "",
		},
		{
			name:         "Add mixed",
			args:         map[string]any{"operation": "add", "a": 10, "b": 5.5},
			expectResult: map[string]any{"result": 15.5},
			expectErr:    "",
		},
		{
			name:         "Subtract floats",
			args:         map[string]any{"operation": "subtract", "a": 10.5, "b": 5.0},
			expectResult: map[string]any{"result": 5.5},
			expectErr:    "",
		},
		{
			name:         "Subtract ints",
			args:         map[string]any{"operation": "subtract", "a": 10, "b": 5},
			expectResult: map[string]any{"result": 5.0},
			expectErr:    "",
		},
		{
			name:         "Missing operation",
			args:         map[string]any{"a": 10, "b": 5},
			expectResult: nil,
			expectErr:    "missing required argument 'operation'",
		},
		{
			name:         "Missing a",
			args:         map[string]any{"operation": "add", "b": 5},
			expectResult: nil,
			expectErr:    "missing required argument 'a'",
		},
		{
			name:         "Missing b",
			args:         map[string]any{"operation": "add", "a": 10},
			expectResult: nil,
			expectErr:    "missing required argument 'b'",
		},
		{
			name:         "Invalid operation type",
			args:         map[string]any{"operation": 123, "a": 10, "b": 5},
			expectResult: nil,
			expectErr:    "argument 'operation' must be a string",
		},
		{
			name:         "Invalid a type",
			args:         map[string]any{"operation": "add", "a": "ten", "b": 5},
			expectResult: nil,
			expectErr:    "argument 'a' must be a number (float64 or int)",
		},
		{
			name:         "Invalid b type",
			args:         map[string]any{"operation": "add", "a": 10, "b": "five"},
			expectResult: nil,
			expectErr:    "argument 'b' must be a number (float64 or int)",
		},
		{
			name:         "Multiply ints",
			args:         map[string]any{"operation": "multiply", "a": 10, "b": 5},
			expectResult: map[string]any{"result": 50.0},
			expectErr:    "",
		},
		{
			name:         "Divide ints",
			args:         map[string]any{"operation": "divide", "a": 10, "b": 5},
			expectResult: map[string]any{"result": 2.0},
			expectErr:    "",
		},
		{
			name:         "Division by zero",
			args:         map[string]any{"operation": "divide", "a": 1, "b": 0},
			expectResult: nil,
			expectErr:    "division by zero is not allowed",
		},
		{
			name:         "Unsupported operation",
			args:         map[string]any{"operation": "mod", "a": 10, "b": 5},
			expectResult: nil,
			expectErr:    "unsupported operation 'mod'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tool.Call(ctx, tc.args)

			if tc.expectErr != "" {
				if err == nil {
					t.Fatalf("Expected error '%s', but got nil", tc.expectErr)
				}
				if err.Error() != tc.expectErr {
					t.Fatalf("Error mismatch: got '%v', want '%s'", err, tc.expectErr)
				}
				if result != nil {
					t.Errorf("Expected nil result on error, but got %v", result)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if !reflect.DeepEqual(result, tc.expectResult) {
					t.Errorf("Result mismatch: got %v, want %v", result, tc.expectResult)
				}
			}
		})
	}
}
