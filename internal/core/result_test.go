package agentflow

import (
	"encoding/json"
	"fmt" // Make sure fmt is imported
	"reflect"
	"testing"
	"time"
)

func TestAgentResult_JSON(t *testing.T) {
	start := time.Now()
	end := start.Add(150 * time.Millisecond)
	duration := end.Sub(start)

	outState := NewState() // NewState returns State value
	outState.Set("result", "success")
	outState.SetMeta("trace", "123")

	original := AgentResult{
		OutputState: &outState, // Pass the address of the State value
		Error:       "something went wrong",
		StartTime:   start,
		EndTime:     end,
		Duration:    duration,
	}

	// Marshal the ADDRESS of the original struct
	data, err := json.MarshalIndent(&original, "", "  ") // <--- CHANGE HERE: Pass &original
	if err != nil {
		t.Fatalf("json.MarshalIndent failed: %v", err)
	}

	fmt.Println("Generated JSON:\n" + string(data)) // Uncommented this line

	var parsed AgentResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// --- Comparisons ---
	if parsed.Error != original.Error {
		t.Errorf("Error mismatch: got %q, want %q", parsed.Error, original.Error)
	}

	// Check if OutputState is nil before accessing its methods
	if parsed.OutputState == nil {
		t.Fatalf("OutputState is nil after unmarshal, want non-nil")
	}
	if original.OutputState == nil {
		// This shouldn't happen in this test case, but good practice
		t.Fatalf("Original OutputState is unexpectedly nil")
	}

	// Use reflect.DeepEqual for State maps (after getting them)
	if !reflect.DeepEqual(parsed.OutputState.GetData(), original.OutputState.GetData()) {
		t.Errorf("OutputState Data mismatch:\ngot:  %#v\nwant: %#v", parsed.OutputState.GetData(), original.OutputState.GetData())
	}
	if !reflect.DeepEqual(parsed.OutputState.GetMetadata(), original.OutputState.GetMetadata()) {
		t.Errorf("OutputState Metadata mismatch:\ngot:  %#v\nwant: %#v", parsed.OutputState.GetMetadata(), original.OutputState.GetMetadata())
	}
	// Compare time up to millisecond precision as JSON might lose nanoseconds
	if !parsed.StartTime.Truncate(time.Millisecond).Equal(original.StartTime.Truncate(time.Millisecond)) {
		t.Errorf("StartTime mismatch: got %v, want %v", parsed.StartTime, original.StartTime)
	}
	if !parsed.EndTime.Truncate(time.Millisecond).Equal(original.EndTime.Truncate(time.Millisecond)) {
		t.Errorf("EndTime mismatch: got %v, want %v", parsed.EndTime, original.EndTime)
	}
	// Compare duration directly
	if parsed.Duration != original.Duration {
		t.Errorf("Duration mismatch: got %v, want %v", parsed.Duration, original.Duration)
	}
}

func TestAgentResult_JSON_NoError(t *testing.T) {
	// Ensure OutputState is initialized correctly for pointer usage
	emptyState := NewState()
	original := AgentResult{
		OutputState: &emptyState, // Pass address
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(10 * time.Millisecond),
		Duration:    10 * time.Millisecond,
	}

	// Also marshal the address here
	data, err := json.MarshalIndent(&original, "", "  ") // <--- CHANGE HERE: Pass &original
	if err != nil {
		t.Fatalf("json.MarshalIndent failed: %v", err)
	}

	var parsed AgentResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if parsed.Error != "" {
		t.Errorf("Expected empty error string, got %q", parsed.Error)
	}
	if parsed.OutputState == nil {
		t.Fatalf("OutputState is nil after unmarshal, want non-nil")
	}
	// Add other checks as needed, e.g., duration
	if parsed.Duration != original.Duration {
		t.Errorf("Duration mismatch: got %v, want %v", parsed.Duration, original.Duration)
	}
}
