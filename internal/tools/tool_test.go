package tools

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// MockTool for testing registry
type MockTool struct {
	name    string
	callFn  func(ctx context.Context, args map[string]any) (map[string]any, error)
	callErr error
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Call(ctx context.Context, args map[string]any) (map[string]any, error) {
	if m.callFn != nil {
		return m.callFn(ctx, args)
	}
	if m.callErr != nil {
		return nil, m.callErr
	}
	return map[string]any{"mock_result": "success", "input_args": args}, nil
}

func TestToolRegistry(t *testing.T) {
	registry := NewToolRegistry()

	tool1 := &MockTool{name: "tool1"}
	tool2 := &MockTool{name: "tool2"}
	tool1Duplicate := &MockTool{name: "tool1"} // Same name

	// Test Register success
	err := registry.Register(tool1)
	if err != nil {
		t.Fatalf("Register(tool1) failed: %v", err)
	}
	err = registry.Register(tool2)
	if err != nil {
		t.Fatalf("Register(tool2) failed: %v", err)
	}

	// Test Register duplicate
	err = registry.Register(tool1Duplicate)
	if err == nil {
		t.Fatal("Register(tool1Duplicate) should have failed but didn't")
	}
	expectedErr := "tool 'tool1' is already registered"
	if err.Error() != expectedErr {
		t.Errorf("Register duplicate error mismatch: got '%v', want '%s'", err, expectedErr)
	}

	// Test Register nil tool
	err = registry.Register(nil)
	if err == nil {
		t.Fatal("Register(nil) should have failed but didn't")
	}
	if err.Error() != "cannot register a nil tool" {
		t.Errorf("Register nil error mismatch: got '%v', want 'cannot register a nil tool'", err)
	}

	// Test Register empty name
	err = registry.Register(&MockTool{name: ""})
	if err == nil {
		t.Fatal("Register empty name should have failed but didn't")
	}
	if err.Error() != "tool name cannot be empty" {
		t.Errorf("Register empty name error mismatch: got '%v', want 'tool name cannot be empty'", err)
	}

	// Test Get found
	retrievedTool, found := registry.Get("tool1")
	if !found {
		t.Fatal("Get('tool1') failed: tool not found")
	}
	if retrievedTool != tool1 {
		t.Error("Get('tool1') returned wrong tool instance")
	}

	// Test Get not found
	_, found = registry.Get("nonexistent")
	if found {
		t.Fatal("Get('nonexistent') should not have found a tool")
	}

	// Test List
	expectedList := []string{"tool1", "tool2"}
	actualList := registry.List()
	sort.Strings(actualList) // Sort for consistent comparison
	if !reflect.DeepEqual(actualList, expectedList) {
		t.Errorf("List() mismatch: got %v, want %v", actualList, expectedList)
	}

	// Test CallTool success
	ctx := context.Background()
	args := map[string]any{"arg1": 123}
	expectedResult := map[string]any{"mock_result": "success", "input_args": args}
	actualResult, err := registry.CallTool(ctx, "tool1", args)
	if err != nil {
		t.Fatalf("CallTool('tool1') failed: %v", err)
	}
	if !reflect.DeepEqual(actualResult, expectedResult) {
		t.Errorf("CallTool('tool1') result mismatch: got %v, want %v", actualResult, expectedResult)
	}

	// Test CallTool not found
	_, err = registry.CallTool(ctx, "nonexistent", args)
	if err == nil {
		t.Fatal("CallTool('nonexistent') should have failed but didn't")
	}
	expectedErr = "tool 'nonexistent' not found in registry"
	if err.Error() != expectedErr {
		t.Errorf("CallTool not found error mismatch: got '%v', want '%s'", err, expectedErr)
	}

	// Test CallTool underlying error
	errorTool := &MockTool{name: "errorTool", callErr: fmt.Errorf("internal tool error")}
	registry.Register(errorTool)
	_, err = registry.CallTool(ctx, "errorTool", args)
	if err == nil {
		t.Fatal("CallTool('errorTool') should have failed but didn't")
	}
	if err.Error() != "internal tool error" {
		t.Errorf("CallTool underlying error mismatch: got '%v', want 'internal tool error'", err)
	}
}
