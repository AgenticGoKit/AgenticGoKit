package tools

import (
	"context"
	"reflect"
	"testing"
)

func TestWebSearchTool(t *testing.T) {
	tool := &WebSearchTool{}
	ctx := context.Background()

	if tool.Name() != "web_search" {
		t.Errorf("Name() mismatch: got %s, want web_search", tool.Name())
	}

	testCases := []struct {
		name         string
		args         map[string]any
		expectResult map[string]any
		expectErr    string
	}{
		{
			name: "Valid query",
			args: map[string]any{"query": "test search"},
			expectResult: map[string]any{
				"results": []string{"Showing results for 'test search'", "Result 1", "Result 2"},
			},
			expectErr: "",
		},
		{
			name: "Specific query",
			args: map[string]any{"query": "capital of France"},
			expectResult: map[string]any{
				"results": []string{"Paris is the capital of France.", "France is a country in Europe."},
			},
			expectErr: "",
		},
		{
			name:         "Missing query",
			args:         map[string]any{},
			expectResult: nil,
			expectErr:    "missing required argument 'query'",
		},
		{
			name:         "Empty query",
			args:         map[string]any{"query": ""},
			expectResult: nil,
			expectErr:    "argument 'query' must be a non-empty string",
		},
		{
			name:         "Invalid query type",
			args:         map[string]any{"query": 123},
			expectResult: nil,
			expectErr:    "argument 'query' must be a non-empty string",
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
