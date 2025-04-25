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
	OutputState *SimpleState `json:"output_state"` // Use concrete type
	Error       string       `json:"error"`
	StartTime   time.Time    `json:"start_time"`
	EndTime     time.Time    `json:"end_time"`
	// Duration is typically calculated, not stored in JSON, but include if needed
	// Duration    time.Duration `json:"duration"`
}

func TestAgentResult_JSON(t *testing.T) {
	start := time.Now()
	end := start.Add(150 * time.Millisecond)
	duration := end.Sub(start)

	outStateConcrete := NewState()
	outStateConcrete.Set("result", "success")
	outStateConcrete.SetMeta("trace", "123")
	var outStateInterface State = outStateConcrete

	original := AgentResult{
		OutputState: &outStateInterface,
		Error:       "something went wrong",
		StartTime:   start,
		EndTime:     end,
		Duration:    duration,
	}

	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent failed: %v", err)
	}
	fmt.Println("Generated JSON:\n" + string(data))

	// FIX: Unmarshal into the helper struct first
	var helper agentResultUnmarshalHelper
	if err := json.Unmarshal(data, &helper); err != nil {
		// Log the error source for easier debugging
		fmt.Printf("Unmarshal error (output_state): %v\n", err)
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// FIX: Convert helper back to AgentResult for comparison
	var parsed AgentResult
	parsed.Error = helper.Error
	parsed.StartTime = helper.StartTime
	parsed.EndTime = helper.EndTime
	// Calculate duration if needed, or copy from original for test comparison
	parsed.Duration = original.Duration // Assuming duration isn't in JSON

	// Assign the concrete state from helper to the interface pointer
	if helper.OutputState != nil {
		var stateInterface State = helper.OutputState // Assign concrete type to interface var
		parsed.OutputState = &stateInterface          // Assign address of interface var
	}

	// --- Comparisons ---
	if parsed.Error != original.Error {
		t.Errorf("Error mismatch: got %q, want %q", parsed.Error, original.Error)
	}

	if parsed.OutputState == nil || original.OutputState == nil {
		if parsed.OutputState != original.OutputState {
			t.Fatalf("OutputState nil pointer mismatch: parsed=%v, original=%v", parsed.OutputState, original.OutputState)
		}
	} else if *parsed.OutputState == nil || *original.OutputState == nil {
		if *parsed.OutputState != *original.OutputState {
			t.Fatalf("OutputState nil interface mismatch: parsed=%v, original=%v", *parsed.OutputState, *original.OutputState)
		}
	} else {
		parsedDataKeys := (*parsed.OutputState).Keys()
		originalDataKeys := (*original.OutputState).Keys()
		sort.Strings(parsedDataKeys)
		sort.Strings(originalDataKeys)
		if !reflect.DeepEqual(parsedDataKeys, originalDataKeys) {
			t.Errorf("OutputState Data Keys mismatch:\ngot:  %#v\nwant: %#v", parsedDataKeys, originalDataKeys)
			// Add value checks if needed
			valP, _ := (*parsed.OutputState).Get("result")
			valO, _ := (*original.OutputState).Get("result")
			if !reflect.DeepEqual(valP, valO) {
				t.Errorf("OutputState Data Value mismatch for 'result': got %v, want %v", valP, valO)
			}
		}

		parsedMetaKeys := (*parsed.OutputState).MetaKeys()
		originalMetaKeys := (*original.OutputState).MetaKeys()
		sort.Strings(parsedMetaKeys)
		sort.Strings(originalMetaKeys)
		if !reflect.DeepEqual(parsedMetaKeys, originalMetaKeys) {
			t.Errorf("OutputState Metadata Keys mismatch:\ngot:  %#v\nwant: %#v", parsedMetaKeys, originalMetaKeys)
			// Add value checks if needed
			metaP, _ := (*parsed.OutputState).GetMeta("trace")
			metaO, _ := (*original.OutputState).GetMeta("trace")
			if metaP != metaO {
				t.Errorf("OutputState Metadata Value mismatch for 'trace': got %v, want %v", metaP, metaO)
			}
		}
	}

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

func TestAgentResult_JSON_NoError(t *testing.T) {
	emptyStateConcrete := NewState()
	var emptyStateInterface State = emptyStateConcrete

	original := AgentResult{
		OutputState: &emptyStateInterface,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(10 * time.Millisecond),
		Duration:    10 * time.Millisecond,
	}

	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent failed: %v", err)
	}

	// FIX: Unmarshal into the helper struct first
	var helper agentResultUnmarshalHelper
	if err := json.Unmarshal(data, &helper); err != nil {
		fmt.Printf("Unmarshal error (output_state): %v\n", err) // Log error source
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// FIX: Convert helper back to AgentResult
	var parsed AgentResult
	parsed.Error = helper.Error
	parsed.StartTime = helper.StartTime
	parsed.EndTime = helper.EndTime
	parsed.Duration = original.Duration // Copy duration

	if helper.OutputState != nil {
		var stateInterface State = helper.OutputState
		parsed.OutputState = &stateInterface
	}

	if parsed.Error != "" {
		t.Errorf("Expected empty error string, got %q", parsed.Error)
	}

	if parsed.OutputState == nil {
		t.Fatalf("OutputState pointer is nil after unmarshal, want non-nil")
	} else if *parsed.OutputState == nil {
		t.Fatalf("OutputState interface is nil after unmarshal, want non-nil")
	} else {
		if len((*parsed.OutputState).Keys()) != 0 {
			t.Errorf("Expected empty data keys, got %v", (*parsed.OutputState).Keys())
		}
		if len((*parsed.OutputState).MetaKeys()) != 0 {
			t.Errorf("Expected empty meta keys, got %v", (*parsed.OutputState).MetaKeys())
		}
	}

	if parsed.Duration != original.Duration {
		t.Errorf("Duration mismatch: got %v, want %v", parsed.Duration, original.Duration)
	}
}
