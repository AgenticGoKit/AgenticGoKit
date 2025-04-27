package agentflow

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort" // Add sort import
	"testing"
	"time"
)

// Helper struct for unmarshalling AgentResult with concrete state type
type agentResultUnmarshalHelper struct {
	OutputState *SimpleState `json:"output_state"`    // Use concrete type pointer (JSON lib often works best with pointers for structs)
	Error       string       `json:"error,omitempty"` // Keep omitempty
	StartTime   time.Time    `json:"start_time"`
	EndTime     time.Time    `json:"end_time"`
	DurationMs  int64        `json:"duration_ms"` // Match MarshalJSON output
}

func TestAgentResult_JSON(t *testing.T) {
	start := time.Now()
	end := start.Add(150 * time.Millisecond)
	duration := end.Sub(start)

	outStateConcrete := NewState() // Returns *SimpleState
	outStateConcrete.Set("result", "success")
	outStateConcrete.SetMeta("trace", "123")

	original := AgentResult{
		OutputState: outStateConcrete, // CHANGED: Assign concrete type directly
		Error:       "something went wrong",
		StartTime:   start,
		EndTime:     end,
		Duration:    duration,
	}

	// FIX: Pass a pointer to original to ensure MarshalJSON is called
	data, err := json.MarshalIndent(&original, "", "  ") // Pass &original
	if err != nil {
		t.Fatalf("json.MarshalIndent failed: %v", err)
	}
	fmt.Println("Generated JSON:\n" + string(data)) // Check this output carefully now

	// Unmarshal back into AgentResult directly (using custom UnmarshalJSON)
	var parsed AgentResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		// Log the error source for easier debugging
		fmt.Printf("Unmarshal error: %v\n", err)
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// --- Comparisons ---
	if parsed.Error != original.Error {
		t.Errorf("Error mismatch: got %q, want %q", parsed.Error, original.Error)
	}

	// Compare OutputState
	if parsed.OutputState == nil || original.OutputState == nil {
		if parsed.OutputState != original.OutputState { // Check if one is nil and the other isn't
			t.Fatalf("OutputState nil mismatch: parsed=%v, original=%v", parsed.OutputState, original.OutputState)
		}
		// Both are nil, which is okay
	} else {
		// Both are non-nil, compare contents
		parsedDataKeys := parsed.OutputState.Keys()
		originalDataKeys := original.OutputState.Keys()
		sort.Strings(parsedDataKeys)
		sort.Strings(originalDataKeys)
		if !reflect.DeepEqual(parsedDataKeys, originalDataKeys) {
			t.Errorf("OutputState Data Keys mismatch:\ngot:  %#v\nwant: %#v", parsedDataKeys, originalDataKeys)
		} else {
			// Add value checks if needed
			valP, _ := parsed.OutputState.Get("result")
			valO, _ := original.OutputState.Get("result")
			if !reflect.DeepEqual(valP, valO) {
				t.Errorf("OutputState Data Value mismatch for 'result': got %v, want %v", valP, valO)
			}
		}

		parsedMetaKeys := parsed.OutputState.MetaKeys()
		originalMetaKeys := original.OutputState.MetaKeys()
		sort.Strings(parsedMetaKeys)
		sort.Strings(originalMetaKeys)
		if !reflect.DeepEqual(parsedMetaKeys, originalMetaKeys) {
			t.Errorf("OutputState Metadata Keys mismatch:\ngot:  %#v\nwant: %#v", parsedMetaKeys, originalMetaKeys)
		} else {
			// Add value checks if needed
			metaP, _ := parsed.OutputState.GetMeta("trace")
			metaO, _ := original.OutputState.GetMeta("trace")
			if metaP != metaO {
				t.Errorf("OutputState Metadata Value mismatch for 'trace': got %v, want %v", metaP, metaO)
			}
		}
	}

	// Compare time fields (allowing for slight precision differences)
	if !parsed.StartTime.Truncate(time.Millisecond).Equal(original.StartTime.Truncate(time.Millisecond)) {
		t.Errorf("StartTime mismatch: got %v, want %v", parsed.StartTime, original.StartTime)
	}
	if !parsed.EndTime.Truncate(time.Millisecond).Equal(original.EndTime.Truncate(time.Millisecond)) {
		t.Errorf("EndTime mismatch: got %v, want %v", parsed.EndTime, original.EndTime)
	}
	// Compare calculated duration (UnmarshalJSON calculates it from DurationMs)
	if parsed.Duration != original.Duration {
		t.Errorf("Duration mismatch: got %v, want %v", parsed.Duration, original.Duration)
	}
}

func TestAgentResult_JSON_NoError(t *testing.T) {
	emptyStateConcrete := NewState() // Returns *SimpleState

	original := AgentResult{
		OutputState: emptyStateConcrete, // CHANGED: Assign concrete type directly
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(10 * time.Millisecond),
		Duration:    10 * time.Millisecond,
	}

	// FIX: Pass a pointer to original to ensure MarshalJSON is called
	data, err := json.MarshalIndent(&original, "", "  ") // Pass &original
	if err != nil {
		t.Fatalf("json.MarshalIndent failed: %v", err)
	}
	// Optional: fmt.Println("Generated JSON (NoError):\n" + string(data))

	// Unmarshal back into AgentResult directly
	var parsed AgentResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		fmt.Printf("Unmarshal error: %v\n", err) // Log error source
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// --- Comparisons ---
	if parsed.Error != "" {
		t.Errorf("Expected empty error string, got %q", parsed.Error)
	}

	if parsed.OutputState == nil { // Check if interface is nil
		t.Fatalf("OutputState is nil after unmarshal, want non-nil empty state")
	} else {
		if len(parsed.OutputState.Keys()) != 0 {
			t.Errorf("Expected empty data keys, got %v", parsed.OutputState.Keys())
		}
		if len(parsed.OutputState.MetaKeys()) != 0 {
			t.Errorf("Expected empty meta keys, got %v", parsed.OutputState.MetaKeys())
		}
	}

	// Compare time fields
	if !parsed.StartTime.Truncate(time.Millisecond).Equal(original.StartTime.Truncate(time.Millisecond)) {
		t.Errorf("StartTime mismatch: got %v, want %v", parsed.StartTime, original.StartTime)
	}
	if !parsed.EndTime.Truncate(time.Millisecond).Equal(original.EndTime.Truncate(time.Millisecond)) {
		t.Errorf("EndTime mismatch: got %v, want %v", parsed.EndTime, original.EndTime)
	}
	if parsed.Duration != original.Duration {
		t.Errorf("Duration mismatch: got %v, want %v", parsed.Duration, original.Duration)
	}
}
